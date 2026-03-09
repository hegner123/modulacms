package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// CreateRelation creates a new content relation.
func (s *ContentService) CreateRelation(ctx context.Context, ac audited.AuditContext, params db.CreateContentRelationParams) (*db.ContentRelations, error) {
	created, err := s.driver.CreateContentRelation(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create content relation: %w", err)
	}
	return created, nil
}

// DeleteRelation deletes a content relation by ID.
func (s *ContentService) DeleteRelation(ctx context.Context, ac audited.AuditContext, id types.ContentRelationID) error {
	if err := s.driver.DeleteContentRelation(ctx, ac, id); err != nil {
		return fmt.Errorf("delete content relation: %w", err)
	}
	return nil
}

// GetRelation retrieves a content relation by ID.
func (s *ContentService) GetRelation(ctx context.Context, id types.ContentRelationID) (*db.ContentRelations, error) {
	rel, err := s.driver.GetContentRelation(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "content_relation", ID: string(id)}
	}
	return rel, nil
}

// ListRelationsBySource returns all content relations where the given ID is the source.
func (s *ContentService) ListRelationsBySource(ctx context.Context, sourceID types.ContentID) (*[]db.ContentRelations, error) {
	rels, err := s.driver.ListContentRelationsBySource(sourceID)
	if err != nil {
		return nil, fmt.Errorf("list content relations by source: %w", err)
	}
	return rels, nil
}

// ListRelationsByTarget returns all content relations where the given ID is the target.
func (s *ContentService) ListRelationsByTarget(ctx context.Context, targetID types.ContentID) (*[]db.ContentRelations, error) {
	rels, err := s.driver.ListContentRelationsByTarget(targetID)
	if err != nil {
		return nil, fmt.Errorf("list content relations by target: %w", err)
	}
	return rels, nil
}

// ListRelationsBySourceAndField returns all content relations for a source+field pair.
func (s *ContentService) ListRelationsBySourceAndField(ctx context.Context, sourceID types.ContentID, fieldID types.FieldID) (*[]db.ContentRelations, error) {
	rels, err := s.driver.ListContentRelationsBySourceAndField(sourceID, fieldID)
	if err != nil {
		return nil, fmt.Errorf("list content relations by source and field: %w", err)
	}
	return rels, nil
}

// UpdateRelationSortOrder updates the sort order of a content relation.
func (s *ContentService) UpdateRelationSortOrder(ctx context.Context, ac audited.AuditContext, params db.UpdateContentRelationSortOrderParams) error {
	if err := s.driver.UpdateContentRelationSortOrder(ctx, ac, params); err != nil {
		return fmt.Errorf("update content relation sort order: %w", err)
	}
	return nil
}
