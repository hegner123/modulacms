package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	modula "github.com/hegner123/modulacms/sdks/go"
)

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

func (b *sdkContentBackend) QueryContent(ctx context.Context, datatype string, params json.RawMessage) (json.RawMessage, error) {
	var qp struct {
		Sort   string            `json:"sort"`
		Limit  int               `json:"limit"`
		Offset int               `json:"offset"`
		Filter map[string]string `json:"filter"`
	}
	if params != nil {
		if err := json.Unmarshal(params, &qp); err != nil {
			return nil, fmt.Errorf("unmarshal query params: %w", err)
		}
	}
	sdkParams := &modula.QueryParams{
		Sort:    qp.Sort,
		Limit:   qp.Limit,
		Offset:  qp.Offset,
		Filters: qp.Filter,
	}
	result, err := b.client.Query.Query(ctx, datatype, sdkParams)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkContentBackend) GetGlobals(ctx context.Context, format string) (json.RawMessage, error) {
	return b.client.Globals.List(ctx)
}

func (b *sdkContentBackend) GetContentFull(ctx context.Context, id string) (json.RawMessage, error) {
	return b.client.ContentDataFull.GetFull(ctx, modula.ContentID(id))
}

func (b *sdkContentBackend) GetContentByRoute(ctx context.Context, routeID string) (json.RawMessage, error) {
	return b.client.ContentDataFull.ListByRoute(ctx, modula.RouteID(routeID))
}

func (b *sdkContentBackend) CreateContentComposite(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.ContentCreateParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal composite create params: %w", err)
	}
	result, err := b.client.ContentComposite.CreateWithFields(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
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

func (b *sdkAdminContentBackend) AdminGetContentFull(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	// List admin content with pagination, then fetch full details for each item.
	page, err := b.client.AdminContentData.ListPaginated(ctx, modula.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	type fullResult struct {
		Items  []json.RawMessage `json:"items"`
		Total  int64             `json:"total"`
		Limit  int64             `json:"limit"`
		Offset int64             `json:"offset"`
	}
	var items []json.RawMessage
	for _, item := range page.Data {
		full, err := b.client.ContentDataFull.AdminGetFull(ctx, item.AdminContentDataID)
		if err != nil {
			return nil, fmt.Errorf("get admin content full %s: %w", item.AdminContentDataID, err)
		}
		items = append(items, full)
	}
	return json.Marshal(fullResult{
		Items:  items,
		Total:  page.Total,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

func (b *sdkAdminContentBackend) GetAdminTree(ctx context.Context, slug string) (json.RawMessage, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug is required for admin tree lookup")
	}
	return b.client.AdminTree.Get(ctx, slug, "")
}

// ---------------------------------------------------------------------------
// VersionBackend
// ---------------------------------------------------------------------------

type sdkVersionBackend struct {
	client *modula.Client
}

func (b *sdkVersionBackend) ListVersions(ctx context.Context, contentID string) (json.RawMessage, error) {
	result, err := b.client.Publishing.ListVersions(ctx, contentID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkVersionBackend) GetVersion(ctx context.Context, versionID string) (json.RawMessage, error) {
	result, err := b.client.Publishing.GetVersion(ctx, versionID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkVersionBackend) CreateVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateVersionRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create version params: %w", err)
	}
	result, err := b.client.Publishing.CreateVersion(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkVersionBackend) DeleteVersion(ctx context.Context, versionID string) error {
	return b.client.Publishing.DeleteVersion(ctx, versionID)
}

func (b *sdkVersionBackend) RestoreVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.RestoreRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal restore version params: %w", err)
	}
	result, err := b.client.Publishing.Restore(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkVersionBackend) AdminListVersions(ctx context.Context, contentID string) (json.RawMessage, error) {
	result, err := b.client.AdminPublishing.ListAdminVersions(ctx, contentID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkVersionBackend) AdminGetVersion(ctx context.Context, versionID string) (json.RawMessage, error) {
	result, err := b.client.AdminPublishing.GetAdminVersion(ctx, versionID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkVersionBackend) AdminCreateVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminVersionRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal admin create version params: %w", err)
	}
	result, err := b.client.AdminPublishing.CreateAdminVersion(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkVersionBackend) AdminDeleteVersion(ctx context.Context, versionID string) error {
	return b.client.AdminPublishing.DeleteVersion(ctx, versionID)
}

func (b *sdkVersionBackend) AdminRestoreVersion(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.AdminRestoreRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal admin restore version params: %w", err)
	}
	result, err := b.client.AdminPublishing.AdminRestore(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
