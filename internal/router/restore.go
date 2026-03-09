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
// REQUEST / RESPONSE TYPES
///////////////////////////////

// RestoreRequest is the JSON body for POST /api/v1/content/restore.
type RestoreRequest struct {
	ContentDataID    types.ContentID        `json:"content_data_id"`
	ContentVersionID types.ContentVersionID `json:"content_version_id"`
}

// AdminRestoreRequest is the JSON body for POST /api/v1/admin/content/restore.
type AdminRestoreRequest struct {
	AdminContentDataID    types.AdminContentID        `json:"admin_content_data_id"`
	AdminContentVersionID types.AdminContentVersionID `json:"admin_content_version_id"`
}

// RestoreResponse is the JSON response for content restore operations.
type RestoreResponse struct {
	Status          string   `json:"status"`
	ContentDataID   string   `json:"content_data_id"`
	RestoredVersion string   `json:"restored_version_id"`
	FieldsRestored  int      `json:"fields_restored"`
	UnmappedFields  []string `json:"unmapped_fields,omitempty"`
}

// AdminRestoreResponse is the JSON response for admin content restore operations.
type AdminRestoreResponse struct {
	Status             string   `json:"status"`
	AdminContentDataID string   `json:"admin_content_data_id"`
	RestoredVersion    string   `json:"restored_version_id"`
	FieldsRestored     int      `json:"fields_restored"`
	UnmappedFields     []string `json:"unmapped_fields,omitempty"`
}

///////////////////////////////
// HANDLERS
///////////////////////////////

// RestoreVersionHandler handles POST requests to restore content from a saved version snapshot.
func RestoreVersionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.ContentDataID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_data_id: %v", err), http.StatusBadRequest)
		return
	}
	if err := req.ContentVersionID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_version_id: %v", err), http.StatusBadRequest)
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

	result, err := svc.Content.RestoreVersion(r.Context(), ac, req.ContentDataID, req.ContentVersionID, user.UserID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RestoreResponse{
		Status:          "restored",
		ContentDataID:   req.ContentDataID.String(),
		RestoredVersion: req.ContentVersionID.String(),
		FieldsRestored:  result.FieldsRestored,
		UnmappedFields:  result.UnmappedFields,
	})
}

// AdminRestoreVersionHandler handles POST requests to restore admin content from a saved version snapshot.
func AdminRestoreVersionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req AdminRestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.AdminContentDataID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin_content_data_id: %v", err), http.StatusBadRequest)
		return
	}
	if err := req.AdminContentVersionID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin_content_version_id: %v", err), http.StatusBadRequest)
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

	result, err := svc.AdminContent.RestoreVersion(r.Context(), ac, req.AdminContentDataID, req.AdminContentVersionID, user.UserID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AdminRestoreResponse{
		Status:             "restored",
		AdminContentDataID: req.AdminContentDataID.String(),
		RestoredVersion:    req.AdminContentVersionID.String(),
		FieldsRestored:     result.FieldsRestored,
		UnmappedFields:     result.UnmappedFields,
	})
}
