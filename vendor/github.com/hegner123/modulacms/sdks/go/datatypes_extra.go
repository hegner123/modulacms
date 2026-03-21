package modula

import (
	"context"
	"fmt"
	"net/url"
)

// DatatypesExtraResource provides specialized datatype operations beyond
// standard CRUD, specifically sort order management. Datatypes can be
// organized in a tree hierarchy with explicit sort ordering within each level.
// It is accessed via [Client].DatatypesExtra.
type DatatypesExtraResource struct {
	http *httpClient
}

// UpdateSortOrder sets the sort order position for a datatype within its parent group.
// Lower values appear first. Use [DatatypesExtraResource.MaxSortOrder] to determine
// the next available position when appending.
func (r *DatatypesExtraResource) UpdateSortOrder(ctx context.Context, datatypeID DatatypeID, sortOrder int64) error {
	body := struct {
		SortOrder int64 `json:"sort_order"`
	}{SortOrder: sortOrder}
	if err := r.http.put(ctx, "/api/v1/datatype/"+string(datatypeID)+"/sort-order", body, nil); err != nil {
		return fmt.Errorf("update datatype sort order %s: %w", string(datatypeID), err)
	}
	return nil
}

// MaxSortOrder returns the highest sort order value currently assigned to datatypes
// under the given parent. Pass nil for parentID to query root-level datatypes.
// To append a new datatype at the end, use the returned value + 1 as the sort order.
// Returns 0 if there are no datatypes under the parent.
func (r *DatatypesExtraResource) MaxSortOrder(ctx context.Context, parentID *DatatypeID) (int64, error) {
	params := url.Values{}
	if parentID != nil {
		params.Set("parent_id", string(*parentID))
	}
	var result struct {
		MaxSortOrder int64 `json:"max_sort_order"`
	}
	if err := r.http.get(ctx, "/api/v1/datatype/max-sort-order", params, &result); err != nil {
		return 0, fmt.Errorf("get max datatype sort order: %w", err)
	}
	return result.MaxSortOrder, nil
}
