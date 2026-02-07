package db

import (
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type AdminContentFields struct {
	AdminContentFieldID types.AdminContentFieldID  `json:"admin_content_field_id"`
	AdminRouteID        sql.NullString             `json:"admin_route_id"`
	AdminContentDataID  string                     `json:"admin_content_data_id"`
	AdminFieldID        types.NullableAdminFieldID `json:"admin_field_id"`
	AdminFieldValue     string                     `json:"admin_field_value"`
	AuthorID            types.NullableUserID       `json:"author_id"`
	DateCreated         types.Timestamp            `json:"date_created"`
	DateModified        types.Timestamp            `json:"date_modified"`
}
type CreateAdminContentFieldParams struct {
	AdminRouteID       sql.NullString             `json:"admin_route_id"`
	AdminContentDataID string                     `json:"admin_content_data_id"`
	AdminFieldID       types.NullableAdminFieldID `json:"admin_field_id"`
	AdminFieldValue    string                     `json:"admin_field_value"`
	AuthorID           types.NullableUserID       `json:"author_id"`
	DateCreated        types.Timestamp            `json:"date_created"`
	DateModified       types.Timestamp            `json:"date_modified"`
}
type UpdateAdminContentFieldParams struct {
	AdminRouteID        sql.NullString             `json:"admin_route_id"`
	AdminContentDataID  string                     `json:"admin_content_data_id"`
	AdminFieldID        types.NullableAdminFieldID `json:"admin_field_id"`
	AdminFieldValue     string                     `json:"admin_field_value"`
	AuthorID            types.NullableUserID       `json:"author_id"`
	DateCreated         types.Timestamp            `json:"date_created"`
	DateModified        types.Timestamp            `json:"date_modified"`
	AdminContentFieldID types.AdminContentFieldID  `json:"admin_content_field_id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

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

// MapStringAdminContentField converts AdminContentFields to StringAdminContentFields for table display
func MapStringAdminContentField(a AdminContentFields) StringAdminContentFields {
	adminRouteID := ""
	if a.AdminRouteID.Valid {
		adminRouteID = a.AdminRouteID.String
	}
	return StringAdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID.String(),
		AdminRouteID:        adminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID.String(),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID.String(),
		DateCreated:         a.DateCreated.String(),
		DateModified:        a.DateModified.String(),
		History:             "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminContentField(a mdb.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}

func (d Database) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdb.CreateAdminContentFieldParams {
	return mdb.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}

func (d Database) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdb.UpdateAdminContentFieldParams {
	return mdb.UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}

// QUERIES

func (d Database) CountAdminContentFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d Database) CreateAdminContentField(s CreateAdminContentFieldParams) AdminContentFields {
	params := d.MapCreateAdminContentFieldParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateAdminContentField(d.Context, params)
	if err != nil {
		fmt.Printf("failed to create admin content field  %v \n", err)
	}
	return d.MapAdminContentField(row)
}
func (d Database) CreateAdminContentFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d Database) DeleteAdminContentField(id types.AdminContentFieldID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminContentField(d.Context, mdb.DeleteAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin content field: %v ", id)
	}
	return nil
}
func (d Database) GetAdminContentField(id types.AdminContentFieldID) (*AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, mdb.GetAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}
func (d Database) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListAdminContentFieldsByRoute(id string) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, mdb.ListAdminContentFieldsByRouteParams{AdminRouteID: StringToNullString(id)})
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) UpdateAdminContentField(s UpdateAdminContentFieldParams) (*string, error) {
	params := d.MapUpdateAdminContentFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminContentField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field id %v\n", s.AdminContentFieldID)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminContentField(a mdbm.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d MysqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbm.CreateAdminContentFieldParams {
	return mdbm.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d MysqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbm.UpdateAdminContentFieldParams {
	return mdbm.UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminContentFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CreateAdminContentField(s CreateAdminContentFieldParams) AdminContentFields {
	params := d.MapCreateAdminContentFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminContentField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminContentField: %v\n", err)
	}
	row, err := queries.GetAdminContentField(d.Context, mdbm.GetAdminContentFieldParams{AdminContentFieldID: params.AdminContentFieldID})
	if err != nil {
		fmt.Printf("Failed to get last inserted AdminContentField: %v\n", err)
	}
	return d.MapAdminContentField(row)
}

func (d MysqlDatabase) CreateAdminContentFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d MysqlDatabase) DeleteAdminContentField(id types.AdminContentFieldID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminContentField(d.Context, mdbm.DeleteAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin content field: %v ", id)
	}

	return nil
}
func (d MysqlDatabase) GetAdminContentField(id types.AdminContentFieldID) (*AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, mdbm.GetAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentFieldsByRoute(id string) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, mdbm.ListAdminContentFieldsByRouteParams{AdminRouteID: StringToNullString(id)})
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) UpdateAdminContentField(s UpdateAdminContentFieldParams) (*string, error) {
	params := d.MapUpdateAdminContentFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateAdminContentField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field id %v\n", s.AdminContentFieldID)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS

func (d PsqlDatabase) MapAdminContentField(a mdbp.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d PsqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbp.CreateAdminContentFieldParams {
	return mdbp.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d PsqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbp.UpdateAdminContentFieldParams {
	return mdbp.UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}

// /QUERIES

func (d PsqlDatabase) CountAdminContentFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CreateAdminContentField(s CreateAdminContentFieldParams) AdminContentFields {
	params := d.MapCreateAdminContentFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminContentField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminContentField  %v \n", err)
	}

	return d.MapAdminContentField(row)
}
func (d PsqlDatabase) CreateAdminContentFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d PsqlDatabase) DeleteAdminContentField(id types.AdminContentFieldID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminContentField(d.Context, mdbp.DeleteAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete admin content field: %v ", id)
	}

	return nil
}
func (d PsqlDatabase) GetAdminContentField(id types.AdminContentFieldID) (*AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, mdbp.GetAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentFieldsByRoute(id string) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, mdbp.ListAdminContentFieldsByRouteParams{AdminRouteID: StringToNullString(id)})
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) UpdateAdminContentField(s UpdateAdminContentFieldParams) (*string, error) {
	params := d.MapUpdateAdminContentFieldParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateAdminContentField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field id %v\n", s.AdminContentFieldID)
	return &u, nil
}
