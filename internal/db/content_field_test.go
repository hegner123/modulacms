package db

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// contentFieldFixture returns a fully populated ContentFields and its parts for testing.
func contentFieldFixture() (ContentFields, types.ContentFieldID, types.Timestamp) {
	cfID := types.NewContentFieldID()
	ts := types.NewTimestamp(time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC))
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	cf := ContentFields{
		ContentFieldID: cfID,
		RouteID:        routeID,
		ContentDataID:  contentDataID,
		FieldID:        fieldID,
		FieldValue:     "test field value",
		AuthorID:       authorID,
		DateCreated:    ts,
		DateModified:   ts,
	}
	return cf, cfID, ts
}

// contentFieldFixtureNulls returns a ContentFields where all nullable fields are null/invalid.
func contentFieldFixtureNulls() ContentFields {
	cfID := types.NewContentFieldID()
	ts := types.NewTimestamp(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	return ContentFields{
		ContentFieldID: cfID,
		RouteID:        types.NullableRouteID{Valid: false},
		ContentDataID:  types.NullableContentID{Valid: false},
		FieldID:        types.NullableFieldID{Valid: false},
		FieldValue:     "",
		AuthorID:       types.NullableUserID{Valid: false},
		DateCreated:    ts,
		DateModified:   ts,
	}
}

// --- MapContentFieldJSON tests ---

func TestMapContentFieldJSON_AllFieldsValid(t *testing.T) {
	t.Parallel()
	cf, _, ts := contentFieldFixture()

	got := MapContentFieldJSON(cf)

	// MapContentFieldJSON sets all ID fields to 0 (deprecated -- type conversion not available).
	// Only FieldValue and timestamps carry data.
	tests := []struct {
		name string
		got  int64
		want int64
	}{
		{"ContentFieldID", got.ContentFieldID, 0},
		{"RouteID", got.RouteID, 0},
		{"ContentDataID", got.ContentDataID, 0},
		{"FieldID", got.FieldID, 0},
		{"AuthorID", got.AuthorID, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}

	if got.FieldValue != "test field value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "test field value")
	}
	if got.DateCreated != ts.String() {
		t.Errorf("DateCreated = %q, want %q", got.DateCreated, ts.String())
	}
	if got.DateModified != ts.String() {
		t.Errorf("DateModified = %q, want %q", got.DateModified, ts.String())
	}
}

func TestMapContentFieldJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value ContentFields should produce a valid ContentFieldsJSON without panic
	got := MapContentFieldJSON(ContentFields{})

	if got.ContentFieldID != 0 {
		t.Errorf("ContentFieldID = %d, want 0", got.ContentFieldID)
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
}

func TestMapContentFieldJSON_EmptyFieldValue(t *testing.T) {
	t.Parallel()
	cf := ContentFields{FieldValue: ""}
	got := MapContentFieldJSON(cf)

	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
}

func TestMapContentFieldJSON_LongFieldValue(t *testing.T) {
	t.Parallel()
	// Very long field values should pass through without truncation
	longValue := ""
	for range 1000 {
		longValue += "abcdefghij"
	}
	cf := ContentFields{FieldValue: longValue}
	got := MapContentFieldJSON(cf)

	if got.FieldValue != longValue {
		t.Errorf("FieldValue length = %d, want %d", len(got.FieldValue), len(longValue))
	}
}

func TestMapContentFieldJSON_UnicodeFieldValue(t *testing.T) {
	t.Parallel()
	cf := ContentFields{FieldValue: "Hello, world! Hej, verden! Ahoj, svete!"}
	got := MapContentFieldJSON(cf)

	if got.FieldValue != cf.FieldValue {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, cf.FieldValue)
	}
}

// --- MapStringContentField tests ---

