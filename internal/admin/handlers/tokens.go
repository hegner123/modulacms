package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
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

// TokensListHandler lists API tokens.
// HTMX requests return partial table rows; full requests include the complete page layout.
func TokensListHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := driver.ListTokens()
		if err != nil {
			utility.DefaultLogger.Error("failed to list tokens", err)
			http.Error(w, "Failed to load tokens", http.StatusInternalServerError)
			return
		}

		var tokens []db.Tokens
		if items != nil {
			tokens = *items
		}

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "API Tokens"}`)
			Render(w, r, pages.TokensListContent(tokens, csrfToken))
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.TokensTableRows(tokens))
			return
		}

		layout := NewAdminData(r, "API Tokens")
		Render(w, r, pages.TokensList(layout, tokens))
	}
}

// TokenCreateHandler generates a random API token and stores it.
// Returns an HTMX response with the new token table rows.
func TokenCreateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenType := r.FormValue("token_type")
		if tokenType == "" {
			tokenType = "api"
		}

		// Generate random token: 32 bytes = 64 hex chars
		tokenBytes := make([]byte, 32)
		if _, randErr := rand.Read(tokenBytes); randErr != nil {
			utility.DefaultLogger.Error("failed to generate token bytes", randErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		tokenValue := hex.EncodeToString(tokenBytes)

		now := time.Now()
		expiresAt := types.NewTimestamp(now.Add(365 * 24 * time.Hour))

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		params := db.CreateTokenParams{
			UserID:    types.NullableUserID{ID: user.UserID, Valid: true},
			TokenType: tokenType,
			Token:     tokenValue,
			IssuedAt:  types.TimestampNow(),
			ExpiresAt: expiresAt,
			Revoked:   false,
		}

		_, createErr := driver.CreateToken(r.Context(), ac, params)
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create token", createErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to create token", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to create token", http.StatusInternalServerError)
			return
		}

		if IsHTMX(r) {
			// Reload the full token list
			items, listErr := driver.ListTokens()
			if listErr != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Token created but failed to reload list", "type": "warning"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			var tokens []db.Tokens
			if items != nil {
				tokens = *items
			}
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "API token created", "type": "success"}}`)
			Render(w, r, partials.TokensTableRows(tokens))
			return
		}
		http.Redirect(w, r, "/admin/users/tokens", http.StatusSeeOther)
	}
}

// TokenDeleteHandler deletes (revokes) an API token by ID.
// Only HTMX DELETE requests are supported.
func TokenDeleteHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
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

		tokenID := r.PathValue("id")
		if tokenID == "" {
			http.Error(w, "Token ID required", http.StatusBadRequest)
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

		if err := driver.DeleteToken(r.Context(), ac, tokenID); err != nil {
			utility.DefaultLogger.Error("failed to delete token", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete token", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Token revoked", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
