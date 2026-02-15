package plugin

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
)

// newTestHookEngine creates a HookEngine with a test Manager and in-memory DB.
func newTestHookEngine(t *testing.T) (*HookEngine, *Manager, *sql.DB) {
	t.Helper()
	conn := newTestDB(t)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       t.TempDir(),
		MaxVMsPerPlugin: 4,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	engine := mgr.HookEngine()
	return engine, mgr, conn
}

func TestHookEngine_HasHooks_NoHooks(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	if engine.HasHooks(audited.HookBeforeCreate, "content_data") {
		t.Error("HasHooks should return false when no hooks are registered")
	}
}

func TestHookEngine_RegisterHooks_HasHooks(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	pending := []PendingHook{
		{
			Event:      audited.HookBeforeCreate,
			Table:      "content_data",
			Priority:   100,
			IsWildcard: false,
			HandlerKey: "hook_before_create_content_data_1",
		},
	}

	engine.RegisterHooks("test_plugin", pending)

	if !engine.HasHooks(audited.HookBeforeCreate, "content_data") {
		t.Error("HasHooks should return true after registering before_create on content_data")
	}

	if engine.HasHooks(audited.HookBeforeUpdate, "content_data") {
		t.Error("HasHooks should return false for unregistered event")
	}

	if engine.HasHooks(audited.HookBeforeCreate, "other_table") {
		t.Error("HasHooks should return false for unregistered table")
	}
}

func TestHookEngine_WildcardHooks(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	pending := []PendingHook{
		{
			Event:      audited.HookBeforeCreate,
			Table:      "*",
			Priority:   100,
			IsWildcard: true,
			HandlerKey: "hook_before_create_wildcard_1",
		},
	}

	engine.RegisterHooks("test_plugin", pending)

	// Wildcard should match any table.
	if !engine.HasHooks(audited.HookBeforeCreate, "content_data") {
		t.Error("wildcard hook should match content_data")
	}
	if !engine.HasHooks(audited.HookBeforeCreate, "users") {
		t.Error("wildcard hook should match users")
	}
	if !engine.HasHooks(audited.HookBeforeCreate, "any_table") {
		t.Error("wildcard hook should match any_table")
	}

	// Different event should not match.
	if engine.HasHooks(audited.HookAfterCreate, "content_data") {
		t.Error("wildcard should not match different event")
	}
}

func TestHookEngine_hookEntryLess_Ordering(t *testing.T) {
	t.Run("lower priority first", func(t *testing.T) {
		a := hookEntry{priority: 50}
		b := hookEntry{priority: 100}
		if !hookEntryLess(a, b) {
			t.Error("priority 50 should sort before 100")
		}
		if hookEntryLess(b, a) {
			t.Error("priority 100 should not sort before 50")
		}
	})

	t.Run("specific before wildcard at equal priority", func(t *testing.T) {
		a := hookEntry{priority: 100, isWildcard: false}
		b := hookEntry{priority: 100, isWildcard: true}
		if !hookEntryLess(a, b) {
			t.Error("specific should sort before wildcard at equal priority")
		}
		if hookEntryLess(b, a) {
			t.Error("wildcard should not sort before specific at equal priority")
		}
	})

	t.Run("registration order as tiebreaker", func(t *testing.T) {
		a := hookEntry{priority: 100, isWildcard: false, regOrder: 0}
		b := hookEntry{priority: 100, isWildcard: false, regOrder: 1}
		if !hookEntryLess(a, b) {
			t.Error("earlier registration should sort first")
		}
		if hookEntryLess(b, a) {
			t.Error("later registration should sort after")
		}
	})
}

func TestHookEngine_GatherEntries_UnapprovedFiltered(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	pending := []PendingHook{
		{
			Event:      audited.HookBeforeCreate,
			Table:      "content_data",
			Priority:   100,
			IsWildcard: false,
			HandlerKey: "hook_before_create_content_data_1",
		},
	}

	engine.RegisterHooks("test_plugin", pending)

	// Without approval, gatherEntries should return nil.
	entries := engine.gatherEntries(audited.HookBeforeCreate, "content_data")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries without approval, got %d", len(entries))
	}

	// Approve the hook.
	key := approvalKeyFor("test_plugin", "before_create", "content_data")
	engine.mu.Lock()
	engine.approved[key] = true
	engine.mu.Unlock()

	entries = engine.gatherEntries(audited.HookBeforeCreate, "content_data")
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after approval, got %d", len(entries))
	}
}

