package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

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

func (b *svcRouteBackend) ListRoutesFull(ctx context.Context) (json.RawMessage, error) {
	routes, err := b.svc.Routes.ListRoutes(ctx)
	if err != nil {
		return nil, err
	}
	if routes == nil {
		return json.Marshal([]any{})
	}
	views := make([]db.RouteFullView, 0, len(*routes))
	for _, r := range *routes {
		view, viewErr := b.svc.Routes.GetRouteFull(ctx, r.RouteID)
		if viewErr != nil {
			continue
		}
		views = append(views, *view)
	}
	return json.Marshal(views)
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
