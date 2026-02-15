package plugin

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	_ "github.com/mattn/go-sqlite3"
	lua "github.com/yuin/gopher-lua"
)

// integrationDBSeq provides unique names for in-memory SQLite databases so
// parallel tests do not share state. Each call to newIntegrationDB gets a
// unique database via file:integN?mode=memory&cache=shared.
var integrationDBSeq atomic.Int64

// testPluginsDir returns the absolute path to testdata/plugins/ relative to this test file.
// Uses runtime.Caller to resolve the path regardless of the working directory.
func testPluginsDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed to resolve test file path")
	}
	return filepath.Join(filepath.Dir(filename), "testdata", "plugins")
}

// newIntegrationDB opens a fresh in-memory SQLite database with foreign keys
// enabled, matching the production plugin pool configuration.
// Uses file:integN?mode=memory&cache=shared with a unique name per call to
// ensure all connections from the sql.DB pool share the same in-memory database
// (required for concurrent goroutine access) while tests remain isolated from
// each other.
// Caller must defer conn.Close() unless Shutdown will close it.
func newIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	seq := integrationDBSeq.Add(1)
	dsn := fmt.Sprintf("file:integ%d?mode=memory&cache=shared", seq)
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("opening test db: %s", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		t.Fatalf("enabling foreign keys: %s", err)
	}
	return conn
}

// loadTestBookmarksManager creates a Manager pointing at testdata/plugins/ and
// calls LoadAll. Returns the manager and the db connection. The caller is
// responsible for calling mgr.Shutdown(ctx) which closes the db connection.
func loadTestBookmarksManager(t *testing.T, cfg ManagerConfig) (*Manager, *sql.DB) {
	t.Helper()
	conn := newIntegrationDB(t)

	if cfg.Directory == "" {
		cfg.Directory = testPluginsDir(t)
	}
	if cfg.MaxVMsPerPlugin <= 0 {
		cfg.MaxVMsPerPlugin = 2
	}
	if cfg.ExecTimeoutSec <= 0 {
		cfg.ExecTimeoutSec = 5
	}
	if cfg.MaxOpsPerExec <= 0 {
		cfg.MaxOpsPerExec = 1000
	}

	mgr := NewManager(cfg, conn, db.DialectSQLite)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		conn.Close()
		t.Fatalf("LoadAll: %s", err)
	}

	return mgr, conn
}

