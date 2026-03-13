package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// NewSDKBackends returns a Backends struct with all 19 fields populated by
// SDK adapter implementations that delegate to the given client over HTTP.
func NewSDKBackends(client *modula.Client) *Backends {
	return &Backends{
		Content:      &sdkContentBackend{client: client},
		AdminContent: &sdkAdminContentBackend{client: client},
		Schema:       &sdkSchemaBackend{client: client},
		AdminSchema:  &sdkAdminSchemaBackend{client: client},
		Media:        &sdkMediaBackend{client: client},
		Routes:       &sdkRouteBackend{client: client},
		AdminRoutes:  &sdkAdminRouteBackend{client: client},
		Users:        &sdkUserBackend{client: client},
		RBAC:         &sdkRBACBackend{client: client},
		Sessions:     &sdkSessionBackend{client: client},
		Tokens:       &sdkTokenBackend{client: client},
		SSHKeys:      &sdkSSHKeyBackend{client: client},
		OAuth:        &sdkOAuthBackend{client: client},
		Tables:       &sdkTableBackend{client: client},
		Plugins:      &sdkPluginBackend{client: client},
		Config:       &sdkConfigBackend{client: client},
		Import:       &sdkImportBackend{client: client},
		Deploy:       &sdkDeployBackend{client: client},
		Health:       &sdkHealthBackend{client: client},
	}
}

// ---------------------------------------------------------------------------
// ContentBackend
// ---------------------------------------------------------------------------

type sdkContentBackend struct {
	client *modula.Client
}

func (b *sdkContentBackend) ListContent(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.client.ContentData.ListPaginated(ctx, modula.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) GetContent(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.ContentData.Get(ctx, modula.ContentID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) CreateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateContentDataParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create content params: %w", err)
	}
	result, err := b.client.ContentData.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) UpdateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateContentDataParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update content params: %w", err)
	}
	result, err := b.client.ContentData.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) DeleteContent(ctx context.Context, id string) error {
	return b.client.ContentData.Delete(ctx, modula.ContentID(id))
}

func (b *sdkContentBackend) GetPage(ctx context.Context, slug, format, locale string) (json.RawMessage, error) {
	return b.client.Content.GetPage(ctx, slug, format, locale)
}

func (b *sdkContentBackend) GetContentTree(ctx context.Context, slug, format string) (json.RawMessage, error) {
	return b.client.AdminTree.Get(ctx, slug, format)
}

func (b *sdkContentBackend) ListContentFields(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.client.ContentFields.ListPaginated(ctx, modula.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) GetContentField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.ContentFields.Get(ctx, modula.ContentFieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) CreateContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateContentFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create content field params: %w", err)
	}
	result, err := b.client.ContentFields.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) UpdateContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateContentFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update content field params: %w", err)
	}
	result, err := b.client.ContentFields.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) DeleteContentField(ctx context.Context, id string) error {
	return b.client.ContentFields.Delete(ctx, modula.ContentFieldID(id))
}

func (b *sdkContentBackend) ReorderContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.ContentReorderRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal content reorder params: %w", err)
	}
	result, err := b.client.ContentReorder.Reorder(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) MoveContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.ContentMoveRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal content move params: %w", err)
	}
	result, err := b.client.ContentReorder.Move(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) SaveContentTree(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.TreeSaveRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal tree save params: %w", err)
	}
	result, err := b.client.ContentTree.Save(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) HealContent(ctx context.Context, dryRun bool) (json.RawMessage, error) {
	result, err := b.client.ContentHeal.Heal(ctx, dryRun)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) BatchUpdateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p any
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal batch update params: %w", err)
	}
	return b.client.ContentBatch.Update(ctx, p)
}

// ---------------------------------------------------------------------------
// AdminContentBackend
// ---------------------------------------------------------------------------

type sdkAdminContentBackend struct {
	client *modula.Client
}

func (b *sdkAdminContentBackend) ListAdminContent(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.client.AdminContentData.ListPaginated(ctx, modula.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) GetAdminContent(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminContentData.Get(ctx, modula.AdminContentID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) CreateAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminContentDataParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin content params: %w", err)
	}
	result, err := b.client.AdminContentData.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) UpdateAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminContentDataParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin content params: %w", err)
	}
	result, err := b.client.AdminContentData.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) DeleteAdminContent(ctx context.Context, id string) error {
	return b.client.AdminContentData.Delete(ctx, modula.AdminContentID(id))
}

