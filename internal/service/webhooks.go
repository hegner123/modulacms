package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/webhooks"
)

// WebhookService encapsulates webhook CRUD, SSRF validation, secret generation,
// test dispatch, delivery tracking, and old delivery pruning.
type WebhookService struct {
	driver     db.DbDriver
	mgr        *config.Manager
	dispatcher publishing.WebhookDispatcher
}

// NewWebhookService creates a WebhookService.
func NewWebhookService(driver db.DbDriver, mgr *config.Manager, dispatcher publishing.WebhookDispatcher) *WebhookService {
	return &WebhookService{driver: driver, mgr: mgr, dispatcher: dispatcher}
}

///////////////////////////////
// INPUT / OUTPUT TYPES
///////////////////////////////

// CreateWebhookInput holds parameters for creating a webhook.
type CreateWebhookInput struct {
	Name     string
	URL      string
	Secret   string // empty -> auto-generate
	Events   []string
	IsActive bool
	Headers  map[string]string
}

// UpdateWebhookInput holds parameters for updating a webhook.
type UpdateWebhookInput struct {
	WebhookID types.WebhookID
	Name      string
	URL       string
	Secret    string
	Events    []string
	IsActive  bool
	Headers   map[string]string
}

// WebhookTestResult holds the outcome of a synchronous webhook test delivery.
type WebhookTestResult struct {
	Status     string // "success" or "failed"
	StatusCode int
	Error      string
	Duration   string
}

// DeliveryRetryResult holds the outcome of re-enqueuing a delivery for retry.
type DeliveryRetryResult struct {
	DeliveryID types.WebhookDeliveryID
	Status     string
}

///////////////////////////////
// SERVICE METHODS
///////////////////////////////

// CreateWebhook validates input, auto-generates a secret if needed, and creates a webhook.
func (s *WebhookService) CreateWebhook(ctx context.Context, ac audited.AuditContext, input CreateWebhookInput, authorID types.UserID) (*db.Webhook, error) {
	ve := &ValidationError{}

	if input.Name == "" {
		ve.Add("name", "is required")
	}
	if input.URL == "" {
		ve.Add("url", "is required")
	}
	if len(input.Events) == 0 {
		ve.Add("events", "at least one event is required")
	}
	if ve.HasErrors() {
		return nil, ve
	}

	// SSRF validation.
	cfg, cfgErr := s.mgr.Config()
	if cfgErr != nil {
		return nil, &InternalError{Err: fmt.Errorf("load config: %w", cfgErr)}
	}
	if err := webhooks.ValidateWebhookURL(input.URL, cfg.WebhookAllowHTTP()); err != nil {
		return nil, NewValidationError("url", fmt.Sprintf("invalid webhook URL: %v", err))
	}

	// Auto-generate secret if empty.
	secret := input.Secret
	if secret == "" {
		secretBytes := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, secretBytes); err != nil {
			return nil, &InternalError{Err: fmt.Errorf("generate secret: %w", err)}
		}
		secret = hex.EncodeToString(secretBytes)
	}

	events := input.Events
	if events == nil {
		events = []string{}
	}
	headers := input.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	now := types.TimestampNow()
	created, err := s.driver.CreateWebhook(ctx, ac, db.CreateWebhookParams{
		Name:         input.Name,
		URL:          input.URL,
		Secret:       secret,
		Events:       events,
		IsActive:     input.IsActive,
		Headers:      headers,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("create webhook: %w", err)}
	}

	return created, nil
}

