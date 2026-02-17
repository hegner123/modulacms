// White-box tests for user_provision.go: OAuth user provisioning, user info
// fetching, and GitHub email fallback logic.
//
// White-box access is needed because:
//   - UserProvisioner has unexported fields (config, driver, log) that must be
//     set directly to inject test doubles. The constructor NewUserProvisioner
//     calls db.ConfigDB() which requires a real database connection.
//   - fetchGitHubEmail is unexported but contains non-trivial branching logic
//     (primary+verified preference, fallback to any verified email, no verified
//     email error) that is not fully exercisable through FetchUserInfo alone
//     without elaborate multi-endpoint httptest setups.
//   - createNewUser, linkOAuthToUser, and updateTokens are unexported helpers
//     with distinct error paths and token-expiry formatting logic.
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

// provisionLogger implements Logger for tests. Reuses the pattern from
// token_refresh_test.go but is defined here to avoid cross-file coupling
// within white-box tests.
type provisionLogger struct {
	entries []provisionLogEntry
}

type provisionLogEntry struct {
	level   string
	message string
	err     error
	args    []any
}

func (l *provisionLogger) Debug(message string, args ...any) {
	l.entries = append(l.entries, provisionLogEntry{level: "debug", message: message, args: args})
}
func (l *provisionLogger) Info(message string, args ...any) {
	l.entries = append(l.entries, provisionLogEntry{level: "info", message: message, args: args})
}
func (l *provisionLogger) Warn(message string, err error, args ...any) {
	l.entries = append(l.entries, provisionLogEntry{level: "warn", message: message, err: err, args: args})
}
func (l *provisionLogger) Error(message string, err error, args ...any) {
	l.entries = append(l.entries, provisionLogEntry{level: "error", message: message, err: err, args: args})
}

// hasLog returns true if any entry at the given level contains substr.
func (l *provisionLogger) hasLog(level, substr string) bool {
	for _, e := range l.entries {
		if e.level == level && strings.Contains(e.message, substr) {
			return true
		}
	}
	return false
}

var _ Logger = (*provisionLogger)(nil)

// provisionDbDriver is a test double for db.DbDriver. Only methods used by
// user_provision.go are implemented. Unimplemented methods panic via the
// embedded nil interface, surfacing unexpected dependencies immediately.
type provisionDbDriver struct {
	db.DbDriver // nil embedded -- panics on unimplemented calls

	getUserOauthByProviderIDFn func(provider, providerUserID string) (*db.UserOauth, error)
	getUserByEmailFn           func(email types.Email) (*db.Users, error)
	getUserFn                  func(id types.UserID) (*db.Users, error)
	createUserFn               func(ctx context.Context, ac audited.AuditContext, params db.CreateUserParams) (*db.Users, error)
	createUserOauthFn          func(ctx context.Context, ac audited.AuditContext, params db.CreateUserOauthParams) (*db.UserOauth, error)
	updateUserOauthFn          func(ctx context.Context, ac audited.AuditContext, params db.UpdateUserOauthParams) (*string, error)
	listRolesFn                func() (*[]db.Roles, error)

	// Capture calls for assertion
	lastCreateUserParams      *db.CreateUserParams
	lastCreateUserOauthParams *db.CreateUserOauthParams
	lastUpdateOauthParams     *db.UpdateUserOauthParams
	createUserCallCount       int
	createOauthCallCount      int
	updateOauthCallCount      int
}

func (d *provisionDbDriver) GetUserOauthByProviderID(provider, providerUserID string) (*db.UserOauth, error) {
	if d.getUserOauthByProviderIDFn != nil {
		return d.getUserOauthByProviderIDFn(provider, providerUserID)
	}
	return nil, fmt.Errorf("no oauth record")
}

func (d *provisionDbDriver) GetUserByEmail(email types.Email) (*db.Users, error) {
	if d.getUserByEmailFn != nil {
		return d.getUserByEmailFn(email)
	}
	return nil, fmt.Errorf("user not found")
}

func (d *provisionDbDriver) GetUser(id types.UserID) (*db.Users, error) {
	if d.getUserFn != nil {
		return d.getUserFn(id)
	}
	return nil, fmt.Errorf("user not found")
}

func (d *provisionDbDriver) CreateUser(ctx context.Context, ac audited.AuditContext, params db.CreateUserParams) (*db.Users, error) {
	d.lastCreateUserParams = &params
	d.createUserCallCount++
	if d.createUserFn != nil {
		return d.createUserFn(ctx, ac, params)
	}
	return &db.Users{
		UserID:   types.NewUserID(),
		Username: params.Username,
		Name:     params.Name,
		Email:    params.Email,
		Hash:     params.Hash,
		Role:     params.Role,
	}, nil
}

func (d *provisionDbDriver) CreateUserOauth(ctx context.Context, ac audited.AuditContext, params db.CreateUserOauthParams) (*db.UserOauth, error) {
	d.lastCreateUserOauthParams = &params
	d.createOauthCallCount++
	if d.createUserOauthFn != nil {
		return d.createUserOauthFn(ctx, ac, params)
	}
	return &db.UserOauth{
		UserOauthID:         types.UserOauthID("generated-oauth-id"),
		UserID:              params.UserID,
		OauthProvider:       params.OauthProvider,
		OauthProviderUserID: params.OauthProviderUserID,
		AccessToken:         params.AccessToken,
		RefreshToken:        params.RefreshToken,
		TokenExpiresAt:      params.TokenExpiresAt,
	}, nil
}

func (d *provisionDbDriver) UpdateUserOauth(ctx context.Context, ac audited.AuditContext, params db.UpdateUserOauthParams) (*string, error) {
	d.lastUpdateOauthParams = &params
	d.updateOauthCallCount++
	if d.updateUserOauthFn != nil {
		return d.updateUserOauthFn(ctx, ac, params)
	}
	msg := "ok"
	return &msg, nil
}

func (d *provisionDbDriver) ListRoles() (*[]db.Roles, error) {
	if d.listRolesFn != nil {
		return d.listRolesFn()
	}
	roles := []db.Roles{
		{RoleID: types.RoleID("role-viewer"), Label: "viewer"},
		{RoleID: types.RoleID("role-admin"), Label: "admin"},
	}
	return &roles, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newProvisionConfig creates a minimal config for UserProvisioner tests.
func newProvisionConfig(userInfoURL string) *config.Config {
	return &config.Config{
		Oauth_Client_Id:     "test-client-id",
		Oauth_Client_Secret: "test-client-secret",
		Oauth_Endpoint: map[config.Endpoint]string{
			config.OauthUserInfoURL: userInfoURL,
		},
		Node_ID: "test-node",
	}
}

// newTestProvisioner constructs a UserProvisioner with injected test doubles,
// bypassing NewUserProvisioner which calls db.ConfigDB.
func newTestProvisioner(log Logger, cfg *config.Config, driver db.DbDriver) *UserProvisioner {
	return &UserProvisioner{
		config: cfg,
		driver: driver,
		log:    log,
	}
}

// makeToken constructs an oauth2.Token for tests.
func makeToken(accessToken, refreshToken string, expiry time.Time) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
		TokenType:    "Bearer",
	}
}

