package plugin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// newTestEngine creates a RequestEngine backed by an in-memory SQLite DB
// with the plugin_requests table already created. Caller must defer pool.Close().
func newTestEngine(t *testing.T, cfg RequestEngineConfig) (*RequestEngine, func()) {
	t.Helper()
	pool := openTestDB(t)

	cfg.AllowLocalhost = true // tests use 127.0.0.1
	e := NewRequestEngine(pool, db.DialectSQLite, cfg)
	if err := e.CreatePluginRequestsTable(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("creating plugin_requests table: %v", err)
	}

	cleanup := func() {
		e.Close()
		pool.Close()
	}
	return e, cleanup
}

// ---------- DB operations ----------

func TestCreatePluginRequestsTable(t *testing.T) {
	pool := openTestDB(t)
	defer pool.Close()

	e := NewRequestEngine(pool, db.DialectSQLite, RequestEngineConfig{})
	defer e.Close()

	if err := e.CreatePluginRequestsTable(context.Background()); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// Idempotent -- second call should succeed.
	if err := e.CreatePluginRequestsTable(context.Background()); err != nil {
		t.Fatalf("second create (idempotent): %v", err)
	}
}

func TestUpsertAndListRequests(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{})
	defer cleanup()

	ctx := context.Background()
	requests := []PendingRequest{
		{Domain: "api.example.com", Description: "Example API"},
		{Domain: "data.other.io", Description: "Data service"},
	}

	if err := e.UpsertRequestRegistrations(ctx, "myplugin", "1.0.0", requests); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	listed, err := e.ListRequests(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected 2 registrations, got %d", len(listed))
	}
	if listed[0].Domain != "api.example.com" {
		t.Errorf("first domain = %q, want api.example.com", listed[0].Domain)
	}
	if listed[0].Approved {
		t.Error("new registration should not be approved")
	}
	if listed[0].PluginVersion != "1.0.0" {
		t.Errorf("version = %q, want 1.0.0", listed[0].PluginVersion)
	}

	// Re-upsert with updated version should not change approval.
	requests[0].Description = "Updated description"
	if err := e.UpsertRequestRegistrations(ctx, "myplugin", "2.0.0", requests); err != nil {
		t.Fatalf("re-upsert: %v", err)
	}

	listed, err = e.ListRequests(ctx)
	if err != nil {
		t.Fatalf("re-list: %v", err)
	}
	if listed[0].PluginVersion != "2.0.0" {
		t.Errorf("version after re-upsert = %q, want 2.0.0", listed[0].PluginVersion)
	}
	if listed[0].Description != "Updated description" {
		t.Errorf("description after re-upsert = %q, want Updated description", listed[0].Description)
	}
}

func TestApproveAndRevokeRequest(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{})
	defer cleanup()

	ctx := context.Background()
	requests := []PendingRequest{{Domain: "api.example.com", Description: "test"}}
	if err := e.UpsertRequestRegistrations(ctx, "myplugin", "1.0.0", requests); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Not approved initially.
	domains := e.ApprovedDomains()
	if len(domains) != 0 {
		t.Errorf("expected 0 approved domains, got %v", domains)
	}

	// Approve.
	if err := e.ApproveRequest(ctx, "myplugin", "api.example.com", "admin@test.com"); err != nil {
		t.Fatalf("approve: %v", err)
	}

	domains = e.ApprovedDomains()
	if len(domains) != 1 || domains[0] != "myplugin:api.example.com" {
		t.Errorf("approved domains = %v, want [myplugin:api.example.com]", domains)
	}

	// DB should also reflect approval.
	listed, err := e.ListRequests(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !listed[0].Approved {
		t.Error("listed request should be approved")
	}

	// Revoke.
	if err := e.RevokeRequest(ctx, "myplugin", "api.example.com"); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	domains = e.ApprovedDomains()
	if len(domains) != 0 {
		t.Errorf("expected 0 approved after revoke, got %v", domains)
	}
}

func TestCleanupOrphanedRequests(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{})
	defer cleanup()

	ctx := context.Background()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		requests := []PendingRequest{{Domain: "api.example.com", Description: "test"}}
		if err := e.UpsertRequestRegistrations(ctx, name, "1.0.0", requests); err != nil {
			t.Fatalf("upsert %s: %v", name, err)
		}
	}

	// Keep only "alpha".
	if err := e.CleanupOrphanedRequests(ctx, []string{"alpha"}); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	listed, err := e.ListRequests(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 after cleanup, got %d", len(listed))
	}
	if listed[0].PluginName != "alpha" {
		t.Errorf("remaining plugin = %q, want alpha", listed[0].PluginName)
	}
}

