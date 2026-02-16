package plugin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/utility"
	lua "github.com/yuin/gopher-lua"
)

// hookEntry is a single registered hook with its dispatch metadata.
type hookEntry struct {
	pluginName string
	event      audited.HookEvent
	table      string // audited table name or "*"
	priority   int
	isWildcard bool
	regOrder   int    // registration order for stable sort
	handlerKey string // key into __hook_handlers table
}

// HookEngineConfig configures the hook engine.
type HookEngineConfig struct {
	HookTimeoutMs        int // per-hook timeout in before-hooks (default 2000)
	EventTimeoutMs       int // per-event total timeout for before-hook chain (default 5000)
	MaxConsecutiveAborts int // circuit breaker threshold (default 10)
	MaxConcurrentAfter   int // max concurrent after-hook goroutines (default 10)
	HookMaxOps           int // reduced op budget for after-hooks (default 100)
	ExecTimeoutMs        int // after-hook execution timeout (default 5000)
}

// HookEngine orchestrates content lifecycle hook dispatch. It implements the
// audited.HookRunner interface and is the central coordinator between the
// audited command layer and plugin VMs.
//
// Thread safety: hookIndex and hasHooks are built once during RegisterHooks
// (called from loadPlugin, which runs sequentially under Manager.mu). After
// LoadAll completes, these are read-only. The approval and circuit breaker
// maps are protected by mu.
type HookEngine struct {
	// hookIndex maps "event:table" to sorted hookEntry slices.
	// Built once during registration, read concurrently during dispatch.
	hookIndex map[string][]hookEntry

	// hasHooks is the fast-path gate. Zero-allocation O(1) check.
	hasHooks map[string]bool

	// hasAnyHook is true if any hooks are registered across all plugins.
	hasAnyHook bool

	// manager provides access to plugin instances for VM checkout.
	manager *Manager

	// mu protects approval and circuit breaker state.
	mu sync.RWMutex

	// approved tracks per (plugin, event, table) approval status.
	// Key format: "plugin:event:table"
	approved map[string]bool

	// disabled tracks per (plugin, event, table) circuit breaker state.
	// Key format: "plugin:event:table"
	disabled map[string]bool

	// consecutiveAborts tracks abort count per (plugin, event, table).
	// Key format: "plugin:event:table"
	consecutiveAborts map[string]int

	// afterWG tracks in-flight after-hook goroutines for shutdown drain.
	afterWG sync.WaitGroup

	// closing is set during Close() to prevent new after-hook dispatches.
	closing atomic.Bool

	// shutdownCtx is used by after-hook goroutines instead of the request
	// context (which may already be done since the HTTP response was sent).
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc

	// afterSem limits concurrent after-hook goroutines (M10).
	afterSem chan struct{}

	cfg HookEngineConfig

	// dbPool and dialect for plugin_hooks table operations.
	dbPool  *sql.DB
	dialect db.Dialect

	logger *utility.Logger
}

// NewHookEngine creates a new HookEngine.
func NewHookEngine(manager *Manager, dbPool *sql.DB, dialect db.Dialect, cfg HookEngineConfig) *HookEngine {
	if cfg.HookTimeoutMs <= 0 {
		cfg.HookTimeoutMs = 2000
	}
	if cfg.EventTimeoutMs <= 0 {
		cfg.EventTimeoutMs = 5000
	}
	if cfg.MaxConsecutiveAborts <= 0 {
		cfg.MaxConsecutiveAborts = 10
	}
	if cfg.MaxConcurrentAfter <= 0 {
		cfg.MaxConcurrentAfter = 10
	}
	if cfg.HookMaxOps <= 0 {
		cfg.HookMaxOps = 100
	}
	if cfg.ExecTimeoutMs <= 0 {
		cfg.ExecTimeoutMs = 5000
	}

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	return &HookEngine{
		hookIndex:         make(map[string][]hookEntry),
		hasHooks:          make(map[string]bool),
		manager:           manager,
		approved:          make(map[string]bool),
		disabled:          make(map[string]bool),
		consecutiveAborts: make(map[string]int),
		shutdownCtx:       shutdownCtx,
		shutdownCancel:    shutdownCancel,
		afterSem:          make(chan struct{}, cfg.MaxConcurrentAfter),
		cfg:               cfg,
		dbPool:            dbPool,
		dialect:           dialect,
		logger:            utility.DefaultLogger,
	}
}

