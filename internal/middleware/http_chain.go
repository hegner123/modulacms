package middleware

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"golang.org/x/time/rate"
)

// DefaultMiddlewareChain returns the standard middleware chain for the application.
// This includes: logging, CORS, authentication, and public endpoint protection.
func DefaultMiddlewareChain(c *config.Config) func(http.Handler) http.Handler {
	return Chain(
		HTTPLoggingMiddleware(),                  // 1. Request/response logging
		CorsMiddleware(c),                        // 2. CORS headers
		HTTPAuthenticationMiddleware(c),          // 3. Session authentication
		HTTPPublicEndpointMiddleware(c),          // 4. Public endpoint protection
	)
}

// AuthenticatedChain returns middleware for authenticated-only endpoints.
// Use this for endpoints that absolutely require authentication.
func AuthenticatedChain(c *config.Config) func(http.Handler) http.Handler {
	return Chain(
		HTTPLoggingMiddleware(),                  // 1. Request/response logging
		CorsMiddleware(c),                        // 2. CORS headers
		HTTPAuthenticationMiddleware(c),          // 3. Session authentication
		HTTPAuthorizationMiddleware(c),           // 4. Require authentication
	)
}

// AuthEndpointChain returns middleware for auth endpoints (login, register, etc).
// Includes rate limiting to prevent brute force attacks.
func AuthEndpointChain(c *config.Config) func(http.Handler) http.Handler {
	authLimiter := NewRateLimiter(rate.Limit(10.0/60.0), 10) // 10 req/min

	return Chain(
		HTTPLoggingMiddleware(),                  // 1. Request/response logging
		CorsMiddleware(c),                        // 2. CORS headers
		authLimiter.Middleware,                   // 3. Rate limiting
	)
}

// PublicAPIChain returns middleware for public API endpoints.
// No authentication required, but includes logging and CORS.
func PublicAPIChain(c *config.Config) func(http.Handler) http.Handler {
	return Chain(
		HTTPLoggingMiddleware(),                  // 1. Request/response logging
		CorsMiddleware(c),                        // 2. CORS headers
	)
}
