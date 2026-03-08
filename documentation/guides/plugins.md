# Managing Plugins

This guide covers how to install, configure, approve, monitor, and troubleshoot plugins as a CMS administrator. For building plugins, see the [Plugin Developer Documentation](../plugins/overview.md).

## What Plugins Do

Plugins extend ModulaCMS with custom functionality provided by third-party developers or your own team. A plugin can:

- Add custom REST API endpoints under `/api/v1/plugins/<plugin_name>/`
- React to content lifecycle events (create, update, delete, publish, archive)
- Store its own data in isolated database tables

Plugins run in sandboxed environments and cannot access the filesystem, execute system commands, or read core CMS database tables. Each plugin's data is isolated from other plugins and from the CMS itself.

## Enabling the Plugin System

The plugin system is disabled by default. Enable it in `config.json`:

```json
{
  "plugin_enabled": true,
  "plugin_directory": "./plugins/"
}
```

`plugin_directory` is the path where ModulaCMS scans for plugin directories. Each subdirectory containing an `init.lua` file is treated as a plugin.

Restart the server after changing `plugin_enabled`.

## Installing a Plugin

Copy the plugin's directory into your `plugin_directory`:

```
plugins/
  task_tracker/
    init.lua
    lib/
      helpers.lua
```

The server picks up new plugins on startup. If `plugin_hot_reload` is enabled, the server detects new plugins automatically without a restart.

### Validating Before Installation

Check that a plugin is well-formed before deploying it:

```bash
modula plugin validate ./plugins/task_tracker
```

This checks the manifest (`plugin_info` table in `init.lua`), file structure, and basic syntax without loading the plugin into the server.

## Configuration

### Essential Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_enabled` | bool | false | Master switch for the plugin system |
| `plugin_directory` | string | `"./plugins/"` | Path to scan for plugin directories |
| `plugin_hot_reload` | bool | false | Auto-reload plugins when Lua files change |

### Resource Limits

These settings protect the CMS from misbehaving plugins:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_max_vms` | int | 4 | VM pool size per plugin |
| `plugin_timeout` | int | 5 | Execution timeout per handler (seconds) |
| `plugin_max_ops` | int | 1000 | Max database operations per request |
| `plugin_rate_limit` | int | 100 | Max requests per second per IP |
| `plugin_max_routes` | int | 50 | Max HTTP routes per plugin |
| `plugin_max_request_body` | int | 1048576 | Max request body size (bytes, default 1 MB) |
| `plugin_max_response_body` | int | 5242880 | Max response body size (bytes, default 5 MB) |

### Circuit Breaker Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_max_failures` | int | 5 | Consecutive failures before the plugin is disabled |
| `plugin_reset_interval` | string | `"60s"` | Wait time before retrying a disabled plugin |
| `plugin_hook_max_consecutive_aborts` | int | 10 | Consecutive hook errors before that hook is disabled |

### Hook Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_hook_reserve_vms` | int | 1 | VMs reserved for hook execution |
| `plugin_hook_max_ops` | int | 100 | Max database operations per hook |
| `plugin_hook_max_concurrent_after` | int | 10 | Max concurrent after-hook goroutines |
| `plugin_hook_timeout_ms` | int | 2000 | Per-hook execution timeout (ms) |
| `plugin_hook_event_timeout_ms` | int | 5000 | Total timeout for all hooks on one event (ms) |

### Proxy Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_trusted_proxies` | []string | [] | IP ranges trusted for X-Forwarded-For parsing |

## Route and Hook Approval

Plugins register HTTP routes and content hooks, but none of them are active until an admin approves them. This is a security gate -- new or updated plugin code does not execute until explicitly allowed.

When a plugin's version changes (in its `plugin_info.version` field), all existing approvals are revoked and must be re-approved.

### Viewing Pending Approvals

```bash
# List all routes with approval status
curl http://localhost:8080/api/v1/admin/plugins/routes \
  -H "Cookie: session=YOUR_SESSION_COOKIE"

# List all hooks with approval status
curl http://localhost:8080/api/v1/admin/plugins/hooks \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Approving via CLI

```bash
# Approve all routes and hooks for a plugin
modula plugin approve task_tracker --all-routes
modula plugin approve task_tracker --all-hooks

# Approve specific items
modula plugin approve task_tracker --route "GET /tasks"
modula plugin approve task_tracker --hook "before_create:content_data"

# Skip confirmation prompts (for CI/CD)
modula plugin approve task_tracker --all-routes --all-hooks --yes
```

### Approving via API

```bash
# Approve routes
curl -X POST http://localhost:8080/api/v1/admin/plugins/routes/approve \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "routes": [
      {"plugin": "task_tracker", "method": "GET", "path": "/tasks"},
      {"plugin": "task_tracker", "method": "POST", "path": "/tasks"}
    ]
  }'

# Approve hooks
curl -X POST http://localhost:8080/api/v1/admin/plugins/hooks/approve \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "hooks": [
      {"plugin": "task_tracker", "event": "after_create", "table": "content_data"}
    ]
  }'
```

### Approving via TUI

Navigate to the Plugins page from the homepage menu. Select a plugin to view its routes and hooks, then approve through the confirmation dialog.

### Revoking Approvals

```bash
modula plugin revoke task_tracker --all-routes
modula plugin revoke task_tracker --route "GET /tasks"
```

## Monitoring

### Plugin Status

```bash
modula plugin info task_tracker
```

This shows:
- Current lifecycle state (Discovered, Loading, Running, Failed, Stopped)
- Circuit breaker state (Closed, Open, Half-Open)
- VM pool utilization
- Route and hook counts with approval status
- Schema drift warnings (if table definitions don't match actual database columns)

### Via API

```bash
# List all plugins with state
curl http://localhost:8080/api/v1/admin/plugins \
  -H "Cookie: session=YOUR_SESSION_COOKIE"

