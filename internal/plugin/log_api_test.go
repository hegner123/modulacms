package plugin

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestRegisterLogAPI(t *testing.T) {
	t.Run("log global is a table after registration", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		RegisterLogAPI(L, "test_plugin")

		logGlobal := L.GetGlobal("log")
		if logGlobal.Type() != lua.LTTable {
			t.Fatalf("expected log global to be LTTable, got %s", logGlobal.Type())
		}
	})

	t.Run("log table has all four level functions", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		RegisterLogAPI(L, "test_plugin")

		logTable := L.GetGlobal("log").(*lua.LTable)

		levels := []string{"info", "warn", "error", "debug"}
		for _, level := range levels {
			fn := L.GetField(logTable, level)
			if fn.Type() != lua.LTFunction {
				t.Errorf("expected log.%s to be LTFunction, got %s", level, fn.Type())
			}
		}
	})

	t.Run("log functions are Go-bound (IsG == true)", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		RegisterLogAPI(L, "test_plugin")

		logTable := L.GetGlobal("log").(*lua.LTable)

		levels := []string{"info", "warn", "error", "debug"}
		for _, level := range levels {
			fn := L.GetField(logTable, level)
			luaFn, ok := fn.(*lua.LFunction)
			if !ok {
				t.Errorf("expected log.%s to be *lua.LFunction, got %T", level, fn)
				continue
			}
			if !luaFn.IsG {
				t.Errorf("expected log.%s to be Go-bound (IsG == true), got false", level)
			}
		}
	})

	t.Run("calling with message string succeeds", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		RegisterLogAPI(L, "test_plugin")

		// Each level should accept a single string argument without error.
		scripts := []string{
			`log.info("info message")`,
			`log.warn("warn message")`,
			`log.error("error message")`,
			`log.debug("debug message")`,
		}

		for _, script := range scripts {
			if err := L.DoString(script); err != nil {
				t.Errorf("expected no error for %q, got: %v", script, err)
			}
		}
	})

	t.Run("calling with message and context table succeeds", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		RegisterLogAPI(L, "test_plugin")

		scripts := []string{
			`log.info("processing", {count = 42, status = "ok"})`,
			`log.warn("rate limit", {threshold = 100})`,
			`log.error("failed", {err = "timeout"})`,
			`log.debug("trace", {request_id = "abc-123"})`,
		}

		for _, script := range scripts {
			if err := L.DoString(script); err != nil {
				t.Errorf("expected no error for %q, got: %v", script, err)
			}
		}
	})

	t.Run("missing message argument raises error", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		RegisterLogAPI(L, "test_plugin")

		// Calling log.info() with no arguments should raise an error because
		// CheckString(1) requires the first argument to be a string.
		calls := []string{
			`log.info()`,
			`log.warn()`,
			`log.error()`,
			`log.debug()`,
		}

		for _, call := range calls {
			if err := L.DoString(call); err == nil {
				t.Errorf("expected error for %q with no arguments, got nil", call)
			}
		}
	})

	t.Run("non-table second argument is silently ignored", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		RegisterLogAPI(L, "test_plugin")

		// Passing a number as the second argument should not error.
		// The roadmap says second arg is optional Lua table; non-table is ignored.
		scripts := []string{
			`log.info("msg", 42)`,
			`log.info("msg", "not a table")`,
			`log.info("msg", true)`,
		}

		for _, script := range scripts {
			if err := L.DoString(script); err != nil {
				t.Errorf("expected no error for %q, got: %v", script, err)
			}
		}
	})

	t.Run("context table with various value types succeeds", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		RegisterLogAPI(L, "test_plugin")

		// Context table can contain strings, numbers, booleans, nil values
		script := `log.info("mixed context", {
			str_val = "hello",
			num_val = 3.14,
			int_val = 42,
			bool_val = true,
			nil_val = nil
		})`

		if err := L.DoString(script); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})
}

func TestLuaValueToString(t *testing.T) {
	cases := []struct {
		name     string
		input    lua.LValue
		expected string
	}{
		{name: "nil", input: lua.LNil, expected: "nil"},
		{name: "true", input: lua.LTrue, expected: "true"},
		{name: "false", input: lua.LFalse, expected: "false"},
		{name: "integer number", input: lua.LNumber(42), expected: "42"},
		{name: "float number", input: lua.LNumber(3.14), expected: "3.14"},
		{name: "zero", input: lua.LNumber(0), expected: "0"},
		{name: "negative", input: lua.LNumber(-7), expected: "-7"},
		{name: "string", input: lua.LString("hello"), expected: "hello"},
		{name: "empty string", input: lua.LString(""), expected: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := luaValueToString(tc.input)
			if result != tc.expected {
				t.Errorf("luaValueToString(%v) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
