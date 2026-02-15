package plugin

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// testdataDir returns the absolute path to the testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get caller file path")
	}
	return filepath.Join(filepath.Dir(file), "testdata")
}

// newSandboxedState creates a sandboxed LState for testing.
// The caller must defer L.Close().
func newSandboxedState(t *testing.T, cfg SandboxConfig) *lua.LState {
	t.Helper()
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	ApplySandbox(L, cfg)
	return L
}

func TestApplySandbox_BlockedGlobals(t *testing.T) {
	blocked := []string{
		"dofile", "loadfile", "load",
		"rawget", "rawset", "rawequal", "rawlen",
	}

	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	for _, name := range blocked {
		t.Run(name, func(t *testing.T) {
			val := L.GetGlobal(name)
			if val != lua.LNil {
				t.Errorf("expected global %q to be nil, got %s", name, val.Type())
			}
		})
	}
}

func TestApplySandbox_BlockedLibraries(t *testing.T) {
	blocked := []string{"io", "os", "debug", "package"}

	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	for _, name := range blocked {
		t.Run(name, func(t *testing.T) {
			val := L.GetGlobal(name)
			if val != lua.LNil {
				t.Errorf("expected library %q to be nil, got %s", name, val.Type())
			}
		})
	}
}

func TestApplySandbox_AllowedGlobals(t *testing.T) {
	// These globals must be present and callable after sandboxing.
	// Functions are tested by checking they are non-nil.
	// Library tables are tested by checking they are tables.
	allowed := []struct {
		name     string
		wantType lua.LValueType
	}{
		{"print", lua.LTFunction},
		{"type", lua.LTFunction},
		{"tostring", lua.LTFunction},
		{"tonumber", lua.LTFunction},
		{"error", lua.LTFunction},
		{"pcall", lua.LTFunction},
		{"xpcall", lua.LTFunction},
		{"ipairs", lua.LTFunction},
		{"pairs", lua.LTFunction},
		{"next", lua.LTFunction},
		{"select", lua.LTFunction},
		{"unpack", lua.LTFunction},
		{"setmetatable", lua.LTFunction},
		{"getmetatable", lua.LTFunction},
		{"table", lua.LTTable},
		{"string", lua.LTTable},
		{"math", lua.LTTable},
	}

	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	for _, tc := range allowed {
		t.Run(tc.name, func(t *testing.T) {
			val := L.GetGlobal(tc.name)
			if val == lua.LNil {
				t.Fatalf("expected global %q to exist, got nil", tc.name)
			}
			if val.Type() != tc.wantType {
				t.Errorf("expected global %q to be %s, got %s", tc.name, tc.wantType, val.Type())
			}
		})
	}
}

func TestApplySandbox_AllowedGlobalsWork(t *testing.T) {
	// Verify that allowed globals actually function correctly, not just exist.
	tests := []struct {
		name   string
		code   string
		expect string
	}{
		{
			name:   "type",
			code:   `return type(42)`,
			expect: "number",
		},
		{
			name:   "tostring",
			code:   `return tostring(42)`,
			expect: "42",
		},
		{
			name:   "tonumber",
			code:   `return tostring(tonumber("42"))`,
			expect: "42",
		},
		{
			name:   "pcall_success",
			code:   `local ok, v = pcall(function() return 1 end); return tostring(ok)`,
			expect: "true",
		},
		{
			name:   "pcall_failure",
			code:   `local ok, v = pcall(function() error("boom") end); return tostring(ok)`,
			expect: "false",
		},
		{
			name:   "select",
			code:   `return tostring(select("#", 1, 2, 3))`,
			expect: "3",
		},
		{
			name:   "table_insert",
			code:   `local t = {}; table.insert(t, "a"); return t[1]`,
			expect: "a",
		},
		{
			name:   "string_upper",
			code:   `return string.upper("hello")`,
			expect: "HELLO",
		},
		{
			name:   "math_floor",
			code:   `return tostring(math.floor(3.7))`,
			expect: "3",
		},
		{
			name:   "pairs_iteration",
			code:   `local t = {a=1}; for k,v in pairs(t) do return k end`,
			expect: "a",
		},
		{
			name:   "ipairs_iteration",
			code:   `local t = {"x","y"}; local r = ""; for i,v in ipairs(t) do r = r..v end; return r`,
			expect: "xy",
		},
		{
			name:   "next",
			code:   `local t = {a=1}; local k = next(t); return k`,
			expect: "a",
		},
		{
			name:   "unpack",
			code:   `local a, b = unpack({10, 20}); return tostring(a + b)`,
			expect: "30",
		},
		{
			name:   "setmetatable_getmetatable",
			code:   `local t = {}; local mt = {__index = function() return "found" end}; setmetatable(t, mt); local m = getmetatable(t); return tostring(m == mt)`,
			expect: "true",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			L := newSandboxedState(t, SandboxConfig{})
			defer L.Close()

			if err := L.DoString(tc.code); err != nil {
				t.Fatalf("DoString failed: %v", err)
			}
			result := L.Get(-1)
			if result.String() != tc.expect {
				t.Errorf("got %q, want %q", result.String(), tc.expect)
			}
		})
	}
}

