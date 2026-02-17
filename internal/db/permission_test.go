// White-box tests for permission.go: wrapper structs, mapper methods
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

// permTestFixture returns a fully populated Permissions and its individual parts.
func permTestFixture() (Permissions, types.PermissionID, string) {
	permID := types.NewPermissionID()
	label := "full_access"
	p := Permissions{
		PermissionID: permID,
		Label:        label,
	}
	return p, permID, label
}

// --- MapStringPermission tests ---

func TestMapStringPermission_AllFields(t *testing.T) {
	t.Parallel()
	perm, permID, label := permTestFixture()

	got := MapStringPermission(perm)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"PermissionID", got.PermissionID, permID.String()},
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

func TestMapStringPermission_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringPermission(Permissions{})

	if got.PermissionID != "" {
		t.Errorf("PermissionID = %q, want empty string", got.PermissionID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapPermission tests ---

func TestDatabase_MapPermission_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	permID := types.NewPermissionID()

	input := mdb.Permissions{
		PermissionID: permID,
		Label:        "read_write",
	}

	got := d.MapPermission(input)

	if got.PermissionID != permID {
		t.Errorf("PermissionID = %v, want %v", got.PermissionID, permID)
	}
	if got.Label != "read_write" {
		t.Errorf("Label = %q, want %q", got.Label, "read_write")
	}
}

func TestDatabase_MapPermission_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapPermission(mdb.Permissions{})

	if got.PermissionID != "" {
		t.Errorf("PermissionID = %v, want zero value", got.PermissionID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapCreatePermissionParams tests ---

func TestDatabase_MapCreatePermissionParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := CreatePermissionParams{
		Label: "read_only",
	}

	got := d.MapCreatePermissionParams(input)

	if got.PermissionID.IsZero() {
		t.Fatal("expected non-zero PermissionID to be generated")
	}
	if got.Label != "read_only" {
		t.Errorf("Label = %q, want %q", got.Label, "read_only")
	}
}

func TestDatabase_MapCreatePermissionParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}

	got1 := d.MapCreatePermissionParams(CreatePermissionParams{})
	got2 := d.MapCreatePermissionParams(CreatePermissionParams{})

	if got1.PermissionID == got2.PermissionID {
		t.Error("two calls produced identical PermissionIDs")
	}
}

func TestDatabase_MapCreatePermissionParams_ZeroInput(t *testing.T) {
	t.Parallel()
	d := Database{}

	got := d.MapCreatePermissionParams(CreatePermissionParams{})

	if got.PermissionID.IsZero() {
		t.Fatal("expected non-zero PermissionID even with zero input")
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapUpdatePermissionParams tests ---

func TestDatabase_MapUpdatePermissionParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	permID := types.NewPermissionID()

	input := UpdatePermissionParams{
		Label:        "media_access",
		PermissionID: permID,
	}

	got := d.MapUpdatePermissionParams(input)

	if got.Label != "media_access" {
		t.Errorf("Label = %q, want %q", got.Label, "media_access")
	}
	if got.PermissionID != permID {
		t.Errorf("PermissionID = %v, want %v", got.PermissionID, permID)
	}
}

func TestDatabase_MapUpdatePermissionParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdatePermissionParams(UpdatePermissionParams{})

	if got.PermissionID != "" {
		t.Errorf("PermissionID = %v, want zero value", got.PermissionID)
	}
}

// --- MySQL MysqlDatabase.MapPermission tests ---

func TestMysqlDatabase_MapPermission_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	permID := types.NewPermissionID()

	input := mdbm.Permissions{
		PermissionID: permID,
		Label:        "test_label",
	}

	got := d.MapPermission(input)

	if got.PermissionID != permID {
		t.Errorf("PermissionID = %v, want %v", got.PermissionID, permID)
	}
	if got.Label != "test_label" {
		t.Errorf("Label = %q, want %q", got.Label, "test_label")
	}
}

func TestMysqlDatabase_MapPermission_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapPermission(mdbm.Permissions{})

	if got.PermissionID != "" {
		t.Errorf("PermissionID = %v, want zero value", got.PermissionID)
	}
}

