package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/tree/ops"
)

// NewServiceBackends creates a Backends wired to a service.Registry for
// in-process (direct mode) operation. Every domain adapter wraps the
// corresponding service with JSON marshaling.
func NewServiceBackends(svc *service.Registry, ac audited.AuditContext) *Backends {
	return &Backends{
		Content:           &svcContentBackend{svc: svc, ac: ac},
		AdminContent:      &svcAdminContentBackend{svc: svc, ac: ac},
		Schema:            &svcSchemaBackend{svc: svc, ac: ac},
		AdminSchema:       &svcAdminSchemaBackend{svc: svc, ac: ac},
		Media:             &svcMediaBackend{svc: svc, ac: ac},
		MediaFolders:      &svcMediaFolderBackend{svc: svc, ac: ac},
		AdminMedia:        &svcAdminMediaBackend{svc: svc, ac: ac},
		AdminMediaFolders: &svcAdminMediaFolderBackend{svc: svc, ac: ac},
		Routes:            &svcRouteBackend{svc: svc, ac: ac},
		AdminRoutes:       &svcAdminRouteBackend{svc: svc, ac: ac},
		Users:             &svcUserBackend{svc: svc, ac: ac},
		RBAC:              &svcRBACBackend{svc: svc, ac: ac},
		Sessions:          &svcSessionBackend{svc: svc, ac: ac},
		Tokens:            &svcTokenBackend{svc: svc, ac: ac},
		SSHKeys:           &svcSSHKeyBackend{svc: svc, ac: ac},
		OAuth:             &svcOAuthBackend{svc: svc, ac: ac},
		Tables:            &svcTableBackend{svc: svc, ac: ac},
		Plugins:           &svcPluginBackend{svc: svc},
		Config:            &svcConfigBackend{svc: svc},
		Import:            &svcImportBackend{svc: svc, ac: ac},
		Deploy:            &svcDeployBackend{svc: svc},
		Health:            &svcHealthBackend{svc: svc},
	}
}

// ---------------------------------------------------------------------------
// ContentBackend
// ---------------------------------------------------------------------------

type svcContentBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcContentBackend) ListContent(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.svc.Content.ListPaginated(ctx, db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) GetContent(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Content.Get(ctx, types.ContentID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) CreateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateContentDataParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create content params: %w", err)
	}
	result, err := b.svc.Content.Create(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) UpdateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		db.UpdateContentDataParams
		Revision int64 `json:"revision"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal update content params: %w", err)
	}
	result, err := b.svc.Content.Update(ctx, b.ac, input.UpdateContentDataParams, input.Revision)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) DeleteContent(ctx context.Context, id string) error {
	_, err := b.svc.Content.Delete(ctx, b.ac, types.ContentID(id), false)
	return err
}

func (b *svcContentBackend) GetPage(ctx context.Context, slug, format, locale string) (json.RawMessage, error) {
	// Content delivery (GetPage) is handled at the HTTP handler layer with
	// publishing snapshots, transform pipelines, and locale resolution.
	// The service layer does not expose a single GetPage method.
	return nil, fmt.Errorf("GetPage is not supported in direct mode; use the REST API")
}

func (b *svcContentBackend) GetContentTree(ctx context.Context, routeID, format string) (json.RawMessage, error) {
	nullableRouteID := types.NullableRouteID{ID: types.RouteID(routeID), Valid: routeID != ""}
	tree, err := b.svc.Content.GetTree(ctx, nullableRouteID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(tree)
}

func (b *svcContentBackend) ListContentFields(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.svc.Content.ListFieldsPaginated(ctx, db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) GetContentField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Content.GetField(ctx, types.ContentFieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) CreateContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateContentFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create content field params: %w", err)
	}
	result, err := b.svc.Content.CreateField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) UpdateContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateContentFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update content field params: %w", err)
	}
	result, err := b.svc.Content.UpdateField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) DeleteContentField(ctx context.Context, id string) error {
	return b.svc.Content.DeleteField(ctx, b.ac, types.ContentFieldID(id))
}

func (b *svcContentBackend) ReorderContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		ParentID   *string  `json:"parent_id"`
		OrderedIDs []string `json:"ordered_ids"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal reorder params: %w", err)
	}

	parentID := ops.EmptyID[types.ContentID]()
	if input.ParentID != nil && *input.ParentID != "" {
		parentID = ops.NullID(types.ContentID(*input.ParentID))
	}

	orderedIDs := make([]types.ContentID, len(input.OrderedIDs))
	for i, id := range input.OrderedIDs {
		orderedIDs[i] = types.ContentID(id)
	}

	updated, err := b.svc.Content.Reorder(ctx, b.ac, parentID, orderedIDs)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int{"updated": updated})
}

func (b *svcContentBackend) MoveContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		NodeID      string  `json:"node_id"`
		NewParentID *string `json:"new_parent_id"`
		Position    int     `json:"position"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal move params: %w", err)
	}

	newParentID := ops.EmptyID[types.ContentID]()
	if input.NewParentID != nil && *input.NewParentID != "" {
		newParentID = ops.NullID(types.ContentID(*input.NewParentID))
	}

	result, err := b.svc.Content.Move(ctx, b.ac, ops.MoveParams[types.ContentID]{
		NodeID:      types.ContentID(input.NodeID),
		NewParentID: newParentID,
		Position:    input.Position,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) SaveContentTree(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	// SaveContentTree is a complex bulk tree operation that has no single
	// service method. The handler layer orchestrates multiple operations.
	return nil, fmt.Errorf("SaveContentTree is not supported in direct mode; use the REST API")
}

func (b *svcContentBackend) HealContent(ctx context.Context, dryRun bool) (json.RawMessage, error) {
	result, err := b.svc.Content.Heal(ctx, b.ac, dryRun)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) BatchUpdateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.BatchUpdateParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal batch update params: %w", err)
	}
	result, err := b.svc.Content.BatchUpdate(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// AdminContentBackend
// ---------------------------------------------------------------------------

type svcAdminContentBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcAdminContentBackend) ListAdminContent(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.svc.AdminContent.ListPaginated(ctx, db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminContentBackend) GetAdminContent(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.AdminContent.Get(ctx, types.AdminContentID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminContentBackend) CreateAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateAdminContentDataParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin content params: %w", err)
	}
	result, err := b.svc.AdminContent.Create(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminContentBackend) UpdateAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		db.UpdateAdminContentDataParams
		Revision int64 `json:"revision"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal update admin content params: %w", err)
	}
	result, err := b.svc.AdminContent.Update(ctx, b.ac, input.UpdateAdminContentDataParams, input.Revision)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminContentBackend) DeleteAdminContent(ctx context.Context, id string) error {
	_, err := b.svc.AdminContent.Delete(ctx, b.ac, types.AdminContentID(id), false)
	return err
}

