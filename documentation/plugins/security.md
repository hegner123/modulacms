# Plugin Security Model

The plugin system is designed so that a malicious or buggy plugin cannot compromise CMS stability, access data it should not, or affect other plugins. This page documents every layer of the security model.

## Lua Sandbox

Source: `internal/plugin/sandbox.go`

Each plugin runs in a restricted Lua 5.1 environment. Before any plugin code executes, the sandbox removes dangerous functions and modules.

### Removed Modules

| Module | Reason |
|--------|--------|
| `io` | No filesystem read/write access |
| `os` | No process execution, environment access, or system calls |
| `package` | No arbitrary module loading from the filesystem |
| `debug` | No VM introspection, stack manipulation, or upvalue access |

### Removed Global Functions

| Function | Reason |
|----------|--------|
| `dofile` | No dynamic code loading from files |
| `loadfile` | No dynamic code loading from files |
| `load` | No dynamic code loading from strings |
| `rawget` | Prevents bypassing frozen module metatables |
| `rawset` | Prevents bypassing frozen module metatables |
| `rawequal` | Prevents bypassing metatable-based protections |
| `rawlen` | Prevents bypassing metatable-based protections |

### What Remains

Only safe standard library functions are available. See the [Lua API reference](lua-api.md#allowed-lua-standard-library) for the complete list. In summary: `base` (type checks, iteration, error handling, metatable get/set), `string` (manipulation, pattern matching), `table` (insert, remove, sort, concat), `math` (arithmetic, random). The `coroutine` library is additionally available for plugins that declare `screens` or `interfaces` in their manifest, enabling the TUI coroutine bridge for plugin-driven terminal UI.

### Frozen Module Protection

All injected API modules (`db`, `http`, `hooks`, `log`, `tui`) are frozen read-only via a metatable proxy:

- **Writes raise an error.** `db.query = nil` or `db.custom = function() end` fails with an error.
- **`getmetatable` returns `"protected"`.** Prevents retrieving and modifying the proxy metatable.
- **`rawget`/`rawset` are removed.** The standard Lua bypass for metatables is unavailable.

This prevents plugins from monkey-patching API modules to escalate privileges or intercept other plugins' calls.

Source: `internal/plugin/sandbox.go` (`FreezeModule`)

## Database Namespace Isolation

All plugin database tables are prefixed with `plugin_<name>_`. A plugin named `bookmarks` that defines a table `links` gets `plugin_bookmarks_links` in the database.

### Enforcement

Table name validation happens at the Go level before any SQL is executed. The `db.*` API functions:

1. Prepend the plugin prefix to the table name provided in Lua.
2. Validate that the resulting table name starts with the correct prefix.
3. Reject any attempt to reference a table outside the plugin's namespace.

### What Plugins Cannot Do

- **Access CMS tables.** Queries against `content_data`, `users`, `media`, or any other CMS table are rejected.
- **Access other plugins' tables.** A plugin named `foo` cannot query `plugin_bar_tasks`.
- **Create cross-namespace foreign keys.** Foreign key `ref_table` values must match the plugin's own prefix.

### Auto-Injected Columns

Every plugin table receives three columns automatically: `id` (TEXT PRIMARY KEY, ULID), `created_at` (TEXT, RFC3339 UTC), `updated_at` (TEXT, RFC3339 UTC). Plugins cannot redefine these columns.

### Schema Drift Detection

After `db.define_table()` creates a table, the system inspects actual database columns and compares them against the Lua definition. Mismatches are logged as warnings and surfaced in the `GET /api/v1/admin/plugins/{name}` response. Drift is advisory only -- the system does not auto-migrate tables.

## Response Header Blocking

Plugin HTTP responses are filtered before being sent to the client. The following headers are silently dropped if set by a plugin:

| Blocked Header | Attack Vector |
|----------------|---------------|
| `access-control-*` | CORS policy override -- a plugin could allow arbitrary origins |
| `set-cookie` | Session fixation or hijacking via injected cookies |
| `transfer-encoding` | Response smuggling via chunked encoding manipulation |
| `content-length` | Response smuggling via length mismatch |
| `cache-control` | Cache poisoning via forced caching of dynamic responses |
| `host` | Request smuggling via host header manipulation |
| `connection` | Request smuggling via connection lifecycle manipulation |

Two security headers are set automatically on all plugin responses and cannot be overridden:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`

## Operation Budgets

Each VM checkout (HTTP handler or hook execution) has a finite operation budget. Every `db.*` call decrements the budget by one. This prevents a single handler from monopolizing database connections with unbounded queries.

| Context | Default Budget | Config Key |
|---------|---------------|------------|
| HTTP request handler | 1000 | `plugin_max_ops` |
| After-hook | 100 | `plugin_hook_max_ops` |
| Before-hook | 0 (all blocked) | -- |

Exceeding the budget raises a Lua error (`ErrOpLimitExceeded`). The handler can catch this with `pcall` but cannot perform further database operations.

### Why Before-Hooks Block db.*

Before-hooks run synchronously inside the CMS database transaction. Plugin database operations use a separate connection pool. On SQLite (single-writer), a plugin `db.*` call would attempt to acquire a write lock while the CMS transaction already holds one, causing a deadlock. On MySQL/PostgreSQL, it would create long-held locks and risk timeouts. The system blocks `db.*` entirely in before-hooks rather than allowing partial access that could fail unpredictably.

Use after-hooks for any database work. After-hooks run after the CMS transaction commits, so there is no lock contention.

## Circuit Breakers

Two independent circuit breaker systems protect the CMS from misbehaving plugins.

### Plugin-Level Circuit Breaker

Source: `internal/plugin/recovery.go`

Tracks consecutive HTTP handler failures and manager operation errors. Does not count hook failures.

| State | Behavior |
|-------|----------|
| Closed | Normal operation. Requests are processed. |
| Open | All requests return 503 immediately. No code executes. |
| Half-Open | One probe request is allowed. Success closes the breaker. Failure re-opens it. |

**Transitions:**

- Closed to Open: `plugin_max_failures` consecutive failures (default 5).
- Open to Half-Open: `plugin_reset_interval` elapses (default 60s).
- Half-Open to Closed: Probe request succeeds.
- Half-Open to Open: Probe request fails.

**Manual reset:** `modulacms plugin enable <name>` resets the circuit breaker and triggers a fresh load.

### Hook-Level Circuit Breaker

Source: `internal/plugin/hook_engine.go`

Each (plugin, event, table) combination has its own circuit breaker. This is more granular than the plugin-level breaker -- a failing `before_create` hook on `content_data` does not affect `after_update` hooks on the same table.

After `plugin_hook_max_consecutive_aborts` consecutive errors (default 10), that specific hook is disabled. It remains disabled until the plugin is reloaded or re-enabled.

Hook failures never feed into the plugin-level circuit breaker. A plugin can have all its hooks disabled while its HTTP endpoints continue to function normally.

## Rate Limiting

Source: `internal/plugin/http_bridge.go`

Per-IP token bucket rate limiter protects plugin endpoints from abuse.

### Design

- Each unique client IP gets its own token bucket.
- Bucket capacity equals the configured rate (default 100 tokens).
- Tokens refill at the configured rate per second.
- When a request arrives and no token is available, the bridge returns HTTP 429 (Too Many Requests).

### Cleanup

Every 5 minutes, the rate limiter scans all entries and removes those not seen in the last 10 minutes. This prevents unbounded memory growth from many unique IPs.

### Trusted Proxies

When `plugin_trusted_proxies` is configured with CIDR ranges, requests arriving from those IPs have their `X-Forwarded-For` header parsed to extract the real client IP. Without trusted proxy configuration, the direct connection IP is used, which behind a load balancer would be the proxy's IP -- causing all clients to share a single rate limit bucket.

## Route and Hook Approval

All routes and hooks start unapproved. This is a security gate: new plugin code does not execute until an admin explicitly approves it.

- **Unapproved routes** return 404 as if they do not exist.
- **Unapproved hooks** are silently skipped.

### Version-Change Revocation

When a plugin's `version` field in `plugin_info` changes, all route and hook approvals are automatically revoked. The admin must re-approve after reviewing the updated code. This prevents a plugin from changing its behavior (e.g., adding data exfiltration to an approved route) without going through the approval gate again.

See [approval workflow](approval.md) for the full approval process.

## Hot Reload Security

Source: `internal/plugin/watcher.go`

The hot reload watcher has several safety limits to prevent abuse:

| Limit | Value | Purpose |
|-------|-------|---------|
| Max .lua files per plugin | 100 | Prevents DoS via creating thousands of files that must be checksummed |
| Max total size per checksum | 10 MB | Prevents DoS via large files consuming CPU during SHA-256 hashing |
| Reload cooldown | 10s per plugin | Prevents reload storms from rapid file changes |
| Max consecutive slow reloads | 3 (>10s each) | Pauses watcher for plugins with systemic initialization issues |
| Debounce delay | 1s | Waits for file writes to settle before triggering reload |

### Blue-Green Safety

Reload creates a new plugin instance alongside the old one. The new instance goes through full initialization (VM pool, `init.lua`, routes, hooks, `on_init()`). Only after the new instance is fully running does the old one drain and shut down.

If the new instance fails at any point during initialization, the old instance remains active. No traffic is dropped during the transition.

### Eviction

After 3 consecutive slow reloads (each taking more than 10 seconds), the watcher pauses monitoring for that plugin. This prevents a plugin with a fundamentally broken initialization from consuming resources on every poll cycle. The plugin continues to run with its last successful version.

## Trusted Proxy Configuration

Accurate client IP identification is critical for rate limiting and `req.client_ip` in plugin handlers. Behind a reverse proxy or load balancer, the direct connection IP is the proxy's IP, not the client's.

Configure trusted proxies with CIDR notation:

```json
{
  "plugin_trusted_proxies": ["10.0.0.0/8", "172.16.0.0/12"]
}
```

When a request arrives from a trusted proxy IP, the rightmost non-trusted IP in the `X-Forwarded-For` chain is used as the client IP. If no trusted proxies are configured, the direct connection IP is used.

Do not add `0.0.0.0/0` or other overly broad ranges. This would allow any client to spoof their IP via the `X-Forwarded-For` header, bypassing rate limiting entirely.
