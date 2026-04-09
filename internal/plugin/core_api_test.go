package plugin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	_ "github.com/mattn/go-sqlite3"
	lua "github.com/yuin/gopher-lua"
)

// -- Test helpers --

// newCoreTestState creates a sandboxed Lua state with both db and core APIs registered.
// Returns the LState, DatabaseAPI (for op count reset / before-hook testing),
// CoreTableAPI, and a cancel function. Caller must defer L.Close() and cancel().
func newCoreTestState(t *testing.T, conn *sql.DB, pluginName string, access PluginCoreAccess) (*lua.LState, *DatabaseAPI, *CoreTableAPI, context.CancelFunc) {
	t.Helper()

	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	ApplySandbox(L, SandboxConfig{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	L.SetContext(ctx)

	dbAPI := NewDatabaseAPI(conn, pluginName, db.DialectSQLite, 1000, nil)
	RegisterDBAPI(L, dbAPI)
	RegisterLogAPI(L, pluginName)

	coreAPI := NewCoreTableAPI(dbAPI, pluginName, db.DialectSQLite, access)
	RegisterCoreAPI(L, coreAPI)

	FreezeModule(L, "db")
	FreezeModule(L, "log")
	FreezeModule(L, "core")

	return L, dbAPI, coreAPI, cancel
}

// createCoreTestTable creates a core-style table (no plugin prefix) for testing.
func createCoreTestTable(t *testing.T, conn *sql.DB, tableName string, columns []db.CreateColumnDef) {
	t.Helper()
	ctx := context.Background()
	err := db.DDLCreateTable(ctx, conn, db.DialectSQLite, db.DDLCreateTableParams{
		Table:       tableName,
		Columns:     columns,
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("failed to create core test table %q: %v", tableName, err)
	}
}

// seedCoreRow inserts a row directly into a core table.
func seedCoreRow(t *testing.T, conn *sql.DB, tableName string, values map[string]any) {
	t.Helper()
	ctx := context.Background()
	_, err := db.QInsert(ctx, conn, db.DialectSQLite, db.InsertParams{
		Table:  tableName,
		Values: values,
	})
	if err != nil {
		t.Fatalf("failed to seed core row in %q: %v", tableName, err)
	}
}

// setupContentFieldsTable creates a content_fields table and seeds rows.
func setupContentFieldsTable(t *testing.T, conn *sql.DB) {
	t.Helper()
	createCoreTestTable(t, conn, "content_fields", []db.CreateColumnDef{
		{Name: "id", Type: db.ColText, NotNull: true, PrimaryKey: true},
		{Name: "content_id", Type: db.ColText, NotNull: true},
		{Name: "field_name", Type: db.ColText, NotNull: true},
		{Name: "field_value", Type: db.ColText},
		{Name: "created_at", Type: db.ColText, NotNull: true},
		{Name: "updated_at", Type: db.ColText, NotNull: true},
	})

	now := time.Now().UTC().Format(time.RFC3339)
	seedCoreRow(t, conn, "content_fields", map[string]any{
		"id": "cf_01", "content_id": "c_01", "field_name": "title",
		"field_value": "Hello World", "created_at": now, "updated_at": now,
	})
	seedCoreRow(t, conn, "content_fields", map[string]any{
		"id": "cf_02", "content_id": "c_01", "field_name": "body",
		"field_value": "Lorem ipsum", "created_at": now, "updated_at": now,
	})
}

// setupUsersTable creates a users table with a hash column for blocked-column testing.
func setupUsersTable(t *testing.T, conn *sql.DB) {
	t.Helper()
	createCoreTestTable(t, conn, "users", []db.CreateColumnDef{
		{Name: "id", Type: db.ColText, NotNull: true, PrimaryKey: true},
		{Name: "username", Type: db.ColText, NotNull: true},
		{Name: "email", Type: db.ColText, NotNull: true},
		{Name: "hash", Type: db.ColText, NotNull: true},
		{Name: "created_at", Type: db.ColText, NotNull: true},
		{Name: "updated_at", Type: db.ColText, NotNull: true},
	})

	now := time.Now().UTC().Format(time.RFC3339)
	seedCoreRow(t, conn, "users", map[string]any{
		"id": "u_01", "username": "alice", "email": "alice@example.com",
		"hash": "secret_hash_value", "created_at": now, "updated_at": now,
	})
}

// setupContentDataTable creates a content_data table for write testing.
func setupContentDataTable(t *testing.T, conn *sql.DB) {
	t.Helper()
	createCoreTestTable(t, conn, "content_data", []db.CreateColumnDef{
		{Name: "id", Type: db.ColText, NotNull: true, PrimaryKey: true},
		{Name: "datatype_id", Type: db.ColText, NotNull: true},
		{Name: "title", Type: db.ColText, NotNull: true},
		{Name: "status", Type: db.ColText, NotNull: true},
		{Name: "created_at", Type: db.ColText, NotNull: true},
		{Name: "updated_at", Type: db.ColText, NotNull: true},
	})
}

// -- Tests: Permission checks --

func TestCoreAPI_WhitelistDenial(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"tokens": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	// "tokens" is not in the whitelist
	err := L.DoString(`local rows, errmsg = core.query("tokens", {}); if errmsg then error(errmsg) end`)
	if err == nil {
		t.Fatal("expected error for non-whitelisted table, got nil")
	}
	if !strings.Contains(err.Error(), "not accessible to plugins") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_PolicyDenial_WriteOnReadOnly(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"datatypes": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	// datatypes is read-only in the policy -- even if approved_access says write
	err := L.DoString(`
		local _, errmsg = core.insert("datatypes", {id = "dt_01", name = "test"})
		if errmsg then error(errmsg) end
	`)
	if err == nil {
		t.Fatal("expected error for write on read-only table, got nil")
	}
	if !strings.Contains(err.Error(), "does not allow write access") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_ApprovedAccessDenial(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	// Plugin has access to content_fields but NOT content_data
	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	err := L.DoString(`local rows, errmsg = core.query("content_data", {}); if errmsg then error(errmsg) end`)
	if err == nil {
		t.Fatal("expected error for unapproved table, got nil")
	}
	if !strings.Contains(err.Error(), "has no approved access") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_NilApprovedAccess(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	// nil approved access -- no core access at all
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", nil)
	defer L.Close()
	defer cancel()

	err := L.DoString(`local rows, errmsg = core.query("content_fields", {}); if errmsg then error(errmsg) end`)
	if err == nil {
		t.Fatal("expected error for nil approved access, got nil")
	}
	if !strings.Contains(err.Error(), "has no approved access") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_ApprovedAccessReadOnly_DeniesWrite(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	// Plugin has read but not write on content_data
	access := PluginCoreAccess{"content_data": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	err := L.DoString(`
		local _, errmsg = core.insert("content_data", {id = "c_01", datatype_id = "dt_01", title = "Test", status = "draft", created_at = "2024-01-01T00:00:00Z", updated_at = "2024-01-01T00:00:00Z"})
		if errmsg then error(errmsg) end
	`)
	if err == nil {
		t.Fatal("expected error for write without write approval, got nil")
	}
	if !strings.Contains(err.Error(), "not approved for") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_SuccessfulAccess(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	err := L.DoString(`
		local rows = core.query("content_fields", {})
		if #rows ~= 2 then
			error("expected 2 rows, got " .. #rows)
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

// -- Tests: Pre-query column validation --

func TestCoreAPI_BlockedColumnInExplicitRequest(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"users": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupUsersTable(t, conn)

	// Explicitly requesting "hash" should be rejected before query
	err := L.DoString(`
		local _, errmsg = core.query("users", {columns = {"id", "hash"}})
		if errmsg then error(errmsg) end
	`)
	if err == nil {
		t.Fatal("expected error for blocked column in explicit request, got nil")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_BlockedColumnInQueryOne(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"users": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupUsersTable(t, conn)

	err := L.DoString(`
		local _, errmsg = core.query_one("users", {columns = {"hash"}, where = {id = "u_01"}})
		if errmsg then error(errmsg) end
	`)
	if err == nil {
		t.Fatal("expected error for blocked column in query_one, got nil")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

// -- Tests: Post-query column stripping --

func TestCoreAPI_BlockedColumnStrippedFromSelectStar(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"users": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupUsersTable(t, conn)

	// SELECT * (no explicit columns) should strip "hash"
	err := L.DoString(`
		local rows = core.query("users", {})
		if #rows ~= 1 then
			error("expected 1 row, got " .. #rows)
		end
		if rows[1].hash ~= nil then
			error("hash column should have been stripped, got: " .. tostring(rows[1].hash))
		end
		if rows[1].username ~= "alice" then
			error("expected username 'alice', got: " .. tostring(rows[1].username))
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_BlockedColumnStrippedFromQueryOne(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"users": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupUsersTable(t, conn)

	err := L.DoString(`
		local row = core.query_one("users", {where = {id = "u_01"}})
		if row == nil then
			error("expected a row, got nil")
		end
		if row.hash ~= nil then
			error("hash column should have been stripped")
		end
		if row.email ~= "alice@example.com" then
			error("expected email 'alice@example.com', got: " .. tostring(row.email))
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_NonBlockedExplicitColumns(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"users": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupUsersTable(t, conn)

	// Requesting only non-blocked columns should work fine
	err := L.DoString(`
		local rows = core.query("users", {columns = {"id", "username", "email"}})
		if #rows ~= 1 then
			error("expected 1 row, got " .. #rows)
		end
		if rows[1].username ~= "alice" then
			error("expected username 'alice'")
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

// -- Tests: Read operations --

func TestCoreAPI_QueryOne(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	err := L.DoString(`
		local row = core.query_one("content_fields", {where = {id = "cf_01"}})
		if row == nil then
			error("expected a row, got nil")
		end
		if row.field_name ~= "title" then
			error("expected field_name 'title', got: " .. tostring(row.field_name))
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_QueryOne_NoMatch(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	err := L.DoString(`
		local row = core.query_one("content_fields", {where = {id = "nonexistent"}})
		if row ~= nil then
			error("expected nil for no match")
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_Count(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	err := L.DoString(`
		local n = core.count("content_fields", {})
		if n ~= 2 then
			error("expected count 2, got " .. tostring(n))
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_Count_WithWhere(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	err := L.DoString(`
		local n = core.count("content_fields", {where = {field_name = "title"}})
		if n ~= 1 then
			error("expected count 1, got " .. tostring(n))
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_Exists(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	err := L.DoString(`
		local yes = core.exists("content_fields", {where = {id = "cf_01"}})
		if not yes then
			error("expected exists to return true")
		end

		local no = core.exists("content_fields", {where = {id = "nonexistent"}})
		if no then
			error("expected exists to return false for nonexistent")
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

// -- Tests: Write operations --

func TestCoreAPI_Insert(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	err := L.DoString(`
		core.insert("content_data", {
			id = "c_99",
			datatype_id = "dt_01",
			title = "Inserted via core",
			status = "draft",
			created_at = "2024-01-01T00:00:00Z",
			updated_at = "2024-01-01T00:00:00Z",
		})
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	// Verify the row was inserted.
	count := countRows(t, conn, "content_data")
	if count != 1 {
		t.Errorf("expected 1 row in content_data, got %d", count)
	}
}

func TestCoreAPI_Update(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)
	now := time.Now().UTC().Format(time.RFC3339)
	seedCoreRow(t, conn, "content_data", map[string]any{
		"id": "c_01", "datatype_id": "dt_01", "title": "Original",
		"status": "draft", "created_at": now, "updated_at": now,
	})

	err := L.DoString(`
		core.update("content_data", {
			set = {title = "Updated via core"},
			where = {id = "c_01"},
		})
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	// Verify the update.
	ctx := context.Background()
	row, qErr := db.QSelectOne(ctx, conn, db.DialectSQLite, db.SelectParams{
		Table:   "content_data",
		Columns: []string{"title"},
		Where:   map[string]any{"id": "c_01"},
	})
	if qErr != nil {
		t.Fatalf("failed to query updated row: %v", qErr)
	}
	if row == nil {
		t.Fatal("expected row, got nil")
	}
	if row["title"] != "Updated via core" {
		t.Errorf("expected title 'Updated via core', got %v", row["title"])
	}
}

func TestCoreAPI_Delete(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)
	now := time.Now().UTC().Format(time.RFC3339)
	seedCoreRow(t, conn, "content_data", map[string]any{
		"id": "c_01", "datatype_id": "dt_01", "title": "to delete",
		"status": "draft", "created_at": now, "updated_at": now,
	})

	err := L.DoString(`
		core.delete("content_data", {
			where = {id = "c_01"},
		})
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	count := countRows(t, conn, "content_data")
	if count != 0 {
		t.Errorf("expected 0 rows after delete, got %d", count)
	}
}

func TestCoreAPI_WriteDeniedOnReadOnlyTable(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	// routes is read-only in the policy
	access := PluginCoreAccess{"routes": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	err := L.DoString(`
		local _, errmsg = core.insert("routes", {id = "r_01", path = "/test"})
		if errmsg then error(errmsg) end
	`)
	if err == nil {
		t.Fatal("expected error for write on read-only routes table, got nil")
	}
	if !strings.Contains(err.Error(), "does not allow write access") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

// -- Tests: Write safety limit --

func TestCoreAPI_UpdateSafetyLimit(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, coreAPI, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	// Seed more rows than the safety limit (set it low for testing).
	coreAPI.maxWriteRows = 5
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range 10 {
		seedCoreRow(t, conn, "content_data", map[string]any{
			"id": fmt.Sprintf("c_%02d", i), "datatype_id": "dt_01",
			"title": "Bulk", "status": "draft", "created_at": now, "updated_at": now,
		})
	}

	// Update that would affect all 10 rows should be rejected.
	err := L.DoString(`
		local _, errmsg = core.update("content_data", {
			set = {status = "published"},
			where = {status = "draft"},
		})
		if errmsg then error(errmsg) end
	`)
	if err == nil {
		t.Fatal("expected error for update exceeding safety limit, got nil")
	}
	if !strings.Contains(err.Error(), "exceeding safety limit") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_DeleteSafetyLimit(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, coreAPI, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)
	coreAPI.maxWriteRows = 3
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range 5 {
		seedCoreRow(t, conn, "content_data", map[string]any{
			"id": fmt.Sprintf("c_%02d", i), "datatype_id": "dt_01",
			"title": "Bulk", "status": "draft", "created_at": now, "updated_at": now,
		})
	}

	err := L.DoString(`
		local _, errmsg = core.delete("content_data", {
			where = {status = "draft"},
		})
		if errmsg then error(errmsg) end
	`)
	if err == nil {
		t.Fatal("expected error for delete exceeding safety limit, got nil")
	}
	if !strings.Contains(err.Error(), "exceeding safety limit") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_UpdateUnderSafetyLimit(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)
	now := time.Now().UTC().Format(time.RFC3339)
	seedCoreRow(t, conn, "content_data", map[string]any{
		"id": "c_01", "datatype_id": "dt_01", "title": "One",
		"status": "draft", "created_at": now, "updated_at": now,
	})

	// 1 row is under the default limit of 100
	err := L.DoString(`
		core.update("content_data", {
			set = {status = "published"},
			where = {id = "c_01"},
		})
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

// -- Tests: No auto-set --

func TestCoreAPI_InsertNoAutoSetID(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	// Insert with explicit ID -- should use that exact ID, not auto-generate
	err := L.DoString(`
		core.insert("content_data", {
			id = "MY_EXPLICIT_ID",
			datatype_id = "dt_01",
			title = "Explicit",
			status = "draft",
			created_at = "2024-06-01T00:00:00Z",
			updated_at = "2024-06-01T00:00:00Z",
		})
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	ctx := context.Background()
	row, qErr := db.QSelectOne(ctx, conn, db.DialectSQLite, db.SelectParams{
		Table: "content_data",
		Where: map[string]any{"id": "MY_EXPLICIT_ID"},
	})
	if qErr != nil {
		t.Fatalf("failed to query: %v", qErr)
	}
	if row == nil {
		t.Fatal("expected row with explicit ID, got nil")
	}
}

func TestCoreAPI_InsertNoAutoSetTimestamps(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	// Insert with specific timestamps -- they should be preserved exactly
	err := L.DoString(`
		core.insert("content_data", {
			id = "c_ts",
			datatype_id = "dt_01",
			title = "Timestamp test",
			status = "draft",
			created_at = "2020-01-01T00:00:00Z",
			updated_at = "2020-01-01T00:00:00Z",
		})
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	ctx := context.Background()
	row, qErr := db.QSelectOne(ctx, conn, db.DialectSQLite, db.SelectParams{
		Table:   "content_data",
		Columns: []string{"created_at", "updated_at"},
		Where:   map[string]any{"id": "c_ts"},
	})
	if qErr != nil {
		t.Fatalf("failed to query: %v", qErr)
	}
	if row == nil {
		t.Fatal("expected row, got nil")
	}
	if row["created_at"] != "2020-01-01T00:00:00Z" {
		t.Errorf("expected created_at '2020-01-01T00:00:00Z', got %v", row["created_at"])
	}
	if row["updated_at"] != "2020-01-01T00:00:00Z" {
		t.Errorf("expected updated_at '2020-01-01T00:00:00Z', got %v", row["updated_at"])
	}
}

// -- Tests: Op budget sharing --

func TestCoreAPI_SharedOpBudget(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, dbAPI, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	// Set a very low op budget.
	dbAPI.maxOpsPerExec = 3
	dbAPI.ResetOpCount()

	// Three operations should succeed.
	err := L.DoString(`
		core.query("content_fields", {})
		core.count("content_fields", {})
		core.exists("content_fields", {})
	`)
	if err != nil {
		t.Fatalf("unexpected error for first 3 ops: %s", err.Error())
	}

	// Fourth should fail (budget exhausted).
	err = L.DoString(`core.query("content_fields", {})`)
	if err == nil {
		t.Fatal("expected error for exceeding op budget, got nil")
	}
	if !strings.Contains(err.Error(), "operation limit exceeded") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_SharedOpBudget_CrossModule(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, dbAPI, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	// Also create a plugin table for db.* calls.
	createTestTable(t, conn, "test_plugin")

	dbAPI.maxOpsPerExec = 2
	dbAPI.ResetOpCount()

	// One db.* call + one core.* call = 2 ops (budget).
	err := L.DoString(`
		db.query("tasks", {})
		core.query("content_fields", {})
	`)
	if err != nil {
		t.Fatalf("unexpected error for 2 cross-module ops: %s", err.Error())
	}

	// Third call should fail regardless of module.
	err = L.DoString(`core.exists("content_fields", {})`)
	if err == nil {
		t.Fatal("expected error for exceeding shared op budget, got nil")
	}
}

// -- Tests: Before-hook blocking --

func TestCoreAPI_BeforeHookBlocking(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, dbAPI, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	// Simulate being inside a before-hook.
	dbAPI.inBeforeHook = true

	err := L.DoString(`core.query("content_fields", {})`)
	if err == nil {
		t.Fatal("expected error when inside before-hook, got nil")
	}
	if !strings.Contains(err.Error(), "not allowed inside before-hooks") {
		t.Errorf("unexpected error: %s", err.Error())
	}

	// Restore and verify it works again.
	dbAPI.inBeforeHook = false
	err = L.DoString(`
		local rows = core.query("content_fields", {})
		if #rows ~= 2 then error("expected 2 rows") end
	`)
	if err != nil {
		t.Fatalf("unexpected error after clearing before-hook flag: %s", err.Error())
	}
}

// -- Tests: Transaction integration --

func TestCoreAPI_TransactionIntegration(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	// Insert via core.* inside a db.transaction -- should use the same tx.
	err := L.DoString(`
		local ok, errmsg = db.transaction(function()
			core.insert("content_data", {
				id = "c_tx",
				datatype_id = "dt_01",
				title = "In transaction",
				status = "draft",
				created_at = "2024-01-01T00:00:00Z",
				updated_at = "2024-01-01T00:00:00Z",
			})
		end)
		if not ok then error(errmsg) end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	// Row should be committed.
	count := countRows(t, conn, "content_data")
	if count != 1 {
		t.Errorf("expected 1 row after committed tx, got %d", count)
	}
}

func TestCoreAPI_TransactionRollback(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	// Insert inside a transaction that errors out -- should be rolled back.
	err := L.DoString(`
		local ok, errmsg = db.transaction(function()
			core.insert("content_data", {
				id = "c_rollback",
				datatype_id = "dt_01",
				title = "Will be rolled back",
				status = "draft",
				created_at = "2024-01-01T00:00:00Z",
				updated_at = "2024-01-01T00:00:00Z",
			})
			error("force rollback")
		end)
		-- ok should be false, errmsg should contain the error
		if ok then error("expected transaction to fail") end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	// Row should NOT be there (rolled back).
	count := countRows(t, conn, "content_data")
	if count != 0 {
		t.Errorf("expected 0 rows after rollback, got %d", count)
	}
}

// -- Tests: validateVM core module check --

func TestCoreAPI_ValidateVM_IntactModule(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	// Also need http and hooks for validateVM to pass.
	RegisterHTTPAPI(L, "test_plugin")
	RegisterHooksAPI(L, "test_plugin")
	FreezeModule(L, "http")
	FreezeModule(L, "hooks")

	pool := &VMPool{pluginName: "test_plugin"}
	if !pool.validateVM(L) {
		t.Error("validateVM should return true for intact core module")
	}
}

func TestCoreAPI_ValidateVM_CorruptedModule(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	RegisterHTTPAPI(L, "test_plugin")
	RegisterHooksAPI(L, "test_plugin")
	FreezeModule(L, "http")
	FreezeModule(L, "hooks")

	// Corrupt the core module by replacing it with a plain table.
	corruptCore := L.NewTable()
	corruptCore.RawSetString("query", lua.LString("not a function"))
	L.SetGlobal("core", corruptCore)

	pool := &VMPool{pluginName: "test_plugin"}
	if pool.validateVM(L) {
		t.Error("validateVM should return false for corrupted core module")
	}
}

func TestCoreAPI_ValidateVM_MissingModule(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, _, cancel := newDBTestState(t, conn, "test_plugin")
	defer L.Close()
	defer cancel()

	RegisterHTTPAPI(L, "test_plugin")
	RegisterHooksAPI(L, "test_plugin")
	FreezeModule(L, "http")
	FreezeModule(L, "hooks")

	// No core module registered at all.
	pool := &VMPool{pluginName: "test_plugin"}
	if pool.validateVM(L) {
		t.Error("validateVM should return false when core module is missing")
	}
}

// -- Tests: Update and Delete require where --

func TestCoreAPI_UpdateRequiresWhere(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	err := L.DoString(`
		core.update("content_data", {set = {title = "new"}})
	`)
	if err == nil {
		t.Fatal("expected error for update without where, got nil")
	}
	if !strings.Contains(err.Error(), "non-empty where") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCoreAPI_DeleteRequiresWhere(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_data": {"read", "write"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentDataTable(t, conn)

	err := L.DoString(`
		core.delete("content_data", {})
	`)
	if err == nil {
		t.Fatal("expected error for delete without where, got nil")
	}
	if !strings.Contains(err.Error(), "non-empty where") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

// -- Tests: Query with order_by and limit --

func TestCoreAPI_QueryWithOrderAndLimit(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	access := PluginCoreAccess{"content_fields": {"read"}}
	L, _, _, cancel := newCoreTestState(t, conn, "test_plugin", access)
	defer L.Close()
	defer cancel()

	setupContentFieldsTable(t, conn)

	err := L.DoString(`
		local rows = core.query("content_fields", {
			order_by = "field_name",
			desc = true,
			limit = 1,
		})
		if #rows ~= 1 then
			error("expected 1 row with limit, got " .. #rows)
		end
		if rows[1].field_name ~= "title" then
			error("expected 'title' (desc order), got " .. tostring(rows[1].field_name))
		end
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}
