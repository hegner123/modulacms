package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

// /////////////////////////////
// STRUCTS
// ////////////////////////////
type AdminDatatypes struct {
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	ParentID        sql.NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	AuthorID        int64          `json:"author_id"`
	DateCreated     sql.NullString `json:"date_created"`
	DateModified    sql.NullString `json:"date_modified"`
	History         sql.NullString `json:"history"`
}
type CreateAdminDatatypeParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}
type ListAdminDatatypeByRouteIdRow struct {
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	AdminRouteID    sql.NullInt64  `json:"admin_route_id"`
	ParentID        sql.NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	History         sql.NullString `json:"history"`
}
type UpdateAdminDatatypeParams struct {
	ParentID        sql.NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	AuthorID        int64          `json:"author_id"`
	DateCreated     sql.NullString `json:"date_created"`
	DateModified    sql.NullString `json:"date_modified"`
	History         sql.NullString `json:"history"`
	AdminDatatypeID int64          `json:"admin_datatype_id"`
}
type UtilityGetAdminDatatypesRow struct {
	AdminDatatypeID int64  `json:"admin_datatype_id"`
	Label           string `json:"label"`
}
type AdminDatatypesHistoryEntry struct {
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	ParentID        sql.NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	AuthorID        int64          `json:"author_id"`
	DateCreated     sql.NullString `json:"date_created"`
	DateModified    sql.NullString `json:"date_modified"`
}
type CreateAdminDatatypeFormParams struct {
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}
type UpdateAdminDatatypeFormParams struct {
	ParentID        string `json:"parent_id"`
	Label           string `json:"label"`
	Type            string `json:"type"`
	AuthorID        string `json:"author_id"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
	History         string `json:"history"`
	AdminDatatypeID string `json:"admin_datatype_id"`
}
type AdminDatatypesJSON struct {
	AdminDatatypeID int64          `json:"admin_datatype_id"`
	ParentID        NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	AuthorID        int64          `json:"author_id"`
	DateCreated     NullString `json:"date_created"`
	DateModified    NullString `json:"date_modified"`
	History         NullString `json:"history"`
}
type CreateAdminDatatypeParamsJSON struct {
	ParentID     NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
}
type UpdateAdminDatatypeParamsJSON struct {
	ParentID        NullInt64  `json:"parent_id"`
	Label           string         `json:"label"`
	Type            string         `json:"type"`
	AuthorID        int64          `json:"author_id"`
	DateCreated     NullString `json:"date_created"`
	DateModified    NullString `json:"date_modified"`
	History         NullString `json:"history"`
	AdminDatatypeID int64          `json:"admin_datatype_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateAdminDatatypeParams(a CreateAdminDatatypeFormParams) CreateAdminDatatypeParams {
	return CreateAdminDatatypeParams{
		ParentID:     StringToNullInt64(a.ParentID),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
		History:      StringToNullString(a.History),
	}
}
func MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeFormParams) UpdateAdminDatatypeParams {
	return UpdateAdminDatatypeParams{
		ParentID:        StringToNullInt64(a.ParentID),
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        StringToInt64(a.AuthorID),
		DateCreated:     StringToNullString(a.DateCreated),
		DateModified:    StringToNullString(a.DateModified),
		History:         StringToNullString(a.History),
		AdminDatatypeID: StringToInt64(a.AdminDatatypeID),
	}
}
func MapStringAdminDatatype(a AdminDatatypes) StringAdminDatatypes {
	return StringAdminDatatypes{
		AdminDatatypeID: strconv.FormatInt(a.AdminDatatypeID, 10),
		ParentID:        ReadNullInt64(a.ParentID),
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        strconv.FormatInt(a.AuthorID, 10),
		DateCreated:     ReadNullString(a.DateCreated),
		DateModified:    ReadNullString(a.DateModified),
		History:         ReadNullString(a.History),
	}
}
func MapCreateAdminDatatypeJSONParams(a CreateAdminDatatypeParamsJSON) CreateAdminDatatypeParams {
	return CreateAdminDatatypeParams{
		ParentID:     a.ParentID.NullInt64,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
		History:      a.History.NullString,
	}
}
func MapUpdateAdminDatatypeJSONParams(a UpdateAdminDatatypeParamsJSON) UpdateAdminDatatypeParams {
	return UpdateAdminDatatypeParams{
		ParentID:        a.ParentID.NullInt64,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated.NullString,
		DateModified:    a.DateModified.NullString,
		History:         a.History.NullString,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
func (d Database) MapAdminDatatype(a mdb.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		History:         a.History,
	}
}
func (d Database) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdb.CreateAdminDatatypeParams {
	return mdb.CreateAdminDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}
func (d Database) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdb.UpdateAdminDatatypeParams {
	return mdb.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		History:         a.History,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// /QUERIES
func (d Database) CountAdminDatatypes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateAdminDatatype(s CreateAdminDatatypeParams) AdminDatatypes {
	params := d.MapCreateAdminDatatypeParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateAdminDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminDatatype(row)
}
func (d Database) CreateAdminDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d Database) DeleteAdminDatatype(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to delete admin datatype: %v ", id)
	}

	return nil
}