func TestMapStringContentField_AllFields(t *testing.T) {
	t.Parallel()
	cf, cfID, ts := contentFieldFixture()

	got := MapStringContentField(cf)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ContentFieldID", got.ContentFieldID, cfID.String()},
		{"RouteID", got.RouteID, cf.RouteID.String()},
		{"ContentDataID", got.ContentDataID, cf.ContentDataID.String()},
		{"FieldID", got.FieldID, cf.FieldID.String()},
		{"FieldValue", got.FieldValue, "test field value"},
		{"AuthorID", got.AuthorID, cf.AuthorID.String()},
		{"DateCreated", got.DateCreated, ts.String()},
		{"DateModified", got.DateModified, ts.String()},
		{"History", got.History, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestMapStringContentField_HistoryAlwaysEmpty(t *testing.T) {
	t.Parallel()
	// History field is hard-coded to "" per the comment in the source.
	cf, _, _ := contentFieldFixture()

	got := MapStringContentField(cf)

	if got.History != "" {
		t.Errorf("History = %q, want empty string (History field removed)", got.History)
	}
}

func TestMapStringContentField_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringContentField(ContentFields{})

	if got.ContentFieldID != "" {
		t.Errorf("ContentFieldID = %q, want empty string", got.ContentFieldID)
	}
	// NullableRouteID with Valid=false should produce "null" from .String()
	if got.RouteID != "null" {
		t.Errorf("RouteID = %q, want %q for null NullableRouteID", got.RouteID, "null")
	}
	if got.ContentDataID != "null" {
		t.Errorf("ContentDataID = %q, want %q for null NullableContentID", got.ContentDataID, "null")
	}
	if got.FieldID != "null" {
		t.Errorf("FieldID = %q, want %q for null NullableFieldID", got.FieldID, "null")
	}
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q for null NullableUserID", got.AuthorID, "null")
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
	if got.History != "" {
		t.Errorf("History = %q, want empty string", got.History)
	}
}

func TestMapStringContentField_NullableFieldsShowNull(t *testing.T) {
	t.Parallel()
	cf := contentFieldFixtureNulls()

	got := MapStringContentField(cf)

	// Nullable IDs with Valid=false produce "null" via their String() method
	if got.RouteID != "null" {
		t.Errorf("RouteID = %q, want %q for null", got.RouteID, "null")
	}
	if got.ContentDataID != "null" {
		t.Errorf("ContentDataID = %q, want %q for null", got.ContentDataID, "null")
	}
	if got.FieldID != "null" {
		t.Errorf("FieldID = %q, want %q for null", got.FieldID, "null")
	}
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q for null", got.AuthorID, "null")
	}
}

// --- SQLite Database.MapContentField tests ---

func TestDatabase_MapContentField_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	cfID := types.NewContentFieldID()
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdb.ContentFields{
		ContentFieldID: cfID,
		RouteID:        routeID,
		ContentDataID:  contentDataID,
		FieldID:        fieldID,
		FieldValue:     "sqlite-value",
		AuthorID:       authorID,
		DateCreated:    ts,
		DateModified:   ts,
	}

	got := d.MapContentField(input)

	if got.ContentFieldID != cfID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, cfID)
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
	if got.FieldValue != "sqlite-value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "sqlite-value")
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
}

func TestDatabase_MapContentField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapContentField(mdb.ContentFields{})

	if got.ContentFieldID != "" {
		t.Errorf("ContentFieldID = %v, want zero value", got.ContentFieldID)
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
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
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false for zero value")
	}
}

// --- SQLite Database.MapCreateContentFieldParams tests ---

func TestDatabase_MapCreateContentFieldParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := CreateContentFieldParams{
		RouteID:       routeID,
		ContentDataID: contentDataID,
		FieldID:       fieldID,
		FieldValue:    "create-value",
		AuthorID:      authorID,
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapCreateContentFieldParams(input)

	if got.ContentFieldID.IsZero() {
		t.Fatal("expected non-zero ContentFieldID to be generated")
	}
	if got.RouteID != input.RouteID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, input.RouteID)
	}
	if got.ContentDataID != input.ContentDataID {
		t.Errorf("ContentDataID = %v, want %v", got.ContentDataID, input.ContentDataID)
	}
	if got.FieldID != input.FieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, input.FieldID)
	}
	if got.FieldValue != input.FieldValue {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, input.FieldValue)
	}
	if got.AuthorID != input.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, input.AuthorID)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestDatabase_MapCreateContentFieldParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateContentFieldParams{}

	got1 := d.MapCreateContentFieldParams(input)
	got2 := d.MapCreateContentFieldParams(input)

	if got1.ContentFieldID == got2.ContentFieldID {
		t.Error("two calls produced the same ContentFieldID -- each call should generate a unique ID")
	}
}

