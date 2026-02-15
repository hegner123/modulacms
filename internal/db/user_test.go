package db

import (
	"context"
	"testing"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// userTestFixture returns a fully populated Users struct and its constituent parts for testing.
func userTestFixture() (Users, types.UserID, types.Email, types.Timestamp) {
	userID := types.NewUserID()
	email := types.Email("alice@example.com")
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))
	user := Users{
		UserID:       userID,
		Username:     "alice",
		Name:         "Alice Smith",
		Email:        email,
		Hash:         "$2a$10$somehashvalue",
		Role:         "admin",
		DateCreated:  ts,
		DateModified: ts,
	}
	return user, userID, email, ts
}

// --- MapStringUser tests ---

func TestMapStringUser_AllFields(t *testing.T) {
	t.Parallel()
	user, userID, email, ts := userTestFixture()

	got := MapStringUser(user)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"UserID", got.UserID, userID.String()},
		{"Username", got.Username, "alice"},
		{"Name", got.Name, "Alice Smith"},
		{"Email", got.Email, email.String()},
		{"Hash", got.Hash, "$2a$10$somehashvalue"},
		{"Role", got.Role, "admin"},
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

func TestMapStringUser_ZeroValueUser(t *testing.T) {
	t.Parallel()
	// Zero-value Users: all fields should convert without panic
	got := MapStringUser(Users{})

	if got.UserID != "" {
		t.Errorf("UserID = %q, want empty string", got.UserID)
	}
	if got.Username != "" {
		t.Errorf("Username = %q, want empty string", got.Username)
	}
	if got.Name != "" {
		t.Errorf("Name = %q, want empty string", got.Name)
	}
	if got.Email != "" {
		t.Errorf("Email = %q, want empty string", got.Email)
	}
	if got.Hash != "" {
		t.Errorf("Hash = %q, want empty string", got.Hash)
	}
	if got.Role != "" {
		t.Errorf("Role = %q, want empty string", got.Role)
	}
}

func TestMapStringUser_UnicodeFields(t *testing.T) {
	t.Parallel()
	// Verify multi-byte characters survive the mapping
	user := Users{
		Username: "usuario",
		Name:     "Jose Garcia",
		Email:    types.Email("jose@ejemplo.es"),
		Role:     "editor",
	}
	got := MapStringUser(user)
	if got.Name != "Jose Garcia" {
		t.Errorf("Name = %q, want %q", got.Name, "Jose Garcia")
	}
	if got.Email != "jose@ejemplo.es" {
		t.Errorf("Email = %q, want %q", got.Email, "jose@ejemplo.es")
	}
}

// --- SQLite Database.MapUser tests ---

