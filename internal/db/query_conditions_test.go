package db

import (
	"fmt"
	"strings"
	"testing"
)

func TestCompare_Build(t *testing.T) {
	tests := []struct {
		name      string
		cond      Compare
		dialect   Dialect
		argOffset int
		wantSQL   string
		wantArgs  []any
		wantNext  int
		wantErr   string
	}{
		{
			name:      "eq sqlite",
			cond:      Compare{Column: "status", Op: OpEq, Value: "active"},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"status" = ?`,
			wantArgs:  []any{"active"},
			wantNext:  2,
		},
		{
			name:      "eq mysql",
			cond:      Compare{Column: "status", Op: OpEq, Value: "active"},
			dialect:   DialectMySQL,
			argOffset: 1,
			wantSQL:   "`status` = ?",
			wantArgs:  []any{"active"},
			wantNext:  2,
		},
		{
			name:      "eq postgres",
			cond:      Compare{Column: "status", Op: OpEq, Value: "active"},
			dialect:   DialectPostgres,
			argOffset: 1,
			wantSQL:   `"status" = $1`,
			wantArgs:  []any{"active"},
			wantNext:  2,
		},
		{
			name:      "neq emits <> not !=",
			cond:      Compare{Column: "status", Op: OpNeq, Value: "deleted"},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"status" <> ?`,
			wantArgs:  []any{"deleted"},
			wantNext:  2,
		},
		{
			name:      "gt",
			cond:      Compare{Column: "priority", Op: OpGt, Value: 5},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"priority" > ?`,
			wantArgs:  []any{5},
			wantNext:  2,
		},
		{
			name:      "lt",
			cond:      Compare{Column: "priority", Op: OpLt, Value: 10},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"priority" < ?`,
			wantArgs:  []any{10},
			wantNext:  2,
		},
		{
			name:      "gte",
			cond:      Compare{Column: "priority", Op: OpGte, Value: 3},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"priority" >= ?`,
			wantArgs:  []any{3},
			wantNext:  2,
		},
		{
			name:      "lte",
			cond:      Compare{Column: "priority", Op: OpLte, Value: 7},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"priority" <= ?`,
			wantArgs:  []any{7},
			wantNext:  2,
		},
		{
			name:      "like",
			cond:      Compare{Column: "name", Op: OpLike, Value: "%test%"},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"name" LIKE ?`,
			wantArgs:  []any{"%test%"},
			wantNext:  2,
		},
		{
			name:      "postgres offset 5",
			cond:      Compare{Column: "age", Op: OpGt, Value: 18},
			dialect:   DialectPostgres,
			argOffset: 5,
			wantSQL:   `"age" > $5`,
			wantArgs:  []any{18},
			wantNext:  6,
		},
		{
			name:      "nil value rejected",
			cond:      Compare{Column: "status", Op: OpEq, Value: nil},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "non-nil value",
		},
		{
			name:      "invalid operator",
			cond:      Compare{Column: "status", Op: "DROP", Value: "x"},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "invalid compare operator",
		},
		{
			name:      "invalid column",
			cond:      Compare{Column: "DROP", Op: OpEq, Value: "x"},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "invalid column",
		},
		{
			name:      "empty column",
			cond:      Compare{Column: "", Op: OpEq, Value: "x"},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "invalid column",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewBuildContext()
			sql, args, next, err := tt.cond.Build(ctx, tt.dialect, tt.argOffset)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sql != tt.wantSQL {
				t.Errorf("sql = %q, want %q", sql, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args len = %d, want %d", len(args), len(tt.wantArgs))
			}
			for i, a := range args {
				if a != tt.wantArgs[i] {
					t.Errorf("args[%d] = %v, want %v", i, a, tt.wantArgs[i])
				}
			}
			if next != tt.wantNext {
				t.Errorf("nextOffset = %d, want %d", next, tt.wantNext)
			}
		})
	}
}

func TestInCondition_Build(t *testing.T) {
	tests := []struct {
		name      string
		cond      InCondition
		dialect   Dialect
		argOffset int
		wantSQL   string
		wantArgs  []any
		wantNext  int
		wantErr   string
	}{
		{
			name:      "sqlite 3 values",
			cond:      InCondition{Column: "status", Values: []any{"a", "b", "c"}},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"status" IN (?, ?, ?)`,
			wantArgs:  []any{"a", "b", "c"},
			wantNext:  4,
		},
		{
			name:      "mysql 3 values",
			cond:      InCondition{Column: "status", Values: []any{"a", "b", "c"}},
			dialect:   DialectMySQL,
			argOffset: 1,
			wantSQL:   "`status` IN (?, ?, ?)",
			wantArgs:  []any{"a", "b", "c"},
			wantNext:  4,
		},
		{
			name:      "postgres 3 values offset 3",
			cond:      InCondition{Column: "status", Values: []any{"a", "b", "c"}},
			dialect:   DialectPostgres,
			argOffset: 3,
			wantSQL:   `"status" IN ($3, $4, $5)`,
			wantArgs:  []any{"a", "b", "c"},
			wantNext:  6,
		},
		{
			name:      "single value",
			cond:      InCondition{Column: "id", Values: []any{42}},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"id" IN (?)`,
			wantArgs:  []any{42},
			wantNext:  2,
		},
		{
			name:      "empty values rejected",
			cond:      InCondition{Column: "status", Values: []any{}},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "at least one value",
		},
		{
			name:      "nil values rejected",
			cond:      InCondition{Column: "status", Values: nil},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "at least one value",
		},
		{
			name:      "exceeds max values",
			cond:      InCondition{Column: "status", Values: make([]any, MaxInValues+1)},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "exceeds maximum values",
		},
		{
			name:      "invalid column",
			cond:      InCondition{Column: "1bad", Values: []any{1}},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "invalid column",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewBuildContext()
			sql, args, next, err := tt.cond.Build(ctx, tt.dialect, tt.argOffset)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sql != tt.wantSQL {
				t.Errorf("sql = %q, want %q", sql, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args len = %d, want %d", len(args), len(tt.wantArgs))
			}
			for i, a := range args {
				if a != tt.wantArgs[i] {
					t.Errorf("args[%d] = %v, want %v", i, a, tt.wantArgs[i])
				}
			}
			if next != tt.wantNext {
				t.Errorf("nextOffset = %d, want %d", next, tt.wantNext)
			}
		})
	}
}

func TestBetweenCondition_Build(t *testing.T) {
	tests := []struct {
		name      string
		cond      BetweenCondition
		dialect   Dialect
		argOffset int
		wantSQL   string
		wantArgs  []any
		wantNext  int
		wantErr   string
	}{
		{
			name:      "sqlite",
			cond:      BetweenCondition{Column: "priority", Low: 1, High: 10},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"priority" BETWEEN ? AND ?`,
			wantArgs:  []any{1, 10},
			wantNext:  3,
		},
		{
			name:      "mysql",
			cond:      BetweenCondition{Column: "priority", Low: 1, High: 10},
			dialect:   DialectMySQL,
			argOffset: 1,
			wantSQL:   "`priority` BETWEEN ? AND ?",
			wantArgs:  []any{1, 10},
			wantNext:  3,
		},
		{
			name:      "postgres offset 4",
			cond:      BetweenCondition{Column: "priority", Low: 1, High: 10},
			dialect:   DialectPostgres,
			argOffset: 4,
			wantSQL:   `"priority" BETWEEN $4 AND $5`,
			wantArgs:  []any{1, 10},
			wantNext:  6,
		},
		{
			name:      "string values",
			cond:      BetweenCondition{Column: "date_created", Low: "2024-01-01", High: "2024-12-31"},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `"date_created" BETWEEN ? AND ?`,
			wantArgs:  []any{"2024-01-01", "2024-12-31"},
			wantNext:  3,
		},
		{
			name:      "invalid column",
			cond:      BetweenCondition{Column: "SELECT", Low: 1, High: 10},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "invalid column",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewBuildContext()
			sql, args, next, err := tt.cond.Build(ctx, tt.dialect, tt.argOffset)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sql != tt.wantSQL {
				t.Errorf("sql = %q, want %q", sql, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args len = %d, want %d", len(args), len(tt.wantArgs))
			}
			for i, a := range args {
				if a != tt.wantArgs[i] {
					t.Errorf("args[%d] = %v, want %v", i, a, tt.wantArgs[i])
				}
			}
			if next != tt.wantNext {
				t.Errorf("nextOffset = %d, want %d", next, tt.wantNext)
			}
		})
	}
}

