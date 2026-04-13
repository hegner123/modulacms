package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// newTestClient creates a modula.Client backed by the given httptest.Server.
func newTestClient(t *testing.T, srv *httptest.Server) *modula.Client {
	t.Helper()
	client, err := modula.NewClient(modula.ClientConfig{
		BaseURL:    srv.URL,
		APIKey:     "test-key",
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
}

// newTestBackends creates SDK backends backed by the given httptest.Server.
func newTestBackends(t *testing.T, srv *httptest.Server) *Backends {
	t.Helper()
	return NewSDKBackends(newTestClient(t, srv))
}

// callTool invokes a tool handler directly, returning the result.
func callTool(t *testing.T, handler server.ToolHandlerFunc, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	result, err := handler(context.Background(), makeReq(args))
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	return result
}

// --- Health tool ---

func TestHandleHealth_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health" {
			t.Errorf("path = %q, want /api/v1/health", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"checks": map[string]bool{"database": true},
		})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleHealth(backends.Health)
	result := callTool(t, handler, nil)

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var health map[string]any
	if err := json.Unmarshal([]byte(text), &health); err != nil {
		t.Fatalf("failed to parse health response: %v", err)
	}
	if health["status"] != "ok" {
		t.Errorf("status = %v, want %q", health["status"], "ok")
	}
}

func TestHandleHealth_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"message":"database down"}`)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleHealth(backends.Health)
	result := callTool(t, handler, nil)

	if !result.IsError {
		t.Fatal("expected error result")
	}

	text := resultText(t, result)
	var detail map[string]any
	if err := json.Unmarshal([]byte(text), &detail); err != nil {
		t.Fatalf("failed to parse error JSON: %v", err)
	}
	if detail["status"] != float64(500) {
		t.Errorf("status = %v, want 500", detail["status"])
	}
	if detail["message"] != "database down" {
		t.Errorf("message = %v, want %q", detail["message"], "database down")
	}
}

// --- Route tools ---

func TestHandleListRoutes(t *testing.T) {
	routes := []modula.Route{
		{RouteID: "rt-001", Slug: "about", Title: "About Us", Status: 1},
		{RouteID: "rt-002", Slug: "blog", Title: "Blog", Status: 1},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/routes" {
			t.Errorf("path = %q, want /api/v1/routes", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(routes)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleListRoutes(backends.Routes)
	result := callTool(t, handler, nil)

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var decoded []modula.Route
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("len = %d, want 2", len(decoded))
	}
	if decoded[0].Slug != "about" {
		t.Errorf("decoded[0].Slug = %q, want %q", decoded[0].Slug, "about")
	}
}

func TestHandleGetRoute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		q := r.URL.Query().Get("q")
		if q != "rt-001" {
			t.Errorf("query q = %q, want %q", q, "rt-001")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modula.Route{
			RouteID: "rt-001", Slug: "about", Title: "About Us", Status: 1,
		})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleGetRoute(backends.Routes)
	result := callTool(t, handler, map[string]any{"id": "rt-001"})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var route modula.Route
	if err := json.Unmarshal([]byte(text), &route); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if route.RouteID != "rt-001" {
		t.Errorf("RouteID = %q, want %q", route.RouteID, "rt-001")
	}
}

func TestHandleGetRoute_MissingID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called when id is missing")
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleGetRoute(backends.Routes)
	result := callTool(t, handler, map[string]any{})

	if !result.IsError {
		t.Fatal("expected error result for missing id")
	}
	text := resultText(t, result)
	if text != "id is required" {
		t.Errorf("text = %q, want %q", text, "id is required")
	}
}

func TestHandleCreateRoute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var params modula.CreateRouteParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Slug != "contact" {
			t.Errorf("slug = %q, want %q", params.Slug, "contact")
		}
		if params.Title != "Contact" {
			t.Errorf("title = %q, want %q", params.Title, "Contact")
		}
		if params.Status != 1 {
			t.Errorf("status = %d, want 1", params.Status)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(modula.Route{
			RouteID: "rt-new", Slug: params.Slug, Title: params.Title, Status: params.Status,
		})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleCreateRoute(backends.Routes)
	result := callTool(t, handler, map[string]any{
		"slug":   "contact",
		"title":  "Contact",
		"status": float64(1),
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var route modula.Route
	if err := json.Unmarshal([]byte(text), &route); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if route.RouteID != "rt-new" {
		t.Errorf("RouteID = %q, want %q", route.RouteID, "rt-new")
	}
	if route.Slug != "contact" {
		t.Errorf("Slug = %q, want %q", route.Slug, "contact")
	}
}

func TestHandleCreateRoute_WithOptionalAuthorID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params modula.CreateRouteParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.AuthorID == nil {
			t.Error("expected author_id to be set")
		} else if *params.AuthorID != "usr-001" {
			t.Errorf("author_id = %q, want %q", *params.AuthorID, "usr-001")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(modula.Route{
			RouteID:  "rt-new",
			Slug:     params.Slug,
			Title:    params.Title,
			Status:   params.Status,
			AuthorID: params.AuthorID,
		})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleCreateRoute(backends.Routes)
	result := callTool(t, handler, map[string]any{
		"slug":      "faq",
		"title":     "FAQ",
		"status":    float64(1),
		"author_id": "usr-001",
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}

func TestHandleDeleteRoute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		q := r.URL.Query().Get("q")
		if q != "rt-del" {
			t.Errorf("query q = %q, want %q", q, "rt-del")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleDeleteRoute(backends.Routes)
	result := callTool(t, handler, map[string]any{"id": "rt-del"})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
	text := resultText(t, result)
	if text != "deleted" {
		t.Errorf("text = %q, want %q", text, "deleted")
	}
}

func TestHandleDeleteRoute_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"message":"route not found"}`)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleDeleteRoute(backends.Routes)
	result := callTool(t, handler, map[string]any{"id": "rt-nonexistent"})

	if !result.IsError {
		t.Fatal("expected error result for 404")
	}

	text := resultText(t, result)
	var detail map[string]any
	if err := json.Unmarshal([]byte(text), &detail); err != nil {
		t.Fatalf("failed to parse error JSON: %v", err)
	}
	if detail["status"] != float64(404) {
		t.Errorf("status = %v, want 404", detail["status"])
	}
}

