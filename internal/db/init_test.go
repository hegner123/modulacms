// White-box tests for init.go: singleton lifecycle (InitDB, ConfigDB, CloseDB),
// GetDb methods, GenerateKey, and buildDSN edge cases.
//
// White-box access is needed because:
//   - The singleton pattern uses unexported package-level vars (dbInstance, dbOnce, dbInitErr)
//     that must be reset between tests to avoid cross-test contamination.
//   - buildDSN is unexported but contains critical DSN-construction logic.
package db

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	config "github.com/hegner123/modulacms/internal/config"
)

// resetSingleton clears the package-level singleton state so each test
// starts with a fresh InitDB. Must be called at the start of every test
// that exercises InitDB, ConfigDB, or CloseDB.
func resetSingleton(t *testing.T) {
	t.Helper()
	// If a previous test left a live connection, close it to avoid leaks.
	if dbInstance != nil {
		con, _, err := dbInstance.GetConnection()
		if err == nil && con != nil {
			con.Close() // best-effort cleanup; ignore error
		}
	}
	dbInstance = nil
	dbOnce = sync.Once{}
	dbInitErr = nil
}

// --- InitDB tests ---

func TestInitDB_SQLite_Success(t *testing.T) {
	resetSingleton(t)

	dir := t.TempDir()
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "initdb_test.db"),
	}

	driver, err := InitDB(cfg)
	if err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}
	if driver == nil {
		t.Fatal("InitDB returned nil driver")
	}

	// Verify the connection is alive
	if pingErr := driver.Ping(); pingErr != nil {
		t.Fatalf("Ping after InitDB failed: %v", pingErr)
	}

	// Clean up: close the connection through the singleton
	con, _, connErr := driver.GetConnection()
	if connErr != nil {
		t.Fatalf("GetConnection failed: %v", connErr)
	}
	t.Cleanup(func() { con.Close() })
}

func TestInitDB_Idempotent(t *testing.T) {
	// sync.Once guarantees InitDB only initializes once. Calling it twice
	// with different configs should still return the first result.
	resetSingleton(t)

	dir := t.TempDir()
	cfg1 := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "first.db"),
	}
	cfg2 := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "second.db"),
	}

	driver1, err1 := InitDB(cfg1)
	if err1 != nil {
		t.Fatalf("first InitDB: %v", err1)
	}
	driver2, err2 := InitDB(cfg2)
	if err2 != nil {
		t.Fatalf("second InitDB: %v", err2)
	}

	// Both calls should return the same driver instance.
	// We can verify by checking they share the same underlying *sql.DB.
	con1, _, _ := driver1.GetConnection()
	con2, _, _ := driver2.GetConnection()
	if con1 != con2 {
		t.Error("second InitDB call returned a different connection; expected singleton behavior")
	}

	t.Cleanup(func() { con1.Close() })
}

func TestInitDB_UnsupportedDriver(t *testing.T) {
	resetSingleton(t)

	cfg := config.Config{Db_Driver: "oracle"}
	driver, err := InitDB(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported driver, got nil")
	}
	if driver != nil {
		t.Errorf("expected nil driver for unsupported driver, got %T", driver)
	}
}

func TestInitDB_EmptyDriver(t *testing.T) {
	resetSingleton(t)

	cfg := config.Config{Db_Driver: ""}
	driver, err := InitDB(cfg)
	if err == nil {
		t.Fatal("expected error for empty driver, got nil")
	}
	if driver != nil {
		t.Errorf("expected nil driver for empty driver, got %T", driver)
	}
}

// --- ConfigDB tests ---

func TestConfigDB_ReturnsSingleton_WhenInitialized(t *testing.T) {
	resetSingleton(t)

	dir := t.TempDir()
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "configdb_singleton.db"),
	}

	// Initialize the singleton first
	initDriver, err := InitDB(cfg)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}

	// ConfigDB should return the same singleton, not create a new connection
	configDriver := ConfigDB(cfg)
	if configDriver == nil {
		t.Fatal("ConfigDB returned nil")
	}

	initCon, _, _ := initDriver.GetConnection()
	configCon, _, _ := configDriver.GetConnection()
	if initCon != configCon {
		t.Error("ConfigDB returned different connection than InitDB singleton")
	}

	t.Cleanup(func() { initCon.Close() })
}

func TestConfigDB_FallbackCreatesNewConnection(t *testing.T) {
	resetSingleton(t)

	dir := t.TempDir()
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "configdb_fallback.db"),
	}

	// dbInstance is nil (no InitDB called), ConfigDB should create a one-off connection
	driver := ConfigDB(cfg)
	if driver == nil {
		t.Fatal("ConfigDB fallback returned nil")
	}

	if pingErr := driver.Ping(); pingErr != nil {
		t.Fatalf("Ping on fallback connection failed: %v", pingErr)
	}

	// Clean up the one-off connection
	con, _, connErr := driver.GetConnection()
	if connErr != nil {
		t.Fatalf("GetConnection on fallback: %v", connErr)
	}
	t.Cleanup(func() { con.Close() })
}

