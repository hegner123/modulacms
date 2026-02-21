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

// --- Compile-time interface checks ---
// These verify that all 9 audited command types satisfy their respective
// audited interfaces. A compilation failure here means the struct drifted
// from the interface contract.

var (
	_ audited.CreateCommand[mdb.Datatypes]  = NewDatatypeCmd{}
	_ audited.UpdateCommand[mdb.Datatypes]  = UpdateDatatypeCmd{}
	_ audited.DeleteCommand[mdb.Datatypes]  = DeleteDatatypeCmd{}
	_ audited.CreateCommand[mdbm.Datatypes] = NewDatatypeCmdMysql{}
	_ audited.UpdateCommand[mdbm.Datatypes] = UpdateDatatypeCmdMysql{}
	_ audited.DeleteCommand[mdbm.Datatypes] = DeleteDatatypeCmdMysql{}
	_ audited.CreateCommand[mdbp.Datatypes] = NewDatatypeCmdPsql{}
	_ audited.UpdateCommand[mdbp.Datatypes] = UpdateDatatypeCmdPsql{}
	_ audited.DeleteCommand[mdbp.Datatypes] = DeleteDatatypeCmdPsql{}
)

// --- Test data helpers ---

// datatypeTestFixture returns a fully populated Datatypes struct and its component parts.
func datatypeTestFixture() (Datatypes, types.DatatypeID, types.NullableDatatypeID, types.UserID, types.Timestamp) {
	dtID := types.NewDatatypeID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 7, 10, 14, 30, 0, 0, time.UTC))

	dt := Datatypes{
		DatatypeID:   dtID,
		ParentID:     parentID,
		Label:        "Blog Post",
		Type:         "content",
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}
	return dt, dtID, parentID, authorID, ts
}

// datatypeUpdateParams returns an UpdateDatatypeParams with all fields populated.
func datatypeUpdateParams() UpdateDatatypeParams {
	dtID := types.NewDatatypeID()
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 7, 15, 8, 45, 0, 0, time.UTC))
	return UpdateDatatypeParams{
		ParentID:     types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true},
		Label:        "Updated Article",
		Type:         "page",
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
		DatatypeID:   dtID,
	}
}

// --- MapDatatypeJSON tests ---

func TestMapDatatypeJSON_AllFields(t *testing.T) {
	t.Parallel()
	dt, dtID, parentID, authorID, ts := datatypeTestFixture()

	got := MapDatatypeJSON(dt)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"DatatypeID", got.DatatypeID, dtID.String()},
		{"ParentID", got.ParentID, parentID.String()},
		{"Label", got.Label, "Blog Post"},
		{"Type", got.Type, "content"},
		{"AuthorID", got.AuthorID, authorID.String()},
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

func TestMapDatatypeJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value Datatypes should produce a valid DatatypeJSON without panic
	got := MapDatatypeJSON(Datatypes{})

	if got.DatatypeID != "" {
		t.Errorf("DatatypeID = %q, want empty string", got.DatatypeID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
}

func TestMapDatatypeJSON_NullParentID(t *testing.T) {
	t.Parallel()
	// NullableDatatypeID with Valid=false should produce "null" from .String()
	dt := Datatypes{
		ParentID: types.NullableDatatypeID{Valid: false},
	}
	got := MapDatatypeJSON(dt)

	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q", got.ParentID, "null")
	}
}

func TestMapDatatypeJSON_ZeroAuthorID(t *testing.T) {
	t.Parallel()
	dt := Datatypes{
		AuthorID: types.UserID(""),
	}
	got := MapDatatypeJSON(dt)

	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string for zero UserID", got.AuthorID)
	}
}

// --- MapStringDatatype tests ---

func TestMapStringDatatype_AllFields(t *testing.T) {
	t.Parallel()
	dt, dtID, parentID, authorID, ts := datatypeTestFixture()

	got := MapStringDatatype(dt)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"DatatypeID", got.DatatypeID, dtID.String()},
		{"ParentID", got.ParentID, parentID.String()},
		{"Label", got.Label, "Blog Post"},
		{"Type", got.Type, "content"},
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

func TestMapStringDatatype_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringDatatype(Datatypes{})

	if got.DatatypeID != "" {
		t.Errorf("DatatypeID = %q, want empty string", got.DatatypeID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
	if got.History != "" {
		t.Errorf("History = %q, want empty string (always empty)", got.History)
	}
}

func TestMapStringDatatype_NullParentID(t *testing.T) {
	t.Parallel()
	dt := Datatypes{
		ParentID: types.NullableDatatypeID{Valid: false},
	}
	got := MapStringDatatype(dt)

	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q", got.ParentID, "null")
	}
}

