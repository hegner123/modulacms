package modula

import (
	"context"
	"encoding/json"
	"net/url"
)

// AdminTreeResource provides access to the admin content tree, which represents
// the hierarchical structure of all content in the CMS. The tree includes draft
// and working content (unlike the public content delivery endpoints, which only
// serve published content).
// It is accessed via [Client].AdminTree.
type AdminTreeResource struct {
	http *httpClient
}

// Get retrieves the content tree rooted at the given datatype slug.
// The format parameter controls the response shape: common values include
// "nested" (recursive tree), "flat" (list with parent pointers), and "" (server default).
// Returns [json.RawMessage] because the output structure varies by format.
// Parse the returned JSON into your own structs based on the requested format.
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
