// Package middleware provides HTTP middleware components for the ModulaCMS web application.
//
// It includes middleware for authentication, CORS handling, and session management,
// which can be composed in an HTTP handler chain. The package is responsible for
// pre-processing HTTP requests before they reach the application's core handlers.
package middleware

import (
	"net/http"
	"strings"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// authcontext is a type for storing authentication information in request context
type authcontext string

// AuthRequest extracts and validates authentication information from the request.
// It retrieves the authentication cookie, verifies it, and returns the authenticated
// user if valid. Falls back to API key authentication via the Authorization header
// when cookie auth is not present or fails.
// Returns nil values if all authentication methods fail or the request is not authenticated.
func AuthRequest(w http.ResponseWriter, r *http.Request, c *config.Config) (*authcontext, *db.Users) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return nil, nil
	}
	var u authcontext = "authenticated"

	// Try cookie auth first
	cookie, err := r.Cookie(c.Cookie_Name)
	if err == nil {
		user, err := UserIsAuth(r, cookie, c)
		if err == nil {
			return &u, user
		}
	}

	// Fall back to API key auth
	return APIKeyAuth(r, c)
}

// APIKeyAuth authenticates a request using an API key from the Authorization header.
// It expects the header format "Bearer <key>", looks up the token in the database,
// and validates that the token is of type "api_key", is not revoked, and has not expired.
// Returns the authenticated context and user on success, or nil values on failure.
func APIKeyAuth(r *http.Request, c *config.Config) (*authcontext, *db.Users) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, nil
	}

	key := strings.TrimPrefix(authHeader, "Bearer ")
	if key == "" {
		return nil, nil
	}

	dbc := db.ConfigDB(*c)

	token, err := dbc.GetTokenByTokenValue(key)
	if err != nil {
		utility.DefaultLogger.Finfo("api key lookup failed", err)
		return nil, nil
	}

	if token.TokenType != "api_key" && token.TokenType != "plugin_api_key" {
		utility.DefaultLogger.Finfo("token is not an api_key", token.TokenType)
		return nil, nil
	}

	if token.Revoked {
		utility.DefaultLogger.Finfo("api key is revoked", token.ID)
		return nil, nil
	}

	expired := utility.TimestampLessThan(token.ExpiresAt.String())
	if expired {
		utility.DefaultLogger.Finfo("api key is expired", token.ID)
		return nil, nil
	}

	if !token.UserID.Valid {
		utility.DefaultLogger.Finfo("api key has no associated user", token.ID)
		return nil, nil
	}

	user, err := dbc.GetUser(token.UserID.ID)
	if err != nil {
		utility.DefaultLogger.Finfo("failed to get user for api key", err)
		return nil, nil
	}

	var u authcontext = "authenticated"
	return &u, user
}
