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

// ConfigMetaResponse is the response from the config meta endpoint.
type ConfigMetaResponse struct {
	Fields     []ConfigFieldMeta `json:"fields"`
	Categories []string          `json:"categories"`
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
