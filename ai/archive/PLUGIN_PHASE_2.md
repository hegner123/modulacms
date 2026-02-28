Phase 2: Plugin HTTP Integration

Context

Phase 1 of the plugin system (Core Engine) is fully implemented and tested: Manager, VMPool, Sandbox, DatabaseAPI, SchemaAPI, LogAPI, and LuaHelpers are all working with comprehensive test coverage. Plugins can define tables, query data, and manage lifecycle via on_init/on_shutdown.

Phase 2 adds HTTP route registration so plugins can serve REST endpoints via the existing net/http.ServeMux. This is the next step in ai/PLUGIN_SYSTEM_ROADMAP.md (lines 813-874).

Trust Model

Plugins are installed by CMS administrators only. There is no plugin marketplace and no remote installation mechanism. Administrators place plugin directories containing init.lua into the configured Plugin_Directory on the server filesystem. This means:

- Plugin code is admin-vetted, not untrusted third-party code
- The primary threat model is buggy plugins, not malicious ones
- Security controls still enforce defense in depth — a bug should not compromise the CMS
- All HTTP endpoints registered by plugins require explicit admin approval before they become reachable

Route Access Control

Plugin endpoints serve two audiences with different security postures:

1. Authenticated routes (default) — for CMS frontend consumption. Require CMS session authentication (cookie or API key) via the existing HTTPPublicEndpointMiddleware. Plugin-level Lua middleware can add further authorization (role checks, etc.) but is not responsible for primary authentication.

2. Public routes — for third-party integrations, webhooks, external consumers. Bypass CMS session authentication. Plugin-level Lua middleware is fully responsible for any authentication (API keys, signatures, etc.).

Route type is declared per-handler via an options table:

    -- Authenticated (default): requires CMS session
    http.handle("GET", "/tasks", handler)
    http.handle("GET", "/tasks", handler, {})          -- equivalent
    http.handle("GET", "/tasks", handler, {public = false})  -- equivalent

    -- Public: no CMS session required
    http.handle("POST", "/webhook", handler, {public = true})

Admin Approval:

All routes start as unapproved (pending). Unapproved routes return 404 ROUTE_NOT_FOUND (not 403, to prevent enumeration). The admin approval workflow:

1. Plugin loads, registers routes via http.handle() at module scope
2. RegisterRoutes extracts metadata, upserts into plugin_routes DB table with approved = false for new routes. If an existing approved route changes its public flag (e.g., authenticated -> public), approval is revoked and the route returns to pending. This prevents a plugin update from silently widening the access surface without admin review.
3. Only previously-approved routes are mounted on ServeMux
4. Admin reviews pending routes via admin API and approves or rejects
5. On approval, the bridge registers the route on ServeMux immediately (Go 1.22+ ServeMux is internally synchronized for concurrent Handle calls).
6. On revocation, the bridge sets route.Approved = false in memory and updates the DB. The route pattern remains registered on ServeMux (cannot be unregistered at runtime), but ServeHTTP checks the in-memory approval flag and returns 404 for revoked routes. Full cleanup (removing stale ServeMux patterns) takes effect on next server restart.

Approval lifecycle on plugin changes:
- Plugin version change (detected by comparing PluginInfo.Version in RegisterRoutes): all routes for that plugin are revoked and must be re-approved. Old plugin_routes rows are deleted, new ones inserted as unapproved.
- Plugin removal (directory deleted, server restarted): orphaned plugin_routes rows are cleaned up by the Manager at the start of LoadAll() — any plugin_name not present in the discovered set has its rows deleted.
- Route public flag change: approval revoked for that specific route (see step 2 above).

Admin API endpoints (require admin role — see authorization below):

    GET  /api/v1/admin/plugins/routes              — list all plugin routes with approval status
    POST /api/v1/admin/plugins/routes/approve       — approve route(s)
    POST /api/v1/admin/plugins/routes/revoke        — revoke route approval

Request/response format:

    // POST /api/v1/admin/plugins/routes/approve
    // Body:
    {"routes": [{"plugin": "task_tracker", "method": "GET", "path": "/tasks"}]}

    // GET /api/v1/admin/plugins/routes
    // Response:
    {"routes": [
      {"plugin": "task_tracker", "method": "GET", "path": "/tasks",
       "public": false, "approved": true, "approved_at": "...", "approved_by": "..."},
      {"plugin": "task_tracker", "method": "POST", "path": "/webhook",
       "public": true, "approved": false, "approved_at": null, "approved_by": null}
    ]}

plugin_routes DB table (created by Manager at startup via raw SQL on the plugin DB pool):

    CREATE TABLE IF NOT EXISTS plugin_routes (
        plugin_name     TEXT NOT NULL,
        method          TEXT NOT NULL,
        path            TEXT NOT NULL,  -- plugin-relative path, e.g. "/tasks"
        public          INTEGER NOT NULL DEFAULT 0,
        approved        INTEGER NOT NULL DEFAULT 0,
        approved_at     TEXT,  -- ISO 8601 timestamp
        approved_by     TEXT,  -- user ID of approving admin
        plugin_version  TEXT NOT NULL DEFAULT '',  -- from PluginInfo.Version, triggers re-approval on change
        created_at      TEXT NOT NULL DEFAULT (datetime('now')),
        PRIMARY KEY (plugin_name, method, path)
    );

CreatePluginRoutesTable() must implement three dialect variants (same table, dialect-appropriate DDL):

SQLite: as shown above (TEXT types, datetime('now') default).
MySQL: TEXT columns, BOOLEAN NOT NULL DEFAULT FALSE for public/approved, DATETIME DEFAULT CURRENT_TIMESTAMP for created_at, composite PRIMARY KEY same as above.
PostgreSQL: TEXT columns, BOOLEAN NOT NULL DEFAULT FALSE for public/approved, TIMESTAMPTZ DEFAULT NOW() for created_at, composite PRIMARY KEY same as above.

The Manager detects the dialect from its existing m.dialect field (set during initPluginPool) and executes the appropriate CREATE TABLE IF NOT EXISTS. This follows the same pattern as CreateAllTables() in internal/db/db.go which switches on dialect. The Manager calls CreatePluginRoutesTable() during startup before LoadAll(). LoadAll() also calls CleanupOrphanedRoutes() which deletes rows for plugin_names not present in the discovered plugin set.

