package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Admin Content Field CRUD (on AdminContentService) ---

// GetField retrieves a single admin content field by ID.
func (s *AdminContentService) GetField(ctx context.Context, id types.AdminContentFieldID) (*db.AdminContentFields, error) {
	cf, err := s.driver.GetAdminContentField(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "admin_content_field", ID: string(id)}
	}
	return cf, nil
}

// ListFields returns all admin content fields.
func (s *AdminContentService) ListFields(ctx context.Context) (*[]db.AdminContentFields, error) {
	return s.driver.ListAdminContentFields()
}

// ListFieldsPaginated returns a paginated list of admin content fields.
func (s *AdminContentService) ListFieldsPaginated(ctx context.Context, params db.PaginationParams) (*db.PaginatedResponse[db.AdminContentFields], error) {
	items, err := s.driver.ListAdminContentFieldsPaginated(params)
	if err != nil {
		return nil, fmt.Errorf("list admin content fields paginated: %w", err)
	}
	total, err := s.driver.CountAdminContentFields()
	if err != nil {
		return nil, fmt.Errorf("count admin content fields: %w", err)
	}
	return &db.PaginatedResponse[db.AdminContentFields]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// CreateField creates a new admin content field.
func (s *AdminContentService) CreateField(ctx context.Context, ac audited.AuditContext, params db.CreateAdminContentFieldParams) (*db.AdminContentFields, error) {
	// Fill in defaults for fields that callers may omit.
	now := types.TimestampNow()
	if params.DateCreated.IsZero() {
		params.DateCreated = now
	}
	if params.DateModified.IsZero() {
		params.DateModified = now
	}
	if params.AuthorID.IsZero() && !ac.UserID.IsZero() {
		params.AuthorID = ac.UserID
	}

	// Resolve root_id from admin content_data when not set.
	if !params.RootID.Valid && params.AdminContentDataID.Valid {
		cd, lookupErr := s.driver.GetAdminContentData(params.AdminContentDataID.ID)
		if lookupErr == nil && cd != nil {
			params.RootID = cd.RootID
			if !params.AdminRouteID.Valid {
				params.AdminRouteID = cd.AdminRouteID
			}
		}
	}

	cf, err := s.driver.CreateAdminContentField(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create admin content field: %w", err)
	}
	return cf, nil
}

// UpdateField updates an existing admin content field.
func (s *AdminContentService) UpdateField(ctx context.Context, ac audited.AuditContext, params db.UpdateAdminContentFieldParams) (*db.AdminContentFields, error) {
	_, err := s.driver.UpdateAdminContentField(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("update admin content field: %w", err)
	}

	updated, err := s.driver.GetAdminContentField(params.AdminContentFieldID)
	if err != nil {
		return nil, fmt.Errorf("update admin field: re-fetch: %w", err)
	}
	return updated, nil
}

// DeleteField removes an admin content field by ID.
func (s *AdminContentService) DeleteField(ctx context.Context, ac audited.AuditContext, id types.AdminContentFieldID) error {
	if err := s.driver.DeleteAdminContentField(ctx, ac, id); err != nil {
		return fmt.Errorf("delete admin content field: %w", err)
	}
	return nil
}
