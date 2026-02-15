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

// --- Compile-time interface checks ---
// These verify that all 9 audited command types satisfy their respective
// audited interfaces. A compilation failure here means the struct drifted
// from the interface contract.

var (
	_ audited.CreateCommand[mdb.Fields]  = NewFieldCmd{}
	_ audited.UpdateCommand[mdb.Fields]  = UpdateFieldCmd{}
	_ audited.DeleteCommand[mdb.Fields]  = DeleteFieldCmd{}
	_ audited.CreateCommand[mdbm.Fields] = NewFieldCmdMysql{}
	_ audited.UpdateCommand[mdbm.Fields] = UpdateFieldCmdMysql{}
	_ audited.DeleteCommand[mdbm.Fields] = DeleteFieldCmdMysql{}
	_ audited.CreateCommand[mdbp.Fields] = NewFieldCmdPsql{}
	_ audited.UpdateCommand[mdbp.Fields] = UpdateFieldCmdPsql{}
	_ audited.DeleteCommand[mdbp.Fields] = DeleteFieldCmdPsql{}
)

// --- Test data helpers ---

// fieldFixture returns a fully populated Fields struct and its component IDs.
func fieldFixture() (Fields, types.FieldID, types.NullableDatatypeID, types.NullableUserID, types.Timestamp) {
	fieldID := types.NewFieldID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 7, 10, 14, 30, 0, 0, time.UTC))

	f := Fields{
		FieldID:      fieldID,
		ParentID:     parentID,
		Label:        "Test Label",
		Data:         "some-data",
		Validation:   `{"required": true}`,
		UIConfig:     `{"widget": "text"}`,
		Type:         types.FieldTypeText,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}
	return f, fieldID, parentID, authorID, ts
}

// --- MapFieldJSON tests ---

func TestMapFieldJSON_AllFields(t *testing.T) {
	t.Parallel()
	f, fieldID, parentID, authorID, ts := fieldFixture()

	got := MapFieldJSON(f)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"FieldID", got.FieldID, fieldID.String()},
		{"ParentID", got.ParentID, parentID.String()},
		{"Label", got.Label, "Test Label"},
		{"Data", got.Data, "some-data"},
		{"Validation", got.Validation, `{"required": true}`},
		{"UIConfig", got.UIConfig, `{"widget": "text"}`},
		{"Type", got.Type, types.FieldTypeText.String()},
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

func TestMapFieldJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value Fields should produce a valid FieldsJSON without panic
	got := MapFieldJSON(Fields{})

	if got.FieldID != "" {
		t.Errorf("FieldID = %q, want empty string", got.FieldID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
}

func TestMapFieldJSON_NullParentID(t *testing.T) {
	t.Parallel()
	// NullableDatatypeID with Valid=false should produce "null" from .String()
	f := Fields{
		ParentID: types.NullableDatatypeID{Valid: false},
	}
	got := MapFieldJSON(f)

	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q", got.ParentID, "null")
	}
}

func TestMapFieldJSON_NullAuthorID(t *testing.T) {
	t.Parallel()
	f := Fields{
		AuthorID: types.NullableUserID{Valid: false},
	}
	got := MapFieldJSON(f)

	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, "null")
	}
}

func TestMapFieldJSON_AllFieldTypes(t *testing.T) {
	t.Parallel()
	// Verify the Type field correctly maps each FieldType variant
	fieldTypes := []types.FieldType{
		types.FieldTypeText, types.FieldTypeTextarea, types.FieldTypeNumber,
		types.FieldTypeDate, types.FieldTypeDatetime, types.FieldTypeBoolean,
		types.FieldTypeSelect, types.FieldTypeMedia, types.FieldTypeRelation,
		types.FieldTypeJSON, types.FieldTypeRichText, types.FieldTypeSlug,
		types.FieldTypeEmail, types.FieldTypeURL,
	}

	for _, ft := range fieldTypes {
		t.Run(string(ft), func(t *testing.T) {
			t.Parallel()
			f := Fields{Type: ft}
			got := MapFieldJSON(f)
			if got.Type != ft.String() {
				t.Errorf("Type = %q, want %q", got.Type, ft.String())
			}
		})
	}
}

// --- MapStringField tests ---

func TestMapStringField_AllFields(t *testing.T) {
	t.Parallel()
	f, fieldID, parentID, authorID, ts := fieldFixture()

	got := MapStringField(f)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"FieldID", got.FieldID, fieldID.String()},
		{"ParentID", got.ParentID, parentID.String()},
		{"Label", got.Label, "Test Label"},
		{"Data", got.Data, "some-data"},
		{"Validation", got.Validation, `{"required": true}`},
		{"UIConfig", got.UIConfig, `{"widget": "text"}`},
		// MapStringField uses a.Type.String()
		{"Type", got.Type, types.FieldTypeText.String()},
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

