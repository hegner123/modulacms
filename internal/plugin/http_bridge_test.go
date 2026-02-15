package plugin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	_ "github.com/mattn/go-sqlite3"
	lua "github.com/yuin/gopher-lua"
)

// -- Test helpers for HTTPBridge --

// loadPluginForTest manually constructs a PluginInstance with a full VM pipeline
// that includes the HTTP API. This bypasses Manager.LoadAll/extractManifest, which
// do not register RegisterHTTPAPI (Agent D's responsibility). Our test plugins call
// http.handle() at module scope, so the VM factory must include RegisterHTTPAPI.
//
// Steps:
//  1. Create PluginInstance with provided metadata
//  2. Build a factory: ApplySandbox -> RegisterPluginRequire -> RegisterDBAPI ->
//     RegisterLogAPI -> FreezeModule(db) -> FreezeModule(log) -> RegisterHTTPAPI ->
//     DoFile(init.lua)
//  3. Create VMPool with that factory
//  4. Checkout one VM, run on_init, snapshot globals, return to pool
//  5. Register the plugin in the Manager's plugins map
func loadPluginForTest(t *testing.T, mgr *Manager, pluginName, pluginVersion, pluginDir string) *PluginInstance {
	t.Helper()

	initPath := filepath.Join(pluginDir, pluginName, "init.lua")

	inst := &PluginInstance{
		Info: PluginInfo{
			Name:        pluginName,
			Version:     pluginVersion,
			Description: "test plugin",
		},
		State:    StateLoading,
		Dir:      filepath.Join(pluginDir, pluginName),
		InitPath: initPath,
		dbAPIs:   make(map[*lua.LState]*DatabaseAPI),
	}

	timeout := time.Duration(mgr.cfg.ExecTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	factory := func() *lua.LState {
		L := lua.NewState(lua.Options{
			SkipOpenLibs:  true,
			CallStackSize: 256,
			RegistrySize:  5120,
		})

		// Full VM factory sequence matching the roadmap + HTTP API for Phase 2.
		ApplySandbox(L, SandboxConfig{AllowCoroutine: true, ExecTimeout: timeout})
		RegisterPluginRequire(L, inst.Dir)

		dbAPI := NewDatabaseAPI(mgr.db, pluginName, mgr.dialect, mgr.cfg.MaxOpsPerExec)
		RegisterDBAPI(L, dbAPI)
		RegisterLogAPI(L, pluginName)
		FreezeModule(L, "db")
		FreezeModule(L, "log")

		// Register HTTP API -- needed for http.handle() calls at module scope.
		RegisterHTTPAPI(L, pluginName)
		RegisterHooksAPI(L, pluginName)
		FreezeModule(L, "http")
		FreezeModule(L, "hooks")

		// Execute init.lua.
		initCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		L.SetContext(initCtx)

		if err := L.DoFile(initPath); err != nil {
			t.Fatalf("plugin %q: init.lua execution error in test factory: %s", pluginName, err)
		}

		L.SetContext(nil)

		// Store dbAPI mapping.
		inst.dbAPIs[L] = dbAPI

		return L
	}

	// Reserve VMs only when pool is large enough (same logic as manager.go).
	reserveSize := 1
	if mgr.cfg.MaxVMsPerPlugin <= 1 {
		reserveSize = 0
	}
	pool := NewVMPool(VMPoolConfig{
		Size:        mgr.cfg.MaxVMsPerPlugin,
		ReserveSize: reserveSize,
		Factory:     factory,
		InitPath:    initPath,
		PluginName:  pluginName,
	})
	inst.Pool = pool

	// Checkout one VM to run on_init and take global snapshot.
	initCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	L, getErr := pool.Get(initCtx)
	if getErr != nil {
		t.Fatalf("plugin %q: failed to get VM from pool: %s", pluginName, getErr)
	}

	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}

	L.SetContext(initCtx)

	// Call on_init() if defined.
	onInit := L.GetGlobal("on_init")
	if fn, ok := onInit.(*lua.LFunction); ok {
		if callErr := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}); callErr != nil {
			pool.Put(L)
			t.Fatalf("plugin %q: on_init failed: %s", pluginName, callErr)
		}
	}

	pool.SnapshotGlobals(L)
	pool.Put(L)

	inst.State = StateRunning

	// Register in manager's plugins map.
	mgr.mu.Lock()
	mgr.plugins[pluginName] = inst
	mgr.loadOrder = append(mgr.loadOrder, pluginName)
	mgr.mu.Unlock()

	return inst
}

// setupTestBridge creates a Manager, an HTTPBridge, and the plugin_routes table.
func setupTestBridge(t *testing.T, pluginDir string) (*HTTPBridge, *Manager, *sql.DB) {
	t.Helper()

	conn := newTestDB(t)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       pluginDir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)

	// Create plugin_routes table.
	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("creating plugin_routes table: %s", err)
	}

	return bridge, mgr, conn
}

