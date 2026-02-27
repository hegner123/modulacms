// Cross-backend integration tests for the db package.
//
// These tests verify that MySQL, PostgreSQL, and SQLite produce identical
// results for the same operations. This is the primary value add of multi-
// backend integration tests: catching dialect-specific bugs that unit tests
// against a single backend cannot detect.
//
// Requires Docker containers running MySQL and PostgreSQL.
// Start them with: just docker-infra
//
// Run with: go test -tags integration -v -timeout 120s ./internal/db/ -run TestCrossBackend
//
// Environment variables for database connections (defaults match docker-compose.full.yml):
//
//	INTEGRATION_MYSQL_DSN    default: modula:modula@tcp(localhost:3306)/modula_integration_test?parseTime=true
//	INTEGRATION_POSTGRES_DSN default: postgres://modula:modula@localhost:5432/modula_integration_test?sslmode=disable
//
//go:build integration

package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	config "github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ---------------------------------------------------------------------------
// Backend setup helpers
// ---------------------------------------------------------------------------

// backendInfo pairs a human-readable name with a DbDriver and its context.
type backendInfo struct {
	name   string
	driver DbDriver
	ctx    context.Context
}

// tryMySQLDriver attempts to create an isolated MySQL database with all tables.
// Returns the driver and true on success, or a zero value and false if MySQL
// is not available (Docker not running).
func tryMySQLDriver(t *testing.T) (MysqlDatabase, bool) {
	t.Helper()

	dsn := os.Getenv("INTEGRATION_MYSQL_DSN")
	if dsn == "" {
		dsn = "modula:modula@tcp(localhost:3306)/?parseTime=true"
	}

	// Connect to the MySQL server (no database selected)
	adminConn, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Logf("MySQL sql.Open failed (skipping MySQL): %v", err)
		return MysqlDatabase{}, false
	}

	if err := adminConn.Ping(); err != nil {
		adminConn.Close()
		t.Logf("MySQL not available (start with `just docker-infra`): %v", err)
		return MysqlDatabase{}, false
	}

	// Create a unique test database
	dbName := fmt.Sprintf("inttest_%s", strings.ToLower(types.NewULID().String()))
	_, err = adminConn.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		adminConn.Close()
		t.Logf("MySQL CREATE DATABASE %s failed (skipping MySQL): %v", dbName, err)
		return MysqlDatabase{}, false
	}

	t.Cleanup(func() {
		adminConn.Exec("DROP DATABASE IF EXISTS " + dbName)
		adminConn.Close()
	})

	// Connect to the test database
	testDSN := fmt.Sprintf("modula:modula@tcp(localhost:3306)/%s?parseTime=true", dbName)
	conn, err := sql.Open("mysql", testDSN)
	if err != nil {
		t.Logf("MySQL sql.Open (test db %s) failed: %v", dbName, err)
		return MysqlDatabase{}, false
	}
	t.Cleanup(func() { conn.Close() })

	ctx := context.Background()
	d := MysqlDatabase{
		Connection: conn,
		Context:    ctx,
		Config:     config.Config{Node_ID: types.NewNodeID().String()},
	}

	if err := d.CreateAllTables(); err != nil {
		t.Logf("MySQL CreateAllTables failed (skipping MySQL): %v", err)
		return MysqlDatabase{}, false
	}

	return d, true
}

// tryPsqlDriver attempts to create an isolated PostgreSQL database with all
// tables. Returns the driver and true on success, or a zero value and false
// if PostgreSQL is not available.
func tryPsqlDriver(t *testing.T) (PsqlDatabase, bool) {
	t.Helper()

	dsn := os.Getenv("INTEGRATION_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://modula:modula@localhost:5432/modula_db?sslmode=disable"
	}

	// Connect to the default database for admin operations
	adminConn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Logf("PostgreSQL sql.Open failed (skipping PostgreSQL): %v", err)
		return PsqlDatabase{}, false
	}

	if err := adminConn.Ping(); err != nil {
		adminConn.Close()
		t.Logf("PostgreSQL not available (start with `just docker-infra`): %v", err)
		return PsqlDatabase{}, false
	}

	// Create a unique test database
	dbName := fmt.Sprintf("inttest_%s", strings.ToLower(types.NewULID().String()))
	_, err = adminConn.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		adminConn.Close()
		t.Logf("PostgreSQL CREATE DATABASE %s failed: %v", dbName, err)
		return PsqlDatabase{}, false
	}

	t.Cleanup(func() {
		// Must close test conn before dropping database
		adminConn.Exec("DROP DATABASE IF EXISTS " + dbName)
		adminConn.Close()
	})

	// Connect to the test database
	testDSN := fmt.Sprintf("postgres://modula:modula@localhost:5432/%s?sslmode=disable", dbName)
	conn, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Logf("PostgreSQL sql.Open (test db %s) failed: %v", dbName, err)
		return PsqlDatabase{}, false
	}
	t.Cleanup(func() { conn.Close() })

	ctx := context.Background()
	d := PsqlDatabase{
		Connection: conn,
		Context:    ctx,
		Config:     config.Config{Node_ID: types.NewNodeID().String()},
	}

	if err := d.CreateAllTables(); err != nil {
		t.Logf("PostgreSQL CreateAllTables failed (skipping PostgreSQL): %v", err)
		return PsqlDatabase{}, false
	}

	return d, true
}