func TestCleanupOrphanedRequestsEmpty(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{})
	defer cleanup()

	ctx := context.Background()
	requests := []PendingRequest{{Domain: "api.example.com", Description: "test"}}
	if err := e.UpsertRequestRegistrations(ctx, "myplugin", "1.0.0", requests); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Empty discovered list clears all.
	if err := e.CleanupOrphanedRequests(ctx, nil); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	listed, err := e.ListRequests(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected 0 after cleanup, got %d", len(listed))
	}
}

func TestLoadApprovals(t *testing.T) {
	pool := openTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	// First engine: create table, upsert, approve.
	e1 := NewRequestEngine(pool, db.DialectSQLite, RequestEngineConfig{AllowLocalhost: true})
	if err := e1.CreatePluginRequestsTable(ctx); err != nil {
		t.Fatalf("create table: %v", err)
	}
	requests := []PendingRequest{{Domain: "api.example.com", Description: "test"}}
	if err := e1.UpsertRequestRegistrations(ctx, "myplugin", "1.0.0", requests); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := e1.ApproveRequest(ctx, "myplugin", "api.example.com", "admin"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	e1.Close()

	// Second engine: load approvals from DB into fresh in-memory map.
	e2 := NewRequestEngine(pool, db.DialectSQLite, RequestEngineConfig{AllowLocalhost: true})
	defer e2.Close()

	if err := e2.LoadApprovals(ctx); err != nil {
		t.Fatalf("load approvals: %v", err)
	}

	domains := e2.ApprovedDomains()
	if len(domains) != 1 || domains[0] != "myplugin:api.example.com" {
		t.Errorf("loaded domains = %v, want [myplugin:api.example.com]", domains)
	}
}

func TestUnregisterPlugin(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{})
	defer cleanup()

	ctx := context.Background()
	for _, name := range []string{"alpha", "beta"} {
		requests := []PendingRequest{{Domain: "api.example.com", Description: "test"}}
		if err := e.UpsertRequestRegistrations(ctx, name, "1.0.0", requests); err != nil {
			t.Fatalf("upsert %s: %v", name, err)
		}
		if err := e.ApproveRequest(ctx, name, "api.example.com", "admin"); err != nil {
			t.Fatalf("approve %s: %v", name, err)
		}
	}

	e.UnregisterPlugin("alpha")

	domains := e.ApprovedDomains()
	if len(domains) != 1 || domains[0] != "beta:api.example.com" {
		t.Errorf("after unregister, domains = %v, want [beta:api.example.com]", domains)
	}
}

// ---------- Execute ----------

func TestExecuteApprovedDomain(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{})
	defer cleanup()

	// Start a test HTTPS server (we'll use HTTP with AllowLocalhost for test).
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer ts.Close()

	// Extract host from test server URL.
	host := strings.TrimPrefix(ts.URL, "http://")

	ctx := context.Background()

	// Approve the domain.
	key := "testplugin:" + strings.Split(host, ":")[0]
	e.mu.Lock()
	e.approved[key] = true
	e.mu.Unlock()

	result, err := e.Execute(ctx, "testplugin", "GET", ts.URL, OutboundRequestOpts{ParseJSON: true})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	status, ok := result["status"].(int)
	if !ok || status != 200 {
		t.Errorf("status = %v, want 200", result["status"])
	}
	if jsonData, ok := result["json"].(map[string]any); ok {
		if jsonData["ok"] != true {
			t.Errorf("json.ok = %v, want true", jsonData["ok"])
		}
	} else {
		t.Error("expected parsed JSON in result")
	}
}

func TestExecuteUnapprovedDomain(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{})
	defer cleanup()

	_, err := e.Execute(context.Background(), "testplugin", "GET", "http://127.0.0.1:9999/test", OutboundRequestOpts{})
	if err == nil {
		t.Fatal("expected error for unapproved domain")
	}
	if !strings.Contains(err.Error(), "domain not approved") {
		t.Errorf("error = %q, want 'domain not approved'", err.Error())
	}
}

