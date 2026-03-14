# Plugin Pipeline System Plan

## Implementation Status (2026-02-27)

The plugin system has been built through the original 4-phase roadmap plus additional work. This plan describes the complete vision. Sections are marked with implementation status:

- **IMPLEMENTED** -- feature exists in code, fully functional
- **IMPLEMENTED (different design)** -- equivalent feature exists but with different architecture than originally planned
- **NOT IMPLEMENTED** -- feature does not exist yet

### What's Built

| Component | Status | Source Files |
|-----------|--------|-------------|
| Core engine (VM pool, sandbox, manager) | IMPLEMENTED | `manager.go`, `pool.go`, `sandbox.go` |
| Database API (full query builder) | IMPLEMENTED | `db_api.go` |
| Schema definition + drift detection | IMPLEMENTED | `schema_api.go` |
| HTTP routes + approval workflow | IMPLEMENTED | `http_api.go`, `http_bridge.go`, `http_request.go` |
| Content hooks + approval workflow | IMPLEMENTED | `hooks_api.go`, `hook_engine.go`, `hook_handlers.go` |
| Circuit breaker (instance + hook level) | IMPLEMENTED | `recovery.go` |
| Metrics | IMPLEMENTED | `metrics.go` |
| Hot reload (blue-green, file watcher) | IMPLEMENTED | `watcher.go` |
| CLI (list, init, validate, info, reload, enable, disable, approve, revoke) | IMPLEMENTED | `cmd/plugin.go`, `cli_commands.go` |
| Plugin validator | IMPLEMENTED | `validator.go` |
| Lua helpers | IMPLEMENTED | `lua_helpers.go` |
| Log API | IMPLEMENTED | `log_api.go` |
| `plugin_routes` table (dynamic) | IMPLEMENTED | `http_bridge.go` |
| `plugin_hooks` table (dynamic) | IMPLEMENTED | `hook_engine.go` |
| `plugins` table (persistent registry) | NOT IMPLEMENTED | -- |
| `pipelines` table (DB-backed config) | NOT IMPLEMENTED | -- |
| Three-state lifecycle | NOT IMPLEMENTED | -- |
| Core Table Gateway | NOT IMPLEMENTED | -- |
| Distributed coordination | NOT IMPLEMENTED | -- |
| Pipeline dry-run | NOT IMPLEMENTED | -- |
| Manifest change detection | NOT IMPLEMENTED | (schema drift exists) |
| TUI screens for plugin management | IMPLEMENTED | `internal/plugin/ui_bridge.go`, `ui_primitives.go`, `ui_renderer.go`, `ui_api.go`, `ui_pool.go`; `internal/tui/screen_plugin_tui.go`, `bubble_plugin.go`, `overlay_plugin.go` |

### Architectural Differences from Plan

The plan envisioned hooks as DB-backed "pipeline" entries configurable at runtime. The implementation uses code-defined hooks registered at module scope via `hooks.on()`, with DB-backed approval status in a `plugin_hooks` table. This is simpler and already functional, but lacks runtime reconfiguration without code changes.

The plan envisioned a `plugins` table for persistent lifecycle state. The implementation uses filesystem discovery with in-memory state only. Plugin routes and hooks have DB-backed approval, but plugin enabled/disabled state is not persisted across restarts.

---

## Design Philosophy

ModulaCMS is a developer-first CMS. Plugin functionality is explicit, auditable, and fully controlled by the developer. Plugins never silently inject behavior. The developer knows WHAT is happening, WHEN, WHERE, and never has to ask WHY.

This stands in contrast to WordPress-style hook systems where plugins insert themselves into the core data path via `add_filter`/`add_action`, creating implicit behavior chains that are difficult to debug and reason about.

---

## Core Concepts

### Two-Tier Plugin Model

**Tier 1 -- Plugin Endpoints with Gateway Access** -- IMPLEMENTED (endpoints), NOT IMPLEMENTED (gateway)
For plugins that add new functionality (forms, e-commerce, analytics). These get their own HTTP endpoints under `/api/v1/plugins/{plugin-name}/` and can read/write core CMS tables through a scoped gateway interface. No chaining needed because they're doing their own thing.