func (b *sdkAdminContentBackend) ReorderAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.AdminContentReorderRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal admin content reorder params: %w", err)
	}
	result, err := b.client.AdminContentReorder.Reorder(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) MoveAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.AdminContentMoveRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal admin content move params: %w", err)
	}
	result, err := b.client.AdminContentReorder.Move(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) ListAdminContentFields(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.client.AdminContentFields.ListPaginated(ctx, modula.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) GetAdminContentField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminContentFields.Get(ctx, modula.AdminContentFieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) CreateAdminContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminContentFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin content field params: %w", err)
	}
	result, err := b.client.AdminContentFields.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) UpdateAdminContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminContentFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin content field params: %w", err)
	}
	result, err := b.client.AdminContentFields.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminContentBackend) DeleteAdminContentField(ctx context.Context, id string) error {
	return b.client.AdminContentFields.Delete(ctx, modula.AdminContentFieldID(id))
}

// ---------------------------------------------------------------------------
// SchemaBackend
// ---------------------------------------------------------------------------

type sdkSchemaBackend struct {
	client *modula.Client
}

func (b *sdkSchemaBackend) ListDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	if full {
		return b.client.Datatypes.RawList(ctx, url.Values{"full": {"true"}})
	}
	result, err := b.client.Datatypes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) GetDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Datatypes.Get(ctx, modula.DatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) CreateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create datatype params: %w", err)
	}
	result, err := b.client.Datatypes.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) UpdateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update datatype params: %w", err)
	}
	result, err := b.client.Datatypes.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) DeleteDatatype(ctx context.Context, id string) error {
	return b.client.Datatypes.Delete(ctx, modula.DatatypeID(id))
}

func (b *sdkSchemaBackend) ListFields(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Fields.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) GetField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Fields.Get(ctx, modula.FieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) CreateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create field params: %w", err)
	}
	result, err := b.client.Fields.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) UpdateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update field params: %w", err)
	}
	result, err := b.client.Fields.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) DeleteField(ctx context.Context, id string) error {
	return b.client.Fields.Delete(ctx, modula.FieldID(id))
}

func (b *sdkSchemaBackend) GetDatatypeFull(ctx context.Context, id string) (json.RawMessage, error) {
	return b.client.Datatypes.RawList(ctx, url.Values{"full": {"true"}, "q": {id}})
}

func (b *sdkSchemaBackend) ListFieldTypes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.FieldTypes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) GetFieldType(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.FieldTypes.Get(ctx, modula.FieldTypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) CreateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create field type params: %w", err)
	}
	result, err := b.client.FieldTypes.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) UpdateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update field type params: %w", err)
	}
	result, err := b.client.FieldTypes.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) DeleteFieldType(ctx context.Context, id string) error {
	return b.client.FieldTypes.Delete(ctx, modula.FieldTypeID(id))
}

// ---------------------------------------------------------------------------
// AdminSchemaBackend
// ---------------------------------------------------------------------------

type sdkAdminSchemaBackend struct {
	client *modula.Client
}

func (b *sdkAdminSchemaBackend) ListAdminDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	if full {
		return b.client.AdminDatatypes.RawList(ctx, url.Values{"full": {"true"}})
	}
	result, err := b.client.AdminDatatypes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) GetAdminDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminDatatypes.Get(ctx, modula.AdminDatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) CreateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin datatype params: %w", err)
	}
	result, err := b.client.AdminDatatypes.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) UpdateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin datatype params: %w", err)
	}
	result, err := b.client.AdminDatatypes.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) DeleteAdminDatatype(ctx context.Context, id string) error {
	return b.client.AdminDatatypes.Delete(ctx, modula.AdminDatatypeID(id))
}

func (b *sdkAdminSchemaBackend) ListAdminFields(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.AdminFields.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) GetAdminField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminFields.Get(ctx, modula.AdminFieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) CreateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin field params: %w", err)
	}
	result, err := b.client.AdminFields.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) UpdateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin field params: %w", err)
	}
	result, err := b.client.AdminFields.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) DeleteAdminField(ctx context.Context, id string) error {
	return b.client.AdminFields.Delete(ctx, modula.AdminFieldID(id))
}

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

