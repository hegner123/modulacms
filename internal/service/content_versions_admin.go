package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/publishing"
)

// ListVersions returns all admin content versions for a given admin content data item.
func (s *AdminContentService) ListVersions(ctx context.Context, adminContentID types.AdminContentID) (*[]db.AdminContentVersion, error) {
	versions, err := s.driver.ListAdminContentVersionsByContent(adminContentID)
	if err != nil {
		return nil, fmt.Errorf("list admin content versions: %w", err)
	}
	return versions, nil
}

// GetVersion retrieves a single admin content version by ID.
func (s *AdminContentService) GetVersion(ctx context.Context, versionID types.AdminContentVersionID) (*db.AdminContentVersion, error) {
	version, err := s.driver.GetAdminContentVersion(versionID)
	if err != nil {
		return nil, &NotFoundError{Resource: "admin_content_version", ID: string(versionID)}
	}
	return version, nil
}

// CreateVersion builds a snapshot from live admin tables and creates a manual
// admin content version. Prunes excess versions asynchronously.
func (s *AdminContentService) CreateVersion(ctx context.Context, ac audited.AuditContext, adminContentID types.AdminContentID, locale string, label string, userID types.UserID) (*db.AdminContentVersion, error) {
	snapshot, err := publishing.BuildAdminSnapshot(s.driver, ctx, adminContentID, locale)
	if err != nil {
		return nil, fmt.Errorf("admin create version: build snapshot: %w", err)
	}

	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("admin create version: marshal snapshot: %w", err)
	}

	maxVersion, err := s.driver.GetAdminMaxVersionNumber(adminContentID, locale)
	if err != nil {
		return nil, fmt.Errorf("admin create version: get max version number: %w", err)
	}
	nextVersion := maxVersion + 1

	now := types.TimestampNow()
	version, err := s.driver.CreateAdminContentVersion(ctx, ac, db.CreateAdminContentVersionParams{
		AdminContentDataID: adminContentID,
		VersionNumber:      nextVersion,
		Locale:             locale,
		Snapshot:           string(snapshotBytes),
		Trigger:            "manual",
		Label:              label,
		Published:          false,
		PublishedBy:        types.NullableUserID{ID: userID, Valid: true},
		DateCreated:        now,
	})
	if err != nil {
		return nil, fmt.Errorf("create admin content version: %w", err)
	}

	cfg, cfgErr := s.mgr.Config()
	if cfgErr == nil {
		retentionCap := cfg.VersionMaxPerContent()
		go publishing.PruneExcessAdminVersions(s.driver, adminContentID, "", retentionCap)
	}

	return version, nil
}

// DeleteVersion removes an admin content version. Rejects deletion of published versions.
func (s *AdminContentService) DeleteVersion(ctx context.Context, ac audited.AuditContext, versionID types.AdminContentVersionID) error {
	version, err := s.driver.GetAdminContentVersion(versionID)
	if err != nil {
		return &NotFoundError{Resource: "admin_content_version", ID: string(versionID)}
	}

	if version.Published {
		return &ConflictError{
			Resource: "admin_content_version",
			ID:       string(versionID),
			Detail:   "cannot delete a published version",
		}
	}

	if err := s.driver.DeleteAdminContentVersion(ctx, ac, versionID); err != nil {
		return fmt.Errorf("delete admin content version: %w", err)
	}
	return nil
}

// RestoreVersion restores admin content from a saved version snapshot.
func (s *AdminContentService) RestoreVersion(ctx context.Context, ac audited.AuditContext, adminContentID types.AdminContentID, versionID types.AdminContentVersionID, userID types.UserID) (*publishing.RestoreResult, error) {
	result, err := publishing.RestoreAdminContent(ctx, s.driver, adminContentID, versionID, userID, ac)
	if err != nil {
		return nil, fmt.Errorf("admin restore content version: %w", err)
	}
	return result, nil
}
