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

// CreateAdminVersionRequest is the JSON body for POST /api/v1/admin/content/versions.
type CreateAdminVersionRequest struct {
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
	Label              string               `json:"label"`
	Locale             string               `json:"locale"`
}

///////////////////////////////
// HANDLERS
///////////////////////////////

// AdminListVersionsHandler handles GET requests to list all versions for an admin content data item.
// Reads admin_content_data_id from the "q" query parameter.
func AdminListVersionsHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	acdID := types.AdminContentID(q)
	if err := acdID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin_content_data_id: %v", err), http.StatusBadRequest)
		return
	}

	versions, err := svc.AdminContent.ListVersions(r.Context(), acdID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, versions)
}

// AdminGetVersionHandler handles GET requests to retrieve a specific admin content version.
// Reads admin_content_version_id from the "q" query parameter.
func AdminGetVersionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	avID := types.AdminContentVersionID(q)
	if err := avID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin_content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	version, err := svc.AdminContent.GetVersion(r.Context(), avID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, version)
}

// AdminCreateManualVersionHandler handles POST requests to create a manual admin snapshot version.
// Reads admin_content_data_id and optional label from the JSON body.
func AdminCreateManualVersionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req CreateAdminVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.AdminContentDataID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin_content_data_id: %v", err), http.StatusBadRequest)
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

	version, err := svc.AdminContent.CreateVersion(r.Context(), ac, req.AdminContentDataID, req.Locale, req.Label, user.UserID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(version)
}

// AdminDeleteVersionHandler handles DELETE requests to remove an admin content version.
// Reads admin_content_version_id from the "q" query parameter.
// Rejects deletion of published versions with 409 Conflict.
func AdminDeleteVersionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	avID := types.AdminContentVersionID(q)
	if err := avID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin_content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := svc.AdminContent.DeleteVersion(r.Context(), ac, avID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
