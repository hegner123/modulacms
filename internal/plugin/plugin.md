# plugin

ModulaCMS Lua plugin system enabling runtime-extensible database tables, queries, and CMS integration. Provides sandboxed Lua VM pools, a query builder API, structured logging, and lifecycle management.

## Overview

The plugin system architecture consists of Manager for discovery and lifecycle, VMPool for concurrent VM management, sandboxed stdlib with db and log modules, and isolated table namespacing. Plugins define tables with prefix enforcement, execute queries via a builder API, and integrate with CMS via lifecycle hooks.

All plugin tables use the prefix plugin_pluginname_. The query builder validates identifiers, preventing SQL injection and namespace violations. Plugins cannot access core CMS tables or other plugins' tables. VM pooling enables concurrent execution with per-checkout operation budgets and health validation.

## Constants

ErrPoolExhausted signals VM pool exhaustion after 100ms timeout. Callers treat this as backpressure. HTTP bridge translates to 503 Service Unavailable.

ErrOpLimitExceeded signals per-checkout operation budget exhaustion. Check via errors.Is(). Budget resets on next VMPool.Get() via ResetOpCount().

acquireTimeout is 100 milliseconds for VM acquisition. Provides backpressure under load rather than tying up goroutines for full execution timeout.

## Types

### PluginState

PluginState represents lifecycle state using iota enum. States are StateDiscovered, StateLoading, StateRunning, StateFailed, StateStopped.

#### String

String() returns human-readable state name. Returns unknown(N) for invalid values.

### PluginInfo

PluginInfo holds manifest metadata extracted from plugin_info global. Contains Name, Version, Description, Author, License, MinCMSVersion, Dependencies slice.

### PluginInstance

PluginInstance represents loaded plugin with state and resources. Fields include Info, State, Dir as absolute path, InitPath to init.lua, FailedReason for human-readable failure, Pool, and dbAPIs map linking each LState to its DatabaseAPI for op count reset.

### ManagerConfig

ManagerConfig controls runtime behavior. Enabled bool, Directory string, MaxVMsPerPlugin default 4, ExecTimeoutSec default 5, MaxOpsPerExec default 1000 per VM checkout.

### Manager

Manager coordinates plugin discovery, loading, lifecycle, and shutdown. No stored context field. All I/O methods accept ctx as first parameter. Fields include cfg, db as separate pool via db.OpenPool(), dialect, plugins map, mu sync.RWMutex, loadOrder preserving topologically sorted successful loads for reverse-order shutdown.

#### NewManager

NewManager creates Manager with config and DB pool. Zero-value fields replaced with defaults. MaxVMsPerPlugin default 4, ExecTimeoutSec default 5, MaxOpsPerExec default 1000.

#### LoadAll

LoadAll discovers plugins in cfg.Directory, validates manifests, resolves dependencies via topological sort, loads each in dependency order. Failed plugins marked StateFailed without preventing others.

Loading sequence per plugin: scan for init.lua subdirectories, create temp sandboxed VM to extract and validate plugin_info, topologically sort by dependencies detecting cycles, create VMPool for each in dependency order, run on_init, snapshot globals.

#### GetPlugin

GetPlugin returns instance by name or nil if not found. Thread-safe via read lock.

#### ListPlugins

ListPlugins returns all loaded instances. Thread-safe via read lock.

#### Shutdown

Shutdown gracefully shuts down all plugins in reverse dependency order. For each plugin checks out VM, calls on_shutdown if defined, returns VM. After all on_shutdown calls closes all VM pools and plugin DB pool.

### VMPool

VMPool provides goroutine-safe pool of pre-initialized LState instances for single plugin. Uses buffered channel for lock-free checkout and return. Validates VM health on every Put.

Lifecycle: NewVMPool creates pool and fills with factory VMs. Get(ctx) checks out VM setting caller's context. Put(L) clears stack, restores globals to post-init state, validates health, returns to pool or replaces if corrupted. Close() drains all VMs preventing further returns.

Thread safety via channel synchronization. initGlobals map written once by SnapshotGlobals and read on every Put. Caller ensures SnapshotGlobals called before concurrent Put.

Fields include states channel, factory function, size int, initPath and pluginName for diagnostics, closed atomic bool, initGlobals map for snapshot after on_init.

#### NewVMPool

NewVMPool creates pool of size VMs using factory. Factory must produce fully sandboxed VMs with ApplySandbox, RegisterPluginRequire, RegisterDBAPI, RegisterLogAPI, FreezeModule applied. Global snapshot NOT taken here, happens after on_init in manager.

#### Get

