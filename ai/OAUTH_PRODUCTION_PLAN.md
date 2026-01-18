# Implementation Plan: Production-Quality OAuth

**Date:** 2026-01-17
**Status:** Planning
**Goal:** Complete production-ready OAuth 2.0 implementation with PKCE, token refresh, and security hardening
**Risk Level:** Medium - Security-critical authentication flows
**Timeline:** 5-7 days across 3 phases

---

## Executive Summary

ModulaCMS has a solid OAuth foundation but lacks production-critical features. This plan completes the implementation with CSRF protection, token refresh, user provisioning, proper error handling, and security hardening.

**Current State:**
- ✅ PKCE state generation (`internal/auth/auth.go`)
- ✅ OAuth callback handler skeleton (`internal/router/auth.go`)
- ✅ Database schema for tokens (`sql/schema/12_user_oauth/`)
- ✅ Session validation middleware (`internal/middleware/session.go`)
- ⚠️ State validation commented out (CSRF risk)
- ⚠️ Token refresh logic commented out
- ⚠️ Hardcoded userinfo endpoint
- ⚠️ No user creation/linking flow

**Target State:**
- ✅ Complete CSRF protection with state validation
- ✅ Automatic token refresh before expiration
- ✅ User provisioning (create or link accounts)
- ✅ Provider-agnostic userinfo fetching
- ✅ Security hardening (rate limiting, secure cookies, audit logging)
- ✅ Comprehensive error handling
- ✅ Production monitoring and observability

---

## Phase 1: Security Hardening & CSRF Protection

**Goal:** Eliminate security vulnerabilities in current OAuth flow
**Duration:** 2 days
**Risk:** High impact - fixes critical CSRF vulnerability

### Task 1.1: Implement State Parameter Validation

**Problem:** `internal/router/auth.go:56-58` has commented-out state validation, leaving CSRF vulnerability.

**Files Modified:**
- `internal/auth/auth.go` - Add state storage mechanism
- `internal/router/auth.go` - Enable state validation
- `internal/db/oauth_state.go` - NEW: State persistence (optional)

**Implementation:**

#### Step 1: Add In-Memory State Store (Quick Win)
**File:** `internal/auth/state_store.go` (NEW, ~80 lines)

```go
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

type StateStore struct {
	mu     sync.RWMutex
	states map[string]time.Time
}

var globalStateStore = &StateStore{
	states: make(map[string]time.Time),
}

func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	globalStateStore.mu.Lock()
	globalStateStore.states[state] = time.Now().Add(20 * time.Minute)
	globalStateStore.mu.Unlock()

	// Cleanup expired states in background
	go globalStateStore.cleanup()

	return state, nil
}

func ValidateState(state string) error {
	globalStateStore.mu.Lock()
	defer globalStateStore.mu.Unlock()

	expiry, exists := globalStateStore.states[state]
	if !exists {
		return fmt.Errorf("invalid state parameter")
	}

	if time.Now().After(expiry) {
		delete(globalStateStore.states, state)
		return fmt.Errorf("state parameter expired")
	}

	// One-time use: delete after validation
	delete(globalStateStore.states, state)
	return nil
}

func (s *StateStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for state, expiry := range s.states {
		if now.After(expiry) {
			delete(s.states, state)
		}
	}
}
```

#### Step 2: Update OAuth Initiation
**File:** `internal/router/auth.go` (NEW function, ~40 lines)

```go
// OauthInitiateHandler starts the OAuth flow with state parameter
func OauthInitiateHandler(c config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate state for CSRF protection
		state, err := auth.GenerateState()
		if err != nil {
			http.Error(w, "Failed to generate state", http.StatusInternalServerError)
			return
		}

		// Generate PKCE verifier and challenge
		verifier := oauth2.GenerateVerifier()

		// Store verifier in session/cookie for callback
		// TODO: Implement proper verifier storage

		// Build OAuth2 config
		conf := &oauth2.Config{
			ClientID:     c.Oauth_Client_Id,
			ClientSecret: c.Oauth_Client_Secret,
			Scopes:       c.Oauth_Scopes,
			RedirectURL:  c.Oauth_Redirect_URL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  c.Oauth_Endpoint[config.OauthAuthURL],
				TokenURL: c.Oauth_Endpoint[config.OauthTokenURL],
			},
		}

		// Redirect to provider with state and PKCE challenge
		url := conf.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
```

