package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/publishing"
)

// Publish builds a snapshot of the admin content tree, stores it as a versioned
// snapshot, and marks the admin content as published.
// When node_level_publish is disabled (default), this publishes the root and
// all descendants. When enabled, only the root node is published.
func (s *AdminContentService) Publish(ctx context.Context, ac audited.AuditContext, adminContentID types.AdminContentID, locale string, userID types.UserID) error {
	cfg, err := s.mgr.Config()
	if err != nil {
		return fmt.Errorf("admin publish: get config: %w", err)
	}
	retentionCap := cfg.VersionMaxPerContent()
	publishAll := !cfg.Node_Level_Publish

	err = publishing.PublishAdminContent(ctx, s.driver, adminContentID, locale, userID, ac, retentionCap, publishAll, s.dispatcher)
	if err != nil {
		if publishing.IsRevisionConflict(err) {
			return &ConflictError{
				Resource: "admin_content_data",
				ID:       string(adminContentID),
				Detail:   err.Error(),
			}
		}
		return fmt.Errorf("admin publish content: %w", err)
	}
	return nil
}

// PublishAll builds a snapshot and marks the root and all descendants as
// published, regardless of node_level_publish configuration.
func (s *AdminContentService) PublishAll(ctx context.Context, ac audited.AuditContext, adminContentID types.AdminContentID, locale string, userID types.UserID) error {
	cfg, err := s.mgr.Config()
	if err != nil {
		return fmt.Errorf("admin publish all: get config: %w", err)
	}
	retentionCap := cfg.VersionMaxPerContent()

	err = publishing.PublishAdminContent(ctx, s.driver, adminContentID, locale, userID, ac, retentionCap, true, s.dispatcher)
	if err != nil {
		if publishing.IsRevisionConflict(err) {
			return &ConflictError{
				Resource: "admin_content_data",
				ID:       string(adminContentID),
				Detail:   err.Error(),
			}
		}
		return fmt.Errorf("admin publish all content: %w", err)
	}
	return nil
}

// Unpublish clears the published flag and resets publish metadata to draft
// for admin content.
func (s *AdminContentService) Unpublish(ctx context.Context, ac audited.AuditContext, adminContentID types.AdminContentID, locale string, userID types.UserID) error {
	err := publishing.UnpublishAdminContent(ctx, s.driver, adminContentID, locale, userID, ac, s.dispatcher)
	if err != nil {
		return fmt.Errorf("admin unpublish content: %w", err)
	}
	return nil
}

// Schedule sets the publish_at field on an admin content data row for the
// scheduler to pick up. Validates that publishAt is in the future.
func (s *AdminContentService) Schedule(ctx context.Context, adminContentID types.AdminContentID, publishAt time.Time) error {
	if publishAt.Before(time.Now()) {
		return NewValidationError("publish_at", "must be in the future")
	}

	now := types.TimestampNow()
	err := s.driver.UpdateAdminContentDataSchedule(ctx, db.UpdateAdminContentDataScheduleParams{
		PublishAt:          types.NewTimestamp(publishAt),
		DateModified:       now,
		AdminContentDataID: adminContentID,
	})
	if err != nil {
		return fmt.Errorf("admin schedule content: %w", err)
	}
	return nil
}