func (b *svcAdminContentBackend) ReorderAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		ParentID   *string  `json:"parent_id"`
		OrderedIDs []string `json:"ordered_ids"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal admin reorder params: %w", err)
	}

	parentID := ops.EmptyID[types.AdminContentID]()
	if input.ParentID != nil && *input.ParentID != "" {
		parentID = ops.NullID(types.AdminContentID(*input.ParentID))
	}

	orderedIDs := make([]types.AdminContentID, len(input.OrderedIDs))
	for i, id := range input.OrderedIDs {
		orderedIDs[i] = types.AdminContentID(id)
	}

	updated, err := b.svc.AdminContent.Reorder(ctx, b.ac, parentID, orderedIDs)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int{"updated": updated})
}

func (b *svcAdminContentBackend) MoveAdminContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		NodeID      string  `json:"node_id"`
		NewParentID *string `json:"new_parent_id"`
		Position    int     `json:"position"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal admin move params: %w", err)
	}

	newParentID := ops.EmptyID[types.AdminContentID]()
	if input.NewParentID != nil && *input.NewParentID != "" {
		newParentID = ops.NullID(types.AdminContentID(*input.NewParentID))
	}

	result, err := b.svc.AdminContent.Move(ctx, b.ac, ops.MoveParams[types.AdminContentID]{
		NodeID:      types.AdminContentID(input.NodeID),
		NewParentID: newParentID,
		Position:    input.Position,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminContentBackend) ListAdminContentFields(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	result, err := b.svc.AdminContent.ListFieldsPaginated(ctx, db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminContentBackend) GetAdminContentField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.AdminContent.GetField(ctx, types.AdminContentFieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminContentBackend) CreateAdminContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateAdminContentFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin content field params: %w", err)
	}
	result, err := b.svc.AdminContent.CreateField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminContentBackend) UpdateAdminContentField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateAdminContentFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin content field params: %w", err)
	}
	result, err := b.svc.AdminContent.UpdateField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminContentBackend) DeleteAdminContentField(ctx context.Context, id string) error {
	return b.svc.AdminContent.DeleteField(ctx, b.ac, types.AdminContentFieldID(id))
}

// ---------------------------------------------------------------------------
// SchemaBackend
// ---------------------------------------------------------------------------

type svcSchemaBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcSchemaBackend) ListDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	if full {
		result, err := b.svc.Schema.ListDatatypesFull(ctx)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)
	}
	result, err := b.svc.Schema.ListDatatypes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) GetDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetDatatype(ctx, types.DatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) CreateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create datatype params: %w", err)
	}
	result, err := b.svc.Schema.CreateDatatype(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) UpdateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update datatype params: %w", err)
	}
	result, err := b.svc.Schema.UpdateDatatype(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) DeleteDatatype(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteDatatype(ctx, b.ac, types.DatatypeID(id))
}

func (b *svcSchemaBackend) ListFields(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Schema.ListFields(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) GetField(ctx context.Context, id string) (json.RawMessage, error) {
	// Service GetField requires roleID and isAdmin for access control.
	// In direct mode (MCP), assume admin access.
	result, err := b.svc.Schema.GetField(ctx, types.FieldID(id), "", true)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) CreateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create field params: %w", err)
	}
	result, err := b.svc.Schema.CreateField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) UpdateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update field params: %w", err)
	}
	result, err := b.svc.Schema.UpdateField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) DeleteField(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteField(ctx, b.ac, types.FieldID(id))
}

func (b *svcSchemaBackend) GetDatatypeFull(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetDatatypeFull(ctx, types.DatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) ListFieldTypes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Schema.ListFieldTypes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) GetFieldType(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetFieldType(ctx, types.FieldTypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) CreateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create field type params: %w", err)
	}
	result, err := b.svc.Schema.CreateFieldType(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) UpdateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update field type params: %w", err)
	}
	result, err := b.svc.Schema.UpdateFieldType(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) DeleteFieldType(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteFieldType(ctx, b.ac, types.FieldTypeID(id))
}

// ---------------------------------------------------------------------------
// AdminSchemaBackend
// ---------------------------------------------------------------------------

type svcAdminSchemaBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcAdminSchemaBackend) ListAdminDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	// AdminDatatypes do not have a ListDatatypesFull equivalent.
	// The full flag is ignored; return all admin datatypes.
	result, err := b.svc.Schema.ListAdminDatatypes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) GetAdminDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetAdminDatatype(ctx, types.AdminDatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) CreateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateAdminDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin datatype params: %w", err)
	}
	result, err := b.svc.Schema.CreateAdminDatatype(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) UpdateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateAdminDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin datatype params: %w", err)
	}
	result, err := b.svc.Schema.UpdateAdminDatatype(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) DeleteAdminDatatype(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteAdminDatatype(ctx, b.ac, types.AdminDatatypeID(id))
}

func (b *svcAdminSchemaBackend) ListAdminFields(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Schema.ListAdminFields(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) GetAdminField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetAdminField(ctx, types.AdminFieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) CreateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateAdminFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin field params: %w", err)
	}
	result, err := b.svc.Schema.CreateAdminField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) UpdateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateAdminFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin field params: %w", err)
	}
	result, err := b.svc.Schema.UpdateAdminField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) DeleteAdminField(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteAdminField(ctx, b.ac, types.AdminFieldID(id))
}