// RegisterHooks processes pending hook registrations from a loaded plugin.
// Called once per plugin during loadPlugin. Must be called before the hasHooks
// map is used for dispatch (i.e., before LoadAll returns).
func (e *HookEngine) RegisterHooks(pluginName string, hooks []PendingHook) {
	for i, h := range hooks {
		entry := hookEntry{
			pluginName: pluginName,
			event:      h.Event,
			table:      h.Table,
			priority:   h.Priority,
			isWildcard: h.IsWildcard,
			regOrder:   i,
			handlerKey: h.HandlerKey,
		}

		// Index by "event:table" for specific hooks.
		key := string(h.Event) + ":" + h.Table
		e.hookIndex[key] = append(e.hookIndex[key], entry)
		e.hasHooks[key] = true

		// Also mark "event:*" for wildcard-aware HasHooks.
		if h.IsWildcard {
			e.hasHooks[string(h.Event)+":*"] = true
		}

		e.hasAnyHook = true
	}

	// Sort all entries by priority, wildcard status, and registration order.
	for key := range e.hookIndex {
		entries := e.hookIndex[key]
		sort.SliceStable(entries, func(i, j int) bool {
			return hookEntryLess(entries[i], entries[j])
		})
		e.hookIndex[key] = entries
	}
}

// hookEntryLess defines the ordering: lower priority first, specific before
// wildcard at equal priority, registration order as tiebreaker (M6).
func hookEntryLess(a, b hookEntry) bool {
	if a.priority != b.priority {
		return a.priority < b.priority
	}
	if a.isWildcard != b.isWildcard {
		return !a.isWildcard
	}
	return a.regOrder < b.regOrder
}

// HasHooks implements audited.HookRunner. Returns true if any hooks are
// registered for the given event and table (including wildcards).
// Zero allocation, O(1) map lookup.
func (e *HookEngine) HasHooks(event audited.HookEvent, table string) bool {
	if !e.hasAnyHook {
		return false
	}
	if e.hasHooks[string(event)+":"+table] {
		return true
	}
	return e.hasHooks[string(event)+":*"]
}

// RunBeforeHooks implements audited.HookRunner. Executes all matching hooks
// for the given event/table synchronously inside the caller's transaction
// context. Returns a *audited.HookError if any hook aborts.
//
// The entity argument is the raw Go struct; StructToMap is called here (not at
// the call site) so the JSON roundtrip only happens when hooks exist.
func (e *HookEngine) RunBeforeHooks(ctx context.Context, event audited.HookEvent, table string, entity any) error {
	entries := e.gatherEntries(event, table)
	if len(entries) == 0 {
		return nil
	}

	// Convert entity to map once for all hooks.
	dataMap, err := audited.StructToMap(entity)
	if err != nil {
		return fmt.Errorf("hook structToMap: %w", err)
	}

	// Inject metadata fields.
	if dataMap == nil {
		dataMap = make(map[string]any)
	}
	dataMap["_table"] = table
	dataMap["_event"] = string(event)

	// S2: Per-event timeout budget.
	eventCtx, eventCancel := context.WithTimeout(ctx, time.Duration(e.cfg.EventTimeoutMs)*time.Millisecond)
	defer eventCancel()

	start := time.Now()

	for _, entry := range entries {
		if err := e.executeBefore(eventCtx, entry, dataMap); err != nil {
			return err
		}
	}

	// S4: Observability -- total before-hook chain duration.
	totalMs := time.Since(start).Milliseconds()
	e.logger.Debug("hook.before.total_duration_ms",
		"event", string(event),
		"table", table,
		"hook_count", len(entries),
		"duration_ms", totalMs,
	)

	return nil
}

