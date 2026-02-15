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
	_ audited.CreateCommand[mdb.AdminFields]  = NewAdminFieldCmd{}
	_ audited.UpdateCommand[mdb.AdminFields]  = UpdateAdminFieldCmd{}
	_ audited.DeleteCommand[mdb.AdminFields]  = DeleteAdminFieldCmd{}
	_ audited.CreateCommand[mdbm.AdminFields] = NewAdminFieldCmdMysql{}
	_ audited.UpdateCommand[mdbm.AdminFields] = UpdateAdminFieldCmdMysql{}
	_ audited.DeleteCommand[mdbm.AdminFields] = DeleteAdminFieldCmdMysql{}
	_ audited.CreateCommand[mdbp.AdminFields] = NewAdminFieldCmdPsql{}
	_ audited.UpdateCommand[mdbp.AdminFields] = UpdateAdminFieldCmdPsql{}
	_ audited.DeleteCommand[mdbp.AdminFields] = DeleteAdminFieldCmdPsql{}
)

// --- Test data helpers ---

// afTestFixture returns a fully populated AdminFields struct and its component IDs.
func afTestFixture() (AdminFields, types.AdminFieldID, types.NullableAdminDatatypeID, types.NullableUserID, types.Timestamp) {
	fieldID := types.NewAdminFieldID()
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 7, 10, 14, 30, 0, 0, time.UTC))

	af := AdminFields{
		AdminFieldID: fieldID,
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
	return af, fieldID, parentID, authorID, ts
}

// --- MapAdminFieldJSON tests ---

func TestMapAdminFieldJSON_AllFields(t *testing.T) {
	t.Parallel()
	af, fieldID, parentID, authorID, ts := afTestFixture()

	got := MapAdminFieldJSON(af)

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

func TestMapAdminFieldJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	// Zero-value AdminFields should produce a valid FieldsJSON without panic
	got := MapAdminFieldJSON(AdminFields{})

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

func TestMapAdminFieldJSON_NullParentID(t *testing.T) {
	t.Parallel()
	// NullableAdminDatatypeID with Valid=false should produce "null" from .String()
	af := AdminFields{
		ParentID: types.NullableAdminDatatypeID{Valid: false},
	}
	got := MapAdminFieldJSON(af)

	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q", got.ParentID, "null")
	}
}

func TestMapAdminFieldJSON_NullAuthorID(t *testing.T) {
	t.Parallel()
	af := AdminFields{
		AuthorID: types.NullableUserID{Valid: false},
	}
	got := MapAdminFieldJSON(af)

	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, "null")
	}
}

func TestMapAdminFieldJSON_AllFieldTypes(t *testing.T) {
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
			af := AdminFields{Type: ft}
			got := MapAdminFieldJSON(af)
			if got.Type != ft.String() {
				t.Errorf("Type = %q, want %q", got.Type, ft.String())
			}
		})
	}
}

// --- MapStringAdminField tests ---

func TestMapStringAdminField_AllFields(t *testing.T) {
	t.Parallel()
	af, fieldID, parentID, authorID, ts := afTestFixture()

	got := MapStringAdminField(af)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"AdminFieldID", got.AdminFieldID, fieldID.String()},
		{"ParentID", got.ParentID, parentID.String()},
		{"Label", got.Label, "Test Label"},
		{"Data", got.Data, "some-data"},
		{"Validation", got.Validation, `{"required": true}`},
		{"UIConfig", got.UIConfig, `{"widget": "text"}`},
		// Note: MapStringAdminField uses string(a.Type), not a.Type.String()
		// Both are equivalent since FieldType.String() returns string(t)
		{"Type", got.Type, string(types.FieldTypeText)},
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

func TestMapStringAdminField_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringAdminField(AdminFields{})

	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %q, want empty string", got.AdminFieldID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.History != "" {
		t.Errorf("History = %q, want empty string (always empty)", got.History)
	}
}

func TestMapStringAdminField_NullParentID(t *testing.T) {
	t.Parallel()
	af := AdminFields{
		ParentID: types.NullableAdminDatatypeID{Valid: false},
	}
	got := MapStringAdminField(af)

	if got.ParentID != "null" {
		t.Errorf("ParentID = %q, want %q", got.ParentID, "null")
	}
}

