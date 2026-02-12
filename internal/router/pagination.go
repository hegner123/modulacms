package router

import (
	"net/http"
	"strconv"

	"github.com/hegner123/modulacms/internal/db"
)

// ParsePaginationParams extracts limit and offset from query parameters.
// Defaults: limit=50, offset=0. Max limit=1000. Negative values are ignored.
func ParsePaginationParams(r *http.Request) db.PaginationParams {
	params := db.PaginationParams{
		Limit:  50,
		Offset: 0,
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			params.Limit = n
		}
	}

	if params.Limit > 1000 {
		params.Limit = 1000
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n >= 0 {
			params.Offset = n
		}
	}

	return params
}

// HasPaginationParams returns true if the request contains limit or offset query parameters.
func HasPaginationParams(r *http.Request) bool {
	return r.URL.Query().Has("limit") || r.URL.Query().Has("offset")
}
