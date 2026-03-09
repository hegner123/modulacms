package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
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

///////////////////////////////
// HTTP HANDLERS
///////////////////////////////

// WebhookListHandler handles GET /api/v1/admin/webhooks.
func WebhookListHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	list, err := svc.Webhooks.ListWebhooks(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}

// WebhookCreateHandler handles POST /api/v1/admin/webhooks.
func WebhookCreateHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req WebhookCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	created, err := svc.Webhooks.CreateWebhook(r.Context(), ac, service.CreateWebhookInput{
		Name:     req.Name,
		URL:      req.URL,
		Secret:   req.Secret,
		Events:   req.Events,
		IsActive: req.IsActive,
		Headers:  req.Headers,
	}, user.UserID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// WebhookGetHandler handles GET /api/v1/admin/webhooks/{id}.
func WebhookGetHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractWebhookID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wh, getErr := svc.Webhooks.GetWebhook(r.Context(), id)
	if getErr != nil {
		service.HandleServiceError(w, r, getErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wh)
}

// WebhookUpdateHandler handles PUT /api/v1/admin/webhooks/{id}.
func WebhookUpdateHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
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

	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	updated, updateErr := svc.Webhooks.UpdateWebhook(r.Context(), ac, service.UpdateWebhookInput{
		WebhookID: id,
		Name:      req.Name,
		URL:       req.URL,
		Secret:    req.Secret,
		Events:    req.Events,
		IsActive:  req.IsActive,
		Headers:   req.Headers,
	})
	if updateErr != nil {
		service.HandleServiceError(w, r, updateErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// WebhookDeleteHandler handles DELETE /api/v1/admin/webhooks/{id}.
func WebhookDeleteHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractWebhookID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	if deleteErr := svc.Webhooks.DeleteWebhook(r.Context(), ac, id); deleteErr != nil {
		service.HandleServiceError(w, r, deleteErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// WebhookTestHandler handles POST /api/v1/admin/webhooks/{id}/test.
// Performs a synchronous HTTP POST to the webhook URL and returns the result.
func WebhookTestHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractWebhookID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, testErr := svc.Webhooks.TestWebhook(r.Context(), id)
	if testErr != nil {
		service.HandleServiceError(w, r, testErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// WebhookDeliveryListHandler handles GET /api/v1/admin/webhooks/{id}/deliveries.
func WebhookDeliveryListHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractWebhookID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deliveries, listErr := svc.Webhooks.ListDeliveries(r.Context(), id)
	if listErr != nil {
		service.HandleServiceError(w, r, listErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(deliveries)
}

// WebhookDeliveryRetryHandler handles POST /api/v1/admin/webhooks/deliveries/{id}/retry.
// Re-enqueues a single delivery for immediate retry via the dispatcher.
func WebhookDeliveryRetryHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	rawID := extractLastPathSegment(r.URL.Path, "retry")
	deliveryID := types.WebhookDeliveryID(rawID)
	if err := deliveryID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid delivery_id: %v", err), http.StatusBadRequest)
		return
	}

	result, retryErr := svc.Webhooks.RetryDelivery(r.Context(), deliveryID)
	if retryErr != nil {
		service.HandleServiceError(w, r, retryErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":      result.Status,
		"delivery_id": result.DeliveryID.String(),
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
