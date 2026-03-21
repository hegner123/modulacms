package modula

import (
	"context"
	"fmt"
	"net/url"
)

// MediaCompositeResource provides composite media operations that span the media
// and content tables atomically. Use these methods to safely inspect or remove media
// items that may be referenced by content fields.
// It is accessed via [Client].MediaComposite.
type MediaCompositeResource struct {
	http *httpClient
}

// GetReferences scans all content fields and returns those that reference the
// given media item. Use this before deleting a media item to understand the
// impact on content, or to build a "used by" UI showing where media is in use.
func (m *MediaCompositeResource) GetReferences(ctx context.Context, id MediaID) (*MediaReferenceScanResponse, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result MediaReferenceScanResponse
	if err := m.http.get(ctx, "/api/v1/media/references", params, &result); err != nil {
		return nil, fmt.Errorf("media references %s: %w", string(id), err)
	}
	return &result, nil
}

// DeleteWithCleanup deletes a media item from storage and the database, and
// atomically clears all content field references that point to it. This is the
// preferred way to delete media that is potentially in use, as it prevents
// broken references in content. Use [MediaCompositeResource.GetReferences] first
// to preview the impact.
func (m *MediaCompositeResource) DeleteWithCleanup(ctx context.Context, id MediaID) error {
	params := url.Values{}
	params.Set("q", string(id))
	params.Set("clean_refs", "true")
	if err := m.http.del(ctx, "/api/v1/media/", params); err != nil {
		return fmt.Errorf("media delete with cleanup %s: %w", string(id), err)
	}
	return nil
}
