package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRFMiddleware_GET_SetsCookie(t *testing.T) {
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Token should be in context
		token, _ := r.Context().Value(CSRFContextKey{}).(string)
		if token == "" {
			t.Error("expected CSRF token in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Check cookie was set
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			found = true
			if c.Value == "" {
				t.Error("csrf_token cookie should have a value")
			}
			if c.HttpOnly {
				t.Error("csrf_token cookie should not be HttpOnly (JS needs to read it)")
			}
		}
	}
	if !found {
		t.Error("expected csrf_token cookie to be set")
	}
}

func TestCSRFMiddleware_POST_RejectsMissingToken(t *testing.T) {
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when CSRF token is missing")
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestCSRFMiddleware_POST_RejectsMismatchedToken(t *testing.T) {
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with mismatched CSRF token")
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "cookie-token"})
	req.Header.Set("X-CSRF-Token", "different-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestCSRFMiddleware_POST_AcceptsValidHeaderToken(t *testing.T) {
	called := false
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})
	req.Header.Set("X-CSRF-Token", "valid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should have been called with valid CSRF token")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCSRFMiddleware_POST_AcceptsValidFormToken(t *testing.T) {
	called := false
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	form := "_csrf=valid-token&name=test"
	req := httptest.NewRequest(http.MethodPost, "/admin/test", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should have been called with valid form CSRF token")
	}
}

func TestCSRFMiddleware_GET_AllowsWithoutToken(t *testing.T) {
	called := false
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("GET request should not require CSRF token")
	}
}

func TestCSRFMiddleware_POST_RejectsEmptyHeaderToken(t *testing.T) {
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with empty CSRF header token")
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})
	req.Header.Set("X-CSRF-Token", "")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Empty header + no form field = empty requestToken, should be rejected
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestCSRFMiddleware_POST_StoresTokenInContext(t *testing.T) {
	var contextToken string
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextToken, _ = r.Context().Value(CSRFContextKey{}).(string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "stored-token"})
	req.Header.Set("X-CSRF-Token", "stored-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if contextToken != "stored-token" {
		t.Errorf("expected context token %q, got %q", "stored-token", contextToken)
	}
}

func TestCSRFMiddleware_HEAD_GeneratesToken(t *testing.T) {
	called := false
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		token, _ := r.Context().Value(CSRFContextKey{}).(string)
		if token == "" {
			t.Error("expected CSRF token in context for HEAD request")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodHead, "/admin/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("HEAD request should pass through without CSRF validation")
	}
}

func TestCSRFMiddleware_OPTIONS_GeneratesToken(t *testing.T) {
	called := false
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/admin/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("OPTIONS request should pass through without CSRF validation")
	}
}

func TestCSRFMiddleware_PUT_RequiresToken(t *testing.T) {
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for PUT without CSRF token")
	}))

	req := httptest.NewRequest(http.MethodPut, "/admin/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestCSRFMiddleware_DELETE_RequiresToken(t *testing.T) {
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for DELETE without CSRF token")
	}))

	req := httptest.NewRequest(http.MethodDelete, "/admin/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestCSRFMiddleware_GET_CookiePath(t *testing.T) {
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			if c.Path != "/admin/" {
				t.Errorf("expected cookie path /admin/, got %s", c.Path)
			}
			if c.SameSite != http.SameSiteStrictMode {
				t.Errorf("expected SameSiteStrictMode, got %v", c.SameSite)
			}
			return
		}
	}
	t.Error("csrf_token cookie not found")
}

func TestCSRFMiddleware_GET_TokenUniqueness(t *testing.T) {
	var tokens []string
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _ := r.Context().Value(CSRFContextKey{}).(string)
		tokens = append(tokens, token)
		w.WriteHeader(http.StatusOK)
	}))

	for range 5 {
		req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	if len(tokens) != 5 {
		t.Fatalf("expected 5 tokens, got %d", len(tokens))
	}
	seen := make(map[string]bool)
	for _, tok := range tokens {
		if seen[tok] {
			t.Errorf("duplicate CSRF token generated: %s", tok)
		}
		seen[tok] = true
	}
}

func TestCSRFContextKey_RoundTrips(t *testing.T) {
	token := "test-token-12345"
	ctx := context.WithValue(context.Background(), CSRFContextKey{}, token)
	got, ok := ctx.Value(CSRFContextKey{}).(string)
	if !ok || got != token {
		t.Errorf("expected %q, got %q", token, got)
	}
}
