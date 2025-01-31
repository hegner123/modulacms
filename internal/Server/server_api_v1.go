package api_v1

import (
	"net/http"
)

type (
	Segment            int
	UrlSegments        []string
	AuthHandlerFunc    func(http.ResponseWriter, *http.Request)
	ApiServerV1        struct {
		UrlSegments   []string
		AuthHandler   AuthHandlerFunc
		DeleteHandler func(http.ResponseWriter, *http.Request, UrlSegments)
		GetHandler    func(http.ResponseWriter, *http.Request, UrlSegments)
		PostHandler   func(http.ResponseWriter, *http.Request, UrlSegments)
		PutHandler    func(http.ResponseWriter, *http.Request, UrlSegments)

	}
)

const AUTHROUTE Segment = 3

func (server ApiServerV1) Routes(w http.ResponseWriter, r *http.Request) {
	switch {
	case checkPath(server.UrlSegments, AUTHROUTE, "admin"):
		server.AuthHandler(w, r)
	case r.Method == http.MethodDelete:
		server.DeleteHandler(w, r, server.UrlSegments)
	case r.Method == http.MethodGet:
		server.GetHandler(w, r, server.UrlSegments)
	case r.Method == http.MethodPost:
		server.PostHandler(w, r, server.UrlSegments)
	case r.Method == http.MethodPut:
		server.PutHandler(w, r, server.UrlSegments)
	}
}

func checkPath(segments []string, index Segment, ref string) bool {
	return segments[index] == ref
}
