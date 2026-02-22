// White-box tests for field_type_gen.go: wrapper structs, mapper methods
// across all three database drivers (SQLite, MySQL, PostgreSQL), string mapping
// for TUI display, and audited command struct accessors.
//
// White-box access is needed because:
//   - Audited command structs have unexported fields (ctx, auditCtx, params, conn)
//     that can only be constructed through the Database/MysqlDatabase/PsqlDatabase
//     factory methods, which require access to the package internals.
//   - We verify that the SQLiteRecorder, MysqlRecorder, and PsqlRecorder package-level
//     vars are correctly wired into command constructors.
package db

import (
	"context"
	"testing"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

func fieldTypeTestFixture() (FieldTypes, types.FieldTypeID, string, string) {
	id := types.NewFieldTypeID()
	typ := "text"
	label := "Text Input"
	ft := FieldTypes{
		FieldTypeID: id,
		Type:        typ,
		Label:       label,
	}
	return ft, id, typ, label
}

// --- MapStringFieldType tests ---

func TestMapStringFieldType_AllFields(t *testing.T) {
	t.Parallel()
	ft, id, typ, label := fieldTypeTestFixture()

	got := MapStringFieldType(ft)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"FieldTypeID", got.FieldTypeID, id.String()},
		{"Type", got.Type, typ},
		{"Label", got.Label, label},
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

func TestMapStringFieldType_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringFieldType(FieldTypes{})

	if got.FieldTypeID != "" {
		t.Errorf("FieldTypeID = %q, want empty string", got.FieldTypeID)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapFieldType tests ---

func TestDatabase_MapFieldType_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	id := types.NewFieldTypeID()

	input := mdb.FieldTypes{
		FieldTypeID: id,
		Type:        "number",
		Label:       "Number",
	}

	got := d.MapFieldType(input)

	if got.FieldTypeID != id {
		t.Errorf("FieldTypeID = %v, want %v", got.FieldTypeID, id)
	}
	if got.Type != "number" {
		t.Errorf("Type = %q, want %q", got.Type, "number")
	}
	if got.Label != "Number" {
		t.Errorf("Label = %q, want %q", got.Label, "Number")
	}
}

func TestDatabase_MapFieldType_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapFieldType(mdb.FieldTypes{})

	if got.FieldTypeID != "" {
		t.Errorf("FieldTypeID = %v, want zero value", got.FieldTypeID)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapCreateFieldTypeParams tests ---

func TestDatabase_MapCreateFieldTypeParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := CreateFieldTypeParams{
		Type:  "date",
		Label: "Date",
	}

	got := d.MapCreateFieldTypeParams(input)

	if got.FieldTypeID.IsZero() {
		t.Fatal("expected non-zero FieldTypeID to be generated")
	}
	if got.Type != "date" {
		t.Errorf("Type = %q, want %q", got.Type, "date")
	}
	if got.Label != "Date" {
		t.Errorf("Label = %q, want %q", got.Label, "Date")
	}
}

func TestDatabase_MapCreateFieldTypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}

	got1 := d.MapCreateFieldTypeParams(CreateFieldTypeParams{})
	got2 := d.MapCreateFieldTypeParams(CreateFieldTypeParams{})

	if got1.FieldTypeID == got2.FieldTypeID {
		t.Error("two calls produced identical FieldTypeIDs")
	}
}