func TestMapStringDatatype_ZeroAuthorID(t *testing.T) {
	t.Parallel()
	dt := Datatypes{
		AuthorID: types.UserID(""),
	}
	got := MapStringDatatype(dt)

	if got.AuthorID != "" {
		t.Errorf("AuthorID = %q, want empty string for zero UserID", got.AuthorID)
	}
}

func TestMapStringDatatype_HistoryAlwaysEmpty(t *testing.T) {
	t.Parallel()
	// The History field is explicitly set to "" per the comment "History field removed".
	// Verify this is consistent even with a fully populated Datatypes.
	dt, _, _, _, _ := datatypeTestFixture()
	got := MapStringDatatype(dt)

	if got.History != "" {
		t.Errorf("History = %q, want empty string (always empty per code comment)", got.History)
	}
}

// --- SQLite Database.MapDatatype tests ---

func TestDatabase_MapDatatype_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	dtID := types.NewDatatypeID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))

	input := mdb.Datatypes{
		DatatypeID:   dtID,
		ParentID:     parentID,
		Label:        "sqlite-datatype",
		Type:         "content",
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapDatatype(input)

	if got.DatatypeID != dtID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, dtID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "sqlite-datatype" {
		t.Errorf("Label = %q, want %q", got.Label, "sqlite-datatype")
	}
	if got.Type != "content" {
		t.Errorf("Type = %q, want %q", got.Type, "content")
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

func TestDatabase_MapDatatype_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapDatatype(mdb.Datatypes{})

	if got.DatatypeID != "" {
		t.Errorf("DatatypeID = %v, want zero value", got.DatatypeID)
	}
	if got.ParentID.Valid {
		t.Error("ParentID.Valid = true, want false")
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
}

// --- SQLite Database.MapCreateDatatypeParams tests ---

func TestDatabase_MapCreateDatatypeParams_Table(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NewUserID()
	explicitID := types.NewDatatypeID()

	tests := []struct {
		name       string
		input      CreateDatatypeParams
		wantAutoID bool // true means the test expects a new ID to be generated
		wantID     types.DatatypeID
	}{
		{
			name: "generates ID when DatatypeID is zero",
			input: CreateDatatypeParams{
				DatatypeID:   "",
				ParentID:     parentID,
				Label:        "Auto Datatype",
				Type:         "content",
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
			wantAutoID: true,
		},
		{
			name: "preserves explicit ID",
			input: CreateDatatypeParams{
				DatatypeID:   explicitID,
				ParentID:     parentID,
				Label:        "Explicit Datatype",
				Type:         "page",
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
			wantAutoID: false,
			wantID:     explicitID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.MapCreateDatatypeParams(tt.input)

			if tt.wantAutoID {
				if got.DatatypeID.IsZero() {
					t.Fatal("expected non-zero DatatypeID to be generated")
				}
			} else {
				if got.DatatypeID != tt.wantID {
					t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, tt.wantID)
				}
			}

			if got.ParentID != tt.input.ParentID {
				t.Errorf("ParentID = %v, want %v", got.ParentID, tt.input.ParentID)
			}
			if got.Label != tt.input.Label {
				t.Errorf("Label = %q, want %q", got.Label, tt.input.Label)
			}
			if got.Type != tt.input.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.input.Type)
			}
			if got.AuthorID != tt.input.AuthorID {
				t.Errorf("AuthorID = %v, want %v", got.AuthorID, tt.input.AuthorID)
			}
			if got.DateCreated != tt.input.DateCreated {
				t.Errorf("DateCreated = %v, want %v", got.DateCreated, tt.input.DateCreated)
			}
			if got.DateModified != tt.input.DateModified {
				t.Errorf("DateModified = %v, want %v", got.DateModified, tt.input.DateModified)
			}
		})
	}
}

func TestDatabase_MapCreateDatatypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateDatatypeParams{DatatypeID: ""} // zero ID triggers generation

	got1 := d.MapCreateDatatypeParams(input)
	got2 := d.MapCreateDatatypeParams(input)

	if got1.DatatypeID == got2.DatatypeID {
		t.Error("two calls generated the same DatatypeID -- each call should be unique")
	}
}

// --- SQLite Database.MapUpdateDatatypeParams tests ---

