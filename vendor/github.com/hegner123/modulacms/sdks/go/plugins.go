package modula

import (
	"context"
	"fmt"
	"net/url"
)

// ---------------------------------------------------------------------------
// Plugin types
// ---------------------------------------------------------------------------

// PluginListItem is a summary of an installed plugin, returned by [PluginsResource.List].
type PluginListItem struct {
	// Name is the unique plugin identifier (matches the plugin directory name).
	Name string `json:"name"`
	// Version is the semver version declared in the plugin manifest.
	Version string `json:"version"`
	// Description is a human-readable summary from the plugin manifest.
	Description string `json:"description"`
	// State is the plugin's lifecycle state: "active", "disabled", or "failed".
	State string `json:"state"`
	// CircuitBreakerState is the fault isolation state: "closed" (healthy),
	// "open" (tripped, calls rejected), or "half-open" (testing recovery).
	// Empty when the plugin has no circuit breaker activity.
	CircuitBreakerState string `json:"circuit_breaker_state,omitempty"`
}

// DriftEntry describes a schema drift detection result for a plugin's database tables.
// Drift occurs when the actual database schema differs from what the plugin manifest declares.
type DriftEntry struct {
	// Table is the database table where drift was detected.
	Table string `json:"table"`
	// Kind is the drift type: "missing" (declared but absent) or "extra" (present but undeclared).
	Kind string `json:"kind"`
	// Column is the specific column that is missing or extra.
	Column string `json:"column"`
}

// PluginInfo is detailed information for a single plugin, returned by [PluginsResource.Get].
// It extends [PluginListItem] with diagnostic fields for troubleshooting.
type PluginInfo struct {
	// Name is the unique plugin identifier (matches the plugin directory name).
	Name string `json:"name"`
	// Version is the semver version declared in the plugin manifest.
	Version string `json:"version"`
	// Description is a human-readable summary from the plugin manifest.
	Description string `json:"description"`
	// Author is the plugin author, if declared in the manifest.
	Author string `json:"author,omitempty"`
	// License is the plugin license identifier (e.g., "MIT"), if declared.
	License string `json:"license,omitempty"`
	// State is the plugin's lifecycle state: "active", "disabled", or "failed".
	State string `json:"state"`
	// FailedReason describes why the plugin entered the "failed" state.
	// Empty when the plugin is active or disabled.
	FailedReason string `json:"failed_reason,omitempty"`
	// CircuitBreakerState is the fault isolation state: "closed", "open", or "half-open".
	CircuitBreakerState string `json:"circuit_breaker_state,omitempty"`
	// CircuitBreakerErrs is the number of consecutive errors that contributed to
	// the circuit breaker state. Resets to 0 when the breaker closes.
	CircuitBreakerErrs int `json:"circuit_breaker_errors,omitempty"`
	// VMsAvailable is the number of idle Lua VMs in the pool ready to handle requests.
	VMsAvailable int `json:"vms_available"`
	// VMsTotal is the total number of Lua VMs allocated for this plugin.
	VMsTotal int `json:"vms_total"`
	// Dependencies lists other plugin names this plugin depends on.
	Dependencies []string `json:"dependencies,omitempty"`
	// SchemaDrift lists database schema mismatches between the plugin manifest
	// and the actual database state. Non-empty drift may indicate the plugin
	// needs a migration or was partially installed.
	SchemaDrift []DriftEntry `json:"schema_drift,omitempty"`
}

// PluginActionResponse is the response from [PluginsResource.Reload].
type PluginActionResponse struct {
	// OK is true if the action completed successfully.
	OK bool `json:"ok"`
	// Plugin is the name of the plugin that was acted upon.
	Plugin string `json:"plugin"`
}

// PluginStateResponse is the response from [PluginsResource.Enable] and [PluginsResource.Disable].
type PluginStateResponse struct {
	// OK is true if the state change completed successfully.
	OK bool `json:"ok"`
	// Plugin is the name of the plugin that was acted upon.
	Plugin string `json:"plugin"`
	// State is the new lifecycle state after the operation ("active" or "disabled").
	State string `json:"state"`
}

// CleanupDryRunResponse is the response from [PluginsResource.CleanupDryRun],
// previewing which orphaned plugin tables would be dropped.
type CleanupDryRunResponse struct {
	// OrphanedTables lists table names that exist in the database but are not
	// claimed by any installed plugin.
	OrphanedTables []string `json:"orphaned_tables"`
	// Count is the number of orphaned tables found.
	Count int `json:"count"`
	// Action describes what would happen (always "dry_run" for this response).
	Action string `json:"action"`
}

// CleanupDropParams holds parameters for [PluginsResource.CleanupDrop].
type CleanupDropParams struct {
	// Confirm must be true to actually drop the tables. The server rejects
	// the request if this is false, as a safety measure.
	Confirm bool `json:"confirm"`
	// Tables is the list of orphaned table names to drop. These should come
	// from a prior [PluginsResource.CleanupDryRun] call.
	Tables []string `json:"tables"`
}

