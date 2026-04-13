package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// PublishingBackend (SDK)
// ---------------------------------------------------------------------------

type sdkPublishingBackend struct {
	client *modula.Client
}

func (b *sdkPublishingBackend) PublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.PublishRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal publish params: %w", err)
	}
	result, err := b.client.Publishing.Publish(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPublishingBackend) UnpublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.PublishRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal unpublish params: %w", err)
	}
	result, err := b.client.Publishing.Unpublish(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPublishingBackend) ScheduleContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.ScheduleRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal schedule params: %w", err)
	}
	result, err := b.client.Publishing.Schedule(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPublishingBackend) AdminPublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.AdminPublishRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal admin publish params: %w", err)
	}
	result, err := b.client.AdminPublishing.AdminPublish(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPublishingBackend) AdminUnpublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.AdminPublishRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal admin unpublish params: %w", err)
	}
	result, err := b.client.AdminPublishing.AdminUnpublish(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPublishingBackend) AdminScheduleContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.AdminScheduleRequest
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal admin schedule params: %w", err)
	}
	result, err := b.client.AdminPublishing.AdminSchedule(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
