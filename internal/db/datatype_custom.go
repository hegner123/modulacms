package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

// DatatypeJSON provides a string-based representation for JSON serialization.
type DatatypeJSON struct {
	DatatypeID   string `json:"datatype_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// MapDatatypeJSON converts Datatypes to DatatypeJSON for JSON serialization.
func MapDatatypeJSON(a Datatypes) DatatypeJSON {
	return DatatypeJSON{
		DatatypeID:   a.DatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// ListDatatypesRoot returns all root-level datatypes.
func (d Database) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesRoot returns all root-level datatypes.
func (d MysqlDatabase) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesRoot returns all root-level datatypes.
func (d PsqlDatabase) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesPaginated returns datatypes with pagination (SQLite).
func (d Database) ListDatatypesPaginated(params PaginationParams) (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypePaginated(d.Context, mdb.ListDatatypePaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes paginated: %w", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesPaginated returns datatypes with pagination (MySQL).
func (d MysqlDatabase) ListDatatypesPaginated(params PaginationParams) (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypePaginated(d.Context, mdbm.ListDatatypePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes paginated: %w", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesPaginated returns datatypes with pagination (PostgreSQL).
func (d PsqlDatabase) ListDatatypesPaginated(params PaginationParams) (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypePaginated(d.Context, mdbp.ListDatatypePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes paginated: %w", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}
