package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// CSRFContextKey is the context key for CSRF tokens.
// Exported so handlers can read the token from context.
type CSRFContextKey struct{}

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CSRFMiddleware implements double-submit cookie CSRF protection.
// For GET/HEAD/OPTIONS: reuses the existing cookie token if present, otherwise
// generates a new one. Reusing avoids invalidating the <meta> tag token during
// HTMX partial navigations (cookie updates but <head> doesn't refresh).
// For POST/PUT/PATCH/DELETE: validates that the token from cookie matches header or form field.
func CSRFMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				// Reuse existing CSRF cookie if present so SPA/HTMX
				// partial navigations don't desync cookie vs meta tag.
				var token string
				if existing, cookieErr := r.Cookie("csrf_token"); cookieErr == nil && existing.Value != "" {
					token = existing.Value
				} else {
					var genErr error
					token, genErr = generateCSRFToken()
					if genErr != nil {
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}
				}
				http.SetCookie(w, &http.Cookie{
					Name:     "csrf_token",
					Value:    token,
					Path:     "/admin/",
					HttpOnly: false, // JS needs to read this
					Secure:   r.TLS != nil,
					SameSite: http.SameSiteStrictMode,
				})
				ctx := context.WithValue(r.Context(), CSRFContextKey{}, token)
				next.ServeHTTP(w, r.WithContext(ctx))
			default:
				// Mutating request: validate token
				cookieToken := ""
				if c, err := r.Cookie("csrf_token"); err == nil {
					cookieToken = c.Value
				}
				if cookieToken == "" {
					http.Error(w, "Forbidden: missing CSRF token", http.StatusForbidden)
					return
				}
				// Check header first, then form field
				requestToken := r.Header.Get("X-CSRF-Token")
				if requestToken == "" {
					// Try form field (ParseForm is idempotent)
					if err := r.ParseForm(); err == nil {
						requestToken = r.FormValue("_csrf")
					}
				}
				if requestToken == "" || requestToken != cookieToken {
					http.Error(w, "Forbidden: invalid CSRF token", http.StatusForbidden)
					return
				}
				// Store token in context for templates that need it
				ctx := context.WithValue(r.Context(), CSRFContextKey{}, cookieToken)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}
