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
		FirstChildID:  types.NullableContentID{ID: types.ContentID("fc-tree-001"), Valid: true},
		NextSiblingID: types.NullableContentID{ID: types.ContentID("ns-tree-001"), Valid: true},
		PrevSiblingID: types.NullableContentID{ID: types.ContentID("ps-tree-001"), Valid: true},
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
		FirstChildID:  types.NullableContentID{ID: types.ContentID("fc-ct-001"), Valid: true},
		NextSiblingID: types.NullableContentID{ID: types.ContentID("ns-ct-001"), Valid: true},
		PrevSiblingID: types.NullableContentID{ID: types.ContentID("ps-ct-001"), Valid: true},
		DatatypeID:    types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true},
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		AuthorID:      types.NewUserID(),
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
		FirstChildID:  types.NullableContentID{ID: types.ContentID("child-1"), Valid: true},
		NextSiblingID: types.NullableContentID{ID: types.ContentID("next-1"), Valid: true},
		PrevSiblingID: types.NullableContentID{ID: types.ContentID("prev-1"), Valid: true},
		DatatypeLabel: "Blog Post",
		DatatypeType:  "dynamic",
		FieldLabel:    "body",
		FieldTypes:    types.FieldType("richtext"),
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
		FieldTypes:    types.FieldType("text"),
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
	authorID := types.NewUserID()

	input := mdb.GetContentTreeByRouteRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  types.NullableContentID{ID: types.ContentID("child-ct"), Valid: true},
		NextSiblingID: types.NullableContentID{ID: types.ContentID("next-ct"), Valid: true},
		PrevSiblingID: types.NullableContentID{ID: types.ContentID("prev-ct"), Valid: true},
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
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string for zero value", got.AuthorID)
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
		FirstChildID:  types.NullableContentID{ID: types.ContentID("mysql-fc"), Valid: true},
		NextSiblingID: types.NullableContentID{ID: types.ContentID("mysql-ns"), Valid: true},
		PrevSiblingID: types.NullableContentID{ID: types.ContentID("mysql-ps"), Valid: true},
		DatatypeLabel: "Product",
		DatatypeType:  "commerce",
		FieldLabel:    "price",
		FieldTypes:    types.FieldType("number"),
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
	authorID := types.NewUserID()

	input := mdbm.GetContentTreeByRouteRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  types.NullableContentID{ID: types.ContentID("mysql-child"), Valid: true},
		NextSiblingID: types.NullableContentID{ID: types.ContentID("mysql-next"), Valid: true},
		PrevSiblingID: types.NullableContentID{ID: types.ContentID("mysql-prev"), Valid: true},
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
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string for zero value", got.AuthorID)
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
		FirstChildID:  types.NullableContentID{ID: types.ContentID("psql-fc"), Valid: true},
		NextSiblingID: types.NullableContentID{ID: types.ContentID("psql-ns"), Valid: true},
		PrevSiblingID: types.NullableContentID{ID: types.ContentID("psql-ps"), Valid: true},
		DatatypeLabel: "Category",
		DatatypeType:  "taxonomy",
		FieldLabel:    "slug",
		FieldTypes:    types.FieldType("slug"),
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
	authorID := types.NewUserID()

	input := mdbp.GetContentTreeByRouteRow{
		ContentDataID: contentID,
		ParentID:      parentID,
		FirstChildID:  types.NullableContentID{ID: types.ContentID("psql-child"), Valid: true},
		NextSiblingID: types.NullableContentID{ID: types.ContentID("psql-next"), Valid: true},
		PrevSiblingID: types.NullableContentID{ID: types.ContentID("psql-prev"), Valid: true},
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
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string for zero value", got.AuthorID)
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
	firstChild := types.NullableContentID{ID: types.ContentID("cross-fc"), Valid: true}
	nextSibling := types.NullableContentID{ID: types.ContentID("cross-ns"), Valid: true}
	prevSibling := types.NullableContentID{ID: types.ContentID("cross-ps"), Valid: true}
	fieldValue := sql.NullString{String: "cross-val", Valid: true}

	sqliteInput := mdb.GetRouteTreeByRouteIDRow{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		DatatypeLabel: "Cross", DatatypeType: "type",
		FieldLabel: "label", FieldTypes: types.FieldType("text"), FieldValue: fieldValue,
	}
	mysqlInput := mdbm.GetRouteTreeByRouteIDRow{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		DatatypeLabel: "Cross", DatatypeType: "type",
		FieldLabel: "label", FieldTypes: types.FieldType("text"), FieldValue: fieldValue,
	}
	psqlInput := mdbp.GetRouteTreeByRouteIDRow{
		ContentDataID: contentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		DatatypeLabel: "Cross", DatatypeType: "type",
		FieldLabel: "label", FieldTypes: types.FieldType("text"), FieldValue: fieldValue,
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
	authorID := types.NewUserID()
	firstChild := types.NullableContentID{ID: types.ContentID("cross-child"), Valid: true}
	nextSibling := types.NullableContentID{ID: types.ContentID("cross-next"), Valid: true}
	prevSibling := types.NullableContentID{ID: types.ContentID("cross-prev"), Valid: true}

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

// --- Edge case: nullable sibling fields in route tree ---
// Tests mixed Valid/Invalid NullableContentID states across all tree pointer fields.

func TestDatabase_MapGetRouteTreeByRouteIDRow_MixedNullStrings(t *testing.T) {
	t.Parallel()
	d := Database{}

	tests := []struct {
		name          string
		firstChild    types.NullableContentID
		nextSibling   types.NullableContentID
		prevSibling   types.NullableContentID
		fieldValue    sql.NullString
		wantFCValid   bool
		wantNSValid   bool
		wantPSValid   bool
		wantFVValid   bool
	}{
		{
			name:        "all valid",
			firstChild:  types.NullableContentID{ID: types.ContentID("fc"), Valid: true},
			nextSibling: types.NullableContentID{ID: types.ContentID("ns"), Valid: true},
			prevSibling: types.NullableContentID{ID: types.ContentID("ps"), Valid: true},
			fieldValue:  sql.NullString{String: "fv", Valid: true},
			wantFCValid: true, wantNSValid: true, wantPSValid: true, wantFVValid: true,
		},
		{
			name:        "all null",
			firstChild:  types.NullableContentID{},
			nextSibling: types.NullableContentID{},
			prevSibling: types.NullableContentID{},
			fieldValue:  sql.NullString{Valid: false},
			wantFCValid: false, wantNSValid: false, wantPSValid: false, wantFVValid: false,
		},
		{
			name:        "first child valid only",
			firstChild:  types.NullableContentID{ID: types.ContentID("only-fc"), Valid: true},
			nextSibling: types.NullableContentID{},
			prevSibling: types.NullableContentID{},
			fieldValue:  sql.NullString{Valid: false},
			wantFCValid: true, wantNSValid: false, wantPSValid: false, wantFVValid: false,
		},
		{
			name:        "field value valid only",
			firstChild:  types.NullableContentID{},
			nextSibling: types.NullableContentID{},
			prevSibling: types.NullableContentID{},
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
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    types.NullableDatatypeID{Valid: false},
		RouteID:       types.NullableRouteID{Valid: false},
		AuthorID:      types.UserID(""),
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
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string for zero UserID", got.AuthorID)
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
				FieldTypes:    ft,
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

// --- Edge case: ValidButEmpty nullable IDs in route tree ---
// NullableContentID{ID: "", Valid: true} is a valid non-null empty ID.

func TestDatabase_MapGetRouteTreeByRouteIDRow_ValidEmptyNullStrings(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := mdb.GetRouteTreeByRouteIDRow{
		ContentDataID: types.NewContentID(),
		FirstChildID:  types.NullableContentID{ID: types.ContentID(""), Valid: true},
		NextSiblingID: types.NullableContentID{ID: types.ContentID(""), Valid: true},
		PrevSiblingID: types.NullableContentID{ID: types.ContentID(""), Valid: true},
		FieldValue:    sql.NullString{String: "", Valid: true},
	}

	got := d.MapGetRouteTreeByRouteIDRow(input)

	// Valid=true should be preserved even with empty string
	if !got.FirstChildID.Valid {
		t.Errorf("FirstChildID.Valid = false, want true for Valid=true empty string")
	}
	if got.FirstChildID.ID != "" {
		t.Errorf("FirstChildID.ID = %q, want empty string", got.FirstChildID.ID)
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

// =============================================================================
// MapAdminContentDataWithDatatypeRow -- all 3 backends
// =============================================================================

// adminContentDataFixture returns shared test data for MapAdminContentDataWithDatatypeRow tests.
func adminContentDataWithDatatypeFixture() (
	types.AdminContentID,
	types.NullableAdminContentID,
	types.NullableAdminContentID,
	types.NullableAdminContentID,
	types.NullableAdminContentID,
	types.NullableAdminRouteID,
	types.NullableAdminDatatypeID,
	types.UserID,
	types.ContentStatus,
	types.Timestamp,
	types.AdminDatatypeID,
	types.NullableAdminDatatypeID,
	types.UserID,
) {
	ts := types.NewTimestamp(time.Date(2025, 11, 5, 10, 30, 0, 0, time.UTC))
	return types.NewAdminContentID(),
		types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		types.NullableAdminRouteID{ID: types.NewAdminRouteID(), Valid: true},
		types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true},
		types.NewUserID(),
		types.ContentStatus("published"),
		ts,
		types.NewAdminDatatypeID(),
		types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true},
		types.NewUserID()
}

func TestDatabase_MapAdminContentDataWithDatatypeRow_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	adminContentID, parentID, firstChild, nextSibling, prevSibling,
		adminRouteID, adminDatatypeID, authorID, status, ts,
		dtAdminDatatypeID, dtParentID, dtAuthorID := adminContentDataWithDatatypeFixture()

	input := mdb.ListAdminContentDataWithDatatypeByRouteRow{
		AdminContentDataID: adminContentID,
		ParentID:           parentID,
		FirstChildID:       firstChild,
		NextSiblingID:      nextSibling,
		PrevSiblingID:      prevSibling,
		AdminRouteID:       adminRouteID,
		AdminDatatypeID:    adminDatatypeID,
		AuthorID:           authorID,
		Status:             status,
		DateCreated:        ts,
		DateModified:       ts,
		DtAdminDatatypeId:  dtAdminDatatypeID,
		DtParentId:         dtParentID,
		DtLabel:            "admin-page",
		DtType:             "admin-type",
		DtAuthorId:         dtAuthorID,
		DtDateCreated:      ts,
		DtDateModified:     ts,
	}

	got := d.MapAdminContentDataWithDatatypeRow(input)

	if got.AdminContentDataID != adminContentID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, adminContentID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.FirstChildID != firstChild {
		t.Errorf("FirstChildID = %v, want %v", got.FirstChildID, firstChild)
	}
	if got.NextSiblingID != nextSibling {
		t.Errorf("NextSiblingID = %v, want %v", got.NextSiblingID, nextSibling)
	}
	if got.PrevSiblingID != prevSibling {
		t.Errorf("PrevSiblingID = %v, want %v", got.PrevSiblingID, prevSibling)
	}
	if got.AdminRouteID != adminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, adminRouteID)
	}
	if got.AdminDatatypeID != adminDatatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, adminDatatypeID)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
	if got.Status != status {
		t.Errorf("Status = %v, want %v", got.Status, status)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
	if got.DtAdminDatatypeID != dtAdminDatatypeID {
		t.Errorf("DtAdminDatatypeID = %v, want %v", got.DtAdminDatatypeID, dtAdminDatatypeID)
	}
	if got.DtParentID != dtParentID {
		t.Errorf("DtParentID = %v, want %v", got.DtParentID, dtParentID)
	}
	if got.DtLabel != "admin-page" {
		t.Errorf("DtLabel = %q, want %q", got.DtLabel, "admin-page")
	}
	if got.DtType != "admin-type" {
		t.Errorf("DtType = %q, want %q", got.DtType, "admin-type")
	}
	if got.DtAuthorID != dtAuthorID {
		t.Errorf("DtAuthorID = %v, want %v", got.DtAuthorID, dtAuthorID)
	}
	if got.DtDateCreated != ts {
		t.Errorf("DtDateCreated = %v, want %v", got.DtDateCreated, ts)
	}
	if got.DtDateModified != ts {
		t.Errorf("DtDateModified = %v, want %v", got.DtDateModified, ts)
	}
}

func TestDatabase_MapAdminContentDataWithDatatypeRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminContentDataWithDatatypeRow(mdb.ListAdminContentDataWithDatatypeByRouteRow{})

	if got.AdminContentDataID != "" {
		t.Errorf("AdminContentDataID = %v, want zero value", got.AdminContentDataID)
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
	if got.AdminRouteID.Valid {
		t.Errorf("AdminRouteID.Valid = true, want false for zero value")
	}
	if got.AdminDatatypeID.Valid {
		t.Errorf("AdminDatatypeID.Valid = true, want false for zero value")
	}
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string for zero value", got.AuthorID)
	}
	if got.Status != "" {
		t.Errorf("Status = %q, want empty string", got.Status)
	}
	if got.DtLabel != "" {
		t.Errorf("DtLabel = %q, want empty string", got.DtLabel)
	}
	if got.DtType != "" {
		t.Errorf("DtType = %q, want empty string", got.DtType)
	}
	if got.DtParentID.Valid {
		t.Errorf("DtParentID.Valid = true, want false for zero value")
	}
}

func TestMysqlDatabase_MapAdminContentDataWithDatatypeRow_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	adminContentID, parentID, firstChild, nextSibling, prevSibling,
		adminRouteID, adminDatatypeID, authorID, status, ts,
		dtAdminDatatypeID, dtParentID, dtAuthorID := adminContentDataWithDatatypeFixture()

	input := mdbm.ListAdminContentDataWithDatatypeByRouteRow{
		AdminContentDataID: adminContentID,
		ParentID:           parentID,
		FirstChildID:       firstChild,
		NextSiblingID:      nextSibling,
		PrevSiblingID:      prevSibling,
		AdminRouteID:       adminRouteID,
		AdminDatatypeID:    adminDatatypeID,
		AuthorID:           authorID,
		Status:             status,
		DateCreated:        ts,
		DateModified:       ts,
		DtAdminDatatypeId:  dtAdminDatatypeID,
		DtParentId:         dtParentID,
		DtLabel:            "admin-page",
		DtType:             "admin-type",
		DtAuthorId:         dtAuthorID,
		DtDateCreated:      ts,
		DtDateModified:     ts,
	}

	got := d.MapAdminContentDataWithDatatypeRow(input)

	if got.AdminContentDataID != adminContentID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, adminContentID)
	}
	if got.DtAdminDatatypeID != dtAdminDatatypeID {
		t.Errorf("DtAdminDatatypeID = %v, want %v", got.DtAdminDatatypeID, dtAdminDatatypeID)
	}
	if got.DtLabel != "admin-page" {
		t.Errorf("DtLabel = %q, want %q", got.DtLabel, "admin-page")
	}
	if got.DtAuthorID != dtAuthorID {
		t.Errorf("DtAuthorID = %v, want %v", got.DtAuthorID, dtAuthorID)
	}
}