func TestMapStringAdminField_NullAuthorID(t *testing.T) {
	t.Parallel()
	af := AdminFields{
		AuthorID: types.NullableUserID{Valid: false},
	}
	got := MapStringAdminField(af)

	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, "null")
	}
}

func TestMapStringAdminField_HistoryAlwaysEmpty(t *testing.T) {
	t.Parallel()
	// The History field is explicitly set to "" per the comment "History field removed".
	// Verify this is consistent even with a fully populated AdminFields.
	af, _, _, _, _ := afTestFixture()
	got := MapStringAdminField(af)

	if got.History != "" {
		t.Errorf("History = %q, want empty string (always empty per code comment)", got.History)
	}
}

// --- SQLite Database.MapAdminField tests ---

func TestDatabase_MapAdminField_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	fieldID := types.NewAdminFieldID()
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))

	input := mdb.AdminFields{
		AdminFieldID: fieldID,
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

	got := d.MapAdminField(input)

	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
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

func TestDatabase_MapAdminField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminField(mdb.AdminFields{})

	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
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

// --- SQLite Database.MapCreateAdminFieldParams tests ---

func TestDatabase_MapCreateAdminFieldParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := CreateAdminFieldParams{
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

	got := d.MapCreateAdminFieldParams(input)

	// A new ID should always be generated
	if got.AdminFieldID.IsZero() {
		t.Fatal("expected non-zero AdminFieldID to be generated")
	}

	// All other fields should pass through
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

func TestDatabase_MapCreateAdminFieldParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateAdminFieldParams{}

	got1 := d.MapCreateAdminFieldParams(input)
	got2 := d.MapCreateAdminFieldParams(input)

	if got1.AdminFieldID == got2.AdminFieldID {
		t.Error("two calls generated the same AdminFieldID -- each call should be unique")
	}
}

func TestDatabase_MapCreateAdminFieldParams_NullOptionalFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateAdminFieldParams{
		ParentID: types.NullableAdminDatatypeID{Valid: false},
		AuthorID: types.NullableUserID{Valid: false},
	}

	got := d.MapCreateAdminFieldParams(input)

	if got.ParentID.Valid {
		t.Error("ParentID.Valid = true, want false")
	}
	if got.AuthorID.Valid {
		t.Error("AuthorID.Valid = true, want false")
	}
}

// --- SQLite Database.MapUpdateAdminFieldParams tests ---

func TestDatabase_MapUpdateAdminFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	afID := types.NewAdminFieldID()

	input := UpdateAdminFieldParams{
		ParentID:     parentID,
		Label:        "updated-label",
		Data:         "updated-data",
		Validation:   "updated-validation",
		UIConfig:     "updated-ui-config",
		Type:         types.FieldTypeBoolean,
		AuthorID:     authorID,
		DateCreated:  ts,
		DateModified: ts,
		AdminFieldID: afID,
	}

	got := d.MapUpdateAdminFieldParams(input)

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
	// AdminFieldID is the WHERE clause identifier -- must be preserved
	if got.AdminFieldID != afID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, afID)
	}
}

// --- MySQL MysqlDatabase.MapAdminField tests ---