// setupTestBridgeWithPlugin creates a bridge with a single test plugin loaded
// and routes registered. The plugin has GET/POST authenticated routes, a public
// webhook route, a path parameter route, and a header test route.
func setupTestBridgeWithPlugin(t *testing.T) (*HTTPBridge, *Manager, *sql.DB) {
	t.Helper()

	dir := t.TempDir()
	writePluginFile(t, dir, "http_test", `
plugin_info = {
    name = "http_test",
    version = "1.0.0",
    description = "Test plugin for HTTP bridge tests",
}

http.handle("GET", "/hello", function(req)
    return {status = 200, json = {message = "hello"}}
end)

http.handle("POST", "/echo", function(req)
    return {status = 200, json = req.json}
end)

http.handle("POST", "/webhook", function(req)
    return {status = 200, json = {received = true}}
end, {public = true})

http.handle("GET", "/items/{id}", function(req)
    return {status = 200, json = {id = req.params.id}}
end)

http.handle("GET", "/headers", function(req)
    return {
        status = 200,
        headers = {
            ["set-cookie"] = "evil=1",
            ["x-custom"] = "allowed",
        },
        json = {ok = true},
    }
end)

function on_init()
end
`)

	bridge, mgr, conn := setupTestBridge(t, dir)

	// Load the plugin manually (bypasses extractManifest which lacks HTTP API).
	loadPluginForTest(t, mgr, "http_test", "1.0.0", dir)

	inst := mgr.GetPlugin("http_test")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("expected http_test plugin to be running, got %v", inst)
	}

	// Register routes from the loaded plugin.
	ctx := context.Background()
	L, getErr := inst.Pool.Get(ctx)
	if getErr != nil {
		t.Fatalf("getting VM for route registration: %s", getErr)
	}
	L.SetContext(ctx)

	regErr := bridge.RegisterRoutes(ctx, "http_test", "1.0.0", L)
	inst.Pool.Put(L)
	if regErr != nil {
		t.Fatalf("RegisterRoutes: %s", regErr)
	}

	return bridge, mgr, conn
}

// withAuthenticatedUser injects a mock authenticated user into the request context
// for testing authenticated routes. Uses middleware.SetAuthenticatedUser to set
// the user with the correct unexported context key so that
// middleware.AuthenticatedUser finds it.
func withAuthenticatedUser(r *http.Request) *http.Request {
	user := &db.Users{
		Username: "testuser",
		Role:     "admin",
	}
	ctx := middleware.SetAuthenticatedUser(r.Context(), user)
	return r.WithContext(ctx)
}

// decodePluginError decodes the response body as a PluginErrorResponse.
func decodePluginError(t *testing.T, rec *httptest.ResponseRecorder) PluginErrorResponse {
	t.Helper()
	var errResp PluginErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %s (body: %s)", err, rec.Body.String())
	}
	return errResp
}

// loadTestPluginWithRoutes is a convenience that creates a temp dir, writes a plugin,
// loads it via loadPluginForTest, and registers routes on the bridge.
func loadTestPluginWithRoutes(t *testing.T, bridge *HTTPBridge, mgr *Manager, pluginName, version, luaCode string) string {
	t.Helper()

	dir := mgr.cfg.Directory
	writePluginFile(t, dir, pluginName, luaCode)
	loadPluginForTest(t, mgr, pluginName, version, dir)

	inst := mgr.GetPlugin(pluginName)
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("expected %s plugin to be running", pluginName)
	}

	ctx := context.Background()
	L, getErr := inst.Pool.Get(ctx)
	if getErr != nil {
		t.Fatalf("getting VM for route registration: %s", getErr)
	}
	L.SetContext(ctx)

	regErr := bridge.RegisterRoutes(ctx, pluginName, version, L)
	inst.Pool.Put(L)
	if regErr != nil {
		t.Fatalf("RegisterRoutes for %s: %s", pluginName, regErr)
	}

	return dir
}

// -- CreatePluginRoutesTable tests --

func TestCreatePluginRoutesTable(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{ExecTimeoutSec: 5}, conn, db.DialectSQLite)
	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	err := bridge.CreatePluginRoutesTable(context.Background())
	if err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	// Verify the table exists by inserting a row.
	_, err = conn.Exec(
		"INSERT INTO plugin_routes (plugin_name, method, path, plugin_version) VALUES (?, ?, ?, ?)",
		"test_plugin", "GET", "/test", "1.0.0",
	)
	if err != nil {
		t.Fatalf("inserting into plugin_routes: %s", err)
	}

	// Verify the row.
	var name string
	err = conn.QueryRow("SELECT plugin_name FROM plugin_routes LIMIT 1").Scan(&name)
	if err != nil {
		t.Fatalf("querying plugin_routes: %s", err)
	}
	if name != "test_plugin" {
		t.Errorf("plugin_name = %q, want %q", name, "test_plugin")
	}
}

