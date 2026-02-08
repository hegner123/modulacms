# ModulaCMS Lua Plugin System -- Roadmap

## Skeptical Architect Review (2026-02-07)

### Initial Assessment: MEDIUM-HIGH Risk

The roadmap is well-organized and correctly leverages the existing `generic` package, but has architectural gaps that will cause production pain if not addressed before implementation.

### Critical Concerns (ordered by severity)

1. ~~**Connection pool starvation is a time bomb, not a "document later" item.** 4 VMs x 5 plugins x 5s timeout = 20 connections held. The app has 25 total. A slow plugin starves the core CMS. Needs a separate `*sql.DB` pool in Phase 1, not Phase 4.~~ **RESOLVED** -- separate `*sql.DB` pool for plugins.

2. ~~**The `generic` package uses `?` placeholders -- SQLite-only.** PostgreSQL needs `$1, $2, ...`. The roadmap doesn't acknowledge that the plugin DB API is implicitly SQLite-only. This must be scoped explicitly or fixed.~~ **RESOLVED** -- `generic` package now supports dialect-aware placeholders.

3. ~~**`CREATE TABLE` column types are SQLite-specific.** `BLOB` vs `BYTEA` (Postgres), `INTEGER` vs `INT` (MySQL). Fine for Phase 1 if scoped, but the type mapping should be designed now since changing allowed types later breaks plugin contracts.~~ **RESOLVED** -- `generic` package now provides 7 abstract column types (`text`, `integer`, `real`, `blob`, `boolean`, `timestamp`, `json`) with cross-dialect mapping via `SQLType()`, plus `CreateTable()` and `CreateIndex()` for dialect-aware DDL.

4. ~~**VM pool `Put()` doesn't validate VM health.** If a plugin does `db = nil`, that VM is permanently broken -- a silent 25% failure rate. `Put()` should verify critical globals are intact.~~ **RESOLVED** -- `Put()` validates Go-bound globals before returning VM to pool; broken VMs are closed and replaced via factory. See pool.go design below.

5. **`db.transaction()` has a deadlock risk.** Nothing prevents Lua code from calling top-level `db.query()` alongside `tx.insert()` inside the same callback. In SQLite without WAL mode, this deadlocks.

6. **Sandbox metatable escape is a security issue, not a testing concern.** `getmetatable`/`setmetatable` can replace the Go-bound `db` functions, breaking cross-plugin table isolation. Must be frozen in Phase 1.

7. **Not all mutations go through `audited` layer.** Methods like `UpdateUserSshKeyLastUsed` bypass audit. The Phase 3 hook injection claim that audited is "the single entry point" is inaccurate -- needs documenting as design decision.

8. **`GetConnection()` couples plugins to driver internals.** Reaching through the `DbDriver` abstraction to extract raw `*sql.DB` is a design smell worth acknowledging.

9. **No per-plugin operation rate limiting.** A plugin in a tight loop can fire thousands of queries within the 5s timeout. Needs a configurable ceiling (e.g., 1000 ops/window).

10. **Dependencies are name-only, no version constraints.** Plugin B upgrades to 2.0 with breaking changes, Plugin A breaks silently.

### Questions That Need Answers

1. What SQLite journal mode? WAL vs DELETE affects transaction behavior significantly.
2. How does Phase 2 wire routes into `NewModulacmsMux` without breaking its signature?
3. What's the recovery path for a failed plugin before Phase 4's circuit breaker?
4. Does plugin removal drop its tables or leave orphaned data?
5. Can duplicate plugin directories cause name collisions?

### What's Actually Good

- Building on the `generic` package is the correct decision -- it's well-built with proper `Executor` interface, identifier validation, and thorough tests.
- Hook injection at `audited` layer instead of wrapping 100+ `DbDriver` methods shows good maintenance cost judgment.
- The sandbox approach (gopher-lua `SkipOpenLibs` + selective loading) is standard and correct.
- Phase 1 scoped to core engine only (no HTTP, no hooks) limits blast radius well.
- The "Known Gaps" section is honest -- a sign of maturity.

### Verdict: Viable with conditions

