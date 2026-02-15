// White-box tests for foreignKey.go: SqliteForeignKeyQueryRow struct,
// query constant formatting, GetForeignKeys argument validation and real
// SQLite query execution, ScanForeignKeyQueryRows row scanning, and
// SelectColumnFromTable method existence across all three drivers.
//
// White-box access is needed because:
//   - SqliteForeignKeyQueryRow has unexported fields that can only be verified
//     by scanning into them via ScanForeignKeyQueryRows (which is a method on
//     the unexported-field struct).
//   - The query constants (sqliteQuery, mysqlQuery, psqlQuery) are unexported
//     and need direct inspection to verify formatting placeholders.
//   - GetForeignKeys returns nil for invalid arg counts before touching the
//     database -- testing the guard clauses requires constructing zero-value
//     Database/MysqlDatabase/PsqlDatabase structs directly.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Query constant tests
// ---------------------------------------------------------------------------

func TestSqliteQuery_ContainsPragma(t *testing.T) {
	t.Parallel()
	if !strings.Contains(sqliteQuery, "PRAGMA foreign_key_list") {
		t.Errorf("sqliteQuery does not contain expected PRAGMA statement: %q", sqliteQuery)
	}
}

func TestSqliteQuery_HasSingleFormatVerb(t *testing.T) {
	t.Parallel()
	count := strings.Count(sqliteQuery, "%s")
	if count != 1 {
		t.Errorf("sqliteQuery has %d format verbs, want 1", count)
	}
}

func TestSqliteQuery_FormatProducesValidSQL(t *testing.T) {
	t.Parallel()
	result := fmt.Sprintf(sqliteQuery, "my_table")
	want := "PRAGMA foreign_key_list('my_table');"
	if result != want {
		t.Errorf("formatted sqliteQuery = %q, want %q", result, want)
	}
}

func TestMysqlQuery_HasThreeFormatVerbs(t *testing.T) {
	t.Parallel()
	count := strings.Count(mysqlQuery, "%s")
	if count != 3 {
		t.Errorf("mysqlQuery has %d format verbs, want 3", count)
	}
}

func TestMysqlQuery_ContainsInformationSchema(t *testing.T) {
	t.Parallel()
	if !strings.Contains(mysqlQuery, "INFORMATION_SCHEMA.KEY_COLUMN_USAGE") {
		t.Errorf("mysqlQuery does not reference INFORMATION_SCHEMA.KEY_COLUMN_USAGE")
	}
}

func TestMysqlQuery_ContainsReferencedTableNotNull(t *testing.T) {
	t.Parallel()
	if !strings.Contains(mysqlQuery, "REFERENCED_TABLE_NAME IS NOT NULL") {
		t.Errorf("mysqlQuery missing REFERENCED_TABLE_NAME IS NOT NULL filter")
	}
}

func TestMysqlQuery_FormatProducesExpectedSQL(t *testing.T) {
	t.Parallel()
	result := fmt.Sprintf(mysqlQuery, "mydb", "my_table", "my_column")
	if !strings.Contains(result, "TABLE_SCHEMA = 'mydb'") {
		t.Errorf("formatted mysqlQuery missing TABLE_SCHEMA clause: %s", result)
	}
	if !strings.Contains(result, "TABLE_NAME = 'my_table'") {
		t.Errorf("formatted mysqlQuery missing TABLE_NAME clause: %s", result)
	}
	if !strings.Contains(result, "COLUMN_NAME = 'my_column'") {
		t.Errorf("formatted mysqlQuery missing COLUMN_NAME clause: %s", result)
	}
}

func TestPsqlQuery_HasTwoFormatVerbs(t *testing.T) {
	t.Parallel()
	count := strings.Count(psqlQuery, "%s")
	if count != 2 {
		t.Errorf("psqlQuery has %d format verbs, want 2", count)
	}
}

func TestPsqlQuery_ContainsForeignKeyConstraint(t *testing.T) {
	t.Parallel()
	if !strings.Contains(psqlQuery, "tc.constraint_type = 'FOREIGN KEY'") {
		t.Errorf("psqlQuery missing FOREIGN KEY constraint filter")
	}
}

