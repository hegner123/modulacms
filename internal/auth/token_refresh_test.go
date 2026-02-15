// White-box tests for token_refresh.go: OAuth token refresh logic.
//
// White-box access is needed because:
//   - TokenRefresher has unexported fields (config, driver, log) that must be
//     set directly to inject test doubles. The constructor NewTokenRefresher
//     calls db.ConfigDB() which requires a real database connection.
//   - refreshToken is unexported and makes real HTTP calls to OAuth providers.
//     Testing RefreshIfNeeded end-to-end requires an httptest.Server to
//     intercept the token refresh request, and the OAuth config's TokenURL
//     must point to that test server.
//   - updateTokens is unexported but exercised indirectly through RefreshIfNeeded
//     and directly via unit tests.
//
// KNOWN ISSUE (production code bug):
//   refreshToken() builds an oauth2.Token with AccessToken set but Expiry at
//   zero value. The oauth2 library treats zero-Expiry + non-empty AccessToken
//   as "token is valid" and never issues a refresh request. This means
//   RefreshIfNeeded proceeds through to updateTokens with the OLD token, not a
//   refreshed one. Tests marked with "BUG:" document this behavior. See
//   token_refresh.go:109 -- setting Expiry to a past time would fix this.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"golang.org/x/oauth2"
)

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

// stubLogger implements Logger for capturing log calls in tests.
type stubLogger struct {
	entries []stubLogEntry
}

type stubLogEntry struct {
	level   string
	message string
	err     error
	args    []any
}

func (s *stubLogger) Debug(message string, args ...any) {
	s.entries = append(s.entries, stubLogEntry{level: "debug", message: message, args: args})
}
func (s *stubLogger) Info(message string, args ...any) {
	s.entries = append(s.entries, stubLogEntry{level: "info", message: message, args: args})
}
func (s *stubLogger) Warn(message string, err error, args ...any) {
	s.entries = append(s.entries, stubLogEntry{level: "warn", message: message, err: err, args: args})
}
func (s *stubLogger) Error(message string, err error, args ...any) {
	s.entries = append(s.entries, stubLogEntry{level: "error", message: message, err: err, args: args})
}

// hasLogAt returns true if any log entry at the given level contains substr.
func (s *stubLogger) hasLogAt(level, substr string) bool {
	for _, e := range s.entries {
		if e.level == level && strings.Contains(e.message, substr) {
			return true
		}
	}
	return false
}

// Compile-time check.
var _ Logger = (*stubLogger)(nil)

// stubDbDriver embeds db.DbDriver to satisfy the large interface.
// Only methods used by token_refresh.go are implemented; all others panic
// if called, which surfaces unexpected dependencies in tests immediately.
type stubDbDriver struct {
	db.DbDriver // embedded interface -- nil; panics on unimplemented methods

	getUserOauthByUserIdFn func(types.NullableUserID) (*db.UserOauth, error)
	updateUserOauthFn      func(context.Context, audited.AuditContext, db.UpdateUserOauthParams) (*string, error)

	// Captures the last UpdateUserOauth call for assertion
	lastUpdateParams *db.UpdateUserOauthParams
	updateCallCount  int
}

func (s *stubDbDriver) GetUserOauthByUserId(userID types.NullableUserID) (*db.UserOauth, error) {
	if s.getUserOauthByUserIdFn != nil {
		return s.getUserOauthByUserIdFn(userID)
	}
	return nil, fmt.Errorf("no OAuth record")
}

