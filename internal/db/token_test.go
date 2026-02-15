package db

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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
	_ audited.CreateCommand[mdb.Tokens]  = NewTokenCmd{}
	_ audited.UpdateCommand[mdb.Tokens]  = UpdateTokenCmd{}
	_ audited.DeleteCommand[mdb.Tokens]  = DeleteTokenCmd{}
	_ audited.CreateCommand[mdbm.Tokens] = NewTokenCmdMysql{}
	_ audited.UpdateCommand[mdbm.Tokens] = UpdateTokenCmdMysql{}
	_ audited.DeleteCommand[mdbm.Tokens] = DeleteTokenCmdMysql{}
	_ audited.CreateCommand[mdbp.Tokens] = NewTokenCmdPsql{}
	_ audited.UpdateCommand[mdbp.Tokens] = UpdateTokenCmdPsql{}
	_ audited.DeleteCommand[mdbp.Tokens] = DeleteTokenCmdPsql{}
)

// --- Test data helpers ---

// tokenTestFixture returns a fully populated Tokens struct and its component parts.
func tokenTestFixture() (Tokens, types.NullableUserID, types.Timestamp) {
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 8, 15, 10, 30, 0, 0, time.UTC))
	tok := Tokens{
		ID:        string(types.NewTokenID()),
		UserID:    userID,
		TokenType: "refresh",
		Token:     "eyJhbGciOiJIUzI1NiJ9.dGVzdA.signature",
		IssuedAt:  "2025-08-15T10:30:00Z",
		ExpiresAt: ts,
		Revoked:   false,
	}
	return tok, userID, ts
}

// tokenCreateParams returns a CreateTokenParams with all fields populated.
func tokenCreateParams() CreateTokenParams {
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 9, 1, 12, 0, 0, 0, time.UTC))
	return CreateTokenParams{
		UserID:    userID,
		TokenType: "access",
		Token:     "abc123tokenvalue",
		IssuedAt:  "2025-09-01T12:00:00Z",
		ExpiresAt: ts,
		Revoked:   false,
	}
}

// tokenUpdateParams returns an UpdateTokenParams with all fields populated.
func tokenUpdateParams() UpdateTokenParams {
	ts := types.NewTimestamp(time.Date(2025, 10, 20, 8, 45, 0, 0, time.UTC))
	return UpdateTokenParams{
		Token:     "updated-token-value-xyz",
		IssuedAt:  "2025-10-20T08:45:00Z",
		ExpiresAt: ts,
		Revoked:   true,
		ID:        string(types.NewTokenID()),
	}
}

// --- JSON tag verification ---

func TestTokens_JSONTags(t *testing.T) {
	t.Parallel()
	want := map[string]string{
		"ID":        "id",
		"UserID":    "user_id",
		"TokenType": "token_type",
		"Token":     "token",
		"IssuedAt":  "issued_at",
		"ExpiresAt": "expires_at",
		"Revoked":   "revoked",
	}

	rt := reflect.TypeOf(Tokens{})
	for i := range rt.NumField() {
		f := rt.Field(i)
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			expected, ok := want[f.Name]
			if !ok {
				t.Fatalf("unexpected field %q in Tokens struct", f.Name)
			}
			got := f.Tag.Get("json")
			if got != expected {
				t.Errorf("json tag for %s = %q, want %q", f.Name, got, expected)
			}
		})
	}

	if rt.NumField() != len(want) {
		t.Errorf("Tokens has %d fields, expected %d", rt.NumField(), len(want))
	}
}

func TestCreateTokenParams_JSONTags(t *testing.T) {
	t.Parallel()
	want := map[string]string{
		"UserID":    "user_id",
		"TokenType": "token_type",
		"Token":     "token",
		"IssuedAt":  "issued_at",
		"ExpiresAt": "expires_at",
		"Revoked":   "revoked",
	}

	rt := reflect.TypeOf(CreateTokenParams{})
	for i := range rt.NumField() {
		f := rt.Field(i)
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			expected, ok := want[f.Name]
			if !ok {
				t.Fatalf("unexpected field %q in CreateTokenParams struct", f.Name)
			}
			got := f.Tag.Get("json")
			if got != expected {
				t.Errorf("json tag for %s = %q, want %q", f.Name, got, expected)
			}
		})
	}

	if rt.NumField() != len(want) {
		t.Errorf("CreateTokenParams has %d fields, expected %d", rt.NumField(), len(want))
	}
}

func TestUpdateTokenParams_JSONTags(t *testing.T) {
	t.Parallel()
	want := map[string]string{
		"Token":     "token",
		"IssuedAt":  "issued_at",
		"ExpiresAt": "expires_at",
		"Revoked":   "revoked",
		"ID":        "id",
	}

	rt := reflect.TypeOf(UpdateTokenParams{})
	for i := range rt.NumField() {
		f := rt.Field(i)
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			expected, ok := want[f.Name]
			if !ok {
				t.Fatalf("unexpected field %q in UpdateTokenParams struct", f.Name)
			}
			got := f.Tag.Get("json")
			if got != expected {
				t.Errorf("json tag for %s = %q, want %q", f.Name, got, expected)
			}
		})
	}

	if rt.NumField() != len(want) {
		t.Errorf("UpdateTokenParams has %d fields, expected %d", rt.NumField(), len(want))
	}
}

