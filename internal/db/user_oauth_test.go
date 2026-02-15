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

// --- Compile-time interface checks ---
// These verify that all 9 audited command types satisfy their respective
// audited interfaces. A compilation failure here means the struct drifted
// from the interface contract.

var (
	_ audited.CreateCommand[mdb.UserOauth]  = NewUserOauthCmd{}
	_ audited.UpdateCommand[mdb.UserOauth]  = UpdateUserOauthCmd{}
	_ audited.DeleteCommand[mdb.UserOauth]  = DeleteUserOauthCmd{}
	_ audited.CreateCommand[mdbm.UserOauth] = NewUserOauthCmdMysql{}
	_ audited.UpdateCommand[mdbm.UserOauth] = UpdateUserOauthCmdMysql{}
	_ audited.DeleteCommand[mdbm.UserOauth] = DeleteUserOauthCmdMysql{}
	_ audited.CreateCommand[mdbp.UserOauth] = NewUserOauthCmdPsql{}
	_ audited.UpdateCommand[mdbp.UserOauth] = UpdateUserOauthCmdPsql{}
	_ audited.DeleteCommand[mdbp.UserOauth] = DeleteUserOauthCmdPsql{}
)

// --- Test data helpers ---

// userOauthTestFixture returns a fully populated UserOauth struct and its component parts.
func userOauthTestFixture() (UserOauth, types.UserOauthID, types.NullableUserID, types.Timestamp) {
	oauthID := types.NewUserOauthID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 8, 15, 10, 30, 0, 0, time.UTC))
	oauth := UserOauth{
		UserOauthID:         oauthID,
		UserID:              userID,
		OauthProvider:       "google",
		OauthProviderUserID: "google-user-12345",
		AccessToken:         "access-token-abc",
		RefreshToken:        "refresh-token-xyz",
		TokenExpiresAt:      "2025-09-15T10:30:00Z",
		DateCreated:         ts,
	}
	return oauth, oauthID, userID, ts
}

// --- MapStringUserOauth tests ---

func TestMapStringUserOauth_AllFields(t *testing.T) {
	t.Parallel()
	oauth, oauthID, userID, ts := userOauthTestFixture()

	got := MapStringUserOauth(oauth)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"UserOauthID", got.UserOauthID, oauthID.String()},
		{"UserID", got.UserID, userID.String()},
		{"OauthProvider", got.OauthProvider, "google"},
		{"OauthProviderUserID", got.OauthProviderUserID, "google-user-12345"},
		{"AccessToken", got.AccessToken, "access-token-abc"},
		{"RefreshToken", got.RefreshToken, "refresh-token-xyz"},
		{"TokenExpiresAt", got.TokenExpiresAt, "2025-09-15T10:30:00Z"},
		{"DateCreated", got.DateCreated, ts.String()},
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

func TestMapStringUserOauth_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringUserOauth(UserOauth{})

	if got.UserOauthID != "" {
		t.Errorf("UserOauthID = %q, want empty string", got.UserOauthID)
	}
	if got.UserID != "" {
		// NullableUserID with zero value should produce empty or its zero representation
		zeroUserID := types.NullableUserID{}
		if got.UserID != zeroUserID.String() {
			t.Errorf("UserID = %q, want %q", got.UserID, zeroUserID.String())
		}
	}
	if got.OauthProvider != "" {
		t.Errorf("OauthProvider = %q, want empty string", got.OauthProvider)
	}
	if got.OauthProviderUserID != "" {
		t.Errorf("OauthProviderUserID = %q, want empty string", got.OauthProviderUserID)
	}
	if got.AccessToken != "" {
		t.Errorf("AccessToken = %q, want empty string", got.AccessToken)
	}
	if got.RefreshToken != "" {
		t.Errorf("RefreshToken = %q, want empty string", got.RefreshToken)
	}
	if got.TokenExpiresAt != "" {
		t.Errorf("TokenExpiresAt = %q, want empty string", got.TokenExpiresAt)
	}
}

func TestMapStringUserOauth_NullUserID(t *testing.T) {
	t.Parallel()
	oauth := UserOauth{
		UserID: types.NullableUserID{Valid: false},
	}
	got := MapStringUserOauth(oauth)
	// Should produce whatever NullableUserID.String() returns for invalid
	if got.UserID != oauth.UserID.String() {
		t.Errorf("UserID = %q, want %q", got.UserID, oauth.UserID.String())
	}
}

// --- SQLite Database.MapUserOauth tests ---

