package plugin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	_ "github.com/mattn/go-sqlite3"
	lua "github.com/yuin/gopher-lua"
)

// -- Test helpers --

// newDBTestState creates a sandboxed Lua state with the full db API registered.
// The state has a 5-second context set. Caller must defer L.Close() and cancel().
// The returned DatabaseAPI can be used to call ResetOpCount() or inspect state.
func newDBTestState(t *testing.T, conn *sql.DB, pluginName string) (*lua.LState, *DatabaseAPI, context.CancelFunc) {
	t.Helper()

	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	ApplySandbox(L, SandboxConfig{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	L.SetContext(ctx)

	api := NewDatabaseAPI(conn, pluginName, db.DialectSQLite, 1000)
	RegisterDBAPI(L, api)
	RegisterLogAPI(L, pluginName)
	FreezeModule(L, "db")
	FreezeModule(L, "log")

	return L, api, cancel
}

// createTestTable creates a table via DDL for testing CRUD operations.
// Creates plugin_<pluginName>_tasks with columns: id, title, status, priority, created_at, updated_at.
func createTestTable(t *testing.T, conn *sql.DB, pluginName string) {
	t.Helper()
	ctx := context.Background()
	fullName := tablePrefix(pluginName) + "tasks"
	err := db.DDLCreateTable(ctx, conn, db.DialectSQLite, db.DDLCreateTableParams{
		Table: fullName,
		Columns: []db.CreateColumnDef{
			{Name: "id", Type: db.ColText, NotNull: true, PrimaryKey: true},
			{Name: "title", Type: db.ColText, NotNull: true},
			{Name: "status", Type: db.ColText, NotNull: true},
			{Name: "priority", Type: db.ColInteger, NotNull: true, Default: "0"},
			{Name: "created_at", Type: db.ColText, NotNull: true},
			{Name: "updated_at", Type: db.ColText, NotNull: true},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
}

// seedTestRow inserts a row directly into the test table for test setup.
func seedTestRow(t *testing.T, conn *sql.DB, pluginName, id, title, status string, priority int) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	fullName := tablePrefix(pluginName) + "tasks"
	_, err := db.QInsert(ctx, conn, db.DialectSQLite, db.InsertParams{
		Table: fullName,
		Values: map[string]any{
			"id":         id,
			"title":      title,
			"status":     status,
			"priority":   priority,
			"created_at": now,
			"updated_at": now,
		},
	})
	if err != nil {
		t.Fatalf("failed to seed test row: %v", err)
	}
}

// countRows returns the number of rows in the specified table.
func countRows(t *testing.T, conn *sql.DB, fullTableName string) int64 {
	t.Helper()
	ctx := context.Background()
	count, err := db.QCount(ctx, conn, db.DialectSQLite, fullTableName, nil)
	if err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	return count
}

// -- Tests: db.query --

func TestDBAPI_Query_ReturnsRows(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "query_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "Task 1", "active", 1)
	seedTestRow(t, conn, pluginName, "id2", "Task 2", "active", 2)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local rows = db.query("tasks", {where = {status = "active"}})
		return #rows
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.query failed: %v", err)
	}
	result := L.Get(-1)
	if n, ok := result.(lua.LNumber); !ok || float64(n) != 2 {
		t.Errorf("expected 2 rows, got %v", result)
	}
}

func TestDBAPI_Query_EmptyWhereReturnsAll(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "query_all_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "Task 1", "active", 1)
	seedTestRow(t, conn, pluginName, "id2", "Task 2", "done", 2)
	seedTestRow(t, conn, pluginName, "id3", "Task 3", "active", 3)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	// Empty opts table (no where) should return all rows.
	code := `
		local rows = db.query("tasks", {})
		return #rows
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.query with empty where failed: %v", err)
	}
	result := L.Get(-1)
	if n, ok := result.(lua.LNumber); !ok || float64(n) != 3 {
		t.Errorf("expected 3 rows, got %v", result)
	}
}

func TestDBAPI_Query_NilWhereReturnsAll(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "query_nil_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "Task 1", "active", 1)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	// No second argument at all.
	code := `
		local rows = db.query("tasks")
		return #rows
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.query without opts failed: %v", err)
	}
	result := L.Get(-1)
	if n, ok := result.(lua.LNumber); !ok || float64(n) != 1 {
		t.Errorf("expected 1 row, got %v", result)
	}
}

func TestDBAPI_Query_EmptyResultReturnsTable(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "query_empty_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local rows = db.query("tasks", {where = {status = "nonexistent"}})
		-- rows should be a table, not nil
		if type(rows) ~= "table" then
			error("expected table, got " .. type(rows))
		end
		return #rows
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.query empty result failed: %v", err)
	}
	result := L.Get(-1)
	if n, ok := result.(lua.LNumber); !ok || float64(n) != 0 {
		t.Errorf("expected 0 rows, got %v", result)
	}
}

func TestDBAPI_Query_WithOrderByAndLimit(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "query_order_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "A", "active", 3)
	seedTestRow(t, conn, pluginName, "id2", "B", "active", 1)
	seedTestRow(t, conn, pluginName, "id3", "C", "active", 2)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local rows = db.query("tasks", {order_by = "title", limit = 2})
		return rows[1].title .. ":" .. rows[2].title
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.query with order_by/limit failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "A:B" {
		t.Errorf("expected 'A:B', got %q", result.String())
	}
}

// -- Tests: db.query_one --

func TestDBAPI_QueryOne_ReturnsSingleRow(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "qone_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "Found", "active", 1)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local row = db.query_one("tasks", {where = {id = "id1"}})
		return row.title
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.query_one failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "Found" {
		t.Errorf("expected 'Found', got %q", result.String())
	}
}

func TestDBAPI_QueryOne_NoMatchReturnsNil(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "qone_nil_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local row = db.query_one("tasks", {where = {id = "nonexistent"}})
		if row == nil then
			return "nil"
		end
		return "not_nil"
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.query_one no match failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "nil" {
		t.Errorf("expected 'nil', got %q", result.String())
	}
}

func TestDBAPI_QueryOne_EmptyWhereReturnsRow(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "qone_empty_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "Task 1", "active", 1)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local row = db.query_one("tasks", {})
		if row == nil then
			return "nil"
		end
		return "found"
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.query_one with empty where failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "found" {
		t.Errorf("expected 'found', got %q", result.String())
	}
}

// -- Tests: db.count --

func TestDBAPI_Count_ReturnsInteger(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "count_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "T1", "active", 1)
	seedTestRow(t, conn, pluginName, "id2", "T2", "done", 2)
	seedTestRow(t, conn, pluginName, "id3", "T3", "active", 3)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local n = db.count("tasks", {where = {status = "active"}})
		return n
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.count failed: %v", err)
	}
	result := L.Get(-1)
	if n, ok := result.(lua.LNumber); !ok || float64(n) != 2 {
		t.Errorf("expected 2, got %v", result)
	}
}

