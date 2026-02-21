package db

import (
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

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
		ContentFieldID: 0,                       // Type conversion not available, set to 0
		RouteID:        0,                       // Type conversion not available, set to 0
		ContentDataID:  0,                       // Type conversion not available, set to 0
		FieldID:        0,                       // Type conversion not available, set to 0
		FieldValue:     a.FieldValue,
		AuthorID:       0,                       // Type conversion not available, set to 0
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
