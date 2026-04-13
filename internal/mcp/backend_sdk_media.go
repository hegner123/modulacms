package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// MediaBackend
// ---------------------------------------------------------------------------

type sdkMediaBackend struct {
	client *modula.Client
}

func (b *sdkMediaBackend) ListMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.client.Media.ListPaginated(ctx, modula.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaBackend) GetMedia(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Media.Get(ctx, modula.MediaID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaBackend) UpdateMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateMediaParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update media params: %w", err)
	}
	result, err := b.client.Media.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaBackend) DeleteMedia(ctx context.Context, id string) error {
	return b.client.Media.Delete(ctx, modula.MediaID(id))
}

func (b *sdkMediaBackend) UploadMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error) {
	result, err := b.client.MediaUpload.Upload(ctx, reader, filename, nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaBackend) MediaHealth(ctx context.Context) (json.RawMessage, error) {
	return b.client.MediaAdmin.Health(ctx)
}

func (b *sdkMediaBackend) MediaCleanup(ctx context.Context) (json.RawMessage, error) {
	return b.client.MediaAdmin.Cleanup(ctx)
}

func (b *sdkMediaBackend) MediaCleanupCheck(ctx context.Context) (json.RawMessage, error) {
	return b.MediaHealth(ctx)
}

func (b *sdkMediaBackend) ListMediaDimensions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.MediaDimensions.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaBackend) GetMediaDimension(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.MediaDimensions.Get(ctx, modula.MediaDimensionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaBackend) CreateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateMediaDimensionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create media dimension params: %w", err)
	}
	result, err := b.client.MediaDimensions.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaBackend) UpdateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateMediaDimensionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update media dimension params: %w", err)
	}
	result, err := b.client.MediaDimensions.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaBackend) DeleteMediaDimension(ctx context.Context, id string) error {
	return b.client.MediaDimensions.Delete(ctx, modula.MediaDimensionID(id))
}

func (b *sdkMediaBackend) DownloadMedia(ctx context.Context, id string) (json.RawMessage, error) {
	url, err := b.client.MediaDownload.GetURL(ctx, modula.MediaID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"url": url})
}

func (b *sdkMediaBackend) GetMediaFull(ctx context.Context) (json.RawMessage, error) {
	return b.client.MediaFull.List(ctx)
}

func (b *sdkMediaBackend) GetMediaReferences(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.MediaComposite.GetReferences(ctx, modula.MediaID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaBackend) ReprocessMedia(ctx context.Context) (json.RawMessage, error) {
	return nil, fmt.Errorf("media reprocess is not available in SDK mode; use the REST API endpoint POST /api/v1/media/reprocess")
}

// ---------------------------------------------------------------------------
// MediaFolderBackend
// ---------------------------------------------------------------------------

type sdkMediaFolderBackend struct {
	client *modula.Client
}

func (b *sdkMediaFolderBackend) ListMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error) {
	// The SDK Resource.List doesn't support query param filtering, so we fetch
	// all folders and filter client-side. Folder counts are typically small.
	all, err := b.client.MediaFoldersData.List(ctx)
	if err != nil {
		return nil, err
	}
	if parentID == "" {
		// Return root folders (no parent)
		roots := make([]modula.MediaFolder, 0)
		for _, f := range all {
			if f.ParentID == nil {
				roots = append(roots, f)
			}
		}
		return json.Marshal(roots)
	}
	// Return children of the given parent
	pid := modula.MediaFolderID(parentID)
	children := make([]modula.MediaFolder, 0)
	for _, f := range all {
		if f.ParentID != nil && *f.ParentID == pid {
			children = append(children, f)
		}
	}
	return json.Marshal(children)
}