func TestDatabase_MapUserOauth_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	oauthID := types.NewUserOauthID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	input := mdb.UserOauth{
		UserOAuthID:         oauthID,
		UserID:              userID,
		OauthProvider:       "github",
		OAuthProviderUserID: "gh-user-99",
		AccessToken:         "sqlite-access",
		RefreshToken:        "sqlite-refresh",
		TokenExpiresAt:      "2025-12-31T23:59:59Z",
		DateCreated:         ts,
	}

	got := d.MapUserOauth(input)

	if got.UserOauthID != oauthID {
		t.Errorf("UserOauthID = %v, want %v", got.UserOauthID, oauthID)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.OauthProvider != "github" {
		t.Errorf("OauthProvider = %q, want %q", got.OauthProvider, "github")
	}
	if got.OauthProviderUserID != "gh-user-99" {
		t.Errorf("OauthProviderUserID = %q, want %q", got.OauthProviderUserID, "gh-user-99")
	}
	if got.AccessToken != "sqlite-access" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "sqlite-access")
	}
	if got.RefreshToken != "sqlite-refresh" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "sqlite-refresh")
	}
	if got.TokenExpiresAt != "2025-12-31T23:59:59Z" {
		t.Errorf("TokenExpiresAt = %q, want %q", got.TokenExpiresAt, "2025-12-31T23:59:59Z")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestDatabase_MapUserOauth_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUserOauth(mdb.UserOauth{})

	if got.UserOauthID != "" {
		t.Errorf("UserOauthID = %v, want zero value", got.UserOauthID)
	}
	if got.OauthProvider != "" {
		t.Errorf("OauthProvider = %q, want empty string", got.OauthProvider)
	}
	if got.AccessToken != "" {
		t.Errorf("AccessToken = %q, want empty string", got.AccessToken)
	}
}

// --- SQLite Database.MapCreateUserOauthParams tests ---

func TestDatabase_MapCreateUserOauthParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := Database{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := CreateUserOauthParams{
		UserID:              userID,
		OauthProvider:       "google",
		OauthProviderUserID: "google-user-1",
		AccessToken:         "access-1",
		RefreshToken:        "refresh-1",
		TokenExpiresAt:      "2025-12-31T23:59:59Z",
		DateCreated:         ts,
	}

	got := d.MapCreateUserOauthParams(input)

	if got.UserOAuthID.IsZero() {
		t.Fatal("expected non-zero UserOAuthID to be generated")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.OauthProvider != "google" {
		t.Errorf("OauthProvider = %q, want %q", got.OauthProvider, "google")
	}
	if got.OAuthProviderUserID != "google-user-1" {
		t.Errorf("OAuthProviderUserID = %q, want %q", got.OAuthProviderUserID, "google-user-1")
	}
	if got.AccessToken != "access-1" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "access-1")
	}
	if got.RefreshToken != "refresh-1" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "refresh-1")
	}
	if got.TokenExpiresAt != "2025-12-31T23:59:59Z" {
		t.Errorf("TokenExpiresAt = %q, want %q", got.TokenExpiresAt, "2025-12-31T23:59:59Z")
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestDatabase_MapCreateUserOauthParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateUserOauthParams{}

	got1 := d.MapCreateUserOauthParams(input)
	got2 := d.MapCreateUserOauthParams(input)

	if got1.UserOAuthID == got2.UserOAuthID {
		t.Error("two consecutive calls produced the same UserOAuthID -- each call should be unique")
	}
}

// --- SQLite Database.MapUpdateUserOauthParams tests ---

func TestDatabase_MapUpdateUserOauthParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	oauthID := types.NewUserOauthID()

	input := UpdateUserOauthParams{
		AccessToken:    "new-access",
		RefreshToken:   "new-refresh",
		TokenExpiresAt: "2026-01-01T00:00:00Z",
		UserOauthID:    oauthID,
	}

	got := d.MapUpdateUserOauthParams(input)

	if got.AccessToken != "new-access" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "new-access")
	}
	if got.RefreshToken != "new-refresh" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "new-refresh")
	}
	if got.TokenExpiresAt != "2026-01-01T00:00:00Z" {
		t.Errorf("TokenExpiresAt = %q, want %q", got.TokenExpiresAt, "2026-01-01T00:00:00Z")
	}
	if got.UserOAuthID != oauthID {
		t.Errorf("UserOAuthID = %v, want %v", got.UserOAuthID, oauthID)
	}
}

// --- MySQL MysqlDatabase.MapUserOauth tests ---

