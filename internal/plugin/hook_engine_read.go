package plugin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/db/audited"
	lua "github.com/yuin/gopher-lua"
)

// HasReadHooks implements audited.ReadHookRunner. Returns true if any
// before_read or after_read hooks are registered for the given table.
func (e *HookEngine) HasReadHooks(table string) bool {
	if !e.hasAnyHook {
		return false
	}
	if e.hasHooks[string(audited.HookBeforeRead)+":"+table] {
		return true
	}
	return e.hasHooks[string(audited.HookAfterRead)+":"+table]
}

// RunBeforeReadHooks implements audited.ReadHookRunner. Executes all matching
// before_read hooks synchronously. Unlike mutation before-hooks, read hooks:
//   - Do NOT set inBeforeHook (db.*, core.*, request.* are all permitted)
//   - Do NOT run inside a database transaction
//   - Return structured responses (NRet: 1) instead of aborting via error()
//
// Returns:
//   - (*ReadHookResponse, state, nil): A hook wants to abort delivery. Write the response.
//   - (nil, state, nil): All hooks returned nil. Proceed with delivery. State contains
//     any _-prefixed keys set by hooks on the data table.
//   - (nil, nil, error): Internal error during hook execution.
func (e *HookEngine) RunBeforeReadHooks(ctx context.Context, table string, data map[string]any) (*audited.ReadHookResponse, map[string]any, error) {
	entries := e.gatherEntries(audited.HookBeforeRead, table)
	if len(entries) == 0 {
		return nil, nil, nil
	}

	// Inject metadata fields.
	if data == nil {
		data = make(map[string]any)
	}
	data["_table"] = table
	data["_event"] = string(audited.HookBeforeRead)

	eventCtx, eventCancel := context.WithTimeout(ctx, time.Duration(e.cfg.EventTimeoutMs)*time.Millisecond)
	defer eventCancel()

	start := time.Now()
	var state map[string]any

	for _, entry := range entries {
		resp, updatedData, err := e.executeBeforeRead(eventCtx, entry, data)
		if err != nil {
			return nil, nil, err
		}
		if updatedData != nil {
			state = updatedData
		}
		if resp != nil {
			e.logger.Debug("hook.before_read.abort",
				"plugin", entry.pluginName,
				"table", table,
				"status", resp.Status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
			return resp, nil, nil
		}
	}

	e.logger.Debug("hook.before_read.total_duration_ms",
		"table", table,
		"hook_count", len(entries),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return nil, state, nil
}

// RunAfterReadHooks implements audited.ReadHookRunner. Executes all matching
// after_read hooks synchronously (NOT fire-and-forget). Collects and returns
// response headers from all hooks. Like before_read, db.* and request.* are
// permitted.
//
// The state map from RunBeforeReadHooks is merged into the data table so
// after-read hooks can access _-prefixed keys set by before-read hooks.
func (e *HookEngine) RunAfterReadHooks(ctx context.Context, table string, data map[string]any, state map[string]any) (map[string]string, error) {
	entries := e.gatherEntries(audited.HookAfterRead, table)
	if len(entries) == 0 {
		return nil, nil
	}

	// Merge state from before_read into data.
	if data == nil {
		data = make(map[string]any)
	}
	if state != nil {
		for k, v := range state {
			data[k] = v
		}
	}
	data["_table"] = table
	data["_event"] = string(audited.HookAfterRead)

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(e.cfg.ExecTimeoutMs)*time.Millisecond)
	defer cancel()

	start := time.Now()
	collectedHeaders := make(map[string]string)

	for _, entry := range entries {
		headers, err := e.executeAfterRead(execCtx, entry, data)
		if err != nil {
			// Log but continue — after-read errors should not break delivery.
			e.logger.Warn(
				fmt.Sprintf("hook after_read error: plugin %q: %s", entry.pluginName, err.Error()),
				nil,
				"event", string(entry.event),
				"table", entry.table,
			)
			continue
		}
		for k, v := range headers {
			collectedHeaders[k] = v
		}
	}

	e.logger.Debug("hook.after_read.total_duration_ms",
		"table", table,
		"hook_count", len(entries),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	if len(collectedHeaders) == 0 {
		return nil, nil
	}
	return collectedHeaders, nil
}

// executeBeforeRead runs a single before_read hook synchronously.
// Does NOT set inBeforeHook — db.*, core.*, and request.* are all permitted.
// Returns the Go map read back from the Lua data table for state capture when
// the hook returns nil.
func (e *HookEngine) executeBeforeRead(eventCtx context.Context, entry hookEntry, dataMap map[string]any) (*audited.ReadHookResponse, map[string]any, error) {
	inst := e.manager.GetPlugin(entry.pluginName)
	if inst == nil || inst.State != StateRunning {
		return nil, nil, nil
	}

	L, getErr := inst.Pool.GetForHook(eventCtx)
	if getErr != nil {
		if errors.Is(getErr, ErrPoolExhausted) {
			e.logger.Warn(
				fmt.Sprintf("hook before_read: VM pool exhausted for plugin %q", entry.pluginName),
				nil,
				"table", entry.table,
			)
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("hook VM checkout: %w", getErr)
	}
	defer inst.Pool.Put(L)

	// NOTE: We intentionally do NOT set inBeforeHook here.
	// Read hooks run outside any DB transaction, so db.*, core.*, and
	// request.* are all safe to call.

	hookCtx, hookCancel := context.WithTimeout(eventCtx, time.Duration(e.cfg.HookTimeoutMs)*time.Millisecond)
	defer hookCancel()

	L.SetContext(hookCtx)

	handlersVal := L.GetGlobal("__hook_handlers")
	handlersTbl, ok := handlersVal.(*lua.LTable)
	if !ok || handlersVal == lua.LNil {
		return nil, nil, nil
	}

	handlerVal := L.GetField(handlersTbl, entry.handlerKey)
	handlerFn, ok := handlerVal.(*lua.LFunction)
	if !ok {
		return nil, nil, nil
	}

	luaData := MapToLuaTable(L, dataMap)
	start := time.Now()

	// NRet: 1 — read hooks can return a structured response table or nil.
	callErr := L.CallByParam(lua.P{
		Fn:      handlerFn,
		NRet:    1,
		Protect: true,
	}, luaData)

	durationMs := time.Since(start).Milliseconds()

	if callErr != nil {
		status := "error"
		if errors.Is(hookCtx.Err(), context.DeadlineExceeded) || errors.Is(eventCtx.Err(), context.DeadlineExceeded) {
			status = "timeout"
		}
		e.logger.Warn(
			"hook.before_read.duration_ms",
			nil,
			"plugin", entry.pluginName,
			"table", entry.table,
			"status", status,
			"duration_ms", durationMs,
		)

		RecordHookExecution(PluginMetricHookBefore, entry.pluginName, string(entry.event), entry.table, status, float64(durationMs))

		// Circuit breaker for read hooks.
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

		// Read hook errors are not fatal — skip this hook and continue.
		return nil, nil, nil
	}

	// Success: reset circuit breaker.
	cbKey := approvalKeyFor(entry.pluginName, string(entry.event), entry.table)
	e.mu.Lock()
	e.consecutiveAborts[cbKey] = 0
	e.mu.Unlock()

	RecordHookExecution(PluginMetricHookBefore, entry.pluginName, string(entry.event), entry.table, "ok", float64(durationMs))

	// Read the return value from the Lua stack.
	retVal := L.Get(-1)
	L.Pop(1)

	// Read the Lua data table back to a Go map while the VM is still checked
	// out. This captures any _-prefixed keys the hook set (state passing).
	updatedData := LuaTableToMap(L, luaData)

	if retVal == lua.LNil || retVal == nil {
		return nil, updatedData, nil
	}

	retTbl, ok := retVal.(*lua.LTable)
	if !ok {
		return nil, updatedData, nil
	}

	// Check for status field — indicates an abort response.
	statusVal := L.GetField(retTbl, "status")
	if statusVal == lua.LNil {
		return nil, updatedData, nil
	}

	statusNum, ok := statusVal.(lua.LNumber)
	if !ok {
		return nil, updatedData, nil
	}

	resp := &audited.ReadHookResponse{
		Status: int(statusNum),
	}

	// Parse headers.
	headersVal := L.GetField(retTbl, "headers")
	if headersTbl, ok := headersVal.(*lua.LTable); ok {
		resp.Headers = make(map[string]string)
		headersTbl.ForEach(func(key, value lua.LValue) {
			if k, ok := key.(lua.LString); ok {
				if v, ok := value.(lua.LString); ok {
					resp.Headers[string(k)] = string(v)
				}
			}
		})
	}

	// Parse json body.
	jsonVal := L.GetField(retTbl, "json")
	if jsonTbl, ok := jsonVal.(*lua.LTable); ok {
		resp.Body = LuaTableToMap(L, jsonTbl)
	}

	return resp, nil, nil
}

// executeAfterRead runs a single after_read hook synchronously.
// Does NOT set inBeforeHook — db.* and request.* are permitted.
// Returns headers to append to the response.
func (e *HookEngine) executeAfterRead(execCtx context.Context, entry hookEntry, dataMap map[string]any) (map[string]string, error) {
	inst := e.manager.GetPlugin(entry.pluginName)
	if inst == nil || inst.State != StateRunning {
		return nil, nil
	}

	L, getErr := inst.Pool.GetForHook(execCtx)
	if getErr != nil {
		return nil, fmt.Errorf("hook VM checkout: %w", getErr)
	}
	defer inst.Pool.Put(L)

	// Reduced op budget for after-hooks (same as mutation after-hooks).
	inst.mu.Lock()
	dbAPI := inst.dbAPIs[L]
	inst.mu.Unlock()

	if dbAPI != nil {
		dbAPI.ResetOpCount(e.cfg.HookMaxOps)
	}

	hookCtx, hookCancel := context.WithTimeout(execCtx, time.Duration(e.cfg.HookTimeoutMs)*time.Millisecond)
	defer hookCancel()

	L.SetContext(hookCtx)

	handlersVal := L.GetGlobal("__hook_handlers")
	handlersTbl, ok := handlersVal.(*lua.LTable)
	if !ok || handlersVal == lua.LNil {
		return nil, nil
	}

	handlerVal := L.GetField(handlersTbl, entry.handlerKey)
	handlerFn, ok := handlerVal.(*lua.LFunction)
	if !ok {
		return nil, nil
	}

	luaData := MapToLuaTable(L, dataMap)
	start := time.Now()

	// NRet: 1 — after-read hooks return { headers = { ... } } or nil.
	callErr := L.CallByParam(lua.P{
		Fn:      handlerFn,
		NRet:    1,
		Protect: true,
	}, luaData)

	durationMs := time.Since(start).Milliseconds()

	status := "ok"
	if callErr != nil {
		status = "error"
		if errors.Is(hookCtx.Err(), context.DeadlineExceeded) {
			status = "timeout"
		}
		e.logger.Warn(
			fmt.Sprintf("hook after_read error: plugin %q: %s", entry.pluginName, callErr.Error()),
			nil,
			"event", string(entry.event),
			"table", entry.table,
		)
		RecordHookExecution(PluginMetricHookAfter, entry.pluginName, string(entry.event), entry.table, status, float64(durationMs))
		return nil, nil
	}

	RecordHookExecution(PluginMetricHookAfter, entry.pluginName, string(entry.event), entry.table, status, float64(durationMs))

	// Read the return value.
	retVal := L.Get(-1)
	L.Pop(1)

	if retVal == lua.LNil || retVal == nil {
		return nil, nil
	}

	retTbl, ok := retVal.(*lua.LTable)
	if !ok {
		return nil, nil
	}

	// Parse headers from return table.
	headersVal := L.GetField(retTbl, "headers")
	headersTbl, ok := headersVal.(*lua.LTable)
	if !ok {
		return nil, nil
	}

	headers := make(map[string]string)
	headersTbl.ForEach(func(key, value lua.LValue) {
		if k, ok := key.(lua.LString); ok {
			if v, ok := value.(lua.LString); ok {
				headers[string(k)] = string(v)
			}
		}
	})

	if len(headers) == 0 {
		return nil, nil
	}
	return headers, nil
}