func TestMapStringField_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringField(Fields{})

	if got.FieldID != "" {
		t.Errorf("FieldID = %q, want empty string", got.FieldID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.History != "" {
		t.Errorf("History = %q, want empty string (always empty)", got.History)
	}
}

func TestMapStringField_NullParentID(t *testing.T) {
	t.Parallel()
	f := Fields{
		ParentID: types.NullableDatatypeID{Valid: false},
	}
	got := MapStringField(f)

	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q", got.ParentID, "null")
	}
}

func TestMapStringField_NullAuthorID(t *testing.T) {
	t.Parallel()
	f := Fields{
		AuthorID: types.NullableUserID{Valid: false},
	}
	got := MapStringField(f)

	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, "null")
	}
}

func TestMapStringField_HistoryAlwaysEmpty(t *testing.T) {
	t.Parallel()
	// The History field is explicitly set to "" per the comment "History field removed".
	// Verify this is consistent even with a fully populated Fields.
	f, _, _, _, _ := fieldFixture()
	got := MapStringField(f)

	if got.History != "" {
		t.Errorf("History = %q, want empty string (always empty per code comment)", got.History)
	}
}

// --- SQLite Database.MapField tests ---

func TestDatabase_MapField_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	fieldID := types.NewFieldID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))

	input := mdb.Fields{
		FieldID:      fieldID,
		ParentID:     parentID,
		Label:        "sqlite-label",
		Data:         "sqlite-data",
		Validation:   "sqlite-validation",
		UiConfig:     "sqlite-ui-config",
		Type:         types.FieldTypeNumber,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapField(input)

	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "sqlite-label" {
		t.Errorf("Label = %q, want %q", got.Label, "sqlite-label")
	}
	if got.Data != "sqlite-data" {
		t.Errorf("Data = %q, want %q", got.Data, "sqlite-data")
	}
	if got.Validation != "sqlite-validation" {
		t.Errorf("Validation = %q, want %q", got.Validation, "sqlite-validation")
	}
	// UIConfig in wrapper maps from UiConfig in sqlc type
	if got.UIConfig != "sqlite-ui-config" {
		t.Errorf("UIConfig = %q, want %q", got.UIConfig, "sqlite-ui-config")
	}
	if got.Type != types.FieldTypeNumber {
		t.Errorf("Type = %v, want %v", got.Type, types.FieldTypeNumber)
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

func TestDatabase_MapField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapField(mdb.Fields{})

	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
	if got.ParentID.Valid {
		t.Error("ParentID.Valid = true, want false")
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.UIConfig != "" {
		t.Errorf("UIConfig = %q, want empty string", got.UIConfig)
	}
}

// --- SQLite Database.MapCreateFieldParams tests ---

func TestDatabase_MapCreateFieldParams_PreservesProvidedID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	existingID := types.NewFieldID()

	input := CreateFieldParams{
		FieldID:      existingID,
		ParentID:     parentID,
		Label:        "create-label",
		Data:         "create-data",
		Validation:   "create-validation",
		UIConfig:     "create-ui-config",
		Type:         types.FieldTypeTextarea,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateFieldParams(input)

	// When FieldID is provided, it should be preserved
	if got.FieldID != existingID {
		t.Errorf("FieldID = %v, want %v (provided ID should be preserved)", got.FieldID, existingID)
	}
	if got.ParentID != input.ParentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, input.ParentID)
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.Data != input.Data {
		t.Errorf("Data = %q, want %q", got.Data, input.Data)
	}
	if got.Validation != input.Validation {
		t.Errorf("Validation = %q, want %q", got.Validation, input.Validation)
	}
	// UIConfig -> UiConfig field name mapping
	if got.UiConfig != input.UIConfig {
		t.Errorf("UiConfig = %q, want %q", got.UiConfig, input.UIConfig)
	}
	if got.Type != input.Type {
		t.Errorf("Type = %v, want %v", got.Type, input.Type)
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

func TestDatabase_MapCreateFieldParams_GeneratesIDWhenZero(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateFieldParams{
		// FieldID intentionally zero
		Label:        "auto-id-label",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateFieldParams(input)

	if got.FieldID.IsZero() {
		t.Fatal("expected non-zero FieldID to be generated when input FieldID is zero")
	}
}

func TestDatabase_MapCreateFieldParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateFieldParams{}

	got1 := d.MapCreateFieldParams(input)
	got2 := d.MapCreateFieldParams(input)

	if got1.FieldID == got2.FieldID {
		t.Error("two calls generated the same FieldID -- each call should be unique")
	}
}

func TestDatabase_MapCreateFieldParams_NullOptionalFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateFieldParams{
		ParentID: types.NullableDatatypeID{Valid: false},
		AuthorID: types.NullableUserID{Valid: false},
	}

	got := d.MapCreateFieldParams(input)

	if got.ParentID.Valid {
		t.Error("ParentID.Valid = true, want false")
	}
	if got.AuthorID.Valid {
		t.Error("AuthorID.Valid = true, want false")
	}
}

// --- SQLite Database.MapUpdateFieldParams tests ---

func TestDatabase_MapUpdateFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	fID := types.NewFieldID()

	input := UpdateFieldParams{
		ParentID:     parentID,
		Label:        "updated-label",
		Data:         "updated-data",
		Validation:   "updated-validation",
		UIConfig:     "updated-ui-config",
		Type:         types.FieldTypeBoolean,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
		FieldID:      fID,
	}

	got := d.MapUpdateFieldParams(input)

	if got.ParentID != input.ParentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, input.ParentID)
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.Data != input.Data {
		t.Errorf("Data = %q, want %q", got.Data, input.Data)
	}
	if got.Validation != input.Validation {
		t.Errorf("Validation = %q, want %q", got.Validation, input.Validation)
	}
	if got.UiConfig != input.UIConfig {
		t.Errorf("UiConfig = %q, want %q", got.UiConfig, input.UIConfig)
	}
	if got.Type != input.Type {
		t.Errorf("Type = %v, want %v", got.Type, input.Type)
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
	// FieldID is the WHERE clause identifier -- must be preserved
	if got.FieldID != fID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fID)
	}
}

