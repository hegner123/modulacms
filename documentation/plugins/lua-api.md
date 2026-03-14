# Lua API Reference

Complete reference for all APIs available to Lua plugins in ModulaCMS. Every function, every parameter, every return value.

Source files: `internal/plugin/db_api.go`, `internal/plugin/http_api.go`, `internal/plugin/hooks_api.go`, `internal/plugin/log_api.go`, `internal/plugin/sandbox.go`, `internal/plugin/ui_api.go`, `internal/plugin/ui_bridge.go`, `internal/plugin/ui_primitives.go`

## db -- Database API

All table names are auto-prefixed with `plugin_<name>_`. Lua code uses short names only (e.g., `"tasks"` becomes `plugin_task_tracker_tasks` in SQL). Plugins cannot access CMS tables or other plugins' tables.

### db.define_table(table, definition)

Creates a plugin table using `CREATE TABLE IF NOT EXISTS`. Call inside `on_init()` only.

Three columns are auto-injected on every table. Do not include them in your columns list:

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PRIMARY KEY | ULID, auto-generated |
| `created_at` | TEXT | RFC3339 UTC timestamp |
| `updated_at` | TEXT | RFC3339 UTC timestamp |

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `table` | string | Yes | Short table name (no prefix) |
| `definition.columns` | table | Yes | Array of column definitions |
| `definition.indexes` | table | No | Array of index definitions |
| `definition.foreign_keys` | table | No | Array of foreign key definitions |

**Column definition fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Column name |
| `type` | string | Yes | One of: `text`, `integer`, `boolean`, `real`, `timestamp`, `json`, `blob` |
| `not_null` | boolean | No | Add NOT NULL constraint |
| `default` | any | No | Default value |

**Index definition fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `columns` | table | Yes | Array of column name strings |
| `unique` | boolean | No | Create unique index (default `false`) |

**Foreign key definition fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `column` | string | Yes | Column in this table |
| `ref_table` | string | Yes | Referenced table (full prefixed name, e.g., `plugin_bookmarks_collections`) |
| `ref_column` | string | Yes | Referenced column |
| `on_delete` | string | No | `CASCADE`, `SET NULL`, `RESTRICT`, or `NO ACTION` |

