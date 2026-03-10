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
// REQUEST / RESPONSE TYPES
///////////////////////////////

// PublishRequest is the JSON body for POST /api/v1/content/publish.
type PublishRequest struct {
	AdminContentDataID types.AdminContentID `json:"content_data_id"`
	Locale             string               `json:"locale"`
}

// PublishResponse is the JSON response for publish and unpublish operations.
type PublishResponse struct {
	Status                string `json:"status"`
	VersionNumber         int64  `json:"version_number,omitempty"`
	AdminContentVersionID string `json:"content_version_id,omitempty"`
	AdminContentDataID    string `json:"content_data_id"`
}

// ScheduleRequest is the JSON body for POST /api/v1/content/schedule.
type ScheduleRequest struct {
	AdminContentDataID types.AdminContentID `json:"content_data_id"`
	PublishAt          string               `json:"publish_at"`
}

// ScheduleResponse is the JSON response for schedule operations.
type ScheduleResponse struct {
	Status             string `json:"status"`
	AdminContentDataID string `json:"content_data_id"`
	PublishAt          string `json:"publish_at"`
}

///////////////////////////////
// HTTP HANDLERS
///////////////////////////////

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
