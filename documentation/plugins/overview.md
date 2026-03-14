# Plugin Development Overview

This section is for plugin developers. If you are a CMS administrator looking to install, configure, and manage plugins, see the [Managing Plugins](../guides/plugins.md) guide instead.

ModulaCMS plugins are Lua scripts that extend the CMS with custom HTTP endpoints, content lifecycle hooks, isolated database storage, and TUI screen interfaces. Plugins run in sandboxed Lua VMs with controlled access to database operations, HTTP routing, terminal UI rendering, and structured logging. The system enforces resource limits, approval gates, and circuit breakers to prevent plugins from affecting core CMS stability.

This page covers architecture and design. For a hands-on walkthrough, see the [tutorial](tutorial.md). For API details, see the [Lua API reference](lua-api.md).

## What Plugins Can Do

- **Custom HTTP endpoints.** Register REST routes under `/api/v1/plugins/<plugin_name>/` with path parameters, middleware, and public or authenticated access.
- **Content lifecycle hooks.** React to create, update, delete, publish, and archive events on CMS content tables. Before-hooks can abort transactions; after-hooks run asynchronously.
- **Isolated database storage.** Create plugin-owned tables with full CRUD, transactions, indexes, and foreign keys. All tables are namespaced to the plugin and invisible to other plugins and the CMS core.
- **Structured logging.** Write log entries at info, warn, error, and debug levels with arbitrary key-value context. Plugin name is injected automatically.
- **TUI screens.** Define standalone pages in the SSH terminal UI using a coroutine bridge. Plugin code yields layout tables (grids, lists, tables, text blocks) and receives key/data events on resume. Screens appear in the TUI sidebar for operators to navigate to.
- **Field interfaces.** Provide custom editors for content fields of type `plugin`. Inline interfaces render within the field row; overlay interfaces open as full-screen modals. The plugin produces and commits field values through the same coroutine protocol.

Plugins cannot access the filesystem, execute system commands, load dynamic code, or read CMS database tables.

## Architecture

### Lua Sandbox

