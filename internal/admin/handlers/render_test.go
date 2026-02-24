package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/hegner123/modulacms/internal/admin"
)

// testComponent creates a simple templ component for testing.
func testComponent(content string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := fmt.Fprint(w, content)
		return err
	})
}

// errorComponent creates a component that always returns an error.
func errorComponent() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return fmt.Errorf("render error")
	})
}

func TestRender_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	component := testComponent("<h1>Hello</h1>")
	Render(rec, req, component)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %s", ct)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Hello") {
		t.Errorf("expected body to contain Hello, got %s", body)
	}
}

func TestRender_Error_NonHTMX(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	Render(rec, req, errorComponent())

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestRender_Error_HTMX(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	Render(rec, req, errorComponent())

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if retarget := rec.Header().Get("HX-Retarget"); retarget != "#none" {
		t.Errorf("expected HX-Retarget=#none, got %s", retarget)
	}
	if trigger := rec.Header().Get("HX-Trigger"); trigger == "" {
		t.Error("expected HX-Trigger header for HTMX error response")
	}
}

func TestRender_BufferFirst(t *testing.T) {
	// Verify that a render error does not send partial content.
	// The errorComponent writes nothing before failing, so the body
	// should contain only the error response, not any HTML fragment.
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	Render(rec, req, errorComponent())

	body := rec.Body.String()
	if strings.Contains(body, "<h1>") {
		t.Error("partial content should not be sent on render error")
	}
}

func TestRender_ContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	Render(rec, req, testComponent("<p>test</p>"))

	ct := rec.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html; charset=utf-8, got %s", ct)
	}
}

func TestRender_EmptyComponent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	Render(rec, req, testComponent(""))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRenderWithOOB_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	primary := testComponent("<main>Primary</main>")
	oob := OOBSwap{
		TargetID:  "pagination",
		Component: testComponent("<nav>Page 1</nav>"),
	}

	RenderWithOOB(rec, req, primary, oob)

	body := rec.Body.String()
	if !strings.Contains(body, "Primary") {
		t.Error("expected primary content in response")
	}
	if !strings.Contains(body, "pagination") {
		t.Error("expected OOB target ID in response")
	}
	if !strings.Contains(body, "hx-swap-oob") {
		t.Error("expected hx-swap-oob attribute in response")
	}
	if !strings.Contains(body, "Page 1") {
		t.Error("expected OOB content in response")
	}
}

func TestRenderWithOOB_PrimaryError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	RenderWithOOB(rec, req, errorComponent(), OOBSwap{
		TargetID:  "test",
		Component: testComponent("should not appear"),
	})

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestRenderWithOOB_OOBError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	RenderWithOOB(rec, req, testComponent("<main>OK</main>"), OOBSwap{
		TargetID:  "broken",
		Component: errorComponent(),
	})

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on OOB render error, got %d", rec.Code)
	}
}

func TestRenderWithOOB_MultipleOOB(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	primary := testComponent("<main>Content</main>")
	oob1 := OOBSwap{
		TargetID:  "sidebar",
		Component: testComponent("<aside>Sidebar</aside>"),
	}
	oob2 := OOBSwap{
		TargetID:  "footer",
		Component: testComponent("<footer>Footer</footer>"),
	}

	RenderWithOOB(rec, req, primary, oob1, oob2)

	body := rec.Body.String()
	if !strings.Contains(body, "Content") {
		t.Error("expected primary content")
	}
	if !strings.Contains(body, "sidebar") {
		t.Error("expected sidebar OOB target")
	}
	if !strings.Contains(body, "Sidebar") {
		t.Error("expected sidebar OOB content")
	}
	if !strings.Contains(body, "footer") {
		t.Error("expected footer OOB target")
	}
	if !strings.Contains(body, "Footer") {
		t.Error("expected footer OOB content")
	}
}

func TestRenderWithOOB_NoOOB(t *testing.T) {
	// RenderWithOOB with zero OOB swaps should behave like Render.
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	RenderWithOOB(rec, req, testComponent("<main>Only</main>"))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Only") {
		t.Error("expected primary content in response")
	}
	if strings.Contains(body, "hx-swap-oob") {
		t.Error("expected no OOB attributes when no swaps provided")
	}
}

func TestRenderWithOOB_HTMXError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	RenderWithOOB(rec, req, errorComponent())

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if retarget := rec.Header().Get("HX-Retarget"); retarget != "#none" {
		t.Errorf("expected HX-Retarget=#none for HTMX error, got %s", retarget)
	}
}

func TestRenderWithOOB_OOBWrappingStructure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	primary := testComponent("")
	oob := OOBSwap{
		TargetID:  "my-target",
		Component: testComponent("inner"),
	}

	RenderWithOOB(rec, req, primary, oob)

	body := rec.Body.String()
	// The OOB div should have id and hx-swap-oob="true"
	expected := `<div id="my-target" hx-swap-oob="true">inner</div>`
	if !strings.Contains(body, expected) {
		t.Errorf("expected OOB wrapper %q in body, got %q", expected, body)
	}
}

func TestCSRFTokenFromContext_ReturnsToken(t *testing.T) {
	token := "test-csrf-token"
	ctx := context.WithValue(context.Background(), admin.CSRFContextKey{}, token)
	got := CSRFTokenFromContext(ctx)
	if got != token {
		t.Errorf("expected %q, got %q", token, got)
	}
}

