package main

import (
	"context"
	"database/sql"
	_ "embed"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbListAdminRoute(db *sql.DB, ctx context.Context) []mdb.Adminroute {
	queries := new(mdb.Queries)
	fetchedAdminRoutes, err := queries.ListAdminRoute(ctx)
	if err != nil {
		logError("failed to get Admin Routes: ", err)
	}
	return fetchedAdminRoutes
}

func dbListRoute(db *sql.DB, ctx context.Context) []mdb.Route {
	queries := new(mdb.Queries)
	fetchedRoutes, err := queries.ListRoute(ctx)
	if err != nil {
		logError("failed to get Routes ", err)
	}
	return fetchedRoutes
}

func dbListUser(db *sql.DB, ctx context.Context) []mdb.User {
	queries := new(mdb.Queries)
	fetchedUsers, err := queries.ListUser(ctx)
	if err != nil {
		logError("failed to get Users ", err)
	}
	return fetchedUsers
}

func dbListMedia(db *sql.DB, ctx context.Context) []mdb.Media {
	queries := new(mdb.Queries)
	fetchedMedias, err := queries.ListMedia(ctx)
	if err != nil {
		logError("failed to get Medias ", err)
	}
	return fetchedMedias 
}

func dbListMediaDimension(db *sql.DB, ctx context.Context) []mdb.MediaDimension {
	queries := new(mdb.Queries)
	fetchedMediaDimensions, err := queries.ListMediaDimension(ctx)
	if err != nil {
		logError("failed to get MediaDimensions ", err)
	}
	return fetchedMediaDimensions
}

func dbListTable(db *sql.DB, ctx context.Context) []mdb.Tables {
	queries := new(mdb.Queries)
	fetchedTables, err := queries.ListTable(ctx)
	if err != nil {
		logError("failed to get Tables ", err)
	}
	return fetchedTables
}

func dbListField(db *sql.DB, ctx context.Context) []mdb.Field {
	queries := new(mdb.Queries)
	fetchedFields, err := queries.ListField(ctx)
	if err != nil {
		logError("failed to get Fields ", err)
	}
	return fetchedFields
}

func dbListFieldsByRoute(db *sql.DB,ctx context.Context, id int64)[]mdb.ListFieldJoinRow{
	queries := new(mdb.Queries)
	fetchedFields, err := queries.ListFieldJoin(ctx,id)
    if err != nil { 
        logError("failed to list and join fields: ", err)
    }
    return fetchedFields
}