func TestExecuteHTTPSRequired(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{AllowLocalhost: false})
	defer cleanup()

	// Override AllowLocalhost to false for this test.
	e.cfg.AllowLocalhost = false

	_, err := e.Execute(context.Background(), "testplugin", "GET", "http://example.com/test", OutboundRequestOpts{})
	if err == nil {
		t.Fatal("expected error for HTTP scheme")
	}
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("error = %q, want mention of HTTPS", err.Error())
	}
}

func TestExecuteUserinfoBlocked(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{})
	defer cleanup()

	_, err := e.Execute(context.Background(), "testplugin", "GET", "https://user:pass@example.com/test", OutboundRequestOpts{})
	if err == nil {
		t.Fatal("expected error for URL with userinfo")
	}
	if !strings.Contains(err.Error(), "userinfo") {
		t.Errorf("error = %q, want mention of userinfo", err.Error())
	}
}

func TestExecuteUnsupportedScheme(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{})
	defer cleanup()

	_, err := e.Execute(context.Background(), "testplugin", "GET", "ftp://example.com/file", OutboundRequestOpts{})
	if err == nil {
		t.Fatal("expected error for ftp scheme")
	}
	if !strings.Contains(err.Error(), "unsupported URL scheme") {
		t.Errorf("error = %q, want 'unsupported URL scheme'", err.Error())
	}
}

func TestExecuteBodySizeLimit(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{MaxRequestBodyBytes: 100})
	defer cleanup()

	// Approve the domain.
	e.mu.Lock()
	e.approved["testplugin:127.0.0.1"] = true
	e.mu.Unlock()

	largeBody := strings.Repeat("x", 200)
	_, err := e.Execute(context.Background(), "testplugin", "POST", "http://127.0.0.1:9999/test", OutboundRequestOpts{Body: largeBody})
	if err == nil {
		t.Fatal("expected error for oversized body")
	}
	if !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Errorf("error = %q, want 'exceeds maximum size'", err.Error())
	}
}

func TestExecuteJSONBodySizeLimit(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{MaxRequestBodyBytes: 10})
	defer cleanup()

	e.mu.Lock()
	e.approved["testplugin:127.0.0.1"] = true
	e.mu.Unlock()

	_, err := e.Execute(context.Background(), "testplugin", "POST", "http://127.0.0.1:9999/test", OutboundRequestOpts{
		JSONBody: map[string]string{"key": "a very long value that exceeds the limit"},
	})
	if err == nil {
		t.Fatal("expected error for oversized JSON body")
	}
	if !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Errorf("error = %q, want 'exceeds maximum size'", err.Error())
	}
}

func TestExecuteResponseSizeLimit(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{MaxResponseBytes: 50})
	defer cleanup()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, strings.Repeat("x", 100))
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "http://")
	e.mu.Lock()
	e.approved["testplugin:"+strings.Split(host, ":")[0]] = true
	e.mu.Unlock()

	_, err := e.Execute(context.Background(), "testplugin", "GET", ts.URL, OutboundRequestOpts{})
	if err == nil {
		t.Fatal("expected error for oversized response")
	}
	if !strings.Contains(err.Error(), "response exceeded maximum size") {
		t.Errorf("error = %q, want 'response exceeded maximum size'", err.Error())
	}
}

// ---------- Rate limiting ----------

func TestRateLimiterBlocks(t *testing.T) {
	// 1 request/min with burst 1 — second request should be blocked.
	e, cleanup := newTestEngine(t, RequestEngineConfig{MaxRequestsPerMin: 1})
	defer cleanup()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "http://")
	e.mu.Lock()
	e.approved["testplugin:"+strings.Split(host, ":")[0]] = true
	e.mu.Unlock()

	// First request should succeed.
	_, err := e.Execute(context.Background(), "testplugin", "GET", ts.URL, OutboundRequestOpts{})
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	// Second should be rate limited.
	_, err = e.Execute(context.Background(), "testplugin", "GET", ts.URL, OutboundRequestOpts{})
	if err == nil {
		t.Fatal("expected rate limit error on second request")
	}
	if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("error = %q, want 'rate limit exceeded'", err.Error())
	}
}

