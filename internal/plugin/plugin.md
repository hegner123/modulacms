# plugin

The plugin package implements the ModulaCMS Lua plugin system for runtime extension of the CMS with sandboxed Lua plugins that can define database tables, execute queries, register HTTP handlers, and integrate with the CMS lifecycle.

## Overview

The plugin system extends ModulaCMS with runtime Lua plugins that run in sandboxed virtual machines with controlled access to database operations, logging, and HTTP routing. Plugins are loaded from directories containing init.lua files, validated, topologically sorted by dependencies, and executed with per-plugin VM pools for concurrency.

Architecture flow: Manager discovers plugins, creates VMPool per plugin with sandboxed LState instances, applies sandbox restrictions via ApplySandbox, registers db log http APIs, freezes modules via read-only proxies, runs on_init lifecycle hook, and snapshots global state for VM restoration on checkout return.

All plugin tables are prefixed with plugin_pluginname_ and validated via query builder identifier checks. Plugins cannot access core CMS tables or other plugins' tables.

### Constants

```go
const (
    StateDiscovered PluginState = iota
    StateLoading
    StateRunning
    StateFailed
    StateStopped
)
```

PluginState enum representing plugin lifecycle states. StateDiscovered means manifest extracted but not yet loaded. StateLoading means dependencies validated and VMPool creation in progress. StateRunning means on_init succeeded and plugin is active. StateFailed means loading or init failed with reason in FailedReason field. StateStopped means on_shutdown has run.

```go
const (
    MaxPluginRequestBody  = 1 << 20
    MaxPluginResponseBody = 5 << 20
    DefaultPluginRateLimit = 100
    PluginRoutePrefix     = "/api/v1/plugins/"
)
```

HTTP bridge constants for request body size limit 1 MB, response body size limit 5 MB, default rate limit per IP 100 requests per second, and route prefix for all plugin HTTP endpoints.

```go
const MaxRoutesPerPlugin = 50
```

Maximum number of HTTP routes a single plugin can register via http.handle. Exceeding this limit raises a Lua error.

```go
const acquireTimeout = 100 * time.Millisecond
```

Maximum time VMPool.Get waits for an available VM. Provides backpressure under load rather than tying up goroutines for the full execution timeout.

### Variables

```go
var ErrPoolExhausted = errors.New("VM pool exhausted")
```

Returned by VMPool.Get when no VM is available within acquisition timeout. Caller should treat as backpressure. HTTP bridge translates to 503 Service Unavailable with Retry-After header.

```go
var ErrOpLimitExceeded = errors.New("operation limit exceeded")
```

Returned when a plugin's per-checkout operation budget is exhausted. Callers can check with errors.Is. Budget resets on next VMPool.Get via DatabaseAPI.ResetOpCount.

```go
var reservedColumns = map[string]bool{
    "id":         true,
    "created_at": true,
    "updated_at": true,
}
```

Columns auto-injected by schema_api.go that plugins cannot define. Plugins including these names in columns list receive validation error.

```go
var validMethods = map[string]bool{
    "GET":    true,
    "POST":   true,
    "PUT":    true,
    "DELETE": true,
    "PATCH":  true,
}
```

Allowlist of HTTP methods accepted by http.handle. Other methods raise Lua error.

```go
var strippedGlobals = []string{
    "dofile", "loadfile", "load",
    "rawget", "rawset", "rawequal", "rawlen",
}
```

Dangerous base library globals removed after OpenBase. These allow arbitrary code loading or bypass metatable protections. setmetatable and getmetatable are intentionally kept for proxy freezing.

```go
var blockedResponseHeaders = map[string]bool{
    "access-control-allow-origin": true,
    "access-control-allow-credentials": true,
    "set-cookie": true,
    "transfer-encoding": true,
    "content-length": true,
    "cache-control": true,
}
```

Response headers that Lua plugins cannot set. Prevents CORS bypass, session fixation, response smuggling, and cache poisoning.

#### func NewManager

```go
func NewManager(cfg ManagerConfig, pool *sql.DB, dialect db.Dialect) *Manager
```

Creates a new plugin Manager with the given configuration. Zero-value config fields are replaced with defaults. MaxVMsPerPlugin defaults to 4, ExecTimeoutSec to 5, MaxOpsPerExec to 1000. The db pool must be a separate connection opened via db.OpenPool for isolation.

#### func NewVMPool