Fix these 6 items for Phase 1 sign-off:

1. ~~**Connection pool isolation** -- separate `*sql.DB` for plugins (non-negotiable)~~ **RESOLVED**
2. ~~**Acknowledge PostgreSQL placeholder problem** -- scope as SQLite-only or fix `generic`~~ **RESOLVED**
3. ~~**VM health checking on `Put()`** -- verify `db`/`log` globals intact, discard broken VMs~~ **RESOLVED**
4. **Scope `db.*` inside transactions** -- replace `db` global with tx-bound executor in callbacks
5. **Freeze `db`/`log` module metatables** -- security concern, not a Phase 4 deferral
6. **Per-plugin operation counter** -- configurable ceiling to prevent DB saturation

> "This is a solid plan with real gaps. Fix the gaps I've identified for Phase 1 and I'd be comfortable seeing it go to implementation."

---

## Context

ModulaCMS is a headless CMS extended by Lua plugins at runtime. The current plugin package (`internal/plugin/plugin.go`) is a 30-line skeleton that creates a gopher-lua VM and maintains a flat list -- no loading, execution, sandboxing, or API exists. Extensive architecture docs (`ai/architecture/PLUGIN_ARCHITECTURE.md`, `ai/docs/PLUGIN_PACKAGE.md`, `ai/refactor/PROBLEM-UPDATE-2026-01-15-PLUGINS.md`) describe the vision: two-tier architecture where core CMS uses typed DbDriver methods and plugins use generic query builders against runtime-defined tables.

This roadmap designs the complete plugin system in 4 phases, starting with the core engine foundation.

---

## Plugin Directory Structure and Manifest

### Filesystem Layout
```
{plugins_directory}/           # Configured path, e.g., ./plugins/
  task_tracker/                # One directory per plugin
    init.lua                   # Required: entry point
    lib/                       # Optional: plugin-local modules
      helpers.lua
  analytics/
    init.lua
```

### Manifest (Lua table in init.lua)
```lua
plugin_info = {
    name        = "task_tracker",       -- Required: [a-z0-9_], max 32 chars
    version     = "1.0.0",             -- Required: semver
    description = "Task tracking",      -- Required
    author      = "Example Corp",       -- Optional
    dependencies = {},                  -- Optional: list of plugin names
}
```

Extracted by executing `init.lua` in a temporary sandboxed VM and reading the `plugin_info` global. Rejected with logged error if missing or invalid.

### Table Namespace Isolation
All plugin-created tables prefixed with `plugin_<plugin_name>_`. Go enforces the prefix -- Lua passes relative names (e.g., `"tasks"` becomes `plugin_task_tracker_tasks`). Validated via existing `db.ValidTableName()` (used by `internal/db/generic/`).

---

## Security Model

### Sandbox
gopher-lua `SkipOpenLibs: true` + selective loading:
- **ALLOWED:** `base` (with `dofile`/`loadfile`/`load`/`rawget`/`rawset` stripped), `table`, `string`, `math`, `coroutine`
- **BLOCKED:** `io`, `os`, `package`, `debug`

### Resource Limits
- **Execution timeout:** `LState.SetContext(ctx)` with deadline. Default 5s per call.
- **VM memory:** `CallStackSize=256`, `RegistrySize=5120`
- **Query limits:** Max 100 rows per `query()`, max 10 ops per transaction

---

## Phase 1: Core Engine (Foundation)

**Goal:** Plugin loading, Lua VM pool, sandboxed database API, schema definition, lifecycle management. No HTTP routes, no hooks.

### New Files

#### `internal/plugin/manager.go` -- L
Central coordinator: discovery, loading, lifecycle, shutdown.

```go
type PluginState string // discovered, loading, running, failed, stopped

type PluginInfo struct {
    Name, Version, Description, Author, License string
    MinCMSVersion string
    Dependencies  []string
}

type PluginInstance struct {
    Info     PluginInfo
    State    PluginState
    Dir      string         // absolute path to plugin directory
    InitPath string         // absolute path to init.lua
    Error    error
    Pool     *VMPool
}

type ManagerConfig struct {
    Enabled         bool
    Directory       string
    MaxVMsPerPlugin int  // default 4
    ExecTimeoutSec  int  // default 5
}

type Manager struct {
    cfg     ManagerConfig
    db      *sql.DB
    plugins map[string]*PluginInstance
    mu      sync.RWMutex
    ctx     context.Context
    cancel  context.CancelFunc
}
```

