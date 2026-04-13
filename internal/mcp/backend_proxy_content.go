package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// Content
// ---------------------------------------------------------------------------

type proxyContentBackend struct{ p *proxyBackends }

func (b *proxyContentBackend) ListContent(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.ListContent(ctx, limit, offset)
}
func (b *proxyContentBackend) GetContent(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.GetContent(ctx, id)
}
func (b *proxyContentBackend) CreateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.CreateContent(ctx, params)
}
func (b *proxyContentBackend) UpdateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.UpdateContent(ctx, params)
}
func (b *proxyContentBackend) DeleteContent(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Content.DeleteContent(ctx, id)
}
func (b *proxyContentBackend) GetPage(ctx context.Context, slug, format, locale string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.GetPage(ctx, slug, format, locale)
}
func (b *proxyContentBackend) GetContentTree(ctx context.Context, slug, format string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.GetContentTree(ctx, slug, format)
}
func (b *proxyContentBackend) ListContentFields(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.ListContentFields(ctx, limit, offset)
}
func (b *proxyContentBackend) GetContentField(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.GetContentField(ctx, id)
}
func (b *proxyContentBackend) CreateContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.CreateContentField(ctx, params)
}
func (b *proxyContentBackend) UpdateContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.UpdateContentField(ctx, params)
}
func (b *proxyContentBackend) DeleteContentField(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Content.DeleteContentField(ctx, id)
}
func (b *proxyContentBackend) ReorderContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.ReorderContent(ctx, params)
}
func (b *proxyContentBackend) MoveContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.MoveContent(ctx, params)
}
func (b *proxyContentBackend) SaveContentTree(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.SaveContentTree(ctx, params)
}
func (b *proxyContentBackend) HealContent(ctx context.Context, dryRun bool) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.HealContent(ctx, dryRun)
}
func (b *proxyContentBackend) BatchUpdateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.BatchUpdateContent(ctx, params)
}
func (b *proxyContentBackend) QueryContent(ctx context.Context, datatype string, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.QueryContent(ctx, datatype, params)
}
func (b *proxyContentBackend) GetGlobals(ctx context.Context, format string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.GetGlobals(ctx, format)
}
func (b *proxyContentBackend) GetContentFull(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.GetContentFull(ctx, id)
}
func (b *proxyContentBackend) GetContentByRoute(ctx context.Context, routeID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.GetContentByRoute(ctx, routeID)
}
func (b *proxyContentBackend) CreateContentComposite(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Content.CreateContentComposite(ctx, params)
}

// ---------------------------------------------------------------------------
// AdminContent
// ---------------------------------------------------------------------------

type proxyAdminContentBackend struct{ p *proxyBackends }

func (b *proxyAdminContentBackend) ListAdminContent(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.ListAdminContent(ctx, limit, offset)
}
func (b *proxyAdminContentBackend) GetAdminContent(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.GetAdminContent(ctx, id)
}
func (b *proxyAdminContentBackend) CreateAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.CreateAdminContent(ctx, params)
}
func (b *proxyAdminContentBackend) UpdateAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.UpdateAdminContent(ctx, params)
}
func (b *proxyAdminContentBackend) DeleteAdminContent(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminContent.DeleteAdminContent(ctx, id)
}
func (b *proxyAdminContentBackend) ReorderAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.ReorderAdminContent(ctx, params)
}
func (b *proxyAdminContentBackend) MoveAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.MoveAdminContent(ctx, params)
}
func (b *proxyAdminContentBackend) ListAdminContentFields(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.ListAdminContentFields(ctx, limit, offset)
}
func (b *proxyAdminContentBackend) GetAdminContentField(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.GetAdminContentField(ctx, id)
}
func (b *proxyAdminContentBackend) CreateAdminContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.CreateAdminContentField(ctx, params)
}
func (b *proxyAdminContentBackend) UpdateAdminContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.UpdateAdminContentField(ctx, params)
}
func (b *proxyAdminContentBackend) DeleteAdminContentField(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminContent.DeleteAdminContentField(ctx, id)
}
func (b *proxyAdminContentBackend) AdminGetContentFull(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.AdminGetContentFull(ctx, limit, offset)
}
func (b *proxyAdminContentBackend) GetAdminTree(ctx context.Context, slug string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminContent.GetAdminTree(ctx, slug)
}

// ---------------------------------------------------------------------------
// Versions
// ---------------------------------------------------------------------------

type proxyVersionBackend struct{ p *proxyBackends }

func (b *proxyVersionBackend) ListVersions(ctx context.Context, contentID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Versions.ListVersions(ctx, contentID)
}
func (b *proxyVersionBackend) GetVersion(ctx context.Context, versionID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Versions.GetVersion(ctx, versionID)
}
func (b *proxyVersionBackend) CreateVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Versions.CreateVersion(ctx, params)
}
func (b *proxyVersionBackend) DeleteVersion(ctx context.Context, versionID string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Versions.DeleteVersion(ctx, versionID)
}
func (b *proxyVersionBackend) RestoreVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Versions.RestoreVersion(ctx, params)
}
func (b *proxyVersionBackend) AdminListVersions(ctx context.Context, contentID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Versions.AdminListVersions(ctx, contentID)
}
func (b *proxyVersionBackend) AdminGetVersion(ctx context.Context, versionID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Versions.AdminGetVersion(ctx, versionID)
}
func (b *proxyVersionBackend) AdminCreateVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Versions.AdminCreateVersion(ctx, params)
}
func (b *proxyVersionBackend) AdminDeleteVersion(ctx context.Context, versionID string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Versions.AdminDeleteVersion(ctx, versionID)
}
func (b *proxyVersionBackend) AdminRestoreVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Versions.AdminRestoreVersion(ctx, params)
}
