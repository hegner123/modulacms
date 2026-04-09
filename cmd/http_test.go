package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// apiClient.do
// ---------------------------------------------------------------------------

func TestAPIClient_Do_SetsAuthorizationHeader(t *testing.T) {
	t.Parallel()

	var gotAuth string
	var gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "test-token-123",
		http:    srv.Client(),
	}

	resp, err := client.do("GET", "/api/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if gotAuth != "Bearer test-token-123" {
		t.Errorf("Authorization header: got %q, want %q", gotAuth, "Bearer test-token-123")
	}

	// No body means no Content-Type header should be set
	if gotContentType != "" {
		t.Errorf("Content-Type should be empty for nil body, got %q", gotContentType)
	}
}

func TestAPIClient_Do_SetsContentTypeWithBody(t *testing.T) {
	t.Parallel()

	var gotContentType string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	body := `{"key":"value"}`
	resp, err := client.do("POST", "/api/test", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if gotContentType != "application/json" {
		t.Errorf("Content-Type: got %q, want %q", gotContentType, "application/json")
	}
	if gotBody != body {
		t.Errorf("body: got %q, want %q", gotBody, body)
	}
}

func TestAPIClient_Do_ConstructsCorrectURL(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	resp, err := client.do("DELETE", "/api/v1/admin/plugins/test-plugin", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if gotMethod != "DELETE" {
		t.Errorf("method: got %q, want %q", gotMethod, "DELETE")
	}
	if gotPath != "/api/v1/admin/plugins/test-plugin" {
		t.Errorf("path: got %q, want %q", gotPath, "/api/v1/admin/plugins/test-plugin")
	}
}

func TestAPIClient_Do_ConnectionError(t *testing.T) {
	t.Parallel()

	client := &apiClient{
		baseURL: "http://127.0.0.1:1", // port 1 -- connection refused
		token:   "tok",
		http:    &http.Client{},
	}

	_, err := client.do("GET", "/api/test", nil)
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
	if !strings.Contains(err.Error(), "sending request") {
		t.Errorf("expected error about sending request, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// handleSimpleResponse
// ---------------------------------------------------------------------------

// newCaptureCmd creates a cobra.Command with captured stdout and stderr buffers.
func newCaptureCmd(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd, stdout, stderr
}

func TestHandleSimpleResponse_Success(t *testing.T) {
	t.Parallel()

	cmd, stdout, _ := newCaptureCmd(t)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	err := handleSimpleResponse(cmd, resp, "my_plugin", "reloaded")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, `Plugin "my_plugin" reloaded successfully.`) {
		t.Errorf("expected success message, got: %q", output)
	}
}

// Note: handleSimpleResponse calls os.Exit(1) for 404 and non-200 responses.
// Testing those paths requires subprocess execution via exec.Command.
// This is intentionally deferred to integration tests since the function's
// design couples exit behavior with response handling.

// ---------------------------------------------------------------------------
// handlerSwap as http.Handler with real HTTP server
// ---------------------------------------------------------------------------

func TestHandlerSwap_WithRealHTTPServer(t *testing.T) {
	t.Parallel()

	swap := &handlerSwap{}
	swap.set(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("initial"))
	}))

	srv := httptest.NewServer(swap)
	t.Cleanup(srv.Close)

	// Verify initial handler
	resp, err := srv.Client().Get(srv.URL + "/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != "initial" {
		t.Errorf("expected 'initial', got %q", string(body))
	}

	// Swap handler
	swap.set(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("swapped"))
	}))

	resp2, err := srv.Client().Get(srv.URL + "/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	if resp2.StatusCode != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", resp2.StatusCode)
	}
	if string(body2) != "swapped" {
		t.Errorf("expected 'swapped', got %q", string(body2))
	}
}

// ---------------------------------------------------------------------------
// fetchRoutesFiltered / fetchHooksFiltered with httptest server
// ---------------------------------------------------------------------------

