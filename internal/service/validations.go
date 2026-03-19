package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ValidationService manages reusable validation configs referenced by fields.
type ValidationService struct {
	driver db.DbDriver
}

// NewValidationService creates a new ValidationService.
func NewValidationService(driver db.DbDriver) *ValidationService {
	return &ValidationService{driver: driver}
}

// --- Public Validations ---

// CreateValidation validates input, creates the record, and returns it.
func (s *ValidationService) CreateValidation(ctx context.Context, ac audited.AuditContext, params db.CreateValidationParams) (*db.Validation, error) {
	var ve ValidationError
	if params.Name == "" {
		ve.Add("name", "Name is required")
	}
	if err := validateConfigJSON(params.Config); err != nil {
		ve.Add("config", err.Error())
	}
	if ve.HasErrors() {
		return nil, &ve
	}

	now := nowUTC()
	params.DateCreated = now
	params.DateModified = now

	v, err := s.driver.CreateValidation(ctx, ac, params)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("create validation: %w", err)}
	}
	return v, nil
}

// UpdateValidation validates input, updates the record, and returns it.
func (s *ValidationService) UpdateValidation(ctx context.Context, ac audited.AuditContext, params db.UpdateValidationParams) (*db.Validation, error) {
	var ve ValidationError
	if params.Name == "" {
		ve.Add("name", "Name is required")
	}
	if err := validateConfigJSON(params.Config); err != nil {
		ve.Add("config", err.Error())
	}
	if ve.HasErrors() {
		return nil, &ve
	}

	existing, err := s.driver.GetValidation(params.ValidationID)
	if err != nil || existing == nil {
		return nil, &NotFoundError{Resource: "validation", ID: params.ValidationID.String()}
	}

	params.DateCreated = existing.DateCreated
	params.DateModified = nowUTC()

	if _, err = s.driver.UpdateValidation(ctx, ac, params); err != nil {
		return nil, &InternalError{Err: fmt.Errorf("update validation: %w", err)}
	}

	updated, err := s.driver.GetValidation(params.ValidationID)
	if err != nil || updated == nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch updated validation: %w", err)}
	}
	return updated, nil
}

// DeleteValidation removes a validation by ID.
// ON DELETE SET NULL handles the FK — fields lose their validation reference.
func (s *ValidationService) DeleteValidation(ctx context.Context, ac audited.AuditContext, id types.ValidationID) error {
	if err := s.driver.DeleteValidation(ctx, ac, id); err != nil {
		return &InternalError{Err: fmt.Errorf("delete validation: %w", err)}
	}
	return nil
}

// GetValidation retrieves a validation by ID.
func (s *ValidationService) GetValidation(id types.ValidationID) (*db.Validation, error) {
	v, err := s.driver.GetValidation(id)
	if err != nil || v == nil {
		return nil, &NotFoundError{Resource: "validation", ID: id.String()}
	}
	return v, nil
}

// ListValidations returns all validations.
func (s *ValidationService) ListValidations() (*[]db.Validation, error) {
	return s.driver.ListValidations()
}

// ListValidationsPaginated returns a paginated list of validations.
func (s *ValidationService) ListValidationsPaginated(params db.PaginationParams) (*db.PaginatedResponse[db.Validation], error) {
	items, err := s.driver.ListValidationsPaginated(params)
	if err != nil {
		return nil, fmt.Errorf("list validations paginated: %w", err)
	}
	total, err := s.driver.CountValidations()
	if err != nil {
		return nil, fmt.Errorf("count validations: %w", err)
	}
	return &db.PaginatedResponse[db.Validation]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// ListValidationsByName searches validations by name substring.
func (s *ValidationService) ListValidationsByName(name string) (*[]db.Validation, error) {
	return s.driver.ListValidationsByName(name)
}

// --- Admin Validations ---

// CreateAdminValidation validates input, creates the record, and returns it.
func (s *ValidationService) CreateAdminValidation(ctx context.Context, ac audited.AuditContext, params db.CreateAdminValidationParams) (*db.AdminValidation, error) {
	var ve ValidationError
	if params.Name == "" {
		ve.Add("name", "Name is required")
	}
	if err := validateConfigJSON(params.Config); err != nil {
		ve.Add("config", err.Error())
	}
	if ve.HasErrors() {
		return nil, &ve
	}

	now := nowUTC()
	params.DateCreated = now
	params.DateModified = now

	v, err := s.driver.CreateAdminValidation(ctx, ac, params)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("create admin validation: %w", err)}
	}
	return v, nil
}