func TestIsNullCondition_Build(t *testing.T) {
	dialects := []struct {
		name    string
		dialect Dialect
		wantSQL string
	}{
		{"sqlite", DialectSQLite, `"description" IS NULL`},
		{"mysql", DialectMySQL, "`description` IS NULL"},
		{"postgres", DialectPostgres, `"description" IS NULL`},
	}

	for _, dd := range dialects {
		t.Run(dd.name, func(t *testing.T) {
			ctx := NewBuildContext()
			cond := IsNullCondition{Column: "description"}
			sql, args, next, err := cond.Build(ctx, dd.dialect, 5)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sql != dd.wantSQL {
				t.Errorf("sql = %q, want %q", sql, dd.wantSQL)
			}
			if len(args) != 0 {
				t.Errorf("expected 0 args, got %d", len(args))
			}
			if next != 5 {
				t.Errorf("nextOffset = %d, want 5 (unchanged)", next)
			}
		})
	}

	t.Run("invalid column", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := IsNullCondition{Column: ""}
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err == nil {
			t.Fatal("expected error for empty column")
		}
	})
}

func TestIsNotNullCondition_Build(t *testing.T) {
	dialects := []struct {
		name    string
		dialect Dialect
		wantSQL string
	}{
		{"sqlite", DialectSQLite, `"description" IS NOT NULL`},
		{"mysql", DialectMySQL, "`description` IS NOT NULL"},
		{"postgres", DialectPostgres, `"description" IS NOT NULL`},
	}

	for _, dd := range dialects {
		t.Run(dd.name, func(t *testing.T) {
			ctx := NewBuildContext()
			cond := IsNotNullCondition{Column: "description"}
			sql, args, next, err := cond.Build(ctx, dd.dialect, 3)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sql != dd.wantSQL {
				t.Errorf("sql = %q, want %q", sql, dd.wantSQL)
			}
			if len(args) != 0 {
				t.Errorf("expected 0 args, got %d", len(args))
			}
			if next != 3 {
				t.Errorf("nextOffset = %d, want 3 (unchanged)", next)
			}
		})
	}
}

