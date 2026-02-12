package db

// PaginationParams holds limit and offset for paginated queries.
type PaginationParams struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

// PaginatedResponse wraps a page of results with pagination metadata.
type PaginatedResponse[T any] struct {
	Data   []T   `json:"data"`
	Total  int64 `json:"total"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}
