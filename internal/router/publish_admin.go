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

///////////////////////////////
// ADMIN SNAPSHOT TYPES
///////////////////////////////

// AdminSnapshot holds the serialized admin content tree data for a published version.
type AdminSnapshot struct {
	ContentData   []db.ContentDataJSON           `json:"content_data"`
	Datatypes     []db.DatatypeJSON              `json:"datatypes"`
	ContentFields []AdminSnapshotContentFieldJSON `json:"content_fields"`
	Fields        []db.FieldsJSON                `json:"fields"`
	Route         AdminSnapshotRoute             `json:"route"`
	SchemaVersion int                            `json:"schema_version"`
}

// AdminSnapshotRoute holds admin route metadata at the time of publish.
type AdminSnapshotRoute struct {
	AdminRouteID string `json:"admin_route_id"`
}

// AdminSnapshotContentFieldJSON is a string-based representation of an admin content field.
type AdminSnapshotContentFieldJSON struct {
	AdminContentFieldID string `json:"admin_content_field_id"`
	AdminRouteID        string `json:"admin_route_id"`
	AdminContentDataID  string `json:"admin_content_data_id"`
	AdminFieldID        string `json:"admin_field_id"`
	AdminFieldValue     string `json:"admin_field_value"`
	AuthorID            string `json:"author_id"`
	DateCreated         string `json:"date_created"`
	DateModified        string `json:"date_modified"`
}

// MapAdminSnapshotContentFieldJSON converts an AdminContentFields to its JSON representation.
func MapAdminSnapshotContentFieldJSON(a db.AdminContentFields) AdminSnapshotContentFieldJSON {
	return AdminSnapshotContentFieldJSON{
		AdminContentFieldID: a.AdminContentFieldID.String(),
		AdminRouteID:        a.AdminRouteID.String(),
		AdminContentDataID:  a.AdminContentDataID.String(),
		AdminFieldID:        a.AdminFieldID.String(),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID.String(),
		DateCreated:         a.DateCreated.String(),
		DateModified:        a.DateModified.String(),
	}
}

///////////////////////////////
// ADMIN REQUEST / RESPONSE TYPES
///////////////////////////////

// AdminPublishRequest is the JSON body for admin publish/unpublish operations.
type AdminPublishRequest struct {
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
}

// AdminPublishResponse is the JSON response for admin publish and unpublish operations.
type AdminPublishResponse struct {
	Status                 string `json:"status"`
	VersionNumber          int64  `json:"version_number,omitempty"`
	AdminContentVersionID  string `json:"admin_content_version_id,omitempty"`
	AdminContentDataID     string `json:"admin_content_data_id"`
}

// AdminScheduleRequest is the JSON body for admin schedule operations.
type AdminScheduleRequest struct {
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
	PublishAt          string                `json:"publish_at"`
}

// AdminScheduleResponse is the JSON response for admin schedule operations.
type AdminScheduleResponse struct {
	Status             string `json:"status"`
	AdminContentDataID string `json:"admin_content_data_id"`
	PublishAt          string `json:"publish_at"`
}

///////////////////////////////
// ADMIN SNAPSHOT BUILDER
///////////////////////////////

