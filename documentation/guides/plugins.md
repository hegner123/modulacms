# Plugins

ModulaCMS includes a Lua-based plugin system for extending the CMS with custom HTTP endpoints, content lifecycle hooks, and isolated database storage. Plugins run in sandboxed Lua VMs (via gopher-lua), with controlled access to database operations, HTTP routing, and structured logging. The system enforces resource limits, approval gates, and circuit breakers to keep plugins from affecting core CMS stability.

## Overview

A plugin is a directory containing an `init.lua` entry point and optional modules in a `lib/` subdirectory. Plugins can:

- Register HTTP endpoints under `/api/v1/plugins/<plugin_name>/`
- React to content lifecycle events (create, update, delete, publish, archive) with hooks
- Create and manage their own database tables (namespaced and isolated from CMS tables)
- Log structured messages to the CMS log

Plugins cannot access the filesystem, execute system commands, or read CMS database tables directly. All plugin database operations are scoped to tables prefixed with `plugin_<name>_`.

## Configuration

Enable the plugin system and configure its runtime limits in `config.json`:

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

### Configuration Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_enabled` | bool | false | Master switch for the plugin system |
| `plugin_directory` | string | `"./plugins/"` | Path to scan for plugin directories |
| `plugin_max_vms` | int | 4 | VM pool size per plugin |
| `plugin_timeout` | int | 5 | Execution timeout per handler/hook (seconds) |
| `plugin_max_ops` | int | 1000 | Max database operations per request handler |
| `plugin_hot_reload` | bool | false | Enable file-watching for live reload |
| `plugin_max_failures` | int | 5 | Consecutive failures before circuit breaker trips |
| `plugin_reset_interval` | string | `"60s"` | Duration before circuit breaker half-opens |
| `plugin_rate_limit` | int | 100 | Max requests per second per IP |
| `plugin_max_routes` | int | 50 | Max routes per plugin |
| `plugin_max_request_body` | int | 1048576 | Max request body size in bytes (1 MB) |
| `plugin_max_response_body` | int | 5242880 | Max response body size in bytes (5 MB) |
| `plugin_trusted_proxies` | []string | [] | IP ranges for X-Forwarded-For parsing |
| `plugin_hook_reserve_vms` | int | 1 | VMs reserved for hook execution per plugin |
| `plugin_hook_max_consecutive_aborts` | int | 10 | Consecutive hook errors before hook-level circuit breaker trips |
| `plugin_hook_max_ops` | int | 100 | Max database operations per after-hook |
| `plugin_hook_max_concurrent_after` | int | 10 | Max concurrent after-hook goroutines |
| `plugin_hook_timeout_ms` | int | 2000 | Per-hook timeout for before-hooks (ms) |
| `plugin_hook_event_timeout_ms` | int | 5000 | Total timeout for before-hook chain per event (ms) |

## Plugin Structure

```
plugins/
  my_plugin/
    init.lua          # Entry point (required)
    lib/
      helpers.lua     # Optional modules loaded via require()
      validators.lua
```

The directory name determines the plugin name, which becomes the database table prefix and route prefix.

### Manifest

Every plugin must define a global `plugin_info` table at the top of `init.lua`:

```lua
plugin_info = {
    name        = "task_tracker",
    version     = "1.0.0",
    description = "Track tasks with a REST API",
    author      = "Dev Team",
    license     = "MIT",
}
```

**Name rules**: lowercase alphanumeric plus underscores, no trailing underscore, max 32 characters. The name must match the directory name.

**Version changes**: When the version string changes, all route and hook approvals are revoked and require re-approval. This prevents stale approvals from applying to updated code.

### Execution Order

The code in `init.lua` runs in a specific order:

```lua
local helpers = require("helpers")     -- 1. Load modules

plugin_info = { ... }                  -- 2. Declare manifest

http.handle("GET", "/tasks", ...)      -- 3. Register routes
http.handle("POST", "/tasks", ...)

hooks.on("before_create", ...)         -- 4. Register hooks
```

Routes, hooks, and `require()` calls must be at module scope (top-level code). Do not register routes or hooks inside `on_init()` -- this raises an error.

### Lifecycle Functions

```lua
-- Runs once after all VMs are created. Use for table creation and seed data.
function on_init()
    db.define_table("tasks", { ... })
    if db.count("tasks", {}) == 0 then
        db.insert("tasks", { title = "Default task" })
    end
    log.info("task_tracker initialized")
end

-- Runs during graceful shutdown. Errors are logged but do not block shutdown.
function on_shutdown()
    log.info("task_tracker shutting down")
end
```