func TestHookEngine_GatherEntries_DisabledFiltered(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	pending := []PendingHook{
		{
			Event:      audited.HookBeforeCreate,
			Table:      "content_data",
			Priority:   100,
			IsWildcard: false,
			HandlerKey: "hook_before_create_content_data_1",
		},
	}

	engine.RegisterHooks("test_plugin", pending)

	// Approve and then disable.
	key := approvalKeyFor("test_plugin", "before_create", "content_data")
	engine.mu.Lock()
	engine.approved[key] = true
	engine.disabled[key] = true
	engine.mu.Unlock()

	entries := engine.gatherEntries(audited.HookBeforeCreate, "content_data")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries when disabled, got %d", len(entries))
	}
}

func TestHookEngine_SetHookEnabled(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	pending := []PendingHook{
		{
			Event:      audited.HookBeforeCreate,
			Table:      "content_data",
			Priority:   100,
			IsWildcard: false,
			HandlerKey: "hook_before_create_content_data_1",
		},
	}

	engine.RegisterHooks("test_plugin", pending)

	// Approve hook.
	key := approvalKeyFor("test_plugin", "before_create", "content_data")
	engine.mu.Lock()
	engine.approved[key] = true
	engine.mu.Unlock()

	// Disable.
	engine.SetHookEnabled("test_plugin", "before_create", "content_data", false)

	engine.mu.RLock()
	disabled := engine.disabled[key]
	engine.mu.RUnlock()

	if !disabled {
		t.Error("hook should be disabled after SetHookEnabled(false)")
	}

	entries := engine.gatherEntries(audited.HookBeforeCreate, "content_data")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries when disabled, got %d", len(entries))
	}

	// Re-enable.
	engine.SetHookEnabled("test_plugin", "before_create", "content_data", true)

	engine.mu.RLock()
	disabled = engine.disabled[key]
	engine.mu.RUnlock()

	if disabled {
		t.Error("hook should be enabled after SetHookEnabled(true)")
	}

	entries = engine.gatherEntries(audited.HookBeforeCreate, "content_data")
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after re-enable, got %d", len(entries))
	}
}

func TestHookEngine_RunBeforeHooks_NoHooks(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	err := engine.RunBeforeHooks(context.Background(), audited.HookBeforeCreate, "content_data", nil)
	if err != nil {
		t.Errorf("RunBeforeHooks should succeed with no hooks: %s", err)
	}
}

func TestHookEngine_RunAfterHooks_NoHooks(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	// Should not panic with no hooks registered.
	engine.RunAfterHooks(context.Background(), audited.HookAfterCreate, "content_data", nil)
}

func TestHookEngine_RunAfterHooks_ClosingPreventsDispatch(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	pending := []PendingHook{
		{
			Event:      audited.HookAfterCreate,
			Table:      "content_data",
			Priority:   100,
			IsWildcard: false,
			HandlerKey: "hook_after_create_content_data_1",
		},
	}
	engine.RegisterHooks("test_plugin", pending)

	// Approve.
	key := approvalKeyFor("test_plugin", "after_create", "content_data")
	engine.mu.Lock()
	engine.approved[key] = true
	engine.mu.Unlock()

	// Close the engine.
	engine.Close(context.Background())

	// Should not dispatch after closing.
	engine.RunAfterHooks(context.Background(), audited.HookAfterCreate, "content_data", map[string]any{"id": "test"})
	// No way to assert non-dispatch other than confirming no panic.
}

func TestHookEngine_Close_Idempotent(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	// Close twice should not panic.
	engine.Close(context.Background())
	engine.Close(context.Background())
}

