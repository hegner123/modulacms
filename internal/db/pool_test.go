package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	config "github.com/hegner123/modulacms/internal/config"
)

func tempSQLiteConfig(t *testing.T) config.Config {
	t.Helper()
	dir := t.TempDir()
	return config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "test.db"),
	}
}

// --- buildDSN tests ---

func TestBuildDSN_SQLite(t *testing.T) {
	cfg := config.Config{Db_Driver: config.Sqlite, Db_URL: "/tmp/test.db"}
	driver, dsn, err := buildDSN(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver != "sqlite3" {
		t.Errorf("driver = %q, want %q", driver, "sqlite3")
	}
	if dsn != "/tmp/test.db" {
		t.Errorf("dsn = %q, want %q", dsn, "/tmp/test.db")
	}
}

func TestBuildDSN_SQLite_Default(t *testing.T) {
	cfg := config.Config{Db_Driver: config.Sqlite, Db_URL: ""}
	_, dsn, err := buildDSN(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dsn != "./modula.db" {
		t.Errorf("dsn = %q, want %q", dsn, "./modula.db")
	}
}

func TestBuildDSN_MySQL(t *testing.T) {
	cfg := config.Config{
		Db_Driver:   config.Mysql,
		Db_User:     "root",
		Db_Password: "secret",
		Db_URL:      "localhost:3306",
		Db_Name:     "cms",
	}
	driver, dsn, err := buildDSN(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver != "mysql" {
		t.Errorf("driver = %q, want %q", driver, "mysql")
	}
	if dsn != "root:secret@tcp(localhost:3306)/cms?parseTime=true" {
		t.Errorf("dsn = %q, want %q", dsn, "root:secret@tcp(localhost:3306)/cms?parseTime=true")
	}
}

func TestBuildDSN_PostgreSQL(t *testing.T) {
	cfg := config.Config{
		Db_Driver:   config.Psql,
		Db_User:     "admin",
		Db_Password: "pass",
		Db_URL:      "localhost:5432",
		Db_Name:     "cms",
	}
	driver, dsn, err := buildDSN(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver != "postgres" {
		t.Errorf("driver = %q, want %q", driver, "postgres")
	}
	if !strings.HasPrefix(dsn, "postgres://admin:pass@localhost:5432/cms") {
		t.Errorf("dsn = %q, expected postgres URL prefix", dsn)
	}
	if !strings.Contains(dsn, "sslmode=disable") {
		t.Errorf("dsn = %q, expected sslmode=disable", dsn)
	}
}

func TestBuildDSN_UnsupportedDriver(t *testing.T) {
	cfg := config.Config{Db_Driver: "oracle"}
	_, _, err := buildDSN(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Errorf("error = %q, expected 'unsupported database driver'", err.Error())
	}
}

// --- OpenPool tests ---

func TestOpenPool_SQLite(t *testing.T) {
	cfg := tempSQLiteConfig(t)
	pc := PoolConfig{MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifetime: time.Minute}

	pool, err := OpenPool(cfg, pc)
	if err != nil {
		t.Fatalf("OpenPool failed: %v", err)
	}
	defer pool.Close()

	// Verify pool is alive
	if err := pool.Ping(); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Verify WAL mode is active
	var journalMode string
	if err := pool.QueryRow("PRAGMA journal_mode;").Scan(&journalMode); err != nil {
		t.Fatalf("PRAGMA journal_mode query failed: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}

	// Verify foreign keys enabled
	var fkEnabled int
	if err := pool.QueryRow("PRAGMA foreign_keys;").Scan(&fkEnabled); err != nil {
		t.Fatalf("PRAGMA foreign_keys query failed: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("foreign_keys = %d, want 1", fkEnabled)
	}
}

func TestOpenPool_PoolLimits(t *testing.T) {
	cfg := tempSQLiteConfig(t)
	pc := PoolConfig{MaxOpenConns: 7, MaxIdleConns: 3, ConnMaxLifetime: 2 * time.Minute}

	pool, err := OpenPool(cfg, pc)
	if err != nil {
		t.Fatalf("OpenPool failed: %v", err)
	}
	defer pool.Close()

	stats := pool.Stats()
	if stats.MaxOpenConnections != 7 {
		t.Errorf("MaxOpenConnections = %d, want 7", stats.MaxOpenConnections)
	}
}

func TestOpenPool_UnsupportedDriver(t *testing.T) {
	cfg := config.Config{Db_Driver: "oracle"}
	_, err := OpenPool(cfg, DefaultPluginPoolConfig())
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}

func TestOpenPool_DualPoolIsolation(t *testing.T) {
	cfg := tempSQLiteConfig(t)

	// Open "core" pool with 5 max connections
	corePc := PoolConfig{MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifetime: time.Minute}
	corePool, err := OpenPool(cfg, corePc)
	if err != nil {
		t.Fatalf("core OpenPool failed: %v", err)
	}
	defer corePool.Close()

	// Open "plugin" pool with 3 max connections
	pluginPc := PoolConfig{MaxOpenConns: 3, MaxIdleConns: 1, ConnMaxLifetime: time.Minute}
	pluginPool, err := OpenPool(cfg, pluginPc)
	if err != nil {
		t.Fatalf("plugin OpenPool failed: %v", err)
	}
	defer pluginPool.Close()

	// Create a test table via core pool
	_, err = corePool.Exec("CREATE TABLE IF NOT EXISTS pool_test (id INTEGER PRIMARY KEY, val TEXT);")
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	// Saturate the plugin pool by holding all its connections
	pluginConns := make([]*sql.Conn, 0, pluginPc.MaxOpenConns)
	for range pluginPc.MaxOpenConns {
		conn, err := pluginPool.Conn(context.Background())
		if err != nil {
			t.Fatalf("plugin pool Conn: %v", err)
		}
		pluginConns = append(pluginConns, conn)
	}

	// Core pool should still be able to serve queries
	var wg sync.WaitGroup
	errCh := make(chan error, 5)
	for i := range 5 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := corePool.Exec("INSERT INTO pool_test (val) VALUES (?);", "test")
			if err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("core pool query while plugin saturated: %v", err)
	}

	// Release plugin connections
	for _, conn := range pluginConns {
		conn.Close()
	}

	// Verify WAL sidecar files exist (confirms WAL mode is active)
	walPath := cfg.Db_URL + "-wal"
	if _, err := os.Stat(walPath); err != nil {
		t.Logf("WAL file check: %v (may not exist until write)", err)
	}
}

func TestDefaultPluginPoolConfig(t *testing.T) {
	pc := DefaultPluginPoolConfig()
	if pc.MaxOpenConns != 10 {
		t.Errorf("MaxOpenConns = %d, want 10", pc.MaxOpenConns)
	}
	if pc.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", pc.MaxIdleConns)
	}
	if pc.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("ConnMaxLifetime = %v, want 5m", pc.ConnMaxLifetime)
	}
}
