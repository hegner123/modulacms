package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// LocalesHandler dispatches collection-level locale operations (no ID required).
func LocalesHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListLocales(w, r, svc)
	case http.MethodPost:
		apiCreateLocale(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// LocaleHandler dispatches item-level locale operations (ID required).
func LocaleHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetLocale(w, r, svc)
	case http.MethodPut:
		apiUpdateLocale(w, r, svc)
	case http.MethodDelete:
		apiDeleteLocale(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// LocalesPublicHandler handles GET requests for unauthenticated clients.
// Returns only enabled locales with a short cache header.
func LocalesPublicHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	locales, err := svc.Locales.ListEnabledLocales(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(locales)
}

// apiGetLocale handles GET requests for a single locale.
func apiGetLocale(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	localeID := types.LocaleID(q)
	if err := localeID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	locale, err := svc.Locales.GetLocale(r.Context(), localeID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(locale)
}

// apiListLocales handles GET requests for listing all locales.
// Supports optional enabled=true query parameter to return only enabled locales.
func apiListLocales(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	enabledParam := r.URL.Query().Get("enabled")
	if enabledParam == "true" {
		locales, err := svc.Locales.ListEnabledLocales(r.Context())
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(locales)
		return
	}

	locales, err := svc.Locales.ListLocales(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(locales)
}

// apiCreateLocale handles POST requests to create a new locale.
func apiCreateLocale(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input service.CreateLocaleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utility.DefaultLogger.Error("failed to decode locale input", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	created, err := svc.Locales.CreateLocale(r.Context(), ac, input)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// apiUpdateLocale handles PUT requests to update an existing locale.
func apiUpdateLocale(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input service.UpdateLocaleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utility.DefaultLogger.Error("failed to decode locale input", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := input.LocaleID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	updated, err := svc.Locales.UpdateLocale(r.Context(), ac, input)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// apiDeleteLocale handles DELETE requests for a locale.
func apiDeleteLocale(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	localeID := types.LocaleID(q)
	if err := localeID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := svc.Locales.DeleteLocale(r.Context(), ac, localeID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
