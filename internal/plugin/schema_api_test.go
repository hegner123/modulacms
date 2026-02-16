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

// openTestDB creates an in-memory SQLite database for testing.
// Caller must defer conn.Close().
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	return conn
}

// newSchemaTestState creates a sandboxed Lua state with db.define_table registered.
// The state has a context with a 5-second timeout set.
// Caller must defer L.Close() and cancel().
func newSchemaTestState(t *testing.T, conn *sql.DB, pluginName string) (*lua.LState, context.CancelFunc) {
	t.Helper()

	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	ApplySandbox(L, SandboxConfig{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	L.SetContext(ctx)

	// Register db.define_table as a global.
	dbTable := L.NewTable()
	dbTable.RawSetString("define_table", L.NewFunction(
		luaDefineTable(L, pluginName, conn, db.DialectSQLite, nil),
	))
	L.SetGlobal("db", dbTable)

	return L, cancel
}

// tableExists checks if a table with the given name exists in the SQLite database.
func tableExists(t *testing.T, conn *sql.DB, tableName string) bool {
	t.Helper()
	var count int
	err := conn.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
		tableName,
	).Scan(&count)
	if err != nil {
		t.Fatalf("failed to check table existence: %v", err)
	}
	return count > 0
}

// indexExists checks if an index with the given name exists in the SQLite database.
func indexExists(t *testing.T, conn *sql.DB, indexName string) bool {
	t.Helper()
	var count int
	err := conn.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?",
		indexName,
	).Scan(&count)
	if err != nil {
		t.Fatalf("failed to check index existence: %v", err)
	}
	return count > 0
}

// getColumnNames returns the column names for a table in the SQLite database.
func getColumnNames(t *testing.T, conn *sql.DB, tableName string) []string {
	t.Helper()
	rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%q)", tableName))
	if err != nil {
		t.Fatalf("failed to get table info: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("failed to scan column info: %v", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows iteration error: %v", err)
	}
	return names
}

func TestDefineTable_ValidCreation(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "test_plugin")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("tasks", {
			columns = {
				{name = "title",    type = "text",    not_null = true},
				{name = "status",   type = "text",    not_null = true, default = "pending"},
				{name = "priority", type = "integer", not_null = true, default = 0},
			},
			indexes = {
				{columns = {"status"}},
				{columns = {"status", "priority"}},
			},
		})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("define_table failed: %v", err)
	}

	// Verify table was created with the correct prefix.
	fullName := "plugin_test_plugin_tasks"
	if !tableExists(t, conn, fullName) {
		t.Fatalf("expected table %q to exist", fullName)
	}

	// Verify column order: id, title, status, priority, created_at, updated_at
	columns := getColumnNames(t, conn, fullName)
	expected := []string{"id", "title", "status", "priority", "created_at", "updated_at"}
	if len(columns) != len(expected) {
		t.Fatalf("expected %d columns, got %d: %v", len(expected), len(columns), columns)
	}
	for i, name := range expected {
		if columns[i] != name {
			t.Errorf("column %d: expected %q, got %q", i, name, columns[i])
		}
	}

	// Verify indexes were created.
	if !indexExists(t, conn, "idx_plugin_test_plugin_tasks_status") {
		t.Error("expected index idx_plugin_test_plugin_tasks_status to exist")
	}
	if !indexExists(t, conn, "idx_plugin_test_plugin_tasks_status_priority") {
		t.Error("expected index idx_plugin_test_plugin_tasks_status_priority to exist")
	}
}

func TestDefineTable_AutoInjectsColumns(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "inject_test")
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
		t.Fatalf("define_table failed: %v", err)
	}

	columns := getColumnNames(t, conn, "plugin_inject_test_items")
	// Must have: id (first), name (user), created_at, updated_at (last two)
	if len(columns) != 4 {
		t.Fatalf("expected 4 columns (id + name + created_at + updated_at), got %d: %v", len(columns), columns)
	}
	if columns[0] != "id" {
		t.Errorf("first column should be 'id', got %q", columns[0])
	}
	if columns[1] != "name" {
		t.Errorf("second column should be 'name', got %q", columns[1])
	}
	if columns[2] != "created_at" {
		t.Errorf("third column should be 'created_at', got %q", columns[2])
	}
	if columns[3] != "updated_at" {
		t.Errorf("fourth column should be 'updated_at', got %q", columns[3])
	}
}

