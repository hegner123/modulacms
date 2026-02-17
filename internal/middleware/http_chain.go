package middleware

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"golang.org/x/time/rate"
)

// DefaultMiddlewareChain returns the standard middleware chain for the application.
// This includes: logging, CORS, authentication, public endpoint protection, and
// permission injection. PermissionInjector is added here (not in AuthenticatedChain)
// to avoid double injection since DefaultMiddlewareChain wraps the entire mux.
// Accepts *config.Manager for hot-reloadable config access.
func DefaultMiddlewareChain(mgr *config.Manager, pc *PermissionCache) func(http.Handler) http.Handler {
	cfg, err := mgr.Config()
	if err != nil {
		// Fallback: return a chain that rejects all requests.
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			})
		}
	}
	return Chain(
		RequestIDMiddleware(),                    // 1. Request ID generation
		HTTPLoggingMiddleware(),                  // 2. Request/response logging
		CorsMiddleware(cfg),                      // 3. CORS headers
		HTTPAuthenticationMiddleware(cfg),        // 4. Session authentication
		HTTPPublicEndpointMiddleware(cfg),        // 5. Public endpoint protection
		PermissionInjector(pc),                   // 6. Permission set injection
	)
}

// AuthenticatedChain returns middleware for authenticated-only endpoints.
// Use this for endpoints that absolutely require authentication.
// Accepts *config.Manager for hot-reloadable config access.
func AuthenticatedChain(mgr *config.Manager) func(http.Handler) http.Handler {
	cfg, err := mgr.Config()
	if err != nil {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			})
		}
	}
	return Chain(
		RequestIDMiddleware(),                    // 1. Request ID generation
		HTTPLoggingMiddleware(),                  // 2. Request/response logging
		CorsMiddleware(cfg),                      // 3. CORS headers
		HTTPAuthenticationMiddleware(cfg),        // 4. Session authentication
		HTTPAuthorizationMiddleware(cfg),         // 5. Require authentication
	)
}

// AuthEndpointChain returns middleware for auth endpoints (login, register, etc).
// Includes rate limiting to prevent brute force attacks.
func AuthEndpointChain(c *config.Config) func(http.Handler) http.Handler {
	authLimiter := NewRateLimiter(rate.Limit(10.0/60.0), 10) // 10 req/min

	return Chain(
		RequestIDMiddleware(),                    // 1. Request ID generation
		HTTPLoggingMiddleware(),                  // 2. Request/response logging
		CorsMiddleware(c),                        // 3. CORS headers
		authLimiter.Middleware,                   // 4. Rate limiting
	)
}

// PublicAPIChain returns middleware for public API endpoints.
// No authentication required, but includes logging and CORS.
func PublicAPIChain(c *config.Config) func(http.Handler) http.Handler {
	return Chain(
		RequestIDMiddleware(),                    // 1. Request ID generation
		HTTPLoggingMiddleware(),                  // 2. Request/response logging
		CorsMiddleware(c),                        // 3. CORS headers
	)
}