// TestIntegration_FullLifecycle verifies the end-to-end plugin discovery, loading,
// state management, and shutdown using the real test fixtures in testdata/plugins/.
//
// The testdata directory contains:
//   - test_bookmarks: valid plugin exercising full CRUD, transactions, require
//   - invalid_no_manifest: missing plugin_info (should be skipped during discovery)
//   - invalid_bad_name: plugin_info.name has spaces/uppercase (should be skipped)
//   - timeout_plugin: on_init has infinite loop (should reach StateFailed)
func TestIntegration_FullLifecycle(t *testing.T) {
	// Use a short timeout so the timeout_plugin fails quickly.
	mgr, conn := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:         true,
		ExecTimeoutSec:  2,
		MaxVMsPerPlugin: 2,
		MaxOpsPerExec:   1000,
	})
	// Shutdown closes the DB connection and all pools.
	defer mgr.Shutdown(context.Background())

	t.Run("test_bookmarks_is_running", func(t *testing.T) {
		inst := mgr.GetPlugin("test_bookmarks")
		if inst == nil {
			t.Fatal("GetPlugin returned nil for test_bookmarks")
		}
		if inst.State != StateRunning {
			t.Errorf("test_bookmarks state = %s, want running (FailedReason: %s)",
				inst.State, inst.FailedReason)
		}
	})

	t.Run("invalid_no_manifest_not_loaded", func(t *testing.T) {
		// invalid_no_manifest has no plugin_info global, so manifest extraction
		// fails and the plugin is never added to the manager's plugin map.
		inst := mgr.GetPlugin("invalid_no_manifest")
		if inst != nil {
			t.Errorf("expected invalid_no_manifest to be absent from plugins, got state=%s", inst.State)
		}
	})

	t.Run("invalid_bad_name_not_loaded", func(t *testing.T) {
		// invalid_bad_name has name "Has Spaces And CAPS" which fails validation.
		// It is never added to the manager's plugin map.
		inst := mgr.GetPlugin("Has Spaces And CAPS")
		if inst != nil {
			t.Errorf("expected invalid_bad_name to be absent from plugins, got state=%s", inst.State)
		}
	})

	t.Run("timeout_plugin_is_failed", func(t *testing.T) {
		inst := mgr.GetPlugin("timeout_plugin")
		if inst == nil {
			t.Fatal("GetPlugin returned nil for timeout_plugin")
		}
		if inst.State != StateFailed {
			t.Errorf("timeout_plugin state = %s, want failed", inst.State)
		}
		if inst.FailedReason == "" {
			t.Error("timeout_plugin FailedReason is empty, expected non-empty")
		}
	})

	t.Run("tables_created_with_correct_prefix", func(t *testing.T) {
		// The test_bookmarks plugin creates two tables: collections and bookmarks.
		// They should be prefixed with plugin_test_bookmarks_.
		for _, tableName := range []string{
			"plugin_test_bookmarks_collections",
			"plugin_test_bookmarks_bookmarks",
		} {
			var count int
			err := conn.QueryRow(
				"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
				tableName,
			).Scan(&count)
			if err != nil {
				t.Fatalf("checking table %q: %s", tableName, err)
			}
			if count != 1 {
				t.Errorf("table %q not found in sqlite_master", tableName)
			}
		}
	})

	t.Run("seed_data_inserted_by_on_init", func(t *testing.T) {
		// on_init inserts 2 collections and ends with 3 bookmarks (after
		// transaction commit/rollback cycle).
		var colCount int
		err := conn.QueryRow("SELECT COUNT(*) FROM plugin_test_bookmarks_collections").Scan(&colCount)
		if err != nil {
			t.Fatalf("counting collections: %s", err)
		}
		if colCount != 2 {
			t.Errorf("collection count = %d, want 2", colCount)
		}

		var bmCount int
		err = conn.QueryRow("SELECT COUNT(*) FROM plugin_test_bookmarks_bookmarks").Scan(&bmCount)
		if err != nil {
			t.Fatalf("counting bookmarks: %s", err)
		}
		if bmCount != 3 {
			t.Errorf("bookmark count = %d, want 3", bmCount)
		}
	})

	t.Run("auto_injected_columns_present", func(t *testing.T) {
		// Verify that id, created_at, updated_at were auto-injected into collections.
		row := conn.QueryRow(
			"SELECT id, created_at, updated_at FROM plugin_test_bookmarks_collections LIMIT 1",
		)
		var id, createdAt, updatedAt string
		err := row.Scan(&id, &createdAt, &updatedAt)
		if err != nil {
			t.Fatalf("scanning auto-injected columns: %s", err)
		}
		if id == "" {
			t.Error("id column is empty")
		}
		if createdAt == "" {
			t.Error("created_at column is empty")
		}
		if updatedAt == "" {
			t.Error("updated_at column is empty")
		}
	})

	t.Run("shutdown_transitions_to_stopped", func(t *testing.T) {
		// We need a fresh manager for the shutdown test since the deferred
		// Shutdown above would double-close. Use a separate db.
		shutdownConn := newIntegrationDB(t)
		shutdownMgr := NewManager(ManagerConfig{
			Enabled:         true,
			Directory:       testPluginsDir(t),
			MaxVMsPerPlugin: 2,
			ExecTimeoutSec:  2,
			MaxOpsPerExec:   1000,
		}, shutdownConn, db.DialectSQLite)

		err := shutdownMgr.LoadAll(context.Background())
		if err != nil {
			t.Fatalf("LoadAll: %s", err)
		}

		inst := shutdownMgr.GetPlugin("test_bookmarks")
		if inst == nil || inst.State != StateRunning {
			t.Fatalf("expected test_bookmarks running, got %v", inst)
		}

		shutdownMgr.Shutdown(context.Background())

		if inst.State != StateStopped {
			t.Errorf("state after shutdown = %s, want stopped", inst.State)
		}

		// timeout_plugin was StateFailed, so it should remain failed (not stopped).
		tp := shutdownMgr.GetPlugin("timeout_plugin")
		if tp == nil {
			t.Fatal("timeout_plugin not in plugins")
		}
		// StateFailed plugins are not in loadOrder, so shutdown skips them.
		// Their pool is still closed, but state remains StateFailed.
		if tp.State != StateFailed {
			t.Errorf("timeout_plugin state after shutdown = %s, want failed", tp.State)
		}
	})
}

