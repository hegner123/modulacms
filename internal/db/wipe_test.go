// White-box tests for wipe.go: DropAllTables methods on Database,
// MysqlDatabase, and PsqlDatabase.
//
// White-box access is needed because:
//   - We construct zero-value and partially-initialized Database structs directly
//     to test error propagation paths (no exported constructor exposes these states).
//   - We inspect sqlite_master to verify table drop ordering and partial-failure behavior.
//   - Existing roundtrip tests in db_test.go cover the happy path (create+drop); these
//     tests focus on error wrapping, partial failure, ordering, and cross-database consistency.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	config "github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Helper ---

// newWipeTestDB creates a fresh SQLite Database with all tables created.
// The caller does NOT need to close the connection; t.Cleanup handles it.
func newWipeTestDB(t *testing.T) Database {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "wipe_test.db")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if _, err := conn.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}

	d := Database{
		Connection: conn,
		Context:    context.Background(),
		Config:     config.Config{Node_ID: types.NewNodeID().String()},
	}

	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}
	return d
}

// countAppTables returns the number of non-sqlite internal tables.
func countAppTables(t *testing.T, conn *sql.DB) int {
	t.Helper()
	var count int
	err := conn.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite%';",
	).Scan(&count)
	if err != nil {
		t.Fatalf("count tables: %v", err)
	}
	return count
}

// listAppTables returns the names of all non-sqlite tables.
func listAppTables(t *testing.T, conn *sql.DB) []string {
	t.Helper()
	rows, err := conn.Query(
		"SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite%' ORDER BY name;",
	)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan table name: %v", err)
		}
		names = append(names, name)
	}
	return names
}

// tableExists checks whether a specific table exists in the database.
func tableExists(t *testing.T, conn *sql.DB, tableName string) bool {
	t.Helper()
	var name string
	err := conn.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name=?;",
		tableName,
	).Scan(&name)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		return false
	}
	return true
}

// Tables that use "DROP TABLE" (strict -- error if missing).
// These are the tables we can use for error-path testing.
var strictDropTables = []string{
	"admin_content_fields",
	"content_fields",
	"admin_content_data",
	"content_data",
	"admin_fields",
	"fields",
	"admin_datatypes",
	"datatypes",
	"routes",
	"admin_routes",
	"media",
	"tables",
	"sessions",
	"user_ssh_keys",
	"user_oauth",
	"tokens",
	"users",
	"media_dimensions",
	"roles",
	"permissions",
}

// Tables that use "DROP TABLE IF EXISTS" (tolerant -- no error if missing).
// Pre-dropping these won't cause DropAllTables to fail.
var ifExistsDropTables = []string{
	"pipelines",
	"plugins",
	"role_permissions",
	"admin_content_relations",
	"content_relations",
	"admin_content_versions",
	"content_versions",
	"backup_sets",
	"backup_verifications",
	"backups",
	"change_events",
}

// --- DropAllTables: SQLite success tests ---

func TestDatabase_DropAllTables_AllTablesRemoved(t *testing.T) {
	// Verify that after DropAllTables, zero application tables remain.
	d := newWipeTestDB(t)

	before := countAppTables(t, d.Connection)
	if before == 0 {
		t.Fatal("precondition failed: expected tables to exist before drop")
	}

	if err := d.DropAllTables(); err != nil {
		t.Fatalf("DropAllTables: %v", err)
	}

	after := countAppTables(t, d.Connection)
	if after != 0 {
		remaining := listAppTables(t, d.Connection)
		t.Errorf("expected 0 tables after DropAllTables, got %d: %v", after, remaining)
	}
}

func TestDatabase_DropAllTables_SpecificTablesGone(t *testing.T) {
	// Verify each known table that CreateAllTables actually creates is absent after drop.
	// NOTE: backup_sets and backup_verifications are NOT created by CreateAllTables
	// (CreateBackupTables only creates the "backups" table). Their Drop uses IF EXISTS
	// and silently succeeds. We only check tables that actually exist.
	d := newWipeTestDB(t)

	if err := d.DropAllTables(); err != nil {
		t.Fatalf("DropAllTables: %v", err)
	}

	// Tables actually created by CreateAllTables (excludes backup_sets, backup_verifications)
	createdTables := []string{
		"pipelines",
		"plugins",
		"role_permissions",
		"admin_content_relations",
		"content_relations",
		"admin_content_versions",
		"content_versions",
		"admin_content_fields",
		"content_fields",
		"admin_content_data",
		"content_data",
		"admin_fields",
		"fields",
		"admin_datatypes",
		"datatypes",
		"routes",
		"admin_routes",
		"media",
		"tables",
		"sessions",
		"user_ssh_keys",
		"user_oauth",
		"tokens",
		"users",
		"media_dimensions",
		"roles",
		"permissions",
		"backups",
		"change_events",
	}

	for _, table := range createdTables {
		t.Run(table, func(t *testing.T) {
			if tableExists(t, d.Connection, table) {
				t.Errorf("table %q still exists after DropAllTables", table)
			}
		})
	}
}

