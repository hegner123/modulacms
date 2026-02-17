# Plugin Development Guide

How to create, structure, and manage Lua plugins for ModulaCMS.

For the complete Lua API reference, see [PLUGIN_API.md](PLUGIN_API.md). For configuration and security details, see [CONFIGURATION.md](CONFIGURATION.md).

## Plugin Directory Structure

```
plugins/
  my_plugin/
    init.lua          # Entry point (required)
    lib/
      helpers.lua     # Optional modules loaded via require()
      validators.lua
```

The plugin name is derived from the directory name and must match `plugin_info.name`.

## Manifest

Every plugin must define a global `plugin_info` table at the top of `init.lua`.

```lua
plugin_info = {
    name        = "my_plugin",              -- required: [a-z0-9_], max 32 chars
    version     = "1.0.0",                  -- required: semantic version
    description = "What this plugin does",  -- required: shown in admin listings
    author      = "Your Name",              -- optional
    license     = "MIT",                    -- optional
    min_cms_version = "1.0.0",              -- optional (reserved for future)
    dependencies = {"other_plugin"},        -- optional: loaded first, cycles rejected
}
```

**Name rules**: lowercase alphanumeric + underscores, no trailing underscore. The name becomes the database table prefix (`plugin_<name>_`) and route prefix (`/api/v1/plugins/<name>/`).

**Version changes**: When the version string changes, all route and hook approvals are revoked and require re-approval.

## Execution Order

### Module Scope (runs for every VM in the pool)

Routes, hooks, and `require()` calls must go at module scope (top-level code in `init.lua`). **Do NOT register routes or hooks inside `on_init()`** -- each VM needs its own registered function reference. Calling `http.handle()` or `hooks.on()` inside `on_init()` raises an error.

```lua
local validators = require("validators")  -- 1. Load modules

plugin_info = { ... }                      -- 2. Declare manifest

http.handle("GET", "/tasks", function(req) -- 3. Register routes
    ...
end)

hooks.on("before_create", "content_data",  -- 4. Register hooks
    function(data) ... end)
```

### on_init() (runs once after all VMs created)

One-time setup: create tables, seed data, validate state.

```lua
function on_init()
    db.define_table("tasks", { ... })

    local count = db.count("tasks", {})
    if count == 0 then
        db.insert("tasks", { title = "Default task" })
    end

    log.info("my_plugin initialized")
end
```

If `on_init()` calls `error()`, the plugin is marked failed and will not serve traffic.

### on_shutdown() (runs during graceful shutdown)

Cleanup. Errors are logged but shutdown continues regardless.

```lua
function on_shutdown()
    log.info("my_plugin shutting down")
end
```

## Database API

Sandboxed database access scoped to plugin-prefixed tables (`plugin_<name>_<table>`). All tables get auto-injected `id` (ULID), `created_at`, and `updated_at` columns.