func (b *sdkMediaFolderBackend) GetMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.MediaFoldersData.Get(ctx, modula.MediaFolderID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaFolderBackend) CreateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateMediaFolderParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create media folder params: %w", err)
	}
	result, err := b.client.MediaFoldersData.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaFolderBackend) UpdateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateMediaFolderParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update media folder params: %w", err)
	}
	result, err := b.client.MediaFoldersData.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaFolderBackend) DeleteMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	err := b.client.MediaFoldersData.Delete(ctx, modula.MediaFolderID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "deleted"})
}

func (b *sdkMediaFolderBackend) MoveMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.MoveMediaParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal move media params: %w", err)
	}
	result, err := b.client.MediaFolders.MoveMedia(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaFolderBackend) GetMediaFolderTree(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.MediaFolders.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkMediaFolderBackend) ListMediaInFolder(ctx context.Context, folderID string, limit, offset int64) (json.RawMessage, error) {
	result, err := b.client.MediaFolders.ListMedia(ctx, modula.MediaFolderID(folderID), modula.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// AdminMediaBackend
// ---------------------------------------------------------------------------

type sdkAdminMediaBackend struct {
	client *modula.Client
}

func (b *sdkAdminMediaBackend) ListAdminMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.client.AdminMediaData.ListPaginated(ctx, modula.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminMediaBackend) GetAdminMedia(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminMediaData.Get(ctx, modula.AdminMediaID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminMediaBackend) UpdateAdminMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminMediaParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin media params: %w", err)
	}
	result, err := b.client.AdminMediaData.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminMediaBackend) DeleteAdminMedia(ctx context.Context, id string) error {
	return b.client.AdminMediaData.Delete(ctx, modula.AdminMediaID(id))
}

func (b *sdkAdminMediaBackend) UploadAdminMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error) {
	result, err := b.client.AdminMediaUpload.Upload(ctx, reader, filename, nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminMediaBackend) ListMediaDimensions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.MediaDimensions.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// AdminMediaFolderBackend
// ---------------------------------------------------------------------------

type sdkAdminMediaFolderBackend struct {
	client *modula.Client
}

func (b *sdkAdminMediaFolderBackend) ListAdminMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error) {
	// The SDK Resource.List doesn't support query param filtering, so we fetch
	// all folders and filter client-side. Folder counts are typically small.
	all, err := b.client.AdminMediaFoldersData.List(ctx)
	if err != nil {
		return nil, err
	}
	if parentID == "" {
		// Return root folders (no parent)
		roots := make([]modula.AdminMediaFolder, 0)
		for _, f := range all {
			if f.ParentID == nil {
				roots = append(roots, f)
			}
		}
		return json.Marshal(roots)
	}
	// Return children of the given parent
	pid := modula.AdminMediaFolderID(parentID)
	children := make([]modula.AdminMediaFolder, 0)
	for _, f := range all {
		if f.ParentID != nil && *f.ParentID == pid {
			children = append(children, f)
		}
	}
	return json.Marshal(children)
}

func (b *sdkAdminMediaFolderBackend) GetAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminMediaFoldersData.Get(ctx, modula.AdminMediaFolderID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminMediaFolderBackend) CreateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminMediaFolderParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin media folder params: %w", err)
	}
	result, err := b.client.AdminMediaFoldersData.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminMediaFolderBackend) UpdateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminMediaFolderParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin media folder params: %w", err)
	}
	result, err := b.client.AdminMediaFoldersData.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminMediaFolderBackend) DeleteAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	err := b.client.AdminMediaFoldersData.Delete(ctx, modula.AdminMediaFolderID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "deleted"})
}

func (b *sdkAdminMediaFolderBackend) MoveAdminMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.MoveAdminMediaParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal move admin media params: %w", err)
	}
	result, err := b.client.AdminMediaFolders.MoveMedia(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminMediaFolderBackend) AdminGetMediaFolderTree(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.AdminMediaFolders.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminMediaFolderBackend) AdminListMediaInFolder(ctx context.Context, folderID string, limit, offset int64) (json.RawMessage, error) {
	result, err := b.client.AdminMediaFolders.ListMedia(ctx, modula.AdminMediaFolderID(folderID), modula.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