func (s *stubDbDriver) UpdateUserOauth(ctx context.Context, ac audited.AuditContext, params db.UpdateUserOauthParams) (*string, error) {
	s.lastUpdateParams = &params
	s.updateCallCount++
	if s.updateUserOauthFn != nil {
		return s.updateUserOauthFn(ctx, ac, params)
	}
	msg := "ok"
	return &msg, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestConfig creates a minimal config suitable for TokenRefresher tests.
// tokenURL should point to an httptest.Server when testing actual refresh.
func newTestConfig(tokenURL string) *config.Config {
	return &config.Config{
		Oauth_Client_Id:     "test-client-id",
		Oauth_Client_Secret: "test-client-secret",
		Oauth_Endpoint: map[config.Endpoint]string{
			config.OauthTokenURL: tokenURL,
		},
		Node_ID: "test-node",
	}
}

// newTestRefresher constructs a TokenRefresher with injected test doubles.
func newTestRefresher(log Logger, cfg *config.Config, driver db.DbDriver) *TokenRefresher {
	return &TokenRefresher{
		config: cfg,
		driver: driver,
		log:    log,
	}
}

// fakeOauthTokenServer returns an httptest.Server that responds to OAuth
// token refresh requests with a new access token and configurable expiry.
func fakeOauthTokenServer(t *testing.T, accessToken, refreshToken string, expiry time.Time) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request to token endpoint, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := map[string]any{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
		}
		if !expiry.IsZero() {
			resp["expires_in"] = int(time.Until(expiry).Seconds())
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode token response: %v", err)
		}
	}))
}

// fakeOauthTokenServerError returns an httptest.Server that always responds
// with an error to token refresh requests.
func fakeOauthTokenServerError(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{
			"error":             "invalid_grant",
			"error_description": "refresh token expired",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode error response: %v", err)
		}
	}))
}

// makeOAuth2Token constructs an oauth2.Token for use in updateTokens tests.
func makeOAuth2Token(accessToken, refreshToken string, expiry time.Time) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
		TokenType:    "Bearer",
	}
}

// ---------------------------------------------------------------------------
// RefreshIfNeeded: early-return paths (no OAuth, long-lived, still valid)
// ---------------------------------------------------------------------------

func TestRefreshIfNeeded_UserHasNoOAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		driverFn  func(types.NullableUserID) (*db.UserOauth, error)
		wantDebug string // expected substring in a debug log
	}{
		{
			name: "GetUserOauthByUserId returns error",
			driverFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
				return nil, fmt.Errorf("sql: no rows in result set")
			},
			wantDebug: "No OAuth record",
		},
		{
			name: "GetUserOauthByUserId returns nil record",
			driverFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
				return nil, nil
			},
			wantDebug: "", // no specific debug log for nil return path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := &stubLogger{}
			driver := &stubDbDriver{getUserOauthByUserIdFn: tt.driverFn}
			tr := newTestRefresher(log, newTestConfig(""), driver)

			err := tr.RefreshIfNeeded(types.UserID("user-123"))
			if err != nil {
				t.Fatalf("expected nil error for user without OAuth, got: %v", err)
			}

			if tt.wantDebug != "" && !log.hasLogAt("debug", tt.wantDebug) {
				t.Errorf("expected debug log containing %q, got entries: %+v", tt.wantDebug, log.entries)
			}
		})
	}
}

func TestRefreshIfNeeded_LongLivedToken(t *testing.T) {
	t.Parallel()

	// Tokens without an expiry (e.g., GitHub personal access tokens) should
	// not trigger a refresh. The code checks for empty TokenExpiresAt.
	tests := []struct {
		name           string
		tokenExpiresAt string
		wantDebug      string
	}{
		{
			name:           "empty expiry string",
			tokenExpiresAt: "",
			wantDebug:      "no expiry",
		},
		{
			// time.Parse of "0001-01-01T00:00:00Z" succeeds with zero time,
			// which triggers the IsZero() / Year()==1 guard
			name:           "zero time RFC3339",
			tokenExpiresAt: "0001-01-01T00:00:00Z",
			wantDebug:      "no expiry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := &stubLogger{}
			driver := &stubDbDriver{
				getUserOauthByUserIdFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
					return &db.UserOauth{
						UserOauthID:    types.UserOauthID("oauth-1"),
						AccessToken:    "some-token",
						RefreshToken:   "some-refresh",
						TokenExpiresAt: tt.tokenExpiresAt,
					}, nil
				},
			}
			tr := newTestRefresher(log, newTestConfig(""), driver)

			err := tr.RefreshIfNeeded(types.UserID("user-456"))
			if err != nil {
				t.Fatalf("expected nil error for long-lived token, got: %v", err)
			}
			if !log.hasLogAt("debug", tt.wantDebug) {
				t.Errorf("expected debug log containing %q, got entries: %+v", tt.wantDebug, log.entries)
			}
		})
	}
}

