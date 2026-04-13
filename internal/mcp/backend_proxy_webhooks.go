package mcp

import (
	"context"
	"encoding/json"
)

type proxyWebhookBackend struct{ p *proxyBackends }

func (b *proxyWebhookBackend) ListWebhooks(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Webhooks.ListWebhooks(ctx)
}

func (b *proxyWebhookBackend) GetWebhook(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Webhooks.GetWebhook(ctx, id)
}

func (b *proxyWebhookBackend) CreateWebhook(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Webhooks.CreateWebhook(ctx, params)
}

func (b *proxyWebhookBackend) UpdateWebhook(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Webhooks.UpdateWebhook(ctx, params)
}

func (b *proxyWebhookBackend) DeleteWebhook(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Webhooks.DeleteWebhook(ctx, id)
}

func (b *proxyWebhookBackend) TestWebhook(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Webhooks.TestWebhook(ctx, id)
}

func (b *proxyWebhookBackend) ListWebhookDeliveries(ctx context.Context, webhookID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Webhooks.ListWebhookDeliveries(ctx, webhookID)
}

func (b *proxyWebhookBackend) RetryWebhookDelivery(ctx context.Context, deliveryID string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Webhooks.RetryWebhookDelivery(ctx, deliveryID)
}
