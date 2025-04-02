package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/db-mysql"
	mdbp "github.com/hegner123/modulacms/db-psql"
	mdb "github.com/hegner123/modulacms/db-sqlite"
)

// /////////////////////////////
// STRUCTS
// ////////////////////////////
type AdminFields struct {
	AdminFieldID int64          `json:"admin_field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type CreateAdminFieldParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}
type ListAdminFieldByRouteIdRow struct {
	AdminFieldID int64          `json:"admin_field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	History      sql.NullString `json:"history"`
}
type ListAdminFieldsByDatatypeIDRow struct {
	AdminFieldID int64          `json:"admin_field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	History      sql.NullString `json:"history"`
}

type UpdateAdminFieldParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
	AdminFieldID int64          `json:"admin_field_id"`
}
type UtilityGetAdminfieldsRow struct {
	AdminFieldID int64  `json:"admin_field_id"`
	Label        string `json:"label"`
}
type AdminFieldsHistoryEntry struct {
	AdminFieldID int64          `json:"admin_field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}
type CreateAdminFieldFormParams struct {
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}
type UpdateAdminFieldFormParams struct {
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
	AdminFieldID string `json:"admin_field_id"`
}
type CreateAdminFieldParamsJSON struct {
	ParentID     NullInt64  `json:"parent_id"`
	Label        string     `json:"label"`
	Data         string     `json:"data"`
	Type         string     `json:"type"`
	AuthorID     int64      `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
}
type UpdateAdminFieldParamsJSON struct {
	ParentID     NullInt64  `json:"parent_id"`
	Label        string     `json:"label"`
	Data         string     `json:"data"`
	Type         string     `json:"type"`
	AuthorID     int64      `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
	AdminFieldID int64      `json:"admin_field_id"`
}

// /////////////////////////////
// GENERIC
// ////////////////////////////
func MapCreateAdminFieldParams(a CreateAdminFieldFormParams) CreateAdminFieldParams {
	return CreateAdminFieldParams{
		ParentID:     SNi64(a.ParentID),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     Si(a.AuthorID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
	}
}
func MapUpdateAdminFieldParams(a UpdateAdminFieldFormParams) UpdateAdminFieldParams {
	return UpdateAdminFieldParams{
		ParentID:     SNi64(a.ParentID),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     Si(a.AuthorID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
		AdminFieldID: Si(a.AdminFieldID),
	}
}
func MapStringAdminField(a AdminFields) StringAdminFields {
	return StringAdminFields{
		AdminFieldID: strconv.FormatInt(a.AdminFieldID, 10),
		ParentID:     strconv.FormatInt(a.ParentID.Int64, 10),
		Label:        AssertString(a.Label),
		Data:         AssertString(a.Data),
		Type:         AssertString(a.Type),
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
		History:      a.History.String,
	}
}
func MapCreateAdminFieldJSONParams(a CreateAdminFieldParamsJSON) CreateAdminFieldParams {
	return CreateAdminFieldParams{
		ParentID:     a.ParentID.NullInt64,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
		History:      a.History.NullString,
	}
}
func MapUpdateAdminFieldJSONParams(a UpdateAdminFieldParamsJSON) UpdateAdminFieldParams {
	return UpdateAdminFieldParams{
		ParentID:     a.ParentID.NullInt64,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
		History:      a.History.NullString,
		AdminFieldID: a.AdminFieldID,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
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
		History:      a.History,
	}
}
func (d Database) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdb.CreateAdminFieldParams {
	return mdb.CreateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
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
		History:      a.History,
		AdminFieldID: a.AdminFieldID,
	}
}

///QUERIES

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

func (d Database) DeleteAdminField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Field: %v ", id)
	}

	return nil
}

func (d Database) GetAdminField(id int64) (*AdminFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, id)
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
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapAdminField(a mdbm.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: int64(a.AdminFieldID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(a.DateCreated.String()),
		DateModified: Ns(a.DateModified.String()),
		History:      a.History,
	}
}
func (d MysqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbm.CreateAdminFieldParams {
	return mdbm.CreateAdminFieldParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         AssertString(a.Data),
		Type:         AssertString(a.Type),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		History:      a.History,
	}
}
func (d MysqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbm.UpdateAdminFieldParams {
	return mdbm.UpdateAdminFieldParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         AssertString(a.Data),
		Type:         AssertString(a.Type),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		History:      a.History,
		AdminFieldID: int32(a.AdminFieldID),
	}
}

///QUERIES

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

func (d MysqlDatabase) DeleteAdminField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Field: %v ", id)
	}

	return nil
}

func (d MysqlDatabase) GetAdminField(id int64) (*AdminFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, int32(id))
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
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapAdminField(a mdbp.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: int64(a.AdminFieldID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(Nt(a.DateCreated)),
		DateModified: Ns(Nt(a.DateModified)),
		History:      a.History,
	}
}
func (d PsqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbp.CreateAdminFieldParams {
	return mdbp.CreateAdminFieldParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         AssertString(a.Data),
		Type:         AssertString(a.Type),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
	}
}
func (d PsqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbp.UpdateAdminFieldParams {
	return mdbp.UpdateAdminFieldParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         AssertString(a.Data),
		Type:         AssertString(a.Type),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
		AdminFieldID: int32(a.AdminFieldID),
	}
}

///QUERIES

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

func (d PsqlDatabase) DeleteAdminField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteAdminField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Admin Field: %v ", id)
	}

	return nil
}
func (d PsqlDatabase) GetAdminField(id int64) (*AdminFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, int32(id))
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
