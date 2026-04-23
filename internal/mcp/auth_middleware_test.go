package mcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
)

// passthrough is a ToolHandlerFunc that returns a success result.
func passthrough(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ok"), nil
}

func buildReq(toolName string) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: toolName,
		},
	}
}

func TestPermissionMiddleware_PublicTool(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	// Public tools should be allowed without any auth context.
	result, err := handler(context.Background(), buildReq("register_user"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success for public tool, got error: %s", resultText(t, result))
	}
}

func TestPermissionMiddleware_Unauthenticated(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	// A tool that requires auth but no user in context.
	result, err := handler(context.Background(), buildReq("list_content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for unauthenticated request")
	}
	text := resultText(t, result)
	if text != errAuthRequired {
		t.Errorf("error = %q, want %q", text, errAuthRequired)
	}
}

func TestPermissionMiddleware_AdminBypass(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	// Admin user bypasses permission checks.
	ctx := middleware.SetAuthenticatedUser(context.Background(), &db.Users{
		UserID: types.UserID("admin-user"),
	})
	ctx = middleware.SetIsAdmin(ctx, true)

	result, err := handler(ctx, buildReq("delete_user"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("admin should bypass permission check, got error: %s", resultText(t, result))
	}
}

func TestPermissionMiddleware_Allowed(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	// User with the required permission.
	ctx := middleware.SetAuthenticatedUser(context.Background(), &db.Users{
		UserID: types.UserID("editor-user"),
	})
	ctx = middleware.SetIsAdmin(ctx, false)
	ctx = middleware.SetPermissions(ctx, middleware.PermissionSet{
		"content:read": {},
	})

	result, err := handler(ctx, buildReq("list_content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success for authorized user, got error: %s", resultText(t, result))
	}
}

func TestPermissionMiddleware_Denied(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	// User without the required permission.
	ctx := middleware.SetAuthenticatedUser(context.Background(), &db.Users{
		UserID: types.UserID("viewer-user"),
	})
	ctx = middleware.SetIsAdmin(ctx, false)
	ctx = middleware.SetPermissions(ctx, middleware.PermissionSet{
		"content:read": {},
	})

	result, err := handler(ctx, buildReq("delete_content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for denied permission")
	}
	text := resultText(t, result)
	expected := "forbidden: requires permission 'content:delete'"
	if text != expected {
		t.Errorf("error = %q, want %q", text, expected)
	}
}

func TestAuditContextFromMCP_WithUser(t *testing.T) {
	ctx := middleware.SetAuthenticatedUser(context.Background(), &db.Users{
		UserID: types.UserID("real-user"),
	})
	ac := AuditContextFromMCP(ctx)
	if ac.UserID != types.UserID("real-user") {
		t.Errorf("UserID = %q, want %q", ac.UserID, "real-user")
	}
}

func TestAuditContextFromMCP_StdioFallback(t *testing.T) {
	fallbackAC := audited.Ctx(types.NewNodeID(), types.UserID("mcp-local"), "", "127.0.0.1")
	ctx := context.WithValue(context.Background(), mcpAuditKey{}, fallbackAC)

	ac := AuditContextFromMCP(ctx)
	if ac.UserID != types.UserID("mcp-local") {
		t.Errorf("UserID = %q, want %q", ac.UserID, "mcp-local")
	}
}

func TestAuditContextFromMCP_Anonymous(t *testing.T) {
	ac := AuditContextFromMCP(context.Background())
	if ac.UserID != types.UserID("mcp-anonymous") {
		t.Errorf("UserID = %q, want %q", ac.UserID, "mcp-anonymous")
	}
}

// ---------------------------------------------------------------------------
// Adversarial tests: attack assumptions in the permission middleware
// ---------------------------------------------------------------------------

// TestPermissionMiddleware_UnknownToolPassthrough verifies that a tool name
// absent from both toolPermissions and publicTools is allowed through without
// any auth check. This is the documented behavior for connection tools, but
// it means any unregistered tool name bypasses permission enforcement entirely.
// If this behavior ever changes (e.g., fail-closed for unknown tools), update
// this test to expect an error instead.
func TestPermissionMiddleware_UnknownToolPassthrough(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	// No user in context at all. An unknown tool should still pass.
	result, err := handler(context.Background(), buildReq("completely_nonexistent_tool"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unknown tool should pass through, got error: %s", resultText(t, result))
	}
}

// TestPermissionMiddleware_EmptyPermissionSet verifies that an authenticated
// non-admin user with zero permissions is denied access to every guarded tool.
func TestPermissionMiddleware_EmptyPermissionSet(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	ctx := middleware.SetAuthenticatedUser(context.Background(), &db.Users{
		UserID: types.UserID("empty-perms-user"),
	})
	ctx = middleware.SetIsAdmin(ctx, false)
	ctx = middleware.SetPermissions(ctx, middleware.PermissionSet{})

	// Try several tools from different categories.
	tools := []string{
		"list_content",
		"create_user",
		"delete_webhook",
		"update_config",
		"list_recent_activity",
	}

	for _, toolName := range tools {
		t.Run(toolName, func(t *testing.T) {
			result, err := handler(ctx, buildReq(toolName))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.IsError {
				t.Errorf("user with empty permission set should be denied %q", toolName)
			}
		})
	}
}

// TestPermissionMiddleware_CaseSensitivity verifies that permission matching
// is case-sensitive. A user with "Content:Read" should NOT pass a check for
// "content:read". PermissionSet is map[string]struct{}, so lookup is exact.
func TestPermissionMiddleware_CaseSensitivity(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	ctx := middleware.SetAuthenticatedUser(context.Background(), &db.Users{
		UserID: types.UserID("case-test-user"),
	})
	ctx = middleware.SetIsAdmin(ctx, false)

	cases := []struct {
		name      string
		grantPerm string
		tool      string
		wantError bool
	}{
		{
			name:      "wrong case is denied",
			grantPerm: "Content:Read",
			tool:      "list_content",
			wantError: true,
		},
		{
			name:      "exact case is allowed",
			grantPerm: "content:read",
			tool:      "list_content",
			wantError: false,
		},
		{
			name:      "all caps is denied",
			grantPerm: "CONTENT:READ",
			tool:      "list_content",
			wantError: true,
		},
		{
			name:      "trailing space is denied",
			grantPerm: "content:read ",
			tool:      "list_content",
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caseCtx := middleware.SetPermissions(ctx, middleware.PermissionSet{
				tc.grantPerm: {},
			})
			result, err := handler(caseCtx, buildReq(tc.tool))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantError && !result.IsError {
				t.Errorf("expected denial with permission %q for tool %q", tc.grantPerm, tc.tool)
			}
			if !tc.wantError && result.IsError {
				t.Errorf("expected success with permission %q for tool %q, got: %s", tc.grantPerm, tc.tool, resultText(t, result))
			}
		})
	}
}

// TestPermissionMiddleware_ConcurrentAccess verifies that PermissionMiddleware
// is safe for concurrent use. The toolPermissions and publicTools maps are
// read-only after init(), so concurrent reads should never race.
func TestPermissionMiddleware_ConcurrentAccess(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	ctx := middleware.SetAuthenticatedUser(context.Background(), &db.Users{
		UserID: types.UserID("concurrent-user"),
	})
	ctx = middleware.SetIsAdmin(ctx, true) // admin bypass, so all calls succeed

	const goroutines int32 = 50
	errs := make(chan error, goroutines)

	for i := range int32(goroutines) {
		go func(idx int32) {
			// Alternate between different tool categories.
			tools := []string{"list_content", "register_user", "delete_webhook", "health", "unknown_tool"}
			toolName := tools[idx%int32(len(tools))]

			result, err := handler(ctx, buildReq(toolName))
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: unexpected error for %q: %v", idx, toolName, err)
				return
			}
			if result.IsError {
				errs <- fmt.Errorf("goroutine %d: unexpected tool error for %q: %s", idx, toolName, result.Content)
				return
			}
			errs <- nil
		}(i)
	}

	for range int32(goroutines) {
		if err := <-errs; err != nil {
			t.Error(err)
		}
	}
}

// TestPermissionMiddleware_MalformedToolNames verifies behavior when the tool
// name is empty, contains whitespace, or uses special characters. None of
// these appear in toolPermissions or publicTools, so they should all fall
// through as unknown tools (allowed without auth, per current behavior).
func TestPermissionMiddleware_MalformedToolNames(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	malformed := []struct {
		name     string
		toolName string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab character", "\t"},
		{"newline", "list\ncontent"},
		{"null byte", "list\x00content"},
		{"unicode", "list_\u200bcontent"}, // zero-width space
		{"sql injection attempt", "'; DROP TABLE permissions; --"},
		{"path traversal", "../../../etc/passwd"},
	}

	for _, tc := range malformed {
		t.Run(tc.name, func(t *testing.T) {
			// No user context. Unknown tools pass through without auth.
			result, err := handler(context.Background(), buildReq(tc.toolName))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Current behavior: unknown tools pass through. If this changes
			// to fail-closed, flip this assertion.
			if result.IsError {
				t.Errorf("malformed tool name %q was rejected, expected passthrough", tc.toolName)
			}
		})
	}
}

// TestPermissionMiddleware_NilPermissionSetWithUser verifies that an
// authenticated non-admin user with a nil PermissionSet (not empty, nil)
// is denied. This tests the fail-closed path: if the permission injector
// middleware didn't run or failed silently, ps will be nil.
func TestPermissionMiddleware_NilPermissionSetWithUser(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	// User present, admin=false, but no SetPermissions call (ps = nil).
	ctx := middleware.SetAuthenticatedUser(context.Background(), &db.Users{
		UserID: types.UserID("nil-perms-user"),
	})
	ctx = middleware.SetIsAdmin(ctx, false)

	result, err := handler(ctx, buildReq("list_content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("nil PermissionSet should be denied (fail-closed)")
	}
	text := resultText(t, result)
	if text != errAuthRequired {
		t.Errorf("error = %q, want %q", text, errAuthRequired)
	}
}

// TestPermissionMiddleware_WrongPermissionForTool verifies that having a
// permission from a different resource does not grant access to unrelated
// tools. Ensures there is no wildcard or prefix matching.
func TestPermissionMiddleware_WrongPermissionForTool(t *testing.T) {
	mw := PermissionMiddleware()
	handler := mw(passthrough)

	ctx := middleware.SetAuthenticatedUser(context.Background(), &db.Users{
		UserID: types.UserID("wrong-perm-user"),
	})
	ctx = middleware.SetIsAdmin(ctx, false)

	cases := []struct {
		name      string
		granted   string
		tool      string
	}{
		{"media perm vs content tool", "media:read", "list_content"},
		{"content:read vs content:delete tool", "content:read", "delete_content"},
		{"roles:read vs permissions tool", "roles:read", "list_permissions"},
		{"webhook:read vs locale tool", "webhook:read", "list_locales"},
		{"content:admin vs config tool", "content:admin", "update_config"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caseCtx := middleware.SetPermissions(ctx, middleware.PermissionSet{
				tc.granted: {},
			})
			result, err := handler(caseCtx, buildReq(tc.tool))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.IsError {
				t.Errorf("permission %q should not grant access to tool %q", tc.granted, tc.tool)
			}
		})
	}
}

func TestInjectAuditContextMiddleware(t *testing.T) {
	ac := audited.Ctx(types.NewNodeID(), types.UserID("stdio-user"), "req-123", "192.168.1.1")
	mw := injectAuditContextMiddleware(ac)

	handler := mw(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		extracted := AuditContextFromMCP(ctx)
		if extracted.UserID != types.UserID("stdio-user") {
			t.Errorf("UserID = %q, want %q", extracted.UserID, "stdio-user")
		}
		return mcp.NewToolResultText("ok"), nil
	})

	result, err := handler(context.Background(), buildReq("health"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", resultText(t, result))
	}
}