// --- MySQL MysqlDatabase.MapCreatePermissionParams tests ---

func TestMysqlDatabase_MapCreatePermissionParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := CreatePermissionParams{
		Label: "route_admin",
	}

	got := d.MapCreatePermissionParams(input)

	if got.PermissionID.IsZero() {
		t.Fatal("expected non-zero PermissionID to be generated")
	}
	if got.Label != "route_admin" {
		t.Errorf("Label = %q, want %q", got.Label, "route_admin")
	}
}

func TestMysqlDatabase_MapCreatePermissionParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	got1 := d.MapCreatePermissionParams(CreatePermissionParams{})
	got2 := d.MapCreatePermissionParams(CreatePermissionParams{})

	if got1.PermissionID == got2.PermissionID {
		t.Error("two calls produced identical PermissionIDs")
	}
}

// --- MySQL MysqlDatabase.MapUpdatePermissionParams tests ---

func TestMysqlDatabase_MapUpdatePermissionParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	permID := types.NewPermissionID()

	input := UpdatePermissionParams{
		Label:        "media_read",
		PermissionID: permID,
	}

	got := d.MapUpdatePermissionParams(input)

	if got.PermissionID != permID {
		t.Errorf("PermissionID = %v, want %v", got.PermissionID, permID)
	}
	if got.Label != "media_read" {
		t.Errorf("Label = %q, want %q", got.Label, "media_read")
	}
}

// --- PostgreSQL PsqlDatabase.MapPermission tests ---

func TestPsqlDatabase_MapPermission_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	permID := types.NewPermissionID()

	input := mdbp.Permissions{
		PermissionID: permID,
		Label:        "psql_label",
	}

	got := d.MapPermission(input)

	if got.PermissionID != permID {
		t.Errorf("PermissionID = %v, want %v", got.PermissionID, permID)
	}
	if got.Label != "psql_label" {
		t.Errorf("Label = %q, want %q", got.Label, "psql_label")
	}
}

func TestPsqlDatabase_MapPermission_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapPermission(mdbp.Permissions{})

	if got.PermissionID != "" {
		t.Errorf("PermissionID = %v, want zero value", got.PermissionID)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreatePermissionParams tests ---

func TestPsqlDatabase_MapCreatePermissionParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := CreatePermissionParams{
		Label: "field_write",
	}

	got := d.MapCreatePermissionParams(input)

	if got.PermissionID.IsZero() {
		t.Fatal("expected non-zero PermissionID to be generated")
	}
	if got.Label != "field_write" {
		t.Errorf("Label = %q, want %q", got.Label, "field_write")
	}
}

func TestPsqlDatabase_MapCreatePermissionParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	got1 := d.MapCreatePermissionParams(CreatePermissionParams{})
	got2 := d.MapCreatePermissionParams(CreatePermissionParams{})

	if got1.PermissionID == got2.PermissionID {
		t.Error("two calls produced identical PermissionIDs")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdatePermissionParams tests ---

func TestPsqlDatabase_MapUpdatePermissionParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	permID := types.NewPermissionID()

	input := UpdatePermissionParams{
		Label:        "session_read",
		PermissionID: permID,
	}

	got := d.MapUpdatePermissionParams(input)

	if got.PermissionID != permID {
		t.Errorf("PermissionID = %v, want %v", got.PermissionID, permID)
	}
	if got.Label != "session_read" {
		t.Errorf("Label = %q, want %q", got.Label, "session_read")
	}
}

// --- Cross-database mapper consistency ---

