package plugin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/time/rate"
)

// Constants for plugin HTTP request/response limits and route prefixing.
const (
	MaxPluginRequestBody  = 1 << 20 // 1 MB default for request bodies
	MaxPluginResponseBody = 5 << 20 // 5 MB default for response bodies
	DefaultPluginRateLimit = 100    // requests per second per IP
	PluginRoutePrefix     = "/api/v1/plugins/"
)

// RouteRegistration holds the metadata for a single plugin-registered HTTP route.
// The bridge maintains one RouteRegistration per method+path combination across
// all plugins. Approval status controls whether the route is dispatched.
type RouteRegistration struct {
	Method     string // GET, POST, PUT, DELETE, PATCH
	Path       string // plugin-relative, e.g., "/tasks" or "/tasks/{id}"
	PluginName string
	FullPath   string // computed: /api/v1/plugins/<plugin>/path
	Public     bool   // true = bypass CMS auth, false = require CMS session
	Approved   bool   // only approved routes are dispatched; unapproved return 404
}

// ipLimiterEntry tracks a per-IP rate limiter with last-seen timestamp for cleanup.
type ipLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// HTTPBridge is the central coordinator between the plugin Manager and the HTTP
// ServeMux. It implements http.Handler and dispatches incoming requests to the
// correct plugin handler via VM checkout from the plugin's pool.
//
// The bridge owns the plugin_routes DB table, manages per-route approval status,
// and enforces rate limiting, body size limits, auth checks, and graceful shutdown.
type HTTPBridge struct {
	manager     *Manager
	routes      map[string]*RouteRegistration // "METHOD /api/v1/plugins/plugin/path" -> reg
	mu          sync.RWMutex
	closing     atomic.Bool    // set during shutdown, new requests get 503
	inflight    sync.WaitGroup // tracks in-progress requests for graceful drain
	maxBodySize int64
	maxRespSize int64
	mux         *http.ServeMux // reference for live route registration on approval
	trustedNets []*net.IPNet   // parsed from config at init
	dbPool      *sql.DB        // plugin DB pool for plugin_routes table queries
	dialect     db.Dialect
	execTimeout time.Duration

	// Per-IP rate limiting implemented directly in the bridge. We cannot use
	// middleware.RateLimiter's unexported getLimiter method, so we manage our
	// own map of per-IP token bucket limiters with periodic cleanup.
	rateMu       sync.Mutex
	rateLimiters map[string]*ipLimiterEntry
	rateLimit    rate.Limit
	rateBurst    int
	cleanupDone  chan struct{} // closed to stop the cleanup goroutine

	// Phase 4 A1: registeredPatterns tracks mux patterns already registered
	// via Handle(). Go's http.ServeMux panics on duplicate Handle() calls,
	// so we skip re-registration for patterns we have already mounted. This
	// also fixes a latent panic in ApproveRoute for already-registered patterns
	// and prevents panics during hot reload when the new plugin version
	// re-registers the same route patterns.
	registeredPatterns map[string]bool
}

// NewHTTPBridge creates a new HTTPBridge tied to the given Manager.
//
// Config field wiring (Plugin_Max_Request_Body, Plugin_Max_Response_Body,
// Plugin_Rate_Limit, Plugin_Trusted_Proxies) is handled by Agent E. For now,
// all values use the package-level constants as defaults.
func NewHTTPBridge(manager *Manager, pool *sql.DB, dialect db.Dialect) *HTTPBridge {
	// Compute exec timeout from the manager's config.
	execTimeout := time.Duration(manager.cfg.ExecTimeoutSec) * time.Second
	if execTimeout <= 0 {
		execTimeout = 5 * time.Second
	}

	b := &HTTPBridge{
		manager:            manager,
		routes:             make(map[string]*RouteRegistration),
		maxBodySize:        MaxPluginRequestBody,
		maxRespSize:        MaxPluginResponseBody,
		dbPool:             pool,
		dialect:            dialect,
		execTimeout:        execTimeout,
		rateLimiters:       make(map[string]*ipLimiterEntry),
		rateLimit:          rate.Limit(DefaultPluginRateLimit),
		rateBurst:          DefaultPluginRateLimit,
		cleanupDone:        make(chan struct{}),
		registeredPatterns: make(map[string]bool),
	}

	// Start background cleanup goroutine for stale IP rate limiters.
	go b.cleanupRateLimiters()

	return b
}

// cleanupRateLimiters periodically removes IP rate limiter entries that have not
// been seen for more than 10 minutes. Without this, public-facing plugin endpoints
// would accumulate unbounded per-IP entries.
func (b *HTTPBridge) cleanupRateLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-b.cleanupDone:
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-10 * time.Minute)
			b.rateMu.Lock()
			for ip, entry := range b.rateLimiters {
				if entry.lastSeen.Before(cutoff) {
					delete(b.rateLimiters, ip)
				}
			}
			b.rateMu.Unlock()
		}
	}
}

