package main

import (
	"context"
	"database/sql"
	_ "embed"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbGetAdminRoute(db *sql.DB, ctx context.Context, slug string) mdb.Adminroute {
	queries := new(mdb.Queries)
	params := mdb.GetAdminRouteParams{
		Column1: "slug",
		Column2: slug,
	}
	fetchedAdminRoute, err := queries.GetAdminRoute(ctx, params)
	if err != nil {
		logError("failed to create database dump in archive: ", err)
	}
	return fetchedAdminRoute
}

func dbGetRoute(db *sql.DB, ctx context.Context, slug string) mdb.Route {
	queries := new(mdb.Queries)
	params := mdb.GetRouteParams{
		Column1: "slug",
		Column2: slug,
	}
	fetchedRoute, err := queries.GetRoute(ctx, params)
	if err != nil {
		logError("failed to get Route ", err)
	}
	return fetchedRoute
}

func dbGetUser(db *sql.DB, ctx context.Context, id int) mdb.User {
	queries := new(mdb.Queries)
	params := mdb.GetUserParams{
		Column1: "id",
		Column2: id,
	}
	fetchedUser, err := queries.GetUser(ctx, params)
	if err != nil {
		logError("failed to get User ", err)
	}
	return fetchedUser
}

func dbGetMedia(db *sql.DB, ctx context.Context, id int) mdb.Media {
	queries := new(mdb.Queries)
	params := mdb.GetMediaParams{
		Column1: "id",
		Column2: id,
	}
	fetchedMedia, err := queries.GetMedia(ctx, params)
	if err != nil {
		logError("failed to get Media ", err)
	}
	return fetchedMedia
}

func dbGetMediaDimension(db *sql.DB, ctx context.Context, id int) mdb.MediaDimension {
	queries := new(mdb.Queries)
	params := mdb.GetMediaDimensionParams{
		Column1: "id",
		Column2: id,
	}
	fetchedMediaDimension, err := queries.GetMediaDimension(ctx, params)
	if err != nil {
		logError("failed to get MediaDimension ", err)
	}
	return fetchedMediaDimension

}

func dbGetTable(db *sql.DB, ctx context.Context, id int) mdb.Tables {
	queries := new(mdb.Queries)
	params := mdb.GetTableParams{
		Column1: "id",
		Column2: id,
	}
	fetchedTable, err := queries.GetTable(ctx, params)
	if err != nil {
		logError("failed to get Table ", err)
	}
	return fetchedTable
}

func dbGetField(db *sql.DB, ctx context.Context, id int) mdb.Field {
	queries := new(mdb.Queries)
	params := mdb.GetFieldParams{
		Column1: "id",
		Column2: id,
	}
	fetchedField, err := queries.GetField(ctx, params)
	if err != nil {
		logError("failed to get Field ", err)
	}
	return fetchedField

}
