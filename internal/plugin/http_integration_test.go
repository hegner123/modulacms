package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/hegner123/modulacms/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// -- Fixture-based integration test helpers --

// fixturePluginsDir is the relative path from the test file to the fixture plugins.
// Go test files run from the package directory, so this path is relative to
// internal/plugin/.
const fixturePluginsDir = "testdata/plugins"

// loadFixturePlugin loads a real fixture plugin from testdata/plugins/ via the
// full VM pipeline (sandbox -> require -> db -> log -> http -> freeze -> DoFile).
// It uses loadPluginForTest from http_bridge_test.go (same package).
//
// After loading, it checks out a VM, registers routes on the bridge, and returns
// the VM to the pool. The plugin is ready for HTTP dispatch after this call.
func loadFixturePlugin(t *testing.T, bridge *HTTPBridge, mgr *Manager, pluginName, version string) {
	t.Helper()

	loadPluginForTest(t, mgr, pluginName, version, fixturePluginsDir)

	inst := mgr.GetPlugin(pluginName)
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("expected %s plugin to be running, got %v", pluginName, inst)
	}

	ctx := context.Background()
	L, getErr := inst.Pool.Get(ctx)
	if getErr != nil {
		t.Fatalf("getting VM for %s route registration: %s", pluginName, getErr)
	}
	L.SetContext(ctx)

	regErr := bridge.RegisterRoutes(ctx, pluginName, version, L)
	inst.Pool.Put(L)
	if regErr != nil {
		t.Fatalf("RegisterRoutes for %s: %s", pluginName, regErr)
	}
}

// setupFixtureBridge creates a Manager and HTTPBridge configured to load
// fixture plugins from testdata/plugins/.
func setupFixtureBridge(t *testing.T) (*HTTPBridge, *Manager, func()) {
	t.Helper()

	conn := newTestDB(t)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       fixturePluginsDir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("creating plugin_routes table: %s", err)
	}

	cleanup := func() {
		bridge.Close(context.Background())
		conn.Close()
	}

	return bridge, mgr, cleanup
}

// approveAllRoutes approves all currently registered routes on the bridge.
func approveAllRoutes(t *testing.T, bridge *HTTPBridge) {
	t.Helper()

	routes := bridge.ListRoutes()
	for _, r := range routes {
		if err := bridge.ApproveRoute(context.Background(), r.PluginName, r.Method, r.Path, "test_admin"); err != nil {
			t.Fatalf("approving route %s %s for %s: %s", r.Method, r.Path, r.PluginName, err)
		}
	}
}

// -- Integration tests --

// TestIntegration_MultiPluginCoexistence loads http_plugin, http_public_plugin,
// and http_params_plugin simultaneously, registers all routes, approves them,
// and verifies each plugin responds correctly to its routes without interference.
func TestIntegration_MultiPluginCoexistence(t *testing.T) {
	bridge, mgr, cleanup := setupFixtureBridge(t)
	defer cleanup()

	// Load three fixture plugins.
	loadFixturePlugin(t, bridge, mgr, "http_plugin", "1.0.0")
	loadFixturePlugin(t, bridge, mgr, "http_public_plugin", "1.0.0")
	loadFixturePlugin(t, bridge, mgr, "http_params_plugin", "1.0.0")

	// Approve all routes.
	approveAllRoutes(t, bridge)

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Verify http_plugin GET /tasks responds.
	t.Run("http_plugin_GET_tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/plugins/http_plugin/tasks", nil)
		req = withAuthenticatedUser(req)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
		}

		var body map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decoding response: %s", err)
		}
		// The tasks key should exist (even if empty due to VM snapshot restore).
		if _, ok := body["tasks"]; !ok {
			t.Error("expected 'tasks' key in response")
		}
	})

	// Verify http_public_plugin GET /status responds (public, no auth needed).
	t.Run("http_public_plugin_GET_status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/plugins/http_public_plugin/status", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
		}

		var body map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decoding response: %s", err)
		}
		if body["status"] != "ok" {
			t.Errorf("status = %v, want %q", body["status"], "ok")
		}
	})

	// Verify http_params_plugin GET /items/{id} responds with auth.
	t.Run("http_params_plugin_GET_items", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/plugins/http_params_plugin/items/xyz789", nil)
		req = withAuthenticatedUser(req)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
		}

		var body map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decoding response: %s", err)
		}
		if body["id"] != "xyz789" {
			t.Errorf("id = %v, want %q", body["id"], "xyz789")
		}
	})

	// Verify route count: http_plugin has 2 (GET+POST /tasks),
	// http_public_plugin has 2 (POST /webhook + GET /status),
	// http_params_plugin has 1 (GET /items/{id}).
	routes := bridge.ListRoutes()
	if len(routes) != 5 {
		t.Errorf("expected 5 total routes across 3 plugins, got %d", len(routes))
	}
}