func TestDatabase_MapCreateFieldTypeParams_ZeroInput(t *testing.T) {
	t.Parallel()
	d := Database{}

	got := d.MapCreateFieldTypeParams(CreateFieldTypeParams{})

	if got.FieldTypeID.IsZero() {
		t.Fatal("expected non-zero FieldTypeID even with zero input")
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapUpdateFieldTypeParams tests ---

func TestDatabase_MapUpdateFieldTypeParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	id := types.NewFieldTypeID()

	input := UpdateFieldTypeParams{
		Type:        "media",
		Label:       "Media",
		FieldTypeID: id,
	}

	got := d.MapUpdateFieldTypeParams(input)

	if got.Type != "media" {
		t.Errorf("Type = %q, want %q", got.Type, "media")
	}
	if got.Label != "Media" {
		t.Errorf("Label = %q, want %q", got.Label, "Media")
	}
	if got.FieldTypeID != id {
		t.Errorf("FieldTypeID = %v, want %v", got.FieldTypeID, id)
	}
}

func TestDatabase_MapUpdateFieldTypeParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateFieldTypeParams(UpdateFieldTypeParams{})

	if got.FieldTypeID != "" {
		t.Errorf("FieldTypeID = %v, want zero value", got.FieldTypeID)
	}
}

// --- MySQL MysqlDatabase.MapFieldType tests ---

func TestMysqlDatabase_MapFieldType_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	id := types.NewFieldTypeID()

	input := mdbm.FieldTypes{
		FieldTypeID: id,
		Type:        "boolean",
		Label:       "Boolean",
	}

	got := d.MapFieldType(input)

	if got.FieldTypeID != id {
		t.Errorf("FieldTypeID = %v, want %v", got.FieldTypeID, id)
	}
	if got.Type != "boolean" {
		t.Errorf("Type = %q, want %q", got.Type, "boolean")
	}
	if got.Label != "Boolean" {
		t.Errorf("Label = %q, want %q", got.Label, "Boolean")
	}
}

func TestMysqlDatabase_MapFieldType_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapFieldType(mdbm.FieldTypes{})

	if got.FieldTypeID != "" {
		t.Errorf("FieldTypeID = %v, want zero value", got.FieldTypeID)
	}
}

// --- MySQL MysqlDatabase.MapCreateFieldTypeParams tests ---

func TestMysqlDatabase_MapCreateFieldTypeParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := CreateFieldTypeParams{
		Type:  "select",
		Label: "Select",
	}

	got := d.MapCreateFieldTypeParams(input)

	if got.FieldTypeID.IsZero() {
		t.Fatal("expected non-zero FieldTypeID to be generated")
	}
	if got.Type != "select" {
		t.Errorf("Type = %q, want %q", got.Type, "select")
	}
	if got.Label != "Select" {
		t.Errorf("Label = %q, want %q", got.Label, "Select")
	}
}

func TestMysqlDatabase_MapCreateFieldTypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	got1 := d.MapCreateFieldTypeParams(CreateFieldTypeParams{})
	got2 := d.MapCreateFieldTypeParams(CreateFieldTypeParams{})

	if got1.FieldTypeID == got2.FieldTypeID {
		t.Error("two calls produced identical FieldTypeIDs")
	}
}

// --- MySQL MysqlDatabase.MapUpdateFieldTypeParams tests ---

func TestMysqlDatabase_MapUpdateFieldTypeParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	id := types.NewFieldTypeID()

	input := UpdateFieldTypeParams{
		Type:        "json",
		Label:       "JSON",
		FieldTypeID: id,
	}

	got := d.MapUpdateFieldTypeParams(input)

	if got.FieldTypeID != id {
		t.Errorf("FieldTypeID = %v, want %v", got.FieldTypeID, id)
	}
	if got.Type != "json" {
		t.Errorf("Type = %q, want %q", got.Type, "json")
	}
	if got.Label != "JSON" {
		t.Errorf("Label = %q, want %q", got.Label, "JSON")
	}
}

// --- PostgreSQL PsqlDatabase.MapFieldType tests ---

func TestPsqlDatabase_MapFieldType_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	id := types.NewFieldTypeID()

	input := mdbp.FieldTypes{
		FieldTypeID: id,
		Type:        "richtext",
		Label:       "Rich Text",
	}

	got := d.MapFieldType(input)

	if got.FieldTypeID != id {
		t.Errorf("FieldTypeID = %v, want %v", got.FieldTypeID, id)
	}
	if got.Type != "richtext" {
		t.Errorf("Type = %q, want %q", got.Type, "richtext")
	}
	if got.Label != "Rich Text" {
		t.Errorf("Label = %q, want %q", got.Label, "Rich Text")
	}
}

