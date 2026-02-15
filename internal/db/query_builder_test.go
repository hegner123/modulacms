package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	})

	_, err = db.Exec(`CREATE TABLE test_items (
		id TEXT PRIMARY KEY NOT NULL,
		name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		priority INTEGER NOT NULL DEFAULT 0,
		description TEXT
	)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	return db
}

func seedRows(t *testing.T, db *sql.DB) {
	t.Helper()
	rows := []struct {
		id, name, status string
		priority         int
		description      *string
	}{
		{"1", "alpha", "active", 3, strPtr("first item")},
		{"2", "beta", "active", 1, nil},
		{"3", "gamma", "archived", 2, strPtr("third item")},
		{"4", "delta", "active", 4, strPtr("fourth item")},
		{"5", "epsilon", "archived", 5, nil},
	}
	for _, r := range rows {
		var desc any
		if r.description != nil {
			desc = *r.description
		}
		_, err := db.Exec(
			`INSERT INTO test_items (id, name, status, priority, description) VALUES (?, ?, ?, ?, ?)`,
			r.id, r.name, r.status, r.priority, desc,
		)
		if err != nil {
			t.Fatalf("seed row %s: %v", r.id, err)
		}
	}
}

func strPtr(s string) *string { return &s }

func TestQSelect(t *testing.T) {
	ctx := context.Background()

	t.Run("all rows default limit", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{Table: "test_items"})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if len(rows) != 5 {
			t.Fatalf("expected 5 rows, got %d", len(rows))
		}
	})

	t.Run("single where condition", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"status": "active"},
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if len(rows) != 3 {
			t.Fatalf("expected 3 active rows, got %d", len(rows))
		}
	})

	t.Run("multiple where conditions AND", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"status": "active", "name": "alpha"},
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		if rows[0]["id"] != "1" {
			t.Fatalf("expected id=1, got %v", rows[0]["id"])
		}
	})

	t.Run("where with nil value IS NULL", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"description": nil},
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("expected 2 null-description rows, got %d", len(rows))
		}
	})

	t.Run("specific columns", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table:   "test_items",
			Columns: []string{"id", "name"},
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if len(rows) != 5 {
			t.Fatalf("expected 5 rows, got %d", len(rows))
		}
		// Should only have id and name columns
		row := rows[0]
		if _, ok := row["status"]; ok {
			t.Fatal("expected status column to be absent")
		}
		if _, ok := row["id"]; !ok {
			t.Fatal("expected id column to be present")
		}
		if _, ok := row["name"]; !ok {
			t.Fatal("expected name column to be present")
		}
	})

	t.Run("order by ASC", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table:   "test_items",
			OrderBy: "priority",
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		// Priority order: beta(1), gamma(2), alpha(3), delta(4), epsilon(5)
		if rows[0]["name"] != "beta" {
			t.Fatalf("expected first row to be beta, got %v", rows[0]["name"])
		}
		if rows[4]["name"] != "epsilon" {
			t.Fatalf("expected last row to be epsilon, got %v", rows[4]["name"])
		}
	})

	t.Run("order by DESC", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table:   "test_items",
			OrderBy: "priority",
			Desc:    true,
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if rows[0]["name"] != "epsilon" {
			t.Fatalf("expected first row to be epsilon, got %v", rows[0]["name"])
		}
		if rows[4]["name"] != "beta" {
			t.Fatalf("expected last row to be beta, got %v", rows[4]["name"])
		}
	})

	t.Run("limit and offset", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table:   "test_items",
			OrderBy: "priority",
			Limit:   2,
			Offset:  1,
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(rows))
		}
		// Skips beta(1), returns gamma(2) and alpha(3)
		if rows[0]["name"] != "gamma" {
			t.Fatalf("expected first row to be gamma, got %v", rows[0]["name"])
		}
		if rows[1]["name"] != "alpha" {
			t.Fatalf("expected second row to be alpha, got %v", rows[1]["name"])
		}
	})

	t.Run("all clauses combined", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table:   "test_items",
			Columns: []string{"id", "name", "priority"},
			Where:   map[string]any{"status": "active"},
			OrderBy: "priority",
			Desc:    true,
			Limit:   2,
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(rows))
		}
		// Active sorted by priority DESC: delta(4), alpha(3), beta(1) -> first 2
		if rows[0]["name"] != "delta" {
			t.Fatalf("expected first row to be delta, got %v", rows[0]["name"])
		}
		if rows[1]["name"] != "alpha" {
			t.Fatalf("expected second row to be alpha, got %v", rows[1]["name"])
		}
	})

	t.Run("limit capped at 10000", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		// Should not error, just cap at 10000
		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Limit: 99999,
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if len(rows) != 5 {
			t.Fatalf("expected 5 rows, got %d", len(rows))
		}
	})

	t.Run("negative limit disables limit", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Limit: -1,
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		if len(rows) != 5 {
			t.Fatalf("expected 5 rows, got %d", len(rows))
		}
	})
}

func TestQSelectOne(t *testing.T) {
	ctx := context.Background()

	t.Run("returns first match", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		row, err := QSelectOne(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"id": "2"},
		})
		if err != nil {
			t.Fatalf("select one: %v", err)
		}
		if row == nil {
			t.Fatal("expected a row, got nil")
		}
		if row["name"] != "beta" {
			t.Fatalf("expected name=beta, got %v", row["name"])
		}
	})

	t.Run("returns nil when no match", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		row, err := QSelectOne(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"id": "nonexistent"},
		})
		if err != nil {
			t.Fatalf("select one: %v", err)
		}
		if row != nil {
			t.Fatalf("expected nil, got %v", row)
		}
	})
}

func TestQInsert(t *testing.T) {
	ctx := context.Background()

	t.Run("insert single row", func(t *testing.T) {
		db := setupTestDB(t)

		result, err := QInsert(ctx, db, DialectSQLite, InsertParams{
			Table: "test_items",
			Values: map[string]any{
				"id":       "100",
				"name":     "zeta",
				"status":   "active",
				"priority": 10,
			},
		})
		if err != nil {
			t.Fatalf("insert: %v", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			t.Fatalf("rows affected: %v", err)
		}
		if affected != 1 {
			t.Fatalf("expected 1 row affected, got %d", affected)
		}

		// Verify with select
		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"id": "100"},
		})
		if err != nil {
			t.Fatalf("verify select: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		if rows[0]["name"] != "zeta" {
			t.Fatalf("expected name=zeta, got %v", rows[0]["name"])
		}
	})

	t.Run("insert with nil value", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QInsert(ctx, db, DialectSQLite, InsertParams{
			Table: "test_items",
			Values: map[string]any{
				"id":          "200",
				"name":        "eta",
				"status":      "active",
				"priority":    0,
				"description": nil,
			},
		})
		if err != nil {
			t.Fatalf("insert: %v", err)
		}

		row, err := QSelectOne(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"id": "200"},
		})
		if err != nil {
			t.Fatalf("verify select: %v", err)
		}
		if row == nil {
			t.Fatal("expected a row")
		}
		if row["description"] != nil {
			t.Fatalf("expected nil description, got %v", row["description"])
		}
	})

	t.Run("empty values error", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QInsert(ctx, db, DialectSQLite, InsertParams{
			Table:  "test_items",
			Values: map[string]any{},
		})
		if err == nil {
			t.Fatal("expected error for empty values")
		}
	})
}

func TestQUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("update matching rows", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		result, err := QUpdate(ctx, db, DialectSQLite, UpdateParams{
			Table: "test_items",
			Set:   map[string]any{"status": "done"},
			Where: map[string]any{"id": "1"},
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			t.Fatalf("rows affected: %v", err)
		}
		if affected != 1 {
			t.Fatalf("expected 1 row affected, got %d", affected)
		}

		row, err := QSelectOne(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"id": "1"},
		})
		if err != nil {
			t.Fatalf("verify select: %v", err)
		}
		if row["status"] != "done" {
			t.Fatalf("expected status=done, got %v", row["status"])
		}
	})

	t.Run("update with multi-column where", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		result, err := QUpdate(ctx, db, DialectSQLite, UpdateParams{
			Table: "test_items",
			Set:   map[string]any{"priority": 99},
			Where: map[string]any{"status": "active", "name": "beta"},
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			t.Fatalf("rows affected: %v", err)
		}
		if affected != 1 {
			t.Fatalf("expected 1 row affected, got %d", affected)
		}

		row, err := QSelectOne(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"id": "2"},
		})
		if err != nil {
			t.Fatalf("verify: %v", err)
		}
		// SQLite returns int64
		if row["priority"] != int64(99) {
			t.Fatalf("expected priority=99, got %v", row["priority"])
		}
	})

	t.Run("update set to NULL", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		_, err := QUpdate(ctx, db, DialectSQLite, UpdateParams{
			Table: "test_items",
			Set:   map[string]any{"description": nil},
			Where: map[string]any{"id": "1"},
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}

		row, err := QSelectOne(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"id": "1"},
		})
		if err != nil {
			t.Fatalf("verify: %v", err)
		}
		if row["description"] != nil {
			t.Fatalf("expected nil description, got %v", row["description"])
		}
	})

	t.Run("empty set error", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QUpdate(ctx, db, DialectSQLite, UpdateParams{
			Table: "test_items",
			Set:   map[string]any{},
			Where: map[string]any{"id": "1"},
		})
		if err == nil {
			t.Fatal("expected error for empty set")
		}
	})

	t.Run("empty where error", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QUpdate(ctx, db, DialectSQLite, UpdateParams{
			Table: "test_items",
			Set:   map[string]any{"status": "done"},
			Where: map[string]any{},
		})
		if err == nil {
			t.Fatal("expected error for empty where")
		}
	})
}

func TestQDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("delete matching rows", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		result, err := QDelete(ctx, db, DialectSQLite, DeleteParams{
			Table: "test_items",
			Where: map[string]any{"id": "3"},
		})
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			t.Fatalf("rows affected: %v", err)
		}
		if affected != 1 {
			t.Fatalf("expected 1 row affected, got %d", affected)
		}

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{Table: "test_items"})
		if err != nil {
			t.Fatalf("verify: %v", err)
		}
		if len(rows) != 4 {
			t.Fatalf("expected 4 remaining rows, got %d", len(rows))
		}
	})

	t.Run("delete with multi-column where", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		result, err := QDelete(ctx, db, DialectSQLite, DeleteParams{
			Table: "test_items",
			Where: map[string]any{"status": "archived"},
		})
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			t.Fatalf("rows affected: %v", err)
		}
		if affected != 2 {
			t.Fatalf("expected 2 rows affected, got %d", affected)
		}

		rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{Table: "test_items"})
		if err != nil {
			t.Fatalf("verify: %v", err)
		}
		if len(rows) != 3 {
			t.Fatalf("expected 3 remaining rows, got %d", len(rows))
		}
	})

	t.Run("empty where error", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QDelete(ctx, db, DialectSQLite, DeleteParams{
			Table: "test_items",
			Where: map[string]any{},
		})
		if err == nil {
			t.Fatal("expected error for empty where")
		}
	})
}

func TestQCount(t *testing.T) {
	ctx := context.Background()

	t.Run("count all", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		count, err := QCount(ctx, db, DialectSQLite, "test_items", nil)
		if err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 5 {
			t.Fatalf("expected 5, got %d", count)
		}
	})

	t.Run("count with where", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		count, err := QCount(ctx, db, DialectSQLite, "test_items", map[string]any{"status": "active"})
		if err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 3 {
			t.Fatalf("expected 3, got %d", count)
		}
	})
}

func TestQExists(t *testing.T) {
	ctx := context.Background()

	t.Run("exists true", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		exists, err := QExists(ctx, db, DialectSQLite, "test_items", map[string]any{"id": "1"})
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if !exists {
			t.Fatal("expected true")
		}
	})

	t.Run("exists false", func(t *testing.T) {
		db := setupTestDB(t)
		seedRows(t, db)

		exists, err := QExists(ctx, db, DialectSQLite, "test_items", map[string]any{"id": "nonexistent"})
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if exists {
			t.Fatal("expected false")
		}
	})
}

func TestValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid table name", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QSelect(ctx, db, DialectSQLite, SelectParams{Table: "DROP TABLE--"})
		if err == nil {
			t.Fatal("expected error for invalid table name")
		}
	})

	t.Run("invalid column in where", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"bad column!": "x"},
		})
		if err == nil {
			t.Fatal("expected error for invalid column name")
		}
	})

	t.Run("invalid column in columns list", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table:   "test_items",
			Columns: []string{"id", "1bad"},
		})
		if err == nil {
			t.Fatal("expected error for invalid column name")
		}
	})

	t.Run("SQL keyword as table name", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QSelect(ctx, db, DialectSQLite, SelectParams{Table: "SELECT"})
		if err == nil {
			t.Fatal("expected error for SQL keyword table name")
		}
	})

	t.Run("SQL keyword as column name", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table:   "test_items",
			Columns: []string{"id", "SELECT"},
		})
		if err == nil {
			t.Fatal("expected error for SQL keyword column name")
		}
	})

	t.Run("SQL keyword in where column", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table: "test_items",
			Where: map[string]any{"DELETE": "x"},
		})
		if err == nil {
			t.Fatal("expected error for SQL keyword where column")
		}
	})

	t.Run("invalid order_by column", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QSelect(ctx, db, DialectSQLite, SelectParams{
			Table:   "test_items",
			OrderBy: "bad col!",
		})
		if err == nil {
			t.Fatal("expected error for invalid order_by column")
		}
	})

	t.Run("invalid table on insert", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QInsert(ctx, db, DialectSQLite, InsertParams{
			Table:  "DROP",
			Values: map[string]any{"id": "1"},
		})
		if err == nil {
			t.Fatal("expected error for SQL keyword table name")
		}
	})

	t.Run("invalid column on insert", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QInsert(ctx, db, DialectSQLite, InsertParams{
			Table:  "test_items",
			Values: map[string]any{"bad col!": "x"},
		})
		if err == nil {
			t.Fatal("expected error for invalid column name")
		}
	})

	t.Run("invalid table on update", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QUpdate(ctx, db, DialectSQLite, UpdateParams{
			Table: "DELETE",
			Set:   map[string]any{"name": "x"},
			Where: map[string]any{"id": "1"},
		})
		if err == nil {
			t.Fatal("expected error for SQL keyword table name")
		}
	})

	t.Run("invalid set column on update", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QUpdate(ctx, db, DialectSQLite, UpdateParams{
			Table: "test_items",
			Set:   map[string]any{"bad!": "x"},
			Where: map[string]any{"id": "1"},
		})
		if err == nil {
			t.Fatal("expected error for invalid set column")
		}
	})

	t.Run("invalid table on delete", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QDelete(ctx, db, DialectSQLite, DeleteParams{
			Table: "TABLE",
			Where: map[string]any{"id": "1"},
		})
		if err == nil {
			t.Fatal("expected error for SQL keyword table name")
		}
	})

	t.Run("invalid table on count", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QCount(ctx, db, DialectSQLite, "INSERT", nil)
		if err == nil {
			t.Fatal("expected error for SQL keyword table name")
		}
	})

	t.Run("invalid table on exists", func(t *testing.T) {
		db := setupTestDB(t)

		_, err := QExists(ctx, db, DialectSQLite, "UPDATE", nil)
		if err == nil {
			t.Fatal("expected error for SQL keyword table name")
		}
	})
}

// ===== DDL TESTS =====

func TestValidateColumnType(t *testing.T) {
	t.Run("valid types", func(t *testing.T) {
		valid := []string{"text", "integer", "real", "blob", "boolean", "timestamp", "json"}
		for _, ct := range valid {
			if err := ValidateColumnType(ct); err != nil {
				t.Errorf("ValidateColumnType(%q) returned error: %v", ct, err)
			}
		}
	})

	t.Run("invalid types", func(t *testing.T) {
		invalid := []string{"", "varchar", "int", "string", "TEXT", "INTEGER", "bool", "datetime"}
		for _, ct := range invalid {
			if err := ValidateColumnType(ct); err == nil {
				t.Errorf("ValidateColumnType(%q) expected error, got nil", ct)
			}
		}
	})
}

func TestSQLType(t *testing.T) {
	cases := []struct {
		dialect  Dialect
		colType  ColumnType
		expected string
	}{
		// SQLite
		{DialectSQLite, ColText, "TEXT"},
		{DialectSQLite, ColInteger, "INTEGER"},
		{DialectSQLite, ColReal, "REAL"},
		{DialectSQLite, ColBlob, "BLOB"},
		{DialectSQLite, ColBoolean, "INTEGER"},
		{DialectSQLite, ColTimestamp, "TEXT"},
		{DialectSQLite, ColJSON, "TEXT"},
		// MySQL
		{DialectMySQL, ColText, "TEXT"},
		{DialectMySQL, ColInteger, "INT"},
		{DialectMySQL, ColReal, "DOUBLE"},
		{DialectMySQL, ColBlob, "BLOB"},
		{DialectMySQL, ColBoolean, "TINYINT(1)"},
		{DialectMySQL, ColTimestamp, "TIMESTAMP"},
		{DialectMySQL, ColJSON, "JSON"},
		// PostgreSQL
		{DialectPostgres, ColText, "TEXT"},
		{DialectPostgres, ColInteger, "INTEGER"},
		{DialectPostgres, ColReal, "DOUBLE PRECISION"},
		{DialectPostgres, ColBlob, "BYTEA"},
		{DialectPostgres, ColBoolean, "BOOLEAN"},
		{DialectPostgres, ColTimestamp, "TIMESTAMP"},
		{DialectPostgres, ColJSON, "JSONB"},
	}

	for _, tc := range cases {
		var dialectName string
		switch tc.dialect {
		case DialectMySQL:
			dialectName = "mysql"
		case DialectPostgres:
			dialectName = "postgres"
		default:
			dialectName = "sqlite"
		}
		name := dialectName + "/" + string(tc.colType)
		t.Run(name, func(t *testing.T) {
			got := SQLType(tc.dialect, tc.colType)
			if got != tc.expected {
				t.Errorf("SQLType(%s, %s) = %q, want %q", dialectName, tc.colType, got, tc.expected)
			}
		})
	}
}

func TestDDLCreateTable_SQLite(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "widgets",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "name", Type: ColText, NotNull: true},
			{Name: "count", Type: ColInteger, NotNull: true, Default: "0"},
			{Name: "active", Type: ColBoolean},
			{Name: "data", Type: ColJSON},
			{Name: "created_at", Type: ColTimestamp, NotNull: true},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("DDLCreateTable: %v", err)
	}

	// Verify table exists by inserting a row
	_, err = QInsert(ctx, db, DialectSQLite, InsertParams{
		Table: "widgets",
		Values: map[string]any{
			"id":         "w1",
			"name":       "Sprocket",
			"count":      5,
			"active":     1,
			"data":       `{"color":"blue"}`,
			"created_at": "2026-02-07T12:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("insert into created table: %v", err)
	}

	rows, err := QSelect(ctx, db, DialectSQLite, SelectParams{
		Table: "widgets",
		Where: map[string]any{"id": "w1"},
	})
	if err != nil {
		t.Fatalf("select from created table: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["name"] != "Sprocket" {
		t.Fatalf("expected name=Sprocket, got %v", rows[0]["name"])
	}
}

func TestDDLCreateTable_WithDefault(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "statuses",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "state", Type: ColText, NotNull: true, Default: "pending"},
			{Name: "priority", Type: ColInteger, NotNull: true, Default: "0"},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("DDLCreateTable: %v", err)
	}

	// Insert with only id to check defaults
	_, err = db.ExecContext(ctx, `INSERT INTO "statuses" ("id") VALUES ('s1')`)
	if err != nil {
		t.Fatalf("insert with defaults: %v", err)
	}

	row, err := QSelectOne(ctx, db, DialectSQLite, SelectParams{
		Table: "statuses",
		Where: map[string]any{"id": "s1"},
	})
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if row["state"] != "pending" {
		t.Fatalf("expected default state=pending, got %v", row["state"])
	}
}

func TestDDLCreateTable_WithDefaultEscaping(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "notes",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "label", Type: ColText, NotNull: true, Default: "it's a test"},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("DDLCreateTable: %v", err)
	}

	_, err = db.ExecContext(ctx, `INSERT INTO "notes" ("id") VALUES ('n1')`)
	if err != nil {
		t.Fatalf("insert with escaped default: %v", err)
	}

	row, err := QSelectOne(ctx, db, DialectSQLite, SelectParams{
		Table: "notes",
		Where: map[string]any{"id": "n1"},
	})
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if row["label"] != "it's a test" {
		t.Fatalf("expected default with quote, got %v", row["label"])
	}
}

func TestDDLCreateTable_WithIndexes(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "tasks",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "status", Type: ColText, NotNull: true},
			{Name: "priority", Type: ColInteger, NotNull: true},
			{Name: "title", Type: ColText, NotNull: true},
		},
		Indexes: []IndexDef{
			{Columns: []string{"status"}},
			{Columns: []string{"status", "priority"}},
			{Columns: []string{"title"}, Unique: true},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("DDLCreateTable with indexes: %v", err)
	}

	// Verify indexes exist by querying sqlite_master (filter out SQLite auto-indexes)
	rows, err := db.QueryContext(ctx,
		`SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='tasks' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	defer rows.Close()

	var indexNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan index name: %v", err)
		}
		indexNames = append(indexNames, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error: %v", err)
	}

	expected := []string{
		"idx_tasks_status",
		"idx_tasks_status_priority",
		"idx_tasks_title",
	}
	if len(indexNames) != len(expected) {
		t.Fatalf("expected %d indexes, got %d: %v", len(expected), len(indexNames), indexNames)
	}
	for i, name := range expected {
		if indexNames[i] != name {
			t.Errorf("index %d: expected %q, got %q", i, name, indexNames[i])
		}
	}
}

