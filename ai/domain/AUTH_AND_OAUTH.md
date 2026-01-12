# AUTH_AND_OAUTH.md

Domain guide for authentication and OAuth implementation.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/AUTH_AND_OAUTH.md`
**Purpose:** Authentication flows, OAuth integration, and session management
**Last Updated:** 2026-01-12

---

## Overview

ModulaCMS supports multiple authentication methods:
- **HTTP API:** Cookie-based session authentication
- **OAuth2:** Multi-provider with PKCE support
- **SSH:** Public key authentication for CLI access
- **Role-Based Access Control (RBAC):** Permissions and roles

---

## Password Hashing

**File:** `internal/auth/auth.go`

```go
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}
```

**Key Points:**
- bcrypt with cost factor 12
- CPU-intensive, resistant to brute force
- Hash stored in `users.hash` column

---

## Session Management

**Schema** (`sql/schema/15_sessions/schema.sql`):
```sql
CREATE TABLE sessions (
	session_id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users ON DELETE CASCADE,
	created_at TEXT DEFAULT CURRENT_TIMESTAMP,
	expires_at TEXT,
	last_access TEXT DEFAULT CURRENT_TIMESTAMP,
	ip_address TEXT,
	user_agent TEXT,
	session_data TEXT
);
```

**Cookie Structure** (`internal/middleware/cookies.go`):
```go
type MiddlewareCookie struct {
	Session string `json:"session"`
	UserId  int64  `json:"userId"`
}
```

**Value:** Base64-encoded JSON

**Authentication Check** (`internal/middleware/session.go`):
```go
func UserIsAuth(r *http.Request, cookie *http.Cookie, c *config.Config) (*db.Users, error) {
	// Decode cookie
	userCookie, err := ReadCookie(cookie)
	if err != nil {
		return nil, err
	}

	// Retrieve session from database
	dbc := db.ConfigDB(*c)
	session, err := dbc.GetSessionByUserId(userCookie.UserId)
	if err != nil {
		return nil, err
	}

	// Validate session data matches cookie
	if userCookie.Session != session.SessionData.String {
		return nil, fmt.Errorf("sessions don't match")
	}

	// Check expiration
	expired := utility.TimestampLessThan(session.ExpiresAt.String)
	if expired {
		return nil, fmt.Errorf("session is expired")
	}

	// Return authenticated user
	return dbc.GetUser(userCookie.UserId)
}
```

**Validation Flow:**
1. Extract cookie from request
2. Base64 decode and unmarshal JSON
3. Query database for session by user_id
4. Compare session data
5. Check expiration timestamp
6. Return user if valid

---

## OAuth2 Implementation

**Configuration** (`internal/config/config.go`):
```go
type Config struct {
	Oauth_Client_Id     string              `json:"oauth_client_id"`
	Oauth_Client_Secret string              `json:"oauth_client_secret"`
	Oauth_Scopes        []string            `json:"oauth_scopes"`
	Oauth_Endpoint      map[Endpoint]string `json:"oauth_endpoint"`
}

