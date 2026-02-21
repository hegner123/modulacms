package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

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
