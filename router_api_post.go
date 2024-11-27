package main

import (
	"fmt"
	"net/http"
)

func apiPostHandler(w http.ResponseWriter, r *http.Request, apiRoute string) {
	postRoute, err := stripCreatePath(r.URL.Path)
	if err != nil {
		fmt.Print("UM, this ain't a url bud.")
		fmt.Printf("\nerror: %s", err)
		return
	}
	switch {
	case matchesPath(postRoute, "adminroute"):
		apiCreateAdminRoute(w, r)
	case matchesPath(postRoute, "datatype"):
		apiCreateDatatype(w, r)
	case matchesPath(postRoute, "field"):
		apiCreateField(w, r)
	case matchesPath(postRoute, "media"):
		apiCreateMedia(w, r)
	case matchesPath(postRoute, "mediadimension"):
		apiCreateMediaDimension(w, r)
	case matchesPath(postRoute, "route"):
		apiCreateRoute(w, r)
	case matchesPath(postRoute, "table"):
		apiCreateTable(w, r)
	case matchesPath(postRoute, "token"):
		apiCreateToken(w, r)
	case matchesPath(postRoute, "user"):
		apiCreateUser(w, r)
	}
}
