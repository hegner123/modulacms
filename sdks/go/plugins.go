package modulacms

import (
	"context"
	"fmt"
	"net/url"
)

// ---------------------------------------------------------------------------
// Plugin types
// ---------------------------------------------------------------------------

// PluginListItem is a summary returned by the plugin list endpoint.
type PluginListItem struct {
	Name                string `json:"name"`
	Version             string `json:"version"`
	Description         string `json:"description"`
	State               string `json:"state"`
	CircuitBreakerState string `json:"circuit_breaker_state,omitempty"`
}

// DriftEntry describes a missing or extra column detected in plugin schema.
type DriftEntry struct {
	Table  string `json:"table"`
	Kind   string `json:"kind"`
	Column string `json:"column"`
}

// PluginInfo is detailed information for a single plugin.
type PluginInfo struct {
	Name                string       `json:"name"`
	Version             string       `json:"version"`
	Description         string       `json:"description"`
	Author              string       `json:"author,omitempty"`
	License             string       `json:"license,omitempty"`
	State               string       `json:"state"`
	FailedReason        string       `json:"failed_reason,omitempty"`
	CircuitBreakerState string       `json:"circuit_breaker_state,omitempty"`
	CircuitBreakerErrs  int          `json:"circuit_breaker_errors,omitempty"`
	VMsAvailable        int          `json:"vms_available"`
	VMsTotal            int          `json:"vms_total"`
	Dependencies        []string     `json:"dependencies,omitempty"`
	SchemaDrift         []DriftEntry `json:"schema_drift,omitempty"`
}

// PluginActionResponse is the response from reload.
type PluginActionResponse struct {
	OK     bool   `json:"ok"`
	Plugin string `json:"plugin"`
}

// PluginStateResponse is the response from enable/disable.
type PluginStateResponse struct {
	OK     bool   `json:"ok"`
	Plugin string `json:"plugin"`
	State  string `json:"state"`
}

// CleanupDryRunResponse is the response from cleanup dry-run (GET).
type CleanupDryRunResponse struct {
	OrphanedTables []string `json:"orphaned_tables"`
	Count          int      `json:"count"`
	Action         string   `json:"action"`
}

// CleanupDropParams holds parameters for the cleanup drop (POST) endpoint.
type CleanupDropParams struct {
	Confirm bool     `json:"confirm"`
	Tables  []string `json:"tables"`
}

// CleanupDropResponse is the response from cleanup drop (POST).
type CleanupDropResponse struct {
	Dropped []string `json:"dropped"`
	Count   int      `json:"count"`
}