func TestDefineTable_ReservedColumnRejected(t *testing.T) {
	reserved := []string{"id", "created_at", "updated_at"}

	for _, name := range reserved {
		t.Run(name, func(t *testing.T) {
			conn := openTestDB(t)
			defer conn.Close()

			L, cancel := newSchemaTestState(t, conn, "reserved_test")
			defer L.Close()
			defer cancel()

			code := fmt.Sprintf(`
				db.define_table("items", {
					columns = {
						{name = %q, type = "text"},
					},
				})
			`, name)
			err := L.DoString(code)
			if err == nil {
				t.Fatalf("expected error for reserved column %q, got nil", name)
			}
			if !strings.Contains(err.Error(), "auto-injected") {
				t.Errorf("expected 'auto-injected' in error, got: %s", err.Error())
			}
		})
	}
}

func TestDefineTable_InvalidColumnType(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "type_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {
				{name = "data", type = "varchar"},
			},
		})
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for invalid column type, got nil")
	}
	if !strings.Contains(err.Error(), "invalid column type") {
		t.Errorf("expected 'invalid column type' in error, got: %s", err.Error())
	}
}

func TestDefineTable_InvalidColumnName(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "name_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {
				{name = "has spaces", type = "text"},
			},
		})
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for invalid column name, got nil")
	}
	if !strings.Contains(err.Error(), "invalid column name") {
		t.Errorf("expected 'invalid column name' in error, got: %s", err.Error())
	}
}