func TestDatabase_MapUser_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	userID := types.NewUserID()
	email := types.Email("bob@example.com")
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	input := mdb.Users{
		UserID:       userID,
		Username:     "bob",
		Name:         "Bob Jones",
		Email:        email,
		Hash:         "hash123",
		Roles:        "editor",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapUser(input)

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Username != "bob" {
		t.Errorf("Username = %q, want %q", got.Username, "bob")
	}
	if got.Name != "Bob Jones" {
		t.Errorf("Name = %q, want %q", got.Name, "Bob Jones")
	}
	if got.Email != email {
		t.Errorf("Email = %v, want %v", got.Email, email)
	}
	if got.Hash != "hash123" {
		t.Errorf("Hash = %q, want %q", got.Hash, "hash123")
	}
	// The sqlc model uses "Roles" but the wrapper uses "Role" -- verify mapping
	if got.Role != "editor" {
		t.Errorf("Role = %q, want %q", got.Role, "editor")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestDatabase_MapUser_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUser(mdb.Users{})

	if got.UserID != "" {
		t.Errorf("UserID = %v, want zero value", got.UserID)
	}
	if got.Username != "" {
		t.Errorf("Username = %q, want empty string", got.Username)
	}
	if got.Role != "" {
		t.Errorf("Role = %q, want empty string", got.Role)
	}
}

// --- SQLite Database.MapCreateUserParams tests ---

func TestDatabase_MapCreateUserParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateUserParams{
		Username:     "newuser",
		Name:         "New User",
		Email:        types.Email("new@example.com"),
		Hash:         "newhash",
		Role:         "viewer",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateUserParams(input)

	// MapCreateUserParams always generates a new UserID via types.NewUserID()
	if got.UserID.IsZero() {
		t.Fatal("expected non-zero UserID to be generated")
	}
	if got.Username != "newuser" {
		t.Errorf("Username = %q, want %q", got.Username, "newuser")
	}
	if got.Name != "New User" {
		t.Errorf("Name = %q, want %q", got.Name, "New User")
	}
	if got.Email != types.Email("new@example.com") {
		t.Errorf("Email = %v, want %v", got.Email, types.Email("new@example.com"))
	}
	if got.Hash != "newhash" {
		t.Errorf("Hash = %q, want %q", got.Hash, "newhash")
	}
	// Wrapper uses "Role" but sqlc expects "Roles"
	if got.Roles != "viewer" {
		t.Errorf("Roles = %q, want %q", got.Roles, "viewer")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestDatabase_MapCreateUserParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateUserParams{Username: "test"}

	// Two calls should produce different IDs
	got1 := d.MapCreateUserParams(input)
	got2 := d.MapCreateUserParams(input)
	if got1.UserID == got2.UserID {
		t.Error("two calls produced the same UserID -- each call should generate a unique ID")
	}
}

// --- SQLite Database.MapUpdateUserParams tests ---

func TestDatabase_MapUpdateUserParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	userID := types.NewUserID()

	input := UpdateUserParams{
		Username:     "updated",
		Name:         "Updated Name",
		Email:        types.Email("updated@example.com"),
		Hash:         "updatedhash",
		Role:         "admin",
		DateCreated:  ts,
		DateModified: ts,
		UserID:       userID,
	}

	got := d.MapUpdateUserParams(input)

	if got.Username != "updated" {
		t.Errorf("Username = %q, want %q", got.Username, "updated")
	}
	if got.Name != "Updated Name" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated Name")
	}
	if got.Email != types.Email("updated@example.com") {
		t.Errorf("Email = %v, want %v", got.Email, types.Email("updated@example.com"))
	}
	if got.Hash != "updatedhash" {
		t.Errorf("Hash = %q, want %q", got.Hash, "updatedhash")
	}
	if got.Roles != "admin" {
		t.Errorf("Roles = %q, want %q", got.Roles, "admin")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
}

// --- MySQL MysqlDatabase.MapUser tests ---

func TestMysqlDatabase_MapUser_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	userID := types.NewUserID()
	email := types.Email("mysql@example.com")
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))

	input := mdbm.Users{
		UserID:       userID,
		Username:     "mysqluser",
		Name:         "MySQL User",
		Email:        email,
		Hash:         "mysqlhash",
		Roles:        "admin",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapUser(input)

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Username != "mysqluser" {
		t.Errorf("Username = %q, want %q", got.Username, "mysqluser")
	}
	if got.Name != "MySQL User" {
		t.Errorf("Name = %q, want %q", got.Name, "MySQL User")
	}
	if got.Email != email {
		t.Errorf("Email = %v, want %v", got.Email, email)
	}
	if got.Hash != "mysqlhash" {
		t.Errorf("Hash = %q, want %q", got.Hash, "mysqlhash")
	}
	if got.Role != "admin" {
		t.Errorf("Role = %q, want %q", got.Role, "admin")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestMysqlDatabase_MapUser_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapUser(mdbm.Users{})

	if got.UserID != "" {
		t.Errorf("UserID = %v, want zero value", got.UserID)
	}
	if got.Role != "" {
		t.Errorf("Role = %q, want empty string", got.Role)
	}
}

// --- MySQL MysqlDatabase.MapCreateUserParams tests ---

func TestMysqlDatabase_MapCreateUserParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateUserParams{
		Username:     "mysqlnew",
		Name:         "MySQL New",
		Email:        types.Email("mysqlnew@example.com"),
		Hash:         "mysqlnewhash",
		Role:         "editor",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateUserParams(input)

	if got.UserID.IsZero() {
		t.Fatal("expected non-zero UserID to be generated")
	}
	if got.Username != "mysqlnew" {
		t.Errorf("Username = %q, want %q", got.Username, "mysqlnew")
	}
	if got.Roles != "editor" {
		t.Errorf("Roles = %q, want %q", got.Roles, "editor")
	}
	if got.Email != types.Email("mysqlnew@example.com") {
		t.Errorf("Email = %v, want %v", got.Email, types.Email("mysqlnew@example.com"))
	}
}

// --- MySQL MysqlDatabase.MapUpdateUserParams tests ---

func TestMysqlDatabase_MapUpdateUserParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	userID := types.NewUserID()

	input := UpdateUserParams{
		Username:     "mysqlupdated",
		Name:         "MySQL Updated",
		Email:        types.Email("mysqlupdated@example.com"),
		Hash:         "mysqlupdatedhash",
		Role:         "viewer",
		DateCreated:  ts,
		DateModified: ts,
		UserID:       userID,
	}

	got := d.MapUpdateUserParams(input)

	if got.Username != "mysqlupdated" {
		t.Errorf("Username = %q, want %q", got.Username, "mysqlupdated")
	}
	if got.Roles != "viewer" {
		t.Errorf("Roles = %q, want %q", got.Roles, "viewer")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Email != types.Email("mysqlupdated@example.com") {
		t.Errorf("Email = %v, want %v", got.Email, types.Email("mysqlupdated@example.com"))
	}
}

// --- PostgreSQL PsqlDatabase.MapUser tests ---

func TestPsqlDatabase_MapUser_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	userID := types.NewUserID()
	email := types.Email("psql@example.com")
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))

	input := mdbp.Users{
		UserID:       userID,
		Username:     "psqluser",
		Name:         "PostgreSQL User",
		Email:        email,
		Hash:         "psqlhash",
		Roles:        "superadmin",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapUser(input)

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Username != "psqluser" {
		t.Errorf("Username = %q, want %q", got.Username, "psqluser")
	}
	if got.Name != "PostgreSQL User" {
		t.Errorf("Name = %q, want %q", got.Name, "PostgreSQL User")
	}
	if got.Email != email {
		t.Errorf("Email = %v, want %v", got.Email, email)
	}
	if got.Hash != "psqlhash" {
		t.Errorf("Hash = %q, want %q", got.Hash, "psqlhash")
	}
	if got.Role != "superadmin" {
		t.Errorf("Role = %q, want %q", got.Role, "superadmin")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
	if got.DateModified != ts {
		t.Errorf("DateModified = %v, want %v", got.DateModified, ts)
	}
}

func TestPsqlDatabase_MapUser_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapUser(mdbp.Users{})

	if got.UserID != "" {
		t.Errorf("UserID = %v, want zero value", got.UserID)
	}
	if got.Role != "" {
		t.Errorf("Role = %q, want empty string", got.Role)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateUserParams tests ---

func TestPsqlDatabase_MapCreateUserParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateUserParams{
		Username:     "psqlnew",
		Name:         "Psql New",
		Email:        types.Email("psqlnew@example.com"),
		Hash:         "psqlnewhash",
		Role:         "admin",
		DateCreated:  ts,
		DateModified: ts,
	}

	got := d.MapCreateUserParams(input)

	if got.UserID.IsZero() {
		t.Fatal("expected non-zero UserID to be generated")
	}
	if got.Username != "psqlnew" {
		t.Errorf("Username = %q, want %q", got.Username, "psqlnew")
	}
	if got.Roles != "admin" {
		t.Errorf("Roles = %q, want %q", got.Roles, "admin")
	}
	if got.Email != types.Email("psqlnew@example.com") {
		t.Errorf("Email = %v, want %v", got.Email, types.Email("psqlnew@example.com"))
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateUserParams tests ---

func TestPsqlDatabase_MapUpdateUserParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	userID := types.NewUserID()

	input := UpdateUserParams{
		Username:     "psqlupdated",
		Name:         "Psql Updated",
		Email:        types.Email("psqlupdated@example.com"),
		Hash:         "psqlupdatedhash",
		Role:         "editor",
		DateCreated:  ts,
		DateModified: ts,
		UserID:       userID,
	}

	got := d.MapUpdateUserParams(input)

	if got.Username != "psqlupdated" {
		t.Errorf("Username = %q, want %q", got.Username, "psqlupdated")
	}
	if got.Roles != "editor" {
		t.Errorf("Roles = %q, want %q", got.Roles, "editor")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Email != types.Email("psqlupdated@example.com") {
		t.Errorf("Email = %v, want %v", got.Email, types.Email("psqlupdated@example.com"))
	}
}

// --- Cross-database mapper consistency ---
// Verifies that all three database mappers produce identical Users from equivalent input.

func TestCrossDatabaseMapUser_Consistency(t *testing.T) {
	t.Parallel()
	userID := types.NewUserID()
	email := types.Email("cross@example.com")
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))

	sqliteInput := mdb.Users{
		UserID: userID, Username: "crossuser", Name: "Cross DB",
		Email: email, Hash: "crosshash", Roles: "admin",
		DateCreated: ts, DateModified: ts,
	}
	mysqlInput := mdbm.Users{
		UserID: userID, Username: "crossuser", Name: "Cross DB",
		Email: email, Hash: "crosshash", Roles: "admin",
		DateCreated: ts, DateModified: ts,
	}
	psqlInput := mdbp.Users{
		UserID: userID, Username: "crossuser", Name: "Cross DB",
		Email: email, Hash: "crosshash", Roles: "admin",
		DateCreated: ts, DateModified: ts,
	}

	sqliteResult := Database{}.MapUser(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapUser(mysqlInput)
	psqlResult := PsqlDatabase{}.MapUser(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateUserParams consistency ---
// All three should auto-generate unique IDs.

func TestCrossDatabaseMapCreateUserParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))

	input := CreateUserParams{
		Username:     "crosscreate",
		Name:         "Cross Create",
		Email:        types.Email("crosscreate@example.com"),
		Hash:         "crosshash",
		Role:         "editor",
		DateCreated:  ts,
		DateModified: ts,
	}

	sqliteResult := Database{}.MapCreateUserParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateUserParams(input)
	psqlResult := PsqlDatabase{}.MapCreateUserParams(input)

	if sqliteResult.UserID.IsZero() {
		t.Error("SQLite: expected non-zero generated UserID")
	}
	if mysqlResult.UserID.IsZero() {
		t.Error("MySQL: expected non-zero generated UserID")
	}
	if psqlResult.UserID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated UserID")
	}

	// Each call should generate a unique ID
	if sqliteResult.UserID == mysqlResult.UserID {
		t.Error("SQLite and MySQL generated the same UserID -- each call should be unique")
	}
	if sqliteResult.UserID == psqlResult.UserID {
		t.Error("SQLite and PostgreSQL generated the same UserID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewUserCmd_AllAccessors(t *testing.T) {
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
	params := CreateUserParams{
		Username:     "cmduser",
		Name:         "Cmd User",
		Email:        types.Email("cmd@example.com"),
		Hash:         "cmdhash",
		Role:         "admin",
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := Database{}.NewUserCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	p, ok := cmd.Params().(CreateUserParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateUserParams", cmd.Params())
	}
	if p.Username != "cmduser" {
		t.Errorf("Params().Username = %q, want %q", p.Username, "cmduser")
	}
	if p.Email != types.Email("cmd@example.com") {
		t.Errorf("Params().Email = %v, want %v", p.Email, types.Email("cmd@example.com"))
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewUserCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	userID := types.NewUserID()
	cmd := NewUserCmd{}

	row := mdb.Users{UserID: userID}
	got := cmd.GetID(row)
	if got != string(userID) {
		t.Errorf("GetID() = %q, want %q", got, string(userID))
	}
}

func TestNewUserCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewUserCmd{}

	row := mdb.Users{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateUserCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateUserParams{
		Username:     "updated",
		Name:         "Updated User",
		Email:        types.Email("updated@example.com"),
		Hash:         "updatedhash",
		Role:         "editor",
		DateCreated:  ts,
		DateModified: ts,
		UserID:       userID,
	}

	cmd := Database{}.UpdateUserCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	if cmd.GetID() != string(userID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(userID))
	}
	p, ok := cmd.Params().(UpdateUserParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateUserParams", cmd.Params())
	}
	if p.Username != "updated" {
		t.Errorf("Params().Username = %q, want %q", p.Username, "updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteUserCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	userID := types.NewUserID()

	cmd := Database{}.DeleteUserCmd(ctx, ac, userID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	if cmd.GetID() != string(userID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(userID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewUserCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-req",
		IP:        "192.168.1.1",
	}
	params := CreateUserParams{
		Username:     "mysqlcmduser",
		Name:         "MySQL Cmd User",
		Email:        types.Email("mysqlcmd@example.com"),
		Hash:         "mysqlcmdhash",
		Role:         "admin",
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := MysqlDatabase{}.NewUserCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	p, ok := cmd.Params().(CreateUserParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateUserParams", cmd.Params())
	}
	if p.Username != "mysqlcmduser" {
		t.Errorf("Params().Username = %q, want %q", p.Username, "mysqlcmduser")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewUserCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	userID := types.NewUserID()
	cmd := NewUserCmdMysql{}

	row := mdbm.Users{UserID: userID}
	got := cmd.GetID(row)
	if got != string(userID) {
		t.Errorf("GetID() = %q, want %q", got, string(userID))
	}
}

func TestUpdateUserCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateUserParams{
		Username:     "mysqlupdated",
		Name:         "MySQL Updated",
		Email:        types.Email("mysqlupdated@example.com"),
		Hash:         "mysqlupdatedhash",
		Role:         "viewer",
		DateCreated:  ts,
		DateModified: ts,
		UserID:       userID,
	}

	cmd := MysqlDatabase{}.UpdateUserCmd(ctx, ac, params)

	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	if cmd.GetID() != string(userID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(userID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateUserParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateUserParams", cmd.Params())
	}
	if p.Username != "mysqlupdated" {
		t.Errorf("Params().Username = %q, want %q", p.Username, "mysqlupdated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteUserCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	userID := types.NewUserID()

	cmd := MysqlDatabase{}.DeleteUserCmd(ctx, ac, userID)

	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	if cmd.GetID() != string(userID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(userID))
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

func TestNewUserCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-req",
		IP:        "172.16.0.1",
	}
	params := CreateUserParams{
		Username:     "psqlcmduser",
		Name:         "Psql Cmd User",
		Email:        types.Email("psqlcmd@example.com"),
		Hash:         "psqlcmdhash",
		Role:         "superadmin",
		DateCreated:  ts,
		DateModified: ts,
	}

	cmd := PsqlDatabase{}.NewUserCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	p, ok := cmd.Params().(CreateUserParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateUserParams", cmd.Params())
	}
	if p.Username != "psqlcmduser" {
		t.Errorf("Params().Username = %q, want %q", p.Username, "psqlcmduser")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewUserCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	userID := types.NewUserID()
	cmd := NewUserCmdPsql{}

	row := mdbp.Users{UserID: userID}
	got := cmd.GetID(row)
	if got != string(userID) {
		t.Errorf("GetID() = %q, want %q", got, string(userID))
	}
}

func TestUpdateUserCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateUserParams{
		Username:     "psqlupdated",
		Name:         "Psql Updated",
		Email:        types.Email("psqlupdated@example.com"),
		Hash:         "psqlupdatedhash",
		Role:         "editor",
		DateCreated:  ts,
		DateModified: ts,
		UserID:       userID,
	}

	cmd := PsqlDatabase{}.UpdateUserCmd(ctx, ac, params)

	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	if cmd.GetID() != string(userID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(userID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateUserParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateUserParams", cmd.Params())
	}
	if p.Username != "psqlupdated" {
		t.Errorf("Params().Username = %q, want %q", p.Username, "psqlupdated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteUserCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	userID := types.NewUserID()

	cmd := PsqlDatabase{}.DeleteUserCmd(ctx, ac, userID)

	if cmd.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "users")
	}
	if cmd.GetID() != string(userID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(userID))
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

func TestAuditedUserCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateUserParams{}
	updateParams := UpdateUserParams{UserID: types.NewUserID()}
	userID := types.NewUserID()

	// SQLite
	sqliteCreate := Database{}.NewUserCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateUserCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteUserCmd(ctx, ac, userID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewUserCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateUserCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteUserCmd(ctx, ac, userID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewUserCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateUserCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteUserCmd(ctx, ac, userID)

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
			if c.name != "users" {
				t.Errorf("TableName() = %q, want %q", c.name, "users")
			}
		})
	}
}

func TestAuditedUserCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateUserParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewUserCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewUserCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewUserCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedUserCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	userID := types.NewUserID()

	t.Run("UpdateCmd GetID returns UserID", func(t *testing.T) {
		t.Parallel()
		params := UpdateUserParams{UserID: userID}

		sqliteCmd := Database{}.UpdateUserCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateUserCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateUserCmd(ctx, ac, params)

		wantID := string(userID)
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

	t.Run("DeleteCmd GetID returns UserID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteUserCmd(ctx, ac, userID)
		mysqlCmd := MysqlDatabase{}.DeleteUserCmd(ctx, ac, userID)
		psqlCmd := PsqlDatabase{}.DeleteUserCmd(ctx, ac, userID)

		wantID := string(userID)
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
		testUserID := types.NewUserID()

		sqliteCmd := NewUserCmd{}
		mysqlCmd := NewUserCmdMysql{}
		psqlCmd := NewUserCmdPsql{}

		wantID := string(testUserID)

		sqliteRow := mdb.Users{UserID: testUserID}
		mysqlRow := mdbm.Users{UserID: testUserID}
		psqlRow := mdbp.Users{UserID: testUserID}

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

// --- Edge case: UpdateUserCmd with empty UserID ---

func TestUpdateUserCmd_GetID_EmptyUserID(t *testing.T) {
	t.Parallel()
	// When UserID is empty, GetID should return empty string
	params := UpdateUserParams{UserID: ""}

	sqliteCmd := Database{}.UpdateUserCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateUserCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateUserCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteUserCmd with empty ID ---

func TestDeleteUserCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.UserID("")

	sqliteCmd := Database{}.DeleteUserCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteUserCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteUserCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.Users]  = NewUserCmd{}
	_ audited.UpdateCommand[mdb.Users]  = UpdateUserCmd{}
	_ audited.DeleteCommand[mdb.Users]  = DeleteUserCmd{}
	_ audited.CreateCommand[mdbm.Users] = NewUserCmdMysql{}
	_ audited.UpdateCommand[mdbm.Users] = UpdateUserCmdMysql{}
	_ audited.DeleteCommand[mdbm.Users] = DeleteUserCmdMysql{}
	_ audited.CreateCommand[mdbp.Users] = NewUserCmdPsql{}
	_ audited.UpdateCommand[mdbp.Users] = UpdateUserCmdPsql{}
	_ audited.DeleteCommand[mdbp.Users] = DeleteUserCmdPsql{}
)

// --- Role/Roles field mapping consistency ---
// The wrapper uses "Role" but all three sqlc-generated models use "Roles".
// This test verifies that mapping happens correctly in both directions.

func TestRolesToRoleFieldMapping_AllDatabases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		role string
	}{
		{"simple role", "admin"},
		{"empty role", ""},
		{"role with spaces", "super admin"},
		{"comma-separated roles", "admin,editor,viewer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// SQLite: sqlc "Roles" -> wrapper "Role"
			sqliteResult := Database{}.MapUser(mdb.Users{Roles: tt.role})
			if sqliteResult.Role != tt.role {
				t.Errorf("SQLite MapUser: Role = %q, want %q", sqliteResult.Role, tt.role)
			}

			// MySQL: sqlc "Roles" -> wrapper "Role"
			mysqlResult := MysqlDatabase{}.MapUser(mdbm.Users{Roles: tt.role})
			if mysqlResult.Role != tt.role {
				t.Errorf("MySQL MapUser: Role = %q, want %q", mysqlResult.Role, tt.role)
			}

			// PostgreSQL: sqlc "Roles" -> wrapper "Role"
			psqlResult := PsqlDatabase{}.MapUser(mdbp.Users{Roles: tt.role})
			if psqlResult.Role != tt.role {
				t.Errorf("PostgreSQL MapUser: Role = %q, want %q", psqlResult.Role, tt.role)
			}

			// Reverse: wrapper "Role" -> sqlc "Roles" (via MapCreateUserParams)
			createInput := CreateUserParams{Role: tt.role}
			sqliteCreate := Database{}.MapCreateUserParams(createInput)
			if sqliteCreate.Roles != tt.role {
				t.Errorf("SQLite MapCreateUserParams: Roles = %q, want %q", sqliteCreate.Roles, tt.role)
			}
			mysqlCreate := MysqlDatabase{}.MapCreateUserParams(createInput)
			if mysqlCreate.Roles != tt.role {
				t.Errorf("MySQL MapCreateUserParams: Roles = %q, want %q", mysqlCreate.Roles, tt.role)
			}
			psqlCreate := PsqlDatabase{}.MapCreateUserParams(createInput)
			if psqlCreate.Roles != tt.role {
				t.Errorf("PostgreSQL MapCreateUserParams: Roles = %q, want %q", psqlCreate.Roles, tt.role)
			}

			// Reverse: wrapper "Role" -> sqlc "Roles" (via MapUpdateUserParams)
			updateInput := UpdateUserParams{Role: tt.role, UserID: types.NewUserID()}
			sqliteUpdate := Database{}.MapUpdateUserParams(updateInput)
			if sqliteUpdate.Roles != tt.role {
				t.Errorf("SQLite MapUpdateUserParams: Roles = %q, want %q", sqliteUpdate.Roles, tt.role)
			}
			mysqlUpdate := MysqlDatabase{}.MapUpdateUserParams(updateInput)
			if mysqlUpdate.Roles != tt.role {
				t.Errorf("MySQL MapUpdateUserParams: Roles = %q, want %q", mysqlUpdate.Roles, tt.role)
			}
			psqlUpdate := PsqlDatabase{}.MapUpdateUserParams(updateInput)
			if psqlUpdate.Roles != tt.role {
				t.Errorf("PostgreSQL MapUpdateUserParams: Roles = %q, want %q", psqlUpdate.Roles, tt.role)
			}
		})
	}
}
