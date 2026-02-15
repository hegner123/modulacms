package db

import (
	"context"
	"encoding/json"
	"testing"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/sqlc-dev/pqtype"
)

// --- Test data helpers ---

// --- SQLite Database.MapRole tests ---

func TestDatabase_MapRole_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	roleID := types.NewRoleID()

	input := mdb.Roles{
		RoleID:      roleID,
		Label:       "editor",
		Permissions: `{"read":true,"write":true}`,
	}

	got := d.MapRole(input)

	if got.RoleID != roleID {
		t.Errorf("RoleID = %v, want %v", got.RoleID, roleID)
	}
	if got.Label != "editor" {
		t.Errorf("Label = %q, want %q", got.Label, "editor")
	}
	if got.Permissions != `{"read":true,"write":true}` {
		t.Errorf("Permissions = %q, want %q", got.Permissions, `{"read":true,"write":true}`)
	}
}

func TestDatabase_MapRole_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapRole(mdb.Roles{})

	if got.RoleID != "" {
		t.Errorf("RoleID = %v, want zero value", got.RoleID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.Permissions != "" {
		t.Errorf("Permissions = %q, want empty string", got.Permissions)
	}
}

// --- SQLite Database.MapCreateRoleParams tests ---

func TestDatabase_MapCreateRoleParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := CreateRoleParams{
		Label:       "viewer",
		Permissions: `{"read":true}`,
	}

	got := d.MapCreateRoleParams(input)

	// MapCreateRoleParams always generates a new RoleID via types.NewRoleID()
	if got.RoleID.IsZero() {
		t.Fatal("expected non-zero RoleID to be generated")
	}
	if got.Label != "viewer" {
		t.Errorf("Label = %q, want %q", got.Label, "viewer")
	}
	if got.Permissions != `{"read":true}` {
		t.Errorf("Permissions = %q, want %q", got.Permissions, `{"read":true}`)
	}
}

func TestDatabase_MapCreateRoleParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateRoleParams{Label: "test"}

	// Two calls should produce different IDs
	got1 := d.MapCreateRoleParams(input)
	got2 := d.MapCreateRoleParams(input)
	if got1.RoleID == got2.RoleID {
		t.Error("two calls produced the same RoleID -- each call should generate a unique ID")
	}
}

// --- SQLite Database.MapUpdateRoleParams tests ---

func TestDatabase_MapUpdateRoleParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	roleID := types.NewRoleID()

	input := UpdateRoleParams{
		Label:       "superadmin",
		Permissions: `{"all":true}`,
		RoleID:      roleID,
	}

	got := d.MapUpdateRoleParams(input)

	if got.Label != "superadmin" {
		t.Errorf("Label = %q, want %q", got.Label, "superadmin")
	}
	if got.Permissions != `{"all":true}` {
		t.Errorf("Permissions = %q, want %q", got.Permissions, `{"all":true}`)
	}
	if got.RoleID != roleID {
		t.Errorf("RoleID = %v, want %v", got.RoleID, roleID)
	}
}

// --- MySQL MysqlDatabase.MapRole tests ---

func TestMysqlDatabase_MapRole_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	roleID := types.NewRoleID()

	input := mdbm.Roles{
		RoleID:      roleID,
		Label:       "editor",
		Permissions: json.RawMessage(`{"read":true,"write":true}`),
	}

	got := d.MapRole(input)

	if got.RoleID != roleID {
		t.Errorf("RoleID = %v, want %v", got.RoleID, roleID)
	}
	if got.Label != "editor" {
		t.Errorf("Label = %q, want %q", got.Label, "editor")
	}
	if got.Permissions != `{"read":true,"write":true}` {
		t.Errorf("Permissions = %q, want %q", got.Permissions, `{"read":true,"write":true}`)
	}
}

