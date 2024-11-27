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
	err := queries.DeleteAdminRoute(ctx, slug)
	if err != nil {
		logError("failed to delete admin route ", err)
	    return fmt.Sprintf("failed to delete Admin Route %s ", slug)
	}
	return fmt.Sprintf("Deleted Admin Route %s successfully", slug)
}

func dbDeleteDataType(db *sql.DB, ctx context.Context, id int64) string {
	queries := mdb.New(db)
	err := queries.DeleteDatatype(ctx, id)
	if err != nil {
		logError("failed to delete Field ", err)
	    return fmt.Sprintf("failed to delete datatype %d ", id)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)
}

func dbDeleteField(db *sql.DB, ctx context.Context, id int64) string {
	queries := mdb.New(db)
	err := queries.DeleteField(ctx, int64(id))
	if err != nil {
		logError("failed to delete Field ", err)
	    return fmt.Sprintf("failed to delete Field %d ", id)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)
}

func dbDeleteMedia(db *sql.DB, ctx context.Context, id int64) string {
	queries := mdb.New(db)
	err := queries.DeleteMedia(ctx, int64(id))
	if err != nil {
		logError("failed to delete Media ", err)
	    return fmt.Sprintf("failed to delete Media %d ", id)
	}
	return fmt.Sprintf("Deleted Media %d successfully", id)
}

func dbDeleteMediaDimension(db *sql.DB, ctx context.Context, id int64) string {
	queries := mdb.New(db)
	err := queries.DeleteMediaDimension(ctx, int64(id))
	if err != nil {
		logError("failed to delete MediaDimension ", err)
	    return fmt.Sprintf("failed to delete MediaDimension %d ", id)
	}
	return fmt.Sprintf("Deleted Media Dimension %d successfully", id)
}

func dbDeleteRoute(db *sql.DB, ctx context.Context, slug string) string {
	queries := mdb.New(db)
	err := queries.DeleteRoute(ctx, slug)
	if err != nil {
		logError("failed to delete Route ", err)
	    return fmt.Sprintf("failed to delete  Route %s ", slug)
	}
	return fmt.Sprintf("Deleted Route %s successfully", slug)
}

func dbDeleteTable(db *sql.DB, ctx context.Context, id int64) string {
	queries := mdb.New(db)
	err := queries.DeleteTable(ctx, id)
	if err != nil {
		logError("failed to delete Table ", err)
	    return fmt.Sprintf("failed to delete table %d ", id)
	}
	return fmt.Sprintf("Deleted Table %d successfully", id)
}

func dbDeleteToken(db *sql.DB, ctx context.Context, id int64) string {
	queries := mdb.New(db)
	err := queries.DeleteToken(ctx, id)
	if err != nil {
		logError("failed to delete Table ", err)
	    return fmt.Sprintf("failed to delete Token %d ", id)
	}
	return fmt.Sprintf("Deleted Table %d successfully", id)
}

func dbDeleteUser(db *sql.DB, ctx context.Context, id int64) string {
	queries := mdb.New(db)
	err := queries.DeleteUser(ctx, id)
	if err != nil {
		logError("failed to delete User ", err)
	    return fmt.Sprintf("failed to delete User %d ", id)
	}
	return fmt.Sprintf("Deleted User %d successfully", id)
}