// RunAfterHooks implements audited.HookRunner. Dispatches matching hooks
// asynchronously (fire-and-forget). Uses the engine's shutdown context,
// not the request context.
func (e *HookEngine) RunAfterHooks(ctx context.Context, event audited.HookEvent, table string, entity any) {
	if e.closing.Load() {
		return
	}

	entries := e.gatherEntries(event, table)
	if len(entries) == 0 {
		return
	}

	// Convert entity to map once for all hooks.
	dataMap, err := audited.StructToMap(entity)
	if err != nil {
		e.logger.Error(
			fmt.Sprintf("hook after structToMap error: %s", err.Error()), nil,
			"event", string(event),
			"table", table,
		)
		return
	}

	if dataMap == nil {
		dataMap = make(map[string]any)
	}
	dataMap["_table"] = table
	dataMap["_event"] = string(event)

	for _, entry := range entries {
		entry := entry // capture for goroutine

		// M10: Acquire semaphore (bounded concurrency).
		select {
		case e.afterSem <- struct{}{}:
			// Acquired.
		case <-e.shutdownCtx.Done():
			return
		}

		e.afterWG.Add(1)
		go func() {
			defer e.afterWG.Done()
			defer func() { <-e.afterSem }() // release semaphore

			e.executeAfter(entry, dataMap)
		}()
	}
}

// gatherEntries collects and merges specific + wildcard hooks for the given
// event and table, filtering by approval and circuit breaker state.
func (e *HookEngine) gatherEntries(event audited.HookEvent, table string) []hookEntry {
	specificKey := string(event) + ":" + table
	wildcardKey := string(event) + ":*"

	specific := e.hookIndex[specificKey]
	wildcard := e.hookIndex[wildcardKey]

	if len(specific) == 0 && len(wildcard) == 0 {
		return nil
	}

	// Merge and filter.
	merged := make([]hookEntry, 0, len(specific)+len(wildcard))

	e.mu.RLock()
	for _, entry := range specific {
		approvalKey := approvalKeyFor(entry.pluginName, string(event), table)
		if !e.approved[approvalKey] {
			continue // M8: unapproved hooks silently skipped
		}
		if e.disabled[approvalKey] {
			continue // M9: circuit-breaker disabled
		}
		merged = append(merged, entry)
	}
	for _, entry := range wildcard {
		approvalKey := approvalKeyFor(entry.pluginName, string(event), "*")
		if !e.approved[approvalKey] {
			continue
		}
		disableKey := approvalKeyFor(entry.pluginName, string(event), table)
		if e.disabled[disableKey] {
			continue
		}
		merged = append(merged, entry)
	}
	e.mu.RUnlock()

	if len(merged) == 0 {
		return nil
	}

	// Re-sort merged list.
	sort.SliceStable(merged, func(i, j int) bool {
		return hookEntryLess(merged[i], merged[j])
	})

	return merged
}

