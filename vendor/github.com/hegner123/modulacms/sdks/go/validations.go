package modula

import (
	"context"
	"fmt"
	"net/url"
)

// ValidationResource provides CRUD operations for reusable validation configurations.
// Validations define input rules (required, min/max length, patterns, etc.) that fields
// can reference by ID. This centralizes validation logic so multiple fields can share
// the same rules.
//
// ValidationResource embeds [Resource] for standard List, Get, Create, Update, Delete
// operations, and adds a search-by-name method.
// It is accessed via [Client].Validations.
type ValidationResource struct {
	*Resource[Validation, CreateValidationParams, UpdateValidationParams, ValidationID]
	http *httpClient
}

func newValidationResource(h *httpClient) *ValidationResource {
	return &ValidationResource{
		Resource: newResource[Validation, CreateValidationParams, UpdateValidationParams, ValidationID](h, "/api/v1/validations"),
		http:     h,
	}
}

// SearchByName returns validations matching the given name substring.
// The server performs a case-insensitive partial match against validation names.
func (r *ValidationResource) SearchByName(ctx context.Context, name string) ([]Validation, error) {
	params := url.Values{}
	params.Set("name", name)
	var result []Validation
	if err := r.http.get(ctx, "/api/v1/validations/search", params, &result); err != nil {
		return nil, fmt.Errorf("search validations by name %q: %w", name, err)
	}
	return result, nil
}

// AdminValidationResource provides CRUD operations for admin-side validation configurations.
// These validations serve the same purpose as public validations but operate within
// the admin content namespace.
//
// AdminValidationResource embeds [Resource] for standard List, Get, Create, Update, Delete
// operations, and adds a search-by-name method.
// It is accessed via [Client].AdminValidations.
type AdminValidationResource struct {
	*Resource[AdminValidation, CreateAdminValidationParams, UpdateAdminValidationParams, AdminValidationID]
	http *httpClient
}

func newAdminValidationResource(h *httpClient) *AdminValidationResource {
	return &AdminValidationResource{
		Resource: newResource[AdminValidation, CreateAdminValidationParams, UpdateAdminValidationParams, AdminValidationID](h, "/api/v1/admin/validations"),
		http:     h,
	}
}

// SearchByName returns admin validations matching the given name substring.
// The server performs a case-insensitive partial match against validation names.
func (r *AdminValidationResource) SearchByName(ctx context.Context, name string) ([]AdminValidation, error) {
	params := url.Values{}
	params.Set("name", name)
	var result []AdminValidation
	if err := r.http.get(ctx, "/api/v1/admin/validations/search", params, &result); err != nil {
		return nil, fmt.Errorf("search admin validations by name %q: %w", name, err)
	}
	return result, nil
}