// TestIntegration_TaskCRUDCycle loads the http_plugin fixture, approves its
// routes, and tests creating a task via POST then retrieving via GET. Because
// VM state is reset per-checkout via snapshot, we verify response shapes rather
// than stateful persistence across requests.
func TestIntegration_TaskCRUDCycle(t *testing.T) {
	bridge, mgr, cleanup := setupFixtureBridge(t)
	defer cleanup()

	loadFixturePlugin(t, bridge, mgr, "http_plugin", "1.0.0")
	approveAllRoutes(t, bridge)

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// POST to create a task.
	t.Run("POST_create_task", func(t *testing.T) {
		payload := map[string]any{"title": "Write integration tests"}
		bodyBytes, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			t.Fatalf("marshaling payload: %s", marshalErr)
		}

		req := httptest.NewRequest("POST", "/api/v1/plugins/http_plugin/tasks",
			bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = withAuthenticatedUser(req)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
		}

		var task map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
			t.Fatalf("decoding response: %s", err)
		}
		if task["title"] != "Write integration tests" {
			t.Errorf("title = %v, want %q", task["title"], "Write integration tests")
		}
		// id should be a number (Lua integer converted to JSON number).
		if _, ok := task["id"]; !ok {
			t.Error("expected 'id' key in task response")
		}
	})

	// GET to list tasks. Due to VM snapshot restore, the task list may be
	// empty (if a different VM serves this request) or contain the task (if
	// the same VM serves both). We verify the response shape, not the contents.
	t.Run("GET_list_tasks_shape", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/plugins/http_plugin/tasks", nil)
		req = withAuthenticatedUser(req)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
		}

		var body map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decoding response: %s", err)
		}
		// The response must have a 'tasks' key (array) and 'processed_by' (from middleware).
		if _, ok := body["tasks"]; !ok {
			t.Error("expected 'tasks' key in response")
		}
		if body["processed_by"] != "http_plugin_middleware" {
			t.Errorf("processed_by = %v, want %q", body["processed_by"], "http_plugin_middleware")
		}
	})
}

// TestIntegration_PathParameters loads the http_params_plugin fixture and
// verifies that path parameters are correctly extracted end-to-end from the
// ServeMux through to the Lua handler's req.params table.
func TestIntegration_PathParameters(t *testing.T) {
	bridge, mgr, cleanup := setupFixtureBridge(t)
	defer cleanup()

	loadFixturePlugin(t, bridge, mgr, "http_params_plugin", "1.0.0")
	approveAllRoutes(t, bridge)

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	tests := []struct {
		name     string
		paramVal string
	}{
		{"simple_id", "abc123"},
		{"numeric_id", "42"},
		{"ulid_style", "01HN8Z3QRXK9VBYN1234ABCDE"},
		{"hyphenated", "my-item-name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/plugins/http_params_plugin/items/"+tt.paramVal, nil)
			req = withAuthenticatedUser(req)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
			}

			var body map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decoding response: %s", err)
			}
			if body["id"] != tt.paramVal {
				t.Errorf("id = %v, want %q", body["id"], tt.paramVal)
			}
			if body["method"] != "GET" {
				t.Errorf("method = %v, want %q", body["method"], "GET")
			}
		})
	}
}