func TestAnd_Build(t *testing.T) {
	t.Run("two children sqlite", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := And{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "active"},
			Compare{Column: "priority", Op: OpGt, Value: 5},
		}}
		sql, args, next, err := cond.Build(ctx, DialectSQLite, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != `("status" = ? AND "priority" > ?)` {
			t.Errorf("sql = %q", sql)
		}
		if len(args) != 2 || args[0] != "active" || args[1] != 5 {
			t.Errorf("args = %v", args)
		}
		if next != 3 {
			t.Errorf("nextOffset = %d, want 3", next)
		}
	})

	t.Run("two children postgres", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := And{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "active"},
			Compare{Column: "priority", Op: OpGt, Value: 5},
		}}
		sql, args, next, err := cond.Build(ctx, DialectPostgres, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != `("status" = $1 AND "priority" > $2)` {
			t.Errorf("sql = %q", sql)
		}
		if len(args) != 2 || args[0] != "active" || args[1] != 5 {
			t.Errorf("args = %v", args)
		}
		if next != 3 {
			t.Errorf("nextOffset = %d, want 3", next)
		}
	})

	t.Run("single child valid", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := And{Conditions: []Condition{
			Compare{Column: "x", Op: OpEq, Value: 1},
		}}
		sql, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != `("x" = ?)` {
			t.Errorf("sql = %q", sql)
		}
	})

	t.Run("empty children rejected", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := And{Conditions: []Condition{}}
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err == nil {
			t.Fatal("expected error for empty children")
		}
		if !strings.Contains(err.Error(), "at least one child") {
			t.Errorf("error = %q", err.Error())
		}
	})

	t.Run("nil children rejected", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := And{Conditions: nil}
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err == nil {
			t.Fatal("expected error for nil children")
		}
	})
}

