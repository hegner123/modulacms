package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

type svcWebhookBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcWebhookBackend) ListWebhooks(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Webhooks.ListWebhooks(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcWebhookBackend) GetWebhook(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Webhooks.GetWebhook(ctx, types.WebhookID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcWebhookBackend) CreateWebhook(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Name     string   `json:"name"`
		URL      string   `json:"url"`
		Secret   string   `json:"secret"`
		Events   []string `json:"events"`
		IsActive bool     `json:"is_active"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal create webhook params: %w", err)
	}
	result, err := b.svc.Webhooks.CreateWebhook(ctx, b.ac, service.CreateWebhookInput{
		Name:     input.Name,
		URL:      input.URL,
		Secret:   input.Secret,
		Events:   input.Events,
		IsActive: input.IsActive,
	}, b.ac.UserID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcWebhookBackend) UpdateWebhook(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		WebhookID string   `json:"webhook_id"`
		Name      string   `json:"name"`
		URL       string   `json:"url"`
		Secret    string   `json:"secret"`
		Events    []string `json:"events"`
		IsActive  bool     `json:"is_active"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal update webhook params: %w", err)
	}
	result, err := b.svc.Webhooks.UpdateWebhook(ctx, b.ac, service.UpdateWebhookInput{
		WebhookID: types.WebhookID(input.WebhookID),
		Name:      input.Name,
		URL:       input.URL,
		Secret:    input.Secret,
		Events:    input.Events,
		IsActive:  input.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcWebhookBackend) DeleteWebhook(ctx context.Context, id string) error {
	return b.svc.Webhooks.DeleteWebhook(ctx, b.ac, types.WebhookID(id))
}

func (b *svcWebhookBackend) TestWebhook(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Webhooks.TestWebhook(ctx, types.WebhookID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcWebhookBackend) ListWebhookDeliveries(ctx context.Context, webhookID string) (json.RawMessage, error) {
	result, err := b.svc.Webhooks.ListDeliveries(ctx, types.WebhookID(webhookID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcWebhookBackend) RetryWebhookDelivery(ctx context.Context, deliveryID string) error {
	_, err := b.svc.Webhooks.RetryDelivery(ctx, types.WebhookDeliveryID(deliveryID))
	return err
}