func TestHookEngine_CreatePluginHooksTable(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	ctx := context.Background()
	err := engine.CreatePluginHooksTable(ctx)
	if err != nil {
		t.Fatalf("CreatePluginHooksTable: %s", err)
	}

	// Verify table exists by doing a select.
	_, err = conn.ExecContext(ctx, "SELECT count(*) FROM plugin_hooks")
	if err != nil {
		t.Fatalf("plugin_hooks table not created: %s", err)
	}

	// Idempotent -- calling again should not fail.
	err = engine.CreatePluginHooksTable(ctx)
	if err != nil {
		t.Fatalf("second CreatePluginHooksTable: %s", err)
	}
}

func TestHookEngine_UpsertHookRegistrations(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	ctx := context.Background()
	if err := engine.CreatePluginHooksTable(ctx); err != nil {
		t.Fatalf("CreatePluginHooksTable: %s", err)
	}

	pending := []PendingHook{
		{
			Event:      audited.HookBeforeCreate,
			Table:      "content_data",
			Priority:   100,
			IsWildcard: false,
			HandlerKey: "hook_1",
		},
		{
			Event:      audited.HookAfterCreate,
			Table:      "*",
			Priority:   50,
			IsWildcard: true,
			HandlerKey: "hook_2",
		},
	}

	err := engine.UpsertHookRegistrations(ctx, "test_plugin", "1.0.0", pending)
	if err != nil {
		t.Fatalf("UpsertHookRegistrations: %s", err)
	}

	// Verify rows were inserted.
	var count int
	err = conn.QueryRowContext(ctx, "SELECT count(*) FROM plugin_hooks WHERE plugin_name = ?", "test_plugin").Scan(&count)
	if err != nil {
		t.Fatalf("counting plugin_hooks: %s", err)
	}
	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}

	// Hooks should not be approved by default.
	engine.mu.RLock()
	approved := engine.approved[approvalKeyFor("test_plugin", "before_create", "content_data")]
	engine.mu.RUnlock()
	if approved {
		t.Error("hook should not be approved by default")
	}
}

func TestHookEngine_ApproveAndRevokeHook(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	ctx := context.Background()
	if err := engine.CreatePluginHooksTable(ctx); err != nil {
		t.Fatalf("CreatePluginHooksTable: %s", err)
	}

	pending := []PendingHook{
		{
			Event:      audited.HookBeforeCreate,
			Table:      "content_data",
			Priority:   100,
			HandlerKey: "hook_1",
		},
	}

	if err := engine.UpsertHookRegistrations(ctx, "test_plugin", "1.0.0", pending); err != nil {
		t.Fatalf("UpsertHookRegistrations: %s", err)
	}

	// Approve.
	if err := engine.ApproveHook(ctx, "test_plugin", "before_create", "content_data", "admin"); err != nil {
		t.Fatalf("ApproveHook: %s", err)
	}

	engine.mu.RLock()
	approved := engine.approved[approvalKeyFor("test_plugin", "before_create", "content_data")]
	engine.mu.RUnlock()
	if !approved {
		t.Error("hook should be approved after ApproveHook")
	}

	// Verify in DB.
	var approvedInt int
	err := conn.QueryRowContext(ctx,
		"SELECT approved FROM plugin_hooks WHERE plugin_name = ? AND event = ? AND table_name = ?",
		"test_plugin", "before_create", "content_data",
	).Scan(&approvedInt)
	if err != nil {
		t.Fatalf("reading approval from DB: %s", err)
	}
	if approvedInt != 1 {
		t.Errorf("DB approved = %d, want 1", approvedInt)
	}

	// Revoke.
	if err := engine.RevokeHook(ctx, "test_plugin", "before_create", "content_data"); err != nil {
		t.Fatalf("RevokeHook: %s", err)
	}

	engine.mu.RLock()
	approved = engine.approved[approvalKeyFor("test_plugin", "before_create", "content_data")]
	engine.mu.RUnlock()
	if approved {
		t.Error("hook should not be approved after RevokeHook")
	}
}

