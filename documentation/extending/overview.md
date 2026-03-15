# Plugin Development Overview

Extend ModulaCMS with custom HTTP endpoints, content lifecycle hooks, isolated database storage, and TUI screens using Lua plugins.

> **Good to know**: This section is for plugin developers. For CMS administrator guidance on installing, configuring, and managing plugins, see the [Configuration](configuration.md) page.

For a hands-on walkthrough, see the [tutorial](tutorial.md). For API details, see the [Lua API reference](lua-api.md).

## What Plugins Can Do

- **Custom HTTP endpoints.** Register REST routes under `/api/v1/plugins/<plugin_name>/` with path parameters, middleware, and public or authenticated access.
- **Content lifecycle hooks.** React to create, update, delete, publish, and archive events on CMS content tables. Before-hooks can abort transactions; after-hooks run asynchronously.
- **Isolated database storage.** Create plugin-owned tables with full CRUD, transactions, indexes, and foreign keys. All tables are namespaced to the plugin and invisible to other plugins and the CMS core.
- **Structured logging.** Write log entries at info, warn, error, and debug levels with arbitrary key-value context. The plugin name is injected automatically.
- **TUI screens.** Define standalone pages in the SSH terminal UI using a coroutine bridge. Plugin code yields layout tables (grids, lists, tables, text blocks) and receives key/data events on resume. Screens appear in the TUI sidebar for operators to navigate to.
- **Field interfaces.** Provide custom editors for content fields of type `plugin`. Inline interfaces render within the field row; overlay interfaces open as full-screen modals. The plugin produces and commits field values through the same coroutine protocol.

Plugins cannot access the filesystem, execute system commands, load dynamic code, or read CMS database tables.

## Architecture

### Sandboxed Execution

Each plugin runs inside a restricted Lua 5.1 environment. ModulaCMS removes dangerous modules (`io`, `os`, `package`, `debug`) and dynamic code loading functions (`load`, `dofile`, `loadfile`) before any plugin code executes. All injected API modules (`db`, `http`, `hooks`, `log`) are frozen read-only -- writes raise an error.

See [security](security.md) for the full sandbox details.

### VM Pool

Each plugin gets a pool of pre-initialized Lua VMs (default 4, configurable via `plugin_max_vms`). The pool uses a three-pool design:

- **General channel.** Serves HTTP request handlers. If all VMs are busy, the request gets a 503 response.
- **Reserved channel.** Serves content hooks exclusively. Sized by `plugin_hook_reserve_vms` (default 1). This guarantees hook execution even when HTTP traffic saturates the general pool.
- **UI pool.** Serves TUI screen and field interface coroutines. VMs are held for the lifetime of a screen/field session. Default 4 VMs per plugin (`plugin_max_ui_vms`). Pool exhaustion shows "Plugin busy" instead of blocking. Only created for plugins that declare `screens` or `interfaces` in their manifest.

### Blue-Green Reload

When you reload a plugin (manually or via hot reload), ModulaCMS creates a new plugin instance alongside the old one. If the new instance loads successfully, it replaces the old one atomically. If the new instance fails, the old one keeps running unchanged.

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

Each plugin's tables are namespaced with a `plugin_<name>_` prefix. A plugin named `bookmarks` that defines a table called `links` gets `plugin_bookmarks_links` in the database. Plugins cannot query tables outside their namespace, and foreign keys must reference tables owned by the same plugin.

Three columns are auto-injected on every plugin table: `id` (ULID primary key), `created_at`, and `updated_at` (RFC3339 timestamps).

### Route Prefix

All plugin HTTP routes are mounted under `/api/v1/plugins/<plugin_name>/`. A route registered as `GET /tasks` on the `task_tracker` plugin becomes `GET /api/v1/plugins/task_tracker/tasks`.

### Hook Injection Points

Hooks attach to CMS content lifecycle events. Before-hooks run synchronously and can abort the operation by calling `error()`. After-hooks run asynchronously after the operation completes. Hooks fire only on CMS content tables, not on plugin tables.

### Approval Gate

Routes and hooks start unapproved. Unapproved routes return 404; unapproved hooks are silently skipped. An admin must explicitly approve each route and hook before it executes. When a plugin's version changes, all approvals are revoked automatically, requiring re-approval of the updated code.

See [approval workflow](approval.md) for details.

### Circuit Breakers

Two independent circuit breaker systems protect the CMS:

- **Plugin-level.** Tracks consecutive HTTP handler failures and manager operation errors. After `plugin_max_failures` consecutive errors (default 5), the breaker opens and all requests return 503 until the reset interval elapses and a probe request succeeds.
- **Hook-level.** Tracks consecutive errors per (plugin, event, table) combination. After `plugin_hook_max_consecutive_aborts` consecutive errors (default 10), that specific hook is disabled until you reload the plugin. Hook failures do not feed into the plugin-level breaker.

See [security](security.md) for the full security model.

## Next Steps

- [Tutorial](tutorial.md) -- build a bookmarks plugin from scratch
- [Lua API Reference](lua-api.md) -- every function, parameter, and return value
- [Configuration](configuration.md) -- all modula.config.json fields with tuning guidance
- [Security](security.md) -- sandbox, isolation, and protection mechanisms
- [Approval Workflow](approval.md) -- route and hook approval in detail
- [Examples](examples.md) -- complete example plugins demonstrating common patterns
- [Configuration](configuration.md) -- all modula.config.json plugin fields with tuning guidance