// buildAdminSnapshot reads the admin content tree from live tables and assembles
// an AdminSnapshot suitable for JSON serialization and storage in admin_content_versions.
// It fetches all content data for the route and filters to the root and its descendants.
func buildAdminSnapshot(d db.DbDriver, ctx context.Context, rootID types.AdminContentID) (*AdminSnapshot, error) {
	// 1. Fetch root content data to determine its route.
	root, err := d.GetAdminContentData(rootID)
	if err != nil {
		return nil, fmt.Errorf("fetch admin content data %s: %w", rootID, err)
	}

	if !root.AdminRouteID.Valid {
		return nil, fmt.Errorf("admin content data %s has no route", rootID)
	}

	// 2. Fetch all content data for this route.
	allCD, err := d.ListAdminContentDataByRoute(root.AdminRouteID)
	if err != nil {
		return nil, fmt.Errorf("list admin content data by route: %w", err)
	}
	if allCD == nil || len(*allCD) == 0 {
		return nil, fmt.Errorf("no admin content data found for route %s", root.AdminRouteID.ID)
	}

	// 3. Filter to root and its descendants by walking the tree (BFS).
	cdIndex := make(map[types.AdminContentID]db.AdminContentData, len(*allCD))
	children := make(map[types.AdminContentID][]types.AdminContentID)
	for _, c := range *allCD {
		cdIndex[c.AdminContentDataID] = c
		if c.ParentID.Valid {
			children[c.ParentID.ID] = append(children[c.ParentID.ID], c.AdminContentDataID)
		}
	}

	var cd []db.AdminContentData
	queue := []types.AdminContentID{rootID}
	visited := make(map[types.AdminContentID]struct{})
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, seen := visited[current]; seen {
			continue
		}
		visited[current] = struct{}{}
		if node, ok := cdIndex[current]; ok {
			cd = append(cd, node)
			queue = append(queue, children[current]...)
		}
	}

	if len(cd) == 0 {
		return nil, fmt.Errorf("no admin content data found for root %s", rootID)
	}

	// 4. Fetch datatypes with cache.
	dtCache := make(map[types.AdminDatatypeID]db.AdminDatatypes)
	dt := make([]db.AdminDatatypes, len(cd))
	for i, c := range cd {
		if !c.AdminDatatypeID.Valid {
			continue
		}
		dtID := c.AdminDatatypeID.ID
		if cached, ok := dtCache[dtID]; ok {
			dt[i] = cached
			continue
		}
		datatype, dtErr := d.GetAdminDatatypeById(dtID)
		if dtErr != nil {
			return nil, fmt.Errorf("fetch admin datatype %s for content %s: %w", dtID, c.AdminContentDataID, dtErr)
		}
		dtCache[dtID] = *datatype
		dt[i] = *datatype
	}

	// 5. Fetch content fields for the entire route, then filter to descendants.
	allCFPtr, err := d.ListAdminContentFieldsByRoute(root.AdminRouteID)
	if err != nil {
		return nil, fmt.Errorf("list admin content fields by route: %w", err)
	}

	var filteredCF []db.AdminContentFields
	var filteredFD []db.AdminFields
	fieldCache := make(map[types.AdminFieldID]db.AdminFields)

	if allCFPtr != nil {
		for _, cf := range *allCFPtr {
			if !cf.AdminContentDataID.Valid {
				continue
			}
			if _, isDescendant := visited[cf.AdminContentDataID.ID]; !isDescendant {
				continue
			}
			if !cf.AdminFieldID.Valid {
				continue
			}

			fID := cf.AdminFieldID.ID
			var fieldDef db.AdminFields
			if cached, ok := fieldCache[fID]; ok {
				fieldDef = cached
			} else {
				fd, fdErr := d.GetAdminField(fID)
				if fdErr != nil {
					return nil, fmt.Errorf("fetch admin field %s: %w", fID, fdErr)
				}
				fieldCache[fID] = *fd
				fieldDef = *fd
			}
			filteredCF = append(filteredCF, cf)
			filteredFD = append(filteredFD, fieldDef)
		}
	}

	// 6. Build route metadata.
	route := AdminSnapshotRoute{
		AdminRouteID: root.AdminRouteID.ID.String(),
	}

	// 7. Convert to JSON types for portable serialization.
	cdJSON := make([]db.ContentDataJSON, len(cd))
	for i, c := range cd {
		cdJSON[i] = db.MapAdminContentDataJSON(c)
	}

	dtJSON := make([]db.DatatypeJSON, len(dt))
	for i, d := range dt {
		dtJSON[i] = db.MapAdminDatatypeJSON(d)
	}

	cfJSON := make([]AdminSnapshotContentFieldJSON, len(filteredCF))
	for i, cf := range filteredCF {
		cfJSON[i] = MapAdminSnapshotContentFieldJSON(cf)
	}

	fdJSON := make([]db.FieldsJSON, len(filteredFD))
	for i, f := range filteredFD {
		fdJSON[i] = db.MapAdminFieldJSON(f)
	}

	return &AdminSnapshot{
		ContentData:   cdJSON,
		Datatypes:     dtJSON,
		ContentFields: cfJSON,
		Fields:        fdJSON,
		Route:         route,
		SchemaVersion: 1,
	}, nil
}