func TestHandleUpdateRoute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		var params modula.UpdateRouteParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Slug != "about" {
			t.Errorf("slug = %q, want %q", params.Slug, "about")
		}
		if params.Title != "About Updated" {
			t.Errorf("title = %q, want %q", params.Title, "About Updated")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modula.Route{
			RouteID: "rt-001", Slug: params.Slug, Title: params.Title, Status: params.Status,
		})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleUpdateRoute(backends.Routes)
	result := callTool(t, handler, map[string]any{
		"id":     "rt-001",
		"slug":   "about",
		"title":  "About Updated",
		"status": float64(1),
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}

// --- Config tools ---

func TestHandleGetConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"config": map[string]any{
				"port":     8080,
				"ssl_port": 8443,
			},
		})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleGetConfig(backends.Config)
	result := callTool(t, handler, nil)

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var config map[string]any
	if err := json.Unmarshal([]byte(text), &config); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	cfgMap, ok := config["config"].(map[string]any)
	if !ok {
		t.Fatal("config field missing or not an object")
	}
	if cfgMap["port"] != float64(8080) {
		t.Errorf("port = %v, want 8080", cfgMap["port"])
	}
}

func TestHandleGetConfigMeta(t *testing.T) {
	meta := map[string]any{
		"fields": []map[string]any{
			{"json_key": "port", "label": "Port", "category": "server", "required": true},
		},
		"categories": []map[string]any{
			{"key": "server", "label": "Server"},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(meta)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleGetConfigMeta(backends.Config)
	result := callTool(t, handler, nil)

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}

func TestHandleUpdateConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":     true,
			"config": map[string]any{"port": 9090},
		})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleUpdateConfig(backends.Config)
	result := callTool(t, handler, map[string]any{
		"updates": map[string]any{"port": float64(9090)},
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}

func TestHandleUpdateConfig_InvalidUpdates(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called with invalid updates param")
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleUpdateConfig(backends.Config)
	result := callTool(t, handler, map[string]any{
		"updates": "not-an-object",
	})

	if !result.IsError {
		t.Fatal("expected error result for invalid updates")
	}
	text := resultText(t, result)
	if text != "updates must be a JSON object of key-value pairs" {
		t.Errorf("text = %q", text)
	}
}

// --- Session tools ---

func TestHandleListSessions(t *testing.T) {
	sessions := []modula.Session{
		{SessionID: "sess-001"},
		{SessionID: "sess-002"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessions)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleListSessions(backends.Sessions)
	result := callTool(t, handler, nil)

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var decoded []modula.Session
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("len = %d, want 2", len(decoded))
	}
}

func TestHandleGetSession(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q != "sess-001" {
			t.Errorf("query q = %q, want %q", q, "sess-001")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modula.Session{SessionID: "sess-001"})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleGetSession(backends.Sessions)
	result := callTool(t, handler, map[string]any{"id": "sess-001"})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}

func TestHandleUpdateSession(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		var params modula.UpdateSessionParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.SessionID != "sess-001" {
			t.Errorf("SessionID = %q, want %q", params.SessionID, "sess-001")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modula.Session{SessionID: "sess-001"})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleUpdateSession(backends.Sessions)
	result := callTool(t, handler, map[string]any{
		"id":         "sess-001",
		"ip_address": "192.168.1.1",
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}

func TestHandleDeleteSession(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleDeleteSession(backends.Sessions)
	result := callTool(t, handler, map[string]any{"id": "sess-001"})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
	text := resultText(t, result)
	if text != "deleted" {
		t.Errorf("text = %q, want %q", text, "deleted")
	}
}

// --- Schema tools (datatypes) ---

func TestHandleListDatatypes(t *testing.T) {
	datatypes := []modula.Datatype{
		{DatatypeID: "dt-001", Label: "Page", Type: "page"},
		{DatatypeID: "dt-002", Label: "Post", Type: "post"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(datatypes)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleListDatatypes(backends.Schema)
	result := callTool(t, handler, map[string]any{})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var decoded []modula.Datatype
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("len = %d, want 2", len(decoded))
	}
}

func TestHandleGetDatatype(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q != "dt-001" {
			t.Errorf("query q = %q, want %q", q, "dt-001")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modula.Datatype{
			DatatypeID: "dt-001", Label: "Page", Type: "page",
		})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleGetDatatype(backends.Schema)
	result := callTool(t, handler, map[string]any{"id": "dt-001"})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var dt modula.Datatype
	if err := json.Unmarshal([]byte(text), &dt); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if dt.Label != "Page" {
		t.Errorf("Label = %q, want %q", dt.Label, "Page")
	}
}

func TestHandleCreateDatatype(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var params modula.CreateDatatypeParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Label != "Event" {
			t.Errorf("Label = %q, want %q", params.Label, "Event")
		}
		if params.Type != "component" {
			t.Errorf("Type = %q, want %q", params.Type, "component")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(modula.Datatype{
			DatatypeID: "dt-new", Label: params.Label, Type: params.Type,
		})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleCreateDatatype(backends.Schema)
	result := callTool(t, handler, map[string]any{
		"label": "Event",
		"type":  "component",
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var dt modula.Datatype
	if err := json.Unmarshal([]byte(text), &dt); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if dt.DatatypeID != "dt-new" {
		t.Errorf("DatatypeID = %q, want %q", dt.DatatypeID, "dt-new")
	}
}

func TestHandleDeleteDatatype(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleDeleteDatatype(backends.Schema)
	result := callTool(t, handler, map[string]any{"id": "dt-001"})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}

// --- Tool registration ---

func TestToolRegistration_AllGroupsRegistered(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	srv := server.NewMCPServer("modulacms-test", "0.0.1")

	// All register functions should not panic
	registerContentTools(srv, backends.Content)
	registerSchemaTools(srv, backends.Schema)
	registerMediaTools(srv, backends.Media, backends.MediaFolders)
	registerMediaFolderTools(srv, backends.MediaFolders)
	registerRouteTools(srv, backends.Routes)
	registerUserTools(srv, backends.Users)
	registerRBACTools(srv, backends.RBAC)
	registerConfigTools(srv, backends.Config)
	registerImportTools(srv, backends.Import)
	registerDeployTools(srv, backends.Deploy)
	registerHealthTools(srv, backends.Health)
	registerSessionTools(srv, backends.Sessions)
	registerTokenTools(srv, backends.Tokens)
	registerSSHKeyTools(srv, backends.SSHKeys)
	registerOAuthTools(srv, backends.OAuth)
	registerTableTools(srv, backends.Tables)
	registerPluginTools(srv, backends.Plugins)
	registerAdminContentTools(srv, backends.AdminContent)
	registerAdminSchemaTools(srv, backends.AdminSchema)
	registerAdminRouteTools(srv, backends.AdminRoutes)
	registerAdminMediaTools(srv, backends.AdminMedia, backends.AdminMediaFolders)
	registerAdminMediaFolderTools(srv, backends.AdminMediaFolders)
	registerPublishingTools(srv, backends.Publishing)
	registerVersionTools(srv, backends.Versions)
	registerWebhookTools(srv, backends.Webhooks)
	registerLocaleTools(srv, backends.Locales)
	registerValidationTools(srv, backends.Validations)
	registerSearchTools(srv, backends.Search)
	registerActivityTools(srv, backends.Activity)
	registerAuthTools(srv, backends.Auth)
}

// --- A2: Partial-update semantics ---
// These tests verify that omitted optional fields produce null (not "") in the
// JSON params the handler passes to the backend. The recording backends capture
// the raw json.RawMessage before the SDK typed structs normalize null -> "".

// recordingUserBackend captures the raw params from UpdateUser.
type recordingUserBackend struct {
	UserBackend
	lastParams json.RawMessage
}

func (r *recordingUserBackend) UpdateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	r.lastParams = params
	return json.RawMessage(`{"user_id":"usr-001"}`), nil
}

// recordingSessionBackend captures the raw params from UpdateSession.
type recordingSessionBackend struct {
	SessionBackend
	lastParams json.RawMessage
}

func (r *recordingSessionBackend) UpdateSession(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	r.lastParams = params
	return json.RawMessage(`{"session_id":"sess-001"}`), nil
}

// recordingSchemaBackend captures the raw params from UpdateDatatype/UpdateField.
type recordingSchemaBackend struct {
	SchemaBackend
	lastParams json.RawMessage
}

func (r *recordingSchemaBackend) UpdateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	r.lastParams = params
	return json.RawMessage(`{"datatype_id":"dt-001"}`), nil
}

func (r *recordingSchemaBackend) UpdateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	r.lastParams = params
	return json.RawMessage(`{"field_id":"fld-001"}`), nil
}

func (r *recordingSchemaBackend) GetDatatypeFull(ctx context.Context, id string) (json.RawMessage, error) {
	if id == "" {
		return json.RawMessage(`[{"datatype_id":"dt-001","fields":[]}]`), nil
	}
	return json.RawMessage(`{"datatype_id":"` + id + `","fields":[]}`), nil
}

// recordingAdminSchemaBackend captures the raw params from admin UpdateDatatype/UpdateField.
type recordingAdminSchemaBackend struct {
	AdminSchemaBackend
	lastParams json.RawMessage
}

func (r *recordingAdminSchemaBackend) UpdateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	r.lastParams = params
	return json.RawMessage(`{"admin_datatype_id":"adt-001"}`), nil
}

func (r *recordingAdminSchemaBackend) UpdateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	r.lastParams = params
	return json.RawMessage(`{"admin_field_id":"afld-001"}`), nil
}

// assertNullFields checks that each named field is JSON null in the raw params.
func assertNullFields(t *testing.T, raw json.RawMessage, fields []string) {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("failed to unmarshal params: %v", err)
	}
	for _, field := range fields {
		v, ok := m[field]
		if !ok {
			continue // absent is acceptable (even better than null)
		}
		if string(v) != "null" {
			t.Errorf("%s = %s, want null (field was omitted, should not overwrite existing data)", field, string(v))
		}
	}
}

func TestHandleUpdateUser_OmittedFieldsAreNull(t *testing.T) {
	rec := &recordingUserBackend{}
	handler := handleUpdateUser(rec)

	result := callTool(t, handler, map[string]any{
		"id":       "usr-001",
		"username": "newname",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	// Verify provided field is present
	var m map[string]json.RawMessage
	if err := json.Unmarshal(rec.lastParams, &m); err != nil {
		t.Fatalf("failed to unmarshal params: %v", err)
	}
	if string(m["username"]) != `"newname"` {
		t.Errorf("username = %s, want %q", string(m["username"]), "newname")
	}

	// Omitted fields must be null, NOT empty string ""
	assertNullFields(t, rec.lastParams, []string{"name", "email", "password", "role"})
}

func TestHandleUpdateSession_OmittedExpiresAtIsNull(t *testing.T) {
	rec := &recordingSessionBackend{}
	handler := handleUpdateSession(rec)

	result := callTool(t, handler, map[string]any{
		"id":         "sess-001",
		"ip_address": "10.0.0.1",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	assertNullFields(t, rec.lastParams, []string{"expires_at"})
}

func TestHandleUpdateDatatype_OmittedNameIsNull(t *testing.T) {
	rec := &recordingSchemaBackend{}
	handler := handleUpdateDatatype(rec)

	result := callTool(t, handler, map[string]any{
		"id":    "dt-001",
		"label": "Blog Post",
		"type":  "collection",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	assertNullFields(t, rec.lastParams, []string{"name"})
}

func TestHandleAdminUpdateDatatype_OmittedNameIsNull(t *testing.T) {
	rec := &recordingAdminSchemaBackend{}
	handler := handleAdminUpdateDatatype(rec)

	result := callTool(t, handler, map[string]any{
		"id":    "adt-001",
		"label": "Sidebar Widget",
		"type":  "single",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	assertNullFields(t, rec.lastParams, []string{"name"})
}

func TestHandleUpdateField_OmittedFieldsAreNull(t *testing.T) {
	rec := &recordingSchemaBackend{}
	handler := handleUpdateField(rec)

	result := callTool(t, handler, map[string]any{
		"id":         "fld-001",
		"label":      "Title",
		"field_type": "text",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	assertNullFields(t, rec.lastParams, []string{"name", "data", "validation", "ui_config"})
}

func TestHandleAdminUpdateField_OmittedFieldsAreNull(t *testing.T) {
	rec := &recordingAdminSchemaBackend{}
	handler := handleAdminUpdateField(rec)

	result := callTool(t, handler, map[string]any{
		"id":         "afld-001",
		"label":      "Heading",
		"field_type": "text",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	assertNullFields(t, rec.lastParams, []string{"name", "data", "validation", "ui_config"})
}

// --- A3: media_cleanup_apply confirm gate ---

// recordingMediaBackend captures calls to MediaCleanup and MediaCleanupCheck.
type recordingMediaBackend struct {
	MediaBackend
	cleanupCalled bool
	checkCalled   bool
}

func (r *recordingMediaBackend) MediaCleanup(ctx context.Context) (json.RawMessage, error) {
	r.cleanupCalled = true
	return json.RawMessage(`{"deleted":3}`), nil
}

func (r *recordingMediaBackend) MediaCleanupCheck(ctx context.Context) (json.RawMessage, error) {
	r.checkCalled = true
	return json.RawMessage(`{"total_objects":10,"tracked_keys":7,"orphaned_keys":["a","b","c"]}`), nil
}

func TestHandleMediaCleanupApply_ConfirmTrue(t *testing.T) {
	rec := &recordingMediaBackend{}
	handler := handleMediaCleanupApply(rec)

	result := callTool(t, handler, map[string]any{"confirm": true})
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
	if !rec.cleanupCalled {
		t.Error("MediaCleanup was not called when confirm=true")
	}
}

func TestHandleMediaCleanupApply_ConfirmFalse(t *testing.T) {
	rec := &recordingMediaBackend{}
	handler := handleMediaCleanupApply(rec)

	result := callTool(t, handler, map[string]any{"confirm": false})
	if !result.IsError {
		t.Fatal("expected error when confirm=false")
	}
	if rec.cleanupCalled {
		t.Error("MediaCleanup was called when confirm=false, should have been rejected")
	}
}

func TestHandleMediaCleanupApply_ConfirmOmitted(t *testing.T) {
	rec := &recordingMediaBackend{}
	handler := handleMediaCleanupApply(rec)

	result := callTool(t, handler, map[string]any{})
	if !result.IsError {
		t.Fatal("expected error when confirm is omitted")
	}
	if rec.cleanupCalled {
		t.Error("MediaCleanup was called when confirm was omitted, should have been rejected")
	}
}

func TestHandleMediaCleanupCheck_ReturnsOrphans(t *testing.T) {
	rec := &recordingMediaBackend{}
	handler := handleMediaCleanupCheck(rec)

	result := callTool(t, handler, nil)
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
	if !rec.checkCalled {
		t.Error("MediaCleanupCheck was not called")
	}
	text := resultText(t, result)
	var body map[string]any
	if err := json.Unmarshal([]byte(text), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	keys, ok := body["orphaned_keys"].([]any)
	if !ok {
		t.Fatal("orphaned_keys missing or not an array")
	}
	if len(keys) != 3 {
		t.Errorf("orphaned_keys length = %d, want 3", len(keys))
	}
}

// --- A7: get_datatype_full requires ID, list_datatypes_full does not ---

func TestHandleGetDatatypeFull_MissingID(t *testing.T) {
	rec := &recordingSchemaBackend{}
	handler := handleGetDatatypeFull(rec)

	result := callTool(t, handler, map[string]any{})
	if !result.IsError {
		t.Fatal("expected error when id is omitted from get_datatype_full")
	}
	text := resultText(t, result)
	if text != "id is required" {
		t.Errorf("error text = %q, want %q", text, "id is required")
	}
}

func TestHandleListDatatypesFull_NoIDNeeded(t *testing.T) {
	rec := &recordingSchemaBackend{}
	// Override GetDatatypeFull to capture the empty-string call
	rec.lastParams = nil
	handler := handleListDatatypesFull(rec)

	result := callTool(t, handler, map[string]any{})
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}

// --- Adversarial: wrong parameter types and empty required IDs ---

func TestRequireString_RejectsNumericID(t *testing.T) {
	// LLMs sometimes send {"id": 123} instead of {"id": "123"}.
	// RequireString must reject this.
	handler := handleGetRoute(nil) // backend not called, error before that
	result := callTool(t, handler, map[string]any{"id": 123})
	if !result.IsError {
		t.Fatal("expected error when id is a number, not a string")
	}
}

func TestRequireString_RejectsMissingID(t *testing.T) {
	handler := handleGetRoute(nil)
	result := callTool(t, handler, map[string]any{})
	if !result.IsError {
		t.Fatal("expected error when id is missing")
	}
}

func TestRequireString_AcceptsEmptyString(t *testing.T) {
	// Empty string passes RequireString (no error). This is a known limitation.
	// Handlers that need non-empty IDs should validate separately.
	// We verify this by calling RequireString directly.
	req := makeReq(map[string]any{"id": ""})
	val, err := req.RequireString("id")
	if err != nil {
		t.Fatalf("RequireString rejected empty string: %v", err)
	}
	if val != "" {
		t.Errorf("val = %q, want empty string", val)
	}
	// This documents: empty string IDs pass RequireString and reach the backend.
	// A future improvement would add non-empty validation at the handler level.
}

func TestGetBool_CoercesStringTrue(t *testing.T) {
	// LLMs sometimes send {"confirm": "true"} (string) instead of {"confirm": true} (bool).
	// GetBool coerces string "true" to true. Document this behavior: the confirm gate
	// accepts string "true", which is acceptable since the intent is unambiguous.
	rec := &recordingMediaBackend{}
	handler := handleMediaCleanupApply(rec)
	result := callTool(t, handler, map[string]any{"confirm": "true"})
	if result.IsError {
		t.Fatalf("unexpected error: GetBool should coerce string 'true' to true: %s", resultText(t, result))
	}
	if !rec.cleanupCalled {
		t.Error("MediaCleanup should have been called since GetBool coerces string 'true'")
	}
}

func TestGetBool_CoercesNumericOne(t *testing.T) {
	// LLMs might send {"confirm": 1} instead of {"confirm": true}.
	// GetBool coerces float64(1) to true (nonzero = true).
	// Document this behavior: numeric 1 passes the confirm gate.
	rec := &recordingMediaBackend{}
	handler := handleMediaCleanupApply(rec)
	result := callTool(t, handler, map[string]any{"confirm": float64(1)})
	if result.IsError {
		t.Fatalf("unexpected error: GetBool should coerce numeric 1 to true: %s", resultText(t, result))
	}
	if !rec.cleanupCalled {
		t.Error("MediaCleanup should have been called since GetBool coerces numeric 1")
	}
}

func TestGetBool_RejectsMapValue(t *testing.T) {
	// Non-coercible types (map, array, nil) should fall through to default (false).
	rec := &recordingMediaBackend{}
	handler := handleMediaCleanupApply(rec)
	result := callTool(t, handler, map[string]any{"confirm": map[string]any{"yes": true}})
	if !result.IsError {
		t.Fatal("expected error: map value should not pass boolean confirm gate")
	}
	if rec.cleanupCalled {
		t.Error("MediaCleanup was called with non-boolean confirm value")
	}
}

// --- Auth header propagation ---

func TestAuthHeaderPropagated(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-key")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)
	handler := handleHealth(backends.Health)
	result := callTool(t, handler, nil)

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}
