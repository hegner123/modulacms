package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// PublishingBackend (Service)
// ---------------------------------------------------------------------------

type svcPublishingBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcPublishingBackend) PublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		ContentDataID string `json:"content_data_id"`
		Locale        string `json:"locale"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal publish params: %w", err)
	}
	locale := input.Locale
	if locale == "" {
		locale = "en"
	}
	result, err := b.svc.Content.Publish(ctx, b.ac, types.ContentID(input.ContentDataID), locale, b.ac.UserID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcPublishingBackend) UnpublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		ContentDataID string `json:"content_data_id"`
		Locale        string `json:"locale"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal unpublish params: %w", err)
	}
	locale := input.Locale
	if locale == "" {
		locale = "en"
	}
	if err := b.svc.Content.Unpublish(ctx, b.ac, types.ContentID(input.ContentDataID), locale, b.ac.UserID); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{
		"status":          "unpublished",
		"content_data_id": input.ContentDataID,
	})
}

func (b *svcPublishingBackend) ScheduleContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		ContentDataID string `json:"content_data_id"`
		PublishAt     string `json:"publish_at"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal schedule params: %w", err)
	}
	publishAt, err := time.Parse(time.RFC3339, input.PublishAt)
	if err != nil {
		return nil, fmt.Errorf("invalid publish_at timestamp: %w", err)
	}
	if err := b.svc.Content.Schedule(ctx, types.ContentID(input.ContentDataID), publishAt); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{
		"status":          "scheduled",
		"content_data_id": input.ContentDataID,
		"publish_at":      input.PublishAt,
	})
}

func (b *svcPublishingBackend) AdminPublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		AdminContentDataID string `json:"admin_content_data_id"`
		Locale             string `json:"locale"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal admin publish params: %w", err)
	}
	locale := input.Locale
	if locale == "" {
		locale = "en"
	}
	if err := b.svc.AdminContent.Publish(ctx, b.ac, types.AdminContentID(input.AdminContentDataID), locale, b.ac.UserID); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{
		"status":                "published",
		"admin_content_data_id": input.AdminContentDataID,
	})
}

func (b *svcPublishingBackend) AdminUnpublishContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		AdminContentDataID string `json:"admin_content_data_id"`
		Locale             string `json:"locale"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal admin unpublish params: %w", err)
	}
	locale := input.Locale
	if locale == "" {
		locale = "en"
	}
	if err := b.svc.AdminContent.Unpublish(ctx, b.ac, types.AdminContentID(input.AdminContentDataID), locale, b.ac.UserID); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{
		"status":                "unpublished",
		"admin_content_data_id": input.AdminContentDataID,
	})
}

func (b *svcPublishingBackend) AdminScheduleContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		AdminContentDataID string `json:"admin_content_data_id"`
		PublishAt          string `json:"publish_at"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal admin schedule params: %w", err)
	}
	publishAt, err := time.Parse(time.RFC3339, input.PublishAt)
	if err != nil {
		return nil, fmt.Errorf("invalid publish_at timestamp: %w", err)
	}
	if err := b.svc.AdminContent.Schedule(ctx, types.AdminContentID(input.AdminContentDataID), publishAt); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{
		"status":                "scheduled",
		"admin_content_data_id": input.AdminContentDataID,
		"publish_at":            input.PublishAt,
	})
}
