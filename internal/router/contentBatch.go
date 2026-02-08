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

// BatchContentUpdateRequest is the JSON body for POST /api/v1/content/batch.
// At least one of ContentData or Fields must be present.
type BatchContentUpdateRequest struct {
	ContentDataID types.ContentID             `json:"content_data_id"`
	ContentData   *db.UpdateContentDataParams `json:"content_data,omitempty"`
	Fields        map[types.FieldID]string    `json:"fields,omitempty"`
}

// BatchContentUpdateResponse summarises what was applied.
// HTTP 200 is always returned; the caller checks fields_failed > 0 or
// content_data_error != "" for partial failures.
type BatchContentUpdateResponse struct {
	ContentDataID      types.ContentID `json:"content_data_id"`
	ContentDataUpdated bool            `json:"content_data_updated"`
	ContentDataError   string          `json:"content_data_error,omitempty"`
	FieldsUpdated      int             `json:"fields_updated"`
	FieldsCreated      int             `json:"fields_created"`
	FieldsFailed       int             `json:"fields_failed"`
	Errors             []string        `json:"errors,omitempty"`
}

// ContentBatchHandler applies an optional content_data update plus a map of
// field value upserts in a single request. This mirrors the TUI pattern in
// HandleUpdateContentFromDialog (internal/cli/commands.go).
func ContentBatchHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	var req BatchContentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if req.ContentDataID.IsZero() {
		http.Error(w, "content_data_id is required", http.StatusBadRequest)
		return
	}

	if req.ContentData == nil && len(req.Fields) == 0 {
		http.Error(w, "at least one of content_data or fields must be provided", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ac := middleware.AuditContextFromRequest(r, c)
	d := db.ConfigDB(c)

	resp := BatchContentUpdateResponse{
		ContentDataID: req.ContentDataID,
	}

	// --- content_data update ---
	if req.ContentData != nil {
		if req.ContentData.ContentDataID != req.ContentDataID {
			req.ContentData.ContentDataID = req.ContentDataID
		}
		_, err := d.UpdateContentData(ctx, ac, *req.ContentData)
		if err != nil {
			utility.DefaultLogger.Error("batch: content_data update failed", err)
			resp.ContentDataError = err.Error()
			resp.Errors = append(resp.Errors, fmt.Sprintf("content_data: %v", err))
		} else {
			resp.ContentDataUpdated = true
		}
	}

	// --- field upserts ---
	if len(req.Fields) > 0 {
		contentDataID := types.NullableContentID{ID: req.ContentDataID, Valid: true}

		existingFields, err := d.ListContentFieldsByContentData(contentDataID)
		if err != nil {
			utility.DefaultLogger.Error("batch: failed to fetch existing content fields", err)
			resp.Errors = append(resp.Errors, fmt.Sprintf("list fields: %v", err))
			resp.FieldsFailed = len(req.Fields)
			writeJSON(w, resp)
			return
		}

		existingMap := make(map[string]db.ContentFields)
		if existingFields != nil {
			for _, cf := range *existingFields {
				if cf.FieldID.Valid {
					existingMap[string(cf.FieldID.ID)] = cf
				}
			}
		}

		// Fetch the content_data row so we have RouteID for new field creates.
		var routeID types.NullableRouteID
		contentData, err := d.GetContentData(req.ContentDataID)
		if err != nil {
			utility.DefaultLogger.Error("batch: failed to fetch content_data for route_id", err)
			resp.Errors = append(resp.Errors, fmt.Sprintf("get content_data: %v", err))
			resp.FieldsFailed = len(req.Fields)
			writeJSON(w, resp)
			return
		}
		routeID = contentData.RouteID

		// Derive author_id from the authenticated user (same as audit context).
		var authorID types.NullableUserID
		if user := middleware.AuthenticatedUser(ctx); user != nil {
			authorID = types.NullableUserID{ID: user.UserID, Valid: !user.UserID.IsZero()}
		}

		for fieldID, value := range req.Fields {
			if existing, ok := existingMap[string(fieldID)]; ok {
				// Update existing field
				_, updateErr := d.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
					ContentFieldID: existing.ContentFieldID,
					RouteID:        existing.RouteID,
					ContentDataID:  contentDataID,
					FieldID:        types.NullableFieldID{ID: fieldID, Valid: true},
					FieldValue:     value,
					AuthorID:       authorID,
					DateCreated:    existing.DateCreated,
					DateModified:   types.TimestampNow(),
				})
				if updateErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("batch: failed to update field %s", fieldID), updateErr)
					resp.FieldsFailed++
					resp.Errors = append(resp.Errors, fmt.Sprintf("update field %s: %v", fieldID, updateErr))
				} else {
					resp.FieldsUpdated++
				}
			} else {
				// Create new field
				created, createErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
					ContentDataID: contentDataID,
					FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
					FieldValue:    value,
					RouteID:       routeID,
					AuthorID:      authorID,
					DateCreated:   types.TimestampNow(),
					DateModified:  types.TimestampNow(),
				})
				if createErr != nil || created == nil || created.ContentFieldID.IsZero() {
					utility.DefaultLogger.Error(fmt.Sprintf("batch: failed to create field %s", fieldID), createErr)
					resp.FieldsFailed++
					resp.Errors = append(resp.Errors, fmt.Sprintf("create field %s: %v", fieldID, createErr))
				} else {
					resp.FieldsCreated++
				}
			}
		}
	}

	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}
