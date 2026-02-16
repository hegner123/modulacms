package plugin

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/audited"
	lua "github.com/yuin/gopher-lua"
)

func newHooksTestVM() *lua.LState {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	ApplySandbox(L, SandboxConfig{AllowCoroutine: true})
	RegisterLogAPI(L, "test_hooks")
	RegisterDBAPI(L, NewDatabaseAPI(nil, "test_hooks", 0, 1000, nil))
	RegisterHTTPAPI(L, "test_hooks")
	RegisterHooksAPI(L, "test_hooks")
	FreezeModule(L, "db")
	FreezeModule(L, "log")
	FreezeModule(L, "http")
	FreezeModule(L, "hooks")
	return L
}

func TestRegisterHooksAPI(t *testing.T) {
	t.Run("hooks global is a table with on function", func(t *testing.T) {
		L := newHooksTestVM()
		defer L.Close()

		hooksVal := L.GetGlobal("hooks")
		if hooksVal == lua.LNil {
			t.Fatal("hooks global is nil")
		}

		// hooks is frozen, so it's a proxy. Reading through __index should work.
		onVal := L.GetField(hooksVal, "on")
		fn, ok := onVal.(*lua.LFunction)
		if !ok {
			t.Fatalf("hooks.on is not a function: %T", onVal)
		}
		if !fn.IsG {
			t.Fatal("hooks.on is not a Go-bound function")
		}
	})

	t.Run("hidden tables created as globals", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test")

		handlersVal := L.GetGlobal("__hook_handlers")
		if handlersVal == lua.LNil {
			t.Fatal("__hook_handlers not created")
		}
		if _, ok := handlersVal.(*lua.LTable); !ok {
			t.Fatalf("__hook_handlers is not a table: %T", handlersVal)
		}

		pendingVal := L.GetGlobal("__hook_pending")
		if pendingVal == lua.LNil {
			t.Fatal("__hook_pending not created")
		}
		if _, ok := pendingVal.(*lua.LTable); !ok {
			t.Fatalf("__hook_pending is not a table: %T", pendingVal)
		}
	})
}

