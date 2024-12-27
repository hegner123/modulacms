package db

import (
	"context"
	"database/sql"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func CountAdminRoutes(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdb.New(db)
	c, err := queries.CountAdminroute(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func CountDatatypes(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdb.New(db)
	c, err := queries.CountDatatype(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func CountField(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdb.New(db)
	c, err := queries.CountField(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func CountMedia(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdb.New(db)
	c, err := queries.CountMedia(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func CountTables(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdb.New(db)
	c, err := queries.CountTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func CountTokens(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdb.New(db)
	c, err := queries.CountTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func CountUsers(db *sql.DB, ctx context.Context) (*int64, error) {
	queries := mdb.New(db)
	c, err := queries.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