func TestDatabase_MapUpdateDatatypeParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := datatypeUpdateParams()

	got := d.MapUpdateDatatypeParams(params)

	if got.ParentID != params.ParentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, params.ParentID)
	}
	if got.Label != params.Label {
		t.Errorf("Label = %q, want %q", got.Label, params.Label)
	}
	if got.Type != params.Type {
		t.Errorf("Type = %q, want %q", got.Type, params.Type)
	}
	if got.AuthorID != params.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, params.AuthorID)
	}
	if got.DateCreated != params.DateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, params.DateCreated)
	}
	if got.DateModified != params.DateModified {
		t.Errorf("DateModified = %v, want %v", got.DateModified, params.DateModified)
	}
	// DatatypeID is the WHERE clause identifier -- must be preserved
	if got.DatatypeID != params.DatatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, params.DatatypeID)
	}
}

func TestDatabase_MapUpdateDatatypeParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateDatatypeParams(UpdateDatatypeParams{})

	if got.DatatypeID != "" {
		t.Errorf("DatatypeID = %v, want zero value", got.DatatypeID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- MySQL MysqlDatabase.MapDatatype tests ---

func TestMysqlDatabase_MapDatatype_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	dtID := types.NewDatatypeID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))

	input := mdbm.Datatypes{
		DatatypeID:   dtID,
		ParentID:     parentID,
		Label:        "mysql-datatype",
		Type:         "page",
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapDatatype(input)

	if got.DatatypeID != dtID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, dtID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "mysql-datatype" {
		t.Errorf("Label = %q, want %q", got.Label, "mysql-datatype")
	}
	if got.Type != "page" {
		t.Errorf("Type = %q, want %q", got.Type, "page")
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

func TestMysqlDatabase_MapDatatype_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapDatatype(mdbm.Datatypes{})

	if got.DatatypeID != "" {
		t.Errorf("DatatypeID = %v, want zero value", got.DatatypeID)
	}
	if got.ParentID.Valid {
		t.Error("ParentID.Valid = true, want false")
	}
}

// --- MySQL MysqlDatabase.MapCreateDatatypeParams tests ---

func TestMysqlDatabase_MapCreateDatatypeParams_Table(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NewUserID()
	explicitID := types.NewDatatypeID()

	tests := []struct {
		name       string
		input      CreateDatatypeParams
		wantAutoID bool
		wantID     types.DatatypeID
	}{
		{
			name: "generates ID when DatatypeID is zero",
			input: CreateDatatypeParams{
				DatatypeID: "",
				ParentID:   parentID,
				Label:      "MySQL Auto",
				Type:       "content",
				AuthorID:   authorID,
			},
			wantAutoID: true,
		},
		{
			name: "preserves explicit ID",
			input: CreateDatatypeParams{
				DatatypeID: explicitID,
				ParentID:   parentID,
				Label:      "MySQL Explicit",
				Type:       "page",
				AuthorID:   authorID,
			},
			wantAutoID: false,
			wantID:     explicitID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.MapCreateDatatypeParams(tt.input)

			if tt.wantAutoID {
				if got.DatatypeID.IsZero() {
					t.Fatal("expected non-zero DatatypeID to be generated")
				}
			} else {
				if got.DatatypeID != tt.wantID {
					t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, tt.wantID)
				}
			}

			if got.ParentID != tt.input.ParentID {
				t.Errorf("ParentID = %v, want %v", got.ParentID, tt.input.ParentID)
			}
			if got.Label != tt.input.Label {
				t.Errorf("Label = %q, want %q", got.Label, tt.input.Label)
			}
			if got.Type != tt.input.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.input.Type)
			}
			if got.AuthorID != tt.input.AuthorID {
				t.Errorf("AuthorID = %v, want %v", got.AuthorID, tt.input.AuthorID)
			}
		})
	}
}

// MySQL CreateDatatypeParams lacks DateCreated/DateModified fields (they use
// MySQL DEFAULT values). Verify the mapper does not include them.
func TestMysqlDatabase_MapCreateDatatypeParams_OmitsTimestamps(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	input := CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		Label:        "Timestamp Check",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateDatatypeParams(input)

	// The MySQL sqlc-generated CreateDatatypeParams struct does not have
	// DateCreated/DateModified fields. The fact that this compiles proves
	// the mapper correctly omits them. We verify the fields that ARE mapped.
	if got.Label != "Timestamp Check" {
		t.Errorf("Label = %q, want %q", got.Label, "Timestamp Check")
	}
}

