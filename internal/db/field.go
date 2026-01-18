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
//STRUCTS
//////////////////////////////

type Fields struct {
	FieldID      int64          `json:"field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}
type CreateFieldParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}
type UpdateFieldParams struct {
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        string         `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
	FieldID      int64          `json:"field_id"`
}
type FieldsHistoryEntry struct {
	FieldID      int64          `json:"field_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	Label        any            `json:"label"`
	Data         string         `json:"data"`
	Type         string         `json:"type"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}
type CreateFieldFormParams struct {
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}
type UpdateFieldFormParams struct {
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
	FieldID      string `json:"field_id"`
}
type FieldsJSON struct {
	FieldID      int64      `json:"field_id"`
	ParentID     NullInt64  `json:"parent_id"`
	Label        any        `json:"label"`
	Data         string     `json:"data"`
	Type         string     `json:"type"`
	AuthorID     int64      `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
}
type CreateFieldParamsJSON struct {
	ParentID     NullInt64  `json:"parent_id"`
	Label        string     `json:"label"`
	Data         string     `json:"data"`
	Type         string     `json:"type"`
	AuthorID     int64      `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
}
type UpdateFieldParamsJSON struct {
	ParentID     NullInt64  `json:"parent_id"`
	Label        string     `json:"label"`
	Data         string     `json:"data"`
	Type         string     `json:"type"`
	AuthorID     int64      `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
	FieldID      int64      `json:"field_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateFieldParams(a CreateFieldFormParams) CreateFieldParams {
	return CreateFieldParams{
		ParentID:     StringToNullInt64(a.ParentID),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
		History:      StringToNullString(a.History),
	}
}

func MapUpdateFieldParams(a UpdateFieldFormParams) UpdateFieldParams {
	return UpdateFieldParams{
		ParentID:     StringToNullInt64(a.ParentID),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
		History:      StringToNullString(a.History),
		FieldID:      StringToInt64(a.FieldID),
	}
}
func MapFieldJSON(a Fields) FieldsJSON {
	return FieldsJSON{
		FieldID: a.FieldID,
		ParentID: NullInt64{
			NullInt64: a.ParentID,
		},
		Label:    a.Label,
		Data:     a.Data,
		Type:     a.Type,
		AuthorID: a.AuthorID,
		DateCreated: NullString{
			NullString: a.DateCreated,
		},
		DateModified: NullString{
			NullString: a.DateModified,
		},
		History: NullString{
			NullString: a.History,
		},
	}
}
func MapCreateFieldJSONParams(a CreateFieldParamsJSON) CreateFieldParams {
	return CreateFieldParams{
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

func MapUpdateFieldJSONParams(a UpdateFieldParamsJSON) UpdateFieldParams {
	return UpdateFieldParams{
		ParentID:     a.ParentID.NullInt64,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
		History:      a.History.NullString,
		FieldID:      a.FieldID,
	}
}
func MapStringField(a Fields) StringFields {
	return StringFields{
		FieldID:      strconv.FormatInt(a.FieldID, 10),
		ParentID:     utility.NullToString(a.ParentID),
		Label:        AssertString(a.Label),
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  utility.NullToString(a.DateCreated),
		DateModified: utility.NullToString(a.DateModified),
		History:      utility.NullToString(a.History),
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
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
		History:      a.History,
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
		History:      a.History,
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
		History:      a.History,
		FieldID:      a.FieldID,
	}
}

///QUERIES

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
func (d Database) DeleteField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteField(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Field: %v ", id)
	}

	return nil
}

func (d Database) GetField(id int64) (*Fields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetField(d.Context, id)
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

func (d Database) ListFieldsByDatatypeID(id int64) (*[]Fields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListFieldByDatatypeID(d.Context, Int64ToNullInt64(id))
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
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapField(a mdbm.Fields) Fields {
	return Fields{
		FieldID:      int64(a.FieldID),
		ParentID:     Int64ToNullInt64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated.String()),
		DateModified: StringToNullString(a.DateModified.String()),
		History:      a.History,
	}
}
func (d MysqlDatabase) MapCreateFieldParams(a CreateFieldParams) mdbm.CreateFieldParams {
	return mdbm.CreateFieldParams{
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		History:      a.History,
	}
}
func (d MysqlDatabase) MapUpdateFieldParams(a UpdateFieldParams) mdbm.UpdateFieldParams {
	return mdbm.UpdateFieldParams{
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		History:      a.History,
		FieldID:      int32(a.FieldID),
	}
}

///QUERIES

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
func (d MysqlDatabase) DeleteField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Field: %v ", id)
	}

	return nil
}
func (d MysqlDatabase) GetField(id int64) (*Fields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetField(d.Context, int32(id))
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
func (d MysqlDatabase) ListFieldsByDatatypeID(id int64) (*[]Fields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListFieldByDatatypeID(d.Context, Int64ToNullInt32(id))
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

// /MAPS
func (d PsqlDatabase) MapField(a mdbp.Fields) Fields {
	return Fields{
		FieldID:      int64(a.FieldID),
		ParentID:     Int64ToNullInt64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  StringToNullString(NullTimeToString(a.DateCreated)),
		DateModified: StringToNullString(NullTimeToString(a.DateModified)),
		History:      a.History,
	}
}
func (d PsqlDatabase) MapCreateFieldParams(a CreateFieldParams) mdbp.CreateFieldParams {
	return mdbp.CreateFieldParams{
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
	}
}
func (d PsqlDatabase) MapUpdateFieldParams(a UpdateFieldParams) mdbp.UpdateFieldParams {
	return mdbp.UpdateFieldParams{
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         a.Data,
		Type:         a.Type,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
		FieldID:      int32(a.FieldID),
	}
}

// /QUERIES
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
func (d PsqlDatabase) DeleteField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Field: %v ", id)
	}

	return nil
}
func (d PsqlDatabase) GetField(id int64) (*Fields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetField(d.Context, int32(id))
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
func (d PsqlDatabase) ListFieldsByDatatypeID(id int64) (*[]Fields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListFieldByDatatypeID(d.Context, Int64ToNullInt32(id))
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