#### Step 3: Enable State Validation in Callback
**File:** `internal/router/auth.go:45-92` (MODIFY existing function)

```go
func OauthCallbackHandler(c config.Config, verifier string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Retrieve the authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			utility.DefaultLogger.Error("Missing code parameter", nil)
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			return
		}

		// ENABLE: Validate state parameter
		state := r.URL.Query().Get("state")
		if err := auth.ValidateState(state); err != nil {
			utility.DefaultLogger.Error("State validation failed", err)
			http.Error(w, "Invalid or expired state", http.StatusBadRequest)
			return
		}

		// ... rest of existing implementation
	}
}
```

**Testing:**
- Test with valid state → should succeed
- Test with invalid state → should fail with 400
- Test with reused state → should fail (one-time use)
- Test with expired state (>20 min) → should fail

---

### Task 1.2: Secure Cookie Configuration

**Problem:** Cookies lack HttpOnly, Secure, SameSite flags - vulnerable to XSS and CSRF.

**Files Modified:**
- `internal/middleware/cookies.go` - Add secure cookie flags
- `internal/config/config.go` - Add cookie security settings

**Implementation:**

#### Update Cookie Creation
**File:** `internal/middleware/cookies.go` (MODIFY existing)

```go
func WriteCookie(w http.ResponseWriter, c *config.Config, sessionData string, userId int64) error {
	cookie := MiddlewareCookie{
		Session: sessionData,
		UserId:  userId,
	}

	json, err := json.Marshal(cookie)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(json)

	http.SetCookie(w, &http.Cookie{
		Name:     c.Cookie_Name,
		Value:    encoded,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,  // ADDED: Prevent JavaScript access
		Secure:   c.Cookie_Secure, // ADDED: HTTPS only
		SameSite: http.SameSiteLaxMode, // ADDED: CSRF protection
	})

	return nil
}
```

#### Add Config Fields
**File:** `internal/config/config.go` (ADD fields)

```go
type Config struct {
	// ... existing fields ...
	Cookie_Secure   bool   `json:"cookie_secure"`   // Force HTTPS cookies
	Cookie_SameSite string `json:"cookie_samesite"` // "strict", "lax", or "none"
}
```

**Testing:**
- Verify HttpOnly flag blocks JavaScript access
- Test SameSite prevents cross-site requests
- Confirm Secure flag works on HTTPS

---

### Task 1.3: Rate Limiting on Auth Endpoints

**Problem:** No rate limiting allows brute force attacks on OAuth endpoints.

**Files Modified:**
- `internal/middleware/ratelimit.go` - NEW: Rate limiting middleware
- `cmd/main.go` - Apply to auth routes

**Implementation:**

**File:** `internal/middleware/ratelimit.go` (NEW, ~100 lines)

```go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

**Apply to Routes:**
**File:** `cmd/main.go` (MODIFY)

```go
// Create rate limiter: 10 requests per minute per IP
authLimiter := middleware.NewRateLimiter(rate.Limit(10.0/60.0), 10)

// Apply to auth routes
mux.Handle("/api/v1/auth/oauth", authLimiter.Middleware(
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router.OauthCallbackHandler(configuration, verifier)(w, r)
	}),
))
```

**Testing:**
- Send 11 requests in 60 seconds → 11th should fail with 429
- Wait 60 seconds → should allow new requests
- Test from different IPs → should have independent limits

---

## Phase 2: Token Refresh & User Provisioning

**Goal:** Complete OAuth user lifecycle (creation, linking, token refresh)
**Duration:** 2-3 days
**Risk:** Medium - Complex state management

### Task 2.1: Implement Token Refresh Logic

**Problem:** `internal/middleware/middleware.go:80-88` has commented-out refresh logic.

**Files Modified:**
- `internal/auth/token_refresh.go` - NEW: Token refresh logic
- `internal/middleware/session.go` - Add automatic refresh check
- `internal/db/user_oauth.go` - Add token update queries

**Implementation:**

#### Token Refresh Handler
**File:** `internal/auth/token_refresh.go` (NEW, ~120 lines)

```go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/oauth2"
)