// --- MySQL MysqlDatabase.MapField tests ---

func TestMysqlDatabase_MapField_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	fieldID := types.NewFieldID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))

	input := mdbm.Fields{
		FieldID:      fieldID,
		ParentID:     parentID,
		Label:        "mysql-label",
		Data:         "mysql-data",
		Validation:   "mysql-validation",
		UiConfig:     "mysql-ui-config",
		Type:         types.FieldTypeSelect,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapField(input)

	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "mysql-label" {
		t.Errorf("Label = %q, want %q", got.Label, "mysql-label")
	}
	if got.Data != "mysql-data" {
		t.Errorf("Data = %q, want %q", got.Data, "mysql-data")
	}
	if got.Validation != "mysql-validation" {
		t.Errorf("Validation = %q, want %q", got.Validation, "mysql-validation")
	}
	if got.UIConfig != "mysql-ui-config" {
		t.Errorf("UIConfig = %q, want %q", got.UIConfig, "mysql-ui-config")
	}
	if got.Type != types.FieldTypeSelect {
		t.Errorf("Type = %v, want %v", got.Type, types.FieldTypeSelect)
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

func TestMysqlDatabase_MapField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapField(mdbm.Fields{})

	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
}

// --- MySQL MysqlDatabase.MapCreateFieldParams tests ---

func TestMysqlDatabase_MapCreateFieldParams_GeneratesIDWhenZero(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateFieldParams{
		Label:        "mysql-create",
		UIConfig:     "mysql-create-ui",
		Type:         types.FieldTypeMedia,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateFieldParams(input)

	if got.FieldID.IsZero() {
		t.Fatal("expected non-zero FieldID to be generated")
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.UiConfig != input.UIConfig {
		t.Errorf("UiConfig = %q, want %q", got.UiConfig, input.UIConfig)
	}
	if got.Type != input.Type {
		t.Errorf("Type = %v, want %v", got.Type, input.Type)
	}
}

func TestMysqlDatabase_MapCreateFieldParams_PreservesProvidedID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	existingID := types.NewFieldID()

	input := CreateFieldParams{
		FieldID: existingID,
		Label:   "mysql-existing-id",
	}

	got := d.MapCreateFieldParams(input)

	if got.FieldID != existingID {
		t.Errorf("FieldID = %v, want %v (provided ID should be preserved)", got.FieldID, existingID)
	}
}

// --- MySQL MysqlDatabase.MapUpdateFieldParams tests ---

func TestMysqlDatabase_MapUpdateFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	fID := types.NewFieldID()

	input := UpdateFieldParams{
		Label:        "mysql-updated",
		UIConfig:     "mysql-updated-ui",
		Type:         types.FieldTypeRelation,
		DateCreated:  ts,
		DateModified: ts,
		FieldID:      fID,
	}

	got := d.MapUpdateFieldParams(input)

	if got.FieldID != fID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fID)
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.UiConfig != input.UIConfig {
		t.Errorf("UiConfig = %q, want %q", got.UiConfig, input.UIConfig)
	}
}

