package main

import (
	"context"
	"database/sql"
	_ "embed"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)


func dbCreateAdminRoute(db *sql.DB, ctx context.Context, s mdb.CreateAdminRouteParams) mdb.Adminroute {
	queries := new(mdb.Queries)
	insertedAdminRoute, err := queries.CreateAdminRoute(ctx, s)
	if err != nil {
		logError("failed to create admin route ", err)
	}

	return insertedAdminRoute
}

func dbCreateField(db *sql.DB, ctx context.Context, s mdb.CreateFieldParams) mdb.Field {

	queries := new(mdb.Queries)
	insertedField, err := queries.CreateField(ctx, s)
	if err != nil {
		logError("failed to create field ", err)
	}

	return insertedField
}

func dbCreateMedia(db *sql.DB, ctx context.Context, s mdb.CreateMediaParams) mdb.Media {

	queries := new(mdb.Queries)
	insertedMedia, err := queries.CreateMedia(ctx, s)
	if err != nil {
		logError("failed to create media ", err)
	}

	return insertedMedia
}

func dbCreateMediaDimension(db *sql.DB, ctx context.Context, s mdb.CreateMediaDimensionParams) mdb.MediaDimension {

	queries := new(mdb.Queries)
	insertedMediaDimension, err := queries.CreateMediaDimension(ctx, s)
	if err != nil {
		logError("failed to create MediaDimension ", err)
	}

	return insertedMediaDimension
}

func dbCreateRoute(db *sql.DB, ctx context.Context, s mdb.CreateRouteParams) mdb.Route {

	queries := new(mdb.Queries)
	insertedRoute, err := queries.CreateRoute(ctx, s)
	if err != nil {
		logError("failed to create route ", err)
	}

	return insertedRoute

}

func dbCreateUser(db *sql.DB, ctx context.Context, s mdb.CreateUserParams) mdb.User {


	queries := new(mdb.Queries)
	insertedUser, err := queries.CreateUser(ctx, s)
	if err != nil {
		logError("failed to create user ", err)
	}

	return insertedUser
}

func dbCreateTable(db *sql.DB, ctx context.Context, s mdb.Tables) mdb.Tables {


	queries := new(mdb.Queries)
	insertedTable, err := queries.CreateTable(ctx, s.Label)
	if err != nil {
		logError("failed to create table ", err)
	}

	return insertedTable


}
