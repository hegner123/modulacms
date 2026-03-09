package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// BatchContentUpdateRequest is the JSON body for POST /api/v1/content/batch.
// At least one of ContentData or Fields must be present.
type BatchContentUpdateRequest struct {
	ContentDataID types.ContentID             `json:"content_data_id"`
	ContentData   *db.UpdateContentDataParams `json:"content_data,omitempty"`
	Fields        map[types.FieldID]string    `json:"fields,omitempty"`
}

// ContentBatchHandler applies an optional content_data update plus a map of
// field value upserts in a single request. This mirrors the TUI pattern in
// HandleUpdateContentFromDialog (internal/tui/commands.go).
func ContentBatchHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req BatchContentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	var authorID types.UserID
	if user := middleware.AuthenticatedUser(ctx); user != nil {
		authorID = user.UserID
	}

	result, err := svc.Content.BatchUpdate(ctx, ac, service.BatchUpdateParams{
		ContentDataID: req.ContentDataID,
		ContentData:   req.ContentData,
		Fields:        req.Fields,
		AuthorID:      authorID,
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, result)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}
