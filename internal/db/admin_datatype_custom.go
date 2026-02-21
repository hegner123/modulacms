package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ListAdminDatatypeByRouteIdRow represents a row from listing admin datatypes by route ID.
type ListAdminDatatypeByRouteIdRow struct {
	AdminDatatypeID types.AdminDatatypeID         `json:"admin_datatype_id"`
	AdminRouteID    types.NullableRouteID         `json:"admin_route_id"`
	ParentID        types.NullableAdminDatatypeID `json:"parent_id"`
	Label           string                        `json:"label"`
	Type            string                        `json:"type"`
}

// UtilityGetAdminDatatypesRow represents a row from the utility query for admin datatypes.
type UtilityGetAdminDatatypesRow struct {
	AdminDatatypeID types.AdminDatatypeID `json:"admin_datatype_id"`
	Label           string                `json:"label"`
}

// GetAdminDatatypeById retrieves an admin datatype by ID.
// This is a legacy alias for GetAdminDatatype required by the DbDriver interface.
func (d Database) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	return d.GetAdminDatatype(id)
}

// GetAdminDatatypeById retrieves an admin datatype by ID.
func (d MysqlDatabase) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	return d.GetAdminDatatype(id)
}

// GetAdminDatatypeById retrieves an admin datatype by ID.
func (d PsqlDatabase) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	return d.GetAdminDatatype(id)
}

// MapAdminDatatypeJSON converts AdminDatatypes to DatatypeJSON for tree building
// by mapping admin datatype ID into the public DatatypeJSON shape.
func MapAdminDatatypeJSON(a AdminDatatypes) DatatypeJSON {
	return DatatypeJSON{
		DatatypeID:   a.AdminDatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// ListAdminDatatypeGlobalId returns all admin datatypes with global scope.
// Note: This query is only available for SQLite.
func (d Database) ListAdminDatatypeGlobalId() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeGlobal(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list admin datatypes global: %w", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypesPaginated returns admin datatypes with pagination (SQLite).
func (d Database) ListAdminDatatypesPaginated(params PaginationParams) (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypePaginated(d.Context, mdb.ListAdminDatatypePaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes paginated: %w", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypesPaginated returns admin datatypes with pagination (MySQL).
func (d MysqlDatabase) ListAdminDatatypesPaginated(params PaginationParams) (*[]AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypePaginated(d.Context, mdbm.ListAdminDatatypePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes paginated: %w", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypesPaginated returns admin datatypes with pagination (PostgreSQL).
func (d PsqlDatabase) ListAdminDatatypesPaginated(params PaginationParams) (*[]AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypePaginated(d.Context, mdbp.ListAdminDatatypePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes paginated: %w", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}
