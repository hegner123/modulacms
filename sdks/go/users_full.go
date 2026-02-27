package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// UsersFullResource provides access to the users/full endpoints which return
// user records with all associated data (roles, permissions, sessions, etc.).
type UsersFullResource struct {
	http *httpClient
}

// List returns all users with their full associated data as raw JSON.
func (u *UsersFullResource) List(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := u.http.get(ctx, "/api/v1/users/full", nil, &result); err != nil {
		return nil, fmt.Errorf("list users full: %w", err)
	}
	return result, nil
}

// Get returns a single user with full associated data as raw JSON.
func (u *UsersFullResource) Get(ctx context.Context, id UserID) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result json.RawMessage
	if err := u.http.get(ctx, "/api/v1/users/full/", params, &result); err != nil {
		return nil, fmt.Errorf("get user full %s: %w", string(id), err)
	}
	return result, nil
}
