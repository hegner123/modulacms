package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// CreateRelation creates a new admin content relation.
func (s *AdminContentService) CreateRelation(ctx context.Context, ac audited.AuditContext, params db.CreateAdminContentRelationParams) (*db.AdminContentRelations, error) {
	created, err := s.driver.CreateAdminContentRelation(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create admin content relation: %w", err)
	}
	return created, nil
}

// DeleteRelation deletes an admin content relation by ID.
func (s *AdminContentService) DeleteRelation(ctx context.Context, ac audited.AuditContext, id types.AdminContentRelationID) error {
	if err := s.driver.DeleteAdminContentRelation(ctx, ac, id); err != nil {
		return fmt.Errorf("delete admin content relation: %w", err)
	}
	return nil
}

// GetRelation retrieves an admin content relation by ID.
func (s *AdminContentService) GetRelation(ctx context.Context, id types.AdminContentRelationID) (*db.AdminContentRelations, error) {
	rel, err := s.driver.GetAdminContentRelation(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "admin_content_relation", ID: string(id)}
	}
	return rel, nil
}

// ListRelationsBySource returns all admin content relations where the given ID is the source.
func (s *AdminContentService) ListRelationsBySource(ctx context.Context, sourceID types.AdminContentID) (*[]db.AdminContentRelations, error) {
	rels, err := s.driver.ListAdminContentRelationsBySource(sourceID)
	if err != nil {
		return nil, fmt.Errorf("list admin content relations by source: %w", err)
	}
	return rels, nil
}

// ListRelationsByTarget returns all admin content relations where the given ID is the target.
func (s *AdminContentService) ListRelationsByTarget(ctx context.Context, targetID types.AdminContentID) (*[]db.AdminContentRelations, error) {
	rels, err := s.driver.ListAdminContentRelationsByTarget(targetID)
	if err != nil {
		return nil, fmt.Errorf("list admin content relations by target: %w", err)
	}
	return rels, nil
}

// ListRelationsBySourceAndField returns all admin content relations for a source+field pair.
func (s *AdminContentService) ListRelationsBySourceAndField(ctx context.Context, sourceID types.AdminContentID, fieldID types.AdminFieldID) (*[]db.AdminContentRelations, error) {
	rels, err := s.driver.ListAdminContentRelationsBySourceAndField(sourceID, fieldID)
	if err != nil {
		return nil, fmt.Errorf("list admin content relations by source and field: %w", err)
	}
	return rels, nil
}

// UpdateRelationSortOrder updates the sort order of an admin content relation.
func (s *AdminContentService) UpdateRelationSortOrder(ctx context.Context, ac audited.AuditContext, params db.UpdateAdminContentRelationSortOrderParams) error {
	if err := s.driver.UpdateAdminContentRelationSortOrder(ctx, ac, params); err != nil {
		return fmt.Errorf("update admin content relation sort order: %w", err)
	}
	return nil
}