// executeBefore runs a single before-hook synchronously.
func (e *HookEngine) executeBefore(eventCtx context.Context, entry hookEntry, dataMap map[string]any) error {
	inst := e.manager.GetPlugin(entry.pluginName)
	if inst == nil || inst.State != StateRunning {
		return nil
	}

	// Check out VM for hook execution.
	L, getErr := inst.Pool.GetForHook(eventCtx)
	if getErr != nil {
		if errors.Is(getErr, ErrPoolExhausted) {
			e.logger.Warn(
				fmt.Sprintf("hook before: VM pool exhausted for plugin %q", entry.pluginName),
				nil,
				"event", string(entry.event),
				"table", entry.table,
			)
			return fmt.Errorf("hook VM pool exhausted for plugin %q", entry.pluginName)
		}
		return fmt.Errorf("hook VM checkout: %w", getErr)
	}
	defer inst.Pool.Put(L)

	// Look up the bound DatabaseAPI and set inBeforeHook flag (M1).
	inst.mu.Lock()
	dbAPI := inst.dbAPIs[L]
	inst.mu.Unlock()

	if dbAPI != nil {
		dbAPI.inBeforeHook = true
		defer func() { dbAPI.inBeforeHook = false }()
	}

	// M11: Per-hook timeout (shorter than the event-level timeout).
	hookCtx, hookCancel := context.WithTimeout(eventCtx, time.Duration(e.cfg.HookTimeoutMs)*time.Millisecond)
	defer hookCancel()

	L.SetContext(hookCtx)

	// Look up handler function from __hook_handlers.
	handlersVal := L.GetGlobal("__hook_handlers")
	handlersTbl, ok := handlersVal.(*lua.LTable)
	if !ok || handlersVal == lua.LNil {
		return nil
	}

	handlerVal := L.GetField(handlersTbl, entry.handlerKey)
	handlerFn, ok := handlerVal.(*lua.LFunction)
	if !ok {
		return nil
	}

	// Convert data map to Lua table.
	luaData := MapToLuaTable(L, dataMap)

	start := time.Now()

	// Call the handler.
	callErr := L.CallByParam(lua.P{
		Fn:      handlerFn,
		NRet:    0,
		Protect: true,
	}, luaData)

	durationMs := time.Since(start).Milliseconds()

	if callErr != nil {
		// S4: Observability.
		status := "error"
		if errors.Is(hookCtx.Err(), context.DeadlineExceeded) || errors.Is(eventCtx.Err(), context.DeadlineExceeded) {
			status = "timeout"
		}
		e.logger.Warn(
			"hook.before.duration_ms",
			nil,
			"plugin", entry.pluginName,
			"event", string(entry.event),
			"table", entry.table,
			"status", status,
			"duration_ms", durationMs,
		)

		// Phase 4: Record hook execution metric.
		RecordHookExecution(PluginMetricHookBefore, entry.pluginName, string(entry.event), entry.table, status, float64(durationMs))

		// M9: Circuit breaker.
		cbKey := approvalKeyFor(entry.pluginName, string(entry.event), entry.table)
		e.mu.Lock()
		e.consecutiveAborts[cbKey]++
		count := e.consecutiveAborts[cbKey]
		if count >= e.cfg.MaxConsecutiveAborts {
			e.disabled[cbKey] = true
			e.logger.Error(
				"hook.circuit_breaker.tripped",
				nil,
				"plugin", entry.pluginName,
				"event", string(entry.event),
				"table", entry.table,
				"abort_count", count,
			)
		}
		e.mu.Unlock()

		// S4: Abort counter observability.
		e.logger.Warn(
			"hook.before.abort",
			nil,
			"plugin", entry.pluginName,
			"event", string(entry.event),
			"table", entry.table,
			"consecutive_count", count,
		)

		// M5: Error sanitization.
		return audited.NewHookError(entry.pluginName, entry.event, entry.table, callErr.Error())
	}

	// Success: reset circuit breaker counter.
	cbKey := approvalKeyFor(entry.pluginName, string(entry.event), entry.table)
	e.mu.Lock()
	e.consecutiveAborts[cbKey] = 0
	e.mu.Unlock()

	// S4: Observability.
	e.logger.Debug("hook.before.duration_ms",
		"plugin", entry.pluginName,
		"event", string(entry.event),
		"table", entry.table,
		"status", "ok",
		"duration_ms", durationMs,
	)

	// Phase 4: Record hook execution metric.
	RecordHookExecution(PluginMetricHookBefore, entry.pluginName, string(entry.event), entry.table, "ok", float64(durationMs))

	return nil
}