// TestIntegration_PluginCRUD verifies that a running plugin's VM can execute
// all CRUD operations: insert, query, query_one, update, delete, count, exists.
func TestIntegration_PluginCRUD(t *testing.T) {
	mgr, _ := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:        true,
		ExecTimeoutSec: 5,
		MaxOpsPerExec:  1000,
	})
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("test_bookmarks")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("test_bookmarks not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)

	// Reset op count for this checkout.
	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}

	L.SetContext(ctx)

	// Execute CRUD operations as Lua code on the already-loaded VM.
	crudCode := `
		-- INSERT a new bookmark
		db.insert("bookmarks", {
			collection_id = "test_crud_col",
			url   = "https://example.com/crud-test",
			title = "CRUD Test Bookmark",
			rating = 3.5,
		})

		-- QUERY: verify it was inserted
		local results = db.query("bookmarks", {
			where = {url = "https://example.com/crud-test"},
		})
		if #results ~= 1 then
			error("expected 1 result from query, got " .. #results)
		end
		local row = results[1]
		if row.title ~= "CRUD Test Bookmark" then
			error("title mismatch: " .. tostring(row.title))
		end

		-- QUERY_ONE: verify single-row fetch
		local one = db.query_one("bookmarks", {
			where = {url = "https://example.com/crud-test"},
		})
		if not one then
			error("query_one returned nil")
		end
		if one.title ~= "CRUD Test Bookmark" then
			error("query_one title mismatch: " .. tostring(one.title))
		end

		-- COUNT before update
		local before_count = db.count("bookmarks", {
			where = {url = "https://example.com/crud-test"},
		})
		if before_count ~= 1 then
			error("expected count 1, got " .. tostring(before_count))
		end

		-- EXISTS: verify
		local found = db.exists("bookmarks", {
			where = {url = "https://example.com/crud-test"},
		})
		if not found then
			error("exists returned false for known row")
		end

		-- UPDATE: change the title
		db.update("bookmarks", {
			set   = {title = "Updated CRUD Bookmark"},
			where = {url = "https://example.com/crud-test"},
		})

		-- Verify the update
		local updated = db.query_one("bookmarks", {
			where = {url = "https://example.com/crud-test"},
		})
		if not updated then
			error("query_one returned nil after update")
		end
		if updated.title ~= "Updated CRUD Bookmark" then
			error("update failed, title = " .. tostring(updated.title))
		end

		-- DELETE
		db.delete("bookmarks", {
			where = {url = "https://example.com/crud-test"},
		})

		-- Verify deletion
		local after_delete = db.count("bookmarks", {
			where = {url = "https://example.com/crud-test"},
		})
		if after_delete ~= 0 then
			error("expected 0 after delete, got " .. tostring(after_delete))
		end

		local gone = db.exists("bookmarks", {
			where = {url = "https://example.com/crud-test"},
		})
		if gone then
			error("exists returned true after deletion")
		end
	`

	if err := L.DoString(crudCode); err != nil {
		t.Fatalf("CRUD operations failed: %s", err)
	}
}

// TestIntegration_PluginTransaction verifies that db.transaction() correctly
// commits on success and rolls back on error within a loaded plugin VM.
func TestIntegration_PluginTransaction(t *testing.T) {
	mgr, _ := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:        true,
		ExecTimeoutSec: 5,
		MaxOpsPerExec:  1000,
	})
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("test_bookmarks")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("test_bookmarks not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)

	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}
	L.SetContext(ctx)

	t.Run("commit", func(t *testing.T) {
		commitCode := `
			-- Get current count before the transaction
			local before = db.count("bookmarks", {})

			local ok, err = db.transaction(function()
				db.insert("bookmarks", {
					collection_id = "tx_test",
					url   = "https://example.com/tx-commit-1",
					title = "TX Commit 1",
				})
				db.insert("bookmarks", {
					collection_id = "tx_test",
					url   = "https://example.com/tx-commit-2",
					title = "TX Commit 2",
				})
			end)

			if not ok then
				error("transaction commit failed: " .. tostring(err))
			end

			-- Both rows should exist after commit
			local after = db.count("bookmarks", {})
			if after ~= before + 2 then
				error("expected count to increase by 2, before=" .. tostring(before) .. " after=" .. tostring(after))
			end
		`
		if err := L.DoString(commitCode); err != nil {
			t.Fatalf("transaction commit test failed: %s", err)
		}
	})

	t.Run("rollback", func(t *testing.T) {
		rollbackCode := `
			local before = db.count("bookmarks", {})

			local ok, err = db.transaction(function()
				db.insert("bookmarks", {
					collection_id = "tx_test",
					url   = "https://example.com/tx-rollback",
					title = "Should Not Persist",
				})
				error("deliberate rollback")
			end)

			if ok then
				error("expected transaction to fail but it succeeded")
			end

			-- Row should NOT exist after rollback
			local after = db.count("bookmarks", {})
			if after ~= before then
				error("rollback failed: count changed from " .. tostring(before) .. " to " .. tostring(after))
			end

			local ghost = db.exists("bookmarks", {
				where = {url = "https://example.com/tx-rollback"},
			})
			if ghost then
				error("rolled-back row still exists")
			end
		`
		if err := L.DoString(rollbackCode); err != nil {
			t.Fatalf("transaction rollback test failed: %s", err)
		}
	})
}

