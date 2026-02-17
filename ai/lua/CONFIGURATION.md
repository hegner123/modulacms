# Plugin System Configuration & Security

## config.json Fields

```json
{
    "plugin_enabled": false,
    "plugin_directory": "./plugins/",
    "plugin_max_vms": 4,
    "plugin_timeout": 5,
    "plugin_max_ops": 1000,
    "plugin_hot_reload": false,
    "plugin_max_failures": 5,
    "plugin_reset_interval": "60s",
    "plugin_rate_limit": 100,
    "plugin_max_routes": 50,
    "plugin_max_request_body": 1048576,
    "plugin_max_response_body": 5242880,
    "plugin_trusted_proxies": [],
    "plugin_hook_reserve_vms": 1,
    "plugin_hook_max_consecutive_aborts": 10,
    "plugin_hook_max_ops": 100,
    "plugin_hook_max_concurrent_after": 10,
    "plugin_hook_timeout_ms": 2000,
    "plugin_hook_event_timeout_ms": 5000
}
```

### Field Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plugin_enabled` | bool | false | Master switch for the plugin system |
| `plugin_directory` | string | `"./plugins/"` | Path to scan for plugin directories |
| `plugin_max_vms` | int | 4 | VM pool size per plugin (general + reserved) |
| `plugin_timeout` | int | 5 | Execution timeout in seconds per handler/hook |
| `plugin_max_ops` | int | 1000 | Max database operations per VM checkout |
| `plugin_hot_reload` | bool | false | Enable file-polling watcher for live reload |
| `plugin_max_failures` | int | 5 | Consecutive failures before circuit breaker trips |
| `plugin_reset_interval` | string | `"60s"` | Duration before circuit breaker half-opens |
| `plugin_rate_limit` | int | 100 | Max requests per second per IP |
| `plugin_max_routes` | int | 50 | Max routes per plugin |
| `plugin_max_request_body` | int | 1048576 | Max request body size in bytes (1 MB) |
| `plugin_max_response_body` | int | 5242880 | Max response body size in bytes (5 MB) |
| `plugin_trusted_proxies` | []string | [] | IP ranges for X-Forwarded-For parsing |
| `plugin_hook_reserve_vms` | int | 1 | Reserved VMs for hook execution |
| `plugin_hook_max_consecutive_aborts` | int | 10 | Hook errors before hook-level circuit breaker trips |
| `plugin_hook_max_ops` | int | 100 | Max db operations per before-hook |
| `plugin_hook_max_concurrent_after` | int | 10 | Max concurrent after-hook goroutines |
| `plugin_hook_timeout_ms` | int | 2000 | Per-hook execution timeout in ms |
| `plugin_hook_event_timeout_ms` | int | 5000 | Per-event chain timeout in ms |

## Security Architecture

### Sandbox

Source: `internal/plugin/sandbox.go`

Lua VMs run in a restricted environment:

- **No filesystem access**: `io`, `os`, `package` modules removed
- **No dynamic code loading**: `dofile`, `loadfile`, `load` removed
- **No metatable bypass**: `rawget`, `rawset`, `rawequal`, `rawlen` removed
- **No introspection**: `debug` module removed
- **Frozen modules**: `db`, `http`, `hooks`, `log` wrapped in read-only metatable proxy (writes raise error, getmetatable returns `"protected"`)

### Table Namespace Isolation

All plugin database tables are prefixed with `plugin_<name>_`. Plugins cannot:
- Access core CMS tables
- Access other plugins' tables
- Use foreign keys referencing tables outside their namespace

Table name validation: `db.ValidTableName()` enforces the prefix.

### Response Header Blocking

Plugins cannot set security-sensitive response headers:

| Blocked Header | Reason |
|---------------|--------|
| `access-control-*` | Prevents CORS policy override |
| `set-cookie` | Prevents session fixation |
| `transfer-encoding` | Prevents response smuggling |
| `content-length` | Prevents response smuggling |
| `cache-control` | Prevents cache poisoning |
| `host`, `connection` | Prevents request smuggling |

### Operation Budgets

Each VM checkout has a finite operation budget that prevents runaway queries:

| Context | Budget | Config Key |
|---------|--------|------------|
| HTTP request handler | 1000 ops | `plugin_max_ops` |
| Before-hook | 100 ops | `plugin_hook_max_ops` |

Exceeding the budget raises `ErrOpLimitExceeded` in Lua.

Additionally, before-hooks block all `db.*` calls entirely (prevents SQLite transaction deadlock since before-hooks run inside the CMS transaction).

### Rate Limiting

Per-IP token bucket rate limiter:
- Default: 100 requests/second
- Entries cached with last-seen timestamp
- Cleanup: every 5 minutes, removes entries unseen >10 minutes
- Exceeding limit returns 429 Too Many Requests