func TestFetchRoutesFiltered(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"routes": [
				{"plugin": "test_plugin", "method": "GET", "path": "/tasks", "approved": false},
				{"plugin": "test_plugin", "method": "POST", "path": "/tasks", "approved": true},
				{"plugin": "other_plugin", "method": "GET", "path": "/items", "approved": false}
			]
		}`))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	t.Run("filter unapproved for test_plugin", func(t *testing.T) {
		t.Parallel()
		routes, err := fetchRoutesFiltered(client, "test_plugin", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(routes))
		}
		if routes[0].Method != "GET" || routes[0].Path != "/tasks" {
			t.Errorf("expected GET /tasks, got %s %s", routes[0].Method, routes[0].Path)
		}
	})

	t.Run("filter approved for test_plugin", func(t *testing.T) {
		t.Parallel()
		routes, err := fetchRoutesFiltered(client, "test_plugin", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(routes))
		}
		if routes[0].Method != "POST" {
			t.Errorf("expected POST method, got %s", routes[0].Method)
		}
	})

	t.Run("filter for nonexistent plugin returns empty", func(t *testing.T) {
		t.Parallel()
		routes, err := fetchRoutesFiltered(client, "nonexistent", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(routes) != 0 {
			t.Errorf("expected 0 routes, got %d", len(routes))
		}
	})
}

func TestFetchHooksFiltered(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"hooks": [
				{"plugin_name": "my_plugin", "event": "after_insert", "table": "content_data", "approved": false},
				{"plugin_name": "my_plugin", "event": "before_delete", "table": "media", "approved": true},
				{"plugin_name": "other_plugin", "event": "after_update", "table": "users", "approved": false}
			]
		}`))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	t.Run("filter unapproved for my_plugin", func(t *testing.T) {
		t.Parallel()
		hooks, err := fetchHooksFiltered(client, "my_plugin", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hooks) != 1 {
			t.Fatalf("expected 1 hook, got %d", len(hooks))
		}
		if hooks[0].Event != "after_insert" || hooks[0].Table != "content_data" {
			t.Errorf("expected after_insert:content_data, got %s:%s", hooks[0].Event, hooks[0].Table)
		}
	})

	t.Run("filter approved for my_plugin", func(t *testing.T) {
		t.Parallel()
		hooks, err := fetchHooksFiltered(client, "my_plugin", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hooks) != 1 {
			t.Fatalf("expected 1 hook, got %d", len(hooks))
		}
		if hooks[0].Event != "before_delete" {
			t.Errorf("expected before_delete, got %s", hooks[0].Event)
		}
	})

	t.Run("filter for nonexistent plugin returns empty", func(t *testing.T) {
		t.Parallel()
		hooks, err := fetchHooksFiltered(client, "nonexistent", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hooks) != 0 {
			t.Errorf("expected 0 hooks, got %d", len(hooks))
		}
	})
}

func TestFetchRoutesFiltered_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	_, err := fetchRoutesFiltered(client, "test", false)
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("expected 'server error' in message, got: %v", err)
	}
}

func TestFetchHooksFiltered_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	_, err := fetchHooksFiltered(client, "test", false)
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("expected 'server error' in message, got: %v", err)
	}
}

func TestFetchRoutesFiltered_InvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	_, err := fetchRoutesFiltered(client, "test", false)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "decoding routes") {
		t.Errorf("expected 'decoding routes' in error, got: %v", err)
	}
}

func TestFetchHooksFiltered_InvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	_, err := fetchHooksFiltered(client, "test", false)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "decoding hooks") {
		t.Errorf("expected 'decoding hooks' in error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// approveRoutes / revokeRoutes / approveHooks / revokeHooks with httptest
// ---------------------------------------------------------------------------

func TestApproveRoutes_Success(t *testing.T) {
	t.Parallel()

	var gotMethod string
	var gotPath string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	routes := []routeItem{{Plugin: "test", Method: "GET", Path: "/tasks"}}

	err := approveRoutes(cmd, client, "test", routes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method: got %q, want POST", gotMethod)
	}
	if gotPath != "/api/v1/admin/plugins/routes/approve" {
		t.Errorf("path: got %q, want /api/v1/admin/plugins/routes/approve", gotPath)
	}
	if !strings.Contains(gotBody, `"method":"GET"`) {
		t.Errorf("expected body to contain route method, got: %s", gotBody)
	}
	if !strings.Contains(stdout.String(), "Approved 1 route(s)") {
		t.Errorf("expected success message, got: %q", stdout.String())
	}
}

func TestRevokeRoutes_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	routes := []routeItem{
		{Plugin: "test", Method: "GET", Path: "/a"},
		{Plugin: "test", Method: "POST", Path: "/b"},
	}

	err := revokeRoutes(cmd, client, "test", routes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Revoked 2 route(s)") {
		t.Errorf("expected revoke message, got: %q", stdout.String())
	}
}

func TestApproveHooks_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	hooks := []hookItem{{Plugin: "test", Event: "after_insert", Table: "content_data"}}

	err := approveHooks(cmd, client, "test", hooks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Approved 1 hook(s)") {
		t.Errorf("expected success message, got: %q", stdout.String())
	}
}

