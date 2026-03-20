package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// errNoConnection is returned when a tool is invoked but no project is connected.
var errNoConnection = fmt.Errorf("no active connection — use switch_project to connect to a CMS instance")

// proxyBackends wraps a ConnectionManager and lazily delegates to SDKBackends
// built from the current client. When the project is switched, subsequent
// calls automatically use the new client.
type proxyBackends struct {
	cm *ConnectionManager
}

func (p *proxyBackends) backends() (*Backends, error) {
	client := p.cm.Client()
	if client == nil {
		return nil, errNoConnection
	}
	return NewSDKBackends(client), nil
}

// NewProxyBackends returns a Backends struct where every method delegates
// through the ConnectionManager's current client. This enables runtime
// project switching without rebuilding the MCP server.
func NewProxyBackends(cm *ConnectionManager) *Backends {
	pb := &proxyBackends{cm: cm}
	return &Backends{
		Content:           &proxyContentBackend{pb},
		AdminContent:      &proxyAdminContentBackend{pb},
		Schema:            &proxySchemaBackend{pb},
		AdminSchema:       &proxyAdminSchemaBackend{pb},
		Media:             &proxyMediaBackend{pb},
		MediaFolders:      &proxyMediaFolderBackend{pb},
		AdminMedia:        &proxyAdminMediaBackend{pb},
		AdminMediaFolders: &proxyAdminMediaFolderBackend{pb},
		Routes:            &proxyRouteBackend{pb},
		AdminRoutes:       &proxyAdminRouteBackend{pb},
		Users:             &proxyUserBackend{pb},
		RBAC:              &proxyRBACBackend{pb},
		Sessions:          &proxySessionBackend{pb},
		Tokens:            &proxyTokenBackend{pb},
		SSHKeys:           &proxySSHKeyBackend{pb},
		OAuth:             &proxyOAuthBackend{pb},
		Tables:            &proxyTableBackend{pb},
		Plugins:           &proxyPluginBackend{pb},
		Config:            &proxyConfigBackend{pb},
		Import:            &proxyImportBackend{pb},
		Deploy:            &proxyDeployBackend{pb},
		Health:            &proxyHealthBackend{pb},
	}
}

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

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

type proxySchemaBackend struct{ p *proxyBackends }

func (b *proxySchemaBackend) ListDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.ListDatatypes(ctx, full)
}
func (b *proxySchemaBackend) GetDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetDatatype(ctx, id)
}
func (b *proxySchemaBackend) CreateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.CreateDatatype(ctx, params)
}
func (b *proxySchemaBackend) UpdateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.UpdateDatatype(ctx, params)
}
func (b *proxySchemaBackend) DeleteDatatype(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Schema.DeleteDatatype(ctx, id)
}
func (b *proxySchemaBackend) ListFields(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.ListFields(ctx)
}
func (b *proxySchemaBackend) GetField(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetField(ctx, id)
}
func (b *proxySchemaBackend) CreateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.CreateField(ctx, params)
}
func (b *proxySchemaBackend) UpdateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.UpdateField(ctx, params)
}
func (b *proxySchemaBackend) DeleteField(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Schema.DeleteField(ctx, id)
}
func (b *proxySchemaBackend) GetDatatypeFull(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetDatatypeFull(ctx, id)
}
func (b *proxySchemaBackend) ListFieldTypes(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.ListFieldTypes(ctx)
}
func (b *proxySchemaBackend) GetFieldType(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetFieldType(ctx, id)
}
func (b *proxySchemaBackend) CreateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.CreateFieldType(ctx, params)
}
func (b *proxySchemaBackend) UpdateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.UpdateFieldType(ctx, params)
}
func (b *proxySchemaBackend) DeleteFieldType(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Schema.DeleteFieldType(ctx, id)
}

// ---------------------------------------------------------------------------
// AdminSchema
// ---------------------------------------------------------------------------

type proxyAdminSchemaBackend struct{ p *proxyBackends }