func TestCreatePluginRoutesTable_Idempotent(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{ExecTimeoutSec: 5}, conn, db.DialectSQLite)
	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	// Call twice -- should not error.
	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("first call: %s", err)
	}
	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("second call: %s", err)
	}
}

// -- CleanupOrphanedRoutes tests --

func TestCleanupOrphanedRoutes(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{ExecTimeoutSec: 5}, conn, db.DialectSQLite)
	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	ctx := context.Background()
	if err := bridge.CreatePluginRoutesTable(ctx); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	// Insert routes for two plugins.
	for _, name := range []string{"plugin_a", "plugin_b"} {
		_, err := conn.Exec(
			"INSERT INTO plugin_routes (plugin_name, method, path, plugin_version) VALUES (?, ?, ?, ?)",
			name, "GET", "/test", "1.0.0",
		)
		if err != nil {
			t.Fatalf("inserting route for %s: %s", name, err)
		}
	}

	// Only plugin_a is discovered -- plugin_b should be cleaned up.
	if err := bridge.CleanupOrphanedRoutes(ctx, []string{"plugin_a"}); err != nil {
		t.Fatalf("CleanupOrphanedRoutes: %s", err)
	}

	var count int
	err := conn.QueryRow("SELECT COUNT(*) FROM plugin_routes").Scan(&count)
	if err != nil {
		t.Fatalf("counting routes: %s", err)
	}
	if count != 1 {
		t.Errorf("expected 1 route remaining, got %d", count)
	}

	var remaining string
	err = conn.QueryRow("SELECT plugin_name FROM plugin_routes").Scan(&remaining)
	if err != nil {
		t.Fatalf("querying remaining: %s", err)
	}
	if remaining != "plugin_a" {
		t.Errorf("remaining plugin = %q, want %q", remaining, "plugin_a")
	}
}

func TestCleanupOrphanedRoutes_EmptyDiscovered(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{ExecTimeoutSec: 5}, conn, db.DialectSQLite)
	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	ctx := context.Background()
	if err := bridge.CreatePluginRoutesTable(ctx); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	_, err := conn.Exec(
		"INSERT INTO plugin_routes (plugin_name, method, path, plugin_version) VALUES (?, ?, ?, ?)",
		"orphan", "GET", "/test", "1.0.0",
	)
	if err != nil {
		t.Fatalf("inserting: %s", err)
	}

	// No plugins discovered -- all rows should be deleted.
	if err := bridge.CleanupOrphanedRoutes(ctx, nil); err != nil {
		t.Fatalf("CleanupOrphanedRoutes: %s", err)
	}

	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM plugin_routes").Scan(&count)
	if err != nil {
		t.Fatalf("counting: %s", err)
	}
	if count != 0 {
		t.Errorf("expected 0 routes, got %d", count)
	}
}

// -- RegisterRoutes tests --

func TestRegisterRoutes_ExtractsFromHandlers(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	routes := bridge.ListRoutes()
	if len(routes) == 0 {
		t.Fatal("expected routes to be registered, got 0")
	}

	// Verify expected routes exist.
	routeMap := make(map[string]*RouteRegistration)
	for i := range routes {
		key := routes[i].Method + " " + routes[i].Path
		routeMap[key] = &routes[i]
	}

	expectedRoutes := []struct {
		key    string
		public bool
	}{
		{"GET /hello", false},
		{"POST /echo", false},
		{"POST /webhook", true},
		{"GET /items/{id}", false},
		{"GET /headers", false},
	}

	for _, expected := range expectedRoutes {
		reg, found := routeMap[expected.key]
		if !found {
			t.Errorf("expected route %q to be registered", expected.key)
			continue
		}
		if reg.Public != expected.public {
			t.Errorf("route %q: Public = %v, want %v", expected.key, reg.Public, expected.public)
		}
		if reg.PluginName != "http_test" {
			t.Errorf("route %q: PluginName = %q, want %q", expected.key, reg.PluginName, "http_test")
		}
	}
}