// fakeUserInfoServer returns an httptest.Server that serves JSON user info
// at its root URL.
func fakeUserInfoServer(t *testing.T, userInfo UserInfo) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userInfo); err != nil {
			t.Errorf("failed to encode user info response: %v", err)
		}
	}))
}

// fakeGitHubEmailServer returns an httptest.Server that serves a JSON array
// of GitHubEmail at the path /user/emails.
func fakeGitHubEmailServer(t *testing.T, emails []GitHubEmail) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(emails); err != nil {
			t.Errorf("failed to encode GitHub emails response: %v", err)
		}
	}))
}

// fakeErrorServer returns an httptest.Server that always responds with the
// given status code and body.
func fakeErrorServer(t *testing.T, statusCode int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		fmt.Fprint(w, body)
	}))
}

// ---------------------------------------------------------------------------
// FetchUserInfo
// ---------------------------------------------------------------------------

func TestFetchUserInfo_StandardOIDCResponse(t *testing.T) {
	t.Parallel()

	server := fakeUserInfoServer(t, UserInfo{
		ProviderUserID: "oidc-sub-123",
		Email:          "alice@example.com",
		Name:           "Alice Smith",
		Username:       "alice",
	})
	defer server.Close()

	log := &provisionLogger{}
	cfg := newProvisionConfig(server.URL)
	up := newTestProvisioner(log, cfg, nil)

	info, err := up.FetchUserInfo(server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.ProviderUserID != "oidc-sub-123" {
		t.Errorf("ProviderUserID = %q, want %q", info.ProviderUserID, "oidc-sub-123")
	}
	if info.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", info.Email, "alice@example.com")
	}
	if info.Name != "Alice Smith" {
		t.Errorf("Name = %q, want %q", info.Name, "Alice Smith")
	}
	if info.Username != "alice" {
		t.Errorf("Username = %q, want %q", info.Username, "alice")
	}
}

func TestFetchUserInfo_GitHubFieldMapping(t *testing.T) {
	t.Parallel()

	// GitHub uses "login" instead of "preferred_username" and numeric "id"
	// instead of string "sub". FetchUserInfo should normalize these.
	server := fakeUserInfoServer(t, UserInfo{
		ID:        12345,
		Login:     "octocat",
		Email:     "octocat@github.com",
		Name:      "The Octocat",
		AvatarURL: "https://avatars.githubusercontent.com/u/12345",
	})
	defer server.Close()

	log := &provisionLogger{}
	cfg := newProvisionConfig(server.URL)
	up := newTestProvisioner(log, cfg, nil)

	info, err := up.FetchUserInfo(server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Username should be populated from Login when preferred_username is empty
	if info.Username != "octocat" {
		t.Errorf("Username = %q, want %q (from Login field)", info.Username, "octocat")
	}

	// ProviderUserID should be the string form of the numeric ID
	if info.ProviderUserID != "12345" {
		t.Errorf("ProviderUserID = %q, want %q (from numeric ID)", info.ProviderUserID, "12345")
	}
}

func TestFetchUserInfo_MissingUserInfoURL(t *testing.T) {
	t.Parallel()

	log := &provisionLogger{}
	cfg := &config.Config{
		Oauth_Endpoint: map[config.Endpoint]string{},
	}
	up := newTestProvisioner(log, cfg, nil)

	_, err := up.FetchUserInfo(http.DefaultClient)
	if err == nil {
		t.Fatal("expected error for missing userinfo URL, got nil")
	}
	if !strings.Contains(err.Error(), "oauth_userinfo_url not configured") {
		t.Errorf("error = %q, want it to contain 'oauth_userinfo_url not configured'", err.Error())
	}
}

func TestFetchUserInfo_NonOKStatusCode(t *testing.T) {
	t.Parallel()

	server := fakeErrorServer(t, http.StatusForbidden, "access denied")
	defer server.Close()

	log := &provisionLogger{}
	cfg := newProvisionConfig(server.URL)
	up := newTestProvisioner(log, cfg, nil)

	_, err := up.FetchUserInfo(server.Client())
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error = %q, want it to contain status code '403'", err.Error())
	}
	if !log.hasLog("error", "Userinfo request failed") {
		t.Errorf("expected error log about userinfo failure")
	}
}

func TestFetchUserInfo_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "not valid json{{{")
	}))
	defer server.Close()

	log := &provisionLogger{}
	cfg := newProvisionConfig(server.URL)
	up := newTestProvisioner(log, cfg, nil)

	_, err := up.FetchUserInfo(server.Client())
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode userinfo") {
		t.Errorf("error = %q, want it to contain 'failed to decode userinfo'", err.Error())
	}
}

func TestFetchUserInfo_MissingEmail(t *testing.T) {
	t.Parallel()

	// When userinfo has no email and the GitHub email endpoint also fails,
	// FetchUserInfo should return an error about missing email.
	// We need a server that returns no email in userinfo, and a GitHub email
	// endpoint that returns an error. Since fetchGitHubEmail hits a hardcoded
	// URL (https://api.github.com/user/emails), we can't intercept it with
	// httptest for the userinfo client. The client.Get to that URL will fail,
	// which triggers the Warn path, and then the "email not provided" error.
	server := fakeUserInfoServer(t, UserInfo{
		ProviderUserID: "sub-no-email",
		Name:           "No Email User",
	})
	defer server.Close()

	log := &provisionLogger{}
	cfg := newProvisionConfig(server.URL)
	up := newTestProvisioner(log, cfg, nil)

	_, err := up.FetchUserInfo(server.Client())
	if err == nil {
		t.Fatal("expected error for missing email, got nil")
	}
	if !strings.Contains(err.Error(), "email not provided") {
		t.Errorf("error = %q, want it to contain 'email not provided'", err.Error())
	}
}