If `on_init()` calls `error()`, the plugin is marked as failed and will not serve traffic.

## Database API

Plugins get sandboxed database access scoped to their own tables. All table names are auto-prefixed with `plugin_<name>_`. In Lua code, use short names only (e.g., `"tasks"` becomes `plugin_task_tracker_tasks` in the database).

### Defining Tables

Call `db.define_table()` inside `on_init()`. Three columns are auto-injected on every table -- do not include them in your definition:

- `id` (TEXT PRIMARY KEY, ULID)
- `created_at` (TEXT, RFC3339 UTC)
- `updated_at` (TEXT, RFC3339 UTC)

```lua
db.define_table("tasks", {
    columns = {
        { name = "title",    type = "text",    not_null = true },
        { name = "status",   type = "text",    not_null = true, default = "todo" },
        { name = "priority", type = "integer", default = 0 },
        { name = "done",     type = "boolean" },
        { name = "metadata", type = "json" },
    },
    indexes = {
        { columns = {"status"} },
        { columns = {"status", "priority"}, unique = false },
    },
    foreign_keys = {
        {
            column     = "category_id",
            ref_table  = "plugin_task_tracker_categories",
            ref_column = "id",
            on_delete  = "CASCADE",
        },
    },
})
```

**Column types**: `text`, `integer`, `boolean`, `real`, `timestamp`, `json`, `blob`

Foreign keys must reference tables owned by the same plugin (namespace isolation is enforced).

### Queries

```lua
-- SELECT with filters, ordering, pagination
local tasks = db.query("tasks", {
    where    = { status = "pending" },
    order_by = "created_at DESC",
    limit    = 10,
    offset   = 0,
})

-- Single row (returns nil if not found)
local task = db.query_one("tasks", { where = { id = task_id } })

-- Count
local total = db.count("tasks", { where = { status = "done" } })

-- Exists check
local found = db.exists("tasks", { where = { id = task_id } })
```

`db.query()` returns an empty table `{}` on no matches (never nil). Omitting the `where` clause returns all rows up to the limit (default 100).

### Mutations

```lua
-- Insert (id, created_at, updated_at auto-set if omitted)
db.insert("tasks", { title = "Fix bug", status = "todo" })

-- Update (both set and where required -- prevents full-table updates)
db.update("tasks", {
    set   = { status = "done" },
    where = { id = task_id },
})

-- Delete (where required -- prevents full-table deletes)
db.delete("tasks", { where = { id = task_id } })
```

### Transactions

```lua
local ok, err = db.transaction(function()
    db.insert("tasks", { title = "Task 1" })
    db.insert("tasks", { title = "Task 2" })
    -- error() inside rolls back the entire transaction
end)

if not ok then
    log.warn("Transaction failed", { err = err })
end
```

Nested transactions are rejected.

### Utilities

```lua
local id  = db.ulid()       -- Generate a 26-character ULID
local now = db.timestamp()  -- Current UTC time as RFC3339 string
```

## HTTP Routes

Register HTTP endpoints under `/api/v1/plugins/<plugin_name>/`.

### Registering Routes

```lua
-- Authenticated (default): requires CMS session
http.handle("GET", "/tasks", function(req)
    local tasks = db.query("tasks", {})
    return { status = 200, json = tasks }
end)

-- Public: no CMS session required
http.handle("POST", "/webhook", function(req)
    return { status = 200, json = { ok = true } }
end, { public = true })
```

**Valid methods**: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`

**Path rules**: Must start with `/`, max 256 characters. Supports `{param}` path parameters. No `..`, `?`, or `#`.

The full URL for a route is `/api/v1/plugins/<plugin_name><path>`. For example, `http.handle("GET", "/tasks", ...)` on the `task_tracker` plugin becomes `GET /api/v1/plugins/task_tracker/tasks`.

Routes require admin approval before serving traffic. Unapproved routes return 404.

### Request Object

| Field | Type | Description |
|-------|------|-------------|
| `req.method` | string | HTTP method |
| `req.path` | string | URL path |
| `req.body` | string | Raw request body |
| `req.client_ip` | string | Client IP (proxy-aware) |
| `req.headers` | table | All headers (lowercase keys) |
| `req.query` | table | URL query parameters |
| `req.params` | table | Path parameters from `{param}` wildcards |
| `req.json` | table | Parsed JSON body (when Content-Type is application/json) |