Current state: HTTP endpoints are fully implemented. Gateway access to core CMS tables is not implemented -- plugins can only access their own `plugin_<name>_*` tables.

**Tier 2 -- Processing Pipelines** -- IMPLEMENTED (different design)
For plugins that extend core behavior (validation, sanitization, permissions, computed fields). These register processors on defined extension points. The CMS runs them as a pipeline during core operations. Chaining is built in.

Current state: Implemented as `hooks.on()` registration at module scope. Hooks are code-defined (not DB-configurable). Before-hooks run synchronously inside transactions, after-hooks run asynchronously post-commit. Priority-based ordering with wildcard support.

A plugin can participate in both tiers. A forms plugin might:
- Register `POST /api/v1/plugins/forms/submit/:id` (Tier 1 -- its own endpoint)
- Register a `before_create` processor on its own `plugin_forms_entries` table (Tier 2 -- validation pipeline)
- Use gateway access to read `content_fields` to know what fields to render (Tier 1 -- reading core tables) -- NOT YET AVAILABLE

### Separation of Concerns

1. **What a plugin can do** -- declared in the manifest (capabilities) -- PARTIALLY IMPLEMENTED (name, version, description, author, license, dependencies; no `capabilities` or `core_access` declarations)
2. **What's wired up** -- code-defined at module scope via `hooks.on()` and `http.handle()`, approval-gated by admin -- IMPLEMENTED
3. **What the CMS supports** -- extension points on audited content operations -- IMPLEMENTED

---

## Validation System (Field-Type-Scoped) -- NOT IMPLEMENTED

Field validation configs are stored as JSON in the `validation` column. The schema is scoped by field type, making invalid combinations structurally impossible.

### Structure

Common fields live at the top level. Type-specific constraints live in a nested object keyed by the field type. Only the matching section should be populated for a given field.

```json
{
  "required": true,
  "text": {
    "min_length": 3,
    "max_length": 255,
    "trim_whitespace": true
  }
}
```

```json
{
  "required": true,
  "number": {
    "min": 0,
    "max": 100,
    "step": 0.5
  }
}
```

```json
{
  "required": true,
  "relation": {
    "min_items": 1,
    "max_items": 5
  }
}
```

### Go Types

```go
type ValidationConfig struct {
    Required bool                `json:"required,omitempty"`
    Text     *TextValidation     `json:"text,omitempty"`
    Number   *NumberValidation   `json:"number,omitempty"`
    Date     *DateValidation     `json:"date,omitempty"`
    Select   *SelectValidation   `json:"select,omitempty"`
    Relation *RelationValidation `json:"relation,omitempty"`
    Media    *MediaValidation    `json:"media,omitempty"`
}

type TextValidation struct {
    MinLength      *int `json:"min_length,omitempty"`
    MaxLength      *int `json:"max_length,omitempty"`
    TrimWhitespace bool `json:"trim_whitespace,omitempty"`
}

type NumberValidation struct {
    Min  *float64 `json:"min,omitempty"`
    Max  *float64 `json:"max,omitempty"`
    Step *float64 `json:"step,omitempty"`
}
```

### Why Field-Type-Scoped

- `FieldType` enum already defines the universe of types. One validation struct per type maps cleanly.
- Third-party admin panels can read the config and know exactly which constraints apply without guessing.
- Prevents nonsensical configs (no `min_length` on a boolean) at the struct level rather than runtime.
- Extends the existing `RelationConfig` pattern already in the codebase.

---

## Plugin Lifecycle -- IMPLEMENTED (different design)

### Current Implementation (5-State Model)

Plugins use a 5-state model driven by filesystem discovery and automatic loading:

| State | Value | Description |
|-------|-------|-------------|
| Discovered | 0 | Directory found, manifest extracted, not yet loaded |
| Loading | 1 | Dependencies validated, VM pool creation in progress |
| Running | 2 | on_init succeeded, fully operational |
| Failed | 3 | Load or runtime error (reason stored in FailedReason) |
| Stopped | 4 | on_shutdown has run or admin disabled |

All discovered plugins are loaded automatically on startup. There is no separate "install" approval step.

### Planned Three-State Model -- NOT IMPLEMENTED

The original plan called for an explicit install step separating discovery from activation:

