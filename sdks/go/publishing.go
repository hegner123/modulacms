package modula

import (
	"context"
	"net/url"
)

// PublishingResource provides publishing, versioning, and restore operations for content.
type PublishingResource struct {
	http   *httpClient
	prefix string // "content" or "admin/content"
}

// Publish publishes content, creating a new version snapshot.
func (p *PublishingResource) Publish(ctx context.Context, req PublishRequest) (*PublishResponse, error) {
	var resp PublishResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/publish", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdminPublish publishes admin content, creating a new version snapshot.
func (p *PublishingResource) AdminPublish(ctx context.Context, req AdminPublishRequest) (*AdminPublishResponse, error) {
	var resp AdminPublishResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/publish", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Unpublish removes the published state from content.
func (p *PublishingResource) Unpublish(ctx context.Context, req PublishRequest) (*PublishResponse, error) {
	var resp PublishResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/unpublish", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdminUnpublish removes the published state from admin content.
func (p *PublishingResource) AdminUnpublish(ctx context.Context, req AdminPublishRequest) (*AdminPublishResponse, error) {
	var resp AdminPublishResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/unpublish", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Schedule sets a future publication time for content.
func (p *PublishingResource) Schedule(ctx context.Context, req ScheduleRequest) (*ScheduleResponse, error) {
	var resp ScheduleResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/schedule", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdminSchedule sets a future publication time for admin content.
func (p *PublishingResource) AdminSchedule(ctx context.Context, req AdminScheduleRequest) (*AdminScheduleResponse, error) {
	var resp AdminScheduleResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/schedule", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListVersions returns all version snapshots for a given content data ID.
func (p *PublishingResource) ListVersions(ctx context.Context, contentDataID string) ([]ContentVersion, error) {
	var versions []ContentVersion
	params := url.Values{"q": {contentDataID}}
	if err := p.http.get(ctx, "/api/v1/"+p.prefix+"/versions", params, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// ListAdminVersions returns all version snapshots for a given admin content data ID.
func (p *PublishingResource) ListAdminVersions(ctx context.Context, adminContentDataID string) ([]AdminContentVersion, error) {
	var versions []AdminContentVersion
	params := url.Values{"q": {adminContentDataID}}
	if err := p.http.get(ctx, "/api/v1/"+p.prefix+"/versions", params, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// GetVersion retrieves a single content version by its ID.
func (p *PublishingResource) GetVersion(ctx context.Context, versionID string) (*ContentVersion, error) {
	var version ContentVersion
	params := url.Values{"q": {versionID}}
	if err := p.http.get(ctx, "/api/v1/"+p.prefix+"/versions/", params, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// GetAdminVersion retrieves a single admin content version by its ID.
func (p *PublishingResource) GetAdminVersion(ctx context.Context, versionID string) (*AdminContentVersion, error) {
	var version AdminContentVersion
	params := url.Values{"q": {versionID}}
	if err := p.http.get(ctx, "/api/v1/"+p.prefix+"/versions/", params, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// CreateVersion manually creates a new version snapshot for content.
func (p *PublishingResource) CreateVersion(ctx context.Context, req CreateVersionRequest) (*ContentVersion, error) {
	var version ContentVersion
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/versions", req, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// CreateAdminVersion manually creates a new version snapshot for admin content.
func (p *PublishingResource) CreateAdminVersion(ctx context.Context, req CreateAdminVersionRequest) (*AdminContentVersion, error) {
	var version AdminContentVersion
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/versions", req, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// DeleteVersion removes a content version by its ID.
func (p *PublishingResource) DeleteVersion(ctx context.Context, versionID string) error {
	params := url.Values{"q": {versionID}}
	return p.http.del(ctx, "/api/v1/"+p.prefix+"/versions/", params)
}

// Restore restores content to a previous version.
func (p *PublishingResource) Restore(ctx context.Context, req RestoreRequest) (*RestoreResponse, error) {
	var resp RestoreResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/restore", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdminRestore restores admin content to a previous version.
func (p *PublishingResource) AdminRestore(ctx context.Context, req AdminRestoreRequest) (*AdminRestoreResponse, error) {
	var resp AdminRestoreResponse
	if err := p.http.post(ctx, "/api/v1/"+p.prefix+"/restore", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
