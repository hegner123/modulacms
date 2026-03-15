package dbmetrics

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/utility"
)

func findMetric(snap map[string]utility.Metric, prefix string, wantLabels map[string]string) (utility.Metric, bool) {
	for key, m := range snap {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		match := true
		for k, v := range wantLabels {
			if m.Labels[k] != v {
				match = false
				break
			}
		}
		if match {
			return m, true
		}
	}
	return utility.Metric{}, false
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3-metrics", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestDriverWrapper_CapturesExec(t *testing.T) {
	utility.GlobalMetrics.Reset()
	db := openTestDB(t)

	_, err := db.ExecContext(context.Background(), "CREATE TABLE test_driver (id TEXT PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	snap := utility.GlobalMetrics.GetSnapshot()
	m, ok := findMetric(snap, utility.MetricDBQueries, map[string]string{
		"operation": "create",
		"driver":    "sqlite",
	})
	if !ok {
		t.Fatal("expected db.queries metric for CREATE")
	}
	if m.Value != 1 {
		t.Errorf("expected 1, got %f", m.Value)
	}
}

func TestDriverWrapper_CapturesQuery(t *testing.T) {
	utility.GlobalMetrics.Reset()
	db := openTestDB(t)

	db.ExecContext(context.Background(), "CREATE TABLE test_q (id TEXT PRIMARY KEY, val INTEGER)")
	db.ExecContext(context.Background(), "INSERT INTO test_q (id, val) VALUES ('a', 1)")
	utility.GlobalMetrics.Reset() // reset after setup

	rows, err := db.QueryContext(context.Background(), "SELECT id, val FROM test_q WHERE val = ?", 1)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	rows.Close()

	snap := utility.GlobalMetrics.GetSnapshot()
	_, ok := findMetric(snap, utility.MetricDBQueries, map[string]string{
		"operation": "select",
		"table":     "test_q",
		"driver":    "sqlite",
	})
	if !ok {
		t.Fatal("expected db.queries metric for SELECT on test_q")
	}

	_, ok = findMetric(snap, utility.MetricDBDuration, map[string]string{
		"operation": "select",
		"table":     "test_q",
	})
	if !ok {
		t.Fatal("expected db.duration metric for SELECT")
	}
}

func TestDriverWrapper_CapturesErrors(t *testing.T) {
	utility.GlobalMetrics.Reset()
	db := openTestDB(t)

	_, err := db.ExecContext(context.Background(), "INSERT INTO nonexistent_table (id) VALUES ('x')")
	if err == nil {
		t.Fatal("expected error for nonexistent table")
	}

	snap := utility.GlobalMetrics.GetSnapshot()
	_, ok := findMetric(snap, utility.MetricDBErrors, map[string]string{
		"driver": "sqlite",
	})
	if !ok {
		t.Fatal("expected db.errors metric on failed query")
	}
}

func TestDriverWrapper_CapturesTransactionQueries(t *testing.T) {
	utility.GlobalMetrics.Reset()
	db := openTestDB(t)

	db.ExecContext(context.Background(), "CREATE TABLE test_tx (id TEXT PRIMARY KEY, val TEXT)")
	utility.GlobalMetrics.Reset()

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	_, err = tx.ExecContext(context.Background(), "INSERT INTO test_tx (id, val) VALUES ('t1', 'hello')")
	if err != nil {
		t.Fatalf("insert in tx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	snap := utility.GlobalMetrics.GetSnapshot()
	_, ok := findMetric(snap, utility.MetricDBQueries, map[string]string{
		"operation": "insert",
		"table":     "test_tx",
	})
	if !ok {
		t.Fatal("expected db.queries metric for INSERT within transaction")
	}
}

func TestDriverWrapper_NoDuplicateWithQSelect(t *testing.T) {
	utility.GlobalMetrics.Reset()
	db := openTestDB(t)

	db.ExecContext(context.Background(), "CREATE TABLE test_dup (id TEXT PRIMARY KEY, name TEXT)")
	db.ExecContext(context.Background(), "INSERT INTO test_dup (id, name) VALUES ('1', 'alice')")
	utility.GlobalMetrics.Reset()

	// Use the query builder's QSelect — should record exactly 1 metric (from driver), not 2.
	rows, err := db.QueryContext(context.Background(), "SELECT id, name FROM test_dup")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	rows.Close()

	snap := utility.GlobalMetrics.GetSnapshot()
	m, ok := findMetric(snap, utility.MetricDBQueries, map[string]string{
		"operation": "select",
		"table":     "test_dup",
	})
	if !ok {
		t.Fatal("expected metric")
	}
	if m.Value != 1 {
		t.Errorf("expected exactly 1 count (no double-counting), got %f", m.Value)
	}
}

func TestDriverWrapper_PreparedStatement(t *testing.T) {
	utility.GlobalMetrics.Reset()
	db := openTestDB(t)

	db.ExecContext(context.Background(), "CREATE TABLE test_ps (id TEXT PRIMARY KEY, val INTEGER)")
	utility.GlobalMetrics.Reset()

	stmt, err := db.PrepareContext(context.Background(), "INSERT INTO test_ps (id, val) VALUES (?, ?)")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	defer stmt.Close()

	for i := range 3 {
		_, err := stmt.ExecContext(context.Background(), string(rune('a'+i)), i)
		if err != nil {
			t.Fatalf("exec prepared %d: %v", i, err)
		}
	}

	snap := utility.GlobalMetrics.GetSnapshot()
	m, ok := findMetric(snap, utility.MetricDBQueries, map[string]string{
		"operation": "insert",
		"table":     "test_ps",
	})
	if !ok {
		t.Fatal("expected db.queries for prepared statement executions")
	}
	if m.Value != 3 {
		t.Errorf("expected 3 from 3 prepared execs, got %f", m.Value)
	}
}