func TestDBAPI_Count_EmptyWhereReturnsTotal(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "count_all_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "T1", "active", 1)
	seedTestRow(t, conn, pluginName, "id2", "T2", "done", 2)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `return db.count("tasks", {})`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.count with empty where failed: %v", err)
	}
	result := L.Get(-1)
	if n, ok := result.(lua.LNumber); !ok || float64(n) != 2 {
		t.Errorf("expected 2, got %v", result)
	}
}

// -- Tests: db.exists --

func TestDBAPI_Exists_ReturnsBoolTrue(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "exists_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "T1", "active", 1)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local e = db.exists("tasks", {where = {id = "id1"}})
		return e
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.exists failed: %v", err)
	}
	result := L.Get(-1)
	if b, ok := result.(lua.LBool); !ok || !bool(b) {
		t.Errorf("expected true, got %v", result)
	}
}

func TestDBAPI_Exists_ReturnsBoolFalse(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "exists_false_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `return db.exists("tasks", {where = {id = "nonexistent"}})`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.exists false failed: %v", err)
	}
	result := L.Get(-1)
	if b, ok := result.(lua.LBool); !ok || bool(b) {
		t.Errorf("expected false, got %v", result)
	}
}

func TestDBAPI_Exists_EmptyWhereChecksAnyRows(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "exists_any_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "T1", "active", 1)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `return db.exists("tasks", {})`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.exists with empty where failed: %v", err)
	}
	result := L.Get(-1)
	if b, ok := result.(lua.LBool); !ok || !bool(b) {
		t.Errorf("expected true for non-empty table, got %v", result)
	}
}

// -- Tests: db.insert --

