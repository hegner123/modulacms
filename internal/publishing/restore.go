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

// RestoreContent deserializes a version snapshot and updates matching content_field
// rows by (field_id, content_data_id). Resets the content status to draft.
func RestoreContent(ctx context.Context, d db.DbDriver, contentDataID types.ContentID, versionID types.ContentVersionID, userID types.UserID, ac audited.AuditContext) (*RestoreResult, error) {
	// 1. Fetch the version.
	version, err := d.GetContentVersion(versionID)
	if err != nil {
		return nil, fmt.Errorf("get content version: %w", err)
	}

	// 2. Verify the version belongs to the requested content_data_id.
	if version.ContentDataID != contentDataID {
		return nil, fmt.Errorf("content_version_id does not belong to the specified content_data_id")
	}

	// 3. Deserialize the snapshot.
	var snapshot Snapshot
	if err := json.Unmarshal([]byte(version.Snapshot), &snapshot); err != nil {
		return nil, fmt.Errorf("deserialize snapshot: %w", err)
	}

	// 4. Fetch current live content fields for this content_data_id.
	nullableID := types.NullableContentID{ID: contentDataID, Valid: true}
	currentFieldsPtr, err := d.ListContentFieldsByContentData(nullableID)
	if err != nil {
		return nil, fmt.Errorf("list current content fields: %w", err)
	}

	// Build a lookup map: (field_id, content_data_id) -> current ContentFields.
	type fieldKey struct {
		fieldID       types.FieldID
		contentDataID types.ContentID
	}
	currentFieldMap := make(map[fieldKey]db.ContentFields)
	if currentFieldsPtr != nil {
		for _, cf := range *currentFieldsPtr {
			if !cf.FieldID.Valid || !cf.ContentDataID.Valid {
				continue
			}
			k := fieldKey{fieldID: cf.FieldID.ID, contentDataID: cf.ContentDataID.ID}
			currentFieldMap[k] = cf
		}
	}

	// 5. Iterate snapshot fields and restore values.
	fieldsRestored := 0
	var unmappedFields []string

	now := types.TimestampNow()

	for _, snapshotField := range snapshot.ContentFields {
		fID := types.FieldID(snapshotField.FieldID)
		cdID := types.ContentID(snapshotField.ContentDataID)

		k := fieldKey{fieldID: fID, contentDataID: cdID}
		currentField, found := currentFieldMap[k]
		if !found {
			// Schema drift: field was deleted or content_data_id changed.
			unmappedFields = append(unmappedFields, snapshotField.FieldID)
			continue
		}

		// Update the field value to the snapshot value.
		_, updateErr := d.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
			RouteID:        currentField.RouteID,
			ContentDataID:  currentField.ContentDataID,
			FieldID:        currentField.FieldID,
			FieldValue:     snapshotField.FieldValue,
			AuthorID:       userID,
			DateCreated:    currentField.DateCreated,
			DateModified:   now,
			ContentFieldID: currentField.ContentFieldID,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("restore: update content field %s failed", currentField.ContentFieldID), updateErr)
			return nil, fmt.Errorf("update content field %s: %w", currentField.ContentFieldID, updateErr)
		}
		fieldsRestored++
	}

	// 6. Reset content status to draft.
	publishErr := d.UpdateContentDataPublishMeta(ctx, db.UpdateContentDataPublishMetaParams{
		Status:        types.ContentStatusDraft,
		PublishedAt:   types.Timestamp{},
		PublishedBy:   types.NullableUserID{},
		DateModified:  now,
		ContentDataID: contentDataID,
	})
	if publishErr != nil {
		return nil, fmt.Errorf("reset content status: %w", publishErr)
	}

	return &RestoreResult{
		FieldsRestored: fieldsRestored,
		UnmappedFields: unmappedFields,
	}, nil
}

