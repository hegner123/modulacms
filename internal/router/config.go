package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/service"
)

// ConfigGetHandler returns the current configuration (redacted).
// Supports optional ?category= query parameter to filter by category.
func ConfigGetHandler(svc *service.Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		category := r.URL.Query().Get("category")
		if category != "" {
			result, err := svc.ConfigSvc.GetConfigByCategory(category)
			if err != nil {
				service.HandleServiceError(w, r, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			// Encode error is non-recoverable (client disconnected);
			// response is already partially written so no recovery is possible.
			json.NewEncoder(w).Encode(result)
			return
		}

		data, err := svc.ConfigSvc.GetConfig()
		if err != nil {
			http.Error(w, "failed to load config", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
}

// ConfigUpdateHandler applies a partial update to the configuration.
// Expects a JSON object body with fields to change.
func ConfigUpdateHandler(svc *service.Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limit request body to 1 MB.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var updates map[string]any
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		result, err := svc.ConfigSvc.UpdateConfig(updates)
		if err != nil {
			var cve *service.ConfigValidationError
			if errors.As(err, &cve) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				// Encode error is non-recoverable (client disconnected);
				// response is already partially written so no recovery is possible.
				json.NewEncoder(w).Encode(map[string]any{
					"ok":     false,
					"errors": cve.Errors,
				})
				return
			}
			service.HandleServiceError(w, r, err)
			return
		}

		redactedBytes, marshalErr := json.Marshal(result.Config)
		if marshalErr != nil {
			http.Error(w, "failed to marshal config", http.StatusInternalServerError)
			return
		}

		response := map[string]any{
			"ok":               true,
			"config":           json.RawMessage(redactedBytes),
			"restart_required": result.RestartRequired,
			"warnings":         result.Warnings,
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected);
		// response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(response)
	})
}

// ConfigSearchIndexHandler returns the full search index of all config fields.
// Combines FieldRegistry metadata with rich help text from HELP_TEXT.md.
func ConfigSearchIndexHandler() http.Handler {
	// Build once at handler creation time — the index is static.
	index := config.BuildSearchIndex()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected);
		// response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(index)
	})
}

// ConfigMetaHandler returns the field metadata registry.
func ConfigMetaHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type fieldJSON struct {
			JSONKey       string `json:"json_key"`
			Label         string `json:"label"`
			Category      string `json:"category"`
			HotReloadable bool   `json:"hot_reloadable"`
			Sensitive     bool   `json:"sensitive"`
			Required      bool   `json:"required"`
			Description   string `json:"description"`
		}

		fields := make([]fieldJSON, 0, len(config.FieldRegistry))
		for _, f := range config.FieldRegistry {
			fields = append(fields, fieldJSON{
				JSONKey:       f.JSONKey,
				Label:         f.Label,
				Category:      string(f.Category),
				HotReloadable: f.HotReloadable,
				Sensitive:     f.Sensitive,
				Required:      f.Required,
				Description:   f.Description,
			})
		}

		type categoryJSON struct {
			Key   string `json:"key"`
			Label string `json:"label"`
		}

		categories := make([]categoryJSON, 0)
		for _, c := range config.AllCategories() {
			categories = append(categories, categoryJSON{
				Key:   string(c),
				Label: config.CategoryLabel(c),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected);
		// response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(map[string]any{
			"fields":     fields,
			"categories": categories,
		})
	})
}
