package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// LocaleResource provides CRUD operations for locales, plus translation creation.
type LocaleResource struct {
	*Resource[Locale, CreateLocaleRequest, UpdateLocaleRequest, LocaleID]
	http *httpClient
}

func newLocaleResource(h *httpClient) *LocaleResource {
	return &LocaleResource{
		Resource: newResource[Locale, CreateLocaleRequest, UpdateLocaleRequest, LocaleID](h, "/api/v1/locales"),
		http:     h,
	}
}

// ListEnabled returns only enabled locales.
func (r *LocaleResource) ListEnabled(ctx context.Context) ([]Locale, error) {
	params := url.Values{}
	params.Set("enabled", "true")
	raw, err := r.RawList(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list enabled locales: %w", err)
	}
	var result []Locale
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode enabled locales: %w", err)
	}
	return result, nil
}

// CreateTranslation creates translated content fields for a content data node in the given locale.
func (r *LocaleResource) CreateTranslation(ctx context.Context, contentDataID string, req CreateTranslationRequest) (*CreateTranslationResponse, error) {
	var resp CreateTranslationResponse
	if err := r.http.post(ctx, "/api/v1/admin/contentdata/"+contentDataID+"/translations", req, &resp); err != nil {
		return nil, fmt.Errorf("create translation for %s: %w", contentDataID, err)
	}
	return &resp, nil
}
