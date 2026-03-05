package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/types"
)

// rowNullableAdminContentID extracts a nullable AdminContentID from a Row.
func rowNullableAdminContentID(row Row, key string) types.NullableAdminContentID {
	s := rowString(row, key)
	if s == "" {
		return types.NullableAdminContentID{}
	}
	return types.NullableAdminContentID{ID: types.AdminContentID(s), Valid: true}
}

// rowNullableAdminRouteID extracts a nullable AdminRouteID from a Row.
func rowNullableAdminRouteID(row Row, key string) types.NullableAdminRouteID {
	s := rowString(row, key)
	if s == "" {
		return types.NullableAdminRouteID{}
	}
	return types.NullableAdminRouteID{ID: types.AdminRouteID(s), Valid: true}
}

// rowNullableAdminFieldID extracts a nullable AdminFieldID from a Row.
func rowNullableAdminFieldID(row Row, key string) types.NullableAdminFieldID {
	s := rowString(row, key)
	if s == "" {
		return types.NullableAdminFieldID{}
	}
	return types.NullableAdminFieldID{ID: types.AdminFieldID(s), Valid: true}
}

// rowToAdminContentFields maps a QSelect Row (map[string]any) to an AdminContentFields struct.
func rowToAdminContentFields(row Row) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: types.AdminContentFieldID(rowString(row, "admin_content_field_id")),
		AdminRouteID:        rowNullableAdminRouteID(row, "admin_route_id"),
		AdminContentDataID:  rowNullableAdminContentID(row, "admin_content_data_id"),
		AdminFieldID:        rowNullableAdminFieldID(row, "admin_field_id"),
		AdminFieldValue:     rowString(row, "admin_field_value"),
		Locale:              rowString(row, "locale"),
		AuthorID:            types.UserID(rowString(row, "author_id")),
		DateCreated:         rowTimestamp(row, "date_created"),
		DateModified:        rowTimestamp(row, "date_modified"),
	}
}

// listAdminContentFieldsByContentDataIDs fetches all admin content fields for a batch of
// admin content data IDs using QSelect with an IN clause. When locale is non-empty,
// only fields matching that locale or the empty locale (non-translatable) are returned.
func listAdminContentFieldsByContentDataIDs(ctx context.Context, conn *sql.DB, dialect Dialect, ids []types.AdminContentID, locale string) (*[]AdminContentFields, error) {
	if len(ids) == 0 {
		empty := []AdminContentFields{}
		return &empty, nil
	}
	vals := make([]any, len(ids))
	for i, id := range ids {
		vals[i] = id.String()
	}
	where := map[string]any{
		"admin_content_data_id": In(vals...),
	}
	var whereOr []map[string]any
	if locale != "" {
		whereOr = []map[string]any{
			{"locale": locale},
			{"locale": ""},
		}
	}
	rows, err := QSelect(ctx, conn, dialect, SelectParams{
		Table:   "admin_content_fields",
		Columns: nil,
		Where:   where,
		WhereOr: whereOr,
		Limit:   -1,
	})
	if err != nil {
		return nil, fmt.Errorf("batch admin content field select: %w", err)
	}
	result := make([]AdminContentFields, 0, len(rows))
	for _, row := range rows {
		result = append(result, rowToAdminContentFields(row))
	}
	return &result, nil
}

// ListAdminContentFieldsByContentDataIDs returns all admin content fields for a batch of
// admin content data IDs, optionally filtered by locale (SQLite).
func (d Database) ListAdminContentFieldsByContentDataIDs(ctx context.Context, ids []types.AdminContentID, locale string) (*[]AdminContentFields, error) {
	return listAdminContentFieldsByContentDataIDs(ctx, d.Connection, DialectSQLite, ids, locale)
}

// ListAdminContentFieldsByContentDataIDs returns all admin content fields for a batch of
// admin content data IDs, optionally filtered by locale (MySQL).
func (d MysqlDatabase) ListAdminContentFieldsByContentDataIDs(ctx context.Context, ids []types.AdminContentID, locale string) (*[]AdminContentFields, error) {
	return listAdminContentFieldsByContentDataIDs(ctx, d.Connection, DialectMySQL, ids, locale)
}

// ListAdminContentFieldsByContentDataIDs returns all admin content fields for a batch of
// admin content data IDs, optionally filtered by locale (PostgreSQL).
func (d PsqlDatabase) ListAdminContentFieldsByContentDataIDs(ctx context.Context, ids []types.AdminContentID, locale string) (*[]AdminContentFields, error) {
	return listAdminContentFieldsByContentDataIDs(ctx, d.Connection, DialectPostgres, ids, locale)
}
