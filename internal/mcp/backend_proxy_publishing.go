package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// PublishingBackend (Proxy)
// ---------------------------------------------------------------------------

type proxyPublishingBackend struct{ p *proxyBackends }

func (b *proxyPublishingBackend) PublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Publishing.PublishContent(ctx, params)
}

func (b *proxyPublishingBackend) UnpublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Publishing.UnpublishContent(ctx, params)
}

func (b *proxyPublishingBackend) ScheduleContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Publishing.ScheduleContent(ctx, params)
}

func (b *proxyPublishingBackend) AdminPublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Publishing.AdminPublishContent(ctx, params)
}

func (b *proxyPublishingBackend) AdminUnpublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Publishing.AdminUnpublishContent(ctx, params)
}

func (b *proxyPublishingBackend) AdminScheduleContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Publishing.AdminScheduleContent(ctx, params)
}
