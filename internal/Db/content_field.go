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
	RouteID        sql.NullInt64  `json:"route_id"`
	ContentDataID    int64          `json:"content_data_id"`
	FieldID          int64          `json:"field_id"`
	FieldValue       string         `json:"field_value"`
	AuthorID         int64          `json:"author_id"`
	DateCreated      sql.NullString `json:"date_created"`
	DateModified     sql.NullString `json:"date_modified"`
	History          sql.NullString `json:"history"`
	ContentFieldID_2 int64          `json:"content_field_id_2"`
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

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateContentFieldParams(a CreateContentFieldFormParams) CreateContentFieldParams {
	return CreateContentFieldParams{
		ContentFieldID: Si(a.ContentFieldID),
		RouteID:        Ni64(Si(a.RouteID)),
		ContentDataID:  Si(a.ContentDataID),
		FieldID:        Si(a.FieldID),
		FieldValue:     a.FieldValue,
		AuthorID:       Si(a.AuthorID),
		DateCreated:    Ns(a.DateCreated),
		DateModified:   Ns(a.DateModified),
		History:        Ns(a.History),
	}
}

func MapUpdateContentFieldParams(a UpdateContentFieldFormParams) UpdateContentFieldParams {
	return UpdateContentFieldParams{
		ContentFieldID:   Si(a.ContentFieldID),
		RouteID:          Ni64(Si(a.RouteID)),
		ContentDataID:    Si(a.ContentDataID),
		FieldID:          Si(a.FieldID),
		FieldValue:       a.FieldValue,
		AuthorID:         Si(a.AuthorID),
		DateCreated:      Ns(a.DateCreated),
		DateModified:     Ns(a.DateModified),
		History:          Ns(a.History),
		ContentFieldID_2: Si(a.ContentFieldID_2),
	}
}

func MapStringContentField(a ContentFields) StringContentFields {
	return StringContentFields{
		ContentFieldID: strconv.FormatInt(a.ContentFieldID, 10),
		RouteID:        strconv.FormatInt(a.RouteID.Int64, 10),
		ContentDataID:  strconv.FormatInt(a.ContentDataID, 10),
		FieldID:        strconv.FormatInt(a.FieldID, 10),
		FieldValue:     a.FieldValue,
		AuthorID:       strconv.FormatInt(a.AuthorID, 10),
		DateCreated:    a.DateCreated.String,
		DateModified:   a.DateModified.String,
		History:        a.History.String,
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

func (d Database) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdb.UpdateContentFieldParams {
	return mdb.UpdateContentFieldParams{
		ContentFieldID:   a.ContentFieldID,
		RouteID:          a.RouteID,
		ContentDataID:    a.ContentDataID,
		FieldID:          a.FieldID,
		FieldValue:       a.FieldValue,
		AuthorID:         a.AuthorID,
		DateCreated:      a.DateCreated,
		DateModified:     a.DateModified,
		History:          a.History,
		ContentFieldID_2: a.ContentFieldID_2,
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
	rows, err := queries.ListContentFieldsByRoute(d.Context, Ni64(routeID))
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
		RouteID:        Ni64(int64(a.RouteID.Int32)),
		ContentDataID:  int64(a.ContentDataID),
		FieldID:        int64(a.FieldID),
		FieldValue:     a.FieldValue,
		AuthorID:       int64(a.AuthorID),
		DateCreated:    Ns(a.DateCreated.String()),
		DateModified:   Ns(a.DateModified.String()),
		History:        a.History,
	}
}

func (d MysqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbm.CreateContentFieldParams {
	return mdbm.CreateContentFieldParams{
		RouteID:       Ni32(a.RouteID.Int64),
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
		RouteID:        Ni32(a.RouteID.Int64),
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
	rows, err := queries.ListContentFieldsByRoute(d.Context, Ni32(routeID))
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
		RouteID:        Ni64(int64(a.RouteID.Int32)),
		ContentDataID:  int64(a.ContentDataID),
		FieldID:        int64(a.FieldID),
		FieldValue:     a.FieldValue,
		AuthorID:       int64(a.AuthorID),
		DateCreated:    Ns(Nt(a.DateCreated)),
		DateModified:   Ns(Nt(a.DateModified)),
		History:        a.History,
	}
}

func (d PsqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbp.CreateContentFieldParams {
	return mdbp.CreateContentFieldParams{
		ContentFieldID: int32(a.ContentFieldID),
		RouteID:        Ni32(a.RouteID.Int64),
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
		RouteID:          Ni32(a.RouteID.Int64),
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
	rows, err := queries.ListContentFieldsByRoute(d.Context, Ni32(routeID))
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
