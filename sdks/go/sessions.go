package modulacms

import (
	"context"
	"net/url"
)

// SessionsResource provides session management operations.
type SessionsResource struct {
	http *httpClient
}

// Update updates an existing session.
func (s *SessionsResource) Update(ctx context.Context, params UpdateSessionParams) (*Session, error) {
	var result Session
	if err := s.http.put(ctx, "/api/v1/sessions/", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Remove deletes a session by ID.
func (s *SessionsResource) Remove(ctx context.Context, id SessionID) error {
	params := url.Values{}
	params.Set("q", string(id))
	return s.http.del(ctx, "/api/v1/sessions/", params)
}
