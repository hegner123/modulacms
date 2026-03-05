package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

// newTestClient creates a modulacms.Client backed by the given httptest.Server.
func newTestClient(t *testing.T, srv *httptest.Server) *modulacms.Client {
	t.Helper()
	client, err := modulacms.NewClient(modulacms.ClientConfig{
		BaseURL:    srv.URL,
		APIKey:     "test-key",
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
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

	client := newTestClient(t, ts)
	handler := handleHealth(client)
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

	client := newTestClient(t, ts)
	handler := handleHealth(client)
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
	routes := []modulacms.Route{
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

	client := newTestClient(t, ts)
	handler := handleListRoutes(client)
	result := callTool(t, handler, nil)

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var decoded []modulacms.Route
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
		json.NewEncoder(w).Encode(modulacms.Route{
			RouteID: "rt-001", Slug: "about", Title: "About Us", Status: 1,
		})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleGetRoute(client)
	result := callTool(t, handler, map[string]any{"id": "rt-001"})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var route modulacms.Route
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

	client := newTestClient(t, ts)
	handler := handleGetRoute(client)
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
		var params modulacms.CreateRouteParams
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
		json.NewEncoder(w).Encode(modulacms.Route{
			RouteID: "rt-new", Slug: params.Slug, Title: params.Title, Status: params.Status,
		})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleCreateRoute(client)
	result := callTool(t, handler, map[string]any{
		"slug":   "contact",
		"title":  "Contact",
		"status": float64(1),
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var route modulacms.Route
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
		var params modulacms.CreateRouteParams
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
		json.NewEncoder(w).Encode(modulacms.Route{
			RouteID:  "rt-new",
			Slug:     params.Slug,
			Title:    params.Title,
			Status:   params.Status,
			AuthorID: params.AuthorID,
		})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleCreateRoute(client)
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

	client := newTestClient(t, ts)
	handler := handleDeleteRoute(client)
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

	client := newTestClient(t, ts)
	handler := handleDeleteRoute(client)
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
		var params modulacms.UpdateRouteParams
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
		json.NewEncoder(w).Encode(modulacms.Route{
			RouteID: "rt-001", Slug: params.Slug, Title: params.Title, Status: params.Status,
		})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleUpdateRoute(client)
	result := callTool(t, handler, map[string]any{
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

	client := newTestClient(t, ts)
	handler := handleGetConfig(client)
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
		"categories": []string{"server"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(meta)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleGetConfigMeta(client)
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

	client := newTestClient(t, ts)
	handler := handleUpdateConfig(client)
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

	client := newTestClient(t, ts)
	handler := handleUpdateConfig(client)
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
	sessions := []modulacms.Session{
		{SessionID: "sess-001"},
		{SessionID: "sess-002"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessions)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleListSessions(client)
	result := callTool(t, handler, nil)

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var decoded []modulacms.Session
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
		json.NewEncoder(w).Encode(modulacms.Session{SessionID: "sess-001"})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleGetSession(client)
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
		var params modulacms.UpdateSessionParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.SessionID != "sess-001" {
			t.Errorf("SessionID = %q, want %q", params.SessionID, "sess-001")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modulacms.Session{SessionID: "sess-001"})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleUpdateSession(client)
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

	client := newTestClient(t, ts)
	handler := handleDeleteSession(client)
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
	datatypes := []modulacms.Datatype{
		{DatatypeID: "dt-001", Label: "Page", Type: "page"},
		{DatatypeID: "dt-002", Label: "Post", Type: "post"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(datatypes)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleListDatatypes(client)
	result := callTool(t, handler, map[string]any{})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var decoded []modulacms.Datatype
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
		json.NewEncoder(w).Encode(modulacms.Datatype{
			DatatypeID: "dt-001", Label: "Page", Type: "page",
		})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleGetDatatype(client)
	result := callTool(t, handler, map[string]any{"id": "dt-001"})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var dt modulacms.Datatype
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
		var params modulacms.CreateDatatypeParams
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
		json.NewEncoder(w).Encode(modulacms.Datatype{
			DatatypeID: "dt-new", Label: params.Label, Type: params.Type,
		})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	handler := handleCreateDatatype(client)
	result := callTool(t, handler, map[string]any{
		"label": "Event",
		"type":  "component",
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	var dt modulacms.Datatype
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

	client := newTestClient(t, ts)
	handler := handleDeleteDatatype(client)
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

	client := newTestClient(t, ts)
	srv := server.NewMCPServer("modulacms-test", "0.0.1")

	// All register functions should not panic
	registerContentTools(srv, client)
	registerSchemaTools(srv, client)
	registerMediaTools(srv, client)
	registerRouteTools(srv, client)
	registerUserTools(srv, client)
	registerRBACTools(srv, client)
	registerConfigTools(srv, client)
	registerImportTools(srv, client)
	registerDeployTools(srv, client)
	registerHealthTools(srv, client)
	registerSessionTools(srv, client)
	registerTokenTools(srv, client)
	registerSSHKeyTools(srv, client)
	registerOAuthTools(srv, client)
	registerTableTools(srv, client)
	registerPluginTools(srv, client)
	registerAdminContentTools(srv, client)
	registerAdminSchemaTools(srv, client)
	registerAdminRouteTools(srv, client)
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

	client := newTestClient(t, ts)
	handler := handleHealth(client)
	result := callTool(t, handler, nil)

	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}
}
