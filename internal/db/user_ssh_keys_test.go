// White-box tests for user_ssh_keys.go: wrapper structs, mapper methods
// across all three database drivers (SQLite, MySQL, PostgreSQL), and audited
// command struct accessors.
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
	"database/sql"
	"testing"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// sshKeyTestFixture returns a fully populated UserSshKeys and its individual parts.
func sshKeyTestFixture() (UserSshKeys, string, types.NullableUserID, string, string, string, string, types.Timestamp, string) {
	keyID := "ssh-key-001"
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	publicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."
	keyType := "ed25519"
	fingerprint := "SHA256:abcdef1234567890"
	label := "work-laptop"
	dateCreated := types.NewTimestamp(time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC))
	lastUsed := "2025-07-01T14:00:00Z"

	key := UserSshKeys{
		SshKeyID:    keyID,
		UserID:      userID,
		PublicKey:   publicKey,
		KeyType:     keyType,
		Fingerprint: fingerprint,
		Label:       label,
		DateCreated: dateCreated,
		LastUsed:    lastUsed,
	}
	return key, keyID, userID, publicKey, keyType, fingerprint, label, dateCreated, lastUsed
}

// --- SQLite Database.MapUserSshKeys tests ---

func TestDatabase_MapUserSshKeys_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	dateCreated := types.NewTimestamp(time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC))

	input := mdb.UserSshKeys{
		SSHKeyID:    "ssh-sqlite-001",
		UserID:      userID,
		PublicKey:   "ssh-rsa AAAAB3...",
		KeyType:     "rsa",
		Fingerprint: "SHA256:sqlite-fp",
		Label:       sql.NullString{String: "my-key", Valid: true},
		DateCreated: dateCreated,
		LastUsed:    sql.NullString{String: "2025-07-10T12:00:00Z", Valid: true},
	}

	got := d.MapUserSshKeys(input)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"SshKeyID", got.SshKeyID, "ssh-sqlite-001"},
		{"PublicKey", got.PublicKey, "ssh-rsa AAAAB3..."},
		{"KeyType", got.KeyType, "rsa"},
		{"Fingerprint", got.Fingerprint, "SHA256:sqlite-fp"},
		{"Label", got.Label, "my-key"},
		{"LastUsed", got.LastUsed, "2025-07-10T12:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.DateCreated != dateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, dateCreated)
	}
}

func TestDatabase_MapUserSshKeys_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUserSshKeys(mdb.UserSshKeys{})

	if got.SshKeyID != "" {
		t.Errorf("SshKeyID = %q, want empty string", got.SshKeyID)
	}
	if got.UserID.Valid {
		t.Errorf("UserID.Valid = true, want false")
	}
	if got.PublicKey != "" {
		t.Errorf("PublicKey = %q, want empty string", got.PublicKey)
	}
	if got.KeyType != "" {
		t.Errorf("KeyType = %q, want empty string", got.KeyType)
	}
	if got.Fingerprint != "" {
		t.Errorf("Fingerprint = %q, want empty string", got.Fingerprint)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string (null label)", got.Label)
	}
	if got.LastUsed != "" {
		t.Errorf("LastUsed = %q, want empty string (null last_used)", got.LastUsed)
	}
}

func TestDatabase_MapUserSshKeys_NullLabel(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := mdb.UserSshKeys{
		SSHKeyID: "ssh-null-label",
		Label:    sql.NullString{Valid: false},
	}

	got := d.MapUserSshKeys(input)
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string for null label", got.Label)
	}
}

