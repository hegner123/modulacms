package utility

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestStartRuntimeMetrics_RecordsGauges(t *testing.T) {
	// Swap GlobalMetrics to an isolated instance so we don't interfere with
	// other tests or the global singleton.
	origMetrics := GlobalMetrics
	GlobalMetrics = NewMetrics()
	t.Cleanup(func() {
		GlobalMetrics = origMetrics
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a very short interval so the test completes quickly.
	StartRuntimeMetrics(ctx, 10*time.Millisecond)

	// Wait long enough for at least one tick to fire.
	time.Sleep(50 * time.Millisecond)

	snap := GlobalMetrics.GetSnapshot()

	memMetric, ok := snap[MetricMemoryUsage]
	if !ok {
		t.Fatal("metric 'memory.usage' not recorded after StartRuntimeMetrics")
	}
	if memMetric.Type != MetricTypeGauge {
		t.Errorf("memory.usage type = %q, want %q", memMetric.Type, MetricTypeGauge)
	}
	if memMetric.Value <= 0 {
		t.Errorf("memory.usage value = %f, want > 0", memMetric.Value)
	}

	goroutineMetric, ok := snap[MetricGoroutines]
	if !ok {
		t.Fatal("metric 'goroutines.count' not recorded after StartRuntimeMetrics")
	}
	if goroutineMetric.Type != MetricTypeGauge {
		t.Errorf("goroutines.count type = %q, want %q", goroutineMetric.Type, MetricTypeGauge)
	}
	if goroutineMetric.Value < 1 {
		t.Errorf("goroutines.count value = %f, want >= 1", goroutineMetric.Value)
	}
}

func TestStartRuntimeMetrics_StopsOnContextCancel(t *testing.T) {
	origMetrics := GlobalMetrics
	GlobalMetrics = NewMetrics()
	t.Cleanup(func() {
		GlobalMetrics = origMetrics
	})

	ctx, cancel := context.WithCancel(context.Background())

	StartRuntimeMetrics(ctx, 10*time.Millisecond)

	// Let it tick a couple of times.
	time.Sleep(50 * time.Millisecond)

	// Cancel the context -- the goroutine should exit.
	cancel()

	// Give the goroutine time to observe the cancellation.
	time.Sleep(20 * time.Millisecond)

	// Record the goroutine count after cancellation.
	countAfterCancel := runtime.NumGoroutine()

	// Reset metrics and wait to confirm no further ticks arrive.
	GlobalMetrics.Reset()
	time.Sleep(50 * time.Millisecond)

	snap := GlobalMetrics.GetSnapshot()
	if _, ok := snap[MetricMemoryUsage]; ok {
		t.Error("runtime metrics goroutine still recording after context cancel")
	}

	// Sanity check: goroutine count should not be growing.
	countLater := runtime.NumGoroutine()
	if countLater > countAfterCancel+2 {
		t.Errorf("goroutine count growing after cancel: %d -> %d", countAfterCancel, countLater)
	}
}

func TestRecordRuntimeMetrics_DirectCall(t *testing.T) {
	origMetrics := GlobalMetrics
	GlobalMetrics = NewMetrics()
	t.Cleanup(func() {
		GlobalMetrics = origMetrics
	})

	recordRuntimeMetrics()

	snap := GlobalMetrics.GetSnapshot()

	if _, ok := snap[MetricMemoryUsage]; !ok {
		t.Fatal("recordRuntimeMetrics did not set memory.usage gauge")
	}
	if _, ok := snap[MetricGoroutines]; !ok {
		t.Fatal("recordRuntimeMetrics did not set goroutines.count gauge")
	}
}
