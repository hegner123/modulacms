package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/db-mysql"
	mdbp "github.com/hegner123/modulacms/db-psql"
	mdb "github.com/hegner123/modulacms/db-sqlite"
)

///////////////////////////////
//STRUCTS
//////////////////////////////
type AdminContentFields struct {
	AdminContentFieldID int64          `json:"admin_content_field_id"`
	AdminRouteID        int64          `json:"admin_route_id"`
	AdminContentDataID  int64          `json:"admin_content_data_id"`
	AdminFieldID        int64          `json:"admin_field_id"`
	AdminFieldValue     string         `json:"admin_field_value"`
	History             sql.NullString `json:"history"`
	DateCreated         sql.NullString `json:"date_created"`
	DateModified        sql.NullString `json:"date_modified"`
}
type CreateAdminContentFieldParams struct {
	AdminContentFieldID int64          `json:"admin_content_field_id"`
	AdminRouteID        int64          `json:"admin_route_id"`
	AdminContentDataID  int64          `json:"admin_content_data_id"`
	AdminFieldID        int64          `json:"admin_field_id"`
	AdminFieldValue     string         `json:"admin_field_value"`
	History             sql.NullString `json:"history"`
	DateCreated         sql.NullString `json:"date_created"`
	DateModified        sql.NullString `json:"date_modified"`
}
type UpdateAdminContentFieldParams struct {
	AdminContentFieldID   int64          `json:"admin_content_field_id"`
	AdminRouteID          int64          `json:"admin_route_id"`
	AdminContentDataID    int64          `json:"admin_content_data_id"`
	AdminFieldID          int64          `json:"admin_field_id"`
	AdminFieldValue       string         `json:"admin_field_value"`
	History               sql.NullString `json:"history"`
	DateCreated           sql.NullString `json:"date_created"`
	DateModified          sql.NullString `json:"date_modified"`
	AdminContentFieldID_2 int64          `json:"admin_content_field_id_2"`
}
type AdminContentFieldsHistoryEntry struct {
	AdminContentFieldID int64          `json:"admin_content_field_id"`
	AdminRouteID        int64          `json:"admin_route_id"`
	AdminContentDataID  int64          `json:"admin_content_data_id"`
	AdminFieldID        int64          `json:"admin_field_id"`
	AdminFieldValue     string         `json:"admin_field_value"`
	DateCreated         sql.NullString `json:"date_created"`
	DateModified        sql.NullString `json:"date_modified"`
}
type CreateAdminContentFieldFormParams struct {
	AdminRouteID        string `json:"admin_route_id"`
	AdminContentFieldID string `json:"admin_content_field_id"`
	AdminContentDataID  string `json:"admin_content_data_id"`
	AdminFieldID        string `json:"admin_field_id"`
	AdminFieldValue     string `json:"admin_field_value"`
	History             string `json:"history"`
	DateCreated         string `json:"date_created"`
	DateModified        string `json:"date_modified"`
}
type UpdateAdminContentFieldFormParams struct {
	AdminRouteID          string `json:"admin_route_id"`
	AdminContentFieldID   string `json:"content_field_id"`
	AdminContentDataID    string `json:"content_data_id"`
	AdminFieldID          string `json:"admin_field_id"`
	AdminFieldValue       string `json:"admin_field_value"`
	History               string `json:"history"`
	DateCreated           string `json:"date_created"`
	DateModified          string `json:"date_modified"`
	AdminContentFieldID_2 string `json:"admin_content_field_id_2"`
}

///////////////////////////////
//GENERIC
//////////////////////////////
func MapCreateAdminContentFieldParams(a CreateAdminContentFieldFormParams) CreateAdminContentFieldParams {
	return CreateAdminContentFieldParams{
		AdminRouteID:        Si(a.AdminRouteID),
		AdminContentFieldID: Si(a.AdminContentFieldID),
		AdminContentDataID:  Si(a.AdminContentDataID),
		AdminFieldID:        Si(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		History:             Ns(a.History),
		DateCreated:         Ns(a.DateCreated),
		DateModified:        Ns(a.DateModified),
	}
}
func MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldFormParams) UpdateAdminContentFieldParams {
	return UpdateAdminContentFieldParams{
		AdminRouteID:          Si(a.AdminRouteID),
		AdminContentFieldID:   Si(a.AdminContentFieldID),
		AdminContentDataID:    Si(a.AdminContentDataID),
		AdminFieldID:          Si(a.AdminFieldID),
		AdminFieldValue:       a.AdminFieldValue,
		History:               Ns(a.History),
		DateCreated:           Ns(a.DateCreated),
		DateModified:          Ns(a.DateModified),
		AdminContentFieldID_2: Si(a.AdminContentFieldID_2),
	}
}
func MapStringAdminContentField(a AdminContentFields) StringAdminContentFields {
	return StringAdminContentFields{
		AdminContentFieldID: strconv.FormatInt(a.AdminContentFieldID, 10),
		AdminContentDataID:  strconv.FormatInt(a.AdminContentDataID, 10),
		AdminFieldID:        strconv.FormatInt(a.AdminFieldID, 10),
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History.String,
		DateCreated:         a.DateCreated.String,
		DateModified:        a.DateModified.String,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

///MAPS
func (d Database) MapAdminContentField(a mdb.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d Database) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdb.CreateAdminContentFieldParams {
	return mdb.CreateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentFieldID: a.AdminContentFieldID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
func (d Database) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdb.UpdateAdminContentFieldParams {
	return mdb.UpdateAdminContentFieldParams{
		AdminRouteID:          a.AdminRouteID,
		AdminContentFieldID:   a.AdminContentFieldID,
		AdminContentDataID:    a.AdminContentDataID,
		AdminFieldID:          a.AdminFieldID,
		AdminFieldValue:       a.AdminFieldValue,
		History:               a.History,
		DateCreated:           a.DateCreated,
		DateModified:          a.DateModified,
		AdminContentFieldID_2: a.AdminContentFieldID_2,
	}
}
///QUERIES
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
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminContentField(row)
}
func (d Database) CreateAdminContentFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d Database) DeleteAdminContentField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminContentField(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Field: %v ", id)
	}

	return nil
}
func (d Database) GetAdminContentField(id int64) (*AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, id)
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
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListAdminContentFieldsByRoute(id int64) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
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
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field id %v\n", s.AdminContentFieldID)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