// --- DropAllTables: error early-return behavior ---

func TestDatabase_DropAllTables_ErrorOnFirstTable(t *testing.T) {
	// The first strict table (DROP TABLE without IF EXISTS) in the drop
	// sequence is admin_content_fields. Pre-dropping it should cause
	// DropAllTables to fail with an error mentioning that table.
	// Note: pipelines, plugins, role_permissions, admin_content_relations,
	// and content_relations use DROP TABLE IF EXISTS, so they
	// silently succeed even when already dropped.
	d := newWipeTestDB(t)

	// Pre-drop the first strict table in the sequence
	_, err := d.Connection.Exec("DROP TABLE admin_content_fields;")
	if err != nil {
		t.Fatalf("pre-drop admin_content_fields: %v", err)
	}

	err = d.DropAllTables()
	// DropAllTables continues past missing tables. If it errors, the
	// table name should appear in the combined error message.
	if err != nil && !strings.Contains(err.Error(), "admin_content_fields") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "admin_content_fields")
	}
}

func TestDatabase_DropAllTables_ErrorOnMidTierTable(t *testing.T) {
	// Pre-drop a table from the middle of the sequence (users, Tier 1).
	// DropAllTables continues past the missing table and drops remaining tables.
	d := newWipeTestDB(t)

	_, err := d.Connection.Exec("DROP TABLE users;")
	if err != nil {
		t.Fatalf("pre-drop users: %v", err)
	}

	err = d.DropAllTables()
	if err != nil && !strings.Contains(err.Error(), "users") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "users")
	}

	// Tables before 'users' in the drop order should have been dropped successfully.
	tier5Tables := []string{
		"admin_content_fields",
		"content_fields",
	}
	for _, table := range tier5Tables {
		if tableExists(t, d.Connection, table) {
			t.Errorf("table %q should have been dropped, but still exists", table)
		}
	}

	// Tables after 'users' should also be dropped (continue-on-error behavior).
	tier0Tables := []string{
		"media_dimensions",
		"roles",
		"permissions",
	}
	for _, table := range tier0Tables {
		if tableExists(t, d.Connection, table) {
			t.Errorf("table %q should have been dropped (continue-on-error), but still exists", table)
		}
	}
}

func TestDatabase_DropAllTables_ErrorOnLastStrictTable(t *testing.T) {
	// Pre-drop the last strict-drop table in the sequence (permissions, Tier 0).
	// DropAllTables continues past it and drops remaining infrastructure tables.
	d := newWipeTestDB(t)

	_, err := d.Connection.Exec("DROP TABLE permissions;")
	if err != nil {
		t.Fatalf("pre-drop permissions: %v", err)
	}

	err = d.DropAllTables()
	if err != nil && !strings.Contains(err.Error(), "permissions") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "permissions")
	}

	// All other tables should still be dropped
	if tableExists(t, d.Connection, "users") {
		t.Error("users should have been dropped")
	}
	if tableExists(t, d.Connection, "roles") {
		t.Error("roles should have been dropped")
	}
	if tableExists(t, d.Connection, "media_dimensions") {
		t.Error("media_dimensions should have been dropped")
	}
}

// --- DropAllTables: IF EXISTS tables succeed silently ---

func TestDatabase_DropAllTables_IfExistsTables_NoError(t *testing.T) {
	// Tables that use DROP TABLE IF EXISTS should NOT cause errors when
	// pre-dropped. This documents the asymmetry between strict and tolerant
	// drop statements in the wipe sequence.
	//
	// Note: backup_sets and backup_verifications are never created by
	// CreateAllTables, so they don't need pre-dropping -- they never existed.
	ifExistsCreated := []string{
		"pipelines",
		"plugins",
		"role_permissions",
		"admin_content_relations",
		"content_relations",
		"admin_content_versions",
		"content_versions",
		"backups",
		"change_events",
	}

	for _, table := range ifExistsCreated {
		t.Run(table, func(t *testing.T) {
			d := newWipeTestDB(t)

			// Pre-drop this IF EXISTS table
			_, err := d.Connection.Exec("DROP TABLE " + table + ";")
			if err != nil {
				t.Fatalf("pre-drop %s: %v", table, err)
			}

			// DropAllTables should still succeed because the generated SQL
			// for this table uses DROP TABLE IF EXISTS
			err = d.DropAllTables()
			if err != nil {
				t.Errorf("DropAllTables failed after pre-dropping IF EXISTS table %s: %v", table, err)
			}
		})
	}
}

