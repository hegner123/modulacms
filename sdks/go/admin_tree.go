package modulacms

import (
	"context"
	"encoding/json"
	"net/url"
)

// AdminTreeResource provides access to the admin content tree.
type AdminTreeResource struct {
	http *httpClient
}

// Get retrieves the content tree for a given slug and format.
// Returns json.RawMessage because the output structure varies by format.
func (a *AdminTreeResource) Get(ctx context.Context, slug string, format string) (json.RawMessage, error) {
	params := url.Values{}
	if format != "" {
		params.Set("format", format)
	}
	var result json.RawMessage
	if err := a.http.get(ctx, "/api/v1/admin/tree/"+slug, params, &result); err != nil {
		return nil, err
	}
	return result, nil
}
