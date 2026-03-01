package router

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/hegner123/modulacms/internal/webhooks"
)

///////////////////////////////
// REQUEST / RESPONSE TYPES
///////////////////////////////

// WebhookCreateRequest is the JSON body for POST /api/v1/admin/webhooks.
type WebhookCreateRequest struct {
	Name     string            `json:"name"`
	URL      string            `json:"url"`
	Secret   string            `json:"secret"`
	Events   []string          `json:"events"`
	IsActive bool              `json:"is_active"`
	Headers  map[string]string `json:"headers"`
}

// WebhookUpdateRequest is the JSON body for PUT /api/v1/admin/webhooks/{id}.
type WebhookUpdateRequest struct {
	Name     string            `json:"name"`
	URL      string            `json:"url"`
	Secret   string            `json:"secret"`
	Events   []string          `json:"events"`
	IsActive bool              `json:"is_active"`
	Headers  map[string]string `json:"headers"`
}

// WebhookTestResponse is returned from POST /api/v1/admin/webhooks/{id}/test.
type WebhookTestResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
	Duration   string `json:"duration"`
}

///////////////////////////////
// HTTP HANDLERS
///////////////////////////////

// WebhookListHandler handles GET /api/v1/admin/webhooks.
func WebhookListHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	d := db.ConfigDB(c)

	list, err := d.ListWebhooks()
	if err != nil {
		utility.DefaultLogger.Error("list webhooks failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}

// WebhookCreateHandler handles POST /api/v1/admin/webhooks.
func WebhookCreateHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req WebhookCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	// Validate URL against SSRF rules.
	if err := webhooks.ValidateWebhookURL(req.URL, c.WebhookAllowHTTP()); err != nil {
		http.Error(w, fmt.Sprintf("invalid webhook URL: %v", err), http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	// Generate a secret if none provided.
	secret := req.Secret
	if secret == "" {
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			utility.DefaultLogger.Error("failed to generate webhook secret", err)
			http.Error(w, "failed to generate secret", http.StatusInternalServerError)
			return
		}
		secret = hex.EncodeToString(secretBytes)
	}

	if req.Events == nil {
		req.Events = []string{}
	}
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}

	d := db.ConfigDB(c)
	ac := middleware.AuditContextFromRequest(r, c)
	now := types.TimestampNow()

	created, err := d.CreateWebhook(r.Context(), ac, db.CreateWebhookParams{
		Name:         req.Name,
		URL:          req.URL,
		Secret:       secret,
		Events:       req.Events,
		IsActive:     req.IsActive,
		Headers:      req.Headers,
		AuthorID:     user.UserID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		utility.DefaultLogger.Error("create webhook failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// WebhookGetHandler handles GET /api/v1/admin/webhooks/{id}.
func WebhookGetHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	id, err := extractWebhookID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d := db.ConfigDB(c)
	wh, err := d.GetWebhook(id)
	if err != nil {
		utility.DefaultLogger.Error("get webhook failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wh)
}

// WebhookUpdateHandler handles PUT /api/v1/admin/webhooks/{id}.
func WebhookUpdateHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	id, err := extractWebhookID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req WebhookUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	// Validate URL against SSRF rules.
	if err := webhooks.ValidateWebhookURL(req.URL, c.WebhookAllowHTTP()); err != nil {
		http.Error(w, fmt.Sprintf("invalid webhook URL: %v", err), http.StatusBadRequest)
		return
	}

	if req.Events == nil {
		req.Events = []string{}
	}
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}

	d := db.ConfigDB(c)
	ac := middleware.AuditContextFromRequest(r, c)

	updateErr := d.UpdateWebhook(r.Context(), ac, db.UpdateWebhookParams{
		WebhookID:    id,
		Name:         req.Name,
		URL:          req.URL,
		Secret:       req.Secret,
		Events:       req.Events,
		IsActive:     req.IsActive,
		Headers:      req.Headers,
		DateModified: types.TimestampNow(),
	})
	if updateErr != nil {
		utility.DefaultLogger.Error("update webhook failed", updateErr)
		http.Error(w, updateErr.Error(), http.StatusInternalServerError)
		return
	}

	updated, err := d.GetWebhook(id)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated webhook", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// WebhookDeleteHandler handles DELETE /api/v1/admin/webhooks/{id}.
func WebhookDeleteHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	id, err := extractWebhookID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d := db.ConfigDB(c)
	ac := middleware.AuditContextFromRequest(r, c)

	if err := d.DeleteWebhook(r.Context(), ac, id); err != nil {
		utility.DefaultLogger.Error("delete webhook failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// WebhookTestHandler handles POST /api/v1/admin/webhooks/{id}/test.
// Performs a synchronous HTTP POST to the webhook URL and returns the result.
func WebhookTestHandler(w http.ResponseWriter, r *http.Request, c config.Config, dispatcher publishing.WebhookDispatcher) {
	id, err := extractWebhookID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d := db.ConfigDB(c)
	wh, err := d.GetWebhook(id)
	if err != nil {
		utility.DefaultLogger.Error("get webhook for test failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate URL at test time too.
	if urlErr := webhooks.ValidateWebhookURL(wh.URL, c.WebhookAllowHTTP()); urlErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WebhookTestResponse{
			Status:   "failed",
			Error:    fmt.Sprintf("URL validation failed: %v", urlErr),
			Duration: "0s",
		})
		return
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
		utility.DefaultLogger.Error("failed to marshal test payload", marshalErr)
		http.Error(w, "failed to marshal test payload", http.StatusInternalServerError)
		return
	}

	signature := webhooks.Sign(wh.Secret, payloadBytes)

	req, reqErr := http.NewRequestWithContext(r.Context(), http.MethodPost, wh.URL, bytes.NewReader(payloadBytes))
	if reqErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WebhookTestResponse{
			Status:   "failed",
			Error:    fmt.Sprintf("request creation failed: %v", reqErr),
			Duration: "0s",
		})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ModulaCMS-Signature", signature)
	req.Header.Set("X-ModulaCMS-Event", "webhook.test")
	req.Header.Set("User-Agent", "ModulaCMS-Webhook/1.0")
	for k, v := range wh.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: time.Duration(c.WebhookTimeout()) * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	start := time.Now()
	resp, doErr := client.Do(req)
	duration := time.Since(start)

	if doErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WebhookTestResponse{
			Status:   "failed",
			Error:    fmt.Sprintf("HTTP request failed: %v", doErr),
			Duration: duration.String(),
		})
		return
	}
	resp.Body.Close()

	status := "success"
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status = "failed"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(WebhookTestResponse{
		Status:     status,
		StatusCode: resp.StatusCode,
		Duration:   duration.String(),
	})
}

// WebhookDeliveryListHandler handles GET /api/v1/admin/webhooks/{id}/deliveries.
func WebhookDeliveryListHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	id, err := extractWebhookID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d := db.ConfigDB(c)
	deliveries, err := d.ListWebhookDeliveriesByWebhook(id)
	if err != nil {
		utility.DefaultLogger.Error("list webhook deliveries failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(deliveries)
}

// WebhookDeliveryRetryHandler handles POST /api/v1/admin/webhooks/deliveries/{id}/retry.
// Re-enqueues a single delivery for immediate retry via the dispatcher.
func WebhookDeliveryRetryHandler(w http.ResponseWriter, r *http.Request, c config.Config, dispatcher publishing.WebhookDispatcher) {
	rawID := extractLastPathSegment(r.URL.Path, "retry")
	deliveryID := types.WebhookDeliveryID(rawID)
	if err := deliveryID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid delivery_id: %v", err), http.StatusBadRequest)
		return
	}

	d := db.ConfigDB(c)
	del, err := d.GetWebhookDelivery(deliveryID)
	if err != nil {
		utility.DefaultLogger.Error("get webhook delivery for retry failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reset status to retrying so the dispatcher picks it up.
	updateErr := d.UpdateWebhookDeliveryStatus(r.Context(), db.UpdateWebhookDeliveryStatusParams{
		Status:      db.DeliveryStatusRetrying,
		Attempts:    del.Attempts,
		NextRetryAt: time.Now().UTC().Format(time.RFC3339),
		DeliveryID:  del.DeliveryID,
	})
	if updateErr != nil {
		utility.DefaultLogger.Error("update delivery status for retry failed", updateErr)
		http.Error(w, updateErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "retrying",
		"delivery_id": del.DeliveryID.String(),
	})
}

///////////////////////////////
// HELPERS
///////////////////////////////

// extractWebhookID extracts the webhook ID from the URL path.
// It looks for a 26-character ULID segment in the path after "webhooks/".
func extractWebhookID(r *http.Request) (types.WebhookID, error) {
	path := r.URL.Path
	idx := strings.Index(path, "webhooks/")
	if idx == -1 {
		return "", fmt.Errorf("webhook ID not found in path")
	}
	rest := path[idx+len("webhooks/"):]
	// Take up to the next slash or end of string.
	if slashIdx := strings.IndexByte(rest, '/'); slashIdx != -1 {
		rest = rest[:slashIdx]
	}
	id := types.WebhookID(rest)
	if err := id.Validate(); err != nil {
		return "", fmt.Errorf("invalid webhook_id: %w", err)
	}
	return id, nil
}

// extractLastPathSegment returns the path segment immediately before the given suffix.
// For example, extractLastPathSegment("/a/b/ULID/retry", "retry") returns "ULID".
func extractLastPathSegment(path, suffix string) string {
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimSuffix(path, suffix)
	path = strings.TrimSuffix(path, "/")
	idx := strings.LastIndexByte(path, '/')
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}
