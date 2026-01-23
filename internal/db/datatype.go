package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type Datatypes struct {
	DatatypeID   types.DatatypeID        `json:"datatype_id"`
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

type CreateDatatypeParams struct {
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

type UpdateDatatypeParams struct {
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
	DatatypeID   types.DatatypeID        `json:"datatype_id"`
}

// DatatypeJSON provides a string-based representation for JSON serialization
type DatatypeJSON struct {
	DatatypeID   string `json:"datatype_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// MapDatatypeJSON converts Datatypes to DatatypeJSON for JSON serialization
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

// MapStringDatatype converts Datatypes to StringDatatypes for table display
func MapStringDatatype(a Datatypes) StringDatatypes {
	return StringDatatypes{
		DatatypeID:   a.DatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
		History:      "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapDatatype(a mdb.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateDatatypeParams(a CreateDatatypeParams) mdb.CreateDatatypeParams {
	return mdb.CreateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdb.UpdateDatatypeParams {
	return mdb.UpdateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		DatatypeID:   a.DatatypeID,
	}
}

// QUERIES

func (d Database) CountDatatypes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d Database) CreateDatatype(s CreateDatatypeParams) Datatypes {
	params := d.MapCreateDatatypeParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatype: %v\n", err)
	}
	return d.MapDatatype(row)
}

func (d Database) DeleteDatatype(id types.DatatypeID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, mdb.DeleteDatatypeParams{DatatypeID: id})
	if err != nil {
		return fmt.Errorf("failed to delete datatype: %v", id)
	}
	return nil
}

func (d Database) GetDatatype(id types.DatatypeID) (*Datatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, mdb.GetDatatypeParams{DatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d Database) ListDatatypes() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
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

func (d Database) UpdateDatatype(s UpdateDatatypeParams) (*string, error) {
	params := d.MapUpdateDatatypeParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapDatatype(a mdbm.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbm.CreateDatatypeParams {
	return mdbm.CreateDatatypeParams{
		ParentID: a.ParentID,
		Label:    a.Label,
		Type:     a.Type,
		AuthorID: a.AuthorID,
	}
}

func (d MysqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbm.UpdateDatatypeParams {
	return mdbm.UpdateDatatypeParams{
		ParentID:   a.ParentID,
		Label:      a.Label,
		Type:       a.Type,
		AuthorID:   a.AuthorID,
		DatatypeID: a.DatatypeID,
	}
}

// QUERIES

func (d MysqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateDatatypeTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateDatatype(s CreateDatatypeParams) Datatypes {
	params := d.MapCreateDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatype: %v\n", err)
	}
	row, err := queries.GetLastDatatype(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Datatype: %v\n", err)
	}
	return d.MapDatatype(row)
}

func (d MysqlDatabase) DeleteDatatype(id types.DatatypeID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, mdbm.DeleteDatatypeParams{DatatypeID: id})
	if err != nil {
		return fmt.Errorf("failed to delete datatype: %v", id)
	}
	return nil
}

func (d MysqlDatabase) GetDatatype(id types.DatatypeID) (*Datatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, mdbm.GetDatatypeParams{DatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d MysqlDatabase) ListDatatypes() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
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

func (d MysqlDatabase) UpdateDatatype(s UpdateDatatypeParams) (*string, error) {
	params := d.MapUpdateDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapDatatype(a mdbp.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbp.CreateDatatypeParams {
	return mdbp.CreateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbp.UpdateDatatypeParams {
	return mdbp.UpdateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		DatatypeID:   a.DatatypeID,
	}
}

// QUERIES

func (d PsqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateDatatypeTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateDatatype(s CreateDatatypeParams) Datatypes {
	params := d.MapCreateDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatype: %v\n", err)
	}
	return d.MapDatatype(row)
}

func (d PsqlDatabase) DeleteDatatype(id types.DatatypeID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteDatatype(d.Context, mdbp.DeleteDatatypeParams{DatatypeID: id})
	if err != nil {
		return fmt.Errorf("failed to delete datatype: %v", id)
	}
	return nil
}

func (d PsqlDatabase) GetDatatype(id types.DatatypeID) (*Datatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, mdbp.GetDatatypeParams{DatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d PsqlDatabase) ListDatatypes() (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
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

func (d PsqlDatabase) UpdateDatatype(s UpdateDatatypeParams) (*string, error) {
	params := d.MapUpdateDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}
