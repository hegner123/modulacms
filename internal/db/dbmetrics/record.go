package dbmetrics

import (
	"time"

	"github.com/hegner123/modulacms/internal/utility"
)

// RecordQueryMetrics records db.queries, db.duration, and (on error) db.errors
// metrics for a completed database query. The driver string should be one of
// "sqlite", "mysql", or "postgres".
func RecordQueryMetrics(query string, driver string, duration time.Duration, err error) {
	info := ParseQuery(query)
	labels := utility.Labels{
		"operation": info.Operation,
		"table":     info.Table,
		"driver":    driver,
	}

	utility.GlobalMetrics.Increment(utility.MetricDBQueries, labels)
	utility.GlobalMetrics.Timing(utility.MetricDBDuration, duration, labels)

	if err != nil {
		utility.GlobalMetrics.Increment(utility.MetricDBErrors, labels)
	}
}
