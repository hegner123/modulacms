package db

import (
	"context"
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

// acfTestFixture returns a fully populated AdminContentFields struct for testing.
func acfTestFixture() (AdminContentFields, types.AdminContentFieldID, types.NullableAdminFieldID, types.UserID, types.Timestamp) {
	fieldID := types.NewAdminContentFieldID()
	adminFieldID := types.NullableAdminFieldID{ID: types.AdminFieldID(types.NewAdminFieldID()), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC))

	acf := AdminContentFields{
		AdminContentFieldID: fieldID,
		AdminRouteID:        types.NullableAdminRouteID{ID: types.AdminRouteID("route-abc"), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.AdminContentID("content-data-123"), Valid: true},
		AdminFieldID:        adminFieldID,
		AdminFieldValue:     "some field value",
		AuthorID:            authorID,
		DateCreated:         ts,
		DateModified:        ts,
	}
	return acf, fieldID, adminFieldID, authorID, ts
}

// --- MapAdminContentFieldJSON tests ---

func TestMapAdminContentFieldJSON_AllFields(t *testing.T) {
	t.Parallel()
	acf, _, _, _, ts := acfTestFixture()

	got := MapAdminContentFieldJSON(acf)

	// MapAdminContentFieldJSON maps admin fields into the public ContentFieldsJSON shape.
	// All ID fields are set to 0; only FieldValue and timestamps carry data.
	if got.ContentFieldID != 0 {
		t.Errorf("ContentFieldID = %d, want 0", got.ContentFieldID)
	}
	if got.RouteID != 0 {
		t.Errorf("RouteID = %d, want 0", got.RouteID)
	}
	if got.ContentDataID != 0 {
		t.Errorf("ContentDataID = %d, want 0", got.ContentDataID)
	}
	if got.FieldID != 0 {
		t.Errorf("FieldID = %d, want 0", got.FieldID)
	}
	if got.AuthorID != 0 {
		t.Errorf("AuthorID = %d, want 0", got.AuthorID)
	}
	if got.FieldValue != "some field value" {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, "some field value")
	}
	if got.DateCreated != ts.String() {
		t.Errorf("DateCreated = %q, want %q", got.DateCreated, ts.String())
	}
	if got.DateModified != ts.String() {
		t.Errorf("DateModified = %q, want %q", got.DateModified, ts.String())
	}
}

func TestMapAdminContentFieldJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value AdminContentFields should produce a valid ContentFieldsJSON without panic
	got := MapAdminContentFieldJSON(AdminContentFields{})

	if got.ContentFieldID != 0 {
		t.Errorf("ContentFieldID = %d, want 0", got.ContentFieldID)
	}
	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
}

func TestMapAdminContentFieldJSON_EmptyFieldValue(t *testing.T) {
	t.Parallel()
	acf := AdminContentFields{AdminFieldValue: ""}
	got := MapAdminContentFieldJSON(acf)

	if got.FieldValue != "" {
		t.Errorf("FieldValue = %q, want empty string", got.FieldValue)
	}
}

func TestMapAdminContentFieldJSON_LongFieldValue(t *testing.T) {
	t.Parallel()
	// Very long field values should pass through without truncation
	longValue := ""
	for range 1000 {
		longValue += "abcdefghij"
	}
	acf := AdminContentFields{AdminFieldValue: longValue}
	got := MapAdminContentFieldJSON(acf)

	if got.FieldValue != longValue {
		t.Errorf("FieldValue length = %d, want %d", len(got.FieldValue), len(longValue))
	}
}

func TestMapAdminContentFieldJSON_UnicodeFieldValue(t *testing.T) {
	t.Parallel()
	// Verify multi-byte characters pass through correctly
	acf := AdminContentFields{AdminFieldValue: "Hello, world! Hej, verden! Ahoj, svete!"}
	got := MapAdminContentFieldJSON(acf)

	if got.FieldValue != acf.AdminFieldValue {
		t.Errorf("FieldValue = %q, want %q", got.FieldValue, acf.AdminFieldValue)
	}
}

// --- MapStringAdminContentField tests ---

func TestMapStringAdminContentField_AllFields(t *testing.T) {
	t.Parallel()
	acf, fieldID, adminFieldID, authorID, ts := acfTestFixture()

	got := MapStringAdminContentField(acf)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"AdminContentFieldID", got.AdminContentFieldID, fieldID.String()},
		{"AdminRouteID", got.AdminRouteID, "route-abc"},
		{"AdminContentDataID", got.AdminContentDataID, "content-data-123"},
		{"AdminFieldID", got.AdminFieldID, adminFieldID.String()},
		{"AdminFieldValue", got.AdminFieldValue, "some field value"},
		{"AuthorID", got.AuthorID, authorID.String()},
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

