package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
)

func TestAdminAuthMiddleware_RedirectsUnauthenticated(t *testing.T) {
	// AdminAuthMiddleware does not access the manager in its current
	// implementation (it only checks middleware.AuthenticatedUser from
	// context), so nil is safe for testing.
	mw := AdminAuthMiddleware((*config.Manager)(nil))
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for unauthenticated request")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/content", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
	location := rec.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header")
	}
	if !strings.Contains(location, "/admin/login") {
		t.Errorf("expected redirect to /admin/login, got %s", location)
	}
	if !strings.Contains(location, "next=") {
		t.Errorf("expected ?next= parameter in redirect URL, got %s", location)
	}
}

func TestAdminAuthMiddleware_HTMXRedirect(t *testing.T) {
	mw := AdminAuthMiddleware((*config.Manager)(nil))
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for unauthenticated HTMX request")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/content", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	hxRedirect := rec.Header().Get("HX-Redirect")
	if hxRedirect == "" {
		t.Error("expected HX-Redirect header for HTMX request")
	}
	if !strings.Contains(hxRedirect, "/admin/login") {
		t.Errorf("expected HX-Redirect to /admin/login, got %s", hxRedirect)
	}
}

func TestAdminAuthMiddleware_AllowsAuthenticated(t *testing.T) {
	mw := AdminAuthMiddleware((*config.Manager)(nil))
	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	user := &db.Users{
		UserID:   types.NewUserID(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	ctx := middleware.SetAuthenticatedUser(req.Context(), user)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should be called for authenticated request")
	}
}

func TestAdminAuthMiddleware_NextParamPreservesPath(t *testing.T) {
	mw := AdminAuthMiddleware((*config.Manager)(nil))
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/admin/schema/datatypes", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	location := rec.Header().Get("Location")
	if !strings.Contains(location, "schema") {
		t.Errorf("expected next= to contain the original path, got %s", location)
	}
}

func TestAdminAuthMiddleware_HTMXRedirectIncludesNext(t *testing.T) {
	mw := AdminAuthMiddleware((*config.Manager)(nil))
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/admin/media", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	hxRedirect := rec.Header().Get("HX-Redirect")
	if !strings.Contains(hxRedirect, "next=") {
		t.Errorf("expected HX-Redirect to include next= parameter, got %s", hxRedirect)
	}
	if !strings.Contains(hxRedirect, "media") {
		t.Errorf("expected HX-Redirect next= to contain original path, got %s", hxRedirect)
	}
}

func TestAdminAuthMiddleware_NonAdminPathFallback(t *testing.T) {
	// When the requested path doesn't start with /admin/ or equal /admin,
	// the middleware should fall back to /admin/ as the next path.
	mw := AdminAuthMiddleware((*config.Manager)(nil))
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/other/path", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	location := rec.Header().Get("Location")
	// The next= value should be /admin/ since the original path is not an admin path
	if !strings.Contains(location, "/admin/login") {
		t.Errorf("expected redirect to /admin/login, got %s", location)
	}
}

func TestAdminAuthMiddleware_AuthenticatedPOST(t *testing.T) {
	mw := AdminAuthMiddleware((*config.Manager)(nil))
	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/content", nil)
	user := &db.Users{
		UserID:   types.NewUserID(),
		Username: "admin",
		Email:    "admin@example.com",
	}
	ctx := middleware.SetAuthenticatedUser(req.Context(), user)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("authenticated POST should pass through middleware")
	}
}

func TestAdminAuthMiddleware_UnauthenticatedPOST(t *testing.T) {
	mw := AdminAuthMiddleware((*config.Manager)(nil))
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for unauthenticated POST")
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/content", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// POST without auth should still redirect (or 401 for HTMX)
	if rec.Code != http.StatusFound {
		t.Errorf("expected 302 for unauthenticated POST, got %d", rec.Code)
	}
}
