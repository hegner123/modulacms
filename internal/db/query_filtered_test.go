package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

// setupFilteredTestDB creates an in-memory SQLite database with a test table and seed data.
func setupFilteredTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`CREATE TABLE test_items (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		status TEXT NOT NULL,
		priority INTEGER NOT NULL,
		email TEXT,
		date_created TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	rows := []struct {
		id, name, status string
		priority         int
		email            *string
		date             string
	}{
		{"1", "alpha", "active", 5, strPtr("a@test.com"), "2024-01-01"},
		{"2", "beta", "draft", 3, nil, "2024-02-01"},
		{"3", "gamma", "active", 8, strPtr("g@test.com"), "2024-03-01"},
		{"4", "delta", "deleted", 1, nil, "2024-04-01"},
		{"5", "epsilon", "active", 10, strPtr("e@test.com"), "2024-05-01"},
	}
	for _, r := range rows {
		_, err := db.Exec(`INSERT INTO test_items (id, name, status, priority, email, date_created) VALUES (?, ?, ?, ?, ?, ?)`,
			r.id, r.name, r.status, r.priority, r.email, r.date)
		if err != nil {
			t.Fatalf("insert seed row %s: %v", r.id, err)
		}
	}
	return db
}

// ===== QSelectFiltered =====

func TestQSelectFiltered_Compare(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: Compare{Column: "status", Op: OpEq, Value: "active"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}
}

func TestQSelectFiltered_In(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: InCondition{Column: "status", Values: []any{"active", "draft"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 4 {
		t.Errorf("expected 4 rows, got %d", len(rows))
	}
}

func TestQSelectFiltered_Between(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: BetweenCondition{Column: "priority", Low: 3, High: 8},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows (priority 3,5,8), got %d", len(rows))
	}
}

func TestQSelectFiltered_IsNull(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: IsNullCondition{Column: "email"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows with null email, got %d", len(rows))
	}
}

func TestQSelectFiltered_IsNotNull(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: IsNotNullCondition{Column: "email"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows with non-null email, got %d", len(rows))
	}
}

func TestQSelectFiltered_And(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table: "test_items",
		Filter: And{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "active"},
			Compare{Column: "priority", Op: OpGt, Value: 5},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows (gamma=8, epsilon=10), got %d", len(rows))
	}
}

func TestQSelectFiltered_Or(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table: "test_items",
		Filter: Or{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "deleted"},
			Compare{Column: "priority", Op: OpGte, Value: 10},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// delta(deleted) + epsilon(priority=10)
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

func TestQSelectFiltered_Not(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: Not{Condition: Compare{Column: "status", Op: OpEq, Value: "active"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 non-active rows, got %d", len(rows))
	}
}

func TestQSelectFiltered_Distinct(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:    "test_items",
		Columns:  []string{"status"},
		Filter:   Compare{Column: "priority", Op: OpGt, Value: 0},
		Distinct: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3 distinct statuses: active, draft, deleted
	if len(rows) != 3 {
		t.Errorf("expected 3 distinct statuses, got %d", len(rows))
	}
}

func TestQSelectFiltered_MultipleOrderBy(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: Compare{Column: "priority", Op: OpGt, Value: 0},
		OrderByCols: []OrderByColumn{
			{Column: "status", Desc: false},
			{Column: "priority", Desc: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(rows))
	}
	// First row should be active with highest priority (epsilon, priority=10)
	if rows[0]["name"] != "epsilon" {
		t.Errorf("expected first row to be epsilon (active, priority=10), got %v", rows[0]["name"])
	}
}

func TestQSelectFiltered_RejectsExcessiveOrderBy(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	cols := make([]OrderByColumn, MaxOrderByCols+1)
	for i := range cols {
		cols[i] = OrderByColumn{Column: "name"}
	}

	_, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:       "test_items",
		Filter:      Compare{Column: "priority", Op: OpGt, Value: 0},
		OrderByCols: cols,
	})
	if err == nil {
		t.Fatal("expected error for too many order_by columns")
	}
	if !strings.Contains(err.Error(), "too many order_by") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestQSelectFiltered_Columns(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:   "test_items",
		Columns: []string{"name", "status"},
		Filter:  Compare{Column: "id", Op: OpEq, Value: "1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if _, ok := rows[0]["id"]; ok {
		t.Error("expected id column to not be in result when only name,status selected")
	}
	if rows[0]["name"] != "alpha" {
		t.Errorf("expected name=alpha, got %v", rows[0]["name"])
	}
}

func TestQSelectFiltered_NilFilter(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil filter")
	}
}

func TestQSelectFiltered_LimitAndOffset(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:       "test_items",
		Filter:      Compare{Column: "priority", Op: OpGt, Value: 0},
		OrderByCols: []OrderByColumn{{Column: "priority", Desc: false}},
		Limit:       2,
		Offset:      1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

// ===== QSelectOneFiltered =====

func TestQSelectOneFiltered(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	row, err := QSelectOneFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: Compare{Column: "id", Op: OpEq, Value: "3"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row == nil {
		t.Fatal("expected non-nil row")
	}
	if row["name"] != "gamma" {
		t.Errorf("expected name=gamma, got %v", row["name"])
	}
}

func TestQSelectOneFiltered_NoMatch(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	row, err := QSelectOneFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: Compare{Column: "id", Op: OpEq, Value: "nonexistent"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row != nil {
		t.Errorf("expected nil row for no match, got %v", row)
	}
}

// ===== QUpdateFiltered =====

func TestQUpdateFiltered_CompoundCondition(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	result, err := QUpdateFiltered(ctx, db, DialectSQLite, FilteredUpdateParams{
		Table: "test_items",
		Set:   map[string]any{"status": "archived"},
		Filter: And{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "active"},
			Compare{Column: "priority", Op: OpLte, Value: 5},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("rows affected: %v", err)
	}
	// alpha (active, priority=5) should be updated
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}

	// Verify the change
	row, err := QSelectOneFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: Compare{Column: "id", Op: OpEq, Value: "1"},
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if row["status"] != "archived" {
		t.Errorf("expected status=archived, got %v", row["status"])
	}
}

func TestQUpdateFiltered_PostgresPlaceholderOffset(t *testing.T) {
	// This test verifies SQL generation only (no real Postgres connection).
	// We use buildFilteredSelectQuery-style validation by building the SQL manually.
	p := FilteredUpdateParams{
		Table: "test_items",
		Set:   map[string]any{"priority": 99, "status": "updated"},
		Filter: And{Conditions: []Condition{
			Compare{Column: "name", Op: OpEq, Value: "alpha"},
			Compare{Column: "email", Op: OpNeq, Value: "nope"},
		}},
	}

	// Simulate the SQL generation path for Postgres
	setKeys := sortedKeys(p.Set)
	var setClauses []string
	argIdx := 1
	for _, k := range setKeys {
		setClauses = append(setClauses, fmt.Sprintf(`"%s" = $%d`, k, argIdx))
		argIdx++
	}

	bctx := NewBuildContext()
	whereSQL, _, nextOffset, err := p.Filter.Build(bctx, DialectPostgres, argIdx)
	if err != nil {
		t.Fatalf("filter build: %v", err)
	}

	query := fmt.Sprintf(`UPDATE "test_items" SET %s WHERE %s`, strings.Join(setClauses, ", "), whereSQL)

	// SET uses $1 (priority), $2 (status); WHERE uses $3 (name), $4 (email)
	wantQuery := `UPDATE "test_items" SET "priority" = $1, "status" = $2 WHERE ("name" = $3 AND "email" <> $4)`
	if query != wantQuery {
		t.Errorf("query = %q, want %q", query, wantQuery)
	}
	if nextOffset != 5 {
		t.Errorf("nextOffset = %d, want 5", nextOffset)
	}
}

func TestQUpdateFiltered_RejectsVacuousCondition(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QUpdateFiltered(ctx, db, DialectSQLite, FilteredUpdateParams{
		Table:  "test_items",
		Set:    map[string]any{"status": "bad"},
		Filter: IsNotNullCondition{Column: "id"},
	})
	if err == nil {
		t.Fatal("expected error for vacuous condition")
	}
	if !strings.Contains(err.Error(), "value binding") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestQUpdateFiltered_NilFilter(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QUpdateFiltered(ctx, db, DialectSQLite, FilteredUpdateParams{
		Table:  "test_items",
		Set:    map[string]any{"status": "bad"},
		Filter: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil filter")
	}
}

func TestQUpdateFiltered_EmptySet(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QUpdateFiltered(ctx, db, DialectSQLite, FilteredUpdateParams{
		Table:  "test_items",
		Set:    map[string]any{},
		Filter: Compare{Column: "id", Op: OpEq, Value: "1"},
	})
	if err == nil {
		t.Fatal("expected error for empty set")
	}
}

func TestQUpdateFiltered_SetNull(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QUpdateFiltered(ctx, db, DialectSQLite, FilteredUpdateParams{
		Table:  "test_items",
		Set:    map[string]any{"email": nil},
		Filter: Compare{Column: "id", Op: OpEq, Value: "1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	row, err := QSelectOneFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: Compare{Column: "id", Op: OpEq, Value: "1"},
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if row["email"] != nil {
		t.Errorf("expected email=nil, got %v", row["email"])
	}
}

// ===== QDeleteFiltered =====

func TestQDeleteFiltered_CompoundCondition(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	result, err := QDeleteFiltered(ctx, db, DialectSQLite, FilteredDeleteParams{
		Table: "test_items",
		Filter: And{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "deleted"},
			Compare{Column: "priority", Op: OpLt, Value: 5},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("rows affected: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row deleted, got %d", affected)
	}

	// Verify 4 remaining
	allRows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: Compare{Column: "priority", Op: OpGt, Value: -1},
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if len(allRows) != 4 {
		t.Errorf("expected 4 remaining rows, got %d", len(allRows))
	}
}

func TestQDeleteFiltered_RejectsVacuousCondition(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QDeleteFiltered(ctx, db, DialectSQLite, FilteredDeleteParams{
		Table:  "test_items",
		Filter: IsNotNullCondition{Column: "id"},
	})
	if err == nil {
		t.Fatal("expected error for vacuous condition")
	}
	if !strings.Contains(err.Error(), "value binding") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestQDeleteFiltered_NilFilter(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QDeleteFiltered(ctx, db, DialectSQLite, FilteredDeleteParams{
		Table:  "test_items",
		Filter: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil filter")
	}
}

// ===== QCountFiltered =====

func TestQCountFiltered(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	count, err := QCountFiltered(ctx, db, DialectSQLite, "test_items",
		Compare{Column: "status", Op: OpEq, Value: "active"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count=3, got %d", count)
	}
}

func TestQCountFiltered_NilFilter(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QCountFiltered(ctx, db, DialectSQLite, "test_items", nil)
	if err == nil {
		t.Fatal("expected error for nil filter")
	}
}

// ===== QExistsFiltered =====

func TestQExistsFiltered_True(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	exists, err := QExistsFiltered(ctx, db, DialectSQLite, "test_items",
		Compare{Column: "name", Op: OpEq, Value: "alpha"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
}

func TestQExistsFiltered_False(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	exists, err := QExistsFiltered(ctx, db, DialectSQLite, "test_items",
		Compare{Column: "name", Op: OpEq, Value: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected exists=false")
	}
}

func TestQExistsFiltered_NilFilter(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QExistsFiltered(ctx, db, DialectSQLite, "test_items", nil)
	if err == nil {
		t.Fatal("expected error for nil filter")
	}
}

// ===== QBulkInsert =====

func TestQBulkInsert_SingleBatch(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	result, err := QBulkInsert(ctx, db, DialectSQLite, BulkInsertParams{
		Table:   "test_items",
		Columns: []string{"id", "name", "status", "priority", "email", "date_created"},
		Rows: [][]any{
			{"10", "zeta", "active", 20, "z@test.com", "2024-06-01"},
			{"11", "eta", "draft", 15, nil, "2024-07-01"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("rows affected: %v", err)
	}
	if affected != 2 {
		t.Errorf("expected 2 rows affected, got %d", affected)
	}

	// Verify inserted
	count, err := QCountFiltered(ctx, db, DialectSQLite, "test_items",
		Compare{Column: "priority", Op: OpGt, Value: -1})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 7 {
		t.Errorf("expected 7 total rows, got %d", count)
	}
}

func TestQBulkInsert_NilValuesAsNull(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QBulkInsert(ctx, db, DialectSQLite, BulkInsertParams{
		Table:   "test_items",
		Columns: []string{"id", "name", "status", "priority", "email", "date_created"},
		Rows: [][]any{
			{"20", "theta", "active", 1, nil, "2024-08-01"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	row, err := QSelectOneFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:  "test_items",
		Filter: Compare{Column: "id", Op: OpEq, Value: "20"},
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if row["email"] != nil {
		t.Errorf("expected email=nil (SQL NULL), got %v", row["email"])
	}
}

func TestQBulkInsert_MultiBatch(t *testing.T) {
	memdb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { memdb.Close() })

	_, err = memdb.Exec(`CREATE TABLE bulk_test (id TEXT PRIMARY KEY, val TEXT NOT NULL)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	// Create enough rows to force multiple batches on SQLite (2 cols: batch = min(999/2, 100) = 100)
	numRows := 250
	rows := make([][]any, numRows)
	for i := range rows {
		rows[i] = []any{fmt.Sprintf("id_%d", i), fmt.Sprintf("val_%d", i)}
	}

	ctx := context.Background()
	result, err := QBulkInsert(ctx, memdb, DialectSQLite, BulkInsertParams{
		Table:   "bulk_test",
		Columns: []string{"id", "val"},
		Rows:    rows,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("rows affected: %v", err)
	}
	if affected != int64(numRows) {
		t.Errorf("expected %d rows affected, got %d", numRows, affected)
	}

	// Verify all rows exist
	var count int64
	err = memdb.QueryRow("SELECT COUNT(*) FROM bulk_test").Scan(&count)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != int64(numRows) {
		t.Errorf("expected %d rows, got %d", numRows, count)
	}
}