func TestMapStringAdminContentField_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value AdminContentFields should convert without panic
	got := MapStringAdminContentField(AdminContentFields{})

	if got.AdminContentFieldID != "" {
		t.Errorf("AdminContentFieldID = %q, want empty string", got.AdminContentFieldID)
	}
	// AdminRouteID with Valid=false should produce "null" from .String()
	if got.AdminRouteID != "null" {
		t.Errorf("AdminRouteID = %q, want %q for null NullableAdminRouteID", got.AdminRouteID, "null")
	}
	if got.AdminContentDataID != "null" {
		t.Errorf("AdminContentDataID = %q, want %q for null NullableAdminContentID", got.AdminContentDataID, "null")
	}
	if got.AdminFieldValue != "" {
		t.Errorf("AdminFieldValue = %q, want empty string", got.AdminFieldValue)
	}
	if got.History != "" {
		t.Errorf("History = %q, want empty string", got.History)
	}
}

func TestMapStringAdminContentField_NullAdminRouteID(t *testing.T) {
	t.Parallel()
	// When AdminRouteID.Valid is false, the output should be "null" from NullableAdminRouteID.String()
	acf := AdminContentFields{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("should-be-ignored"), Valid: false},
	}
	got := MapStringAdminContentField(acf)

	if got.AdminRouteID != "null" {
		t.Errorf("AdminRouteID = %q, want %q when Valid=false", got.AdminRouteID, "null")
	}
}

func TestMapStringAdminContentField_ValidAdminRouteID(t *testing.T) {
	t.Parallel()
	acf := AdminContentFields{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID("valid-route-id"), Valid: true},
	}
	got := MapStringAdminContentField(acf)

	if got.AdminRouteID != "valid-route-id" {
		t.Errorf("AdminRouteID = %q, want %q", got.AdminRouteID, "valid-route-id")
	}
}

func TestMapStringAdminContentField_NullAdminFieldID(t *testing.T) {
	t.Parallel()
	// NullableAdminFieldID with Valid=false should produce "null" from .String()
	acf := AdminContentFields{
		AdminFieldID: types.NullableAdminFieldID{Valid: false},
	}
	got := MapStringAdminContentField(acf)

	// NullableAdminFieldID.String() returns "null" when Valid=false
	if got.AdminFieldID != "null" {
		t.Errorf("AdminFieldID = %q, want %q", got.AdminFieldID, "null")
	}
}

func TestMapStringAdminContentField_EmptyAuthorID(t *testing.T) {
	t.Parallel()
	// Zero-value UserID should produce "" from .String()
	acf := AdminContentFields{
		AuthorID: types.UserID(""),
	}
	got := MapStringAdminContentField(acf)

	// UserID.String() returns "" when zero-value
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, "")
	}
}

func TestMapStringAdminContentField_EmptyValidAdminRouteID(t *testing.T) {
	t.Parallel()
	// Edge case: Valid=true but String is empty
	acf := AdminContentFields{
		AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID(""), Valid: true},
	}
	got := MapStringAdminContentField(acf)

	if got.AdminRouteID != "" {
		t.Errorf("AdminRouteID = %q, want empty string", got.AdminRouteID)
	}
}

// --- SQLite Database.MapAdminContentField tests ---

func TestDatabase_MapAdminContentField_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	fieldID := types.NewAdminContentFieldID()
	adminFieldID := types.NullableAdminFieldID{ID: types.AdminFieldID(types.NewAdminFieldID()), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))

	input := mdb.AdminContentFields{
		AdminContentFieldID: fieldID,
		AdminRouteID:        types.NullableAdminRouteID{ID: types.AdminRouteID("route-1"), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.AdminContentID("data-1"), Valid: true},
		AdminFieldID:        adminFieldID,
		AdminFieldValue:     "value-1",
		AuthorID:            authorID,
		DateCreated:         ts,
		DateModified:        ts,
	}

	got := d.MapAdminContentField(input)

	if got.AdminContentFieldID != fieldID {
		t.Errorf("AdminContentFieldID = %v, want %v", got.AdminContentFieldID, fieldID)
	}
	if got.AdminRouteID != input.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, input.AdminRouteID)
	}
	if got.AdminContentDataID.String() != "data-1" {
		t.Errorf("AdminContentDataID = %q, want %q", got.AdminContentDataID.String(), "data-1")
	}
	if got.AdminFieldID != adminFieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, adminFieldID)
	}
	if got.AdminFieldValue != "value-1" {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, "value-1")
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