// allBackends returns backend drivers for all available databases.
// SQLite is always available; MySQL and PostgreSQL are included only if
// their Docker containers are running. Skips the test if fewer than 2
// backends are available.
func allBackends(t *testing.T) []backendInfo {
	t.Helper()

	backends := []backendInfo{
		{
			name:   "sqlite",
			driver: testIntegrationDB(t),
			ctx:    context.Background(),
		},
	}

	if m, ok := tryMySQLDriver(t); ok {
		backends = append(backends, backendInfo{
			name:   "mysql",
			driver: m,
			ctx:    m.Context,
		})
	}

	if p, ok := tryPsqlDriver(t); ok {
		backends = append(backends, backendInfo{
			name:   "postgres",
			driver: p,
			ctx:    p.Context,
		})
	}

	if len(backends) < 2 {
		t.Skip("At least 2 backends required for cross-backend tests; start MySQL/PostgreSQL with `just docker-infra`")
	}

	return backends
}

// backendAuditCtx creates an AuditContext for a backend with no user.
func backendAuditCtx(driver DbDriver) audited.AuditContext {
	conn, ctx, err := driver.GetConnection()
	if err != nil || conn == nil || ctx == nil {
		// Fallback: use a generic node ID
		return audited.Ctx(types.NewNodeID(), types.UserID(""), "integration-test", "127.0.0.1")
	}
	return audited.Ctx(types.NewNodeID(), types.UserID(""), "integration-test", "127.0.0.1")
}

// backendAuditCtxWithUser creates an AuditContext with a real UserID.
func backendAuditCtxWithUser(userID types.UserID) audited.AuditContext {
	return audited.Ctx(types.NewNodeID(), userID, "integration-test", "127.0.0.1")
}

// ---------------------------------------------------------------------------
// Seed data helpers (backend-agnostic, using DbDriver interface)
// ---------------------------------------------------------------------------

// crossBackendSeed holds the seeded entities for cross-backend tests.
type crossBackendSeed struct {
	Driver     DbDriver
	Ctx        context.Context
	AC         audited.AuditContext
	User       *Users
	Role       *Roles
	Route      *Routes
	RouteID    types.NullableRouteID
	DtRootID   types.NullableDatatypeID
	DtPageID   types.NullableDatatypeID
	FieldID    types.NullableFieldID
	Permission *Permissions
}