func (b *proxyAdminSchemaBackend) ListAdminDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.ListAdminDatatypes(ctx, full)
}
func (b *proxyAdminSchemaBackend) GetAdminDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.GetAdminDatatype(ctx, id)
}
func (b *proxyAdminSchemaBackend) CreateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.CreateAdminDatatype(ctx, params)
}
func (b *proxyAdminSchemaBackend) UpdateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.UpdateAdminDatatype(ctx, params)
}
func (b *proxyAdminSchemaBackend) DeleteAdminDatatype(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminSchema.DeleteAdminDatatype(ctx, id)
}
func (b *proxyAdminSchemaBackend) ListAdminFields(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.ListAdminFields(ctx)
}
func (b *proxyAdminSchemaBackend) GetAdminField(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.GetAdminField(ctx, id)
}
func (b *proxyAdminSchemaBackend) CreateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.CreateAdminField(ctx, params)
}
func (b *proxyAdminSchemaBackend) UpdateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.UpdateAdminField(ctx, params)
}
func (b *proxyAdminSchemaBackend) DeleteAdminField(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminSchema.DeleteAdminField(ctx, id)
}

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

// ---------------------------------------------------------------------------
// Routes
// ---------------------------------------------------------------------------

type proxyRouteBackend struct{ p *proxyBackends }

func (b *proxyRouteBackend) ListRoutes(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Routes.ListRoutes(ctx)
}
func (b *proxyRouteBackend) GetRoute(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Routes.GetRoute(ctx, id)
}
func (b *proxyRouteBackend) CreateRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Routes.CreateRoute(ctx, params)
}
func (b *proxyRouteBackend) UpdateRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Routes.UpdateRoute(ctx, params)
}
func (b *proxyRouteBackend) DeleteRoute(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Routes.DeleteRoute(ctx, id)
}

// ---------------------------------------------------------------------------
// AdminRoutes
// ---------------------------------------------------------------------------

type proxyAdminRouteBackend struct{ p *proxyBackends }

func (b *proxyAdminRouteBackend) ListAdminRoutes(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminRoutes.ListAdminRoutes(ctx)
}
func (b *proxyAdminRouteBackend) GetAdminRoute(ctx context.Context, slug string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminRoutes.GetAdminRoute(ctx, slug)
}
func (b *proxyAdminRouteBackend) CreateAdminRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminRoutes.CreateAdminRoute(ctx, params)
}
func (b *proxyAdminRouteBackend) UpdateAdminRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminRoutes.UpdateAdminRoute(ctx, params)
}
func (b *proxyAdminRouteBackend) DeleteAdminRoute(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminRoutes.DeleteAdminRoute(ctx, id)
}
func (b *proxyAdminRouteBackend) ListAdminFieldTypes(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminRoutes.ListAdminFieldTypes(ctx)
}
func (b *proxyAdminRouteBackend) GetAdminFieldType(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminRoutes.GetAdminFieldType(ctx, id)
}
func (b *proxyAdminRouteBackend) CreateAdminFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminRoutes.CreateAdminFieldType(ctx, params)
}
func (b *proxyAdminRouteBackend) UpdateAdminFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminRoutes.UpdateAdminFieldType(ctx, params)
}
func (b *proxyAdminRouteBackend) DeleteAdminFieldType(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminRoutes.DeleteAdminFieldType(ctx, id)
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

type proxyUserBackend struct{ p *proxyBackends }

func (b *proxyUserBackend) Whoami(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.Whoami(ctx)
}
func (b *proxyUserBackend) ListUsers(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.ListUsers(ctx)
}
func (b *proxyUserBackend) GetUser(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.GetUser(ctx, id)
}
func (b *proxyUserBackend) CreateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.CreateUser(ctx, params)
}
func (b *proxyUserBackend) UpdateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.UpdateUser(ctx, params)
}
func (b *proxyUserBackend) DeleteUser(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Users.DeleteUser(ctx, id)
}
func (b *proxyUserBackend) ListUsersFull(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.ListUsersFull(ctx)
}
func (b *proxyUserBackend) GetUserFull(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.GetUserFull(ctx, id)
}