func TestDatabase_MapUserSshKeys_ValidEmptyLabel(t *testing.T) {
	t.Parallel()
	d := Database{}
	// Valid=true but empty string -- should pass through as empty
	input := mdb.UserSshKeys{
		SSHKeyID: "ssh-empty-label",
		Label:    sql.NullString{String: "", Valid: true},
	}

	got := d.MapUserSshKeys(input)
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

func TestDatabase_MapUserSshKeys_NullLastUsed(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := mdb.UserSshKeys{
		SSHKeyID: "ssh-null-lastused",
		LastUsed: sql.NullString{Valid: false},
	}

	got := d.MapUserSshKeys(input)
	if got.LastUsed != "" {
		t.Errorf("LastUsed = %q, want empty string for null last_used", got.LastUsed)
	}
}

func TestDatabase_MapUserSshKeys_ValidLastUsed(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := mdb.UserSshKeys{
		SSHKeyID: "ssh-valid-lastused",
		LastUsed: sql.NullString{String: "2025-12-25T00:00:00Z", Valid: true},
	}

	got := d.MapUserSshKeys(input)
	if got.LastUsed != "2025-12-25T00:00:00Z" {
		t.Errorf("LastUsed = %q, want %q", got.LastUsed, "2025-12-25T00:00:00Z")
	}
}

func TestDatabase_MapUserSshKeys_NullUserID(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := mdb.UserSshKeys{
		SSHKeyID: "ssh-null-userid",
		UserID:   types.NullableUserID{Valid: false},
	}

	got := d.MapUserSshKeys(input)
	if got.UserID.Valid {
		t.Errorf("UserID.Valid = true, want false")
	}
}

// --- MySQL MysqlDatabase.MapUserSshKeys tests ---

func TestMysqlDatabase_MapUserSshKeys_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	dateCreated := types.NewTimestamp(time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC))
	lastUsedTime := time.Date(2025, 7, 10, 12, 0, 0, 0, time.UTC)

	input := mdbm.UserSshKeys{
		SSHKeyID:    "ssh-mysql-001",
		UserID:      userID,
		PublicKey:   "ssh-rsa AAAAB3...",
		KeyType:     "rsa",
		Fingerprint: "SHA256:mysql-fp",
		Label:       sql.NullString{String: "mysql-key", Valid: true},
		DateCreated: dateCreated,
		LastUsed:    sql.NullTime{Time: lastUsedTime, Valid: true},
	}

	got := d.MapUserSshKeys(input)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"SshKeyID", got.SshKeyID, "ssh-mysql-001"},
		{"PublicKey", got.PublicKey, "ssh-rsa AAAAB3..."},
		{"KeyType", got.KeyType, "rsa"},
		{"Fingerprint", got.Fingerprint, "SHA256:mysql-fp"},
		{"Label", got.Label, "mysql-key"},
		{"LastUsed", got.LastUsed, lastUsedTime.Format(time.RFC3339)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.DateCreated != dateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, dateCreated)
	}
}

func TestMysqlDatabase_MapUserSshKeys_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapUserSshKeys(mdbm.UserSshKeys{})

	if got.SshKeyID != "" {
		t.Errorf("SshKeyID = %q, want empty string", got.SshKeyID)
	}
	if got.UserID.Valid {
		t.Errorf("UserID.Valid = true, want false")
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string (null label)", got.Label)
	}
	if got.LastUsed != "" {
		t.Errorf("LastUsed = %q, want empty string (null last_used)", got.LastUsed)
	}
}

func TestMysqlDatabase_MapUserSshKeys_NullLabel(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := mdbm.UserSshKeys{
		SSHKeyID: "ssh-mysql-null-label",
		Label:    sql.NullString{Valid: false},
	}

	got := d.MapUserSshKeys(input)
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string for null label", got.Label)
	}
}

func TestMysqlDatabase_MapUserSshKeys_NullLastUsed(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := mdbm.UserSshKeys{
		SSHKeyID: "ssh-mysql-null-lastused",
		LastUsed: sql.NullTime{Valid: false},
	}

	got := d.MapUserSshKeys(input)
	if got.LastUsed != "" {
		t.Errorf("LastUsed = %q, want empty string for null last_used", got.LastUsed)
	}
}

func TestMysqlDatabase_MapUserSshKeys_LastUsed_RFC3339Format(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	tests := []struct {
		name    string
		time    time.Time
		wantStr string
	}{
		{
			"UTC midnight",
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			"2025-01-01T00:00:00Z",
		},
		{
			"specific timestamp",
			time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			"2025-12-31T23:59:59Z",
		},
		{
			"midday",
			time.Date(2025, 6, 15, 12, 30, 45, 0, time.UTC),
			"2025-06-15T12:30:45Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbm.UserSshKeys{
				SSHKeyID: "time-test",
				LastUsed: sql.NullTime{Time: tt.time, Valid: true},
			}
			got := d.MapUserSshKeys(input)
			if got.LastUsed != tt.wantStr {
				t.Errorf("LastUsed = %q, want %q", got.LastUsed, tt.wantStr)
			}
		})
	}
}

