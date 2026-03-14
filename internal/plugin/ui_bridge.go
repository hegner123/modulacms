package plugin

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// CoroutineBridge manages a single gopher-lua coroutine for plugin UI rendering.
// It bridges the Bubbletea Update/View cycle with Lua's coroutine yield/resume
// pattern. The bridge is used by screens, inline field interfaces, and overlay
// field interfaces — all three share the same lifecycle.
//
// Thread safety: CoroutineBridge is NOT thread-safe. It must only be called
// from the Bubbletea update goroutine (single goroutine guarantee).
type CoroutineBridge struct {
	parentL     *lua.LState    // checked out from UIVMPool; returned on Close
	thread      *lua.LState    // child coroutine created from parentL
	entryFn     *lua.LFunction // screen(ctx) or interface(ctx)
	plugin      *PluginInstance
	started     bool
	done        bool
	renderingUI bool // true while coroutine is active; used by UI pool drain
	lastErr     error
}

// NewCoroutineBridge creates a bridge for a plugin UI coroutine.
// L must be a VM checked out from the UI pool. fn is the screen() or interface()
// entry function loaded from the plugin's Lua file.
func NewCoroutineBridge(plugin *PluginInstance, L *lua.LState, fn *lua.LFunction) *CoroutineBridge {
	return &CoroutineBridge{
		parentL: L,
		entryFn: fn,
		plugin:  plugin,
	}
}

// Start initializes the coroutine and sends the init event. Returns the first
// yield value (either a layout or an action). The initEvent table is passed as
// the ctx argument to the screen/interface function.
func (cb *CoroutineBridge) Start(initEvent *lua.LTable) (YieldValue, error) {
	if cb.started {
		return YieldValue{}, fmt.Errorf("coroutine already started")
	}
	if cb.done {
		return YieldValue{}, fmt.Errorf("coroutine is done")
	}

	cb.thread, _ = cb.parentL.NewThread()
	cb.started = true
	cb.renderingUI = true

	// Call: screen(ctx) or interface(ctx) with initEvent as ctx.
	// The function runs until its first coroutine.yield().
	st, err, values := cb.parentL.Resume(cb.thread, cb.entryFn, initEvent)
	if err != nil {
		cb.done = true
		cb.renderingUI = false
		cb.lastErr = err
		return YieldValue{}, fmt.Errorf("coroutine start error: %w", err)
	}

	if st == lua.ResumeOK {
		// Function returned without yielding — treat as quit/cancel.
		cb.done = true
		cb.renderingUI = false
		return YieldValue{IsAction: true, Action: &PluginAction{Name: "quit"}}, nil
	}

	// st == lua.ResumeYield
	return cb.parseYield(values)
}

// Resume sends an event to the coroutine and returns the next yield value.
// The event table is the return value of coroutine.yield() in Lua.
func (cb *CoroutineBridge) Resume(event *lua.LTable) (YieldValue, error) {
	if !cb.started {
		return YieldValue{}, fmt.Errorf("coroutine not started")
	}
	if cb.done {
		return YieldValue{}, fmt.Errorf("coroutine is done")
	}

	st, err, values := cb.parentL.Resume(cb.thread, cb.entryFn, event)
	if err != nil {
		cb.done = true
		cb.renderingUI = false
		cb.lastErr = err
		return YieldValue{}, fmt.Errorf("coroutine resume error: %w", err)
	}

	if st == lua.ResumeOK {
		// Function returned — treat as quit/cancel.
		cb.done = true
		cb.renderingUI = false
		return YieldValue{IsAction: true, Action: &PluginAction{Name: "quit"}}, nil
	}

	return cb.parseYield(values)
}

// Close terminates the coroutine. The parent LState is NOT returned to the pool
// here — the caller (screen/bubble/overlay host) is responsible for returning
// it via UIVMPool.Put(). This separation allows the host to control timing.
func (cb *CoroutineBridge) Close() {
	cb.done = true
	cb.renderingUI = false
}

// Done returns true if the coroutine has finished (returned or errored).
func (cb *CoroutineBridge) Done() bool {
	return cb.done
}

// RenderingUI returns true while the coroutine is active. Used by the UI pool
// drain logic to detect VMs that should not be reclaimed.
func (cb *CoroutineBridge) RenderingUI() bool {
	return cb.renderingUI
}

// ParentL returns the parent LState for pool return on close.
func (cb *CoroutineBridge) ParentL() *lua.LState {
	return cb.parentL
}