func TestQBulkInsert_MultiBatchRollsBackOnFailure(t *testing.T) {
	memdb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { memdb.Close() })

	_, err = memdb.Exec(`CREATE TABLE rollback_test (id TEXT PRIMARY KEY, val TEXT NOT NULL)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	// Create 200 rows; insert first 100 successfully, then duplicate ID in second batch
	rows := make([][]any, 200)
	for i := range 100 {
		rows[i] = []any{fmt.Sprintf("id_%d", i), fmt.Sprintf("val_%d", i)}
	}
	// Second batch starts at 100 but uses a duplicate ID from first batch
	rows[100] = []any{"id_0", "duplicate"} // will fail (PRIMARY KEY conflict)
	for i := 101; i < 200; i++ {
		rows[i] = []any{fmt.Sprintf("id_%d", i), fmt.Sprintf("val_%d", i)}
	}

	ctx := context.Background()
	_, err = QBulkInsert(ctx, memdb, DialectSQLite, BulkInsertParams{
		Table:   "rollback_test",
		Columns: []string{"id", "val"},
		Rows:    rows,
	})
	if err == nil {
		t.Fatal("expected error from duplicate key")
	}

	// Verify rollback: table should be empty (transaction rolled back)
	var count int64
	scanErr := memdb.QueryRow("SELECT COUNT(*) FROM rollback_test").Scan(&count)
	if scanErr != nil {
		t.Fatalf("count: %v", scanErr)
	}
	if count != 0 {
		t.Errorf("expected 0 rows after rollback, got %d", count)
	}
}

func TestQBulkInsert_RejectsExcessiveRows(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	rows := make([][]any, MaxBulkInsertRows+1)
	for i := range rows {
		rows[i] = []any{fmt.Sprintf("id_%d", i+100), "x", "active", 1, nil, "2024-01-01"}
	}

	_, err := QBulkInsert(ctx, db, DialectSQLite, BulkInsertParams{
		Table:   "test_items",
		Columns: []string{"id", "name", "status", "priority", "email", "date_created"},
		Rows:    rows,
	})
	if err == nil {
		t.Fatal("expected error for too many rows")
	}
	if !strings.Contains(err.Error(), "too many rows") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestQBulkInsert_RejectsEmptyRows(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QBulkInsert(ctx, db, DialectSQLite, BulkInsertParams{
		Table:   "test_items",
		Columns: []string{"id", "name"},
		Rows:    [][]any{},
	})
	if err == nil {
		t.Fatal("expected error for empty rows")
	}
}

func TestQBulkInsert_RejectsEmptyColumns(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QBulkInsert(ctx, db, DialectSQLite, BulkInsertParams{
		Table:   "test_items",
		Columns: []string{},
		Rows:    [][]any{{"x"}},
	})
	if err == nil {
		t.Fatal("expected error for empty columns")
	}
}

func TestQBulkInsert_RejectsMismatchedRowLength(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QBulkInsert(ctx, db, DialectSQLite, BulkInsertParams{
		Table:   "test_items",
		Columns: []string{"id", "name"},
		Rows:    [][]any{{"1", "x", "extra"}},
	})
	if err == nil {
		t.Fatal("expected error for mismatched row length")
	}
	if !strings.Contains(err.Error(), "values, expected") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestQBulkInsert_RejectsInvalidColumn(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QBulkInsert(ctx, db, DialectSQLite, BulkInsertParams{
		Table:   "test_items",
		Columns: []string{"id", "DROP"},
		Rows:    [][]any{{"1", "x"}},
	})
	if err == nil {
		t.Fatal("expected error for invalid column name")
	}
}

func TestQBulkInsert_RejectsInvalidTable(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	_, err := QBulkInsert(ctx, db, DialectSQLite, BulkInsertParams{
		Table:   "DROP",
		Columns: []string{"id"},
		Rows:    [][]any{{"1"}},
	})
	if err == nil {
		t.Fatal("expected error for invalid table name")
	}
}

// ===== Batch Sizing =====

func TestBulkBatchSize_SQLite(t *testing.T) {
	tests := []struct {
		cols int
		want int
	}{
		{2, 100},   // 999/2 = 499, capped at 100
		{10, 99},   // 999/10 = 99
		{64, 15},   // 999/64 = 15
		{100, 9},   // 999/100 = 9
		{999, 1},   // 999/999 = 1
		{1000, 0},  // 999/1000 = 0 → clamped to 1
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("cols=%d", tt.cols), func(t *testing.T) {
			got := bulkBatchSize(DialectSQLite, tt.cols)
			want := tt.want
			if tt.cols >= 1000 {
				want = 1 // clamped
			}
			if got != want {
				t.Errorf("bulkBatchSize(SQLite, %d) = %d, want %d", tt.cols, got, want)
			}
		})
	}
}

func TestBulkBatchSize_MySQL(t *testing.T) {
	got := bulkBatchSize(DialectMySQL, 10)
	if got != 500 {
		t.Errorf("bulkBatchSize(MySQL, 10) = %d, want 500", got)
	}
}

func TestBulkBatchSize_Postgres(t *testing.T) {
	got := bulkBatchSize(DialectPostgres, 10)
	if got != 1000 {
		t.Errorf("bulkBatchSize(Postgres, 10) = %d, want 1000", got)
	}
}

// ===== QBulkInsert with *sql.Tx =====

func TestQBulkInsert_WithExistingTransaction(t *testing.T) {
	memdb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { memdb.Close() })

	_, err = memdb.Exec(`CREATE TABLE tx_test (id TEXT PRIMARY KEY, val TEXT NOT NULL)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	ctx := context.Background()
	tx, err := memdb.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	// Single batch via tx
	_, err = QBulkInsert(ctx, tx, DialectSQLite, BulkInsertParams{
		Table:   "tx_test",
		Columns: []string{"id", "val"},
		Rows: [][]any{
			{"1", "a"},
			{"2", "b"},
		},
	})
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			t.Fatalf("insert: %v (rollback: %v)", err, rollbackErr)
		}
		t.Fatalf("insert: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	var count int64
	scanErr := memdb.QueryRow("SELECT COUNT(*) FROM tx_test").Scan(&count)
	if scanErr != nil {
		t.Fatalf("count: %v", scanErr)
	}
	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}
}

