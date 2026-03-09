package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

///////////////////////////////
// ADMIN REQUEST / RESPONSE TYPES
///////////////////////////////

// AdminPublishRequest is the JSON body for admin publish/unpublish operations.
type AdminPublishRequest struct {
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
	Locale             string               `json:"locale"`
}

// AdminPublishResponse is the JSON response for admin publish and unpublish operations.
type AdminPublishResponse struct {
	Status                string `json:"status"`
	VersionNumber         int64  `json:"version_number,omitempty"`
	AdminContentVersionID string `json:"admin_content_version_id,omitempty"`
	AdminContentDataID    string `json:"admin_content_data_id"`
}

// AdminScheduleRequest is the JSON body for admin schedule operations.
type AdminScheduleRequest struct {
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
	PublishAt          string               `json:"publish_at"`
}

// AdminScheduleResponse is the JSON response for admin schedule operations.
type AdminScheduleResponse struct {
	Status             string `json:"status"`
	AdminContentDataID string `json:"admin_content_data_id"`
	PublishAt          string `json:"publish_at"`
}

///////////////////////////////
// ADMIN HTTP HANDLERS
///////////////////////////////

// AdminPublishHandler handles POST requests to publish admin content.
func AdminPublishHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req AdminPublishRequest
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

	err = svc.AdminContent.Publish(r.Context(), ac, req.AdminContentDataID, req.Locale, user.UserID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AdminPublishResponse{
		Status:             "published",
		AdminContentDataID: req.AdminContentDataID.String(),
	})
}

// AdminUnpublishHandler handles POST requests to unpublish admin content.
func AdminUnpublishHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req AdminPublishRequest
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

	if err := svc.AdminContent.Unpublish(r.Context(), ac, req.AdminContentDataID, req.Locale, user.UserID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AdminPublishResponse{
		Status:             "draft",
		AdminContentDataID: req.AdminContentDataID.String(),
	})
}

// AdminScheduleHandler handles POST requests to schedule admin content for publication.
func AdminScheduleHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req AdminScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.AdminContentDataID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin_content_data_id: %v", err), http.StatusBadRequest)
		return
	}

	publishAt, err := time.Parse(time.RFC3339, req.PublishAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid publish_at: must be RFC3339 format: %v", err), http.StatusBadRequest)
		return
	}

	if err := svc.AdminContent.Schedule(r.Context(), req.AdminContentDataID, publishAt); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AdminScheduleResponse{
		Status:             "scheduled",
		AdminContentDataID: req.AdminContentDataID.String(),
		PublishAt:          req.PublishAt,
	})
}
