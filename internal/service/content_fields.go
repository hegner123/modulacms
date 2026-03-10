package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/validation"
)

// --- Content Field CRUD (on ContentService) ---

// GetField retrieves a single content field by ID.
func (s *ContentService) GetField(ctx context.Context, id types.ContentFieldID) (*db.ContentFields, error) {
	cf, err := s.driver.GetContentField(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "content_field", ID: string(id)}
	}
	return cf, nil
}

// ListFields returns all content fields.
func (s *ContentService) ListFields(ctx context.Context) (*[]db.ContentFields, error) {
	return s.driver.ListContentFields()
}

// ListFieldsPaginated returns a paginated list of content fields.
func (s *ContentService) ListFieldsPaginated(ctx context.Context, params db.PaginationParams) (*db.PaginatedResponse[db.ContentFields], error) {
	items, err := s.driver.ListContentFieldsPaginated(params)
	if err != nil {
		return nil, fmt.Errorf("list content fields paginated: %w", err)
	}
	total, err := s.driver.CountContentFields()
	if err != nil {
		return nil, fmt.Errorf("count content fields: %w", err)
	}
	return &db.PaginatedResponse[db.ContentFields]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// CreateField creates a new content field after validating the value against
// the field definition's type and validation rules.
func (s *ContentService) CreateField(ctx context.Context, ac audited.AuditContext, params db.CreateContentFieldParams) (*db.ContentFields, error) {
	if !params.FieldID.Valid || params.FieldID.ID.IsZero() {
		return nil, NewValidationError("field_id", "field_id is required")
	}

	fieldDef, err := s.driver.GetField(params.FieldID.ID)
	if err != nil {
		return nil, &NotFoundError{Resource: "field", ID: string(params.FieldID.ID)}
	}

	if fe := validateContentFieldValue(fieldDef, params.FieldValue); fe != nil {
		return nil, fe
	}

	cf, err := s.driver.CreateContentField(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create content field: %w", err)
	}
	return cf, nil
}

// UpdateField updates an existing content field after validating the value.
func (s *ContentService) UpdateField(ctx context.Context, ac audited.AuditContext, params db.UpdateContentFieldParams) (*db.ContentFields, error) {
	if !params.FieldID.Valid || params.FieldID.ID.IsZero() {
		return nil, NewValidationError("field_id", "field_id is required")
	}

	fieldDef, err := s.driver.GetField(params.FieldID.ID)
	if err != nil {
		return nil, &NotFoundError{Resource: "field", ID: string(params.FieldID.ID)}
	}

	if fe := validateContentFieldValue(fieldDef, params.FieldValue); fe != nil {
		return nil, fe
	}

	_, err = s.driver.UpdateContentField(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("update content field: %w", err)
	}

	updated, err := s.driver.GetContentField(params.ContentFieldID)
	if err != nil {
		return nil, fmt.Errorf("update field: re-fetch: %w", err)
	}
	return updated, nil
}

// DeleteField removes a content field by ID.
func (s *ContentService) DeleteField(ctx context.Context, ac audited.AuditContext, id types.ContentFieldID) error {
	if err := s.driver.DeleteContentField(ctx, ac, id); err != nil {
		return fmt.Errorf("delete content field: %w", err)
	}
	return nil
}

// ListFieldsByContentDataAndLocale returns content fields filtered by content data ID and locale.
func (s *ContentService) ListFieldsByContentDataAndLocale(ctx context.Context, contentDataID types.NullableContentID, locale string) (*[]db.ContentFields, error) {
	return s.driver.ListContentFieldsByContentDataAndLocale(contentDataID, locale)
}

// validateContentFieldValue runs validation.ValidateField and converts the
// result to a service-layer ValidationError.
func validateContentFieldValue(fieldDef *db.Fields, value string) *ValidationError {
	fe := validation.ValidateField(validation.FieldInput{
		FieldID:    fieldDef.FieldID,
		Label:      fieldDef.Label,
		FieldType:  fieldDef.Type,
		Value:      value,
		Validation: fieldDef.Validation,
		Data:       fieldDef.Data,
	})
	if fe != nil {
		return NewValidationError(fe.Label, fe.Error())
	}
	return nil
}