// seedCrossBackend creates the minimum set of entities needed for content tree
// and audit trail tests on any backend. Uses only DbDriver interface methods.
func seedCrossBackend(t *testing.T, driver DbDriver) crossBackendSeed {
	t.Helper()
	_, ctx, err := driver.GetConnection()
	if err != nil {
		t.Fatalf("GetConnection: %v", err)
	}
	ac := backendAuditCtx(driver)
	now := types.TimestampNow()

	perm, err := driver.CreatePermission(ctx, ac, CreatePermissionParams{
		Label: "cross-test-perm",
	})
	if err != nil {
		t.Fatalf("seed CreatePermission: %v", err)
	}

	role, err := driver.CreateRole(ctx, ac, CreateRoleParams{Label: "cross-test-role"})
	if err != nil {
		t.Fatalf("seed CreateRole: %v", err)
	}

	user, err := driver.CreateUser(ctx, ac, CreateUserParams{
		Username:     "crossuser",
		Name:         "Cross Backend User",
		Email:        types.Email("cross@test.com"),
		Hash:         "fakehash",
		Role:         role.RoleID.String(),
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}

	acUser := backendAuditCtxWithUser(user.UserID)

	route, err := driver.CreateRoute(ctx, acUser, CreateRouteParams{
		Slug:         types.Slug("cross-test"),
		Title:        "Cross Backend Route",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateRoute: %v", err)
	}

	dtRoot, err := driver.CreateDatatype(ctx, acUser, CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "Root",
		Type:         "_root",
		AuthorID:     user.UserID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateDatatype(_root): %v", err)
	}

	dtPage, err := driver.CreateDatatype(ctx, acUser, CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "Page",
		Type:         "page",
		AuthorID:     user.UserID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateDatatype(page): %v", err)
	}

	field, err := driver.CreateField(ctx, acUser, CreateFieldParams{
		FieldID:      types.NewFieldID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "title",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateField: %v", err)
	}

	return crossBackendSeed{
		Driver:     driver,
		Ctx:        ctx,
		AC:         acUser,
		User:       user,
		Role:       role,
		Route:      route,
		RouteID:    types.NullableRouteID{ID: route.RouteID, Valid: true},
		DtRootID:   types.NullableDatatypeID{ID: dtRoot.DatatypeID, Valid: true},
		DtPageID:   types.NullableDatatypeID{ID: dtPage.DatatypeID, Valid: true},
		FieldID:    types.NullableFieldID{ID: field.FieldID, Valid: true},
		Permission: perm,
	}
}

// createNode creates a content_data row on any backend via DbDriver.
func createNode(t *testing.T, s crossBackendSeed, datatypeID types.NullableDatatypeID, parentID, firstChildID, nextSiblingID, prevSiblingID types.NullableContentID) *ContentData {
	t.Helper()
	now := types.TimestampNow()

	created, err := s.Driver.CreateContentData(s.Ctx, s.AC, CreateContentDataParams{
		RouteID:       s.RouteID,
		ParentID:      parentID,
		FirstChildID:  firstChildID,
		NextSiblingID: nextSiblingID,
		PrevSiblingID: prevSiblingID,
		DatatypeID:    datatypeID,
		AuthorID:      s.User.UserID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("createNode: %v", err)
	}
	return created
}

// nullCID returns an invalid NullableContentID (no reference).
func nullCID() types.NullableContentID {
	return types.NullableContentID{Valid: false}
}

// validCID wraps a ContentID in a valid NullableContentID.
func validCID(id types.ContentID) types.NullableContentID {
	return types.NullableContentID{ID: id, Valid: true}
}

// ---------------------------------------------------------------------------
// Cross-backend consistency tests
// ---------------------------------------------------------------------------

// TestCrossBackend_RoleUserRoute_RoundTrip verifies that creating and reading
// back roles, users, and routes produces identical results across all backends.
func TestCrossBackend_RoleUserRoute_RoundTrip(t *testing.T) {
	backends := allBackends(t)

	type result struct {
		name       string
		roleLabel  string
		userName   string
		userEmail  string
		routeTitle string
		routeSlug  string
	}

	var results []result

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)

			// Read back the role
			roles, err := b.driver.ListRoles()
			if err != nil {
				t.Fatalf("ListRoles: %v", err)
			}
			if roles == nil || len(*roles) == 0 {
				t.Fatal("expected at least 1 role")
			}
			// Find our role
			var foundRole *Roles
			for i := range *roles {
				if (*roles)[i].RoleID == s.Role.RoleID {
					foundRole = &(*roles)[i]
					break
				}
			}
			if foundRole == nil {
				t.Fatal("seeded role not found in ListRoles")
			}

			// Read back the user
			users, err := b.driver.ListUsers()
			if err != nil {
				t.Fatalf("ListUsers: %v", err)
			}
			if users == nil || len(*users) == 0 {
				t.Fatal("expected at least 1 user")
			}
			var foundUser *Users
			for i := range *users {
				if (*users)[i].UserID == s.User.UserID {
					foundUser = &(*users)[i]
					break
				}
			}
			if foundUser == nil {
				t.Fatal("seeded user not found in ListUsers")
			}

			// Read back the route
			routes, err := b.driver.ListRoutes()
			if err != nil {
				t.Fatalf("ListRoutes: %v", err)
			}
			if routes == nil || len(*routes) == 0 {
				t.Fatal("expected at least 1 route")
			}
			var foundRoute *Routes
			for i := range *routes {
				if (*routes)[i].RouteID == s.Route.RouteID {
					foundRoute = &(*routes)[i]
					break
				}
			}
			if foundRoute == nil {
				t.Fatal("seeded route not found in ListRoutes")
			}

			results = append(results, result{
				name:       b.name,
				roleLabel:  foundRole.Label,
				userName:   foundUser.Username,
				userEmail:  string(foundUser.Email),
				routeTitle: foundRoute.Title,
				routeSlug:  string(foundRoute.Slug),
			})
		})
	}

	// Verify all backends produced the same values
	if len(results) < 2 {
		return // only one backend available, nothing to compare
	}
	ref := results[0]
	for _, r := range results[1:] {
		if r.roleLabel != ref.roleLabel {
			t.Errorf("role label: %s=%q vs %s=%q", ref.name, ref.roleLabel, r.name, r.roleLabel)
		}
		if r.userName != ref.userName {
			t.Errorf("username: %s=%q vs %s=%q", ref.name, ref.userName, r.name, r.userName)
		}
		if r.userEmail != ref.userEmail {
			t.Errorf("email: %s=%q vs %s=%q", ref.name, ref.userEmail, r.name, r.userEmail)
		}
		if r.routeTitle != ref.routeTitle {
			t.Errorf("route title: %s=%q vs %s=%q", ref.name, ref.routeTitle, r.name, r.routeTitle)
		}
		if r.routeSlug != ref.routeSlug {
			t.Errorf("route slug: %s=%q vs %s=%q", ref.name, ref.routeSlug, r.name, r.routeSlug)
		}
	}
}