func TestDBAPI_Insert_AutoSetsIDAndTimestamps(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "insert_auto_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.insert("tasks", {title = "Auto ID", status = "pending", priority = 0})`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.insert failed: %v", err)
	}

	// Verify a row was inserted.
	fullName := tablePrefix(pluginName) + "tasks"
	if countRows(t, conn, fullName) != 1 {
		t.Fatal("expected 1 row after insert")
	}

	// Verify auto-generated fields.
	ctx := context.Background()
	row, err := db.QSelectOne(ctx, conn, db.DialectSQLite, db.SelectParams{Table: fullName})
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if row == nil {
		t.Fatal("expected a row")
	}

	// ID should be a 26-char ULID string.
	idVal, ok := row["id"]
	if !ok {
		t.Fatal("expected 'id' column")
	}
	idStr := fmt.Sprintf("%v", idVal)
	if len(idStr) != 26 {
		t.Errorf("auto-generated id should be 26 chars (ULID), got %d: %q", len(idStr), idStr)
	}

	// created_at and updated_at should be RFC3339 strings.
	for _, col := range []string{"created_at", "updated_at"} {
		val := fmt.Sprintf("%v", row[col])
		_, parseErr := time.Parse(time.RFC3339, val)
		if parseErr != nil {
			t.Errorf("auto-set %s is not valid RFC3339: %q", col, val)
		}
	}
}

func TestDBAPI_Insert_ExplicitIDPreserved(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "insert_explicit_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.insert("tasks", {id = "MY_CUSTOM_ID_00000000000000", title = "Explicit", status = "pending", priority = 0})`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.insert with explicit id failed: %v", err)
	}

	// Verify the explicit ID was used.
	fullName := tablePrefix(pluginName) + "tasks"
	ctx := context.Background()
	row, err := db.QSelectOne(ctx, conn, db.DialectSQLite, db.SelectParams{
		Table: fullName,
		Where: map[string]any{"id": "MY_CUSTOM_ID_00000000000000"},
	})
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if row == nil {
		t.Fatal("expected row with explicit ID")
	}
}

func TestDBAPI_Insert_ExplicitTimestampsPreserved(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "insert_ts_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.insert("tasks", {
		title = "TS",
		status = "done",
		priority = 0,
		created_at = "2020-01-01T00:00:00Z",
		updated_at = "2020-06-15T12:00:00Z",
	})`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.insert with explicit timestamps failed: %v", err)
	}

	fullName := tablePrefix(pluginName) + "tasks"
	ctx := context.Background()
	row, err := db.QSelectOne(ctx, conn, db.DialectSQLite, db.SelectParams{Table: fullName})
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}

	if fmt.Sprintf("%v", row["created_at"]) != "2020-01-01T00:00:00Z" {
		t.Errorf("expected explicit created_at, got %v", row["created_at"])
	}
	if fmt.Sprintf("%v", row["updated_at"]) != "2020-06-15T12:00:00Z" {
		t.Errorf("expected explicit updated_at, got %v", row["updated_at"])
	}
}

// -- Tests: db.update --

func TestDBAPI_Update_AutoSetsUpdatedAt(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "update_auto_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "Original", "active", 1)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.update("tasks", {set = {title = "Updated"}, where = {id = "id1"}})`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.update failed: %v", err)
	}

	// Verify update applied.
	fullName := tablePrefix(pluginName) + "tasks"
	ctx := context.Background()
	row, err := db.QSelectOne(ctx, conn, db.DialectSQLite, db.SelectParams{
		Table: fullName,
		Where: map[string]any{"id": "id1"},
	})
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if fmt.Sprintf("%v", row["title"]) != "Updated" {
		t.Errorf("expected title 'Updated', got %v", row["title"])
	}

	// updated_at should be auto-set to a valid RFC3339 timestamp.
	val := fmt.Sprintf("%v", row["updated_at"])
	_, parseErr := time.Parse(time.RFC3339, val)
	if parseErr != nil {
		t.Errorf("auto-set updated_at is not valid RFC3339: %q", val)
	}
}

