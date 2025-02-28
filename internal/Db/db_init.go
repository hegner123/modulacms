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

	config "github.com/hegner123/modulacms/internal/Config"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/*
var SqlFiles embed.FS

func (dbSrc Database) GetDb() TestDbDriver {
	ctx := context.Background()

	if dbSrc.Src == "" {
		dbSrc.Src = "./modula.db"
	}
	db, err := sql.Open("sqlite3", dbSrc.Src)
	if err != nil {
		fmt.Printf("db exec err db_init 007 : %s\n", err)
        dbSrc.Err = err
		return dbSrc
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s\n", err)
        dbSrc.Err = err
		return dbSrc
	}
    dbSrc.Connection = db
    dbSrc.Context = ctx
    dbSrc.Err = nil
	return dbSrc
}
func (dbSrc MysqlDatabase) GetDb() TestDbDriver {
	ctx := context.Background()

	if dbSrc.Src == "" {
		dbSrc.Src = "./modula.db"
	}
	db, err := sql.Open("sqlite3", dbSrc.Src)
	if err != nil {
		fmt.Printf("db exec err db_init 007 : %s\n", err)
        dbSrc.Err = err
		return dbSrc
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s\n", err)
        dbSrc.Err = err
		return dbSrc
	}
    dbSrc.Connection = db
    dbSrc.Context = ctx
    dbSrc.Err = nil
	return dbSrc
}
func (dbSrc PsqlDatabase) GetDb() TestDbDriver {
	ctx := context.Background()

	if dbSrc.Src == "" {
		dbSrc.Src = "./modula.db"
	}
	db, err := sql.Open("sqlite3", dbSrc.Src)
	if err != nil {
		fmt.Printf("db exec err db_init 007 : %s\n", err)
        dbSrc.Err = err
		return dbSrc
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s\n", err)
        dbSrc.Err = err
		return dbSrc
	}
    dbSrc.Connection = db
    dbSrc.Context = ctx
    dbSrc.Err = nil
	return dbSrc
}

func DbConfig(env config.Config) TestDbDriver {
	switch env.Db_Driver {
	case config.Sqlite:
		d := Database{Src: env.Db_Name}
		dbc := d.GetDb()
		return dbc
	case config.Mysql:
		d := MysqlDatabase{Src: env.Db_Name}
		dbc := d.GetDb()
		return dbc
	case config.Psql:
		d := PsqlDatabase{Src: env.Db_Name}
		dbc := d.GetDb()
		return dbc
	}
	return nil
}

func (init Database) InitDb(Db Database, v *bool, database string) error {
	tables, err := ReadSchemaFiles(v)
	if err != nil {
		return err
	}
	if _, err := Db.Connection.ExecContext(Db.Context, tables); err != nil {
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
