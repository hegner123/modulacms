package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"net/url"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	config "github.com/hegner123/modulacms/internal/config"
	utility "github.com/hegner123/modulacms/internal/utility"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// dbInstance, dbOnce, and dbInitErr manage the singleton database connection pool.
var (
	dbInstance DbDriver
	dbOnce    sync.Once
	dbInitErr error
)


// GetDb initializes a SQLite database connection and configures connection pooling.
func (d Database) GetDb(verbose *bool) DbDriver {
	if *verbose {
		utility.DefaultLogger.Info("Connecting to SQLite database...")
	}
	ctx := context.Background()

	// Use default path if not specified
	if d.Src == "" {
		d.Src = "./modula.db"
		utility.DefaultLogger.Info("Using default database path", "path", d.Src)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", d.Src)
	if err != nil {
		errWithContext := fmt.Errorf("failed to open SQLite database: %w", err)
		utility.DefaultLogger.Error("Database connection error", errWithContext, "path", d.Src)
		d.Err = errWithContext
		return d
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys=1;")
	if err != nil {
		errWithContext := fmt.Errorf("failed to enable foreign keys: %w", err)
		utility.DefaultLogger.Error("Database configuration error", errWithContext)
		d.Err = errWithContext
		return d
	}

	// Enable WAL mode for concurrent reader support
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		errWithContext := fmt.Errorf("failed to enable WAL mode: %w", err)
		utility.DefaultLogger.Error("Database configuration error", errWithContext)
		d.Err = errWithContext
		return d
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	d.Connection = db
	d.Context = ctx
	d.Err = nil
	return d
}
// GetDb initializes a MySQL database connection and configures connection pooling.
func (d MysqlDatabase) GetDb(verbose *bool) DbDriver {
	if *verbose {
		utility.DefaultLogger.Info("Connecting to MySQL database...")
	}
	ctx := context.Background()

	// Create connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", d.Config.Db_User, d.Config.Db_Password, d.Config.Db_URL, d.Config.Db_Name)

	// Hide password in logs
	sanitizedDsn := fmt.Sprintf("%s:****@tcp(%s)/%s?parseTime=true", d.Config.Db_User, d.Config.Db_URL, d.Config.Db_Name)
	utility.DefaultLogger.Info("Preparing MySQL connection", "dsn", sanitizedDsn)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		errWithContext := fmt.Errorf("failed to open MySQL database: %w", err)
		utility.DefaultLogger.Error("Database connection error", errWithContext, "host", d.Config.Db_URL)
		d.Err = errWithContext
		return d
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		errWithContext := fmt.Errorf("failed to connect to MySQL database: %w", err)
		utility.DefaultLogger.Error("Database ping error", errWithContext, "host", d.Config.Db_URL)
		d.Err = errWithContext
		return d
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	utility.DefaultLogger.Info("MySQL database connected successfully", "database", d.Config.Db_Name)
	d.Connection = db
	d.Context = ctx
	d.Err = nil
	return d
}
// GetDb initializes a PostgreSQL database connection and configures connection pooling.
func (d PsqlDatabase) GetDb(verbose *bool) DbDriver {
	if *verbose {
		utility.DefaultLogger.Info("Connecting to PostgreSQL database...")
	}
	ctx := context.Background()

	// Create connection string (url.UserPassword escapes special chars in credentials)
	connURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.Config.Db_User, d.Config.Db_Password),
		Host:     d.Config.Db_URL,
		Path:     d.Config.Db_Name,
		RawQuery: "sslmode=disable",
	}
	connStr := connURL.String()

	// Hide password in logs
	sanitizedConnStr := fmt.Sprintf("postgres://%s:****@%s/%s?sslmode=disable",
		d.Config.Db_User, d.Config.Db_URL, d.Config.Db_Name)
	utility.DefaultLogger.Info("Preparing PostgreSQL connection", "connection", sanitizedConnStr)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		errWithContext := fmt.Errorf("failed to open PostgreSQL database: %w", err)
		utility.DefaultLogger.Error("Database connection error", errWithContext, "host", d.Config.Db_URL)
		d.Err = errWithContext
		return d
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		errWithContext := fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
		utility.DefaultLogger.Error("Database ping error", errWithContext, "host", d.Config.Db_URL)
		d.Err = errWithContext
		return d
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	utility.DefaultLogger.Info("PostgreSQL database connected successfully", "database", d.Config.Db_Name)
	d.Connection = db
	d.Context = ctx
	d.Err = nil
	return d
}