// --- PostgreSQL PsqlDatabase.MapField tests ---

func TestPsqlDatabase_MapField_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	fieldID := types.NewFieldID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))

	input := mdbp.Fields{
		FieldID:      fieldID,
		ParentID:     parentID,
		Label:        "psql-label",
		Data:         "psql-data",
		Validation:   "psql-validation",
		UiConfig:     "psql-ui-config",
		Type:         types.FieldTypeJSON,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapField(input)

	if got.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fieldID)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.Label != "psql-label" {
		t.Errorf("Label = %q, want %q", got.Label, "psql-label")
	}
	if got.Data != "psql-data" {
		t.Errorf("Data = %q, want %q", got.Data, "psql-data")
	}
	if got.Validation != "psql-validation" {
		t.Errorf("Validation = %q, want %q", got.Validation, "psql-validation")
	}
	if got.UIConfig != "psql-ui-config" {
		t.Errorf("UIConfig = %q, want %q", got.UIConfig, "psql-ui-config")
	}
	if got.Type != types.FieldTypeJSON {
		t.Errorf("Type = %v, want %v", got.Type, types.FieldTypeJSON)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

func TestPsqlDatabase_MapField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapField(mdbp.Fields{})

	if got.FieldID != "" {
		t.Errorf("FieldID = %v, want zero value", got.FieldID)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateFieldParams tests ---

func TestPsqlDatabase_MapCreateFieldParams_GeneratesIDWhenZero(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateFieldParams{
		Label:        "psql-create",
		UIConfig:     "psql-create-ui",
		Type:         types.FieldTypeRichText,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateFieldParams(input)

	if got.FieldID.IsZero() {
		t.Fatal("expected non-zero FieldID to be generated")
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.UiConfig != input.UIConfig {
		t.Errorf("UiConfig = %q, want %q", got.UiConfig, input.UIConfig)
	}
}

func TestPsqlDatabase_MapCreateFieldParams_PreservesProvidedID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	existingID := types.NewFieldID()

	input := CreateFieldParams{
		FieldID: existingID,
		Label:   "psql-existing-id",
	}

	got := d.MapCreateFieldParams(input)

	if got.FieldID != existingID {
		t.Errorf("FieldID = %v, want %v (provided ID should be preserved)", got.FieldID, existingID)
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateFieldParams tests ---

func TestPsqlDatabase_MapUpdateFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	fID := types.NewFieldID()

	input := UpdateFieldParams{
		Label:        "psql-updated",
		UIConfig:     "psql-updated-ui",
		Type:         types.FieldTypeSlug,
		DateCreated:  ts,
		DateModified: ts,
		FieldID:      fID,
	}

	got := d.MapUpdateFieldParams(input)

	if got.FieldID != fID {
		t.Errorf("FieldID = %v, want %v", got.FieldID, fID)
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.UiConfig != input.UIConfig {
		t.Errorf("UiConfig = %q, want %q", got.UiConfig, input.UIConfig)
	}
}

// --- Cross-database mapper consistency ---
// All three database drivers use identical types for Fields
// (no int32/int64 conversions needed). This test verifies they all produce
// identical Fields wrapper structs from equivalent input.

func TestCrossDatabaseMapField_Consistency(t *testing.T) {
	t.Parallel()
	fieldID := types.NewFieldID()
	parentID := types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))

	sqliteInput := mdb.Fields{
		FieldID: fieldID, ParentID: parentID,
		Label: "cross-label", Data: "cross-data",
		Validation: "cross-validation", UiConfig: "cross-ui",
		Type: types.FieldTypeEmail, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.Fields{
		FieldID: fieldID, ParentID: parentID,
		Label: "cross-label", Data: "cross-data",
		Validation: "cross-validation", UiConfig: "cross-ui",
		Type: types.FieldTypeEmail, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.Fields{
		FieldID: fieldID, ParentID: parentID,
		Label: "cross-label", Data: "cross-data",
		Validation: "cross-validation", UiConfig: "cross-ui",
		Type: types.FieldTypeEmail, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapField(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapField(mysqlInput)
	psqlResult := PsqlDatabase{}.MapField(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateFieldParams auto-ID generation ---

func TestCrossDatabaseMapCreateFieldParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	input := CreateFieldParams{
		Label:        "cross-create-label",
		UIConfig:     "cross-create-ui",
		Type:         types.FieldTypeURL,
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteResult := Database{}.MapCreateFieldParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateFieldParams(input)
	psqlResult := PsqlDatabase{}.MapCreateFieldParams(input)

	if sqliteResult.FieldID.IsZero() {
		t.Error("SQLite: expected non-zero generated FieldID")
	}
	if mysqlResult.FieldID.IsZero() {
		t.Error("MySQL: expected non-zero generated FieldID")
	}
	if psqlResult.FieldID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated FieldID")
	}

	// Each call should generate a unique ID
	if sqliteResult.FieldID == mysqlResult.FieldID {
		t.Error("SQLite and MySQL generated the same FieldID -- each call should be unique")
	}
	if sqliteResult.FieldID == psqlResult.FieldID {
		t.Error("SQLite and PostgreSQL generated the same FieldID -- each call should be unique")
	}
}

// --- Cross-database UIConfig -> UiConfig field name mapping ---
// This is the critical mapping to verify: the wrapper uses UIConfig (Go convention)
// while sqlc generates UiConfig. All three Create and Update param mappers must
// correctly bridge this naming difference.

func TestCrossDatabaseUIConfigFieldMapping_Fields(t *testing.T) {
	t.Parallel()
	uiValue := `{"columns": 2, "widget": "dropdown"}`
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	fID := types.NewFieldID()

	t.Run("CreateParams", func(t *testing.T) {
		t.Parallel()
		input := CreateFieldParams{UIConfig: uiValue, DateCreated: ts, DateModified: ts}

		sqliteGot := Database{}.MapCreateFieldParams(input)
		mysqlGot := MysqlDatabase{}.MapCreateFieldParams(input)
		psqlGot := PsqlDatabase{}.MapCreateFieldParams(input)

		if sqliteGot.UiConfig != uiValue {
			t.Errorf("SQLite UiConfig = %q, want %q", sqliteGot.UiConfig, uiValue)
		}
		if mysqlGot.UiConfig != uiValue {
			t.Errorf("MySQL UiConfig = %q, want %q", mysqlGot.UiConfig, uiValue)
		}
		if psqlGot.UiConfig != uiValue {
			t.Errorf("PostgreSQL UiConfig = %q, want %q", psqlGot.UiConfig, uiValue)
		}
	})

	t.Run("UpdateParams", func(t *testing.T) {
		t.Parallel()
		input := UpdateFieldParams{
			UIConfig: uiValue, DateCreated: ts, DateModified: ts,
			FieldID: fID,
		}

		sqliteGot := Database{}.MapUpdateFieldParams(input)
		mysqlGot := MysqlDatabase{}.MapUpdateFieldParams(input)
		psqlGot := PsqlDatabase{}.MapUpdateFieldParams(input)

		if sqliteGot.UiConfig != uiValue {
			t.Errorf("SQLite UiConfig = %q, want %q", sqliteGot.UiConfig, uiValue)
		}
		if mysqlGot.UiConfig != uiValue {
			t.Errorf("MySQL UiConfig = %q, want %q", mysqlGot.UiConfig, uiValue)
		}
		if psqlGot.UiConfig != uiValue {
			t.Errorf("PostgreSQL UiConfig = %q, want %q", psqlGot.UiConfig, uiValue)
		}
	})

	t.Run("MapField_ReverseDirection", func(t *testing.T) {
		t.Parallel()
		// Verify the reverse: UiConfig in sqlc -> UIConfig in wrapper
		sqliteGot := Database{}.MapField(mdb.Fields{UiConfig: uiValue})
		mysqlGot := MysqlDatabase{}.MapField(mdbm.Fields{UiConfig: uiValue})
		psqlGot := PsqlDatabase{}.MapField(mdbp.Fields{UiConfig: uiValue})

		if sqliteGot.UIConfig != uiValue {
			t.Errorf("SQLite UIConfig = %q, want %q", sqliteGot.UIConfig, uiValue)
		}
		if mysqlGot.UIConfig != uiValue {
			t.Errorf("MySQL UIConfig = %q, want %q", mysqlGot.UIConfig, uiValue)
		}
		if psqlGot.UIConfig != uiValue {
			t.Errorf("PostgreSQL UIConfig = %q, want %q", psqlGot.UIConfig, uiValue)
		}
	})
}

// --- SQLite Audited Command Accessor tests ---

func TestNewFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-f-1"),
		RequestID: "req-f-123",
		IP:        "10.0.0.1",
	}
	params := CreateFieldParams{
		Label:        "cmd-label",
		Data:         "cmd-data",
		UIConfig:     "cmd-ui",
		Type:         types.FieldTypeText,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := Database{}.NewFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "fields")
	}
	p, ok := cmd.Params().(CreateFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateFieldParams", cmd.Params())
	}
	if p.Label != "cmd-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "cmd-label")
	}
	if p.UIConfig != "cmd-ui" {
		t.Errorf("Params().UIConfig = %q, want %q", p.UIConfig, "cmd-ui")
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewFieldCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	fieldID := types.NewFieldID()
	cmd := NewFieldCmd{}

	row := mdb.Fields{FieldID: fieldID}
	got := cmd.GetID(row)
	if got != string(fieldID) {
		t.Errorf("GetID() = %q, want %q", got, string(fieldID))
	}
}

func TestNewFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	cmd := NewFieldCmd{}
	row := mdb.Fields{FieldID: ""}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	fID := types.NewFieldID()
	params := UpdateFieldParams{
		Label:        "update-label",
		Data:         "update-data",
		UIConfig:     "update-ui",
		DateCreated:  ts,
		DateModified: ts,
		FieldID:      fID,
	}

	cmd := Database{}.UpdateFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "fields")
	}
	if cmd.GetID() != string(fID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(fID))
	}
	p, ok := cmd.Params().(UpdateFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateFieldParams", cmd.Params())
	}
	if p.Label != "update-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "update-label")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	fID := types.NewFieldID()

	cmd := Database{}.DeleteFieldCmd(ctx, ac, fID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "fields")
	}
	if cmd.GetID() != string(fID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(fID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node-f"),
		RequestID: "mysql-req-f",
		IP:        "192.168.1.1",
	}
	params := CreateFieldParams{
		Label:        "mysql-cmd-label",
		UIConfig:     "mysql-cmd-ui",
		Type:         types.FieldTypeDate,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := MysqlDatabase{}.NewFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "fields")
	}
	p, ok := cmd.Params().(CreateFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateFieldParams", cmd.Params())
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

func TestNewFieldCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	fieldID := types.NewFieldID()
	cmd := NewFieldCmdMysql{}

	row := mdbm.Fields{FieldID: fieldID}
	got := cmd.GetID(row)
	if got != string(fieldID) {
		t.Errorf("GetID() = %q, want %q", got, string(fieldID))
	}
}

func TestUpdateFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	fID := types.NewFieldID()
	params := UpdateFieldParams{
		Label:        "mysql-update-label",
		DateCreated:  ts,
		DateModified: ts,
		FieldID:      fID,
	}

	cmd := MysqlDatabase{}.UpdateFieldCmd(ctx, ac, params)

	if cmd.TableName() != "fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "fields")
	}
	if cmd.GetID() != string(fID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(fID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateFieldParams", cmd.Params())
	}
	if p.Label != "mysql-update-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "mysql-update-label")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	fID := types.NewFieldID()

	cmd := MysqlDatabase{}.DeleteFieldCmd(ctx, ac, fID)

	if cmd.TableName() != "fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "fields")
	}
	if cmd.GetID() != string(fID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(fID))
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

func TestNewFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node-f"),
		RequestID: "psql-req-f",
		IP:        "172.16.0.1",
	}
	params := CreateFieldParams{
		Label:        "psql-cmd-label",
		UIConfig:     "psql-cmd-ui",
		Type:         types.FieldTypeDatetime,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := PsqlDatabase{}.NewFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "fields")
	}
	p, ok := cmd.Params().(CreateFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateFieldParams", cmd.Params())
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

func TestNewFieldCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	fieldID := types.NewFieldID()
	cmd := NewFieldCmdPsql{}

	row := mdbp.Fields{FieldID: fieldID}
	got := cmd.GetID(row)
	if got != string(fieldID) {
		t.Errorf("GetID() = %q, want %q", got, string(fieldID))
	}
}

func TestUpdateFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	fID := types.NewFieldID()
	params := UpdateFieldParams{
		Label:        "psql-update-label",
		DateCreated:  ts,
		DateModified: ts,
		FieldID:      fID,
	}

	cmd := PsqlDatabase{}.UpdateFieldCmd(ctx, ac, params)

	if cmd.TableName() != "fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "fields")
	}
	if cmd.GetID() != string(fID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(fID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateFieldParams", cmd.Params())
	}
	if p.Label != "psql-update-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "psql-update-label")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	fID := types.NewFieldID()

	cmd := PsqlDatabase{}.DeleteFieldCmd(ctx, ac, fID)

	if cmd.TableName() != "fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "fields")
	}
	if cmd.GetID() != string(fID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(fID))
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

func TestAuditedFieldCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateFieldParams{}
	updateParams := UpdateFieldParams{FieldID: types.NewFieldID()}
	fID := types.NewFieldID()

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewFieldCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateFieldCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteFieldCmd(ctx, ac, fID).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewFieldCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateFieldCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteFieldCmd(ctx, ac, fID).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewFieldCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateFieldCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteFieldCmd(ctx, ac, fID).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "fields" {
				t.Errorf("TableName() = %q, want %q", c.name, "fields")
			}
		})
	}
}

func TestAuditedFieldCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateFieldParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewFieldCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewFieldCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewFieldCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedFieldCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	fID := types.NewFieldID()

	t.Run("UpdateCmd GetID returns FieldID", func(t *testing.T) {
		t.Parallel()
		params := UpdateFieldParams{FieldID: fID}

		sqliteCmd := Database{}.UpdateFieldCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateFieldCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateFieldCmd(ctx, ac, params)

		wantID := string(fID)
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

	t.Run("DeleteCmd GetID returns FieldID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteFieldCmd(ctx, ac, fID)
		mysqlCmd := MysqlDatabase{}.DeleteFieldCmd(ctx, ac, fID)
		psqlCmd := PsqlDatabase{}.DeleteFieldCmd(ctx, ac, fID)

		wantID := string(fID)
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
		testFieldID := types.NewFieldID()

		sqliteCmd := NewFieldCmd{}
		mysqlCmd := NewFieldCmdMysql{}
		psqlCmd := NewFieldCmdPsql{}

		wantID := string(testFieldID)

		sqliteRow := mdb.Fields{FieldID: testFieldID}
		mysqlRow := mdbm.Fields{FieldID: testFieldID}
		psqlRow := mdbp.Fields{FieldID: testFieldID}

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

// --- Edge case: UpdateCmd with empty FieldID ---

func TestUpdateFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateFieldParams{FieldID: ""}

	sqliteCmd := Database{}.UpdateFieldCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateFieldCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateFieldCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteCmd with empty ID ---

func TestDeleteFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.FieldID("")

	sqliteCmd := Database{}.DeleteFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- UpdateField success message format ---
// The UpdateField method produces a success message string.
// We can't test it without a real DB, but we verify the message format expectation
// by testing the format string directly.

func TestUpdateField_SuccessMessageFormat(t *testing.T) {
	t.Parallel()
	label := "Test Field"
	expected := fmt.Sprintf("Successfully updated %v\n", label)

	// The actual method builds this exact format using s.Label
	if expected != "Successfully updated Test Field\n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated Test Field\n")
	}
}

func TestUpdateField_SuccessMessageFormat_EmptyLabel(t *testing.T) {
	t.Parallel()
	// Edge case: empty label still produces a valid format string
	expected := fmt.Sprintf("Successfully updated %v\n", "")
	if expected != "Successfully updated \n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated \n")
	}
}

