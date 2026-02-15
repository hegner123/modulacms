package db

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// routeTreeFixture returns a fully populated GetRouteTreeByRouteIDRow and its component values.
func routeTreeFixture() (GetRouteTreeByRouteIDRow, types.ContentID) {
	contentID := types.NewContentID()
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}

	return GetRouteTreeByRouteIDRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "fc-tree-001", Valid: true},
		NextSiblingID: sql.NullString{String: "ns-tree-001", Valid: true},
		PrevSiblingID: sql.NullString{String: "ps-tree-001", Valid: true},
		DatatypeLabel: "Article",
		DatatypeType:  "content",
		FieldLabel:    "title",
		FieldType:     types.FieldType("text"),
		FieldValue:    sql.NullString{String: "Hello World", Valid: true},
	}, contentID
}

// contentTreeFixture returns a fully populated GetContentTreeByRouteRow and its component values.
func contentTreeFixture() (GetContentTreeByRouteRow, types.ContentID, types.Timestamp) {
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 7, 10, 14, 30, 0, 0, time.UTC))

	return GetContentTreeByRouteRow{
		ContentDataID: contentID,
		ParentID:      types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FirstChildID:  sql.NullString{String: "fc-ct-001", Valid: true},
		NextSiblingID: sql.NullString{String: "ns-ct-001", Valid: true},
		PrevSiblingID: sql.NullString{String: "ps-ct-001", Valid: true},
		DatatypeID:    types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true},
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		AuthorID:      types.NullableUserID{ID: types.NewUserID(), Valid: true},
		DateCreated:   ts,
		DateModified:  ts,
		Status:        types.ContentStatus("published"),
		DatatypeLabel: "Page",
		DatatypeType:  "static",
	}, contentID, ts
}

// --- SQLite Database.MapGetRouteTreeByRouteIDRow tests ---

func TestDatabase_MapGetRouteTreeByRouteIDRow_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	contentID := types.NewContentID()
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}

	input := mdb.GetRouteTreeByRouteIDRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "child-1", Valid: true},
		NextSiblingID: sql.NullString{String: "next-1", Valid: true},
		PrevSiblingID: sql.NullString{String: "prev-1", Valid: true},
		DatatypeLabel: "Blog Post",
		DatatypeType:  "dynamic",
		FieldLabel:    "body",
		FieldType:     types.FieldType("richtext"),
		FieldValue:    sql.NullString{String: "<p>content</p>", Valid: true},
	}

	got := d.MapGetRouteTreeByRouteIDRow(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != input.FirstChildID {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, input.FirstChildID)
	}
	if got.NextSiblingID != input.NextSiblingID {
		t.Errorf("NextSiblingID = %v, want %v", got.NextSiblingID, input.NextSiblingID)
	}
	if got.PrevSiblingID != input.PrevSiblingID {
		t.Errorf("PrevSiblingID = %v, want %v", got.PrevSiblingID, input.PrevSiblingID)
	}
	if got.DatatypeLabel != "Blog Post" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "Blog Post")
	}
	if got.DatatypeType != "dynamic" {
		t.Errorf("DatatypeType = %q, want %q", got.DatatypeType, "dynamic")
	}
	if got.FieldLabel != "body" {
		t.Errorf("FieldLabel = %q, want %q", got.FieldLabel, "body")
	}
	if got.FieldType != types.FieldType("richtext") {
		t.Errorf("FieldType = %v, want %v", got.FieldType, types.FieldType("richtext"))
	}
	if got.FieldValue != input.FieldValue {
		t.Errorf("FieldValue = %v, want %v", got.FieldValue, input.FieldValue)
	}
}

func TestDatabase_MapGetRouteTreeByRouteIDRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapGetRouteTreeByRouteIDRow(mdb.GetRouteTreeByRouteIDRow{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %v, want zero value", got.ContentDataID)
	}
	if got.ParentID.Valid {
		t.Errorf("ParentID.Valid = true, want false for zero value")
	}
	if got.FirstChildID.Valid {
		t.Errorf("FirstChildID.Valid = true, want false for zero value")
	}
	if got.NextSiblingID.Valid {
		t.Errorf("NextSiblingID.Valid = true, want false for zero value")
	}
	if got.PrevSiblingID.Valid {
		t.Errorf("PrevSiblingID.Valid = true, want false for zero value")
	}
	if got.DatatypeLabel != "" {
		t.Errorf("DatatypeLabel = %q, want empty string", got.DatatypeLabel)
	}
	if got.DatatypeType != "" {
		t.Errorf("DatatypeType = %q, want empty string", got.DatatypeType)
	}
	if got.FieldLabel != "" {
		t.Errorf("FieldLabel = %q, want empty string", got.FieldLabel)
	}
	if got.FieldType != "" {
		t.Errorf("FieldType = %v, want zero value", got.FieldType)
	}
	if got.FieldValue.Valid {
		t.Errorf("FieldValue.Valid = true, want false for zero value")
	}
}