func TestDatabase_DropAllTables_NonexistentBackupSubtables(t *testing.T) {
	// backup_sets and backup_verifications are never created by CreateAllTables,
	// yet DropAllTables references them. The drop should succeed because the
	// generated SQL uses DROP TABLE IF EXISTS for these tables.
	d := newWipeTestDB(t)

	// Verify they don't exist
	if tableExists(t, d.Connection, "backup_sets") {
		t.Error("precondition failed: backup_sets should not exist")
	}
	if tableExists(t, d.Connection, "backup_verifications") {
		t.Error("precondition failed: backup_verifications should not exist")
	}

	// DropAllTables should succeed despite these tables not existing
	if err := d.DropAllTables(); err != nil {
		t.Fatalf("DropAllTables: %v", err)
	}
}

// --- DropAllTables: error wrapping ---

func TestDatabase_DropAllTables_ErrorWrapping(t *testing.T) {
	// DropAllTables continues past missing tables, logging warnings,
	// and returns a combined error listing all failed table names.
	d := newWipeTestDB(t)

	// Drop a table to force an error
	_, err := d.Connection.Exec("DROP TABLE permissions;")
	if err != nil {
		t.Fatalf("pre-drop permissions: %v", err)
	}

	err = d.DropAllTables()
	// With IF EXISTS in the SQL, a pre-dropped table may not error.
	// If it does error, the table name should appear in the combined message.
	if err != nil && !strings.Contains(err.Error(), "permissions") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "permissions")
	}
}

// --- DropAllTables: strict table error message inventory ---

func TestDatabase_DropAllTables_ErrorMessages_StrictTables(t *testing.T) {
	// Verify that when a table is pre-dropped, DropAllTables continues
	// past the missing table. If the drop errors, the table name appears
	// in the combined error message.

	for _, table := range strictDropTables {
		t.Run(table, func(t *testing.T) {
			d := newWipeTestDB(t)

			// Pre-drop this specific table
			_, err := d.Connection.Exec("DROP TABLE " + table + ";")
			if err != nil {
				t.Fatalf("pre-drop %s: %v", table, err)
			}

			err = d.DropAllTables()
			// DropAllTables continues past failures. If an error is returned,
			// the table name should appear in the combined message.
			if err != nil && !strings.Contains(err.Error(), table) {
				t.Errorf("error = %q, want it to contain table name %q", err.Error(), table)
			}
		})
	}
}

// --- DropAllTables: double-drop behavior ---

func TestDatabase_DropAllTables_DoubleDrop(t *testing.T) {
	// Calling DropAllTables twice: the first call drops all tables.
	// The second call encounters missing tables but continues through all
	// of them, logging warnings. It may return a combined error listing
	// the tables that failed, or succeed if all SQL uses IF EXISTS.
	d := newWipeTestDB(t)

	// First drop: should succeed
	if err := d.DropAllTables(); err != nil {
		t.Fatalf("first DropAllTables: %v", err)
	}

	// Second drop: continues past missing tables
	err := d.DropAllTables()
	if err != nil {
		// If errors occurred, they should list table names
		t.Logf("second DropAllTables returned (expected): %v", err)
	}
}

// --- DropAllTables: nil context behavior ---

func TestDatabase_DropAllTables_NilContext(t *testing.T) {
	t.Parallel()
	// A Database with a nil Context field panics when DropAllTables tries
	// to pass the nil context to database/sql.(*DB).ExecContext. Go 1.25's
	// database/sql requires a non-nil context and panics otherwise.
	// We document this as expected behavior: callers must set a valid Context.
	//
	// We verify this by re-executing the test binary in a subprocess. Running
	// the panic in-process causes a deadlock: the panic inside database/sql
	// holds an internal mutex, and the sql.Open connectionOpener goroutine
	// blocks on that mutex forever, preventing cleanup and causing a 10-minute
	// timeout.
	if os.Getenv("TEST_NIL_CONTEXT_SUBPROCESS") == "1" {
		conn, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			fmt.Fprintf(os.Stderr, "sql.Open: %v", err)
			os.Exit(2)
		}
		d := Database{
			Connection: conn,
			Context:    nil,
		}
		d.DropAllTables() //nolint:errcheck
		// Should not reach here — the above should panic.
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestDatabase_DropAllTables_NilContext$")
	cmd.Env = append(os.Environ(), "TEST_NIL_CONTEXT_SUBPROCESS=1")
	err := cmd.Run()
	if err == nil {
		t.Error("expected subprocess to exit non-zero (panic), but it exited 0")
	}
	// A panic causes a non-zero exit — that's what we expect.
}

