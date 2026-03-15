package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/utility"
)

// findMetric searches the snapshot for a metric with the given name prefix and matching labels.
func findMetric(snap map[string]utility.Metric, name string, wantLabels map[string]string) (utility.Metric, bool) {
	for key, m := range snap {
		if !strings.HasPrefix(key, name) {
			continue
		}
		if wantLabels == nil {
			return m, true
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

func TestHTTPMetricsMiddleware_RecordsRequestCounter(t *testing.T) {
	utility.GlobalMetrics.Reset()

	handler := HTTPMetricsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	snap := utility.GlobalMetrics.GetSnapshot()

	m, ok := findMetric(snap, utility.MetricHTTPRequests, map[string]string{
		"method": "GET",
		"status": "200",
	})
	if !ok {
		t.Fatal("expected http.requests metric with method=GET, status=200")
	}
	if m.Value != 1 {
		t.Errorf("expected counter value 1, got %f", m.Value)
	}
}

func TestHTTPMetricsMiddleware_RecordsDuration(t *testing.T) {
	utility.GlobalMetrics.Reset()

	handler := HTTPMetricsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	snap := utility.GlobalMetrics.GetSnapshot()

	m, ok := findMetric(snap, utility.MetricHTTPDuration, map[string]string{
		"method": "POST",
	})
	if !ok {
		t.Fatal("expected http.duration metric to be recorded")
	}
	if m.Type != "histogram" {
		t.Errorf("expected histogram type, got %q", m.Type)
	}
}

func TestHTTPMetricsMiddleware_RecordsErrors(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		wantError bool
	}{
		{"200 no error", http.StatusOK, false},
		{"201 no error", http.StatusCreated, false},
		{"301 no error", http.StatusMovedPermanently, false},
		{"400 is error", http.StatusBadRequest, true},
		{"404 is error", http.StatusNotFound, true},
		{"500 is error", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utility.GlobalMetrics.Reset()

			handler := HTTPMetricsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			snap := utility.GlobalMetrics.GetSnapshot()

			_, hasError := findMetric(snap, utility.MetricHTTPErrors, nil)
			if tt.wantError && !hasError {
				t.Error("expected http.errors metric to be recorded")
			}
			if !tt.wantError && hasError {
				t.Error("expected no http.errors metric")
			}
		})
	}
}

func TestHTTPMetricsMiddleware_DefaultStatusCode(t *testing.T) {
	utility.GlobalMetrics.Reset()

	// Handler that writes body without calling WriteHeader — Go defaults to 200.
	handler := HTTPMetricsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	snap := utility.GlobalMetrics.GetSnapshot()
	_, ok := findMetric(snap, utility.MetricHTTPRequests, map[string]string{"status": "200"})
	if !ok {
		t.Error("expected status=200 when WriteHeader not called")
	}
}

func TestHTTPMetricsMiddleware_RouteLabel(t *testing.T) {
	utility.GlobalMetrics.Reset()

	handler := HTTPMetricsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Without a real ServeMux, r.Pattern is empty → "unknown".
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	snap := utility.GlobalMetrics.GetSnapshot()
	_, ok := findMetric(snap, utility.MetricHTTPRequests, map[string]string{"route": "unknown"})
	if !ok {
		t.Error("expected route=unknown when r.Pattern is empty")
	}
}

func TestHTTPMetricsMiddleware_MultipleRequests(t *testing.T) {
	utility.GlobalMetrics.Reset()

	handler := HTTPMetricsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for range 5 {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	snap := utility.GlobalMetrics.GetSnapshot()
	m, ok := findMetric(snap, utility.MetricHTTPRequests, map[string]string{
		"method": "GET",
		"status": "200",
	})
	if !ok {
		t.Fatal("expected http.requests metric")
	}
	if m.Value != 5 {
		t.Errorf("expected counter value 5 after 5 requests, got %f", m.Value)
	}
}

func TestHTTPMetricsMiddleware_WithServeMux(t *testing.T) {
	utility.GlobalMetrics.Reset()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := HTTPMetricsMiddleware()(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/items/abc123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	snap := utility.GlobalMetrics.GetSnapshot()
	_, ok := findMetric(snap, utility.MetricHTTPRequests, map[string]string{
		"route": "GET /api/items/{id}",
	})
	if !ok {
		t.Error("expected route label to match ServeMux pattern 'GET /api/items/{id}'")
	}
}
