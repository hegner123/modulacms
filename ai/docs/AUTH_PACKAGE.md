# AUTH_PACKAGE.md

Package guide for authentication primitives in ModulaCMS.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/AUTH_PACKAGE.md`
**Purpose:** Implementation details for password hashing, OAuth state management, and authentication utilities
**Last Updated:** 2026-01-12

---

## Overview

The `internal/auth/` package provides low-level authentication utilities used throughout ModulaCMS:
- **Password Hashing:** bcrypt implementation with cost factor 12
- **OAuth State Generation:** CSRF protection for OAuth flows
- **Hash Comparison:** Constant-time comparison to prevent timing attacks

**Important:** This package provides **primitives only**. Full authentication flows are handled by:
- `internal/middleware/` - Session validation and request authentication
- `internal/router/auth.go` - OAuth callback handling
- `internal/router/userOauth.go` - OAuth token CRUD operations
- `internal/router/sessions.go` - Session management
- `internal/router/tokens.go` - API token management

---

## Package Structure

**Files:**
```
internal/auth/
├── auth.go          # Core authentication utilities
└── config.json      # Example OAuth configuration (should not be in production)
```

**Key Functions:**
- `HashPassword(password string) (string, error)`
- `compareHashes(hash1, hash2 string) bool`
- `writeStateOauthCookie(w http.ResponseWriter) string`
- `generateStateOauthCookie() (*http.Cookie, string)`

---

## Password Hashing

**File:** `internal/auth/auth.go:14-19`

```go
// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	// Use cost of 12 (which is a good balance of security and performance)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}
```

**Cost Factor:** 12
- Balances security and performance
- ~0.2-0.3 seconds per hash on modern hardware
- Resistant to brute force attacks

**Usage Pattern:**
```go
// When creating a new user
hashedPassword, err := auth.HashPassword(plainPassword)
if err != nil {
	return fmt.Errorf("failed to hash password: %w", err)
}

user := db.CreateUserParams{
	Username: username,
	Email:    email,
	Hash:     hashedPassword,
	// ... other fields
}
```

**Stored In:** `users.hash` column (TEXT)

**Validation:** Use `bcrypt.CompareHashAndPassword()` in application code:
```go
err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedPassword))
if err != nil {
	return fmt.Errorf("invalid password")
}
```

---

## Hash Comparison (Constant-Time)

**File:** `internal/auth/auth.go:21-29`

```go
func compareHashes(hash1, hash2 string) bool {
	hash1Bytes, err1 := hex.DecodeString(hash1)
	hash2Bytes, err2 := hex.DecodeString(hash2)

	if err1 != nil || err2 != nil || len(hash1Bytes) != len(hash2Bytes) {
		return false
	}
	return subtle.ConstantTimeCompare(hash1Bytes, hash2Bytes) == 1
}
```

**Purpose:** Prevent timing attacks when comparing sensitive values
- Uses `crypto/subtle.ConstantTimeCompare()`
- Takes same time regardless of where mismatch occurs
- Protects against side-channel attacks

**Note:** This function is currently unexported. For session validation, see `internal/middleware/session.go:39` which uses direct string comparison. Consider using constant-time comparison for session tokens in production.

---

## OAuth State Cookie Generation

**File:** `internal/auth/auth.go:31-47`

### writeStateOauthCookie

```go
func writeStateOauthCookie(w http.ResponseWriter) string {
	cookie, state := generateStateOauthCookie()
	http.SetCookie(w, cookie)
	return state
}
```

**Purpose:** Generate and set OAuth state cookie in one operation
- Returns state string for validation
- Sets cookie in HTTP response

### generateStateOauthCookie

```go
func generateStateOauthCookie() (*http.Cookie, string) {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{
		Name:    "oauthstate",
		Value:   state,
		Expires: expiration,
	}

	return &cookie, state
}
```

**State Parameters:**
- **Entropy:** 16 random bytes (128 bits)
- **Encoding:** Base64 URL-safe encoding
- **Expiration:** 20 minutes
- **Cookie Name:** `"oauthstate"`

**Security Features:**
- Cryptographically secure random via `crypto/rand`
- Short expiration window (20 minutes)
- Used for CSRF protection in OAuth flow

---

## OAuth Implementation Details

### Configuration Structure

**File:** `internal/config/config.go:48-51`

```go
type Config struct {
	Oauth_Client_Id     string              `json:"oauth_client_id"`
	Oauth_Client_Secret string              `json:"oauth_client_secret"`
	Oauth_Scopes        []string            `json:"oauth_scopes"`
	Oauth_Endpoint      map[Endpoint]string `json:"oauth_endpoint"`
}
```

**Endpoint Constants:**
```go
type Endpoint string