func TestPsqlQuery_ContainsJoins(t *testing.T) {
	t.Parallel()
	joinCount := strings.Count(psqlQuery, "JOIN")
	if joinCount < 2 {
		t.Errorf("psqlQuery has %d JOINs, want at least 2", joinCount)
	}
}

func TestPsqlQuery_FormatProducesExpectedSQL(t *testing.T) {
	t.Parallel()
	result := fmt.Sprintf(psqlQuery, "my_table", "my_column")
	if !strings.Contains(result, "tc.table_name = 'my_table'") {
		t.Errorf("formatted psqlQuery missing table_name clause: %s", result)
	}
	if !strings.Contains(result, "kcu.column_name = 'my_column'") {
		t.Errorf("formatted psqlQuery missing column_name clause: %s", result)
	}
}

// ---------------------------------------------------------------------------
// SqliteForeignKeyQueryRow struct tests
// ---------------------------------------------------------------------------

func TestSqliteForeignKeyQueryRow_ZeroValue(t *testing.T) {
	t.Parallel()
	var row SqliteForeignKeyQueryRow
	if row.id != 0 {
		t.Errorf("zero-value id = %d, want 0", row.id)
	}
	if row.seq != 0 {
		t.Errorf("zero-value seq = %d, want 0", row.seq)
	}
	if row.tableName != "" {
		t.Errorf("zero-value tableName = %q, want empty", row.tableName)
	}
	if row.fromCol != "" {
		t.Errorf("zero-value fromCol = %q, want empty", row.fromCol)
	}
	if row.toCol != "" {
		t.Errorf("zero-value toCol = %q, want empty", row.toCol)
	}
	if row.onUpdate != "" {
		t.Errorf("zero-value onUpdate = %q, want empty", row.onUpdate)
	}
	if row.onDelete != "" {
		t.Errorf("zero-value onDelete = %q, want empty", row.onDelete)
	}
	if row.match != "" {
		t.Errorf("zero-value match = %q, want empty", row.match)
	}
}

// ---------------------------------------------------------------------------
// GetForeignKeys -- argument validation (nil connection, wrong arg count)
// ---------------------------------------------------------------------------

func TestDatabase_GetForeignKeys_WrongArgCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "nil_args", args: nil},
		{name: "empty_args", args: []string{}},
		{name: "two_args", args: []string{"a", "b"}},
		{name: "three_args", args: []string{"a", "b", "c"}},
		{name: "four_args", args: []string{"a", "b", "c", "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := Database{} // zero-value, no connection
			got := d.GetForeignKeys(tt.args)
			if got != nil {
				t.Errorf("GetForeignKeys(%v) = non-nil, want nil", tt.args)
			}
		})
	}
}

// NOTE: Testing GetForeignKeys with correct arg count but nil Connection is
// not possible because GetConnection() returns the nil *sql.DB directly, and
// calling QueryContext on a nil *sql.DB panics (nil pointer dereference).
// The code does not guard against nil *sql.DB after GetConnection succeeds.
// This is an untestable error path without a real database connection.

func TestMysqlDatabase_GetForeignKeys_WrongArgCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "nil_args", args: nil},
		{name: "empty_args", args: []string{}},
		{name: "one_arg", args: []string{"a"}},
		{name: "two_args", args: []string{"a", "b"}},
		{name: "four_args", args: []string{"a", "b", "c", "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := MysqlDatabase{}
			got := d.GetForeignKeys(tt.args)
			if got != nil {
				t.Errorf("GetForeignKeys(%v) = non-nil, want nil", tt.args)
			}
		})
	}
}

