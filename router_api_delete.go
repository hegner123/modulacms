package main

import (
	"fmt"
	"net/http"
)

func apiDeleteHandler(w http.ResponseWriter, r *http.Request, apiRoute string) {
	deleteRoute, err := stripDeletePath(r.URL.Path)
	if err != nil {
		fmt.Print("UM, this ain't a url bud.")
		fmt.Printf("\nerror: %s", err)
		return
	}
	switch {
	case matchesPath(deleteRoute, "adminroute"):
		err := apiDeleteAdminRoute(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(deleteRoute, "datatype"):
		err := apiDeleteDataType(w, r)
		if err != nil {
			logError("failed to delete datatype: ", err)
		}
	case matchesPath(deleteRoute, "field"):
		err := apiDeleteField(w, r)
		if err != nil {
			logError("failed to delete field: ", err)
		}
	case matchesPath(deleteRoute, "media"):
		err := apiDeleteMedia(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(deleteRoute, "mediadimension"):
		err := apiDeleteMediaDimension(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(deleteRoute, "route"):
		err := apiDeleteRoute(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(deleteRoute, "table"):
		err := apiDeleteTable(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(deleteRoute, "token"):
		err := apiDeleteToken(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(deleteRoute, "user"):
		err := apiDeleteUser(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	}
}