func TestCSRFTokenFromContext_ReturnsEmptyWhenMissing(t *testing.T) {
	ctx := context.Background()
	got := CSRFTokenFromContext(ctx)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestCSRFTokenFromContext_ReturnsEmptyForWrongType(t *testing.T) {
	// If the context value is not a string, the type assertion should
	// return the zero value (empty string) instead of panicking.
	ctx := context.WithValue(context.Background(), admin.CSRFContextKey{}, 12345)
	got := CSRFTokenFromContext(ctx)
	if got != "" {
		t.Errorf("expected empty string for wrong type, got %q", got)
	}
}

func TestIsHTMX_True(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	req.Header.Set("HX-Request", "true")
	if !IsHTMX(req) {
		t.Error("expected IsHTMX to return true when HX-Request header is set")
	}
}

func TestIsHTMX_False(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	if IsHTMX(req) {
		t.Error("expected IsHTMX to return false when HX-Request header is absent")
	}
}

func TestNewPaginationData(t *testing.T) {
	t.Run("basic pagination", func(t *testing.T) {
		pd := NewPaginationData(100, 10, 0, "#content", "/admin/content")
		if pd.Current != 1 {
			t.Errorf("expected current page 1, got %d", pd.Current)
		}
		if pd.TotalPages != 10 {
			t.Errorf("expected 10 total pages, got %d", pd.TotalPages)
		}
		if pd.Total != 100 {
			t.Errorf("expected total 100, got %d", pd.Total)
		}
	})

	t.Run("second page", func(t *testing.T) {
		pd := NewPaginationData(100, 10, 10, "#content", "/admin/content")
		if pd.Current != 2 {
			t.Errorf("expected current page 2, got %d", pd.Current)
		}
	})

	t.Run("zero limit defaults to 50", func(t *testing.T) {
		pd := NewPaginationData(200, 0, 0, "#content", "/admin/content")
		if pd.Limit != 50 {
			t.Errorf("expected limit to default to 50, got %d", pd.Limit)
		}
	})

	t.Run("partial last page", func(t *testing.T) {
		pd := NewPaginationData(15, 10, 0, "#content", "/admin/content")
		if pd.TotalPages != 2 {
			t.Errorf("expected 2 total pages for 15 items with limit 10, got %d", pd.TotalPages)
		}
	})
}

func TestPaginationData_Pages(t *testing.T) {
	pd := NewPaginationData(30, 10, 0, "#content", "/admin/content")
	pages := pd.Pages()
	if len(pages) != 3 {
		t.Fatalf("expected 3 pages, got %d", len(pages))
	}
	for i, p := range pages {
		if p != i+1 {
			t.Errorf("expected page %d at index %d, got %d", i+1, i, p)
		}
	}
}

func TestPaginationData_URLForPage(t *testing.T) {
	pd := NewPaginationData(100, 10, 0, "#content", "/admin/content")

	url1 := pd.URLForPage(1)
	if !strings.Contains(url1, "limit=10") {
		t.Errorf("expected limit=10 in URL, got %s", url1)
	}
	if !strings.Contains(url1, "offset=0") {
		t.Errorf("expected offset=0 for page 1, got %s", url1)
	}

	url2 := pd.URLForPage(2)
	if !strings.Contains(url2, "offset=10") {
		t.Errorf("expected offset=10 for page 2, got %s", url2)
	}

	url3 := pd.URLForPage(3)
	if !strings.Contains(url3, "offset=20") {
		t.Errorf("expected offset=20 for page 3, got %s", url3)
	}
}

func TestPaginationData_URLForPage_WithQueryParams(t *testing.T) {
	pd := NewPaginationData(100, 10, 0, "#content", "/admin/content?search=test")
	url := pd.URLForPage(1)
	// When base URL already has ?, separator should be &
	if !strings.Contains(url, "?search=test&limit=") {
		t.Errorf("expected & separator for existing query params, got %s", url)
	}
}

func TestParsePagination(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/content", nil)
		limit, offset := ParsePagination(req)
		if limit != 50 {
			t.Errorf("expected default limit 50, got %d", limit)
		}
		if offset != 0 {
			t.Errorf("expected default offset 0, got %d", offset)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/content?limit=25&offset=50", nil)
		limit, offset := ParsePagination(req)
		if limit != 25 {
			t.Errorf("expected limit 25, got %d", limit)
		}
		if offset != 50 {
			t.Errorf("expected offset 50, got %d", offset)
		}
	})

	t.Run("invalid limit uses default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/content?limit=abc", nil)
		limit, _ := ParsePagination(req)
		if limit != 50 {
			t.Errorf("expected default limit 50 for invalid input, got %d", limit)
		}
	})

	t.Run("negative limit uses default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/content?limit=-5", nil)
		limit, _ := ParsePagination(req)
		if limit != 50 {
			t.Errorf("expected default limit 50 for negative input, got %d", limit)
		}
	})

	t.Run("limit exceeding max uses default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/content?limit=5000", nil)
		limit, _ := ParsePagination(req)
		if limit != 50 {
			t.Errorf("expected default limit 50 for exceeding max, got %d", limit)
		}
	})

	t.Run("negative offset uses default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/content?offset=-10", nil)
		_, offset := ParsePagination(req)
		if offset != 0 {
			t.Errorf("expected default offset 0 for negative input, got %d", offset)
		}
	})
}