// ---------------------------------------------------------------------------
// MediaBackend
// ---------------------------------------------------------------------------

type svcMediaBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcMediaBackend) ListMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	items, total, err := b.svc.Media.ListMediaPaginated(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	resp := db.PaginatedResponse[mcpMediaResponse]{
		Limit:  limit,
		Offset: offset,
	}
	if items != nil {
		resp.Data = toMCPMediaList(*items)
	}
	if total != nil {
		resp.Total = *total
	}
	return json.Marshal(resp)
}

func (b *svcMediaBackend) GetMedia(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Media.GetMedia(ctx, types.MediaID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(toMCPMediaResponse(*result))
}

func (b *svcMediaBackend) UpdateMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdateMediaMetadataParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update media params: %w", err)
	}
	result, err := b.svc.Media.UpdateMediaMetadata(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) DeleteMedia(ctx context.Context, id string) error {
	return b.svc.Media.DeleteMedia(ctx, b.ac, types.MediaID(id))
}

func (b *svcMediaBackend) UploadMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error) {
	// The MediaService.Upload requires multipart.File and *multipart.FileHeader,
	// which are HTTP-specific types. In direct mode we cannot easily construct
	// these from an io.Reader. Return an unsupported error directing callers
	// to use the REST API for media upload.
	return nil, fmt.Errorf("media upload is not supported in direct mode; use the REST API")
}

func (b *svcMediaBackend) MediaHealth(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Media.MediaHealth(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) MediaCleanup(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Media.MediaCleanup(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) ListMediaDimensions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Media.ListMediaDimensions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) GetMediaDimension(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Media.GetMediaDimension(ctx, id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) CreateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateMediaDimensionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create media dimension params: %w", err)
	}
	result, err := b.svc.Media.CreateMediaDimension(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) UpdateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateMediaDimensionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update media dimension params: %w", err)
	}
	result, err := b.svc.Media.UpdateMediaDimension(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) DeleteMediaDimension(ctx context.Context, id string) error {
	return b.svc.Media.DeleteMediaDimension(ctx, b.ac, id)
}

// ---------------------------------------------------------------------------
// MediaFolderBackend
// ---------------------------------------------------------------------------

type svcMediaFolderBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcMediaFolderBackend) ListMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error) {
	d := b.svc.Driver()
	if parentID != "" {
		pid := types.MediaFolderID(parentID)
		if err := pid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		folders, err := d.ListMediaFoldersByParent(pid)
		if err != nil {
			return nil, err
		}
		return json.Marshal(folders)
	}
	folders, err := d.ListMediaFoldersAtRoot()
	if err != nil {
		return nil, err
	}
	return json.Marshal(folders)
}

func (b *svcMediaFolderBackend) GetMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()
	fid := types.MediaFolderID(id)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	folder, err := d.GetMediaFolder(fid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "media_folder", ID: id}
	}
	return json.Marshal(folder)
}

func (b *svcMediaFolderBackend) CreateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create media folder params: %w", err)
	}
	if p.Name == "" {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "name", Message: "required"}}}
	}

	var parentID types.NullableMediaFolderID
	if p.ParentID != nil && *p.ParentID != "" {
		pid := types.MediaFolderID(*p.ParentID)
		if err := pid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		if _, err := d.GetMediaFolder(pid); err != nil {
			return nil, &service.NotFoundError{Resource: "parent_folder", ID: *p.ParentID}
		}
		parentID = types.NullableMediaFolderID{ID: pid, Valid: true}

		breadcrumb, err := d.GetMediaFolderBreadcrumb(pid)
		if err != nil {
			return nil, fmt.Errorf("check folder depth: %w", err)
		}
		if len(breadcrumb)+1 > 10 {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: "creating this folder would exceed maximum folder depth of 10"}}}
		}
	}

	if err := d.ValidateMediaFolderName(p.Name, parentID); err != nil {
		return nil, &service.ConflictError{Resource: "media_folder", Detail: err.Error()}
	}

	now := types.NewTimestamp(time.Now().UTC())
	folder, err := d.CreateMediaFolder(ctx, b.ac, db.CreateMediaFolderParams{
		Name:         p.Name,
		ParentID:     parentID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(folder)
}

func (b *svcMediaFolderBackend) UpdateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		FolderID string  `json:"folder_id"`
		Name     *string `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update media folder params: %w", err)
	}

	fid := types.MediaFolderID(p.FolderID)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	existing, err := d.GetMediaFolder(fid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "media_folder", ID: p.FolderID}
	}

	name := existing.Name
	parentID := existing.ParentID

	if p.Name != nil {
		if *p.Name == "" {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "name", Message: "cannot be empty"}}}
		}
		name = *p.Name
	}

	parentChanged := false
	if p.ParentID != nil {
		parentChanged = true
		if *p.ParentID == "" {
			parentID = types.NullableMediaFolderID{}
		} else {
			pid := types.MediaFolderID(*p.ParentID)
			if err := pid.Validate(); err != nil {
				return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
			}
			parentID = types.NullableMediaFolderID{ID: pid, Valid: true}
		}
	}

	if parentChanged {
		if err := d.ValidateMediaFolderMove(fid, parentID); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: err.Error()}}}
		}
	}

	nameChanged := p.Name != nil && name != existing.Name
	if nameChanged || parentChanged {
		if err := d.ValidateMediaFolderName(name, parentID); err != nil {
			return nil, &service.ConflictError{Resource: "media_folder", Detail: err.Error()}
		}
	}

	_, err = d.UpdateMediaFolder(ctx, b.ac, db.UpdateMediaFolderParams{
		FolderID:     fid,
		Name:         name,
		ParentID:     parentID,
		DateModified: types.NewTimestamp(time.Now().UTC()),
	})
	if err != nil {
		return nil, err
	}

	updated, err := d.GetMediaFolder(fid)
	if err != nil {
		return nil, err
	}
	return json.Marshal(updated)
}

func (b *svcMediaFolderBackend) DeleteMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()

	fid := types.MediaFolderID(id)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	if _, err := d.GetMediaFolder(fid); err != nil {
		return nil, &service.NotFoundError{Resource: "media_folder", ID: id}
	}

	children, err := d.ListMediaFoldersByParent(fid)
	if err != nil {
		return nil, err
	}
	childCount := 0
	if children != nil {
		childCount = len(*children)
	}

	folderNullable := types.NullableMediaFolderID{ID: fid, Valid: true}
	mediaCount, err := d.CountMediaByFolder(folderNullable)
	if err != nil {
		return nil, err
	}

	mc := int64(0)
	if mediaCount != nil {
		mc = *mediaCount
	}

	if childCount > 0 || mc > 0 {
		return json.Marshal(map[string]any{
			"error":         fmt.Sprintf("cannot delete folder: contains %d child folder(s) and %d media item(s)", childCount, mc),
			"child_folders": childCount,
			"media_items":   mc,
		})
	}

	if err := d.DeleteMediaFolder(ctx, b.ac, fid); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "deleted"})
}

func (b *svcMediaFolderBackend) MoveMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		MediaIDs []string `json:"media_ids"`
		FolderID *string  `json:"folder_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal move media params: %w", err)
	}

	if len(p.MediaIDs) == 0 {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "media_ids", Message: "required and cannot be empty"}}}
	}
	if len(p.MediaIDs) > 100 {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "media_ids", Message: "batch size cannot exceed 100 items"}}}
	}

	var folderID types.NullableMediaFolderID
	if p.FolderID != nil && *p.FolderID != "" {
		fid := types.MediaFolderID(*p.FolderID)
		if err := fid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		if _, err := d.GetMediaFolder(fid); err != nil {
			return nil, &service.NotFoundError{Resource: "media_folder", ID: *p.FolderID}
		}
		folderID = types.NullableMediaFolderID{ID: fid, Valid: true}
	}

	now := types.NewTimestamp(time.Now().UTC())
	moved := 0
	for _, idStr := range p.MediaIDs {
		mid := types.MediaID(idStr)
		if err := mid.Validate(); err != nil {
			continue
		}
		err := d.MoveMediaToFolder(ctx, b.ac, db.MoveMediaToFolderParams{
			FolderID:     folderID,
			DateModified: now,
			MediaID:      mid,
		})
		if err != nil {
			continue
		}
		moved++
	}

	return json.Marshal(map[string]int{"moved": moved})
}