Get checks out VM from pool with 100ms acquisition timeout for backpressure. On success sets caller's context via L.SetContext(ctx) for execution timeout. Caller MUST call Put(L) when done even on Lua execution failure. Caller should call DatabaseAPI.ResetOpCount() after Get() to reset per-checkout operation budget. Returns ErrPoolExhausted if no VM available.

#### Put

Put returns VM to pool after use. Steps: clear Lua stack, restore global snapshot removing any created after on_init, validate VM health verifying db and log modules intact, return to channel if healthy and pool not closed, close VM directly if healthy and pool closed, close and replace if unhealthy.

Put never blocks. If channel full VM closed with warning logged.

#### Close

Close drains all VMs from pool and prevents further returns. VMs currently checked out closed when returned via Put checking closed flag. Logs diagnostic if pool not fully drained indicating VMs checked out at shutdown.

#### SnapshotGlobals

SnapshotGlobals records all current global names on given VM. Must be called once after on_init on first VM. Snapshot shared across all pool VMs since all loaded same init.lua. After this call Put removes any global not in snapshot.

### SandboxConfig

SandboxConfig controls sandbox behavior. AllowCoroutine enables coroutine library, disabled by default. ExecTimeout as maximum wall-clock duration for single plugin execution via LState.SetContext(ctx), default 5s.

### DatabaseAPI

DatabaseAPI provides sandboxed db module for single plugin. INVARIANT: each instance bound to exactly one LState, never shared across VMs. 1:1 binding means no concurrent access to currentExec, inTx, or opCount.

Fields include conn, currentExec normally conn swapped to sql.Tx inside transactions, pluginName, dialect, maxRows default 100 hard cap 10000, maxTxOps default 10, inTx preventing nested transactions, opCount incremented on every db call reset on Get, maxOpsPerExec default 1000 configurable via ManagerConfig.

#### NewDatabaseAPI

NewDatabaseAPI creates instance bound to connection and plugin. INVARIANT: each instance bound to exactly one LState, never shared. maxOpsPerExec default 1000 if zero.

#### ResetOpCount

ResetOpCount resets per-checkout operation counter to zero. Called by Manager after VMPool.Get() before plugin code executes.

### TableDefinition

TableDefinition holds parsed and validated schema for plugin-defined table. Fields include PluginName, TableName without prefix, FullName as plugin_name_table, Columns slice, Indexes slice, ForeignKeys slice.

## Functions

### ApplySandbox

ApplySandbox configures VM with safe stdlib subset. Loads base, table, string, math libraries plus coroutine if cfg.AllowCoroutine. Strips dofile, loadfile, load, rawget, rawset, rawequal, rawlen. Never loads io, os, package, debug, channel. LState must have SkipOpenLibs true.

### RegisterPluginRequire

RegisterPluginRequire replaces global require with sandboxed loader. Resolves modules only from pluginDir/lib/name.lua. Module names must be simple identifiers, path traversal rejected. Loaded modules cached, subsequent require returns cached value. Uses L.ArgError for validation failures, L.RaiseError for load failures.

### FreezeModule

FreezeModule replaces global module table with read-only proxy. Real functions moved to hidden backing table. Proxy delegates reads via __index, rejects writes via __newindex, __metatable prevents inspection or replacement.

After freezing db.query(...) works via delegation, db.query = nil raises error, getmetatable(db) returns protected string, setmetatable(db, {}) raises error, pairs(db) returns nothing as documented DX limitation.

### RegisterDBAPI

RegisterDBAPI creates db Lua table with all database operations and sets as global. DatabaseAPI instance must be bound to exactly one LState, never shared. After calling should call FreezeModule(L, db) for read-only.

DB module provides query, query_one, count, exists, insert, update, delete, transaction, ulid, timestamp, define_table.

### RegisterLogAPI

RegisterLogAPI creates log Lua table with info, warn, error, debug functions bound to utility.DefaultLogger. Plugin name included as structured field on every log call. Each function signature log.level(message, context_table) where context_table optional key-value pairs.

### LuaTableToMap

LuaTableToMap converts Lua table with string keys to Go map. Non-string keys skipped. Nested tables recursively converted via LuaValueToGo.

### MapToLuaTable

MapToLuaTable converts Go map string any to Lua table. Values converted via GoValueToLua.

### GoValueToLua

GoValueToLua converts Go value to Lua value. Supported types: string to LString, int64 float64 int int32 to LNumber, bool to LBool, nil to LNil, bytes to LString, map string any to recursive MapToLuaTable, slice any to sequence table 1-indexed, slice map string any to sequence table of tables 1-indexed.