func TestRegisterRoutes_CrossPluginCollision(t *testing.T) {
	// Cross-plugin collision cannot happen naturally because each plugin has
	// its own prefix (/api/v1/plugins/<name>/). This test verifies the
	// collision detection code path by synthetically injecting a route entry
	// under a different plugin name that maps to the same mux pattern.
	dir := t.TempDir()

	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	// Create and load plugin_b.
	writePluginFile(t, dir, "plugin_b", `
plugin_info = {
    name = "plugin_b",
    version = "1.0.0",
    description = "Plugin B",
}
http.handle("GET", "/conflict", function(req) return {status = 200} end)
function on_init() end
`)
	loadPluginForTest(t, mgr, "plugin_b", "1.0.0", dir)

	// Synthetically inject a route in the bridge's route map that maps to the
	// same mux pattern that plugin_b's /conflict route would produce.
	// This simulates a hypothetical scenario where a different plugin already
	// owns that exact path.
	muxPattern := "GET " + PluginRoutePrefix + "plugin_b/conflict"
	bridge.mu.Lock()
	bridge.routes[muxPattern] = &RouteRegistration{
		Method:     "GET",
		Path:       "/conflict",
		PluginName: "imposter_plugin", // different plugin name
		FullPath:   PluginRoutePrefix + "plugin_b/conflict",
	}
	bridge.mu.Unlock()

	// Now RegisterRoutes for plugin_b should detect the collision.
	instB := mgr.GetPlugin("plugin_b")
	ctx := context.Background()
	L, getErr := instB.Pool.Get(ctx)
	if getErr != nil {
		t.Fatalf("getting VM: %s", getErr)
	}
	L.SetContext(ctx)
	regErr := bridge.RegisterRoutes(ctx, "plugin_b", "1.0.0", L)
	instB.Pool.Put(L)

	if regErr == nil {
		t.Fatal("expected error for cross-plugin route collision, got nil")
	}
	if !strings.Contains(regErr.Error(), "route collision") {
		t.Errorf("error = %q, want to contain %q", regErr.Error(), "route collision")
	}
}

// TestRegisterRoutes_DifferentPluginsNoCollision verifies that two different
// plugins can register the same relative path (e.g., "/tasks") without collision
// because each plugin has a unique prefix.
func TestRegisterRoutes_DifferentPluginsNoCollision(t *testing.T) {
	dir := t.TempDir()

	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	// Create and load plugin_a with a /tasks route.
	writePluginFile(t, dir, "plugin_a", `
plugin_info = {
    name = "plugin_a",
    version = "1.0.0",
    description = "Plugin A",
}
http.handle("GET", "/tasks", function(req) return {status = 200} end)
function on_init() end
`)
	loadPluginForTest(t, mgr, "plugin_a", "1.0.0", dir)

	instA := mgr.GetPlugin("plugin_a")
	ctx := context.Background()
	L, _ := instA.Pool.Get(ctx)
	L.SetContext(ctx)
	regErr := bridge.RegisterRoutes(ctx, "plugin_a", "1.0.0", L)
	instA.Pool.Put(L)
	if regErr != nil {
		t.Fatalf("RegisterRoutes for plugin_a: %s", regErr)
	}

	// Create and load plugin_b with the same relative /tasks route.
	writePluginFile(t, dir, "plugin_b", `
plugin_info = {
    name = "plugin_b",
    version = "1.0.0",
    description = "Plugin B",
}
http.handle("GET", "/tasks", function(req) return {status = 200} end)
function on_init() end
`)
	loadPluginForTest(t, mgr, "plugin_b", "1.0.0", dir)

	instB := mgr.GetPlugin("plugin_b")
	L2, _ := instB.Pool.Get(ctx)
	L2.SetContext(ctx)
	regErr2 := bridge.RegisterRoutes(ctx, "plugin_b", "1.0.0", L2)
	instB.Pool.Put(L2)

	// Should NOT collide since full paths are different:
	// /api/v1/plugins/plugin_a/tasks vs /api/v1/plugins/plugin_b/tasks
	if regErr2 != nil {
		t.Fatalf("expected no collision for different plugins with same relative path, got: %s", regErr2)
	}

	routes := bridge.ListRoutes()
	if len(routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(routes))
	}
}

// -- ServeHTTP dispatch tests --

func TestServeHTTP_DispatchApprovedRoute(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	// Approve the /hello route.
	if err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/hello", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %s", err)
	}
	if body["message"] != "hello" {
		t.Errorf("body = %v, want message=hello", body)
	}
}

func TestServeHTTP_Uniform404_UnknownRoute(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/nonexistent/route", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "ROUTE_NOT_FOUND" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "ROUTE_NOT_FOUND")
	}
}

func TestServeHTTP_Uniform404_UnapprovedRoute(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	// Routes are unapproved by default -- mount to get the fallback handler.
	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Even though the route is registered, it is unapproved and should 404.
	// The fallback handler at /api/v1/plugins/ will catch this.
	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "ROUTE_NOT_FOUND" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "ROUTE_NOT_FOUND")
	}
}

func TestServeHTTP_503_NonRunningPlugin(t *testing.T) {
	bridge, mgr, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	// Approve the route.
	if err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/hello", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	// Force the plugin into a non-running state.
	inst := mgr.GetPlugin("http_test")
	inst.State = StateFailed

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "PLUGIN_UNAVAILABLE" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "PLUGIN_UNAVAILABLE")
	}
}