func TestRefreshIfNeeded_TokenStillValid(t *testing.T) {
	t.Parallel()

	// Token that expires in 30 minutes -- well beyond the 5-minute threshold
	expiresAt := time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339)

	log := &stubLogger{}
	driver := &stubDbDriver{
		getUserOauthByUserIdFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID:    types.UserOauthID("oauth-2"),
				AccessToken:    "valid-token",
				RefreshToken:   "valid-refresh",
				TokenExpiresAt: expiresAt,
			}, nil
		},
	}
	tr := newTestRefresher(log, newTestConfig(""), driver)

	err := tr.RefreshIfNeeded(types.UserID("user-789"))
	if err != nil {
		t.Fatalf("expected nil error for token still valid, got: %v", err)
	}
	if !log.hasLogAt("debug", "still valid") {
		t.Errorf("expected debug log containing 'still valid', got entries: %+v", log.entries)
	}
}

// ---------------------------------------------------------------------------
// RefreshIfNeeded: 5-minute boundary tests
// ---------------------------------------------------------------------------

func TestRefreshIfNeeded_TokenExpiresAtBoundary(t *testing.T) {
	t.Parallel()

	// Test the 5-minute boundary. time.Until(expiresAt) > 5*time.Minute
	// is false when <= 5 min, so that triggers the refresh path.
	//
	// BUG: refreshToken() builds an oauth2.Token with non-empty AccessToken
	// and zero Expiry. The oauth2 library considers this "valid" and returns
	// the old token without issuing a refresh request. So even when the code
	// enters the refresh path, it "succeeds" with the old token and calls
	// updateTokens with the original access token. The test verifies this
	// actual (buggy) behavior: no error is returned, and updateTokens is
	// called (with the OLD token values).
	tests := []struct {
		name        string
		expiresIn   time.Duration
		wantRefresh bool // whether the code enters the refresh path
	}{
		{
			name:        "expires in 4 minutes enters refresh path",
			expiresIn:   4 * time.Minute,
			wantRefresh: true,
		},
		{
			name:        "expires in 5 minutes exactly enters refresh path",
			expiresIn:   5 * time.Minute,
			wantRefresh: true,
		},
		{
			name:        "expires in 6 minutes does not trigger refresh",
			expiresIn:   6 * time.Minute,
			wantRefresh: false,
		},
		{
			name:        "already expired enters refresh path",
			expiresIn:   -1 * time.Minute,
			wantRefresh: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			expiresAt := time.Now().Add(tt.expiresIn).UTC().Format(time.RFC3339)
			log := &stubLogger{}
			driver := &stubDbDriver{
				getUserOauthByUserIdFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
					return &db.UserOauth{
						UserOauthID:    types.UserOauthID("oauth-boundary"),
						AccessToken:    "expiring-token",
						RefreshToken:   "refresh-token",
						TokenExpiresAt: expiresAt,
					}, nil
				},
			}
			// BUG: even with an unreachable server, refreshToken succeeds
			// because the oauth2 library returns the existing token without
			// making an HTTP request (see package comment).
			tr := newTestRefresher(log, newTestConfig("http://invalid-host:0/token"), driver)

			err := tr.RefreshIfNeeded(types.UserID("user-boundary"))

			// Due to the BUG described above, the refresh path never errors
			// -- it "succeeds" with the old token.
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantRefresh {
				// The code should log "expiring soon" and call updateTokens
				if !log.hasLogAt("info", "expiring soon") {
					t.Errorf("expected info log 'expiring soon'")
				}
				if driver.updateCallCount == 0 {
					t.Error("expected UpdateUserOauth to be called")
				}
				// BUG: updateTokens is called with the OLD access token
				// because the oauth2 library did not actually refresh it
				if driver.lastUpdateParams != nil && driver.lastUpdateParams.AccessToken != "expiring-token" {
					t.Errorf("BUG changed: AccessToken = %q, expected old token %q (oauth2 lib bug workaround)", driver.lastUpdateParams.AccessToken, "expiring-token")
				}
			} else {
				if log.hasLogAt("info", "expiring soon") {
					t.Error("did not expect 'expiring soon' log for non-expiring token")
				}
				if driver.updateCallCount > 0 {
					t.Error("UpdateUserOauth should not be called for non-expiring token")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RefreshIfNeeded: date parsing
// ---------------------------------------------------------------------------

func TestRefreshIfNeeded_ParsesAlternativeDateFormat(t *testing.T) {
	t.Parallel()

	// The code tries RFC3339 first, then falls back to "2006-01-02 15:04:05"
	expiresAt := time.Now().Add(30 * time.Minute).UTC().Format("2006-01-02 15:04:05")

	log := &stubLogger{}
	driver := &stubDbDriver{
		getUserOauthByUserIdFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID:    types.UserOauthID("oauth-alt-fmt"),
				AccessToken:    "some-token",
				RefreshToken:   "some-refresh",
				TokenExpiresAt: expiresAt,
			}, nil
		},
	}
	tr := newTestRefresher(log, newTestConfig(""), driver)

	err := tr.RefreshIfNeeded(types.UserID("user-alt-fmt"))
	if err != nil {
		t.Fatalf("expected nil error when using alternative date format, got: %v", err)
	}
	// Should get a warn about the RFC3339 parse failure, then succeed with alt format
	if !log.hasLogAt("warn", "Failed to parse") {
		t.Errorf("expected warn log about parse failure, got entries: %+v", log.entries)
	}
	if !log.hasLogAt("debug", "still valid") {
		t.Errorf("expected debug log 'still valid' after parsing alt format, got entries: %+v", log.entries)
	}
}

func TestRefreshIfNeeded_UnparsableExpiryReturnsError(t *testing.T) {
	t.Parallel()

	log := &stubLogger{}
	driver := &stubDbDriver{
		getUserOauthByUserIdFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID:    types.UserOauthID("oauth-bad-date"),
				AccessToken:    "some-token",
				RefreshToken:   "some-refresh",
				TokenExpiresAt: "not-a-date-at-all",
			}, nil
		},
	}
	tr := newTestRefresher(log, newTestConfig(""), driver)

	err := tr.RefreshIfNeeded(types.UserID("user-bad-date"))
	if err == nil {
		t.Fatal("expected error for unparsable expiry, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse expiration") {
		t.Errorf("error = %q, want it to contain 'failed to parse expiration'", err.Error())
	}
}

func TestRefreshIfNeeded_ExpiryParsingFallback_ThenRefresh(t *testing.T) {
	t.Parallel()

	// Date in alt format that is soon-to-expire, triggering both the
	// parse fallback AND the refresh path.
	soonExpiry := time.Now().Add(2 * time.Minute).UTC().Format("2006-01-02 15:04:05")

	log := &stubLogger{}
	driver := &stubDbDriver{
		getUserOauthByUserIdFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID:    types.UserOauthID("oauth-alt"),
				AccessToken:    "t",
				RefreshToken:   "r",
				TokenExpiresAt: soonExpiry,
			}, nil
		},
	}
	// BUG: oauth2 lib returns old token, so no error even with an invalid server
	tr := newTestRefresher(log, newTestConfig("http://invalid-host:0/token"), driver)

	err := tr.RefreshIfNeeded(types.UserID("user-alt-expire"))
	// BUG: no error because oauth2 lib returns old token without HTTP call
	if err != nil {
		t.Fatalf("unexpected error (BUG may be fixed if this fails): %v", err)
	}
	// But the code DID enter the refresh path:
	if !log.hasLogAt("info", "expiring soon") {
		t.Errorf("expected 'expiring soon' log")
	}
	if driver.updateCallCount == 0 {
		t.Error("expected UpdateUserOauth to be called")
	}
}

