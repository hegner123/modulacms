package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// adminDatatypeFixture returns a fully populated AdminDatatypes and its parts for testing.
func adminDatatypeFixture() (AdminDatatypes, types.AdminDatatypeID, types.Timestamp) {
	datatypeID := types.NewAdminDatatypeID()
	ts := types.NewTimestamp(time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC))
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	adt := AdminDatatypes{
		AdminDatatypeID: datatypeID,
		ParentID:        parentID,
		Label:           "Blog Post",
		Type:            "collection",
		AuthorID:        authorID,
		DateCreated:     ts,
		DateModified:    ts,
	}
	return adt, datatypeID, ts
}

// adminDatatypeFixtureNulls returns an AdminDatatypes where all nullable fields are null/invalid.
func adminDatatypeFixtureNulls() AdminDatatypes {
	datatypeID := types.NewAdminDatatypeID()
	ts := types.NewTimestamp(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))

	return AdminDatatypes{
		AdminDatatypeID: datatypeID,
		ParentID:        types.NullableAdminDatatypeID{Valid: false},
		Label:           "Minimal",
		Type:            "single",
		AuthorID:        types.NullableUserID{Valid: false},
		DateCreated:     ts,
		DateModified:    ts,
	}
}

// --- MapAdminDatatypeJSON tests ---

func TestMapAdminDatatypeJSON_AllFieldsValid(t *testing.T) {
	t.Parallel()
	adt, datatypeID, ts := adminDatatypeFixture()

	got := MapAdminDatatypeJSON(adt)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"DatatypeID", got.DatatypeID, datatypeID.String()},
		{"ParentID", got.ParentID, adt.ParentID.String()},
		{"Label", got.Label, "Blog Post"},
		{"Type", got.Type, "collection"},
		{"AuthorID", got.AuthorID, adt.AuthorID.String()},
		{"DateCreated", got.DateCreated, ts.String()},
		{"DateModified", got.DateModified, ts.String()},
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

func TestMapAdminDatatypeJSON_NullableIDsProduceNullString(t *testing.T) {
	t.Parallel()
	// When ParentID and AuthorID are null (Valid=false),
	// their String() method returns "null".
	adt := adminDatatypeFixtureNulls()

	got := MapAdminDatatypeJSON(adt)

	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q for null NullableAdminDatatypeID", got.ParentID, "null")
	}
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q for null NullableUserID", got.AuthorID, "null")
	}
}

func TestMapAdminDatatypeJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value AdminDatatypes should not panic
	got := MapAdminDatatypeJSON(AdminDatatypes{})

	if got.DatatypeID != "" {
		t.Errorf("DatatypeID = %q, want empty string for zero AdminDatatypeID", got.DatatypeID)
	}
	// NullableXxx with zero value has Valid=false, so String() returns "null"
	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q", got.ParentID, "null")
	}
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, "null")
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
}

func TestMapAdminDatatypeJSON_MapsAdminIDToPublicShape(t *testing.T) {
	t.Parallel()
	// Verify that AdminDatatypeID is mapped into DatatypeJSON.DatatypeID.
	// This is the core purpose of this function: admin-specific ID -> public shape.
	adt, _, _ := adminDatatypeFixture()
	got := MapAdminDatatypeJSON(adt)

	if got.DatatypeID != adt.AdminDatatypeID.String() {
		t.Errorf("DatatypeID = %q, want AdminDatatypeID %q", got.DatatypeID, adt.AdminDatatypeID.String())
	}
}

// --- MapStringAdminDatatype tests ---

func TestMapStringAdminDatatype_AllFields(t *testing.T) {
	t.Parallel()
	adt, datatypeID, ts := adminDatatypeFixture()

	got := MapStringAdminDatatype(adt)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"AdminDatatypeID", got.AdminDatatypeID, datatypeID.String()},
		{"ParentID", got.ParentID, adt.ParentID.String()},
		{"Label", got.Label, "Blog Post"},
		{"Type", got.Type, "collection"},
		{"AuthorID", got.AuthorID, adt.AuthorID.String()},
		{"DateCreated", got.DateCreated, ts.String()},
		{"DateModified", got.DateModified, ts.String()},
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