func TestApplySandbox_CoroutineControlled(t *testing.T) {
	t.Run("coroutine_disabled", func(t *testing.T) {
		L := newSandboxedState(t, SandboxConfig{AllowCoroutine: false})
		defer L.Close()

		val := L.GetGlobal("coroutine")
		if val != lua.LNil {
			t.Errorf("expected coroutine to be nil when AllowCoroutine is false, got %s", val.Type())
		}
	})

	t.Run("coroutine_enabled", func(t *testing.T) {
		L := newSandboxedState(t, SandboxConfig{AllowCoroutine: true})
		defer L.Close()

		val := L.GetGlobal("coroutine")
		if val == lua.LNil {
			t.Fatal("expected coroutine to exist when AllowCoroutine is true")
		}
		if val.Type() != lua.LTTable {
			t.Errorf("expected coroutine to be a table, got %s", val.Type())
		}

		// Verify coroutine actually works.
		code := `
			local co = coroutine.create(function() coroutine.yield(42) end)
			local ok, val = coroutine.resume(co)
			return tostring(val)
		`
		if err := L.DoString(code); err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
		result := L.Get(-1)
		if result.String() != "42" {
			t.Errorf("got %q, want %q", result.String(), "42")
		}
	})
}

func TestApplySandbox_TimeoutEnforcement(t *testing.T) {
	L := newSandboxedState(t, SandboxConfig{ExecTimeout: 100 * time.Millisecond})
	defer L.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	L.SetContext(ctx)

	// Infinite loop -- should be killed by context cancellation.
	err := L.DoString(`while true do end`)
	if err == nil {
		t.Fatal("expected timeout error from infinite loop, got nil")
	}

	// Verify the error is context-related.
	if err.Error() != context.DeadlineExceeded.Error() {
		// gopher-lua wraps the context error in its own format.
		// Accept any error -- the key contract is that it does not hang.
		t.Logf("timeout error (acceptable): %v", err)
	}
}

func TestFreezeModule_WritesRejected(t *testing.T) {
	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	// Create a module with a function.
	mod := L.NewTable()
	L.SetField(mod, "greet", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("hello"))
		return 1
	}))
	L.SetGlobal("mymod", mod)

	FreezeModule(L, "mymod")

	// Attempt to overwrite an existing key.
	err := L.DoString(`mymod.greet = nil`)
	if err == nil {
		t.Fatal("expected error when writing to frozen module, got nil")
	}
	if got := err.Error(); !containsAll(got, "cannot modify frozen module", "mymod", "greet") {
		t.Errorf("unexpected error message: %s", got)
	}

	// Attempt to add a new key.
	err = L.DoString(`mymod.new_field = true`)
	if err == nil {
		t.Fatal("expected error when adding to frozen module, got nil")
	}
	if got := err.Error(); !containsAll(got, "cannot modify frozen module", "mymod", "new_field") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestFreezeModule_ReadsWork(t *testing.T) {
	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	// Create a module with a Go-bound function.
	mod := L.NewTable()
	L.SetField(mod, "greet", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		L.Push(lua.LString("hello " + name))
		return 1
	}))
	L.SetGlobal("mymod", mod)

	FreezeModule(L, "mymod")

	// Read should delegate through __index to the backing table.
	if err := L.DoString(`return mymod.greet("world")`); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "hello world" {
		t.Errorf("got %q, want %q", result.String(), "hello world")
	}
}

