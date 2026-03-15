package deploy

import (
	"context"
	"database/sql"
	"testing"

	"github.com/hegner123/modulacms/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

// newPluginTestDB creates an in-memory SQLite database with a plugin table.
func newPluginTestDB(t *testing.T) *sql.DB {
	t.Helper()
	pool, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	// Create the tables table (used by discoverPluginTables via ListTables).
	_, err = pool.Exec(`
		CREATE TABLE plugin_test_items (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			count INTEGER NOT NULL,
			score REAL,
			created_at TEXT
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestIsPluginTable(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"plugin_blog_posts", true},
		{"plugin_test_items", true},
		{"content_data", false},
		{"users", false},
		{"plugin_", false},
		{"plugin_xy", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPluginTable(tt.name); got != tt.want {
				t.Errorf("isPluginTable(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestColumnsMatch(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want bool
	}{
		{"equal", []string{"id", "name"}, []string{"id", "name"}, true},
		{"different order", []string{"name", "id"}, []string{"id", "name"}, false},
		{"different length", []string{"id"}, []string{"id", "name"}, false},
		{"empty", []string{}, []string{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := columnsMatch(tt.a, tt.b); got != tt.want {
				t.Errorf("columnsMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildIntColumnMap(t *testing.T) {
	cols := []string{"id", "name", "count", "score"}
	meta := []db.ColumnMeta{
		{Name: "id", IsInteger: false},
		{Name: "name", IsInteger: false},
		{Name: "count", IsInteger: true},
		{Name: "score", IsInteger: false},
	}

	intCols := buildIntColumnMap(cols, meta)
	if !intCols[2] {
		t.Error("expected column index 2 (count) to be integer")
	}
	if intCols[0] || intCols[1] || intCols[3] {
		t.Error("non-integer columns should not be in map")
	}
}

func TestCoerceWithIntMap(t *testing.T) {
	rows := [][]any{
		{"abc", "hello", float64(42), float64(3.14)},
		{"def", "world", float64(99), nil},
	}
	intCols := map[int]bool{2: true}

	result := coerceWithIntMap(rows, intCols)
	if result[0][2] != int64(42) {
		t.Errorf("expected int64(42), got %T(%v)", result[0][2], result[0][2])
	}
	if result[1][2] != int64(99) {
		t.Errorf("expected int64(99), got %T(%v)", result[1][2], result[1][2])
	}
	// Non-integer columns untouched.
	if result[0][3] != float64(3.14) {
		t.Errorf("expected float64(3.14), got %T(%v)", result[0][3], result[0][3])
	}
	// Nil stays nil.
	if result[1][3] != nil {
		t.Errorf("expected nil, got %v", result[1][3])
	}
}

func TestCoerceWithIntMap_Empty(t *testing.T) {
	rows := [][]any{{"a", "b"}}
	result := coerceWithIntMap(rows, nil)
	if len(result) != 1 || result[0][0] != "a" {
		t.Error("empty intCols should return rows unchanged")
	}
}

func TestExtractColNames(t *testing.T) {
	cols := []db.ColumnMeta{
		{Name: "id", IsInteger: false},
		{Name: "count", IsInteger: true},
	}
	names := extractColNames(cols)
	if len(names) != 2 || names[0] != "id" || names[1] != "count" {
		t.Errorf("extractColNames returned %v", names)
	}
}

func TestIntrospectColumns_SQLite(t *testing.T) {
	pool := newPluginTestDB(t)
	defer pool.Close()

	ops := &sqliteTestDeployOps{pool: pool}
	ctx := context.Background()

	cols, err := ops.IntrospectColumns(ctx, db.DBTable("plugin_test_items"))
	if err != nil {
		t.Fatalf("IntrospectColumns: %v", err)
	}

	expected := map[string]bool{
		"id":         false,
		"name":       false,
		"count":      true,
		"score":      false,
		"created_at": false,
	}

	if len(cols) != len(expected) {
		t.Fatalf("expected %d columns, got %d", len(expected), len(cols))
	}

	for _, c := range cols {
		wantInt, ok := expected[c.Name]
		if !ok {
			t.Errorf("unexpected column %q", c.Name)
			continue
		}
		if c.IsInteger != wantInt {
			t.Errorf("column %q: IsInteger = %v, want %v", c.Name, c.IsInteger, wantInt)
		}
	}
}

func TestIntrospectColumns_NonexistentTable(t *testing.T) {
	pool := newPluginTestDB(t)
	defer pool.Close()

	ops := &sqliteTestDeployOps{pool: pool}
	ctx := context.Background()

	_, err := ops.IntrospectColumns(ctx, db.DBTable("plugin_nonexistent"))
	if err == nil {
		t.Error("expected error for nonexistent table")
	}
}

func TestValidatePayload_SkipsPluginULIDCheck(t *testing.T) {
	// Plugin table rows with non-ULID IDs should not cause validation errors.
	payload := &SyncPayload{
		Manifest: SyncManifest{
			SchemaVersion: "",
			Tables:        []string{"plugin_blog_posts"},
			RowCounts:     map[string]int{"plugin_blog_posts": 1},
		},
		Tables: map[string]TableData{
			"plugin_blog_posts": {
				Columns: []string{"post_id", "title"},
				Rows:    [][]any{{"not-a-ulid-12345678", "Hello"}},
			},
		},
		UserRefs: map[string]string{},
	}

	// Recompute hash and schema version so those checks pass.
	hash, _ := computePayloadHash(payload.Tables)
	payload.Manifest.PayloadHash = hash
	payload.Manifest.SchemaVersion = computeSchemaVersion(payload.Tables)

	errs := ValidatePayload(payload, nil)
	for _, e := range errs {
		if e.Table == "plugin_blog_posts" && e.Phase == "validate" {
			t.Errorf("unexpected validation error for plugin table: %s", e.Message)
		}
	}
}

func TestValidatePayload_PluginRowWidthCheck(t *testing.T) {
	payload := &SyncPayload{
		Manifest: SyncManifest{
			Tables:    []string{"plugin_blog_posts"},
			RowCounts: map[string]int{"plugin_blog_posts": 1},
		},
		Tables: map[string]TableData{
			"plugin_blog_posts": {
				Columns: []string{"id", "title", "body"},
				Rows:    [][]any{{"1", "Hello"}}, // only 2 values for 3 columns
			},
		},
		UserRefs: map[string]string{},
	}

	hash, _ := computePayloadHash(payload.Tables)
	payload.Manifest.PayloadHash = hash
	payload.Manifest.SchemaVersion = computeSchemaVersion(payload.Tables)

	errs := ValidatePayload(payload, nil)
	found := false
	for _, e := range errs {
		if e.Table == "plugin_blog_posts" && e.Phase == "validate" {
			found = true
		}
	}
	if !found {
		t.Error("expected row width validation error for plugin table with mismatched column count")
	}
}