// --- PostgreSQL PsqlDatabase.MapUserSshKeys tests ---

func TestPsqlDatabase_MapUserSshKeys_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	dateCreated := types.NewTimestamp(time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC))
	lastUsedTime := time.Date(2025, 7, 10, 12, 0, 0, 0, time.UTC)

	input := mdbp.UserSshKeys{
		SSHKeyID:    "ssh-psql-001",
		UserID:      userID,
		PublicKey:   "ssh-rsa AAAAB3...",
		KeyType:     "rsa",
		Fingerprint: "SHA256:psql-fp",
		Label:       sql.NullString{String: "psql-key", Valid: true},
		DateCreated: dateCreated,
		LastUsed:    sql.NullTime{Time: lastUsedTime, Valid: true},
	}

	got := d.MapUserSshKeys(input)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"SshKeyID", got.SshKeyID, "ssh-psql-001"},
		{"PublicKey", got.PublicKey, "ssh-rsa AAAAB3..."},
		{"KeyType", got.KeyType, "rsa"},
		{"Fingerprint", got.Fingerprint, "SHA256:psql-fp"},
		{"Label", got.Label, "psql-key"},
		{"LastUsed", got.LastUsed, lastUsedTime.Format(time.RFC3339)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.DateCreated != dateCreated {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, dateCreated)
	}
}

func TestPsqlDatabase_MapUserSshKeys_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapUserSshKeys(mdbp.UserSshKeys{})

	if got.SshKeyID != "" {
		t.Errorf("SshKeyID = %q, want empty string", got.SshKeyID)
	}
	if got.UserID.Valid {
		t.Errorf("UserID.Valid = true, want false")
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string (null label)", got.Label)
	}
	if got.LastUsed != "" {
		t.Errorf("LastUsed = %q, want empty string (null last_used)", got.LastUsed)
	}
}

func TestPsqlDatabase_MapUserSshKeys_NullLabel(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := mdbp.UserSshKeys{
		SSHKeyID: "ssh-psql-null-label",
		Label:    sql.NullString{Valid: false},
	}

	got := d.MapUserSshKeys(input)
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string for null label", got.Label)
	}
}

func TestPsqlDatabase_MapUserSshKeys_NullLastUsed(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := mdbp.UserSshKeys{
		SSHKeyID: "ssh-psql-null-lastused",
		LastUsed: sql.NullTime{Valid: false},
	}

	got := d.MapUserSshKeys(input)
	if got.LastUsed != "" {
		t.Errorf("LastUsed = %q, want empty string for null last_used", got.LastUsed)
	}
}

func TestPsqlDatabase_MapUserSshKeys_LastUsed_RFC3339Format(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	tests := []struct {
		name    string
		time    time.Time
		wantStr string
	}{
		{
			"UTC midnight",
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			"2025-01-01T00:00:00Z",
		},
		{
			"specific timestamp",
			time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			"2025-12-31T23:59:59Z",
		},
		{
			"midday",
			time.Date(2025, 6, 15, 12, 30, 45, 0, time.UTC),
			"2025-06-15T12:30:45Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdbp.UserSshKeys{
				SSHKeyID: "time-test",
				LastUsed: sql.NullTime{Time: tt.time, Valid: true},
			}
			got := d.MapUserSshKeys(input)
			if got.LastUsed != tt.wantStr {
				t.Errorf("LastUsed = %q, want %q", got.LastUsed, tt.wantStr)
			}
		})
	}
}

// --- Cross-database mapper consistency ---