// executeAfter runs a single after-hook asynchronously.
func (e *HookEngine) executeAfter(entry hookEntry, dataMap map[string]any) {
	if e.closing.Load() {
		return
	}

	inst := e.manager.GetPlugin(entry.pluginName)
	if inst == nil || inst.State != StateRunning {
		return
	}

	// Check out VM using shutdown context (request context may be done).
	L, getErr := inst.Pool.GetForHook(e.shutdownCtx)
	if getErr != nil {
		e.logger.Warn(
			fmt.Sprintf("hook after: VM checkout failed for plugin %q: %s", entry.pluginName, getErr.Error()),
			nil,
			"event", string(entry.event),
			"table", entry.table,
		)
		return
	}
	defer inst.Pool.Put(L)

	// M10: Reduced op budget for after-hooks.
	inst.mu.Lock()
	dbAPI := inst.dbAPIs[L]
	inst.mu.Unlock()

	if dbAPI != nil {
		dbAPI.ResetOpCount(e.cfg.HookMaxOps)
	}

	// After-hook timeout uses the standard exec timeout (not the shorter hook timeout).
	execCtx, cancel := context.WithTimeout(e.shutdownCtx, time.Duration(e.cfg.ExecTimeoutMs)*time.Millisecond)
	defer cancel()

	L.SetContext(execCtx)

	// Look up handler.
	handlersVal := L.GetGlobal("__hook_handlers")
	handlersTbl, ok := handlersVal.(*lua.LTable)
	if !ok || handlersVal == lua.LNil {
		return
	}

	handlerVal := L.GetField(handlersTbl, entry.handlerKey)
	handlerFn, ok := handlerVal.(*lua.LFunction)
	if !ok {
		return
	}

	luaData := MapToLuaTable(L, dataMap)

	start := time.Now()

	callErr := L.CallByParam(lua.P{
		Fn:      handlerFn,
		NRet:    0,
		Protect: true,
	}, luaData)

	durationMs := time.Since(start).Milliseconds()

	status := "ok"
	if callErr != nil {
		status = "error"
		if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			status = "timeout"
		}
		e.logger.Warn(
			fmt.Sprintf("hook after error: plugin %q: %s", entry.pluginName, callErr.Error()),
			nil,
			"event", string(entry.event),
			"table", entry.table,
		)
	}

	// S4: Observability.
	e.logger.Debug("hook.after.duration_ms",
		"plugin", entry.pluginName,
		"event", string(entry.event),
		"table", entry.table,
		"status", status,
		"duration_ms", durationMs,
	)

	// Phase 4: Record hook execution metric.
	RecordHookExecution(PluginMetricHookAfter, entry.pluginName, string(entry.event), entry.table, status, float64(durationMs))
}

// Close shuts down the hook engine: sets closing flag, cancels shutdown context,
// and waits for in-flight after-hooks to complete (with ctx deadline as backstop).
//
// Must be called AFTER HTTP servers stop but BEFORE Manager.Shutdown closes pools.
func (e *HookEngine) Close(ctx context.Context) {
	e.closing.Store(true)
	e.shutdownCancel()

	done := make(chan struct{})
	go func() {
		e.afterWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All after-hooks completed.
	case <-ctx.Done():
		e.logger.Warn(
			"HookEngine shutdown: context expired before all after-hooks drained", nil,
		)
	}
}

// SetHookEnabled enables or disables a specific hook at runtime. When disabled,
// the hook is skipped during dispatch. Used for admin control and circuit breaker
// recovery. Restart clears all disabled state (hooks are rebuilt from init.lua).
func (e *HookEngine) SetHookEnabled(plugin, event, table string, enabled bool) {
	key := approvalKeyFor(plugin, event, table)
	e.mu.Lock()
	defer e.mu.Unlock()

	if enabled {
		delete(e.disabled, key)
		e.consecutiveAborts[key] = 0
	} else {
		e.disabled[key] = true
	}
}