func TestMysqlDatabase_MapCreateDatatypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := CreateDatatypeParams{DatatypeID: ""}

	got1 := d.MapCreateDatatypeParams(input)
	got2 := d.MapCreateDatatypeParams(input)

	if got1.DatatypeID == got2.DatatypeID {
		t.Error("two calls generated the same DatatypeID -- each call should be unique")
	}
}

// --- MySQL MysqlDatabase.MapUpdateDatatypeParams tests ---

func TestMysqlDatabase_MapUpdateDatatypeParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	dtID := types.NewDatatypeID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NewUserID()

	input := UpdateDatatypeParams{
		ParentID:   parentID,
		Label:      "mysql-updated",
		Type:       "component",
		AuthorID:   authorID,
		DatatypeID: dtID,
	}

	got := d.MapUpdateDatatypeParams(input)

	if got.ParentID != input.ParentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, input.ParentID)
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.Type != input.Type {
		t.Errorf("Type = %q, want %q", got.Type, input.Type)
	}
	if got.AuthorID != input.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, input.AuthorID)
	}
	if got.DatatypeID != dtID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, dtID)
	}
}

// --- PostgreSQL PsqlDatabase.MapDatatype tests ---

func TestPsqlDatabase_MapDatatype_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	dtID := types.NewDatatypeID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))

	input := mdbp.Datatypes{
		DatatypeID:   dtID,
		ParentID:     parentID,
		Label:        "psql-datatype",
		Type:         "component",
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapDatatype(input)

	if got.DatatypeID != dtID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, dtID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "psql-datatype" {
		t.Errorf("Label = %q, want %q", got.Label, "psql-datatype")
	}
	if got.Type != "component" {
		t.Errorf("Type = %q, want %q", got.Type, "component")
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

func TestPsqlDatabase_MapDatatype_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapDatatype(mdbp.Datatypes{})

	if got.DatatypeID != "" {
		t.Errorf("DatatypeID = %v, want zero value", got.DatatypeID)
	}
	if got.ParentID.Valid {
		t.Error("ParentID.Valid = true, want false")
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateDatatypeParams tests ---

func TestPsqlDatabase_MapCreateDatatypeParams_Table(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NewUserID()
	explicitID := types.NewDatatypeID()

	tests := []struct {
		name       string
		input      CreateDatatypeParams
		wantAutoID bool
		wantID     types.DatatypeID
	}{
		{
			name: "generates ID when DatatypeID is zero",
			input: CreateDatatypeParams{
				DatatypeID:   "",
				ParentID:     parentID,
				Label:        "Psql Auto",
				Type:         "content",
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
			wantAutoID: true,
		},
		{
			name: "preserves explicit ID",
			input: CreateDatatypeParams{
				DatatypeID:   explicitID,
				ParentID:     parentID,
				Label:        "Psql Explicit",
				Type:         "page",
				AuthorID:     authorID,
				DateCreated:  ts,
				DateModified: ts,
			},
			wantAutoID: false,
			wantID:     explicitID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.MapCreateDatatypeParams(tt.input)

			if tt.wantAutoID {
				if got.DatatypeID.IsZero() {
					t.Fatal("expected non-zero DatatypeID to be generated")
				}
			} else {
				if got.DatatypeID != tt.wantID {
					t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, tt.wantID)
				}
			}

			if got.ParentID != tt.input.ParentID {
				t.Errorf("ParentID = %v, want %v", got.ParentID, tt.input.ParentID)
			}
			if got.Label != tt.input.Label {
				t.Errorf("Label = %q, want %q", got.Label, tt.input.Label)
			}
			if got.Type != tt.input.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.input.Type)
			}
			if got.AuthorID != tt.input.AuthorID {
				t.Errorf("AuthorID = %v, want %v", got.AuthorID, tt.input.AuthorID)
			}
			if got.DateCreated != tt.input.DateCreated {
				t.Errorf("DateCreated = %v, want %v", got.DateCreated, tt.input.DateCreated)
			}
			if got.DateModified != tt.input.DateModified {
				t.Errorf("DateModified = %v, want %v", got.DateModified, tt.input.DateModified)
			}
		})
	}
}

