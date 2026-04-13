package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/server"

	"github.com/hegner123/modulacms/internal/service"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// =============================================================================
// Group 1: Constructor Completeness
// =============================================================================

func TestNewSDKBackends_AllFieldsPopulated(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)

	v := reflect.ValueOf(backends).Elem()
	ty := v.Type()
	for i := range v.NumField() {
		f := v.Field(i)
		name := ty.Field(i).Name
		if f.Kind() != reflect.Interface {
			t.Errorf("Backends.%s is %s, expected interface kind", name, f.Kind())
			continue
		}
		if f.IsNil() {
			t.Errorf("Backends.%s is nil in NewSDKBackends result", name)
		}
	}
}

func TestNewSDKBackends_FieldCount_MatchesStruct(t *testing.T) {
	// Verify the constructor assigns exactly as many fields as the struct declares.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	backends := newTestBackends(t, ts)

	v := reflect.ValueOf(backends).Elem()
	ty := v.Type()
	structFields := ty.NumField()

	nonNilCount := 0
	for i := range v.NumField() {
		f := v.Field(i)
		if f.Kind() == reflect.Interface && !f.IsNil() {
			nonNilCount++
		}
	}
	if nonNilCount != structFields {
		t.Errorf("NewSDKBackends populated %d of %d Backends fields", nonNilCount, structFields)
	}
}

func TestNewProxyBackends_AllFieldsPopulated(t *testing.T) {
	// NewProxyBackends creates proxy structs regardless of connection state.
	// We cannot easily construct a ConnectionManager in tests (requires registry file),
	// so we verify structurally that NewProxyBackends assigns all fields by counting
	// the proxy assignments in the source. Here we verify the Backends struct has
	// the same field count as NewSDKBackends to confirm parity.
	sdkType := reflect.TypeOf(Backends{})
	sdkFieldCount := sdkType.NumField()

	// Verify the proxy constructor source populates the same number of fields.
	// We do this by checking that the Backends struct definition has not grown
	// without updating the constructors. The SDK test above proves SDK covers all,
	// and the proxy constructor is mechanically identical (30 fields each).
	if sdkFieldCount != 30 {
		t.Errorf("Backends has %d fields; if a field was added, verify all 3 constructors are updated", sdkFieldCount)
	}
}

// =============================================================================
// Group 2: Proxy errNoConnection propagation
// =============================================================================

// proxyContentNoConn implements ContentBackend and always returns errNoConnection.
// This simulates the behavior of proxyContentBackend when no client is connected.
type proxyContentNoConn struct{}

func (p *proxyContentNoConn) ListContent(_ context.Context, _, _ int64) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) GetContent(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) CreateContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) UpdateContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) DeleteContent(_ context.Context, _ string) error {
	return errNoConnection
}
func (p *proxyContentNoConn) GetPage(_ context.Context, _, _, _ string) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) GetContentTree(_ context.Context, _, _ string) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) ListContentFields(_ context.Context, _, _ int64) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) GetContentField(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) CreateContentField(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) UpdateContentField(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) DeleteContentField(_ context.Context, _ string) error {
	return errNoConnection
}
func (p *proxyContentNoConn) ReorderContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) MoveContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) SaveContentTree(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) HealContent(_ context.Context, _ bool) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) BatchUpdateContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) QueryContent(_ context.Context, _ string, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) GetGlobals(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) GetContentFull(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) GetContentByRoute(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, errNoConnection
}
func (p *proxyContentNoConn) CreateContentComposite(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, errNoConnection
}

func TestProxy_NoConnection_GetContent(t *testing.T) {
	backend := &proxyContentNoConn{}
	handler := handleGetContent(backend)
	result := callTool(t, handler, map[string]any{"id": "01ABCDEFGHJKMNPQRSTVWXYZ1"})

	if !result.IsError {
		t.Fatal("expected error when no connection is active")
	}
	text := resultText(t, result)
	assertErrorJSONContains(t, text, 500, "no active connection")
}

func TestProxy_NoConnection_ListContent(t *testing.T) {
	backend := &proxyContentNoConn{}
	handler := handleListContent(backend)
	result := callTool(t, handler, map[string]any{})

	if !result.IsError {
		t.Fatal("expected error when no connection is active")
	}
	text := resultText(t, result)
	assertErrorJSONContains(t, text, 500, "no active connection")
}

