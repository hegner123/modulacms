package handlers

import (
	"net"
	"net/http"
	"strconv"
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
)

// LocaleSettingsHandler renders the locale management page.
// Always accessible so users can discover and configure i18n settings.
// CRUD mutations remain gated behind i18n being enabled.
func LocaleSettingsHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			utility.DefaultLogger.Error("failed to load config", err)
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		i18nEnabled := cfg.I18nEnabled()

		locales, localesErr := driver.ListLocales()
		if localesErr != nil {
			utility.DefaultLogger.Error("failed to list locales", localesErr)
			http.Error(w, "Failed to load locales", http.StatusInternalServerError)
			return
		}

		localeList := make([]db.Locale, 0)
		if locales != nil {
			localeList = *locales
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
func LocaleEditDialogHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
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

		locale, localeErr := driver.GetLocale(types.LocaleID(id))
		if localeErr != nil {
			utility.DefaultLogger.Error("failed to get locale", localeErr)
			http.Error(w, "Locale not found", http.StatusNotFound)
			return
		}

		allLocales, listErr := driver.ListLocales()
		if listErr != nil {
			utility.DefaultLogger.Error("failed to list locales", listErr)
			allLocales = &[]db.Locale{}
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		Render(w, r, pages.LocaleEditDialog(*locale, *allLocales, csrfToken))
	}
}

// LocaleCreateHandler creates a new locale from a form submission.
// Returns the updated locale table rows as an HTMX partial.
func LocaleCreateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
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

		if code == "" || label == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Code and Label are required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		// If setting as default, clear existing default first
		if isDefault {
			if clearErr := driver.ClearDefaultLocale(r.Context()); clearErr != nil {
				utility.DefaultLogger.Error("failed to clear default locale", clearErr)
			}
		}

		_, createErr := driver.CreateLocale(r.Context(), ac, db.CreateLocaleParams{
			Code:         code,
			Label:        label,
			IsDefault:    isDefault,
			IsEnabled:    isEnabled,
			FallbackCode: fallbackCode,
			SortOrder:    sortOrder,
			DateCreated:  types.NewTimestamp(time.Now()),
		})
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create locale", createErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to create locale", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Locale created", "type": "success"}}`)
		renderLocaleTableRows(w, r, driver)
	}
}

// LocaleUpdateHandler updates an existing locale.
// Returns the updated locale table rows as an HTMX partial.
func LocaleUpdateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
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

		existing, getErr := driver.GetLocale(types.LocaleID(id))
		if getErr != nil {
			http.Error(w, "Locale not found", http.StatusNotFound)
			return
		}

		code := strings.TrimSpace(r.FormValue("code"))
		label := strings.TrimSpace(r.FormValue("label"))
		isDefault := r.FormValue("is_default") == "true"
		isEnabled := r.FormValue("is_enabled") == "true"
		fallbackCode := strings.TrimSpace(r.FormValue("fallback_code"))
		sortOrderStr := strings.TrimSpace(r.FormValue("sort_order"))

		sortOrder := existing.SortOrder
		if sortOrderStr != "" {
			if n, parseIntErr := strconv.ParseInt(sortOrderStr, 10, 64); parseIntErr == nil {
				sortOrder = n
			}
		}

		if code == "" || label == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Code and Label are required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		// If setting as default, clear existing default first
		if isDefault && !existing.IsDefault {
			if clearErr := driver.ClearDefaultLocale(r.Context()); clearErr != nil {
				utility.DefaultLogger.Error("failed to clear default locale", clearErr)
			}
		}

		updateErr := driver.UpdateLocale(r.Context(), ac, db.UpdateLocaleParams{
			LocaleID:     types.LocaleID(id),
			Code:         code,
			Label:        label,
			IsDefault:    isDefault,
			IsEnabled:    isEnabled,
			FallbackCode: fallbackCode,
			SortOrder:    sortOrder,
			DateCreated:  existing.DateCreated,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error("failed to update locale", updateErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update locale", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Locale updated", "type": "success"}}`)
		renderLocaleTableRows(w, r, driver)
	}
}

// LocaleDeleteHandler deletes a locale by ID.
// Returns the updated locale table rows as an HTMX partial.
func LocaleDeleteHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
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

		// Prevent deleting the default locale
		locale, getErr := driver.GetLocale(types.LocaleID(id))
		if getErr != nil {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Locale not found", "type": "error"}}`)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if locale.IsDefault {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Cannot delete the default locale", "type": "error"}}`)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		deleteErr := driver.DeleteLocale(r.Context(), ac, types.LocaleID(id))
		if deleteErr != nil {
			utility.DefaultLogger.Error("failed to delete locale", deleteErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete locale", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Locale deleted", "type": "success"}}`)
		renderLocaleTableRows(w, r, driver)
	}
}

// renderLocaleTableRows loads all locales and renders the table body partial.
func renderLocaleTableRows(w http.ResponseWriter, r *http.Request, driver db.DbDriver) {
	locales, listErr := driver.ListLocales()
	if listErr != nil {
		utility.DefaultLogger.Error("failed to list locales after mutation", listErr)
		http.Error(w, "Failed to reload locales", http.StatusInternalServerError)
		return
	}

	localeList := make([]db.Locale, 0)
	if locales != nil {
		localeList = *locales
	}

	csrfToken := CSRFTokenFromContext(r.Context())
	Render(w, r, partials.LocaleTableRows(localeList, csrfToken))
}
