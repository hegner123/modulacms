// Shared integration test helpers for the db package.
//
// These helpers create isolated SQLite databases with the full schema
// bootstrapped, enabling CRUD integration tests across all entity files.
// Each call creates a fresh database in t.TempDir(), safe for t.Parallel().
package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	config "github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// testIntegrationDB creates an isolated SQLite database with all tables.
// The database lives in t.TempDir() (not :memory: -- WAL mode needs a file).
// Connection cleanup is registered via t.Cleanup.
func testIntegrationDB(t *testing.T) Database {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "integration.db")
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

// testAuditCtx returns an AuditContext suitable for entities with no FK to users.
func testAuditCtx(d Database) audited.AuditContext {
	return audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "test", "127.0.0.1")
}

// testAuditCtxWithUser returns an AuditContext with a real UserID from seed data.
func testAuditCtxWithUser(d Database, userID types.UserID) audited.AuditContext {
	return audited.Ctx(types.NodeID(d.Config.Node_ID), userID, "test", "127.0.0.1")
}

// SeedData holds pointers to minimal FK-satisfying records inserted by testSeededDB.
type SeedData struct {
	Permission    *Permissions
	Role          *Roles
	User          *Users
	Route         *Routes
	AdminRoute    *AdminRoutes
	Datatype      *Datatypes
	AdminDatatype *AdminDatatypes
	Field         *Fields
	AdminField    *AdminFields
}

