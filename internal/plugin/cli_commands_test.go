package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// -- Test helpers --

// newTestManager creates a Manager with in-memory SQLite for handler tests.
// Returns the manager and a cleanup function.
func newTestManager(t *testing.T) (*Manager, func()) {
	t.Helper()

	pool, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("opening in-memory sqlite: %v", err)
	}

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       t.TempDir(),
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
		MaxFailures:     3,
		ResetInterval:   60 * time.Second,
	}, pool, db.DialectSQLite)

	cleanup := func() {
		mgr.Shutdown(context.Background())
		// Pool is closed by Shutdown; second close is benign no-op.
	}

	return mgr, cleanup
}

// addTestPlugin inserts a plugin instance into the manager's plugin map
// for handler tests. This bypasses the full loading flow.
func addTestPlugin(mgr *Manager, name string, state PluginState, pool *VMPool) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	inst := &PluginInstance{
		Info: PluginInfo{
			Name:        name,
			Version:     "1.0.0",
			Description: "Test plugin " + name,
		},
		State: state,
		Pool:  pool,
		CB:    NewCircuitBreaker(name, 3, 60*time.Second),
	}
	mgr.plugins[name] = inst
}

// -- PluginListHandler tests --

func TestPluginListHandler_EmptyList(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/admin/plugins", nil)
	w := httptest.NewRecorder()

	PluginListHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	plugins, ok := body["plugins"].([]any)
	if !ok {
		t.Fatal("expected plugins array in response")
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestPluginListHandler_WithPlugins(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	addTestPlugin(mgr, "alpha", StateRunning, nil)
	addTestPlugin(mgr, "beta", StateFailed, nil)

	req := httptest.NewRequest("GET", "/api/v1/admin/plugins", nil)
	w := httptest.NewRecorder()

	PluginListHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	plugins, ok := body["plugins"].([]any)
	if !ok {
		t.Fatal("expected plugins array")
	}
	if len(plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(plugins))
	}
}

// -- PluginInfoHandler tests --

func TestPluginInfoHandler_Found(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	pool := newTestPool(2)
	defer pool.Close()
	addTestPlugin(mgr, "test_info", StateRunning, pool)

	req := httptest.NewRequest("GET", "/api/v1/admin/plugins/test_info", nil)
	req.SetPathValue("name", "test_info")
	w := httptest.NewRecorder()

	PluginInfoHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if body["name"] != "test_info" {
		t.Errorf("name = %v, want %q", body["name"], "test_info")
	}
	if body["state"] != "running" {
		t.Errorf("state = %v, want %q", body["state"], "running")
	}

	// Should include VM counts from the pool.
	vmsTotal, ok := body["vms_total"].(float64)
	if !ok || vmsTotal != 2 {
		t.Errorf("vms_total = %v, want 2", body["vms_total"])
	}
}

func TestPluginInfoHandler_NotFound(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/admin/plugins/nonexistent", nil)
	req.SetPathValue("name", "nonexistent")
	w := httptest.NewRecorder()

	PluginInfoHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPluginInfoHandler_MissingPathValue(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/admin/plugins/", nil)
	// Do not set path value.
	w := httptest.NewRecorder()

	PluginInfoHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// -- PluginReloadHandler tests --

func TestPluginReloadHandler_PluginNotFound(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/v1/admin/plugins/nonexistent/reload", nil)
	req.SetPathValue("name", "nonexistent")
	w := httptest.NewRecorder()

	PluginReloadHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPluginReloadHandler_MissingName(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/v1/admin/plugins//reload", nil)
	w := httptest.NewRecorder()

	PluginReloadHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// -- PluginDisableHandler tests --

func TestPluginDisableHandler_PluginNotFound(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/v1/admin/plugins/nonexistent/disable", nil)
	req.SetPathValue("name", "nonexistent")
	w := httptest.NewRecorder()

	PluginDisableHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 (not found error), got %d", w.Code)
	}
}

func TestPluginDisableHandler_AlreadyStopped(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	addTestPlugin(mgr, "stopped_plugin", StateStopped, nil)

	req := httptest.NewRequest("POST", "/api/v1/admin/plugins/stopped_plugin/disable", nil)
	req.SetPathValue("name", "stopped_plugin")
	w := httptest.NewRecorder()

	PluginDisableHandler(mgr).ServeHTTP(w, req)

	// Should return error because plugin is already stopped.
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 (already stopped), got %d", w.Code)
	}
}

// -- PluginEnableHandler tests --

func TestPluginEnableHandler_PluginNotFound(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/v1/admin/plugins/nonexistent/enable", nil)
	req.SetPathValue("name", "nonexistent")
	w := httptest.NewRecorder()

	PluginEnableHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// -- PluginCleanupListHandler tests --

func TestPluginCleanupListHandler_EmptyResult(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/admin/plugins/cleanup", nil)
	w := httptest.NewRecorder()

	PluginCleanupListHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	count, ok := body["count"].(float64)
	if !ok {
		t.Fatal("expected count field")
	}
	if count != 0 {
		t.Errorf("expected 0 orphaned tables, got %v", count)
	}
}

// -- PluginCleanupDropHandler tests --

func TestPluginCleanupDropHandler_MissingConfirm(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	body := `{"tables": ["plugin_old_stuff"]}`
	req := httptest.NewRequest("POST", "/api/v1/admin/plugins/cleanup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	PluginCleanupDropHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (missing confirm), got %d", w.Code)
	}
}

func TestPluginCleanupDropHandler_EmptyTables(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	body := `{"confirm": true, "tables": []}`
	req := httptest.NewRequest("POST", "/api/v1/admin/plugins/cleanup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	PluginCleanupDropHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (empty tables), got %d", w.Code)
	}
}

func TestPluginCleanupDropHandler_InvalidJSON(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/v1/admin/plugins/cleanup", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	PluginCleanupDropHandler(mgr).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (invalid JSON), got %d", w.Code)
	}
}

