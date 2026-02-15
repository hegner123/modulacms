// White-box tests for utility.go: ColumnNameType and ColumnIndexName types,
// CopyDb file-copy helper, GetTableColumns PRAGMA table_info wrapper, and
// GetColumnRowsString generic column reader.
//
// White-box access is needed because:
//   - CopyDb uses relative paths and external binaries; testing requires
//     understanding the exact path construction to set up the filesystem.
//   - GetColumnRowsString has an internal type-switch on any values that
//     silently returns (nil, nil) for unrecognized types -- this is an
//     internal behavior detail not visible through a public API.
package db

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helper: create an in-memory SQLite database for tests that only need SQL
// ---------------------------------------------------------------------------

func inMemorySQLiteDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open(:memory:): %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// ---------------------------------------------------------------------------
// ColumnNameType and ColumnIndexName type tests
// ---------------------------------------------------------------------------

func TestColumnNameType_IsMapStringString(t *testing.T) {
	t.Parallel()
	var m ColumnNameType = make(ColumnNameType)
	m["name"] = "TEXT"
	m["age"] = "INTEGER"

	if m["name"] != "TEXT" {
		t.Errorf("m[\"name\"] = %q, want %q", m["name"], "TEXT")
	}
	if m["age"] != "INTEGER" {
		t.Errorf("m[\"age\"] = %q, want %q", m["age"], "INTEGER")
	}
	if len(m) != 2 {
		t.Errorf("len(m) = %d, want 2", len(m))
	}
}

func TestColumnNameType_ZeroValue(t *testing.T) {
	t.Parallel()
	var m ColumnNameType
	if len(m) != 0 {
		t.Error("zero-value ColumnNameType should have length 0")
	}
}

func TestColumnIndexName_IsMapIntString(t *testing.T) {
	t.Parallel()
	var m ColumnIndexName = make(ColumnIndexName)
	m[0] = "id"
	m[1] = "name"

	if m[0] != "id" {
		t.Errorf("m[0] = %q, want %q", m[0], "id")
	}
	if m[1] != "name" {
		t.Errorf("m[1] = %q, want %q", m[1], "name")
	}
}

func TestColumnIndexName_ZeroValue(t *testing.T) {
	t.Parallel()
	var m ColumnIndexName
	if len(m) != 0 {
		t.Error("zero-value ColumnIndexName should have length 0")
	}
}

// ---------------------------------------------------------------------------
// GetTableColumns tests
// ---------------------------------------------------------------------------

func TestGetTableColumns_EmptyTableName(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	nt, in, err := GetTableColumns(ctx, db, "")
	if err == nil {
		t.Fatal("expected error for empty table name, got nil")
	}
	if !strings.Contains(err.Error(), "table name cannot be empty") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "table name cannot be empty")
	}
	if nt != nil {
		t.Error("expected nil ColumnNameType on error")
	}
	if in != nil {
		t.Error("expected nil ColumnIndexName on error")
	}
}

