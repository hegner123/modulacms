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
	AuthorID      types.UserID             `json:"author_id"`
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
	FieldID    types.FieldID    `json:"field_id"`
	Label      string           `json:"label"`
	Type       types.FieldType  `json:"type"`
	DatatypeID types.DatatypeID `json:"datatype_id"`
}

// AdminContentDataWithDatatypeRow represents an admin content data row joined with its datatype.
type AdminContentDataWithDatatypeRow struct {
	AdminContentDataID types.AdminContentID          `json:"admin_content_data_id"`
	ParentID           types.NullableAdminContentID  `json:"parent_id"`
	FirstChildID       types.NullableAdminContentID  `json:"first_child_id"`
	NextSiblingID      types.NullableAdminContentID  `json:"next_sibling_id"`
	PrevSiblingID      types.NullableAdminContentID  `json:"prev_sibling_id"`
	AdminRouteID       types.NullableAdminRouteID    `json:"admin_route_id"`
	AdminDatatypeID    types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AuthorID           types.UserID                  `json:"author_id"`
	Status             types.ContentStatus           `json:"status"`
	DateCreated        types.Timestamp               `json:"date_created"`
	DateModified       types.Timestamp               `json:"date_modified"`
	DtAdminDatatypeID  types.AdminDatatypeID         `json:"dt_admin_datatype_id"`
	DtParentID         types.NullableAdminDatatypeID `json:"dt_parent_id"`
	DtLabel            string                        `json:"dt_label"`
	DtType             string                        `json:"dt_type"`
	DtAuthorID         types.UserID                  `json:"dt_author_id"`
	DtDateCreated      types.Timestamp               `json:"dt_date_created"`
	DtDateModified     types.Timestamp               `json:"dt_date_modified"`
}

// ContentFieldWithFieldRow represents a content field row joined with its field definition.
type ContentFieldWithFieldRow struct {
	ContentFieldID types.ContentFieldID    `json:"content_field_id"`
	RouteID        types.NullableRouteID   `json:"route_id"`
	ContentDataID  types.NullableContentID `json:"content_data_id"`
	FieldID        types.NullableFieldID   `json:"field_id"`
	FieldValue     string                  `json:"field_value"`
	AuthorID       types.NullableUserID    `json:"author_id"`
	DateCreated    types.Timestamp         `json:"date_created"`
	DateModified   types.Timestamp         `json:"date_modified"`
	FFieldID       types.FieldID           `json:"f_field_id"`
	FLabel         string                  `json:"f_label"`
	FType          types.FieldType         `json:"f_type"`
}

// UserWithRoleLabelRow represents a user joined with the role label.
type UserWithRoleLabelRow struct {
	UserID       types.UserID    `json:"user_id"`
	Username     string          `json:"username"`
	Name         string          `json:"name"`
	Email        types.Email     `json:"email"`
	Role         string          `json:"role"`
	RoleLabel    string          `json:"role_label"`
	DateCreated  types.Timestamp `json:"date_created"`
	DateModified types.Timestamp `json:"date_modified"`
}

// FieldWithSortOrderRow represents a field definition joined with its sort order from datatypes_fields.
type FieldWithSortOrderRow struct {
	SortOrder  int64           `json:"sort_order"`
	FieldID    types.FieldID   `json:"field_id"`
	Label      string          `json:"label"`
	Type       types.FieldType `json:"type"`
	Data       string          `json:"data"`
	Validation string          `json:"validation"`
	UIConfig   string          `json:"ui_config"`
}

