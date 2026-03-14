package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/hegner123/modulacms/internal/search"
)

// SearchHandler handles GET /api/v1/search
func SearchHandler(w http.ResponseWriter, r *http.Request, searchSvc *search.Service) {
	if searchSvc == nil {
		http.Error(w, "search is not enabled", http.StatusNotFound)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "missing required parameter: q", http.StatusBadRequest)
		return
	}

	opts := search.SearchOptions{
		DatatypeName: r.URL.Query().Get("type"),
		Locale:       r.URL.Query().Get("locale"),
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			opts.Limit = v
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			opts.Offset = v
		}
	}

	// Use prefix matching by default, disable with prefix=false
	usePrefix := true
	if prefixStr := r.URL.Query().Get("prefix"); prefixStr == "false" {
		usePrefix = false
	}

	var resp search.SearchResponse
	if usePrefix {
		resp = searchSvc.SearchWithPrefix(q, opts)
	} else {
		resp = searchSvc.Search(q, opts)
	}

	w.Header().Set("Content-Type", "application/json")
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(resp)
}

// SearchRebuildHandler handles POST /api/v1/admin/search/rebuild
func SearchRebuildHandler(w http.ResponseWriter, r *http.Request, searchSvc *search.Service) {
	if searchSvc == nil {
		http.Error(w, "search is not enabled", http.StatusNotFound)
		return
	}

	if err := searchSvc.Rebuild(); err != nil {
		http.Error(w, "rebuild failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stats := searchSvc.Stats()
	w.Header().Set("Content-Type", "application/json")
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"documents": stats.Documents,
		"terms":     stats.Terms,
		"mem_bytes": stats.MemEstimate,
	})
}