**Discovered** -- files exist in the plugins directory. The CMS sees them. Nothing runs. No code from the plugin has been executed, not even to read the manifest.

**Installed** -- developer explicitly runs `modulacms plugin install <name>`. The CMS reads the manifest, validates it, shows the approval screen (capabilities, requested pipelines, core access). Developer approves. Plugin gets a row in the `plugins` table with `status = installed`.

**Enabled** -- developer runs `modulacms plugin enable <name>`. VMs spin up, `on_init` fires, tables are created, endpoints register. Pipeline entries (if approved during install) become active.

### Planned Install Flow -- NOT IMPLEMENTED

```
$ modulacms plugin install field-validator

Reading manifest from plugins/field-validator/init.lua...

  field-validator v1.2.0
  "Content field validation with configurable rules"
  by: Example Corp

  Capabilities requested:
    Pipelines:
      content_fields   before_create   validate()   priority 10
      content_fields   before_update   validate()   priority 20

    Endpoints:
      GET  /api/v1/plugin/field-validator/rules
      POST /api/v1/plugin/field-validator/rules

    Core table access:
      content_fields   read
      datatypes        read

  Install? [y/N]: y

  Installed. Run 'modulacms plugin enable field-validator' to activate.
  Recommended pipelines will be created on enable.
  Adjust with 'modulacms pipelines configure'.
```

### Plugin Manifest -- IMPLEMENTED (partial)

Current manifest supports: `name`, `version`, `description`, `author`, `license`, `min_cms_version`, `dependencies`.

Not implemented: `capabilities` (pipeline declarations) and `core_access` (core table permissions).

```lua
plugin_info = {
    name = "field-validator",
    version = "1.2.0",
    description = "Content field validation with configurable rules",
    author = "Example Corp",
    dependencies = {},

    -- NOT YET SUPPORTED:
    capabilities = {
        {table = "content_fields", op = "before_create", handler = "validate", priority = 10},
        {table = "content_fields", op = "before_update", handler = "validate", priority = 20},
    },

    core_access = {
        content_fields = {"read"},
        datatypes = {"read"},
    }
}
```

### Manifest Change Detection -- NOT IMPLEMENTED

Schema drift detection exists (compares DDL definitions against actual table columns). Capability drift detection (new hooks, routes, core access changes) does not exist.

The plan called for:

```
$ modulacms plugin list

NAME               STATUS      VERSION  PIPELINES  ENDPOINTS  NOTES
field-validator    enabled     1.2.0    2 active   1
analytics          enabled     1.0.0    3 active   2          ! manifest changed
```

```
$ modulacms plugin inspect analytics

Manifest changes detected (installed: v0.9.0, current: v1.0.0):

  Added capabilities:
    + content_data   before_create   (NEW)
    + users          read            (NEW)

  Run 'modulacms plugin update analytics' to review and approve.
```

**Current behavior**: When the plugin version string changes, all route and hook approvals are revoked automatically, forcing re-approval. This provides a safety gate against silent capability changes, but does not show a diff of what changed.

---

## Pipeline System -- IMPLEMENTED (different design)

### Extension Points -- IMPLEMENTED

The CMS defines a fixed set of extension points on content operations. These are the only places plugins can participate.

| Extension Point  | Phase          | Can Modify Data | Can Reject Write |
|------------------|----------------|-----------------|------------------|
| `before_create`  | in transaction | yes             | yes              |
| `after_create`   | post-commit    | no              | no               |
| `before_update`  | in transaction | yes             | yes              |
| `after_update`   | post-commit    | no              | no               |
| `before_delete`  | in transaction | no              | yes              |
| `after_delete`   | post-commit    | no              | no               |
| `before_publish` | in transaction | yes             | yes              |
| `after_publish`  | post-commit    | no              | no               |
| `before_archive` | in transaction | yes             | yes              |
| `after_archive`  | post-commit    | no              | no               |

`before_*` hooks run inside the database transaction. They can modify the data being written or reject the operation entirely. If a before hook rejects, the transaction rolls back and the client gets an error.

`after_*` hooks run after the transaction commits. They receive a copy of the data but cannot modify it. They are fire-and-forget, suitable for analytics, notifications, cache invalidation, search indexing.

