package modula

import (
	"context"
	"fmt"
)

// WebhookResource provides CRUD operations for webhooks, plus test and delivery management.
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

// Test sends a test event to the webhook and returns the delivery result.
func (r *WebhookResource) Test(ctx context.Context, id WebhookID) (*WebhookTestResponse, error) {
	var resp WebhookTestResponse
	if err := r.http.post(ctx, fmt.Sprintf("/api/v1/admin/webhooks/%s/test", id), nil, &resp); err != nil {
		return nil, fmt.Errorf("test webhook %s: %w", id, err)
	}
	return &resp, nil
}

// ListDeliveries returns delivery history for a webhook.
func (r *WebhookResource) ListDeliveries(ctx context.Context, id WebhookID) ([]WebhookDelivery, error) {
	var resp []WebhookDelivery
	if err := r.http.get(ctx, fmt.Sprintf("/api/v1/admin/webhooks/%s/deliveries", id), nil, &resp); err != nil {
		return nil, fmt.Errorf("list deliveries for webhook %s: %w", id, err)
	}
	return resp, nil
}

// RetryDelivery re-enqueues a failed delivery for retry.
func (r *WebhookResource) RetryDelivery(ctx context.Context, deliveryID WebhookDeliveryID) error {
	if err := r.http.post(ctx, fmt.Sprintf("/api/v1/admin/webhooks/deliveries/%s/retry", deliveryID), nil, nil); err != nil {
		return fmt.Errorf("retry delivery %s: %w", deliveryID, err)
	}
	return nil
}
