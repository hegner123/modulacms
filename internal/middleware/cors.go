package middleware

import (
	"net/http"
	"slices"
	"strings"

	config "github.com/hegner123/modulacms/internal/config"
)

// CorsHandler sets CORS headers based on configuration
func Cors(w http.ResponseWriter, r *http.Request, c *config.Config) {
	CorsWithConfig(w, r, c)
}

// CorsWithConfig allows specifying a config for CORS settings
func CorsWithConfig(w http.ResponseWriter, r *http.Request, c *config.Config) {
	// Get origin from request
	origin := r.Header.Get("Origin")

	// Check if the origin is allowed
	allowedOrigin := getAllowedOrigin(origin, c.Cors_Origins)
	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	}

	// Set allowed methods
	if len(c.Cors_Methods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.Cors_Methods, ", "))
	}

	// Set allowed headers
	if len(c.Cors_Headers) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.Cors_Headers, ", "))
	}

	// Set credentials allowed if configured
	if c.Cors_Credentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// For preflight requests, respond with no content
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}

// getAllowedOrigin returns the allowed origin if the request origin is permitted
// Returns empty string if origin is not allowed
func getAllowedOrigin(requestOrigin string, allowedOrigins []string) string {
	// If no origins specified or empty list, don't set any CORS headers
	if len(allowedOrigins) == 0 {
		return ""
	}

	// Check for wildcard
	if slices.Contains(allowedOrigins, "*") {
		return "*"
	}

	// Check if the request origin matches any of the allowed origins
	if slices.Contains(allowedOrigins, requestOrigin) {
		return requestOrigin
	}

	// Origin not allowed
	return ""
}
