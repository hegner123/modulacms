package modula

// PaginationParams specifies limit and offset for paginated requests.
type PaginationParams struct {
	Limit  int64
	Offset int64
}

// PaginatedResponse wraps a page of results with pagination metadata.
type PaginatedResponse[T any] struct {
	Data   []T   `json:"data"`
	Total  int64 `json:"total"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}