func TestDBAPI_Update_ExplicitUpdatedAtPreserved(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "update_ts_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "Original", "active", 1)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.update("tasks", {
		set = {title = "Updated", updated_at = "2025-12-25T00:00:00Z"},
		where = {id = "id1"},
	})`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.update with explicit updated_at failed: %v", err)
	}

	fullName := tablePrefix(pluginName) + "tasks"
	ctx := context.Background()
	row, err := db.QSelectOne(ctx, conn, db.DialectSQLite, db.SelectParams{
		Table: fullName,
		Where: map[string]any{"id": "id1"},
	})
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if fmt.Sprintf("%v", row["updated_at"]) != "2025-12-25T00:00:00Z" {
		t.Errorf("expected explicit updated_at, got %v", row["updated_at"])
	}
}

func TestDBAPI_Update_EmptyWhereRaisesError(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "update_no_where_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.update("tasks", {set = {title = "Bad"}, where = {}})`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for update with empty where, got nil")
	}
	if !strings.Contains(err.Error(), "non-empty where") {
		t.Errorf("expected 'non-empty where' in error, got: %s", err.Error())
	}
}

func TestDBAPI_Update_NoWhereFieldRaisesError(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "update_missing_where_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.update("tasks", {set = {title = "Bad"}})`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for update without where, got nil")
	}
	if !strings.Contains(err.Error(), "non-empty where") {
		t.Errorf("expected 'non-empty where' in error, got: %s", err.Error())
	}
}

// -- Tests: db.delete --

func TestDBAPI_Delete_RemovesRows(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "delete_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "T1", "active", 1)
	seedTestRow(t, conn, pluginName, "id2", "T2", "done", 2)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.delete("tasks", {where = {id = "id1"}})`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.delete failed: %v", err)
	}

	fullName := tablePrefix(pluginName) + "tasks"
	if countRows(t, conn, fullName) != 1 {
		t.Error("expected 1 row remaining after delete")
	}
}

func TestDBAPI_Delete_EmptyWhereRaisesError(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "delete_no_where_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.delete("tasks", {where = {}})`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for delete with empty where, got nil")
	}
	if !strings.Contains(err.Error(), "non-empty where") {
		t.Errorf("expected 'non-empty where' in error, got: %s", err.Error())
	}
}

func TestDBAPI_Delete_NoWhereFieldRaisesError(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "delete_missing_where_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `db.delete("tasks", {})`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for delete without where, got nil")
	}
	if !strings.Contains(err.Error(), "non-empty where") {
		t.Errorf("expected 'non-empty where' in error, got: %s", err.Error())
	}
}

// -- Tests: db.transaction --

func TestDBAPI_Transaction_CommitsOnSuccess(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "tx_commit_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local ok, err = db.transaction(function()
			db.insert("tasks", {title = "TX 1", status = "active", priority = 1})
			db.insert("tasks", {title = "TX 2", status = "active", priority = 2})
		end)
		return tostring(ok)
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.transaction failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "true" {
		t.Errorf("expected 'true', got %q", result.String())
	}

	// Verify both rows were committed.
	fullName := tablePrefix(pluginName) + "tasks"
	if countRows(t, conn, fullName) != 2 {
		t.Errorf("expected 2 rows after successful transaction, got %d", countRows(t, conn, fullName))
	}
}

func TestDBAPI_Transaction_RollsBackOnError(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "tx_rollback_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local ok, err = db.transaction(function()
			db.insert("tasks", {title = "Will rollback", status = "active", priority = 1})
			error("something went wrong")
		end)
		return tostring(ok) .. ":" .. tostring(err)
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.transaction with error failed: %v", err)
	}
	result := L.Get(-1)
	if !strings.HasPrefix(result.String(), "false:") {
		t.Errorf("expected 'false:...' got %q", result.String())
	}

	// Verify no rows were committed (rollback).
	fullName := tablePrefix(pluginName) + "tasks"
	if countRows(t, conn, fullName) != 0 {
		t.Errorf("expected 0 rows after rollback, got %d", countRows(t, conn, fullName))
	}
}

func TestDBAPI_Transaction_NestedRaisesError(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "tx_nested_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local ok, err = db.transaction(function()
			db.transaction(function()
				-- should never reach here
			end)
		end)
		return tostring(ok) .. ":" .. tostring(err)
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.transaction nested failed: %v", err)
	}
	result := L.Get(-1)
	if !strings.HasPrefix(result.String(), "false:") {
		t.Errorf("expected 'false:...' got %q", result.String())
	}
	if !strings.Contains(result.String(), "nested") {
		t.Errorf("expected error message about nested transactions, got %q", result.String())
	}
}

