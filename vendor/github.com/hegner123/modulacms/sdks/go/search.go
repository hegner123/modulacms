package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// SearchResource provides full-text search over published content.
// Search indexes published content only and requires no authentication.
// It is accessed via [Client].Search.
type SearchResource struct {
	http *httpClient
}

// SearchOptions configures filtering and pagination for a search query.
// All fields are optional; zero values are omitted from the request.
type SearchOptions struct {
	// Type filters results to a specific datatype name (e.g., "blog_post").
	Type string
	// Locale filters results to a specific locale code (e.g., "en", "fr").
	Locale string
	// Limit is the maximum number of results to return. Server default is 20.
	Limit int
	// Offset is the number of results to skip for pagination.
	Offset int
	// Prefix enables prefix matching on the last query term for search-as-you-type.
	// Defaults to true on the server when not specified.
	Prefix *bool
}

// SearchResult represents a single search hit with relevance score and snippet.
type SearchResult struct {
	ID            string  `json:"id"`
	ContentDataID string  `json:"content_data_id"`
	RouteSlug     string  `json:"route_slug"`
	RouteTitle    string  `json:"route_title"`
	DatatypeName  string  `json:"datatype_name"`
	DatatypeLabel string  `json:"datatype_label"`
	Locale        string  `json:"locale,omitempty"`
	Section       string  `json:"section,omitempty"`
	SectionAnchor string  `json:"section_anchor,omitempty"`
	Score         float64 `json:"score"`
	Snippet       string  `json:"snippet"`
	PublishedAt   string  `json:"published_at"`
}

// SearchResponse is the envelope returned by a search query.
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
}

// Search executes a full-text search against published content.
// The query parameter is required. Pass nil for opts to use server defaults.
//
// Example:
//
//	resp, err := client.Search.Search(ctx, "installation guide", &modula.SearchOptions{
//	    Type:  "doc_page",
//	    Limit: 10,
//	})
func (r *SearchResource) Search(ctx context.Context, query string, opts *SearchOptions) (*SearchResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("modula: search query is required")
	}
	p := url.Values{}
	p.Set("q", query)
	if opts != nil {
		if opts.Type != "" {
			p.Set("type", opts.Type)
		}
		if opts.Locale != "" {
			p.Set("locale", opts.Locale)
		}
		if opts.Limit > 0 {
			p.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Offset > 0 {
			p.Set("offset", strconv.Itoa(opts.Offset))
		}
		if opts.Prefix != nil {
			p.Set("prefix", strconv.FormatBool(*opts.Prefix))
		}
	}
	var raw json.RawMessage
	if err := r.http.get(ctx, "/api/v1/search", p, &raw); err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	var result SearchResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal search response: %w", err)
	}
	return &result, nil
}