func TestGlobalRateLimiterBlocks(t *testing.T) {
	// Global limit of 1/min, burst 1.
	e, cleanup := newTestEngine(t, RequestEngineConfig{
		MaxRequestsPerMin:       600,
		GlobalMaxRequestsPerMin: 1,
	})
	defer cleanup()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "http://")
	e.mu.Lock()
	e.approved["testplugin:"+strings.Split(host, ":")[0]] = true
	e.mu.Unlock()

	// First request should succeed.
	_, err := e.Execute(context.Background(), "testplugin", "GET", ts.URL, OutboundRequestOpts{})
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	// Second should be globally rate limited.
	_, err = e.Execute(context.Background(), "testplugin", "GET", ts.URL, OutboundRequestOpts{})
	if err == nil {
		t.Fatal("expected global rate limit error")
	}
	if !strings.Contains(err.Error(), "global outbound request rate limit") {
		t.Errorf("error = %q, want 'global outbound request rate limit'", err.Error())
	}
}

// ---------- Circuit breaker ----------

func TestCircuitBreakerTrips(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{
		MaxRequestsPerMin:  6000, // high limit so rate limiter doesn't interfere
		CBMaxFailures:      2,
		CBResetIntervalSec: 3600, // long reset so it stays open
	})
	defer cleanup()

	// Server that returns 500s.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "http://")
	key := "testplugin:" + strings.Split(host, ":")[0]
	e.mu.Lock()
	e.approved[key] = true
	e.mu.Unlock()

	// First 2 requests get 500 responses (not errors, but CB counts them).
	for i := range 2 {
		result, err := e.Execute(context.Background(), "testplugin", "GET", ts.URL, OutboundRequestOpts{})
		if err != nil {
			t.Fatalf("request %d failed: %v", i+1, err)
		}
		if result["status"] != 500 {
			t.Errorf("request %d status = %v, want 500", i+1, result["status"])
		}
	}

	// Third request should be blocked by circuit breaker.
	_, err := e.Execute(context.Background(), "testplugin", "GET", ts.URL, OutboundRequestOpts{})
	if err == nil {
		t.Fatal("expected circuit breaker error")
	}
	if !strings.Contains(err.Error(), "circuit breaker open") {
		t.Errorf("error = %q, want 'circuit breaker open'", err.Error())
	}
}

func TestCircuitBreakerResetsOnSuccess(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{
		MaxRequestsPerMin: 6000, // high limit so rate limiter doesn't interfere
		CBMaxFailures:     3,
	})
	defer cleanup()

	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "http://")
	key := "testplugin:" + strings.Split(host, ":")[0]
	e.mu.Lock()
	e.approved[key] = true
	e.mu.Unlock()

	// 2 failures, then success.
	for range 3 {
		_, execErr := e.Execute(context.Background(), "testplugin", "GET", ts.URL, OutboundRequestOpts{})
		if execErr != nil {
			t.Fatalf("unexpected error: %v", execErr)
		}
	}

	// CB should be reset -- verify by checking internal state.
	e.cbMu.RLock()
	cb := e.circuitBreakers[key]
	e.cbMu.RUnlock()

	if cb != nil && cb.disabled {
		t.Error("circuit breaker should not be disabled after success")
	}
	if cb != nil && cb.consecutiveFailures != 0 {
		t.Errorf("consecutive failures = %d, want 0", cb.consecutiveFailures)
	}
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	e, cleanup := newTestEngine(t, RequestEngineConfig{
		CBMaxFailures:      1,
		CBResetIntervalSec: 1, // 1 second reset for test speed
	})
	defer cleanup()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "http://")
	key := "testplugin:" + strings.Split(host, ":")[0]
	e.mu.Lock()
	e.approved[key] = true
	e.mu.Unlock()

	// Manually trip the circuit breaker.
	e.cbMu.Lock()
	e.circuitBreakers[key] = &domainCB{
		consecutiveFailures: 1,
		disabled:            true,
		lastFailure:         time.Now().Add(-2 * time.Second), // past reset interval
	}
	e.cbMu.Unlock()

	// Half-open probe should succeed since lastFailure is past reset interval.
	result, err := e.Execute(context.Background(), "testplugin", "GET", ts.URL, OutboundRequestOpts{})
	if err != nil {
		t.Fatalf("half-open probe failed: %v", err)
	}
	if result["status"] != 200 {
		t.Errorf("status = %v, want 200", result["status"])
	}

	// CB should be reset after successful probe.
	e.cbMu.RLock()
	cb := e.circuitBreakers[key]
	e.cbMu.RUnlock()
	if cb.disabled {
		t.Error("circuit breaker should be reset after successful probe")
	}
}

// ---------- Config defaults ----------

