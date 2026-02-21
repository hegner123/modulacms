package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ListAdminFieldByRouteIdRow represents a result row from listing admin fields by route ID.
type ListAdminFieldByRouteIdRow struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                        `json:"label"`
	Data         string                        `json:"data"`
	Validation   string                        `json:"validation"`
	UIConfig     string                        `json:"ui_config"`
	Type         types.FieldType               `json:"type"`
}

// ListAdminFieldsByDatatypeIDRow represents a result row from listing admin fields by datatype ID.
type ListAdminFieldsByDatatypeIDRow struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                        `json:"label"`
	Data         string                        `json:"data"`
	Validation   string                        `json:"validation"`
	UIConfig     string                        `json:"ui_config"`
	Type         types.FieldType               `json:"type"`
}

// UtilityGetAdminfieldsRow represents a result row from utility admin fields retrieval.
type UtilityGetAdminfieldsRow struct {
	AdminFieldID types.AdminFieldID `json:"admin_field_id"`
	Label        string             `json:"label"`
}

// ListAdminFieldsPaginated returns admin fields with pagination (SQLite).
func (d Database) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdb.ListAdminFieldPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %w", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminFieldsPaginated returns admin fields with pagination (MySQL).
func (d MysqlDatabase) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdbm.ListAdminFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %w", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminFieldsPaginated returns admin fields with pagination (PostgreSQL).
func (d PsqlDatabase) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdbp.ListAdminFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %w", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// MapAdminFieldJSON converts AdminFields to FieldsJSON for tree building
// by mapping admin field ID into the public FieldsJSON shape.
func MapAdminFieldJSON(a AdminFields) FieldsJSON {
	return FieldsJSON{
		FieldID:      a.AdminFieldID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UIConfig,
		Type:         a.Type.String(),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}