func TestDDLCreateTable_WithForeignKeys(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// Enable FK enforcement for SQLite
	_, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	// Create parent table
	err = DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "projects",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "name", Type: ColText, NotNull: true},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("DDLCreateTable projects: %v", err)
	}

	// Create child table with FK
	err = DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "tasks",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "project_id", Type: ColText, NotNull: true},
			{Name: "title", Type: ColText, NotNull: true},
		},
		ForeignKeys: []ForeignKeyDef{
			{
				Column:    "project_id",
				RefTable:  "projects",
				RefColumn: "id",
				OnDelete:  "CASCADE",
			},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("DDLCreateTable tasks with FK: %v", err)
	}

	// Insert parent
	_, err = QInsert(ctx, db, DialectSQLite, InsertParams{
		Table:  "projects",
		Values: map[string]any{"id": "p1", "name": "Project Alpha"},
	})
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}

	// Insert child referencing parent
	_, err = QInsert(ctx, db, DialectSQLite, InsertParams{
		Table:  "tasks",
		Values: map[string]any{"id": "t1", "project_id": "p1", "title": "Task 1"},
	})
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	// Try inserting child with invalid FK reference — should fail
	_, err = QInsert(ctx, db, DialectSQLite, InsertParams{
		Table:  "tasks",
		Values: map[string]any{"id": "t2", "project_id": "nonexistent", "title": "Task 2"},
	})
	if err == nil {
		t.Fatal("expected FK violation error, got nil")
	}

	// Delete parent — should cascade and remove child
	_, err = QDelete(ctx, db, DialectSQLite, DeleteParams{
		Table: "projects",
		Where: map[string]any{"id": "p1"},
	})
	if err != nil {
		t.Fatalf("delete project: %v", err)
	}

	count, err := QCount(ctx, db, DialectSQLite, "tasks", nil)
	if err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 tasks after cascade delete, got %d", count)
	}
}

