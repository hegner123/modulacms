# middleware

Package middleware provides HTTP and SSH middleware components for ModulaCMS. It includes authentication, authorization, CORS, rate limiting, session management, request logging, and audit helpers for both web and SSH interfaces.

## Overview

The middleware package implements the chain-of-responsibility pattern for HTTP request processing and SSH session handling. HTTP middleware wraps handlers to add cross-cutting concerns like authentication, CORS headers, and logging. SSH middleware uses the Charm Wish framework to provide similar functionality for terminal sessions.

Key capabilities include panic recovery, request ID generation, cookie-based and API key authentication, per-IP rate limiting, session validation, HTTP request metrics recording, audit context extraction, and predefined middleware chains for common use cases.

## Constants

No exported constants.

## Types

### authcontext

```go
type authcontext string
```

Unexported type used as a key for storing authentication information in request context. The value "authenticated" is used throughout the package to store and retrieve user data from context.

### MiddlewareCookie

```go
type MiddlewareCookie struct {
    Session string       `json:"session"`
    UserId  types.UserID `json:"userId"`
}
```

Represents the structure of authentication cookies. Contains session identifier and user ID. Used for cookie-based authentication in HTTP requests.

### RateLimiter

```go
type RateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
    cleanup  time.Duration
}
```

Implements per-IP rate limiting using the token bucket algorithm. Tracks individual limiters for each IP address and enforces configurable rate limits to prevent abuse of authentication endpoints.

## Functions

### AuthRequest

```go
func AuthRequest(w http.ResponseWriter, r *http.Request, c *config.Config) (*authcontext, *db.Users)
```

Extracts and validates authentication information from the request. Retrieves the authentication cookie, verifies it, and returns the authenticated user if valid. Falls back to API key authentication via the Authorization header when cookie auth is not present or fails. Returns nil values if all authentication methods fail.

### APIKeyAuth

```go
func APIKeyAuth(r *http.Request, c *config.Config) (*authcontext, *db.Users)
```

Authenticates a request using an API key from the Authorization header. Expects the header format "Bearer key", looks up the token in the database, and validates that the token is of type "api_key", is not revoked, and has not expired. Returns the authenticated context and user on success.

### CorsMiddleware

```go
func CorsMiddleware(c *config.Config) func(http.Handler) http.Handler
```

Returns middleware that wraps an http.Handler and adds CORS headers based on configuration. Automatically responds to preflight OPTIONS requests with proper CORS headers.

### Cors

```go
func Cors(w http.ResponseWriter, r *http.Request, c *config.Config)
```

Sets CORS headers based on configuration. Direct function for adding CORS headers without middleware wrapping.

### CorsWithConfig

```go
func CorsWithConfig(w http.ResponseWriter, r *http.Request, c *config.Config)
```

Allows specifying a config for CORS settings. Validates origin against allowed origins, sets allowed methods and headers, handles credentials flag, and responds to preflight requests.

### SetCookieHandler

```go
func SetCookieHandler(w http.ResponseWriter, c *http.Cookie)
```

Sets a cookie in the HTTP response and writes a basic response body. Logs the headers and bytes written for debugging purposes. Used for testing cookie functionality.

### ReadCookie

```go
func ReadCookie(c *http.Cookie) (*MiddlewareCookie, error)
```

Decodes and deserializes a cookie value into a MiddlewareCookie struct. Validates the cookie, base64 decodes its value, and unmarshals the JSON data. Returns an error if any step fails.

### WriteCookie

```go
func WriteCookie(w http.ResponseWriter, c *config.Config, sessionData string, userId types.UserID) error
```

Creates and sets a secure authentication cookie with proper security flags. Encodes the session data and user ID as base64-encoded JSON and applies security settings from configuration including HttpOnly, Secure, and SameSite. Returns an error if encoding or cookie creation fails.

### NewRateLimiter

```go
func NewRateLimiter(r rate.Limit, b int) *RateLimiter
```

Creates a new rate limiter with the specified rate and burst size. The rate parameter controls how many requests per second are allowed. The burst parameter controls how many requests can be made in a short burst. Example usage: NewRateLimiter 0.16667 with burst 10 allows 10 requests per minute.

### Middleware (RateLimiter)

```go
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler
```

Returns an HTTP middleware handler that enforces rate limits. If a client exceeds the rate limit, it receives a 429 Too Many Requests response. Each IP address is tracked independently.

### Size (RateLimiter)

```go
func (rl *RateLimiter) Size() int
```

Returns the number of active rate limiters. Primarily useful for testing and monitoring.

### Clear (RateLimiter)

```go
func (rl *RateLimiter) Clear()
```

Removes all rate limiters. Should only be used in testing to reset state between tests.

### RecoveryMiddleware

```go
func RecoveryMiddleware() func(http.Handler) http.Handler
```

Catches panics in HTTP handlers, recovers, reports the error via utility.CaptureError, and returns a 500 Internal Server Error response. Should be the outermost middleware in the chain.

### HTTPMetricsMiddleware