func TestDatabase_MapGetRouteTreeByRouteIDRow_NullFieldValue(t *testing.T) {
	t.Parallel()
	d := Database{}
	// FieldValue is sql.NullString -- test that null is preserved through mapping
	input := mdb.GetRouteTreeByRouteIDRow{
		ContentDataID: types.NewContentID(),
		FieldValue:    sql.NullString{Valid: false},
		FieldLabel:    "optional_field",
		FieldType:     types.FieldType("text"),
	}

	got := d.MapGetRouteTreeByRouteIDRow(input)

	if got.FieldValue.Valid {
		t.Errorf("FieldValue.Valid = true, want false for null field value")
	}
	if got.FieldValue.String != "" {
		t.Errorf("FieldValue.String = %q, want empty string for null", got.FieldValue.String)
	}
}

// --- SQLite Database.MapGetContentTreeByRouteRow tests ---

func TestDatabase_MapGetContentTreeByRouteRow_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 5, 20, 8, 0, 0, 0, time.UTC))
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdb.GetContentTreeByRouteRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "child-ct", Valid: true},
		NextSiblingID: sql.NullString{String: "next-ct", Valid: true},
		PrevSiblingID: sql.NullString{String: "prev-ct", Valid: true},
		DatatypeID:    datatypeID,
		RouteID:       routeID,
		AuthorID:      authorID,
		DateCreated:   ts,
		DateModified:  ts,
		Status:        types.ContentStatus("draft"),
		DatatypeLabel: "Section",
		DatatypeType:  "container",
	}

	got := d.MapGetContentTreeByRouteRow(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != input.FirstChildID {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, input.FirstChildID)
	}
	if got.NextSiblingID != input.NextSiblingID {
		t.Errorf("NextSiblingID = %v, want %v", got.NextSiblingID, input.NextSiblingID)
	}
	if got.PrevSiblingID != input.PrevSiblingID {
		t.Errorf("PrevSiblingID = %v, want %v", got.PrevSiblingID, input.PrevSiblingID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
	if got.Status != types.ContentStatus("draft") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("draft"))
	}
	if got.DatatypeLabel != "Section" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "Section")
	}
	if got.DatatypeType != "container" {
		t.Errorf("DatatypeType = %q, want %q", got.DatatypeType, "container")
	}
}

func TestDatabase_MapGetContentTreeByRouteRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapGetContentTreeByRouteRow(mdb.GetContentTreeByRouteRow{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %v, want zero value", got.ContentDataID)
	}
	if got.ParentID.Valid {
		t.Errorf("ParentID.Valid = true, want false for zero value")
	}
	if got.FirstChildID.Valid {
		t.Errorf("FirstChildID.Valid = true, want false for zero value")
	}
	if got.DatatypeID.Valid {
		t.Errorf("DatatypeID.Valid = true, want false for zero value")
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false for zero value")
	}
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false for zero value")
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
	if got.DatatypeLabel != "" {
		t.Errorf("DatatypeLabel = %q, want empty string", got.DatatypeLabel)
	}
	if got.DatatypeType != "" {
		t.Errorf("DatatypeType = %q, want empty string", got.DatatypeType)
	}
}

// --- SQLite Database.MapGetFieldDefinitionsByRouteRow tests ---

func TestDatabase_MapGetFieldDefinitionsByRouteRow_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	fieldID := types.NewFieldID()
	datatypeID := types.NewDatatypeID()

	input := mdb.GetFieldDefinitionsByRouteRow{
		FieldID:    fieldID,
		Label:      "headline",
		Type:       types.FieldType("text"),
		DatatypeID: datatypeID,
	}

	got := d.MapGetFieldDefinitionsByRouteRow(input)

	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.Label != "headline" {
		t.Errorf("Label = %q, want %q", got.Label, "headline")
	}
	if got.Type != types.FieldType("text") {
		t.Errorf("Type = %v, want %v", got.Type, types.FieldType("text"))
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
}

func TestDatabase_MapGetFieldDefinitionsByRouteRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapGetFieldDefinitionsByRouteRow(mdb.GetFieldDefinitionsByRouteRow{})

	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %v, want zero value", got.Type)
	}
	if !got.DatatypeID.IsZero() {
		t.Errorf("DatatypeID = %v, want zero value", got.DatatypeID)
	}
}

func TestDatabase_MapGetFieldDefinitionsByRouteRow_EmptyDatatypeID(t *testing.T) {
	t.Parallel()
	d := Database{}
	// A field definition can exist without being bound to a datatype
	input := mdb.GetFieldDefinitionsByRouteRow{
		FieldID:    types.NewFieldID(),
		Label:      "global_field",
		Type:       types.FieldType("number"),
		DatatypeID: types.DatatypeID(""),
	}

	got := d.MapGetFieldDefinitionsByRouteRow(input)

	if !got.DatatypeID.IsZero() {
		t.Errorf("DatatypeID = %v, want zero value for empty DatatypeID", got.DatatypeID)
	}
	if got.Label != "global_field" {
		t.Errorf("Label = %q, want %q", got.Label, "global_field")
	}
}

