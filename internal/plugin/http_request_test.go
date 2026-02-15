package plugin

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// --- writePluginError tests ---

func TestWritePluginError(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		code      string
		message   string
		requestID string
	}{
		{
			name:      "400 INVALID_REQUEST with request ID",
			status:    http.StatusBadRequest,
			code:      "INVALID_REQUEST",
			message:   "body too large",
			requestID: "req-abc-123",
		},
		{
			name:      "404 ROUTE_NOT_FOUND without request ID",
			status:    http.StatusNotFound,
			code:      "ROUTE_NOT_FOUND",
			message:   "no matching route",
			requestID: "",
		},
		{
			name:      "500 HANDLER_ERROR",
			status:    http.StatusInternalServerError,
			code:      "HANDLER_ERROR",
			message:   "internal plugin error",
			requestID: "req-def-456",
		},
		{
			name:      "503 PLUGIN_UNAVAILABLE",
			status:    http.StatusServiceUnavailable,
			code:      "PLUGIN_UNAVAILABLE",
			message:   "plugin not running",
			requestID: "req-ghi-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writePluginError(rec, tt.status, tt.code, tt.message, tt.requestID)

			// Verify status code.
			if rec.Code != tt.status {
				t.Fatalf("expected status %d, got %d", tt.status, rec.Code)
			}

			// Verify Content-Type.
			ct := rec.Header().Get("Content-Type")
			if ct != "application/json" {
				t.Fatalf("expected Content-Type application/json, got %q", ct)
			}

			// Verify JSON structure matches PluginErrorResponse schema.
			var resp PluginErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if resp.Error.Code != tt.code {
				t.Fatalf("expected error code %q, got %q", tt.code, resp.Error.Code)
			}
			if resp.Error.Message != tt.message {
				t.Fatalf("expected error message %q, got %q", tt.message, resp.Error.Message)
			}
			if resp.Error.RequestID != tt.requestID {
				t.Fatalf("expected request_id %q, got %q", tt.requestID, resp.Error.RequestID)
			}
		})
	}
}

func TestWritePluginError_OmitsEmptyRequestID(t *testing.T) {
	rec := httptest.NewRecorder()
	writePluginError(rec, 404, "ROUTE_NOT_FOUND", "not found", "")

	// When requestID is empty, the JSON should omit the field entirely
	// (due to the omitempty tag).
	body := rec.Body.String()
	if strings.Contains(body, "request_id") {
		t.Fatalf("expected request_id to be omitted when empty, got body: %s", body)
	}
}

// --- BuildLuaRequest tests ---