func TestPsqlDatabase_MapFieldType_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapFieldType(mdbp.FieldTypes{})

	if got.FieldTypeID != "" {
		t.Errorf("FieldTypeID = %v, want zero value", got.FieldTypeID)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateFieldTypeParams tests ---

func TestPsqlDatabase_MapCreateFieldTypeParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := CreateFieldTypeParams{
		Type:  "slug",
		Label: "Slug",
	}

	got := d.MapCreateFieldTypeParams(input)

	if got.FieldTypeID.IsZero() {
		t.Fatal("expected non-zero FieldTypeID to be generated")
	}
	if got.Type != "slug" {
		t.Errorf("Type = %q, want %q", got.Type, "slug")
	}
	if got.Label != "Slug" {
		t.Errorf("Label = %q, want %q", got.Label, "Slug")
	}
}

func TestPsqlDatabase_MapCreateFieldTypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	got1 := d.MapCreateFieldTypeParams(CreateFieldTypeParams{})
	got2 := d.MapCreateFieldTypeParams(CreateFieldTypeParams{})

	if got1.FieldTypeID == got2.FieldTypeID {
		t.Error("two calls produced identical FieldTypeIDs")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateFieldTypeParams tests ---

func TestPsqlDatabase_MapUpdateFieldTypeParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	id := types.NewFieldTypeID()

	input := UpdateFieldTypeParams{
		Type:        "email",
		Label:       "Email",
		FieldTypeID: id,
	}

	got := d.MapUpdateFieldTypeParams(input)

	if got.FieldTypeID != id {
		t.Errorf("FieldTypeID = %v, want %v", got.FieldTypeID, id)
	}
	if got.Type != "email" {
		t.Errorf("Type = %q, want %q", got.Type, "email")
	}
	if got.Label != "Email" {
		t.Errorf("Label = %q, want %q", got.Label, "Email")
	}
}

// --- Cross-database mapper consistency ---

func TestCrossDatabaseMapFieldType_Consistency(t *testing.T) {
	t.Parallel()
	id := types.NewFieldTypeID()

	sqliteInput := mdb.FieldTypes{
		FieldTypeID: id,
		Type:        "url",
		Label:       "URL",
	}
	mysqlInput := mdbm.FieldTypes{
		FieldTypeID: id,
		Type:        "url",
		Label:       "URL",
	}
	psqlInput := mdbp.FieldTypes{
		FieldTypeID: id,
		Type:        "url",
		Label:       "URL",
	}

	sqliteResult := Database{}.MapFieldType(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapFieldType(mysqlInput)
	psqlResult := PsqlDatabase{}.MapFieldType(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateFieldTypeParams - ID generation ---

func TestCrossDatabaseMapCreateFieldTypeParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()

	input := CreateFieldTypeParams{
		Type:  "relation",
		Label: "Relation",
	}

	sqliteResult := Database{}.MapCreateFieldTypeParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateFieldTypeParams(input)
	psqlResult := PsqlDatabase{}.MapCreateFieldTypeParams(input)

	if sqliteResult.FieldTypeID.IsZero() {
		t.Error("SQLite: expected non-zero generated FieldTypeID")
	}
	if mysqlResult.FieldTypeID.IsZero() {
		t.Error("MySQL: expected non-zero generated FieldTypeID")
	}
	if psqlResult.FieldTypeID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated FieldTypeID")
	}

	// Each call should generate a unique ID
	if sqliteResult.FieldTypeID == mysqlResult.FieldTypeID {
		t.Error("SQLite and MySQL generated the same FieldTypeID -- each call should be unique")
	}
	if sqliteResult.FieldTypeID == psqlResult.FieldTypeID {
		t.Error("SQLite and PostgreSQL generated the same FieldTypeID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewFieldTypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-ft-001",
		IP:        "10.0.0.1",
	}
	params := CreateFieldTypeParams{
		Type:  "text",
		Label: "Text Input",
	}

	cmd := Database{}.NewFieldTypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "field_types")
	}
	p, ok := cmd.Params().(CreateFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateFieldTypeParams", cmd.Params())
	}
	if p.Type != "text" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "text")
	}
	if p.Label != "Text Input" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "Text Input")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewFieldTypeCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	id := types.NewFieldTypeID()
	cmd := NewFieldTypeCmd{}

	row := mdb.FieldTypes{FieldTypeID: id}
	got := cmd.GetID(row)
	if got != string(id) {
		t.Errorf("GetID() = %q, want %q", got, string(id))
	}
}

func TestNewFieldTypeCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewFieldTypeCmd{}

	row := mdb.FieldTypes{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateFieldTypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewFieldTypeID()
	params := UpdateFieldTypeParams{
		Type:        "boolean",
		Label:       "Boolean",
		FieldTypeID: id,
	}

	cmd := Database{}.UpdateFieldTypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "field_types")
	}
	if cmd.GetID() != string(id) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(id))
	}
	p, ok := cmd.Params().(UpdateFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateFieldTypeParams", cmd.Params())
	}
	if p.Type != "boolean" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "boolean")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteFieldTypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewFieldTypeID()

	cmd := Database{}.DeleteFieldTypeCmd(ctx, ac, id)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "field_types")
	}
	if cmd.GetID() != string(id) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(id))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewFieldTypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-ft-001",
		IP:        "192.168.1.1",
	}
	params := CreateFieldTypeParams{
		Type:  "textarea",
		Label: "Text Area",
	}

	cmd := MysqlDatabase{}.NewFieldTypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "field_types")
	}
	p, ok := cmd.Params().(CreateFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateFieldTypeParams", cmd.Params())
	}
	if p.Type != "textarea" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "textarea")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewFieldTypeCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	id := types.NewFieldTypeID()
	cmd := NewFieldTypeCmdMysql{}

	row := mdbm.FieldTypes{FieldTypeID: id}
	got := cmd.GetID(row)
	if got != string(id) {
		t.Errorf("GetID() = %q, want %q", got, string(id))
	}
}

func TestUpdateFieldTypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewFieldTypeID()
	params := UpdateFieldTypeParams{
		Type:        "date",
		Label:       "Date",
		FieldTypeID: id,
	}

	cmd := MysqlDatabase{}.UpdateFieldTypeCmd(ctx, ac, params)

	if cmd.TableName() != "field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "field_types")
	}
	if cmd.GetID() != string(id) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(id))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateFieldTypeParams", cmd.Params())
	}
	if p.Type != "date" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "date")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteFieldTypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewFieldTypeID()

	cmd := MysqlDatabase{}.DeleteFieldTypeCmd(ctx, ac, id)

	if cmd.TableName() != "field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "field_types")
	}
	if cmd.GetID() != string(id) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(id))
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

func TestNewFieldTypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-ft-001",
		IP:        "172.16.0.1",
	}
	params := CreateFieldTypeParams{
		Type:  "relation",
		Label: "Relation",
	}

	cmd := PsqlDatabase{}.NewFieldTypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "field_types")
	}
	p, ok := cmd.Params().(CreateFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateFieldTypeParams", cmd.Params())
	}
	if p.Type != "relation" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "relation")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewFieldTypeCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	id := types.NewFieldTypeID()
	cmd := NewFieldTypeCmdPsql{}

	row := mdbp.FieldTypes{FieldTypeID: id}
	got := cmd.GetID(row)
	if got != string(id) {
		t.Errorf("GetID() = %q, want %q", got, string(id))
	}
}

func TestUpdateFieldTypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewFieldTypeID()
	params := UpdateFieldTypeParams{
		Type:        "datetime",
		Label:       "Date & Time",
		FieldTypeID: id,
	}

	cmd := PsqlDatabase{}.UpdateFieldTypeCmd(ctx, ac, params)

	if cmd.TableName() != "field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "field_types")
	}
	if cmd.GetID() != string(id) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(id))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateFieldTypeParams", cmd.Params())
	}
	if p.Type != "datetime" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "datetime")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteFieldTypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewFieldTypeID()

	cmd := PsqlDatabase{}.DeleteFieldTypeCmd(ctx, ac, id)

	if cmd.TableName() != "field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "field_types")
	}
	if cmd.GetID() != string(id) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(id))
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

func TestAuditedFieldTypeCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateFieldTypeParams{}
	updateParams := UpdateFieldTypeParams{
		FieldTypeID: types.NewFieldTypeID(),
	}
	id := types.NewFieldTypeID()

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewFieldTypeCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateFieldTypeCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteFieldTypeCmd(ctx, ac, id).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewFieldTypeCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateFieldTypeCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteFieldTypeCmd(ctx, ac, id).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewFieldTypeCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateFieldTypeCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteFieldTypeCmd(ctx, ac, id).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "field_types" {
				t.Errorf("TableName() = %q, want %q", c.name, "field_types")
			}
		})
	}
}

func TestAuditedFieldTypeCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateFieldTypeParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewFieldTypeCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewFieldTypeCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewFieldTypeCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedFieldTypeCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	id := types.NewFieldTypeID()

	t.Run("UpdateCmd GetID returns field type ID", func(t *testing.T) {
		t.Parallel()
		params := UpdateFieldTypeParams{
			FieldTypeID: id,
		}

		sqliteCmd := Database{}.UpdateFieldTypeCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateFieldTypeCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateFieldTypeCmd(ctx, ac, params)

		wantID := string(id)
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

	t.Run("DeleteCmd GetID returns field type ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteFieldTypeCmd(ctx, ac, id)
		mysqlCmd := MysqlDatabase{}.DeleteFieldTypeCmd(ctx, ac, id)
		psqlCmd := PsqlDatabase{}.DeleteFieldTypeCmd(ctx, ac, id)

		wantID := string(id)
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
		testID := types.NewFieldTypeID()
		wantID := string(testID)

		sqliteCmd := NewFieldTypeCmd{}
		mysqlCmd := NewFieldTypeCmdMysql{}
		psqlCmd := NewFieldTypeCmdPsql{}

		sqliteRow := mdb.FieldTypes{FieldTypeID: testID}
		mysqlRow := mdbm.FieldTypes{FieldTypeID: testID}
		psqlRow := mdbp.FieldTypes{FieldTypeID: testID}

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

// --- Edge cases: empty IDs ---

func TestUpdateFieldTypeCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateFieldTypeParams{
		FieldTypeID: "",
	}

	sqliteCmd := Database{}.UpdateFieldTypeCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateFieldTypeCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateFieldTypeCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteFieldTypeCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.FieldTypeID("")

	sqliteCmd := Database{}.DeleteFieldTypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteFieldTypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteFieldTypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.FieldTypes]  = NewFieldTypeCmd{}
	_ audited.UpdateCommand[mdb.FieldTypes]  = UpdateFieldTypeCmd{}
	_ audited.DeleteCommand[mdb.FieldTypes]  = DeleteFieldTypeCmd{}
	_ audited.CreateCommand[mdbm.FieldTypes] = NewFieldTypeCmdMysql{}
	_ audited.UpdateCommand[mdbm.FieldTypes] = UpdateFieldTypeCmdMysql{}
	_ audited.DeleteCommand[mdbm.FieldTypes] = DeleteFieldTypeCmdMysql{}
	_ audited.CreateCommand[mdbp.FieldTypes] = NewFieldTypeCmdPsql{}
	_ audited.UpdateCommand[mdbp.FieldTypes] = UpdateFieldTypeCmdPsql{}
	_ audited.DeleteCommand[mdbp.FieldTypes] = DeleteFieldTypeCmdPsql{}
)
