package mcp

import (
	"context"
	"encoding/json"
)

type proxyLocaleBackend struct{ p *proxyBackends }

func (b *proxyLocaleBackend) ListLocales(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Locales.ListLocales(ctx)
}

func (b *proxyLocaleBackend) ListAdminLocales(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Locales.ListAdminLocales(ctx)
}

func (b *proxyLocaleBackend) GetLocale(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Locales.GetLocale(ctx, id)
}

func (b *proxyLocaleBackend) CreateLocale(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Locales.CreateLocale(ctx, params)
}

func (b *proxyLocaleBackend) UpdateLocale(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Locales.UpdateLocale(ctx, params)
}

func (b *proxyLocaleBackend) DeleteLocale(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Locales.DeleteLocale(ctx, id)
}

func (b *proxyLocaleBackend) CreateTranslation(ctx context.Context, contentDataID string, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Locales.CreateTranslation(ctx, contentDataID, params)
}

func (b *proxyLocaleBackend) AdminCreateTranslation(ctx context.Context, adminContentDataID string, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Locales.AdminCreateTranslation(ctx, adminContentDataID, params)
}
