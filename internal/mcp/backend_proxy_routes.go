package mcp

import (
	"context"
	"encoding/json"
)

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
func (b *proxyRouteBackend) ListRoutesFull(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Routes.ListRoutesFull(ctx)
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