// --- SQLite Database.MapGetContentFieldsByRouteRow tests ---

func TestDatabase_MapGetContentFieldsByRouteRow_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	contentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}

	input := mdb.GetContentFieldsByRouteRow{
		ContentDataID: contentID,
		FieldID:       fieldID,
		FieldValue:    "Lorem ipsum dolor sit amet",
	}

	got := d.MapGetContentFieldsByRouteRow(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.FieldValue != "Lorem ipsum dolor sit amet" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "Lorem ipsum dolor sit amet")
	}
}

func TestDatabase_MapGetContentFieldsByRouteRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapGetContentFieldsByRouteRow(mdb.GetContentFieldsByRouteRow{})

	if got.ContentDataID.Valid {
		t.Errorf("ContentDataID.Valid = true, want false for zero value")
	}
	if got.FieldID.Valid {
		t.Errorf("FieldID.Valid = true, want false for zero value")
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
}

func TestDatabase_MapGetContentFieldsByRouteRow_NullIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	// Both ContentDataID and FieldID can be nullable
	input := mdb.GetContentFieldsByRouteRow{
		ContentDataID: types.NullableContentID{Valid: false},
		FieldID:       types.NullableFieldID{Valid: false},
		FieldValue:    "value-with-null-ids",
	}

	got := d.MapGetContentFieldsByRouteRow(input)

	if got.ContentDataID.Valid {
		t.Errorf("ContentDataID.Valid = true, want false")
	}
	if got.FieldID.Valid {
		t.Errorf("FieldID.Valid = true, want false")
	}
	if got.FieldValue != "value-with-null-ids" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "value-with-null-ids")
	}
}

// --- MySQL MysqlDatabase mapper tests ---

func TestMysqlDatabase_MapGetRouteTreeByRouteIDRow_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	contentID := types.NewContentID()
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}

	input := mdbm.GetRouteTreeByRouteIDRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "mysql-fc", Valid: true},
		NextSiblingID: sql.NullString{String: "mysql-ns", Valid: true},
		PrevSiblingID: sql.NullString{String: "mysql-ps", Valid: true},
		DatatypeLabel: "Product",
		DatatypeType:  "commerce",
		FieldLabel:    "price",
		FieldType:     types.FieldType("number"),
		FieldValue:    sql.NullString{String: "29.99", Valid: true},
	}

	got := d.MapGetRouteTreeByRouteIDRow(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != input.FirstChildID {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, input.FirstChildID)
	}
	if got.NextSiblingID != input.NextSiblingID {
		t.Errorf("NextSiblingID = %v, want %v", got.NextSiblingID, input.NextSiblingID)
	}
	if got.PrevSiblingID != input.PrevSiblingID {
		t.Errorf("PrevSiblingID = %v, want %v", got.PrevSiblingID, input.PrevSiblingID)
	}
	if got.DatatypeLabel != "Product" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "Product")
	}
	if got.DatatypeType != "commerce" {
		t.Errorf("DatatypeType = %q, want %q", got.DatatypeType, "commerce")
	}
	if got.FieldLabel != "price" {
		t.Errorf("FieldLabel = %q, want %q", got.FieldLabel, "price")
	}
	if got.FieldType != types.FieldType("number") {
		t.Errorf("FieldType = %v, want %v", got.FieldType, types.FieldType("number"))
	}
	if got.FieldValue != input.FieldValue {
		t.Errorf("FieldValue = %v, want %v", got.FieldValue, input.FieldValue)
	}
}

func TestMysqlDatabase_MapGetRouteTreeByRouteIDRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapGetRouteTreeByRouteIDRow(mdbm.GetRouteTreeByRouteIDRow{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %v, want zero value", got.ContentDataID)
	}
	if got.ParentID.Valid {
		t.Errorf("ParentID.Valid = true, want false for zero value")
	}
	if got.FieldType != "" {
		t.Errorf("FieldType = %v, want zero value", got.FieldType)
	}
	if got.FieldValue.Valid {
		t.Errorf("FieldValue.Valid = true, want false for zero value")
	}
}

func TestMysqlDatabase_MapGetContentTreeByRouteRow_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 9, 0, 0, 0, time.UTC))
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbm.GetContentTreeByRouteRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "mysql-child", Valid: true},
		NextSiblingID: sql.NullString{String: "mysql-next", Valid: true},
		PrevSiblingID: sql.NullString{String: "mysql-prev", Valid: true},
		DatatypeID:    datatypeID,
		RouteID:       routeID,
		AuthorID:      authorID,
		DateCreated:   ts,
		DateModified:  ts,
		Status:        types.ContentStatus("published"),
		DatatypeLabel: "MySQLPage",
		DatatypeType:  "mysql-type",
	}

	got := d.MapGetContentTreeByRouteRow(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
	if got.Status != types.ContentStatus("published") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("published"))
	}
	if got.DatatypeLabel != "MySQLPage" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "MySQLPage")
	}
	if got.DatatypeType != "mysql-type" {
		t.Errorf("DatatypeType = %q, want %q", got.DatatypeType, "mysql-type")
	}
}

