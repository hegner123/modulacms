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
	if db.Status == open {
		return &sql.DB{}, context.TODO(), errors.New("db_is already open")
	}
	db.Context = context.Background()

	if s == "" {
		db.Src = "./modula.db"
	}
	db.Connection, db.Err = sql.Open("sqlite3", db.Src)
	if db.Err != nil {
		fmt.Printf("db exec err db_init 007 : %s\n", err)
	}
	v, err := db.Connection.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("db exec err db_init 008 : %s\n", err)
	}
	fmt.Println(v)
	return db.Connection, db.Context, nil
}
