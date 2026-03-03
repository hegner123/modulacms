package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// ContentVersionsListHandler lists content versions filtered by content_id query parameter.
// Registered as: GET /api/v1/contentversions
func ContentVersionsListHandler(w http.ResponseWriter, r *http.Request, d db.DbDriver) {
	contentIDStr := r.URL.Query().Get("content_id")
	if contentIDStr == "" {
		http.Error(w, "content_id query parameter is required", http.StatusBadRequest)
		return
	}

	contentID := types.ContentID(contentIDStr)
	if err := contentID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid content_id", err)
		http.Error(w, fmt.Sprintf("invalid content_id: %v", err), http.StatusBadRequest)
		return
	}

	versions, err := d.ListContentVersionsByContent(contentID)
	if err != nil {
		utility.DefaultLogger.Error("failed to list content versions by content", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(versions)
}
