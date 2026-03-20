# Lua Plugin API Reference

Complete reference for all APIs available to Lua plugins in ModulaCMS.

## Lua Environment

### Allowed Standard Library

| Module | Available Functions |
|--------|-------------------|
| `base` | type, tostring, tonumber, pairs, ipairs, next, select, unpack, error, pcall, xpcall, setmetatable, getmetatable |
| `string` | find, sub, len, format, match, gmatch, gsub, rep, reverse, byte, char, lower, upper |
| `table` | insert, remove, sort, concat |
| `math` | floor, ceil, max, min, abs, sqrt, huge, pi, random, randomseed |

### Removed (Sandboxed Out)

- `io`, `os`, `package`, `debug` -- no filesystem, process, or introspection access
- `dofile`, `loadfile`, `load` -- no dynamic code loading
- `rawget`, `rawset`, `rawequal`, `rawlen` -- no metatable bypass

All injected modules (`db`, `http`, `hooks`, `log`) are frozen read-only via metatable proxy.

---

## db -- Database API

Source: `internal/plugin/db_api.go`

All table names are auto-prefixed with `plugin_<name>_`. Lua code uses short names only (e.g., `"tasks"` becomes `plugin_task_tracker_tasks` in SQL).

### db.define_table(table, definition)

Creates a plugin table (IF NOT EXISTS). Call in `on_init()`.

Three columns are auto-injected -- do NOT include them in your columns list:
- `id` (TEXT PRIMARY KEY, ULID)
- `created_at` (TEXT, RFC3339 UTC)
- `updated_at` (TEXT, RFC3339 UTC)

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
            ref_table  = "plugin_task_tracker_categories", -- must use same plugin prefix
            ref_column = "id",
            on_delete  = "CASCADE",
        },
    },
})
```

**Column types**: `text`, `integer`, `boolean`, `real`, `timestamp`, `json`, `blob`

**Raises error** if: reserved column name used, empty columns list, invalid type, foreign key references another plugin's table.

### db.insert(table, values)

Inserts a row. Auto-sets `id`, `created_at`, `updated_at` if not provided. Explicit values are never overridden.

```lua
db.insert("tasks", {
    id          = db.ulid(),           -- optional, auto-generated if omitted
    title       = "Fix bug",
    status      = "todo",
    created_at  = db.timestamp(),      -- optional, auto-set if omitted
})
```

Returns nothing on success. On error: `nil, errmsg`.

### db.query(table, opts) -> table

Returns a sequence table of row tables. Returns empty table `{}` on no matches (never nil).

```lua
local tasks = db.query("tasks", {
    where    = { status = "todo", category_id = "01ABC..." },
    order_by = "created_at",
    desc     = true,
    limit    = 50,
    offset   = 0,
})

for _, task in ipairs(tasks) do
    log.info("Task: " .. task.title)
end
```

**opts fields**:
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `where` | table | nil | Column=value equality filters (AND). Values can be condition constructors: `db.gt(v)`, `db.gte(v)`, `db.lt(v)`, `db.lte(v)`, `db.like(v)`, `db.not_like(v)`, `db.in_list(t)`, `db.not_in(t)`, `db.between(a,b)`, `db.is_null()`, `db.is_not_null()` |
| `order_by` | string | nil | Column name to sort by (must be a valid identifier, no spaces) |
| `desc` | boolean | false | Sort descending when true (used with order_by) |
| `limit` | number | 100 | Max rows returned |
| `offset` | number | 0 | Skip N rows |

Omitting `where` returns all rows (up to limit).

### db.query_one(table, opts) -> table or nil

Returns a single row table, or `nil` if no match.

```lua
local task = db.query_one("tasks", { where = { id = task_id } })
if not task then
    return { status = 404, json = { error = "not found" } }
end
```

**opts fields**: `where` (table), `order_by` (string).

### db.count(table, opts) -> number

Returns row count as integer.

```lua
local total = db.count("tasks", {})                              -- all rows
local done  = db.count("tasks", { where = { status = "done" } }) -- filtered
```

### db.exists(table, opts) -> boolean

Returns `true` if at least one row matches.

```lua
if not db.exists("tasks", { where = { id = id } }) then
    return { status = 404, json = { error = "not found" } }
end
```

### db.update(table, opts)

Updates rows matching `where`. Both `set` and `where` are required and must be non-empty (safety: prevents full-table updates). Auto-sets `updated_at` if not in `set`.

```lua
db.update("tasks", {
    set   = { status = "done", title = "Fixed bug" },
    where = { id = task_id },
})
```

### db.delete(table, opts)

Deletes rows matching `where`. `where` is required and must be non-empty (safety: prevents full-table deletes).

```lua
db.delete("tasks", { where = { id = task_id } })
```

### db.transaction(fn) -> boolean, string|nil

Wraps multiple operations in a single database transaction. Nested transactions are rejected.

```lua
local ok, err = db.transaction(function()
    db.insert("categories", { name = "Bug" })
    db.insert("categories", { name = "Feature" })
    -- If any call errors, the entire transaction rolls back.
end)