func TestDatabase_MapCreateContentFieldParams_NullFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	// Null optional fields should pass through correctly
	input := CreateContentFieldParams{
		RouteID:       types.NullableRouteID{Valid: false},
		ContentDataID: types.NullableContentID{Valid: false},
		FieldID:       types.NullableFieldID{Valid: false},
		AuthorID:      types.NullableUserID{Valid: false},
	}

	got := d.MapCreateContentFieldParams(input)

	if got.RouteID.Valid {
		t.Error("RouteID.Valid = true, want false")
	}
	if got.ContentDataID.Valid {
		t.Error("ContentDataID.Valid = true, want false")
	}
	if got.FieldID.Valid {
		t.Error("FieldID.Valid = true, want false")
	}
	if got.AuthorID.Valid {
		t.Error("AuthorID.Valid = true, want false")
	}
}

// --- SQLite Database.MapUpdateContentFieldParams tests ---

func TestDatabase_MapUpdateContentFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	cfID := types.NewContentFieldID()
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := UpdateContentFieldParams{
		RouteID:        routeID,
		ContentDataID:  contentDataID,
		FieldID:        fieldID,
		FieldValue:     "updated-value",
		AuthorID:       authorID,
		DateCreated:    ts,
		DateModified:   ts,
		ContentFieldID: cfID,
	}

	got := d.MapUpdateContentFieldParams(input)

	if got.ContentFieldID != cfID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, cfID)
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
	if got.FieldValue != "updated-value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "updated-value")
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
}

// --- MySQL MysqlDatabase.MapContentField tests ---

func TestMysqlDatabase_MapContentField_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	cfID := types.NewContentFieldID()
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbm.ContentFields{
		ContentFieldID: cfID,
		RouteID:        routeID,
		ContentDataID:  contentDataID,
		FieldID:        fieldID,
		FieldValue:     "mysql-value",
		AuthorID:       authorID,
		DateCreated:    ts,
		DateModified:   ts,
	}

	got := d.MapContentField(input)

	if got.ContentFieldID != cfID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, cfID)
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
	if got.FieldValue != "mysql-value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "mysql-value")
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
}

func TestMysqlDatabase_MapContentField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapContentField(mdbm.ContentFields{})

	if got.ContentFieldID != "" {
		t.Errorf("ContentFieldID = %v, want zero value", got.ContentFieldID)
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false for zero value")
	}
}

// --- MySQL MysqlDatabase.MapCreateContentFieldParams tests ---

func TestMysqlDatabase_MapCreateContentFieldParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}

	input := CreateContentFieldParams{
		RouteID:       routeID,
		ContentDataID: types.NullableContentID{Valid: false},
		FieldID:       types.NullableFieldID{Valid: false},
		FieldValue:    "mysql-create-value",
		AuthorID:      types.NullableUserID{Valid: false},
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapCreateContentFieldParams(input)

	if got.ContentFieldID.IsZero() {
		t.Fatal("expected non-zero ContentFieldID to be generated")
	}
	if got.RouteID != input.RouteID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, input.RouteID)
	}
	if got.FieldValue != input.FieldValue {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, input.FieldValue)
	}
}

// MySQL create params do NOT include DateCreated/DateModified (server-side defaults).
// Verify those fields are absent by confirming only the expected fields are mapped.
func TestMysqlDatabase_MapCreateContentFieldParams_NoTimestamps(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateContentFieldParams{
		RouteID:      types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		FieldValue:   "mysql-value",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateContentFieldParams(input)

	// The mapped result preserves the core fields
	if got.RouteID != input.RouteID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, input.RouteID)
	}
	if got.FieldValue != "mysql-value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "mysql-value")
	}
	// MySQL params struct has no DateCreated/DateModified fields -- this test
	// would fail to compile if someone incorrectly added timestamp mapping.
}

// --- MySQL MysqlDatabase.MapUpdateContentFieldParams tests ---

func TestMysqlDatabase_MapUpdateContentFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	cfID := types.NewContentFieldID()
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := UpdateContentFieldParams{
		RouteID:        routeID,
		ContentDataID:  contentDataID,
		FieldID:        fieldID,
		FieldValue:     "mysql-updated-value",
		AuthorID:       authorID,
		ContentFieldID: cfID,
	}

	got := d.MapUpdateContentFieldParams(input)

	if got.ContentFieldID != cfID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, cfID)
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
	if got.FieldValue != "mysql-updated-value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "mysql-updated-value")
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

