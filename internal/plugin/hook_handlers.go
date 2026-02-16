package plugin

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/middleware"
)

// PluginHooksListHandler returns all registered hooks with their approval status.
// GET /api/v1/admin/plugins/hooks -- any authenticated (read-only).
func PluginHooksListHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hookEngine := mgr.HookEngine()
		if hookEngine == nil {
			w.Header().Set("Content-Type", "application/json")
			// Encode error is non-recoverable (client disconnected or similar);
			// the response is already partially written so no recovery is possible.
			json.NewEncoder(w).Encode(map[string]any{"hooks": []HookRegistration{}})
			return
		}

		hooks := hookEngine.ListHooks()
		if hooks == nil {
			hooks = []HookRegistration{}
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected or similar);
		// the response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(map[string]any{"hooks": hooks})
	})
}

// PluginHooksApproveHandler approves one or more plugin hooks.
// POST /api/v1/admin/plugins/hooks/approve -- adminOnly (mutating).
// Idempotent: approving an already-approved hook is a no-op.
func PluginHooksApproveHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Body size limit: 1 MB. Must be applied before json.NewDecoder.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req struct {
			Hooks []struct {
				Plugin string `json:"plugin"`
				Event  string `json:"event"`
				Table  string `json:"table"`
			} `json:"hooks"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.Hooks) == 0 {
			http.Error(w, `"hooks" list required`, http.StatusBadRequest)
			return
		}

		hookEngine := mgr.HookEngine()
		if hookEngine == nil {
			http.Error(w, "hook engine not available", http.StatusServiceUnavailable)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		approvedBy := ""
		if user != nil {
			approvedBy = user.Username
		}

		var errs []string
		for _, hook := range req.Hooks {
			if err := hookEngine.ApproveHook(r.Context(), hook.Plugin, hook.Event, hook.Table, approvedBy); err != nil {
				errs = append(errs, err.Error())
			}
		}

		if len(errs) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			// Encode error is non-recoverable (client disconnected or similar);
			// the response is already partially written so no recovery is possible.
			json.NewEncoder(w).Encode(map[string]any{"errors": errs})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected or similar);
		// the response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
}

// PluginHooksRevokeHandler revokes approval for one or more plugin hooks.
// POST /api/v1/admin/plugins/hooks/revoke -- adminOnly (mutating).
// Idempotent: revoking an already-revoked hook is a no-op.
func PluginHooksRevokeHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Body size limit: 1 MB. Must be applied before json.NewDecoder.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req struct {
			Hooks []struct {
				Plugin string `json:"plugin"`
				Event  string `json:"event"`
				Table  string `json:"table"`
			} `json:"hooks"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.Hooks) == 0 {
			http.Error(w, `"hooks" list required`, http.StatusBadRequest)
			return
		}

		hookEngine := mgr.HookEngine()
		if hookEngine == nil {
			http.Error(w, "hook engine not available", http.StatusServiceUnavailable)
			return
		}

		var errs []string
		for _, hook := range req.Hooks {
			if err := hookEngine.RevokeHook(r.Context(), hook.Plugin, hook.Event, hook.Table); err != nil {
				errs = append(errs, err.Error())
			}
		}

		if len(errs) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			// Encode error is non-recoverable (client disconnected or similar);
			// the response is already partially written so no recovery is possible.
			json.NewEncoder(w).Encode(map[string]any{"errors": errs})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected or similar);
		// the response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
}