func TestMysqlDatabase_MapAdminContentDataWithDatatypeRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminContentDataWithDatatypeRow(mdbm.ListAdminContentDataWithDatatypeByRouteRow{})

	if got.AdminContentDataID != "" {
		t.Errorf("AdminContentDataID = %v, want zero value", got.AdminContentDataID)
	}
	if got.ParentID.Valid {
		t.Errorf("ParentID.Valid = true, want false for zero value")
	}
	if got.AdminRouteID.Valid {
		t.Errorf("AdminRouteID.Valid = true, want false for zero value")
	}
}

func TestPsqlDatabase_MapAdminContentDataWithDatatypeRow_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	adminContentID, parentID, firstChild, nextSibling, prevSibling,
		adminRouteID, adminDatatypeID, authorID, status, ts,
		dtAdminDatatypeID, dtParentID, dtAuthorID := adminContentDataWithDatatypeFixture()

	input := mdbp.ListAdminContentDataWithDatatypeByRouteRow{
		AdminContentDataID: adminContentID,
		ParentID:           parentID,
		FirstChildID:       firstChild,
		NextSiblingID:      nextSibling,
		PrevSiblingID:      prevSibling,
		AdminRouteID:       adminRouteID,
		AdminDatatypeID:    adminDatatypeID,
		AuthorID:           authorID,
		Status:             status,
		DateCreated:        ts,
		DateModified:       ts,
		DtAdminDatatypeId:  dtAdminDatatypeID,
		DtParentId:         dtParentID,
		DtLabel:            "admin-page",
		DtType:             "admin-type",
		DtAuthorId:         dtAuthorID,
		DtDateCreated:      ts,
		DtDateModified:     ts,
	}

	got := d.MapAdminContentDataWithDatatypeRow(input)

	if got.AdminContentDataID != adminContentID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, adminContentID)
	}
	if got.DtAdminDatatypeID != dtAdminDatatypeID {
		t.Errorf("DtAdminDatatypeID = %v, want %v", got.DtAdminDatatypeID, dtAdminDatatypeID)
	}
	if got.DtLabel != "admin-page" {
		t.Errorf("DtLabel = %q, want %q", got.DtLabel, "admin-page")
	}
	if got.DtAuthorID != dtAuthorID {
		t.Errorf("DtAuthorID = %v, want %v", got.DtAuthorID, dtAuthorID)
	}
}