### Response Object

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | number | 200 | HTTP status code |
| `json` | table | nil | Serialized as JSON (sets Content-Type automatically) |
| `body` | string | nil | Raw string body (used only if `json` is nil) |
| `headers` | table | nil | Custom response headers |

### Middleware

Register middleware that runs before all route handlers. Return a response table to short-circuit, or nil to continue:

```lua
http.use(function(req)
    if not req.headers["x-api-key"] then
        return { status = 401, json = { error = "missing api key" } }
    end
    return nil
end)
```

## Content Hooks

React to content lifecycle events on CMS tables.

### Registering Hooks

```lua
-- Before-hooks run synchronously inside the CMS transaction.
-- Call error() to abort the transaction.
hooks.on("before_create", "content_data", function(data)
    if data.title and data.title == "" then
        error("title cannot be empty")
    end
end, { priority = 50 })

-- After-hooks run asynchronously (fire-and-forget).
hooks.on("after_update", "content_data", function(data)
    db.insert("activity_log", {
        action     = "content_updated",
        content_id = data.id,
    })
end)

-- Wildcard: runs for all tables.
hooks.on("after_delete", "*", function(data)
    log.warn("Record deleted", { table_name = data._table, id = data.id })
end)
```

### Hook Events

| Event | Timing | Can Abort? | db.* Access? |
|-------|--------|------------|--------------|
| `before_create` | Inside transaction | Yes (`error()`) | No |
| `after_create` | After commit | No | Yes |
| `before_update` | Inside transaction | Yes | No |
| `after_update` | After commit | No | Yes |
| `before_delete` | Inside transaction | Yes | No |
| `after_delete` | After commit | No | Yes |
| `before_publish` | Inside transaction | Yes | No |
| `after_publish` | After commit | No | Yes |
| `before_archive` | Inside transaction | Yes | No |
| `after_archive` | After commit | No | Yes |

`before_publish` and `before_archive` trigger automatically when content status transitions to `"published"` or `"archived"` (on the `content_data` table only).

### Key Constraints

- **db.* calls are blocked inside before-hooks.** Before-hooks run inside the CMS database transaction. Plugin database calls use a separate connection pool that would deadlock with the active transaction (especially on SQLite). Use after-hooks for any database work.
- After-hooks run with a reduced operation budget (100 ops vs 1000 for HTTP handlers).
- Hooks require admin approval before they fire. Unapproved hooks are silently skipped.

### Hook Options

- `priority`: 1-1000, lower runs first (default 100). Table-specific hooks run before wildcard hooks at equal priority.
- Max 50 hooks per plugin.
- Before-hook timeout: 2s per hook, 5s per event chain (configurable).

### Hook Data

The handler receives a table with all entity fields plus:

| Field | Description |
|-------|-------------|
| `data._table` | The CMS table name |
| `data._event` | The event name |

### Hook Circuit Breaker

Each (plugin, event, table) combination has its own circuit breaker. After 10 consecutive errors (configurable via `plugin_hook_max_consecutive_aborts`), that specific hook is disabled until the plugin is reloaded or re-enabled. Hook failures do not feed into the plugin-level circuit breaker.

## Logging

Structured logging with automatic `plugin: <name>` context:

```lua
log.info("Task created", { id = task_id, user = user_name })
log.warn("Deprecated API used")
log.error("Failed to process", { reason = err })
log.debug("Trace detail", { step = 5, value = result })
```

Context table key-value pairs are flattened into structured log arguments.

## Module System

Split plugin code into modules under `lib/`:

```lua
-- lib/validators.lua
local M = {}
function M.not_empty(s)
    return type(s) == "string" and #s > 0
end
return M

-- init.lua
local v = require("validators")
if v.not_empty(title) then ... end
```

Only files under the plugin's own `lib/` directory are loadable. Path traversal (`..`, `/`, `\`) is rejected. Modules are cached after first load.

## Route and Hook Approval

All routes and hooks start unapproved. This is a security gate: new code does not execute until an admin explicitly approves it.

### Approving via CLI

```bash
# Approve all routes and hooks
modulacms plugin approve my_plugin --all-routes
modulacms plugin approve my_plugin --all-hooks

# Approve specific items
modulacms plugin approve my_plugin --route "GET /tasks"
modulacms plugin approve my_plugin --hook "before_create:content_data"

# Revoke
modulacms plugin revoke my_plugin --all-routes
modulacms plugin revoke my_plugin --route "GET /tasks"
```