// UpdateAdminValidation validates input, updates the record, and returns it.
func (s *ValidationService) UpdateAdminValidation(ctx context.Context, ac audited.AuditContext, params db.UpdateAdminValidationParams) (*db.AdminValidation, error) {
	var ve ValidationError
	if params.Name == "" {
		ve.Add("name", "Name is required")
	}
	if err := validateConfigJSON(params.Config); err != nil {
		ve.Add("config", err.Error())
	}
	if ve.HasErrors() {
		return nil, &ve
	}

	existing, err := s.driver.GetAdminValidation(params.AdminValidationID)
	if err != nil || existing == nil {
		return nil, &NotFoundError{Resource: "admin_validation", ID: params.AdminValidationID.String()}
	}

	params.DateCreated = existing.DateCreated
	params.DateModified = nowUTC()

	if _, err = s.driver.UpdateAdminValidation(ctx, ac, params); err != nil {
		return nil, &InternalError{Err: fmt.Errorf("update admin validation: %w", err)}
	}

	updated, err := s.driver.GetAdminValidation(params.AdminValidationID)
	if err != nil || updated == nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch updated admin validation: %w", err)}
	}
	return updated, nil
}

// DeleteAdminValidation removes an admin validation by ID.
func (s *ValidationService) DeleteAdminValidation(ctx context.Context, ac audited.AuditContext, id types.AdminValidationID) error {
	if err := s.driver.DeleteAdminValidation(ctx, ac, id); err != nil {
		return &InternalError{Err: fmt.Errorf("delete admin validation: %w", err)}
	}
	return nil
}

// GetAdminValidation retrieves an admin validation by ID.
func (s *ValidationService) GetAdminValidation(id types.AdminValidationID) (*db.AdminValidation, error) {
	v, err := s.driver.GetAdminValidation(id)
	if err != nil || v == nil {
		return nil, &NotFoundError{Resource: "admin_validation", ID: id.String()}
	}
	return v, nil
}

// ListAdminValidations returns all admin validations.
func (s *ValidationService) ListAdminValidations() (*[]db.AdminValidation, error) {
	return s.driver.ListAdminValidations()
}

// ListAdminValidationsPaginated returns a paginated list of admin validations.
func (s *ValidationService) ListAdminValidationsPaginated(params db.PaginationParams) (*db.PaginatedResponse[db.AdminValidation], error) {
	items, err := s.driver.ListAdminValidationsPaginated(params)
	if err != nil {
		return nil, fmt.Errorf("list admin validations paginated: %w", err)
	}
	total, err := s.driver.CountAdminValidations()
	if err != nil {
		return nil, fmt.Errorf("count admin validations: %w", err)
	}
	return &db.PaginatedResponse[db.AdminValidation]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// ListAdminValidationsByName searches admin validations by name substring.
func (s *ValidationService) ListAdminValidationsByName(name string) (*[]db.AdminValidation, error) {
	return s.driver.ListAdminValidationsByName(name)
}

// --- Helpers ---

// ResolveValidationConfig fetches the config JSON for a field's validation_id.
// Returns empty string if the validation_id is null/zero.
func (s *ValidationService) ResolveValidationConfig(id types.NullableValidationID) string {
	if !id.Valid || id.ID.IsZero() {
		return ""
	}
	v, err := s.driver.GetValidation(id.ID)
	if err != nil || v == nil {
		return ""
	}
	return v.Config
}

// ResolveAdminValidationConfig fetches the config JSON for an admin field's validation_id.
// Returns empty string if the validation_id is null/zero.
func (s *ValidationService) ResolveAdminValidationConfig(id types.NullableAdminValidationID) string {
	if !id.Valid || id.ID.IsZero() {
		return ""
	}
	v, err := s.driver.GetAdminValidation(id.ID)
	if err != nil || v == nil {
		return ""
	}
	return v.Config
}

// validateConfigJSON parses and validates a validation config JSON string.
func validateConfigJSON(config string) error {
	if config == "" || config == types.EmptyJSON {
		return nil
	}
	vc, err := types.ParseValidationConfig(config)
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return types.ValidateValidationConfig(vc)
}
