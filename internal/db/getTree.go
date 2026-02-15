package db

import (
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
//STRUCTS
//////////////////////////////

// GetRouteTreeByRouteIDRow represents a row from the route tree query with datatype and field metadata.
type GetRouteTreeByRouteIDRow struct {
	ContentDataID types.ContentID         `json:"content_data_id"`
	ParentID      types.NullableContentID `json:"parent_id"`
	FirstChildID  types.NullableContentID `json:"first_child_id"`
	NextSiblingID types.NullableContentID `json:"next_sibling_id"`
	PrevSiblingID types.NullableContentID `json:"prev_sibling_id"`
	DatatypeLabel string                  `json:"datatype_label"`
	DatatypeType  string                  `json:"datatype_type"`
	FieldLabel    string                  `json:"field_label"`
	FieldType     types.FieldType         `json:"field_type"`
	FieldValue    sql.NullString          `json:"field_value"`
}

// GetContentTreeByRouteRow represents a content tree node with datatype information.
type GetContentTreeByRouteRow struct {
	ContentDataID types.ContentID          `json:"content_data_id"`
	ParentID      types.NullableContentID  `json:"parent_id"`
	FirstChildID  types.NullableContentID  `json:"first_child_id"`
	NextSiblingID types.NullableContentID  `json:"next_sibling_id"`
	PrevSiblingID types.NullableContentID  `json:"prev_sibling_id"`
	DatatypeID    types.NullableDatatypeID `json:"datatype_id"`
	RouteID       types.NullableRouteID    `json:"route_id"`
	AuthorID      types.NullableUserID     `json:"author_id"`
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
	Status        types.ContentStatus      `json:"status"`
	DatatypeLabel string                   `json:"datatype_label"`
	DatatypeType  string                   `json:"datatype_type"`
}

// GetContentFieldsByRouteRow represents a content field value for a specific content node.
type GetContentFieldsByRouteRow struct {
	ContentDataID types.NullableContentID `json:"content_data_id"`
	FieldID       types.NullableFieldID   `json:"field_id"`
	FieldValue    string                  `json:"field_value"`
}

// GetFieldDefinitionsByRouteRow represents a field definition with its associated datatype.
type GetFieldDefinitionsByRouteRow struct {
	FieldID    types.FieldID   `json:"field_id"`
	Label      string          `json:"label"`
	Type       types.FieldType `json:"type"`
	DatatypeID types.DatatypeID `json:"datatype_id"`
}

///////////////////////////////
//SQLITE
//////////////////////////////

// MapGetRouteTreeByRouteIDRow maps SQLite route tree row to wrapper struct.
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

// MapGetContentTreeByRouteRow maps SQLite content tree row to wrapper struct.
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
		Status:        a.Status,
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
	}
}

// MapGetFieldDefinitionsByRouteRow maps SQLite field definitions row to wrapper struct.
func (d Database) MapGetFieldDefinitionsByRouteRow(a mdb.GetFieldDefinitionsByRouteRow) GetFieldDefinitionsByRouteRow {
	return GetFieldDefinitionsByRouteRow{
		Label:      a.Label,
		FieldID:    a.FieldID,
		Type:       a.Type,
		DatatypeID: a.DatatypeID,
	}
}

// MapGetContentFieldsByRouteRow maps SQLite content fields row to wrapper struct.
func (d Database) MapGetContentFieldsByRouteRow(a mdb.GetContentFieldsByRouteRow) GetContentFieldsByRouteRow {
	return GetContentFieldsByRouteRow{
		ContentDataID: a.ContentDataID,
		FieldID:       a.FieldID,
		FieldValue:    a.FieldValue,
	}
}

