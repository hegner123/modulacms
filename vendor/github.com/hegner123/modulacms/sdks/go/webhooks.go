package modula

import (
	"context"
	"fmt"
)

// WebhookResource provides CRUD operations for webhooks, plus test firing and
// delivery history management. Webhooks are event-driven HTTP callbacks that notify
// external services when CMS events occur (content created, published, deleted, etc.).
//
// WebhookResource embeds [Resource] for standard List, Get, Create, Update, Delete
// operations, and adds webhook-specific methods for testing and delivery management.
// It is accessed via [Client].Webhooks.
type WebhookResource struct {
	*Resource[Webhook, CreateWebhookRequest, UpdateWebhookRequest, WebhookID]
	http *httpClient
}

func newWebhookResource(h *httpClient) *WebhookResource {
	return &WebhookResource{
		Resource: newResource[Webhook, CreateWebhookRequest, UpdateWebhookRequest, WebhookID](h, "/api/v1/admin/webhooks"),
		http:     h,
	}
}

// Test sends a synthetic test event to the webhook's URL and returns the delivery result.
// Use this to verify that the webhook endpoint is reachable and responds correctly.
// The test event uses a special event type that the receiver can identify as a test.
func (r *WebhookResource) Test(ctx context.Context, id WebhookID) (*WebhookTestResponse, error) {
	var resp WebhookTestResponse
	if err := r.http.post(ctx, fmt.Sprintf("/api/v1/admin/webhooks/%s/test", id), nil, &resp); err != nil {
		return nil, fmt.Errorf("test webhook %s: %w", id, err)
	}
	return &resp, nil
}

// ListDeliveries returns the delivery history for a webhook, including both successful
// and failed attempts. Each [WebhookDelivery] contains the HTTP status code, response body,
// and timing information for the delivery attempt.
func (r *WebhookResource) ListDeliveries(ctx context.Context, id WebhookID) ([]WebhookDelivery, error) {
	var resp []WebhookDelivery
	if err := r.http.get(ctx, fmt.Sprintf("/api/v1/admin/webhooks/%s/deliveries", id), nil, &resp); err != nil {
		return nil, fmt.Errorf("list deliveries for webhook %s: %w", id, err)
	}
	return resp, nil
}

// RetryDelivery re-enqueues a failed delivery for another attempt. The original
// event payload is resent to the webhook URL. Only failed deliveries can be retried;
// attempting to retry a successful delivery returns an error.
func (r *WebhookResource) RetryDelivery(ctx context.Context, deliveryID WebhookDeliveryID) error {
	if err := r.http.post(ctx, fmt.Sprintf("/api/v1/admin/webhooks/deliveries/%s/retry", deliveryID), nil, nil); err != nil {
		return fmt.Errorf("retry delivery %s: %w", deliveryID, err)
	}
	return nil
}
