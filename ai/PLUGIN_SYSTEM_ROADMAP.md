# ModulaCMS Lua Plugin System -- Roadmap

## Review History

This roadmap has passed five review rounds:

1. **Skeptical Architect Review (2026-02-07)** -- identified 10 critical concerns (connection pool starvation, dialect portability, metatable escape, transaction deadlocks, VM health checking, rate limiting, etc.). All architectural concerns resolved. Remaining open items carried to Known Gaps.

2. **Lua DX Review (2026-02-08)** -- identified 8 API design concerns (equality-only WHERE, missing count/exists, `#` trap, contract globals, `define_table` execution model, restricted require, error convention). All applied to roadmap.

3. **Final Pre-Implementation Review (2026-02-08)** -- verified infrastructure claims against codebase, caught 6 documentation accuracy issues (mapping table signatures, config naming, initPlugins helper, on_init semantics, empty WHERE docs, table cleanup policy). All corrected below.

4. **Critical Analysis Review (2026-02-08)** -- identified 14 issues: PluginState as raw string (not iota enum), error stored on long-lived struct, Manager holding stored context (ambiguity with parameter context), WithTransaction wrong package path, missing empty WHERE semantics for reads, pairs() undocumented on frozen modules, pool exhaustion with no backpressure, gopher-lua timeout limitation on Go-bound calls, WAL mode prerequisite undocumented, no VM global snapshot/restore, error wrapping with %s not %w, config naming inconsistency, VMPool.Close drain unsafety, thread safety caveat on DatabaseAPI. All corrected below.

5. **Internal Consistency Review (2026-02-08)** -- caught 8 issues introduced or exposed by round 4 edits: three conflicting VMPool struct definitions (consolidated to one), Get() return type change not propagated to loading/shutdown sequences, Get() referencing non-existent p.dbAPI field (ResetOpCount moved to caller), ManagerConfig missing MaxOpsPerExec field, NewManager() signature contradicting context paragraph, review count header stale, Phase 3 still claiming "single entry point" despite Known Gap #4, query_one with empty WHERE non-deterministic without order_by. All corrected below.

**Status: APPROVED FOR IMPLEMENTATION**

---

## Context

ModulaCMS is a headless CMS extended by Lua plugins at runtime. The current plugin package (`internal/plugin/plugin.go`) is a 30-line skeleton that creates a gopher-lua VM and maintains a flat list -- no loading, execution, sandboxing, or API exists. Extensive architecture docs (`ai/architecture/PLUGIN_ARCHITECTURE.md`, `ai/docs/PLUGIN_PACKAGE.md`, `ai/refactor/PROBLEM-UPDATE-2026-01-15-PLUGINS.md`) describe the vision: two-tier architecture where core CMS uses typed DbDriver methods and plugins use the `db.Q*` query builder functions against runtime-defined tables.

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

### Plugin Contract Globals

These globals are part of the plugin API contract. Everything else must be declared `local`.

| Global | Type | Required | When Read |
|--------|------|----------|-----------|
| `plugin_info` | table | Yes | During manifest extraction (temp VM) |
| `on_init` | function | No | Once per plugin via Manager, in dependency order |
| `on_shutdown` | function | No | Once per plugin via Manager, in reverse dependency order |

**`on_init` execution model:** The Manager checks out one VM from the pool, calls `on_init()`, and returns the VM. This happens **once per plugin**, not per VM. Other VMs in the pool do not run `on_init`. This means `db.define_table()` (which uses `CREATE TABLE IF NOT EXISTS`) runs exactly once per plugin load. `db.define_table()` must be called inside `on_init()`, not at module scope, because the temp VM used for manifest extraction has no `db` API and is discarded after reading `plugin_info`.

**`db.define_table()` is not enforced to `on_init()` only** -- calling it from an HTTP handler or hook callback would succeed (idempotent DDL). However, this is convention: schema should be defined at startup, not at request time. Phase 4 may add a flag to restrict `define_table` to the init lifecycle.

### Table Namespace Isolation
All plugin-created tables prefixed with `plugin_<plugin_name>_`. Go enforces the prefix -- Lua passes relative names (e.g., `"tasks"` becomes `plugin_task_tracker_tasks`). Validated via existing `db.ValidTableName()` (in `internal/db/query_builder.go`).

---

## Security Model

### Sandbox
gopher-lua `SkipOpenLibs: true` + selective loading:
- **ALLOWED:** `base` (with `dofile`/`loadfile`/`load`/`rawget`/`rawset` stripped), `table`, `string`, `math`, `coroutine`
- **BLOCKED:** `io`, `os`, `package`, `debug`
- **RESTRICTED:** `require` — replaced with a custom loader that resolves only within `<plugin_dir>/lib/`. Module names map to files: `require("helpers")` loads `<plugin_dir>/lib/helpers.lua`. Paths outside the plugin directory are rejected. This enables multi-file plugins without opening filesystem access.

### Resource Limits
- **Execution timeout:** `LState.SetContext(ctx)` with deadline. Default 5s per call.
- **VM memory:** `CallStackSize=256`, `RegistrySize=5120`
- **Query limits:** Max 100 rows per `query()`, max 10 ops per transaction
- **Operation budget:** Max 1000 `db.*` calls per VM checkout (configurable via `Plugin_Max_Ops`). Reset on `VMPool.Get()`. Prevents tight-loop DB saturation within the 5s timeout.

**Timeout limitation:** gopher-lua checks context cancellation at Lua instruction boundaries only. Go-bound functions (e.g., a slow SQL query executing inside `db.query()`) block until the Go call returns — the 5s deadline does not interrupt in-flight database operations. The timeout protects against infinite Lua loops and runaway Lua computation, but not against slow queries. Database-level query timeouts should be enforced separately via the `*sql.DB` connection's `SetConnMaxLifetime` or per-query context deadlines passed through to the query builder. The `db.*` API functions should propagate `L.Context()` to all `QSelect`/`QInsert`/etc. calls so database drivers respect the deadline where supported.

### Prerequisites
- **SQLite WAL mode:** Plugin transactions use executor swapping (see db_api.go). Without WAL mode, SQLite allows only one writer at a time and a read on a separate connection during a write will block or return SQLITE_BUSY. The plugin pool uses a separate `*sql.DB` from the core CMS pool, so concurrent access between core and plugin operations requires WAL mode. **WAL mode must be enabled on the SQLite database before plugins are loaded.** Verify with `PRAGMA journal_mode;` — it should return `wal`. If the CMS does not already set this, `initPluginPool()` or `Manager.LoadAll()` should execute `PRAGMA journal_mode=WAL;` on startup and verify the result.

---

## Phase 1: Core Engine (Foundation)

**Goal:** Plugin loading, Lua VM pool, sandboxed database API, schema definition, lifecycle management. No HTTP routes, no hooks.

### New Files

#### `internal/plugin/manager.go` -- L
Central coordinator: discovery, loading, lifecycle, shutdown.