func TestOr_Build(t *testing.T) {
	t.Run("two children sqlite", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := Or{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "active"},
			Compare{Column: "status", Op: OpEq, Value: "draft"},
		}}
		sql, args, next, err := cond.Build(ctx, DialectSQLite, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != `("status" = ? OR "status" = ?)` {
			t.Errorf("sql = %q", sql)
		}
		if len(args) != 2 || args[0] != "active" || args[1] != "draft" {
			t.Errorf("args = %v", args)
		}
		if next != 3 {
			t.Errorf("nextOffset = %d, want 3", next)
		}
	})

	t.Run("two children postgres offset 3", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := Or{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "active"},
			Compare{Column: "status", Op: OpEq, Value: "draft"},
		}}
		sql, args, next, err := cond.Build(ctx, DialectPostgres, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != `("status" = $3 OR "status" = $4)` {
			t.Errorf("sql = %q", sql)
		}
		if len(args) != 2 {
			t.Fatalf("args len = %d, want 2", len(args))
		}
		if next != 5 {
			t.Errorf("nextOffset = %d, want 5", next)
		}
	})

	t.Run("empty children rejected", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := Or{Conditions: []Condition{}}
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err == nil {
			t.Fatal("expected error for empty children")
		}
	})
}

func TestNot_Build(t *testing.T) {
	t.Run("negate compare sqlite", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := Not{Condition: Compare{Column: "status", Op: OpEq, Value: "deleted"}}
		sql, args, next, err := cond.Build(ctx, DialectSQLite, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != `NOT ("status" = ?)` {
			t.Errorf("sql = %q", sql)
		}
		if len(args) != 1 || args[0] != "deleted" {
			t.Errorf("args = %v", args)
		}
		if next != 2 {
			t.Errorf("nextOffset = %d, want 2", next)
		}
	})

	t.Run("negate compare postgres", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := Not{Condition: Compare{Column: "status", Op: OpEq, Value: "deleted"}}
		sql, _, next, err := cond.Build(ctx, DialectPostgres, 7)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != `NOT ("status" = $7)` {
			t.Errorf("sql = %q", sql)
		}
		if next != 8 {
			t.Errorf("nextOffset = %d, want 8", next)
		}
	})

	t.Run("nil condition rejected", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := Not{Condition: nil}
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err == nil {
			t.Fatal("expected error for nil condition")
		}
		if !strings.Contains(err.Error(), "non-nil") {
			t.Errorf("error = %q", err.Error())
		}
	})
}