### Current Implementation: Code-Defined Hooks with DB-Backed Approval

Hooks are registered at module scope in `init.lua` via `hooks.on()`:

```lua
hooks.on("before_create", "content_data", function(data)
    if not data.title then error("title required") end
    return data
end, { priority = 50 })

hooks.on("after_update", "content_data", function(data)
    log.info("Content updated", { id = data.id })
end)
```

Hook registrations are stored in a `plugin_hooks` table (created dynamically by the hook engine) with an `is_approved` column. Unapproved hooks are silently skipped. Admin approves/revokes via CLI or API.

**Hook engine architecture** (`hook_engine.go`):
- In-memory index: `"event:table" -> sorted []hookEntry`
- O(1) fast-path gate: `hasHooks` map + `hasAnyHook` bool
- Sort order: priority (lower first), specific before wildcard, registration order
- Per-hook circuit breaker: consecutive abort tracking per (plugin, event, table)
- Before-hooks: synchronous, inside CMS transaction, `db.*` calls BLOCKED (SQLite deadlock prevention)
- After-hooks: async goroutines, bounded concurrency (default 10), reduced op budget (default 100)

### Planned: DB-Backed Pipeline Configuration -- NOT IMPLEMENTED

The original plan called for pipeline definitions stored in a database table, configurable at runtime:

```sql
CREATE TABLE pipelines (
    pipeline_id TEXT PRIMARY KEY NOT NULL,
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    plugin_name TEXT NOT NULL,
    handler TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 50,
    enabled INTEGER NOT NULL DEFAULT 1,
    config TEXT NOT NULL DEFAULT '{}',
    date_created TEXT NOT NULL,
    date_modified TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_pipeline_unique
    ON pipelines(table_name, operation, plugin_name);
```

#### Plugins Table -- NOT IMPLEMENTED

```sql
CREATE TABLE plugins (
    plugin_id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL UNIQUE,
    version TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'installed',
    capabilities TEXT NOT NULL DEFAULT '{}',
    approved_access TEXT NOT NULL DEFAULT '{}',
    date_installed TEXT NOT NULL,
    date_modified TEXT NOT NULL
);
```

### In-Memory Registry -- IMPLEMENTED (different design)

The plan called for a `PipelineRegistry` built from the `pipelines` DB table:

```go
type PipelineRegistry struct {
    mu     sync.RWMutex
    chains map[string][]PipelineEntry  // "content_data.before_create" -> sorted entries
}
```

The current implementation uses `HookEngine` with an equivalent in-memory index built from code-defined registrations:

```go
// hook_engine.go
hookIndex map[string][]hookEntry  // "event:table" -> sorted entries
hasHooks  map[string]bool          // fast-path gate
hasAnyHook bool                    // zero-allocation O(1) check
```

Functionally equivalent. The difference is the source: code-defined hooks vs DB-backed pipelines.

### Audited Command Integration -- IMPLEMENTED

The hook engine implements the `audited.HookRunner` interface, integrating directly into the audited command layer. Every audited write checks for registered hooks at the appropriate extension points.

### Pipeline Execution -- IMPLEMENTED

Processors run sequentially within a phase. Each processor is a Lua VM checkout from the respective plugin's pool.

```
Request -> Auth -> Core handler -> Audited command begins transaction
  -> Run before_update hooks:
      -> checkout validator VM -> handler(data) -> return
      -> checkout sanitizer VM -> handler(data) -> return
  -> Execute write
  -> Run after_update hooks (async, post-commit)
  -> Response
```

If a before-hook calls `error()`, the pipeline aborts and the transaction rolls back. The client gets a clear error.

### Audit Identity -- NOT IMPLEMENTED

When a plugin writes to core tables through a pipeline or gateway, the audit trail should capture both the human actor and the plugin involvement.

```go
type AuditContext struct {
    UserID    types.UserID
    PluginID  string  // empty for core operations
    Operation string
}
```

### Pipeline Introspection -- NOT IMPLEMENTED

