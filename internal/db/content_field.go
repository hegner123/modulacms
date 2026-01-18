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


// /////////////////////////////
// STRUCTS
// ////////////////////////////
type ContentFields struct {
	ContentFieldID int64          `json:"content_field_id"`
	RouteID        sql.NullInt64  `json:"route_id"`
	ContentDataID  int64          `json:"content_data_id"`
	FieldID        int64          `json:"field_id"`
	FieldValue     string         `json:"field_value"`
	AuthorID       int64          `json:"author_id"`
	DateCreated    sql.NullString `json:"date_created"`
	DateModified   sql.NullString `json:"date_modified"`
	History        sql.NullString `json:"history"`
}

type CreateContentFieldParams struct {
	ContentFieldID int64          `json:"content_field_id"`
	RouteID        sql.NullInt64  `json:"route_id"`
	ContentDataID  int64          `json:"content_data_id"`
	FieldID        int64          `json:"field_id"`
	FieldValue     string         `json:"field_value"`
	AuthorID       int64          `json:"author_id"`
	DateCreated    sql.NullString `json:"date_created"`
	DateModified   sql.NullString `json:"date_modified"`
	History        sql.NullString `json:"history"`
}

type UpdateContentFieldParams struct {
	ContentFieldID   int64          `json:"content_field_id"`
	RouteID          sql.NullInt64  `json:"route_id"`
	ContentDataID    int64          `json:"content_data_id"`
	FieldID          int64          `json:"field_id"`
	FieldValue       string         `json:"field_value"`
	AuthorID         int64          `json:"author_id"`
	DateCreated      sql.NullString `json:"date_created"`
	DateModified     sql.NullString `json:"date_modified"`
	History          sql.NullString `json:"history"`
	ContentFieldID_2 int64          `json:"content_field_id_2"`
}
type CreateContentFieldParamsJSON struct {
	ContentFieldID int64      `json:"content_field_id"`
	RouteID        NullInt64  `json:"route_id"`
	ContentDataID  int64      `json:"content_data_id"`
	FieldID        int64      `json:"field_id"`
	FieldValue     string     `json:"field_value"`
	AuthorID       int64      `json:"author_id"`
	DateCreated    NullString `json:"date_created"`
	DateModified   NullString `json:"date_modified"`
	History        NullString `json:"history"`
}

type UpdateContentFieldParamsJSON struct {
	ContentFieldID   int64      `json:"content_field_id"`
	RouteID          NullInt64  `json:"route_id"`
	ContentDataID    int64      `json:"content_data_id"`
	FieldID          int64      `json:"field_id"`
	FieldValue       string     `json:"field_value"`
	AuthorID         int64      `json:"author_id"`
	DateCreated      NullString `json:"date_created"`
	DateModified     NullString `json:"date_modified"`
	History          NullString `json:"history"`
	ContentFieldID_2 int64      `json:"content_field_id_2"`
}

type ContentFieldsHistoryEntry struct {
	ContentFieldID int64          `json:"content_field_id"`
	RouteID        sql.NullInt64  `json:"route_id"`
	ContentDataID  int64          `json:"content_data_id"`
	FieldID        int64          `json:"field_id"`
	FieldValue     string         `json:"field_value"`
	AuthorID       int64          `json:"author_id"`
	DateCreated    sql.NullString `json:"date_created"`
	DateModified   sql.NullString `json:"date_modified"`
}

type CreateContentFieldFormParams struct {
	RouteID        string `json:"route_id"`
	ContentFieldID string `json:"content_field_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	FieldValue     string `json:"field_value"`
	AuthorID       string `json:"author_id"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
	History        string `json:"history"`
}