///MAPS
func (d MysqlDatabase) MapAdminContentField(a mdbm.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminRouteID:        int64(a.AdminRouteID.Int32),
		AdminContentFieldID: int64(a.AdminContentFieldID),
		AdminContentDataID:  int64(a.AdminContentDataID),
		AdminFieldID:        int64(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History,
		DateCreated:         Ns(Nt(a.DateCreated)),
		DateModified:        Ns(Nt(a.DateModified)),
	}
}
func (d MysqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbm.CreateAdminContentFieldParams {
	return mdbm.CreateAdminContentFieldParams{
		AdminRouteID:        Ni32(a.AdminRouteID),
		AdminContentFieldID: int32(a.AdminContentFieldID),
		AdminContentDataID:  int32(a.AdminContentDataID),
		AdminFieldID:        int32(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History,
		DateCreated:         StringToNTime(a.DateCreated.String),
		DateModified:        StringToNTime(a.DateModified.String),
	}
}
func (d MysqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbm.UpdateAdminContentFieldParams {
	return mdbm.UpdateAdminContentFieldParams{
		AdminRouteID:          Ni32(a.AdminRouteID),
		AdminContentFieldID:   int32(a.AdminContentFieldID),
		AdminContentDataID:    int32(a.AdminContentDataID),
		AdminFieldID:          int32(a.AdminFieldID),
		AdminFieldValue:       a.AdminFieldValue,
		History:               a.History,
		DateCreated:           StringToNTime(a.DateCreated.String),
		DateModified:          StringToNTime(a.DateModified.String),
		AdminContentFieldID_2: int32(a.AdminContentFieldID_2),
	}
}
///QUERIES
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
	row, err := queries.GetLastAdminContentField(d.Context)
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
func (d MysqlDatabase) DeleteAdminContentField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Field: %v ", id)
	}

	return nil
}
func (d MysqlDatabase) GetAdminContentField(id int64) (*AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, int32(id))
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
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentFieldsByRoute(id int64) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, Ni32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
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
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field id %v\n", s.AdminContentFieldID)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

///MAPS
func (d PsqlDatabase) MapAdminContentField(a mdbp.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: int64(a.AdminContentFieldID),
		AdminContentDataID:  int64(a.AdminContentDataID),
		AdminFieldID:        int64(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History,
		DateCreated:         Ns(Nt(a.DateCreated)),
		DateModified:        Ns(Nt(a.DateModified)),
	}
}
func (d PsqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbp.CreateAdminContentFieldParams {
	return mdbp.CreateAdminContentFieldParams{
		AdminRouteID:        Ni32(a.AdminRouteID),
		AdminContentFieldID: int32(a.AdminContentFieldID),
		AdminContentDataID:  int32(a.AdminContentDataID),
		AdminFieldID:        int32(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History,
		DateCreated:         StringToNTime(a.DateCreated.String),
		DateModified:        StringToNTime(a.DateModified.String),
	}
}
func (d PsqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbp.UpdateAdminContentFieldParams {
	return mdbp.UpdateAdminContentFieldParams{
		AdminRouteID:          Ni32(a.AdminRouteID),
		AdminContentFieldID:   int32(a.AdminContentFieldID),
		AdminContentDataID:    int32(a.AdminContentDataID),
		AdminFieldID:          int32(a.AdminFieldID),
		AdminFieldValue:       a.AdminFieldValue,
		History:               a.History,
		DateCreated:           StringToNTime(a.DateCreated.String),
		DateModified:          StringToNTime(a.DateModified.String),
		AdminContentFieldID_2: int32(a.AdminContentFieldID_2),
	}
}

///QUERIES
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
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminContentField(row)
}
func (d PsqlDatabase) CreateAdminContentFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d PsqlDatabase) DeleteAdminContentField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Field: %v ", id)
	}

	return nil
}
func (d PsqlDatabase) GetAdminContentField(id int64) (*AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, int32(id))
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
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentFieldsByRoute(id int64) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, Ni32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
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
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field id %v\n", s.AdminContentFieldID)
	return &u, nil
}