```
$ modulacms pipelines list

TABLE              OPERATION        PLUGIN           HANDLER      PRI  ENABLED
content_fields     before_create    field-validator   validate     10   yes
content_fields     before_update    field-validator   validate     10   yes
content_fields     before_update    sanitizer         sanitize     20   yes
content_data       after_create     analytics         track        10   yes
content_data       after_update     analytics         track        10   yes
```

```
$ modulacms pipelines show content_fields

before_create:
  1. field-validator.validate  (priority 10)

before_update:
  1. field-validator.validate  (priority 10)
  2. sanitizer.sanitize        (priority 20)

after_create:   (none)
after_update:   (none)
before_delete:  (none)
after_delete:   (none)
```

**Current equivalent**: Hooks can be listed via `GET /api/v1/admin/plugins/hooks` which returns all registered hooks with approval status, event, table, plugin name, and priority.

### Pipeline Execution Logging -- IMPLEMENTED (partial)

Hook execution timing is logged via slog at debug level with plugin name, event, table, and duration. The formatted pipeline-chain output shown below is not implemented:

```
[2026-02-11 14:32:05] pipeline content_fields.before_update
  -> field-validator.validate   2.1ms  pass  (data modified: false)
  -> sanitizer.sanitize         0.8ms  pass  (data modified: true, fields: [value])
  -> audited write              1.2ms  ok
  total: 4.1ms
```

### Pipeline Dry-Run -- NOT IMPLEMENTED

```
$ modulacms pipelines test content_fields before_update \
    --data '{"value": "test <script>alert(1)</script>"}'

Pipeline: content_fields.before_update
Input:  {"value": "test <script>alert(1)</script>"}

  -> field-validator.validate   pass  (no modifications)
  -> sanitizer.sanitize         pass  modified: {"value": "test "}

Output: {"value": "test "}
(dry run -- no data written)
```

### Validation at Wire-Up Time -- NOT IMPLEMENTED

When a pipeline entry is created (via TUI or CLI), the CMS validates:

1. Plugin exists and is loaded
2. Capability match -- the plugin must declare that it handles this table/op combination (or `*` wildcard)
3. Handler exists -- the Lua function must be defined in the plugin's VM
4. No duplicate -- unique constraint prevents double-registration
5. Priority conflicts -- warn if two processors share the same priority on the same extension point

### Wildcard Policy -- IMPLEMENTED

Plugins can register `table = "*"` in `hooks.on()` for wildcard hooks. Wildcard hooks run after specific-table hooks at equal priority.

---

## Core Table Gateway (Tier 1 -- Scoped Access) -- NOT IMPLEMENTED

Plugins that need to interact with core CMS tables (not just their own `plugin_*` tables) do so through a scoped gateway. Access is declared in the manifest and approved at install time.

### Manifest Declaration

```lua
core_access = {
    content_fields = {"read", "update"},
    datatypes = {"read"},
}
```

### Lua API

Only approved operations are exposed. Unapproved access raises an error.

```lua
-- inside a plugin HTTP handler
function handle_validated_write(req)
    local field = core.content_fields.get(req.params.id)    -- works (read granted)

    local ok, err = validate_against_rules(field, req.body)
    if not ok then
        return {status = 400, body = {error = err}}
    end

    core.content_fields.update(req.params.id, {              -- works (update granted)
        data = req.body.data
    })

    return {status = 200}
end

-- core.content_fields.delete(id)  -- ERROR: not granted
-- core.users.list()               -- ERROR: no access to users table
```

Under the hood, `core.content_fields.update` goes through the audited command infrastructure. Change events, audit logs, and pipeline processing all apply as if the write came from a core endpoint.

---

## Distributed Deployment -- NOT IMPLEMENTED

### DB as Coordination Layer

Plugin state and pipeline configuration live in the database. All running instances derive their runtime behavior from the same DB. No config file sync needed.

- Instance A enables a plugin -> writes to `plugins` and `pipelines` tables
- Instance B detects the change on next poll -> loads VMs, activates pipelines
- Instance C detects the change on next poll -> same

### Change Detection

Each instance polls `plugins` and `pipelines` tables on an interval (e.g. 5 seconds). When changes are detected:

1. If a plugin was enabled and files exist locally: spin up VM pool, register endpoints, activate pipelines
2. If a plugin was disabled: shut down VM pool, deregister endpoints, deactivate pipelines
3. If a plugin was enabled but files are missing locally: trigger graceful self-eviction