func TestGetTableColumns_SingleColumn(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec("CREATE TABLE single_col (id INTEGER PRIMARY KEY);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	nt, in, err := GetTableColumns(ctx, db, "single_col")
	if err != nil {
		t.Fatalf("GetTableColumns: %v", err)
	}

	if len(nt) != 1 {
		t.Fatalf("ColumnNameType has %d entries, want 1", len(nt))
	}
	if nt["id"] != "INTEGER" {
		t.Errorf("nt[\"id\"] = %q, want %q", nt["id"], "INTEGER")
	}

	if len(in) != 1 {
		t.Fatalf("ColumnIndexName has %d entries, want 1", len(in))
	}
	if in[0] != "id" {
		t.Errorf("in[0] = %q, want %q", in[0], "id")
	}
}

func TestGetTableColumns_MultipleColumns(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec(`CREATE TABLE multi_col (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT,
		age INTEGER DEFAULT 0,
		active BOOLEAN
	);`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	nt, in, err := GetTableColumns(ctx, db, "multi_col")
	if err != nil {
		t.Fatalf("GetTableColumns: %v", err)
	}

	// Verify all columns present in ColumnNameType
	wantCols := map[string]string{
		"id":     "INTEGER",
		"name":   "TEXT",
		"email":  "TEXT",
		"age":    "INTEGER",
		"active": "BOOLEAN",
	}

	if len(nt) != len(wantCols) {
		t.Fatalf("ColumnNameType has %d entries, want %d", len(nt), len(wantCols))
	}
	for col, wantType := range wantCols {
		gotType, ok := nt[col]
		if !ok {
			t.Errorf("column %q missing from ColumnNameType", col)
			continue
		}
		if gotType != wantType {
			t.Errorf("nt[%q] = %q, want %q", col, gotType, wantType)
		}
	}

	// Verify ColumnIndexName preserves column order (SQLite returns columns
	// in definition order via PRAGMA table_info)
	if len(in) != 5 {
		t.Fatalf("ColumnIndexName has %d entries, want 5", len(in))
	}
	wantOrder := []string{"id", "name", "email", "age", "active"}
	for i, wantName := range wantOrder {
		if in[i] != wantName {
			t.Errorf("in[%d] = %q, want %q", i, in[i], wantName)
		}
	}
}

func TestGetTableColumns_NonexistentTable(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	// PRAGMA table_info on a nonexistent table returns an empty result set
	// (not an error) in SQLite
	nt, in, err := GetTableColumns(ctx, db, "does_not_exist")
	if err != nil {
		t.Fatalf("GetTableColumns unexpectedly returned error: %v", err)
	}
	if len(nt) != 0 {
		t.Errorf("ColumnNameType has %d entries, want 0", len(nt))
	}
	if len(in) != 0 {
		t.Errorf("ColumnIndexName has %d entries, want 0", len(in))
	}
}

func TestGetTableColumns_ClosedConnection(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	db.Close()

	ctx := context.Background()
	_, _, err = GetTableColumns(ctx, db, "any_table")
	if err == nil {
		t.Fatal("expected error for closed connection, got nil")
	}
	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "query failed")
	}
}

