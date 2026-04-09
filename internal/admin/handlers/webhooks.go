package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// WebhookSettingsHandler renders the webhook management page.
func WebhookSettingsHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.Webhooks.ListWebhooks(r.Context())
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		webhookList := make([]db.Webhook, 0)
		if list != nil {
			webhookList = *list
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "webhook Settings"}`)
			RenderWithOOB(w, r, pages.WebhookSettingsContent(webhookList, csrfToken),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.WebhookAddDialog(csrfToken)})
			return
		}

		layout := NewAdminData(r, "webhook Settings")
		Render(w, r, pages.WebhookSettings(layout, webhookList))
	}
}

// WebhookDetailHandler renders a webhook detail/edit page with delivery log.
func WebhookDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing webhook ID", http.StatusBadRequest)
			return
		}

		wh, err := svc.Webhooks.GetWebhook(r.Context(), types.WebhookID(id))
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		deliveries, delErr := svc.Webhooks.ListDeliveries(r.Context(), wh.WebhookID)
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
func WebhookCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		url := strings.TrimSpace(r.FormValue("url"))
		secret := strings.TrimSpace(r.FormValue("secret"))
		eventsRaw := strings.TrimSpace(r.FormValue("events"))
		isActive := r.FormValue("is_active") == "true"

		events := parseCommaSeparated(eventsRaw)
		if len(events) == 0 {
			events = []string{"*"}
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

		_, createErr := svc.Webhooks.CreateWebhook(r.Context(), ac, service.CreateWebhookInput{
			Name:     name,
			URL:      url,
			Secret:   secret,
			Events:   events,
			IsActive: isActive,
			Headers:  map[string]string{},
		}, user.UserID)
		if createErr != nil {
			service.HandleServiceError(w, r, createErr)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "webhook created", "type": "success"}}`)
		renderWebhookTableRows(w, r, svc)
	}
}

// WebhookUpdateHandler updates an existing webhook from a form submission.
func WebhookUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing webhook ID", http.StatusBadRequest)
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

		events := parseCommaSeparated(eventsRaw)

		cfg, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}
		ac := middleware.AuditContextFromRequest(r, *cfg)

		_, updateErr := svc.Webhooks.UpdateWebhook(r.Context(), ac, service.UpdateWebhookInput{
			WebhookID: types.WebhookID(id),
			Name:      name,
			URL:       url,
			Secret:    secret,
			Events:    events,
			IsActive:  isActive,
			Headers:   map[string]string{},
		})
		if updateErr != nil {
			service.HandleServiceError(w, r, updateErr)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "webhook updated", "type": "success"}}`)
		renderWebhookTableRows(w, r, svc)
	}
}

// WebhookDeleteHandler deletes a webhook by ID.
func WebhookDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing webhook ID", http.StatusBadRequest)
			return
		}

		cfg, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}
		ac := middleware.AuditContextFromRequest(r, *cfg)

		deleteErr := svc.Webhooks.DeleteWebhook(r.Context(), ac, types.WebhookID(id))
		if deleteErr != nil {
			service.HandleServiceError(w, r, deleteErr)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "webhook deleted", "type": "success"}}`)
		renderWebhookTableRows(w, r, svc)
	}
}

// WebhookTestHandler sends a test delivery to a webhook and returns the result as JSON.
func WebhookTestHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing webhook ID", http.StatusBadRequest)
			return
		}

		result, err := svc.Webhooks.TestWebhook(r.Context(), types.WebhookID(id))
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

// renderWebhookTableRows loads all webhooks and renders the table body partial.
func renderWebhookTableRows(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	list, listErr := svc.Webhooks.ListWebhooks(r.Context())
	if listErr != nil {
		utility.DefaultLogger.Error("failed to list webhooks after mutation", listErr)
		http.Error(w, "failed to reload webhooks", http.StatusInternalServerError)
		return
	}

	webhookList := make([]db.Webhook, 0)
	if list != nil {
		webhookList = *list
	}

	csrfToken := CSRFTokenFromContext(r.Context())
	Render(w, r, partials.WebhookTableRows(webhookList, csrfToken))
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