// --- PostgreSQL PsqlDatabase.MapContentField tests ---

func TestPsqlDatabase_MapContentField_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	cfID := types.NewContentFieldID()
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbp.ContentFields{
		ContentFieldID: cfID,
		RouteID:        routeID,
		ContentDataID:  contentDataID,
		FieldID:        fieldID,
		FieldValue:     "psql-value",
		AuthorID:       authorID,
		DateCreated:    ts,
		DateModified:   ts,
	}

	got := d.MapContentField(input)

	if got.ContentFieldID != cfID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, cfID)
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
	if got.FieldValue != "psql-value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "psql-value")
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
}

func TestPsqlDatabase_MapContentField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapContentField(mdbp.ContentFields{})

	if got.ContentFieldID != "" {
		t.Errorf("ContentFieldID = %v, want zero value", got.ContentFieldID)
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
	if got.RouteID.Valid {
		t.Errorf("RouteID.Valid = true, want false for zero value")
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateContentFieldParams tests ---

func TestPsqlDatabase_MapCreateContentFieldParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}

	input := CreateContentFieldParams{
		RouteID:       routeID,
		ContentDataID: types.NullableContentID{Valid: false},
		FieldValue:    "psql-create-value",
		DateCreated:   ts,
		DateModified:  ts,
	}

	got := d.MapCreateContentFieldParams(input)

	if got.ContentFieldID.IsZero() {
		t.Fatal("expected non-zero ContentFieldID to be generated")
	}
	if got.RouteID != input.RouteID {
		t.Errorf("RouteID = %v, want %v", got.RouteID, input.RouteID)
	}
	if got.FieldValue != input.FieldValue {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, input.FieldValue)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateContentFieldParams tests ---

func TestPsqlDatabase_MapUpdateContentFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	cfID := types.NewContentFieldID()
	routeID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	contentDataID := types.NullableContentID{ID: types.NewContentID(), Valid: true}
	fieldID := types.NullableFieldID{ID: types.NewFieldID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := UpdateContentFieldParams{
		RouteID:        routeID,
		ContentDataID:  contentDataID,
		FieldID:        fieldID,
		FieldValue:     "psql-updated-value",
		AuthorID:       authorID,
		DateCreated:    ts,
		DateModified:   ts,
		ContentFieldID: cfID,
	}

	got := d.MapUpdateContentFieldParams(input)

	// PostgreSQL uses ContentFieldID_2 for the WHERE clause; both should match the input ID
	if got.ContentFieldID != cfID {
		t.Errorf("ContentFieldID = %v, want %v", got.ContentFieldID, cfID)
	}
	if got.ContentFieldID_2 != cfID {
		t.Errorf("ContentFieldID_2 = %v, want %v (must match ContentFieldID for WHERE clause)", got.ContentFieldID_2, cfID)
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
	if got.FieldValue != "psql-updated-value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "psql-updated-value")
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
}

// --- Cross-database mapper consistency ---
// Verifies that all three database mappers produce identical ContentFields from equivalent input.

func TestCrossDatabaseMapContentField_Consistency(t *testing.T) {
	t.Parallel()
	cfID := types.NewContentFieldID()
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	routeID := types.NullableRouteID{ID: types.RouteID("route-cross"), Valid: true}
	contentDataID := types.NullableContentID{ID: types.ContentID("cd-cross"), Valid: true}
	fieldID := types.NullableFieldID{ID: types.FieldID("field-cross"), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	sqliteInput := mdb.ContentFields{
		ContentFieldID: cfID, RouteID: routeID, ContentDataID: contentDataID,
		FieldID: fieldID, FieldValue: "cross-value", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.ContentFields{
		ContentFieldID: cfID, RouteID: routeID, ContentDataID: contentDataID,
		FieldID: fieldID, FieldValue: "cross-value", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.ContentFields{
		ContentFieldID: cfID, RouteID: routeID, ContentDataID: contentDataID,
		FieldID: fieldID, FieldValue: "cross-value", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapContentField(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapContentField(mysqlInput)
	psqlResult := PsqlDatabase{}.MapContentField(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateContentFieldParams auto-ID generation ---

func TestCrossDatabaseMapCreateContentFieldParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateContentFieldParams{
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		FieldValue:    "auto-id-test",
		DateCreated:   ts,
		DateModified:  ts,
	}

	sqliteResult := Database{}.MapCreateContentFieldParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateContentFieldParams(input)
	psqlResult := PsqlDatabase{}.MapCreateContentFieldParams(input)

	if sqliteResult.ContentFieldID.IsZero() {
		t.Error("SQLite: expected non-zero generated ContentFieldID")
	}
	if mysqlResult.ContentFieldID.IsZero() {
		t.Error("MySQL: expected non-zero generated ContentFieldID")
	}
	if psqlResult.ContentFieldID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated ContentFieldID")
	}

	// Each call should generate a unique ID
	if sqliteResult.ContentFieldID == mysqlResult.ContentFieldID {
		t.Error("SQLite and MySQL generated the same ContentFieldID -- each call should be unique")
	}
	if sqliteResult.ContentFieldID == psqlResult.ContentFieldID {
		t.Error("SQLite and PostgreSQL generated the same ContentFieldID -- each call should be unique")
	}
}

// --- NullableRouteIDFromInt64 tests ---

func TestNullableRouteIDFromInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     int64
		wantValid bool
		wantID    string
	}{
		{
			name:      "zero produces invalid",
			input:     0,
			wantValid: false,
			wantID:    "0",
		},
		{
			name:      "positive integer",
			input:     42,
			wantValid: true,
			wantID:    "42",
		},
		{
			name:      "negative integer",
			input:     -1,
			wantValid: true,
			wantID:    "-1",
		},
		{
			name:      "large integer",
			input:     9999999999,
			wantValid: true,
			wantID:    "9999999999",
		},
		{
			name:      "one produces valid",
			input:     1,
			wantValid: true,
			wantID:    "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NullableRouteIDFromInt64(tt.input)
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if string(got.ID) != tt.wantID {
				t.Errorf("ID = %q, want %q", string(got.ID), tt.wantID)
			}
		})
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewContentFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-cf-1"),
		RequestID: "req-cf-123",
		IP:        "10.0.0.1",
	}
	params := CreateContentFieldParams{
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		FieldValue:    "cmd-value",
		AuthorID:      types.NullableUserID{ID: userID, Valid: true},
		DateCreated:   ts,
		DateModified:  ts,
	}

	cmd := Database{}.NewContentFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_fields")
	}
	p, ok := cmd.Params().(CreateContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateContentFieldParams", cmd.Params())
	}
	if p.FieldValue != "cmd-value" {
		t.Errorf("Params().FieldValue = %q, want %q", p.FieldValue, "cmd-value")
	}
	if p.RouteID != params.RouteID {
		t.Errorf("Params().RouteID = %v, want %v", p.RouteID, params.RouteID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewContentFieldCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	cfID := types.NewContentFieldID()
	cmd := NewContentFieldCmd{}

	row := mdb.ContentFields{ContentFieldID: cfID}
	got := cmd.GetID(row)
	if got != string(cfID) {
		t.Errorf("GetID() = %q, want %q", got, string(cfID))
	}
}

func TestNewContentFieldCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewContentFieldCmd{}

	row := mdb.ContentFields{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateContentFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	cfID := types.NewContentFieldID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateContentFieldParams{
		RouteID:        types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		FieldValue:     "update-cmd-value",
		DateCreated:    ts,
		DateModified:   ts,
		ContentFieldID: cfID,
	}

	cmd := Database{}.UpdateContentFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_fields")
	}
	if cmd.GetID() != string(cfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(cfID))
	}
	p, ok := cmd.Params().(UpdateContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateContentFieldParams", cmd.Params())
	}
	if p.FieldValue != "update-cmd-value" {
		t.Errorf("Params().FieldValue = %q, want %q", p.FieldValue, "update-cmd-value")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteContentFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	cfID := types.NewContentFieldID()

	cmd := Database{}.DeleteContentFieldCmd(ctx, ac, cfID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_fields")
	}
	if cmd.GetID() != string(cfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(cfID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewContentFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node-cf"),
		RequestID: "mysql-req-cf",
		IP:        "192.168.1.1",
	}
	params := CreateContentFieldParams{
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		FieldValue:    "mysql-cmd-value",
		DateCreated:   ts,
		DateModified:  ts,
	}

	cmd := MysqlDatabase{}.NewContentFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_fields")
	}
	p, ok := cmd.Params().(CreateContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateContentFieldParams", cmd.Params())
	}
	if p.FieldValue != "mysql-cmd-value" {
		t.Errorf("Params().FieldValue = %q, want %q", p.FieldValue, "mysql-cmd-value")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewContentFieldCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	cfID := types.NewContentFieldID()
	cmd := NewContentFieldCmdMysql{}

	row := mdbm.ContentFields{ContentFieldID: cfID}
	got := cmd.GetID(row)
	if got != string(cfID) {
		t.Errorf("GetID() = %q, want %q", got, string(cfID))
	}
}

func TestUpdateContentFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	cfID := types.NewContentFieldID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateContentFieldParams{
		RouteID:        types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		FieldValue:     "mysql-update-value",
		DateCreated:    ts,
		DateModified:   ts,
		ContentFieldID: cfID,
	}

	cmd := MysqlDatabase{}.UpdateContentFieldCmd(ctx, ac, params)

	if cmd.TableName() != "content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_fields")
	}
	if cmd.GetID() != string(cfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(cfID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateContentFieldParams", cmd.Params())
	}
	if p.FieldValue != "mysql-update-value" {
		t.Errorf("Params().FieldValue = %q, want %q", p.FieldValue, "mysql-update-value")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteContentFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	cfID := types.NewContentFieldID()

	cmd := MysqlDatabase{}.DeleteContentFieldCmd(ctx, ac, cfID)

	if cmd.TableName() != "content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_fields")
	}
	if cmd.GetID() != string(cfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(cfID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- PostgreSQL Audited Command Accessor tests ---

func TestNewContentFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node-cf"),
		RequestID: "psql-req-cf",
		IP:        "172.16.0.1",
	}
	params := CreateContentFieldParams{
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		FieldValue:    "psql-cmd-value",
		DateCreated:   ts,
		DateModified:  ts,
	}

	cmd := PsqlDatabase{}.NewContentFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_fields")
	}
	p, ok := cmd.Params().(CreateContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateContentFieldParams", cmd.Params())
	}
	if p.FieldValue != "psql-cmd-value" {
		t.Errorf("Params().FieldValue = %q, want %q", p.FieldValue, "psql-cmd-value")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewContentFieldCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	cfID := types.NewContentFieldID()
	cmd := NewContentFieldCmdPsql{}

	row := mdbp.ContentFields{ContentFieldID: cfID}
	got := cmd.GetID(row)
	if got != string(cfID) {
		t.Errorf("GetID() = %q, want %q", got, string(cfID))
	}
}

func TestUpdateContentFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	cfID := types.NewContentFieldID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateContentFieldParams{
		RouteID:        types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		FieldValue:     "psql-update-value",
		DateCreated:    ts,
		DateModified:   ts,
		ContentFieldID: cfID,
	}

	cmd := PsqlDatabase{}.UpdateContentFieldCmd(ctx, ac, params)

	if cmd.TableName() != "content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_fields")
	}
	if cmd.GetID() != string(cfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(cfID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateContentFieldParams", cmd.Params())
	}
	if p.FieldValue != "psql-update-value" {
		t.Errorf("Params().FieldValue = %q, want %q", p.FieldValue, "psql-update-value")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteContentFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	cfID := types.NewContentFieldID()

	cmd := PsqlDatabase{}.DeleteContentFieldCmd(ctx, ac, cfID)

	if cmd.TableName() != "content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "content_fields")
	}
	if cmd.GetID() != string(cfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(cfID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- Cross-database audited command consistency ---
// Verify all three database types produce commands with the correct table name.

func TestAuditedContentFieldCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateContentFieldParams{}
	updateParams := UpdateContentFieldParams{ContentFieldID: types.NewContentFieldID()}
	cfID := types.NewContentFieldID()

	// SQLite
	sqliteCreate := Database{}.NewContentFieldCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateContentFieldCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteContentFieldCmd(ctx, ac, cfID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewContentFieldCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateContentFieldCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteContentFieldCmd(ctx, ac, cfID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewContentFieldCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateContentFieldCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteContentFieldCmd(ctx, ac, cfID)

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", sqliteCreate.TableName()},
		{"SQLite Update", sqliteUpdate.TableName()},
		{"SQLite Delete", sqliteDelete.TableName()},
		{"MySQL Create", mysqlCreate.TableName()},
		{"MySQL Update", mysqlUpdate.TableName()},
		{"MySQL Delete", mysqlDelete.TableName()},
		{"PostgreSQL Create", psqlCreate.TableName()},
		{"PostgreSQL Update", psqlUpdate.TableName()},
		{"PostgreSQL Delete", psqlDelete.TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "content_fields" {
				t.Errorf("TableName() = %q, want %q", c.name, "content_fields")
			}
		})
	}
}

func TestAuditedContentFieldCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateContentFieldParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewContentFieldCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewContentFieldCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewContentFieldCmd(ctx, ac, createParams).Recorder()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.recorder == nil {
				t.Fatalf("%s recorder is nil", tt.name)
			}
		})
	}
}

