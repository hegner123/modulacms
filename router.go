package main

import (
	"net/http"
	"os"
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
	case checkPath(segments, ENDPOINT, "state"):
		stateHandler(w, r,segments)
	default:
		handleClientRoutes(w, r)
	}
}

func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	wd, err := os.Getwd()
	if err != nil {
		logError("failed to : ", err)
	}
	filePath := r.URL.Path
	segments := strings.Split(filePath, "/")

	if len(segments) > 1 {
		segments = segments[2:]
	}
	trimmedPath := wd + "/" + strings.Join(segments, "/")
	file, err := os.Open(trimmedPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, _ := file.Stat()
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}