func TestMysqlDatabase_MapGetContentTreeByRouteRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapGetContentTreeByRouteRow(mdbm.GetContentTreeByRouteRow{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %v, want zero value", got.ContentDataID)
	}
	if got.DatatypeID.Valid {
		t.Errorf("DatatypeID.Valid = true, want false for zero value")
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false for zero value")
	}
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false for zero value")
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
}

func TestMysqlDatabase_MapGetFieldDefinitionsByRouteRow_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	fieldID := types.NewFieldID()
	datatypeID := types.NewDatatypeID()

	input := mdbm.GetFieldDefinitionsByRouteRow{
		FieldID:    fieldID,
		Label:      "mysql_field",
		Type:       types.FieldType("textarea"),
		DatatypeID: datatypeID,
	}

	got := d.MapGetFieldDefinitionsByRouteRow(input)

	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.Label != "mysql_field" {
		t.Errorf("Label = %q, want %q", got.Label, "mysql_field")
	}
	if got.Type != types.FieldType("textarea") {
		t.Errorf("Type = %v, want %v", got.Type, types.FieldType("textarea"))
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
}

func TestMysqlDatabase_MapGetFieldDefinitionsByRouteRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapGetFieldDefinitionsByRouteRow(mdbm.GetFieldDefinitionsByRouteRow{})

	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %v, want zero value", got.Type)
	}
	if !got.DatatypeID.IsZero() {
		t.Errorf("DatatypeID = %v, want zero value", got.DatatypeID)
	}
}

func TestMysqlDatabase_MapGetContentFieldsByRouteRow_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	contentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}

	input := mdbm.GetContentFieldsByRouteRow{
		ContentDataID: contentID,
		FieldID:       fieldID,
		FieldValue:    "mysql field value",
	}

	got := d.MapGetContentFieldsByRouteRow(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.FieldValue != "mysql field value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "mysql field value")
	}
}

func TestMysqlDatabase_MapGetContentFieldsByRouteRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapGetContentFieldsByRouteRow(mdbm.GetContentFieldsByRouteRow{})

	if got.ContentDataID.Valid {
		t.Errorf("ContentDataID.Valid = true, want false for zero value")
	}
	if got.FieldID.Valid {
		t.Errorf("FieldID.Valid = true, want false for zero value")
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
}

// --- PostgreSQL PsqlDatabase mapper tests ---

func TestPsqlDatabase_MapGetRouteTreeByRouteIDRow_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	contentID := types.NewContentID()
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}

	input := mdbp.GetRouteTreeByRouteIDRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "psql-fc", Valid: true},
		NextSiblingID: sql.NullString{String: "psql-ns", Valid: true},
		PrevSiblingID: sql.NullString{String: "psql-ps", Valid: true},
		DatatypeLabel: "Category",
		DatatypeType:  "taxonomy",
		FieldLabel:    "slug",
		FieldType:     types.FieldType("slug"),
		FieldValue:    sql.NullString{String: "my-category", Valid: true},
	}

	got := d.MapGetRouteTreeByRouteIDRow(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != input.FirstChildID {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, input.FirstChildID)
	}
	if got.NextSiblingID != input.NextSiblingID {
		t.Errorf("NextSiblingID = %v, want %v", got.NextSiblingID, input.NextSiblingID)
	}
	if got.PrevSiblingID != input.PrevSiblingID {
		t.Errorf("PrevSiblingID = %v, want %v", got.PrevSiblingID, input.PrevSiblingID)
	}
	if got.DatatypeLabel != "Category" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "Category")
	}
	if got.DatatypeType != "taxonomy" {
		t.Errorf("DatatypeType = %q, want %q", got.DatatypeType, "taxonomy")
	}
	if got.FieldLabel != "slug" {
		t.Errorf("FieldLabel = %q, want %q", got.FieldLabel, "slug")
	}
	if got.FieldType != types.FieldType("slug") {
		t.Errorf("FieldType = %v, want %v", got.FieldType, types.FieldType("slug"))
	}
	if got.FieldValue != input.FieldValue {
		t.Errorf("FieldValue = %v, want %v", got.FieldValue, input.FieldValue)
	}
}

func TestPsqlDatabase_MapGetRouteTreeByRouteIDRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapGetRouteTreeByRouteIDRow(mdbp.GetRouteTreeByRouteIDRow{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %v, want zero value", got.ContentDataID)
	}
	if got.ParentID.Valid {
		t.Errorf("ParentID.Valid = true, want false for zero value")
	}
	if got.FieldType != "" {
		t.Errorf("FieldType = %v, want zero value", got.FieldType)
	}
	if got.FieldValue.Valid {
		t.Errorf("FieldValue.Valid = true, want false for zero value")
	}
}