func TestBuildLuaRequest_GETWithQueryParams(t *testing.T) {
	L := newTestState()
	defer L.Close()

	r := httptest.NewRequest("GET", "/tasks?page=2&limit=10", nil)
	tbl, err := BuildLuaRequest(L, r, "192.168.1.100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// method
	method := L.GetField(tbl, "method")
	if method.String() != "GET" {
		t.Fatalf("expected method=GET, got %v", method)
	}

	// path
	path := L.GetField(tbl, "path")
	if path.String() != "/tasks" {
		t.Fatalf("expected path=/tasks, got %v", path)
	}

	// client_ip comes from the parameter, not from the request
	clientIP := L.GetField(tbl, "client_ip")
	if clientIP.String() != "192.168.1.100" {
		t.Fatalf("expected client_ip=192.168.1.100, got %v", clientIP)
	}

	// query params
	queryTbl := L.GetField(tbl, "query")
	qt, ok := queryTbl.(*lua.LTable)
	if !ok {
		t.Fatalf("expected query to be *LTable, got %T", queryTbl)
	}
	page := L.GetField(qt, "page")
	if page.String() != "2" {
		t.Fatalf("expected query.page=2, got %v", page)
	}
	limit := L.GetField(qt, "limit")
	if limit.String() != "10" {
		t.Fatalf("expected query.limit=10, got %v", limit)
	}

	// body should be empty for GET
	body := L.GetField(tbl, "body")
	if body.String() != "" {
		t.Fatalf("expected empty body for GET, got %q", body.String())
	}

	// json should be nil for non-JSON request
	jsonVal := L.GetField(tbl, "json")
	if jsonVal != lua.LNil {
		t.Fatalf("expected json=nil for GET request, got %T(%v)", jsonVal, jsonVal)
	}
}

func TestBuildLuaRequest_POSTWithJSONBody(t *testing.T) {
	L := newTestState()
	defer L.Close()

	bodyContent := `{"title":"Buy groceries","priority":1}`
	r := httptest.NewRequest("POST", "/tasks", strings.NewReader(bodyContent))
	r.Header.Set("Content-Type", "application/json")

	tbl, err := BuildLuaRequest(L, r, "10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// body should have the raw JSON string
	body := L.GetField(tbl, "body")
	if body.String() != bodyContent {
		t.Fatalf("expected body=%q, got %q", bodyContent, body.String())
	}

	// json should be populated
	jsonVal := L.GetField(tbl, "json")
	if jsonVal == lua.LNil {
		t.Fatal("expected json to be populated for application/json Content-Type")
	}

	jsonTbl, ok := jsonVal.(*lua.LTable)
	if !ok {
		t.Fatalf("expected json to be *LTable, got %T", jsonVal)
	}

	title := L.GetField(jsonTbl, "title")
	if title.String() != "Buy groceries" {
		t.Fatalf("expected json.title=Buy groceries, got %v", title)
	}

	priority := L.GetField(jsonTbl, "priority")
	if num, ok := priority.(lua.LNumber); !ok || float64(num) != 1 {
		t.Fatalf("expected json.priority=1, got %v", priority)
	}
}

func TestBuildLuaRequest_POSTWithJSONContentTypeAndCharset(t *testing.T) {
	L := newTestState()
	defer L.Close()

	bodyContent := `{"key":"value"}`
	r := httptest.NewRequest("POST", "/data", strings.NewReader(bodyContent))
	r.Header.Set("Content-Type", "application/json; charset=utf-8")

	tbl, err := BuildLuaRequest(L, r, "10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// json should be populated even with charset parameter
	jsonVal := L.GetField(tbl, "json")
	if jsonVal == lua.LNil {
		t.Fatal("expected json to be populated for application/json with charset")
	}
}

func TestBuildLuaRequest_POSTWithFormBody(t *testing.T) {
	L := newTestState()
	defer L.Close()

	bodyContent := "username=alice&password=secret"
	r := httptest.NewRequest("POST", "/login", strings.NewReader(bodyContent))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tbl, err := BuildLuaRequest(L, r, "10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// body should have the raw form content
	body := L.GetField(tbl, "body")
	if body.String() != bodyContent {
		t.Fatalf("expected body=%q, got %q", bodyContent, body.String())
	}

	// json should be nil for non-JSON Content-Type
	jsonVal := L.GetField(tbl, "json")
	if jsonVal != lua.LNil {
		t.Fatalf("expected json=nil for form Content-Type, got %T(%v)", jsonVal, jsonVal)
	}
}

func TestBuildLuaRequest_HeadersNormalizedToLowercase(t *testing.T) {
	L := newTestState()
	defer L.Close()

	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Custom-Header", "custom-value")
	r.Header.Set("Authorization", "Bearer token123")

	tbl, err := BuildLuaRequest(L, r, "10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	headersTbl := L.GetField(tbl, "headers")
	ht, ok := headersTbl.(*lua.LTable)
	if !ok {
		t.Fatalf("expected headers to be *LTable, got %T", headersTbl)
	}

	// Headers should be accessible via lowercase keys
	customVal := L.GetField(ht, "x-custom-header")
	if customVal.String() != "custom-value" {
		t.Fatalf("expected x-custom-header=custom-value, got %v", customVal)
	}

	authVal := L.GetField(ht, "authorization")
	if authVal.String() != "Bearer token123" {
		t.Fatalf("expected authorization=Bearer token123, got %v", authVal)
	}
}

func TestBuildLuaRequest_PathParamsFromPattern(t *testing.T) {
	L := newTestState()
	defer L.Close()

	// Simulate a request that was matched by ServeMux with a pattern.
	// We need to use a real ServeMux to get r.Pattern and r.PathValue set.
	mux := http.NewServeMux()
	var capturedReq *http.Request
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/items/abc-123", nil)
	mux.ServeHTTP(rec, req)

	if capturedReq == nil {
		t.Fatal("handler was not called by ServeMux")
	}

	tbl, err := BuildLuaRequest(L, capturedReq, "10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paramsTbl := L.GetField(tbl, "params")
	pt, ok := paramsTbl.(*lua.LTable)
	if !ok {
		t.Fatalf("expected params to be *LTable, got %T", paramsTbl)
	}

	idVal := L.GetField(pt, "id")
	if idVal.String() != "abc-123" {
		t.Fatalf("expected params.id=abc-123, got %v", idVal)
	}
}

func TestBuildLuaRequest_ClientIPFromParameter(t *testing.T) {
	L := newTestState()
	defer L.Close()

	// The request's RemoteAddr is different from the clientIP parameter.
	// BuildLuaRequest should use the parameter, not RemoteAddr.
	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "192.168.0.1:54321"

	tbl, err := BuildLuaRequest(L, r, "203.0.113.50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	clientIP := L.GetField(tbl, "client_ip")
	if clientIP.String() != "203.0.113.50" {
		t.Fatalf("expected client_ip=203.0.113.50 (from parameter), got %v", clientIP)
	}
}

func TestBuildLuaRequest_MalformedJSON(t *testing.T) {
	L := newTestState()
	defer L.Close()

	r := httptest.NewRequest("POST", "/data", strings.NewReader("{invalid json"))
	r.Header.Set("Content-Type", "application/json")

	tbl, err := BuildLuaRequest(L, r, "10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// body should still have the raw content
	body := L.GetField(tbl, "body")
	if body.String() != "{invalid json" {
		t.Fatalf("expected raw body content, got %q", body.String())
	}

	// json should be nil for malformed JSON
	jsonVal := L.GetField(tbl, "json")
	if jsonVal != lua.LNil {
		t.Fatalf("expected json=nil for malformed JSON, got %T(%v)", jsonVal, jsonVal)
	}
}

func TestBuildLuaRequest_EmptyBody(t *testing.T) {
	L := newTestState()
	defer L.Close()

	r := httptest.NewRequest("POST", "/data", strings.NewReader(""))
	r.Header.Set("Content-Type", "application/json")

	tbl, err := BuildLuaRequest(L, r, "10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// body should be empty
	body := L.GetField(tbl, "body")
	if body.String() != "" {
		t.Fatalf("expected empty body, got %q", body.String())
	}

	// json should be nil for empty body even with JSON Content-Type
	jsonVal := L.GetField(tbl, "json")
	if jsonVal != lua.LNil {
		t.Fatalf("expected json=nil for empty body, got %T(%v)", jsonVal, jsonVal)
	}
}

func TestBuildLuaRequest_BodyTooLarge(t *testing.T) {
	L := newTestState()
	defer L.Close()

	// Use http.MaxBytesReader to simulate body size enforcement.
	largeBody := strings.Repeat("x", 1024)
	r := httptest.NewRequest("POST", "/data", strings.NewReader(largeBody))

	// Wrap the body with MaxBytesReader (limit 10 bytes).
	rec := httptest.NewRecorder()
	r.Body = http.MaxBytesReader(rec, r.Body, 10)

	_, err := BuildLuaRequest(L, r, "10.0.0.1")
	if err == nil {
		t.Fatal("expected error for body exceeding MaxBytesReader limit")
	}
}

// --- goValueToLua tests ---

func TestGoValueToLua_HTTP(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tests := []struct {
		name    string
		input   any
		checkFn func(t *testing.T, v lua.LValue)
	}{
		{
			name:  "nil",
			input: nil,
			checkFn: func(t *testing.T, v lua.LValue) {
				if v != lua.LNil {
					t.Fatalf("expected LNil, got %T(%v)", v, v)
				}
			},
		},
		{
			name:  "bool true",
			input: true,
			checkFn: func(t *testing.T, v lua.LValue) {
				if v != lua.LTrue {
					t.Fatalf("expected LTrue, got %v", v)
				}
			},
		},
		{
			name:  "float64 (JSON number)",
			input: float64(42.5),
			checkFn: func(t *testing.T, v lua.LValue) {
				n, ok := v.(lua.LNumber)
				if !ok || float64(n) != 42.5 {
					t.Fatalf("expected 42.5, got %v", v)
				}
			},
		},
		{
			name:  "string",
			input: "hello",
			checkFn: func(t *testing.T, v lua.LValue) {
				s, ok := v.(lua.LString)
				if !ok || string(s) != "hello" {
					t.Fatalf("expected hello, got %v", v)
				}
			},
		},
		{
			name:  "int (defensive)",
			input: 42,
			checkFn: func(t *testing.T, v lua.LValue) {
				n, ok := v.(lua.LNumber)
				if !ok || float64(n) != 42 {
					t.Fatalf("expected 42, got %v", v)
				}
			},
		},
		{
			name:  "int64 (defensive)",
			input: int64(99),
			checkFn: func(t *testing.T, v lua.LValue) {
				n, ok := v.(lua.LNumber)
				if !ok || float64(n) != 99 {
					t.Fatalf("expected 99, got %v", v)
				}
			},
		},
		{
			name:  "slice",
			input: []any{"a", float64(2), true},
			checkFn: func(t *testing.T, v lua.LValue) {
				tbl, ok := v.(*lua.LTable)
				if !ok {
					t.Fatalf("expected *LTable, got %T", v)
				}
				if tbl.Len() != 3 {
					t.Fatalf("expected length 3, got %d", tbl.Len())
				}
				if tbl.RawGetInt(1).String() != "a" {
					t.Fatalf("expected [1]=a, got %v", tbl.RawGetInt(1))
				}
			},
		},
		{
			name:  "map",
			input: map[string]any{"key": "val"},
			checkFn: func(t *testing.T, v lua.LValue) {
				tbl, ok := v.(*lua.LTable)
				if !ok {
					t.Fatalf("expected *LTable, got %T", v)
				}
				keyVal := tbl.RawGetString("key")
				if keyVal.String() != "val" {
					t.Fatalf("expected key=val, got %v", keyVal)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := goValueToLua(L, tt.input)
			tt.checkFn(t, result)
		})
	}
}

// --- WriteLuaResponse tests ---

func TestWriteLuaResponse_JSONResponse(t *testing.T) {
	L := newTestState()
	defer L.Close()

	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(200))

	jsonTbl := L.NewTable()
	jsonTbl.RawSetString("message", lua.LString("ok"))
	jsonTbl.RawSetString("count", lua.LNumber(42))
	respTbl.RawSetString("json", jsonTbl)

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "req-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]any
	if jsonErr := json.Unmarshal(rec.Body.Bytes(), &body); jsonErr != nil {
		t.Fatalf("failed to unmarshal response: %v", jsonErr)
	}
	if body["message"] != "ok" {
		t.Fatalf("expected message=ok, got %v", body["message"])
	}
	if body["count"] != float64(42) {
		t.Fatalf("expected count=42, got %v", body["count"])
	}
}

func TestWriteLuaResponse_RawBodyResponse(t *testing.T) {
	L := newTestState()
	defer L.Close()

	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(201))
	respTbl.RawSetString("body", lua.LString("plain text response"))

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "req-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != 201 {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	if rec.Body.String() != "plain text response" {
		t.Fatalf("expected body %q, got %q", "plain text response", rec.Body.String())
	}
}

func TestWriteLuaResponse_BodyAndJSON_JSONWins(t *testing.T) {
	L := newTestState()
	defer L.Close()

	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(200))
	respTbl.RawSetString("body", lua.LString("should be ignored"))

	jsonTbl := L.NewTable()
	jsonTbl.RawSetString("winner", lua.LString("json"))
	respTbl.RawSetString("json", jsonTbl)

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "req-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// JSON should win over body.
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]any
	if jsonErr := json.Unmarshal(rec.Body.Bytes(), &body); jsonErr != nil {
		t.Fatalf("failed to unmarshal response: %v", jsonErr)
	}
	if body["winner"] != "json" {
		t.Fatalf("expected winner=json, got %v", body["winner"])
	}
}

func TestWriteLuaResponse_BlockedHeadersFiltered(t *testing.T) {
	L := newTestState()
	defer L.Close()

	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(200))
	respTbl.RawSetString("body", lua.LString("ok"))

	headersTbl := L.NewTable()
	headersTbl.RawSetString("Set-Cookie", lua.LString("evil=1"))
	headersTbl.RawSetString("Access-Control-Allow-Origin", lua.LString("*"))
	headersTbl.RawSetString("X-Custom", lua.LString("allowed"))
	headersTbl.RawSetString("Transfer-Encoding", lua.LString("chunked"))
	headersTbl.RawSetString("Content-Length", lua.LString("999"))
	headersTbl.RawSetString("Cache-Control", lua.LString("public, max-age=31536000"))
	respTbl.RawSetString("headers", headersTbl)

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "req-blocked")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Blocked headers should NOT be present.
	if rec.Header().Get("Set-Cookie") != "" {
		t.Fatal("Set-Cookie should be blocked")
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("Access-Control-Allow-Origin should be blocked")
	}
	if rec.Header().Get("Transfer-Encoding") != "" {
		t.Fatal("Transfer-Encoding should be blocked")
	}

	// Custom header should pass through.
	if rec.Header().Get("X-Custom") != "allowed" {
		t.Fatalf("expected X-Custom=allowed, got %q", rec.Header().Get("X-Custom"))
	}

	// Cache-Control should be the default (no-store), not the plugin's value.
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected Cache-Control=no-store, got %q", rec.Header().Get("Cache-Control"))
	}
}

func TestWriteLuaResponse_SecurityHeadersAlwaysPresent(t *testing.T) {
	L := newTestState()
	defer L.Close()

	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(200))
	respTbl.RawSetString("body", lua.LString("ok"))

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "req-sec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// X-Content-Type-Options: nosniff
	xcto := rec.Header().Get("X-Content-Type-Options")
	if xcto != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options=nosniff, got %q", xcto)
	}

	// X-Frame-Options: DENY
	xfo := rec.Header().Get("X-Frame-Options")
	if xfo != "DENY" {
		t.Fatalf("expected X-Frame-Options=DENY, got %q", xfo)
	}

	// Cache-Control: no-store
	cc := rec.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Fatalf("expected Cache-Control=no-store, got %q", cc)
	}
}

