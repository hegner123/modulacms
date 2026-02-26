package utility

import (
	"sync"
	"testing"
	"time"
)

// ============================================================
// NewMetrics
// ============================================================

func TestNewMetrics(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics() returned nil")
	}
	// Fresh metrics should have empty snapshot
	snap := m.GetSnapshot()
	if len(snap) != 0 {
		t.Errorf("fresh Metrics has %d entries, want 0", len(snap))
	}
}

// ============================================================
// Counter
// ============================================================

func TestMetrics_Counter(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	m.Counter("requests", 1, nil)
	m.Counter("requests", 1, nil)
	m.Counter("requests", 3, nil)

	snap := m.GetSnapshot()
	metric, ok := snap["requests"]
	if !ok {
		t.Fatal("metric 'requests' not found in snapshot")
	}
	if metric.Type != MetricTypeCounter {
		t.Errorf("type = %q, want %q", metric.Type, MetricTypeCounter)
	}
	if metric.Value != 5 {
		t.Errorf("value = %f, want 5", metric.Value)
	}
}

func TestMetrics_Counter_WithLabels(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	labels := Labels{"method": "GET", "path": "/api"}
	m.Counter("http.requests", 1, labels)
	m.Counter("http.requests", 2, labels)

	snap := m.GetSnapshot()
	// The key includes labels, so we need to find it
	found := false
	for _, metric := range snap {
		if metric.Type == MetricTypeCounter && metric.Value == 3 {
			found = true
			if metric.Labels["method"] != "GET" {
				t.Errorf("label method = %q, want %q", metric.Labels["method"], "GET")
			}
			break
		}
	}
	if !found {
		t.Error("counter with labels not found in snapshot")
	}
}

// ============================================================
// Increment
// ============================================================

func TestMetrics_Increment(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	m.Increment("hits", nil)
	m.Increment("hits", nil)

	snap := m.GetSnapshot()
	metric, ok := snap["hits"]
	if !ok {
		t.Fatal("metric 'hits' not found")
	}
	if metric.Value != 2 {
		t.Errorf("Increment twice: value = %f, want 2", metric.Value)
	}
}

// ============================================================
// Gauge
// ============================================================

func TestMetrics_Gauge(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	// Gauge should replace, not accumulate
	m.Gauge("connections", 10, nil)
	m.Gauge("connections", 5, nil)
	m.Gauge("connections", 42, nil)

	snap := m.GetSnapshot()
	metric, ok := snap["connections"]
	if !ok {
		t.Fatal("metric 'connections' not found")
	}
	if metric.Type != MetricTypeGauge {
		t.Errorf("type = %q, want %q", metric.Type, MetricTypeGauge)
	}
	if metric.Value != 42 {
		t.Errorf("gauge value = %f, want 42 (last set value)", metric.Value)
	}
}

// ============================================================
// Histogram
// ============================================================

func TestMetrics_Histogram(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	m.Histogram("latency", 100, nil)
	m.Histogram("latency", 200, nil)
	m.Histogram("latency", 300, nil)

	snap := m.GetSnapshot()
	metric, ok := snap["latency"]
	if !ok {
		t.Fatal("metric 'latency' not found")
	}
	if metric.Type != MetricTypeHistogram {
		t.Errorf("type = %q, want %q", metric.Type, MetricTypeHistogram)
	}
	// Average of 100, 200, 300 = 200
	if metric.Value != 200 {
		t.Errorf("histogram average = %f, want 200", metric.Value)
	}
}

func TestMetrics_Histogram_SingleValue(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.Histogram("single", 42.5, nil)

	snap := m.GetSnapshot()
	if snap["single"].Value != 42.5 {
		t.Errorf("single histogram value = %f, want 42.5", snap["single"].Value)
	}
}

// ============================================================
// Timing
// ============================================================

func TestMetrics_Timing(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	m.Timing("request_time", 150*time.Millisecond, nil)

	snap := m.GetSnapshot()
	metric, ok := snap["request_time"]
	if !ok {
		t.Fatal("metric 'request_time' not found")
	}
	// 150ms = 150 milliseconds
	if metric.Value != 150 {
		t.Errorf("timing value = %f, want 150", metric.Value)
	}
}

// ============================================================
// Reset
// ============================================================

func TestMetrics_Reset(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	m.Counter("a", 5, nil)
	m.Gauge("b", 10, nil)
	m.Histogram("c", 20, nil)

	if len(m.GetSnapshot()) == 0 {
		t.Fatal("metrics should not be empty before reset")
	}

	m.Reset()
	snap := m.GetSnapshot()
	if len(snap) != 0 {
		t.Errorf("after Reset(), snapshot has %d entries, want 0", len(snap))
	}
}

// ============================================================
// GetSnapshot
// ============================================================

func TestMetrics_GetSnapshot_Isolation(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	m.Counter("x", 1, nil)
	snap := m.GetSnapshot()

	// Modifying the snapshot should not affect the metrics
	snap["x"] = Metric{Value: 999}

	snap2 := m.GetSnapshot()
	if snap2["x"].Value != 1 {
		t.Errorf("snapshot modification affected metrics: got %f, want 1", snap2["x"].Value)
	}
}

