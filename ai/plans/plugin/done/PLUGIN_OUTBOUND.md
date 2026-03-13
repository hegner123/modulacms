# Plan: Plugin Outbound Request Module

## Context

Plugins currently cannot make outbound HTTP requests. This module adds a `request` Lua API that allows plugins to register domains they need to call, with admin approve/deny control matching the existing route and hook patterns. This enables plugins to integrate with external APIs (webhooks, third-party services, etc.) while maintaining the same security model.

## Lua API Surface

```lua
-- Module scope: register domains
request.register("api.example.com", { description = "Webhook notifications" })
request.register("data.external.io")

-- Runtime: make requests (only to approved domains)
local resp = request.send("POST", "https://api.example.com/webhook", {
  headers = {["authorization"] = "Bearer token123"},
  json = {event = "published", id = "abc"},
  body = "raw string body",       -- mutually exclusive with json
  timeout = 5,                    -- seconds, capped by engine max
  parse_json = true               -- opt-in: parse response body as JSON into resp.json
})
-- resp = {status = 200, body = "...", headers = {...}, json = {...}}
-- resp.json is only present when parse_json = true AND Content-Type is application/json
-- On failure: resp = {error = "domain not approved"}

-- Convenience methods:
local resp = request.get(url [, opts])
local resp = request.post(url [, opts])
local resp = request.put(url [, opts])
local resp = request.delete(url [, opts])
local resp = request.patch(url [, opts])
```

### Error Contract

Two distinct error categories matching existing Lua API patterns:

**Raises Lua error (L.RaiseError) -- programming mistakes, halt execution:**
- Phase guard violations: `request.send()` called at module scope, inside `on_init()`, or during `on_shutdown()`
- Before-hook violations: `request.send()` called inside a before-hook handler
- Argument validation: missing/invalid method, missing URL, invalid opts type
- URL with userinfo: `request.send()` called with a URL containing `user:pass@host`
- `request.register()` called outside module scope
- Domain count limit exceeded (50 per plugin)
- Request body too large (exceeds `MaxRequestBodyBytes`)

**Returns error table `{error = "..."}` -- operational failures, plugin can handle:**
- Domain not approved: `{error = "domain not approved: api.example.com"}`
- Circuit breaker open: `{error = "circuit breaker open for domain: api.example.com"}`
- Rate limit exceeded: `{error = "rate limit exceeded for domain: api.example.com"}`
- Global rate limit exceeded: `{error = "global outbound request rate limit exceeded"}`
- SSRF blocked: `{error = "request to private/reserved IP address blocked"}`
- Connection refused: `{error = "connection refused: api.example.com"}`
- DNS resolution failed: `{error = "dns lookup failed: api.example.com"}`
- Request timeout: `{error = "request timed out after 5s"}`
- Response too large: `{error = "response exceeded maximum size (1048576 bytes)"}`
- TLS handshake failure: `{error = "tls handshake failed: api.example.com"}`
- Context canceled (shutdown): `{error = "request canceled"}`

### Retry Policy

Retries are intentionally excluded from the engine. Reasons:
- Retries multiply outbound traffic and interact unpredictably with rate limits
- Only the plugin author knows which failures are retryable for their use case (a 429 from Stripe is retryable; a 400 is not)
- Plugin authors can implement retries in Lua with full control over backoff and max attempts
- Keeps the engine simple and predictable

## Admin API Endpoints

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | /api/v1/admin/plugins/requests | authenticated | List all registered domains + approval status |
| POST | /api/v1/admin/plugins/requests/approve | adminOnly | Approve domains: `{"requests": [{"plugin": "...", "domain": "..."}]}` |
| POST | /api/v1/admin/plugins/requests/revoke | adminOnly | Revoke domains (same body format) |

## New Files

### 1. internal/plugin/request_api.go -- Lua API

- `PendingRequest` struct: `{Domain, Description}`
- `requestAPIState` struct: holds `engine *RequestEngine`, `pluginName string`, `inBeforeHook bool`
  - `inBeforeHook` set to `true` by hook_engine before executing before-hooks, reset after (same pattern as `DatabaseAPI.inBeforeHook`)
  - INVARIANT: each `requestAPIState` instance is bound to exactly one LState (1:1), same as `DatabaseAPI`. Never share across VMs.
