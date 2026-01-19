package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/utility"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type AdminContentData struct {
	AdminContentDataID int64          `json:"admin_content_data_id"`
	ParentID           sql.NullInt64  `json:"parent_id"`
	AdminRouteID       int64          `json:"admin_route_id"`
	AdminDatatypeID    int64          `json:"admin_datatype_id"`
	AuthorID           int64          `json:"author_id"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
	History            sql.NullString `json:"history"`
}
type CreateAdminContentDataParams struct {
	ParentID        sql.NullInt64  `json:"parent_id"`
	AdminRouteID    int64          `json:"admin_route_id"`
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	AuthorID        int64          `json:"author_id"`
	DateCreated     sql.NullString `json:"date_created"`
	DateModified    sql.NullString `json:"date_modified"`
	History         sql.NullString `json:"history"`
}
type UpdateAdminContentDataParams struct {
	ParentID           sql.NullInt64  `json:"parent_id"`
	AdminRouteID       int64          `json:"admin_route_id"`
	AdminDatatypeID    int64          `json:"admin_datatype_id"`
	AuthorID           int64          `json:"author_id"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
	History            sql.NullString `json:"history"`
	AdminContentDataID int64          `json:"admin_content_data_id"`
}
type AdminContentDataHistoryEntry struct {
	AdminContentDataID int64          `json:"admin_content_data_id"`
	ParentID           sql.NullInt64  `json:"parent_id"`
	AdminRouteID       int64          `json:"admin_route_id"`
	AdminDatatypeID    int64          `json:"admin_datatype_id"`
	AuthorID           int64          `json:"author_id"`
	DateCreated        sql.NullString `json:"date_created"`
	DateModified       sql.NullString `json:"date_modified"`
}
type CreateAdminContentDataFormParams struct {
	ParentID        string `json:"parent_id"`
	AdminRouteID    string `json:"admin_route_id"`
	AdminDatatypeID string `json:"admin_datatype_id"`
	AuthorID        string `json:"author_id"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
	History         string `json:"history"`
}
type UpdateAdminContentDataFormParams struct {
	ParentID           string `json:"parent_id"`
	AdminRouteID       string `json:"admin_route_id"`
	AdminDatatypeID    string `json:"admin_datatype_id"`
	AuthorID           string `json:"author_id"`
	DateCreated        string `json:"date_created"`
	DateModified       string `json:"date_modified"`
	History            string `json:"history"`
	AdminContentDataID string `json:"admin_content_data_id"`
}
type AdminContentDataJSON struct {
	AdminContentDataID int64      `json:"admin_content_data_id"`
	ParentID           NullInt64  `json:"parent_id"`
	AdminRouteID       int64      `json:"admin_route_id"`
	AdminDatatypeID    int64      `json:"admin_datatype_id"`
	AuthorID           int64      `json:"author_id"`
	DateCreated        NullString `json:"date_created"`
	DateModified       NullString `json:"date_modified"`
	History            NullString `json:"history"`
}
type CreateAdminContentDataParamsJSON struct {
	ParentID        NullInt64  `json:"parent_id"`
	AdminRouteID    int64      `json:"admin_route_id"`
	AdminDatatypeID int64      `json:"admin_datatype_id"`
	AuthorID        int64      `json:"author_id"`
	DateCreated     NullString `json:"date_created"`
	DateModified    NullString `json:"date_modified"`
	History         NullString `json:"history"`
}
type UpdateAdminContentDataParamsJSON struct {
	ParentID           NullInt64  `json:"parent_id"`
	AdminRouteID       int64      `json:"admin_route_id"`
	AdminDatatypeID    int64      `json:"admin_datatype_id"`
	AuthorID           int64      `json:"author_id"`
	DateCreated        NullString `json:"date_created"`
	DateModified       NullString `json:"date_modified"`
	History            NullString `json:"history"`
	AdminContentDataID int64      `json:"admin_content_data_id"`
}

// /////////////////////////////
// GENERIC
// ////////////////////////////

func MapCreateAdminContentDataParams(a CreateAdminContentDataFormParams) CreateAdminContentDataParams {
	return CreateAdminContentDataParams{
		ParentID:        StringToNullInt64(a.ParentID),
		AdminRouteID:    StringToInt64(a.AdminRouteID),
		AdminDatatypeID: StringToInt64(a.AdminDatatypeID),
		AuthorID:        StringToInt64(a.AuthorID),
		DateCreated:     StringToNullString(a.DateCreated),
		DateModified:    StringToNullString(a.DateModified),
		History:         StringToNullString(a.History),
	}
}
func MapUpdateAdminContentDataParams(a UpdateAdminContentDataFormParams) UpdateAdminContentDataParams {
	return UpdateAdminContentDataParams{
		ParentID:           StringToNullInt64(a.ParentID),
		AdminRouteID:       StringToInt64(a.AdminRouteID),
		AdminDatatypeID:    StringToInt64(a.AdminDatatypeID),
		AuthorID:           StringToInt64(a.AuthorID),
		DateCreated:        StringToNullString(a.DateCreated),
		DateModified:       StringToNullString(a.DateModified),
		History:            StringToNullString(a.History),
		AdminContentDataID: StringToInt64(a.AdminContentDataID),
	}
}
func MapStringAdminContentData(a AdminContentData) StringAdminContentData {
	return StringAdminContentData{
		AdminContentDataID: strconv.FormatInt(a.AdminContentDataID, 10),
		ParentID:           utility.NullToString(a.ParentID),
		AdminRouteID:       strconv.FormatInt(a.AdminRouteID, 10),
		AdminDatatypeID:    strconv.FormatInt(a.AdminDatatypeID, 10),
		AuthorID:           strconv.FormatInt(a.AuthorID, 10),
		DateCreated:        utility.NullToString(a.DateCreated),
		DateModified:       utility.NullToString(a.DateModified),
		History:            utility.NullToString(a.History),
	}
}

func CreateAdminContentJSONParams(a CreateAdminContentDataParamsJSON) CreateAdminContentDataParams {
	return CreateAdminContentDataParams{
		ParentID:        a.ParentID.NullInt64,
		AdminRouteID:    a.AdminRouteID,
		AdminDatatypeID: a.AdminDatatypeID,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated.NullString,
		DateModified:    a.DateModified.NullString,
		History:         a.History.NullString,
	}
}

func UpdateAdminContentDataJSONParams(a UpdateAdminContentDataParamsJSON) UpdateAdminContentDataParams {
	return UpdateAdminContentDataParams{
		ParentID:        a.ParentID.NullInt64,
		AdminRouteID:    a.AdminRouteID,
		AdminDatatypeID: a.AdminDatatypeID,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated.NullString,
		DateModified:    a.DateModified.NullString,
		History:         a.History.NullString,
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapAdminContentData(a mdb.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		History:            a.History,
	}
}

func (d Database) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdb.CreateAdminContentDataParams {
	return mdb.CreateAdminContentDataParams{
		ParentID:        a.ParentID,
		AdminRouteID:    a.AdminRouteID,
		AdminDatatypeID: a.AdminDatatypeID,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		History:         a.History,
	}
}

func (d Database) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdb.UpdateAdminContentDataParams {
	return mdb.UpdateAdminContentDataParams{
		ParentID:           a.ParentID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		History:            a.History,
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
		return fmt.Errorf("failed to delete admin content data: %v\n ", id)
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
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
		ParentID:           NullInt32ToNullInt64(a.ParentID),
		AdminContentDataID: int64(a.AdminContentDataID),
		AdminRouteID:       int64(a.AdminRouteID),
		AdminDatatypeID:    int64(a.AdminDatatypeID),
		AuthorID:           int64(a.AuthorID),
		DateCreated:        StringToNullString(a.DateCreated.String()),
		DateModified:       StringToNullString(a.DateModified.String()),
		History:            a.History,
	}
}
func (d MysqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbm.CreateAdminContentDataParams {
	return mdbm.CreateAdminContentDataParams{
		ParentID:        NullInt64ToNullInt32(a.ParentID),
		AdminRouteID:    int32(a.AdminRouteID),
		AdminDatatypeID: int32(a.AdminDatatypeID),
		AuthorID:        int32(a.AuthorID),
		DateCreated:     StringToNTime(a.DateCreated.String).Time,
		DateModified:    StringToNTime(a.DateModified.String).Time,
		History:         a.History,
	}
}
func (d MysqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbm.UpdateAdminContentDataParams {
	return mdbm.UpdateAdminContentDataParams{
		ParentID:           NullInt64ToNullInt32(a.ParentID),
		AdminRouteID:       int32(a.AdminRouteID),
		AdminDatatypeID:    int32(a.AdminDatatypeID),
		AuthorID:           int32(a.AuthorID),
		DateCreated:        StringToNTime(a.DateCreated.String).Time,
		DateModified:       StringToNTime(a.DateModified.String).Time,
		History:            a.History,
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
		return fmt.Errorf("failed to delete admin content data: %v ", id)
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
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
	rows, err := queries.ListAdminContentDataByRoute(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
		ParentID:           NullInt32ToNullInt64(a.ParentID),
		AdminRouteID:       int64(a.AdminRouteID),
		AdminDatatypeID:    int64(a.AdminDatatypeID),
		AuthorID:           int64(a.AuthorID),
		DateCreated:        StringToNullString(NullTimeToString(a.DateCreated)),
		DateModified:       StringToNullString(NullTimeToString(a.DateModified)),
		History:            a.History,
	}
}
func (d PsqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbp.CreateAdminContentDataParams {
	return mdbp.CreateAdminContentDataParams{
		ParentID:        NullInt64ToNullInt32(a.ParentID),
		AdminRouteID:    int32(a.AdminRouteID),
		AdminDatatypeID: int32(a.AdminDatatypeID),
		AuthorID:        int32(a.AuthorID),
		DateCreated:     StringToNTime(a.DateCreated.String),
		DateModified:    StringToNTime(a.DateModified.String),
		History:         a.History,
	}
}
func (d PsqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbp.UpdateAdminContentDataParams {
	return mdbp.UpdateAdminContentDataParams{
		ParentID:           NullInt64ToNullInt32(a.ParentID),
		AdminRouteID:       int32(a.AdminRouteID),
		AdminDatatypeID:    int32(a.AdminDatatypeID),
		AuthorID:           int32(a.AuthorID),
		DateCreated:        StringToNTime(a.DateCreated.String),
		DateModified:       StringToNTime(a.DateModified.String),
		History:            a.History,
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
		return fmt.Errorf("failed to delete admin content data: %v ", id)
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
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
	rows, err := queries.ListAdminContentDataByRoute(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get datatypes: %v", err)
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
