# Middleware

Composable middleware for HTTP and SSH servers.

## HTTP Middleware

### Pre-built Chains

```go
// Global - used in main.go
middleware.DefaultMiddlewareChain(mgr, pc)(mux)

// Protected endpoints (requires auth)
middleware.AuthenticatedChain(mgr)(handler)

// Auth endpoints (login, register) - includes rate limiting
middleware.AuthEndpointChain(mgr)(handler)

// Public API endpoints
middleware.PublicAPIChain(mgr)(handler)
```

### Components

- `RecoveryMiddleware()` - Panic recovery and error capture
- `RequestIDMiddleware()` - UUID per-request ID generation
- `HTTPLoggingMiddleware()` - Request/response logging
- `HTTPMetricsMiddleware()` - Request metrics recording
- `HTTPAuthenticationMiddleware(c)` - Session validation, populates context
- `HTTPAuthorizationMiddleware(c)` - Requires authentication
- `CorsMiddleware(c)` - CORS headers
- `ClientIPMiddleware()` - Client IP resolution
- `UserAgentMiddleware()` - User-Agent parsing
- `RateLimiter` - Per-IP rate limiting

## Authorization (RBAC)

```go
// Permission cache (loaded at startup, refreshed every 60s)
pc := middleware.NewPermissionCache()
pc.Load(driver)
pc.StartPeriodicRefresh(ctx, driver, 60*time.Second)

// Per-route permission middleware
middleware.RequirePermission("content:read")(handler)
middleware.RequireResourcePermission("content")(handler)  // auto-maps HTTP method
middleware.RequireAnyPermission("content:read", "media:read")(handler)
middleware.RequireAllPermissions("content:read", "content:update")(handler)

// Context helpers
middleware.ContextPermissions(ctx)
middleware.ContextIsAdmin(ctx)
middleware.ValidatePermissionLabel("content:read")
```

## SSH Middleware

```go
wish.WithMiddleware(
    middleware.SSHSessionLoggingMiddleware(c),
    middleware.SSHAuthenticationMiddleware(c),
    middleware.SSHAuthorizationMiddleware(c),
    cli.CliMiddleware(v, c),
    logging.Middleware(),
)
```

### Components

- `SSHSessionLoggingMiddleware(c)` - Logs SSH connections
- `SSHAuthenticationMiddleware(c)` - Validates keys, populates context
- `SSHAuthorizationMiddleware(c)` - Checks auth or provisioning
- `PublicKeyHandler(c)` - Validates public key structure

## Custom Chains

```go
middleware.Chain(
    middleware.HTTPLoggingMiddleware(),
    middleware.CorsMiddleware(c),
    customMiddleware,
)(handler)
```

See godoc for detailed API documentation.