// testSeededDB creates an integration DB and inserts a minimal set of records
// that satisfy foreign key constraints. The returned SeedData provides pointers
// to every created entity for use as FK references in downstream tests.
func testSeededDB(t *testing.T) (Database, SeedData) {
	t.Helper()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)
	now := types.TimestampNow()

	// Tier 0: no FKs
	perm, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
		Label: "test-permission",
	})
	if err != nil {
		t.Fatalf("seed CreatePermission: %v", err)
	}

	role, err := d.CreateRole(ctx, ac, CreateRoleParams{
		Label: "test-role",
	})
	if err != nil {
		t.Fatalf("seed CreateRole: %v", err)
	}

	// Tier 1: user references role by string label
	user, err := d.CreateUser(ctx, ac, CreateUserParams{
		Username:     "testuser",
		Name:         "Test User",
		Email:        types.Email("test@example.com"),
		Hash:         "fakehash",
		Role:         role.RoleID.String(),
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}

	// Switch to user-aware audit context for remaining entities
	acUser := testAuditCtxWithUser(d, user.UserID)
	authorID := types.NullableUserID{ID: user.UserID, Valid: true}

	// Tier 2: entities referencing user
	route, err := d.CreateRoute(ctx, acUser, CreateRouteParams{
		Slug:         types.Slug("test-route"),
		Title:        "Test Route",
		Status:       1,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateRoute: %v", err)
	}

	adminRoute, err := d.CreateAdminRoute(ctx, acUser, CreateAdminRouteParams{
		Slug:         types.Slug("test-admin-route"),
		Title:        "Test Admin Route",
		Status:       1,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateAdminRoute: %v", err)
	}

	datatype, err := d.CreateDatatype(ctx, acUser, CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "test-datatype",
		Type:         "page",
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateDatatype: %v", err)
	}

	adminDatatype, err := d.CreateAdminDatatype(ctx, acUser, CreateAdminDatatypeParams{
		ParentID:     types.NullableAdminDatatypeID{},
		Label:        "test-admin-datatype",
		Type:         "page",
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateAdminDatatype: %v", err)
	}

	field, err := d.CreateField(ctx, acUser, CreateFieldParams{
		FieldID:      types.NewFieldID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "test-field",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateField: %v", err)
	}

	adminField, err := d.CreateAdminField(ctx, acUser, CreateAdminFieldParams{
		ParentID:     types.NullableAdminDatatypeID{},
		Label:        "test-admin-field",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateAdminField: %v", err)
	}

	return d, SeedData{
		Permission:    perm,
		Role:          role,
		User:          user,
		Route:         route,
		AdminRoute:    adminRoute,
		Datatype:      datatype,
		AdminDatatype: adminDatatype,
		Field:         field,
		AdminField:    adminField,
	}
}

// ===== SMOKE TESTS =====

func TestTestIntegrationDB_Smoke(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	permCount, err := d.CountPermissions()
	if err != nil {
		t.Fatalf("CountPermissions: %v", err)
	}
	if *permCount != 0 {
		t.Errorf("CountPermissions = %d, want 0", *permCount)
	}

	roleCount, err := d.CountRoles()
	if err != nil {
		t.Fatalf("CountRoles: %v", err)
	}
	if *roleCount != 0 {
		t.Errorf("CountRoles = %d, want 0", *roleCount)
	}
}

func TestTestSeededDB_Smoke(t *testing.T) {
	t.Parallel()
	_, seed := testSeededDB(t)

	// All seed data fields must be non-nil with non-zero IDs
	if seed.Permission == nil || seed.Permission.PermissionID.IsZero() {
		t.Error("seed.Permission is nil or has zero ID")
	}
	if seed.Role == nil || seed.Role.RoleID.IsZero() {
		t.Error("seed.Role is nil or has zero ID")
	}
	if seed.User == nil || seed.User.UserID.IsZero() {
		t.Error("seed.User is nil or has zero ID")
	}
	if seed.Route == nil || seed.Route.RouteID.IsZero() {
		t.Error("seed.Route is nil or has zero ID")
	}
	if seed.AdminRoute == nil || seed.AdminRoute.AdminRouteID.IsZero() {
		t.Error("seed.AdminRoute is nil or has zero ID")
	}
	if seed.Datatype == nil || seed.Datatype.DatatypeID.IsZero() {
		t.Error("seed.Datatype is nil or has zero ID")
	}
	if seed.AdminDatatype == nil || seed.AdminDatatype.AdminDatatypeID.IsZero() {
		t.Error("seed.AdminDatatype is nil or has zero ID")
	}
	if seed.Field == nil || seed.Field.FieldID.IsZero() {
		t.Error("seed.Field is nil or has zero ID")
	}
	if seed.AdminField == nil || seed.AdminField.AdminFieldID.IsZero() {
		t.Error("seed.AdminField is nil or has zero ID")
	}
}

func TestTestSeededDB_Counts(t *testing.T) {
	t.Parallel()
	d, _ := testSeededDB(t)

	permCount, err := d.CountPermissions()
	if err != nil {
		t.Fatalf("CountPermissions: %v", err)
	}
	if *permCount != 1 {
		t.Errorf("CountPermissions = %d, want 1", *permCount)
	}

	roleCount, err := d.CountRoles()
	if err != nil {
		t.Fatalf("CountRoles: %v", err)
	}
	if *roleCount != 1 {
		t.Errorf("CountRoles = %d, want 1", *roleCount)
	}

	userCount, err := d.CountUsers()
	if err != nil {
		t.Fatalf("CountUsers: %v", err)
	}
	if *userCount != 1 {
		t.Errorf("CountUsers = %d, want 1", *userCount)
	}
}

func TestTestIntegrationDB_Parallel_Isolation(t *testing.T) {
	t.Parallel()

	for i := range 3 {
		label := []string{"alpha", "beta", "gamma"}[i]
		t.Run(label, func(t *testing.T) {
			t.Parallel()
			d := testIntegrationDB(t)
			ctx := d.Context
			ac := testAuditCtx(d)

			_, err := d.CreatePermission(ctx, ac, CreatePermissionParams{
				Label: label,
			})
			if err != nil {
				t.Fatalf("CreatePermission(%s): %v", label, err)
			}

			count, err := d.CountPermissions()
			if err != nil {
				t.Fatalf("CountPermissions: %v", err)
			}
			if *count != 1 {
				t.Errorf("CountPermissions = %d, want 1 (isolation broken)", *count)
			}
		})
	}
}