func TestMetrics_GetSnapshot_MixedTypes(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	m.Counter("counter_a", 10, nil)
	m.Gauge("gauge_b", 20, nil)
	m.Histogram("hist_c", 30, nil)

	snap := m.GetSnapshot()
	if len(snap) != 3 {
		t.Errorf("snapshot has %d entries, want 3", len(snap))
	}
	if snap["counter_a"].Type != MetricTypeCounter {
		t.Errorf("counter_a type = %q, want %q", snap["counter_a"].Type, MetricTypeCounter)
	}
	if snap["gauge_b"].Type != MetricTypeGauge {
		t.Errorf("gauge_b type = %q, want %q", snap["gauge_b"].Type, MetricTypeGauge)
	}
	if snap["hist_c"].Type != MetricTypeHistogram {
		t.Errorf("hist_c type = %q, want %q", snap["hist_c"].Type, MetricTypeHistogram)
	}
}

func TestMetrics_GetSnapshot_Timestamps(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.Counter("ts_test", 1, nil)

	before := time.Now()
	snap := m.GetSnapshot()
	after := time.Now()

	ts := snap["ts_test"].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("snapshot timestamp %v not between %v and %v", ts, before, after)
	}
}

// ============================================================
// Concurrent access (race detector test)
// ============================================================

func TestMetrics_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent counter increments
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range iterations {
			m.Counter("concurrent_counter", 1, nil)
		}
	}()

	// Concurrent gauge sets
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range iterations {
			m.Gauge("concurrent_gauge", float64(i), nil)
		}
	}()

	// Concurrent histogram records
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range iterations {
			m.Histogram("concurrent_hist", float64(i), nil)
		}
	}()

	// Concurrent snapshots
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range iterations {
			m.GetSnapshot()
		}
	}()

	// Concurrent resets (less frequent)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 10 {
			m.Reset()
		}
	}()

	wg.Wait()

	// If we get here without a race detector complaint, the locking is correct.
	// The final values are nondeterministic due to concurrent resets, so we
	// just verify it does not panic.
}

// ============================================================
// MeasureTime
// ============================================================

func TestMeasureTime(t *testing.T) {
	// Not parallel because it uses GlobalMetrics
	origMetrics := GlobalMetrics
	GlobalMetrics = NewMetrics()
	t.Cleanup(func() {
		GlobalMetrics = origMetrics
	})

	called := false
	MeasureTime("test_measure", nil, func() {
		called = true
		// Simulate some work
		time.Sleep(5 * time.Millisecond)
	})

	if !called {
		t.Fatal("MeasureTime did not call the function")
	}

	snap := GlobalMetrics.GetSnapshot()
	metric, ok := snap["test_measure"]
	if !ok {
		t.Fatal("metric 'test_measure' not found after MeasureTime")
	}
	if metric.Type != MetricTypeHistogram {
		t.Errorf("type = %q, want %q", metric.Type, MetricTypeHistogram)
	}
	// Should be at least 5ms
	if metric.Value < 5 {
		t.Errorf("measured time = %f ms, expected at least 5", metric.Value)
	}
}

// ============================================================
// MeasureTimeCtx
// ============================================================

func TestMeasureTimeCtx_Success(t *testing.T) {
	origMetrics := GlobalMetrics
	GlobalMetrics = NewMetrics()
	t.Cleanup(func() {
		GlobalMetrics = origMetrics
	})

	err := MeasureTimeCtx("ctx_measure", nil, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := GlobalMetrics.GetSnapshot()
	if _, ok := snap["ctx_measure"]; !ok {
		t.Fatal("metric 'ctx_measure' not found")
	}
}

func TestMeasureTimeCtx_Error(t *testing.T) {
	origMetrics := GlobalMetrics
	GlobalMetrics = NewMetrics()
	t.Cleanup(func() {
		GlobalMetrics = origMetrics
	})

	wantErr := "something failed"
	err := MeasureTimeCtx("ctx_err", nil, func() error {
		return &testError{msg: wantErr}
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != wantErr {
		t.Errorf("error = %q, want %q", err.Error(), wantErr)
	}

	// Metric should still be recorded even on error
	snap := GlobalMetrics.GetSnapshot()
	if _, ok := snap["ctx_err"]; !ok {
		t.Fatal("metric 'ctx_err' not recorded on error path")
	}
}

// ============================================================
// MetricType constants
// ============================================================

func TestMetricType_Values(t *testing.T) {
	t.Parallel()
	if MetricTypeCounter != "counter" {
		t.Errorf("MetricTypeCounter = %q, want %q", MetricTypeCounter, "counter")
	}
	if MetricTypeGauge != "gauge" {
		t.Errorf("MetricTypeGauge = %q, want %q", MetricTypeGauge, "gauge")
	}
	if MetricTypeHistogram != "histogram" {
		t.Errorf("MetricTypeHistogram = %q, want %q", MetricTypeHistogram, "histogram")
	}
}

// ============================================================
// keyWithLabels (tested indirectly, but verify isolation)
// ============================================================

func TestMetrics_SameNameDifferentLabels(t *testing.T) {
	t.Parallel()
	m := NewMetrics()

	m.Counter("requests", 1, Labels{"method": "GET"})
	m.Counter("requests", 1, Labels{"method": "POST"})
	m.Counter("requests", 1, nil) // no labels

	snap := m.GetSnapshot()
	// Should have 3 distinct entries
	if len(snap) != 3 {
		t.Errorf("expected 3 distinct metric entries, got %d", len(snap))
		for k, v := range snap {
			t.Logf("  %q: %+v", k, v)
		}
	}
}

// testError is a minimal error implementation for tests
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