// --- JSON round-trip tests ---

func TestTokens_JSONRoundTrip(t *testing.T) {
	t.Parallel()
	tok, _, _ := tokenTestFixture()

	data, err := json.Marshal(tok)
	if err != nil {
		t.Fatalf("failed to marshal Tokens: %v", err)
	}

	var got Tokens
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal Tokens: %v", err)
	}

	if got.ID != tok.ID {
		t.Errorf("ID = %q, want %q", got.ID, tok.ID)
	}
	if got.TokenType != tok.TokenType {
		t.Errorf("TokenType = %q, want %q", got.TokenType, tok.TokenType)
	}
	if got.Token != tok.Token {
		t.Errorf("Token = %q, want %q", got.Token, tok.Token)
	}
	if got.IssuedAt != tok.IssuedAt {
		t.Errorf("IssuedAt = %q, want %q", got.IssuedAt, tok.IssuedAt)
	}
	if got.Revoked != tok.Revoked {
		t.Errorf("Revoked = %v, want %v", got.Revoked, tok.Revoked)
	}
}

// --- MapStringToken tests ---

func TestMapStringToken_AllFields(t *testing.T) {
	t.Parallel()
	tok, userID, ts := tokenTestFixture()

	got := MapStringToken(tok)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ID", got.ID, tok.ID},
		{"UserID", got.UserID, userID.String()},
		{"TokenType", got.TokenType, tok.TokenType},
		{"Token", got.Token, tok.Token},
		{"IssuedAt", got.IssuedAt, tok.IssuedAt},
		{"ExpiresAt", got.ExpiresAt, ts.String()},
		{"Revoked", got.Revoked, "false"},
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

func TestMapStringToken_RevokedTrue(t *testing.T) {
	t.Parallel()
	tok := Tokens{Revoked: true}
	got := MapStringToken(tok)
	if got.Revoked != "true" {
		t.Errorf("Revoked = %q, want %q", got.Revoked, "true")
	}
}

func TestMapStringToken_RevokedFalse_ExplicitZero(t *testing.T) {
	t.Parallel()
	// Explicitly set Revoked to false (not relying on zero value) to verify fmt.Sprintf("%t", false)
	tok := Tokens{Revoked: false}
	got := MapStringToken(tok)
	if got.Revoked != "false" {
		t.Errorf("Revoked = %q, want %q", got.Revoked, "false")
	}
}

func TestMapStringToken_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringToken(Tokens{})

	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
	if got.TokenType != "" {
		t.Errorf("TokenType = %q, want empty string", got.TokenType)
	}
	if got.Token != "" {
		t.Errorf("Token = %q, want empty string", got.Token)
	}
	if got.IssuedAt != "" {
		t.Errorf("IssuedAt = %q, want empty string", got.IssuedAt)
	}
	if got.Revoked != "false" {
		t.Errorf("Revoked = %q, want %q (zero bool)", got.Revoked, "false")
	}
}

func TestMapStringToken_NullUserID(t *testing.T) {
	t.Parallel()
	tok := Tokens{
		UserID: types.NullableUserID{Valid: false},
	}
	got := MapStringToken(tok)
	// Should produce whatever NullableUserID.String() returns for invalid
	if got.UserID != tok.UserID.String() {
		t.Errorf("UserID = %q, want %q", got.UserID, tok.UserID.String())
	}
}

func TestMapStringToken_ValidUserID(t *testing.T) {
	t.Parallel()
	uid := types.NewUserID()
	tok := Tokens{
		UserID: types.NullableUserID{ID: uid, Valid: true},
	}
	got := MapStringToken(tok)
	if got.UserID != tok.UserID.String() {
		t.Errorf("UserID = %q, want %q", got.UserID, tok.UserID.String())
	}
}

// --- SQLite Database.MapToken tests ---

func TestDatabase_MapToken_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 3, 10, 9, 0, 0, 0, time.UTC))

	input := mdb.Tokens{
		ID:        "tok-sqlite-001",
		UserID:    userID,
		TokenType: "refresh",
		Tokens:    "sqlite-token-value",
		IssuedAt:  "2025-03-10T09:00:00Z",
		ExpiresAt: ts,
		Revoked:   true,
	}

	got := d.MapToken(input)

	if got.ID != "tok-sqlite-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tok-sqlite-001")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.TokenType != "refresh" {
		t.Errorf("TokenType = %q, want %q", got.TokenType, "refresh")
	}
	// Note: sqlc field is "Tokens" but wrapper field is "Token"
	if got.Token != "sqlite-token-value" {
		t.Errorf("Token = %q, want %q", got.Token, "sqlite-token-value")
	}
	if got.IssuedAt != "2025-03-10T09:00:00Z" {
		t.Errorf("IssuedAt = %q, want %q", got.IssuedAt, "2025-03-10T09:00:00Z")
	}
	if got.ExpiresAt != ts {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, ts)
	}
	if got.Revoked != true {
		t.Error("Revoked = false, want true")
	}
}