Key methods: `NewManager()`, `LoadAll(ctx)`, `Shutdown(ctx)`, `GetPlugin(name)`, `ListPlugins()`

**Loading sequence per plugin:**
1. Scan `plugins_directory` for subdirectories with `init.lua`
2. Create temp sandboxed VM, execute `init.lua`, extract/validate `plugin_info`
3. Validate dependencies (listed plugins must exist)
4. Initialize dedicated `VMPool`
5. Call `on_init()` if defined
6. Set state to `Running`

**Shutdown sequence:** Call `on_shutdown()` per plugin, close all VM pools, set state to `Stopped`.

#### `internal/plugin/pool.go` -- M
Goroutine-safe pool of pre-initialized `lua.LState` instances per plugin, with health validation on return.

```go
type VMPool struct {
    states     chan *lua.LState
    factory    func() *lua.LState
    size       int
    initPath   string
    pluginName string  // for diagnostic logging
}
```

Key methods: `NewVMPool(size, initPath, pluginName, sandbox, dbAPI)`, `Get(ctx)`, `Put(L)`, `Close()`

- Channel-based pool (buffered channel of `*lua.LState`)
- `Get()` applies `L.SetContext(ctx)` before returning (timeout enforcement)
- `Put()` validates VM health before returning to pool (see below)

**VM Health Check on `Put()`:**

`Put(L)` performs these steps in order:
1. `L.SetTop(0)` -- clear call stack (existing behavior)
2. `validateVM(L)` -- verify critical Go-bound globals are intact
3. If healthy: return VM to the channel
4. If unhealthy: `L.Close()`, log warning with plugin name and which global was corrupted, create fresh VM via `factory()`, put replacement in channel

`validateVM(L)` checks:
- `L.GetGlobal("db")` is `*lua.LTable` (not `LNil`, `LString`, etc.)
- `L.GetField(dbTable, "query")` is `*lua.LFunction` with `.IsG == true`
- Same check for `"insert"`, `"update"`, `"delete"`, `"transaction"`, `"define_table"`
- `L.GetGlobal("log")` is `*lua.LTable`
- `L.GetField(logTable, "info")` is `*lua.LFunction` with `.IsG == true`