func TestRequestEngineDefaults(t *testing.T) {
	cfg := RequestEngineConfig{}
	requestEngineDefaults(&cfg)

	if cfg.DefaultTimeoutSec != 10 {
		t.Errorf("DefaultTimeoutSec = %d, want 10", cfg.DefaultTimeoutSec)
	}
	if cfg.MaxResponseBytes != 1_048_576 {
		t.Errorf("MaxResponseBytes = %d, want 1048576", cfg.MaxResponseBytes)
	}
	if cfg.MaxRequestBodyBytes != 1_048_576 {
		t.Errorf("MaxRequestBodyBytes = %d, want 1048576", cfg.MaxRequestBodyBytes)
	}
	if cfg.MaxRequestsPerMin != 60 {
		t.Errorf("MaxRequestsPerMin = %d, want 60", cfg.MaxRequestsPerMin)
	}
	// GlobalMaxRequestsPerMin: 0 means unlimited (no global cap).
	// Default only applies for negative values.
	if cfg.GlobalMaxRequestsPerMin != 0 {
		t.Errorf("GlobalMaxRequestsPerMin = %d, want 0 (unlimited)", cfg.GlobalMaxRequestsPerMin)
	}
	if cfg.CBMaxFailures != 5 {
		t.Errorf("CBMaxFailures = %d, want 5", cfg.CBMaxFailures)
	}
	if cfg.CBResetIntervalSec != 60 {
		t.Errorf("CBResetIntervalSec = %d, want 60", cfg.CBResetIntervalSec)
	}
}

func TestRequestEngineDefaultsPreservesExplicit(t *testing.T) {
	cfg := RequestEngineConfig{
		DefaultTimeoutSec:       30,
		MaxResponseBytes:        500,
		MaxRequestBodyBytes:     500,
		MaxRequestsPerMin:       120,
		GlobalMaxRequestsPerMin: 1200,
		CBMaxFailures:           10,
		CBResetIntervalSec:      120,
	}
	requestEngineDefaults(&cfg)

	if cfg.DefaultTimeoutSec != 30 {
		t.Errorf("DefaultTimeoutSec = %d, want 30", cfg.DefaultTimeoutSec)
	}
	if cfg.MaxResponseBytes != 500 {
		t.Errorf("MaxResponseBytes = %d, want 500", cfg.MaxResponseBytes)
	}
	if cfg.MaxRequestsPerMin != 120 {
		t.Errorf("MaxRequestsPerMin = %d, want 120", cfg.MaxRequestsPerMin)
	}
	if cfg.GlobalMaxRequestsPerMin != 1200 {
		t.Errorf("GlobalMaxRequestsPerMin = %d, want 1200", cfg.GlobalMaxRequestsPerMin)
	}
}

// ---------- categorizeHTTPError ----------

func TestCategorizeHTTPError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "timeout", err: fmt.Errorf("context deadline exceeded"), want: "timed out"},
		{name: "connection refused", err: fmt.Errorf("connection refused"), want: "connection refused"},
		{name: "dns failure", err: fmt.Errorf("no such host"), want: "dns lookup failed"},
		{name: "tls error", err: fmt.Errorf("tls handshake failure"), want: "tls handshake failed"},
		{name: "canceled", err: fmt.Errorf("context canceled"), want: "request canceled"},
		{name: "ssrf blocked", err: fmt.Errorf("private/reserved IP blocked"), want: "private/reserved IP address blocked"},
		{name: "generic", err: fmt.Errorf("something else"), want: "request failed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := categorizeHTTPError(tc.err, "example.com", 10)
			if !strings.Contains(msg, tc.want) {
				t.Errorf("categorizeHTTPError(%q) = %q, want to contain %q", tc.err, msg, tc.want)
			}
		})
	}
}

// ---------- Unsupported dialect ----------

func TestCreatePluginRequestsTableUnsupportedDialect(t *testing.T) {
	pool := openTestDB(t)
	defer pool.Close()

	e := NewRequestEngine(pool, db.Dialect(99), RequestEngineConfig{})
	defer e.Close()

	err := e.CreatePluginRequestsTable(context.Background())
	if err == nil {
		t.Fatal("expected error for unsupported dialect")
	}
	if !strings.Contains(err.Error(), "unsupported dialect") {
		t.Errorf("error = %q, want 'unsupported dialect'", err.Error())
	}
}
