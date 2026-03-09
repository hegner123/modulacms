package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/hegner123/modulacms/internal/validation"
)

// BatchUpdateParams is the input for ContentService.BatchUpdate.
type BatchUpdateParams struct {
	ContentDataID types.ContentID
	ContentData   *db.UpdateContentDataParams
	Fields        map[types.FieldID]string
	AuthorID      types.UserID
}

// BatchUpdateResult summarises what was applied in a batch update.
type BatchUpdateResult struct {
	ContentDataID      types.ContentID `json:"content_data_id"`
	ContentDataUpdated bool            `json:"content_data_updated"`
	ContentDataError   string          `json:"content_data_error,omitempty"`
	FieldsUpdated      int             `json:"fields_updated"`
	FieldsCreated      int             `json:"fields_created"`
	FieldsFailed       int             `json:"fields_failed"`
	Errors             []string        `json:"errors,omitempty"`
}

// BatchUpdate applies an optional content_data update plus a map of field value
// upserts in a single operation. Returns partial results — the caller checks
// FieldsFailed > 0 or ContentDataError != "" for partial failures.
func (s *ContentService) BatchUpdate(ctx context.Context, ac audited.AuditContext, params BatchUpdateParams) (*BatchUpdateResult, error) {
	if params.ContentDataID.IsZero() {
		return nil, NewValidationError("content_data_id", "required")
	}
	if params.ContentData == nil && len(params.Fields) == 0 {
		return nil, NewValidationError("request", "at least one of content_data or fields must be provided")
	}

	result := &BatchUpdateResult{
		ContentDataID: params.ContentDataID,
	}

	// --- content_data update ---
	if params.ContentData != nil {
		if params.ContentData.ContentDataID != params.ContentDataID {
			params.ContentData.ContentDataID = params.ContentDataID
		}
		_, err := s.driver.UpdateContentData(ctx, ac, *params.ContentData)
		if err != nil {
			utility.DefaultLogger.Error("batch: content_data update failed", err)
			result.ContentDataError = err.Error()
			result.Errors = append(result.Errors, fmt.Sprintf("content_data: %v", err))
		} else {
			result.ContentDataUpdated = true
		}
	}

	// --- field upserts ---
	if len(params.Fields) == 0 {
		return result, nil
	}

	contentDataID := types.NullableContentID{ID: params.ContentDataID, Valid: true}

	existingFields, err := s.driver.ListContentFieldsByContentData(contentDataID)
	if err != nil {
		utility.DefaultLogger.Error("batch: failed to fetch existing content fields", err)
		result.Errors = append(result.Errors, fmt.Sprintf("list fields: %v", err))
		result.FieldsFailed = len(params.Fields)
		return result, nil
	}

	existingMap := make(map[string]db.ContentFields)
	if existingFields != nil {
		for _, cf := range *existingFields {
			if cf.FieldID.Valid {
				existingMap[string(cf.FieldID.ID)] = cf
			}
		}
	}

	// Fetch content_data row for RouteID.
	contentData, err := s.driver.GetContentData(params.ContentDataID)
	if err != nil {
		utility.DefaultLogger.Error("batch: failed to fetch content_data for route_id", err)
		result.Errors = append(result.Errors, fmt.Sprintf("get content_data: %v", err))
		result.FieldsFailed = len(params.Fields)
		return result, nil
	}
	routeID := contentData.RouteID

	// --- validation ---
	fieldIDs := make([]types.FieldID, 0, len(params.Fields))
	for fid := range params.Fields {
		fieldIDs = append(fieldIDs, fid)
	}

	fieldDefs, fieldDefsErr := s.driver.GetFieldsByIDs(ctx, fieldIDs)
	if fieldDefsErr != nil {
		utility.DefaultLogger.Error("batch: failed to fetch field definitions", fieldDefsErr)
		result.Errors = append(result.Errors, fmt.Sprintf("get field definitions: %v", fieldDefsErr))
		result.FieldsFailed = len(params.Fields)
		return result, nil
	}

	fieldDefMap := make(map[types.FieldID]db.Fields, len(fieldDefs))
	for _, fd := range fieldDefs {
		fieldDefMap[fd.FieldID] = fd
	}

	// Check for unknown field IDs.
	var unknownIDs []string
	for _, fid := range fieldIDs {
		if _, ok := fieldDefMap[fid]; !ok {
			unknownIDs = append(unknownIDs, string(fid))
		}
	}
	if len(unknownIDs) > 0 {
		return nil, NewValidationError("fields", fmt.Sprintf("unknown field IDs: %v", unknownIDs))
	}

	// Build validation inputs and validate.
	valInputs := make([]validation.FieldInput, 0, len(params.Fields))
	for fid, value := range params.Fields {
		fd := fieldDefMap[fid]
		valInputs = append(valInputs, validation.FieldInput{
			FieldID:    fid,
			Label:      fd.Label,
			FieldType:  fd.Type,
			Value:      value,
			Validation: fd.Validation,
			Data:       fd.Data,
		})
	}

	ve := validation.ValidateBatch(valInputs)
	if ve.HasErrors() {
		return nil, &ValidationError{
			Errors: validationFieldErrorsToService(ve),
		}
	}

	for fieldID, value := range params.Fields {
		if existing, ok := existingMap[string(fieldID)]; ok {
			_, updateErr := s.driver.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
				ContentFieldID: existing.ContentFieldID,
				RouteID:        existing.RouteID,
				ContentDataID:  contentDataID,
				FieldID:        types.NullableFieldID{ID: fieldID, Valid: true},
				FieldValue:     value,
				AuthorID:       params.AuthorID,
				DateCreated:    existing.DateCreated,
				DateModified:   types.TimestampNow(),
			})
			if updateErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("batch: failed to update field %s", fieldID), updateErr)
				result.FieldsFailed++
				result.Errors = append(result.Errors, fmt.Sprintf("update field %s: %v", fieldID, updateErr))
			} else {
				result.FieldsUpdated++
			}
		} else {
			created, createErr := s.driver.CreateContentField(ctx, ac, db.CreateContentFieldParams{
				ContentDataID: contentDataID,
				FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
				FieldValue:    value,
				RouteID:       routeID,
				AuthorID:      params.AuthorID,
				DateCreated:   types.TimestampNow(),
				DateModified:  types.TimestampNow(),
			})
			if createErr != nil || created == nil || created.ContentFieldID.IsZero() {
				utility.DefaultLogger.Error(fmt.Sprintf("batch: failed to create field %s", fieldID), createErr)
				result.FieldsFailed++
				result.Errors = append(result.Errors, fmt.Sprintf("create field %s: %v", fieldID, createErr))
			} else {
				result.FieldsCreated++
			}
		}
	}

	return result, nil
}

// validationFieldErrorsToService converts validation.ValidationErrors to service FieldErrors.
func validationFieldErrorsToService(ve validation.ValidationErrors) []FieldError {
	out := make([]FieldError, 0, len(ve.Fields))
	for _, fe := range ve.Fields {
		for _, msg := range fe.Messages {
			out = append(out, FieldError{Field: string(fe.FieldID), Message: msg})
		}
	}
	return out
}