// LastError returns the last error that caused the coroutine to terminate.
func (cb *CoroutineBridge) LastError() error {
	return cb.lastErr
}

// parseYield interprets the values yielded by the coroutine.
// If the yielded table has an "action" field, it's an action yield.
// Otherwise, it's a layout/primitive yield.
func (cb *CoroutineBridge) parseYield(values []lua.LValue) (YieldValue, error) {
	if len(values) == 0 {
		return YieldValue{}, fmt.Errorf("coroutine yielded no values")
	}

	tbl, ok := values[0].(*lua.LTable)
	if !ok {
		return YieldValue{}, fmt.Errorf("coroutine yielded non-table value: %s", values[0].Type().String())
	}

	// Check for action yield.
	actionVal := tbl.RawGetString("action")
	if actionVal != lua.LNil {
		action, err := ParseAction(tbl)
		if err != nil {
			return YieldValue{}, fmt.Errorf("invalid action yield: %w", err)
		}
		return YieldValue{IsAction: true, Action: action}, nil
	}

	// Layout/primitive yield.
	typeVal := tbl.RawGetString("type")
	typeStr, ok := typeVal.(lua.LString)
	if !ok {
		return YieldValue{}, fmt.Errorf("yield missing 'type' field")
	}

	if string(typeStr) == "grid" {
		layout, err := ParseLayout(tbl)
		if err != nil {
			return YieldValue{}, fmt.Errorf("invalid layout yield: %w", err)
		}
		return YieldValue{Layout: layout}, nil
	}

	// Single primitive (inline interface or standalone use).
	prim, err := ParsePrimitive(tbl)
	if err != nil {
		return YieldValue{}, fmt.Errorf("invalid primitive yield: %w", err)
	}
	return YieldValue{Primitive: prim}, nil
}

// YieldValue represents the parsed result of a coroutine yield.
type YieldValue struct {
	IsAction  bool
	Layout    *PluginLayout    // non-nil for grid yields (screens, overlay interfaces)
	Primitive PluginPrimitive  // non-nil for single primitive yields (inline interfaces)
	Action    *PluginAction    // non-nil for action yields
}

// PluginAction represents an action yielded by the plugin coroutine.
type PluginAction struct {
	Name   string         // "navigate", "confirm", "toast", "fetch", "request", "commit", "cancel", "quit"
	Params map[string]any // action-specific parameters
}

// PluginLayout represents a grid layout yielded by the plugin coroutine.
type PluginLayout struct {
	Columns []PluginColumn
	Hints   []PluginHint
}

// PluginColumn is a column in the plugin grid layout.
type PluginColumn struct {
	Span  int
	Cells []PluginCell
}

// PluginCell is a cell within a plugin column.
type PluginCell struct {
	Title   string
	Height  float64
	Content PluginPrimitive
}

// PluginHint is a key hint for the statusbar.
type PluginHint struct {
	Key   string
	Label string
}

// BuildEvent creates a Lua event table from Go values.
// This is the Go → Lua direction for sending events to the coroutine.
func BuildEvent(L *lua.LState, eventType string, fields map[string]lua.LValue) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString(eventType))
	for k, v := range fields {
		tbl.RawSetString(k, v)
	}
	return tbl
}

// BuildInitEvent creates the init event for screen coroutines.
func BuildInitEvent(L *lua.LState, width, height int, params map[string]string) *lua.LTable {
	fields := map[string]lua.LValue{
		"protocol_version": lua.LNumber(1),
		"width":            lua.LNumber(width),
		"height":           lua.LNumber(height),
	}

	if len(params) > 0 {
		paramsTbl := L.NewTable()
		for k, v := range params {
			paramsTbl.RawSetString(k, lua.LString(v))
		}
		fields["params"] = paramsTbl
	}

	return BuildEvent(L, "init", fields)
}

// BuildFieldInitEvent creates the init event for field interface coroutines.
func BuildFieldInitEvent(L *lua.LState, width, height int, value string, config map[string]string) *lua.LTable {
	fields := map[string]lua.LValue{
		"protocol_version": lua.LNumber(1),
		"width":            lua.LNumber(width),
		"height":           lua.LNumber(height),
		"value":            lua.LString(value),
	}

	if len(config) > 0 {
		configTbl := L.NewTable()
		for k, v := range config {
			configTbl.RawSetString(k, lua.LString(v))
		}
		fields["config"] = configTbl
	}

	return BuildEvent(L, "init", fields)
}

