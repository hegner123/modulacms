package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Resource provides type-safe CRUD operations for a single API entity type.
//
// The four type parameters are:
//   - Entity: the struct returned by the API (e.g. [ContentData], [User]).
//   - CreateParams: the struct sent when creating a new entity.
//   - UpdateParams: the struct sent when updating an existing entity.
//   - ID: the branded string type used to identify this entity (e.g. [ContentID]).
//     The ~string constraint ensures any named string type works.
//
// Resource instances are created internally by [NewClient] and exposed as
// fields on [Client]. Do not construct them directly.
//
// All methods accept a [context.Context] for cancellation and deadline
// propagation. Errors from the API are returned as [*ApiError] and can be
// inspected with [IsNotFound], [IsUnauthorized], and similar helpers.
type Resource[Entity any, CreateParams any, UpdateParams any, ID ~string] struct {
	path string
	http *httpClient
}

// List returns all entities of this type as an unordered slice.
// For large collections, prefer [Resource.ListPaginated] to avoid loading
// the entire dataset into memory.
func (r *Resource[E, C, U, ID]) List(ctx context.Context) ([]E, error) {
	var result []E
	if err := r.http.get(ctx, r.path, nil, &result); err != nil {
		return nil, fmt.Errorf("list %s: %w", r.path, err)
	}
	return result, nil
}

// Get returns a single entity by its branded ID.
// Returns [*ApiError] with status 404 if the entity does not exist;
// use [IsNotFound] to check for this condition.
func (r *Resource[E, C, U, ID]) Get(ctx context.Context, id ID) (*E, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result E
	if err := r.http.get(ctx, r.path+"/", params, &result); err != nil {
		return nil, fmt.Errorf("get %s %s: %w", r.path, string(id), err)
	}
	return &result, nil
}

// Create creates a new entity from the given parameters and returns the
// server-assigned entity, which includes the generated ID and timestamps.
// The params struct is JSON-encoded and sent as the request body.
func (r *Resource[E, C, U, ID]) Create(ctx context.Context, params C) (*E, error) {
	var result E
	if err := r.http.post(ctx, r.path, params, &result); err != nil {
		return nil, fmt.Errorf("create %s: %w", r.path, err)
	}
	return &result, nil
}

// Update applies a partial or full update to an existing entity and returns
// the updated version. The UpdateParams struct typically includes the entity's
// ID field to identify which record to update.
func (r *Resource[E, C, U, ID]) Update(ctx context.Context, params U) (*E, error) {
	var result E
	if err := r.http.put(ctx, r.path+"/", params, &result); err != nil {
		return nil, fmt.Errorf("update %s: %w", r.path, err)
	}
	return &result, nil
}

// Delete permanently removes an entity by its branded ID.
// Returns nil on success, even if the entity was already absent on some
// server implementations. Returns [*ApiError] on failure.
func (r *Resource[E, C, U, ID]) Delete(ctx context.Context, id ID) error {
	params := url.Values{}
	params.Set("q", string(id))
	if err := r.http.del(ctx, r.path+"/", params); err != nil {
		return fmt.Errorf("delete %s %s: %w", r.path, string(id), err)
	}
	return nil
}

// ListPaginated returns a single page of entities along with pagination
// metadata (total count, limit, offset). Use [PaginationParams] to control
// page size and position.
//
// To iterate through all pages:
//
//	offset := int64(0)
//	for {
//	    page, err := client.Users.ListPaginated(ctx, modula.PaginationParams{Limit: 50, Offset: offset})
//	    if err != nil { return err }
//	    process(page.Data)
//	    offset += int64(len(page.Data))
//	    if offset >= page.Total { break }
//	}
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

// Count returns the total number of entities of this type without
// transferring the entity data itself. Useful for display or deciding
// whether pagination is needed.
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

// ListWithParams performs a list request with custom query parameters and
// decodes the response into the provided result pointer. This is useful when
// the server supports query-parameter filtering on a list endpoint (e.g.
// `?parent_id=...` to list children of a specific parent).
func (r *Resource[E, C, U, ID]) ListWithParams(ctx context.Context, params url.Values, result any) error {
	if err := r.http.get(ctx, r.path, params, result); err != nil {
		return fmt.Errorf("list %s: %w", r.path, err)
	}
	return nil
}

// RawList returns the raw JSON response body for a list request, allowing
// the caller to perform custom decoding. This is useful when you need access
// to the original JSON structure or when the standard type parameters do not
// match the response format (e.g. custom query parameters that change the
// response shape). The params argument is sent as URL query parameters.
func (r *Resource[E, C, U, ID]) RawList(ctx context.Context, params url.Values) (json.RawMessage, error) {
	var result json.RawMessage
	if err := r.http.get(ctx, r.path, params, &result); err != nil {
		return nil, fmt.Errorf("raw list %s: %w", r.path, err)
	}
	return result, nil
}
