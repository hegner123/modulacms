package router

import (
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

// ContentVersionsListHandler lists content versions filtered by content_id query parameter.
// Registered as: GET /api/v1/contentversions
func ContentVersionsListHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	contentIDStr := r.URL.Query().Get("content_id")
	if contentIDStr == "" {
		http.Error(w, "content_id query parameter is required", http.StatusBadRequest)
		return
	}

	contentID := types.ContentID(contentIDStr)
	if err := contentID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_id: %v", err), http.StatusBadRequest)
		return
	}

	versions, err := svc.Content.ListVersions(r.Context(), contentID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, versions)
}