func TestDDLCreateTable_Validation(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	t.Run("empty table name", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "",
			Columns: []CreateColumnDef{
				{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			},
		})
		if err == nil {
			t.Fatal("expected error for empty table name")
		}
	})

	t.Run("empty columns", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table:   "empty_cols",
			Columns: []CreateColumnDef{},
		})
		if err == nil {
			t.Fatal("expected error for empty columns")
		}
	})

	t.Run("no primary key", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "no_pk",
			Columns: []CreateColumnDef{
				{Name: "name", Type: ColText, NotNull: true},
			},
		})
		if err == nil {
			t.Fatal("expected error for no primary key")
		}
	})

	t.Run("multiple primary keys", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "multi_pk",
			Columns: []CreateColumnDef{
				{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
				{Name: "id2", Type: ColText, PrimaryKey: true, NotNull: true},
			},
		})
		if err == nil {
			t.Fatal("expected error for multiple primary keys")
		}
	})

	t.Run("invalid column type", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "bad_type",
			Columns: []CreateColumnDef{
				{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
				{Name: "data", Type: "varchar"},
			},
		})
		if err == nil {
			t.Fatal("expected error for invalid column type")
		}
	})

	t.Run("invalid column name", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "bad_col",
			Columns: []CreateColumnDef{
				{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
				{Name: "bad name!", Type: ColText},
			},
		})
		if err == nil {
			t.Fatal("expected error for invalid column name")
		}
	})

	t.Run("duplicate column name", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "dup_col",
			Columns: []CreateColumnDef{
				{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
				{Name: "name", Type: ColText},
				{Name: "name", Type: ColInteger},
			},
		})
		if err == nil {
			t.Fatal("expected error for duplicate column name")
		}
	})

	t.Run("too many columns", func(t *testing.T) {
		cols := make([]CreateColumnDef, 65)
		cols[0] = CreateColumnDef{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true}
		for i := 1; i < 65; i++ {
			cols[i] = CreateColumnDef{Name: fmt.Sprintf("col_%d", i), Type: ColText}
		}
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table:   "too_many",
			Columns: cols,
		})
		if err == nil {
			t.Fatal("expected error for too many columns")
		}
	})

	t.Run("invalid FK ON DELETE action", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "bad_fk",
			Columns: []CreateColumnDef{
				{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
				{Name: "ref_id", Type: ColText},
			},
			ForeignKeys: []ForeignKeyDef{
				{Column: "ref_id", RefTable: "other_table", RefColumn: "id", OnDelete: "DESTROY"},
			},
		})
		if err == nil {
			t.Fatal("expected error for invalid ON DELETE action")
		}
	})

	t.Run("FK column not in definitions", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "fk_missing_col",
			Columns: []CreateColumnDef{
				{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			},
			ForeignKeys: []ForeignKeyDef{
				{Column: "ref_id", RefTable: "other_table", RefColumn: "id", OnDelete: "CASCADE"},
			},
		})
		if err == nil {
			t.Fatal("expected error for FK column not in definitions")
		}
	})

	t.Run("index column not in definitions", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "idx_missing_col",
			Columns: []CreateColumnDef{
				{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			},
			Indexes: []IndexDef{
				{Columns: []string{"nonexistent"}},
			},
		})
		if err == nil {
			t.Fatal("expected error for index column not in definitions")
		}
	})

	t.Run("SQL keyword as table name", func(t *testing.T) {
		err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
			Table: "SELECT",
			Columns: []CreateColumnDef{
				{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			},
		})
		if err == nil {
			t.Fatal("expected error for SQL keyword table name")
		}
	})
}

