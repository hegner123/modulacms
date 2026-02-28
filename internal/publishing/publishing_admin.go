package publishing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// BuildAdminSnapshot reads the admin content tree from live tables and assembles
// an AdminSnapshot suitable for JSON serialization and storage in admin_content_versions.
// It fetches all content data for the route and filters to the root and its descendants.
func BuildAdminSnapshot(d db.DbDriver, ctx context.Context, rootID types.AdminContentID) (*AdminSnapshot, error) {
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

// PublishAdminContent builds a snapshot and publishes admin content.
// retentionCap: max versions to keep (0 = unlimited). Async pruning runs after return.
func PublishAdminContent(ctx context.Context, d db.DbDriver, rootID types.AdminContentID, userID types.UserID, ac audited.AuditContext, retentionCap int) error {
	// Read root's current revision for TOCTOU guard.
	root, err := d.GetAdminContentData(rootID)
	if err != nil {
		return fmt.Errorf("fetch admin root %s: %w", rootID, err)
	}
	revisionBefore := root.Revision

	// Build snapshot.
	snapshot, err := BuildAdminSnapshot(d, ctx, rootID)
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
	go PruneExcessAdminVersions(d, rootID, "", retentionCap)

	return nil
}

// UnpublishAdminContent clears the published flag and resets publish metadata.
func UnpublishAdminContent(ctx context.Context, d db.DbDriver, rootID types.AdminContentID, userID types.UserID, ac audited.AuditContext) error {
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
