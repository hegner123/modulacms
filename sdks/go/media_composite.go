package modula

import (
	"context"
	"fmt"
	"net/url"
)

// MediaCompositeResource provides composite media operations such as
// reference scanning and delete with reference cleanup.
type MediaCompositeResource struct {
	http *httpClient
}

// GetReferences returns all content fields that reference the given media item.
func (m *MediaCompositeResource) GetReferences(ctx context.Context, id MediaID) (*MediaReferenceScanResponse, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result MediaReferenceScanResponse
	if err := m.http.get(ctx, "/api/v1/media/references", params, &result); err != nil {
		return nil, fmt.Errorf("media references %s: %w", string(id), err)
	}
	return &result, nil
}

// DeleteWithCleanup deletes a media item and cleans up all content field
// references that point to it.
func (m *MediaCompositeResource) DeleteWithCleanup(ctx context.Context, id MediaID) error {
	params := url.Values{}
	params.Set("q", string(id))
	params.Set("clean_refs", "true")
	if err := m.http.del(ctx, "/api/v1/media/", params); err != nil {
		return fmt.Errorf("media delete with cleanup %s: %w", string(id), err)
	}
	return nil
}