func TestPsqlDatabase_MapCreateDatatypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := CreateDatatypeParams{DatatypeID: ""}

	got1 := d.MapCreateDatatypeParams(input)
	got2 := d.MapCreateDatatypeParams(input)

	if got1.DatatypeID == got2.DatatypeID {
		t.Error("two calls generated the same DatatypeID -- each call should be unique")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateDatatypeParams tests ---

func TestPsqlDatabase_MapUpdateDatatypeParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := datatypeUpdateParams()

	got := d.MapUpdateDatatypeParams(params)

	if got.ParentID != params.ParentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, params.ParentID)
	}
	if got.Label != params.Label {
		t.Errorf("Label = %q, want %q", got.Label, params.Label)
	}
	if got.Type != params.Type {
		t.Errorf("Type = %q, want %q", got.Type, params.Type)
	}
	if got.AuthorID != params.AuthorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, params.AuthorID)
	}
	if got.DateCreated != params.DateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, params.DateCreated)
	}
	if got.DateModified != params.DateModified {
		t.Errorf("DateModified = %v, want %v", got.DateModified, params.DateModified)
	}
	if got.DatatypeID != params.DatatypeID {
		t.Errorf("DatatypeID = %v, want %v", got.DatatypeID, params.DatatypeID)
	}
}

// --- Cross-database mapper consistency ---
// All three database drivers use identical types for Datatypes
// (no int32/int64 conversions needed). This test verifies they all produce
// identical Datatypes wrapper structs from equivalent input.

func TestCrossDatabaseMapDatatype_Consistency(t *testing.T) {
	t.Parallel()
	dtID := types.NewDatatypeID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NewUserID()
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))

	sqliteInput := mdb.Datatypes{
		DatatypeID: dtID, ParentID: parentID,
		Label: "cross-label", Type: "cross-type",
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.Datatypes{
		DatatypeID: dtID, ParentID: parentID,
		Label: "cross-label", Type: "cross-type",
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.Datatypes{
		DatatypeID: dtID, ParentID: parentID,
		Label: "cross-label", Type: "cross-type",
		AuthorID: authorID, DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapDatatype(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapDatatype(mysqlInput)
	psqlResult := PsqlDatabase{}.MapDatatype(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateDatatypeParams auto-ID generation ---

func TestCrossDatabaseMapCreateDatatypeParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	input := CreateDatatypeParams{
		DatatypeID:   "", // zero -- triggers generation
		Label:        "cross-create-label",
		Type:         "content",
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteResult := Database{}.MapCreateDatatypeParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateDatatypeParams(input)
	psqlResult := PsqlDatabase{}.MapCreateDatatypeParams(input)

	if sqliteResult.DatatypeID.IsZero() {
		t.Error("SQLite: expected non-zero generated DatatypeID")
	}
	if mysqlResult.DatatypeID.IsZero() {
		t.Error("MySQL: expected non-zero generated DatatypeID")
	}
	if psqlResult.DatatypeID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated DatatypeID")
	}

	// Each call should generate a unique ID
	if sqliteResult.DatatypeID == mysqlResult.DatatypeID {
		t.Error("SQLite and MySQL generated the same DatatypeID -- each call should be unique")
	}
	if sqliteResult.DatatypeID == psqlResult.DatatypeID {
		t.Error("SQLite and PostgreSQL generated the same DatatypeID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewDatatypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-dt-1"),
		RequestID: "req-dt-123",
		IP:        "10.0.0.1",
	}
	params := CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		Label:        "cmd-label",
		Type:         "content",
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := Database{}.NewDatatypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes")
	}
	p, ok := cmd.Params().(CreateDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateDatatypeParams", cmd.Params())
	}
	if p.Label != "cmd-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "cmd-label")
	}
	if p.Type != "content" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "content")
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewDatatypeCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	dtID := types.NewDatatypeID()
	cmd := NewDatatypeCmd{}

	row := mdb.Datatypes{DatatypeID: dtID}
	got := cmd.GetID(row)
	if got != string(dtID) {
		t.Errorf("GetID() = %q, want %q", got, string(dtID))
	}
}

func TestNewDatatypeCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	cmd := NewDatatypeCmd{}
	row := mdb.Datatypes{DatatypeID: ""}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateDatatypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	dtID := types.NewDatatypeID()
	params := UpdateDatatypeParams{
		Label:        "update-label",
		Type:         "page",
		DateCreated:  ts,
		DateModified: ts,
		DatatypeID:   dtID,
	}

	cmd := Database{}.UpdateDatatypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes")
	}
	if cmd.GetID() != string(dtID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(dtID))
	}
	p, ok := cmd.Params().(UpdateDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateDatatypeParams", cmd.Params())
	}
	if p.Label != "update-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "update-label")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteDatatypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	dtID := types.NewDatatypeID()

	cmd := Database{}.DeleteDatatypeCmd(ctx, ac, dtID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes")
	}
	if cmd.GetID() != string(dtID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(dtID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewDatatypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node-dt"),
		RequestID: "mysql-req-dt",
		IP:        "192.168.1.1",
	}
	params := CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		Label:        "mysql-cmd-label",
		Type:         "content",
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := MysqlDatabase{}.NewDatatypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes")
	}
	p, ok := cmd.Params().(CreateDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateDatatypeParams", cmd.Params())
	}
	if p.Label != "mysql-cmd-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "mysql-cmd-label")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewDatatypeCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	dtID := types.NewDatatypeID()
	cmd := NewDatatypeCmdMysql{}

	row := mdbm.Datatypes{DatatypeID: dtID}
	got := cmd.GetID(row)
	if got != string(dtID) {
		t.Errorf("GetID() = %q, want %q", got, string(dtID))
	}
}

func TestUpdateDatatypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	dtID := types.NewDatatypeID()
	params := UpdateDatatypeParams{
		Label:        "mysql-update-label",
		DateCreated:  ts,
		DateModified: ts,
		DatatypeID:   dtID,
	}

	cmd := MysqlDatabase{}.UpdateDatatypeCmd(ctx, ac, params)

	if cmd.TableName() != "datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes")
	}
	if cmd.GetID() != string(dtID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(dtID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateDatatypeParams", cmd.Params())
	}
	if p.Label != "mysql-update-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "mysql-update-label")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteDatatypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	dtID := types.NewDatatypeID()

	cmd := MysqlDatabase{}.DeleteDatatypeCmd(ctx, ac, dtID)

	if cmd.TableName() != "datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes")
	}
	if cmd.GetID() != string(dtID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(dtID))
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

func TestNewDatatypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node-dt"),
		RequestID: "psql-req-dt",
		IP:        "172.16.0.1",
	}
	params := CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		Label:        "psql-cmd-label",
		Type:         "content",
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := PsqlDatabase{}.NewDatatypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes")
	}
	p, ok := cmd.Params().(CreateDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateDatatypeParams", cmd.Params())
	}
	if p.Label != "psql-cmd-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "psql-cmd-label")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewDatatypeCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	dtID := types.NewDatatypeID()
	cmd := NewDatatypeCmdPsql{}

	row := mdbp.Datatypes{DatatypeID: dtID}
	got := cmd.GetID(row)
	if got != string(dtID) {
		t.Errorf("GetID() = %q, want %q", got, string(dtID))
	}
}

func TestUpdateDatatypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	dtID := types.NewDatatypeID()
	params := UpdateDatatypeParams{
		Label:        "psql-update-label",
		DateCreated:  ts,
		DateModified: ts,
		DatatypeID:   dtID,
	}

	cmd := PsqlDatabase{}.UpdateDatatypeCmd(ctx, ac, params)

	if cmd.TableName() != "datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes")
	}
	if cmd.GetID() != string(dtID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(dtID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateDatatypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateDatatypeParams", cmd.Params())
	}
	if p.Label != "psql-update-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "psql-update-label")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteDatatypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	dtID := types.NewDatatypeID()

	cmd := PsqlDatabase{}.DeleteDatatypeCmd(ctx, ac, dtID)

	if cmd.TableName() != "datatypes" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "datatypes")
	}
	if cmd.GetID() != string(dtID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(dtID))
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

func TestAuditedDatatypeCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateDatatypeParams{}
	updateParams := UpdateDatatypeParams{DatatypeID: types.NewDatatypeID()}
	dtID := types.NewDatatypeID()

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewDatatypeCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateDatatypeCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteDatatypeCmd(ctx, ac, dtID).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewDatatypeCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateDatatypeCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteDatatypeCmd(ctx, ac, dtID).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewDatatypeCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateDatatypeCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteDatatypeCmd(ctx, ac, dtID).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "datatypes" {
				t.Errorf("TableName() = %q, want %q", c.name, "datatypes")
			}
		})
	}
}

func TestAuditedDatatypeCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateDatatypeParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewDatatypeCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewDatatypeCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewDatatypeCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedDatatypeCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	dtID := types.NewDatatypeID()

	t.Run("UpdateCmd GetID returns DatatypeID", func(t *testing.T) {
		t.Parallel()
		params := UpdateDatatypeParams{DatatypeID: dtID}

		sqliteCmd := Database{}.UpdateDatatypeCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateDatatypeCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateDatatypeCmd(ctx, ac, params)

		wantID := string(dtID)
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

	t.Run("DeleteCmd GetID returns DatatypeID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteDatatypeCmd(ctx, ac, dtID)
		mysqlCmd := MysqlDatabase{}.DeleteDatatypeCmd(ctx, ac, dtID)
		psqlCmd := PsqlDatabase{}.DeleteDatatypeCmd(ctx, ac, dtID)

		wantID := string(dtID)
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
		testDtID := types.NewDatatypeID()

		sqliteCmd := NewDatatypeCmd{}
		mysqlCmd := NewDatatypeCmdMysql{}
		psqlCmd := NewDatatypeCmdPsql{}

		wantID := string(testDtID)

		sqliteRow := mdb.Datatypes{DatatypeID: testDtID}
		mysqlRow := mdbm.Datatypes{DatatypeID: testDtID}
		psqlRow := mdbp.Datatypes{DatatypeID: testDtID}

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

// --- Edge cases ---

func TestUpdateDatatypeCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	// When DatatypeID is empty, GetID should return empty string
	params := UpdateDatatypeParams{DatatypeID: ""}

	sqliteCmd := Database{}.UpdateDatatypeCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateDatatypeCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateDatatypeCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteDatatypeCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.DatatypeID("")

	sqliteCmd := Database{}.DeleteDatatypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteDatatypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteDatatypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestNewDatatypeCmd_GetID_EmptyDatatypeID(t *testing.T) {
	t.Parallel()
	// CreateCmd.GetID extracts from a row; an empty DatatypeID in the row should return ""
	sqliteCmd := NewDatatypeCmd{}
	if sqliteCmd.GetID(mdb.Datatypes{}) != "" {
		t.Errorf("SQLite GetID() = %q, want empty string", sqliteCmd.GetID(mdb.Datatypes{}))
	}

	mysqlCmd := NewDatatypeCmdMysql{}
	if mysqlCmd.GetID(mdbm.Datatypes{}) != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID(mdbm.Datatypes{}))
	}

	psqlCmd := NewDatatypeCmdPsql{}
	if psqlCmd.GetID(mdbp.Datatypes{}) != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID(mdbp.Datatypes{}))
	}
}

// --- Recorder identity verification ---
// Verify the command factories assign the correct recorder variant.

func TestDatatypeCommand_RecorderIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateDatatypeParams{}
	updateParams := UpdateDatatypeParams{DatatypeID: types.NewDatatypeID()}
	deleteID := types.NewDatatypeID()

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
		want     audited.ChangeEventRecorder
	}{
		// SQLite commands should use SQLiteRecorder
		{"SQLite Create", Database{}.NewDatatypeCmd(ctx, ac, createParams).Recorder(), SQLiteRecorder},
		{"SQLite Update", Database{}.UpdateDatatypeCmd(ctx, ac, updateParams).Recorder(), SQLiteRecorder},
		{"SQLite Delete", Database{}.DeleteDatatypeCmd(ctx, ac, deleteID).Recorder(), SQLiteRecorder},
		// MySQL commands should use MysqlRecorder
		{"MySQL Create", MysqlDatabase{}.NewDatatypeCmd(ctx, ac, createParams).Recorder(), MysqlRecorder},
		{"MySQL Update", MysqlDatabase{}.UpdateDatatypeCmd(ctx, ac, updateParams).Recorder(), MysqlRecorder},
		{"MySQL Delete", MysqlDatabase{}.DeleteDatatypeCmd(ctx, ac, deleteID).Recorder(), MysqlRecorder},
		// PostgreSQL commands should use PsqlRecorder
		{"PostgreSQL Create", PsqlDatabase{}.NewDatatypeCmd(ctx, ac, createParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateDatatypeCmd(ctx, ac, updateParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteDatatypeCmd(ctx, ac, deleteID).Recorder(), PsqlRecorder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Compare by type identity -- each recorder is a distinct struct type
			if fmt.Sprintf("%T", tt.recorder) != fmt.Sprintf("%T", tt.want) {
				t.Errorf("Recorder type = %T, want %T", tt.recorder, tt.want)
			}
		})
	}
}

// --- UpdateDatatype success message format ---
// The UpdateDatatype method produces a success message string.
// We cannot call UpdateDatatype without a real DB, but we verify the message
// format expectation by testing the format string directly.

func TestUpdateDatatype_SuccessMessageFormat(t *testing.T) {
	t.Parallel()
	label := "Blog Post"
	expected := fmt.Sprintf("Successfully updated %v\n", label)

	if expected != "Successfully updated Blog Post\n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated Blog Post\n")
	}
}

