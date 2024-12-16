package db

import (
	"context"
	"database/sql"
	_ "embed"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func dbListAdminDatatypes(db *sql.DB, ctx context.Context) []mdb.AdminDatatypes {
	queries := mdb.New(db)
	fetchedAdminDatatypes, err := queries.ListAdminDatatype(ctx)
	if err != nil {
		logError("failed to get Admin Datatypes: ", err)
	}
	return fetchedAdminDatatypes
}

func dbListAdminFields(db *sql.DB, ctx context.Context) []mdb.AdminFields {
	queries := mdb.New(db)
	fetchedAdminFields, err := queries.ListAdminField(ctx)
	if err != nil {
		logError("failed to get Admin Fields: ", err)
	}
	return fetchedAdminFields
}

func dbListAdminRoute(db *sql.DB, ctx context.Context) []mdb.AdminRoutes {
	queries := mdb.New(db)
	fetchedAdminRoutes, err := queries.ListAdminRoute(ctx)
	if err != nil {
		logError("failed to get Admin Routes: ", err)
	}
	return fetchedAdminRoutes
}

func dbListDatatype(db *sql.DB, ctx context.Context) []mdb.Datatypes {
	queries := mdb.New(db)
	fetchedDatatypes, err := queries.ListDatatype(ctx)
	if err != nil {
		logError("failed to get Datatypes ", err)
	}
	return fetchedDatatypes
}

func dbListField(db *sql.DB, ctx context.Context) []mdb.Fields {
	queries := mdb.New(db)
	fetchedFields, err := queries.ListField(ctx)
	if err != nil {
		logError("failed to get Fields ", err)
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

func dbListMediaDimension(db *sql.DB, ctx context.Context) []mdb.MediaDimensions {
	queries := mdb.New(db)
	fetchedMediaDimensions, err := queries.ListMediaDimension(ctx)
	if err != nil {
		logError("failed to get MediaDimensions ", err)
	}
	return fetchedMediaDimensions
}

func dbListRoute(db *sql.DB, ctx context.Context) []mdb.Routes {
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

func dbListUser(db *sql.DB, ctx context.Context) []mdb.Users {
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

func dbListDatatypeById(db *sql.DB, ctx context.Context, routeId int64) []mdb.ListDatatypeByRouteIdRow {
	queries := mdb.New(db)
	fetchedDatatypes, err := queries.ListDatatypeByRouteId(ctx, ni64(routeId))
	if err != nil {
		logError("failed to get Users ", err)
	}
	return fetchedDatatypes
}

func dbListFieldById(db *sql.DB, ctx context.Context, routeId int64) []mdb.ListFieldByRouteIdRow {
	queries := mdb.New(db)
	fetchedDatatypes, err := queries.ListFieldByRouteId(ctx, ni64(routeId))
	if err != nil {
		logError("failed to get Users ", err)
	}
	return fetchedDatatypes
}

func dbListAdminFieldByAdminDtId(db *sql.DB, ctx context.Context, admin_datatype_id int64) []mdb.ListAdminFieldByAdminDtIdRow {
	queries := mdb.New(db)
	fetchedFields, err := queries.ListAdminFieldByAdminDtId(ctx, ni64(admin_datatype_id))
	if err != nil {
		logError("failed to get Users ", err)
	}
	return fetchedFields
}

func dbListAdminDatatypeByAdminRouteId(db *sql.DB, ctx context.Context, adminRouteId int64) []mdb.ListAdminDatatypeByRouteIdRow {
	queries := mdb.New(db)
	fetchedAdminDatatypes, err := queries.ListAdminDatatypeByRouteId(ctx, ni64(adminRouteId))
	if err != nil {
		logError("failed to get Users ", err)
	}
	return fetchedAdminDatatypes
}