func TestPsqlDatabase_MapAdminContentDataWithDatatypeRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminContentDataWithDatatypeRow(mdbp.ListAdminContentDataWithDatatypeByRouteRow{})

	if got.AdminContentDataID != "" {
		t.Errorf("AdminContentDataID = %v, want zero value", got.AdminContentDataID)
	}
	if got.ParentID.Valid {
		t.Errorf("ParentID.Valid = true, want false for zero value")
	}
	if got.AdminRouteID.Valid {
		t.Errorf("AdminRouteID.Valid = true, want false for zero value")
	}
}

// --- Cross-database consistency for AdminContentDataWithDatatypeRow ---

func TestCrossDatabase_MapAdminContentDataWithDatatypeRow_Consistency(t *testing.T) {
	t.Parallel()
	adminContentID, parentID, firstChild, nextSibling, prevSibling,
		adminRouteID, adminDatatypeID, authorID, status, ts,
		dtAdminDatatypeID, dtParentID, dtAuthorID := adminContentDataWithDatatypeFixture()

	sqliteInput := mdb.ListAdminContentDataWithDatatypeByRouteRow{
		AdminContentDataID: adminContentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		AdminRouteID: adminRouteID, AdminDatatypeID: adminDatatypeID,
		AuthorID: authorID, Status: status, DateCreated: ts, DateModified: ts,
		DtAdminDatatypeId: dtAdminDatatypeID, DtParentId: dtParentID,
		DtLabel: "cross", DtType: "cross-type",
		DtAuthorId: dtAuthorID, DtDateCreated: ts, DtDateModified: ts,
	}
	mysqlInput := mdbm.ListAdminContentDataWithDatatypeByRouteRow{
		AdminContentDataID: adminContentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		AdminRouteID: adminRouteID, AdminDatatypeID: adminDatatypeID,
		AuthorID: authorID, Status: status, DateCreated: ts, DateModified: ts,
		DtAdminDatatypeId: dtAdminDatatypeID, DtParentId: dtParentID,
		DtLabel: "cross", DtType: "cross-type",
		DtAuthorId: dtAuthorID, DtDateCreated: ts, DtDateModified: ts,
	}
	psqlInput := mdbp.ListAdminContentDataWithDatatypeByRouteRow{
		AdminContentDataID: adminContentID, ParentID: parentID,
		FirstChildID: firstChild, NextSiblingID: nextSibling, PrevSiblingID: prevSibling,
		AdminRouteID: adminRouteID, AdminDatatypeID: adminDatatypeID,
		AuthorID: authorID, Status: status, DateCreated: ts, DateModified: ts,
		DtAdminDatatypeId: dtAdminDatatypeID, DtParentId: dtParentID,
		DtLabel: "cross", DtType: "cross-type",
		DtAuthorId: dtAuthorID, DtDateCreated: ts, DtDateModified: ts,
	}

	sqliteResult := Database{}.MapAdminContentDataWithDatatypeRow(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminContentDataWithDatatypeRow(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminContentDataWithDatatypeRow(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// =============================================================================
// MapAdminContentFieldsWithFieldRow -- all 3 backends
// =============================================================================

func TestDatabase_MapAdminContentFieldsWithFieldRow_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 8, 20, 15, 0, 0, 0, time.UTC))
	adminContentFieldID := types.NewAdminContentFieldID()
	adminRouteID := types.NullableAdminRouteID{ID: types.NewAdminRouteID(), Valid: true}
	adminContentDataID := types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true}
	adminFieldID := types.NullableAdminFieldID{ID: types.NewAdminFieldID(), Valid: true}
	authorID := types.NewUserID()
	fAdminFieldID := types.NewAdminFieldID()
	fParentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	fAuthorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdb.ListAdminContentFieldsWithFieldByRouteRow{
		AdminContentFieldID: adminContentFieldID,
		AdminRouteID:        adminRouteID,
		AdminContentDataID:  adminContentDataID,
		AdminFieldID:        adminFieldID,
		AdminFieldValue:     "admin-value-text",
		AuthorID:            authorID,
		DateCreated:         ts,
		DateModified:        ts,
		FAdminFieldId:       fAdminFieldID,
		FParentId:           fParentID,
		FLabel:              "admin-field-label",
		FData:               "field-data",
		FValidation:         `{"required":true}`,
		FUiConfig:           `{"width":"full"}`,
		FType:               types.FieldType("textarea"),
		FAuthorId:           fAuthorID,
		FDateCreated:        ts,
		FDateModified:       ts,
	}

	got := d.MapAdminContentFieldsWithFieldRow(input)

	if got.AdminContentFieldID != adminContentFieldID {
		t.Errorf("AdminContentFieldID = %v, want %v", got.AdminContentFieldID, adminContentFieldID)
	}
	if got.AdminRouteID != adminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, adminRouteID)
	}
	if got.AdminContentDataID != adminContentDataID {
		t.Errorf("AdminContentDataID = %v, want %v", got.AdminContentDataID, adminContentDataID)
	}
	if got.AdminFieldID != adminFieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, adminFieldID)
	}
	if got.AdminFieldValue != "admin-value-text" {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, "admin-value-text")
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
	if got.FAdminFieldID != fAdminFieldID {
		t.Errorf("FAdminFieldID = %v, want %v", got.FAdminFieldID, fAdminFieldID)
	}
	if got.FParentID != fParentID {
		t.Errorf("FParentID = %v, want %v", got.FParentID, fParentID)
	}
	if got.FLabel != "admin-field-label" {
		t.Errorf("FLabel = %q, want %q", got.FLabel, "admin-field-label")
	}
	if got.FData != "field-data" {
		t.Errorf("FData = %q, want %q", got.FData, "field-data")
	}
	if got.FValidation != `{"required":true}` {
		t.Errorf("FValidation = %q, want %q", got.FValidation, `{"required":true}`)
	}
	if got.FUIConfig != `{"width":"full"}` {
		t.Errorf("FUIConfig = %q, want %q", got.FUIConfig, `{"width":"full"}`)
	}
	if got.FType != types.FieldType("textarea") {
		t.Errorf("FType = %v, want %v", got.FType, types.FieldType("textarea"))
	}
	if got.FAuthorID != fAuthorID {
		t.Errorf("FAuthorID = %v, want %v", got.FAuthorID, fAuthorID)
	}
	if got.FDateCreated != ts {
		t.Errorf("FDateCreated = %v, want %v", got.FDateCreated, ts)
	}
	if got.FDateModified != ts {
		t.Errorf("FDateModified = %v, want %v", got.FDateModified, ts)
	}
}