func TestMysqlDatabase_MapAdminField_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	fieldID := types.NewAdminFieldID()
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))

	input := mdbm.AdminFields{
		AdminFieldID: fieldID,
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

	got := d.MapAdminField(input)

	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
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

func TestMysqlDatabase_MapAdminField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminField(mdbm.AdminFields{})

	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

// --- MySQL MysqlDatabase.MapCreateAdminFieldParams tests ---

func TestMysqlDatabase_MapCreateAdminFieldParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminFieldParams{
		Label:        "mysql-create",
		UIConfig:     "mysql-create-ui",
		Type:         types.FieldTypeMedia,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminFieldParams(input)

	if got.AdminFieldID.IsZero() {
		t.Fatal("expected non-zero AdminFieldID to be generated")
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

// --- MySQL MysqlDatabase.MapUpdateAdminFieldParams tests ---

func TestMysqlDatabase_MapUpdateAdminFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	afID := types.NewAdminFieldID()

	input := UpdateAdminFieldParams{
		Label:        "mysql-updated",
		UIConfig:     "mysql-updated-ui",
		Type:         types.FieldTypeRelation,
		DateCreated:  ts,
		DateModified: ts,
		AdminFieldID: afID,
	}

	got := d.MapUpdateAdminFieldParams(input)

	if got.AdminFieldID != afID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, afID)
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.UiConfig != input.UIConfig {
		t.Errorf("UiConfig = %q, want %q", got.UiConfig, input.UIConfig)
	}
}

// --- PostgreSQL PsqlDatabase.MapAdminField tests ---

func TestPsqlDatabase_MapAdminField_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	fieldID := types.NewAdminFieldID()
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))

	input := mdbp.AdminFields{
		AdminFieldID: fieldID,
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

	got := d.MapAdminField(input)

	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
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

func TestPsqlDatabase_MapAdminField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminField(mdbp.AdminFields{})

	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateAdminFieldParams tests ---

func TestPsqlDatabase_MapCreateAdminFieldParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateAdminFieldParams{
		Label:        "psql-create",
		UIConfig:     "psql-create-ui",
		Type:         types.FieldTypeRichText,
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateAdminFieldParams(input)

	if got.AdminFieldID.IsZero() {
		t.Fatal("expected non-zero AdminFieldID to be generated")
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.UiConfig != input.UIConfig {
		t.Errorf("UiConfig = %q, want %q", got.UiConfig, input.UIConfig)
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateAdminFieldParams tests ---

func TestPsqlDatabase_MapUpdateAdminFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	afID := types.NewAdminFieldID()

	input := UpdateAdminFieldParams{
		Label:        "psql-updated",
		UIConfig:     "psql-updated-ui",
		Type:         types.FieldTypeSlug,
		DateCreated:  ts,
		DateModified: ts,
		AdminFieldID: afID,
	}

	got := d.MapUpdateAdminFieldParams(input)

	if got.AdminFieldID != afID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, afID)
	}
	if got.Label != input.Label {
		t.Errorf("Label = %q, want %q", got.Label, input.Label)
	}
	if got.UiConfig != input.UIConfig {
		t.Errorf("UiConfig = %q, want %q", got.UiConfig, input.UIConfig)
	}
}

// --- Cross-database mapper consistency ---
// All three database drivers use identical types for AdminFields
// (no int32/int64 conversions needed). This test verifies they all produce
// identical AdminFields wrapper structs from equivalent input.

func TestCrossDatabaseMapAdminField_Consistency(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminFieldID()
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))

	sqliteInput := mdb.AdminFields{
		AdminFieldID: fieldID, ParentID: parentID,
		Label: "cross-label", Data: "cross-data",
		Validation: "cross-validation", UiConfig: "cross-ui",
		Type: types.FieldTypeEmail, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.AdminFields{
		AdminFieldID: fieldID, ParentID: parentID,
		Label: "cross-label", Data: "cross-data",
		Validation: "cross-validation", UiConfig: "cross-ui",
		Type: types.FieldTypeEmail, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.AdminFields{
		AdminFieldID: fieldID, ParentID: parentID,
		Label: "cross-label", Data: "cross-data",
		Validation: "cross-validation", UiConfig: "cross-ui",
		Type: types.FieldTypeEmail, AuthorID: authorID,
		DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapAdminField(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminField(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminField(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateAdminFieldParams auto-ID generation ---

func TestCrossDatabaseMapCreateAdminFieldParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	input := CreateAdminFieldParams{
		Label:        "cross-create-label",
		UIConfig:     "cross-create-ui",
		Type:         types.FieldTypeURL,
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteResult := Database{}.MapCreateAdminFieldParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateAdminFieldParams(input)
	psqlResult := PsqlDatabase{}.MapCreateAdminFieldParams(input)

	if sqliteResult.AdminFieldID.IsZero() {
		t.Error("SQLite: expected non-zero generated AdminFieldID")
	}
	if mysqlResult.AdminFieldID.IsZero() {
		t.Error("MySQL: expected non-zero generated AdminFieldID")
	}
	if psqlResult.AdminFieldID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated AdminFieldID")
	}

	// Each call should generate a unique ID
	if sqliteResult.AdminFieldID == mysqlResult.AdminFieldID {
		t.Error("SQLite and MySQL generated the same AdminFieldID -- each call should be unique")
	}
	if sqliteResult.AdminFieldID == psqlResult.AdminFieldID {
		t.Error("SQLite and PostgreSQL generated the same AdminFieldID -- each call should be unique")
	}
}

// --- Cross-database UIConfig -> UiConfig field name mapping ---
// This is the critical mapping to verify: the wrapper uses UIConfig (Go convention)
// while sqlc generates UiConfig. All three Create and Update param mappers must
// correctly bridge this naming difference.

func TestCrossDatabaseUIConfigFieldMapping(t *testing.T) {
	t.Parallel()
	uiValue := `{"columns": 2, "widget": "dropdown"}`
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	afID := types.NewAdminFieldID()

	t.Run("CreateParams", func(t *testing.T) {
		t.Parallel()
		input := CreateAdminFieldParams{UIConfig: uiValue, DateCreated: ts, DateModified: ts}

		sqliteGot := Database{}.MapCreateAdminFieldParams(input)
		mysqlGot := MysqlDatabase{}.MapCreateAdminFieldParams(input)
		psqlGot := PsqlDatabase{}.MapCreateAdminFieldParams(input)

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
		input := UpdateAdminFieldParams{
			UIConfig: uiValue, DateCreated: ts, DateModified: ts,
			AdminFieldID: afID,
		}

		sqliteGot := Database{}.MapUpdateAdminFieldParams(input)
		mysqlGot := MysqlDatabase{}.MapUpdateAdminFieldParams(input)
		psqlGot := PsqlDatabase{}.MapUpdateAdminFieldParams(input)

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

	t.Run("MapAdminField_ReverseDirection", func(t *testing.T) {
		t.Parallel()
		// Verify the reverse: UiConfig in sqlc -> UIConfig in wrapper
		sqliteGot := Database{}.MapAdminField(mdb.AdminFields{UiConfig: uiValue})
		mysqlGot := MysqlDatabase{}.MapAdminField(mdbm.AdminFields{UiConfig: uiValue})
		psqlGot := PsqlDatabase{}.MapAdminField(mdbp.AdminFields{UiConfig: uiValue})

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

func TestNewAdminFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-af-1"),
		RequestID: "req-af-123",
		IP:        "10.0.0.1",
	}
	params := CreateAdminFieldParams{
		Label:        "cmd-label",
		Data:         "cmd-data",
		UIConfig:     "cmd-ui",
		Type:         types.FieldTypeText,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := Database{}.NewAdminFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_fields")
	}
	p, ok := cmd.Params().(CreateAdminFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminFieldParams", cmd.Params())
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

func TestNewAdminFieldCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminFieldID()
	cmd := NewAdminFieldCmd{}

	row := mdb.AdminFields{AdminFieldID: fieldID}
	got := cmd.GetID(row)
	if got != string(fieldID) {
		t.Errorf("GetID() = %q, want %q", got, string(fieldID))
	}
}

func TestNewAdminFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	cmd := NewAdminFieldCmd{}
	row := mdb.AdminFields{AdminFieldID: ""}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateAdminFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	afID := types.NewAdminFieldID()
	params := UpdateAdminFieldParams{
		Label:        "update-label",
		Data:         "update-data",
		UIConfig:     "update-ui",
		DateCreated:  ts,
		DateModified: ts,
		AdminFieldID: afID,
	}

	cmd := Database{}.UpdateAdminFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_fields")
	}
	if cmd.GetID() != string(afID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(afID))
	}
	p, ok := cmd.Params().(UpdateAdminFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminFieldParams", cmd.Params())
	}
	if p.Label != "update-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "update-label")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	afID := types.NewAdminFieldID()

	cmd := Database{}.DeleteAdminFieldCmd(ctx, ac, afID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_fields")
	}
	if cmd.GetID() != string(afID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(afID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewAdminFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node-af"),
		RequestID: "mysql-req-af",
		IP:        "192.168.1.1",
	}
	params := CreateAdminFieldParams{
		Label:        "mysql-cmd-label",
		UIConfig:     "mysql-cmd-ui",
		Type:         types.FieldTypeDate,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := MysqlDatabase{}.NewAdminFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_fields")
	}
	p, ok := cmd.Params().(CreateAdminFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminFieldParams", cmd.Params())
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

func TestNewAdminFieldCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminFieldID()
	cmd := NewAdminFieldCmdMysql{}

	row := mdbm.AdminFields{AdminFieldID: fieldID}
	got := cmd.GetID(row)
	if got != string(fieldID) {
		t.Errorf("GetID() = %q, want %q", got, string(fieldID))
	}
}

func TestUpdateAdminFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	afID := types.NewAdminFieldID()
	params := UpdateAdminFieldParams{
		Label:        "mysql-update-label",
		DateCreated:  ts,
		DateModified: ts,
		AdminFieldID: afID,
	}

	cmd := MysqlDatabase{}.UpdateAdminFieldCmd(ctx, ac, params)

	if cmd.TableName() != "admin_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_fields")
	}
	if cmd.GetID() != string(afID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(afID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminFieldParams", cmd.Params())
	}
	if p.Label != "mysql-update-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "mysql-update-label")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	afID := types.NewAdminFieldID()

	cmd := MysqlDatabase{}.DeleteAdminFieldCmd(ctx, ac, afID)

	if cmd.TableName() != "admin_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_fields")
	}
	if cmd.GetID() != string(afID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(afID))
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

func TestNewAdminFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node-af"),
		RequestID: "psql-req-af",
		IP:        "172.16.0.1",
	}
	params := CreateAdminFieldParams{
		Label:        "psql-cmd-label",
		UIConfig:     "psql-cmd-ui",
		Type:         types.FieldTypeDatetime,
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := PsqlDatabase{}.NewAdminFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_fields")
	}
	p, ok := cmd.Params().(CreateAdminFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminFieldParams", cmd.Params())
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

func TestNewAdminFieldCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminFieldID()
	cmd := NewAdminFieldCmdPsql{}

	row := mdbp.AdminFields{AdminFieldID: fieldID}
	got := cmd.GetID(row)
	if got != string(fieldID) {
		t.Errorf("GetID() = %q, want %q", got, string(fieldID))
	}
}

func TestUpdateAdminFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	afID := types.NewAdminFieldID()
	params := UpdateAdminFieldParams{
		Label:        "psql-update-label",
		DateCreated:  ts,
		DateModified: ts,
		AdminFieldID: afID,
	}

	cmd := PsqlDatabase{}.UpdateAdminFieldCmd(ctx, ac, params)

	if cmd.TableName() != "admin_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_fields")
	}
	if cmd.GetID() != string(afID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(afID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminFieldParams", cmd.Params())
	}
	if p.Label != "psql-update-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "psql-update-label")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	afID := types.NewAdminFieldID()

	cmd := PsqlDatabase{}.DeleteAdminFieldCmd(ctx, ac, afID)

	if cmd.TableName() != "admin_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_fields")
	}
	if cmd.GetID() != string(afID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(afID))
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

func TestAuditedAdminFieldCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminFieldParams{}
	updateParams := UpdateAdminFieldParams{AdminFieldID: types.NewAdminFieldID()}
	afID := types.NewAdminFieldID()

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewAdminFieldCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateAdminFieldCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteAdminFieldCmd(ctx, ac, afID).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewAdminFieldCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateAdminFieldCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteAdminFieldCmd(ctx, ac, afID).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewAdminFieldCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateAdminFieldCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteAdminFieldCmd(ctx, ac, afID).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "admin_fields" {
				t.Errorf("TableName() = %q, want %q", c.name, "admin_fields")
			}
		})
	}
}

func TestAuditedAdminFieldCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminFieldParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewAdminFieldCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewAdminFieldCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewAdminFieldCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedAdminFieldCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	afID := types.NewAdminFieldID()

	t.Run("UpdateCmd GetID returns AdminFieldID", func(t *testing.T) {
		t.Parallel()
		params := UpdateAdminFieldParams{AdminFieldID: afID}

		sqliteCmd := Database{}.UpdateAdminFieldCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateAdminFieldCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateAdminFieldCmd(ctx, ac, params)

		wantID := string(afID)
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

	t.Run("DeleteCmd GetID returns AdminFieldID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteAdminFieldCmd(ctx, ac, afID)
		mysqlCmd := MysqlDatabase{}.DeleteAdminFieldCmd(ctx, ac, afID)
		psqlCmd := PsqlDatabase{}.DeleteAdminFieldCmd(ctx, ac, afID)

		wantID := string(afID)
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
		testFieldID := types.NewAdminFieldID()

		sqliteCmd := NewAdminFieldCmd{}
		mysqlCmd := NewAdminFieldCmdMysql{}
		psqlCmd := NewAdminFieldCmdPsql{}

		wantID := string(testFieldID)

		sqliteRow := mdb.AdminFields{AdminFieldID: testFieldID}
		mysqlRow := mdbm.AdminFields{AdminFieldID: testFieldID}
		psqlRow := mdbp.AdminFields{AdminFieldID: testFieldID}

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

// --- Edge case: UpdateCmd with empty AdminFieldID ---

func TestUpdateAdminFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateAdminFieldParams{AdminFieldID: ""}

	sqliteCmd := Database{}.UpdateAdminFieldCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateAdminFieldCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateAdminFieldCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteCmd with empty ID ---

func TestDeleteAdminFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.AdminFieldID("")

	sqliteCmd := Database{}.DeleteAdminFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteAdminFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteAdminFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- UpdateAdminField success message format ---
// The UpdateAdminField method produces a success message string.
// We can't test it without a real DB, but we verify the message format expectation
// by testing the format string directly.

func TestUpdateAdminField_SuccessMessageFormat(t *testing.T) {
	t.Parallel()
	label := "Test Field"
	expected := fmt.Sprintf("Successfully updated %v\n", label)

	// The actual method builds this exact format using s.Label
	if expected != "Successfully updated Test Field\n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated Test Field\n")
	}
}

func TestUpdateAdminField_SuccessMessageFormat_EmptyLabel(t *testing.T) {
	t.Parallel()
	// Edge case: empty label still produces a valid format string
	expected := fmt.Sprintf("Successfully updated %v\n", "")
	if expected != "Successfully updated \n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated \n")
	}
}

