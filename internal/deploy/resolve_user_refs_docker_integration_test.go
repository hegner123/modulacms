//go:build integration

package deploy

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/db"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// ---------------------------------------------------------------------------
// resolveUserRefs cross-backend integration test (PostgreSQL + MySQL)
//
// This is the core regression test for the $1 placeholder fix. Before the fix,
// resolveUserRefs hardcoded "?" in its SELECT COUNT(*) query, which worked on
// SQLite and MySQL but produced a syntax error on PostgreSQL (which requires $1).
//
// These tests require Docker containers for PostgreSQL and MySQL.
// Run with: go test -tags=integration -v -timeout 120s ./internal/deploy/...
//
// Environment variables (with defaults matching docker-compose):
//   MODULACMS_TEST_PSQL_DSN  (default: postgres://modula:modula@localhost:5432/modula_db?sslmode=disable)
//   MODULACMS_TEST_MYSQL_DSN (default: modula:modula@tcp(localhost:3306)/modula_db?parseTime=true)
// ---------------------------------------------------------------------------

const (
	defaultPsqlDSN  = "postgres://modula:modula@localhost:5432/modula_db?sslmode=disable"
	defaultMySQLDSN = "modula:modula@tcp(localhost:3306)/modula_db?parseTime=true"
)

func psqlDSN() string {
	if v := os.Getenv("MODULACMS_TEST_PSQL_DSN"); v != "" {
		return v
	}
	return defaultPsqlDSN
}

func mysqlDSN() string {
	if v := os.Getenv("MODULACMS_TEST_MYSQL_DSN"); v != "" {
		return v
	}
	return defaultMySQLDSN
}

// connectWithRetry attempts to open and ping a database connection with retries.
// Docker containers can take time to become ready.
func connectWithRetry(t *testing.T, driverName, dsn string, maxRetries int) *sql.DB {
	t.Helper()
	var pool *sql.DB
	var err error

	for attempt := range maxRetries {
		pool, err = sql.Open(driverName, dsn)
		if err != nil {
			t.Logf("attempt %d: sql.Open(%s): %v", attempt+1, driverName, err)
			time.Sleep(2 * time.Second)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = pool.PingContext(ctx)
		cancel()
		if err == nil {
			return pool
		}

		t.Logf("attempt %d: ping %s: %v", attempt+1, driverName, err)
		pool.Close()
		time.Sleep(2 * time.Second)
	}

	t.Fatalf("failed to connect to %s after %d attempts: %v", driverName, maxRetries, err)
	return nil
}

// ---------------------------------------------------------------------------
// PostgreSQL backend
// ---------------------------------------------------------------------------

// setupPsqlDeployDB connects to the test PostgreSQL instance, creates
// isolated test tables (prefixed to avoid collision), seeds a viewer role,
// and returns the pool, real DeployOps, and viewer role ID.
func setupPsqlDeployDB(t *testing.T) (*sql.DB, db.DeployOps, string) {
	t.Helper()

	pool := connectWithRetry(t, "postgres", psqlDSN(), 5)
	t.Cleanup(func() { pool.Close() })

	ctx := context.Background()

	// Use a unique schema per test to avoid table collisions.
	schema := fmt.Sprintf("test_resolve_%d", time.Now().UnixNano())
	if _, err := pool.ExecContext(ctx, "CREATE SCHEMA "+schema+";"); err != nil {
		t.Fatalf("create schema %s: %v", schema, err)
	}
	t.Cleanup(func() {
		pool.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE;")
	})

	// Set search_path for this connection.
	if _, err := pool.ExecContext(ctx, "SET search_path TO "+schema+";"); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	// Create roles table.
	if _, err := pool.ExecContext(ctx, `CREATE TABLE roles (
		role_id TEXT PRIMARY KEY NOT NULL,
		label TEXT NOT NULL UNIQUE,
		system_protected INTEGER NOT NULL DEFAULT 0
	);`); err != nil {
		t.Fatalf("create roles: %v", err)
	}

	// Create users table matching the PostgreSQL schema.
	if _, err := pool.ExecContext(ctx, `CREATE TABLE users (
		user_id TEXT PRIMARY KEY NOT NULL,
		username TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		hash TEXT NOT NULL,
		role TEXT NOT NULL REFERENCES roles(role_id),
		date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`); err != nil {
		t.Fatalf("create users: %v", err)
	}

	// Seed viewer role.
	viewerRoleID := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	if _, err := pool.ExecContext(ctx,
		"INSERT INTO roles (role_id, label, system_protected) VALUES ($1, $2, $3);",
		viewerRoleID, "viewer", 1,
	); err != nil {
		t.Fatalf("seed viewer role: %v", err)
	}

	// Build real DeployOps via the constructor.
	driver := db.PsqlDatabase{Connection: pool}
	ops, err := db.NewDeployOps(driver)
	if err != nil {
		t.Fatalf("NewDeployOps: %v", err)
	}

	return pool, ops, viewerRoleID
}

// seedExistingUserPsql inserts a user using PostgreSQL placeholders.
func seedExistingUserPsql(t *testing.T, pool *sql.DB, userID, username, roleID string) {
	t.Helper()
	if _, err := pool.Exec(
		"INSERT INTO users (user_id, username, name, email, hash, role, date_created, date_modified) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);",
		userID, username, "Real "+username, username+"@example.com", "bcrypt-hash", roleID,
		"2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z",
	); err != nil {
		t.Fatalf("seed user %s: %v", username, err)
	}
}

// queryUserPsql retrieves a user row from PostgreSQL.
func queryUserPsql(t *testing.T, pool *sql.DB, userID string) *userRow {
	t.Helper()
	var u userRow
	err := pool.QueryRow(
		"SELECT user_id, username, name, email, hash, role FROM users WHERE user_id = $1;",
		userID,
	).Scan(&u.UserID, &u.Username, &u.Name, &u.Email, &u.Hash, &u.Role)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		t.Fatalf("queryUserPsql %s: %v", userID, err)
	}
	return &u
}

