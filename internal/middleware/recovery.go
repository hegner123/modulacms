package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/hegner123/modulacms/internal/utility"
)

// RecoveryMiddleware catches panics in HTTP handlers, reports them via
// CaptureError, and returns a 500 response. Placed early in the middleware
// chain so it wraps all downstream handlers.
func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rv := recover(); rv != nil {
					stack := string(debug.Stack())
					err := fmt.Errorf("panic: %v", rv)

					utility.CaptureError(err, map[string]any{
						"handler": "http",
						"method":  r.Method,
						"path":    r.URL.Path,
						"stack":   stack,
					})

					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
