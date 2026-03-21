package modula

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Admin Media Folder Types
// ---------------------------------------------------------------------------

// AdminMediaFolder represents a folder for organizing admin media assets.
// Mirrors [MediaFolder] but uses [AdminMediaFolderID] for the admin-specific
// media namespace. Folders form a tree hierarchy via ParentID, with root
// folders having a nil ParentID.
type AdminMediaFolder struct {
	FolderID     AdminMediaFolderID  `json:"folder_id"`
	Name         string              `json:"name"`
	ParentID     *AdminMediaFolderID `json:"parent_id"`
	DateCreated  Timestamp           `json:"date_created"`
	DateModified Timestamp           `json:"date_modified"`
}

// CreateAdminMediaFolderParams contains parameters for creating a new admin
// media folder. Name is required. ParentID determines placement in the folder
// hierarchy; nil creates a root-level folder.
type CreateAdminMediaFolderParams struct {
	Name     string              `json:"name"`
	ParentID *AdminMediaFolderID `json:"parent_id"`
}

// UpdateAdminMediaFolderParams contains parameters for renaming or moving an
// admin media folder. FolderID identifies the folder to update.
type UpdateAdminMediaFolderParams struct {
	FolderID AdminMediaFolderID  `json:"folder_id"`
	Name     string              `json:"name"`
	ParentID *AdminMediaFolderID `json:"parent_id"`
}

// AdminMediaFolderTreeNode represents a folder in the admin media folder tree
// with its children recursively nested.
type AdminMediaFolderTreeNode struct {
	AdminMediaFolder
	Children []AdminMediaFolderTreeNode `json:"children"`
}

// MoveAdminMediaParams contains parameters for batch-moving admin media items
// to a folder.
type MoveAdminMediaParams struct {
	MediaIDs []AdminMediaID      `json:"media_ids"`
	FolderID *AdminMediaFolderID `json:"folder_id"`
}

// MoveAdminMediaResponse contains the result of a batch admin media move
// operation.
type MoveAdminMediaResponse struct {
	Moved int `json:"moved"`
}

// ---------------------------------------------------------------------------
// Admin Media Folders Resource
// ---------------------------------------------------------------------------

// AdminMediaFoldersResource provides admin media folder operations beyond
// standard CRUD, including folder tree retrieval, listing media within a
// folder, and batch moving media items between folders.
// It is accessed via [Client].AdminMediaFolders for specialized operations.
// Standard CRUD is available via [Client].AdminMediaFoldersData.
type AdminMediaFoldersResource struct {
	http *httpClient
}

// Tree returns the full admin media folder tree as a nested hierarchy.
// Root folders are returned at the top level with their children recursively
// nested.
func (r *AdminMediaFoldersResource) Tree(ctx context.Context) ([]AdminMediaFolderTreeNode, error) {
	var result []AdminMediaFolderTreeNode
	if err := r.http.get(ctx, "/api/v1/adminmedia-folders/tree", nil, &result); err != nil {
		return nil, fmt.Errorf("admin media folder tree: %w", err)
	}
	return result, nil
}

// ListMedia returns a paginated list of admin media items within a specific
// folder.
func (r *AdminMediaFoldersResource) ListMedia(ctx context.Context, folderID AdminMediaFolderID, p PaginationParams) (*PaginatedResponse[AdminMedia], error) {
	params := url.Values{}
	params.Set("limit", strconv.FormatInt(p.Limit, 10))
	params.Set("offset", strconv.FormatInt(p.Offset, 10))
	var result PaginatedResponse[AdminMedia]
	if err := r.http.get(ctx, "/api/v1/adminmedia-folders/"+url.PathEscape(string(folderID))+"/media", params, &result); err != nil {
		return nil, fmt.Errorf("list admin media in folder %s: %w", folderID, err)
	}
	return &result, nil
}

// MoveMedia batch-moves admin media items to a target folder (or to root if
// FolderID is nil).
func (r *AdminMediaFoldersResource) MoveMedia(ctx context.Context, params MoveAdminMediaParams) (*MoveAdminMediaResponse, error) {
	var result MoveAdminMediaResponse
	if err := r.http.post(ctx, "/api/v1/adminmedia/move", params, &result); err != nil {
		return nil, fmt.Errorf("move admin media: %w", err)
	}
	return &result, nil
}
