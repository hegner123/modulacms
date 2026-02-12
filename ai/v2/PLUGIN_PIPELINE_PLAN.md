# Plugin Pipeline System Plan

## Design Philosophy

ModulaCMS is a developer-first CMS. Plugin functionality is explicit, auditable, and fully controlled by the developer. Plugins never silently inject behavior. The developer knows WHAT is happening, WHEN, WHERE, and never has to ask WHY.

This stands in contrast to WordPress-style hook systems where plugins insert themselves into the core data path via `add_filter`/`add_action`, creating implicit behavior chains that are difficult to debug and reason about.

---

## Core Concepts

### Two-Tier Plugin Model

**Tier 1 -- Plugin Endpoints with Gateway Access**
For plugins that add new functionality (forms, e-commerce, analytics). These get their own HTTP endpoints under `/api/v1/plugin/{plugin-name}/` and can read/write core CMS tables through a scoped gateway interface. No chaining needed because they're doing their own thing.

**Tier 2 -- Processing Pipelines**
For plugins that extend core behavior (validation, sanitization, permissions, computed fields). These register processors on defined extension points. The CMS runs them as a pipeline during core operations. Chaining is built in.

A plugin can participate in both tiers. A forms plugin might:
- Register `POST /api/v1/plugin/forms/submit/:id` (Tier 1 -- its own endpoint)
- Register a `before_create` processor on its own `plugin_forms_entries` table (Tier 2 -- validation pipeline)
- Use gateway access to read `content_fields` to know what fields to render (Tier 1 -- reading core tables)

### Separation of Concerns

1. **What a plugin can do** -- declared in the manifest (capabilities)
2. **What's wired up** -- configured by the developer (pipeline definitions)
3. **What the CMS supports** -- extension points exist on every table, cost nothing until activated

---

## Validation System (Field-Type-Scoped)

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

## Plugin Lifecycle

### Three-State Model

Plugins are discovered via filesystem but activated via explicit developer action.

**Discovered** -- files exist in the plugins directory. The CMS sees them. Nothing runs. No code from the plugin has been executed, not even to read the manifest.

**Installed** -- developer explicitly runs `modulacms plugin install <name>`. The CMS reads the manifest, validates it, shows the approval screen (capabilities, requested pipelines, core access). Developer approves. Plugin gets a row in the `plugins` table with `status = installed`.

**Enabled** -- developer runs `modulacms plugin enable <name>`. VMs spin up, `on_init` fires, tables are created, endpoints register. Pipeline entries (if approved during install) become active.

### Install Flow

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

### Plugin Manifest

Plugins declare capabilities in their manifest. The CMS validates these at install time. The developer controls what actually gets wired.

```lua
plugin_info = {
    name = "field-validator",
    version = "1.2.0",
    description = "Content field validation with configurable rules",
    author = "Example Corp",
    dependencies = {},

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

### Manifest Change Detection

The `plugins` table stores a manifest snapshot at install time. If a plugin updates and its capabilities change, the CMS detects the drift:

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

The plugin keeps running with its old approved access until the developer explicitly approves the new capabilities. No silent privilege escalation.

---

## Pipeline System

### Extension Points

The CMS defines a fixed set of extension points on content operations. These are the only places plugins can participate. Not arbitrary function hooks.

| Extension Point  | Phase          | Can Modify Data | Can Reject Write |
|------------------|----------------|-----------------|------------------|
| `before_create`  | in transaction | yes             | yes              |
| `after_create`   | post-commit    | no              | no               |
| `before_update`  | in transaction | yes             | yes              |
| `after_update`   | post-commit    | no              | no               |
| `before_delete`  | in transaction | no              | yes              |
| `after_delete`   | post-commit    | no              | no               |

`before_*` hooks run inside the database transaction. They can modify the data being written or reject the operation entirely. If a before hook rejects, the transaction rolls back and the client gets an error.

`after_*` hooks run after the transaction commits. They receive a copy of the data but cannot modify it. They are fire-and-forget, suitable for analytics, notifications, cache invalidation, search indexing.

### Pipeline Configuration (DB-Backed)

Pipeline definitions live in the database, not config files. This enables:
- Runtime configuration through TUI or CLI without restarts
- Auditable changes through the change event system
- Coordination across distributed instances

#### Schema

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

#### Plugins Table

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

### In-Memory Registry

At startup (and on pipeline config change), the manager builds an in-memory registry from the `pipelines` table:

```go
type PipelineRegistry struct {
    mu     sync.RWMutex
    chains map[string][]PipelineEntry  // "content_data.before_create" -> sorted entries
}