func TestPsqlDatabase_MapGetContentTreeByRouteRow_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 9, 1, 16, 45, 0, 0, time.UTC))
	parentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbp.GetContentTreeByRouteRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  sql.NullString{String: "psql-child", Valid: true},
		NextSiblingID: sql.NullString{String: "psql-next", Valid: true},
		PrevSiblingID: sql.NullString{String: "psql-prev", Valid: true},
		DatatypeID:    datatypeID,
		RouteID:       routeID,
		AuthorID:      authorID,
		DateCreated:   ts,
		DateModified:  ts,
		Status:        types.ContentStatus("archived"),
		DatatypeLabel: "PsqlPage",
		DatatypeType:  "psql-type",
	}

	got := d.MapGetContentTreeByRouteRow(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
	if got.Status != types.ContentStatus("archived") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("archived"))
	}
	if got.DatatypeLabel != "PsqlPage" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "PsqlPage")
	}
	if got.DatatypeType != "psql-type" {
		t.Errorf("DatatypeType = %q, want %q", got.DatatypeType, "psql-type")
	}
}

func TestPsqlDatabase_MapGetContentTreeByRouteRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapGetContentTreeByRouteRow(mdbp.GetContentTreeByRouteRow{})

	if got.ContentDataID != "" {
		t.Errorf("ContentDataID = %v, want zero value", got.ContentDataID)
	}
	if got.DatatypeID.Valid {
		t.Errorf("DatatypeID.Valid = true, want false for zero value")
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false for zero value")
	}
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false for zero value")
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
}

func TestPsqlDatabase_MapGetFieldDefinitionsByRouteRow_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	fieldID := types.NewFieldID()
	datatypeID := types.NewDatatypeID()

	input := mdbp.GetFieldDefinitionsByRouteRow{
		FieldID:    fieldID,
		Label:      "psql_field",
		Type:       types.FieldType("boolean"),
		DatatypeID: datatypeID,
	}

	got := d.MapGetFieldDefinitionsByRouteRow(input)

	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.Label != "psql_field" {
		t.Errorf("Label = %q, want %q", got.Label, "psql_field")
	}
	if got.Type != types.FieldType("boolean") {
		t.Errorf("Type = %v, want %v", got.Type, types.FieldType("boolean"))
	}
	if got.DatatypeID != datatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, datatypeID)
	}
}

func TestPsqlDatabase_MapGetFieldDefinitionsByRouteRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapGetFieldDefinitionsByRouteRow(mdbp.GetFieldDefinitionsByRouteRow{})

	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %v, want zero value", got.Type)
	}
	if !got.DatatypeID.IsZero() {
		t.Errorf("DatatypeID = %v, want zero value", got.DatatypeID)
	}
}

func TestPsqlDatabase_MapGetContentFieldsByRouteRow_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	contentID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}

	input := mdbp.GetContentFieldsByRouteRow{
		ContentDataID: contentID,
		FieldID:       fieldID,
		FieldValue:    "psql field value",
	}

	got := d.MapGetContentFieldsByRouteRow(input)

	if got.ContentDataID != contentID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.FieldValue != "psql field value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "psql field value")
	}
}

func TestPsqlDatabase_MapGetContentFieldsByRouteRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapGetContentFieldsByRouteRow(mdbp.GetContentFieldsByRouteRow{})

	if got.ContentDataID.Valid {
		t.Errorf("ContentDataID.Valid = true, want false for zero value")
	}
	if got.FieldID.Valid {
		t.Errorf("FieldID.Valid = true, want false for zero value")
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
}

// --- Cross-database mapper consistency tests ---
// All three drivers should produce identical wrapper structs from equivalent input.

