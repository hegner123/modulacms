package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	modula "github.com/hegner123/modulacms/sdks/go"
)

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

func (b *sdkRouteBackend) ListRoutesFull(ctx context.Context) (json.RawMessage, error) {
	routes, err := b.client.Routes.List(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]json.RawMessage, 0, len(routes))
	for _, r := range routes {
		view, viewErr := b.client.RoutesFull.GetFull(ctx, r.RouteID)
		if viewErr != nil {
			continue
		}
		views = append(views, view)
	}
	return json.Marshal(views)
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