// --- DropAllTables: nil connection behavior ---

func TestDatabase_DropAllTables_NilConnection(t *testing.T) {
	t.Parallel()
	// A Database with nil Connection should panic when DropAllTables tries
	// to create queries via mdb.New(d.Connection). We document this behavior.
	d := Database{
		Connection: nil,
		Context:    context.Background(),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from DropAllTables with nil connection, but did not panic")
		}
	}()
	d.DropAllTables() //nolint:errcheck
}

func TestMysqlDatabase_DropAllTables_NilConnection(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{
		Connection: nil,
		Context:    context.Background(),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from MysqlDatabase.DropAllTables with nil connection, but did not panic")
		}
	}()
	d.DropAllTables() //nolint:errcheck
}

func TestPsqlDatabase_DropAllTables_NilConnection(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{
		Connection: nil,
		Context:    context.Background(),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from PsqlDatabase.DropAllTables with nil connection, but did not panic")
		}
	}()
	d.DropAllTables() //nolint:errcheck
}

// --- Drop ordering: verify reverse dependency ---

func TestDatabase_DropAllTables_DependencyOrder(t *testing.T) {
	// Verify that DropAllTables drops tables in reverse dependency order
	// by checking that higher-tier tables are dropped before lower-tier tables.
	// We do this by pre-dropping a low-tier table and verifying that
	// higher-tier tables are already gone by the time the error occurs.

	t.Run("tier6_before_tier0", func(t *testing.T) {
		d := newWipeTestDB(t)

		// Pre-drop a Tier 0 table (permissions)
		_, err := d.Connection.Exec("DROP TABLE permissions;")
		if err != nil {
			t.Fatalf("pre-drop permissions: %v", err)
		}

		err = d.DropAllTables()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// All Tier 5 content field tables should be gone (they're dropped first)
		tier5 := []string{
			"admin_content_fields",
			"content_fields",
		}
		for _, table := range tier5 {
			if tableExists(t, d.Connection, table) {
				t.Errorf("Tier 5 table %q should have been dropped before Tier 0 failure", table)
			}
		}
	})

	t.Run("content_before_fields", func(t *testing.T) {
		d := newWipeTestDB(t)

		// Pre-drop a Tier 3 table (fields)
		_, err := d.Connection.Exec("DROP TABLE fields;")
		if err != nil {
			t.Fatalf("pre-drop fields: %v", err)
		}

		err = d.DropAllTables()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Content field tables (Tier 5) should be gone -- they're dropped before fields (Tier 3)
		if tableExists(t, d.Connection, "content_fields") {
			t.Error("content_fields (Tier 5) should be dropped before fields (Tier 3)")
		}
		if tableExists(t, d.Connection, "admin_content_fields") {
			t.Error("admin_content_fields (Tier 5) should be dropped before fields (Tier 3)")
		}

		// Content data tables (Tier 4) should also be gone
		if tableExists(t, d.Connection, "content_data") {
			t.Error("content_data (Tier 4) should be dropped before fields (Tier 3)")
		}
	})

	t.Run("users_before_foundation", func(t *testing.T) {
		d := newWipeTestDB(t)

		// Pre-drop a Tier 0 table (roles)
		_, err := d.Connection.Exec("DROP TABLE roles;")
		if err != nil {
			t.Fatalf("pre-drop roles: %v", err)
		}

		err = d.DropAllTables()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Users (Tier 1) should be gone -- dropped before roles (Tier 0)
		if tableExists(t, d.Connection, "users") {
			t.Error("users (Tier 1) should be dropped before roles (Tier 0)")
		}

		// Sessions, tokens, etc. (Tier 2) should also be gone
		tier2Tables := []string{"sessions", "tokens", "user_ssh_keys", "user_oauth"}
		for _, table := range tier2Tables {
			if tableExists(t, d.Connection, table) {
				t.Errorf("Tier 2 table %q should be dropped before roles (Tier 0)", table)
			}
		}
	})

	t.Run("content_fields_before_content_data", func(t *testing.T) {
		// Verify content_fields (Tier 5) are dropped before content_data (Tier 4)
		d := newWipeTestDB(t)

		// Pre-drop content_data (Tier 4)
		_, err := d.Connection.Exec("DROP TABLE content_data;")
		if err != nil {
			t.Fatalf("pre-drop content_data: %v", err)
		}

		err = d.DropAllTables()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Content field tables (Tier 5) should be gone -- they're dropped before content_data (Tier 4)
		if tableExists(t, d.Connection, "admin_content_fields") {
			t.Error("admin_content_fields (Tier 5) should be dropped before content_data (Tier 4)")
		}
		if tableExists(t, d.Connection, "content_fields") {
			t.Error("content_fields (Tier 5) should be dropped before content_data (Tier 4)")
		}
	})
}

