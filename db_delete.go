package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbDeleteAdminRoute(db *sql.DB, ctx context.Context, slug string) string {
	queries := mdb.New(db)
	err := queries.DeleteAdminRoute(ctx, ns(slug))
	if err != nil {
		logError("failed to delete admin route ", err)
	}
	return fmt.Sprintf("Deleted Admin Route %s successfully", slug)
}

func dbDeleteRoute(db *sql.DB, ctx context.Context, slug string) string {
	queries := mdb.New(db)
	err := queries.DeleteRoute(ctx, ns(slug))
	if err != nil {
		logError("failed to delete Route ", err)
	}
	return fmt.Sprintf("Deleted Route %s successfully", slug)
}

func dbDeleteUser(db *sql.DB, ctx context.Context, id int) string {
	queries := mdb.New(db)
	err := queries.DeleteUser(ctx, int64(id))
	if err != nil {
		logError("failed to delete User ", err)
	}
	return fmt.Sprintf("Deleted User %d successfully", id)
}

func dbDeleteMedia(db *sql.DB, ctx context.Context, id int) string {
	queries := mdb.New(db)
	err := queries.DeleteMedia(ctx, int64(id))
	if err != nil {
		logError("failed to delete Media ", err)
	}
	return fmt.Sprintf("Deleted Media %d successfully", id)
}

func dbDeleteMediaDimension(db *sql.DB, ctx context.Context, id int) string {
	queries := mdb.New(db)
	err := queries.DeleteMediaDimension(ctx, int64(id))
	if err != nil {
		logError("failed to delete MediaDimension ", err)
	}
	return fmt.Sprintf("Deleted Media Dimension %d successfully", id)

}

func dbDeleteTable(db *sql.DB, ctx context.Context, id int) string {
	queries := mdb.New(db)
	err := queries.DeleteTable(ctx, int64(id))
	if err != nil {
		logError("failed to delete Table ", err)
	}
	return fmt.Sprintf("Deleted Table %d successfully", id)
}

func dbDeleteField(db *sql.DB, ctx context.Context, id int) string {
	queries := mdb.New(db)
	err := queries.DeleteField(ctx, int64(id))
	if err != nil {
		logError("failed to delete Field ", err)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)

}