func TestCrossDatabaseMapUserSshKeys_ConsistencyWithLabel(t *testing.T) {
	t.Parallel()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	dateCreated := types.NewTimestamp(time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC))
	lastUsedTime := time.Date(2025, 7, 10, 12, 0, 0, 0, time.UTC)
	lastUsedStr := lastUsedTime.Format(time.RFC3339)

	sqliteInput := mdb.UserSshKeys{
		SSHKeyID:    "cross-ssh-001",
		UserID:      userID,
		PublicKey:   "ssh-ed25519 AAAA...",
		KeyType:     "ed25519",
		Fingerprint: "SHA256:cross-fp",
		Label:       sql.NullString{String: "cross-label", Valid: true},
		DateCreated: dateCreated,
		LastUsed:    sql.NullString{String: lastUsedStr, Valid: true},
	}
	mysqlInput := mdbm.UserSshKeys{
		SSHKeyID:    "cross-ssh-001",
		UserID:      userID,
		PublicKey:   "ssh-ed25519 AAAA...",
		KeyType:     "ed25519",
		Fingerprint: "SHA256:cross-fp",
		Label:       sql.NullString{String: "cross-label", Valid: true},
		DateCreated: dateCreated,
		LastUsed:    sql.NullTime{Time: lastUsedTime, Valid: true},
	}
	psqlInput := mdbp.UserSshKeys{
		SSHKeyID:    "cross-ssh-001",
		UserID:      userID,
		PublicKey:   "ssh-ed25519 AAAA...",
		KeyType:     "ed25519",
		Fingerprint: "SHA256:cross-fp",
		Label:       sql.NullString{String: "cross-label", Valid: true},
		DateCreated: dateCreated,
		LastUsed:    sql.NullTime{Time: lastUsedTime, Valid: true},
	}

	sqliteResult := Database{}.MapUserSshKeys(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapUserSshKeys(mysqlInput)
	psqlResult := PsqlDatabase{}.MapUserSshKeys(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

func TestCrossDatabaseMapUserSshKeys_ConsistencyWithNulls(t *testing.T) {
	t.Parallel()
	nullUser := types.NullableUserID{Valid: false}
	dateCreated := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	sqliteInput := mdb.UserSshKeys{
		SSHKeyID:    "null-cross-001",
		UserID:      nullUser,
		PublicKey:   "ssh-rsa AAAA...",
		KeyType:     "rsa",
		Fingerprint: "SHA256:null-fp",
		Label:       sql.NullString{Valid: false},
		DateCreated: dateCreated,
		LastUsed:    sql.NullString{Valid: false},
	}
	mysqlInput := mdbm.UserSshKeys{
		SSHKeyID:    "null-cross-001",
		UserID:      nullUser,
		PublicKey:   "ssh-rsa AAAA...",
		KeyType:     "rsa",
		Fingerprint: "SHA256:null-fp",
		Label:       sql.NullString{Valid: false},
		DateCreated: dateCreated,
		LastUsed:    sql.NullTime{Valid: false},
	}
	psqlInput := mdbp.UserSshKeys{
		SSHKeyID:    "null-cross-001",
		UserID:      nullUser,
		PublicKey:   "ssh-rsa AAAA...",
		KeyType:     "rsa",
		Fingerprint: "SHA256:null-fp",
		Label:       sql.NullString{Valid: false},
		DateCreated: dateCreated,
		LastUsed:    sql.NullTime{Valid: false},
	}

	sqliteResult := Database{}.MapUserSshKeys(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapUserSshKeys(mysqlInput)
	psqlResult := PsqlDatabase{}.MapUserSshKeys(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL differ with nulls:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL differ with nulls:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewUserSshKeyCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-ssh-1"),
		RequestID: "req-ssh-001",
		IP:        "10.0.0.1",
	}
	params := CreateUserSshKeyParams{
		UserID:      types.NullableUserID{ID: userID, Valid: true},
		PublicKey:   "ssh-ed25519 AAAA...",
		KeyType:     "ed25519",
		Fingerprint: "SHA256:test-fp",
		Label:       "test-label",
		DateCreated: types.NewTimestamp(time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)),
	}

	cmd := Database{}.NewUserSshKeyCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_ssh_keys" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_ssh_keys")
	}
	p, ok := cmd.Params().(CreateUserSshKeyParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateUserSshKeyParams", cmd.Params())
	}
	if p.PublicKey != "ssh-ed25519 AAAA..." {
		t.Errorf("Params().PublicKey = %q, want %q", p.PublicKey, "ssh-ed25519 AAAA...")
	}
	if p.KeyType != "ed25519" {
		t.Errorf("Params().KeyType = %q, want %q", p.KeyType, "ed25519")
	}
	if p.Fingerprint != "SHA256:test-fp" {
		t.Errorf("Params().Fingerprint = %q, want %q", p.Fingerprint, "SHA256:test-fp")
	}
	if p.Label != "test-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "test-label")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewUserSshKeyCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	cmd := NewUserSshKeyCmd{}

	row := mdb.UserSshKeys{SSHKeyID: "ssh-row-001"}
	got := cmd.GetID(row)
	if got != "ssh-row-001" {
		t.Errorf("GetID() = %q, want %q", got, "ssh-row-001")
	}
}

func TestNewUserSshKeyCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewUserSshKeyCmd{}

	row := mdb.UserSshKeys{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestDeleteUserSshKeyCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := "ssh-del-001"

	cmd := Database{}.DeleteUserSshKeyCmd(ctx, ac, id)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_ssh_keys" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_ssh_keys")
	}
	if cmd.GetID() != "ssh-del-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "ssh-del-001")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestUpdateUserSshKeyLabelCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := "ssh-upd-label-001"
	label := "new-label"

	cmd := Database{}.UpdateUserSshKeyLabelCmd(ctx, ac, id, label)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_ssh_keys" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_ssh_keys")
	}
	if cmd.GetID() != "ssh-upd-label-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "ssh-upd-label-001")
	}
	// Params returns a map[string]any for this command
	paramsRaw := cmd.Params()
	paramsMap, ok := paramsRaw.(map[string]any)
	if !ok {
		t.Fatalf("Params() returned %T, want map[string]any", paramsRaw)
	}
	if paramsMap["id"] != "ssh-upd-label-001" {
		t.Errorf("Params()[\"id\"] = %v, want %q", paramsMap["id"], "ssh-upd-label-001")
	}
	if paramsMap["label"] != "new-label" {
		t.Errorf("Params()[\"label\"] = %v, want %q", paramsMap["label"], "new-label")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestUpdateUserSshKeyLabelCmd_EmptyLabel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	cmd := Database{}.UpdateUserSshKeyLabelCmd(ctx, ac, "ssh-id", "")

	paramsMap, ok := cmd.Params().(map[string]any)
	if !ok {
		t.Fatalf("Params() returned %T, want map[string]any", cmd.Params())
	}
	if paramsMap["label"] != "" {
		t.Errorf("Params()[\"label\"] = %v, want empty string", paramsMap["label"])
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewUserSshKeyCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-ssh-node"),
		RequestID: "mysql-ssh-001",
		IP:        "192.168.1.1",
	}
	params := CreateUserSshKeyParams{
		PublicKey:   "ssh-rsa AAAA...",
		KeyType:     "rsa",
		Fingerprint: "SHA256:mysql-test-fp",
		Label:       "mysql-label",
	}

	cmd := MysqlDatabase{}.NewUserSshKeyCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_ssh_keys" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_ssh_keys")
	}
	p, ok := cmd.Params().(CreateUserSshKeyParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateUserSshKeyParams", cmd.Params())
	}
	if p.Label != "mysql-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "mysql-label")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewUserSshKeyCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewUserSshKeyCmdMysql{}

	row := mdbm.UserSshKeys{SSHKeyID: "mysql-ssh-row-001"}
	got := cmd.GetID(row)
	if got != "mysql-ssh-row-001" {
		t.Errorf("GetID() = %q, want %q", got, "mysql-ssh-row-001")
	}
}

func TestDeleteUserSshKeyCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := "ssh-mysql-del-001"

	cmd := MysqlDatabase{}.DeleteUserSshKeyCmd(ctx, ac, id)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_ssh_keys" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_ssh_keys")
	}
	if cmd.GetID() != "ssh-mysql-del-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "ssh-mysql-del-001")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestUpdateUserSshKeyLabelCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := "ssh-mysql-upd-label-001"
	label := "mysql-new-label"

	cmd := MysqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, id, label)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_ssh_keys" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_ssh_keys")
	}
	if cmd.GetID() != "ssh-mysql-upd-label-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "ssh-mysql-upd-label-001")
	}
	paramsMap, ok := cmd.Params().(map[string]any)
	if !ok {
		t.Fatalf("Params() returned %T, want map[string]any", cmd.Params())
	}
	if paramsMap["id"] != "ssh-mysql-upd-label-001" {
		t.Errorf("Params()[\"id\"] = %v, want %q", paramsMap["id"], "ssh-mysql-upd-label-001")
	}
	if paramsMap["label"] != "mysql-new-label" {
		t.Errorf("Params()[\"label\"] = %v, want %q", paramsMap["label"], "mysql-new-label")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- PostgreSQL Audited Command Accessor tests ---

func TestNewUserSshKeyCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-ssh-node"),
		RequestID: "psql-ssh-001",
		IP:        "172.16.0.1",
	}
	params := CreateUserSshKeyParams{
		PublicKey:   "ssh-ed25519 AAAA...",
		KeyType:     "ed25519",
		Fingerprint: "SHA256:psql-test-fp",
		Label:       "psql-label",
	}

	cmd := PsqlDatabase{}.NewUserSshKeyCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_ssh_keys" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_ssh_keys")
	}
	p, ok := cmd.Params().(CreateUserSshKeyParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateUserSshKeyParams", cmd.Params())
	}
	if p.Label != "psql-label" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "psql-label")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewUserSshKeyCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewUserSshKeyCmdPsql{}

	row := mdbp.UserSshKeys{SSHKeyID: "psql-ssh-row-001"}
	got := cmd.GetID(row)
	if got != "psql-ssh-row-001" {
		t.Errorf("GetID() = %q, want %q", got, "psql-ssh-row-001")
	}
}

func TestDeleteUserSshKeyCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := "ssh-psql-del-001"

	cmd := PsqlDatabase{}.DeleteUserSshKeyCmd(ctx, ac, id)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_ssh_keys" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_ssh_keys")
	}
	if cmd.GetID() != "ssh-psql-del-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "ssh-psql-del-001")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestUpdateUserSshKeyLabelCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := "ssh-psql-upd-label-001"
	label := "psql-new-label"

	cmd := PsqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, id, label)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_ssh_keys" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_ssh_keys")
	}
	if cmd.GetID() != "ssh-psql-upd-label-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "ssh-psql-upd-label-001")
	}
	paramsMap, ok := cmd.Params().(map[string]any)
	if !ok {
		t.Fatalf("Params() returned %T, want map[string]any", cmd.Params())
	}
	if paramsMap["id"] != "ssh-psql-upd-label-001" {
		t.Errorf("Params()[\"id\"] = %v, want %q", paramsMap["id"], "ssh-psql-upd-label-001")
	}
	if paramsMap["label"] != "psql-new-label" {
		t.Errorf("Params()[\"label\"] = %v, want %q", paramsMap["label"], "psql-new-label")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- Cross-database audited command consistency ---

func TestAuditedUserSshKeyCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateUserSshKeyParams{}
	id := "ssh-consistency-del"

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewUserSshKeyCmd(ctx, ac, createParams).TableName()},
		{"SQLite UpdateLabel", Database{}.UpdateUserSshKeyLabelCmd(ctx, ac, id, "lbl").TableName()},
		{"SQLite Delete", Database{}.DeleteUserSshKeyCmd(ctx, ac, id).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewUserSshKeyCmd(ctx, ac, createParams).TableName()},
		{"MySQL UpdateLabel", MysqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, id, "lbl").TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteUserSshKeyCmd(ctx, ac, id).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewUserSshKeyCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL UpdateLabel", PsqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, id, "lbl").TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteUserSshKeyCmd(ctx, ac, id).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "user_ssh_keys" {
				t.Errorf("TableName() = %q, want %q", c.name, "user_ssh_keys")
			}
		})
	}
}

func TestAuditedUserSshKeyCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateUserSshKeyParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite Create", Database{}.NewUserSshKeyCmd(ctx, ac, createParams).Recorder()},
		{"SQLite UpdateLabel", Database{}.UpdateUserSshKeyLabelCmd(ctx, ac, "id", "lbl").Recorder()},
		{"SQLite Delete", Database{}.DeleteUserSshKeyCmd(ctx, ac, "id").Recorder()},
		{"MySQL Create", MysqlDatabase{}.NewUserSshKeyCmd(ctx, ac, createParams).Recorder()},
		{"MySQL UpdateLabel", MysqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, "id", "lbl").Recorder()},
		{"MySQL Delete", MysqlDatabase{}.DeleteUserSshKeyCmd(ctx, ac, "id").Recorder()},
		{"PostgreSQL Create", PsqlDatabase{}.NewUserSshKeyCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL UpdateLabel", PsqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, "id", "lbl").Recorder()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteUserSshKeyCmd(ctx, ac, "id").Recorder()},
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

func TestAuditedUserSshKeyCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	id := "ssh-getid-cross-001"

	t.Run("UpdateLabelCmd GetID returns key ID", func(t *testing.T) {
		t.Parallel()

		sqliteCmd := Database{}.UpdateUserSshKeyLabelCmd(ctx, ac, id, "lbl")
		mysqlCmd := MysqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, id, "lbl")
		psqlCmd := PsqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, id, "lbl")

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

	t.Run("DeleteCmd GetID returns key ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteUserSshKeyCmd(ctx, ac, id)
		mysqlCmd := MysqlDatabase{}.DeleteUserSshKeyCmd(ctx, ac, id)
		psqlCmd := PsqlDatabase{}.DeleteUserSshKeyCmd(ctx, ac, id)

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
		wantID := "ssh-create-row-001"

		sqliteCmd := NewUserSshKeyCmd{}
		mysqlCmd := NewUserSshKeyCmdMysql{}
		psqlCmd := NewUserSshKeyCmdPsql{}

		sqliteRow := mdb.UserSshKeys{SSHKeyID: wantID}
		mysqlRow := mdbm.UserSshKeys{SSHKeyID: wantID}
		psqlRow := mdbp.UserSshKeys{SSHKeyID: wantID}

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

func TestUpdateUserSshKeyLabelCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()

	sqliteCmd := Database{}.UpdateUserSshKeyLabelCmd(context.Background(), audited.AuditContext{}, "", "lbl")
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateUserSshKeyLabelCmd(context.Background(), audited.AuditContext{}, "", "lbl")
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateUserSshKeyLabelCmd(context.Background(), audited.AuditContext{}, "", "lbl")
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteUserSshKeyCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := ""

	sqliteCmd := Database{}.DeleteUserSshKeyCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteUserSshKeyCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteUserSshKeyCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- UserSshKeys wrapper struct tests ---

func TestUserSshKeys_ZeroValue(t *testing.T) {
	t.Parallel()
	var key UserSshKeys

	if key.SshKeyID != "" {
		t.Errorf("SshKeyID = %q, want empty string", key.SshKeyID)
	}
	if key.UserID.Valid {
		t.Errorf("UserID.Valid = true, want false")
	}
	if key.PublicKey != "" {
		t.Errorf("PublicKey = %q, want empty string", key.PublicKey)
	}
	if key.KeyType != "" {
		t.Errorf("KeyType = %q, want empty string", key.KeyType)
	}
	if key.Fingerprint != "" {
		t.Errorf("Fingerprint = %q, want empty string", key.Fingerprint)
	}
	if key.Label != "" {
		t.Errorf("Label = %q, want empty string", key.Label)
	}
	if key.LastUsed != "" {
		t.Errorf("LastUsed = %q, want empty string", key.LastUsed)
	}
}

func TestCreateUserSshKeyParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var params CreateUserSshKeyParams

	if params.UserID.Valid {
		t.Errorf("UserID.Valid = true, want false")
	}
	if params.PublicKey != "" {
		t.Errorf("PublicKey = %q, want empty string", params.PublicKey)
	}
	if params.KeyType != "" {
		t.Errorf("KeyType = %q, want empty string", params.KeyType)
	}
	if params.Fingerprint != "" {
		t.Errorf("Fingerprint = %q, want empty string", params.Fingerprint)
	}
	if params.Label != "" {
		t.Errorf("Label = %q, want empty string", params.Label)
	}
}

// --- UpdateLabel Params map key consistency ---

func TestUpdateUserSshKeyLabelCmd_Params_MapKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	tests := []struct {
		name  string
		id    string
		label string
	}{
		{"normal values", "ssh-key-123", "my-laptop"},
		{"empty label", "ssh-key-456", ""},
		{"empty id", "", "some-label"},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Test all three drivers produce consistent map structure
			sqliteCmd := Database{}.UpdateUserSshKeyLabelCmd(ctx, ac, tt.id, tt.label)
			mysqlCmd := MysqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, tt.id, tt.label)
			psqlCmd := PsqlDatabase{}.UpdateUserSshKeyLabelCmd(ctx, ac, tt.id, tt.label)

			for _, pair := range []struct {
				driver string
				params any
			}{
				{"SQLite", sqliteCmd.Params()},
				{"MySQL", mysqlCmd.Params()},
				{"PostgreSQL", psqlCmd.Params()},
			} {
				m, ok := pair.params.(map[string]any)
				if !ok {
					t.Fatalf("%s Params() returned %T, want map[string]any", pair.driver, pair.params)
				}
				if m["id"] != tt.id {
					t.Errorf("%s Params()[\"id\"] = %v, want %q", pair.driver, m["id"], tt.id)
				}
				if m["label"] != tt.label {
					t.Errorf("%s Params()[\"label\"] = %v, want %q", pair.driver, m["label"], tt.label)
				}
			}
		})
	}
}

