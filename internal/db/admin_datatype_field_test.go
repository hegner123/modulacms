// White-box tests for admin_datatype_field.go: wrapper structs, mapper methods
// across all three database drivers (SQLite, MySQL, PostgreSQL), string mapping
// for TUI display, and audited command struct accessors.
//
// White-box access is needed because:
//   - Audited command structs have unexported fields (ctx, auditCtx, params, conn,
//     recorder) that can only be constructed through the Database/MysqlDatabase/
//     PsqlDatabase factory methods, which require access to the package internals.
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

// adfTestFixture returns fully populated test data for AdminDatatypeFields tests.
func adfTestFixture() (AdminDatatypeFields, string, types.AdminDatatypeID, types.AdminFieldID) {
	id := string(types.NewAdminDatatypeFieldID())
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()
	adf := AdminDatatypeFields{
		ID:              id,
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}
	return adf, id, datatypeID, fieldID
}

// --- MapStringAdminDatatypeField tests ---

func TestMapStringAdminDatatypeField_AllFields(t *testing.T) {
	t.Parallel()
	adf, id, datatypeID, fieldID := adfTestFixture()

	got := MapStringAdminDatatypeField(adf)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ID", got.ID, id},
		{"AdminDatatypeID", got.AdminDatatypeID, datatypeID.String()},
		{"AdminFieldID", got.AdminFieldID, fieldID.String()},
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

func TestMapStringAdminDatatypeField_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringAdminDatatypeField(AdminDatatypeFields{})

	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
	// Zero-value AdminDatatypeID is empty string, .String() returns ""
	if got.AdminDatatypeID != "" {
		t.Errorf("AdminDatatypeID = %q, want %q", got.AdminDatatypeID, "")
	}
	// Zero-value AdminFieldID is empty string, .String() returns ""
	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %q, want %q", got.AdminFieldID, "")
	}
}

// --- SQLite Database.MapAdminDatatypeField tests ---

func TestDatabase_MapAdminDatatypeField_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	id := string(types.NewAdminDatatypeFieldID())
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := mdb.AdminDatatypesFields{
		ID:              id,
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	got := d.MapAdminDatatypeField(input)

	if got.ID != id {
		t.Errorf("ID = %v, want %v", got.ID, id)
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
}

func TestDatabase_MapAdminDatatypeField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapAdminDatatypeField(mdb.AdminDatatypesFields{})

	if got.ID != "" {
		t.Errorf("ID = %v, want zero value", got.ID)
	}
	if got.AdminDatatypeID != "" {
		t.Errorf("AdminDatatypeID = %v, want zero value", got.AdminDatatypeID)
	}
	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

// --- SQLite Database.MapCreateAdminDatatypeFieldParams tests ---

func TestDatabase_MapCreateAdminDatatypeFieldParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := CreateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	got := d.MapCreateAdminDatatypeFieldParams(input)

	// A new ID should always be generated (non-empty string)
	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
}

func TestDatabase_MapCreateAdminDatatypeFieldParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}

	// Two calls should produce different IDs
	got1 := d.MapCreateAdminDatatypeFieldParams(CreateAdminDatatypeFieldParams{})
	got2 := d.MapCreateAdminDatatypeFieldParams(CreateAdminDatatypeFieldParams{})

	if got1.ID == got2.ID {
		t.Error("two calls produced identical IDs")
	}
}

// --- SQLite Database.MapUpdateAdminDatatypeFieldParams tests ---

func TestDatabase_MapUpdateAdminDatatypeFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	id := string(types.NewAdminDatatypeFieldID())
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
		ID:              id,
	}

	got := d.MapUpdateAdminDatatypeFieldParams(input)

	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
	if got.ID != id {
		t.Errorf("ID = %v, want %v", got.ID, id)
	}
}