Critical Architectural Decision: Handler Function Binding

Problem: *lua.LFunction references are bound to their creating LState. A handler function created on VM-A cannot be called on VM-B. Since the VMPool has N VMs (default 4), each created by the factory which runs init.lua, each VM has its own copy of the handler functions.

Solution: Per-VM handler registry with Go-side route metadata.

1. http.handle() is called at module scope in init.lua (NOT inside on_init()). This differs from db.define_table() which must be inside on_init(). The reason: every VM executes init.lua via the factory, so every VM registers its own handler functions. The Go bridge deduplicates route metadata.
2. Each VM stores handlers in a Lua table __http_handlers (keyed by METHOD /path). The http.handle() Go-bound function writes to this table.
3. The HTTPBridge stores route metadata once (method, path, plugin name, access type, approval status). On request: bridge checks approval, checks auth if authenticated route, checks out a VM, looks up the handler from that VM's __http_handlers table, calls it, writes the HTTP response, returns the VM.
4. http.use() middleware follows the same pattern: stored in __http_middleware Lua table per-VM, called in order before the handler.

Runtime guard: The http.handle() Go-bound function checks a phase flag on the LState (set during on_init execution). If http.handle() is called during on_init, it raises error("http.handle() must be called at module scope, not inside on_init()"). This prevents the silent failure mode where only one VM gets the handler function. The flag is set via a Go-side boolean stored in the LState's registry table before calling on_init(), and cleared after on_init() returns.

Request Lifecycle

HTTP request -> ServeMux matches registered pattern "GET /api/v1/plugins/task_tracker/tasks/{id}"
  -> HTTPBridge.ServeHTTP()
    -> Set default security headers (X-Content-Type-Options: nosniff, X-Frame-Options: DENY)
    -> Extract plugin name from URL prefix (/api/v1/plugins/<name>/)
    -> Look up route in bridge's route registry (method + full path -> plugin name)
    -> If no route match: 404 ROUTE_NOT_FOUND (uniform for all unmatched paths)
    -> Check route.Approved -- if false: 404 ROUTE_NOT_FOUND (not 403, prevents enumeration)
    -> If route is authenticated: check CMS auth context from middleware
       -> If unauthenticated: 401 UNAUTHORIZED
    -> If route is public: skip CMS auth check (Lua middleware handles its own auth)
    -> Rate limit check (Plugin_Rate_Limit per IP)
    -> bridge.inflight.Add(1); defer bridge.inflight.Done()
    -> Check bridge.closing atomic flag -- if true, return 503 PLUGIN_UNAVAILABLE
    -> Apply body size limit: r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
    -> Create scoped context: execCtx, cancel := context.WithTimeout(r.Context(), execTimeout)
    -> pool.Get(execCtx) -> checkout VM
    -> dbAPI.ResetOpCount()
    -> L.SetContext(execCtx) -- propagates deadline into Lua VM execution
    -> Build LuaRequest table from *http.Request (using r.PathValue() for path params)
    -> Run middleware chain from __http_middleware (shared req table, mutations propagate)
    -> Look up handler: L.GetField(__http_handlers, "GET /tasks")
    -> L.CallByParam(handler, luaReqTable)
    -> Read LuaResponse from return value
    -> Enforce response size limit (MaxPluginResponseBody)
    -> Filter blocked response headers
    -> Write HTTP response (status, headers, JSON body)
    -> Audit log via audited command pattern (internal/db/audited/): plugin name, method, path, status, duration, request ID. Records a change_event with operation type, request metadata, and timestamp.
    -> pool.Put(L) -> return VM

Timeout layers (innermost to outermost):
1. 100ms -- pool.Get() acquisition timeout (fast-fail when all VMs busy)
2. 5s -- ExecTimeoutSec via L.SetContext(execCtx) (Lua execution deadline)
3. 15s -- http.Server.WriteTimeout (backstop, kills connection)

Error Response Schema

All bridge error responses use a consistent JSON schema, distinct from the core router's http.Error() plain text. Plugin consumers are external developers who need machine-readable errors.

type PluginErrorResponse struct {
    Error PluginErrorDetail `json:"error"`
}

type PluginErrorDetail struct {
    Code      string `json:"code"`
    Message   string `json:"message"`
    RequestID string `json:"request_id,omitempty"` // from RequestIDMiddleware
}

Error codes:

| Status | Code               | When                              |
|--------|--------------------|-----------------------------------|
| 400    | INVALID_REQUEST    | Body too large, malformed JSON    |
| 401    | UNAUTHORIZED       | Authenticated route, no CMS session |
| 404    | ROUTE_NOT_FOUND    | No matching route OR unapproved route (uniform) |
| 500    | HANDLER_ERROR      | Lua error() in handler            |
| 503    | PLUGIN_UNAVAILABLE | Plugin not in StateRunning        |
| 503    | POOL_EXHAUSTED     | All VMs busy (+ Retry-After: 1)   |
| 504    | HANDLER_TIMEOUT    | Context deadline exceeded         |

Note on 405: If a client sends an HTTP method that is not registered for a given path (e.g., PATCH to a route with only GET/POST), ServeMux may return its own plain-text 405. To maintain JSON consistency, the fallback handler at "/api/v1/plugins/" catches all unmatched method+path combos and returns 404 ROUTE_NOT_FOUND in JSON. The bridge does NOT attempt to return 405 — this avoids revealing which methods are registered (enumeration prevention). All unmatched requests are uniformly 404.

Helper: writePluginError(w, status, code, message, requestID) sets Content-Type: application/json and encodes the struct. HANDLER_ERROR responses always return the generic message "internal plugin error" to clients. The full Lua error with stack trace is logged server-side with the request ID for operator correlation.

Size Indicators

S (~50-100 lines changed), M (~100-300), L (~300-500), XL (~500+). These estimate net lines of Go code per file, not counting tests.

New Files

1. internal/plugin/http_bridge.go -- XL

Central coordinator between Manager and ServeMux. Implements http.Handler.

