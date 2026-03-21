package modula

import (
	"context"
	"net/url"
)

// SessionsResource provides session management operations for the authenticated user.
// Sessions represent active login sessions (browser, API, SSH). Admins can view
// and manage all sessions; non-admin users can only manage their own.
// It is accessed via [Client].Sessions.
type SessionsResource struct {
	http *httpClient
}

// List returns all active sessions. For admin users this includes all sessions
// across all users; for non-admin users only their own sessions are returned.
func (s *SessionsResource) List(ctx context.Context) ([]Session, error) {
	var result []Session
	if err := s.http.get(ctx, "/api/v1/sessions", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Get returns a single session by its ID.
// Returns an [*ApiError] with status 404 if the session does not exist or
// the caller lacks permission to view it.
func (s *SessionsResource) Get(ctx context.Context, id SessionID) (*Session, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result Session
	if err := s.http.get(ctx, "/api/v1/sessions/", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update modifies an existing session's metadata and returns the updated session.
func (s *SessionsResource) Update(ctx context.Context, params UpdateSessionParams) (*Session, error) {
	var result Session
	if err := s.http.put(ctx, "/api/v1/sessions/", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Remove deletes a session by its ID, effectively logging out the associated user
// from that particular session. Deleting the caller's own session is equivalent
// to calling [AuthResource.Logout].
func (s *SessionsResource) Remove(ctx context.Context, id SessionID) error {
	params := url.Values{}
	params.Set("q", string(id))
	return s.http.del(ctx, "/api/v1/sessions/", params)
}
