package modulacms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Resource provides type-safe CRUD operations for an API entity.
// ID is constrained to ~string so any branded ID type works.
type Resource[Entity any, CreateParams any, UpdateParams any, ID ~string] struct {
	path string
	http *httpClient
}

// List returns all entities of this type.
func (r *Resource[E, C, U, ID]) List(ctx context.Context) ([]E, error) {
	var result []E
	if err := r.http.get(ctx, r.path, nil, &result); err != nil {
		return nil, fmt.Errorf("list %s: %w", r.path, err)
	}
	return result, nil
}

// Get returns a single entity by ID.
func (r *Resource[E, C, U, ID]) Get(ctx context.Context, id ID) (*E, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result E
	if err := r.http.get(ctx, r.path+"/", params, &result); err != nil {
		return nil, fmt.Errorf("get %s %s: %w", r.path, string(id), err)
	}
	return &result, nil
}

// Create creates a new entity and returns it.
func (r *Resource[E, C, U, ID]) Create(ctx context.Context, params C) (*E, error) {
	var result E
	if err := r.http.post(ctx, r.path, params, &result); err != nil {
		return nil, fmt.Errorf("create %s: %w", r.path, err)
	}
	return &result, nil
}

// Update updates an existing entity and returns the updated version.
func (r *Resource[E, C, U, ID]) Update(ctx context.Context, params U) (*E, error) {
	var result E
	if err := r.http.put(ctx, r.path+"/", params, &result); err != nil {
		return nil, fmt.Errorf("update %s: %w", r.path, err)
	}
	return &result, nil
}

// Delete removes an entity by ID.
func (r *Resource[E, C, U, ID]) Delete(ctx context.Context, id ID) error {
	params := url.Values{}
	params.Set("q", string(id))
	if err := r.http.del(ctx, r.path+"/", params); err != nil {
		return fmt.Errorf("delete %s %s: %w", r.path, string(id), err)
	}
	return nil
}

// ListPaginated returns a page of entities with pagination metadata.
func (r *Resource[E, C, U, ID]) ListPaginated(ctx context.Context, p PaginationParams) (*PaginatedResponse[E], error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", p.Limit))
	params.Set("offset", fmt.Sprintf("%d", p.Offset))
	var result PaginatedResponse[E]
	if err := r.http.get(ctx, r.path, params, &result); err != nil {
		return nil, fmt.Errorf("list paginated %s: %w", r.path, err)
	}
	return &result, nil
}

// Count returns the total number of entities of this type.
func (r *Resource[E, C, U, ID]) Count(ctx context.Context) (int64, error) {
	params := url.Values{}
	params.Set("count", "true")
	var result struct {
		Count int64 `json:"count"`
	}
	if err := r.http.get(ctx, r.path, params, &result); err != nil {
		return 0, fmt.Errorf("count %s: %w", r.path, err)
	}
	return result.Count, nil
}

// RawList returns the raw JSON for list requests, useful for custom decoding.
func (r *Resource[E, C, U, ID]) RawList(ctx context.Context, params url.Values) (json.RawMessage, error) {
	var result json.RawMessage
	if err := r.http.get(ctx, r.path, params, &result); err != nil {
		return nil, fmt.Errorf("raw list %s: %w", r.path, err)
	}
	return result, nil
}