func TestFreezeModule_GetmetatableReturnsProtected(t *testing.T) {
	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	mod := L.NewTable()
	L.SetField(mod, "noop", L.NewFunction(func(L *lua.LState) int { return 0 }))
	L.SetGlobal("mymod", mod)

	FreezeModule(L, "mymod")

	if err := L.DoString(`return getmetatable(mymod)`); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "protected" {
		t.Errorf("getmetatable returned %q, want %q", result.String(), "protected")
	}
}

func TestFreezeModule_SetmetatableRaisesError(t *testing.T) {
	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	mod := L.NewTable()
	L.SetField(mod, "noop", L.NewFunction(func(L *lua.LState) int { return 0 }))
	L.SetGlobal("mymod", mod)

	FreezeModule(L, "mymod")

	err := L.DoString(`setmetatable(mymod, {})`)
	if err == nil {
		t.Fatal("expected error from setmetatable on frozen module, got nil")
	}
	// gopher-lua returns "cannot change a protected metatable" for __metatable-guarded tables.
	if got := err.Error(); !strings.Contains(got, "protected metatable") {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestFreezeModule_PairsReturnsNothing(t *testing.T) {
	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	mod := L.NewTable()
	L.SetField(mod, "greet", L.NewFunction(func(L *lua.LState) int { return 0 }))
	L.SetGlobal("mymod", mod)

	FreezeModule(L, "mymod")

	// pairs() iterates the proxy table's own keys, which is empty.
	// The count should be 0.
	code := `
		local count = 0
		for k, v in pairs(mymod) do
			count = count + 1
		end
		return tostring(count)
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "0" {
		t.Errorf("pairs() yielded %s items, want 0 (proxy is empty)", result.String())
	}
}

func TestFreezeModule_NilGlobalIsNoOp(t *testing.T) {
	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	// FreezeModule on a nonexistent global should not panic or create artifacts.
	FreezeModule(L, "nonexistent")

	val := L.GetGlobal("nonexistent")
	if val != lua.LNil {
		t.Errorf("expected nonexistent global to remain nil, got %s", val.Type())
	}
}

func TestRegisterPluginRequire_LoadFromLib(t *testing.T) {
	td := testdataDir(t)
	pluginDir := filepath.Join(td, "sandbox_test_plugin")

	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()
	RegisterPluginRequire(L, pluginDir)

	code := `
		local helpers = require("helpers")
		return helpers.greet("world")
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "hello world" {
		t.Errorf("got %q, want %q", result.String(), "hello world")
	}
}

func TestRegisterPluginRequire_RejectsPathTraversal(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
	}{
		{"dotdot", `require("..")`},
		{"dotdot_prefix", `require("../evil")`},
		{"forward_slash", `require("lib/helpers")`},
		{"backslash", `require("lib\\helpers")`},
		{"embedded_dotdot", `require("foo..bar")`},
	}

	td := testdataDir(t)
	pluginDir := filepath.Join(td, "sandbox_test_plugin")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			L := newSandboxedState(t, SandboxConfig{})
			defer L.Close()
			RegisterPluginRequire(L, pluginDir)

			err := L.DoString(tc.moduleName)
			if err == nil {
				t.Fatal("expected error for path traversal attempt, got nil")
			}
			if got := err.Error(); !strings.Contains(got, "must be a simple name") {
				t.Errorf("unexpected error: %s", got)
			}
		})
	}
}

