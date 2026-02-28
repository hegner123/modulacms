package publishing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// PruneExcessVersions removes the oldest unlabeled, unpublished versions that
// exceed the retention cap. It counts total versions first and only deletes the
// excess (total - cap). If cap is 0, pruning is disabled (unlimited retention).
func PruneExcessVersions(d db.DbDriver, contentDataID types.ContentID, locale string, cap int) {
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

// PruneExcessAdminVersions removes the oldest unlabeled, unpublished admin
// versions that exceed the retention cap.
func PruneExcessAdminVersions(d db.DbDriver, adminContentDataID types.AdminContentID, locale string, cap int) {
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

// BuildSnapshot reads the content tree from live tables and assembles a Snapshot
// suitable for JSON serialization and storage in content_versions.
func BuildSnapshot(d db.DbDriver, ctx context.Context, rootID types.ContentID) (*Snapshot, error) {
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

// PublishContent builds a snapshot of the content tree, stores it as a new
// content version marked as published, and updates the content data's publish
// metadata. It uses optimistic locking via the revision field to prevent
// publishing stale data. retentionCap: max versions to keep (0 = unlimited).
// Async pruning runs after return.
func PublishContent(ctx context.Context, d db.DbDriver, rootID types.ContentID, userID types.UserID, ac audited.AuditContext, retentionCap int) (*db.ContentVersion, error) {
	// 1. Read root's current revision for TOCTOU guard.
	root, err := d.GetContentData(rootID)
	if err != nil {
		return nil, fmt.Errorf("fetch root content data %s: %w", rootID, err)
	}
	revisionBefore := root.Revision

	// 2. Build snapshot from live tables.
	snapshot, err := BuildSnapshot(d, ctx, rootID)
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
	go PruneExcessVersions(d, rootID, "", retentionCap)

	return version, nil
}

// UnpublishContent clears the published flag on all versions for the given
// content data and resets the publish metadata to draft status.
func UnpublishContent(ctx context.Context, d db.DbDriver, rootID types.ContentID, userID types.UserID, ac audited.AuditContext) error {
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

// IsRevisionConflict checks whether an error is a TOCTOU revision conflict.
// Uses string containment because the error is constructed locally, not from
// an external source.
func IsRevisionConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Check for the specific prefix from PublishContent's TOCTOU guard.
	return len(msg) >= 9 && msg[:9] == "conflict:"
}
