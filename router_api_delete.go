package main

import (
	"net/http"
)

func apiDeleteHandler(w http.ResponseWriter, r *http.Request, segments []string) {
	switch {
	case checkPath(segments, DBMETHOD, "adminroute"):
		err := apiDeleteAdminRoute(w, r)
		if err != nil {
			logError("failed to delete adminroute", err)
		}
	case checkPath(segments, DBMETHOD, "datatype"):
		err := apiDeleteDataType(w, r)
		if err != nil {
			logError("failed to delete datatype", err)
		}
	case checkPath(segments, DBMETHOD, "field"):
		err := apiDeleteField(w, r)
		if err != nil {
			logError("failed to delete field", err)
		}
	case checkPath(segments, DBMETHOD, "media"):
		err := apiDeleteMedia(w, r)
		if err != nil {
			logError("failed to delete media", err)
		}
	case checkPath(segments, DBMETHOD, "mediadimension"):
		err := apiDeleteMediaDimension(w, r)
		if err != nil {
			logError("failed to delete mediadimension", err)
		}
	case checkPath(segments, DBMETHOD, "route"):
		err := apiDeleteRoute(w, r)
		if err != nil {
			logError("failed to delete route", err)
		}
	case checkPath(segments, DBMETHOD, "table"):
		err := apiDeleteTable(w, r)
		if err != nil {
			logError("failed to delete table", err)
		}
	case checkPath(segments, DBMETHOD, "token"):
		err := apiDeleteToken(w, r)
		if err != nil {
			logError("failed to delete token", err)
		}
	case checkPath(segments, DBMETHOD, "user"):
		err := apiDeleteUser(w, r)
		if err != nil {
			logError("failed to delete user", err)
		}
	}
}