### Self-Eviction for Container Deployments

When an instance detects a plugin is enabled but the plugin files are not present on the local filesystem:

```go
func (w *PluginWatcher) onPluginChange(p PluginRecord) {
    if p.Status == "enabled" && !w.pluginFilesExist(p.Name) {
        w.healthStatus.Store(false)  // health endpoint returns 503
        w.logger.Warn("plugin files missing, draining before shutdown",
            "plugin", p.Name,
        )
        time.Sleep(w.drainDuration)  // wait for LB to stop sending traffic
        w.shutdownCh <- syscall.SIGTERM
    }
}
```

The deployment sequence:

1. Push new container image with plugin files included
2. Run `modulacms plugin install && modulacms plugin enable` on any running instance
3. Instances without the files detect the mismatch
4. Health check fails -> load balancer stops routing traffic to that instance
5. Instance drains in-flight requests, then exits
6. Orchestrator (k8s, ECS, etc.) sees instance down, starts replacement from new image
7. New instance boots with plugin files, reads DB, sees plugin is enabled, loads everything
8. Health check passes -> load balancer adds it back

Go binary boot time is milliseconds. The window between "DB says enabled" and "new instance is ready" is negligible.

### What Each Instance Manages

**In the DB (shared):** Plugin status, pipeline definitions, approved capabilities. Source of truth.

**In memory (per-instance):** VM pools, pipeline registry map, endpoint registrations. Derived state, rebuilt from DB.

**On disk (per-instance):** Plugin Lua files. Deployed with the container image.

### Concurrent Schema Creation

`db.define_table` uses `CREATE TABLE IF NOT EXISTS`. Multiple instances running `on_init` concurrently is safe -- all try to create the same table, only one succeeds, the rest no-op.

---

## User Scenarios

### User A -- No Plugins -- IMPLEMENTED

- Plugin manager starts, finds no plugins in directory
- Hook engine has empty index
- Every write does a map lookup, gets nil, continues normally
- Overhead: effectively zero

### User B -- Plugin Endpoints Only (Sandboxed Tables) -- IMPLEMENTED

- Plugin loads, creates `plugin_forms_entries` table
- Registers `POST /api/v1/plugins/forms/submit/:id`
- No hooks registered
- Core writes are completely unaffected
- Plugin reads/writes only its own namespaced tables

### User C -- Analytics Plugin (After-Hooks) -- IMPLEMENTED

- Plugin loads, creates `plugin_analytics_events` table
- Registers `GET /api/v1/plugins/analytics/events`
- Registers after-hooks via `hooks.on()`:
  - `after_create` on `content_data`
  - `after_update` on `content_data`
- Every content write fires the after-hook, analytics plugin writes to its own table
- Core write performance unaffected (after-hooks are async, post-commit)
- Admin panel fetches analytics via the plugin's GET endpoint

### User D -- Composed Pipeline (Multiple Plugins) -- IMPLEMENTED

- Validator plugin: registers `before_create`, `before_update` hooks on `content_data`
- Sanitizer plugin: registers `before_create`, `before_update` hooks on `content_data`
- Analytics plugin: registers `after_*` hooks with wildcard table
- Hooks ordered by priority (lower first, specific before wildcard)

Data flow: request -> validator -> sanitizer -> audited write -> analytics. Each step is a Lua VM checkout from the respective plugin's pool. If the validator rejects, the write never happens and the client gets a 422.

---

## Comparison with WordPress -- UPDATED

