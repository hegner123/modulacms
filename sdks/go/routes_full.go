package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// RoutesFullResource provides access to the composed route endpoint that returns
// a route with its content tree. The response shape depends on the content
// structure and datatype configuration for the route.
// It is accessed via [Client].RoutesFull.
type RoutesFullResource struct {
	http *httpClient
}

// GetFull returns a route with its associated content tree as raw JSON.
// The response is returned as [json.RawMessage] because the nested content
// tree structure varies by route configuration and content schema.
func (r *RoutesFullResource) GetFull(ctx context.Context, id RouteID) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result json.RawMessage
	if err := r.http.get(ctx, "/api/v1/routes/full", params, &result); err != nil {
		return nil, fmt.Errorf("get route full %s: %w", string(id), err)
	}
	return result, nil
}
