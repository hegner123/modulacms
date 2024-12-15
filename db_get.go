package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbGetAdminDatatypeGlobalId(db *sql.DB, ctx context.Context) mdb.AdminDatatypes {
	queries := mdb.New(db)
	fetchedGlobalAdminDatatypeId, err := queries.GetGlobalAdminDatatypeId(ctx)
	if err != nil {
		logError("failed to get Global AdminDatatypes", err)
	}
	return fetchedGlobalAdminDatatypeId
}

func dbGetAdminDatatypeById(db *sql.DB, ctx context.Context, id int64) mdb.AdminDatatypes {
	queries := mdb.New(db)
	fetchedAdminDatatype, err := queries.GetAdminDatatype(ctx, id)
	if err != nil {
		logError("failed to get Global AdminDatatypes", err)
	}
	return fetchedAdminDatatype
}

func dbGetRootAdIdByAdRtId(db *sql.DB, ctx context.Context, adminRtId int64) sql.NullInt64  {
	queries := mdb.New(db)
    res := sql.NullInt64{Int64: int64(0),Valid: true}
	fetchedAdminDatatype, err := queries.GetRootAdminDtByAdminRtId(ctx, ni64(adminRtId))
	if err != nil {
        fmt.Printf("adminRtId %d\n", adminRtId)
		logError("failed to get  Admin Datatype", err)
        res = sql.NullInt64{Valid:false}
        return res
	}
    res.Int64 = fetchedAdminDatatype.AdminDtID
	return res 
}

func dbGetAdminField(db *sql.DB, ctx context.Context, id int64) mdb.AdminFields {
	queries := mdb.New(db)
	fetchedAdminField, err := queries.GetAdminField(ctx, id)
	if err != nil {
		logError("failed to get Global AdminDatatypes", err)
	}
	return fetchedAdminField
}

func dbGetAdminRoute(db *sql.DB, ctx context.Context, slug string) mdb.AdminRoutes {
	queries := mdb.New(db)
	fetchedAdminRoute, err := queries.GetAdminRouteBySlug(ctx, slug)
	if err != nil {
		logError("failed to get admin route", err)
	}
	return fetchedAdminRoute
}

func dbGetDatatype(db *sql.DB, ctx context.Context, id int64) mdb.Datatypes {
	queries := mdb.New(db)
	fetchedDatatype, err := queries.GetDatatype(ctx, id)
	if err != nil {
		logError("failed to get Datatype ", err)
	}
	return fetchedDatatype
}

func dbGetField(db *sql.DB, ctx context.Context, id int64) mdb.Fields {
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

func dbGetMediaDimension(db *sql.DB, ctx context.Context, id int64) mdb.MediaDimensions {
	queries := mdb.New(db)
	fetchedMediaDimension, err := queries.GetMediaDimension(ctx, id)
	if err != nil {
		logError("failed to get MediaDimension ", err)
	}
	return fetchedMediaDimension
}

func dbGetRoute(db *sql.DB, ctx context.Context, slug string) mdb.Routes {
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

func dbGetToken(db *sql.DB, ctx context.Context, id int64) mdb.Tokens {
	queries := mdb.New(db)
	fetchedToken, err := queries.GetToken(ctx, id)
	if err != nil {
		logError("failed to get Token ", err)
	}
	return fetchedToken
}

func dbGetTokenByUserId(db *sql.DB, ctx context.Context, userId int64) mdb.Tokens {
	queries := mdb.New(db)
	fetchedToken, err := queries.GetTokenByUserId(ctx, userId)
	if err != nil {
		logError("failed to get Token ", err)
	}
	return fetchedToken
}

func dbGetUser(db *sql.DB, ctx context.Context, id int64) (mdb.Users, error) {
	queries := mdb.New(db)
	fetchedUser, err := queries.GetUser(ctx, id)
	if err != nil {
		logError("failed to get User ", err)
		return fetchedUser, err
	}
	return fetchedUser, nil
}

func dbGetUserByEmail(db *sql.DB, ctx context.Context, email string) mdb.Users {
	queries := mdb.New(db)
	fetchedUser, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		logError("failed to get User via email ", err)
	}
	return fetchedUser
}