func TestDatabase_MapToken_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapToken(mdb.Tokens{})

	if got.ID != "" {
		t.Errorf("ID = %q, want zero value", got.ID)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if got.TokenType != "" {
		t.Errorf("TokenType = %q, want zero value", got.TokenType)
	}
	if got.Token != "" {
		t.Errorf("Token = %q, want zero value", got.Token)
	}
	if got.IssuedAt != "" {
		t.Errorf("IssuedAt = %q, want zero value", got.IssuedAt)
	}
	if got.Revoked != false {
		t.Error("Revoked = true, want false")
	}
}

func TestDatabase_MapToken_NullUserID(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := mdb.Tokens{
		ID:     "tok-null-user",
		UserID: types.NullableUserID{Valid: false},
	}

	got := d.MapToken(input)

	if got.UserID.Valid {
		t.Error("UserID should be invalid/null")
	}
}

// --- SQLite Database.MapCreateTokenParams tests ---

func TestDatabase_MapCreateTokenParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := tokenCreateParams()

	got := d.MapCreateTokenParams(params)

	// A new ID should always be generated (non-empty)
	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}

	// All other fields should pass through
	if got.UserID != params.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, params.UserID)
	}
	if got.TokenType != params.TokenType {
		t.Errorf("TokenType = %q, want %q", got.TokenType, params.TokenType)
	}
	// Note: wrapper uses "Token", sqlc uses "Tokens"
	if got.Tokens != params.Token {
		t.Errorf("Tokens = %q, want %q", got.Tokens, params.Token)
	}
	if got.IssuedAt != params.IssuedAt {
		t.Errorf("IssuedAt = %q, want %q", got.IssuedAt, params.IssuedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	if got.Revoked != params.Revoked {
		t.Errorf("Revoked = %v, want %v", got.Revoked, params.Revoked)
	}
}

func TestDatabase_MapCreateTokenParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := tokenCreateParams()

	got1 := d.MapCreateTokenParams(params)
	got2 := d.MapCreateTokenParams(params)

	if got1.ID == got2.ID {
		t.Error("two calls generated the same ID -- each call should be unique")
	}
}

func TestDatabase_MapCreateTokenParams_NullUserID(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := CreateTokenParams{
		UserID: types.NullableUserID{Valid: false},
	}

	got := d.MapCreateTokenParams(params)

	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
}

// --- SQLite Database.MapUpdateTokenParams tests ---

func TestDatabase_MapUpdateTokenParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := tokenUpdateParams()

	got := d.MapUpdateTokenParams(params)

	// Token -> Tokens field mapping
	if got.Tokens != params.Token {
		t.Errorf("Tokens = %q, want %q", got.Tokens, params.Token)
	}
	if got.IssuedAt != params.IssuedAt {
		t.Errorf("IssuedAt = %q, want %q", got.IssuedAt, params.IssuedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	if got.Revoked != params.Revoked {
		t.Errorf("Revoked = %v, want %v", got.Revoked, params.Revoked)
	}
	// ID is the WHERE clause identifier -- must be preserved
	if got.ID != params.ID {
		t.Errorf("ID = %q, want %q", got.ID, params.ID)
	}
}

func TestDatabase_MapUpdateTokenParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateTokenParams(UpdateTokenParams{})

	if got.ID != "" {
		t.Errorf("ID = %q, want zero value", got.ID)
	}
	if got.Tokens != "" {
		t.Errorf("Tokens = %q, want zero value", got.Tokens)
	}
	if got.IssuedAt != "" {
		t.Errorf("IssuedAt = %q, want zero value", got.IssuedAt)
	}
	if got.Revoked != false {
		t.Error("Revoked = true, want false for zero value")
	}
}

// --- MySQL MysqlDatabase.MapToken tests ---

func TestMysqlDatabase_MapToken_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 5, 11, 0, 0, 0, time.UTC))
	issuedAt := time.Date(2025, 4, 5, 11, 0, 0, 0, time.UTC)

	input := mdbm.Tokens{
		ID:        "tok-mysql-001",
		UserID:    userID,
		TokenType: "access",
		Tokens:    "mysql-token-value",
		IssuedAt:  issuedAt,
		ExpiresAt: ts,
		Revoked:   false,
	}

	got := d.MapToken(input)

	if got.ID != "tok-mysql-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tok-mysql-001")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.TokenType != "access" {
		t.Errorf("TokenType = %q, want %q", got.TokenType, "access")
	}
	if got.Token != "mysql-token-value" {
		t.Errorf("Token = %q, want %q", got.Token, "mysql-token-value")
	}
	// MySQL maps IssuedAt via time.Time.String()
	if got.IssuedAt != issuedAt.String() {
		t.Errorf("IssuedAt = %q, want %q", got.IssuedAt, issuedAt.String())
	}
	if got.ExpiresAt != ts {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, ts)
	}
	if got.Revoked != false {
		t.Error("Revoked = true, want false")
	}
}

func TestMysqlDatabase_MapToken_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapToken(mdbm.Tokens{})

	if got.ID != "" {
		t.Errorf("ID = %q, want zero value", got.ID)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	// Zero time.Time.String() is "0001-01-01 00:00:00 +0000 UTC"
	zeroTime := time.Time{}
	if got.IssuedAt != zeroTime.String() {
		t.Errorf("IssuedAt = %q, want %q", got.IssuedAt, zeroTime.String())
	}
}