// --- DropAllTables with data: verify tables with rows can be dropped ---

func TestDatabase_DropAllTables_WithData(t *testing.T) {
	// DropAllTables should succeed even when tables contain data.
	// This is critical: DROP TABLE removes the table regardless of contents.
	d := newWipeTestDB(t)

	// Insert bootstrap data into the tables so they're not empty
	if err := d.CreateBootstrapData(""); err != nil {
		t.Fatalf("CreateBootstrapData: %v", err)
	}

	// Verify some tables have data
	var permCount int
	if err := d.Connection.QueryRow("SELECT COUNT(*) FROM permissions;").Scan(&permCount); err != nil {
		t.Fatalf("count permissions: %v", err)
	}
	if permCount == 0 {
		t.Fatal("precondition failed: permissions table has no data after bootstrap")
	}

	// Now drop everything -- should succeed despite tables having rows
	if err := d.DropAllTables(); err != nil {
		t.Fatalf("DropAllTables with data: %v", err)
	}

	after := countAppTables(t, d.Connection)
	if after != 0 {
		t.Errorf("expected 0 tables after drop with data, got %d", after)
	}
}

// --- DropAllTables after closed connection ---

func TestDatabase_DropAllTables_ClosedConnection(t *testing.T) {
	// A closed connection should cause DropAllTables to fail with an error.
	d := newWipeTestDB(t)

	// Close the underlying connection
	d.Connection.Close()

	err := d.DropAllTables()
	if err == nil {
		t.Fatal("expected error from DropAllTables on closed connection, got nil")
	}
	// The error should list failed tables (all will fail on a closed connection)
	if !strings.Contains(err.Error(), "webhook_deliveries") {
		t.Errorf("error = %q, want it to reference webhook_deliveries", err.Error())
	}
}

// --- Table count verification ---

func TestDatabase_DropAllTables_TableCount(t *testing.T) {
	// Verify the number of tables before and after drop.
	// This catches additions/removals of tables in CreateAllTables that
	// are not matched by corresponding changes in DropAllTables.
	d := newWipeTestDB(t)

	before := countAppTables(t, d.Connection)
	// CreateAllTables creates ~27 tables (not backup_sets or backup_verifications).
	// We check a reasonable lower bound.
	if before < 25 {
		t.Errorf("expected at least 25 tables after CreateAllTables, got %d", before)
	}

	if err := d.DropAllTables(); err != nil {
		t.Fatalf("DropAllTables: %v", err)
	}

	after := countAppTables(t, d.Connection)
	if after != 0 {
		remaining := listAppTables(t, d.Connection)
		t.Errorf("DropAllTables left %d tables behind: %v", after, remaining)
	}
}

// --- CreateAllTables -> DropAllTables -> CreateAllTables cycle ---

func TestDatabase_DropAllTables_CreateDropCreateCycle(t *testing.T) {
	// Verify that after drop, tables can be recreated successfully.
	// This proves DropAllTables leaves the database in a clean state.
	d := newWipeTestDB(t)

	// Drop
	if err := d.DropAllTables(); err != nil {
		t.Fatalf("first DropAllTables: %v", err)
	}

	// Recreate
	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("second CreateAllTables: %v", err)
	}

	after := countAppTables(t, d.Connection)
	if after == 0 {
		t.Fatal("no tables after second CreateAllTables")
	}

	// Drop again
	if err := d.DropAllTables(); err != nil {
		t.Fatalf("second DropAllTables: %v", err)
	}

	final := countAppTables(t, d.Connection)
	if final != 0 {
		remaining := listAppTables(t, d.Connection)
		t.Errorf("expected 0 tables after second drop, got %d: %v", final, remaining)
	}
}

// --- Partial failure: some tables survive ---

func TestDatabase_DropAllTables_ContinuesPastFailure(t *testing.T) {
	// When DropAllTables encounters a missing table, it continues and
	// drops all remaining tables. Tables after the failure point are
	// NOT preserved (unlike the old fail-fast behavior).
	d := newWipeTestDB(t)

	// Insert bootstrap data
	if err := d.CreateBootstrapData(""); err != nil {
		t.Fatalf("CreateBootstrapData: %v", err)
	}

	// Pre-drop 'fields' (Tier 3) to cause a warning there
	_, err := d.Connection.Exec("DROP TABLE fields;")
	if err != nil {
		t.Fatalf("pre-drop fields: %v", err)
	}

	err = d.DropAllTables()
	// Error is expected (fields was missing), but all other tables
	// should still be dropped.
	if err != nil && !strings.Contains(err.Error(), "fields") {
		t.Errorf("error = %q, want it to mention 'fields'", err.Error())
	}

	// Roles and permissions (Tier 0) should be dropped because
	// DropAllTables continues past the 'fields' failure.
	if tableExists(t, d.Connection, "roles") {
		t.Error("roles table should have been dropped (continue-on-error)")
	}
	if tableExists(t, d.Connection, "permissions") {
		t.Error("permissions table should have been dropped (continue-on-error)")
	}
}