// allowIP checks whether the given IP address is within the per-IP rate limit.
// Returns true if the request is allowed, false if rate limited.
func (b *HTTPBridge) allowIP(ip string) bool {
	b.rateMu.Lock()
	entry, exists := b.rateLimiters[ip]
	if !exists {
		entry = &ipLimiterEntry{
			limiter:  rate.NewLimiter(b.rateLimit, b.rateBurst),
			lastSeen: time.Now(),
		}
		b.rateLimiters[ip] = entry
	} else {
		entry.lastSeen = time.Now()
	}
	b.rateMu.Unlock()
	return entry.limiter.Allow()
}

// CreatePluginRoutesTable creates the plugin_routes table using raw SQL
// appropriate for the configured dialect. Called by the Manager at startup
// before LoadAll.
func (b *HTTPBridge) CreatePluginRoutesTable(ctx context.Context) error {
	var ddl string

	switch b.dialect {
	case db.DialectSQLite:
		ddl = `CREATE TABLE IF NOT EXISTS plugin_routes (
    plugin_name     TEXT NOT NULL,
    method          TEXT NOT NULL,
    path            TEXT NOT NULL,
    public          INTEGER NOT NULL DEFAULT 0,
    approved        INTEGER NOT NULL DEFAULT 0,
    approved_at     TEXT,
    approved_by     TEXT,
    plugin_version  TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (plugin_name, method, path)
)`
	case db.DialectMySQL:
		ddl = `CREATE TABLE IF NOT EXISTS plugin_routes (
    plugin_name     TEXT NOT NULL,
    method          TEXT NOT NULL,
    path            TEXT NOT NULL,
    public          BOOLEAN NOT NULL DEFAULT FALSE,
    approved        BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at     TEXT,
    approved_by     TEXT,
    plugin_version  TEXT NOT NULL DEFAULT '',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (plugin_name(191), method(10), path(191))
)`
	case db.DialectPostgres:
		ddl = `CREATE TABLE IF NOT EXISTS plugin_routes (
    plugin_name     TEXT NOT NULL,
    method          TEXT NOT NULL,
    path            TEXT NOT NULL,
    public          BOOLEAN NOT NULL DEFAULT FALSE,
    approved        BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at     TEXT,
    approved_by     TEXT,
    plugin_version  TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (plugin_name, method, path)
)`
	default:
		return fmt.Errorf("unsupported dialect for plugin_routes table: %d", b.dialect)
	}

	_, err := b.dbPool.ExecContext(ctx, ddl)
	if err != nil {
		return fmt.Errorf("creating plugin_routes table: %w", err)
	}

	return nil
}

// CleanupOrphanedRoutes deletes rows from plugin_routes where plugin_name is
// not in the discovered set. Called by the Manager after scanning for plugin
// directories but before loading plugins.
func (b *HTTPBridge) CleanupOrphanedRoutes(ctx context.Context, discoveredPlugins []string) error {
	if len(discoveredPlugins) == 0 {
		// No plugins discovered -- delete all rows.
		_, err := b.dbPool.ExecContext(ctx, "DELETE FROM plugin_routes")
		if err != nil {
			return fmt.Errorf("cleaning all orphaned routes: %w", err)
		}
		return nil
	}

	// Build parameterized IN clause. SQLite and MySQL use ?, PostgreSQL uses $N.
	placeholders := make([]string, len(discoveredPlugins))
	args := make([]any, len(discoveredPlugins))
	for i, name := range discoveredPlugins {
		switch b.dialect {
		case db.DialectPostgres:
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		default:
			placeholders[i] = "?"
		}
		args[i] = name
	}

	query := fmt.Sprintf(
		"DELETE FROM plugin_routes WHERE plugin_name NOT IN (%s)",
		strings.Join(placeholders, ", "),
	)

	_, err := b.dbPool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("cleaning orphaned routes: %w", err)
	}

	return nil
}

// registerOnMux safely registers a pattern on the ServeMux, skipping if the
// pattern was already registered (A1). Go's http.ServeMux panics on duplicate
// Handle() calls, so this guard is essential for hot reload and re-approval.
func (b *HTTPBridge) registerOnMux(pattern string) {
	if b.mux == nil {
		return
	}
	if b.registeredPatterns[pattern] {
		return // already registered -- skip to avoid ServeMux panic
	}
	b.mux.Handle(pattern, b)
	b.registeredPatterns[pattern] = true
}