func TestNestedConditions(t *testing.T) {
	t.Run("AND containing OR sqlite", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := And{Conditions: []Condition{
			Compare{Column: "active", Op: OpEq, Value: true},
			Or{Conditions: []Condition{
				Compare{Column: "role", Op: OpEq, Value: "admin"},
				Compare{Column: "role", Op: OpEq, Value: "editor"},
			}},
		}}
		sql, args, next, err := cond.Build(ctx, DialectSQLite, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `("active" = ? AND ("role" = ? OR "role" = ?))`
		if sql != want {
			t.Errorf("sql = %q, want %q", sql, want)
		}
		if len(args) != 3 {
			t.Fatalf("args len = %d, want 3", len(args))
		}
		if args[0] != true || args[1] != "admin" || args[2] != "editor" {
			t.Errorf("args = %v", args)
		}
		if next != 4 {
			t.Errorf("nextOffset = %d, want 4", next)
		}
	})

	t.Run("AND containing OR postgres", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := And{Conditions: []Condition{
			Compare{Column: "active", Op: OpEq, Value: true},
			Or{Conditions: []Condition{
				Compare{Column: "role", Op: OpEq, Value: "admin"},
				Compare{Column: "role", Op: OpEq, Value: "editor"},
			}},
		}}
		sql, _, next, err := cond.Build(ctx, DialectPostgres, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `("active" = $1 AND ("role" = $2 OR "role" = $3))`
		if sql != want {
			t.Errorf("sql = %q, want %q", sql, want)
		}
		if next != 4 {
			t.Errorf("nextOffset = %d, want 4", next)
		}
	})

	t.Run("NOT wrapping AND", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := Not{Condition: And{Conditions: []Condition{
			Compare{Column: "status", Op: OpEq, Value: "deleted"},
			IsNullCondition{Column: "restored_at"},
		}}}
		sql, args, next, err := cond.Build(ctx, DialectSQLite, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `NOT (("status" = ? AND "restored_at" IS NULL))`
		if sql != want {
			t.Errorf("sql = %q, want %q", sql, want)
		}
		if len(args) != 1 || args[0] != "deleted" {
			t.Errorf("args = %v", args)
		}
		if next != 2 {
			t.Errorf("nextOffset = %d, want 2", next)
		}
	})

	t.Run("complex nested OR of ANDs postgres", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := Or{Conditions: []Condition{
			And{Conditions: []Condition{
				Compare{Column: "status", Op: OpEq, Value: "active"},
				Compare{Column: "priority", Op: OpGt, Value: 5},
			}},
			And{Conditions: []Condition{
				Compare{Column: "status", Op: OpEq, Value: "urgent"},
				Compare{Column: "priority", Op: OpGt, Value: 0},
			}},
		}}
		sql, args, next, err := cond.Build(ctx, DialectPostgres, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `(("status" = $1 AND "priority" > $2) OR ("status" = $3 AND "priority" > $4))`
		if sql != want {
			t.Errorf("sql = %q, want %q", sql, want)
		}
		if len(args) != 4 {
			t.Fatalf("args len = %d, want 4", len(args))
		}
		if args[0] != "active" || args[1] != 5 || args[2] != "urgent" || args[3] != 0 {
			t.Errorf("args = %v", args)
		}
		if next != 5 {
			t.Errorf("nextOffset = %d, want 5", next)
		}
	})

	t.Run("mixed leaf types in AND", func(t *testing.T) {
		ctx := NewBuildContext()
		cond := And{Conditions: []Condition{
			Compare{Column: "name", Op: OpLike, Value: "%test%"},
			InCondition{Column: "status", Values: []any{"a", "b"}},
			BetweenCondition{Column: "priority", Low: 1, High: 10},
			IsNotNullCondition{Column: "email"},
		}}
		sql, args, next, err := cond.Build(ctx, DialectPostgres, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `("name" LIKE $1 AND "status" IN ($2, $3) AND "priority" BETWEEN $4 AND $5 AND "email" IS NOT NULL)`
		if sql != want {
			t.Errorf("sql = %q, want %q", sql, want)
		}
		if len(args) != 5 {
			t.Fatalf("args len = %d, want 5", len(args))
		}
		if next != 6 {
			t.Errorf("nextOffset = %d, want 6", next)
		}
	})
}

func TestCondition_DepthLimit(t *testing.T) {
	t.Run("depth exactly at limit succeeds", func(t *testing.T) {
		// Build a chain of depth exactly MaxConditionDepth
		var cond Condition = Compare{Column: "x", Op: OpEq, Value: 1}
		for range MaxConditionDepth {
			cond = And{Conditions: []Condition{cond}}
		}
		ctx := NewBuildContext()
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err != nil {
			t.Fatalf("depth %d should succeed, got: %v", MaxConditionDepth, err)
		}
	})

	t.Run("depth exceeds limit fails", func(t *testing.T) {
		// Build a chain of depth MaxConditionDepth+1
		var cond Condition = Compare{Column: "x", Op: OpEq, Value: 1}
		for range MaxConditionDepth + 1 {
			cond = And{Conditions: []Condition{cond}}
		}
		ctx := NewBuildContext()
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err == nil {
			t.Fatal("expected depth limit error")
		}
		if !strings.Contains(err.Error(), "maximum depth") {
			t.Errorf("error = %q", err.Error())
		}
	})
}

func TestCondition_NodeCountLimit(t *testing.T) {
	t.Run("exactly at limit succeeds", func(t *testing.T) {
		// Build flat AND with MaxConditionNodes-1 children (1 And + N-1 Compare = N nodes)
		children := make([]Condition, MaxConditionNodes-1)
		for i := range children {
			children[i] = Compare{Column: "x", Op: OpEq, Value: i}
		}
		cond := And{Conditions: children}
		ctx := NewBuildContext()
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err != nil {
			t.Fatalf("exactly %d nodes should succeed, got: %v", MaxConditionNodes, err)
		}
	})

	t.Run("exceeds limit fails", func(t *testing.T) {
		// Build flat AND with MaxConditionNodes children (1 And + N Compare = N+1 nodes)
		children := make([]Condition, MaxConditionNodes)
		for i := range children {
			children[i] = Compare{Column: "x", Op: OpEq, Value: i}
		}
		cond := And{Conditions: children}
		ctx := NewBuildContext()
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err == nil {
			t.Fatal("expected node count limit error")
		}
		if !strings.Contains(err.Error(), "maximum node count") {
			t.Errorf("error = %q", err.Error())
		}
	})
}

func TestCondition_InMaxValues(t *testing.T) {
	t.Run("exactly at limit succeeds", func(t *testing.T) {
		vals := make([]any, MaxInValues)
		for i := range vals {
			vals[i] = i
		}
		cond := InCondition{Column: "x", Values: vals}
		ctx := NewBuildContext()
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err != nil {
			t.Fatalf("exactly %d IN values should succeed, got: %v", MaxInValues, err)
		}
	})

	t.Run("exceeds limit fails", func(t *testing.T) {
		vals := make([]any, MaxInValues+1)
		for i := range vals {
			vals[i] = i
		}
		cond := InCondition{Column: "x", Values: vals}
		ctx := NewBuildContext()
		_, _, _, err := cond.Build(ctx, DialectSQLite, 1)
		if err == nil {
			t.Fatal("expected IN max values error")
		}
		if !strings.Contains(err.Error(), "exceeds maximum values") {
			t.Errorf("error = %q", err.Error())
		}
	})
}

func TestValidateCondition(t *testing.T) {
	t.Run("valid compare", func(t *testing.T) {
		err := ValidateCondition(Compare{Column: "x", Op: OpEq, Value: 1})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nil condition", func(t *testing.T) {
		err := ValidateCondition(nil)
		if err == nil {
			t.Fatal("expected error for nil condition")
		}
	})

	t.Run("invalid column caught", func(t *testing.T) {
		err := ValidateCondition(Compare{Column: "DROP", Op: OpEq, Value: 1})
		if err == nil {
			t.Fatal("expected error for SQL keyword column")
		}
	})

	t.Run("nested invalid caught", func(t *testing.T) {
		err := ValidateCondition(And{Conditions: []Condition{
			Compare{Column: "good", Op: OpEq, Value: 1},
			Compare{Column: "", Op: OpEq, Value: 2},
		}})
		if err == nil {
			t.Fatal("expected error for empty column in nested condition")
		}
	})
}

func TestHasValueBinding(t *testing.T) {
	tests := []struct {
		name string
		cond Condition
		want bool
	}{
		{"compare", Compare{Column: "x", Op: OpEq, Value: 1}, true},
		{"in", InCondition{Column: "x", Values: []any{1}}, true},
		{"between", BetweenCondition{Column: "x", Low: 1, High: 10}, true},
		{"is null", IsNullCondition{Column: "x"}, false},
		{"is not null", IsNotNullCondition{Column: "x"}, false},
		{
			"and with compare",
			And{Conditions: []Condition{
				IsNullCondition{Column: "a"},
				Compare{Column: "b", Op: OpEq, Value: 1},
			}},
			true,
		},
		{
			"and with only nulls",
			And{Conditions: []Condition{
				IsNullCondition{Column: "a"},
				IsNotNullCondition{Column: "b"},
			}},
			false,
		},
		{
			"or with compare",
			Or{Conditions: []Condition{
				IsNullCondition{Column: "a"},
				Compare{Column: "b", Op: OpEq, Value: 1},
			}},
			true,
		},
		{
			"or with only nulls",
			Or{Conditions: []Condition{
				IsNullCondition{Column: "a"},
				IsNotNullCondition{Column: "b"},
			}},
			false,
		},
		{
			"not with compare",
			Not{Condition: Compare{Column: "x", Op: OpEq, Value: 1}},
			true,
		},
		{
			"not with is null",
			Not{Condition: IsNullCondition{Column: "x"}},
			false,
		},
		{
			"deeply nested with value",
			And{Conditions: []Condition{
				Or{Conditions: []Condition{
					Not{Condition: InCondition{Column: "x", Values: []any{1, 2}}},
				}},
			}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasValueBinding(tt.cond)
			if got != tt.want {
				t.Errorf("HasValueBinding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCondition_AllDialects(t *testing.T) {
	// Verify a compound condition produces correct SQL across all 3 dialects.
	cond := And{Conditions: []Condition{
		Compare{Column: "status", Op: OpEq, Value: "active"},
		InCondition{Column: "role", Values: []any{"admin", "editor"}},
		BetweenCondition{Column: "priority", Low: 1, High: 10},
		IsNotNullCondition{Column: "email"},
	}}

	dialects := []struct {
		name    string
		dialect Dialect
		wantSQL string
	}{
		{
			"sqlite",
			DialectSQLite,
			`("status" = ? AND "role" IN (?, ?) AND "priority" BETWEEN ? AND ? AND "email" IS NOT NULL)`,
		},
		{
			"mysql",
			DialectMySQL,
			"(`status` = ? AND `role` IN (?, ?) AND `priority` BETWEEN ? AND ? AND `email` IS NOT NULL)",
		},
		{
			"postgres",
			DialectPostgres,
			`("status" = $1 AND "role" IN ($2, $3) AND "priority" BETWEEN $4 AND $5 AND "email" IS NOT NULL)`,
		},
	}

	for _, dd := range dialects {
		t.Run(dd.name, func(t *testing.T) {
			ctx := NewBuildContext()
			sql, args, next, err := cond.Build(ctx, dd.dialect, 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sql != dd.wantSQL {
				t.Errorf("sql = %q, want %q", sql, dd.wantSQL)
			}
			// 1 (status) + 2 (role IN) + 2 (between) = 5 args
			if len(args) != 5 {
				t.Fatalf("args len = %d, want 5", len(args))
			}
			if args[0] != "active" {
				t.Errorf("args[0] = %v, want active", args[0])
			}
			if args[1] != "admin" || args[2] != "editor" {
				t.Errorf("args[1:3] = %v", args[1:3])
			}
			if args[3] != 1 || args[4] != 10 {
				t.Errorf("args[3:5] = %v", args[3:5])
			}
			// 5 value bindings + 1 IS NOT NULL (no arg) = next offset 6
			if next != 6 {
				t.Errorf("nextOffset = %d, want 6", next)
			}
		})
	}
}

func TestCondition_PostgresPlaceholderSequence(t *testing.T) {
	// Verify $N numbering is contiguous across a complex tree.
	cond := Or{Conditions: []Condition{
		And{Conditions: []Condition{
			Compare{Column: "a", Op: OpEq, Value: "v1"},
			InCondition{Column: "b", Values: []any{"v2", "v3"}},
		}},
		And{Conditions: []Condition{
			BetweenCondition{Column: "c", Low: 10, High: 20},
			Compare{Column: "d", Op: OpNeq, Value: "v4"},
		}},
	}}

	ctx := NewBuildContext()
	sql, args, next, err := cond.Build(ctx, DialectPostgres, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `(("a" = $1 AND "b" IN ($2, $3)) OR ("c" BETWEEN $4 AND $5 AND "d" <> $6))`
	if sql != want {
		t.Errorf("sql = %q, want %q", sql, want)
	}
	if len(args) != 6 {
		t.Fatalf("args len = %d, want 6", len(args))
	}
	if next != 7 {
		t.Errorf("nextOffset = %d, want 7", next)
	}

	// Verify each $N appears exactly once and in sequence
	for i := 1; i <= 6; i++ {
		ph := fmt.Sprintf("$%d", i)
		count := strings.Count(sql, ph)
		if count != 1 {
			t.Errorf("placeholder %s appears %d times, want 1", ph, count)
		}
	}
}

// ===== AggregateCondition Tests =====

func TestAggregateCondition_Build(t *testing.T) {
	tests := []struct {
		name      string
		cond      AggregateCondition
		dialect   Dialect
		argOffset int
		wantSQL   string
		wantArgs  []any
		wantNext  int
		wantErr   string
	}{
		{
			name: "COUNT(*) > 5 sqlite",
			cond: AggregateCondition{
				Agg:   AggregateColumn{Func: "COUNT", Arg: "*"},
				Op:    OpGt,
				Value: 5,
			},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantSQL:   `COUNT(*) > ?`,
			wantArgs:  []any{5},
			wantNext:  2,
		},
		{
			name: "SUM(amount) >= 100 postgres",
			cond: AggregateCondition{
				Agg:   AggregateColumn{Func: "SUM", Arg: "amount"},
				Op:    OpGte,
				Value: 100,
			},
			dialect:   DialectPostgres,
			argOffset: 1,
			wantSQL:   `SUM("amount") >= $1`,
			wantArgs:  []any{100},
			wantNext:  2,
		},
		{
			name: "OpLike rejected",
			cond: AggregateCondition{
				Agg:   AggregateColumn{Func: "COUNT", Arg: "*"},
				Op:    OpLike,
				Value: "x",
			},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "LIKE",
		},
		{
			name: "nil value rejected",
			cond: AggregateCondition{
				Agg:   AggregateColumn{Func: "COUNT", Arg: "*"},
				Op:    OpGt,
				Value: nil,
			},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "non-nil value",
		},
		{
			name: "invalid func rejected",
			cond: AggregateCondition{
				Agg:   AggregateColumn{Func: "INVALID", Arg: "*"},
				Op:    OpGt,
				Value: 1,
			},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "invalid aggregate function",
		},
		{
			name: "star with SUM rejected",
			cond: AggregateCondition{
				Agg:   AggregateColumn{Func: "SUM", Arg: "*"},
				Op:    OpGt,
				Value: 1,
			},
			dialect:   DialectSQLite,
			argOffset: 1,
			wantErr:   "does not support *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewBuildContext()
			sql, args, next, err := tt.cond.Build(ctx, tt.dialect, tt.argOffset)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sql != tt.wantSQL {
				t.Errorf("sql = %q, want %q", sql, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args len = %d, want %d", len(args), len(tt.wantArgs))
			}
			for i, a := range args {
				if a != tt.wantArgs[i] {
					t.Errorf("args[%d] = %v, want %v", i, a, tt.wantArgs[i])
				}
			}
			if next != tt.wantNext {
				t.Errorf("nextOffset = %d, want %d", next, tt.wantNext)
			}
		})
	}
}

func TestAggregateCondition_HasValueBinding(t *testing.T) {
	got := HasValueBinding(AggregateCondition{
		Agg:   AggregateColumn{Func: "COUNT", Arg: "*"},
		Op:    OpGt,
		Value: 5,
	})
	if !got {
		t.Error("HasValueBinding(AggregateCondition{...}) = false, want true")
	}
}
