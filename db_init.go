package main

import (
	"context"
	"database/sql"
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	mdb "github.com/hegner123/modulacms/db-sqlite"
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
	return strings.Join(result, "\n"), nil
}

func createSetupInserts(db *sql.DB, ctx context.Context,modify string) {

	times := timestampS()
	dbCreateAdminRoute(db, ctx, mdb.CreateAdminRouteParams{
		Author:       ns("system"),
		Authorid:     ns("0"),
		Slug:         ns("/test1"+modify),
		Title:        ns("Test"),
		Status:       ni(0),
		Datecreated:  ns(times),
		Datemodified: ns(times),
		Content:      ns("Test content"),
		Template:     ns("page.html"),
	})
	dbCreateRoute(db, ctx, mdb.CreateRouteParams{
		Author:       ns("system"),
		Authorid:     ns("0"),
		Slug:         ns("/test1"+modify),
		Title:        ns("Test"),
		Status:       ni(0),
		Datecreated:  ns(times),
		Datemodified: ns(times),
		Content:      ns("Test content"),
		Template:     ns("page.html"),
	})
	dbCreateMedia(db, ctx, mdb.CreateMediaParams{
		Name:               ns("test.png"),
		Displayname:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             ns("system"),
		Authorid:           ni(0),
		Datecreated:        ns(times),
		Datemodified:       ns(times),
		Url:                ns("public/2024/11/test1.png"+modify),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		Optimizedmobile:    ns("public/2024/11/test-mobile.png"),
		Optimizedtablet:    ns("public/2024/11/test-tablet.png"),
		Optimizeddesktop:   ns("public/2024/11/test-desktop.png"),
		Optimizedultrawide: ns("public/2024/11/test-ultra.png"),
	})

	dbCreateField(db, ctx, mdb.CreateFieldParams{
		Routeid:      int64(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       ns("system"),
		Authorid:     ns("0"),
		Datecreated:  ns(times),
		Datemodified: ns(times),
	})
	dbCreateUser(db, ctx, mdb.CreateUserParams{
		Datecreated:  ns(times),
		Datemodified: ns(times),
		Username:     ns("system"),
		Name:         ns("system"),
		Email:        ns("system@modulacms.com"+modify),
		Hash:         ns("has"),
		Role:         ns("admin"),
	})
	dbCreateMediaDimension(db, ctx, mdb.CreateMediaDimensionParams{
		Label:  ns("Tablet"+modify),
		Width:  ni(1920),
		Height: ni(1080),
	})

	dbCreateTable(db, ctx, mdb.Tables{Label: ns("Test1"+modify)})
}