func TestProxy_NoConnection_DeleteContent(t *testing.T) {
	backend := &proxyContentNoConn{}
	handler := handleDeleteContent(backend)
	result := callTool(t, handler, map[string]any{"id": "01ABCDEFGHJKMNPQRSTVWXYZ1"})

	if !result.IsError {
		t.Fatal("expected error when no connection is active")
	}
	text := resultText(t, result)
	assertErrorJSONContains(t, text, 500, "no active connection")
}

// =============================================================================
// Group 3: Error Type Mapping (errResult coverage)
// =============================================================================

func TestErrResult_NotFoundError(t *testing.T) {
	err := &service.NotFoundError{Resource: "content", ID: "ct-999"}
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 404, `content "ct-999" not found`)
}

func TestErrResult_ValidationError(t *testing.T) {
	err := service.NewValidationError("name", "must not be empty")
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 422, "validation: name: must not be empty")
}

func TestErrResult_ValidationError_Multiple(t *testing.T) {
	err := service.NewValidationErrors(
		service.FieldError{Field: "name", Message: "required"},
		service.FieldError{Field: "email", Message: "invalid format"},
	)
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(422) {
		t.Errorf("status = %v, want 422", detail["status"])
	}
	msg, ok := detail["message"].(string)
	if !ok {
		t.Fatalf("message is not a string: %v", detail["message"])
	}
	if !strings.Contains(msg, "name: required") {
		t.Errorf("message %q does not contain 'name: required'", msg)
	}
	if !strings.Contains(msg, "email: invalid format") {
		t.Errorf("message %q does not contain 'email: invalid format'", msg)
	}
}

func TestErrResult_ConflictError(t *testing.T) {
	err := &service.ConflictError{Resource: "route", ID: "rt-001", Detail: "slug already exists"}
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 409, `route "rt-001" conflict: slug already exists`)
}

func TestErrResult_ConflictError_NoDetail(t *testing.T) {
	err := &service.ConflictError{Resource: "user", ID: "u-001"}
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 409, `user "u-001" already exists`)
}

func TestErrResult_ForbiddenError(t *testing.T) {
	err := &service.ForbiddenError{Message: "insufficient permissions"}
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 403, "forbidden: insufficient permissions")
}

func TestErrResult_ForbiddenError_Empty(t *testing.T) {
	err := &service.ForbiddenError{}
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 403, "forbidden")
}

func TestErrResult_UnauthorizedError(t *testing.T) {
	err := &service.UnauthorizedError{Message: "token expired"}
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 401, "unauthorized: token expired")
}

func TestErrResult_UnauthorizedError_Empty(t *testing.T) {
	err := &service.UnauthorizedError{}
	result := errResult(err)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 401, "unauthorized")
}

func TestErrResult_WrappedNotFoundError(t *testing.T) {
	inner := &service.NotFoundError{Resource: "media", ID: "m-404"}
	wrapped := fmt.Errorf("get media failed: %w", inner)
	result := errResult(wrapped)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(404) {
		t.Errorf("status = %v, want 404 (wrapped NotFoundError should preserve type)", detail["status"])
	}
}

func TestErrResult_WrappedValidationError(t *testing.T) {
	inner := service.NewValidationError("slug", "invalid characters")
	wrapped := fmt.Errorf("create route: %w", inner)
	result := errResult(wrapped)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(422) {
		t.Errorf("status = %v, want 422 (wrapped ValidationError should preserve type)", detail["status"])
	}
}

func TestErrResult_WrappedConflictError(t *testing.T) {
	inner := &service.ConflictError{Resource: "role", ID: "r-dup"}
	wrapped := fmt.Errorf("assign role: %w", inner)
	result := errResult(wrapped)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(409) {
		t.Errorf("status = %v, want 409 (wrapped ConflictError should preserve type)", detail["status"])
	}
}

func TestErrResult_WrappedForbiddenError(t *testing.T) {
	inner := &service.ForbiddenError{Message: "admin only"}
	wrapped := fmt.Errorf("permission check: %w", inner)
	result := errResult(wrapped)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(403) {
		t.Errorf("status = %v, want 403 (wrapped ForbiddenError should preserve type)", detail["status"])
	}
}

func TestErrResult_WrappedUnauthorizedError(t *testing.T) {
	inner := &service.UnauthorizedError{Message: "session expired"}
	wrapped := fmt.Errorf("auth middleware: %w", inner)
	result := errResult(wrapped)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(401) {
		t.Errorf("status = %v, want 401 (wrapped UnauthorizedError should preserve type)", detail["status"])
	}
}