func TestRevokeHooks_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	hooks := []hookItem{
		{Plugin: "test", Event: "after_insert", Table: "content_data"},
		{Plugin: "test", Event: "before_delete", Table: "media"},
	}

	err := revokeHooks(cmd, client, "test", hooks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Revoked 2 hook(s)") {
		t.Errorf("expected revoke message, got: %q", stdout.String())
	}
}

// ---------------------------------------------------------------------------
// approveAllRoutes / revokeAllRoutes / approveAllHooks / revokeAllHooks
// (empty list case -- no pending items)
// ---------------------------------------------------------------------------

func TestApproveAllRoutes_NoPending(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"routes": []}`))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	err := approveAllRoutes(cmd, client, "test_plugin", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "no pending routes") {
		t.Errorf("expected 'No pending routes' message, got: %q", stdout.String())
	}
}

func TestRevokeAllRoutes_NoApproved(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"routes": []}`))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	err := revokeAllRoutes(cmd, client, "test_plugin", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "no approved routes") {
		t.Errorf("expected 'No approved routes' message, got: %q", stdout.String())
	}
}

func TestApproveAllHooks_NoPending(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"hooks": []}`))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	err := approveAllHooks(cmd, client, "test_plugin", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "no pending hooks") {
		t.Errorf("expected 'No pending hooks' message, got: %q", stdout.String())
	}
}

func TestRevokeAllHooks_NoApproved(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"hooks": []}`))
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	err := revokeAllHooks(cmd, client, "test_plugin", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "no approved hooks") {
		t.Errorf("expected 'No approved hooks' message, got: %q", stdout.String())
	}
}

// ---------------------------------------------------------------------------
// approveAllRoutes with --yes (skipPrompt) flow-through to approve endpoint
// ---------------------------------------------------------------------------

func TestApproveAllRoutes_WithYes_ApprovesPending(t *testing.T) {
	t.Parallel()

	// Track which endpoints get hit
	var approveHit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/admin/plugins/routes" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"routes": [
					{"plugin": "my_plugin", "method": "GET", "path": "/items", "approved": false},
					{"plugin": "my_plugin", "method": "POST", "path": "/items", "approved": false}
				]
			}`))
		case r.URL.Path == "/api/v1/admin/plugins/routes/approve" && r.Method == "POST":
			approveHit = true
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	err := approveAllRoutes(cmd, client, "my_plugin", true) // skipPrompt = true
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approveHit {
		t.Error("expected approve endpoint to be called")
	}
	if !strings.Contains(stdout.String(), "Approved 2 route(s)") {
		t.Errorf("expected 'Approved 2 route(s)' message, got: %q", stdout.String())
	}
}

func TestRevokeAllHooks_WithYes_RevokesApproved(t *testing.T) {
	t.Parallel()

	var revokeHit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/admin/plugins/hooks" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"hooks": [
					{"plugin_name": "my_plugin", "event": "after_insert", "table": "content_data", "approved": true}
				]
			}`))
		case r.URL.Path == "/api/v1/admin/plugins/hooks/revoke" && r.Method == "POST":
			revokeHit = true
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	client := &apiClient{
		baseURL: srv.URL,
		token:   "tok",
		http:    srv.Client(),
	}

	cmd, stdout, _ := newCaptureCmd(t)
	err := revokeAllHooks(cmd, client, "my_plugin", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !revokeHit {
		t.Error("expected revoke endpoint to be called")
	}
	if !strings.Contains(stdout.String(), "Revoked 1 hook(s)") {
		t.Errorf("expected 'Revoked 1 hook(s)' message, got: %q", stdout.String())
	}
}
