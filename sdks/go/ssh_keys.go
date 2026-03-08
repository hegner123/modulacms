package modula

import (
	"context"
	"fmt"
	"net/url"
)

// SSHKeysResource provides SSH key management operations for the authenticated user.
// SSH keys are used to authenticate with the ModulaCMS SSH server, which provides
// access to the Bubbletea TUI for content management. Each key is stored with its
// public key material and fingerprint for identification.
// It is accessed via [Client].SSHKeys.
type SSHKeysResource struct {
	http *httpClient
}

// List returns all SSH keys registered for the authenticated user.
// Each [SshKeyListItem] contains the key's metadata (name, fingerprint, creation date)
// but not the full public key material.
func (s *SSHKeysResource) List(ctx context.Context) ([]SshKeyListItem, error) {
	var result []SshKeyListItem
	if err := s.http.get(ctx, "/api/v1/ssh-keys", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Create registers a new SSH public key for the authenticated user.
// The [CreateSSHKeyParams] must include the public key in OpenSSH authorized_keys format.
// Returns the created [SshKey] with its server-assigned ID and computed fingerprint.
func (s *SSHKeysResource) Create(ctx context.Context, params CreateSSHKeyParams) (*SshKey, error) {
	var result SshKey
	if err := s.http.post(ctx, "/api/v1/ssh-keys", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByFingerprint returns a single SSH key matching the given fingerprint string.
// The fingerprint should be in the standard SSH format (e.g., "SHA256:...").
// Returns an [*ApiError] with status 404 if no key matches the fingerprint.
func (s *SSHKeysResource) GetByFingerprint(ctx context.Context, fingerprint string) (*SshKeyListItem, error) {
	params := url.Values{}
	params.Set("fingerprint", fingerprint)
	var result SshKeyListItem
	if err := s.http.get(ctx, "/api/v1/ssh-keys", params, &result); err != nil {
		return nil, fmt.Errorf("get ssh key by fingerprint %s: %w", fingerprint, err)
	}
	return &result, nil
}

// Delete removes an SSH key by its ID. After deletion, the key can no longer
// be used to authenticate SSH connections to the CMS server.
func (s *SSHKeysResource) Delete(ctx context.Context, id UserSshKeyID) error {
	params := url.Values{}
	params.Set("q", string(id))
	return s.http.del(ctx, "/api/v1/ssh-keys/", params)
}