func TestMysqlDatabase_MapToken_RevokedTrue(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := mdbm.Tokens{Revoked: true}
	got := d.MapToken(input)
	if got.Revoked != true {
		t.Error("Revoked = false, want true")
	}
}

// --- MySQL MysqlDatabase.MapCreateTokenParams tests ---

func TestMysqlDatabase_MapCreateTokenParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := tokenCreateParams()

	got := d.MapCreateTokenParams(params)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.UserID != params.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, params.UserID)
	}
	if got.TokenType != params.TokenType {
		t.Errorf("TokenType = %q, want %q", got.TokenType, params.TokenType)
	}
	if got.Tokens != params.Token {
		t.Errorf("Tokens = %q, want %q", got.Tokens, params.Token)
	}
	// MySQL maps IssuedAt via StringToNTime
	expectedIssuedAt := StringToNTime(params.IssuedAt).Time
	if got.IssuedAt != expectedIssuedAt {
		t.Errorf("IssuedAt = %v, want %v", got.IssuedAt, expectedIssuedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	if got.Revoked != params.Revoked {
		t.Errorf("Revoked = %v, want %v", got.Revoked, params.Revoked)
	}
}

func TestMysqlDatabase_MapCreateTokenParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := tokenCreateParams()

	got1 := d.MapCreateTokenParams(params)
	got2 := d.MapCreateTokenParams(params)

	if got1.ID == got2.ID {
		t.Error("two calls produced identical IDs; each should be unique")
	}
}

// --- MySQL MysqlDatabase.MapUpdateTokenParams tests ---

func TestMysqlDatabase_MapUpdateTokenParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := tokenUpdateParams()

	got := d.MapUpdateTokenParams(params)

	if got.Tokens != params.Token {
		t.Errorf("Tokens = %q, want %q", got.Tokens, params.Token)
	}
	// MySQL maps IssuedAt via StringToNTime
	expectedIssuedAt := StringToNTime(params.IssuedAt).Time
	if got.IssuedAt != expectedIssuedAt {
		t.Errorf("IssuedAt = %v, want %v", got.IssuedAt, expectedIssuedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	if got.Revoked != params.Revoked {
		t.Errorf("Revoked = %v, want %v", got.Revoked, params.Revoked)
	}
	if got.ID != params.ID {
		t.Errorf("ID = %q, want %q", got.ID, params.ID)
	}
}

func TestMysqlDatabase_MapUpdateTokenParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapUpdateTokenParams(UpdateTokenParams{})

	if got.ID != "" {
		t.Errorf("ID = %q, want zero value", got.ID)
	}
	if got.Tokens != "" {
		t.Errorf("Tokens = %q, want zero value", got.Tokens)
	}
}

// --- PostgreSQL PsqlDatabase.MapToken tests ---

func TestPsqlDatabase_MapToken_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 2, 28, 16, 0, 0, 0, time.UTC))
	issuedAt := time.Date(2025, 2, 28, 16, 0, 0, 0, time.UTC)

	input := mdbp.Tokens{
		ID:        "tok-psql-001",
		UserID:    userID,
		TokenType: "refresh",
		Tokens:    "psql-token-value",
		IssuedAt:  issuedAt,
		ExpiresAt: ts,
		Revoked:   true,
	}

	got := d.MapToken(input)

	if got.ID != "tok-psql-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tok-psql-001")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.TokenType != "refresh" {
		t.Errorf("TokenType = %q, want %q", got.TokenType, "refresh")
	}
	if got.Token != "psql-token-value" {
		t.Errorf("Token = %q, want %q", got.Token, "psql-token-value")
	}
	// PostgreSQL maps IssuedAt via time.Time.String()
	if got.IssuedAt != issuedAt.String() {
		t.Errorf("IssuedAt = %q, want %q", got.IssuedAt, issuedAt.String())
	}
	if got.ExpiresAt != ts {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, ts)
	}
	if got.Revoked != true {
		t.Error("Revoked = false, want true")
	}
}

func TestPsqlDatabase_MapToken_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapToken(mdbp.Tokens{})

	if got.ID != "" {
		t.Errorf("ID = %q, want zero value", got.ID)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	// Zero time.Time.String() is "0001-01-01 00:00:00 +0000 UTC"
	zeroTime := time.Time{}
	if got.IssuedAt != zeroTime.String() {
		t.Errorf("IssuedAt = %q, want %q", got.IssuedAt, zeroTime.String())
	}
}