### Circuit Breakers

**Plugin-level** (source: `internal/plugin/recovery.go`):

| State | Behavior |
|-------|----------|
| Closed | Normal operation, requests allowed |
| Open | All requests rejected immediately |
| Half-Open | One probe request allowed; success closes, failure re-opens |

Transition: Closed -> Open after N consecutive failures (`plugin_max_failures`). Open -> Half-Open after reset interval (`plugin_reset_interval`).

Admin can force reset via `modulacms plugin enable <name>` or API.

**Hook-level** (source: `internal/plugin/hook_engine.go`):

Per (plugin, event, table) key. After N consecutive errors (`plugin_hook_max_consecutive_aborts`), that specific hook is disabled until the plugin is reloaded. Hook failures do NOT feed into the plugin-level circuit breaker.

### VM Pool Architecture

Source: `internal/plugin/pool.go`

**Two-channel design**:
- General channel: serves HTTP requests
- Reserved channel: serves hooks (guaranteed availability)

**Checkout lifecycle**:
1. `Get()`: acquire from general channel (100ms timeout)
2. Set context on VM
3. Execute handler/hook
4. `Put()`: clear stack, restore globals, validate health, return to channel

**Health validation**: VMs checked on every return. Corrupted VMs are replaced with fresh instances.

**Pool lifecycle** (Phase 4 tri-state):
- Open: normal Get/Put
- Draining: Get returns ErrPoolExhausted, Put still accepts returns
- Closed: both Get and Put close VMs directly

### Hot Reload Security

Source: `internal/plugin/watcher.go`

| Limit | Value | Purpose |
|-------|-------|---------|
| Max .lua files per plugin | 100 | Prevents DoS via file count |
| Max total size per checksum | 10 MB | Prevents DoS via file size |
| Reload cooldown | 10s per plugin | Prevents reload storms |
| Max consecutive slow reloads | 3 (>10s each) | Pauses watcher on systemic issues |

**Blue-green reload**: New instance created alongside old. If new fails, old keeps running. If new succeeds, old is drained and replaced atomically.

## Admin CLI Commands

### Offline (no server required)

```bash
modulacms plugin list              # List discovered plugins
modulacms plugin init <name>       # Scaffold new plugin directory
modulacms plugin validate <path>   # Validate manifest without loading
```

### Online (requires running server)

```bash
modulacms plugin info <name>       # Plugin details and status
modulacms plugin reload <name>     # Trigger hot reload
modulacms plugin enable <name>     # Reset circuit breaker, reload
modulacms plugin disable <name>    # Stop plugin, trip circuit breaker

# Approval
modulacms plugin approve <name> --all-routes
modulacms plugin approve <name> --all-hooks
modulacms plugin approve <name> --route "GET /tasks"
modulacms plugin approve <name> --hook "before_create:content_data"

# Revocation
modulacms plugin revoke <name> --all-routes
modulacms plugin revoke <name> --route "GET /tasks"
```

## Admin API Endpoints

All require admin authentication.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/plugins` | List all plugins |
| GET | `/api/v1/admin/plugins/{name}` | Plugin details |
| POST | `/api/v1/admin/plugins/{name}/reload` | Trigger hot reload |
| POST | `/api/v1/admin/plugins/{name}/enable` | Reset circuit breaker |
| POST | `/api/v1/admin/plugins/{name}/disable` | Trip circuit breaker |
| GET | `/api/v1/admin/plugins/routes` | List routes with approval status |
| POST | `/api/v1/admin/plugins/routes/approve` | Approve routes |
| POST | `/api/v1/admin/plugins/routes/revoke` | Revoke route approval |
| GET | `/api/v1/admin/plugins/hooks` | List hooks with approval status |
| POST | `/api/v1/admin/plugins/hooks/approve` | Approve hooks |
| POST | `/api/v1/admin/plugins/hooks/revoke` | Revoke hook approval |

## Metrics

Source: `internal/plugin/metrics.go`

| Metric | Labels | Description |
|--------|--------|-------------|
| `plugin.http.requests` | plugin, method, status | HTTP request count |
| `plugin.http.duration_ms` | plugin, method, status | HTTP request latency |
| `plugin.hook.before` | plugin, event, table, status | Before-hook count |
| `plugin.hook.after` | plugin, event, table, status | After-hook count |
| `plugin.hook.duration_ms` | plugin, event, table, status | Hook latency |
| `plugin.errors` | plugin, type | Error count |
| `plugin.circuit_breaker.trip` | plugin | CB trip count |
| `plugin.reload` | plugin, status | Reload event count |
| `plugin.vm.available` | plugin | Available VM gauge |
