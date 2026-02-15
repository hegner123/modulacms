package plugin

import (
	"fmt"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestRegisterHTTPAPI(t *testing.T) {
	t.Run("http global is a table with handle and use functions", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		httpGlobal := L.GetGlobal("http")
		if httpGlobal.Type() != lua.LTTable {
			t.Fatalf("expected http global to be LTTable, got %s", httpGlobal.Type())
		}

		httpTable := httpGlobal.(*lua.LTable)
		for _, name := range []string{"handle", "use"} {
			fn := L.GetField(httpTable, name)
			if fn.Type() != lua.LTFunction {
				t.Errorf("expected http.%s to be LTFunction, got %s", name, fn.Type())
			}
			luaFn, ok := fn.(*lua.LFunction)
			if !ok {
				t.Errorf("expected http.%s to be *lua.LFunction, got %T", name, fn)
				continue
			}
			if !luaFn.IsG {
				t.Errorf("expected http.%s to be Go-bound (IsG == true), got false", name)
			}
		}
	})

	t.Run("hidden tables are created as globals", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		for _, name := range []string{"__http_handlers", "__http_route_meta", "__http_middleware"} {
			val := L.GetGlobal(name)
			if val.Type() != lua.LTTable {
				t.Errorf("expected %s to be LTTable, got %s", name, val.Type())
			}
		}
	})
}