// ---------------------------------------------------------------------------
// RouteBackend
// ---------------------------------------------------------------------------

type svcRouteBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcRouteBackend) ListRoutes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Routes.ListRoutes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRouteBackend) GetRoute(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Routes.GetRoute(ctx, types.RouteID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRouteBackend) CreateRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreateRouteInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create route params: %w", err)
	}
	result, err := b.svc.Routes.CreateRoute(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRouteBackend) UpdateRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdateRouteInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update route params: %w", err)
	}
	result, err := b.svc.Routes.UpdateRoute(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRouteBackend) DeleteRoute(ctx context.Context, id string) error {
	return b.svc.Routes.DeleteRoute(ctx, b.ac, types.RouteID(id))
}

// ---------------------------------------------------------------------------
// AdminRouteBackend
// ---------------------------------------------------------------------------

type svcAdminRouteBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcAdminRouteBackend) ListAdminRoutes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Routes.ListAdminRoutes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminRouteBackend) GetAdminRoute(ctx context.Context, slug string) (json.RawMessage, error) {
	result, err := b.svc.Routes.GetAdminRouteBySlug(ctx, types.Slug(slug))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminRouteBackend) CreateAdminRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreateRouteInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin route params: %w", err)
	}
	result, err := b.svc.Routes.CreateAdminRoute(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminRouteBackend) UpdateAdminRoute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdateRouteInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin route params: %w", err)
	}
	result, err := b.svc.Routes.UpdateAdminRoute(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminRouteBackend) DeleteAdminRoute(ctx context.Context, id string) error {
	return b.svc.Routes.DeleteAdminRoute(ctx, b.ac, types.AdminRouteID(id))
}

func (b *svcAdminRouteBackend) ListAdminFieldTypes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Schema.ListAdminFieldTypes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminRouteBackend) GetAdminFieldType(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetAdminFieldType(ctx, types.AdminFieldTypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminRouteBackend) CreateAdminFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateAdminFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin field type params: %w", err)
	}
	result, err := b.svc.Schema.CreateAdminFieldType(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminRouteBackend) UpdateAdminFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateAdminFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin field type params: %w", err)
	}
	result, err := b.svc.Schema.UpdateAdminFieldType(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminRouteBackend) DeleteAdminFieldType(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteAdminFieldType(ctx, b.ac, types.AdminFieldTypeID(id))
}

// ---------------------------------------------------------------------------
// UserBackend
// ---------------------------------------------------------------------------

type svcUserBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcUserBackend) Whoami(ctx context.Context) (json.RawMessage, error) {
	// Whoami identifies the MCP operator. In direct mode the audit context
	// carries the user identity set at construction time.
	user, err := b.svc.Users.GetUser(ctx, b.ac.UserID)
	if err != nil {
		return nil, err
	}
	user.Hash = "" // never expose password hash
	return json.Marshal(user)
}

func (b *svcUserBackend) ListUsers(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Users.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	if result != nil {
		sanitizeUserList(*result)
	}
	return json.Marshal(result)
}

func (b *svcUserBackend) GetUser(ctx context.Context, id string) (json.RawMessage, error) {
	user, err := b.svc.Users.GetUser(ctx, types.UserID(id))
	if err != nil {
		return nil, err
	}
	user.Hash = "" // never expose password hash
	return json.Marshal(user)
}

func (b *svcUserBackend) CreateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreateUserInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create user params: %w", err)
	}
	// MCP operates as admin
	p.IsAdmin = true
	result, err := b.svc.Users.CreateUser(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	result.Hash = "" // never expose password hash
	return json.Marshal(result)
}

func (b *svcUserBackend) UpdateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdateUserInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update user params: %w", err)
	}
	// MCP operates as admin
	p.IsAdmin = true
	result, err := b.svc.Users.UpdateUser(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	result.Hash = "" // never expose password hash
	return json.Marshal(result)
}

