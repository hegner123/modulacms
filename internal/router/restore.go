package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

///////////////////////////////
// REQUEST / RESPONSE TYPES
///////////////////////////////

// RestoreRequest is the JSON body for POST /api/v1/content/restore.
type RestoreRequest struct {
	ContentDataID    types.ContentID        `json:"content_data_id"`
	ContentVersionID types.ContentVersionID `json:"content_version_id"`
}

// AdminRestoreRequest is the JSON body for POST /api/v1/admin/content/restore.
type AdminRestoreRequest struct {
	AdminContentDataID    types.AdminContentID        `json:"admin_content_data_id"`
	AdminContentVersionID types.AdminContentVersionID `json:"admin_content_version_id"`
}

// RestoreResponse is the JSON response for content restore operations.
type RestoreResponse struct {
	Status          string   `json:"status"`
	ContentDataID   string   `json:"content_data_id"`
	RestoredVersion string   `json:"restored_version_id"`
	FieldsRestored  int      `json:"fields_restored"`
	UnmappedFields  []string `json:"unmapped_fields,omitempty"`
}

// AdminRestoreResponse is the JSON response for admin content restore operations.
type AdminRestoreResponse struct {
	Status             string   `json:"status"`
	AdminContentDataID string   `json:"admin_content_data_id"`
	RestoredVersion    string   `json:"restored_version_id"`
	FieldsRestored     int      `json:"fields_restored"`
	UnmappedFields     []string `json:"unmapped_fields,omitempty"`
}

///////////////////////////////
// HANDLERS
///////////////////////////////