// CleanupDropResponse is the response from [PluginsResource.CleanupDrop],
// confirming which tables were actually dropped.
type CleanupDropResponse struct {
	// Dropped lists the table names that were successfully dropped.
	Dropped []string `json:"dropped"`
	// Count is the number of tables that were dropped.
	Count int `json:"count"`
}

// PluginRoute describes an HTTP route registered by a plugin. Plugin routes require
// admin approval before they become active and start serving traffic.
type PluginRoute struct {
	// Plugin is the name of the plugin that registered this route.
	Plugin string `json:"plugin"`
	// Method is the HTTP method (GET, POST, PUT, DELETE, etc.).
	Method string `json:"method"`
	// Path is the URL path pattern for this route.
	Path string `json:"path"`
	// Public is true if the route does not require authentication.
	Public bool `json:"public"`
	// Approved is true if an admin has approved this route for serving.
	// Unapproved routes are registered but return 403 to callers.
	Approved bool `json:"approved"`
}

// RouteApprovalItem identifies a specific plugin route for approval or revocation.
// The combination of Plugin, Method, and Path uniquely identifies a route.
type RouteApprovalItem struct {
	// Plugin is the name of the plugin that owns the route.
	Plugin string `json:"plugin"`
	// Method is the HTTP method of the route to approve/revoke.
	Method string `json:"method"`
	// Path is the URL path of the route to approve/revoke.
	Path string `json:"path"`
}

// PluginHook describes a database event hook registered by a plugin. Hooks fire
// on content mutations (create, update, delete) and require admin approval before
// they become active.
type PluginHook struct {
	// PluginName is the name of the plugin that registered this hook.
	PluginName string `json:"plugin_name"`
	// Event is the mutation event type (e.g., "before_create", "after_update", "after_delete").
	Event string `json:"event"`
	// Table is the database table this hook listens on (e.g., "content_data", "media").
	Table string `json:"table"`
	// Priority determines execution order when multiple hooks fire on the same event.
	// Lower values execute first.
	Priority int `json:"priority"`
	// Approved is true if an admin has approved this hook for execution.
	// Unapproved hooks are registered but do not fire.
	Approved bool `json:"approved"`
	// IsWildcard is true if the hook listens on all tables (Table is "*").
	IsWildcard bool `json:"is_wildcard"`
}