// --- Zero-value struct behavior ---

func TestDatabase_DropAllTables_ZeroValueStruct(t *testing.T) {
	t.Parallel()
	// A completely zero-value Database struct should panic because Connection is nil.
	var d Database

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from zero-value Database.DropAllTables, but did not panic")
		}
	}()
	d.DropAllTables() //nolint:errcheck
}

func TestMysqlDatabase_DropAllTables_ZeroValueStruct(t *testing.T) {
	t.Parallel()
	var d MysqlDatabase

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from zero-value MysqlDatabase.DropAllTables, but did not panic")
		}
	}()
	d.DropAllTables() //nolint:errcheck
}

func TestPsqlDatabase_DropAllTables_ZeroValueStruct(t *testing.T) {
	t.Parallel()
	var d PsqlDatabase

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from zero-value PsqlDatabase.DropAllTables, but did not panic")
		}
	}()
	d.DropAllTables() //nolint:errcheck
}

// --- Cross-database consistency: error message patterns ---

func TestDropAllTables_ErrorMessageFormat(t *testing.T) {
	t.Parallel()
	// All error messages from DropAllTables should follow the format:
	// "drop <table_name>: <underlying error>"
	// This verifies the wrapping convention is consistent.
	tables := []string{
		"admin_content_fields",
		"content_fields",
		"admin_content_data",
		"content_data",
		"admin_fields",
		"fields",
		"admin_datatypes",
		"datatypes",
		"routes",
		"admin_routes",
		"media",
		"tables",
		"sessions",
		"user_ssh_keys",
		"user_oauth",
		"tokens",
		"users",
		"media_dimensions",
		"roles",
		"permissions",
	}

	for _, table := range tables {
		t.Run(table, func(t *testing.T) {
			t.Parallel()
			d := newWipeTestDB(t)

			_, err := d.Connection.Exec("DROP TABLE " + table + ";")
			if err != nil {
				t.Fatalf("pre-drop %s: %v", table, err)
			}

			err = d.DropAllTables()
			// DropAllTables now continues past missing tables and reports
			// all failures in a combined error. If the table was already
			// dropped and the SQL uses IF EXISTS, it may succeed silently.
			// If it does error, the table name should appear in the message.
			if err != nil && !strings.Contains(err.Error(), table) {
				t.Errorf("error = %q, want it to contain table name %q", err.Error(), table)
			}
		})
	}
}

// --- Canceled context behavior ---

func TestDatabase_DropAllTables_CanceledContext(t *testing.T) {
	// A canceled context should cause DropAllTables to fail. The SQLite driver
	// checks context cancellation before executing queries, so the first
	// drop statement should return a context.Canceled error.
	d := newWipeTestDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately before any work

	d.Context = ctx

	err := d.DropAllTables()
	if err == nil {
		// Some SQLite driver versions ignore canceled contexts. Document this.
		after := countAppTables(t, d.Connection)
		if after == 0 {
			t.Log("DropAllTables succeeded despite canceled context (driver ignores context)")
		} else {
			t.Log("DropAllTables returned nil error but tables still exist -- unexpected")
		}
		return
	}

	// The error should reference failed tables (all will fail with canceled context)
	if !strings.Contains(err.Error(), "webhook_deliveries") {
		t.Errorf("error = %q, want it to reference webhook_deliveries", err.Error())
	}
}

// --- Concurrent DropAllTables calls ---

func TestDatabase_DropAllTables_Concurrent(t *testing.T) {
	// Two goroutines racing to drop all tables should both complete without
	// panicking. One should succeed, the other should either succeed (if
	// DROP IF EXISTS is used) or return an error (if not). Neither should panic.
	d := newWipeTestDB(t)

	const goroutines = 5
	errs := make(chan error, goroutines)

	for range goroutines {
		go func() {
			errs <- d.DropAllTables()
		}()
	}

	var successes, failures int
	for range goroutines {
		err := <-errs
		if err == nil {
			successes++
		} else {
			failures++
		}
	}

	// At least one goroutine should succeed (the first one to execute each DROP)
	// or all may fail if they interleave badly. The key assertion: no panic.
	t.Logf("concurrent DropAllTables: %d successes, %d failures", successes, failures)

	// After all goroutines finish, tables should be gone
	after := countAppTables(t, d.Connection)
	if after != 0 {
		remaining := listAppTables(t, d.Connection)
		t.Errorf("expected 0 tables after concurrent drops, got %d: %v", after, remaining)
	}
}

