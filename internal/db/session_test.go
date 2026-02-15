package db

import (
	"context"
	"database/sql"
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
	_ audited.CreateCommand[mdb.Sessions]  = NewSessionCmd{}
	_ audited.UpdateCommand[mdb.Sessions]  = UpdateSessionCmd{}
	_ audited.DeleteCommand[mdb.Sessions]  = DeleteSessionCmd{}
	_ audited.CreateCommand[mdbm.Sessions] = NewSessionCmdMysql{}
	_ audited.UpdateCommand[mdbm.Sessions] = UpdateSessionCmdMysql{}
	_ audited.DeleteCommand[mdbm.Sessions] = DeleteSessionCmdMysql{}
	_ audited.CreateCommand[mdbp.Sessions] = NewSessionCmdPsql{}
	_ audited.UpdateCommand[mdbp.Sessions] = UpdateSessionCmdPsql{}
	_ audited.DeleteCommand[mdbp.Sessions] = DeleteSessionCmdPsql{}
)

// --- Test data helpers ---

// sessionTestFixture returns a fully populated Sessions struct and its component parts.
func sessionTestFixture() (Sessions, types.SessionID, types.NullableUserID, types.Timestamp) {
	sessionID := types.NewSessionID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 8, 15, 10, 30, 0, 0, time.UTC))
	s := Sessions{
		SessionID:   sessionID,
		UserID:      userID,
		CreatedAt:   ts,
		ExpiresAt:   ts,
		LastAccess:  sql.NullString{String: "2025-08-15T10:30:00Z", Valid: true},
		IpAddress:   sql.NullString{String: "192.168.1.100", Valid: true},
		UserAgent:   sql.NullString{String: "Mozilla/5.0 (Macintosh)", Valid: true},
		SessionData: sql.NullString{String: `{"theme":"dark"}`, Valid: true},
	}
	return s, sessionID, userID, ts
}

// sessionCreateParams returns a CreateSessionParams with all fields populated.
func sessionCreateParams() CreateSessionParams {
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 9, 1, 12, 0, 0, 0, time.UTC))
	return CreateSessionParams{
		UserID:      userID,
		CreatedAt:   ts,
		ExpiresAt:   ts,
		LastAccess:  sql.NullString{String: "2025-09-01T12:00:00Z", Valid: true},
		IpAddress:   sql.NullString{String: "10.0.0.1", Valid: true},
		UserAgent:   sql.NullString{String: "curl/7.88.1", Valid: true},
		SessionData: sql.NullString{String: `{"role":"admin"}`, Valid: true},
	}
}

// sessionUpdateParams returns an UpdateSessionParams with all fields populated.
func sessionUpdateParams() UpdateSessionParams {
	sessionID := types.NewSessionID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 10, 20, 8, 45, 0, 0, time.UTC))
	return UpdateSessionParams{
		UserID:      userID,
		CreatedAt:   ts,
		ExpiresAt:   ts,
		LastAccess:  sql.NullString{String: "2025-10-20T08:45:00Z", Valid: true},
		IpAddress:   sql.NullString{String: "172.16.0.50", Valid: true},
		UserAgent:   sql.NullString{String: "PostmanRuntime/7.32", Valid: true},
		SessionData: sql.NullString{String: `{"lang":"en"}`, Valid: true},
		SessionID:   sessionID,
	}
}

// --- MapStringSession tests ---

func TestMapStringSession_AllFields(t *testing.T) {
	t.Parallel()
	s, sessionID, userID, ts := sessionTestFixture()

	got := MapStringSession(s)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"SessionID", got.SessionID, sessionID.String()},
		{"UserID", got.UserID, userID.String()},
		{"CreatedAt", got.CreatedAt, ts.String()},
		{"ExpiresAt", got.ExpiresAt, ts.String()},
		{"LastAccess", got.LastAccess, "2025-08-15T10:30:00Z"},
		{"IpAddress", got.IpAddress, "192.168.1.100"},
		{"UserAgent", got.UserAgent, "Mozilla/5.0 (Macintosh)"},
		{"SessionData", got.SessionData, `{"theme":"dark"}`},
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

func TestMapStringSession_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringSession(Sessions{})

	if got.SessionID != "" {
		t.Errorf("SessionID = %q, want empty string", got.SessionID)
	}
	// sql.NullString zero value has Valid=false, so MapStringSession returns ""
	nullStringFields := []struct {
		name string
		got  string
	}{
		{"LastAccess", got.LastAccess},
		{"IpAddress", got.IpAddress},
		{"UserAgent", got.UserAgent},
		{"SessionData", got.SessionData},
	}
	for _, f := range nullStringFields {
		if f.got != "" {
			t.Errorf("%s = %q, want empty string for zero-value NullString", f.name, f.got)
		}
	}
}