# Detailed info for one plugin
curl http://localhost:8080/api/v1/admin/plugins/task_tracker \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Listing Plugins

```bash
modula plugin list
```

## Lifecycle Management

### Enable / Disable

```bash
modula plugin enable task_tracker   # reset circuit breaker, reload
modula plugin disable task_tracker  # stop plugin, trip circuit breaker
```

Disabling a plugin immediately stops it from serving traffic. Enabling resets the circuit breaker and reloads the plugin.

### Reload

```bash
modula plugin reload task_tracker
```

Triggers a blue-green reload: a new instance is created alongside the old one. If it loads successfully, it replaces the old instance atomically. If it fails, the old instance keeps running.

### Hot Reload

When `plugin_hot_reload` is `true`, the server watches for Lua file changes every 2 seconds and reloads automatically. Safety limits prevent reload storms:

- 1-second debounce window
- 10-second cooldown between reloads per plugin
- After 3 consecutive slow reloads (>10s each), auto-reload pauses for that plugin

Hot reload is intended for development. Disable it in production.

## Circuit Breaker

The circuit breaker protects the CMS from failing plugins.

**Plugin-level**: After consecutive failures (default 5), the circuit breaker opens and all requests to the plugin return HTTP 503. After the reset interval (default 60s), a single probe request is allowed. If it succeeds, the breaker closes. If it fails, it re-opens.

| State | Behavior |
|-------|----------|
| Closed | Normal operation |
| Open | All requests return 503 |
| Half-Open | One probe request allowed |

**Hook-level**: Each (plugin, event, table) combination has its own breaker. After 10 consecutive errors (configurable), that specific hook is disabled. Hook failures do not affect the plugin-level circuit breaker.

Reset the circuit breaker manually:

```bash
modula plugin enable task_tracker
```

## Cleanup

Over time, plugins may leave behind orphaned database tables (e.g., after a plugin is removed but its tables remain). The cleanup endpoints handle this.

```bash
# Dry run: list orphaned tables without deleting
curl http://localhost:8080/api/v1/admin/plugins/cleanup \
  -H "Cookie: session=YOUR_SESSION_COOKIE"

# Drop orphaned tables (requires confirmation)
curl -X POST http://localhost:8080/api/v1/admin/plugins/cleanup \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"tables": ["plugin_old_plugin_tasks", "plugin_old_plugin_categories"]}'
```

## Troubleshooting

**Plugin shows "Failed" state**: Check the server logs for the `on_init()` error. Common causes: missing table definition, invalid SQL, dependency plugin not loaded.

**Routes return 404**: Routes must be approved before they serve traffic. Check `modula plugin info <name>` for unapproved routes.

**Hooks not firing**: Hooks must be approved. Check hook approval status. Also verify the event and table name match what the plugin registered.

**Circuit breaker tripped**: Check logs for the error pattern. Fix the underlying issue, then `modula plugin enable <name>` to reset.

**Schema drift warnings**: The plugin's table definition doesn't match the actual database columns. This is advisory -- the plugin still works, but the developer should update the definition or migrate the table.

**Rate limiting (429 responses)**: Increase `plugin_rate_limit` or investigate why the client is sending too many requests.

## Admin API Reference

All endpoints require authentication. Admin endpoints require the `plugins:admin` permission; read endpoints require `plugins:read`.

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/admin/plugins` | `plugins:read` | List all plugins with state |
| GET | `/api/v1/admin/plugins/{name}` | `plugins:read` | Plugin details and circuit breaker state |
| POST | `/api/v1/admin/plugins/{name}/reload` | `plugins:admin` | Trigger hot reload |
| POST | `/api/v1/admin/plugins/{name}/enable` | `plugins:admin` | Reset circuit breaker, reload |
| POST | `/api/v1/admin/plugins/{name}/disable` | `plugins:admin` | Stop plugin |
| GET | `/api/v1/admin/plugins/cleanup` | `plugins:admin` | List orphaned tables (dry run) |
| POST | `/api/v1/admin/plugins/cleanup` | `plugins:admin` | Drop orphaned tables |
| GET | `/api/v1/admin/plugins/routes` | `plugins:read` | List routes with approval status |
| POST | `/api/v1/admin/plugins/routes/approve` | `plugins:admin` | Approve routes |
| POST | `/api/v1/admin/plugins/routes/revoke` | `plugins:admin` | Revoke route approvals |
| GET | `/api/v1/admin/plugins/hooks` | `plugins:read` | List hooks with approval status |
| POST | `/api/v1/admin/plugins/hooks/approve` | `plugins:admin` | Approve hooks |
| POST | `/api/v1/admin/plugins/hooks/revoke` | `plugins:admin` | Revoke hook approvals |

## CLI Reference

| Command | Server Required | Description |
|---------|----------------|-------------|
| `modula plugin list` | No | List discovered plugins |
| `modula plugin validate <path>` | No | Validate a plugin without loading |
| `modula plugin info <name>` | Yes | Plugin status, circuit breaker, approvals |
| `modula plugin reload <name>` | Yes | Blue-green hot reload |
| `modula plugin enable <name>` | Yes | Reset circuit breaker, reload |
| `modula plugin disable <name>` | Yes | Stop plugin |
| `modula plugin approve <name> [flags]` | Yes | Approve routes/hooks |
| `modula plugin revoke <name> [flags]` | Yes | Revoke approvals |