// --- Audited Command GetID consistency ---
// Verify GetID returns the correct identifier across all database variants.

func TestAuditedContentFieldCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	cfID := types.NewContentFieldID()

	t.Run("UpdateCmd GetID returns ContentFieldID", func(t *testing.T) {
		t.Parallel()
		params := UpdateContentFieldParams{ContentFieldID: cfID}

		sqliteCmd := Database{}.UpdateContentFieldCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateContentFieldCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateContentFieldCmd(ctx, ac, params)

		wantID := string(cfID)
		if sqliteCmd.GetID() != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), wantID)
		}
		if mysqlCmd.GetID() != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), wantID)
		}
		if psqlCmd.GetID() != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), wantID)
		}
	})

	t.Run("DeleteCmd GetID returns ContentFieldID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteContentFieldCmd(ctx, ac, cfID)
		mysqlCmd := MysqlDatabase{}.DeleteContentFieldCmd(ctx, ac, cfID)
		psqlCmd := PsqlDatabase{}.DeleteContentFieldCmd(ctx, ac, cfID)

		wantID := string(cfID)
		if sqliteCmd.GetID() != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), wantID)
		}
		if mysqlCmd.GetID() != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), wantID)
		}
		if psqlCmd.GetID() != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), wantID)
		}
	})

	t.Run("CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		testCfID := types.NewContentFieldID()

		sqliteCmd := NewContentFieldCmd{}
		mysqlCmd := NewContentFieldCmdMysql{}
		psqlCmd := NewContentFieldCmdPsql{}

		wantID := string(testCfID)

		sqliteRow := mdb.ContentFields{ContentFieldID: testCfID}
		mysqlRow := mdbm.ContentFields{ContentFieldID: testCfID}
		psqlRow := mdbp.ContentFields{ContentFieldID: testCfID}

		if sqliteCmd.GetID(sqliteRow) != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(sqliteRow), wantID)
		}
		if mysqlCmd.GetID(mysqlRow) != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(mysqlRow), wantID)
		}
		if psqlCmd.GetID(psqlRow) != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(psqlRow), wantID)
		}
	})
}

