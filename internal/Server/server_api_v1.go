package api_v1

import (
	"net/http"

	config "github.com/hegner123/modulacms/internal/Config"
)

type (
	Segment         int
	UrlSegments     []string
	AuthHandlerFunc func(http.ResponseWriter, *http.Request)
	ApiServerV1     struct {
		Config        config.Config
		UrlSegments   []string
		AuthHandler   AuthHandlerFunc
		DeleteHandler func(http.ResponseWriter, *http.Request, UrlSegments, config.Config)
		GetHandler    func(http.ResponseWriter, *http.Request, UrlSegments, config.Config)
		PostHandler   func(http.ResponseWriter, *http.Request, UrlSegments, config.Config)
		PutHandler    func(http.ResponseWriter, *http.Request, UrlSegments, config.Config)
	}
)

const AUTHROUTE Segment = 3

func (server ApiServerV1) Routes(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodDelete:
		server.DeleteHandler(w, r, server.UrlSegments, server.Config)
	case r.Method == http.MethodGet:
		server.GetHandler(w, r, server.UrlSegments, server.Config)
	case r.Method == http.MethodPost:
		server.PostHandler(w, r, server.UrlSegments, server.Config)
	case r.Method == http.MethodPut:
		server.PutHandler(w, r, server.UrlSegments, server.Config)
	}
}

/*
func checkPath(segments []string, index Segment, ref string) bool {
    if len(segments)<=int(index){
	fmt.Println("index out of range")
	fmt.Println(len(segments))
        return false
    }
	fmt.Println("check path")
	fmt.Println(index)
	fmt.Printf("\nSegments %v\n", segments[index])
	return segments[index] == ref
}
*/
