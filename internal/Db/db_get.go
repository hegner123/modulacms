package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func DbGetAdminDatatypeGlobalId(db *sql.DB, ctx context.Context) mdb.AdminDatatypes {
	queries := mdb.New(db)
	fetchedGlobalAdminDatatypeId, err := queries.GetGlobalAdminDatatypeId(ctx)
	if err != nil {
		logError("failed to get Global AdminDatatypes", err)
	}
	return fetchedGlobalAdminDatatypeId
}

func DbGetAdminDatatypeById(db *sql.DB, ctx context.Context, id int64) mdb.AdminDatatypes {
	queries := mdb.New(db)
	fetchedAdminDatatype, err := queries.GetAdminDatatype(ctx, id)
	if err != nil {
		logError("failed to get Global AdminDatatypes", err)
	}
	return fetchedAdminDatatype
}

func DbGetRootAdIdByAdRtId(db *sql.DB, ctx context.Context, adminRtId int64) sql.NullInt64  {
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

func DbGetAdminField(db *sql.DB, ctx context.Context, id int64) mdb.AdminFields {
	queries := mdb.New(db)
	fetchedAdminField, err := queries.GetAdminField(ctx, id)
	if err != nil {
		logError("failed to get Global AdminDatatypes", err)
	}
	return fetchedAdminField
}

func DbGetAdminRoute(db *sql.DB, ctx context.Context, slug string) mdb.AdminRoutes {
	queries := mdb.New(db)
	fetchedAdminRoute, err := queries.GetAdminRouteBySlug(ctx, slug)
	if err != nil {
		logError("failed to get admin route", err)
	}
	return fetchedAdminRoute
}

func DbGetDatatype(db *sql.DB, ctx context.Context, id int64) mdb.Datatypes {
	queries := mdb.New(db)
	fetchedDatatype, err := queries.GetDatatype(ctx, id)
	if err != nil {
		logError("failed to get Datatype ", err)
	}
	return fetchedDatatype
}

func DbGetField(db *sql.DB, ctx context.Context, id int64) mdb.Fields {
	queries := mdb.New(db)
	fetchedField, err := queries.GetField(ctx, id)
	if err != nil {
		logError("failed to get Field ", err)
	}
	return fetchedField
}

func DbGetMedia(db *sql.DB, ctx context.Context, id int64) mdb.Media {
	queries := mdb.New(db)
	fetchedMedia, err := queries.GetMedia(ctx, id)
	if err != nil {
		logError("failed to get Media ", err)
	}
	return fetchedMedia
}

func DbGetMediaDimension(db *sql.DB, ctx context.Context, id int64) mdb.MediaDimensions {
	queries := mdb.New(db)
	fetchedMediaDimension, err := queries.GetMediaDimension(ctx, id)
	if err != nil {
		logError("failed to get MediaDimension ", err)
	}
	return fetchedMediaDimension
}

func DbGetRoute(db *sql.DB, ctx context.Context, slug string) mdb.Routes {
	queries := mdb.New(db)
	fetchedRoute, err := queries.GetRoute(ctx, slug)
	if err != nil {
		logError("failed to get Route ", err)
	}
	return fetchedRoute
}

func DbGetTable(db *sql.DB, ctx context.Context, id int64) mdb.Tables {
	queries := mdb.New(db)
	fetchedTable, err := queries.GetTable(ctx, id)
	if err != nil {
		logError("failed to get Table ", err)
	}
	return fetchedTable
}

func DbGetToken(db *sql.DB, ctx context.Context, id int64) mdb.Tokens {
	queries := mdb.New(db)
	fetchedToken, err := queries.GetToken(ctx, id)
	if err != nil {
		logError("failed to get Token ", err)
	}
	return fetchedToken
}

func DbGetTokenByUserId(db *sql.DB, ctx context.Context, userId int64) mdb.Tokens {
	queries := mdb.New(db)
	fetchedToken, err := queries.GetTokenByUserId(ctx, userId)
	if err != nil {
		logError("failed to get Token ", err)
	}
	return fetchedToken
}

func DbGetUser(db *sql.DB, ctx context.Context, id int64) (mdb.Users, error) {
	queries := mdb.New(db)
	fetchedUser, err := queries.GetUser(ctx, id)
	if err != nil {
		logError("failed to get User ", err)
		return fetchedUser, err
	}
	return fetchedUser, nil
}

func DbGetUserByEmail(db *sql.DB, ctx context.Context, email string) mdb.Users {
	queries := mdb.New(db)
	fetchedUser, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		logError("failed to get User via email ", err)
	}
	return fetchedUser
}