// GetRouteTreeByRouteID retrieves the complete route tree including field values for a route.
func (d Database) GetRouteTreeByRouteID(routeID types.NullableRouteID) (*[]GetRouteTreeByRouteIDRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetRouteTreeByRouteID(d.Context, mdb.GetRouteTreeByRouteIDParams{RouteID: routeID})
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

// GetContentTreeByRoute retrieves all content nodes for a route with datatype metadata.
func (d Database) GetContentTreeByRoute(routeID types.NullableRouteID) (*[]GetContentTreeByRouteRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetContentTreeByRoute(d.Context, mdb.GetContentTreeByRouteParams{RouteID: routeID})
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

// GetFieldDefinitionsByRoute retrieves all field definitions associated with a route.
func (d Database) GetFieldDefinitionsByRoute(routeID types.NullableRouteID) (*[]GetFieldDefinitionsByRouteRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetFieldDefinitionsByRoute(d.Context, mdb.GetFieldDefinitionsByRouteParams{RouteID: routeID})
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

// GetContentFieldsByRoute retrieves all content field values for a route.
func (d Database) GetContentFieldsByRoute(routeID types.NullableRouteID) (*[]GetContentFieldsByRouteRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetContentFieldsByRoute(d.Context, mdb.GetContentFieldsByRouteParams{RouteID: routeID})
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

// MapGetRouteTreeByRouteIDRow maps MySQL route tree row to wrapper struct.
func (d MysqlDatabase) MapGetRouteTreeByRouteIDRow(a mdbm.GetRouteTreeByRouteIDRow) GetRouteTreeByRouteIDRow {
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

// MapGetContentTreeByRouteRow maps MySQL content tree row to wrapper struct.
func (d MysqlDatabase) MapGetContentTreeByRouteRow(a mdbm.GetContentTreeByRouteRow) GetContentTreeByRouteRow {
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
		Status:        a.Status,
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
	}
}

// MapGetFieldDefinitionsByRouteRow maps MySQL field definitions row to wrapper struct.
func (d MysqlDatabase) MapGetFieldDefinitionsByRouteRow(a mdbm.GetFieldDefinitionsByRouteRow) GetFieldDefinitionsByRouteRow {
	return GetFieldDefinitionsByRouteRow{
		FieldID:    a.FieldID,
		Label:      a.Label,
		Type:       a.Type,
		DatatypeID: a.DatatypeID,
	}
}

// MapGetContentFieldsByRouteRow maps MySQL content fields row to wrapper struct.
func (d MysqlDatabase) MapGetContentFieldsByRouteRow(a mdbm.GetContentFieldsByRouteRow) GetContentFieldsByRouteRow {
	return GetContentFieldsByRouteRow{
		ContentDataID: a.ContentDataID,
		FieldID:       a.FieldID,
		FieldValue:    a.FieldValue,
	}
}

// GetRouteTreeByRouteID retrieves the complete route tree including field values for a route.
func (d MysqlDatabase) GetRouteTreeByRouteID(routeID types.NullableRouteID) (*[]GetRouteTreeByRouteIDRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetRouteTreeByRouteID(d.Context, mdbm.GetRouteTreeByRouteIDParams{RouteID: routeID})
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

// GetContentTreeByRoute retrieves all content nodes for a route with datatype metadata.
func (d MysqlDatabase) GetContentTreeByRoute(routeID types.NullableRouteID) (*[]GetContentTreeByRouteRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetContentTreeByRoute(d.Context, mdbm.GetContentTreeByRouteParams{RouteID: routeID})
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

// GetFieldDefinitionsByRoute retrieves all field definitions associated with a route.
func (d MysqlDatabase) GetFieldDefinitionsByRoute(routeID types.NullableRouteID) (*[]GetFieldDefinitionsByRouteRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetFieldDefinitionsByRoute(d.Context, mdbm.GetFieldDefinitionsByRouteParams{RouteID: routeID})
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

// GetContentFieldsByRoute retrieves all content field values for a route.
func (d MysqlDatabase) GetContentFieldsByRoute(routeID types.NullableRouteID) (*[]GetContentFieldsByRouteRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetContentFieldsByRoute(d.Context, mdbm.GetContentFieldsByRouteParams{RouteID: routeID})
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
//POSTGRES
//////////////////////////////

// MapGetRouteTreeByRouteIDRow maps PostgreSQL route tree row to wrapper struct.
func (d PsqlDatabase) MapGetRouteTreeByRouteIDRow(a mdbp.GetRouteTreeByRouteIDRow) GetRouteTreeByRouteIDRow {
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

// MapGetContentTreeByRouteRow maps PostgreSQL content tree row to wrapper struct.
func (d PsqlDatabase) MapGetContentTreeByRouteRow(a mdbp.GetContentTreeByRouteRow) GetContentTreeByRouteRow {
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
		Status:        a.Status,
		DatatypeLabel: a.DatatypeLabel,
		DatatypeType:  a.DatatypeType,
	}
}

// MapGetFieldDefinitionsByRouteRow maps PostgreSQL field definitions row to wrapper struct.
func (d PsqlDatabase) MapGetFieldDefinitionsByRouteRow(a mdbp.GetFieldDefinitionsByRouteRow) GetFieldDefinitionsByRouteRow {
	return GetFieldDefinitionsByRouteRow{
		FieldID:    a.FieldID,
		Label:      a.Label,
		Type:       a.Type,
		DatatypeID: a.DatatypeID,
	}
}

// MapGetContentFieldsByRouteRow maps PostgreSQL content fields row to wrapper struct.
func (d PsqlDatabase) MapGetContentFieldsByRouteRow(a mdbp.GetContentFieldsByRouteRow) GetContentFieldsByRouteRow {
	return GetContentFieldsByRouteRow{
		ContentDataID: a.ContentDataID,
		FieldID:       a.FieldID,
		FieldValue:    a.FieldValue,
	}
}

// GetRouteTreeByRouteID retrieves the complete route tree including field values for a route.
func (d PsqlDatabase) GetRouteTreeByRouteID(routeID types.NullableRouteID) (*[]GetRouteTreeByRouteIDRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetRouteTreeByRouteID(d.Context, mdbp.GetRouteTreeByRouteIDParams{RouteID: routeID})
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

// GetContentTreeByRoute retrieves all content nodes for a route with datatype metadata.
func (d PsqlDatabase) GetContentTreeByRoute(routeID types.NullableRouteID) (*[]GetContentTreeByRouteRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetContentTreeByRoute(d.Context, mdbp.GetContentTreeByRouteParams{RouteID: routeID})
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

// GetFieldDefinitionsByRoute retrieves all field definitions associated with a route.
func (d PsqlDatabase) GetFieldDefinitionsByRoute(routeID types.NullableRouteID) (*[]GetFieldDefinitionsByRouteRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetFieldDefinitionsByRoute(d.Context, mdbp.GetFieldDefinitionsByRouteParams{RouteID: routeID})
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

// GetContentFieldsByRoute retrieves all content field values for a route.
func (d PsqlDatabase) GetContentFieldsByRoute(routeID types.NullableRouteID) (*[]GetContentFieldsByRouteRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetContentFieldsByRoute(d.Context, mdbp.GetContentFieldsByRouteParams{RouteID: routeID})
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