// TestCrossBackend_ContentData_CreateReadDelete verifies that content_data CRUD
// operations produce consistent results across all backends.
func TestCrossBackend_ContentData_CreateReadDelete(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)

			// Create a root node
			root := createNode(t, s, s.DtRootID, nullCID(), nullCID(), nullCID(), nullCID())
			if root.ContentDataID.IsZero() {
				t.Fatal("root ContentDataID is zero")
			}

			// Read it back
			got, err := s.Driver.GetContentData(root.ContentDataID)
			if err != nil {
				t.Fatalf("GetContentData: %v", err)
			}
			if got.ContentDataID != root.ContentDataID {
				t.Errorf("ID mismatch: got %v, want %v", got.ContentDataID, root.ContentDataID)
			}
			if got.Status != types.ContentStatusPublished {
				t.Errorf("status = %q, want %q", got.Status, types.ContentStatusPublished)
			}
			if !got.RouteID.Valid || got.RouteID.ID != s.Route.RouteID {
				t.Errorf("route_id mismatch")
			}

			// Delete it
			err = s.Driver.DeleteContentData(s.Ctx, s.AC, root.ContentDataID)
			if err != nil {
				t.Fatalf("DeleteContentData: %v", err)
			}

			// Verify it is gone
			count, err := s.Driver.CountContentData()
			if err != nil {
				t.Fatalf("CountContentData: %v", err)
			}
			if *count != 0 {
				t.Errorf("CountContentData = %d after delete, want 0", *count)
			}
		})
	}
}