```go
type PluginState int

const (
    StateDiscovered PluginState = iota
    StateLoading
    StateRunning
    StateFailed
    StateStopped
)

func (s PluginState) String() string {
    switch s {
    case StateDiscovered:
        return "discovered"
    case StateLoading:
        return "loading"
    case StateRunning:
        return "running"
    case StateFailed:
        return "failed"
    case StateStopped:
        return "stopped"
    default:
        return fmt.Sprintf("unknown(%d)", s)
    }
}

type PluginInfo struct {
    Name, Version, Description, Author, License string
    MinCMSVersion string
    Dependencies  []string
}

type PluginInstance struct {
    Info         PluginInfo
    State        PluginState
    Dir          string         // absolute path to plugin directory
    InitPath     string         // absolute path to init.lua
    FailedReason string         // human-readable failure message; empty when State != StateFailed
    Pool         *VMPool
}

type ManagerConfig struct {
    Enabled         bool
    Directory       string
    MaxVMsPerPlugin int  // default 4
    ExecTimeoutSec  int  // default 5
    MaxOpsPerExec   int  // default 1000, per VM checkout
}

type Manager struct {
    cfg     ManagerConfig
    db      *sql.DB         // separate plugin pool via db.OpenPool()
    dialect db.Dialect
    plugins map[string]*PluginInstance
    mu      sync.RWMutex
}
```

**Context handling:** `Manager` does not store a `context.Context` field. Stored contexts on long-lived structs create ambiguity about which context governs a given operation. Instead, all Manager methods that perform I/O accept `ctx context.Context` as their first parameter. The caller (cmd/serve.go) passes the application lifecycle context. This follows the Go convention: contexts flow through call chains, not struct fields.

Key methods: `NewManager(cfg, db, dialect)`, `LoadAll(ctx)`, `Shutdown(ctx)`, `GetPlugin(name)`, `ListPlugins()`

**Loading sequence per plugin:**
1. Scan `plugins_directory` for subdirectories with `init.lua`
2. Create temp sandboxed VM (no `db` API registered), execute `init.lua`, extract/validate `plugin_info`. Temp VM is discarded after this step — side effects do not persist.
3. Validate dependencies (listed plugins must exist)
4. Initialize dedicated `VMPool` (production VMs with full `db`/`log` APIs and sandboxed `require`)
5. Check out one VM via `pool.Get(ctx)`. If `ErrPoolExhausted` (should not happen on a fresh pool — treat as fatal plugin load error), set state to `StateFailed` and skip. Call `on_init()` if defined (this is where `db.define_table()` runs). Take global snapshot via `pool.snapshotGlobals(L)`. Return VM to pool. Runs **once per plugin**, not per VM.
6. Set state to `StateRunning`

**Shutdown sequence** (reverse dependency order — dependents shut down before their dependencies):
1. For each plugin:
   a. Check out one VM from the pool via `pool.Get(shutdownCtx)`. If error (pool exhausted or context expired — all VMs stuck), skip this plugin's `on_shutdown()` and log a warning.
   b. Call `on_shutdown()` if the function exists in globals
   c. Return the VM via `pool.Put(L)` (health check runs but result is irrelevant)
   d. Set state to `StateStopped`
2. After `on_shutdown()` has run (or been skipped) for ALL plugins: close all VM pools (`pool.Close()` calls `L.Close()` on every VM in the channel)
3. Close the plugin `*sql.DB` pool

Key constraint: `on_shutdown()` runs ON a pooled VM, so the pool must still be open during step 1. The shutdown context should have a generous timeout (e.g., 10s total). If `pool.Get()` times out (all VMs stuck), skip that plugin's `on_shutdown()` and log a warning.

#### `internal/plugin/pool.go` -- M
Goroutine-safe pool of pre-initialized `lua.LState` instances per plugin, with health validation on return.

```go
type VMPool struct {
    states      chan *lua.LState
    factory     func() *lua.LState
    size        int
    initPath    string
    pluginName  string           // for diagnostic logging
    closed      atomic.Bool      // drain safety flag (see Close section)
    initGlobals map[string]bool  // snapshot of global names after init (see Put section)
}
```

Key methods: `NewVMPool(size, initPath, pluginName, sandbox, dbAPI)`, `Get(ctx) (*lua.LState, error)`, `Put(L)`, `Close()`

- Channel-based pool (buffered channel of `*lua.LState`)
- `Get()` applies `L.SetContext(ctx)` before returning (timeout enforcement)
- `Put()` validates VM health before returning to pool (see below)

**Pool Exhaustion and Backpressure:**

When all VMs are checked out, `Get(ctx)` blocks on the channel receive until a VM is returned or the context expires. For HTTP handlers (Phase 2), a blocked `Get()` ties up a goroutine for up to the full execution timeout (5s default), which under sustained load can cascade into goroutine accumulation.

Mitigation: `Get()` uses a short acquisition timeout separate from the execution timeout. If no VM is available within 100ms, `Get()` returns a `ErrPoolExhausted` sentinel error. The HTTP bridge (Phase 2) translates this to HTTP 503 Service Unavailable with a `Retry-After` header. For Phase 1 (no HTTP), pool exhaustion during `on_init()` is a fatal plugin load error.

```go
var ErrPoolExhausted = errors.New("VM pool exhausted")

func (p *VMPool) Get(ctx context.Context) (*lua.LState, error) {
    // Short acquisition timeout — fail fast under load
    acquireCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
    defer cancel()

    select {
    case L := <-p.states:
        L.SetContext(ctx) // execution timeout uses the caller's context
        return L, nil
    case <-acquireCtx.Done():
        return nil, ErrPoolExhausted
    }
}
```

**Operation counter reset:** `DatabaseAPI.ResetOpCount()` is called by the Manager (or HTTP bridge) after `Get()` returns, not by the pool itself. The pool has no reference to `DatabaseAPI` — it manages `LState` lifecycle only. The `DatabaseAPI` is embedded as upvalues in the Go-bound Lua functions during VM factory creation, not stored on the pool. The caller pattern is:

```go
L, err := pool.Get(ctx)
if err != nil { return err }
defer pool.Put(L)
dbAPI.ResetOpCount()  // Manager holds dbAPI reference per-VM
// ... execute plugin code ...
```

In practice, each VM's `DatabaseAPI` can be stored in a `map[*lua.LState]*DatabaseAPI` on the Manager, or the reset can be triggered inside the Lua-bound function wrappers on first call after checkout (lazy reset). The exact wiring is an implementation detail — the contract is: op count resets once per checkout, before plugin code executes.

**VMPool.Close Drain Safety:**

`Close()` must not close the channel while VMs may still be returned via `Put()`. The close sequence is:
1. Set the `closed` atomic flag — `Put()` checks this and calls `L.Close()` directly instead of sending to the channel
2. Drain all VMs currently in the channel, calling `L.Close()` on each
3. Log if any VMs were not returned within a timeout (indicates leaked checkouts)