// UpdateWebhook validates input, re-checks SSRF on URL change, preserves
// immutable fields, and updates a webhook.
func (s *WebhookService) UpdateWebhook(ctx context.Context, ac audited.AuditContext, input UpdateWebhookInput) (*db.Webhook, error) {
	existing, err := s.driver.GetWebhook(input.WebhookID)
	if err != nil {
		return nil, &NotFoundError{Resource: "webhook", ID: input.WebhookID.String()}
	}

	ve := &ValidationError{}
	if input.Name == "" {
		ve.Add("name", "is required")
	}
	if input.URL == "" {
		ve.Add("url", "is required")
	}
	if len(input.Events) == 0 {
		ve.Add("events", "at least one event is required")
	}
	if ve.HasErrors() {
		return nil, ve
	}

	// SSRF re-check on URL change.
	cfg, cfgErr := s.mgr.Config()
	if cfgErr != nil {
		return nil, &InternalError{Err: fmt.Errorf("load config: %w", cfgErr)}
	}
	if err := webhooks.ValidateWebhookURL(input.URL, cfg.WebhookAllowHTTP()); err != nil {
		return nil, NewValidationError("url", fmt.Sprintf("invalid webhook URL: %v", err))
	}

	events := input.Events
	if events == nil {
		events = []string{}
	}
	headers := input.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	// Preserve immutable fields from existing record.
	updateErr := s.driver.UpdateWebhook(ctx, ac, db.UpdateWebhookParams{
		WebhookID:    input.WebhookID,
		Name:         input.Name,
		URL:          input.URL,
		Secret:       input.Secret,
		Events:       events,
		IsActive:     input.IsActive,
		Headers:      headers,
		DateCreated:  existing.DateCreated,
		DateModified: types.TimestampNow(),
	})
	if updateErr != nil {
		return nil, &InternalError{Err: fmt.Errorf("update webhook: %w", updateErr)}
	}

	updated, err := s.driver.GetWebhook(input.WebhookID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch updated webhook: %w", err)}
	}

	return updated, nil
}

// DeleteWebhook removes a webhook by ID.
func (s *WebhookService) DeleteWebhook(ctx context.Context, ac audited.AuditContext, id types.WebhookID) error {
	if err := id.Validate(); err != nil {
		return NewValidationError("webhook_id", fmt.Sprintf("invalid: %v", err))
	}

	if err := s.driver.DeleteWebhook(ctx, ac, id); err != nil {
		return &InternalError{Err: fmt.Errorf("delete webhook: %w", err)}
	}

	return nil
}

// GetWebhook retrieves a webhook by ID. Returns NotFoundError if not found.
func (s *WebhookService) GetWebhook(ctx context.Context, id types.WebhookID) (*db.Webhook, error) {
	wh, err := s.driver.GetWebhook(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "webhook", ID: id.String()}
	}
	return wh, nil
}

// ListWebhooks returns all webhooks.
func (s *WebhookService) ListWebhooks(ctx context.Context) (*[]db.Webhook, error) {
	list, err := s.driver.ListWebhooks()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list webhooks: %w", err)}
	}
	return list, nil
}

// ListWebhooksPaginated returns webhooks with pagination and total count.
func (s *WebhookService) ListWebhooksPaginated(ctx context.Context, limit, offset int64) (*[]db.Webhook, *int64, error) {
	items, err := s.driver.ListWebhooksPaginated(db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, nil, &InternalError{Err: fmt.Errorf("list webhooks paginated: %w", err)}
	}
	total, err := s.driver.CountWebhooks()
	if err != nil {
		return nil, nil, &InternalError{Err: fmt.Errorf("count webhooks: %w", err)}
	}
	return items, total, nil
}