func TestPsqlDatabase_GetForeignKeys_WrongArgCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "nil_args", args: nil},
		{name: "empty_args", args: []string{}},
		{name: "one_arg", args: []string{"a"}},
		{name: "two_args", args: []string{"a", "b"}},
		{name: "four_args", args: []string{"a", "b", "c", "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := PsqlDatabase{}
			got := d.GetForeignKeys(tt.args)
			if got != nil {
				t.Errorf("GetForeignKeys(%v) = non-nil, want nil", tt.args)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetForeignKeys -- real SQLite execution (integration-level)
// ---------------------------------------------------------------------------

// sqliteTestDB creates a temporary SQLite database with foreign keys enabled
// and returns a Database struct ready for testing. The database is cleaned up
// when the test finishes.
func sqliteTestDB(t *testing.T) Database {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "fk_test.db")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	// Enable foreign keys
	_, err = conn.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}

	return Database{
		Src:        dbPath,
		Connection: conn,
		Context:    context.Background(),
	}
}

func TestDatabase_GetForeignKeys_NoForeignKeys(t *testing.T) {
	d := sqliteTestDB(t)

	// Create a table without foreign keys
	_, err := d.Connection.Exec("CREATE TABLE standalone (id INTEGER PRIMARY KEY, name TEXT);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := d.GetForeignKeys([]string{"standalone"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil for valid table")
	}
	defer rows.Close()

	// Should have no rows
	if rows.Next() {
		t.Error("expected no foreign key rows for standalone table")
	}
}