Add `--yes` to skip confirmation prompts (for CI/CD pipelines).

### Approving via API

```bash
# List routes with approval status
curl http://localhost:8080/api/v1/admin/plugins/routes \
  -H "Cookie: session=YOUR_SESSION_COOKIE"

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
      {"plugin": "task_tracker", "event": "before_create", "table": "content_data"}
    ]
  }'
```

Approvals are idempotent: approving an already-approved item is a no-op.

### Approving via TUI

The SSH TUI includes a Plugins page accessible from the homepage menu. Select a plugin to view details and approve routes or hooks through a confirmation dialog.

## Hot Reload

When `plugin_hot_reload` is `true`, the system watches for `.lua` file changes every 2 seconds. Changes trigger a blue-green reload: a new plugin instance is created alongside the old one. If the new instance loads successfully, it replaces the old one atomically. If it fails, the old instance keeps running.

Reload limits:
- 1-second debounce window (waits for file writes to settle)
- 10-second cooldown between reloads per plugin
- Max 100 `.lua` files and 10 MB total per plugin for checksumming
- After 3 consecutive slow reloads (>10s each), the watcher pauses for that plugin

Manual reload via CLI or API:

```bash
modulacms plugin reload task_tracker
```

```bash
curl -X POST http://localhost:8080/api/v1/admin/plugins/task_tracker/reload \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## Circuit Breaker

**Plugin-level**: After 5 consecutive failures (configurable), the circuit breaker opens. All requests return 503 until the reset interval (60s) elapses and a probe request succeeds. HTTP handler errors and manager operations (reload, init) feed into this breaker. Hook failures do not.

| State | Behavior |
|-------|----------|
| Closed | Normal operation |
| Open | All requests rejected (503) |
| Half-Open | One probe request allowed; success closes, failure re-opens |

Reset the circuit breaker manually:

```bash
modulacms plugin enable task_tracker
```

## CLI Commands

### Offline (no running server required)

```bash
modulacms plugin list                 # List discovered plugins
modulacms plugin init my_plugin       # Scaffold a new plugin directory
modulacms plugin validate ./plugins/my_plugin  # Validate manifest without loading
```

The `init` command accepts optional flags:

```bash
modulacms plugin init my_plugin \
  --version 1.0.0 \
  --description "My plugin" \
  --author "Dev Team" \
  --license MIT