// TestCrossBackend_ContentTree_SiblingPointers verifies that sibling pointer
// chains produce identical tree structures across all backends.
func TestCrossBackend_ContentTree_SiblingPointers(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)

			// Create root with 3 children
			root := createNode(t, s, s.DtRootID, nullCID(), nullCID(), nullCID(), nullCID())
			nodeA := createNode(t, s, s.DtPageID, validCID(root.ContentDataID), nullCID(), nullCID(), nullCID())
			nodeB := createNode(t, s, s.DtPageID, validCID(root.ContentDataID), nullCID(), nullCID(), nullCID())
			nodeC := createNode(t, s, s.DtPageID, validCID(root.ContentDataID), nullCID(), nullCID(), nullCID())

			now := types.TimestampNow()

			// Set sibling order: C -> A -> B
			// Root's first_child = C
			_, err := s.Driver.UpdateContentData(s.Ctx, s.AC, UpdateContentDataParams{
				ContentDataID: root.ContentDataID,
				RouteID:       s.RouteID,
				ParentID:      nullCID(),
				FirstChildID:  validCID(nodeC.ContentDataID),
				NextSiblingID: nullCID(),
				PrevSiblingID: nullCID(),
				DatatypeID:    s.DtRootID,
				AuthorID:      s.User.UserID,
				Status:        types.ContentStatusPublished,
				DateCreated:   now,
				DateModified:  now,
			})
			if err != nil {
				t.Fatalf("UpdateContentData(root): %v", err)
			}

			// C: prev=nil, next=A
			_, err = s.Driver.UpdateContentData(s.Ctx, s.AC, UpdateContentDataParams{
				ContentDataID: nodeC.ContentDataID,
				RouteID:       s.RouteID,
				ParentID:      validCID(root.ContentDataID),
				FirstChildID:  nullCID(),
				NextSiblingID: validCID(nodeA.ContentDataID),
				PrevSiblingID: nullCID(),
				DatatypeID:    s.DtPageID,
				AuthorID:      s.User.UserID,
				Status:        types.ContentStatusPublished,
				DateCreated:   now,
				DateModified:  now,
			})
			if err != nil {
				t.Fatalf("UpdateContentData(C): %v", err)
			}

			// A: prev=C, next=B
			_, err = s.Driver.UpdateContentData(s.Ctx, s.AC, UpdateContentDataParams{
				ContentDataID: nodeA.ContentDataID,
				RouteID:       s.RouteID,
				ParentID:      validCID(root.ContentDataID),
				FirstChildID:  nullCID(),
				NextSiblingID: validCID(nodeB.ContentDataID),
				PrevSiblingID: validCID(nodeC.ContentDataID),
				DatatypeID:    s.DtPageID,
				AuthorID:      s.User.UserID,
				Status:        types.ContentStatusPublished,
				DateCreated:   now,
				DateModified:  now,
			})
			if err != nil {
				t.Fatalf("UpdateContentData(A): %v", err)
			}

			// B: prev=A, next=nil
			_, err = s.Driver.UpdateContentData(s.Ctx, s.AC, UpdateContentDataParams{
				ContentDataID: nodeB.ContentDataID,
				RouteID:       s.RouteID,
				ParentID:      validCID(root.ContentDataID),
				FirstChildID:  nullCID(),
				NextSiblingID: nullCID(),
				PrevSiblingID: validCID(nodeA.ContentDataID),
				DatatypeID:    s.DtPageID,
				AuthorID:      s.User.UserID,
				Status:        types.ContentStatusPublished,
				DateCreated:   now,
				DateModified:  now,
			})
			if err != nil {
				t.Fatalf("UpdateContentData(B): %v", err)
			}

			// Read back and verify sibling chain via GetContentTreeByRoute
			rows, err := s.Driver.GetContentTreeByRoute(s.RouteID)
			if err != nil {
				t.Fatalf("GetContentTreeByRoute: %v", err)
			}
			if rows == nil || len(*rows) != 4 {
				count := 0
				if rows != nil {
					count = len(*rows)
				}
				t.Fatalf("expected 4 rows, got %d", count)
			}

			// Build a lookup map: ContentDataID -> row
			rowMap := make(map[types.ContentID]GetContentTreeByRouteRow)
			for _, r := range *rows {
				rowMap[r.ContentDataID] = r
			}

			// Verify root's first_child is C
			rootRow := rowMap[root.ContentDataID]
			if !rootRow.FirstChildID.Valid || rootRow.FirstChildID.ID != nodeC.ContentDataID {
				t.Errorf("root.first_child = %v, want %v", rootRow.FirstChildID, nodeC.ContentDataID)
			}

			// Verify C -> A -> B chain
			cRow := rowMap[nodeC.ContentDataID]
			if !cRow.NextSiblingID.Valid || cRow.NextSiblingID.ID != nodeA.ContentDataID {
				t.Errorf("C.next_sibling = %v, want %v", cRow.NextSiblingID, nodeA.ContentDataID)
			}
			if cRow.PrevSiblingID.Valid {
				t.Errorf("C.prev_sibling should be NULL, got %v", cRow.PrevSiblingID.ID)
			}

			aRow := rowMap[nodeA.ContentDataID]
			if !aRow.NextSiblingID.Valid || aRow.NextSiblingID.ID != nodeB.ContentDataID {
				t.Errorf("A.next_sibling = %v, want %v", aRow.NextSiblingID, nodeB.ContentDataID)
			}
			if !aRow.PrevSiblingID.Valid || aRow.PrevSiblingID.ID != nodeC.ContentDataID {
				t.Errorf("A.prev_sibling = %v, want %v", aRow.PrevSiblingID, nodeC.ContentDataID)
			}

			bRow := rowMap[nodeB.ContentDataID]
			if bRow.NextSiblingID.Valid {
				t.Errorf("B.next_sibling should be NULL, got %v", bRow.NextSiblingID.ID)
			}
			if !bRow.PrevSiblingID.Valid || bRow.PrevSiblingID.ID != nodeA.ContentDataID {
				t.Errorf("B.prev_sibling = %v, want %v", bRow.PrevSiblingID, nodeA.ContentDataID)
			}
		})
	}
}

// TestCrossBackend_AuditTrail_CreateUpdateDelete verifies that the audited
// command pipeline records identical change events across all backends.
func TestCrossBackend_AuditTrail_CreateUpdateDelete(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)
			now := types.TimestampNow()

			// Create
			created, err := s.Driver.CreateContentData(s.Ctx, s.AC, CreateContentDataParams{
				RouteID:       s.RouteID,
				ParentID:      nullCID(),
				FirstChildID:  nullCID(),
				NextSiblingID: nullCID(),
				PrevSiblingID: nullCID(),
				DatatypeID:    s.DtPageID,
				AuthorID:      s.User.UserID,
				Status:        types.ContentStatusDraft,
				DateCreated:   now,
				DateModified:  now,
			})
			if err != nil {
				t.Fatalf("Create: %v", err)
			}

			// Update: draft -> published
			_, err = s.Driver.UpdateContentData(s.Ctx, s.AC, UpdateContentDataParams{
				ContentDataID: created.ContentDataID,
				RouteID:       s.RouteID,
				ParentID:      nullCID(),
				FirstChildID:  nullCID(),
				NextSiblingID: nullCID(),
				PrevSiblingID: nullCID(),
				DatatypeID:    s.DtPageID,
				AuthorID:      s.User.UserID,
				Status:        types.ContentStatusPublished,
				DateCreated:   now,
				DateModified:  types.TimestampNow(),
			})
			if err != nil {
				t.Fatalf("Update: %v", err)
			}

			// Delete
			err = s.Driver.DeleteContentData(s.Ctx, s.AC, created.ContentDataID)
			if err != nil {
				t.Fatalf("Delete: %v", err)
			}

			// Verify audit trail: should have 3 events (create + update + delete)
			events, err := s.Driver.GetChangeEventsByRecord("content_data", string(created.ContentDataID))
			if err != nil {
				t.Fatalf("GetChangeEventsByRecord: %v", err)
			}
			if events == nil || len(*events) != 3 {
				count := 0
				if events != nil {
					count = len(*events)
				}
				t.Fatalf("expected 3 change events, got %d", count)
			}

			// Verify all three operation types present
			opCounts := make(map[types.Operation]int)
			for _, ev := range *events {
				opCounts[ev.Operation]++
			}
			if opCounts[types.OpInsert] != 1 {
				t.Errorf("INSERT count = %d, want 1", opCounts[types.OpInsert])
			}
			if opCounts[types.OpUpdate] != 1 {
				t.Errorf("UPDATE count = %d, want 1", opCounts[types.OpUpdate])
			}
			if opCounts[types.OpDelete] != 1 {
				t.Errorf("DELETE count = %d, want 1", opCounts[types.OpDelete])
			}
		})
	}
}

