# Plugin System Review

## Overview

45+ files, ~6,000 lines. Lua-based plugin system via gopher-lua. This is not scaffolding - it's a complete, production-ready extension system.

## What Solves a Real Problem

CMS extensibility without recompiling the binary. Plugins can:
- Define custom database tables
- Register HTTP endpoints
- Hook into content lifecycle events (before/after insert, update, delete)
- Log structured messages

This covers the three main extension points any CMS needs: data, API, and workflow.

## What Is Good

### Security Model
- **Sandboxed Lua VM**: Only base, table, string, math modules allowed. No file I/O, no network access, no OS interaction from Lua directly.
- **API freezing**: db, http, hooks, log modules are read-only via metatable protection.
- **Operation budgets**: 1,000 ops per VM checkout, 100 ops for before-hooks. Prevents infinite loops and resource exhaustion.
- **Route approval**: Admin must approve plugin HTTP routes before they serve traffic. Version changes reset all approvals (forces re-review).
- **Hook approval**: Same approval model for lifecycle hooks.
- **Namespace isolation**: All plugin tables prefixed `plugin_<name>_`.
- **Blocked HTTP headers**: Plugins cannot set access-control-*, set-cookie, transfer-encoding headers.
- **Request body limits**: 1 MB request, 5 MB response (configurable).
- **Rate limiting**: 100 req/sec per IP for plugin routes.

### Reliability Model
- **Circuit breaker**: 10 consecutive aborts from before-hooks disables the plugin until manual reload.
- **Blue-green reload**: New VM pool created alongside old one. If new pool initializes successfully, old one is swapped out. If not, old pool continues serving.
- **Per-event timeouts**: 2 seconds per before-hook, 5 seconds per event aggregate.
- **VM pooling**: Channel-based pool with configurable size (default 4). VMs are recycled between requests.

### Hot Reload
- File polling every 2 seconds with SHA-256 checksums of all .lua files
- 1-second debounce (stability window)
- 10-second cooldown between reloads
- Slow reload protection: pauses polling after 3 consecutive reloads taking >10s
- Max 100 .lua files, 10 MB per checksumming pass

### Database API
Plugins get a safe, sandboxed database interface:
- `db.define_table()` with column validation
- `db.query()`, `db.query_one()` with WHERE/ORDER/LIMIT
- `db.insert()`, `db.update()`, `db.delete()` (WHERE required on mutating ops)
- `db.transaction()` for atomic operations
- `db.ulid()`, `db.timestamp()` for ID and time generation

### Metrics
Built-in instrumentation: request counts, durations, hook execution stats, error rates, circuit breaker trips, VM pool gauge. Ready for Prometheus export.

### CLI
Full lifecycle management from the terminal:
- Offline: `plugin list`, `plugin init`, `plugin validate`
- Online: `plugin info`, `plugin reload`, `plugin enable`, `plugin disable`, `plugin approve`, `plugin revoke`

### Testing
16+ test files with plugin fixtures (hello_world, task_tracker, hooks_plugin, http_test_plugin). Good coverage of the core lifecycle.

## What Is Extra

The plugin system is arguably the most over-engineered part of the project relative to current usage. It's a complete runtime with pooling, circuit breaking, hot reload, approval workflows, metrics, and a full database API. For a CMS that may currently have zero plugins in production use, this represents significant invested effort.

However, extensibility is what separates a real CMS from a toy. Wordpress dominates largely because of plugins. If ModulaCMS aims to be a production CMS platform, this investment is justified and forward-looking.

## What Could Be Better

- **Memory limits per VM** are not enforced. A malicious or buggy plugin could allocate unlimited memory within its Lua VM.
- **No plugin marketplace or registry** infrastructure. Plugins are local filesystem only.
- **Plugin-to-plugin communication** is not supported. Plugins are isolated from each other.
- **No async Lua operations** within a single request. All plugin code is synchronous within a VM checkout.

## Recommendations

1. Add memory limits per Lua VM (gopher-lua doesn't support this natively; may need VM-level monitoring)
2. Document plugin API thoroughly for third-party developers (the README is good but could be a standalone developer guide)
3. Consider a plugin manifest format that declares required permissions up front