func TestDatabase_MapAdminContentFieldsWithFieldRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminContentFieldsWithFieldRow(mdb.ListAdminContentFieldsWithFieldByRouteRow{})

	if got.AdminContentFieldID != "" {
		t.Errorf("AdminContentFieldID = %v, want zero value", got.AdminContentFieldID)
	}
	if got.AdminRouteID.Valid {
		t.Errorf("AdminRouteID.Valid = true, want false for zero value")
	}
	if got.AdminContentDataID.Valid {
		t.Errorf("AdminContentDataID.Valid = true, want false for zero value")
	}
	if got.AdminFieldID.Valid {
		t.Errorf("AdminFieldID.Valid = true, want false for zero value")
	}
	if got.AdminFieldValue != "" {
		t.Errorf("AdminFieldValue = %q, want empty string", got.AdminFieldValue)
	}
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string for zero value", got.AuthorID)
	}
	if got.FAdminFieldID != "" {
		t.Errorf("FAdminFieldID = %v, want zero value", got.FAdminFieldID)
	}
	if got.FParentID.Valid {
		t.Errorf("FParentID.Valid = true, want false for zero value")
	}
	if got.FLabel != "" {
		t.Errorf("FLabel = %q, want empty string", got.FLabel)
	}
	if got.FData != "" {
		t.Errorf("FData = %q, want empty string", got.FData)
	}
	if got.FType != "" {
		t.Errorf("FType = %v, want zero value", got.FType)
	}
	if got.FAuthorID.Valid {
		t.Errorf("FAuthorID.Valid = true, want false for zero value")
	}
}

func TestMysqlDatabase_MapAdminContentFieldsWithFieldRow_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 8, 20, 15, 0, 0, 0, time.UTC))
	adminContentFieldID := types.NewAdminContentFieldID()
	fAdminFieldID := types.NewAdminFieldID()

	input := mdbm.ListAdminContentFieldsWithFieldByRouteRow{
		AdminContentFieldID: adminContentFieldID,
		AdminRouteID:        types.NullableAdminRouteID{ID: types.NewAdminRouteID(), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		AdminFieldID:        types.NullableAdminFieldID{ID: types.NewAdminFieldID(), Valid: true},
		AdminFieldValue:     "mysql-admin-value",
		AuthorID:            types.NewUserID(),
		DateCreated:         ts,
		DateModified:        ts,
		FAdminFieldId:       fAdminFieldID,
		FParentId:           types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true},
		FLabel:              "mysql-admin-field",
		FData:               "mysql-data",
		FValidation:         "{}",
		FUiConfig:           "{}",
		FType:               types.FieldType("text"),
		FAuthorId:           types.NullableUserID{ID: types.NewUserID(), Valid: true},
		FDateCreated:        ts,
		FDateModified:       ts,
	}

	got := d.MapAdminContentFieldsWithFieldRow(input)

	if got.AdminContentFieldID != adminContentFieldID {
		t.Errorf("AdminContentFieldID = %v, want %v", got.AdminContentFieldID, adminContentFieldID)
	}
	if got.FAdminFieldID != fAdminFieldID {
		t.Errorf("FAdminFieldID = %v, want %v", got.FAdminFieldID, fAdminFieldID)
	}
	if got.AdminFieldValue != "mysql-admin-value" {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, "mysql-admin-value")
	}
	if got.FLabel != "mysql-admin-field" {
		t.Errorf("FLabel = %q, want %q", got.FLabel, "mysql-admin-field")
	}
}

func TestMysqlDatabase_MapAdminContentFieldsWithFieldRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminContentFieldsWithFieldRow(mdbm.ListAdminContentFieldsWithFieldByRouteRow{})

	if got.AdminContentFieldID != "" {
		t.Errorf("AdminContentFieldID = %v, want zero value", got.AdminContentFieldID)
	}
	if got.AdminRouteID.Valid {
		t.Errorf("AdminRouteID.Valid = true, want false for zero value")
	}
	if got.AdminFieldID.Valid {
		t.Errorf("AdminFieldID.Valid = true, want false for zero value")
	}
	if got.FType != "" {
		t.Errorf("FType = %v, want zero value", got.FType)
	}
}

func TestPsqlDatabase_MapAdminContentFieldsWithFieldRow_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 8, 20, 15, 0, 0, 0, time.UTC))
	adminContentFieldID := types.NewAdminContentFieldID()
	fAdminFieldID := types.NewAdminFieldID()

	input := mdbp.ListAdminContentFieldsWithFieldByRouteRow{
		AdminContentFieldID: adminContentFieldID,
		AdminRouteID:        types.NullableAdminRouteID{ID: types.NewAdminRouteID(), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		AdminFieldID:        types.NullableAdminFieldID{ID: types.NewAdminFieldID(), Valid: true},
		AdminFieldValue:     "psql-admin-value",
		AuthorID:            types.NewUserID(),
		DateCreated:         ts,
		DateModified:        ts,
		FAdminFieldId:       fAdminFieldID,
		FParentId:           types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true},
		FLabel:              "psql-admin-field",
		FData:               "psql-data",
		FValidation:         "{}",
		FUiConfig:           "{}",
		FType:               types.FieldType("richtext"),
		FAuthorId:           types.NullableUserID{ID: types.NewUserID(), Valid: true},
		FDateCreated:        ts,
		FDateModified:       ts,
	}

	got := d.MapAdminContentFieldsWithFieldRow(input)

	if got.AdminContentFieldID != adminContentFieldID {
		t.Errorf("AdminContentFieldID = %v, want %v", got.AdminContentFieldID, adminContentFieldID)
	}
	if got.FAdminFieldID != fAdminFieldID {
		t.Errorf("FAdminFieldID = %v, want %v", got.FAdminFieldID, fAdminFieldID)
	}
	if got.AdminFieldValue != "psql-admin-value" {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, "psql-admin-value")
	}
	if got.FType != types.FieldType("richtext") {
		t.Errorf("FType = %v, want %v", got.FType, types.FieldType("richtext"))
	}
}

func TestPsqlDatabase_MapAdminContentFieldsWithFieldRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminContentFieldsWithFieldRow(mdbp.ListAdminContentFieldsWithFieldByRouteRow{})

	if got.AdminContentFieldID != "" {
		t.Errorf("AdminContentFieldID = %v, want zero value", got.AdminContentFieldID)
	}
	if got.AdminRouteID.Valid {
		t.Errorf("AdminRouteID.Valid = true, want false for zero value")
	}
	if got.AdminFieldID.Valid {
		t.Errorf("AdminFieldID.Valid = true, want false for zero value")
	}
	if got.FType != "" {
		t.Errorf("FType = %v, want zero value", got.FType)
	}
}

// --- Cross-database consistency for AdminContentFieldsWithFieldRow ---

