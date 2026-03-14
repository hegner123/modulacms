# Plugin Configuration

All plugin settings live in `modula.config.json`. This page documents every field with detailed explanations of behavior and tuning guidance.

## Complete Configuration Block

```json
{
  "plugin_enabled": true,
  "plugin_directory": "./plugins/",
  "plugin_max_vms": 4,
  "plugin_timeout": 5,
  "plugin_max_ops": 1000,
  "plugin_hot_reload": false,
  "plugin_max_failures": 5,
  "plugin_reset_interval": "60s",
  "plugin_rate_limit": 100,
  "plugin_max_routes": 50,
  "plugin_max_request_body": 1048576,
  "plugin_max_response_body": 5242880,
  "plugin_trusted_proxies": [],
  "plugin_hook_reserve_vms": 1,
  "plugin_hook_max_consecutive_aborts": 10,
  "plugin_hook_max_ops": 100,
  "plugin_hook_max_concurrent_after": 10,
  "plugin_hook_timeout_ms": 2000,
  "plugin_hook_event_timeout_ms": 5000
}
```

## Master Switch and Directory

### plugin_enabled

| Type | Default |
|------|---------|
| bool | `false` |

Master switch for the entire plugin system. When `false`, no plugins are loaded, no routes are registered, and no hooks fire. The plugin directory is not scanned.

Set to `true` to enable plugin loading on server startup.

### plugin_directory

| Type | Default |
|------|---------|
| string | `"./plugins/"` |

Path to the directory scanned for plugin subdirectories. Can be absolute or relative to the CMS working directory. Each subdirectory containing an `init.lua` file is treated as a plugin candidate.

## VM Pool Tuning

### plugin_max_vms

| Type | Default | Range |
|------|---------|-------|
| int | `4` | 1+ |

Total VM pool size per plugin for HTTP and hook execution. This count is split between two channels:

- **General channel**: `plugin_max_vms - plugin_hook_reserve_vms` VMs. Serves HTTP request handlers.
- **Reserved channel**: `plugin_hook_reserve_vms` VMs. Serves content hooks exclusively.

The two-channel design guarantees hook execution even when HTTP traffic saturates the general pool. HTTP requests that cannot acquire a VM within 100ms receive a 503 response.

A separate UI VM pool (`plugin_max_ui_vms`) is used for TUI screens and field interfaces. See below.

**Tuning guidance:**
- For plugins with light HTTP traffic and few hooks: `2` is sufficient.
- For plugins under heavy HTTP load: increase to `8-16`. Each VM is a gopher-lua state with its own stack and globals, consuming roughly 1-2 MB of memory.
- The value must be greater than `plugin_hook_reserve_vms`.

### plugin_hook_reserve_vms

| Type | Default | Range |
|------|---------|-------|
| int | `1` | 0+ |

Number of VMs reserved exclusively for hook execution. These VMs are never used for HTTP requests.

Setting this to `0` means hooks compete with HTTP requests for the same VM pool. This is acceptable for plugins that do not register hooks, but risks hook starvation under HTTP load for plugins that do.

Must be less than `plugin_max_vms`.

### plugin_max_ui_vms

| Type | Default | Range |
|------|---------|-------|
| int | `4` | 1+ |

Maximum concurrent TUI screen and field interface sessions per plugin. Each active plugin screen or field interface in the SSH TUI holds one UI VM for the lifetime of the session.

UI VMs are in a separate pool from the general and reserved pools. They have no acquisition timeout -- if the pool is full, the screen shows "Plugin busy" instead of blocking.

In a multi-user SSH environment, each concurrent operator viewing a plugin screen or editing a plugin field consumes one UI VM. With the default of 4, four operators can use plugin UIs simultaneously per plugin.

Only created for plugins that declare `screens` or `interfaces` in their manifest. Plugins without UI declarations do not allocate UI VMs.

## Timeout Settings

### plugin_timeout

| Type | Default | Unit |
|------|---------|------|
| int | `5` | seconds |

Maximum execution time for a single HTTP request handler or after-hook. If the Lua code does not return within this duration, the VM execution is killed and an error is returned (500 for HTTP, logged for hooks).

This is the outer timeout. Individual database operations within the handler have their own timeouts derived from this value.

### plugin_hook_timeout_ms

| Type | Default | Unit |
|------|---------|------|
| int | `2000` | milliseconds |

