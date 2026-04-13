package utility

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

// ============================================================
// ConsoleProvider
// ============================================================

func TestConsoleProvider_Implements_ObservabilityProvider(t *testing.T) {
	t.Parallel()
	// Compile-time check that ConsoleProvider implements ObservabilityProvider
	var _ ObservabilityProvider = (*ConsoleProvider)(nil)
}

func TestConsoleProvider_SendMetric(t *testing.T) {
	t.Parallel()
	p := NewConsoleProvider()
	err := p.SendMetric(Metric{Name: "test", Type: MetricTypeCounter, Value: 1})
	if err != nil {
		t.Fatalf("SendMetric returned error: %v", err)
	}
}

func TestConsoleProvider_SendError(t *testing.T) {
	t.Parallel()
	p := NewConsoleProvider()
	err := p.SendError(errors.New("test error"), map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("SendError returned error: %v", err)
	}
}

func TestConsoleProvider_Flush(t *testing.T) {
	t.Parallel()
	p := NewConsoleProvider()
	err := p.Flush(5 * time.Second)
	if err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}
}

func TestConsoleProvider_Close(t *testing.T) {
	t.Parallel()
	p := NewConsoleProvider()
	err := p.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

// ============================================================
// SentryProvider
// ============================================================

func TestSentryProvider_Implements_ObservabilityProvider(t *testing.T) {
	t.Parallel()
	var _ ObservabilityProvider = (*SentryProvider)(nil)
}

func TestNewSentryProvider(t *testing.T) {
	t.Parallel()
	p, err := NewSentryProvider(ObservabilityConfig{
		DSN:         "https://key@sentry.example.com/1",
		Environment: "test",
	})
	if err != nil {
		t.Fatalf("NewSentryProvider returned error: %v", err)
	}
	if p == nil {
		t.Fatal("NewSentryProvider returned nil")
	}
}

func TestSentryProvider_AllMethods(t *testing.T) {
	t.Parallel()
	p, err := NewSentryProvider(ObservabilityConfig{
		DSN: "https://key@sentry.example.com/1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All methods should succeed without error (placeholder implementation)
	if err := p.SendMetric(Metric{Name: "test"}); err != nil {
		t.Errorf("SendMetric error: %v", err)
	}
	if err := p.SendError(errors.New("test"), nil); err != nil {
		t.Errorf("SendError error: %v", err)
	}
	if err := p.Flush(time.Second); err != nil {
		t.Errorf("Flush error: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}

// ============================================================
// NewObservabilityClient
// ============================================================

func TestNewObservabilityClient_Disabled(t *testing.T) {
	t.Parallel()
	client, err := NewObservabilityClient(ObservabilityConfig{
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client even when disabled")
	}
	if client.enabled {
		t.Error("client should be disabled")
	}
}

func TestNewObservabilityClient_ConsoleProvider(t *testing.T) {
	t.Parallel()
	client, err := NewObservabilityClient(ObservabilityConfig{
		Enabled:       true,
		Provider:      "console",
		FlushInterval: "10s",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if !client.enabled {
		t.Error("client should be enabled")
	}
	if client.flushInterval != 10*time.Second {
		t.Errorf("flush interval = %v, want 10s", client.flushInterval)
	}
}

func TestNewObservabilityClient_SentryProvider(t *testing.T) {
	t.Parallel()
	client, err := NewObservabilityClient(ObservabilityConfig{
		Enabled:       true,
		Provider:      "sentry",
		DSN:           "https://key@sentry.example.com/1",
		FlushInterval: "5s",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewObservabilityClient_UnsupportedProvider(t *testing.T) {
	t.Parallel()
	_, err := NewObservabilityClient(ObservabilityConfig{
		Enabled:  true,
		Provider: "datadog",
	})
	if err == nil {
		t.Fatal("expected error for unsupported provider, got nil")
	}
}

func TestNewObservabilityClient_InvalidFlushInterval_DefaultsTo30s(t *testing.T) {
	t.Parallel()
	// An invalid FlushInterval should default to 30s without preventing
	// provider creation. Previously, the ParseDuration error leaked into the
	// provider error check (fixed by separating the error variables).
	client, err := NewObservabilityClient(ObservabilityConfig{
		Enabled:       true,
		Provider:      "console",
		FlushInterval: "not-a-duration",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.flushInterval != 30*time.Second {
		t.Errorf("flush interval = %v, want 30s", client.flushInterval)
	}
}

// ============================================================
// ObservabilityClient lifecycle: Start + Stop
// ============================================================

func TestObservabilityClient_StartStop_Enabled(t *testing.T) {
	t.Parallel()

	client, err := NewObservabilityClient(ObservabilityConfig{
		Enabled:       true,
		Provider:      "console",
		FlushInterval: "50ms",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.Start(ctx)

	// Let at least one flush tick happen
	time.Sleep(100 * time.Millisecond)

	if err := client.Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
}

func TestObservabilityClient_StartStop_Disabled(t *testing.T) {
	t.Parallel()

	client, err := NewObservabilityClient(ObservabilityConfig{
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	client.Start(ctx) // Should be a no-op

	if err := client.Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
}

func TestObservabilityClient_Start_ContextCancellation(t *testing.T) {
	t.Parallel()

	client, err := NewObservabilityClient(ObservabilityConfig{
		Enabled:       true,
		Provider:      "console",
		FlushInterval: "1h", // Long interval -- cancellation should stop it
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	client.Start(ctx)

	// Cancel should cause the goroutine to exit
	cancel()

	// Give the goroutine time to react to cancellation
	time.Sleep(50 * time.Millisecond)

	// Stop should complete quickly since the goroutine already exited
	done := make(chan error, 1)
	go func() {
		done <- client.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Stop returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop timed out -- goroutine likely did not exit on context cancellation")
	}
}

// ============================================================
// ObservabilityClient.SendError
// ============================================================

func TestObservabilityClient_SendError_Disabled(t *testing.T) {
	t.Parallel()
	client, err := NewObservabilityClient(ObservabilityConfig{Enabled: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = client.SendError(errors.New("test"), nil)
	if err != nil {
		t.Fatalf("SendError on disabled client returned error: %v", err)
	}
}

func TestObservabilityClient_SendError_Enabled(t *testing.T) {
	t.Parallel()
	client, err := NewObservabilityClient(ObservabilityConfig{
		Enabled:       true,
		Provider:      "console",
		FlushInterval: "10s",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = client.SendError(errors.New("test error"), map[string]any{"user_id": "123"})
	if err != nil {
		t.Fatalf("SendError returned error: %v", err)
	}
}

// ============================================================
// CaptureError
// ============================================================

func TestCaptureError_WithNilGlobal(t *testing.T) {
	// Not parallel because we modify GlobalObservability
	orig := GlobalObservability
	t.Cleanup(func() {
		GlobalObservability = orig
	})

	GlobalObservability = nil
	// Should not panic when GlobalObservability is nil
	CaptureError(errors.New("test"), nil)
}

func TestCaptureError_WithClient(t *testing.T) {
	orig := GlobalObservability
	t.Cleanup(func() {
		GlobalObservability = orig
	})

	client, err := NewObservabilityClient(ObservabilityConfig{
		Enabled:       true,
		Provider:      "console",
		FlushInterval: "10s",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	GlobalObservability = client

	// Should not panic and should call provider
	CaptureError(errors.New("captured error"), map[string]any{"module": "test"})
}

// ============================================================
// flush (tested indirectly through recordingProvider)
// ============================================================

func TestObservabilityClient_Flush_SendsMetrics(t *testing.T) {
	t.Parallel()

	rp := &recordingProvider{}
	m := NewMetrics()
	m.Counter("flush_test", 1, nil)

	client := &ObservabilityClient{
		provider:      rp,
		metrics:       m,
		flushInterval: time.Hour,
		stopChan:      make(chan struct{}),
		enabled:       true,
	}

	client.flush()

	rp.mu.Lock()
	defer rp.mu.Unlock()
	if len(rp.metrics) != 1 {
		t.Errorf("expected 1 metric sent, got %d", len(rp.metrics))
	}
}

// recordingProvider captures metrics for verification
type recordingProvider struct {
	mu      sync.Mutex
	metrics []Metric
	errors  []error
}

func (r *recordingProvider) SendMetric(metric Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = append(r.metrics, metric)
	return nil
}

func (r *recordingProvider) SendError(err error, context map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, err)
	return nil
}

func (r *recordingProvider) Flush(timeout time.Duration) error { return nil }
func (r *recordingProvider) Close() error                      { return nil }
func (r *recordingProvider) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}
func (r *recordingProvider) CaptureRequestError(err error, req *http.Request, context map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, err)
}

// ============================================================
// initSentryWithMockTransport sets up a Sentry client with a MockTransport
// and returns the transport for event inspection. Not parallel-safe because
// sentry.Init replaces the global client.
// ============================================================

func initSentryWithMockTransport(t *testing.T) *sentry.MockTransport {
	t.Helper()
	transport := &sentry.MockTransport{}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://key@sentry.example.com/1",
		Transport:        transport,
		AttachStacktrace: true,
	})
	if err != nil {
		t.Fatalf("sentry.Init with MockTransport: %v", err)
	}
	return transport
}

// ============================================================
// SentryProvider: MockTransport event verification
// ============================================================

func TestSentryProvider_SendError_CapturesEvent(t *testing.T) {
	transport := initSentryWithMockTransport(t)

	p := &SentryProvider{config: ObservabilityConfig{DSN: "https://key@sentry.example.com/1"}}
	testErr := errors.New("database connection failed")
	err := p.SendError(testErr, map[string]any{"table": "users", "op": "insert"})
	if err != nil {
		t.Fatalf("SendError returned error: %v", err)
	}
	sentry.Flush(2 * time.Second)

	events := transport.Events()
	if len(events) == 0 {
		t.Fatal("expected at least 1 Sentry event, got 0")
	}

	event := events[len(events)-1]
	if len(event.Exception) == 0 {
		t.Fatal("event has no exceptions")
	}
	if event.Exception[0].Value != "database connection failed" {
		t.Errorf("exception value = %q, want %q", event.Exception[0].Value, "database connection failed")
	}

	tableVal, ok := event.Extra["table"]
	if !ok {
		t.Error("expected 'table' in event extras")
	} else if tableVal != "users" {
		t.Errorf("extra 'table' = %v, want 'users'", tableVal)
	}
}

func TestSentryProvider_SendMetric_AddsBreadcrumb(t *testing.T) {
	transport := initSentryWithMockTransport(t)

	p := &SentryProvider{config: ObservabilityConfig{DSN: "https://key@sentry.example.com/1"}}
	err := p.SendMetric(Metric{Name: "http_requests", Type: MetricTypeCounter, Value: 42})
	if err != nil {
		t.Fatalf("SendMetric returned error: %v", err)
	}

	// Breadcrumbs attach to the next error event, so send one.
	p.SendError(errors.New("trigger"), nil)
	sentry.Flush(2 * time.Second)

	events := transport.Events()
	if len(events) == 0 {
		t.Fatal("expected at least 1 event, got 0")
	}

	event := events[len(events)-1]
	found := false
	for _, bc := range event.Breadcrumbs {
		if bc.Category == "metric" && bc.Data["name"] == "http_requests" {
			found = true
			break
		}
	}
	if !found {
		t.Error("metric breadcrumb not found on event")
	}
}

func TestSentryProvider_CaptureRequestError_UsesRequestHub(t *testing.T) {
	transport := initSentryWithMockTransport(t)

	p := &SentryProvider{config: ObservabilityConfig{DSN: "https://key@sentry.example.com/1"}}

	// Simulate a request with a hub on context (as HTTPMiddleware would set).
	hub := sentry.CurrentHub().Clone()
	hub.Scope().SetTag("request_id", "req-abc")
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	req := httptest.NewRequest(http.MethodPost, "/api/content", nil).WithContext(ctx)

	p.CaptureRequestError(errors.New("handler panic"), req, map[string]any{"stack": "..."})
	sentry.Flush(2 * time.Second)

	events := transport.Events()
	if len(events) == 0 {
		t.Fatal("expected at least 1 event, got 0")
	}

	event := events[len(events)-1]
	if event.Tags["request_id"] != "req-abc" {
		t.Errorf("expected tag request_id=req-abc, got %q", event.Tags["request_id"])
	}
}

// ============================================================
// HTTPMiddleware: SentryProvider
// ============================================================

func TestSentryProvider_HTTPMiddleware_CreatesTransaction(t *testing.T) {
	transport := initSentryWithMockTransport(t)

	// Re-init with tracing enabled so transactions are sampled.
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://key@sentry.example.com/1",
		Transport:        transport,
		TracesSampleRate: 1.0,
		EnableTracing:    true,
	})
	if err != nil {
		t.Fatalf("sentry.Init: %v", err)
	}

	p := &SentryProvider{config: ObservabilityConfig{}}
	handler := p.HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify hub is on context inside the handler.
		hub := sentry.GetHubFromContext(r.Context())
		if hub == nil {
			t.Error("expected Sentry hub on request context")
		}
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/content", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	sentry.Flush(2 * time.Second)
	events := transport.Events()

	// Find the transaction event.
	var txn *sentry.Event
	for _, e := range events {
		if e.Type == "transaction" {
			txn = e
			break
		}
	}
	if txn == nil {
		t.Fatal("expected a transaction event, found none")
	}
	if txn.Tags["http.method"] != "POST" {
		t.Errorf("transaction tag http.method = %q, want POST", txn.Tags["http.method"])
	}
	if txn.Tags["http.path"] != "/api/content" {
		t.Errorf("transaction tag http.path = %q, want /api/content", txn.Tags["http.path"])
	}
	if txn.Tags["http.status_code"] != "201" {
		t.Errorf("transaction tag http.status_code = %q, want 201", txn.Tags["http.status_code"])
	}
}

func TestSentryProvider_HTTPMiddleware_SetsErrorStatus(t *testing.T) {
	transport := initSentryWithMockTransport(t)

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://key@sentry.example.com/1",
		Transport:        transport,
		TracesSampleRate: 1.0,
		EnableTracing:    true,
	})
	if err != nil {
		t.Fatalf("sentry.Init: %v", err)
	}

	p := &SentryProvider{config: ObservabilityConfig{}}
	handler := p.HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	sentry.Flush(2 * time.Second)
	events := transport.Events()

	var txn *sentry.Event
	for _, e := range events {
		if e.Type == "transaction" {
			txn = e
			break
		}
	}
	if txn == nil {
		t.Fatal("expected a transaction event, found none")
	}
	if txn.Tags["http.status_code"] != "500" {
		t.Errorf("transaction tag http.status_code = %q, want 500", txn.Tags["http.status_code"])
	}
}

// ============================================================
// HTTPMiddleware: ConsoleProvider
// ============================================================

func TestConsoleProvider_HTTPMiddleware_PassThrough(t *testing.T) {
	t.Parallel()
	p := NewConsoleProvider()
	handler := p.HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rec.Body.String())
	}
}

func TestConsoleProvider_CaptureRequestError(t *testing.T) {
	t.Parallel()
	p := NewConsoleProvider()
	req := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	// Should not panic.
	p.CaptureRequestError(errors.New("test"), req, map[string]any{"key": "val"})
}

// ============================================================
// ObservabilityClient: HTTPMiddleware delegation
// ============================================================

func TestObservabilityClient_HTTPMiddleware_Disabled(t *testing.T) {
	t.Parallel()
	client := &ObservabilityClient{enabled: false}
	mw := client.HTTPMiddleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestObservabilityClient_HTTPMiddleware_DelegatesToProvider(t *testing.T) {
	t.Parallel()
	called := false
	provider := &recordingProvider{}
	client := &ObservabilityClient{
		enabled:  true,
		provider: provider,
	}

	// The recordingProvider's HTTPMiddleware is a pass-through.
	mw := client.HTTPMiddleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("inner handler was not called")
	}
}

func TestObservabilityClient_CaptureRequestError_Disabled(t *testing.T) {
	t.Parallel()
	client := &ObservabilityClient{enabled: false}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Should not panic.
	client.CaptureRequestError(errors.New("test"), req, nil)
}

func TestObservabilityClient_CaptureRequestError_DelegatesToProvider(t *testing.T) {
	t.Parallel()
	rp := &recordingProvider{}
	client := &ObservabilityClient{
		enabled:  true,
		provider: rp,
	}
	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	client.CaptureRequestError(errors.New("provider error"), req, map[string]any{"key": "val"})

	rp.mu.Lock()
	defer rp.mu.Unlock()
	if len(rp.errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(rp.errors))
	}
	if rp.errors[0].Error() != "provider error" {
		t.Errorf("error = %q, want 'provider error'", rp.errors[0].Error())
	}
}

// ============================================================
// Edge cases and assumption violations
// ============================================================

func TestSentryProvider_SendError_NilContextMap(t *testing.T) {
	transport := initSentryWithMockTransport(t)

	p := &SentryProvider{config: ObservabilityConfig{DSN: "https://key@sentry.example.com/1"}}
	err := p.SendError(errors.New("nil ctx"), nil)
	if err != nil {
		t.Fatalf("SendError returned error: %v", err)
	}
	sentry.Flush(2 * time.Second)

	events := transport.Events()
	if len(events) == 0 {
		t.Fatal("expected at least 1 event, got 0")
	}
	if events[len(events)-1].Exception[0].Value != "nil ctx" {
		t.Errorf("exception value = %q, want 'nil ctx'", events[len(events)-1].Exception[0].Value)
	}
}

func TestSentryProvider_CaptureRequestError_NoHubOnContext(t *testing.T) {
	// When no hub is on the request context, CaptureRequestError should
	// fall back to CurrentHub without panicking.
	transport := initSentryWithMockTransport(t)

	p := &SentryProvider{config: ObservabilityConfig{DSN: "https://key@sentry.example.com/1"}}
	req := httptest.NewRequest(http.MethodGet, "/no-hub", nil)

	p.CaptureRequestError(errors.New("no hub error"), req, nil)
	sentry.Flush(2 * time.Second)

	events := transport.Events()
	if len(events) == 0 {
		t.Fatal("expected at least 1 event, got 0")
	}
	if events[len(events)-1].Exception[0].Value != "no hub error" {
		t.Errorf("exception = %q, want 'no hub error'", events[len(events)-1].Exception[0].Value)
	}
}

func TestSentryProvider_CaptureRequestError_NilContextMap(t *testing.T) {
	transport := initSentryWithMockTransport(t)

	p := &SentryProvider{config: ObservabilityConfig{DSN: "https://key@sentry.example.com/1"}}
	req := httptest.NewRequest(http.MethodGet, "/nil-ctx", nil)

	p.CaptureRequestError(errors.New("nil ctx req"), req, nil)
	sentry.Flush(2 * time.Second)

	events := transport.Events()
	if len(events) == 0 {
		t.Fatal("expected at least 1 event, got 0")
	}
}

func TestSentryProvider_HTTPMiddleware_DefaultStatusWhenWriteHeaderNotCalled(t *testing.T) {
	transport := initSentryWithMockTransport(t)

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://key@sentry.example.com/1",
		Transport:        transport,
		TracesSampleRate: 1.0,
		EnableTracing:    true,
	})
	if err != nil {
		t.Fatalf("sentry.Init: %v", err)
	}

	p := &SentryProvider{config: ObservabilityConfig{}}
	handler := p.HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write body without explicit WriteHeader. Go implicitly sends 200.
		w.Write([]byte("implicit 200"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/implicit", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	sentry.Flush(2 * time.Second)
	events := transport.Events()

	var txn *sentry.Event
	for _, e := range events {
		if e.Type == "transaction" {
			txn = e
			break
		}
	}
	if txn == nil {
		t.Fatal("expected a transaction event")
	}
	// Default status should be 200 when WriteHeader is never called.
	if txn.Tags["http.status_code"] != "200" {
		t.Errorf("http.status_code = %q, want '200'", txn.Tags["http.status_code"])
	}
}

func TestSentryProvider_HTTPMiddleware_QueryStringCaptured(t *testing.T) {
	transport := initSentryWithMockTransport(t)

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://key@sentry.example.com/1",
		Transport:        transport,
		TracesSampleRate: 1.0,
		EnableTracing:    true,
	})
	if err != nil {
		t.Fatalf("sentry.Init: %v", err)
	}

	p := &SentryProvider{config: ObservabilityConfig{}}
	handler := p.HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&page=2", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	sentry.Flush(2 * time.Second)
	events := transport.Events()

	var txn *sentry.Event
	for _, e := range events {
		if e.Type == "transaction" {
			txn = e
			break
		}
	}
	if txn == nil {
		t.Fatal("expected a transaction event")
	}
	// Transaction name should use path only, not query string.
	if txn.Transaction != "GET /search" {
		t.Errorf("transaction name = %q, want 'GET /search'", txn.Transaction)
	}
}

func TestSentryProvider_HTTPMiddleware_StatusCodeMapping(t *testing.T) {
	// Verify different HTTP status codes produce the correct Sentry span status.
	tests := []struct {
		name       string
		statusCode int
		wantTag    string
	}{
		{"200 OK", http.StatusOK, "200"},
		{"201 Created", http.StatusCreated, "201"},
		{"204 No Content", http.StatusNoContent, "204"},
		{"400 Bad Request", http.StatusBadRequest, "400"},
		{"401 Unauthorized", http.StatusUnauthorized, "401"},
		{"403 Forbidden", http.StatusForbidden, "403"},
		{"404 Not Found", http.StatusNotFound, "404"},
		{"429 Too Many Requests", http.StatusTooManyRequests, "429"},
		{"500 Internal Server Error", http.StatusInternalServerError, "500"},
		{"502 Bad Gateway", http.StatusBadGateway, "502"},
		{"503 Service Unavailable", http.StatusServiceUnavailable, "503"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transport := &sentry.MockTransport{}
			err := sentry.Init(sentry.ClientOptions{
				Dsn:              "https://key@sentry.example.com/1",
				Transport:        transport,
				TracesSampleRate: 1.0,
				EnableTracing:    true,
			})
			if err != nil {
				t.Fatalf("sentry.Init: %v", err)
			}

			p := &SentryProvider{config: ObservabilityConfig{}}
			handler := p.HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			sentry.Flush(2 * time.Second)
			events := transport.Events()

			var txn *sentry.Event
			for _, e := range events {
				if e.Type == "transaction" {
					txn = e
					break
				}
			}
			if txn == nil {
				t.Fatal("expected a transaction event")
			}
			if txn.Tags["http.status_code"] != tc.wantTag {
				t.Errorf("http.status_code = %q, want %q", txn.Tags["http.status_code"], tc.wantTag)
			}
		})
	}
}

func TestNewSentryProvider_EmptyDSN(t *testing.T) {
	// sentry.Init with an empty DSN succeeds but creates a no-op client.
	// The provider should still be usable.
	p, err := NewSentryProvider(ObservabilityConfig{DSN: ""})
	if err != nil {
		t.Fatalf("NewSentryProvider with empty DSN returned error: %v", err)
	}
	// SendError should not panic on a no-op client.
	err = p.SendError(errors.New("test"), nil)
	if err != nil {
		t.Errorf("SendError on no-op client returned error: %v", err)
	}
}

func TestSentryProvider_TagsApplied(t *testing.T) {
	// NewSentryProvider calls sentry.Init internally, which would overwrite a
	// MockTransport. Instead, test the tag application path directly: init
	// with MockTransport, apply tags via ConfigureScope (same as
	// NewSentryProvider does), and verify they appear on events.
	transport := initSentryWithMockTransport(t)

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("region", "us-east")
		scope.SetTag("service", "cms")
	})

	p := &SentryProvider{config: ObservabilityConfig{}}
	p.SendError(errors.New("tag test"), nil)
	sentry.Flush(2 * time.Second)

	events := transport.Events()
	if len(events) == 0 {
		t.Fatal("expected at least 1 event")
	}

	event := events[len(events)-1]
	if event.Tags["region"] != "us-east" {
		t.Errorf("tag 'region' = %q, want 'us-east'", event.Tags["region"])
	}
	if event.Tags["service"] != "cms" {
		t.Errorf("tag 'service' = %q, want 'cms'", event.Tags["service"])
	}
}

func TestObservabilityClient_HTTPMiddleware_NilProvider(t *testing.T) {
	t.Parallel()
	// enabled=true but nil provider should return pass-through, not panic.
	client := &ObservabilityClient{enabled: true, provider: nil}
	mw := client.HTTPMiddleware()

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestObservabilityClient_CaptureRequestError_NilProvider(t *testing.T) {
	t.Parallel()
	// enabled=true but nil provider should not panic.
	client := &ObservabilityClient{enabled: true, provider: nil}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	client.CaptureRequestError(errors.New("test"), req, nil)
}

func TestSentryProvider_FlushTimeout(t *testing.T) {
	// With MockTransport, Flush always returns true. Verify the success path.
	transport := initSentryWithMockTransport(t)
	_ = transport

	p := &SentryProvider{config: ObservabilityConfig{DSN: "https://key@sentry.example.com/1"}}
	err := p.Flush(1 * time.Millisecond)
	if err != nil {
		t.Errorf("Flush with MockTransport returned error: %v", err)
	}
}

func TestNewObservabilityClientFromProvider(t *testing.T) {
	t.Parallel()
	rp := &recordingProvider{}
	client := NewObservabilityClientFromProvider(rp)

	if !client.enabled {
		t.Error("client should be enabled")
	}

	err := client.SendError(errors.New("test"), nil)
	if err != nil {
		t.Fatalf("SendError returned error: %v", err)
	}

	rp.mu.Lock()
	defer rp.mu.Unlock()
	if len(rp.errors) != 1 {
		t.Errorf("expected 1 error recorded, got %d", len(rp.errors))
	}
}
