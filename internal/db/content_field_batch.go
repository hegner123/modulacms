package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/types"
)

// listContentFieldsByContentDataIDs fetches all content fields for a batch of
// content data IDs using QSelect with an IN clause. When locale is non-empty,
// only fields matching that locale or the empty locale (non-translatable) are returned.
func listContentFieldsByContentDataIDs(ctx context.Context, conn *sql.DB, dialect Dialect, ids []types.ContentID, locale string) (*[]ContentFields, error) {
	if len(ids) == 0 {
		empty := []ContentFields{}
		return &empty, nil
	}
	vals := make([]any, len(ids))
	for i, id := range ids {
		vals[i] = id.String()
	}
	where := map[string]any{
		"content_data_id": In(vals...),
	}
	var whereOr []map[string]any
	if locale != "" {
		whereOr = []map[string]any{
			{"locale": locale},
			{"locale": ""},
		}
	}
	rows, err := QSelect(ctx, conn, dialect, SelectParams{
		Table:   "content_fields",
		Columns: nil,
		Where:   where,
		WhereOr: whereOr,
		Limit:   -1,
	})
	if err != nil {
		return nil, fmt.Errorf("batch content field select: %w", err)
	}
	result := make([]ContentFields, 0, len(rows))
	for _, row := range rows {
		result = append(result, rowToContentFields(row))
	}
	return &result, nil
}

// rowToContentFields maps a QSelect Row (map[string]any) to a ContentFields struct.
func rowToContentFields(row Row) ContentFields {
	cf := ContentFields{}
	cf.ContentFieldID = types.ContentFieldID(rowString(row, "content_field_id"))
	cf.RouteID = rowNullableRouteID(row, "route_id")
	cf.ContentDataID = rowNullableContentID(row, "content_data_id")
	cf.FieldID = rowNullableFieldID(row, "field_id")
	cf.FieldValue = rowString(row, "field_value")
	cf.Locale = rowString(row, "locale")
	cf.AuthorID = types.UserID(rowString(row, "author_id"))
	cf.DateCreated = rowTimestamp(row, "date_created")
	cf.DateModified = rowTimestamp(row, "date_modified")
	return cf
}

// rowString extracts a string value from a Row, handling both string and []byte
// (SQLite returns []byte for TEXT columns via database/sql).
func rowString(row Row, key string) string {
	v, ok := row[key]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// rowTimestamp extracts a Timestamp from a Row by delegating to Timestamp.Scan.
func rowTimestamp(row Row, key string) types.Timestamp {
	v, ok := row[key]
	if !ok || v == nil {
		return types.Timestamp{}
	}
	var ts types.Timestamp
	// Scan handles string, []byte, time.Time, and nil.
	// Ignore scan errors — return zero Timestamp on failure.
	_ = ts.Scan(v) // Scan only fails on unparseable strings — no recovery possible for raw row data
	return ts
}

// rowNullableContentID extracts a nullable ContentID from a Row.
func rowNullableContentID(row Row, key string) types.NullableContentID {
	s := rowString(row, key)
	if s == "" {
		return types.NullableContentID{}
	}
	return types.NullableContentID{ID: types.ContentID(s), Valid: true}
}

// rowNullableRouteID extracts a nullable RouteID from a Row.
func rowNullableRouteID(row Row, key string) types.NullableRouteID {
	s := rowString(row, key)
	if s == "" {
		return types.NullableRouteID{}
	}
	return types.NullableRouteID{ID: types.RouteID(s), Valid: true}
}

// rowNullableFieldID extracts a nullable FieldID from a Row.
func rowNullableFieldID(row Row, key string) types.NullableFieldID {
	s := rowString(row, key)
	if s == "" {
		return types.NullableFieldID{}
	}
	return types.NullableFieldID{ID: types.FieldID(s), Valid: true}
}

// ListContentFieldsByContentDataIDs returns all content fields for a batch of content
// data IDs, optionally filtered by locale (SQLite).
func (d Database) ListContentFieldsByContentDataIDs(ctx context.Context, ids []types.ContentID, locale string) (*[]ContentFields, error) {
	return listContentFieldsByContentDataIDs(ctx, d.Connection, DialectSQLite, ids, locale)
}

// ListContentFieldsByContentDataIDs returns all content fields for a batch of content
// data IDs, optionally filtered by locale (MySQL).
func (d MysqlDatabase) ListContentFieldsByContentDataIDs(ctx context.Context, ids []types.ContentID, locale string) (*[]ContentFields, error) {
	return listContentFieldsByContentDataIDs(ctx, d.Connection, DialectMySQL, ids, locale)
}

// ListContentFieldsByContentDataIDs returns all content fields for a batch of content
// data IDs, optionally filtered by locale (PostgreSQL).
func (d PsqlDatabase) ListContentFieldsByContentDataIDs(ctx context.Context, ids []types.ContentID, locale string) (*[]ContentFields, error) {
	return listContentFieldsByContentDataIDs(ctx, d.Connection, DialectPostgres, ids, locale)
}
