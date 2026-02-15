package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// PluginListHandler returns all loaded plugins with their state and metadata.
// GET /api/v1/admin/plugins -- any authenticated (read-only, S2).
func PluginListHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plugins := mgr.ListPlugins()

		type pluginJSON struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			State       string `json:"state"`
			CBState     string `json:"circuit_breaker_state,omitempty"`
		}

		result := make([]pluginJSON, 0, len(plugins))
		for _, inst := range plugins {
			p := pluginJSON{
				Name:        inst.Info.Name,
				Version:     inst.Info.Version,
				Description: inst.Info.Description,
				State:       inst.State.String(),
			}
			if inst.CB != nil {
				p.CBState = inst.CB.State().String()
			}
			result = append(result, p)
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected or similar);
		// the response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(map[string]any{"plugins": result})
	})
}

// PluginInfoHandler returns detailed information about a single plugin.
// GET /api/v1/admin/plugins/{name} -- any authenticated (read-only, S2).
func PluginInfoHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.Error(w, "plugin name required", http.StatusBadRequest)
			return
		}

		inst := mgr.GetPlugin(name)
		if inst == nil {
			http.Error(w, "plugin not found", http.StatusNotFound)
			return
		}

		type infoJSON struct {
			Name          string       `json:"name"`
			Version       string       `json:"version"`
			Description   string       `json:"description"`
			Author        string       `json:"author,omitempty"`
			License       string       `json:"license,omitempty"`
			State         string       `json:"state"`
			FailedReason  string       `json:"failed_reason,omitempty"`
			CBState       string       `json:"circuit_breaker_state,omitempty"`
			CBErrors      int          `json:"circuit_breaker_errors,omitempty"`
			VMsAvailable  int          `json:"vms_available"`
			VMsTotal      int          `json:"vms_total"`
			Dependencies  []string     `json:"dependencies,omitempty"`
			SchemaDrift   []DriftEntry `json:"schema_drift,omitempty"`
		}

		info := infoJSON{
			Name:         inst.Info.Name,
			Version:      inst.Info.Version,
			Description:  inst.Info.Description,
			Author:       inst.Info.Author,
			License:      inst.Info.License,
			State:        inst.State.String(),
			FailedReason: inst.FailedReason,
			Dependencies: inst.Info.Dependencies,
			SchemaDrift:  inst.SchemaDrift, // S7: surfaced in admin response
		}

		if inst.CB != nil {
			info.CBState = inst.CB.State().String()
			info.CBErrors = inst.CB.ConsecutiveErrors()
		}

		if inst.Pool != nil {
			info.VMsAvailable = inst.Pool.AvailableCount()
			info.VMsTotal = inst.Pool.PoolSize()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})
}

// PluginReloadHandler triggers a reload of a specific plugin.
// POST /api/v1/admin/plugins/{name}/reload -- adminOnly (mutating, S2).
// S9: Enforces 10s per-plugin cooldown between reloads.
func PluginReloadHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.Error(w, "plugin name required", http.StatusBadRequest)
			return
		}

		inst := mgr.GetPlugin(name)
		if inst == nil {
			http.Error(w, "plugin not found", http.StatusNotFound)
			return
		}

		// S9: Check reload cooldown.
		if mgr.watcher != nil && mgr.watcher.IsPluginCooldownActive(name) {
			w.Header().Set("Retry-After", "10")
			http.Error(w, "reload cooldown active (10s)", http.StatusTooManyRequests)
			return
		}

		err := mgr.ReloadPlugin(r.Context(), name)
		if err != nil {
			utility.DefaultLogger.Error(
				fmt.Sprintf("admin reload for plugin %q failed: %s", name, err.Error()), nil)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Record cooldown.
		if mgr.watcher != nil {
			mgr.watcher.SetReloadCooldown(name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "plugin": name})
	})
}

// PluginEnableHandler re-enables a disabled plugin (resets CB and reloads).
// POST /api/v1/admin/plugins/{name}/enable -- adminOnly (S2, S5).
func PluginEnableHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.Error(w, "plugin name required", http.StatusBadRequest)
			return
		}

		inst := mgr.GetPlugin(name)
		if inst == nil {
			http.Error(w, "plugin not found", http.StatusNotFound)
			return
		}

		// Extract admin user from request context for audit trail (S5).
		adminUser := ""
		user := middleware.AuthenticatedUser(r.Context())
		if user != nil {
			adminUser = user.Username
		}

		err := mgr.EnablePlugin(r.Context(), name, adminUser)
		if err != nil {
			utility.DefaultLogger.Error(
				fmt.Sprintf("admin enable for plugin %q failed: %s", name, err.Error()), nil)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "plugin": name, "state": "running"})
	})
}

// PluginDisableHandler disables a running plugin.
// POST /api/v1/admin/plugins/{name}/disable -- adminOnly (S2).
func PluginDisableHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.Error(w, "plugin name required", http.StatusBadRequest)
			return
		}

		err := mgr.DisablePlugin(r.Context(), name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "plugin": name, "state": "stopped"})
	})
}

// PluginCleanupListHandler returns orphaned plugin tables (dry-run).
// GET /api/v1/admin/plugins/cleanup -- adminOnly (S2).
func PluginCleanupListHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orphaned, err := mgr.ListOrphanedTables(r.Context())
		if err != nil {
			utility.DefaultLogger.Error(
				fmt.Sprintf("listing orphaned tables failed: %s", err.Error()), nil)
			http.Error(w, "failed to list orphaned tables", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"orphaned_tables": orphaned,
			"count":           len(orphaned),
			"action":          "dry_run",
		})
	})
}

// PluginCleanupDropHandler drops orphaned plugin tables.
// POST /api/v1/admin/plugins/cleanup -- adminOnly (destructive, S1, S2).
// Requires {"confirm": true, "tables": [...]} in body.
func PluginCleanupDropHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Confirm bool     `json:"confirm"`
			Tables  []string `json:"tables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if !req.Confirm {
			http.Error(w, `{"confirm": true} required to drop tables`, http.StatusBadRequest)
			return
		}

		if len(req.Tables) == 0 {
			http.Error(w, `"tables" list required`, http.StatusBadRequest)
			return
		}

		dropped, err := mgr.DropOrphanedTables(r.Context(), req.Tables)
		if err != nil {
			utility.DefaultLogger.Error(
				fmt.Sprintf("dropping orphaned tables failed: %s", err.Error()), nil)
			http.Error(w, "failed to drop orphaned tables", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"dropped": dropped,
			"count":   len(dropped),
		})
	})
}