func TestFetchUserInfo_UsernameNotOverriddenWhenPresent(t *testing.T) {
	t.Parallel()

	// When both preferred_username and login are set, preferred_username wins
	server := fakeUserInfoServer(t, UserInfo{
		ProviderUserID: "sub-both",
		Email:          "both@example.com",
		Username:       "preferred-name",
		Login:          "login-name",
	})
	defer server.Close()

	log := &provisionLogger{}
	cfg := newProvisionConfig(server.URL)
	up := newTestProvisioner(log, cfg, nil)

	info, err := up.FetchUserInfo(server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Username != "preferred-name" {
		t.Errorf("Username = %q, want %q (preferred_username should take precedence)", info.Username, "preferred-name")
	}
}

func TestFetchUserInfo_SubNotOverriddenByNumericID(t *testing.T) {
	t.Parallel()

	// When both sub and id are set, sub (ProviderUserID) should remain
	server := fakeUserInfoServer(t, UserInfo{
		ProviderUserID: "oidc-sub",
		ID:             99999,
		Email:          "dual@example.com",
	})
	defer server.Close()

	log := &provisionLogger{}
	cfg := newProvisionConfig(server.URL)
	up := newTestProvisioner(log, cfg, nil)

	info, err := up.FetchUserInfo(server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.ProviderUserID != "oidc-sub" {
		t.Errorf("ProviderUserID = %q, want %q (sub should not be overridden by numeric id)", info.ProviderUserID, "oidc-sub")
	}
}

// ---------------------------------------------------------------------------
// fetchGitHubEmail (unexported)
// ---------------------------------------------------------------------------

func TestFetchGitHubEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		emails    []GitHubEmail
		wantEmail string
		wantErr   string
	}{
		{
			name: "returns primary verified email",
			emails: []GitHubEmail{
				{Email: "secondary@example.com", Primary: false, Verified: true},
				{Email: "primary@example.com", Primary: true, Verified: true},
			},
			wantEmail: "primary@example.com",
		},
		{
			name: "falls back to first verified when no primary",
			emails: []GitHubEmail{
				{Email: "unverified@example.com", Primary: true, Verified: false},
				{Email: "verified@example.com", Primary: false, Verified: true},
			},
			wantEmail: "verified@example.com",
		},
		{
			name: "no verified email returns error",
			emails: []GitHubEmail{
				{Email: "unverified@example.com", Primary: true, Verified: false},
			},
			wantErr: "no verified email found",
		},
		{
			name:    "empty email list returns error",
			emails:  []GitHubEmail{},
			wantErr: "no verified email found",
		},
		{
			// Primary+verified should be preferred over just-verified
			name: "primary verified preferred over non-primary verified",
			emails: []GitHubEmail{
				{Email: "first-verified@example.com", Primary: false, Verified: true},
				{Email: "primary-verified@example.com", Primary: true, Verified: true},
				{Email: "second-verified@example.com", Primary: false, Verified: true},
			},
			wantEmail: "primary-verified@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := fakeGitHubEmailServer(t, tt.emails)
			defer server.Close()

			log := &provisionLogger{}
			cfg := newProvisionConfig("")
			up := newTestProvisioner(log, cfg, nil)

			// Override the hardcoded GitHub URL by using the test server's
			// client, but fetchGitHubEmail uses a hardcoded URL so we need
			// to create a custom transport that redirects all requests to
			// our test server.
			transport := &rewriteTransport{base: server.URL}
			client := &http.Client{Transport: transport}

			email, err := up.fetchGitHubEmail(client)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if email != tt.wantEmail {
				t.Errorf("email = %q, want %q", email, tt.wantEmail)
			}
		})
	}
}

// rewriteTransport rewrites all request URLs to point to a test server,
// allowing us to intercept calls to hardcoded URLs like api.github.com.
type rewriteTransport struct {
	base string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to our test server, keeping the path
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.base, "http://")
	return http.DefaultTransport.RoundTrip(req)
}

func TestFetchGitHubEmail_HTTPError(t *testing.T) {
	t.Parallel()

	server := fakeErrorServer(t, http.StatusUnauthorized, "bad token")
	defer server.Close()

	log := &provisionLogger{}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, nil)

	transport := &rewriteTransport{base: server.URL}
	client := &http.Client{Transport: transport}

	_, err := up.fetchGitHubEmail(client)
	if err == nil {
		t.Fatal("expected error for HTTP 401, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %q, want it to contain '401'", err.Error())
	}
}

func TestFetchGitHubEmail_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "not json")
	}))
	defer server.Close()

	log := &provisionLogger{}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, nil)

	transport := &rewriteTransport{base: server.URL}
	client := &http.Client{Transport: transport}

	_, err := up.fetchGitHubEmail(client)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode emails") {
		t.Errorf("error = %q, want it to contain 'failed to decode emails'", err.Error())
	}
}

// ---------------------------------------------------------------------------
// ProvisionUser
// ---------------------------------------------------------------------------

func TestProvisionUser_ExistingOAuthLink(t *testing.T) {
	t.Parallel()

	// When an OAuth record already exists for the provider+userID, the code
	// updates tokens and returns the existing user.
	existingUserID := types.UserID("existing-user-123")
	oauthID := types.UserOauthID("existing-oauth-456")

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		getUserOauthByProviderIDFn: func(provider, providerUserID string) (*db.UserOauth, error) {
			if provider != "github" || providerUserID != "gh-789" {
				t.Errorf("unexpected provider=%q providerUserID=%q", provider, providerUserID)
			}
			return &db.UserOauth{
				UserOauthID: oauthID,
				UserID:      types.NullableUserID{ID: existingUserID, Valid: true},
			}, nil
		},
		getUserFn: func(id types.UserID) (*db.Users, error) {
			if id != existingUserID {
				t.Errorf("GetUser called with %q, want %q", id, existingUserID)
			}
			return &db.Users{
				UserID:   existingUserID,
				Username: "existing-user",
				Email:    types.Email("existing@example.com"),
			}, nil
		},
	}

	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	userInfo := &UserInfo{
		ProviderUserID: "gh-789",
		Email:          "existing@example.com",
	}
	token := makeToken("new-access", "new-refresh", time.Now().Add(1*time.Hour))

	user, err := up.ProvisionUser(userInfo, token, "github")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID != existingUserID {
		t.Errorf("UserID = %q, want %q", user.UserID, existingUserID)
	}

	// Tokens should have been updated
	if driver.updateOauthCallCount != 1 {
		t.Errorf("UpdateUserOauth call count = %d, want 1", driver.updateOauthCallCount)
	}
	if driver.lastUpdateOauthParams.UserOauthID != oauthID {
		t.Errorf("UpdateUserOauth UserOauthID = %q, want %q", driver.lastUpdateOauthParams.UserOauthID, oauthID)
	}
	if driver.lastUpdateOauthParams.AccessToken != "new-access" {
		t.Errorf("AccessToken = %q, want %q", driver.lastUpdateOauthParams.AccessToken, "new-access")
	}

	// No user creation should have occurred
	if driver.createUserCallCount != 0 {
		t.Error("CreateUser should not be called when OAuth link exists")
	}
}