func TestDBAPI_Transaction_InTxResetAfterRollback(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "tx_reset_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	// First transaction fails.
	code1 := `
		local ok, err = db.transaction(function()
			error("fail")
		end)
		return tostring(ok)
	`
	if err := L.DoString(code1); err != nil {
		t.Fatalf("first transaction failed: %v", err)
	}

	// Second transaction should work (inTx was reset).
	code2 := `
		local ok, err = db.transaction(function()
			db.insert("tasks", {title = "After reset", status = "active", priority = 0})
		end)
		return tostring(ok)
	`
	if err := L.DoString(code2); err != nil {
		t.Fatalf("second transaction failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "true" {
		t.Errorf("expected 'true' for second transaction, got %q", result.String())
	}

	fullName := tablePrefix(pluginName) + "tasks"
	if countRows(t, conn, fullName) != 1 {
		t.Errorf("expected 1 row from second transaction")
	}
}

// -- Tests: db.ulid and db.timestamp --

func TestDBAPI_ULID_ReturnsValidString(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, _, cancel := newDBTestState(t, conn, "ulid_test")
	defer L.Close()
	defer cancel()

	code := `return db.ulid()`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.ulid failed: %v", err)
	}
	result := L.Get(-1)
	s := result.String()
	if len(s) != 26 {
		t.Errorf("ULID should be 26 characters, got %d: %q", len(s), s)
	}
}

func TestDBAPI_ULID_GeneratesUniqueValues(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, _, cancel := newDBTestState(t, conn, "ulid_unique_test")
	defer L.Close()
	defer cancel()

	code := `
		local a = db.ulid()
		local b = db.ulid()
		if a == b then
			return "duplicate"
		end
		return "unique"
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.ulid uniqueness test failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "unique" {
		t.Error("expected two ULIDs to be different")
	}
}

func TestDBAPI_Timestamp_ReturnsRFC3339(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, _, cancel := newDBTestState(t, conn, "ts_test")
	defer L.Close()
	defer cancel()

	code := `return db.timestamp()`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.timestamp failed: %v", err)
	}
	result := L.Get(-1)
	s := result.String()
	_, parseErr := time.Parse(time.RFC3339, s)
	if parseErr != nil {
		t.Errorf("db.timestamp did not return valid RFC3339: %q", s)
	}
}

func TestDBAPI_Timestamp_IsUTC(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, _, cancel := newDBTestState(t, conn, "ts_utc_test")
	defer L.Close()
	defer cancel()

	code := `return db.timestamp()`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.timestamp failed: %v", err)
	}
	result := L.Get(-1)
	s := result.String()
	if !strings.HasSuffix(s, "Z") {
		t.Errorf("expected UTC timestamp ending in 'Z', got %q", s)
	}
}

// -- Tests: Table name prefixing --

func TestDBAPI_TablePrefixing(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "prefix_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "T1", "active", 1)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	// Lua passes "tasks" (relative), Go prefixes to "plugin_prefix_test_tasks".
	code := `
		local rows = db.query("tasks", {})
		return #rows
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("query with table prefixing failed: %v", err)
	}
	result := L.Get(-1)
	if n, ok := result.(lua.LNumber); !ok || float64(n) != 1 {
		t.Errorf("expected 1 row, got %v", result)
	}
}

func TestDBAPI_PrefixTable_InvalidTableNameReturnsError(t *testing.T) {
	// Table names with spaces or special characters should be caught.
	_, err := prefixTable("test", "has spaces")
	if err == nil {
		t.Fatal("expected error for table name with spaces")
	}
}

// -- Tests: SQL injection prevention --

func TestDBAPI_InjectionPrevention_TableName(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, _, cancel := newDBTestState(t, conn, "injection_test")
	defer L.Close()
	defer cancel()

	// Attempt SQL injection via table name.
	// prefixTable validates the name, so this should return nil, errmsg (recoverable error).
	code := `
		local rows, err = db.query("tasks; DROP TABLE users --", {})
		if rows == nil and err ~= nil then
			return "blocked:" .. err
		end
		return "not_blocked"
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	result := L.Get(-1)
	if !strings.HasPrefix(result.String(), "blocked:") {
		t.Errorf("expected injection to be blocked, got %q", result.String())
	}
}

func TestDBAPI_InjectionPrevention_WhereKey(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "inj_where_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	// Attempt injection via WHERE column name.
	// The query builder validates column names, returning an error.
	// db.query returns nil, errmsg for query-level errors.
	code := `
		local rows, err = db.query("tasks", {where = {["1=1; --"] = "x"}})
		if rows == nil and err ~= nil then
			return "blocked:" .. err
		end
		return "not_blocked"
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	result := L.Get(-1)
	if !strings.HasPrefix(result.String(), "blocked:") {
		t.Errorf("expected injection in where key to be blocked, got %q", result.String())
	}
}