type UpdateContentFieldFormParams struct {
	RouteID          string `json:"route_id"`
	ContentFieldID   string `json:"content_field_id"`
	ContentDataID    string `json:"content_data_id"`
	FieldID          string `json:"field_id"`
	FieldValue       string `json:"field_value"`
	AuthorID         string `json:"author_id"`
	DateCreated      string `json:"date_created"`
	DateModified     string `json:"date_modified"`
	History          string `json:"history"`
	ContentFieldID_2 string `json:"content_field_id_2"`
}
type ContentFieldsJSON struct {
	ContentFieldID int64      `json:"content_field_id"`
	RouteID        NullInt64  `json:"route_id"`
	ContentDataID  int64      `json:"content_data_id"`
	FieldID        int64      `json:"field_id"`
	FieldValue     string     `json:"field_value"`
	AuthorID       int64      `json:"author_id"`
	DateCreated    NullString `json:"date_created"`
	DateModified   NullString `json:"date_modified"`
	History        NullString `json:"history"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateContentFieldParams(a CreateContentFieldFormParams) CreateContentFieldParams {
	return CreateContentFieldParams{
		ContentFieldID: StringToInt64(a.ContentFieldID),
		RouteID:        Int64ToNullInt64(StringToInt64(a.RouteID)),
		ContentDataID:  StringToInt64(a.ContentDataID),
		FieldID:        StringToInt64(a.FieldID),
		FieldValue:     a.FieldValue,
		AuthorID:       StringToInt64(a.AuthorID),
		DateCreated:    StringToNullString(a.DateCreated),
		DateModified:   StringToNullString(a.DateModified),
		History:        StringToNullString(a.History),
	}
}

func MapUpdateContentFieldParams(a UpdateContentFieldFormParams) UpdateContentFieldParams {
	return UpdateContentFieldParams{
		ContentFieldID:   StringToInt64(a.ContentFieldID),
		RouteID:          Int64ToNullInt64(StringToInt64(a.RouteID)),
		ContentDataID:    StringToInt64(a.ContentDataID),
		FieldID:          StringToInt64(a.FieldID),
		FieldValue:       a.FieldValue,
		AuthorID:         StringToInt64(a.AuthorID),
		DateCreated:      StringToNullString(a.DateCreated),
		DateModified:     StringToNullString(a.DateModified),
		History:          StringToNullString(a.History),
		ContentFieldID_2: StringToInt64(a.ContentFieldID_2),
	}
}
func MapCreateContentFieldJSONParams(a CreateContentFieldParamsJSON) CreateContentFieldParams {
	return CreateContentFieldParams{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID.NullInt64,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated.NullString,
		DateModified:   a.DateModified.NullString,
		History:        a.History.NullString,
	}
}

func MapUpdateContentFieldJSONParams(a UpdateContentFieldParamsJSON) UpdateContentFieldParams {
	return UpdateContentFieldParams{
		RouteID:        a.RouteID.NullInt64,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated.NullString,
		DateModified:   a.DateModified.NullString,
		History:        a.History.NullString,
		ContentFieldID: a.ContentFieldID,
	}
}
func MapContentFieldJSON(a ContentFields) ContentFieldsJSON {
	return ContentFieldsJSON{
		ContentFieldID: a.ContentFieldID,
		RouteID:        NullInt64{a.RouteID},
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    NullString{a.DateCreated},
		DateModified:   NullString{a.DateModified},
		History:        NullString{a.History},
	}
}

func MapStringContentField(a ContentFields) StringContentFields {
	return StringContentFields{
		ContentFieldID: strconv.FormatInt(a.ContentFieldID, 10),
		RouteID:        utility.NullToString(a.RouteID),
		ContentDataID:  strconv.FormatInt(a.ContentDataID, 10),
		FieldID:        strconv.FormatInt(a.FieldID, 10),
		FieldValue:     a.FieldValue,
		AuthorID:       strconv.FormatInt(a.AuthorID, 10),
		DateCreated:    utility.NullToString(a.DateCreated),
		DateModified:   utility.NullToString(a.DateModified),
		History:        utility.NullToString(a.History),
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// MAPS
func (d Database) MapContentField(a mdb.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
		History:        a.History,
	}
}

func (d Database) MapCreateContentFieldParams(a CreateContentFieldParams) mdb.CreateContentFieldParams {
	return mdb.CreateContentFieldParams{
		RouteID:       a.RouteID,
		ContentDataID: a.ContentDataID,
		FieldID:       a.FieldID,
		FieldValue:    a.FieldValue,
		AuthorID:      a.AuthorID,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
		History:       a.History,
	}
}

func (d Database) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdb.UpdateContentFieldParams {
	return mdb.UpdateContentFieldParams{
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
		History:        a.History,
		ContentFieldID: a.ContentFieldID,
	}
}

// QUERIES
func (d Database) CountContentFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateContentFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateContentFieldTable(d.Context)
	return err
}

func (d Database) CreateContentField(s CreateContentFieldParams) ContentFields {
	params := d.MapCreateContentFieldParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateContentField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentField: %v\n", err)
	}
	return d.MapContentField(row)
}

func (d Database) DeleteContentField(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteContentField(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete ContentField: %v ", id)
	}
	return nil
}

func (d Database) GetContentField(id int64) (*ContentFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentField(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapContentField(row)
	return &res, nil
}

func (d Database) ListContentFields() (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields: %v\n", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentFieldsByRoute(routeID int64) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, Int64ToNullInt64(routeID))
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route: %v\n", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateContentField(s UpdateContentFieldParams) (*string, error) {
	params := d.MapUpdateContentFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateContentField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field %v\n", s.ContentFieldID)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

// MAPS
func (d MysqlDatabase) MapContentField(a mdbm.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: int64(a.ContentFieldID),
		RouteID:        Int64ToNullInt64(int64(a.RouteID.Int32)),
		ContentDataID:  int64(a.ContentDataID),
		FieldID:        int64(a.FieldID),
		FieldValue:     a.FieldValue,
		AuthorID:       int64(a.AuthorID),
		DateCreated:    StringToNullString(a.DateCreated.String()),
		DateModified:   StringToNullString(a.DateModified.String()),
		History:        a.History,
	}
}

func (d MysqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbm.CreateContentFieldParams {
	return mdbm.CreateContentFieldParams{
		RouteID:       Int64ToNullInt32(a.RouteID.Int64),
		ContentDataID: int32(a.ContentDataID),
		FieldID:       int32(a.FieldID),
		FieldValue:    a.FieldValue,
		AuthorID:      int32(a.AuthorID),
		History:       a.History,
	}
}

func (d MysqlDatabase) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdbm.UpdateContentFieldParams {
	return mdbm.UpdateContentFieldParams{
		ContentFieldID: int32(a.ContentFieldID),
		RouteID:        Int64ToNullInt32(a.RouteID.Int64),
		ContentDataID:  int32(a.ContentDataID),
		FieldID:        int32(a.FieldID),
		FieldValue:     a.FieldValue,
		AuthorID:       int32(a.AuthorID),
		History:        a.History,
	}
}

// QUERIES
func (d MysqlDatabase) CountContentFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateContentFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentFieldTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateContentField(s CreateContentFieldParams) ContentFields {
	params := d.MapCreateContentFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentField: %v\n", err)
	}
	row, err := queries.GetLastContentField(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted ContentField: %v\n", err)
	}
	return d.MapContentField(row)
}

func (d MysqlDatabase) DeleteContentField(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete ContentField: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetContentField(id int64) (*ContentFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentField(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapContentField(row)
	return &res, nil
}

func (d MysqlDatabase) ListContentFields() (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields: %v\n", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentFieldsByRoute(routeID int64) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, Int64ToNullInt32(routeID))
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route: %v\n", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateContentField(s UpdateContentFieldParams) (*string, error) {
	params := d.MapUpdateContentFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateContentField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field %v\n", s.ContentFieldID)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// MAPS
func (d PsqlDatabase) MapContentField(a mdbp.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: int64(a.ContentFieldID),
		RouteID:        Int64ToNullInt64(int64(a.RouteID.Int32)),
		ContentDataID:  int64(a.ContentDataID),
		FieldID:        int64(a.FieldID),
		FieldValue:     a.FieldValue,
		AuthorID:       int64(a.AuthorID),
		DateCreated:    StringToNullString(NullTimeToString(a.DateCreated)),
		DateModified:   StringToNullString(NullTimeToString(a.DateModified)),
		History:        a.History,
	}
}

func (d PsqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbp.CreateContentFieldParams {
	return mdbp.CreateContentFieldParams{
		ContentFieldID: int32(a.ContentFieldID),
		RouteID:        Int64ToNullInt32(a.RouteID.Int64),
		ContentDataID:  int32(a.ContentDataID),
		FieldID:        int32(a.FieldID),
		FieldValue:     a.FieldValue,
		AuthorID:       int32(a.AuthorID),
		DateCreated:    StringToNTime(a.DateCreated.String),
		DateModified:   StringToNTime(a.DateModified.String),
		History:        a.History,
	}
}

func (d PsqlDatabase) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdbp.UpdateContentFieldParams {
	return mdbp.UpdateContentFieldParams{
		ContentFieldID:   int32(a.ContentFieldID),
		RouteID:          Int64ToNullInt32(a.RouteID.Int64),
		ContentDataID:    int32(a.ContentDataID),
		FieldID:          int32(a.FieldID),
		FieldValue:       a.FieldValue,
		AuthorID:         int32(a.AuthorID),
		DateCreated:      StringToNTime(a.DateCreated.String),
		DateModified:     StringToNTime(a.DateModified.String),
		History:          a.History,
		ContentFieldID_2: int32(a.ContentFieldID_2),
	}
}

// QUERIES
func (d PsqlDatabase) CountContentFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateContentFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateContentFieldTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateContentField(s CreateContentFieldParams) ContentFields {
	params := d.MapCreateContentFieldParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateContentField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentField: %v\n", err)
	}
	return d.MapContentField(row)
}

func (d PsqlDatabase) DeleteContentField(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteContentField(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete ContentField: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetContentField(id int64) (*ContentFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentField(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapContentField(row)
	return &res, nil
}

func (d PsqlDatabase) ListContentFields() (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields: %v\n", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentFieldsByRoute(routeID int64) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, Int64ToNullInt32(routeID))
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route: %v\n", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateContentField(s UpdateContentFieldParams) (*string, error) {
	params := d.MapUpdateContentFieldParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateContentField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field %v\n", s.ContentFieldID)
	return &u, nil
}
