package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/hegner123/modulacms/internal/admin"
	"github.com/hegner123/modulacms/internal/admin/layouts"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// Render writes a templ component to the response.
// Buffer-first: if rendering fails, no bytes have been sent and we can return 500.
func Render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	buf := new(bytes.Buffer)
	if err := component.Render(r.Context(), buf); err != nil {
		utility.DefaultLogger.Error("render failed", err)
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Retarget", "#none")
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Render error", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// OOBSwap represents an out-of-band HTMX swap fragment.
// By default uses outerHTML swap. Set InnerHTML to true to preserve the
// target element and only replace its children.
type OOBSwap struct {
	TargetID  string
	Component templ.Component
	InnerHTML bool
}

// RenderWithOOB renders a primary component plus out-of-band swap fragments.
// All components are buffered before writing for atomic delivery.
func RenderWithOOB(w http.ResponseWriter, r *http.Request, primary templ.Component, oob ...OOBSwap) {
	buf := new(bytes.Buffer)

	if err := primary.Render(r.Context(), buf); err != nil {
		utility.DefaultLogger.Error("primary render failed", err)
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Retarget", "#none")
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Render error", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, swap := range oob {
		swapMode := "true"
		if swap.InnerHTML {
			swapMode = "innerHTML"
		}
		fmt.Fprintf(buf, `<div id="%s" hx-swap-oob="%s">`, template.HTMLEscapeString(swap.TargetID), swapMode)
		if err := swap.Component.Render(r.Context(), buf); err != nil {
			utility.DefaultLogger.Error("OOB render failed", err, "target", swap.TargetID)
			if r.Header.Get("HX-Request") != "" {
				w.Header().Set("HX-Retarget", "#none")
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Render error", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(buf, `</div>`)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// CSRFTokenFromContext retrieves the CSRF token from context.
func CSRFTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(admin.CSRFContextKey{}).(string)
	return token
}

// NewAdminData builds the common layout data from the request context.
func NewAdminData(r *http.Request, title string) layouts.AdminData {
	user := middleware.AuthenticatedUser(r.Context())
	return layouts.AdminData{
		Title:       title,
		CurrentPath: r.URL.Path,
		User:        user,
		Permissions: middleware.ContextPermissions(r.Context()),
		IsAdmin:     middleware.ContextIsAdmin(r.Context()),
		CSRFToken:   CSRFTokenFromContext(r.Context()),
		Version:     utility.Version,
	}
}

// PaginationData holds pagination state for templates.
type PaginationData struct {
	Current    int
	TotalPages int
	Total      int64
	Limit      int64
	Target     string
	BaseURL    string
}

// Pages returns the list of page numbers for pagination.
func (p PaginationData) Pages() []int {
	pages := make([]int, 0, p.TotalPages)
	for i := 1; i <= p.TotalPages; i++ {
		pages = append(pages, i)
	}
	return pages
}

// URLForPage returns the URL for a specific page number.
func (p PaginationData) URLForPage(page int) string {
	offset := int64((page - 1)) * p.Limit
	sep := "?"
	if strings.Contains(p.BaseURL, "?") {
		sep = "&"
	}
	return p.BaseURL + sep + "limit=" + strconv.FormatInt(p.Limit, 10) + "&offset=" + strconv.FormatInt(offset, 10)
}

// NewPaginationData creates pagination data from total count and request params.
func NewPaginationData(total int64, limit int64, offset int64, target string, baseURL string) PaginationData {
	if limit <= 0 {
		limit = 50
	}
	totalPages := int((total + limit - 1) / limit)
	current := int(offset/limit) + 1
	return PaginationData{
		Current:    current,
		TotalPages: totalPages,
		Total:      total,
		Limit:      limit,
		Target:     target,
		BaseURL:    baseURL,
	}
}

// ParsePagination parses limit and offset from request query params.
func ParsePagination(r *http.Request) (limit int64, offset int64) {
	limit = 50
	offset = 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return
}

// IsHTMX returns true if the request is an HTMX request.
func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") != ""
}

// IsNavHTMX returns true if the request is an HTMX navigation request
// targeting #main-content (sidebar nav), as opposed to pagination or
// other partial HTMX swaps.
func IsNavHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") != "" && r.Header.Get("HX-Target") == "main-content"
}

// RenderNav handles the HTMX navigation vs full-page branching.
// For sidebar HTMX requests, it renders the content-only component and
// sets an HX-Trigger header so the client updates the page title.
// For direct URL access (no HTMX), it renders the full page with layout.
func RenderNav(w http.ResponseWriter, r *http.Request, title string, content, fullPage templ.Component) {
	if IsNavHTMX(r) {
		safeTitle := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(title)
		w.Header().Set("HX-Trigger", `{"pageTitle": "`+safeTitle+`"}`)
		Render(w, r, content)
		return
	}
	Render(w, r, fullPage)
}

// ResolveUserName looks up a user by ID and returns a display name.
// Prefers Name, falls back to Username, then Email. Returns "" if not found.
func ResolveUserName(driver db.DbDriver, userID types.UserID) string {
	if userID.IsZero() {
		return ""
	}
	user, err := driver.GetUser(userID)
	if err != nil || user == nil {
		return ""
	}
	if user.Name != "" {
		return user.Name
	}
	if user.Username != "" {
		return user.Username
	}
	return string(user.Email)
}

// ResolveNullableUserName is like ResolveUserName but for NullableUserID fields.
func ResolveNullableUserName(driver db.DbDriver, userID types.NullableUserID) string {
	if !userID.Valid || userID.ID.IsZero() {
		return ""
	}
	return ResolveUserName(driver, userID.ID)
}

// BuildUserNameMap resolves a set of NullableUserIDs to display names.
// Deduplicates lookups so each user is fetched at most once.
func BuildUserNameMap(driver db.DbDriver, userIDs []types.NullableUserID) map[string]string {
	names := make(map[string]string, len(userIDs))
	for _, uid := range userIDs {
		idStr := uid.String()
		if idStr == "" {
			continue
		}
		if _, seen := names[idStr]; seen {
			continue
		}
		names[idStr] = ResolveNullableUserName(driver, uid)
	}
	return names
}

// HasPermission checks if the current request has a specific permission.
func HasPermission(r *http.Request, perm string) bool {
	if middleware.ContextIsAdmin(r.Context()) {
		return true
	}
	return middleware.ContextPermissions(r.Context()).Has(perm)
}