**VM Health Check on `Put()`:**

`Put(L)` performs these steps in order:
1. `L.SetTop(0)` -- clear Lua stack (not globals — globals persist)
2. `restoreGlobalSnapshot(L)` -- reset globals to post-init state (see below)
3. `validateVM(L)` -- verify critical Go-bound globals are intact
4. If healthy: return VM to the channel (or `L.Close()` directly if pool is closed)
5. If unhealthy: `L.Close()`, log warning with plugin name and which global was corrupted, create fresh VM via `factory()`, put replacement in channel

**VM Global Snapshot/Restore:**

`L.SetTop(0)` clears the Lua *stack*, not globals. Without cleanup, a handler that writes `counter = (counter or 0) + 1` leaks state across invocations — and across *different* handler/hook callbacks on the same VM. This is shared mutable state that most plugin authors won't expect.

After `on_init()` completes on the first VM, the pool takes a snapshot of all global names via `snapshotGlobals()`. This snapshot is shared by all VMs in the pool (they all loaded the same `init.lua`, so their baseline global sets are identical). On `Put()`, any global not in the snapshot is set to `nil`:

```go
// Called once after init.lua is loaded and on_init() has run
func (p *VMPool) snapshotGlobals(L *lua.LState) {
    p.initGlobals = make(map[string]bool)
    globals := L.Get(lua.GlobalsIndex).(*lua.LTable)
    globals.ForEach(func(key, _ lua.LValue) {
        if s, ok := key.(lua.LString); ok {
            p.initGlobals[string(s)] = true
        }
    })
}

// Called by Put() before validateVM
func (p *VMPool) restoreGlobalSnapshot(L *lua.LState) {
    var toRemove []string
    globals := L.Get(lua.GlobalsIndex).(*lua.LTable)
    globals.ForEach(func(key, _ lua.LValue) {
        if s, ok := key.(lua.LString); ok {
            if !p.initGlobals[string(s)] {
                toRemove = append(toRemove, string(s))
            }
        }
    })
    for _, name := range toRemove {
        L.SetGlobal(name, lua.LNil)
    }
}
```

**Limitations:** This removes *new* globals but does not restore *mutated* globals (e.g., plugin sets `plugin_info.name = "hacked"`). Full deep-clone snapshotting is too expensive. The snapshot approach catches the common case (forgetting `local`) while the VM health check catches the critical case (corrupted `db`/`log` modules). Plugin authors are still told to use `local` exclusively — the snapshot is a safety net, not license to use globals.

`validateVM(L)` checks:
- `L.GetGlobal("db")` is `*lua.LTable` (not `LNil`, `LString`, etc.)
- `L.GetField(dbTable, "query")` is `*lua.LFunction` with `.IsG == true`
- Same check for `"query_one"`, `"count"`, `"exists"`, `"insert"`, `"update"`, `"delete"`, `"transaction"`, `"define_table"`
- `L.GetGlobal("log")` is `*lua.LTable`
- `L.GetField(logTable, "info")` is `*lua.LFunction` with `.IsG == true`

