package handlers

import (
	"net/http"
	"strconv"

	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/search"
	"github.com/hegner123/modulacms/internal/service"
)

// AdminSearchHandler handles GET /admin/search and returns a partial
// search results dropdown for the topbar global search.
func AdminSearchHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !svc.Search.Available() {
			w.WriteHeader(http.StatusOK)
			return
		}

		q := r.URL.Query().Get("q")
		if q == "" {
			w.WriteHeader(http.StatusOK)
			return
		}

		opts := search.SearchOptions{
			Limit: 8,
		}
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
				opts.Limit = v
			}
		}
		if dtName := r.URL.Query().Get("type"); dtName != "" {
			opts.DatatypeName = dtName
		}

		resp, err := svc.Search.Search(r.Context(), q, true, opts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		Render(w, r, partials.SearchResults(resp.Results, resp.Total, q))
	}
}
