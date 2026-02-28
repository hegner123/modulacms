package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// pruneExcessVersions removes the oldest unlabeled, unpublished versions that
// exceed the retention cap. It counts total versions first and only deletes the
// excess (total - cap). If cap is 0, pruning is disabled (unlimited retention).
func pruneExcessVersions(d db.DbDriver, contentDataID types.ContentID, locale string, cap int) {
	if cap <= 0 {
		return
	}
	total, err := d.CountContentVersionsByContent(contentDataID)
	if err != nil {
		utility.DefaultLogger.Error(fmt.Sprintf("prune: count versions for %s failed", contentDataID), err)
		return
	}
	if total == nil || *total <= int64(cap) {
		return
	}
	deleteCount := *total - int64(cap)
	pruneErr := d.PruneOldVersions(contentDataID, locale, deleteCount)
	if pruneErr != nil {
		utility.DefaultLogger.Error(fmt.Sprintf("prune: delete excess versions for %s failed", contentDataID), pruneErr)
	}
}

// pruneExcessAdminVersions removes the oldest unlabeled, unpublished admin
// versions that exceed the retention cap.
func pruneExcessAdminVersions(d db.DbDriver, adminContentDataID types.AdminContentID, locale string, cap int) {
	if cap <= 0 {
		return
	}
	total, err := d.CountAdminContentVersionsByContent(adminContentDataID)
	if err != nil {
		utility.DefaultLogger.Error(fmt.Sprintf("prune: count admin versions for %s failed", adminContentDataID), err)
		return
	}
	if total == nil || *total <= int64(cap) {
		return
	}
	deleteCount := *total - int64(cap)
	pruneErr := d.PruneAdminOldVersions(adminContentDataID, locale, deleteCount)
	if pruneErr != nil {
		utility.DefaultLogger.Error(fmt.Sprintf("prune: delete excess admin versions for %s failed", adminContentDataID), pruneErr)
	}
}

///////////////////////////////
// SNAPSHOT TYPES
///////////////////////////////

// Snapshot holds the serialized content tree data for a published version.
// It stores the raw parallel slices that can be fed back into model.BuildTree
// to reconstruct the content tree on delivery.
type Snapshot struct {
	ContentData   []db.ContentDataJSON       `json:"content_data"`
	Datatypes     []db.DatatypeJSON          `json:"datatypes"`
	ContentFields []SnapshotContentFieldJSON `json:"content_fields"`
	Fields        []db.FieldsJSON            `json:"fields"`
	Route         SnapshotRoute              `json:"route"`
	SchemaVersion int                        `json:"schema_version"`
}

// SnapshotRoute holds route metadata at the time of publish.
type SnapshotRoute struct {
	RouteID string `json:"route_id"`
	Slug    string `json:"slug"`
	Title   string `json:"title"`
}

// SnapshotContentFieldJSON is a string-based representation of a content field
// for snapshot serialization. The existing ContentFieldsJSON type is deprecated
// and uses int64 IDs which are incompatible with the ULID-based typed IDs.
type SnapshotContentFieldJSON struct {
	ContentFieldID string `json:"content_field_id"`
	RouteID        string `json:"route_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	FieldValue     string `json:"field_value"`
	AuthorID       string `json:"author_id"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
}

// MapSnapshotContentFieldJSON converts a ContentFields to its JSON representation.
func MapSnapshotContentFieldJSON(a db.ContentFields) SnapshotContentFieldJSON {
	return SnapshotContentFieldJSON{
		ContentFieldID: a.ContentFieldID.String(),
		RouteID:        a.RouteID.String(),
		ContentDataID:  a.ContentDataID.String(),
		FieldID:        a.FieldID.String(),
		FieldValue:     a.FieldValue,
		AuthorID:       a.AuthorID.String(),
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
	}
}

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
// SNAPSHOT BUILDER
///////////////////////////////

