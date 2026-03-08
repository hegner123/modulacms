package modula

import (
	"context"
	"encoding/json"
)

// ContentBatchResource provides atomic batch content operations via
// POST /api/v1/content/batch.
//
// Batch operations allow multiple content mutations (creates, updates, deletes)
// to be submitted in a single HTTP request and executed atomically on the
// server. This is useful for editors that buffer changes locally and flush
// them in one round-trip.
//
// For tree-structure changes (inserting, moving, and deleting nodes with
// automatic pointer rewiring), prefer [ContentTreeResource].Save instead,
// which provides a higher-level API purpose-built for tree manipulation.
//
// Access this resource via [Client].ContentBatch:
//
//	resp, err := client.ContentBatch.Update(ctx, batchPayload)
type ContentBatchResource struct {
	http *httpClient
}

// Update performs an atomic batch content update by POSTing the given request
// body to the batch endpoint.
//
// The req parameter is serialized as JSON and sent directly to the server.
// Its shape is not enforced by the SDK because the batch protocol supports
// heterogeneous operation lists. Consult the ModulaCMS API documentation for
// the expected batch request format.
//
// Returns [json.RawMessage] because the response structure varies depending on
// the operations included in the batch. The caller is responsible for
// unmarshalling the result.
//
// Returns an [*ApiError] if the server rejects the batch or the authenticated
// user lacks the content:update permission.
func (b *ContentBatchResource) Update(ctx context.Context, req any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := b.http.post(ctx, "/api/v1/content/batch", req, &result); err != nil {
		return nil, err
	}
	return result, nil
}
