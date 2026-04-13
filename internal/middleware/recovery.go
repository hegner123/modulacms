package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/hegner123/modulacms/internal/utility"
)

// RecoveryMiddleware catches panics in HTTP handlers, reports them via the
// observability provider, and returns a 500 response. When an ObservabilityClient
// is provided, uses CaptureRequestError for per-request context enrichment
// (e.g. Sentry hub with transaction scope). Falls back to the global
// CaptureError when obs is nil.
func RecoveryMiddleware(obs *utility.ObservabilityClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rv := recover(); rv != nil {
					stack := string(debug.Stack())
					err := fmt.Errorf("panic: %v", rv)
					ctx := map[string]any{
						"handler": "http",
						"method":  r.Method,
						"path":    r.URL.Path,
						"stack":   stack,
					}

					if obs != nil {
						obs.CaptureRequestError(err, r, ctx)
					}
					utility.CaptureError(err, ctx)

					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
