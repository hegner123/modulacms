package main

import (
	"net/http"
	"strings"
)

func stateHandler(w http.ResponseWriter, r *http.Request, segments []string) {
	if len(segments) < 3 {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}
	// Specfic Htmx Functions
	route := strings.Join(segments, "/")

    switch route{
    case "sidebar/dashboard/clicked":

    }
}