func TestWriteLuaResponse_ResponseTooLarge(t *testing.T) {
	L := newTestState()
	defer L.Close()

	// Create a response body larger than the limit.
	largeBody := strings.Repeat("x", 1024)
	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(200))
	respTbl.RawSetString("body", lua.LString(largeBody))

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 100, "req-too-large")

	// Should return an error.
	if err == nil {
		t.Fatal("expected error for response exceeding size limit")
	}

	// Should write a 500 RESPONSE_TOO_LARGE error.
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var resp PluginErrorResponse
	if jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp); jsonErr != nil {
		t.Fatalf("failed to unmarshal error response: %v", jsonErr)
	}
	if resp.Error.Code != "RESPONSE_TOO_LARGE" {
		t.Fatalf("expected error code RESPONSE_TOO_LARGE, got %q", resp.Error.Code)
	}
}

func TestWriteLuaResponse_DefaultStatus200(t *testing.T) {
	L := newTestState()
	defer L.Close()

	// No explicit status set.
	respTbl := L.NewTable()
	respTbl.RawSetString("body", lua.LString("ok"))

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != 200 {
		t.Fatalf("expected default status 200, got %d", rec.Code)
	}
}

func TestWriteLuaResponse_EmptyResponse(t *testing.T) {
	L := newTestState()
	defer L.Close()

	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(204))

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != 204 {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}
}