func TestHTTPHandle(t *testing.T) {
	t.Run("stores handler in __http_handlers", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.handle("GET", "/tasks", function(req)
				return {status = 200}
			end)
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		handlers := L.GetGlobal("__http_handlers").(*lua.LTable)
		handler := L.GetField(handlers, "GET /tasks")
		if handler.Type() != lua.LTFunction {
			t.Fatalf("expected handler at 'GET /tasks' to be LTFunction, got %s", handler.Type())
		}
	})

	t.Run("options table parsing with public flag", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.handle("POST", "/webhook", function(req)
				return {status = 200}
			end, {public = true})
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		routeMeta := L.GetGlobal("__http_route_meta").(*lua.LTable)
		meta := L.GetField(routeMeta, "POST /webhook")
		if meta.Type() != lua.LTTable {
			t.Fatalf("expected route meta at 'POST /webhook' to be LTTable, got %s", meta.Type())
		}

		metaTbl := meta.(*lua.LTable)
		publicVal := L.GetField(metaTbl, "public")
		if publicVal != lua.LTrue {
			t.Errorf("expected public = true, got %v", publicVal)
		}
	})

	t.Run("default public flag is false", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.handle("GET", "/tasks", function(req)
				return {status = 200}
			end)
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		routeMeta := L.GetGlobal("__http_route_meta").(*lua.LTable)
		meta := L.GetField(routeMeta, "GET /tasks").(*lua.LTable)
		publicVal := L.GetField(meta, "public")
		if publicVal != lua.LFalse {
			t.Errorf("expected public = false, got %v", publicVal)
		}
	})

	t.Run("multiple routes stored with correct keys", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.handle("GET", "/tasks", function(req) return {status = 200} end)
			http.handle("POST", "/tasks", function(req) return {status = 201} end)
			http.handle("GET", "/tasks/{id}", function(req) return {status = 200} end)
			http.handle("DELETE", "/tasks/{id}", function(req) return {status = 204} end)
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		handlers := L.GetGlobal("__http_handlers").(*lua.LTable)
		expectedKeys := []string{"GET /tasks", "POST /tasks", "GET /tasks/{id}", "DELETE /tasks/{id}"}
		for _, key := range expectedKeys {
			handler := L.GetField(handlers, key)
			if handler.Type() != lua.LTFunction {
				t.Errorf("expected handler at %q to be LTFunction, got %s", key, handler.Type())
			}
		}
	})

	t.Run("duplicate route detection", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.handle("GET", "/tasks", function(req) return {status = 200} end)
		`)
		if err != nil {
			t.Fatalf("first registration should succeed, got: %v", err)
		}

		// Second registration of the same method+path should fail.
		err = L.DoString(`
			http.handle("GET", "/tasks", function(req) return {status = 200} end)
		`)
		if err == nil {
			t.Fatal("expected error for duplicate route, got nil")
		}
		if !strings.Contains(err.Error(), "duplicate route") {
			t.Errorf("expected error to mention 'duplicate route', got: %v", err)
		}
		if !strings.Contains(err.Error(), "GET /tasks") {
			t.Errorf("expected error to mention 'GET /tasks', got: %v", err)
		}
	})

	t.Run("same path different methods is not a duplicate", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.handle("GET", "/tasks", function(req) return {status = 200} end)
			http.handle("POST", "/tasks", function(req) return {status = 201} end)
		`)
		if err != nil {
			t.Fatalf("expected no error for different methods on same path, got: %v", err)
		}
	})

	t.Run("invalid method rejection", func(t *testing.T) {
		methods := []string{"INVALID", "OPTIONS", "HEAD", "CONNECT", "TRACE", "get", "post", ""}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				L := lua.NewState()
				defer L.Close()
				RegisterHTTPAPI(L, "test_plugin")

				err := L.DoString(fmt.Sprintf(`
					http.handle(%q, "/tasks", function(req) return {status = 200} end)
				`, method))
				if err == nil {
					t.Fatalf("expected error for method %q, got nil", method)
				}
				if !strings.Contains(err.Error(), "invalid HTTP method") {
					t.Errorf("expected error to mention 'invalid HTTP method', got: %v", err)
				}
			})
		}
	})

	t.Run("valid methods accepted", func(t *testing.T) {
		validMethodList := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range validMethodList {
			t.Run(method, func(t *testing.T) {
				L := lua.NewState()
				defer L.Close()
				RegisterHTTPAPI(L, "test_plugin")

				err := L.DoString(fmt.Sprintf(`
					http.handle(%q, "/tasks", function(req) return {status = 200} end)
				`, method))
				if err != nil {
					t.Fatalf("expected method %q to be accepted, got: %v", method, err)
				}
			})
		}
	})

	t.Run("invalid path rejection", func(t *testing.T) {
		cases := []struct {
			name string
			path string
		}{
			{name: "path traversal", path: "../escape"},
			{name: "path traversal with slash", path: "/foo/../bar"},
			{name: "query string", path: "/path?query=1"},
			{name: "fragment", path: "/path#section"},
			{name: "missing leading slash", path: "tasks"},
			{name: "empty string", path: ""},
			{name: "exceeds 256 chars", path: "/" + strings.Repeat("a", 256)},
			{name: "space character", path: "/path with spaces"},
			{name: "backslash", path: "/path\\escape"},
			{name: "semicolon", path: "/path;injection"},
			{name: "percent encoding", path: "/path%20encoded"},
			{name: "angle bracket", path: "/path<script>"},
			{name: "pipe character", path: "/path|other"},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				L := lua.NewState()
				defer L.Close()
				RegisterHTTPAPI(L, "test_plugin")

				// Use CallByParam with Protect: true to catch the error without
				// corrupting the VM state for subsequent checks.
				httpTable := L.GetGlobal("http").(*lua.LTable)
				handleFn := L.GetField(httpTable, "handle").(*lua.LFunction)

				handler := L.NewFunction(func(L *lua.LState) int { return 0 })

				err := L.CallByParam(lua.P{
					Fn:      handleFn,
					NRet:    0,
					Protect: true,
				}, lua.LString("GET"), lua.LString(tc.path), handler)

				if err == nil {
					t.Fatalf("expected error for path %q, got nil", tc.path)
				}
			})
		}
	})

	t.Run("valid paths accepted", func(t *testing.T) {
		paths := []string{
			"/tasks",
			"/tasks/{id}",
			"/api/v1/items",
			"/items/{id}/sub-items",
			"/webhook.endpoint",
			"/a",
			"/" + strings.Repeat("a", 255), // exactly 256 chars total, at the limit
		}

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				L := lua.NewState()
				defer L.Close()
				RegisterHTTPAPI(L, "test_plugin")

				err := L.DoString(fmt.Sprintf(`
					http.handle("GET", %q, function(req) return {status = 200} end)
				`, path))
				if err != nil {
					t.Fatalf("expected path %q to be accepted, got: %v", path, err)
				}
			})
		}
	})

	t.Run("route count limit", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		// Register exactly MaxRoutesPerPlugin routes (should all succeed).
		for i := range MaxRoutesPerPlugin {
			err := L.DoString(fmt.Sprintf(`
				http.handle("GET", "/route%d", function(req) return {status = 200} end)
			`, i))
			if err != nil {
				t.Fatalf("route %d (of %d) should succeed, got: %v", i, MaxRoutesPerPlugin, err)
			}
		}

		// The 51st route should fail.
		err := L.DoString(`
			http.handle("GET", "/one_too_many", function(req) return {status = 200} end)
		`)
		if err == nil {
			t.Fatal("expected error for exceeding route limit, got nil")
		}
		if !strings.Contains(err.Error(), "maximum route limit") {
			t.Errorf("expected error to mention 'maximum route limit', got: %v", err)
		}
	})

	t.Run("phase guard rejects handle inside on_init", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		// Simulate the Manager setting the phase flag before on_init.
		reg := L.Get(lua.RegistryIndex).(*lua.LTable)
		L.SetField(reg, "in_init", lua.LTrue)

		err := L.DoString(`
			http.handle("GET", "/should-fail", function(req) return {status = 200} end)
		`)
		if err == nil {
			t.Fatal("expected error for http.handle() inside on_init, got nil")
		}
		if !strings.Contains(err.Error(), "module scope") {
			t.Errorf("expected error to mention 'module scope', got: %v", err)
		}
		if !strings.Contains(err.Error(), "on_init") {
			t.Errorf("expected error to mention 'on_init', got: %v", err)
		}
	})

	t.Run("handle works when phase flag is false", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		// Explicitly set in_init to false (simulates after on_init returns).
		reg := L.Get(lua.RegistryIndex).(*lua.LTable)
		L.SetField(reg, "in_init", lua.LFalse)

		err := L.DoString(`
			http.handle("GET", "/should-work", function(req) return {status = 200} end)
		`)
		if err != nil {
			t.Fatalf("expected no error when in_init is false, got: %v", err)
		}
	})

	t.Run("handle works when phase flag is not set", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		// No phase flag set at all (module scope -- the normal case).
		err := L.DoString(`
			http.handle("GET", "/normal", function(req) return {status = 200} end)
		`)
		if err != nil {
			t.Fatalf("expected no error when in_init is not set, got: %v", err)
		}
	})

	t.Run("options with public false", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.handle("GET", "/private", function(req)
				return {status = 200}
			end, {public = false})
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		routeMeta := L.GetGlobal("__http_route_meta").(*lua.LTable)
		meta := L.GetField(routeMeta, "GET /private").(*lua.LTable)
		publicVal := L.GetField(meta, "public")
		if publicVal != lua.LFalse {
			t.Errorf("expected public = false, got %v", publicVal)
		}
	})

	t.Run("options with empty table defaults to public false", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.handle("GET", "/default", function(req)
				return {status = 200}
			end, {})
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		routeMeta := L.GetGlobal("__http_route_meta").(*lua.LTable)
		meta := L.GetField(routeMeta, "GET /default").(*lua.LTable)
		publicVal := L.GetField(meta, "public")
		if publicVal != lua.LFalse {
			t.Errorf("expected public = false for empty options, got %v", publicVal)
		}
	})
}