**Design decisions:**
- **Check `IsG == true`, not pointer equality.** `LFunction.IsG` is `true` only for Go-bound functions. Catches replacement with pure Lua functions or non-function values. Does not require storing original pointers.
- **Replace, don't shrink.** Broken VMs are closed and replaced via `factory()`. Pool size stays constant. No degradation under corruption.
- **Lightweight.** 7 `GetGlobal`/`GetField` calls + type assertions. No Lua execution, no allocations. Nanosecond-scale overhead per `Put()`.
- **Log, don't panic.** Corrupted VM means plugin code is misbehaving. Log a warning with specifics (plugin name, which field failed) for admin investigation. Never crash the server.
- **Defense in depth with metatable freezing.** Health check is the fallback -- even if a metatable attack bypasses sandbox freezing (concern #6), the VM gets discarded on return.
- **Compatible with proxy pattern.** After module freezing, `L.GetGlobal("db")` returns the proxy (`*lua.LTable`), and `L.GetField(proxy, "query")` follows `__index` to the backing table's Go-bound function. All type assertions and `IsG` checks pass unchanged.

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

**Restricted `require` for plugin-local modules:**

The native `package` library is blocked, but a custom `require` is registered that resolves only within the plugin's own directory. Module names map to `<plugin_dir>/lib/<name>.lua` — no path traversal, no absolute paths, no loading outside the plugin directory.

```go
// RegisterPluginRequire replaces the global require with a sandboxed loader.
// Only resolves modules from <pluginDir>/lib/<name>.lua.
func RegisterPluginRequire(L *lua.LState, pluginDir string) {
    loaded := L.NewTable() // cache: module name -> returned value

    L.SetGlobal("require", L.NewFunction(func(L *lua.LState) int {
        name := L.CheckString(1)

        // Reject path traversal and absolute paths
        if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
            L.ArgError(1, fmt.Sprintf("invalid module name %q: must be a simple name, not a path", name))
            return 0
        }

        // Return cached module if already loaded
        if cached := L.GetField(loaded, name); cached != lua.LNil {
            L.Push(cached)
            return 1
        }

        path := filepath.Join(pluginDir, "lib", name+".lua")
        if _, err := os.Stat(path); err != nil {
            L.ArgError(1, fmt.Sprintf("module %q not found at %s", name, path))
            return 0
        }

        if err := L.DoFile(path); err != nil {
            L.RaiseError("error loading module %q: %s", name, err.Error())
            return 0
        }

        // Module returns a value (convention: return a table)
        ret := L.Get(-1)
        L.SetField(loaded, name, ret)
        L.Push(ret)
        return 1
    }))
}
```

**VM factory call sequence** (updated):
1. `ApplySandbox(L, cfg)` -- strip dangerous globals, load safe libs
2. `RegisterPluginRequire(L, pluginDir)` -- sandboxed module loader
3. `RegisterDBAPI(L, dbAPI)` -- populate `db` global with Go-bound functions
4. `RegisterLogAPI(L, logAPI)` -- populate `log` global with Go-bound functions
5. `FreezeModule(L, "db")` -- replace `db` with read-only proxy
6. `FreezeModule(L, "log")` -- replace `log` with read-only proxy

**`setmetatable`/`getmetatable` intentionally kept:** These are legitimate Lua patterns for OOP-style metatables on plugin-local tables (e.g., defining a local "class" with `__tostring` or `__index`). Stripping them would cripple normal Lua programming. The threat of replacing module metatables is fully blocked by the `FreezeModule` proxy pattern (`__metatable = "protected"`), and the `Put()` health check catches any corruption that slips through.

**Module Freezing via Proxy Tables:**

After `RegisterDBAPI` and `RegisterLogAPI` populate the `db` and `log` globals with Go-bound functions, `FreezeModule` replaces each global with an empty proxy table. The real functions live in a hidden backing table. Since the proxy is always empty, gopher-lua's internal `RawGetString` check always misses, routing all reads through `__index` and all writes through `__newindex`.

**Why a simple `__newindex` doesn't work:** gopher-lua's `setFieldString` (the internal handler for `table.field = value`) calls `RawGetString(key)` first. If the key **already exists**, it calls `RawSetString` directly -- `__newindex` is never invoked. So `db.query = nil` on a table where `query` already exists bypasses `__newindex` entirely.

```go
// FreezeModule replaces a global module table with a read-only proxy.
// The real functions are moved to a hidden backing table; the proxy
// delegates reads via __index and rejects writes via __newindex.
// __metatable prevents getmetatable/setmetatable from inspecting or
// replacing the metatable. rawget/rawset are already stripped by the sandbox.
func FreezeModule(L *lua.LState, moduleName string) {
    backing := L.GetGlobal(moduleName)
    if backing == lua.LNil {
        return
    }

    proxy := L.NewTable()
    mt := L.NewTable()

    // All reads delegate to the backing table
    L.SetField(mt, "__index", backing)

    // All writes raise an error
    L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
        key := L.CheckString(2)
        L.ArgError(2, fmt.Sprintf("cannot modify frozen module %q (key %q)", moduleName, key))
        return 0
    }))

    // Prevent metatable inspection/replacement
    L.SetField(mt, "__metatable", lua.LString("protected"))

    L.SetMetatable(proxy, mt)
    L.SetGlobal(moduleName, proxy)
}
```

**VM factory call sequence:** See updated sequence in sandbox.go section above (includes `RegisterPluginRequire` step).

**Attack surface:**

| Attack Vector | Outcome |
|---------------|---------|
| `db.query = nil` | `__newindex` fires -> Lua error "cannot modify frozen module" |
| `db.query = function() end` | `__newindex` fires -> Lua error |
| `db.new_field = true` | `__newindex` fires -> Lua error |
| `local q = db.query` | `__index` fires -> returns real Go-bound function |
| `db.query("tasks", {})` | `__index` fires -> returns Go function -> call proceeds normally |
| `getmetatable(db)` | Returns `"protected"` string (not the real metatable) |
| `setmetatable(db, {})` | Lua error: "cannot change a protected metatable" |
| `rawget(db, "query")` | `rawget` stripped by sandbox -> Lua error "attempt to call a nil value". Even if not stripped, would return `nil` (proxy is empty) — read-only, cannot corrupt VM. |
| `rawset(db, "query", nil)` | `rawset` stripped by sandbox -> Lua error "attempt to call a nil value" |
| `for k,v in pairs(db) do` | Iterates the proxy table's own keys — which is empty. Returns nothing. |
| `for k,v in pairs(db) do` (with `__pairs`) | Not implemented. gopher-lua does not support `__pairs` metamethod. |

**`pairs()` limitation:** Because the proxy table is empty (all real functions live in the hidden backing table), `pairs(db)` and `ipairs(db)` return nothing. This is a DX surprise for plugin authors debugging with `for k,v in pairs(db) do print(k,v) end`. gopher-lua does not support the `__pairs` metamethod (Lua 5.2+), so there is no way to make iteration work on frozen proxies. **Document this in plugin developer docs:** "The `db` and `log` modules do not support iteration. Call functions by name (e.g., `db.query(...)`)."

**Design decisions:**
- **Proxy pattern over `__newindex` alone.** As documented above, `__newindex` only fires for keys that don't exist in the table. The proxy is always empty, so every access (read or write) goes through metamethods.
- **`__metatable` as string.** Setting `__metatable` to a non-nil value causes `getmetatable()` to return that value instead of the real metatable, and causes `setmetatable()` to error. The string `"protected"` is conventional.
- **Freeze after registration.** `FreezeModule` must run after `RegisterDBAPI`/`RegisterLogAPI` populate the globals, since it moves the populated table to the backing position.
- **No performance impact.** Every `db.*` call already crosses the Go-Lua boundary. One extra metamethod dispatch per call is negligible compared to the SQL execution cost.
- **Compatible with VM health checks.** `L.GetGlobal("db")` returns the proxy, and `L.GetField(proxy, "query")` follows `__index` to the backing table's Go-bound function. `IsG == true` checks pass unchanged.
- **Defense in depth.** Even if a future gopher-lua change or edge case bypasses the proxy, the `Put()` health check (pool.go) catches corrupted VMs and replaces them.

#### `internal/plugin/db_api.go` -- L (downgraded from XL)
Sandboxed database API exposed as `db` Lua module. Delegates to the query builder in `internal/db/query_builder.go` which provides composable `QSelect` (WHERE maps, ORDER BY, LIMIT, OFFSET), `QInsert`, `QUpdate`, `QDelete`, `QCount`, and `QExists` -- all with `db.ValidTableName()`/`db.ValidColumnName()` validation, dialect-aware placeholders, and parameterized values.

**Lua API:**
```lua
-- Structured query (auto-prefixes table name)
local rows = db.query("tasks", {where = {status = "active"}, order_by = "created_at", limit = 50})
local row  = db.query_one("tasks", {where = {id = "abc"}})

-- Aggregates
local n    = db.count("tasks", {where = {status = "active"}})  -- returns integer
local has  = db.exists("tasks", {where = {id = "abc"}})        -- returns boolean

-- Mutations
db.insert("tasks", {id = db.ulid(), title = "My task", status = "active"})
db.update("tasks", {set = {status = "done"}, where = {id = "abc"}})
db.delete("tasks", {where = {id = "abc"}})

-- Transactions (scoped executor swap; all db.* calls route through *sql.Tx)
-- Nested db.transaction() calls are rejected. Lua error() triggers rollback.
local ok, err = db.transaction(function()
    db.insert("tasks", {id = db.ulid(), title = "Task 1"})
    db.insert("tasks", {id = db.ulid(), title = "Task 2"})
end)

-- Helpers
local id = db.ulid()        -- generate ULID
local now = db.timestamp()   -- RFC3339 UTC timestamp (replaces os.date which is sandboxed)
```

**Phase 1 Query Scope:** The `where` parameter supports **equality conditions only** (`AND`-joined). Operators like `>`, `<`, `IN`, `LIKE`, `IS NOT NULL`, and `OR` are not available in Phase 1. This is an intentional scope limitation — the structured API prevents SQL injection categorically and covers the foundational CRUD patterns. Advanced query capabilities (operator syntax, raw parameterized queries) will be added after the core engine is proven and tested. Plugin authors needing complex queries in Phase 1 should decompose them into multiple equality queries with Lua-side filtering.

```go
type DatabaseAPI struct {
    conn          *sql.DB
    currentExec   db.Executor // normally conn; swapped to *sql.Tx inside transaction callbacks
    pluginName    string
    dialect       db.Dialect
    maxRows       int  // default 100, hard cap 10000 (enforced by db.QSelect)
    maxTxOps      int  // default 10
    inTx          bool // prevents nested db.transaction() calls
    opCount       int  // incremented on every db.* call, reset on VMPool.Get()
    maxOpsPerExec int  // default 1000, configurable via ManagerConfig
}
```

**Mapping from Lua API to `db` query builder:**

| Lua call | Go call |
|----------|---------|
| `db.query(table, opts)` | `db.QSelect(ctx, api.currentExec, dialect, SelectParams{Table: prefixed, Where: opts.where, OrderBy: opts.order_by, Limit: opts.limit})` |
| `db.query_one(table, opts)` | `db.QSelectOne(ctx, api.currentExec, dialect, SelectParams{...})` |
| `db.insert(table, values)` | `db.QInsert(ctx, api.currentExec, dialect, InsertParams{Table: prefixed, Values: values})` |
| `db.update(table, opts)` | `db.QUpdate(ctx, api.currentExec, dialect, UpdateParams{Table: prefixed, Set: opts.set, Where: opts.where})` |
| `db.count(table, opts)` | `db.QCount(ctx, api.currentExec, dialect, prefixed, opts.where)` |
| `db.exists(table, opts)` | `db.QExists(ctx, api.currentExec, dialect, prefixed, opts.where)` |
| `db.delete(table, opts)` | `db.QDelete(ctx, api.currentExec, dialect, DeleteParams{Table: prefixed, Where: opts.where})` |
| `db.transaction(fn)` | `types.WithTransaction()` (from `internal/db/types/transaction.go`) + scoped executor swap (`currentExec = tx`) + `L.CallByParam` with `Protect: true` |

All `db.*` methods use `api.currentExec` (not `api.conn` directly). The `db.Executor` interface accepts both `*sql.DB` and `*sql.Tx`, so during a transaction all operations automatically route through the `*sql.Tx`. Table names auto-prefixed with `plugin_<name>_`. The query builder handles all identifier validation and parameterization.

**Transaction Scoping via Executor Swap:**

The core problem: inside `db.transaction(function() ... end)`, nothing prevents Lua code from calling `db.query()` which would use the raw `*sql.DB` instead of the transaction's `*sql.Tx`. In SQLite (even with WAL mode, which only allows concurrent readers — not concurrent writers on separate connections) this deadlocks; in all drivers it breaks atomicity.

The solution: `DatabaseAPI.currentExec` defaults to `conn` but is temporarily swapped to the `*sql.Tx` for the duration of the transaction callback. All `db.*` methods use `currentExec`, so they automatically participate in the transaction without any action from plugin authors.

```go
func (api *DatabaseAPI) luaTransaction(L *lua.LState) int {
    if api.inTx {
        L.ArgError(1, "nested transactions are not allowed")
        return 0
    }

    fn := L.CheckFunction(1)
    ctx := L.Context()

    err := types.WithTransaction(ctx, api.conn, func(tx *sql.Tx) error {
        // Swap executor -- all db.* calls now go through the transaction
        prevExec := api.currentExec
        api.currentExec = tx
        api.inTx = true
        defer func() {
            api.currentExec = prevExec
            api.inTx = false
        }()

        // Call Lua function; Protect: true means error() returns err instead of panic
        if err := L.CallByParam(lua.P{
            Fn:      fn,
            NRet:    0,
            Protect: true,
        }); err != nil {
            return err // triggers rollback via WithTransaction
        }
        return nil
    })

    if err != nil {
        L.Push(lua.LFalse)
        L.Push(lua.LString(err.Error()))
        return 2
    }

    L.Push(lua.LTrue)
    L.Push(lua.LNil)
    return 2
}
```

**Design decisions:**
- **No separate `tx` table in Lua.** The callback is `function()` not `function(tx)`. All `db.*` calls automatically route through the transaction. This eliminates the entire class of bugs where plugin authors accidentally use `db.*` instead of `tx.*`.
- **Thread safety relies on 1:1 binding.** Each `DatabaseAPI` instance is bound to exactly one `lua.LState`, and each `LState` is used by one goroutine at a time via the VM pool. This means no concurrent access to `currentExec`, `inTx`, or `opCount`. **This invariant must be enforced architecturally:** the `DatabaseAPI` is created inside the VM factory and never shared. If Phase 4 hot reload or recovery ever reuses a `DatabaseAPI` across VMs, a data race will occur. Add a comment in `NewDatabaseAPI()`: `// INVARIANT: each DatabaseAPI instance is bound to exactly one LState. Never share across VMs.`
- **Nested transaction prevention.** `inTx` flag rejects `db.transaction()` inside an existing callback. SQLite doesn't support nested transactions, and SAVEPOINT adds complexity with no clear plugin use case.
- **Scoped restore via defer.** `currentExec` is always restored even if the Lua callback panics. The `defer` in `WithTransaction` handles rollback.
- **Resolves both concern #4 and #5.** Scoping `db.*` to the transaction (concern #4) eliminates the deadlock risk (concern #5) because no second connection is ever acquired during a transaction.

**Per-Plugin Operation Rate Limiting:**

Every `db.*` call (query, insert, update, delete) increments `api.opCount`. When `opCount > maxOpsPerExec`, the call returns a Lua error: `"plugin <name> exceeded maximum operations per execution (1000)"`. The counter resets after `VMPool.Get()` returns (the Manager calls `dbAPI.ResetOpCount()` before executing plugin code), NOT in `Put()`. This means the budget is per-checkout, not per-VM-lifetime.

```go
// Called at the start of every db.query/insert/update/delete
func (api *DatabaseAPI) checkOpLimit(L *lua.LState) error {
    api.opCount++
    if api.opCount > api.maxOpsPerExec {
        return fmt.Errorf("plugin %q exceeded maximum operations per execution (%d): %w",
            api.pluginName, api.maxOpsPerExec, ErrOpLimitExceeded)
    }
    return nil
}

// Sentinel error for operation budget exhaustion. Callers can check with errors.Is().
var ErrOpLimitExceeded = errors.New("operation limit exceeded")

// Called by VMPool.Get() before returning VM to caller
func (api *DatabaseAPI) ResetOpCount() {
    api.opCount = 0
}
```

**Why per-checkout, not per-second:** A time-based rate limiter needs a background goroutine or token bucket, adds complexity, and the 5-second execution timeout already caps wall-clock abuse. The per-checkout counter is stateless, zero-cost, and directly limits the damage a single plugin invocation can do.

**Error Convention:**

All `db.*` functions follow a consistent error contract based on whether the error is recoverable (caller can handle) or a programming error (caller made a mistake):

| Situation | Convention | Example |
|-----------|-----------|---------|
| Query returns no rows | Returns empty table (not nil) | `db.query()` on valid table, no matches |
| `query_one` returns no row | Returns `nil` | `if not row then` is idiomatic |
| Constraint violation | Returns `nil, errmsg` | Duplicate key, FK violation |
| Table does not exist | Returns `nil, errmsg` | Typo in table name |
| Transaction failure | Returns `false, errmsg` | Deadlock, timeout |
| Wrong argument type | Raises `error()` | Passing number where table expected |
| Missing required argument | Raises `error()` | `db.insert()` with no table name |
| Operation budget exceeded | Raises `error()` | Unrecoverable within this checkout |
| Empty `where` in `db.update()` | Raises `error()` | Safety: prevents full-table update |
| Empty `where` in `db.delete()` | Raises `error()` | Safety: prevents full-table delete |
| Empty/nil `where` in `db.query()` | Returns all rows (up to limit) | Full-table scan is safe for reads |
| Empty/nil `where` in `db.query_one()` | Returns an arbitrary row | `LIMIT 1` with no filter; row order is non-deterministic without `order_by` |
| Empty/nil `where` in `db.count()` | Returns total row count | `SELECT COUNT(*)` with no filter |
| Empty/nil `where` in `db.exists()` | Returns `true` if table has any rows | Equivalent to `SELECT EXISTS(...)` |

**Empty WHERE policy:** Read operations (`query`, `query_one`, `count`, `exists`) allow empty or nil `where` — this is a common pattern for "get all" or "check if table has data." Write operations (`update`, `delete`) reject empty `where` with `error()` as a safety guard against accidental full-table mutation. This asymmetry is intentional: reads are safe to scan; writes are dangerous to broadcast.

Recoverable errors return `nil, errmsg` (or `false, errmsg` for transactions). Programming errors raise via `error()`. Plugin authors use `if not result then` for recoverable errors and `pcall` only when they need to catch programming errors.

**Note on `db.transaction()`:** Both callback failures (Lua `error()` inside the callback) and commit failures (database-level commit error) return `false, errmsg`. Plugin authors cannot distinguish the two — both trigger rollback. This is intentional: the plugin's response to either case is the same (log the error, retry or report to user).

**Row marshaling (`db.Row` to Lua table):** `string` -> `LString`, `int64`/`float64` -> `LNumber`, `bool` -> `LBool`, `nil` -> `LNil`, `[]byte` -> `LString`.

#### `internal/plugin/schema_api.go` -- L
Lua-defined schema creation, validated by Go. Wraps `db.DDLCreateTable()` and `db.DDLCreateIndex()` for cross-dialect DDL with plugin-specific DX conveniences.

**7 Abstract Column Types** (via `db.ValidColumnTypes`):

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
- **FK namespace validation:** ForeignKey `ref_table` must start with same `plugin_<name>_` prefix (enforced by schema_api.go, not the query builder).

**Auto-Injection Contract:**

The names `id`, `created_at`, and `updated_at` are reserved. If a plugin's `columns` list includes any of them, `schema_api.go` rejects with a validation error: `"column 'id' is auto-injected and cannot be defined manually"`.

Injected column definitions (always the same regardless of dialect — type mapping handled by `db.SQLType()`):
- `id`: `{Name: "id", Type: "text", NotNull: true, PrimaryKey: true}` — first column
- `created_at`: `{Name: "created_at", Type: "text", NotNull: true}` — after last user column
- `updated_at`: `{Name: "updated_at", Type: "text", NotNull: true}` — last column

Runtime behavior in `db_api.go`:
- **`db.insert()`:** If the values map lacks `id`, auto-generates via ULID. If it lacks `created_at` or `updated_at`, auto-sets both to RFC3339 UTC. If the plugin explicitly provides any of these, the explicit value is used (no override).
- **`db.update()`:** Auto-sets `updated_at` in the `set` map to RFC3339 UTC. If the plugin explicitly provides `updated_at` in `set`, the explicit value is used. `id` and `created_at` are never auto-set on update.

**Lua API (called inside `on_init()`):**
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

-- What actually gets created (dialect-aware via db.DDLCreateTable):
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
    Columns    []db.CreateColumnDef
    Indexes    []db.IndexDef
    ForeignKeys []db.ForeignKeyDef
}
```

**Implementation:** schema_api.go translates the Lua table definition into `db.DDLCreateTableParams`, auto-injects id/timestamp columns, prefixes the table name, validates FK namespace isolation, then calls `db.DDLCreateTable()`. All type mapping and DDL generation is handled by the query builder's cross-dialect infrastructure.

**Validation:** Type in `db.ValidColumnTypes` (7 types), names pass `db.ValidColumnName()`, max 64 columns, exactly one PK (auto-injected). Uses `CREATE TABLE IF NOT EXISTS` (idempotent). Cross-dialect support is built-in via `db.SQLType()` — no SQLite-only limitation.

#### `internal/plugin/lua_helpers.go` -- S
Go<->Lua value conversion utilities.

```go
func LuaTableToMap(L *lua.LState, tbl *lua.LTable) map[string]any
func MapToLuaTable(L *lua.LState, m map[string]any) *lua.LTable
func GoValueToLua(L *lua.LState, v any) lua.LValue
func LuaValueToGo(v lua.LValue) any
func SQLRowsToLuaTable(L *lua.LState, rows *sql.Rows) (*lua.LTable, error)
```

**Query result contract:** `SQLRowsToLuaTable` always returns a Lua **sequence** — a contiguous 1-indexed array of row tables. Empty result sets return an empty table (not `nil`). This is a hard contract that plugin authors depend on: `#results`, `ipairs(results)`, and `results[1]` must all work correctly. `db.query_one()` returns a single row table on match, or `nil` on no match (allowing `if not row then`).

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

