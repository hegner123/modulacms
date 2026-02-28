package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
)

// RichtextToolbarHandler returns the global richtext toolbar configuration as a JSON array.
// GET /admin/api/config/richtext-toolbar
func RichtextToolbarHandler(mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected or similar);
		// the response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(cfg.RichtextToolbar())
	}
}