// BuildKeyEvent creates a key press event table.
func BuildKeyEvent(L *lua.LState, key string) *lua.LTable {
	return BuildEvent(L, "key", map[string]lua.LValue{
		"key": lua.LString(key),
	})
}

// BuildResizeEvent creates a window resize event table.
func BuildResizeEvent(L *lua.LState, width, height int) *lua.LTable {
	return BuildEvent(L, "resize", map[string]lua.LValue{
		"width":  lua.LNumber(width),
		"height": lua.LNumber(height),
	})
}

// BuildDataEvent creates an async data response event table.
func BuildDataEvent(L *lua.LState, id string, ok bool, result lua.LValue, errMsg string) *lua.LTable {
	fields := map[string]lua.LValue{
		"id": lua.LString(id),
		"ok": lua.LBool(ok),
	}
	if ok {
		fields["result"] = result
	} else {
		fields["error"] = lua.LString(errMsg)
	}
	return BuildEvent(L, "data", fields)
}

// BuildDialogEvent creates a dialog response event table.
func BuildDialogEvent(L *lua.LState, accepted bool) *lua.LTable {
	return BuildEvent(L, "dialog", map[string]lua.LValue{
		"accepted": lua.LBool(accepted),
	})
}

// BuildFocusEvent creates a focus change event table.
func BuildFocusEvent(L *lua.LState, panel int) *lua.LTable {
	return BuildEvent(L, "focus", map[string]lua.LValue{
		"panel": lua.LNumber(panel),
	})
}

// ParseAction parses an action yield table into a PluginAction.
func ParseAction(tbl *lua.LTable) (*PluginAction, error) {
	actionVal := tbl.RawGetString("action")
	actionStr, ok := actionVal.(lua.LString)
	if !ok {
		return nil, fmt.Errorf("action field is not a string")
	}

	name := string(actionStr)
	validActions := map[string]bool{
		"navigate": true, "confirm": true, "toast": true,
		"fetch": true, "request": true, "commit": true,
		"cancel": true, "quit": true,
	}
	if !validActions[name] {
		return nil, fmt.Errorf("unknown action %q", name)
	}

	params := make(map[string]any)
	tbl.ForEach(func(k, v lua.LValue) {
		key, ok := k.(lua.LString)
		if !ok || string(key) == "action" {
			return
		}
		params[string(key)] = luaValueToGo(v)
	})

	return &PluginAction{Name: name, Params: params}, nil
}

// luaValueToGo converts a Lua value to a Go value for action params.
func luaValueToGo(v lua.LValue) any {
	switch val := v.(type) {
	case lua.LBool:
		return bool(val)
	case lua.LNumber:
		return float64(val)
	case lua.LString:
		return string(val)
	case *lua.LTable:
		return luaTableToMap(val)
	case *lua.LNilType:
		return nil
	default:
		return v.String()
	}
}

// luaTableToMap converts a Lua table to a Go map or slice.
// If all keys are sequential integers starting from 1, returns a slice.
// Otherwise returns a map.
func luaTableToMap(tbl *lua.LTable) any {
	// Check if it's an array (sequential integer keys from 1).
	maxN := tbl.MaxN()
	if maxN > 0 {
		isArray := true
		count := 0
		tbl.ForEach(func(k, v lua.LValue) {
			count++
		})
		if count == maxN {
			arr := make([]any, 0, maxN)
			for i := 1; i <= maxN; i++ {
				arr = append(arr, luaValueToGo(tbl.RawGetInt(i)))
			}
			return arr
		}
		isArray = false
		_ = isArray
	}

	m := make(map[string]any)
	tbl.ForEach(func(k, v lua.LValue) {
		switch key := k.(type) {
		case lua.LString:
			m[string(key)] = luaValueToGo(v)
		case lua.LNumber:
			m[fmt.Sprintf("%g", float64(key))] = luaValueToGo(v)
		}
	})
	return m
}

// KeyString converts a tea.KeyPressMsg-style key name to the string format
// used by the plugin event protocol.
func KeyString(key string) string {
	// Normalize common key names to the protocol's expected format.
	// tea.KeyPressMsg.String() already produces "enter", "esc", "tab",
	// "ctrl+c", etc. which matches our protocol.
	return strings.ToLower(key)
}