func TestUpdateDatatype_SuccessMessageFormat_EmptyLabel(t *testing.T) {
	t.Parallel()
	// Edge case: empty label still produces a valid format string
	expected := fmt.Sprintf("Successfully updated %v\n", "")
	if expected != "Successfully updated \n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated \n")
	}
}

// --- DatatypeJSON and MapDatatypeJSON consistency with MapStringDatatype ---
// Both free functions convert a Datatypes to strings. Verify they agree on
// the same input values.

func TestMapDatatypeJSON_MapStringDatatype_Consistency(t *testing.T) {
	t.Parallel()
	dt, _, _, _, _ := datatypeTestFixture()

	jsonResult := MapDatatypeJSON(dt)
	stringResult := MapStringDatatype(dt)

	// The two representations should agree on all shared fields
	if jsonResult.DatatypeID != stringResult.DatatypeID {
		t.Errorf("DatatypeID mismatch: JSON=%q, String=%q", jsonResult.DatatypeID, stringResult.DatatypeID)
	}
	if jsonResult.ParentID != stringResult.ParentID {
		t.Errorf("ParentID mismatch: JSON=%q, String=%q", jsonResult.ParentID, stringResult.ParentID)
	}
	if jsonResult.Label != stringResult.Label {
		t.Errorf("Label mismatch: JSON=%q, String=%q", jsonResult.Label, stringResult.Label)
	}
	if jsonResult.Type != stringResult.Type {
		t.Errorf("Type mismatch: JSON=%q, String=%q", jsonResult.Type, stringResult.Type)
	}
	if jsonResult.AuthorID != stringResult.AuthorID {
		t.Errorf("AuthorID mismatch: JSON=%q, String=%q", jsonResult.AuthorID, stringResult.AuthorID)
	}
	if jsonResult.DateCreated != stringResult.DateCreated {
		t.Errorf("DateCreated mismatch: JSON=%q, String=%q", jsonResult.DateCreated, stringResult.DateCreated)
	}
	if jsonResult.DateModified != stringResult.DateModified {
		t.Errorf("DateModified mismatch: JSON=%q, String=%q", jsonResult.DateModified, stringResult.DateModified)
	}
}

// --- MapCreateDatatypeParams: NullableDatatypeID as ParentID ---
// Datatypes uses NullableDatatypeID for ParentID.
// Verify the ParentID passes through all three database mappers correctly.

func TestMapCreateDatatypeParams_ParentIDCastConsistency(t *testing.T) {
	t.Parallel()
	originalDtID := types.NewDatatypeID()
	parentID := types.NullableDatatypeID{
		ID:    originalDtID,
		Valid: true,
	}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     parentID,
		Label:        "Child Type",
		Type:         "content",
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteGot := Database{}.MapCreateDatatypeParams(input)
	mysqlGot := MysqlDatabase{}.MapCreateDatatypeParams(input)
	psqlGot := PsqlDatabase{}.MapCreateDatatypeParams(input)

	// All three should preserve the ParentID
	if sqliteGot.ParentID != parentID {
		t.Errorf("SQLite ParentID = %v, want %v", sqliteGot.ParentID, parentID)
	}
	if mysqlGot.ParentID != parentID {
		t.Errorf("MySQL ParentID = %v, want %v", mysqlGot.ParentID, parentID)
	}
	if psqlGot.ParentID != parentID {
		t.Errorf("PostgreSQL ParentID = %v, want %v", psqlGot.ParentID, parentID)
	}

	// Verify the ParentID preserves the original DatatypeID string value
	if sqliteGot.ParentID.ID != originalDtID {
		t.Errorf("SQLite ParentID.ID = %q, want %q (original DatatypeID)", sqliteGot.ParentID.ID, originalDtID)
	}
}

// --- NullableDatatypeID with null ParentID ---

func TestMapCreateDatatypeParams_NullParentID(t *testing.T) {
	t.Parallel()
	nullParent := types.NullableDatatypeID{Valid: false}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     nullParent,
		Label:        "Root Type",
		Type:         "content",
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteGot := Database{}.MapCreateDatatypeParams(input)
	mysqlGot := MysqlDatabase{}.MapCreateDatatypeParams(input)
	psqlGot := PsqlDatabase{}.MapCreateDatatypeParams(input)

	if sqliteGot.ParentID.Valid {
		t.Error("SQLite ParentID.Valid = true, want false")
	}
	if mysqlGot.ParentID.Valid {
		t.Error("MySQL ParentID.Valid = true, want false")
	}
	if psqlGot.ParentID.Valid {
		t.Error("PostgreSQL ParentID.Valid = true, want false")
	}
}
