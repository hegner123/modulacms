# ModulaCMS Middleware Structure

Clean, composable middleware architecture with no legacy code.

## File Organization

### HTTP Middleware

#### `http_middleware.go` - Core HTTP middleware components
- `HTTPLoggingMiddleware()` - Request/response logging with duration
- `HTTPAuthenticationMiddleware(c)` - Session validation, populates context
- `HTTPAuthorizationMiddleware(c)` - Blocks unauthenticated requests
- `HTTPPublicEndpointMiddleware(c)` - Protects API routes, allows public endpoints
- `Chain(...)` - Helper to compose multiple middleware
- `responseWriter` - Response wrapper to capture status codes

#### `http_chain.go` - Pre-built middleware chains
- `DefaultMiddlewareChain(c)` - Global application chain (logging, CORS, auth, public endpoint protection)
- `AuthenticatedChain(c)` - For protected endpoints (logging, CORS, auth, authorization)
- `AuthEndpointChain(c)` - For login/register (logging, CORS, rate limiting)
- `PublicAPIChain(c)` - For public content (logging, CORS)

#### `cors.go` - CORS handling
- `CorsMiddleware(c)` - Composable CORS middleware
- `Cors(w, r, c)` - Direct CORS function (called by middleware)
- `CorsWithConfig(w, r, c)` - CORS implementation
- `getAllowedOrigin(...)` - Origin validation

#### `ratelimit.go` - Rate limiting
- `RateLimiter` - Per-IP rate limiter with token bucket algorithm
- `NewRateLimiter(rate, burst)` - Create rate limiter
- `Middleware(next)` - Rate limiting middleware
- `getLimiter(ip)` - Get/create limiter for IP
- `getIP(r)` - Extract IP from request (handles proxies)
- `cleanupLimiters()` - Background cleanup goroutine

#### `middleware.go` - Shared utilities
- `authcontext` - Type for authentication context key
- `AuthRequest(w, r, c)` - Extract and validate session cookie
- Used by `HTTPAuthenticationMiddleware`

#### `session.go` - Session validation
- `UserIsAuth(r, cookie, c)` - Validate session cookie and return user
- Used by `AuthRequest`

#### `cookies.go` - Cookie utilities
- Cookie creation and management functions
- Used by authentication handlers

### SSH Middleware

#### `ssh_middleware.go` - Core SSH middleware components
- `SSHAuthenticationMiddleware(c)` - Validates SSH keys, populates context
- `SSHAuthorizationMiddleware(c)` - Ensures user is authenticated or needs provisioning
- `SSHSessionLoggingMiddleware(c)` - Logs SSH sessions
- `SSHRateLimitMiddleware(c)` - Placeholder for SSH rate limiting
- `FingerprintSHA256(key)` - Generate SHA256 fingerprint
- `ParseSSHPublicKey(str)` - Parse SSH public key string

#### `ssh_public_key_handler.go` - SSH key validation callback
- `PublicKeyHandler(c)` - Callback for Wish's `WithPublicKeyAuth`
- Validates key structure, allows all valid keys through
- Authentication happens in middleware chain

## Usage Examples

### HTTP

```go
// Global middleware (in main.go)
mux := router.NewModulacmsMux(*configuration)
middlewareHandler := middleware.DefaultMiddlewareChain(configuration)(mux)

// Protected endpoint (in router)
mux.Handle("POST /api/v1/ssh-keys",
    middleware.AuthenticatedChain(&c)(http.HandlerFunc(AddSSHKeyHandler))
)

// Auth endpoint (in router)
mux.Handle("POST /api/v1/auth/login",
    middleware.AuthEndpointChain(&c)(http.HandlerFunc(LoginHandler))
)

// Custom chain
customChain := middleware.Chain(
    middleware.HTTPLoggingMiddleware(),
    middleware.CorsMiddleware(&c),
    customRateLimiter.Middleware,
)
mux.Handle("/custom", customChain(handler))
```

### SSH

```go
// SSH server (in main.go)
sshServer, err := wish.NewServer(
    wish.WithAddress(net.JoinHostPort(host, port)),
    wish.WithHostKeyPath(".ssh/id_ed25519"),
    wish.WithPublicKeyAuth(middleware.PublicKeyHandler(configuration)),
    wish.WithMiddleware(
        middleware.SSHSessionLoggingMiddleware(configuration),
        middleware.SSHAuthenticationMiddleware(configuration),
        middleware.SSHAuthorizationMiddleware(configuration),
        cli.CliMiddleware(app.VerboseFlag, configuration),
        logging.Middleware(),
    ),
)
```

## Middleware Execution Order

### HTTP Request Flow

```
HTTP Request
    ↓
HTTPLoggingMiddleware (start timer, log request)
    ↓
CorsMiddleware (set CORS headers)
    ↓
HTTPAuthenticationMiddleware (validate session, populate context)
    ↓
HTTPPublicEndpointMiddleware (check if public, require auth for API)
    ↓
Application Handler (mux)
    ↓
HTTPLoggingMiddleware (log response, duration)
```

### SSH Connection Flow

```
SSH Connection
    ↓
PublicKeyHandler (validate key structure) → returns true
    ↓
SSHSessionLoggingMiddleware (log connection start)
    ↓
SSHAuthenticationMiddleware (lookup user by fingerprint, set context)
    ↓
SSHAuthorizationMiddleware (check auth or provisioning)
    ↓
CliMiddleware (launch TUI or provisioning wizard)
    ↓
logging.Middleware (framework logging)
    ↓
Session ends → SSHSessionLoggingMiddleware (log connection end)
```

## Context Values

### HTTP

| Key | Type | Set By | Used By |
|-----|------|--------|---------|
| `"authenticated"` | `*db.Users` | HTTPAuthenticationMiddleware | All handlers |

### SSH

| Key | Type | Set By | Used By |
|-----|------|--------|---------|
| `authenticated` | `bool` | SSHAuthenticationMiddleware | Authorization, CLI |
| `needs_provisioning` | `bool` | SSHAuthenticationMiddleware | Authorization, CLI |
| `user` | `*db.Users` | SSHAuthenticationMiddleware | CLI |
| `user_id` | `int64` | SSHAuthenticationMiddleware | CLI |
| `ssh_fingerprint` | `string` | SSHAuthenticationMiddleware | CLI (provisioning) |
| `ssh_key_type` | `string` | SSHAuthenticationMiddleware | CLI (provisioning) |
| `ssh_public_key` | `string` | SSHAuthenticationMiddleware | CLI (provisioning) |

## Key Principles

1. **Single Responsibility** - Each middleware does one thing well
2. **Composable** - Middleware can be easily combined and reordered
3. **No Legacy Code** - All old monolithic middleware removed
4. **Type-Safe** - Compile-time guarantees
5. **Testable** - Each middleware can be tested independently
6. **Well-Documented** - Clear purpose and usage for each component
7. **Consistent** - Same pattern for HTTP and SSH
8. **Clean Separation** - Authentication vs Authorization as separate steps