```go
func NewVMPool(size int, factory func() *lua.LState, initPath string, pluginName string) *VMPool
```

Creates a pool of size pre-initialized Lua VMs using provided factory function. Factory must produce fully sandboxed VMs with ApplySandbox, RegisterPluginRequire, RegisterDBAPI, RegisterLogAPI, RegisterHTTPAPI, and FreezeModule all applied. initPath is absolute path to plugin's init.lua for diagnostic logging. pluginName is used in log messages when unhealthy VMs are detected. Global snapshot is NOT taken here, happens after on_init in manager because on_init may define new globals that should be part of baseline.

#### func NewDatabaseAPI

```go
func NewDatabaseAPI(conn *sql.DB, pluginName string, dialect db.Dialect, maxOpsPerExec int) *DatabaseAPI
```

Creates a new DatabaseAPI bound to given connection and plugin. INVARIANT: each DatabaseAPI instance is bound to exactly one LState, never share across VMs. maxOpsPerExec defaults to 1000 if zero or negative.

#### func NewHTTPBridge

```go
func NewHTTPBridge(manager *Manager, pool *sql.DB, dialect db.Dialect) *HTTPBridge
```

Creates a new HTTPBridge tied to given Manager. Config field wiring for Plugin_Max_Request_Body, Plugin_Max_Response_Body, Plugin_Rate_Limit, Plugin_Trusted_Proxies uses package-level constants as defaults. Starts background cleanup goroutine for stale IP rate limiters every 5 minutes, removing entries not seen for more than 10 minutes.

#### func ApplySandbox

```go
func ApplySandbox(L *lua.LState, cfg SandboxConfig)
```

Configures a Lua VM with a safe stdlib subset. Loads only base, table, string, math libraries, plus coroutine if cfg.AllowCoroutine is true. Dangerous globals are stripped after loading. The io, os, package, debug, and channel libraries are never loaded. The LState must have been created with SkipOpenLibs set to true.

#### func RegisterPluginRequire

```go
func RegisterPluginRequire(L *lua.LState, pluginDir string)
```

Replaces the global require with a sandboxed loader. Only resolves modules from pluginDir/lib/name.lua. Module names must be simple identifiers, path traversal characters rejected. Loaded modules are cached, subsequent require calls for same name return cached value. Uses L.ArgError for validation failures like bad module name or module not found. Uses L.RaiseError for load failures like syntax error in module file.

#### func FreezeModule

```go
func FreezeModule(L *lua.LState, moduleName string)
```

Replaces a global module table with a read-only proxy. Real functions are moved to hidden backing table, proxy delegates reads via __index metatable and rejects writes via __newindex. __metatable prevents getmetatable/setmetatable from inspecting or replacing the metatable. After freezing, db.query works via __index delegation, db.query assignment raises error, getmetatable returns "protected" string, setmetatable raises error. Limitation: pairs returns nothing because proxy is empty, documented DX tradeoff.

#### func RegisterDBAPI

```go
func RegisterDBAPI(L *lua.LState, api *DatabaseAPI)
```

Creates a db Lua table with all database operation functions and sets it as a global. Provided DatabaseAPI instance must be bound to exactly one LState and must not be shared. After calling RegisterDBAPI, the caller should call FreezeModule(L, "db") to make the module read-only.

#### func RegisterLogAPI

```go
func RegisterLogAPI(L *lua.LState, pluginName string)
```

Creates a log Lua table with info, warn, error, and debug functions bound to utility.DefaultLogger. Plugin name is included as structured field on every log call so operators can trace log output back to originating plugin. Each Lua function signature: log.level(message, context_table). message is string required, context_table is table optional with key-value pairs appended as structured args.

#### func RegisterHTTPAPI

```go
func RegisterHTTPAPI(L *lua.LState, pluginName string)
```

Creates an http global Lua table with handle and use functions. Also creates three hidden global tables used by HTTPBridge to read registered routes at load time: __http_handlers maps "METHOD /path" to handler LFunction, __http_route_meta maps "METHOD /path" to metadata table with public flag, __http_middleware is ordered array of middleware LFunctions. These tables are part of global snapshot taken by SnapshotGlobals and must NOT be modified after snapshot. Bridge reads them on every request dispatch to look up correct handler for checked-out VM.

#### func BuildLuaRequest

```go
func BuildLuaRequest(L *lua.LState, r *http.Request, clientIP string) (*lua.LTable, error)
```