// RegisterRoutes extracts route metadata from a checked-out VM's __http_handlers
// and __http_route_meta tables, upserts into the plugin_routes DB table, and
// mounts approved routes on the ServeMux.
//
// Called once per plugin inside loadPlugin, after SnapshotGlobals and before
// pool.Put. The VM (L) must still be checked out.
//
// Version change handling: if the stored plugin_version differs from the current
// version, all existing routes for the plugin are deleted and re-inserted as
// unapproved. This forces admin re-approval after a plugin update.
func (b *HTTPBridge) RegisterRoutes(ctx context.Context, pluginName string, pluginVersion string, L *lua.LState) error {
	// Read __http_handlers keys to find registered routes.
	handlersVal := L.GetGlobal("__http_handlers")
	handlersTbl, ok := handlersVal.(*lua.LTable)
	if !ok || handlersVal == lua.LNil {
		// No routes registered -- nothing to do.
		return nil
	}

	// Read __http_route_meta for public flags.
	routeMetaVal := L.GetGlobal("__http_route_meta")
	routeMetaTbl, _ := routeMetaVal.(*lua.LTable)

	// Collect route registrations from Lua state.
	type pendingRoute struct {
		method string
		path   string
		public bool
	}
	var pending []pendingRoute

	handlersTbl.ForEach(func(key, _ lua.LValue) {
		keyStr, isStr := key.(lua.LString)
		if !isStr {
			return
		}
		parts := strings.SplitN(string(keyStr), " ", 2)
		if len(parts) != 2 {
			return
		}
		method := parts[0]
		path := parts[1]

		isPublic := false
		if routeMetaTbl != nil {
			meta := L.GetField(routeMetaTbl, string(keyStr))
			if metaTbl, ok := meta.(*lua.LTable); ok {
				publicVal := L.GetField(metaTbl, "public")
				if pb, ok := publicVal.(lua.LBool); ok {
					isPublic = bool(pb)
				}
			}
		}

		pending = append(pending, pendingRoute{method: method, path: path, public: isPublic})
	})

	if len(pending) == 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Check for version change: query existing version from DB.
	versionChanged, versionErr := b.checkVersionChange(ctx, pluginName, pluginVersion)
	if versionErr != nil {
		return fmt.Errorf("checking plugin version: %w", versionErr)
	}

	if versionChanged {
		// Delete all existing routes for this plugin (forces re-approval).
		// A5: Use dialect-aware query for PostgreSQL $N placeholder support.
		var delQuery string
		switch b.dialect {
		case db.DialectPostgres:
			delQuery = "DELETE FROM plugin_routes WHERE plugin_name = $1"
		default:
			delQuery = "DELETE FROM plugin_routes WHERE plugin_name = ?"
		}
		if _, delErr := b.dbPool.ExecContext(ctx, delQuery, pluginName); delErr != nil {
			return fmt.Errorf("deleting routes for version change: %w", delErr)
		}
		utility.DefaultLogger.Info(
			fmt.Sprintf("plugin %q version changed, all routes revoked for re-approval", pluginName),
		)
	}

	for _, r := range pending {
		fullPath := PluginRoutePrefix + pluginName + r.path
		muxPattern := r.method + " " + fullPath

		// Cross-plugin collision check.
		if existing, exists := b.routes[muxPattern]; exists && existing.PluginName != pluginName {
			return fmt.Errorf(
				"route collision: %s %s already registered by plugin %q",
				r.method, fullPath, existing.PluginName,
			)
		}

		// Upsert into plugin_routes DB table.
		upsertErr := b.upsertRoute(ctx, pluginName, pluginVersion, r.method, r.path, r.public)
		if upsertErr != nil {
			return fmt.Errorf("upserting route %s %s: %w", r.method, r.path, upsertErr)
		}

		// Read back approval status from DB.
		approved, publicFlag, readErr := b.readRouteApproval(ctx, pluginName, r.method, r.path)
		if readErr != nil {
			return fmt.Errorf("reading route approval %s %s: %w", r.method, r.path, readErr)
		}

		reg := &RouteRegistration{
			Method:     r.method,
			Path:       r.path,
			PluginName: pluginName,
			FullPath:   fullPath,
			Public:     publicFlag,
			Approved:   approved,
		}

		b.routes[muxPattern] = reg

		// If approved, register on ServeMux (A1: safe against duplicate registration).
		if approved {
			b.registerOnMux(muxPattern)
		}
	}

	return nil
}

