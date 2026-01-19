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
// user if valid. Returns nil values if authentication fails or is not required.
func AuthRequest(w http.ResponseWriter, r *http.Request, c *config.Config) (*authcontext, *db.Users) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return nil, nil
	}
	var u authcontext = "authenticated"
	cookie, err := r.Cookie(c.Cookie_Name)
	if err != nil {
		utility.DefaultLogger.Finfo("cookie not found", err)
		return nil, nil
	}

	user, err := UserIsAuth(r, cookie, c)
	if err != nil {
		return nil, nil
	}
	return &u, user
}