func TestDDLCreateTable_IfNotExists(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	params := DDLCreateTableParams{
		Table: "idempotent_table",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "value", Type: ColText},
		},
		IfNotExists: true,
	}

	// First creation
	err := DDLCreateTable(ctx, db, DialectSQLite, params)
	if err != nil {
		t.Fatalf("first DDLCreateTable: %v", err)
	}

	// Second creation — should not error
	err = DDLCreateTable(ctx, db, DialectSQLite, params)
	if err != nil {
		t.Fatalf("second DDLCreateTable (idempotent): %v", err)
	}
}

func TestDDLCreateTable_WithoutIfNotExists(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	params := DDLCreateTableParams{
		Table: "non_idempotent",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
		},
		IfNotExists: false,
	}

	err := DDLCreateTable(ctx, db, DialectSQLite, params)
	if err != nil {
		t.Fatalf("first DDLCreateTable: %v", err)
	}

	// Second creation without IF NOT EXISTS — should error
	err = DDLCreateTable(ctx, db, DialectSQLite, params)
	if err == nil {
		t.Fatal("expected error for duplicate table without IF NOT EXISTS")
	}
}

func TestDDLCreateTable_UniqueColumn(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "emails",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "email", Type: ColText, NotNull: true, Unique: true},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("DDLCreateTable: %v", err)
	}

	_, err = QInsert(ctx, db, DialectSQLite, InsertParams{
		Table:  "emails",
		Values: map[string]any{"id": "e1", "email": "test@example.com"},
	})
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}

	// Duplicate email should fail
	_, err = QInsert(ctx, db, DialectSQLite, InsertParams{
		Table:  "emails",
		Values: map[string]any{"id": "e2", "email": "test@example.com"},
	})
	if err == nil {
		t.Fatal("expected unique constraint violation")
	}
}