func TestMysqlDatabase_MapRole_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapRole(mdbm.Roles{})

	if got.RoleID != "" {
		t.Errorf("RoleID = %v, want zero value", got.RoleID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	// NullString with Valid=false: String field is "" so Permissions should be ""
	if got.Permissions != "" {
		t.Errorf("Permissions = %q, want empty string", got.Permissions)
	}
}

func TestMysqlDatabase_MapRole_NullPermissions(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	// When Permissions is NULL in MySQL, NullString.Valid is false and String is ""
	input := mdbm.Roles{
		RoleID:      types.NewRoleID(),
		Label:       "viewer",
		Permissions: nil,
	}

	got := d.MapRole(input)

	// The mapper uses .String which returns "" when Valid is false
	if got.Permissions != "" {
		t.Errorf("Permissions = %q, want empty string for NULL permissions", got.Permissions)
	}
}

// --- MySQL MysqlDatabase.MapCreateRoleParams tests ---

func TestMysqlDatabase_MapCreateRoleParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := CreateRoleParams{
		Label:       "mysqlrole",
		Permissions: `{"read":true}`,
	}

	got := d.MapCreateRoleParams(input)

	if got.RoleID.IsZero() {
		t.Fatal("expected non-zero RoleID to be generated")
	}
	if got.Label != "mysqlrole" {
		t.Errorf("Label = %q, want %q", got.Label, "mysqlrole")
	}
	// Permissions should be wrapped in json.RawMessage
	if len(got.Permissions) == 0 {
		t.Fatal("Permissions is empty, want non-empty for non-empty string")
	}
	if string(got.Permissions) != `{"read":true}` {
		t.Errorf("Permissions = %q, want %q", string(got.Permissions), `{"read":true}`)
	}
}

func TestMysqlDatabase_MapCreateRoleParams_EmptyPermissions(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	// Empty string should produce empty json.RawMessage
	input := CreateRoleParams{
		Label:       "empty-perms",
		Permissions: "",
	}

	got := d.MapCreateRoleParams(input)

	if string(got.Permissions) != "" {
		t.Errorf("Permissions = %q, want empty", string(got.Permissions))
	}
}

func TestMysqlDatabase_MapCreateRoleParams_NullStringPermissions(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	// The literal string "null" should produce json.RawMessage("null")
	input := CreateRoleParams{
		Label:       "null-perms",
		Permissions: "null",
	}

	got := d.MapCreateRoleParams(input)

	if string(got.Permissions) != "null" {
		t.Errorf("Permissions = %q, want %q", string(got.Permissions), "null")
	}
}

// --- MySQL MysqlDatabase.MapUpdateRoleParams tests ---

func TestMysqlDatabase_MapUpdateRoleParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	roleID := types.NewRoleID()

	input := UpdateRoleParams{
		Label:       "mysqlupdated",
		Permissions: `{"updated":true}`,
		RoleID:      roleID,
	}

	got := d.MapUpdateRoleParams(input)

	if got.Label != "mysqlupdated" {
		t.Errorf("Label = %q, want %q", got.Label, "mysqlupdated")
	}
	if got.RoleID != roleID {
		t.Errorf("RoleID = %v, want %v", got.RoleID, roleID)
	}
	if len(got.Permissions) == 0 {
		t.Fatal("Permissions is empty, want non-empty for non-empty string")
	}
	if string(got.Permissions) != `{"updated":true}` {
		t.Errorf("Permissions = %q, want %q", string(got.Permissions), `{"updated":true}`)
	}
}

// --- PostgreSQL PsqlDatabase.MapRole tests ---

func TestPsqlDatabase_MapRole_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	roleID := types.NewRoleID()

	input := mdbp.Roles{
		RoleID:      roleID,
		Label:       "admin",
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(`{"all":true}`), Valid: true},
	}

	got := d.MapRole(input)

	if got.RoleID != roleID {
		t.Errorf("RoleID = %v, want %v", got.RoleID, roleID)
	}
	if got.Label != "admin" {
		t.Errorf("Label = %q, want %q", got.Label, "admin")
	}
	if got.Permissions != `{"all":true}` {
		t.Errorf("Permissions = %q, want %q", got.Permissions, `{"all":true}`)
	}
}