Maximum execution time for a single before-hook handler. Before-hooks run synchronously inside the CMS database transaction, so this timeout is intentionally short to prevent long-running hooks from blocking CMS operations.

If a before-hook exceeds this timeout, it is treated as an error: the hook's circuit breaker is incremented, and the transaction is aborted.

### plugin_hook_event_timeout_ms

| Type | Default | Unit |
|------|---------|------|
| int | `5000` | milliseconds |

Maximum total time for the entire before-hook chain on a single event. Multiple plugins may register before-hooks on the same event and table. This timeout caps the total time spent running all of them.

If the chain exceeds this timeout, remaining hooks are skipped and the transaction is aborted.

This timeout applies to the chain as a whole, not to individual hooks. A chain of 5 hooks that each take 900ms would exceed the 5000ms event timeout, even though each individual hook is under the 2000ms per-hook timeout.

## Operation Budgets

### plugin_max_ops

| Type | Default |
|------|---------|
| int | `1000` |

Maximum number of database operations (`db.insert`, `db.query`, `db.update`, `db.delete`, etc.) allowed per HTTP request handler execution. Each `db.*` call decrements the budget. When the budget reaches zero, subsequent `db.*` calls raise a Lua error.

This prevents runaway queries from monopolizing database connections. A well-designed handler typically uses 5-20 operations.

### plugin_hook_max_ops

| Type | Default |
|------|---------|
| int | `100` |

Maximum database operations per after-hook execution. After-hooks run asynchronously with fire-and-forget semantics, so the budget is lower than for HTTP handlers.

**Before-hooks always have a budget of zero.** All `db.*` calls are blocked inside before-hooks because they run inside the CMS database transaction, and plugin database operations use a separate connection pool. Calling `db.*` in a before-hook would deadlock on SQLite (single-writer) and risk long-held locks on MySQL/PostgreSQL.

## Rate Limiting

### plugin_rate_limit

| Type | Default | Unit |
|------|---------|------|
| int | `100` | requests per second |

Per-IP token bucket rate limit for plugin HTTP endpoints. Each unique client IP gets its own bucket.

When a client exceeds the rate, the plugin bridge returns HTTP 429 (Too Many Requests).

Implementation details:
- Token bucket with burst equal to the rate (e.g., 100 tokens, refilling at 100/s).
- Inactive entries (no requests for 10 minutes) are cleaned up every 5 minutes to prevent memory growth.
- The rate applies per-IP across all plugins, not per-plugin.

### plugin_trusted_proxies

| Type | Default |
|------|---------|
| []string | `[]` |

CIDR ranges for trusted reverse proxies. When a request arrives from a trusted proxy IP, the `X-Forwarded-For` header is parsed to extract the real client IP for rate limiting and `req.client_ip`.

```json
{
  "plugin_trusted_proxies": ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
}
```

If empty, `req.client_ip` uses the direct connection IP.

## Size Limits

### plugin_max_routes

| Type | Default |
|------|---------|
| int | `50` |

Maximum number of HTTP routes a single plugin can register. Attempting to register more routes raises an error during plugin loading.

### plugin_max_request_body

| Type | Default | Unit |
|------|---------|------|
| int | `1048576` | bytes (1 MB) |

Maximum request body size for plugin HTTP endpoints. Requests with a body exceeding this size are rejected with HTTP 413 (Request Entity Too Large) before the handler executes.

### plugin_max_response_body

| Type | Default | Unit |
|------|---------|------|
| int | `5242880` | bytes (5 MB) |

Maximum response body size for plugin HTTP endpoints. Responses exceeding this size are truncated.

## Circuit Breaker Tuning

### plugin_max_failures

| Type | Default |
|------|---------|
| int | `5` |

Consecutive failures before the plugin-level circuit breaker trips. Failures include HTTP handler errors and manager operation errors (reload, init). Hook failures do not count.

When the circuit breaker is open, all HTTP requests to the plugin return 503 until the reset interval elapses.

### plugin_reset_interval

| Type | Default |
|------|---------|
| string | `"60s"` |

Duration before the circuit breaker transitions from open to half-open. Accepts Go duration strings: `"30s"`, `"2m"`, `"1m30s"`.

In the half-open state, one probe request is allowed through. If it succeeds, the breaker closes (normal operation). If it fails, the breaker re-opens for another reset interval.