// buildSnapshot reads the content tree from live tables and assembles a Snapshot
// suitable for JSON serialization and storage in content_versions.
func buildSnapshot(d db.DbDriver, ctx context.Context, rootID types.ContentID) (*Snapshot, error) {
	// 1. Fetch the content data tree (root + all descendants).
	cdPtr, err := d.GetContentDataDescendants(ctx, rootID)
	if err != nil {
		return nil, fmt.Errorf("fetch descendants for %s: %w", rootID, err)
	}
	if cdPtr == nil || len(*cdPtr) == 0 {
		return nil, fmt.Errorf("no content data found for root %s", rootID)
	}
	cd := *cdPtr

	// 2. Fetch datatypes with cache (parallel slice: dt[i] is for cd[i]).
	dtCache := make(map[types.DatatypeID]db.Datatypes)
	dt := make([]db.Datatypes, len(cd))
	for i, c := range cd {
		if !c.DatatypeID.Valid {
			continue
		}
		dtID := c.DatatypeID.ID
		if cached, ok := dtCache[dtID]; ok {
			dt[i] = cached
			continue
		}
		datatype, dtErr := d.GetDatatype(dtID)
		if dtErr != nil {
			return nil, fmt.Errorf("fetch datatype %s for content %s: %w", dtID, c.ContentDataID, dtErr)
		}
		dtCache[dtID] = *datatype
		dt[i] = *datatype
	}

	// 3. Fetch content fields and field definitions for all content data rows.
	var allCF []db.ContentFields
	var allFD []db.Fields
	fieldCache := make(map[types.FieldID]db.Fields)

	for _, c := range cd {
		nullableID := types.NullableContentID{ID: c.ContentDataID, Valid: true}
		cfPtr, cfErr := d.ListContentFieldsByContentData(nullableID)
		if cfErr != nil {
			return nil, fmt.Errorf("fetch content fields for %s: %w", c.ContentDataID, cfErr)
		}
		if cfPtr == nil {
			continue
		}
		for _, cf := range *cfPtr {
			if !cf.FieldID.Valid {
				continue
			}
			fID := cf.FieldID.ID
			var fieldDef db.Fields
			if cached, ok := fieldCache[fID]; ok {
				fieldDef = cached
			} else {
				fd, fdErr := d.GetField(fID)
				if fdErr != nil {
					return nil, fmt.Errorf("fetch field %s: %w", fID, fdErr)
				}
				fieldCache[fID] = *fd
				fieldDef = *fd
			}
			allCF = append(allCF, cf)
			allFD = append(allFD, fieldDef)
		}
	}

	// 4. Resolve the route from the root content data.
	var route SnapshotRoute
	if cd[0].RouteID.Valid {
		r, rErr := d.GetRoute(cd[0].RouteID.ID)
		if rErr != nil {
			return nil, fmt.Errorf("fetch route %s: %w", cd[0].RouteID.ID, rErr)
		}
		route = SnapshotRoute{
			RouteID: r.RouteID.String(),
			Slug:    string(r.Slug),
			Title:   r.Title,
		}
	}

	// 5. Convert to JSON types for portable serialization.
	cdJSON := make([]db.ContentDataJSON, len(cd))
	for i, c := range cd {
		cdJSON[i] = db.MapContentDataJSON(c)
	}

	dtJSON := make([]db.DatatypeJSON, len(dt))
	for i, d := range dt {
		dtJSON[i] = db.MapDatatypeJSON(d)
	}

	cfJSON := make([]SnapshotContentFieldJSON, len(allCF))
	for i, cf := range allCF {
		cfJSON[i] = MapSnapshotContentFieldJSON(cf)
	}

	fdJSON := make([]db.FieldsJSON, len(allFD))
	for i, f := range allFD {
		fdJSON[i] = db.MapFieldJSON(f)
	}

	return &Snapshot{
		ContentData:   cdJSON,
		Datatypes:     dtJSON,
		ContentFields: cfJSON,
		Fields:        fdJSON,
		Route:         route,
		SchemaVersion: 1,
	}, nil
}

///////////////////////////////
// PUBLISH / UNPUBLISH LOGIC
///////////////////////////////

