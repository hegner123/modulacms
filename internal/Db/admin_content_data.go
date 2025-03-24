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
// STRUCTS
//////////////////////////////

// ----------------------------
// STRUCTS
// ---------------------------
type AdminContentData struct {
	AdminContentDataID int64          `json:"admin_content_data_id"`
	AdminRouteID       int64          `json:"admin_route_id"`
	AdminDatatypeID    int64          `json:"admin_datatype_id"`
	History            sql.NullString `json:"history"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
}
type CreateAdminContentDataParams struct {
	AdminRouteID    int64          `json:"admin_route_id"`
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	History         sql.NullString `json:"history"`
	DateCreated     sql.NullString `json:"date_created"`
	DateModified    sql.NullString `json:"date_modified"`
}
type UpdateAdminContentDataParams struct {
	AdminRouteID       int64          `json:"admin_route_id"`
	AdminDatatypeID    int64          `json:"admin_datatype_id"`
	History            sql.NullString `json:"history"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
	AdminContentDataID int64          `json:"admin_content_data_id"`
}
type AdminContentDataHistoryEntry struct {
	AdminContentDataID int64          `json:"admin_content_data_id"`
	AdminRouteID       int64          `json:"admin_route_id"`
	AdminDatatypeID    int64          `json:"admin_datatype_id"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
}
type CreateAdminContentDataFormParams struct {
	AdminRouteID    string `json:"admin_route_id"`
	AdminDatatypeID string `json:"admin_datatype_id"`
	History         string `json:"history"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
}
type UpdateAdminContentDataFormParams struct {
	AdminRouteID       string `json:"admin_route_id"`
	AdminDatatypeID    string `json:"admin_datatype_id"`
	History            string `json:"history"`
	DateCreated        string `json:"date_created"`
	DateModified       string `json:"date_modified"`
	AdminContentDataID string `json:"admin_content_data_id"`
}

// /////////////////////////////
// GENERIC
// ////////////////////////////

