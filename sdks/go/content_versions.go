package modula

import (
	"context"
	"fmt"
	"net/url"
)

// ContentVersionsResource provides access to content version history.
type ContentVersionsResource struct {
	http *httpClient
}

// ListByContent returns all versions for a given content item.
func (r *ContentVersionsResource) ListByContent(ctx context.Context, contentID ContentID) ([]ContentVersion, error) {
	params := url.Values{}
	params.Set("content_id", string(contentID))
	var result []ContentVersion
	if err := r.http.get(ctx, "/api/v1/contentversions", params, &result); err != nil {
		return nil, fmt.Errorf("list content versions by content %s: %w", string(contentID), err)
	}
	return result, nil
}