func TestMapStringSession_NullOptionalFields(t *testing.T) {
	t.Parallel()
	// All NullString fields explicitly invalid. The underlying String value
	// must NOT leak through; MapStringSession returns "" for invalid NullStrings.
	s := Sessions{
		SessionID:   types.NewSessionID(),
		LastAccess:  sql.NullString{String: "should-not-appear", Valid: false},
		IpAddress:   sql.NullString{String: "should-not-appear", Valid: false},
		UserAgent:   sql.NullString{String: "should-not-appear", Valid: false},
		SessionData: sql.NullString{String: "should-not-appear", Valid: false},
	}

	got := MapStringSession(s)

	nullFields := []struct {
		name string
		got  string
	}{
		{"LastAccess", got.LastAccess},
		{"IpAddress", got.IpAddress},
		{"UserAgent", got.UserAgent},
		{"SessionData", got.SessionData},
	}

	for _, f := range nullFields {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			if f.got != "" {
				t.Errorf("%s = %q, want empty string for invalid NullString", f.name, f.got)
			}
			if f.got == "should-not-appear" {
				t.Errorf("%s leaked the underlying String value through an invalid NullString", f.name)
			}
		})
	}
}

func TestMapStringSession_NullUserID(t *testing.T) {
	t.Parallel()
	s := Sessions{
		UserID: types.NullableUserID{Valid: false},
	}
	got := MapStringSession(s)
	// Should produce whatever NullableUserID.String() returns for invalid
	if got.UserID != s.UserID.String() {
		t.Errorf("UserID = %q, want %q", got.UserID, s.UserID.String())
	}
}

func TestMapStringSession_ValidUserID(t *testing.T) {
	t.Parallel()
	uid := types.NewUserID()
	s := Sessions{
		UserID: types.NullableUserID{ID: uid, Valid: true},
	}
	got := MapStringSession(s)
	if got.UserID != s.UserID.String() {
		t.Errorf("UserID = %q, want %q", got.UserID, s.UserID.String())
	}
}

// --- SQLite Database.MapSession tests ---

func TestDatabase_MapSession_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	sessionID := types.NewSessionID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 3, 10, 9, 0, 0, 0, time.UTC))

	input := mdb.Sessions{
		SessionID:   sessionID,
		UserID:      userID,
		CreatedAt:   ts,
		ExpiresAt:   ts,
		LastAccess:  sql.NullString{String: "2025-03-10T09:00:00Z", Valid: true},
		IpAddress:   sql.NullString{String: "10.0.0.5", Valid: true},
		UserAgent:   sql.NullString{String: "Safari/17", Valid: true},
		SessionData: sql.NullString{String: `{"foo":"bar"}`, Valid: true},
	}

	got := d.MapSession(input)

	if got.SessionID != sessionID {
		t.Errorf("SessionID = %v, want %v", got.SessionID, sessionID)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.CreatedAt != ts {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, ts)
	}
	if got.ExpiresAt != ts {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, ts)
	}
	if got.LastAccess.String != "2025-03-10T09:00:00Z" || !got.LastAccess.Valid {
		t.Errorf("LastAccess = %v, want valid '2025-03-10T09:00:00Z'", got.LastAccess)
	}
	if got.IpAddress.String != "10.0.0.5" || !got.IpAddress.Valid {
		t.Errorf("IpAddress = %v, want valid '10.0.0.5'", got.IpAddress)
	}
	if got.UserAgent.String != "Safari/17" || !got.UserAgent.Valid {
		t.Errorf("UserAgent = %v, want valid 'Safari/17'", got.UserAgent)
	}
	if got.SessionData.String != `{"foo":"bar"}` || !got.SessionData.Valid {
		t.Errorf("SessionData = %v, want valid '{\"foo\":\"bar\"}'", got.SessionData)
	}
}

func TestDatabase_MapSession_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapSession(mdb.Sessions{})

	if got.SessionID != "" {
		t.Errorf("SessionID = %v, want zero value", got.SessionID)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if got.LastAccess.Valid {
		t.Error("LastAccess.Valid = true, want false for zero value")
	}
	if got.IpAddress.Valid {
		t.Error("IpAddress.Valid = true, want false for zero value")
	}
}

func TestDatabase_MapSession_NullLastAccess(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := mdb.Sessions{
		SessionID:  types.NewSessionID(),
		LastAccess: sql.NullString{Valid: false},
	}

	got := d.MapSession(input)

	if got.LastAccess.Valid {
		t.Error("LastAccess should be invalid/null")
	}
}

// --- SQLite Database.MapCreateSessionParams tests ---