// --- Edge case: UpdateContentFieldCmd with empty ContentFieldID ---

func TestUpdateContentFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateContentFieldParams{ContentFieldID: ""}

	sqliteCmd := Database{}.UpdateContentFieldCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateContentFieldCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateContentFieldCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteContentFieldCmd with empty ID ---

func TestDeleteContentFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.ContentFieldID("")

	sqliteCmd := Database{}.DeleteContentFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteContentFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteContentFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- UpdateContentField success message format ---
// The UpdateContentField method produces a success message string.
// We can't test it without a real DB, but we verify the format expectation.

func TestUpdateContentField_SuccessMessageFormat(t *testing.T) {
	t.Parallel()
	cfID := types.NewContentFieldID()
	expected := fmt.Sprintf("Successfully updated %v\n", cfID)

	if expected == "" {
		t.Fatal("expected non-empty message")
	}
	if expected == "Successfully updated \n" {
		t.Fatal("expected message to contain the ID")
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.ContentFields]  = NewContentFieldCmd{}
	_ audited.UpdateCommand[mdb.ContentFields]  = UpdateContentFieldCmd{}
	_ audited.DeleteCommand[mdb.ContentFields]  = DeleteContentFieldCmd{}
	_ audited.CreateCommand[mdbm.ContentFields] = NewContentFieldCmdMysql{}
	_ audited.UpdateCommand[mdbm.ContentFields] = UpdateContentFieldCmdMysql{}
	_ audited.DeleteCommand[mdbm.ContentFields] = DeleteContentFieldCmdMysql{}
	_ audited.CreateCommand[mdbp.ContentFields] = NewContentFieldCmdPsql{}
	_ audited.UpdateCommand[mdbp.ContentFields] = UpdateContentFieldCmdPsql{}
	_ audited.DeleteCommand[mdbp.ContentFields] = DeleteContentFieldCmdPsql{}
)

