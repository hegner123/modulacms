package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func ListAdminDatatypes(db *sql.DB, ctx context.Context) (*[]mdb.AdminDatatypes, error) {
	queries := mdb.New(db)
	fetchedAdminDatatypes, err := queries.ListAdminDatatype(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Datatypes: %v\n", err)
	}
	return &fetchedAdminDatatypes, nil
}

func ListAdminFields(db *sql.DB, ctx context.Context) (*[]mdb.AdminFields, error) {
	queries := mdb.New(db)
	fetchedAdminFields, err := queries.ListAdminField(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	return &fetchedAdminFields, nil
}

func ListAdminRoute(db *sql.DB, ctx context.Context) (*[]mdb.AdminRoutes, error) {
	queries := mdb.New(db)
	fetchedAdminRoutes, err := queries.ListAdminRoute(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Routes: %v\n", err)
	}
	return &fetchedAdminRoutes, nil
}

func ListDatatype(db *sql.DB, ctx context.Context) (*[]mdb.Datatypes, error) {
	queries := mdb.New(db)
	fetchedDatatypes, err := queries.ListDatatype(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	return &fetchedDatatypes, nil
}

func ListField(db *sql.DB, ctx context.Context) (*[]mdb.Fields, error) {
	queries := mdb.New(db)
	fetchedFields, err := queries.ListField(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	return &fetchedFields, nil
}

func ListMedia(db *sql.DB, ctx context.Context) (*[]mdb.Media, error) {
	queries := mdb.New(db)
	fetchedMedias, err := queries.ListMedia(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Medias: %v\n", err)
	}
	return &fetchedMedias, nil
}

func ListMediaDimension(db *sql.DB, ctx context.Context) (*[]mdb.MediaDimensions, error) {
	queries := mdb.New(db)
	fetchedMediaDimensions, err := queries.ListMediaDimension(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get MediaDimensions: %v\n", err)
	}
	return &fetchedMediaDimensions, nil
}

func ListRoute(db *sql.DB, ctx context.Context) (*[]mdb.Routes, error) {
	queries := mdb.New(db)
	fetchedRoutes, err := queries.ListRoute(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes: %v\n", err)
	}
	return &fetchedRoutes, nil
}

func ListTable(db *sql.DB, ctx context.Context) (*[]mdb.Tables, error) {
	queries := mdb.New(db)
	fetchedTables, err := queries.ListTable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tables: %v\n", err)
	}
	return &fetchedTables, nil
}

func ListUser(db *sql.DB, ctx context.Context) (*[]mdb.Users, error) {
	queries := mdb.New(db)
	fetchedUsers, err := queries.ListUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	return &fetchedUsers, nil
}
func ListTokens(db *sql.DB, ctx context.Context) (*[]mdb.Tokens, error) {
	queries := mdb.New(db)
	fetchedTokens, err := queries.ListTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	return &fetchedTokens, nil
}

func ListTokenDependencies(db *sql.DB, ctx context.Context, id int64) {
	// TODO implement dependency checking for delete candidate
}

func ListDatatypeById(db *sql.DB, ctx context.Context, routeId int64) (*[]mdb.ListDatatypeByRouteIdRow, error) {
	queries := mdb.New(db)
	fetchedDatatypes, err := queries.ListDatatypeByRouteId(ctx, ni64(routeId))
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	return &fetchedDatatypes, nil
}

func ListFieldByRouteId(db *sql.DB, ctx context.Context, routeId int64) (*[]mdb.ListFieldByRouteIdRow, error) {
	queries := mdb.New(db)
	fetchedDatatypes, err := queries.ListFieldByRouteId(ctx, ni64(routeId))
	if err != nil {
		return nil, fmt.Errorf("failed to get fields with route %d: %v\n", routeId, err)
	}
	return &fetchedDatatypes, nil
}

func ListAdminFieldByAdminDtId(db *sql.DB, ctx context.Context, admin_datatype_id int64) (*[]mdb.ListAdminFieldByAdminDtIdRow, error) {
	queries := mdb.New(db)
	fetchedFields, err := queries.ListAdminFieldByAdminDtId(ctx, ni64(admin_datatype_id))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields by AdminDatatypes id: %v\n ", err)
	}
	return &fetchedFields, nil
}

func ListAdminDatatypeByAdminRouteId(db *sql.DB, ctx context.Context, adminRouteId int64) (*[]mdb.ListAdminDatatypeByRouteIdRow, error) {
	queries := mdb.New(db)
	fetchedAdminDatatypes, err := queries.ListAdminDatatypeByRouteId(ctx, ni64(adminRouteId))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes by AdminRouteId %v\n", err)
	}
	return &fetchedAdminDatatypes, nil
}

func ListAdminDatatypeChildren(db *sql.DB, ctx context.Context, parentId int64) (*[]mdb.AdminDatatypes, error) {
	queries := mdb.New(db)
	fetchedAdminDatatypeChildren, err := queries.ListAdminDatatypeChildren(ctx, ni64(parentId))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes by AdminRouteId %v\n", err)
	}
	return &fetchedAdminDatatypeChildren, nil

}