func TestCrossDatabase_MapAdminContentFieldsWithFieldRow_Consistency(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	adminContentFieldID := types.NewAdminContentFieldID()
	adminRouteID := types.NullableAdminRouteID{ID: types.NewAdminRouteID(), Valid: true}
	adminContentDataID := types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true}
	adminFieldID := types.NullableAdminFieldID{ID: types.NewAdminFieldID(), Valid: true}
	authorID := types.NewUserID()
	fAdminFieldID := types.NewAdminFieldID()
	fParentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	fAuthorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	sqliteInput := mdb.ListAdminContentFieldsWithFieldByRouteRow{
		AdminContentFieldID: adminContentFieldID, AdminRouteID: adminRouteID,
		AdminContentDataID: adminContentDataID, AdminFieldID: adminFieldID,
		AdminFieldValue: "cross-val", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
		FAdminFieldId: fAdminFieldID, FParentId: fParentID,
		FLabel: "cross-label", FData: "data", FValidation: "{}", FUiConfig: "{}",
		FType: types.FieldType("text"), FAuthorId: fAuthorID,
		FDateCreated: ts, FDateModified: ts,
	}
	mysqlInput := mdbm.ListAdminContentFieldsWithFieldByRouteRow{
		AdminContentFieldID: adminContentFieldID, AdminRouteID: adminRouteID,
		AdminContentDataID: adminContentDataID, AdminFieldID: adminFieldID,
		AdminFieldValue: "cross-val", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
		FAdminFieldId: fAdminFieldID, FParentId: fParentID,
		FLabel: "cross-label", FData: "data", FValidation: "{}", FUiConfig: "{}",
		FType: types.FieldType("text"), FAuthorId: fAuthorID,
		FDateCreated: ts, FDateModified: ts,
	}
	psqlInput := mdbp.ListAdminContentFieldsWithFieldByRouteRow{
		AdminContentFieldID: adminContentFieldID, AdminRouteID: adminRouteID,
		AdminContentDataID: adminContentDataID, AdminFieldID: adminFieldID,
		AdminFieldValue: "cross-val", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
		FAdminFieldId: fAdminFieldID, FParentId: fParentID,
		FLabel: "cross-label", FData: "data", FValidation: "{}", FUiConfig: "{}",
		FType: types.FieldType("text"), FAuthorId: fAuthorID,
		FDateCreated: ts, FDateModified: ts,
	}

	sqliteResult := Database{}.MapAdminContentFieldsWithFieldRow(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminContentFieldsWithFieldRow(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminContentFieldsWithFieldRow(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// =============================================================================
// MapContentFieldWithFieldRow -- all 3 backends
// =============================================================================

func TestDatabase_MapContentFieldWithFieldRow_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 2, 14, 12, 0, 0, 0, time.UTC))
	contentFieldID := types.NewContentFieldID()
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NewUserID()
	fFieldID := types.NewFieldID()

	input := mdb.ListContentFieldsWithFieldByContentDataRow{
		ContentFieldID: contentFieldID,
		RouteID:        routeID,
		ContentDataID:  contentDataID,
		FieldID:        fieldID,
		FieldValue:     "content-field-value",
		AuthorID:       authorID,
		DateCreated:    ts,
		DateModified:   ts,
		FFieldId:       fFieldID,
		FLabel:         "field-label",
		FType:          types.FieldType("text"),
	}

	got := d.MapContentFieldWithFieldRow(input)

	if got.ContentFieldID != contentFieldID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, contentFieldID)
	}
	if got.RouteID != routeID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, routeID)
	}
	if got.ContentDataID != contentDataID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, contentDataID)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.FieldValue != "content-field-value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "content-field-value")
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
	if got.FFieldID != fFieldID {
		t.Errorf("FFieldID = %v, want %v", got.FFieldID, fFieldID)
	}
	if got.FLabel != "field-label" {
		t.Errorf("FLabel = %q, want %q", got.FLabel, "field-label")
	}
	if got.FType != types.FieldType("text") {
		t.Errorf("FType = %v, want %v", got.FType, types.FieldType("text"))
	}
}

func TestDatabase_MapContentFieldWithFieldRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapContentFieldWithFieldRow(mdb.ListContentFieldsWithFieldByContentDataRow{})

	if got.ContentFieldID != "" {
		t.Errorf("ContentFieldID = %v, want zero value", got.ContentFieldID)
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false for zero value")
	}
	if got.ContentDataID.Valid {
		t.Errorf("ContentDataID.Valid = true, want false for zero value")
	}
	if got.FieldID.Valid {
		t.Errorf("FieldID.Valid = true, want false for zero value")
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string for zero value", got.AuthorID)
	}
	if got.FFieldID != "" {
		t.Errorf("FFieldID = %v, want zero value", got.FFieldID)
	}
	if got.FLabel != "" {
		t.Errorf("FLabel = %q, want empty string", got.FLabel)
	}
	if got.FType != "" {
		t.Errorf("FType = %v, want zero value", got.FType)
	}
}

func TestMysqlDatabase_MapContentFieldWithFieldRow_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 2, 14, 12, 0, 0, 0, time.UTC))
	contentFieldID := types.NewContentFieldID()
	fFieldID := types.NewFieldID()

	input := mdbm.ListContentFieldsWithFieldByContentDataRow{
		ContentFieldID: contentFieldID,
		RouteID:        types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		ContentDataID:  types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FieldID:        types.NullableFieldID{ID: types.NewFieldID(), Valid: true},
		FieldValue:     "mysql-content-field",
		AuthorID:       types.NewUserID(),
		DateCreated:    ts,
		DateModified:   ts,
		FFieldId:       fFieldID,
		FLabel:         "mysql-field-label",
		FType:          types.FieldType("number"),
	}

	got := d.MapContentFieldWithFieldRow(input)

	if got.ContentFieldID != contentFieldID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, contentFieldID)
	}
	if got.FFieldID != fFieldID {
		t.Errorf("FFieldID = %v, want %v", got.FFieldID, fFieldID)
	}
	if got.FieldValue != "mysql-content-field" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "mysql-content-field")
	}
	if got.FLabel != "mysql-field-label" {
		t.Errorf("FLabel = %q, want %q", got.FLabel, "mysql-field-label")
	}
	if got.FType != types.FieldType("number") {
		t.Errorf("FType = %v, want %v", got.FType, types.FieldType("number"))
	}
}

func TestMysqlDatabase_MapContentFieldWithFieldRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapContentFieldWithFieldRow(mdbm.ListContentFieldsWithFieldByContentDataRow{})

	if got.ContentFieldID != "" {
		t.Errorf("ContentFieldID = %v, want zero value", got.ContentFieldID)
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false for zero value")
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
	if got.FType != "" {
		t.Errorf("FType = %v, want zero value", got.FType)
	}
}

func TestPsqlDatabase_MapContentFieldWithFieldRow_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 2, 14, 12, 0, 0, 0, time.UTC))
	contentFieldID := types.NewContentFieldID()
	fFieldID := types.NewFieldID()

	input := mdbp.ListContentFieldsWithFieldByContentDataRow{
		ContentFieldID: contentFieldID,
		RouteID:        types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		ContentDataID:  types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FieldID:        types.NullableFieldID{ID: types.NewFieldID(), Valid: true},
		FieldValue:     "psql-content-field",
		AuthorID:       types.NewUserID(),
		DateCreated:    ts,
		DateModified:   ts,
		FFieldId:       fFieldID,
		FLabel:         "psql-field-label",
		FType:          types.FieldType("date"),
	}

	got := d.MapContentFieldWithFieldRow(input)

	if got.ContentFieldID != contentFieldID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, contentFieldID)
	}
	if got.FFieldID != fFieldID {
		t.Errorf("FFieldID = %v, want %v", got.FFieldID, fFieldID)
	}
	if got.FieldValue != "psql-content-field" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "psql-content-field")
	}
	if got.FType != types.FieldType("date") {
		t.Errorf("FType = %v, want %v", got.FType, types.FieldType("date"))
	}
}

func TestPsqlDatabase_MapContentFieldWithFieldRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapContentFieldWithFieldRow(mdbp.ListContentFieldsWithFieldByContentDataRow{})

	if got.ContentFieldID != "" {
		t.Errorf("ContentFieldID = %v, want zero value", got.ContentFieldID)
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false for zero value")
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
	if got.FType != "" {
		t.Errorf("FType = %v, want zero value", got.FType)
	}
}

// --- Cross-database consistency for ContentFieldWithFieldRow ---

