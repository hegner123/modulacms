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

type ContentData struct {
	ContentDataID int64          `json:"content_data_id"`
	ParentID      sql.NullInt64  `json:"parent_id"`
	RouteID       int64          `json:"route_id"`
	DatatypeID    int64          `json:"datatype_id"`
	AuthorID      int64          `json:"author_id"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
	History       sql.NullString `json:"history"`
}

type CreateContentDataParams struct {
	RouteID      int64          `json:"route_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	DatatypeID   int64          `json:"datatype_id"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	History      sql.NullString `json:"history"`
}

type UpdateContentDataParams struct {
	RouteID       int64          `json:"route_id"`
	ParentID      sql.NullInt64  `json:"parent_id"`
	DatatypeID    int64          `json:"datatype_id"`
	AuthorID      int64          `json:"author_id"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
	History       sql.NullString `json:"history"`
	ContentDataID int64          `json:"content_data_id"`
}

type ContentDataHistoryEntry struct {
	ContentDataID int64          `json:"content_data_id"`
	RouteID       int64          `json:"route_id"`
	ParentID      sql.NullInt64  `json:"parent_id"`
	DatatypeID    int64          `json:"datatype_id"`
	AuthorID      int64          `json:"author_id"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
}

type CreateContentDataFormParams struct {
	RouteID      string `json:"route_id"`
	ParentID     string `json:"parent_id"`
	DatatypeID   string `json:"datatype_id"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

type UpdateContentDataFormParams struct {
	RouteID       string `json:"route_id"`
	ParentID      string `json:"parent_id"`
	DatatypeID    string `json:"datatype_id"`
	AuthorID      string `json:"author_id"`
	DateCreated   string `json:"date_created"`
	DateModified  string `json:"date_modified"`
	History       string `json:"history"`
	ContentDataID string `json:"content_data_id"`
}

type ContentDataJSON struct {
	ContentDataID int64      `json:"content_data_id"`
	ParentID      NullInt64  `json:"parent_id"`
	RouteID       int64      `json:"route_id"`
	DatatypeID    int64      `json:"datatype_id"`
	AuthorID      int64      `json:"author_id"`
	DateCreated   NullString `json:"date_created"`
	DateModified  NullString `json:"date_modified"`
	History       NullString `json:"history"`
}
type CreateContentDataParamsJSON struct {
	RouteID      int64      `json:"route_id"`
	ParentID     NullInt64  `json:"parent_id"`
	DatatypeID   int64      `json:"datatype_id"`
	AuthorID     int64      `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	History      NullString `json:"history"`
}