// CreatePluginHooksTable creates the plugin_hooks approval table using DDL
// appropriate for the configured dialect. Called by Manager.LoadAll before
// loading plugins.
func (e *HookEngine) CreatePluginHooksTable(ctx context.Context) error {
	var ddl string

	switch e.dialect {
	case db.DialectSQLite:
		ddl = `CREATE TABLE IF NOT EXISTS plugin_hooks (
    plugin_name    TEXT    NOT NULL,
    event          TEXT    NOT NULL,
    table_name     TEXT    NOT NULL,
    approved       INTEGER NOT NULL DEFAULT 0,
    approved_at    TEXT,
    approved_by    TEXT,
    plugin_version TEXT    NOT NULL DEFAULT '',
    PRIMARY KEY (plugin_name, event, table_name)
)`
	case db.DialectMySQL:
		ddl = `CREATE TABLE IF NOT EXISTS plugin_hooks (
    plugin_name    VARCHAR(255) NOT NULL,
    event          VARCHAR(64)  NOT NULL,
    table_name     VARCHAR(255) NOT NULL,
    approved       TINYINT(1)   NOT NULL DEFAULT 0,
    approved_at    TIMESTAMP    NULL DEFAULT NULL,
    approved_by    VARCHAR(255),
    plugin_version VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (plugin_name, event, table_name)
)`
	case db.DialectPostgres:
		ddl = `CREATE TABLE IF NOT EXISTS plugin_hooks (
    plugin_name    TEXT    NOT NULL,
    event          TEXT    NOT NULL,
    table_name     TEXT    NOT NULL,
    approved       BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at    TIMESTAMPTZ,
    approved_by    TEXT,
    plugin_version TEXT    NOT NULL DEFAULT '',
    PRIMARY KEY (plugin_name, event, table_name)
)`
	default:
		return fmt.Errorf("unsupported dialect for plugin_hooks table: %d", e.dialect)
	}

	_, err := e.dbPool.ExecContext(ctx, ddl)
	if err != nil {
		return fmt.Errorf("creating plugin_hooks table: %w", err)
	}

	return nil
}

// UpsertHookRegistrations records hook registrations in the plugin_hooks table
// and loads approval state into memory. Called once per plugin during loadPlugin.
func (e *HookEngine) UpsertHookRegistrations(ctx context.Context, pluginName, pluginVersion string, hooks []PendingHook) error {
	for _, h := range hooks {
		if err := e.upsertHook(ctx, pluginName, pluginVersion, string(h.Event), h.Table); err != nil {
			return fmt.Errorf("upserting hook %s %s: %w", h.Event, h.Table, err)
		}

		// Load approval state.
		approved, readErr := e.readHookApproval(ctx, pluginName, string(h.Event), h.Table)
		if readErr != nil {
			return fmt.Errorf("reading hook approval %s %s: %w", h.Event, h.Table, readErr)
		}

		key := approvalKeyFor(pluginName, string(h.Event), h.Table)
		e.mu.Lock()
		e.approved[key] = approved
		e.mu.Unlock()
	}

	return nil
}

// upsertHook inserts or updates a single hook registration.
func (e *HookEngine) upsertHook(ctx context.Context, pluginName, pluginVersion, event, tableName string) error {
	switch e.dialect {
	case db.DialectSQLite:
		_, err := e.dbPool.ExecContext(ctx, `
			INSERT INTO plugin_hooks (plugin_name, event, table_name, approved, plugin_version)
			VALUES (?, ?, ?, 0, ?)
			ON CONFLICT(plugin_name, event, table_name) DO UPDATE SET
				plugin_version = excluded.plugin_version
		`, pluginName, event, tableName, pluginVersion)
		return err

	case db.DialectMySQL:
		_, err := e.dbPool.ExecContext(ctx, `
			INSERT INTO plugin_hooks (plugin_name, event, table_name, approved, plugin_version)
			VALUES (?, ?, ?, 0, ?)
			ON DUPLICATE KEY UPDATE
				plugin_version = VALUES(plugin_version)
		`, pluginName, event, tableName, pluginVersion)
		return err

	case db.DialectPostgres:
		_, err := e.dbPool.ExecContext(ctx, `
			INSERT INTO plugin_hooks (plugin_name, event, table_name, approved, plugin_version)
			VALUES ($1, $2, $3, FALSE, $4)
			ON CONFLICT (plugin_name, event, table_name) DO UPDATE SET
				plugin_version = EXCLUDED.plugin_version
		`, pluginName, event, tableName, pluginVersion)
		return err

	default:
		return fmt.Errorf("unsupported dialect: %d", e.dialect)
	}
}

