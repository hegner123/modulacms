package main

import (
	"context"
	"database/sql"
	_ "embed"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbListAdminRoute(db *sql.DB, ctx context.Context) []mdb.Adminroute {
	queries := mdb.New(db)
	fetchedAdminRoutes, err := queries.ListAdminRoute(ctx)
	if err != nil {
		logError("failed to get Admin Routes: ", err)
	}
	return fetchedAdminRoutes
}

func dbListDatatype(db *sql.DB, ctx context.Context) []mdb.Datatype {
	queries := mdb.New(db)
	fetchedDatatypes, err := queries.ListDatatype(ctx)
	if err != nil {
		logError("failed to get Datatypes ", err)
	}
	return fetchedDatatypes
}

func dbListField(db *sql.DB, ctx context.Context) []mdb.Field {
	queries := mdb.New(db)
	fetchedFields, err := queries.ListField(ctx)
	if err != nil {
		logError("failed to get Fields ", err)
	}
	return fetchedFields
}

func dbListFieldsByRoute(db *sql.DB, ctx context.Context, id int64) []mdb.ListFieldJoinRow {
	queries := mdb.New(db)
	fetchedFields, err := queries.ListFieldJoin(ctx, ni64(id))
	if err != nil {
		logError("failed to list and join fields: ", err)
	}
	return fetchedFields
}

func dbListMedia(db *sql.DB, ctx context.Context) []mdb.Media {
	queries := mdb.New(db)
	fetchedMedias, err := queries.ListMedia(ctx)
	if err != nil {
		logError("failed to get Medias ", err)
	}
	return fetchedMedias
}

func dbListMediaDimension(db *sql.DB, ctx context.Context) []mdb.MediaDimension {
	queries := mdb.New(db)
	fetchedMediaDimensions, err := queries.ListMediaDimension(ctx)
	if err != nil {
		logError("failed to get MediaDimensions ", err)
	}
	return fetchedMediaDimensions
}

func dbListRoute(db *sql.DB, ctx context.Context) []mdb.Route {
	queries := mdb.New(db)
	fetchedRoutes, err := queries.ListRoute(ctx)
	if err != nil {
		logError("failed to get Routes ", err)
	}
	return fetchedRoutes
}

func dbListTable(db *sql.DB, ctx context.Context) []mdb.Tables {
	queries := mdb.New(db)
	fetchedTables, err := queries.ListTable(ctx)
	if err != nil {
		logError("failed to get Tables ", err)
	}
	return fetchedTables
}

func dbListUser(db *sql.DB, ctx context.Context) []mdb.User {
	queries := mdb.New(db)
	fetchedUsers, err := queries.ListUser(ctx)
	if err != nil {
		logError("failed to get Users ", err)
	}
	return fetchedUsers
}

func dbListTokenDependencies(db *sql.DB, ctx context.Context, id int64) {
	// TODO implement dependency checking for delete candidate

}
