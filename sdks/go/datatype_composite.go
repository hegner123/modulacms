package modula

import (
	"context"
	"fmt"
	"net/url"
)

// DatatypeCompositeResource provides composite datatype operations that span
// the datatype, field, and content tables atomically. Use this for operations
// like cascade deletion where removing a datatype must also clean up all
// associated fields, content nodes, and content field values.
// It is accessed via [Client].DatatypeComposite.
type DatatypeCompositeResource struct {
	http *httpClient
}

// DeleteCascade deletes a datatype and all associated data: fields, content nodes,
// content field values, and content relations. This is irreversible. The returned
// [*DatatypeCascadeDeleteResponse] reports the number of items removed per table.
// Returns an [*ApiError] with status 404 if the datatype does not exist, or 403
// if the caller lacks permission.
func (d *DatatypeCompositeResource) DeleteCascade(ctx context.Context, id DatatypeID) (*DatatypeCascadeDeleteResponse, error) {
	params := url.Values{}
	params.Set("q", string(id))
	params.Set("cascade", "true")
	var result DatatypeCascadeDeleteResponse
	if err := d.http.delBody(ctx, "/api/v1/datatype/", params, &result); err != nil {
		return nil, fmt.Errorf("datatype cascade delete %s: %w", string(id), err)
	}
	return &result, nil
}
