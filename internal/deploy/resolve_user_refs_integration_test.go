package deploy

import (
	"context"
	"database/sql"
	"testing"

	"github.com/hegner123/modulacms/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

// ---------------------------------------------------------------------------
// resolveUserRefs cross-backend integration test (SQLite only -- no Docker)
//
// This test exercises the real sqliteDeployOps (from db.NewDeployOps) against
// an in-memory SQLite database, verifying that the Placeholder(1) method
// produces correct SQL ("?") and that resolveUserRefs correctly:
//   - skips existing users
//   - creates placeholder users with [sync] prefix, locked hash, and viewer role
//   - handles mixed existing + missing user sets
//   - handles empty UserRefs gracefully
//
// The PostgreSQL ($1 placeholder) and MySQL (?) backends are tested in
// resolve_user_refs_docker_integration_test.go behind the "integration" build tag.
// ---------------------------------------------------------------------------

// setupSQLiteDeployDB creates a minimal in-memory SQLite database with
// roles and users tables, inserts a viewer role, and returns the pool,
// the real DeployOps from db.NewDeployOps, and the viewer role ID.
func setupSQLiteDeployDB(t *testing.T) (*sql.DB, db.DeployOps, string) {
	t.Helper()

	pool, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	// Enable foreign keys for realistic constraint behavior.
	if _, err := pool.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}

	// Create minimal roles table (users.role references roles.role_id).
	if _, err := pool.Exec(`CREATE TABLE roles (
		role_id TEXT PRIMARY KEY NOT NULL,
		label TEXT NOT NULL UNIQUE,
		system_protected INTEGER NOT NULL DEFAULT 0
	);`); err != nil {
		t.Fatalf("create roles: %v", err)
	}

	// Create users table matching the real schema.
	if _, err := pool.Exec(`CREATE TABLE users (
		user_id TEXT PRIMARY KEY NOT NULL,
		username TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		hash TEXT NOT NULL,
		role TEXT NOT NULL REFERENCES roles(role_id),
		date_created TEXT DEFAULT CURRENT_TIMESTAMP,
		date_modified TEXT DEFAULT CURRENT_TIMESTAMP
	);`); err != nil {
		t.Fatalf("create users: %v", err)
	}

	// Seed viewer role.
	viewerRoleID := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	if _, err := pool.Exec(
		"INSERT INTO roles (role_id, label, system_protected) VALUES (?, ?, ?);",
		viewerRoleID, "viewer", 1,
	); err != nil {
		t.Fatalf("seed viewer role: %v", err)
	}

	// Build DeployOps using the real constructor -- this gives us
	// the actual sqliteDeployOps with the real Placeholder and BulkInsert.
	driver := db.Database{Connection: pool}
	ops, err := db.NewDeployOps(driver)
	if err != nil {
		t.Fatalf("NewDeployOps: %v", err)
	}

	return pool, ops, viewerRoleID
}

// seedExistingUser inserts a user directly into the database.
func seedExistingUser(t *testing.T, pool *sql.DB, userID, username, roleID string) {
	t.Helper()
	if _, err := pool.Exec(
		"INSERT INTO users (user_id, username, name, email, hash, role, date_created, date_modified) VALUES (?, ?, ?, ?, ?, ?, ?, ?);",
		userID, username, "Real "+username, username+"@example.com", "bcrypt-hash", roleID,
		"2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z",
	); err != nil {
		t.Fatalf("seed user %s: %v", username, err)
	}
}

// queryUser retrieves a user row by user_id. Returns nil if not found.
func queryUser(t *testing.T, pool *sql.DB, userID string) *userRow {
	t.Helper()
	var u userRow
	err := pool.QueryRow(
		"SELECT user_id, username, name, email, hash, role FROM users WHERE user_id = ?;",
		userID,
	).Scan(&u.UserID, &u.Username, &u.Name, &u.Email, &u.Hash, &u.Role)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		t.Fatalf("queryUser %s: %v", userID, err)
	}
	return &u
}

type userRow struct {
	UserID   string
	Username string
	Name     string
	Email    string
	Hash     string
	Role     string
}

// countUsers returns the total number of rows in the users table.
func countUsers(t *testing.T, pool *sql.DB) int64 {
	t.Helper()
	var count int64
	if err := pool.QueryRow("SELECT COUNT(*) FROM users;").Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	return count
}

// ---------------------------------------------------------------------------
// SQLite backend -- runs without Docker, no build tag needed
// ---------------------------------------------------------------------------

