package modula

import (
	"context"
	"encoding/json"
	"fmt"
)

// MediaFullResource provides access to the composed media listing endpoint
// that returns media items with author name information.
// It is accessed via [Client].MediaFull.
type MediaFullResource struct {
	http *httpClient
}

// List returns all media items with their author names as raw JSON.
// The response is returned as [json.RawMessage] because the shape includes
// joined author data that extends the basic Media struct.
func (r *MediaFullResource) List(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := r.http.get(ctx, "/api/v1/media/full", nil, &result); err != nil {
		return nil, fmt.Errorf("list media full: %w", err)
	}
	return result, nil
}
