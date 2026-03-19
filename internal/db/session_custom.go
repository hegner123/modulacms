package db

import (
	"database/sql"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

// GetSessionByToken retrieves a session by its session_data token value. (SQLite)
func (d Database) GetSessionByToken(token string) (*Sessions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetSessionByToken(d.Context, mdb.GetSessionByTokenParams{
		SessionData: sql.NullString{String: token, Valid: token != ""},
	})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

// MYSQL

// GetSessionByToken retrieves a session by its session_data token value. (MySQL)
func (d MysqlDatabase) GetSessionByToken(token string) (*Sessions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetSessionByToken(d.Context, mdbm.GetSessionByTokenParams{
		SessionData: sql.NullString{String: token, Valid: token != ""},
	})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

// PSQL

// GetSessionByToken retrieves a session by its session_data token value. (PostgreSQL)
func (d PsqlDatabase) GetSessionByToken(token string) (*Sessions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetSessionByToken(d.Context, mdbp.GetSessionByTokenParams{
		SessionData: sql.NullString{String: token, Valid: token != ""},
	})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}