func TestHookEngine_CleanupOrphanedHooks(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	ctx := context.Background()
	if err := engine.CreatePluginHooksTable(ctx); err != nil {
		t.Fatalf("CreatePluginHooksTable: %s", err)
	}

	// Insert hooks for two plugins.
	pending1 := []PendingHook{{Event: audited.HookBeforeCreate, Table: "t1", HandlerKey: "h1"}}
	pending2 := []PendingHook{{Event: audited.HookAfterCreate, Table: "t2", HandlerKey: "h2"}}

	if err := engine.UpsertHookRegistrations(ctx, "plugin_a", "1.0.0", pending1); err != nil {
		t.Fatalf("upsert plugin_a: %s", err)
	}
	if err := engine.UpsertHookRegistrations(ctx, "plugin_b", "1.0.0", pending2); err != nil {
		t.Fatalf("upsert plugin_b: %s", err)
	}

	// Only plugin_a is discovered.
	if err := engine.CleanupOrphanedHooks(ctx, []string{"plugin_a"}); err != nil {
		t.Fatalf("CleanupOrphanedHooks: %s", err)
	}

	// plugin_b's hooks should be deleted.
	var count int
	err := conn.QueryRowContext(ctx, "SELECT count(*) FROM plugin_hooks WHERE plugin_name = ?", "plugin_b").Scan(&count)
	if err != nil {
		t.Fatalf("counting: %s", err)
	}
	if count != 0 {
		t.Errorf("expected 0 hooks for plugin_b after cleanup, got %d", count)
	}

	// plugin_a's hooks should remain.
	err = conn.QueryRowContext(ctx, "SELECT count(*) FROM plugin_hooks WHERE plugin_name = ?", "plugin_a").Scan(&count)
	if err != nil {
		t.Fatalf("counting: %s", err)
	}
	if count != 1 {
		t.Errorf("expected 1 hook for plugin_a after cleanup, got %d", count)
	}
}

func TestHookEngine_CleanupOrphanedHooks_EmptyDiscovered(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	ctx := context.Background()
	if err := engine.CreatePluginHooksTable(ctx); err != nil {
		t.Fatalf("CreatePluginHooksTable: %s", err)
	}

	pending := []PendingHook{{Event: audited.HookBeforeCreate, Table: "t1", HandlerKey: "h1"}}
	if err := engine.UpsertHookRegistrations(ctx, "plugin_a", "1.0.0", pending); err != nil {
		t.Fatalf("upsert: %s", err)
	}

	// Empty discovered set should delete all.
	if err := engine.CleanupOrphanedHooks(ctx, []string{}); err != nil {
		t.Fatalf("CleanupOrphanedHooks: %s", err)
	}

	var count int
	err := conn.QueryRowContext(ctx, "SELECT count(*) FROM plugin_hooks").Scan(&count)
	if err != nil {
		t.Fatalf("counting: %s", err)
	}
	if count != 0 {
		t.Errorf("expected 0 hooks after cleanup, got %d", count)
	}
}

func TestHookEngine_MergedOrdering(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	// Register specific hook at priority 100 and wildcard at priority 100.
	// Specific should come before wildcard at the same priority.
	specificHook := []PendingHook{
		{
			Event:      audited.HookBeforeCreate,
			Table:      "content_data",
			Priority:   100,
			IsWildcard: false,
			HandlerKey: "specific_hook",
		},
	}
	wildcardHook := []PendingHook{
		{
			Event:      audited.HookBeforeCreate,
			Table:      "*",
			Priority:   100,
			IsWildcard: true,
			HandlerKey: "wildcard_hook",
		},
	}

	engine.RegisterHooks("plugin_a", specificHook)
	engine.RegisterHooks("plugin_b", wildcardHook)

	// Approve both.
	engine.mu.Lock()
	engine.approved[approvalKeyFor("plugin_a", "before_create", "content_data")] = true
	engine.approved[approvalKeyFor("plugin_b", "before_create", "*")] = true
	engine.mu.Unlock()

	entries := engine.gatherEntries(audited.HookBeforeCreate, "content_data")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].isWildcard {
		t.Error("specific hook should sort before wildcard at equal priority")
	}
	if !entries[1].isWildcard {
		t.Error("wildcard hook should sort after specific at equal priority")
	}
}

