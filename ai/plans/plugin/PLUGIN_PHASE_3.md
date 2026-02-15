 Ready to code?
                                                                                                                                                                                                Here is Claude's plan:
╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌ Phase 3: Content Hooks (Revised)
                                                                                                                                                                                                Context
                                                                                                                                                                                                Phase 1 (Core Engine) and Phase 2 (HTTP Integration) of the plugin system are fully implemented and tested. Phase 3 adds content lifecycle hooks so plugins can react to CMS entity
 mutations. Hooks fire at the audited command layer -- the single chokepoint through which all database mutations pass (internal/db/audited/audited.go). This covers all 26+ audited entities
 without touching any handler code.

 This revision incorporates mitigations for 7 critical/medium concerns identified during architectural review: SQLite deadlock risk, VM pool exhaustion, fast-path performance, shutdown
 safety, error sanitization, wildcard ordering, and float64 coercion. A subsequent security review added 5 more mitigations (M8-M12): per-table hook approval, before-hook circuit breaker,
 after-hook resource budgets, before-hook timeout, and detectStatusTransition specification. An architectural stress-test added 5 structural fixes (S1-S5):
 hook approval DDL, per-event timeout budget, after-hook HasHooks gate, hook observability, and dbAPIs map leak cleanup. A final LLM instruction review
 resolved agent ownership conflicts, added the formal hooks Lua module specification, fixed the Update HasHooks gate, and disambiguated implementation details for multi-agent execution.

 ---
 Design

 Hook Injection Point (unchanged from original)

 The audited.Create/Update/Delete generic functions are the sole mutation path. The injection approach:

 1. Add a HookRunner interface field to AuditContext (nil = no hooks, backward compatible)
 2. Inside audited.Create/Update/Delete, check auditCtx.HookRunner != nil AND HasHooks(event, table) before any work
 3. Before-hooks run inside the transaction (can abort via error, rolling back)
 4. After-hooks run post-commit (fire-and-forget, errors logged)

 Zero changes to existing handlers or command structs.

 Context Injection (unchanged from original)

 A middleware injects HookRunner into request context. AuditContextFromRequest in internal/middleware/audit_helpers.go extracts it automatically. Non-HTTP paths (CLI, bootstrap, SSH TUI) use
 audited.Ctx() directly -- HookRunner stays nil, hooks don't fire. This is an intentional design decision: hooks only fire for HTTP-originated mutations. If hook support for CLI/TUI mutations
 is added in a future phase, the shutdown sequence must account for SSH server draining before hook engine close.

 ---
 Mitigations

 M1: SQLite Deadlock Prevention -- Block db.* in Before-Hooks

 Problem: Before-hooks run inside types.WithTransaction on the main pool. Plugin DatabaseAPI uses a separate *sql.DB pool. SQLite file-level locks mean a read from the plugin pool deadlocks
 when the main pool holds an EXCLUSIVE write lock.

 Solution: Add inBeforeHook bool flag to DatabaseAPI. Check in checkOpLimit() (the existing gate called by every db.* function). HookEngine.executeBefore sets the flag before invoking the
 Lua handler and clears it on defer.

 // db_api.go -- add to checkOpLimit:
 if api.inBeforeHook {
     return fmt.Errorf("plugin %q: db.* calls are not allowed inside before-hooks", api.pluginName)
 }

 After-hooks run post-commit with no transaction held, so db.* calls work normally there.

 Files: internal/plugin/db_api.go, internal/plugin/hook_engine.go

 M2: VM Pool Reservation -- Decouple Hook VMs from HTTP VMs

 Problem: With 4 VMs per plugin, HTTP requests can exhaust the pool. Before-hook pool.Get() fails after 100ms, rolling back the CMS transaction.

 Solution: Split VMPool into two channels: general (for HTTP) and reserved (for hooks). New GetForHook(ctx) method tries general (non-blocking), then falls back to reserved (with timeout).
 Default: 1 reserved VM. Configurable via plugin_hook_reserve_vms.

 // pool.go
 type VMPool struct {
     general  chan *lua.LState // HTTP requests draw from here
     reserved chan *lua.LState // hooks try general first, then reserved
     // ...
 }

 func (p *VMPool) GetForHook(ctx context.Context) (*lua.LState, error) {
     // Try general pool first (non-blocking)
     select {
     case L := <-p.general:
         return L, nil
     default:
     }
     // Fall back to reserved with same acquireTimeout as Get() (100ms default)
     acquireCtx, cancel := context.WithTimeout(ctx, p.acquireTimeout)
     defer cancel()
     select {
     case L := <-p.reserved:
         return L, nil
     case <-acquireCtx.Done():
         return nil, ErrPoolExhausted
     }
 }

 Put() returns VMs to whichever channel has capacity (reserved first). For plugins with no registered hooks, promote the reserved VM to general.

 Files: internal/plugin/pool.go, internal/plugin/pool_test.go, internal/plugin/manager.go, internal/config/config.go

 M3: Fast-Path Check -- Zero-Allocation "Any Hooks?" Gate

 Problem: Without a fast path, every audited mutation pays for an interface call and potential structToMap even when no hooks exist.

 Solution: HookEngine maintains hasAnyHook bool and hasHooks map[string]bool (keyed by "event:table"), built once at registration time. The HookRunner interface includes a HasHooks(event,
 table) bool method. structToMap is called inside RunBeforeHooks/RunAfterHooks, not at the call site.

 // audited.go -- Create, inside transaction:
 if auditCtx.HookRunner != nil && auditCtx.HookRunner.HasHooks(HookBeforeCreate, cmd.TableName()) {
     if err := auditCtx.HookRunner.RunBeforeHooks(ctx, HookBeforeCreate, cmd.TableName(), cmd.Params()); err != nil {
         return err  // rolls back transaction
     }
 }

 When no hooks match: O(1) map lookup, zero allocations, no JSON roundtrip.

 Files: internal/db/audited/hooks.go, internal/plugin/hook_engine.go

 M4: After-Hook Shutdown Awareness

 Problem: After-hook goroutines use context.Background(). During shutdown, they may try to check out VMs from closed pools.

 Solution: HookEngine tracks in-flight after-hooks via sync.WaitGroup and closing atomic.Bool. HookEngine.Close(ctx) sets closing, waits for the WaitGroup with context deadline. Called in
 serve.go shutdown AFTER HTTP servers stop but BEFORE Manager.Shutdown closes VM pools.

 // serve.go shutdown order:
 httpServer.Shutdown(shutdownCtx)
 httpsServer.Shutdown(shutdownCtx)
 bridge.Close(shutdownCtx)
 hookEngine.Close(shutdownCtx)   // drain after-hooks while pools still open
 pluginManager.Shutdown(shutdownCtx) // closes pools

 Files: internal/plugin/hook_engine.go, cmd/serve.go

 M5: Error Sanitization

 Problem: Before-hook error("msg") propagates raw Lua error to HTTP client, potentially leaking internal state.

 Solution: HookError sentinel type in internal/db/audited/hooks.go. Logs the original Lua error message. Returns sanitized "operation blocked by plugin %q" to caller. Handlers can
 type-assert *audited.HookError to return 422 instead of 500.

 type HookError struct {
     PluginName      string
     Event           HookEvent
     Table           string
     originalMessage string  // unexported -- prevents accidental inclusion in HTTP responses
 }

 func (e *HookError) Error() string {
     return fmt.Sprintf("operation blocked by plugin %q", e.PluginName)
 }

 // LogMessage returns the original Lua error message for structured logging only.
 func (e *HookError) LogMessage() string {
     return e.originalMessage
 }

 Files: internal/db/audited/hooks.go, internal/plugin/hook_engine.go

 M6: Wildcard + Specific Ordering

 Rule: At equal priority, specific hooks run before wildcard hooks. At equal priority and wildcard status, registration order is preserved (stable sort).

 func hookEntryLess(a, b hookEntry) bool {
     if a.priority != b.priority { return a.priority < b.priority }
     if a.isWildcard != b.isWildcard { return !a.isWildcard }
     return a.regOrder < b.regOrder
 }

 Files: internal/plugin/hook_engine.go

 M7: Float64 Coercion -- json.Number + Documentation

 Solution: structToMap is a two-step JSON roundtrip: (1) marshal the Go struct to JSON bytes via json.Marshal, (2) decode those bytes back into map[string]any via json.NewDecoder(bytes.NewReader(b))
 with dec.UseNumber(). This two-step approach is necessary because json.Unmarshal into map[string]any coerces numbers to float64, while the decoder path preserves them as json.Number
 (string-backed). Add json.Number case to GoValueToLua in lua_helpers.go. Document that all Go numeric fields become Lua float64 (lossless for integers up to 2^53, CMS IDs are ULID strings
 so not affected).

 Files: internal/db/audited/hooks.go, internal/plugin/lua_helpers.go

 M8: Per-Table Hook Approval -- Plugin Permission Model

 Problem: Without an authorization boundary, any plugin registering a wildcard hook on "*" receives the full serialized entity data for every mutation across all 26+ audited tables, including
 sensitive tables (users, sessions, ssh_keys). This bypasses the namespace isolation enforced by prefixTable() for db.* operations.

 Solution: Hooks use a per-table per-plugin approval model (same pattern as Phase 2 HTTP route approval). During on_init, hooks.on() calls are recorded as pending registrations. The HookEngine
 only dispatches to hooks whose (plugin, event, table) tuple has been approved by a developer. Unapproved hooks are silently skipped at dispatch time (not at registration time, so the plugin
 loads without error but hooks do not fire until approved).

 S1: Approval storage uses a plugin_hooks table (same pattern as plugin_routes from Phase 2). The HookEngine owns this table. Tri-dialect DDL:

 -- SQLite
 CREATE TABLE IF NOT EXISTS plugin_hooks (
     plugin_name    TEXT    NOT NULL,
     event          TEXT    NOT NULL,
     table_name     TEXT    NOT NULL,
     approved       INTEGER NOT NULL DEFAULT 0,
     approved_at    DATETIME,
     approved_by    TEXT,
     plugin_version TEXT    NOT NULL DEFAULT '',
     PRIMARY KEY (plugin_name, event, table_name)
 );

 -- MySQL
 CREATE TABLE IF NOT EXISTS plugin_hooks (
     plugin_name    VARCHAR(255) NOT NULL,
     event          VARCHAR(64)  NOT NULL,
     table_name     VARCHAR(255) NOT NULL,
     approved       TINYINT(1)   NOT NULL DEFAULT 0,
     approved_at    TIMESTAMP    NULL DEFAULT NULL,
     approved_by    VARCHAR(255),
     plugin_version VARCHAR(255) NOT NULL DEFAULT '',
     PRIMARY KEY (plugin_name, event, table_name)
 );

 -- PostgreSQL
 CREATE TABLE IF NOT EXISTS plugin_hooks (
     plugin_name    TEXT    NOT NULL,
     event          TEXT    NOT NULL,
     table_name     TEXT    NOT NULL,
     approved       BOOLEAN NOT NULL DEFAULT FALSE,
     approved_at    TIMESTAMPTZ,
     approved_by    TEXT,
     plugin_version TEXT    NOT NULL DEFAULT '',
     PRIMARY KEY (plugin_name, event, table_name)
 );

 The table is created in HookEngine.CreatePluginHooksTable(ctx) called from Manager.LoadAll(), same as HTTPBridge.CreatePluginRoutesTable(). The method switches DDL by dialect (same
 pattern as CreatePluginRoutesTable). Upsert logic follows the plugin_routes pattern: if a plugin version changes and the (event, table_name) pair changes, approved resets to 0.

 Approval/revocation methods: HookEngine.ApproveHook(ctx, plugin, event, table, approvedBy) and HookEngine.RevokeHook(ctx, plugin, event, table) update the DB and the in-memory dispatch map.
 These are wired to admin API endpoints in Agent F (same pattern as HTTPBridge.ApproveRoute/RevokeRoute).

 Wildcard ("*") hooks require explicit wildcard approval. Approving specific tables does NOT approve wildcard. Approving wildcard does NOT auto-approve specific tables. They are independent
 permission entries.

 Re-approval on change: If a plugin's init.lua registers different hooks than what was previously approved (new tables, new events), the new registrations require fresh approval. Existing
 approved hooks continue to fire. This prevents a plugin update from silently expanding its table access.

 Files: internal/plugin/hook_engine.go, internal/plugin/manager.go, internal/plugin/hooks_api.go

 M9: Before-Hook Circuit Breaker

 Problem: A malicious or buggy plugin can register before-hooks (including wildcards) that call error() on every invocation, blocking all CMS mutations indefinitely. The only recovery is
 removing the plugin from the filesystem and restarting.

 Solution: HookEngine tracks a consecutive-abort counter per plugin per (event, table) pair. After N consecutive aborts (configurable via plugin_hook_max_consecutive_aborts, default 10), the
 hook is auto-disabled and a slog.Error is emitted with the plugin name, event, table, and abort count. The hook remains disabled until the server restarts or an admin re-enables it.

 Admin runtime control: Add an internal method HookEngine.SetHookEnabled(plugin, event, table, enabled bool) that can be wired to an admin API endpoint in a future phase. For now, restart
 clears the disabled state (hook registrations are rebuilt from init.lua on load).

 A successful hook execution (no error) resets the counter to zero.

 Files: internal/plugin/hook_engine.go, internal/config/config.go

 M10: After-Hook Resource Budgets

 Problem: After-hooks are fire-and-forget goroutines with db.* access. Without bounded concurrency or reduced operation budgets, a burst of mutations can spawn unlimited after-hook goroutines
 that exhaust the plugin DB pool or the database itself.

 Solution: Two controls:

 1. Reduced op budget: ResetOpCount() is called before each after-hook execution with a reduced budget. New config field plugin_hook_max_ops (default 100, vs 1000 for HTTP requests).
 Signature: ResetOpCount(limitOverride ...int). Zero args uses api.maxOpsPerExec (the existing default). One arg uses the provided limit. Called as ResetOpCount(cfg.Plugin_Hook_Max_Ops)
 before after-hook execution.

 2. Global concurrency limiter: HookEngine maintains a semaphore (buffered channel) of size plugin_hook_max_concurrent_after (default 10). After-hook goroutines acquire the semaphore before
 VM checkout and release on completion. If the semaphore is full, the goroutine blocks (with shutdown-aware context) rather than spawning unboundedly.

 // hook_engine.go
 type HookEngine struct {
     // ...
     afterSem chan struct{} // buffered to maxConcurrentAfter
 }

 Files: internal/plugin/hook_engine.go, internal/plugin/db_api.go, internal/config/config.go

 M11: Before-Hook Execution Timeout

 Problem: Before-hooks share the 30-second transaction timeout. A hung hook holds the transaction (and SQLite's file-level lock) for up to 30 seconds, blocking all other writes.

 Solution: Wrap before-hook Lua execution in a dedicated context.WithTimeout. New config field plugin_hook_timeout_ms (default 2000). This is deliberately shorter than the transaction timeout
 to leave headroom for the actual mutation and audit record. The VM's LState context is set to this derived context before execution.

 // hook_engine.go -- executeBefore:
 hookCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Plugin_Hook_Timeout_Ms)*time.Millisecond)
 defer cancel()
 L.SetContext(hookCtx)

 After-hooks use the existing execTimeout (5s default) since they are not holding a transaction.

 S2: Per-event timeout budget. The plugin_hook_timeout_ms applies PER INDIVIDUAL HOOK, not per event. If 3 plugins each register a before_create hook on content_data, the worst case is
 3 * 2s = 6s inside the transaction. To cap total before-hook time per event, add plugin_hook_event_timeout_ms (default 5000). RunBeforeHooks creates a single context.WithTimeout for the
 entire hook chain. Individual hooks inherit this parent context (and their own per-hook sub-timeout, whichever is shorter). This means: single hook = 2s max, three hooks = 5s max total,
 not 6s. The event-level timeout must be shorter than the 30s transaction timeout and should be documented as the controlling limit for multi-hook scenarios.

 Files: internal/plugin/hook_engine.go, internal/config/config.go

 M12: detectStatusTransition Specification

 Problem: The status transition detection logic determines whether before_publish and before_archive hooks fire. Incorrect detection could miss validation (false negative) or cause spurious
 aborts (false positive).

 Solution: Fully specified detection rules:

 Applicable tables: Only "content_data". All other tables skip detection entirely.

 Field compared: "status" (case-sensitive string comparison).

 Transition rules:
 - before.Status != "published" AND params contains status field AND params.Status == "published" -> emit before_publish / after_publish
 - before.Status != "archived" AND params contains status field AND params.Status == "archived" -> emit before_archive / after_archive
 - If before entity is nil (should not happen for Update, but defensively): skip detection, emit no extra events.
 - If params does not contain a status field: skip detection (partial update not touching status).
 - Unknown status values: no extra events. Only the exact strings "published" and "archived" trigger detection.

 Implementation: detectStatusTransition receives the before entity map and params map (both map[string]any after structToMap). Returns a []HookEvent slice (0, 1, or 2 entries).

 Files: internal/db/audited/hooks.go (or internal/plugin/hook_engine.go, implementer's choice)

 ---
 Structural Fixes (from architectural stress-test)

 S4: Hook Observability

 Problem: Hooks fundamentally change the request latency profile (before-hooks add up to 5s inside transactions) and the error surface (hooks can abort mutations). Without metrics, the
 cause of write latency regressions and unexpected 422s will be invisible.

 Solution: HookEngine emits structured slog events at key points. The existing observability config (Sentry/Datadog/NewRelic) already captures slog output.

 Metrics emitted:
 - hook.before.duration_ms: Per-hook execution time (plugin, event, table, status=ok|error|timeout)
 - hook.before.total_duration_ms: Total before-hook chain time per event (event, table, hook_count)
 - hook.before.abort: Counter for before-hook aborts (plugin, event, table, consecutive_count)
 - hook.after.duration_ms: Per-hook execution time (plugin, event, table, status=ok|error|timeout)
 - hook.after.queued: Gauge of goroutines waiting on the afterSem semaphore
 - hook.circuit_breaker.tripped: Event when a hook is auto-disabled (plugin, event, table, abort_count)

 All metrics use slog.With() for structured fields. No new dependencies. The HookEngine accepts a *slog.Logger at construction.

 For the hot path (HasHooks returning false), zero logging occurs. Metrics are only emitted when hooks actually execute.

 Files: internal/plugin/hook_engine.go

 S5: dbAPIs Map Leak Cleanup

 Problem: When Put() detects an unhealthy VM and replaces it via the factory, the old *lua.LState key remains in inst.dbAPIs forever. Phase 3 amplifies this because hooks increase VM
 checkout frequency, increasing the probability of unhealthy detection and replacement.

 Solution: Put() must delete the old key from inst.dbAPIs before calling factory. The PluginInstance needs a method RemoveDBAPI(L *lua.LState) called by the pool's replacement path.

 // pool.go -- Put, in the unhealthy replacement branch:
 if p.onReplace != nil {
     p.onReplace(L) // callback to clean up inst.dbAPIs[L]
 }

 VMPool gains an onReplace func(*lua.LState) field set during construction. PluginInstance must gain a sync.Mutex field (inst.mu) to protect dbAPIs map
 access from concurrent onReplace callbacks. Manager.loadPlugin sets onReplace to:
 func(old *lua.LState) {
     inst.mu.Lock()
     delete(inst.dbAPIs, old)
     inst.mu.Unlock()
     old.Close()
 }

 This also properly closes the old LState, which the current code does via L.Close() inline but does not clean up the map entry.

 Files: internal/plugin/pool.go, internal/plugin/manager.go

 ---
 HookRunner Interface (revised)

 internal/db/audited/hooks.go:

 type HookEvent string

 const (
     HookBeforeCreate  HookEvent = "before_create"
     HookAfterCreate   HookEvent = "after_create"
     HookBeforeUpdate  HookEvent = "before_update"
     HookAfterUpdate   HookEvent = "after_update"
     HookBeforeDelete  HookEvent = "before_delete"
     HookAfterDelete   HookEvent = "after_delete"
     HookBeforePublish HookEvent = "before_publish"
     HookAfterPublish  HookEvent = "after_publish"
     HookBeforeArchive HookEvent = "before_archive"
     HookAfterArchive  HookEvent = "after_archive"
 )

 type HookRunner interface {
     HasHooks(event HookEvent, table string) bool
     RunBeforeHooks(ctx context.Context, event HookEvent, table string, entity any) error
     RunAfterHooks(ctx context.Context, event HookEvent, table string, entity any)
 }

 Key change from original: RunBeforeHooks accepts entity any (raw struct), not map[string]any. The structToMap JSON roundtrip happens inside the implementation only when hooks actually
 match. HasHooks enables the fast-path gate in audited.go.

 ---
 Modified Audited Functions (revised)

 Create

 1. If HookRunner != nil && HasHooks(before_create, table):
    a. RunBeforeHooks(ctx, before_create, table, cmd.Params()) -- inside tx
    b. Error -> return, tx rolls back
 2. cmd.Execute (existing)
 3. Record change event (existing)
 4. Capture afterEntity from created result
 --- tx commits ---
 5. If HookRunner != nil && HasHooks(after_create, table): RunAfterHooks(ctx, after_create, table, afterEntity)

 S3: After-hook HasHooks gate. The HasHooks check for after events is critical: without it, structToMap runs inside RunAfterHooks on every mutation even when no after-hooks exist. The
 HasHooks check for afterEntity also determines whether afterEntity needs to be captured at all -- if no after-hooks exist for the table, skip the capture and avoid the allocation.

 Structural change: Capture tx result and fire after-hooks outside the transaction (same pattern as Update/Delete in original plan).

 Update (with publish/archive detection)

 hasUpdateHooks = HookRunner != nil && (HasHooks(before_update, table) || HasHooks(before_publish, table) || HasHooks(before_archive, table))

 1. cmd.GetBefore (existing)
 2. If hasUpdateHooks:
    a. If HasHooks(before_update, table): RunBeforeHooks(before_update, table, cmd.Params()) -- inside tx
    b. detectStatusTransition(table, before, params) -> extra events ([]HookEvent)
    c. For each extra event with HasHooks: RunBeforeHooks(before_publish/before_archive, table, cmd.Params())
    d. Error on any -> return, tx rolls back
 3. cmd.Execute (existing)
 4. Record change event (existing)
 --- tx commits ---
 5. If HasHooks(after_update, table): RunAfterHooks(after_update, table, cmd.Params())
    Then for each extra event: if HasHooks(after_publish/after_archive, table): RunAfterHooks(event, table, cmd.Params())

 Note: The union gate in step 2 ensures publish/archive detection runs even when no before_update hooks exist but before_publish or before_archive hooks do.

 Hook data shape for all Update events (before_update, before_publish, before_archive, and their after counterparts): the Lua handler receives cmd.Params() -- the update params struct, NOT
 the full entity. This matches Create (which receives create params) and Delete (which receives the before entity). The before entity from cmd.GetBefore() is NOT passed to hooks; plugins
 that need the before state should use after-hooks where db.* is available.

 Delete

 1. cmd.GetBefore (existing)
 2. If HookRunner != nil && HasHooks(before_delete, table):
    a. RunBeforeHooks(before_delete, table, beforeEntity)
    b. Error -> return, tx rolls back
 3. cmd.Execute (existing)
 4. Record change event (existing)
 --- tx commits ---
 5. If HookRunner != nil && HasHooks(after_delete, table): RunAfterHooks(after_delete, table, beforeEntity)

 ---
 hooks Lua Module

 `hooks` is a new global Lua table with a single function: `on`. It follows the same patterns as `db`, `log`, and `http`:
 - Registered via RegisterHooksAPI(L, inst) in hooks_api.go (Agent B)
 - Frozen via FreezeModule(L, "hooks") in the factory sequence (Agent F)
 - Validated by validateVM in sandbox.go (Agent F)
 - Called at MODULE SCOPE ONLY (same phase guard as http.handle()). Blocked during on_init because each VM needs its own registered Lua function reference.

 Factory call sequence (updated -- new steps marked with +):
 1. ApplySandbox
 2. RegisterPluginRequire
 3. RegisterDBAPI
 4. RegisterLogAPI
 5. RegisterHTTPAPI
 + 6. RegisterHooksAPI
 7. FreezeModule("db")
 8. FreezeModule("log")
 9. FreezeModule("http")
 + 10. FreezeModule("hooks")

 hooks.on(event, table, handler, opts)

 Parameters:
 - event: string, one of the HookEvent constants ("before_create", "after_create", etc.)
 - table: string, the audited table name or "*" for wildcard
 - handler: function(data) -- the Lua callback
 - opts: optional table with { priority = number }

 Priority: integer, default 100, valid range 1-1000. Lower numbers execute first. Values outside 1-1000 are clamped.
 Maximum 50 hooks per plugin (total across all events and tables). Exceeding the limit returns an error from hooks.on().

 Data injection: Before dispatching to the Lua handler, the HookEngine injects two metadata fields into the data map:
 - _table: string, the audited table name (e.g., "content_data")
 - _event: string, the hook event (e.g., "before_create")
 These use underscore-prefixed names to avoid collision with CMS entity fields (no entity field starts with underscore).

 ---
 Lua API Examples

 hooks.on("before_create", "content_data", function(data)
     -- data = {field1 = val1, ..., _table = "content_data", _event = "before_create"}
     -- db.* calls are NOT available in before-hooks (deadlock prevention)
     -- Return nil = no modification
     -- Call error("msg") = abort (rolls back transaction, msg is sanitized)
     if data.title == "" then
         error("title is required")
     end
 end)

 hooks.on("after_update", "*", function(data)
     -- Wildcard: fires for ALL approved tables (after specific hooks at equal priority)
     -- db.* calls ARE available in after-hooks
     log.info("entity updated in " .. data._table)
 end)

 hooks.on("before_publish", "content_data", function(data)
     -- Fires when content_data.status transitions to "published"
 end, {priority = 50})

 Documented constraints:
 - Before-hooks: db.* calls are blocked (returns error). Use data argument for inspection.
 - After-hooks: fire-and-forget. No delivery or ordering guarantee relative to HTTP response.
 - After-hook goroutines use HookEngine's shutdown context (derived from a cancellable context.Context stored on HookEngine, cancelled during Close()). NOT the request context, which
   may already be done since the HTTP response has been sent.
 - All Go numeric fields become Lua numbers (float64). IDs are ULID strings, unaffected.
 - Specific hooks run before wildcard hooks at equal priority. Lower priority number = runs first.
 - Maximum 50 hooks per plugin total (across all events and tables). A plugin with 10 tables * 6 events = 60 hooks exceeds this; use wildcards to stay within budget.
 - Hooks require per-table developer approval before they fire. Unapproved hooks are silently skipped.
 - Before-hooks have a 2-second per-hook timeout and 5-second per-event total timeout (both configurable). Exceeding either aborts and rolls back.
 - After-hooks run with a reduced 100-op budget (vs 1000 for HTTP handlers) and bounded concurrency (max 10 concurrent).
 - Hook execution emits structured slog metrics (duration, abort count, circuit breaker trips). No logging when no hooks match.

 ---
 New Files

 File: internal/db/audited/hooks.go
 Size: S
 Purpose: HookRunner interface, HookEvent constants, HookError type, structToMap
 ────────────────────────────────────────
 File: internal/plugin/hooks_api.go
 Size: M
 Purpose: RegisterHooksAPI, hooksOnFn, event/table validation, phase guard
 ────────────────────────────────────────
 File: internal/plugin/hook_engine.go
 Size: XL
 Purpose: HookEngine struct, RegisterHooks, RunBeforeHooks/RunAfterHooks, executeBefore/After, HasHooks, fast-path maps, afterWG shutdown, hookEntryLess ordering
 ────────────────────────────────────────
 File: internal/plugin/hooks_api_test.go
 Size: L
 Purpose: Unit tests for hooks.on() registration and validation
 ────────────────────────────────────────
 File: internal/plugin/hook_engine_test.go
 Size: XL
 Purpose: Before/after execution, priority, timeout, publish detection, fast-path, pool reservation
 ────────────────────────────────────────
 File: internal/plugin/hook_integration_test.go
 Size: L
 Purpose: End-to-end: load plugin with hooks, trigger audited ops
 ────────────────────────────────────────
 File: internal/plugin/testdata/plugins/hooks_plugin/init.lua
 Size: S
 Purpose: Basic before/after create/update/delete hooks
 ────────────────────────────────────────
 File: internal/plugin/testdata/plugins/hooks_publish_plugin/init.lua
 Size: S
 Purpose: Publish/archive detection hooks
 ────────────────────────────────────────
 File: internal/plugin/testdata/plugins/hooks_priority_plugin/init.lua
 Size: S
 Purpose: Priority and wildcard ordering test
 ────────────────────────────────────────
 File: internal/plugin/testdata/plugins/hooks_abort_plugin/init.lua
 Size: S
 Purpose: Before-hook that calls error()
 ────────────────────────────────────────
 File: internal/plugin/testdata/plugins/hooks_wildcard_plugin/init.lua
 Size: S
 Purpose: Wildcard "*" table hooks
 ────────────────────────────────────────
 File: internal/plugin/testdata/plugins/hooks_db_blocked_plugin/init.lua
 Size: S
 Purpose: Verifies db.* blocked in before-hooks

 Modified Files

 ┌────────────────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │                  File                  │                                                 Changes                                                  │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/audited/context.go         │ Add HookRunner field to AuditContext                                                                     │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/audited/audited.go         │ Hook invocation in Create/Update/Delete with HasHooks fast-path; restructure for post-commit after-hooks │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/plugin/pool.go                │ Split into general/reserved channels; add GetForHook; update Put to return to either channel             │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/plugin/pool_test.go           │ Update tests for dual-channel pool, add GetForHook tests                                                 │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/plugin/manager.go             │ Add hookEngine field, create in LoadAll, call RegisterHooks in loadPlugin, expose HookEngine() getter    │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/plugin/sandbox.go             │ Add hooks to validateVM, FreezeModule for hooks                                                          │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/plugin/db_api.go              │ Add inBeforeHook field, gate in checkOpLimit                                                             │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/plugin/lua_helpers.go         │ Add json.Number case to GoValueToLua                                                                     │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/middleware/audit_helpers.go   │ Extract HookRunner from request context                                                                  │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/middleware/http_middleware.go │ No signature change -- HookRunner injected as separate middleware in serve.go                            │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/config/config.go              │ Add Plugin_Hook_Reserve_VMs (1), Plugin_Hook_Max_Consecutive_Aborts (10), Plugin_Hook_Max_Ops (100),     │
 │                                        │ Plugin_Hook_Max_Concurrent_After (10), Plugin_Hook_Timeout_Ms (2000), Plugin_Hook_Event_Timeout_Ms (5000)  │
 ├────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ cmd/serve.go                           │ Wire HookEngine into middleware (context injection), add hookEngine.Close in shutdown sequence           │
 └────────────────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 Key Existing Code to Reuse

 - internal/plugin/lua_helpers.go -- MapToLuaTable, GoValueToLua, LuaTableToMap
 - internal/plugin/http_api.go -- Module-scope registration with phase guard and hidden globals pattern
 - internal/plugin/http_bridge.go -- VM checkout, timeout, op count reset pattern
 - internal/db/audited/audited.go -- Create/Update/Delete to modify
 - internal/middleware/audit_helpers.go -- AuditContextFromRequest to extend
 - internal/middleware/http_middleware.go -- Context key pattern (unexported struct type)

 ---
 Multi-Agent Decomposition

 Parallel Group 1 (no dependencies):

 - Agent A -- internal/db/audited/hooks.go + audited/context.go modification
 HookRunner interface (with HasHooks), HookEvent constants, HookError type, structToMap (with json.Number), detectStatusTransition. Add HookRunner field to AuditContext. Unit tests.
 - Agent B -- internal/plugin/hooks_api.go + hooks_api_test.go
 RegisterHooksAPI, hooksOnFn, event/table validation, phase guard (module scope only, blocked during on_init -- same as http.handle()), priority validation (1-1000, default 100),
 count limit (max 50 per plugin total). Records hook registrations as pending -- does NOT filter by approval at registration time. Follows http_api.go pattern.
 Tests for all validation paths including phase guard, priority clamping, and count limit.
 - Agent C -- internal/plugin/pool.go refactor + pool_test.go
 Split VMPool into general/reserved channels. Add GetForHook. Update Put to return to either channel. Update NewVMPool to accept reserveSize. Add onReplace callback for dbAPIs map
 cleanup (S5): Put() calls onReplace(oldL) before factory replacement, so the old *lua.LState key is deleted from inst.dbAPIs. Update all existing tests.

 Parallel Group 2 (depends on Group 1):

 - Agent D -- internal/plugin/hook_engine.go + hook_engine_test.go
 HookEngine struct with hasHooks map, hookIndex, afterWG, closing flag, afterSem, consecutiveAborts map, shutdownCtx/cancel. RegisterHooks, RunBeforeHooks/RunAfterHooks,
 executeBefore/After (fully implemented here -- Agent E does NOT implement these), HasHooks, hookEntryLess, Close, SetHookEnabled. Per-table approval check at dispatch time (M8):
 only approved hooks are dispatched, unapproved are silently skipped. Circuit breaker (M9), after-hook semaphore (M10), before-hook timeout with per-event budget (M11/S2),
 structured slog observability (S4). Uses HookRunner from A, registration from B, GetForHook from C.

 executeBefore implementation steps (binding for Agent D):
 (1) Check out VM via pool.GetForHook(ctx)
 (2) Look up inst.dbAPIs[L] for the bound DatabaseAPI
 (3) Set dbAPI.inBeforeHook = true
 (4) Create per-hook context.WithTimeout(eventCtx, hookTimeout)
 (5) Set L.SetContext(hookCtx)
 (6) Call the registered Lua handler
 (7) Defer: clear inBeforeHook, cancel hookCtx, return VM via pool.Put(L)

 executeAfter implementation steps (binding for Agent D):
 (1) Check HookEngine.closing; if true, skip
 (2) Acquire afterSem (blocking, with shutdownCtx)
 (3) afterWG.Add(1)
 (4) In goroutine: check out VM via pool.GetForHook(shutdownCtx), look up dbAPI, call ResetOpCount(hookMaxOps),
     execute handler with execTimeout, return VM, release afterSem, afterWG.Done()

 Sequential (depends on all above):

 - Agent E -- internal/plugin/db_api.go + internal/db/audited/audited.go + internal/plugin/lua_helpers.go modifications
 Scope (Agent E does NOT implement executeBefore/executeAfter -- those are fully owned by Agent D):
 (1) db_api.go: Add inBeforeHook bool field. Add gate in checkOpLimit (M1). Change ResetOpCount signature to ResetOpCount(limitOverride ...int) (M10).
 (2) audited.go: Add hook invocation points to Create/Update/Delete with HasHooks fast-path gates. Restructure Create to capture afterEntity inside tx and fire after-hooks
     post-commit. Restructure Update with union HasHooks gate for publish/archive detection (see "Modified Audited Functions" section).
 (3) lua_helpers.go: Add json.Number case to GoValueToLua.
 - Agent F -- Wiring + integration
 manager.go: HookEngine creation, loadPlugin integration (add RegisterHooksAPI + FreezeModule("hooks") to factory sequence per "hooks Lua Module" section), onReplace callback
 wiring for S5, add sync.Mutex field to PluginInstance for dbAPIs protection (S5). sandbox.go: add "hooks" to validateVM. audit_helpers.go: extract HookRunner from request context.
 config.go: add Plugin_Hook_Event_Timeout_Ms (5000). cmd/serve.go: middleware wiring (HookRunner context injection), shutdown sequence.
 HookEngine.CreatePluginHooksTable in LoadAll (S1). ApproveHook/RevokeHook admin API endpoints (same pattern as HTTPBridge.ApproveRoute/RevokeRoute).
 All test fixtures. hook_integration_test.go.

 ---
 Verification

 # Unit tests for audited hook types
 go test -v ./internal/db/audited/ -run TestHook -count=1

 # Unit tests for hooks API registration
 go test -v ./internal/plugin/ -run TestHooksAPI -count=1

 # VM pool reservation tests
 go test -v ./internal/plugin/ -run TestVMPool -count=1

 # Unit tests for hook engine
 go test -v ./internal/plugin/ -run TestHookEngine -count=1

 # Integration tests
 go test -v ./internal/plugin/ -run TestHookIntegration -count=1

 # Full plugin package (Phase 1+2+3)
 go test -v ./internal/plugin/ -count=1

 # Audited package (no regressions)
 go test -v ./internal/db/audited/ -count=1

 # Middleware (audit_helpers changes)
 go test -v ./internal/middleware/ -count=1

 # Full regression
 go test ./... -count=1

 # Compile check
 just check