// TestIntegration_VMPoolReuse verifies that VMs are properly recycled:
// op count is reset, and globals set during one checkout are cleaned on return.
func TestIntegration_VMPoolReuse(t *testing.T) {
	mgr, _ := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:        true,
		ExecTimeoutSec: 5,
		MaxOpsPerExec:  1000,
	})
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("test_bookmarks")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("test_bookmarks not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First checkout: set a global and do some work.
	L1, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("first Pool.Get: %s", err)
	}

	if dbAPI, ok := inst.dbAPIs[L1]; ok {
		dbAPI.ResetOpCount()
	}
	L1.SetContext(ctx)

	// Set a rogue global (simulating plugin forgetting `local`).
	if err := L1.DoString(`leaked_global = "should_be_cleaned"`); err != nil {
		t.Fatalf("setting leaked global: %s", err)
	}

	// Do some db operations to increment the op count.
	if err := L1.DoString(`local n = db.count("bookmarks", {})`); err != nil {
		t.Fatalf("db.count: %s", err)
	}

	// Verify op count is non-zero before return.
	if dbAPI, ok := inst.dbAPIs[L1]; ok {
		if dbAPI.opCount == 0 {
			t.Error("opCount should be > 0 after db operation")
		}
	}

	// Return VM to pool. This triggers restoreGlobalSnapshot.
	inst.Pool.Put(L1)

	// Second checkout: may get the same or different VM.
	L2, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("second Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L2)

	// Reset op count for the new checkout (as the Manager would do).
	if dbAPI, ok := inst.dbAPIs[L2]; ok {
		dbAPI.ResetOpCount()

		// Verify op count was reset.
		if dbAPI.opCount != 0 {
			t.Errorf("opCount after reset = %d, want 0", dbAPI.opCount)
		}
	}

	L2.SetContext(ctx)

	// Verify the leaked global was cleaned up by restoreGlobalSnapshot.
	// This only works if L2 is the same VM as L1 (which it likely is with
	// pool size 2 and one VM checked out). If it's a different VM, the global
	// was never set on it, which also proves isolation.
	leaked := L2.GetGlobal("leaked_global")
	if leaked != lua.LNil {
		t.Errorf("leaked_global should be nil after pool return, got %s", leaked.String())
	}
}