func TestMapStringAdminDatatype_HistoryAlwaysEmpty(t *testing.T) {
	t.Parallel()
	// History field is hard-coded to "" per the comment in the source.
	adt, _, _ := adminDatatypeFixture()

	got := MapStringAdminDatatype(adt)

	if got.History != "" {
		t.Errorf("History = %q, want empty string (History field removed)", got.History)
	}
}

func TestMapStringAdminDatatype_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringAdminDatatype(AdminDatatypes{})

	if got.AdminDatatypeID != "" {
		t.Errorf("AdminDatatypeID = %q, want empty string", got.AdminDatatypeID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
}

func TestMapStringAdminDatatype_NullableFieldsShowNull(t *testing.T) {
	t.Parallel()
	adt := adminDatatypeFixtureNulls()

	got := MapStringAdminDatatype(adt)

	// Nullable IDs with Valid=false produce "null" via their String() method
	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q for null", got.ParentID, "null")
	}
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q for null", got.AuthorID, "null")
	}
}

// --- SQLite Database.MapAdminDatatype tests ---

func TestDatabase_MapAdminDatatype_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	datatypeID := types.NewAdminDatatypeID()
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdb.AdminDatatypes{
		AdminDatatypeID: datatypeID,
		ParentID:        parentID,
		Label:           "Article",
		Type:            "collection",
		AuthorID:        authorID,
		DateCreated:     ts,
		DateModified:    ts,
	}

	got := d.MapAdminDatatype(input)

	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "Article" {
		t.Errorf("Label = %q, want %q", got.Label, "Article")
	}
	if got.Type != "collection" {
		t.Errorf("Type = %q, want %q", got.Type, "collection")
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

func TestDatabase_MapAdminDatatype_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminDatatype(mdb.AdminDatatypes{})

	if got.AdminDatatypeID != "" {
		t.Errorf("AdminDatatypeID = %v, want zero value", got.AdminDatatypeID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
	if got.ParentID.Valid {
		t.Errorf("ParentID.Valid = true, want false for zero value")
	}
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false for zero value")
	}
}

// --- SQLite Database.MapCreateAdminDatatypeParams tests ---

func TestDatabase_MapCreateAdminDatatypeParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminDatatypeParams{
		ParentID:     types.NullableAdminDatatypeID{Valid: false},
		Label:        "Page",
		Type:         "single",
		AuthorID:     types.NullableUserID{ID: types.NewUserID(), Valid: true},
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminDatatypeParams(input)

	if got.AdminDatatypeID.IsZero() {
		t.Fatal("expected non-zero AdminDatatypeID to be generated")
	}
	if got.Label != "Page" {
		t.Errorf("Label = %q, want %q", got.Label, "Page")
	}
	if got.Type != "single" {
		t.Errorf("Type = %q, want %q", got.Type, "single")
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

func TestDatabase_MapCreateAdminDatatypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateAdminDatatypeParams{Label: "test"}

	got1 := d.MapCreateAdminDatatypeParams(input)
	got2 := d.MapCreateAdminDatatypeParams(input)
	if got1.AdminDatatypeID == got2.AdminDatatypeID {
		t.Error("two calls produced the same AdminDatatypeID -- each call should generate a unique ID")
	}
}

func TestDatabase_MapCreateAdminDatatypeParams_PreservesNullableFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	parentID := types.NullableAdminDatatypeID{ID: types.AdminDatatypeID("parent-123"), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := CreateAdminDatatypeParams{
		ParentID: parentID,
		AuthorID: authorID,
		Label:    "Test",
		Type:     "collection",
	}

	got := d.MapCreateAdminDatatypeParams(input)

	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

// --- SQLite Database.MapUpdateAdminDatatypeParams tests ---

func TestDatabase_MapUpdateAdminDatatypeParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	datatypeID := types.NewAdminDatatypeID()
	parentID := types.NullableAdminDatatypeID{ID: types.AdminDatatypeID("parent-updated"), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := UpdateAdminDatatypeParams{
		ParentID:        parentID,
		Label:           "Updated Label",
		Type:            "single",
		AuthorID:        authorID,
		DateCreated:     ts,
		DateModified:    ts,
		AdminDatatypeID: datatypeID,
	}

	got := d.MapUpdateAdminDatatypeParams(input)

	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "Updated Label" {
		t.Errorf("Label = %q, want %q", got.Label, "Updated Label")
	}
	if got.Type != "single" {
		t.Errorf("Type = %q, want %q", got.Type, "single")
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

// --- MySQL MysqlDatabase.MapAdminDatatype tests ---

func TestMysqlDatabase_MapAdminDatatype_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	datatypeID := types.NewAdminDatatypeID()
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbm.AdminDatatypes{
		AdminDatatypeID: datatypeID,
		ParentID:        parentID,
		Label:           "MySQL Article",
		Type:            "collection",
		AuthorID:        authorID,
		DateCreated:     ts,
		DateModified:    ts,
	}

	got := d.MapAdminDatatype(input)

	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "MySQL Article" {
		t.Errorf("Label = %q, want %q", got.Label, "MySQL Article")
	}
	if got.Type != "collection" {
		t.Errorf("Type = %q, want %q", got.Type, "collection")
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

func TestMysqlDatabase_MapAdminDatatype_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminDatatype(mdbm.AdminDatatypes{})

	if got.AdminDatatypeID != "" {
		t.Errorf("AdminDatatypeID = %v, want zero value", got.AdminDatatypeID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
}

// --- MySQL MysqlDatabase.MapCreateAdminDatatypeParams tests ---

func TestMysqlDatabase_MapCreateAdminDatatypeParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminDatatypeParams{
		Label:        "MySQL Page",
		Type:         "single",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminDatatypeParams(input)

	if got.AdminDatatypeID.IsZero() {
		t.Fatal("expected non-zero AdminDatatypeID to be generated")
	}
	if got.Label != "MySQL Page" {
		t.Errorf("Label = %q, want %q", got.Label, "MySQL Page")
	}
	if got.Type != "single" {
		t.Errorf("Type = %q, want %q", got.Type, "single")
	}
}

// --- MySQL MysqlDatabase.MapUpdateAdminDatatypeParams tests ---

func TestMysqlDatabase_MapUpdateAdminDatatypeParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	datatypeID := types.NewAdminDatatypeID()

	input := UpdateAdminDatatypeParams{
		ParentID:        types.NullableAdminDatatypeID{Valid: false},
		Label:           "MySQL Updated",
		Type:            "collection",
		AuthorID:        types.NullableUserID{Valid: false},
		DateCreated:     ts,
		DateModified:    ts,
		AdminDatatypeID: datatypeID,
	}

	got := d.MapUpdateAdminDatatypeParams(input)

	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.Label != "MySQL Updated" {
		t.Errorf("Label = %q, want %q", got.Label, "MySQL Updated")
	}
	if got.Type != "collection" {
		t.Errorf("Type = %q, want %q", got.Type, "collection")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

// --- PostgreSQL PsqlDatabase.MapAdminDatatype tests ---

func TestPsqlDatabase_MapAdminDatatype_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	datatypeID := types.NewAdminDatatypeID()
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbp.AdminDatatypes{
		AdminDatatypeID: datatypeID,
		ParentID:        parentID,
		Label:           "Psql Article",
		Type:            "collection",
		AuthorID:        authorID,
		DateCreated:     ts,
		DateModified:    ts,
	}

	got := d.MapAdminDatatype(input)

	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "Psql Article" {
		t.Errorf("Label = %q, want %q", got.Label, "Psql Article")
	}
	if got.Type != "collection" {
		t.Errorf("Type = %q, want %q", got.Type, "collection")
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

func TestPsqlDatabase_MapAdminDatatype_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminDatatype(mdbp.AdminDatatypes{})

	if got.AdminDatatypeID != "" {
		t.Errorf("AdminDatatypeID = %v, want zero value", got.AdminDatatypeID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateAdminDatatypeParams tests ---

func TestPsqlDatabase_MapCreateAdminDatatypeParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminDatatypeParams{
		Label:        "Psql Page",
		Type:         "single",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminDatatypeParams(input)

	if got.AdminDatatypeID.IsZero() {
		t.Fatal("expected non-zero AdminDatatypeID to be generated")
	}
	if got.Label != "Psql Page" {
		t.Errorf("Label = %q, want %q", got.Label, "Psql Page")
	}
	if got.Type != "single" {
		t.Errorf("Type = %q, want %q", got.Type, "single")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateAdminDatatypeParams tests ---

func TestPsqlDatabase_MapUpdateAdminDatatypeParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	datatypeID := types.NewAdminDatatypeID()

	input := UpdateAdminDatatypeParams{
		ParentID:        types.NullableAdminDatatypeID{Valid: false},
		Label:           "Psql Updated",
		Type:            "collection",
		AuthorID:        types.NullableUserID{Valid: false},
		DateCreated:     ts,
		DateModified:    ts,
		AdminDatatypeID: datatypeID,
	}

	got := d.MapUpdateAdminDatatypeParams(input)

	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.Label != "Psql Updated" {
		t.Errorf("Label = %q, want %q", got.Label, "Psql Updated")
	}
	if got.Type != "collection" {
		t.Errorf("Type = %q, want %q", got.Type, "collection")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

// --- Cross-database mapper consistency ---
// Verifies that all three database mappers produce identical AdminDatatypes from equivalent input.

func TestCrossDatabaseMapAdminDatatype_Consistency(t *testing.T) {
	t.Parallel()
	datatypeID := types.NewAdminDatatypeID()
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	parentID := types.NullableAdminDatatypeID{ID: types.AdminDatatypeID("parent-cross"), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	sqliteInput := mdb.AdminDatatypes{
		AdminDatatypeID: datatypeID, ParentID: parentID,
		Label: "Cross Test", Type: "collection",
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.AdminDatatypes{
		AdminDatatypeID: datatypeID, ParentID: parentID,
		Label: "Cross Test", Type: "collection",
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.AdminDatatypes{
		AdminDatatypeID: datatypeID, ParentID: parentID,
		Label: "Cross Test", Type: "collection",
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapAdminDatatype(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminDatatype(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminDatatype(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateAdminDatatypeParams consistency ---
// All three should auto-generate unique IDs.

func TestCrossDatabaseMapCreateAdminDatatypeParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminDatatypeParams{
		Label:        "Cross Create",
		Type:         "single",
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteResult := Database{}.MapCreateAdminDatatypeParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateAdminDatatypeParams(input)
	psqlResult := PsqlDatabase{}.MapCreateAdminDatatypeParams(input)

	if sqliteResult.AdminDatatypeID.IsZero() {
		t.Error("SQLite: expected non-zero generated AdminDatatypeID")
	}
	if mysqlResult.AdminDatatypeID.IsZero() {
		t.Error("MySQL: expected non-zero generated AdminDatatypeID")
	}
	if psqlResult.AdminDatatypeID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated AdminDatatypeID")
	}

	// Each call should generate a unique ID
	if sqliteResult.AdminDatatypeID == mysqlResult.AdminDatatypeID {
		t.Error("SQLite and MySQL generated the same AdminDatatypeID -- each call should be unique")
	}
	if sqliteResult.AdminDatatypeID == psqlResult.AdminDatatypeID {
		t.Error("SQLite and PostgreSQL generated the same AdminDatatypeID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewAdminDatatypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-123",
		IP:        "10.0.0.1",
	}
	params := CreateAdminDatatypeParams{
		Label:        "Cmd Test",
		Type:         "collection",
		AuthorID:     types.NullableUserID{ID: userID, Valid: true},
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := Database{}.NewAdminDatatypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes")
	}
	p, ok := cmd.Params().(CreateAdminDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminDatatypeParams", cmd.Params())
	}
	if p.Label != "Cmd Test" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "Cmd Test")
	}
	if p.Type != "collection" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "collection")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminDatatypeCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	datatypeID := types.NewAdminDatatypeID()
	cmd := NewAdminDatatypeCmd{}

	row := mdb.AdminDatatypes{AdminDatatypeID: datatypeID}
	got := cmd.GetID(row)
	if got != string(datatypeID) {
		t.Errorf("GetID() = %q, want %q", got, string(datatypeID))
	}
}

func TestNewAdminDatatypeCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewAdminDatatypeCmd{}

	row := mdb.AdminDatatypes{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateAdminDatatypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	datatypeID := types.NewAdminDatatypeID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateAdminDatatypeParams{
		Label:           "Updated",
		Type:            "single",
		DateCreated:     ts,
		DateModified:    ts,
		AdminDatatypeID: datatypeID,
	}

	cmd := Database{}.UpdateAdminDatatypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes")
	}
	if cmd.GetID() != string(datatypeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(datatypeID))
	}
	p, ok := cmd.Params().(UpdateAdminDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminDatatypeParams", cmd.Params())
	}
	if p.Label != "Updated" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "Updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminDatatypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	datatypeID := types.NewAdminDatatypeID()

	cmd := Database{}.DeleteAdminDatatypeCmd(ctx, ac, datatypeID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes")
	}
	if cmd.GetID() != string(datatypeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(datatypeID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewAdminDatatypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-req",
		IP:        "192.168.1.1",
	}
	params := CreateAdminDatatypeParams{
		Label:        "MySQL Cmd",
		Type:         "collection",
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := MysqlDatabase{}.NewAdminDatatypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes")
	}
	p, ok := cmd.Params().(CreateAdminDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminDatatypeParams", cmd.Params())
	}
	if p.Label != "MySQL Cmd" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "MySQL Cmd")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminDatatypeCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	datatypeID := types.NewAdminDatatypeID()
	cmd := NewAdminDatatypeCmdMysql{}

	row := mdbm.AdminDatatypes{AdminDatatypeID: datatypeID}
	got := cmd.GetID(row)
	if got != string(datatypeID) {
		t.Errorf("GetID() = %q, want %q", got, string(datatypeID))
	}
}

func TestUpdateAdminDatatypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	datatypeID := types.NewAdminDatatypeID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateAdminDatatypeParams{
		Label:           "MySQL Updated",
		Type:            "single",
		DateCreated:     ts,
		DateModified:    ts,
		AdminDatatypeID: datatypeID,
	}

	cmd := MysqlDatabase{}.UpdateAdminDatatypeCmd(ctx, ac, params)

	if cmd.TableName() != "admin_datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes")
	}
	if cmd.GetID() != string(datatypeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(datatypeID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminDatatypeParams", cmd.Params())
	}
	if p.Label != "MySQL Updated" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "MySQL Updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminDatatypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	datatypeID := types.NewAdminDatatypeID()

	cmd := MysqlDatabase{}.DeleteAdminDatatypeCmd(ctx, ac, datatypeID)

	if cmd.TableName() != "admin_datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes")
	}
	if cmd.GetID() != string(datatypeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(datatypeID))
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

func TestNewAdminDatatypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-req",
		IP:        "172.16.0.1",
	}
	params := CreateAdminDatatypeParams{
		Label:        "Psql Cmd",
		Type:         "collection",
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := PsqlDatabase{}.NewAdminDatatypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes")
	}
	p, ok := cmd.Params().(CreateAdminDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminDatatypeParams", cmd.Params())
	}
	if p.Label != "Psql Cmd" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "Psql Cmd")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminDatatypeCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	datatypeID := types.NewAdminDatatypeID()
	cmd := NewAdminDatatypeCmdPsql{}

	row := mdbp.AdminDatatypes{AdminDatatypeID: datatypeID}
	got := cmd.GetID(row)
	if got != string(datatypeID) {
		t.Errorf("GetID() = %q, want %q", got, string(datatypeID))
	}
}

func TestUpdateAdminDatatypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	datatypeID := types.NewAdminDatatypeID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateAdminDatatypeParams{
		Label:           "Psql Updated",
		Type:            "single",
		DateCreated:     ts,
		DateModified:    ts,
		AdminDatatypeID: datatypeID,
	}

	cmd := PsqlDatabase{}.UpdateAdminDatatypeCmd(ctx, ac, params)

	if cmd.TableName() != "admin_datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes")
	}
	if cmd.GetID() != string(datatypeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(datatypeID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminDatatypeParams", cmd.Params())
	}
	if p.Label != "Psql Updated" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "Psql Updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminDatatypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	datatypeID := types.NewAdminDatatypeID()

	cmd := PsqlDatabase{}.DeleteAdminDatatypeCmd(ctx, ac, datatypeID)

	if cmd.TableName() != "admin_datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes")
	}
	if cmd.GetID() != string(datatypeID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(datatypeID))
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

func TestAuditedAdminDatatypeCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateAdminDatatypeParams{}
	updateParams := UpdateAdminDatatypeParams{AdminDatatypeID: types.NewAdminDatatypeID()}
	datatypeID := types.NewAdminDatatypeID()

	// SQLite
	sqliteCreate := Database{}.NewAdminDatatypeCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateAdminDatatypeCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteAdminDatatypeCmd(ctx, ac, datatypeID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewAdminDatatypeCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateAdminDatatypeCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteAdminDatatypeCmd(ctx, ac, datatypeID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewAdminDatatypeCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateAdminDatatypeCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteAdminDatatypeCmd(ctx, ac, datatypeID)

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
			if c.name != "admin_datatypes" {
				t.Errorf("TableName() = %q, want %q", c.name, "admin_datatypes")
			}
		})
	}
}

func TestAuditedAdminDatatypeCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminDatatypeParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewAdminDatatypeCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewAdminDatatypeCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewAdminDatatypeCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedAdminDatatypeCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	datatypeID := types.NewAdminDatatypeID()

	t.Run("UpdateCmd GetID returns AdminDatatypeID", func(t *testing.T) {
		t.Parallel()
		params := UpdateAdminDatatypeParams{AdminDatatypeID: datatypeID}

		sqliteCmd := Database{}.UpdateAdminDatatypeCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateAdminDatatypeCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateAdminDatatypeCmd(ctx, ac, params)

		wantID := string(datatypeID)
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

	t.Run("DeleteCmd GetID returns AdminDatatypeID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteAdminDatatypeCmd(ctx, ac, datatypeID)
		mysqlCmd := MysqlDatabase{}.DeleteAdminDatatypeCmd(ctx, ac, datatypeID)
		psqlCmd := PsqlDatabase{}.DeleteAdminDatatypeCmd(ctx, ac, datatypeID)

		wantID := string(datatypeID)
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
		testDatatypeID := types.NewAdminDatatypeID()

		sqliteCmd := NewAdminDatatypeCmd{}
		mysqlCmd := NewAdminDatatypeCmdMysql{}
		psqlCmd := NewAdminDatatypeCmdPsql{}

		wantID := string(testDatatypeID)

		sqliteRow := mdb.AdminDatatypes{AdminDatatypeID: testDatatypeID}
		mysqlRow := mdbm.AdminDatatypes{AdminDatatypeID: testDatatypeID}
		psqlRow := mdbp.AdminDatatypes{AdminDatatypeID: testDatatypeID}

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

// --- Edge case: UpdateAdminDatatypeCmd with empty AdminDatatypeID ---

func TestUpdateAdminDatatypeCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateAdminDatatypeParams{AdminDatatypeID: ""}

	sqliteCmd := Database{}.UpdateAdminDatatypeCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateAdminDatatypeCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateAdminDatatypeCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteAdminDatatypeCmd with empty ID ---

func TestDeleteAdminDatatypeCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.AdminDatatypeID("")

	sqliteCmd := Database{}.DeleteAdminDatatypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteAdminDatatypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteAdminDatatypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.AdminDatatypes]  = NewAdminDatatypeCmd{}
	_ audited.UpdateCommand[mdb.AdminDatatypes]  = UpdateAdminDatatypeCmd{}
	_ audited.DeleteCommand[mdb.AdminDatatypes]  = DeleteAdminDatatypeCmd{}
	_ audited.CreateCommand[mdbm.AdminDatatypes] = NewAdminDatatypeCmdMysql{}
	_ audited.UpdateCommand[mdbm.AdminDatatypes] = UpdateAdminDatatypeCmdMysql{}
	_ audited.DeleteCommand[mdbm.AdminDatatypes] = DeleteAdminDatatypeCmdMysql{}
	_ audited.CreateCommand[mdbp.AdminDatatypes] = NewAdminDatatypeCmdPsql{}
	_ audited.UpdateCommand[mdbp.AdminDatatypes] = UpdateAdminDatatypeCmdPsql{}
	_ audited.DeleteCommand[mdbp.AdminDatatypes] = DeleteAdminDatatypeCmdPsql{}
)

// --- Struct field correctness ---
// Verify that the wrapper AdminDatatypes struct and param structs hold values correctly via JSON.

func TestAdminDatatypesStruct_JSONTags(t *testing.T) {
	t.Parallel()
	adt, _, _ := adminDatatypeFixture()

	data, err := json.Marshal(adt)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"admin_datatype_id", "parent_id", "label",
		"type", "author_id", "date_created", "date_modified",
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

func TestCreateAdminDatatypeParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	p := CreateAdminDatatypeParams{
		Label:        "JSON Test",
		Type:         "collection",
		DateCreated:  ts,
		DateModified: ts,
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
		"parent_id", "label", "type",
		"author_id", "date_created", "date_modified",
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

func TestUpdateAdminDatatypeParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	datatypeID := types.NewAdminDatatypeID()
	p := UpdateAdminDatatypeParams{
		Label:           "JSON Update Test",
		Type:            "single",
		DateCreated:     ts,
		DateModified:    ts,
		AdminDatatypeID: datatypeID,
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
		"parent_id", "label", "type",
		"author_id", "date_created", "date_modified",
		"admin_datatype_id",
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

// --- Additional struct JSON tag tests ---

func TestListAdminDatatypeByRouteIdRow_JSONTags(t *testing.T) {
	t.Parallel()
	row := ListAdminDatatypeByRouteIdRow{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		AdminRouteID:    types.NullableRouteID{Valid: false},
		ParentID:        types.NullableAdminDatatypeID{Valid: false},
		Label:           "Route Row",
		Type:            "collection",
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
		"admin_datatype_id", "admin_route_id", "parent_id",
		"label", "type",
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

func TestUtilityGetAdminDatatypesRow_JSONTags(t *testing.T) {
	t.Parallel()
	row := UtilityGetAdminDatatypesRow{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		Label:           "Utility Row",
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
		"admin_datatype_id", "label",
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

// --- MapAdminDatatypeJSON: label and type edge cases ---

func TestMapAdminDatatypeJSON_SpecialCharacters(t *testing.T) {
	t.Parallel()
	// Labels and types with special characters should pass through unchanged.
	tests := []struct {
		name  string
		label string
		typ   string
	}{
		{"unicode", "Artikeltyp \u00e4\u00f6\u00fc", "Sammlung"},
		{"spaces", "My Blog Post Type", "my collection"},
		{"empty_strings", "", ""},
		{"special_chars", "Type <script>", "col/sub"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			adt := AdminDatatypes{
				AdminDatatypeID: types.NewAdminDatatypeID(),
				Label:           tt.label,
				Type:            tt.typ,
			}
			got := MapAdminDatatypeJSON(adt)
			if got.Label != tt.label {
				t.Errorf("Label = %q, want %q", got.Label, tt.label)
			}
			if got.Type != tt.typ {
				t.Errorf("Type = %q, want %q", got.Type, tt.typ)
			}
		})
	}
}

// --- MapStringAdminDatatype: timestamp formatting ---

func TestMapStringAdminDatatype_TimestampFormatting(t *testing.T) {
	t.Parallel()
	// Verify that different timestamps are preserved exactly via String() conversion.
	tests := []struct {
		name string
		ts   time.Time
	}{
		{"epoch", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"recent", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},
		{"midnight", time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := types.NewTimestamp(tt.ts)
			adt := AdminDatatypes{
				AdminDatatypeID: types.NewAdminDatatypeID(),
				Label:           "ts-test",
				DateCreated:     ts,
				DateModified:    ts,
			}
			got := MapStringAdminDatatype(adt)
			if got.DateCreated != ts.String() {
				t.Errorf("DateCreated = %q, want %q", got.DateCreated, ts.String())
			}
			if got.DateModified != ts.String() {
				t.Errorf("DateModified = %q, want %q", got.DateModified, ts.String())
			}
		})
	}
}
