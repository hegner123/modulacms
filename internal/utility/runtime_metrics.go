package utility

import (
	"context"
	"runtime"
	"time"
)

// StartRuntimeMetrics starts a background goroutine that periodically samples
// Go runtime statistics and records them as gauge metrics via GlobalMetrics.
//
// Recorded metrics:
//   - memory.usage     -- current heap allocation in bytes (runtime.MemStats.Alloc)
//   - goroutines.count -- current number of goroutines (runtime.NumGoroutine)
//
// The goroutine exits when ctx is cancelled. A typical interval is 15 seconds.
func StartRuntimeMetrics(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				recordRuntimeMetrics()
			}
		}
	}()
}

// recordRuntimeMetrics samples runtime stats and writes them to GlobalMetrics.
func recordRuntimeMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	GlobalMetrics.Gauge(MetricMemoryUsage, float64(memStats.Alloc), nil)
	GlobalMetrics.Gauge(MetricGoroutines, float64(runtime.NumGoroutine()), nil)
}