func TestDefineTable_Idempotent(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "idempotent_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {
				{name = "name", type = "text", not_null = true},
			},
		})
	`
	// Call twice -- should not error due to IF NOT EXISTS.
	if err := L.DoString(code); err != nil {
		t.Fatalf("first define_table failed: %v", err)
	}
	if err := L.DoString(code); err != nil {
		t.Fatalf("second define_table failed (idempotency broken): %v", err)
	}

	if !tableExists(t, conn, "plugin_idempotent_test_items") {
		t.Fatal("table should exist after idempotent creation")
	}
}

func TestDefineTable_FKNamespaceValidation(t *testing.T) {
	t.Run("valid_same_prefix", func(t *testing.T) {
		conn := openTestDB(t)
		defer conn.Close()

		L, cancel := newSchemaTestState(t, conn, "fk_test")
		defer L.Close()
		defer cancel()

		// First create the referenced table.
		code := `
			db.define_table("categories", {
				columns = {
					{name = "name", type = "text", not_null = true},
				},
			})
		`
		if err := L.DoString(code); err != nil {
			t.Fatalf("define categories table failed: %v", err)
		}

		// Now create a table that references it with the full prefixed name.
		code = `
			db.define_table("items", {
				columns = {
					{name = "name", type = "text", not_null = true},
					{name = "category_id", type = "text"},
				},
				foreign_keys = {
					{column = "category_id", ref_table = "plugin_fk_test_categories", ref_column = "id", on_delete = "CASCADE"},
				},
			})
		`
		if err := L.DoString(code); err != nil {
			t.Fatalf("define items table with FK failed: %v", err)
		}

		if !tableExists(t, conn, "plugin_fk_test_items") {
			t.Fatal("expected items table to exist")
		}
	})

	t.Run("rejected_different_prefix", func(t *testing.T) {
		conn := openTestDB(t)
		defer conn.Close()

		L, cancel := newSchemaTestState(t, conn, "fk_test")
		defer L.Close()
		defer cancel()

		code := `
			db.define_table("items", {
				columns = {
					{name = "name", type = "text", not_null = true},
					{name = "other_id", type = "text"},
				},
				foreign_keys = {
					{column = "other_id", ref_table = "plugin_other_plugin_things", ref_column = "id"},
				},
			})
		`
		err := L.DoString(code)
		if err == nil {
			t.Fatal("expected error for FK referencing different plugin prefix, got nil")
		}
		if !strings.Contains(err.Error(), "namespace isolation") {
			t.Errorf("expected 'namespace isolation' in error, got: %s", err.Error())
		}
	})

	t.Run("rejected_core_table", func(t *testing.T) {
		conn := openTestDB(t)
		defer conn.Close()

		L, cancel := newSchemaTestState(t, conn, "fk_test")
		defer L.Close()
		defer cancel()

		code := `
			db.define_table("items", {
				columns = {
					{name = "name", type = "text", not_null = true},
					{name = "user_id", type = "text"},
				},
				foreign_keys = {
					{column = "user_id", ref_table = "users", ref_column = "id"},
				},
			})
		`
		err := L.DoString(code)
		if err == nil {
			t.Fatal("expected error for FK referencing core CMS table, got nil")
		}
		if !strings.Contains(err.Error(), "namespace isolation") {
			t.Errorf("expected 'namespace isolation' in error, got: %s", err.Error())
		}
	})
}

func TestDefineTable_EmptyColumnsRejected(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "empty_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {},
		})
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for empty columns, got nil")
	}
	if !strings.Contains(err.Error(), "columns cannot be empty") {
		t.Errorf("expected 'columns cannot be empty' in error, got: %s", err.Error())
	}
}

func TestDefineTable_MissingTableNameRejected(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "noname_test")
	defer L.Close()
	defer cancel()

	// Calling define_table without arguments raises a type error from CheckString.
	code := `db.define_table()`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for missing table name, got nil")
	}
}

func TestDefineTable_EmptyTableNameRejected(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "emptyname_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("", {
			columns = {
				{name = "x", type = "text"},
			},
		})
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for empty table name, got nil")
	}
	if !strings.Contains(err.Error(), "table name cannot be empty") {
		t.Errorf("expected 'table name cannot be empty' in error, got: %s", err.Error())
	}
}

func TestDefineTable_AllColumnTypesAccepted(t *testing.T) {
	// All 7 abstract column types must be accepted.
	columnTypes := []string{"text", "integer", "real", "blob", "boolean", "timestamp", "json"}

	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "types_test")
	defer L.Close()
	defer cancel()

	// Build a columns definition with all 7 types.
	var colDefs []string
	for i, ct := range columnTypes {
		colDefs = append(colDefs, fmt.Sprintf(`{name = "col_%d", type = %q}`, i, ct))
	}

	code := fmt.Sprintf(`
		db.define_table("all_types", {
			columns = {
				%s
			},
		})
	`, strings.Join(colDefs, ",\n\t\t\t\t"))

	if err := L.DoString(code); err != nil {
		t.Fatalf("define_table with all column types failed: %v", err)
	}

	if !tableExists(t, conn, "plugin_types_test_all_types") {
		t.Fatal("expected table to exist")
	}

	// Verify all columns exist: id + 7 user columns + created_at + updated_at = 10
	columns := getColumnNames(t, conn, "plugin_types_test_all_types")
	if len(columns) != 10 {
		t.Fatalf("expected 10 columns, got %d: %v", len(columns), columns)
	}
}

func TestDefineTable_MaxColumnsValidation(t *testing.T) {
	// The DDL layer enforces max 64 columns. With 3 auto-injected columns,
	// the plugin can define at most 61 user columns. Test that 62 user columns
	// (65 total) triggers the max columns error from DDLCreateTable.
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "maxcol_test")
	defer L.Close()
	defer cancel()

	var colDefs []string
	for i := range 62 {
		colDefs = append(colDefs, fmt.Sprintf(`{name = "col_%d", type = "text"}`, i))
	}

	code := fmt.Sprintf(`
		db.define_table("wide_table", {
			columns = {
				%s
			},
		})
	`, strings.Join(colDefs, ",\n\t\t\t\t"))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for too many columns, got nil")
	}
	if !strings.Contains(err.Error(), "too many columns") {
		t.Errorf("expected 'too many columns' in error, got: %s", err.Error())
	}
}

func TestDefineTable_IndexCreation(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "idx_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {
				{name = "name",     type = "text",    not_null = true},
				{name = "category", type = "text"},
				{name = "priority", type = "integer"},
			},
			indexes = {
				{columns = {"name"}, unique = true},
				{columns = {"category", "priority"}},
			},
		})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("define_table failed: %v", err)
	}

	// Verify unique index was created.
	if !indexExists(t, conn, "idx_plugin_idx_test_items_name") {
		t.Error("expected unique index on name to exist")
	}

	// Verify composite index was created.
	if !indexExists(t, conn, "idx_plugin_idx_test_items_category_priority") {
		t.Error("expected composite index on category_priority to exist")
	}
}

func TestDefineTable_DefaultValues(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "default_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {
				{name = "status", type = "text", not_null = true, default = "pending"},
				{name = "count",  type = "integer", not_null = true, default = 0},
			},
		})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("define_table failed: %v", err)
	}

	// Insert a row with only the required columns to verify defaults work.
	fullName := "plugin_default_test_items"
	_, err := conn.Exec(
		fmt.Sprintf(`INSERT INTO %q ("id", "created_at", "updated_at") VALUES ('test_id', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`, fullName),
	)
	if err != nil {
		t.Fatalf("insert with defaults failed: %v", err)
	}

	var status string
	var count int
	err = conn.QueryRow(fmt.Sprintf(`SELECT "status", "count" FROM %q WHERE "id" = 'test_id'`, fullName)).Scan(&status, &count)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if status != "pending" {
		t.Errorf("expected default status 'pending', got %q", status)
	}
	if count != 0 {
		t.Errorf("expected default count 0, got %d", count)
	}
}

func TestDefineTable_MissingColumnsField(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "nocols_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {})
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for missing columns field, got nil")
	}
	// Should mention 'columns' in the error.
	if !strings.Contains(err.Error(), "columns") {
		t.Errorf("expected error mentioning 'columns', got: %s", err.Error())
	}
}

func TestDefineTable_ColumnMissingName(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "noname_col_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {
				{type = "text"},
			},
		})
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for column missing name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error mentioning 'name', got: %s", err.Error())
	}
}

func TestDefineTable_ColumnMissingType(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "notype_col_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {
				{name = "x"},
			},
		})
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for column missing type, got nil")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Errorf("expected error mentioning 'type', got: %s", err.Error())
	}
}

func TestDefineTable_SQLKeywordColumnNameRejected(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "keyword_test")
	defer L.Close()
	defer cancel()

	// "SELECT" is a SQL keyword and should be rejected as a column name.
	code := `
		db.define_table("items", {
			columns = {
				{name = "SELECT", type = "text"},
			},
		})
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for SQL keyword column name, got nil")
	}
}

