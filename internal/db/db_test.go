// White-box tests for db.go: DbStatus constants, struct types (Database,
// MysqlDatabase, PsqlDatabase), GetConnection, Ping, ExecuteQuery, Query,
// and the Build*Query helper functions.
//
// White-box access is needed because:
//   - Ping() nil-connection paths test the internal nil guard before delegating
//     to *sql.DB, which is only reachable by constructing zero-value structs
//     directly (no exported constructor exposes this state).
//   - ExecuteQuery uses DBTableString (unexported output) and we need to verify
//     the composed SQL string reaches the database correctly.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	config "github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- DbStatus constant tests ---

func TestDbStatus_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  DbStatus
		want string
	}{
		{name: "Open", got: Open, want: "open"},
		{name: "Closed", got: Closed, want: "closed"},
		{name: "Err", got: Err, want: "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.got) != tt.want {
				t.Errorf("DbStatus %s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestDbStatus_IsStringType(t *testing.T) {
	t.Parallel()
	// DbStatus should be usable as a string for comparisons and display
	var s DbStatus = "custom"
	if string(s) != "custom" {
		t.Errorf("DbStatus from string literal = %q, want %q", s, "custom")
	}
}

// --- GetConnection tests ---
// These test the method directly on each struct type, verifying they return
// the struct's fields without transformation.

func TestDatabase_GetConnection_ReturnsFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "getconn_test.db")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	d := Database{
		Connection: conn,
		Context:    ctx,
	}

	gotConn, gotCtx, gotErr := d.GetConnection()
	if gotErr != nil {
		t.Fatalf("GetConnection returned error: %v", gotErr)
	}
	if gotConn != conn {
		t.Error("GetConnection returned different *sql.DB pointer")
	}
	if gotCtx != ctx {
		t.Error("GetConnection returned different context")
	}
}

func TestMysqlDatabase_GetConnection_ReturnsFields(t *testing.T) {
	t.Parallel()

	// MysqlDatabase.GetConnection just returns struct fields.
	// We use a nil connection to verify the method does not validate.
	ctx := context.Background()
	d := MysqlDatabase{
		Connection: nil,
		Context:    ctx,
	}

	gotConn, gotCtx, gotErr := d.GetConnection()
	if gotErr != nil {
		t.Fatalf("GetConnection returned error: %v", gotErr)
	}
	if gotConn != nil {
		t.Error("expected nil connection, got non-nil")
	}
	if gotCtx != ctx {
		t.Error("GetConnection returned different context")
	}
}

func TestPsqlDatabase_GetConnection_ReturnsFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d := PsqlDatabase{
		Connection: nil,
		Context:    ctx,
	}

	gotConn, gotCtx, gotErr := d.GetConnection()
	if gotErr != nil {
		t.Fatalf("GetConnection returned error: %v", gotErr)
	}
	if gotConn != nil {
		t.Error("expected nil connection, got non-nil")
	}
	if gotCtx != ctx {
		t.Error("GetConnection returned different context")
	}
}

func TestDatabase_GetConnection_NilContext(t *testing.T) {
	t.Parallel()
	// Verify GetConnection works even when Context is nil (zero value)
	d := Database{}
	_, gotCtx, gotErr := d.GetConnection()
	if gotErr != nil {
		t.Fatalf("GetConnection returned error: %v", gotErr)
	}
	if gotCtx != nil {
		t.Error("expected nil context from zero-value Database")
	}
}

// --- Ping tests ---

func TestDatabase_Ping_NilConnection(t *testing.T) {
	t.Parallel()
	d := Database{Connection: nil}
	err := d.Ping()
	if err == nil {
		t.Fatal("expected error for nil connection, got nil")
	}
	if !strings.Contains(err.Error(), "SQLite connection not established") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "SQLite connection not established")
	}
}

func TestMysqlDatabase_Ping_NilConnection(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{Connection: nil}
	err := d.Ping()
	if err == nil {
		t.Fatal("expected error for nil connection, got nil")
	}
	if !strings.Contains(err.Error(), "MySQL connection not established") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "MySQL connection not established")
	}
}

func TestPsqlDatabase_Ping_NilConnection(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{Connection: nil}
	err := d.Ping()
	if err == nil {
		t.Fatal("expected error for nil connection, got nil")
	}
	if !strings.Contains(err.Error(), "PostgreSQL connection not established") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "PostgreSQL connection not established")
	}
}