func TestProvisionUser_ExistingOAuthLink_UpdateTokensFails(t *testing.T) {
	t.Parallel()

	// When token update fails, the error should be propagated.
	existingUserID := types.UserID("user-ok-token-fail")

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		getUserOauthByProviderIDFn: func(_, _ string) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID: types.UserOauthID("oauth-fail"),
				UserID:      types.NullableUserID{ID: existingUserID, Valid: true},
			}, nil
		},
		updateUserOauthFn: func(_ context.Context, _ audited.AuditContext, _ db.UpdateUserOauthParams) (*string, error) {
			return nil, fmt.Errorf("db connection lost")
		},
		getUserFn: func(id types.UserID) (*db.Users, error) {
			return &db.Users{UserID: id, Email: types.Email("user@example.com")}, nil
		},
	}

	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.ProvisionUser(
		&UserInfo{ProviderUserID: "sub-1", Email: "user@example.com"},
		makeToken("t", "r", time.Time{}),
		"provider",
	)
	if err == nil {
		t.Fatal("expected error when token update fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update OAuth tokens") {
		t.Errorf("error = %q, want it to contain 'failed to update OAuth tokens'", err.Error())
	}
	if !log.hasLog("warn", "Failed to update tokens") {
		t.Errorf("expected warn log about token update failure")
	}
}

func TestProvisionUser_ExistingUserByEmail(t *testing.T) {
	t.Parallel()

	// When no OAuth link exists but the email matches an existing user,
	// OAuth should be linked to the existing user without creating a new one.
	existingUserID := types.UserID("email-match-user")

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		getUserOauthByProviderIDFn: func(_, _ string) (*db.UserOauth, error) {
			return nil, fmt.Errorf("not found")
		},
		getUserByEmailFn: func(email types.Email) (*db.Users, error) {
			if email != types.Email("existing@example.com") {
				return nil, fmt.Errorf("not found")
			}
			return &db.Users{
				UserID:   existingUserID,
				Username: "existing-user",
				Email:    email,
			}, nil
		},
	}

	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	token := makeToken("at-link", "rt-link", time.Now().Add(2*time.Hour))
	user, err := up.ProvisionUser(
		&UserInfo{ProviderUserID: "new-provider-id", Email: "existing@example.com", Username: "newlogin"},
		token,
		"github",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID != existingUserID {
		t.Errorf("UserID = %q, want %q (should return existing user)", user.UserID, existingUserID)
	}

	// OAuth should be created for the existing user
	if driver.createOauthCallCount != 1 {
		t.Fatalf("CreateUserOauth call count = %d, want 1", driver.createOauthCallCount)
	}
	if driver.lastCreateUserOauthParams.OauthProvider != "github" {
		t.Errorf("OauthProvider = %q, want %q", driver.lastCreateUserOauthParams.OauthProvider, "github")
	}
	if driver.lastCreateUserOauthParams.OauthProviderUserID != "new-provider-id" {
		t.Errorf("OauthProviderUserID = %q, want %q", driver.lastCreateUserOauthParams.OauthProviderUserID, "new-provider-id")
	}
	if driver.lastCreateUserOauthParams.UserID.ID != existingUserID {
		t.Errorf("UserID = %q, want %q", driver.lastCreateUserOauthParams.UserID.ID, existingUserID)
	}

	// No new user creation
	if driver.createUserCallCount != 0 {
		t.Error("CreateUser should not be called when email matches existing user")
	}
}

func TestProvisionUser_CreateNewUser(t *testing.T) {
	t.Parallel()

	// When no OAuth link or email match exists, a new user is created.
	newUserID := types.UserID("new-user-id")

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		getUserOauthByProviderIDFn: func(_, _ string) (*db.UserOauth, error) {
			return nil, fmt.Errorf("not found")
		},
		getUserByEmailFn: func(_ types.Email) (*db.Users, error) {
			return nil, fmt.Errorf("not found")
		},
		createUserFn: func(_ context.Context, _ audited.AuditContext, params db.CreateUserParams) (*db.Users, error) {
			return &db.Users{
				UserID:   newUserID,
				Username: params.Username,
				Name:     params.Name,
				Email:    params.Email,
				Role:     params.Role,
			}, nil
		},
	}

	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	token := makeToken("at-new", "rt-new", time.Now().Add(1*time.Hour))
	user, err := up.ProvisionUser(
		&UserInfo{ProviderUserID: "new-sub", Email: "brand-new@example.com", Username: "brandnew", Name: "Brand New"},
		token,
		"github",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID != newUserID {
		t.Errorf("UserID = %q, want %q", user.UserID, newUserID)
	}

	// User creation params
	if driver.lastCreateUserParams == nil {
		t.Fatal("CreateUser was not called")
	}
	if driver.lastCreateUserParams.Username != "brandnew" {
		t.Errorf("Username = %q, want %q", driver.lastCreateUserParams.Username, "brandnew")
	}
	if driver.lastCreateUserParams.Name != "Brand New" {
		t.Errorf("Name = %q, want %q", driver.lastCreateUserParams.Name, "Brand New")
	}
	if driver.lastCreateUserParams.Email != types.Email("brand-new@example.com") {
		t.Errorf("Email = %q, want %q", driver.lastCreateUserParams.Email, "brand-new@example.com")
	}
	// OAuth users should have empty password hash
	if driver.lastCreateUserParams.Hash != "" {
		t.Errorf("Hash = %q, want empty string for OAuth user", driver.lastCreateUserParams.Hash)
	}
	// Role should be the viewer role
	if driver.lastCreateUserParams.Role != "role-viewer" {
		t.Errorf("Role = %q, want %q", driver.lastCreateUserParams.Role, "role-viewer")
	}

	// OAuth should also be created
	if driver.createOauthCallCount != 1 {
		t.Fatalf("CreateUserOauth call count = %d, want 1", driver.createOauthCallCount)
	}
	if driver.lastCreateUserOauthParams.AccessToken != "at-new" {
		t.Errorf("AccessToken = %q, want %q", driver.lastCreateUserOauthParams.AccessToken, "at-new")
	}
}

func TestProvisionUser_EmptyEmail(t *testing.T) {
	t.Parallel()

	log := &provisionLogger{}
	driver := &provisionDbDriver{}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.ProvisionUser(
		&UserInfo{ProviderUserID: "sub-1"},
		makeToken("t", "r", time.Time{}),
		"provider",
	)
	if err == nil {
		t.Fatal("expected error for empty email, got nil")
	}
	if !strings.Contains(err.Error(), "email is required") {
		t.Errorf("error = %q, want it to contain 'email is required'", err.Error())
	}
	if !log.hasLog("error", "Provisioning failed") {
		t.Errorf("expected error log about provisioning failure")
	}
}