```go
func HTTPMetricsMiddleware() func(http.Handler) http.Handler
```

Records HTTP request metrics to utility.GlobalMetrics. Tracks request counts (http.requests counter), request duration (http.duration histogram), and error counts (http.errors counter) with labels for method, path, and status code.

### HTTPLoggingMiddleware

```go
func HTTPLoggingMiddleware() func(http.Handler) http.Handler
```

Logs HTTP requests and responses including method, path, remote address, status code, and duration. Uses a response wrapper to capture status code.

### HTTPAuthenticationMiddleware

```go
func HTTPAuthenticationMiddleware(c *config.Config) func(http.Handler) http.Handler
```

Validates session cookies and populates request context with authenticated user. Injects user into context using authcontext key. Continues without auth context if authentication fails.

### HTTPAuthorizationMiddleware

```go
func HTTPAuthorizationMiddleware(c *config.Config) func(http.Handler) http.Handler
```

Blocks unauthenticated requests to protected endpoints. Use this on routes that require authentication. Returns 401 Unauthorized if no authenticated user in context.

### HTTPPublicEndpointMiddleware

```go
func HTTPPublicEndpointMiddleware(c *config.Config) func(http.Handler) http.Handler
```

Allows public endpoints through, blocks others. Used as a global middleware to protect all API routes by default. Allows exact matches to PublicEndpoints list and non-API routes. Checks authentication for all other API routes.

### AuthenticatedUser

```go
func AuthenticatedUser(ctx context.Context) *db.Users
```

Extracts the authenticated user from the request context. Returns nil if no user is authenticated. Uses authcontext key to retrieve user.

### Chain

```go
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler
```

Applies multiple middleware in sequence from left to right. Example usage: Chain middleware1, middleware2, middleware3 applies in that order.

### UserIsAuth

```go
func UserIsAuth(r *http.Request, cookie *http.Cookie, c *config.Config) (*db.Users, error)
```

Validates a user's authentication status based on the provided cookie. Verifies that the session in the cookie matches the one in the database, checks if the session is still valid and not expired, refreshes OAuth tokens if needed, and retrieves the user data. Returns the user object if authentication is successful.

### AuditContextFromRequest

```go
func AuditContextFromRequest(r *http.Request, c config.Config) audited.AuditContext
```

Builds an AuditContext from the HTTP request. Extracts the authenticated user, client IP, and request ID from the context. Used for audit logging and change event tracking.

### AuditContextFromCLI

```go
func AuditContextFromCLI(c config.Config, userID types.UserID) audited.AuditContext
```

Builds an AuditContext for CLI and TUI operations. Uses "cli" as the IP address and includes node ID and user ID.

### DefaultMiddlewareChain

```go
func DefaultMiddlewareChain(mgr *config.Manager, pc *PermissionCache) func(http.Handler) http.Handler
```

Returns the standard middleware chain for the application. Includes, in order: (1) panic recovery, (2) request ID generation, (3) client IP resolution, (4) user agent parsing, (5) request/response logging, (6) HTTP request metrics recording, (7) CORS, (8) session authentication, (9) public endpoint protection, (10) permission set injection.

### AuthenticatedChain

```go
func AuthenticatedChain(mgr *config.Manager) func(http.Handler) http.Handler
```

Returns middleware for authenticated-only endpoints. Use this for endpoints that absolutely require authentication. Includes request ID, logging, CORS, authentication, and authorization middleware.

### AuthEndpointChain

```go
func AuthEndpointChain(c *config.Config) func(http.Handler) http.Handler
```

Returns middleware for auth endpoints like login and register. Includes rate limiting set to 10 requests per minute to prevent brute force attacks.

### PublicAPIChain

```go
func PublicAPIChain(c *config.Config) func(http.Handler) http.Handler
```

Returns middleware for public API endpoints. No authentication required but includes logging and CORS.

### PublicKeyHandler

```go
func PublicKeyHandler(c *config.Config) func(ssh.Context, ssh.PublicKey) bool
```

SSH public key authentication callback. Validates that the key is structurally valid and allows the connection through. Actual authentication and authorization happens in the middleware chain. Logs fingerprint and key type.

### RequestIDMiddleware

```go
func RequestIDMiddleware() func(http.Handler) http.Handler
```

Generates a UUID per request, stores it in the context, and sets the X-Request-ID response header. Uses UUID v4 format generated with crypto/rand.

### RequestIDFromContext

```go
func RequestIDFromContext(ctx context.Context) string
```

Extracts the request ID from the context. Returns an empty string if no request ID is present.

### SSHAuthenticationMiddleware

```go
func SSHAuthenticationMiddleware(c *config.Config) wish.Middleware
```

Validates SSH keys and populates session context. Should run early in the middleware chain to set up authentication state. Looks up user by SSH key fingerprint, marks unregistered keys for provisioning, and updates last used timestamp for registered keys.

### SSHAuthorizationMiddleware

```go
func SSHAuthorizationMiddleware(c *config.Config) wish.Middleware
```

