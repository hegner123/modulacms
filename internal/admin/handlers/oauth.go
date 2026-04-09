package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// OAuthCreateHandler handles POST /admin/oauth/{id}.
// Links an OAuth provider to the user specified by {id}.
func OAuthCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		if userID == "" {
			http.Error(w, "missing user ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		provider := strings.TrimSpace(r.FormValue("oauth_provider"))
		providerUserID := strings.TrimSpace(r.FormValue("oauth_provider_user_id"))
		accessToken := strings.TrimSpace(r.FormValue("access_token"))
		refreshToken := strings.TrimSpace(r.FormValue("refresh_token"))

		if provider == "" || providerUserID == "" || accessToken == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Provider, Provider User ID, and Access Token are required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err := svc.OAuth.CreateUserOauth(r.Context(), ac, db.CreateUserOauthParams{
			UserID:              types.NullableUserID{ID: types.UserID(userID), Valid: true},
			OauthProvider:       provider,
			OauthProviderUserID: providerUserID,
			AccessToken:         accessToken,
			RefreshToken:        refreshToken,
			TokenExpiresAt:      types.NewTimestamp(time.Now().Add(365 * 24 * time.Hour)),
			DateCreated:         types.TimestampNow(),
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create oauth connection", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to link provider", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "OAuth provider linked", "type": "success"}}`)
		renderOAuthRows(w, r, svc, userID)
	}
}

// OAuthDeleteHandler handles DELETE /admin/oauth/{id}.
// Unlinks an OAuth provider connection.
func OAuthDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		oauthID := r.PathValue("id")
		if oauthID == "" {
			http.Error(w, "OAuth connection ID required", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to unlink provider", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.OAuth.DeleteUserOauth(r.Context(), ac, types.UserOauthID(oauthID))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete oauth connection", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to unlink provider", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "OAuth provider unlinked", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// renderOAuthRows reloads and renders OAuth connection rows for a user.
func renderOAuthRows(w http.ResponseWriter, r *http.Request, svc *service.Registry, userID string) {
	var conns []db.UserOauth
	entry, err := svc.Driver().GetUserOauthByUserId(types.NullableUserID{ID: types.UserID(userID), Valid: true})
	if err == nil && entry != nil {
		conns = []db.UserOauth{*entry}
	}

	csrfToken := CSRFTokenFromContext(r.Context())
	Render(w, r, partials.UserOauthRows(conns, userID, csrfToken))
}
