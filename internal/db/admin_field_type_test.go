// White-box tests for admin_field_type_gen.go: wrapper structs, mapper methods
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

func adminFieldTypeTestFixture() (AdminFieldTypes, types.AdminFieldTypeID, string, string) {
	id := types.NewAdminFieldTypeID()
	typ := "text"
	label := "Text Input"
	ft := AdminFieldTypes{
		AdminFieldTypeID: id,
		Type:             typ,
		Label:            label,
	}
	return ft, id, typ, label
}

// --- MapStringAdminFieldType tests ---

func TestMapStringAdminFieldType_AllFields(t *testing.T) {
	t.Parallel()
	ft, id, typ, label := adminFieldTypeTestFixture()

	got := MapStringAdminFieldType(ft)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"AdminFieldTypeID", got.AdminFieldTypeID, id.String()},
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

func TestMapStringAdminFieldType_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringAdminFieldType(AdminFieldTypes{})

	if got.AdminFieldTypeID != "" {
		t.Errorf("AdminFieldTypeID = %q, want empty string", got.AdminFieldTypeID)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapAdminFieldType tests ---

func TestDatabase_MapAdminFieldType_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	id := types.NewAdminFieldTypeID()

	input := mdb.AdminFieldTypes{
		AdminFieldTypeID: id,
		Type:             "number",
		Label:            "Number",
	}

	got := d.MapAdminFieldType(input)

	if got.AdminFieldTypeID != id {
		t.Errorf("AdminFieldTypeID = %v, want %v", got.AdminFieldTypeID, id)
	}
	if got.Type != "number" {
		t.Errorf("Type = %q, want %q", got.Type, "number")
	}
	if got.Label != "Number" {
		t.Errorf("Label = %q, want %q", got.Label, "Number")
	}
}

func TestDatabase_MapAdminFieldType_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminFieldType(mdb.AdminFieldTypes{})

	if got.AdminFieldTypeID != "" {
		t.Errorf("AdminFieldTypeID = %v, want zero value", got.AdminFieldTypeID)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapCreateAdminFieldTypeParams tests ---

func TestDatabase_MapCreateAdminFieldTypeParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := CreateAdminFieldTypeParams{
		Type:  "date",
		Label: "Date",
	}

	got := d.MapCreateAdminFieldTypeParams(input)

	if got.AdminFieldTypeID.IsZero() {
		t.Fatal("expected non-zero AdminFieldTypeID to be generated")
	}
	if got.Type != "date" {
		t.Errorf("Type = %q, want %q", got.Type, "date")
	}
	if got.Label != "Date" {
		t.Errorf("Label = %q, want %q", got.Label, "Date")
	}
}

func TestDatabase_MapCreateAdminFieldTypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}

	got1 := d.MapCreateAdminFieldTypeParams(CreateAdminFieldTypeParams{})
	got2 := d.MapCreateAdminFieldTypeParams(CreateAdminFieldTypeParams{})

	if got1.AdminFieldTypeID == got2.AdminFieldTypeID {
		t.Error("two calls produced identical AdminFieldTypeIDs")
	}
}

func TestDatabase_MapCreateAdminFieldTypeParams_ZeroInput(t *testing.T) {
	t.Parallel()
	d := Database{}

	got := d.MapCreateAdminFieldTypeParams(CreateAdminFieldTypeParams{})

	if got.AdminFieldTypeID.IsZero() {
		t.Fatal("expected non-zero AdminFieldTypeID even with zero input")
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty string", got.Type)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapUpdateAdminFieldTypeParams tests ---

func TestDatabase_MapUpdateAdminFieldTypeParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	id := types.NewAdminFieldTypeID()

	input := UpdateAdminFieldTypeParams{
		Type:             "media",
		Label:            "Media",
		AdminFieldTypeID: id,
	}

	got := d.MapUpdateAdminFieldTypeParams(input)

	if got.Type != "media" {
		t.Errorf("Type = %q, want %q", got.Type, "media")
	}
	if got.Label != "Media" {
		t.Errorf("Label = %q, want %q", got.Label, "Media")
	}
	if got.AdminFieldTypeID != id {
		t.Errorf("AdminFieldTypeID = %v, want %v", got.AdminFieldTypeID, id)
	}
}

func TestDatabase_MapUpdateAdminFieldTypeParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateAdminFieldTypeParams(UpdateAdminFieldTypeParams{})

	if got.AdminFieldTypeID != "" {
		t.Errorf("AdminFieldTypeID = %v, want zero value", got.AdminFieldTypeID)
	}
}

