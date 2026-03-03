package router

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// LocalesHandler dispatches collection-level locale operations (no ID required).
func LocalesHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListLocales(w, r, c)
	case http.MethodPost:
		apiCreateLocale(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// LocaleHandler dispatches item-level locale operations (ID required).
func LocaleHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetLocale(w, r, c)
	case http.MethodPut:
		apiUpdateLocale(w, r, c)
	case http.MethodDelete:
		apiDeleteLocale(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// LocalesPublicHandler handles GET requests for unauthenticated clients.
// Returns only enabled locales with a short cache header.
func LocalesPublicHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	d := db.ConfigDB(c)

	locales, err := d.ListEnabledLocales()
	if err != nil {
		utility.DefaultLogger.Error("failed to list enabled locales", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(locales)
}

// apiGetLocale handles GET requests for a single locale.
func apiGetLocale(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	localeID := types.LocaleID(q)
	if err := localeID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	locale, err := d.GetLocale(localeID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(locale)
	return nil
}

// apiListLocales handles GET requests for listing all locales.
// Supports optional enabled=true query parameter to return only enabled locales.
func apiListLocales(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	enabledParam := r.URL.Query().Get("enabled")
	if enabledParam == "true" {
		locales, err := d.ListEnabledLocales()
		if err != nil {
			utility.DefaultLogger.Error("", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(locales)
		return nil
	}

	locales, err := d.ListLocales()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(locales)
	return nil
}

// apiCreateLocale handles POST requests to create a new locale.
func apiCreateLocale(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var params db.CreateLocaleParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if err := validateLocaleCode(params.Code); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if params.FallbackCode != "" {
		if err := validateFallbackChain(d, params.FallbackCode, params.Code); err != nil {
			utility.DefaultLogger.Error("", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return err
		}
	}

	if params.IsDefault {
		if err := d.ClearDefaultLocale(r.Context()); err != nil {
			utility.DefaultLogger.Error("failed to clear default locale", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
	}

	ac := middleware.AuditContextFromRequest(r, c)
	created, err := d.CreateLocale(r.Context(), ac, params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
	return nil
}

// apiUpdateLocale handles PUT requests to update an existing locale.
func apiUpdateLocale(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var params db.UpdateLocaleParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if err := params.LocaleID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if err := validateLocaleCode(params.Code); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if params.FallbackCode != "" {
		if err := validateFallbackChain(d, params.FallbackCode, params.Code); err != nil {
			utility.DefaultLogger.Error("", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return err
		}
	}

	if params.IsDefault {
		if err := d.ClearDefaultLocale(r.Context()); err != nil {
			utility.DefaultLogger.Error("failed to clear default locale", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
	}

	ac := middleware.AuditContextFromRequest(r, c)
	err = d.UpdateLocale(r.Context(), ac, params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetLocale(params.LocaleID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated locale", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteLocale handles DELETE requests for a locale.
func apiDeleteLocale(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	localeID := types.LocaleID(q)
	if err := localeID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Prevent deleting the default locale.
	existing, err := d.GetLocale(localeID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	if existing.IsDefault {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  "forbidden",
			"detail": "cannot delete the default locale",
		})
		return nil
	}

	ac := middleware.AuditContextFromRequest(r, c)
	err = d.DeleteLocale(r.Context(), ac, localeID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// validateLocaleCode validates a locale code against a basic BCP 47 pattern:
// 2-5 lowercase letters optionally followed by a hyphen and 2-3 alphanumeric characters.
// Uses character-by-character validation -- no regex.
func validateLocaleCode(code string) error {
	if len(code) == 0 {
		return fmt.Errorf("locale code must not be empty")
	}

	// Find the hyphen position, if any.
	hyphenIdx := -1
	for i := range len(code) {
		if code[i] == '-' {
			hyphenIdx = i
			break
		}
	}

	var primary, subtag string
	if hyphenIdx == -1 {
		primary = code
	} else {
		primary = code[:hyphenIdx]
		subtag = code[hyphenIdx+1:]
	}

	// Primary subtag: 2-5 lowercase ASCII letters.
	if len(primary) < 2 || len(primary) > 5 {
		return fmt.Errorf("locale code primary subtag must be 2-5 lowercase letters, got %q", primary)
	}
	for i := range len(primary) {
		ch := primary[i]
		if ch < 'a' || ch > 'z' {
			return fmt.Errorf("locale code primary subtag must contain only lowercase letters, found %q at position %d", string(ch), i)
		}
	}

	// If there is a hyphen, the subtag is required: 2-3 alphanumeric characters.
	if hyphenIdx != -1 {
		if len(subtag) < 2 || len(subtag) > 3 {
			return fmt.Errorf("locale code subtag must be 2-3 alphanumeric characters, got %q", subtag)
		}
		for i := range len(subtag) {
			ch := subtag[i]
			isLower := ch >= 'a' && ch <= 'z'
			isUpper := ch >= 'A' && ch <= 'Z'
			isDigit := ch >= '0' && ch <= '9'
			if !isLower && !isUpper && !isDigit {
				return fmt.Errorf("locale code subtag must contain only alphanumeric characters, found %q at position %d", string(ch), i)
			}
		}
	}

	return nil
}

// validateFallbackChain verifies that fallbackCode references an existing locale
// and that the chain does not form a cycle. Walks at most 2 hops.
func validateFallbackChain(d db.DbDriver, fallbackCode string, selfCode string) error {
	if fallbackCode == selfCode {
		return fmt.Errorf("locale cannot fall back to itself")
	}

	// Hop 1: the direct fallback must exist.
	fb1, err := d.GetLocaleByCode(fallbackCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("fallback locale %q does not exist", fallbackCode)
		}
		// Wrapped errors from the db layer: check if the underlying cause is not-found.
		return fmt.Errorf("failed to look up fallback locale %q: %w", fallbackCode, err)
	}

	// Hop 2: if the direct fallback itself has a fallback, check it does not point back to selfCode.
	if fb1.FallbackCode == "" {
		return nil
	}
	if fb1.FallbackCode == selfCode {
		return fmt.Errorf("fallback chain cycle detected: %s -> %s -> %s", selfCode, fallbackCode, selfCode)
	}

	// Verify the second-hop locale exists.
	_, err = d.GetLocaleByCode(fb1.FallbackCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("second-hop fallback locale %q (via %q) does not exist", fb1.FallbackCode, fallbackCode)
		}
		return fmt.Errorf("failed to look up second-hop fallback locale %q: %w", fb1.FallbackCode, err)
	}

	return nil
}
