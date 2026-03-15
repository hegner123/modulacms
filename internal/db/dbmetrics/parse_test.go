package dbmetrics

import (
	"testing"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantOp    string
		wantTable string
	}{
		// SELECT
		{
			name:      "simple select",
			query:     "SELECT * FROM users WHERE id = ?",
			wantOp:    "select",
			wantTable: "users",
		},
		{
			name:      "select with qualified column",
			query:     `SELECT u.id FROM "users" u`,
			wantOp:    "select",
			wantTable: "users",
		},
		{
			name:      "select lowercase",
			query:     "select * from tasks where status = 1",
			wantOp:    "select",
			wantTable: "tasks",
		},
		{
			name:      "select with schema prefix",
			query:     "SELECT * FROM public.users WHERE id = $1",
			wantOp:    "select",
			wantTable: "users",
		},
		{
			name:      "select with backtick-quoted table",
			query:     "SELECT * FROM `users` WHERE id = 1",
			wantOp:    "select",
			wantTable: "users",
		},
		{
			name:      "select count",
			query:     "SELECT COUNT(*) FROM sessions",
			wantOp:    "select",
			wantTable: "sessions",
		},
		{
			name:      "select with join",
			query:     "SELECT u.id, r.name FROM users u JOIN roles r ON u.role_id = r.id",
			wantOp:    "select",
			wantTable: "users",
		},

		// INSERT
		{
			name:      "simple insert",
			query:     "INSERT INTO tasks (id, name) VALUES (?, ?)",
			wantOp:    "insert",
			wantTable: "tasks",
		},
		{
			name:      "insert with schema",
			query:     "INSERT INTO public.content_data (id) VALUES ($1)",
			wantOp:    "insert",
			wantTable: "content_data",
		},
		{
			name:      "insert with backticks",
			query:     "INSERT INTO `change_events` (event_id) VALUES (?)",
			wantOp:    "insert",
			wantTable: "change_events",
		},

		// UPDATE
		{
			name:      "simple update",
			query:     "UPDATE sessions SET token = ? WHERE id = ?",
			wantOp:    "update",
			wantTable: "sessions",
		},
		{
			name:      "update with schema",
			query:     "UPDATE public.users SET name = $1 WHERE id = $2",
			wantOp:    "update",
			wantTable: "users",
		},
		{
			name:      "update with backtick-quoted table",
			query:     "UPDATE `datatypes` SET status = 1 WHERE id = ?",
			wantOp:    "update",
			wantTable: "datatypes",
		},

		// DELETE
		{
			name:      "simple delete",
			query:     "DELETE FROM old_data WHERE created_at < ?",
			wantOp:    "delete",
			wantTable: "old_data",
		},
		{
			name:      "delete with double-quoted table",
			query:     `DELETE FROM "media_dimensions" WHERE media_id = $1`,
			wantOp:    "delete",
			wantTable: "media_dimensions",
		},

		// CREATE
		{
			name:      "create table",
			query:     "CREATE TABLE foo (id TEXT PRIMARY KEY, name TEXT NOT NULL)",
			wantOp:    "create",
			wantTable: "foo",
		},
		{
			name:      "create table if not exists",
			query:     "CREATE TABLE IF NOT EXISTS bar (id INTEGER PRIMARY KEY)",
			wantOp:    "create",
			wantTable: "bar",
		},
		{
			name:      "create index",
			query:     "CREATE INDEX idx_users_email ON users (email)",
			wantOp:    "create",
			wantTable: "users",
		},

		// ALTER
		{
			name:      "alter table",
			query:     "ALTER TABLE users ADD COLUMN avatar TEXT",
			wantOp:    "alter",
			wantTable: "users",
		},

		// DROP
		{
			name:      "drop table",
			query:     "DROP TABLE IF EXISTS sessions",
			wantOp:    "drop",
			wantTable: "sessions",
		},

		// PRAGMA
		{
			name:      "pragma table_info",
			query:     "PRAGMA table_info(users)",
			wantOp:    "pragma",
			wantTable: "table_info",
		},
		{
			name:      "pragma foreign_keys",
			query:     "PRAGMA foreign_keys = ON",
			wantOp:    "pragma",
			wantTable: "foreign_keys",
		},

		// WITH (CTE)
		{
			name:      "with cte select",
			query:     "WITH cte AS (SELECT id FROM tasks) SELECT * FROM cte",
			wantOp:    "select",
			wantTable: "cte",
		},
		{
			name:      "with cte insert",
			query:     "WITH src AS (SELECT 1 AS id) INSERT INTO targets (id) SELECT id FROM src",
			wantOp:    "insert",
			wantTable: "targets",
		},

		// Comments
		{
			name:      "line comment before select",
			query:     "-- fetch users\nSELECT * FROM users",
			wantOp:    "select",
			wantTable: "users",
		},
		{
			name:      "block comment inline",
			query:     "SELECT /* all columns */ * FROM /* table */ roles WHERE id = ?",
			wantOp:    "select",
			wantTable: "roles",
		},

		// Edge cases
		{
			name:      "empty string",
			query:     "",
			wantOp:    "other",
			wantTable: "",
		},
		{
			name:      "whitespace only",
			query:     "   \t\n  ",
			wantOp:    "other",
			wantTable: "",
		},
		{
			name:      "unknown statement",
			query:     "EXPLAIN SELECT * FROM users",
			wantOp:    "other",
			wantTable: "",
		},
		{
			name:      "select with subquery in where",
			query:     "SELECT * FROM users WHERE id IN (SELECT user_id FROM sessions)",
			wantOp:    "select",
			wantTable: "users",
		},

		// Mixed case
		{
			name:      "mixed case select",
			query:     "Select * From content_fields Where datatype_id = ?",
			wantOp:    "select",
			wantTable: "content_fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQuery(tt.query)
			if got.Operation != tt.wantOp {
				t.Errorf("Operation = %q, want %q", got.Operation, tt.wantOp)
			}
			if got.Table != tt.wantTable {
				t.Errorf("Table = %q, want %q", got.Table, tt.wantTable)
			}
		})
	}
}

func BenchmarkParseQuery(b *testing.B) {
	queries := []string{
		"SELECT * FROM users WHERE id = ?",
		"INSERT INTO tasks (id, name, status) VALUES (?, ?, ?)",
		"UPDATE sessions SET token = ? WHERE id = ?",
		"DELETE FROM old_data WHERE created_at < ?",
		"CREATE TABLE IF NOT EXISTS foo (id TEXT PRIMARY KEY)",
		"WITH cte AS (SELECT id FROM tasks) SELECT * FROM cte",
	}

	b.ResetTimer()
	for range b.N {
		for _, q := range queries {
			ParseQuery(q)
		}
	}
}