func TestRegisterPluginRequire_RejectsNonexistentModule(t *testing.T) {
	td := testdataDir(t)
	pluginDir := filepath.Join(td, "sandbox_test_plugin")

	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()
	RegisterPluginRequire(L, pluginDir)

	err := L.DoString(`require("nonexistent")`)
	if err == nil {
		t.Fatal("expected error for nonexistent module, got nil")
	}
	if got := err.Error(); !strings.Contains(got, "not found") {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestRegisterPluginRequire_CachesModules(t *testing.T) {
	td := testdataDir(t)
	pluginDir := filepath.Join(td, "sandbox_test_plugin")

	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()
	RegisterPluginRequire(L, pluginDir)

	// Load the module twice and verify both references point to the same table.
	code := `
		local h1 = require("helpers")
		local h2 = require("helpers")
		return tostring(h1 == h2)
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "true" {
		t.Errorf("require caching failed: h1 == h2 returned %s, want true", result.String())
	}
}

func TestRegisterPluginRequire_ReplacesNativeRequire(t *testing.T) {
	// gopher-lua's OpenBase registers a native require (and module) function
	// as part of baseFuncs. RegisterPluginRequire replaces it with the
	// sandboxed version. Verify the replacement actually happens.
	td := testdataDir(t)
	pluginDir := filepath.Join(td, "sandbox_test_plugin")

	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	// Before RegisterPluginRequire, the native require from OpenBase exists.
	nativeRequire := L.GetGlobal("require")
	if nativeRequire == lua.LNil {
		t.Fatal("expected native require to exist after OpenBase")
	}

	RegisterPluginRequire(L, pluginDir)

	// After RegisterPluginRequire, require should be a different function.
	sandboxedRequire := L.GetGlobal("require")
	if sandboxedRequire == lua.LNil {
		t.Fatal("expected sandboxed require to exist after RegisterPluginRequire")
	}

	// The sandboxed require should reject path traversal (native would not).
	err := L.DoString(`require("../evil")`)
	if err == nil {
		t.Fatal("expected sandboxed require to reject path traversal")
	}
	if got := err.Error(); !strings.Contains(got, "must be a simple name") {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestFreezeModule_MultipleModulesIndependent(t *testing.T) {
	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	// Create two modules.
	mod1 := L.NewTable()
	L.SetField(mod1, "value", lua.LString("one"))
	L.SetGlobal("mod1", mod1)

	mod2 := L.NewTable()
	L.SetField(mod2, "value", lua.LString("two"))
	L.SetGlobal("mod2", mod2)

	FreezeModule(L, "mod1")
	FreezeModule(L, "mod2")

	// Verify both modules are independently readable.
	code := `return mod1.value .. ":" .. mod2.value`
	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "one:two" {
		t.Errorf("got %q, want %q", result.String(), "one:two")
	}

	// Verify writes to either module fail independently.
	err := L.DoString(`mod1.value = "hacked"`)
	if err == nil {
		t.Fatal("expected error writing to frozen mod1")
	}
	err = L.DoString(`mod2.value = "hacked"`)
	if err == nil {
		t.Fatal("expected error writing to frozen mod2")
	}
}

func TestFreezeModule_GoFunctionCallableAfterFreeze(t *testing.T) {
	L := newSandboxedState(t, SandboxConfig{})
	defer L.Close()

	// Simulate the pattern used by db_api.go: register Go-bound functions,
	// then freeze the module. Verify the functions remain callable.
	mod := L.NewTable()
	L.SetField(mod, "query", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("query_result"))
		return 1
	}))
	L.SetField(mod, "insert", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	L.SetGlobal("db", mod)

	FreezeModule(L, "db")

	// Verify function calls work through the proxy.
	code := `
		local r1 = db.query()
		local r2 = db.insert()
		return r1 .. ":" .. tostring(r2)
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "query_result:true" {
		t.Errorf("got %q, want %q", result.String(), "query_result:true")
	}
}

// containsAll checks that s contains all of the given substrings.
func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

