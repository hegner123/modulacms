package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type DbStatus string

const (
	open   DbStatus = "open"
	closed DbStatus = "closed"
	err    DbStatus = "error"
)

type Database struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Context        context.Context
}

func (db Database) GetDb(s string) (*sql.DB, context.Context, error) {
	if db.status == open {
		return &sql.DB{}, context.TODO(), errors.New("db_is already open")
	}
	db.context = context.Background()

	if s == "" {
		db.src = "./modula.db"
	}
	db.connection, db.err = sql.Open("sqlite3", db.src)
	if db.err != nil {
		fmt.Printf("db exec err db_init 007 : %s\n", err)
	}
	v, err := db.connection.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s\n", err)
	}
	fmt.Println(v)
	return db.connection, db.context, nil
}