// --- MySQL MysqlDatabase.MapAdminFieldType tests ---

func TestMysqlDatabase_MapAdminFieldType_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	id := types.NewAdminFieldTypeID()

	input := mdbm.AdminFieldTypes{
		AdminFieldTypeID: id,
		Type:             "boolean",
		Label:            "Boolean",
	}

	got := d.MapAdminFieldType(input)

	if got.AdminFieldTypeID != id {
		t.Errorf("AdminFieldTypeID = %v, want %v", got.AdminFieldTypeID, id)
	}
	if got.Type != "boolean" {
		t.Errorf("Type = %q, want %q", got.Type, "boolean")
	}
	if got.Label != "Boolean" {
		t.Errorf("Label = %q, want %q", got.Label, "Boolean")
	}
}

func TestMysqlDatabase_MapAdminFieldType_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminFieldType(mdbm.AdminFieldTypes{})

	if got.AdminFieldTypeID != "" {
		t.Errorf("AdminFieldTypeID = %v, want zero value", got.AdminFieldTypeID)
	}
}

// --- MySQL MysqlDatabase.MapCreateAdminFieldTypeParams tests ---

func TestMysqlDatabase_MapCreateAdminFieldTypeParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := CreateAdminFieldTypeParams{
		Type:  "select",
		Label: "Select",
	}

	got := d.MapCreateAdminFieldTypeParams(input)

	if got.AdminFieldTypeID.IsZero() {
		t.Fatal("expected non-zero AdminFieldTypeID to be generated")
	}
	if got.Type != "select" {
		t.Errorf("Type = %q, want %q", got.Type, "select")
	}
	if got.Label != "Select" {
		t.Errorf("Label = %q, want %q", got.Label, "Select")
	}
}

func TestMysqlDatabase_MapCreateAdminFieldTypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	got1 := d.MapCreateAdminFieldTypeParams(CreateAdminFieldTypeParams{})
	got2 := d.MapCreateAdminFieldTypeParams(CreateAdminFieldTypeParams{})

	if got1.AdminFieldTypeID == got2.AdminFieldTypeID {
		t.Error("two calls produced identical AdminFieldTypeIDs")
	}
}

// --- MySQL MysqlDatabase.MapUpdateAdminFieldTypeParams tests ---

func TestMysqlDatabase_MapUpdateAdminFieldTypeParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	id := types.NewAdminFieldTypeID()

	input := UpdateAdminFieldTypeParams{
		Type:             "json",
		Label:            "JSON",
		AdminFieldTypeID: id,
	}

	got := d.MapUpdateAdminFieldTypeParams(input)

	if got.AdminFieldTypeID != id {
		t.Errorf("AdminFieldTypeID = %v, want %v", got.AdminFieldTypeID, id)
	}
	if got.Type != "json" {
		t.Errorf("Type = %q, want %q", got.Type, "json")
	}
	if got.Label != "JSON" {
		t.Errorf("Label = %q, want %q", got.Label, "JSON")
	}
}

// --- PostgreSQL PsqlDatabase.MapAdminFieldType tests ---

func TestPsqlDatabase_MapAdminFieldType_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	id := types.NewAdminFieldTypeID()

	input := mdbp.AdminFieldTypes{
		AdminFieldTypeID: id,
		Type:             "richtext",
		Label:            "Rich Text",
	}

	got := d.MapAdminFieldType(input)

	if got.AdminFieldTypeID != id {
		t.Errorf("AdminFieldTypeID = %v, want %v", got.AdminFieldTypeID, id)
	}
	if got.Type != "richtext" {
		t.Errorf("Type = %q, want %q", got.Type, "richtext")
	}
	if got.Label != "Rich Text" {
		t.Errorf("Label = %q, want %q", got.Label, "Rich Text")
	}
}

