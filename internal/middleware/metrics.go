package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/utility"
)

// HTTPMetricsMiddleware records HTTP request metrics to utility.GlobalMetrics.
//
// Metrics recorded:
//   - http.requests — counter, every request (labels: method, route, status)
//   - http.duration — histogram, request duration in ms (labels: method, route)
//   - http.errors   — counter, 4xx/5xx responses only (labels: method, route, status)
//
// Uses r.Pattern (Go 1.22+ ServeMux route pattern) for the "route" label to keep
// cardinality bounded. Falls back to "unknown" if no pattern is set.
func HTTPMetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			route := r.Pattern
			if route == "" {
				route = "unknown"
			}

			recordHTTPMetrics(r.Method, route, rw.statusCode, duration)
		})
	}
}

// recordHTTPMetrics records request counter, duration histogram, and error counter.
func recordHTTPMetrics(method, route string, status int, duration time.Duration) {
	statusStr := fmt.Sprintf("%d", status)

	utility.GlobalMetrics.Increment(utility.MetricHTTPRequests, utility.Labels{
		"method": method,
		"route":  route,
		"status": statusStr,
	})

	utility.GlobalMetrics.Timing(utility.MetricHTTPDuration, duration, utility.Labels{
		"method": method,
		"route":  route,
	})

	if status >= 400 {
		utility.GlobalMetrics.Increment(utility.MetricHTTPErrors, utility.Labels{
			"method": method,
			"route":  route,
			"status": statusStr,
		})
	}
}