// checkVersionChange queries the DB for any existing route with a different
// plugin_version for the given plugin. Returns true if a version change is
// detected or if no rows exist yet (first load -- not a version "change" per se,
// but no deletion needed).
func (b *HTTPBridge) checkVersionChange(ctx context.Context, pluginName string, newVersion string) (bool, error) {
	// A5: Dialect-aware query for PostgreSQL $N placeholder support.
	var query string
	switch b.dialect {
	case db.DialectPostgres:
		query = "SELECT plugin_version FROM plugin_routes WHERE plugin_name = $1 LIMIT 1"
	default:
		query = "SELECT plugin_version FROM plugin_routes WHERE plugin_name = ? LIMIT 1"
	}

	var storedVersion string
	err := b.dbPool.QueryRowContext(ctx, query, pluginName).Scan(&storedVersion)

	if errors.Is(err, sql.ErrNoRows) {
		// No existing routes -- first load. No version change to handle.
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("querying stored version: %w", err)
	}

	return storedVersion != newVersion, nil
}

// upsertRoute inserts or updates a route in the plugin_routes table.
// For new routes, approved defaults to false. For existing routes where the
// public flag changed, approval is revoked (plan requirement: "If an existing
// approved route changes its public flag, approval is revoked").
func (b *HTTPBridge) upsertRoute(ctx context.Context, pluginName, pluginVersion, method, path string, public bool) error {
	publicInt := 0
	if public {
		publicInt = 1
	}

	switch b.dialect {
	case db.DialectSQLite:
		// SQLite: Use INSERT OR IGNORE for new rows, then conditionally update.
		// If the public flag changed, revoke approval by resetting approved to 0.
		_, err := b.dbPool.ExecContext(ctx, `
			INSERT INTO plugin_routes (plugin_name, method, path, public, approved, plugin_version)
			VALUES (?, ?, ?, ?, 0, ?)
			ON CONFLICT(plugin_name, method, path) DO UPDATE SET
				plugin_version = excluded.plugin_version,
				public = excluded.public,
				approved = CASE
					WHEN plugin_routes.public != excluded.public THEN 0
					ELSE plugin_routes.approved
				END,
				approved_at = CASE
					WHEN plugin_routes.public != excluded.public THEN NULL
					ELSE plugin_routes.approved_at
				END,
				approved_by = CASE
					WHEN plugin_routes.public != excluded.public THEN NULL
					ELSE plugin_routes.approved_by
				END
		`, pluginName, method, path, publicInt, pluginVersion)
		return err

	case db.DialectMySQL:
		_, err := b.dbPool.ExecContext(ctx, `
			INSERT INTO plugin_routes (plugin_name, method, path, public, approved, plugin_version)
			VALUES (?, ?, ?, ?, 0, ?)
			ON DUPLICATE KEY UPDATE
				plugin_version = VALUES(plugin_version),
				public = VALUES(public),
				approved = CASE
					WHEN plugin_routes.public != VALUES(public) THEN 0
					ELSE plugin_routes.approved
				END,
				approved_at = CASE
					WHEN plugin_routes.public != VALUES(public) THEN NULL
					ELSE plugin_routes.approved_at
				END,
				approved_by = CASE
					WHEN plugin_routes.public != VALUES(public) THEN NULL
					ELSE plugin_routes.approved_by
				END
		`, pluginName, method, path, publicInt, pluginVersion)
		return err

	case db.DialectPostgres:
		_, err := b.dbPool.ExecContext(ctx, `
			INSERT INTO plugin_routes (plugin_name, method, path, public, approved, plugin_version)
			VALUES ($1, $2, $3, $4, FALSE, $5)
			ON CONFLICT (plugin_name, method, path) DO UPDATE SET
				plugin_version = EXCLUDED.plugin_version,
				public = EXCLUDED.public,
				approved = CASE
					WHEN plugin_routes.public != EXCLUDED.public THEN FALSE
					ELSE plugin_routes.approved
				END,
				approved_at = CASE
					WHEN plugin_routes.public != EXCLUDED.public THEN NULL
					ELSE plugin_routes.approved_at
				END,
				approved_by = CASE
					WHEN plugin_routes.public != EXCLUDED.public THEN NULL
					ELSE plugin_routes.approved_by
				END
		`, pluginName, method, path, publicInt, pluginVersion)
		return err

	default:
		return fmt.Errorf("unsupported dialect: %d", b.dialect)
	}
}

// readRouteApproval reads the approved and public flags from the plugin_routes table.
func (b *HTTPBridge) readRouteApproval(ctx context.Context, pluginName, method, path string) (approved bool, public bool, err error) {
	// A5: Dialect-aware query for PostgreSQL $N placeholder support.
	var query string
	switch b.dialect {
	case db.DialectPostgres:
		query = "SELECT CASE WHEN approved THEN 1 ELSE 0 END, CASE WHEN public THEN 1 ELSE 0 END FROM plugin_routes WHERE plugin_name = $1 AND method = $2 AND path = $3"
	default:
		query = "SELECT approved, public FROM plugin_routes WHERE plugin_name = ? AND method = ? AND path = ?"
	}

	var approvedInt, publicInt int
	err = b.dbPool.QueryRowContext(ctx, query, pluginName, method, path).Scan(&approvedInt, &publicInt)
	if err != nil {
		return false, false, err
	}
	return approvedInt != 0, publicInt != 0, nil
}

