package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/hegner123/modulacms/internal/webhooks"
)

// WebhookSettingsHandler renders the webhook management page.
func WebhookSettingsHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := driver.ListWebhooks()
		if err != nil {
			utility.DefaultLogger.Error("failed to list webhooks", err)
			http.Error(w, "Failed to load webhooks", http.StatusInternalServerError)
			return
		}

		webhookList := make([]db.Webhook, 0)
		if list != nil {
			webhookList = *list
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Webhook Settings"}`)
			RenderWithOOB(w, r, pages.WebhookSettingsContent(webhookList, csrfToken),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.WebhookAddDialog(csrfToken)})
			return
		}

		layout := NewAdminData(r, "Webhook Settings")
		Render(w, r, pages.WebhookSettings(layout, webhookList))
	}
}

// WebhookDetailHandler renders a webhook detail/edit page with delivery log.
func WebhookDetailHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing webhook ID", http.StatusBadRequest)
			return
		}

		wh, err := driver.GetWebhook(types.WebhookID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get webhook", err)
			http.Error(w, "Webhook not found", http.StatusNotFound)
			return
		}

		deliveries, delErr := driver.ListWebhookDeliveriesByWebhook(wh.WebhookID)
		if delErr != nil {
			utility.DefaultLogger.Error("failed to list deliveries", delErr)
			deliveries = &[]db.WebhookDelivery{}
		}

		deliveryList := make([]db.WebhookDelivery, 0)
		if deliveries != nil {
			deliveryList = *deliveries
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Webhook: `+wh.Name+`"}`)
			Render(w, r, pages.WebhookDetailContent(*wh, deliveryList, csrfToken))
			return
		}

		layout := NewAdminData(r, "Webhook: "+wh.Name)
		Render(w, r, pages.WebhookDetail(layout, *wh, deliveryList))
	}
}

// WebhookCreateHandler creates a new webhook from a form submission.
func WebhookCreateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		url := strings.TrimSpace(r.FormValue("url"))
		secret := strings.TrimSpace(r.FormValue("secret"))
		eventsRaw := strings.TrimSpace(r.FormValue("events"))
		isActive := r.FormValue("is_active") == "true"

		if name == "" || url == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Name and URL are required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if urlErr := webhooks.ValidateWebhookURL(url, cfg.WebhookAllowHTTP()); urlErr != nil {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid webhook URL", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		// Parse comma-separated events.
		events := parseCommaSeparated(eventsRaw)

		// Generate secret if empty.
		if secret == "" {
			secret = types.NewWebhookID().String() // Good enough random string
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		now := types.TimestampNow()
		_, createErr := driver.CreateWebhook(r.Context(), ac, db.CreateWebhookParams{
			Name:         name,
			URL:          url,
			Secret:       secret,
			Events:       events,
			IsActive:     isActive,
			Headers:      map[string]string{},
			AuthorID:     user.UserID,
			DateCreated:  now,
			DateModified: now,
		})
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create webhook", createErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to create webhook", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Webhook created", "type": "success"}}`)
		renderWebhookTableRows(w, r, driver)
	}
}

// WebhookUpdateHandler updates an existing webhook from a form submission.
func WebhookUpdateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing webhook ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		url := strings.TrimSpace(r.FormValue("url"))
		secret := strings.TrimSpace(r.FormValue("secret"))
		eventsRaw := strings.TrimSpace(r.FormValue("events"))
		isActive := r.FormValue("is_active") == "true"

		if name == "" || url == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Name and URL are required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if urlErr := webhooks.ValidateWebhookURL(url, cfg.WebhookAllowHTTP()); urlErr != nil {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid webhook URL", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		events := parseCommaSeparated(eventsRaw)

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		updateErr := driver.UpdateWebhook(r.Context(), ac, db.UpdateWebhookParams{
			WebhookID:    types.WebhookID(id),
			Name:         name,
			URL:          url,
			Secret:       secret,
			Events:       events,
			IsActive:     isActive,
			Headers:      map[string]string{},
			DateModified: types.TimestampNow(),
		})
		if updateErr != nil {
			utility.DefaultLogger.Error("failed to update webhook", updateErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update webhook", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Webhook updated", "type": "success"}}`)
		renderWebhookTableRows(w, r, driver)
	}
}

// WebhookDeleteHandler deletes a webhook by ID.
func WebhookDeleteHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing webhook ID", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		deleteErr := driver.DeleteWebhook(r.Context(), ac, types.WebhookID(id))
		if deleteErr != nil {
			utility.DefaultLogger.Error("failed to delete webhook", deleteErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete webhook", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Webhook deleted", "type": "success"}}`)
		renderWebhookTableRows(w, r, driver)
	}
}

// WebhookTestHandler sends a test delivery to a webhook and returns the result as JSON.
func WebhookTestHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing webhook ID", http.StatusBadRequest)
			return
		}

		wh, getErr := driver.GetWebhook(types.WebhookID(id))
		if getErr != nil {
			http.Error(w, "Webhook not found", http.StatusNotFound)
			return
		}

		// Perform a synchronous test using the REST API test handler logic.
		result := testWebhookSync(r.Context(), *wh, *cfg)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

// renderWebhookTableRows loads all webhooks and renders the table body partial.
func renderWebhookTableRows(w http.ResponseWriter, r *http.Request, driver db.DbDriver) {
	list, listErr := driver.ListWebhooks()
	if listErr != nil {
		utility.DefaultLogger.Error("failed to list webhooks after mutation", listErr)
		http.Error(w, "Failed to reload webhooks", http.StatusInternalServerError)
		return
	}

	webhookList := make([]db.Webhook, 0)
	if list != nil {
		webhookList = *list
	}

	csrfToken := CSRFTokenFromContext(r.Context())
	Render(w, r, partials.WebhookTableRows(webhookList, csrfToken))
}

// testWebhookSync performs a synchronous test POST to a webhook URL.
func testWebhookSync(ctx context.Context, wh db.Webhook, cfg config.Config) map[string]any {
	if urlErr := webhooks.ValidateWebhookURL(wh.URL, cfg.WebhookAllowHTTP()); urlErr != nil {
		return map[string]any{"status": "failed", "error": urlErr.Error(), "duration": "0s"}
	}

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
		return map[string]any{"status": "failed", "error": marshalErr.Error(), "duration": "0s"}
	}

	signature := webhooks.Sign(wh.Secret, payloadBytes)

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(payloadBytes))
	if reqErr != nil {
		return map[string]any{"status": "failed", "error": reqErr.Error(), "duration": "0s"}
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
		return map[string]any{"status": "failed", "error": doErr.Error(), "duration": duration.String()}
	}
	resp.Body.Close()

	status := "success"
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status = "failed"
	}

	return map[string]any{"status": status, "status_code": resp.StatusCode, "duration": duration.String()}
}

// parseCommaSeparated splits a comma-separated string into a trimmed slice.
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
