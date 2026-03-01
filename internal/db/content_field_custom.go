package db

import (
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ListContentFieldsByContentDataAndLocale returns content fields for a content data ID filtered by locale (SQLite).
// Returns fields matching the given locale plus non-translatable fields (locale = "").
func (d Database) ListContentFieldsByContentDataAndLocale(contentDataID types.NullableContentID, locale string) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentDataAndLocale(d.Context, mdb.ListContentFieldsByContentDataAndLocaleParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by contentDataID and locale: %w", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentFieldsByContentDataAndLocale returns content fields for a content data ID filtered by locale (MySQL).
func (d MysqlDatabase) ListContentFieldsByContentDataAndLocale(contentDataID types.NullableContentID, locale string) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentDataAndLocale(d.Context, mdbm.ListContentFieldsByContentDataAndLocaleParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by contentDataID and locale: %w", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentFieldsByContentDataAndLocale returns content fields for a content data ID filtered by locale (PostgreSQL).
func (d PsqlDatabase) ListContentFieldsByContentDataAndLocale(contentDataID types.NullableContentID, locale string) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentDataAndLocale(d.Context, mdbp.ListContentFieldsByContentDataAndLocaleParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by contentDataID and locale: %w", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentFieldsByRouteAndLocale returns content fields for a route filtered by locale (SQLite).
// Returns fields matching the given locale plus non-translatable fields (locale = "").
func (d Database) ListContentFieldsByRouteAndLocale(routeID types.NullableRouteID, locale string) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsByRouteAndLocale(d.Context, mdb.ListContentFieldsByRouteAndLocaleParams{
		RouteID: routeID,
		Locale:  locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by routeID and locale: %w", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentFieldsByRouteAndLocale returns content fields for a route filtered by locale (MySQL).
func (d MysqlDatabase) ListContentFieldsByRouteAndLocale(routeID types.NullableRouteID, locale string) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsByRouteAndLocale(d.Context, mdbm.ListContentFieldsByRouteAndLocaleParams{
		RouteID: routeID,
		Locale:  locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by routeID and locale: %w", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentFieldsByRouteAndLocale returns content fields for a route filtered by locale (PostgreSQL).
func (d PsqlDatabase) ListContentFieldsByRouteAndLocale(routeID types.NullableRouteID, locale string) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByRouteAndLocale(d.Context, mdbp.ListContentFieldsByRouteAndLocaleParams{
		RouteID: routeID,
		Locale:  locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by routeID and locale: %w", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ContentFieldsJSON is used for JSON serialization in model package
// Deprecated: Will be removed in future version. Use typed ContentFields directly.
type ContentFieldsJSON struct {
	ContentFieldID int64  `json:"content_field_id"`
	RouteID        int64  `json:"route_id"`
	ContentDataID  int64  `json:"content_data_id"`
	FieldID        int64  `json:"field_id"`
	FieldValue     string `json:"field_value"`
	AuthorID       int64  `json:"author_id"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
}

// MapContentFieldJSON converts ContentFields to ContentFieldsJSON for JSON serialization
// Deprecated: Will be removed in future version
func MapContentFieldJSON(a ContentFields) ContentFieldsJSON {
	return ContentFieldsJSON{
		ContentFieldID: 0, // Type conversion not available, set to 0
		RouteID:        0, // Type conversion not available, set to 0
		ContentDataID:  0, // Type conversion not available, set to 0
		FieldID:        0, // Type conversion not available, set to 0
		FieldValue:     a.FieldValue,
		AuthorID:       0, // Type conversion not available, set to 0
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
	}
}

// NullableRouteIDFromInt64 converts int64 to NullableRouteID for backward compatibility.
// Deprecated: Use types.NullableRouteID directly
func NullableRouteIDFromInt64(id int64) types.NullableRouteID {
	return types.NullableRouteID{ID: types.RouteID(strconv.FormatInt(id, 10)), Valid: id != 0}
}

// ListContentFieldsPaginated returns content fields with pagination (SQLite).
func (d Database) ListContentFieldsPaginated(params PaginationParams) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsPaginated(d.Context, mdb.ListContentFieldsPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields paginated: %w", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentFieldsPaginated returns content fields with pagination (MySQL).
func (d MysqlDatabase) ListContentFieldsPaginated(params PaginationParams) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsPaginated(d.Context, mdbm.ListContentFieldsPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields paginated: %w", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListContentFieldsPaginated returns content fields with pagination (PostgreSQL).
func (d PsqlDatabase) ListContentFieldsPaginated(params PaginationParams) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsPaginated(d.Context, mdbp.ListContentFieldsPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields paginated: %w", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