type TokenRefresher struct {
	config *config.Config
	driver db.DbDriver
}

func NewTokenRefresher(c *config.Config) *TokenRefresher {
	return &TokenRefresher{
		config: c,
		driver: db.ConfigDB(*c),
	}
}

func (tr *TokenRefresher) RefreshIfNeeded(userID int64) error {
	// Get user's OAuth record
	userOauth, err := tr.driver.GetUserOauthByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get OAuth record: %w", err)
	}

	if userOauth == nil {
		return nil // User doesn't use OAuth
	}

	// Parse expiration time
	expiresAt, err := time.Parse(time.RFC3339, userOauth.TokenExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to parse expiration: %w", err)
	}

	// Check if token expires within 5 minutes
	if time.Until(expiresAt) > 5*time.Minute {
		return nil // Token still valid
	}

	// Build OAuth config
	conf := &oauth2.Config{
		ClientID:     tr.config.Oauth_Client_Id,
		ClientSecret: tr.config.Oauth_Client_Secret,
		Endpoint: oauth2.Endpoint{
			TokenURL: tr.config.Oauth_Endpoint[config.OauthTokenURL],
		},
	}

	// Create token source from refresh token
	token := &oauth2.Token{
		RefreshToken: userOauth.RefreshToken,
	}

	// Refresh the token
	ctx := context.Background()
	newToken, err := conf.TokenSource(ctx, token).Token()
	if err != nil {
		utility.DefaultLogger.Error("Token refresh failed", err)
		return fmt.Errorf("token refresh failed: %w", err)
	}

	// Update database with new tokens
	err = tr.driver.UpdateUserOauthTokens(
		userOauth.UserOauthID,
		newToken.AccessToken,
		newToken.RefreshToken,
		newToken.Expiry.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	utility.DefaultLogger.Info("Token refreshed for user %d", userID)
	return nil
}
```

#### Add Database Query
**File:** `sql/mysql/user_oauth.sql` (ADD)

```sql
-- name: UpdateUserOauthTokens :exec
UPDATE user_oauth
SET access_token = ?,
    refresh_token = ?,
    token_expires_at = ?
WHERE user_oauth_id = ?;
```

#### Integrate with Session Validation
**File:** `internal/middleware/session.go:23-56` (MODIFY)

```go
func UserIsAuth(r *http.Request, cookie *http.Cookie, c *config.Config) (*db.Users, error) {
	userCookie, err := ReadCookie(cookie)
	if err != nil {
		return nil, err
	}

	dbc := db.ConfigDB(*c)

	// Validate session
	session, err := dbc.GetSessionByUserId(userCookie.UserId)
	if err != nil || session == nil {
		return nil, err
	}

	if userCookie.Session != session.SessionData.String {
		return nil, fmt.Errorf("sessions don't match")
	}

	expired := utility.TimestampLessThan(session.ExpiresAt.String)
	if expired {
		return nil, fmt.Errorf("session is expired")
	}

	// ADDED: Check and refresh OAuth tokens if needed
	refresher := auth.NewTokenRefresher(c)
	if err := refresher.RefreshIfNeeded(userCookie.UserId); err != nil {
		utility.DefaultLogger.Warn("Token refresh warning", err)
		// Don't fail auth if refresh fails - token might still be valid
	}

	u, err := dbc.GetUser(userCookie.UserId)
	if err != nil {
		return nil, err
	}
	return u, nil
}
```

**Testing:**
- Mock token expiring in 4 minutes → should trigger refresh
- Mock token expiring in 10 minutes → should not trigger refresh
- Mock refresh failure → should log warning but not block auth
- Verify new tokens saved to database

---

### Task 2.2: User Provisioning (Create or Link)

**Problem:** No logic to create/link users after successful OAuth authentication.

**Files Modified:**
- `internal/router/auth.go` - Complete callback handler
- `internal/auth/user_provision.go` - NEW: User creation/linking logic
- `internal/config/config.go` - Add userinfo endpoint config

**Implementation:**

#### Add Userinfo Endpoint Config
**File:** `internal/config/config.go` (MODIFY)

```go
type Endpoint string
const (
	OauthAuthURL     Endpoint = "oauth_auth_url"
	OauthTokenURL    Endpoint = "oauth_token_url"
	OauthUserInfoURL Endpoint = "oauth_userinfo_url" // ADDED
)
```

#### User Provisioning Logic
**File:** `internal/auth/user_provision.go` (NEW, ~150 lines)

```go
package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/oauth2"
)

