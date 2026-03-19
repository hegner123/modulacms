package modula

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// MediaFoldersResource provides media folder operations beyond standard CRUD,
// including folder tree retrieval, listing media within a folder, and batch
// moving media items between folders.
// It is accessed via [Client].MediaFolders for specialized operations.
// Standard CRUD is available via [Client].MediaFoldersResource.
type MediaFoldersResource struct {
	http *httpClient
}

// Tree returns the full folder tree as a nested hierarchy.
// Root folders are returned at the top level with their children recursively nested.
func (r *MediaFoldersResource) Tree(ctx context.Context) ([]MediaFolderTreeNode, error) {
	var result []MediaFolderTreeNode
	if err := r.http.get(ctx, "/api/v1/media-folders/tree", nil, &result); err != nil {
		return nil, fmt.Errorf("media folder tree: %w", err)
	}
	return result, nil
}

// ListMedia returns a paginated list of media items within a specific folder.
func (r *MediaFoldersResource) ListMedia(ctx context.Context, folderID MediaFolderID, p PaginationParams) (*PaginatedResponse[Media], error) {
	params := url.Values{}
	params.Set("limit", strconv.FormatInt(p.Limit, 10))
	params.Set("offset", strconv.FormatInt(p.Offset, 10))
	var result PaginatedResponse[Media]
	if err := r.http.get(ctx, "/api/v1/media-folders/"+url.PathEscape(string(folderID))+"/media", params, &result); err != nil {
		return nil, fmt.Errorf("list media in folder %s: %w", folderID, err)
	}
	return &result, nil
}

// MoveMedia batch-moves media items to a target folder (or to root if FolderID is nil).
func (r *MediaFoldersResource) MoveMedia(ctx context.Context, params MoveMediaParams) (*MoveMediaResponse, error) {
	var result MoveMediaResponse
	if err := r.http.post(ctx, "/api/v1/media/move", params, &result); err != nil {
		return nil, fmt.Errorf("move media: %w", err)
	}
	return &result, nil
}
