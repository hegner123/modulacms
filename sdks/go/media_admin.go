package modula

import (
	"context"
	"encoding/json"
	"fmt"
)

// MediaAdminResource provides media storage administration operations including
// health checks and orphan cleanup. These are admin-only operations for maintaining
// the media storage backend (S3 or local filesystem).
// It is accessed via [Client].MediaAdmin.
type MediaAdminResource struct {
	http *httpClient
}

// Health returns the media storage health status as raw JSON.
// The response includes connectivity status for the configured storage backend,
// available space, and configuration details. Use this to verify that the
// storage backend is accessible and properly configured.
func (m *MediaAdminResource) Health(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := m.http.get(ctx, "/api/v1/media/health", nil, &result); err != nil {
		return nil, fmt.Errorf("media health: %w", err)
	}
	return result, nil
}

// Cleanup scans for and removes orphaned media files that exist in storage but have
// no corresponding database record. Returns a JSON summary of files removed and
// storage space reclaimed. This is a destructive operation that cannot be undone.
func (m *MediaAdminResource) Cleanup(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := m.http.delBody(ctx, "/api/v1/media/cleanup", nil, &result); err != nil {
		return nil, fmt.Errorf("media cleanup: %w", err)
	}
	return result, nil
}