func TestWriteLuaResponse_JSONArrayResponse(t *testing.T) {
	L := newTestState()
	defer L.Close()

	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(200))

	// Create a Lua array (sequence table).
	arr := L.NewTable()
	item1 := L.NewTable()
	item1.RawSetString("id", lua.LString("1"))
	item1.RawSetString("name", lua.LString("alice"))
	arr.Append(item1)

	item2 := L.NewTable()
	item2.RawSetString("id", lua.LString("2"))
	item2.RawSetString("name", lua.LString("bob"))
	arr.Append(item2)

	respTbl.RawSetString("json", arr)

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body []map[string]any
	if jsonErr := json.Unmarshal(rec.Body.Bytes(), &body); jsonErr != nil {
		t.Fatalf("failed to unmarshal array response: %v", jsonErr)
	}
	if len(body) != 2 {
		t.Fatalf("expected 2 items, got %d", len(body))
	}
	if body[0]["name"] != "alice" {
		t.Fatalf("expected first name=alice, got %v", body[0]["name"])
	}
}

// --- LuaValueToGo round-trip through WriteLuaResponse tests ---
// The WriteLuaResponse function uses LuaValueToGo from lua_helpers.go
// for Lua->Go conversion. These tests verify correct behavior through
// the JSON response path.

func TestWriteLuaResponse_LuaMapToJSON(t *testing.T) {
	L := newTestState()
	defer L.Close()

	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(200))

	jsonTbl := L.NewTable()
	jsonTbl.RawSetString("name", lua.LString("alice"))
	jsonTbl.RawSetString("active", lua.LTrue)
	jsonTbl.RawSetString("score", lua.LNumber(95.5))
	respTbl.RawSetString("json", jsonTbl)

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body map[string]any
	if jsonErr := json.Unmarshal(rec.Body.Bytes(), &body); jsonErr != nil {
		t.Fatalf("failed to unmarshal response: %v", jsonErr)
	}
	if body["name"] != "alice" {
		t.Fatalf("expected name=alice, got %v", body["name"])
	}
	if body["active"] != true {
		t.Fatalf("expected active=true, got %v", body["active"])
	}
	if body["score"] != float64(95.5) {
		t.Fatalf("expected score=95.5, got %v", body["score"])
	}
}

