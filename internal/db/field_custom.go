package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

// ListFieldsPaginated returns fields with pagination (SQLite).
func (d Database) ListFieldsPaginated(params PaginationParams) (*[]Fields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListFieldPaginated(d.Context, mdb.ListFieldPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields paginated: %w", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListFieldsPaginated returns fields with pagination (MySQL).
func (d MysqlDatabase) ListFieldsPaginated(params PaginationParams) (*[]Fields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListFieldPaginated(d.Context, mdbm.ListFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields paginated: %w", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListFieldsPaginated returns fields with pagination (PostgreSQL).
func (d PsqlDatabase) ListFieldsPaginated(params PaginationParams) (*[]Fields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListFieldPaginated(d.Context, mdbp.ListFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields paginated: %w", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

// FieldsJSON is used for JSON serialization in model package
type FieldsJSON struct {
	FieldID      string `json:"field_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Validation   string `json:"validation"`
	UIConfig     string `json:"ui_config"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// MapFieldJSON converts Fields to FieldsJSON for JSON serialization
func MapFieldJSON(a Fields) FieldsJSON {
	return FieldsJSON{
		FieldID:      a.FieldID.String(),
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