func TestDatabase_MapUpdateAdminDatatypeFieldParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateAdminDatatypeFieldParams(UpdateAdminDatatypeFieldParams{})

	if got.ID != "" {
		t.Errorf("ID = %v, want zero value", got.ID)
	}
	if got.AdminDatatypeID != "" {
		t.Errorf("AdminDatatypeID = %v, want zero value", got.AdminDatatypeID)
	}
	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

// --- MySQL MysqlDatabase.MapAdminDatatypeField tests ---

func TestMysqlDatabase_MapAdminDatatypeField_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	id := string(types.NewAdminDatatypeFieldID())
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := mdbm.AdminDatatypesFields{
		ID:              id,
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	got := d.MapAdminDatatypeField(input)

	if got.ID != id {
		t.Errorf("ID = %v, want %v", got.ID, id)
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
}

func TestMysqlDatabase_MapAdminDatatypeField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapAdminDatatypeField(mdbm.AdminDatatypesFields{})

	if got.ID != "" {
		t.Errorf("ID = %v, want zero value", got.ID)
	}
	if got.AdminDatatypeID != "" {
		t.Errorf("AdminDatatypeID = %v, want zero value", got.AdminDatatypeID)
	}
	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

// --- MySQL MysqlDatabase.MapCreateAdminDatatypeFieldParams tests ---

func TestMysqlDatabase_MapCreateAdminDatatypeFieldParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := CreateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	got := d.MapCreateAdminDatatypeFieldParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
}

// --- MySQL MysqlDatabase.MapUpdateAdminDatatypeFieldParams tests ---

func TestMysqlDatabase_MapUpdateAdminDatatypeFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	id := string(types.NewAdminDatatypeFieldID())
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
		ID:              id,
	}

	got := d.MapUpdateAdminDatatypeFieldParams(input)

	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
	if got.ID != id {
		t.Errorf("ID = %v, want %v", got.ID, id)
	}
}

// --- PostgreSQL PsqlDatabase.MapAdminDatatypeField tests ---

func TestPsqlDatabase_MapAdminDatatypeField_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	id := string(types.NewAdminDatatypeFieldID())
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := mdbp.AdminDatatypesFields{
		ID:              id,
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	got := d.MapAdminDatatypeField(input)

	if got.ID != id {
		t.Errorf("ID = %v, want %v", got.ID, id)
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
}

func TestPsqlDatabase_MapAdminDatatypeField_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapAdminDatatypeField(mdbp.AdminDatatypesFields{})

	if got.ID != "" {
		t.Errorf("ID = %v, want zero value", got.ID)
	}
	if got.AdminDatatypeID != "" {
		t.Errorf("AdminDatatypeID = %v, want zero value", got.AdminDatatypeID)
	}
	if got.AdminFieldID != "" {
		t.Errorf("AdminFieldID = %v, want zero value", got.AdminFieldID)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateAdminDatatypeFieldParams tests ---

func TestPsqlDatabase_MapCreateAdminDatatypeFieldParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := CreateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	got := d.MapCreateAdminDatatypeFieldParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateAdminDatatypeFieldParams tests ---

func TestPsqlDatabase_MapUpdateAdminDatatypeFieldParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	id := string(types.NewAdminDatatypeFieldID())
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
		ID:              id,
	}

	got := d.MapUpdateAdminDatatypeFieldParams(input)

	if got.AdminDatatypeID != datatypeID {
		t.Errorf("AdminDatatypeID = %v, want %v", got.AdminDatatypeID, datatypeID)
	}
	if got.AdminFieldID != fieldID {
		t.Errorf("AdminFieldID = %v, want %v", got.AdminFieldID, fieldID)
	}
	if got.ID != id {
		t.Errorf("ID = %v, want %v", got.ID, id)
	}
}

// --- Cross-database mapper consistency ---
// Verify that all three database mappers produce identical AdminDatatypeFields
// from equivalent input.

