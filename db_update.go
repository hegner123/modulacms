package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbUpdateAdminRoute(db *sql.DB, ctx context.Context, s mdb.UpdateAdminRouteParams) string {
	queries := new(mdb.Queries)
	err := queries.UpdateAdminRoute(ctx, s)
	if err != nil {
		logError("failed to update admin route ", err)
	}
	return fmt.Sprintf("Successfully updated %v\n", s.Slug)
}

func dbUpdateField(db *sql.DB, ctx context.Context, s mdb.UpdateFieldParams) string {
	queries := new(mdb.Queries)
	err := queries.UpdateField(ctx, s)
	if err != nil {
		logError("failed to update field ", err)
	}
	return fmt.Sprintf("Successfully updated %v\n", s.Label)
}

func dbUpdateMedia(db *sql.DB, ctx context.Context, s mdb.UpdateMediaParams) string {
	queries := new(mdb.Queries)
	err := queries.UpdateMedia(ctx, s)
	if err != nil {
		logError("failed to update media ", err)
	}
	return fmt.Sprintf("Successfully updated %v\n", s.Name)
}

func dbUpdateMediaDimension(db *sql.DB, ctx context.Context, s mdb.UpdateMediaDimensionParams) string {
	queries := new(mdb.Queries)
	err := queries.UpdateMediaDimension(ctx, s)
	if err != nil {
		logError("failed to update MediaDimension ", err)
	}
	return fmt.Sprintf("Successfully updated %v\n", s.Label)
}

func dbUpdateRoute(db *sql.DB, ctx context.Context, s mdb.UpdateRouteParams) string {
	queries := new(mdb.Queries)
	err := queries.UpdateRoute(ctx, s)
	if err != nil {
		logError("failed to update route ", err)
	}
	return fmt.Sprintf("Successfully updated %v\n", s.Title)
}

func dbUpdateUser(db *sql.DB, ctx context.Context, s mdb.UpdateUserParams) string {
	queries := new(mdb.Queries)
	err := queries.UpdateUser(ctx, s)
	if err != nil {
		logError("failed to update user ", err)
	}
	return fmt.Sprintf("Successfully updated %v\n", s.Name)
}

func dbUpdateTable(db *sql.DB, ctx context.Context, s mdb.UpdateTableParams) string {
	queries := new(mdb.Queries)
	err := queries.UpdateTable(ctx, s)
	if err != nil {
		logError("failed to update table ", err)
	}
	return fmt.Sprintf("Successfully updated %v\n", s.Label)
}
