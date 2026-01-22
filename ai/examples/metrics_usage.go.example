package utility

import (
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	// http.ResponseWriter
	statusCode int
	written    bool
}

// Example: How to instrument HTTP handlers with metrics
// This is example code showing the pattern - integrate directly into middleware package
/*
func ExampleHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Track active connections
		GlobalMetrics.Increment(MetricActiveConnections, Labels{
			"type": "http",
		})
		defer GlobalMetrics.Counter(MetricActiveConnections, -1, Labels{
			"type": "http",
		})

		// Track request count
		GlobalMetrics.Increment(MetricHTTPRequests, Labels{
			"method": r.Method,
			"path":   r.URL.Path,
		})

		// Create response wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Execute handler
		next.ServeHTTP(rw, r)

		// Record duration
		duration := time.Since(start)
		GlobalMetrics.Timing(MetricHTTPDuration, duration, Labels{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": http.StatusText(rw.statusCode),
		})

		// Track errors (4xx and 5xx)
		if rw.statusCode >= 400 {
			GlobalMetrics.Increment(MetricHTTPErrors, Labels{
				"method": r.Method,
				"path":   r.URL.Path,
				"status": http.StatusText(rw.statusCode),
			})
		}
	})
}
*/

// Example: How to track database queries
func ExampleDatabaseQuery(queryName string, fn func() error) error {
	start := time.Now()

	// Increment query counter
	GlobalMetrics.Increment(MetricDBQueries, Labels{
		"query": queryName,
	})

	// Execute query
	err := fn()

	// Record duration
	duration := time.Since(start)
	GlobalMetrics.Timing(MetricDBDuration, duration, Labels{
		"query": queryName,
	})

	// Track errors
	if err != nil {
		GlobalMetrics.Increment(MetricDBErrors, Labels{
			"query": queryName,
			"error": err.Error(),
		})

		// Send error to observability platform
		CaptureError(err, map[string]any{
			"query":    queryName,
			"duration": duration.Milliseconds(),
		})
	}

	return err
}

// Example: How to track SSH connections
func ExampleSSHConnection(user string, fn func() error) error {
	GlobalMetrics.Increment(MetricSSHConnections, Labels{
		"user": user,
	})

	err := fn()

	if err != nil {
		GlobalMetrics.Increment(MetricSSHErrors, Labels{
			"user":  user,
			"error": err.Error(),
		})

		CaptureError(err, map[string]any{
			"user":      user,
			"subsystem": "ssh",
		})
	}

	return err
}

// Example: How to track cache performance
func ExampleCacheGet(key string) (interface{}, bool) {
	// Simulated cache lookup
	value, found := map[string]interface{}{}[key]

	if found {
		GlobalMetrics.Increment(MetricCacheHits, Labels{
			"cache": "primary",
		})
	} else {
		GlobalMetrics.Increment(MetricCacheMisses, Labels{
			"cache": "primary",
		})
	}

	return value, found
}

// Example: How to track system metrics (memory, goroutines)
func ExampleCollectSystemMetrics() {
	// In production, you'd use runtime.ReadMemStats() and runtime.NumGoroutine()

	// Example memory usage in bytes
	memoryBytes := float64(1024 * 1024 * 512) // 512 MB
	GlobalMetrics.Gauge(MetricMemoryUsage, memoryBytes, Labels{
		"type": "heap",
	})

	// Example goroutine count
	goroutineCount := float64(100)
	GlobalMetrics.Gauge(MetricGoroutines, goroutineCount, nil)
}

// Example: Using MeasureTime helper
func ExampleBusinessLogic() {
	MeasureTime("business.process_order", Labels{"type": "ecommerce"}, func() {
		// Your business logic here
		time.Sleep(50 * time.Millisecond)
	})
}

// Example: Using MeasureTimeCtx with error handling
func ExampleBusinessLogicWithError() error {
	return MeasureTimeCtx("business.create_user", Labels{"source": "api"}, func() error {
		// Your business logic here
		time.Sleep(20 * time.Millisecond)
		return nil
	})
}
