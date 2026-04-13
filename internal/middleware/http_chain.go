package middleware

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/time/rate"
)

// DefaultMiddlewareChain returns the standard middleware chain for the application.
// This includes: logging, CORS, authentication, public endpoint protection, and
// permission injection. PermissionInjector is added here (not in AuthenticatedChain)
// to avoid double injection since DefaultMiddlewareChain wraps the entire mux.
// Accepts *config.Manager for hot-reloadable config access and *utility.ObservabilityClient
// for provider-specific HTTP transaction tracing.
func DefaultMiddlewareChain(mgr *config.Manager, pc *PermissionCache, obs *utility.ObservabilityClient) func(http.Handler) http.Handler {
	cfg, err := mgr.Config()
	if err != nil {
		// Fallback: return a chain that rejects all requests.
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			})
		}
	}

	// The observability middleware is provider-specific (Sentry creates
	// transactions, console is a pass-through). Selected at startup via the
	// observability_provider config field.
	var obsMw func(http.Handler) http.Handler
	if obs != nil {
		obsMw = obs.HTTPMiddleware()
	} else {
		obsMw = func(next http.Handler) http.Handler { return next }
	}

	return Chain(
		RecoveryMiddleware(obs),           // 1. Panic recovery + error capture
		obsMw,                             // 2. Observability transaction tracing
		RequestIDMiddleware(),             // 3. Request ID generation
		ClientIPMiddleware(),              // 4. Client IP resolution
		UserAgentMiddleware(),             // 5. User-Agent + Client Hints parsing
		HTTPLoggingMiddleware(),           // 6. Request/response logging
		HTTPMetricsMiddleware(),           // 7. Request metrics recording
		CorsMiddleware(cfg),               // 8. CORS headers
		HTTPAuthenticationMiddleware(cfg), // 9. Session authentication
		HTTPPublicEndpointMiddleware(cfg), // 10. Public endpoint protection
		PermissionInjector(pc),            // 11. Permission set injection
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
		RequestIDMiddleware(),             // 1. Request ID generation
		ClientIPMiddleware(),              // 2. Client IP resolution
		UserAgentMiddleware(),             // 3. User-Agent + Client Hints parsing
		HTTPLoggingMiddleware(),           // 4. Request/response logging
		CorsMiddleware(cfg),               // 5. CORS headers
		HTTPAuthenticationMiddleware(cfg), // 6. Session authentication
		HTTPAuthorizationMiddleware(cfg),  // 7. Require authentication
	)
}

// AuthEndpointChain returns middleware for auth endpoints (login, register, etc).
// Includes rate limiting to prevent brute force attacks.
func AuthEndpointChain(c *config.Config) func(http.Handler) http.Handler {
	authLimiter := NewRateLimiter(rate.Limit(10.0/60.0), 10) // 10 req/min

	return Chain(
		RequestIDMiddleware(),   // 1. Request ID generation
		ClientIPMiddleware(),    // 2. Client IP resolution
		UserAgentMiddleware(),   // 3. User-Agent + Client Hints parsing
		HTTPLoggingMiddleware(), // 4. Request/response logging
		CorsMiddleware(c),       // 5. CORS headers
		authLimiter.Middleware,  // 6. Rate limiting
	)
}

// PublicAPIChain returns middleware for public API endpoints.
// No authentication required, but includes logging and CORS.
func PublicAPIChain(c *config.Config) func(http.Handler) http.Handler {
	return Chain(
		RequestIDMiddleware(),   // 1. Request ID generation
		ClientIPMiddleware(),    // 2. Client IP resolution
		UserAgentMiddleware(),   // 3. User-Agent + Client Hints parsing
		HTTPLoggingMiddleware(), // 4. Request/response logging
		CorsMiddleware(c),       // 5. CORS headers
	)
}