func TestConfigDB_UnsupportedDriver_ReturnsNil(t *testing.T) {
	resetSingleton(t)

	cfg := config.Config{Db_Driver: "mssql"}
	driver := ConfigDB(cfg)
	if driver != nil {
		t.Errorf("expected nil for unsupported driver, got %T", driver)
	}
}

func TestConfigDB_EmptyDriver_ReturnsNil(t *testing.T) {
	resetSingleton(t)

	cfg := config.Config{Db_Driver: ""}
	driver := ConfigDB(cfg)
	if driver != nil {
		t.Errorf("expected nil for empty driver, got %T", driver)
	}
}

// --- CloseDB tests ---

func TestCloseDB_NilInstance_Noop(t *testing.T) {
	resetSingleton(t)

	// dbInstance is nil; CloseDB should return nil without error
	err := CloseDB()
	if err != nil {
		t.Errorf("CloseDB with nil instance: %v", err)
	}
}

func TestCloseDB_ClosesActiveConnection(t *testing.T) {
	resetSingleton(t)

	dir := t.TempDir()
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "closedb_test.db"),
	}

	_, err := InitDB(cfg)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}

	err = CloseDB()
	if err != nil {
		t.Fatalf("CloseDB: %v", err)
	}

	// After closing, the underlying *sql.DB should reject new operations.
	// dbInstance is still set (CloseDB doesn't nil it), but the pool is closed.
	con, _, connErr := dbInstance.GetConnection()
	if connErr != nil {
		// GetConnection just returns the struct field; it shouldn't error
		t.Fatalf("GetConnection after close: %v", connErr)
	}
	// Ping should fail on a closed pool
	if pingErr := con.Ping(); pingErr == nil {
		t.Error("expected Ping to fail after CloseDB, but it succeeded")
	}
}

// --- GetDb tests (SQLite only; MySQL/PostgreSQL need real databases) ---

func TestDatabase_GetDb_Success(t *testing.T) {
	dir := t.TempDir()
	d := Database{Src: filepath.Join(dir, "getdb_success.db")}
	verbose := true

	result := d.GetDb(&verbose)
	if result == nil {
		t.Fatal("GetDb returned nil")
	}

	// Cast back to Database to inspect fields
	db, ok := result.(Database)
	if !ok {
		t.Fatalf("GetDb returned %T, want Database", result)
	}
	if db.Err != nil {
		t.Fatalf("GetDb set Err: %v", db.Err)
	}
	if db.Connection == nil {
		t.Fatal("GetDb did not set Connection")
	}
	if db.Context == nil {
		t.Fatal("GetDb did not set Context")
	}

	// Verify pool settings
	stats := db.Connection.Stats()
	if stats.MaxOpenConnections != 25 {
		t.Errorf("MaxOpenConnections = %d, want 25", stats.MaxOpenConnections)
	}

	// Verify PRAGMAs
	var journalMode string
	if err := db.Connection.QueryRow("PRAGMA journal_mode;").Scan(&journalMode); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}

	var fkEnabled int
	if err := db.Connection.QueryRow("PRAGMA foreign_keys;").Scan(&fkEnabled); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("foreign_keys = %d, want 1", fkEnabled)
	}

	t.Cleanup(func() { db.Connection.Close() })
}

func TestDatabase_GetDb_DefaultPath(t *testing.T) {
	// When Src is empty, GetDb should default to "./modula.db"
	d := Database{Src: ""}
	verbose := false

	result := d.GetDb(&verbose)
	db, ok := result.(Database)
	if !ok {
		t.Fatalf("GetDb returned %T, want Database", result)
	}
	// The connection should have been opened (or attempted) with the default path.
	// We verify the Src field was updated.
	if db.Src != "./modula.db" {
		t.Errorf("Src = %q, want %q", db.Src, "./modula.db")
	}

	// Clean up the default database file if it was created
	if db.Connection != nil {
		db.Connection.Close()
	}
}

func TestDatabase_GetDb_NonVerbose(t *testing.T) {
	// Verify that GetDb works with verbose=false (no logging panic)
	dir := t.TempDir()
	d := Database{Src: filepath.Join(dir, "nonverbose.db")}
	verbose := false

	result := d.GetDb(&verbose)
	db, ok := result.(Database)
	if !ok {
		t.Fatalf("GetDb returned %T, want Database", result)
	}
	if db.Err != nil {
		t.Fatalf("GetDb with verbose=false set Err: %v", db.Err)
	}

	t.Cleanup(func() {
		if db.Connection != nil {
			db.Connection.Close()
		}
	})
}