type PipelineEntry struct {
    PluginName string
    Handler    string
    Priority   int
    Config     map[string]any
}

func (r *PipelineRegistry) Before(table, op string) []PipelineEntry {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.chains[table+".before_"+op]
}

func (r *PipelineRegistry) After(table, op string) []PipelineEntry {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.chains[table+".after_"+op]
}
```

Map lookup under a read lock. When no pipelines exist, returns nil. Cost: nanoseconds.

### Audited Command Integration

The pipeline check integrates into the existing audited command layer. Every write already goes through `audited.Create`, `audited.Update`, `audited.Delete`.

```go
func Create[T any](cmd CreateCommand[T]) (T, error) {
    // ... begin transaction ...

    // Pipeline check -- map lookup, nil if nothing registered
    if processors := cmd.Registry().Before(cmd.TableName(), "create"); processors != nil {
        params, err := RunPipeline(cmd.Context(), processors, cmd.Params())
        if err != nil {
            // rollback
            return zero, fmt.Errorf("pipeline rejected: %w", err)
        }
        cmd.SetParams(params)
    }

    result, err := cmd.Execute(ctx, tx)
    // ... commit ...

    // After hooks -- fire and forget, outside transaction
    if processors := cmd.Registry().After(cmd.TableName(), "create"); processors != nil {
        go RunPipelineAsync(context.Background(), processors, result)
    }

    return result, nil
}
```

### Pipeline Execution

Processors run sequentially within a phase. Each processor is a Lua VM checkout from the respective plugin's pool.

```
Request -> Auth -> Core handler -> Audited command begins transaction
  -> Run before_update pipeline:
      -> checkout validator VM -> process(data, ctx) -> return data'
      -> checkout sanitizer VM -> process(data', ctx) -> return data''
  -> Execute write with data''
  -> Run after_update pipeline (async, post-commit)
  -> Response
```

If a before-hook processor rejects, the pipeline aborts and the write never happens. The client gets a clear error indicating which processor rejected and why.

### Audit Identity

When a plugin writes to core tables through a pipeline or gateway, the audit trail should capture both the human actor and the plugin involvement.

```go
type AuditContext struct {
    UserID    types.UserID
    PluginID  string  // empty for core operations
    Operation string
}
```

### Pipeline Introspection

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

### Pipeline Execution Logging

```
[2026-02-11 14:32:05] pipeline content_fields.before_update
  -> field-validator.validate   2.1ms  pass  (data modified: false)
  -> sanitizer.sanitize         0.8ms  pass  (data modified: true, fields: [value])
  -> audited write              1.2ms  ok
  total: 4.1ms
```

On rejection:

```
[2026-02-11 14:32:05] pipeline content_fields.before_update
  -> field-validator.validate   1.9ms  REJECTED  "min_length: title must be at least 3 characters"
  -> pipeline aborted, write not executed
```

### Pipeline Dry-Run

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

### Validation at Wire-Up Time

When a pipeline entry is created (via TUI or CLI), the CMS validates:

1. Plugin exists and is loaded
2. Capability match -- the plugin must declare that it handles this table/op combination (or `*` wildcard)
3. Handler exists -- the Lua function must be defined in the plugin's VM
4. No duplicate -- unique constraint prevents double-registration
5. Priority conflicts -- warn if two processors share the same priority on the same extension point

### Wildcard Policy

Plugins may declare `table = "*"` in capabilities (can process any table). However, pipeline wiring must be explicit per-table. No wildcard wiring to start. This can be relaxed later if the verbosity becomes painful, but it's easier to relax a constraint than to add one.

---

## Core Table Gateway (Tier 1 -- Scoped Access)

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

## Distributed Deployment

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

### User A -- No Plugins

- Plugin manager starts, finds no plugins in directory
- Pipeline registry is empty map
- Every write does a map lookup, gets nil, continues normally
- Overhead: effectively zero

### User B -- Plugin Endpoints Only (Sandboxed Tables)

- Plugin loads, creates `plugin_forms_entries` table
- Registers `POST /api/v1/plugin/forms/submit/:id`
- No pipeline entries configured
- Core writes are completely unaffected
- Plugin reads/writes only its own namespaced tables

### User C -- Analytics Plugin (After-Hooks)

- Plugin loads, creates `plugin_analytics_events` table
- Registers `GET /api/v1/plugin/analytics/events`
- Developer configures after-hook pipelines:
  - `content_data.after_create -> analytics.track_write (priority 10)`
  - `content_data.after_update -> analytics.track_write (priority 10)`
- Every content write fires the after-hook, analytics plugin writes to its own table
- Core write performance unaffected (after-hooks are async, post-commit)
- Admin panel fetches analytics via the plugin's GET endpoint

### User D -- Composed Pipeline (Multiple Plugins)

- Validator plugin: offers `before_create`, `before_update` on `content_fields`
- Sanitizer plugin: offers `before_create`, `before_update` on `content_fields`
- Analytics plugin: offers `after_*` on any table
- Developer wires the pipeline:

```
content_fields.before_create:
  1. field-validator.validate    (priority 10)
  2. sanitizer.sanitize          (priority 20)

content_fields.before_update:
  1. field-validator.validate    (priority 10)
  2. sanitizer.sanitize          (priority 20)

content_fields.after_create:
  1. analytics.track_write       (priority 10)
```

Data flow: request -> validator -> sanitizer -> audited write -> analytics. Each step is a Lua VM checkout from the respective plugin's pool. If the validator rejects, the write never happens and the client gets a 400.

---

## Comparison with WordPress

| Aspect                        | WordPress                                      | ModulaCMS                                          |
|-------------------------------|------------------------------------------------|---------------------------------------------------|
| Hook registration             | Plugin code calls `add_filter()` at load time  | Plugin declares capabilities, developer wires pipelines |
| Who controls what runs        | Plugin authors                                 | CMS developer/admin                               |
| Removing a hook               | Edit plugin code or deactivate entirely        | Disable one pipeline entry                         |
| Debugging "what runs on save" | grep codebase for `add_filter('save_post')`    | `modulacms pipelines show content_data`            |
| Performance when no plugins   | Still loads hook infrastructure                | Map lookup returns nil                             |
| Bad plugin impact             | Runs in main PHP thread, can take down site    | Timeout kills VM, write fails cleanly              |
| Chaining                      | Implicit via priority in `add_filter`          | Explicit via pipeline priority, visible in config  |
| Discovery                     | Filesystem scan, activate with one click       | Filesystem discovery, explicit install and enable   |
| Plugin updates                | Silent capability changes                       | Manifest diff, developer approval required         |
| Distributed coordination      | Not applicable (single-process PHP)            | DB-backed state, automatic instance coordination   |

---

## Implementation Phases

This plan builds on the existing Phase 1 (core engine) and replaces/refines the original Phase 2-4 roadmap items.

### Phase 2A: Plugin Lifecycle and Pipelines Schema
- Add `plugins` table and `pipelines` table schemas
- Implement three-state lifecycle (discovered/installed/enabled)
- CLI commands: `plugin install`, `plugin enable`, `plugin disable`, `plugin list`, `plugin inspect`
- Pipeline registry (in-memory, built from DB)
- Pipeline CLI commands: `pipelines list`, `pipelines show`, `pipelines configure`

### Phase 2B: Pipeline Execution
- Integrate pipeline checks into audited command layer
- Before-hook execution (in-transaction, sequential, can modify/reject)
- After-hook execution (post-commit, async, fire-and-forget)
- Pipeline execution logging with timing
- Timeout enforcement per processor

### Phase 2C: HTTP Integration
- Plugin endpoint registration via `http.handle(method, path, handler)`
- All routes namespaced under `/api/v1/plugin/{plugin_name}/`
- Request/response marshaling (LuaRequest, LuaResponse)
- Pool exhaustion -> HTTP 503 with Retry-After header

### Phase 2D: Core Table Gateway
- Scoped access API (`core.{table}.get`, `core.{table}.update`, etc.)
- Access enforcement based on approved capabilities
- Writes go through audited command infrastructure
- Audit identity includes plugin provenance

### Phase 3: Distributed Coordination
- Plugin watcher (polls DB for state changes)
- Hot pipeline reconfiguration (rebuild registry without restart)
- Self-eviction for missing plugin files
- Health endpoint integration for load balancer draining

### Phase 4: Production Hardening
- Pipeline dry-run mode
- Manifest change detection and developer approval flow
- Priority conflict warnings
- Circuit breaker (N failures -> temp disable processor)
- Per-plugin metrics (execution time, error rate, rejection rate)
- TUI screens for plugin and pipeline management