func TestCrossDatabaseMapPermission_Consistency(t *testing.T) {
	t.Parallel()
	permID := types.NewPermissionID()

	sqliteInput := mdb.Permissions{
		PermissionID: permID,
		Label:        "full_access",
	}
	mysqlInput := mdbm.Permissions{
		PermissionID: permID,
		Label:        "full_access",
	}
	psqlInput := mdbp.Permissions{
		PermissionID: permID,
		Label:        "full_access",
	}

	sqliteResult := Database{}.MapPermission(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapPermission(mysqlInput)
	psqlResult := PsqlDatabase{}.MapPermission(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreatePermissionParams - ID generation ---

func TestCrossDatabaseMapCreatePermissionParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()

	input := CreatePermissionParams{
		Label: "route_manage",
	}

	sqliteResult := Database{}.MapCreatePermissionParams(input)
	mysqlResult := MysqlDatabase{}.MapCreatePermissionParams(input)
	psqlResult := PsqlDatabase{}.MapCreatePermissionParams(input)

	if sqliteResult.PermissionID.IsZero() {
		t.Error("SQLite: expected non-zero generated PermissionID")
	}
	if mysqlResult.PermissionID.IsZero() {
		t.Error("MySQL: expected non-zero generated PermissionID")
	}
	if psqlResult.PermissionID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated PermissionID")
	}

	// Each call should generate a unique ID
	if sqliteResult.PermissionID == mysqlResult.PermissionID {
		t.Error("SQLite and MySQL generated the same PermissionID -- each call should be unique")
	}
	if sqliteResult.PermissionID == psqlResult.PermissionID {
		t.Error("SQLite and PostgreSQL generated the same PermissionID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewPermissionCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-perm-001",
		IP:        "10.0.0.1",
	}
	params := CreatePermissionParams{
		Label: "user_admin",
	}

	cmd := Database{}.NewPermissionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
	p, ok := cmd.Params().(CreatePermissionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreatePermissionParams", cmd.Params())
	}
	if p.Label != "user_admin" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "user_admin")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewPermissionCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	permID := types.NewPermissionID()
	cmd := NewPermissionCmd{}

	row := mdb.Permissions{PermissionID: permID}
	got := cmd.GetID(row)
	if got != string(permID) {
		t.Errorf("GetID() = %q, want %q", got, string(permID))
	}
}

func TestNewPermissionCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewPermissionCmd{}

	row := mdb.Permissions{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdatePermissionCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	permID := types.NewPermissionID()
	params := UpdatePermissionParams{
		Label:        "role_read",
		PermissionID: permID,
	}

	cmd := Database{}.UpdatePermissionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
	if cmd.GetID() != string(permID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(permID))
	}
	p, ok := cmd.Params().(UpdatePermissionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdatePermissionParams", cmd.Params())
	}
	if p.Label != "role_read" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "role_read")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeletePermissionCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	permID := types.NewPermissionID()

	cmd := Database{}.DeletePermissionCmd(ctx, ac, permID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
	if cmd.GetID() != string(permID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(permID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewPermissionCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-perm-001",
		IP:        "192.168.1.1",
	}
	params := CreatePermissionParams{
		Label: "content_manage",
	}

	cmd := MysqlDatabase{}.NewPermissionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
	p, ok := cmd.Params().(CreatePermissionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreatePermissionParams", cmd.Params())
	}
	if p.Label != "content_manage" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "content_manage")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewPermissionCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	permID := types.NewPermissionID()
	cmd := NewPermissionCmdMysql{}

	row := mdbm.Permissions{PermissionID: permID}
	got := cmd.GetID(row)
	if got != string(permID) {
		t.Errorf("GetID() = %q, want %q", got, string(permID))
	}
}

func TestUpdatePermissionCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	permID := types.NewPermissionID()
	params := UpdatePermissionParams{
		Label:        "media_write",
		PermissionID: permID,
	}

	cmd := MysqlDatabase{}.UpdatePermissionCmd(ctx, ac, params)

	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
	if cmd.GetID() != string(permID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(permID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdatePermissionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdatePermissionParams", cmd.Params())
	}
	if p.Label != "media_write" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "media_write")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeletePermissionCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	permID := types.NewPermissionID()

	cmd := MysqlDatabase{}.DeletePermissionCmd(ctx, ac, permID)

	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
	if cmd.GetID() != string(permID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(permID))
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

func TestNewPermissionCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-perm-001",
		IP:        "172.16.0.1",
	}
	params := CreatePermissionParams{
		Label: "datatype_admin",
	}

	cmd := PsqlDatabase{}.NewPermissionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
	p, ok := cmd.Params().(CreatePermissionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreatePermissionParams", cmd.Params())
	}
	if p.Label != "datatype_admin" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "datatype_admin")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewPermissionCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	permID := types.NewPermissionID()
	cmd := NewPermissionCmdPsql{}

	row := mdbp.Permissions{PermissionID: permID}
	got := cmd.GetID(row)
	if got != string(permID) {
		t.Errorf("GetID() = %q, want %q", got, string(permID))
	}
}

func TestUpdatePermissionCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	permID := types.NewPermissionID()
	params := UpdatePermissionParams{
		Label:        "field_read",
		PermissionID: permID,
	}

	cmd := PsqlDatabase{}.UpdatePermissionCmd(ctx, ac, params)

	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
	if cmd.GetID() != string(permID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(permID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdatePermissionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdatePermissionParams", cmd.Params())
	}
	if p.Label != "field_read" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "field_read")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeletePermissionCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	permID := types.NewPermissionID()

	cmd := PsqlDatabase{}.DeletePermissionCmd(ctx, ac, permID)

	if cmd.TableName() != "permissions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "permissions")
	}
	if cmd.GetID() != string(permID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(permID))
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

func TestAuditedPermissionCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreatePermissionParams{}
	updateParams := UpdatePermissionParams{
		PermissionID: types.NewPermissionID(),
	}
	permID := types.NewPermissionID()

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewPermissionCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdatePermissionCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeletePermissionCmd(ctx, ac, permID).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewPermissionCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdatePermissionCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeletePermissionCmd(ctx, ac, permID).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewPermissionCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdatePermissionCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeletePermissionCmd(ctx, ac, permID).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "permissions" {
				t.Errorf("TableName() = %q, want %q", c.name, "permissions")
			}
		})
	}
}

func TestAuditedPermissionCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreatePermissionParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewPermissionCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewPermissionCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewPermissionCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedPermissionCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	permID := types.NewPermissionID()

	t.Run("UpdateCmd GetID returns permission ID", func(t *testing.T) {
		t.Parallel()
		params := UpdatePermissionParams{
			PermissionID: permID,
		}

		sqliteCmd := Database{}.UpdatePermissionCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdatePermissionCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdatePermissionCmd(ctx, ac, params)

		wantID := string(permID)
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

	t.Run("DeleteCmd GetID returns permission ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeletePermissionCmd(ctx, ac, permID)
		mysqlCmd := MysqlDatabase{}.DeletePermissionCmd(ctx, ac, permID)
		psqlCmd := PsqlDatabase{}.DeletePermissionCmd(ctx, ac, permID)

		wantID := string(permID)
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
		testPermID := types.NewPermissionID()
		wantID := string(testPermID)

		sqliteCmd := NewPermissionCmd{}
		mysqlCmd := NewPermissionCmdMysql{}
		psqlCmd := NewPermissionCmdPsql{}

		sqliteRow := mdb.Permissions{PermissionID: testPermID}
		mysqlRow := mdbm.Permissions{PermissionID: testPermID}
		psqlRow := mdbp.Permissions{PermissionID: testPermID}

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

func TestUpdatePermissionCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdatePermissionParams{
		PermissionID: "",
	}

	sqliteCmd := Database{}.UpdatePermissionCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdatePermissionCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdatePermissionCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeletePermissionCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.PermissionID("")

	sqliteCmd := Database{}.DeletePermissionCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeletePermissionCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeletePermissionCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.Permissions]  = NewPermissionCmd{}
	_ audited.UpdateCommand[mdb.Permissions]  = UpdatePermissionCmd{}
	_ audited.DeleteCommand[mdb.Permissions]  = DeletePermissionCmd{}
	_ audited.CreateCommand[mdbm.Permissions] = NewPermissionCmdMysql{}
	_ audited.UpdateCommand[mdbm.Permissions] = UpdatePermissionCmdMysql{}
	_ audited.DeleteCommand[mdbm.Permissions] = DeletePermissionCmdMysql{}
	_ audited.CreateCommand[mdbp.Permissions] = NewPermissionCmdPsql{}
	_ audited.UpdateCommand[mdbp.Permissions] = UpdatePermissionCmdPsql{}
	_ audited.DeleteCommand[mdbp.Permissions] = DeletePermissionCmdPsql{}
)
