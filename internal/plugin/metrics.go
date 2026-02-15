package plugin

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/utility"
)

// Plugin-scoped metric name constants. These are instrumented at coarse-grained
// boundaries (HTTP requests, hook events, reload events) where the recording
// cost is negligible relative to the operation cost.
//
// Deferred until profiling justifies: per-Get()/Put() checkout timing,
// per-DB-op counting. These can be added later behind a Plugin_Verbose_Metrics
// flag if needed.
const (
	PluginMetricHTTPRequests       = "plugin.http.requests"
	PluginMetricHTTPDuration       = "plugin.http.duration_ms"
	PluginMetricHookBefore         = "plugin.hook.before"
	PluginMetricHookAfter          = "plugin.hook.after"
	PluginMetricHookDuration       = "plugin.hook.duration_ms"
	PluginMetricErrors             = "plugin.errors"
	PluginMetricCircuitBreakerTrip = "plugin.circuit_breaker.trip"
	PluginMetricReload             = "plugin.reload"
	PluginMetricVMAvailable        = "plugin.vm.available"
)

// RecordHTTPRequest records a single HTTP request handled by a plugin.
// Called once per HTTP request in ServeHTTP -- negligible cost relative to
// Lua execution.
func RecordHTTPRequest(pluginName, method string, status int, durationMs float64) {
	labels := utility.Labels{
		"plugin": pluginName,
		"method": method,
		"status": fmt.Sprintf("%d", status),
	}
	utility.GlobalMetrics.Increment(PluginMetricHTTPRequests, labels)
	utility.GlobalMetrics.Histogram(PluginMetricHTTPDuration, durationMs, labels)
}

// RecordHookExecution records a single hook execution (before or after).
// Called once per hook event in executeBefore/executeAfter -- negligible cost
// relative to DB + Lua execution.
func RecordHookExecution(metricName, pluginName, event, table, status string, durationMs float64) {
	labels := utility.Labels{
		"plugin": pluginName,
		"event":  event,
		"table":  table,
		"status": status,
	}
	utility.GlobalMetrics.Increment(metricName, labels)
	utility.GlobalMetrics.Histogram(PluginMetricHookDuration, durationMs, labels)
}

// RecordReload records a plugin reload event with its outcome.
func RecordReload(pluginName, status string) {
	labels := utility.Labels{
		"plugin": pluginName,
		"status": status,
	}
	utility.GlobalMetrics.Increment(PluginMetricReload, labels)
}

// RecordCircuitBreakerTrip records when a plugin's circuit breaker trips.
// This is a rare event -- no performance concern.
func RecordCircuitBreakerTrip(pluginName string) {
	labels := utility.Labels{
		"plugin": pluginName,
	}
	utility.GlobalMetrics.Increment(PluginMetricCircuitBreakerTrip, labels)
}

// RecordError records a generic plugin error.
func RecordError(pluginName, errorType string) {
	labels := utility.Labels{
		"plugin":     pluginName,
		"error_type": errorType,
	}
	utility.GlobalMetrics.Increment(PluginMetricErrors, labels)
}

// SnapshotVMAvailability records the current VM availability for all plugins.
// Called periodically by the watcher (every 2s) -- negligible overhead.
func SnapshotVMAvailability(plugins map[string]*PluginInstance) {
	for name, inst := range plugins {
		if inst.Pool == nil {
			continue
		}
		labels := utility.Labels{
			"plugin": name,
		}
		utility.GlobalMetrics.Gauge(PluginMetricVMAvailable, float64(inst.Pool.AvailableCount()), labels)
	}
}