///////////////////////////////
// ADMIN PUBLISH / UNPUBLISH LOGIC
///////////////////////////////

// publishAdminContent builds a snapshot and publishes admin content.
func publishAdminContent(ctx context.Context, d db.DbDriver, rootID types.AdminContentID, userID types.UserID, ac audited.AuditContext, cfg config.Config) error {
	// Read root's current revision for TOCTOU guard.
	root, err := d.GetAdminContentData(rootID)
	if err != nil {
		return fmt.Errorf("fetch admin root %s: %w", rootID, err)
	}
	revisionBefore := root.Revision

	// Build snapshot.
	snapshot, err := buildAdminSnapshot(d, ctx, rootID)
	if err != nil {
		return fmt.Errorf("build admin snapshot: %w", err)
	}

	// TOCTOU guard.
	rootAfter, err := d.GetAdminContentData(rootID)
	if err != nil {
		return fmt.Errorf("re-fetch admin root for revision check: %w", err)
	}
	if rootAfter.Revision != revisionBefore {
		return fmt.Errorf("conflict: admin content was modified during snapshot build (revision %d -> %d)", revisionBefore, rootAfter.Revision)
	}

	// Marshal snapshot.
	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal admin snapshot: %w", err)
	}

	// Get next version number.
	maxVersion, err := d.GetAdminMaxVersionNumber(rootID, "")
	if err != nil {
		return fmt.Errorf("get admin max version number: %w", err)
	}
	nextVersion := maxVersion + 1

	// Clear published flag.
	if clearErr := d.ClearAdminPublishedFlag(rootID, ""); clearErr != nil {
		return fmt.Errorf("clear admin published flag: %w", clearErr)
	}

	// Create new version.
	now := types.TimestampNow()
	_, err = d.CreateAdminContentVersion(ctx, ac, db.CreateAdminContentVersionParams{
		AdminContentDataID: rootID,
		VersionNumber:      nextVersion,
		Locale:             "",
		Snapshot:           string(snapshotBytes),
		Trigger:            "publish",
		Label:              "",
		Published:          true,
		PublishedBy:        types.NullableUserID{ID: userID, Valid: true},
		DateCreated:        now,
	})
	if err != nil {
		return fmt.Errorf("create admin content version: %w", err)
	}

	// Update publish metadata.
	publishErr := d.UpdateAdminContentDataPublishMeta(ctx, db.UpdateAdminContentDataPublishMetaParams{
		Status:             types.ContentStatusPublished,
		PublishedAt:        now,
		PublishedBy:        types.NullableUserID{ID: userID, Valid: true},
		DateModified:       now,
		AdminContentDataID: rootID,
	})
	if publishErr != nil {
		return fmt.Errorf("update admin publish metadata: %w", publishErr)
	}

	// Async prune.
	retentionCap := cfg.VersionMaxPerContent()
	go pruneExcessAdminVersions(d, rootID, "", retentionCap)

	return nil
}

// unpublishAdminContent clears the published flag and resets publish metadata.
func unpublishAdminContent(ctx context.Context, d db.DbDriver, rootID types.AdminContentID, userID types.UserID, ac audited.AuditContext) error {
	if err := d.ClearAdminPublishedFlag(rootID, ""); err != nil {
		return fmt.Errorf("clear admin published flag: %w", err)
	}

	now := types.TimestampNow()
	return d.UpdateAdminContentDataPublishMeta(ctx, db.UpdateAdminContentDataPublishMetaParams{
		Status:             types.ContentStatusDraft,
		PublishedAt:        types.Timestamp{},
		PublishedBy:        types.NullableUserID{},
		DateModified:       now,
		AdminContentDataID: rootID,
	})
}

///////////////////////////////
// ADMIN HTTP HANDLERS
///////////////////////////////

// AdminPublishHandler handles POST requests to publish admin content.
func AdminPublishHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
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

	err := publishAdminContent(r.Context(), d, req.AdminContentDataID, user.UserID, ac, c)
	if err != nil {
		utility.DefaultLogger.Error("admin publish content failed", err)
		if isRevisionConflict(err) {
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
func AdminUnpublishHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
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

	if err := unpublishAdminContent(r.Context(), d, req.AdminContentDataID, user.UserID, ac); err != nil {
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