const (
	OauthAuthURL  Endpoint = "oauth_auth_url"
	OauthTokenURL Endpoint = "oauth_token_url"
)
```

### Example Configuration

**File:** `internal/auth/config.json` (example only, not for production)

```json
{
	"oauth_client_id": "Ov23liFoy8pVGnAnGgrE",
	"oauth_client_secret": "f57dda6a58faa59e4803f08efca11362478dcd3c",
	"oauth_scopes": ["user"],
	"oauth_endpoint": {
		"oauth_auth_url": "https://github.com/login/oauth/authorize",
		"oauth_token_url": "https://github.com/login/oauth2/token"
	}
}
```

---

## OAuth Callback Handler

**File:** `internal/router/auth.go:45-92`

```go
func OauthCallbackHandler(c config.Config, verifier string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 1. Retrieve authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			return
		}

		// 2. Build OAuth2 configuration
		conf := &oauth2.Config{
			ClientID:     c.Oauth_Client_Id,
			ClientSecret: c.Oauth_Client_Secret,
			Scopes:       c.Oauth_Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  c.Oauth_Endpoint[config.OauthAuthURL],
				TokenURL: c.Oauth_Endpoint[config.OauthTokenURL],
			},
		}

		// 3. Exchange code for token with PKCE verifier
		token, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error exchanging token: %v", err), http.StatusInternalServerError)
			return
		}

		// 4. Use token to create authenticated client
		client := conf.Client(ctx, token)

		// 5. Fetch user info from provider
		resp, err := client.Get("https://api.example.com/userinfo")
		// ... handle response
	}
}
```

**Flow:**
1. Extract authorization code from callback URL
2. Build `oauth2.Config` from application config
3. Exchange code for access token (with PKCE verifier)
4. Create authenticated HTTP client
5. Fetch user info from OAuth provider
6. Store tokens in `user_oauth` table

**PKCE Support:** Uses `oauth2.VerifierOption(verifier)` for enhanced security

---

## Integration with Middleware

### Session Validation

**File:** `internal/middleware/session.go:23-56`

The middleware validates sessions by:
1. Extracting cookie from request
2. Decoding Base64 JSON cookie value
3. Querying session from database
4. Comparing session tokens
5. Checking expiration timestamp
6. Returning authenticated user

**Flow:**
```go
func UserIsAuth(r *http.Request, cookie *http.Cookie, c *config.Config) (*db.Users, error) {
	// 1. Decode cookie
	userCookie, err := ReadCookie(cookie)
	if err != nil {
		return nil, err
	}

	// 2. Get session from database
	dbc := db.ConfigDB(*c)
	session, err := dbc.GetSessionByUserId(userCookie.UserId)
	if err != nil || session == nil {
		return nil, err
	}

	// 3. Validate session data
	if userCookie.Session != session.SessionData.String {
		return nil, fmt.Errorf("sessions don't match")
	}

	// 4. Check expiration
	expired := utility.TimestampLessThan(session.ExpiresAt.String)
	if expired {
		return nil, fmt.Errorf("session is expired")
	}

	// 5. Return authenticated user
	return dbc.GetUser(userCookie.UserId)
}
```

### Cookie Structure

**File:** `internal/middleware/cookies.go:14-17`

```go
type MiddlewareCookie struct {
	Session string `json:"session"`
	UserId  int64  `json:"userId"`
}
```

**Encoding:** Base64-encoded JSON
```go
// Reading a cookie
func ReadCookie(c *http.Cookie) (*MiddlewareCookie, error) {
	k := MiddlewareCookie{}

	// Validate cookie
	err := c.Valid()
	if err != nil {
		return nil, err
	}

	// Base64 decode
	cv := c.Value
	b, err := base64.StdEncoding.DecodeString(cv)
	if err != nil {
		return nil, err
	}

	// JSON unmarshal
	err = json.Unmarshal(b, &k)
	if err != nil {
		return nil, err
	}

	return &k, nil
}
```

### Main Middleware Chain

**File:** `internal/middleware/middleware.go:25-50`

```go
func Serve(next http.Handler, c *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Cors(w, r, c)

		u, user := AuthRequest(w, r, c)
		if u != nil {
			// Inject authenticated user into context
			ctx := context.WithValue(r.Context(), u, user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Block unauthenticated API requests
		if strings.Contains(r.URL.Path, "api") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

**Authentication Flow:**
1. Apply CORS headers
2. Attempt authentication via `AuthRequest()`
3. If authenticated: inject user into context
4. If unauthenticated + API route: return 401
5. Otherwise: proceed to next handler

---

## OAuth Token Management

### UserOauth CRUD Operations

**File:** `internal/router/userOauth.go`

#### Create OAuth Connection
```go
POST /api/v1/user-oauth

Body:
{
	"user_id": 123,
	"oauth_provider": "github",
	"oauth_provider_user_id": "12345678",
	"access_token": "gho_...",
	"refresh_token": "ghr_...",
	"token_expires_at": "2026-01-13T12:00:00Z"
}
```

**Implementation:**
```go
func ApiCreateUserOauth(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	var newUserOauth db.CreateUserOauthParams
	err = json.NewDecoder(r.Body).Decode(&newUserOauth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	createdUserOauth, err := d.CreateUserOauth(newUserOauth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUserOauth)
	return nil
}
```

#### Update OAuth Tokens
```go
PUT /api/v1/user-oauth/:id

Body:
{
	"user_oauth_id": 1,
	"access_token": "new_token",
	"refresh_token": "new_refresh",
	"token_expires_at": "2026-01-14T12:00:00Z"
}
```

#### Delete OAuth Connection
```go
DELETE /api/v1/user-oauth/:id?q=1
```

---

## Session Management

**File:** `internal/router/sessions.go`

### Update Session
```go
PUT /api/v1/sessions/:id

Body:
{
	"session_id": 123,
	"expires_at": "2026-01-13T12:00:00Z",
	"last_access": "2026-01-12T14:30:00Z"
}
```

### Logout (Delete Session)
```go
DELETE /api/v1/sessions/:id?q=123
```

**Implementation:**
```go
func apiDeleteSession(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	q := r.URL.Query().Get("q")
	sID, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	err = d.DeleteSession(sID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
```

---

## API Token Management

**File:** `internal/router/tokens.go`

### Token Operations

**Create Token:**
```go
POST /api/v1/tokens

Body:
{
	"user_id": 123,
	"token_type": "api_access",
	"token": "tok_...",
	"issued_at": "2026-01-12T12:00:00Z",
	"expires_at": "2027-01-12T12:00:00Z",
	"revoked": false
}
```

**Get Token:**
```go
GET /api/v1/tokens/:id?q=123
```

**Update Token (Revoke):**
```go
PUT /api/v1/tokens/:id

Body:
{
	"id": 123,
	"revoked": true
}
```

**Delete Token:**
```go
DELETE /api/v1/tokens/:id?q=123
```

---

## Adding a New OAuth Provider

### Step 1: Register OAuth Application

Register your application with the OAuth provider:
- **GitHub:** https://github.com/settings/developers
- **Google:** https://console.cloud.google.com/apis/credentials
- **Azure AD:** https://portal.azure.com/ → App registrations
- **Okta:** Your Okta admin panel

Obtain:
- Client ID
- Client Secret
- Authorization URL
- Token URL
- Required scopes

### Step 2: Update Configuration

**File:** `config.json` or environment variables

```json
{
	"oauth_client_id": "your-new-provider-client-id",
	"oauth_client_secret": "your-new-provider-client-secret",
	"oauth_scopes": ["openid", "profile", "email"],
	"oauth_endpoint": {
		"oauth_auth_url": "https://provider.example.com/oauth2/authorize",
		"oauth_token_url": "https://provider.example.com/oauth2/token"
	}
}
```

### Step 3: Configure Callback Route

**File:** `cmd/main.go` or router initialization

```go
// Assuming you have a verifier from PKCE generation
verifier := "your-pkce-verifier"

mux.HandleFunc("/api/v1/auth/oauth/callback",
	router.OauthCallbackHandler(configuration, verifier))
```

### Step 4: Customize User Info Endpoint

**File:** `internal/router/auth.go:82-87`

Modify the user info fetch to match your provider's API:

```go
// For GitHub
resp, err := client.Get("https://api.github.com/user")

// For Google
resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")

// For Azure AD
resp, err := client.Get("https://graph.microsoft.com/v1.0/me")

// For Okta
resp, err := client.Get("https://your-domain.okta.com/oauth2/v1/userinfo")
```

### Step 5: Map Provider User ID

Parse the response and extract the provider's user ID:

```go
type GitHubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

var ghUser GitHubUser
json.NewDecoder(resp.Body).Decode(&ghUser)

// Store in user_oauth table
userOauth := db.CreateUserOauthParams{
	UserID:              existingUserId,
	OauthProvider:       "github",
	OauthProviderUserID: strconv.FormatInt(ghUser.ID, 10),
	AccessToken:         token.AccessToken,
	RefreshToken:        token.RefreshToken,
	TokenExpiresAt:      token.Expiry.Format(time.RFC3339),
}
```

---

## Security Considerations

### Implemented

1. **bcrypt Cost Factor 12:** Resistant to brute force
2. **PKCE for OAuth:** Enhanced security for authorization code flow
3. **Constant-Time Comparison:** Available via `compareHashes()` (not currently used)
4. **Secure Random State:** 128-bit entropy for OAuth state
5. **Short-Lived State Cookies:** 20-minute expiration
6. **Database-Backed Sessions:** Centralized validation and revocation

### Recommended Improvements

1. **Use Constant-Time Comparison for Sessions:**
   ```go
   // Instead of:
   if userCookie.Session != session.SessionData.String {

   // Use:
   if !auth.CompareHashes(userCookie.Session, session.SessionData.String) {
   ```

2. **Add Secure Cookie Flags:**
   ```go
   cookie := http.Cookie{
   	Name:     c.Cookie_Name,
   	Value:    encodedValue,
   	Expires:  expiration,
   	HttpOnly: true,        // Prevent JavaScript access
   	Secure:   true,        // HTTPS only
   	SameSite: http.SameSiteStrictMode, // CSRF protection
   	Path:     "/",
   }
   ```

3. **Implement Token Refresh Logic:**
   - Check `token_expires_at` before API calls
   - Automatically refresh with `refresh_token`
   - Update `user_oauth` table with new tokens

4. **Add Rate Limiting:**
   - Login attempts per IP
   - OAuth callback requests
   - Session validation checks

5. **Session Cleanup:**
   - Periodic job to delete expired sessions
   - Implement in `internal/maintenance/` or cron

6. **State Validation:**
   - Currently commented out in `auth.go:57-58`
   - Validate state parameter against stored cookie
   - Prevent CSRF attacks on OAuth flow

---

## Database Schema Integration

### Users Table
**Schema:** `sql/schema/4_users/schema.sql`
- Stores bcrypt hashed passwords in `hash` column

### Sessions Table
**Schema:** `sql/schema/15_sessions/schema.sql`
- Tracks active user sessions
- Validates against cookie values

### User OAuth Table
**Schema:** `sql/schema/12_user_oauth/schema.sql`
- Stores OAuth provider tokens
- Links users to external OAuth accounts

### Tokens Table
**Schema:** `sql/schema/11_tokens/schema.sql`
- Manages API access tokens
- Tracks revocation status

See **[AUTH_AND_OAUTH.md](../domain/AUTH_AND_OAUTH.md)** for complete schema details.

---

## Testing Authentication

### Unit Test Password Hashing
```go
func TestHashPassword(t *testing.T) {
	password := "testPassword123"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify hash can be validated
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("Hash verification failed: %v", err)
	}

	// Verify wrong password fails
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrongPassword"))
	if err == nil {
		t.Error("Wrong password should not validate")
	}
}
```

### Test OAuth State Generation
```go
func TestGenerateStateOauthCookie(t *testing.T) {
	cookie, state := auth.GenerateStateOauthCookie()

	if cookie.Name != "oauthstate" {
		t.Errorf("Expected cookie name 'oauthstate', got '%s'", cookie.Name)
	}

	if cookie.Value != state {
		t.Error("Cookie value should match returned state")
	}

	if len(state) < 20 {
		t.Error("State should be at least 20 characters")
	}

	// Check expiration is set
	if cookie.Expires.IsZero() {
		t.Error("Cookie expiration not set")
	}
}
```

### Integration Test Session Validation
```go
func TestUserIsAuth(t *testing.T) {
	// Setup test database and config
	c := config.Config{/* test config */}

	// Create test user and session
	user := createTestUser(t, c)
	session := createTestSession(t, c, user.UserId)

	// Create test cookie
	cookieData := middleware.MiddlewareCookie{
		Session: session.SessionData.String,
		UserId:  user.UserId,
	}
	cookieJSON, _ := json.Marshal(cookieData)
	encodedCookie := base64.StdEncoding.EncodeToString(cookieJSON)

	cookie := &http.Cookie{
		Name:  c.Cookie_Name,
		Value: encodedCookie,
	}

	// Test authentication
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.AddCookie(cookie)

	authUser, err := middleware.UserIsAuth(req, cookie, &c)
	if err != nil {
		t.Fatalf("UserIsAuth failed: %v", err)
	}

	if authUser.UserId != user.UserId {
		t.Errorf("Expected user ID %d, got %d", user.UserId, authUser.UserId)
	}
}
```

---

## Common Patterns

### Creating a New User with Password
```go
import (
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/db"
)

func createUserWithPassword(username, email, plainPassword string, c config.Config) (*db.Users, error) {
	// Hash password
	hashedPassword, err := auth.HashPassword(plainPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	d := db.ConfigDB(c)
	user, err := d.CreateUser(db.CreateUserParams{
		Username: username,
		Email:    email,
		Hash:     hashedPassword,
		Role:     4, // Default role
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
```

### Validating User Login
```go
import "golang.org/x/crypto/bcrypt"

func validateLogin(email, password string, c config.Config) (*db.Users, error) {
	d := db.ConfigDB(c)

	// Get user by email
	user, err := d.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Compare password hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}
```

### Creating Session After Login
```go
import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

func createSession(user *db.Users, w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	// Generate session token
	b := make([]byte, 32)
	rand.Read(b)
	sessionToken := base64.URLEncoding.EncodeToString(b)

	// Create session in database
	session, err := d.CreateSession(db.CreateSessionParams{
		UserID:      user.UserId,
		SessionData: sessionToken,
		ExpiresAt:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create cookie
	cookieData := middleware.MiddlewareCookie{
		Session: sessionToken,
		UserId:  user.UserId,
	}
	cookieJSON, _ := json.Marshal(cookieData)
	encodedCookie := base64.StdEncoding.EncodeToString(cookieJSON)

	cookie := &http.Cookie{
		Name:     c.Cookie_Name,
		Value:    encodedCookie,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}

	http.SetCookie(w, cookie)
	return nil
}
```

---

## Related Documentation

- **[AUTH_AND_OAUTH.md](../domain/AUTH_AND_OAUTH.md)** - High-level authentication flows and domain concepts
- **[MIDDLEWARE_PACKAGE.md](MIDDLEWARE_PACKAGE.md)** - Middleware integration and request handling
- **[DB_PACKAGE.md](DB_PACKAGE.md)** - Database operations for users, sessions, and tokens
- **[CLAUDE.md](../../CLAUDE.md)** - Security considerations and configuration

---

## Quick Reference

**Key Files:**
- `internal/auth/auth.go` - Core authentication utilities
- `internal/middleware/session.go` - Session validation logic
- `internal/middleware/cookies.go` - Cookie encoding/decoding
- `internal/router/auth.go` - OAuth callback handler
- `internal/router/userOauth.go` - OAuth token CRUD
- `internal/router/sessions.go` - Session management
- `internal/router/tokens.go` - API token management

**Key Functions:**
- `auth.HashPassword()` - bcrypt hash with cost 12
- `auth.generateStateOauthCookie()` - OAuth state with 20min expiration
- `middleware.UserIsAuth()` - Complete session validation
- `middleware.ReadCookie()` - Base64 JSON cookie decoding
- `router.OauthCallbackHandler()` - Exchange code for OAuth tokens

**Security:**
- bcrypt cost: 12
- OAuth state: 128-bit entropy, 20min expiration
- PKCE: Supported via `oauth2.VerifierOption()`
- Constant-time comparison: Available but not used for sessions

**Recommended:**
- Add secure cookie flags (HttpOnly, Secure, SameSite)
- Implement token refresh logic
- Add rate limiting on auth endpoints
- Enable state parameter validation
- Regular session cleanup job
