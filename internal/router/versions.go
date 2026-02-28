package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/utility"
)

///////////////////////////////
// REQUEST TYPES
///////////////////////////////

// CreateVersionRequest is the JSON body for POST /api/v1/content/versions.
type CreateVersionRequest struct {
	ContentDataID types.ContentID `json:"content_data_id"`
	Label         string          `json:"label"`
}

///////////////////////////////
// HANDLERS
///////////////////////////////

// ListVersionsHandler handles GET requests to list all versions for a content data item.
// Reads content_data_id from the "q" query parameter.
func ListVersionsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	cdID := types.ContentID(q)
	if err := cdID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid content_data_id", err)
		http.Error(w, fmt.Sprintf("invalid content_data_id: %v", err), http.StatusBadRequest)
		return
	}

	versions, err := d.ListContentVersionsByContent(cdID)
	if err != nil {
		utility.DefaultLogger.Error("list content versions failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(versions)
}

// GetVersionHandler handles GET requests to retrieve a specific content version.
// Reads content_version_id from the "q" query parameter.
func GetVersionHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	vID := types.ContentVersionID(q)
	if err := vID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid content_version_id", err)
		http.Error(w, fmt.Sprintf("invalid content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	version, err := d.GetContentVersion(vID)
	if err != nil {
		utility.DefaultLogger.Error("get content version failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(version)
}

// CreateManualVersionHandler handles POST requests to create a manual snapshot version.
// Reads content_data_id and optional label from the JSON body.
func CreateManualVersionHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
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

	d := db.ConfigDB(c)
	ctx := r.Context()

	// Build snapshot from live tables.
	snapshot, err := publishing.BuildSnapshot(d, ctx, req.ContentDataID)
	if err != nil {
		utility.DefaultLogger.Error("build snapshot for manual version failed", err)
		http.Error(w, fmt.Sprintf("failed to build snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		utility.DefaultLogger.Error("marshal snapshot failed", err)
		http.Error(w, fmt.Sprintf("failed to marshal snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	// Get next version number.
	maxVersion, err := d.GetMaxVersionNumber(req.ContentDataID, "")
	if err != nil {
		utility.DefaultLogger.Error("get max version number failed", err)
		http.Error(w, fmt.Sprintf("failed to get version number: %v", err), http.StatusInternalServerError)
		return
	}
	nextVersion := maxVersion + 1

	ac := middleware.AuditContextFromRequest(r, c)
	now := types.TimestampNow()

	version, err := d.CreateContentVersion(ctx, ac, db.CreateContentVersionParams{
		ContentDataID: req.ContentDataID,
		VersionNumber: nextVersion,
		Locale:        "",
		Snapshot:      string(snapshotBytes),
		Trigger:       "manual",
		Label:         req.Label,
		Published:     false,
		PublishedBy:   types.NullableUserID{ID: user.UserID, Valid: true},
		DateCreated:   now,
	})
	if err != nil {
		utility.DefaultLogger.Error("create manual content version failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Async: prune old versions if retention cap exceeded.
	retentionCap := c.VersionMaxPerContent()
	go publishing.PruneExcessVersions(d, req.ContentDataID, "", retentionCap)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(version)
}

// DeleteVersionHandler handles DELETE requests to remove a content version.
// Reads content_version_id from the "q" query parameter.
// Rejects deletion of published versions with 409 Conflict.
func DeleteVersionHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	vID := types.ContentVersionID(q)
	if err := vID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid content_version_id", err)
		http.Error(w, fmt.Sprintf("invalid content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	// Fetch version to check published status before deleting.
	version, err := d.GetContentVersion(vID)
	if err != nil {
		utility.DefaultLogger.Error("get content version for delete failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if version.Published {
		http.Error(w, "cannot delete a published version", http.StatusConflict)
		return
	}

	ctx := r.Context()
	ac := middleware.AuditContextFromRequest(r, c)

	if err := d.DeleteContentVersion(ctx, ac, vID); err != nil {
		utility.DefaultLogger.Error("delete content version failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