**Design decisions:**
- **Check `IsG == true`, not pointer equality.** `LFunction.IsG` is `true` only for Go-bound functions. Catches replacement with pure Lua functions or non-function values. Does not require storing original pointers.
- **Replace, don't shrink.** Broken VMs are closed and replaced via `factory()`. Pool size stays constant. No degradation under corruption.
- **Lightweight.** 7 `GetGlobal`/`GetField` calls + type assertions. No Lua execution, no allocations. Nanosecond-scale overhead per `Put()`.
- **Log, don't panic.** Corrupted VM means plugin code is misbehaving. Log a warning with specifics (plugin name, which field failed) for admin investigation. Never crash the server.
- **Defense in depth with metatable freezing.** Health check is the fallback -- even if a metatable attack bypasses sandbox freezing (concern #6), the VM gets discarded on return.

#### `internal/plugin/sandbox.go` -- M
Configures Lua VM with safe stdlib subset.

```go
type SandboxConfig struct {
    AllowCoroutine bool
    ExecTimeout    time.Duration
}

func ApplySandbox(L *lua.LState, cfg SandboxConfig)
```

Loads `base`, `table`, `string`, `math` individually. Strips `dofile`, `loadfile`, `load`, `rawget`, `rawset`, `rawequal`, `rawlen` from globals.

#### `internal/plugin/db_api.go` -- L (downgraded from XL)
Sandboxed database API exposed as `db` Lua module. Delegates to existing `internal/db/generic/` package which already provides composable SELECT (WHERE maps, ORDER BY, LIMIT, OFFSET), INSERT, UPDATE, DELETE, Count, and Exists -- all with `db.ValidTableName()`/`db.ValidColumnName()` validation and parameterized values.

**Lua API:**
```lua
-- Structured query (auto-prefixes table name)
local rows = db.query("tasks", {where = {status = "active"}, order_by = "created_at", limit = 50})
local row  = db.query_one("tasks", {where = {id = "abc"}})

-- Mutations
db.insert("tasks", {id = db.ulid(), title = "My task", status = "active"})
db.update("tasks", {set = {status = "done"}, where = {id = "abc"}})
db.delete("tasks", {where = {id = "abc"}})

-- Transactions (uses L.PCall for error recovery; Lua error() triggers rollback)
local ok, err = db.transaction(function(tx)
    tx.insert("tasks", {id = db.ulid(), title = "Task 1"})
    tx.insert("tasks", {id = db.ulid(), title = "Task 2"})
end)

-- Helpers
local id = db.ulid()        -- generate ULID
local now = db.timestamp()   -- RFC3339 UTC timestamp (replaces os.date which is sandboxed)
```

```go
type DatabaseAPI struct {
    conn       *sql.DB
    pluginName string
    maxRows    int  // default 100, hard cap 10000 (enforced by generic.Select)
    maxTxOps   int  // default 10
}
```

**Mapping from Lua API to `generic` package:**

| Lua call | Go call |
|----------|---------|
| `db.query(table, opts)` | `generic.Select(ctx, exec, SelectParams{Table: prefixed, Where: opts.where, OrderBy: opts.order_by, Limit: opts.limit})` |
| `db.query_one(table, opts)` | `generic.SelectOne(ctx, exec, SelectParams{...})` |
| `db.insert(table, values)` | `generic.Insert(ctx, exec, InsertParams{Table: prefixed, Values: values})` |
| `db.update(table, opts)` | `generic.Update(ctx, exec, UpdateParams{Table: prefixed, Set: opts.set, Where: opts.where})` |
| `db.delete(table, opts)` | `generic.Delete(ctx, exec, DeleteParams{Table: prefixed, Where: opts.where})` |
| `db.transaction(fn)` | `sql.DB.BeginTx()` + pass `*sql.Tx` as `generic.Executor` + `L.PCall` for error recovery |

The `generic.Executor` interface accepts both `*sql.DB` and `*sql.Tx`, so transaction support is built-in. Table names auto-prefixed with `plugin_<name>_`. The `generic` package handles all identifier validation and parameterization.

**Row marshaling (generic.Row to Lua table):** `string` -> `LString`, `int64`/`float64` -> `LNumber`, `bool` -> `LBool`, `nil` -> `LNil`, `[]byte` -> `LString`.

#### `internal/plugin/schema_api.go` -- L
Lua-defined schema creation, validated by Go. Wraps `generic.CreateTable()` and `generic.CreateIndex()` for cross-dialect DDL with plugin-specific DX conveniences.

**7 Abstract Column Types** (via `generic.ValidColumnTypes`):

| Lua Type      | SQLite    | MySQL            | PostgreSQL         |
|---------------|-----------|------------------|--------------------|
| `"text"`      | `TEXT`    | `TEXT`           | `TEXT`             |
| `"integer"`   | `INTEGER` | `INT`            | `INTEGER`          |
| `"real"`      | `REAL`    | `DOUBLE`         | `DOUBLE PRECISION` |
| `"blob"`      | `BLOB`    | `BLOB`           | `BYTEA`            |
| `"boolean"`   | `INTEGER` | `TINYINT(1)`     | `BOOLEAN`          |
| `"timestamp"` | `TEXT`    | `TIMESTAMP`      | `TIMESTAMP`        |
| `"json"`      | `TEXT`    | `JSON`           | `JSONB`            |

**DX Helpers** (auto-injected by schema_api.go, NOT written by plugin developers):
- **Auto-ID column:** `id TEXT NOT NULL PRIMARY KEY` injected as first column. Plugins use `db.ulid()` to generate values.
- **Auto-timestamps:** `created_at` and `updated_at` columns injected after user columns. `db.insert()` auto-sets both; `db.update()` auto-sets `updated_at`.
- **Table prefixing:** `"tasks"` becomes `plugin_<name>_tasks` automatically.
- **FK namespace validation:** ForeignKey `ref_table` must start with same `plugin_<name>_` prefix (enforced by schema_api.go, not generic package).

**Lua API:**
```lua
-- Plugin developer writes this (no id, no timestamps, no prefix):
db.define_table("tasks", {
    columns = {
        {name = "title",    type = "text",    not_null = true},
        {name = "status",   type = "text",    not_null = true, default = "pending"},
        {name = "priority", type = "integer", not_null = true, default = 0},
        {name = "body",     type = "json"},
    },
    indexes = {
        {columns = {"status"}},
        {columns = {"status", "priority"}},
    },
})

-- What actually gets created (dialect-aware via generic.CreateTable):
-- CREATE TABLE IF NOT EXISTS "plugin_task_tracker_tasks" (
--     "id" TEXT NOT NULL PRIMARY KEY,
--     "title" TEXT NOT NULL,
--     "status" TEXT NOT NULL DEFAULT 'pending',
--     "priority" INTEGER NOT NULL DEFAULT '0',
--     "body" TEXT,
--     "created_at" TEXT NOT NULL,
--     "updated_at" TEXT NOT NULL
-- )
-- + CREATE INDEX IF NOT EXISTS "idx_plugin_task_tracker_tasks_status" ...
-- + CREATE INDEX IF NOT EXISTS "idx_plugin_task_tracker_tasks_status_priority" ...
```

**Lua Runtime Helpers** (provided by db_api.go):
```lua
local id = db.ulid()        -- "01HZQK..." (26-char ULID string)
local now = db.timestamp()  -- "2026-02-07T14:30:00Z" (RFC3339 UTC)

-- insert auto-sets id, created_at, updated_at:
db.insert("tasks", {title = "My task", status = "active"})

-- update auto-sets updated_at:
db.update("tasks", {set = {status = "done"}, where = {id = "01HZQK..."}})
```

```go
type TableDefinition struct {
    PluginName string
    TableName  string            // without prefix
    FullName   string            // plugin_<name>_<table>
    Columns    []generic.CreateColumnDef
    Indexes    []generic.IndexDef
    ForeignKeys []generic.ForeignKeyDef
}
```

**Implementation:** schema_api.go translates the Lua table definition into `generic.CreateTableParams`, auto-injects id/timestamp columns, prefixes the table name, validates FK namespace isolation, then calls `generic.CreateTable()`. All type mapping and DDL generation is handled by the generic package's cross-dialect infrastructure.

**Validation:** Type in `generic.ValidColumnTypes` (7 types), names pass `generic.ValidColumnName()`, max 64 columns, exactly one PK (auto-injected). Uses `CREATE TABLE IF NOT EXISTS` (idempotent). Cross-dialect support is built-in via `generic.SQLType()` — no SQLite-only limitation.

#### `internal/plugin/lua_helpers.go` -- S
Go<->Lua value conversion utilities.

```go
func LuaTableToMap(L *lua.LState, tbl *lua.LTable) map[string]any
func MapToLuaTable(L *lua.LState, m map[string]any) *lua.LTable
func GoValueToLua(L *lua.LState, v any) lua.LValue
func LuaValueToGo(v lua.LValue) any
func SQLRowsToLuaTable(L *lua.LState, rows *sql.Rows) (*lua.LTable, error)
```

#### `internal/plugin/log_api.go` -- S
Scoped logging exposed as `log` Lua module.

```lua
log.info("Processing started", {count = 42})
log.warn("Rate limit approaching")
log.error("Failed to process", {err = tostring(e)})
log.debug("Trace info")
```

Delegates to `utility.DefaultLogger` with plugin name as field.

### Existing Files to Modify

#### `internal/plugin/plugin.go` -- REPLACE (S)
Current 30-line skeleton replaced entirely. May become thin re-export or deleted if `manager.go` serves as entry point.

#### `internal/config/config.go` -- ADD fields (S)
```go
Plugins_Enabled    bool   `json:"plugins_enabled"`
Plugins_Directory  string `json:"plugins_directory"`
Plugins_Max_VMs    int    `json:"plugins_max_vms"`     // per plugin, default 4
Plugins_Timeout    int    `json:"plugins_timeout"`      // seconds, default 5
```

#### `cmd/serve.go` -- ADD init (S)
Insert between observability and install check:
```go
pluginCleanup := initPlugins(rootCtx, cfg, driver)
defer pluginCleanup()
```

#### `cmd/helpers.go` -- ADD helper (S)
```go
func initPlugins(ctx context.Context, cfg *config.Config, driver db.DbDriver) func() {
    // If !cfg.Plugins_Enabled, return noop
    // Get *sql.DB from driver.GetConnection()
    // Create Manager, call LoadAll, return Shutdown closure
}
```

### Testing Strategy

| Test File | Coverage |
|-----------|----------|
| `manager_test.go` | Discovery from temp dirs, lifecycle transitions, dependency validation |
| `pool_test.go` | Concurrent Get/Put, context cancellation, pool exhaustion (blocks, no crash) |
| `sandbox_test.go` | Blocked globals are nil, allowed globals work, timeout enforcement |
| `db_api_test.go` | All CRUD on in-memory SQLite, table prefixing, injection prevention, type marshaling |
| `schema_api_test.go` | Valid creation, validation failures, idempotency |
| `lua_helpers_test.go` | Round-trip: Go map -> Lua table -> Go map |
| `plugin_integration_test.go` | Load test plugin from `testdata/`, define table, insert, query, update, delete, transaction |

**Test fixtures:**
```
internal/plugin/testdata/plugins/
  valid_plugin/init.lua         -- Valid manifest + on_init creates table
  invalid_no_manifest/init.lua  -- Missing plugin_info
  invalid_bad_name/init.lua     -- plugin_info.name = "has spaces"
  timeout_plugin/init.lua       -- on_init has infinite loop
```

### Critical Existing Files Referenced
- `internal/db/generic/generic.go` -- **Primary foundation for plugin DB API.** Provides `Select`, `SelectOne`, `Insert`, `Update`, `Delete`, `Count`, `Exists` with composable WHERE/ORDER/LIMIT, `Executor` interface for `*sql.DB`/`*sql.Tx` interop, full identifier validation, parameterized values, 10000-row hard cap. Also provides cross-dialect DDL: 7 abstract `ColumnType` values with `SQLType()` mapping, `CreateTable()` with indexes/FKs, and standalone `CreateIndex()`. Fully tested.
- `internal/db/audited/audited.go` -- Transaction pattern reference; also the natural hook injection point for Phase 3 (preferred over wrapping full DbDriver interface)
- `internal/db/init.go` -- Singleton DB pattern. Note: `GetConnection()` returns a stored `context.Context` -- plugin manager should discard it and use its own `rootCtx`.
- `cmd/serve.go` -- Startup integration point. Plugin init must happen before `router.NewModulacmsMux()` (line 123) for Phase 2 route registration to work.

---

## Phase 2: HTTP Integration

**Goal:** Plugins register HTTP endpoints served via existing `net/http.ServeMux`. All plugin routes namespaced under `/api/v1/plugins/<plugin_name>/`.

### New Files

#### `internal/plugin/http_api.go` -- XL
Route registration and request/response marshaling.

**Lua API:**
```lua
http.handle("GET", "/tasks", function(req)
    local tasks = db.query("tasks", {order_by = "created_at", limit = 50})
    return {status = 200, json = tasks}
end)

http.handle("POST", "/tasks", function(req)
    local body = req.json
    db.insert("tasks", {id = db.ulid(), title = body.title, status = "pending"})
    return {status = 201, json = {message = "created"}}
end)
```

```go
type LuaRequest struct {
    Method, Path, Body, RemoteAddr, UserID string
    Headers, Query, PathParams map[string]string
    JSON any
}

type LuaResponse struct {
    Status  int
    Headers map[string]string
    Body    string
    JSON    any
}

type HTTPBridge struct {
    manager *Manager
    mux     *http.ServeMux
    routes  []RouteRegistration
}
```

**Request flow:** HTTP -> ServeMux -> Bridge extracts plugin name -> checks out VM from pool -> converts Request to LuaRequest -> calls Lua handler -> converts LuaResponse -> writes HTTP response -> returns VM

#### `internal/plugin/http_middleware.go` -- M
Plugin-scoped middleware.

```lua
http.use(function(req)
    if not req.headers["X-API-Key"] then
        return {status = 401, json = {error = "API key required"}}
    end
    return nil  -- continue to handler
end)
```

### Existing Files to Modify
- `internal/router/mux.go` -- Accept optional `*plugin.HTTPBridge`, mount catch-all handler (M)
- `cmd/serve.go` -- Wire bridge to mux constructor (S)

---

## Phase 3: Content Hooks

**Goal:** Plugins react to content lifecycle events. Before-hooks run inside the mutation transaction (can abort). After-hooks run post-commit (fire-and-forget).

### New Files

#### `internal/plugin/hooks.go` -- XL
Hook registration, ordering, and execution engine.

**Lua API:**
```lua
hooks.on("before_create", "content_data", function(data)
    if not data.title then error("title required") end
    return data
end)

hooks.on("after_publish", "content_data", function(data)
    log.info("Published", {id = data.content_data_id})
end)
```

**Events:** `before_create`, `after_create`, `before_update`, `after_update`, `before_delete`, `after_delete`, `before_publish`, `after_publish`

**Execution semantics:**
- **Before hooks:** Inside transaction, sequential by priority, each receives previous output. `error()` rolls back.
- **After hooks:** Post-commit, concurrent per plugin, errors logged but don't affect committed data.

#### Hook injection approach -- inject at `audited` layer, NOT via DbDriver wrapper

The `DbDriver` interface has 100+ methods. Wrapping it is a massive maintenance surface. Instead, inject hooks at the `audited.Create`/`audited.Update`/`audited.Delete` generic functions, which are already the single entry point for all mutations with transactions. Add an optional `HookRunner` parameter (or embed in `AuditContext`) that the audited functions call at the appropriate points within the transaction.

```go
// In audited package, extend the command interfaces or AuditContext:
type HookRunner interface {
    RunBefore(ctx context.Context, event HookEvent, table string, data map[string]any) (map[string]any, error)
    RunAfter(ctx context.Context, event HookEvent, table string, data map[string]any)
}
```

This keeps the `db` package clean, avoids wrapping 100+ methods, and hooks execute inside the existing transaction boundary naturally.

### Existing Files to Modify
- `internal/db/audited/audited.go` -- Add optional `HookRunner` to Create/Update/Delete flows (M)
- `cmd/serve.go` -- Pass `HookRunner` from plugin manager into audit context (S)

---

## Phase 4: Production Hardening

**Goal:** Operational robustness -- hot reload, monitoring, error isolation, CLI management.

### New Files

| File | Size | Purpose |
|------|------|---------|
| `internal/plugin/watcher.go` | L | File polling for hot reload (detect changes, graceful restart per plugin) |
| `internal/plugin/metrics.go` | M | Per-plugin metrics: VM usage, call counts, error rates, avg exec time |
| `internal/plugin/recovery.go` | M | `SafeExecute()` with panic recovery + `CircuitBreaker` (5 failures -> temp disable) |
| `internal/plugin/cli_commands.go` | M | TUI/CLI commands: `plugin list/enable/disable/reload/info` |

### Config Additions
```go
Plugins_Hot_Reload     bool   `json:"plugins_hot_reload"`
Plugins_Max_Failures   int    `json:"plugins_max_failures"`    // circuit breaker, default 5
Plugins_Reset_Interval string `json:"plugins_reset_interval"`  // default "60s"
```

---

## Known Gaps and Reviewer Findings

Items identified during go-backend-reviewer review that must be addressed during implementation:

1. **Dependency cycle detection** -- `LoadAll` must topologically sort plugins and detect circular dependencies. Not just "listed plugins must exist."
2. **Schema evolution** -- `CREATE TABLE IF NOT EXISTS` is idempotent for creation, but adding columns in plugin v1.1.0 silently uses old schema. Migration story deferred to Phase 4 (document as known limitation in Phase 1).
3. **DB connection isolation** -- Plugins share the app's `*sql.DB` pool (25 connections). A plugin holding connections for 5s across N plugins can starve CMS queries. Phase 1 should document this; Phase 4 should add an optional separate connection pool for plugins.
4. **VM global state leakage** -- `Put()` with `L.SetTop(0)` preserves globals between handler invocations. Plugin authors must use `local` exclusively. Document this requirement. Phase 4 should evaluate resetting globals to post-init snapshot.
5. **Transaction error semantics** -- `db.transaction(fn)` must use `L.PCall` with error recovery. Lua `error()` inside the callback triggers rollback. This must be explicitly implemented, not left to default gopher-lua error handling.
6. **Plugin name collision** -- Names like `a_` and `a` produce `plugin_a__tasks` vs `plugin_a_tasks`. Validation must reject trailing underscores or enforce separator uniqueness.
7. **Sandbox metatable escape** -- Even with `rawget`/`rawset` stripped, `getmetatable`/`setmetatable` on the `db`/`log` modules could replace Go-bound functions. Testing must cover this; consider freezing module metatables.

---

## Performance Notes

- **VM Pool:** Default 4 VMs per plugin. Increase for HTTP-heavy plugins, 2 sufficient for hook-only.
- **Boundary crossing:** Structured query API minimizes crossings -- one table in, one table out per operation.
- **Connection sharing:** Plugins share app's `*sql.DB` pool. Consider separate pool if plugin load is high.
- **Startup:** Plugins loaded sequentially (dependency order), VM pool init parallelized per plugin.

---

## Example Plugin: Task Tracker

```lua
plugin_info = {
    name        = "task_tracker",
    version     = "1.0.0",
    description = "Simple task tracking for content workflows",
}

-- id, created_at, updated_at are auto-injected by schema_api.go
db.define_table("tasks", {
    columns = {
        {name = "title",       type = "text",    not_null = true},
        {name = "description", type = "text"},
        {name = "status",      type = "text",    not_null = true, default = "pending"},
        {name = "priority",    type = "integer", not_null = true, default = 0},
        {name = "content_id",  type = "text"},
    },
    indexes = {
        {columns = {"status"}},
        {columns = {"status", "priority"}},
    },
})

function on_init()
    log.info("Task tracker initialized")
    local existing = db.query("tasks", {limit = 1})
    if #existing == 0 then
        -- id, created_at, updated_at are auto-set by db.insert()
        db.insert("tasks", {
            title = "Review plugin system",
            status = "pending", priority = 1,
        })
    end
end

function on_shutdown()
    log.info("Task tracker shutting down")
end
```

---

## Multi-Agent Work Decomposition (Phase 1)

**Parallel group 1 (no dependencies):**
- Agent A: `sandbox.go` + tests
- Agent B: `lua_helpers.go` + tests
- Agent C: `log_api.go` + tests
- Agent D: Config changes

**Parallel group 2 (depends on group 1):**
- Agent E: `pool.go` + tests (needs sandbox)
- Agent F: `schema_api.go` + tests (needs lua_helpers)

**Parallel group 3 (depends on group 2):**
- Agent G: `db_api.go` + tests (needs pool, lua_helpers, schema_api)

**Sequential (depends on all):**
- Agent H: `manager.go` + tests
- Agent I: `cmd/serve.go` + `cmd/helpers.go` integration
- Agent J: Integration tests

---

## Verification

After Phase 1 implementation:
1. `just test` passes (no regressions)
2. `go test ./internal/plugin/...` passes all unit + integration tests
3. Place example plugin in `./plugins/task_tracker/init.lua`
4. Start with `./modulacms-x86 serve` and `plugins_enabled: true` in config
5. Verify logs show "Task tracker initialized" and table creation
6. Connect to SQLite and confirm `plugin_task_tracker_tasks` table exists with seed data