func TestDefineTable_TablePrefixingCorrect(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "myplugin")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("my_data", {
			columns = {
				{name = "value", type = "text"},
			},
		})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("define_table failed: %v", err)
	}

	// The table should be named "plugin_myplugin_my_data", not "my_data".
	if tableExists(t, conn, "my_data") {
		t.Error("raw table name should not exist (should be prefixed)")
	}
	if !tableExists(t, conn, "plugin_myplugin_my_data") {
		t.Fatal("expected prefixed table 'plugin_myplugin_my_data' to exist")
	}
}

func TestDefineTable_UniqueColumnConstraint(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "unique_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("items", {
			columns = {
				{name = "email", type = "text", not_null = true, unique = true},
			},
		})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("define_table failed: %v", err)
	}

	fullName := "plugin_unique_test_items"

	// Insert one row successfully.
	_, err := conn.Exec(
		fmt.Sprintf(`INSERT INTO %q ("id", "email", "created_at", "updated_at") VALUES ('id1', 'a@b.com', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`, fullName),
	)
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	// Second insert with same email should fail due to UNIQUE constraint.
	_, err = conn.Exec(
		fmt.Sprintf(`INSERT INTO %q ("id", "email", "created_at", "updated_at") VALUES ('id2', 'a@b.com', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`, fullName),
	)
	if err == nil {
		t.Fatal("expected unique constraint violation, got nil")
	}
}

