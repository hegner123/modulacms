package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/webhooks"
)

// Publish builds a snapshot of the content tree, stores it as a versioned
// snapshot, and marks the content as published. Returns the created version.
func (s *ContentService) Publish(ctx context.Context, ac audited.AuditContext, contentID types.ContentID, locale string, userID types.UserID) (*db.ContentVersion, error) {
	cfg, err := s.mgr.Config()
	if err != nil {
		return nil, fmt.Errorf("publish: get config: %w", err)
	}
	retentionCap := cfg.VersionMaxPerContent()

	version, err := publishing.PublishContent(ctx, s.driver, contentID, locale, userID, ac, retentionCap, s.dispatcher, nil)
	if err != nil {
		if publishing.IsRevisionConflict(err) {
			return nil, &ConflictError{
				Resource: "content_data",
				ID:       string(contentID),
				Detail:   err.Error(),
			}
		}
		return nil, fmt.Errorf("publish content: %w", err)
	}
	return version, nil
}

// Unpublish clears the published flag and resets publish metadata to draft.
func (s *ContentService) Unpublish(ctx context.Context, ac audited.AuditContext, contentID types.ContentID, locale string, userID types.UserID) error {
	err := publishing.UnpublishContent(ctx, s.driver, contentID, locale, userID, ac, s.dispatcher, nil)
	if err != nil {
		return fmt.Errorf("unpublish content: %w", err)
	}
	return nil
}

// Schedule sets the publish_at field on a content data row for the scheduler
// to pick up. Validates that publishAt is in the future.
func (s *ContentService) Schedule(ctx context.Context, contentID types.ContentID, publishAt time.Time) error {
	if publishAt.Before(time.Now()) {
		return NewValidationError("publish_at", "must be in the future")
	}

	now := types.TimestampNow()
	err := s.driver.UpdateContentDataSchedule(ctx, db.UpdateContentDataScheduleParams{
		PublishAt:     types.NewTimestamp(publishAt),
		DateModified:  now,
		ContentDataID: contentID,
	})
	if err != nil {
		return fmt.Errorf("schedule content: %w", err)
	}

	if s.dispatcher != nil {
		s.dispatcher.Dispatch(ctx, webhooks.EventContentScheduled, map[string]any{
			"content_data_id": contentID.String(),
			"publish_at":      publishAt.UTC().Format(time.RFC3339),
		})
	}

	return nil
}