func (b *svcUserBackend) DeleteUser(ctx context.Context, id string) error {
	return b.svc.Users.DeleteUser(ctx, b.ac, types.UserID(id))
}

func (b *svcUserBackend) ListUsersFull(ctx context.Context) (json.RawMessage, error) {
	// ListUsersFull assembles full user views with related entities.
	users, err := b.svc.Users.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return json.Marshal([]any{})
	}
	views := make([]db.UserFullView, 0, len(*users))
	for _, u := range *users {
		view, viewErr := b.svc.Users.GetUserFull(ctx, u.UserID)
		if viewErr != nil {
			continue
		}
		views = append(views, *view)
	}
	return json.Marshal(views)
}

func (b *svcUserBackend) GetUserFull(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Users.GetUserFull(ctx, types.UserID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// sanitizeUserList zeroes the Hash field on all users in a slice.
func sanitizeUserList(users []db.Users) {
	for i := range users {
		users[i].Hash = ""
	}
}

// ---------------------------------------------------------------------------
// RBACBackend
// ---------------------------------------------------------------------------

type svcRBACBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcRBACBackend) ListRoles(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.RBAC.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) GetRole(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.RBAC.GetRole(ctx, types.RoleID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) CreateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreateRoleInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create role params: %w", err)
	}
	result, err := b.svc.RBAC.CreateRole(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) UpdateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdateRoleInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update role params: %w", err)
	}
	result, err := b.svc.RBAC.UpdateRole(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) DeleteRole(ctx context.Context, id string) error {
	return b.svc.RBAC.DeleteRole(ctx, b.ac, types.RoleID(id))
}

func (b *svcRBACBackend) ListPermissions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.RBAC.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) GetPermission(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.RBAC.GetPermission(ctx, types.PermissionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) CreatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreatePermissionInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create permission params: %w", err)
	}
	result, err := b.svc.RBAC.CreatePermission(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) UpdatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdatePermissionInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update permission params: %w", err)
	}
	result, err := b.svc.RBAC.UpdatePermission(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) DeletePermission(ctx context.Context, id string) error {
	return b.svc.RBAC.DeletePermission(ctx, b.ac, types.PermissionID(id))
}

func (b *svcRBACBackend) AssignRolePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateRolePermissionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal assign role permission params: %w", err)
	}
	result, err := b.svc.RBAC.CreateRolePermission(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) RemoveRolePermission(ctx context.Context, id string) error {
	return b.svc.RBAC.DeleteRolePermission(ctx, b.ac, types.RolePermissionID(id))
}

func (b *svcRBACBackend) ListRolePermissions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.RBAC.ListRolePermissions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) GetRolePermission(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.RBAC.GetRolePermission(ctx, types.RolePermissionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) ListRolePermissionsByRole(ctx context.Context, roleID string) (json.RawMessage, error) {
	result, err := b.svc.RBAC.ListRolePermissionsByRoleID(ctx, types.RoleID(roleID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// SessionBackend
// ---------------------------------------------------------------------------

type svcSessionBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcSessionBackend) ListSessions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Sessions.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSessionBackend) GetSession(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Sessions.GetSession(ctx, types.SessionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSessionBackend) UpdateSession(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateSessionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update session params: %w", err)
	}
	result, err := b.svc.Sessions.UpdateSession(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSessionBackend) DeleteSession(ctx context.Context, id string) error {
	return b.svc.Sessions.DeleteSession(ctx, b.ac, types.SessionID(id))
}

// ---------------------------------------------------------------------------
// TokenBackend
// ---------------------------------------------------------------------------

type svcTokenBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcTokenBackend) ListTokens(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Tokens.ListTokens(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTokenBackend) GetToken(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Tokens.GetToken(ctx, id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTokenBackend) CreateToken(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreateTokenInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create token params: %w", err)
	}
	result, err := b.svc.Tokens.CreateToken(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTokenBackend) DeleteToken(ctx context.Context, id string) error {
	return b.svc.Tokens.DeleteToken(ctx, b.ac, id)
}

// ---------------------------------------------------------------------------
// SSHKeyBackend
// ---------------------------------------------------------------------------

type svcSSHKeyBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcSSHKeyBackend) ListSSHKeys(ctx context.Context) (json.RawMessage, error) {
	// List SSH keys for the audit context user.
	userID := types.NullableUserID{ID: b.ac.UserID, Valid: !b.ac.UserID.IsZero()}
	result, err := b.svc.SSHKeys.ListKeys(ctx, userID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSSHKeyBackend) CreateSSHKey(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.AddSSHKeyInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create ssh key params: %w", err)
	}
	// Default to the audit context user if not set.
	if p.UserID.IsZero() {
		p.UserID = b.ac.UserID
	}
	result, err := b.svc.SSHKeys.AddKey(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSSHKeyBackend) DeleteSSHKey(ctx context.Context, id string) error {
	return b.svc.SSHKeys.DeleteKey(ctx, b.ac, b.ac.UserID, id)
}

// ---------------------------------------------------------------------------
// OAuthBackend
// ---------------------------------------------------------------------------

type svcOAuthBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcOAuthBackend) ListUsersOAuth(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Driver().ListUserOauths()
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcOAuthBackend) GetUserOAuth(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Driver().GetUserOauth(types.UserOauthID(id))
	if err != nil {
		return nil, &service.NotFoundError{Resource: "user_oauth", ID: id}
	}
	return json.Marshal(result)
}

func (b *svcOAuthBackend) CreateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateUserOauthParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create user oauth params: %w", err)
	}
	result, err := b.svc.OAuth.CreateUserOauth(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcOAuthBackend) UpdateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateUserOauthParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update user oauth params: %w", err)
	}
	result, err := b.svc.OAuth.UpdateUserOauth(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	// UpdateUserOauth returns *string. Fetch full entity if we have an ID.
	if result != nil {
		entity, fetchErr := b.svc.Driver().GetUserOauth(p.UserOauthID)
		if fetchErr == nil {
			return json.Marshal(entity)
		}
	}
	return json.Marshal(result)
}

func (b *svcOAuthBackend) DeleteUserOAuth(ctx context.Context, id string) error {
	return b.svc.OAuth.DeleteUserOauth(ctx, b.ac, types.UserOauthID(id))
}

// ---------------------------------------------------------------------------
// TableBackend
// ---------------------------------------------------------------------------

type svcTableBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcTableBackend) ListTables(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Tables.ListTables(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTableBackend) GetTable(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Tables.GetTable(ctx, id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTableBackend) CreateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	// TableService does not have a CreateTable method. Call the driver directly.
	var p db.CreateTableParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create table params: %w", err)
	}
	result, err := b.svc.Driver().CreateTable(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTableBackend) UpdateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateTableParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update table params: %w", err)
	}
	result, err := b.svc.Tables.UpdateTable(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTableBackend) DeleteTable(ctx context.Context, id string) error {
	return b.svc.Tables.DeleteTable(ctx, b.ac, id)
}

