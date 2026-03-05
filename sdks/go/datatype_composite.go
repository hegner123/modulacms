package modula

import (
	"context"
	"fmt"
	"net/url"
)

// DatatypeCompositeResource provides composite datatype operations such as
// cascade delete (removing a datatype and all of its associated content).
type DatatypeCompositeResource struct {
	http *httpClient
}

// DeleteCascade deletes a datatype and all content associated with it.
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
