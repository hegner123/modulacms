package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/utility"
)

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	handler := RecoveryMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRecoveryMiddleware_CatchesPanic(t *testing.T) {
	handler := RecoveryMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Should not panic — recovery middleware catches it.
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestRecoveryMiddleware_PanicWithNonStringValue(t *testing.T) {
	// Panics can carry any type, not just strings.
	handler := RecoveryMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(42)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestRecoveryMiddleware_PanicWithErrorValue(t *testing.T) {
	handler := RecoveryMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("wrapped error"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestRecoveryMiddleware_WithObsClient_DelegatesCaptureRequestError(t *testing.T) {
	rp := &testRecordingProvider{}
	client := utility.NewObservabilityClientFromProvider(rp)

	handler := RecoveryMiddleware(client)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("obs panic")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	rp.mu.Lock()
	defer rp.mu.Unlock()
	if len(rp.requestErrors) == 0 {
		t.Fatal("expected CaptureRequestError to be called, got 0 calls")
	}
	if rp.requestErrors[0].method != "POST" {
		t.Errorf("captured method = %q, want POST", rp.requestErrors[0].method)
	}
	if rp.requestErrors[0].path != "/api/data" {
		t.Errorf("captured path = %q, want /api/data", rp.requestErrors[0].path)
	}
}

// testRecordingProvider implements ObservabilityProvider for middleware tests.
type testRecordingProvider struct {
	mu            sync.Mutex
	requestErrors []capturedRequest
}

type capturedRequest struct {
	method string
	path   string
	err    error
}

func (p *testRecordingProvider) SendMetric(utility.Metric) error                         { return nil }
func (p *testRecordingProvider) SendError(error, map[string]any) error                   { return nil }
func (p *testRecordingProvider) Flush(time.Duration) error                               { return nil }
func (p *testRecordingProvider) Close() error                                            { return nil }
func (p *testRecordingProvider) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}
func (p *testRecordingProvider) CaptureRequestError(err error, r *http.Request, _ map[string]any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requestErrors = append(p.requestErrors, capturedRequest{
		method: r.Method,
		path:   r.URL.Path,
		err:    err,
	})
}
