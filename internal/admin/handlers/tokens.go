package handlers

import (
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// TokensListHandler lists API tokens.
// HTMX requests return partial table rows; full requests include the complete page layout.
func TokensListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := svc.Tokens.ListTokens(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list tokens", err)
			http.Error(w, "Failed to load tokens", http.StatusInternalServerError)
			return
		}

		var tokens []db.Tokens
		if items != nil {
			tokens = *items
		}

		userIDs := make([]types.NullableUserID, len(tokens))
		for i, t := range tokens {
			userIDs[i] = t.UserID
		}
		userNames := BuildUserNameMap(svc.Driver(), userIDs)

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "API Tokens"}`)
			Render(w, r, pages.TokensListContent(tokens, csrfToken, userNames))
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.TokensTableRows(tokens, userNames))
			return
		}

		layout := NewAdminData(r, "API Tokens")
		Render(w, r, pages.TokensList(layout, tokens, userNames))
	}
}

// TokenDetailHandler shows token metadata.
func TokenDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenID := r.PathValue("id")
		if tokenID == "" {
			http.Error(w, "Token ID required", http.StatusBadRequest)
			return
		}

		token, err := svc.Tokens.GetToken(r.Context(), tokenID)
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		userName := ResolveNullableUserName(svc.Driver(), token.UserID)

		layout := NewAdminData(r, "Token")
		RenderNav(w, r, "Token",
			pages.TokenDetailContent(*token, userName),
			pages.TokenDetail(layout, *token, userName))
	}
}

// TokenCreateHandler generates a random API token and stores it.
// Returns an HTMX response with the new token table rows.
func TokenCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := svc.Config()
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

		expiresAt := types.NewTimestamp(time.Now().Add(365 * 24 * time.Hour))

		input := service.CreateTokenInput{
			UserID:    types.NullableUserID{ID: user.UserID, Valid: true},
			TokenType: tokenType,
			Expiry:    expiresAt,
		}

		ac := middleware.AuditContextFromRequest(r, *cfg)
		result, createErr := svc.Tokens.CreateToken(r.Context(), ac, input)
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
			items, listErr := svc.Tokens.ListTokens(r.Context())
			if listErr != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Token created but failed to reload list", "type": "warning"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			var tokens []db.Tokens
			if items != nil {
				tokens = *items
			}
			userIDs := make([]types.NullableUserID, len(tokens))
			for i, t := range tokens {
				userIDs[i] = t.UserID
			}
			userNames := BuildUserNameMap(svc.Driver(), userIDs)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Token created", "type": "success"}}`)
			RenderWithOOB(w, r, partials.TokensTableRows(tokens, userNames),
				OOBSwap{TargetID: "admin-dialogs", Component: partials.TokenRevealDialog(result.RawToken)})
			return
		}
		http.Redirect(w, r, "/admin/users/tokens", http.StatusSeeOther)
	}
}

// TokenDeleteHandler deletes (revokes) an API token by ID.
// Only HTMX DELETE requests are supported.
func TokenDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := svc.Config()
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

		ac := middleware.AuditContextFromRequest(r, *cfg)

		if err := svc.Tokens.DeleteToken(r.Context(), ac, tokenID); err != nil {
			utility.DefaultLogger.Error("failed to delete token", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete token", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Token revoked", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
