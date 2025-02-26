package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func GetAdminDatatypeGlobalId(db *sql.DB, ctx context.Context) (*mdb.AdminDatatypes, error) {
	queries := mdb.New(db)
	fetchedGlobalAdminDatatypeId, err := queries.GetGlobalAdminDatatypeId(ctx)
	if err != nil {
		return nil, err
	}
	return &fetchedGlobalAdminDatatypeId, nil
}

func GetAdminDatatypeById(db *sql.DB, ctx context.Context, id int64) (*mdb.AdminDatatypes, error) {
	queries := mdb.New(db)
	fetchedAdminDatatype, err := queries.GetAdminDatatype(ctx, id)
	if err != nil {
		return nil, err
	}
	return &fetchedAdminDatatype, nil
}

func GetRootAdIdByAdRtId(db *sql.DB, ctx context.Context, adminRtId int64) (*sql.NullInt64, error) {
	queries := mdb.New(db)
	res := sql.NullInt64{Int64: int64(0), Valid: true}
	fetchedAdminDatatype, err := queries.GetRootAdminDtByAdminRtId(ctx, ni64(adminRtId))
	if err != nil {
		fmt.Printf("adminRtId %d\n", adminRtId)
		res = sql.NullInt64{Valid: false}
		return nil, err
	}
	res.Int64 = fetchedAdminDatatype.AdminDtID
	return &res, nil
}

func GetAdminFieldID(db *sql.DB, ctx context.Context, id int64) (*mdb.AdminFields, error) {
	queries := mdb.New(db)
	fetchedAdminField, err := queries.GetAdminField(ctx, id)
	if err != nil {
		return nil, err
	}
	return &fetchedAdminField, nil
}

func GetAdminRoute(db *sql.DB, ctx context.Context, slug string) (*mdb.AdminRoutes, error) {
	queries := mdb.New(db)
	fetchedAdminRoute, err := queries.GetAdminRouteBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return &fetchedAdminRoute, nil
}

func GetDatatype(db *sql.DB, ctx context.Context, id int64) (*mdb.Datatypes, error) {
	queries := mdb.New(db)
	fetchedDatatype, err := queries.GetDatatype(ctx, id)
	if err != nil {
		return nil, err
	}
	return &fetchedDatatype, nil
}

func GetField(db *sql.DB, ctx context.Context, id int64) (*mdb.Fields, error) {
	queries := mdb.New(db)
	fetchedField, err := queries.GetField(ctx, id)
	if err != nil {
		return nil, err
	}
	return &fetchedField, nil
}

func GetMedia(db *sql.DB, ctx context.Context, id int64) (*mdb.Media, error) {
	queries := mdb.New(db)
	fetchedMedia, err := queries.GetMedia(ctx, id)
	if err != nil {
		return nil, err
	}
	return &fetchedMedia, nil
}

func GetMediaDimension(db *sql.DB, ctx context.Context, id int64) (*mdb.MediaDimensions, error) {
	queries := mdb.New(db)
	fetchedMediaDimension, err := queries.GetMediaDimension(ctx, id)
	if err != nil {
		return nil, err
	}
	return &fetchedMediaDimension, nil
}

func GetRoute(db *sql.DB, ctx context.Context, slug string) (*mdb.Routes, error) {
	queries := mdb.New(db)
	fetchedRoute, err := queries.GetRoute(ctx, slug)
	if err != nil {
		return nil, err
	}
	return &fetchedRoute, nil
}

func GetTable(db *sql.DB, ctx context.Context, id int64) (*mdb.Tables, error) {
	queries := mdb.New(db)
	fetchedTable, err := queries.GetTable(ctx, id)
	if err != nil {
		return nil, err
	}
	return &fetchedTable, nil
}

func GetToken(db *sql.DB, ctx context.Context, id int64) (*mdb.Tokens, error) {
	queries := mdb.New(db)
	fetchedToken, err := queries.GetToken(ctx, id)
	if err != nil {
		return nil, err

	}
	return &fetchedToken, nil
}

func GetTokenByUserId(db *sql.DB, ctx context.Context, userId int64) (*[]mdb.Tokens, error) {
	queries := mdb.New(db)
	fetchedToken, err := queries.GetTokensByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	return &fetchedToken, nil
}

func GetUser(db *sql.DB, ctx context.Context, id int64) (*mdb.Users, error) {
	queries := mdb.New(db)
	fetchedUser, err := queries.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &fetchedUser, nil
}

func GetUserByEmail(db *sql.DB, ctx context.Context, email string) (*mdb.Users, error) {
	queries := mdb.New(db)
	fetchedUser, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &fetchedUser, nil
}