Each plugin runs inside [gopher-lua](https://github.com/yuin/gopher-lua) VMs with a restricted standard library. Dangerous modules (`io`, `os`, `package`, `debug`) and dynamic code loading functions (`load`, `dofile`, `loadfile`) are removed before any plugin code executes. The `rawget`, `rawset`, `rawequal`, and `rawlen` functions are also stripped to prevent metatable bypass. All injected API modules (`db`, `http`, `hooks`, `log`) are frozen read-only via metatable proxy -- writes raise an error, and `getmetatable` returns `"protected"`.

Source: `internal/plugin/sandbox.go`

### VM Pool

Each plugin gets a pool of pre-initialized Lua VMs (default 4, configurable via `plugin_max_vms`). The pool uses a three-pool design:

- **General channel.** Serves HTTP request handlers. VMs are checked out with a 100ms timeout; if all are busy, the request gets a 503 response.
- **Reserved channel.** Serves content hooks exclusively. Sized by `plugin_hook_reserve_vms` (default 1). This guarantees hook execution even when HTTP traffic saturates the general pool.
- **UI pool.** Serves TUI screen and field interface coroutines. VMs are held for the lifetime of a screen/field session with no acquisition timeout. Default 4 VMs per plugin (`plugin_max_ui_vms`). Pool exhaustion shows "Plugin busy" instead of blocking. Only created for plugins that declare `screens` or `interfaces` in their manifest.

On checkout, the VM receives an operation budget and execution context. On return, the pool clears the Lua stack, restores global variables to their initial snapshot, and validates VM health. Corrupted VMs are replaced with fresh instances.

The general and reserved pools share three lifecycle phases:

| Phase | Get() | Put() |
|-------|-------|-------|
| Open | Normal checkout | Normal return |
| Draining | Returns `ErrPoolExhausted` | Accepts returns |
| Closed | Closes VM directly | Closes VM directly |

The UI pool has independent drain semantics: active UI coroutines continue on old VMs until the user navigates away, then the old VM is closed rather than returned.

Source: `internal/plugin/pool.go`, `internal/plugin/ui_pool.go`

### Blue-Green Reload

When a plugin is reloaded (manually or via hot reload), the system creates a new plugin instance alongside the old one. The new instance goes through full initialization: VM pool creation, `init.lua` loading, route and hook registration, `on_init()` execution. If the new instance loads successfully, the old instance's pool is drained and replaced atomically. If the new instance fails, the old one keeps running unchanged.

Source: `internal/plugin/manager.go` (`ReloadPlugin`)

## Plugin Lifecycle

A plugin moves through five states:

| State | Value | Description |
|-------|-------|-------------|
| Discovered | 0 | Directory found, not yet loaded |
| Loading | 1 | VM pool being created, `init.lua` executing |
| Running | 2 | Serving HTTP requests and hooks |
| Failed | 3 | Initialization or runtime error; not serving traffic |
| Stopped | 4 | Administratively disabled |

State transitions:

```
Discovered -> Loading -> Running
                 |           |
                 v           v
               Failed     Stopped
                 |           |
                 +--> Loading (via enable/reload)
```

A failed plugin can be retried with `modulacms plugin enable <name>`, which resets the circuit breaker and triggers a fresh load. A stopped plugin can be re-enabled the same way.

## How Plugins Interact with the CMS

### Database Isolation

Plugin tables are prefixed with `plugin_<name>_`. A plugin named `bookmarks` that defines a table called `links` gets `plugin_bookmarks_links` in the database. The prefix is enforced at the API level -- plugins cannot query tables outside their namespace. Foreign keys must reference tables owned by the same plugin.

Three columns are auto-injected on every plugin table: `id` (ULID primary key), `created_at`, and `updated_at` (RFC3339 timestamps).

### Route Prefix

All plugin HTTP routes are mounted under `/api/v1/plugins/<plugin_name>/`. A route registered as `GET /tasks` on the `task_tracker` plugin becomes `GET /api/v1/plugins/task_tracker/tasks`.

### Hook Injection Points

Hooks attach to CMS content lifecycle events. Before-hooks run synchronously inside the CMS database transaction and can abort it by calling `error()`. After-hooks run asynchronously after the transaction commits. Hooks fire only on CMS tables (e.g., `content_data`), not on plugin tables.

### Approval Gate

Routes and hooks start unapproved. Unapproved routes return 404; unapproved hooks are silently skipped. An admin must explicitly approve each route and hook before it executes. When a plugin's version changes, all approvals are revoked automatically, requiring re-approval of the updated code.

See [approval workflow](approval.md) for details.

### Circuit Breakers

Two independent circuit breaker systems protect the CMS:

- **Plugin-level.** Tracks consecutive HTTP handler failures and manager operation errors. After `plugin_max_failures` consecutive errors (default 5), the breaker opens and all requests return 503 until the reset interval elapses and a probe request succeeds.
- **Hook-level.** Tracks consecutive errors per (plugin, event, table) combination. After `plugin_hook_max_consecutive_aborts` consecutive errors (default 10), that specific hook is disabled until the plugin is reloaded. Hook failures do not feed into the plugin-level breaker.

See [security](security.md) for the full security model.

## Further Reading

- [Tutorial](tutorial.md) -- build a bookmarks plugin from scratch
- [Lua API Reference](lua-api.md) -- every function, parameter, and return value
- [Configuration](configuration.md) -- all modula.config.json fields with tuning guidance
- [Security](security.md) -- sandbox, isolation, and protection mechanisms
- [Approval Workflow](approval.md) -- route and hook approval in detail
- [Examples](examples.md) -- complete example plugins demonstrating common patterns
- [Managing Plugins](../guides/plugins.md) -- admin guide for installing, configuring, and approving plugins
