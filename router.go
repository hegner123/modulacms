package main

import (
	"fmt"
	"net/http"
	"path/filepath"
)

func router(w http.ResponseWriter, r *http.Request) {
    fmt.Println(r.URL.Path)
	switch {
	case hasFileExtension(r.URL.Path):
		fmt.Print("static route\n")
		staticFileHandler(w, r)
	case checkPath(r.URL.Path, "api"):
		fmt.Print("api/v1 route\n")
		apiRoutes(w, r)
	case checkPath(r.URL.Path, "admin"):
		fmt.Print("admin route\n")
		handleAdminRoutes(w, r)
	case r.URL.Path == "/404":
		fmt.Print("404 route\n")
		notFoundHandler(w, r)
	default:
		fmt.Print("client route\n")
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