```

### Online (requires running server)

Online commands authenticate via a Bearer token auto-generated at server startup (written to `<config_dir>/.plugin-api-token`). Pass `--token <value>` for CI/CD.

```bash
modulacms plugin info my_plugin       # Plugin details and status
modulacms plugin reload my_plugin     # Blue-green hot reload
modulacms plugin enable my_plugin     # Reset circuit breaker, reload
modulacms plugin disable my_plugin    # Stop plugin, trip circuit breaker
```

## Sandboxing and Security

### Lua Environment

Plugins run in a restricted Lua environment. Available standard library:

| Module | Available Functions |
|--------|-------------------|
| `base` | type, tostring, tonumber, pairs, ipairs, next, select, unpack, error, pcall, xpcall, setmetatable, getmetatable |
| `string` | find, sub, len, format, match, gmatch, gsub, rep, reverse, byte, char, lower, upper |
| `table` | insert, remove, sort, concat |
| `math` | floor, ceil, max, min, abs, sqrt, huge, pi, random, randomseed |

Removed: `io`, `os`, `package`, `debug`, `dofile`, `loadfile`, `load`, `rawget`, `rawset`, `rawequal`, `rawlen`. No filesystem access, no process execution, no dynamic code loading, no metatable bypass.

All injected modules (`db`, `http`, `hooks`, `log`) are frozen read-only via metatable proxy.

### Database Isolation

All plugin tables are prefixed with `plugin_<name>_`. Plugins cannot access core CMS tables, other plugins' tables, or create foreign keys to tables outside their namespace.

### Response Header Restrictions

Plugins cannot set these response headers:

| Header | Reason |
|--------|--------|
| `access-control-*` | Prevents CORS policy override |
| `set-cookie` | Prevents session manipulation |
| `transfer-encoding` | Prevents response smuggling |
| `content-length` | Prevents response smuggling |
| `cache-control` | Prevents cache poisoning |
| `host`, `connection` | Prevents request smuggling |

Security headers `X-Content-Type-Options: nosniff` and `X-Frame-Options: DENY` are set automatically on all plugin responses.

### Operation Budgets

| Context | Max Operations | Configurable Via |
|---------|---------------|-----------------|
| HTTP request handler | 1000 | `plugin_max_ops` |
| After-hook | 100 | `plugin_hook_max_ops` |
| Before-hook | 0 (blocked) | -- |

Exceeding the budget raises an error in Lua. Before-hooks block all `db.*` calls entirely to prevent transaction deadlocks.

### Rate Limiting

Per-IP token bucket: 100 requests/second by default. Exceeding the limit returns HTTP 429.

## Error Handling

| Error Source | Behavior |
|-------------|----------|
| `error()` in `on_init()` | Plugin marked as failed, does not serve traffic |
| `error()` in `on_shutdown()` | Logged, shutdown continues |
| `error()` in before-hook | CMS transaction aborted, HTTP 422 returned |
| `error()` in after-hook | Logged, does not affect response |
| `error()` in HTTP handler | HTTP 500 returned |
| Operation limit exceeded | Error raised in Lua |

## Admin API Reference

All plugin management endpoints require authentication. Admin-only endpoints require the `plugins:admin` permission; read endpoints require `plugins:read`.

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/admin/plugins` | `plugins:read` | List all plugins with state |
| GET | `/api/v1/admin/plugins/{name}` | `plugins:read` | Plugin details, VM counts, circuit breaker state |
| POST | `/api/v1/admin/plugins/{name}/reload` | `plugins:admin` | Trigger hot reload |
| POST | `/api/v1/admin/plugins/{name}/enable` | `plugins:admin` | Reset circuit breaker, reload |
| POST | `/api/v1/admin/plugins/{name}/disable` | `plugins:admin` | Stop plugin, trip circuit breaker |
| GET | `/api/v1/admin/plugins/cleanup` | `plugins:admin` | List orphaned plugin tables (dry-run) |
| POST | `/api/v1/admin/plugins/cleanup` | `plugins:admin` | Drop orphaned tables (requires confirmation) |
| GET | `/api/v1/admin/plugins/routes` | `plugins:read` | List all routes with approval status |
| POST | `/api/v1/admin/plugins/routes/approve` | `plugins:admin` | Approve routes |
| POST | `/api/v1/admin/plugins/routes/revoke` | `plugins:admin` | Revoke route approval |
| GET | `/api/v1/admin/plugins/hooks` | `plugins:read` | List all hooks with approval status |
| POST | `/api/v1/admin/plugins/hooks/approve` | `plugins:admin` | Approve hooks |
| POST | `/api/v1/admin/plugins/hooks/revoke` | `plugins:admin` | Revoke hook approval |

## Developer Workflow

1. **Scaffold**: `modulacms plugin init my_plugin`
2. **Develop**: Write `init.lua`, add modules in `lib/`
3. **Validate**: `modulacms plugin validate ./plugins/my_plugin`
4. **Deploy**: Copy the plugin directory to the server's `plugin_directory` path
5. **Start**: The server picks up the plugin on startup, or the watcher auto-reloads if `plugin_hot_reload` is enabled
6. **Approve**: `modulacms plugin approve my_plugin --all-routes --all-hooks`
7. **Monitor**: `modulacms plugin info my_plugin` to check state and circuit breaker status
8. **Update**: Edit Lua files; the watcher auto-reloads via blue-green deployment with no downtime

## Notes

- **Schema drift detection.** After `db.define_table()` creates a table, the system inspects actual columns and compares them against the definition. Mismatches are logged as warnings and surfaced in the `GET /api/v1/admin/plugins/{name}` response. Drift is advisory only -- the system does not auto-migrate tables.
- **Plugin lifecycle states**: Discovered (0), Loading (1), Running (2), Failed (3), Stopped (4).
- **Dependencies.** Plugins can declare dependencies in `plugin_info.dependencies`. Dependencies are loaded first. Circular dependencies are rejected.
- **Shutdown order.** During graceful shutdown: file watcher stops, HTTP bridge drains inflight requests, hook engine drains inflight after-hooks, then the manager closes pools in reverse dependency order and calls `on_shutdown()`.
- **VM pool.** Each plugin gets a pool of Lua VMs. The pool has two channels: a general channel for HTTP requests and a reserved channel for hooks, so hooks always have VM availability even when HTTP traffic saturates the general pool.