// RestoreAdminContent deserializes an admin version snapshot and updates matching
// admin_content_field rows by (admin_field_id, admin_content_data_id). Resets the
// content status to draft.
func RestoreAdminContent(ctx context.Context, d db.DbDriver, adminContentDataID types.AdminContentID, versionID types.AdminContentVersionID, userID types.UserID, ac audited.AuditContext) (*RestoreResult, error) {
	// 1. Fetch the admin version.
	version, err := d.GetAdminContentVersion(versionID)
	if err != nil {
		return nil, fmt.Errorf("get admin content version: %w", err)
	}

	// 2. Verify the version belongs to the requested admin_content_data_id.
	if version.AdminContentDataID != adminContentDataID {
		return nil, fmt.Errorf("admin_content_version_id does not belong to the specified admin_content_data_id")
	}

	// 3. Deserialize the admin snapshot.
	var snapshot AdminSnapshot
	if err := json.Unmarshal([]byte(version.Snapshot), &snapshot); err != nil {
		return nil, fmt.Errorf("deserialize admin snapshot: %w", err)
	}

	// 4. Fetch current live admin content fields.
	adminCD, err := d.GetAdminContentData(adminContentDataID)
	if err != nil {
		return nil, fmt.Errorf("get admin content data: %w", err)
	}

	if !adminCD.AdminRouteID.Valid {
		return nil, fmt.Errorf("admin content data has no associated route")
	}

	allFieldsPtr, err := d.ListAdminContentFieldsByRoute(adminCD.AdminRouteID)
	if err != nil {
		return nil, fmt.Errorf("list admin content fields: %w", err)
	}

	// Build a lookup map: (admin_field_id, admin_content_data_id) -> current AdminContentFields,
	// filtered to only fields belonging to the target admin_content_data_id.
	type adminFieldKey struct {
		adminFieldID       types.AdminFieldID
		adminContentDataID types.AdminContentID
	}
	currentFieldMap := make(map[adminFieldKey]db.AdminContentFields)
	if allFieldsPtr != nil {
		for _, cf := range *allFieldsPtr {
			if !cf.AdminFieldID.Valid || !cf.AdminContentDataID.Valid {
				continue
			}
			if cf.AdminContentDataID.ID != adminContentDataID {
				continue
			}
			k := adminFieldKey{adminFieldID: cf.AdminFieldID.ID, adminContentDataID: cf.AdminContentDataID.ID}
			currentFieldMap[k] = cf
		}
	}

	// 5. Iterate snapshot fields and restore values.
	fieldsRestored := 0
	var unmappedFields []string

	now := types.TimestampNow()

	for _, snapshotField := range snapshot.ContentFields {
		fID := types.AdminFieldID(snapshotField.AdminFieldID)
		cdID := types.AdminContentID(snapshotField.AdminContentDataID)

		k := adminFieldKey{adminFieldID: fID, adminContentDataID: cdID}
		currentField, found := currentFieldMap[k]
		if !found {
			// Schema drift: field was deleted or content_data_id changed.
			unmappedFields = append(unmappedFields, snapshotField.AdminFieldID)
			continue
		}

		// Update the field value to the snapshot value.
		_, updateErr := d.UpdateAdminContentField(ctx, ac, db.UpdateAdminContentFieldParams{
			AdminRouteID:        currentField.AdminRouteID,
			AdminContentDataID:  currentField.AdminContentDataID,
			AdminFieldID:        currentField.AdminFieldID,
			AdminFieldValue:     snapshotField.AdminFieldValue,
			AuthorID:            userID,
			DateCreated:         currentField.DateCreated,
			DateModified:        now,
			AdminContentFieldID: currentField.AdminContentFieldID,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("admin restore: update admin content field %s failed", currentField.AdminContentFieldID), updateErr)
			return nil, fmt.Errorf("update admin content field %s: %w", currentField.AdminContentFieldID, updateErr)
		}
		fieldsRestored++
	}

	// 6. Reset admin content status to draft.
	publishErr := d.UpdateAdminContentDataPublishMeta(ctx, db.UpdateAdminContentDataPublishMetaParams{
		Status:             types.ContentStatusDraft,
		PublishedAt:        types.Timestamp{},
		PublishedBy:        types.NullableUserID{},
		DateModified:       now,
		AdminContentDataID: adminContentDataID,
	})
	if publishErr != nil {
		return nil, fmt.Errorf("reset admin content status: %w", publishErr)
	}

	return &RestoreResult{
		FieldsRestored: fieldsRestored,
		UnmappedFields: unmappedFields,
	}, nil
}