- `RegisterRequestAPI(L, pluginName, engine) *requestAPIState`: creates `__request_pending` hidden table + public `request` table
- `request.register(domain [, opts])`: module-scope only, phase guard via `__vm_phase` registry key (must be `"module_scope"`), validates domain, count limit (50), duplicate detection
- `request.send(method, url [, opts])`: runtime only with three guards:
  1. Phase guard: checks `__vm_phase` registry key; raises Lua error if value is `"module_scope"`, `"init"`, or `"shutdown"` (only `"runtime"` is allowed)
  2. Before-hook guard: checks `requestAPIState.inBeforeHook`; raises Lua error if true (before-hooks run inside caller's transaction -- a 10s HTTP call would hold DB locks)
  3. URL validation: parses URL via `net/url.Parse`, rejects URLs with userinfo (`user:pass@host` -- serves no legitimate purpose, common SSRF parsing confusion vector)
  4. Validates `body`/`json` options are not both set; raises Lua error if both present
  5. Checks serialized body size against `MaxRequestBodyBytes`; raises Lua error if exceeded
  6. Calls `engine.Execute()`, returns `{status, body, headers}` or `{error = "..."}`
  7. If `opts.parse_json` is true, attempts to parse response body as JSON into a `json` field on the response table
- `request.get/post/put/delete/patch`: convenience wrappers around `send`
- `ReadPendingRequests(L) []PendingRequest`: extracts from `__request_pending` after init.lua
- `validateDomain(domain)`: no scheme, no path, no port, no wildcards, `a-zA-Z0-9.-` only, max 253 chars
- **Domain matching in Execute:** URL is parsed via `net/url.Parse`. Hostname is extracted via `url.Hostname()` (strips port). Hostname must be an exact case-insensitive match against the approved domain list. No suffix matching, no prefix matching, no wildcard matching.

### 2. internal/plugin/request_engine.go -- Engine

**Struct:**
```
RequestEngine {
    approved        map[string]bool           // "plugin:domain" -> approved
    mu              sync.RWMutex              // protects approved map
    client          *http.Client              // shared, redirect-disabled, configured transport
    rateLimiters    map[string]*rateLimiterEntry  // "plugin:domain" -> token bucket
    rateMu          sync.Mutex
    globalLimiter   *rate.Limiter             // aggregate cap across all plugins/domains
    circuitBreakers map[string]*domainCB      // "plugin:domain" -> circuit breaker state
    cbMu            sync.RWMutex
    cleanupDone     chan struct{}              // closed to stop the rate limiter cleanup goroutine
    pool            *sql.DB
    dialect         string
    cfg             RequestEngineConfig
}
```

The `cleanupDone` channel is created in `NewRequestEngine()` and closed in `Close()`. The rate limiter cleanup goroutine selects on this channel to know when to stop (matching the `HTTPBridge.cleanupDone` pattern from `http_bridge.go:75`).

**Key format invariant:** All map keys use `"plugin:domain"` format. This is safe because plugin names are validated as `[a-z0-9_]` (see `ValidateManifest` in `manager.go:783-791`) and domain names are validated as `[a-zA-Z0-9.-]` (see `validateDomain`). Neither character set includes `:`, so the delimiter is unambiguous. If plugin name validation is ever relaxed to allow `:`, the key format must be updated.

**Config:**
```
RequestEngineConfig {
    DefaultTimeoutSec      int    // default 10, max per-request timeout
    MaxResponseBytes       int64  // default 1MB (1_048_576)
    MaxRequestBodyBytes    int64  // default 1MB (1_048_576)
    MaxRequestsPerMin      int    // default 60, per plugin per domain
    GlobalMaxRequestsPerMin int   // default 600, aggregate across all plugins/domains; 0 = unlimited
    CBMaxFailures          int    // default 5, consecutive failures to trip
    CBResetIntervalSec     int    // default 60, seconds before half-open probe
    AllowLocalhost         bool   // default false; when true, HTTP to 127.0.0.1/::1 is permitted (development only)
}
```

**HTTP Transport configuration:**
```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
    TLSHandshakeTimeout: 10 * time.Second,
    DialContext:         ssrfSafeDialer(cfg).DialContext,  // SSRF-safe dialer
}
client := &http.Client{
    Transport:     transport,
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse  // no redirect following
    },
    // No default timeout here -- per-request context.WithTimeout is used
}
```

**SSRF-Safe Dialer (`ssrfSafeDialer`):**

Custom `net.Dialer` wrapper with a `Control` function that inspects resolved IP addresses before connection establishment:

```go
func ssrfSafeDialer(cfg RequestEngineConfig) *net.Dialer {
    return &net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
        Control: func(network, address string, c syscall.RawConn) error {
            host, _, err := net.SplitHostPort(address)
            if err != nil {
                return fmt.Errorf("failed to parse address %q: %w", address, err)
            }
            ip := net.ParseIP(host)
            if ip == nil {
                return fmt.Errorf("invalid IP in resolved address: %q", host)
            }
            if isBlockedIP(ip, cfg.AllowLocalhost) {
                return fmt.Errorf("request to private/reserved IP blocked: %s", ip)
            }
            return nil
        },
    }
}
```

Blocked IP ranges (`isBlockedIP(ip, allowLocalhost)`):
- `127.0.0.0/8` -- loopback (allowed only when `AllowLocalhost` is true)
- `10.0.0.0/8` -- RFC 1918 private
- `172.16.0.0/12` -- RFC 1918 private
- `192.168.0.0/16` -- RFC 1918 private
- `169.254.0.0/16` -- link-local (includes AWS/GCP/Azure metadata at 169.254.169.254)
- `fc00::/7` -- IPv6 unique local
- `fe80::/10` -- IPv6 link-local
- `::1/128` -- IPv6 loopback (allowed only when `AllowLocalhost` is true)
- `0.0.0.0/8` -- "this" network
- `100.64.0.0/10` -- shared address space (CGN)
- `192.0.0.0/24` -- IETF protocol assignments
- `198.18.0.0/15` -- benchmarking

All private/reserved IPs are blocked by default. Localhost (`127.0.0.0/8` and `::1`) is only allowed when `AllowLocalhost` is explicitly enabled in config (default: false). This is intended for development environments only -- enabling it in production is dangerous as it allows plugins to probe any localhost service. The SSRF check runs after DNS resolution via the dialer Control hook, so DNS rebinding attacks that resolve a public domain to a private IP are caught.

**Per-Domain Circuit Breaker** (inline map pattern, matching hook_engine):
```
domainCB {
    consecutiveFailures int
    disabled            bool
    lastFailure         time.Time
}
```

Trip conditions:
- Network errors (connection refused, DNS failure, TLS error, timeout) increment `consecutiveFailures`
- HTTP 5xx responses increment `consecutiveFailures`
- HTTP 2xx/3xx/4xx responses reset `consecutiveFailures` to 0
- When `consecutiveFailures >= CBMaxFailures` (default 5), set `disabled = true`
- When `disabled` and `time.Since(lastFailure) >= CBResetIntervalSec`, allow one probe request (half-open). Success resets to closed; failure re-trips.

Rationale: per-plugin-per-domain so that a dead Stripe API does not block Slack webhooks from the same plugin, while also providing isolation between plugins hitting the same domain. The key format `"plugin:domain"` means two plugins calling the same dead API will each independently send probe requests during half-open state.

**Per-Domain Rate Limiter:**

Uses `golang.org/x/time/rate` token bucket (matching HTTP bridge pattern):
```
rateLimiterEntry {
    limiter  *rate.Limiter   // MaxRequestsPerMin / 60 tokens/sec, burst = min(MaxRequestsPerMin / 60, 5)
    lastSeen time.Time
}
```

Burst is capped at 5 regardless of rate. This prevents concentrated request spikes when admins set high rate limits (e.g., 600/min would give burst=10 without the cap). The cap ensures requests are spread over time.

Key: `"plugin:domain"`. Background cleanup goroutine removes entries not seen for 10+ minutes (same as HTTP bridge). This means a plugin calling 3 APIs gets 60/min to each, not 60/min shared.

**Global Rate Limiter:**

In addition to per-plugin-per-domain rate limiting, a global `*rate.Limiter` caps total outbound requests across all plugins at `GlobalMaxRequestsPerMin` (default 600, burst capped at 5). This acts as a backstop -- 20 plugins each at 60/min/domain with 5 approved domains could produce 6,000 requests/min without this cap, which could trigger abuse complaints from hosting providers. Checked in `Execute()` after the per-domain rate limiter. Set to 0 to disable.

**DB table `plugin_requests` -- tri-dialect DDL:**

SQLite:
```sql
CREATE TABLE IF NOT EXISTS plugin_requests (
    plugin_name    TEXT    NOT NULL,
    domain         TEXT    NOT NULL,
    description    TEXT    NOT NULL DEFAULT '',
    approved       INTEGER NOT NULL DEFAULT 0,
    approved_at    TEXT,
    approved_by    TEXT,
    plugin_version TEXT    NOT NULL DEFAULT '',
    created_at     TEXT    NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (plugin_name, domain)
)
```

MySQL:
```sql
CREATE TABLE IF NOT EXISTS plugin_requests (
    plugin_name    VARCHAR(255) NOT NULL,
    domain         VARCHAR(253) NOT NULL,
    description    TEXT         NOT NULL,
    approved       TINYINT(1)   NOT NULL DEFAULT 0,
    approved_at    TIMESTAMP    NULL DEFAULT NULL,
    approved_by    VARCHAR(255),
    plugin_version VARCHAR(255) NOT NULL DEFAULT '',
    created_at     DATETIME     DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (plugin_name(191), domain(253))
)
```

PostgreSQL:
```sql
CREATE TABLE IF NOT EXISTS plugin_requests (
    plugin_name    TEXT    NOT NULL,
    domain         TEXT    NOT NULL,
    description    TEXT    NOT NULL DEFAULT '',
    approved       BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at    TIMESTAMPTZ,
    approved_by    TEXT,
    plugin_version TEXT    NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (plugin_name, domain)
)
```

Follows the same column patterns as `plugin_hooks` and `plugin_routes`. MySQL uses prefix key lengths on the primary key to stay within InnoDB's 3072-byte limit with utf8mb4 (4 bytes/char): `plugin_name(191)` = 764 bytes + `domain(253)` = 1012 bytes = 1776 bytes total, well under the 3072-byte limit. Domain column uses VARCHAR(253) since `validateDomain` caps domains at 253 chars.

**`RequestRegistration` struct (matches `HookRegistration` pattern from `hook_engine.go:909-916`):**
```go
type RequestRegistration struct {
    PluginName    string `json:"plugin_name"`
    Domain        string `json:"domain"`
    Description   string `json:"description"`
    Approved      bool   `json:"approved"`
    PluginVersion string `json:"plugin_version"`
}
```
Note: `created_at` and `approved_at`/`approved_by` are intentionally omitted from the struct (matching `HookRegistration` which also omits `approved_at`/`approved_by`). These columns exist in the DDL for DB-level audit but are not exposed in the list API response.

**Methods (follow hook_engine.go pattern):**
- `CreatePluginRequestsTable(ctx)`, `CleanupOrphanedRequests(ctx, discovered)`
- `UpsertRequestRegistrations(ctx, pluginName, version, requests)` -- ON CONFLICT updates `plugin_version` and `description` only, does NOT revoke approval (matches hook_engine behavior, NOT route bridge behavior). Logs version change at INFO level for audit trail.
- `ApproveRequest(ctx, pluginName, domain, approvedBy)`, `RevokeRequest(ctx, pluginName, domain)`
- `ListRequests(ctx) []RequestRegistration` -- queries the `plugin_requests` DB table directly (not the in-memory `approved` map). The in-memory map only stores `"plugin:domain" -> bool` which cannot reconstruct individual fields like `Description` and `PluginVersion`. This matches `HookEngine.ListHooks()` which reads from the structured `hookIndex`, not the flat `approved` map. Sort results by `(plugin_name, domain)` for deterministic output.
- `UnregisterPlugin(pluginName)` -- clears in-memory maps (approved, rate limiters, circuit breakers)
- `Close()` -- takes no parameters (unlike `HookEngine.Close(ctx context.Context)`, there is no WaitGroup to drain, so no context is needed). (1) Closes `cleanupDone` channel to stop the rate limiter cleanup goroutine, (2) closes idle transport connections via `client.CloseIdleConnections()`. Does NOT drain in-flight requests; those are governed by their per-request `context.WithTimeout` and will complete or timeout naturally when the parent context (from shutdown) is canceled. Called AFTER `hookEngine.Close()` to ensure all after-hook goroutines (which may call `request.send()`) have completed first.

**Execute(ctx, pluginName, method, urlStr, opts) (map, error):**

1. Parse URL via `net/url.Parse`, reject URLs with userinfo (`url.User != nil`), validate scheme (HTTPS required; HTTP only when `AllowLocalhost` is true and hostname is `127.0.0.1`/`::1`)
2. Extract hostname via `url.Hostname()`, exact case-insensitive match against approval in-memory map
3. Check per-domain circuit breaker; return `{error}` if open
4. Check per-domain rate limiter; return `{error}` if exceeded
5. Check global rate limiter; return `{error = "global outbound request rate limit exceeded"}` if exceeded
6. Serialize body: if `opts.json`, marshal to JSON and check size against `MaxRequestBodyBytes`; if `opts.body` (string), check size
7. Build `*http.Request` with `context.WithTimeout(ctx, clamp(opts.timeout, 1, engine.DefaultTimeoutSec))`. The timeout is clamped to a minimum of 1 second (prevents `timeout=0` which would immediately cancel via `context.WithTimeout(ctx, 0)`) and a maximum of `DefaultTimeoutSec` (default 10s). If `opts.timeout` is not provided or is negative, `DefaultTimeoutSec` is used.
8. Set headers from opts; auto-set `Content-Type: application/json` for JSON body; always set `User-Agent: ModulaCMS-Plugin/1.0`
9. Execute via shared client (SSRF-safe dialer catches private IPs at connection time)
10. Read response body with `io.LimitReader` at `MaxResponseBytes`; if body exceeds limit, return `{error}`
11. Record circuit breaker result (success resets, network error / 5xx increments)
12. Record metrics via `RecordOutboundRequest`
13. Return `{Status, Body, Headers}`. If `opts.parse_json` is true and response Content-Type is `application/json`, parse body into `json` field. JSON parsing is opt-in to prevent a malicious server from influencing processing behavior via Content-Type headers (deeply nested JSON can consume CPU even within the 1MB size cap).

### 3. internal/plugin/request_handlers.go -- Admin Handlers

- `PluginRequestsListHandler(mgr)`: GET, returns `{"requests": [...]}`
- `PluginRequestsApproveHandler(mgr)`: POST, body `{"requests": [{"plugin","domain"}]}`, extracts `approvedBy`
- `PluginRequestsRevokeHandler(mgr)`: POST, same body format
- Follows exact pattern of hook_handlers.go (body size limit, error collection, JSON responses)

### 4. internal/plugin/request_ssrf.go -- SSRF Protection

- `isBlockedIP(ip net.IP, allowLocalhost bool) bool`: checks IP against blocked CIDR ranges; skips loopback check when `allowLocalhost` is true
- `ssrfSafeDialer(cfg RequestEngineConfig) *net.Dialer`: returns dialer with Control hook; passes `cfg.AllowLocalhost` to `isBlockedIP`
- `var blockedCIDRs []*net.IPNet`: package-level slice, initialized in `init()` from the CIDR list above
- Separated into its own file for testability: unit tests can verify each blocked range independently

## Modified Files

### 5. internal/plugin/manager.go

- Add `requestEngine *RequestEngine` field + `RequestEngine()` accessor
- `NewManager()`: create engine after `NewHookEngine()` (at ~line 198), matching the existing pattern:
```go
mgr.requestEngine = NewRequestEngine(pool, dialect, RequestEngineConfig{
    DefaultTimeoutSec:      cfg.RequestTimeoutSec,
    MaxResponseBytes:       cfg.RequestMaxResponseBytes,
    MaxRequestBodyBytes:    cfg.RequestMaxRequestBodyBytes,
    MaxRequestsPerMin:      cfg.RequestMaxPerMin,
    GlobalMaxRequestsPerMin: cfg.RequestGlobalMaxPerMin,
    CBMaxFailures:          cfg.RequestCBMaxFailures,
    CBResetIntervalSec:     cfg.RequestCBResetIntervalSec,
    AllowLocalhost:         cfg.RequestAllowLocalhost,
})
```

**PluginInstance changes:**
- Add `requestAPIs map[*lua.LState]*requestAPIState` field to `PluginInstance` struct (parallel to existing `dbAPIs map[*lua.LState]*DatabaseAPI`)
- Protected by existing `inst.mu` mutex (same lock protects both maps)
- Initialize in `loadPlugin()` factory: `requestAPIs: make(map[*lua.LState]*requestAPIState)` (set on `inst` alongside `inst.dbAPIs[L] = dbAPI`)
- Initialize in `ReloadPlugin()` new instance (at `manager.go:985`, alongside existing `dbAPIs: make(...)`): add `requestAPIs: make(map[*lua.LState]*requestAPIState)` to the `PluginInstance` struct literal

**loadPlugin() factory function changes:**
- After `RegisterRequestAPI(L, pluginName, m.requestEngine)`, store the returned `*requestAPIState` in `inst.requestAPIs[L] = reqAPI` (under `inst.mu` lock, same as `inst.dbAPIs[L] = dbAPI`)
- Add `FreezeModule(L, "request")` after hooks

**OnReplace callback changes:**
- Add `delete(inst.requestAPIs, oldL)` alongside existing `delete(inst.dbAPIs, oldL)` inside the `OnReplace` callback, under the same `inst.mu` lock. This prevents stale `requestAPIState` references when unhealthy VMs are replaced.

**Phase tracking consolidation:**
- Replace the existing `in_init` registry key (currently set in `manager.go:487`, checked in `http_api.go:88`, `http_api.go:149`, and `hooks_api.go:85`) with a single `__vm_phase` registry key. Values: `"module_scope"` (during factory/init.lua execution), `"init"` (during `on_init()`), `"runtime"` (normal execution), `"shutdown"` (during `on_shutdown()`).
- Factory function sets `__vm_phase = "module_scope"` immediately after `ApplySandbox(L, ...)` and before any API registration calls (`RegisterDBAPI`, `RegisterLogAPI`, `RegisterHTTPAPI`, `RegisterHooksAPI`, `RegisterRequestAPI`). This ensures the phase is set before `DoFile(inst.InitPath)` executes `init.lua`, which may call `request.register()`.
- Before `on_init()`: set `__vm_phase = "init"` (replaces `L.SetField(regTbl, "in_init", lua.LTrue)`).
- After `on_init()` returns: set `__vm_phase = "runtime"` (replaces `L.SetField(regTbl, "in_init", lua.LNil)`).
- `http.handle()` phase guard (in `http_api.go:88`): check `__vm_phase == "init"` instead of `in_init == LTrue`.
- `http.use()` phase guard (in `http_api.go:149`): check `__vm_phase == "init"` instead of `in_init == LTrue`. This is a third check site that must not be missed -- leaving it unchanged would silently disable the `http.use()` phase guard since `in_init` is never set.
- `hooks.on()` phase guard (in `hooks_api.go:85`): same change as `http.handle()`.
- `request.register()`: check `__vm_phase == "module_scope"` (registration only during module scope).
- `request.send()`: check `__vm_phase == "runtime"` (only value that permits outbound calls).
- This eliminates two independent sources of lifecycle truth, preventing desync during error paths.

**Test file updates for phase tracking refactor:**
- `http_api_test.go:338`: change `L.SetField(reg, "in_init", lua.LTrue)` to set `__vm_phase = "init"` on the registry table.
- `http_api_test.go:359-381`: change `L.SetField(reg, "in_init", lua.LFalse)` to set `__vm_phase = "runtime"` (or simply do not set the phase, since `"runtime"` is the post-init default).
- `hooks_api_test.go:194`: change `L.SetField(regTbl, "in_init", lua.LTrue)` to set `__vm_phase = "init"` on the registry table.

**Other changes:**
- `LoadAll()`: after hook table setup, create `plugin_requests` table, register in tables registry, cleanup orphans
- `loadPlugin()` post-hooks: read `__request_pending`, call `UpsertRequestRegistrations()`. Note: `ReadPendingRequests(L)` is NOT called during manifest extraction (`ExtractManifest` / `registerManifestStubs`), only during full `loadPlugin()`.
- `registerManifestStubs()` (at `manager.go:906`): add `"request"` to the string slice `[]string{"db", "log", "http", "hooks"}` so it becomes `[]string{"db", "log", "http", "hooks", "request"}` (noop stub; `ReadPendingRequests` is never called on stub VMs)
- `ListOrphanedTables()` (at `manager.go:1133`): add `"plugin_requests": true` to `systemTables` map (alongside existing `"plugin_routes"` and `"plugin_hooks"` entries). Without this, the cleanup endpoint would flag `plugin_requests` as orphaned and could drop it. Note: `DropOrphanedTables` calls `ListOrphanedTables()` internally, so the exclusion propagates automatically -- no separate change needed there.
- `ReloadPlugin()`: call `m.requestEngine.UnregisterPlugin(name)` in the same block as the existing unregister calls (after `bridge.UnregisterPlugin(name)` at line 1013 and `hookEngine.UnregisterPlugin(name)` at line 1016), guarded by a nil check:
```go
if m.requestEngine != nil {
    m.requestEngine.UnregisterPlugin(name)
}
```
Note: there is a brief window during reload where the approved map is empty for the plugin (between `UnregisterPlugin` clearing the map and `UpsertRequestRegistrations` repopulating from DB). Any concurrent `request.send()` during this window will get `{error = "domain not approved"}`. This is acceptable and matches the equivalent route registration window in `HTTPBridge`.
- `shutdownPlugin()` (at `manager.go:632-674`): before calling `on_shutdown()` (line ~662), set `__vm_phase = "shutdown"` on the checked-out LState's registry table. This ensures `request.send()` raises a Lua error if called during `on_shutdown()`, and also blocks `http.handle()`, `http.use()`, and `hooks.on()` registration during shutdown. Insert after `L.SetContext(shutdownCtx)` and before the `on_shutdown` global lookup.
- `ManagerConfig`: add `RequestTimeoutSec`, `RequestMaxResponseBytes`, `RequestMaxRequestBodyBytes`, `RequestMaxPerMin`, `RequestGlobalMaxPerMin`, `RequestCBMaxFailures`, `RequestCBResetIntervalSec`, `RequestAllowLocalhost`

### 6. internal/plugin/hook_engine.go

- `executeBefore()`: after looking up `dbAPI` via `inst.dbAPIs[L]` (existing pattern at `hook_engine.go:380-387`), also look up `reqAPI` via `inst.requestAPIs[L]` under the same `inst.mu.Lock()` call. The map lookup is protected by the lock; the flag setting happens after `inst.mu.Unlock()` (matching the existing `dbAPI` pattern -- the flag is per-VM so no concurrent access once the correct instance is retrieved):
```go
inst.mu.Lock()
dbAPI := inst.dbAPIs[L]
reqAPI := inst.requestAPIs[L]
inst.mu.Unlock()

if dbAPI != nil {
    dbAPI.inBeforeHook = true
    defer func() { dbAPI.inBeforeHook = false }()
}
if reqAPI != nil {
    reqAPI.inBeforeHook = true
    defer func() { reqAPI.inBeforeHook = false }()
}
```

### 7. internal/plugin/http_bridge.go

- `MountAdminEndpoints()`: register 3 new endpoints in the "literal-path endpoints" section (before the wildcard `{name}` endpoints, matching the hooks registration pattern at lines 1005-1010):
```go
mux.Handle("GET /api/v1/admin/plugins/requests",
    authChain(PluginRequestsListHandler(mgr)))
mux.Handle("POST /api/v1/admin/plugins/requests/approve",
    authChain(adminOnlyFn(PluginRequestsApproveHandler(mgr))))
mux.Handle("POST /api/v1/admin/plugins/requests/revoke",
    authChain(adminOnlyFn(PluginRequestsRevokeHandler(mgr))))
```
Note: GET uses `authChain` only (any authenticated user can view); POST uses `authChain(adminOnlyFn(...))` (admin-only for approve/revoke). This matches the hooks endpoint auth pattern exactly.
- `ServeHTTP()`: no `__vm_phase` changes needed. VMs are snapshotted after `on_init()` completes, at which point `__vm_phase` is already `"runtime"`. Checked-out VMs inherit this value from the snapshot. Do NOT reset phase after handler returns -- the VM goes back to the pool and must remain in `"runtime"` phase for subsequent checkouts.

### 8. internal/plugin/metrics.go

- Add `PluginMetricOutboundRequests`, `PluginMetricOutboundDuration`, `PluginMetricOutboundCircuitBreak` constants
- Add `RecordOutboundRequest(pluginName, method, domain, status, durationMs)` -- logs only plugin name, method, domain, status code, and duration. MUST NOT log request headers (which contain Authorization tokens), request body, or response body.

### 9. internal/config/config.go

- Add fields:
  - `Plugin_Request_Timeout int` (default 10)
  - `Plugin_Request_Max_Response int64` (default 1_048_576 = 1MB)
  - `Plugin_Request_Max_Body int64` (default 1_048_576 = 1MB)
  - `Plugin_Request_Rate_Limit int` (default 60)
  - `Plugin_Request_Global_Rate_Limit int` (default 600; 0 = unlimited)
  - `Plugin_Request_CB_Max_Failures int` (default 5)
  - `Plugin_Request_CB_Reset_Interval int` (default 60)
  - `Plugin_Request_Allow_Localhost bool` (default false; development only -- allows HTTP to 127.0.0.1/::1)

### 10. cmd/helpers.go

- `initPluginManager()`: pass all new config fields to `ManagerConfig` (including `RequestGlobalMaxPerMin` and `RequestAllowLocalhost`)

### 11. cmd/serve.go

- Shutdown sequence: add `requestEngine.Close()` AFTER `hookEngine.Close()`, not before it. The existing shutdown order (at `serve.go:323-339`) is: `StopWatcher` -> `bridge.Close` -> `hookEngine.Close` -> `pluginManager.Shutdown`. The new order is: `StopWatcher` -> `bridge.Close` -> `hookEngine.Close` -> `requestEngine.Close()` -> `pluginManager.Shutdown`. Insert after the `hookEngine.Close(shutdownCtx)` call and before `pluginManager.Shutdown(shutdownCtx)`:
```go
if pluginManager != nil {
    if reqEngine := pluginManager.RequestEngine(); reqEngine != nil {
        reqEngine.Close()
    }
}
```
Rationale: `hookEngine.Close()` drains its `afterWG` (waits for all after-hook goroutines to complete). After-hook goroutines may call `request.send()`. If `requestEngine.Close()` runs first, it calls `client.CloseIdleConnections()` which could yank connections from under in-flight requests in after-hooks, producing confusing "connection reset" errors and false circuit breaker failures.
- `requestEngine.Close()` cancels idle connections and stops the rate limiter cleanup goroutine. In-flight requests (if any remain despite the ordering) are governed by their per-request `context.WithTimeout` and will complete or timeout naturally.

## Security

- **HTTPS enforced**: All outbound requests require HTTPS. HTTP is only permitted to `127.0.0.1`/`::1` when `AllowLocalhost` is explicitly enabled in config (default: false, development only). Enabling in production is dangerous -- it allows plugins to probe localhost services.
- **TLS certificate verification**: Uses Go's default TLS configuration, which validates server certificates against the host OS trust store. No custom CA support is provided. In environments with corporate proxies or internal CAs, plugins calling internal HTTPS services will get TLS errors (which the circuit breaker will trip on). Deployers must add custom CAs to the host OS trust store or use the standard `SSL_CERT_FILE` / `SSL_CERT_DIR` environment variables that Go respects.
- **SSRF protection**: dialer Control hook blocks connections to private/reserved IP ranges after DNS resolution. Catches DNS rebinding attacks. Blocked ranges: RFC 1918, link-local, loopback, CGN, cloud metadata (169.254.0.0/16). Localhost exception is opt-in via `AllowLocalhost`.
- **Domain matching**: URL parsed via `net/url.Parse`, hostname extracted via `url.Hostname()`, exact case-insensitive match against approved domains. No suffix/prefix/wildcard matching. URLs with userinfo (`user:pass@host`) are rejected entirely (common SSRF parsing confusion vector).
- **Domain-level approval**: default-deny, admin must approve each domain per plugin
- **Version stability**: version changes do NOT revoke approvals (matches hook_engine pattern). Version is recorded in DB for audit. If route-bridge-style revocation is desired in the future, it can be added as a config flag.
- **Per-plugin-per-domain circuit breaker**: 5 consecutive failures trips the breaker. 60s reset interval with half-open probe. Prevents VM pool exhaustion from dead external services.
- **Per-plugin-per-domain rate limiting**: 60 requests/min per plugin per domain (not shared across domains). Uses `golang.org/x/time/rate` token bucket. Burst capped at 5 regardless of rate.
- **Global rate limiting**: aggregate cap of 600 requests/min across all plugins/domains. Prevents excessive outbound traffic from the CMS server.
- **Response size limit**: `io.LimitReader` at `MaxResponseBytes` (default 1MB)
- **Request body size limit**: checked before sending, default 1MB
- **Request timeout**: `context.WithTimeout`, capped by engine max (default 10s)
- **Before-hook blocking**: `request.send()` raises Lua error if called inside before-hook handler (prevents holding DB transaction locks during outbound HTTP calls)
- **Shutdown-phase blocking**: `request.send()` raises Lua error if called during `on_shutdown()` (`__vm_phase == "shutdown"`). Request engine is closed after hook engine drains, so this guard prevents calls after transport teardown.
- **Unified phase tracking**: Single `__vm_phase` registry key replaces the `in_init` key. Prevents desync between phase guards during error paths.
- **No redirect following**: returns 3xx responses as-is (prevents open-redirect SSRF amplification)
- **User-Agent**: always set to `ModulaCMS-Plugin/1.0`
- **Correlation ID**: `Execute()` extracts the request ID from its `ctx` parameter via `middleware.RequestIDFromContext(ctx)` (defined in `internal/middleware/request_id.go:27`, already takes `context.Context`). If non-empty, it is forwarded as an `X-Request-ID` header on the outbound request. This enables tracing "inbound request -> plugin hook -> outbound webhook" across systems. If no request ID exists in context, the header is omitted. Import: `"github.com/hegner123/modulacms/internal/middleware"`.
- **No credential logging**: `RecordOutboundRequest` logs only plugin name, method, domain, status code, and duration. Request/response headers and bodies are never logged.
- **JSON parsing opt-in**: response body JSON parsing requires explicit `parse_json = true` in opts. Prevents server-controlled Content-Type headers from influencing processing behavior.
- **Future: plugin secrets store**: Plugin credentials (API keys, tokens) should not be hardcoded in Lua source. A future enhancement will provide a per-plugin secrets store where admins set key-value pairs that are injected at runtime. Until then, plugin authors should use environment-aware configuration files outside of version control.

## Design Decisions

### Domain limit: 50 per plugin
Matches `MaxRoutesPerPlugin` (50) and `MaxHooksPerPlugin` (100, but hooks have event+table combinations). 50 distinct external domains covers even ambitious integration plugins (Zapier-style). A plugin needing more than 50 domains is likely doing something wrong or should be split. Note: there is no aggregate cap across all plugins. Memory footprint scales with `plugins * domains_per_plugin` for rate limiter and circuit breaker entries. At typical scale (< 20 plugins) this is negligible; the rate limiter cleanup goroutine removes idle entries after 10 minutes.

### No retries in engine
See Lua API Surface section. Plugin authors implement retries in Lua with full control.

### Version changes don't revoke approvals
The hook_engine pattern (upsert updates version, preserves approval) is the right model here. Routes use revocation-on-version-change because routes are inbound security boundaries (a new version might expose different URL patterns). Outbound request domains are an operational concern -- a plugin that talked to `api.stripe.com` in v1.0 almost certainly needs it in v1.1 too. Silent revocation causes invisible webhook failures with no admin notification, which is worse than the marginal security benefit.

### Per-plugin-per-domain rate limits and circuit breakers
A plugin integrating with Stripe, SendGrid, and Slack should not have a dead Stripe API cause rate limiting or circuit breaking on the healthy Slack and SendGrid connections. The key format `"plugin:domain"` also provides isolation between plugins -- one plugin's abuse of a domain does not consume another plugin's budget. A global aggregate rate limiter (default 600/min) acts as a backstop to prevent the CMS server from generating excessive outbound traffic regardless of individual plugin limits.

### SSRF check at dialer level (not URL level)
URL-level checks (parsing the hostname and checking if it's a private IP) miss DNS rebinding attacks where a public domain resolves to a private IP. The dialer Control hook fires after DNS resolution but before TCP connection, catching this attack vector.

## Verification

1. `just check` -- compile check
2. `go test -v ./internal/plugin/ -run TestRequest` -- request API and engine tests
3. `go test -v ./internal/plugin/ -run TestSSRF` -- SSRF protection unit tests (each blocked CIDR range)
4. `go test -v ./internal/plugin/ -run TestCircuitBreaker` -- circuit breaker state transitions
5. `just test` -- full test suite
6. Manual: create a test plugin with `request.register("httpbin.org")`, approve via admin API, verify `request.get("https://httpbin.org/get")` works, revoke and verify it returns `{error = "domain not approved: httpbin.org"}`