// ---------------------------------------------------------------------------
// RBAC
// ---------------------------------------------------------------------------

type proxyRBACBackend struct{ p *proxyBackends }

func (b *proxyRBACBackend) ListRoles(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.ListRoles(ctx)
}
func (b *proxyRBACBackend) GetRole(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.GetRole(ctx, id)
}
func (b *proxyRBACBackend) CreateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.CreateRole(ctx, params)
}
func (b *proxyRBACBackend) UpdateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.UpdateRole(ctx, params)
}
func (b *proxyRBACBackend) DeleteRole(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.RBAC.DeleteRole(ctx, id)
}
func (b *proxyRBACBackend) ListPermissions(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.ListPermissions(ctx)
}
func (b *proxyRBACBackend) GetPermission(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.GetPermission(ctx, id)
}
func (b *proxyRBACBackend) CreatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.CreatePermission(ctx, params)
}
func (b *proxyRBACBackend) UpdatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.UpdatePermission(ctx, params)
}
func (b *proxyRBACBackend) DeletePermission(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.RBAC.DeletePermission(ctx, id)
}
func (b *proxyRBACBackend) AssignRolePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.AssignRolePermission(ctx, params)
}
func (b *proxyRBACBackend) RemoveRolePermission(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.RBAC.RemoveRolePermission(ctx, id)
}
func (b *proxyRBACBackend) ListRolePermissions(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.ListRolePermissions(ctx)
}
func (b *proxyRBACBackend) GetRolePermission(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.GetRolePermission(ctx, id)
}
func (b *proxyRBACBackend) ListRolePermissionsByRole(ctx context.Context, roleID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.ListRolePermissionsByRole(ctx, roleID)
}

// ---------------------------------------------------------------------------
// Sessions, Tokens, SSHKeys, OAuth, Tables, Plugins, Config, Import, Deploy, Health
// ---------------------------------------------------------------------------

type proxySessionBackend struct{ p *proxyBackends }

func (b *proxySessionBackend) ListSessions(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Sessions.ListSessions(ctx)
}
func (b *proxySessionBackend) GetSession(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Sessions.GetSession(ctx, id)
}
func (b *proxySessionBackend) UpdateSession(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Sessions.UpdateSession(ctx, params)
}
func (b *proxySessionBackend) DeleteSession(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Sessions.DeleteSession(ctx, id)
}

type proxyTokenBackend struct{ p *proxyBackends }

func (b *proxyTokenBackend) ListTokens(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tokens.ListTokens(ctx)
}
func (b *proxyTokenBackend) GetToken(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tokens.GetToken(ctx, id)
}
func (b *proxyTokenBackend) CreateToken(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tokens.CreateToken(ctx, params)
}
func (b *proxyTokenBackend) DeleteToken(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Tokens.DeleteToken(ctx, id)
}

type proxySSHKeyBackend struct{ p *proxyBackends }

func (b *proxySSHKeyBackend) ListSSHKeys(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.SSHKeys.ListSSHKeys(ctx)
}
func (b *proxySSHKeyBackend) CreateSSHKey(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.SSHKeys.CreateSSHKey(ctx, params)
}
func (b *proxySSHKeyBackend) DeleteSSHKey(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.SSHKeys.DeleteSSHKey(ctx, id)
}

type proxyOAuthBackend struct{ p *proxyBackends }

func (b *proxyOAuthBackend) ListUsersOAuth(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.OAuth.ListUsersOAuth(ctx)
}
func (b *proxyOAuthBackend) GetUserOAuth(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.OAuth.GetUserOAuth(ctx, id)
}
func (b *proxyOAuthBackend) CreateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.OAuth.CreateUserOAuth(ctx, params)
}
func (b *proxyOAuthBackend) UpdateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.OAuth.UpdateUserOAuth(ctx, params)
}
func (b *proxyOAuthBackend) DeleteUserOAuth(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.OAuth.DeleteUserOAuth(ctx, id)
}