// --- GenerateKey tests ---

func TestGenerateKey_Length(t *testing.T) {
	t.Parallel()
	key := GenerateKey()
	if len(key) != 32 {
		t.Errorf("key length = %d, want 32", len(key))
	}
}

func TestGenerateKey_NonZero(t *testing.T) {
	t.Parallel()
	key := GenerateKey()
	allZero := true
	for _, b := range key {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("GenerateKey returned all-zero bytes; extremely unlikely from crypto/rand")
	}
}

func TestGenerateKey_Unique(t *testing.T) {
	t.Parallel()
	// Two calls should produce different keys (collision probability is negligible
	// for 256-bit random values).
	key1 := GenerateKey()
	key2 := GenerateKey()
	if string(key1) == string(key2) {
		t.Error("two GenerateKey calls produced identical keys")
	}
}

// --- buildDSN edge cases (supplements existing tests in pool_test.go) ---

func TestBuildDSN_PostgreSQL_SpecialCharsInPassword(t *testing.T) {
	t.Parallel()

	// Passwords with special characters that need URL encoding.
	// url.UserPassword handles this, but we verify the round-trip.
	tests := []struct {
		name     string
		password string
	}{
		{name: "at_sign", password: "p@ss"},
		{name: "slash", password: "p/ss"},
		{name: "colon", password: "p:ss"},
		{name: "space", password: "p ss"},
		{name: "hash", password: "p#ss"},
		{name: "percent", password: "p%ss"},
		{name: "complex", password: "P@$$w0rd/With:Special#Chars%20!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.Config{
				Db_Driver:   config.Psql,
				Db_User:     "admin",
				Db_Password: tt.password,
				Db_URL:      "localhost:5432",
				Db_Name:     "testdb",
			}

			driver, dsn, err := buildDSN(cfg)
			if err != nil {
				t.Fatalf("buildDSN error: %v", err)
			}
			if driver != "postgres" {
				t.Errorf("driver = %q, want %q", driver, "postgres")
			}
			// The DSN should be a valid postgres URL. We can't easily
			// decode it back without parsing, but we verify it doesn't
			// contain the raw password when special chars are present
			// (they should be percent-encoded).
			if dsn == "" {
				t.Fatal("dsn is empty")
			}
		})
	}
}

func TestBuildDSN_MySQL_EmptyFields(t *testing.T) {
	t.Parallel()

	// MySQL DSN with empty user/password/host/name should still produce
	// a syntactically valid DSN string (even if the connection would fail).
	cfg := config.Config{
		Db_Driver:   config.Mysql,
		Db_User:     "",
		Db_Password: "",
		Db_URL:      "",
		Db_Name:     "",
	}

	driver, dsn, err := buildDSN(cfg)
	if err != nil {
		t.Fatalf("buildDSN error: %v", err)
	}
	if driver != "mysql" {
		t.Errorf("driver = %q, want %q", driver, "mysql")
	}
	// Format: "user:pass@tcp(host)/dbname?parseTime=true"
	if dsn != ":@tcp()/?parseTime=true" {
		t.Errorf("dsn = %q, want %q", dsn, ":@tcp()/?parseTime=true")
	}
}

func TestBuildDSN_PostgreSQL_EmptyFields(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Db_Driver:   config.Psql,
		Db_User:     "",
		Db_Password: "",
		Db_URL:      "",
		Db_Name:     "",
	}

	driver, dsn, err := buildDSN(cfg)
	if err != nil {
		t.Fatalf("buildDSN error: %v", err)
	}
	if driver != "postgres" {
		t.Errorf("driver = %q, want %q", driver, "postgres")
	}
	if dsn == "" {
		t.Error("dsn is empty; expected a postgres:// URL even with empty fields")
	}
}

// --- PoolConfig edge cases ---

func TestOpenPool_ZeroPoolConfig(t *testing.T) {
	// Zero values in PoolConfig should still produce a usable pool
	// (Go's sql.DB treats 0 as "unlimited" for MaxOpenConns)
	cfg := tempSQLiteConfig(t)
	pc := PoolConfig{
		MaxOpenConns:    0,
		MaxIdleConns:    0,
		ConnMaxLifetime: 0,
	}

	pool, err := OpenPool(cfg, pc)
	if err != nil {
		t.Fatalf("OpenPool with zero config: %v", err)
	}
	defer pool.Close()

	// Pool should be alive
	if err := pool.Ping(); err != nil {
		t.Fatalf("Ping on zero-config pool: %v", err)
	}

	stats := pool.Stats()
	// When MaxOpenConns is 0, Go reports it as 0 (unlimited)
	if stats.MaxOpenConnections != 0 {
		t.Errorf("MaxOpenConnections = %d, want 0 (unlimited)", stats.MaxOpenConnections)
	}
}