The following DB pool fields **already exist** in config.go (added during plugin pool infrastructure work):
```go
// EXISTING — do not re-add:
Plugin_DB_MaxOpenConns    int    `json:"plugin_db_max_open_conns"`
Plugin_DB_MaxIdleConns    int    `json:"plugin_db_max_idle_conns"`
Plugin_DB_ConnMaxLifetime string `json:"plugin_db_conn_max_lifetime"`
```

Add these new fields for plugin system configuration. Uses `Plugin_` prefix (no `DB` qualifier) because these govern the plugin runtime, not the database pool:
```go
Plugin_Enabled    bool   `json:"plugin_enabled"`
Plugin_Directory  string `json:"plugin_directory"`
Plugin_Max_VMs    int    `json:"plugin_max_vms"`     // per plugin, default 4
Plugin_Timeout    int    `json:"plugin_timeout"`      // seconds, default 5
Plugin_Max_Ops    int    `json:"plugin_max_ops"`      // per VM checkout, default 1000
```

**Naming convention:** `Plugin_DB_*` = database pool tuning (already exists). `Plugin_*` = plugin runtime behavior (new). The two-level prefix disambiguates without requiring a nested config struct.

#### `cmd/serve.go` -- MODIFY init (S)
The plugin `*sql.DB` pool is already opened by `initPluginPool()` at serve.go:86. Replace the `_ = pluginPool` placeholder with the Manager init:
```go
// Already exists:
pluginPool, pluginPoolCleanup, err := initPluginPool(cfg)
// ...
defer pluginPoolCleanup()

// Replace "_ = pluginPool" with:
pluginManager := initPluginManager(rootCtx, cfg, pluginPool)
defer pluginManager.Shutdown(rootCtx)
```