// -- Tests: Type marshaling round-trip --

func TestDBAPI_TypeMarshalingRoundTrip(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "marshal_test"

	// Create a table with various column types.
	fullName := tablePrefix(pluginName) + "data"
	ctx := context.Background()
	err := db.DDLCreateTable(ctx, conn, db.DialectSQLite, db.DDLCreateTableParams{
		Table: fullName,
		Columns: []db.CreateColumnDef{
			{Name: "id", Type: db.ColText, NotNull: true, PrimaryKey: true},
			{Name: "str_col", Type: db.ColText},
			{Name: "num_col", Type: db.ColReal},
			{Name: "int_col", Type: db.ColInteger},
			{Name: "bool_col", Type: db.ColBoolean},
			{Name: "created_at", Type: db.ColText, NotNull: true},
			{Name: "updated_at", Type: db.ColText, NotNull: true},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("create test table failed: %v", err)
	}

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		db.insert("data", {
			str_col = "hello",
			num_col = 3.14,
			int_col = 42,
			bool_col = 1,
		})
		local row = db.query_one("data", {})
		return row.str_col .. ":" .. tostring(row.num_col) .. ":" .. tostring(row.int_col) .. ":" .. tostring(row.bool_col)
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("type marshaling test failed: %v", err)
	}
	result := L.Get(-1)
	// SQLite stores numbers as int64 or float64; exact format depends on the driver.
	// The key contract is that the data round-trips without error.
	if result == lua.LNil {
		t.Fatal("expected non-nil result")
	}
	// Verify the string part is correct at minimum.
	if !strings.HasPrefix(result.String(), "hello:") {
		t.Errorf("expected result to start with 'hello:', got %q", result.String())
	}
}

// -- Tests: Operation budget --

func TestDBAPI_OpBudget_ExceededRaisesError(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "opbudget_test"
	createTestTable(t, conn, pluginName)

	L, api, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	// Set a very low budget for testing.
	api.maxOpsPerExec = 3

	code := `
		db.count("tasks", {})  -- op 1
		db.count("tasks", {})  -- op 2
		db.count("tasks", {})  -- op 3
		db.count("tasks", {})  -- op 4 -- should exceed
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error when op budget exceeded")
	}
	if !strings.Contains(err.Error(), "exceeded maximum operations") {
		t.Errorf("expected 'exceeded maximum operations' in error, got: %s", err.Error())
	}
}

func TestDBAPI_OpBudget_SentinelError(t *testing.T) {
	api := NewDatabaseAPI(nil, "test", db.DialectSQLite, 2)
	// Exhaust the budget.
	api.opCount = 3

	err := api.checkOpLimit()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrOpLimitExceeded) {
		t.Errorf("expected ErrOpLimitExceeded, got %v", err)
	}
}

func TestDBAPI_OpBudget_ResetsOnResetOpCount(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "opreset_test"
	createTestTable(t, conn, pluginName)

	L, api, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	api.maxOpsPerExec = 2

	// Use 2 ops.
	code := `
		db.count("tasks", {})
		db.count("tasks", {})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("first batch failed: %v", err)
	}

	// Reset the counter (simulates a new checkout).
	api.ResetOpCount()

	// Should be able to run 2 more ops.
	code = `
		db.count("tasks", {})
		db.count("tasks", {})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("second batch after reset failed: %v", err)
	}
}

// -- Tests: Error convention --

func TestDBAPI_ErrorConvention_RecoverableReturnsNilErrmsg(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "err_recover_test"
	// Do NOT create the table -- query will fail with a recoverable error.

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	// Query against a nonexistent table should return nil, errmsg.
	code := `
		local rows, err = db.query("nonexistent_table", {})
		if rows == nil and err ~= nil then
			return "recoverable:" .. err
		end
		return "unexpected"
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("recoverable error test failed: %v", err)
	}
	result := L.Get(-1)
	if !strings.HasPrefix(result.String(), "recoverable:") {
		t.Errorf("expected 'recoverable:...', got %q", result.String())
	}
}

