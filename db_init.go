package main

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

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/*
var sqlFiles embed.FS

func getDb(dbName Database) (*sql.DB, context.Context, error) {
	ctx := context.Background()

	if dbName.DB == "" {
		dbName.DB = "./modula.db"
	}
	db, err := sql.Open("sqlite3", dbName.DB)
	if err != nil {
		fmt.Printf("db exec err db_init 007 : %s\n", err)
		return nil, ctx, err
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s\n", err)
		return nil, ctx, err
	}
	return db, ctx, nil
}

func initDb(db *sql.DB, ctx context.Context) error {
	tables, err := readSchemaFiles()
	if err != nil {
		logError("couldn't read schema files.", err)
	}

	if _, err := db.ExecContext(ctx, tables); err != nil {
		return err
	}

	return nil
}

func readSchemaFiles() (string, error) {
	var result []string

	// Walk through the embedded file system
	err := fs.WalkDir(sqlFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Handle traversal errors
		}
		// Match only files named "schema.sql"
		if !d.IsDir() && filepath.Base(path) == "schema.sql" {
			data, err := sqlFiles.ReadFile(path)
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
	// Join all the file contents
    fmt.Println(strings.Join(result, "\n"))
	return strings.Join(result, "\n"), nil
}

func generateKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}
	return key
}
