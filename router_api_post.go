package main

import (
	"net/http"
)

func apiPostHandler(w http.ResponseWriter, r *http.Request, segments []string) {
	switch {
	case checkPath(segments, DBMETHOD, "adminroute"):
		apiCreateAdminRoute(w, r)
	case checkPath(segments, DBMETHOD, "datatype"):
		apiCreateDatatype(w, r)
	case checkPath(segments, DBMETHOD, "field"):
		apiCreateField(w, r)
	case checkPath(segments, DBMETHOD, "media"):
		apiCreateMedia(w, r)
	case checkPath(segments, DBMETHOD, "mediadimension"):
		apiCreateMediaDimension(w, r)
	case checkPath(segments, DBMETHOD, "route"):
		apiCreateRoute(w, r)
	case checkPath(segments, DBMETHOD, "table"):
		apiCreateTable(w, r)
	case checkPath(segments, DBMETHOD, "token"):
		apiCreateToken(w, r)
	case checkPath(segments, DBMETHOD, "user"):
		apiCreateUser(w, r)
	}
}
