package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	modula "github.com/hegner123/modulacms/sdks/go"
)

type sdkWebhookBackend struct {
	client *modula.Client
}

func (b *sdkWebhookBackend) ListWebhooks(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Webhooks.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkWebhookBackend) GetWebhook(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Webhooks.Get(ctx, modula.WebhookID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkWebhookBackend) CreateWebhook(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateWebhookRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create webhook params: %w", err)
	}
	result, err := b.client.Webhooks.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkWebhookBackend) UpdateWebhook(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateWebhookRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update webhook params: %w", err)
	}
	result, err := b.client.Webhooks.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkWebhookBackend) DeleteWebhook(ctx context.Context, id string) error {
	return b.client.Webhooks.Delete(ctx, modula.WebhookID(id))
}

func (b *sdkWebhookBackend) TestWebhook(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Webhooks.Test(ctx, modula.WebhookID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkWebhookBackend) ListWebhookDeliveries(ctx context.Context, webhookID string) (json.RawMessage, error) {
	result, err := b.client.Webhooks.ListDeliveries(ctx, modula.WebhookID(webhookID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkWebhookBackend) RetryWebhookDelivery(ctx context.Context, deliveryID string) error {
	return b.client.Webhooks.RetryDelivery(ctx, modula.WebhookDeliveryID(deliveryID))
}
