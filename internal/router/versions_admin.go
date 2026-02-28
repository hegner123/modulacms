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

// CreateAdminVersionRequest is the JSON body for POST /api/v1/admin/content/versions.
type CreateAdminVersionRequest struct {
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
	Label              string               `json:"label"`
}

///////////////////////////////
// HANDLERS
///////////////////////////////

// AdminListVersionsHandler handles GET requests to list all versions for an admin content data item.
// Reads admin_content_data_id from the "q" query parameter.
func AdminListVersionsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	acdID := types.AdminContentID(q)
	if err := acdID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid admin_content_data_id", err)
		http.Error(w, fmt.Sprintf("invalid admin_content_data_id: %v", err), http.StatusBadRequest)
		return
	}

	versions, err := d.ListAdminContentVersionsByContent(acdID)
	if err != nil {
		utility.DefaultLogger.Error("list admin content versions failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(versions)
}

// AdminGetVersionHandler handles GET requests to retrieve a specific admin content version.
// Reads admin_content_version_id from the "q" query parameter.
func AdminGetVersionHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	avID := types.AdminContentVersionID(q)
	if err := avID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid admin_content_version_id", err)
		http.Error(w, fmt.Sprintf("invalid admin_content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	version, err := d.GetAdminContentVersion(avID)
	if err != nil {
		utility.DefaultLogger.Error("get admin content version failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(version)
}

// AdminCreateManualVersionHandler handles POST requests to create a manual admin snapshot version.
// Reads admin_content_data_id and optional label from the JSON body.
func AdminCreateManualVersionHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
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

	d := db.ConfigDB(c)
	ctx := r.Context()

	// Build snapshot from live admin tables.
	snapshot, err := publishing.BuildAdminSnapshot(d, ctx, req.AdminContentDataID)
	if err != nil {
		utility.DefaultLogger.Error("build admin snapshot for manual version failed", err)
		http.Error(w, fmt.Sprintf("failed to build snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		utility.DefaultLogger.Error("marshal admin snapshot failed", err)
		http.Error(w, fmt.Sprintf("failed to marshal snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	// Get next version number.
	maxVersion, err := d.GetAdminMaxVersionNumber(req.AdminContentDataID, "")
	if err != nil {
		utility.DefaultLogger.Error("get admin max version number failed", err)
		http.Error(w, fmt.Sprintf("failed to get version number: %v", err), http.StatusInternalServerError)
		return
	}
	nextVersion := maxVersion + 1

	ac := middleware.AuditContextFromRequest(r, c)
	now := types.TimestampNow()

	version, err := d.CreateAdminContentVersion(ctx, ac, db.CreateAdminContentVersionParams{
		AdminContentDataID: req.AdminContentDataID,
		VersionNumber:      nextVersion,
		Locale:             "",
		Snapshot:           string(snapshotBytes),
		Trigger:            "manual",
		Label:              req.Label,
		Published:          false,
		PublishedBy:        types.NullableUserID{ID: user.UserID, Valid: true},
		DateCreated:        now,
	})
	if err != nil {
		utility.DefaultLogger.Error("create manual admin content version failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Async: prune old admin versions if retention cap exceeded.
	retentionCap := c.VersionMaxPerContent()
	go publishing.PruneExcessAdminVersions(d, req.AdminContentDataID, "", retentionCap)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(version)
}

// AdminDeleteVersionHandler handles DELETE requests to remove an admin content version.
// Reads admin_content_version_id from the "q" query parameter.
// Rejects deletion of published versions with 409 Conflict.
func AdminDeleteVersionHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	avID := types.AdminContentVersionID(q)
	if err := avID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid admin_content_version_id", err)
		http.Error(w, fmt.Sprintf("invalid admin_content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	// Fetch version to check published status before deleting.
	version, err := d.GetAdminContentVersion(avID)
	if err != nil {
		utility.DefaultLogger.Error("get admin content version for delete failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if version.Published {
		http.Error(w, "cannot delete a published version", http.StatusConflict)
		return
	}

	ctx := r.Context()
	ac := middleware.AuditContextFromRequest(r, c)

	if err := d.DeleteAdminContentVersion(ctx, ac, avID); err != nil {
		utility.DefaultLogger.Error("delete admin content version failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