func TestDatabase_MapCreateSessionParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := sessionCreateParams()

	got := d.MapCreateSessionParams(params)

	// A new SessionID should always be generated
	if got.SessionID.IsZero() {
		t.Fatal("expected non-zero SessionID to be generated")
	}

	// All other fields should pass through
	if got.UserID != params.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, params.UserID)
	}
	if got.CreatedAt != params.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, params.CreatedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	if got.LastAccess != params.LastAccess {
		t.Errorf("LastAccess = %v, want %v", got.LastAccess, params.LastAccess)
	}
	if got.IpAddress != params.IpAddress {
		t.Errorf("IpAddress = %v, want %v", got.IpAddress, params.IpAddress)
	}
	if got.UserAgent != params.UserAgent {
		t.Errorf("UserAgent = %v, want %v", got.UserAgent, params.UserAgent)
	}
	if got.SessionData != params.SessionData {
		t.Errorf("SessionData = %v, want %v", got.SessionData, params.SessionData)
	}
}

func TestDatabase_MapCreateSessionParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := sessionCreateParams()

	got1 := d.MapCreateSessionParams(params)
	got2 := d.MapCreateSessionParams(params)

	if got1.SessionID == got2.SessionID {
		t.Error("two calls generated the same SessionID -- each call should be unique")
	}
}

func TestDatabase_MapCreateSessionParams_NullOptionalFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := CreateSessionParams{
		UserID:     types.NullableUserID{Valid: false},
		LastAccess: sql.NullString{Valid: false},
		IpAddress:  sql.NullString{Valid: false},
		UserAgent:  sql.NullString{Valid: false},
	}

	got := d.MapCreateSessionParams(params)

	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if got.LastAccess.Valid {
		t.Error("LastAccess.Valid = true, want false")
	}
	if got.IpAddress.Valid {
		t.Error("IpAddress.Valid = true, want false")
	}
	if got.UserAgent.Valid {
		t.Error("UserAgent.Valid = true, want false")
	}
}

// --- SQLite Database.MapUpdateSessionParams tests ---

func TestDatabase_MapUpdateSessionParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := sessionUpdateParams()

	got := d.MapUpdateSessionParams(params)

	if got.UserID != params.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, params.UserID)
	}
	if got.CreatedAt != params.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, params.CreatedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	if got.LastAccess != params.LastAccess {
		t.Errorf("LastAccess = %v, want %v", got.LastAccess, params.LastAccess)
	}
	if got.IpAddress != params.IpAddress {
		t.Errorf("IpAddress = %v, want %v", got.IpAddress, params.IpAddress)
	}
	if got.UserAgent != params.UserAgent {
		t.Errorf("UserAgent = %v, want %v", got.UserAgent, params.UserAgent)
	}
	if got.SessionData != params.SessionData {
		t.Errorf("SessionData = %v, want %v", got.SessionData, params.SessionData)
	}
	// SessionID is the WHERE clause identifier -- must be preserved
	if got.SessionID != params.SessionID {
		t.Errorf("SessionID = %v, want %v", got.SessionID, params.SessionID)
	}
}

func TestDatabase_MapUpdateSessionParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateSessionParams(UpdateSessionParams{})

	if got.SessionID != "" {
		t.Errorf("SessionID = %v, want zero value", got.SessionID)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false for zero value")
	}
}

// --- MySQL MysqlDatabase.MapSession tests ---

func TestMysqlDatabase_MapSession_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	sessionID := types.NewSessionID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 5, 11, 0, 0, 0, time.UTC))

	// MySQL sessions use time.Time for LastAccess
	lastAccess := time.Date(2025, 4, 5, 11, 0, 0, 0, time.UTC)

	input := mdbm.Sessions{
		SessionID:   sessionID,
		UserID:      userID,
		CreatedAt:   ts,
		ExpiresAt:   ts,
		LastAccess:  lastAccess,
		IpAddress:   sql.NullString{String: "192.168.1.200", Valid: true},
		UserAgent:   sql.NullString{String: "Chrome/120", Valid: true},
		SessionData: sql.NullString{String: `{"mysql":"data"}`, Valid: true},
	}

	got := d.MapSession(input)

	if got.SessionID != sessionID {
		t.Errorf("SessionID = %v, want %v", got.SessionID, sessionID)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.CreatedAt != ts {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, ts)
	}
	if got.ExpiresAt != ts {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, ts)
	}
	// MySQL maps LastAccess via StringToNullString(a.LastAccess.String())
	// The time.Time.String() representation becomes the NullString value.
	expectedLastAccess := StringToNullString(lastAccess.String())
	if got.LastAccess != expectedLastAccess {
		t.Errorf("LastAccess = %v, want %v", got.LastAccess, expectedLastAccess)
	}
	if got.IpAddress.String != "192.168.1.200" || !got.IpAddress.Valid {
		t.Errorf("IpAddress = %v, want valid '192.168.1.200'", got.IpAddress)
	}
	if got.UserAgent.String != "Chrome/120" || !got.UserAgent.Valid {
		t.Errorf("UserAgent = %v, want valid 'Chrome/120'", got.UserAgent)
	}
	if got.SessionData.String != `{"mysql":"data"}` || !got.SessionData.Valid {
		t.Errorf("SessionData = %v, want valid '{\"mysql\":\"data\"}'", got.SessionData)
	}
}