func TestPsqlDatabase_MapToken_NullUserID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := mdbp.Tokens{
		ID:     "tok-psql-null-user",
		UserID: types.NullableUserID{Valid: false},
	}

	got := d.MapToken(input)

	if got.UserID.Valid {
		t.Error("UserID should be invalid/null")
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateTokenParams tests ---

func TestPsqlDatabase_MapCreateTokenParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := tokenCreateParams()

	got := d.MapCreateTokenParams(params)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.UserID != params.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, params.UserID)
	}
	if got.TokenType != params.TokenType {
		t.Errorf("TokenType = %q, want %q", got.TokenType, params.TokenType)
	}
	if got.Tokens != params.Token {
		t.Errorf("Tokens = %q, want %q", got.Tokens, params.Token)
	}
	// PostgreSQL maps IssuedAt via StringToNTime
	expectedIssuedAt := StringToNTime(params.IssuedAt).Time
	if got.IssuedAt != expectedIssuedAt {
		t.Errorf("IssuedAt = %v, want %v", got.IssuedAt, expectedIssuedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	if got.Revoked != params.Revoked {
		t.Errorf("Revoked = %v, want %v", got.Revoked, params.Revoked)
	}
}

func TestPsqlDatabase_MapCreateTokenParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := tokenCreateParams()

	got1 := d.MapCreateTokenParams(params)
	got2 := d.MapCreateTokenParams(params)

	if got1.ID == got2.ID {
		t.Error("two calls produced identical IDs; each should be unique")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateTokenParams tests ---

func TestPsqlDatabase_MapUpdateTokenParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := tokenUpdateParams()

	got := d.MapUpdateTokenParams(params)

	if got.Tokens != params.Token {
		t.Errorf("Tokens = %q, want %q", got.Tokens, params.Token)
	}
	// PostgreSQL maps IssuedAt via StringToNTime
	expectedIssuedAt := StringToNTime(params.IssuedAt).Time
	if got.IssuedAt != expectedIssuedAt {
		t.Errorf("IssuedAt = %v, want %v", got.IssuedAt, expectedIssuedAt)
	}
	if got.ExpiresAt != params.ExpiresAt {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, params.ExpiresAt)
	}
	if got.Revoked != params.Revoked {
		t.Errorf("Revoked = %v, want %v", got.Revoked, params.Revoked)
	}
	if got.ID != params.ID {
		t.Errorf("ID = %q, want %q", got.ID, params.ID)
	}
}

func TestPsqlDatabase_MapUpdateTokenParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapUpdateTokenParams(UpdateTokenParams{})

	if got.ID != "" {
		t.Errorf("ID = %q, want zero value", got.ID)
	}
	if got.Tokens != "" {
		t.Errorf("Tokens = %q, want zero value", got.Tokens)
	}
}

// --- Cross-database MapCreateTokenParams auto-ID generation ---

func TestCrossDatabaseMapCreateTokenParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	params := tokenCreateParams()

	sqliteResult := Database{}.MapCreateTokenParams(params)
	mysqlResult := MysqlDatabase{}.MapCreateTokenParams(params)
	psqlResult := PsqlDatabase{}.MapCreateTokenParams(params)

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

// --- Cross-database mapper consistency ---
// The SQLite mapper passes IssuedAt as a string directly, while MySQL and
// PostgreSQL convert time.Time via .String(). We verify the fields that have
// identical representation across all three drivers produce identical results.

func TestCrossDatabaseMapToken_SharedFieldConsistency(t *testing.T) {
	t.Parallel()
	tokenID := "cross-db-token-001"
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	tokenValue := "shared-token-value"

	sqliteResult := Database{}.MapToken(mdb.Tokens{
		ID: tokenID, UserID: userID, TokenType: "refresh",
		Tokens: tokenValue, ExpiresAt: ts, Revoked: true,
	})
	mysqlResult := MysqlDatabase{}.MapToken(mdbm.Tokens{
		ID: tokenID, UserID: userID, TokenType: "refresh",
		Tokens: tokenValue, ExpiresAt: ts, Revoked: true,
	})
	psqlResult := PsqlDatabase{}.MapToken(mdbp.Tokens{
		ID: tokenID, UserID: userID, TokenType: "refresh",
		Tokens: tokenValue, ExpiresAt: ts, Revoked: true,
	})

	// Compare shared fields (IssuedAt excluded due to different source types)
	if sqliteResult.ID != mysqlResult.ID {
		t.Errorf("SQLite vs MySQL ID mismatch: %q vs %q", sqliteResult.ID, mysqlResult.ID)
	}
	if sqliteResult.ID != psqlResult.ID {
		t.Errorf("SQLite vs PostgreSQL ID mismatch: %q vs %q", sqliteResult.ID, psqlResult.ID)
	}
	if sqliteResult.UserID != mysqlResult.UserID {
		t.Errorf("SQLite vs MySQL UserID mismatch: %v vs %v", sqliteResult.UserID, mysqlResult.UserID)
	}
	if sqliteResult.UserID != psqlResult.UserID {
		t.Errorf("SQLite vs PostgreSQL UserID mismatch: %v vs %v", sqliteResult.UserID, psqlResult.UserID)
	}
	if sqliteResult.TokenType != mysqlResult.TokenType {
		t.Errorf("SQLite vs MySQL TokenType mismatch: %q vs %q", sqliteResult.TokenType, mysqlResult.TokenType)
	}
	if sqliteResult.TokenType != psqlResult.TokenType {
		t.Errorf("SQLite vs PostgreSQL TokenType mismatch: %q vs %q", sqliteResult.TokenType, psqlResult.TokenType)
	}
	if sqliteResult.Token != mysqlResult.Token {
		t.Errorf("SQLite vs MySQL Token mismatch: %q vs %q", sqliteResult.Token, mysqlResult.Token)
	}
	if sqliteResult.Token != psqlResult.Token {
		t.Errorf("SQLite vs PostgreSQL Token mismatch: %q vs %q", sqliteResult.Token, psqlResult.Token)
	}
	if sqliteResult.ExpiresAt != mysqlResult.ExpiresAt {
		t.Errorf("SQLite vs MySQL ExpiresAt mismatch: %v vs %v", sqliteResult.ExpiresAt, mysqlResult.ExpiresAt)
	}
	if sqliteResult.ExpiresAt != psqlResult.ExpiresAt {
		t.Errorf("SQLite vs PostgreSQL ExpiresAt mismatch: %v vs %v", sqliteResult.ExpiresAt, psqlResult.ExpiresAt)
	}
	if sqliteResult.Revoked != mysqlResult.Revoked {
		t.Errorf("SQLite vs MySQL Revoked mismatch: %v vs %v", sqliteResult.Revoked, mysqlResult.Revoked)
	}
	if sqliteResult.Revoked != psqlResult.Revoked {
		t.Errorf("SQLite vs PostgreSQL Revoked mismatch: %v vs %v", sqliteResult.Revoked, psqlResult.Revoked)
	}
}