// PluginRoute is a route registered by a plugin.
type PluginRoute struct {
	Plugin   string `json:"plugin"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Public   bool   `json:"public"`
	Approved bool   `json:"approved"`
}

// RouteApprovalItem identifies a plugin route for approval or revocation.
type RouteApprovalItem struct {
	Plugin string `json:"plugin"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

// PluginHook is a hook registered by a plugin.
type PluginHook struct {
	PluginName string `json:"plugin_name"`
	Event      string `json:"event"`
	Table      string `json:"table"`
	Priority   int    `json:"priority"`
	Approved   bool   `json:"approved"`
	IsWildcard bool   `json:"is_wildcard"`
}

// HookApprovalItem identifies a plugin hook for approval or revocation.
type HookApprovalItem struct {
	Plugin string `json:"plugin"`
	Event  string `json:"event"`
	Table  string `json:"table"`
}

// ---------------------------------------------------------------------------
// Response envelopes (unexported, used for JSON decode then unwrap)
// ---------------------------------------------------------------------------

type pluginListEnvelope struct {
	Plugins []PluginListItem `json:"plugins"`
}

type routeListEnvelope struct {
	Routes []PluginRoute `json:"routes"`
}

type hookListEnvelope struct {
	Hooks []PluginHook `json:"hooks"`
}

type routeApprovalBody struct {
	Routes []RouteApprovalItem `json:"routes"`
}

type hookApprovalBody struct {
	Hooks []HookApprovalItem `json:"hooks"`
}

// ---------------------------------------------------------------------------
// PluginsResource — plugin management
// ---------------------------------------------------------------------------

// PluginsResource provides plugin management operations.
type PluginsResource struct {
	http *httpClient
}

// List returns all installed plugins.
func (p *PluginsResource) List(ctx context.Context) ([]PluginListItem, error) {
	var envelope pluginListEnvelope
	if err := p.http.get(ctx, "/api/v1/admin/plugins", nil, &envelope); err != nil {
		return nil, fmt.Errorf("list plugins: %w", err)
	}
	return envelope.Plugins, nil
}

// Get returns detailed info for a specific plugin.
func (p *PluginsResource) Get(ctx context.Context, name string) (*PluginInfo, error) {
	var result PluginInfo
	path := "/api/v1/admin/plugins/" + url.PathEscape(name)
	if err := p.http.get(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("get plugin %q: %w", name, err)
	}
	return &result, nil
}

// Reload reloads a plugin from disk.
func (p *PluginsResource) Reload(ctx context.Context, name string) (*PluginActionResponse, error) {
	var result PluginActionResponse
	path := "/api/v1/admin/plugins/" + url.PathEscape(name) + "/reload"
	if err := p.http.post(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("reload plugin %q: %w", name, err)
	}
	return &result, nil
}

// Enable enables a disabled plugin.
func (p *PluginsResource) Enable(ctx context.Context, name string) (*PluginStateResponse, error) {
	var result PluginStateResponse
	path := "/api/v1/admin/plugins/" + url.PathEscape(name) + "/enable"
	if err := p.http.post(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("enable plugin %q: %w", name, err)
	}
	return &result, nil
}

// Disable disables an active plugin.
func (p *PluginsResource) Disable(ctx context.Context, name string) (*PluginStateResponse, error) {
	var result PluginStateResponse
	path := "/api/v1/admin/plugins/" + url.PathEscape(name) + "/disable"
	if err := p.http.post(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("disable plugin %q: %w", name, err)
	}
	return &result, nil
}

// CleanupDryRun returns a list of orphaned plugin tables without dropping them.
func (p *PluginsResource) CleanupDryRun(ctx context.Context) (*CleanupDryRunResponse, error) {
	var result CleanupDryRunResponse
	if err := p.http.get(ctx, "/api/v1/admin/plugins/cleanup", nil, &result); err != nil {
		return nil, fmt.Errorf("cleanup dry-run: %w", err)
	}
	return &result, nil
}

// CleanupDrop drops orphaned plugin tables.
func (p *PluginsResource) CleanupDrop(ctx context.Context, params CleanupDropParams) (*CleanupDropResponse, error) {
	var result CleanupDropResponse
	if err := p.http.post(ctx, "/api/v1/admin/plugins/cleanup", params, &result); err != nil {
		return nil, fmt.Errorf("cleanup drop: %w", err)
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// PluginRoutesResource — route approval
// ---------------------------------------------------------------------------

// PluginRoutesResource provides plugin route approval operations.
type PluginRoutesResource struct {
	http *httpClient
}

// List returns all plugin-registered routes with their approval status.
func (r *PluginRoutesResource) List(ctx context.Context) ([]PluginRoute, error) {
	var envelope routeListEnvelope
	if err := r.http.get(ctx, "/api/v1/admin/plugins/routes", nil, &envelope); err != nil {
		return nil, fmt.Errorf("list plugin routes: %w", err)
	}
	return envelope.Routes, nil
}

// Approve approves one or more plugin routes.
func (r *PluginRoutesResource) Approve(ctx context.Context, routes []RouteApprovalItem) error {
	body := routeApprovalBody{Routes: routes}
	if err := r.http.post(ctx, "/api/v1/admin/plugins/routes/approve", body, nil); err != nil {
		return fmt.Errorf("approve plugin routes: %w", err)
	}
	return nil
}

// Revoke revokes approval for one or more plugin routes.
func (r *PluginRoutesResource) Revoke(ctx context.Context, routes []RouteApprovalItem) error {
	body := routeApprovalBody{Routes: routes}
	if err := r.http.post(ctx, "/api/v1/admin/plugins/routes/revoke", body, nil); err != nil {
		return fmt.Errorf("revoke plugin routes: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// PluginHooksResource — hook approval
// ---------------------------------------------------------------------------

// PluginHooksResource provides plugin hook approval operations.
type PluginHooksResource struct {
	http *httpClient
}

// List returns all plugin-registered hooks with their approval status.
func (h *PluginHooksResource) List(ctx context.Context) ([]PluginHook, error) {
	var envelope hookListEnvelope
	if err := h.http.get(ctx, "/api/v1/admin/plugins/hooks", nil, &envelope); err != nil {
		return nil, fmt.Errorf("list plugin hooks: %w", err)
	}
	return envelope.Hooks, nil
}

// Approve approves one or more plugin hooks.
func (h *PluginHooksResource) Approve(ctx context.Context, hooks []HookApprovalItem) error {
	body := hookApprovalBody{Hooks: hooks}
	if err := h.http.post(ctx, "/api/v1/admin/plugins/hooks/approve", body, nil); err != nil {
		return fmt.Errorf("approve plugin hooks: %w", err)
	}
	return nil
}

// Revoke revokes approval for one or more plugin hooks.
func (h *PluginHooksResource) Revoke(ctx context.Context, hooks []HookApprovalItem) error {
	body := hookApprovalBody{Hooks: hooks}
	if err := h.http.post(ctx, "/api/v1/admin/plugins/hooks/revoke", body, nil); err != nil {
		return fmt.Errorf("revoke plugin hooks: %w", err)
	}
	return nil
}