func TestCrossDatabase_MapContentFieldWithFieldRow_Consistency(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))
	contentFieldID := types.NewContentFieldID()
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NewUserID()
	fFieldID := types.NewFieldID()

	sqliteInput := mdb.ListContentFieldsWithFieldByContentDataRow{
		ContentFieldID: contentFieldID, RouteID: routeID, ContentDataID: contentDataID,
		FieldID: fieldID, FieldValue: "cross", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
		FFieldId: fFieldID, FLabel: "cross-label", FType: types.FieldType("text"),
	}
	mysqlInput := mdbm.ListContentFieldsWithFieldByContentDataRow{
		ContentFieldID: contentFieldID, RouteID: routeID, ContentDataID: contentDataID,
		FieldID: fieldID, FieldValue: "cross", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
		FFieldId: fFieldID, FLabel: "cross-label", FType: types.FieldType("text"),
	}
	psqlInput := mdbp.ListContentFieldsWithFieldByContentDataRow{
		ContentFieldID: contentFieldID, RouteID: routeID, ContentDataID: contentDataID,
		FieldID: fieldID, FieldValue: "cross", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
		FFieldId: fFieldID, FLabel: "cross-label", FType: types.FieldType("text"),
	}

	sqliteResult := Database{}.MapContentFieldWithFieldRow(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapContentFieldWithFieldRow(mysqlInput)
	psqlResult := PsqlDatabase{}.MapContentFieldWithFieldRow(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// =============================================================================
// MapUserWithRoleLabelRow -- all 3 backends
// =============================================================================

func TestDatabase_MapUserWithRoleLabelRow_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	userID := types.NewUserID()
	email := types.Email("admin@example.com")

	input := mdb.ListUsersWithRoleLabelRow{
		UserID:       userID,
		Username:     "admin",
		Name:         "Admin User",
		Email:        email,
		Roles:        "admin-role-id",
		RoleLabel:    "Administrator",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapUserWithRoleLabelRow(input)

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Username != "admin" {
		t.Errorf("Username = %q, want %q", got.Username, "admin")
	}
	if got.Name != "Admin User" {
		t.Errorf("Name = %q, want %q", got.Name, "Admin User")
	}
	if got.Email != email {
		t.Errorf("Email = %v, want %v", got.Email, email)
	}
	// Note: source field is "Roles" but wrapper field is "Role"
	if got.Role != "admin-role-id" {
		t.Errorf("Role = %q, want %q", got.Role, "admin-role-id")
	}
	if got.RoleLabel != "Administrator" {
		t.Errorf("RoleLabel = %q, want %q", got.RoleLabel, "Administrator")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestDatabase_MapUserWithRoleLabelRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUserWithRoleLabelRow(mdb.ListUsersWithRoleLabelRow{})

	if got.UserID != "" {
		t.Errorf("UserID = %v, want zero value", got.UserID)
	}
	if got.Username != "" {
		t.Errorf("Username = %q, want empty string", got.Username)
	}
	if got.Name != "" {
		t.Errorf("Name = %q, want empty string", got.Name)
	}
	if got.Email != "" {
		t.Errorf("Email = %v, want zero value", got.Email)
	}
	if got.Role != "" {
		t.Errorf("Role = %q, want empty string", got.Role)
	}
	if got.RoleLabel != "" {
		t.Errorf("RoleLabel = %q, want empty string", got.RoleLabel)
	}
}

func TestMysqlDatabase_MapUserWithRoleLabelRow_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	userID := types.NewUserID()
	email := types.Email("editor@example.com")

	input := mdbm.ListUsersWithRoleLabelRow{
		UserID:       userID,
		Username:     "editor",
		Name:         "Editor User",
		Email:        email,
		Roles:        "editor-role-id",
		RoleLabel:    "Editor",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapUserWithRoleLabelRow(input)

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Username != "editor" {
		t.Errorf("Username = %q, want %q", got.Username, "editor")
	}
	if got.Role != "editor-role-id" {
		t.Errorf("Role = %q, want %q", got.Role, "editor-role-id")
	}
	if got.RoleLabel != "Editor" {
		t.Errorf("RoleLabel = %q, want %q", got.RoleLabel, "Editor")
	}
}

func TestMysqlDatabase_MapUserWithRoleLabelRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapUserWithRoleLabelRow(mdbm.ListUsersWithRoleLabelRow{})

	if got.UserID != "" {
		t.Errorf("UserID = %v, want zero value", got.UserID)
	}
	if got.Role != "" {
		t.Errorf("Role = %q, want empty string", got.Role)
	}
	if got.RoleLabel != "" {
		t.Errorf("RoleLabel = %q, want empty string", got.RoleLabel)
	}
}

func TestPsqlDatabase_MapUserWithRoleLabelRow_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	userID := types.NewUserID()
	email := types.Email("viewer@example.com")

	input := mdbp.ListUsersWithRoleLabelRow{
		UserID:       userID,
		Username:     "viewer",
		Name:         "Viewer User",
		Email:        email,
		Roles:        "viewer-role-id",
		RoleLabel:    "Viewer",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapUserWithRoleLabelRow(input)

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Username != "viewer" {
		t.Errorf("Username = %q, want %q", got.Username, "viewer")
	}
	if got.Role != "viewer-role-id" {
		t.Errorf("Role = %q, want %q", got.Role, "viewer-role-id")
	}
	if got.RoleLabel != "Viewer" {
		t.Errorf("RoleLabel = %q, want %q", got.RoleLabel, "Viewer")
	}
}

func TestPsqlDatabase_MapUserWithRoleLabelRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapUserWithRoleLabelRow(mdbp.ListUsersWithRoleLabelRow{})

	if got.UserID != "" {
		t.Errorf("UserID = %v, want zero value", got.UserID)
	}
	if got.Role != "" {
		t.Errorf("Role = %q, want empty string", got.Role)
	}
	if got.RoleLabel != "" {
		t.Errorf("RoleLabel = %q, want empty string", got.RoleLabel)
	}
}

// --- Cross-database consistency for UserWithRoleLabelRow ---

func TestCrossDatabase_MapUserWithRoleLabelRow_Consistency(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC))
	userID := types.NewUserID()
	email := types.Email("cross@example.com")

	sqliteInput := mdb.ListUsersWithRoleLabelRow{
		UserID: userID, Username: "cross", Name: "Cross User", Email: email,
		Roles: "role-cross", RoleLabel: "Cross", DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.ListUsersWithRoleLabelRow{
		UserID: userID, Username: "cross", Name: "Cross User", Email: email,
		Roles: "role-cross", RoleLabel: "Cross", DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.ListUsersWithRoleLabelRow{
		UserID: userID, Username: "cross", Name: "Cross User", Email: email,
		Roles: "role-cross", RoleLabel: "Cross", DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapUserWithRoleLabelRow(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapUserWithRoleLabelRow(mysqlInput)
	psqlResult := PsqlDatabase{}.MapUserWithRoleLabelRow(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// =============================================================================
// MapFieldWithSortOrderRow -- all 3 backends
// Note: MySQL and PostgreSQL use int32 for SortOrder, SQLite uses int64.
// The mapper casts int32 -> int64 for MySQL/PostgreSQL. This is a critical
// behavior to verify because a dropped cast would silently truncate.
// =============================================================================

func TestDatabase_MapFieldWithSortOrderRow_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	fieldID := types.NewFieldID()

	input := mdb.ListFieldsWithSortOrderByDatatypeIDRow{
		SortOrder:  42,
		FieldID:    fieldID,
		Label:      "sort-field",
		Type:       types.FieldType("select"),
		Data:       `{"options":["a","b"]}`,
		Validation: `{"required":true}`,
		UiConfig:   `{"display":"dropdown"}`,
	}

	got := d.MapFieldWithSortOrderRow(input)

	if got.SortOrder != 42 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 42)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.Label != "sort-field" {
		t.Errorf("Label = %q, want %q", got.Label, "sort-field")
	}
	if got.Type != types.FieldType("select") {
		t.Errorf("Type = %v, want %v", got.Type, types.FieldType("select"))
	}
	if got.Data != `{"options":["a","b"]}` {
		t.Errorf("Data = %q, want %q", got.Data, `{"options":["a","b"]}`)
	}
	if got.Validation != `{"required":true}` {
		t.Errorf("Validation = %q, want %q", got.Validation, `{"required":true}`)
	}
	if got.UIConfig != `{"display":"dropdown"}` {
		t.Errorf("UIConfig = %q, want %q", got.UIConfig, `{"display":"dropdown"}`)
	}
}

func TestDatabase_MapFieldWithSortOrderRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapFieldWithSortOrderRow(mdb.ListFieldsWithSortOrderByDatatypeIDRow{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %v, want zero value", got.Type)
	}
	if got.Data != "" {
		t.Errorf("Data = %q, want empty string", got.Data)
	}
	if got.Validation != "" {
		t.Errorf("Validation = %q, want empty string", got.Validation)
	}
	if got.UIConfig != "" {
		t.Errorf("UIConfig = %q, want empty string", got.UIConfig)
	}
}

// TestDatabase_MapFieldWithSortOrderRow_LargeSortOrder verifies that the
// SQLite int64 SortOrder preserves values beyond int32 range.
func TestDatabase_MapFieldWithSortOrderRow_LargeSortOrder(t *testing.T) {
	t.Parallel()
	d := Database{}

	// Value beyond int32 max (2,147,483,647)
	input := mdb.ListFieldsWithSortOrderByDatatypeIDRow{
		SortOrder: 3_000_000_000,
		FieldID:   types.NewFieldID(),
		Label:     "large-sort",
		Type:      types.FieldType("text"),
	}

	got := d.MapFieldWithSortOrderRow(input)

	if got.SortOrder != 3_000_000_000 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, int64(3_000_000_000))
	}
}

// TestMysqlDatabase_MapFieldWithSortOrderRow_Int32ToInt64Cast verifies the
// int32 -> int64 cast that happens for MySQL. This is the primary behavior
// difference between the SQLite mapper (identity) and MySQL/PostgreSQL mappers.
func TestMysqlDatabase_MapFieldWithSortOrderRow_Int32ToInt64Cast(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	tests := []struct {
		name      string
		sortOrder int32
		want      int64
	}{
		{name: "zero", sortOrder: 0, want: 0},
		{name: "positive", sortOrder: 100, want: 100},
		{name: "negative", sortOrder: -1, want: -1},
		{name: "max int32", sortOrder: 2_147_483_647, want: 2_147_483_647},
		{name: "min int32", sortOrder: -2_147_483_648, want: -2_147_483_648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbm.ListFieldsWithSortOrderByDatatypeIDRow{
				SortOrder: tt.sortOrder,
				FieldID:   types.NewFieldID(),
				Label:     "mysql-sort",
				Type:      types.FieldType("text"),
			}

			got := d.MapFieldWithSortOrderRow(input)

			if got.SortOrder != tt.want {
				t.Errorf("SortOrder = %d, want %d", got.SortOrder, tt.want)
			}
		})
	}
}