// ---------------------------------------------------------------------------
// RouteBackend
// ---------------------------------------------------------------------------

type sdkRouteBackend struct {
	client *modula.Client
}

func (b *sdkRouteBackend) ListRoutes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Routes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRouteBackend) GetRoute(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Routes.Get(ctx, modula.RouteID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRouteBackend) CreateRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateRouteParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create route params: %w", err)
	}
	result, err := b.client.Routes.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRouteBackend) UpdateRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateRouteParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update route params: %w", err)
	}
	result, err := b.client.Routes.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRouteBackend) DeleteRoute(ctx context.Context, id string) error {
	return b.client.Routes.Delete(ctx, modula.RouteID(id))
}

// ---------------------------------------------------------------------------
// AdminRouteBackend
// ---------------------------------------------------------------------------

type sdkAdminRouteBackend struct {
	client *modula.Client
}

func (b *sdkAdminRouteBackend) ListAdminRoutes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.AdminRoutes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminRouteBackend) GetAdminRoute(ctx context.Context, slug string) (json.RawMessage, error) {
	result, err := b.client.AdminRoutes.Get(ctx, modula.AdminRouteID(slug))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminRouteBackend) CreateAdminRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminRouteParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin route params: %w", err)
	}
	result, err := b.client.AdminRoutes.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminRouteBackend) UpdateAdminRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminRouteParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin route params: %w", err)
	}
	result, err := b.client.AdminRoutes.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminRouteBackend) DeleteAdminRoute(ctx context.Context, id string) error {
	return b.client.AdminRoutes.Delete(ctx, modula.AdminRouteID(id))
}

func (b *sdkAdminRouteBackend) ListAdminFieldTypes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.AdminFieldTypes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminRouteBackend) GetAdminFieldType(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminFieldTypes.Get(ctx, modula.AdminFieldTypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminRouteBackend) CreateAdminFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin field type params: %w", err)
	}
	result, err := b.client.AdminFieldTypes.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminRouteBackend) UpdateAdminFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin field type params: %w", err)
	}
	result, err := b.client.AdminFieldTypes.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminRouteBackend) DeleteAdminFieldType(ctx context.Context, id string) error {
	return b.client.AdminFieldTypes.Delete(ctx, modula.AdminFieldTypeID(id))
}

// ---------------------------------------------------------------------------
// UserBackend
// ---------------------------------------------------------------------------

type sdkUserBackend struct {
	client *modula.Client
}

func (b *sdkUserBackend) Whoami(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Auth.Me(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) ListUsers(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Users.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) GetUser(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Users.Get(ctx, modula.UserID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) CreateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateUserParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create user params: %w", err)
	}
	result, err := b.client.Users.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) UpdateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateUserParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update user params: %w", err)
	}
	result, err := b.client.Users.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) DeleteUser(ctx context.Context, id string) error {
	return b.client.Users.Delete(ctx, modula.UserID(id))
}

func (b *sdkUserBackend) ListUsersFull(ctx context.Context) (json.RawMessage, error) {
	return b.client.UsersFull.List(ctx)
}

func (b *sdkUserBackend) GetUserFull(ctx context.Context, id string) (json.RawMessage, error) {
	return b.client.UsersFull.Get(ctx, modula.UserID(id))
}

// ---------------------------------------------------------------------------
// RBACBackend
// ---------------------------------------------------------------------------

type sdkRBACBackend struct {
	client *modula.Client
}

func (b *sdkRBACBackend) ListRoles(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Roles.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) GetRole(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Roles.Get(ctx, modula.RoleID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) CreateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateRoleParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create role params: %w", err)
	}
	result, err := b.client.Roles.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) UpdateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateRoleParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update role params: %w", err)
	}
	result, err := b.client.Roles.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) DeleteRole(ctx context.Context, id string) error {
	return b.client.Roles.Delete(ctx, modula.RoleID(id))
}