// TestIntegration_ErrorHandler loads the http_error_plugin fixture and verifies
// that a Lua error() inside a handler produces a 500 HANDLER_ERROR with a
// generic message (not the raw Lua error string).
func TestIntegration_ErrorHandler(t *testing.T) {
	bridge, mgr, cleanup := setupFixtureBridge(t)
	defer cleanup()

	loadFixturePlugin(t, bridge, mgr, "http_error_plugin", "1.0.0")
	approveAllRoutes(t, bridge)

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_error_plugin/fail", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "HANDLER_ERROR" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "HANDLER_ERROR")
	}
	if errResp.Error.Message != "internal plugin error" {
		t.Errorf("error message = %q, want generic %q", errResp.Error.Message, "internal plugin error")
	}
	// Verify the Lua error string is NOT leaked to the client.
	if errResp.Error.Message == "deliberate test error" {
		t.Error("Lua error string leaked to client -- must use generic message")
	}
}

// TestIntegration_TimeoutHandler loads the http_timeout_plugin fixture with a
// short execution timeout and verifies that an infinite loop in a handler
// produces a 504 HANDLER_TIMEOUT response.
func TestIntegration_TimeoutHandler(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	// Use a short timeout (1 second) so the test completes quickly.
	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       fixturePluginsDir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  1,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("creating plugin_routes table: %s", err)
	}

	loadFixturePlugin(t, bridge, mgr, "http_timeout_plugin", "1.0.0")
	approveAllRoutes(t, bridge)

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_timeout_plugin/hang", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusGatewayTimeout, rec.Body.String())
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "HANDLER_TIMEOUT" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "HANDLER_TIMEOUT")
	}
}

// TestIntegration_MiddlewareEnrichment loads the http_middleware_plugin fixture
// and verifies that middleware-set fields on the request table are visible to
// the handler and returned in the response.
func TestIntegration_MiddlewareEnrichment(t *testing.T) {
	bridge, mgr, cleanup := setupFixtureBridge(t)
	defer cleanup()

	loadFixturePlugin(t, bridge, mgr, "http_middleware_plugin", "1.0.0")
	approveAllRoutes(t, bridge)

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_middleware_plugin/check", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %s", err)
	}

	if body["custom_field"] != "enriched_value" {
		t.Errorf("custom_field = %v, want %q", body["custom_field"], "enriched_value")
	}

	// request_count is a Lua number, which JSON decodes as float64.
	if body["request_count"] != float64(1) {
		t.Errorf("request_count = %v (type %T), want %v", body["request_count"], body["request_count"], float64(1))
	}
}

// TestIntegration_BlockedHeaders loads the http_blocked_headers fixture and
// verifies that blocked response headers (set-cookie, access-control-allow-origin)
// are filtered out while allowed headers (x-custom, x-plugin-id) pass through.
func TestIntegration_BlockedHeaders(t *testing.T) {
	bridge, mgr, cleanup := setupFixtureBridge(t)
	defer cleanup()

	loadFixturePlugin(t, bridge, mgr, "http_blocked_headers", "1.0.0")
	approveAllRoutes(t, bridge)

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_blocked_headers/headers", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Blocked headers must be absent.
	if rec.Header().Get("Set-Cookie") != "" {
		t.Errorf("Set-Cookie header should be blocked, got %q", rec.Header().Get("Set-Cookie"))
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Access-Control-Allow-Origin header should be blocked, got %q",
			rec.Header().Get("Access-Control-Allow-Origin"))
	}

	// Allowed headers must be present.
	if rec.Header().Get("X-Custom") != "allowed" {
		t.Errorf("X-Custom = %q, want %q", rec.Header().Get("X-Custom"), "allowed")
	}
	if rec.Header().Get("X-Plugin-Id") != "test" {
		t.Errorf("X-Plugin-Id = %q, want %q", rec.Header().Get("X-Plugin-Id"), "test")
	}

	// Verify the body is still correct.
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %s", err)
	}
	if body["ok"] != true {
		t.Errorf("body ok = %v, want true", body["ok"])
	}
}