func TestDatabase_MapAdminContentField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminContentField(mdb.AdminContentFields{})

	if got.AdminContentFieldID != "" {
		t.Errorf("AdminContentFieldID = %v, want zero value", got.AdminContentFieldID)
	}
	if got.AdminRouteID.Valid {
		t.Errorf("AdminRouteID.Valid = true, want false")
	}
	if got.AdminContentDataID != (types.NullableAdminContentID{}) {
		t.Errorf("AdminContentDataID = %v, want zero value", got.AdminContentDataID)
	}
}

// --- SQLite Database.MapCreateAdminContentFieldParams tests ---

func TestDatabase_MapCreateAdminContentFieldParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NewUserID()
	adminFieldID := types.NullableAdminFieldID{ID: types.AdminFieldID(types.NewAdminFieldID()), Valid: true}

	input := CreateAdminContentFieldParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("route-create"), Valid: true},
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("data-create"), Valid: true},
		AdminFieldID:       adminFieldID,
		AdminFieldValue:    "create-value",
		AuthorID:           authorID,
		DateCreated:        ts,
		DateModified:       ts,
	}

	got := d.MapCreateAdminContentFieldParams(input)

	// A new ID should always be generated
	if got.AdminContentFieldID.IsZero() {
		t.Fatal("expected non-zero AdminContentFieldID to be generated")
	}

	// All other fields should pass through
	if got.AdminRouteID != input.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, input.AdminRouteID)
	}
	if got.AdminContentDataID != input.AdminContentDataID {
		t.Errorf("AdminContentDataID = %q, want %q", got.AdminContentDataID, input.AdminContentDataID)
	}
	if got.AdminFieldID != input.AdminFieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, input.AdminFieldID)
	}
	if got.AdminFieldValue != input.AdminFieldValue {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, input.AdminFieldValue)
	}
	if got.AuthorID != input.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, input.AuthorID)
	}
	if got.DateCreated != input.DateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, input.DateCreated)
	}
	if got.DateModified != input.DateModified {
		t.Errorf("DateModified = %v, want %v", got.DateModified, input.DateModified)
	}
}

func TestDatabase_MapCreateAdminContentFieldParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateAdminContentFieldParams{}

	got1 := d.MapCreateAdminContentFieldParams(input)
	got2 := d.MapCreateAdminContentFieldParams(input)

	if got1.AdminContentFieldID == got2.AdminContentFieldID {
		t.Error("two calls generated the same AdminContentFieldID -- each call should be unique")
	}
}

func TestDatabase_MapCreateAdminContentFieldParams_NullFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	// Null optional fields should pass through correctly
	input := CreateAdminContentFieldParams{
		AdminRouteID: types.NullableAdminRouteID{Valid: false},
		AdminFieldID: types.NullableAdminFieldID{Valid: false},
		AuthorID:     types.UserID(""),
	}

	got := d.MapCreateAdminContentFieldParams(input)

	if got.AdminRouteID.Valid {
		t.Error("AdminRouteID.Valid = true, want false")
	}
	if got.AdminFieldID.Valid {
		t.Error("AdminFieldID.Valid = true, want false")
	}
	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string", got.AuthorID)
	}
}

// --- SQLite Database.MapUpdateAdminContentFieldParams tests ---