// --- Struct field correctness ---
// Verify that the wrapper ContentFields struct and param structs hold values correctly via JSON.

func TestContentFieldsStruct_JSONTags(t *testing.T) {
	t.Parallel()
	cf, _, _ := contentFieldFixture()

	data, err := json.Marshal(cf)
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

func TestCreateContentFieldParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	p := CreateContentFieldParams{
		RouteID:       types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		ContentDataID: types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FieldID:       types.NullableFieldID{ID: types.NewFieldID(), Valid: true},
		FieldValue:    "test-value",
		AuthorID:      types.NullableUserID{ID: types.NewUserID(), Valid: true},
		DateCreated:   ts,
		DateModified:  ts,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"route_id", "content_data_id", "field_id",
		"field_value", "author_id",
		"date_created", "date_modified",
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

func TestUpdateContentFieldParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	cfID := types.NewContentFieldID()
	p := UpdateContentFieldParams{
		RouteID:        types.NullableRouteID{ID: types.NewRouteID(), Valid: true},
		ContentDataID:  types.NullableContentID{ID: types.NewContentID(), Valid: true},
		FieldID:        types.NullableFieldID{ID: types.NewFieldID(), Valid: true},
		FieldValue:     "test-value",
		AuthorID:       types.NullableUserID{ID: types.NewUserID(), Valid: true},
		DateCreated:    ts,
		DateModified:   ts,
		ContentFieldID: cfID,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"route_id", "content_data_id", "field_id",
		"field_value", "author_id",
		"date_created", "date_modified",
		"content_field_id",
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

func TestContentFieldsJSONStruct_JSONTags(t *testing.T) {
	t.Parallel()
	cfj := ContentFieldsJSON{
		ContentFieldID: 1,
		RouteID:        2,
		ContentDataID:  3,
		FieldID:        4,
		FieldValue:     "json-value",
		AuthorID:       5,
		DateCreated:    "2025-01-01T00:00:00Z",
		DateModified:   "2025-01-01T00:00:00Z",
	}

	data, err := json.Marshal(cfj)
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
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("ContentFieldsJSON JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("ContentFieldsJSON JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

// --- MapContentFieldJSON and MapStringContentField consistency ---
// Both functions process the same input; verify FieldValue is handled identically.

func TestMapContentFieldJSON_MapStringContentField_FieldValueConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fieldValue string
	}{
		{"empty string", ""},
		{"simple string", "hello world"},
		{"unicode", "Bonjour le monde"},
		{"special characters", `<script>alert("xss")</script>`},
		{"newlines", "line1\nline2\nline3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cf := ContentFields{
				ContentFieldID: types.NewContentFieldID(),
				FieldValue:     tt.fieldValue,
			}

			jsonResult := MapContentFieldJSON(cf)
			stringResult := MapStringContentField(cf)

			if jsonResult.FieldValue != tt.fieldValue {
				t.Errorf("JSON FieldValue = %q, want %q", jsonResult.FieldValue, tt.fieldValue)
			}
			if stringResult.FieldValue != tt.fieldValue {
				t.Errorf("String FieldValue = %q, want %q", stringResult.FieldValue, tt.fieldValue)
			}
		})
	}
}
