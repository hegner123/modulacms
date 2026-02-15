package modulacms

import (
	"context"
	"net/url"
)

// SSHKeysResource provides SSH key management operations.
type SSHKeysResource struct {
	http *httpClient
}

// List returns all SSH keys for the authenticated user.
func (s *SSHKeysResource) List(ctx context.Context) ([]SshKeyListItem, error) {
	var result []SshKeyListItem
	if err := s.http.get(ctx, "/api/v1/ssh-keys", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Create adds a new SSH key for the authenticated user.
func (s *SSHKeysResource) Create(ctx context.Context, params CreateSSHKeyParams) (*SshKey, error) {
	var result SshKey
	if err := s.http.post(ctx, "/api/v1/ssh-keys", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes an SSH key by ID.
func (s *SSHKeysResource) Delete(ctx context.Context, id UserSshKeyID) error {
	params := url.Values{}
	params.Set("q", string(id))
	return s.http.del(ctx, "/api/v1/ssh-keys/", params)
}
