package modula

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ---------------------------------------------------------------------------
// Config types
// ---------------------------------------------------------------------------

// ConfigResponse is the response from the config GET endpoint.
type ConfigResponse struct {
	Config map[string]any `json:"config"`
}

// ConfigUpdateResponse is the response from the config PATCH endpoint.
type ConfigUpdateResponse struct {
	OK              bool           `json:"ok"`
	Config          map[string]any `json:"config"`
	RestartRequired []string       `json:"restart_required,omitempty"`
	Warnings        []string       `json:"warnings,omitempty"`
}

// ConfigFieldMeta describes a single config field.
type ConfigFieldMeta struct {
	JSONKey       string `json:"json_key"`
	Label         string `json:"label"`
	Category      string `json:"category"`
	HotReloadable bool   `json:"hot_reloadable"`
	Sensitive     bool   `json:"sensitive"`
	Required      bool   `json:"required"`
	Description   string `json:"description"`
}

// ConfigCategory describes a config category with its key and display label.
type ConfigCategory struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// ConfigMetaResponse is the response from the config meta endpoint.
type ConfigMetaResponse struct {
	Fields     []ConfigFieldMeta `json:"fields"`
	Categories []ConfigCategory  `json:"categories"`
}

// ---------------------------------------------------------------------------
// ConfigResource
// ---------------------------------------------------------------------------

// ConfigResource provides config management operations.
type ConfigResource struct {
	http *httpClient
}

// Get returns the current server config (redacted).
// An optional category filter narrows the result to fields in that category.
func (c *ConfigResource) Get(ctx context.Context, category string) (*ConfigResponse, error) {
	var params url.Values
	if category != "" {
		params = url.Values{"category": {category}}
	}
	var result ConfigResponse
	if err := c.http.get(ctx, "/api/v1/admin/config", params, &result); err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	return &result, nil
}

// Update applies a partial config update.
func (c *ConfigResource) Update(ctx context.Context, updates map[string]any) (*ConfigUpdateResponse, error) {
	var result ConfigUpdateResponse
	if err := c.http.jsonRequest(ctx, http.MethodPatch, "/api/v1/admin/config", updates, &result); err != nil {
		return nil, fmt.Errorf("update config: %w", err)
	}
	return &result, nil
}

// Meta returns the field metadata registry.
func (c *ConfigResource) Meta(ctx context.Context) (*ConfigMetaResponse, error) {
	var result ConfigMetaResponse
	if err := c.http.get(ctx, "/api/v1/admin/config/meta", nil, &result); err != nil {
		return nil, fmt.Errorf("get config meta: %w", err)
	}
	return &result, nil
}

// ConfigSearchIndexEntry describes a single config field with its metadata,
// help text, default value, and example. Returned by [ConfigResource.SearchIndex].
type ConfigSearchIndexEntry struct {
	Key           string `json:"key"`
	Label         string `json:"label"`
	Category      string `json:"category"`
	CategoryLabel string `json:"category_label"`
	Description   string `json:"description"`
	HelpText      string `json:"help_text"`
	Default       string `json:"default"`
	Example       string `json:"example"`
	HotReloadable bool   `json:"hot_reloadable"`
	Sensitive     bool   `json:"sensitive"`
	Required      bool   `json:"required"`
}

// SearchIndex returns a searchable index of all config fields, combining field
// registry metadata with help text. Useful for building config search UIs.
func (c *ConfigResource) SearchIndex(ctx context.Context) ([]ConfigSearchIndexEntry, error) {
	var result []ConfigSearchIndexEntry
	if err := c.http.get(ctx, "/api/v1/admin/config/search-index", nil, &result); err != nil {
		return nil, fmt.Errorf("get config search index: %w", err)
	}
	return result, nil
}
