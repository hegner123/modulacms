package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

type svcLocaleBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcLocaleBackend) ListLocales(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Locales.ListEnabledLocales(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcLocaleBackend) ListAdminLocales(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Locales.ListLocales(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcLocaleBackend) GetLocale(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Locales.GetLocale(ctx, types.LocaleID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcLocaleBackend) CreateLocale(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreateLocaleInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create locale params: %w", err)
	}
	result, err := b.svc.Locales.CreateLocale(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcLocaleBackend) UpdateLocale(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdateLocaleInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update locale params: %w", err)
	}
	result, err := b.svc.Locales.UpdateLocale(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcLocaleBackend) DeleteLocale(ctx context.Context, id string) error {
	return b.svc.Locales.DeleteLocale(ctx, b.ac, types.LocaleID(id))
}

func (b *svcLocaleBackend) CreateTranslation(ctx context.Context, contentDataID string, params json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Locale string `json:"locale"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create translation params: %w", err)
	}
	result, err := b.svc.Locales.CreateTranslation(ctx, b.ac, types.ContentID(contentDataID), p.Locale, b.ac.UserID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcLocaleBackend) AdminCreateTranslation(ctx context.Context, adminContentDataID string, params json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Locale string `json:"locale"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal admin create translation params: %w", err)
	}
	result, err := b.svc.Locales.CreateAdminTranslation(ctx, b.ac, types.AdminContentID(adminContentDataID), p.Locale, b.ac.UserID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