func TestMysqlDatabase_MapUserOauth_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	oauthID := types.NewUserOauthID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))
	tokenExpires := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	input := mdbm.UserOauth{
		UserOAuthID:         oauthID,
		UserID:              userID,
		OauthProvider:       "microsoft",
		OAuthProviderUserID: "ms-user-42",
		AccessToken:         "mysql-access",
		RefreshToken:        "mysql-refresh",
		TokenExpiresAt:      tokenExpires,
		DateCreated:         ts,
	}

	got := d.MapUserOauth(input)

	if got.UserOauthID != oauthID {
		t.Errorf("UserOauthID = %v, want %v", got.UserOauthID, oauthID)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.OauthProvider != "microsoft" {
		t.Errorf("OauthProvider = %q, want %q", got.OauthProvider, "microsoft")
	}
	if got.OauthProviderUserID != "ms-user-42" {
		t.Errorf("OauthProviderUserID = %q, want %q", got.OauthProviderUserID, "ms-user-42")
	}
	if got.AccessToken != "mysql-access" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "mysql-access")
	}
	if got.RefreshToken != "mysql-refresh" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "mysql-refresh")
	}
	// MySQL TokenExpiresAt is time.Time, mapped via .String()
	if got.TokenExpiresAt != tokenExpires.String() {
		t.Errorf("TokenExpiresAt = %q, want %q", got.TokenExpiresAt, tokenExpires.String())
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestMysqlDatabase_MapUserOauth_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapUserOauth(mdbm.UserOauth{})

	if got.UserOauthID != "" {
		t.Errorf("UserOauthID = %v, want zero value", got.UserOauthID)
	}
	if got.OauthProvider != "" {
		t.Errorf("OauthProvider = %q, want empty string", got.OauthProvider)
	}
	// time.Time{}.String() produces a non-empty string
	zeroTimeStr := time.Time{}.String()
	if got.TokenExpiresAt != zeroTimeStr {
		t.Errorf("TokenExpiresAt = %q, want %q", got.TokenExpiresAt, zeroTimeStr)
	}
}

// --- MySQL MysqlDatabase.MapCreateUserOauthParams tests ---

func TestMysqlDatabase_MapCreateUserOauthParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := CreateUserOauthParams{
		UserID:              userID,
		OauthProvider:       "facebook",
		OauthProviderUserID: "fb-user-7",
		AccessToken:         "mysql-access-new",
		RefreshToken:        "mysql-refresh-new",
		TokenExpiresAt:      "2025-12-31T23:59:59Z",
		DateCreated:         ts,
	}

	got := d.MapCreateUserOauthParams(input)

	if got.UserOAuthID.IsZero() {
		t.Fatal("expected non-zero UserOAuthID to be generated")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.OauthProvider != "facebook" {
		t.Errorf("OauthProvider = %q, want %q", got.OauthProvider, "facebook")
	}
	if got.OAuthProviderUserID != "fb-user-7" {
		t.Errorf("OAuthProviderUserID = %q, want %q", got.OAuthProviderUserID, "fb-user-7")
	}
	if got.AccessToken != "mysql-access-new" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "mysql-access-new")
	}
	if got.RefreshToken != "mysql-refresh-new" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "mysql-refresh-new")
	}
	// MySQL uses time.Time via ParseTime
	expectedTime := ParseTime("2025-12-31T23:59:59Z")
	if !got.TokenExpiresAt.Equal(expectedTime) {
		t.Errorf("TokenExpiresAt = %v, want %v", got.TokenExpiresAt, expectedTime)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

// --- MySQL MysqlDatabase.MapUpdateUserOauthParams tests ---

func TestMysqlDatabase_MapUpdateUserOauthParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	oauthID := types.NewUserOauthID()

	input := UpdateUserOauthParams{
		AccessToken:    "mysql-updated-access",
		RefreshToken:   "mysql-updated-refresh",
		TokenExpiresAt: "2026-06-15T12:00:00Z",
		UserOauthID:    oauthID,
	}

	got := d.MapUpdateUserOauthParams(input)

	if got.AccessToken != "mysql-updated-access" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "mysql-updated-access")
	}
	if got.RefreshToken != "mysql-updated-refresh" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "mysql-updated-refresh")
	}
	expectedTime := ParseTime("2026-06-15T12:00:00Z")
	if !got.TokenExpiresAt.Equal(expectedTime) {
		t.Errorf("TokenExpiresAt = %v, want %v", got.TokenExpiresAt, expectedTime)
	}
	if got.UserOAuthID != oauthID {
		t.Errorf("UserOAuthID = %v, want %v", got.UserOAuthID, oauthID)
	}
}

// --- PostgreSQL PsqlDatabase.MapUserOauth tests ---

