package db

import (
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

///////////////////////////////
//STRUCTS
//////////////////////////////

type GetRouteTreeByRouteIDRow struct {
	ContentDataID int64          `json:"content_data_id"`
	ParentID      sql.NullInt64  `json:"parent_id"`
	FirstChildID  sql.NullInt64  `json:"first_child_id"`
	NextSiblingID sql.NullInt64  `json:"next_sibling_id"`
	PrevSiblingID sql.NullInt64  `json:"prev_sibling_id"`
	DatatypeLabel string         `json:"datatype_label"`
	DatatypeType  string         `json:"datatype_type"`
	FieldLabel    string         `json:"field_label"`
	FieldType     string         `json:"field_type"`
	FieldValue    sql.NullString `json:"field_value"`
}
type GetContentTreeByRouteRow struct {
	ContentDataID int64          `json:"content_data_id"`
	ParentID      sql.NullInt64  `json:"parent_id"`
	FirstChildID  sql.NullInt64  `json:"first_child_id"`
	NextSiblingID sql.NullInt64  `json:"next_sibling_id"`
	PrevSiblingID sql.NullInt64  `json:"prev_sibling_id"`
	DatatypeID    int64          `json:"datatype_id"`
	RouteID       int64          `json:"route_id"`
	AuthorID      int64          `json:"author_id"`
	DateCreated   sql.NullString `json:"date_created"`
	DateModified  sql.NullString `json:"date_modified"`
	DatatypeLabel string         `json:"datatype_label"`
	DatatypeType  string         `json:"datatype_type"`
}

type GetContentFieldsByRouteRow struct {
	ContentDataID int64  `json:"content_data_id"`
	FieldID       int64  `json:"field_id"`
	FieldValue    string `json:"field_value"`
}

type GetFieldDefinitionsByRouteRow struct {
	FieldID    int64  `json:"field_id"`
	Label      string `json:"label"`
	Type       string `json:"type"`
	DatatypeID int64  `json:"datatype_id"`
}

///////////////////////////////
//SQLITE
//////////////////////////////

func (d Database) MapGetRouteTreeByRouteIDRow(a mdb.GetRouteTreeByRouteIDRow) GetRouteTreeByRouteIDRow {
	return GetRouteTreeByRouteIDRow{
		ContentDataID: a.ContentDataID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
		FieldLabel:    a.FieldLabel,
		FieldType:     a.FieldType,
		FieldValue:    a.FieldValue,
	}
}

// GetContentTreeByRouteRow
func (d Database) MapGetContentTreeByRouteRow(a mdb.GetContentTreeByRouteRow) GetContentTreeByRouteRow {
	return GetContentTreeByRouteRow{
		ContentDataID: a.ContentDataID,
		ParentID:      a.ParentID,
		FirstChildID:  a.FirstChildID,
		NextSiblingID: a.NextSiblingID,
		PrevSiblingID: a.PrevSiblingID,
		DatatypeID:    a.DatatypeID,
		RouteID:       a.RouteID,
		AuthorID:      a.AuthorID,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
	}
}

// GetFieldDefinitionsByRouteRow
func (d Database) MapGetFieldDefinitionsByRouteRow(a mdb.GetFieldDefinitionsByRouteRow) GetFieldDefinitionsByRouteRow {
	return GetFieldDefinitionsByRouteRow{
		Label:      a.Label,
		FieldID:    a.FieldID,
		Type:       a.Type,
		DatatypeID: a.DatatypeID,
	}
}

// GetContentFieldsByRouteRow
func (d Database) MapGetContentFieldsByRouteRow(a mdb.GetContentFieldsByRouteRow) GetContentFieldsByRouteRow {
	return GetContentFieldsByRouteRow{
		ContentDataID: a.ContentDataID,
		FieldID:       a.FieldID,
		FieldValue:    a.FieldValue,
	}
}

func (d Database) GetRouteTreeByRouteID(routeID int64) (*[]GetRouteTreeByRouteIDRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetRouteTreeByRouteID(d.Context, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route tree: %v", err)
	}
	res := []GetRouteTreeByRouteIDRow{}
	for _, v := range rows {
		m := d.MapGetRouteTreeByRouteIDRow(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) GetContentTreeByRoute(id int64) (*[]GetContentTreeByRouteRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetContentTreeByRoute(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get content tree: %v", err)
	}
	res := []GetContentTreeByRouteRow{}
	for _, v := range rows {
		m := d.MapGetContentTreeByRouteRow(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) GetFieldDefinitionsByRoute(id int64) (*[]GetFieldDefinitionsByRouteRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetFieldDefinitionsByRoute(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get field definitions: %v", err)
	}
	res := []GetFieldDefinitionsByRouteRow{}
	for _, v := range rows {
		m := d.MapGetFieldDefinitionsByRouteRow(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) GetContentFieldsByRoute(id int64) (*[]GetContentFieldsByRouteRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetContentFieldsByRoute(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []GetContentFieldsByRouteRow{}
	for _, v := range rows {
		m := d.MapGetContentFieldsByRouteRow(v)
		res = append(res, m)
	}
	return &res, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

func (d MysqlDatabase) MapGetRouteTreeByRouteIDRow(a mdbm.GetRouteTreeByRouteIDRow) GetRouteTreeByRouteIDRow {
	return GetRouteTreeByRouteIDRow{
		ContentDataID: int64(a.ContentDataID),
		ParentID:      NullInt32ToNullInt64(a.ParentID),
		FirstChildID:  NullInt32ToNullInt64(a.FirstChildID),
		NextSiblingID: NullInt32ToNullInt64(a.NextSiblingID),
		PrevSiblingID: NullInt32ToNullInt64(a.PrevSiblingID),
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
		FieldLabel:    a.FieldLabel,
		FieldType:     a.FieldType,
		FieldValue:    a.FieldValue,
	}
}

// GetContentTreeByRouteRow
func (d MysqlDatabase) MapGetContentTreeByRouteRow(a mdbm.GetContentTreeByRouteRow) GetContentTreeByRouteRow {
	return GetContentTreeByRouteRow{
		ContentDataID: int64(a.ContentDataID),
		ParentID:      NullInt32ToNullInt64(a.ParentID),
		FirstChildID:  NullInt32ToNullInt64(a.FirstChildID),
		NextSiblingID: NullInt32ToNullInt64(a.NextSiblingID),
		PrevSiblingID: NullInt32ToNullInt64(a.PrevSiblingID),
		DatatypeID:    int64(a.DatatypeID.Int32),
		RouteID:       int64(a.RouteID.Int32),
		AuthorID:      int64(a.AuthorID),
		DateCreated:   TimeToNullString(a.DateCreated),
		DateModified:  TimeToNullString(a.DateModified),
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
	}
}

// GetFieldDefinitionsByRouteRow
func (d MysqlDatabase) MapGetFieldDefinitionsByRouteRow(a mdbm.GetFieldDefinitionsByRouteRow) GetFieldDefinitionsByRouteRow {
	return GetFieldDefinitionsByRouteRow{
		Label:      a.Label,
		FieldID:    int64(a.FieldID),
		Type:       a.Type,
		DatatypeID: int64(a.DatatypeID),
	}
}

// GetContentFieldsByRouteRow
func (d MysqlDatabase) MapGetContentFieldsByRouteRow(a mdbm.GetContentFieldsByRouteRow) GetContentFieldsByRouteRow {
	return GetContentFieldsByRouteRow{
		ContentDataID: int64(a.ContentDataID),
		FieldID:       int64(a.FieldID),
		FieldValue:    a.FieldValue,
	}
}

func (d MysqlDatabase) GetRouteTreeByRouteID(routeID int64) (*[]GetRouteTreeByRouteIDRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetRouteTreeByRouteID(d.Context, Int64ToNullInt32(routeID))
	if err != nil {
		return nil, fmt.Errorf("failed to get route tree: %v", err)
	}
	res := []GetRouteTreeByRouteIDRow{}
	for _, v := range rows {
		m := d.MapGetRouteTreeByRouteIDRow(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) GetContentTreeByRoute(id int64) (*[]GetContentTreeByRouteRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetContentTreeByRoute(d.Context, Int64ToNullInt32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get route tree: %v", err)
	}
	res := []GetContentTreeByRouteRow{}
	for _, v := range rows {
		m := d.MapGetContentTreeByRouteRow(v)
		res = append(res, m)
	}
	return &res, nil

}
func (d MysqlDatabase) GetFieldDefinitionsByRoute(id int64) (*[]GetFieldDefinitionsByRouteRow, error) {

	return nil, nil
}
func (d MysqlDatabase) GetContentFieldsByRoute(id int64) (*[]GetContentFieldsByRouteRow, error) {

	return nil, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

func (d PsqlDatabase) MapGetRouteTreeByRouteIDRow(a mdbp.GetRouteTreeByRouteIDRow) GetRouteTreeByRouteIDRow {
	return GetRouteTreeByRouteIDRow{
		ContentDataID: int64(a.ContentDataID),
		ParentID:      NullInt32ToNullInt64(a.ParentID),
		FirstChildID:  NullInt32ToNullInt64(a.FirstChildID),
		NextSiblingID: NullInt32ToNullInt64(a.NextSiblingID),
		PrevSiblingID: NullInt32ToNullInt64(a.PrevSiblingID),
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
		FieldLabel:    a.FieldLabel,
		FieldType:     a.FieldType,
		FieldValue:    a.FieldValue,
	}
}
func (d PsqlDatabase) MapGetContentTreeByRouteRow(a mdbm.GetContentTreeByRouteRow) GetContentTreeByRouteRow {
	return GetContentTreeByRouteRow{
		ContentDataID: int64(a.ContentDataID),
		ParentID:      NullInt32ToNullInt64(a.ParentID),
		FirstChildID:  NullInt32ToNullInt64(a.FirstChildID),
		NextSiblingID: NullInt32ToNullInt64(a.NextSiblingID),
		PrevSiblingID: NullInt32ToNullInt64(a.PrevSiblingID),
		DatatypeID:    int64(a.DatatypeID.Int32),
		RouteID:       int64(a.RouteID.Int32),
		AuthorID:      int64(a.AuthorID),
		DateCreated:   TimeToNullString(a.DateCreated),
		DateModified:  TimeToNullString(a.DateModified),
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
	}
}

// GetFieldDefinitionsByRouteRow
func (d PsqlDatabase) MapGetFieldDefinitionsByRouteRow(a mdbm.GetFieldDefinitionsByRouteRow) GetFieldDefinitionsByRouteRow {
	return GetFieldDefinitionsByRouteRow{
		Label:      a.Label,
		FieldID:    int64(a.FieldID),
		Type:       a.Type,
		DatatypeID: int64(a.DatatypeID),
	}
}

// GetContentFieldsByRouteRow
func (d PsqlDatabase) MapGetContentFieldsByRouteRow(a mdbm.GetContentFieldsByRouteRow) GetContentFieldsByRouteRow {
	return GetContentFieldsByRouteRow{
		ContentDataID: int64(a.ContentDataID),
		FieldID:       int64(a.FieldID),
		FieldValue:    a.FieldValue,
	}
}

func (d PsqlDatabase) GetRouteTreeByRouteID(routeID int64) (*[]GetRouteTreeByRouteIDRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetRouteTreeByRouteID(d.Context, Int64ToNullInt32(routeID))
	if err != nil {
		return nil, fmt.Errorf("failed to get route tree: %v", err)
	}
	res := []GetRouteTreeByRouteIDRow{}
	for _, v := range rows {
		m := d.MapGetRouteTreeByRouteIDRow(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) GetContentTreeByRoute(id int64) (*[]GetContentTreeByRouteRow, error) {
	return nil, nil

}
func (d PsqlDatabase) GetFieldDefinitionsByRoute(id int64) (*[]GetFieldDefinitionsByRouteRow, error) {
	return nil, nil

}
func (d PsqlDatabase) GetContentFieldsByRoute(id int64) (*[]GetContentFieldsByRouteRow, error) {

	return nil, nil
}