// ---------------------------------------------------------------------------
// PluginBackend
// ---------------------------------------------------------------------------

type svcPluginBackend struct {
	svc *service.Registry
}

func (b *svcPluginBackend) ListPlugins(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Plugins.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcPluginBackend) GetPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.svc.Plugins.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcPluginBackend) ReloadPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	if err := b.svc.Plugins.Reload(ctx, name); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "reloaded", "name": name})
}

func (b *svcPluginBackend) EnablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	if err := b.svc.Plugins.Enable(ctx, name); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "enabled", "name": name})
}

func (b *svcPluginBackend) DisablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	if err := b.svc.Plugins.Disable(ctx, name); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "disabled", "name": name})
}

func (b *svcPluginBackend) PluginCleanupDryRun(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Plugins.CleanupDryRun(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"orphaned_tables": result})
}

func (b *svcPluginBackend) PluginCleanupDrop(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Tables []string `json:"tables"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal plugin cleanup drop params: %w", err)
	}
	result, err := b.svc.Plugins.CleanupDrop(ctx, input.Tables)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"dropped": result})
}

func (b *svcPluginBackend) ListPluginRoutes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Plugins.ListRoutes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcPluginBackend) ApprovePluginRoutes(ctx context.Context, params json.RawMessage) error {
	var input struct {
		Routes     []service.RouteApprovalInput `json:"routes"`
		ApprovedBy string                       `json:"approved_by"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return fmt.Errorf("unmarshal approve plugin routes params: %w", err)
	}
	return b.svc.Plugins.ApproveRoutes(ctx, input.Routes, input.ApprovedBy)
}

func (b *svcPluginBackend) RevokePluginRoutes(ctx context.Context, params json.RawMessage) error {
	var input struct {
		Routes []service.RouteApprovalInput `json:"routes"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return fmt.Errorf("unmarshal revoke plugin routes params: %w", err)
	}
	return b.svc.Plugins.RevokeRoutes(ctx, input.Routes)
}

func (b *svcPluginBackend) ListPluginHooks(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Plugins.ListHooks(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcPluginBackend) ApprovePluginHooks(ctx context.Context, params json.RawMessage) error {
	var input struct {
		Hooks      []service.HookApprovalInput `json:"hooks"`
		ApprovedBy string                      `json:"approved_by"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return fmt.Errorf("unmarshal approve plugin hooks params: %w", err)
	}
	return b.svc.Plugins.ApproveHooks(ctx, input.Hooks, input.ApprovedBy)
}