// countUsersPsql returns total users from PostgreSQL.
func countUsersPsql(t *testing.T, pool *sql.DB) int64 {
	t.Helper()
	var count int64
	if err := pool.QueryRow("SELECT COUNT(*) FROM users;").Scan(&count); err != nil {
		t.Fatalf("count users psql: %v", err)
	}
	return count
}

// TestIntegration_ResolveUserRefs_PostgreSQL is the primary regression test
// for the $1 placeholder fix. Before the fix, this test would fail with:
//   "could not determine data type of parameter $1"
// because resolveUserRefs hardcoded "?" instead of using ops.Placeholder(1).
func TestIntegration_ResolveUserRefs_PostgreSQL(t *testing.T) {
	pool, ops, viewerRoleID := setupPsqlDeployDB(t)

	// Scenario: sync payload references alice (exists) and bob (missing).
	seedExistingUserPsql(t, pool, "01ALICE0000000000000000000", "alice", viewerRoleID)

	userRefs := map[string]string{
		"01ALICE0000000000000000000": "alice",
		"01BOB00000000000000000000N": "bob",
	}

	// Run through ImportAtomic which uses a real PostgreSQL transaction
	// with session_replication_role = 'replica' to disable FK triggers.
	err := ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
		return resolveUserRefs(ctx, ex, ops, userRefs, viewerRoleID)
	})
	if err != nil {
		t.Fatalf("resolveUserRefs on PostgreSQL: %v", err)
	}

	t.Run("correct user count after resolve", func(t *testing.T) {
		if c := countUsersPsql(t, pool); c != 2 {
			t.Fatalf("expected 2 users, got %d", c)
		}
	})

	t.Run("existing user is unchanged", func(t *testing.T) {
		alice := queryUserPsql(t, pool, "01ALICE0000000000000000000")
		if alice == nil {
			t.Fatal("alice should exist")
		}
		if alice.Username != "alice" {
			t.Errorf("alice username = %q, want %q", alice.Username, "alice")
		}
		if alice.Hash != "bcrypt-hash" {
			t.Errorf("alice hash should be unchanged, got %q", alice.Hash)
		}
	})

	t.Run("placeholder user created with correct properties", func(t *testing.T) {
		bob := queryUserPsql(t, pool, "01BOB00000000000000000000N")
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
	})
}

