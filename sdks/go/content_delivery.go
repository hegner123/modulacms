package modulacms

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// ContentDeliveryResource provides public content delivery via slug-based routing.
type ContentDeliveryResource struct {
	http *httpClient
}

// GetPage retrieves a page by its slug. The format parameter controls
// the response structure (e.g. "clean", "contentful", "sanity", etc.).
// Returns json.RawMessage because the output structure varies by format.
func (c *ContentDeliveryResource) GetPage(ctx context.Context, slug string, format string) (json.RawMessage, error) {
	params := url.Values{}
	if format != "" {
		params.Set("format", format)
	}
	path := "/" + strings.TrimLeft(slug, "/")
	var result json.RawMessage
	if err := c.http.get(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return result, nil
}