Ensures the user is authenticated before proceeding. Can be used to protect specific endpoints or require authentication. Allows through users needing provisioning for provisioning flow.

### SSHRateLimitMiddleware

```go
func SSHRateLimitMiddleware(c *config.Config) wish.Middleware
```

Limits connection attempts per IP to prevent brute force attacks on SSH keys. Currently unimplemented, passes through all requests.

### SSHSessionLoggingMiddleware

```go
func SSHSessionLoggingMiddleware(c *config.Config) wish.Middleware
```

Logs SSH session details including remote address, user, session start, and session end.

### FingerprintSHA256

```go
func FingerprintSHA256(key ssh.PublicKey) string
```

Generates a SHA256 fingerprint from an SSH public key. Returns format "SHA256:base64hash" matching modern SSH clients.

### ParseSSHPublicKey

```go
func ParseSSHPublicKey(publicKeyStr string) (keyType string, fingerprint string, err error)
```

Parses an SSH public key string in authorized_keys format and returns the key type and fingerprint. Returns error if parsing fails.

## Variables

### PublicEndpoints

```go
var PublicEndpoints = []string{
    "/api/v1/auth/login",
    "/api/v1/auth/register",
    "/api/v1/auth/logout",
    "/api/v1/auth/reset",
    "/api/v1/auth/me",
    "/api/v1/auth/oauth/login",
    "/api/v1/auth/oauth/callback",
    "/api/v1/health",
    "/favicon.ico",
}
```

Lists API endpoints that do not require authentication. Used by HTTPPublicEndpointMiddleware to allow unauthenticated access to specific routes.

## Authorization (RBAC)

### PermissionCache

```go
type PermissionCache struct { ... }
```

In-memory role-to-permissions map for lock-free reads. Uses build-then-swap pattern for concurrent safety.

### NewPermissionCache

```go
func NewPermissionCache() *PermissionCache
```

Creates a new empty PermissionCache.

### PermissionCache.Load

```go
func (pc *PermissionCache) Load(driver db.RBACRepository) error
```

Loads all role-permission mappings from the database. Builds a new map and atomically swaps it in.

### PermissionCache.PermissionsForRole

```go
func (pc *PermissionCache) PermissionsForRole(roleID types.RoleID) PermissionSet
```

Returns the PermissionSet for a given role.

### PermissionCache.IsAdmin

```go
func (pc *PermissionCache) IsAdmin(roleID types.RoleID) bool
```

Checks if a role is the admin role.

### PermissionCache.StartPeriodicRefresh

```go
func (pc *PermissionCache) StartPeriodicRefresh(ctx context.Context, driver db.RBACRepository, interval time.Duration)
```

Starts a background goroutine that refreshes the permission cache at the specified interval.

### PermissionInjector

```go
func PermissionInjector(pc *PermissionCache) func(http.Handler) http.Handler
```

Middleware that resolves the authenticated user's role to a PermissionSet and stores it in the request context.

### RequirePermission

```go
func RequirePermission(permission string) func(http.Handler) http.Handler
```

Middleware that checks for a single permission. Returns 403 if missing.

### RequireResourcePermission

```go
func RequireResourcePermission(resource string) func(http.Handler) http.Handler
```

Middleware that auto-maps HTTP method to operation (GET→read, POST→create, PUT/PATCH→update, DELETE→delete).

### RequireAnyPermission

```go
func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler
```

Middleware that requires at least one of the given permissions (OR logic).

### RequireAllPermissions

```go
func RequireAllPermissions(permissions ...string) func(http.Handler) http.Handler
```

Middleware that requires all of the given permissions (AND logic).

### ContextPermissions

```go
func ContextPermissions(ctx context.Context) PermissionSet
```

Extracts the PermissionSet from the request context.

### ContextIsAdmin

```go
func ContextIsAdmin(ctx context.Context) bool
```

Checks if the user in context is an admin. Admin bypass uses this boolean, not a wildcard in PermissionSet.

### ValidatePermissionLabel

```go
func ValidatePermissionLabel(label string) error
```

Validates `resource:operation` format (character-by-character, no regex).

### SetAuthenticatedUser

```go
func SetAuthenticatedUser(ctx context.Context, user *db.Users) context.Context
```

Sets the authenticated user in the request context.

### ClientIPMiddleware

```go
func ClientIPMiddleware() func(http.Handler) http.Handler
```

Resolves client IP from request headers and stores in context.

### ClientIPFromContext

```go
func ClientIPFromContext(ctx context.Context) string
```

Extracts client IP from the request context.

### UserAgentMiddleware

```go
func UserAgentMiddleware() func(http.Handler) http.Handler
```

Parses User-Agent and Client Hints headers and stores UserAgentInfo in context.

### UserAgentInfoFromContext

```go
func UserAgentInfoFromContext(ctx context.Context) *UserAgentInfo
```

Extracts UserAgentInfo from the request context.

### HookRunnerMiddleware

```go
func HookRunnerMiddleware(runner webhook.HookRunner) func(http.Handler) http.Handler
```

Middleware for storing a webhook hook runner in context.