func MapCreateAdminContentDataParams(a CreateAdminContentDataFormParams) CreateAdminContentDataParams {
	return CreateAdminContentDataParams{
		AdminRouteID:    Si(a.AdminRouteID),
		AdminDatatypeID: Si(a.AdminDatatypeID),
		History:         Ns(a.History),
		DateCreated:     Ns(a.DateCreated),
		DateModified:    Ns(a.DateModified),
	}
}
func MapUpdateAdminContentDataParams(a UpdateAdminContentDataFormParams) UpdateAdminContentDataParams {
	return UpdateAdminContentDataParams{
		AdminRouteID:       Si(a.AdminRouteID),
		AdminDatatypeID:    Si(a.AdminDatatypeID),
		History:            Ns(a.History),
		DateCreated:        Ns(a.DateCreated),
		DateModified:       Ns(a.DateModified),
		AdminContentDataID: Si(a.AdminContentDataID),
	}
}
func MapStringAdminContentData(a AdminContentData) StringAdminContentData {
	return StringAdminContentData{
		AdminContentDataID: strconv.FormatInt(a.AdminContentDataID, 10),
		AdminRouteID:       strconv.FormatInt(a.AdminRouteID, 10),
		AdminDatatypeID:    strconv.FormatInt(a.AdminDatatypeID, 10),
		History:            a.History.String,
		DateCreated:        a.DateCreated.String,
		DateModified:       a.DateModified.String,
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminContentData(a mdb.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: a.AdminContentDataID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		History:            a.History,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
}

func (d Database) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdb.CreateAdminContentDataParams {
	return mdb.CreateAdminContentDataParams{
		AdminRouteID:    a.AdminRouteID,
		AdminDatatypeID: a.AdminDatatypeID,
		History:         a.History,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

func (d Database) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdb.UpdateAdminContentDataParams {
	return mdb.UpdateAdminContentDataParams{
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		History:            a.History,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		AdminContentDataID: a.AdminContentDataID,
	}
}

// QUERIES

func (d Database) CountAdminContentData() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d Database) CreateAdminContentData(s CreateAdminContentDataParams) AdminContentData {
	params := d.MapCreateAdminContentDataParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateAdminContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminContentData(row)
}
func (d Database) CreateAdminContentDataTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}
func (d Database) DeleteAdminContentData(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminContentData(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Data: %v ", id)
	}

	return nil
}
func (d Database) GetAdminContentData(id int64) (*AdminContentData, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}
func (d Database) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListAdminContentDataByRoute(id int64) (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRoute(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) UpdateAdminContentData(s UpdateAdminContentDataParams) (*string, error) {
	params := d.MapUpdateAdminContentDataParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content Data id %v\n", s.AdminDatatypeID)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapAdminContentData(a mdbm.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: int64(a.AdminContentDataID),
		AdminRouteID:       int64(a.AdminRouteID.Int32),
		AdminDatatypeID:    int64(a.AdminDatatypeID.Int32),
		History:            a.History,
		DateCreated:        Ns(Nt(a.DateCreated)),
		DateModified:       Ns(Nt(a.DateModified)),
	}
}
func (d MysqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbm.CreateAdminContentDataParams {
	return mdbm.CreateAdminContentDataParams{
		AdminRouteID:    Ni32(a.AdminRouteID),
		AdminDatatypeID: Ni32(a.AdminDatatypeID),
		History:         a.History,
		DateCreated:     StringToNTime(a.DateCreated.String),
		DateModified:    StringToNTime(a.DateModified.String),
	}
}
func (d MysqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbm.UpdateAdminContentDataParams {
	return mdbm.UpdateAdminContentDataParams{
		AdminRouteID:       Ni32(a.AdminRouteID),
		AdminDatatypeID:    Ni32(a.AdminDatatypeID),
		History:            a.History,
		DateCreated:        StringToNTime(a.DateCreated.String),
		DateModified:       StringToNTime(a.DateModified.String),
		AdminContentDataID: int32(a.AdminContentDataID),
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminContentData() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CreateAdminContentData(s CreateAdminContentDataParams) AdminContentData {
	params := d.MapCreateAdminContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminContentData: %v\n", err)
	}
	row, err := queries.GetLastAdminContentData(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted AdminContentData: %v\n", err)
	}
	return d.MapAdminContentData(row)
}
func (d MysqlDatabase) CreateAdminContentDataTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}
func (d MysqlDatabase) DeleteAdminContentData(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Data: %v ", id)
	}

	return nil
}
func (d MysqlDatabase) GetAdminContentData(id int64) (*AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentDataByRoute(id int64) (*[]AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRoute(d.Context, Ni32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) UpdateAdminContentData(s UpdateAdminContentDataParams) (*string, error) {
	params := d.MapUpdateAdminContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateAdminContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content Data id %v\n", s.AdminDatatypeID)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS

func (d PsqlDatabase) MapAdminContentData(a mdbp.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: int64(a.AdminContentDataID),
		AdminRouteID:       int64(a.AdminRouteID.Int32),
		AdminDatatypeID:    int64(a.AdminDatatypeID.Int32),
		History:            a.History,
		DateCreated:        Ns(Nt(a.DateCreated)),
		DateModified:       Ns(Nt(a.DateModified)),
	}
}
func (d PsqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbp.CreateAdminContentDataParams {
	return mdbp.CreateAdminContentDataParams{
		AdminRouteID:    Ni32(a.AdminRouteID),
		AdminDatatypeID: Ni32(a.AdminDatatypeID),
		History:         a.History,
		DateCreated:     StringToNTime(a.DateCreated.String),
		DateModified:    StringToNTime(a.DateModified.String),
	}
}
func (d PsqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbp.UpdateAdminContentDataParams {
	return mdbp.UpdateAdminContentDataParams{
		AdminRouteID:       Ni32(a.AdminRouteID),
		AdminDatatypeID:    Ni32(a.AdminDatatypeID),
		History:            a.History,
		DateCreated:        StringToNTime(a.DateCreated.String),
		DateModified:       StringToNTime(a.DateModified.String),
		AdminContentDataID: int32(a.AdminContentDataID),
	}
}

// /QUERIES

func (d PsqlDatabase) CountAdminContentData() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CreateAdminContentData(s CreateAdminContentDataParams) AdminContentData {
	params := d.MapCreateAdminContentDataParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminContentData(row)
}
func (d PsqlDatabase) CreateAdminContentDataTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminContentDataTable(d.Context)
	return err
}
func (d PsqlDatabase) DeleteAdminContentData(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Content Data: %v ", id)
	}

	return nil
}
func (d PsqlDatabase) GetAdminContentData(id int64) (*AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentDataByRoute(id int64) (*[]AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentDataByRoute(d.Context, Ni32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) UpdateAdminContentData(s UpdateAdminContentDataParams) (*string, error) {
	params := d.MapUpdateAdminContentDataParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateAdminContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content Data id %v\n", s.AdminDatatypeID)
	return &u, nil
}