// --- Fixture helper test ---

func TestSshKeyTestFixture_ReturnsPopulatedValues(t *testing.T) {
	t.Parallel()
	key, keyID, userID, publicKey, keyType, fingerprint, label, dateCreated, lastUsed := sshKeyTestFixture()

	if keyID == "" {
		t.Fatal("keyID is empty")
	}
	if !userID.Valid {
		t.Fatal("userID.Valid = false, want true")
	}
	if publicKey == "" {
		t.Fatal("publicKey is empty")
	}
	if keyType == "" {
		t.Fatal("keyType is empty")
	}
	if fingerprint == "" {
		t.Fatal("fingerprint is empty")
	}
	if label == "" {
		t.Fatal("label is empty")
	}
	if !dateCreated.Valid {
		t.Fatal("dateCreated.Valid = false, want true")
	}
	if lastUsed == "" {
		t.Fatal("lastUsed is empty")
	}

	if key.SshKeyID != keyID {
		t.Errorf("key.SshKeyID = %q, want %q", key.SshKeyID, keyID)
	}
	if key.UserID != userID {
		t.Errorf("key.UserID = %v, want %v", key.UserID, userID)
	}
	if key.PublicKey != publicKey {
		t.Errorf("key.PublicKey = %q, want %q", key.PublicKey, publicKey)
	}
	if key.KeyType != keyType {
		t.Errorf("key.KeyType = %q, want %q", key.KeyType, keyType)
	}
	if key.Fingerprint != fingerprint {
		t.Errorf("key.Fingerprint = %q, want %q", key.Fingerprint, fingerprint)
	}
	if key.Label != label {
		t.Errorf("key.Label = %q, want %q", key.Label, label)
	}
	if key.DateCreated != dateCreated {
		t.Errorf("key.DateCreated = %v, want %v", key.DateCreated, dateCreated)
	}
	if key.LastUsed != lastUsed {
		t.Errorf("key.LastUsed = %q, want %q", key.LastUsed, lastUsed)
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.UserSshKeys]  = NewUserSshKeyCmd{}
	_ audited.DeleteCommand[mdb.UserSshKeys]  = DeleteUserSshKeyCmd{}
	_ audited.UpdateCommand[mdb.UserSshKeys]  = UpdateUserSshKeyLabelCmd{}
	_ audited.CreateCommand[mdbm.UserSshKeys] = NewUserSshKeyCmdMysql{}
	_ audited.DeleteCommand[mdbm.UserSshKeys] = DeleteUserSshKeyCmdMysql{}
	_ audited.UpdateCommand[mdbm.UserSshKeys] = UpdateUserSshKeyLabelCmdMysql{}
	_ audited.CreateCommand[mdbp.UserSshKeys] = NewUserSshKeyCmdPsql{}
	_ audited.DeleteCommand[mdbp.UserSshKeys] = DeleteUserSshKeyCmdPsql{}
	_ audited.UpdateCommand[mdbp.UserSshKeys] = UpdateUserSshKeyLabelCmdPsql{}
)
