package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/utility"
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
func AdminPublishHandler(w http.ResponseWriter, r *http.Request, c config.Config, dispatcher publishing.WebhookDispatcher) {
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

	d := db.ConfigDB(c)
	ac := middleware.AuditContextFromRequest(r, c)

	err := publishing.PublishAdminContent(r.Context(), d, req.AdminContentDataID, req.Locale, user.UserID, ac, c.VersionMaxPerContent(), dispatcher)
	if err != nil {
		utility.DefaultLogger.Error("admin publish content failed", err)
		if publishing.IsRevisionConflict(err) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
func AdminUnpublishHandler(w http.ResponseWriter, r *http.Request, c config.Config, dispatcher publishing.WebhookDispatcher) {
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

	d := db.ConfigDB(c)
	ac := middleware.AuditContextFromRequest(r, c)

	if err := publishing.UnpublishAdminContent(r.Context(), d, req.AdminContentDataID, req.Locale, user.UserID, ac, dispatcher); err != nil {
		utility.DefaultLogger.Error("admin unpublish content failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
func AdminScheduleHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
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
	if publishAt.Before(time.Now()) {
		http.Error(w, "publish_at must be in the future", http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	d := db.ConfigDB(c)
	now := types.TimestampNow()

	pubErr := d.UpdateAdminContentDataSchedule(r.Context(), db.UpdateAdminContentDataScheduleParams{
		PublishAt:          types.NewTimestamp(publishAt),
		DateModified:       now,
		AdminContentDataID: req.AdminContentDataID,
	})
	if pubErr != nil {
		utility.DefaultLogger.Error("admin schedule content failed", pubErr)
		http.Error(w, pubErr.Error(), http.StatusInternalServerError)
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
