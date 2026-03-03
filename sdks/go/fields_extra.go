package modula

import (
	"context"
	"fmt"
	"net/url"
)

// FieldsExtraResource provides specialized field operations beyond standard CRUD.
type FieldsExtraResource struct {
	http *httpClient
}

// UpdateSortOrder updates the sort order for a field.
func (r *FieldsExtraResource) UpdateSortOrder(ctx context.Context, fieldID FieldID, sortOrder int64) error {
	body := struct {
		SortOrder int64 `json:"sort_order"`
	}{SortOrder: sortOrder}
	if err := r.http.put(ctx, "/api/v1/fields/"+string(fieldID)+"/sort-order", body, nil); err != nil {
		return fmt.Errorf("update field sort order %s: %w", string(fieldID), err)
	}
	return nil
}

// MaxSortOrder returns the maximum sort order value for fields under a parent datatype.
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
