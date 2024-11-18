package main

import (
	"database/sql"
	"embed"
	_ "embed"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/insert/*.sql
var sqlFiles embed.FS

//go:embed sql/init/init_fields.sql
var fieldsTable string

//go:embed sql/init/init_insert_fields.sql
var insertFields string

//go:embed sql/init/init_elements.sql
var elementsTable string

//go:embed sql/init/init_elements.sql
var insertElements string

//go:embed sql/init/init_attributes.sql
var attributesTable string

//go:embed sql/init/init_insert_attributes.sql
var insertAttribute string

//go:embed sql/init/init_adminroutes.sql
var adminRoutesTable string

//go:embed sql/init/init_insert_adminroutes.sql
var insertAdminRoutes string

//go:embed sql/init/init_routes.sql
var routesTable string

//go:embed sql/init/init_users.sql
var usersTable string

//go:embed sql/init/init_insert_users.sql
var insertSystemUser string

//go:embed sql/init/init_media.sql
var mediaTable string

//go:embed sql/init/init_insert_media.sql
var insertMedia string

//go:embed sql/init/init_md.sql
var mediaDimensionTable string

//go:embed sql/init/init_insert_md.sql
var insertMediaDimensions string

//go:embed sql/init/init_insert_tables.sql
var insertDefaultTables string

func getDb(dbName Database) (*sql.DB, error) {
	if dbName.DB == "" {
		dbName.DB = "./modula.db"
	}
	db, err := sql.Open("sqlite3", dbName.DB)
	if err != nil {
		fmt.Printf("db exec err db_init 007 : %s\n", err)
		return nil, err
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s\n", err)
		return nil, err
	}
	return db, nil
}

func initializeDatabase(db *sql.DB, reset bool) error {
	if reset {
        resetDatabase(db)
	}

	tables := []string{tables, insertDefaultTables, usersTable, adminRoutesTable, routesTable, fieldsTable, mediaTable, mediaDimensionTable, elementsTable, attributesTable}
	rows := []string{insertAdminRoutes, insertElements, insertFields, insertAttribute, insertSystemUser, insertMedia, insertMediaDimensions}

	err := forEachStatement(db, tables, "tables")
	if err != nil {
		fmt.Printf("db exec err db_init 001 : %s\n", err)
	}
	err = forEachStatement(db, rows, "rows")
	if err != nil {
		fmt.Printf("db exec err db_init 002  : %s\n", err)
	}
	return err
}

func initializeClientDatabase(clientDB string, clientReset bool) (*sql.DB, error) {
	dbName := RemoveTLD(clientDB)
	db, err := sql.Open("sqlite3", "./"+dbName+".db")
	if err != nil {
		fmt.Printf("db exec err db_init 004 : %s\n", err)
		return nil, err
	}
	res, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 005 : %s\n", err)
		return nil, err
	}
	fmt.Print(res)
	return db, nil
}

func resetDatabase(db *sql.DB) {

	res, err := db.Exec(`
            DROP TABLE IF EXISTS users;
            DROP TABLE IF EXISTS routes;
            DROP TABLE IF EXISTS adminroutes;
            DROP TABLE IF EXISTS fields;
            DROP TABLE IF EXISTS media;
            DROP TABLE IF EXISTS media_dimensions;
            DROP TABLE IF EXISTS tables;
            DROP TABLE IF EXISTS elements;
            DROP TABLE IF EXISTS attributes;`)
	if err != nil {
		fmt.Printf("db exec err db_init 006 : %s\n", err)
		log.Fatal("I CAN'T FIND THE DATABASE CAPTIN!!!!\n Oh GOD IT'S GOT MY LEG!!!!")
	}
	if res != nil {
		log.Print(res)
	}
}