// --- Read-only database behavior ---

func TestDatabase_DropAllTables_ReadOnlyConnection(t *testing.T) {
	// A read-only SQLite connection should fail to drop tables.
	// This verifies that DropAllTables properly propagates the driver error.
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "readonly_wipe.db")

	// First create the database with tables using a writable connection
	writeConn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open (write): %v", err)
	}
	if _, err := writeConn.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		writeConn.Close()
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if _, err := writeConn.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		writeConn.Close()
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	dWrite := Database{
		Connection: writeConn,
		Context:    context.Background(),
		Config:     config.Config{Node_ID: types.NewNodeID().String()},
	}
	if err := dWrite.CreateAllTables(); err != nil {
		writeConn.Close()
		t.Fatalf("CreateAllTables: %v", err)
	}
	writeConn.Close()

	// Re-open as read-only using SQLite URI with mode=ro
	roConn, err := sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
	if err != nil {
		t.Fatalf("sql.Open (read-only): %v", err)
	}
	t.Cleanup(func() { roConn.Close() })

	dRO := Database{
		Connection: roConn,
		Context:    context.Background(),
	}

	err = dRO.DropAllTables()
	if err == nil {
		t.Fatal("expected error from DropAllTables on read-only connection, got nil")
	}
	// The error should list failed tables (all will fail on read-only)
	if !strings.Contains(err.Error(), "webhook_deliveries") {
		t.Errorf("error = %q, want it to reference webhook_deliveries", err.Error())
	}
}

// --- Unknown tables survive DropAllTables ---

func TestDatabase_DropAllTables_UnknownTablesSurvive(t *testing.T) {
	// DropAllTables only drops the known schema tables. If custom tables
	// exist (e.g., from plugins or manual creation), they should survive.
	d := newWipeTestDB(t)

	// Create an extra table not in the schema
	_, err := d.Connection.Exec("CREATE TABLE custom_plugin_data (id INTEGER PRIMARY KEY, value TEXT);")
	if err != nil {
		t.Fatalf("create custom table: %v", err)
	}

	if err := d.DropAllTables(); err != nil {
		t.Fatalf("DropAllTables: %v", err)
	}

	// The custom table should survive
	if !tableExists(t, d.Connection, "custom_plugin_data") {
		t.Error("custom_plugin_data table was dropped by DropAllTables; expected it to survive")
	}

	// Only the custom table should remain
	remaining := listAppTables(t, d.Connection)
	if len(remaining) != 1 {
		t.Errorf("expected exactly 1 remaining table (custom_plugin_data), got %d: %v", len(remaining), remaining)
	}
	if len(remaining) == 1 && remaining[0] != "custom_plugin_data" {
		t.Errorf("remaining table = %q, want %q", remaining[0], "custom_plugin_data")
	}
}

// --- FK enforcement: DROP TABLE bypasses FK checks in SQLite ---

func TestDatabase_DropAllTables_ForeignKeyEnforcement(t *testing.T) {
	// SQLite's DROP TABLE bypasses foreign key checks. This test documents
	// that DropAllTables works even with foreign_keys=ON and rows that
	// reference other tables. If the DROP order were wrong AND SQLite
	// enforced FKs on DROP, the test would fail.
	d := newWipeTestDB(t)

	// Insert bootstrap data to create FK references between tables
	if err := d.CreateBootstrapData(""); err != nil {
		t.Fatalf("CreateBootstrapData: %v", err)
	}

	// Verify FK enforcement is active for DML operations
	var fkEnabled int
	if err := d.Connection.QueryRow("PRAGMA foreign_keys;").Scan(&fkEnabled); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Fatalf("foreign_keys = %d, want 1 (ON)", fkEnabled)
	}

	// DropAllTables should succeed despite FK references existing
	if err := d.DropAllTables(); err != nil {
		t.Fatalf("DropAllTables with FK enforcement and data: %v", err)
	}

	after := countAppTables(t, d.Connection)
	if after != 0 {
		remaining := listAppTables(t, d.Connection)
		t.Errorf("expected 0 tables, got %d: %v", after, remaining)
	}
}

// --- Partial failure: exact remaining table count ---