func TestMysqlDatabase_MapFieldWithSortOrderRow_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	fieldID := types.NewFieldID()

	input := mdbm.ListFieldsWithSortOrderByDatatypeIDRow{
		SortOrder:  7,
		FieldID:    fieldID,
		Label:      "mysql-sort-field",
		Type:       types.FieldType("media"),
		Data:       "mysql-data",
		Validation: "mysql-validation",
		UiConfig:   "mysql-ui",
	}

	got := d.MapFieldWithSortOrderRow(input)

	if got.SortOrder != 7 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 7)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.Label != "mysql-sort-field" {
		t.Errorf("Label = %q, want %q", got.Label, "mysql-sort-field")
	}
	if got.UIConfig != "mysql-ui" {
		t.Errorf("UIConfig = %q, want %q", got.UIConfig, "mysql-ui")
	}
}

func TestMysqlDatabase_MapFieldWithSortOrderRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapFieldWithSortOrderRow(mdbm.ListFieldsWithSortOrderByDatatypeIDRow{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// TestPsqlDatabase_MapFieldWithSortOrderRow_Int32ToInt64Cast mirrors the MySQL
// test -- PostgreSQL also uses int32 for SortOrder and the mapper must cast.
func TestPsqlDatabase_MapFieldWithSortOrderRow_Int32ToInt64Cast(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	tests := []struct {
		name      string
		sortOrder int32
		want      int64
	}{
		{name: "zero", sortOrder: 0, want: 0},
		{name: "positive", sortOrder: 255, want: 255},
		{name: "negative", sortOrder: -1, want: -1},
		{name: "max int32", sortOrder: 2_147_483_647, want: 2_147_483_647},
		{name: "min int32", sortOrder: -2_147_483_648, want: -2_147_483_648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbp.ListFieldsWithSortOrderByDatatypeIDRow{
				SortOrder: tt.sortOrder,
				FieldID:   types.NewFieldID(),
				Label:     "psql-sort",
				Type:      types.FieldType("text"),
			}

			got := d.MapFieldWithSortOrderRow(input)

			if got.SortOrder != tt.want {
				t.Errorf("SortOrder = %d, want %d", got.SortOrder, tt.want)
			}
		})
	}
}

func TestPsqlDatabase_MapFieldWithSortOrderRow_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	fieldID := types.NewFieldID()

	input := mdbp.ListFieldsWithSortOrderByDatatypeIDRow{
		SortOrder:  3,
		FieldID:    fieldID,
		Label:      "psql-sort-field",
		Type:       types.FieldType("boolean"),
		Data:       "psql-data",
		Validation: "psql-validation",
		UiConfig:   "psql-ui",
	}

	got := d.MapFieldWithSortOrderRow(input)

	if got.SortOrder != 3 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, 3)
	}
	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.Label != "psql-sort-field" {
		t.Errorf("Label = %q, want %q", got.Label, "psql-sort-field")
	}
	if got.UIConfig != "psql-ui" {
		t.Errorf("UIConfig = %q, want %q", got.UIConfig, "psql-ui")
	}
}

func TestPsqlDatabase_MapFieldWithSortOrderRow_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapFieldWithSortOrderRow(mdbp.ListFieldsWithSortOrderByDatatypeIDRow{})

	if got.SortOrder != 0 {
		t.Errorf("SortOrder = %d, want 0", got.SortOrder)
	}
	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- Cross-database consistency for FieldWithSortOrderRow ---
// Note: SQLite uses int64 natively for SortOrder, MySQL/PostgreSQL use int32.
// The wrapper normalizes to int64. When the value fits in int32, all three
// backends should produce identical results.

