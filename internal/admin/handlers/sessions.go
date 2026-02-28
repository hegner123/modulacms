package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// SessionsListHandler lists all active sessions.
// HTMX requests return partial table rows; full requests include the complete page layout.
func SessionsListHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := driver.ListSessions()
		if err != nil {
			utility.DefaultLogger.Error("failed to list sessions", err)
			http.Error(w, "Failed to load sessions", http.StatusInternalServerError)
			return
		}

		var sessions []db.Sessions
		if items != nil {
			sessions = *items
		}

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Sessions"}`)
			Render(w, r, pages.SessionsListContent(sessions))
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.SessionsTableRows(sessions))
			return
		}

		layout := NewAdminData(r, "Sessions")
		Render(w, r, pages.SessionsList(layout, sessions))
	}
}

// SessionDeleteHandler revokes a session by ID.
// Only HTMX DELETE requests are supported.
func SessionDeleteHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionID := r.PathValue("id")
		if sessionID == "" {
			http.Error(w, "Session ID required", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		if err := driver.DeleteSession(r.Context(), ac, types.SessionID(sessionID)); err != nil {
			utility.DefaultLogger.Error("failed to delete session", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to revoke session", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Session revoked", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