// TestIntegration_OpBudgetEnforcement verifies that the per-checkout operation
// budget is enforced: a plugin that exceeds MaxOpsPerExec gets an error raised.
func TestIntegration_OpBudgetEnforcement(t *testing.T) {
	// Use a very low op budget so we can trigger it quickly.
	conn := newIntegrationDB(t)
	dir := t.TempDir()

	// Create a simple plugin that defines a table in on_init.
	writePluginFile(t, dir, "budget_test", `
plugin_info = {
    name        = "budget_test",
    version     = "1.0.0",
    description = "Tests op budget enforcement",
}

function on_init()
    db.define_table("items", {
        columns = {
            {name = "title", type = "text", not_null = true},
        },
    })
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   5, // Very low budget: 5 operations max
	}, conn, db.DialectSQLite)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("budget_test")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("budget_test not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)

	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}
	L.SetContext(ctx)

	// Attempt more than 5 db operations. Each insert counts as 1 operation.
	// With MaxOpsPerExec=5, the 6th operation should fail.
	budgetCode := `
		local ok_count = 0
		for i = 1, 10 do
			local success, errmsg = pcall(function()
				db.insert("items", {title = "item " .. tostring(i)})
			end)
			if success then
				ok_count = ok_count + 1
			else
				-- The error message should mention exceeded maximum operations
				if not string.find(tostring(errmsg), "exceeded maximum operations") then
					error("unexpected error: " .. tostring(errmsg))
				end
				return ok_count
			end
		end
		error("expected op budget to be exceeded, but all 10 inserts succeeded (ok_count=" .. ok_count .. ")")
	`

	if err := L.DoString(budgetCode); err != nil {
		t.Fatalf("op budget test failed: %s", err)
	}
}

// TestIntegration_SandboxedRequire verifies that the sandboxed require() loader
// works correctly in the context of a full plugin load: the test_bookmarks plugin
// uses require("validators") which loads from lib/validators.lua.
func TestIntegration_SandboxedRequire(t *testing.T) {
	mgr, _ := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:        true,
		ExecTimeoutSec: 5,
		MaxOpsPerExec:  1000,
	})
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("test_bookmarks")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("test_bookmarks not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)
	L.SetContext(ctx)

	// The validators module should be cached from init.lua. Verify it's accessible
	// and its functions work.
	requireCode := `
		local v = require("validators")
		if type(v) ~= "table" then
			error("require('validators') returned " .. type(v) .. ", expected table")
		end
		if not v.is_valid_url("https://example.com") then
			error("is_valid_url returned false for valid URL")
		end
		if v.is_valid_url("not-a-url") then
			error("is_valid_url returned true for invalid URL")
		end
		if v.trim("  hello  ") ~= "hello" then
			error("trim failed: got '" .. v.trim("  hello  ") .. "'")
		end
	`
	if err := L.DoString(requireCode); err != nil {
		t.Fatalf("sandboxed require test failed: %s", err)
	}
}

// TestIntegration_FrozenModules verifies that db and log modules are frozen
// (read-only) on loaded plugin VMs.
func TestIntegration_FrozenModules(t *testing.T) {
	mgr, _ := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:        true,
		ExecTimeoutSec: 5,
		MaxOpsPerExec:  1000,
	})
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("test_bookmarks")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("test_bookmarks not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)
	L.SetContext(ctx)

	t.Run("db_module_frozen", func(t *testing.T) {
		err := L.DoString(`
			local ok, err = pcall(function()
				db.query = nil
			end)
			if ok then
				error("expected write to frozen db module to fail")
			end
		`)
		if err != nil {
			t.Fatalf("frozen db module test failed: %s", err)
		}
	})

	t.Run("log_module_frozen", func(t *testing.T) {
		err := L.DoString(`
			local ok, err = pcall(function()
				log.info = nil
			end)
			if ok then
				error("expected write to frozen log module to fail")
			end
		`)
		if err != nil {
			t.Fatalf("frozen log module test failed: %s", err)
		}
	})

	t.Run("db_reads_still_work", func(t *testing.T) {
		err := L.DoString(`
			local n = db.count("bookmarks", {})
			if type(n) ~= "number" then
				error("db.count returned " .. type(n) .. ", expected number")
			end
		`)
		if err != nil {
			t.Fatalf("frozen module read test failed: %s", err)
		}
	})

	t.Run("getmetatable_returns_protected", func(t *testing.T) {
		err := L.DoString(`
			local mt = getmetatable(db)
			if mt ~= "protected" then
				error("getmetatable(db) returned " .. tostring(mt) .. ", expected 'protected'")
			end
		`)
		if err != nil {
			t.Fatalf("metatable protection test failed: %s", err)
		}
	})
}

// TestIntegration_SandboxBlocksDangerousGlobals verifies that the sandbox strips
// dangerous globals from loaded plugin VMs.
func TestIntegration_SandboxBlocksDangerousGlobals(t *testing.T) {
	mgr, _ := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:        true,
		ExecTimeoutSec: 5,
		MaxOpsPerExec:  1000,
	})
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("test_bookmarks")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("test_bookmarks not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)
	L.SetContext(ctx)

	blockedGlobals := []string{
		"dofile", "loadfile", "load",
		"rawget", "rawset", "rawequal",
	}

	for _, name := range blockedGlobals {
		t.Run(name+"_is_nil", func(t *testing.T) {
			code := `
				if ` + name + ` ~= nil then
					error("` + name + ` should be nil in sandbox")
				end
			`
			if err := L.DoString(code); err != nil {
				t.Fatalf("sandbox check for %s failed: %s", name, err)
			}
		})
	}
}

// TestIntegration_ConcurrentVMCheckout verifies that multiple goroutines can
// check out VMs from the same plugin pool without data races or corruption.
func TestIntegration_ConcurrentVMCheckout(t *testing.T) {
	mgr, _ := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:         true,
		ExecTimeoutSec:  5,
		MaxVMsPerPlugin: 4,
		MaxOpsPerExec:   1000,
	})
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("test_bookmarks")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("test_bookmarks not running: %v", inst)
	}

	// Run 4 concurrent goroutines each doing a db.count on the same table.
	const numWorkers = 4
	errCh := make(chan error, numWorkers)

	for i := range numWorkers {
		go func(workerID int) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			L, getErr := inst.Pool.Get(ctx)
			if getErr != nil {
				errCh <- getErr
				return
			}
			defer inst.Pool.Put(L)

			if dbAPI, ok := inst.dbAPIs[L]; ok {
				dbAPI.ResetOpCount()
			}
			L.SetContext(ctx)

			code := `
				local n = db.count("collections", {})
				if type(n) ~= "number" then
					error("worker got non-number count")
				end
			`
			errCh <- L.DoString(code)
		}(i)
	}

	for range numWorkers {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent checkout error: %s", err)
		}
	}
}

// TestIntegration_TablePrefixIsolation verifies that plugins cannot access tables
// outside their namespace prefix. A Lua query referencing a non-prefixed table
// should fail validation.
func TestIntegration_TablePrefixIsolation(t *testing.T) {
	conn := newIntegrationDB(t)
	dir := t.TempDir()

	writePluginFile(t, dir, "isolated", `
plugin_info = {
    name        = "isolated",
    version     = "1.0.0",
    description = "Tests table prefix isolation",
}

function on_init()
    db.define_table("my_data", {
        columns = {
            {name = "value", type = "text"},
        },
    })
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("isolated")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("isolated not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)

	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}
	L.SetContext(ctx)

	// The plugin passes "my_data" which becomes "plugin_isolated_my_data" -- this should work.
	err = L.DoString(`db.insert("my_data", {value = "ok"})`)
	if err != nil {
		t.Fatalf("insert into own table failed: %s", err)
	}

	// But the prefixing is enforced by Go. The Lua code cannot bypass it because
	// every table name argument goes through prefixTable(). The table name
	// "plugin_isolated_my_data" is what actually gets sent to SQL. Even if the plugin
	// tries to query a raw core table name like "users", it becomes
	// "plugin_isolated_users" which does not exist.
	err = L.DoString(`
		local results, errmsg = db.query("users", {})
		-- This should return nil + error because plugin_isolated_users does not exist
		if results ~= nil then
			error("expected nil result for non-existent table, got a table")
		end
	`)
	if err != nil {
		t.Fatalf("table prefix isolation test failed: %s", err)
	}
}

// TestIntegration_EmptyWhereBlockedOnUpdateDelete verifies that db.update and
// db.delete raise errors when called without a where clause (safety guard).
func TestIntegration_EmptyWhereBlockedOnUpdateDelete(t *testing.T) {
	conn := newIntegrationDB(t)
	dir := t.TempDir()

	writePluginFile(t, dir, "where_test", `
plugin_info = {
    name        = "where_test",
    version     = "1.0.0",
    description = "Tests empty where rejection",
}

function on_init()
    db.define_table("items", {
        columns = {
            {name = "title", type = "text", not_null = true},
        },
    })
    db.insert("items", {title = "test item"})
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("where_test")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("where_test not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)

	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}
	L.SetContext(ctx)

	t.Run("update_without_where", func(t *testing.T) {
		err := L.DoString(`
			local ok, errmsg = pcall(function()
				db.update("items", {set = {title = "changed"}})
			end)
			if ok then
				error("expected error for update without where")
			end
			if not string.find(tostring(errmsg), "non%-empty where") then
				error("unexpected error message: " .. tostring(errmsg))
			end
		`)
		if err != nil {
			t.Fatalf("empty where update test failed: %s", err)
		}
	})

	t.Run("delete_without_where", func(t *testing.T) {
		err := L.DoString(`
			local ok, errmsg = pcall(function()
				db.delete("items", {})
			end)
			if ok then
				error("expected error for delete without where")
			end
			if not string.find(tostring(errmsg), "non%-empty where") then
				error("unexpected error message: " .. tostring(errmsg))
			end
		`)
		if err != nil {
			t.Fatalf("empty where delete test failed: %s", err)
		}
	})
}

// TestIntegration_DBHelpers verifies that db.ulid() and db.timestamp() work
// correctly in loaded plugin VMs.
func TestIntegration_DBHelpers(t *testing.T) {
	mgr, _ := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:        true,
		ExecTimeoutSec: 5,
		MaxOpsPerExec:  1000,
	})
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("test_bookmarks")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("test_bookmarks not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)
	L.SetContext(ctx)

	t.Run("ulid_generates_unique_ids", func(t *testing.T) {
		err := L.DoString(`
			local id1 = db.ulid()
			local id2 = db.ulid()
			if type(id1) ~= "string" or id1 == "" then
				error("db.ulid() returned invalid value: " .. tostring(id1))
			end
			if id1 == id2 then
				error("db.ulid() returned duplicate: " .. id1)
			end
			-- ULID should be 26 characters
			if #id1 ~= 26 then
				error("db.ulid() length = " .. #id1 .. ", expected 26")
			end
		`)
		if err != nil {
			t.Fatalf("db.ulid test failed: %s", err)
		}
	})

	t.Run("timestamp_returns_rfc3339", func(t *testing.T) {
		err := L.DoString(`
			local ts = db.timestamp()
			if type(ts) ~= "string" or ts == "" then
				error("db.timestamp() returned invalid value: " .. tostring(ts))
			end
			-- RFC3339 timestamps contain "T" and "Z" (for UTC)
			if not string.find(ts, "T") then
				error("timestamp missing T separator: " .. ts)
			end
			if not string.find(ts, "Z") then
				error("timestamp missing Z suffix (not UTC): " .. ts)
			end
		`)
		if err != nil {
			t.Fatalf("db.timestamp test failed: %s", err)
		}
	})
}

// TestIntegration_PluginLogAPI verifies that log functions are available and
// callable on loaded plugin VMs without errors.
func TestIntegration_PluginLogAPI(t *testing.T) {
	mgr, _ := loadTestBookmarksManager(t, ManagerConfig{
		Enabled:        true,
		ExecTimeoutSec: 5,
		MaxOpsPerExec:  1000,
	})
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("test_bookmarks")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("test_bookmarks not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)
	L.SetContext(ctx)

	// Verify all log levels work without error.
	logCode := `
		log.info("integration test info message")
		log.warn("integration test warn message")
		log.error("integration test error message")
		log.debug("integration test debug message")
		log.info("with context", {key = "value", count = 42})
	`
	if err := L.DoString(logCode); err != nil {
		t.Fatalf("log API test failed: %s", err)
	}
}

// TestIntegration_NestedTransactionRejected verifies that nested db.transaction()
// calls are rejected with an error.
func TestIntegration_NestedTransactionRejected(t *testing.T) {
	conn := newIntegrationDB(t)
	dir := t.TempDir()

	writePluginFile(t, dir, "nested_tx", `
plugin_info = {
    name        = "nested_tx",
    version     = "1.0.0",
    description = "Tests nested transaction rejection",
}

function on_init()
    db.define_table("data", {
        columns = {
            {name = "value", type = "text"},
        },
    })
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}
	defer mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("nested_tx")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("nested_tx not running: %v", inst)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := inst.Pool.Get(ctx)
	if err != nil {
		t.Fatalf("Pool.Get: %s", err)
	}
	defer inst.Pool.Put(L)

	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}
	L.SetContext(ctx)

	nestedCode := `
		local ok, err = pcall(function()
			db.transaction(function()
				db.insert("data", {value = "outer"})
				db.transaction(function()
					db.insert("data", {value = "inner"})
				end)
			end)
		end)
		if ok then
			error("expected nested transaction to be rejected")
		end
		if not string.find(tostring(err), "nested") then
			error("unexpected error: " .. tostring(err))
		end
	`
	if err := L.DoString(nestedCode); err != nil {
		// The error could also be raised by ArgError which uses L.RaiseError under the hood.
		// If pcall catches it, the Lua code above handles verification.
		// If pcall doesn't catch it, the error propagates here.
		if !strings.Contains(err.Error(), "nested") {
			t.Fatalf("nested transaction test failed: %s", err)
		}
	}
}