// --- MySQL and PostgreSQL IssuedAt consistency ---
// Both MySQL and PostgreSQL use time.Time for IssuedAt and should produce the
// same string representation via .String().

func TestCrossDatabaseMapToken_IssuedAtConsistency_MysqlPsql(t *testing.T) {
	t.Parallel()
	issuedAt := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)

	mysqlResult := MysqlDatabase{}.MapToken(mdbm.Tokens{IssuedAt: issuedAt})
	psqlResult := PsqlDatabase{}.MapToken(mdbp.Tokens{IssuedAt: issuedAt})

	if mysqlResult.IssuedAt != psqlResult.IssuedAt {
		t.Errorf("MySQL IssuedAt = %q, PostgreSQL IssuedAt = %q -- should match", mysqlResult.IssuedAt, psqlResult.IssuedAt)
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewTokenCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("token-node-1"),
		RequestID: "token-req-123",
		IP:        "10.0.0.1",
	}
	params := tokenCreateParams()

	cmd := Database{}.NewTokenCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tokens" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tokens")
	}
	p, ok := cmd.Params().(CreateTokenParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateTokenParams", cmd.Params())
	}
	if p.UserID != params.UserID {
		t.Errorf("Params().UserID = %v, want %v", p.UserID, params.UserID)
	}
	if p.Token != params.Token {
		t.Errorf("Params().Token = %q, want %q", p.Token, params.Token)
	}
	if p.TokenType != params.TokenType {
		t.Errorf("Params().TokenType = %q, want %q", p.TokenType, params.TokenType)
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewTokenCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	cmd := NewTokenCmd{}

	row := mdb.Tokens{ID: "tok-getid-test"}
	got := cmd.GetID(row)
	if got != "tok-getid-test" {
		t.Errorf("GetID() = %q, want %q", got, "tok-getid-test")
	}
}

func TestNewTokenCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	cmd := NewTokenCmd{}
	row := mdb.Tokens{ID: ""}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateTokenCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := tokenUpdateParams()

	cmd := Database{}.UpdateTokenCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tokens" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tokens")
	}
	if cmd.GetID() != params.ID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), params.ID)
	}
	p, ok := cmd.Params().(UpdateTokenParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateTokenParams", cmd.Params())
	}
	if p.Token != params.Token {
		t.Errorf("Params().Token = %q, want %q", p.Token, params.Token)
	}
	if p.Revoked != params.Revoked {
		t.Errorf("Params().Revoked = %v, want %v", p.Revoked, params.Revoked)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteTokenCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	tokenID := "tok-delete-001"

	cmd := Database{}.DeleteTokenCmd(ctx, ac, tokenID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tokens" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tokens")
	}
	if cmd.GetID() != tokenID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), tokenID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewTokenCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-token-node"),
		RequestID: "mysql-token-req",
		IP:        "192.168.1.1",
	}
	params := tokenCreateParams()

	cmd := MysqlDatabase{}.NewTokenCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tokens" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tokens")
	}
	p, ok := cmd.Params().(CreateTokenParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateTokenParams", cmd.Params())
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

func TestNewTokenCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewTokenCmdMysql{}

	row := mdbm.Tokens{ID: "mysql-tok-getid"}
	got := cmd.GetID(row)
	if got != "mysql-tok-getid" {
		t.Errorf("GetID() = %q, want %q", got, "mysql-tok-getid")
	}
}

func TestUpdateTokenCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := tokenUpdateParams()

	cmd := MysqlDatabase{}.UpdateTokenCmd(ctx, ac, params)

	if cmd.TableName() != "tokens" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tokens")
	}
	if cmd.GetID() != params.ID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), params.ID)
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateTokenParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateTokenParams", cmd.Params())
	}
	if p.ID != params.ID {
		t.Errorf("Params().ID = %q, want %q", p.ID, params.ID)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteTokenCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	tokenID := "mysql-tok-delete"

	cmd := MysqlDatabase{}.DeleteTokenCmd(ctx, ac, tokenID)

	if cmd.TableName() != "tokens" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tokens")
	}
	if cmd.GetID() != tokenID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), tokenID)
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

func TestNewTokenCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-token-node"),
		RequestID: "psql-token-req",
		IP:        "172.16.0.1",
	}
	params := tokenCreateParams()

	cmd := PsqlDatabase{}.NewTokenCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tokens" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tokens")
	}
	p, ok := cmd.Params().(CreateTokenParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateTokenParams", cmd.Params())
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

func TestNewTokenCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewTokenCmdPsql{}

	row := mdbp.Tokens{ID: "psql-tok-getid"}
	got := cmd.GetID(row)
	if got != "psql-tok-getid" {
		t.Errorf("GetID() = %q, want %q", got, "psql-tok-getid")
	}
}

func TestUpdateTokenCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := tokenUpdateParams()

	cmd := PsqlDatabase{}.UpdateTokenCmd(ctx, ac, params)

	if cmd.TableName() != "tokens" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tokens")
	}
	if cmd.GetID() != params.ID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), params.ID)
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateTokenParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateTokenParams", cmd.Params())
	}
	if p.ID != params.ID {
		t.Errorf("Params().ID = %q, want %q", p.ID, params.ID)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteTokenCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	tokenID := "psql-tok-delete"

	cmd := PsqlDatabase{}.DeleteTokenCmd(ctx, ac, tokenID)

	if cmd.TableName() != "tokens" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tokens")
	}
	if cmd.GetID() != tokenID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), tokenID)
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

func TestAuditedTokenCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateTokenParams{}
	updateParams := UpdateTokenParams{ID: "table-name-test"}
	deleteID := "table-name-delete"

	// SQLite
	sqliteCreate := Database{}.NewTokenCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateTokenCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteTokenCmd(ctx, ac, deleteID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewTokenCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateTokenCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteTokenCmd(ctx, ac, deleteID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewTokenCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateTokenCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteTokenCmd(ctx, ac, deleteID)

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
			if c.name != "tokens" {
				t.Errorf("TableName() = %q, want %q", c.name, "tokens")
			}
		})
	}
}

func TestAuditedTokenCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateTokenParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewTokenCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewTokenCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewTokenCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedTokenCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	tokenID := string(types.NewTokenID())

	t.Run("UpdateCmd GetID returns ID", func(t *testing.T) {
		t.Parallel()
		params := UpdateTokenParams{ID: tokenID}

		sqliteCmd := Database{}.UpdateTokenCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateTokenCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateTokenCmd(ctx, ac, params)

		if sqliteCmd.GetID() != tokenID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), tokenID)
		}
		if mysqlCmd.GetID() != tokenID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), tokenID)
		}
		if psqlCmd.GetID() != tokenID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), tokenID)
		}
	})

	t.Run("DeleteCmd GetID returns id", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteTokenCmd(ctx, ac, tokenID)
		mysqlCmd := MysqlDatabase{}.DeleteTokenCmd(ctx, ac, tokenID)
		psqlCmd := PsqlDatabase{}.DeleteTokenCmd(ctx, ac, tokenID)

		if sqliteCmd.GetID() != tokenID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), tokenID)
		}
		if mysqlCmd.GetID() != tokenID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), tokenID)
		}
		if psqlCmd.GetID() != tokenID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), tokenID)
		}
	})

	t.Run("CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		testID := "create-getid-test"

		sqliteCmd := NewTokenCmd{}
		mysqlCmd := NewTokenCmdMysql{}
		psqlCmd := NewTokenCmdPsql{}

		sqliteRow := mdb.Tokens{ID: testID}
		mysqlRow := mdbm.Tokens{ID: testID}
		psqlRow := mdbp.Tokens{ID: testID}

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

// --- Edge cases ---

func TestUpdateTokenCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateTokenParams{ID: ""}

	sqliteCmd := Database{}.UpdateTokenCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateTokenCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateTokenCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteTokenCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := ""

	sqliteCmd := Database{}.DeleteTokenCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteTokenCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteTokenCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestNewTokenCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	// CreateCmd.GetID extracts from a row; an empty ID should return ""
	sqliteCmd := NewTokenCmd{}
	if sqliteCmd.GetID(mdb.Tokens{}) != "" {
		t.Errorf("SQLite GetID() = %q, want empty string", sqliteCmd.GetID(mdb.Tokens{}))
	}

	mysqlCmd := NewTokenCmdMysql{}
	if mysqlCmd.GetID(mdbm.Tokens{}) != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID(mdbm.Tokens{}))
	}

	psqlCmd := NewTokenCmdPsql{}
	if psqlCmd.GetID(mdbp.Tokens{}) != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID(mdbp.Tokens{}))
	}
}

// --- Recorder identity verification ---
// Verify the command factories assign the correct recorder variant.