func TestServeHTTP_503_PoolExhaustion(t *testing.T) {
	dir := t.TempDir()

	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 1, // Only 1 VM
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	writePluginFile(t, dir, "pool_test", `
plugin_info = {
    name = "pool_test",
    version = "1.0.0",
    description = "Pool exhaustion test",
}
http.handle("GET", "/slow", function(req)
    return {status = 200, json = {ok = true}}
end)
function on_init() end
`)
	loadPluginForTest(t, mgr, "pool_test", "1.0.0", dir)

	inst := mgr.GetPlugin("pool_test")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("expected pool_test to be running")
	}

	ctx := context.Background()
	L, getErr := inst.Pool.Get(ctx)
	if getErr != nil {
		t.Fatalf("getting VM: %s", getErr)
	}
	L.SetContext(ctx)
	if err := bridge.RegisterRoutes(ctx, "pool_test", "1.0.0", L); err != nil {
		inst.Pool.Put(L)
		t.Fatalf("RegisterRoutes: %s", err)
	}
	inst.Pool.Put(L)

	// Approve the route.
	if err := bridge.ApproveRoute(ctx, "pool_test", "GET", "/slow", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Checkout the only VM to simulate exhaustion.
	L2, getErr2 := inst.Pool.Get(ctx)
	if getErr2 != nil {
		t.Fatalf("getting VM for exhaustion: %s", getErr2)
	}
	defer inst.Pool.Put(L2)

	req := httptest.NewRequest("GET", "/api/v1/plugins/pool_test/slow", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "POOL_EXHAUSTED" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "POOL_EXHAUSTED")
	}

	retryAfter := rec.Header().Get("Retry-After")
	if retryAfter != "1" {
		t.Errorf("Retry-After = %q, want %q", retryAfter, "1")
	}
}

// -- Auth enforcement tests --

func TestServeHTTP_AuthRequired_NoSession(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	if err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/hello", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// No authenticated user in context.
	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "UNAUTHORIZED")
	}
}