// TestCrossBackend_ContentField_RoundTrip verifies that content fields can be
// created and read back identically across all backends.
func TestCrossBackend_ContentField_RoundTrip(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)

			root := createNode(t, s, s.DtRootID, nullCID(), nullCID(), nullCID(), nullCID())
			now := types.TimestampNow()

			// Create a content field
			cf, err := s.Driver.CreateContentField(s.Ctx, s.AC, CreateContentFieldParams{
				RouteID:       s.RouteID,
				ContentDataID: validCID(root.ContentDataID),
				FieldID:       s.FieldID,
				FieldValue:    "Hello from cross-backend test",
				AuthorID:      s.User.UserID,
				DateCreated:   now,
				DateModified:  now,
			})
			if err != nil {
				t.Fatalf("CreateContentField: %v", err)
			}
			if cf.ContentFieldID.IsZero() {
				t.Fatal("ContentFieldID is zero")
			}

			// Read back via ListContentFieldsByContentData
			fields, err := s.Driver.ListContentFieldsByContentData(validCID(root.ContentDataID))
			if err != nil {
				t.Fatalf("ListContentFieldsByContentData: %v", err)
			}
			if fields == nil || len(*fields) != 1 {
				count := 0
				if fields != nil {
					count = len(*fields)
				}
				t.Fatalf("expected 1 content field, got %d", count)
			}

			gotField := (*fields)[0]
			if gotField.FieldValue != "Hello from cross-backend test" {
				t.Errorf("field_value = %q, want %q", gotField.FieldValue, "Hello from cross-backend test")
			}
			if gotField.ContentFieldID != cf.ContentFieldID {
				t.Errorf("ContentFieldID mismatch")
			}
		})
	}
}

// TestCrossBackend_Datatype_CreateAndList verifies that datatypes round-trip
// consistently across all backends.
func TestCrossBackend_Datatype_CreateAndList(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)

			// List datatypes -- should include _root and page from seed
			dtList, err := s.Driver.ListDatatypes()
			if err != nil {
				t.Fatalf("ListDatatypes: %v", err)
			}
			if dtList == nil || len(*dtList) < 2 {
				count := 0
				if dtList != nil {
					count = len(*dtList)
				}
				t.Fatalf("expected at least 2 datatypes, got %d", count)
			}

			// Find _root and page
			var foundRoot, foundPage bool
			for _, dt := range *dtList {
				if dt.Type == "_root" && dt.Label == "Root" {
					foundRoot = true
				}
				if dt.Type == "page" && dt.Label == "Page" {
					foundPage = true
				}
			}
			if !foundRoot {
				t.Error("_root datatype not found")
			}
			if !foundPage {
				t.Error("page datatype not found")
			}

			// Get by ID
			rootDT, err := s.Driver.GetDatatype(s.DtRootID.ID)
			if err != nil {
				t.Fatalf("GetDatatype(_root): %v", err)
			}
			if rootDT.Type != "_root" {
				t.Errorf("_root type = %q", rootDT.Type)
			}
			if rootDT.Label != "Root" {
				t.Errorf("_root label = %q", rootDT.Label)
			}
		})
	}
}

// TestCrossBackend_Count_EmptyDatabase verifies that count operations return
// zero for empty tables across all backends.
func TestCrossBackend_Count_EmptyDatabase(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			// Use a freshly-created database with no seed data
			_, ctx, err := b.driver.GetConnection()
			if err != nil {
				t.Fatalf("GetConnection: %v", err)
			}
			_ = ctx

			// ContentData should be zero (no seed data for content)
			cdCount, err := b.driver.CountContentData()
			if err != nil {
				t.Fatalf("CountContentData: %v", err)
			}
			if *cdCount != 0 {
				t.Errorf("CountContentData = %d, want 0", *cdCount)
			}
		})
	}
}

