package router

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// authcontext is used to retrieve user from context (must match middleware type)
type authcontext string

// AddSSHKeyHandler handles POST /api/v1/ssh-keys
// Allows authenticated users to add an SSH public key to their account.
func AddSSHKeyHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	// Get authenticated user from context
	var authCtx authcontext = "authenticated"
	userVal := r.Context().Value(authCtx)
	if userVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	authUser, ok := userVal.(*db.Users)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req struct {
		PublicKey string `json:"public_key"`
		Label     string `json:"label"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and parse SSH public key
	keyType, fingerprint, err := middleware.ParseSSHPublicKey(req.PublicKey)
	if err != nil {
		utility.DefaultLogger.Error("Invalid SSH public key", err)
		http.Error(w, "Invalid SSH public key format", http.StatusBadRequest)
		return
	}

	// Check if key already exists
	dbc := db.ConfigDB(c)
	existing, _ := dbc.GetUserSshKeyByFingerprint(fingerprint)
	if existing != nil {
		http.Error(w, "SSH key already registered", http.StatusConflict)
		return
	}

	// Create SSH key record
	sshKey, err := dbc.CreateUserSshKey(db.CreateUserSshKeyParams{
		UserID:      types.NullableUserID{ID: authUser.UserID, Valid: true},
		PublicKey:   req.PublicKey,
		KeyType:     keyType,
		Fingerprint: fingerprint,
		Label:       req.Label,
		DateCreated: types.TimestampNow(),
	})

	if err != nil {
		utility.DefaultLogger.Error("Failed to create SSH key", err)
		http.Error(w, "Failed to add SSH key", http.StatusInternalServerError)
		return
	}

	utility.DefaultLogger.Info("SSH key added for user %d: %s", authUser.UserID, fingerprint)

	// Return created SSH key
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sshKey)
}

// ListSSHKeysHandler handles GET /api/v1/ssh-keys
// Returns all SSH keys for the authenticated user.
func ListSSHKeysHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	// Get authenticated user from context
	var authCtx authcontext = "authenticated"
	userVal := r.Context().Value(authCtx)
	if userVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	authUser, ok := userVal.(*db.Users)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	// Get user's SSH keys
	dbc := db.ConfigDB(c)
	keys, err := dbc.ListUserSshKeys(types.NullableUserID{ID: authUser.UserID, Valid: true})
	if err != nil {
		utility.DefaultLogger.Error("Failed to list SSH keys", err)
		http.Error(w, "Failed to retrieve SSH keys", http.StatusInternalServerError)
		return
	}

	// Return SSH keys (without full public key for security)
	type SSHKeyResponse struct {
		SSHKEY_ID   int64  `json:"ssh_key_id"`
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteSSHKeyHandler handles DELETE /api/v1/ssh-keys/:id
// Allows authenticated users to delete their own SSH keys.
func DeleteSSHKeyHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	// Get authenticated user from context
	var authCtx authcontext = "authenticated"
	userVal := r.Context().Value(authCtx)
	if userVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	authUser, ok := userVal.(*db.Users)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	// Get SSH key ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}

	keyID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		http.Error(w, "Invalid SSH key ID", http.StatusBadRequest)
		return
	}

	// Verify the key belongs to the authenticated user
	dbc := db.ConfigDB(c)
	sshKey, err := dbc.GetUserSshKey(keyID)
	if err != nil {
		http.Error(w, "SSH key not found", http.StatusNotFound)
		return
	}

	if !sshKey.UserID.Valid || sshKey.UserID.ID != authUser.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete the SSH key
	err = dbc.DeleteUserSshKey(keyID)
	if err != nil {
		utility.DefaultLogger.Error("Failed to delete SSH key", err)
		http.Error(w, "Failed to delete SSH key", http.StatusInternalServerError)
		return
	}

	utility.DefaultLogger.Info("SSH key deleted for user %d: %d", authUser.UserID, keyID)

	w.WriteHeader(http.StatusNoContent)
}