func TestDDLCreateTable_AllDialects_DDLString(t *testing.T) {
	// This test verifies the DDL string generation for each dialect
	// by checking that SQLType returns the correct types used in the DDL.
	// We can't execute MySQL/Postgres DDL without those databases,
	// but we verify the type mapping is correct.

	t.Run("mysql types", func(t *testing.T) {
		expectations := map[ColumnType]string{
			ColText:      "TEXT",
			ColInteger:   "INT",
			ColReal:      "DOUBLE",
			ColBlob:      "BLOB",
			ColBoolean:   "TINYINT(1)",
			ColTimestamp: "TIMESTAMP",
			ColJSON:      "JSON",
		}
		for ct, expected := range expectations {
			got := SQLType(DialectMySQL, ct)
			if got != expected {
				t.Errorf("MySQL SQLType(%s) = %q, want %q", ct, got, expected)
			}
		}
	})

	t.Run("postgres types", func(t *testing.T) {
		expectations := map[ColumnType]string{
			ColText:      "TEXT",
			ColInteger:   "INTEGER",
			ColReal:      "DOUBLE PRECISION",
			ColBlob:      "BYTEA",
			ColBoolean:   "BOOLEAN",
			ColTimestamp: "TIMESTAMP",
			ColJSON:      "JSONB",
		}
		for ct, expected := range expectations {
			got := SQLType(DialectPostgres, ct)
			if got != expected {
				t.Errorf("Postgres SQLType(%s) = %q, want %q", ct, got, expected)
			}
		}
	})
}

