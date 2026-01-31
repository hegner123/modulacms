package db

import (
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
)

// DropAllTables drops all database tables in reverse dependency order (SQLite)
func (d Database) DropAllTables() error {
	queries := mdb.New(d.Connection)
	return queries.DropAllTables(d.Context)
}

// DropAllTables drops all database tables in reverse dependency order (MySQL)
func (d MysqlDatabase) DropAllTables() error {
	queries := mdbm.New(d.Connection)
	return queries.DropAllTables(d.Context)
}

// DropAllTables drops all database tables in reverse dependency order (PostgreSQL)
func (d PsqlDatabase) DropAllTables() error {
	queries := mdbp.New(d.Connection)
	return queries.DropAllTables(d.Context)
}