func TestHTTPUse(t *testing.T) {
	t.Run("stores middleware in __http_middleware", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.use(function(req)
				return nil
			end)
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		middleware := L.GetGlobal("__http_middleware").(*lua.LTable)
		first := L.RawGetInt(middleware, 1)
		if first.Type() != lua.LTFunction {
			t.Fatalf("expected middleware[1] to be LTFunction, got %s", first.Type())
		}
	})

	t.Run("multiple middleware stored in order", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.use(function(req) req.mw1 = true; return nil end)
			http.use(function(req) req.mw2 = true; return nil end)
			http.use(function(req) req.mw3 = true; return nil end)
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		middleware := L.GetGlobal("__http_middleware").(*lua.LTable)
		if middleware.Len() != 3 {
			t.Fatalf("expected 3 middleware entries, got %d", middleware.Len())
		}

		// Verify all three are functions.
		for i := 1; i <= 3; i++ {
			entry := L.RawGetInt(middleware, i)
			if entry.Type() != lua.LTFunction {
				t.Errorf("expected middleware[%d] to be LTFunction, got %s", i, entry.Type())
			}
		}
	})

	t.Run("phase guard rejects use inside on_init", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		// Simulate the Manager setting the phase flag before on_init.
		reg := L.Get(lua.RegistryIndex).(*lua.LTable)
		L.SetField(reg, "in_init", lua.LTrue)

		err := L.DoString(`
			http.use(function(req) return nil end)
		`)
		if err == nil {
			t.Fatal("expected error for http.use() inside on_init, got nil")
		}
		if !strings.Contains(err.Error(), "module scope") {
			t.Errorf("expected error to mention 'module scope', got: %v", err)
		}
		if !strings.Contains(err.Error(), "on_init") {
			t.Errorf("expected error to mention 'on_init', got: %v", err)
		}
	})

	t.Run("use works when phase flag is not set", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`
			http.use(function(req) return nil end)
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("use rejects non-function argument", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		RegisterHTTPAPI(L, "test_plugin")

		err := L.DoString(`http.use("not a function")`)
		if err == nil {
			t.Fatal("expected error for non-function argument, got nil")
		}

		err = L.DoString(`http.use(42)`)
		if err == nil {
			t.Fatal("expected error for number argument, got nil")
		}

		err = L.DoString(`http.use({})`)
		if err == nil {
			t.Fatal("expected error for table argument, got nil")
		}
	})
}

func TestValidateRoutePath(t *testing.T) {
	t.Run("valid paths", func(t *testing.T) {
		paths := []string{
			"/",
			"/tasks",
			"/tasks/{id}",
			"/api/v1/items",
			"/items/{id}/sub-items",
			"/webhook.endpoint",
			"/a",
			"/ABC/DEF",
			"/path-with-dashes",
			"/path_with_underscores",
			"/path.with.dots",
			"/" + strings.Repeat("x", 255), // 256 chars total -- at the limit
		}

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				if err := validateRoutePath(path); err != nil {
					t.Errorf("expected path %q to be valid, got: %v", path, err)
				}
			})
		}
	})

	t.Run("invalid paths", func(t *testing.T) {
		cases := []struct {
			name string
			path string
		}{
			{name: "empty", path: ""},
			{name: "no leading slash", path: "tasks"},
			{name: "path traversal", path: "/foo/../bar"},
			{name: "path traversal at start", path: "../escape"},
			{name: "double dot only", path: "/.."},
			{name: "query string", path: "/path?a=1"},
			{name: "fragment", path: "/path#sec"},
			{name: "exceeds 256", path: "/" + strings.Repeat("a", 256)},
			{name: "space", path: "/has space"},
			{name: "backslash", path: "/back\\slash"},
			{name: "semicolon", path: "/semi;colon"},
			{name: "percent", path: "/pct%20"},
			{name: "at sign", path: "/at@sign"},
			{name: "exclamation", path: "/bang!"},
			{name: "asterisk", path: "/star*"},
			{name: "plus", path: "/plus+"},
			{name: "equals", path: "/eq=val"},
			{name: "comma", path: "/a,b"},
			{name: "tilde", path: "/~user"},
			{name: "angle bracket", path: "/a<b"},
			{name: "pipe", path: "/a|b"},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if err := validateRoutePath(tc.path); err == nil {
					t.Errorf("expected path %q to be invalid, got nil", tc.path)
				}
			})
		}
	})
}

func TestIsValidPathChar(t *testing.T) {
	t.Run("allowed characters", func(t *testing.T) {
		allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/_{}.-"
		for _, c := range []byte(allowed) {
			if !isValidPathChar(c) {
				t.Errorf("expected %q to be a valid path char", string(c))
			}
		}
	})

	t.Run("disallowed characters", func(t *testing.T) {
		disallowed := " !@#$%^&*()+=[]|\\:;'\"<>,?~`"
		for _, c := range []byte(disallowed) {
			if isValidPathChar(c) {
				t.Errorf("expected %q to NOT be a valid path char", string(c))
			}
		}
	})
}