// readHookApproval reads the approval status from the plugin_hooks table.
func (e *HookEngine) readHookApproval(ctx context.Context, pluginName, event, tableName string) (bool, error) {
	var approvedInt int

	var query string
	var args []any
	switch e.dialect {
	case db.DialectPostgres:
		query = "SELECT CASE WHEN approved THEN 1 ELSE 0 END FROM plugin_hooks WHERE plugin_name = $1 AND event = $2 AND table_name = $3"
		args = []any{pluginName, event, tableName}
	default:
		query = "SELECT approved FROM plugin_hooks WHERE plugin_name = ? AND event = ? AND table_name = ?"
		args = []any{pluginName, event, tableName}
	}

	err := e.dbPool.QueryRowContext(ctx, query, args...).Scan(&approvedInt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return approvedInt != 0, nil
}

// ApproveHook approves a specific hook, updating the DB and in-memory state.
func (e *HookEngine) ApproveHook(ctx context.Context, pluginName, event, tableName, approvedBy string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	var query string
	var args []any
	switch e.dialect {
	case db.DialectPostgres:
		query = "UPDATE plugin_hooks SET approved = TRUE, approved_at = $1, approved_by = $2 WHERE plugin_name = $3 AND event = $4 AND table_name = $5"
		args = []any{now, approvedBy, pluginName, event, tableName}
	default:
		query = "UPDATE plugin_hooks SET approved = 1, approved_at = ?, approved_by = ? WHERE plugin_name = ? AND event = ? AND table_name = ?"
		args = []any{now, approvedBy, pluginName, event, tableName}
	}

	_, err := e.dbPool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("approving hook in DB: %w", err)
	}

	key := approvalKeyFor(pluginName, event, tableName)
	e.mu.Lock()
	e.approved[key] = true
	e.mu.Unlock()

	return nil
}

// RevokeHook revokes approval for a specific hook.
func (e *HookEngine) RevokeHook(ctx context.Context, pluginName, event, tableName string) error {
	var query string
	var args []any
	switch e.dialect {
	case db.DialectPostgres:
		query = "UPDATE plugin_hooks SET approved = FALSE, approved_at = NULL, approved_by = NULL WHERE plugin_name = $1 AND event = $2 AND table_name = $3"
		args = []any{pluginName, event, tableName}
	default:
		query = "UPDATE plugin_hooks SET approved = 0, approved_at = NULL, approved_by = NULL WHERE plugin_name = ? AND event = ? AND table_name = ?"
		args = []any{pluginName, event, tableName}
	}

	_, err := e.dbPool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("revoking hook in DB: %w", err)
	}

	key := approvalKeyFor(pluginName, event, tableName)
	e.mu.Lock()
	e.approved[key] = false
	e.mu.Unlock()

	return nil
}

// CleanupOrphanedHooks deletes rows from plugin_hooks where plugin_name is
// not in the discovered set.
func (e *HookEngine) CleanupOrphanedHooks(ctx context.Context, discoveredPlugins []string) error {
	if len(discoveredPlugins) == 0 {
		_, err := e.dbPool.ExecContext(ctx, "DELETE FROM plugin_hooks")
		if err != nil {
			return fmt.Errorf("cleaning all orphaned hooks: %w", err)
		}
		return nil
	}

	placeholders := make([]string, len(discoveredPlugins))
	args := make([]any, len(discoveredPlugins))
	for i, name := range discoveredPlugins {
		switch e.dialect {
		case db.DialectPostgres:
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		default:
			placeholders[i] = "?"
		}
		args[i] = name
	}

	query := fmt.Sprintf(
		"DELETE FROM plugin_hooks WHERE plugin_name NOT IN (%s)",
		strings.Join(placeholders, ", "),
	)

	_, err := e.dbPool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("cleaning orphaned hooks: %w", err)
	}

	return nil
}

