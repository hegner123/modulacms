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
	AdminContentDataID types.AdminContentID `json:"content_data_id"`
	Label              string               `json:"label"`
	Locale             string               `json:"locale"`
}

///////////////////////////////
// HANDLERS
///////////////////////////////

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