// AdminContentFieldsWithFieldRow represents an admin content field row joined with its field definition.
type AdminContentFieldsWithFieldRow struct {
	AdminContentFieldID types.AdminContentFieldID     `json:"admin_content_field_id"`
	AdminRouteID        types.NullableAdminRouteID    `json:"admin_route_id"`
	AdminContentDataID  types.NullableAdminContentID  `json:"admin_content_data_id"`
	AdminFieldID        types.NullableAdminFieldID    `json:"admin_field_id"`
	AdminFieldValue     string                        `json:"admin_field_value"`
	AuthorID            types.NullableUserID          `json:"author_id"`
	DateCreated         types.Timestamp               `json:"date_created"`
	DateModified        types.Timestamp               `json:"date_modified"`
	FAdminFieldID       types.AdminFieldID            `json:"f_admin_field_id"`
	FParentID           types.NullableAdminDatatypeID `json:"f_parent_id"`
	FLabel              string                        `json:"f_label"`
	FData               string                        `json:"f_data"`
	FValidation         string                        `json:"f_validation"`
	FUIConfig           string                        `json:"f_ui_config"`
	FType               types.FieldType               `json:"f_type"`
	FAuthorID           types.NullableUserID          `json:"f_author_id"`
	FDateCreated        types.Timestamp               `json:"f_date_created"`
	FDateModified       types.Timestamp               `json:"f_date_modified"`
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

// MapAdminContentDataWithDatatypeRow maps SQLite admin content+datatype JOIN row to wrapper struct.
func (d Database) MapAdminContentDataWithDatatypeRow(a mdb.ListAdminContentDataWithDatatypeByRouteRow) AdminContentDataWithDatatypeRow {
	return AdminContentDataWithDatatypeRow{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		DtAdminDatatypeID:  a.DtAdminDatatypeId,
		DtParentID:         a.DtParentId,
		DtLabel:            a.DtLabel,
		DtType:             a.DtType,
		DtAuthorID:         a.DtAuthorId,
		DtDateCreated:      a.DtDateCreated,
		DtDateModified:     a.DtDateModified,
	}
}

// MapAdminContentFieldsWithFieldRow maps SQLite admin content field+field JOIN row to wrapper struct.
func (d Database) MapAdminContentFieldsWithFieldRow(a mdb.ListAdminContentFieldsWithFieldByRouteRow) AdminContentFieldsWithFieldRow {
	return AdminContentFieldsWithFieldRow{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		FAdminFieldID:       a.FAdminFieldId,
		FParentID:           a.FParentId,
		FLabel:              a.FLabel,
		FData:               a.FData,
		FValidation:         a.FValidation,
		FUIConfig:           a.FUiConfig,
		FType:               a.FType,
		FAuthorID:           a.FAuthorId,
		FDateCreated:        a.FDateCreated,
		FDateModified:       a.FDateModified,
	}
}

// ListAdminContentDataWithDatatypeByRoute retrieves admin content data joined with datatypes for a route.
func (d Database) ListAdminContentDataWithDatatypeByRoute(routeID types.NullableAdminRouteID) (*[]AdminContentDataWithDatatypeRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentDataWithDatatypeByRoute(d.Context, mdb.ListAdminContentDataWithDatatypeByRouteParams{AdminRouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content data with datatypes: %w", err)
	}
	res := []AdminContentDataWithDatatypeRow{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentDataWithDatatypeRow(v))
	}
	return &res, nil
}

// ListAdminContentFieldsWithFieldByRoute retrieves admin content fields joined with field definitions for a route.
func (d Database) ListAdminContentFieldsWithFieldByRoute(routeID types.NullableAdminRouteID) (*[]AdminContentFieldsWithFieldRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsWithFieldByRoute(d.Context, mdb.ListAdminContentFieldsWithFieldByRouteParams{AdminRouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content fields with field definitions: %w", err)
	}
	res := []AdminContentFieldsWithFieldRow{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentFieldsWithFieldRow(v))
	}
	return &res, nil
}

// MapContentFieldWithFieldRow maps SQLite content field+field JOIN row to wrapper struct.
func (d Database) MapContentFieldWithFieldRow(a mdb.ListContentFieldsWithFieldByContentDataRow) ContentFieldWithFieldRow {
	return ContentFieldWithFieldRow{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
		FFieldID:       a.FFieldId,
		FLabel:         a.FLabel,
		FType:          a.FType,
	}
}

// ListContentFieldsWithFieldByContentData retrieves content fields joined with field definitions for a content data item.
func (d Database) ListContentFieldsWithFieldByContentData(contentDataID types.NullableContentID) (*[]ContentFieldWithFieldRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsWithFieldByContentData(d.Context, mdb.ListContentFieldsWithFieldByContentDataParams{ContentDataID: contentDataID})
	if err != nil {
		return nil, fmt.Errorf("failed to list content fields with field definitions: %w", err)
	}
	res := []ContentFieldWithFieldRow{}
	for _, v := range rows {
		res = append(res, d.MapContentFieldWithFieldRow(v))
	}
	return &res, nil
}

// MapUserWithRoleLabelRow maps SQLite user+role JOIN row to wrapper struct.
func (d Database) MapUserWithRoleLabelRow(a mdb.ListUsersWithRoleLabelRow) UserWithRoleLabelRow {
	return UserWithRoleLabelRow{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Role:         a.Roles,
		RoleLabel:    a.RoleLabel,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// ListUsersWithRoleLabel retrieves all users with their role label.
func (d Database) ListUsersWithRoleLabel() (*[]UserWithRoleLabelRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListUsersWithRoleLabel(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list users with role label: %w", err)
	}
	res := []UserWithRoleLabelRow{}
	for _, v := range rows {
		res = append(res, d.MapUserWithRoleLabelRow(v))
	}
	return &res, nil
}

// MapFieldWithSortOrderRow maps SQLite field+sort_order JOIN row to wrapper struct.
func (d Database) MapFieldWithSortOrderRow(a mdb.ListFieldsWithSortOrderByDatatypeIDRow) FieldWithSortOrderRow {
	return FieldWithSortOrderRow{
		SortOrder:  a.SortOrder,
		FieldID:    a.FieldID,
		Label:      a.Label,
		Type:       a.Type,
		Data:       a.Data,
		Validation: a.Validation,
		UIConfig:   a.UiConfig,
	}
}

// ListFieldsWithSortOrderByDatatypeID retrieves field definitions with sort order for a datatype.
func (d Database) ListFieldsWithSortOrderByDatatypeID(datatypeID types.DatatypeID) (*[]FieldWithSortOrderRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListFieldsWithSortOrderByDatatypeID(d.Context, mdb.ListFieldsWithSortOrderByDatatypeIDParams{DatatypeID: datatypeID})
	if err != nil {
		return nil, fmt.Errorf("failed to list fields with sort order: %w", err)
	}
	res := []FieldWithSortOrderRow{}
	for _, v := range rows {
		res = append(res, d.MapFieldWithSortOrderRow(v))
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

// MapAdminContentDataWithDatatypeRow maps MySQL admin content+datatype JOIN row to wrapper struct.
func (d MysqlDatabase) MapAdminContentDataWithDatatypeRow(a mdbm.ListAdminContentDataWithDatatypeByRouteRow) AdminContentDataWithDatatypeRow {
	return AdminContentDataWithDatatypeRow{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		DtAdminDatatypeID:  a.DtAdminDatatypeId,
		DtParentID:         a.DtParentId,
		DtLabel:            a.DtLabel,
		DtType:             a.DtType,
		DtAuthorID:         a.DtAuthorId,
		DtDateCreated:      a.DtDateCreated,
		DtDateModified:     a.DtDateModified,
	}
}

// MapAdminContentFieldsWithFieldRow maps MySQL admin content field+field JOIN row to wrapper struct.
func (d MysqlDatabase) MapAdminContentFieldsWithFieldRow(a mdbm.ListAdminContentFieldsWithFieldByRouteRow) AdminContentFieldsWithFieldRow {
	return AdminContentFieldsWithFieldRow{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		FAdminFieldID:       a.FAdminFieldId,
		FParentID:           a.FParentId,
		FLabel:              a.FLabel,
		FData:               a.FData,
		FValidation:         a.FValidation,
		FUIConfig:           a.FUiConfig,
		FType:               a.FType,
		FAuthorID:           a.FAuthorId,
		FDateCreated:        a.FDateCreated,
		FDateModified:       a.FDateModified,
	}
}

// ListAdminContentDataWithDatatypeByRoute retrieves admin content data joined with datatypes for a route.
func (d MysqlDatabase) ListAdminContentDataWithDatatypeByRoute(routeID types.NullableAdminRouteID) (*[]AdminContentDataWithDatatypeRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentDataWithDatatypeByRoute(d.Context, mdbm.ListAdminContentDataWithDatatypeByRouteParams{AdminRouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content data with datatypes: %w", err)
	}
	res := []AdminContentDataWithDatatypeRow{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentDataWithDatatypeRow(v))
	}
	return &res, nil
}

// ListAdminContentFieldsWithFieldByRoute retrieves admin content fields joined with field definitions for a route.
func (d MysqlDatabase) ListAdminContentFieldsWithFieldByRoute(routeID types.NullableAdminRouteID) (*[]AdminContentFieldsWithFieldRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsWithFieldByRoute(d.Context, mdbm.ListAdminContentFieldsWithFieldByRouteParams{AdminRouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content fields with field definitions: %w", err)
	}
	res := []AdminContentFieldsWithFieldRow{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentFieldsWithFieldRow(v))
	}
	return &res, nil
}

// MapContentFieldWithFieldRow maps MySQL content field+field JOIN row to wrapper struct.
func (d MysqlDatabase) MapContentFieldWithFieldRow(a mdbm.ListContentFieldsWithFieldByContentDataRow) ContentFieldWithFieldRow {
	return ContentFieldWithFieldRow{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
		FFieldID:       a.FFieldId,
		FLabel:         a.FLabel,
		FType:          a.FType,
	}
}

// ListContentFieldsWithFieldByContentData retrieves content fields joined with field definitions for a content data item.
func (d MysqlDatabase) ListContentFieldsWithFieldByContentData(contentDataID types.NullableContentID) (*[]ContentFieldWithFieldRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsWithFieldByContentData(d.Context, mdbm.ListContentFieldsWithFieldByContentDataParams{ContentDataID: contentDataID})
	if err != nil {
		return nil, fmt.Errorf("failed to list content fields with field definitions: %w", err)
	}
	res := []ContentFieldWithFieldRow{}
	for _, v := range rows {
		res = append(res, d.MapContentFieldWithFieldRow(v))
	}
	return &res, nil
}

// MapUserWithRoleLabelRow maps MySQL user+role JOIN row to wrapper struct.
func (d MysqlDatabase) MapUserWithRoleLabelRow(a mdbm.ListUsersWithRoleLabelRow) UserWithRoleLabelRow {
	return UserWithRoleLabelRow{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Role:         a.Roles,
		RoleLabel:    a.RoleLabel,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// ListUsersWithRoleLabel retrieves all users with their role label.
func (d MysqlDatabase) ListUsersWithRoleLabel() (*[]UserWithRoleLabelRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListUsersWithRoleLabel(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list users with role label: %w", err)
	}
	res := []UserWithRoleLabelRow{}
	for _, v := range rows {
		res = append(res, d.MapUserWithRoleLabelRow(v))
	}
	return &res, nil
}

// MapFieldWithSortOrderRow maps MySQL field+sort_order JOIN row to wrapper struct.
func (d MysqlDatabase) MapFieldWithSortOrderRow(a mdbm.ListFieldsWithSortOrderByDatatypeIDRow) FieldWithSortOrderRow {
	return FieldWithSortOrderRow{
		SortOrder:  int64(a.SortOrder),
		FieldID:    a.FieldID,
		Label:      a.Label,
		Type:       a.Type,
		Data:       a.Data,
		Validation: a.Validation,
		UIConfig:   a.UiConfig,
	}
}

// ListFieldsWithSortOrderByDatatypeID retrieves field definitions with sort order for a datatype.
func (d MysqlDatabase) ListFieldsWithSortOrderByDatatypeID(datatypeID types.DatatypeID) (*[]FieldWithSortOrderRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListFieldsWithSortOrderByDatatypeID(d.Context, mdbm.ListFieldsWithSortOrderByDatatypeIDParams{DatatypeID: datatypeID})
	if err != nil {
		return nil, fmt.Errorf("failed to list fields with sort order: %w", err)
	}
	res := []FieldWithSortOrderRow{}
	for _, v := range rows {
		res = append(res, d.MapFieldWithSortOrderRow(v))
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

// MapAdminContentDataWithDatatypeRow maps PostgreSQL admin content+datatype JOIN row to wrapper struct.
func (d PsqlDatabase) MapAdminContentDataWithDatatypeRow(a mdbp.ListAdminContentDataWithDatatypeByRouteRow) AdminContentDataWithDatatypeRow {
	return AdminContentDataWithDatatypeRow{
		AdminContentDataID: a.AdminContentDataID,
		ParentID:           a.ParentID,
		FirstChildID:       a.FirstChildID,
		NextSiblingID:      a.NextSiblingID,
		PrevSiblingID:      a.PrevSiblingID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		AuthorID:           a.AuthorID,
		Status:             a.Status,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		DtAdminDatatypeID:  a.DtAdminDatatypeId,
		DtParentID:         a.DtParentId,
		DtLabel:            a.DtLabel,
		DtType:             a.DtType,
		DtAuthorID:         a.DtAuthorId,
		DtDateCreated:      a.DtDateCreated,
		DtDateModified:     a.DtDateModified,
	}
}

// MapAdminContentFieldsWithFieldRow maps PostgreSQL admin content field+field JOIN row to wrapper struct.
func (d PsqlDatabase) MapAdminContentFieldsWithFieldRow(a mdbp.ListAdminContentFieldsWithFieldByRouteRow) AdminContentFieldsWithFieldRow {
	return AdminContentFieldsWithFieldRow{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		FAdminFieldID:       a.FAdminFieldId,
		FParentID:           a.FParentId,
		FLabel:              a.FLabel,
		FData:               a.FData,
		FValidation:         a.FValidation,
		FUIConfig:           a.FUiConfig,
		FType:               a.FType,
		FAuthorID:           a.FAuthorId,
		FDateCreated:        a.FDateCreated,
		FDateModified:       a.FDateModified,
	}
}

// ListAdminContentDataWithDatatypeByRoute retrieves admin content data joined with datatypes for a route.
func (d PsqlDatabase) ListAdminContentDataWithDatatypeByRoute(routeID types.NullableAdminRouteID) (*[]AdminContentDataWithDatatypeRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentDataWithDatatypeByRoute(d.Context, mdbp.ListAdminContentDataWithDatatypeByRouteParams{AdminRouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content data with datatypes: %w", err)
	}
	res := []AdminContentDataWithDatatypeRow{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentDataWithDatatypeRow(v))
	}
	return &res, nil
}

// ListAdminContentFieldsWithFieldByRoute retrieves admin content fields joined with field definitions for a route.
func (d PsqlDatabase) ListAdminContentFieldsWithFieldByRoute(routeID types.NullableAdminRouteID) (*[]AdminContentFieldsWithFieldRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsWithFieldByRoute(d.Context, mdbp.ListAdminContentFieldsWithFieldByRouteParams{AdminRouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content fields with field definitions: %w", err)
	}
	res := []AdminContentFieldsWithFieldRow{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentFieldsWithFieldRow(v))
	}
	return &res, nil
}

// MapContentFieldWithFieldRow maps PostgreSQL content field+field JOIN row to wrapper struct.
func (d PsqlDatabase) MapContentFieldWithFieldRow(a mdbp.ListContentFieldsWithFieldByContentDataRow) ContentFieldWithFieldRow {
	return ContentFieldWithFieldRow{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
		FFieldID:       a.FFieldId,
		FLabel:         a.FLabel,
		FType:          a.FType,
	}
}

// ListContentFieldsWithFieldByContentData retrieves content fields joined with field definitions for a content data item.
func (d PsqlDatabase) ListContentFieldsWithFieldByContentData(contentDataID types.NullableContentID) (*[]ContentFieldWithFieldRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsWithFieldByContentData(d.Context, mdbp.ListContentFieldsWithFieldByContentDataParams{ContentDataID: contentDataID})
	if err != nil {
		return nil, fmt.Errorf("failed to list content fields with field definitions: %w", err)
	}
	res := []ContentFieldWithFieldRow{}
	for _, v := range rows {
		res = append(res, d.MapContentFieldWithFieldRow(v))
	}
	return &res, nil
}

// MapUserWithRoleLabelRow maps PostgreSQL user+role JOIN row to wrapper struct.
func (d PsqlDatabase) MapUserWithRoleLabelRow(a mdbp.ListUsersWithRoleLabelRow) UserWithRoleLabelRow {
	return UserWithRoleLabelRow{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Role:         a.Roles,
		RoleLabel:    a.RoleLabel,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// ListUsersWithRoleLabel retrieves all users with their role label.
func (d PsqlDatabase) ListUsersWithRoleLabel() (*[]UserWithRoleLabelRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListUsersWithRoleLabel(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list users with role label: %w", err)
	}
	res := []UserWithRoleLabelRow{}
	for _, v := range rows {
		res = append(res, d.MapUserWithRoleLabelRow(v))
	}
	return &res, nil
}

// MapFieldWithSortOrderRow maps PostgreSQL field+sort_order JOIN row to wrapper struct.
func (d PsqlDatabase) MapFieldWithSortOrderRow(a mdbp.ListFieldsWithSortOrderByDatatypeIDRow) FieldWithSortOrderRow {
	return FieldWithSortOrderRow{
		SortOrder:  int64(a.SortOrder),
		FieldID:    a.FieldID,
		Label:      a.Label,
		Type:       a.Type,
		Data:       a.Data,
		Validation: a.Validation,
		UIConfig:   a.UiConfig,
	}
}

// ListFieldsWithSortOrderByDatatypeID retrieves field definitions with sort order for a datatype.
func (d PsqlDatabase) ListFieldsWithSortOrderByDatatypeID(datatypeID types.DatatypeID) (*[]FieldWithSortOrderRow, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListFieldsWithSortOrderByDatatypeID(d.Context, mdbp.ListFieldsWithSortOrderByDatatypeIDParams{DatatypeID: datatypeID})
	if err != nil {
		return nil, fmt.Errorf("failed to list fields with sort order: %w", err)
	}
	res := []FieldWithSortOrderRow{}
	for _, v := range rows {
		res = append(res, d.MapFieldWithSortOrderRow(v))
	}
	return &res, nil
}