if not ok then
    log.warn("Transaction failed", { err = err })
end
```

Returns `true, nil` on commit. Returns `false, errmsg` on rollback.

### db.ulid() -> string

Generates a 26-character ULID (time-sortable, globally unique).

```lua
local id = db.ulid()  -- e.g., "01HXYZ..."
```

### db.timestamp() -> string

Returns current UTC time as RFC3339 string. Replaces `os.date` which is sandboxed.

```lua
local now = db.timestamp()  -- e.g., "2026-02-15T12:00:00Z"
```

### db.timestamp_ago(seconds) -> string

Returns (now - N seconds) as an RFC3339 UTC string. Used for time window comparisons
where `os.time()` is sandboxed out. RFC3339 with zero-padded UTC fields sorts
lexicographically, so string comparison operators work correctly.

```lua
local one_hour_ago = db.timestamp_ago(3600)
if record.created_at < one_hour_ago then
    -- record is older than 1 hour
end
```

---

## http -- HTTP Route API

Source: `internal/plugin/http_api.go`

### http.handle(method, path, handler [, options])

Registers an HTTP route. **Must be called at module scope** (top-level code), NOT inside `on_init()`.

```lua
http.handle("GET", "/tasks/{id}", function(req)
    local task = db.query_one("tasks", { where = { id = req.params.id } })
    if not task then
        return { status = 404, json = { error = "not found" } }
    end
    return { status = 200, json = task }
end, { public = true })
```

**Arguments**:
| Arg | Type | Description |
|-----|------|-------------|
| method | string | `GET`, `POST`, `PUT`, `DELETE`, `PATCH` |
| path | string | Starts with `/`, max 256 chars. Supports `{param}` path parameters |
| handler | function | Receives request table, returns response table |
| options | table | Optional. `{ public = true }` bypasses CMS auth (default: authenticated) |

**Full URL**: `/api/v1/plugins/<plugin_name><path>`

Routes require admin approval before serving traffic. Unapproved routes return 404.

### Request Table

| Field | Type | Description |
|-------|------|-------------|
| `req.method` | string | HTTP method (e.g., `"GET"`) |
| `req.path` | string | Full URL path |
| `req.body` | string | Raw request body |
| `req.client_ip` | string | Client IP (proxy-aware, no port) |
| `req.headers` | table | All headers (lowercase keys) |
| `req.query` | table | URL query parameters (`?name=value`) |
| `req.params` | table | Path parameters from `{param}` wildcards |
| `req.json` | table | Parsed JSON body (only when Content-Type is application/json) |

### Response Table

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | number | 200 | HTTP status code |
| `json` | table | nil | Serialized as JSON (sets Content-Type automatically) |
| `body` | string | nil | Raw string body (used only if `json` is nil) |
| `headers` | table | nil | Custom response headers |

**Blocked response headers**: `access-control-*`, `set-cookie`, `transfer-encoding`, `content-length`, `cache-control`, `host`, `connection`.

### http.use(middleware_function)

Appends middleware executed before route handlers.

```lua
http.use(function(req)
    if not req.headers["x-api-key"] then
        return { status = 401, json = { error = "missing api key" } }
    end
    -- Return nil to continue to route handler.
    return nil
end)
```

---

## hooks -- Content Lifecycle Hooks

Source: `internal/plugin/hooks_api.go`, `internal/plugin/hook_engine.go`

### hooks.on(event, table, handler [, options])

Registers a content lifecycle hook. **Must be called at module scope**, NOT inside `on_init()`.

```lua
hooks.on("before_create", "content_data", function(data)
    if data.title and data.title == "" then
        error("title cannot be empty")  -- aborts the transaction
    end
end, { priority = 50 })
```

**Arguments**:
| Arg | Type | Description |
|-----|------|-------------|
| event | string | Hook event name (see table below) |
| table | string | CMS table name (e.g., `"content_data"`), or `"*"` for wildcard |
| handler | function | Receives data table with entity fields |
| options | table | Optional. `{ priority = <1-1000> }` (lower runs first, default 100) |

**Events**:
| Event | Timing | Can Abort? | Has db.* Access? |
|-------|--------|-----------|------------------|
| `before_create` | Inside transaction | Yes (call `error()`) | No |
| `after_create` | After commit | No | Yes |
| `before_update` | Inside transaction | Yes | No |
| `after_update` | After commit | No | Yes |
| `before_delete` | Inside transaction | Yes | No |
| `after_delete` | After commit | No | Yes |
| `before_publish` | Inside transaction | Yes | No |
| `after_publish` | After commit | No | Yes |
| `before_archive` | Inside transaction | Yes | No |
| `after_archive` | After commit | No | Yes |

**Handler data fields**:
| Field | Description |
|-------|-------------|
| `data._table` | The table name |
| `data._event` | The event name |
| `data.*` | All entity fields from the CMS table |

**Key constraints**:
- Before-hooks run synchronously inside the CMS transaction. `error()` aborts the transaction.
- After-hooks run asynchronously (fire-and-forget). Errors are logged, not propagated.
- `db.*` calls are blocked inside before-hooks (prevents SQLite deadlock).
- Hooks require admin approval before they fire.

**Hook circuit breaker**: After 10 consecutive errors (configurable), the hook is auto-disabled until the plugin is reloaded or re-enabled.

---

## log -- Structured Logging

Source: `internal/plugin/log_api.go`

### log.info(message [, context])
### log.warn(message [, context])
### log.error(message [, context])
### log.debug(message [, context])

```lua
log.info("Task created", { id = task_id, title = "Fix bug" })
log.warn("Category seeding failed", { err = err })
log.error("Unexpected state", { status = status })
log.debug("Query result", { count = #results })
```

Plugin name is automatically included in every log entry. Context table key-value pairs are flattened into structured log arguments.

---

## require -- Module Loading

Source: `internal/plugin/sandbox.go`

Loads modules from the plugin's own `lib/` directory.

```lua
local validators = require("validators")
-- Resolves to: <plugin_dir>/lib/validators.lua
```

**Rules**:
- Only files under `<plugin_dir>/lib/` are loadable
- Path traversal (`..`, `/`, `\`) is rejected
- Modules are cached after first load
- Modules have the same sandboxed environment as `init.lua`
- By convention, modules return a table of functions

```lua
-- lib/helpers.lua
local M = {}
function M.trim(s)
    if type(s) ~= "string" then return s end
    return s:match("^%s*(.-)%s*$")
end
return M
```

---

## Operation Limits

| Limit | Default | Config Key |
|-------|---------|------------|
| Operations per request | 1000 | `plugin_max_ops` |
| Operations per before-hook | 100 | `plugin_hook_max_ops` |
| Execution timeout | 5s | `plugin_timeout` |
| Max routes per plugin | 50 | `plugin_max_routes` |
| Max hooks per plugin | 50 | -- |
| Request body size | 1 MB | `plugin_max_request_body` |
| Response body size | 5 MB | `plugin_max_response_body` |
| Rate limit | 100 req/s per IP | `plugin_rate_limit` |

## TUI Module (`tui.*`)

The `tui` module provides constructors for building terminal UI primitives. Available only to plugins that declare `screens` or `interfaces` in their manifest. The module is frozen (read-only).

### Constructors

| Function | Arguments | Returns |
|----------|-----------|---------|
| `tui.grid(columns, hints?)` | columns: array of column tables; hints: optional array of `{key, label}` | `{type="grid", columns=..., hints=...}` |
| `tui.column(span, cells)` | span: int (out of 12); cells: array of cell tables | `{span=..., cells=...}` |
| `tui.cell(title, height, content)` | title: string; height: float; content: primitive table | `{title=..., height=..., content=...}` |
| `tui.list(items, cursor?)` | items: array of `{label, id, faint?, bold?}`; cursor: int (default 0) | `{type="list", items=..., cursor=...}` |
| `tui.detail(fields)` | fields: array of `{label, value, faint?}` | `{type="detail", fields=...}` |
| `tui.text(lines)` | lines: array of strings or `{text, bold?, faint?, accent?}` | `{type="text", lines=...}` |
| `tui.table(headers, rows, cursor?)` | headers: array of strings; rows: array of string arrays; cursor: int | `{type="table", ...}` |
| `tui.input(id, value?, placeholder?)` | id: string; value/placeholder: optional strings | `{type="input", ...}` |
| `tui.select_field(id, options, selected?)` | id: string; options: array of `{label, value}`; selected: int | `{type="select", ...}` |
| `tui.tree(nodes, cursor?)` | nodes: array of `{label, id, expanded?, children?}`; cursor: int | `{type="tree", ...}` |
| `tui.progress(value, label?)` | value: float 0.0-1.0; label: optional string | `{type="progress", ...}` |

All constructors produce plain Lua tables identical to hand-built equivalents. They are pure sugar with no side effects.

### Coroutine Protocol

Screen functions (`screens/<name>.lua`) define a `function screen(ctx)` that runs as a coroutine. The function yields layout tables and receives event tables on resume:

**Events (Go → Lua):**
- `{type="init", protocol_version=1, width=N, height=N, params={...}}`
- `{type="key", key="j"}` / `{type="key", key="enter"}` / `{type="key", key="ctrl+c"}`
- `{type="resize", width=N, height=N}`
- `{type="data", id="...", ok=bool, result=... or error="..."}`
- `{type="dialog", accepted=bool}`

**Actions (Lua → Go via yield):**
- `{action="navigate", plugin="name", screen="name", params={...}}`
- `{action="confirm", title="...", message="..."}`
- `{action="toast", message="...", level="success"}`
- `{action="request", id="...", method="GET", url="..."}`
- `{action="commit", value="..."}` (field interfaces only)
- `{action="cancel"}` (field interfaces only)
- `{action="quit"}` (screens only)
