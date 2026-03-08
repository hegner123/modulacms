package modula

import (
	"context"
	"fmt"
	"net/url"
)

// AdminDatatypesExtraResource provides specialized admin datatype operations
// beyond standard CRUD, specifically sort order management for the admin-side
// datatype hierarchy. Admin datatypes define the schema for internal/administrative
// content (as opposed to public-facing content datatypes).
// It is accessed via [Client].AdminDatatypesExtra.
type AdminDatatypesExtraResource struct {
	http *httpClient
}

// UpdateSortOrder sets the sort order position for an admin datatype within its
// parent group. Lower values appear first. Use [AdminDatatypesExtraResource.MaxSortOrder]
// to determine the next available position when appending.
func (r *AdminDatatypesExtraResource) UpdateSortOrder(ctx context.Context, id AdminDatatypeID, sortOrder int64) error {
	body := struct {
		SortOrder int64 `json:"sort_order"`
	}{SortOrder: sortOrder}
	if err := r.http.put(ctx, "/api/v1/admindatatypes/"+string(id)+"/sort-order", body, nil); err != nil {
		return fmt.Errorf("update admin datatype sort order %s: %w", string(id), err)
	}
	return nil
}

// MaxSortOrder returns the highest sort order value currently assigned to admin
// datatypes under the given parent. Pass nil for parentID to query root-level
// admin datatypes. To append a new admin datatype at the end, use the returned
// value + 1 as the sort order. Returns 0 if there are no admin datatypes under the parent.
func (r *AdminDatatypesExtraResource) MaxSortOrder(ctx context.Context, parentID *AdminDatatypeID) (int64, error) {
	params := url.Values{}
	if parentID != nil {
		params.Set("parent_id", string(*parentID))
	}
	var result struct {
		MaxSortOrder int64 `json:"max_sort_order"`
	}
	if err := r.http.get(ctx, "/api/v1/admindatatypes/max-sort-order", params, &result); err != nil {
		return 0, fmt.Errorf("get max admin datatype sort order: %w", err)
	}
	return result.MaxSortOrder, nil
}
