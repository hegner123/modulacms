package middleware

import (
	"net/http"
	"slices"
	"strings"

	config "github.com/hegner123/modulacms/internal/config"
)

// CorsHandler sets CORS headers based on configuration
func Cors(w http.ResponseWriter, r *http.Request) {
	CorsWithConfig(w, r, config.Env)
}

// CorsWithConfig allows specifying a config for CORS settings
func CorsWithConfig(w http.ResponseWriter, r *http.Request, conf config.Config) {
	// Get origin from request
	origin := r.Header.Get("Origin")

	// Check if the origin is allowed
	allowedOrigin := getAllowedOrigin(origin, conf.Cors_Origins)
	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	}

	// Set allowed methods
	if len(conf.Cors_Methods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(conf.Cors_Methods, ", "))
	}

	// Set allowed headers
	if len(conf.Cors_Headers) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(conf.Cors_Headers, ", "))
	}

	// Set credentials allowed if configured
	if conf.Cors_Credentials {
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
