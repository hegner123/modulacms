package modula

import (
	"context"
	"fmt"
	"net/url"
)

// FieldsExtraResource provides specialized field operations beyond standard CRUD,
// specifically sort order management. Fields within a datatype are displayed in
// sort order, and these methods allow reordering without a full field update.
// It is accessed via [Client].FieldsExtra.
type FieldsExtraResource struct {
	http *httpClient
}

// UpdateSortOrder sets the sort order position for a field within its parent datatype.
// Lower values appear first in the field list. Use [FieldsExtraResource.MaxSortOrder]
// to determine the next available position when appending a new field.
func (r *FieldsExtraResource) UpdateSortOrder(ctx context.Context, fieldID FieldID, sortOrder int64) error {
	body := struct {
		SortOrder int64 `json:"sort_order"`
	}{SortOrder: sortOrder}
	if err := r.http.put(ctx, "/api/v1/fields/"+string(fieldID)+"/sort-order", body, nil); err != nil {
		return fmt.Errorf("update field sort order %s: %w", string(fieldID), err)
	}
	return nil
}

// MaxSortOrder returns the highest sort order value currently assigned to fields
// under the given parent datatype. To append a new field at the end of the list,
// use the returned value + 1 as the sort order. Returns 0 if the datatype has no fields.
func (r *FieldsExtraResource) MaxSortOrder(ctx context.Context, parentID DatatypeID) (int64, error) {
	params := url.Values{}
	params.Set("parent_id", string(parentID))
	var result struct {
		MaxSortOrder int64 `json:"max_sort_order"`
	}
	if err := r.http.get(ctx, "/api/v1/fields/max-sort-order", params, &result); err != nil {
		return 0, fmt.Errorf("get max sort order for parent %s: %w", string(parentID), err)
	}
	return result.MaxSortOrder, nil
}
