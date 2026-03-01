package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// QueryResource provides content query operations.
type QueryResource struct {
	http *httpClient
}

// QueryParams configures a content query request.
type QueryParams struct {
	Sort    string
	Limit   int
	Offset  int
	Locale  string
	Status  string
	Filters map[string]string
}

// QueryResult is the response envelope for a content query.
type QueryResult struct {
	Data     []QueryItem   `json:"data"`
	Total    int64         `json:"total"`
	Limit    int64         `json:"limit"`
	Offset   int64         `json:"offset"`
	Datatype QueryDatatype `json:"datatype"`
}

// QueryItem represents a single content item in a query result.
type QueryItem struct {
	ContentDataID string            `json:"content_data_id"`
	DatatypeID    string            `json:"datatype_id"`
	AuthorID      string            `json:"author_id"`
	Status        string            `json:"status"`
	DateCreated   string            `json:"date_created"`
	DateModified  string            `json:"date_modified"`
	PublishedAt   string            `json:"published_at"`
	Fields        map[string]string `json:"fields"`
}

// QueryDatatype holds the datatype metadata in a query result.
type QueryDatatype struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// Query fetches a filtered, sorted, paginated list of content items by datatype name.
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