func TestDBAPI_ErrorConvention_ProgrammingErrorRaises(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "err_raise_test"

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	// Calling insert without arguments (wrong type) should raise error().
	code := `db.insert()`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected raised error for missing arguments")
	}
}

func TestDBAPI_ErrorConvention_ConstraintViolation(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "err_constraint_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	// Insert two rows with the same ID (PK violation).
	code := `
		db.insert("tasks", {id = "DUPE_ID_0000000000000000000", title = "First", status = "active", priority = 0})
		local nil_val, err = db.insert("tasks", {id = "DUPE_ID_0000000000000000000", title = "Dupe", status = "active", priority = 0})
		if nil_val == nil and err ~= nil then
			return "constraint:" .. err
		end
		return "no_error"
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("constraint test failed: %v", err)
	}
	result := L.Get(-1)
	if !strings.HasPrefix(result.String(), "constraint:") {
		t.Errorf("expected 'constraint:...', got %q", result.String())
	}
}

// -- Tests: NewDatabaseAPI defaults --

func TestNewDatabaseAPI_Defaults(t *testing.T) {
	api := NewDatabaseAPI(nil, "test_plugin", db.DialectSQLite, 0)

	if api.maxRows != 100 {
		t.Errorf("expected maxRows=100, got %d", api.maxRows)
	}
	if api.maxTxOps != 10 {
		t.Errorf("expected maxTxOps=10, got %d", api.maxTxOps)
	}
	if api.maxOpsPerExec != 1000 {
		t.Errorf("expected maxOpsPerExec=1000 (default for 0 input), got %d", api.maxOpsPerExec)
	}
	if api.inTx {
		t.Error("expected inTx=false")
	}
	if api.opCount != 0 {
		t.Errorf("expected opCount=0, got %d", api.opCount)
	}
}

func TestNewDatabaseAPI_CustomMaxOps(t *testing.T) {
	api := NewDatabaseAPI(nil, "test_plugin", db.DialectSQLite, 500)
	if api.maxOpsPerExec != 500 {
		t.Errorf("expected maxOpsPerExec=500, got %d", api.maxOpsPerExec)
	}
}

// -- Tests: prefixTable helper --

func TestPrefixTable_Valid(t *testing.T) {
	result, err := prefixTable("my_plugin", "tasks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "plugin_my_plugin_tasks" {
		t.Errorf("expected 'plugin_my_plugin_tasks', got %q", result)
	}
}

func TestPrefixTable_InvalidCharacters(t *testing.T) {
	_, err := prefixTable("my_plugin", "bad-name")
	if err == nil {
		t.Fatal("expected error for table name with hyphens")
	}
}

// -- Tests: Full CRUD cycle through Lua --

func TestDBAPI_FullCRUDCycle(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "crud_test"
	createTestTable(t, conn, pluginName)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		-- INSERT
		db.insert("tasks", {title = "CRUD Task", status = "active", priority = 5})

		-- QUERY
		local rows = db.query("tasks", {where = {status = "active"}})
		if #rows ~= 1 then error("expected 1 row, got " .. #rows) end
		local id = rows[1].id

		-- QUERY_ONE
		local row = db.query_one("tasks", {where = {id = id}})
		if row.title ~= "CRUD Task" then error("wrong title") end

		-- COUNT
		local n = db.count("tasks", {})
		if n ~= 1 then error("expected count 1") end

		-- EXISTS
		if not db.exists("tasks", {where = {id = id}}) then error("expected exists") end

		-- UPDATE
		db.update("tasks", {set = {status = "done"}, where = {id = id}})
		local updated = db.query_one("tasks", {where = {id = id}})
		if updated.status ~= "done" then error("update failed") end

		-- DELETE
		db.delete("tasks", {where = {id = id}})
		local deleted = db.query_one("tasks", {where = {id = id}})
		if deleted ~= nil then error("delete failed") end

		return "ok"
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("full CRUD cycle failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "ok" {
		t.Errorf("expected 'ok', got %q", result.String())
	}
}

// -- Tests: define_table via db.define_table --

func TestDBAPI_DefineTable_Accessible(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "deftbl_test"

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {
				{name = "name", type = "text", not_null = true},
			},
		})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("db.define_table via db API failed: %v", err)
	}

	if !tableExists(t, conn, "plugin_deftbl_test_items") {
		t.Fatal("expected table to be created via db.define_table")
	}
}

// -- Tests: Transaction with query operations --

func TestDBAPI_Transaction_QueryInsideWorks(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()
	pluginName := "tx_query_test"
	createTestTable(t, conn, pluginName)
	seedTestRow(t, conn, pluginName, "id1", "Existing", "active", 1)

	L, _, cancel := newDBTestState(t, conn, pluginName)
	defer L.Close()
	defer cancel()

	code := `
		local ok, err = db.transaction(function()
			local n = db.count("tasks", {})
			if n ~= 1 then error("expected 1 row, got " .. n) end
			db.insert("tasks", {title = "New in TX", status = "active", priority = 2})
			local n2 = db.count("tasks", {})
			if n2 ~= 2 then error("expected 2 rows, got " .. n2) end
		end)
		if not ok then error("tx failed: " .. tostring(err)) end
		return tostring(db.count("tasks", {}))
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("transaction with queries failed: %v", err)
	}
	result := L.Get(-1)
	if result.String() != "2" {
		t.Errorf("expected '2' rows after transaction, got %q", result.String())
	}
}