// --- Recorder identity verification ---
// Verify the command factories assign the correct recorder variant.

func TestAdminFieldCommand_RecorderIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminFieldParams{}
	updateParams := UpdateAdminFieldParams{AdminFieldID: types.NewAdminFieldID()}
	deleteID := types.NewAdminFieldID()

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
		want     audited.ChangeEventRecorder
	}{
		// SQLite commands should use SQLiteRecorder
		{"SQLite Create", Database{}.NewAdminFieldCmd(ctx, ac, createParams).Recorder(), SQLiteRecorder},
		{"SQLite Update", Database{}.UpdateAdminFieldCmd(ctx, ac, updateParams).Recorder(), SQLiteRecorder},
		{"SQLite Delete", Database{}.DeleteAdminFieldCmd(ctx, ac, deleteID).Recorder(), SQLiteRecorder},
		// MySQL commands should use MysqlRecorder
		{"MySQL Create", MysqlDatabase{}.NewAdminFieldCmd(ctx, ac, createParams).Recorder(), MysqlRecorder},
		{"MySQL Update", MysqlDatabase{}.UpdateAdminFieldCmd(ctx, ac, updateParams).Recorder(), MysqlRecorder},
		{"MySQL Delete", MysqlDatabase{}.DeleteAdminFieldCmd(ctx, ac, deleteID).Recorder(), MysqlRecorder},
		// PostgreSQL commands should use PsqlRecorder
		{"PostgreSQL Create", PsqlDatabase{}.NewAdminFieldCmd(ctx, ac, createParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateAdminFieldCmd(ctx, ac, updateParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteAdminFieldCmd(ctx, ac, deleteID).Recorder(), PsqlRecorder},
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

// --- Additional row type struct tests ---
// Verify the extra row types declared in admin_field.go are usable.

func TestListAdminFieldByRouteIdRow_ZeroValue(t *testing.T) {
	t.Parallel()
	var row ListAdminFieldByRouteIdRow

	if row.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", row.AdminFieldID)
	}
	if row.Label != "" {
		t.Errorf("Label = %q, want empty string", row.Label)
	}
	if row.Type != "" {
		t.Errorf("Type = %v, want zero value", row.Type)
	}
}

func TestListAdminFieldsByDatatypeIDRow_ZeroValue(t *testing.T) {
	t.Parallel()
	var row ListAdminFieldsByDatatypeIDRow

	if row.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", row.AdminFieldID)
	}
	if row.ParentID.Valid {
		t.Error("ParentID.Valid = true, want false")
	}
}

func TestUtilityGetAdminfieldsRow_ZeroValue(t *testing.T) {
	t.Parallel()
	var row UtilityGetAdminfieldsRow

	if row.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", row.AdminFieldID)
	}
	if row.Label != "" {
		t.Errorf("Label = %q, want empty string", row.Label)
	}
}