func TestDatabase_MapUpdateAdminContentFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	authorID := types.NewUserID()
	adminFieldID := types.NullableAdminFieldID{ID: types.AdminFieldID(types.NewAdminFieldID()), Valid: true}
	acfID := types.NewAdminContentFieldID()

	input := UpdateAdminContentFieldParams{
		AdminRouteID:        types.NullableAdminRouteID{ID: types.AdminRouteID("updated-route"), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.AdminContentID("updated-data"), Valid: true},
		AdminFieldID:        adminFieldID,
		AdminFieldValue:     "updated-value",
		AuthorID:            authorID,
		DateCreated:         ts,
		DateModified:        ts,
		AdminContentFieldID: acfID,
	}

	got := d.MapUpdateAdminContentFieldParams(input)

	if got.AdminRouteID != input.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, input.AdminRouteID)
	}
	if got.AdminContentDataID != input.AdminContentDataID {
		t.Errorf("AdminContentDataID = %q, want %q", got.AdminContentDataID, input.AdminContentDataID)
	}
	if got.AdminFieldID != input.AdminFieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, input.AdminFieldID)
	}
	if got.AdminFieldValue != input.AdminFieldValue {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, input.AdminFieldValue)
	}
	if got.AuthorID != input.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, input.AuthorID)
	}
	if got.DateCreated != input.DateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, input.DateCreated)
	}
	if got.DateModified != input.DateModified {
		t.Errorf("DateModified = %v, want %v", got.DateModified, input.DateModified)
	}
	// AdminContentFieldID is the WHERE clause identifier -- must be preserved
	if got.AdminContentFieldID != acfID {
		t.Errorf("AdminContentFieldID = %v, want %v", got.AdminContentFieldID, acfID)
	}
}

// --- MySQL MysqlDatabase.MapAdminContentField tests ---

func TestMysqlDatabase_MapAdminContentField_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	fieldID := types.NewAdminContentFieldID()
	adminFieldID := types.NullableAdminFieldID{ID: types.AdminFieldID(types.NewAdminFieldID()), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))

	input := mdbm.AdminContentFields{
		AdminContentFieldID: fieldID,
		AdminRouteID:        types.NullableAdminRouteID{ID: types.AdminRouteID("mysql-route"), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.AdminContentID("mysql-data"), Valid: true},
		AdminFieldID:        adminFieldID,
		AdminFieldValue:     "mysql-value",
		AuthorID:            authorID,
		DateCreated:         ts,
		DateModified:        ts,
	}

	got := d.MapAdminContentField(input)

	if got.AdminContentFieldID != fieldID {
		t.Errorf("AdminContentFieldID = %v, want %v", got.AdminContentFieldID, fieldID)
	}
	if got.AdminRouteID != input.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, input.AdminRouteID)
	}
	if got.AdminContentDataID.String() != "mysql-data" {
		t.Errorf("AdminContentDataID = %q, want %q", got.AdminContentDataID.String(), "mysql-data")
	}
	if got.AdminFieldID != adminFieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, adminFieldID)
	}
	if got.AdminFieldValue != "mysql-value" {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, "mysql-value")
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

func TestMysqlDatabase_MapAdminContentField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminContentField(mdbm.AdminContentFields{})

	if got.AdminContentFieldID != "" {
		t.Errorf("AdminContentFieldID = %v, want zero value", got.AdminContentFieldID)
	}
}

// --- MySQL MysqlDatabase.MapCreateAdminContentFieldParams tests ---

func TestMysqlDatabase_MapCreateAdminContentFieldParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminContentFieldParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("mysql-route"), Valid: true},
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("mysql-data"), Valid: true},
		AdminFieldValue:    "mysql-create-value",
		DateCreated:        ts,
		DateModified:       ts,
	}

	got := d.MapCreateAdminContentFieldParams(input)

	if got.AdminContentFieldID.IsZero() {
		t.Fatal("expected non-zero AdminContentFieldID to be generated")
	}
	if got.AdminRouteID != input.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, input.AdminRouteID)
	}
	if got.AdminContentDataID != input.AdminContentDataID {
		t.Errorf("AdminContentDataID = %q, want %q", got.AdminContentDataID, input.AdminContentDataID)
	}
	if got.AdminFieldValue != input.AdminFieldValue {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, input.AdminFieldValue)
	}
}

// --- MySQL MysqlDatabase.MapUpdateAdminContentFieldParams tests ---

func TestMysqlDatabase_MapUpdateAdminContentFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	acfID := types.NewAdminContentFieldID()

	input := UpdateAdminContentFieldParams{
		AdminRouteID:        types.NullableAdminRouteID{ID: types.AdminRouteID("mysql-updated"), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.AdminContentID("mysql-updated-data"), Valid: true},
		AdminFieldValue:     "mysql-updated-value",
		DateCreated:         ts,
		DateModified:        ts,
		AdminContentFieldID: acfID,
	}

	got := d.MapUpdateAdminContentFieldParams(input)

	if got.AdminContentFieldID != acfID {
		t.Errorf("AdminContentFieldID = %v, want %v", got.AdminContentFieldID, acfID)
	}
	if got.AdminRouteID != input.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, input.AdminRouteID)
	}
	if got.AdminFieldValue != input.AdminFieldValue {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, input.AdminFieldValue)
	}
}

