# Middleware

Composable middleware for HTTP and SSH servers.

## HTTP Middleware

### Pre-built Chains

```go
// Global - used in main.go
middleware.DefaultMiddlewareChain(c)(mux)

// Protected endpoints (requires auth)
middleware.AuthenticatedChain(c)(handler)

// Auth endpoints (login, register) - includes rate limiting
middleware.AuthEndpointChain(c)(handler)

// Public API endpoints
middleware.PublicAPIChain(c)(handler)
```

### Components

- `HTTPLoggingMiddleware()` - Request/response logging
- `HTTPAuthenticationMiddleware(c)` - Session validation, populates context
- `HTTPAuthorizationMiddleware(c)` - Requires authentication
- `CorsMiddleware(c)` - CORS headers
- `RateLimiter` - Per-IP rate limiting

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