// MountOn stores the ServeMux reference and registers all approved routes as
// real ServeMux patterns. Also registers a fallback handler at the plugin route
// prefix for uniform 404 responses on unmatched paths.
//
// Each approved route is registered as its own pattern (e.g.,
// "GET /api/v1/plugins/task_tracker/tasks") so that Go 1.22+ ServeMux handles
// path parameter matching and precedence.
func (b *HTTPBridge) MountOn(mux *http.ServeMux) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.mux = mux

	// Register all approved routes (A1: safe against duplicate registration).
	for pattern, reg := range b.routes {
		if reg.Approved {
			b.registerOnMux(pattern)
		}
	}

	// Register fallback for uniform 404 on all unmatched /api/v1/plugins/ requests.
	b.registerOnMux(PluginRoutePrefix)
}

// ServeHTTP is the main dispatch handler for plugin HTTP requests. It follows
// the Request Lifecycle defined in PLUGIN_PHASE_2.md:
//
//  1. Set security headers
//  2. Look up route via r.Pattern (O(1))
//  3. Check approval status
//  4. Check auth (if not public)
//  5. Rate limit
//  6. Track inflight / check closing
//  7. Apply body size limit
//  8. Create scoped context with timeout
//  9. Checkout VM from pool
//  10. Reset DB op count
//  11. Build Lua request
//  12. Run middleware chain
//  13. Call handler
//  14. Write response
func (b *HTTPBridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Step 1: Set default security headers. These are always present and cannot
	// be overridden by plugins (Cache-Control is in blockedResponseHeaders).
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Cache-Control", "no-store")

	// Step 2: Get request ID from context.
	requestID := RequestIDFromContext(r)

	// Step 3: Look up route by r.Pattern (set by ServeMux to the matched pattern).
	b.mu.RLock()
	route, ok := b.routes[r.Pattern]
	b.mu.RUnlock()

	if !ok {
		// Uniform 404 for all unmatched paths -- prevents plugin enumeration.
		writePluginError(w, http.StatusNotFound, "ROUTE_NOT_FOUND",
			"the requested plugin endpoint does not exist", requestID)
		return
	}

	// Step 4: Check approval status. Unapproved routes get 404 (not 403)
	// to prevent enumeration of registered-but-unapproved routes.
	if !route.Approved {
		writePluginError(w, http.StatusNotFound, "ROUTE_NOT_FOUND",
			"the requested plugin endpoint does not exist", requestID)
		return
	}

	// Step 5: Check auth. For authenticated routes (not public), require a
	// CMS session via middleware.AuthenticatedUser.
	if !route.Public {
		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			writePluginError(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"authentication required", requestID)
			return
		}
	}

	// Step 6: Rate limit check.
	clientIP := extractClientIP(r, b.trustedNets)
	if !b.allowIP(clientIP) {
		w.Header().Set("Retry-After", "1")
		writePluginError(w, http.StatusTooManyRequests, "RATE_LIMITED",
			"rate limit exceeded", requestID)
		return
	}

	// Step 7: Track inflight requests for graceful shutdown.
	b.inflight.Add(1)
	defer b.inflight.Done()

	// Step 8: Check closing flag. New requests after Close() get 503.
	if b.closing.Load() {
		writePluginError(w, http.StatusServiceUnavailable, "PLUGIN_UNAVAILABLE",
			"plugin system is shutting down", requestID)
		return
	}

	// Step 9: Apply body size limit.
	r.Body = http.MaxBytesReader(w, r.Body, b.maxBodySize)

	// Step 10: Create scoped context with execution timeout.
	execCtx, cancel := context.WithTimeout(r.Context(), b.execTimeout)
	defer cancel()

	// Step 11: Get plugin instance and check state.
	inst := b.manager.GetPlugin(route.PluginName)
	if inst == nil || inst.State != StateRunning {
		writePluginError(w, http.StatusServiceUnavailable, "PLUGIN_UNAVAILABLE",
			"plugin is not running", requestID)
		return
	}

	// Phase 4: Check circuit breaker before VM checkout.
	if inst.CB != nil && !inst.CB.Allow() {
		writePluginError(w, http.StatusServiceUnavailable, "CIRCUIT_BREAKER_OPEN",
			"plugin is temporarily disabled due to repeated failures", requestID)
		return
	}

	// Phase 4: Start timing for metrics.
	httpStart := time.Now()

	// Step 12: Checkout VM from pool.
	L, getErr := inst.Pool.Get(execCtx)
	if getErr != nil {
		if errors.Is(getErr, ErrPoolExhausted) {
			w.Header().Set("Retry-After", "1")
			writePluginError(w, http.StatusServiceUnavailable, "POOL_EXHAUSTED",
				"all plugin VMs are busy, try again shortly", requestID)
			return
		}
		writePluginError(w, http.StatusServiceUnavailable, "PLUGIN_UNAVAILABLE",
			"failed to acquire plugin VM", requestID)
		return
	}
	defer inst.Pool.Put(L)

	// Step 13: Reset DB API op count for this checkout (S5: protected by inst.mu).
	inst.mu.Lock()
	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}
	inst.mu.Unlock()

	// Step 14: Set execution context on VM for timeout enforcement.
	L.SetContext(execCtx)

	// Step 15: Build Lua request table from HTTP request.
	luaReq, buildErr := BuildLuaRequest(L, r, clientIP)
	if buildErr != nil {
		// Check for MaxBytesError (body too large).
		var maxBytesErr *http.MaxBytesError
		if errors.As(buildErr, &maxBytesErr) {
			writePluginError(w, http.StatusBadRequest, "INVALID_REQUEST",
				"request body too large", requestID)
			return
		}
		writePluginError(w, http.StatusBadRequest, "INVALID_REQUEST",
			"failed to read request", requestID)
		return
	}

	// Step 16: Run middleware chain from __http_middleware.
	middlewareTbl := L.GetGlobal("__http_middleware")
	if mwTbl, ok := middlewareTbl.(*lua.LTable); ok && middlewareTbl != lua.LNil {
		earlyResponse, mwErr := b.runMiddlewareChain(L, mwTbl, luaReq, requestID)
		if mwErr != nil {
			// Check both the error chain and the context directly. gopher-lua
			// wraps context.DeadlineExceeded inside *lua.ApiError which does not
			// implement Unwrap(), so errors.Is may not reach the sentinel. Checking
			// execCtx.Err() is the reliable fallback.
			if errors.Is(mwErr, context.DeadlineExceeded) || errors.Is(execCtx.Err(), context.DeadlineExceeded) {
				writePluginError(w, http.StatusGatewayTimeout, "HANDLER_TIMEOUT",
					"plugin execution timed out", requestID)
				return
			}
			utility.DefaultLogger.Error(
				fmt.Sprintf("plugin %q middleware error [request_id=%s]: %s",
					route.PluginName, requestID, mwErr.Error()), nil)
			writePluginError(w, http.StatusInternalServerError, "HANDLER_ERROR",
				"internal plugin error", requestID)
			return
		}
		if earlyResponse != nil {
			// Middleware returned an early response -- skip handler.
			respErr := WriteLuaResponse(w, L, earlyResponse, b.maxRespSize, requestID)
			if respErr != nil {
				utility.DefaultLogger.Error(
					fmt.Sprintf("plugin %q middleware response write error [request_id=%s]: %s",
						route.PluginName, requestID, respErr.Error()), nil)
			}
			return
		}
	}

	// Step 17: Look up handler from this VM's __http_handlers table.
	// The handler key is the plugin-relative "METHOD /path" (e.g., "GET /tasks"),
	// which matches what http.handle() stored.
	handlerKey := route.Method + " " + route.Path
	handlersTbl := L.GetGlobal("__http_handlers")
	if handlersTbl == lua.LNil {
		writePluginError(w, http.StatusInternalServerError, "HANDLER_ERROR",
			"internal plugin error", requestID)
		return
	}
	handlerFn := L.GetField(handlersTbl, handlerKey)
	luaFn, ok := handlerFn.(*lua.LFunction)
	if !ok {
		utility.DefaultLogger.Error(
			fmt.Sprintf("plugin %q handler not found for key %q [request_id=%s]",
				route.PluginName, handlerKey, requestID), nil)
		writePluginError(w, http.StatusInternalServerError, "HANDLER_ERROR",
			"internal plugin error", requestID)
		return
	}

	// Step 18: Call handler with the Lua request table.
	callErr := L.CallByParam(lua.P{
		Fn:      luaFn,
		NRet:    1,
		Protect: true,
	}, luaReq)

	if callErr != nil {
		// Phase 4: Record failure on circuit breaker.
		if inst.CB != nil {
			inst.CB.RecordFailure()
		}
		httpDuration := float64(time.Since(httpStart).Milliseconds())

		// Check both the error chain and the context directly. gopher-lua
		// wraps context.DeadlineExceeded inside *lua.ApiError which does not
		// implement Unwrap(), so errors.Is may not reach the sentinel. Checking
		// execCtx.Err() is the reliable fallback.
		if errors.Is(callErr, context.DeadlineExceeded) || errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			RecordHTTPRequest(route.PluginName, r.Method, http.StatusGatewayTimeout, httpDuration)
			writePluginError(w, http.StatusGatewayTimeout, "HANDLER_TIMEOUT",
				"plugin execution timed out", requestID)
			return
		}
		RecordHTTPRequest(route.PluginName, r.Method, http.StatusInternalServerError, httpDuration)
		utility.DefaultLogger.Error(
			fmt.Sprintf("plugin %q handler error [request_id=%s]: %s",
				route.PluginName, requestID, callErr.Error()), nil)
		writePluginError(w, http.StatusInternalServerError, "HANDLER_ERROR",
			"internal plugin error", requestID)
		return
	}

	// Step 19: Read response from Lua stack.
	respVal := L.Get(-1)
	L.Pop(1)

	respTbl, ok := respVal.(*lua.LTable)
	if !ok {
		utility.DefaultLogger.Error(
			fmt.Sprintf("plugin %q handler returned non-table response [request_id=%s]",
				route.PluginName, requestID), nil)
		writePluginError(w, http.StatusInternalServerError, "HANDLER_ERROR",
			"internal plugin error", requestID)
		return
	}

	// Step 20: Write HTTP response.
	respErr := WriteLuaResponse(w, L, respTbl, b.maxRespSize, requestID)
	if respErr != nil {
		utility.DefaultLogger.Error(
			fmt.Sprintf("plugin %q response write error [request_id=%s]: %s",
				route.PluginName, requestID, respErr.Error()), nil)
	}

	// Phase 4: Record success on circuit breaker and emit metrics.
	if inst.CB != nil {
		inst.CB.RecordSuccess()
	}
	httpDuration := float64(time.Since(httpStart).Milliseconds())
	RecordHTTPRequest(route.PluginName, r.Method, http.StatusOK, httpDuration)
}

