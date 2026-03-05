package middleware

import (
	"context"
	"net"
	"net/http"
)

type clientIPKey struct{}

// ClientIPMiddleware resolves the client IP once per request and stores it
// in context. Checks X-Forwarded-For and X-Real-IP headers for proxied
// requests, then falls back to net.SplitHostPort(r.RemoteAddr).
func ClientIPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := resolveClientIP(r)
			ctx := context.WithValue(r.Context(), clientIPKey{}, ip)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClientIPFromContext extracts the client IP from the context.
// Returns an empty string if not present.
func ClientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}

// resolveClientIP extracts the client IP from proxy headers or RemoteAddr.
func resolveClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For may contain multiple IPs; the first is the client
		for i := range len(forwarded) {
			if forwarded[i] == ',' {
				return forwarded[:i]
			}
		}
		return forwarded
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
