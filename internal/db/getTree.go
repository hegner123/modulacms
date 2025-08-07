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
	DatatypeLabel string         `json:"datatype_label"`
	DatatypeType  string         `json:"datatype_type"`
	FieldLabel    string         `json:"field_label"`
	FieldType     string         `json:"field_type"`
	FieldValue    sql.NullString `json:"field_value"`
}

///////////////////////////////
//SQLITE
//////////////////////////////

func (d Database) MapGetRouteTreeByRouteIDRow(a mdb.GetRouteTreeByRouteIDRow) GetRouteTreeByRouteIDRow {
	return GetRouteTreeByRouteIDRow{
		ContentDataID: a.ContentDataID,
		ParentID:      a.ParentID,
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
		FieldLabel:    a.FieldLabel,
		FieldType:     a.FieldType,
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

///////////////////////////////
//MYSQL
//////////////////////////////

func (d MysqlDatabase) MapGetRouteTreeByRouteIDRow(a mdbm.GetRouteTreeByRouteIDRow) GetRouteTreeByRouteIDRow {
	return GetRouteTreeByRouteIDRow{
		ContentDataID: int64(a.ContentDataID),
		ParentID:      NullInt32ToNullInt64(a.ParentID),
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
		FieldLabel:    a.FieldLabel,
		FieldType:     a.FieldType,
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

///////////////////////////////
//POSTGRES
//////////////////////////////

func (d PsqlDatabase) MapGetRouteTreeByRouteIDRow(a mdbp.GetRouteTreeByRouteIDRow) GetRouteTreeByRouteIDRow {
	return GetRouteTreeByRouteIDRow{
		ContentDataID: int64(a.ContentDataID),
		ParentID:      NullInt32ToNullInt64(a.ParentID),
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
		FieldLabel:    a.FieldLabel,
		FieldType:     a.FieldType,
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