// runMiddlewareChain iterates the __http_middleware table and calls each
// middleware function with the request table. If a middleware returns a non-nil
// table, it is treated as an early response and execution stops.
//
// Returns the early response table (or nil) and any error from the Lua call.
func (b *HTTPBridge) runMiddlewareChain(L *lua.LState, mwTbl *lua.LTable, luaReq *lua.LTable, requestID string) (*lua.LTable, error) {
	mwLen := mwTbl.Len()
	for i := 1; i <= mwLen; i++ {
		mwFn := L.RawGetInt(mwTbl, i)
		luaFn, ok := mwFn.(*lua.LFunction)
		if !ok {
			continue
		}

		callErr := L.CallByParam(lua.P{
			Fn:      luaFn,
			NRet:    1,
			Protect: true,
		}, luaReq)

		if callErr != nil {
			return nil, callErr
		}

		retVal := L.Get(-1)
		L.Pop(1)

		// If middleware returns a non-nil table, treat as early response.
		if retVal != lua.LNil {
			if retTbl, ok := retVal.(*lua.LTable); ok {
				return retTbl, nil
			}
		}
	}

	return nil, nil
}

// ApproveRoute approves a plugin route, updating the DB and in-memory state.
// If the ServeMux is set, the route pattern is registered immediately.
func (b *HTTPBridge) ApproveRoute(ctx context.Context, pluginName, method, path, approvedBy string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	fullPath := PluginRoutePrefix + pluginName + path
	muxPattern := method + " " + fullPath

	route, exists := b.routes[muxPattern]
	if !exists {
		return fmt.Errorf("route %s %s not found for plugin %q", method, path, pluginName)
	}

	// A5: Dialect-aware query for PostgreSQL $N placeholder support.
	now := time.Now().UTC().Format(time.RFC3339)
	var approveQuery string
	switch b.dialect {
	case db.DialectPostgres:
		approveQuery = "UPDATE plugin_routes SET approved = TRUE, approved_at = $1, approved_by = $2 WHERE plugin_name = $3 AND method = $4 AND path = $5"
	default:
		approveQuery = "UPDATE plugin_routes SET approved = 1, approved_at = ?, approved_by = ? WHERE plugin_name = ? AND method = ? AND path = ?"
	}
	_, err := b.dbPool.ExecContext(ctx, approveQuery, now, approvedBy, pluginName, method, path)
	if err != nil {
		return fmt.Errorf("approving route in DB: %w", err)
	}

	route.Approved = true

	// Register on ServeMux (A1: safe against duplicate registration).
	b.registerOnMux(muxPattern)

	return nil
}