// TestCrossBackend_Permission_CRUD exercises permission create, list, and
// delete across backends.
func TestCrossBackend_Permission_CRUD(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			_, ctx, err := b.driver.GetConnection()
			if err != nil {
				t.Fatalf("GetConnection: %v", err)
			}
			ac := backendAuditCtx(b.driver)

			// Create
			perm, err := b.driver.CreatePermission(ctx, ac, CreatePermissionParams{
				Label: "test:read",
			})
			if err != nil {
				t.Fatalf("CreatePermission: %v", err)
			}
			if perm.PermissionID.IsZero() {
				t.Fatal("PermissionID is zero")
			}

			// Count
			count, err := b.driver.CountPermissions()
			if err != nil {
				t.Fatalf("CountPermissions: %v", err)
			}
			if *count < 1 {
				t.Errorf("CountPermissions = %d, want >= 1", *count)
			}

			// List
			perms, err := b.driver.ListPermissions()
			if err != nil {
				t.Fatalf("ListPermissions: %v", err)
			}
			var found bool
			for _, p := range *perms {
				if p.PermissionID == perm.PermissionID {
					found = true
					if p.Label != "test:read" {
						t.Errorf("label = %q, want %q", p.Label, "test:read")
					}
					break
				}
			}
			if !found {
				t.Error("created permission not found in list")
			}

			// Delete
			err = b.driver.DeletePermission(ctx, ac, perm.PermissionID)
			if err != nil {
				t.Fatalf("DeletePermission: %v", err)
			}
		})
	}
}

// TestCrossBackend_ForeignKeyEnforcement verifies that foreign key constraints
// are enforced identically across all backends. A content_data row referencing
// a non-existent route should fail.
func TestCrossBackend_ForeignKeyEnforcement(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)
			now := types.TimestampNow()

			// Try to create content_data with a non-existent route_id
			badRouteID := types.NullableRouteID{ID: types.NewRouteID(), Valid: true}

			_, err := s.Driver.CreateContentData(s.Ctx, s.AC, CreateContentDataParams{
				RouteID:       badRouteID,
				ParentID:      nullCID(),
				FirstChildID:  nullCID(),
				NextSiblingID: nullCID(),
				PrevSiblingID: nullCID(),
				DatatypeID:    s.DtPageID,
				AuthorID:      s.User.UserID,
				Status:        types.ContentStatusDraft,
				DateCreated:   now,
				DateModified:  now,
			})
			if err == nil {
				t.Fatal("expected FK violation error, got nil")
			}

			// Verify no content_data was created (transaction rolled back)
			count, err := s.Driver.CountContentData()
			if err != nil {
				t.Fatalf("CountContentData: %v", err)
			}
			if *count != 0 {
				t.Errorf("CountContentData = %d after FK violation, want 0", *count)
			}
		})
	}
}

// TestCrossBackend_NullHandling verifies that NULL values round-trip correctly
// across all backends. This catches dialect-specific NULL coercion bugs.
func TestCrossBackend_NullHandling(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)

			// Create a content_data with all nullable pointers set to NULL
			root := createNode(t, s, s.DtRootID, nullCID(), nullCID(), nullCID(), nullCID())

			got, err := s.Driver.GetContentData(root.ContentDataID)
			if err != nil {
				t.Fatalf("GetContentData: %v", err)
			}

			// All pointer fields should be NULL
			if got.ParentID.Valid {
				t.Errorf("parent_id should be NULL, got %v", got.ParentID.ID)
			}
			if got.FirstChildID.Valid {
				t.Errorf("first_child_id should be NULL, got %v", got.FirstChildID.ID)
			}
			if got.NextSiblingID.Valid {
				t.Errorf("next_sibling_id should be NULL, got %v", got.NextSiblingID.ID)
			}
			if got.PrevSiblingID.Valid {
				t.Errorf("prev_sibling_id should be NULL, got %v", got.PrevSiblingID.ID)
			}

			// Now set parent_id to a valid value and verify others stay NULL
			now := types.TimestampNow()
			child := createNode(t, s, s.DtPageID, validCID(root.ContentDataID), nullCID(), nullCID(), nullCID())

			gotChild, err := s.Driver.GetContentData(child.ContentDataID)
			if err != nil {
				t.Fatalf("GetContentData(child): %v", err)
			}
			if !gotChild.ParentID.Valid || gotChild.ParentID.ID != root.ContentDataID {
				t.Errorf("child parent_id = %v, want %v", gotChild.ParentID, root.ContentDataID)
			}
			if gotChild.FirstChildID.Valid {
				t.Errorf("child first_child_id should be NULL")
			}
			if gotChild.NextSiblingID.Valid {
				t.Errorf("child next_sibling_id should be NULL")
			}
			if gotChild.PrevSiblingID.Valid {
				t.Errorf("child prev_sibling_id should be NULL")
			}

			_ = now // used implicitly via createNode
		})
	}
}

