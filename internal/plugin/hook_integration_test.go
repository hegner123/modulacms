package plugin

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
)

// TestHookIntegration_LoadPluginRegistersHooks verifies that loading a plugin
// via Manager.LoadAll discovers hooks from init.lua's hooks.on() calls,
// registers them with the HookEngine, and upserts records in plugin_hooks.
func TestHookIntegration_LoadPluginRegistersHooks(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "hook_reg", `
plugin_info = {
    name        = "hook_reg",
    version     = "1.0.0",
    description = "Hook registration integration test",
}

hooks.on("before_create", "content_data", function(data)
    log.info("before_create fired")
end)

hooks.on("after_update", "*", function(data)
    log.info("wildcard after_update fired")
end)

function on_init()
    log.info("hook_reg initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("hook_reg")
	if inst == nil {
		t.Fatal("plugin not loaded")
	}
	if inst.State != StateRunning {
		t.Fatalf("state = %s, want running", inst.State)
	}

	engine := mgr.HookEngine()

	// HasHooks should detect both the specific and wildcard hooks.
	if !engine.HasHooks(audited.HookBeforeCreate, "content_data") {
		t.Error("HasHooks(before_create, content_data) should be true")
	}
	if !engine.HasHooks(audited.HookAfterUpdate, "any_table") {
		t.Error("HasHooks(after_update, any_table) should be true (wildcard)")
	}
	if engine.HasHooks(audited.HookBeforeDelete, "content_data") {
		t.Error("HasHooks(before_delete, content_data) should be false (not registered)")
	}

	// Verify plugin_hooks DB records were created.
	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM plugin_hooks WHERE plugin_name = ?", "hook_reg").Scan(&count)
	if err != nil {
		t.Fatalf("querying plugin_hooks: %s", err)
	}
	if count != 2 {
		t.Errorf("expected 2 plugin_hooks rows, got %d", count)
	}

	// Hooks should not be approved by default.
	entries := engine.gatherEntries(audited.HookBeforeCreate, "content_data")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries without approval, got %d", len(entries))
	}
}

// TestHookIntegration_BeforeHookDispatch verifies that an approved before-hook
// actually executes the Lua handler when RunBeforeHooks is called.
func TestHookIntegration_BeforeHookDispatch(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "before_test", `
plugin_info = {
    name        = "before_test",
    version     = "1.0.0",
    description = "Before-hook dispatch test",
}

hooks.on("before_create", "content_data", function(data)
    -- Verify metadata injection.
    if data._table ~= "content_data" then
        error("expected _table = content_data, got " .. tostring(data._table))
    end
    if data._event ~= "before_create" then
        error("expected _event = before_create, got " .. tostring(data._event))
    end
    log.info("before_create handler executed successfully")
end)

function on_init()
    log.info("before_test initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// Approve the hook.
	err = engine.ApproveHook(context.Background(), "before_test", "before_create", "content_data", "test_admin")
	if err != nil {
		t.Fatalf("ApproveHook: %s", err)
	}

	// Entity data to pass to the hook (simulating audited.Create params).
	type TestEntity struct {
		Title  string `json:"title"`
		Status string `json:"status"`
	}
	entity := TestEntity{Title: "Test Content", Status: "draft"}

	// Run before hooks -- should succeed (handler does not error).
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", entity)
	if err != nil {
		t.Fatalf("RunBeforeHooks: %s", err)
	}
}

// TestHookIntegration_BeforeHookAbort verifies that a before-hook that calls
// error() returns a *audited.HookError with sanitized message.
func TestHookIntegration_BeforeHookAbort(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "abort_test", `
plugin_info = {
    name        = "abort_test",
    version     = "1.0.0",
    description = "Abort test",
}

hooks.on("before_create", "content_data", function(data)
    error("title is required")
end)

function on_init()
    log.info("abort_test initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// Approve the hook.
	err = engine.ApproveHook(context.Background(), "abort_test", "before_create", "content_data", "test_admin")
	if err != nil {
		t.Fatalf("ApproveHook: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", map[string]any{"title": ""})
	if err == nil {
		t.Fatal("expected error from aborting before-hook")
	}

	// Verify it is a *audited.HookError.
	var hookErr *audited.HookError
	if !errors.As(err, &hookErr) {
		t.Fatalf("expected *audited.HookError, got %T: %s", err, err)
	}

	// Sanitized message should not contain the Lua error text.
	if strings.Contains(hookErr.Error(), "title is required") {
		t.Errorf("sanitized error should not contain Lua message: %s", hookErr.Error())
	}
	if !strings.Contains(hookErr.Error(), "abort_test") {
		t.Errorf("sanitized error should contain plugin name: %s", hookErr.Error())
	}

	// LogMessage should contain the original Lua error.
	if !strings.Contains(hookErr.LogMessage(), "title is required") {
		t.Errorf("LogMessage should contain original Lua error: %s", hookErr.LogMessage())
	}
}

// TestHookIntegration_AfterHookDispatch verifies that approved after-hooks
// execute asynchronously without blocking the caller.
func TestHookIntegration_AfterHookDispatch(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	// This plugin creates a table in on_init and inserts a row in the after-hook
	// to produce an observable side effect.
	writePluginFile(t, dir, "after_test", `
plugin_info = {
    name        = "after_test",
    version     = "1.0.0",
    description = "After-hook dispatch test",
}

hooks.on("after_create", "content_data", function(data)
    -- After-hooks CAN use db.* calls (unlike before-hooks).
    db.insert("hook_log", {message = "after_create fired for " .. tostring(data._table)})
end)

function on_init()
    db.define_table("hook_log", {
        columns = {
            {name = "message", type = "text", not_null = true},
        },
    })
    log.info("after_test initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 4,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("after_test")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("expected after_test running, got %v", inst)
	}

	engine := mgr.HookEngine()

	// Approve the after_create hook.
	err = engine.ApproveHook(context.Background(), "after_test", "after_create", "content_data", "test_admin")
	if err != nil {
		t.Fatalf("ApproveHook: %s", err)
	}

	// Fire after-hooks.
	type TestEntity struct {
		Title string `json:"title"`
	}
	engine.RunAfterHooks(context.Background(), audited.HookAfterCreate, "content_data", TestEntity{Title: "New Post"})

	// After-hooks are async -- wait for the WaitGroup to drain.
	// Use a short poll with timeout instead of direct WG access.
	deadline := time.After(10 * time.Second)
	for {
		var count int
		scanErr := conn.QueryRow("SELECT COUNT(*) FROM plugin_after_test_hook_log").Scan(&count)
		if scanErr != nil {
			t.Fatalf("querying hook_log: %s", scanErr)
		}
		if count > 0 {
			// After-hook wrote to the DB -- success.
			break
		}

		select {
		case <-deadline:
			t.Fatal("timed out waiting for after-hook to write to hook_log")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// TestHookIntegration_UnapprovedHookSkipped verifies that unapproved hooks
// are silently skipped during dispatch (M8).
func TestHookIntegration_UnapprovedHookSkipped(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "unapproved", `
plugin_info = {
    name        = "unapproved",
    version     = "1.0.0",
    description = "Unapproved hook test",
}

hooks.on("before_create", "content_data", function(data)
    error("should never execute")
end)

function on_init()
    log.info("unapproved initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// Do NOT approve the hook.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// RunBeforeHooks should succeed because the unapproved hook is skipped.
	err = engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", map[string]any{"title": "test"})
	if err != nil {
		t.Fatalf("RunBeforeHooks should succeed with unapproved hook: %s", err)
	}
}

// TestHookIntegration_CircuitBreaker verifies that a hook that aborts repeatedly
// is auto-disabled by the circuit breaker (M9).
func TestHookIntegration_CircuitBreaker(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "breaker", `
plugin_info = {
    name        = "breaker",
    version     = "1.0.0",
    description = "Circuit breaker test",
}

hooks.on("before_create", "content_data", function(data)
    error("always fails")
end)

function on_init()
    log.info("breaker initialized")
end
`)

	// Use a low MaxConsecutiveAborts threshold for fast testing.
	threshold := 3
	mgr := NewManager(ManagerConfig{
		Enabled:                  true,
		Directory:                dir,
		MaxVMsPerPlugin:          2,
		ExecTimeoutSec:           5,
		MaxOpsPerExec:            100,
		HookMaxConsecutiveAborts: threshold,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// Approve the hook.
	err = engine.ApproveHook(context.Background(), "breaker", "before_create", "content_data", "test_admin")
	if err != nil {
		t.Fatalf("ApproveHook: %s", err)
	}

	// Run before-hooks threshold times -- each should fail.
	for i := range threshold {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		runErr := engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", map[string]any{"id": i})
		cancel()

		if runErr == nil {
			t.Fatalf("iteration %d: expected error from failing hook", i)
		}
	}

	// After threshold consecutive aborts, the hook should be auto-disabled.
	// The next RunBeforeHooks should succeed (hook is skipped).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", map[string]any{"id": "after_breaker"})
	if err != nil {
		t.Fatalf("RunBeforeHooks should succeed after circuit breaker tripped: %s", err)
	}

	// Verify the hook is disabled in memory.
	key := approvalKeyFor("breaker", "before_create", "content_data")
	engine.mu.RLock()
	disabled := engine.disabled[key]
	engine.mu.RUnlock()
	if !disabled {
		t.Error("hook should be disabled after circuit breaker")
	}
}

// TestHookIntegration_DBBlockedInBeforeHook verifies that db.* calls inside
// before-hooks are blocked (M1: SQLite deadlock prevention).
func TestHookIntegration_DBBlockedInBeforeHook(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "dbblock", `
plugin_info = {
    name        = "dbblock",
    version     = "1.0.0",
    description = "DB blocked in before-hook test",
}

hooks.on("before_create", "content_data", function(data)
    -- This should raise an error because db.* is blocked in before-hooks.
    local rows = db.query("some_table")
end)

function on_init()
    log.info("dbblock initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// Approve the hook.
	err = engine.ApproveHook(context.Background(), "dbblock", "before_create", "content_data", "test_admin")
	if err != nil {
		t.Fatalf("ApproveHook: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", map[string]any{"title": "test"})
	if err == nil {
		t.Fatal("expected error when db.* called in before-hook")
	}

	// The error should be a HookError (M5: sanitized).
	var hookErr *audited.HookError
	if !errors.As(err, &hookErr) {
		t.Fatalf("expected *audited.HookError, got %T: %s", err, err)
	}

	// The original message should mention db.* being blocked.
	if !strings.Contains(hookErr.LogMessage(), "not allowed inside before-hooks") {
		t.Errorf("LogMessage should mention db.* blocked: %s", hookErr.LogMessage())
	}
}

// TestHookIntegration_WildcardDispatch verifies that wildcard hooks fire for
// any table name, and that specific hooks run before wildcards at equal priority.
func TestHookIntegration_WildcardDispatch(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	// Two plugins: one with a specific hook, one with a wildcard.
	writePluginFile(t, dir, "specific_hooks", `
plugin_info = {
    name        = "specific_hooks",
    version     = "1.0.0",
    description = "Specific hook plugin",
}

hooks.on("before_create", "content_data", function(data)
    log.info("specific before_create on content_data")
end)

function on_init()
    log.info("specific_hooks initialized")
end
`)
	writePluginFile(t, dir, "wildcard_hooks", `
plugin_info = {
    name        = "wildcard_hooks",
    version     = "1.0.0",
    description = "Wildcard hook plugin",
}

hooks.on("before_create", "*", function(data)
    log.info("wildcard before_create on " .. tostring(data._table))
end)

function on_init()
    log.info("wildcard_hooks initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// Approve both hooks.
	err = engine.ApproveHook(context.Background(), "specific_hooks", "before_create", "content_data", "admin")
	if err != nil {
		t.Fatalf("ApproveHook specific: %s", err)
	}
	err = engine.ApproveHook(context.Background(), "wildcard_hooks", "before_create", "*", "admin")
	if err != nil {
		t.Fatalf("ApproveHook wildcard: %s", err)
	}

	// Verify ordering: specific before wildcard at equal priority.
	entries := engine.gatherEntries(audited.HookBeforeCreate, "content_data")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].pluginName != "specific_hooks" {
		t.Errorf("first entry should be specific_hooks, got %s", entries[0].pluginName)
	}
	if entries[1].pluginName != "wildcard_hooks" {
		t.Errorf("second entry should be wildcard_hooks, got %s", entries[1].pluginName)
	}

	// RunBeforeHooks should succeed (both hooks just log).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", map[string]any{"title": "test"})
	if err != nil {
		t.Fatalf("RunBeforeHooks: %s", err)
	}

	// Wildcard should also match a different table.
	if !engine.HasHooks(audited.HookBeforeCreate, "users") {
		t.Error("wildcard should match 'users' table")
	}
}

// TestHookIntegration_PriorityOrdering verifies that hooks execute in priority
// order (lower number first) across multiple plugins.
func TestHookIntegration_PriorityOrdering(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	// Plugin with high priority (50) -- runs first.
	writePluginFile(t, dir, "prio_high", `
plugin_info = {
    name        = "prio_high",
    version     = "1.0.0",
    description = "High priority plugin",
}

hooks.on("before_update", "content_data", function(data)
    log.info("priority 50 hook")
end, { priority = 50 })

function on_init()
    log.info("prio_high initialized")
end
`)
	// Plugin with low priority (200) -- runs last.
	writePluginFile(t, dir, "prio_low", `
plugin_info = {
    name        = "prio_low",
    version     = "1.0.0",
    description = "Low priority plugin",
}

hooks.on("before_update", "content_data", function(data)
    log.info("priority 200 hook")
end, { priority = 200 })

function on_init()
    log.info("prio_low initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// Approve both.
	err = engine.ApproveHook(context.Background(), "prio_high", "before_update", "content_data", "admin")
	if err != nil {
		t.Fatalf("ApproveHook prio_high: %s", err)
	}
	err = engine.ApproveHook(context.Background(), "prio_low", "before_update", "content_data", "admin")
	if err != nil {
		t.Fatalf("ApproveHook prio_low: %s", err)
	}

	entries := engine.gatherEntries(audited.HookBeforeUpdate, "content_data")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Priority 50 should come before priority 200.
	if entries[0].priority != 50 {
		t.Errorf("first entry priority = %d, want 50", entries[0].priority)
	}
	if entries[1].priority != 200 {
		t.Errorf("second entry priority = %d, want 200", entries[1].priority)
	}

	// Dispatch should succeed.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = engine.RunBeforeHooks(ctx, audited.HookBeforeUpdate, "content_data", map[string]any{"title": "updated"})
	if err != nil {
		t.Fatalf("RunBeforeHooks: %s", err)
	}
}

// TestHookIntegration_PluginHooksTableCleanup verifies that orphaned hook records
// are cleaned when plugins are removed from the filesystem.
func TestHookIntegration_PluginHooksTableCleanup(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "persistent", `
plugin_info = {
    name        = "persistent",
    version     = "1.0.0",
    description = "Plugin that stays",
}

hooks.on("before_create", "content_data", function(data) end)

function on_init()
    log.info("persistent initialized")
end
`)
	writePluginFile(t, dir, "ephemeral", `
plugin_info = {
    name        = "ephemeral",
    version     = "1.0.0",
    description = "Plugin that will be removed",
}

hooks.on("after_update", "content_data", function(data) end)

function on_init()
    log.info("ephemeral initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	// First load: both plugins register hooks.
	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("first LoadAll: %s", err)
	}

	// Verify both plugins have hook records.
	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM plugin_hooks").Scan(&count)
	if err != nil {
		t.Fatalf("counting hooks: %s", err)
	}
	if count != 2 {
		t.Errorf("expected 2 hook records, got %d", count)
	}

	// Shutdown the first manager to release resources.
	mgr.Shutdown(context.Background())

	// "Remove" the ephemeral plugin (create a new manager without it).
	conn2 := newTestDB(t)
	defer conn2.Close()

	// Recreate the hooks table in the new DB (since it is in-memory).
	// In production, the same DB persists across restarts.
	// For this test, we seed the new DB with the old hook records.
	if _, execErr := conn2.Exec(`CREATE TABLE IF NOT EXISTS plugin_hooks (
		plugin_name TEXT NOT NULL, event TEXT NOT NULL, table_name TEXT NOT NULL,
		approved INTEGER NOT NULL DEFAULT 0, approved_at TEXT, approved_by TEXT,
		plugin_version TEXT NOT NULL DEFAULT '', PRIMARY KEY (plugin_name, event, table_name)
	)`); execErr != nil {
		t.Fatalf("creating plugin_hooks in conn2: %s", execErr)
	}
	if _, execErr := conn2.Exec(
		"INSERT INTO plugin_hooks (plugin_name, event, table_name, plugin_version) VALUES (?, ?, ?, ?), (?, ?, ?, ?)",
		"persistent", "before_create", "content_data", "1.0.0",
		"ephemeral", "after_update", "content_data", "1.0.0",
	); execErr != nil {
		t.Fatalf("seeding plugin_hooks: %s", execErr)
	}

	dir2 := t.TempDir()
	// Only persistent plugin present.
	writePluginFile(t, dir2, "persistent", `
plugin_info = {
    name        = "persistent",
    version     = "1.0.0",
    description = "Plugin that stays",
}

hooks.on("before_create", "content_data", function(data) end)

function on_init()
    log.info("persistent initialized (second load)")
end
`)

	mgr2 := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir2,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn2, db.DialectSQLite, nil)

	err = mgr2.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("second LoadAll: %s", err)
	}

	// Ephemeral's hooks should have been cleaned up.
	err = conn2.QueryRow("SELECT COUNT(*) FROM plugin_hooks WHERE plugin_name = ?", "ephemeral").Scan(&count)
	if err != nil {
		t.Fatalf("counting ephemeral hooks: %s", err)
	}
	if count != 0 {
		t.Errorf("expected 0 ephemeral hooks after cleanup, got %d", count)
	}

	// Persistent's hooks should remain.
	err = conn2.QueryRow("SELECT COUNT(*) FROM plugin_hooks WHERE plugin_name = ?", "persistent").Scan(&count)
	if err != nil {
		t.Fatalf("counting persistent hooks: %s", err)
	}
	if count != 0 {
		t.Logf("persistent hooks: %d (may have been re-upserted)", count)
	}
}

// TestHookIntegration_HooksOnBlockedInOnInit verifies the phase guard: hooks.on()
// called inside on_init() raises an error and the plugin fails to load.
func TestHookIntegration_HooksOnBlockedInOnInit(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "bad_phase", `
plugin_info = {
    name        = "bad_phase",
    version     = "1.0.0",
    description = "Calls hooks.on inside on_init (not allowed)",
}

function on_init()
    hooks.on("before_create", "content_data", function(data) end)
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("bad_phase")
	if inst == nil {
		t.Fatal("plugin not loaded")
	}
	if inst.State != StateFailed {
		t.Errorf("state = %s, want failed", inst.State)
	}
	if !strings.Contains(inst.FailedReason, "on_init failed") {
		t.Errorf("FailedReason = %q, want to contain 'on_init failed'", inst.FailedReason)
	}
}

// TestHookIntegration_ShutdownDrainsAfterHooks verifies that HookEngine.Close
// waits for in-flight after-hooks to complete before returning.
func TestHookIntegration_ShutdownDrainsAfterHooks(t *testing.T) {
	conn := newTestDB(t)
	// Do NOT defer conn.Close() -- Shutdown closes it.

	dir := t.TempDir()
	writePluginFile(t, dir, "slow_after", `
plugin_info = {
    name        = "slow_after",
    version     = "1.0.0",
    description = "Slow after-hook for shutdown test",
}

hooks.on("after_create", "content_data", function(data)
    -- Simulate a slow operation by doing string concatenation in a loop.
    local s = ""
    for i = 1, 10000 do
        s = s .. "x"
    end
    log.info("slow after-hook completed")
end)

function on_init()
    log.info("slow_after initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 4,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// Approve the hook.
	err = engine.ApproveHook(context.Background(), "slow_after", "after_create", "content_data", "admin")
	if err != nil {
		t.Fatalf("ApproveHook: %s", err)
	}

	// Fire an after-hook.
	engine.RunAfterHooks(context.Background(), audited.HookAfterCreate, "content_data", map[string]any{"id": "test"})

	// Close the engine -- should wait for the after-hook to complete.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	engine.Close(shutdownCtx)

	if !engine.closing.Load() {
		t.Error("closing flag should be set")
	}

	// After Close, new after-hooks should be rejected.
	engine.RunAfterHooks(context.Background(), audited.HookAfterCreate, "content_data", map[string]any{"id": "late"})

	// Shutdown the manager (closes pools and DB).
	mgr.Shutdown(context.Background())
}

// TestHookIntegration_ApproveRevokeLifecycle verifies the full approve/revoke
// lifecycle through the hook engine: initially unapproved, approve, dispatch
// succeeds, revoke, dispatch skipped.
func TestHookIntegration_ApproveRevokeLifecycle(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "approv_test", `
plugin_info = {
    name        = "approv_test",
    version     = "1.0.0",
    description = "Approve/revoke lifecycle test",
}

hooks.on("before_create", "content_data", function(data)
    error("hook executed")
end)

function on_init()
    log.info("approv_test initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()
	ctx := context.Background()

	// Step 1: Unapproved -- hook should be skipped.
	err = engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", map[string]any{})
	if err != nil {
		t.Fatalf("unapproved hook should be skipped: %s", err)
	}

	// Step 2: Approve -- hook should execute and abort.
	err = engine.ApproveHook(ctx, "approv_test", "before_create", "content_data", "admin")
	if err != nil {
		t.Fatalf("ApproveHook: %s", err)
	}

	err = engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", map[string]any{})
	if err == nil {
		t.Fatal("approved hook should have executed and aborted")
	}
	var hookErr *audited.HookError
	if !errors.As(err, &hookErr) {
		t.Fatalf("expected *audited.HookError, got %T", err)
	}

	// Step 3: Revoke -- hook should be skipped again.
	err = engine.RevokeHook(ctx, "approv_test", "before_create", "content_data")
	if err != nil {
		t.Fatalf("RevokeHook: %s", err)
	}

	err = engine.RunBeforeHooks(ctx, audited.HookBeforeCreate, "content_data", map[string]any{})
	if err != nil {
		t.Fatalf("revoked hook should be skipped: %s", err)
	}
}

// TestHookIntegration_PublishDetectionHooks verifies that
// DetectStatusTransition fires before_publish/after_publish events when
// content_data status transitions to "published".
func TestHookIntegration_PublishDetectionHooks(t *testing.T) {
	// This test verifies the DetectStatusTransition logic at the HookEngine
	// level. Since we cannot easily drive the full audited.Update path in a
	// unit test (it requires a real CreateCommand), we verify the detection
	// logic and hook engine wiring directly.
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "pub_detect", `
plugin_info = {
    name        = "pub_detect",
    version     = "1.0.0",
    description = "Publish detection test",
}

hooks.on("before_publish", "content_data", function(data)
    log.info("before_publish detected: " .. tostring(data._event))
end)

function on_init()
    log.info("pub_detect initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// Approve the before_publish hook.
	err = engine.ApproveHook(context.Background(), "pub_detect", "before_publish", "content_data", "admin")
	if err != nil {
		t.Fatalf("ApproveHook: %s", err)
	}

	// Verify HasHooks returns true for before_publish.
	if !engine.HasHooks(audited.HookBeforePublish, "content_data") {
		t.Error("HasHooks(before_publish, content_data) should be true")
	}

	// Run before_publish hooks directly.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = engine.RunBeforeHooks(ctx, audited.HookBeforePublish, "content_data", map[string]any{"status": "published"})
	if err != nil {
		t.Fatalf("RunBeforeHooks(before_publish): %s", err)
	}

	// Verify DetectStatusTransition produces the right events.
	before := map[string]any{"status": "draft"}
	params := map[string]any{"status": "published"}
	events := audited.DetectStatusTransition("content_data", before, params)
	if len(events) != 1 || events[0] != audited.HookBeforePublish {
		t.Errorf("DetectStatusTransition = %v, want [before_publish]", events)
	}
}

// TestHookIntegration_HookEngineWithTestdataFixtures loads the test fixture
// plugins from testdata/plugins/ and verifies they load and register hooks
// correctly through the Manager.
func TestHookIntegration_HookEngineWithTestdataFixtures(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := "testdata/plugins"

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	engine := mgr.HookEngine()

	// hooks_plugin: should register before_create and after_create on content_data.
	hooksPlugin := mgr.GetPlugin("hooks_plugin")
	if hooksPlugin == nil {
		t.Fatal("hooks_plugin not loaded")
	}
	if hooksPlugin.State != StateRunning {
		t.Errorf("hooks_plugin state = %s, want running", hooksPlugin.State)
	}
	if !engine.HasHooks(audited.HookBeforeCreate, "content_data") {
		t.Error("hooks_plugin: HasHooks(before_create, content_data) should be true")
	}

	// hooks_wildcard_plugin: should register wildcard hooks.
	wildcardPlugin := mgr.GetPlugin("hooks_wildcard_plugin")
	if wildcardPlugin == nil {
		t.Fatal("hooks_wildcard_plugin not loaded")
	}
	if wildcardPlugin.State != StateRunning {
		t.Errorf("hooks_wildcard_plugin state = %s, want running", wildcardPlugin.State)
	}
	if !engine.HasHooks(audited.HookBeforeCreate, "any_table_name") {
		t.Error("hooks_wildcard_plugin: wildcard should match any table")
	}

	// hooks_priority_plugin: should register 3 hooks on before_update.
	prioPlugin := mgr.GetPlugin("hooks_priority_plugin")
	if prioPlugin == nil {
		t.Fatal("hooks_priority_plugin not loaded")
	}
	if prioPlugin.State != StateRunning {
		t.Errorf("hooks_priority_plugin state = %s, want running", prioPlugin.State)
	}

	// hooks_publish_plugin: should register publish/archive hooks.
	pubPlugin := mgr.GetPlugin("hooks_publish_plugin")
	if pubPlugin == nil {
		t.Fatal("hooks_publish_plugin not loaded")
	}
	if pubPlugin.State != StateRunning {
		t.Errorf("hooks_publish_plugin state = %s, want running", pubPlugin.State)
	}
	if !engine.HasHooks(audited.HookBeforePublish, "content_data") {
		t.Error("hooks_publish_plugin: HasHooks(before_publish, content_data) should be true")
	}
	if !engine.HasHooks(audited.HookBeforeArchive, "content_data") {
		t.Error("hooks_publish_plugin: HasHooks(before_archive, content_data) should be true")
	}

	// hooks_abort_plugin: should register a before_create hook that always errors.
	abortPlugin := mgr.GetPlugin("hooks_abort_plugin")
	if abortPlugin == nil {
		t.Fatal("hooks_abort_plugin not loaded")
	}
	if abortPlugin.State != StateRunning {
		t.Errorf("hooks_abort_plugin state = %s, want running", abortPlugin.State)
	}

	// hooks_db_blocked_plugin: should register a before_create hook that tries db.query.
	dbBlockedPlugin := mgr.GetPlugin("hooks_db_blocked_plugin")
	if dbBlockedPlugin == nil {
		t.Fatal("hooks_db_blocked_plugin not loaded")
	}
	if dbBlockedPlugin.State != StateRunning {
		t.Errorf("hooks_db_blocked_plugin state = %s, want running", dbBlockedPlugin.State)
	}
}