type Endpoint string
const (
	OauthAuthURL  Endpoint = "oauth_auth_url"
	OauthTokenURL Endpoint = "oauth_token_url"
)
```

**Callback Handler** (`internal/router/auth.go`):
```go
func OauthCallbackHandler(c config.Config, verifier string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			return
		}

		// Build OAuth2 config
		conf := &oauth2.Config{
			ClientID:     c.Oauth_Client_Id,
			ClientSecret: c.Oauth_Client_Secret,
			Scopes:       c.Oauth_Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  c.Oauth_Endpoint[config.OauthAuthURL],
				TokenURL: c.Oauth_Endpoint[config.OauthTokenURL],
			},
		}

		// Exchange code for token with PKCE verifier
		token, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error exchanging token: %v", err), http.StatusInternalServerError)
			return
		}

		// Use token to create authenticated client
		client := conf.Client(ctx, token)
		// Fetch user info from provider...
	}
}
```

**PKCE State Management** (`internal/auth/auth.go`):
```go
func generateStateOauthCookie() (*http.Cookie, string) {
	expiration := time.Now().Add(20 * time.Minute)
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

**Supported Features:**
- Authorization Code Flow
- PKCE (Proof Key for Code Exchange)
- State parameter validation
- Token refresh storage
- Multi-provider support

---

## OAuth Token Storage

**Schema** (`sql/schema/12_user_oauth/schema.sql`):
```sql
CREATE TABLE IF NOT EXISTS user_oauth (
	user_oauth_id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users ON DELETE CASCADE,
	oauth_provider TEXT NOT NULL,
	oauth_provider_user_id TEXT NOT NULL,
	access_token TEXT NOT NULL,
	refresh_token TEXT NOT NULL,
	token_expires_at TEXT NOT NULL,
	date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Go Model:**
```go
type UserOauth struct {
	UserOauthID         int64  `json:"user_oauth_id"`
	UserID              int64  `json:"user_id"`
	OauthProvider       string `json:"oauth_provider"`
	OauthProviderUserID string `json:"oauth_provider_user_id"`
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	TokenExpiresAt      string `json:"token_expires_at"`
	DateCreated         string `json:"date_created"`
}
```

---

## SSH Key Authentication

**SSH Server** (`cmd/main.go`):
```go
sshServer, err := wish.NewServer(
	wish.WithAddress(net.JoinHostPort(host, configuration.SSH_Port)),
	wish.WithHostKeyPath(".ssh/id_ed25519"),
	wish.WithMiddleware(
		cli.CliMiddleware(app.VerboseFlag, configuration),
		logging.Middleware(),
	),
)
```

**Features:**
- Charmbracelet Wish SSH server
- Ed25519 host key
- Public key authentication via SSH protocol
- Launches Bubbletea TUI on connection

**CLI Middleware** (`internal/cli/middleware.go`):
```go
func CliMiddleware(v *bool, c *config.Config) wish.Middleware {
	teaHandler := func(s ssh.Session) *tea.Program {
		pty, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal")
			return nil
		}
		m, _ := InitialModel(v, c)
		m.Term = pty.Term
		m.Width = pty.Window.Width
		m.Height = pty.Window.Height
		return tea.NewProgram(&m, bubbletea.MakeOptions(s)...)
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
```

**SSH Flow:**
1. Client initiates connection with public key
2. Wish verifies against authorized keys
3. PTY (pseudo-terminal) established
4. Bubbletea TUI launched in session
5. CLI management interface available

---

## Role-Based Access Control

**Permissions Schema** (`sql/schema/1_permissions/schema.sql`):
```sql
CREATE TABLE IF NOT EXISTS permissions (
	permission_id INTEGER PRIMARY KEY,
	table_id INTEGER NOT NULL,
	mode INTEGER NOT NULL,
	label TEXT NOT NULL
);
```

**Modes:**
- READ
- CREATE
- UPDATE
- DELETE

**Roles Schema** (`sql/schema/2_roles/schema.sql`):
```sql
CREATE TABLE IF NOT EXISTS roles (
	role_id INTEGER PRIMARY KEY,
	label TEXT NOT NULL UNIQUE,
	permissions TEXT NOT NULL UNIQUE
);
```

**Users Schema** (`sql/schema/4_users/schema.sql`):
```sql
CREATE TABLE IF NOT EXISTS users (
	user_id INTEGER PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	email TEXT NOT NULL,
	hash TEXT NOT NULL,
	role INTEGER NOT NULL DEFAULT 4 REFERENCES roles ON DELETE SET DEFAULT,
	date_created TEXT DEFAULT CURRENT_TIMESTAMP,
	date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

**Permission Structure** (`sql/permissions.json`):
```json
{
	"create": [
		{ "admin_datatypes": true },
		{ "content_data": true }
	],
	"read": [
		{ "admin_datatypes": true },
		{ "content_data": true }
	],
	"update": [
		{ "admin_datatypes": true }
	],
	"delete": [
		{ "admin_datatypes": true }
	]
}
```

**Relationships:**
- Users → Roles (via role_id)
- Roles store permissions as JSON
- Default role: ID 4
- Cascade: ON DELETE SET DEFAULT

---

## Token Management

**Schema** (`sql/schema/11_tokens/schema.sql`):
```sql
CREATE TABLE IF NOT EXISTS tokens (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users ON DELETE CASCADE,
	token_type TEXT NOT NULL,
	token TEXT NOT NULL UNIQUE,
	issued_at TEXT NOT NULL,
	expires_at TEXT NOT NULL,
	revoked BOOLEAN NOT NULL DEFAULT 0
);
```

**Token Types:**
- API access tokens
- OAuth refresh tokens
- Session tokens
- Password reset tokens

**Operations:**
```sql
-- Create token
INSERT INTO tokens (user_id, token_type, token, issued_at, expires_at, revoked)
VALUES (?, ?, ?, ?, ?, ?)

-- Revoke token
UPDATE tokens SET revoked = 1 WHERE id = ?

-- Get valid tokens
SELECT * FROM tokens
WHERE user_id = ? AND revoked = 0 AND expires_at > CURRENT_TIMESTAMP
```

---

## Middleware Authentication

**Main Middleware** (`internal/middleware/middleware.go`):
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

**Authentication Chain:**
1. Apply CORS headers
2. Extract cookie
3. Validate cookie structure
4. Query session from database
5. Compare session tokens
6. Check expiration
7. Inject user into context
8. Block API requests if unauthenticated

---

## API Endpoints

**Authentication:**
```
POST /api/v1/auth/register      # User registration
GET  /api/v1/auth/oauth        # OAuth callback
POST /api/v1/auth/reset        # Password reset
```

**Users:**
```
GET    /api/v1/users           # List users
POST   /api/v1/users           # Create user
GET    /api/v1/users/:id       # Get user
PUT    /api/v1/users/:id       # Update user
DELETE /api/v1/users/:id       # Delete user
```

**OAuth Tokens:**
```
POST   /api/v1/user-oauth      # Create OAuth link
PUT    /api/v1/user-oauth/:id  # Update tokens
DELETE /api/v1/user-oauth/:id  # Delete OAuth link
```

**Sessions:**
```
PUT    /api/v1/sessions/:id    # Update session
DELETE /api/v1/sessions/:id    # Logout
```

---

## Provider Configuration

**Example: Generic OAuth** (config.json):
```json
{
	"oauth_client_id": "your-client-id",
	"oauth_client_secret": "your-client-secret",
	"oauth_scopes": ["openid", "profile", "email"],
	"oauth_endpoint": {
		"oauth_auth_url": "https://provider.example.com/authorize",
		"oauth_token_url": "https://provider.example.com/token"
	}
}
```

**Example: Azure AD**:
```json
{
	"oauth_client_id": "azure-app-id",
	"oauth_client_secret": "azure-app-secret",
	"oauth_scopes": ["openid", "profile", "email", "offline_access"],
	"oauth_endpoint": {
		"oauth_auth_url": "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		"oauth_token_url": "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	}
}
```

**Supported Providers:**
- Any OpenID Connect compatible provider
- Azure AD, Okta, Google, GitHub, etc.
- Configurable authorization/token endpoints
- Customizable scope requests

---

## Schema Relationships

```
users
├── user_id (PK)
├── role (FK → roles.role_id)
└── hash (bcrypt password)

roles
├── role_id (PK)
└── permissions (JSON)

sessions
├── session_id (PK)
├── user_id (FK → users.user_id)
├── session_data (token)
└── expires_at

user_oauth
├── user_oauth_id (PK)
├── user_id (FK → users.user_id)
├── oauth_provider
├── access_token
└── refresh_token

tokens
├── id (PK)
├── user_id (FK → users.user_id)
├── token (unique)
├── expires_at
└── revoked (boolean)
```

---

## Security Considerations

**Implemented:**
- bcrypt cost factor 12
- PKCE for OAuth
- Session expiration tracking
- Database-backed session validation
- Parameterized SQL queries (via sqlc)
- Ed25519 SSH keys

**Recommended:**
- Enable HTTPS (Let's Encrypt built-in)
- Rotate SSH keys periodically
- Implement token refresh logic
- Add rate limiting on auth endpoints
- Regular session cleanup
- Secure cookie flags (SameSite, Secure)

---

## Related Documentation

- **[CLAUDE.md](../CLAUDE.md)** - Configuration management
- **[CONTENT_MODEL.md](../architecture/CONTENT_MODEL.md)** - User relationships in content

---

## Quick Reference

**Key Files:**
- `internal/auth/auth.go` - Password hashing, OAuth state
- `internal/middleware/session.go` - Session validation
- `internal/middleware/cookies.go` - Cookie handling
- `internal/router/auth.go` - OAuth callback
- `cmd/main.go` - SSH server setup
- `sql/schema/4_users/` - Users table
- `sql/schema/15_sessions/` - Sessions table
- `sql/schema/12_user_oauth/` - OAuth tokens

**Key Operations:**
- Hash password: bcrypt cost 12
- Validate session: Cookie → DB → Expiration check
- OAuth flow: Authorize → Code → Exchange → Token
- SSH auth: Public key via Wish server
- RBAC: Users → Roles → Permissions (JSON)
