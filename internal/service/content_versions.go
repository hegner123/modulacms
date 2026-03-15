package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/webhooks"
)

// ListVersions returns all content versions for a given content data item.
func (s *ContentService) ListVersions(ctx context.Context, contentID types.ContentID) (*[]db.ContentVersion, error) {
	versions, err := s.driver.ListContentVersionsByContent(contentID)
	if err != nil {
		return nil, fmt.Errorf("list content versions: %w", err)
	}
	return versions, nil
}

// GetVersion retrieves a single content version by ID.
func (s *ContentService) GetVersion(ctx context.Context, versionID types.ContentVersionID) (*db.ContentVersion, error) {
	version, err := s.driver.GetContentVersion(versionID)
	if err != nil {
		return nil, &NotFoundError{Resource: "content_version", ID: string(versionID)}
	}
	return version, nil
}

// CreateVersion builds a snapshot from live tables and creates a manual
// content version. Prunes excess versions asynchronously.
func (s *ContentService) CreateVersion(ctx context.Context, ac audited.AuditContext, contentID types.ContentID, locale string, label string, userID types.UserID) (*db.ContentVersion, error) {
	snapshot, err := publishing.BuildSnapshot(s.driver, ctx, contentID, locale)
	if err != nil {
		return nil, fmt.Errorf("create version: build snapshot: %w", err)
	}

	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("create version: marshal snapshot: %w", err)
	}

	maxVersion, err := s.driver.GetMaxVersionNumber(contentID, locale)
	if err != nil {
		return nil, fmt.Errorf("create version: get max version number: %w", err)
	}
	nextVersion := maxVersion + 1

	now := types.TimestampNow()
	version, err := s.driver.CreateContentVersion(ctx, ac, db.CreateContentVersionParams{
		ContentDataID: contentID,
		VersionNumber: nextVersion,
		Locale:        locale,
		Snapshot:      string(snapshotBytes),
		Trigger:       "manual",
		Label:         label,
		Published:     false,
		PublishedBy:   types.NullableUserID{ID: userID, Valid: true},
		DateCreated:   now,
	})
	if err != nil {
		return nil, fmt.Errorf("create content version: %w", err)
	}

	cfg, cfgErr := s.mgr.Config()
	if cfgErr == nil {
		retentionCap := cfg.VersionMaxPerContent()
		go publishing.PruneExcessVersions(s.driver, contentID, "", retentionCap)
	}

	if s.dispatcher != nil {
		s.dispatcher.Dispatch(ctx, webhooks.EventVersionCreated, map[string]any{
			"content_data_id":    contentID.String(),
			"content_version_id": version.ContentVersionID.String(),
			"version_number":     version.VersionNumber,
			"locale":             locale,
			"trigger":            "manual",
		})
	}

	return version, nil
}

// DeleteVersion removes a content version. Rejects deletion of published versions.
func (s *ContentService) DeleteVersion(ctx context.Context, ac audited.AuditContext, versionID types.ContentVersionID) error {
	version, err := s.driver.GetContentVersion(versionID)
	if err != nil {
		return &NotFoundError{Resource: "content_version", ID: string(versionID)}
	}

	if version.Published {
		return &ConflictError{
			Resource: "content_version",
			ID:       string(versionID),
			Detail:   "cannot delete a published version",
		}
	}

	if err := s.driver.DeleteContentVersion(ctx, ac, versionID); err != nil {
		return fmt.Errorf("delete content version: %w", err)
	}
	return nil
}

// RestoreVersion restores content from a saved version snapshot.
func (s *ContentService) RestoreVersion(ctx context.Context, ac audited.AuditContext, contentID types.ContentID, versionID types.ContentVersionID, userID types.UserID) (*publishing.RestoreResult, error) {
	result, err := publishing.RestoreContent(ctx, s.driver, contentID, versionID, userID, ac)
	if err != nil {
		return nil, fmt.Errorf("restore content version: %w", err)
	}
	return result, nil
}