// HookApprovalItem identifies a specific plugin hook for approval or revocation.
// The combination of Plugin, Event, and Table uniquely identifies a hook.
type HookApprovalItem struct {
	// Plugin is the name of the plugin that owns the hook.
	Plugin string `json:"plugin"`
	// Event is the mutation event type of the hook to approve/revoke.
	Event string `json:"event"`
	// Table is the database table of the hook to approve/revoke.
	Table string `json:"table"`
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

// PluginsResource provides plugin lifecycle management operations including
// listing, inspecting, enabling, disabling, reloading, and orphaned table cleanup.
// Plugins are Lua scripts that extend CMS functionality with custom routes, hooks,
// and database tables. Each plugin runs in an isolated Lua VM pool with a circuit
// breaker for fault isolation.
// It is accessed via [Client].Plugins.
type PluginsResource struct {
	http *httpClient
}

// List returns a summary of all installed plugins, including their current state
// and circuit breaker status. Both active and disabled plugins are included.
func (p *PluginsResource) List(ctx context.Context) ([]PluginListItem, error) {
	var envelope pluginListEnvelope
	if err := p.http.get(ctx, "/api/v1/admin/plugins", nil, &envelope); err != nil {
		return nil, fmt.Errorf("list plugins: %w", err)
	}
	return envelope.Plugins, nil
}

// Get returns detailed information for a specific plugin by name, including
// VM pool stats, schema drift analysis, dependencies, and failure diagnostics.
// Returns an [*ApiError] with status 404 if no plugin with the given name is installed.
func (p *PluginsResource) Get(ctx context.Context, name string) (*PluginInfo, error) {
	var result PluginInfo
	path := "/api/v1/admin/plugins/" + url.PathEscape(name)
	if err := p.http.get(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("get plugin %q: %w", name, err)
	}
	return &result, nil
}

// Reload reloads a plugin from disk, re-reading its manifest and Lua source files,
// and reinitializing its VM pool. Use this after modifying plugin files on the server.
// The plugin is briefly unavailable during reload. Returns an [*ApiError] if the
// plugin fails to reinitialize (e.g., syntax errors in Lua source).
func (p *PluginsResource) Reload(ctx context.Context, name string) (*PluginActionResponse, error) {
	var result PluginActionResponse
	path := "/api/v1/admin/plugins/" + url.PathEscape(name) + "/reload"
	if err := p.http.post(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("reload plugin %q: %w", name, err)
	}
	return &result, nil
}

// Enable activates a disabled plugin, starting its VM pool and registering
// its routes and hooks. The plugin must be in the "disabled" state.
func (p *PluginsResource) Enable(ctx context.Context, name string) (*PluginStateResponse, error) {
	var result PluginStateResponse
	path := "/api/v1/admin/plugins/" + url.PathEscape(name) + "/enable"
	if err := p.http.post(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("enable plugin %q: %w", name, err)
	}
	return &result, nil
}

// Disable deactivates an active plugin, stopping its VM pool and unregistering
// its routes and hooks. The plugin remains installed but does not process any requests.
func (p *PluginsResource) Disable(ctx context.Context, name string) (*PluginStateResponse, error) {
	var result PluginStateResponse
	path := "/api/v1/admin/plugins/" + url.PathEscape(name) + "/disable"
	if err := p.http.post(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("disable plugin %q: %w", name, err)
	}
	return &result, nil
}

// CleanupDryRun scans the database for tables that were created by plugins that are
// no longer installed. Returns a preview of which tables would be dropped without
// actually dropping them. Use this to inspect before calling [PluginsResource.CleanupDrop].
func (p *PluginsResource) CleanupDryRun(ctx context.Context) (*CleanupDryRunResponse, error) {
	var result CleanupDryRunResponse
	if err := p.http.get(ctx, "/api/v1/admin/plugins/cleanup", nil, &result); err != nil {
		return nil, fmt.Errorf("cleanup dry-run: %w", err)
	}
	return &result, nil
}

// CleanupDrop permanently drops orphaned plugin tables from the database.
// The params must include Confirm: true and a Tables list (typically from a prior
// [PluginsResource.CleanupDryRun] call). This operation is irreversible.
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

// PluginRoutesResource provides plugin route approval and management operations.
// Plugin-registered HTTP routes require explicit admin approval before they become
// active. Unapproved routes exist in the registry but return 403 to callers.
// It is accessed via [Client].PluginRoutes.
type PluginRoutesResource struct {
	http *httpClient
}

// List returns all plugin-registered HTTP routes across all plugins,
// including their approval status. Both approved and unapproved routes are returned.
func (r *PluginRoutesResource) List(ctx context.Context) ([]PluginRoute, error) {
	var envelope routeListEnvelope
	if err := r.http.get(ctx, "/api/v1/admin/plugins/routes", nil, &envelope); err != nil {
		return nil, fmt.Errorf("list plugin routes: %w", err)
	}
	return envelope.Routes, nil
}

// Approve grants approval to one or more plugin routes, allowing them to serve
// traffic. Each route is identified by its plugin name, HTTP method, and path.
func (r *PluginRoutesResource) Approve(ctx context.Context, routes []RouteApprovalItem) error {
	body := routeApprovalBody{Routes: routes}
	if err := r.http.post(ctx, "/api/v1/admin/plugins/routes/approve", body, nil); err != nil {
		return fmt.Errorf("approve plugin routes: %w", err)
	}
	return nil
}

// Revoke removes approval from one or more plugin routes, causing them to
// return 403 to callers until re-approved. The routes remain registered but inactive.
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

// PluginHooksResource provides plugin hook approval and management operations.
// Plugin-registered database hooks require explicit admin approval before they
// fire on content mutations. Unapproved hooks are registered but do not execute.
// It is accessed via [Client].PluginHooks.
type PluginHooksResource struct {
	http *httpClient
}

// List returns all plugin-registered database hooks across all plugins,
// including their approval status, priority, and wildcard flag.
func (h *PluginHooksResource) List(ctx context.Context) ([]PluginHook, error) {
	var envelope hookListEnvelope
	if err := h.http.get(ctx, "/api/v1/admin/plugins/hooks", nil, &envelope); err != nil {
		return nil, fmt.Errorf("list plugin hooks: %w", err)
	}
	return envelope.Hooks, nil
}

// Approve grants approval to one or more plugin hooks, allowing them to fire
// on content mutations. Each hook is identified by its plugin name, event type, and table.
func (h *PluginHooksResource) Approve(ctx context.Context, hooks []HookApprovalItem) error {
	body := hookApprovalBody{Hooks: hooks}
	if err := h.http.post(ctx, "/api/v1/admin/plugins/hooks/approve", body, nil); err != nil {
		return fmt.Errorf("approve plugin hooks: %w", err)
	}
	return nil
}

// Revoke removes approval from one or more plugin hooks, preventing them from
// firing on content mutations. The hooks remain registered but inactive until re-approved.
func (h *PluginHooksResource) Revoke(ctx context.Context, hooks []HookApprovalItem) error {
	body := hookApprovalBody{Hooks: hooks}
	if err := h.http.post(ctx, "/api/v1/admin/plugins/hooks/revoke", body, nil); err != nil {
		return fmt.Errorf("revoke plugin hooks: %w", err)
	}
	return nil
}