func (d Database) GetAdminDatatypeById(id int64) (*AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d Database) ListAdminDatatypeGlobalId() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeGlobal(d.Context)
	if err != nil {
		return nil, err
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateAdminDatatype(s UpdateAdminDatatypeParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin datatype, %v ", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapAdminDatatype(a mdbm.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: int64(a.AdminDatatypeID),
		ParentID:        Int64ToNullInt64(int64(a.ParentID.Int32)),
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        int64(a.AuthorID),
		DateCreated:     StringToNullString(a.DateCreated.String()),
		DateModified:    StringToNullString(a.DateModified.String()),
		History:         a.History,
	}
}
func (d MysqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbm.CreateAdminDatatypeParams {
	return mdbm.CreateAdminDatatypeParams{
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		History:      a.History,
	}
}
func (d MysqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbm.UpdateAdminDatatypeParams {
	return mdbm.UpdateAdminDatatypeParams{
		ParentID:        Int64ToNullInt32(a.ParentID.Int64),
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        int32(a.AuthorID),
		DateCreated:     StringToNTime(a.DateCreated.String).Time,
		DateModified:    StringToNTime(a.DateModified.String).Time,
		History:         a.History,
		AdminDatatypeID: int32(a.AdminDatatypeID),
	}
}

// /QUERIES
func (d MysqlDatabase) CountAdminDatatypes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateAdminDatatype(s CreateAdminDatatypeParams) AdminDatatypes {
	params := d.MapCreateAdminDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype: %v\n", err)
	}
	row, err := queries.GetLastAdminDatatype(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted AdminDatatype: %v\n", err)
	}
	return d.MapAdminDatatype(row)
}

func (d MysqlDatabase) CreateAdminDatatypeTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d MysqlDatabase) DeleteAdminDatatype(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete admin datatype: %v ", id)
	}

	return nil
}


func (d MysqlDatabase) GetAdminDatatypeById(id int64) (*AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}
func (d MysqlDatabase) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) UpdateAdminDatatype(s UpdateAdminDatatypeParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateAdminDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin datatype, %v ", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapAdminDatatype(a mdbp.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: int64(a.AdminDatatypeID),
		ParentID:        Int64ToNullInt64(int64(a.ParentID.Int32)),
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        int64(a.AuthorID),
		DateCreated:     StringToNullString(NullTimeToString(a.DateCreated)),
		DateModified:    StringToNullString(NullTimeToString(a.DateModified)),
		History:         a.History,
	}
}
func (d PsqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbp.CreateAdminDatatypeParams {
	return mdbp.CreateAdminDatatypeParams{
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
	}
}
func (d PsqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbp.UpdateAdminDatatypeParams {
	return mdbp.UpdateAdminDatatypeParams{
		ParentID:        Int64ToNullInt32(a.ParentID.Int64),
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        int32(a.AuthorID),
		DateCreated:     StringToNTime(a.DateCreated.String),
		DateModified:    StringToNTime(a.DateModified.String),
		History:         a.History,
		AdminDatatypeID: int32(a.AdminDatatypeID),
	}
}

// /QUERIES
func (d PsqlDatabase) CountAdminDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateAdminDatatype(s CreateAdminDatatypeParams) AdminDatatypes {
	params := d.MapCreateAdminDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateAdminDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminDatatype  %v \n", err)
	}

	return d.MapAdminDatatype(row)
}

func (d PsqlDatabase) CreateAdminDatatypeTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

func (d PsqlDatabase) DeleteAdminDatatype(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminDatatype(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to delete admin datatype: %v ", id)
	}

	return nil
}
/*
func (d PsqlDatabase) GetAdminDatatypeGlobalId() (*AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetGlobalAdminDatatypeId(d.Context)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}
*/

func (d PsqlDatabase) GetAdminDatatypeById(id int64) (*AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d PsqlDatabase) ListAdminDatatypes() (*[]AdminDatatypes, error) {

	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateAdminDatatype(s UpdateAdminDatatypeParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateAdminDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin datatype, %v ", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}