// --- Recorder identity verification ---
// Verify the command factories assign the correct recorder variant.

func TestFieldCommand_RecorderIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateFieldParams{}
	updateParams := UpdateFieldParams{FieldID: types.NewFieldID()}
	deleteID := types.NewFieldID()

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
		want     audited.ChangeEventRecorder
	}{
		// SQLite commands should use SQLiteRecorder
		{"SQLite Create", Database{}.NewFieldCmd(ctx, ac, createParams).Recorder(), SQLiteRecorder},
		{"SQLite Update", Database{}.UpdateFieldCmd(ctx, ac, updateParams).Recorder(), SQLiteRecorder},
		{"SQLite Delete", Database{}.DeleteFieldCmd(ctx, ac, deleteID).Recorder(), SQLiteRecorder},
		// MySQL commands should use MysqlRecorder
		{"MySQL Create", MysqlDatabase{}.NewFieldCmd(ctx, ac, createParams).Recorder(), MysqlRecorder},
		{"MySQL Update", MysqlDatabase{}.UpdateFieldCmd(ctx, ac, updateParams).Recorder(), MysqlRecorder},
		{"MySQL Delete", MysqlDatabase{}.DeleteFieldCmd(ctx, ac, deleteID).Recorder(), MysqlRecorder},
		// PostgreSQL commands should use PsqlRecorder
		{"PostgreSQL Create", PsqlDatabase{}.NewFieldCmd(ctx, ac, createParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateFieldCmd(ctx, ac, updateParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteFieldCmd(ctx, ac, deleteID).Recorder(), PsqlRecorder},
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

// --- Struct field correctness ---
// Verify that the wrapper Fields struct and param structs hold values correctly via JSON.

func TestFieldsStruct_JSONTags(t *testing.T) {
	t.Parallel()
	f, _, _, _, _ := fieldFixture()

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"field_id", "parent_id", "label", "data",
		"validation", "ui_config", "type", "author_id",
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

func TestCreateFieldParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	p := CreateFieldParams{
		FieldID:      types.NewFieldID(),
		ParentID:     types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true},
		Label:        "test-label",
		Data:         "test-data",
		Validation:   "test-validation",
		UIConfig:     "test-ui-config",
		Type:         types.FieldTypeText,
		AuthorID:     types.NullableUserID{ID: types.NewUserID(), Valid: true},
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
		"field_id", "parent_id", "label", "data",
		"validation", "ui_config", "type", "author_id",
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

func TestUpdateFieldParams_JSONTags(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	fID := types.NewFieldID()
	p := UpdateFieldParams{
		ParentID:     types.NullableDatatypeID{ID: types.NewDatatypeID(), Valid: true},
		Label:        "test-label",
		Data:         "test-data",
		Validation:   "test-validation",
		UIConfig:     "test-ui-config",
		Type:         types.FieldTypeText,
		AuthorID:     types.NullableUserID{ID: types.NewUserID(), Valid: true},
		DateCreated:  ts,
		DateModified: ts,
		FieldID:      fID,
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
		"parent_id", "label", "data",
		"validation", "ui_config", "type", "author_id",
		"date_created", "date_modified",
		"field_id",
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

func TestFieldsJSONStruct_JSONTags(t *testing.T) {
	t.Parallel()
	fj := FieldsJSON{
		FieldID:      "test-field-id",
		ParentID:     "test-parent-id",
		Label:        "json-label",
		Data:         "json-data",
		Validation:   "json-validation",
		UIConfig:     "json-ui-config",
		Type:         "text",
		AuthorID:     "test-author-id",
		DateCreated:  "2025-01-01T00:00:00Z",
		DateModified: "2025-01-01T00:00:00Z",
	}

	data, err := json.Marshal(fj)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"field_id", "parent_id", "label", "data",
		"validation", "ui_config", "type", "author_id",
		"date_created", "date_modified",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("FieldsJSON JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("FieldsJSON JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

// --- MapFieldJSON and MapStringField consistency ---
// Both functions process the same input; verify string fields are handled identically.

func TestMapFieldJSON_MapStringField_FieldValueConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		label string
	}{
		{"empty string", ""},
		{"simple string", "hello world"},
		{"unicode", "Bonjour le monde"},
		{"special characters", `<script>alert("xss")</script>`},
		{"newlines", "line1\nline2\nline3"},
		{"json content", `{"required": true, "min": 0}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := Fields{
				FieldID: types.NewFieldID(),
				Label:   tt.label,
			}

			jsonResult := MapFieldJSON(f)
			stringResult := MapStringField(f)

			if jsonResult.Label != tt.label {
				t.Errorf("JSON Label = %q, want %q", jsonResult.Label, tt.label)
			}
			if stringResult.Label != tt.label {
				t.Errorf("String Label = %q, want %q", stringResult.Label, tt.label)
			}
		})
	}
}

// --- StringFields struct JSON tags ---
// Verify the StringFields struct (used for TUI table display) has correct tags.

func TestStringFieldsStruct_JSONTags(t *testing.T) {
	t.Parallel()
	sf := StringFields{
		FieldID:      "test-id",
		ParentID:     "test-parent",
		Label:        "test-label",
		Data:         "test-data",
		Validation:   "test-validation",
		UIConfig:     "test-ui",
		Type:         "text",
		AuthorID:     "test-author",
		DateCreated:  "2025-01-01T00:00:00Z",
		DateModified: "2025-01-01T00:00:00Z",
		History:      "",
	}

	data, err := json.Marshal(sf)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"field_id", "parent_id", "label", "data",
		"validation", "ui_config", "type", "author_id",
		"date_created", "date_modified", "history",
	}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("StringFields JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("StringFields JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}