// -- PluginCleanupDrop integration test (with real orphaned tables) --

func TestPluginCleanupDrop_DropsOrphanedTable(t *testing.T) {
	pool, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("opening in-memory sqlite: %v", err)
	}

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       t.TempDir(),
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, pool, db.DialectSQLite)
	defer mgr.Shutdown(context.Background())

	// Create an orphaned table (no plugin claims it).
	_, execErr := pool.Exec("CREATE TABLE plugin_removed_tasks (id TEXT PRIMARY KEY)")
	if execErr != nil {
		t.Fatalf("creating orphaned table: %v", execErr)
	}

	// Verify it shows up in the dry-run.
	req := httptest.NewRequest("GET", "/api/v1/admin/plugins/cleanup", nil)
	w := httptest.NewRecorder()
	PluginCleanupListHandler(mgr).ServeHTTP(w, req)

	var listBody map[string]any
	if err := json.NewDecoder(w.Body).Decode(&listBody); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}

	orphanedRaw, ok := listBody["orphaned_tables"].([]any)
	if !ok {
		t.Fatal("expected orphaned_tables array")
	}
	if len(orphanedRaw) != 1 {
		t.Fatalf("expected 1 orphaned table, got %d", len(orphanedRaw))
	}
	if orphanedRaw[0].(string) != "plugin_removed_tasks" {
		t.Errorf("expected plugin_removed_tasks, got %v", orphanedRaw[0])
	}

	// Drop it.
	dropBody := `{"confirm": true, "tables": ["plugin_removed_tasks"]}`
	dropReq := httptest.NewRequest("POST", "/api/v1/admin/plugins/cleanup", strings.NewReader(dropBody))
	dropReq.Header.Set("Content-Type", "application/json")
	dropW := httptest.NewRecorder()
	PluginCleanupDropHandler(mgr).ServeHTTP(dropW, dropReq)

	if dropW.Code != http.StatusOK {
		t.Fatalf("expected 200 for drop, got %d: %s", dropW.Code, dropW.Body.String())
	}

	var dropResp map[string]any
	if err := json.NewDecoder(dropW.Body).Decode(&dropResp); err != nil {
		t.Fatalf("decoding drop response: %v", err)
	}

	dropped, ok := dropResp["dropped"].([]any)
	if !ok || len(dropped) != 1 {
		t.Errorf("expected 1 dropped table, got %v", dropResp["dropped"])
	}

	// Verify the table is gone.
	var count int
	row := pool.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='plugin_removed_tasks'")
	if scanErr := row.Scan(&count); scanErr != nil {
		t.Fatalf("checking table existence: %v", scanErr)
	}
	if count != 0 {
		t.Error("orphaned table should have been dropped")
	}
}

// -- Response content-type checks --

func TestPluginListHandler_ContentType(t *testing.T) {
	mgr, cleanup := newTestManager(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/admin/plugins", nil)
	w := httptest.NewRecorder()

	PluginListHandler(mgr).ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}
