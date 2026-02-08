package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// SandboxConfig controls the Lua VM sandbox behavior.
type SandboxConfig struct {
	// AllowCoroutine enables the coroutine library in sandboxed VMs.
	// Disabled by default because coroutines can be used to circumvent
	// instruction counting in some Lua implementations, though gopher-lua
	// handles this safely. Enable only if plugins need cooperative multitasking.
	AllowCoroutine bool

	// ExecTimeout is the maximum wall-clock duration for a single plugin
	// execution (on_init, HTTP handler, hook callback). Applied via
	// LState.SetContext(ctx) with deadline. Default 5s.
	ExecTimeout time.Duration
}

// strippedGlobals are dangerous base library globals removed after OpenBase.
// These allow arbitrary code loading or bypass metatable protections.
// setmetatable/getmetatable are intentionally kept -- see roadmap sandbox.go section.
// rawlen is included for completeness even though gopher-lua (Lua 5.1) does not
// define it; setting a nonexistent global to nil is harmless.
var strippedGlobals = []string{
	"dofile",
	"loadfile",
	"load",
	"rawget",
	"rawset",
	"rawequal",
	"rawlen",
}

// ApplySandbox configures a Lua VM with a safe stdlib subset.
// It loads only base, table, string, and math libraries (plus coroutine if
// cfg.AllowCoroutine is true). Dangerous globals are stripped after loading.
//
// The io, os, package, debug, and channel libraries are never loaded.
//
// The LState must have been created with SkipOpenLibs: true.
func ApplySandbox(L *lua.LState, cfg SandboxConfig) {
	// Load safe libraries individually using the same pattern as LState.OpenLibs():
	// push the opener function, push the library name, call with 1 arg 0 returns.
	safeLibs := []struct {
		name   string
		opener lua.LGFunction
	}{
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	}

	if cfg.AllowCoroutine {
		safeLibs = append(safeLibs, struct {
			name   string
			opener lua.LGFunction
		}{lua.CoroutineLibName, lua.OpenCoroutine})
	}

	for _, lib := range safeLibs {
		L.Push(L.NewFunction(lib.opener))
		L.Push(lua.LString(lib.name))
		L.Call(1, 0)
	}

	// Strip dangerous globals that were registered by OpenBase.
	for _, name := range strippedGlobals {
		L.SetGlobal(name, lua.LNil)
	}
}

// RegisterPluginRequire replaces the global require with a sandboxed loader.
// Only resolves modules from <pluginDir>/lib/<name>.lua. Module names must be
// simple identifiers -- path traversal (.. / \) is rejected. Loaded modules
// are cached: subsequent require() calls for the same name return the cached value.
//
// Uses L.ArgError for validation failures (bad module name, module not found)
// and L.RaiseError for load failures (syntax error in module file).
func RegisterPluginRequire(L *lua.LState, pluginDir string) {
	loaded := L.NewTable() // cache: module name -> returned value

	L.SetGlobal("require", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)

		// Reject path traversal and absolute paths.
		// Check for "..", "/", and "\" anywhere in the module name.
		if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
			L.ArgError(1, fmt.Sprintf("invalid module name %q: must be a simple name, not a path", name))
			return 0
		}

		// Return cached module if already loaded.
		if cached := L.GetField(loaded, name); cached != lua.LNil {
			L.Push(cached)
			return 1
		}

		path := filepath.Join(pluginDir, "lib", name+".lua")
		if _, err := os.Stat(path); err != nil {
			L.ArgError(1, fmt.Sprintf("module %q not found at %s", name, path))
			return 0
		}

		if err := L.DoFile(path); err != nil {
			L.RaiseError("error loading module %q: %s", name, err.Error())
			return 0
		}

		// Module returns a value (convention: return a table).
		// DoFile pushes the return value onto the stack.
		ret := L.Get(-1)
		L.SetField(loaded, name, ret)
		L.Push(ret)
		return 1
	}))
}

// FreezeModule replaces a global module table with a read-only proxy.
// The real functions are moved to a hidden backing table; the proxy
// delegates reads via __index and rejects writes via __newindex.
// __metatable prevents getmetatable/setmetatable from inspecting or
// replacing the metatable. rawget/rawset are already stripped by the sandbox.
//
// After freezing:
//   - db.query(...) works (reads delegate to backing table via __index)
//   - db.query = nil raises an error (writes intercepted by __newindex)
//   - getmetatable(db) returns "protected" (not the real metatable)
//   - setmetatable(db, {}) raises an error
//   - pairs(db) returns nothing (proxy is empty -- documented DX limitation)
func FreezeModule(L *lua.LState, moduleName string) {
	backing := L.GetGlobal(moduleName)
	if backing == lua.LNil {
		return
	}

	proxy := L.NewTable()
	mt := L.NewTable()

	// All reads delegate to the backing table.
	L.SetField(mt, "__index", backing)

	// All writes raise an error. The proxy is always empty, so gopher-lua's
	// internal RawGetString check always misses, guaranteeing __newindex fires
	// for every write attempt.
	L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(2)
		L.ArgError(2, fmt.Sprintf("cannot modify frozen module %q (key %q)", moduleName, key))
		return 0
	}))

	// Prevent metatable inspection/replacement.
	// Setting __metatable to a non-nil value causes getmetatable() to return
	// that value instead of the real metatable, and causes setmetatable() to error.
	L.SetField(mt, "__metatable", lua.LString("protected"))

	L.SetMetatable(proxy, mt)
	L.SetGlobal(moduleName, proxy)
}
