package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// HTTPLoggingMiddleware logs HTTP requests and responses
func HTTPLoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			utility.DefaultLogger.Finfo("HTTP request started", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)

			// Create a response wrapper to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			utility.DefaultLogger.Finfo("HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration", duration.String(),
			)
		})
	}
}

// HTTPAuthenticationMiddleware validates session cookies and populates request context
func HTTPAuthenticationMiddleware(c *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, user := AuthRequest(w, r, c)
			if authCtx != nil && user != nil {
				// Inject authenticated user into context
				// Use the dereferenced value as key so other middleware
				// can look it up with authcontext("authenticated")
				ctx := context.WithValue(r.Context(), *authCtx, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No authenticated user - continue without auth context
			next.ServeHTTP(w, r)
		})
	}
}

// HTTPAuthorizationMiddleware blocks unauthenticated requests to protected endpoints
// Use this on routes that require authentication
func HTTPAuthorizationMiddleware(c *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is authenticated
			var authCtx authcontext = "authenticated"
			user := r.Context().Value(authCtx)

			if user == nil {
				utility.DefaultLogger.Fwarn("Unauthorized HTTP request", nil, r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// PublicEndpoints lists API endpoints that don't require authentication
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

// HTTPPublicEndpointMiddleware allows public endpoints through, blocks others
// This is used as a global middleware to protect all /api/* routes by default
func HTTPPublicEndpointMiddleware(c *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Plugin routes handle their own auth -- the bridge enforces per-route
			// approval checks and authenticated/public route distinctions.
			if strings.HasPrefix(r.URL.Path, "/api/v1/plugins/") {
				next.ServeHTTP(w, r)
				return
			}


			// Allow public endpoints (exact match or with trailing slash)
			for _, endpoint := range PublicEndpoints {
				if r.URL.Path == endpoint || r.URL.Path == endpoint+"/" {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Allow non-API routes (like /)
			if !strings.HasPrefix(r.URL.Path, "/api") {
				next.ServeHTTP(w, r)
				return
			}

			// Check if authenticated for API routes
			var authCtx authcontext = "authenticated"
			user := r.Context().Value(authCtx)

			if user == nil {
				utility.DefaultLogger.Fwarn("Unauthorized API access", nil, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				// Encode error is non-recoverable (client disconnected or similar);
				// the response is already partially written so no recovery is possible.
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthenticatedUser extracts the authenticated user from the request context.
// Returns nil if no user is authenticated.
func AuthenticatedUser(ctx context.Context) *db.Users {
	var key authcontext = "authenticated"
	user, _ := ctx.Value(key).(*db.Users)
	return user
}

// SetAuthenticatedUser returns a new context with the given user set as the
// authenticated user. This uses the same unexported context key as the auth
// middleware, so AuthenticatedUser will find it. Intended for use by the
// plugin bridge tests and any other test that needs to simulate an authenticated
// request without running the full middleware chain.
func SetAuthenticatedUser(ctx context.Context, user *db.Users) context.Context {
	var key authcontext = "authenticated"
	return context.WithValue(ctx, key, user)
}

// Chain applies multiple middleware in sequence (left to right)
// Example: Chain(middleware1, middleware2, middleware3)(handler)
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the HTTP status code before delegating to the underlying ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