func (b *sdkRBACBackend) ListPermissions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Permissions.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) GetPermission(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Permissions.Get(ctx, modula.PermissionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) CreatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreatePermissionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create permission params: %w", err)
	}
	result, err := b.client.Permissions.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) UpdatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdatePermissionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update permission params: %w", err)
	}
	result, err := b.client.Permissions.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) DeletePermission(ctx context.Context, id string) error {
	return b.client.Permissions.Delete(ctx, modula.PermissionID(id))
}

func (b *sdkRBACBackend) AssignRolePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateRolePermissionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal assign role permission params: %w", err)
	}
	result, err := b.client.RolePermissions.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) RemoveRolePermission(ctx context.Context, id string) error {
	return b.client.RolePermissions.Delete(ctx, modula.RolePermissionID(id))
}

func (b *sdkRBACBackend) ListRolePermissions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.RolePermissions.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) GetRolePermission(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.RolePermissions.Get(ctx, modula.RolePermissionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) ListRolePermissionsByRole(ctx context.Context, roleID string) (json.RawMessage, error) {
	result, err := b.client.RolePermissions.ListByRole(ctx, modula.RoleID(roleID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// SessionBackend
// ---------------------------------------------------------------------------

type sdkSessionBackend struct {
	client *modula.Client
}

func (b *sdkSessionBackend) ListSessions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Sessions.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSessionBackend) GetSession(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Sessions.Get(ctx, modula.SessionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSessionBackend) UpdateSession(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateSessionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update session params: %w", err)
	}
	result, err := b.client.Sessions.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSessionBackend) DeleteSession(ctx context.Context, id string) error {
	return b.client.Sessions.Remove(ctx, modula.SessionID(id))
}

// ---------------------------------------------------------------------------
// TokenBackend
// ---------------------------------------------------------------------------

type sdkTokenBackend struct {
	client *modula.Client
}

func (b *sdkTokenBackend) ListTokens(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Tokens.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTokenBackend) GetToken(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Tokens.Get(ctx, modula.TokenID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTokenBackend) CreateToken(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateTokenParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create token params: %w", err)
	}
	result, err := b.client.Tokens.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTokenBackend) DeleteToken(ctx context.Context, id string) error {
	return b.client.Tokens.Delete(ctx, modula.TokenID(id))
}

// ---------------------------------------------------------------------------
// SSHKeyBackend
// ---------------------------------------------------------------------------

type sdkSSHKeyBackend struct {
	client *modula.Client
}

func (b *sdkSSHKeyBackend) ListSSHKeys(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.SSHKeys.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSSHKeyBackend) CreateSSHKey(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateSSHKeyParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create ssh key params: %w", err)
	}
	result, err := b.client.SSHKeys.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSSHKeyBackend) DeleteSSHKey(ctx context.Context, id string) error {
	return b.client.SSHKeys.Delete(ctx, modula.UserSshKeyID(id))
}

// ---------------------------------------------------------------------------
// OAuthBackend
// ---------------------------------------------------------------------------

type sdkOAuthBackend struct {
	client *modula.Client
}

func (b *sdkOAuthBackend) ListUsersOAuth(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.UsersOauth.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkOAuthBackend) GetUserOAuth(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.UsersOauth.Get(ctx, modula.UserOauthID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkOAuthBackend) CreateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateUserOauthParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create user oauth params: %w", err)
	}
	result, err := b.client.UsersOauth.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkOAuthBackend) UpdateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateUserOauthParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update user oauth params: %w", err)
	}
	result, err := b.client.UsersOauth.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkOAuthBackend) DeleteUserOAuth(ctx context.Context, id string) error {
	return b.client.UsersOauth.Delete(ctx, modula.UserOauthID(id))
}

// ---------------------------------------------------------------------------
// TableBackend
// ---------------------------------------------------------------------------

type sdkTableBackend struct {
	client *modula.Client
}

func (b *sdkTableBackend) ListTables(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Tables.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTableBackend) GetTable(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Tables.Get(ctx, modula.TableID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTableBackend) CreateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateTableParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create table params: %w", err)
	}
	result, err := b.client.Tables.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTableBackend) UpdateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateTableParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update table params: %w", err)
	}
	result, err := b.client.Tables.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTableBackend) DeleteTable(ctx context.Context, id string) error {
	return b.client.Tables.Delete(ctx, modula.TableID(id))
}

// ---------------------------------------------------------------------------
// PluginBackend
// ---------------------------------------------------------------------------