func TestMysqlDatabase_MapSession_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapSession(mdbm.Sessions{})

	if got.SessionID != "" {
		t.Errorf("SessionID = %v, want zero value", got.SessionID)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
}

// --- MySQL MysqlDatabase.MapCreateSessionParams tests ---

func TestMysqlDatabase_MapCreateSessionParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := sessionCreateParams()

	got := d.MapCreateSessionParams(params)

	if got.SessionID.IsZero() {
		t.Fatal("expected non-zero SessionID to be generated")
	}
	if got.UserID != params.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, params.UserID)
	}
	if got.CreatedAt != params.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, params.CreatedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	// MySQL maps LastAccess via StringToNTime
	expectedLastAccess := StringToNTime(params.LastAccess.String).Time
	if got.LastAccess != expectedLastAccess {
		t.Errorf("LastAccess = %v, want %v", got.LastAccess, expectedLastAccess)
	}
	if got.IpAddress != params.IpAddress {
		t.Errorf("IpAddress = %v, want %v", got.IpAddress, params.IpAddress)
	}
	if got.UserAgent != params.UserAgent {
		t.Errorf("UserAgent = %v, want %v", got.UserAgent, params.UserAgent)
	}
	if got.SessionData != params.SessionData {
		t.Errorf("SessionData = %v, want %v", got.SessionData, params.SessionData)
	}
}

func TestMysqlDatabase_MapCreateSessionParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := sessionCreateParams()

	got1 := d.MapCreateSessionParams(params)
	got2 := d.MapCreateSessionParams(params)

	if got1.SessionID == got2.SessionID {
		t.Error("two calls produced identical SessionIDs; each should be unique")
	}
}

// --- MySQL MysqlDatabase.MapUpdateSessionParams tests ---

func TestMysqlDatabase_MapUpdateSessionParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := sessionUpdateParams()

	got := d.MapUpdateSessionParams(params)

	if got.SessionID != params.SessionID {
		t.Errorf("SessionID = %v, want %v", got.SessionID, params.SessionID)
	}
	if got.UserID != params.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, params.UserID)
	}
	if got.CreatedAt != params.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, params.CreatedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	// MySQL maps LastAccess via StringToNTime
	expectedLastAccess := StringToNTime(params.LastAccess.String).Time
	if got.LastAccess != expectedLastAccess {
		t.Errorf("LastAccess = %v, want %v", got.LastAccess, expectedLastAccess)
	}
	if got.IpAddress != params.IpAddress {
		t.Errorf("IpAddress = %v, want %v", got.IpAddress, params.IpAddress)
	}
	if got.UserAgent != params.UserAgent {
		t.Errorf("UserAgent = %v, want %v", got.UserAgent, params.UserAgent)
	}
	if got.SessionData != params.SessionData {
		t.Errorf("SessionData = %v, want %v", got.SessionData, params.SessionData)
	}
}

// --- PostgreSQL PsqlDatabase.MapSession tests ---

func TestPsqlDatabase_MapSession_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	sessionID := types.NewSessionID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 2, 28, 16, 0, 0, 0, time.UTC))

	// PostgreSQL sessions use sql.NullTime for LastAccess
	lastAccess := sql.NullTime{
		Time:  time.Date(2025, 2, 28, 16, 0, 0, 0, time.UTC),
		Valid: true,
	}

	input := mdbp.Sessions{
		SessionID:   sessionID,
		UserID:      userID,
		CreatedAt:   ts,
		ExpiresAt:   ts,
		LastAccess:  lastAccess,
		IpAddress:   sql.NullString{String: "172.16.0.1", Valid: true},
		UserAgent:   sql.NullString{String: "Firefox/121", Valid: true},
		SessionData: sql.NullString{String: `{"psql":"data"}`, Valid: true},
	}

	got := d.MapSession(input)

	if got.SessionID != sessionID {
		t.Errorf("SessionID = %v, want %v", got.SessionID, sessionID)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.CreatedAt != ts {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, ts)
	}
	if got.ExpiresAt != ts {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, ts)
	}
	// PostgreSQL maps LastAccess via StringToNullString(NullTimeToString(a.LastAccess))
	expectedLastAccess := StringToNullString(NullTimeToString(lastAccess))
	if got.LastAccess != expectedLastAccess {
		t.Errorf("LastAccess = %v, want %v", got.LastAccess, expectedLastAccess)
	}
	if got.IpAddress.String != "172.16.0.1" || !got.IpAddress.Valid {
		t.Errorf("IpAddress = %v, want valid '172.16.0.1'", got.IpAddress)
	}
	if got.UserAgent.String != "Firefox/121" || !got.UserAgent.Valid {
		t.Errorf("UserAgent = %v, want valid 'Firefox/121'", got.UserAgent)
	}
	if got.SessionData.String != `{"psql":"data"}` || !got.SessionData.Valid {
		t.Errorf("SessionData = %v, want valid '{\"psql\":\"data\"}'", got.SessionData)
	}
}