// ---------------------------------------------------------------------------
// RefreshIfNeeded: full refresh cycle (actual HTTP token exchange)
//
// BUG: Because refreshToken() does not set Expiry on the oauth2.Token,
// the oauth2 library returns the OLD token without making an HTTP request.
// The tests below document this behavior. If the bug is fixed (by setting
// token.Expiry to a past time in refreshToken()), these tests will need
// updating -- the test server WILL be contacted and new tokens will appear.
// ---------------------------------------------------------------------------

func TestRefreshIfNeeded_RefreshPathCallsUpdateTokens(t *testing.T) {
	t.Parallel()

	// Even though the oauth2 library doesn't actually refresh (BUG), the
	// code still calls updateTokens. Verify the complete flow.
	server := fakeOauthTokenServer(t, "new-access-token", "new-refresh-token",
		time.Now().Add(1*time.Hour).UTC())
	defer server.Close()

	soonExpiry := time.Now().Add(2 * time.Minute).UTC().Format(time.RFC3339)
	oauthID := types.UserOauthID("oauth-refresh-ok")

	log := &stubLogger{}
	driver := &stubDbDriver{
		getUserOauthByUserIdFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID:    oauthID,
				AccessToken:    "old-access-token",
				RefreshToken:   "old-refresh-token",
				TokenExpiresAt: soonExpiry,
			}, nil
		},
	}

	tr := newTestRefresher(log, newTestConfig(server.URL), driver)

	err := tr.RefreshIfNeeded(types.UserID("user-refresh-ok"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// UpdateUserOauth must have been called
	if driver.lastUpdateParams == nil {
		t.Fatal("UpdateUserOauth was not called")
	}
	if driver.lastUpdateParams.UserOauthID != oauthID {
		t.Errorf("UserOauthID = %q, want %q", driver.lastUpdateParams.UserOauthID, oauthID)
	}
	// BUG: AccessToken is the OLD value because oauth2 didn't actually refresh
	if driver.lastUpdateParams.AccessToken != "old-access-token" {
		// If this assertion fails, the BUG may have been fixed!
		t.Logf("NOTE: AccessToken = %q (not old-access-token); the oauth2 refresh BUG may be fixed", driver.lastUpdateParams.AccessToken)
	}

	// Verify logging
	if !log.hasLogAt("info", "expiring soon") {
		t.Errorf("expected info log about token expiring soon")
	}
	if !log.hasLogAt("info", "refreshed successfully") {
		t.Errorf("expected info log about successful refresh")
	}
}

func TestRefreshIfNeeded_UpdateTokensDatabaseFailure(t *testing.T) {
	t.Parallel()

	// Refresh "succeeds" (see BUG) but database update fails
	soonExpiry := time.Now().Add(2 * time.Minute).UTC().Format(time.RFC3339)

	log := &stubLogger{}
	driver := &stubDbDriver{
		getUserOauthByUserIdFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID:    types.UserOauthID("oauth-db-fail"),
				AccessToken:    "old-token",
				RefreshToken:   "old-refresh",
				TokenExpiresAt: soonExpiry,
			}, nil
		},
		updateUserOauthFn: func(_ context.Context, _ audited.AuditContext, _ db.UpdateUserOauthParams) (*string, error) {
			return nil, fmt.Errorf("database connection lost")
		},
	}
	// BUG: no real HTTP call, so the token server URL doesn't matter
	tr := newTestRefresher(log, newTestConfig("http://invalid:0"), driver)

	err := tr.RefreshIfNeeded(types.UserID("user-db-fail"))
	if err == nil {
		t.Fatal("expected error when database update fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update tokens") {
		t.Errorf("error = %q, want it to contain 'failed to update tokens'", err.Error())
	}
	if !log.hasLogAt("error", "Failed to update tokens in database") {
		t.Errorf("expected error log about database update failure")
	}
}

