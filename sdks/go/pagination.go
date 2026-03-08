package modula

// PaginationParams controls which page of results to retrieve from a
// paginated list endpoint. Pass these to [Resource.ListPaginated].
type PaginationParams struct {
	// Limit is the maximum number of items to return in this page.
	// The server may cap this to a maximum value (typically 100).
	Limit int64

	// Offset is the zero-based index of the first item to return.
	// For the first page, use 0. For subsequent pages, increment by
	// the page size (Limit).
	Offset int64
}

// PaginatedResponse wraps a single page of results with metadata needed
// to implement pagination controls. T is the entity type being listed.
type PaginatedResponse[T any] struct {
	// Data is the slice of entities on this page.
	Data []T `json:"data"`

	// Total is the total number of entities matching the query across
	// all pages. Use this with Offset to determine whether more pages
	// remain.
	Total int64 `json:"total"`

	// Limit is the maximum number of items that were requested.
	Limit int64 `json:"limit"`

	// Offset is the starting index of this page within the full result set.
	Offset int64 `json:"offset"`
}