// TestDetectStatusTransition tests the M12 status transition detection.
func TestDetectStatusTransition(t *testing.T) {
	t.Run("non-content_data returns nil", func(t *testing.T) {
		events := audited.DetectStatusTransition("users", map[string]any{"status": "draft"}, map[string]any{"status": "published"})
		if events != nil {
			t.Errorf("expected nil for non-content_data, got %v", events)
		}
	})

	t.Run("publish transition detected", func(t *testing.T) {
		events := audited.DetectStatusTransition("content_data",
			map[string]any{"status": "draft"},
			map[string]any{"status": "published"},
		)
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0] != audited.HookBeforePublish {
			t.Errorf("expected before_publish, got %s", events[0])
		}
	})

	t.Run("archive transition detected", func(t *testing.T) {
		events := audited.DetectStatusTransition("content_data",
			map[string]any{"status": "published"},
			map[string]any{"status": "archived"},
		)
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0] != audited.HookBeforeArchive {
			t.Errorf("expected before_archive, got %s", events[0])
		}
	})

	t.Run("no status change returns nil", func(t *testing.T) {
		events := audited.DetectStatusTransition("content_data",
			map[string]any{"status": "published"},
			map[string]any{"title": "new title"},
		)
		if events != nil {
			t.Errorf("expected nil for no status change, got %v", events)
		}
	})

	t.Run("same status returns nil", func(t *testing.T) {
		events := audited.DetectStatusTransition("content_data",
			map[string]any{"status": "published"},
			map[string]any{"status": "published"},
		)
		if events != nil {
			t.Errorf("expected nil when status unchanged, got %v", events)
		}
	})

	t.Run("nil before returns nil", func(t *testing.T) {
		events := audited.DetectStatusTransition("content_data",
			nil,
			map[string]any{"status": "published"},
		)
		if events != nil {
			t.Errorf("expected nil when before is nil, got %v", events)
		}
	})
}

func TestBeforeToAfterEvent(t *testing.T) {
	tests := []struct {
		input    audited.HookEvent
		expected audited.HookEvent
	}{
		{audited.HookBeforeCreate, audited.HookAfterCreate},
		{audited.HookBeforeUpdate, audited.HookAfterUpdate},
		{audited.HookBeforeDelete, audited.HookAfterDelete},
		{audited.HookBeforePublish, audited.HookAfterPublish},
		{audited.HookBeforeArchive, audited.HookAfterArchive},
		{audited.HookAfterCreate, audited.HookAfterCreate}, // no mapping, returns self
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := audited.BeforeToAfterEvent(tt.input)
			if got != tt.expected {
				t.Errorf("BeforeToAfterEvent(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result, err := audited.StructToMap(nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("struct converted to map", func(t *testing.T) {
		type testStruct struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		result, err := audited.StructToMap(testStruct{Name: "test", Count: 42})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if result["name"] != "test" {
			t.Errorf("name = %v, want test", result["name"])
		}
		// json.Number is preserved via UseNumber().
		// The actual type depends on the decoder, but it should be convertible to string "42".
	})
}

func TestHookError(t *testing.T) {
	t.Run("Error returns sanitized message", func(t *testing.T) {
		he := audited.NewHookError("my_plugin", audited.HookBeforeCreate, "content_data", "internal lua error")
		msg := he.Error()
		if msg != `operation blocked by plugin "my_plugin"` {
			t.Errorf("Error() = %q, expected sanitized message", msg)
		}
	})

	t.Run("LogMessage returns original", func(t *testing.T) {
		he := audited.NewHookError("my_plugin", audited.HookBeforeCreate, "content_data", "internal lua error")
		if he.LogMessage() != "internal lua error" {
			t.Errorf("LogMessage() = %q, want original", he.LogMessage())
		}
	})

	t.Run("HookError is detectable via type assertion", func(t *testing.T) {
		he := audited.NewHookError("my_plugin", audited.HookBeforeCreate, "content_data", "test")
		var err error = he

		var hookErr *audited.HookError
		if !errors.As(err, &hookErr) {
			t.Error("errors.As should find *HookError")
		}
		if hookErr.PluginName != "my_plugin" {
			t.Errorf("PluginName = %q, want my_plugin", hookErr.PluginName)
		}
	})
}

func TestHookEngine_Close_WaitsForAfterHooks(t *testing.T) {
	engine, _, conn := newTestHookEngine(t)
	defer conn.Close()

	// Close with a short timeout should complete quickly when no hooks are running.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	engine.Close(ctx)

	if !engine.closing.Load() {
		t.Error("closing flag should be set after Close")
	}
}
