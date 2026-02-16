package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
)

// ConfigGetHandler returns the current configuration (redacted).
// Supports optional ?category= query parameter to filter by category.
func ConfigGetHandler(mgr *config.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		redacted := config.RedactedConfig(*cfg)

		category := r.URL.Query().Get("category")
		if category != "" {
			fields := config.FieldsByCategory(config.FieldCategory(category))
			if len(fields) == 0 {
				http.Error(w, "unknown category", http.StatusBadRequest)
				return
			}

			result := make(map[string]any)
			result["category"] = category
			fieldValues := make(map[string]string)
			for _, f := range fields {
				fieldValues[f.JSONKey] = config.ConfigFieldString(redacted, f.JSONKey)
			}
			result["fields"] = fieldValues

			w.Header().Set("Content-Type", "application/json")
			// Encode error is non-recoverable (client disconnected);
			// response is already partially written so no recovery is possible.
			json.NewEncoder(w).Encode(result)
			return
		}

		data, err := config.RedactedJSON(*cfg)
		if err != nil {
			http.Error(w, "failed to marshal config", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
}

// ConfigUpdateHandler applies a partial update to the configuration.
// Expects a JSON object body with fields to change.
func ConfigUpdateHandler(mgr *config.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limit request body to 1 MB.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var updates map[string]any
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(updates) == 0 {
			http.Error(w, "no updates provided", http.StatusBadRequest)
			return
		}

		result, err := mgr.Update(updates)
		if err != nil && !result.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			// Encode error is non-recoverable (client disconnected);
			// response is already partially written so no recovery is possible.
			json.NewEncoder(w).Encode(map[string]any{
				"ok":     false,
				"errors": result.Errors,
			})
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "failed to read updated config", http.StatusInternalServerError)
			return
		}

		redacted := config.RedactedConfig(*cfg)
		redactedBytes, marshalErr := json.Marshal(redacted)
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
