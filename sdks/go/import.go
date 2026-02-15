package modulacms

import (
	"context"
	"encoding/json"
)

// ImportResource provides CMS data import operations.
type ImportResource struct {
	http *httpClient
}

// Contentful imports data from Contentful format.
func (i *ImportResource) Contentful(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/contentful", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Sanity imports data from Sanity format.
func (i *ImportResource) Sanity(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/sanity", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Strapi imports data from Strapi format.
func (i *ImportResource) Strapi(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/strapi", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// WordPress imports data from WordPress format.
func (i *ImportResource) WordPress(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/wordpress", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Clean imports data using the clean/normalized format.
func (i *ImportResource) Clean(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/clean", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Bulk imports data in bulk.
func (i *ImportResource) Bulk(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