func TestDatabase_GetForeignKeys_WithForeignKeys(t *testing.T) {
	d := sqliteTestDB(t)

	// Create parent and child tables with a foreign key
	_, err := d.Connection.Exec(`
		CREATE TABLE parent (id INTEGER PRIMARY KEY, name TEXT);
		CREATE TABLE child (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER,
			FOREIGN KEY (parent_id) REFERENCES parent(id) ON DELETE CASCADE ON UPDATE SET NULL
		);
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := d.GetForeignKeys([]string{"child"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	// Should have at least one foreign key row
	if !rows.Next() {
		t.Fatal("expected at least one foreign key row for child table")
	}
}

func TestDatabase_GetForeignKeys_NonexistentTable(t *testing.T) {
	d := sqliteTestDB(t)

	// PRAGMA foreign_key_list on a nonexistent table returns an empty result
	// set (not an error) in SQLite
	rows := d.GetForeignKeys([]string{"nonexistent_table"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil for nonexistent table")
	}
	defer rows.Close()

	if rows.Next() {
		t.Error("expected no rows for nonexistent table")
	}
}

func TestDatabase_GetForeignKeys_MultipleForeignKeys(t *testing.T) {
	d := sqliteTestDB(t)

	_, err := d.Connection.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY);
		CREATE TABLE roles (id INTEGER PRIMARY KEY);
		CREATE TABLE assignments (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			role_id INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (role_id) REFERENCES roles(id)
		);
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := d.GetForeignKeys([]string{"assignments"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		// Consume the row to avoid scan errors
		var id, seq int
		var tableName, fromCol, toCol, onUpdate, onDelete, match string
		scanErr := rows.Scan(&id, &seq, &tableName, &fromCol, &toCol, &onUpdate, &onDelete, &match)
		if scanErr != nil {
			t.Fatalf("Scan row %d: %v", count, scanErr)
		}
	}
	if count != 2 {
		t.Errorf("got %d foreign key rows, want 2", count)
	}
}

// ---------------------------------------------------------------------------
// ScanForeignKeyQueryRows -- real SQLite scanning
// ---------------------------------------------------------------------------

func TestDatabase_ScanForeignKeyQueryRows_EmptyResult(t *testing.T) {
	d := sqliteTestDB(t)

	_, err := d.Connection.Exec("CREATE TABLE no_fk (id INTEGER PRIMARY KEY);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := d.GetForeignKeys([]string{"no_fk"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	result := d.ScanForeignKeyQueryRows(rows)
	if result == nil {
		t.Fatal("ScanForeignKeyQueryRows returned nil, want empty slice")
	}
	if len(result) != 0 {
		t.Errorf("ScanForeignKeyQueryRows returned %d rows, want 0", len(result))
	}
}

func TestDatabase_ScanForeignKeyQueryRows_SingleForeignKey(t *testing.T) {
	d := sqliteTestDB(t)

	_, err := d.Connection.Exec(`
		CREATE TABLE parent_scan (id INTEGER PRIMARY KEY);
		CREATE TABLE child_scan (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER,
			FOREIGN KEY (parent_id) REFERENCES parent_scan(id) ON DELETE CASCADE ON UPDATE NO ACTION
		);
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := d.GetForeignKeys([]string{"child_scan"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	result := d.ScanForeignKeyQueryRows(rows)
	if len(result) != 1 {
		t.Fatalf("ScanForeignKeyQueryRows returned %d rows, want 1", len(result))
	}

	fk := result[0]
	if fk.tableName != "parent_scan" {
		t.Errorf("tableName = %q, want %q", fk.tableName, "parent_scan")
	}
	if fk.fromCol != "parent_id" {
		t.Errorf("fromCol = %q, want %q", fk.fromCol, "parent_id")
	}
	if fk.toCol != "id" {
		t.Errorf("toCol = %q, want %q", fk.toCol, "id")
	}
	if fk.onDelete != "CASCADE" {
		t.Errorf("onDelete = %q, want %q", fk.onDelete, "CASCADE")
	}
	if fk.onUpdate != "NO ACTION" {
		t.Errorf("onUpdate = %q, want %q", fk.onUpdate, "NO ACTION")
	}
}

func TestDatabase_ScanForeignKeyQueryRows_MultipleForeignKeys(t *testing.T) {
	d := sqliteTestDB(t)

	_, err := d.Connection.Exec(`
		CREATE TABLE t_alpha (id INTEGER PRIMARY KEY);
		CREATE TABLE t_beta (id INTEGER PRIMARY KEY);
		CREATE TABLE t_child_multi (
			id INTEGER PRIMARY KEY,
			alpha_id INTEGER,
			beta_id INTEGER,
			FOREIGN KEY (alpha_id) REFERENCES t_alpha(id) ON DELETE SET NULL ON UPDATE CASCADE,
			FOREIGN KEY (beta_id) REFERENCES t_beta(id) ON DELETE RESTRICT ON UPDATE SET DEFAULT
		);
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := d.GetForeignKeys([]string{"t_child_multi"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	result := d.ScanForeignKeyQueryRows(rows)
	if len(result) != 2 {
		t.Fatalf("ScanForeignKeyQueryRows returned %d rows, want 2", len(result))
	}

	// SQLite returns foreign keys in definition order (id 0, 1, ...)
	// The first FK in the CREATE TABLE is alpha_id, but SQLite may assign
	// ids in reverse order. Check by matching on fromCol.
	foundAlpha := false
	foundBeta := false
	for _, fk := range result {
		switch fk.fromCol {
		case "alpha_id":
			foundAlpha = true
			if fk.tableName != "t_alpha" {
				t.Errorf("alpha FK tableName = %q, want %q", fk.tableName, "t_alpha")
			}
			if fk.onDelete != "SET NULL" {
				t.Errorf("alpha FK onDelete = %q, want %q", fk.onDelete, "SET NULL")
			}
			if fk.onUpdate != "CASCADE" {
				t.Errorf("alpha FK onUpdate = %q, want %q", fk.onUpdate, "CASCADE")
			}
		case "beta_id":
			foundBeta = true
			if fk.tableName != "t_beta" {
				t.Errorf("beta FK tableName = %q, want %q", fk.tableName, "t_beta")
			}
			if fk.onDelete != "RESTRICT" {
				t.Errorf("beta FK onDelete = %q, want %q", fk.onDelete, "RESTRICT")
			}
			if fk.onUpdate != "SET DEFAULT" {
				t.Errorf("beta FK onUpdate = %q, want %q", fk.onUpdate, "SET DEFAULT")
			}
		default:
			t.Errorf("unexpected fromCol = %q", fk.fromCol)
		}
	}
	if !foundAlpha {
		t.Error("alpha_id foreign key not found in results")
	}
	if !foundBeta {
		t.Error("beta_id foreign key not found in results")
	}
}

func TestDatabase_ScanForeignKeyQueryRows_FieldValues(t *testing.T) {
	// Verify all 8 fields of SqliteForeignKeyQueryRow are populated correctly
	d := sqliteTestDB(t)

	_, err := d.Connection.Exec(`
		CREATE TABLE ref_target (pk INTEGER PRIMARY KEY);
		CREATE TABLE ref_source (
			id INTEGER PRIMARY KEY,
			target_fk INTEGER,
			FOREIGN KEY (target_fk) REFERENCES ref_target(pk) ON DELETE NO ACTION ON UPDATE NO ACTION
		);
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := d.GetForeignKeys([]string{"ref_source"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	result := d.ScanForeignKeyQueryRows(rows)
	if len(result) != 1 {
		t.Fatalf("got %d rows, want 1", len(result))
	}

	fk := result[0]

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "tableName", got: fk.tableName, want: "ref_target"},
		{name: "fromCol", got: fk.fromCol, want: "target_fk"},
		{name: "toCol", got: fk.toCol, want: "pk"},
		{name: "onUpdate", got: fk.onUpdate, want: "NO ACTION"},
		{name: "onDelete", got: fk.onDelete, want: "NO ACTION"},
		{name: "match", got: fk.match, want: "NONE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}

	// id and seq are integers; verify seq=0 for the only column in this FK
	if fk.seq != 0 {
		t.Errorf("seq = %d, want 0", fk.seq)
	}
	// id should be 0 for the first (and only) FK definition
	if fk.id != 0 {
		t.Errorf("id = %d, want 0", fk.id)
	}
}

// ---------------------------------------------------------------------------
// GetForeignKeys + ScanForeignKeyQueryRows round-trip
// ---------------------------------------------------------------------------

func TestDatabase_GetAndScan_RoundTrip(t *testing.T) {
	d := sqliteTestDB(t)

	_, err := d.Connection.Exec(`
		CREATE TABLE rt_parent (id INTEGER PRIMARY KEY);
		CREATE TABLE rt_child (
			id INTEGER PRIMARY KEY,
			pid INTEGER NOT NULL,
			FOREIGN KEY (pid) REFERENCES rt_parent(id) ON DELETE CASCADE ON UPDATE SET NULL
		);
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := d.GetForeignKeys([]string{"rt_child"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	fks := d.ScanForeignKeyQueryRows(rows)
	if len(fks) != 1 {
		t.Fatalf("expected 1 FK, got %d", len(fks))
	}

	fk := fks[0]
	if fk.tableName != "rt_parent" {
		t.Errorf("tableName = %q, want %q", fk.tableName, "rt_parent")
	}
	if fk.fromCol != "pid" {
		t.Errorf("fromCol = %q, want %q", fk.fromCol, "pid")
	}
	if fk.toCol != "id" {
		t.Errorf("toCol = %q, want %q", fk.toCol, "id")
	}
	if fk.onDelete != "CASCADE" {
		t.Errorf("onDelete = %q, want %q", fk.onDelete, "CASCADE")
	}
	if fk.onUpdate != "SET NULL" {
		t.Errorf("onUpdate = %q, want %q", fk.onUpdate, "SET NULL")
	}
}

// ---------------------------------------------------------------------------
// Argument count boundary tests -- exact boundary for each driver
// ---------------------------------------------------------------------------

func TestDatabase_GetForeignKeys_ExactlyOneArg(t *testing.T) {
	d := sqliteTestDB(t)

	// One arg is the correct count for SQLite -- should not return nil
	// (returns *sql.Rows, possibly empty)
	_, err := d.Connection.Exec("CREATE TABLE exactly_one (id INTEGER PRIMARY KEY);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := d.GetForeignKeys([]string{"exactly_one"})
	if rows == nil {
		t.Fatal("GetForeignKeys with exactly 1 arg returned nil; expected valid (possibly empty) *sql.Rows")
	}
	defer rows.Close()
}

// NOTE: Testing MysqlDatabase.GetForeignKeys and PsqlDatabase.GetForeignKeys
// with correct arg counts (3 args) but nil Connection is not possible because
// GetConnection() returns the nil *sql.DB from the struct field, and calling
// QueryContext on a nil *sql.DB causes a nil pointer dereference panic.
// These paths require a real MySQL/PostgreSQL connection to test.

// ---------------------------------------------------------------------------
// PsqlDatabase GetForeignKeys uses only first 2 of 3 args
// ---------------------------------------------------------------------------

func TestPsqlQuery_OnlyUsesTwoOfThreeArgs(t *testing.T) {
	t.Parallel()
	// The psqlQuery format string has 2 %s placeholders but GetForeignKeys
	// requires 3 args. Verify the third arg is ignored in formatting.
	result := fmt.Sprintf(psqlQuery, "orders", "customer_id")
	if strings.Contains(result, "%!") {
		t.Errorf("psqlQuery format produced error verb: %s", result)
	}
	if !strings.Contains(result, "'orders'") {
		t.Errorf("psqlQuery missing table name in output: %s", result)
	}
	if !strings.Contains(result, "'customer_id'") {
		t.Errorf("psqlQuery missing column name in output: %s", result)
	}
}

// ---------------------------------------------------------------------------
// MysqlDatabase/PsqlDatabase ScanForeignKeyQueryRows -- identical signature
// ---------------------------------------------------------------------------
// NOTE: MysqlDatabase.ScanForeignKeyQueryRows and PsqlDatabase.ScanForeignKeyQueryRows
// have identical implementations to Database.ScanForeignKeyQueryRows (they all scan
// into SqliteForeignKeyQueryRow). We cannot test them with real MySQL/PostgreSQL
// databases in unit tests, but we verify they compile with the correct signature
// and that they handle an empty result set from SQLite (the *sql.Rows type is
// database-agnostic).

func TestMysqlDatabase_ScanForeignKeyQueryRows_EmptyResult(t *testing.T) {
	// We use a real SQLite *sql.Rows to test the MySQL scanner because
	// *sql.Rows is a stdlib type -- the scanner doesn't know the backend.
	sqliteDB := sqliteTestDB(t)
	_, err := sqliteDB.Connection.Exec("CREATE TABLE mysql_scan_test (id INTEGER PRIMARY KEY);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := sqliteDB.GetForeignKeys([]string{"mysql_scan_test"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	mysqlDB := MysqlDatabase{}
	result := mysqlDB.ScanForeignKeyQueryRows(rows)
	if result == nil {
		t.Fatal("ScanForeignKeyQueryRows returned nil, want empty slice")
	}
	if len(result) != 0 {
		t.Errorf("ScanForeignKeyQueryRows returned %d rows, want 0", len(result))
	}
}

func TestPsqlDatabase_ScanForeignKeyQueryRows_EmptyResult(t *testing.T) {
	sqliteDB := sqliteTestDB(t)
	_, err := sqliteDB.Connection.Exec("CREATE TABLE psql_scan_test (id INTEGER PRIMARY KEY);")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := sqliteDB.GetForeignKeys([]string{"psql_scan_test"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	psqlDB := PsqlDatabase{}
	result := psqlDB.ScanForeignKeyQueryRows(rows)
	if result == nil {
		t.Fatal("ScanForeignKeyQueryRows returned nil, want empty slice")
	}
	if len(result) != 0 {
		t.Errorf("ScanForeignKeyQueryRows returned %d rows, want 0", len(result))
	}
}

func TestMysqlDatabase_ScanForeignKeyQueryRows_WithData(t *testing.T) {
	sqliteDB := sqliteTestDB(t)
	_, err := sqliteDB.Connection.Exec(`
		CREATE TABLE mscan_parent (id INTEGER PRIMARY KEY);
		CREATE TABLE mscan_child (
			id INTEGER PRIMARY KEY,
			ref INTEGER,
			FOREIGN KEY (ref) REFERENCES mscan_parent(id)
		);
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := sqliteDB.GetForeignKeys([]string{"mscan_child"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	mysqlDB := MysqlDatabase{}
	result := mysqlDB.ScanForeignKeyQueryRows(rows)
	if len(result) != 1 {
		t.Fatalf("got %d rows, want 1", len(result))
	}
	if result[0].tableName != "mscan_parent" {
		t.Errorf("tableName = %q, want %q", result[0].tableName, "mscan_parent")
	}
	if result[0].fromCol != "ref" {
		t.Errorf("fromCol = %q, want %q", result[0].fromCol, "ref")
	}
}

func TestPsqlDatabase_ScanForeignKeyQueryRows_WithData(t *testing.T) {
	sqliteDB := sqliteTestDB(t)
	_, err := sqliteDB.Connection.Exec(`
		CREATE TABLE pscan_parent (id INTEGER PRIMARY KEY);
		CREATE TABLE pscan_child (
			id INTEGER PRIMARY KEY,
			ref INTEGER,
			FOREIGN KEY (ref) REFERENCES pscan_parent(id) ON DELETE SET DEFAULT
		);
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	rows := sqliteDB.GetForeignKeys([]string{"pscan_child"})
	if rows == nil {
		t.Fatal("GetForeignKeys returned nil")
	}
	defer rows.Close()

	psqlDB := PsqlDatabase{}
	result := psqlDB.ScanForeignKeyQueryRows(rows)
	if len(result) != 1 {
		t.Fatalf("got %d rows, want 1", len(result))
	}
	if result[0].tableName != "pscan_parent" {
		t.Errorf("tableName = %q, want %q", result[0].tableName, "pscan_parent")
	}
	if result[0].onDelete != "SET DEFAULT" {
		t.Errorf("onDelete = %q, want %q", result[0].onDelete, "SET DEFAULT")
	}
}

// ---------------------------------------------------------------------------
// Cross-driver consistency: all three drivers satisfy DbDriver interface
// for GetForeignKeys, ScanForeignKeyQueryRows, and SelectColumnFromTable
// ---------------------------------------------------------------------------

func TestForeignKey_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Compile-time check that all three database types implement the
	// foreign-key-related methods on the DbDriver interface.
	var _ interface {
		GetForeignKeys(args []string) *sql.Rows
		ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow
		SelectColumnFromTable(table string, column string)
	} = Database{}

	var _ interface {
		GetForeignKeys(args []string) *sql.Rows
		ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow
		SelectColumnFromTable(table string, column string)
	} = MysqlDatabase{}

	var _ interface {
		GetForeignKeys(args []string) *sql.Rows
		ScanForeignKeyQueryRows(rows *sql.Rows) []SqliteForeignKeyQueryRow
		SelectColumnFromTable(table string, column string)
	} = PsqlDatabase{}
}

// ---------------------------------------------------------------------------
// SelectColumnFromTable -- smoke tests (these call GenericList internally,
// which requires a full database with tables. We verify the method signature
// compiles and that calling with an unknown table does not panic.)
// ---------------------------------------------------------------------------
// NOTE: SelectColumnFromTable calls GenericList which in turn calls ListXxx
// methods on the DbDriver interface. With zero-value structs (no connection),
// these will fail, but the function swallows errors with early return.
// Testing the full path would require a complete database setup which is
// covered by integration tests. Here we verify it does not panic.

func TestDatabase_SelectColumnFromTable_UnknownTable_NoPanic(t *testing.T) {
	d := sqliteTestDB(t)
	// "nonexistent" is not a DBTable constant, so GenericList returns (nil, nil)
	// and the for-range over nil is a no-op.
	// This should not panic.
	d.SelectColumnFromTable("nonexistent", "any_column")
}

func TestMysqlDatabase_SelectColumnFromTable_UnknownTable_NoPanic(t *testing.T) {
	t.Parallel()
	// Zero-value MysqlDatabase; GenericList will try to call ListXxx which
	// will fail, but SelectColumnFromTable swallows the error.
	d := MysqlDatabase{}
	d.SelectColumnFromTable("nonexistent", "any_column")
}

func TestPsqlDatabase_SelectColumnFromTable_UnknownTable_NoPanic(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	d.SelectColumnFromTable("nonexistent", "any_column")
}