type sdkPluginBackend struct {
	client *modula.Client
}

func (b *sdkPluginBackend) ListPlugins(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Plugins.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) GetPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.client.Plugins.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) ReloadPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.client.Plugins.Reload(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) EnablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.client.Plugins.Enable(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) DisablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.client.Plugins.Disable(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) PluginCleanupDryRun(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Plugins.CleanupDryRun(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) PluginCleanupDrop(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CleanupDropParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal cleanup drop params: %w", err)
	}
	result, err := b.client.Plugins.CleanupDrop(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) ListPluginRoutes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.PluginRoutes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) ApprovePluginRoutes(ctx context.Context, params json.RawMessage) error {
	var items []modula.RouteApprovalItem
	if err := json.Unmarshal(params, &items); err != nil {
		return fmt.Errorf("unmarshal route approval items: %w", err)
	}
	return b.client.PluginRoutes.Approve(ctx, items)
}

func (b *sdkPluginBackend) RevokePluginRoutes(ctx context.Context, params json.RawMessage) error {
	var items []modula.RouteApprovalItem
	if err := json.Unmarshal(params, &items); err != nil {
		return fmt.Errorf("unmarshal route revocation items: %w", err)
	}
	return b.client.PluginRoutes.Revoke(ctx, items)
}

func (b *sdkPluginBackend) ListPluginHooks(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.PluginHooks.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) ApprovePluginHooks(ctx context.Context, params json.RawMessage) error {
	var items []modula.HookApprovalItem
	if err := json.Unmarshal(params, &items); err != nil {
		return fmt.Errorf("unmarshal hook approval items: %w", err)
	}
	return b.client.PluginHooks.Approve(ctx, items)
}

func (b *sdkPluginBackend) RevokePluginHooks(ctx context.Context, params json.RawMessage) error {
	var items []modula.HookApprovalItem
	if err := json.Unmarshal(params, &items); err != nil {
		return fmt.Errorf("unmarshal hook revocation items: %w", err)
	}
	return b.client.PluginHooks.Revoke(ctx, items)
}

// ---------------------------------------------------------------------------
// ConfigBackend
// ---------------------------------------------------------------------------

type sdkConfigBackend struct {
	client *modula.Client
}

func (b *sdkConfigBackend) GetConfig(ctx context.Context, category string) (json.RawMessage, error) {
	result, err := b.client.Config.Get(ctx, category)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkConfigBackend) GetConfigMeta(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Config.Meta(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkConfigBackend) UpdateConfig(ctx context.Context, updates map[string]any) (json.RawMessage, error) {
	result, err := b.client.Config.Update(ctx, updates)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// ImportBackend
// ---------------------------------------------------------------------------

type sdkImportBackend struct {
	client *modula.Client
}

func (b *sdkImportBackend) ImportContent(ctx context.Context, format string, data any) (json.RawMessage, error) {
	switch format {
	case "contentful":
		return b.client.Import.Contentful(ctx, data)
	case "sanity":
		return b.client.Import.Sanity(ctx, data)
	case "strapi":
		return b.client.Import.Strapi(ctx, data)
	case "wordpress":
		return b.client.Import.WordPress(ctx, data)
	case "clean":
		return b.client.Import.Clean(ctx, data)
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}
}

func (b *sdkImportBackend) ImportBulk(ctx context.Context, format string, data any) (json.RawMessage, error) {
	return b.client.Import.Bulk(ctx, format, data)
}

// ---------------------------------------------------------------------------
// DeployBackend
// ---------------------------------------------------------------------------

type sdkDeployBackend struct {
	client *modula.Client
}

func (b *sdkDeployBackend) DeployHealth(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Deploy.Health(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkDeployBackend) DeployExport(ctx context.Context, tables []string) (json.RawMessage, error) {
	return b.client.Deploy.Export(ctx, tables)
}

func (b *sdkDeployBackend) DeployImport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	result, err := b.client.Deploy.Import(ctx, payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkDeployBackend) DeployDryRun(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	result, err := b.client.Deploy.DryRunImport(ctx, payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// HealthBackend
// ---------------------------------------------------------------------------

type sdkHealthBackend struct {
	client *modula.Client
}

func (b *sdkHealthBackend) Health(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Health.Check(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
