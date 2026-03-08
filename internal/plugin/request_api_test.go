package plugin

import (
	"context"
	"fmt"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// mockExecutor implements outboundExecutor for testing.
type mockExecutor struct {
	lastMethod string
	lastURL    string
	lastOpts   OutboundRequestOpts
	result     map[string]any
	err        error
}

func (m *mockExecutor) Execute(_ context.Context, pluginName, method, urlStr string, opts OutboundRequestOpts) (map[string]any, error) {
	m.lastMethod = method
	m.lastURL = urlStr
	m.lastOpts = opts
	return m.result, m.err
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		wantErr string
	}{
		{name: "valid domain", domain: "api.example.com"},
		{name: "valid subdomain", domain: "data.external.io"},
		{name: "valid with hyphen", domain: "my-api.example.com"},
		{name: "empty", domain: "", wantErr: "cannot be empty"},
		{name: "too long", domain: strings.Repeat("a", 254), wantErr: "exceeds 253"},
		{name: "has scheme", domain: "https://example.com", wantErr: "scheme"},
		{name: "has path", domain: "example.com/api", wantErr: "path"},
		{name: "has port", domain: "example.com:8080", wantErr: "port"},
		{name: "has wildcard", domain: "*.example.com", wantErr: "wildcard"},
		{name: "invalid char space", domain: "example .com", wantErr: "invalid character"},
		{name: "invalid char underscore", domain: "my_api.example.com", wantErr: "invalid character"},
		{name: "no dot", domain: "localhost", wantErr: "at least one dot"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDomain(tc.domain)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestRequestRegister(t *testing.T) {
	t.Run("registers domain at module scope", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.register("api.example.com", {description = "Test API"})`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		reqs := ReadPendingRequests(L)
		if len(reqs) != 1 {
			t.Fatalf("expected 1 pending request, got %d", len(reqs))
		}
		if reqs[0].Domain != "api.example.com" {
			t.Errorf("expected domain 'api.example.com', got %q", reqs[0].Domain)
		}
		if reqs[0].Description != "Test API" {
			t.Errorf("expected description 'Test API', got %q", reqs[0].Description)
		}
	})

	t.Run("normalizes domain to lowercase", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.register("API.Example.COM")`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		reqs := ReadPendingRequests(L)
		if len(reqs) != 1 {
			t.Fatalf("expected 1 pending request, got %d", len(reqs))
		}
		if reqs[0].Domain != "api.example.com" {
			t.Errorf("expected lowercase domain, got %q", reqs[0].Domain)
		}
	})

	t.Run("rejects during on_init", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "init")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.register("api.example.com")`)
		if err == nil {
			t.Fatal("expected error for request.register inside on_init")
		}
		if !strings.Contains(err.Error(), "module scope") {
			t.Errorf("expected error about module scope, got: %v", err)
		}
	})

	t.Run("rejects during runtime", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.register("api.example.com")`)
		if err == nil {
			t.Fatal("expected error for request.register at runtime")
		}
		if !strings.Contains(err.Error(), "module scope") {
			t.Errorf("expected error about module scope, got: %v", err)
		}
	})

	t.Run("duplicate is idempotent", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`
			request.register("api.example.com", {description = "First"})
			request.register("api.example.com", {description = "Second"})
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		reqs := ReadPendingRequests(L)
		if len(reqs) != 1 {
			t.Fatalf("expected 1 pending request (duplicate ignored), got %d", len(reqs))
		}
		if reqs[0].Description != "First" {
			t.Errorf("expected original description 'First', got %q", reqs[0].Description)
		}
	})

	t.Run("count limit enforced", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		// Register MaxDomainsPerPlugin domains.
		for i := range MaxDomainsPerPlugin {
			err := L.DoString(fmt.Sprintf(`request.register("d%d.example.com")`, i))
			if err != nil {
				t.Fatalf("domain %d registration failed: %s", i, err)
			}
		}

		// The next one should fail.
		err := L.DoString(`request.register("overflow.example.com")`)
		if err == nil {
			t.Fatal("expected error for exceeding domain limit")
		}
		if !strings.Contains(err.Error(), "maximum domain registration limit") {
			t.Errorf("expected error about limit, got: %v", err)
		}
	})

	t.Run("validates domain format", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.register("https://bad.com")`)
		if err == nil {
			t.Fatal("expected error for domain with scheme")
		}
	})

	t.Run("description is optional", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.register("api.example.com")`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		reqs := ReadPendingRequests(L)
		if len(reqs) != 1 {
			t.Fatalf("expected 1 pending request, got %d", len(reqs))
		}
		if reqs[0].Description != "" {
			t.Errorf("expected empty description, got %q", reqs[0].Description)
		}
	})
}

func TestRequestSend(t *testing.T) {
	t.Run("phase guard rejects at module scope", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.send("GET", "https://api.example.com/test")`)
		if err == nil {
			t.Fatal("expected error for send at module scope")
		}
		if !strings.Contains(err.Error(), "module scope") {
			t.Errorf("expected error about module scope, got: %v", err)
		}
	})

	t.Run("phase guard rejects during init", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "init")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.send("GET", "https://api.example.com/test")`)
		if err == nil {
			t.Fatal("expected error for send during init")
		}
		if !strings.Contains(err.Error(), "on_init") {
			t.Errorf("expected error about on_init, got: %v", err)
		}
	})

	t.Run("phase guard rejects during shutdown", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "shutdown")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.send("GET", "https://api.example.com/test")`)
		if err == nil {
			t.Fatal("expected error for send during shutdown")
		}
		if !strings.Contains(err.Error(), "on_shutdown") {
			t.Errorf("expected error about on_shutdown, got: %v", err)
		}
	})

	t.Run("before-hook guard rejects", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")
		state := RegisterRequestAPI(L, "test_plugin", nil)
		state.inBeforeHook = true

		ctx := context.Background()
		L.SetContext(ctx)

		err := L.DoString(`request.send("GET", "https://api.example.com/test")`)
		if err == nil {
			t.Fatal("expected error for send inside before-hook")
		}
		if !strings.Contains(err.Error(), "before-hook") {
			t.Errorf("expected error about before-hook, got: %v", err)
		}
	})

	t.Run("rejects URL with userinfo", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")
		RegisterRequestAPI(L, "test_plugin", nil)

		ctx := context.Background()
		L.SetContext(ctx)

		err := L.DoString(`request.send("GET", "https://user:pass@api.example.com/test")`)
		if err == nil {
			t.Fatal("expected error for URL with userinfo")
		}
		if !strings.Contains(err.Error(), "userinfo") {
			t.Errorf("expected error about userinfo, got: %v", err)
		}
	})

	t.Run("rejects URL without scheme", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")
		RegisterRequestAPI(L, "test_plugin", nil)

		ctx := context.Background()
		L.SetContext(ctx)

		err := L.DoString(`request.send("GET", "api.example.com/test")`)
		if err == nil {
			t.Fatal("expected error for URL without scheme")
		}
		if !strings.Contains(err.Error(), "scheme and host") {
			t.Errorf("expected error about scheme and host, got: %v", err)
		}
	})

	t.Run("rejects invalid HTTP method", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")
		RegisterRequestAPI(L, "test_plugin", nil)

		ctx := context.Background()
		L.SetContext(ctx)

		err := L.DoString(`request.send("INVALID", "https://api.example.com/test")`)
		if err == nil {
			t.Fatal("expected error for invalid method")
		}
		if !strings.Contains(err.Error(), "invalid HTTP method") {
			t.Errorf("expected error about invalid method, got: %v", err)
		}
	})

	t.Run("rejects body and json together", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")
		RegisterRequestAPI(L, "test_plugin", nil)

		ctx := context.Background()
		L.SetContext(ctx)

		err := L.DoString(`request.send("POST", "https://api.example.com/test", {body = "raw", json = {key = "val"}})`)
		if err == nil {
			t.Fatal("expected error for body + json")
		}
		if !strings.Contains(err.Error(), "mutually exclusive") {
			t.Errorf("expected error about mutual exclusivity, got: %v", err)
		}
	})

	t.Run("returns error table when engine is nil", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")
		RegisterRequestAPI(L, "test_plugin", nil) // nil engine

		ctx := context.Background()
		L.SetContext(ctx)

		err := L.DoString(`
			local resp = request.send("GET", "https://api.example.com/test")
			assert(resp.error == "outbound requests not configured", "expected error about not configured, got: " .. tostring(resp.error))
		`)
		if err != nil {
			t.Fatalf("expected no Lua error, got: %v", err)
		}
	})

	t.Run("calls engine and returns response", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")

		mock := &mockExecutor{
			result: map[string]any{
				"status":  200,
				"body":    `{"ok":true}`,
				"headers": map[string]string{"content-type": "application/json"},
			},
		}
		RegisterRequestAPI(L, "test_plugin", mock)

		ctx := context.Background()
		L.SetContext(ctx)

		err := L.DoString(`
			local resp = request.send("POST", "https://api.example.com/webhook", {
				headers = {["authorization"] = "Bearer token123"},
				json = {event = "published"},
				timeout = 5,
				parse_json = true
			})
			assert(resp.status == 200, "expected status 200, got: " .. tostring(resp.status))
			assert(resp.body == '{"ok":true}', "unexpected body: " .. tostring(resp.body))
			assert(resp.headers["content-type"] == "application/json", "unexpected content-type")
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if mock.lastMethod != "POST" {
			t.Errorf("expected method POST, got %q", mock.lastMethod)
		}
		if mock.lastURL != "https://api.example.com/webhook" {
			t.Errorf("expected URL https://api.example.com/webhook, got %q", mock.lastURL)
		}
		if mock.lastOpts.Headers["authorization"] != "Bearer token123" {
			t.Errorf("expected authorization header, got %v", mock.lastOpts.Headers)
		}
		if mock.lastOpts.Timeout != 5 {
			t.Errorf("expected timeout 5, got %d", mock.lastOpts.Timeout)
		}
		if !mock.lastOpts.ParseJSON {
			t.Error("expected parse_json true")
		}
	})

	t.Run("engine error returns error table", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")

		mock := &mockExecutor{
			err: fmt.Errorf("domain not approved: api.example.com"),
		}
		RegisterRequestAPI(L, "test_plugin", mock)

		ctx := context.Background()
		L.SetContext(ctx)

		err := L.DoString(`
			local resp = request.send("GET", "https://api.example.com/test")
			assert(resp.error == "domain not approved: api.example.com",
				"expected domain not approved error, got: " .. tostring(resp.error))
		`)
		if err != nil {
			t.Fatalf("expected no Lua error, got: %v", err)
		}
	})
}

func TestRequestConvenienceMethods(t *testing.T) {
	methods := []struct {
		luaFn      string
		httpMethod string
	}{
		{"get", "GET"},
		{"post", "POST"},
		{"put", "PUT"},
		{"delete", "DELETE"},
		{"patch", "PATCH"},
	}

	for _, tc := range methods {
		t.Run(tc.luaFn, func(t *testing.T) {
			L := lua.NewState(lua.Options{SkipOpenLibs: true})
			defer L.Close()
			ApplySandbox(L, SandboxConfig{})
			setVMPhase(L, "runtime")

			mock := &mockExecutor{
				result: map[string]any{"status": 200, "body": "ok"},
			}
			RegisterRequestAPI(L, "test_plugin", mock)

			ctx := context.Background()
			L.SetContext(ctx)

			err := L.DoString(fmt.Sprintf(`
				local resp = request.%s("https://api.example.com/test")
				assert(resp.status == 200, "expected 200")
			`, tc.luaFn))
			if err != nil {
				t.Fatalf("expected no error for request.%s, got: %v", tc.luaFn, err)
			}
			if mock.lastMethod != tc.httpMethod {
				t.Errorf("expected method %s, got %q", tc.httpMethod, mock.lastMethod)
			}
		})
	}

	t.Run("convenience method with opts", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "runtime")

		mock := &mockExecutor{
			result: map[string]any{"status": 201, "body": "created"},
		}
		RegisterRequestAPI(L, "test_plugin", mock)

		ctx := context.Background()
		L.SetContext(ctx)

		err := L.DoString(`
			local resp = request.post("https://api.example.com/items", {
				json = {name = "test"},
				timeout = 3
			})
			assert(resp.status == 201, "expected 201")
		`)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if mock.lastMethod != "POST" {
			t.Errorf("expected POST, got %q", mock.lastMethod)
		}
		if mock.lastOpts.Timeout != 3 {
			t.Errorf("expected timeout 3, got %d", mock.lastOpts.Timeout)
		}
	})

	t.Run("convenience method phase guard", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`request.get("https://api.example.com/test")`)
		if err == nil {
			t.Fatal("expected error for convenience method at module scope")
		}
		if !strings.Contains(err.Error(), "module scope") {
			t.Errorf("expected module scope error, got: %v", err)
		}
	})
}

func TestReadPendingRequests(t *testing.T) {
	t.Run("empty when no registrations", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		reqs := ReadPendingRequests(L)
		if len(reqs) != 0 {
			t.Fatalf("expected 0 pending requests, got %d", len(reqs))
		}
	})

	t.Run("returns all registrations", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})
		setVMPhase(L, "module_scope")
		RegisterRequestAPI(L, "test_plugin", nil)

		err := L.DoString(`
			request.register("api.example.com", {description = "API"})
			request.register("data.external.io")
		`)
		if err != nil {
			t.Fatalf("registration failed: %v", err)
		}

		reqs := ReadPendingRequests(L)
		if len(reqs) != 2 {
			t.Fatalf("expected 2 pending requests, got %d", len(reqs))
		}

		// Build a map for order-independent checking.
		byDomain := make(map[string]PendingRequest, len(reqs))
		for _, r := range reqs {
			byDomain[r.Domain] = r
		}

		r1, ok := byDomain["api.example.com"]
		if !ok {
			t.Fatal("expected api.example.com in pending requests")
		}
		if r1.Description != "API" {
			t.Errorf("expected description 'API', got %q", r1.Description)
		}

		r2, ok := byDomain["data.external.io"]
		if !ok {
			t.Fatal("expected data.external.io in pending requests")
		}
		if r2.Description != "" {
			t.Errorf("expected empty description, got %q", r2.Description)
		}
	})

	t.Run("returns nil when no __request_pending table", func(t *testing.T) {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		ApplySandbox(L, SandboxConfig{})

		reqs := ReadPendingRequests(L)
		if reqs != nil {
			t.Fatalf("expected nil, got %v", reqs)
		}
	})
}

func TestVMPhaseHelpers(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setVMPhase(L, "module_scope")
		if got := vmPhase(L); got != "module_scope" {
			t.Errorf("expected 'module_scope', got %q", got)
		}

		setVMPhase(L, "init")
		if got := vmPhase(L); got != "init" {
			t.Errorf("expected 'init', got %q", got)
		}

		setVMPhase(L, "runtime")
		if got := vmPhase(L); got != "runtime" {
			t.Errorf("expected 'runtime', got %q", got)
		}

		setVMPhase(L, "shutdown")
		if got := vmPhase(L); got != "shutdown" {
			t.Errorf("expected 'shutdown', got %q", got)
		}
	})

	t.Run("returns empty when not set", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		if got := vmPhase(L); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}