// RestoreVersionHandler handles POST requests to restore content from a saved version snapshot.
// It deserializes the snapshot's content fields, matches them against the current live fields
// by field_id + content_data_id, updates the field values, and resets the content status to draft.
func RestoreVersionHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.ContentDataID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_data_id: %v", err), http.StatusBadRequest)
		return
	}
	if err := req.ContentVersionID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	d := db.ConfigDB(c)
	ctx := r.Context()
	ac := middleware.AuditContextFromRequest(r, c)

	// 1. Fetch the version.
	version, err := d.GetContentVersion(req.ContentVersionID)
	if err != nil {
		utility.DefaultLogger.Error("restore: get content version failed", err)
		http.Error(w, fmt.Sprintf("failed to get content version: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Verify the version belongs to the requested content_data_id.
	if version.ContentDataID != req.ContentDataID {
		http.Error(w, "content_version_id does not belong to the specified content_data_id", http.StatusBadRequest)
		return
	}

	// 3. Deserialize the snapshot.
	var snapshot Snapshot
	if err := json.Unmarshal([]byte(version.Snapshot), &snapshot); err != nil {
		utility.DefaultLogger.Error("restore: unmarshal snapshot failed", err)
		http.Error(w, fmt.Sprintf("failed to deserialize snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	// 4. Fetch current live content fields for this content_data_id.
	nullableID := types.NullableContentID{ID: req.ContentDataID, Valid: true}
	currentFieldsPtr, err := d.ListContentFieldsByContentData(nullableID)
	if err != nil {
		utility.DefaultLogger.Error("restore: list content fields failed", err)
		http.Error(w, fmt.Sprintf("failed to list current content fields: %v", err), http.StatusInternalServerError)
		return
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
			AuthorID:       user.UserID,
			DateCreated:    currentField.DateCreated,
			DateModified:   now,
			ContentFieldID: currentField.ContentFieldID,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("restore: update content field %s failed", currentField.ContentFieldID), updateErr)
			http.Error(w, fmt.Sprintf("failed to update content field %s: %v", currentField.ContentFieldID, updateErr), http.StatusInternalServerError)
			return
		}
		fieldsRestored++
	}

	// 6. Reset content status to draft.
	publishErr := d.UpdateContentDataPublishMeta(ctx, db.UpdateContentDataPublishMetaParams{
		Status:        types.ContentStatusDraft,
		PublishedAt:   types.Timestamp{},
		PublishedBy:   types.NullableUserID{},
		DateModified:  now,
		ContentDataID: req.ContentDataID,
	})
	if publishErr != nil {
		utility.DefaultLogger.Error("restore: update publish meta failed", publishErr)
		http.Error(w, fmt.Sprintf("failed to reset content status: %v", publishErr), http.StatusInternalServerError)
		return
	}

	// 7. Return response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(RestoreResponse{
		Status:          "restored",
		ContentDataID:   req.ContentDataID.String(),
		RestoredVersion: req.ContentVersionID.String(),
		FieldsRestored:  fieldsRestored,
		UnmappedFields:  unmappedFields,
	})
}

// AdminRestoreVersionHandler handles POST requests to restore admin content from a saved version snapshot.
// It deserializes the admin snapshot's content fields, matches them against the current live fields
// by admin_field_id + admin_content_data_id, updates the field values, and resets the content status to draft.
func AdminRestoreVersionHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req AdminRestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.AdminContentDataID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin_content_data_id: %v", err), http.StatusBadRequest)
		return
	}
	if err := req.AdminContentVersionID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin_content_version_id: %v", err), http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	d := db.ConfigDB(c)
	ctx := r.Context()
	ac := middleware.AuditContextFromRequest(r, c)

	// 1. Fetch the admin version.
	version, err := d.GetAdminContentVersion(req.AdminContentVersionID)
	if err != nil {
		utility.DefaultLogger.Error("admin restore: get admin content version failed", err)
		http.Error(w, fmt.Sprintf("failed to get admin content version: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Verify the version belongs to the requested admin_content_data_id.
	if version.AdminContentDataID != req.AdminContentDataID {
		http.Error(w, "admin_content_version_id does not belong to the specified admin_content_data_id", http.StatusBadRequest)
		return
	}

	// 3. Deserialize the admin snapshot.
	var snapshot AdminSnapshot
	if err := json.Unmarshal([]byte(version.Snapshot), &snapshot); err != nil {
		utility.DefaultLogger.Error("admin restore: unmarshal snapshot failed", err)
		http.Error(w, fmt.Sprintf("failed to deserialize admin snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	// 4. Fetch current live admin content fields.
	// There is no ListAdminContentFieldsByContentData method, so fetch by route
	// from the admin content data and filter by content_data_id.
	adminCD, err := d.GetAdminContentData(req.AdminContentDataID)
	if err != nil {
		utility.DefaultLogger.Error("admin restore: get admin content data failed", err)
		http.Error(w, fmt.Sprintf("failed to get admin content data: %v", err), http.StatusInternalServerError)
		return
	}

	if !adminCD.AdminRouteID.Valid {
		http.Error(w, "admin content data has no associated route", http.StatusBadRequest)
		return
	}

	allFieldsPtr, err := d.ListAdminContentFieldsByRoute(adminCD.AdminRouteID)
	if err != nil {
		utility.DefaultLogger.Error("admin restore: list admin content fields failed", err)
		http.Error(w, fmt.Sprintf("failed to list admin content fields: %v", err), http.StatusInternalServerError)
		return
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
			if cf.AdminContentDataID.ID != req.AdminContentDataID {
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
			AuthorID:            user.UserID,
			DateCreated:         currentField.DateCreated,
			DateModified:        now,
			AdminContentFieldID: currentField.AdminContentFieldID,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("admin restore: update admin content field %s failed", currentField.AdminContentFieldID), updateErr)
			http.Error(w, fmt.Sprintf("failed to update admin content field %s: %v", currentField.AdminContentFieldID, updateErr), http.StatusInternalServerError)
			return
		}
		fieldsRestored++
	}

	// 6. Reset admin content status to draft.
	publishErr := d.UpdateAdminContentDataPublishMeta(ctx, db.UpdateAdminContentDataPublishMetaParams{
		Status:             types.ContentStatusDraft,
		PublishedAt:        types.Timestamp{},
		PublishedBy:        types.NullableUserID{},
		DateModified:       now,
		AdminContentDataID: req.AdminContentDataID,
	})
	if publishErr != nil {
		utility.DefaultLogger.Error("admin restore: update publish meta failed", publishErr)
		http.Error(w, fmt.Sprintf("failed to reset admin content status: %v", publishErr), http.StatusInternalServerError)
		return
	}

	// 7. Return response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(AdminRestoreResponse{
		Status:             "restored",
		AdminContentDataID: req.AdminContentDataID.String(),
		RestoredVersion:    req.AdminContentVersionID.String(),
		FieldsRestored:     fieldsRestored,
		UnmappedFields:     unmappedFields,
	})
}