func TestCrossDatabase_MapFieldWithSortOrderRow_Consistency(t *testing.T) {
	t.Parallel()
	fieldID := types.NewFieldID()

	sqliteInput := mdb.ListFieldsWithSortOrderByDatatypeIDRow{
		SortOrder: 99, FieldID: fieldID, Label: "cross-sort",
		Type: types.FieldType("text"), Data: "d", Validation: "v", UiConfig: "u",
	}
	mysqlInput := mdbm.ListFieldsWithSortOrderByDatatypeIDRow{
		SortOrder: 99, FieldID: fieldID, Label: "cross-sort",
		Type: types.FieldType("text"), Data: "d", Validation: "v", UiConfig: "u",
	}
	psqlInput := mdbp.ListFieldsWithSortOrderByDatatypeIDRow{
		SortOrder: 99, FieldID: fieldID, Label: "cross-sort",
		Type: types.FieldType("text"), Data: "d", Validation: "v", UiConfig: "u",
	}

	sqliteResult := Database{}.MapFieldWithSortOrderRow(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapFieldWithSortOrderRow(mysqlInput)
	psqlResult := PsqlDatabase{}.MapFieldWithSortOrderRow(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// =============================================================================
// JSON tag verification for remaining wrapper structs
// =============================================================================

func TestAdminContentDataWithDatatypeRow_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	row := AdminContentDataWithDatatypeRow{
		AdminContentDataID: types.NewAdminContentID(),
		ParentID:           types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		FirstChildID:       types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		NextSiblingID:      types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		PrevSiblingID:      types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		AdminRouteID:       types.NullableAdminRouteID{ID: types.NewAdminRouteID(), Valid: true},
		AdminDatatypeID:    types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true},
		AuthorID:           types.NewUserID(),
		Status:             types.ContentStatus("published"),
		DateCreated:        ts,
		DateModified:       ts,
		DtAdminDatatypeID:  types.NewAdminDatatypeID(),
		DtParentID:         types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true},
		DtLabel:            "test",
		DtType:             "page",
		DtAuthorID:         types.NewUserID(),
		DtDateCreated:      ts,
		DtDateModified:     ts,
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
		"admin_content_data_id", "parent_id", "first_child_id",
		"next_sibling_id", "prev_sibling_id",
		"admin_route_id", "admin_datatype_id",
		"author_id", "status", "date_created", "date_modified",
		"dt_admin_datatype_id", "dt_parent_id", "dt_label", "dt_type",
		"dt_author_id", "dt_date_created", "dt_date_modified",
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

func TestAdminContentFieldsWithFieldRow_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	row := AdminContentFieldsWithFieldRow{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        types.NullableAdminRouteID{ID: types.NewAdminRouteID(), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.NewAdminContentID(), Valid: true},
		AdminFieldID:        types.NullableAdminFieldID{ID: types.NewAdminFieldID(), Valid: true},
		AdminFieldValue:     "value",
		AuthorID:            types.NewUserID(),
		DateCreated:         ts,
		DateModified:        ts,
		FAdminFieldID:       types.NewAdminFieldID(),
		FParentID:           types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true},
		FLabel:              "label",
		FData:               "data",
		FValidation:         "{}",
		FUIConfig:           "{}",
		FType:               types.FieldType("text"),
		FAuthorID:           types.NullableUserID{ID: types.NewUserID(), Valid: true},
		FDateCreated:        ts,
		FDateModified:       ts,
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
		"admin_content_field_id", "admin_route_id", "admin_content_data_id",
		"admin_field_id", "admin_field_value", "author_id",
		"date_created", "date_modified",
		"f_admin_field_id", "f_parent_id", "f_label", "f_data",
		"f_validation", "f_ui_config", "f_type",
		"f_author_id", "f_date_created", "f_date_modified",
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

func TestContentFieldWithFieldRow_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	row := ContentFieldWithFieldRow{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		ContentDataID:  types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FieldID:        types.NullableFieldID{ID: types.NewFieldID(), Valid: true},
		FieldValue:     "value",
		AuthorID:       types.NewUserID(),
		DateCreated:    ts,
		DateModified:   ts,
		FFieldID:       types.NewFieldID(),
		FLabel:         "label",
		FType:          types.FieldType("text"),
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
		"content_field_id", "route_id", "content_data_id",
		"field_id", "field_value", "author_id",
		"date_created", "date_modified",
		"f_field_id", "f_label", "f_type",
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

func TestUserWithRoleLabelRow_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	row := UserWithRoleLabelRow{
		UserID:       types.NewUserID(),
		Username:     "user",
		Name:         "Test User",
		Email:        types.Email("test@example.com"),
		Role:         "role-id",
		RoleLabel:    "Admin",
		DateCreated:  ts,
		DateModified: ts,
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
		"user_id", "username", "name", "email",
		"role", "role_label", "date_created", "date_modified",
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

func TestFieldWithSortOrderRow_JSONTags(t *testing.T) {
	t.Parallel()
	row := FieldWithSortOrderRow{
		SortOrder:  1,
		FieldID:    types.NewFieldID(),
		Label:      "test",
		Type:       types.FieldType("text"),
		Data:       "data",
		Validation: "{}",
		UIConfig:   "{}",
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
		"sort_order", "field_id", "label", "type",
		"data", "validation", "ui_config",
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

// =============================================================================
// Edge case: AdminContentFieldsWithFieldRow with all nullable fields invalid
// This tests the "orphan admin content field" scenario -- an admin content field
// that exists but has no route, no content data, no field, and no author.
// =============================================================================

func TestDatabase_MapAdminContentFieldsWithFieldRow_AllNullableFieldsInvalid(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	input := mdb.ListAdminContentFieldsWithFieldByRouteRow{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        types.NullableAdminRouteID{Valid: false},
		AdminContentDataID:  types.NullableAdminContentID{Valid: false},
		AdminFieldID:        types.NullableAdminFieldID{Valid: false},
		AdminFieldValue:     "orphan-value",
		AuthorID:            types.UserID(""),
		DateCreated:         ts,
		DateModified:        ts,
		FAdminFieldId:       types.NewAdminFieldID(),
		FParentId:           types.NullableAdminDatatypeID{Valid: false},
		FLabel:              "orphan-field",
		FData:               "",
		FValidation:         "{}",
		FUiConfig:           "{}",
		FType:               types.FieldType("text"),
		FAuthorId:           types.NullableUserID{Valid: false},
		FDateCreated:        ts,
		FDateModified:       ts,
	}

	got := d.MapAdminContentFieldsWithFieldRow(input)

	if got.AdminRouteID.Valid {
		t.Errorf("AdminRouteID.Valid = true, want false")
	}
	if got.AdminContentDataID.Valid {
		t.Errorf("AdminContentDataID.Valid = true, want false")
	}
	if got.AdminFieldID.Valid {
		t.Errorf("AdminFieldID.Valid = true, want false")
	}
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string", got.AuthorID)
	}
	if got.FParentID.Valid {
		t.Errorf("FParentID.Valid = true, want false")
	}
	if got.FAuthorID.Valid {
		t.Errorf("FAuthorID.Valid = true, want false")
	}
	// Non-nullable fields should still be mapped
	if got.AdminFieldValue != "orphan-value" {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, "orphan-value")
	}
	if got.FLabel != "orphan-field" {
		t.Errorf("FLabel = %q, want %q", got.FLabel, "orphan-field")
	}
}

// =============================================================================
// Edge case: ContentFieldWithFieldRow with empty FieldValue
// Empty string is a valid value distinct from "no field value set".
// =============================================================================

func TestDatabase_MapContentFieldWithFieldRow_EmptyFieldValue(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	input := mdb.ListContentFieldsWithFieldByContentDataRow{
		ContentFieldID: types.NewContentFieldID(),
		RouteID:        types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		ContentDataID:  types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FieldID:        types.NullableFieldID{ID: types.NewFieldID(), Valid: true},
		FieldValue:     "",
		AuthorID:       types.NewUserID(),
		DateCreated:    ts,
		DateModified:   ts,
		FFieldId:       types.NewFieldID(),
		FLabel:         "empty-value-field",
		FType:          types.FieldType("text"),
	}

	got := d.MapContentFieldWithFieldRow(input)

	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
	// IDs should still be mapped correctly despite empty value
	if !got.ContentDataID.Valid {
		t.Errorf("ContentDataID.Valid = false, want true")
	}
	if !got.FieldID.Valid {
		t.Errorf("FieldID.Valid = false, want true")
	}
}

// =============================================================================
// Edge case: UserWithRoleLabelRow with special characters
// Verifies that unicode, special chars in name/email pass through unmolested.
// =============================================================================

func TestDatabase_MapUserWithRoleLabelRow_SpecialCharacters(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		username string
		userName string
		email    types.Email
	}{
		{
			name:     "unicode name",
			username: "yuki",
			userName: "Tanaka Yuki",
			email:    types.Email("yuki@example.com"),
		},
		{
			name:     "hyphenated email",
			username: "john-doe",
			userName: "John Doe",
			email:    types.Email("john-doe@my-company.example.com"),
		},
		{
			name:     "plus addressing",
			username: "user",
			userName: "Test User",
			email:    types.Email("user+tag@example.com"),
		},
		{
			name:     "empty name",
			username: "anonymous",
			userName: "",
			email:    types.Email("anon@example.com"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdb.ListUsersWithRoleLabelRow{
				UserID:       types.NewUserID(),
				Username:     tt.username,
				Name:         tt.userName,
				Email:        tt.email,
				Roles:        "role",
				RoleLabel:    "Label",
				DateCreated:  ts,
				DateModified: ts,
			}

			got := d.MapUserWithRoleLabelRow(input)

			if got.Username != tt.username {
				t.Errorf("Username = %q, want %q", got.Username, tt.username)
			}
			if got.Name != tt.userName {
				t.Errorf("Name = %q, want %q", got.Name, tt.userName)
			}
			if got.Email != tt.email {
				t.Errorf("Email = %v, want %v", got.Email, tt.email)
			}
		})
	}
}

// =============================================================================
// Edge case: FieldWithSortOrderRow negative sort order
// Negative sort orders should be preserved through the mapper.
// =============================================================================

func TestDatabase_MapFieldWithSortOrderRow_NegativeSortOrder(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := mdb.ListFieldsWithSortOrderByDatatypeIDRow{
		SortOrder: -10,
		FieldID:   types.NewFieldID(),
		Label:     "negative-sort",
		Type:      types.FieldType("text"),
	}

	got := d.MapFieldWithSortOrderRow(input)

	if got.SortOrder != -10 {
		t.Errorf("SortOrder = %d, want %d", got.SortOrder, -10)
	}
}