// ===== compositeResult =====

func TestCompositeResult(t *testing.T) {
	r := compositeResult{rowsAffected: 42}

	affected, err := r.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected error: %v", err)
	}
	if affected != 42 {
		t.Errorf("RowsAffected = %d, want 42", affected)
	}

	_, err = r.LastInsertId()
	if err == nil {
		t.Error("expected error from LastInsertId")
	}
}

// ===== SQL Generation Verification (Postgres Placeholders) =====

func TestFilteredSelect_PostgresSQL(t *testing.T) {
	p := FilteredSelectParams{
		Table:   "test_items",
		Columns: []string{"name", "status"},
		Filter: And{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "active"},
			InCondition{Column: "priority", Values: []any{1, 2, 3}},
		}},
		OrderByCols: []OrderByColumn{
			{Column: "priority", Desc: true},
		},
		Distinct: true,
		Limit:    10,
		Offset:   5,
	}

	query, args, err := buildFilteredSelectQuery(DialectPostgres, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `SELECT DISTINCT "name", "status" FROM "test_items" WHERE ("status" = $1 AND "priority" IN ($2, $3, $4)) ORDER BY "priority" DESC LIMIT 10 OFFSET 5`
	if query != want {
		t.Errorf("query = %q, want %q", query, want)
	}
	if len(args) != 4 {
		t.Fatalf("args len = %d, want 4", len(args))
	}
	if args[0] != "active" || args[1] != 1 || args[2] != 2 || args[3] != 3 {
		t.Errorf("args = %v", args)
	}
}