func TestDefineTable_FKOnDeleteCascade(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	// Enable foreign keys in SQLite (they are off by default).
	_, err := conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	L, cancel := newSchemaTestState(t, conn, "cascade_test")
	defer L.Close()
	defer cancel()

	// Create parent table.
	code := `
		db.define_table("parents", {
			columns = {
				{name = "name", type = "text", not_null = true},
			},
		})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("define parents failed: %v", err)
	}

	// Create child table with FK to parent.
	code = `
		db.define_table("children", {
			columns = {
				{name = "parent_id", type = "text", not_null = true},
				{name = "label",     type = "text"},
			},
			foreign_keys = {
				{column = "parent_id", ref_table = "plugin_cascade_test_parents", ref_column = "id", on_delete = "cascade"},
			},
		})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("define children failed: %v", err)
	}

	// Insert parent and child.
	_, err = conn.Exec(`INSERT INTO "plugin_cascade_test_parents" ("id", "name", "created_at", "updated_at") VALUES ('p1', 'Parent', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert parent failed: %v", err)
	}
	_, err = conn.Exec(`INSERT INTO "plugin_cascade_test_children" ("id", "parent_id", "label", "created_at", "updated_at") VALUES ('c1', 'p1', 'Child', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert child failed: %v", err)
	}

	// Delete parent -- child should be cascade deleted.
	_, err = conn.Exec(`DELETE FROM "plugin_cascade_test_parents" WHERE "id" = 'p1'`)
	if err != nil {
		t.Fatalf("delete parent failed: %v", err)
	}

	var count int
	err = conn.QueryRow(`SELECT COUNT(*) FROM "plugin_cascade_test_children"`).Scan(&count)
	if err != nil {
		t.Fatalf("count children failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 children after cascade delete, got %d", count)
	}
}

func TestDefineTable_NoIndexes(t *testing.T) {
	// Verify that defining a table without indexes works fine.
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "noidx_test")
	defer L.Close()
	defer cancel()

	code := `
		db.define_table("simple", {
			columns = {
				{name = "value", type = "text"},
			},
		})
	`
	if err := L.DoString(code); err != nil {
		t.Fatalf("define_table failed: %v", err)
	}

	if !tableExists(t, conn, "plugin_noidx_test_simple") {
		t.Fatal("expected table to exist")
	}
}

func TestDefineTable_DuplicateColumnNameRejected(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	L, cancel := newSchemaTestState(t, conn, "dup_test")
	defer L.Close()
	defer cancel()

	// The DDL layer validates duplicate column names.
	code := `
		db.define_table("items", {
			columns = {
				{name = "title", type = "text"},
				{name = "title", type = "text"},
			},
		})
	`
	err := L.DoString(code)
	if err == nil {
		t.Fatal("expected error for duplicate column name, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate column") {
		t.Errorf("expected 'duplicate column' in error, got: %s", err.Error())
	}
}

func TestDefineTable_FKMissingFields(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "missing column",
			code: `
				db.define_table("items", {
					columns = {{name = "ref_id", type = "text"}},
					foreign_keys = {{ref_table = "plugin_fk_miss_test_other", ref_column = "id"}},
				})
			`,
			want: "column",
		},
		{
			name: "missing ref_table",
			code: `
				db.define_table("items", {
					columns = {{name = "ref_id", type = "text"}},
					foreign_keys = {{column = "ref_id", ref_column = "id"}},
				})
			`,
			want: "ref_table",
		},
		{
			name: "missing ref_column",
			code: `
				db.define_table("items", {
					columns = {{name = "ref_id", type = "text"}},
					foreign_keys = {{column = "ref_id", ref_table = "plugin_fk_miss_test_other"}},
				})
			`,
			want: "ref_column",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn := openTestDB(t)
			defer conn.Close()

			L, cancel := newSchemaTestState(t, conn, "fk_miss_test")
			defer L.Close()
			defer cancel()

			err := L.DoString(tc.code)
			if err == nil {
				t.Fatal("expected error for missing FK field, got nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("expected error containing %q, got: %s", tc.want, err.Error())
			}
		})
	}
}

