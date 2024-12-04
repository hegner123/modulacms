package main

import (
	"context"
	"database/sql"
	_ "embed"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbGetAdminDatatypeGlobalId(db *sql.DB, ctx context.Context) mdb.AdminDatatype{
	queries := mdb.New(db)
    fetchedGlobalAdminDatatypeId, err := queries.GetGlobalAdminDatatypeId(ctx)
	if err != nil {
		logError("failed to get admin route", err)
	}
    return fetchedGlobalAdminDatatypeId
}

func dbGetAdminRoute(db *sql.DB, ctx context.Context, slug string) mdb.AdminRoute {
	queries := mdb.New(db)
	fetchedAdminRoute, err := queries.GetAdminRouteBySlug(ctx, slug)
	if err != nil {
		logError("failed to get admin route", err)
	}
	return fetchedAdminRoute
}

func dbGetDatatype(db *sql.DB, ctx context.Context, id int64) mdb.Datatype {
	queries := mdb.New(db)
	fetchedDatatype, err := queries.GetDatatype(ctx, id)
	if err != nil {
		logError("failed to get Datatype ", err)
	}
	return fetchedDatatype
}

func dbGetField(db *sql.DB, ctx context.Context, id int64) mdb.Field {
	queries := mdb.New(db)
	fetchedField, err := queries.GetField(ctx, id)
	if err != nil {
		logError("failed to get Field ", err)
	}
	return fetchedField
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

func dbGetRoute(db *sql.DB, ctx context.Context, slug string) mdb.Route {
	queries := mdb.New(db)
	fetchedRoute, err := queries.GetRoute(ctx, slug)
	if err != nil {
		logError("failed to get Route ", err)
	}
	return fetchedRoute
}

func dbGetTable(db *sql.DB, ctx context.Context, id int64) mdb.Tables {
	queries := mdb.New(db)
	fetchedTable, err := queries.GetTable(ctx, id)
	if err != nil {
		logError("failed to get Table ", err)
	}
	return fetchedTable
}

func dbGetToken(db *sql.DB, ctx context.Context, id int64) mdb.Token {
	queries := mdb.New(db)
	fetchedToken, err := queries.GetToken(ctx, id)
	if err != nil {
		logError("failed to get Token ", err)
	}
	return fetchedToken
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

func dbGetUserByEmail(db *sql.DB, ctx context.Context, email string) mdb.User {
	queries := mdb.New(db)
	fetchedUser, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		logError("failed to get User via email ", err)
	}
	return fetchedUser
}
