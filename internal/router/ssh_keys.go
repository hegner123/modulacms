package router

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// AddSSHKeyHandler handles POST /api/v1/ssh-keys
// Allows authenticated users to add an SSH public key to their account.
func AddSSHKeyHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	authUser := middleware.AuthenticatedUser(r.Context())
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		PublicKey string `json:"public_key"`
		Label     string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cfg, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	input := service.AddSSHKeyInput{
		UserID:    authUser.UserID,
		PublicKey: req.PublicKey,
		Label:     req.Label,
	}

	created, err := svc.SSHKeys.AddKey(r.Context(), ac, input)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// ListSSHKeysHandler handles GET /api/v1/ssh-keys
// Returns all SSH keys for the authenticated user.
// Supports optional fingerprint query parameter to return a single key matching the fingerprint.
func ListSSHKeysHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	authUser := middleware.AuthenticatedUser(r.Context())
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// If fingerprint query param is provided, return the matching key.
	fingerprintParam := r.URL.Query().Get("fingerprint")
	if fingerprintParam != "" {
		sshKey, err := svc.SSHKeys.GetKeyByFingerprint(r.Context(), fingerprintParam)
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		// Only return the key if it belongs to the authenticated user.
		if !sshKey.UserID.Valid || sshKey.UserID.ID != authUser.UserID {
			http.Error(w, "SSH key not found", http.StatusNotFound)
			return
		}

		writeJSON(w, sshKey)
		return
	}

	keys, err := svc.SSHKeys.ListKeys(r.Context(), types.NullableUserID{ID: authUser.UserID, Valid: true})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	// Return SSH keys (without full public key for security)
	type SSHKeyResponse struct {
		SSHKEY_ID   string `json:"ssh_key_id"`
		KeyType     string `json:"key_type"`
		Fingerprint string `json:"fingerprint"`
		Label       string `json:"label"`
		DateCreated string `json:"date_created"`
		LastUsed    string `json:"last_used"`
	}

	response := make([]SSHKeyResponse, len(*keys))
	for i, key := range *keys {
		response[i] = SSHKeyResponse{
			SSHKEY_ID:   key.SshKeyID,
			KeyType:     key.KeyType,
			Fingerprint: key.Fingerprint,
			Label:       key.Label,
			DateCreated: key.DateCreated.String(),
			LastUsed:    key.LastUsed,
		}
	}

	writeJSON(w, response)
}

// DeleteSSHKeyHandler handles DELETE /api/v1/ssh-keys/:id
// Allows authenticated users to delete their own SSH keys.
func DeleteSSHKeyHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	authUser := middleware.AuthenticatedUser(r.Context())
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get SSH key ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}

	keyID := pathParts[3]
	if keyID == "" {
		http.Error(w, "Invalid SSH key ID", http.StatusBadRequest)
		return
	}

	cfg, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	if err := svc.SSHKeys.DeleteKey(r.Context(), ac, authUser.UserID, keyID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
