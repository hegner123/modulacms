package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func UpdateAdminDatatype(db *sql.DB, ctx context.Context, s mdb.UpdateAdminDatatypeParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateAdminDatatype(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin datatype, %v ", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func UpdateAdminField(db *sql.DB, ctx context.Context, s mdb.UpdateAdminFieldParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateAdminField(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func UpdateAdminRoute(db *sql.DB, ctx context.Context, s mdb.UpdateAdminRouteParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateAdminRoute(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}

func UpdateDatatype(db *sql.DB, ctx context.Context, s mdb.UpdateDatatypeParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateDatatype(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func UpdateField(db *sql.DB, ctx context.Context, s mdb.UpdateFieldParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateField(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func UpdateMedia(db *sql.DB, ctx context.Context, s mdb.UpdateMediaParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateMedia(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update media, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Name)
	return &u, nil
}

func UpdateMediaDimension(db *sql.DB, ctx context.Context, s mdb.UpdateMediaDimensionParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateMediaDimension(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update media dimension, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func UpdateRoute(db *sql.DB, ctx context.Context, s mdb.UpdateRouteParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateRoute(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}

func UpdateTable(db *sql.DB, ctx context.Context, s mdb.UpdateTableParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateTable(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update table, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func UpdateToken(db *sql.DB, ctx context.Context, s mdb.UpdateTokenParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateToken(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update token, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

func UpdateUser(db *sql.DB, ctx context.Context, s mdb.UpdateUserParams) (*string, error) {
	queries := mdb.New(db)
	err := queries.UpdateUser(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update user, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Name)
	return &u, nil
}