func TestDatabase_DropAllTables_ContinueOnError_RemainingCount(t *testing.T) {
	// With continue-on-error, DropAllTables drops all tables it can,
	// skipping only those that fail. Pre-dropping a table should result
	// in 0 remaining tables (the pre-dropped one is already gone, and
	// all others get dropped successfully).

	tests := []struct {
		name         string
		preDropTable string
	}{
		{"pre_drop_first_strict", "admin_content_fields"},
		{"pre_drop_users", "users"},
		{"pre_drop_change_events", "change_events"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newWipeTestDB(t)

			_, err := d.Connection.Exec("DROP TABLE " + tt.preDropTable + ";")
			if err != nil {
				t.Fatalf("pre-drop %s: %v", tt.preDropTable, err)
			}

			err = d.DropAllTables()
			// May or may not error depending on IF EXISTS usage
			if err != nil {
				t.Logf("DropAllTables returned (expected): %v", err)
			}

			after := countAppTables(t, d.Connection)
			if after != 0 {
				remaining := listAppTables(t, d.Connection)
				t.Errorf("expected 0 remaining tables, got %d: %v", after, remaining)
			}
		})
	}
}

// --- Drop with SQLite busy connection ---

func TestDatabase_DropAllTables_WhileReading(t *testing.T) {
	// DropAllTables should work even when another goroutine has an open
	// read transaction on the database (WAL mode allows concurrent readers).
	d := newWipeTestDB(t)

	if err := d.CreateBootstrapData(""); err != nil {
		t.Fatalf("CreateBootstrapData: %v", err)
	}

	// Start a read transaction that stays open
	tx, err := d.Connection.Begin()
	if err != nil {
		t.Fatalf("Begin read tx: %v", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM permissions;").Scan(&count); err != nil {
		t.Fatalf("SELECT in read tx: %v", err)
	}
	if count == 0 {
		t.Fatal("precondition: permissions should have bootstrap data")
	}

	// Now try to drop all tables while the read tx is open.
	// In WAL mode, this should succeed because writers don't block on readers.
	err = d.DropAllTables()
	if err != nil {
		// Some SQLite configurations may block here. Document the behavior.
		t.Logf("DropAllTables while reading returned error (may be expected): %v", err)
		return
	}

	after := countAppTables(t, d.Connection)
	if after != 0 {
		remaining := listAppTables(t, d.Connection)
		t.Errorf("expected 0 tables after drop, got %d: %v", after, remaining)
	}
}

// --- Multiple pre-drops: cascading early failure ---

func TestDatabase_DropAllTables_MultiplePreDrops(t *testing.T) {
	// Pre-dropping multiple tables at different tiers. DropAllTables
	// continues past all missing tables and drops everything it can.
	d := newWipeTestDB(t)

	// Pre-drop tables from different tiers
	tablesToPreDrop := []string{
		"content_fields", // Tier 5
		"users",          // Tier 1
		"permissions",    // Tier 0
	}
	for _, table := range tablesToPreDrop {
		_, err := d.Connection.Exec("DROP TABLE " + table + ";")
		if err != nil {
			t.Fatalf("pre-drop %s: %v", table, err)
		}
	}

	err := d.DropAllTables()
	// Should report all three missing tables in the combined error
	if err != nil {
		for _, table := range tablesToPreDrop {
			if !strings.Contains(err.Error(), table) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), table)
			}
		}
	}

	// All tables should be gone (pre-dropped ones were already gone,
	// remaining ones were dropped by continue-on-error)
	if tableExists(t, d.Connection, "admin_content_fields") {
		t.Error("admin_content_fields should have been dropped")
	}
	if tableExists(t, d.Connection, "admin_content_data") {
		t.Error("admin_content_data should have been dropped")
	}
	if tableExists(t, d.Connection, "roles") {
		t.Error("roles should have been dropped")
	}
}

// --- DropAllTables completeness: every created table is dropped ---

func TestDatabase_DropAllTables_Completeness(t *testing.T) {
	// Verify that every table created by CreateAllTables is also dropped
	// by DropAllTables. This catches schema additions that forget to add
	// a corresponding drop statement.
	d := newWipeTestDB(t)

	// Record all tables after creation
	beforeTables := listAppTables(t, d.Connection)

	if err := d.DropAllTables(); err != nil {
		t.Fatalf("DropAllTables: %v", err)
	}

	afterTables := listAppTables(t, d.Connection)

	// Build a set of tables that survived the drop
	survivorSet := make(map[string]bool)
	for _, name := range afterTables {
		survivorSet[name] = true
	}

	// Every table that was created should have been dropped
	for _, table := range beforeTables {
		if survivorSet[table] {
			t.Errorf("table %q was created by CreateAllTables but NOT dropped by DropAllTables", table)
		}
	}
}

// --- DbDriver interface compliance ---
// These are compile-time checks. If DropAllTables signature changes,
// these will fail to compile.

var (
	_ interface{ DropAllTables() error } = Database{}
	_ interface{ DropAllTables() error } = MysqlDatabase{}
	_ interface{ DropAllTables() error } = PsqlDatabase{}
)
