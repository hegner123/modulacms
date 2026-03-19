package modula

import (
	"context"
	"fmt"
	"net/url"
)

// ContentVersionsResource provides read access to content version history.
// Each time content is published or a snapshot is triggered, the server creates
// a [ContentVersion] record containing the full serialized state of the content
// at that point in time.
//
// Version history is immutable: versions are never updated or deleted through
// normal operation. This provides a complete audit trail and enables rollback
// to any previous published state.
//
// Access this resource via [Client].ContentVersions:
//
//	versions, err := client.ContentVersions.ListByContent(ctx, contentID)
type ContentVersionsResource struct {
	http *httpClient
}

// ListByContent returns all versions for a given content item, ordered by
// version number descending (most recent first).
//
// Each returned [ContentVersion] includes the full snapshot JSON, version
// number, locale, trigger reason, and whether the version is currently
// the published version.
//
// Returns an empty slice (not nil) if the content item exists but has no
// versions. Returns an [*ApiError] with status 404 if the content item
// does not exist.
func (r *ContentVersionsResource) ListByContent(ctx context.Context, contentID ContentID) ([]ContentVersion, error) {
	params := url.Values{}
	params.Set("content_id", string(contentID))
	var result []ContentVersion
	if err := r.http.get(ctx, "/api/v1/contentversions", params, &result); err != nil {
		return nil, fmt.Errorf("list content versions by content %s: %w", string(contentID), err)
	}
	return result, nil
}
