package db

import (
	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// GetUserOauth retrieves a user OAuth record by ID (SQLite).
func (d Database) GetUserOauth(id types.UserOauthID) (*UserOauth, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, mdb.GetUserOauthParams{UserOAuthID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

// MYSQL

// GetUserOauth retrieves a user OAuth record by ID (MySQL).
func (d MysqlDatabase) GetUserOauth(id types.UserOauthID) (*UserOauth, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, mdbm.GetUserOauthParams{UserOAuthID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

// PSQL

// GetUserOauth retrieves a user OAuth record by ID (PostgreSQL).
func (d PsqlDatabase) GetUserOauth(id types.UserOauthID) (*UserOauth, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, mdbp.GetUserOauthParams{UserOAuthID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}
