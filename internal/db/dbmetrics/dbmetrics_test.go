package dbmetrics

import (
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/utility"
)

func TestRecordQueryMetrics(t *testing.T) {
	t.Run("successful query records queries and duration", func(t *testing.T) {
		utility.GlobalMetrics.Reset()

		RecordQueryMetrics("SELECT * FROM users WHERE id = ?", "sqlite", 5*time.Millisecond, nil)

		snap := utility.GlobalMetrics.GetSnapshot()

		// Check counter
		counterKey := utility.MetricDBQueries + ",driver=sqlite,operation=select,table=users"
		m, ok := snap[counterKey]
		if !ok {
			t.Fatalf("missing metric key %q; snapshot keys: %v", counterKey, snapshotKeys(snap))
		}
		if m.Value != 1 {
			t.Errorf("counter value = %v, want 1", m.Value)
		}

		// Check timing
		timingKey := utility.MetricDBDuration + ",driver=sqlite,operation=select,table=users"
		tm, ok := snap[timingKey]
		if !ok {
			t.Fatalf("missing metric key %q", timingKey)
		}
		if tm.Value <= 0 {
			t.Errorf("timing value = %v, want > 0", tm.Value)
		}

		// No error metric should exist
		errKey := utility.MetricDBErrors + ",driver=sqlite,operation=select,table=users"
		if _, ok := snap[errKey]; ok {
			t.Error("error metric should not exist for successful query")
		}
	})

	t.Run("failed query records queries, duration, and errors", func(t *testing.T) {
		utility.GlobalMetrics.Reset()

		RecordQueryMetrics(
			"INSERT INTO tasks (id) VALUES (?)",
			"mysql",
			2*time.Millisecond,
			errForTest("constraint violation"),
		)

		snap := utility.GlobalMetrics.GetSnapshot()

		counterKey := utility.MetricDBQueries + ",driver=mysql,operation=insert,table=tasks"
		if _, ok := snap[counterKey]; !ok {
			t.Fatalf("missing counter metric %q", counterKey)
		}

		errKey := utility.MetricDBErrors + ",driver=mysql,operation=insert,table=tasks"
		em, ok := snap[errKey]
		if !ok {
			t.Fatalf("missing error metric %q", errKey)
		}
		if em.Value != 1 {
			t.Errorf("error counter = %v, want 1", em.Value)
		}
	})

	t.Run("multiple queries accumulate counters", func(t *testing.T) {
		utility.GlobalMetrics.Reset()

		for range 5 {
			RecordQueryMetrics("SELECT * FROM sessions", "postgres", time.Millisecond, nil)
		}

		snap := utility.GlobalMetrics.GetSnapshot()
		counterKey := utility.MetricDBQueries + ",driver=postgres,operation=select,table=sessions"
		m, ok := snap[counterKey]
		if !ok {
			t.Fatalf("missing metric key %q", counterKey)
		}
		if m.Value != 5 {
			t.Errorf("counter value = %v, want 5", m.Value)
		}
	})

	t.Run("different tables produce different metric keys", func(t *testing.T) {
		utility.GlobalMetrics.Reset()

		RecordQueryMetrics("SELECT * FROM users", "sqlite", time.Millisecond, nil)
		RecordQueryMetrics("SELECT * FROM roles", "sqlite", time.Millisecond, nil)

		snap := utility.GlobalMetrics.GetSnapshot()

		usersKey := utility.MetricDBQueries + ",driver=sqlite,operation=select,table=users"
		rolesKey := utility.MetricDBQueries + ",driver=sqlite,operation=select,table=roles"

		if _, ok := snap[usersKey]; !ok {
			t.Errorf("missing users metric key %q", usersKey)
		}
		if _, ok := snap[rolesKey]; !ok {
			t.Errorf("missing roles metric key %q", rolesKey)
		}
	})
}

// snapshotKeys returns the keys of a metric snapshot for debugging.
func snapshotKeys(snap map[string]utility.Metric) []string {
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	return keys
}

// errForTest is a simple error implementation for test use.
type errForTest string

func (e errForTest) Error() string { return string(e) }
