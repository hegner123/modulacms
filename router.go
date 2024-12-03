package main

import (
	"net/http"
	"path/filepath"
	"strings"
)

type Segment int

const (
	CLIENT    Segment = 2
	CROUTE    Segment = 3
	ADMIN     Segment = 2
	AROUTE    Segment = 3
	ENDPOINT  Segment = 1
	VERSION   Segment = 2
	DBMETHOD  Segment = 3
	TABLE     Segment = 4
	AUTHROUTE Segment = 3
)

func router(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")

	switch {
	case hasFileExtension(r.URL.Path):
		staticFileHandler(w, r)
	case checkPath(segments, ENDPOINT, "api"):
		apiRoutes(w, r, segments)
	case checkPath(segments, ENDPOINT, "admin"):
		handleAdminRoutes(w, r, segments)
	case checkPath(segments, ENDPOINT, "404"):
		notFoundHandler(w, r)
	default:
		handleClientRoutes(w, r)
	}
}

func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("public", r.URL.Path)
	switch {
	case filepath.Ext(filePath) == ".css":
		w.Header().Set("Content-Type", "text/css")
	case filepath.Ext(filePath) == ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case filepath.Ext(filePath) == ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case filepath.Ext(filePath) == ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case filepath.Ext(filePath) == ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case filepath.Ext(filePath) == ".bmp":
		w.Header().Set("Content-Type", "image/bmp")
	case filepath.Ext(filePath) == ".tiff":
		w.Header().Set("Content-Type", "image/tiff")
	case filepath.Ext(filePath) == ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case filepath.Ext(filePath) == ".heif":
		w.Header().Set("Content-Type", "image/heif")
	case filepath.Ext(filePath) == ".heic":
		w.Header().Set("Content-Type", "image/heic")
	case filepath.Ext(filePath) == ".png":
		w.Header().Set("Content-Type", "image/png")
	case filepath.Ext(filePath) == ".txt":
		w.Header().Set("Content-Type", "text/plain")

	}

	http.ServeFile(w, r, filePath)
}
