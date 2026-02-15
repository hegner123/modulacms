package plugin

import (
	"testing"

	"github.com/hegner123/modulacms/internal/utility"
)

// -- RecordHTTPRequest --

func TestRecordHTTPRequest_IncrementsCounter(t *testing.T) {
	// Reset global metrics so previous test runs don't pollute.
	utility.GlobalMetrics.Reset()

	RecordHTTPRequest("test_plugin", "GET", 200, 42.0)

	snap := utility.GlobalMetrics.GetSnapshot()

	// The counter key is prefixed with the metric name and labels.
	// We check that at least one counter was incremented.
	found := false
	for key, m := range snap {
		if m.Type == utility.MetricTypeCounter && containsSubstring(key, PluginMetricHTTPRequests) {
			found = true
			if m.Value < 1 {
				t.Errorf("expected counter >= 1, got %f for key %q", m.Value, key)
			}
		}
	}
	if !found {
		t.Errorf("expected counter metric %q to be recorded", PluginMetricHTTPRequests)
	}
}

func TestRecordHTTPRequest_RecordsHistogram(t *testing.T) {
	utility.GlobalMetrics.Reset()

	RecordHTTPRequest("test_plugin", "POST", 201, 100.5)

	snap := utility.GlobalMetrics.GetSnapshot()

	found := false
	for key, m := range snap {
		if m.Type == utility.MetricTypeHistogram && containsSubstring(key, PluginMetricHTTPDuration) {
			found = true
			if m.Value < 100.0 {
				t.Errorf("expected histogram avg >= 100.0, got %f for key %q", m.Value, key)
			}
		}
	}
	if !found {
		t.Errorf("expected histogram metric %q to be recorded", PluginMetricHTTPDuration)
	}
}

// -- RecordHookExecution --

func TestRecordHookExecution_IncrementsCounterAndHistogram(t *testing.T) {
	utility.GlobalMetrics.Reset()

	RecordHookExecution(PluginMetricHookBefore, "test_plugin", "content.create", "posts", "success", 50.0)

	snap := utility.GlobalMetrics.GetSnapshot()

	foundCounter := false
	foundHistogram := false
	for key, m := range snap {
		if m.Type == utility.MetricTypeCounter && containsSubstring(key, PluginMetricHookBefore) {
			foundCounter = true
		}
		if m.Type == utility.MetricTypeHistogram && containsSubstring(key, PluginMetricHookDuration) {
			foundHistogram = true
		}
	}

	if !foundCounter {
		t.Errorf("expected counter metric %q", PluginMetricHookBefore)
	}
	if !foundHistogram {
		t.Errorf("expected histogram metric %q", PluginMetricHookDuration)
	}
}

// -- RecordReload --

func TestRecordReload_IncrementsCounter(t *testing.T) {
	utility.GlobalMetrics.Reset()

	RecordReload("test_plugin", "success")

	snap := utility.GlobalMetrics.GetSnapshot()

	found := false
	for key, m := range snap {
		if m.Type == utility.MetricTypeCounter && containsSubstring(key, PluginMetricReload) {
			found = true
			if m.Value < 1 {
				t.Errorf("expected counter >= 1, got %f", m.Value)
			}
		}
	}
	if !found {
		t.Errorf("expected counter metric %q", PluginMetricReload)
	}
}

// -- RecordCircuitBreakerTrip --

func TestRecordCircuitBreakerTrip_IncrementsCounter(t *testing.T) {
	utility.GlobalMetrics.Reset()

	RecordCircuitBreakerTrip("test_plugin")

	snap := utility.GlobalMetrics.GetSnapshot()

	found := false
	for key, m := range snap {
		if m.Type == utility.MetricTypeCounter && containsSubstring(key, PluginMetricCircuitBreakerTrip) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected counter metric %q", PluginMetricCircuitBreakerTrip)
	}
}

// -- RecordError --

func TestRecordError_IncrementsCounter(t *testing.T) {
	utility.GlobalMetrics.Reset()

	RecordError("test_plugin", "panic")

	snap := utility.GlobalMetrics.GetSnapshot()

	found := false
	for key, m := range snap {
		if m.Type == utility.MetricTypeCounter && containsSubstring(key, PluginMetricErrors) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected counter metric %q", PluginMetricErrors)
	}
}

// -- SnapshotVMAvailability --

func TestSnapshotVMAvailability_RecordsGauges(t *testing.T) {
	utility.GlobalMetrics.Reset()

	pool := newTestPool(3)
	defer pool.Close()

	plugins := map[string]*PluginInstance{
		"test_plugin": {
			Info: PluginInfo{Name: "test_plugin"},
			Pool: pool,
		},
	}

	SnapshotVMAvailability(plugins)

	snap := utility.GlobalMetrics.GetSnapshot()

	found := false
	for key, m := range snap {
		if m.Type == utility.MetricTypeGauge && containsSubstring(key, PluginMetricVMAvailable) {
			found = true
			if m.Value != 3 {
				t.Errorf("expected gauge = 3, got %f for key %q", m.Value, key)
			}
		}
	}
	if !found {
		t.Errorf("expected gauge metric %q", PluginMetricVMAvailable)
	}
}

func TestSnapshotVMAvailability_SkipsNilPool(t *testing.T) {
	utility.GlobalMetrics.Reset()

	plugins := map[string]*PluginInstance{
		"nil_pool": {
			Info: PluginInfo{Name: "nil_pool"},
			Pool: nil,
		},
	}

	// Should not panic.
	SnapshotVMAvailability(plugins)
}

// -- helper --

// containsSubstring checks if s contains substr. Used for metric key matching
// since keys include label suffixes.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
