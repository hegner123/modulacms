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
// an in-memory SQLite database, verifying that resolveUserRefs correctly:
//   - skips existing users (no remap needed)
//   - remaps missing users to the first admin user
//   - handles empty UserRefs gracefully
//   - never creates placeholder users
// ---------------------------------------------------------------------------

// setupSQLiteDeployDB creates a minimal in-memory SQLite database with
// roles, users, and content tables, inserts an admin role and admin user,
// and returns the pool, the real DeployOps, and the admin user ID.
func setupSQLiteDeployDB(t *testing.T) (*sql.DB, db.DeployOps, string) {
	t.Helper()

	pool, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	if _, err := pool.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}

	for _, ddl := range []string{
		`CREATE TABLE roles (role_id TEXT PRIMARY KEY NOT NULL, label TEXT NOT NULL UNIQUE, system_protected INTEGER NOT NULL DEFAULT 0);`,
		`CREATE TABLE users (user_id TEXT PRIMARY KEY NOT NULL, username TEXT NOT NULL UNIQUE, name TEXT NOT NULL, email TEXT NOT NULL, hash TEXT NOT NULL, role TEXT NOT NULL REFERENCES roles(role_id), date_created TEXT DEFAULT CURRENT_TIMESTAMP, date_modified TEXT DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE content_data (content_data_id TEXT PRIMARY KEY, author_id TEXT, published_by TEXT);`,
		`CREATE TABLE content_fields (content_field_id TEXT PRIMARY KEY, author_id TEXT);`,
		`CREATE TABLE admin_content_data (content_data_id TEXT PRIMARY KEY, author_id TEXT, published_by TEXT);`,
		`CREATE TABLE admin_content_fields (content_field_id TEXT PRIMARY KEY, author_id TEXT);`,
	} {
		if _, err := pool.Exec(ddl); err != nil {
			t.Fatalf("ddl: %v", err)
		}
	}

	// Seed admin role and admin user.
	adminRoleID := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	adminUserID := "01ADMIN000000000000000000AA"
	if _, err := pool.Exec(
		"INSERT INTO roles (role_id, label, system_protected) VALUES (?, ?, ?);",
		adminRoleID, "admin", 1,
	); err != nil {
		t.Fatalf("seed admin role: %v", err)
	}
	if _, err := pool.Exec(
		"INSERT INTO users (user_id, username, name, email, hash, role, date_created, date_modified) VALUES (?, ?, ?, ?, ?, ?, ?, ?);",
		adminUserID, "admin", "Admin", "admin@example.com", "bcrypt-hash", adminRoleID,
		"2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z",
	); err != nil {
		t.Fatalf("seed admin user: %v", err)
	}

	driver := db.Database{Connection: pool}
	ops, err := db.NewDeployOps(driver)
	if err != nil {
		t.Fatalf("NewDeployOps: %v", err)
	}

	return pool, ops, adminUserID
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
	pool, ops, adminUserID := setupSQLiteDeployDB(t)

	// Scenario: payload references two users. "alice" exists. "bob" does not.
	// Content has bob as author — should get remapped to admin.
	aliceRoleID := "01ARZ3NDEKTSV4RRFFQ69G5FAV" // same admin role for simplicity
	seedExistingUser(t, pool, "01ALICE0000000000000000000", "alice", aliceRoleID)

	// Insert content referencing the missing user.
	bobID := "01BOB00000000000000000000N"
	if _, err := pool.Exec(
		"INSERT INTO content_data (content_data_id, author_id, published_by) VALUES (?, ?, ?);",
		"01CONTENT00000000000000000", bobID, bobID,
	); err != nil {
		t.Fatalf("insert content: %v", err)
	}

	userRefs := map[string]string{
		"01ALICE0000000000000000000": "alice",
		bobID:                       "bob",
	}

	err := ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
		_, rErr := resolveUserRefs(ctx, ex, ops, userRefs)
		return rErr
	})
	if err != nil {
		t.Fatalf("resolveUserRefs: %v", err)
	}

	t.Run("existing user is not modified", func(t *testing.T) {
		var username string
		if err := pool.QueryRow("SELECT username FROM users WHERE user_id = ?;", "01ALICE0000000000000000000").Scan(&username); err != nil {
			t.Fatalf("query alice: %v", err)
		}
		if username != "alice" {
			t.Errorf("alice username = %q, want %q", username, "alice")
		}
	})

	t.Run("missing user content remapped to admin", func(t *testing.T) {
		var authorID, publishedBy string
		if err := pool.QueryRow(
			"SELECT author_id, published_by FROM content_data WHERE content_data_id = ?;",
			"01CONTENT00000000000000000",
		).Scan(&authorID, &publishedBy); err != nil {
			t.Fatalf("query content: %v", err)
		}
		if authorID != adminUserID {
			t.Errorf("author_id = %q, want admin %q", authorID, adminUserID)
		}
		if publishedBy != adminUserID {
			t.Errorf("published_by = %q, want admin %q", publishedBy, adminUserID)
		}
	})

	t.Run("no placeholder user created", func(t *testing.T) {
		// Should be: admin + alice = 2 users. No bob placeholder.
		if c := countUsers(t, pool); c != 2 {
			t.Errorf("expected 2 users (admin + alice), got %d", c)
		}
	})
}

func TestIntegration_ResolveUserRefs_SQLite_EmptyRefs(t *testing.T) {
	pool, ops, _ := setupSQLiteDeployDB(t)

	beforeCount := countUsers(t, pool)
	err := ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
		_, rErr := resolveUserRefs(ctx, ex, ops, map[string]string{})
		return rErr
	})
	if err != nil {
		t.Fatalf("resolveUserRefs with empty refs: %v", err)
	}
	if after := countUsers(t, pool); after != beforeCount {
		t.Errorf("user count changed: %d -> %d", beforeCount, after)
	}
}

func TestIntegration_ResolveUserRefs_SQLite_AllExisting(t *testing.T) {
	pool, ops, _ := setupSQLiteDeployDB(t)

	adminRoleID := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	seedExistingUser(t, pool, "01EXISTING1000000000000000A", "user_a", adminRoleID)
	seedExistingUser(t, pool, "01EXISTING2000000000000000B", "user_b", adminRoleID)

	userRefs := map[string]string{
		"01EXISTING1000000000000000A": "user_a",
		"01EXISTING2000000000000000B": "user_b",
	}

	var remapped map[string]string
	err := ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
		var rErr error
		remapped, rErr = resolveUserRefs(ctx, ex, ops, userRefs)
		return rErr
	})
	if err != nil {
		t.Fatalf("resolveUserRefs: %v", err)
	}

	if len(remapped) != 0 {
		t.Errorf("expected 0 remapped users, got %d", len(remapped))
	}
	// admin + user_a + user_b = 3
	if c := countUsers(t, pool); c != 3 {
		t.Errorf("expected 3 users (no new ones), got %d", c)
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