func (b *svcPluginBackend) RevokePluginHooks(ctx context.Context, params json.RawMessage) error {
	var input struct {
		Hooks []service.HookApprovalInput `json:"hooks"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return fmt.Errorf("unmarshal revoke plugin hooks params: %w", err)
	}
	return b.svc.Plugins.RevokeHooks(ctx, input.Hooks)
}

// ---------------------------------------------------------------------------
// ConfigBackend
// ---------------------------------------------------------------------------

type svcConfigBackend struct {
	svc *service.Registry
}

func (b *svcConfigBackend) GetConfig(ctx context.Context, category string) (json.RawMessage, error) {
	if category != "" {
		result, err := b.svc.ConfigSvc.GetConfigByCategory(category)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)
	}
	data, err := b.svc.ConfigSvc.GetConfig()
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func (b *svcConfigBackend) GetConfigMeta(ctx context.Context) (json.RawMessage, error) {
	fields, categories := b.svc.ConfigSvc.GetFieldMetadata()
	return json.Marshal(map[string]any{
		"fields":     fields,
		"categories": categories,
	})
}

func (b *svcConfigBackend) UpdateConfig(ctx context.Context, updates map[string]any) (json.RawMessage, error) {
	result, err := b.svc.ConfigSvc.UpdateConfig(updates)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// ImportBackend
// ---------------------------------------------------------------------------

type svcImportBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcImportBackend) ImportContent(ctx context.Context, format string, data any) (json.RawMessage, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal import data: %w", err)
	}
	result, err := b.svc.Import.ImportContent(ctx, b.ac, service.ImportContentInput{
		Format: config.OutputFormat(format),
		Body:   body,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcImportBackend) ImportBulk(ctx context.Context, format string, data any) (json.RawMessage, error) {
	// ImportBulk is not yet implemented in the ImportService.
	// Delegate to ImportContent as a reasonable fallback.
	return b.ImportContent(ctx, format, data)
}

// ---------------------------------------------------------------------------
// DeployBackend
// ---------------------------------------------------------------------------

type svcDeployBackend struct {
	svc *service.Registry
}

func (b *svcDeployBackend) DeployHealth(ctx context.Context) (json.RawMessage, error) {
	// DeployService is a placeholder with no methods. Return a status indicator.
	return json.Marshal(map[string]string{"status": "not implemented in direct mode"})
}

func (b *svcDeployBackend) DeployExport(ctx context.Context, tables []string) (json.RawMessage, error) {
	return nil, fmt.Errorf("deploy export is not supported in direct mode")
}

func (b *svcDeployBackend) DeployImport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("deploy import is not supported in direct mode")
}

func (b *svcDeployBackend) DeployDryRun(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("deploy dry run is not supported in direct mode")
}

// ---------------------------------------------------------------------------
// HealthBackend
// ---------------------------------------------------------------------------

type svcHealthBackend struct {
	svc *service.Registry
}

func (b *svcHealthBackend) Health(ctx context.Context) (json.RawMessage, error) {
	resp := map[string]any{
		"status": "ok",
		"checks": map[string]bool{},
	}

	checks := resp["checks"].(map[string]bool)

	// Database check via driver Ping.
	driver := b.svc.Driver()
	if driver != nil {
		if err := driver.Ping(); err != nil {
			checks["database"] = false
			resp["status"] = "degraded"
		} else {
			checks["database"] = true
		}
	} else {
		checks["database"] = false
		resp["status"] = "degraded"
	}

	// Plugin health check if available.
	if b.svc.Plugins != nil {
		pluginHealth, err := b.svc.Plugins.Health(ctx)
		if err != nil {
			checks["plugins"] = false
		} else {
			checks["plugins"] = pluginHealth.Healthy
			if !pluginHealth.Healthy {
				resp["status"] = "degraded"
			}
		}
	}

	return json.Marshal(resp)
}

// ---------------------------------------------------------------------------
// AdminMediaBackend
// ---------------------------------------------------------------------------

type svcAdminMediaBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcAdminMediaBackend) ListAdminMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	d := b.svc.Driver()
	items, err := d.ListAdminMediaPaginated(db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	total, err := d.CountAdminMedia()
	if err != nil {
		return nil, err
	}
	resp := db.PaginatedResponse[mcpAdminMediaResponse]{
		Limit:  limit,
		Offset: offset,
	}
	if items != nil {
		resp.Data = toMCPAdminMediaList(*items)
	}
	if total != nil {
		resp.Total = *total
	}
	return json.Marshal(resp)
}

func (b *svcAdminMediaBackend) GetAdminMedia(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()
	mid := types.AdminMediaID(id)
	if err := mid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	result, err := d.GetAdminMedia(mid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media", ID: id}
	}
	return json.Marshal(toMCPAdminMediaResponse(*result))
}

func (b *svcAdminMediaBackend) UpdateAdminMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		AdminMediaID string   `json:"admin_media_id"`
		Name         *string  `json:"name"`
		DisplayName  *string  `json:"display_name"`
		Alt          *string  `json:"alt"`
		Caption      *string  `json:"caption"`
		Description  *string  `json:"description"`
		Class        *string  `json:"class"`
		FocalX       *float64 `json:"focal_x"`
		FocalY       *float64 `json:"focal_y"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin media params: %w", err)
	}

	mid := types.AdminMediaID(p.AdminMediaID)
	if err := mid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "admin_media_id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	existing, err := d.GetAdminMedia(mid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media", ID: p.AdminMediaID}
	}

	updateParams := db.UpdateAdminMediaParams{
		AdminMediaID: mid,
		Name:         existing.Name,
		DisplayName:  existing.DisplayName,
		Alt:          existing.Alt,
		Caption:      existing.Caption,
		Description:  existing.Description,
		Class:        existing.Class,
		Mimetype:     existing.Mimetype,
		Dimensions:   existing.Dimensions,
		URL:          existing.URL,
		Srcset:       existing.Srcset,
		FocalX:       existing.FocalX,
		FocalY:       existing.FocalY,
		AuthorID:     existing.AuthorID,
		FolderID:     existing.FolderID,
		DateCreated:  existing.DateCreated,
		DateModified: types.NewTimestamp(time.Now().UTC()),
	}

	if p.Name != nil {
		updateParams.Name = db.NewNullString(*p.Name)
	}
	if p.DisplayName != nil {
		updateParams.DisplayName = db.NewNullString(*p.DisplayName)
	}
	if p.Alt != nil {
		updateParams.Alt = db.NewNullString(*p.Alt)
	}
	if p.Caption != nil {
		updateParams.Caption = db.NewNullString(*p.Caption)
	}
	if p.Description != nil {
		updateParams.Description = db.NewNullString(*p.Description)
	}
	if p.Class != nil {
		updateParams.Class = db.NewNullString(*p.Class)
	}
	if p.FocalX != nil {
		updateParams.FocalX = types.NullableFloat64{Float64: *p.FocalX, Valid: true}
	}
	if p.FocalY != nil {
		updateParams.FocalY = types.NullableFloat64{Float64: *p.FocalY, Valid: true}
	}

	if _, err := d.UpdateAdminMedia(ctx, b.ac, updateParams); err != nil {
		return nil, err
	}

	updated, err := d.GetAdminMedia(mid)
	if err != nil {
		return nil, err
	}
	return json.Marshal(toMCPAdminMediaResponse(*updated))
}

func (b *svcAdminMediaBackend) DeleteAdminMedia(ctx context.Context, id string) error {
	d := b.svc.Driver()
	mid := types.AdminMediaID(id)
	if err := mid.Validate(); err != nil {
		return &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	return d.DeleteAdminMedia(ctx, b.ac, mid)
}

func (b *svcAdminMediaBackend) UploadAdminMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error) {
	// The MediaService.Upload requires multipart.File and *multipart.FileHeader,
	// which are HTTP-specific types. In direct mode we cannot easily construct
	// these from an io.Reader. Return an unsupported error directing callers
	// to use the REST API for admin media upload.
	return nil, fmt.Errorf("admin media upload is not supported in direct mode; use the REST API")
}

func (b *svcAdminMediaBackend) ListMediaDimensions(ctx context.Context) (json.RawMessage, error) {
	// Dimensions are shared between public and admin media.
	result, err := b.svc.Media.ListMediaDimensions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// AdminMediaFolderBackend
// ---------------------------------------------------------------------------

type svcAdminMediaFolderBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcAdminMediaFolderBackend) ListAdminMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error) {
	d := b.svc.Driver()
	if parentID != "" {
		pid := types.AdminMediaFolderID(parentID)
		if err := pid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		folders, err := d.ListAdminMediaFoldersByParent(pid)
		if err != nil {
			return nil, err
		}
		return json.Marshal(folders)
	}
	folders, err := d.ListAdminMediaFoldersAtRoot()
	if err != nil {
		return nil, err
	}
	return json.Marshal(folders)
}