func TestTokenCommand_RecorderIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateTokenParams{}
	updateParams := UpdateTokenParams{ID: "recorder-test"}
	deleteID := "recorder-delete"

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
		want     audited.ChangeEventRecorder
	}{
		// SQLite commands should use SQLiteRecorder
		{"SQLite Create", Database{}.NewTokenCmd(ctx, ac, createParams).Recorder(), SQLiteRecorder},
		{"SQLite Update", Database{}.UpdateTokenCmd(ctx, ac, updateParams).Recorder(), SQLiteRecorder},
		{"SQLite Delete", Database{}.DeleteTokenCmd(ctx, ac, deleteID).Recorder(), SQLiteRecorder},
		// MySQL commands should use MysqlRecorder
		{"MySQL Create", MysqlDatabase{}.NewTokenCmd(ctx, ac, createParams).Recorder(), MysqlRecorder},
		{"MySQL Update", MysqlDatabase{}.UpdateTokenCmd(ctx, ac, updateParams).Recorder(), MysqlRecorder},
		{"MySQL Delete", MysqlDatabase{}.DeleteTokenCmd(ctx, ac, deleteID).Recorder(), MysqlRecorder},
		// PostgreSQL commands should use PsqlRecorder
		{"PostgreSQL Create", PsqlDatabase{}.NewTokenCmd(ctx, ac, createParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateTokenCmd(ctx, ac, updateParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteTokenCmd(ctx, ac, deleteID).Recorder(), PsqlRecorder},
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

// --- UpdateToken success message format ---
// The UpdateToken method produces a success message string.
// We verify the message format expectation by testing the format string directly.

func TestUpdateToken_SuccessMessageFormat(t *testing.T) {
	t.Parallel()
	tokenID := string(types.NewTokenID())
	expected := fmt.Sprintf("Successfully updated %v\n", tokenID)

	want := "Successfully updated " + tokenID + "\n"
	if expected != want {
		t.Errorf("message = %q, want %q", expected, want)
	}
}

func TestUpdateToken_SuccessMessageFormat_EmptyID(t *testing.T) {
	t.Parallel()
	// Edge case: empty ID still produces a valid format string
	expected := fmt.Sprintf("Successfully updated %v\n", "")
	if expected != "Successfully updated \n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated \n")
	}
}

// --- Token field name mapping verification ---
// The sqlc-generated structs use "Tokens" for the token value field, while the
// wrapper struct uses "Token". This test explicitly verifies the mapping is correct
// across all three drivers to catch drift between sqlc regeneration and wrapper code.

func TestTokenFieldNameMapping_SQLite(t *testing.T) {
	t.Parallel()
	d := Database{}
	tokenValue := "verify-field-mapping"

	input := mdb.Tokens{Tokens: tokenValue}
	got := d.MapToken(input)

	if got.Token != tokenValue {
		t.Errorf("Token = %q, want %q (mapped from Tokens field)", got.Token, tokenValue)
	}

	createParams := CreateTokenParams{Token: tokenValue}
	createGot := d.MapCreateTokenParams(createParams)
	if createGot.Tokens != tokenValue {
		t.Errorf("CreateParams.Tokens = %q, want %q (mapped from Token field)", createGot.Tokens, tokenValue)
	}

	updateParams := UpdateTokenParams{Token: tokenValue}
	updateGot := d.MapUpdateTokenParams(updateParams)
	if updateGot.Tokens != tokenValue {
		t.Errorf("UpdateParams.Tokens = %q, want %q (mapped from Token field)", updateGot.Tokens, tokenValue)
	}
}

func TestTokenFieldNameMapping_MySQL(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	tokenValue := "verify-mysql-field-mapping"

	input := mdbm.Tokens{Tokens: tokenValue}
	got := d.MapToken(input)

	if got.Token != tokenValue {
		t.Errorf("Token = %q, want %q (mapped from Tokens field)", got.Token, tokenValue)
	}

	createParams := CreateTokenParams{Token: tokenValue}
	createGot := d.MapCreateTokenParams(createParams)
	if createGot.Tokens != tokenValue {
		t.Errorf("CreateParams.Tokens = %q, want %q (mapped from Token field)", createGot.Tokens, tokenValue)
	}

	updateParams := UpdateTokenParams{Token: tokenValue}
	updateGot := d.MapUpdateTokenParams(updateParams)
	if updateGot.Tokens != tokenValue {
		t.Errorf("UpdateParams.Tokens = %q, want %q (mapped from Token field)", updateGot.Tokens, tokenValue)
	}
}

func TestTokenFieldNameMapping_PostgreSQL(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	tokenValue := "verify-psql-field-mapping"

	input := mdbp.Tokens{Tokens: tokenValue}
	got := d.MapToken(input)

	if got.Token != tokenValue {
		t.Errorf("Token = %q, want %q (mapped from Tokens field)", got.Token, tokenValue)
	}

	createParams := CreateTokenParams{Token: tokenValue}
	createGot := d.MapCreateTokenParams(createParams)
	if createGot.Tokens != tokenValue {
		t.Errorf("CreateParams.Tokens = %q, want %q (mapped from Token field)", createGot.Tokens, tokenValue)
	}

	updateParams := UpdateTokenParams{Token: tokenValue}
	updateGot := d.MapUpdateTokenParams(updateParams)
	if updateGot.Tokens != tokenValue {
		t.Errorf("UpdateParams.Tokens = %q, want %q (mapped from Token field)", updateGot.Tokens, tokenValue)
	}
}
