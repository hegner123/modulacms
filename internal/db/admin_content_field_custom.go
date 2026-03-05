package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ListAdminContentFieldsByContentData returns all admin content fields for a content item (SQLite).
func (d Database) ListAdminContentFieldsByContentData(contentDataID types.NullableAdminContentID) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByContentData(d.Context, mdb.ListAdminContentFieldsByContentDataParams{
		AdminContentDataID: contentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by contentDataID: %w", err)
	}
	res := make([]AdminContentFields, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentField(v))
	}
	return &res, nil
}

// ListAdminContentFieldsByContentData returns all admin content fields for a content item (MySQL).
func (d MysqlDatabase) ListAdminContentFieldsByContentData(contentDataID types.NullableAdminContentID) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByContentData(d.Context, mdbm.ListAdminContentFieldsByContentDataParams{
		AdminContentDataID: contentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by contentDataID: %w", err)
	}
	res := make([]AdminContentFields, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentField(v))
	}
	return &res, nil
}

// ListAdminContentFieldsByContentData returns all admin content fields for a content item (PostgreSQL).
func (d PsqlDatabase) ListAdminContentFieldsByContentData(contentDataID types.NullableAdminContentID) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByContentData(d.Context, mdbp.ListAdminContentFieldsByContentDataParams{
		AdminContentDataID: contentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by contentDataID: %w", err)
	}
	res := make([]AdminContentFields, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentField(v))
	}
	return &res, nil
}

// ListAdminContentFieldsByContentDataAndLocale returns admin content fields filtered by locale (SQLite).
// Returns fields matching the given locale plus non-translatable fields (locale = "").
func (d Database) ListAdminContentFieldsByContentDataAndLocale(contentDataID types.NullableAdminContentID, locale string) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByContentDataAndLocale(d.Context, mdb.ListAdminContentFieldsByContentDataAndLocaleParams{
		AdminContentDataID: contentDataID,
		Locale:             locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by contentDataID and locale: %w", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminContentFieldsByContentDataAndLocale returns admin content fields filtered by locale (MySQL).
func (d MysqlDatabase) ListAdminContentFieldsByContentDataAndLocale(contentDataID types.NullableAdminContentID, locale string) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByContentDataAndLocale(d.Context, mdbm.ListAdminContentFieldsByContentDataAndLocaleParams{
		AdminContentDataID: contentDataID,
		Locale:             locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by contentDataID and locale: %w", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminContentFieldsByContentDataAndLocale returns admin content fields filtered by locale (PostgreSQL).
func (d PsqlDatabase) ListAdminContentFieldsByContentDataAndLocale(contentDataID types.NullableAdminContentID, locale string) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByContentDataAndLocale(d.Context, mdbp.ListAdminContentFieldsByContentDataAndLocaleParams{
		AdminContentDataID: contentDataID,
		Locale:             locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by contentDataID and locale: %w", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminContentFieldsByRouteAndLocale returns admin content fields for a route filtered by locale (SQLite).
func (d Database) ListAdminContentFieldsByRouteAndLocale(routeID types.NullableAdminRouteID, locale string) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRouteAndLocale(d.Context, mdb.ListAdminContentFieldsByRouteAndLocaleParams{
		AdminRouteID: routeID,
		Locale:       locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by routeID and locale: %w", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminContentFieldsByRouteAndLocale returns admin content fields for a route filtered by locale (MySQL).
func (d MysqlDatabase) ListAdminContentFieldsByRouteAndLocale(routeID types.NullableAdminRouteID, locale string) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRouteAndLocale(d.Context, mdbm.ListAdminContentFieldsByRouteAndLocaleParams{
		AdminRouteID: routeID,
		Locale:       locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by routeID and locale: %w", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminContentFieldsByRouteAndLocale returns admin content fields for a route filtered by locale (PostgreSQL).
func (d PsqlDatabase) ListAdminContentFieldsByRouteAndLocale(routeID types.NullableAdminRouteID, locale string) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRouteAndLocale(d.Context, mdbp.ListAdminContentFieldsByRouteAndLocaleParams{
		AdminRouteID: routeID,
		Locale:       locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by routeID and locale: %w", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminContentFieldsPaginated returns admin content fields with pagination (SQLite).
func (d Database) ListAdminContentFieldsPaginated(params PaginationParams) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsPaginated(d.Context, mdb.ListAdminContentFieldsPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields paginated: %w", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminContentFieldsPaginated returns admin content fields with pagination (MySQL).
func (d MysqlDatabase) ListAdminContentFieldsPaginated(params PaginationParams) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsPaginated(d.Context, mdbm.ListAdminContentFieldsPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields paginated: %w", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminContentFieldsPaginated returns admin content fields with pagination (PostgreSQL).
func (d PsqlDatabase) ListAdminContentFieldsPaginated(params PaginationParams) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsPaginated(d.Context, mdbp.ListAdminContentFieldsPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields paginated: %w", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// MapAdminContentFieldJSON converts AdminContentFields to ContentFieldsJSON for tree building.
// Maps admin field value into the public ContentFieldsJSON shape so BuildNodes works unchanged.
func MapAdminContentFieldJSON(a AdminContentFields) ContentFieldsJSON {
	return ContentFieldsJSON{
		ContentFieldID: 0,
		RouteID:        0,
		ContentDataID:  0,
		FieldID:        0,
		FieldValue:     a.AdminFieldValue,
		AuthorID:       0,
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
	}
}