func TestPsqlDatabase_MapUserOauth_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	oauthID := types.NewUserOauthID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))
	tokenExpires := time.Date(2025, 11, 30, 18, 45, 0, 0, time.UTC)

	input := mdbp.UserOauth{
		UserOAuthID:         oauthID,
		UserID:              userID,
		OauthProvider:       "apple",
		OAuthProviderUserID: "apple-user-88",
		AccessToken:         "psql-access",
		RefreshToken:        "psql-refresh",
		TokenExpiresAt:      tokenExpires,
		DateCreated:         ts,
	}

	got := d.MapUserOauth(input)

	if got.UserOauthID != oauthID {
		t.Errorf("UserOauthID = %v, want %v", got.UserOauthID, oauthID)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.OauthProvider != "apple" {
		t.Errorf("OauthProvider = %q, want %q", got.OauthProvider, "apple")
	}
	if got.OauthProviderUserID != "apple-user-88" {
		t.Errorf("OauthProviderUserID = %q, want %q", got.OauthProviderUserID, "apple-user-88")
	}
	if got.AccessToken != "psql-access" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "psql-access")
	}
	if got.RefreshToken != "psql-refresh" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "psql-refresh")
	}
	// PostgreSQL TokenExpiresAt is time.Time, mapped via .String()
	if got.TokenExpiresAt != tokenExpires.String() {
		t.Errorf("TokenExpiresAt = %q, want %q", got.TokenExpiresAt, tokenExpires.String())
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

func TestPsqlDatabase_MapUserOauth_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapUserOauth(mdbp.UserOauth{})

	if got.UserOauthID != "" {
		t.Errorf("UserOauthID = %v, want zero value", got.UserOauthID)
	}
	if got.OauthProvider != "" {
		t.Errorf("OauthProvider = %q, want empty string", got.OauthProvider)
	}
	zeroTimeStr := time.Time{}.String()
	if got.TokenExpiresAt != zeroTimeStr {
		t.Errorf("TokenExpiresAt = %q, want %q", got.TokenExpiresAt, zeroTimeStr)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateUserOauthParams tests ---

func TestPsqlDatabase_MapCreateUserOauthParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := CreateUserOauthParams{
		UserID:              userID,
		OauthProvider:       "linkedin",
		OauthProviderUserID: "li-user-55",
		AccessToken:         "psql-access-new",
		RefreshToken:        "psql-refresh-new",
		TokenExpiresAt:      "2025-12-31T23:59:59Z",
		DateCreated:         ts,
	}

	got := d.MapCreateUserOauthParams(input)

	if got.UserOAuthID.IsZero() {
		t.Fatal("expected non-zero UserOAuthID to be generated")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.OauthProvider != "linkedin" {
		t.Errorf("OauthProvider = %q, want %q", got.OauthProvider, "linkedin")
	}
	if got.OAuthProviderUserID != "li-user-55" {
		t.Errorf("OAuthProviderUserID = %q, want %q", got.OAuthProviderUserID, "li-user-55")
	}
	if got.AccessToken != "psql-access-new" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "psql-access-new")
	}
	if got.RefreshToken != "psql-refresh-new" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "psql-refresh-new")
	}
	expectedTime := ParseTime("2025-12-31T23:59:59Z")
	if !got.TokenExpiresAt.Equal(expectedTime) {
		t.Errorf("TokenExpiresAt = %v, want %v", got.TokenExpiresAt, expectedTime)
	}
	if got.DateCreated != ts {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, ts)
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateUserOauthParams tests ---

func TestPsqlDatabase_MapUpdateUserOauthParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	oauthID := types.NewUserOauthID()

	input := UpdateUserOauthParams{
		AccessToken:    "psql-updated-access",
		RefreshToken:   "psql-updated-refresh",
		TokenExpiresAt: "2026-03-01T08:00:00Z",
		UserOauthID:    oauthID,
	}

	got := d.MapUpdateUserOauthParams(input)

	if got.AccessToken != "psql-updated-access" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "psql-updated-access")
	}
	if got.RefreshToken != "psql-updated-refresh" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "psql-updated-refresh")
	}
	expectedTime := ParseTime("2026-03-01T08:00:00Z")
	if !got.TokenExpiresAt.Equal(expectedTime) {
		t.Errorf("TokenExpiresAt = %v, want %v", got.TokenExpiresAt, expectedTime)
	}
	if got.UserOAuthID != oauthID {
		t.Errorf("UserOAuthID = %v, want %v", got.UserOAuthID, oauthID)
	}
}

// --- Cross-database mapper consistency ---
// This test verifies that all three database mappers produce identical UserOauth
// from equivalent input, proving the abstraction layer works correctly.
// Note: SQLite uses string for TokenExpiresAt while MySQL/PostgreSQL use time.Time,
// so we test each pair's string fields match where the source types align.