// TestIntegration_Placeholder_PostgreSQL verifies the real psqlDeployOps
// returns "$N" format placeholders, which is the entire fix being tested.
func TestIntegration_Placeholder_PostgreSQL(t *testing.T) {
	_, ops, _ := setupPsqlDeployDB(t)

	tests := []struct {
		n    int
		want string
	}{
		{1, "$1"},
		{2, "$2"},
		{10, "$10"},
		{100, "$100"},
	}

	for _, tt := range tests {
		got := ops.Placeholder(tt.n)
		if got != tt.want {
			t.Errorf("Placeholder(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// MySQL backend
// ---------------------------------------------------------------------------

// setupMySQLDeployDB connects to the test MySQL instance, creates the
// required tables, seeds a viewer role, and returns the pool, real
// DeployOps, and viewer role ID.
func setupMySQLDeployDB(t *testing.T) (*sql.DB, db.DeployOps, string) {
	t.Helper()

	pool := connectWithRetry(t, "mysql", mysqlDSN(), 5)
	t.Cleanup(func() { pool.Close() })

	ctx := context.Background()

	// Drop and recreate tables for test isolation.
	// MySQL does not have schema-level isolation like PostgreSQL, so we
	// use DROP IF EXISTS + CREATE.
	if _, err := pool.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;"); err != nil {
		t.Fatalf("disable FK checks: %v", err)
	}
	if _, err := pool.ExecContext(ctx, "DROP TABLE IF EXISTS users;"); err != nil {
		t.Fatalf("drop users: %v", err)
	}
	if _, err := pool.ExecContext(ctx, "DROP TABLE IF EXISTS roles;"); err != nil {
		t.Fatalf("drop roles: %v", err)
	}
	if _, err := pool.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;"); err != nil {
		t.Fatalf("enable FK checks: %v", err)
	}

	t.Cleanup(func() {
		pool.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;")
		pool.ExecContext(ctx, "DROP TABLE IF EXISTS users;")
		pool.ExecContext(ctx, "DROP TABLE IF EXISTS roles;")
		pool.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;")
	})

	// Create roles table.
	if _, err := pool.ExecContext(ctx, `CREATE TABLE roles (
		role_id VARCHAR(26) PRIMARY KEY NOT NULL,
		label VARCHAR(255) NOT NULL UNIQUE,
		system_protected TINYINT NOT NULL DEFAULT 0
	);`); err != nil {
		t.Fatalf("create roles: %v", err)
	}

	// Create users table matching the MySQL schema.
	if _, err := pool.ExecContext(ctx, `CREATE TABLE users (
		user_id VARCHAR(26) PRIMARY KEY NOT NULL,
		username VARCHAR(255) NOT NULL UNIQUE,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		hash TEXT NOT NULL,
		role VARCHAR(26) NOT NULL,
		date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
		CONSTRAINT fk_users_role FOREIGN KEY (role) REFERENCES roles (role_id)
	);`); err != nil {
		t.Fatalf("create users: %v", err)
	}

	// Seed viewer role.
	viewerRoleID := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	if _, err := pool.ExecContext(ctx,
		"INSERT INTO roles (role_id, label, system_protected) VALUES (?, ?, ?);",
		viewerRoleID, "viewer", 1,
	); err != nil {
		t.Fatalf("seed viewer role: %v", err)
	}

	// Build real DeployOps via the constructor.
	driver := db.MysqlDatabase{Connection: pool}
	ops, err := db.NewDeployOps(driver)
	if err != nil {
		t.Fatalf("NewDeployOps: %v", err)
	}

	return pool, ops, viewerRoleID
}

// seedExistingUserMySQL inserts a user using MySQL placeholders.
func seedExistingUserMySQL(t *testing.T, pool *sql.DB, userID, username, roleID string) {
	t.Helper()
	if _, err := pool.Exec(
		"INSERT INTO users (user_id, username, name, email, hash, role) VALUES (?, ?, ?, ?, ?, ?);",
		userID, username, "Real "+username, username+"@example.com", "bcrypt-hash", roleID,
	); err != nil {
		t.Fatalf("seed user %s: %v", username, err)
	}
}

// queryUserMySQL retrieves a user row from MySQL.
func queryUserMySQL(t *testing.T, pool *sql.DB, userID string) *userRow {
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
		t.Fatalf("queryUserMySQL %s: %v", userID, err)
	}
	return &u
}

// countUsersMySQL returns total users from MySQL.
func countUsersMySQL(t *testing.T, pool *sql.DB) int64 {
	t.Helper()
	var count int64
	if err := pool.QueryRow("SELECT COUNT(*) FROM users;").Scan(&count); err != nil {
		t.Fatalf("count users mysql: %v", err)
	}
	return count
}

func TestIntegration_ResolveUserRefs_MySQL(t *testing.T) {
	pool, ops, viewerRoleID := setupMySQLDeployDB(t)

	// Scenario: same as PostgreSQL test -- alice exists, bob is missing.
	seedExistingUserMySQL(t, pool, "01ALICE0000000000000000000", "alice", viewerRoleID)

	userRefs := map[string]string{
		"01ALICE0000000000000000000": "alice",
		"01BOB00000000000000000000N": "bob",
	}

	err := ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
		return resolveUserRefs(ctx, ex, ops, userRefs, viewerRoleID)
	})
	if err != nil {
		t.Fatalf("resolveUserRefs on MySQL: %v", err)
	}

	t.Run("correct user count after resolve", func(t *testing.T) {
		if c := countUsersMySQL(t, pool); c != 2 {
			t.Fatalf("expected 2 users, got %d", c)
		}
	})

	t.Run("existing user is unchanged", func(t *testing.T) {
		alice := queryUserMySQL(t, pool, "01ALICE0000000000000000000")
		if alice == nil {
			t.Fatal("alice should exist")
		}
		if alice.Username != "alice" {
			t.Errorf("alice username = %q, want %q", alice.Username, "alice")
		}
	})

	t.Run("placeholder user created with correct properties", func(t *testing.T) {
		bob := queryUserMySQL(t, pool, "01BOB00000000000000000000N")
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
	})
}