func TestIntegration_ResolveUserRefs_SQLite(t *testing.T) {
	pool, ops, viewerRoleID := setupSQLiteDeployDB(t)

	// Scenario: a blog sync payload references two users.
	// "alice" exists in the target DB. "bob" does not.
	seedExistingUser(t, pool, "01ALICE0000000000000000000", "alice", viewerRoleID)

	userRefs := map[string]string{
		"01ALICE0000000000000000000": "alice",
		"01BOB00000000000000000000N": "bob",
	}

	// Run resolveUserRefs inside ImportAtomic to match real usage.
	err := ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
		return resolveUserRefs(ctx, ex, ops, userRefs, viewerRoleID)
	})
	if err != nil {
		t.Fatalf("resolveUserRefs: %v", err)
	}

	t.Run("existing user is not duplicated", func(t *testing.T) {
		if c := countUsers(t, pool); c != 2 {
			t.Fatalf("expected 2 users (1 existing + 1 placeholder), got %d", c)
		}
		alice := queryUser(t, pool, "01ALICE0000000000000000000")
		if alice == nil {
			t.Fatal("alice should exist")
		}
		// Original user properties should be unchanged.
		if alice.Username != "alice" {
			t.Errorf("alice username = %q, want %q", alice.Username, "alice")
		}
		if alice.Hash != "bcrypt-hash" {
			t.Errorf("alice hash should be unchanged, got %q", alice.Hash)
		}
	})

	t.Run("missing user gets placeholder account", func(t *testing.T) {
		bob := queryUser(t, pool, "01BOB00000000000000000000N")
		if bob == nil {
			t.Fatal("bob placeholder should have been created")
		}
		if bob.Username != "[sync] bob" {
			t.Errorf("username = %q, want %q", bob.Username, "[sync] bob")
		}
		if bob.Email != "bob@sync.placeholder" {
			t.Errorf("email = %q, want %q", bob.Email, "bob@sync.placeholder")
		}
		if bob.Hash != "" {
			t.Errorf("hash should be empty (locked account), got %q", bob.Hash)
		}
		if bob.Role != viewerRoleID {
			t.Errorf("role = %q, want viewer role %q", bob.Role, viewerRoleID)
		}
		if bob.Name != "" {
			t.Errorf("name = %q, want empty string", bob.Name)
		}
	})

	t.Run("empty user refs is a no-op", func(t *testing.T) {
		beforeCount := countUsers(t, pool)
		err := ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
			return resolveUserRefs(ctx, ex, ops, map[string]string{}, viewerRoleID)
		})
		if err != nil {
			t.Fatalf("resolveUserRefs with empty refs: %v", err)
		}
		if after := countUsers(t, pool); after != beforeCount {
			t.Errorf("user count changed: %d -> %d", beforeCount, after)
		}
	})
}

// TestIntegration_ResolveUserRefs_SQLite_AllNew verifies the path where
// every user in the refs is missing -- all should be created as placeholders.
func TestIntegration_ResolveUserRefs_SQLite_AllNew(t *testing.T) {
	pool, ops, viewerRoleID := setupSQLiteDeployDB(t)

	userRefs := map[string]string{
		"01XUSER10000000000000000000": "user1",
		"01XUSER20000000000000000000": "user2",
		"01XUSER30000000000000000000": "user3",
	}

	err := ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
		return resolveUserRefs(ctx, ex, ops, userRefs, viewerRoleID)
	})
	if err != nil {
		t.Fatalf("resolveUserRefs: %v", err)
	}

	if c := countUsers(t, pool); c != 3 {
		t.Fatalf("expected 3 placeholder users, got %d", c)
	}

	for uid, uname := range userRefs {
		u := queryUser(t, pool, uid)
		if u == nil {
			t.Errorf("user %s (%s) not created", uid, uname)
			continue
		}
		if u.Username != "[sync] "+uname {
			t.Errorf("user %s: username = %q, want %q", uid, u.Username, "[sync] "+uname)
		}
	}
}

// TestIntegration_ResolveUserRefs_SQLite_AllExisting verifies no rows are
// inserted when every referenced user already exists.
func TestIntegration_ResolveUserRefs_SQLite_AllExisting(t *testing.T) {
	pool, ops, viewerRoleID := setupSQLiteDeployDB(t)

	seedExistingUser(t, pool, "01EXISTING1000000000000000A", "user_a", viewerRoleID)
	seedExistingUser(t, pool, "01EXISTING2000000000000000B", "user_b", viewerRoleID)

	userRefs := map[string]string{
		"01EXISTING1000000000000000A": "user_a",
		"01EXISTING2000000000000000B": "user_b",
	}

	err := ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
		return resolveUserRefs(ctx, ex, ops, userRefs, viewerRoleID)
	})
	if err != nil {
		t.Fatalf("resolveUserRefs: %v", err)
	}

	if c := countUsers(t, pool); c != 2 {
		t.Errorf("expected 2 users (no new ones), got %d", c)
	}
}

// TestIntegration_Placeholder_SQLite verifies that the real sqliteDeployOps
// returns "?" for any parameter position, matching SQLite/MySQL convention.
func TestIntegration_Placeholder_SQLite(t *testing.T) {
	t.Parallel()

	_, ops, _ := setupSQLiteDeployDB(t)

	for _, n := range []int{1, 2, 5, 100} {
		got := ops.Placeholder(n)
		if got != "?" {
			t.Errorf("Placeholder(%d) = %q, want %q", n, got, "?")
		}
	}
}