// -- Tests: ResetOpCount --

func TestDBAPI_ResetOpCount(t *testing.T) {
	api := NewDatabaseAPI(nil, "test", db.DialectSQLite, 1000)
	api.opCount = 500
	api.ResetOpCount()
	if api.opCount != 0 {
		t.Errorf("expected opCount=0 after reset, got %d", api.opCount)
	}
}

// -- Tests: checkOpLimit --

func TestDBAPI_CheckOpLimit_IncrementsCounter(t *testing.T) {
	api := NewDatabaseAPI(nil, "test", db.DialectSQLite, 1000)

	for range 5 {
		err := api.checkOpLimit()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if api.opCount != 5 {
		t.Errorf("expected opCount=5, got %d", api.opCount)
	}
}

func TestDBAPI_CheckOpLimit_ExceedsBudget(t *testing.T) {
	api := NewDatabaseAPI(nil, "test", db.DialectSQLite, 3)

	for range 3 {
		if err := api.checkOpLimit(); err != nil {
			t.Fatalf("unexpected error within budget: %v", err)
		}
	}

	// 4th call should exceed.
	err := api.checkOpLimit()
	if err == nil {
		t.Fatal("expected error when budget exceeded")
	}
	if !errors.Is(err, ErrOpLimitExceeded) {
		t.Errorf("expected ErrOpLimitExceeded, got %v", err)
	}
}

// -- Tests: Nil where conversions --

func TestDBAPI_ParseWhereFromLua_NilWhereField(t *testing.T) {
	L := newTestState()
	defer L.Close()
	ApplySandbox(L, SandboxConfig{})

	tbl := L.NewTable()
	// No "where" field set.
	result := parseWhereFromLua(L, tbl)
	if result != nil {
		t.Errorf("expected nil for missing where field, got %v", result)
	}
}

func TestDBAPI_ParseWhereFromLua_EmptyWhereTable(t *testing.T) {
	L := newTestState()
	defer L.Close()
	ApplySandbox(L, SandboxConfig{})

	tbl := L.NewTable()
	L.SetField(tbl, "where", L.NewTable())
	result := parseWhereFromLua(L, tbl)
	// Empty table produces empty map.
	if result == nil {
		t.Fatal("expected empty map for empty where table, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

// -- Tests: luaStringField --

func TestDBAPI_LuaStringField(t *testing.T) {
	L := newTestState()
	defer L.Close()
	ApplySandbox(L, SandboxConfig{})

	tbl := L.NewTable()
	L.SetField(tbl, "order_by", lua.LString("title"))
	L.SetField(tbl, "number_val", lua.LNumber(42))

	if got := luaStringField(L, tbl, "order_by"); got != "title" {
		t.Errorf("expected 'title', got %q", got)
	}
	if got := luaStringField(L, tbl, "number_val"); got != "" {
		t.Errorf("expected empty string for number field, got %q", got)
	}
	if got := luaStringField(L, tbl, "missing"); got != "" {
		t.Errorf("expected empty string for missing field, got %q", got)
	}
}