An admin can force-reset the circuit breaker at any time with `modulacms plugin enable <name>`.

### plugin_hook_max_consecutive_aborts

| Type | Default |
|------|---------|
| int | `10` |

Consecutive errors before a hook-level circuit breaker trips. Each (plugin, event, table) combination has its own breaker. When tripped, that specific hook is disabled until the plugin is reloaded or re-enabled.

Hook circuit breakers are independent of the plugin-level circuit breaker.

### plugin_hook_max_concurrent_after

| Type | Default |
|------|---------|
| int | `10` |

Maximum number of after-hook goroutines running concurrently across all plugins. After-hooks are fire-and-forget, but this semaphore prevents unbounded goroutine growth during burst writes.

When the limit is reached, additional after-hooks block until a slot becomes available or the shutdown context is cancelled.

## Hot Reload Settings

### plugin_hot_reload

| Type | Default |
|------|---------|
| bool | `false` |

Enable file-polling watcher for live plugin reload during development. When `true`, the system polls for `.lua` file changes every 2 seconds.

Change detection uses SHA-256 checksums of all `.lua` files in the plugin directory. When a checksum changes, the plugin undergoes blue-green reload: a new instance is created alongside the old one. If the new instance loads successfully, it replaces the old one atomically. If the new instance fails, the old one keeps running.

**Safety limits (not configurable):**

| Limit | Value | Purpose |
|-------|-------|---------|
| Debounce delay | 1 second | Wait for file writes to settle before reloading |
| Reload cooldown | 10 seconds per plugin | Prevent reload storms during rapid iteration |
| Max .lua files per plugin | 100 | Prevent DoS via file count during checksumming |
| Max total size per checksum | 10 MB | Prevent DoS via file size during checksumming |
| Max consecutive slow reloads | 3 (>10s each) | Pause watcher for plugin on systemic issues |

**Do not enable in production.** File polling adds overhead, and the debounce/cooldown windows create brief periods where changes are not yet live. Use `modulacms plugin reload <name>` for controlled production updates.

## Example Configurations

### Development

Generous limits, hot reload enabled, fast circuit breaker reset for iterative development:

```json
{
  "plugin_enabled": true,
  "plugin_directory": "./plugins/",
  "plugin_max_vms": 2,
  "plugin_timeout": 10,
  "plugin_max_ops": 5000,
  "plugin_hot_reload": true,
  "plugin_max_failures": 10,
  "plugin_reset_interval": "10s",
  "plugin_rate_limit": 1000,
  "plugin_max_routes": 100,
  "plugin_max_request_body": 10485760,
  "plugin_max_response_body": 10485760,
  "plugin_trusted_proxies": [],
  "plugin_hook_reserve_vms": 1,
  "plugin_hook_max_consecutive_aborts": 50,
  "plugin_hook_max_ops": 500,
  "plugin_hook_max_concurrent_after": 20,
  "plugin_hook_timeout_ms": 5000,
  "plugin_hook_event_timeout_ms": 10000
}
```

### Production

Strict limits, hot reload off, trusted proxy for load balancer:

```json
{
  "plugin_enabled": true,
  "plugin_directory": "/opt/modulacms/plugins/",
  "plugin_max_vms": 8,
  "plugin_timeout": 5,
  "plugin_max_ops": 500,
  "plugin_hot_reload": false,
  "plugin_max_failures": 3,
  "plugin_reset_interval": "120s",
  "plugin_rate_limit": 50,
  "plugin_max_routes": 30,
  "plugin_max_request_body": 524288,
  "plugin_max_response_body": 2097152,
  "plugin_trusted_proxies": ["10.0.0.0/8"],
  "plugin_hook_reserve_vms": 2,
  "plugin_hook_max_consecutive_aborts": 5,
  "plugin_hook_max_ops": 50,
  "plugin_hook_max_concurrent_after": 5,
  "plugin_hook_timeout_ms": 1000,
  "plugin_hook_event_timeout_ms": 3000
}
```

Key differences from development:
- Higher VM count (8) to handle concurrent load, with 2 reserved for hooks.
- Lower operation budgets to limit blast radius of misbehaving plugins.
- Lower rate limit per IP.
- Smaller request/response body limits.
- Shorter hook timeouts.
- Stricter circuit breaker thresholds (trip after 3 failures, 2-minute reset).
- Hot reload disabled -- use `modulacms plugin reload` for controlled updates.
