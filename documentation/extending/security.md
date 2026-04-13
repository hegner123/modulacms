# Plugin Security Model

How ModulaCMS prevents plugins from compromising CMS stability, accessing unauthorized data, or affecting other plugins.

## Lua Sandbox

Each plugin runs in a restricted Lua 5.1 environment. Before any plugin code executes, ModulaCMS removes dangerous functions and modules.

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

### Available Standard Library

Only safe standard library functions remain. See the [Lua API reference](/docs/extending/lua-api#allowed-lua-standard-library) for the complete list. In summary: `base` (type checks, iteration, error handling, metatable get/set), `string` (manipulation, pattern matching), `table` (insert, remove, sort, concat), `math` (arithmetic, random).

> **Good to know**: The `coroutine` library is additionally available for plugins that declare `screens` or `interfaces` in their manifest, enabling the TUI coroutine bridge for plugin-driven terminal UI.

### Frozen Module Protection

All injected API modules (`db`, `http`, `hooks`, `log`, `tui`) are frozen read-only via a metatable proxy:

- **Writes raise an error.** `db.query = nil` or `db.custom = function() end` fails with an error.
- **`getmetatable` returns `"protected"`.** Prevents retrieving and modifying the proxy metatable.
- **`rawget`/`rawset` are removed.** The standard Lua bypass for metatables is unavailable.

This prevents plugins from monkey-patching API modules to escalate privileges or intercept other plugins' calls.

## Database Namespace Isolation

All plugin database tables are prefixed with `plugin_<name>_`. A plugin named `bookmarks` that defines a table `links` gets `plugin_bookmarks_links` in the database.

### What Plugins Cannot Access

- **CMS tables.** Queries against content, users, media, or any other CMS table are rejected.
- **Other plugins' tables.** A plugin named `foo` cannot query `plugin_bar_tasks`.
- **Cross-namespace foreign keys.** Foreign key `ref_table` values must match the plugin's own prefix.

### Auto-Injected Columns

Every plugin table receives three columns automatically: `id` (ULID primary key), `created_at` (RFC3339 UTC), `updated_at` (RFC3339 UTC). Plugins cannot redefine these columns.

### Schema Drift Detection

After `db.define_table()` creates a table, ModulaCMS compares the actual database columns against the Lua definition. Mismatches appear as warnings in the `GET /api/v1/admin/plugins/{name}` response. Drift is advisory only -- ModulaCMS does not auto-migrate tables.

## Response Header Blocking

Plugin HTTP responses are filtered before reaching the client. The following headers are silently dropped if a plugin sets them:

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

Each plugin execution context has a finite operation budget. Every `db.*` call decrements the budget by one, preventing a single handler from monopolizing database connections.

| Context | Default Budget | Config Key |
|---------|---------------|------------|
| HTTP request handler | 1000 | `plugin_max_ops` |
| After-hook | 100 | `plugin_hook_max_ops` |
| Before-hook | 0 (all blocked) | -- |

Exceeding the budget raises a Lua error. The handler can catch this with `pcall` but cannot perform further database operations.

> **Good to know**: Before-hooks block all `db.*` calls because they run inside the CMS database transaction. A plugin database call from within a before-hook would deadlock on SQLite or risk long-held locks on other databases. Use after-hooks for database work.

## Circuit Breakers

Two independent circuit breaker systems protect the CMS from misbehaving plugins.

### Plugin-Level Circuit Breaker

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

Each (plugin, event, table) combination has its own circuit breaker. This is more granular than the plugin-level breaker -- a failing `before_create` hook on `content_data` does not affect `after_update` hooks on the same table.

After `plugin_hook_max_consecutive_aborts` consecutive errors (default 10), that specific hook is disabled until you reload or re-enable the plugin.

Hook failures never feed into the plugin-level circuit breaker. A plugin can have all its hooks disabled while its HTTP endpoints continue to function normally.

## Rate Limiting

A per-IP token bucket rate limiter protects plugin endpoints from abuse.

- Each unique client IP gets its own token bucket.
- Bucket capacity equals the configured rate (default 100 tokens, refilling at 100/s).
- When no token is available, the request receives HTTP 429 (Too Many Requests).
- Inactive entries are cleaned up automatically to prevent memory growth.

### Trusted Proxies

When `plugin_trusted_proxies` is configured with CIDR ranges, ModulaCMS parses the `X-Forwarded-For` header from those IPs to extract the real client IP. Without trusted proxy configuration, the direct connection IP is used -- behind a load balancer, this means all clients share a single rate limit bucket.

## Route and Hook Approval

All routes and hooks start unapproved. Plugin code does not execute until an admin explicitly approves it.

- **Unapproved routes** return 404 as if they don't exist.
- **Unapproved hooks** are silently skipped.

When a plugin's `version` field changes, ModulaCMS automatically revokes all approvals. The admin must re-approve after reviewing the updated code. See [approval workflow](/docs/extending/approval) for the full process.

## Hot Reload Security

The hot reload watcher has several safety limits to prevent abuse:

| Limit | Value | Purpose |
|-------|-------|---------|
| Max .lua files per plugin | 100 | Prevents DoS via creating thousands of files that must be checksummed |
| Max total size per checksum | 10 MB | Prevents DoS via large files consuming CPU during SHA-256 hashing |
| Reload cooldown | 10s per plugin | Prevents reload storms from rapid file changes |
| Max consecutive slow reloads | 3 (>10s each) | Pauses watcher for plugins with systemic initialization issues |
| Debounce delay | 1s | Waits for file writes to settle before triggering reload |

### Blue-Green Safety

ModulaCMS creates a new plugin instance alongside the old one during reload. Only after the new instance fully initializes does the old one shut down. If the new instance fails, the old one remains active and no traffic is dropped.

### Eviction

After 3 consecutive slow reloads (each taking more than 10 seconds), the watcher pauses monitoring for that plugin. The plugin continues to run with its last successful version.

## Trusted Proxy Configuration

Accurate client IP identification is critical for rate limiting and `req.client_ip` in plugin handlers. Behind a reverse proxy or load balancer, the direct connection IP is the proxy's IP, not the client's.

Configure trusted proxies with CIDR notation:

```json
{
  "plugin_trusted_proxies": ["10.0.0.0/8", "172.16.0.0/12"]
}
```

When a request arrives from a trusted proxy IP, ModulaCMS uses the rightmost non-trusted IP in the `X-Forwarded-For` chain as the client IP. If no trusted proxies are configured, the direct connection IP is used.

Do not add `0.0.0.0/0` or other overly broad ranges. This would allow any client to spoof their IP via the `X-Forwarded-For` header, bypassing rate limiting entirely.