func TestPsqlDatabase_MapAdminFieldType_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminFieldType(mdbp.AdminFieldTypes{})

	if got.AdminFieldTypeID != "" {
		t.Errorf("AdminFieldTypeID = %v, want zero value", got.AdminFieldTypeID)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateAdminFieldTypeParams tests ---

func TestPsqlDatabase_MapCreateAdminFieldTypeParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := CreateAdminFieldTypeParams{
		Type:  "slug",
		Label: "Slug",
	}

	got := d.MapCreateAdminFieldTypeParams(input)

	if got.AdminFieldTypeID.IsZero() {
		t.Fatal("expected non-zero AdminFieldTypeID to be generated")
	}
	if got.Type != "slug" {
		t.Errorf("Type = %q, want %q", got.Type, "slug")
	}
	if got.Label != "Slug" {
		t.Errorf("Label = %q, want %q", got.Label, "Slug")
	}
}

func TestPsqlDatabase_MapCreateAdminFieldTypeParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	got1 := d.MapCreateAdminFieldTypeParams(CreateAdminFieldTypeParams{})
	got2 := d.MapCreateAdminFieldTypeParams(CreateAdminFieldTypeParams{})

	if got1.AdminFieldTypeID == got2.AdminFieldTypeID {
		t.Error("two calls produced identical AdminFieldTypeIDs")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateAdminFieldTypeParams tests ---

func TestPsqlDatabase_MapUpdateAdminFieldTypeParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	id := types.NewAdminFieldTypeID()

	input := UpdateAdminFieldTypeParams{
		Type:             "email",
		Label:            "Email",
		AdminFieldTypeID: id,
	}

	got := d.MapUpdateAdminFieldTypeParams(input)

	if got.AdminFieldTypeID != id {
		t.Errorf("AdminFieldTypeID = %v, want %v", got.AdminFieldTypeID, id)
	}
	if got.Type != "email" {
		t.Errorf("Type = %q, want %q", got.Type, "email")
	}
	if got.Label != "Email" {
		t.Errorf("Label = %q, want %q", got.Label, "Email")
	}
}

// --- Cross-database mapper consistency ---

func TestCrossDatabaseMapAdminFieldType_Consistency(t *testing.T) {
	t.Parallel()
	id := types.NewAdminFieldTypeID()

	sqliteInput := mdb.AdminFieldTypes{
		AdminFieldTypeID: id,
		Type:             "url",
		Label:            "URL",
	}
	mysqlInput := mdbm.AdminFieldTypes{
		AdminFieldTypeID: id,
		Type:             "url",
		Label:            "URL",
	}
	psqlInput := mdbp.AdminFieldTypes{
		AdminFieldTypeID: id,
		Type:             "url",
		Label:            "URL",
	}

	sqliteResult := Database{}.MapAdminFieldType(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminFieldType(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminFieldType(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateAdminFieldTypeParams - ID generation ---

func TestCrossDatabaseMapCreateAdminFieldTypeParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()

	input := CreateAdminFieldTypeParams{
		Type:  "relation",
		Label: "Relation",
	}

	sqliteResult := Database{}.MapCreateAdminFieldTypeParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateAdminFieldTypeParams(input)
	psqlResult := PsqlDatabase{}.MapCreateAdminFieldTypeParams(input)

	if sqliteResult.AdminFieldTypeID.IsZero() {
		t.Error("SQLite: expected non-zero generated AdminFieldTypeID")
	}
	if mysqlResult.AdminFieldTypeID.IsZero() {
		t.Error("MySQL: expected non-zero generated AdminFieldTypeID")
	}
	if psqlResult.AdminFieldTypeID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated AdminFieldTypeID")
	}

	// Each call should generate a unique ID
	if sqliteResult.AdminFieldTypeID == mysqlResult.AdminFieldTypeID {
		t.Error("SQLite and MySQL generated the same AdminFieldTypeID -- each call should be unique")
	}
	if sqliteResult.AdminFieldTypeID == psqlResult.AdminFieldTypeID {
		t.Error("SQLite and PostgreSQL generated the same AdminFieldTypeID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewAdminFieldTypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-aft-001",
		IP:        "10.0.0.1",
	}
	params := CreateAdminFieldTypeParams{
		Type:  "text",
		Label: "Text Input",
	}

	cmd := Database{}.NewAdminFieldTypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_field_types")
	}
	p, ok := cmd.Params().(CreateAdminFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminFieldTypeParams", cmd.Params())
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

func TestNewAdminFieldTypeCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	id := types.NewAdminFieldTypeID()
	cmd := NewAdminFieldTypeCmd{}

	row := mdb.AdminFieldTypes{AdminFieldTypeID: id}
	got := cmd.GetID(row)
	if got != string(id) {
		t.Errorf("GetID() = %q, want %q", got, string(id))
	}
}

func TestNewAdminFieldTypeCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewAdminFieldTypeCmd{}

	row := mdb.AdminFieldTypes{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateAdminFieldTypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewAdminFieldTypeID()
	params := UpdateAdminFieldTypeParams{
		Type:             "boolean",
		Label:            "Boolean",
		AdminFieldTypeID: id,
	}

	cmd := Database{}.UpdateAdminFieldTypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_field_types")
	}
	if cmd.GetID() != string(id) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(id))
	}
	p, ok := cmd.Params().(UpdateAdminFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminFieldTypeParams", cmd.Params())
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

func TestDeleteAdminFieldTypeCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewAdminFieldTypeID()

	cmd := Database{}.DeleteAdminFieldTypeCmd(ctx, ac, id)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_field_types")
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

func TestNewAdminFieldTypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-aft-001",
		IP:        "192.168.1.1",
	}
	params := CreateAdminFieldTypeParams{
		Type:  "textarea",
		Label: "Text Area",
	}

	cmd := MysqlDatabase{}.NewAdminFieldTypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_field_types")
	}
	p, ok := cmd.Params().(CreateAdminFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminFieldTypeParams", cmd.Params())
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

func TestNewAdminFieldTypeCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	id := types.NewAdminFieldTypeID()
	cmd := NewAdminFieldTypeCmdMysql{}

	row := mdbm.AdminFieldTypes{AdminFieldTypeID: id}
	got := cmd.GetID(row)
	if got != string(id) {
		t.Errorf("GetID() = %q, want %q", got, string(id))
	}
}

func TestUpdateAdminFieldTypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewAdminFieldTypeID()
	params := UpdateAdminFieldTypeParams{
		Type:             "date",
		Label:            "Date",
		AdminFieldTypeID: id,
	}

	cmd := MysqlDatabase{}.UpdateAdminFieldTypeCmd(ctx, ac, params)

	if cmd.TableName() != "admin_field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_field_types")
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
	p, ok := cmd.Params().(UpdateAdminFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminFieldTypeParams", cmd.Params())
	}
	if p.Type != "date" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "date")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminFieldTypeCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewAdminFieldTypeID()

	cmd := MysqlDatabase{}.DeleteAdminFieldTypeCmd(ctx, ac, id)

	if cmd.TableName() != "admin_field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_field_types")
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

func TestNewAdminFieldTypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-aft-001",
		IP:        "172.16.0.1",
	}
	params := CreateAdminFieldTypeParams{
		Type:  "relation",
		Label: "Relation",
	}

	cmd := PsqlDatabase{}.NewAdminFieldTypeCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_field_types")
	}
	p, ok := cmd.Params().(CreateAdminFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminFieldTypeParams", cmd.Params())
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

func TestNewAdminFieldTypeCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	id := types.NewAdminFieldTypeID()
	cmd := NewAdminFieldTypeCmdPsql{}

	row := mdbp.AdminFieldTypes{AdminFieldTypeID: id}
	got := cmd.GetID(row)
	if got != string(id) {
		t.Errorf("GetID() = %q, want %q", got, string(id))
	}
}

func TestUpdateAdminFieldTypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewAdminFieldTypeID()
	params := UpdateAdminFieldTypeParams{
		Type:             "datetime",
		Label:            "Date & Time",
		AdminFieldTypeID: id,
	}

	cmd := PsqlDatabase{}.UpdateAdminFieldTypeCmd(ctx, ac, params)

	if cmd.TableName() != "admin_field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_field_types")
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
	p, ok := cmd.Params().(UpdateAdminFieldTypeParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminFieldTypeParams", cmd.Params())
	}
	if p.Type != "datetime" {
		t.Errorf("Params().Type = %q, want %q", p.Type, "datetime")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminFieldTypeCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := types.NewAdminFieldTypeID()

	cmd := PsqlDatabase{}.DeleteAdminFieldTypeCmd(ctx, ac, id)

	if cmd.TableName() != "admin_field_types" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_field_types")
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

func TestAuditedAdminFieldTypeCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminFieldTypeParams{}
	updateParams := UpdateAdminFieldTypeParams{
		AdminFieldTypeID: types.NewAdminFieldTypeID(),
	}
	id := types.NewAdminFieldTypeID()

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewAdminFieldTypeCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateAdminFieldTypeCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteAdminFieldTypeCmd(ctx, ac, id).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewAdminFieldTypeCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateAdminFieldTypeCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteAdminFieldTypeCmd(ctx, ac, id).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewAdminFieldTypeCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateAdminFieldTypeCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteAdminFieldTypeCmd(ctx, ac, id).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "admin_field_types" {
				t.Errorf("TableName() = %q, want %q", c.name, "admin_field_types")
			}
		})
	}
}

func TestAuditedAdminFieldTypeCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminFieldTypeParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewAdminFieldTypeCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewAdminFieldTypeCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewAdminFieldTypeCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedAdminFieldTypeCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	id := types.NewAdminFieldTypeID()

	t.Run("UpdateCmd GetID returns admin field type ID", func(t *testing.T) {
		t.Parallel()
		params := UpdateAdminFieldTypeParams{
			AdminFieldTypeID: id,
		}

		sqliteCmd := Database{}.UpdateAdminFieldTypeCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateAdminFieldTypeCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateAdminFieldTypeCmd(ctx, ac, params)

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

	t.Run("DeleteCmd GetID returns admin field type ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteAdminFieldTypeCmd(ctx, ac, id)
		mysqlCmd := MysqlDatabase{}.DeleteAdminFieldTypeCmd(ctx, ac, id)
		psqlCmd := PsqlDatabase{}.DeleteAdminFieldTypeCmd(ctx, ac, id)

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
		testID := types.NewAdminFieldTypeID()
		wantID := string(testID)

		sqliteCmd := NewAdminFieldTypeCmd{}
		mysqlCmd := NewAdminFieldTypeCmdMysql{}
		psqlCmd := NewAdminFieldTypeCmdPsql{}

		sqliteRow := mdb.AdminFieldTypes{AdminFieldTypeID: testID}
		mysqlRow := mdbm.AdminFieldTypes{AdminFieldTypeID: testID}
		psqlRow := mdbp.AdminFieldTypes{AdminFieldTypeID: testID}

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

func TestUpdateAdminFieldTypeCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateAdminFieldTypeParams{
		AdminFieldTypeID: "",
	}

	sqliteCmd := Database{}.UpdateAdminFieldTypeCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateAdminFieldTypeCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateAdminFieldTypeCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteAdminFieldTypeCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.AdminFieldTypeID("")

	sqliteCmd := Database{}.DeleteAdminFieldTypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteAdminFieldTypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteAdminFieldTypeCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.AdminFieldTypes]  = NewAdminFieldTypeCmd{}
	_ audited.UpdateCommand[mdb.AdminFieldTypes]  = UpdateAdminFieldTypeCmd{}
	_ audited.DeleteCommand[mdb.AdminFieldTypes]  = DeleteAdminFieldTypeCmd{}
	_ audited.CreateCommand[mdbm.AdminFieldTypes] = NewAdminFieldTypeCmdMysql{}
	_ audited.UpdateCommand[mdbm.AdminFieldTypes] = UpdateAdminFieldTypeCmdMysql{}
	_ audited.DeleteCommand[mdbm.AdminFieldTypes] = DeleteAdminFieldTypeCmdMysql{}
	_ audited.CreateCommand[mdbp.AdminFieldTypes] = NewAdminFieldTypeCmdPsql{}
	_ audited.UpdateCommand[mdbp.AdminFieldTypes] = UpdateAdminFieldTypeCmdPsql{}
	_ audited.DeleteCommand[mdbp.AdminFieldTypes] = DeleteAdminFieldTypeCmdPsql{}
)