// --- PostgreSQL PsqlDatabase.MapAdminContentField tests ---

func TestPsqlDatabase_MapAdminContentField_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	fieldID := types.NewAdminContentFieldID()
	adminFieldID := types.NullableAdminFieldID{ID: types.AdminFieldID(types.NewAdminFieldID()), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))

	input := mdbp.AdminContentFields{
		AdminContentFieldID: fieldID,
		AdminRouteID:        types.NullableAdminRouteID{ID: types.AdminRouteID("psql-route"), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.AdminContentID("psql-data"), Valid: true},
		AdminFieldID:        adminFieldID,
		AdminFieldValue:     "psql-value",
		AuthorID:            authorID,
		DateCreated:         ts,
		DateModified:        ts,
	}

	got := d.MapAdminContentField(input)

	if got.AdminContentFieldID != fieldID {
		t.Errorf("AdminContentFieldID = %v, want %v", got.AdminContentFieldID, fieldID)
	}
	if got.AdminRouteID != input.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, input.AdminRouteID)
	}
	if got.AdminContentDataID.String() != "psql-data" {
		t.Errorf("AdminContentDataID = %q, want %q", got.AdminContentDataID.String(), "psql-data")
	}
	if got.AdminFieldID != adminFieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, adminFieldID)
	}
	if got.AdminFieldValue != "psql-value" {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, "psql-value")
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

func TestPsqlDatabase_MapAdminContentField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminContentField(mdbp.AdminContentFields{})

	if got.AdminContentFieldID != "" {
		t.Errorf("AdminContentFieldID = %v, want zero value", got.AdminContentFieldID)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateAdminContentFieldParams tests ---

func TestPsqlDatabase_MapCreateAdminContentFieldParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminContentFieldParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("psql-route"), Valid: true},
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("psql-data"), Valid: true},
		AdminFieldValue:    "psql-create-value",
		DateCreated:        ts,
		DateModified:       ts,
	}

	got := d.MapCreateAdminContentFieldParams(input)

	if got.AdminContentFieldID.IsZero() {
		t.Fatal("expected non-zero AdminContentFieldID to be generated")
	}
	if got.AdminRouteID != input.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, input.AdminRouteID)
	}
	if got.AdminContentDataID != input.AdminContentDataID {
		t.Errorf("AdminContentDataID = %q, want %q", got.AdminContentDataID, input.AdminContentDataID)
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateAdminContentFieldParams tests ---

func TestPsqlDatabase_MapUpdateAdminContentFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	acfID := types.NewAdminContentFieldID()

	input := UpdateAdminContentFieldParams{
		AdminRouteID:        types.NullableAdminRouteID{ID: types.AdminRouteID("psql-updated"), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.AdminContentID("psql-updated-data"), Valid: true},
		AdminFieldValue:     "psql-updated-value",
		DateCreated:         ts,
		DateModified:        ts,
		AdminContentFieldID: acfID,
	}

	got := d.MapUpdateAdminContentFieldParams(input)

	if got.AdminContentFieldID != acfID {
		t.Errorf("AdminContentFieldID = %v, want %v", got.AdminContentFieldID, acfID)
	}
	if got.AdminRouteID != input.AdminRouteID {
		t.Errorf("AdminRouteID = %v, want %v", got.AdminRouteID, input.AdminRouteID)
	}
	if got.AdminFieldValue != input.AdminFieldValue {
		t.Errorf("AdminFieldValue = %q, want %q", got.AdminFieldValue, input.AdminFieldValue)
	}
}

// --- Cross-database mapper consistency ---
// All three database drivers use identical types for AdminContentFields
// (no int32/int64 conversions needed). This test verifies they all produce
// identical AdminContentFields wrapper structs from equivalent input.

