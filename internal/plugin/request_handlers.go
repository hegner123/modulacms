package plugin

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/middleware"
)

// PluginRequestsListHandler returns all registered outbound domains with approval status.
// GET /api/v1/admin/plugins/requests -- any authenticated (read-only).
func PluginRequestsListHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engine := mgr.RequestEngine()
		if engine == nil {
			w.Header().Set("Content-Type", "application/json")
			// Encode error is non-recoverable (client disconnected or similar);
			// the response is already partially written so no recovery is possible.
			json.NewEncoder(w).Encode(map[string]any{"requests": []RequestRegistration{}})
			return
		}

		reqs, err := engine.ListRequests(r.Context())
		if err != nil {
			http.Error(w, "failed to list requests", http.StatusInternalServerError)
			return
		}
		if reqs == nil {
			reqs = []RequestRegistration{}
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected or similar);
		// the response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(map[string]any{"requests": reqs})
	})
}

// PluginRequestsApproveHandler approves one or more outbound request domains.
// POST /api/v1/admin/plugins/requests/approve -- adminOnly (mutating).
// Idempotent: approving an already-approved domain is a no-op.
func PluginRequestsApproveHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Body size limit: 1 MB. Must be applied before json.NewDecoder.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req struct {
			Requests []struct {
				Plugin string `json:"plugin"`
				Domain string `json:"domain"`
			} `json:"requests"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.Requests) == 0 {
			http.Error(w, `"requests" list required`, http.StatusBadRequest)
			return
		}

		engine := mgr.RequestEngine()
		if engine == nil {
			http.Error(w, "request engine not available", http.StatusServiceUnavailable)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		approvedBy := ""
		if user != nil {
			approvedBy = user.Username
		}

		var errs []string
		for _, item := range req.Requests {
			if err := engine.ApproveRequest(r.Context(), item.Plugin, item.Domain, approvedBy); err != nil {
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

// PluginRequestsRevokeHandler revokes approval for one or more outbound request domains.
// POST /api/v1/admin/plugins/requests/revoke -- adminOnly (mutating).
// Idempotent: revoking an already-revoked domain is a no-op.
func PluginRequestsRevokeHandler(mgr *Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Body size limit: 1 MB. Must be applied before json.NewDecoder.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req struct {
			Requests []struct {
				Plugin string `json:"plugin"`
				Domain string `json:"domain"`
			} `json:"requests"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.Requests) == 0 {
			http.Error(w, `"requests" list required`, http.StatusBadRequest)
			return
		}

		engine := mgr.RequestEngine()
		if engine == nil {
			http.Error(w, "request engine not available", http.StatusServiceUnavailable)
			return
		}

		var errs []string
		for _, item := range req.Requests {
			if err := engine.RevokeRequest(r.Context(), item.Plugin, item.Domain); err != nil {
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
