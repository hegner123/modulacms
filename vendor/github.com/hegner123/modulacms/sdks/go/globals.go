package modula

import (
	"context"
	"encoding/json"
)

// GlobalsResource provides access to global content trees via
// GET /api/v1/globals. Global content items are _global-typed root nodes
// whose published trees are available site-wide (e.g. navigation, footer,
// site settings).
//
// Access this resource via [Client].Globals:
//
//	globals, err := client.Globals.List(ctx)
//
// The response is returned as [json.RawMessage] because each entry contains
// a recursively nested content tree whose shape depends on the datatype
// schema configuration.
type GlobalsResource struct {
	http *httpClient
}

// List returns all published global content trees.
//
// Each entry in the returned array contains a content_data_id, datatype
// metadata, and a fully composed content tree. The array is empty (not nil)
// when no global content exists.
//
// Returns [json.RawMessage] because the tree structure is recursive and
// schema-dependent. The caller is responsible for unmarshalling into an
// appropriate struct or inspecting the raw bytes.
func (g *GlobalsResource) List(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := g.http.get(ctx, "/api/v1/globals", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