| Aspect                        | WordPress                                      | ModulaCMS (current)                               | ModulaCMS (planned)                               |
|-------------------------------|------------------------------------------------|---------------------------------------------------|---------------------------------------------------|
| Hook registration             | Plugin code calls `add_filter()` at load time  | Plugin code calls `hooks.on()` at module scope    | Plugin declares capabilities, developer wires pipelines |
| Who controls what runs        | Plugin authors                                 | Plugin code + admin approval gate                 | CMS developer/admin (DB-configurable)             |
| Removing a hook               | Edit plugin code or deactivate entirely        | Revoke approval via CLI/API                       | Disable one pipeline entry                         |
| Debugging "what runs on save" | grep codebase for `add_filter('save_post')`    | `GET /api/v1/admin/plugins/hooks`                 | `modulacms pipelines show content_data`            |
| Performance when no plugins   | Still loads hook infrastructure                | Map lookup returns nil                             | Map lookup returns nil                             |
| Bad plugin impact             | Runs in main PHP thread, can take down site    | Timeout kills VM, circuit breaker trips           | Same                                               |
| Chaining                      | Implicit via priority in `add_filter`          | Explicit via priority in `hooks.on()`             | Explicit via pipeline priority, visible in config  |
| Discovery                     | Filesystem scan, activate with one click       | Filesystem discovery, auto-load, approval gate    | Filesystem discovery, explicit install and enable   |
| Plugin updates                | Silent capability changes                       | Version change revokes all approvals              | Manifest diff, developer approval required         |
| Distributed coordination      | Not applicable (single-process PHP)            | File-based hot reload per instance                | DB-backed state, automatic instance coordination   |

---

## Implementation Phases -- STATUS

This plan builds on the existing Phase 1 (core engine) and replaces/refines the original Phase 2-4 roadmap items.

### Phase 2A: Plugin Lifecycle and Pipelines Schema -- NOT IMPLEMENTED
- Add `plugins` table and `pipelines` table schemas
- Implement three-state lifecycle (discovered/installed/enabled)
- CLI commands: `plugin install`, `plugin inspect`
- Pipeline registry (in-memory, built from DB)
- Pipeline CLI commands: `pipelines list`, `pipelines show`, `pipelines configure`

### Phase 2B: Pipeline Execution -- IMPLEMENTED
- Integrate pipeline checks into audited command layer (via `audited.HookRunner` interface)
- Before-hook execution (in-transaction, synchronous, can reject via `error()`)
- After-hook execution (post-commit, async, fire-and-forget, bounded concurrency)
- Hook execution logging with timing (slog debug level)
- Timeout enforcement per hook (2s per hook, 5s per event chain)
- Per-hook circuit breaker (10 consecutive aborts disables hook)

### Phase 2C: HTTP Integration -- IMPLEMENTED
- Plugin endpoint registration via `http.handle(method, path, handler)`
- All routes namespaced under `/api/v1/plugins/{plugin_name}/`
- Request/response marshaling (`BuildLuaRequest`/`WriteLuaResponse`)
- Pool exhaustion -> HTTP 503
- Route approval workflow (DB-backed `plugin_routes` table)
- Rate limiting (per-IP token bucket)
- Middleware support via `http.use()`
- Body/response size limits
- Security headers (X-Content-Type-Options, X-Frame-Options)
- Blocked response headers (11 entries)
- Graceful shutdown with inflight request drain

### Phase 2D: Core Table Gateway -- NOT IMPLEMENTED
- Scoped access API (`core.{table}.get`, `core.{table}.update`, etc.)
- Access enforcement based on approved capabilities
- Writes go through audited command infrastructure
- Audit identity includes plugin provenance

### Phase 3: Distributed Coordination -- NOT IMPLEMENTED
- Plugin watcher (polls DB for state changes)
- Hot pipeline reconfiguration (rebuild registry without restart)
- Self-eviction for missing plugin files
- Health endpoint integration for load balancer draining

### Phase 4: Production Hardening -- MOSTLY IMPLEMENTED

| Feature | Status |
|---------|--------|
| Circuit breaker (N failures -> temp disable) | IMPLEMENTED |
| Per-plugin metrics (execution time, error rate) | IMPLEMENTED |
| Hot reload (blue-green, file watcher) | IMPLEMENTED |
| CLI commands (list, init, validate, info, reload, enable, disable, approve, revoke) | IMPLEMENTED |
| Admin API endpoints (13 endpoints) | IMPLEMENTED |
| Pipeline dry-run mode | NOT IMPLEMENTED |
| Manifest change detection (capability drift) | NOT IMPLEMENTED (schema drift exists) |
| Priority conflict warnings | NOT IMPLEMENTED |
| TUI screens for plugin and pipeline management | IMPLEMENTED (plugin TUI coroutine bridge, screens, field interfaces) |
