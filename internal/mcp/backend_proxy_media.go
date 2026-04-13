package mcp

import (
	"context"
	"encoding/json"
	"io"
)

// ---------------------------------------------------------------------------
// Media
// ---------------------------------------------------------------------------

type proxyMediaBackend struct{ p *proxyBackends }

func (b *proxyMediaBackend) ListMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.ListMedia(ctx, limit, offset)
}
func (b *proxyMediaBackend) GetMedia(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.GetMedia(ctx, id)
}
func (b *proxyMediaBackend) UpdateMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.UpdateMedia(ctx, params)
}
func (b *proxyMediaBackend) DeleteMedia(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Media.DeleteMedia(ctx, id)
}
func (b *proxyMediaBackend) UploadMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.UploadMedia(ctx, reader, filename)
}
func (b *proxyMediaBackend) MediaHealth(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.MediaHealth(ctx)
}
func (b *proxyMediaBackend) MediaCleanup(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.MediaCleanup(ctx)
}
func (b *proxyMediaBackend) MediaCleanupCheck(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.MediaCleanupCheck(ctx)
}
func (b *proxyMediaBackend) ListMediaDimensions(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.ListMediaDimensions(ctx)
}
func (b *proxyMediaBackend) GetMediaDimension(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.GetMediaDimension(ctx, id)
}
func (b *proxyMediaBackend) CreateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.CreateMediaDimension(ctx, params)
}
func (b *proxyMediaBackend) UpdateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.UpdateMediaDimension(ctx, params)
}
func (b *proxyMediaBackend) DeleteMediaDimension(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Media.DeleteMediaDimension(ctx, id)
}
func (b *proxyMediaBackend) DownloadMedia(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.DownloadMedia(ctx, id)
}
func (b *proxyMediaBackend) GetMediaFull(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.GetMediaFull(ctx)
}
func (b *proxyMediaBackend) GetMediaReferences(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.GetMediaReferences(ctx, id)
}
func (b *proxyMediaBackend) ReprocessMedia(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Media.ReprocessMedia(ctx)
}

// ---------------------------------------------------------------------------
// MediaFolders
// ---------------------------------------------------------------------------

type proxyMediaFolderBackend struct{ p *proxyBackends }

func (b *proxyMediaFolderBackend) ListMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.MediaFolders.ListMediaFolders(ctx, parentID)
}
func (b *proxyMediaFolderBackend) GetMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.MediaFolders.GetMediaFolder(ctx, id)
}
func (b *proxyMediaFolderBackend) CreateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.MediaFolders.CreateMediaFolder(ctx, params)
}
func (b *proxyMediaFolderBackend) UpdateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.MediaFolders.UpdateMediaFolder(ctx, params)
}
func (b *proxyMediaFolderBackend) DeleteMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.MediaFolders.DeleteMediaFolder(ctx, id)
}
func (b *proxyMediaFolderBackend) MoveMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.MediaFolders.MoveMediaToFolder(ctx, params)
}
func (b *proxyMediaFolderBackend) GetMediaFolderTree(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.MediaFolders.GetMediaFolderTree(ctx)
}
func (b *proxyMediaFolderBackend) ListMediaInFolder(ctx context.Context, folderID string, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.MediaFolders.ListMediaInFolder(ctx, folderID, limit, offset)
}

// ---------------------------------------------------------------------------
// AdminMedia
// ---------------------------------------------------------------------------

type proxyAdminMediaBackend struct{ p *proxyBackends }

func (b *proxyAdminMediaBackend) ListAdminMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMedia.ListAdminMedia(ctx, limit, offset)
}
func (b *proxyAdminMediaBackend) GetAdminMedia(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMedia.GetAdminMedia(ctx, id)
}
func (b *proxyAdminMediaBackend) UpdateAdminMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMedia.UpdateAdminMedia(ctx, params)
}
func (b *proxyAdminMediaBackend) DeleteAdminMedia(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminMedia.DeleteAdminMedia(ctx, id)
}
func (b *proxyAdminMediaBackend) UploadAdminMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMedia.UploadAdminMedia(ctx, reader, filename)
}
func (b *proxyAdminMediaBackend) ListMediaDimensions(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMedia.ListMediaDimensions(ctx)
}

// ---------------------------------------------------------------------------
// AdminMediaFolders
// ---------------------------------------------------------------------------

type proxyAdminMediaFolderBackend struct{ p *proxyBackends }

func (b *proxyAdminMediaFolderBackend) ListAdminMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMediaFolders.ListAdminMediaFolders(ctx, parentID)
}
func (b *proxyAdminMediaFolderBackend) GetAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMediaFolders.GetAdminMediaFolder(ctx, id)
}
func (b *proxyAdminMediaFolderBackend) CreateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMediaFolders.CreateAdminMediaFolder(ctx, params)
}
func (b *proxyAdminMediaFolderBackend) UpdateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMediaFolders.UpdateAdminMediaFolder(ctx, params)
}
func (b *proxyAdminMediaFolderBackend) DeleteAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMediaFolders.DeleteAdminMediaFolder(ctx, id)
}
func (b *proxyAdminMediaFolderBackend) MoveAdminMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMediaFolders.MoveAdminMediaToFolder(ctx, params)
}
func (b *proxyAdminMediaFolderBackend) AdminGetMediaFolderTree(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMediaFolders.AdminGetMediaFolderTree(ctx)
}
func (b *proxyAdminMediaFolderBackend) AdminListMediaInFolder(ctx context.Context, folderID string, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminMediaFolders.AdminListMediaInFolder(ctx, folderID, limit, offset)
}