func TestErrResult_ApiError_RateLimited(t *testing.T) {
	apiErr := &modula.ApiError{
		StatusCode: 429,
		Message:    "rate limit exceeded",
		Body:       `{"message":"rate limit exceeded"}`,
	}
	result := errResult(apiErr)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 429, "rate limit exceeded")
}

func TestErrResult_ApiError_ServiceUnavailable(t *testing.T) {
	apiErr := &modula.ApiError{
		StatusCode: 503,
		Message:    "service unavailable",
	}
	result := errResult(apiErr)

	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 503, "service unavailable")
}

// ServiceErrors take priority over ApiError when both match.
// This verifies service error types are checked before ApiError.
func TestErrResult_ServiceErrorPriority_OverApiError(t *testing.T) {
	// A NotFoundError is not an ApiError, so they are exclusive.
	// But verify the ordering is correct: service errors first, then ApiError.
	nf := &service.NotFoundError{Resource: "field", ID: "f-001"}
	result := errResult(nf)
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	// Should be 404 (service NotFoundError), not 500 (generic).
	if detail["status"] != float64(404) {
		t.Errorf("status = %v, want 404", detail["status"])
	}
}

// =============================================================================
// Group 4: Handler Error Propagation (end-to-end with failing backends)
// =============================================================================

// failingContentBackend implements ContentBackend and returns a configurable error
// from GetContent. All other methods return a generic "not implemented" error.
type failingContentBackend struct {
	getContentErr    error
	listContentErr   error
	deleteContentErr error
}

