package modula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// LocaleResource provides CRUD operations for locales and translation management.
// Locales represent supported languages/regions in the CMS (e.g., "en-US", "fr-FR").
// Content can be translated into any enabled locale, with each translation stored
// as locale-specific content field values.
//
// LocaleResource embeds [Resource] for standard List, Get, Create, Update, Delete
// operations, and adds locale-specific methods for filtering and translation creation.
// It is accessed via [Client].Locales.
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

// ListEnabled returns only locales that are currently enabled for content translation.
// Disabled locales are excluded from the result. Use this to populate locale pickers
// or determine which translations are available for content.
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

// CreateTranslation creates locale-specific content field values for a content node.
// The contentDataID identifies the content node to translate, and the request specifies
// the target locale and translated field values. If translations already exist for the
// given locale, they are updated rather than duplicated.
func (r *LocaleResource) CreateTranslation(ctx context.Context, contentDataID string, req CreateTranslationRequest) (*CreateTranslationResponse, error) {
	var resp CreateTranslationResponse
	if err := r.http.post(ctx, "/api/v1/admin/contentdata/"+contentDataID+"/translations", req, &resp); err != nil {
		return nil, fmt.Errorf("create translation for %s: %w", contentDataID, err)
	}
	return &resp, nil
}

// CreateAdminTranslation creates locale-specific content field values for an admin content node.
// The adminContentDataID identifies the admin content node to translate, and the request
// specifies the target locale. This is the admin-side equivalent of [LocaleResource.CreateTranslation].
func (r *LocaleResource) CreateAdminTranslation(ctx context.Context, adminContentDataID string, req CreateTranslationRequest) (*CreateTranslationResponse, error) {
	var resp CreateTranslationResponse
	if err := r.http.post(ctx, "/api/v1/admin/admincontentdata/"+adminContentDataID+"/translations", req, &resp); err != nil {
		return nil, fmt.Errorf("create admin translation for %s: %w", adminContentDataID, err)
	}
	return &resp, nil
}
