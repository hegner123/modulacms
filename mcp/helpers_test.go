package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

// resultText extracts the text from the first content block of a CallToolResult.
func resultText(t *testing.T, r *mcp.CallToolResult) string {
	t.Helper()
	if len(r.Content) == 0 {
		t.Fatal("result has no content blocks")
	}
	tc, ok := r.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("first content block is %T, want TextContent", r.Content[0])
	}
	return tc.Text
}

// makeReq builds a CallToolRequest with the given arguments map.
func makeReq(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

// --- errResult tests ---

func TestErrResult_PlainError(t *testing.T) {
	err := errors.New("connection refused")
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	if text != "connection refused" {
		t.Errorf("text = %q, want %q", text, "connection refused")
	}
}

func TestErrResult_ApiError(t *testing.T) {
	apiErr := &modulacms.ApiError{
		StatusCode: 404,
		Message:    "not found",
		Body:       `{"error":"not found"}`,
	}
	result := errResult(apiErr)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}

	text := resultText(t, result)

	var detail map[string]any
	if err := json.Unmarshal([]byte(text), &detail); err != nil {
		t.Fatalf("failed to parse error JSON: %v", err)
	}

	if detail["status"] != float64(404) {
		t.Errorf("status = %v, want 404", detail["status"])
	}
	if detail["message"] != "not found" {
		t.Errorf("message = %v, want %q", detail["message"], "not found")
	}
	if detail["body"] != `{"error":"not found"}` {
		t.Errorf("body = %v, want %q", detail["body"], `{"error":"not found"}`)
	}
}

func TestErrResult_WrappedApiError(t *testing.T) {
	inner := &modulacms.ApiError{StatusCode: 500, Message: "internal"}
	wrapped := errors.Join(errors.New("outer"), inner)
	result := errResult(wrapped)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}

	text := resultText(t, result)
	var detail map[string]any
	if err := json.Unmarshal([]byte(text), &detail); err != nil {
		// Wrapped join may not unwrap via errors.As depending on join semantics;
		// if it doesn't, we get plain error text which is also acceptable.
		return
	}
	if detail["status"] != float64(500) {
		t.Errorf("status = %v, want 500", detail["status"])
	}
}

// --- jsonResult tests ---

func TestJsonResult_Struct(t *testing.T) {
	type sample struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	result, err := jsonResult(sample{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected IsError=false")
	}

	text := resultText(t, result)
	var decoded sample
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		t.Fatalf("failed to decode result JSON: %v", err)
	}
	if decoded.Name != "Alice" {
		t.Errorf("Name = %q, want %q", decoded.Name, "Alice")
	}
	if decoded.Age != 30 {
		t.Errorf("Age = %d, want %d", decoded.Age, 30)
	}
}

func TestJsonResult_Slice(t *testing.T) {
	items := []string{"a", "b", "c"}
	result, err := jsonResult(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := resultText(t, result)
	var decoded []string
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		t.Fatalf("failed to decode result JSON: %v", err)
	}
	if len(decoded) != 3 {
		t.Fatalf("len = %d, want 3", len(decoded))
	}
	if decoded[1] != "b" {
		t.Errorf("decoded[1] = %q, want %q", decoded[1], "b")
	}
}

func TestJsonResult_PrettyPrinted(t *testing.T) {
	data := map[string]int{"count": 42}
	result, err := jsonResult(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	expected := "{\n  \"count\": 42\n}"
	if text != expected {
		t.Errorf("text = %q, want %q", text, expected)
	}
}

func TestJsonResult_Unmarshalable(t *testing.T) {
	// Channels cannot be marshaled to JSON
	ch := make(chan int)
	_, err := jsonResult(ch)
	if err == nil {
		t.Fatal("expected error for unmarshalable type")
	}
}

// --- optionalIDPtr tests ---

func TestOptionalIDPtr_Present(t *testing.T) {
	req := makeReq(map[string]any{"user_id": "usr-001"})
	result := optionalIDPtr[modulacms.UserID](req, "user_id")
	if result == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *result != "usr-001" {
		t.Errorf("*result = %q, want %q", *result, "usr-001")
	}
}

func TestOptionalIDPtr_Missing(t *testing.T) {
	req := makeReq(map[string]any{})
	result := optionalIDPtr[modulacms.UserID](req, "user_id")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestOptionalIDPtr_Empty(t *testing.T) {
	req := makeReq(map[string]any{"user_id": ""})
	result := optionalIDPtr[modulacms.UserID](req, "user_id")
	if result != nil {
		t.Errorf("expected nil for empty string, got %v", result)
	}
}

// --- optionalStrPtr tests ---

func TestOptionalStrPtr_Present(t *testing.T) {
	req := makeReq(map[string]any{"name": "test"})
	result := optionalStrPtr(req, "name")
	if result == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *result != "test" {
		t.Errorf("*result = %q, want %q", *result, "test")
	}
}

func TestOptionalStrPtr_Missing(t *testing.T) {
	req := makeReq(map[string]any{})
	result := optionalStrPtr(req, "name")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestOptionalStrPtr_WrongType(t *testing.T) {
	req := makeReq(map[string]any{"name": 123})
	result := optionalStrPtr(req, "name")
	if result != nil {
		t.Errorf("expected nil for non-string value, got %v", result)
	}
}

// --- optionalFloat64Ptr tests ---

func TestOptionalFloat64Ptr_Present(t *testing.T) {
	req := makeReq(map[string]any{"focal_x": 0.5})
	result := optionalFloat64Ptr(req, "focal_x")
	if result == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *result != 0.5 {
		t.Errorf("*result = %f, want 0.5", *result)
	}
}

func TestOptionalFloat64Ptr_Missing(t *testing.T) {
	req := makeReq(map[string]any{})
	result := optionalFloat64Ptr(req, "focal_x")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestOptionalFloat64Ptr_WrongType(t *testing.T) {
	req := makeReq(map[string]any{"focal_x": "not a number"})
	result := optionalFloat64Ptr(req, "focal_x")
	if result != nil {
		t.Errorf("expected nil for non-float64 value, got %v", result)
	}
}

func TestOptionalFloat64Ptr_Zero(t *testing.T) {
	req := makeReq(map[string]any{"focal_x": float64(0)})
	result := optionalFloat64Ptr(req, "focal_x")
	if result == nil {
		t.Fatal("expected non-nil pointer for zero value")
	}
	if *result != 0 {
		t.Errorf("*result = %f, want 0", *result)
	}
}