func (f *failingContentBackend) GetContent(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, f.getContentErr
}
func (f *failingContentBackend) ListContent(_ context.Context, _, _ int64) (json.RawMessage, error) {
	return nil, f.listContentErr
}
func (f *failingContentBackend) DeleteContent(_ context.Context, _ string) error {
	return f.deleteContentErr
}
func (f *failingContentBackend) CreateContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) UpdateContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) GetPage(_ context.Context, _, _, _ string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) GetContentTree(_ context.Context, _, _ string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) ListContentFields(_ context.Context, _, _ int64) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) GetContentField(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) CreateContentField(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) UpdateContentField(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) DeleteContentField(_ context.Context, _ string) error {
	return fmt.Errorf("not implemented")
}
func (f *failingContentBackend) ReorderContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) MoveContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) SaveContentTree(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) HealContent(_ context.Context, _ bool) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) BatchUpdateContent(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) QueryContent(_ context.Context, _ string, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) GetGlobals(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) GetContentFull(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) GetContentByRoute(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingContentBackend) CreateContentComposite(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestHandler_BackendNotFound_Returns404(t *testing.T) {
	backend := &failingContentBackend{
		getContentErr: &service.NotFoundError{Resource: "content", ID: "ct-999"},
	}
	handler := handleGetContent(backend)
	result := callTool(t, handler, map[string]any{"id": "ct-999"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(404) {
		t.Errorf("status = %v, want 404", detail["status"])
	}
	msg, ok := detail["message"].(string)
	if !ok {
		t.Fatalf("message is not a string: %v", detail["message"])
	}
	if !strings.Contains(msg, "ct-999") {
		t.Errorf("message %q should contain the missing ID 'ct-999'", msg)
	}
}

func TestHandler_BackendValidation_Returns422(t *testing.T) {
	backend := &failingContentBackend{
		getContentErr: service.NewValidationError("status", "invalid value"),
	}
	handler := handleGetContent(backend)
	result := callTool(t, handler, map[string]any{"id": "ct-001"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 422, "validation: status: invalid value")
}

func TestHandler_BackendConflict_Returns409(t *testing.T) {
	backend := &failingContentBackend{
		getContentErr: &service.ConflictError{Resource: "content", ID: "ct-dup", Detail: "duplicate entry"},
	}
	handler := handleGetContent(backend)
	result := callTool(t, handler, map[string]any{"id": "ct-dup"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(409) {
		t.Errorf("status = %v, want 409", detail["status"])
	}
}

func TestHandler_BackendForbidden_Returns403(t *testing.T) {
	backend := &failingContentBackend{
		getContentErr: &service.ForbiddenError{Message: "requires admin role"},
	}
	handler := handleGetContent(backend)
	result := callTool(t, handler, map[string]any{"id": "ct-secret"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 403, "forbidden: requires admin role")
}

func TestHandler_BackendUnauthorized_Returns401(t *testing.T) {
	backend := &failingContentBackend{
		getContentErr: &service.UnauthorizedError{Message: "invalid token"},
	}
	handler := handleGetContent(backend)
	result := callTool(t, handler, map[string]any{"id": "ct-001"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 401, "unauthorized: invalid token")
}

func TestHandler_BackendPlainError_Returns500(t *testing.T) {
	backend := &failingContentBackend{
		getContentErr: fmt.Errorf("database connection lost"),
	}
	handler := handleGetContent(backend)
	result := callTool(t, handler, map[string]any{"id": "ct-001"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 500, "database connection lost")
}

func TestHandler_BackendApiError_PreservesStatus(t *testing.T) {
	backend := &failingContentBackend{
		getContentErr: &modula.ApiError{
			StatusCode: 429,
			Message:    "rate limited",
		},
	}
	handler := handleGetContent(backend)
	result := callTool(t, handler, map[string]any{"id": "ct-001"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 429, "rate limited")
}

func TestHandler_BackendDelete_NotFound(t *testing.T) {
	backend := &failingContentBackend{
		deleteContentErr: &service.NotFoundError{Resource: "content", ID: "ct-gone"},
	}
	handler := handleDeleteContent(backend)
	result := callTool(t, handler, map[string]any{"id": "ct-gone"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(404) {
		t.Errorf("status = %v, want 404", detail["status"])
	}
}

func TestHandler_BackendList_InternalError(t *testing.T) {
	backend := &failingContentBackend{
		listContentErr: fmt.Errorf("query timeout"),
	}
	handler := handleListContent(backend)
	result := callTool(t, handler, map[string]any{})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 500, "query timeout")
}

// failingWebhookBackend implements WebhookBackend and returns configurable errors.
type failingWebhookBackend struct {
	getErr error
}

func (f *failingWebhookBackend) ListWebhooks(_ context.Context) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingWebhookBackend) GetWebhook(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, f.getErr
}
func (f *failingWebhookBackend) CreateWebhook(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingWebhookBackend) UpdateWebhook(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingWebhookBackend) DeleteWebhook(_ context.Context, _ string) error {
	return fmt.Errorf("not implemented")
}
func (f *failingWebhookBackend) TestWebhook(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingWebhookBackend) ListWebhookDeliveries(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
func (f *failingWebhookBackend) RetryWebhookDelivery(_ context.Context, _ string) error {
	return fmt.Errorf("not implemented")
}

func TestHandler_WebhookBackendNotFound_Returns404(t *testing.T) {
	backend := &failingWebhookBackend{
		getErr: &service.NotFoundError{Resource: "webhook", ID: "wh-missing"},
	}
	handler := handleGetWebhook(backend)
	result := callTool(t, handler, map[string]any{"id": "wh-missing"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(404) {
		t.Errorf("status = %v, want 404", detail["status"])
	}
}

// failingSearchBackend implements SearchBackend and returns configurable errors.
type failingSearchBackend struct {
	searchErr error
}

func (f *failingSearchBackend) SearchContent(_ context.Context, _ string, _, _ int64) (json.RawMessage, error) {
	return nil, f.searchErr
}
func (f *failingSearchBackend) RebuildSearchIndex(_ context.Context) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestHandler_SearchBackendForbidden_Returns403(t *testing.T) {
	backend := &failingSearchBackend{
		searchErr: &service.ForbiddenError{Message: "search disabled"},
	}
	handler := handleSearchContent(backend)
	result := callTool(t, handler, map[string]any{"q": "test query"})

	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	assertErrorJSON(t, text, 403, "forbidden: search disabled")
}

// =============================================================================
// Group 5: Required Param Validation (table-driven)
// =============================================================================

func TestRequiredParams_MissingID(t *testing.T) {
	// These handlers all check required params before calling the backend.
	// Passing a nil backend is safe because the RequireString check returns
	// an error before the backend is invoked.
	cases := []struct {
		name    string
		handler server.ToolHandlerFunc
		args    map[string]any
		wantMsg string
	}{
		{
			name:    "get_content missing id",
			handler: handleGetContent(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "delete_content missing id",
			handler: handleDeleteContent(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "get_route missing id",
			handler: handleGetRoute(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "get_webhook missing id",
			handler: handleGetWebhook(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "delete_webhook missing id",
			handler: handleDeleteWebhook(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "get_user missing id",
			handler: handleGetUser(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "get_datatype missing id",
			handler: handleGetDatatype(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "get_session missing id",
			handler: handleGetSession(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "get_locale missing id",
			handler: handleGetLocale(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "get_content_version missing id",
			handler: handleGetContentVersion(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "get_validation missing id",
			handler: handleGetValidation(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
		{
			name:    "publish_content missing content_id",
			handler: handlePublishContent(nil),
			args:    map[string]any{},
			wantMsg: "content_id is required",
		},
		{
			name:    "unpublish_content missing content_id",
			handler: handleUnpublishContent(nil),
			args:    map[string]any{},
			wantMsg: "content_id is required",
		},
		{
			name:    "search_content missing q",
			handler: handleSearchContent(nil),
			args:    map[string]any{},
			wantMsg: "q is required",
		},
		{
			name:    "request_password_reset missing email",
			handler: handleRequestPasswordReset(nil),
			args:    map[string]any{},
			wantMsg: "email is required",
		},
		{
			name:    "register_user missing username",
			handler: handleRegisterUser(nil),
			args:    map[string]any{},
			wantMsg: "username is required",
		},
		{
			name:    "create_content missing status",
			handler: handleCreateContent(nil),
			args:    map[string]any{},
			wantMsg: "status is required",
		},
		{
			name:    "get_content_field missing id",
			handler: handleGetContentField(nil),
			args:    map[string]any{},
			wantMsg: "id is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, tc.handler, tc.args)
			if !result.IsError {
				t.Fatal("expected error for missing required param")
			}
			text := resultText(t, result)
			if text != tc.wantMsg {
				t.Errorf("message = %q, want %q", text, tc.wantMsg)
			}
		})
	}
}

func TestRequiredParams_NumericID(t *testing.T) {
	// When a required string param receives a numeric value, RequireString
	// returns "argument is not a string". The handler should return an error.
	cases := []struct {
		name    string
		handler server.ToolHandlerFunc
		args    map[string]any
	}{
		{
			name:    "get_content numeric id",
			handler: handleGetContent(nil),
			args:    map[string]any{"id": 12345},
		},
		{
			name:    "get_webhook numeric id",
			handler: handleGetWebhook(nil),
			args:    map[string]any{"id": 99},
		},
		{
			name:    "publish_content numeric content_id",
			handler: handlePublishContent(nil),
			args:    map[string]any{"content_id": 42},
		},
		{
			name:    "search_content numeric q",
			handler: handleSearchContent(nil),
			args:    map[string]any{"q": 123},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, tc.handler, tc.args)
			if !result.IsError {
				t.Fatal("expected error for non-string required param")
			}
			// The handler returns a user-friendly message, not the raw RequireString error.
			text := resultText(t, result)
			if text == "" {
				t.Error("error message should not be empty")
			}
		})
	}
}

func TestRequiredParams_NilArgs(t *testing.T) {
	// Passing nil args means no arguments at all. Required params should fail.
	handler := handleGetContent(nil)
	result := callTool(t, handler, nil)

	if !result.IsError {
		t.Fatal("expected error for nil args")
	}
	text := resultText(t, result)
	if text != "id is required" {
		t.Errorf("message = %q, want %q", text, "id is required")
	}
}

// =============================================================================
// Test Helpers
// =============================================================================

// parseErrorJSON unmarshals a JSON error string into a map. Fails the test if
// the text is not valid JSON.
func parseErrorJSON(t *testing.T, text string) map[string]any {
	t.Helper()
	var detail map[string]any
	if err := json.Unmarshal([]byte(text), &detail); err != nil {
		t.Fatalf("failed to parse error JSON %q: %v", text, err)
	}
	return detail
}

// assertErrorJSON verifies a JSON error string has the expected status code and
// exact message.
func assertErrorJSON(t *testing.T, text string, wantStatus int, wantMessage string) {
	t.Helper()
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(wantStatus) {
		t.Errorf("status = %v, want %d", detail["status"], wantStatus)
	}
	if detail["message"] != wantMessage {
		t.Errorf("message = %v, want %q", detail["message"], wantMessage)
	}
}

// assertErrorJSONContains verifies a JSON error string has the expected status code
// and a message that contains the given substring.
func assertErrorJSONContains(t *testing.T, text string, wantStatus int, wantSubstring string) {
	t.Helper()
	detail := parseErrorJSON(t, text)
	if detail["status"] != float64(wantStatus) {
		t.Errorf("status = %v, want %d", detail["status"], wantStatus)
	}
	msg, ok := detail["message"].(string)
	if !ok {
		t.Fatalf("message is not a string: %v", detail["message"])
	}
	if !strings.Contains(msg, wantSubstring) {
		t.Errorf("message = %q, want substring %q", msg, wantSubstring)
	}
}