type proxyTableBackend struct{ p *proxyBackends }

func (b *proxyTableBackend) ListTables(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tables.ListTables(ctx)
}
func (b *proxyTableBackend) GetTable(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tables.GetTable(ctx, id)
}
func (b *proxyTableBackend) CreateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tables.CreateTable(ctx, params)
}
func (b *proxyTableBackend) UpdateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tables.UpdateTable(ctx, params)
}
func (b *proxyTableBackend) DeleteTable(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Tables.DeleteTable(ctx, id)
}

type proxyPluginBackend struct{ p *proxyBackends }

func (b *proxyPluginBackend) ListPlugins(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.ListPlugins(ctx)
}
func (b *proxyPluginBackend) GetPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.GetPlugin(ctx, name)
}
func (b *proxyPluginBackend) ReloadPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.ReloadPlugin(ctx, name)
}
func (b *proxyPluginBackend) EnablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.EnablePlugin(ctx, name)
}
func (b *proxyPluginBackend) DisablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.DisablePlugin(ctx, name)
}
func (b *proxyPluginBackend) PluginCleanupDryRun(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.PluginCleanupDryRun(ctx)
}
func (b *proxyPluginBackend) PluginCleanupDrop(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.PluginCleanupDrop(ctx, params)
}
func (b *proxyPluginBackend) ListPluginRoutes(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.ListPluginRoutes(ctx)
}
func (b *proxyPluginBackend) ApprovePluginRoutes(ctx context.Context, params json.RawMessage) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Plugins.ApprovePluginRoutes(ctx, params)
}
func (b *proxyPluginBackend) RevokePluginRoutes(ctx context.Context, params json.RawMessage) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Plugins.RevokePluginRoutes(ctx, params)
}
func (b *proxyPluginBackend) ListPluginHooks(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.ListPluginHooks(ctx)
}
func (b *proxyPluginBackend) ApprovePluginHooks(ctx context.Context, params json.RawMessage) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Plugins.ApprovePluginHooks(ctx, params)
}
func (b *proxyPluginBackend) RevokePluginHooks(ctx context.Context, params json.RawMessage) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Plugins.RevokePluginHooks(ctx, params)
}

type proxyConfigBackend struct{ p *proxyBackends }

func (b *proxyConfigBackend) GetConfig(ctx context.Context, category string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Config.GetConfig(ctx, category)
}
func (b *proxyConfigBackend) GetConfigMeta(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Config.GetConfigMeta(ctx)
}
func (b *proxyConfigBackend) UpdateConfig(ctx context.Context, updates map[string]any) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Config.UpdateConfig(ctx, updates)
}

type proxyImportBackend struct{ p *proxyBackends }

func (b *proxyImportBackend) ImportContent(ctx context.Context, format string, data any) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Import.ImportContent(ctx, format, data)
}
func (b *proxyImportBackend) ImportBulk(ctx context.Context, format string, data any) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Import.ImportBulk(ctx, format, data)
}

type proxyDeployBackend struct{ p *proxyBackends }

func (b *proxyDeployBackend) DeployHealth(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Deploy.DeployHealth(ctx)
}
func (b *proxyDeployBackend) DeployExport(ctx context.Context, tables []string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Deploy.DeployExport(ctx, tables)
}
func (b *proxyDeployBackend) DeployImport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Deploy.DeployImport(ctx, payload)
}
func (b *proxyDeployBackend) DeployDryRun(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Deploy.DeployDryRun(ctx, payload)
}

type proxyHealthBackend struct{ p *proxyBackends }

func (b *proxyHealthBackend) Health(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Health.Health(ctx)
}
