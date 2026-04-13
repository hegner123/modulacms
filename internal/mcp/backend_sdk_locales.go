package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	modula "github.com/hegner123/modulacms/sdks/go"
)

type sdkLocaleBackend struct {
	client *modula.Client
}

func (b *sdkLocaleBackend) ListLocales(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Locales.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkLocaleBackend) ListAdminLocales(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Locales.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkLocaleBackend) GetLocale(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Locales.Get(ctx, modula.LocaleID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkLocaleBackend) CreateLocale(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateLocaleRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create locale params: %w", err)
	}
	result, err := b.client.Locales.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkLocaleBackend) UpdateLocale(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateLocaleRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update locale params: %w", err)
	}
	result, err := b.client.Locales.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkLocaleBackend) DeleteLocale(ctx context.Context, id string) error {
	return b.client.Locales.Delete(ctx, modula.LocaleID(id))
}

func (b *sdkLocaleBackend) CreateTranslation(ctx context.Context, contentDataID string, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateTranslationRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create translation params: %w", err)
	}
	result, err := b.client.Locales.CreateTranslation(ctx, contentDataID, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkLocaleBackend) AdminCreateTranslation(ctx context.Context, adminContentDataID string, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateTranslationRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal admin create translation params: %w", err)
	}
	result, err := b.client.Locales.CreateAdminTranslation(ctx, adminContentDataID, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
