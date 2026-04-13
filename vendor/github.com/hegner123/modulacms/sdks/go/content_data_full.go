package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ContentDataFullResource provides access to composed content data endpoints
// that return content items with joined author, datatype, and field data.
// These endpoints return richer responses than the basic CRUD endpoints.
// It is accessed via [Client].ContentDataFull.
type ContentDataFullResource struct {
	http *httpClient
}

// GetFull returns a single content data item with its author, datatype, and fields.
// The response is returned as [json.RawMessage] because the shape includes nested
// associations that vary by datatype configuration.
func (r *ContentDataFullResource) GetFull(ctx context.Context, id ContentID) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result json.RawMessage
	if err := r.http.get(ctx, "/api/v1/contentdata/full", params, &result); err != nil {
		return nil, fmt.Errorf("get content data full %s: %w", string(id), err)
	}
	return result, nil
}

// ListByRoute returns all content data items belonging to the given route.
// The response is returned as [json.RawMessage] because the shape includes
// nested associations that vary by datatype configuration.
func (r *ContentDataFullResource) ListByRoute(ctx context.Context, routeID RouteID) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("q", string(routeID))
	var result json.RawMessage
	if err := r.http.get(ctx, "/api/v1/contentdata/by-route", params, &result); err != nil {
		return nil, fmt.Errorf("list content data by route %s: %w", string(routeID), err)
	}
	return result, nil
}

// AdminGetFull returns a single admin content data item with its author, datatype, and fields.
// The response is returned as [json.RawMessage] because the shape includes nested
// associations that vary by datatype configuration.
func (r *ContentDataFullResource) AdminGetFull(ctx context.Context, id AdminContentID) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result json.RawMessage
	if err := r.http.get(ctx, "/api/v1/admincontentdatas/full", params, &result); err != nil {
		return nil, fmt.Errorf("get admin content data full %s: %w", string(id), err)
	}
	return result, nil
}
