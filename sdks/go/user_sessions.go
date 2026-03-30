package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// UserSessionsResource provides access to session information for a specific user.
// It is accessed via [Client].UserSessions.
type UserSessionsResource struct {
	http *httpClient
}

// GetByUser returns session information for the given user as raw JSON.
// The response shape depends on the session data stored for the user.
// Requires sessions:read permission.
func (u *UserSessionsResource) GetByUser(ctx context.Context, userID UserID) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("q", string(userID))
	var result json.RawMessage
	if err := u.http.get(ctx, "/api/v1/users/sessions", params, &result); err != nil {
		return nil, fmt.Errorf("get user sessions %s: %w", string(userID), err)
	}
	return result, nil
}