func TestCrossDatabaseMapUserOauth_Consistency(t *testing.T) {
	t.Parallel()
	oauthID := types.NewUserOauthID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	tokenExpiresStr := "2025-12-31T23:59:59Z"
	tokenExpiresTime := ParseTime(tokenExpiresStr)

	sqliteInput := mdb.UserOauth{
		UserOAuthID:         oauthID,
		UserID:              userID,
		OauthProvider:       "cross-provider",
		OAuthProviderUserID: "cross-user-1",
		AccessToken:         "cross-access",
		RefreshToken:        "cross-refresh",
		TokenExpiresAt:      tokenExpiresStr,
		DateCreated:         ts,
	}
	mysqlInput := mdbm.UserOauth{
		UserOAuthID:         oauthID,
		UserID:              userID,
		OauthProvider:       "cross-provider",
		OAuthProviderUserID: "cross-user-1",
		AccessToken:         "cross-access",
		RefreshToken:        "cross-refresh",
		TokenExpiresAt:      tokenExpiresTime,
		DateCreated:         ts,
	}
	psqlInput := mdbp.UserOauth{
		UserOAuthID:         oauthID,
		UserID:              userID,
		OauthProvider:       "cross-provider",
		OAuthProviderUserID: "cross-user-1",
		AccessToken:         "cross-access",
		RefreshToken:        "cross-refresh",
		TokenExpiresAt:      tokenExpiresTime,
		DateCreated:         ts,
	}

	sqliteResult := Database{}.MapUserOauth(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapUserOauth(mysqlInput)
	psqlResult := PsqlDatabase{}.MapUserOauth(psqlInput)

	// All fields except TokenExpiresAt should be identical across all three
	if sqliteResult.UserOauthID != mysqlResult.UserOauthID {
		t.Errorf("UserOauthID mismatch: sqlite=%v, mysql=%v", sqliteResult.UserOauthID, mysqlResult.UserOauthID)
	}
	if sqliteResult.UserID != mysqlResult.UserID {
		t.Errorf("UserID mismatch: sqlite=%v, mysql=%v", sqliteResult.UserID, mysqlResult.UserID)
	}
	if sqliteResult.OauthProvider != mysqlResult.OauthProvider {
		t.Errorf("OauthProvider mismatch: sqlite=%q, mysql=%q", sqliteResult.OauthProvider, mysqlResult.OauthProvider)
	}
	if sqliteResult.OauthProviderUserID != mysqlResult.OauthProviderUserID {
		t.Errorf("OauthProviderUserID mismatch: sqlite=%q, mysql=%q", sqliteResult.OauthProviderUserID, mysqlResult.OauthProviderUserID)
	}
	if sqliteResult.AccessToken != mysqlResult.AccessToken {
		t.Errorf("AccessToken mismatch: sqlite=%q, mysql=%q", sqliteResult.AccessToken, mysqlResult.AccessToken)
	}
	if sqliteResult.RefreshToken != mysqlResult.RefreshToken {
		t.Errorf("RefreshToken mismatch: sqlite=%q, mysql=%q", sqliteResult.RefreshToken, mysqlResult.RefreshToken)
	}
	if sqliteResult.DateCreated != mysqlResult.DateCreated {
		t.Errorf("DateCreated mismatch: sqlite=%v, mysql=%v", sqliteResult.DateCreated, mysqlResult.DateCreated)
	}

	// MySQL and PostgreSQL should produce identical results (both use time.Time)
	if mysqlResult != psqlResult {
		t.Errorf("MySQL and PostgreSQL produced different results:\n  mysql: %+v\n  psql:  %+v", mysqlResult, psqlResult)
	}
}

// --- Cross-database MapCreateUserOauthParams auto-ID generation ---

func TestCrossDatabaseMapCreateUserOauthParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := CreateUserOauthParams{
		UserID:              userID,
		OauthProvider:       "cross-create",
		OauthProviderUserID: "cross-create-user",
		AccessToken:         "cross-access",
		RefreshToken:        "cross-refresh",
		TokenExpiresAt:      "2025-12-31T23:59:59Z",
		DateCreated:         ts,
	}

	sqliteResult := Database{}.MapCreateUserOauthParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateUserOauthParams(input)
	psqlResult := PsqlDatabase{}.MapCreateUserOauthParams(input)

	if sqliteResult.UserOAuthID.IsZero() {
		t.Error("SQLite: expected non-zero generated UserOAuthID")
	}
	if mysqlResult.UserOAuthID.IsZero() {
		t.Error("MySQL: expected non-zero generated UserOAuthID")
	}
	if psqlResult.UserOAuthID.IsZero() {
		t.Error("PostgreSQL: expected non-zero generated UserOAuthID")
	}

	// Each call should generate a unique ID
	if sqliteResult.UserOAuthID == mysqlResult.UserOAuthID {
		t.Error("SQLite and MySQL generated the same UserOAuthID -- each call should be unique")
	}
	if sqliteResult.UserOAuthID == psqlResult.UserOAuthID {
		t.Error("SQLite and PostgreSQL generated the same UserOAuthID -- each call should be unique")
	}
	if mysqlResult.UserOAuthID == psqlResult.UserOAuthID {
		t.Error("MySQL and PostgreSQL generated the same UserOAuthID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewUserOauthCmd_AllAccessors(t *testing.T) {
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
	oauthUserID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	params := CreateUserOauthParams{
		UserID:              oauthUserID,
		OauthProvider:       "cmd-test-provider",
		OauthProviderUserID: "cmd-test-user",
		AccessToken:         "cmd-access",
		RefreshToken:        "cmd-refresh",
		TokenExpiresAt:      "2025-12-31T23:59:59Z",
		DateCreated:         ts,
	}

	cmd := Database{}.NewUserOauthCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_oauth" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_oauth")
	}
	p, ok := cmd.Params().(CreateUserOauthParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateUserOauthParams", cmd.Params())
	}
	if p.OauthProvider != "cmd-test-provider" {
		t.Errorf("Params().OauthProvider = %q, want %q", p.OauthProvider, "cmd-test-provider")
	}
	if p.AccessToken != "cmd-access" {
		t.Errorf("Params().AccessToken = %q, want %q", p.AccessToken, "cmd-access")
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewUserOauthCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	oauthID := types.NewUserOauthID()
	cmd := NewUserOauthCmd{}

	row := mdb.UserOauth{UserOAuthID: oauthID}
	got := cmd.GetID(row)
	if got != string(oauthID) {
		t.Errorf("GetID() = %q, want %q", got, string(oauthID))
	}
}

func TestUpdateUserOauthCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	oauthID := types.NewUserOauthID()
	params := UpdateUserOauthParams{
		AccessToken:    "updated-access",
		RefreshToken:   "updated-refresh",
		TokenExpiresAt: "2026-01-15T00:00:00Z",
		UserOauthID:    oauthID,
	}

	cmd := Database{}.UpdateUserOauthCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_oauth" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_oauth")
	}
	if cmd.GetID() != string(oauthID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(oauthID))
	}
	p, ok := cmd.Params().(UpdateUserOauthParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateUserOauthParams", cmd.Params())
	}
	if p.AccessToken != "updated-access" {
		t.Errorf("Params().AccessToken = %q, want %q", p.AccessToken, "updated-access")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteUserOauthCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	oauthID := types.NewUserOauthID()

	cmd := Database{}.DeleteUserOauthCmd(ctx, ac, oauthID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_oauth" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_oauth")
	}
	if cmd.GetID() != string(oauthID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(oauthID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewUserOauthCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-req",
		IP:        "192.168.1.1",
	}
	params := CreateUserOauthParams{
		UserID:              types.NullableUserID{ID: types.NewUserID(), Valid: true},
		OauthProvider:       "mysql-cmd-provider",
		OauthProviderUserID: "mysql-cmd-user",
		AccessToken:         "mysql-cmd-access",
		RefreshToken:        "mysql-cmd-refresh",
		TokenExpiresAt:      "2025-12-31T23:59:59Z",
		DateCreated:         ts,
	}

	cmd := MysqlDatabase{}.NewUserOauthCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_oauth" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_oauth")
	}
	p, ok := cmd.Params().(CreateUserOauthParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateUserOauthParams", cmd.Params())
	}
	if p.OauthProvider != "mysql-cmd-provider" {
		t.Errorf("Params().OauthProvider = %q, want %q", p.OauthProvider, "mysql-cmd-provider")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewUserOauthCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	oauthID := types.NewUserOauthID()
	cmd := NewUserOauthCmdMysql{}

	row := mdbm.UserOauth{UserOAuthID: oauthID}
	got := cmd.GetID(row)
	if got != string(oauthID) {
		t.Errorf("GetID() = %q, want %q", got, string(oauthID))
	}
}

func TestUpdateUserOauthCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	oauthID := types.NewUserOauthID()
	params := UpdateUserOauthParams{
		AccessToken:    "mysql-updated-access",
		RefreshToken:   "mysql-updated-refresh",
		TokenExpiresAt: "2026-02-15T12:00:00Z",
		UserOauthID:    oauthID,
	}

	cmd := MysqlDatabase{}.UpdateUserOauthCmd(ctx, ac, params)

	if cmd.TableName() != "user_oauth" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_oauth")
	}
	if cmd.GetID() != string(oauthID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(oauthID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateUserOauthParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateUserOauthParams", cmd.Params())
	}
	if p.AccessToken != "mysql-updated-access" {
		t.Errorf("Params().AccessToken = %q, want %q", p.AccessToken, "mysql-updated-access")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteUserOauthCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	oauthID := types.NewUserOauthID()

	cmd := MysqlDatabase{}.DeleteUserOauthCmd(ctx, ac, oauthID)

	if cmd.TableName() != "user_oauth" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_oauth")
	}
	if cmd.GetID() != string(oauthID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(oauthID))
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

func TestNewUserOauthCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-req",
		IP:        "172.16.0.1",
	}
	params := CreateUserOauthParams{
		UserID:              types.NullableUserID{ID: types.NewUserID(), Valid: true},
		OauthProvider:       "psql-cmd-provider",
		OauthProviderUserID: "psql-cmd-user",
		AccessToken:         "psql-cmd-access",
		RefreshToken:        "psql-cmd-refresh",
		TokenExpiresAt:      "2025-12-31T23:59:59Z",
		DateCreated:         ts,
	}

	cmd := PsqlDatabase{}.NewUserOauthCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "user_oauth" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_oauth")
	}
	p, ok := cmd.Params().(CreateUserOauthParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateUserOauthParams", cmd.Params())
	}
	if p.OauthProvider != "psql-cmd-provider" {
		t.Errorf("Params().OauthProvider = %q, want %q", p.OauthProvider, "psql-cmd-provider")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewUserOauthCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	oauthID := types.NewUserOauthID()
	cmd := NewUserOauthCmdPsql{}

	row := mdbp.UserOauth{UserOAuthID: oauthID}
	got := cmd.GetID(row)
	if got != string(oauthID) {
		t.Errorf("GetID() = %q, want %q", got, string(oauthID))
	}
}

func TestUpdateUserOauthCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	oauthID := types.NewUserOauthID()
	params := UpdateUserOauthParams{
		AccessToken:    "psql-updated-access",
		RefreshToken:   "psql-updated-refresh",
		TokenExpiresAt: "2026-04-01T06:00:00Z",
		UserOauthID:    oauthID,
	}

	cmd := PsqlDatabase{}.UpdateUserOauthCmd(ctx, ac, params)

	if cmd.TableName() != "user_oauth" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_oauth")
	}
	if cmd.GetID() != string(oauthID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(oauthID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateUserOauthParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateUserOauthParams", cmd.Params())
	}
	if p.RefreshToken != "psql-updated-refresh" {
		t.Errorf("Params().RefreshToken = %q, want %q", p.RefreshToken, "psql-updated-refresh")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteUserOauthCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	oauthID := types.NewUserOauthID()

	cmd := PsqlDatabase{}.DeleteUserOauthCmd(ctx, ac, oauthID)

	if cmd.TableName() != "user_oauth" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "user_oauth")
	}
	if cmd.GetID() != string(oauthID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(oauthID))
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
// name and the correct recorder type.

func TestAuditedUserOauthCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}

	createParams := CreateUserOauthParams{}
	updateParams := UpdateUserOauthParams{UserOauthID: types.NewUserOauthID()}
	oauthID := types.NewUserOauthID()

	// SQLite
	sqliteCreate := Database{}.NewUserOauthCmd(ctx, ac, createParams)
	sqliteUpdate := Database{}.UpdateUserOauthCmd(ctx, ac, updateParams)
	sqliteDelete := Database{}.DeleteUserOauthCmd(ctx, ac, oauthID)

	// MySQL
	mysqlCreate := MysqlDatabase{}.NewUserOauthCmd(ctx, ac, createParams)
	mysqlUpdate := MysqlDatabase{}.UpdateUserOauthCmd(ctx, ac, updateParams)
	mysqlDelete := MysqlDatabase{}.DeleteUserOauthCmd(ctx, ac, oauthID)

	// PostgreSQL
	psqlCreate := PsqlDatabase{}.NewUserOauthCmd(ctx, ac, createParams)
	psqlUpdate := PsqlDatabase{}.UpdateUserOauthCmd(ctx, ac, updateParams)
	psqlDelete := PsqlDatabase{}.DeleteUserOauthCmd(ctx, ac, oauthID)

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
			if c.name != "user_oauth" {
				t.Errorf("TableName() = %q, want %q", c.name, "user_oauth")
			}
		})
	}
}

func TestAuditedUserOauthCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateUserOauthParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewUserOauthCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewUserOauthCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewUserOauthCmd(ctx, ac, createParams).Recorder()},
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

func TestAuditedUserOauthCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	oauthID := types.NewUserOauthID()

	t.Run("UpdateCmd GetID returns UserOauthID", func(t *testing.T) {
		t.Parallel()
		params := UpdateUserOauthParams{UserOauthID: oauthID}

		sqliteCmd := Database{}.UpdateUserOauthCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateUserOauthCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateUserOauthCmd(ctx, ac, params)

		wantID := string(oauthID)
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

	t.Run("DeleteCmd GetID returns oauth ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteUserOauthCmd(ctx, ac, oauthID)
		mysqlCmd := MysqlDatabase{}.DeleteUserOauthCmd(ctx, ac, oauthID)
		psqlCmd := PsqlDatabase{}.DeleteUserOauthCmd(ctx, ac, oauthID)

		wantID := string(oauthID)
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
		testOauthID := types.NewUserOauthID()

		sqliteCmd := NewUserOauthCmd{}
		mysqlCmd := NewUserOauthCmdMysql{}
		psqlCmd := NewUserOauthCmdPsql{}

		wantID := string(testOauthID)

		sqliteRow := mdb.UserOauth{UserOAuthID: testOauthID}
		mysqlRow := mdbm.UserOauth{UserOAuthID: testOauthID}
		psqlRow := mdbp.UserOauth{UserOAuthID: testOauthID}

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

// --- Edge case: UpdateUserOauthCmd with empty ID ---

func TestUpdateUserOauthCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateUserOauthParams{UserOauthID: ""}

	sqliteCmd := Database{}.UpdateUserOauthCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateUserOauthCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateUserOauthCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Edge case: DeleteUserOauthCmd with empty ID ---

func TestDeleteUserOauthCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.UserOauthID("")

	sqliteCmd := Database{}.DeleteUserOauthCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteUserOauthCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteUserOauthCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- TokenExpiresAt string-to-time conversion consistency ---
// MySQL and PostgreSQL MapCreateUserOauthParams/MapUpdateUserOauthParams
// use ParseTime to convert the string TokenExpiresAt to time.Time.
// Verify the conversion is consistent between both backends.

func TestTokenExpiresAt_ParseTime_Consistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"RFC3339 with timezone", "2025-12-31T23:59:59Z"},
		{"RFC3339 with offset", "2025-06-15T12:30:00+02:00"},
		{"midnight UTC", "2026-01-01T00:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			createInput := CreateUserOauthParams{TokenExpiresAt: tt.input}
			updateInput := UpdateUserOauthParams{TokenExpiresAt: tt.input}

			mysqlCreate := MysqlDatabase{}.MapCreateUserOauthParams(createInput)
			psqlCreate := PsqlDatabase{}.MapCreateUserOauthParams(createInput)
			mysqlUpdate := MysqlDatabase{}.MapUpdateUserOauthParams(updateInput)
			psqlUpdate := PsqlDatabase{}.MapUpdateUserOauthParams(updateInput)

			if !mysqlCreate.TokenExpiresAt.Equal(psqlCreate.TokenExpiresAt) {
				t.Errorf("MapCreateUserOauthParams: MySQL=%v, PostgreSQL=%v", mysqlCreate.TokenExpiresAt, psqlCreate.TokenExpiresAt)
			}
			if !mysqlUpdate.TokenExpiresAt.Equal(psqlUpdate.TokenExpiresAt) {
				t.Errorf("MapUpdateUserOauthParams: MySQL=%v, PostgreSQL=%v", mysqlUpdate.TokenExpiresAt, psqlUpdate.TokenExpiresAt)
			}
		})
	}
}