Unsupported types return LNil. Plugin code never sees Go-internal types. Silently converting to nil safer than panicking in production.

### LuaValueToGo

LuaValueToGo converts Lua value to Go value. LString to string, LNumber to float64, LBool to bool, LNil to nil, LTable to map string any if any string key or slice any if pure sequence.

For tables: any string keys converts to map including integer-keyed entries with string-converted keys. Only consecutive integer keys starting at 1 converts to slice any.

### RowsToLuaTable

RowsToLuaTable converts slice of db.Row results to Lua sequence table. Each row becomes Lua table. Empty slices return empty table never nil. Query result contract allows safe use of #results, ipairs(results), results[1].

Converts from already-scanned Go representation avoiding coupling to sql.Rows lifecycle management.

## Query Operations

Query functions accept table name and optional opts table. Table names auto-prefixed with plugin_pluginname_. Where clauses as key-value maps. Returns empty table on no matches, nil plus error message on error.

db.query(table, opts) returns sequence table of row tables. opts fields: where map, order_by string, limit int default 100, offset int. Empty where allowed returning all rows up to limit.

db.query_one(table, opts) returns single row table or nil. opts fields: where map, order_by string. Empty where allowed returning arbitrary row.

db.count(table, opts) returns integer. opts field: where map. Empty where allowed returning total row count.

db.exists(table, opts) returns boolean. opts field: where map. Empty where allowed returning true if table has any rows.

## Mutation Operations

db.insert(table, values) auto-sets id as ULID, created_at and updated_at as RFC3339 UTC if not provided. Explicit values never overridden. Returns nothing on success, nil plus error message on error.

db.update(table, opts) requires non-empty set map and non-empty where map. Empty where raises error preventing full-table update. Auto-sets updated_at in set if not provided. Returns nothing on success, nil plus error message on error.

db.delete(table, opts) requires non-empty where map. Empty where raises error preventing full-table delete. Returns nothing on success, nil plus error message on error.

db.transaction(fn) executes function in transaction. Nested transactions rejected. Inside callback all db calls route through sql.Tx via executor swap. Returns true nil on success, false error message on error.

## Utility Operations

db.ulid() returns string. Generates new ULID using thread-safe types.NewULID().

db.timestamp() returns string. Returns current time as RFC3339 UTC. Replaces sandboxed os.date().

db.define_table(tableName, definition) parses arguments, validates and prefixes table name, auto-injects id created_at updated_at columns, validates FK namespace isolation, executes DDLCreateTable with IfNotExists true.

Definition table fields: columns sequence required each with name type not_null default unique, indexes sequence optional each with columns sequence and unique bool, foreign_keys sequence optional each with column ref_table ref_column on_delete.

Reserved columns id created_at updated_at auto-injected, cannot be manually defined. FK namespace isolation requires ref_table start with plugin_pluginname_ prefix preventing reference to core CMS or other plugins.

## VM Factory Call Sequence

Factory produces fully sandboxed VMs following roadmap-specified order: ApplySandbox, RegisterPluginRequire, RegisterDBAPI, RegisterLogAPI, FreezeModule(db), FreezeModule(log). Execute init.lua to define globals. Clear context after factory init, caller sets via Get().

Store dbAPI mapping on instance for op count reset. VM usable even if init.lua partially executes, health check on Put catches critical corruption.

## Health Validation

validateVM checks db and log modules intact. Validates db global is LTable, db.query db.query_one db.count db.exists db.insert db.update db.delete db.transaction db.define_table all LFunction with IsG true, log global is LTable, log.info is LFunction with IsG true.

IsG true only for Go-bound functions. Catches replacement with pure Lua functions or non-function values without storing original pointers. Lightweight approximately 11 GetGlobal/GetField calls plus type assertions. No Lua execution, no allocations. Nanosecond-scale overhead per Put.

## Manifest Validation

extractManifest creates temp sandboxed VM, executes init.lua, reads plugin_info global, validates manifest fields. Temp VM discarded after extraction. Applies sandbox and RegisterPluginRequire so plugins using require at file scope can have manifest extracted. No db or log APIs during manifest extraction.

validateManifest checks required fields Name Version Description present and valid. Name max 32 chars, lowercase alphanumeric underscore only, no trailing underscore preventing prefix collisions.

## Dependency Resolution

topologicalSort orders plugins by dependency relationships using Kahn algorithm. Returns error if circular dependency detected. Builds adjacency list and in-degree count. Edge A to B means A depends on B, B must load before A.

Start with nodes having in-degree zero. Sort initial queue for deterministic ordering. Process dependents reducing in-degree. If not all plugins processed cycle exists, find participants for error message.