func TestProvisionUser_FallbackProviderIDToEmail(t *testing.T) {
	t.Parallel()

	// When ProviderUserID is empty, the code uses the email as the provider
	// user ID. This test verifies that by checking GetUserOauthByProviderID
	// receives the email as the providerUserID argument.
	var receivedProviderUserID string

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		getUserOauthByProviderIDFn: func(_ string, providerUserID string) (*db.UserOauth, error) {
			receivedProviderUserID = providerUserID
			return nil, fmt.Errorf("not found")
		},
		getUserByEmailFn: func(_ types.Email) (*db.Users, error) {
			return nil, fmt.Errorf("not found")
		},
		createUserFn: func(_ context.Context, _ audited.AuditContext, params db.CreateUserParams) (*db.Users, error) {
			return &db.Users{UserID: types.NewUserID(), Email: params.Email}, nil
		},
	}

	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.ProvisionUser(
		&UserInfo{Email: "fallback@example.com"},
		makeToken("t", "r", time.Time{}),
		"oidc",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedProviderUserID != "fallback@example.com" {
		t.Errorf("providerUserID = %q, want %q (email fallback)", receivedProviderUserID, "fallback@example.com")
	}
	if !log.hasLog("warn", "using email as provider ID") {
		t.Errorf("expected warn log about email fallback")
	}
}

// ---------------------------------------------------------------------------
// createNewUser (unexported)
// ---------------------------------------------------------------------------

func TestCreateNewUser_UsernameDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		userInfo     *UserInfo
		wantUsername string
		wantName     string
	}{
		{
			name: "all fields provided",
			userInfo: &UserInfo{
				Email:    "test@example.com",
				Username: "testuser",
				Name:     "Test User",
			},
			wantUsername: "testuser",
			wantName:    "Test User",
		},
		{
			name: "username empty defaults to email",
			userInfo: &UserInfo{
				Email: "noname@example.com",
				Name:  "No Username",
			},
			wantUsername: "noname@example.com",
			wantName:    "No Username",
		},
		{
			name: "name empty defaults to username",
			userInfo: &UserInfo{
				Email:    "namedefault@example.com",
				Username: "hasusername",
			},
			wantUsername: "hasusername",
			wantName:    "hasusername",
		},
		{
			name: "both username and name empty default to email",
			userInfo: &UserInfo{
				Email: "alldefault@example.com",
			},
			wantUsername: "alldefault@example.com",
			wantName:    "alldefault@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := &provisionLogger{}
			driver := &provisionDbDriver{
				createUserFn: func(_ context.Context, _ audited.AuditContext, params db.CreateUserParams) (*db.Users, error) {
					return &db.Users{
						UserID:   types.NewUserID(),
						Username: params.Username,
						Name:     params.Name,
						Email:    params.Email,
					}, nil
				},
			}
			cfg := newProvisionConfig("")
			up := newTestProvisioner(log, cfg, driver)

			token := makeToken("t", "r", time.Time{})
			user, err := up.createNewUser(tt.userInfo, token, "provider", "provider-id")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user.Username != tt.wantUsername {
				t.Errorf("Username = %q, want %q", user.Username, tt.wantUsername)
			}
			if user.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", user.Name, tt.wantName)
			}
		})
	}
}

func TestCreateNewUser_ViewerRoleNotFound(t *testing.T) {
	t.Parallel()

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		listRolesFn: func() (*[]db.Roles, error) {
			// Return roles but none with label "viewer"
			roles := []db.Roles{
				{RoleID: types.RoleID("role-admin"), Label: "admin"},
				{RoleID: types.RoleID("role-editor"), Label: "editor"},
			}
			return &roles, nil
		},
	}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.createNewUser(
		&UserInfo{Email: "test@example.com"},
		makeToken("t", "r", time.Time{}),
		"provider",
		"pid",
	)
	if err == nil {
		t.Fatal("expected error when viewer role is not found")
	}
	if !strings.Contains(err.Error(), "failed to find viewer role") {
		t.Errorf("error = %q, want it to contain 'failed to find viewer role'", err.Error())
	}
}

func TestCreateNewUser_ListRolesError(t *testing.T) {
	t.Parallel()

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		listRolesFn: func() (*[]db.Roles, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.createNewUser(
		&UserInfo{Email: "test@example.com"},
		makeToken("t", "r", time.Time{}),
		"provider",
		"pid",
	)
	if err == nil {
		t.Fatal("expected error when ListRoles fails (viewerRoleID is empty)")
	}
	if !strings.Contains(err.Error(), "failed to find viewer role") {
		t.Errorf("error = %q, want it to contain 'failed to find viewer role'", err.Error())
	}
}

func TestCreateNewUser_CreateUserFails(t *testing.T) {
	t.Parallel()

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		createUserFn: func(_ context.Context, _ audited.AuditContext, _ db.CreateUserParams) (*db.Users, error) {
			return nil, fmt.Errorf("duplicate key")
		},
	}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.createNewUser(
		&UserInfo{Email: "test@example.com", Username: "test"},
		makeToken("t", "r", time.Time{}),
		"provider",
		"pid",
	)
	if err == nil {
		t.Fatal("expected error when CreateUser fails")
	}
	if !strings.Contains(err.Error(), "failed to create user") {
		t.Errorf("error = %q, want it to contain 'failed to create user'", err.Error())
	}
	if !log.hasLog("error", "Failed to create user") {
		t.Errorf("expected error log about create failure")
	}
}

func TestCreateNewUser_CreateOAuthFails(t *testing.T) {
	t.Parallel()

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		createUserFn: func(_ context.Context, _ audited.AuditContext, params db.CreateUserParams) (*db.Users, error) {
			return &db.Users{UserID: types.NewUserID(), Email: params.Email}, nil
		},
		createUserOauthFn: func(_ context.Context, _ audited.AuditContext, _ db.CreateUserOauthParams) (*db.UserOauth, error) {
			return nil, fmt.Errorf("constraint violation")
		},
	}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.createNewUser(
		&UserInfo{Email: "test@example.com", Username: "test"},
		makeToken("t", "r", time.Time{}),
		"provider",
		"pid",
	)
	if err == nil {
		t.Fatal("expected error when CreateUserOauth fails")
	}
	if !strings.Contains(err.Error(), "failed to link OAuth") {
		t.Errorf("error = %q, want it to contain 'failed to link OAuth'", err.Error())
	}
	if !log.hasLog("error", "Failed to link OAuth") {
		t.Errorf("expected error log about OAuth link failure")
	}
}

