package modula

import (
	"context"
	"fmt"
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

// CreateTranslation creates translated content fields for a content data node in the given locale.
func (r *LocaleResource) CreateTranslation(ctx context.Context, contentDataID string, req CreateTranslationRequest) (*CreateTranslationResponse, error) {
	var resp CreateTranslationResponse
	if err := r.http.post(ctx, "/api/v1/admin/contentdata/"+contentDataID+"/translations", req, &resp); err != nil {
		return nil, fmt.Errorf("create translation for %s: %w", contentDataID, err)
	}
	return &resp, nil
}