type UpdateContentDataParamsJSON struct {
	RouteID       int64      `json:"route_id"`
	ParentID      NullInt64  `json:"parent_id"`
	DatatypeID    int64      `json:"datatype_id"`
	AuthorID      int64      `json:"author_id"`
	DateCreated   NullString `json:"date_created"`
	DateModified  NullString `json:"date_modified"`
	History       NullString `json:"history"`
	ContentDataID int64      `json:"content_data_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateContentDataParams(a CreateContentDataFormParams) CreateContentDataParams {
	return CreateContentDataParams{
		RouteID:      StringToInt64(a.RouteID),
		ParentID:     StringToNullInt64(a.ParentID),
		DatatypeID:   StringToInt64(a.DatatypeID),
		AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
		History:      StringToNullString(a.History),
	}
}

func MapUpdateContentDataParams(a UpdateContentDataFormParams) UpdateContentDataParams {
	return UpdateContentDataParams{
		RouteID:       StringToInt64(a.RouteID),
		ParentID:      StringToNullInt64(a.ParentID),
		DatatypeID:    StringToInt64(a.DatatypeID),
		AuthorID:      StringToInt64(a.AuthorID),
		DateCreated:   StringToNullString(a.DateCreated),
		DateModified:  StringToNullString(a.DateModified),
		History:       StringToNullString(a.History),
		ContentDataID: StringToInt64(a.ContentDataID),
	}
}

func MapContentDataJSON(a ContentData) ContentDataJSON {
	return ContentDataJSON{
		ContentDataID: a.ContentDataID,
		ParentID:      NullInt64{a.ParentID},
		RouteID:       a.RouteID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		DateCreated:   NullString{a.DateCreated},
		DateModified:  NullString{a.DateModified},
		History:       NullString{a.History},
	}
}

func MapCreateContentDataJSONParams(a CreateContentDataParamsJSON) CreateContentDataParams {
	return CreateContentDataParams{
		RouteID:      a.RouteID,
		ParentID:     a.ParentID.NullInt64,
		DatatypeID:   a.DatatypeID,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
		History:      a.History.NullString,
	}
}

func MapUpdateContentDataJSONParams(a UpdateContentDataParamsJSON) UpdateContentDataParams {
	return UpdateContentDataParams{
		RouteID:       a.RouteID,
		ParentID:      a.ParentID.NullInt64,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		DateCreated:   a.DateCreated.NullString,
		DateModified:  a.DateModified.NullString,
		History:       a.History.NullString,
		ContentDataID: a.ContentDataID,
	}
}

func MapStringContentData(a ContentData) StringContentData {
	return StringContentData{
		ContentDataID: strconv.FormatInt(a.ContentDataID, 10),
		RouteID:       strconv.FormatInt(a.RouteID, 10),
		ParentID:      strconv.FormatInt(a.ParentID.Int64, 10),
		DatatypeID:    strconv.FormatInt(a.DatatypeID, 10),
		AuthorID:      strconv.FormatInt(a.AuthorID, 10),
		DateCreated:   a.DateCreated.String,
		DateModified:  a.DateModified.String,
		History:       a.History.String,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
func (d Database) MapContentData(a mdb.ContentData) ContentData {
	return ContentData{
		ContentDataID: a.ContentDataID,
		RouteID:       a.RouteID,
		ParentID:      a.ParentID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
		History:       a.History,
	}
}

func (d Database) MapCreateContentDataParams(a CreateContentDataParams) mdb.CreateContentDataParams {
	return mdb.CreateContentDataParams{
		RouteID:      a.RouteID,
		ParentID:     a.ParentID,
		DatatypeID:   a.DatatypeID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}

func (d Database) MapUpdateContentDataParams(a UpdateContentDataParams) mdb.UpdateContentDataParams {
	return mdb.UpdateContentDataParams{
		RouteID:       a.RouteID,
		ParentID:      a.ParentID,
		DatatypeID:    a.DatatypeID,
		AuthorID:      a.AuthorID,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
		History:       a.History,
		ContentDataID: a.ContentDataID,
	}
}

// /QUERIES
func (d Database) CountContentData() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateContentDataTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

func (d Database) CreateContentData(s CreateContentDataParams) ContentData {
	params := d.MapCreateContentDataParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentData: %v\n", err)
	}
	return d.MapContentData(row)
}

func (d Database) DeleteContentData(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteContentData(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete ContentData: %v ", id)
	}
	return nil
}

func (d Database) GetContentData(id int64) (*ContentData, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentData(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

func (d Database) ListContentData() (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentDataByRoute(routeID int64) (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentDataByRoute(d.Context, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData by route: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateContentData(s UpdateContentDataParams) (*string, error) {
	params := d.MapUpdateContentDataParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content data %v\n", s.ContentDataID)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapContentData(a mdbm.ContentData) ContentData {
	return ContentData{
		ContentDataID: int64(a.ContentDataID),
		RouteID:       int64(a.RouteID.Int32),
		ParentID:      Int64ToNullInt64(int64(a.ParentID.Int32)),
		DatatypeID:    int64(a.DatatypeID.Int32),
		AuthorID:      int64(a.AuthorID),
		DateCreated:   StringToNullString(a.DateCreated.String()),
		DateModified:  StringToNullString(a.DateModified.String()),
		History:       a.History,
	}
}

func (d MysqlDatabase) MapCreateContentDataParams(a CreateContentDataParams) mdbm.CreateContentDataParams {
	return mdbm.CreateContentDataParams{
		RouteID:      Int64ToNullInt32(a.RouteID),
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		DatatypeID:   Int64ToNullInt32(a.DatatypeID),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		History:      a.History,
	}
}

func (d MysqlDatabase) MapUpdateContentDataParams(a UpdateContentDataParams) mdbm.UpdateContentDataParams {
	return mdbm.UpdateContentDataParams{
		RouteID:       Int64ToNullInt32(a.RouteID),
		ParentID:      Int64ToNullInt32(a.ParentID.Int64),
		DatatypeID:    Int64ToNullInt32(a.DatatypeID),
		AuthorID:      int32(a.AuthorID),
		DateCreated:   StringToNTime(a.DateCreated.String).Time,
		DateModified:  StringToNTime(a.DateModified.String).Time,
		History:       a.History,
		ContentDataID: int32(a.ContentDataID),
	}
}

// /QUERIES
func (d MysqlDatabase) CountContentData() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateContentDataTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateContentData(s CreateContentDataParams) ContentData {
	params := d.MapCreateContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentData: %v\n", err)
	}
	row, err := queries.GetLastContentData(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted ContentData: %v\n", err)
	}
	return d.MapContentData(row)
}

func (d MysqlDatabase) DeleteContentData(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete ContentData: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetContentData(id int64) (*ContentData, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentData(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

func (d MysqlDatabase) ListContentData() (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentDataByRoute(routeID int64) (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentDataByRoute(d.Context, Int64ToNullInt32(routeID))
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData by route: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateContentData(s UpdateContentDataParams) (*string, error) {
	params := d.MapUpdateContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content data %v\n", s.ContentDataID)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapContentData(a mdbp.ContentData) ContentData {
	return ContentData{
		ContentDataID: int64(a.ContentDataID),
		RouteID:       int64(a.RouteID.Int32),
		ParentID:      Int64ToNullInt64(int64(a.ParentID.Int32)),
		DatatypeID:    int64(a.DatatypeID.Int32),
		AuthorID:      int64(a.AuthorID),
		DateCreated:   StringToNullString(NullTimeToString(a.DateCreated)),
		DateModified:  StringToNullString(NullTimeToString(a.DateModified)),
		History:       a.History,
	}
}

func (d PsqlDatabase) MapCreateContentDataParams(a CreateContentDataParams) mdbp.CreateContentDataParams {
	return mdbp.CreateContentDataParams{
		RouteID:      Int64ToNullInt32(a.RouteID),
		ParentID:     Int64ToNullInt32(a.ParentID.Int64),
		DatatypeID:   Int64ToNullInt32(a.DatatypeID),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		History:      a.History,
	}
}

func (d PsqlDatabase) MapUpdateContentDataParams(a UpdateContentDataParams) mdbp.UpdateContentDataParams {
	return mdbp.UpdateContentDataParams{
		RouteID:       Int64ToNullInt32(a.RouteID),
		ParentID:      Int64ToNullInt32(a.ParentID.Int64),
		DatatypeID:    Int64ToNullInt32(a.DatatypeID),
		AuthorID:      int32(a.AuthorID),
		DateCreated:   StringToNTime(a.DateCreated.String),
		DateModified:  StringToNTime(a.DateModified.String),
		History:       a.History,
		ContentDataID: int32(a.ContentDataID),
	}
}

// /QUERIES
func (d PsqlDatabase) CountContentData() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateContentDataTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateContentDataTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateContentData(s CreateContentDataParams) ContentData {
	params := d.MapCreateContentDataParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentData: %v\n", err)
	}
	return d.MapContentData(row)
}

func (d PsqlDatabase) DeleteContentData(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteContentData(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete ContentData: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetContentData(id int64) (*ContentData, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentData(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

func (d PsqlDatabase) ListContentData() (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentDataByRoute(routeID int64) (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentDataByRoute(d.Context, Int64ToNullInt32(routeID))
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentData by route: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateContentData(s UpdateContentDataParams) (*string, error) {
	params := d.MapUpdateContentDataParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content data %v\n", s.ContentDataID)
	return &u, nil
}
