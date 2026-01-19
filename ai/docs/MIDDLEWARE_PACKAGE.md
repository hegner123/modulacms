# MIDDLEWARE_PACKAGE.md

HTTP middleware implementation for ModulaCMS request processing.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware`
**Purpose:** Provide HTTP middleware components for authentication, CORS handling, session management, and cookie operations.
**Last Updated:** 2026-01-12

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Core Components](#core-components)
4. [Middleware Chain](#middleware-chain)
5. [Authentication Flow](#authentication-flow)
6. [CORS Configuration](#cors-configuration)
7. [Cookie Management](#cookie-management)
8. [Session Validation](#session-validation)
9. [Adding New Middleware](#adding-new-middleware)
10. [Testing](#testing)
11. [Configuration](#configuration)
12. [Common Patterns](#common-patterns)
13. [Related Documentation](#related-documentation)
14. [Quick Reference](#quick-reference)

---

## Overview

The middleware package provides HTTP request pre-processing for the ModulaCMS web application. It implements a composable middleware chain that handles:

- **Authentication**: Cookie-based session authentication with database validation
- **CORS**: Cross-Origin Resource Sharing with configurable origins, methods, and headers
- **Cookie Operations**: Base64-encoded JSON cookie reading and writing
- **Session Management**: Session validation with expiration checking
- **API Protection**: Blocks unauthorized access to `/api/*` endpoints

**Design Philosophy:**
- Single middleware chain wraps the entire application router
- Configuration-driven behavior (no hardcoded values)
- Context-based user injection for authenticated requests
- Early returns for performance (failed auth doesn't reach handlers)

---

## Architecture

### Middleware Flow Diagram

```
HTTP Request
    ↓
middleware.Serve()
    ↓
Cors() → Set CORS headers
    ↓
AuthRequest() → Extract & validate cookie
    ↓
    ├─ Cookie valid? → UserIsAuth()
    │   ↓
    │   ├─ Session valid? → Inject user into context → next.ServeHTTP()
    │   └─ Session invalid? → Check if /api route
    │       ↓
    │       ├─ /api route? → 401 Unauthorized
    │       └─ Not /api? → next.ServeHTTP() (public route)
    └─ No cookie? → Check if /api route
        ↓
        ├─ /api route? → 401 Unauthorized
        └─ Not /api? → next.ServeHTTP() (public route)
```

### Package Structure

```
internal/middleware/
├── middleware.go     # Main middleware chain and authentication
├── cors.go          # CORS header management
├── cookies.go       # Cookie reading and writing utilities
├── session.go       # Session validation and user authentication
└── tests/
    └── cors_test.go # CORS middleware tests
```

---

## Core Components

### 1. middleware.go

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/middleware.go`

Main middleware entry point. Implements the `Serve` function that wraps the application router.

**Key Functions:**

```go
// Serve is the main middleware handler
func Serve(next http.Handler, c *config.Config) http.Handler

// AuthRequest extracts and validates authentication
func AuthRequest(w http.ResponseWriter, r *http.Request, c *config.Config) (*authcontext, *db.Users)

// GetURLSegments splits URL path into segments
func GetURLSegments(path string) []string
```

**Authentication Context Type:**

```go
type authcontext string
```

Used as a key to store authenticated user data in request context:

```go
var u authcontext = "authenticated"
ctx := context.WithValue(r.Context(), u, user)
```

---

### 2. cors.go

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/cors.go`

CORS (Cross-Origin Resource Sharing) handling with configuration-driven behavior.

**Key Functions:**

```go
// Cors sets CORS headers based on configuration
func Cors(w http.ResponseWriter, r *http.Request, c *config.Config)

// CorsWithConfig allows specifying a config for CORS settings
func CorsWithConfig(w http.ResponseWriter, r *http.Request, c *config.Config)

// getAllowedOrigin returns the allowed origin if permitted
func getAllowedOrigin(requestOrigin string, allowedOrigins []string) string
```

**CORS Headers Set:**
- `Access-Control-Allow-Origin`: Allowed origin (exact match or `*`)
- `Access-Control-Allow-Methods`: Allowed HTTP methods (GET, POST, PUT, DELETE, OPTIONS)
- `Access-Control-Allow-Headers`: Allowed request headers (Content-Type, Authorization, etc.)
- `Access-Control-Allow-Credentials`: Whether credentials are allowed (`true` or not set)

**Preflight Handling:**
- OPTIONS requests receive 204 No Content response after setting CORS headers

---

### 3. cookies.go

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/cookies.go`

Cookie reading and writing utilities using base64-encoded JSON.

**Key Functions:**

```go
// SetCookieHandler sets a cookie in the HTTP response
func SetCookieHandler(w http.ResponseWriter, c *http.Cookie)

// ReadCookie decodes and deserializes a cookie value
func ReadCookie(c *http.Cookie) (*MiddlewareCookie, error)
```

**Cookie Format:**
- Cookies are JSON objects serialized to base64
- Cookie value structure defined by `MiddlewareCookie` in session.go

---

### 4. session.go

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/session.go`

Session validation and user authentication using database-backed sessions.

**Key Types:**

```go
type MiddlewareCookie struct {
    Session string `json:"session"`  // Session identifier
    UserId  int64  `json:"userId"`   // User ID
}
```

**Key Functions:**

```go
// UserIsAuth validates a user's authentication status
func UserIsAuth(r *http.Request, cookie *http.Cookie, c *config.Config) (*db.Users, error)
```

**Authentication Steps:**
1. Decode cookie (base64 JSON)
2. Extract `userId` and `session` from cookie
3. Query database for session by userId
4. Compare cookie session with database session
5. Check if session is expired (using `utility.TimestampLessThan`)
6. Retrieve and return user record

---

## Middleware Chain

### Integration in main.go

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go:138`

The middleware chain is established at application startup:

```go
mux := router.NewModulacmsMux(*configuration)
middlewareHandler := middleware.Serve(mux, configuration)

httpServer := &http.Server{
    Addr:         configuration.Client_Site + configuration.Port,
    Handler:      middlewareHandler,  // ← Middleware wraps mux
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}

httpsServer := &http.Server{
    Addr:         configuration.Client_Site + configuration.SSL_Port,
    TLSConfig:    manager.TLSConfig(),
    Handler:      middlewareHandler,  // ← Same middleware for HTTPS
}
```

**Request Flow:**

```
HTTP Request → middlewareHandler (middleware.Serve)
    ↓
CORS headers applied
    ↓
Authentication attempted
    ↓
    ├─ Authenticated? → User injected into context → Router (mux) → Handler
    └─ Not authenticated + /api route? → 401 Unauthorized
    └─ Not authenticated + public route? → Router (mux) → Handler
```

### Middleware Order

ModulaCMS uses a single middleware function (`Serve`) that executes operations in this order:

1. **CORS** (always runs first)
2. **Authentication** (cookie extraction and validation)
3. **API Protection** (blocks unauthenticated `/api/*` requests)
4. **Context Injection** (adds user to request context if authenticated)
5. **Next Handler** (passes to router)

**Important:** This is NOT a composable chain like Express.js or Echo. There is one middleware function that does everything. To add new functionality, you modify `Serve()` directly.

---

## Authentication Flow

### Cookie-Based Authentication

ModulaCMS uses session cookies for authentication. Here's the complete flow:

#### 1. User Login (Not in middleware - handled by router)

```go
// In internal/router/auth.go or similar
// 1. Validate credentials
// 2. Create session in database
// 3. Create cookie with session ID and user ID
cookie := &http.Cookie{
    Name:  config.Cookie_Name,
    Value: base64EncodedJSON,  // {session: "abc123", userId: 42}
    // ... other cookie properties
}
http.SetCookie(w, cookie)
```

#### 2. Subsequent Requests (Middleware validates)

```go
// In middleware.Serve()
func Serve(next http.Handler, c *config.Config) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        Cors(w, r, c)  // Always set CORS first

        u, user := AuthRequest(w, r, c)
        if u != nil {
            // User authenticated - inject into context
            ctx := context.WithValue(r.Context(), u, user)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }

        // Not authenticated - check if API route
        if strings.Contains(r.URL.Path, "api") {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte(fmt.Sprintf("Unauthorized Request to %s", r.URL.Path)))
            return
        }

        // Public route - allow through
        next.ServeHTTP(w, r)
    })
}
```

#### 3. Authentication Validation (AuthRequest → UserIsAuth)

```go
// Extract cookie
func AuthRequest(w http.ResponseWriter, r *http.Request, c *config.Config) (*authcontext, *db.Users) {
    // Skip favicon
    if strings.Contains(r.URL.Path, "favicon.ico") {
        return nil, nil
    }

    // Get cookie
    cookie, err := r.Cookie(c.Cookie_Name)
    if err != nil {
        return nil, nil  // No cookie = not authenticated
    }

    // Validate session
    user, err := UserIsAuth(r, cookie, c)
    if err != nil {
        return nil, nil  // Invalid session
    }

    var u authcontext = "authenticated"
    return &u, user
}

// Validate session
func UserIsAuth(r *http.Request, cookie *http.Cookie, c *config.Config) (*db.Users, error) {
    // 1. Decode cookie
    userCookie, err := ReadCookie(cookie)
    if err != nil {
        return nil, err
    }

    // 2. Connect to database
    dbc := db.ConfigDB(*c)

    // 3. Get session from database
    session, err := dbc.GetSessionByUserId(userCookie.UserId)
    if err != nil || session == nil {
        return nil, err
    }

    // 4. Compare session IDs
    if userCookie.Session != session.SessionData.String {
        return nil, fmt.Errorf("sessions don't match")
    }

    // 5. Check expiration
    expired := utility.TimestampLessThan(session.ExpiresAt.String)
    if expired {
        return nil, fmt.Errorf("session is expired")
    }

    // 6. Get user record
    u, err := dbc.GetUser(userCookie.UserId)
    if err != nil {
        return nil, err
    }

    return u, nil
}
```

### Retrieving Authenticated User in Handlers

In your router handlers, retrieve the authenticated user from context:

```go
func MyHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
    // Extract user from context
    var authKey middleware.authcontext = "authenticated"
    userInterface := r.Context().Value(authKey)

    if userInterface != nil {
        user := userInterface.(*db.Users)
        // User is authenticated - use user.ID, user.Email, etc.
    } else {
        // User is not authenticated (public route)
    }
}
```

**Note:** The `authcontext` type is unexported, so you can't access it outside the middleware package. This is a design limitation. Consider exporting it or providing a helper function.

---

## CORS Configuration

### Configuration Structure

CORS behavior is controlled by these fields in `config.Config`:

```go
type Config struct {
    Cors_Origins     []string `json:"cors_origins"`      // Allowed origins
    Cors_Methods     []string `json:"cors_methods"`      // Allowed HTTP methods
    Cors_Headers     []string `json:"cors_headers"`      // Allowed headers
    Cors_Credentials bool     `json:"cors_credentials"`  // Allow credentials
    // ... other fields
}
```

### Example Configuration (config.json)

```json
{
  "cors_origins": [
    "http://localhost:3000",
    "https://admin.example.com",
    "https://example.com"
  ],
  "cors_methods": [
    "GET",
    "POST",
    "PUT",
    "DELETE",
    "OPTIONS"
  ],
  "cors_headers": [
    "Content-Type",
    "Authorization",
    "X-Requested-With"
  ],
  "cors_credentials": true
}
```

### CORS Behavior

**Origin Matching:**
- **Exact match**: Origin must be in `cors_origins` list
- **Wildcard**: `"*"` allows all origins (use cautiously)
- **No match**: No CORS headers are set (browser blocks request)

**Credentials:**
- When `cors_credentials: true`, browser can send cookies/auth headers
- Requires exact origin match (cannot use `*` with credentials)

**Preflight Requests:**
- Browser sends OPTIONS request before actual request
- Middleware returns 204 No Content with CORS headers
- Browser proceeds with actual request if allowed

### CORS Testing

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/tests/cors_test.go`

Tests cover:
- Default CORS settings with allowed origin
- Multiple allowed origins with specific origin
- Rejected origins (not in allowed list)
- Wildcard origin (`*`)
- Credentials enabled/disabled
- Preflight OPTIONS requests

**Example Test:**

```go
func TestCorsWithConfig(t *testing.T) {
    config := config.Config{
        Cors_Origins:     []string{"http://localhost:3000"},
        Cors_Methods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        Cors_Headers:     []string{"Content-Type", "Authorization"},
        Cors_Credentials: true,
    }

    req := httptest.NewRequest("GET", "http://example.com/foo", nil)
    req.Header.Set("Origin", "http://localhost:3000")

    w := httptest.NewRecorder()
    middleware.CorsWithConfig(w, req, &config)

    // Verify headers
    if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
        t.Error("Origin not set correctly")
    }
}
```

---

## Cookie Management

### Cookie Structure

Cookies use base64-encoded JSON for structured data:

```go
type MiddlewareCookie struct {
    Session string `json:"session"`  // Session identifier (UUID or similar)
    UserId  int64  `json:"userId"`   // User database ID
}
```

**Example Cookie Value:**

```
Raw JSON:    {"session":"abc123def456","userId":42}
Base64:      eyJzZXNzaW9uIjoiYWJjMTIzZGVmNDU2IiwidXNlcklkIjo0Mn0=
Cookie:      Cookie: modulacms_session=eyJzZXNzaW9uIjoiYWJjMTIzZGVmNDU2IiwidXNlcklkIjo0Mn0=
```

### Reading Cookies

```go
// 1. Get cookie from request
cookie, err := r.Cookie(config.Cookie_Name)
if err != nil {
    // Cookie not found
}

// 2. Decode cookie
middlewareCookie, err := middleware.ReadCookie(cookie)
if err != nil {
    // Invalid cookie format or base64 decoding failed
}

// 3. Access cookie data
sessionID := middlewareCookie.Session
userID := middlewareCookie.UserId
```

**ReadCookie Process:**
1. Validate cookie (check expiration, etc.)
2. Base64 decode the cookie value
3. JSON unmarshal into `MiddlewareCookie` struct

### Writing Cookies

```go
// 1. Create cookie data
cookieData := middleware.MiddlewareCookie{
    Session: "abc123def456",
    UserId:  42,
}

// 2. Marshal to JSON
jsonData, _ := json.Marshal(cookieData)

// 3. Base64 encode
encodedValue := base64.StdEncoding.EncodeToString(jsonData)

// 4. Create HTTP cookie
cookie := &http.Cookie{
    Name:     config.Cookie_Name,
    Value:    encodedValue,
    Path:     "/",
    HttpOnly: true,
    Secure:   true,  // HTTPS only
    SameSite: http.SameSiteStrictMode,
    MaxAge:   3600,  // 1 hour
}

// 5. Set cookie
http.SetCookie(w, cookie)
```

**Security Considerations:**
- Always use `HttpOnly: true` to prevent JavaScript access
- Use `Secure: true` in production (HTTPS only)
- Use `SameSite: Strict` or `Lax` to prevent CSRF attacks
- Set appropriate `MaxAge` or `Expires` for session duration

---

## Session Validation

### Session Database Schema

**Table:** `sessions` (schema 15_sessions)

```sql
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    session_data TEXT,        -- Session identifier (stored in cookie)
    expires_at TIMESTAMP,     -- Session expiration timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### Session Validation Process

1. **Extract Cookie**: Get session cookie from request
2. **Decode Cookie**: Base64 decode and JSON unmarshal
3. **Database Query**: Retrieve session by `user_id`
4. **Session Match**: Compare cookie session with database `session_data`
5. **Expiration Check**: Verify `expires_at` is in the future
6. **User Retrieval**: Get user record from `users` table

**Code Implementation:**

```go
func UserIsAuth(r *http.Request, cookie *http.Cookie, c *config.Config) (*db.Users, error) {
    // Decode cookie
    userCookie, err := ReadCookie(cookie)
    if err != nil {
        return nil, err
    }

    // Get database connection
    dbc := db.ConfigDB(*c)

    // Get session from database
    session, err := dbc.GetSessionByUserId(userCookie.UserId)
    if err != nil || session == nil {
        utility.DefaultLogger.Error("Error retrieving session or no sessions found:", err)
        return nil, err
    }

    // Verify session matches
    if userCookie.Session != session.SessionData.String {
        err := fmt.Errorf("sessions don't match")
        utility.DefaultLogger.Warn("", err)
        return nil, err
    }

    // Check expiration
    expired := utility.TimestampLessThan(session.ExpiresAt.String)
    if expired {
        return nil, fmt.Errorf("session is expired")
    }

    // Get user
    u, err := dbc.GetUser(userCookie.UserId)
    if err != nil {
        return nil, err
    }

    return u, nil
}
```

### Session Lifecycle

**1. Session Creation** (during login):
```sql
INSERT INTO sessions (user_id, session_data, expires_at)
VALUES (42, 'abc123def456', '2026-01-13 12:00:00');
```

**2. Session Validation** (every request with cookie):
```sql
SELECT * FROM sessions WHERE user_id = 42;
```

**3. Session Expiration**:
- Sessions expire based on `expires_at` timestamp
- Expired sessions return authentication failure
- Client should redirect to login page

**4. Session Revocation** (logout):
```sql
DELETE FROM sessions WHERE user_id = 42;
-- OR update session_data to invalidate
UPDATE sessions SET session_data = NULL WHERE user_id = 42;
```

---

## Adding New Middleware

### Current Limitation

ModulaCMS uses a single `Serve()` function rather than a composable middleware chain. To add new middleware functionality, you must modify the `Serve()` function directly.

### Step-by-Step Guide

#### 1. Determine Middleware Placement

Decide where your middleware logic should run in the chain:

```go
func Serve(next http.Handler, c *config.Config) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Pre-CORS processing (rarely needed)

        Cors(w, r, c)  // CORS always runs first

        // 2. Pre-authentication processing (logging, request ID, etc.)

        u, user := AuthRequest(w, r, c)

        // 3. Post-authentication processing (based on auth result)

        if u != nil {
            ctx := context.WithValue(r.Context(), u, user)

            // 4. Authenticated user processing (rate limiting, etc.)

            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }

        // 5. Unauthenticated user processing

        if strings.Contains(r.URL.Path, "api") {
            w.WriteHeader(http.StatusUnauthorized)
            msg := fmt.Sprintf("Unauthorized Request to %s", string(r.URL.Path))
            w.Write([]byte(msg))
            return
        }

        // 6. Public route processing

        next.ServeHTTP(w, r)
    })
}
```

#### 2. Add Configuration (if needed)

If your middleware needs configuration, add fields to `config.Config`:

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/config/config.go`

```go
type Config struct {
    // ... existing fields

    // Rate limiting example
    RateLimit_Enabled       bool     `json:"rate_limit_enabled"`
    RateLimit_RequestsPer   int      `json:"rate_limit_requests_per"`
    RateLimit_Duration      string   `json:"rate_limit_duration"`
}
```

#### 3. Implement Middleware Logic

Create a new file in `internal/middleware/` or add to existing files:

**Example: Rate Limiting**

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/ratelimit.go`

```go
package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/hegner123/modulacms/internal/config"
)

type rateLimiter struct {
    visitors map[string]*visitor
    mu       sync.RWMutex
    limit    int
    duration time.Duration
}

type visitor struct {
    lastSeen  time.Time
    count     int
}

var limiter *rateLimiter

// InitRateLimiter initializes the rate limiter with config
func InitRateLimiter(c *config.Config) {
    if !c.RateLimit_Enabled {
        return
    }

    duration, _ := time.ParseDuration(c.RateLimit_Duration)
    limiter = &rateLimiter{
        visitors: make(map[string]*visitor),
        limit:    c.RateLimit_RequestsPer,
        duration: duration,
    }

    // Cleanup goroutine
    go limiter.cleanupVisitors()
}

// CheckRateLimit checks if IP is rate limited
func CheckRateLimit(ip string) bool {
    if limiter == nil {
        return false  // Rate limiting disabled
    }

    limiter.mu.Lock()
    defer limiter.mu.Unlock()

    v, exists := limiter.visitors[ip]
    if !exists {
        limiter.visitors[ip] = &visitor{
            lastSeen: time.Now(),
            count:    1,
        }
        return false
    }

    if time.Since(v.lastSeen) > limiter.duration {
        v.count = 1
        v.lastSeen = time.Now()
        return false
    }

    v.count++
    v.lastSeen = time.Now()

    return v.count > limiter.limit
}

func (rl *rateLimiter) cleanupVisitors() {
    for {
        time.Sleep(rl.duration)
        rl.mu.Lock()
        for ip, v := range rl.visitors {
            if time.Since(v.lastSeen) > rl.duration {
                delete(rl.visitors, ip)
            }
        }
        rl.mu.Unlock()
    }
}
```

#### 4. Integrate into Serve()

Modify the `Serve()` function to use your new middleware:

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/middleware.go`

```go
func Serve(next http.Handler, c *config.Config) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        Cors(w, r, c)

        // Add rate limiting check
        ip := getIP(r)
        if CheckRateLimit(ip) {
            w.WriteHeader(http.StatusTooManyRequests)
            w.Write([]byte("Rate limit exceeded"))
            return
        }

        u, user := AuthRequest(w, r, c)
        if u != nil {
            ctx := context.WithValue(r.Context(), u, user)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }
        if strings.Contains(r.URL.Path, "api") {
            w.WriteHeader(http.StatusUnauthorized)
            msg := fmt.Sprintf("Unauthorized Request to %s", string(r.URL.Path))
            w.Write([]byte(msg))
            return
        }

        next.ServeHTTP(w, r)
    })
}

func getIP(r *http.Request) string {
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        return strings.Split(forwarded, ",")[0]
    }
    return strings.Split(r.RemoteAddr, ":")[0]
}
```

#### 5. Initialize Middleware (if needed)

If your middleware requires initialization, call it from `main.go`:

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go`

```go
func main() {
    // ... existing code

    configuration := config.ReadConfiguration()

    // Initialize rate limiter
    middleware.InitRateLimiter(configuration)

    // ... rest of main
}
```

#### 6. Add Tests

Create tests for your new middleware:

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/tests/ratelimit_test.go`

```go
package middleware_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/hegner123/modulacms/internal/config"
    "github.com/hegner123/modulacms/internal/middleware"
)

func TestRateLimiting(t *testing.T) {
    config := &config.Config{
        RateLimit_Enabled:     true,
        RateLimit_RequestsPer: 5,
        RateLimit_Duration:    "1m",
    }

    middleware.InitRateLimiter(config)

    ip := "192.168.1.1"

    // First 5 requests should succeed
    for i := 0; i < 5; i++ {
        if middleware.CheckRateLimit(ip) {
            t.Errorf("Request %d should not be rate limited", i+1)
        }
    }

    // 6th request should be rate limited
    if !middleware.CheckRateLimit(ip) {
        t.Error("6th request should be rate limited")
    }
}
```

### Alternative: Composable Middleware (Future Improvement)

To enable composable middleware (recommended for complex applications), consider refactoring to:

```go
type Middleware func(http.Handler) http.Handler

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}

// Usage in main.go
middlewareHandler := middleware.Chain(
    mux,
    middleware.Cors(configuration),
    middleware.RateLimit(configuration),
    middleware.Authentication(configuration),
    middleware.Logging(),
)
```

This would allow independent middleware functions that can be composed in any order.

---

## Testing

### Test Organization

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/tests/`

Tests are in a separate `tests/` directory using the `middleware_test` package.

### Testing Strategy

#### 1. Unit Tests (Individual Functions)

Test individual middleware functions in isolation:

```go
func TestCorsWithConfig(t *testing.T) {
    // Test CORS header setting
}

func TestReadCookie(t *testing.T) {
    // Test cookie decoding
}

func TestUserIsAuth(t *testing.T) {
    // Test session validation (requires test database)
}
```

#### 2. Integration Tests (Full Middleware Chain)

Test the complete middleware chain with `httptest`:

```go
func TestMiddlewareChain(t *testing.T) {
    // Create test configuration
    config := &config.Config{
        Cookie_Name:      "test_session",
        Cors_Origins:     []string{"http://localhost:3000"},
        Cors_Methods:     []string{"GET", "POST"},
        Cors_Headers:     []string{"Content-Type"},
        Cors_Credentials: true,
    }

    // Create test handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    // Wrap with middleware
    middlewareHandler := middleware.Serve(handler, config)

    // Create test request
    req := httptest.NewRequest("GET", "http://example.com/api/test", nil)
    req.Header.Set("Origin", "http://localhost:3000")

    // Record response
    w := httptest.NewRecorder()

    // Execute request
    middlewareHandler.ServeHTTP(w, req)

    // Assert results
    if w.Code != http.StatusUnauthorized {
        t.Errorf("Expected 401 for unauthenticated API request, got %d", w.Code)
    }
}
```

#### 3. CORS Tests

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/tests/cors_test.go`

Existing CORS tests cover:
- Allowed origins (exact match)
- Multiple allowed origins
- Rejected origins (not in list)
- Wildcard origin (`*`)
- Credentials enabled/disabled
- Preflight OPTIONS requests

**Run CORS Tests:**

```bash
go test -v ./internal/middleware/tests -run TestCors
```

#### 4. Authentication Tests (Requires Database)

Authentication tests need a test database:

```go
func TestAuthenticationFlow(t *testing.T) {
    // 1. Set up test database
    testDB := setupTestDB()
    defer cleanupTestDB(testDB)

    // 2. Create test user and session
    userID := int64(1)
    sessionID := "test-session-123"
    createTestSession(testDB, userID, sessionID)

    // 3. Create cookie
    cookieData := middleware.MiddlewareCookie{
        Session: sessionID,
        UserId:  userID,
    }
    jsonData, _ := json.Marshal(cookieData)
    encodedValue := base64.StdEncoding.EncodeToString(jsonData)

    cookie := &http.Cookie{
        Name:  "test_session",
        Value: encodedValue,
    }

    // 4. Test authentication
    req := httptest.NewRequest("GET", "http://example.com/api/test", nil)
    req.AddCookie(cookie)

    config := testConfig()
    u, user := middleware.AuthRequest(httptest.NewRecorder(), req, config)

    // 5. Verify results
    if u == nil || user == nil {
        t.Error("Authentication should succeed with valid session")
    }
    if user.ID != userID {
        t.Errorf("Expected user ID %d, got %d", userID, user.ID)
    }
}
```

### Running Tests

```bash
# All middleware tests
go test -v ./internal/middleware/tests

# Specific test
go test -v ./internal/middleware/tests -run TestCorsWithConfig

# With coverage
go test -v ./internal/middleware/tests -cover

# Generate coverage report
go test ./internal/middleware/tests -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Configuration

### Complete Configuration Reference

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/config/config.go`

Middleware-related configuration fields:

```go
type Config struct {
    // Cookie configuration
    Cookie_Name         string   `json:"cookie_name"`       // Name of session cookie
    Cookie_Duration     string   `json:"cookie_duration"`   // Cookie lifetime (e.g., "24h")

    // CORS configuration
    Cors_Origins        []string `json:"cors_origins"`      // Allowed origins
    Cors_Methods        []string `json:"cors_methods"`      // Allowed HTTP methods
    Cors_Headers        []string `json:"cors_headers"`      // Allowed headers
    Cors_Credentials    bool     `json:"cors_credentials"`  // Allow credentials

    // Authentication configuration
    Auth_Salt           string   `json:"auth_salt"`         // Salt for password hashing

    // Database configuration (for session validation)
    Db_Driver           DbDriver `json:"db_driver"`         // sqlite, mysql, postgres
    Db_URL              string   `json:"db_url"`            // Database connection URL
    Db_Name             string   `json:"db_name"`           // Database name
    Db_User             string   `json:"db_username"`       // Database username
    Db_Password         string   `json:"db_password"`       // Database password

    // ... other fields
}
```

### Example config.json

```json
{
  "cookie_name": "modulacms_session",
  "cookie_duration": "24h",

  "cors_origins": [
    "http://localhost:3000",
    "http://localhost:5173",
    "https://admin.example.com",
    "https://example.com"
  ],
  "cors_methods": [
    "GET",
    "POST",
    "PUT",
    "DELETE",
    "PATCH",
    "OPTIONS"
  ],
  "cors_headers": [
    "Content-Type",
    "Authorization",
    "X-Requested-With",
    "Accept",
    "Origin"
  ],
  "cors_credentials": true,

  "auth_salt": "your-random-salt-here",

  "db_driver": "sqlite",
  "db_url": "file:./modulacms.db",
  "db_name": "modulacms"
}
```

### Environment Variables

Sensitive configuration can be loaded from environment variables instead of JSON:

```bash
# Database credentials
export MODULACMS_DB_USER="dbuser"
export MODULACMS_DB_PASSWORD="secure-password"

# OAuth secrets
export MODULACMS_OAUTH_CLIENT_SECRET="oauth-secret"

# Salt
export MODULACMS_AUTH_SALT="random-salt-value"
```

Configuration loading priority:
1. Environment variables (highest priority)
2. JSON configuration file
3. Default values (lowest priority)

---

## Common Patterns

### Pattern 1: Extracting Authenticated User

In any handler that receives authenticated requests:

```go
func MyAPIHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
    // Note: This is a limitation - authcontext is unexported
    // You cannot easily access the authenticated user from handlers

    // Workaround: Re-validate session or use a different approach
    // OR: Export authcontext constant from middleware package
}
```

**Recommended Fix:** Export the context key:

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/middleware.go`

```go
// Export the context key
type AuthContext string
const AuthContextKey AuthContext = "authenticated"

func Serve(next http.Handler, c *config.Config) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        Cors(w, r, c)

        u, user := AuthRequest(w, r, c)
        if u != nil {
            // Use exported constant
            ctx := context.WithValue(r.Context(), AuthContextKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }
        // ... rest of function
    })
}
```

Then in handlers:

```go
import "github.com/hegner123/modulacms/internal/middleware"

func MyAPIHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
    userInterface := r.Context().Value(middleware.AuthContextKey)
    if userInterface == nil {
        // Shouldn't happen for /api routes, but handle it
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    user := userInterface.(*db.Users)
    // Use user.ID, user.Email, etc.
}
```

### Pattern 2: Bypassing Authentication for Specific Routes

To allow unauthenticated access to certain API routes:

```go
func Serve(next http.Handler, c *config.Config) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        Cors(w, r, c)

        u, user := AuthRequest(w, r, c)
        if u != nil {
            ctx := context.WithValue(r.Context(), u, user)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }

        // Define public API routes
        publicAPIRoutes := []string{
            "/api/v1/auth/register",
            "/api/v1/auth/login",
            "/api/v1/auth/reset",
            "/api/v1/health",
        }

        // Check if current path is public
        isPublicAPI := false
        for _, route := range publicAPIRoutes {
            if strings.HasPrefix(r.URL.Path, route) {
                isPublicAPI = true
                break
            }
        }

        // Block if API route and not public
        if strings.Contains(r.URL.Path, "api") && !isPublicAPI {
            w.WriteHeader(http.StatusUnauthorized)
            msg := fmt.Sprintf("Unauthorized Request to %s", r.URL.Path)
            w.Write([]byte(msg))
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### Pattern 3: Logging All Requests

Add request logging to the middleware:

```go
func Serve(next http.Handler, c *config.Config) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Log incoming request
        utility.DefaultLogger.Info("Request: %s %s from %s",
            r.Method,
            r.URL.Path,
            r.RemoteAddr,
        )

        Cors(w, r, c)

        u, user := AuthRequest(w, r, c)
        if u != nil {
            utility.DefaultLogger.Info("Authenticated user: %d", user.ID)
            ctx := context.WithValue(r.Context(), u, user)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }

        // ... rest of function
    })
}
```

### Pattern 4: Custom Cookie Settings

To customize cookie behavior beyond the config:

```go
// When setting cookies in auth handlers
cookie := &http.Cookie{
    Name:     config.Cookie_Name,
    Value:    encodedValue,
    Path:     "/",
    Domain:   config.Client_Site,  // Restrict to specific domain
    HttpOnly: true,                 // Prevent JavaScript access
    Secure:   true,                 // HTTPS only (production)
    SameSite: http.SameSiteStrictMode,  // CSRF protection
    MaxAge:   86400,                // 24 hours in seconds
}

http.SetCookie(w, cookie)
```

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - Database abstraction used for session validation

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` - Adding new middleware features
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md` - Testing strategies

**Packages:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/AUTH_PACKAGE.md` - OAuth authentication (related to middleware auth)

**Domain:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/AUTH_AND_OAUTH.md` - Authentication flows and session management

**Reference:**
- `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md` - Project development guidelines
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/FILE_TREE.md` - Complete project structure

---

## Quick Reference

### File Locations

```
internal/middleware/
├── middleware.go     # Main Serve() function and AuthRequest()
├── cors.go          # CORS header management
├── cookies.go       # Cookie reading/writing utilities
├── session.go       # Session validation (UserIsAuth)
└── tests/
    └── cors_test.go # CORS tests
```

### Key Functions

```go
// Main middleware chain
middleware.Serve(next http.Handler, c *config.Config) http.Handler

// CORS handling
middleware.Cors(w http.ResponseWriter, r *http.Request, c *config.Config)

// Authentication
middleware.AuthRequest(w, r, c) (*authcontext, *db.Users)
middleware.UserIsAuth(r, cookie, c) (*db.Users, error)

// Cookie operations
middleware.ReadCookie(cookie *http.Cookie) (*MiddlewareCookie, error)
middleware.SetCookieHandler(w http.ResponseWriter, c *http.Cookie)
```

### Configuration Fields

```go
config.Cookie_Name         // Session cookie name
config.Cookie_Duration     // Cookie lifetime
config.Cors_Origins        // Allowed origins
config.Cors_Methods        // Allowed HTTP methods
config.Cors_Headers        // Allowed headers
config.Cors_Credentials    // Allow credentials (true/false)
```

### Integration Point

```go
// cmd/main.go:138
mux := router.NewModulacmsMux(*configuration)
middlewareHandler := middleware.Serve(mux, configuration)

httpServer := &http.Server{
    Handler: middlewareHandler,  // ← Middleware wraps router
    // ...
}
```

### Request Flow

```
HTTP Request
    ↓
middleware.Serve()
    ↓
CORS headers (Cors)
    ↓
Authentication (AuthRequest → UserIsAuth)
    ↓
    ├─ Valid session? → Inject user context → Router → Handler
    ├─ Invalid + /api? → 401 Unauthorized
    └─ Invalid + public? → Router → Handler
```

### Testing Commands

```bash
# All middleware tests
go test -v ./internal/middleware/tests

# CORS tests only
go test -v ./internal/middleware/tests -run TestCors

# With coverage
go test -v ./internal/middleware/tests -cover
```

### Common Tasks

**Add new middleware functionality:**
1. Modify `Serve()` function in `middleware.go`
2. Add configuration fields to `config.Config` if needed
3. Create new file for complex logic (e.g., `ratelimit.go`)
4. Add tests in `tests/` directory
5. Update this documentation

**Debug authentication issues:**
1. Check session exists in database: `SELECT * FROM sessions WHERE user_id = ?`
2. Verify cookie value matches: Compare base64-decoded cookie with `session_data`
3. Check expiration: Verify `expires_at` is in the future
4. Enable debug logging: `utility.DefaultLogger.Debug()`

**Configure CORS:**
1. Edit `cors_origins` in config.json
2. Add allowed methods to `cors_methods`
3. Add required headers to `cors_headers`
4. Set `cors_credentials: true` for cookies/auth headers
5. Test with browser DevTools Network tab

---

**Last Updated:** 2026-01-12
**Maintained By:** ModulaCMS Core Team
**Status:** Complete