// ===== Filtered Aggregate Tests =====

func TestBuildFilteredSelectQuery_Aggregates(t *testing.T) {
	tests := []struct {
		name     string
		params   FilteredSelectParams
		dialect  Dialect
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "COUNT(*) only no plain columns",
			params: FilteredSelectParams{
				Table:      "test_items",
				Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*"}},
				Limit:      -1,
			},
			dialect:  DialectSQLite,
			wantSQL:  `SELECT COUNT(*) FROM "test_items"`,
			wantArgs: nil,
		},
		{
			name: "COUNT(*) with alias",
			params: FilteredSelectParams{
				Table:      "test_items",
				Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*", Alias: "total"}},
				Limit:      -1,
			},
			dialect:  DialectSQLite,
			wantSQL:  `SELECT COUNT(*) AS "total" FROM "test_items"`,
			wantArgs: nil,
		},
		{
			name: "mixed columns and aggregates",
			params: FilteredSelectParams{
				Table:      "test_items",
				Columns:    []string{"status"},
				Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*"}},
				Filter:     Compare{Column: "priority", Op: OpGt, Value: 0},
				Limit:      -1,
			},
			dialect:  DialectSQLite,
			wantSQL:  `SELECT "status", COUNT(*) FROM "test_items" WHERE "priority" > ?`,
			wantArgs: []any{0},
		},
		{
			name: "SUM with column arg",
			params: FilteredSelectParams{
				Table:      "test_items",
				Aggregates: []AggregateColumn{{Func: "SUM", Arg: "priority"}},
				Limit:      -1,
			},
			dialect:  DialectSQLite,
			wantSQL:  `SELECT SUM("priority") FROM "test_items"`,
			wantArgs: nil,
		},
		{
			name: "multiple aggregates",
			params: FilteredSelectParams{
				Table: "test_items",
				Aggregates: []AggregateColumn{
					{Func: "COUNT", Arg: "*"},
					{Func: "AVG", Arg: "priority"},
				},
				Limit: -1,
			},
			dialect:  DialectSQLite,
			wantSQL:  `SELECT COUNT(*), AVG("priority") FROM "test_items"`,
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args, err := buildFilteredSelectQuery(tt.dialect, tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if query != tt.wantSQL {
				t.Errorf("query = %q, want %q", query, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args len = %d, want %d", len(args), len(tt.wantArgs))
			}
			for i, a := range args {
				if a != tt.wantArgs[i] {
					t.Errorf("args[%d] = %v, want %v", i, a, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestBuildFilteredSelectQuery_GroupBy(t *testing.T) {
	t.Run("valid GROUP BY", func(t *testing.T) {
		query, _, err := buildFilteredSelectQuery(DialectSQLite, FilteredSelectParams{
			Table:      "test_items",
			Columns:    []string{"status"},
			Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*"}},
			GroupBy:    []string{"status"},
			Limit:      -1,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(query, `GROUP BY "status"`) {
			t.Errorf("query missing GROUP BY: %q", query)
		}
	})

	t.Run("invalid GROUP BY column name", func(t *testing.T) {
		_, _, err := buildFilteredSelectQuery(DialectSQLite, FilteredSelectParams{
			Table:      "test_items",
			Columns:    []string{"status"},
			Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*"}},
			GroupBy:    []string{"DROP"},
			Limit:      -1,
		})
		if err == nil {
			t.Fatal("expected error for invalid group_by column")
		}
		if !strings.Contains(err.Error(), "invalid group_by column") {
			t.Errorf("error = %q", err.Error())
		}
	})
}

func TestBuildFilteredSelectQuery_Having(t *testing.T) {
	t.Run("valid HAVING with AggregateCondition", func(t *testing.T) {
		query, args, err := buildFilteredSelectQuery(DialectSQLite, FilteredSelectParams{
			Table:      "test_items",
			Columns:    []string{"status"},
			Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*"}},
			GroupBy:    []string{"status"},
			Having:     AggregateCondition{Agg: AggregateColumn{Func: "COUNT", Arg: "*"}, Op: OpGt, Value: 5},
			Limit:      -1,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(query, `HAVING COUNT(*) > ?`) {
			t.Errorf("query missing HAVING: %q", query)
		}
		if len(args) != 1 || args[0] != 5 {
			t.Errorf("args = %v, want [5]", args)
		}
	})

	t.Run("HAVING without GroupBy rejected", func(t *testing.T) {
		_, _, err := buildFilteredSelectQuery(DialectSQLite, FilteredSelectParams{
			Table:      "test_items",
			Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*"}},
			Having:     AggregateCondition{Agg: AggregateColumn{Func: "COUNT", Arg: "*"}, Op: OpGt, Value: 5},
			Limit:      -1,
		})
		if err == nil {
			t.Fatal("expected error for HAVING without GROUP BY")
		}
		if !strings.Contains(err.Error(), "having requires group_by") {
			t.Errorf("error = %q", err.Error())
		}
	})
}

func TestBuildFilteredSelectQuery_NilFilter_WithAggregates(t *testing.T) {
	t.Run("nil filter accepted with aggregates", func(t *testing.T) {
		_, _, err := buildFilteredSelectQuery(DialectSQLite, FilteredSelectParams{
			Table:      "test_items",
			Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*"}},
			Limit:      -1,
		})
		if err != nil {
			t.Fatalf("expected nil filter to be accepted with aggregates, got: %v", err)
		}
	})

	t.Run("nil filter accepted with GroupBy", func(t *testing.T) {
		_, _, err := buildFilteredSelectQuery(DialectSQLite, FilteredSelectParams{
			Table:   "test_items",
			Columns: []string{"status"},
			GroupBy: []string{"status"},
			Limit:   -1,
		})
		if err != nil {
			t.Fatalf("expected nil filter to be accepted with GroupBy, got: %v", err)
		}
	})

	t.Run("nil filter rejected without aggregates or GroupBy", func(t *testing.T) {
		_, _, err := buildFilteredSelectQuery(DialectSQLite, FilteredSelectParams{
			Table: "test_items",
		})
		if err == nil {
			t.Fatal("expected error for nil filter without aggregates")
		}
		if !strings.Contains(err.Error(), "non-nil filter") {
			t.Errorf("error = %q", err.Error())
		}
	})
}

func TestBuildFilteredSelectQuery_ColumnCap(t *testing.T) {
	// Create enough columns + aggregates to exceed maxSelectColumns (64)
	cols := make([]string, 60)
	for i := range cols {
		cols[i] = fmt.Sprintf("col_%d", i)
	}
	aggs := []AggregateColumn{
		{Func: "COUNT", Arg: "*"},
		{Func: "SUM", Arg: "priority"},
		{Func: "AVG", Arg: "priority"},
		{Func: "MIN", Arg: "priority"},
		{Func: "MAX", Arg: "priority"},
	}

	_, _, err := buildFilteredSelectQuery(DialectSQLite, FilteredSelectParams{
		Table:      "test_items",
		Columns:    cols,
		Aggregates: aggs,
	})
	if err == nil {
		t.Fatal("expected error for too many select columns")
	}
	if !strings.Contains(err.Error(), "too many select columns") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestBuildFilteredSelectQuery_Having_PostgresPlaceholders(t *testing.T) {
	query, args, err := buildFilteredSelectQuery(DialectPostgres, FilteredSelectParams{
		Table:      "test_items",
		Columns:    []string{"status"},
		Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*", Alias: "cnt"}},
		Filter:     Compare{Column: "priority", Op: OpGt, Value: 0},
		GroupBy:    []string{"status"},
		Having:     AggregateCondition{Agg: AggregateColumn{Func: "COUNT", Arg: "*"}, Op: OpGte, Value: 2},
		Limit:      -1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WHERE uses $1 for priority, HAVING uses $2 for count threshold
	if !strings.Contains(query, `$1`) {
		t.Errorf("query missing $1 in WHERE: %q", query)
	}
	if !strings.Contains(query, `$2`) {
		t.Errorf("query missing $2 in HAVING: %q", query)
	}
	if len(args) != 2 {
		t.Fatalf("args len = %d, want 2", len(args))
	}
	if args[0] != 0 {
		t.Errorf("args[0] = %v, want 0 (WHERE value)", args[0])
	}
	if args[1] != 2 {
		t.Errorf("args[1] = %v, want 2 (HAVING value)", args[1])
	}

	// Verify full query structure
	want := `SELECT "status", COUNT(*) AS "cnt" FROM "test_items" WHERE "priority" > $1 GROUP BY "status" HAVING COUNT(*) >= $2`
	if query != want {
		t.Errorf("query = %q, want %q", query, want)
	}
}

func TestQSelectFiltered_GroupByAggregate(t *testing.T) {
	db := setupFilteredTestDB(t)
	ctx := context.Background()

	// Seed data has: active(3 rows), draft(1), deleted(1)
	rows, err := QSelectFiltered(ctx, db, DialectSQLite, FilteredSelectParams{
		Table:      "test_items",
		Columns:    []string{"status"},
		Aggregates: []AggregateColumn{{Func: "COUNT", Arg: "*", Alias: "cnt"}},
		GroupBy:    []string{"status"},
		OrderByCols: []OrderByColumn{
			{Column: "status", Desc: false},
		},
		Limit: -1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect 3 groups: active, deleted, draft
	if len(rows) != 3 {
		t.Fatalf("expected 3 grouped rows, got %d", len(rows))
	}

	// Verify one of the groups has the expected count
	// active has 3 rows (alpha, gamma, epsilon)
	foundActive := false
	for _, row := range rows {
		if row["status"] == "active" {
			cnt, ok := row["cnt"].(int64)
			if !ok {
				t.Fatalf("expected cnt to be int64, got %T", row["cnt"])
			}
			if cnt != 3 {
				t.Errorf("expected active count=3, got %d", cnt)
			}
			foundActive = true
		}
	}
	if !foundActive {
		t.Error("expected to find 'active' group in results")
	}
}
