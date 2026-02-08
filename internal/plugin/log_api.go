package plugin

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/utility"
	lua "github.com/yuin/gopher-lua"
)

// RegisterLogAPI creates a "log" Lua table with info, warn, error, and debug
// functions bound to the utility.DefaultLogger. The plugin name is included as
// a structured field ("plugin", name) on every log call so operators can trace
// log output back to the originating plugin.
//
// Each Lua function signature: log.<level>(message [, context_table])
//   - message (string, required): the log line
//   - context_table (table, optional): key-value pairs appended as structured args
//
// The roadmap specifies this file as size S -- it is intentionally small and
// delegates all formatting/output to utility.DefaultLogger.
func RegisterLogAPI(L *lua.LState, pluginName string) {
	logTable := L.NewTable()

	logTable.RawSetString("info", L.NewFunction(logFn(pluginName, utility.INFO)))
	logTable.RawSetString("warn", L.NewFunction(logFn(pluginName, utility.WARN)))
	logTable.RawSetString("error", L.NewFunction(logFn(pluginName, utility.ERROR)))
	logTable.RawSetString("debug", L.NewFunction(logFn(pluginName, utility.DEBUG)))

	L.SetGlobal("log", logTable)
}

// logFn returns a Go-bound Lua function that logs at the given level.
// The returned function reads:
//
//	arg 1: message string (required)
//	arg 2: context table (optional) -- keys and values are flattened into the
//	        structured args passed to the logger
//
// utility.Logger has an asymmetric API:
//
//	Debug(msg, args...) / Info(msg, args...)        -- no error param
//	Warn(msg, err, args...) / Error(msg, err, args...) -- error as 2nd param
//
// Since Lua callers have no Go error object, Warn and Error receive nil for the
// error parameter. The plugin name and any context key-value pairs are passed
// through the variadic args.
func logFn(pluginName string, level utility.LogLevel) lua.LGFunction {
	return func(L *lua.LState) int {
		msg := L.CheckString(1)

		// Build structured args: always start with "plugin", <name>
		args := []any{"plugin", pluginName}

		// If a context table is provided, flatten its key-value pairs into args.
		if L.GetTop() >= 2 {
			ctxVal := L.Get(2)
			if ctxTable, ok := ctxVal.(*lua.LTable); ok {
				ctxTable.ForEach(func(key, value lua.LValue) {
					args = append(args, luaValueToString(key), luaValueToString(value))
				})
			}
			// If arg 2 is not a table, silently ignore it. The roadmap specifies
			// the second argument as an optional Lua table; passing a non-table
			// is not an error -- the extra argument is simply ignored. This matches
			// the lenient convention used by most logging libraries.
		}

		logger := utility.DefaultLogger

		switch level {
		case utility.DEBUG:
			logger.Debug(msg, args...)
		case utility.INFO:
			logger.Info(msg, args...)
		case utility.WARN:
			// Warn(msg, err, args...) -- pass nil for err since Lua has no Go error
			logger.Warn(msg, nil, args...)
		case utility.ERROR:
			// Error(msg, err, args...) -- pass nil for err since Lua has no Go error
			logger.Error(msg, nil, args...)
		default:
			logger.Info(msg, args...)
		}

		return 0
	}
}

// luaValueToString converts a Lua value to its string representation for use
// as a structured logging argument. This is intentionally simple -- log context
// values are always rendered as strings for the log line.
func luaValueToString(v lua.LValue) string {
	switch val := v.(type) {
	case *lua.LNilType:
		return "nil"
	case lua.LBool:
		if val {
			return "true"
		}
		return "false"
	case lua.LNumber:
		return fmt.Sprintf("%g", float64(val))
	case lua.LString:
		return string(val)
	default:
		return v.String()
	}
}