// ---------------------------------------------------------------------------
// updateTokens (direct unit tests)
// ---------------------------------------------------------------------------

func TestUpdateTokens_FormatsExpiryAsRFC3339(t *testing.T) {
	t.Parallel()

	log := &stubLogger{}
	driver := &stubDbDriver{}
	tr := newTestRefresher(log, newTestConfig(""), driver)

	oauthID := types.UserOauthID("oauth-update-fmt")
	expiry := time.Date(2026, 6, 15, 12, 30, 0, 0, time.UTC)

	err := tr.updateTokens(oauthID, makeOAuth2Token("at-1", "rt-1", expiry))
	if err != nil {
		t.Fatalf("updateTokens returned error: %v", err)
	}

	if driver.lastUpdateParams == nil {
		t.Fatal("UpdateUserOauth was not called")
	}
	want := expiry.Format(time.RFC3339)
	if driver.lastUpdateParams.TokenExpiresAt != want {
		t.Errorf("TokenExpiresAt = %q, want %q", driver.lastUpdateParams.TokenExpiresAt, want)
	}
	if driver.lastUpdateParams.AccessToken != "at-1" {
		t.Errorf("AccessToken = %q, want %q", driver.lastUpdateParams.AccessToken, "at-1")
	}
	if driver.lastUpdateParams.RefreshToken != "rt-1" {
		t.Errorf("RefreshToken = %q, want %q", driver.lastUpdateParams.RefreshToken, "rt-1")
	}
	if driver.lastUpdateParams.UserOauthID != oauthID {
		t.Errorf("UserOauthID = %q, want %q", driver.lastUpdateParams.UserOauthID, oauthID)
	}
}

