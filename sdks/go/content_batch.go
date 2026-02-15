package modulacms

import (
	"context"
	"encoding/json"
)

// ContentBatchResource provides atomic batch content operations.
type ContentBatchResource struct {
	http *httpClient
}

// Update performs an atomic batch content update.
// Returns json.RawMessage because the response structure varies.
func (b *ContentBatchResource) Update(ctx context.Context, req any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := b.http.post(ctx, "/api/v1/content/batch", req, &result); err != nil {
		return nil, err
	}
	return result, nil
}
