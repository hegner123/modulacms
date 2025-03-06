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
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/*
var SqlFiles embed.FS

func (d Database) GetDb() DbDriver {
	fmt.Printf("Connecting to sqlite3 Db.......")
	ctx := context.Background()

	if d.Src == "" {
		d.Src = "./modula.db"
	}

	db, err := sql.Open("sqlite3", d.Src)
	if err != nil {
		fmt.Printf("ERROR\ndb exec err db_init 007 : %s\n", err)
		d.Err = err
		return d
	}

	fmt.Printf("OK\n")

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s\n", err)
		d.Err = err
		return d
	}

	d.Connection = db
	d.Context = ctx
	d.Err = nil
	return d
}
func (d MysqlDatabase) GetDb() DbDriver {
	fmt.Printf("Connecting to mysql Db......")
	ctx := context.Background()

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", d.Config.Db_User, d.Config.Db_Password, d.Config.Db_URL, d.Config.Db_Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("ERROR\nError opening database: %v", err)
	}

	fmt.Printf("OK\n")

	fmt.Printf("Connection:%v\n", dsn)
	d.Connection = db
	d.Context = ctx
	d.Err = nil
	return d
}
func (d PsqlDatabase) GetDb() DbDriver {
	fmt.Printf("Connecting to postgres Db.......")
	ctx := context.Background()

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", d.Config.Db_User, d.Config.Db_Password, d.Config.Db_URL, d.Config.Db_Name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("ERROR\nError opening database: %v", err)
	}
	fmt.Printf("OK\n")

	fmt.Printf("Connection: %v\n", connStr)
	d.Connection = db
	d.Context = ctx
	d.Err = nil
	return d
}

func ConfigDB(env config.Config) DbDriver {
	switch env.Db_Driver {
	case config.Sqlite:
		d := Database{Src: env.Db_Name, Config: env}
		dbc := d.GetDb()
		return dbc
	case config.Mysql:
		d := MysqlDatabase{Src: env.Db_Name, Config: env}
		dbc := d.GetDb()
		return dbc
	case config.Psql:
		d := PsqlDatabase{Src: env.Db_Name, Config: env}
		dbc := d.GetDb()
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
