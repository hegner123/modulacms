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
// REQUEST / RESPONSE TYPES
///////////////////////////////

// PublishRequest is the JSON body for POST /api/v1/content/publish.
type PublishRequest struct {
	ContentDataID types.ContentID `json:"content_data_id"`
}

// PublishResponse is the JSON response for publish and unpublish operations.
type PublishResponse struct {
	Status           string `json:"status"`
	VersionNumber    int64  `json:"version_number,omitempty"`
	ContentVersionID string `json:"content_version_id,omitempty"`
	ContentDataID    string `json:"content_data_id"`
}

// ScheduleRequest is the JSON body for POST /api/v1/content/schedule.
type ScheduleRequest struct {
	ContentDataID types.ContentID `json:"content_data_id"`
	PublishAt     string          `json:"publish_at"`
}

// ScheduleResponse is the JSON response for schedule operations.
type ScheduleResponse struct {
	Status        string `json:"status"`
	ContentDataID string `json:"content_data_id"`
	PublishAt     string `json:"publish_at"`
}

///////////////////////////////
// HTTP HANDLERS
///////////////////////////////

// PublishHandler handles POST requests to publish content.
// It builds a snapshot of the content tree, stores it as a versioned snapshot,
// and marks the content as published.
func PublishHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req PublishRequest
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

	d := db.ConfigDB(c)
	ac := middleware.AuditContextFromRequest(r, c)

	version, err := publishing.PublishContent(r.Context(), d, req.ContentDataID, user.UserID, ac, c.VersionMaxPerContent())
	if err != nil {
		utility.DefaultLogger.Error("publish content failed", err)
		// Return 409 Conflict for TOCTOU revision mismatch.
		if publishing.IsRevisionConflict(err) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(PublishResponse{
		Status:           "published",
		VersionNumber:    version.VersionNumber,
		ContentVersionID: version.ContentVersionID.String(),
		ContentDataID:    req.ContentDataID.String(),
	})
}

// UnpublishHandler handles POST requests to unpublish content.
// It clears the published flag and resets publish metadata to draft.
func UnpublishHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req PublishRequest
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

	d := db.ConfigDB(c)
	ac := middleware.AuditContextFromRequest(r, c)

	if err := publishing.UnpublishContent(r.Context(), d, req.ContentDataID, user.UserID, ac); err != nil {
		utility.DefaultLogger.Error("unpublish content failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(PublishResponse{
		Status:        "draft",
		ContentDataID: req.ContentDataID.String(),
	})
}

// ScheduleHandler handles POST requests to schedule content for future publication.
// It sets the publish_at field on the content data for the scheduler to pick up.
func ScheduleHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.ContentDataID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_data_id: %v", err), http.StatusBadRequest)
		return
	}

	// Validate publish_at is a valid RFC3339 timestamp in the future.
	publishAt, err := time.Parse(time.RFC3339, req.PublishAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid publish_at: must be RFC3339 format (e.g. 2026-03-01T00:00:00Z): %v", err), http.StatusBadRequest)
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

	// Set publish_at on the content data. The scheduler will pick this up
	// and call PublishContent when the time arrives.
	pubErr := d.UpdateContentDataSchedule(r.Context(), db.UpdateContentDataScheduleParams{
		PublishAt:     types.NewTimestamp(publishAt),
		DateModified:  now,
		ContentDataID: req.ContentDataID,
	})
	if pubErr != nil {
		utility.DefaultLogger.Error("schedule content failed", pubErr)
		http.Error(w, pubErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(ScheduleResponse{
		Status:        "scheduled",
		ContentDataID: req.ContentDataID.String(),
		PublishAt:     req.PublishAt,
	})
}
