package plugin

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
	lua "github.com/yuin/gopher-lua"
)

// MaxHooksPerPlugin is the maximum number of hooks a single plugin can register
// via hooks.on(). Exceeding this limit raises a Lua error.
const MaxHooksPerPlugin = 50

// PendingHook holds the metadata for a single hook registration captured during
// init.lua module-scope execution. The HookEngine processes these after loading.
type PendingHook struct {
	Event      audited.HookEvent
	Table      string // audited table name or "*" for wildcard
	Priority   int    // 1-1000, default 100
	IsWildcard bool
	// HandlerKey is the Lua registry key for the handler function, stored in
	// __hook_handlers[handlerKey]. This allows the HookEngine to look up the
	// correct handler from any checked-out VM (all VMs load the same init.lua).
	HandlerKey string
}

// RegisterHooksAPI creates a "hooks" global Lua table with a single function: on.
//
// It also creates two hidden global tables used by the HookEngine to read
// registered hooks at load time:
//   - __hook_handlers:  handlerKey -> handler LFunction
//   - __hook_pending:   ordered array of pending hook metadata tables
//
// These tables are part of the global snapshot taken by SnapshotGlobals and
// must NOT be modified after snapshot. The engine reads them once during loadPlugin.
//
// Phase guard: hooks.on() must only be called at module scope (during init.lua
// execution by the factory), not inside on_init(). This follows the same pattern
// as http.handle() -- each VM needs its own registered Lua function reference.
func RegisterHooksAPI(L *lua.LState, pluginName string) {
	handlers := L.NewTable()
	pending := L.NewTable()

	L.SetGlobal("__hook_handlers", handlers)
	L.SetGlobal("__hook_pending", pending)

	hooksTable := L.NewTable()
	hooksTable.RawSetString("on", L.NewFunction(hooksOnFn(L, pluginName, handlers, pending)))

	L.SetGlobal("hooks", hooksTable)
}

// hooksOnFn returns the Go-bound function for hooks.on(event, table, handler [, opts]).
//
// Validation order:
//  1. Event against ValidHookEvents
//  2. Table name (non-empty string, or "*" for wildcard)
//  3. Handler is a function
//  4. Phase guard (rejects calls inside on_init)
//  5. Hook count limit
//  6. Options table parsing (priority)
//  7. Store handler and pending metadata
func hooksOnFn(L *lua.LState, pluginName string, handlers *lua.LTable, pending *lua.LTable) lua.LGFunction {
	return func(L *lua.LState) int {
		eventStr := L.CheckString(1)
		tableName := L.CheckString(2)
		handler := L.CheckFunction(3)

		// 1. Validate event.
		_, validEvent := audited.ValidHookEvents[eventStr]
		if !validEvent {
			L.ArgError(1, fmt.Sprintf("invalid hook event %q", eventStr))
			return 0
		}

		// 2. Validate table name.
		if tableName == "" {
			L.ArgError(2, "table name cannot be empty")
			return 0
		}

		// 3. Phase guard: reject calls inside on_init().
		registryTbl := L.Get(lua.RegistryIndex)
		if regTbl, ok := registryTbl.(*lua.LTable); ok {
			inInit := L.GetField(regTbl, "in_init")
			if inInit == lua.LTrue {
				L.RaiseError("hooks.on() must be called at module scope, not inside on_init()")
				return 0
			}
		}

		// 4. Hook count limit.
		count := 0
		pending.ForEach(func(_, _ lua.LValue) {
			count++
		})
		if count >= MaxHooksPerPlugin {
			L.RaiseError("plugin %q exceeded maximum hook limit (%d)", pluginName, MaxHooksPerPlugin)
			return 0
		}

		// 5. Parse options table (4th argument, optional).
		priority := 100
		if L.GetTop() >= 4 {
			optVal := L.Get(4)
			if optTbl, ok := optVal.(*lua.LTable); ok {
				prioVal := L.GetField(optTbl, "priority")
				if n, ok := prioVal.(lua.LNumber); ok {
					p := int(n)
					if p < 1 {
						p = 1
					}
					if p > 1000 {
						p = 1000
					}
					priority = p
				}
			}
		}

		// 6. Generate handler key and store the handler function.
		nextIdx := pending.Len() + 1
		handlerKey := fmt.Sprintf("hook_%s_%s_%d", eventStr, tableName, nextIdx)

		handlers.RawSetString(handlerKey, handler)

		// 7. Store pending hook metadata as a Lua table in the pending array.
		meta := L.NewTable()
		meta.RawSetString("event", lua.LString(eventStr))
		meta.RawSetString("table", lua.LString(tableName))
		meta.RawSetString("priority", lua.LNumber(priority))
		meta.RawSetString("handler_key", lua.LString(handlerKey))
		if tableName == "*" {
			meta.RawSetString("is_wildcard", lua.LTrue)
		} else {
			meta.RawSetString("is_wildcard", lua.LFalse)
		}

		L.RawSetInt(pending, nextIdx, meta)

		return 0
	}
}

// ReadPendingHooks extracts PendingHook entries from a checked-out VM's
// __hook_pending table. Called once per plugin during loadPlugin.
func ReadPendingHooks(L *lua.LState) []PendingHook {
	pendingVal := L.GetGlobal("__hook_pending")
	pendingTbl, ok := pendingVal.(*lua.LTable)
	if !ok || pendingVal == lua.LNil {
		return nil
	}

	var hooks []PendingHook
	pendingLen := pendingTbl.Len()
	for i := 1; i <= pendingLen; i++ {
		entry := L.RawGetInt(pendingTbl, i)
		entryTbl, ok := entry.(*lua.LTable)
		if !ok {
			continue
		}

		eventStr := ""
		if s, ok := L.GetField(entryTbl, "event").(lua.LString); ok {
			eventStr = string(s)
		}

		tableName := ""
		if s, ok := L.GetField(entryTbl, "table").(lua.LString); ok {
			tableName = string(s)
		}

		priority := 100
		if n, ok := L.GetField(entryTbl, "priority").(lua.LNumber); ok {
			priority = int(n)
		}

		handlerKey := ""
		if s, ok := L.GetField(entryTbl, "handler_key").(lua.LString); ok {
			handlerKey = string(s)
		}

		isWildcard := false
		if b, ok := L.GetField(entryTbl, "is_wildcard").(lua.LBool); ok {
			isWildcard = bool(b)
		}

		event, valid := audited.ValidHookEvents[eventStr]
		if !valid {
			continue
		}

		hooks = append(hooks, PendingHook{
			Event:      event,
			Table:      tableName,
			Priority:   priority,
			IsWildcard: isWildcard,
			HandlerKey: handlerKey,
		})
	}

	return hooks
}
