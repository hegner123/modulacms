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

type ContentData struct {
	ContentDataID int64          `json:"content_data_id"`
	RouteID       int64          `json:"route_id"`
	ParentID      sql.NullInt64  `json:"parent_id"`
	DatatypeID    int64          `json:"datatype_id"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
	History       sql.NullString `json:"history"`
}

type CreateContentDataParams struct {
	RouteID      int64          `json:"route_id"`
	ParentID     sql.NullInt64  `json:"parent_id"`
	DatatypeID   int64          `json:"datatype_id"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type UpdateContentDataParams struct {
	RouteID       int64          `json:"route_id"`
	ParentID      sql.NullInt64  `json:"parent_id"`
	DatatypeID    int64          `json:"datatype_id"`
	History       sql.NullString `json:"history"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
	ContentDataID int64          `json:"content_data_id"`
}

type ContentDataHistoryEntry struct {
	ContentDataID int64          `json:"content_data_id"`
	RouteID       int64          `json:"route_id"`
	ParentID      sql.NullInt64  `json:"parent_id"`
	DatatypeID    int64          `json:"datatype_id"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
}

type CreateContentDataFormParams struct {
	RouteID      string `json:"route_id"`
	ParentID     string `json:"parent_id"`
	DatatypeID   string `json:"datatype_id"`
	History      string `json:"history"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

type UpdateContentDataFormParams struct {
	RouteID       string `json:"route_id"`
	ParentID      string `json:"parent_id"`
	DatatypeID    string `json:"datatype_id"`
	History       string `json:"history"`
	DateCreated   string `json:"date_created"`
	DateModified  string `json:"date_modified"`
	ContentDataID string `json:"content_data_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateContentDataParams(a CreateContentDataFormParams) CreateContentDataParams {
	return CreateContentDataParams{
		RouteID:      Si(a.RouteID),
		ParentID:     SNi64(a.ParentID),
		DatatypeID:   Si(a.DatatypeID),
		History:      Ns(a.History),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
	}
}

func MapUpdateContentDataParams(a UpdateContentDataFormParams) UpdateContentDataParams {
	return UpdateContentDataParams{
		RouteID:       Si(a.RouteID),
		ParentID:      SNi64(a.ParentID),
		DatatypeID:    Si(a.DatatypeID),
		History:       Ns(a.History),
		DateCreated:   Ns(a.DateCreated),
		DateModified:  Ns(a.DateModified),
		ContentDataID: Si(a.ContentDataID),
	}
}

func MapStringContentData(a ContentData) StringContentData {
	return StringContentData{
		ContentDataID: strconv.FormatInt(a.ContentDataID, 10),
		RouteID:       strconv.FormatInt(a.RouteID, 10),
		ParentID:      strconv.FormatInt(a.ParentID.Int64, 10),
		DatatypeID:    strconv.FormatInt(a.DatatypeID, 10),
		History:       a.History.String,
		DateCreated:   a.DateCreated.String,
		DateModified:  a.DateModified.String,
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
		History:       a.History,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

func (d Database) MapCreateContentDataParams(a CreateContentDataParams) mdb.CreateContentDataParams {
	return mdb.CreateContentDataParams{
		RouteID:      a.RouteID,
		ParentID:     a.ParentID,
		DatatypeID:   a.DatatypeID,
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateContentDataParams(a UpdateContentDataParams) mdb.UpdateContentDataParams {
	return mdb.UpdateContentDataParams{
		RouteID:       a.RouteID,
		ParentID:      a.ParentID,
		DatatypeID:    a.DatatypeID,
		History:       a.History,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
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
		ParentID:      Ni64(int64(a.ParentID.Int32)),
		DatatypeID:    int64(a.DatatypeID.Int32),
		History:       a.History,
		DateCreated:   Ns(Nt(a.DateCreated)),
		DateModified:  Ns(Nt(a.DateModified)),
	}
}

func (d MysqlDatabase) MapCreateContentDataParams(a CreateContentDataParams) mdbm.CreateContentDataParams {
	return mdbm.CreateContentDataParams{
		RouteID:      Ni32(a.RouteID),
		ParentID:     Ni32(a.ParentID.Int64),
		DatatypeID:   Ni32(a.DatatypeID),
		History:      a.History,
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
	}
}

func (d MysqlDatabase) MapUpdateContentDataParams(a UpdateContentDataParams) mdbm.UpdateContentDataParams {
	return mdbm.UpdateContentDataParams{
		RouteID:       Ni32(a.RouteID),
		ParentID:      Ni32(a.ParentID.Int64),
		DatatypeID:    Ni32(a.DatatypeID),
		History:       a.History,
		DateCreated:   StringToNTime(a.DateCreated.String),
		DateModified:  StringToNTime(a.DateModified.String),
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
	rows, err := queries.ListContentDataByRoute(d.Context, Ni32(routeID))
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
		ParentID:      Ni64(int64(a.ParentID.Int32)),
		DatatypeID:    int64(a.DatatypeID.Int32),
		History:       a.History,
		DateCreated:   Ns(Nt(a.DateCreated)),
		DateModified:  Ns(Nt(a.DateModified)),
	}
}

func (d PsqlDatabase) MapCreateContentDataParams(a CreateContentDataParams) mdbp.CreateContentDataParams {
	return mdbp.CreateContentDataParams{
		RouteID:      Ni32(a.RouteID),
		ParentID:     Ni32(a.ParentID.Int64),
		DatatypeID:   Ni32(a.DatatypeID),
		History:      a.History,
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
	}
}

func (d PsqlDatabase) MapUpdateContentDataParams(a UpdateContentDataParams) mdbp.UpdateContentDataParams {
	return mdbp.UpdateContentDataParams{
		RouteID:       Ni32(a.RouteID),
		ParentID:      Ni32(a.ParentID.Int64),
		DatatypeID:    Ni32(a.DatatypeID),
		History:       a.History,
		DateCreated:   StringToNTime(a.DateCreated.String),
		DateModified:  StringToNTime(a.DateModified.String),
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
	rows, err := queries.ListContentDataByRoute(d.Context, Ni32(routeID))
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
