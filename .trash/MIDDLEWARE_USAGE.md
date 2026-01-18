## Composable Middleware Usage in ModulaCMS Router

This guide shows how to use the new composable middleware system for HTTP routes.

### Old Pattern (Deprecated)

```go
// ❌ Hard to read, manually nested
mux.Handle("POST /api/v1/auth/login",
    corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        LoginHandler(w, r, c)
    })))
)
```

### New Pattern (Recommended)

```go
// ✅ Clear, composable, easy to modify
mux.Handle("POST /api/v1/auth/login",
    middleware.AuthEndpointChain(&c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        LoginHandler(w, r, c)
    }))
)
```

## Available Middleware Chains

### 1. AuthEndpointChain - For login, register, password reset
Includes: Logging, CORS, Rate Limiting (10 req/min)

```go
mux.Handle("POST /api/v1/auth/login",
    middleware.AuthEndpointChain(&c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        LoginHandler(w, r, c)
    }))
)
```

### 2. AuthenticatedChain - For endpoints requiring authentication
Includes: Logging, CORS, Authentication, Authorization

```go
mux.Handle("POST /api/v1/ssh-keys",
    middleware.AuthenticatedChain(&c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        AddSSHKeyHandler(w, r, c)
    }))
)
```

### 3. PublicAPIChain - For public API endpoints
Includes: Logging, CORS

```go
mux.Handle("GET /api/v1/content",
    middleware.PublicAPIChain(&c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ContentHandler(w, r, c)
    }))
)
```

### 4. DefaultMiddlewareChain - Global middleware (used by Serve)
Includes: Logging, CORS, Authentication, Public Endpoint Protection

```go
// This is what Serve() uses internally
handler := middleware.DefaultMiddlewareChain(&c)(mux)
```

## Custom Middleware Chains

Build your own chains using the `Chain` helper:

```go
import "github.com/hegner123/modulacms/internal/middleware"

// Custom chain with custom rate limiting
customRateLimiter := middleware.NewRateLimiter(rate.Limit(5.0/60.0), 5) // 5 req/min

customChain := middleware.Chain(
    middleware.HTTPLoggingMiddleware(),
    middleware.CorsMiddleware(&c),
    customRateLimiter.Middleware,
    middleware.HTTPAuthenticationMiddleware(&c),
    middleware.HTTPAuthorizationMiddleware(&c),
)

mux.Handle("POST /api/v1/sensitive-operation",
    customChain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        SensitiveOperationHandler(w, r, c)
    }))
)
```

## Individual Middleware Components

### HTTPLoggingMiddleware
Logs all HTTP requests and responses with duration

```go
middleware.HTTPLoggingMiddleware()
```

### CorsMiddleware
Adds CORS headers based on config

```go
middleware.CorsMiddleware(&c)
```

### HTTPAuthenticationMiddleware
Validates session cookies and populates context with user

```go
middleware.HTTPAuthenticationMiddleware(&c)
```

### HTTPAuthorizationMiddleware
Blocks requests without authenticated user

```go
middleware.HTTPAuthorizationMiddleware(&c)
```

### RateLimiter.Middleware
Per-IP rate limiting with token bucket algorithm

```go
limiter := middleware.NewRateLimiter(rate.Limit(10.0/60.0), 10)
limiter.Middleware
```

## Complete Router Example

```go
func NewModulacmsMux(c config.Config) *http.ServeMux {
    mux := http.NewServeMux()

    // Auth endpoints (public with rate limiting)
    mux.Handle("POST /api/v1/auth/login",
        middleware.AuthEndpointChain(&c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            LoginHandler(w, r, c)
        }))
    )

    // SSH key management (requires authentication)
    mux.Handle("POST /api/v1/ssh-keys",
        middleware.AuthenticatedChain(&c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            AddSSHKeyHandler(w, r, c)
        }))
    )

    mux.Handle("GET /api/v1/ssh-keys",
        middleware.AuthenticatedChain(&c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ListSSHKeysHandler(w, r, c)
        }))
    )

    mux.Handle("DELETE /api/v1/ssh-keys/",
        middleware.AuthenticatedChain(&c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            DeleteSSHKeyHandler(w, r, c)
        }))
    )

    // Public content endpoints
    mux.Handle("GET /api/v1/content",
        middleware.PublicAPIChain(&c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ContentHandler(w, r, c)
        }))
    )

    return mux
}
```

## Migration Strategy

1. **Identify endpoint types** (auth, authenticated, public)
2. **Apply appropriate chain** to each route
3. **Remove manual middleware nesting**
4. **Test authentication flows**

## Benefits

✅ **Clear and readable** - No nested parentheses
✅ **Consistent** - Same middleware on similar endpoints
✅ **Composable** - Easy to add/remove/reorder middleware
✅ **Testable** - Each middleware can be tested independently
✅ **Type-safe** - Compile-time guarantees
✅ **Documented** - Each chain has a clear purpose