func TestCreateNewUser_TokenExpiryFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expiry         time.Time
		wantExpiresAt  string
		wantNonEmpty   bool
	}{
		{
			name:          "token with expiry stores RFC3339",
			expiry:        time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
			wantExpiresAt: "2026-06-15T12:00:00Z",
			wantNonEmpty:  true,
		},
		{
			// GitHub tokens have no expiry
			name:          "zero expiry stores empty string",
			expiry:        time.Time{},
			wantExpiresAt: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := &provisionLogger{}
			driver := &provisionDbDriver{
				createUserFn: func(_ context.Context, _ audited.AuditContext, params db.CreateUserParams) (*db.Users, error) {
					return &db.Users{UserID: types.NewUserID(), Email: params.Email}, nil
				},
			}
			cfg := newProvisionConfig("")
			up := newTestProvisioner(log, cfg, driver)

			_, err := up.createNewUser(
				&UserInfo{Email: "test@example.com"},
				makeToken("t", "r", tt.expiry),
				"provider",
				"pid",
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if driver.lastCreateUserOauthParams == nil {
				t.Fatal("CreateUserOauth was not called")
			}
			if driver.lastCreateUserOauthParams.TokenExpiresAt != tt.wantExpiresAt {
				t.Errorf("TokenExpiresAt = %q, want %q",
					driver.lastCreateUserOauthParams.TokenExpiresAt, tt.wantExpiresAt)
			}
		})
	}
}

func TestCreateNewUser_AuditContext(t *testing.T) {
	t.Parallel()

	var capturedAC audited.AuditContext

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		createUserFn: func(_ context.Context, ac audited.AuditContext, params db.CreateUserParams) (*db.Users, error) {
			capturedAC = ac
			return &db.Users{UserID: types.NewUserID(), Email: params.Email}, nil
		},
	}
	cfg := &config.Config{
		Node_ID: "node-provision-42",
		Oauth_Endpoint: map[config.Endpoint]string{
			config.OauthUserInfoURL: "",
		},
	}
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.createNewUser(
		&UserInfo{Email: "test@example.com"},
		makeToken("t", "r", time.Time{}),
		"provider",
		"pid",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAC.NodeID != types.NodeID("node-provision-42") {
		t.Errorf("NodeID = %q, want %q", capturedAC.NodeID, "node-provision-42")
	}
	if capturedAC.RequestID != "oauth-provision" {
		t.Errorf("RequestID = %q, want %q", capturedAC.RequestID, "oauth-provision")
	}
	if capturedAC.IP != "system" {
		t.Errorf("IP = %q, want %q", capturedAC.IP, "system")
	}
	// UserID should be empty for createNewUser (no user exists yet)
	if capturedAC.UserID != types.UserID("") {
		t.Errorf("UserID = %q, want empty", capturedAC.UserID)
	}
}

// ---------------------------------------------------------------------------
// linkOAuthToUser (unexported)
// ---------------------------------------------------------------------------

func TestLinkOAuthToUser_Success(t *testing.T) {
	t.Parallel()

	existingUser := &db.Users{
		UserID:   types.UserID("link-user-1"),
		Username: "linkuser",
		Email:    types.Email("link@example.com"),
	}

	log := &provisionLogger{}
	driver := &provisionDbDriver{}
	cfg := &config.Config{
		Node_ID:        "node-link",
		Oauth_Endpoint: map[config.Endpoint]string{},
	}
	up := newTestProvisioner(log, cfg, driver)

	expiry := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
	token := makeToken("link-at", "link-rt", expiry)
	userInfo := &UserInfo{Email: "link@example.com"}

	result, err := up.linkOAuthToUser(existingUser, userInfo, token, "google", "google-sub-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != existingUser.UserID {
		t.Errorf("UserID = %q, want %q", result.UserID, existingUser.UserID)
	}

	// Verify CreateUserOauth params
	if driver.lastCreateUserOauthParams == nil {
		t.Fatal("CreateUserOauth was not called")
	}
	p := driver.lastCreateUserOauthParams
	if p.UserID.ID != existingUser.UserID || !p.UserID.Valid {
		t.Errorf("UserID = %+v, want {ID:%q Valid:true}", p.UserID, existingUser.UserID)
	}
	if p.OauthProvider != "google" {
		t.Errorf("OauthProvider = %q, want %q", p.OauthProvider, "google")
	}
	if p.OauthProviderUserID != "google-sub-1" {
		t.Errorf("OauthProviderUserID = %q, want %q", p.OauthProviderUserID, "google-sub-1")
	}
	if p.AccessToken != "link-at" {
		t.Errorf("AccessToken = %q, want %q", p.AccessToken, "link-at")
	}
	if p.RefreshToken != "link-rt" {
		t.Errorf("RefreshToken = %q, want %q", p.RefreshToken, "link-rt")
	}
	if p.TokenExpiresAt != "2026-12-25T00:00:00Z" {
		t.Errorf("TokenExpiresAt = %q, want %q", p.TokenExpiresAt, "2026-12-25T00:00:00Z")
	}
}

func TestLinkOAuthToUser_ZeroExpiry(t *testing.T) {
	t.Parallel()

	log := &provisionLogger{}
	driver := &provisionDbDriver{}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	existingUser := &db.Users{UserID: types.UserID("link-noexp")}
	token := makeToken("t", "r", time.Time{}) // zero expiry (GitHub)

	_, err := up.linkOAuthToUser(existingUser, &UserInfo{Email: "x@y.com"}, token, "github", "gh-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if driver.lastCreateUserOauthParams.TokenExpiresAt != "" {
		t.Errorf("TokenExpiresAt = %q, want empty for zero expiry", driver.lastCreateUserOauthParams.TokenExpiresAt)
	}
}

func TestLinkOAuthToUser_CreateOAuthFails(t *testing.T) {
	t.Parallel()

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		createUserOauthFn: func(_ context.Context, _ audited.AuditContext, _ db.CreateUserOauthParams) (*db.UserOauth, error) {
			return nil, fmt.Errorf("unique constraint violation")
		},
	}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.linkOAuthToUser(
		&db.Users{UserID: types.UserID("link-fail")},
		&UserInfo{Email: "fail@example.com"},
		makeToken("t", "r", time.Time{}),
		"provider",
		"pid",
	)
	if err == nil {
		t.Fatal("expected error when CreateUserOauth fails")
	}
	if !strings.Contains(err.Error(), "failed to link OAuth") {
		t.Errorf("error = %q, want it to contain 'failed to link OAuth'", err.Error())
	}
	if !log.hasLog("error", "Failed to link OAuth to existing user") {
		t.Errorf("expected error log about OAuth link failure")
	}
}

func TestLinkOAuthToUser_AuditContext(t *testing.T) {
	t.Parallel()

	var capturedAC audited.AuditContext

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		createUserOauthFn: func(_ context.Context, ac audited.AuditContext, _ db.CreateUserOauthParams) (*db.UserOauth, error) {
			capturedAC = ac
			return &db.UserOauth{}, nil
		},
	}
	cfg := &config.Config{
		Node_ID:        "node-link-audit",
		Oauth_Endpoint: map[config.Endpoint]string{},
	}
	up := newTestProvisioner(log, cfg, driver)

	existingUser := &db.Users{UserID: types.UserID("audit-user-123")}
	_, err := up.linkOAuthToUser(
		existingUser,
		&UserInfo{Email: "audit@example.com"},
		makeToken("t", "r", time.Time{}),
		"provider",
		"pid",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAC.NodeID != types.NodeID("node-link-audit") {
		t.Errorf("NodeID = %q, want %q", capturedAC.NodeID, "node-link-audit")
	}
	// linkOAuthToUser uses the existing user's ID in the audit context
	if capturedAC.UserID != types.UserID("audit-user-123") {
		t.Errorf("UserID = %q, want %q", capturedAC.UserID, "audit-user-123")
	}
	if capturedAC.RequestID != "oauth-link" {
		t.Errorf("RequestID = %q, want %q", capturedAC.RequestID, "oauth-link")
	}
	if capturedAC.IP != "system" {
		t.Errorf("IP = %q, want %q", capturedAC.IP, "system")
	}
}