func TestDDLCreateIndex(t *testing.T) {
	ctx := context.Background()

	t.Run("single column index", func(t *testing.T) {
		db := openTestDB(t)
		createTestTable(t, ctx, db)

		err := DDLCreateIndex(ctx, db, DialectSQLite, DDLCreateIndexParams{
			Table:       "test_ddl",
			Columns:     []string{"status"},
			IfNotExists: true,
		})
		if err != nil {
			t.Fatalf("DDLCreateIndex: %v", err)
		}

		// Verify index exists
		assertIndexExists(t, ctx, db, "test_ddl", "idx_test_ddl_status")
	})

	t.Run("composite index", func(t *testing.T) {
		db := openTestDB(t)
		createTestTable(t, ctx, db)

		err := DDLCreateIndex(ctx, db, DialectSQLite, DDLCreateIndexParams{
			Table:       "test_ddl",
			Columns:     []string{"status", "priority"},
			IfNotExists: true,
		})
		if err != nil {
			t.Fatalf("DDLCreateIndex: %v", err)
		}

		assertIndexExists(t, ctx, db, "test_ddl", "idx_test_ddl_status_priority")
	})

	t.Run("unique index", func(t *testing.T) {
		db := openTestDB(t)
		createTestTable(t, ctx, db)

		err := DDLCreateIndex(ctx, db, DialectSQLite, DDLCreateIndexParams{
			Table:       "test_ddl",
			Columns:     []string{"name"},
			Unique:      true,
			IfNotExists: true,
		})
		if err != nil {
			t.Fatalf("DDLCreateIndex unique: %v", err)
		}

		assertIndexExists(t, ctx, db, "test_ddl", "idx_test_ddl_name")

		// Verify uniqueness enforced
		_, err = db.ExecContext(ctx, `INSERT INTO "test_ddl" ("id", "name", "status", "priority") VALUES ('a', 'dup', 'x', 0)`)
		if err != nil {
			t.Fatalf("first insert: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO "test_ddl" ("id", "name", "status", "priority") VALUES ('b', 'dup', 'y', 1)`)
		if err == nil {
			t.Fatal("expected unique index violation")
		}
	})

	t.Run("if not exists idempotent", func(t *testing.T) {
		db := openTestDB(t)
		createTestTable(t, ctx, db)

		params := DDLCreateIndexParams{
			Table:       "test_ddl",
			Columns:     []string{"status"},
			IfNotExists: true,
		}

		err := DDLCreateIndex(ctx, db, DialectSQLite, params)
		if err != nil {
			t.Fatalf("first DDLCreateIndex: %v", err)
		}

		err = DDLCreateIndex(ctx, db, DialectSQLite, params)
		if err != nil {
			t.Fatalf("second DDLCreateIndex (idempotent): %v", err)
		}
	})

	t.Run("empty columns error", func(t *testing.T) {
		db := openTestDB(t)

		err := DDLCreateIndex(ctx, db, DialectSQLite, DDLCreateIndexParams{
			Table:   "test_ddl",
			Columns: []string{},
		})
		if err == nil {
			t.Fatal("expected error for empty columns")
		}
	})

	t.Run("invalid table name error", func(t *testing.T) {
		db := openTestDB(t)

		err := DDLCreateIndex(ctx, db, DialectSQLite, DDLCreateIndexParams{
			Table:   "DROP",
			Columns: []string{"id"},
		})
		if err == nil {
			t.Fatal("expected error for SQL keyword table name")
		}
	})

	t.Run("invalid column name error", func(t *testing.T) {
		db := openTestDB(t)

		err := DDLCreateIndex(ctx, db, DialectSQLite, DDLCreateIndexParams{
			Table:   "test_ddl",
			Columns: []string{"bad col!"},
		})
		if err == nil {
			t.Fatal("expected error for invalid column name")
		}
	})
}

func TestDDLCreateTable_FKWithOnDelete(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	_, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	// Test SET NULL on delete
	err = DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "authors",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "name", Type: ColText, NotNull: true},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("create authors: %v", err)
	}

	err = DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "articles",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "author_id", Type: ColText},
			{Name: "title", Type: ColText, NotNull: true},
		},
		ForeignKeys: []ForeignKeyDef{
			{Column: "author_id", RefTable: "authors", RefColumn: "id", OnDelete: "SET NULL"},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("create articles: %v", err)
	}

	_, err = QInsert(ctx, db, DialectSQLite, InsertParams{
		Table:  "authors",
		Values: map[string]any{"id": "a1", "name": "Alice"},
	})
	if err != nil {
		t.Fatalf("insert author: %v", err)
	}

	_, err = QInsert(ctx, db, DialectSQLite, InsertParams{
		Table:  "articles",
		Values: map[string]any{"id": "art1", "author_id": "a1", "title": "Article 1"},
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}

	// Delete author — should set author_id to NULL
	_, err = QDelete(ctx, db, DialectSQLite, DeleteParams{
		Table: "authors",
		Where: map[string]any{"id": "a1"},
	})
	if err != nil {
		t.Fatalf("delete author: %v", err)
	}

	row, err := QSelectOne(ctx, db, DialectSQLite, SelectParams{
		Table: "articles",
		Where: map[string]any{"id": "art1"},
	})
	if err != nil {
		t.Fatalf("select article: %v", err)
	}
	if row["author_id"] != nil {
		t.Fatalf("expected author_id=nil after SET NULL, got %v", row["author_id"])
	}
}

