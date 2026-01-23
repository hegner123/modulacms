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

type Fields struct {
	FieldID      types.FieldID           `json:"field_id"`
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}
type CreateFieldParams struct {
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}
type UpdateFieldParams struct {
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
	FieldID      types.FieldID           `json:"field_id"`
}
// FieldsJSON is used for JSON serialization in model package
type FieldsJSON struct {
	FieldID      string `json:"field_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
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
		Type:         a.Type.String(),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringField converts Fields to StringFields for table display
func MapStringField(a Fields) StringFields {
	return StringFields{
		FieldID:      a.FieldID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type.String(),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
		History:      "", // History field removed from schema
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapField(a mdb.Fields) Fields {
	return Fields{
		FieldID:      a.FieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateFieldParams(a CreateFieldParams) mdb.CreateFieldParams {
	return mdb.CreateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateFieldParams(a UpdateFieldParams) mdb.UpdateFieldParams {
	return mdb.UpdateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		FieldID:      a.FieldID,
	}
}

// QUERIES

func (d Database) CountFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateFieldTable(d.Context)
	return err
}
func (d Database) CreateField(s CreateFieldParams) Fields {
	params := d.MapCreateFieldParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateField  %v \n", err)
	}

	return d.MapField(row)
}
func (d Database) DeleteField(id types.FieldID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteField(d.Context, mdb.DeleteFieldParams{FieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete field: %v\n", id)
	}

	return nil
}

func (d Database) GetField(id types.FieldID) (*Fields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetField(d.Context, mdb.GetFieldParams{FieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapField(row)
	return &res, nil
}

func (d Database) ListFields() (*[]Fields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListFieldsByDatatypeID(id types.NullableContentID) (*[]Fields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListFieldByDatatypeID(d.Context, mdb.ListFieldByDatatypeIDParams{ParentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateField(s UpdateFieldParams) (*string, error) {
	params := d.MapUpdateFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapField(a mdbm.Fields) Fields {
	return Fields{
		FieldID:      a.FieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateFieldParams(a CreateFieldParams) mdbm.CreateFieldParams {
	return mdbm.CreateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapUpdateFieldParams(a UpdateFieldParams) mdbm.UpdateFieldParams {
	return mdbm.UpdateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		FieldID:      a.FieldID,
	}
}

// QUERIES

func (d MysqlDatabase) CountFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CreateFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateFieldTable(d.Context)
	return err
}
func (d MysqlDatabase) CreateField(s CreateFieldParams) Fields {
	params := d.MapCreateFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateField: %v\n", err)
	}
	row, err := queries.GetLastField(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Field: %v\n", err)
	}
	return d.MapField(row)
}
func (d MysqlDatabase) DeleteField(id types.FieldID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteField(d.Context, mdbm.DeleteFieldParams{FieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete field: %v", id)
	}

	return nil
}
func (d MysqlDatabase) GetField(id types.FieldID) (*Fields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetField(d.Context, mdbm.GetFieldParams{FieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapField(row)
	return &res, nil
}
func (d MysqlDatabase) ListFields() (*[]Fields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListFieldsByDatatypeID(id types.NullableContentID) (*[]Fields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListFieldByDatatypeID(d.Context, mdbm.ListFieldByDatatypeIDParams{ParentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) UpdateField(s UpdateFieldParams) (*string, error) {
	params := d.MapUpdateFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapField(a mdbp.Fields) Fields {
	return Fields{
		FieldID:      a.FieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateFieldParams(a CreateFieldParams) mdbp.CreateFieldParams {
	return mdbp.CreateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateFieldParams(a UpdateFieldParams) mdbp.UpdateFieldParams {
	return mdbp.UpdateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		FieldID:      a.FieldID,
	}
}

// QUERIES

func (d PsqlDatabase) CountFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CreateFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateFieldTable(d.Context)
	return err
}
func (d PsqlDatabase) CreateField(s CreateFieldParams) Fields {
	params := d.MapCreateFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateField  %v \n", err)
	}

	return d.MapField(row)
}
func (d PsqlDatabase) DeleteField(id types.FieldID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteField(d.Context, mdbp.DeleteFieldParams{FieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete field: %v", id)
	}

	return nil
}
func (d PsqlDatabase) GetField(id types.FieldID) (*Fields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetField(d.Context, mdbp.GetFieldParams{FieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapField(row)
	return &res, nil
}
func (d PsqlDatabase) ListFields() (*[]Fields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListFieldsByDatatypeID(id types.NullableContentID) (*[]Fields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListFieldByDatatypeID(d.Context, mdbp.ListFieldByDatatypeIDParams{ParentID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) UpdateField(s UpdateFieldParams) (*string, error) {
	params := d.MapUpdateFieldParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}