func TestHooksOn(t *testing.T) {
	t.Run("valid event and table stores handler and pending", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		err := L.DoString(`
			hooks.on("before_create", "content_data", function(data)
				log.info("hook fired")
			end)
		`)
		if err != nil {
			t.Fatalf("hooks.on() raised error: %s", err)
		}

		hooks := ReadPendingHooks(L)
		if len(hooks) != 1 {
			t.Fatalf("expected 1 pending hook, got %d", len(hooks))
		}
		if hooks[0].Event != audited.HookBeforeCreate {
			t.Errorf("event = %q, want %q", hooks[0].Event, audited.HookBeforeCreate)
		}
		if hooks[0].Table != "content_data" {
			t.Errorf("table = %q, want %q", hooks[0].Table, "content_data")
		}
		if hooks[0].Priority != 100 {
			t.Errorf("priority = %d, want 100", hooks[0].Priority)
		}
		if hooks[0].IsWildcard {
			t.Error("is_wildcard should be false")
		}
	})

	t.Run("wildcard table sets is_wildcard", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		err := L.DoString(`
			hooks.on("after_update", "*", function(data) end)
		`)
		if err != nil {
			t.Fatalf("hooks.on() raised error: %s", err)
		}

		hooks := ReadPendingHooks(L)
		if len(hooks) != 1 {
			t.Fatalf("expected 1 pending hook, got %d", len(hooks))
		}
		if !hooks[0].IsWildcard {
			t.Error("is_wildcard should be true for * table")
		}
		if hooks[0].Table != "*" {
			t.Errorf("table = %q, want *", hooks[0].Table)
		}
	})

	t.Run("custom priority clamping", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		err := L.DoString(`
			hooks.on("before_create", "t1", function() end, { priority = 0 })
			hooks.on("before_create", "t2", function() end, { priority = 2000 })
			hooks.on("before_create", "t3", function() end, { priority = 500 })
		`)
		if err != nil {
			t.Fatalf("hooks.on() raised error: %s", err)
		}

		hooks := ReadPendingHooks(L)
		if len(hooks) != 3 {
			t.Fatalf("expected 3 hooks, got %d", len(hooks))
		}
		if hooks[0].Priority != 1 {
			t.Errorf("priority 0 should clamp to 1, got %d", hooks[0].Priority)
		}
		if hooks[1].Priority != 1000 {
			t.Errorf("priority 2000 should clamp to 1000, got %d", hooks[1].Priority)
		}
		if hooks[2].Priority != 500 {
			t.Errorf("priority 500 should stay 500, got %d", hooks[2].Priority)
		}
	})

	t.Run("invalid event rejected", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		err := L.DoString(`
			hooks.on("invalid_event", "content_data", function() end)
		`)
		if err == nil {
			t.Fatal("expected error for invalid event")
		}
	})

	t.Run("empty table name rejected", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		err := L.DoString(`
			hooks.on("before_create", "", function() end)
		`)
		if err == nil {
			t.Fatal("expected error for empty table name")
		}
	})

	t.Run("phase guard rejects inside on_init", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		// Set phase flag.
		regTbl := L.Get(lua.RegistryIndex).(*lua.LTable)
		L.SetField(regTbl, "in_init", lua.LTrue)

		err := L.DoString(`
			hooks.on("before_create", "content_data", function() end)
		`)
		if err == nil {
			t.Fatal("expected error for hooks.on inside on_init")
		}
	})

	t.Run("hook count limit enforced", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		// Register MaxHooksPerPlugin hooks (should succeed).
		for i := range MaxHooksPerPlugin {
			err := L.DoString(`hooks.on("before_create", "t` + string(rune('a'+i%26)) + `", function() end)`)
			if err != nil {
				t.Fatalf("hook %d registration failed: %s", i, err)
			}
		}

		// The next one should fail.
		err := L.DoString(`hooks.on("before_create", "overflow", function() end)`)
		if err == nil {
			t.Fatal("expected error for exceeding hook limit")
		}
	})

	t.Run("multiple hooks stored in order", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		err := L.DoString(`
			hooks.on("before_create", "content_data", function() end)
			hooks.on("after_create", "content_data", function() end)
			hooks.on("before_update", "*", function() end)
		`)
		if err != nil {
			t.Fatalf("hooks.on() raised error: %s", err)
		}

		hooks := ReadPendingHooks(L)
		if len(hooks) != 3 {
			t.Fatalf("expected 3 hooks, got %d", len(hooks))
		}

		if hooks[0].Event != audited.HookBeforeCreate {
			t.Errorf("hook[0].Event = %q, want before_create", hooks[0].Event)
		}
		if hooks[1].Event != audited.HookAfterCreate {
			t.Errorf("hook[1].Event = %q, want after_create", hooks[1].Event)
		}
		if hooks[2].Event != audited.HookBeforeUpdate {
			t.Errorf("hook[2].Event = %q, want before_update", hooks[2].Event)
		}
		if !hooks[2].IsWildcard {
			t.Error("hook[2] should be wildcard")
		}
	})

	t.Run("all valid events accepted", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		for eventStr := range audited.ValidHookEvents {
			err := L.DoString(`hooks.on("` + eventStr + `", "t", function() end)`)
			if err != nil {
				t.Errorf("event %q rejected: %s", eventStr, err)
			}
		}

		hooks := ReadPendingHooks(L)
		if len(hooks) != len(audited.ValidHookEvents) {
			t.Errorf("expected %d hooks, got %d", len(audited.ValidHookEvents), len(hooks))
		}
	})
}

func TestReadPendingHooks(t *testing.T) {
	t.Run("no hooks returns nil", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		hooks := ReadPendingHooks(L)
		if hooks != nil {
			t.Errorf("expected nil, got %v", hooks)
		}
	})

	t.Run("handler key matches between pending and handlers", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		RegisterHooksAPI(L, "test_plugin")

		err := L.DoString(`
			hooks.on("before_delete", "users", function() end)
		`)
		if err != nil {
			t.Fatalf("hooks.on() raised error: %s", err)
		}

		hooks := ReadPendingHooks(L)
		if len(hooks) != 1 {
			t.Fatalf("expected 1 hook, got %d", len(hooks))
		}

		// Verify the handler key exists in __hook_handlers.
		handlersVal := L.GetGlobal("__hook_handlers")
		handlersTbl, ok := handlersVal.(*lua.LTable)
		if !ok {
			t.Fatal("__hook_handlers is not a table")
		}

		handlerFn := L.GetField(handlersTbl, hooks[0].HandlerKey)
		if _, ok := handlerFn.(*lua.LFunction); !ok {
			t.Errorf("handler key %q not found in __hook_handlers", hooks[0].HandlerKey)
		}
	})
}
