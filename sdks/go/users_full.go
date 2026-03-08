package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// UsersFullResource provides access to the users/full endpoints, which return
// user records with all associated data (roles, permissions, sessions, SSH keys, etc.)
// in a single response. This is an admin-only resource intended for user management
// dashboards where the full user profile is needed.
//
// The response is returned as [json.RawMessage] because the shape includes nested
// associations that vary by server configuration. Parse it into your own struct
// as needed.
//
// It is accessed via [Client].UsersFull.
type UsersFullResource struct {
	http *httpClient
}

// List returns all users with their full associated data as raw JSON.
// Requires admin permissions. The returned JSON is an array of user objects,
// each containing nested roles, permissions, sessions, and OAuth provider links.
func (u *UsersFullResource) List(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := u.http.get(ctx, "/api/v1/users/full", nil, &result); err != nil {
		return nil, fmt.Errorf("list users full: %w", err)
	}
	return result, nil
}

// Get returns a single user with full associated data as raw JSON.
// The returned JSON object contains the user's profile along with nested roles,
// permissions, sessions, SSH keys, and OAuth provider links.
// Returns an [*ApiError] with status 404 if the user does not exist.
func (u *UsersFullResource) Get(ctx context.Context, id UserID) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result json.RawMessage
	if err := u.http.get(ctx, "/api/v1/users/full/", params, &result); err != nil {
		return nil, fmt.Errorf("get user full %s: %w", string(id), err)
	}
	return result, nil
}