func TestTablePrefix(t *testing.T) {
	tests := []struct {
		pluginName string
		expected   string
	}{
		{"task_tracker", "plugin_task_tracker_"},
		{"analytics", "plugin_analytics_"},
		{"my_plugin", "plugin_my_plugin_"},
	}

	for _, tc := range tests {
		t.Run(tc.pluginName, func(t *testing.T) {
			result := tablePrefix(tc.pluginName)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestParseColumns_NonTableEntry(t *testing.T) {
	L := newTestState()
	defer L.Close()
	ApplySandbox(L, SandboxConfig{})

	// Build a columns table with a non-table entry.
	tbl := L.NewTable()
	tbl.Append(lua.LString("not_a_table"))

	_, err := parseColumns(L, tbl, "test")
	if err == nil {
		t.Fatal("expected error for non-table column entry, got nil")
	}
	if !strings.Contains(err.Error(), "each column must be a table") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestParseIndexes_NonTableEntry(t *testing.T) {
	L := newTestState()
	defer L.Close()
	ApplySandbox(L, SandboxConfig{})

	tbl := L.NewTable()
	tbl.Append(lua.LString("not_a_table"))

	_, err := parseIndexes(L, tbl)
	if err == nil {
		t.Fatal("expected error for non-table index entry, got nil")
	}
	if !strings.Contains(err.Error(), "each index must be a table") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestParseForeignKeys_NonTableEntry(t *testing.T) {
	L := newTestState()
	defer L.Close()
	ApplySandbox(L, SandboxConfig{})

	tbl := L.NewTable()
	tbl.Append(lua.LString("not_a_table"))

	_, err := parseForeignKeys(L, tbl, "plugin_test_")
	if err == nil {
		t.Fatal("expected error for non-table FK entry, got nil")
	}
	if !strings.Contains(err.Error(), "each foreign_key must be a table") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestLuaBoolField(t *testing.T) {
	L := newTestState()
	defer L.Close()
	ApplySandbox(L, SandboxConfig{})

	tbl := L.NewTable()
	L.SetField(tbl, "yes", lua.LTrue)
	L.SetField(tbl, "no", lua.LFalse)
	L.SetField(tbl, "str", lua.LString("hello"))

	if !luaBoolField(L, tbl, "yes") {
		t.Error("expected true for 'yes' field")
	}
	if luaBoolField(L, tbl, "no") {
		t.Error("expected false for 'no' field")
	}
	if luaBoolField(L, tbl, "str") {
		t.Error("expected false for non-bool field")
	}
	if luaBoolField(L, tbl, "missing") {
		t.Error("expected false for missing field")
	}
}

func TestLuaStringOrNumberField(t *testing.T) {
	L := newTestState()
	defer L.Close()
	ApplySandbox(L, SandboxConfig{})

	tbl := L.NewTable()
	L.SetField(tbl, "str_val", lua.LString("hello"))
	L.SetField(tbl, "num_val", lua.LNumber(42))
	L.SetField(tbl, "bool_val", lua.LTrue)

	if got := luaStringOrNumberField(L, tbl, "str_val"); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
	if got := luaStringOrNumberField(L, tbl, "num_val"); got != "42" {
		t.Errorf("expected '42', got %q", got)
	}
	if got := luaStringOrNumberField(L, tbl, "bool_val"); got != "" {
		t.Errorf("expected empty string for bool, got %q", got)
	}
	if got := luaStringOrNumberField(L, tbl, "missing"); got != "" {
		t.Errorf("expected empty string for missing, got %q", got)
	}
}
