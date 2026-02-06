package db

import (
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type ContentFields struct {
	ContentFieldID types.ContentFieldID    `json:"content_field_id"`
	RouteID        types.NullableRouteID   `json:"route_id"`
	ContentDataID  types.NullableContentID `json:"content_data_id"`
	FieldID        types.NullableFieldID   `json:"field_id"`
	FieldValue     string                  `json:"field_value"`
	AuthorID       types.NullableUserID    `json:"author_id"`
	DateCreated    types.Timestamp         `json:"date_created"`
	DateModified   types.Timestamp         `json:"date_modified"`
}

type CreateContentFieldParams struct {
	RouteID       types.NullableRouteID   `json:"route_id"`
	ContentDataID types.NullableContentID `json:"content_data_id"`
	FieldID       types.NullableFieldID   `json:"field_id"`
	FieldValue    string                  `json:"field_value"`
	AuthorID      types.NullableUserID    `json:"author_id"`
	DateCreated   types.Timestamp         `json:"date_created"`
	DateModified  types.Timestamp         `json:"date_modified"`
}

type UpdateContentFieldParams struct {
	RouteID        types.NullableRouteID   `json:"route_id"`
	ContentDataID  types.NullableContentID `json:"content_data_id"`
	FieldID        types.NullableFieldID   `json:"field_id"`
	FieldValue     string                  `json:"field_value"`
	AuthorID       types.NullableUserID    `json:"author_id"`
	DateCreated    types.Timestamp         `json:"date_created"`
	DateModified   types.Timestamp         `json:"date_modified"`
	ContentFieldID types.ContentFieldID    `json:"content_field_id"`
}

// FormParams and JSON variants removed - use typed params directly

// ContentFieldsJSON is used for JSON serialization in model package
// Deprecated: Will be removed in future version. Use typed ContentFields directly.
type ContentFieldsJSON struct {
	ContentFieldID int64  `json:"content_field_id"`
	RouteID        int64  `json:"route_id"`
	ContentDataID  int64  `json:"content_data_id"`
	FieldID        int64  `json:"field_id"`
	FieldValue     string `json:"field_value"`
	AuthorID       int64  `json:"author_id"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
}

// MapContentFieldJSON converts ContentFields to ContentFieldsJSON for JSON serialization
// Deprecated: Will be removed in future version
func MapContentFieldJSON(a ContentFields) ContentFieldsJSON {
	return ContentFieldsJSON{
		ContentFieldID: 0,                       // Type conversion not available, set to 0
		RouteID:        0,                       // Type conversion not available, set to 0
		ContentDataID:  0,                       // Type conversion not available, set to 0
		FieldID:        0,                       // Type conversion not available, set to 0
		FieldValue:     a.FieldValue,
		AuthorID:       0,                       // Type conversion not available, set to 0
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
	}
}

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

func MapStringContentField(a ContentFields) StringContentFields {
	return StringContentFields{
		ContentFieldID: a.ContentFieldID.String(),
		RouteID:        a.RouteID.String(),
		ContentDataID:  a.ContentDataID.String(),
		FieldID:        a.FieldID.String(),
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID.String(),
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
		History:        "", // History field removed from ContentFields
	}
}

///////////////////////////////
// SQLITE
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
	}
}

func (d Database) MapCreateContentFieldParams(a CreateContentFieldParams) mdb.CreateContentFieldParams {
	return mdb.CreateContentFieldParams{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
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

func (d Database) DeleteContentField(id types.ContentFieldID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteContentField(d.Context, mdb.DeleteContentFieldParams{ContentFieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete ContentField: %v", id)
	}
	return nil
}

func (d Database) GetContentField(id types.ContentFieldID) (*ContentFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentField(d.Context, mdb.GetContentFieldParams{ContentFieldID: id})
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
		return nil, fmt.Errorf("failed to get ContentFields: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentFieldsByRoute(routeID types.NullableRouteID) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, mdb.ListContentFieldsByRouteParams{RouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentFieldsByContentData(contentDataID types.NullableContentID) (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentData(d.Context, mdb.ListContentFieldsByContentDataParams{ContentDataID: contentDataID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by content data: %v", err)
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
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapContentField(a mdbm.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbm.CreateContentFieldParams {
	return mdbm.CreateContentFieldParams{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
	}
}

func (d MysqlDatabase) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdbm.UpdateContentFieldParams {
	return mdbm.UpdateContentFieldParams{
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		ContentFieldID: a.ContentFieldID,
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

func (d MysqlDatabase) DeleteContentField(id types.ContentFieldID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteContentField(d.Context, mdbm.DeleteContentFieldParams{ContentFieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete ContentField: %v", id)
	}
	return nil
}

func (d MysqlDatabase) GetContentField(id types.ContentFieldID) (*ContentFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentField(d.Context, mdbm.GetContentFieldParams{ContentFieldID: id})
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
		return nil, fmt.Errorf("failed to get ContentFields: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentFieldsByRoute(routeID types.NullableRouteID) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, mdbm.ListContentFieldsByRouteParams{RouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) ListContentFieldsByContentData(contentDataID types.NullableContentID) (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentData(d.Context, mdbm.ListContentFieldsByContentDataParams{ContentDataID: contentDataID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by content data: %v", err)
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
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapContentField(a mdbp.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: a.ContentFieldID,
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbp.CreateContentFieldParams {
	return mdbp.CreateContentFieldParams{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        a.RouteID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdbp.UpdateContentFieldParams {
	return mdbp.UpdateContentFieldParams{
		ContentFieldID:   a.ContentFieldID,
		RouteID:          a.RouteID,
		ContentDataID:    a.ContentDataID,
		FieldID:          a.FieldID,
		FieldValue:       a.FieldValue,
		AuthorID:         a.AuthorID,
		DateCreated:      a.DateCreated,
		DateModified:     a.DateModified,
		ContentFieldID_2: a.ContentFieldID,
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

func (d PsqlDatabase) DeleteContentField(id types.ContentFieldID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteContentField(d.Context, mdbp.DeleteContentFieldParams{ContentFieldID: id})
	if err != nil {
		return fmt.Errorf("failed to delete ContentField: %v", id)
	}
	return nil
}

func (d PsqlDatabase) GetContentField(id types.ContentFieldID) (*ContentFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentField(d.Context, mdbp.GetContentFieldParams{ContentFieldID: id})
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
		return nil, fmt.Errorf("failed to get ContentFields: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentFieldsByRoute(routeID types.NullableRouteID) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByRoute(d.Context, mdbp.ListContentFieldsByRouteParams{RouteID: routeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by route: %v", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) ListContentFieldsByContentData(contentDataID types.NullableContentID) (*[]ContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentFieldsByContentData(d.Context, mdbp.ListContentFieldsByContentDataParams{ContentDataID: contentDataID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ContentFields by content data: %v", err)
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

// Utility function for backward compatibility
// Deprecated: Use types.NullableRouteID directly
func NullableRouteIDFromInt64(id int64) types.NullableRouteID {
	return types.NullableRouteID{ID: types.RouteID(strconv.FormatInt(id, 10)), Valid: id != 0}
}