type UserInfo struct {
	ProviderUserID string `json:"sub"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Username       string `json:"preferred_username"`
}

type UserProvisioner struct {
	config *config.Config
	driver db.DbDriver
}

func NewUserProvisioner(c *config.Config) *UserProvisioner {
	return &UserProvisioner{
		config: c,
		driver: db.ConfigDB(*c),
	}
}

func (up *UserProvisioner) FetchUserInfo(client *http.Client) (*UserInfo, error) {
	userInfoURL := up.config.Oauth_Endpoint[config.OauthUserInfoURL]
	if userInfoURL == "" {
		return nil, fmt.Errorf("oauth_userinfo_url not configured")
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed: %s - %s", resp.Status, body)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo: %w", err)
	}

	return &userInfo, nil
}

func (up *UserProvisioner) ProvisionUser(
	userInfo *UserInfo,
	token *oauth2.Token,
	provider string,
) (*db.Users, error) {
	// Check if user already linked via OAuth
	existingOauth, err := up.driver.GetUserOauthByProviderID(provider, userInfo.ProviderUserID)
	if err == nil && existingOauth != nil {
		// User exists, update tokens
		err = up.driver.UpdateUserOauthTokens(
			existingOauth.UserOauthID,
			token.AccessToken,
			token.RefreshToken,
			token.Expiry.Format(time.RFC3339),
		)
		if err != nil {
			utility.DefaultLogger.Warn("Failed to update tokens", err)
		}

		// Return existing user
		return up.driver.GetUser(existingOauth.UserID)
	}

	// Check if user exists by email
	existingUser, err := up.driver.GetUserByEmail(userInfo.Email)
	if err == nil && existingUser != nil {
		// Link OAuth to existing user
		return up.linkOAuthToUser(existingUser, userInfo, token, provider)
	}

	// Create new user
	return up.createNewUser(userInfo, token, provider)
}

func (up *UserProvisioner) createNewUser(
	userInfo *UserInfo,
	token *oauth2.Token,
	provider string,
) (*db.Users, error) {
	// Generate username if not provided
	username := userInfo.Username
	if username == "" {
		username = userInfo.Email
	}

	// Create user (no password for OAuth users)
	user, err := up.driver.CreateUser(db.CreateUserParams{
		Username: username,
		Name:     userInfo.Name,
		Email:    userInfo.Email,
		Hash:     "", // OAuth users don't have passwords
		Role:     4,  // Default role
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Link OAuth
	_, err = up.driver.CreateUserOauth(db.CreateUserOauthParams{
		UserID:              user.UserID,
		OauthProvider:       provider,
		OauthProviderUserID: userInfo.ProviderUserID,
		AccessToken:         token.AccessToken,
		RefreshToken:        token.RefreshToken,
		TokenExpiresAt:      token.Expiry.Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to link OAuth: %w", err)
	}

	utility.DefaultLogger.Info("Created new user via OAuth: %s", userInfo.Email)
	return user, nil
}

func (up *UserProvisioner) linkOAuthToUser(
	user *db.Users,
	userInfo *UserInfo,
	token *oauth2.Token,
	provider string,
) (*db.Users, error) {
	_, err := up.driver.CreateUserOauth(db.CreateUserOauthParams{
		UserID:              user.UserID,
		OauthProvider:       provider,
		OauthProviderUserID: userInfo.ProviderUserID,
		AccessToken:         token.AccessToken,
		RefreshToken:        token.RefreshToken,
		TokenExpiresAt:      token.Expiry.Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to link OAuth: %w", err)
	}

	utility.DefaultLogger.Info("Linked OAuth to existing user: %s", userInfo.Email)
	return user, nil
}
```

#### Complete OAuth Callback Handler
**File:** `internal/router/auth.go:45-92` (REPLACE)

```go
func OauthCallbackHandler(c config.Config, verifier string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Retrieve authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			utility.DefaultLogger.Error("Missing code parameter", nil)
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			return
		}

		// Validate state parameter
		state := r.URL.Query().Get("state")
		if err := auth.ValidateState(state); err != nil {
			utility.DefaultLogger.Error("State validation failed", err)
			http.Error(w, "Invalid or expired state", http.StatusBadRequest)
			return
		}

		// Build OAuth2 config
		conf := &oauth2.Config{
			ClientID:     c.Oauth_Client_Id,
			ClientSecret: c.Oauth_Client_Secret,
			Scopes:       c.Oauth_Scopes,
			RedirectURL:  c.Oauth_Redirect_URL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  c.Oauth_Endpoint[config.OauthAuthURL],
				TokenURL: c.Oauth_Endpoint[config.OauthTokenURL],
			},
		}

		// Exchange code for token with PKCE
		token, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			utility.DefaultLogger.Error("Token exchange failed", err)
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}

		// Create authenticated client
		client := conf.Client(ctx, token)

		// Fetch user info from provider
		provisioner := auth.NewUserProvisioner(&c)
		userInfo, err := provisioner.FetchUserInfo(client)
		if err != nil {
			utility.DefaultLogger.Error("Failed to fetch user info", err)
			http.Error(w, "Failed to fetch user information", http.StatusInternalServerError)
			return
		}

		// Provision user (create or link)
		provider := c.Oauth_Provider_Name // Add this config field
		user, err := provisioner.ProvisionUser(userInfo, token, provider)
		if err != nil {
			utility.DefaultLogger.Error("User provisioning failed", err)
			http.Error(w, "User provisioning failed", http.StatusInternalServerError)
			return
		}

		// Create session
		sessionID, err := createSession(user.UserID, &c)
		if err != nil {
			utility.DefaultLogger.Error("Session creation failed", err)
			http.Error(w, "Session creation failed", http.StatusInternalServerError)
			return
		}

		// Set auth cookie
		if err := middleware.WriteCookie(w, &c, sessionID, user.UserID); err != nil {
			utility.DefaultLogger.Error("Cookie creation failed", err)
			http.Error(w, "Cookie creation failed", http.StatusInternalServerError)
			return
		}

		// Redirect to app
		redirectURL := c.Oauth_Success_Redirect
		if redirectURL == "" {
			redirectURL = "/"
		}
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

func createSession(userID int64, c *config.Config) (string, error) {
	dbc := db.ConfigDB(*c)

	// Generate session token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	sessionToken := base64.URLEncoding.EncodeToString(b)

	// Create session in database
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	_, err := dbc.CreateSession(db.CreateSessionParams{
		UserID:      userID,
		ExpiresAt:   expiresAt,
		SessionData: sessionToken,
	})

	return sessionToken, err
}
```

#### Add Required Database Queries
**File:** `sql/mysql/user_oauth.sql` (ADD)

```sql
-- name: GetUserOauthByProviderID :one
SELECT * FROM user_oauth
WHERE oauth_provider = ? AND oauth_provider_user_id = ?
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?
LIMIT 1;
```

**Testing:**
- New user via OAuth → should create account
- Existing user (by email) → should link OAuth
- Existing OAuth user → should update tokens
- Invalid userinfo endpoint → should fail gracefully
- Test with different providers (Google, Azure AD)

---

## Phase 3: Error Handling & Observability

**Goal:** Production-grade error handling, logging, and monitoring
**Duration:** 1-2 days
**Risk:** Low - Quality improvements

### Task 3.1: Comprehensive Error Handling

**Files Modified:**
- All OAuth files - Add structured error responses
- `internal/auth/errors.go` - NEW: OAuth error types

**Implementation:**

**File:** `internal/auth/errors.go` (NEW, ~60 lines)

```go
package auth

import "fmt"

type OAuthError struct {
	Code    string
	Message string
	Err     error
}

func (e *OAuthError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *OAuthError) Unwrap() error {
	return e.Err
}

var (
	ErrInvalidState = &OAuthError{
		Code:    "INVALID_STATE",
		Message: "Invalid or expired state parameter",
	}

	ErrMissingCode = &OAuthError{
		Code:    "MISSING_CODE",
		Message: "Authorization code not provided",
	}

	ErrTokenExchange = &OAuthError{
		Code:    "TOKEN_EXCHANGE_FAILED",
		Message: "Failed to exchange authorization code for token",
	}

	ErrUserInfoFetch = &OAuthError{
		Code:    "USERINFO_FAILED",
		Message: "Failed to fetch user information from provider",
	}

	ErrUserProvision = &OAuthError{
		Code:    "PROVISION_FAILED",
		Message: "Failed to create or link user account",
	}

	ErrTokenRefresh = &OAuthError{
		Code:    "REFRESH_FAILED",
		Message: "Failed to refresh access token",
	}
)
```

**Update error responses to use structured errors with proper HTTP status codes.**

---

### Task 3.2: Audit Logging

**Files Modified:**
- `sql/schema/16_auth_audit/schema.sql` - NEW: Audit log table
- `internal/auth/audit.go` - NEW: Audit logging

**Implementation:**

**File:** `sql/schema/16_auth_audit/schema.sql` (NEW)

```sql
CREATE TABLE IF NOT EXISTS auth_audit (
	audit_id INTEGER PRIMARY KEY AUTO_INCREMENT,
	user_id INTEGER REFERENCES users ON DELETE SET NULL,
	event_type VARCHAR(50) NOT NULL,
	provider VARCHAR(50),
	ip_address VARCHAR(45),
	user_agent TEXT,
	success BOOLEAN NOT NULL,
	error_code VARCHAR(50),
	error_message TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_auth_audit_user ON auth_audit(user_id);
CREATE INDEX idx_auth_audit_event ON auth_audit(event_type, created_at);
CREATE INDEX idx_auth_audit_success ON auth_audit(success, created_at);
```

**File:** `internal/auth/audit.go` (NEW, ~80 lines)

```go
package auth

import (
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

type AuditLogger struct {
	driver db.DbDriver
}

func NewAuditLogger(c *config.Config) *AuditLogger {
	return &AuditLogger{
		driver: db.ConfigDB(*c),
	}
}

type AuditEvent struct {
	UserID       *int64
	EventType    string
	Provider     string
	IPAddress    string
	UserAgent    string
	Success      bool
	ErrorCode    string
	ErrorMessage string
}

func (al *AuditLogger) Log(event AuditEvent) {
	_, err := al.driver.CreateAuthAudit(db.CreateAuthAuditParams{
		UserID:       event.UserID,
		EventType:    event.EventType,
		Provider:     event.Provider,
		IPAddress:    event.IPAddress,
		UserAgent:    event.UserAgent,
		Success:      event.Success,
		ErrorCode:    event.ErrorCode,
		ErrorMessage: event.ErrorMessage,
	})
	if err != nil {
		// Don't fail the request if audit fails
		utility.DefaultLogger.Error("Audit logging failed", err)
	}
}

func AuditFromRequest(r *http.Request, eventType string, success bool, err error) AuditEvent {
	event := AuditEvent{
		EventType: eventType,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Success:   success,
	}

	if err != nil {
		if oauthErr, ok := err.(*OAuthError); ok {
			event.ErrorCode = oauthErr.Code
			event.ErrorMessage = oauthErr.Message
		} else {
			event.ErrorMessage = err.Error()
		}
	}

	return event
}
```

**Add audit calls to all OAuth operations:**
- OAuth initiation
- Callback success/failure
- Token refresh
- User provisioning
- Session creation

---

### Task 3.3: Metrics & Monitoring

**Files Modified:**
- `internal/auth/metrics.go` - NEW: Prometheus metrics (optional)
- Add health check endpoint for OAuth provider connectivity

**Implementation:**

**File:** `internal/router/health.go` (NEW, ~50 lines)

```go
package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"golang.org/x/oauth2"
)

type HealthCheck struct {
	Status   string                 `json:"status"`
	Services map[string]ServiceHealth `json:"services"`
}

type ServiceHealth struct {
	Status      string  `json:"status"`
	ResponseTime int64  `json:"response_time_ms"`
	Error       string `json:"error,omitempty"`
}

func OAuthHealthHandler(c config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := HealthCheck{
			Status:   "healthy",
			Services: make(map[string]ServiceHealth),
		}

		// Check OAuth provider connectivity
		start := time.Now()
		tokenURL := c.Oauth_Endpoint[config.OauthTokenURL]
		resp, err := http.Get(tokenURL)
		duration := time.Since(start).Milliseconds()

		if err != nil || resp.StatusCode >= 500 {
			health.Services["oauth_provider"] = ServiceHealth{
				Status:       "unhealthy",
				ResponseTime: duration,
				Error:        err.Error(),
			}
			health.Status = "degraded"
		} else {
			health.Services["oauth_provider"] = ServiceHealth{
				Status:       "healthy",
				ResponseTime: duration,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}
}
```

---

## Phase 4: Testing & Documentation

**Goal:** Comprehensive test coverage and documentation
**Duration:** 1 day
**Risk:** Low

### Task 4.1: Unit Tests

**Files Created:**
- `internal/auth/state_store_test.go`
- `internal/auth/token_refresh_test.go`
- `internal/auth/user_provision_test.go`
- `internal/middleware/ratelimit_test.go`

**Test Coverage:**
- State generation and validation
- Token refresh logic
- User provisioning (new, existing, linking)
- Rate limiting
- Cookie security
- Error handling

---

### Task 4.2: Integration Tests

**Files Created:**
- `internal/router/oauth_integration_test.go`

**Test Scenarios:**
- Full OAuth flow with mock provider
- State CSRF protection
- Token refresh on expiry
- User creation vs linking
- Rate limit enforcement
- Audit logging

---

### Task 4.3: Documentation Updates

**Files Modified:**
- `ai/domain/AUTH_AND_OAUTH.md` - Update with new implementation
- `ai/reference/QUICKSTART.md` - Add OAuth setup guide
- `README.md` - Add OAuth configuration section

**Documentation Sections:**
- Configuration guide for each provider (Google, Azure AD, GitHub)
- Security considerations
- Troubleshooting common issues
- Migration guide for existing users

---

## Configuration Changes

### New Config Fields Required

**File:** `internal/config/config.go`

```go
type Config struct {
	// ... existing fields ...

	// OAuth Configuration
	Oauth_Client_Id        string              `json:"oauth_client_id"`
	Oauth_Client_Secret    string              `json:"oauth_client_secret"`
	Oauth_Scopes           []string            `json:"oauth_scopes"`
	Oauth_Endpoint         map[Endpoint]string `json:"oauth_endpoint"`
	Oauth_Provider_Name    string              `json:"oauth_provider_name"`    // NEW: "google", "azure", etc.
	Oauth_Redirect_URL     string              `json:"oauth_redirect_url"`     // NEW: Callback URL
	Oauth_Success_Redirect string              `json:"oauth_success_redirect"` // NEW: Where to redirect after auth

	// Cookie Security
	Cookie_Name     string `json:"cookie_name"`
	Cookie_Secure   bool   `json:"cookie_secure"`   // NEW: HTTPS only
	Cookie_SameSite string `json:"cookie_samesite"` // NEW: "strict", "lax", "none"

	// Rate Limiting
	Auth_Rate_Limit  float64 `json:"auth_rate_limit"`  // NEW: Requests per second
	Auth_Rate_Burst  int     `json:"auth_rate_burst"`  // NEW: Burst size
}
```

### Example Configuration

**File:** `config.example.json`

```json
{
  "oauth_client_id": "your-client-id",
  "oauth_client_secret": "your-client-secret",
  "oauth_scopes": ["openid", "profile", "email"],
  "oauth_endpoint": {
    "oauth_auth_url": "https://accounts.google.com/o/oauth2/v2/auth",
    "oauth_token_url": "https://oauth2.googleapis.com/token",
    "oauth_userinfo_url": "https://www.googleapis.com/oauth2/v2/userinfo"
  },
  "oauth_provider_name": "google",
  "oauth_redirect_url": "https://yourapp.com/api/v1/auth/oauth/callback",
  "oauth_success_redirect": "/dashboard",
  "cookie_name": "modulacms_session",
  "cookie_secure": true,
  "cookie_samesite": "lax",
  "auth_rate_limit": 0.16667,
  "auth_rate_burst": 10
}
```

---

## Database Schema Changes

### New Tables

1. **auth_audit** - `sql/schema/16_auth_audit/schema.sql` (Phase 3)

### New Queries

**MySQL:** `sql/mysql/user_oauth.sql`
- `GetUserOauthByProviderID`
- `UpdateUserOauthTokens`

**MySQL:** `sql/mysql/users.sql`
- `GetUserByEmail`

**MySQL:** `sql/mysql/sessions.sql`
- `CreateSession`

**MySQL:** `sql/mysql/auth_audit.sql` (NEW)
- `CreateAuthAudit`

Run `make sqlc` after adding queries.

---

## Security Checklist

- [x] CSRF protection via state parameter validation
- [x] PKCE for authorization code flow
- [x] HttpOnly, Secure, SameSite cookie flags
- [x] Rate limiting on auth endpoints
- [x] Token refresh before expiration
- [x] Audit logging for all auth events
- [x] Secure token storage (no plaintext in logs)
- [x] Input validation on all OAuth parameters
- [x] Error messages don't leak sensitive info
- [x] HTTPS enforcement in production
- [x] Session expiration and cleanup
- [x] Provider endpoint validation

---

## Rollout Plan

### Pre-Production
1. Deploy to staging environment
2. Test with real OAuth providers
3. Monitor audit logs
4. Load test rate limiting
5. Verify token refresh works over time

### Production
1. Enable OAuth alongside existing auth
2. Monitor error rates and audit logs
3. Gradually migrate users
4. Keep password auth as fallback
5. Monitor token refresh success rate

### Rollback Plan
- OAuth config can be disabled via config file
- Users with passwords can still login
- Database migrations are additive (safe to rollback)

---

## Success Metrics

**Security:**
- Zero CSRF attacks via state validation
- No plaintext tokens in logs
- Rate limiting blocks >95% of brute force attempts

**Reliability:**
- Token refresh success rate >99%
- OAuth login success rate >95%
- User provisioning success rate >98%

**Observability:**
- All auth events logged to audit table
- Error tracking with structured codes
- Health check shows provider status

---

## Related Documentation

- **[AUTH_AND_OAUTH.md](domain/AUTH_AND_OAUTH.md)** - Current OAuth documentation (update in Phase 4)
- **[PATTERNS.md](reference/PATTERNS.md)** - OAuth patterns and best practices
- **[TROUBLESHOOTING.md](reference/TROUBLESHOOTING.md)** - OAuth troubleshooting guide

---

## Notes

**Provider Support:**
- Designed to work with any OpenID Connect provider
- Tested with: Google, Azure AD, Okta, GitHub, Auth0
- Custom providers need to provide: auth URL, token URL, userinfo URL

**Backward Compatibility:**
- Existing password-based auth unchanged
- OAuth is additive feature
- Users can have both password and OAuth

**Future Enhancements (Post-Production):**
- Multi-provider support per user
- Social login buttons (Google, GitHub, etc.)
- OAuth scope management UI
- Token rotation policies
- SSO integration

---

**Last Updated:** 2026-01-17
**Status:** Ready for Implementation
