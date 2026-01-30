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

type AdminFields struct {
	AdminFieldID types.AdminFieldID      `json:"admin_field_id"`
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

type CreateAdminFieldParams struct {
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

type UpdateAdminFieldParams struct {
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Type         types.FieldType         `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
	AdminFieldID types.AdminFieldID      `json:"admin_field_id"`
}

type ListAdminFieldByRouteIdRow struct {
	AdminFieldID types.AdminFieldID      `json:"admin_field_id"`
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Type         types.FieldType         `json:"type"`
}

type ListAdminFieldsByDatatypeIDRow struct {
	AdminFieldID types.AdminFieldID      `json:"admin_field_id"`
	ParentID     types.NullableContentID `json:"parent_id"`
	Label        string                  `json:"label"`
	Data         string                  `json:"data"`
	Type         types.FieldType         `json:"type"`
}

type UtilityGetAdminfieldsRow struct {
	AdminFieldID types.AdminFieldID `json:"admin_field_id"`
	Label        string             `json:"label"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapAdminFieldJSON converts AdminFields to FieldsJSON for tree building.
// Maps admin field ID into the public FieldsJSON shape so BuildNodes works unchanged.
func MapAdminFieldJSON(a AdminFields) FieldsJSON {
	return FieldsJSON{
		FieldID:      a.AdminFieldID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type.String(),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringAdminField converts AdminFields to StringAdminFields for table display
func MapStringAdminField(a AdminFields) StringAdminFields {
	return StringAdminFields{
		AdminFieldID: a.AdminFieldID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Data:         a.Data,
		Type:         string(a.Type),
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

func (d Database) MapAdminField(a mdb.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdb.CreateAdminFieldParams {
	return mdb.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdb.UpdateAdminFieldParams {
	return mdb.UpdateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		AdminFieldID: a.AdminFieldID,
	}
}

// QUERIES

func (d Database) CountAdminFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateAdminField(s CreateAdminFieldParams) AdminFields {
	params := d.MapCreateAdminFieldParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateAdminField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminField  %v \n", err)
	}

	return d.MapAdminField(row)
}

func (d Database) CreateAdminFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

func (d Database) DeleteAdminField(id types.AdminFieldID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, mdb.DeleteAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin field: %v", id)
	}

	return nil
}

func (d Database) GetAdminField(id types.AdminFieldID) (*AdminFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, mdb.GetAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

func (d Database) ListAdminFields() (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateAdminField(s UpdateAdminFieldParams) (*string, error) {
	params := d.MapUpdateAdminFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminField(a mdbm.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbm.CreateAdminFieldParams {
	return mdbm.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbm.UpdateAdminFieldParams {
	return mdbm.UpdateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		AdminFieldID: a.AdminFieldID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateAdminField(s CreateAdminFieldParams) AdminFields {
	params := d.MapCreateAdminFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminField: %v\n", err)
	}
	row, err := queries.GetLastAdminField(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted AdminField: %v\n", err)
	}
	return d.MapAdminField(row)
}

func (d MysqlDatabase) CreateAdminFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

func (d MysqlDatabase) DeleteAdminField(id types.AdminFieldID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, mdbm.DeleteAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin field: %v", id)
	}

	return nil
}

func (d MysqlDatabase) GetAdminField(id types.AdminFieldID) (*AdminFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, mdbm.GetAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

func (d MysqlDatabase) ListAdminFields() (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateAdminField(s UpdateAdminFieldParams) (*string, error) {
	params := d.MapUpdateAdminFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateAdminField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapAdminField(a mdbp.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbp.CreateAdminFieldParams {
	return mdbp.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbp.UpdateAdminFieldParams {
	return mdbp.UpdateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		AdminFieldID: a.AdminFieldID,
	}
}

// QUERIES

func (d PsqlDatabase) CountAdminFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateAdminField(s CreateAdminFieldParams) AdminFields {
	params := d.MapCreateAdminFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminField  %v \n", err)
	}

	return d.MapAdminField(row)
}

func (d PsqlDatabase) CreateAdminFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

func (d PsqlDatabase) DeleteAdminField(id types.AdminFieldID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, mdbp.DeleteAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin field: %v", id)
	}

	return nil
}

func (d PsqlDatabase) GetAdminField(id types.AdminFieldID) (*AdminFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, mdbp.GetAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

func (d PsqlDatabase) ListAdminFields() (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateAdminField(s UpdateAdminFieldParams) (*string, error) {
	params := d.MapUpdateAdminFieldParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateAdminField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}
