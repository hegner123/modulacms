package modula

import (
	"context"
	"encoding/json"
)

// ImportResource provides CMS data import operations from various headless CMS
// formats. Each method accepts data in the source CMS's export format and
// transforms it into ModulaCMS's content model (datatypes, fields, content nodes).
//
// The response is [json.RawMessage] containing an import summary with counts of
// created datatypes, fields, and content items, plus any warnings or errors
// encountered during transformation.
//
// It is accessed via [Client].Import.
type ImportResource struct {
	http *httpClient
}

// Contentful imports data from Contentful's export format. The data parameter should
// be a Contentful space export (content types, entries, and assets). The server
// maps Contentful content types to datatypes and entries to content nodes.
func (i *ImportResource) Contentful(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/contentful", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Sanity imports data from Sanity's NDJSON export format. The data parameter should
// be Sanity documents including schema types. The server maps Sanity document types
// to datatypes and documents to content nodes.
func (i *ImportResource) Sanity(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/sanity", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Strapi imports data from Strapi's export format. The data parameter should
// contain Strapi content types and entries. The server maps Strapi collection types
// to datatypes and entries to content nodes.
func (i *ImportResource) Strapi(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/strapi", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// WordPress imports data from WordPress's WXR (WordPress eXtended RSS) export format.
// The data parameter should contain WordPress posts, pages, categories, and media.
// The server maps WordPress post types to datatypes and posts/pages to content nodes.
func (i *ImportResource) WordPress(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/wordpress", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Clean imports data using ModulaCMS's clean/normalized format. This is the
// preferred format for programmatic imports, as it maps directly to the CMS's
// internal data model without transformation. Use this when building custom
// import pipelines or migrating from unsupported CMS platforms.
func (i *ImportResource) Clean(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import/clean", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Bulk performs a bulk import of data using the server's generic import endpoint.
// The data parameter should contain a collection of records to import. This method
// auto-detects the format or uses the raw payload directly for high-volume imports.
func (i *ImportResource) Bulk(ctx context.Context, data any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := i.http.post(ctx, "/api/v1/import", data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