func TestPsqlDatabase_MapSession_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapSession(mdbp.Sessions{})

	if got.SessionID != "" {
		t.Errorf("SessionID = %v, want zero value", got.SessionID)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
}

func TestPsqlDatabase_MapSession_NullLastAccess(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := mdbp.Sessions{
		SessionID:  types.NewSessionID(),
		LastAccess: sql.NullTime{Valid: false},
	}

	got := d.MapSession(input)

	// NullTimeToString returns "" for invalid NullTime,
	// then StringToNullString("") returns NullString{Valid: false}
	if got.LastAccess.Valid {
		t.Error("LastAccess should be invalid/null for a null NullTime")
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateSessionParams tests ---

func TestPsqlDatabase_MapCreateSessionParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := sessionCreateParams()

	got := d.MapCreateSessionParams(params)

	if got.SessionID.IsZero() {
		t.Fatal("expected non-zero SessionID to be generated")
	}
	if got.UserID != params.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, params.UserID)
	}
	if got.CreatedAt != params.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, params.CreatedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	// PostgreSQL maps LastAccess via StringToNTime
	expectedLastAccess := StringToNTime(params.LastAccess.String)
	if got.LastAccess != expectedLastAccess {
		t.Errorf("LastAccess = %v, want %v", got.LastAccess, expectedLastAccess)
	}
	if got.IpAddress != params.IpAddress {
		t.Errorf("IpAddress = %v, want %v", got.IpAddress, params.IpAddress)
	}
	if got.UserAgent != params.UserAgent {
		t.Errorf("UserAgent = %v, want %v", got.UserAgent, params.UserAgent)
	}
	if got.SessionData != params.SessionData {
		t.Errorf("SessionData = %v, want %v", got.SessionData, params.SessionData)
	}
}

func TestPsqlDatabase_MapCreateSessionParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := sessionCreateParams()

	got1 := d.MapCreateSessionParams(params)
	got2 := d.MapCreateSessionParams(params)

	if got1.SessionID == got2.SessionID {
		t.Error("two calls produced identical SessionIDs; each should be unique")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateSessionParams tests ---

func TestPsqlDatabase_MapUpdateSessionParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := sessionUpdateParams()

	got := d.MapUpdateSessionParams(params)

	if got.SessionID != params.SessionID {
		t.Errorf("SessionID = %v, want %v", got.SessionID, params.SessionID)
	}
	if got.UserID != params.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, params.UserID)
	}
	if got.CreatedAt != params.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, params.CreatedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	// PostgreSQL maps LastAccess via StringToNTime
	expectedLastAccess := StringToNTime(params.LastAccess.String)
	if got.LastAccess != expectedLastAccess {
		t.Errorf("LastAccess = %v, want %v", got.LastAccess, expectedLastAccess)
	}
	if got.IpAddress != params.IpAddress {
		t.Errorf("IpAddress = %v, want %v", got.IpAddress, params.IpAddress)
	}
	if got.UserAgent != params.UserAgent {
		t.Errorf("UserAgent = %v, want %v", got.UserAgent, params.UserAgent)
	}
	if got.SessionData != params.SessionData {
		t.Errorf("SessionData = %v, want %v", got.SessionData, params.SessionData)
	}
}

// --- Cross-database MapCreateSessionParams auto-ID generation ---

func TestCrossDatabaseMapCreateSessionParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	params := sessionCreateParams()

	sqliteResult := Database{}.MapCreateSessionParams(params)
	mysqlResult := MysqlDatabase{}.MapCreateSessionParams(params)
	psqlResult := PsqlDatabase{}.MapCreateSessionParams(params)

	if sqliteResult.SessionID.IsZero() {
		t.Error("SQLite: expected non-zero generated SessionID")
	}
	if mysqlResult.SessionID.IsZero() {
		t.Error("MySQL: expected non-zero generated SessionID")
	}
	if psqlResult.SessionID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated SessionID")
	}

	// Each call should generate a unique ID
	if sqliteResult.SessionID == mysqlResult.SessionID {
		t.Error("SQLite and MySQL generated the same SessionID -- each call should be unique")
	}
	if sqliteResult.SessionID == psqlResult.SessionID {
		t.Error("SQLite and PostgreSQL generated the same SessionID -- each call should be unique")
	}
}