func TestCrossDatabaseMapAdminDatatypeField_Consistency(t *testing.T) {
	t.Parallel()
	id := string(types.NewAdminDatatypeFieldID())
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	sqliteInput := mdb.AdminDatatypesFields{
		ID:              id,
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}
	mysqlInput := mdbm.AdminDatatypesFields{
		ID:              id,
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}
	psqlInput := mdbp.AdminDatatypesFields{
		ID:              id,
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	sqliteResult := Database{}.MapAdminDatatypeField(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapAdminDatatypeField(mysqlInput)
	psqlResult := PsqlDatabase{}.MapAdminDatatypeField(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateAdminDatatypeFieldParams - ID generation ---

func TestCrossDatabaseMapCreateAdminDatatypeFieldParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()

	input := CreateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	sqliteResult := Database{}.MapCreateAdminDatatypeFieldParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateAdminDatatypeFieldParams(input)
	psqlResult := PsqlDatabase{}.MapCreateAdminDatatypeFieldParams(input)

	if sqliteResult.ID == "" {
		t.Error("SQLite: expected non-empty generated ID")
	}
	if mysqlResult.ID == "" {
		t.Error("MySQL: expected non-empty generated ID")
	}
	if psqlResult.ID == "" {
		t.Error("PostgreSQL: expected non-empty generated ID")
	}

	// Each call should generate a unique ID
	if sqliteResult.ID == mysqlResult.ID {
		t.Error("SQLite and MySQL generated the same ID -- each call should be unique")
	}
	if sqliteResult.ID == psqlResult.ID {
		t.Error("SQLite and PostgreSQL generated the same ID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewAdminDatatypeFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-adf-001",
		IP:        "10.0.0.1",
	}
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()
	params := CreateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	cmd := Database{}.NewAdminDatatypeFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes_fields")
	}
	p, ok := cmd.Params().(CreateAdminDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminDatatypeFieldParams", cmd.Params())
	}
	if p.AdminDatatypeID != datatypeID {
		t.Errorf("Params().AdminDatatypeID = %v, want %v", p.AdminDatatypeID, datatypeID)
	}
	if p.AdminFieldID != fieldID {
		t.Errorf("Params().AdminFieldID = %v, want %v", p.AdminFieldID, fieldID)
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminDatatypeFieldCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	id := string(types.NewAdminDatatypeFieldID())
	cmd := NewAdminDatatypeFieldCmd{}

	row := mdb.AdminDatatypesFields{ID: id}
	got := cmd.GetID(row)
	if got != id {
		t.Errorf("GetID() = %q, want %q", got, id)
	}
}

func TestUpdateAdminDatatypeFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := string(types.NewAdminDatatypeFieldID())
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()
	params := UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
		ID:              id,
	}

	cmd := Database{}.UpdateAdminDatatypeFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes_fields")
	}
	if cmd.GetID() != id {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), id)
	}
	p, ok := cmd.Params().(UpdateAdminDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminDatatypeFieldParams", cmd.Params())
	}
	if p.AdminDatatypeID != datatypeID {
		t.Errorf("Params().AdminDatatypeID = %v, want %v", p.AdminDatatypeID, datatypeID)
	}
	if p.AdminFieldID != fieldID {
		t.Errorf("Params().AdminFieldID = %v, want %v", p.AdminFieldID, fieldID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminDatatypeFieldCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := string(types.NewAdminDatatypeFieldID())

	cmd := Database{}.DeleteAdminDatatypeFieldCmd(ctx, ac, id)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes_fields")
	}
	if cmd.GetID() != id {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), id)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewAdminDatatypeFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-adf-001",
		IP:        "192.168.1.1",
	}
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()
	params := CreateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	cmd := MysqlDatabase{}.NewAdminDatatypeFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes_fields")
	}
	p, ok := cmd.Params().(CreateAdminDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminDatatypeFieldParams", cmd.Params())
	}
	if p.AdminDatatypeID != datatypeID {
		t.Errorf("Params().AdminDatatypeID = %v, want %v", p.AdminDatatypeID, datatypeID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminDatatypeFieldCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	id := string(types.NewAdminDatatypeFieldID())
	cmd := NewAdminDatatypeFieldCmdMysql{}

	row := mdbm.AdminDatatypesFields{ID: id}
	got := cmd.GetID(row)
	if got != id {
		t.Errorf("GetID() = %q, want %q", got, id)
	}
}

func TestUpdateAdminDatatypeFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := string(types.NewAdminDatatypeFieldID())
	params := UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		AdminFieldID:    types.NewAdminFieldID(),
		ID:              id,
	}

	cmd := MysqlDatabase{}.UpdateAdminDatatypeFieldCmd(ctx, ac, params)

	if cmd.TableName() != "admin_datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes_fields")
	}
	if cmd.GetID() != id {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), id)
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminDatatypeFieldParams", cmd.Params())
	}
	if p.ID != id {
		t.Errorf("Params().ID = %q, want %q", p.ID, id)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminDatatypeFieldCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := string(types.NewAdminDatatypeFieldID())

	cmd := MysqlDatabase{}.DeleteAdminDatatypeFieldCmd(ctx, ac, id)

	if cmd.TableName() != "admin_datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes_fields")
	}
	if cmd.GetID() != id {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), id)
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

func TestNewAdminDatatypeFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-adf-001",
		IP:        "172.16.0.1",
	}
	datatypeID := types.NewAdminDatatypeID()
	fieldID := types.NewAdminFieldID()
	params := CreateAdminDatatypeFieldParams{
		AdminDatatypeID: datatypeID,
		AdminFieldID:    fieldID,
	}

	cmd := PsqlDatabase{}.NewAdminDatatypeFieldCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "admin_datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes_fields")
	}
	p, ok := cmd.Params().(CreateAdminDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateAdminDatatypeFieldParams", cmd.Params())
	}
	if p.AdminFieldID != fieldID {
		t.Errorf("Params().AdminFieldID = %v, want %v", p.AdminFieldID, fieldID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewAdminDatatypeFieldCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	id := string(types.NewAdminDatatypeFieldID())
	cmd := NewAdminDatatypeFieldCmdPsql{}

	row := mdbp.AdminDatatypesFields{ID: id}
	got := cmd.GetID(row)
	if got != id {
		t.Errorf("GetID() = %q, want %q", got, id)
	}
}

func TestUpdateAdminDatatypeFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := string(types.NewAdminDatatypeFieldID())
	params := UpdateAdminDatatypeFieldParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		AdminFieldID:    types.NewAdminFieldID(),
		ID:              id,
	}

	cmd := PsqlDatabase{}.UpdateAdminDatatypeFieldCmd(ctx, ac, params)

	if cmd.TableName() != "admin_datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes_fields")
	}
	if cmd.GetID() != id {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), id)
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateAdminDatatypeFieldParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateAdminDatatypeFieldParams", cmd.Params())
	}
	if p.ID != id {
		t.Errorf("Params().ID = %q, want %q", p.ID, id)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteAdminDatatypeFieldCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := string(types.NewAdminDatatypeFieldID())

	cmd := PsqlDatabase{}.DeleteAdminDatatypeFieldCmd(ctx, ac, id)

	if cmd.TableName() != "admin_datatypes_fields" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "admin_datatypes_fields")
	}
	if cmd.GetID() != id {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), id)
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

func TestAuditedAdminDatatypeFieldCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminDatatypeFieldParams{}
	updateParams := UpdateAdminDatatypeFieldParams{
		ID: string(types.NewAdminDatatypeFieldID()),
	}
	id := string(types.NewAdminDatatypeFieldID())

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewAdminDatatypeFieldCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateAdminDatatypeFieldCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteAdminDatatypeFieldCmd(ctx, ac, id).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewAdminDatatypeFieldCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateAdminDatatypeFieldCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteAdminDatatypeFieldCmd(ctx, ac, id).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewAdminDatatypeFieldCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateAdminDatatypeFieldCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteAdminDatatypeFieldCmd(ctx, ac, id).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "admin_datatypes_fields" {
				t.Errorf("TableName() = %q, want %q", c.name, "admin_datatypes_fields")
			}
		})
	}
}

func TestAuditedAdminDatatypeFieldCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateAdminDatatypeFieldParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewAdminDatatypeFieldCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewAdminDatatypeFieldCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewAdminDatatypeFieldCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedAdminDatatypeFieldCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	id := string(types.NewAdminDatatypeFieldID())

	t.Run("UpdateCmd GetID returns ID", func(t *testing.T) {
		t.Parallel()
		params := UpdateAdminDatatypeFieldParams{
			ID: id,
		}

		sqliteCmd := Database{}.UpdateAdminDatatypeFieldCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateAdminDatatypeFieldCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateAdminDatatypeFieldCmd(ctx, ac, params)

		if sqliteCmd.GetID() != id {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), id)
		}
		if mysqlCmd.GetID() != id {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), id)
		}
		if psqlCmd.GetID() != id {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), id)
		}
	})

	t.Run("DeleteCmd GetID returns ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteAdminDatatypeFieldCmd(ctx, ac, id)
		mysqlCmd := MysqlDatabase{}.DeleteAdminDatatypeFieldCmd(ctx, ac, id)
		psqlCmd := PsqlDatabase{}.DeleteAdminDatatypeFieldCmd(ctx, ac, id)

		if sqliteCmd.GetID() != id {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), id)
		}
		if mysqlCmd.GetID() != id {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), id)
		}
		if psqlCmd.GetID() != id {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), id)
		}
	})

	t.Run("CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		testID := string(types.NewAdminDatatypeFieldID())

		sqliteCmd := NewAdminDatatypeFieldCmd{}
		mysqlCmd := NewAdminDatatypeFieldCmdMysql{}
		psqlCmd := NewAdminDatatypeFieldCmdPsql{}

		sqliteRow := mdb.AdminDatatypesFields{ID: testID}
		mysqlRow := mdbm.AdminDatatypesFields{ID: testID}
		psqlRow := mdbp.AdminDatatypesFields{ID: testID}

		if sqliteCmd.GetID(sqliteRow) != testID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(sqliteRow), testID)
		}
		if mysqlCmd.GetID(mysqlRow) != testID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(mysqlRow), testID)
		}
		if psqlCmd.GetID(psqlRow) != testID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(psqlRow), testID)
		}
	})
}

// --- Edge cases: empty IDs ---

func TestUpdateAdminDatatypeFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateAdminDatatypeFieldParams{
		ID: "",
	}

	sqliteCmd := Database{}.UpdateAdminDatatypeFieldCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateAdminDatatypeFieldCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateAdminDatatypeFieldCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteAdminDatatypeFieldCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := ""

	sqliteCmd := Database{}.DeleteAdminDatatypeFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteAdminDatatypeFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteAdminDatatypeFieldCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.AdminDatatypesFields]  = NewAdminDatatypeFieldCmd{}
	_ audited.UpdateCommand[mdb.AdminDatatypesFields]  = UpdateAdminDatatypeFieldCmd{}
	_ audited.DeleteCommand[mdb.AdminDatatypesFields]  = DeleteAdminDatatypeFieldCmd{}
	_ audited.CreateCommand[mdbm.AdminDatatypesFields] = NewAdminDatatypeFieldCmdMysql{}
	_ audited.UpdateCommand[mdbm.AdminDatatypesFields] = UpdateAdminDatatypeFieldCmdMysql{}
	_ audited.DeleteCommand[mdbm.AdminDatatypesFields] = DeleteAdminDatatypeFieldCmdMysql{}
	_ audited.CreateCommand[mdbp.AdminDatatypesFields] = NewAdminDatatypeFieldCmdPsql{}
	_ audited.UpdateCommand[mdbp.AdminDatatypesFields] = UpdateAdminDatatypeFieldCmdPsql{}
	_ audited.DeleteCommand[mdbp.AdminDatatypesFields] = DeleteAdminDatatypeFieldCmdPsql{}
)
