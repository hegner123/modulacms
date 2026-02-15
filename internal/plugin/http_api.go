package plugin

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// MaxRoutesPerPlugin is the maximum number of HTTP routes a single plugin can
// register via http.handle(). Exceeding this limit raises a Lua error.
const MaxRoutesPerPlugin = 50

// validMethods is the allowlist of HTTP methods accepted by http.handle().
var validMethods = map[string]bool{
	"GET":    true,
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
	"PATCH":  true,
}

// RegisterHTTPAPI creates an "http" global Lua table with two Go-bound functions:
// handle(method, path, handler [, options]) and use(middleware_function).
//
// It also creates three hidden global tables used by the HTTPBridge to read
// registered routes at load time:
//   - __http_handlers:    "METHOD /path" -> handler LFunction
//   - __http_route_meta:  "METHOD /path" -> {public = bool}
//   - __http_middleware:   ordered array of middleware LFunctions
//
// These tables are part of the global snapshot taken by SnapshotGlobals and
// must NOT be modified after snapshot. The bridge reads them on every request
// dispatch to look up the correct handler for the checked-out VM.
func RegisterHTTPAPI(L *lua.LState, pluginName string) {
	// Create the hidden state tables that http.handle() and http.use() write to.
	handlers := L.NewTable()
	routeMeta := L.NewTable()
	middleware := L.NewTable()

	L.SetGlobal("__http_handlers", handlers)
	L.SetGlobal("__http_route_meta", routeMeta)
	L.SetGlobal("__http_middleware", middleware)

	// Create the public http module table.
	httpTable := L.NewTable()

	httpTable.RawSetString("handle", L.NewFunction(httpHandleFn(L, pluginName, handlers, routeMeta)))
	httpTable.RawSetString("use", L.NewFunction(httpUseFn(L, middleware)))

	L.SetGlobal("http", httpTable)
}

// httpHandleFn returns the Go-bound function for http.handle(method, path, handler [, options]).
//
// Validation order:
//  1. Method against allowlist
//  2. Path via validateRoutePath
//  3. Phase guard (rejects calls inside on_init)
//  4. Duplicate detection
//  5. Route count limit
//  6. Options table parsing
//  7. Store handler and metadata
func httpHandleFn(L *lua.LState, pluginName string, handlers *lua.LTable, routeMeta *lua.LTable) lua.LGFunction {
	return func(L *lua.LState) int {
		method := L.CheckString(1)
		path := L.CheckString(2)
		handler := L.CheckFunction(3)

		// 1. Validate method.
		if !validMethods[method] {
			L.ArgError(1, "invalid HTTP method: must be GET, POST, PUT, DELETE, or PATCH")
			return 0
		}

		// 2. Validate path.
		if err := validateRoutePath(path); err != nil {
			L.ArgError(2, "invalid route path: must start with /, max 256 chars, no '..' or query characters")
			return 0
		}

		// 3. Phase guard: reject calls inside on_init().
		// The Manager sets "in_init" = LTrue in the LState registry before calling
		// on_init(), and clears it after. http.handle() must only be called at
		// module scope (during init.lua execution by the factory), not inside on_init().
		registryTbl := L.Get(lua.RegistryIndex)
		if regTbl, ok := registryTbl.(*lua.LTable); ok {
			inInit := L.GetField(regTbl, "in_init")
			if inInit == lua.LTrue {
				L.RaiseError("http.handle() must be called at module scope, not inside on_init()")
				return 0
			}
		}

		// 4. Duplicate detection.
		key := method + " " + path
		existing := L.GetField(handlers, key)
		if existing != lua.LNil {
			L.ArgError(1, fmt.Sprintf("duplicate route: %s already registered", key))
			return 0
		}

		// 5. Route count limit.
		count := 0
		handlers.ForEach(func(_, _ lua.LValue) {
			count++
		})
		if count >= MaxRoutesPerPlugin {
			L.RaiseError("plugin exceeded maximum route limit (%d)", MaxRoutesPerPlugin)
			return 0
		}

		// 6. Parse options table (4th argument, optional).
		isPublic := false
		if L.GetTop() >= 4 {
			optVal := L.Get(4)
			if optTbl, ok := optVal.(*lua.LTable); ok {
				publicVal := L.GetField(optTbl, "public")
				if publicBool, ok := publicVal.(lua.LBool); ok {
					isPublic = bool(publicBool)
				}
			}
			// If arg 4 is not a table, silently ignore it. The options argument
			// is optional and non-table values are treated as "no options".
		}

		// 7. Store handler and metadata.
		handlers.RawSetString(key, handler)

		meta := L.NewTable()
		meta.RawSetString("public", lua.LBool(isPublic))
		routeMeta.RawSetString(key, meta)

		return 0
	}
}

// httpUseFn returns the Go-bound function for http.use(middleware_function).
//
// Appends the middleware function to the __http_middleware table (ordered array).
// Rejects calls inside on_init() via the same phase guard as http.handle().
func httpUseFn(L *lua.LState, middleware *lua.LTable) lua.LGFunction {
	return func(L *lua.LState) int {
		fn := L.CheckFunction(1)

		// Phase guard: reject calls inside on_init().
		registryTbl := L.Get(lua.RegistryIndex)
		if regTbl, ok := registryTbl.(*lua.LTable); ok {
			inInit := L.GetField(regTbl, "in_init")
			if inInit == lua.LTrue {
				L.RaiseError("http.use() must be called at module scope, not inside on_init()")
				return 0
			}
		}

		// Append to the middleware table. Lua arrays are 1-indexed, so the next
		// index is Len()+1.
		nextIdx := middleware.Len() + 1
		L.RawSetInt(middleware, nextIdx, fn)

		return 0
	}
}

// validateRoutePath checks that a route path is safe for use in ServeMux registration.
// Returns nil if valid, or an error describing the violation.
//
// Validation rules (from PLUGIN_PHASE_2.md):
//   - Must start with "/"
//   - Must not contain ".." (path traversal)
//   - Must not contain "?" (query string)
//   - Must not contain "#" (fragment)
//   - Must not exceed 256 characters
//   - Must contain only allowed characters: a-zA-Z0-9/_{}.-
func validateRoutePath(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("path must not be empty")
	}
	if path[0] != '/' {
		return fmt.Errorf("path must start with /")
	}
	if len(path) > 256 {
		return fmt.Errorf("path exceeds 256 characters")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path must not contain '..'")
	}
	if strings.Contains(path, "?") {
		return fmt.Errorf("path must not contain '?'")
	}
	if strings.Contains(path, "#") {
		return fmt.Errorf("path must not contain '#'")
	}

	for i := range len(path) {
		if !isValidPathChar(path[i]) {
			return fmt.Errorf("path contains invalid character %q at position %d", string(path[i]), i)
		}
	}

	return nil
}

// isValidPathChar returns true if the byte is in the allowed path character set:
// a-z A-Z 0-9 / _ { } . -
//
// Uses byte-level checks (not regex) per CLAUDE.md rules.
func isValidPathChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '/' || c == '_' ||
		c == '{' || c == '}' || c == '.' || c == '-'
}