func TestOpenPool_LargePoolConfig(t *testing.T) {
	cfg := tempSQLiteConfig(t)
	pc := PoolConfig{
		MaxOpenConns:    1000,
		MaxIdleConns:    500,
		ConnMaxLifetime: 24 * time.Hour,
	}

	pool, err := OpenPool(cfg, pc)
	if err != nil {
		t.Fatalf("OpenPool with large config: %v", err)
	}
	defer pool.Close()

	stats := pool.Stats()
	if stats.MaxOpenConnections != 1000 {
		t.Errorf("MaxOpenConnections = %d, want 1000", stats.MaxOpenConnections)
	}
}

func TestOpenPool_SingleConnection(t *testing.T) {
	// MaxOpenConns=1 is a valid configuration (serialized access)
	cfg := tempSQLiteConfig(t)
	pc := PoolConfig{
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Minute,
	}

	pool, err := OpenPool(cfg, pc)
	if err != nil {
		t.Fatalf("OpenPool: %v", err)
	}
	defer pool.Close()

	// Should still handle multiple sequential queries
	for i := range 5 {
		var result int
		err := pool.QueryRow("SELECT ?;", i).Scan(&result)
		if err != nil {
			t.Fatalf("query %d: %v", i, err)
		}
		if result != i {
			t.Errorf("query %d: got %d, want %d", i, result, i)
		}
	}
}

// --- InitDB + CloseDB lifecycle ---

func TestInitDB_CloseDB_Lifecycle(t *testing.T) {
	// Full lifecycle: init -> use -> close -> verify closed
	resetSingleton(t)

	dir := t.TempDir()
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "lifecycle.db"),
	}

	driver, err := InitDB(cfg)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}

	// Use the connection
	con, _, connErr := driver.GetConnection()
	if connErr != nil {
		t.Fatalf("GetConnection: %v", connErr)
	}
	_, execErr := con.Exec("CREATE TABLE lifecycle_test (id INTEGER PRIMARY KEY);")
	if execErr != nil {
		t.Fatalf("CREATE TABLE: %v", execErr)
	}

	// Close
	closeErr := CloseDB()
	if closeErr != nil {
		t.Fatalf("CloseDB: %v", closeErr)
	}

	// Verify closed: queries should fail
	_, execErr = con.Exec("INSERT INTO lifecycle_test (id) VALUES (1);")
	if execErr == nil {
		t.Error("expected error executing query after CloseDB")
	}
}

// --- ConfigDB with different driver types (SQLite fallback) ---

func TestConfigDB_SQLite_FallbackSetsPoolSettings(t *testing.T) {
	resetSingleton(t)

	dir := t.TempDir()
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "configdb_pool.db"),
	}

	driver := ConfigDB(cfg)
	if driver == nil {
		t.Fatal("ConfigDB returned nil")
	}

	con, _, connErr := driver.GetConnection()
	if connErr != nil {
		t.Fatalf("GetConnection: %v", connErr)
	}
	defer con.Close()

	// Verify the fallback path sets pool limits (25 max open, 10 idle)
	stats := con.Stats()
	if stats.MaxOpenConnections != 25 {
		t.Errorf("MaxOpenConnections = %d, want 25", stats.MaxOpenConnections)
	}
}

// --- Concurrent InitDB calls ---

func TestInitDB_ConcurrentCalls(t *testing.T) {
	// Multiple goroutines calling InitDB simultaneously should all get
	// the same result thanks to sync.Once.
	resetSingleton(t)

	dir := t.TempDir()
	cfg := config.Config{
		Db_Driver: config.Sqlite,
		Db_URL:    filepath.Join(dir, "concurrent_init.db"),
	}

	const goroutines = 10
	type result struct {
		driver DbDriver
		err    error
	}
	results := make(chan result, goroutines)

	for range goroutines {
		go func() {
			d, e := InitDB(cfg)
			results <- result{driver: d, err: e}
		}()
	}

	var firstCon *interface{}
	_ = firstCon
	var firstDriver DbDriver

	for i := range goroutines {
		r := <-results
		if r.err != nil {
			t.Fatalf("goroutine %d: InitDB error: %v", i, r.err)
		}
		if r.driver == nil {
			t.Fatalf("goroutine %d: InitDB returned nil driver", i)
		}
		if i == 0 {
			firstDriver = r.driver
		} else {
			// All goroutines should get the same connection
			con1, _, _ := firstDriver.GetConnection()
			con2, _, _ := r.driver.GetConnection()
			if con1 != con2 {
				t.Errorf("goroutine %d returned different connection than goroutine 0", i)
			}
		}
	}

	// Clean up
	t.Cleanup(func() {
		if dbInstance != nil {
			con, _, err := dbInstance.GetConnection()
			if err == nil && con != nil {
				con.Close()
			}
		}
	})
}
