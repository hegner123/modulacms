Phase 4: Plugin Production Hardening -- Implementation Plan
Context

 Phases 1-3 of the ModulaCMS plugin system are fully implemented (~19k lines):
 core engine, HTTP integration, and content hooks. Phase 4 adds operational
 robustness: hot reload, metrics, circuit breaker recovery, and admin management
 endpoints. These are essential for running plugins in production without
 requiring CMS restarts for every plugin change and providing visibility into
 plugin health.

 Architecture Review

 This plan incorporates fixes for 7 critical concerns raised during architecture
 review, plus 11 security hardening fixes (2 Critical, 2 High, 4 Medium,
 3 Low) from security review:

 ┌────────────────────────────────────────────────────────────────────────────┬──────────────────────────────────────────────────────────┬───────────────────────────────────────────┐
 │                                     #                                      │                         Concern                          │                Resolution                 │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ 1                                                                          │ VMPool.Drain() conflicts with existing Close() semantics │ Tri-state pool                            │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ lifecycle: open → draining → closed (new draining atomic.Bool)             │                                                          │                                           │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ 2                                                                          │ Old PluginInstance data race during reload               │ Old instance kept alive until             │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ inflight drains; new instance created independently                        │                                                          │                                           │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ 3                                                                          │ Two circuit breakers with undefined interaction          │ Plugin-level CB tracks                    │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ HTTP/manager failures only; hook aborts stay isolated to hook-level CB     │                                                          │                                           │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ 4                                                                          │ No debounce on file watcher                              │ 1s stability timer: reset on each change, │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ reload only after 1s of quiet                                              │                                                          │                                           │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ 5                                                                          │ DropOrphanedTables is a data loss footgun                │ Exclude StateFailed plugins,              │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ require confirm: true, log every DROP, dry-run default                     │                                                          │                                           │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ 6                                                                          │ Metrics overhead on hot paths unverified                 │ Instrument only coarse-grained            │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ paths (HTTP requests, hook events); skip per-Get/Put/DB-op instrumentation │                                                          │                                           │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ 7                                                                          │ NewModulacmsMux signature change unnecessary             │ Route admin endpoints                     │
 ├────────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ through HTTPBridge (which already holds Manager reference)                 │                                                          │                                           │
 └────────────────────────────────────────────────────────────────────────────┴──────────────────────────────────────────────────────────┴───────────────────────────────────────────┘

 Security Review Fixes

 ┌────┬───────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │ #  │ Severity  │ Fix                                                                                                  │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ S1 │ Critical  │ DropOrphanedTables no longer accepts arbitrary []string. Internally re-lists orphaned tables,         │
 │    │           │ intersects with caller's requested list, validates each name via db.ValidTableName, and               │
 │    │           │ re-verifies orphan status at execution time to close TOCTOU window between GET and POST.              │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ S2 │ Critical  │ Admin endpoints specify explicit authorization: read-only (list, info) = any authenticated;           │
 │    │           │ all mutating (reload, enable, disable) + cleanup (both GET and POST) = adminOnly.                     │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ S3 │ High      │ Watcher uses try-lock on reloadMu (skip + log warning if held). Circuit breaker tracks reload         │
 │    │           │ duration -- 3 consecutive reloads exceeding 10s pauses watcher polling for that plugin.               │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ S4 │ High      │ computeChecksum uses os.Lstat and skips symlinks/non-regular files. Enforces max 100 .lua files       │
 │    │           │ per plugin directory and max 10 MB total bytes per checksumming pass.                                 │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ S5 │ Medium    │ EnablePlugin and CircuitBreaker.Reset emit slog audit events with admin user (from request ctx),     │
 │    │           │ plugin name, prior CB state, and consecutive failure count at time of trip.                           │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ S6 │ Medium    │ If Drain() returns false (timeout), trip the plugin-level circuit breaker to prevent a               │
 │    │           │ reload-drain-timeout cycle from masking stuck handlers. Requires admin Reset to re-enable.            │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ S7 │ Medium    │ Schema drift warnings use slog.Warn (not Info/Debug) and drift details are surfaced in               │
 │    │           │ PluginInfoHandler admin response so operators can see stale schemas without digging through logs.     │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ S8 │ Medium    │ Orphaned table name matching uses known-plugin prefix matching (tablePrefix function pattern),       │
 │    │           │ not delimiter parsing. Plugin names can contain underscores, so "second _ separator" is unreliable.   │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ S9 │ Low       │ Per-plugin reload cooldown (10s) on the admin reload endpoint to prevent reload storms.              │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │S10 │ Low       │ Plugin_Hot_Reload explicitly defaults to false (zero value). Production opt-in only.                 │
 ├────┼───────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │S11 │ Low       │ Verify StopWatcher() is called before bridge.Close() in shutdown sequence to prevent reload          │
 │    │           │ racing with shutdown.                                                                                │
 └────┴───────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────┘

 Architect Review Fixes

 5 concerns raised during skeptical-architect review of the original plan.
 All are addressed in the sections below.

 ┌────┬──────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │ #  │ Severity │ Fix                                                                                                  │
 ├────┼──────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ A1 │ Critical │ Go's http.ServeMux panics on duplicate Handle() calls. Never unregister patterns from the mux.      │
 │    │          │ UnregisterPlugin only clears the in-memory routes map; ServeHTTP already returns 404 for             │
 │    │          │ unmatched map entries. Add registeredPatterns set to HTTPBridge to skip re-registration of            │
 │    │          │ already-mounted patterns. This also fixes a latent panic in ApproveRoute for already-registered      │
 │    │          │ patterns.                                                                                            │
 ├────┼──────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ A2 │ Critical │ Drain() polling via len(p.general)+len(p.reserved) has a data race: between the two len() reads,    │
 │    │          │ a concurrent Put() can move a VM between channels, causing false positive (premature close) or       │
 │    │          │ false negative (timeout). Replace with atomic Int32 checkout counter: increment in Get/GetForHook,   │
 │    │          │ decrement in Put, Drain waits for counter to reach zero.                                             │
 ├────┼──────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ A3 │ High     │ Reload sequence drains old pool (step 6) before creating new instance (step 9). If loadPlugin       │
 │    │          │ fails (syntax error, missing dep), the plugin is dead with no rollback. Fix: blue-green reload --    │
 │    │          │ create new instance FIRST, only drain old after new is confirmed working. If new fails, old keeps    │
 │    │          │ running with a logged warning.                                                                       │
 ├────┼──────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ A4 │ Medium   │ Admin endpoint /api/v1/admin/plugins/cleanup would be caught by {name} wildcard if not explicitly   │
 │    │          │ registered first. MountAdminEndpoints must register all literal-path endpoints (cleanup, routes)     │
 │    │          │ before wildcard {name} patterns. Go 1.22+ ServeMux gives literal segments precedence over            │
 │    │          │ wildcards, but explicit registration order documents intent.                                         │
 ├────┼──────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ A5 │ Medium   │ checkVersionChange, readRouteApproval, ApproveRoute, RevokeRoute, and the DELETE in                 │
 │    │          │ RegisterRoutes all use ? placeholders -- broken on PostgreSQL (needs $N). Fix: add dialect-aware     │
 │    │          │ query branches matching the pattern already used by upsertRoute and CleanupOrphanedRoutes.           │
 │    │          │ Pre-existing bug, but Phase 4 hot reload makes all these paths reachable from the new reload code.   │
 └────┴──────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────┘

 ---
 New Files

 1. internal/plugin/recovery.go (~250 lines)

 Plugin-level circuit breaker + panic-safe execution wrapper. Scope boundary:
  this CB tracks consecutive failures from HTTP handler execution and manager
 operations (reload, init) only. Hook execution failures are tracked by the
 existing hook-level circuit breaker in hook_engine.go and do NOT feed into the
  plugin-level CB. This prevents a buggy hook from disabling working HTTP routes.

 type CircuitState int // CircuitClosed, CircuitOpen, CircuitHalfOpen

 type CircuitBreaker struct {
     mu                sync.Mutex
     state             CircuitState
     consecutiveErrors int
     maxFailures       int           // config Plugin_Max_Failures, default 5
     resetInterval     time.Duration // config Plugin_Reset_Interval, default 60s
     lastFailure       time.Time
     pluginName        string
 }

 func NewCircuitBreaker(pluginName string, maxFailures int, resetInterval
 time.Duration) *CircuitBreaker
 func (cb *CircuitBreaker) Allow() bool           // Closed=yes, Open=yes if
 reset interval elapsed (->HalfOpen), HalfOpen=yes (probe)
 func (cb *CircuitBreaker) RecordSuccess()         // resets count, ->Closed
 func (cb *CircuitBreaker) RecordFailure() bool    // returns true if just
 tripped
 func (cb *CircuitBreaker) State() CircuitState
 func (cb *CircuitBreaker) Reset(adminUser string)  // admin force reset ->Closed
 // security fix S5: Reset emits slog.Warn with adminUser, plugin name,
 // prior state, and consecutive failure count. The adminUser string is
 // extracted from the request context by the caller (EnablePlugin handler)
 // and passed in directly -- Reset does not accept context.Context because
 // it is a pure state transition with no I/O.

 type SafeExecuteResult struct {
     Err      error
     Panicked bool
     PanicVal any
 }

 func SafeExecute(cb *CircuitBreaker, fn func() error) SafeExecuteResult

 SafeExecute: checks cb.Allow() first (returns error if open), runs fn with
  defer/recover, records success/failure on circuit breaker, emits metrics.

 Circuit breaker interaction design decision:

 ┌──────────────────────────────────┬────────────────────┬─────────────────────────────────────┐
 │          Failure Source          │     Which CB?      │              Rationale              │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ HTTP handler error/panic/timeout │ Plugin-level CB    │ HTTP routes are the                 │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ plugin's primary interface       │                    │                                     │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ HTTP VM pool exhaustion          │ Plugin-level CB    │ Indicates plugin health             │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ degradation                      │                    │                                     │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ Reload init failure              │ Plugin-level CB    │ Plugin code is broken               │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ Before-hook abort (error())      │ Hook-level CB only │ Hook is buggy, but HTTP             │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ routes may work fine             │                    │                                     │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ After-hook error                 │ Hook-level CB only │ Fire-and-forget; does not indicate  │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ plugin failure                   │                    │                                     │
 ├──────────────────────────────────┼────────────────────┼─────────────────────────────────────┤
 │ Hook VM pool exhaustion          │ Neither            │ Transient contention, not a failure │
 └──────────────────────────────────┴────────────────────┴─────────────────────────────────────┘

 Add CB *CircuitBreaker field to PluginInstance, created during loadPlugin.

 Test file: recovery_test.go (~300 lines) -- state transitions, panic recovery,
  concurrent safety, circuit trip/reset.

 2. internal/plugin/metrics.go (~180 lines)

 Metric name constants and recording helper functions wrapping
 utility.GlobalMetrics. No new types needed -- utility.Labels,
 utility.GlobalMetrics already exist.

 Instrumentation philosophy: Instrument coarse-grained boundaries (HTTP
 request lifecycle, hook event lifecycle, reload events) where the metric
 recording cost is negligible relative to the operation cost. Do NOT instrument
 per-Get()/Put()/checkOpLimit() hot paths -- GlobalMetrics uses
 sync.RWMutex internally, and adding mutex contention to every VM checkout and
 DB op is not justified until profiling proves otherwise.

 // Constants
 const (
     MetricHTTPRequests       = "plugin.http.requests"
     MetricHTTPDuration       = "plugin.http.duration_ms"
     MetricHookBefore         = "plugin.hook.before"
     MetricHookAfter          = "plugin.hook.after"
     MetricHookDuration       = "plugin.hook.duration_ms"
     MetricErrors             = "plugin.errors"
     MetricCircuitBreakerTrip = "plugin.circuit_breaker.trip"
     MetricReload             = "plugin.reload"
     MetricVMAvailable        = "plugin.vm.available"
 )

 // Helper functions -- coarse-grained recording only
 func RecordHTTPRequest(pluginName, method string, status int, durationMs
 float64)
 func RecordHookExecution(metricName, pluginName, event, table, status string,
 durationMs float64)
 func RecordReload(pluginName, status string)
 func SnapshotVMAvailability(plugins map[string]*PluginInstance)  // periodic
 gauge snapshot by watcher

 Instrumentation points (additions to existing files):
 - http_bridge.go ServeHTTP: timing + RecordHTTPRequest (one call per HTTP
 request -- negligible vs. Lua execution cost)
 - hook_engine.go executeBefore/executeAfter: RecordHookExecution (one call
 per hook event -- negligible vs. DB + Lua cost)
 - recovery.go SafeExecute: RecordCircuitBreakerTrip on trip (rare event)
 - watcher.go poll tick: SnapshotVMAvailability (every 2s -- negligible)

 Deferred until profiling justifies: per-Get()/Put() checkout timing,
 per-DB-op counting. These can be added later behind a Plugin_Verbose_Metrics
 flag if needed.

 Test file: metrics_test.go (~120 lines) -- verify correct metric names/labels
 emitted.

 3. internal/plugin/watcher.go (~400 lines)

 File-polling hot reload with debounce and graceful per-plugin restart.

 type Watcher struct {
     manager        *Manager
     pollInterval   time.Duration     // default 2s
     debounceDelay  time.Duration     // default 1s -- wait for file stability
 before reload
     checksums      map[string]string // pluginName -> SHA-256 of all .lua files
     pendingReloads map[string]pendingReload // pluginName -> checksum + first seen time
 }
 type pendingReload struct {
     checksum  string
     firstSeen time.Time
     mu             sync.Mutex
     reloadMu       sync.Mutex        // serialize reloads (one at a time)
     logger         *utility.Logger
 }

 func NewWatcher(manager *Manager, pollInterval time.Duration) *Watcher
 func (w *Watcher) Run(ctx context.Context)           // blocking poll loop,
 cancel via ctx
 func (w *Watcher) InitialChecksums()                  // baseline after LoadAll
 func (w *Watcher) poll() []string                     // returns changed plugin
 names
 func (w *Watcher) reloadPlugin(ctx context.Context, pluginName string) error
 func computeChecksum(dir string) (string, error)      // SHA-256 of all .lua
 files -- security fix S4: uses os.Lstat, skips symlinks and non-regular
 files, enforces max 100 .lua files per directory and max 10 MB total bytes
 per checksumming pass. Returns error if limits exceeded (plugin skipped for
 change detection with logged warning).

 Debounce logic: When poll() detects a checksum change for a plugin, it
 records the plugin name in pendingReloads with the current time. On each
 subsequent poll tick, it re-checksums pending plugins. If the checksum is still
 different from baseline AND has been stable (unchanged) for debounceDelay
 (1s), the reload fires. If the checksum changes again within the debounce
 window, the timer resets. This prevents reloading mid-save when an editor writes
  multiple files sequentially.

 poll tick 1: detect change for "task_tracker" → pendingReloads["task_tracker"] =
  {checksum: "abc", time: now}
 poll tick 2 (2s later): re-checksum → same "abc" → 2s > 1s debounce → trigger
 reload
 -- OR --
 poll tick 2: re-checksum → changed to "def" → reset:
 pendingReloads["task_tracker"] = {checksum: "def", time: now}
 poll tick 3: re-checksum → same "def" → 2s > 1s debounce → trigger reload

 Reload sequence (blue-green -- architect fix A3 + security fix S3):

 The original plan drained the old pool BEFORE creating the new instance.
 This meant a failed reload (syntax error, missing dependency) left the
 plugin dead with no rollback -- the old pool was already destroyed.

 The fix uses a blue-green approach: create the new instance first, only
 drain the old after the new is confirmed working. If the new instance
 fails to load, the old instance keeps running and a warning is logged.

 1. Try-lock reloadMu (non-blocking). If already held, log warning
 "reload already in progress for another plugin, skipping %s" and return
 without blocking. This is the ONLY acquisition of reloadMu -- there is no
 separate blocking Acquire. This prevents a queue of reload requests from
 compounding drain wait times. Track consecutive slow reloads (>10s) per
 plugin -- after 3 consecutive slow reloads, pause watcher polling for that
 plugin and emit a warning. The admin reload endpoint can still force a
 reload. Defer reloadMu.Unlock() immediately after successful try-lock.
 2. Read current PluginInstance as oldInst under Manager.mu.RLock()
 3. Re-extract manifest (version may have changed)
 4. Check dependency validity: if new manifest has different dependencies,
 verify they exist and are StateRunning. If unmet, log warning and abort
 reload (old instance keeps running).
 5. Create new PluginInstance independently (new pool, new dbAPIs map --
 shares nothing with old instance). The new instance starts in StateLoading.
 6. Call loadPlugin(ctx, newInst) -- creates new pool, runs on_init,
 registers routes/hooks via bridge and hookEngine.

 ** DECISION POINT: did loadPlugin succeed? **

 If loadPlugin FAILED (newInst.State == StateFailed):
   - Old instance stays running. No state change, no drain.
   - Log slog.Warn: "plugin %q reload failed: %s (old version still running)"
   - Record RecordReload(name, "error")
   - Close newInst.Pool (cleanup the failed attempt)
   - Return error (watcher logs it, continues polling for next change)

 If loadPlugin SUCCEEDED (newInst.State == StateRunning):
   7. Set oldInst.State = StateLoading (bridge/hooks skip loading plugins)
   8. Call bridge.UnregisterPlugin(pluginName) -- clears OLD in-memory routes
   only (new routes were registered in step 6; DB approval rows are preserved)
   9. Call hookEngine.UnregisterPlugin(pluginName) -- clears OLD in-memory hook
   index entries (new hooks were registered in step 6; DB approval rows preserved)
   10. Under Manager.mu.Lock(), swap m.plugins[name] to new instance. From
   this point, new requests go to the new instance.
   11. Drain old pool via oldInst.Pool.Drain(10s):
     - Sets draining flag (old pool's Get/GetForHook return ErrPoolExhausted)
     - Waits for checked-out VMs to be returned via Put() using atomic counter
     - After all VMs are back (or timeout): set closed, drain channels, close
   all VMs
     - Security fix S6: if Drain returns false (timeout), trip the plugin-level
   circuit breaker on the NEW instance, since stuck old handlers suggest the
   plugin has systemic issues. Requires admin Reset to re-enable.
   12. Record RecordReload(name, "success")

 Why blue-green is safe: Steps 6-10 create the new instance while the old
 is still serving. The swap in step 10 is atomic under Manager.mu.Lock.
 The drain in step 11 happens AFTER the swap, so new requests go to the
 new instance while old in-flight requests complete against the old pool.

 Route registration during blue-green: loadPlugin (step 6) calls
 bridge.RegisterRoutes which adds entries to the in-memory routes map and
 calls mux.Handle for new patterns. For patterns that already exist (same
 routes in new version), the registeredPatterns set (fix A1) prevents
 duplicate mux.Handle panics. The old routes for removed endpoints stay in
 the mux but are cleared from the in-memory map in step 8, so ServeHTTP
 returns 404 for them.

 Hook registration during blue-green: loadPlugin (step 6) calls
 hookEngine.RegisterHooks which adds entries to hookIndex. The old entries
 are removed in step 9 via UnregisterPlugin. Between steps 6-9, both old
 and new hook entries coexist in the index -- this is harmless because
 hook dispatch filters by plugin name and the entries have distinct
 handlerKeys tied to different VMs.

 New methods needed on existing types:
 - VMPool.Drain(timeout time.Duration) bool -- see pool.go section below
 - HTTPBridge.UnregisterPlugin(ctx, pluginName) -- remove in-memory route
 registrations
 - HookEngine.UnregisterPlugin(pluginName) -- remove in-memory hook index
 entries, rebuild hasHooks/hasAnyHook

 Test file: watcher_test.go (~300 lines) -- checksum determinism, change
 detection, debounce behavior, reload state transitions.

 4. internal/plugin/cli_commands.go (~350 lines)

 HTTP admin endpoints for plugin management. Registered via
 HTTPBridge.MountAdminEndpoints(mux, authChain) to avoid changing
 NewModulacmsMux signature (bridge already holds manager reference).

 Admin HTTP handlers (explicit authorization levels -- security fix S2):

 func PluginListHandler(mgr *Manager) http.Handler        // GET
 /api/v1/admin/plugins              -- any authenticated (read-only)
 func PluginInfoHandler(mgr *Manager) http.Handler        // GET
 /api/v1/admin/plugins/{name}       -- any authenticated (read-only)
 func PluginReloadHandler(mgr *Manager) http.Handler      // POST
 /api/v1/admin/plugins/{name}/reload  -- adminOnly (mutating, resource intensive)
 // security fix S9: enforces 10s per-plugin cooldown between reloads.
 // Returns 429 Too Many Requests if called within cooldown window.
 func PluginEnableHandler(mgr *Manager) http.Handler      // POST
 /api/v1/admin/plugins/{name}/enable  -- adminOnly (changes plugin state)
 func PluginDisableHandler(mgr *Manager) http.Handler     // POST
 /api/v1/admin/plugins/{name}/disable -- adminOnly (changes plugin state)
 func PluginCleanupListHandler(mgr *Manager) http.Handler // GET
 /api/v1/admin/plugins/cleanup        -- adminOnly (reveals internal DB schema)
 func PluginCleanupDropHandler(mgr *Manager) http.Handler // POST
 /api/v1/admin/plugins/cleanup        -- adminOnly (destructive: drops tables)

 MountAdminEndpoints registration order (architect fix A4):

 Go 1.22+ ServeMux gives literal path segments precedence over wildcards.
 However, to make intent explicit and prevent confusion, MountAdminEndpoints
 registers all literal-path endpoints BEFORE wildcard {name} patterns.
 The adminOnly function is passed in from mux.go (where it is defined).

 func (b *HTTPBridge) MountAdminEndpoints(
     mux *http.ServeMux,
     authChain func(http.Handler) http.Handler,
     adminOnlyFn func(http.Handler) http.Handler,
 ) {
     mgr := b.manager

     // 1. Register literal-path endpoints first (these beat {name} wildcard):
     mux.Handle("GET /api/v1/admin/plugins",
         authChain(PluginListHandler(mgr)))

     mux.Handle("GET /api/v1/admin/plugins/cleanup",
         authChain(adminOnlyFn(PluginCleanupListHandler(mgr))))
     mux.Handle("POST /api/v1/admin/plugins/cleanup",
         authChain(adminOnlyFn(PluginCleanupDropHandler(mgr))))

     // 2. Register wildcard {name} endpoints:
     mux.Handle("GET /api/v1/admin/plugins/{name}",
         authChain(PluginInfoHandler(mgr)))
     mux.Handle("POST /api/v1/admin/plugins/{name}/reload",
         authChain(adminOnlyFn(PluginReloadHandler(mgr))))
     mux.Handle("POST /api/v1/admin/plugins/{name}/enable",
         authChain(adminOnlyFn(PluginEnableHandler(mgr))))
     mux.Handle("POST /api/v1/admin/plugins/{name}/disable",
         authChain(adminOnlyFn(PluginDisableHandler(mgr))))
 }

 Precedence verification (Go 1.22+ ServeMux rules):
 - GET /api/v1/admin/plugins           -> PluginListHandler (exact match)
 - GET /api/v1/admin/plugins/cleanup   -> PluginCleanupListHandler (literal
   "cleanup" beats wildcard {name})
 - GET /api/v1/admin/plugins/routes    -> pluginRoutesListHandler (already
   registered in mux.go, literal "routes" beats {name})
 - GET /api/v1/admin/plugins/my_plugin -> PluginInfoHandler ({name} wildcard)
 - POST /api/v1/admin/plugins/cleanup  -> PluginCleanupDropHandler (literal)

 New Manager methods:
 func (m *Manager) ReloadPlugin(ctx context.Context, name string) error
 func (m *Manager) DisablePlugin(ctx context.Context, name string) error  //
 ->StateStopped, trip CB
 func (m *Manager) EnablePlugin(ctx context.Context, name string) error   //
 reset CB, reload -- security fix S5: emits slog.Warn audit event with admin
 user (from ctx), plugin name, prior CB state, and failure count before reset
 func (m *Manager) ListOrphanedTables(ctx context.Context) ([]string, error)
 func (m *Manager) DropOrphanedTables(ctx context.Context, requestedTables []string) error

 Orphaned table detection (safe by design):

 Orphaned = table has plugin_ prefix AND its plugin name prefix does NOT match
 ANY plugin in m.plugins regardless of state. Key: StateFailed and
 StateStopped plugins are NOT considered orphaned -- they are known plugins
 with real data.

 Per-dialect queries:
 - SQLite: SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'plugin_%'
 - MySQL: SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name LIKE 'plugin_%'
 - PostgreSQL: SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'plugin_%'

 Orphaned table name matching (security fix S1 -- no delimiter parsing):

 Plugin names can contain underscores (e.g., task_tracker), so parsing by
 "the second _ separator" would misidentify tables. Instead, for each table
 with plugin_ prefix, check if tableName starts with "plugin_" + knownName +
 "_" for every known plugin name in m.plugins (all states). Use the same
 tablePrefix function from db_api.go (which builds "plugin_" + pluginName +
 "_") for consistent matching. If no known plugin claims the table, it is
 orphaned. System tables plugin_routes and plugin_hooks are excluded by name.

 DropOrphanedTables safety measures (security fix S1 -- closed TOCTOU):

 DropOrphanedTables does NOT trust the caller's requestedTables list
 directly. Instead:
 1. Internally re-calls ListOrphanedTables to get the current orphan set
 2. Intersects with requestedTables -- only tables in BOTH lists proceed
 3. Validates each table name via db.ValidTableName (identifier injection
 prevention -- SQL parameterized queries cannot bind table names, so
 identifier allowlist validation is mandatory). db.ValidTableName is
 defined in internal/db/query_builder.go and validates: first character
 must be lowercase a-z, remaining characters must be lowercase a-z,
 digits 0-9, or underscore, minimum length 1, maximum length 64. Use
 a character-by-character loop (no regex). If the function does not
 already exist, it MUST be created there before DropOrphanedTables can
 use it. Delegate creation to modulacms-db-developer if needed.
 4. Re-verifies orphan status at execution time, closing the TOCTOU window
 between the dry-run GET and the destructive POST

 Additional safety measures (unchanged from architecture review):
 1. GET endpoint returns dry-run list -- shows what WOULD be dropped, never
 drops
 2. POST requires {"confirm": true} in body -- missing or false returns 400
 3. Excludes tables for any known plugin (all states including StateFailed,
 StateStopped)
 4. Excludes system tables (plugin_routes, plugin_hooks)
 5. Logs every DROP with table name before execution
 6. Returns list of dropped tables in response for audit trail

 Test file: cli_commands_test.go (~200 lines) -- orphaned table detection,
 StateFailed exclusion, confirm requirement, enable/disable, handler JSON output.

 ---
 Modifications to Existing Files

 internal/config/config.go -- add 3 fields

 Plugin_Hot_Reload     bool   `json:"plugin_hot_reload"`      // default false
 (zero value) -- production opt-in only (security fix S10)
 Plugin_Max_Failures   int    `json:"plugin_max_failures"`    // circuit breaker,
  default 5
 Plugin_Reset_Interval string `json:"plugin_reset_interval"`  // default "60s"

 internal/plugin/manager.go -- add fields + methods

 - Add to ManagerConfig: HotReload bool, MaxFailures int, ResetInterval time.Duration
 - Add to Manager: watcher *Watcher
 - Add to PluginInstance: CB *CircuitBreaker
 - Modify NewManager: parse defaults for MaxFailures (5) and ResetInterval
 (60s)
 - Modify loadPlugin: create CircuitBreaker per plugin
 - Add methods: StartWatcher(ctx), ReloadPlugin(ctx, name),
 DisablePlugin(ctx, name), EnablePlugin(ctx, name),
 ListOrphanedTables(ctx), DropOrphanedTables(ctx, tables)

 internal/plugin/pool.go -- tri-state lifecycle + Drain method

 Critical fix: tri-state pool lifecycle. The existing pool has a single
 closed atomic.Bool. Adding Drain() requires a draining atomic.Bool that is
 distinct from closed.

 Architect review fix A2: The original plan polled len(p.general) +
 len(p.reserved) to detect when all VMs were returned. This has a data race:
 between reading the two channel lengths, a concurrent Put() via
 returnToPool() can move a VM from one channel to the other, causing
 premature closure (false positive) or missed target (timeout). The fix uses
 an atomic Int32 checkout counter that is incremented in Get/GetForHook and
 decremented in Put. Drain waits for the counter to reach zero.

 ┌─────────────────────────────┬──────────┬────────┬─────────────────────────┬─────────────────────────┐
 │            State            │ draining │ closed │     Get() behavior      │     Put() behavior      │
 ├─────────────────────────────┼──────────┼────────┼─────────────────────────┼─────────────────────────┤
 │ Open                        │ false    │ false  │ Normal checkout         │ Return to channel       │
 ├─────────────────────────────┼──────────┼────────┼─────────────────────────┼─────────────────────────┤
 │ Draining                    │ true     │ false  │ Return ErrPoolExhausted │ Return to channel (VMs  │
 ├─────────────────────────────┼──────────┼────────┼─────────────────────────┼─────────────────────────┤
 │ accumulate for drain count) │          │        │                         │                         │
 ├─────────────────────────────┼──────────┼────────┼─────────────────────────┼─────────────────────────┤
 │ Closed                      │ true     │ true   │ Return ErrPoolExhausted │ Call L.Close() directly │
 └─────────────────────────────┴──────────┴────────┴─────────────────────────┴─────────────────────────┘

 // Add to VMPool struct:
 draining   atomic.Bool   // true during Drain() -- rejects Get but accepts Put returns
 checkedOut atomic.Int32  // A2: tracks VMs currently checked out (incremented in Get, decremented in Put)

 Modify Get(): add draining check + increment counter:
 func (p *VMPool) Get(ctx context.Context) (*lua.LState, error) {
     if p.draining.Load() {
         return nil, ErrPoolExhausted
     }
     acquireCtx, cancel := context.WithTimeout(ctx, acquireTimeout)
     defer cancel()
     select {
     case L := <-p.general:
         p.checkedOut.Add(1)  // A2: track checkout
         L.SetContext(ctx)
         return L, nil
     case <-acquireCtx.Done():
         return nil, ErrPoolExhausted
     }
 }

 Modify GetForHook(): add draining check + increment counter:
 func (p *VMPool) GetForHook(ctx context.Context) (*lua.LState, error) {
     if p.draining.Load() {
         return nil, ErrPoolExhausted
     }
     // Try general pool first (non-blocking).
     select {
     case L := <-p.general:
         p.checkedOut.Add(1)  // A2: track checkout
         L.SetContext(ctx)
         return L, nil
     default:
     }
     // Fall back to reserved with acquisition timeout.
     acquireCtx, cancel := context.WithTimeout(ctx, acquireTimeout)
     defer cancel()
     select {
     case L := <-p.reserved:
         p.checkedOut.Add(1)  // A2: track checkout
         L.SetContext(ctx)
         return L, nil
     case <-acquireCtx.Done():
         return nil, ErrPoolExhausted
     }
 }

 Modify Put(): decrement counter before returning VM:
 func (p *VMPool) Put(L *lua.LState) {
     p.checkedOut.Add(-1)  // A2: track return (before any channel operations)
     // ... rest of existing Put logic unchanged ...
 }

 func (p *VMPool) Drain(timeout time.Duration) bool {
     p.draining.Store(true)
     // A2: Wait for all checked-out VMs to be returned via atomic counter.
     // This avoids the data race in len(general)+len(reserved) snapshots.
     deadline := time.After(timeout)
     ticker := time.NewTicker(50 * time.Millisecond)
     defer ticker.Stop()
     for {
         select {
         case <-deadline:
             // Timeout: some VMs still checked out. Log warning, proceed to
             // close.
             // Security fix S6: caller (watcher.reloadPlugin) must trip the
             // plugin-level circuit breaker when Drain returns false, to
             // prevent a reload-drain-timeout cycle from masking stuck
             // handlers. Requires admin Reset to re-enable the plugin.
             p.closed.Store(true)
             p.drainChannels()
             return false
         case <-ticker.C:
             n := p.checkedOut.Load()
             if n < 0 {
                 // Counter went negative -- indicates a bug (Put without matching Get).
                 // Log warning but treat as drained to avoid infinite loop.
                 utility.DefaultLogger.Warn(fmt.Sprintf("plugin %q: checkout counter is negative (%d), possible double-Put", p.pluginName, n), nil)
             }
             if n <= 0 {
                 // All VMs returned. Close them.
                 p.closed.Store(true)
                 p.drainChannels()
                 return true
             }
         }
     }
 }

 Note: AvailableCount() remains as-is for diagnostics (it documents that
 it is "not safe for capacity decisions"). The atomic counter is the
 authoritative source for Drain correctness.

 internal/plugin/http_bridge.go -- add UnregisterPlugin + SafeExecute +
 mux safety + PostgreSQL placeholder fix + metrics

 Architect fix A1 -- registeredPatterns set:

 Add to HTTPBridge struct:
 registeredPatterns map[string]bool  // mux patterns already registered via Handle()

 Initialize in NewHTTPBridge:
 registeredPatterns: make(map[string]bool),

 Add helper method:
 func (b *HTTPBridge) registerOnMux(pattern string) {
     if b.mux == nil {
         return
     }
     if b.registeredPatterns[pattern] {
         return  // already registered -- skip to avoid ServeMux panic
     }
     b.mux.Handle(pattern, b)
     b.registeredPatterns[pattern] = true
 }

 Replace all calls to b.mux.Handle(pattern, b) with b.registerOnMux(pattern):
 - RegisterRoutes line 367: b.registerOnMux(muxPattern)
 - MountOn loop: b.registerOnMux(pattern)
 - MountOn fallback: b.registerOnMux(PluginRoutePrefix) -- only once
 - ApproveRoute line 818: b.registerOnMux(muxPattern)

 This eliminates the latent panic in ApproveRoute when re-approving an
 already-registered pattern, and prevents panics during hot reload when
 the new plugin version re-registers the same route patterns.

 Mux pattern lifetime: patterns are never removed from http.ServeMux (Go
 stdlib does not support it). Stale mux entries dispatch to ServeHTTP, which
 checks the in-memory routes map and returns 404 for unmatched entries.
 Over months of uptime with frequent reloads, stale entries accumulate but
 are negligible in memory (one map entry per historical route pattern).

 Architect fix A5 -- PostgreSQL placeholder fix:

 The following methods use ? placeholders which fail on PostgreSQL. Fix by
 adding dialect switch branches matching the existing pattern in upsertRoute:

 checkVersionChange: (existing line 381)
   switch b.dialect {
   case db.DialectPostgres:
       query = "SELECT plugin_version FROM plugin_routes WHERE plugin_name = $1 LIMIT 1"
   default:
       query = "SELECT plugin_version FROM plugin_routes WHERE plugin_name = ? LIMIT 1"
   }

 readRouteApproval: (existing line 483)
   switch b.dialect {
   case db.DialectPostgres:
       query = "SELECT approved, public FROM plugin_routes WHERE plugin_name = $1 AND method = $2 AND path = $3"
   default:
       query = "SELECT approved, public FROM plugin_routes WHERE plugin_name = ? AND method = ? AND path = ?"
   }

 RegisterRoutes DELETE on version change: (existing line 322)
   switch b.dialect {
   case db.DialectPostgres:
       delQuery = "DELETE FROM plugin_routes WHERE plugin_name = $1"
   default:
       delQuery = "DELETE FROM plugin_routes WHERE plugin_name = ?"
   }

 ApproveRoute UPDATE: (existing line 806)
   switch b.dialect {
   case db.DialectPostgres:
       query = "UPDATE plugin_routes SET approved = TRUE, approved_at = $1, approved_by = $2 WHERE plugin_name = $3 AND method = $4 AND path = $5"
   default:
       query = "UPDATE plugin_routes SET approved = 1, approved_at = ?, approved_by = ? WHERE plugin_name = ? AND method = ? AND path = ?"
   }

 RevokeRoute UPDATE: (existing line 839)
   switch b.dialect {
   case db.DialectPostgres:
       query = "UPDATE plugin_routes SET approved = FALSE, approved_at = NULL, approved_by = NULL WHERE plugin_name = $1 AND method = $2 AND path = $3"
   default:
       query = "UPDATE plugin_routes SET approved = 0, approved_at = NULL, approved_by = NULL WHERE plugin_name = ? AND method = ? AND path = ?"
   }

 Other changes (unchanged from original plan):

 - Add UnregisterPlugin(ctx context.Context, pluginName string) -- removes
 in-memory routes for plugin (DB rows kept for approval persistence). Does
 NOT touch the mux (fix A1: patterns are never removed from ServeMux).
 - Add MountAdminEndpoints(mux *http.ServeMux, authChain func(http.Handler)
 http.Handler, adminOnlyFn func(http.Handler) http.Handler) -- registers
 admin plugin management endpoints on the mux. Called from mux.go instead
 of changing NewModulacmsMux signature. Takes adminOnly as parameter since
 it is defined in mux.go (see cli_commands.go section for registration order).
 - Wrap handler call in SafeExecute(inst.CB, ...) in ServeHTTP
 - Check inst.CB.Allow() before VM checkout (returns 503 if open)
 - Add RecordHTTPRequest timing around handler execution

 internal/plugin/hook_engine.go -- add UnregisterPlugin + metrics

 - Add UnregisterPlugin(pluginName string) -- removes entries from hookIndex
 for the given plugin, rebuilds hasHooks map and hasAnyHook flag from
 remaining entries
 - Add RecordHookExecution in executeBefore/executeAfter (one call per event,
 not per hook)
 - No changes to hook-level circuit breaker -- it remains independent from
 plugin-level CB

 internal/plugin/schema_api.go -- schema drift detection (Known Gap #9)

 After DDLCreateTable call in luaDefineTable, add:
 - introspectTableColumns(ctx, exec, dialect, tableName) ([]string, error) --
 queries actual column names per dialect
 - checkSchemaDrift(pluginName, fullName, expected, actual) -- compares and
 logs warnings for missing/extra columns
 - Non-fatal: drift is advisory (logs at slog.Warn level, does not error)
 - Security fix S7: drift results are stored on the PluginInstance (e.g.,
 SchemaDrift []DriftEntry field) and surfaced in PluginInfoHandler admin
 response, so operators can see stale schemas from the API without digging
 through log files.

 Operator action on drift: The slog.Warn includes the specific
 missing/extra column names and a message: "plugin %q table %q has schema drift -- add migration logic in on_init() or recreate the table". The operator can
 either update the plugin code to handle the drift or manually ALTER TABLE. Full
 migration support is out of scope for Phase 4.

 internal/router/mux.go -- register admin endpoints via bridge

 No signature change to NewModulacmsMux. Instead, the bridge exposes
 MountAdminEndpoints. The adminOnly function (already defined in mux.go)
 is passed as a parameter so the bridge does not duplicate auth logic.

 // In mux.go, after bridge.MountOn(mux):
 if bridge != nil {
     bridge.MountOn(mux)
     bridge.MountAdminEndpoints(mux, authChain, adminOnly)
 }

 MountAdminEndpoints registers the 7 admin endpoints using bridge.manager
 internally. This avoids passing *Manager separately -- the bridge already
 holds the reference. The adminOnly wrapper is defined in mux.go and passed
 in because it accesses middleware.AuthenticatedUser which is a router-layer
 concern, not a plugin-layer concern.

 cmd/helpers.go -- update initPluginManager

 - Parse Plugin_Max_Failures (default 5) and Plugin_Reset_Interval (parse
 duration, default "60s")
 - Pass HotReload, MaxFailures, ResetInterval to ManagerConfig
 - After LoadAll, if HotReload && mgr != nil: call mgr.StartWatcher(ctx)

 cmd/serve.go -- watcher shutdown

 - No signature changes to NewModulacmsMux
 - Watcher goroutine is cancelled by rootCancel() (already in shutdown flow, no
  new code needed)
 - Security fix S11: Watcher MUST stop before bridge.Close() to prevent
 reload racing with shutdown. Add mgr.StopWatcher() call before bridge
 shutdown in the signal handler. Shutdown order: StopWatcher → bridge.Close
 → hookEngine.Close → manager.Shutdown. Verify this order in integration
 tests.

 ---
 Execution Order

 Parallel Group 1 (no dependencies between them)

 - Agent A: recovery.go + recovery_test.go (self-contained)
 - Agent B: metrics.go + metrics_test.go (self-contained)
 - Agent C: Config additions (config.go -- 3 fields)
 - Agent D: Schema drift detection in schema_api.go + test additions

 Parallel Group 2 (depends on Group 1)

 - Agent E: pool.go modifications (tri-state lifecycle, atomic checkout
 counter [A2], Drain method) -- needs metrics.go constants
 - Agent F: manager.go modifications (CB field, new methods, watcher start)
  -- needs recovery.go, config
 - Agent G: hook_engine.go modifications (UnregisterPlugin, metrics) --
 needs metrics.go

 Parallel Group 3 (depends on Group 2)

 - Agent H: watcher.go + watcher_test.go -- needs manager changes, pool
 Drain, bridge/hook UnregisterPlugin. Implements blue-green reload [A3].
 - Agent I: http_bridge.go modifications (UnregisterPlugin,
 registeredPatterns [A1], PostgreSQL placeholder fix [A5],
 MountAdminEndpoints with explicit route order [A4], SafeExecute, metrics)
 -- needs recovery.go, metrics.go

 Sequential (depends on all above)

 - Agent J: cli_commands.go + cli_commands_test.go -- needs all Manager
 methods
 - Agent K: router/mux.go + cmd/helpers.go + cmd/serve.go -- needs
 bridge.MountAdminEndpoints
 - Agent L: Integration tests + test fixture plugins

 ---
 Test Fixtures

 internal/plugin/testdata/plugins/
   reload_plugin/          -- init.lua that can be modified mid-test for reload
   failing_plugin/         -- always errors on handler call (circuit breaker
 testing)

 ---
 Verification

 1. go test ./internal/plugin/ -count=1 -v -- all new + existing tests pass
 2. just test -- no regressions across full test suite
 3. Manual test: place plugin in ./plugins/, start server with
 plugin_hot_reload: true, modify init.lua, wait >1s debounce, verify reload in
 logs
 4. Manual test: create plugin that errors on every call, verify circuit breaker
 trips after 5 failures, verify auto-recovery after 60s
 5. HTTP API: GET /api/v1/admin/plugins returns plugin list with state/metrics
 6. HTTP API: POST /api/v1/admin/plugins/{name}/reload triggers reload
 7. HTTP API: GET /api/v1/admin/plugins/cleanup lists orphaned tables (dry-run)
 8. HTTP API: POST /api/v1/admin/plugins/cleanup with {"confirm": true} drops
  orphaned tables
 9. Verify StateFailed plugin tables are NOT listed as orphaned

 ---
 Key Files Reference

 ┌───────────────────────────────────────┬──────────────────────────────────────────────────────────┐
 │                 File                  │                         Purpose                          │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/plugin/manager.go:102        │ Manager struct -- add watcher, modify                    │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ loadPlugin                            │                                                          │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/plugin/manager.go:61         │ PluginInstance struct -- add CB field                    │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/plugin/pool.go:44            │ VMPool struct -- add draining flag, Drain                │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ method                                │                                                          │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/plugin/http_bridge.go        │ HTTPBridge -- add UnregisterPlugin,                      │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ MountAdminEndpoints, SafeExecute wrap │                                                          │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/plugin/hook_engine.go        │ HookEngine -- add UnregisterPlugin                       │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/plugin/schema_api.go         │ luaDefineTable -- add drift detection after              │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ DDL                                   │                                                          │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/config/config.go             │ Config struct -- add 3 fields                            │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/router/mux.go:212            │ bridge.MountAdminEndpoints call (no signature            │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ change)                               │                                                          │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ cmd/helpers.go:153                    │ initPluginManager -- pass new config fields              │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ cmd/serve.go:259                      │ Shutdown sequence -- add StopWatcher before bridge.Close │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │                                       │                                                          │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/utility/metrics.go           │ GlobalMetrics, Labels -- used by metrics.go              │
 ├───────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ helpers                               │                                                          │
 └───────────────────────────────────────┴──────────────────────────────────────────────────────────┘
