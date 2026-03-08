package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// QueryResource provides filtered, sorted, paginated content queries by datatype.
// This is the primary way to fetch content for frontend rendering, supporting
// field-level filtering, locale selection, status filtering, and pagination.
// It is accessed via [Client].Query.
type QueryResource struct {
	http *httpClient
}

// QueryParams configures filtering, sorting, and pagination for a content query.
// All fields are optional; zero values are omitted from the request.
type QueryParams struct {
	// Sort specifies the sort field and direction (e.g., "date_created", "-date_modified").
	// Prefix with "-" for descending order.
	Sort string
	// Limit is the maximum number of items to return. Server default applies when 0.
	Limit int
	// Offset is the number of items to skip for pagination. Starts at 0.
	Offset int
	// Locale filters results to content with translations in the given locale code
	// (e.g., "en-US", "fr-FR"). When empty, the default locale is used.
	Locale string
	// Status filters results by content status (e.g., "published", "draft").
	// When empty, all statuses are returned.
	Status string
	// Filters is a map of field-name to value for field-level filtering.
	// Keys correspond to field names defined on the datatype; values are
	// matched exactly against content field values.
	Filters map[string]string
}

// QueryResult is the paginated response envelope for a content query.
type QueryResult struct {
	// Data contains the content items matching the query, up to Limit items.
	Data []QueryItem `json:"data"`
	// Total is the total number of items matching the query across all pages.
	Total int64 `json:"total"`
	// Limit is the page size used for this query.
	Limit int64 `json:"limit"`
	// Offset is the number of items skipped before this page.
	Offset int64 `json:"offset"`
	// Datatype contains metadata about the datatype that was queried.
	Datatype QueryDatatype `json:"datatype"`
}

// QueryItem represents a single content item in a query result, with its
// field values flattened into a string map keyed by field name.
type QueryItem struct {
	// ContentDataID is the unique ULID of this content item.
	ContentDataID string `json:"content_data_id"`
	// DatatypeID is the ULID of the datatype this content belongs to.
	DatatypeID string `json:"datatype_id"`
	// AuthorID is the ULID of the user who created this content.
	AuthorID string `json:"author_id"`
	// Status is the content's lifecycle status ("draft", "published", "scheduled").
	Status string `json:"status"`
	// DateCreated is the ISO 8601 timestamp when the content was created.
	DateCreated string `json:"date_created"`
	// DateModified is the ISO 8601 timestamp when the content was last modified.
	DateModified string `json:"date_modified"`
	// PublishedAt is the ISO 8601 timestamp when the content was published.
	// Empty if the content has never been published.
	PublishedAt string `json:"published_at"`
	// Fields is a map of field name to field value for this content item.
	// All values are strings; parse them according to the field type defined
	// on the datatype schema.
	Fields map[string]string `json:"fields"`
}

// QueryDatatype holds the datatype metadata included in a [QueryResult].
type QueryDatatype struct {
	// Name is the datatype's programmatic name (slug), used in API paths.
	Name string `json:"name"`
	// Label is the datatype's human-readable display name.
	Label string `json:"label"`
}

// Query fetches a filtered, sorted, paginated list of content items for the given
// datatype name (slug). The datatype parameter is required and must match an existing
// datatype's Name field. Pass nil for params to use server defaults (no filters,
// default sort and pagination).
//
// Example:
//
//	result, err := client.Query.Query(ctx, "blog-posts", &modula.QueryParams{
//	    Sort:   "-date_created",
//	    Limit:  10,
//	    Status: "published",
//	})
func (r *QueryResource) Query(ctx context.Context, datatype string, params *QueryParams) (*QueryResult, error) {
	if datatype == "" {
		return nil, fmt.Errorf("modula: datatype name is required")
	}
	p := url.Values{}
	if params != nil {
		if params.Sort != "" {
			p.Set("sort", params.Sort)
		}
		if params.Limit > 0 {
			p.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Offset > 0 {
			p.Set("offset", strconv.Itoa(params.Offset))
		}
		if params.Locale != "" {
			p.Set("locale", params.Locale)
		}
		if params.Status != "" {
			p.Set("status", params.Status)
		}
		for k, v := range params.Filters {
			p.Set(k, v)
		}
	}
	var raw json.RawMessage
	if err := r.http.get(ctx, "/api/v1/query/"+datatype, p, &raw); err != nil {
		return nil, fmt.Errorf("query %s: %w", datatype, err)
	}
	var result QueryResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal query result: %w", err)
	}
	return &result, nil
}
