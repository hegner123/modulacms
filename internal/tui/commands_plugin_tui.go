package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/plugin"
	lua "github.com/yuin/gopher-lua"
)

// PluginScreenSetupCmd creates a command that checks out a VM, loads the screen
// function, creates a CoroutineBridge, and sends a PluginScreenInitMsg to the
// PluginTUIScreen for coroutine initialization.
func PluginScreenSetupCmd(pluginName, screenName string, params map[string]string, width, height int, mgr *plugin.Manager) tea.Cmd {
	return func() tea.Msg {
		if mgr == nil {
			return PluginScreenErrorMsg{
				PluginName: pluginName,
				ScreenName: screenName,
				Error:      "plugin manager not available",
			}
		}

		inst := mgr.GetPlugin(pluginName)
		if inst == nil {
			return PluginScreenErrorMsg{
				PluginName: pluginName,
				ScreenName: screenName,
				Error:      fmt.Sprintf("plugin %q not found", pluginName),
			}
		}

		if inst.State != plugin.StateRunning {
			return PluginScreenErrorMsg{
				PluginName: pluginName,
				ScreenName: screenName,
				Error:      fmt.Sprintf("plugin %q is %s (not running)", pluginName, inst.State),
			}
		}

		// Check out a VM from the pool.
		// Use a generous timeout for UI VM checkout since the user is waiting.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		L, err := inst.Pool.Get(ctx)
		if err != nil {
			return PluginScreenErrorMsg{
				PluginName: pluginName,
				ScreenName: screenName,
				Error:      fmt.Sprintf("plugin %q: VM pool exhausted", pluginName),
			}
		}

		// Load the screen file.
		screenPath := filepath.Join(inst.Dir, "screens", screenName+".lua")
		if err := L.DoFile(screenPath); err != nil {
			inst.Pool.Put(L)
			return PluginScreenErrorMsg{
				PluginName: pluginName,
				ScreenName: screenName,
				Error:      fmt.Sprintf("loading screen %q: %s", screenName, err),
			}
		}

		// Extract the screen function.
		fn := L.GetGlobal("screen")
		luaFn, ok := fn.(*lua.LFunction)
		if !ok {
			inst.Pool.Put(L)
			return PluginScreenErrorMsg{
				PluginName: pluginName,
				ScreenName: screenName,
				Error:      fmt.Sprintf("screen %q: 'screen' function not found", screenName),
			}
		}

		bridge := plugin.NewCoroutineBridge(inst, L, luaFn)

		return PluginScreenInitMsg{
			Bridge: bridge,
			L:      L,
			Width:  width,
			Height: height,
			Params: params,
		}
	}
}

// PluginRequestCmd executes an outbound HTTP request via the RequestEngine
// and returns the result as a PluginDataMsg.
func PluginRequestCmd(id string, action *plugin.PluginAction, mgr *plugin.Manager, pluginName string) tea.Cmd {
	return func() tea.Msg {
		if mgr == nil {
			return PluginDataMsg{ID: id, OK: false, Error: "plugin manager not available"}
		}

		re := mgr.RequestEngine()
		if re == nil {
			return PluginDataMsg{ID: id, OK: false, Error: "request engine not available"}
		}

		method, _ := action.Params["method"].(string)
		if method == "" {
			method = "GET"
		}
		urlStr, _ := action.Params["url"].(string)
		if urlStr == "" {
			return PluginDataMsg{ID: id, OK: false, Error: "request missing 'url'"}
		}

		opts := plugin.OutboundRequestOpts{
			ParseJSON: true,
		}

		// Extract optional headers.
		if hdrs, ok := action.Params["headers"].(map[string]any); ok {
			opts.Headers = make(map[string]string, len(hdrs))
			for k, v := range hdrs {
				opts.Headers[k] = fmt.Sprintf("%v", v)
			}
		}

		// Extract optional body.
		if body, ok := action.Params["body"].(string); ok {
			opts.Body = body
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := re.Execute(ctx, pluginName, method, urlStr, opts)
		if err != nil {
			return PluginDataMsg{ID: id, OK: false, Error: err.Error()}
		}

		// Convert Go map to a simple representation for the data event.
		// The PluginTUIScreen will need to convert this to a Lua table
		// when building the data event for the coroutine.
		return PluginDataMsg{
			ID:     id,
			OK:     true,
			Result: goMapToLuaTable(result),
		}
	}
}

// goMapToLuaTable converts a Go map to a Lua table. This is used for
// converting request results to Lua values for the data event.
// Note: this creates tables on a nil LState which is not ideal,
// but gopher-lua LTable is just a Go struct — no VM needed for creation.
func goMapToLuaTable(m map[string]any) *lua.LTable {
	tbl := &lua.LTable{}
	for k, v := range m {
		tbl.RawSetString(k, goValueToLua(v))
	}
	return tbl
}

func goValueToLua(v any) lua.LValue {
	if v == nil {
		return lua.LNil
	}
	switch val := v.(type) {
	case bool:
		return lua.LBool(val)
	case float64:
		return lua.LNumber(val)
	case int:
		return lua.LNumber(float64(val))
	case int64:
		return lua.LNumber(float64(val))
	case string:
		return lua.LString(val)
	case map[string]any:
		return goMapToLuaTable(val)
	case []any:
		tbl := &lua.LTable{}
		for i, item := range val {
			tbl.RawSetInt(i+1, goValueToLua(item))
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", val))
	}
}