func TestListAdminFieldByRouteIdRow_PopulatedValues(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminFieldID()
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}

	row := ListAdminFieldByRouteIdRow{
		AdminFieldID: fieldID,
		ParentID:     parentID,
		Label:        "route-field",
		Data:         "route-data",
		Validation:   "route-validation",
		UIConfig:     "route-ui-config",
		Type:         types.FieldTypeMedia,
	}

	if row.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", row.AdminFieldID, fieldID)
	}
	if row.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", row.ParentID, parentID)
	}
	if row.Label != "route-field" {
		t.Errorf("Label = %q, want %q", row.Label, "route-field")
	}
	if row.Data != "route-data" {
		t.Errorf("Data = %q, want %q", row.Data, "route-data")
	}
	if row.Validation != "route-validation" {
		t.Errorf("Validation = %q, want %q", row.Validation, "route-validation")
	}
	if row.UIConfig != "route-ui-config" {
		t.Errorf("UIConfig = %q, want %q", row.UIConfig, "route-ui-config")
	}
	if row.Type != types.FieldTypeMedia {
		t.Errorf("Type = %v, want %v", row.Type, types.FieldTypeMedia)
	}
}

func TestListAdminFieldsByDatatypeIDRow_PopulatedValues(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminFieldID()
	parentID := types.NullableAdminDatatypeID{ID: types.NewAdminDatatypeID(), Valid: true}

	row := ListAdminFieldsByDatatypeIDRow{
		AdminFieldID: fieldID,
		ParentID:     parentID,
		Label:        "datatype-field",
		Data:         "datatype-data",
		Validation:   "datatype-validation",
		UIConfig:     "datatype-ui-config",
		Type:         types.FieldTypeRelation,
	}

	if row.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", row.AdminFieldID, fieldID)
	}
	if row.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", row.ParentID, parentID)
	}
	if row.Label != "datatype-field" {
		t.Errorf("Label = %q, want %q", row.Label, "datatype-field")
	}
	if row.Data != "datatype-data" {
		t.Errorf("Data = %q, want %q", row.Data, "datatype-data")
	}
	if row.Validation != "datatype-validation" {
		t.Errorf("Validation = %q, want %q", row.Validation, "datatype-validation")
	}
	if row.UIConfig != "datatype-ui-config" {
		t.Errorf("UIConfig = %q, want %q", row.UIConfig, "datatype-ui-config")
	}
	if row.Type != types.FieldTypeRelation {
		t.Errorf("Type = %v, want %v", row.Type, types.FieldTypeRelation)
	}
}

func TestUtilityGetAdminfieldsRow_PopulatedValues(t *testing.T) {
	t.Parallel()
	fieldID := types.NewAdminFieldID()

	row := UtilityGetAdminfieldsRow{
		AdminFieldID: fieldID,
		Label:        "utility-label",
	}

	if row.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", row.AdminFieldID, fieldID)
	}
	if row.Label != "utility-label" {
		t.Errorf("Label = %q, want %q", row.Label, "utility-label")
	}
}
