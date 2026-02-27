package modula

import (
	"context"
	"encoding/json"
	"fmt"
)

// MediaAdminResource provides media administration operations (health, cleanup).
type MediaAdminResource struct {
	http *httpClient
}

// Health returns the media storage health status.
func (m *MediaAdminResource) Health(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := m.http.get(ctx, "/api/v1/media/health", nil, &result); err != nil {
		return nil, fmt.Errorf("media health: %w", err)
	}
	return result, nil
}

// Cleanup runs orphaned media cleanup and returns the result.
func (m *MediaAdminResource) Cleanup(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := m.http.delBody(ctx, "/api/v1/media/cleanup", nil, &result); err != nil {
		return nil, fmt.Errorf("media cleanup: %w", err)
	}
	return result, nil
}