// TestCrossBackend_Pagination verifies that paginated list queries return
// consistent results across backends.
func TestCrossBackend_Pagination(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)
			now := types.TimestampNow()

			// Create 3 additional datatypes (seed already has _root and page = 2)
			for i := range 3 {
				_, err := s.Driver.CreateDatatype(s.Ctx, s.AC, CreateDatatypeParams{
					DatatypeID:   types.NewDatatypeID(),
					ParentID:     types.NullableDatatypeID{},
					Label:        fmt.Sprintf("pagtest-%d", i),
					Type:         fmt.Sprintf("pagtest-%d", i),
					AuthorID:     s.User.UserID,
					DateCreated:  now,
					DateModified: now,
				})
				if err != nil {
					t.Fatalf("CreateDatatype(%d): %v", i, err)
				}
			}

			// Total datatypes = 5 (2 from seed + 3 created)

			// Page 1: limit 2, offset 0
			page1, err := s.Driver.ListDatatypesPaginated(PaginationParams{Limit: 2, Offset: 0})
			if err != nil {
				t.Fatalf("ListDatatypesPaginated page1: %v", err)
			}
			if page1 == nil || len(*page1) != 2 {
				count := 0
				if page1 != nil {
					count = len(*page1)
				}
				t.Fatalf("page1: expected 2 items, got %d", count)
			}

			// Page 2: limit 2, offset 2
			page2, err := s.Driver.ListDatatypesPaginated(PaginationParams{Limit: 2, Offset: 2})
			if err != nil {
				t.Fatalf("ListDatatypesPaginated page2: %v", err)
			}
			if page2 == nil || len(*page2) != 2 {
				count := 0
				if page2 != nil {
					count = len(*page2)
				}
				t.Fatalf("page2: expected 2 items, got %d", count)
			}

			// Pages should not overlap
			for _, p1 := range *page1 {
				for _, p2 := range *page2 {
					if p1.DatatypeID == p2.DatatypeID {
						t.Errorf("pages overlap: %v appears on both page 1 and page 2", p1.DatatypeID)
					}
				}
			}

			// Beyond last page
			beyond, err := s.Driver.ListDatatypesPaginated(PaginationParams{Limit: 10, Offset: 100})
			if err != nil {
				t.Fatalf("ListDatatypesPaginated beyond: %v", err)
			}
			if beyond != nil && len(*beyond) != 0 {
				t.Errorf("beyond last page: expected 0 items, got %d", len(*beyond))
			}
		})
	}
}

// TestCrossBackend_StatusUpdate verifies that content_data status transitions
// persist correctly across backends.
func TestCrossBackend_StatusUpdate(t *testing.T) {
	backends := allBackends(t)

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			s := seedCrossBackend(t, b.driver)
			now := types.TimestampNow()

			// Create as draft
			created, err := s.Driver.CreateContentData(s.Ctx, s.AC, CreateContentDataParams{
				RouteID:       s.RouteID,
				ParentID:      nullCID(),
				FirstChildID:  nullCID(),
				NextSiblingID: nullCID(),
				PrevSiblingID: nullCID(),
				DatatypeID:    s.DtPageID,
				AuthorID:      s.User.UserID,
				Status:        types.ContentStatusDraft,
				DateCreated:   now,
				DateModified:  now,
			})
			if err != nil {
				t.Fatalf("Create: %v", err)
			}

			// Verify draft
			got, err := s.Driver.GetContentData(created.ContentDataID)
			if err != nil {
				t.Fatalf("GetContentData(draft): %v", err)
			}
			if got.Status != types.ContentStatusDraft {
				t.Errorf("status = %q, want draft", got.Status)
			}

			// Update to published
			_, err = s.Driver.UpdateContentData(s.Ctx, s.AC, UpdateContentDataParams{
				ContentDataID: created.ContentDataID,
				RouteID:       s.RouteID,
				ParentID:      nullCID(),
				FirstChildID:  nullCID(),
				NextSiblingID: nullCID(),
				PrevSiblingID: nullCID(),
				DatatypeID:    s.DtPageID,
				AuthorID:      s.User.UserID,
				Status:        types.ContentStatusPublished,
				DateCreated:   now,
				DateModified:  types.TimestampNow(),
			})
			if err != nil {
				t.Fatalf("Update to published: %v", err)
			}

			// Verify published
			got, err = s.Driver.GetContentData(created.ContentDataID)
			if err != nil {
				t.Fatalf("GetContentData(published): %v", err)
			}
			if got.Status != types.ContentStatusPublished {
				t.Errorf("status = %q, want published", got.Status)
			}
		})
	}
}
