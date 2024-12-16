package api_v1

import (
	"net/http"
)

type (
	Segment            int
	UrlSegments        []string
	RequestHandlerFunc func(http.ResponseWriter, *http.Request, UrlSegments)
	AuthHandlerFunc    func(http.ResponseWriter, *http.Request)
	ApiServerV1        struct {
		UrlSegments   []string
		AuthHandler   AuthHandlerFunc
		DeleteHandler RequestHandlerFunc
		GetHandler    RequestHandlerFunc
		PostHandler   RequestHandlerFunc
		PutHandler    RequestHandlerFunc
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
