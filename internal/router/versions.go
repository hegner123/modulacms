package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

///////////////////////////////
// REQUEST TYPES
///////////////////////////////

// CreateVersionRequest is the JSON body for POST /api/v1/content/versions.
type CreateVersionRequest struct {
	ContentDataID types.ContentID `json:"content_data_id"`
	Label         string          `json:"label"`
	Locale        string          `json:"locale"`
}

///////////////////////////////
// HANDLERS
///////////////////////////////

// ListVersionsHandler handles GET requests to list all versions for a content data item.
// Reads content_data_id from the "q" query parameter.
func ListVersionsHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	cdID := types.ContentID(q)
	if err := cdID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_data_id: %v", err), http.StatusBadRequest)
		return
	}

	versions, err := svc.Content.ListVersions(r.Context(), cdID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, versions)
}

// GetVersionHandler handles GET requests to retrieve a specific content version.
// Reads content_version_id from the "q" query parameter.
func GetVersionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	vID := types.ContentVersionID(q)
	if err := vID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	version, err := svc.Content.GetVersion(r.Context(), vID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, version)
}

// CreateManualVersionHandler handles POST requests to create a manual snapshot version.
// Reads content_data_id and optional label from the JSON body.
func CreateManualVersionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req CreateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.ContentDataID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_data_id: %v", err), http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	version, err := svc.Content.CreateVersion(r.Context(), ac, req.ContentDataID, req.Locale, req.Label, user.UserID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(version)
}

// DeleteVersionHandler handles DELETE requests to remove a content version.
// Reads content_version_id from the "q" query parameter.
// Rejects deletion of published versions with 409 Conflict.
func DeleteVersionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	vID := types.ContentVersionID(q)
	if err := vID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := svc.Content.DeleteVersion(r.Context(), ac, vID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