func (b *svcAdminMediaFolderBackend) GetAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()
	fid := types.AdminMediaFolderID(id)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	folder, err := d.GetAdminMediaFolder(fid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media_folder", ID: id}
	}
	return json.Marshal(folder)
}

func (b *svcAdminMediaFolderBackend) CreateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin media folder params: %w", err)
	}
	if p.Name == "" {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "name", Message: "required"}}}
	}

	var parentID types.NullableAdminMediaFolderID
	if p.ParentID != nil && *p.ParentID != "" {
		pid := types.AdminMediaFolderID(*p.ParentID)
		if err := pid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		if _, err := d.GetAdminMediaFolder(pid); err != nil {
			return nil, &service.NotFoundError{Resource: "parent_folder", ID: *p.ParentID}
		}
		parentID = types.NullableAdminMediaFolderID{ID: pid, Valid: true}

		breadcrumb, err := d.GetAdminMediaFolderBreadcrumb(pid)
		if err != nil {
			return nil, fmt.Errorf("check folder depth: %w", err)
		}
		if len(breadcrumb)+1 > 10 {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: "creating this folder would exceed maximum folder depth of 10"}}}
		}
	}

	if err := d.ValidateAdminMediaFolderName(p.Name, parentID); err != nil {
		return nil, &service.ConflictError{Resource: "admin_media_folder", Detail: err.Error()}
	}

	now := types.NewTimestamp(time.Now().UTC())
	folder, err := d.CreateAdminMediaFolder(ctx, b.ac, db.CreateAdminMediaFolderParams{
		Name:         p.Name,
		ParentID:     parentID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(folder)
}

func (b *svcAdminMediaFolderBackend) UpdateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		FolderID string  `json:"folder_id"`
		Name     *string `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin media folder params: %w", err)
	}

	fid := types.AdminMediaFolderID(p.FolderID)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	existing, err := d.GetAdminMediaFolder(fid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media_folder", ID: p.FolderID}
	}

	name := existing.Name
	parentID := existing.ParentID

	if p.Name != nil {
		if *p.Name == "" {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "name", Message: "cannot be empty"}}}
		}
		name = *p.Name
	}

	parentChanged := false
	if p.ParentID != nil {
		parentChanged = true
		if *p.ParentID == "" {
			parentID = types.NullableAdminMediaFolderID{}
		} else {
			pid := types.AdminMediaFolderID(*p.ParentID)
			if err := pid.Validate(); err != nil {
				return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
			}
			parentID = types.NullableAdminMediaFolderID{ID: pid, Valid: true}
		}
	}

	if parentChanged {
		if err := d.ValidateAdminMediaFolderMove(fid, parentID); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: err.Error()}}}
		}
	}

	nameChanged := p.Name != nil && name != existing.Name
	if nameChanged || parentChanged {
		if err := d.ValidateAdminMediaFolderName(name, parentID); err != nil {
			return nil, &service.ConflictError{Resource: "admin_media_folder", Detail: err.Error()}
		}
	}

	_, err = d.UpdateAdminMediaFolder(ctx, b.ac, db.UpdateAdminMediaFolderParams{
		AdminFolderID: fid,
		Name:          name,
		ParentID:      parentID,
		DateModified:  types.NewTimestamp(time.Now().UTC()),
	})
	if err != nil {
		return nil, err
	}

	updated, err := d.GetAdminMediaFolder(fid)
	if err != nil {
		return nil, err
	}
	return json.Marshal(updated)
}

func (b *svcAdminMediaFolderBackend) DeleteAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()

	fid := types.AdminMediaFolderID(id)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	if _, err := d.GetAdminMediaFolder(fid); err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media_folder", ID: id}
	}

	children, err := d.ListAdminMediaFoldersByParent(fid)
	if err != nil {
		return nil, err
	}
	childCount := 0
	if children != nil {
		childCount = len(*children)
	}

	folderNullable := types.NullableAdminMediaFolderID{ID: fid, Valid: true}
	mediaCount, err := d.CountAdminMediaByFolder(folderNullable)
	if err != nil {
		return nil, err
	}

	mc := int64(0)
	if mediaCount != nil {
		mc = *mediaCount
	}

	if childCount > 0 || mc > 0 {
		return json.Marshal(map[string]any{
			"error":         fmt.Sprintf("cannot delete folder: contains %d child folder(s) and %d media item(s)", childCount, mc),
			"child_folders": childCount,
			"media_items":   mc,
		})
	}

	if err := d.DeleteAdminMediaFolder(ctx, b.ac, fid); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "deleted"})
}

func (b *svcAdminMediaFolderBackend) MoveAdminMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		MediaIDs []string `json:"media_ids"`
		FolderID *string  `json:"folder_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal move admin media params: %w", err)
	}

	if len(p.MediaIDs) == 0 {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "media_ids", Message: "required and cannot be empty"}}}
	}
	if len(p.MediaIDs) > 100 {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "media_ids", Message: "batch size cannot exceed 100 items"}}}
	}

	var folderID types.NullableAdminMediaFolderID
	if p.FolderID != nil && *p.FolderID != "" {
		fid := types.AdminMediaFolderID(*p.FolderID)
		if err := fid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		if _, err := d.GetAdminMediaFolder(fid); err != nil {
			return nil, &service.NotFoundError{Resource: "admin_media_folder", ID: *p.FolderID}
		}
		folderID = types.NullableAdminMediaFolderID{ID: fid, Valid: true}
	}

	now := types.NewTimestamp(time.Now().UTC())
	moved := 0
	for _, idStr := range p.MediaIDs {
		mid := types.AdminMediaID(idStr)
		if err := mid.Validate(); err != nil {
			continue
		}
		err := d.MoveAdminMediaToFolder(ctx, b.ac, db.MoveAdminMediaToFolderParams{
			FolderID:     folderID,
			DateModified: now,
			AdminMediaID: mid,
		})
		if err != nil {
			continue
		}
		moved++
	}

	return json.Marshal(map[string]int{"moved": moved})
}