See [PLUGIN_API.md](PLUGIN_API.md#db----database-api) for the complete API reference.

### Table Definition

Call `db.define_table()` inside `on_init()`. Three columns are auto-injected -- do not include them in your columns list (`id`, `created_at`, `updated_at`).

```lua
db.define_table("tasks", {
    columns = {
        { name = "title",       type = "text",    not_null = true },
        { name = "status",      type = "text",    not_null = true, default = "todo" },
        { name = "category_id", type = "text" },
        { name = "priority",    type = "integer", default = 0 },
        { name = "done",        type = "boolean" },
        { name = "weight",      type = "real" },
        { name = "metadata",    type = "json" },
    },
    indexes = {
        { columns = {"status"} },
        { columns = {"status", "priority"}, unique = false },
    },
    foreign_keys = {
        {
            column     = "category_id",
            ref_table  = "plugin_my_plugin_categories", -- must use same plugin prefix
            ref_column = "id",
            on_delete  = "CASCADE",
        },
    },
})
```

**Column types**: `text`, `integer`, `boolean`, `real`, `timestamp`, `json`, `blob`

Foreign key `ref_table` must reference another table owned by the same plugin (namespace isolation enforced).

### Queries and Mutations

```lua
-- SELECT with filtering, ordering, pagination
local tasks = db.query("tasks", {
    where    = { status = "pending" },
    order_by = "created_at DESC",
    limit    = 10,
    offset   = 0,
})

-- Single row (LIMIT 1), returns nil if not found
local task = db.query_one("tasks", { where = { id = task_id } })

-- Aggregates
local count  = db.count("tasks", { where = { status = "done" } })
local exists = db.exists("tasks", { where = { id = task_id } })

-- Insert (id, created_at, updated_at auto-set if omitted)
db.insert("tasks", { title = "New task", status = "pending" })

-- Update (both set and where are required)
db.update("tasks", {
    set   = { status = "done" },
    where = { id = task_id },
})

-- Delete (where is required)
db.delete("tasks", { where = { id = task_id } })
```

### Transactions

```lua
local ok, err = db.transaction(function()
    db.insert("tasks", { title = "Task 1" })
    db.insert("tasks", { title = "Task 2" })
    -- error() inside rolls back
end)

if not ok then
    log.warn("Transaction failed", { err = err })
end
```

### Utilities

```lua
local id  = db.ulid()        -- Generate a new ULID
local now = db.timestamp()   -- RFC3339 UTC timestamp
```

## HTTP Routes

Register HTTP endpoints under `/api/v1/plugins/<plugin_name>/`.

### Route Registration

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

Routes require admin approval before serving traffic. Unapproved routes return 404.

**Valid methods**: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`

**Path rules**: must start with `/`, no `..`, `?`, `#`, max 256 chars. Supports `{param}` path parameters.

### Request Object

| Field            | Type   | Description                                    |
|------------------|--------|------------------------------------------------|
| `req.method`     | string | HTTP method                                    |
| `req.path`       | string | URL path                                       |
| `req.body`       | string | Raw request body                               |
| `req.client_ip`  | string | Client IP (proxy-aware)                        |
| `req.headers`    | table  | All headers (lowercase keys)                   |
| `req.query`      | table  | URL query parameters                           |
| `req.params`     | table  | Path parameters from `{param}` wildcards       |
| `req.json`       | table  | Parsed JSON body (when Content-Type is JSON)   |

### Response Object

| Field              | Type   | Description                                  |
|--------------------|--------|----------------------------------------------|
| `response.status`  | number | HTTP status code (default 200)               |
| `response.headers` | table  | Custom response headers                      |
| `response.body`    | string | Response body string                         |
| `response.json`    | table  | JSON response (takes priority over `body`)   |

**Blocked response headers**: `access-control-*`, `set-cookie`, `transfer-encoding`, `content-length`, `cache-control`, `host`, `connection`. Auto-set security headers: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`.

### Middleware

Middleware runs before route handlers. Return a response table to short-circuit, or `nil` to continue.

```lua
http.use(function(req)
    if req.headers["x-api-key"] ~= "secret" then
        return { status = 401, json = { error = "Unauthorized" } }
    end
    -- return nil to continue to the route handler
end)
```

### Limits

| Limit | Default | Config Key |
|-------|---------|------------|
| Max routes per plugin | 50 | `plugin_max_routes` |
| Request body | 1 MB | `plugin_max_request_body` |
| Response body | 5 MB | `plugin_max_response_body` |
| Rate limit | 100 req/s per IP | `plugin_rate_limit` |
| Execution timeout | 5s | `plugin_timeout` |

## Content Hooks

React to content lifecycle events on audited CMS tables.

### Hook Registration

```lua
-- Before-hooks run synchronously inside the DB transaction.
-- Return nothing to allow, call error() to abort.
hooks.on("before_create", "content_data", function(data)
    if not data.slug or data.slug == "" then
        error("slug is required")  -- aborts the CMS transaction
    end
end, { priority = 50 })

-- After-hooks run asynchronously (fire-and-forget).
hooks.on("after_update", "content_data", function(data)
    log.info("Content updated", { id = data.id })
end)

-- Wildcard: runs for all tables.
hooks.on("after_delete", "*", function(data)
    log.warn("Record deleted", { table = data._table, id = data.id })
end)
```

### Events

| Event             | Timing        | Can Abort? | Has db.* Access? |
|-------------------|---------------|------------|------------------|
| `before_create`   | Synchronous   | Yes        | No               |
| `after_create`    | Asynchronous  | No         | Yes              |
| `before_update`   | Synchronous   | Yes        | No               |
| `after_update`    | Asynchronous  | No         | Yes              |
| `before_delete`   | Synchronous   | Yes        | No               |
| `after_delete`    | Asynchronous  | No         | Yes              |
| `before_publish`  | Synchronous   | Yes        | No               |
| `after_publish`   | Asynchronous  | No         | Yes              |
| `before_archive`  | Synchronous   | Yes        | No               |
| `after_archive`   | Asynchronous  | No         | Yes              |

`before_publish` and `before_archive` are triggered automatically when content status transitions to `"published"` or `"archived"` (on `content_data` table only).

### Critical Constraints

- **db.* calls are blocked inside before-hooks.** Before-hooks run inside the CMS transaction. Plugin `db.*` calls use a separate connection pool that would deadlock with the active CMS transaction (especially on SQLite). Use after-hooks for any database work.
- After-hooks run post-commit with a reduced operation budget (100 ops vs 1000 for HTTP handlers).
- Hooks require admin approval before they fire. Unapproved hooks are silently skipped.

### Hook Data

The handler receives a table with all entity fields plus:

| Field         | Description        |
|---------------|--------------------|
| `data._table` | The CMS table name |
| `data._event` | The event name     |

### Options

- `priority`: 1-1000, lower runs first (default 100). Specific table hooks run before wildcard hooks at equal priority.
- Max 50 hooks per plugin.
- Before-hook timeout: 2s per hook, 5s per event chain (configurable).
- After-hooks: bounded concurrency (10 goroutines), reduced op budget (100 ops).

### Hook Circuit Breaker

Each (plugin, event, table) combination has a circuit breaker. After 10 consecutive aborts (configurable), that specific hook is disabled until the plugin is reloaded or re-enabled via admin. Hook failures do NOT feed into the plugin-level circuit breaker.

## Logging

Structured logging with automatic `plugin: <name>` context.

```lua
log.info("Task created", { id = task_id, user = user_name })
log.warn("Deprecated API used")
log.error("Failed to process", { reason = err })
log.debug("Trace detail", { step = 5, value = result })
```

## Module System

Plugins can split code into modules under `lib/`:

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

Only files under the plugin's own `lib/` directory are loadable (sandbox enforced). Path traversal (`..`, `/`, `\`) is rejected. Modules are cached after first load.

## Common Patterns

### CRUD REST API

```lua
-- List with filters
http.handle("GET", "/items", function(req)
    local opts = { order_by = "created_at DESC" }
    if req.query and req.query.status then
        opts.where = { status = req.query.status }
    end
    local items = db.query("items", opts)
    return { status = 200, json = { items = items, count = #items } }
end)

-- Create
http.handle("POST", "/items", function(req)
    if not req.json or not req.json.title then
        return { status = 400, json = { error = "title required" } }
    end
    local id = db.ulid()
    db.insert("items", { id = id, title = req.json.title })
    local item = db.query_one("items", { where = { id = id } })
    return { status = 201, json = item }
end)

-- Read by ID
http.handle("GET", "/items/{id}", function(req)
    local item = db.query_one("items", { where = { id = req.params.id } })
    if not item then
        return { status = 404, json = { error = "not found" } }
    end
    return { status = 200, json = item }
end)

-- Update
http.handle("PUT", "/items/{id}", function(req)
    if not db.exists("items", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end
    db.update("items", { set = req.json, where = { id = req.params.id } })
    local item = db.query_one("items", { where = { id = req.params.id } })
    return { status = 200, json = item }
end)

-- Delete
http.handle("DELETE", "/items/{id}", function(req)
    if not db.exists("items", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end
    db.delete("items", { where = { id = req.params.id } })
    return { status = 200, json = { deleted = true } }
end)
```

### Transaction-Wrapped Seed Data

```lua
function on_init()
    db.define_table("categories", {
        columns = {
            { name = "name", type = "text", not_null = true, unique = true },
        },
    })

    if db.count("categories", {}) == 0 then
        local ok, err = db.transaction(function()
            db.insert("categories", { name = "General" })
            db.insert("categories", { name = "Bug" })
            db.insert("categories", { name = "Feature" })
        end)
        if not ok then
            log.warn("Seed failed (likely already exist)", { err = err })
        end
    end
end
```

### Content Validation Hook

```lua
hooks.on("before_create", "content_data", function(data)
    if data.title and data.title:match("^%s*$") then
        error("content title cannot be empty")  -- aborts the CMS transaction
    end
end, { priority = 50 })
```

Note: before-hooks can read `data` fields for validation, but cannot call `db.*` functions. Use after-hooks for any database work.

### Publish Guard

```lua
hooks.on("before_publish", "content_data", function(data)
    if not data.slug or data.slug == "" then
        error("cannot publish content without a slug")
    end
end)
```

### After-Hook Notifications

```lua
hooks.on("after_create", "content_data", function(data)
    -- After-hooks have db.* access (reduced 100 op budget)
    db.insert("activity_log", {
        action = "content_created",
        content_id = data.id,
        title = data.title,
    })
end)
```

## Plugin Lifecycle States

| State | Value | Description |
|-------|-------|-------------|
| Discovered | 0 | Directory found, manifest extracted, not yet loaded |
| Loading | 1 | Dependencies validated, VM pool creation in progress |
| Running | 2 | on_init succeeded, fully operational |
| Failed | 3 | Load or runtime error (reason stored in FailedReason) |
| Stopped | 4 | on_shutdown has run or admin disabled |

## Route & Hook Approval

All routes and hooks start unapproved. Unapproved routes return 404; unapproved hooks are silently skipped.

### CLI

```bash
# Approve
modulacms plugin approve my_plugin --all-routes
modulacms plugin approve my_plugin --all-hooks
modulacms plugin approve my_plugin --route "GET /tasks"
modulacms plugin approve my_plugin --hook "before_create:content_data"

# Revoke
modulacms plugin revoke my_plugin --all-routes
modulacms plugin revoke my_plugin --route "GET /tasks"
modulacms plugin revoke my_plugin --hook "before_create:content_data"
```

Add `--yes` to skip confirmation prompts (for CI/CD).

### Admin API

```bash
# Routes
GET  /api/v1/admin/plugins/routes            # List all routes with approval status
POST /api/v1/admin/plugins/routes/approve     # Approve routes
POST /api/v1/admin/plugins/routes/revoke      # Revoke approval

# Hooks
GET  /api/v1/admin/plugins/hooks             # List all hooks with approval status
POST /api/v1/admin/plugins/hooks/approve     # Approve hooks
POST /api/v1/admin/plugins/hooks/revoke      # Revoke approval
```

**Route approve/revoke request body:**

```json
{
    "routes": [
        {"plugin": "my_plugin", "method": "GET", "path": "/tasks"}
    ]
}
```

**Hook approve/revoke request body:**

```json
{
    "hooks": [
        {"plugin": "my_plugin", "event": "before_create", "table": "content_data"}
    ]
}
```

Idempotent: approving an already-approved item or revoking an already-revoked item is a no-op.

### TUI

The SSH TUI includes a Plugins page accessible from the homepage menu. Select a plugin to view details and perform actions: Enable, Disable, Reload, Approve Routes, Approve Hooks. Approve actions show a confirmation dialog before executing.

## Hot Reload

When `plugin_hot_reload: true` in config, the file watcher polls every 2s for `.lua` file changes.

**Blue-green reload**: A new plugin instance is created alongside the old one. If the new instance loads successfully, it replaces the old one atomically. If it fails, the old instance keeps running.

**Limits**:
- Debounce: 1s stability window (waits for file writes to settle)
- Cooldown: 10s between reloads per plugin
- Max 100 `.lua` files, 10 MB total per checksumming pass
- After 3 consecutive slow reloads (>10s each), the watcher pauses for that plugin

**Manual reload via CLI or API**:

```bash
modulacms plugin reload my_plugin
```

```bash
POST /api/v1/admin/plugins/{name}/reload
```

## Circuit Breaker

**Plugin-level**: After 5 consecutive failures (configurable), the plugin's circuit breaker opens. All requests return 503 until the reset interval (60s) elapses and a probe request succeeds. Feeds from HTTP handler errors and manager operations (reload, init). Does NOT include hook failures.

**Hook-level**: After 10 consecutive hook errors (configurable), that specific hook is disabled until the plugin is reloaded. Per (plugin, event, table) combination.

Admin can reset the plugin-level circuit breaker via:

```bash
modulacms plugin enable my_plugin
```

## Error Handling

| Error Source | Behavior |
|-------------|----------|
| `error()` in `on_init()` | Plugin marked as Failed, does not serve traffic |
| `error()` in `on_shutdown()` | Logged, shutdown continues |
| `error()` in before-hook | CMS transaction aborted, 422 returned to client |
| `error()` in after-hook | Logged, does not affect response |
| `error()` in HTTP handler | 500 returned to client |
| Panic in Go handler | Captured by SafeExecute, recorded as circuit breaker failure |
| Operation limit exceeded | `ErrOpLimitExceeded` raised to Lua |

## CLI Commands

### Offline (no server required)

```bash
# List discovered plugins
modulacms plugin list

# Scaffold a new plugin
modulacms plugin init my_plugin
modulacms plugin init my_plugin --version 1.0.0 --description "My plugin" --author "Dev" --license MIT

# Validate without loading
modulacms plugin validate ./plugins/my_plugin
```

### Online (requires running server)

Online commands authenticate via a Bearer token auto-generated at server startup (written to `<config_dir>/.plugin-api-token`). Pass `--token <value>` for CI/CD.

```bash
# Plugin info and status
modulacms plugin info my_plugin

# Lifecycle management
modulacms plugin reload my_plugin     # Blue-green hot reload
modulacms plugin enable my_plugin     # Reset circuit breaker, reload
modulacms plugin disable my_plugin    # Stop plugin, trip circuit breaker

# Approval (see Route & Hook Approval section above)
modulacms plugin approve my_plugin --all-routes
modulacms plugin revoke my_plugin --all-hooks
```

## Admin API

Plugin management endpoints under `/api/v1/admin/plugins/`.

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/admin/plugins` | Any authed | List all plugins with state |
| GET | `/api/v1/admin/plugins/{name}` | Any authed | Plugin details, VM counts, CB state, schema drift |
| POST | `/api/v1/admin/plugins/{name}/reload` | Admin only | Trigger hot reload (10s cooldown) |
| POST | `/api/v1/admin/plugins/{name}/enable` | Admin only | Reset circuit breaker, reload plugin |
| POST | `/api/v1/admin/plugins/{name}/disable` | Admin only | Stop plugin, trip circuit breaker |
| GET | `/api/v1/admin/plugins/cleanup` | Admin only | List orphaned plugin tables (dry-run) |
| POST | `/api/v1/admin/plugins/cleanup` | Admin only | Drop orphaned tables (requires `{"confirm": true, "tables": [...]}`) |
| GET | `/api/v1/admin/plugins/routes` | Any authed | List all routes with approval status |
| POST | `/api/v1/admin/plugins/routes/approve` | Admin only | Approve routes |
| POST | `/api/v1/admin/plugins/routes/revoke` | Admin only | Revoke route approval |
| GET | `/api/v1/admin/plugins/hooks` | Any authed | List all hooks with approval status |
| POST | `/api/v1/admin/plugins/hooks/approve` | Admin only | Approve hooks |
| POST | `/api/v1/admin/plugins/hooks/revoke` | Admin only | Revoke hook approval |

## Schema Drift Detection

After `db.define_table()` creates a table, the system introspects actual columns and compares against the definition. Mismatches (missing or extra columns) are logged as warnings and surfaced in the `GET /api/v1/admin/plugins/{name}` admin response.

Drift is advisory only. Operators should update plugin code to handle drift or manually `ALTER TABLE`.

## Configuration

All fields in `config.json` under the `plugin_*` namespace. See [CONFIGURATION.md](CONFIGURATION.md) for the complete reference.

Key fields:

| Field | Default | Description |
|-------|---------|-------------|
| `plugin_enabled` | false | Enable plugin system |
| `plugin_directory` | `./plugins/` | Path to plugins directory |
| `plugin_max_vms` | 4 | VM pool size per plugin |
| `plugin_timeout` | 5 | Execution timeout (seconds) |
| `plugin_max_ops` | 1000 | Max DB operations per VM checkout |
| `plugin_hot_reload` | false | Enable file-watching hot reload |
| `plugin_max_failures` | 5 | Circuit breaker failure threshold |
| `plugin_reset_interval` | `60s` | Circuit breaker reset interval |
| `plugin_rate_limit` | 100 | Requests per second per IP |
| `plugin_max_routes` | 50 | Max routes per plugin |
| `plugin_max_request_body` | 1048576 | Max request body bytes (1 MB) |
| `plugin_max_response_body` | 5242880 | Max response body bytes (5 MB) |
| `plugin_hook_reserve_vms` | 1 | VMs reserved for hooks per plugin |
| `plugin_hook_max_consecutive_aborts` | 10 | Hook circuit breaker threshold |
| `plugin_hook_max_ops` | 100 | Reduced op budget for after-hooks |
| `plugin_hook_timeout_ms` | 2000 | Per-hook timeout for before-hooks (ms) |
| `plugin_hook_event_timeout_ms` | 5000 | Per-event total timeout for before-hooks (ms) |

## Shutdown Order

1. Stop watcher (prevents reload during shutdown)
2. Close HTTP bridge (drain inflight requests)
3. Close hook engine (drain inflight after-hooks)
4. Shutdown manager (close pools in reverse dependency order, call `on_shutdown`)

## Developer Workflow

1. **Scaffold**: `modulacms plugin init my_plugin`
2. **Develop**: Write `init.lua`, add modules in `lib/`
3. **Validate**: `modulacms plugin validate ./plugins/my_plugin`
4. **Deploy**: Copy plugin directory to the server's `plugin_directory`
5. **Restart or hot reload**: Server picks up the plugin on startup, or the watcher auto-reloads if `plugin_hot_reload: true`
6. **Approve**: `modulacms plugin approve my_plugin --all-routes --all-hooks`
7. **Monitor**: `modulacms plugin info my_plugin` (check state, circuit breaker, schema drift)
8. **Update**: Edit Lua files; watcher auto-reloads via blue-green deployment (no downtime)
