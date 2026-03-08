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
	SortOrder    string `json:"sort_order"`
	Name         string `json:"name"`
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
		SortOrder:    fmt.Sprintf("%d", a.SortOrder),
		Name:         a.Name,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// GetDatatypeByName returns the datatype matching the given name (SQLite).
func (d Database) GetDatatypeByName(name string) (*Datatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetDatatypeByName(d.Context, mdb.GetDatatypeByNameParams{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatype by name %q: %w", name, err)
	}
	m := d.MapDatatype(row)
	return &m, nil
}

// GetDatatypeByName returns the datatype matching the given name (MySQL).
func (d MysqlDatabase) GetDatatypeByName(name string) (*Datatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetDatatypeByName(d.Context, mdbm.GetDatatypeByNameParams{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatype by name %q: %w", name, err)
	}
	m := d.MapDatatype(row)
	return &m, nil
}

// GetDatatypeByName returns the datatype matching the given name (PostgreSQL).
func (d PsqlDatabase) GetDatatypeByName(name string) (*Datatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetDatatypeByName(d.Context, mdbp.GetDatatypeByNameParams{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatype by name %q: %w", name, err)
	}
	m := d.MapDatatype(row)
	return &m, nil
}

// GetDatatypeByType returns the first datatype matching the given type string (SQLite).
func (d Database) GetDatatypeByType(t string) (*Datatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetDatatypeByType(d.Context, mdb.GetDatatypeByTypeParams{Type: t})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatype by type %q: %w", t, err)
	}
	m := d.MapDatatype(row)
	return &m, nil
}

// GetDatatypeByType returns the first datatype matching the given type string (MySQL).
func (d MysqlDatabase) GetDatatypeByType(t string) (*Datatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetDatatypeByType(d.Context, mdbm.GetDatatypeByTypeParams{Type: t})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatype by type %q: %w", t, err)
	}
	m := d.MapDatatype(row)
	return &m, nil
}

// GetDatatypeByType returns the first datatype matching the given type string (PostgreSQL).
func (d PsqlDatabase) GetDatatypeByType(t string) (*Datatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetDatatypeByType(d.Context, mdbp.GetDatatypeByTypeParams{Type: t})
	if err != nil {
		return nil, fmt.Errorf("failed to get datatype by type %q: %w", t, err)
	}
	m := d.MapDatatype(row)
	return &m, nil
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

// ListDatatypesGlobal returns all datatypes with type _global.
func (d Database) ListDatatypesGlobal() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeGlobal(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get global Datatypes: %w", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesGlobal returns all datatypes with type _global.
func (d MysqlDatabase) ListDatatypesGlobal() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeGlobal(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get global Datatypes: %w", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesGlobal returns all datatypes with type _global.
func (d PsqlDatabase) ListDatatypesGlobal() (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeGlobal(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get global Datatypes: %w", err)
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
