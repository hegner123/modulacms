# Plugin System

Runtime Lua plugin system for ModulaCMS. Plugins can define database tables, handle HTTP requests, react to content lifecycle events, and integrate with the CMS without recompilation.

Built on [gopher-lua](https://github.com/yuin/gopher-lua) (pure Go Lua VM). Zero external dependencies.

## Plugin Structure

```
plugins/
  my_plugin/
    init.lua          # Entry point (required)
    lib/
      helpers.lua     # Optional modules loaded via require("helpers")
```

### Manifest

Every plugin declares metadata via a `plugin_info` global in `init.lua`:

```lua
plugin_info = {
  name         = "task_tracker",        -- required, [a-z0-9_], max 32 chars
  version      = "1.0.0",              -- required
  description  = "Task management",     -- required
  author       = "Developer Name",      -- optional
  license      = "MIT",                 -- optional
  min_cms_version = "1.0.0",           -- optional
  dependencies = {"other_plugin"},      -- optional, loaded in dependency order
}
```

### Lifecycle Functions

```lua
function on_init()
  -- Called once during plugin load.
  -- Define tables, insert seed data, register routes/hooks.
end

function on_shutdown()
  -- Called during graceful shutdown.
  -- Clean up resources.
end
```

## Lua API Reference

### Database (`db.*`)

Sandboxed database access scoped to plugin-prefixed tables (`plugin_<name>_<table>`). All tables get auto-injected `id` (ULID), `created_at`, and `updated_at` columns.

#### Table Definition

```lua
db.define_table("tasks", {
  columns = {
    {name = "title",   type = "text",    not_null = true},
    {name = "status",  type = "text",    default = "pending"},
    {name = "score",   type = "integer", unique = true},
  },
  indexes = {
    {columns = {"status"}, unique = false},
  },
  foreign_keys = {
    {column = "user_id", ref_table = "plugin_task_tracker_users",
     ref_column = "id", on_delete = "CASCADE"},
  },
})
```

Foreign key `ref_table` must reference another table owned by the same plugin (namespace isolation enforced).

#### Queries

```lua
-- SELECT with filtering, ordering, pagination
rows = db.query("tasks", {
  where    = {status = "pending"},
  order_by = "created_at DESC",
  limit    = 10,
  offset   = 0,
})

-- Single row (LIMIT 1)
row = db.query_one("tasks", {where = {id = task_id}})

-- Aggregates
count  = db.count("tasks", {where = {status = "done"}})
exists = db.exists("tasks", {where = {id = task_id}})
```

#### Mutations

```lua
db.insert("tasks", {title = "New task", status = "pending"})
-- id, created_at, updated_at are auto-set

db.update("tasks", {
  set   = {status = "done"},
  where = {id = task_id},        -- where is required (no full-table updates)
})

db.delete("tasks", {
  where = {id = task_id},        -- where is required (no full-table deletes)
})
```

#### Transactions

```lua
local ok, err = db.transaction(function()
  db.insert("tasks", {title = "Task 1"})
  db.insert("tasks", {title = "Task 2"})
  -- error() inside rolls back
end)
```

#### Utilities

```lua
id  = db.ulid()       -- Generate a new ULID
now = db.timestamp()   -- RFC3339 UTC timestamp
```

#### Limits

- Per-checkout operation budget: 1000 ops (configurable via `plugin_max_ops`)
- Before-hooks get a reduced budget (default 100 ops)
- Exceeding the limit returns `ErrOpLimitExceeded`

### HTTP Routes (`http.*`)

Register HTTP endpoints under `/api/v1/plugins/<plugin_name>/`.

```lua
http.handle("GET", "/tasks", function(req)
  local tasks = db.query("tasks", {})
  return {status = 200, json = tasks}
end)

http.handle("POST", "/tasks", function(req)
  if not req.json or not req.json.title then
    return {status = 400, json = {error = "title required"}}
  end
  db.insert("tasks", {title = req.json.title})
  return {status = 201, json = {ok = true}}
end, {public = true})  -- public = true skips CMS session auth
```

#### Request Object

| Field            | Type   | Description                                    |
|------------------|--------|------------------------------------------------|
| `req.method`     | string | HTTP method                                    |
| `req.path`       | string | URL path                                       |
| `req.body`       | string | Raw request body                               |
| `req.client_ip`  | string | Client IP (proxy-aware)                        |
| `req.headers`    | table  | All headers (lowercase keys)                   |
| `req.query`      | table  | URL query parameters                           |
| `req.params`     | table  | Path parameters from `{id}` wildcards          |
| `req.json`       | table  | Parsed JSON body (when Content-Type is JSON)   |

#### Response Object

| Field              | Type   | Description                                  |
|--------------------|--------|----------------------------------------------|
| `response.status`  | number | HTTP status code (default 200)               |
| `response.headers` | table  | Custom response headers                      |
| `response.body`    | string | Response body string                         |
| `response.json`    | table  | JSON response (takes priority over `body`)   |

#### Middleware

```lua
http.use(function(req)
  if req.headers["x-api-key"] ~= "secret" then
    return {status = 401, json = {error = "Unauthorized"}}
  end
  -- return nil to continue to the route handler
end)
```

#### Limits

- Max 50 routes per plugin
- Request body: 1 MB default (`plugin_max_request_body`)
- Response body: 5 MB default (`plugin_max_response_body`)
- Rate limit: 100 req/sec per IP (`plugin_rate_limit`)
- Execution timeout: 5s default (`plugin_timeout`)
- Valid methods: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`
- Path rules: must start with `/`, no `..`, `?`, `#`, max 256 chars

#### Blocked Response Headers

Plugins cannot set: `access-control-*`, `set-cookie`, `transfer-encoding`, `content-length`, `host`, `connection`, `cache-control`.

### Content Hooks (`hooks.*`)

React to content lifecycle events (insert, update, delete) on audited tables.

```lua
-- Before-hooks run synchronously inside the DB transaction.
-- Return nothing to allow, call error() to abort.
hooks.on("before_insert", "content_data", function(data)
  if not data.slug or data.slug == "" then
    error("slug is required")
  end
end, {priority = 50})

-- After-hooks run asynchronously (fire-and-forget).
hooks.on("after_update", "content_data", function(data)
  log.info("Content updated", {id = data.id})
end)

-- Wildcard: runs for all tables.
hooks.on("after_delete", "*", function(data)
  log.warn("Record deleted", {table = data._table, id = data.id})
end)
```

#### Events

| Event             | Timing        | Can abort? | Use case                    |
|-------------------|---------------|------------|-----------------------------|
| `before_insert`   | Synchronous   | Yes        | Validation, enrichment      |
| `after_insert`    | Asynchronous  | No         | Notifications, indexing     |
| `before_update`   | Synchronous   | Yes        | Validation, field guards    |
| `after_update`    | Asynchronous  | No         | Cache invalidation          |
| `before_delete`   | Synchronous   | Yes        | Referential integrity       |
| `after_delete`    | Asynchronous  | No         | Cleanup, audit logging      |

#### Options

- `priority`: 1-1000, lower runs first (default 100)
- Max 50 hooks per plugin
- Before-hook timeout: 2s per hook, 5s per event chain
- After-hooks use reduced op budget (100 ops) and bounded concurrency (10)
- Before-hooks block `db.*` calls to prevent SQLite transaction deadlocks

#### Circuit Breaker

Each (plugin, event, table) combination has a circuit breaker. After 10 consecutive aborts, that hook is disabled until admin re-enables it via `SetHookEnabled` or the plugin is reloaded.

### Logging (`log.*`)

Structured logging with automatic `plugin: <name>` context.

```lua
log.info("Task created", {id = task_id, user = user_name})
log.warn("Deprecated API used")
log.error("Failed to process", {reason = err})
log.debug("Trace detail", {step = 5, value = result})
```

### Module System (`require`)

Plugins can split code into modules under `lib/`:

```lua
-- lib/validators.lua
return {
  not_empty = function(s) return s and #s > 0 end,
}

-- init.lua
local v = require("validators")
if v.not_empty(title) then ... end
```

Only files under the plugin's own `lib/` directory are loadable (sandbox enforced).

## Approval System

Routes and hooks require admin approval before they become active. This prevents untrusted plugins from exposing endpoints or intercepting content operations without review.

### Route Approval

```
GET  /api/v1/admin/plugins/routes          -- List all routes with approval status
POST /api/v1/admin/plugins/routes/approve   -- Approve a route
POST /api/v1/admin/plugins/routes/revoke    -- Revoke approval
```

### Hook Approval

Hooks are approved/revoked via the `HookEngine.ApproveHook` and `RevokeHook` methods. Unapproved hooks are silently skipped during dispatch.

### Version Change

When a plugin's version changes, all route and hook approval rows are deleted and re-created as unapproved. This forces re-review after plugin updates.

## CLI Commands

Plugin management via the `modulacms plugin` command. Commands are split into offline (filesystem only, no server needed) and online (requires a running server).

### Offline Commands

Offline commands read from the filesystem and config.json only. They work regardless of `plugin_enabled` and do not require a database connection.

#### `plugin list`

Scan the plugin directory and print a summary table.

```
$ modulacms plugin list
NAME              VERSION   DESCRIPTION
hello_world       1.0.0     Example greeting plugin
task_tracker      2.1.0     Task management system
broken_plugin [invalid]
```

Reads `plugin_directory` from config.json. Calls `ValidatePlugin()` on each subdirectory. Invalid plugins are shown with `[invalid]` marker.

#### `plugin init <name>`

Scaffold a new plugin directory with `init.lua` and `lib/`.

```
$ modulacms plugin init my_plugin
```

Validates the name using the same rules as the plugin Manager (`[a-z0-9_]`, max 32 chars, no trailing underscore). Creates `<plugin_directory>/<name>/` and `<name>/lib/`.

Interactive mode (default when stdout is a terminal): prompts for version, description, author, license via `huh.Form`.

Non-interactive mode (all flags provided, or stdout is not a terminal):

```
$ modulacms plugin init my_plugin \
    --version 1.0.0 \
    --description "My plugin" \
    --author "Dev" \
    --license MIT
```

| Flag | Default | Required (non-interactive) |
|------|---------|---------------------------|
| `--version` | `0.1.0` | No |
| `--description` | | Yes |
| `--author` | | No |
| `--license` | `MIT` | No |

#### `plugin validate <path>`

Validate a plugin directory without loading it into the runtime.

```
$ modulacms plugin validate ./plugins/my_plugin
Plugin "my_plugin" v1.0.0 is valid.
  1 warning(s) found.
```

Checks: directory exists, `init.lua` exists, Lua syntax (parse+compile without executing), manifest fields, name rules. Returns structured errors (exit 1) and warnings (exit 0).

### Online Commands

Online commands send HTTP requests to the running server's admin API. Authentication uses a Bearer token auto-generated at server startup and written to `<config_dir>/.plugin-api-token`.

State changes (enable/disable/reload) are handled in-process by the Manager -- no server restart needed.

All online commands accept `--token <value>` to override the token file (for CI/CD or remote use).

#### `plugin info <name>`

Show detailed plugin information from the running server.

```
$ modulacms plugin info task_tracker
Name:          task_tracker
Version:       2.1.0
Description:   Task management system
Author:        Dev
License:       MIT
State:         running
VMs:           4/4 available
```

Fields shown: name, version, description, author, license, state, failed reason (if failed), circuit breaker state/errors, VMs available/total, dependencies, schema drift.

#### `plugin reload <name>`

Trigger a hot reload of a plugin. Uses blue-green deployment internally.

```
$ modulacms plugin reload task_tracker
Plugin "task_tracker" reloaded successfully.
```

#### `plugin enable <name>`

Re-enable a disabled plugin. Resets the circuit breaker and reloads the plugin.

```
$ modulacms plugin enable task_tracker
Plugin "task_tracker" enabled successfully.
```

#### `plugin disable <name>`

Disable a running plugin. Stops the plugin and trips the circuit breaker.

```
$ modulacms plugin disable task_tracker
Plugin "task_tracker" disabled successfully.
```

#### `plugin approve <name>`

Approve routes and/or hooks for a plugin to become active.

```
# Approve a single route
$ modulacms plugin approve my_plugin --route "GET /tasks"

# Approve a single hook
$ modulacms plugin approve my_plugin --hook "after_insert:content_data"

# Approve all pending routes (with confirmation)
$ modulacms plugin approve my_plugin --all-routes

# Approve all pending hooks (skip prompt for CI/CD)
$ modulacms plugin approve my_plugin --all-hooks --yes
```

| Flag | Description |
|------|-------------|
| `--route "METHOD /path"` | Approve a specific route |
| `--hook "event:table"` | Approve a specific hook |
| `--all-routes` | Approve all unapproved routes (prompts for confirmation) |
| `--all-hooks` | Approve all unapproved hooks (prompts for confirmation) |
| `--yes` | Skip confirmation prompt |

Bulk operations (`--all-routes`, `--all-hooks`) list pending items to stderr and prompt before proceeding. The prompt is skipped with `--yes` or when stdout is not a terminal.

Idempotent: approving an already-approved item is a no-op. When zero pending items exist, prints "No pending routes/hooks" and exits 0.

#### `plugin revoke <name>`

Revoke approvals for routes and/or hooks. Same flags as `approve`.

```
$ modulacms plugin revoke my_plugin --route "GET /tasks"
$ modulacms plugin revoke my_plugin --all-hooks --yes
```

Idempotent: revoking an already-revoked item is a no-op.

### Justfile Recipes

```
just plugin-list                    # List installed plugins
just plugin-init <name>             # Create a new plugin scaffold
just plugin-validate <path>         # Validate a plugin
just plugin-info <name>             # Show plugin details (server required)
just plugin-reload <name>           # Reload a plugin (server required)
just plugin-enable <name>           # Enable a plugin (server required)
just plugin-disable <name>          # Disable a plugin (server required)
```

## TUI (SSH)

The SSH TUI includes a Plugins page accessible from the homepage menu. SSH access is inherently admin-level.

**Plugins List Page**: 3-panel layout showing all plugins with Name, Version, State, and Circuit Breaker columns. Select a plugin to navigate to its detail page.

**Plugin Detail Page**: Shows plugin metadata and an action menu:

| Action | Description |
|--------|-------------|
| Enable Plugin | Re-enable a disabled plugin (resets CB) |
| Disable Plugin | Stop the plugin (trips CB) |
| Reload Plugin | Blue-green hot reload |
| Approve Routes | Lists pending routes, shows confirmation dialog, then approves |
| Approve Hooks | Lists pending hooks, shows confirmation dialog, then approves |

Approve actions always show a confirmation dialog listing the pending items before executing.

## Admin API

Plugin management endpoints registered at `/api/v1/admin/plugins/`.

### Plugin Management

| Method | Path                                      | Auth       | Description                           |
|--------|-------------------------------------------|------------|---------------------------------------|
| GET    | `/api/v1/admin/plugins`                   | Any authed | List all plugins with state           |
| GET    | `/api/v1/admin/plugins/{name}`            | Any authed | Plugin details, VM counts, CB state, schema drift |
| POST   | `/api/v1/admin/plugins/{name}/reload`     | Admin only | Trigger hot reload (10s cooldown)     |
| POST   | `/api/v1/admin/plugins/{name}/enable`     | Admin only | Reset circuit breaker, reload plugin  |
| POST   | `/api/v1/admin/plugins/{name}/disable`    | Admin only | Stop plugin, trip circuit breaker     |
| GET    | `/api/v1/admin/plugins/cleanup`           | Admin only | List orphaned plugin tables (dry-run) |
| POST   | `/api/v1/admin/plugins/cleanup`           | Admin only | Drop orphaned tables (requires `{"confirm": true, "tables": [...]}`) |

### Route Approval

Registered in `internal/router/mux.go`. All routes require admin approval before becoming active.

| Method | Path                                      | Auth       | Description |
|--------|-------------------------------------------|------------|-------------|
| GET    | `/api/v1/admin/plugins/routes`            | Any authed | List all plugin routes with approval status |
| POST   | `/api/v1/admin/plugins/routes/approve`    | Admin only | Approve routes |
| POST   | `/api/v1/admin/plugins/routes/revoke`     | Admin only | Revoke route approval |

**Approve/Revoke request body:**

```json
{
  "routes": [
    {"plugin": "my_plugin", "method": "GET", "path": "/tasks"}
  ]
}
```

**List response:**

```json
{
  "routes": [
    {
      "plugin": "my_plugin",
      "method": "GET",
      "path": "/tasks",
      "approved": false,
      "public": false,
      "plugin_version": "1.0.0"
    }
  ]
}
```

### Hook Approval

Registered in `internal/plugin/http_bridge.go` via `MountAdminEndpoints()`.

| Method | Path                                      | Auth       | Description |
|--------|-------------------------------------------|------------|-------------|
| GET    | `/api/v1/admin/plugins/hooks`             | Any authed | List all registered hooks with approval status |
| POST   | `/api/v1/admin/plugins/hooks/approve`     | Admin only | Approve hooks |
| POST   | `/api/v1/admin/plugins/hooks/revoke`      | Admin only | Revoke hook approval |

**Approve/Revoke request body:**

```json
{
  "hooks": [
    {"plugin": "my_plugin", "event": "after_insert", "table": "content_data"}
  ]
}
```

**List response:**

```json
{
  "hooks": [
    {
      "plugin_name": "my_plugin",
      "event": "after_insert",
      "table": "content_data",
      "priority": 100,
      "approved": false,
      "is_wildcard": false
    }
  ]
}
```

**Idempotency:** Approving an already-approved item or revoking an already-revoked item is a no-op (200 OK, no error).

**Body limit:** All POST endpoints enforce a 1 MB body limit via `http.MaxBytesReader`.

**Error response format:**

```json
{"errors": ["hook not found: my_plugin:bad_event:bad_table"]}
```

### Authentication

Online CLI commands authenticate using a Bearer token auto-generated at server startup:

1. `cmd/serve.go` generates a 32-byte `crypto/rand` token, hex-encodes it
2. The token is inserted into the `tokens` table (type `api_key`, tied to the system user, no expiry)
3. The token value is written to `<config_dir>/.plugin-api-token` (mode 0600)
4. The existing `APIKeyAuth` middleware validates `Authorization: Bearer <token>` against the tokens table
5. On graceful shutdown, the token row is deleted and the file removed
6. Stale tokens from ungraceful shutdowns are cleaned up on next startup

For CI/CD, pass `--token <value>` to any online command instead of relying on the token file.

## VM Pool

Each plugin gets a pool of pre-initialized Lua VMs (default 4, configurable via `plugin_max_vms`).

- **General pool**: Serves HTTP requests
- **Reserved pool**: Dedicated VMs for content hooks (default 1, `plugin_hook_reserve_vms`)
- **Checkout timeout**: 100ms (returns `ErrPoolExhausted` if none available)
- **Health validation**: VMs are checked on every `Put()` call; corrupted VMs are replaced
- **Global snapshot**: Baseline globals captured after `init.lua`; drift detected on return

### Tri-State Lifecycle

| State    | `Get()` behavior      | `Put()` behavior           |
|----------|-----------------------|----------------------------|
| Open     | Normal checkout       | Return to channel          |
| Draining | `ErrPoolExhausted`    | Return to channel (counted)|
| Closed   | `ErrPoolExhausted`    | Close VM directly          |

## Sandboxing

Lua VMs run in a restricted environment:

**Allowed stdlib**: `base`, `table`, `string`, `math`

**Removed**: `io`, `os`, `package`, `debug`, `channel`

**Stripped globals**: `dofile`, `loadfile`, `load`, `rawget`, `rawset`, `rawequal`, `rawlen`

**Kept**: `setmetatable`, `getmetatable`, `pcall`, `xpcall`, `error`, `type`, `tostring`, `tonumber`, `pairs`, `ipairs`, `next`, `select`, `unpack`

**Module freeze**: All registered API modules (`db`, `http`, `hooks`, `log`) are made read-only via metatable proxy after registration.

## Circuit Breaker

Plugin-level circuit breaker tracks consecutive failures from HTTP handlers and manager operations (reload, init). Hook failures are tracked separately by the hook-level circuit breaker and do not feed into the plugin-level CB.

| State    | Behavior                                                      |
|----------|---------------------------------------------------------------|
| Closed   | Normal operation                                              |
| Open     | All requests rejected with 503; auto-probe after reset interval |
| HalfOpen | One probe request allowed; success closes, failure re-opens   |

Defaults: 5 consecutive failures to trip, 60s reset interval. Configurable via `plugin_max_failures` and `plugin_reset_interval`.

Admin reset via `POST /api/v1/admin/plugins/{name}/enable` emits an audit event with admin user, plugin name, prior CB state, and failure count.

## Hot Reload

File-polling watcher detects changes to `.lua` files and triggers per-plugin reload. Disabled by default (`plugin_hot_reload: false`).

### Behavior

- Poll interval: 2s
- Debounce: 1s stability window (waits for file writes to settle)
- Cooldown: 10s between reloads per plugin
- Checksum: SHA-256 of all `.lua` files (content + filenames)
- Symlinks and non-regular files are skipped
- Max 100 `.lua` files and 10 MB total per checksumming pass

### Blue-Green Reload

1. Create new plugin instance (new pool, new DB APIs)
2. Run `on_init`, register routes and hooks on new instance
3. If new instance **fails**: old instance keeps running, warning logged
4. If new instance **succeeds**: swap in-memory reference under lock, then drain old pool
5. If drain times out: trip circuit breaker on new instance (requires admin reset)

### Slow Reload Protection

After 3 consecutive reloads exceeding 10s, the watcher pauses polling for that plugin. Admin can still force reload via the API.

## Schema Drift Detection

After `db.define_table()` creates a table, the system introspects actual columns and compares against the definition. Mismatches (missing or extra columns) are logged as warnings and surfaced in the `GET /api/v1/admin/plugins/{name}` admin response.

Drift is advisory only. Operators should update plugin code to handle drift or manually `ALTER TABLE`.

## Plugin Lifecycle States

| State        | Description                                         |
|--------------|-----------------------------------------------------|
| `discovered` | Directory found, manifest not yet parsed            |
| `loading`    | `init.lua` executing, tables/routes/hooks registering|
| `running`    | Fully operational, serving requests                 |
| `failed`     | Load or runtime error (reason stored in `FailedReason`)|
| `stopped`    | Disabled by admin                                   |

## Database Tables

The plugin system creates two system tables:

### `plugin_routes`

HTTP route registrations with approval status.

| Column           | Type | Description                  |
|------------------|------|------------------------------|
| `plugin_name`    | TEXT | Plugin identifier            |
| `method`         | TEXT | HTTP method                  |
| `path`           | TEXT | Route path                   |
| `public`         | BOOL | Bypasses CMS session auth    |
| `approved`       | BOOL | Admin-approved for serving   |
| `approved_at`    | TEXT | Approval timestamp           |
| `approved_by`    | TEXT | Approving admin              |
| `plugin_version` | TEXT | Version at registration time |
| `created_at`     | TEXT | Row creation timestamp       |

### `plugin_hooks`

Content lifecycle hook registrations with approval status.

| Column           | Type | Description                  |
|------------------|------|------------------------------|
| `plugin_name`    | TEXT | Plugin identifier            |
| `event`          | TEXT | Hook event type              |
| `table_name`     | TEXT | Target table or `*`          |
| `approved`       | BOOL | Admin-approved for dispatch  |
| `approved_at`    | TEXT | Approval timestamp           |
| `approved_by`    | TEXT | Approving admin              |
| `plugin_version` | TEXT | Version at registration time |

### Plugin-defined tables

Tables created via `db.define_table()` are prefixed with `plugin_<name>_` and get auto-injected `id`, `created_at`, `updated_at` columns.

## Configuration

All fields in `config.json` under the `plugin_*` namespace:

| Field                                | Type     | Default   | Description                                    |
|--------------------------------------|----------|-----------|------------------------------------------------|
| `plugin_enabled`                     | bool     | false     | Enable plugin system                           |
| `plugin_directory`                   | string   | ./plugins/| Path to plugins directory                      |
| `plugin_max_vms`                     | int      | 4         | VM pool size per plugin                        |
| `plugin_timeout`                     | int      | 5         | Execution timeout (seconds)                    |
| `plugin_max_ops`                     | int      | 1000      | Max DB operations per VM checkout              |
| `plugin_db_max_open_conns`           | int      | 0         | Plugin DB pool max open connections            |
| `plugin_db_max_idle_conns`           | int      | 0         | Plugin DB pool max idle connections            |
| `plugin_db_conn_max_lifetime`        | string   | ""        | Plugin DB connection max lifetime              |
| `plugin_max_request_body`            | int64    | 1048576   | Max request body bytes (1 MB)                  |
| `plugin_max_response_body`           | int64    | 5242880   | Max response body bytes (5 MB)                 |
| `plugin_rate_limit`                  | int      | 100       | Requests per second per IP                     |
| `plugin_max_routes`                  | int      | 50        | Max routes per plugin                          |
| `plugin_trusted_proxies`             | []string | []        | CIDR list for proxy-aware IP extraction        |
| `plugin_hook_reserve_vms`            | int      | 1         | VMs reserved for hooks per plugin              |
| `plugin_hook_max_consecutive_aborts` | int      | 10        | Hook circuit breaker threshold                 |
| `plugin_hook_max_ops`                | int      | 100       | Reduced op budget for after-hooks              |
| `plugin_hook_max_concurrent_after`   | int      | 10        | Max concurrent after-hook goroutines           |
| `plugin_hook_timeout_ms`             | int      | 2000      | Per-hook timeout for before-hooks (ms)         |
| `plugin_hook_event_timeout_ms`       | int      | 5000      | Per-event total timeout for before-hooks (ms)  |
| `plugin_hot_reload`                  | bool     | false     | Enable file-watching hot reload                |
| `plugin_max_failures`                | int      | 5         | Circuit breaker failure threshold              |
| `plugin_reset_interval`              | string   | 60s       | Circuit breaker reset interval                 |

## Metrics

Recorded via `utility.GlobalMetrics`. Coarse-grained instrumentation at HTTP request and hook event boundaries.

| Metric                         | Labels                              | Description                    |
|--------------------------------|-------------------------------------|--------------------------------|
| `plugin.http.requests`         | plugin, method, status              | HTTP request count             |
| `plugin.http.duration_ms`      | plugin, method, status              | HTTP request duration          |
| `plugin.hook.before`           | plugin, event, table, status        | Before-hook execution count    |
| `plugin.hook.after`            | plugin, event, table, status        | After-hook execution count     |
| `plugin.hook.duration_ms`      | plugin, event, table, status        | Hook execution duration        |
| `plugin.errors`                | plugin, type                        | Error count by type            |
| `plugin.circuit_breaker.trip`  | plugin                              | Circuit breaker trip count     |
| `plugin.reload`                | plugin, status                      | Reload event count             |
| `plugin.vm.available`          | plugin                              | Available VMs (periodic gauge) |

## Shutdown Order

1. Stop watcher (prevents reload during shutdown)
2. Close HTTP bridge (drain inflight requests)
3. Close hook engine (drain inflight after-hooks)
4. Shutdown manager (close pools in reverse dependency order, call `on_shutdown`)

## File Map

| File               | Purpose                                              |
|--------------------|------------------------------------------------------|
| `manager.go`       | Plugin discovery, loading, dependency resolution, lifecycle |
| `pool.go`          | VM pool with checkout/return, health validation, drain |
| `sandbox.go`       | VM security restrictions, stdlib filtering, module freeze |
| `db_api.go`        | `db.*` Lua module (query, insert, update, delete, transaction) |
| `schema_api.go`    | `db.define_table()` DDL, schema drift detection      |
| `hooks_api.go`     | `hooks.on()` Lua registration API                    |
| `hook_engine.go`   | Hook dispatch, approval, circuit breaker, before/after execution |
| `http_api.go`      | `http.handle()` and `http.use()` Lua registration    |
| `http_bridge.go`   | HTTP dispatcher, route approval, rate limiting, admin mount |
| `http_request.go`  | HTTP request/response Lua conversion                 |
| `log_api.go`       | `log.*` Lua module                                   |
| `lua_helpers.go`   | Lua/Go type conversion utilities                     |
| `recovery.go`      | Plugin-level circuit breaker, panic-safe execution    |
| `metrics.go`       | Metric recording helpers                             |
| `watcher.go`       | File-polling hot reload with debounce                |
| `cli_commands.go`  | Admin HTTP endpoint handlers (list, info, reload, enable, disable, cleanup) |
| `hook_handlers.go` | Hook list/approve/revoke HTTP handlers               |
| `validator.go`     | Offline plugin validation (ValidatePlugin, checkSyntax) |
| `validator_test.go`| Validator tests against testdata fixtures            |

### CLI Files

| File               | Purpose                                              |
|--------------------|------------------------------------------------------|
| `cmd/plugin.go`    | All CLI subcommands (list, init, validate, info, reload, enable, disable, approve, revoke) |
| `cmd/root.go`      | Plugin command registration                          |

## Example Plugins

Two example plugins in `examples/plugins/` demonstrate the full plugin API:

### `hello_world`

Minimal getting-started plugin. Single `GET` route returning `{"message": "Hello from ModulaCMS!"}`. Heavily commented explaining manifest fields, `on_init`/`on_shutdown` lifecycle, and `http.handle` usage.

### `task_tracker`

Full-featured example showing all major APIs:
- `db.define_table` for tasks and categories tables
- CRUD routes (`GET`/`POST`/`PUT`/`DELETE` on `/tasks` and `/categories`)
- `hooks.on` for before/after content hooks
- `db.transaction` for seed data
- `require("validators")` to load `lib/validators.lua` module
- `log.info`, `log.warn`, `log.debug` structured logging
- `db.count`, `db.exists`, `db.query_one`, `db.ulid`, `db.timestamp`