#### `cmd/helpers.go` -- ADD helper (S)
```go
func initPluginManager(ctx context.Context, cfg *config.Config, pool *sql.DB) *plugin.Manager {
    // If !cfg.Plugin_Enabled, return no-op Manager (or nil with nil checks at call sites)
    // Create Manager with the already-opened pool, call LoadAll
    // Do NOT open a new pool -- pool is created by initPluginPool() in serve.go
}
```

### Testing Strategy

| Test File | Coverage |
|-----------|----------|
| `manager_test.go` | Discovery from temp dirs, lifecycle transitions, dependency validation |
| `pool_test.go` | Concurrent Get/Put, context cancellation, pool exhaustion returns ErrPoolExhausted within 100ms (not hang), Close() drains safely with checked-out VMs, global snapshot removes new globals on Put(), global snapshot preserves init-time globals |
| `sandbox_test.go` | Blocked globals are nil, allowed globals work, timeout enforcement, frozen module writes rejected, frozen module reads work, getmetatable returns "protected", setmetatable raises error, restricted require loads from lib/ only, require rejects path traversal, pairs(db) returns nothing (documented DX limitation) |
| `db_api_test.go` | All CRUD + count/exists on in-memory SQLite, table prefixing, injection prevention, type marshaling, error convention (nil/errmsg for recoverable, error() for programming errors), query results are always sequences, empty WHERE allowed for reads, empty WHERE raises error() for update/delete, operation budget exceeded returns ErrOpLimitExceeded, operation budget resets on Get() |
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
- `internal/db/query_builder.go` -- **Primary foundation for plugin DB API.** Provides `QSelect`, `QSelectOne`, `QInsert`, `QUpdate`, `QDelete`, `QCount`, `QExists` with composable WHERE/ORDER/LIMIT, `Executor` interface for `*sql.DB`/`*sql.Tx` interop, `Dialect` type for cross-database placeholders, full identifier validation, parameterized values, 10000-row hard cap. Also provides cross-dialect DDL: 7 abstract `ColumnType` values with `SQLType()` mapping, `DDLCreateTable()` with indexes/FKs, and standalone `DDLCreateIndex()`. Fully tested.
- `internal/db/audited/audited.go` -- Transaction pattern reference; also the natural hook injection point for Phase 3 (preferred over wrapping full DbDriver interface)
- `internal/db/init.go` -- Singleton DB pattern and plugin pool factory. `OpenPool()` creates an independent `*sql.DB` with configurable pool limits for plugin isolation. `DefaultPluginPoolConfig()` returns conservative defaults (10 max open, 5 max idle). Plugin manager uses `OpenPool()` instead of `GetConnection()` to avoid coupling to the core CMS pool.
- `cmd/serve.go` -- Startup integration point. Plugin init must happen before `router.NewModulacmsMux()` (line 123) for Phase 2 route registration to work.

