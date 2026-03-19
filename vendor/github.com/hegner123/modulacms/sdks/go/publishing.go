package modula

import (
	"context"
	"net/url"
)

// PublishingResource provides publishing, versioning, and restore operations for content.
// It manages the content lifecycle: draft -> published -> scheduled, with version
// snapshots for history tracking and rollback.
//
// Two instances exist on [Client]: Publishing (for public content) and AdminPublishing
// (for admin content). The methods are identical; they operate on different content tables.
//
// Version snapshots are immutable point-in-time copies of content field values,
// created automatically on publish and available for manual creation. Restoring a
// version replaces the current draft with the snapshot's field values.
type PublishingResource struct {
	http   *httpClient
	prefix string // "content" or "admin/content"
}

// Publish transitions content from draft to published status, creating an immutable
// version snapshot of the current field values. The content becomes immediately
// available through the public content delivery API.
func (p *PublishingResource) Publish(ctx context.Context, req PublishRequest) (*PublishResponse, error) {
	var resp PublishResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/publish", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdminPublish transitions admin content from draft to published status, creating
// an immutable version snapshot. This is the admin-content equivalent of [PublishingResource.Publish].
func (p *PublishingResource) AdminPublish(ctx context.Context, req AdminPublishRequest) (*AdminPublishResponse, error) {
	var resp AdminPublishResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/publish", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Unpublish reverts published content back to draft status, removing it from the
// public content delivery API. The existing version snapshots are preserved.
func (p *PublishingResource) Unpublish(ctx context.Context, req PublishRequest) (*PublishResponse, error) {
	var resp PublishResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/unpublish", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdminUnpublish reverts published admin content back to draft status.
// This is the admin-content equivalent of [PublishingResource.Unpublish].
func (p *PublishingResource) AdminUnpublish(ctx context.Context, req AdminPublishRequest) (*AdminPublishResponse, error) {
	var resp AdminPublishResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/unpublish", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Schedule sets a future publication time for content. The content remains in draft
// status until the scheduled time, when the server automatically publishes it.
// Scheduling creates a version snapshot at the time of the schedule call.
func (p *PublishingResource) Schedule(ctx context.Context, req ScheduleRequest) (*ScheduleResponse, error) {
	var resp ScheduleResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/schedule", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdminSchedule sets a future publication time for admin content.
// This is the admin-content equivalent of [PublishingResource.Schedule].
func (p *PublishingResource) AdminSchedule(ctx context.Context, req AdminScheduleRequest) (*AdminScheduleResponse, error) {
	var resp AdminScheduleResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/schedule", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListVersions returns all version snapshots for a given content data ID, ordered
// by creation time (newest first). Each [ContentVersion] contains the snapshot's
// field values and metadata (author, timestamp, publish status at time of creation).
func (p *PublishingResource) ListVersions(ctx context.Context, contentDataID string) ([]ContentVersion, error) {
	var versions []ContentVersion
	params := url.Values{"q": {contentDataID}}
	if err := p.http.get(ctx, "/api/v1/"+p.prefix+"/versions", params, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// ListAdminVersions returns all version snapshots for a given admin content data ID.
// This is the admin-content equivalent of [PublishingResource.ListVersions].
func (p *PublishingResource) ListAdminVersions(ctx context.Context, adminContentDataID string) ([]AdminContentVersion, error) {
	var versions []AdminContentVersion
	params := url.Values{"q": {adminContentDataID}}
	if err := p.http.get(ctx, "/api/v1/"+p.prefix+"/versions", params, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// GetVersion retrieves a single content version snapshot by its ID.
// Returns an [*ApiError] with status 404 if the version does not exist.
func (p *PublishingResource) GetVersion(ctx context.Context, versionID string) (*ContentVersion, error) {
	var version ContentVersion
	params := url.Values{"q": {versionID}}
	if err := p.http.get(ctx, "/api/v1/"+p.prefix+"/versions/", params, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// GetAdminVersion retrieves a single admin content version snapshot by its ID.
// This is the admin-content equivalent of [PublishingResource.GetVersion].
func (p *PublishingResource) GetAdminVersion(ctx context.Context, versionID string) (*AdminContentVersion, error) {
	var version AdminContentVersion
	params := url.Values{"q": {versionID}}
	if err := p.http.get(ctx, "/api/v1/"+p.prefix+"/versions/", params, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// CreateVersion manually creates a new version snapshot for content without
// changing its publish status. Use this to save a checkpoint of the current
// field values before making further edits.
func (p *PublishingResource) CreateVersion(ctx context.Context, req CreateVersionRequest) (*ContentVersion, error) {
	var version ContentVersion
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/versions", req, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// CreateAdminVersion manually creates a new version snapshot for admin content.
// This is the admin-content equivalent of [PublishingResource.CreateVersion].
func (p *PublishingResource) CreateAdminVersion(ctx context.Context, req CreateAdminVersionRequest) (*AdminContentVersion, error) {
	var version AdminContentVersion
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/versions", req, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// DeleteVersion permanently removes a content version snapshot by its ID.
// This does not affect the current content; it only removes the historical snapshot.
func (p *PublishingResource) DeleteVersion(ctx context.Context, versionID string) error {
	params := url.Values{"q": {versionID}}
	return p.http.del(ctx, "/api/v1/"+p.prefix+"/versions/", params)
}

// Restore replaces the current draft content field values with those from a
// previous version snapshot. The content's publish status is not changed; you
// must call [PublishingResource.Publish] separately to make the restored content live.
// A new version snapshot is created before the restore to preserve the current state.
func (p *PublishingResource) Restore(ctx context.Context, req RestoreRequest) (*RestoreResponse, error) {
	var resp RestoreResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/restore", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdminRestore restores admin content to a previous version snapshot.
// This is the admin-content equivalent of [PublishingResource.Restore].
func (p *PublishingResource) AdminRestore(ctx context.Context, req AdminRestoreRequest) (*AdminRestoreResponse, error) {
	var resp AdminRestoreResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/restore", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