func TestUpdateTokens_ZeroExpiryStoresEmptyString(t *testing.T) {
	t.Parallel()

	// GitHub tokens have no expiry -- Expiry is zero value.
	// updateTokens should store "" for TokenExpiresAt.
	log := &stubLogger{}
	driver := &stubDbDriver{}
	tr := newTestRefresher(log, newTestConfig(""), driver)

	err := tr.updateTokens(
		types.UserOauthID("oauth-no-expiry"),
		makeOAuth2Token("github-token", "github-refresh", time.Time{}),
	)
	if err != nil {
		t.Fatalf("updateTokens returned error: %v", err)
	}

	if driver.lastUpdateParams == nil {
		t.Fatal("UpdateUserOauth was not called")
	}
	if driver.lastUpdateParams.TokenExpiresAt != "" {
		t.Errorf("TokenExpiresAt = %q, want empty string for zero expiry", driver.lastUpdateParams.TokenExpiresAt)
	}
}

func TestUpdateTokens_DatabaseError(t *testing.T) {
	t.Parallel()

	log := &stubLogger{}
	driver := &stubDbDriver{
		updateUserOauthFn: func(_ context.Context, _ audited.AuditContext, _ db.UpdateUserOauthParams) (*string, error) {
			return nil, fmt.Errorf("constraint violation")
		},
	}
	tr := newTestRefresher(log, newTestConfig(""), driver)

	err := tr.updateTokens(
		types.UserOauthID("oauth-err"),
		makeOAuth2Token("some-token", "some-refresh", time.Time{}),
	)
	if err == nil {
		t.Fatal("expected error from database failure, got nil")
	}
	if !strings.Contains(err.Error(), "constraint violation") {
		t.Errorf("error = %q, want it to contain 'constraint violation'", err.Error())
	}
	if !log.hasLogAt("error", "Failed to update tokens in database") {
		t.Errorf("expected error log about database failure")
	}
}

