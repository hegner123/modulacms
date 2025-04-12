package db

import (
	"database/sql"
	"fmt"
	"strconv"

	utility "github.com/hegner123/modulacms/internal/utility"
	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

// /////////////////////////////
// STRUCTS
// ////////////////////////////
type AdminContentFields struct {
	AdminContentFieldID int64          `json:"admin_content_field_id"`
	AdminRouteID        sql.NullInt64  `json:"admin_route_id"`
	AdminContentDataID  int64          `json:"admin_content_data_id"`
	AdminFieldID        int64          `json:"admin_field_id"`
	AdminFieldValue     string         `json:"admin_field_value"`
	AuthorID            int64          `json:"author_id"`
	DateCreated         sql.NullString `json:"date_created"`
	DateModified        sql.NullString `json:"date_modified"`
	History             sql.NullString `json:"history"`
}
type CreateAdminContentFieldParams struct {
	AdminRouteID       sql.NullInt64  `json:"admin_route_id"`
	AdminContentDataID int64          `json:"admin_content_data_id"`
	AdminFieldID       int64          `json:"admin_field_id"`
	AdminFieldValue    string         `json:"admin_field_value"`
	AuthorID           int64          `json:"author_id"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
	History            sql.NullString `json:"history"`
}
type UpdateAdminContentFieldParams struct {
	AdminRouteID        sql.NullInt64  `json:"admin_route_id"`
	AdminContentDataID  int64          `json:"admin_content_data_id"`
	AdminFieldID        int64          `json:"admin_field_id"`
	AdminFieldValue     string         `json:"admin_field_value"`
	AuthorID            int64          `json:"author_id"`
	DateCreated         sql.NullString `json:"date_created"`
	DateModified        sql.NullString `json:"date_modified"`
	History             sql.NullString `json:"history"`
	AdminContentFieldID int64          `json:"admin_content_field_id"`
}
type AdminContentFieldsHistoryEntry struct {
	AdminContentFieldID int64          `json:"admin_content_field_id"`
	AdminRouteID        sql.NullInt64  `json:"admin_route_id"`
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
	AuthorID            string `json:"author_id"`
	DateCreated         string `json:"date_created"`
	DateModified        string `json:"date_modified"`
	History             string `json:"history"`
}
type UpdateAdminContentFieldFormParams struct {
	AdminRouteID        string `json:"admin_route_id"`
	AdminContentDataID  string `json:"admin_content_data_id"`
	AdminFieldID        string `json:"admin_field_id"`
	AdminFieldValue     string `json:"admin_field_value"`
	AuthorID            string `json:"author_id"`
	DateCreated         string `json:"date_created"`
	DateModified        string `json:"date_modified"`
	History             string `json:"history"`
	AdminContentFieldID string `json:"admin_content_field_id"`
}
type AdminContentFieldsJSON struct {
	AdminContentFieldID int64      `json:"admin_content_field_id"`
	AdminRouteID        NullInt64  `json:"admin_route_id"`
	AdminContentDataID  int64      `json:"admin_content_data_id"`
	AdminFieldID        int64      `json:"admin_field_id"`
	AdminFieldValue     string     `json:"admin_field_value"`
	AuthorID            int64      `json:"author_id"`
	DateCreated         NullString `json:"date_created"`
	DateModified        NullString `json:"date_modified"`
	History             NullString `json:"history"`
}
type CreateAdminContentFieldParamsJSON struct {
	AdminRouteID       NullInt64  `json:"admin_route_id"`
	AdminContentDataID int64      `json:"admin_content_data_id"`
	AdminFieldID       int64      `json:"admin_field_id"`
	AdminFieldValue    string     `json:"admin_field_value"`
	AuthorID           int64      `json:"author_id"`
	DateCreated        NullString `json:"date_created"`
	DateModified       NullString `json:"date_modified"`
	History            NullString `json:"history"`
}
type UpdateAdminContentFieldParamsJSON struct {
	AdminRouteID        NullInt64  `json:"admin_route_id"`
	AdminContentDataID  int64      `json:"admin_content_data_id"`
	AdminFieldID        int64      `json:"admin_field_id"`
	AdminFieldValue     string     `json:"admin_field_value"`
	AuthorID            int64      `json:"author_id"`
	DateCreated         NullString `json:"date_created"`
	DateModified        NullString `json:"date_modified"`
	History             NullString `json:"history"`
	AdminContentFieldID int64      `json:"admin_content_field_id"`
}

// /////////////////////////////
// GENERIC
// ////////////////////////////
func MapCreateAdminContentFieldParams(a CreateAdminContentFieldFormParams) CreateAdminContentFieldParams {
	return CreateAdminContentFieldParams{
		AdminRouteID:       StringToNullInt64(a.AdminRouteID),
		AdminContentDataID: StringToInt64(a.AdminContentDataID),
		AdminFieldID:       StringToInt64(a.AdminFieldID),
		AdminFieldValue:    a.AdminFieldValue,
		AuthorID:           StringToInt64(a.AuthorID),
		DateCreated:        StringToNullString(a.DateCreated),
		DateModified:       StringToNullString(a.DateModified),
		History:            StringToNullString(a.History),
	}
}
func MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldFormParams) UpdateAdminContentFieldParams {
	return UpdateAdminContentFieldParams{
		AdminRouteID:        Int64ToNullInt64(StringToInt64(a.AdminRouteID)),
		AdminContentDataID:  StringToInt64(a.AdminContentDataID),
		AdminFieldID:        StringToInt64(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            StringToInt64(a.AuthorID),
		DateCreated:         StringToNullString(a.DateCreated),
		DateModified:        StringToNullString(a.DateModified),
		History:             StringToNullString(a.History),
        AdminContentFieldID: StringToInt64(a.AdminContentFieldID),
	}
}
func MapStringAdminContentField(a AdminContentFields) StringAdminContentFields {
	return StringAdminContentFields{
		AdminContentFieldID: strconv.FormatInt(a.AdminContentFieldID, 10),
		AdminRouteID:        ReadNullInt64(a.AdminRouteID),
		AdminContentDataID:  strconv.FormatInt(a.AdminContentDataID, 10),
		AdminFieldID:        strconv.FormatInt(a.AdminFieldID, 10),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            strconv.FormatInt(a.AuthorID, 10),
		DateCreated:         ReadNullString(a.DateCreated),
		DateModified:        ReadNullString(a.DateModified),
		History:             ReadNullString(a.History),
	}
}
func MapAdminContentFieldJSON(a AdminContentFieldsJSON) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID.NullInt64,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated.NullString,
		DateModified:        a.DateModified.NullString,
		History:             a.History.NullString,
	}
}
func MapUpdateAdminContentFieldJSONParams(a UpdateAdminContentFieldParamsJSON) UpdateAdminContentFieldParams {
	return UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID.NullInt64,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated.NullString,
		DateModified:        a.DateModified.NullString,
		History:             a.History.NullString,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}
func MapCreateAdminContentFieldJSONParams(a CreateAdminContentFieldParamsJSON) CreateAdminContentFieldParams {
	return CreateAdminContentFieldParams{
		AdminRouteID:       a.AdminRouteID.NullInt64,
		AdminContentDataID: a.AdminContentDataID,
		AdminFieldID:       a.AdminFieldID,
		AdminFieldValue:    a.AdminFieldValue,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated.NullString,
		DateModified:       a.DateModified.NullString,
		History:            a.History.NullString,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
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
		History:             a.History,
	}
}
func (d Database) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdb.CreateAdminContentFieldParams {
	return mdb.CreateAdminContentFieldParams{
		AdminRouteID:       a.AdminRouteID,
		AdminContentDataID: a.AdminContentDataID,
		AdminFieldID:       a.AdminFieldID,
		AdminFieldValue:    a.AdminFieldValue,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		History:            a.History,
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
		History:             a.History,
        AdminContentFieldID: a.AdminContentFieldID,
	}
}

// /QUERIES
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
func (d Database) DeleteAdminContentField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminContentField(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete admin content field: %v ", id)
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
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, Int64ToNullInt64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
    utility.DefaultLogger.Fdebug("",s)
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

// /MAPS
func (d MysqlDatabase) MapAdminContentField(a mdbm.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminRouteID:        Int64ToNullInt64(int64(a.AdminRouteID.Int32)),
		AdminContentFieldID: int64(a.AdminContentFieldID),
		AdminContentDataID:  int64(a.AdminContentDataID),
		AdminFieldID:        int64(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            int64(a.AuthorID),
		DateCreated:         StringToNullString(a.DateCreated.String()),
		DateModified:        StringToNullString(a.DateModified.String()),
		History:             a.History,
	}
}
func (d MysqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbm.CreateAdminContentFieldParams {
	return mdbm.CreateAdminContentFieldParams{
		AdminRouteID:       Int64ToNullInt32(a.AdminRouteID.Int64),
		AdminContentDataID: int32(a.AdminContentDataID),
		AdminFieldID:       int32(a.AdminFieldID),
		AdminFieldValue:    a.AdminFieldValue,
		AuthorID:           int32(a.AuthorID),
		DateCreated:        NullStringToTime(a.DateCreated),
		DateModified:       NullStringToTime(a.DateModified),
		History:            a.History,
	}
}
func (d MysqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbm.UpdateAdminContentFieldParams {
	return mdbm.UpdateAdminContentFieldParams{
		AdminRouteID:        Int64ToNullInt32(a.AdminRouteID.Int64),
		AdminContentDataID:  int32(a.AdminContentDataID),
		AdminFieldID:        int32(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            int32(a.AuthorID),
		DateCreated:         NullStringToTime(a.DateCreated),
		DateModified:        NullStringToTime(a.DateModified),
		History:             a.History,
		AdminContentFieldID: int32(a.AdminContentFieldID),
	}
}

// /QUERIES
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
		return fmt.Errorf("failed to delete admin content field: %v ", id)
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
		return nil, fmt.Errorf("failed to get content fields: %v", err)
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
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, Int64ToNullInt32(id))
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
		AdminContentFieldID: int64(a.AdminContentFieldID),
		AdminRouteID:        NullInt32ToNullInt64(a.AdminRouteID),
		AdminContentDataID:  int64(a.AdminContentDataID),
		AdminFieldID:        int64(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            int64(a.AuthorID),
		DateCreated:         StringToNullString(NullTimeToString(a.DateCreated)),
		DateModified:        StringToNullString(NullTimeToString(a.DateModified)),
		History:             a.History,
	}
}
func (d PsqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbp.CreateAdminContentFieldParams {
	return mdbp.CreateAdminContentFieldParams{
		AdminRouteID:       Int64ToNullInt32(a.AdminRouteID.Int64),
		AdminContentDataID: int32(a.AdminContentDataID),
		AdminFieldID:       int32(a.AdminFieldID),
		AdminFieldValue:    a.AdminFieldValue,
		AuthorID:           int32(a.AuthorID),
		DateCreated:        NullStringToNullTime(a.DateCreated),
		DateModified:       NullStringToNullTime(a.DateModified),
		History:            a.History,
	}
}
func (d PsqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbp.UpdateAdminContentFieldParams {
	return mdbp.UpdateAdminContentFieldParams{
		AdminRouteID:        Int64ToNullInt32(a.AdminRouteID.Int64),
		AdminContentDataID:  int32(a.AdminContentDataID),
		AdminFieldID:        int32(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            int32(a.AuthorID),
		DateCreated:         NullStringToNullTime(a.DateCreated),
		DateModified:        NullStringToNullTime(a.DateModified),
		History:             a.History,
		AdminContentFieldID: int32(a.AdminContentFieldID),
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
		return fmt.Errorf("failed to delete admin content field: %v ", id)
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
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, Int64ToNullInt32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