// InitDB initializes the singleton database connection pool.
// Call once at application startup before any handlers execute.
func InitDB(env config.Config) (DbDriver, error) {
	dbOnce.Do(func() {
		verbose := true
		switch env.Db_Driver {
		case config.Sqlite:
			d := Database{Src: env.Db_URL, Config: env}
			dbInstance = d.GetDb(&verbose)
		case config.Mysql:
			d := MysqlDatabase{Src: env.Db_Name, Config: env}
			dbInstance = d.GetDb(&verbose)
		case config.Psql:
			d := PsqlDatabase{Src: env.Db_Name, Config: env}
			dbInstance = d.GetDb(&verbose)
		default:
			dbInitErr = fmt.Errorf("unsupported database driver: %s", env.Db_Driver)
			return
		}

		if dbInstance == nil {
			dbInitErr = fmt.Errorf("failed to initialize database: nil driver returned")
			return
		}

		// Verify the connection is alive
		if err := dbInstance.Ping(); err != nil {
			dbInitErr = fmt.Errorf("database ping failed after init: %w", err)
			dbInstance = nil
			return
		}

		utility.DefaultLogger.Info("Database pool initialized", "driver", string(env.Db_Driver))
	})

	return dbInstance, dbInitErr
}

// ConfigDB returns the singleton database driver.
// If InitDB has not been called, it falls back to creating a new connection
// (backward-compatible for CLI/install paths that run before the server starts).
func ConfigDB(env config.Config) DbDriver {
	if dbInstance != nil {
		return dbInstance
	}

	// Fallback: create a one-off connection for CLI/install usage
	verbose := false
	switch env.Db_Driver {
	case config.Sqlite:
		d := Database{Src: env.Db_URL, Config: env}
		return d.GetDb(&verbose)
	case config.Mysql:
		d := MysqlDatabase{Src: env.Db_Name, Config: env}
		return d.GetDb(&verbose)
	case config.Psql:
		d := PsqlDatabase{Src: env.Db_Name, Config: env}
		return d.GetDb(&verbose)
	}
	return nil
}

// CloseDB closes the singleton database connection pool.
// Call during graceful shutdown.
func CloseDB() error {
	if dbInstance == nil {
		return nil
	}

	con, _, err := dbInstance.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection for close: %w", err)
	}

	utility.DefaultLogger.Info("Closing database connection pool")
	return con.Close()
}



// buildDSN returns the Go sql driver name and data source name for the
// configured database driver without opening a connection.
func buildDSN(cfg config.Config) (driverName string, dsn string, err error) {
	switch cfg.Db_Driver {
	case config.Sqlite:
		src := cfg.Db_URL
		if src == "" {
			src = "./modula.db"
		}
		return "sqlite3", src, nil

	case config.Mysql:
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", cfg.Db_User, cfg.Db_Password, cfg.Db_URL, cfg.Db_Name)
		return "mysql", dsn, nil

	case config.Psql:
		connURL := url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(cfg.Db_User, cfg.Db_Password),
			Host:     cfg.Db_URL,
			Path:     cfg.Db_Name,
			RawQuery: "sslmode=disable",
		}
		return "postgres", connURL.String(), nil

	default:
		return "", "", fmt.Errorf("unsupported database driver: %q", string(cfg.Db_Driver))
	}
}

// PoolConfig holds connection pool tuning parameters.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultPluginPoolConfig returns conservative pool limits suitable for
// plugin workloads that should not starve the core CMS connection pool.
func DefaultPluginPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// OpenPool opens a new *sql.DB with independent pool limits.
// The returned pool is NOT a singleton — the caller owns it and must close it.
func OpenPool(cfg config.Config, pc PoolConfig) (*sql.DB, error) {
	driverName, dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("buildDSN: %w", err)
	}

	pool, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open(%s): %w", driverName, err)
	}

	pool.SetMaxOpenConns(pc.MaxOpenConns)
	pool.SetMaxIdleConns(pc.MaxIdleConns)
	pool.SetConnMaxLifetime(pc.ConnMaxLifetime)

	switch cfg.Db_Driver {
	case config.Sqlite:
		if _, err := pool.Exec("PRAGMA foreign_keys=1;"); err != nil {
			pool.Close() //nolint:errcheck // best-effort on setup failure
			return nil, fmt.Errorf("PRAGMA foreign_keys: %w", err)
		}
		if _, err := pool.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			pool.Close() //nolint:errcheck // best-effort on setup failure
			return nil, fmt.Errorf("PRAGMA journal_mode: %w", err)
		}

	case config.Mysql, config.Psql:
		if err := pool.Ping(); err != nil {
			pool.Close() //nolint:errcheck // best-effort on setup failure
			return nil, fmt.Errorf("ping %s: %w", driverName, err)
		}
	}

	return pool, nil
}

// GenerateKey generates a cryptographically random 32-byte key.
func GenerateKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		utility.DefaultLogger.Fatal("Failed to generate key: %v", err)
	}
	return key
}
