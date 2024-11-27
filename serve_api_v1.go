package main

import (
	"fmt"
	"net/http"
)

func apiRoutes(w http.ResponseWriter, r *http.Request) {
	apiRoute, err := stripAPIPath(r.URL.Path)
	if err != nil {
		fmt.Print("UM, this ain't a url bud.")
		fmt.Printf("\nerror: %s", err)
		return
	}

	switch {
	case r.Method == http.MethodDelete:
		apiDeleteHandler(w, r, apiRoute)
	case r.Method == http.MethodGet:
		apiGetHandler(w, r, apiRoute)
	case r.Method == http.MethodPost:
		apiPostHandler(w, r, apiRoute)
	case r.Method == http.MethodPut:
		apiPutHandler(w, r, apiRoute)
	case matchesPath(apiRoute, "admin/auth"):
		err := r.ParseForm()
		if err != nil {
			logError("failed to ParseForm ", err)
		}
		// status, err := handleAuth(r.Form)
		if err != nil {
			logError("failed to handle auth: ", err)
		}
		w.Header().Set("Content-Type", "application/json")
	}
}