// TestIntegration_Placeholder_MySQL verifies MySQL returns "?" for all positions.
func TestIntegration_Placeholder_MySQL(t *testing.T) {
	_, ops, _ := setupMySQLDeployDB(t)

	for _, n := range []int{1, 2, 5, 100} {
		got := ops.Placeholder(n)
		if got != "?" {
			t.Errorf("Placeholder(%d) = %q, want %q", n, got, "?")
		}
	}
}

// ---------------------------------------------------------------------------
// Cross-backend consistency: identical input produces identical results
// ---------------------------------------------------------------------------

// TestIntegration_ResolveUserRefs_CrossBackend runs the same resolveUserRefs
// scenario against all three backends and verifies they produce identical
// outcomes. This is the definitive test that the Placeholder fix works.
func TestIntegration_ResolveUserRefs_CrossBackend(t *testing.T) {
	type backendSetup struct {
		name       string
		pool       *sql.DB
		ops        db.DeployOps
		roleID     string
		seedUser   func(t *testing.T, pool *sql.DB, userID, username, roleID string)
		queryUser  func(t *testing.T, pool *sql.DB, userID string) *userRow
		countUsers func(t *testing.T, pool *sql.DB) int64
	}

	// SQLite (always available).
	sqlitePool, sqliteOps, sqliteRoleID := setupSQLiteDeployDB(t)

	// PostgreSQL (requires Docker).
	psqlPool, psqlOps, psqlRoleID := setupPsqlDeployDB(t)

	// MySQL (requires Docker).
	mysqlPool, mysqlOps, mysqlRoleID := setupMySQLDeployDB(t)

	backends := []backendSetup{
		{
			name: "sqlite", pool: sqlitePool, ops: sqliteOps, roleID: sqliteRoleID,
			seedUser: seedExistingUser, queryUser: queryUser, countUsers: countUsers,
		},
		{
			name: "postgresql", pool: psqlPool, ops: psqlOps, roleID: psqlRoleID,
			seedUser: seedExistingUserPsql, queryUser: queryUserPsql, countUsers: countUsersPsql,
		},
		{
			name: "mysql", pool: mysqlPool, ops: mysqlOps, roleID: mysqlRoleID,
			seedUser: seedExistingUserMySQL, queryUser: queryUserMySQL, countUsers: countUsersMySQL,
		},
	}

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			// Seed identical starting state: one existing user.
			b.seedUser(t, b.pool, "01CROSS0000000000000000000", "crossuser", b.roleID)

			// Same userRefs on all backends.
			userRefs := map[string]string{
				"01CROSS0000000000000000000": "crossuser",
				"01NEWCR0000000000000000000": "newuser",
			}

			err := b.ops.ImportAtomic(context.Background(), func(ctx context.Context, ex db.Executor) error {
				return resolveUserRefs(ctx, ex, b.ops, userRefs, b.roleID)
			})
			if err != nil {
				t.Fatalf("[%s] resolveUserRefs: %v", b.name, err)
			}

			// All backends should have exactly 2 users.
			if c := b.countUsers(t, b.pool); c != 2 {
				t.Errorf("[%s] expected 2 users, got %d", b.name, c)
			}

			// Existing user should be unchanged.
			existing := b.queryUser(t, b.pool, "01CROSS0000000000000000000")
			if existing == nil {
				t.Fatalf("[%s] existing user vanished", b.name)
			}
			if existing.Username != "crossuser" {
				t.Errorf("[%s] existing username = %q, want %q", b.name, existing.Username, "crossuser")
			}

			// New user should have identical placeholder properties across backends.
			newUser := b.queryUser(t, b.pool, "01NEWCR0000000000000000000")
			if newUser == nil {
				t.Fatalf("[%s] new placeholder user not created", b.name)
			}
			if newUser.Username != "[sync] newuser" {
				t.Errorf("[%s] username = %q, want %q", b.name, newUser.Username, "[sync] newuser")
			}
			if newUser.Email != "newuser@sync.placeholder" {
				t.Errorf("[%s] email = %q, want %q", b.name, newUser.Email, "newuser@sync.placeholder")
			}
			if newUser.Hash != "" {
				t.Errorf("[%s] hash = %q, want empty", b.name, newUser.Hash)
			}
			if newUser.Role != b.roleID {
				t.Errorf("[%s] role = %q, want %q", b.name, newUser.Role, b.roleID)
			}
		})
	}
}
