package handlers

import (
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// UserSSHKeyAddHandler handles POST /admin/users/{id}/ssh-keys.
// Adds an SSH key to the user and returns updated table rows.
func UserSSHKeyAddHandler(svc *service.Registry) http.HandlerFunc {
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

		label := strings.TrimSpace(r.FormValue("label"))
		publicKey := strings.TrimSpace(r.FormValue("public_key"))

		if publicKey == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "public key is required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err := svc.SSHKeys.AddKey(r.Context(), ac, service.AddSSHKeyInput{
			UserID:    types.UserID(userID),
			PublicKey: publicKey,
			Label:     label,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to add ssh key", err)
			service.HandleServiceError(w, r, err)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "SSH key added", "type": "success"}}`)
		renderSSHKeysRows(w, r, svc, userID)
	}
}

// UserSSHKeyDeleteHandler handles DELETE /admin/users/{id}/ssh-keys/{keyId}.
// Deletes the SSH key and returns a 200 for HTMX row removal.
func UserSSHKeyDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.PathValue("id")
		keyID := r.PathValue("keyId")
		if userID == "" || keyID == "" {
			http.Error(w, "missing user ID or key ID", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete SSH key", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.SSHKeys.DeleteKey(r.Context(), ac, types.UserID(userID), keyID)
		if err != nil {
			utility.DefaultLogger.Error("failed to delete ssh key", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete SSH key", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "SSH key deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// renderSSHKeysRows reloads and renders the SSH keys table rows for a user.
func renderSSHKeysRows(w http.ResponseWriter, r *http.Request, svc *service.Registry, userID string) {
	keyList, err := svc.SSHKeys.ListKeys(r.Context(), types.NullableUserID{ID: types.UserID(userID), Valid: true})
	if err != nil {
		utility.DefaultLogger.Error("failed to reload ssh keys", err)
		http.Error(w, "failed to reload SSH keys", http.StatusInternalServerError)
		return
	}

	var keys []db.UserSshKeys
	if keyList != nil {
		keys = *keyList
	}

	csrfToken := CSRFTokenFromContext(r.Context())
	Render(w, r, partials.UserSshKeysRows(keys, userID, csrfToken))
}