// TestWebhook performs a synchronous HTTP POST to the webhook URL and returns the result.
// Does NOT create a delivery record.
func (s *WebhookService) TestWebhook(ctx context.Context, id types.WebhookID) (*WebhookTestResult, error) {
	wh, err := s.driver.GetWebhook(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "webhook", ID: id.String()}
	}

	// SSRF re-check at test time (DNS rebinding defense).
	cfg, cfgErr := s.mgr.Config()
	if cfgErr != nil {
		return nil, &InternalError{Err: fmt.Errorf("load config: %w", cfgErr)}
	}
	if urlErr := webhooks.ValidateWebhookURL(wh.URL, cfg.WebhookAllowHTTP()); urlErr != nil {
		return &WebhookTestResult{
			Status:   "failed",
			Error:    fmt.Sprintf("URL validation failed: %v", urlErr),
			Duration: "0s",
		}, nil
	}

	// Build a test payload.
	testPayload := webhooks.Payload{
		ID:         types.NewWebhookDeliveryID().String(),
		Event:      "webhook.test",
		OccurredAt: time.Now().UTC(),
		Data: map[string]any{
			"webhook_id": wh.WebhookID.String(),
			"message":    "This is a test delivery from ModulaCMS.",
		},
	}

	payloadBytes, marshalErr := json.Marshal(testPayload)
	if marshalErr != nil {
		return nil, &InternalError{Err: fmt.Errorf("marshal test payload: %w", marshalErr)}
	}

	signature := webhooks.Sign(wh.Secret, payloadBytes)

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(payloadBytes))
	if reqErr != nil {
		return &WebhookTestResult{
			Status:   "failed",
			Error:    fmt.Sprintf("request creation failed: %v", reqErr),
			Duration: "0s",
		}, nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ModulaCMS-Signature", signature)
	req.Header.Set("X-ModulaCMS-Event", "webhook.test")
	req.Header.Set("User-Agent", "ModulaCMS-Webhook/1.0")
	for k, v := range wh.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.WebhookTimeout()) * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	start := time.Now()
	resp, doErr := client.Do(req)
	duration := time.Since(start)

	if doErr != nil {
		return &WebhookTestResult{
			Status:   "failed",
			Error:    fmt.Sprintf("HTTP request failed: %v", doErr),
			Duration: duration.String(),
		}, nil
	}
	resp.Body.Close()

	status := "success"
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status = "failed"
	}

	return &WebhookTestResult{
		Status:     status,
		StatusCode: resp.StatusCode,
		Duration:   duration.String(),
	}, nil
}

// ListDeliveries returns all deliveries for a specific webhook.
func (s *WebhookService) ListDeliveries(ctx context.Context, webhookID types.WebhookID) (*[]db.WebhookDelivery, error) {
	deliveries, err := s.driver.ListWebhookDeliveriesByWebhook(webhookID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list deliveries: %w", err)}
	}
	return deliveries, nil
}

// RetryDelivery re-enqueues a delivery for immediate retry via the dispatcher.
// Returns a DeliveryRetryResult with status "queued" -- the actual delivery
// result is tracked asynchronously.
func (s *WebhookService) RetryDelivery(ctx context.Context, deliveryID types.WebhookDeliveryID) (*DeliveryRetryResult, error) {
	del, err := s.driver.GetWebhookDelivery(deliveryID)
	if err != nil {
		return nil, &NotFoundError{Resource: "webhook_delivery", ID: deliveryID.String()}
	}

	// Verify the parent webhook exists.
	_, whErr := s.driver.GetWebhook(del.WebhookID)
	if whErr != nil {
		return nil, &NotFoundError{Resource: "webhook", ID: del.WebhookID.String()}
	}

	// Reset status to retrying so the dispatcher picks it up.
	updateErr := s.driver.UpdateWebhookDeliveryStatus(ctx, db.UpdateWebhookDeliveryStatusParams{
		Status:      db.DeliveryStatusRetrying,
		Attempts:    del.Attempts,
		NextRetryAt: time.Now().UTC().Format(time.RFC3339),
		DeliveryID:  del.DeliveryID,
	})
	if updateErr != nil {
		return nil, &InternalError{Err: fmt.Errorf("update delivery status for retry: %w", updateErr)}
	}

	return &DeliveryRetryResult{
		DeliveryID: deliveryID,
		Status:     "queued",
	}, nil
}

// PruneDeliveries removes completed deliveries older than the given time.
func (s *WebhookService) PruneDeliveries(ctx context.Context, olderThan time.Time) error {
	if err := s.driver.PruneOldDeliveries(ctx, types.NewTimestamp(olderThan)); err != nil {
		return &InternalError{Err: fmt.Errorf("prune deliveries: %w", err)}
	}
	return nil
}