// RevokeRoute revokes approval for a plugin route, updating the DB and
// in-memory state. The ServeMux pattern stays registered but ServeHTTP
// checks the Approved flag and returns 404 for revoked routes.
func (b *HTTPBridge) RevokeRoute(ctx context.Context, pluginName, method, path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	fullPath := PluginRoutePrefix + pluginName + path
	muxPattern := method + " " + fullPath

	route, exists := b.routes[muxPattern]
	if !exists {
		return fmt.Errorf("route %s %s not found for plugin %q", method, path, pluginName)
	}

	// A5: Dialect-aware query for PostgreSQL $N placeholder support.
	var revokeQuery string
	switch b.dialect {
	case db.DialectPostgres:
		revokeQuery = "UPDATE plugin_routes SET approved = FALSE, approved_at = NULL, approved_by = NULL WHERE plugin_name = $1 AND method = $2 AND path = $3"
	default:
		revokeQuery = "UPDATE plugin_routes SET approved = 0, approved_at = NULL, approved_by = NULL WHERE plugin_name = ? AND method = ? AND path = ?"
	}
	_, err := b.dbPool.ExecContext(ctx, revokeQuery, pluginName, method, path)
	if err != nil {
		return fmt.Errorf("revoking route in DB: %w", err)
	}

	route.Approved = false

	return nil
}

