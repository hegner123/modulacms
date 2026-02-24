package admin

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/middleware"
)

// AdminAuthMiddleware checks for an authenticated user in context.
// If no user is present, redirects to /admin/login with ?next= parameter.
// For HTMX requests, uses HX-Redirect header instead of 302.
func AdminAuthMiddleware(mgr *config.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := middleware.AuthenticatedUser(r.Context())
			if user == nil {
				nextPath := r.URL.Path
				if !strings.HasPrefix(nextPath, "/admin/") && nextPath != "/admin" {
					nextPath = "/admin/"
				}
				loginURL := "/admin/login?next=" + url.QueryEscape(nextPath)

				if r.Header.Get("HX-Request") != "" {
					w.Header().Set("HX-Redirect", loginURL)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Redirect(w, r, loginURL, http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