---

## Phase 2: HTTP Integration — COMPLETE (2026-02-15)

**Goal:** Plugins register HTTP endpoints served via existing `net/http.ServeMux`. All plugin routes namespaced under `/api/v1/plugins/<plugin_name>/`.

**Status: IMPLEMENTED** — Full implementation plan in `ai/plans/plugin/PLUGIN_PHASE_2.md` (768 lines).

### Implementation Summary

Implemented via 6 parallel/sequential agents (A-F). All tests pass, no regressions.

#### New Files Created

| File | Lines | Purpose |
|------|-------|---------|
| `internal/plugin/http_api.go` | 213 | Lua `http.handle()` and `http.use()` registration with method/path validation, phase guard, duplicate detection, route count limit (50/plugin) |
| `internal/plugin/http_request.go` | 437 | Request/response Go↔Lua conversion, `PluginErrorResponse` JSON schema, blocked response headers (11 entries), `extractClientIP` with trusted proxy support, `BuildLuaRequest`/`WriteLuaResponse` |
| `internal/plugin/http_bridge.go` | 882 | Central coordinator: `plugin_routes` DB table (3-dialect DDL), route approval workflow, ServeMux dispatch (20-step request lifecycle), per-IP rate limiting with cleanup, body/response size limits, auth enforcement, graceful shutdown with inflight drain |
| `internal/plugin/http_api_test.go` | 620 | Unit tests for Lua API registration (66 test cases) |
| `internal/plugin/http_request_test.go` | 1117 | Unit tests for request/response conversion (36 tests) |
| `internal/plugin/http_bridge_test.go` | 1826 | Unit tests for bridge dispatch, approval, auth, rate limiting, security headers, middleware chain (30+ tests) |
| `internal/plugin/http_integration_test.go` | ~350 | End-to-end tests using external fixture plugins (7 integration tests) |

**Test fixtures** (7 plugins in `internal/plugin/testdata/plugins/`):
- `http_plugin/` — Task tracker with GET/POST /tasks, middleware enrichment
- `http_public_plugin/` — Public routes (webhook, status)
- `http_params_plugin/` — Path parameter extraction via `{id}`
- `http_error_plugin/` — Handler `error()` produces generic 500
- `http_timeout_plugin/` — Infinite loop triggers 504 HANDLER_TIMEOUT
- `http_middleware_plugin/` — Middleware enriches request, handler reads enriched field
- `http_blocked_headers/` — Blocked header filtering (Set-Cookie, CORS stripped)

#### Existing Files Modified

| File | Changes |
|------|---------|
| `internal/plugin/manager.go` | Bridge field + SetBridge/Bridge methods, `RegisterHTTPAPI` in VM factory, `FreezeModule(http)`, phase flag (`in_init`) before/after `on_init()`, `bridge.RegisterRoutes` call, idempotent Shutdown via `sync.Once` |
| `internal/plugin/pool.go` | HTTP module health check in `validateVM` (verifies `http` table and `http.handle` is Go-bound) |
| `internal/plugin/pool_test.go` | HTTP module stubs in test VM factory |
| `internal/config/config.go` | 5 new fields: `Plugin_Max_Request_Body`, `Plugin_Max_Response_Body`, `Plugin_Rate_Limit`, `Plugin_Max_Routes`, `Plugin_Trusted_Proxies` |
| `internal/middleware/http_middleware.go` | `SetAuthenticatedUser` helper for test injection, plugin prefix exemption in `HTTPPublicEndpointMiddleware` |
| `internal/middleware/ratelimit.go` | `ipLimiterEntry` with `lastSeen` tracking, cleanup goroutine eviction, `Close()` method |
| `internal/router/mux.go` | Bridge mounting, 3 admin endpoints (list/approve/revoke routes), `adminOnly` wrapper |
| `cmd/serve.go` | Bridge wiring, explicit ordered shutdown (bridge.Close → manager.Shutdown) |
| `cmd/helpers.go` | Bridge creation before `LoadAll` |