Converts http.Request into Lua table suitable for passing to plugin handler function. clientIP parameter is proxy-aware IP extracted by extractClientIP with no port. Returned table fields: method string, path string from r.URL.Path, body string from raw body bytes always set regardless of Content-Type, client_ip string from provided parameter, headers table with all request headers keys normalized to lowercase, query table from r.URL.Query flattened to first value per key, params table with path parameters from r.Pattern via r.PathValue, json field with parsed JSON body ONLY when Content-Type starts with application/json. Returns nil and error if body cannot be read like http.MaxBytesError from body size enforcement. Caller is responsible for writing 400 response.

#### func WriteLuaResponse

```go
func WriteLuaResponse(w http.ResponseWriter, L *lua.LState, responseTbl *lua.LTable, maxRespSize int64, requestID string) error
```

Reads Lua response table and writes HTTP response. Response table expected fields: status number defaults to 200, headers table optional, body string optional, json any optional takes priority over body if both present. Enforces response size limits, filters blocked headers, sets default security headers. Returns error if response exceeds maxRespSize. Security headers always present: X-Content-Type-Options nosniff, X-Frame-Options DENY, Cache-Control no-store. Plugin headers filtered against blockedResponseHeaders list. JSON responses get Content-Type application/json.

## Database Operations

The db module provides query, query_one, count, exists, insert, update, delete, transaction, ulid, timestamp, and define_table functions to Lua plugins. All table names are automatically prefixed with plugin_pluginname_ and validated. Operations are counted per VM checkout with configurable limit.

#### db.query

`db.query(table, opts) -> sequence table of row tables | nil, errmsg`

Queries rows from plugin table. opts fields: where table optional, order_by string optional, limit number optional defaults to 100, offset number optional. Returns empty table on no matches, nil plus errmsg on error. Empty or nil where allowed returns all rows up to limit.

#### db.query_one

`db.query_one(table, opts) -> row table | nil | nil, errmsg`

Queries single row from plugin table. opts fields: where table optional, order_by string optional. Returns nil on no match, nil plus errmsg on error. Empty or nil where allowed returns arbitrary row.

#### db.count

`db.count(table, opts) -> integer | nil, errmsg`

Counts rows in plugin table. opts fields: where table optional. Empty or nil where allowed returns total row count.

#### db.exists

`db.exists(table, opts) -> boolean | nil, errmsg`

Checks if any row exists in plugin table. opts fields: where table optional. Empty or nil where allowed returns true if table has any rows.

#### db.insert

`db.insert(table, values) -> nil | nil, errmsg`

Inserts row into plugin table. Auto-sets id as ULID, created_at and updated_at as RFC3339 UTC if not provided. Explicit values never overridden. Returns nothing on success, nil plus errmsg on error.

#### db.update

`db.update(table, opts) -> nil | nil, errmsg`

Updates rows in plugin table. opts fields: set table required non-empty, where table required non-empty. Auto-sets updated_at in set map if not provided. Empty where raises error for safety prevents full-table update. Returns nothing on success, nil plus errmsg on error.

#### db.delete

`db.delete(table, opts) -> nil | nil, errmsg`

Deletes rows from plugin table. opts fields: where table required non-empty. Empty where raises error for safety prevents full-table delete. Returns nothing on success, nil plus errmsg on error.

#### db.transaction

`db.transaction(fn) -> true, nil | false, errmsg`

Executes fn inside a database transaction. Nested transactions rejected with error. Inside callback, all db calls automatically route through the transaction. Returns true and nil on commit, false and errmsg on rollback.

#### db.ulid

`db.ulid() -> string`

Generates a new ULID using thread-safe types.NewULID function. Returns 26-character string.

#### db.timestamp

`db.timestamp() -> string`

Returns current time as RFC3339 UTC string. Replaces os.date which is sandboxed out.

#### db.define_table

`db.define_table(tableName, definition) -> nil | raises error`

Creates plugin table with auto-injected id, created_at, updated_at columns. definition fields: columns sequence required non-empty with each entry having name string, type string, not_null bool optional, default string or number optional, unique bool optional. indexes sequence optional with each entry having columns sequence and unique bool. foreign_keys sequence optional with each entry having column string, ref_table string, ref_column string, on_delete string optional. Reserved column names id, created_at, updated_at rejected. Column types validated via db.ValidateColumnType. Foreign key ref_table must start with same plugin prefix for namespace isolation. Executes DDL with IfNotExists true.