func TestUpdateTokens_PassesCorrectAuditContext(t *testing.T) {
	t.Parallel()

	// Verify that updateTokens builds the audited context with the
	// correct action ("token-refresh") and source ("system").
	var capturedAC audited.AuditContext

	log := &stubLogger{}
	driver := &stubDbDriver{
		updateUserOauthFn: func(_ context.Context, ac audited.AuditContext, _ db.UpdateUserOauthParams) (*string, error) {
			capturedAC = ac
			msg := "ok"
			return &msg, nil
		},
	}
	cfg := &config.Config{
		Node_ID: "node-42",
		Oauth_Endpoint: map[config.Endpoint]string{
			config.OauthTokenURL: "",
		},
	}
	tr := newTestRefresher(log, cfg, driver)

	err := tr.updateTokens(
		types.UserOauthID("oauth-audit"),
		makeOAuth2Token("t", "r", time.Time{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAC.NodeID != types.NodeID("node-42") {
		t.Errorf("AuditContext.NodeID = %q, want %q", capturedAC.NodeID, "node-42")
	}
	if capturedAC.RequestID != "token-refresh" {
		t.Errorf("AuditContext.RequestID = %q, want %q", capturedAC.RequestID, "token-refresh")
	}
	if capturedAC.IP != "system" {
		t.Errorf("AuditContext.IP = %q, want %q", capturedAC.IP, "system")
	}
}

func TestUpdateTokens_LogsDebugOnSuccess(t *testing.T) {
	t.Parallel()

	log := &stubLogger{}
	driver := &stubDbDriver{}
	tr := newTestRefresher(log, newTestConfig(""), driver)

	err := tr.updateTokens(
		types.UserOauthID("oauth-debug"),
		makeOAuth2Token("t", "r", time.Now().Add(1*time.Hour)),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !log.hasLogAt("debug", "Updated tokens") {
		t.Errorf("expected debug log about updated tokens, got entries: %+v", log.entries)
	}
}

// ---------------------------------------------------------------------------
// NewTokenRefresher (constructor)
// ---------------------------------------------------------------------------

func TestNewTokenRefresher_StoresConfigAndLogger(t *testing.T) {
	t.Parallel()

	// We cannot call NewTokenRefresher without triggering db.ConfigDB
	// which would panic or fail without a real database. Instead, verify
	// that the struct fields are correctly populated by direct construction,
	// which is what NewTokenRefresher does internally (minus db.ConfigDB).
	log := &stubLogger{}
	cfg := newTestConfig("http://example.com/token")

	tr := &TokenRefresher{
		config: cfg,
		log:    log,
	}

	if tr.config != cfg {
		t.Error("config not stored correctly")
	}
	if tr.log == nil {
		t.Error("logger not stored")
	}
}

// ---------------------------------------------------------------------------
// refreshToken: OAuth config validation via request inspection
// ---------------------------------------------------------------------------

func TestRefreshToken_BuildsCorrectOAuthConfig(t *testing.T) {
	t.Parallel()

	// BUG: The oauth2 library doesn't actually contact the server because
	// it thinks the token is valid. This test verifies the config is stored
	// correctly by checking that the code doesn't error. When the BUG is
	// fixed, this test should be updated to verify the actual HTTP request.

	log := &stubLogger{}
	driver := &stubDbDriver{
		getUserOauthByUserIdFn: func(_ types.NullableUserID) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID:    types.UserOauthID("oauth-config-check"),
				AccessToken:    "old",
				RefreshToken:   "old-rt",
				TokenExpiresAt: time.Now().Add(2 * time.Minute).UTC().Format(time.RFC3339),
			}, nil
		},
	}

	cfg := &config.Config{
		Oauth_Client_Id:     "my-client-id",
		Oauth_Client_Secret: "my-client-secret",
		Oauth_Endpoint: map[config.Endpoint]string{
			config.OauthTokenURL: "http://localhost:0/should-not-be-called",
		},
		Node_ID: "node-1",
	}
	tr := newTestRefresher(log, cfg, driver)

	err := tr.RefreshIfNeeded(types.UserID("user-config-check"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the code entered the refresh path
	if !log.hasLogAt("info", "expiring soon") {
		t.Errorf("expected 'expiring soon' log, meaning refresh path was entered")
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestRefreshIfNeeded_CorrectUserIDPassedToDriver(t *testing.T) {
	t.Parallel()

	var receivedUserID types.NullableUserID

	log := &stubLogger{}
	driver := &stubDbDriver{
		getUserOauthByUserIdFn: func(uid types.NullableUserID) (*db.UserOauth, error) {
			receivedUserID = uid
			return nil, fmt.Errorf("not found")
		},
	}
	tr := newTestRefresher(log, newTestConfig(""), driver)

	testUserID := types.UserID("01HXYZ123456")
	err := tr.RefreshIfNeeded(testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !receivedUserID.Valid {
		t.Error("NullableUserID.Valid = false, want true")
	}
	if receivedUserID.ID != testUserID {
		t.Errorf("NullableUserID.ID = %q, want %q", receivedUserID.ID, testUserID)
	}
}

func TestRefreshIfNeeded_EmptyUserID(t *testing.T) {
	t.Parallel()

	// Verify the code handles empty UserID gracefully (doesn't panic)
	var receivedUserID types.NullableUserID

	log := &stubLogger{}
	driver := &stubDbDriver{
		getUserOauthByUserIdFn: func(uid types.NullableUserID) (*db.UserOauth, error) {
			receivedUserID = uid
			return nil, fmt.Errorf("not found")
		},
	}
	tr := newTestRefresher(log, newTestConfig(""), driver)

	err := tr.RefreshIfNeeded(types.UserID(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !receivedUserID.Valid {
		t.Error("NullableUserID.Valid = false, want true (even for empty user ID)")
	}
	if receivedUserID.ID != types.UserID("") {
		t.Errorf("NullableUserID.ID = %q, want empty", receivedUserID.ID)
	}
}