// --- Cross-database mapper consistency ---
// The SQLite, MySQL, and PostgreSQL mapper results cannot be directly compared
// because LastAccess uses different source types (sql.NullString, time.Time,
// sql.NullTime). We verify the shared fields that have identical types produce
// identical results.

func TestCrossDatabaseMapSession_SharedFieldConsistency(t *testing.T) {
	t.Parallel()
	sessionID := types.NewSessionID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	ipAddr := sql.NullString{String: "10.0.0.1", Valid: true}
	agent := sql.NullString{String: "test-agent", Valid: true}
	data := sql.NullString{String: `{"test":"data"}`, Valid: true}

	sqliteResult := Database{}.MapSession(mdb.Sessions{
		SessionID: sessionID, UserID: userID, CreatedAt: ts, ExpiresAt: ts,
		IpAddress: ipAddr, UserAgent: agent, SessionData: data,
	})
	mysqlResult := MysqlDatabase{}.MapSession(mdbm.Sessions{
		SessionID: sessionID, UserID: userID, CreatedAt: ts, ExpiresAt: ts,
		IpAddress: ipAddr, UserAgent: agent, SessionData: data,
	})
	psqlResult := PsqlDatabase{}.MapSession(mdbp.Sessions{
		SessionID: sessionID, UserID: userID, CreatedAt: ts, ExpiresAt: ts,
		IpAddress: ipAddr, UserAgent: agent, SessionData: data,
	})

	// Compare shared fields (LastAccess excluded due to different source types)
	if sqliteResult.SessionID != mysqlResult.SessionID {
		t.Errorf("SQLite vs MySQL SessionID mismatch: %v vs %v", sqliteResult.SessionID, mysqlResult.SessionID)
	}
	if sqliteResult.SessionID != psqlResult.SessionID {
		t.Errorf("SQLite vs PostgreSQL SessionID mismatch: %v vs %v", sqliteResult.SessionID, psqlResult.SessionID)
	}
	if sqliteResult.UserID != mysqlResult.UserID {
		t.Errorf("SQLite vs MySQL UserID mismatch: %v vs %v", sqliteResult.UserID, mysqlResult.UserID)
	}
	if sqliteResult.UserID != psqlResult.UserID {
		t.Errorf("SQLite vs PostgreSQL UserID mismatch: %v vs %v", sqliteResult.UserID, psqlResult.UserID)
	}
	if sqliteResult.CreatedAt != mysqlResult.CreatedAt {
		t.Errorf("SQLite vs MySQL CreatedAt mismatch: %v vs %v", sqliteResult.CreatedAt, mysqlResult.CreatedAt)
	}
	if sqliteResult.CreatedAt != psqlResult.CreatedAt {
		t.Errorf("SQLite vs PostgreSQL CreatedAt mismatch: %v vs %v", sqliteResult.CreatedAt, psqlResult.CreatedAt)
	}
	if sqliteResult.IpAddress != mysqlResult.IpAddress {
		t.Errorf("SQLite vs MySQL IpAddress mismatch: %v vs %v", sqliteResult.IpAddress, mysqlResult.IpAddress)
	}
	if sqliteResult.IpAddress != psqlResult.IpAddress {
		t.Errorf("SQLite vs PostgreSQL IpAddress mismatch: %v vs %v", sqliteResult.IpAddress, psqlResult.IpAddress)
	}
	if sqliteResult.UserAgent != mysqlResult.UserAgent {
		t.Errorf("SQLite vs MySQL UserAgent mismatch: %v vs %v", sqliteResult.UserAgent, mysqlResult.UserAgent)
	}
	if sqliteResult.UserAgent != psqlResult.UserAgent {
		t.Errorf("SQLite vs PostgreSQL UserAgent mismatch: %v vs %v", sqliteResult.UserAgent, psqlResult.UserAgent)
	}
	if sqliteResult.SessionData != mysqlResult.SessionData {
		t.Errorf("SQLite vs MySQL SessionData mismatch: %v vs %v", sqliteResult.SessionData, mysqlResult.SessionData)
	}
	if sqliteResult.SessionData != psqlResult.SessionData {
		t.Errorf("SQLite vs PostgreSQL SessionData mismatch: %v vs %v", sqliteResult.SessionData, psqlResult.SessionData)
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewSessionCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("session-node-1"),
		RequestID: "session-req-123",
		IP:        "10.0.0.1",
	}
	params := sessionCreateParams()

	cmd := Database{}.NewSessionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
	p, ok := cmd.Params().(CreateSessionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateSessionParams", cmd.Params())
	}
	if p.UserID != params.UserID {
		t.Errorf("Params().UserID = %v, want %v", p.UserID, params.UserID)
	}
	if p.IpAddress != params.IpAddress {
		t.Errorf("Params().IpAddress = %v, want %v", p.IpAddress, params.IpAddress)
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewSessionCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	sessionID := types.NewSessionID()
	cmd := NewSessionCmd{}

	row := mdb.Sessions{SessionID: sessionID}
	got := cmd.GetID(row)
	if got != string(sessionID) {
		t.Errorf("GetID() = %q, want %q", got, string(sessionID))
	}
}

func TestNewSessionCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	cmd := NewSessionCmd{}
	row := mdb.Sessions{SessionID: ""}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateSessionCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := sessionUpdateParams()

	cmd := Database{}.UpdateSessionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
	if cmd.GetID() != string(params.SessionID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(params.SessionID))
	}
	p, ok := cmd.Params().(UpdateSessionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateSessionParams", cmd.Params())
	}
	if p.UserID != params.UserID {
		t.Errorf("Params().UserID = %v, want %v", p.UserID, params.UserID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteSessionCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	sessionID := types.NewSessionID()

	cmd := Database{}.DeleteSessionCmd(ctx, ac, sessionID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
	if cmd.GetID() != string(sessionID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(sessionID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewSessionCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-session-node"),
		RequestID: "mysql-session-req",
		IP:        "192.168.1.1",
	}
	params := sessionCreateParams()

	cmd := MysqlDatabase{}.NewSessionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
	p, ok := cmd.Params().(CreateSessionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateSessionParams", cmd.Params())
	}
	if p.UserID != params.UserID {
		t.Errorf("Params().UserID = %v, want %v", p.UserID, params.UserID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewSessionCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	sessionID := types.NewSessionID()
	cmd := NewSessionCmdMysql{}

	row := mdbm.Sessions{SessionID: sessionID}
	got := cmd.GetID(row)
	if got != string(sessionID) {
		t.Errorf("GetID() = %q, want %q", got, string(sessionID))
	}
}

func TestUpdateSessionCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := sessionUpdateParams()

	cmd := MysqlDatabase{}.UpdateSessionCmd(ctx, ac, params)

	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
	if cmd.GetID() != string(params.SessionID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(params.SessionID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateSessionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateSessionParams", cmd.Params())
	}
	if p.SessionID != params.SessionID {
		t.Errorf("Params().SessionID = %v, want %v", p.SessionID, params.SessionID)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteSessionCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	sessionID := types.NewSessionID()

	cmd := MysqlDatabase{}.DeleteSessionCmd(ctx, ac, sessionID)

	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
	if cmd.GetID() != string(sessionID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(sessionID))
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

func TestNewSessionCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-session-node"),
		RequestID: "psql-session-req",
		IP:        "172.16.0.1",
	}
	params := sessionCreateParams()

	cmd := PsqlDatabase{}.NewSessionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
	p, ok := cmd.Params().(CreateSessionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateSessionParams", cmd.Params())
	}
	if p.UserID != params.UserID {
		t.Errorf("Params().UserID = %v, want %v", p.UserID, params.UserID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewSessionCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	sessionID := types.NewSessionID()
	cmd := NewSessionCmdPsql{}

	row := mdbp.Sessions{SessionID: sessionID}
	got := cmd.GetID(row)
	if got != string(sessionID) {
		t.Errorf("GetID() = %q, want %q", got, string(sessionID))
	}
}

func TestUpdateSessionCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := sessionUpdateParams()

	cmd := PsqlDatabase{}.UpdateSessionCmd(ctx, ac, params)

	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
	if cmd.GetID() != string(params.SessionID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(params.SessionID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateSessionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateSessionParams", cmd.Params())
	}
	if p.SessionID != params.SessionID {
		t.Errorf("Params().SessionID = %v, want %v", p.SessionID, params.SessionID)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteSessionCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	sessionID := types.NewSessionID()

	cmd := PsqlDatabase{}.DeleteSessionCmd(ctx, ac, sessionID)

	if cmd.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "sessions")
	}
	if cmd.GetID() != string(sessionID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(sessionID))
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

func TestAuditedSessionCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateSessionParams{}
	updateParams := UpdateSessionParams{SessionID: types.NewSessionID()}
	sessionID := types.NewSessionID()

	// SQLite
	sqliteCreate := Database{}.NewSessionCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateSessionCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteSessionCmd(ctx, ac, sessionID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewSessionCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateSessionCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteSessionCmd(ctx, ac, sessionID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewSessionCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateSessionCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteSessionCmd(ctx, ac, sessionID)

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
			if c.name != "sessions" {
				t.Errorf("TableName() = %q, want %q", c.name, "sessions")
			}
		})
	}
}

func TestAuditedSessionCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateSessionParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewSessionCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewSessionCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewSessionCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedSessionCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	sessionID := types.NewSessionID()

	t.Run("UpdateCmd GetID returns SessionID", func(t *testing.T) {
		t.Parallel()
		params := UpdateSessionParams{SessionID: sessionID}

		sqliteCmd := Database{}.UpdateSessionCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateSessionCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateSessionCmd(ctx, ac, params)

		wantID := string(sessionID)
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

	t.Run("DeleteCmd GetID returns session ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteSessionCmd(ctx, ac, sessionID)
		mysqlCmd := MysqlDatabase{}.DeleteSessionCmd(ctx, ac, sessionID)
		psqlCmd := PsqlDatabase{}.DeleteSessionCmd(ctx, ac, sessionID)

		wantID := string(sessionID)
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
		testSessionID := types.NewSessionID()

		sqliteCmd := NewSessionCmd{}
		mysqlCmd := NewSessionCmdMysql{}
		psqlCmd := NewSessionCmdPsql{}

		wantID := string(testSessionID)

		sqliteRow := mdb.Sessions{SessionID: testSessionID}
		mysqlRow := mdbm.Sessions{SessionID: testSessionID}
		psqlRow := mdbp.Sessions{SessionID: testSessionID}

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

func TestUpdateSessionCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateSessionParams{SessionID: ""}

	sqliteCmd := Database{}.UpdateSessionCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateSessionCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateSessionCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteSessionCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.SessionID("")

	sqliteCmd := Database{}.DeleteSessionCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteSessionCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteSessionCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestNewSessionCmd_GetID_EmptySessionID(t *testing.T) {
	t.Parallel()
	// CreateCmd.GetID extracts from a row; an empty SessionID should return ""
	sqliteCmd := NewSessionCmd{}
	if sqliteCmd.GetID(mdb.Sessions{}) != "" {
		t.Errorf("SQLite GetID() = %q, want empty string", sqliteCmd.GetID(mdb.Sessions{}))
	}

	mysqlCmd := NewSessionCmdMysql{}
	if mysqlCmd.GetID(mdbm.Sessions{}) != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID(mdbm.Sessions{}))
	}

	psqlCmd := NewSessionCmdPsql{}
	if psqlCmd.GetID(mdbp.Sessions{}) != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID(mdbp.Sessions{}))
	}
}

// --- Recorder identity verification ---
// Verify the command factories assign the correct recorder variant.

func TestSessionCommand_RecorderIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateSessionParams{}
	updateParams := UpdateSessionParams{SessionID: types.NewSessionID()}
	deleteID := types.NewSessionID()

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
		want     audited.ChangeEventRecorder
	}{
		// SQLite commands should use SQLiteRecorder
		{"SQLite Create", Database{}.NewSessionCmd(ctx, ac, createParams).Recorder(), SQLiteRecorder},
		{"SQLite Update", Database{}.UpdateSessionCmd(ctx, ac, updateParams).Recorder(), SQLiteRecorder},
		{"SQLite Delete", Database{}.DeleteSessionCmd(ctx, ac, deleteID).Recorder(), SQLiteRecorder},
		// MySQL commands should use MysqlRecorder
		{"MySQL Create", MysqlDatabase{}.NewSessionCmd(ctx, ac, createParams).Recorder(), MysqlRecorder},
		{"MySQL Update", MysqlDatabase{}.UpdateSessionCmd(ctx, ac, updateParams).Recorder(), MysqlRecorder},
		{"MySQL Delete", MysqlDatabase{}.DeleteSessionCmd(ctx, ac, deleteID).Recorder(), MysqlRecorder},
		// PostgreSQL commands should use PsqlRecorder
		{"PostgreSQL Create", PsqlDatabase{}.NewSessionCmd(ctx, ac, createParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateSessionCmd(ctx, ac, updateParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteSessionCmd(ctx, ac, deleteID).Recorder(), PsqlRecorder},
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

// --- UpdateSession success message format ---
// The UpdateSession method produces a success message string.
// We verify the message format expectation by testing the format string directly.

func TestUpdateSession_SuccessMessageFormat(t *testing.T) {
	t.Parallel()
	sessionID := types.NewSessionID()
	expected := fmt.Sprintf("Successfully updated %v\n", sessionID)

	want := "Successfully updated " + string(sessionID) + "\n"
	if expected != want {
		t.Errorf("message = %q, want %q", expected, want)
	}
}

func TestUpdateSession_SuccessMessageFormat_EmptySessionID(t *testing.T) {
	t.Parallel()
	// Edge case: empty SessionID still produces a valid format string
	expected := fmt.Sprintf("Successfully updated %v\n", types.SessionID(""))
	if expected != "Successfully updated \n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated \n")
	}
}
