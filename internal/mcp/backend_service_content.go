package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/tree/ops"
)

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

	result, err := b.svc.Content.Reorder(ctx, b.ac, parentID, orderedIDs)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int{"updated": result.Updated})
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

func (b *svcContentBackend) QueryContent(ctx context.Context, datatype string, params json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("query_content is not available in direct mode; use remote mode with an API key")
}

func (b *svcContentBackend) GetGlobals(ctx context.Context, format string) (json.RawMessage, error) {
	return nil, fmt.Errorf("get_globals is not available in direct mode; use remote mode with an API key")
}

func (b *svcContentBackend) GetContentFull(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Content.GetFull(ctx, types.ContentID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) GetContentByRoute(ctx context.Context, routeID string) (json.RawMessage, error) {
	result, err := b.svc.Content.ListByRoute(ctx, types.RouteID(routeID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcContentBackend) CreateContentComposite(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("create_content_composite is not available in direct mode; use remote mode with an API key")
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

	result, err := b.svc.AdminContent.Reorder(ctx, b.ac, parentID, orderedIDs)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int{"updated": result.Updated})
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

func (b *svcAdminContentBackend) AdminGetContentFull(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	page, err := b.svc.AdminContent.ListPaginated(ctx, db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	type fullItem struct {
		*db.AdminContentDataView
	}
	type fullResult struct {
		Items  []fullItem `json:"items"`
		Total  int64      `json:"total"`
		Limit  int64      `json:"limit"`
		Offset int64      `json:"offset"`
	}
	var items []fullItem
	for _, item := range page.Data {
		view, err := b.svc.AdminContent.GetFull(ctx, item.AdminContentDataID)
		if err != nil {
			return nil, fmt.Errorf("get admin content full %s: %w", item.AdminContentDataID, err)
		}
		items = append(items, fullItem{view})
	}
	return json.Marshal(fullResult{
		Items:  items,
		Total:  page.Total,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

func (b *svcAdminContentBackend) GetAdminTree(ctx context.Context, slug string) (json.RawMessage, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug is required for admin tree lookup")
	}
	d := b.svc.Driver()
	route, err := d.GetAdminRoute(types.Slug(slug))
	if err != nil {
		return nil, fmt.Errorf("get admin route %q: %w", slug, err)
	}
	routeID := types.NullableAdminRouteID{ID: route.AdminRouteID, Valid: true}
	contentData, err := d.ListAdminContentDataWithDatatypeByRoute(routeID)
	if err != nil {
		return nil, fmt.Errorf("list admin content data with datatypes: %w", err)
	}
	contentFields, err := d.ListAdminContentFieldsWithFieldByRoute(routeID)
	if err != nil {
		return nil, fmt.Errorf("list admin content fields with fields: %w", err)
	}
	return json.Marshal(map[string]any{
		"route":          route,
		"content_data":   contentData,
		"content_fields": contentFields,
	})
}

// ---------------------------------------------------------------------------
// VersionBackend
// ---------------------------------------------------------------------------

type svcVersionBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcVersionBackend) ListVersions(ctx context.Context, contentID string) (json.RawMessage, error) {
	result, err := b.svc.Content.ListVersions(ctx, types.ContentID(contentID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcVersionBackend) GetVersion(ctx context.Context, versionID string) (json.RawMessage, error) {
	result, err := b.svc.Content.GetVersion(ctx, types.ContentVersionID(versionID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcVersionBackend) CreateVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		ContentDataID string `json:"content_data_id"`
		Label         string `json:"label"`
		Locale        string `json:"locale"`
		UserID        string `json:"user_id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal create version params: %w", err)
	}
	if input.Locale == "" {
		input.Locale = "en"
	}
	result, err := b.svc.Content.CreateVersion(ctx, b.ac, types.ContentID(input.ContentDataID), input.Locale, input.Label, types.UserID(input.UserID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcVersionBackend) DeleteVersion(ctx context.Context, versionID string) error {
	return b.svc.Content.DeleteVersion(ctx, b.ac, types.ContentVersionID(versionID))
}

func (b *svcVersionBackend) RestoreVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		ContentDataID    string `json:"content_data_id"`
		ContentVersionID string `json:"content_version_id"`
		UserID           string `json:"user_id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal restore version params: %w", err)
	}
	result, err := b.svc.Content.RestoreVersion(ctx, b.ac, types.ContentID(input.ContentDataID), types.ContentVersionID(input.ContentVersionID), types.UserID(input.UserID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcVersionBackend) AdminListVersions(ctx context.Context, contentID string) (json.RawMessage, error) {
	result, err := b.svc.AdminContent.ListVersions(ctx, types.AdminContentID(contentID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcVersionBackend) AdminGetVersion(ctx context.Context, versionID string) (json.RawMessage, error) {
	result, err := b.svc.AdminContent.GetVersion(ctx, types.AdminContentVersionID(versionID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcVersionBackend) AdminCreateVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		AdminContentDataID string `json:"admin_content_data_id"`
		Label              string `json:"label"`
		Locale             string `json:"locale"`
		UserID             string `json:"user_id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal admin create version params: %w", err)
	}
	if input.Locale == "" {
		input.Locale = "en"
	}
	result, err := b.svc.AdminContent.CreateVersion(ctx, b.ac, types.AdminContentID(input.AdminContentDataID), input.Locale, input.Label, types.UserID(input.UserID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcVersionBackend) AdminDeleteVersion(ctx context.Context, versionID string) error {
	return b.svc.AdminContent.DeleteVersion(ctx, b.ac, types.AdminContentVersionID(versionID))
}

func (b *svcVersionBackend) AdminRestoreVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		AdminContentDataID    string `json:"admin_content_data_id"`
		AdminContentVersionID string `json:"admin_content_version_id"`
		UserID                string `json:"user_id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal admin restore version params: %w", err)
	}
	result, err := b.svc.AdminContent.RestoreVersion(ctx, b.ac, types.AdminContentID(input.AdminContentDataID), types.AdminContentVersionID(input.AdminContentVersionID), types.UserID(input.UserID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
