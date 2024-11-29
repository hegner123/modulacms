package main

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbCreateAdminRoute(db *sql.DB, ctx context.Context, s mdb.CreateAdminRouteParams) mdb.Adminroute {
	queries := mdb.New(db)
	insertedAdminRoute, err := queries.CreateAdminRoute(ctx, s)
	if err != nil {
		logError("failed to create admin route ", err)
	}

	return insertedAdminRoute
}

func dbCreateDataType(db *sql.DB, ctx context.Context, s mdb.CreateDatatypeParams) (mdb.Datatype, error) {
	queries := mdb.New(db)
	insertedDatatype, err := queries.CreateDatatype(ctx, s)
	if err != nil {
		logError("failed to create field ", err)
		return insertedDatatype, err
	}

	return insertedDatatype, nil
}

func dbCreateField(db *sql.DB, ctx context.Context, s mdb.CreateFieldParams) (mdb.Field, error) {
	queries := mdb.New(db)
	insertedField, err := queries.CreateField(ctx, s)
	if err != nil {
		logError("failed to create field ", err)
		return insertedField, err
	}

	return insertedField, nil
}

func dbCreateMedia(db *sql.DB, ctx context.Context, s mdb.CreateMediaParams) mdb.Media {
	queries := mdb.New(db)
	insertedMedia, err := queries.CreateMedia(ctx, s)
	if err != nil {
		logError("failed to create media ", err)
	}

	return insertedMedia
}

func dbCreateMediaDimension(db *sql.DB, ctx context.Context, s mdb.CreateMediaDimensionParams) mdb.MediaDimension {
	queries := mdb.New(db)
	insertedMediaDimension, err := queries.CreateMediaDimension(ctx, s)
	if err != nil {
		logError("failed to create MediaDimension ", err)
	}

	return insertedMediaDimension
}

func dbCreateRoute(db *sql.DB, ctx context.Context, s mdb.CreateRouteParams) mdb.Route {
	queries := mdb.New(db)
	insertedRoute, err := queries.CreateRoute(ctx, s)
	if err != nil {
		logError("failed to create route ", err)
	}

	return insertedRoute
}

func dbCreateTable(db *sql.DB, ctx context.Context, s mdb.Tables) mdb.Tables {
	queries := mdb.New(db)
	insertedTable, err := queries.CreateTable(ctx, s.Label)
	if err != nil {
		logError("failed to create table ", err)
	}

	return insertedTable
}

func dbCreateToken(db *sql.DB, ctx context.Context, s mdb.CreateTokenParams) mdb.Token {
	queries := mdb.New(db)
	insertedToken, err := queries.CreateToken(ctx, s)
	if err != nil {
		logError("failed to create table ", err)
	}

	return insertedToken
}

func dbCreateUser(db *sql.DB, ctx context.Context, s mdb.CreateUserParams) mdb.User {
	queries := mdb.New(db)
	insertedUser, err := queries.CreateUser(ctx, s)
	if err != nil {
		splitErr := strings.Split(popError(err), ".")
		property := splitErr[len(splitErr)-1]
		v := getColumnValue(property, s)

		logError("failed to create user ", err, property, v)
	}

	return insertedUser
}
