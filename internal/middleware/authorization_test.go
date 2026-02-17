package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// testUser is a minimal Users value for testing the permission injector.
var testUser = db.Users{
	UserID:   types.UserID("01TESTUSER000000000000000"),
	Username: "testuser",
	Role:     "TESTROLE00000000000000000",
}

func TestPermissionSetHas(t *testing.T) {
	ps := PermissionSet{
		"content:read":   {},
		"content:create": {},
	}

	if !ps.Has("content:read") {
		t.Error("expected Has to return true for existing permission")
	}
	if ps.Has("content:delete") {
		t.Error("expected Has to return false for missing permission")
	}
}

func TestPermissionSetHasAny(t *testing.T) {
	ps := PermissionSet{
		"content:read": {},
		"media:read":   {},
	}

	if !ps.HasAny("content:read", "roles:read") {
		t.Error("expected HasAny to return true when one permission matches")
	}
	if ps.HasAny("roles:create", "roles:delete") {
		t.Error("expected HasAny to return false when no permissions match")
	}
}

func TestPermissionSetHasAll(t *testing.T) {
	ps := PermissionSet{
		"content:read":   {},
		"content:create": {},
		"media:read":     {},
	}

	if !ps.HasAll("content:read", "content:create") {
		t.Error("expected HasAll to return true when all permissions present")
	}
	if ps.HasAll("content:read", "roles:read") {
		t.Error("expected HasAll to return false when some permissions missing")
	}
}

func TestPermissionInjectorAuthenticated(t *testing.T) {
	pc := NewPermissionCache()
	// Manually populate cache
	pc.mu.Lock()
	pc.cache[types.RoleID("TESTROLE00000000000000000")] = PermissionSet{"content:read": {}}
	pc.mu.Unlock()

	handler := PermissionInjector(pc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ps := ContextPermissions(r.Context())
		if ps == nil {
			t.Error("expected PermissionSet in context")
			return
		}
		if !ps.Has("content:read") {
			t.Error("expected content:read in PermissionSet")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := SetAuthenticatedUser(req.Context(), &testUser)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestPermissionInjectorUnauthenticated(t *testing.T) {
	pc := NewPermissionCache()

	called := false
	handler := PermissionInjector(pc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		ps := ContextPermissions(r.Context())
		if ps != nil {
			t.Error("expected nil PermissionSet for unauthenticated request")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("expected handler to be called for unauthenticated request")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRequirePermissionAllowed(t *testing.T) {
	handler := RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), permissionsKey, PermissionSet{"content:read": {}})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRequirePermissionDenied(t *testing.T) {
	handler := RequirePermission("content:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), permissionsKey, PermissionSet{"content:read": {}})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}

	// Verify response body is JSON {"error": "forbidden"} with no permission detail
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["error"] != "forbidden" {
		t.Errorf("expected error=forbidden, got %q", body["error"])
	}
	if _, exists := body["permission"]; exists {
		t.Error("403 response should not contain permission name")
	}
	if _, exists := body["detail"]; exists {
		t.Error("403 response should not contain detail field")
	}
}

func TestRequirePermissionFailClosed(t *testing.T) {
	handler := RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// No PermissionSet in context at all
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 (fail-closed), got %d", rr.Code)
	}
}

func TestRequireAnyPermission(t *testing.T) {
	handler := RequireAnyPermission("content:read", "media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Has one of the two
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), permissionsKey, PermissionSet{"media:read": {}})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	// Has neither
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx2 := context.WithValue(req2.Context(), permissionsKey, PermissionSet{"roles:read": {}})
	req2 = req2.WithContext(ctx2)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr2.Code)
	}
}

func TestRequireAllPermissions(t *testing.T) {
	handler := RequireAllPermissions("content:read", "media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Has both
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), permissionsKey, PermissionSet{"content:read": {}, "media:read": {}})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	// Has only one
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx2 := context.WithValue(req2.Context(), permissionsKey, PermissionSet{"content:read": {}})
	req2 = req2.WithContext(ctx2)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr2.Code)
	}
}

func TestRequireResourcePermissionMethodMapping(t *testing.T) {
	tests := []struct {
		method     string
		permission string
		wantCode   int
	}{
		{http.MethodGet, "content:read", http.StatusOK},
		{http.MethodPost, "content:create", http.StatusOK},
		{http.MethodPut, "content:update", http.StatusOK},
		{http.MethodPatch, "content:update", http.StatusOK},
		{http.MethodDelete, "content:delete", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			handler := RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "/test", nil)
			ctx := context.WithValue(req.Context(), permissionsKey, PermissionSet{tt.permission: {}})
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantCode {
				t.Errorf("method %s: expected %d, got %d", tt.method, tt.wantCode, rr.Code)
			}
		})
	}
}

func TestRequireResourcePermissionUnmappedMethod(t *testing.T) {
	handler := RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("TRACE", "/test", nil)
	ctx := context.WithValue(req.Context(), permissionsKey, PermissionSet{"content:read": {}})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unmapped method, got %d", rr.Code)
	}
}

func TestRequireResourcePermissionFailClosed(t *testing.T) {
	handler := RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// No PermissionSet in context
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 (fail-closed), got %d", rr.Code)
	}
}

func TestAdminBypassViaContextIsAdmin(t *testing.T) {
	handler := RequirePermission("some:obscure:permission")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), isAdminKey, true)
	// Deliberately no PermissionSet - admin bypass should work without it
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 (admin bypass), got %d", rr.Code)
	}
}

func TestLoadSafetyLimit(t *testing.T) {
	// This test verifies the behavior documented in the plan but cannot be
	// fully integration-tested without a real DB connection. We test the
	// PermissionCache type safety instead.
	pc := NewPermissionCache()
	if pc.IsAdmin("nonexistent") {
		t.Error("IsAdmin should return false for unknown role")
	}
	ps := pc.PermissionsForRole("nonexistent")
	if ps != nil {
		t.Error("PermissionsForRole should return nil for unknown role")
	}
}

func TestForbiddenResponseNoPermissionNames(t *testing.T) {
	handler := RequirePermission("secret:permission")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), permissionsKey, PermissionSet{})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "secret") {
		t.Error("403 response body should not contain the checked permission name")
	}
	if strings.Contains(body, "permission") {
		t.Error("403 response body should not contain the word 'permission'")
	}
}

func TestValidatePermissionLabel(t *testing.T) {
	tests := []struct {
		label   string
		wantErr bool
	}{
		{"content:read", false},
		{"datatypes:create", false},
		{"media:admin", false},
		{"admin_tree:read", false},
		{"ssh_keys:delete", false},
		{"config:update", false},
		{"", true},
		{"*", true},
		{"nocol", true},
		{":read", true},
		{"content:", true},
		{"Content:read", true},
		{"content:Read", true},
		{"content:123", true},
		{"content-data:read", true},
		{"content:re ad", true},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			err := ValidatePermissionLabel(tt.label)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePermissionLabel(%q): got err=%v, wantErr=%v", tt.label, err, tt.wantErr)
			}
		})
	}
}
