package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// CreateSystemUser inserts the protected system user with the well-known SystemUserID.
// This bypasses the audited command pattern because bootstrap runs before any users exist.
func (d Database) CreateSystemUser(params CreateUserParams) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.CreateUser(d.Context, mdb.CreateUserParams{
		UserID:       types.SystemUserID,
		Username:     params.Username,
		Name:         params.Name,
		Email:        params.Email,
		Hash:         params.Hash,
		Roles:        params.Role,
		DateCreated:  params.DateCreated,
		DateModified: params.DateModified,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create system user: %w", err)
	}
	r := d.MapUser(row)
	return &r, nil
}

func (d MysqlDatabase) CreateSystemUser(params CreateUserParams) (*Users, error) {
	queries := mdbm.New(d.Connection)
	p := mdbm.CreateUserParams{
		UserID:       types.SystemUserID,
		Username:     params.Username,
		Name:         params.Name,
		Email:        params.Email,
		Hash:         params.Hash,
		Roles:        params.Role,
		DateCreated:  params.DateCreated,
		DateModified: params.DateModified,
	}
	if err := queries.CreateUser(d.Context, p); err != nil {
		return nil, fmt.Errorf("failed to create system user: %w", err)
	}
	row, err := queries.GetUser(d.Context, mdbm.GetUserParams{UserID: p.UserID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch system user after create: %w", err)
	}
	r := d.MapUser(row)
	return &r, nil
}

func (d PsqlDatabase) CreateSystemUser(params CreateUserParams) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateUser(d.Context, mdbp.CreateUserParams{
		UserID:       types.SystemUserID,
		Username:     params.Username,
		Name:         params.Name,
		Email:        params.Email,
		Hash:         params.Hash,
		Roles:        params.Role,
		DateCreated:  params.DateCreated,
		DateModified: params.DateModified,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create system user: %w", err)
	}
	r := d.MapUser(row)
	return &r, nil
}