// UnregisterPlugin removes all hook index entries for the given plugin.
// Rebuilds the hasHooks map and hasAnyHook flag from remaining entries.
// Called during hot reload (Phase 4) after the new instance has registered
// its hooks, to remove the old instance's entries.
func (e *HookEngine) UnregisterPlugin(pluginName string) {
	// Remove entries from hookIndex.
	for key, entries := range e.hookIndex {
		filtered := entries[:0]
		for _, entry := range entries {
			if entry.pluginName != pluginName {
				filtered = append(filtered, entry)
			}
		}
		if len(filtered) == 0 {
			delete(e.hookIndex, key)
		} else {
			e.hookIndex[key] = filtered
		}
	}

	// Rebuild hasHooks from remaining entries.
	newHasHooks := make(map[string]bool)
	newHasAnyHook := false
	for key, entries := range e.hookIndex {
		if len(entries) > 0 {
			newHasHooks[key] = true
			newHasAnyHook = true
			for _, entry := range entries {
				if entry.isWildcard {
					newHasHooks[string(entry.event)+":*"] = true
				}
			}
		}
	}
	e.hasHooks = newHasHooks
	e.hasAnyHook = newHasAnyHook

	// Clear approval and circuit breaker state for this plugin.
	e.mu.Lock()
	for key := range e.approved {
		if strings.HasPrefix(key, pluginName+":") {
			delete(e.approved, key)
		}
	}
	for key := range e.disabled {
		if strings.HasPrefix(key, pluginName+":") {
			delete(e.disabled, key)
		}
	}
	for key := range e.consecutiveAborts {
		if strings.HasPrefix(key, pluginName+":") {
			delete(e.consecutiveAborts, key)
		}
	}
	e.mu.Unlock()
}

// HookRegistration represents a single registered hook with its metadata
// and approval status. Used by ListHooks() for the admin API response.
type HookRegistration struct {
	PluginName string `json:"plugin_name"`
	Event      string `json:"event"`
	Table      string `json:"table"`
	Priority   int    `json:"priority"`
	Approved   bool   `json:"approved"`
	IsWildcard bool   `json:"is_wildcard"`
}

// ListHooks returns all registered hooks with their approval status.
// Thread-safe: reads hookIndex (immutable after LoadAll) and approval map
// (protected by mu). Used by the admin hooks list endpoint.
func (e *HookEngine) ListHooks() []HookRegistration {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []HookRegistration
	for _, entries := range e.hookIndex {
		for _, entry := range entries {
			approvalKey := approvalKeyFor(entry.pluginName, string(entry.event), entry.table)
			reg := HookRegistration{
				PluginName: entry.pluginName,
				Event:      string(entry.event),
				Table:      entry.table,
				Priority:   entry.priority,
				Approved:   e.approved[approvalKey],
				IsWildcard: entry.isWildcard,
			}
			result = append(result, reg)
		}
	}

	// Sort for deterministic output: plugin name, event, table.
	sort.Slice(result, func(i, j int) bool {
		if result[i].PluginName != result[j].PluginName {
			return result[i].PluginName < result[j].PluginName
		}
		if result[i].Event != result[j].Event {
			return result[i].Event < result[j].Event
		}
		return result[i].Table < result[j].Table
	})

	return result
}

// approvalKeyFor builds the map key for approval and circuit breaker lookups.
func approvalKeyFor(plugin, event, table string) string {
	return plugin + ":" + event + ":" + table
}