// --- Edge case: CreateUserOauthParams with NullableUserID Valid=false ---

func TestMapCreateUserOauthParams_NullUserID(t *testing.T) {
	t.Parallel()
	nullUserID := types.NullableUserID{Valid: false}
	input := CreateUserOauthParams{
		UserID:              nullUserID,
		OauthProvider:       "null-user-provider",
		OauthProviderUserID: "null-user-id",
		AccessToken:         "null-access",
		RefreshToken:        "null-refresh",
		TokenExpiresAt:      "2025-12-31T23:59:59Z",
	}

	sqliteGot := Database{}.MapCreateUserOauthParams(input)
	if sqliteGot.UserID != nullUserID {
		t.Errorf("SQLite UserID = %v, want %v", sqliteGot.UserID, nullUserID)
	}

	mysqlGot := MysqlDatabase{}.MapCreateUserOauthParams(input)
	if mysqlGot.UserID != nullUserID {
		t.Errorf("MySQL UserID = %v, want %v", mysqlGot.UserID, nullUserID)
	}

	psqlGot := PsqlDatabase{}.MapCreateUserOauthParams(input)
	if psqlGot.UserID != nullUserID {
		t.Errorf("PostgreSQL UserID = %v, want %v", psqlGot.UserID, nullUserID)
	}
}

// --- Edge case: empty string fields ---

func TestMapUserOauth_EmptyStringFields(t *testing.T) {
	t.Parallel()
	oauthID := types.NewUserOauthID()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	sqliteInput := mdb.UserOauth{
		UserOAuthID:         oauthID,
		OauthProvider:       "",
		OAuthProviderUserID: "",
		AccessToken:         "",
		RefreshToken:        "",
		TokenExpiresAt:      "",
		DateCreated:         ts,
	}

	got := Database{}.MapUserOauth(sqliteInput)

	if got.OauthProvider != "" {
		t.Errorf("OauthProvider = %q, want empty string", got.OauthProvider)
	}
	if got.OauthProviderUserID != "" {
		t.Errorf("OauthProviderUserID = %q, want empty string", got.OauthProviderUserID)
	}
	if got.AccessToken != "" {
		t.Errorf("AccessToken = %q, want empty string", got.AccessToken)
	}
	if got.RefreshToken != "" {
		t.Errorf("RefreshToken = %q, want empty string", got.RefreshToken)
	}
	if got.TokenExpiresAt != "" {
		t.Errorf("TokenExpiresAt = %q, want empty string", got.TokenExpiresAt)
	}
}

// --- Edge case: long string values ---

func TestMapUserOauth_LongStringValues(t *testing.T) {
	t.Parallel()
	oauthID := types.NewUserOauthID()
	ts := types.NewTimestamp(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	// Build a long token string (1024 characters)
	longToken := make([]byte, 1024)
	for i := range longToken {
		longToken[i] = 'a'
	}
	longTokenStr := string(longToken)

	input := mdb.UserOauth{
		UserOAuthID:    oauthID,
		AccessToken:    longTokenStr,
		RefreshToken:   longTokenStr,
		TokenExpiresAt: "2025-12-31T23:59:59Z",
		DateCreated:    ts,
	}

	got := Database{}.MapUserOauth(input)

	if got.AccessToken != longTokenStr {
		t.Errorf("AccessToken length = %d, want %d", len(got.AccessToken), len(longTokenStr))
	}
	if got.RefreshToken != longTokenStr {
		t.Errorf("RefreshToken length = %d, want %d", len(got.RefreshToken), len(longTokenStr))
	}
}