**Raises error if:**
- Reserved column name used (`id`, `created_at`, `updated_at`)
- Empty columns list
- Invalid column type
- Foreign key references a table outside the plugin's namespace

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
        { name = "attachment",  type = "blob" },
    },
    indexes = {
        { columns = {"status"} },
        { columns = {"category_id"} },
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

### db.insert(table, values)

Inserts a row. Auto-sets `id` (ULID), `created_at`, and `updated_at` if not provided. Explicit values are never overridden.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `table` | string | Yes | Short table name |
| `values` | table | Yes | Column-value pairs |

**Returns:** Nothing on success. On error: raises Lua error.

```lua
db.insert("tasks", {
    id     = db.ulid(),        -- optional, auto-generated if omitted
    title  = "Fix bug",
    status = "todo",
})
```

### db.query(table, opts) -> table

Returns an array of row tables. Returns empty table `{}` on no matches (never `nil`).

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `table` | string | Yes | -- | Short table name |
| `opts.where` | table | No | nil | Column=value equality filters (AND logic) |
| `opts.order_by` | string | No | nil | SQL ORDER BY clause (e.g., `"created_at DESC"`) |
| `opts.limit` | number | No | 100 | Max rows returned |
| `opts.offset` | number | No | 0 | Skip N rows |

**Returns:** Array table of row tables. Each row is a table with column names as keys.

```lua
local tasks = db.query("tasks", {
    where    = { status = "todo", category_id = "01ABC..." },
    order_by = "created_at DESC",
    limit    = 50,
    offset   = 0,
})

for _, task in ipairs(tasks) do
    log.info("Task: " .. task.title)
end
```

Omitting `where` returns all rows up to the limit.

### db.query_one(table, opts) -> table or nil

Returns a single row table, or `nil` if no match.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `table` | string | Yes | Short table name |
| `opts.where` | table | No | Column=value equality filters |
| `opts.order_by` | string | No | SQL ORDER BY clause |

**Returns:** Row table with column names as keys, or `nil`.

```lua
local task = db.query_one("tasks", { where = { id = task_id } })
if not task then
    return { status = 404, json = { error = "not found" } }
end
```

### db.count(table, opts) -> number

Returns row count as integer.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `table` | string | Yes | Short table name |
| `opts.where` | table | No | Column=value equality filters |

**Returns:** Integer count.

```lua
local total = db.count("tasks", {})                              -- all rows
local done  = db.count("tasks", { where = { status = "done" } }) -- filtered
```

### db.exists(table, opts) -> boolean

Returns `true` if at least one row matches.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `table` | string | Yes | Short table name |
| `opts.where` | table | No | Column=value equality filters |

**Returns:** Boolean.

```lua
if not db.exists("tasks", { where = { id = id } }) then
    return { status = 404, json = { error = "not found" } }
end
```

### db.update(table, opts)

Updates rows matching `where`. Both `set` and `where` are required and must be non-empty. This safety constraint prevents accidental full-table updates. Auto-sets `updated_at` if not included in `set`.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `table` | string | Yes | Short table name |
| `opts.set` | table | Yes | Column=value pairs to update (must be non-empty) |
| `opts.where` | table | Yes | Column=value equality filters (must be non-empty) |

**Returns:** Nothing on success. On error: raises Lua error.

```lua
db.update("tasks", {
    set   = { status = "done", title = "Fixed bug" },
    where = { id = task_id },
})
```

### db.delete(table, opts)

Deletes rows matching `where`. `where` is required and must be non-empty. This safety constraint prevents accidental full-table deletes.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `table` | string | Yes | Short table name |
| `opts.where` | table | Yes | Column=value equality filters (must be non-empty) |

**Returns:** Nothing on success. On error: raises Lua error.

```lua
db.delete("tasks", { where = { id = task_id } })
```

### db.transaction(fn) -> boolean, string|nil

Wraps multiple operations in a single database transaction. If `fn` calls `error()` or any `db.*` call inside fails, the entire transaction rolls back. Nested transactions are rejected.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `fn` | function | Yes | Function containing database operations |

**Returns:** `true, nil` on commit. `false, error_message` on rollback.

```lua
local ok, err = db.transaction(function()
    db.insert("categories", { name = "Bug" })
    db.insert("categories", { name = "Feature" })
    -- error() inside rolls back the entire transaction
end)

if not ok then
    log.warn("Transaction failed", { err = err })
end
```

### db.ulid() -> string

Generates a 26-character ULID (Universally Unique Lexicographically Sortable Identifier). Time-sortable and globally unique.

**Returns:** String, 26 characters.

```lua
local id = db.ulid()  -- e.g., "01HXYZ..."
```

### db.timestamp() -> string

Returns current UTC time as an RFC3339 string. This replaces `os.date()`, which is sandboxed out.

**Returns:** String in RFC3339 format.

```lua
local now = db.timestamp()  -- e.g., "2026-02-15T12:00:00Z"
```

---

## http -- HTTP Route API

### http.handle(method, path, handler [, options])

Registers an HTTP route. Must be called at module scope (top-level code), not inside `on_init()`.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `method` | string | Yes | `GET`, `POST`, `PUT`, `DELETE`, or `PATCH` |
| `path` | string | Yes | Starts with `/`, max 256 characters. Supports `{param}` path parameters. No `..`, `?`, or `#`. |
| `handler` | function | Yes | Receives request table, must return response table |
| `options` | table | No | `{ public = true }` bypasses CMS session auth (default: authenticated) |

**Full URL:** `/api/v1/plugins/<plugin_name><path>`

Routes require admin approval before serving traffic. Unapproved routes return 404.

```lua
http.handle("GET", "/tasks/{id}", function(req)
    local task = db.query_one("tasks", { where = { id = req.params.id } })
    if not task then
        return { status = 404, json = { error = "not found" } }
    end
    return { status = 200, json = task }
end, { public = true })
```

### Request Table

The `req` table passed to route handlers and middleware:

| Field | Type | Description |
|-------|------|-------------|
| `req.method` | string | HTTP method (e.g., `"GET"`, `"POST"`) |
| `req.path` | string | Full URL path |
| `req.body` | string | Raw request body |
| `req.client_ip` | string | Client IP address (proxy-aware via trusted proxies, no port) |
| `req.headers` | table | All request headers (keys are lowercase) |
| `req.query` | table | URL query parameters (`?name=value`) |
| `req.params` | table | Path parameters from `{param}` wildcards |
| `req.json` | table | Parsed JSON body (present only when Content-Type is `application/json`) |

### Response Table

The table returned by route handlers:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | number | 200 | HTTP status code |
| `json` | table | nil | Serialized as JSON response (sets `Content-Type: application/json` automatically) |
| `body` | string | nil | Raw string response body (used only if `json` is nil) |
| `headers` | table | nil | Custom response headers (key-value string pairs) |

If both `json` and `body` are nil, an empty response body is sent. If both are set, `json` takes precedence.

**Blocked response headers** (set by plugins raises no error, but the header is silently dropped):

| Header | Reason |
|--------|--------|
| `access-control-*` | Prevents CORS policy override |
| `set-cookie` | Prevents session manipulation |
| `transfer-encoding` | Prevents response smuggling |
| `content-length` | Prevents response smuggling |
| `cache-control` | Prevents cache poisoning |
| `host` | Prevents request smuggling |
| `connection` | Prevents request smuggling |

Two security headers are set automatically on all plugin responses: `X-Content-Type-Options: nosniff` and `X-Frame-Options: DENY`.

### http.use(middleware_function)

Appends middleware that runs before all route handlers for this plugin. Middleware functions receive the same request table as route handlers.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `middleware_function` | function | Yes | Receives request table. Return response table to short-circuit, or `nil` to continue. |

Middleware executes in registration order. The first middleware to return a non-nil response short-circuits the chain.

```lua
http.use(function(req)
    if not req.headers["x-api-key"] then
        return { status = 401, json = { error = "missing api key" } }
    end
    return nil  -- continue to route handler
end)
```

---

## hooks -- Content Lifecycle Hooks

### hooks.on(event, table, handler [, options])

Registers a content lifecycle hook. Must be called at module scope (top-level code), not inside `on_init()`.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `event` | string | Yes | Hook event name (see events table) |
| `table` | string | Yes | CMS table name (e.g., `"content_data"`), or `"*"` for wildcard |
| `handler` | function | Yes | Receives data table with entity fields |
| `options` | table | No | `{ priority = <1-1000> }` (lower runs first, default 100) |

Hooks require admin approval before they fire. Unapproved hooks are silently skipped.

Max 50 hooks per plugin.

### Hook Events

| Event | Timing | Can Abort? | db.* Access? | Notes |
|-------|--------|------------|--------------|-------|
| `before_create` | Inside CMS transaction | Yes (`error()`) | No | |
| `after_create` | After commit | No | Yes | |
| `before_update` | Inside CMS transaction | Yes (`error()`) | No | |
| `after_update` | After commit | No | Yes | |
| `before_delete` | Inside CMS transaction | Yes (`error()`) | No | |
| `after_delete` | After commit | No | Yes | |
| `before_publish` | Inside CMS transaction | Yes (`error()`) | No | Fires on status transition to `"published"` (`content_data` table only) |
| `after_publish` | After commit | No | Yes | Fires on status transition to `"published"` (`content_data` table only) |
| `before_archive` | Inside CMS transaction | Yes (`error()`) | No | Fires on status transition to `"archived"` (`content_data` table only) |
| `after_archive` | After commit | No | Yes | Fires on status transition to `"archived"` (`content_data` table only) |

**Before-hooks** run synchronously inside the CMS database transaction. Calling `error()` aborts the transaction and returns HTTP 422 to the client. `db.*` calls are blocked inside before-hooks because plugin database operations use a separate connection pool that would deadlock with the active CMS transaction (especially on SQLite).

**After-hooks** run asynchronously (fire-and-forget) after the CMS transaction commits. Errors are logged but do not affect the HTTP response. After-hooks have full `db.*` access with a reduced operation budget (default 100 ops vs 1000 for HTTP handlers).

**Wildcard hooks** (table `"*"`) run for all CMS tables. At equal priority, table-specific hooks run before wildcard hooks.

### Handler Data Table

The `data` table passed to hook handlers:

| Field | Type | Description |
|-------|------|-------------|
| `data._table` | string | The CMS table name that triggered the event |
| `data._event` | string | The event name (e.g., `"before_create"`) |
| `data.*` | varies | All entity fields from the CMS table row |

### Hook Priority

Priority range: 1 to 1000. Lower values run first. Default: 100.

At equal priority, ordering is:
1. Table-specific hooks before wildcard hooks
2. Registration order within the same plugin

```lua
hooks.on("before_create", "content_data", function(data)
    if data.title and data.title == "" then
        error("title cannot be empty")
    end
end, { priority = 50 })  -- runs before default priority (100)
```

### Hook Timeouts

| Timeout | Default | Config Key | Description |
|---------|---------|------------|-------------|
| Per-hook | 2000ms | `plugin_hook_timeout_ms` | Maximum execution time for a single before-hook |
| Per-event chain | 5000ms | `plugin_hook_event_timeout_ms` | Maximum total time for all before-hooks on one event |

After-hooks use the general execution timeout (`plugin_timeout`).

### Hook Circuit Breaker

Each (plugin, event, table) combination has its own circuit breaker. After `plugin_hook_max_consecutive_aborts` consecutive errors (default 10), that specific hook is disabled until the plugin is reloaded or re-enabled. Hook failures do not feed into the plugin-level circuit breaker.

---

## log -- Structured Logging

All log functions accept an optional context table. Plugin name is automatically included in every log entry.

### log.info(message [, context])

Log an informational message.

```lua
log.info("Task created", { id = task_id, title = "Fix bug" })
```

### log.warn(message [, context])

Log a warning.

```lua
log.warn("Category seeding failed", { err = err })
```

### log.error(message [, context])

Log an error.

```lua
log.error("Unexpected state", { status = status })
```

### log.debug(message [, context])

Log a debug message. Typically only visible with debug-level logging enabled.

```lua
log.debug("Query result", { count = #results })
```

**Parameters (all four functions):**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `message` | string | Yes | Log message |
| `context` | table | No | Key-value pairs flattened into structured log arguments |

---

## require -- Module Loading

Loads Lua modules from the plugin's own `lib/` directory.

```lua
local validators = require("validators")
-- Resolves to: <plugin_dir>/lib/validators.lua
```

**Rules:**
- Only files under `<plugin_dir>/lib/` are loadable
- Path traversal (`..`, `/`, `\`) in module names is rejected
- Modules are cached after first load (subsequent `require` calls return the cached value)
- Modules run in the same sandboxed environment as `init.lua`
- By convention, modules return a table of functions

**Example module** (`lib/helpers.lua`):

```lua
local M = {}

function M.trim(s)
    if type(s) ~= "string" then return s end
    return s:match("^%s*(.-)%s*$")
end

function M.is_valid_url(url)
    if type(url) ~= "string" then return false end
    if url == "" then return false end
    return url:sub(1, 7) == "http://" or url:sub(1, 8) == "https://"
end

return M
```

---

## tui -- TUI Screen API

Available only to plugins that declare `screens` or `interfaces` in their manifest. The `tui` module provides constructors for building terminal UI primitives. All constructors are frozen (read-only) and produce plain Lua tables identical to hand-built equivalents.

Plugins that use TUI screens get the `coroutine` library enabled in their sandbox. Screen functions (`screens/<name>.lua`) define a `function screen(ctx)` entry point that runs as a coroutine, yielding layout tables and receiving event tables on resume.

### tui.grid(columns [, hints])

Creates a grid layout container. Used as the top-level yield value for screens and overlay field interfaces.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `columns` | table | Yes | Array of column tables (from `tui.column`) |
| `hints` | table | No | Array of `{key = "n", label = "new"}` for statusbar hints |

### tui.column(span, cells)

Creates a grid column.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `span` | int | Yes | Width units out of 12 |
| `cells` | table | Yes | Array of cell tables (from `tui.cell`) |

### tui.cell(title, height, content)

Creates a cell within a column.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | Yes | Panel title |
| `height` | number | Yes | Proportional height (0.0-1.0) |
| `content` | table | Yes | A primitive table (list, detail, text, etc.) |

### tui.list(items [, cursor])

Creates a vertical item list with cursor.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `items` | table | Yes | Array of `{label = "...", id = "...", faint = bool, bold = bool}` |
| `cursor` | int | No | Selected index (default 0) |

### tui.detail(fields)

Creates a key-value pair display.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `fields` | table | Yes | Array of `{label = "Name", value = "Test", faint = bool}` |

### tui.text(lines)

Creates a styled text block.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `lines` | table | Yes | Array of strings or `{text = "...", bold = bool, faint = bool, accent = bool}` |

### tui.table(headers, rows [, cursor])

Creates a table with headers and rows.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `headers` | table | Yes | Array of header strings |
| `rows` | table | Yes | Array of string arrays |
| `cursor` | int | No | Selected row index (default 0) |

### tui.input(id [, value [, placeholder]])

Creates a text input field.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Input identifier |
| `value` | string | No | Current value (default `""`) |
| `placeholder` | string | No | Placeholder text (default `""`) |

### tui.select_field(id, options [, selected])

Creates an option selector. Named `select_field` because `select` is a Lua reserved word.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Select identifier |
| `options` | table | Yes | Array of `{label = "All", value = ""}` |
| `selected` | int | No | Selected index (default 0) |

### tui.tree(nodes [, cursor])

Creates a hierarchical expandable tree.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `nodes` | table | Yes | Array of `{label = "...", id = "...", expanded = bool, children = {...}}` |
| `cursor` | int | No | Selected index (default 0) |

### tui.progress(value [, label])

Creates a progress bar.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `value` | number | Yes | Progress 0.0 to 1.0 |
| `label` | string | No | Label text (default `""`) |

### Coroutine Protocol

Screen entry points yield layout tables and receive event tables:

```lua
-- screens/main.lua
function screen(ctx)
    -- ctx: { protocol_version, width, height, params }
    while true do
        local event = coroutine.yield({
            type = "grid",
            columns = {
                tui.column(3, { tui.cell("List", 1.0, tui.list(items, cursor)) }),
                tui.column(9, { tui.cell("Detail", 1.0, tui.detail(fields)) }),
            },
            hints = { { key = "q", label = "quit" } },
        })
        if event.type == "key" and event.key == "q" then return end
    end
end
```

**Events (received on resume):**

| Event | Fields | Description |
|-------|--------|-------------|
| `init` | `protocol_version`, `width`, `height`, `params` | First event on startup |
| `key` | `key` (e.g., `"j"`, `"enter"`, `"ctrl+c"`) | Key press |
| `resize` | `width`, `height` | Terminal resized |
| `data` | `id`, `ok`, `result` or `error` | Async data response |
| `dialog` | `accepted` | Confirmation dialog response |

**Actions (yielded instead of layouts):**

| Action | Fields | Description |
|--------|--------|-------------|
| `navigate` | `plugin`, `screen`, `params` | Navigate to a plugin screen |
| `confirm` | `title`, `message` | Show confirmation dialog |
| `toast` | `message`, `level` | Show notification |
| `request` | `id`, `method`, `url`, `headers`, `body` | Async HTTP request |
| `commit` | `value` | Commit field value (interfaces only) |
| `cancel` | -- | Cancel without changing value (interfaces only) |
| `quit` | -- | Exit screen |

### Field Interfaces

Field interface entry points use the same protocol but are declared in `interfaces/<name>.lua` with `function interface(ctx)`. The `ctx` table includes `value` (current field value) and `config` (field data config). Inline interfaces yield single primitives; overlay interfaces yield grid layouts.

```lua
-- interfaces/swatch.lua
function interface(ctx)
    local color = ctx.value or "#000000"
    while true do
        local event = coroutine.yield({
            type = "text",
            lines = { { text = "██ " .. color, accent = true } },
        })
        if event.type == "key" and event.key == "enter" then
            coroutine.yield({ action = "commit", value = color })
            return
        end
    end
end
```

---

## Allowed Lua Standard Library

The plugin sandbox provides a restricted subset of Lua 5.1 standard library functions. Everything not listed here is unavailable.

### base

`type`, `tostring`, `tonumber`, `pairs`, `ipairs`, `next`, `select`, `unpack`, `error`, `pcall`, `xpcall`, `setmetatable`, `getmetatable`

### string

`string.find`, `string.sub`, `string.len`, `string.format`, `string.match`, `string.gmatch`, `string.gsub`, `string.rep`, `string.reverse`, `string.byte`, `string.char`, `string.lower`, `string.upper`

Also available via the string metatable: `s:find(...)`, `s:sub(...)`, etc.

### table

`table.insert`, `table.remove`, `table.sort`, `table.concat`

### math

`math.floor`, `math.ceil`, `math.max`, `math.min`, `math.abs`, `math.sqrt`, `math.huge`, `math.pi`, `math.random`, `math.randomseed`

### Removed (Sandboxed Out)

| Symbol | Reason |
|--------|--------|
| `io` | No filesystem access |
| `os` | No process/system access |
| `package` | No arbitrary module loading |
| `debug` | No VM introspection |
| `dofile` | No dynamic code loading from files |
| `loadfile` | No dynamic code loading from files |
| `load` | No dynamic code loading from strings |
| `rawget` | No metatable bypass (protects frozen modules) |
| `rawset` | No metatable bypass (protects frozen modules) |
| `rawequal` | No metatable bypass |
| `rawlen` | No metatable bypass |

---

## Operation Limits

| Limit | Default | Config Key | Description |
|-------|---------|------------|-------------|
| Operations per HTTP request handler | 1000 | `plugin_max_ops` | Exceeding raises Lua error |
| Operations per after-hook | 100 | `plugin_hook_max_ops` | Exceeding raises Lua error |
| Operations per before-hook | 0 (all blocked) | -- | `db.*` calls raise error in before-hooks |
| Execution timeout | 5s | `plugin_timeout` | VM execution killed after timeout |
| Max routes per plugin | 50 | `plugin_max_routes` | Exceeding during registration raises error |
| Max hooks per plugin | 50 | -- | Exceeding during registration raises error |
| Request body size | 1 MB | `plugin_max_request_body` | Larger bodies rejected with 413 |
| Response body size | 5 MB | `plugin_max_response_body` | Larger responses truncated |
| Rate limit | 100 req/s per IP | `plugin_rate_limit` | Exceeding returns 429 |