const (
    MaxPluginRequestBody  = 1 << 20      // 1 MB default for request bodies
    MaxPluginResponseBody = 5 << 20      // 5 MB default for response bodies
    DefaultPluginRateLimit = 100         // requests per second per IP
    MaxRoutesPerPlugin    = 50           // maximum routes a single plugin can register
)

// blockedResponseHeaders cannot be set by Lua plugins.
// Prevents CORS bypass, session fixation, and response smuggling.
var blockedResponseHeaders = map[string]bool{
    "access-control-allow-origin":      true,
    "access-control-allow-credentials": true,
    "access-control-allow-methods":     true,
    "access-control-allow-headers":     true,
    "access-control-expose-headers":    true,
    "set-cookie":                       true,
    "transfer-encoding":               true,
    "content-length":                  true,
    "host":                            true,
    "connection":                      true,
    "cache-control":                   true,
}

type RouteRegistration struct {
    Method     string // GET, POST, PUT, DELETE, PATCH
    Path       string // plugin-relative, e.g., "/tasks" or "/tasks/{id}"
    PluginName string
    FullPath   string // computed: /api/v1/plugins/<plugin>/path
    Public     bool   // true = bypass CMS auth, false = require CMS session
    Approved   bool   // only approved routes are dispatched; unapproved return 404
}

type HTTPBridge struct {
    manager       *Manager
    routes        map[string]*RouteRegistration // "GET /api/v1/plugins/task_tracker/tasks" -> reg
    mu            sync.RWMutex
    closing       atomic.Bool    // set during shutdown, new requests get 503
    inflight      sync.WaitGroup // tracks in-progress requests for graceful drain
    maxBodySize   int64          // from config, defaults to MaxPluginRequestBody
    maxRespSize   int64          // from config, defaults to MaxPluginResponseBody
    rateLimiter   *middleware.RateLimiter // bridge creates its own instance via middleware.NewRateLimiter(rate.Limit(cfg.Plugin_Rate_Limit), cfg.Plugin_Rate_Limit) -- burst equals rate
    mux           *http.ServeMux // reference for live route registration on approval
    trustedNets   []*net.IPNet   // parsed from Plugin_Trusted_Proxies config at init
}

Key methods:
- NewHTTPBridge(manager *Manager, cfg config.Config) *HTTPBridge
- RegisterRoutes(pluginName string, L *lua.LState) -- called once per plugin inside loadPlugin, between SnapshotGlobals and pool.Put; reads __http_handlers table, upserts into plugin_routes DB table, mounts approved routes
- MountOn(mux *http.ServeMux) -- stores mux reference, registers approved routes as real ServeMux patterns plus fallback "/api/v1/plugins/" for 404s
- ApproveRoute(pluginName, method, path string, approvedBy string) error -- updates DB, registers on ServeMux if not already mounted
- RevokeRoute(pluginName, method, path string) error -- updates DB, marks route unapproved in memory
- ListRoutes() []RouteRegistration -- returns all routes with approval status
- ServeHTTP(w http.ResponseWriter, r *http.Request) -- the main dispatch
- Close(ctx context.Context) -- graceful shutdown

Route path validation: http.handle() validates the path argument before registration:
- Must start with "/"
- Must not contain ".." (path traversal)
- Must not contain "?" or "#" (query string/fragment characters)
- Must not exceed 256 characters
- Must match [a-zA-Z0-9/_{}.-] (alphanumeric, slashes, braces for params, dots, hyphens)
- Violations raise error() at registration time

Path Parameter Matching: Use Go 1.22+ ServeMux directly, NOT a catch-all.

Instead of a single catch-all handler, MountOn() registers each approved plugin route as a real ServeMux pattern:

    mux.Handle("GET /api/v1/plugins/task_tracker/tasks", bridge)
    mux.Handle("GET /api/v1/plugins/task_tracker/tasks/{id}", bridge)
    mux.Handle("POST /api/v1/plugins/task_tracker/tasks", bridge)
    mux.Handle("/api/v1/plugins/", bridge) // fallback for uniform 404

Inside ServeHTTP(), use r.PathValue("id") to populate req.params. This avoids implementing a custom path matcher inside the bridge. ServeMux handles precedence ("/tasks/{id}" vs "/tasks/special") and trailing slash normalization.

The route map key format matches the ServeMux pattern for O(1) lookup when dispatching: the bridge reads r.Pattern (available since Go 1.22, our go.mod specifies Go 1.26) to identify the RouteRegistration, then extracts path values for any {param} segments. r.Pattern is set by ServeMux to the matched pattern string (e.g., "GET /api/v1/plugins/task_tracker/tasks/{id}") and maps directly to the bridge.routes key:

    route, ok := b.routes[r.Pattern]  // O(1) lookup
    if !ok {
        writePluginError(w, 404, "ROUTE_NOT_FOUND", ...)
        return
    }

The fallback handler at "/api/v1/plugins/" returns a uniform 404 ROUTE_NOT_FOUND for ALL unmatched requests regardless of whether the plugin exists, whether the plugin is running, or whether the route is unapproved. This prevents plugin enumeration.

Graceful Shutdown:

Close(ctx context.Context) mirrors the server's httpServer.Shutdown() pattern from cmd/serve.go:

1. bridge.closing.Store(true) -- new requests get immediate 503 PLUGIN_UNAVAILABLE
2. Wait for bridge.inflight (tracked via sync.WaitGroup incremented/decremented in ServeHTTP())
3. ctx deadline provides the backstop (uses the shutdown context's 30s timeout from serve.go)

Call bridge.Close(shutdownCtx) BEFORE manager.Shutdown() in serve.go. This ensures no new VM checkouts happen before pools close. The existing VMPool.Close() drain safety (the closed atomic flag + direct L.Close() on Put) handles VMs returned after pool closure.

2. internal/plugin/http_api.go -- L

Registers http Lua module with handle() and use() functions.

Lua API (called at module scope in init.lua):

-- Authenticated route (default): requires CMS session
http.handle("GET", "/tasks", function(req)
    local tasks = db.query("tasks", {order_by = "created_at", limit = 50})
    return {status = 200, json = tasks}
end)

http.handle("POST", "/tasks", function(req)
    local body = req.json
    db.insert("tasks", {title = body.title, status = "pending"})
    return {status = 201, json = {message = "created"}}
end)

http.handle("GET", "/tasks/{id}", function(req)
    local task = db.query_one("tasks", {where = {id = req.params.id}})
    if not task then
        return {status = 404, json = {error = "not found"}}
    end
    return {status = 200, json = task}
end)

-- Public route: no CMS session required, plugin manages its own auth
http.handle("POST", "/webhook", function(req)
    local sig = req.headers["x-webhook-signature"]
    if not verify_signature(sig, req.body) then
        return {status = 401, json = {error = "invalid signature"}}
    end
    db.insert("events", {payload = req.body})
    return {status = 200, json = {ok = true}}
end, {public = true})

-- Plugin-scoped middleware (runs before all handlers in this plugin)
-- Middleware and handler share the same Lua request table (pass-by-reference).
-- Mutations propagate to subsequent middleware and the handler.
http.use(function(req)
    req.request_start = os.clock()
    return nil  -- continue to handler
end)

Go types:
type LuaRequest struct {
    Method   string
    Path     string
    Body     string
    ClientIP string            // proxy-aware IP (same extraction as rate limiter), no port
    Headers  map[string]string // normalized to lowercase keys
    Query    map[string]string
    Params   map[string]string // path parameters from {id} patterns via r.PathValue()
    JSON     any               // parsed JSON body (only when Content-Type: application/json), nil otherwise
}

type LuaResponse struct {
    Status  int
    Headers map[string]string // filtered through blockedResponseHeaders
    Body    string            // raw body (mutually exclusive with JSON)
    JSON    any               // marshaled to JSON response
}

If both body and json are present in the Lua response table, json takes priority and body is ignored. Log a warning: "plugin %s handler returned both body and json; using json".

Note: RemoteAddr is NOT exposed to Lua. Instead, ClientIP provides a proxy-aware client IP with port stripped. This reduces PII surface and gives plugins a consistent IP regardless of reverse proxy topology.

ClientIP extraction (shared by rate limiter and LuaRequest): uses a trusted proxy-aware algorithm, NOT the existing getIP() which blindly trusts X-Forwarded-For. The implementation:
1. If Plugin_Trusted_Proxies config is empty (default): use r.RemoteAddr only, strip port. Ignore X-Forwarded-For and X-Real-IP entirely. This is the safe default for deployments without a reverse proxy.
2. If Plugin_Trusted_Proxies is set (e.g., ["10.0.0.0/8", "172.16.0.0/12"]): parse X-Forwarded-For from right to left, skip entries matching trusted proxy CIDRs, return the first non-trusted IP. This prevents spoofing because the rightmost entry is set by the closest proxy.
This is a new function (extractClientIP) in http_bridge.go, NOT a reuse of the existing middleware getIP(). The existing getIP() in ratelimit.go has known spoofing issues and is not modified by Phase 2.

RegisterHTTPAPI(L *lua.LState, pluginName string):
- Creates http global table with handle and use Go-bound functions
- handle accepts 3 required args (method, path, function) and 1 optional arg (options table)
- handle stores handler in __http_handlers Lua table on the VM
- handle stores route metadata (including public flag) in __http_route_meta Lua table on the VM. Key format matches __http_handlers: "METHOD /path" (e.g., "GET /tasks"). Value is a Lua table containing the options passed to http.handle():

    -- __http_route_meta["GET /tasks"] = { public = false }
    -- __http_route_meta["POST /webhook"] = { public = true }

  The Go-bound handle function writes it as:
    meta := L.NewTable()
    meta.RawSetString("public", lua.LBool(isPublic))
    routeMeta.RawSetString(key, meta)  // key = "GET /tasks"

  Using a table (rather than a bare boolean) keeps the structure extensible if future options are added to http.handle() (e.g., rate limit overrides, custom timeout, required roles).
- use stores middleware in __http_middleware Lua table on the VM
- All hidden tables (prefixed with __) are part of the global snapshot

Path validation: http.handle() validates the path argument before registration:
- Must start with "/"
- Must not contain ".." (path traversal)
- Must not contain "?" or "#" (query string/fragment characters)
- Must not exceed 256 characters
- Must match pattern [a-zA-Z0-9/_{}.-] (alphanumeric, slashes, braces for params, dots, hyphens)
- Violations raise error() at registration time:
    L.ArgError(2, "invalid route path: must start with /, max 256 chars, no '..' or query characters")

Method validation: http.handle() validates method against an allowlist before registration:

    var validMethods = map[string]bool{
        "GET": true, "POST": true, "PUT": true,
        "DELETE": true, "PATCH": true,
    }
    if !validMethods[method] {
        L.ArgError(1, "invalid HTTP method: must be GET, POST, PUT, DELETE, or PATCH")
    }

Route count limit: http.handle() tracks registrations per plugin. If a plugin exceeds MaxRoutesPerPlugin (default 50), it raises:

    error("plugin exceeded maximum route limit (50)")

Duplicate route detection: http.handle() checks __http_handlers for an existing key before inserting. If the key "GET /tasks" already exists, it raises error() immediately:

    L.ArgError(1, "duplicate route: GET /tasks already registered")

This follows the existing error convention from db_api.go where programming mistakes (wrong argument types, missing required args) raise error(). Cross-plugin duplicates are detected in RegisterRoutes() and cause StateFailed for the later-loading plugin.

Phase guard: http.handle() and http.use() check a phase flag (stored in the LState's registry table) that is set to true during on_init() execution. If the flag is set, they raise:

    error("http.handle() must be called at module scope, not inside on_init()")

This prevents the silent failure mode where only the first VM (the one that runs on_init) gets the handler function.

3. internal/plugin/http_request.go -- M

Converts *http.Request to Lua table and Lua response table to HTTP response.

- BuildLuaRequest(L *lua.LState, r *http.Request, params map[string]string) *lua.LTable
- WriteLuaResponse(w http.ResponseWriter, L *lua.LState, responseTbl *lua.LTable, maxRespSize int64) error
- Handles JSON parsing of request body (Content-Type: application/json ONLY)
- Handles JSON marshaling of response (when json key is present)
- Normalizes header names to lowercase for Lua access

Content-Type enforcement:
- req.json is populated ONLY when Content-Type header is "application/json" (case-insensitive, charset parameters ignored)
- req.body is always set to the raw body string regardless of Content-Type
- When Content-Type is not application/json, req.json is nil in Lua

Body size enforcement happens BEFORE this function is called (http.MaxBytesReader applied in ServeHTTP). If the body exceeds the limit, json.NewDecoder will return a *http.MaxBytesError that BuildLuaRequest surfaces as a 400 INVALID_REQUEST.

Response size enforcement: WriteLuaResponse marshals the response body (raw string or JSON), then checks the byte length against maxRespSize. If exceeded, returns 500 RESPONSE_TOO_LARGE (logged server-side with request ID) instead of the plugin response.

Response header filtering: WriteLuaResponse iterates the plugin's Headers map and skips any key present in blockedResponseHeaders (compared lowercase). A warning is logged when a plugin attempts to set a blocked header.

Default security headers are set by ServeHTTP BEFORE WriteLuaResponse, so they are always present:
- X-Content-Type-Options: nosniff (NOT overridable by plugins)
- X-Frame-Options: DENY (NOT overridable by plugins)
- Cache-Control: no-store (NOT overridable by plugins — added to blockedResponseHeaders)

Cache-Control is blocked because a buggy plugin setting "Cache-Control: public, max-age=31536000" on user-specific data would cause CDNs and shared proxies to cache private responses. Plugins that need caching should document it and the CMS reverse proxy layer should handle cache headers, not the plugin.

Middleware Request Mutation

Middleware and handler share the same Lua request table (pass-by-reference). Since gopher-lua tables are reference types, middleware can enrich requests by adding or modifying fields:

http.use(function(req)
    req.user_id = validate_api_key(req.headers["x-api-key"])
    return nil  -- continue with enriched req
end)

The handler sees req.user_id set by middleware. This is the intended pattern for auth middleware to communicate identity to handlers, analogous to how Go middleware communicates via r.WithContext().

Security note for plugin authors: middleware execution order is registration order. Later middleware can overwrite fields set by earlier middleware. Keep auth-critical logic in a single middleware to avoid confused deputy issues.

Middleware execution:
- If middleware returns a non-nil table: treat as early response, skip remaining middleware and handler
- If middleware returns nil: continue to next middleware (or handler if last)
- Mutations to req persist across the entire chain

Existing Files to Modify

4. internal/plugin/manager.go -- M

Changes:
- Add HTTPBridge field to Manager struct
- Add shutdownOnce sync.Once field to Manager -- makes Shutdown() idempotent (prevents deadlock if called twice)
- Add CreatePluginRoutesTable() method -- creates the plugin_routes DB table at startup (raw SQL via m.db, dialect-aware)
- Add CleanupOrphanedRoutes(discoveredPlugins []string) method -- deletes plugin_routes rows for plugins not in the discovered set
- Call CreatePluginRoutesTable() at the start of LoadAll, then CleanupOrphanedRoutes() after scanning for plugin directories
- In loadPlugin(), between SnapshotGlobals (line 313) and pool.Put (line 316), call bridge.RegisterRoutes(pluginName, L) to extract route metadata from the checked-out VM
- Set phase flag on LState registry before calling on_init(), clear it after on_init() returns -- this enables the http.handle() phase guard
- Export GetPlugin and the dbAPIs map (or add a helper CheckoutVM(ctx, pluginName) / ReturnVM(pluginName, L) pair) so the bridge can manage VM lifecycle
- Add Bridge() *HTTPBridge getter
- Wrap Shutdown() body in m.shutdownOnce.Do() for idempotency

loadPlugin() execution order after pool creation (existing line references from manager.go):

    1. pool.Get(ctx)                        -- checkout VM (already ran factory -> init.lua, http.handle() succeeded at module scope)
    2. Reset dbAPI op count                  -- existing (line ~305)
    3. Set phase flag: L registry "in_init" = true  -- NEW: blocks http.handle() inside on_init
    4. Call on_init() if defined             -- existing (line ~308)
    5. Clear phase flag: L registry "in_init" = false  -- NEW: re-allows http.handle() (defensive)
    6. inst.Pool.SnapshotGlobals(L)          -- existing (line 313)
    7. bridge.RegisterRoutes(pluginName, L)  -- NEW: reads __http_handlers + __http_route_meta, upserts DB, mounts approved routes
    8. inst.Pool.Put(L)                      -- existing (line 316)
    9. inst.State = StateRunning             -- existing (line 319)

Key timing: http.handle() works at module scope (step 1, during factory's L.DoFile) because the phase flag is NOT set. It fails inside on_init (step 4) because the phase flag IS set (step 3). RegisterRoutes (step 7) runs after the phase flag is cleared (step 5) and after SnapshotGlobals (step 6), while the VM is still checked out.

5. internal/plugin/pool.go -- S

Changes:
- Add GetDBAPI(L *lua.LState) *DatabaseAPI method on PluginInstance (or the Manager) so the bridge can reset op count after checkout
- No structural changes to the pool itself

Note on acquire timeout: The pool's 100ms acquireTimeout is appropriate for fast-fail semantics. With a default pool size of 4 and ExecTimeoutSec of 5s, the pool can sustain ~0.8 requests/second per handler. Plugin authors must size their pool for expected concurrency. The pool size is already configurable via Plugin_Max_VMs in config. A future enhancement could make acquireTimeout separately configurable for HTTP traffic vs lifecycle operations, but the current 100ms fast-fail with 503 + Retry-After is operationally correct for HTTP.

6. internal/plugin/sandbox.go -- S

Changes:
- Add http module to validateVM() health check (validate http.handle and http.use are Go-bound functions)
- Add FreezeModule(L, "http") to factory call sequence (after RegisterHTTPAPI)

7. internal/router/mux.go -- M

Changes:
- Add optional *plugin.HTTPBridge parameter to NewModulacmsMux()
- If bridge is non-nil, call bridge.MountOn(mux) to register approved plugin routes
- Add admin plugin route management endpoints (wrapped in AuthenticatedChain + admin role check):
  - GET /api/v1/admin/plugins/routes
  - POST /api/v1/admin/plugins/routes/approve
  - POST /api/v1/admin/plugins/routes/revoke
- Plugin routes are registered as individual ServeMux patterns BEFORE the "/" slug handler so they take priority

Admin authorization: The route management handlers extract the authenticated user from context (via middleware.AuthenticatedUser(r.Context())) and verify the user has an admin role. AuthenticatedChain ensures the user is logged in; the handler itself checks the role. This is necessary because AuthenticatedChain (and HTTPAuthorizationMiddleware) only check that a user is authenticated — they do not check roles. Route approval is a security-sensitive operation (makes code reachable from the internet) and must require admin privileges.

Note: existing mux.go uses inline middleware wrapping (corsMiddleware(authLimiter.Middleware(...))). The admin plugin routes use AuthenticatedChain from http_chain.go instead — this is intentional as the cleaner pattern:

func NewModulacmsMux(c config.Config, bridge *plugin.HTTPBridge) *http.ServeMux {
    // ... existing routes ...
    if bridge != nil {
        bridge.MountOn(mux)
        // Admin route management endpoints (AuthenticatedChain + admin role check in handler)
        authChain := middleware.AuthenticatedChain(&c)
        mux.Handle("GET /api/v1/admin/plugins/routes", authChain(pluginRoutesListHandler(bridge)))
        mux.Handle("POST /api/v1/admin/plugins/routes/approve", authChain(adminOnly(pluginRoutesApproveHandler(bridge))))
        mux.Handle("POST /api/v1/admin/plugins/routes/revoke", authChain(adminOnly(pluginRoutesRevokeHandler(bridge))))
    }
    // ... slug handler ...
}

adminOnly(next http.Handler) http.Handler is a small wrapper that extracts the user via middleware.AuthenticatedUser(r.Context()), checks user.Role == "admin" (db.Users.Role field, string type, defined in internal/db/user.go:26), and returns 403 Forbidden if the check fails. The list endpoint does NOT require admin (allows any authenticated user to see route status for debugging) but approve/revoke do. The /admin/ URL prefix groups these endpoints under the admin namespace for URL consistency, not as an access control mechanism.

The bridge's MountOn() registers real ServeMux patterns for approved routes plus a fallback "/api/v1/plugins/" for uniform 404 responses.

The Go-level DefaultMiddlewareChain wraps the entire mux in serve.go, so all plugin routes inherit request ID, logging, CORS, and session context from the existing chain. The HTTPPublicEndpointMiddleware exempts all `/api/v1/plugins/` paths (see middleware changes below), allowing the bridge to handle its own per-route auth enforcement. The auth middleware (step 4 in DefaultMiddlewareChain) still runs before the public endpoint gate (step 5), so session context is populated in r.Context() when a valid session exists — the bridge reads it via middleware.AuthenticatedUser(r.Context()) for authenticated routes.

8. cmd/serve.go -- M

Changes at line 151:
// Before: mux := router.NewModulacmsMux(*cfg)
// After:
var bridge *plugin.HTTPBridge
if pluginManager != nil {
    bridge = pluginManager.Bridge()
}
mux := router.NewModulacmsMux(*cfg, bridge)

Shutdown sequence update:

The existing code has `defer pluginManager.Shutdown(rootCtx)` at line 101 of serve.go which would cause a double-shutdown deadlock (Manager.Shutdown acquires m.mu.Lock, called twice = deadlock). Fix: remove the defer and make shutdown explicit in the post-signal block only.

In cmd/serve.go, in the RunE closure:

1. REMOVE line 101: `defer pluginManager.Shutdown(rootCtx)`

2. After the existing sshServer.Shutdown() call (line 240) and BEFORE
   the "Servers gracefully stopped" log (line 245), insert:

    if bridge != nil {
        bridge.Close(shutdownCtx) // drain in-flight plugin HTTP requests
    }
    if pluginManager != nil {
        pluginManager.Shutdown(shutdownCtx) // run on_shutdown, close pools
    }

The full shutdown order in the post-<-done block (line 223 onward) is:
  1. httpServer.Shutdown(shutdownCtx)   -- line 232: drain HTTP connections
  2. httpsServer.Shutdown(shutdownCtx)  -- line 236: drain HTTPS connections
  3. sshServer.Shutdown(shutdownCtx)    -- line 240: drain SSH connections
  4. bridge.Close(shutdownCtx)          -- NEW: drain in-flight plugin HTTP, set closing flag
  5. pluginManager.Shutdown(shutdownCtx) -- NEW: run on_shutdown hooks, close VM pools
  6. "Servers gracefully stopped" log   -- line 245

The bridge drains AFTER HTTP servers stop accepting new connections (step 1-2) but BEFORE the manager closes pools (step 5). This ensures no new plugin HTTP requests arrive while the bridge is draining, and pools remain open until all in-flight requests complete.

Additionally, make Manager.Shutdown() idempotent as defense-in-depth: add a `shutdownOnce sync.Once` field to Manager. Wrap the shutdown body in shutdownOnce.Do(). This prevents deadlock even if Shutdown is accidentally called twice due to future code changes.

9. internal/plugin/manager.go factory function -- S

Update the VM factory inside loadPlugin to include HTTP API registration:
// Updated factory call sequence:
// 1. ApplySandbox
// 2. RegisterPluginRequire
// 3. RegisterDBAPI
// 4. RegisterLogAPI
// 5. RegisterHTTPAPI   <- NEW
// 6. FreezeModule("db")
// 7. FreezeModule("log")
// 8. FreezeModule("http")  <- NEW

10. internal/config/config.go -- S

Add new Plugin_* config fields:

    Plugin_Max_Request_Body  int64    `json:"plugin_max_request_body"`   // bytes, default 1MB
    Plugin_Max_Response_Body int64    `json:"plugin_max_response_body"`  // bytes, default 5MB
    Plugin_Rate_Limit        int      `json:"plugin_rate_limit"`         // req/sec per IP, default 100
    Plugin_Max_Routes        int      `json:"plugin_max_routes"`         // per plugin, default 50
    Plugin_Trusted_Proxies   []string `json:"plugin_trusted_proxies"`    // CIDR list, empty = use RemoteAddr only

All new Plugin_* config fields follow the existing zero-value-means-default pattern (same as Plugin_Max_VMs / Plugin_Timeout). NewHTTPBridge applies defaults: if cfg.Plugin_Max_Request_Body == 0, use MaxPluginRequestBody; if cfg.Plugin_Max_Response_Body == 0, use MaxPluginResponseBody; if cfg.Plugin_Rate_Limit == 0, use DefaultPluginRateLimit; if cfg.Plugin_Max_Routes == 0, use MaxRoutesPerPlugin.

11. internal/middleware/http_middleware.go -- S

Changes:

Add a plugin route prefix exemption to HTTPPublicEndpointMiddleware (line 93). Before the existing PublicEndpoints iteration, add:

    // Plugin routes handle their own auth — the bridge enforces per-route
    // approval checks and authenticated/public route distinctions.
    if strings.HasPrefix(r.URL.Path, "/api/v1/plugins/") {
        next.ServeHTTP(w, r)
        return
    }

This allows all requests to `/api/v1/plugins/*` to pass through to the bridge regardless of CMS session status. The bridge's ServeHTTP already enforces per-route auth: it checks route.Approved (404 if not), checks route.Public (skip auth if true), and returns 401 UNAUTHORIZED for authenticated routes without a CMS session.

This approach is correct because:
1. The auth middleware (step 4 in DefaultMiddlewareChain) runs before the public endpoint gate (step 5), so session context is populated in r.Context() when a valid session exists.
2. Unauthenticated requests to nonexistent plugin paths get a JSON 404 from the bridge's fallback handler, matching the plan's uniform 404 goal for enumeration prevention.
3. No PublicEndpoints synchronization is needed — no atomic.Value, no copy-on-write, no add-on-approve / remove-on-revoke. The bridge handles all auth decisions independently.

The existing `var PublicEndpoints` and the static endpoint list are NOT modified. They continue to work as before for core CMS endpoints.

12. internal/middleware/ratelimit.go -- S

Changes:
- Implement actual cleanup in cleanupLimiters(). The existing function is a no-op with a TODO. With public-facing plugin endpoints, distinct IPs accumulate without bound.
- Add a lastSeen time.Time field to the per-IP limiter entry. Update it on each Allow() call.
- cleanupLimiters runs on a ticker (every 5 minutes). Entries with lastSeen older than 10 minutes are deleted.
- Add a `done chan struct{}` field and a `Close()` method that signals the cleanup goroutine to exit. The bridge calls rateLimiter.Close() during bridge.Close(). This prevents goroutine leaks in tests that create/destroy bridges.
- This is a pre-existing bug but is now critical because public plugin endpoints expose the rate limiter to internet-scale IP cardinality.

Route Deduplication Strategy

Since every VM in the pool runs init.lua (which calls http.handle()), route metadata would be registered N times. Solution:

1. The RegisterHTTPAPI Go-bound function stores handlers in the per-VM __http_handlers table (this is correct -- every VM needs its own handler reference)
2. Inside loadPlugin(), between SnapshotGlobals and pool.Put, the Manager calls bridge.RegisterRoutes(pluginName, L) exactly ONCE using the already checked-out VM
3. RegisterRoutes reads __http_handlers and __http_route_meta keys, upserts route metadata into the plugin_routes DB table, reads approval status back, and mounts approved routes as ServeMux patterns
4. At request time, the bridge knows which plugin owns which route (Go map lookup), checks approval status, checks out any VM from that plugin's pool, and fetches the handler function from that VM's __http_handlers

Cross-plugin duplicate detection: RegisterRoutes checks the bridge's route map before inserting. If a route's full path collides with an existing registration from another plugin, the later plugin is set to StateFailed with a logged error.

Route lifecycle on plugin update:
- Version unchanged, same routes: approval status preserved from DB.
- Version changed: all routes for this plugin are deleted from plugin_routes and re-inserted as unapproved. Admin must re-approve after a version bump.
- Route added: new row inserted as unapproved.
- Route removed: stale row remains in DB but is not mounted. Cleaned up by admin via API or on next restart by CleanupOrphanedRoutes.
- Route public flag changed (same version): approval revoked for that specific route, must be re-approved.

Middleware Chain

Plugin middleware registered via http.use() runs in registration order BEFORE the matched handler. Middleware and handler share the same Lua request table -- mutations propagate by reference.

The execution flow per request:

1. Bridge checks route approval and auth
2. Checks out VM
3. Reads __http_middleware table (Lua array)
4. Iterates: call each middleware with request table
   - If middleware returns non-nil table -> treat as early response, skip handler
   - If middleware returns nil -> continue to next middleware (req mutations persist)
5. Call the matched handler (sees any req fields added by middleware)
6. Write response

Test Fixtures

internal/plugin/testdata/plugins/
  http_plugin/init.lua              -- GET/POST /tasks (authenticated), middleware
  http_public_plugin/init.lua       -- POST /webhook (public), GET /status (public)
  http_params_plugin/init.lua       -- GET /items/{id} path params
  http_error_plugin/init.lua        -- handler that raises error()
  http_timeout_plugin/init.lua      -- handler with infinite loop
  http_middleware_plugin/init.lua   -- middleware that enriches req, handler reads enriched field
  http_blocked_headers/init.lua     -- handler that tries to set Set-Cookie and CORS headers

Testing Strategy

| Test File                | Coverage |
|--------------------------|----------|
| http_bridge_test.go      | Route registration, dispatch to correct plugin, uniform 404 for unknown/unapproved/revoked routes, 503 for non-running plugins, 503+Retry-After for pool exhaustion, concurrent requests, graceful shutdown drain, body size rejection, response size rejection, rate limiting (with trusted proxy config), auth enforcement for authenticated routes, public route bypass (no CMS session required), admin approval/revoke lifecycle, header blocking, security headers present, audit logging, version change revokes all approvals, orphan cleanup |
| http_api_test.go         | http.handle() stores handler in __http_handlers, options table parsing (public flag), http.use() stores middleware, duplicate route detection (error()), invalid method rejection (error()), invalid path rejection (error() for "..", "?", "#", >256 chars, missing leading "/"), route count limit (error()), phase guard (error if called inside on_init) |
| http_request_test.go     | Request conversion (headers lowercase, query, JSON body only when Content-Type matches, path params via PathValue, ClientIP not RemoteAddr), response writing (status, headers, JSON, raw body), response header filtering (blocked headers skipped + warned), error response formatting (PluginErrorResponse schema with request_id), body too large -> 400, response too large -> 500 |
| http_integration_test.go | End-to-end: load plugin with HTTP routes, httptest.NewRecorder, full CRUD cycle via plugin endpoints, middleware rejection, path parameters, middleware req mutation propagation, authenticated vs public route behavior, unapproved route returns 404 |

Pool exhaustion test pattern (from existing pool_test.go):
- Pool size = 1
- Check out the only VM (don't return it)
- Send HTTP request via httptest
- Assert rec.Code == 503
- Assert Retry-After header present
- Do NOT test timing -- the pool test already covers the 100ms acquisition behavior

Middleware mutation test:
- Plugin middleware sets req.custom_field = "enriched"
- Handler reads req.custom_field and includes it in JSON response
- Test asserts response contains the enriched field

Auth enforcement test:
- Authenticated route without CMS session -> 401 UNAUTHORIZED
- Authenticated route with CMS session -> 200 (handler executes)
- Public route without CMS session -> 200 (handler executes)
- Public route with CMS session -> 200 (handler executes, session info available but not required)

Approval test:
- Unapproved route -> 404 ROUTE_NOT_FOUND (not 403)
- Approve route via bridge.ApproveRoute() -> subsequent request returns 200
- Revoke route via bridge.RevokeRoute() -> subsequent request returns 404
- Approve public route -> public route reachable without CMS session
- Revoke public route -> returns 404
- Plugin version change -> all routes revoked, return 404 until re-approved
- Plugin removed -> orphaned plugin_routes rows cleaned up on next LoadAll

Admin API authorization test:
- Non-admin authenticated user calls approve -> 403 Forbidden
- Admin authenticated user calls approve -> 200 OK
- Unauthenticated user calls approve -> 401 Unauthorized
- Non-admin can call list (GET) -> 200 OK

Path validation test:
- http.handle("GET", "../escape", handler) -> error()
- http.handle("GET", "/path?query", handler) -> error()
- http.handle("GET", "", handler) -> error()
- http.handle("GET", "/a" * 257, handler) -> error()
- http.handle("GET", "/valid/{id}", handler) -> succeeds

Header blocking test:
- Handler returns {headers = {["set-cookie"] = "evil=1", ["x-custom"] = "ok"}}
- Assert Set-Cookie NOT in response headers
- Assert X-Custom IS in response headers
- Assert X-Content-Type-Options: nosniff always present

Multi-Agent Work Decomposition

Parallel group 1 (no dependencies):
- Agent A: http_api.go + http_api_test.go -- Lua module registration, handle() with options table, use(), method validation, path validation, route count limit, duplicate detection, phase guard
  Exports: RegisterHTTPAPI(L *lua.LState, pluginName string)
- Agent B: http_request.go + http_request_test.go -- Request/response conversion, PluginErrorResponse schema, writePluginError helper, Content-Type enforcement, response size limit, header blocking (including Cache-Control), security headers, extractClientIP with trusted proxy support
  Exports: BuildLuaRequest(L *lua.LState, r *http.Request, params map[string]string) *lua.LTable, WriteLuaResponse(w http.ResponseWriter, L *lua.LState, responseTbl *lua.LTable, maxRespSize int64) error, writePluginError(w http.ResponseWriter, status int, code string, message string, requestID string), extractClientIP(r *http.Request, trustedProxies []*net.IPNet) string, PluginErrorResponse, PluginErrorDetail, LuaRequest, LuaResponse, blockedResponseHeaders

Parallel group 2 (depends on group 1):
- Agent C: http_bridge.go + http_bridge_test.go -- Bridge coordinator, per-route ServeMux registration, dispatch, auth enforcement (bridge handles all auth decisions for plugin routes), approval checking (with version-aware re-approval), rate limiting, body size limit, scoped context/timeout, graceful shutdown, audit logging, plugin_routes DB table, admin approval/revoke methods, orphan cleanup

Sequential (depends on all):
- Agent D: Modify manager.go, pool.go, sandbox.go -- Wire HTTP API into factory, add health checks, phase flag for on_init guard, RegisterRoutes insertion point, CreatePluginRoutesTable, CleanupOrphanedRoutes, idempotent Shutdown (sync.Once)
- Agent E: Modify config.go, mux.go, serve.go, http_middleware.go, ratelimit.go -- New config fields (including Plugin_Trusted_Proxies), admin route management endpoints with adminOnly wrapper, wire bridge into router and startup, explicit shutdown sequence (remove defer), plugin prefix exemption in HTTPPublicEndpointMiddleware, rate limiter cleanup goroutine
- Agent F: Integration tests + test fixtures (including auth, approval, public/authenticated routes, middleware mutation, header blocking, path validation, admin role enforcement, version change re-approval, orphan cleanup)

Verification

# Unit tests for new files
go test -v ./internal/plugin/ -run TestHTTP -count=1

# Full plugin package
go test -v ./internal/plugin/ -count=1

# Middleware changes
go test -v ./internal/middleware/ -count=1

# Router changes
go test -v ./internal/router/ -count=1

# Ensure no regressions
go test ./... -count=1

# Manual smoke test
# 1. Create plugins/http_demo/init.lua with authenticated + public routes
# 2. Start server: ./modulacms-x86 serve
# 3. curl http://localhost:8080/api/v1/plugins/http_demo/tasks -> 404 (unapproved)
# 4. curl -X POST http://localhost:8080/api/v1/admin/plugins/routes/approve (with admin session)
# 5. curl http://localhost:8080/api/v1/plugins/http_demo/tasks (with session) -> 200
# 6. curl http://localhost:8080/api/v1/plugins/http_demo/tasks (no session) -> 401
# 7. curl http://localhost:8080/api/v1/plugins/http_demo/webhook (no session) -> 200 (public)
# 8. curl with >1MB body -> 400 INVALID_REQUEST
# 9. curl unknown route -> 404 ROUTE_NOT_FOUND with JSON schema
# 10. Verify X-Content-Type-Options: nosniff on all plugin responses
