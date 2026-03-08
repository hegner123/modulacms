package modula

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// ContentDeliveryResource provides public (unauthenticated) content delivery
// via slug-based routing. It maps directly to the GET /api/v1/content/{slug}
// endpoint, which resolves a URL slug to its published content tree and returns
// the result in one of several configurable response formats.
//
// Access this resource via [Client].Content:
//
//	page, err := client.Content.GetPage(ctx, "/about", "clean", "")
//
// The returned [json.RawMessage] must be decoded by the caller because the
// response shape varies by format (e.g. "clean" produces a flat object,
// "contentful" produces a Contentful-compatible structure, etc.).
type ContentDeliveryResource struct {
	http *httpClient
}

// GetPage retrieves published content by its URL slug.
//
// Parameters:
//   - slug: the URL path segment identifying the content (e.g. "/about", "/blog/my-post").
//     A leading "/" is stripped automatically.
//   - format: controls the response structure. Supported values are "clean", "raw",
//     "contentful", "sanity", "strapi", and "wordpress". Pass an empty string to use
//     the server default.
//   - locale: when non-empty, requests content translated to the given locale code
//     (e.g. "en-US", "fr"). Pass an empty string for the default locale.
//
// GetPage returns [json.RawMessage] because each format produces a different JSON
// shape. The caller is responsible for unmarshalling into the appropriate struct
// or inspecting the raw bytes.
//
// Returns an [*ApiError] with status 404 if no published content exists at the
// given slug.
func (c *ContentDeliveryResource) GetPage(ctx context.Context, slug string, format string, locale string) (json.RawMessage, error) {
	params := url.Values{}
	if format != "" {
		params.Set("format", format)
	}
	if locale != "" {
		params.Set("locale", locale)
	}
	path := "/api/v1/content/" + strings.TrimLeft(slug, "/")
	var result json.RawMessage
	if err := c.http.get(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return result, nil
}
