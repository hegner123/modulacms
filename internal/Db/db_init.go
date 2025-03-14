package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	config "github.com/hegner123/modulacms/internal/Config"
	utility "github.com/hegner123/modulacms/internal/Utility"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/*
var SqlFiles embed.FS

func (d Database) GetDb(verbose *bool) DbDriver {
	if *verbose {
		utility.LogHeader("Connecting to SQLite database...")
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
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		errWithContext := fmt.Errorf("failed to enable foreign keys: %w", err)
		utility.DefaultLogger.Error("Database configuration error", errWithContext)
		d.Err = errWithContext
		return d
	}

	d.Connection = db
	d.Context = ctx
	d.Err = nil
	return d
}
func (d MysqlDatabase) GetDb(verbose *bool) DbDriver {
	if *verbose {
		utility.LogHeader("Connecting to MySQL database...")
	}
	ctx := context.Background()

	// Create connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", d.Config.Db_User, d.Config.Db_Password, d.Config.Db_URL, d.Config.Db_Name)

	// Hide password in logs
	sanitizedDsn := fmt.Sprintf("%s:****@tcp(%s)/%s", d.Config.Db_User, d.Config.Db_URL, d.Config.Db_Name)
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

	utility.DefaultLogger.Info("MySQL database connected successfully", "database", d.Config.Db_Name)
	d.Connection = db
	d.Context = ctx
	d.Err = nil
	return d
}
func (d PsqlDatabase) GetDb(verbose *bool) DbDriver {
	if *verbose {
		utility.LogHeader("Connecting to PostgreSQL database...")
	}
	ctx := context.Background()

	// Create connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		d.Config.Db_User, d.Config.Db_Password, d.Config.Db_URL, d.Config.Db_Name)

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

	utility.DefaultLogger.Info("PostgreSQL database connected successfully", "database", d.Config.Db_Name)
	d.Connection = db
	d.Context = ctx
	d.Err = nil
	return d
}

func ConfigDB(env config.Config) DbDriver {
	verbose := false
	switch env.Db_Driver {
	case config.Sqlite:
		d := Database{Src: env.Db_Name, Config: env}
		dbc := d.GetDb(&verbose)
		return dbc
	case config.Mysql:
		d := MysqlDatabase{Src: env.Db_Name, Config: env}
		dbc := d.GetDb(&verbose)
		return dbc
	case config.Psql:
		d := PsqlDatabase{Src: env.Db_Name, Config: env}
		dbc := d.GetDb(&verbose)
		return dbc
	}
	return nil
}

func (d Database) InitDb(v *bool) error {
	tables, err := ReadSchemaFiles(v)
	if err != nil {
		return err
	}
	if _, err := d.Connection.ExecContext(d.Context, tables); err != nil {
		return err
	}

	return nil
}
func (d MysqlDatabase) InitDb(v *bool) error {
	tables, err := ReadSchemaFiles(v)
	if err != nil {
		return err
	}
	if _, err := d.Connection.ExecContext(d.Context, tables); err != nil {
		return err
	}

	return nil
}
func (d PsqlDatabase) InitDb(v *bool) error {
	tables, err := ReadSchemaFiles(v)
	if err != nil {
		return err
	}
	if _, err := d.Connection.ExecContext(d.Context, tables); err != nil {
		return err
	}

	return nil
}

func ReadSchemaFiles(verbose *bool) (string, error) {
	var result []string

	err := fs.WalkDir(SqlFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Base(path) == "schema.sql" {
			data, err := SqlFiles.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}
			result = append(result, string(data))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if *verbose {
		fmt.Println(strings.Join(result, "\n"))
	}
	return strings.Join(result, "\n"), nil
}

func GenerateKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}
	return key
}

func (d MysqlDatabase) CreateTables() {

}