func TestPsqlDatabase_MapRole_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapRole(mdbp.Roles{})

	if got.RoleID != "" {
		t.Errorf("RoleID = %v, want zero value", got.RoleID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	// NullRawMessage with nil RawMessage: string(nil) == ""
	if got.Permissions != "" {
		t.Errorf("Permissions = %q, want empty string", got.Permissions)
	}
}

func TestPsqlDatabase_MapRole_NullPermissions(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	// When Permissions is NULL in PostgreSQL, NullRawMessage has nil RawMessage
	input := mdbp.Roles{
		RoleID:      types.NewRoleID(),
		Label:       "viewer",
		Permissions: pqtype.NullRawMessage{RawMessage: nil, Valid: false},
	}

	got := d.MapRole(input)

	// The mapper uses string(a.Permissions.RawMessage) which is "" for nil
	if got.Permissions != "" {
		t.Errorf("Permissions = %q, want empty string for NULL permissions", got.Permissions)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateRoleParams tests ---

func TestPsqlDatabase_MapCreateRoleParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := CreateRoleParams{
		Label:       "psqlrole",
		Permissions: `{"read":true}`,
	}

	got := d.MapCreateRoleParams(input)

	if got.RoleID.IsZero() {
		t.Fatal("expected non-zero RoleID to be generated")
	}
	if got.Label != "psqlrole" {
		t.Errorf("Label = %q, want %q", got.Label, "psqlrole")
	}
	// Permissions should be wrapped in NullRawMessage via json.RawMessage
	wantJSON := `{"read":true}`
	if string(got.Permissions.RawMessage) != wantJSON {
		t.Errorf("Permissions.RawMessage = %q, want %q", string(got.Permissions.RawMessage), wantJSON)
	}
}

func TestPsqlDatabase_MapCreateRoleParams_EmptyPermissions(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	// Empty permissions: json.RawMessage("") is a non-nil empty RawMessage
	input := CreateRoleParams{
		Label:       "empty-perms",
		Permissions: "",
	}

	got := d.MapCreateRoleParams(input)

	if string(got.Permissions.RawMessage) != "" {
		t.Errorf("Permissions.RawMessage = %q, want empty", string(got.Permissions.RawMessage))
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateRoleParams tests ---

func TestPsqlDatabase_MapUpdateRoleParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	roleID := types.NewRoleID()

	input := UpdateRoleParams{
		Label:       "psqlupdated",
		Permissions: `{"updated":true}`,
		RoleID:      roleID,
	}

	got := d.MapUpdateRoleParams(input)

	if got.Label != "psqlupdated" {
		t.Errorf("Label = %q, want %q", got.Label, "psqlupdated")
	}
	if got.RoleID != roleID {
		t.Errorf("RoleID = %v, want %v", got.RoleID, roleID)
	}
	wantJSON := `{"updated":true}`
	if string(got.Permissions.RawMessage) != wantJSON {
		t.Errorf("Permissions.RawMessage = %q, want %q", string(got.Permissions.RawMessage), wantJSON)
	}
}

// --- Cross-database mapper consistency ---
// Verifies that all three database mappers produce identical Roles from equivalent input.

func TestCrossDatabaseMapRole_Consistency(t *testing.T) {
	t.Parallel()
	roleID := types.NewRoleID()
	perms := `{"read":true,"write":false}`

	sqliteInput := mdb.Roles{
		RoleID: roleID, Label: "crossrole", Permissions: perms,
	}
	mysqlInput := mdbm.Roles{
		RoleID: roleID, Label: "crossrole",
		Permissions: json.RawMessage(perms),
	}
	psqlInput := mdbp.Roles{
		RoleID: roleID, Label: "crossrole",
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(perms), Valid: true},
	}

	sqliteResult := Database{}.MapRole(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapRole(mysqlInput)
	psqlResult := PsqlDatabase{}.MapRole(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateRoleParams consistency ---
// All three should auto-generate unique IDs.

func TestCrossDatabaseMapCreateRoleParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()

	input := CreateRoleParams{
		Label:       "crosscreate",
		Permissions: `{"cross":true}`,
	}

	sqliteResult := Database{}.MapCreateRoleParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateRoleParams(input)
	psqlResult := PsqlDatabase{}.MapCreateRoleParams(input)

	if sqliteResult.RoleID.IsZero() {
		t.Error("SQLite: expected non-zero generated RoleID")
	}
	if mysqlResult.RoleID.IsZero() {
		t.Error("MySQL: expected non-zero generated RoleID")
	}
	if psqlResult.RoleID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated RoleID")
	}

	// Each call should generate a unique ID
	if sqliteResult.RoleID == mysqlResult.RoleID {
		t.Error("SQLite and MySQL generated the same RoleID -- each call should be unique")
	}
	if sqliteResult.RoleID == psqlResult.RoleID {
		t.Error("SQLite and PostgreSQL generated the same RoleID -- each call should be unique")
	}
}

// --- Permissions field mapping edge cases ---
// The three drivers store permissions differently: plain string (SQLite),
// json.RawMessage (MySQL), pqtype.NullRawMessage (PostgreSQL).
// This test verifies correct round-trip behavior for various input shapes.

func TestPermissionsFieldMapping_AllDatabases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		perms string
	}{
		{"valid JSON object", `{"read":true,"write":false}`},
		{"empty string", ""},
		{"JSON array", `["read","write","delete"]`},
		{"nested JSON", `{"admin":{"read":true,"write":true},"viewer":{"read":true}}`},
		// Unicode in permissions values
		{"unicode permissions", `{"label":"edici\u00f3n","active":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// SQLite: plain string round-trip
			sqliteResult := Database{}.MapRole(mdb.Roles{Permissions: tt.perms})
			if sqliteResult.Permissions != tt.perms {
				t.Errorf("SQLite MapRole: Permissions = %q, want %q", sqliteResult.Permissions, tt.perms)
			}

			// MySQL: NullString.String always returns the String field
			mysqlInput := mdbm.Roles{
				Permissions: json.RawMessage(tt.perms),
			}
			mysqlResult := MysqlDatabase{}.MapRole(mysqlInput)
			if mysqlResult.Permissions != tt.perms {
				t.Errorf("MySQL MapRole: Permissions = %q, want %q", mysqlResult.Permissions, tt.perms)
			}

			// PostgreSQL: json.RawMessage conversion
			psqlInput := mdbp.Roles{
				Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(tt.perms), Valid: tt.perms != ""},
			}
			psqlResult := PsqlDatabase{}.MapRole(psqlInput)
			if psqlResult.Permissions != tt.perms {
				t.Errorf("PostgreSQL MapRole: Permissions = %q, want %q", psqlResult.Permissions, tt.perms)
			}

			// Reverse: wrapper -> driver via MapCreateRoleParams
			createInput := CreateRoleParams{Permissions: tt.perms}
			sqliteCreate := Database{}.MapCreateRoleParams(createInput)
			if sqliteCreate.Permissions != tt.perms {
				t.Errorf("SQLite MapCreateRoleParams: Permissions = %q, want %q", sqliteCreate.Permissions, tt.perms)
			}

			// MySQL: wraps in json.RawMessage directly
			mysqlCreate := MysqlDatabase{}.MapCreateRoleParams(createInput)
			if string(mysqlCreate.Permissions) != tt.perms {
				t.Errorf("MySQL MapCreateRoleParams: Permissions = %q, want %q", string(mysqlCreate.Permissions), tt.perms)
			}

			// PostgreSQL: wraps in json.RawMessage
			psqlCreate := PsqlDatabase{}.MapCreateRoleParams(createInput)
			if string(psqlCreate.Permissions.RawMessage) != tt.perms {
				t.Errorf("PostgreSQL MapCreateRoleParams: Permissions.RawMessage = %q, want %q",
					string(psqlCreate.Permissions.RawMessage), tt.perms)
			}
		})
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewRoleCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-123",
		IP:        "10.0.0.1",
	}
	params := CreateRoleParams{
		Label:       "cmdrole",
		Permissions: `{"cmd":true}`,
	}

	cmd := Database{}.NewRoleCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "roles" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "roles")
	}
	p, ok := cmd.Params().(CreateRoleParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateRoleParams", cmd.Params())
	}
	if p.Label != "cmdrole" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "cmdrole")
	}
	if p.Permissions != `{"cmd":true}` {
		t.Errorf("Params().Permissions = %q, want %q", p.Permissions, `{"cmd":true}`)
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewRoleCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	roleID := types.NewRoleID()
	cmd := NewRoleCmd{}

	row := mdb.Roles{RoleID: roleID}
	got := cmd.GetID(row)
	if got != string(roleID) {
		t.Errorf("GetID() = %q, want %q", got, string(roleID))
	}
}

func TestNewRoleCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewRoleCmd{}

	row := mdb.Roles{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateRoleCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	roleID := types.NewRoleID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateRoleParams{
		Label:       "updated",
		Permissions: `{"updated":true}`,
		RoleID:      roleID,
	}

	cmd := Database{}.UpdateRoleCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "roles" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "roles")
	}
	if cmd.GetID() != string(roleID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(roleID))
	}
	p, ok := cmd.Params().(UpdateRoleParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateRoleParams", cmd.Params())
	}
	if p.Label != "updated" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteRoleCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	roleID := types.NewRoleID()

	cmd := Database{}.DeleteRoleCmd(ctx, ac, roleID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "roles" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "roles")
	}
	if cmd.GetID() != string(roleID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(roleID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewRoleCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-req",
		IP:        "192.168.1.1",
	}
	params := CreateRoleParams{
		Label:       "mysqlcmdrole",
		Permissions: `{"mysql":true}`,
	}

	cmd := MysqlDatabase{}.NewRoleCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "roles" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "roles")
	}
	p, ok := cmd.Params().(CreateRoleParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateRoleParams", cmd.Params())
	}
	if p.Label != "mysqlcmdrole" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "mysqlcmdrole")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewRoleCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	roleID := types.NewRoleID()
	cmd := NewRoleCmdMysql{}

	row := mdbm.Roles{RoleID: roleID}
	got := cmd.GetID(row)
	if got != string(roleID) {
		t.Errorf("GetID() = %q, want %q", got, string(roleID))
	}
}

func TestUpdateRoleCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	roleID := types.NewRoleID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateRoleParams{
		Label:       "mysqlupdated",
		Permissions: `{"mysqlupdated":true}`,
		RoleID:      roleID,
	}

	cmd := MysqlDatabase{}.UpdateRoleCmd(ctx, ac, params)

	if cmd.TableName() != "roles" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "roles")
	}
	if cmd.GetID() != string(roleID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(roleID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateRoleParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateRoleParams", cmd.Params())
	}
	if p.Label != "mysqlupdated" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "mysqlupdated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteRoleCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	roleID := types.NewRoleID()

	cmd := MysqlDatabase{}.DeleteRoleCmd(ctx, ac, roleID)

	if cmd.TableName() != "roles" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "roles")
	}
	if cmd.GetID() != string(roleID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(roleID))
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

func TestNewRoleCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-req",
		IP:        "172.16.0.1",
	}
	params := CreateRoleParams{
		Label:       "psqlcmdrole",
		Permissions: `{"psql":true}`,
	}

	cmd := PsqlDatabase{}.NewRoleCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "roles" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "roles")
	}
	p, ok := cmd.Params().(CreateRoleParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateRoleParams", cmd.Params())
	}
	if p.Label != "psqlcmdrole" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "psqlcmdrole")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewRoleCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	roleID := types.NewRoleID()
	cmd := NewRoleCmdPsql{}

	row := mdbp.Roles{RoleID: roleID}
	got := cmd.GetID(row)
	if got != string(roleID) {
		t.Errorf("GetID() = %q, want %q", got, string(roleID))
	}
}

func TestUpdateRoleCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	roleID := types.NewRoleID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateRoleParams{
		Label:       "psqlupdated",
		Permissions: `{"psqlupdated":true}`,
		RoleID:      roleID,
	}

	cmd := PsqlDatabase{}.UpdateRoleCmd(ctx, ac, params)

	if cmd.TableName() != "roles" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "roles")
	}
	if cmd.GetID() != string(roleID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(roleID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateRoleParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateRoleParams", cmd.Params())
	}
	if p.Label != "psqlupdated" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "psqlupdated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteRoleCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	roleID := types.NewRoleID()

	cmd := PsqlDatabase{}.DeleteRoleCmd(ctx, ac, roleID)

	if cmd.TableName() != "roles" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "roles")
	}
	if cmd.GetID() != string(roleID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(roleID))
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

func TestAuditedRoleCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateRoleParams{}
	updateParams := UpdateRoleParams{RoleID: types.NewRoleID()}
	roleID := types.NewRoleID()

	// SQLite
	sqliteCreate := Database{}.NewRoleCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateRoleCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteRoleCmd(ctx, ac, roleID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewRoleCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateRoleCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteRoleCmd(ctx, ac, roleID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewRoleCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateRoleCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteRoleCmd(ctx, ac, roleID)

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
			if c.name != "roles" {
				t.Errorf("TableName() = %q, want %q", c.name, "roles")
			}
		})
	}
}

func TestAuditedRoleCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateRoleParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewRoleCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewRoleCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewRoleCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedRoleCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	roleID := types.NewRoleID()

	t.Run("UpdateCmd GetID returns RoleID", func(t *testing.T) {
		t.Parallel()
		params := UpdateRoleParams{RoleID: roleID}

		sqliteCmd := Database{}.UpdateRoleCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateRoleCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateRoleCmd(ctx, ac, params)

		wantID := string(roleID)
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

	t.Run("DeleteCmd GetID returns RoleID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteRoleCmd(ctx, ac, roleID)
		mysqlCmd := MysqlDatabase{}.DeleteRoleCmd(ctx, ac, roleID)
		psqlCmd := PsqlDatabase{}.DeleteRoleCmd(ctx, ac, roleID)

		wantID := string(roleID)
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
		testRoleID := types.NewRoleID()

		sqliteCmd := NewRoleCmd{}
		mysqlCmd := NewRoleCmdMysql{}
		psqlCmd := NewRoleCmdPsql{}

		wantID := string(testRoleID)

		sqliteRow := mdb.Roles{RoleID: testRoleID}
		mysqlRow := mdbm.Roles{RoleID: testRoleID}
		psqlRow := mdbp.Roles{RoleID: testRoleID}

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

// --- Edge case: UpdateRoleCmd with empty RoleID ---

func TestUpdateRoleCmd_GetID_EmptyRoleID(t *testing.T) {
	t.Parallel()
	// When RoleID is empty, GetID should return empty string
	params := UpdateRoleParams{RoleID: ""}

	sqliteCmd := Database{}.UpdateRoleCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateRoleCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateRoleCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteRoleCmd with empty ID ---

func TestDeleteRoleCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.RoleID("")

	sqliteCmd := Database{}.DeleteRoleCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteRoleCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteRoleCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.Roles]  = NewRoleCmd{}
	_ audited.UpdateCommand[mdb.Roles]  = UpdateRoleCmd{}
	_ audited.DeleteCommand[mdb.Roles]  = DeleteRoleCmd{}
	_ audited.CreateCommand[mdbm.Roles] = NewRoleCmdMysql{}
	_ audited.UpdateCommand[mdbm.Roles] = UpdateRoleCmdMysql{}
	_ audited.DeleteCommand[mdbm.Roles] = DeleteRoleCmdMysql{}
	_ audited.CreateCommand[mdbp.Roles] = NewRoleCmdPsql{}
	_ audited.UpdateCommand[mdbp.Roles] = UpdateRoleCmdPsql{}
	_ audited.DeleteCommand[mdbp.Roles] = DeleteRoleCmdPsql{}
)

// --- Struct field correctness ---
// Verify that the wrapper Roles struct and param structs hold values correctly.

func TestRolesStruct_JSONTags(t *testing.T) {
	t.Parallel()
	// Verify that creating a Roles struct and marshaling it preserves field names
	roleID := types.NewRoleID()
	r := Roles{
		RoleID:      roleID,
		Label:       "test-role",
		Permissions: `{"key":"value"}`,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify JSON field names match the struct tags
	expectedFields := []string{"role_id", "label", "permissions"}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}

	// Verify no extra fields
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

func TestCreateRoleParams_JSONTags(t *testing.T) {
	t.Parallel()
	p := CreateRoleParams{
		Label:       "new",
		Permissions: `{"p":1}`,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{"label", "permissions"}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}

func TestUpdateRoleParams_JSONTags(t *testing.T) {
	t.Parallel()
	roleID := types.NewRoleID()
	p := UpdateRoleParams{
		Label:       "updated",
		Permissions: `{"p":2}`,
		RoleID:      roleID,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{"label", "permissions", "role_id"}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON output missing expected field %q", field)
		}
	}
	if len(m) != len(expectedFields) {
		t.Errorf("JSON output has %d fields, want %d", len(m), len(expectedFields))
	}
}
