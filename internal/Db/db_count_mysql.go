package db

import (
	"context"
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/db-mysql"
)

func (d MysqlDatabase) CountAdminRoutes(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdbm.New(db)
	c, err := queries.CountAdminroute(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CountDatatypes(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdbm.New(db)
	c, err := queries.CountDatatype(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CountField(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdbm.New(db)
	c, err := queries.CountField(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CountMedia(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdbm.New(db)
	c, err := queries.CountMedia(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CountTables(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdbm.New(db)
	c, err := queries.CountTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CountTokens(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdbm.New(db)
	c, err := queries.CountTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CountUsers(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdbm.New(db)
	c, err := queries.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
