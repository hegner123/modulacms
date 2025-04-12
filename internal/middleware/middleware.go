// Package middleware provides HTTP middleware components for the ModulaCMS web application.
//
// It includes middleware for authentication, CORS handling, and session management,
// which can be composed in an HTTP handler chain. The package is responsible for
// pre-processing HTTP requests before they reach the application's core handlers.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// authcontext is a type for storing authentication information in request context
type authcontext string

// Serve is the main middleware handler that manages authentication and CORS
// for incoming HTTP requests. It authenticates requests, adds user data to the context
// for authenticated requests, and blocks unauthorized access to API endpoints.
func Serve(next http.Handler, c *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Cors(w, r, c)

		u, user := AuthRequest(w, r, c)
		if u != nil {
			// Inject authenticated user information into the request context for downstream handlers
			ctx := context.WithValue(r.Context(), u, user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if strings.Contains(r.URL.Path, "api") {
			w.WriteHeader(http.StatusUnauthorized)
			msg := fmt.Sprintf("Unauthorized Request to %s", string(r.URL.Path))
			_, err := w.Write([]byte(msg))
			if err != nil {
				utility.DefaultLogger.Error("", err)
				return
			}
			return
		}

		next.ServeHTTP(w, r)

	})
}

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
		utility.DefaultLogger.Info("cookie not found", err)
		return nil, nil
	}

	user, err := UserIsAuth(r, cookie, c)
	if err != nil {
		return nil, nil
	}
	return &u, user
}

// GetURLSegments splits a URL path into segments for routing and analysis.
// It separates the path by forward slashes and returns the resulting segments as a slice.
func GetURLSegments(path string) []string {
	return strings.Split(path, "/")
}

/*
func refreshTokenIfNeeded(t string) (*db.Users, error) {
	u := db.Users{
		Email: t,
	}

	return &u, nil

}
*/
