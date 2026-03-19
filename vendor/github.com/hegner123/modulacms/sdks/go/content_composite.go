package modula

import (
	"context"
	"fmt"
	"net/url"
)

// ContentCompositeResource provides composite content operations that combine
// multiple steps into single API calls. These endpoints reduce round-trips for
// common multi-table workflows:
//
//   - [ContentCompositeResource.CreateWithFields] atomically creates a
//     content_data node and its associated content_field rows in one request.
//   - [ContentCompositeResource.DeleteRecursive] removes a content node and
//     all of its descendants (children, grandchildren, etc.) in one request,
//     walking the tree via sibling and child pointers.
//
// Access this resource via [Client].ContentComposite:
//
//	resp, err := client.ContentComposite.CreateWithFields(ctx, params)
type ContentCompositeResource struct {
	http *httpClient
}

// CreateWithFields creates a content_data node and its associated content_field
// rows in a single atomic request via POST /api/v1/content/create.
//
// The [ContentCreateParams].Fields map is keyed by field slug (as defined in
// the datatype's field definitions) with string values. The server creates one
// content_field row per map entry.
//
// The returned [ContentCreateResponse] includes the created content_data node,
// all successfully created fields, and counts/errors for any fields that
// failed validation.
//
// Returns an [*ApiError] if the datatype does not exist, required fields are
// missing, or the authenticated user lacks the content:create permission.
func (c *ContentCompositeResource) CreateWithFields(ctx context.Context, params ContentCreateParams) (*ContentCreateResponse, error) {
	var result ContentCreateResponse
	if err := c.http.post(ctx, "/api/v1/content/create", params, &result); err != nil {
		return nil, fmt.Errorf("content create with fields: %w", err)
	}
	return &result, nil
}

// DeleteRecursive deletes a content node and all of its descendants by walking
// the tree structure (first_child_id, next_sibling_id pointers) and removing
// every reachable node. The operation also removes all content_field rows
// associated with each deleted node.
//
// The returned [RecursiveDeleteResponse] includes the root node's ID, the
// total number of deleted nodes, and a list of all deleted IDs.
//
// This is a destructive operation. Once completed, the deleted nodes and
// their fields cannot be recovered through the API.
//
// Returns an [*ApiError] with status 404 if the root node does not exist, or
// 403 if the authenticated user lacks the content:delete permission.
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