func TestWriteLuaResponse_LuaSequenceToJSONArray(t *testing.T) {
	L := newTestState()
	defer L.Close()

	respTbl := L.NewTable()
	respTbl.RawSetString("status", lua.LNumber(200))

	arr := L.NewTable()
	arr.Append(lua.LString("a"))
	arr.Append(lua.LNumber(2))
	arr.Append(lua.LTrue)
	respTbl.RawSetString("json", arr)

	rec := httptest.NewRecorder()
	err := WriteLuaResponse(rec, L, respTbl, 5<<20, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body []any
	if jsonErr := json.Unmarshal(rec.Body.Bytes(), &body); jsonErr != nil {
		t.Fatalf("failed to unmarshal response: %v", jsonErr)
	}
	if len(body) != 3 {
		t.Fatalf("expected length 3, got %d", len(body))
	}
	if body[0] != "a" {
		t.Fatalf("expected [0]=a, got %v", body[0])
	}
	if body[1] != float64(2) {
		t.Fatalf("expected [1]=2, got %v", body[1])
	}
	if body[2] != true {
		t.Fatalf("expected [2]=true, got %v", body[2])
	}
}

// --- extractClientIP tests ---

func TestExtractClientIP_NoTrustedProxies_SimpleRemoteAddr(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "192.168.1.100:54321"

	ip := extractClientIP(r, nil)
	if ip != "192.168.1.100" {
		t.Fatalf("expected 192.168.1.100, got %q", ip)
	}
}

func TestExtractClientIP_NoTrustedProxies_XFFIgnored(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "192.168.1.100:54321"
	r.Header.Set("X-Forwarded-For", "10.0.0.1, 172.16.0.1")

	// Without trusted proxies, XFF is completely ignored.
	ip := extractClientIP(r, nil)
	if ip != "192.168.1.100" {
		t.Fatalf("expected 192.168.1.100 (RemoteAddr), got %q", ip)
	}
}

func TestExtractClientIP_TrustedProxies_MixedXFF(t *testing.T) {
	_, cidr, _ := net.ParseCIDR("10.0.0.0/8")
	trustedProxies := []*net.IPNet{cidr}

	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "10.0.0.1:54321"
	// XFF from left to right: client, proxy1, proxy2
	// From right to left: 10.0.0.5 (trusted), 203.0.113.50 (not trusted)
	r.Header.Set("X-Forwarded-For", "203.0.113.50, 10.0.0.5")

	ip := extractClientIP(r, trustedProxies)
	if ip != "203.0.113.50" {
		t.Fatalf("expected 203.0.113.50 (first non-trusted from right), got %q", ip)
	}
}

func TestExtractClientIP_TrustedProxies_AllTrustedFallsBackToRemoteAddr(t *testing.T) {
	_, cidr1, _ := net.ParseCIDR("10.0.0.0/8")
	_, cidr2, _ := net.ParseCIDR("172.16.0.0/12")
	trustedProxies := []*net.IPNet{cidr1, cidr2}

	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "10.0.0.1:54321"
	r.Header.Set("X-Forwarded-For", "10.0.0.2, 172.16.0.3")

	// All XFF entries are trusted, so fall back to RemoteAddr.
	ip := extractClientIP(r, trustedProxies)
	if ip != "10.0.0.1" {
		t.Fatalf("expected 10.0.0.1 (RemoteAddr fallback), got %q", ip)
	}
}

func TestExtractClientIP_PortStripping_IPv4(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "203.0.113.50:8080"

	ip := extractClientIP(r, nil)
	if ip != "203.0.113.50" {
		t.Fatalf("expected 203.0.113.50, got %q", ip)
	}
}

func TestExtractClientIP_PortStripping_IPv6(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "[::1]:8080"

	ip := extractClientIP(r, nil)
	if ip != "::1" {
		t.Fatalf("expected ::1, got %q", ip)
	}
}

func TestExtractClientIP_TrustedProxies_IPv6InXFF(t *testing.T) {
	_, cidr, _ := net.ParseCIDR("::1/128")
	trustedProxies := []*net.IPNet{cidr}

	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "[::1]:54321"
	r.Header.Set("X-Forwarded-For", "2001:db8::1, ::1")

	ip := extractClientIP(r, trustedProxies)
	if ip != "2001:db8::1" {
		t.Fatalf("expected 2001:db8::1 (non-trusted), got %q", ip)
	}
}

func TestExtractClientIP_NoTrustedProxies_BareIPNoPort(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	// Some test setups may have RemoteAddr without a port.
	r.RemoteAddr = "192.168.1.1"

	ip := extractClientIP(r, nil)
	if ip != "192.168.1.1" {
		t.Fatalf("expected 192.168.1.1, got %q", ip)
	}
}

func TestExtractClientIP_EmptyTrustedProxies(t *testing.T) {
	// Empty slice (not nil) should behave the same as nil.
	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "192.168.1.100:54321"
	r.Header.Set("X-Forwarded-For", "10.0.0.1")

	ip := extractClientIP(r, []*net.IPNet{})
	if ip != "192.168.1.100" {
		t.Fatalf("expected 192.168.1.100 (RemoteAddr, XFF ignored), got %q", ip)
	}
}

func TestExtractClientIP_TrustedProxies_NoXFF(t *testing.T) {
	_, cidr, _ := net.ParseCIDR("10.0.0.0/8")
	trustedProxies := []*net.IPNet{cidr}

	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "10.0.0.1:54321"
	// No XFF header at all.

	ip := extractClientIP(r, trustedProxies)
	if ip != "10.0.0.1" {
		t.Fatalf("expected 10.0.0.1 (RemoteAddr fallback), got %q", ip)
	}
}

func TestExtractClientIP_TrustedProxies_XFFWithSpaces(t *testing.T) {
	_, cidr, _ := net.ParseCIDR("10.0.0.0/8")
	trustedProxies := []*net.IPNet{cidr}

	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "10.0.0.1:54321"
	// XFF entries with extra spaces.
	r.Header.Set("X-Forwarded-For", "  203.0.113.50 , 10.0.0.5 ")

	ip := extractClientIP(r, trustedProxies)
	if ip != "203.0.113.50" {
		t.Fatalf("expected 203.0.113.50, got %q", ip)
	}
}

// --- extractPatternParams tests ---

func TestExtractPatternParams(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "no params",
			pattern:  "GET /api/v1/tasks",
			expected: nil,
		},
		{
			name:     "single param",
			pattern:  "GET /api/v1/tasks/{id}",
			expected: []string{"id"},
		},
		{
			name:     "multiple params",
			pattern:  "GET /api/v1/plugins/{plugin}/tasks/{id}",
			expected: []string{"plugin", "id"},
		},
		{
			name:     "wildcard param",
			pattern:  "GET /files/{path...}",
			expected: []string{"path"},
		},
		{
			name:     "empty pattern",
			pattern:  "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPatternParams(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d params, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, exp := range tt.expected {
				if result[i] != exp {
					t.Fatalf("param[%d]: expected %q, got %q", i, exp, result[i])
				}
			}
		})
	}
}

// --- stripPort tests ---

func TestStripPort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"IPv4 with port", "192.168.1.1:8080", "192.168.1.1"},
		{"IPv4 without port", "192.168.1.1", "192.168.1.1"},
		{"IPv6 with port", "[::1]:8080", "::1"},
		{"IPv6 without port", "::1", "::1"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripPort(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// --- blockedResponseHeaders coverage ---

func TestBlockedResponseHeaders_AllExpectedEntries(t *testing.T) {
	// Verify all headers from the spec are present.
	expected := []string{
		"access-control-allow-origin",
		"access-control-allow-credentials",
		"access-control-allow-methods",
		"access-control-allow-headers",
		"access-control-expose-headers",
		"set-cookie",
		"transfer-encoding",
		"content-length",
		"host",
		"connection",
		"cache-control",
	}

	for _, h := range expected {
		if !blockedResponseHeaders[h] {
			t.Fatalf("expected %q to be in blockedResponseHeaders", h)
		}
	}

	if len(blockedResponseHeaders) != len(expected) {
		t.Fatalf("expected %d blocked headers, got %d", len(expected), len(blockedResponseHeaders))
	}
}