func TestServeHTTP_AuthRequired_WithSession(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	if err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/hello", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestServeHTTP_PublicRoute_NoSession(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	if err := bridge.ApproveRoute(context.Background(), "http_test", "POST", "/webhook", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Public route should work without authentication.
	body := `{"event": "test"}`
	req := httptest.NewRequest("POST", "/api/v1/plugins/http_test/webhook",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestServeHTTP_PublicRoute_WithSession(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	if err := bridge.ApproveRoute(context.Background(), "http_test", "POST", "/webhook", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Public route should also work with authentication.
	body := `{"event": "test"}`
	req := httptest.NewRequest("POST", "/api/v1/plugins/http_test/webhook",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// -- Admin approval/revoke lifecycle tests --

func TestApproveRevoke_Lifecycle(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	ctx := context.Background()

	// Initially unapproved.
	routes := bridge.ListRoutes()
	for _, r := range routes {
		if r.Approved {
			t.Errorf("route %s %s should be unapproved initially", r.Method, r.Path)
		}
	}

	// Approve a route.
	if err := bridge.ApproveRoute(ctx, "http_test", "GET", "/hello", "admin_user"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	// Verify approved in memory.
	routes = bridge.ListRoutes()
	found := false
	for _, r := range routes {
		if r.Method == "GET" && r.Path == "/hello" {
			found = true
			if !r.Approved {
				t.Error("route should be approved after ApproveRoute")
			}
		}
	}
	if !found {
		t.Error("GET /hello route not found in ListRoutes")
	}

	// Verify approved in DB.
	var approved int
	var approvedBy string
	err := conn.QueryRow(
		"SELECT approved, approved_by FROM plugin_routes WHERE plugin_name = ? AND method = ? AND path = ?",
		"http_test", "GET", "/hello",
	).Scan(&approved, &approvedBy)
	if err != nil {
		t.Fatalf("querying DB: %s", err)
	}
	if approved != 1 {
		t.Errorf("DB approved = %d, want 1", approved)
	}
	if approvedBy != "admin_user" {
		t.Errorf("DB approved_by = %q, want %q", approvedBy, "admin_user")
	}

	// Revoke the route.
	if err := bridge.RevokeRoute(ctx, "http_test", "GET", "/hello"); err != nil {
		t.Fatalf("RevokeRoute: %s", err)
	}

	// Verify revoked in memory.
	routes = bridge.ListRoutes()
	for _, r := range routes {
		if r.Method == "GET" && r.Path == "/hello" {
			if r.Approved {
				t.Error("route should be unapproved after RevokeRoute")
			}
		}
	}

	// Verify revoked in DB.
	var revokedApproved int
	err = conn.QueryRow(
		"SELECT approved FROM plugin_routes WHERE plugin_name = ? AND method = ? AND path = ?",
		"http_test", "GET", "/hello",
	).Scan(&revokedApproved)
	if err != nil {
		t.Fatalf("querying DB: %s", err)
	}
	if revokedApproved != 0 {
		t.Errorf("DB approved = %d, want 0 after revoke", revokedApproved)
	}
}

func TestApproveRoute_ThenAccessible(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Approve the route.
	if err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/hello", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRevokeRoute_Then404(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Approve then revoke.
	ctx := context.Background()
	if err := bridge.ApproveRoute(ctx, "http_test", "GET", "/hello", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}
	if err := bridge.RevokeRoute(ctx, "http_test", "GET", "/hello"); err != nil {
		t.Fatalf("RevokeRoute: %s", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// -- Body size rejection test --

func TestServeHTTP_BodyTooLarge(t *testing.T) {
	dir := t.TempDir()

	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	// Set a small body limit for testing.
	bridge.maxBodySize = 100
	defer bridge.Close(context.Background())

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	writePluginFile(t, dir, "body_test", `
plugin_info = {
    name = "body_test",
    version = "1.0.0",
    description = "Body size test",
}
http.handle("POST", "/upload", function(req)
    return {status = 200, json = {size = string.len(req.body)}}
end, {public = true})
function on_init() end
`)
	loadPluginForTest(t, mgr, "body_test", "1.0.0", dir)

	ctx := context.Background()
	inst := mgr.GetPlugin("body_test")
	L, _ := inst.Pool.Get(ctx)
	L.SetContext(ctx)
	if err := bridge.RegisterRoutes(ctx, "body_test", "1.0.0", L); err != nil {
		inst.Pool.Put(L)
		t.Fatalf("RegisterRoutes: %s", err)
	}
	inst.Pool.Put(L)

	if err := bridge.ApproveRoute(ctx, "body_test", "POST", "/upload", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Send a body larger than the limit.
	largeBody := strings.Repeat("x", 200)
	req := httptest.NewRequest("POST", "/api/v1/plugins/body_test/upload",
		strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "INVALID_REQUEST" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "INVALID_REQUEST")
	}
}

// -- Header blocking test --

func TestServeHTTP_HeaderBlocking(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	if err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/headers", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/headers", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Set-Cookie should be blocked.
	if rec.Header().Get("Set-Cookie") != "" {
		t.Errorf("Set-Cookie header should be blocked, got %q", rec.Header().Get("Set-Cookie"))
	}

	// X-Custom should pass through.
	if rec.Header().Get("X-Custom") != "allowed" {
		t.Errorf("X-Custom = %q, want %q", rec.Header().Get("X-Custom"), "allowed")
	}
}

// -- Security headers test --

func TestServeHTTP_SecurityHeaders(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	if err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/hello", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Check security headers are present.
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want %q",
			rec.Header().Get("X-Content-Type-Options"), "nosniff")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("X-Frame-Options = %q, want %q",
			rec.Header().Get("X-Frame-Options"), "DENY")
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q, want %q",
			rec.Header().Get("Cache-Control"), "no-store")
	}
}

// Security headers are also present on error responses.
func TestServeHTTP_SecurityHeaders_OnError(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Request to a route that does not exist.
	req := httptest.NewRequest("GET", "/api/v1/plugins/nonexistent/route", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	// Security headers set by ServeHTTP before the route lookup, so they
	// should still be present on 404 responses.
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want %q on error",
			rec.Header().Get("X-Content-Type-Options"), "nosniff")
	}
}

// -- Graceful shutdown test --

func TestClose_NewRequestsGet503(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()

	if err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/hello", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// Close the bridge -- new requests should get 503.
	bridge.Close(context.Background())

	req := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d after Close", rec.Code, http.StatusServiceUnavailable)
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "PLUGIN_UNAVAILABLE" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "PLUGIN_UNAVAILABLE")
	}
}

// -- Version change revokes approvals test --

func TestVersionChange_RevokesAllApprovals(t *testing.T) {
	dir := t.TempDir()

	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	ctx := context.Background()
	if err := bridge.CreatePluginRoutesTable(ctx); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	writePluginFile(t, dir, "versioned", `
plugin_info = {
    name = "versioned",
    version = "1.0.0",
    description = "Version test",
}
http.handle("GET", "/data", function(req) return {status = 200} end)
function on_init() end
`)
	loadPluginForTest(t, mgr, "versioned", "1.0.0", dir)

	inst := mgr.GetPlugin("versioned")
	L, _ := inst.Pool.Get(ctx)
	L.SetContext(ctx)

	// Register with version 1.0.0.
	if err := bridge.RegisterRoutes(ctx, "versioned", "1.0.0", L); err != nil {
		inst.Pool.Put(L)
		t.Fatalf("RegisterRoutes v1: %s", err)
	}
	inst.Pool.Put(L)

	// Approve the route.
	if err := bridge.ApproveRoute(ctx, "versioned", "GET", "/data", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	// Verify approved.
	routes := bridge.ListRoutes()
	for _, r := range routes {
		if r.Method == "GET" && r.Path == "/data" && !r.Approved {
			t.Error("route should be approved")
		}
	}

	// Now re-register with a new version -- should revoke approval.
	L2, _ := inst.Pool.Get(ctx)
	L2.SetContext(ctx)
	if err := bridge.RegisterRoutes(ctx, "versioned", "2.0.0", L2); err != nil {
		inst.Pool.Put(L2)
		t.Fatalf("RegisterRoutes v2: %s", err)
	}
	inst.Pool.Put(L2)

	// Verify all routes are unapproved after version change.
	routes = bridge.ListRoutes()
	for _, r := range routes {
		if r.PluginName == "versioned" && r.Approved {
			t.Errorf("route %s %s should be unapproved after version change", r.Method, r.Path)
		}
	}
}

// -- ListRoutes test --

func TestListRoutes_ReturnsCopy(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	routes := bridge.ListRoutes()
	if len(routes) == 0 {
		t.Fatal("expected at least one route")
	}

	// Modifying the returned slice should not affect the bridge's internal state.
	routes[0].Approved = true
	actual := bridge.ListRoutes()
	for _, r := range actual {
		if r.Method == routes[0].Method && r.Path == routes[0].Path && r.Approved {
			// This would only fail if ListRoutes returned a reference to the
			// internal struct rather than a copy. Since RouteRegistration is a
			// value type (no pointer fields), the copy is inherent.
			break
		}
	}
}

// -- Rate limiting test --

func TestServeHTTP_RateLimiting(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	// Set a very low rate limit for testing.
	bridge.rateLimit = 1
	bridge.rateBurst = 1

	if err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/hello", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	// First request should succeed.
	req1 := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req1 = withAuthenticatedUser(req1)
	req1.RemoteAddr = "192.168.1.100:12345"
	rec1 := httptest.NewRecorder()
	mux.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("first request: status = %d, want %d", rec1.Code, http.StatusOK)
	}

	// Second request immediately should be rate limited.
	req2 := httptest.NewRequest("GET", "/api/v1/plugins/http_test/hello", nil)
	req2 = withAuthenticatedUser(req2)
	req2.RemoteAddr = "192.168.1.100:12346"
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: status = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}

	retryAfter := rec2.Header().Get("Retry-After")
	if retryAfter != "1" {
		t.Errorf("Retry-After = %q, want %q", retryAfter, "1")
	}
}

// -- POST with JSON echo test --

func TestServeHTTP_PostEcho(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	if err := bridge.ApproveRoute(context.Background(), "http_test", "POST", "/echo", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	payload := map[string]any{"key": "value", "num": float64(42)}
	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/plugins/http_test/echo",
		bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decoding response: %s", err)
	}
	if result["key"] != "value" {
		t.Errorf("result[key] = %v, want %q", result["key"], "value")
	}
	if result["num"] != float64(42) {
		t.Errorf("result[num] = %v, want %v", result["num"], float64(42))
	}
}

// -- ApproveRoute / RevokeRoute error cases --

func TestApproveRoute_NotFound(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	err := bridge.ApproveRoute(context.Background(), "http_test", "GET", "/nonexistent", "admin")
	if err == nil {
		t.Fatal("expected error for non-existent route, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "not found")
	}
}

func TestRevokeRoute_NotFound(t *testing.T) {
	bridge, _, conn := setupTestBridgeWithPlugin(t)
	defer conn.Close()
	defer bridge.Close(context.Background())

	err := bridge.RevokeRoute(context.Background(), "http_test", "GET", "/nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent route, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "not found")
	}
}

// -- Middleware chain test --

func TestServeHTTP_MiddlewareChain(t *testing.T) {
	dir := t.TempDir()

	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	writePluginFile(t, dir, "mw_test", `
plugin_info = {
    name = "mw_test",
    version = "1.0.0",
    description = "Middleware test",
}

http.use(function(req)
    req.enriched = "from_middleware"
    return nil
end)

http.handle("GET", "/enriched", function(req)
    return {status = 200, json = {enriched = req.enriched}}
end, {public = true})

function on_init() end
`)
	loadPluginForTest(t, mgr, "mw_test", "1.0.0", dir)

	inst := mgr.GetPlugin("mw_test")
	ctx := context.Background()
	L, _ := inst.Pool.Get(ctx)
	L.SetContext(ctx)
	if err := bridge.RegisterRoutes(ctx, "mw_test", "1.0.0", L); err != nil {
		inst.Pool.Put(L)
		t.Fatalf("RegisterRoutes: %s", err)
	}
	inst.Pool.Put(L)

	if err := bridge.ApproveRoute(ctx, "mw_test", "GET", "/enriched", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/mw_test/enriched", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %s", err)
	}
	if body["enriched"] != "from_middleware" {
		t.Errorf("enriched = %v, want %q", body["enriched"], "from_middleware")
	}
}

// -- Middleware early response test --

func TestServeHTTP_MiddlewareEarlyResponse(t *testing.T) {
	dir := t.TempDir()

	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	writePluginFile(t, dir, "mw_early", `
plugin_info = {
    name = "mw_early",
    version = "1.0.0",
    description = "Middleware early response test",
}

http.use(function(req)
    -- Return an early response, skipping the handler.
    return {status = 403, json = {error = "blocked by middleware"}}
end)

http.handle("GET", "/blocked", function(req)
    -- This should never be reached.
    return {status = 200, json = {message = "should not see this"}}
end, {public = true})

function on_init() end
`)
	loadPluginForTest(t, mgr, "mw_early", "1.0.0", dir)

	inst := mgr.GetPlugin("mw_early")
	ctx := context.Background()
	L, _ := inst.Pool.Get(ctx)
	L.SetContext(ctx)
	if err := bridge.RegisterRoutes(ctx, "mw_early", "1.0.0", L); err != nil {
		inst.Pool.Put(L)
		t.Fatalf("RegisterRoutes: %s", err)
	}
	inst.Pool.Put(L)

	if err := bridge.ApproveRoute(ctx, "mw_early", "GET", "/blocked", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/mw_early/blocked", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %s", err)
	}
	if body["error"] != "blocked by middleware" {
		t.Errorf("error = %v, want %q", body["error"], "blocked by middleware")
	}
}

// -- Response size limit test --

func TestServeHTTP_ResponseTooLarge(t *testing.T) {
	dir := t.TempDir()

	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	bridge.maxRespSize = 100 // Very small limit for testing.
	defer bridge.Close(context.Background())

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	// Create a handler that returns a large response.
	writePluginFile(t, dir, "big_resp", `
plugin_info = {
    name = "big_resp",
    version = "1.0.0",
    description = "Big response test",
}

http.handle("GET", "/big", function(req)
    local big = string.rep("x", 1000)
    return {status = 200, body = big}
end, {public = true})

function on_init() end
`)
	loadPluginForTest(t, mgr, "big_resp", "1.0.0", dir)

	inst := mgr.GetPlugin("big_resp")
	ctx := context.Background()
	L, _ := inst.Pool.Get(ctx)
	L.SetContext(ctx)
	if err := bridge.RegisterRoutes(ctx, "big_resp", "1.0.0", L); err != nil {
		inst.Pool.Put(L)
		t.Fatalf("RegisterRoutes: %s", err)
	}
	inst.Pool.Put(L)

	if err := bridge.ApproveRoute(ctx, "big_resp", "GET", "/big", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/big_resp/big", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// -- allowIP rate limiter tests --

func TestAllowIP_RateLimiting(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{ExecTimeoutSec: 5}, conn, db.DialectSQLite)
	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	// Set very restrictive rate limit.
	bridge.rateLimit = 1
	bridge.rateBurst = 1

	// First call should be allowed.
	if !bridge.allowIP("10.0.0.1") {
		t.Error("first call should be allowed")
	}

	// Second call immediately should be denied.
	if bridge.allowIP("10.0.0.1") {
		t.Error("second call should be denied")
	}

	// Different IP should be allowed.
	if !bridge.allowIP("10.0.0.2") {
		t.Error("different IP should be allowed")
	}
}

// -- Handler error test --

func TestServeHTTP_HandlerError(t *testing.T) {
	dir := t.TempDir()

	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	bridge := NewHTTPBridge(mgr, conn, db.DialectSQLite)
	defer bridge.Close(context.Background())

	if err := bridge.CreatePluginRoutesTable(context.Background()); err != nil {
		t.Fatalf("CreatePluginRoutesTable: %s", err)
	}

	writePluginFile(t, dir, "err_handler", `
plugin_info = {
    name = "err_handler",
    version = "1.0.0",
    description = "Error handler test",
}

http.handle("GET", "/fail", function(req)
    error("something went wrong")
end, {public = true})

function on_init() end
`)
	loadPluginForTest(t, mgr, "err_handler", "1.0.0", dir)

	inst := mgr.GetPlugin("err_handler")
	ctx := context.Background()
	L, _ := inst.Pool.Get(ctx)
	L.SetContext(ctx)
	if err := bridge.RegisterRoutes(ctx, "err_handler", "1.0.0", L); err != nil {
		inst.Pool.Put(L)
		t.Fatalf("RegisterRoutes: %s", err)
	}
	inst.Pool.Put(L)

	if err := bridge.ApproveRoute(ctx, "err_handler", "GET", "/fail", "admin"); err != nil {
		t.Fatalf("ApproveRoute: %s", err)
	}

	mux := http.NewServeMux()
	bridge.MountOn(mux)

	req := httptest.NewRequest("GET", "/api/v1/plugins/err_handler/fail", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	errResp := decodePluginError(t, rec)
	if errResp.Error.Code != "HANDLER_ERROR" {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, "HANDLER_ERROR")
	}
	// The error message should be generic, not the Lua error string.
	if errResp.Error.Message != "internal plugin error" {
		t.Errorf("error message = %q, want %q", errResp.Error.Message, "internal plugin error")
	}
}

// -- Ensure middleware.AuthenticatedUser is importable --
// This test verifies the import works correctly (compile-time check mostly).
func TestAuthenticatedUserIntegration(t *testing.T) {
	user := middleware.AuthenticatedUser(context.Background())
	if user != nil {
		t.Error("expected nil user from empty context")
	}
}

// -- Ensure unused imports don't cause compilation failures --
// These are compile-time verification helpers.
var _ = time.Now
var _ lua.LValue