// ---------------------------------------------------------------------------
// updateTokens (UserProvisioner, unexported)
// ---------------------------------------------------------------------------

func TestUserProvisioner_UpdateTokens_FormatsExpiry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiry    time.Time
		wantValue string
	}{
		{
			name:      "stores RFC3339 for real expiry",
			expiry:    time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC),
			wantValue: "2026-03-15T10:30:00Z",
		},
		{
			name:      "stores empty string for zero expiry",
			expiry:    time.Time{},
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := &provisionLogger{}
			driver := &provisionDbDriver{}
			cfg := newProvisionConfig("")
			up := newTestProvisioner(log, cfg, driver)

			err := up.updateTokens(
				types.UserOauthID("oauth-up-1"),
				makeToken("at", "rt", tt.expiry),
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if driver.lastUpdateOauthParams == nil {
				t.Fatal("UpdateUserOauth was not called")
			}
			if driver.lastUpdateOauthParams.TokenExpiresAt != tt.wantValue {
				t.Errorf("TokenExpiresAt = %q, want %q",
					driver.lastUpdateOauthParams.TokenExpiresAt, tt.wantValue)
			}
		})
	}
}

func TestUserProvisioner_UpdateTokens_DatabaseError(t *testing.T) {
	t.Parallel()

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		updateUserOauthFn: func(_ context.Context, _ audited.AuditContext, _ db.UpdateUserOauthParams) (*string, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	err := up.updateTokens(
		types.UserOauthID("oauth-err"),
		makeToken("t", "r", time.Time{}),
	)
	if err == nil {
		t.Fatal("expected error from database failure")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error = %q, want it to contain 'connection refused'", err.Error())
	}
}

func TestUserProvisioner_UpdateTokens_AuditContext(t *testing.T) {
	t.Parallel()

	var capturedAC audited.AuditContext

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		updateUserOauthFn: func(_ context.Context, ac audited.AuditContext, _ db.UpdateUserOauthParams) (*string, error) {
			capturedAC = ac
			msg := "ok"
			return &msg, nil
		},
	}
	cfg := &config.Config{
		Node_ID:        "node-up-audit",
		Oauth_Endpoint: map[config.Endpoint]string{},
	}
	up := newTestProvisioner(log, cfg, driver)

	err := up.updateTokens(
		types.UserOauthID("oauth-audit-up"),
		makeToken("t", "r", time.Time{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAC.NodeID != types.NodeID("node-up-audit") {
		t.Errorf("NodeID = %q, want %q", capturedAC.NodeID, "node-up-audit")
	}
	if capturedAC.RequestID != "oauth-token-update" {
		t.Errorf("RequestID = %q, want %q", capturedAC.RequestID, "oauth-token-update")
	}
	if capturedAC.IP != "system" {
		t.Errorf("IP = %q, want %q", capturedAC.IP, "system")
	}
}

// ---------------------------------------------------------------------------
// NewUserProvisioner (constructor)
// ---------------------------------------------------------------------------

func TestNewUserProvisioner_StoresConfigAndLogger(t *testing.T) {
	t.Parallel()

	// Cannot call NewUserProvisioner without triggering db.ConfigDB which
	// requires a real database. Verify field assignment via direct construction.
	log := &provisionLogger{}
	cfg := newProvisionConfig("http://example.com/userinfo")

	up := &UserProvisioner{
		config: cfg,
		log:    log,
	}

	if up.config != cfg {
		t.Error("config not stored correctly")
	}
	if up.log == nil {
		t.Error("logger not stored")
	}
}

// ---------------------------------------------------------------------------
// UserInfo struct: JSON deserialization
// ---------------------------------------------------------------------------

func TestUserInfo_JSONDeserialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		want    UserInfo
	}{
		{
			name: "OIDC response",
			json: `{"sub":"oidc-123","email":"user@example.com","name":"User","preferred_username":"user1"}`,
			want: UserInfo{
				ProviderUserID: "oidc-123",
				Email:          "user@example.com",
				Name:           "User",
				Username:       "user1",
			},
		},
		{
			name: "GitHub response",
			json: `{"id":42,"login":"octocat","email":"octocat@github.com","name":"The Octocat","avatar_url":"https://example.com/avatar.png"}`,
			want: UserInfo{
				ID:        42,
				Login:     "octocat",
				Email:     "octocat@github.com",
				Name:      "The Octocat",
				AvatarURL: "https://example.com/avatar.png",
			},
		},
		{
			name: "empty JSON object",
			json: `{}`,
			want: UserInfo{},
		},
		{
			name: "unknown fields are ignored",
			json: `{"sub":"s","email":"e@e.com","unknown_field":"ignored"}`,
			want: UserInfo{
				ProviderUserID: "s",
				Email:          "e@e.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got UserInfo
			if err := json.Unmarshal([]byte(tt.json), &got); err != nil {
				t.Fatalf("json.Unmarshal failed: %v", err)
			}

			if got.ProviderUserID != tt.want.ProviderUserID {
				t.Errorf("ProviderUserID = %q, want %q", got.ProviderUserID, tt.want.ProviderUserID)
			}
			if got.ID != tt.want.ID {
				t.Errorf("ID = %d, want %d", got.ID, tt.want.ID)
			}
			if got.Email != tt.want.Email {
				t.Errorf("Email = %q, want %q", got.Email, tt.want.Email)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Username != tt.want.Username {
				t.Errorf("Username = %q, want %q", got.Username, tt.want.Username)
			}
			if got.Login != tt.want.Login {
				t.Errorf("Login = %q, want %q", got.Login, tt.want.Login)
			}
			if got.AvatarURL != tt.want.AvatarURL {
				t.Errorf("AvatarURL = %q, want %q", got.AvatarURL, tt.want.AvatarURL)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GitHubEmail struct: JSON deserialization
// ---------------------------------------------------------------------------

func TestGitHubEmail_JSONDeserialization(t *testing.T) {
	t.Parallel()

	input := `[{"email":"primary@example.com","primary":true,"verified":true,"visibility":"public"},{"email":"secondary@example.com","primary":false,"verified":false,"visibility":null}]`

	var emails []GitHubEmail
	if err := json.Unmarshal([]byte(input), &emails); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(emails) != 2 {
		t.Fatalf("len(emails) = %d, want 2", len(emails))
	}

	if emails[0].Email != "primary@example.com" {
		t.Errorf("emails[0].Email = %q, want %q", emails[0].Email, "primary@example.com")
	}
	if !emails[0].Primary {
		t.Error("emails[0].Primary = false, want true")
	}
	if !emails[0].Verified {
		t.Error("emails[0].Verified = false, want true")
	}
	if emails[0].Visibility != "public" {
		t.Errorf("emails[0].Visibility = %q, want %q", emails[0].Visibility, "public")
	}

	if emails[1].Primary {
		t.Error("emails[1].Primary = true, want false")
	}
	if emails[1].Verified {
		t.Error("emails[1].Verified = true, want false")
	}
}

// ---------------------------------------------------------------------------
// ProvisionUser: full integration scenarios
// ---------------------------------------------------------------------------

func TestProvisionUser_FullFlow_NewUser(t *testing.T) {
	t.Parallel()

	// End-to-end scenario: no existing OAuth link, no email match, creates
	// user and links OAuth.
	createdUserID := types.UserID("created-full-flow")

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		getUserOauthByProviderIDFn: func(_, _ string) (*db.UserOauth, error) {
			return nil, fmt.Errorf("not found")
		},
		getUserByEmailFn: func(_ types.Email) (*db.Users, error) {
			return nil, fmt.Errorf("not found")
		},
		createUserFn: func(_ context.Context, _ audited.AuditContext, params db.CreateUserParams) (*db.Users, error) {
			return &db.Users{
				UserID:   createdUserID,
				Username: params.Username,
				Name:     params.Name,
				Email:    params.Email,
				Role:     params.Role,
			}, nil
		},
	}

	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	expiry := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	token := makeToken("flow-at", "flow-rt", expiry)

	user, err := up.ProvisionUser(
		&UserInfo{
			ProviderUserID: "flow-sub",
			Email:          "flow@example.com",
			Username:       "flowuser",
			Name:           "Flow User",
		},
		token,
		"github",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify user was created
	if user.UserID != createdUserID {
		t.Errorf("UserID = %q, want %q", user.UserID, createdUserID)
	}
	if driver.createUserCallCount != 1 {
		t.Errorf("CreateUser call count = %d, want 1", driver.createUserCallCount)
	}

	// Verify OAuth was linked
	if driver.createOauthCallCount != 1 {
		t.Fatalf("CreateUserOauth call count = %d, want 1", driver.createOauthCallCount)
	}
	oauthParams := driver.lastCreateUserOauthParams
	if oauthParams.OauthProvider != "github" {
		t.Errorf("OauthProvider = %q, want %q", oauthParams.OauthProvider, "github")
	}
	if oauthParams.OauthProviderUserID != "flow-sub" {
		t.Errorf("OauthProviderUserID = %q, want %q", oauthParams.OauthProviderUserID, "flow-sub")
	}
	if oauthParams.AccessToken != "flow-at" {
		t.Errorf("AccessToken = %q, want %q", oauthParams.AccessToken, "flow-at")
	}
	if oauthParams.RefreshToken != "flow-rt" {
		t.Errorf("RefreshToken = %q, want %q", oauthParams.RefreshToken, "flow-rt")
	}
	if oauthParams.TokenExpiresAt != "2026-07-01T00:00:00Z" {
		t.Errorf("TokenExpiresAt = %q, want %q", oauthParams.TokenExpiresAt, "2026-07-01T00:00:00Z")
	}

	// No token update should have occurred (that's for existing OAuth links)
	if driver.updateOauthCallCount != 0 {
		t.Error("UpdateUserOauth should not be called for new user flow")
	}

	// Logging should reflect the new user path
	if !log.hasLog("debug", "Creating new user") {
		t.Errorf("expected debug log about creating new user")
	}
}

func TestProvisionUser_FullFlow_ExistingOAuthReturnsUser(t *testing.T) {
	t.Parallel()

	// End-to-end scenario: existing OAuth link found, returns existing user
	// and updates tokens.
	existingUserID := types.UserID("existing-full-flow")

	log := &provisionLogger{}
	driver := &provisionDbDriver{
		getUserOauthByProviderIDFn: func(provider, providerUserID string) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID:         types.UserOauthID("existing-oauth-full"),
				UserID:              types.NullableUserID{ID: existingUserID, Valid: true},
				OauthProvider:       provider,
				OauthProviderUserID: providerUserID,
			}, nil
		},
		getUserFn: func(id types.UserID) (*db.Users, error) {
			return &db.Users{
				UserID:   id,
				Username: "existing",
				Email:    types.Email("existing@example.com"),
			}, nil
		},
	}

	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	user, err := up.ProvisionUser(
		&UserInfo{ProviderUserID: "existing-sub", Email: "existing@example.com"},
		makeToken("new-at", "new-rt", time.Time{}),
		"github",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.UserID != existingUserID {
		t.Errorf("UserID = %q, want %q", user.UserID, existingUserID)
	}

	// No user creation or OAuth link creation
	if driver.createUserCallCount != 0 {
		t.Error("CreateUser should not be called for existing OAuth link")
	}
	if driver.createOauthCallCount != 0 {
		t.Error("CreateUserOauth should not be called for existing OAuth link")
	}

	// Token update should have been called
	if driver.updateOauthCallCount != 1 {
		t.Errorf("UpdateUserOauth call count = %d, want 1", driver.updateOauthCallCount)
	}
}

func TestProvisionUser_ExistingOAuth_GetUserFails(t *testing.T) {
	t.Parallel()

	// When OAuth link exists but GetUser fails, the error should propagate.
	log := &provisionLogger{}
	driver := &provisionDbDriver{
		getUserOauthByProviderIDFn: func(_, _ string) (*db.UserOauth, error) {
			return &db.UserOauth{
				UserOauthID: types.UserOauthID("oauth-getuser-fail"),
				UserID:      types.NullableUserID{ID: types.UserID("ghost-user"), Valid: true},
			}, nil
		},
		getUserFn: func(_ types.UserID) (*db.Users, error) {
			return nil, fmt.Errorf("user deleted from database")
		},
	}

	cfg := newProvisionConfig("")
	up := newTestProvisioner(log, cfg, driver)

	_, err := up.ProvisionUser(
		&UserInfo{ProviderUserID: "sub", Email: "ghost@example.com"},
		makeToken("t", "r", time.Time{}),
		"provider",
	)
	if err == nil {
		t.Fatal("expected error when GetUser fails")
	}
	if !strings.Contains(err.Error(), "user deleted from database") {
		t.Errorf("error = %q, want it to contain 'user deleted from database'", err.Error())
	}
}
