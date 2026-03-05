package modula

import (
	"context"
	"fmt"
	"net/url"
)

// ContentCompositeResource provides composite content operations that combine
// multiple steps into single API calls (e.g. create content with fields,
// recursive delete).
type ContentCompositeResource struct {
	http *httpClient
}

// CreateWithFields creates a content node and its fields in a single request.
func (c *ContentCompositeResource) CreateWithFields(ctx context.Context, params ContentCreateParams) (*ContentCreateResponse, error) {
	var result ContentCreateResponse
	if err := c.http.post(ctx, "/api/v1/content/create", params, &result); err != nil {
		return nil, fmt.Errorf("content create with fields: %w", err)
	}
	return &result, nil
}

// DeleteRecursive deletes a content node and all of its descendants.
func (c *ContentCompositeResource) DeleteRecursive(ctx context.Context, id ContentID) (*RecursiveDeleteResponse, error) {
	params := url.Values{}
	params.Set("q", string(id))
	params.Set("recursive", "true")
	var result RecursiveDeleteResponse
	if err := c.http.delBody(ctx, "/api/v1/contentdata/", params, &result); err != nil {
		return nil, fmt.Errorf("recursive delete %s: %w", string(id), err)
	}
	return &result, nil
}
