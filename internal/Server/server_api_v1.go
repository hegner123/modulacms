package api_v1

import (
	"net/http"
)

func apiRoutes(w http.ResponseWriter, r *http.Request, urlSegments []string) {
	switch {
	case r.Method == http.MethodDelete:
		apiDeleteHandler(w, r, urlSegments)
	case r.Method == http.MethodGet:
		apiGetHandler(w, r, urlSegments)
	case r.Method == http.MethodPost:
		apiPostHandler(w, r, urlSegments)
	case r.Method == http.MethodPut:
		apiPutHandler(w, r, urlSegments)
	case checkPath(urlSegments, AUTHROUTE, "admin"):
		handleAdminAuth(w, r)
	}
}