// ===== DDL TEST HELPERS =====

// openTestDB creates a fresh in-memory SQLite database for testing.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	})
	return db
}

// createTestTable creates a simple test table for DDL tests.
func createTestTable(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	err := DDLCreateTable(ctx, db, DialectSQLite, DDLCreateTableParams{
		Table: "test_ddl",
		Columns: []CreateColumnDef{
			{Name: "id", Type: ColText, PrimaryKey: true, NotNull: true},
			{Name: "name", Type: ColText, NotNull: true},
			{Name: "status", Type: ColText, NotNull: true},
			{Name: "priority", Type: ColInteger, NotNull: true},
		},
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("create test_ddl table: %v", err)
	}
}

// assertIndexExists checks that a named index exists on the given table in SQLite.
func assertIndexExists(t *testing.T, ctx context.Context, db *sql.DB, table, indexName string) {
	t.Helper()
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND tbl_name=? AND name=?`,
		table, indexName,
	).Scan(&count)
	if err != nil {
		t.Fatalf("check index existence: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected index %q on table %q to exist", indexName, table)
	}
}

func TestDialect(t *testing.T) {
	t.Run("DialectFromString", func(t *testing.T) {
		cases := []struct {
			input    string
			expected Dialect
		}{
			{"sqlite", DialectSQLite},
			{"mysql", DialectMySQL},
			{"postgres", DialectPostgres},
			{"unknown", DialectSQLite},
			{"", DialectSQLite},
		}
		for _, tc := range cases {
			got := DialectFromString(tc.input)
			if got != tc.expected {
				t.Errorf("DialectFromString(%q) = %d, want %d", tc.input, got, tc.expected)
			}
		}
	})

	t.Run("quoteIdent", func(t *testing.T) {
		if got := quoteIdent(DialectSQLite, "name"); got != `"name"` {
			t.Errorf("SQLite quoteIdent = %s, want %s", got, `"name"`)
		}
		if got := quoteIdent(DialectMySQL, "name"); got != "`name`" {
			t.Errorf("MySQL quoteIdent = %s, want %s", got, "`name`")
		}
		if got := quoteIdent(DialectPostgres, "name"); got != `"name"` {
			t.Errorf("Postgres quoteIdent = %s, want %s", got, `"name"`)
		}
	})

	t.Run("placeholder", func(t *testing.T) {
		if got := placeholder(DialectSQLite, 1); got != "?" {
			t.Errorf("SQLite placeholder = %s, want ?", got)
		}
		if got := placeholder(DialectMySQL, 3); got != "?" {
			t.Errorf("MySQL placeholder = %s, want ?", got)
		}
		if got := placeholder(DialectPostgres, 1); got != "$1" {
			t.Errorf("Postgres placeholder(1) = %s, want $1", got)
		}
		if got := placeholder(DialectPostgres, 5); got != "$5" {
			t.Errorf("Postgres placeholder(5) = %s, want $5", got)
		}
	})
}

func TestTransactions(t *testing.T) {
	ctx := context.Background()

	t.Run("commit both inserts", func(t *testing.T) {
		db := setupTestDB(t)

		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}

		_, err = QInsert(ctx, tx, DialectSQLite, InsertParams{
			Table: "test_items",
			Values: map[string]any{
				"id": "tx1", "name": "txalpha", "status": "active", "priority": 1,
			},
		})
		if err != nil {
			t.Fatalf("insert 1: %v", err)
		}

		_, err = QInsert(ctx, tx, DialectSQLite, InsertParams{
			Table: "test_items",
			Values: map[string]any{
				"id": "tx2", "name": "txbeta", "status": "active", "priority": 2,
			},
		})
		if err != nil {
			t.Fatalf("insert 2: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("commit: %v", err)
		}

		count, err := QCount(ctx, db, DialectSQLite, "test_items", nil)
		if err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 2 {
			t.Fatalf("expected 2 rows after commit, got %d", count)
		}
	})

	t.Run("rollback on error", func(t *testing.T) {
		db := setupTestDB(t)

		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}

		_, err = QInsert(ctx, tx, DialectSQLite, InsertParams{
			Table: "test_items",
			Values: map[string]any{
				"id": "tx1", "name": "txalpha", "status": "active", "priority": 1,
			},
		})
		if err != nil {
			t.Fatalf("insert 1: %v", err)
		}

		// Second insert with duplicate PK will fail
		_, err = QInsert(ctx, tx, DialectSQLite, InsertParams{
			Table: "test_items",
			Values: map[string]any{
				"id": "tx1", "name": "txduplicate", "status": "active", "priority": 2,
			},
		})
		if err == nil {
			t.Fatal("expected error for duplicate PK")
		}

		if rbErr := tx.Rollback(); rbErr != nil {
			t.Fatalf("rollback: %v", rbErr)
		}

		count, err := QCount(ctx, db, DialectSQLite, "test_items", nil)
		if err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected 0 rows after rollback, got %d", count)
		}
	})
}