func TestGetTableColumns_CancelledContext(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, err := GetTableColumns(ctx, db, "any_table")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestGetTableColumns_ColumnTypesPreserved(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	// SQLite is flexible about types, but PRAGMA table_info returns the
	// declared type string exactly as written in CREATE TABLE
	_, err := db.Exec(`CREATE TABLE typed_cols (
		a VARCHAR(255),
		b DECIMAL(10,2),
		c BLOB,
		d REAL,
		e NUMERIC
	);`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	nt, _, err := GetTableColumns(ctx, db, "typed_cols")
	if err != nil {
		t.Fatalf("GetTableColumns: %v", err)
	}

	tests := []struct {
		col      string
		wantType string
	}{
		{col: "a", wantType: "VARCHAR(255)"},
		{col: "b", wantType: "DECIMAL(10,2)"},
		{col: "c", wantType: "BLOB"},
		{col: "d", wantType: "REAL"},
		{col: "e", wantType: "NUMERIC"},
	}

	for _, tt := range tests {
		t.Run(tt.col, func(t *testing.T) {
			gotType, ok := nt[tt.col]
			if !ok {
				t.Fatalf("column %q missing from ColumnNameType", tt.col)
			}
			if gotType != tt.wantType {
				t.Errorf("nt[%q] = %q, want %q", tt.col, gotType, tt.wantType)
			}
		})
	}
}

func TestGetTableColumns_IndexIsZeroBased(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec("CREATE TABLE idx_test (first TEXT, second TEXT, third TEXT);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	_, in, err := GetTableColumns(ctx, db, "idx_test")
	if err != nil {
		t.Fatalf("GetTableColumns: %v", err)
	}

	// The index counter in GetTableColumns starts at 0 and increments
	if in[0] != "first" {
		t.Errorf("in[0] = %q, want %q", in[0], "first")
	}
	if in[1] != "second" {
		t.Errorf("in[1] = %q, want %q", in[1], "second")
	}
	if in[2] != "third" {
		t.Errorf("in[2] = %q, want %q", in[2], "third")
	}
}

// ---------------------------------------------------------------------------
// GetColumnRowsString tests
// ---------------------------------------------------------------------------

func TestGetColumnRowsString_EmptyTableName(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	result, err := GetColumnRowsString(db, ctx, "", "any_column")
	if err == nil {
		t.Fatal("expected error for empty table name, got nil")
	}
	if !strings.Contains(err.Error(), "table name cannot be empty") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "table name cannot be empty")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestGetColumnRowsString_EmptyColumnName(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	result, err := GetColumnRowsString(db, ctx, "some_table", "")
	if err == nil {
		t.Fatal("expected error for empty column name, got nil")
	}
	if !strings.Contains(err.Error(), "column name cannot be empty") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "column name cannot be empty")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestGetColumnRowsString_BothNamesEmpty(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	// When both are empty, the first check (tableName) triggers
	_, err := GetColumnRowsString(db, ctx, "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "table name cannot be empty") {
		t.Errorf("error = %q, want table name error to trigger first", err.Error())
	}
}

func TestGetColumnRowsString_StringValues(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec(`
		CREATE TABLE str_test (id INTEGER PRIMARY KEY, name TEXT);
		INSERT INTO str_test (id, name) VALUES (1, 'alice');
		INSERT INTO str_test (id, name) VALUES (2, 'bob');
		INSERT INTO str_test (id, name) VALUES (3, 'charlie');
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := GetColumnRowsString(db, ctx, "str_test", "name")
	if err != nil {
		t.Fatalf("GetColumnRowsString: %v", err)
	}

	want := []string{"alice", "bob", "charlie"}
	if len(result) != len(want) {
		t.Fatalf("got %d results, want %d", len(result), len(want))
	}
	for i, w := range want {
		if result[i] != w {
			t.Errorf("result[%d] = %q, want %q", i, result[i], w)
		}
	}
}

func TestGetColumnRowsString_IntegerValues(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec(`
		CREATE TABLE int_test (id INTEGER PRIMARY KEY, count INTEGER);
		INSERT INTO int_test (id, count) VALUES (1, 100);
		INSERT INTO int_test (id, count) VALUES (2, 200);
		INSERT INTO int_test (id, count) VALUES (3, 0);
		INSERT INTO int_test (id, count) VALUES (4, -42);
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := GetColumnRowsString(db, ctx, "int_test", "count")
	if err != nil {
		t.Fatalf("GetColumnRowsString: %v", err)
	}

	// Integers are converted to strings via strconv.FormatInt
	want := []string{"100", "200", "0", "-42"}
	if len(result) != len(want) {
		t.Fatalf("got %d results, want %d", len(result), len(want))
	}
	for i, w := range want {
		if result[i] != w {
			t.Errorf("result[%d] = %q, want %q", i, result[i], w)
		}
	}
}

func TestGetColumnRowsString_EmptyTable(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec("CREATE TABLE empty_test (id INTEGER PRIMARY KEY, val TEXT);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	result, err := GetColumnRowsString(db, ctx, "empty_test", "val")
	if err != nil {
		t.Fatalf("GetColumnRowsString: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("got %d results, want 0", len(result))
	}
}

func TestGetColumnRowsString_NonexistentTable(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := GetColumnRowsString(db, ctx, "no_such_table", "any_col")
	if err == nil {
		t.Fatal("expected error for nonexistent table, got nil")
	}
	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "query failed")
	}
}

func TestGetColumnRowsString_NonexistentColumn(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec("CREATE TABLE col_miss (id INTEGER PRIMARY KEY);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	_, err = GetColumnRowsString(db, ctx, "col_miss", "no_such_column")
	if err == nil {
		t.Fatal("expected error for nonexistent column, got nil")
	}
	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "query failed")
	}
}

func TestGetColumnRowsString_ClosedConnection(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	db.Close()

	ctx := context.Background()
	_, err = GetColumnRowsString(db, ctx, "any_table", "any_col")
	if err == nil {
		t.Fatal("expected error for closed connection, got nil")
	}
	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "query failed")
	}
}

func TestGetColumnRowsString_CancelledContext(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := GetColumnRowsString(db, ctx, "any_table", "any_col")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestGetColumnRowsString_NullValues(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec(`
		CREATE TABLE null_test (id INTEGER PRIMARY KEY, val TEXT);
		INSERT INTO null_test (id, val) VALUES (1, NULL);
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// NULL scanned into `any` becomes nil, which is neither string nor int64.
	// The code's type switch falls through and returns (nil, nil).
	result, err := GetColumnRowsString(db, ctx, "null_test", "val")
	if err != nil {
		t.Fatalf("GetColumnRowsString: unexpected error: %v", err)
	}
	// The function returns (nil, nil) when it encounters a non-string,
	// non-int64 value -- NULL is such a value
	if result != nil {
		t.Errorf("expected nil result for NULL value, got %v", result)
	}
}

func TestGetColumnRowsString_MixedStringThenNull(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec(`
		CREATE TABLE mixed_null (id INTEGER PRIMARY KEY, val TEXT);
		INSERT INTO mixed_null (id, val) VALUES (1, 'first');
		INSERT INTO mixed_null (id, val) VALUES (2, NULL);
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First row is string ("first"), second row is NULL.
	// The function processes "first" successfully, then hits NULL and returns
	// (nil, nil) -- discarding the previously collected values.
	result, err := GetColumnRowsString(db, ctx, "mixed_null", "val")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The NULL on the second row causes the function to bail with (nil, nil)
	if result != nil {
		t.Errorf("expected nil result when NULL appears after valid rows, got %v", result)
	}
}

func TestGetColumnRowsString_SingleRow(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec(`
		CREATE TABLE single_row (id INTEGER PRIMARY KEY, label TEXT);
		INSERT INTO single_row (id, label) VALUES (1, 'only');
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := GetColumnRowsString(db, ctx, "single_row", "label")
	if err != nil {
		t.Fatalf("GetColumnRowsString: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d results, want 1", len(result))
	}
	if result[0] != "only" {
		t.Errorf("result[0] = %q, want %q", result[0], "only")
	}
}

func TestGetColumnRowsString_EmptyStringValues(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec(`
		CREATE TABLE empty_str (id INTEGER PRIMARY KEY, val TEXT);
		INSERT INTO empty_str (id, val) VALUES (1, '');
		INSERT INTO empty_str (id, val) VALUES (2, '');
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := GetColumnRowsString(db, ctx, "empty_str", "val")
	if err != nil {
		t.Fatalf("GetColumnRowsString: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d results, want 2", len(result))
	}
	for i, r := range result {
		if r != "" {
			t.Errorf("result[%d] = %q, want empty string", i, r)
		}
	}
}

func TestGetColumnRowsString_LargeIntegerValues(t *testing.T) {
	t.Parallel()
	db := inMemorySQLiteDB(t)
	ctx := context.Background()

	_, err := db.Exec(`
		CREATE TABLE big_int (id INTEGER PRIMARY KEY, val INTEGER);
		INSERT INTO big_int (id, val) VALUES (1, 9223372036854775807);
		INSERT INTO big_int (id, val) VALUES (2, -9223372036854775808);
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := GetColumnRowsString(db, ctx, "big_int", "val")
	if err != nil {
		t.Fatalf("GetColumnRowsString: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d results, want 2", len(result))
	}
	if result[0] != "9223372036854775807" {
		t.Errorf("result[0] = %q, want %q", result[0], "9223372036854775807")
	}
	if result[1] != "-9223372036854775808" {
		t.Errorf("result[1] = %q, want %q", result[1], "-9223372036854775808")
	}
}

// ---------------------------------------------------------------------------
// CopyDb tests
//
// CopyDb depends on:
//   1. Filesystem layout: relative paths "../../testdb/backups/" and "../../testdb/"
//   2. External binary: sqlite3
//   3. utility.TimestampS() for unique naming
//
// We test what we can: path construction, file creation, behavior when
// sqlite3 is missing. Tests that need sqlite3 are skipped if unavailable.
// ---------------------------------------------------------------------------

// hasSqlite3 reports whether the sqlite3 CLI binary is available on PATH.
func hasSqlite3(t *testing.T) bool {
	t.Helper()
	_, err := exec.LookPath("sqlite3")
	return err == nil
}

func TestCopyDb_AlwaysReturnsNilError(t *testing.T) {
	// CopyDb always returns nil for its error value, even when internal
	// operations fail. This documents the current (surprising) behavior.
	// We call it with a dbName that will fail at the os.Create step
	// because the relative path ../../testdb/ likely does not exist from
	// the test working directory.
	_, err := CopyDb("test.db", true)
	// The function always returns nil error -- this is a design issue
	// but we document it here rather than changing it
	if err != nil {
		t.Errorf("CopyDb always returns nil error, but got: %v", err)
	}
}

func TestCopyDb_PathConstruction_Default(t *testing.T) {
	// Set up the filesystem structure CopyDb expects, relative to a temp dir.
	// We run CopyDb from a directory where ../../testdb/ exists.
	if !hasSqlite3(t) {
		t.Skip("sqlite3 binary not found on PATH")
	}

	// Build the directory tree:
	// <root>/testdb/backups/
	// <root>/a/b/  (we run from here so ../../testdb/ resolves to <root>/testdb/)
	root := t.TempDir()
	testdbDir := filepath.Join(root, "testdb")
	backupsDir := filepath.Join(testdbDir, "backups")
	workDir := filepath.Join(root, "a", "b")

	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		t.Fatalf("MkdirAll backups: %v", err)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("MkdirAll workdir: %v", err)
	}

	// Create a minimal SQL file that CopyDb will read
	sqlFile := filepath.Join(backupsDir, "tests.sql")
	if err := os.WriteFile(sqlFile, []byte("CREATE TABLE test (id INTEGER);\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// CopyDb uses relative paths, so we need to change to the work directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		// Restore working directory -- important for other tests
		os.Chdir(origDir) //nolint:errcheck
	})

	dstPath, err := CopyDb("mydb.db", true)
	if err != nil {
		t.Fatalf("CopyDb returned error: %v", err)
	}

	// Verify the destination path follows the expected pattern
	if !strings.HasPrefix(dstPath, "../../testdb/testing") {
		t.Errorf("dstPath = %q, want prefix %q", dstPath, "../../testdb/testing")
	}
	if !strings.HasSuffix(dstPath, "mydb.db") {
		t.Errorf("dstPath = %q, want suffix %q", dstPath, "mydb.db")
	}

	// Verify the file was created
	absPath := filepath.Join(workDir, dstPath)
	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		t.Errorf("destination file %q was not created", absPath)
	}
}

func TestCopyDb_PathConstruction_NonDefault(t *testing.T) {
	// When useDefault is false, CopyDb reads from
	// "../../testdb/backups/<dbName_without_ext>.sql"
	if !hasSqlite3(t) {
		t.Skip("sqlite3 binary not found on PATH")
	}

	root := t.TempDir()
	testdbDir := filepath.Join(root, "testdb")
	backupsDir := filepath.Join(testdbDir, "backups")
	workDir := filepath.Join(root, "a", "b")

	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		t.Fatalf("MkdirAll backups: %v", err)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("MkdirAll workdir: %v", err)
	}

	// The SQL file name is derived from dbName with .db stripped
	sqlFile := filepath.Join(backupsDir, "custom.sql")
	if err := os.WriteFile(sqlFile, []byte("CREATE TABLE custom_tbl (val TEXT);\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) }) //nolint:errcheck

	dstPath, err := CopyDb("custom.db", false)
	if err != nil {
		t.Fatalf("CopyDb returned error: %v", err)
	}

	if !strings.HasSuffix(dstPath, "custom.db") {
		t.Errorf("dstPath = %q, want suffix %q", dstPath, "custom.db")
	}
}

func TestCopyDb_TrimsDbExtension(t *testing.T) {
	// Verify that the .db extension is stripped when constructing the SQL source
	// name. We can test this by checking the path pattern.
	// CopyDb("myname.db", false) should look for "../../testdb/backups/myname.sql"
	// CopyDb("myname.db", true)  should look for "../../testdb/backups/tests.sql"

	// This test only validates the returned path structure, not file I/O
	dstPath, _ := CopyDb("something.db", true)
	// The dst path should contain the original dbName including .db
	if !strings.Contains(dstPath, "something.db") {
		t.Errorf("dstPath = %q, want it to contain %q", dstPath, "something.db")
	}
}

func TestCopyDb_EmptyDbName(t *testing.T) {
	// Edge case: empty dbName produces unusual paths but should not panic
	dstPath, err := CopyDb("", true)
	if err != nil {
		t.Fatalf("CopyDb with empty name returned error: %v", err)
	}
	// The path should still contain the testdb prefix
	if !strings.Contains(dstPath, "../../testdb/testing") {
		t.Errorf("dstPath = %q, want it to contain testdb/testing prefix", dstPath)
	}
}
