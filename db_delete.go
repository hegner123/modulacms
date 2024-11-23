package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbDeleteAdminRoute(db *sql.DB, ctx context.Context, id int) string {
	queries := new(mdb.Queries)
	params := mdb.DeleteAdminRouteParams{
		Column1: "id",
		Column2: id,
	}
	err := queries.DeleteAdminRoute(ctx, params)
	if err != nil {
		logError("failed to create database dump in archive: ", err)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)
}

func dbDeleteRoute(db *sql.DB, ctx context.Context, id int) string {
	queries := new(mdb.Queries)
	params := mdb.DeleteRouteParams{
		Column1: "id",
		Column2: id,
	}
	err := queries.DeleteRoute(ctx, params)
	if err != nil {
		logError("failed to get Route ", err)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)
}

func dbDeleteUser(db *sql.DB, ctx context.Context, id int) string {
	queries := new(mdb.Queries)
	params := mdb.DeleteUserParams{
		Column1: "id",
		Column2: id,
	}
	err := queries.DeleteUser(ctx, params)
	if err != nil {
		logError("failed to get User ", err)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)
}

func dbDeleteMedia(db *sql.DB, ctx context.Context, id int) string {
	queries := new(mdb.Queries)
	params := mdb.DeleteMediaParams{
		Column1: "id",
		Column2: id,
	}
	err := queries.DeleteMedia(ctx, params)
	if err != nil {
		logError("failed to get Media ", err)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)
}

func dbDeleteMediaDimension(db *sql.DB, ctx context.Context, id int) string {
	queries := new(mdb.Queries)
	params := mdb.DeleteMediaDimensionParams{
		Column1: "id",
		Column2: id,
	}
	err := queries.DeleteMediaDimension(ctx, params)
	if err != nil {
		logError("failed to get MediaDimension ", err)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)

}

func dbDeleteTable(db *sql.DB, ctx context.Context, id int) string {
	queries := new(mdb.Queries)
	params := mdb.DeleteTableParams{
		Column1: "id",
		Column2: id,
	}
	err := queries.DeleteTable(ctx, params)
	if err != nil {
		logError("failed to get Table ", err)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)
}

func dbDeleteField(db *sql.DB, ctx context.Context, id int) string {
	queries := new(mdb.Queries)
	params := mdb.DeleteFieldParams{
		Column1: "id",
		Column2: id,
	}
	err := queries.DeleteField(ctx, params)
	if err != nil {
		logError("failed to get Field ", err)
	}
	return fmt.Sprintf("Deleted Field %d successfully", id)

}