#### Key Architecture Decisions

- **Admin approval workflow**: Routes start unapproved, admin approves via `POST /api/v1/admin/plugins/routes/approve`. Unapproved routes return uniform 404 (prevents enumeration).
- **Version change re-approval**: Plugin version bump deletes all routes from `plugin_routes` DB table, forces admin re-approval.
- **Per-VM handler binding**: Each VM has its own `__http_handlers` Lua table. Route metadata registered once (first VM), but handler lookup uses the checked-out VM's table.
- **Public flag change revokes approval**: If a route changes between public/authenticated, approval is revoked for that specific route.
- **Bridge owns auth for plugin routes**: `HTTPPublicEndpointMiddleware` exempts `/api/v1/plugins/` prefix; the bridge enforces public/authenticated distinctions per-route.
- **Trusted proxy IP extraction**: New `extractClientIP` (right-to-left X-Forwarded-For) replaces middleware `getIP` which blindly trusts XFF.
- **gopher-lua timeout detection**: `errors.Is(callErr, context.DeadlineExceeded)` fails because `*lua.ApiError` doesn't implement `Unwrap()`. Fixed with `execCtx.Err()` fallback check.

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

The `DbDriver` interface has 100+ methods. Wrapping it is a massive maintenance surface. Instead, inject hooks at the `audited.Create`/`audited.Update`/`audited.Delete` generic functions, which are the primary entry point for content mutations with transactions. (Note: not all mutations go through `audited` — see Known Gap #4. Hooks fire on audited mutations only; internal CMS operations like `UpdateUserSshKeyLastUsed` bypass hooks by design.) Add an optional `HookRunner` parameter (or embed in `AuditContext`) that the audited functions call at the appropriate points within the transaction.

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
Plugin_Hot_Reload     bool   `json:"plugin_hot_reload"`
Plugin_Max_Failures   int    `json:"plugin_max_failures"`    // circuit breaker, default 5
Plugin_Reset_Interval string `json:"plugin_reset_interval"`  // default "60s"
```

---

## Known Gaps and Design Decisions

Items to address during implementation or document as intentional limitations:

### Phase 1 Implementation Requirements

1. **Dependency cycle detection** -- `LoadAll` must topologically sort plugins and detect circular dependencies. Not just "listed plugins must exist."
2. **Plugin name collision** -- Names like `a_` and `a` produce `plugin_a__tasks` vs `plugin_a_tasks`. Validation must reject trailing underscores or enforce separator uniqueness.
3. **VM global state leakage** -- **Mitigated in Phase 1** via `restoreGlobalSnapshot()` in `Put()` (see pool.go section). New globals created after init are removed on VM return. Mutated pre-existing globals (e.g., modifying a table value in `plugin_info`) are NOT restored — full deep-clone is too expensive. Plugin authors must still use `local` exclusively; the snapshot is a safety net. Document this in plugin developer docs.
4. **Not all mutations go through `audited` layer** -- Methods like `UpdateUserSshKeyLastUsed` bypass audit. The Phase 3 hook injection claim that audited is "the single entry point" is inaccurate. Document as design decision: hooks fire on audited mutations only; non-audited mutations are internal CMS operations not relevant to plugin hooks.
5. **Dependencies are name-only, no version constraints** -- Plugin B upgrades to 2.0 with breaking changes, Plugin A breaks silently. Acceptable for Phase 1; version constraint syntax deferred.
6. **Duplicate plugin directories** -- If two directories contain plugins with the same `plugin_info.name`, the second is rejected with a logged error during `LoadAll`. First-discovered wins (filesystem scan order).

### Intentional Phase 1 Scope Limitations

7. **Equality-only WHERE** -- The structured `db.query()` API supports only `AND`-joined equality conditions. Operators (`>`, `<`, `IN`, `LIKE`, `IS NOT NULL`) and `OR` are not available. Advanced query capabilities (operator syntax in WHERE maps, and/or `db.query_raw()` with parameterized queries) will be added after the core engine is proven and tested.
8. **No cross-table JOINs** -- `db.query()` accepts a single table name. Plugins with related tables must perform multiple queries with Lua-side joining. JOIN support will be evaluated after the foundational query API is stable.
9. **No schema migration path** -- `db.define_table` uses `CREATE TABLE IF NOT EXISTS`. Adding columns in a plugin update silently uses the old schema. `db.define_table` should detect schema drift (column count or names changed) and log a warning with plugin name and table name. Full migration support deferred to Phase 4.

### Design Decisions (Resolved Questions)

10. **Plugin tables are never auto-dropped.** When a plugin directory is removed and the CMS restarts, its prefixed tables remain in the database. This is intentional: auto-dropping risks data loss, and orphaned tables are harmless. A `plugin cleanup` CLI command (Phase 4) can list and optionally drop orphaned `plugin_*` tables.
11. **Failed plugin recovery before Phase 4 circuit breaker:** A failed plugin stays in `failed` state with its error logged. It does not affect other plugins. Admin must fix the plugin code and restart the CMS. No automatic retry.
12. **Phase 2 route wiring:** `NewModulacmsMux` accepts an optional `*plugin.HTTPBridge` parameter. Plugin routes are mounted as a single catch-all handler under `/api/v1/plugins/`. The mux signature change is additive (new optional parameter).
13. **Hot reload forward-compatibility:** The Phase 1 channel-based pool design is forward-compatible with Phase 4 hot reload. Drain is implemented by: stop accepting new `Get()` calls (close a done channel), wait for all checked-out VMs to be returned via `Put()`, then `Close()` the old pool and create a new one. No structural changes needed to pool.go.

---

## Performance Notes

- **VM Pool:** Default 4 VMs per plugin. Increase for HTTP-heavy plugins, 2 sufficient for hook-only.
- **Boundary crossing:** Structured query API minimizes crossings -- one table in, one table out per operation.
- **Connection isolation:** Plugins use a separate `*sql.DB` pool via `db.OpenPool()` (default 10 max open). Core CMS pool (25 connections) is unaffected by plugin load.
- **Startup:** Plugins loaded sequentially (dependency order), VM pool init parallelized per plugin.

---

## Example Plugin: Task Tracker

```lua
-- Plugin contract globals: plugin_info (required), on_init (optional), on_shutdown (optional)
-- Everything else must be declared local.

plugin_info = {
    name        = "task_tracker",
    version     = "1.0.0",
    description = "Simple task tracking for content workflows",
}

function on_init()
    -- db.define_table must be called inside on_init(), not at module scope.
    -- The temp VM used for manifest extraction is discarded; only on_init()
    -- runs in the production VM with full db access.
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

    log.info("Task tracker initialized")

    -- db.exists() avoids fetching rows just to check presence
    if not db.exists("tasks", {}) then
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
