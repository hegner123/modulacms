package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// LocaleSettingsHandler renders the locale management page.
// Always accessible so users can discover and configure i18n settings.
// CRUD mutations remain gated behind i18n being enabled.
func LocaleSettingsHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := svc.Config()
		if err != nil {
			utility.DefaultLogger.Error("failed to load config", err)
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		i18nEnabled := cfg.I18nEnabled()

		localeList, localesErr := svc.Locales.ListLocales(r.Context())
		if localesErr != nil {
			utility.DefaultLogger.Error("failed to list locales", localesErr)
			http.Error(w, "Failed to load locales", http.StatusInternalServerError)
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Locale Settings"}`)
			RenderWithOOB(w, r, pages.LocaleSettingsContent(localeList, csrfToken, i18nEnabled),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.LocaleAddDialog(localeList, csrfToken, i18nEnabled)})
			return
		}

		layout := NewAdminData(r, "Locale Settings")
		Render(w, r, pages.LocaleSettings(layout, localeList, i18nEnabled))
	}
}

// LocaleEditDialogHandler returns the edit dialog partial for a locale.
func LocaleEditDialogHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := svc.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !cfg.I18nEnabled() {
			http.Error(w, "i18n is not enabled", http.StatusNotFound)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing locale ID", http.StatusBadRequest)
			return
		}

		locale, localeErr := svc.Locales.GetLocale(r.Context(), types.LocaleID(id))
		if localeErr != nil {
			utility.DefaultLogger.Error("failed to get locale", localeErr)
			http.Error(w, "Locale not found", http.StatusNotFound)
			return
		}

		allLocales, listErr := svc.Locales.ListLocales(r.Context())
		if listErr != nil {
			utility.DefaultLogger.Error("failed to list locales", listErr)
			allLocales = []db.Locale{}
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		Render(w, r, pages.LocaleEditDialog(*locale, allLocales, csrfToken))
	}
}

// LocaleCreateHandler creates a new locale from a form submission.
// Returns the updated locale table rows as an HTMX partial.
func LocaleCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := svc.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !cfg.I18nEnabled() {
			http.Error(w, "i18n is not enabled", http.StatusNotFound)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		code := strings.TrimSpace(r.FormValue("code"))
		label := strings.TrimSpace(r.FormValue("label"))
		isDefault := r.FormValue("is_default") == "true"
		isEnabled := r.FormValue("is_enabled") == "true"
		fallbackCode := strings.TrimSpace(r.FormValue("fallback_code"))
		sortOrderStr := strings.TrimSpace(r.FormValue("sort_order"))

		sortOrder := int64(0)
		if sortOrderStr != "" {
			if n, parseIntErr := strconv.ParseInt(sortOrderStr, 10, 64); parseIntErr == nil {
				sortOrder = n
			}
		}

		ac := middleware.AuditContextFromRequest(r, *cfg)

		input := service.CreateLocaleInput{
			Code:         code,
			Label:        label,
			IsDefault:    isDefault,
			IsEnabled:    isEnabled,
			FallbackCode: fallbackCode,
			SortOrder:    sortOrder,
		}

		_, createErr := svc.Locales.CreateLocale(r.Context(), ac, input)
		if createErr != nil {
			service.HandleServiceError(w, r, createErr)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Locale created", "type": "success"}}`)
		renderLocaleTableRows(w, r, svc)
	}
}

// LocaleUpdateHandler updates an existing locale.
// Returns the updated locale table rows as an HTMX partial.
func LocaleUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := svc.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !cfg.I18nEnabled() {
			http.Error(w, "i18n is not enabled", http.StatusNotFound)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing locale ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		code := strings.TrimSpace(r.FormValue("code"))
		label := strings.TrimSpace(r.FormValue("label"))
		isDefault := r.FormValue("is_default") == "true"
		isEnabled := r.FormValue("is_enabled") == "true"
		fallbackCode := strings.TrimSpace(r.FormValue("fallback_code"))
		sortOrderStr := strings.TrimSpace(r.FormValue("sort_order"))

		// Get existing locale to preserve sort_order if not provided.
		existing, getErr := svc.Locales.GetLocale(r.Context(), types.LocaleID(id))
		if getErr != nil {
			service.HandleServiceError(w, r, getErr)
			return
		}

		sortOrder := existing.SortOrder
		if sortOrderStr != "" {
			if n, parseIntErr := strconv.ParseInt(sortOrderStr, 10, 64); parseIntErr == nil {
				sortOrder = n
			}
		}

		ac := middleware.AuditContextFromRequest(r, *cfg)

		input := service.UpdateLocaleInput{
			LocaleID:     types.LocaleID(id),
			Code:         code,
			Label:        label,
			IsDefault:    isDefault,
			IsEnabled:    isEnabled,
			FallbackCode: fallbackCode,
			SortOrder:    sortOrder,
		}

		_, updateErr := svc.Locales.UpdateLocale(r.Context(), ac, input)
		if updateErr != nil {
			service.HandleServiceError(w, r, updateErr)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Locale updated", "type": "success"}}`)
		renderLocaleTableRows(w, r, svc)
	}
}

// LocaleDeleteHandler deletes a locale by ID.
// Returns the updated locale table rows as an HTMX partial.
func LocaleDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := svc.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !cfg.I18nEnabled() {
			http.Error(w, "i18n is not enabled", http.StatusNotFound)
			return
		}

		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing locale ID", http.StatusBadRequest)
			return
		}

		ac := middleware.AuditContextFromRequest(r, *cfg)

		deleteErr := svc.Locales.DeleteLocale(r.Context(), ac, types.LocaleID(id))
		if deleteErr != nil {
			service.HandleServiceError(w, r, deleteErr)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Locale deleted", "type": "success"}}`)
		renderLocaleTableRows(w, r, svc)
	}
}

// renderLocaleTableRows loads all locales and renders the table body partial.
func renderLocaleTableRows(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	localeList, listErr := svc.Locales.ListLocales(r.Context())
	if listErr != nil {
		utility.DefaultLogger.Error("failed to list locales after mutation", listErr)
		http.Error(w, "Failed to reload locales", http.StatusInternalServerError)
		return
	}

	csrfToken := CSRFTokenFromContext(r.Context())
	Render(w, r, partials.LocaleTableRows(localeList, csrfToken))
}