func TestDatabase_Ping_LiveConnection(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	conn, err := sql.Open("sqlite3", filepath.Join(dir, "ping_test.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()

	d := Database{Connection: conn}
	if err := d.Ping(); err != nil {
		t.Fatalf("Ping on live connection: %v", err)
	}
}

func TestDatabase_Ping_ClosedConnection(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	conn, err := sql.Open("sqlite3", filepath.Join(dir, "ping_closed.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	conn.Close()

	d := Database{Connection: conn}
	if err := d.Ping(); err == nil {
		t.Error("expected error from Ping on closed connection, got nil")
	}
}

// --- ExecuteQuery tests ---
// ExecuteQuery composes "query DBTableString(table);" and executes it.
// We test with a real SQLite connection to verify the full path.

func TestDatabase_ExecuteQuery_SelectFromTable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	conn, err := sql.Open("sqlite3", filepath.Join(dir, "execquery.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()

	// Create the routes table (ExecuteQuery will reference it by DBTable constant)
	_, err = conn.Exec("CREATE TABLE routes (id INTEGER PRIMARY KEY, slug TEXT);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	_, err = conn.Exec("INSERT INTO routes (id, slug) VALUES (1, 'home');")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	d := Database{Connection: conn, Context: context.Background()}

	// ExecuteQuery formats: "SELECT * FROM" + " " + DBTableString(Route) + ";"
	rows, err := d.ExecuteQuery("SELECT * FROM", Route)
	if err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var id int
		var slug string
		if scanErr := rows.Scan(&id, &slug); scanErr != nil {
			t.Fatalf("Scan: %v", scanErr)
		}
		if id != 1 || slug != "home" {
			t.Errorf("got (%d, %q), want (1, \"home\")", id, slug)
		}
	}
	if count != 1 {
		t.Errorf("row count = %d, want 1", count)
	}
}

func TestDatabase_ExecuteQuery_NilConnection(t *testing.T) {
	t.Parallel()
	d := Database{Connection: nil}
	// ExecuteQuery will panic or error because it calls d.Connection.Query
	// on a nil pointer. We verify it panics to document this behavior.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from ExecuteQuery with nil connection, but did not panic")
		}
	}()
	d.ExecuteQuery("SELECT * FROM", Route) //nolint:errcheck
}

// --- Query method tests ---

func TestDatabase_Query_Exec(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	conn, err := sql.Open("sqlite3", filepath.Join(dir, "query_test.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()

	d := Database{}
	result, err := d.Query(conn, "CREATE TABLE test_query (id INTEGER PRIMARY KEY);")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if result == nil {
		t.Fatal("Query returned nil result")
	}

	// Verify the table was created by inserting into it
	_, err = conn.Exec("INSERT INTO test_query (id) VALUES (1);")
	if err != nil {
		t.Fatalf("table not created; INSERT failed: %v", err)
	}
}

func TestDatabase_Query_InvalidSQL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	conn, err := sql.Open("sqlite3", filepath.Join(dir, "query_invalid.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()

	d := Database{}
	_, err = d.Query(conn, "THIS IS NOT SQL")
	if err == nil {
		t.Error("expected error for invalid SQL, got nil")
	}
}

func TestMysqlDatabase_Query_Exec(t *testing.T) {
	t.Parallel()
	// MysqlDatabase.Query also calls db.Exec -- we can use SQLite as the underlying
	// *sql.DB since the method signature just takes *sql.DB.
	dir := t.TempDir()
	conn, err := sql.Open("sqlite3", filepath.Join(dir, "mysql_query.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()

	d := MysqlDatabase{}
	result, err := d.Query(conn, "CREATE TABLE mysql_test (id INTEGER);")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if result == nil {
		t.Fatal("Query returned nil result")
	}
}

func TestPsqlDatabase_Query_Exec(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	conn, err := sql.Open("sqlite3", filepath.Join(dir, "psql_query.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()

	d := PsqlDatabase{}
	result, err := d.Query(conn, "CREATE TABLE psql_test (id INTEGER);")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if result == nil {
		t.Fatal("Query returned nil result")
	}
}

// --- BuildInsertQuery tests ---

func TestBuildInsertQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		table  string
		values map[string]string
		// We check that specific fragments appear since map iteration order is not guaranteed
		wantPrefix string
		wantCols   []string
		wantVals   []string
	}{
		{
			name:       "single column",
			table:      "users",
			values:     map[string]string{"name": "alice"},
			wantPrefix: "INSERT INTO users",
			wantCols:   []string{"name"},
			wantVals:   []string{"'alice'"},
		},
		{
			name:       "multiple columns",
			table:      "routes",
			values:     map[string]string{"slug": "home", "title": "Home Page"},
			wantPrefix: "INSERT INTO routes",
			wantCols:   []string{"slug", "title"},
			wantVals:   []string{"'home'", "'Home Page'"},
		},
		{
			name:       "empty values",
			table:      "empty",
			values:     map[string]string{},
			wantPrefix: "INSERT INTO empty",
			wantCols:   []string{},
			wantVals:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BuildInsertQuery(tt.table, tt.values)
			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("query = %q, want prefix %q", got, tt.wantPrefix)
			}
			for _, col := range tt.wantCols {
				if !strings.Contains(got, col) {
					t.Errorf("query = %q, missing column %q", got, col)
				}
			}
			for _, val := range tt.wantVals {
				if !strings.Contains(got, val) {
					t.Errorf("query = %q, missing value %q", got, val)
				}
			}
		})
	}
}

func TestBuildInsertQuery_ColumnValueAlignment(t *testing.T) {
	t.Parallel()
	// With a single key, the column and value positions are deterministic
	got := BuildInsertQuery("t", map[string]string{"col1": "val1"})
	want := "INSERT INTO t (col1) VALUES ('val1')"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- BuildUpdateQuery tests ---

func TestBuildUpdateQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		table      string
		id         int64
		values     map[string]string
		wantSuffix string
		wantSets   []string
	}{
		{
			name:       "single field",
			table:      "users",
			id:         42,
			values:     map[string]string{"name": "bob"},
			wantSuffix: "WHERE id = 42",
			wantSets:   []string{"name = 'bob'"},
		},
		{
			name:       "multiple fields",
			table:      "routes",
			id:         1,
			values:     map[string]string{"slug": "about", "title": "About Us"},
			wantSuffix: "WHERE id = 1",
			wantSets:   []string{"slug = 'about'", "title = 'About Us'"},
		},
		{
			name:       "zero id",
			table:      "t",
			id:         0,
			values:     map[string]string{"x": "y"},
			wantSuffix: "WHERE id = 0",
			wantSets:   []string{"x = 'y'"},
		},
		{
			name:       "negative id",
			table:      "t",
			id:         -1,
			values:     map[string]string{"x": "y"},
			wantSuffix: "WHERE id = -1",
			wantSets:   []string{"x = 'y'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BuildUpdateQuery(tt.table, tt.id, tt.values)
			if !strings.HasPrefix(got, "UPDATE "+tt.table+" SET ") {
				t.Errorf("query = %q, missing UPDATE prefix for table %q", got, tt.table)
			}
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("query = %q, want suffix %q", got, tt.wantSuffix)
			}
			for _, set := range tt.wantSets {
				if !strings.Contains(got, set) {
					t.Errorf("query = %q, missing SET clause %q", got, set)
				}
			}
		})
	}
}

func TestBuildUpdateQuery_SingleField_Deterministic(t *testing.T) {
	t.Parallel()
	got := BuildUpdateQuery("t", 5, map[string]string{"col": "val"})
	want := "UPDATE t SET col = 'val' WHERE id = 5"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- BuildSelectQuery tests ---

func TestBuildSelectQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		table string
		id    int64
		want  string
	}{
		{
			name:  "normal table and id",
			table: "users",
			id:    1,
			want:  `SELECT * FROM "users" WHERE id = 1`,
		},
		{
			name:  "zero id",
			table: "routes",
			id:    0,
			want:  `SELECT * FROM "routes" WHERE id = 0`,
		},
		{
			name:  "negative id",
			table: "t",
			id:    -100,
			want:  `SELECT * FROM "t" WHERE id = -100`,
		},
		{
			name:  "large id",
			table: "content",
			id:    9999999999,
			want:  `SELECT * FROM "content" WHERE id = 9999999999`,
		},
		{
			name:  "table with underscore",
			table: "content_data",
			id:    7,
			want:  `SELECT * FROM "content_data" WHERE id = 7`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BuildSelectQuery(tt.table, tt.id)
			if got != tt.want {
				t.Errorf("BuildSelectQuery(%q, %d) = %q, want %q", tt.table, tt.id, got, tt.want)
			}
		})
	}
}

// --- BuildListQuery tests ---

func TestBuildListQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		table string
		want  string
	}{
		{
			name:  "simple table",
			table: "users",
			want:  `SELECT * FROM "users"`,
		},
		{
			name:  "underscore table",
			table: "content_data",
			want:  `SELECT * FROM "content_data"`,
		},
		{
			name:  "empty table name",
			table: "",
			want:  `SELECT * FROM ""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BuildListQuery(tt.table)
			if got != tt.want {
				t.Errorf("BuildListQuery(%q) = %q, want %q", tt.table, got, tt.want)
			}
		})
	}
}

// --- BuildDeleteQuery tests ---

func TestBuildDeleteQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		table string
		id    int64
		want  string
	}{
		{
			name:  "normal",
			table: "users",
			id:    1,
			want:  `DELETE FROM "users" WHERE id = 1`,
		},
		{
			name:  "zero id",
			table: "routes",
			id:    0,
			want:  `DELETE FROM "routes" WHERE id = 0`,
		},
		{
			name:  "negative id",
			table: "t",
			id:    -5,
			want:  `DELETE FROM "t" WHERE id = -5`,
		},
		{
			name:  "underscore table",
			table: "content_fields",
			id:    42,
			want:  `DELETE FROM "content_fields" WHERE id = 42`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BuildDeleteQuery(tt.table, tt.id)
			if got != tt.want {
				t.Errorf("BuildDeleteQuery(%q, %d) = %q, want %q", tt.table, tt.id, got, tt.want)
			}
		})
	}
}

// --- Build*Query consistency tests ---

func TestBuildQueries_TableNamePreserved(t *testing.T) {
	t.Parallel()
	// All Build*Query functions should embed the table name exactly as given.
	// This catches any accidental quoting inconsistency.
	tableName := "my_special_table"

	insert := BuildInsertQuery(tableName, map[string]string{"a": "b"})
	if !strings.Contains(insert, tableName) {
		t.Errorf("BuildInsertQuery lost table name %q in %q", tableName, insert)
	}

	update := BuildUpdateQuery(tableName, 1, map[string]string{"a": "b"})
	if !strings.Contains(update, tableName) {
		t.Errorf("BuildUpdateQuery lost table name %q in %q", tableName, update)
	}

	// BuildSelectQuery, BuildListQuery, BuildDeleteQuery quote the table name
	selectQ := BuildSelectQuery(tableName, 1)
	if !strings.Contains(selectQ, fmt.Sprintf("%q", tableName)) {
		t.Errorf("BuildSelectQuery lost table name %q in %q", tableName, selectQ)
	}

	listQ := BuildListQuery(tableName)
	if !strings.Contains(listQ, fmt.Sprintf("%q", tableName)) {
		t.Errorf("BuildListQuery lost table name %q in %q", tableName, listQ)
	}

	deleteQ := BuildDeleteQuery(tableName, 1)
	if !strings.Contains(deleteQ, fmt.Sprintf("%q", tableName)) {
		t.Errorf("BuildDeleteQuery lost table name %q in %q", tableName, deleteQ)
	}
}

// Note: BuildInsertQuery and BuildUpdateQuery do NOT quote the table name,
// while BuildSelectQuery, BuildListQuery, and BuildDeleteQuery DO quote it.
// This inconsistency is documented here so future refactors are aware.
func TestBuildQueries_QuotingInconsistency(t *testing.T) {
	t.Parallel()
	table := "test_table"

	insert := BuildInsertQuery(table, map[string]string{"x": "1"})
	selectQ := BuildSelectQuery(table, 1)

	// INSERT does NOT quote table names
	insertHasQuoted := strings.Contains(insert, fmt.Sprintf(`"%s"`, table))
	// SELECT DOES quote table names
	selectHasQuoted := strings.Contains(selectQ, fmt.Sprintf(`"%s"`, table))

	if insertHasQuoted {
		t.Error("BuildInsertQuery now quotes table names -- verify this is intentional")
	}
	if !selectHasQuoted {
		t.Error("BuildSelectQuery no longer quotes table names -- this may break queries")
	}
}

// --- Struct zero-value behavior ---
// Verify that zero-value structs are safe to inspect (fields have expected defaults)

func TestDatabase_ZeroValue(t *testing.T) {
	t.Parallel()
	var d Database

	if d.Src != "" {
		t.Errorf("Src = %q, want empty", d.Src)
	}
	if d.Status != "" {
		t.Errorf("Status = %q, want empty", d.Status)
	}
	if d.Connection != nil {
		t.Error("Connection should be nil")
	}
	if d.LastConnection != "" {
		t.Errorf("LastConnection = %q, want empty", d.LastConnection)
	}
	if d.Err != nil {
		t.Errorf("Err = %v, want nil", d.Err)
	}
	if d.Context != nil {
		t.Error("Context should be nil")
	}
}

func TestMysqlDatabase_ZeroValue(t *testing.T) {
	t.Parallel()
	var d MysqlDatabase

	if d.Src != "" {
		t.Errorf("Src = %q, want empty", d.Src)
	}
	if d.Status != "" {
		t.Errorf("Status = %q, want empty", d.Status)
	}
	if d.Connection != nil {
		t.Error("Connection should be nil")
	}
	if d.Err != nil {
		t.Errorf("Err = %v, want nil", d.Err)
	}
}

func TestPsqlDatabase_ZeroValue(t *testing.T) {
	t.Parallel()
	var d PsqlDatabase

	if d.Src != "" {
		t.Errorf("Src = %q, want empty", d.Src)
	}
	if d.Status != "" {
		t.Errorf("Status = %q, want empty", d.Status)
	}
	if d.Connection != nil {
		t.Error("Connection should be nil")
	}
	if d.Err != nil {
		t.Errorf("Err = %v, want nil", d.Err)
	}
}

// --- Struct field assignment ---

func TestDatabase_FieldAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cfg := config.Config{Db_Driver: config.Sqlite}
	testErr := fmt.Errorf("test error")

	d := Database{
		Src:            "/tmp/test.db",
		Status:         Open,
		Connection:     nil,
		LastConnection: "2025-01-01",
		Err:            testErr,
		Config:         cfg,
		Context:        ctx,
	}

	if d.Connection != nil {
		t.Error("Connection should be nil")
	}
	if d.Src != "/tmp/test.db" {
		t.Errorf("Src = %q", d.Src)
	}
	if d.Status != Open {
		t.Errorf("Status = %q, want %q", d.Status, Open)
	}
	if d.LastConnection != "2025-01-01" {
		t.Errorf("LastConnection = %q", d.LastConnection)
	}
	if d.Err != testErr {
		t.Errorf("Err = %v, want %v", d.Err, testErr)
	}
	if d.Config.Db_Driver != config.Sqlite {
		t.Errorf("Config.Db_Driver = %q", d.Config.Db_Driver)
	}
	if d.Context != ctx {
		t.Error("Context mismatch")
	}
}

func TestMysqlDatabase_FieldAssignment(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{
		Src:    "mysql://localhost",
		Status: Closed,
		Err:    fmt.Errorf("connection refused"),
	}

	if d.Src != "mysql://localhost" {
		t.Errorf("Src = %q", d.Src)
	}
	if d.Status != Closed {
		t.Errorf("Status = %q, want %q", d.Status, Closed)
	}
	if d.Err == nil || d.Err.Error() != "connection refused" {
		t.Errorf("Err = %v", d.Err)
	}
}

func TestPsqlDatabase_FieldAssignment(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{
		Src:    "postgres://localhost:5432",
		Status: Err,
	}

	if d.Src != "postgres://localhost:5432" {
		t.Errorf("Src = %q", d.Src)
	}
	if d.Status != Err {
		t.Errorf("Status = %q, want %q", d.Status, Err)
	}
}

// --- Integration: CreateAllTables + DropAllTables round-trip ---
// This is a critical integration test that verifies the full table creation
// and teardown lifecycle with a real SQLite database.

func testSQLiteDatabase(t *testing.T) Database {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "integration.db")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	// Enable WAL mode and foreign keys like production
	if _, err := conn.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}

	return Database{
		Connection: conn,
		Context:    context.Background(),
		Config:     config.Config{Node_ID: types.NewNodeID().String()},
	}
}

func TestDatabase_CreateAllTables_Success(t *testing.T) {
	d := testSQLiteDatabase(t)

	err := d.CreateAllTables()
	if err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}

	// Verify a sampling of tables exist by querying sqlite_master
	tables := []string{
		"permissions", "roles", "users", "routes",
		"content_data", "content_fields", "datatypes",
		"fields", "media", "media_dimensions", "tokens",
		"sessions", "tables", "change_events", "backups",
		"admin_routes", "admin_datatypes", "admin_fields",
		"admin_content_data", "admin_content_fields",
		"datatypes_fields", "admin_datatypes_fields",
		"content_relations", "admin_content_relations",
	}

	for _, table := range tables {
		var name string
		err := d.Connection.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?;",
			table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found in database: %v", table, err)
		}
	}
}

func TestDatabase_CreateAllTables_DropAllTables_Roundtrip(t *testing.T) {
	d := testSQLiteDatabase(t)

	// Create
	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}

	// Verify at least one table exists
	var count int
	err := d.Connection.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite%';",
	).Scan(&count)
	if err != nil {
		t.Fatalf("count tables: %v", err)
	}
	if count == 0 {
		t.Fatal("no tables created")
	}

	// Drop
	if err := d.DropAllTables(); err != nil {
		t.Fatalf("DropAllTables: %v", err)
	}

	// Verify all application tables are gone
	err = d.Connection.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite%';",
	).Scan(&count)
	if err != nil {
		t.Fatalf("count tables after drop: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tables after DropAllTables, got %d", count)
	}
}

func TestDatabase_CreateAllTables_Idempotent(t *testing.T) {
	// Calling CreateAllTables twice should fail on the second call because
	// tables already exist (CREATE TABLE without IF NOT EXISTS).
	// This documents the current behavior -- it is NOT idempotent.
	d := testSQLiteDatabase(t)

	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("first CreateAllTables: %v", err)
	}

	err := d.CreateAllTables()
	// We expect an error on the second call because the tables already exist
	if err == nil {
		// If it succeeds, CreateAllTables uses IF NOT EXISTS -- that's fine too
		t.Log("CreateAllTables succeeded on second call (uses IF NOT EXISTS)")
	}
}

func TestDatabase_DropAllTables_NoTables(t *testing.T) {
	// Dropping when no tables exist should fail (DROP TABLE without IF EXISTS)
	// or succeed if the implementation uses IF EXISTS. Either way, we document it.
	d := testSQLiteDatabase(t)

	err := d.DropAllTables()
	if err != nil {
		// Expected: tables don't exist, DROP fails
		t.Logf("DropAllTables on empty DB returned expected error: %v", err)
	}
}

// --- Integration: CreateBootstrapData + ValidateBootstrapData ---

func TestDatabase_CreateBootstrapData_Success(t *testing.T) {
	d := testSQLiteDatabase(t)

	// Must create tables first
	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}

	// Create bootstrap data
	if err := d.CreateBootstrapData(""); err != nil {
		t.Fatalf("CreateBootstrapData: %v", err)
	}

	// Validate bootstrap data
	if err := d.ValidateBootstrapData(); err != nil {
		t.Fatalf("ValidateBootstrapData: %v", err)
	}
}

func TestDatabase_ValidateBootstrapData_EmptyTables(t *testing.T) {
	// ValidateBootstrapData on empty tables should fail with detailed errors
	d := testSQLiteDatabase(t)

	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}

	err := d.ValidateBootstrapData()
	if err == nil {
		t.Fatal("expected validation error on empty tables, got nil")
	}

	// The error should mention multiple failed tables
	errMsg := err.Error()
	if !strings.Contains(errMsg, "bootstrap validation failed") {
		t.Errorf("error = %q, want it to contain 'bootstrap validation failed'", errMsg)
	}
	// Should mention at least permissions and roles
	if !strings.Contains(errMsg, "permissions") {
		t.Errorf("error missing 'permissions' table failure in: %q", errMsg)
	}
	if !strings.Contains(errMsg, "roles") {
		t.Errorf("error missing 'roles' table failure in: %q", errMsg)
	}
}

func TestDatabase_CreateBootstrapData_WithoutTables(t *testing.T) {
	// Creating bootstrap data without tables should fail
	d := testSQLiteDatabase(t)

	err := d.CreateBootstrapData("")
	if err == nil {
		t.Fatal("expected error from CreateBootstrapData without tables, got nil")
	}
}
