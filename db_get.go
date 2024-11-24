package main

import (
	"context"
	"database/sql"
	_ "embed"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbGetAdminRoute(db *sql.DB, ctx context.Context, slug string) mdb.Adminroute {
	queries := mdb.New(db)
	params := ns(slug)
	fetchedAdminRoute, err := queries.GetAdminRoute(ctx, params)
	if err != nil {
		logError("failed to create database dump in archive: ", err)
	}
	return fetchedAdminRoute
}

func dbGetRoute(db *sql.DB, ctx context.Context, slug string) mdb.Route {
	queries := mdb.New(db)
	fetchedRoute, err := queries.GetRoute(ctx, ns(slug))
	if err != nil {
		logError("failed to get Route ", err)
	}
	return fetchedRoute
}

func dbGetUser(db *sql.DB, ctx context.Context, id int64) (mdb.User, error) {
	queries := mdb.New(db)
	fetchedUser, err := queries.GetUser(ctx, id)
	if err != nil {
		logError("failed to get User ", err)
		return fetchedUser, err
	}
	return fetchedUser, nil
}
func dbGetUserId(db *sql.DB, ctx context.Context, id int64) int64 {
	queries := mdb.New(db)
	fetchedUser, err := queries.GetUserId(ctx, id)
	if err != nil {
		logError("failed to get UserId ", err)
	}
	return fetchedUser
}

func dbGetMedia(db *sql.DB, ctx context.Context, id int64) mdb.Media {
	queries := mdb.New(db)
	fetchedMedia, err := queries.GetMedia(ctx, id)
	if err != nil {
		logError("failed to get Media ", err)
	}
	return fetchedMedia
}

func dbGetMediaDimension(db *sql.DB, ctx context.Context, id int64) mdb.MediaDimension {
	queries := mdb.New(db)
	fetchedMediaDimension, err := queries.GetMediaDimension(ctx, id)
	if err != nil {
		logError("failed to get MediaDimension ", err)
	}
	return fetchedMediaDimension

}

func dbGetTable(db *sql.DB, ctx context.Context, id int64) mdb.Tables {
	queries := mdb.New(db)
	fetchedTable, err := queries.GetTable(ctx, id)
	if err != nil {
		logError("failed to get Table ", err)
	}
	return fetchedTable
}

func dbGetField(db *sql.DB, ctx context.Context, id int64) mdb.Field {
	queries := mdb.New(db)
	fetchedField, err := queries.GetField(ctx, id)
	if err != nil {
		logError("failed to get Field ", err)
	}
	return fetchedField

}