// publishContent builds a snapshot of the content tree, stores it as a new
// content version marked as published, and updates the content data's publish
// metadata. It uses optimistic locking via the revision field to prevent
// publishing stale data.
func publishContent(ctx context.Context, d db.DbDriver, rootID types.ContentID, userID types.UserID, ac audited.AuditContext, cfg config.Config) (*db.ContentVersion, error) {
	// 1. Read root's current revision for TOCTOU guard.
	root, err := d.GetContentData(rootID)
	if err != nil {
		return nil, fmt.Errorf("fetch root content data %s: %w", rootID, err)
	}
	revisionBefore := root.Revision

	// 2. Build snapshot from live tables.
	snapshot, err := buildSnapshot(d, ctx, rootID)
	if err != nil {
		return nil, fmt.Errorf("build snapshot: %w", err)
	}

	// 3. TOCTOU guard: verify revision hasn't changed during snapshot build.
	rootAfter, err := d.GetContentData(rootID)
	if err != nil {
		return nil, fmt.Errorf("re-fetch root for revision check: %w", err)
	}
	if rootAfter.Revision != revisionBefore {
		return nil, fmt.Errorf("conflict: content was modified during snapshot build (revision %d -> %d)", revisionBefore, rootAfter.Revision)
	}

	// 4. Marshal snapshot to JSON.
	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	// 5. Get next version number.
	maxVersion, err := d.GetMaxVersionNumber(rootID, "")
	if err != nil {
		return nil, fmt.Errorf("get max version number: %w", err)
	}
	nextVersion := maxVersion + 1

	// 6. Clear published flag on all existing versions for this content+locale.
	if clearErr := d.ClearPublishedFlag(rootID, ""); clearErr != nil {
		return nil, fmt.Errorf("clear published flag: %w", clearErr)
	}

	// 7. Create new content version with published=true.
	now := types.TimestampNow()
	version, err := d.CreateContentVersion(ctx, ac, db.CreateContentVersionParams{
		ContentDataID: rootID,
		VersionNumber: nextVersion,
		Locale:        "",
		Snapshot:      string(snapshotBytes),
		Trigger:       "publish",
		Label:         "",
		Published:     true,
		PublishedBy:   types.NullableUserID{ID: userID, Valid: true},
		DateCreated:   now,
	})
	if err != nil {
		return nil, fmt.Errorf("create content version: %w", err)
	}

	// 8. Update content data publish metadata.
	publishErr := d.UpdateContentDataPublishMeta(ctx, db.UpdateContentDataPublishMetaParams{
		Status:        types.ContentStatusPublished,
		PublishedAt:   now,
		PublishedBy:   types.NullableUserID{ID: userID, Valid: true},
		DateModified:  now,
		ContentDataID: rootID,
	})
	if publishErr != nil {
		return nil, fmt.Errorf("update publish metadata: %w", publishErr)
	}

	// 9. Async: prune old versions if retention cap exceeded.
	retentionCap := cfg.VersionMaxPerContent()
	go pruneExcessVersions(d, rootID, "", retentionCap)

	return version, nil
}

// unpublishContent clears the published flag on all versions for the given
// content data and resets the publish metadata to draft status.
func unpublishContent(ctx context.Context, d db.DbDriver, rootID types.ContentID, userID types.UserID, ac audited.AuditContext) error {
	// 1. Clear published flag on all versions.
	if err := d.ClearPublishedFlag(rootID, ""); err != nil {
		return fmt.Errorf("clear published flag: %w", err)
	}

	// 2. Reset content data publish metadata to draft.
	now := types.TimestampNow()
	return d.UpdateContentDataPublishMeta(ctx, db.UpdateContentDataPublishMetaParams{
		Status:        types.ContentStatusDraft,
		PublishedAt:   types.Timestamp{},
		PublishedBy:   types.NullableUserID{},
		DateModified:  now,
		ContentDataID: rootID,
	})
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

	version, err := publishContent(r.Context(), d, req.ContentDataID, user.UserID, ac, c)
	if err != nil {
		utility.DefaultLogger.Error("publish content failed", err)
		// Return 409 Conflict for TOCTOU revision mismatch.
		if isRevisionConflict(err) {
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

	if err := unpublishContent(r.Context(), d, req.ContentDataID, user.UserID, ac); err != nil {
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
	// and call publishContent when the time arrives.
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

///////////////////////////////
// HELPERS
///////////////////////////////

// isRevisionConflict checks whether an error is a TOCTOU revision conflict.
// Uses string containment because the error is constructed locally, not from
// an external source.
func isRevisionConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Check for the specific prefix from publishContent's TOCTOU guard.
	return len(msg) >= 9 && msg[:9] == "conflict:"
}