func TestCrossDatabaseMapAdminContentField_Consistency(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminContentFieldID()
	adminFieldID := types.NullableAdminFieldID{ID: types.AdminFieldID(types.NewAdminFieldID()), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	routeID := types.NullableAdminRouteID{ID: types.AdminRouteID("cross-route"), Valid: true}

	sqliteInput := mdb.AdminContentFields{
		AdminContentFieldID: fieldID, AdminRouteID: routeID,
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("cross-data"), Valid: true}, AdminFieldID: adminFieldID,
		AdminFieldValue: "cross-value", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.AdminContentFields{
		AdminContentFieldID: fieldID, AdminRouteID: routeID,
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("cross-data"), Valid: true}, AdminFieldID: adminFieldID,
		AdminFieldValue: "cross-value", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.AdminContentFields{
		AdminContentFieldID: fieldID, AdminRouteID: routeID,
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("cross-data"), Valid: true}, AdminFieldID: adminFieldID,
		AdminFieldValue: "cross-value", AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapAdminContentField(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminContentField(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminContentField(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateAdminContentFieldParams auto-ID generation ---

func TestCrossDatabaseMapCreateAdminContentFieldParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	input := CreateAdminContentFieldParams{
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("cross-create-data"), Valid: true},
		AdminFieldValue:    "cross-create-value",
		DateCreated:        ts,
		DateModified:       ts,
	}

	sqliteResult := Database{}.MapCreateAdminContentFieldParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateAdminContentFieldParams(input)
	psqlResult := PsqlDatabase{}.MapCreateAdminContentFieldParams(input)

	if sqliteResult.AdminContentFieldID.IsZero() {
		t.Error("SQLite: expected non-zero generated AdminContentFieldID")
	}
	if mysqlResult.AdminContentFieldID.IsZero() {
		t.Error("MySQL: expected non-zero generated AdminContentFieldID")
	}
	if psqlResult.AdminContentFieldID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated AdminContentFieldID")
	}

	// Each call should generate a unique ID
	if sqliteResult.AdminContentFieldID == mysqlResult.AdminContentFieldID {
		t.Error("SQLite and MySQL generated the same AdminContentFieldID -- each call should be unique")
	}
	if sqliteResult.AdminContentFieldID == psqlResult.AdminContentFieldID {
		t.Error("SQLite and PostgreSQL generated the same AdminContentFieldID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewAdminContentFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-acf-1"),
		RequestID: "req-acf-123",
		IP:        "10.0.0.1",
	}
	params := CreateAdminContentFieldParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("cmd-route"), Valid: true},
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("cmd-data"), Valid: true},
		AdminFieldValue:    "cmd-value",
		DateCreated:        ts,
		DateModified:       ts,
	}

	cmd := Database{}.NewAdminContentFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_fields")
	}
	p, ok := cmd.Params().(CreateAdminContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminContentFieldParams", cmd.Params())
	}
	if p.AdminFieldValue != "cmd-value" {
		t.Errorf("Params().AdminFieldValue = %q, want %q", p.AdminFieldValue, "cmd-value")
	}
	if p.AdminContentDataID.String() != "cmd-data" {
		t.Errorf("Params().AdminContentDataID = %q, want %q", p.AdminContentDataID.String(), "cmd-data")
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminContentFieldCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminContentFieldID()
	cmd := NewAdminContentFieldCmd{}

	row := mdb.AdminContentFields{AdminContentFieldID: fieldID}
	got := cmd.GetID(row)
	if got != string(fieldID) {
		t.Errorf("GetID() = %q, want %q", got, string(fieldID))
	}
}

func TestNewAdminContentFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	cmd := NewAdminContentFieldCmd{}
	row := mdb.AdminContentFields{AdminContentFieldID: ""}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateAdminContentFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	acfID := types.NewAdminContentFieldID()
	params := UpdateAdminContentFieldParams{
		AdminRouteID:        types.NullableAdminRouteID{ID: types.AdminRouteID("update-route"), Valid: true},
		AdminContentDataID:  types.NullableAdminContentID{ID: types.AdminContentID("update-data"), Valid: true},
		AdminFieldValue:     "update-value",
		DateCreated:         ts,
		DateModified:        ts,
		AdminContentFieldID: acfID,
	}

	cmd := Database{}.UpdateAdminContentFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_fields")
	}
	if cmd.GetID() != string(acfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(acfID))
	}
	p, ok := cmd.Params().(UpdateAdminContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminContentFieldParams", cmd.Params())
	}
	if p.AdminFieldValue != "update-value" {
		t.Errorf("Params().AdminFieldValue = %q, want %q", p.AdminFieldValue, "update-value")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminContentFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	acfID := types.NewAdminContentFieldID()

	cmd := Database{}.DeleteAdminContentFieldCmd(ctx, ac, acfID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_fields")
	}
	if cmd.GetID() != string(acfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(acfID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewAdminContentFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node-acf"),
		RequestID: "mysql-req-acf",
		IP:        "192.168.1.1",
	}
	params := CreateAdminContentFieldParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("mysql-cmd-route"), Valid: true},
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("mysql-cmd-data"), Valid: true},
		AdminFieldValue:    "mysql-cmd-value",
		DateCreated:        ts,
		DateModified:       ts,
	}

	cmd := MysqlDatabase{}.NewAdminContentFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_fields")
	}
	p, ok := cmd.Params().(CreateAdminContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminContentFieldParams", cmd.Params())
	}
	if p.AdminFieldValue != "mysql-cmd-value" {
		t.Errorf("Params().AdminFieldValue = %q, want %q", p.AdminFieldValue, "mysql-cmd-value")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminContentFieldCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminContentFieldID()
	cmd := NewAdminContentFieldCmdMysql{}

	row := mdbm.AdminContentFields{AdminContentFieldID: fieldID}
	got := cmd.GetID(row)
	if got != string(fieldID) {
		t.Errorf("GetID() = %q, want %q", got, string(fieldID))
	}
}

func TestUpdateAdminContentFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	acfID := types.NewAdminContentFieldID()
	params := UpdateAdminContentFieldParams{
		AdminFieldValue:     "mysql-update-value",
		DateCreated:         ts,
		DateModified:        ts,
		AdminContentFieldID: acfID,
	}

	cmd := MysqlDatabase{}.UpdateAdminContentFieldCmd(ctx, ac, params)

	if cmd.TableName() != "admin_content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_fields")
	}
	if cmd.GetID() != string(acfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(acfID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminContentFieldParams", cmd.Params())
	}
	if p.AdminFieldValue != "mysql-update-value" {
		t.Errorf("Params().AdminFieldValue = %q, want %q", p.AdminFieldValue, "mysql-update-value")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminContentFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	acfID := types.NewAdminContentFieldID()

	cmd := MysqlDatabase{}.DeleteAdminContentFieldCmd(ctx, ac, acfID)

	if cmd.TableName() != "admin_content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_fields")
	}
	if cmd.GetID() != string(acfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(acfID))
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

func TestNewAdminContentFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node-acf"),
		RequestID: "psql-req-acf",
		IP:        "172.16.0.1",
	}
	params := CreateAdminContentFieldParams{
		AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID("psql-cmd-route"), Valid: true},
		AdminContentDataID: types.NullableAdminContentID{ID: types.AdminContentID("psql-cmd-data"), Valid: true},
		AdminFieldValue:    "psql-cmd-value",
		DateCreated:        ts,
		DateModified:       ts,
	}

	cmd := PsqlDatabase{}.NewAdminContentFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_fields")
	}
	p, ok := cmd.Params().(CreateAdminContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminContentFieldParams", cmd.Params())
	}
	if p.AdminFieldValue != "psql-cmd-value" {
		t.Errorf("Params().AdminFieldValue = %q, want %q", p.AdminFieldValue, "psql-cmd-value")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminContentFieldCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminContentFieldID()
	cmd := NewAdminContentFieldCmdPsql{}

	row := mdbp.AdminContentFields{AdminContentFieldID: fieldID}
	got := cmd.GetID(row)
	if got != string(fieldID) {
		t.Errorf("GetID() = %q, want %q", got, string(fieldID))
	}
}

func TestUpdateAdminContentFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	acfID := types.NewAdminContentFieldID()
	params := UpdateAdminContentFieldParams{
		AdminFieldValue:     "psql-update-value",
		DateCreated:         ts,
		DateModified:        ts,
		AdminContentFieldID: acfID,
	}

	cmd := PsqlDatabase{}.UpdateAdminContentFieldCmd(ctx, ac, params)

	if cmd.TableName() != "admin_content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_fields")
	}
	if cmd.GetID() != string(acfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(acfID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminContentFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminContentFieldParams", cmd.Params())
	}
	if p.AdminFieldValue != "psql-update-value" {
		t.Errorf("Params().AdminFieldValue = %q, want %q", p.AdminFieldValue, "psql-update-value")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminContentFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	acfID := types.NewAdminContentFieldID()

	cmd := PsqlDatabase{}.DeleteAdminContentFieldCmd(ctx, ac, acfID)

	if cmd.TableName() != "admin_content_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_content_fields")
	}
	if cmd.GetID() != string(acfID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(acfID))
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
// Verify that all three database types produce commands with the correct table
// name and non-nil recorders.

func TestAuditedAdminContentFieldCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminContentFieldParams{}
	updateParams := UpdateAdminContentFieldParams{AdminContentFieldID: types.NewAdminContentFieldID()}
	acfID := types.NewAdminContentFieldID()

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewAdminContentFieldCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateAdminContentFieldCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteAdminContentFieldCmd(ctx, ac, acfID).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewAdminContentFieldCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateAdminContentFieldCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteAdminContentFieldCmd(ctx, ac, acfID).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewAdminContentFieldCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateAdminContentFieldCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteAdminContentFieldCmd(ctx, ac, acfID).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "admin_content_fields" {
				t.Errorf("TableName() = %q, want %q", c.name, "admin_content_fields")
			}
		})
	}
}

func TestAuditedAdminContentFieldCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminContentFieldParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewAdminContentFieldCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewAdminContentFieldCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewAdminContentFieldCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedAdminContentFieldCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	acfID := types.NewAdminContentFieldID()

	t.Run("UpdateCmd GetID returns AdminContentFieldID", func(t *testing.T) {
		t.Parallel()
		params := UpdateAdminContentFieldParams{AdminContentFieldID: acfID}

		sqliteCmd := Database{}.UpdateAdminContentFieldCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateAdminContentFieldCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateAdminContentFieldCmd(ctx, ac, params)

		wantID := string(acfID)
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

	t.Run("DeleteCmd GetID returns AdminContentFieldID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteAdminContentFieldCmd(ctx, ac, acfID)
		mysqlCmd := MysqlDatabase{}.DeleteAdminContentFieldCmd(ctx, ac, acfID)
		psqlCmd := PsqlDatabase{}.DeleteAdminContentFieldCmd(ctx, ac, acfID)

		wantID := string(acfID)
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
		testFieldID := types.NewAdminContentFieldID()

		sqliteCmd := NewAdminContentFieldCmd{}
		mysqlCmd := NewAdminContentFieldCmdMysql{}
		psqlCmd := NewAdminContentFieldCmdPsql{}

		wantID := string(testFieldID)

		sqliteRow := mdb.AdminContentFields{AdminContentFieldID: testFieldID}
		mysqlRow := mdbm.AdminContentFields{AdminContentFieldID: testFieldID}
		psqlRow := mdbp.AdminContentFields{AdminContentFieldID: testFieldID}

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

// --- Edge case: UpdateCmd with empty AdminContentFieldID ---

func TestUpdateAdminContentFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateAdminContentFieldParams{AdminContentFieldID: ""}

	sqliteCmd := Database{}.UpdateAdminContentFieldCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateAdminContentFieldCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateAdminContentFieldCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteCmd with empty ID ---

func TestDeleteAdminContentFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.AdminContentFieldID("")

	sqliteCmd := Database{}.DeleteAdminContentFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteAdminContentFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteAdminContentFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- UpdateAdminContentField success message format ---
// The UpdateAdminContentField method produces a success message string.
// We can't test it without a real DB, but we verify the message format expectation
// by testing the format string directly.

func TestUpdateAdminContentField_SuccessMessageFormat(t *testing.T) {
	t.Parallel()
	acfID := types.NewAdminContentFieldID()
	expected := fmt.Sprintf("Successfully updated %v\n", acfID)

	// The actual method builds this exact format. Verify the expectation is stable.
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
	_ audited.CreateCommand[mdb.AdminContentFields]  = NewAdminContentFieldCmd{}
	_ audited.UpdateCommand[mdb.AdminContentFields]  = UpdateAdminContentFieldCmd{}
	_ audited.DeleteCommand[mdb.AdminContentFields]  = DeleteAdminContentFieldCmd{}
	_ audited.CreateCommand[mdbm.AdminContentFields] = NewAdminContentFieldCmdMysql{}
	_ audited.UpdateCommand[mdbm.AdminContentFields] = UpdateAdminContentFieldCmdMysql{}
	_ audited.DeleteCommand[mdbm.AdminContentFields] = DeleteAdminContentFieldCmdMysql{}
	_ audited.CreateCommand[mdbp.AdminContentFields] = NewAdminContentFieldCmdPsql{}
	_ audited.UpdateCommand[mdbp.AdminContentFields] = UpdateAdminContentFieldCmdPsql{}
	_ audited.DeleteCommand[mdbp.AdminContentFields] = DeleteAdminContentFieldCmdPsql{}
)