func TestCrossDatabase_MapGetRouteTreeByRouteIDRow_Consistency(t *testing.T) {
	t.Parallel()
	contentID := types.NewContentID()
	parentID := types.NullableContentID{ID: types.ContentID("parent-cross"), Valid: true}
	firstChild := sql.NullString{String: "cross-fc", Valid: true}
	nextSibling := sql.NullString{String: "cross-ns", Valid: true}
	prevSibling := sql.NullString{String: "cross-ps", Valid: true}
	fieldValue := sql.NullString{String: "cross-val", Valid: true}

	sqliteInput := mdb.GetRouteTreeByRouteIDRow{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		DatatypeLabel: "Cross", DatatypeType: "type",
		FieldLabel: "label", FieldType: types.FieldType("text"), FieldValue: fieldValue,
	}
	mysqlInput := mdbm.GetRouteTreeByRouteIDRow{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		DatatypeLabel: "Cross", DatatypeType: "type",
		FieldLabel: "label", FieldType: types.FieldType("text"), FieldValue: fieldValue,
	}
	psqlInput := mdbp.GetRouteTreeByRouteIDRow{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		DatatypeLabel: "Cross", DatatypeType: "type",
		FieldLabel: "label", FieldType: types.FieldType("text"), FieldValue: fieldValue,
	}

	sqliteResult := Database{}.MapGetRouteTreeByRouteIDRow(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapGetRouteTreeByRouteIDRow(mysqlInput)
	psqlResult := PsqlDatabase{}.MapGetRouteTreeByRouteIDRow(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

func TestCrossDatabase_MapGetContentTreeByRouteRow_Consistency(t *testing.T) {
	t.Parallel()
	contentID := types.NewContentID()
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 12, 0, 0, 0, time.UTC))
	parentID := types.NullableContentID{ID: types.ContentID("parent-cross"), Valid: true}
	datatypeID := types.NullableDatatypeID{ID: types.DatatypeID("dt-cross"), Valid: true}
	routeID := types.NullableRouteID{ID: types.RouteID("route-cross"), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	firstChild := sql.NullString{String: "cross-child", Valid: true}
	nextSibling := sql.NullString{String: "cross-next", Valid: true}
	prevSibling := sql.NullString{String: "cross-prev", Valid: true}

	sqliteInput := mdb.GetContentTreeByRouteRow{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		DatatypeID: datatypeID, RouteID: routeID, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
		Status: types.ContentStatus("published"), DatatypeLabel: "Cross", DatatypeType: "type",
	}
	mysqlInput := mdbm.GetContentTreeByRouteRow{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		DatatypeID: datatypeID, RouteID: routeID, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
		Status: types.ContentStatus("published"), DatatypeLabel: "Cross", DatatypeType: "type",
	}
	psqlInput := mdbp.GetContentTreeByRouteRow{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		DatatypeID: datatypeID, RouteID: routeID, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
		Status: types.ContentStatus("published"), DatatypeLabel: "Cross", DatatypeType: "type",
	}

	sqliteResult := Database{}.MapGetContentTreeByRouteRow(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapGetContentTreeByRouteRow(mysqlInput)
	psqlResult := PsqlDatabase{}.MapGetContentTreeByRouteRow(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

func TestCrossDatabase_MapGetFieldDefinitionsByRouteRow_Consistency(t *testing.T) {
	t.Parallel()
	fieldID := types.NewFieldID()
	datatypeID := types.DatatypeID("dt-cross")

	sqliteInput := mdb.GetFieldDefinitionsByRouteRow{
		FieldID: fieldID, Label: "cross_field", Type: types.FieldType("date"), DatatypeID: datatypeID,
	}
	mysqlInput := mdbm.GetFieldDefinitionsByRouteRow{
		FieldID: fieldID, Label: "cross_field", Type: types.FieldType("date"), DatatypeID: datatypeID,
	}
	psqlInput := mdbp.GetFieldDefinitionsByRouteRow{
		FieldID: fieldID, Label: "cross_field", Type: types.FieldType("date"), DatatypeID: datatypeID,
	}

	sqliteResult := Database{}.MapGetFieldDefinitionsByRouteRow(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapGetFieldDefinitionsByRouteRow(mysqlInput)
	psqlResult := PsqlDatabase{}.MapGetFieldDefinitionsByRouteRow(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

func TestCrossDatabase_MapGetContentFieldsByRouteRow_Consistency(t *testing.T) {
	t.Parallel()
	contentID := types.NullableContentID{ID: types.ContentID("content-cross"), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}

	sqliteInput := mdb.GetContentFieldsByRouteRow{
		ContentDataID: contentID, FieldID: fieldID, FieldValue: "cross-value",
	}
	mysqlInput := mdbm.GetContentFieldsByRouteRow{
		ContentDataID: contentID, FieldID: fieldID, FieldValue: "cross-value",
	}
	psqlInput := mdbp.GetContentFieldsByRouteRow{
		ContentDataID: contentID, FieldID: fieldID, FieldValue: "cross-value",
	}

	sqliteResult := Database{}.MapGetContentFieldsByRouteRow(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapGetContentFieldsByRouteRow(mysqlInput)
	psqlResult := PsqlDatabase{}.MapGetContentFieldsByRouteRow(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- JSON tag verification for all wrapper structs ---

func TestGetRouteTreeByRouteIDRow_JSONTags(t *testing.T) {
	t.Parallel()
	row, _ := routeTreeFixture()

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"content_data_id", "parent_id", "first_child_id",
		"next_sibling_id", "prev_sibling_id",
		"datatype_label", "datatype_type",
		"field_label", "field_type", "field_value",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

func TestGetContentTreeByRouteRow_JSONTags(t *testing.T) {
	t.Parallel()
	row, _, _ := contentTreeFixture()

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"content_data_id", "parent_id", "first_child_id",
		"next_sibling_id", "prev_sibling_id",
		"datatype_id", "route_id", "author_id",
		"date_created", "date_modified", "status",
		"datatype_label", "datatype_type",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

func TestGetContentFieldsByRouteRow_JSONTags(t *testing.T) {
	t.Parallel()
	row := GetContentFieldsByRouteRow{
		ContentDataID: types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FieldID:       types.NullableFieldID{ID: types.NewFieldID(), Valid: true},
		FieldValue:    "test-value",
	}

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"content_data_id", "field_id", "field_value",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

func TestGetFieldDefinitionsByRouteRow_JSONTags(t *testing.T) {
	t.Parallel()
	row := GetFieldDefinitionsByRouteRow{
		FieldID:    types.NewFieldID(),
		Label:      "test_label",
		Type:       types.FieldType("text"),
		DatatypeID: types.NewDatatypeID(),
	}

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"field_id", "label", "type", "datatype_id",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

// --- Edge case: NullString sibling fields in route tree ---
// Tests mixed Valid/Invalid NullString states across all tree pointer fields.

func TestDatabase_MapGetRouteTreeByRouteIDRow_MixedNullStrings(t *testing.T) {
	t.Parallel()
	d := Database{}

	tests := []struct {
		name          string
		firstChild    sql.NullString
		nextSibling   sql.NullString
		prevSibling   sql.NullString
		fieldValue    sql.NullString
		wantFCValid   bool
		wantNSValid   bool
		wantPSValid   bool
		wantFVValid   bool
	}{
		{
			name:        "all valid",
			firstChild:  sql.NullString{String: "fc", Valid: true},
			nextSibling: sql.NullString{String: "ns", Valid: true},
			prevSibling: sql.NullString{String: "ps", Valid: true},
			fieldValue:  sql.NullString{String: "fv", Valid: true},
			wantFCValid: true, wantNSValid: true, wantPSValid: true, wantFVValid: true,
		},
		{
			name:        "all null",
			firstChild:  sql.NullString{Valid: false},
			nextSibling: sql.NullString{Valid: false},
			prevSibling: sql.NullString{Valid: false},
			fieldValue:  sql.NullString{Valid: false},
			wantFCValid: false, wantNSValid: false, wantPSValid: false, wantFVValid: false,
		},
		{
			name:        "first child valid only",
			firstChild:  sql.NullString{String: "only-fc", Valid: true},
			nextSibling: sql.NullString{Valid: false},
			prevSibling: sql.NullString{Valid: false},
			fieldValue:  sql.NullString{Valid: false},
			wantFCValid: true, wantNSValid: false, wantPSValid: false, wantFVValid: false,
		},
		{
			name:        "field value valid only",
			firstChild:  sql.NullString{Valid: false},
			nextSibling: sql.NullString{Valid: false},
			prevSibling: sql.NullString{Valid: false},
			fieldValue:  sql.NullString{String: "only-fv", Valid: true},
			wantFCValid: false, wantNSValid: false, wantPSValid: false, wantFVValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdb.GetRouteTreeByRouteIDRow{
				ContentDataID: types.NewContentID(),
				FirstChildID:  tt.firstChild,
				NextSiblingID: tt.nextSibling,
				PrevSiblingID: tt.prevSibling,
				FieldValue:    tt.fieldValue,
			}

			got := d.MapGetRouteTreeByRouteIDRow(input)

			if got.FirstChildID.Valid != tt.wantFCValid {
				t.Errorf("FirstChildID.Valid = %v, want %v", got.FirstChildID.Valid, tt.wantFCValid)
			}
			if got.NextSiblingID.Valid != tt.wantNSValid {
				t.Errorf("NextSiblingID.Valid = %v, want %v", got.NextSiblingID.Valid, tt.wantNSValid)
			}
			if got.PrevSiblingID.Valid != tt.wantPSValid {
				t.Errorf("PrevSiblingID.Valid = %v, want %v", got.PrevSiblingID.Valid, tt.wantPSValid)
			}
			if got.FieldValue.Valid != tt.wantFVValid {
				t.Errorf("FieldValue.Valid = %v, want %v", got.FieldValue.Valid, tt.wantFVValid)
			}
		})
	}
}

// --- Edge case: content tree with all nullable IDs invalid ---

func TestDatabase_MapGetContentTreeByRouteRow_AllNullableIDsInvalid(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	input := mdb.GetContentTreeByRouteRow{
		ContentDataID: types.NewContentID(),
		ParentID:      types.NullableContentID{Valid: false},
		FirstChildID:  sql.NullString{Valid: false},
		NextSiblingID: sql.NullString{Valid: false},
		PrevSiblingID: sql.NullString{Valid: false},
		DatatypeID:    types.NullableDatatypeID{Valid: false},
		RouteID:       types.NullableRouteID{Valid: false},
		AuthorID:      types.NullableUserID{Valid: false},
		DateCreated:   ts,
		DateModified:  ts,
		Status:        types.ContentStatus("draft"),
		DatatypeLabel: "Orphan",
		DatatypeType:  "orphan_type",
	}

	got := d.MapGetContentTreeByRouteRow(input)

	if got.ParentID.Valid {
		t.Errorf("ParentID.Valid = true, want false")
	}
	if got.FirstChildID.Valid {
		t.Errorf("FirstChildID.Valid = true, want false")
	}
	if got.NextSiblingID.Valid {
		t.Errorf("NextSiblingID.Valid = true, want false")
	}
	if got.PrevSiblingID.Valid {
		t.Errorf("PrevSiblingID.Valid = true, want false")
	}
	if got.DatatypeID.Valid {
		t.Errorf("DatatypeID.Valid = true, want false")
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false")
	}
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false")
	}
	// Non-nullable fields should still be mapped correctly
	if got.DatatypeLabel != "Orphan" {
		t.Errorf("DatatypeLabel = %q, want %q", got.DatatypeLabel, "Orphan")
	}
	if got.DatatypeType != "orphan_type" {
		t.Errorf("DatatypeType = %q, want %q", got.DatatypeType, "orphan_type")
	}
	if got.Status != types.ContentStatus("draft") {
		t.Errorf("Status = %v, want %v", got.Status, types.ContentStatus("draft"))
	}
}

// --- Edge case: field types via table-driven test ---
// Verify various FieldType values pass through the mapper correctly.

func TestDatabase_MapGetRouteTreeByRouteIDRow_FieldTypeVariants(t *testing.T) {
	t.Parallel()
	d := Database{}

	fieldTypes := []types.FieldType{
		"text", "textarea", "number", "date", "datetime",
		"boolean", "select", "media", "relation", "json",
		"richtext", "slug", "email", "url", "",
	}

	for _, ft := range fieldTypes {
		t.Run(string(ft), func(t *testing.T) {
			t.Parallel()
			input := mdb.GetRouteTreeByRouteIDRow{
				ContentDataID: types.NewContentID(),
				FieldType:     ft,
			}

			got := d.MapGetRouteTreeByRouteIDRow(input)

			if got.FieldType != ft {
				t.Errorf("FieldType = %v, want %v", got.FieldType, ft)
			}
		})
	}
}

// --- Edge case: ContentStatus variants in content tree mapper ---

func TestDatabase_MapGetContentTreeByRouteRow_StatusVariants(t *testing.T) {
	t.Parallel()
	d := Database{}

	statuses := []types.ContentStatus{
		"draft", "published", "archived", "pending", "",
	}

	for _, s := range statuses {
		t.Run(string(s), func(t *testing.T) {
			t.Parallel()
			input := mdb.GetContentTreeByRouteRow{
				ContentDataID: types.NewContentID(),
				Status:        s,
			}

			got := d.MapGetContentTreeByRouteRow(input)

			if got.Status != s {
				t.Errorf("Status = %v, want %v", got.Status, s)
			}
		})
	}
}

// --- Edge case: empty string FieldValue (distinct from null) ---

func TestDatabase_MapGetContentFieldsByRouteRow_EmptyStringValue(t *testing.T) {
	t.Parallel()
	d := Database{}

	// FieldValue is a non-nullable string -- empty string is a valid value
	input := mdb.GetContentFieldsByRouteRow{
		ContentDataID: types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FieldID:       types.NullableFieldID{ID: types.NewFieldID(), Valid: true},
		FieldValue:    "",
	}

	got := d.MapGetContentFieldsByRouteRow(input)

	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
	// IDs should still be mapped correctly
	if !got.ContentDataID.Valid {
		t.Errorf("ContentDataID.Valid = false, want true")
	}
	if !got.FieldID.Valid {
		t.Errorf("FieldID.Valid = false, want true")
	}
}

// --- Edge case: ValidButEmpty NullString in route tree ---
// sql.NullString{String: "", Valid: true} is a valid non-null empty string.

func TestDatabase_MapGetRouteTreeByRouteIDRow_ValidEmptyNullStrings(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := mdb.GetRouteTreeByRouteIDRow{
		ContentDataID: types.NewContentID(),
		FirstChildID:  sql.NullString{String: "", Valid: true},
		NextSiblingID: sql.NullString{String: "", Valid: true},
		PrevSiblingID: sql.NullString{String: "", Valid: true},
		FieldValue:    sql.NullString{String: "", Valid: true},
	}

	got := d.MapGetRouteTreeByRouteIDRow(input)

	// Valid=true should be preserved even with empty string
	if !got.FirstChildID.Valid {
		t.Errorf("FirstChildID.Valid = false, want true for Valid=true empty string")
	}
	if got.FirstChildID.String != "" {
		t.Errorf("FirstChildID.String = %q, want empty string", got.FirstChildID.String)
	}
	if !got.NextSiblingID.Valid {
		t.Errorf("NextSiblingID.Valid = false, want true for Valid=true empty string")
	}
	if !got.PrevSiblingID.Valid {
		t.Errorf("PrevSiblingID.Valid = false, want true for Valid=true empty string")
	}
	if !got.FieldValue.Valid {
		t.Errorf("FieldValue.Valid = false, want true for Valid=true empty string")
	}
}
