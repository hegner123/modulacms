//go:build dev

package admin

import (
	"io/fs"
	"net/http"
)

// StaticFS returns the filesystem-backed static file system for live CSS/JS iteration.
func StaticFS() (http.FileSystem, error) {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}

// CacheControl is a no-op in dev mode so browsers always fetch fresh assets.
func CacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		next.ServeHTTP(w, r)
	})
}
