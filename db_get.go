package main

import (
	"context"
	"database/sql"
	_ "embed"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbGetAdminRoute(db *sql.DB, ctx context.Context, id int) (x,error){}

func dbGetRoute(db *sql.DB, ctx context.Context, id int) (x,error){}

func dbGetUser(db *sql.DB, ctx context.Context, id int) (x,error){}

func dbGetMedia(db *sql.DB, ctx context.Context, id int) (x,error){}

func dbGetMediaDimension(db *sql.DB, ctx context.Context, id int) (x,error){}

func dbGetTable(db *sql.DB, ctx context.Context, id int) (x,error){}

func dbGetField(db *sql.DB, ctx context.Context, id int) (x,error){}