// ListRoutes returns a copy of all registered route registrations.
// Thread-safe via read lock.
func (b *HTTPBridge) ListRoutes() []RouteRegistration {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]RouteRegistration, 0, len(b.routes))
	for _, reg := range b.routes {
		result = append(result, *reg)
	}
	return result
}

// Close performs graceful shutdown of the bridge:
//  1. Sets closing flag so new requests get 503
//  2. Waits for inflight requests to complete (with ctx deadline as backstop)
//  3. Stops the rate limiter cleanup goroutine
func (b *HTTPBridge) Close(ctx context.Context) {
	// Step 1: Signal that we are closing. New requests will get 503.
	b.closing.Store(true)

	// Step 2: Wait for inflight requests with ctx deadline as backstop.
	done := make(chan struct{})
	go func() {
		b.inflight.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All inflight requests completed.
	case <-ctx.Done():
		// Deadline exceeded -- some requests may still be running.
		utility.DefaultLogger.Warn(
			"HTTPBridge shutdown: context expired before all inflight requests drained", nil)
	}

	// Step 3: Stop the cleanup goroutine.
	close(b.cleanupDone)
}

// UnregisterPlugin removes all in-memory route registrations for the given plugin.
// DB rows are kept for approval persistence. The mux patterns are never removed
// (Go's http.ServeMux does not support it) -- ServeHTTP already returns 404 for
// routes not in the in-memory map (A1).
func (b *HTTPBridge) UnregisterPlugin(pluginName string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for key, reg := range b.routes {
		if reg.PluginName == pluginName {
			delete(b.routes, key)
		}
	}
}

// MountAdminEndpoints registers plugin management admin endpoints on the mux.
// Called from router/mux.go after bridge.MountOn(mux). The adminOnly function
// is passed as a parameter since it is defined in mux.go (router-layer concern).
//
// A4: Register literal-path endpoints BEFORE wildcard {name} patterns.
// Go 1.22+ ServeMux gives literal segments precedence, but explicit order
// documents intent.
func (b *HTTPBridge) MountAdminEndpoints(
	mux *http.ServeMux,
	authChain func(http.Handler) http.Handler,
	readPerm func(http.Handler) http.Handler,
	adminPerm func(http.Handler) http.Handler,
) {
	mgr := b.manager

	// 1. Register literal-path endpoints first (these beat {name} wildcard):
	mux.Handle("GET /api/v1/admin/plugins",
		authChain(readPerm(PluginListHandler(mgr))))

	mux.Handle("GET /api/v1/admin/plugins/cleanup",
		authChain(adminPerm(PluginCleanupListHandler(mgr))))
	mux.Handle("POST /api/v1/admin/plugins/cleanup",
		authChain(adminPerm(PluginCleanupDropHandler(mgr))))

	mux.Handle("GET /api/v1/admin/plugins/hooks",
		authChain(readPerm(PluginHooksListHandler(mgr))))
	mux.Handle("POST /api/v1/admin/plugins/hooks/approve",
		authChain(adminPerm(PluginHooksApproveHandler(mgr))))
	mux.Handle("POST /api/v1/admin/plugins/hooks/revoke",
		authChain(adminPerm(PluginHooksRevokeHandler(mgr))))

	// 2. Register wildcard {name} endpoints:
	mux.Handle("GET /api/v1/admin/plugins/{name}",
		authChain(readPerm(PluginInfoHandler(mgr))))
	mux.Handle("POST /api/v1/admin/plugins/{name}/reload",
		authChain(adminPerm(PluginReloadHandler(mgr))))
	mux.Handle("POST /api/v1/admin/plugins/{name}/enable",
		authChain(adminPerm(PluginEnableHandler(mgr))))
	mux.Handle("POST /api/v1/admin/plugins/{name}/disable",
		authChain(adminPerm(PluginDisableHandler(mgr))))
}
