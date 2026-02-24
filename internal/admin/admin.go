package admin

import (
	"io/fs"
	"net/http"
)

// StaticFS returns the embedded static file system for serving admin assets.
func StaticFS() (http.FileSystem, error) {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}

// CacheControl wraps a handler with aggressive caching headers for static assets.
func CacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
