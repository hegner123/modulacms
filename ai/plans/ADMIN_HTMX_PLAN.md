# Admin Panel HTMX + Web Components Refactor Plan

## Goals

Replace the React 19 SPA (`admin/`) with a server-rendered admin panel using templ components, HTMX for dynamic interactions, and Web Components for client-side widgets. The block editor (`<block-editor>`) is a standalone web component at `/Users/home/Documents/Code/blockEditor/` and will be copied in as-is.

## Architecture Overview

```
Browser                          Go Server
  |                                 |
  |  GET /admin/content             |
  |------------------------------->  |
  |  <full HTML page>               |  layouts.Admin(data) { pages.ContentList(items) }
  |  <------------------------------| (templ component renders to ResponseWriter)
  |                                 |
  |  hx-get /admin/content?page=2   |
  |------------------------------->  |
  |  <partial: table rows only>     |  partials.ContentTableRows(items)
  |  <------------------------------| (templ component renders fragment only)
  |                                 |
  |  <mcms-media-picker>            |  Web Component (Light DOM — HTMX-compatible)
  |  dispatches custom event        |  No server round-trip for open/close
  |  hx-get to load media grid      |  HTMX loads media items from server
  |                                 |
```

### Principles

1. **Server renders HTML** — templ components generate type-safe HTML at compile time
2. **HTMX handles interactions** — form submissions, pagination, search, delete confirmations, tab switching, live validation
3. **Web Components for stateful widgets** — media picker, tree navigator, toast system, dialogs, field renderers
4. **JSON API stays untouched** — the existing `/api/v1/` endpoints continue serving SDKs (TypeScript, Go, Swift), TUI, and external consumers. The `@modulacms/admin-sdk` remains published for third-party developers.
5. **New `/admin/` routes return HTML** — separate handler layer that calls the same `DbDriver` methods but renders templ components instead of JSON
6. **HTMX required for mutations** — Read-only pages (lists, details) work as plain HTML links. Create/update/delete operations require HTMX (no progressive enhancement for writes — admin panels require JS).
7. **Go-only build** — `templ generate` compiles `.templ` files to Go code. No Node.js, no bundler. Plain JS for Web Components. CSS via a single stylesheet.

## Package Structure

```
internal/
  admin/                          # NEW — admin panel package
    admin.go                      # RegisterAdminRoutes(mux, mgr, driver, pc)
    render.go                     # Render/RenderPartial/RenderMulti helpers
    csrf.go                       # CSRF token generation and validation (double-submit cookie)
    middleware.go                 # Admin-specific middleware (auth redirect, HTMX detection, CSRF)
    handlers/                     # One file per resource area
      dashboard.go
      content.go
      datatypes.go
      fields.go
      field_types.go
      media.go
      routes.go
      users.go
      roles.go
      tokens.go
      sessions.go
      plugins.go
      import.go
      audit.go
      settings.go
      auth.go                    # Login/logout (HTML form + redirect)

    layouts/                     # templ layout components
      base.templ                 # <!DOCTYPE>, <head>, HTMX script, CSS link
      admin.templ                # Sidebar + topbar + {children...}
      auth.templ                 # Centered card layout for login

    pages/                       # templ full-page components
      dashboard.templ
      content_list.templ
      content_edit.templ         # Field editing + block editor web component
      datatypes_list.templ
      datatype_detail.templ
      fields_list.templ
      field_detail.templ
      field_types_list.templ
      media_list.templ
      media_detail.templ
      routes_list.templ
      admin_routes_list.templ
      users_list.templ
      user_detail.templ
      roles_list.templ
      tokens_list.templ
      sessions_list.templ
      plugins_list.templ
      plugin_detail.templ
      import.templ
      audit.templ
      settings.templ
      login.templ

    partials/                    # templ partial components (HTMX swap targets)
      content_table_rows.templ
      content_tree_nodes.templ
      content_field_form.templ
      datatypes_table_rows.templ
      datatypes_field_list.templ
      fields_table_rows.templ
      media_grid_items.templ
      media_upload_result.templ
      users_table_rows.templ
      roles_table_rows.templ
      roles_permission_list.templ
      tokens_table_rows.templ
      sessions_table_rows.templ
      plugins_table_rows.templ
      audit_table_rows.templ
      toast.templ                # Toast notification (for OOB swaps)
      pagination.templ
      empty_state.templ
      delete_confirm.templ
      form_field.templ           # Reusable form field with error state

    components/                  # templ reusable UI components (shared across pages)
      sidebar.templ              # Navigation sidebar
      topbar.templ               # Top bar with user menu
      data_table.templ           # Table wrapper with header sorting
      status_badge.templ         # Status indicator
      icon.templ                 # SVG icon helper

    embed.go                     # //go:embed static/* directive
    static/                      # Go embed — static assets (CSS, JS only)
      css/
        admin.css                # All styles (design tokens, layout, components)
      js/
        htmx.min.js              # HTMX library (vendored)
        block-editor.js          # Block editor web component (copied from external repo)
        components/              # Web Components (all Light DOM)
          mcms-toast.js
          mcms-dialog.js
          mcms-data-table.js
          mcms-media-picker.js
          mcms-tree-nav.js
          mcms-field-renderer.js
          mcms-confirm.js
          mcms-search.js
        admin.js                 # Minimal glue: register components, HTMX config, CSRF header
      icons/                     # SVG icon sprites
```

## Why templ Instead of html/template

templ compiles `.templ` files into Go code via `templ generate`. This eliminates several problems the skeptic review identified with `html/template`:

| Problem with html/template | How templ solves it |
|---|---|
| Template parse errors at runtime | Compile-time errors — broken templates don't build |
| Template registry needed to compose layouts + partials | Function composition — `Admin(data) { ContentList(items) }` |
| FuncMap for helper functions | Call any Go function directly in templ expressions |
| `any` data type — no type safety | Typed function parameters — `ContentList(items []db.ContentData)` |
| Partial vs full page requires two code paths per template | Same component renders in both contexts — handler picks layout or not |
| Buffer-first rendering bolt-on | Built in — templ components write to `io.Writer` and return `error` |
| Template inheritance doesn't work in Go | Children pattern — `templ Admin() { {children...} }` |

### templ Component Model

```go
// layouts/base.templ
package layouts

templ Base(title string, csrfToken string) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
        <meta name="csrf-token" content={ csrfToken }/>
        <title>{ title } — ModulaCMS</title>
        <link rel="stylesheet" href="/admin/static/css/admin.css"/>
        <script src="/admin/static/js/htmx.min.js"></script>
        <script src="/admin/static/js/admin.js" defer></script>
    </head>
    <body>
        { children... }
    </body>
    </html>
}

// layouts/admin.templ
package layouts

import "github.com/hegner123/modulacms/internal/admin/components"

templ Admin(data AdminData) {
    @Base(data.Title, data.CSRFToken) {
        <div class="admin-layout">
            @components.Sidebar(data.CurrentPath, data.Permissions)
            <div class="admin-main">
                @components.Topbar(data.User)
                <main id="main-content" class="admin-content">
                    { children... }
                </main>
            </div>
        </div>
        <mcms-toast position="bottom-right" duration="5000"></mcms-toast>
    }
}

// layouts/auth.templ
package layouts

templ Auth(title string, csrfToken string) {
    @Base(title, csrfToken) {
        <div class="auth-layout">
            <div class="auth-card">
                { children... }
            </div>
        </div>
    }
}
```

### Page Components

```go
// pages/datatypes_list.templ
package pages

import (
    "github.com/hegner123/modulacms/internal/admin/layouts"
    "github.com/hegner123/modulacms/internal/admin/partials"
    "github.com/hegner123/modulacms/internal/db"
)

templ DatatypesList(layout layouts.AdminData, items []db.Datatypes, pagination PaginationData) {
    @layouts.Admin(layout) {
        <div class="page-header">
            <h1>Datatypes</h1>
            <button onclick="document.getElementById('create-dialog').setAttribute('open','')">
                Create Datatype
            </button>
        </div>
        <mcms-search
            hx-get="/admin/schema/datatypes"
            hx-target="#datatypes-table-body"
            hx-trigger="input changed delay:300ms"
            placeholder="Search datatypes...">
        </mcms-search>
        <mcms-data-table sortable>
            <table>
                <thead>
                    <tr>
                        <th data-sort="name">Name</th>
                        <th data-sort="label">Label</th>
                        <th data-sort="date_modified">Modified</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody id="datatypes-table-body">
                    @partials.DatatypesTableRows(items)
                </tbody>
            </table>
        </mcms-data-table>
        <div id="pagination">
            @partials.Pagination(pagination)
        </div>
    }
}
```

### Partial Components

```go
// partials/datatypes_table_rows.templ
package partials

import (
    "github.com/hegner123/modulacms/internal/db"
    "github.com/hegner123/modulacms/internal/admin/components"
)

templ DatatypesTableRows(items []db.Datatypes) {
    for _, dt := range items {
        <tr>
            <td>
                <a href={ templ.SafeURL("/admin/schema/datatypes/" + dt.DatatypeID.String()) }>
                    { dt.Name }
                </a>
            </td>
            <td>{ dt.Label }</td>
            <td>{ components.FormatDate(dt.DateModified) }</td>
            <td>
                <mcms-confirm
                    hx-delete={ "/admin/schema/datatypes/" + dt.DatatypeID.String() }
                    hx-target="closest tr"
                    hx-swap="outerHTML swap:500ms"
                    label="Delete"
                    confirm-label="Confirm">
                </mcms-confirm>
            </td>
        </tr>
    }
}

// partials/pagination.templ
package partials

templ Pagination(data PaginationData) {
    if data.TotalPages > 1 {
        <nav class="pagination">
            for _, page := range data.Pages() {
                if page == data.Current {
                    <span class="pagination-current">{ intToStr(page) }</span>
                } else {
                    <a hx-get={ data.URLForPage(page) }
                       hx-target={ data.Target }
                       hx-swap="innerHTML"
                       hx-push-url="true">
                        { intToStr(page) }
                    </a>
                }
            }
        </nav>
    }
}

// partials/form_field.templ
package partials

templ FormField(name string, label string, value string, fieldErr string) {
    <div class={ "form-field", templ.KV("has-error", fieldErr != "") }>
        <label for={ name }>{ label }</label>
        <input id={ name } name={ name } value={ value }/>
        if fieldErr != "" {
            <span class="field-error">{ fieldErr }</span>
        }
    </div>
}
```

### Shared Components

```go
// components/sidebar.templ
package components

import "github.com/hegner123/modulacms/internal/middleware"

type NavItem struct {
    Label      string
    Href       string
    Icon       string
    Permission string // empty = always visible
}

var navItems = []NavItem{
    {Label: "Dashboard", Href: "/admin/", Icon: "home"},
    {Label: "Content", Href: "/admin/content", Icon: "file-text", Permission: "content:read"},
    {Label: "Media", Href: "/admin/media", Icon: "image", Permission: "media:read"},
    {Label: "Datatypes", Href: "/admin/schema/datatypes", Icon: "blocks", Permission: "datatypes:read"},
    {Label: "Fields", Href: "/admin/schema/fields", Icon: "form-input", Permission: "fields:read"},
    {Label: "Routes", Href: "/admin/routes", Icon: "globe", Permission: "routes:read"},
    {Label: "Users", Href: "/admin/users", Icon: "users", Permission: "users:read"},
    {Label: "Roles", Href: "/admin/users/roles", Icon: "shield", Permission: "roles:read"},
    {Label: "Tokens", Href: "/admin/users/tokens", Icon: "key", Permission: "tokens:read"},
    {Label: "Plugins", Href: "/admin/plugins", Icon: "puzzle", Permission: "plugins:read"},
    {Label: "Import", Href: "/admin/import", Icon: "upload", Permission: "import:create"},
    {Label: "Audit", Href: "/admin/audit", Icon: "history"},
    {Label: "Settings", Href: "/admin/settings", Icon: "settings", Permission: "config:read"},
}

templ Sidebar(currentPath string, perms middleware.PermissionSet) {
    <nav class="sidebar">
        <div class="sidebar-brand">
            <a href="/admin/">ModulaCMS</a>
        </div>
        <ul class="sidebar-nav">
            for _, item := range navItems {
                if item.Permission == "" || perms.Has(item.Permission) {
                    <li>
                        <a href={ templ.SafeURL(item.Href) }
                           class={ templ.KV("active", isActive(currentPath, item.Href)) }>
                            @Icon(item.Icon)
                            { item.Label }
                        </a>
                    </li>
                }
            }
        </ul>
    </nav>
}

// isActive returns true if the current path matches the nav item's href.
// Uses path-boundary matching: /admin/users matches /admin/users and /admin/users/123
// but NOT /admin/users/roles (which is its own nav item).
// Special case: sub-resources like /admin/users/roles and /admin/users/tokens
// are separate nav items, so we check for exact prefix + next char is '/' or end-of-string,
// then exclude paths that match a more specific nav item.
func isActive(current, href string) bool {
    if href == "/admin/" {
        return current == "/admin/" || current == "/admin"
    }
    if !strings.HasPrefix(current, href) {
        return false
    }
    // Exact match
    if len(current) == len(href) {
        return true
    }
    // Prefix match only if next character is '/' (path boundary)
    return current[len(href)] == '/'
}
```

## Handler Pattern

Handlers are plain Go functions that call templ components. The handler decides whether to render a full page or a partial based on the `HX-Request` header.

```go
// handlers/datatypes.go
package handlers

func DatatypesListHandler(driver db.DbDriver) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        params := ParsePaginationParams(r)
        items, err := driver.ListDatatypesPaginated(params)
        if err != nil {
            http.Error(w, "failed to load datatypes", http.StatusInternalServerError)
            return
        }
        total, err := driver.CountDatatypes()
        if err != nil {
            http.Error(w, "failed to count datatypes", http.StatusInternalServerError)
            return
        }

        pagination := NewPaginationData(*total, params, "#datatypes-table-body", "/admin/schema/datatypes")

        // HTMX request: return partial only
        if r.Header.Get("HX-Request") != "" {
            Render(w, r, partials.DatatypesTableRows(*items))
            return
        }

        // Full page request: wrap in layout
        layout := NewAdminData(r, "Datatypes")
        Render(w, r, pages.DatatypesList(layout, *items, pagination))
    }
}
```

### Render Helpers

```go
// render.go
package admin

// Render writes a templ component to the response.
// templ components buffer internally before flushing to the writer, so a render
// error means no bytes have been sent yet and we can safely return a 500.
func Render(w http.ResponseWriter, r *http.Request, component templ.Component) {
    buf := new(bytes.Buffer)
    if err := component.Render(r.Context(), buf); err != nil {
        utility.DefaultLogger.Error("render failed", "error", err)
        if r.Header.Get("HX-Request") != "" {
            // HTMX request: return a toast-triggering error so the UI shows feedback
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

// RenderWithOOB renders a primary component plus out-of-band swap fragments in a single response.
// All components are buffered before writing to ensure atomic delivery — if any component
// fails to render, no partial HTML has been sent and we can return a clean error.
func RenderWithOOB(w http.ResponseWriter, r *http.Request, primary templ.Component, oob ...OOBSwap) {
    buf := new(bytes.Buffer)

    // Render primary into buffer
    if err := primary.Render(r.Context(), buf); err != nil {
        utility.DefaultLogger.Error("primary render failed", "error", err)
        if r.Header.Get("HX-Request") != "" {
            w.Header().Set("HX-Retarget", "#none")
            w.Header().Set("HX-Trigger", `{"showToast": {"message": "Render error", "type": "error"}}`)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Render OOB swaps into same buffer
    for _, swap := range oob {
        fmt.Fprintf(buf, `<div id="%s" hx-swap-oob="true">`, template.HTMLEscapeString(swap.TargetID))
        if err := swap.Component.Render(r.Context(), buf); err != nil {
            utility.DefaultLogger.Error("OOB render failed", "error", err, "target", swap.TargetID)
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

    // All renders succeeded — flush buffer to client
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    buf.WriteTo(w)
}

type OOBSwap struct {
    TargetID  string
    Component templ.Component
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
    }
}
```

### Layout Data Types

```go
// layouts/types.go
package layouts

import (
    "github.com/hegner123/modulacms/internal/db"
    "github.com/hegner123/modulacms/internal/middleware"
)

type AdminData struct {
    Title       string
    CurrentPath string
    User        *db.Users
    Permissions middleware.PermissionSet
    IsAdmin     bool
    CSRFToken   string
}
```

## CSRF Protection

Same double-submit cookie pattern as before. CSRF token is passed to layouts as a typed string parameter instead of via FuncMap.

```go
// csrf.go — unchanged from previous plan revision
```

**templ integration:**

```go
// In base.templ:
<meta name="csrf-token" content={ csrfToken }/>

// In form partials:
templ CSRFField(token string) {
    <input type="hidden" name="_csrf" value={ token }/>
}

// In admin.js — inject CSRF header into all HTMX requests:
document.body.addEventListener('htmx:configRequest', (e) => {
    const meta = document.querySelector('meta[name="csrf-token"]');
    if (meta) e.detail.headers['X-CSRF-Token'] = meta.content;
});
```

## Routing Design

### URL Structure

All admin HTML routes live under `/admin/`. Mutations use `POST` for creates and updates (HTML form compatible). `DELETE` uses HTMX only.

| URL Pattern | Handler | Page |
|---|---|---|
| `GET /admin/` | `DashboardHandler` | Dashboard |
| `GET /admin/login` | `LoginPageHandler` | Login form |
| `POST /admin/login` | `LoginSubmitHandler` | Process login, redirect |
| `POST /admin/logout` | `LogoutHandler` | Clear session, redirect to login |
| `GET /admin/content` | `ContentListHandler` | Content tree + list |
| `GET /admin/content/{id}` | `ContentEditHandler` | Content editor with block editor |
| `POST /admin/content` | `ContentCreateHandler` | Create content |
| `POST /admin/content/{id}` | `ContentUpdateHandler` | Update content |
| `DELETE /admin/content/{id}` | `ContentDeleteHandler` | Delete content (HTMX only) |
| `POST /admin/content/reorder` | `ContentReorderHandler` | Reorder children |
| `POST /admin/content/move` | `ContentMoveHandler` | Move across parents |
| `GET /admin/schema/datatypes` | `DatatypesListHandler` | Datatypes list |
| `GET /admin/schema/datatypes/{id}` | `DatatypeDetailHandler` | Datatype detail + linked fields |
| `POST /admin/schema/datatypes` | `DatatypeCreateHandler` | Create datatype |
| `POST /admin/schema/datatypes/{id}` | `DatatypeUpdateHandler` | Update datatype |
| `DELETE /admin/schema/datatypes/{id}` | `DatatypeDeleteHandler` | Delete datatype (HTMX only) |
| `GET /admin/schema/fields` | `FieldsListHandler` | Fields list |
| `GET /admin/schema/fields/{id}` | `FieldDetailHandler` | Field detail + config |
| `POST /admin/schema/fields` | `FieldCreateHandler` | Create field |
| `POST /admin/schema/fields/{id}` | `FieldUpdateHandler` | Update field |
| `DELETE /admin/schema/fields/{id}` | `FieldDeleteHandler` | Delete field (HTMX only) |
| `GET /admin/schema/field-types` | `FieldTypesListHandler` | Field types list |
| `GET /admin/media` | `MediaListHandler` | Media grid |
| `GET /admin/media/{id}` | `MediaDetailHandler` | Media detail + edit |
| `POST /admin/media` | `MediaUploadHandler` | Upload (multipart) |
| `POST /admin/media/{id}` | `MediaUpdateHandler` | Update media metadata |
| `DELETE /admin/media/{id}` | `MediaDeleteHandler` | Delete media (HTMX only) |
| `GET /admin/routes` | `RoutesListHandler` | Routes list |
| `GET /admin/routes/admin` | `AdminRoutesListHandler` | Admin routes list |
| `POST /admin/routes` | `RouteCreateHandler` | Create route |
| `POST /admin/routes/{id}` | `RouteUpdateHandler` | Update route |
| `DELETE /admin/routes/{id}` | `RouteDeleteHandler` | Delete route (HTMX only) |
| `GET /admin/users` | `UsersListHandler` | Users list |
| `GET /admin/users/{id}` | `UserDetailHandler` | User detail + edit |
| `POST /admin/users` | `UserCreateHandler` | Create user |
| `POST /admin/users/{id}` | `UserUpdateHandler` | Update user |
| `DELETE /admin/users/{id}` | `UserDeleteHandler` | Delete user (HTMX only) |
| `GET /admin/users/roles` | `RolesListHandler` | Roles + permissions |
| `POST /admin/users/roles` | `RoleCreateHandler` | Create role |
| `POST /admin/users/roles/{id}` | `RoleUpdateHandler` | Update role |
| `DELETE /admin/users/roles/{id}` | `RoleDeleteHandler` | Delete role (HTMX only) |
| `GET /admin/users/tokens` | `TokensListHandler` | API tokens list |
| `POST /admin/users/tokens` | `TokenCreateHandler` | Create token |
| `DELETE /admin/users/tokens/{id}` | `TokenDeleteHandler` | Delete token (HTMX only) |
| `GET /admin/sessions` | `SessionsListHandler` | Active sessions |
| `DELETE /admin/sessions/{id}` | `SessionDeleteHandler` | Revoke session (HTMX only) |
| `GET /admin/plugins` | `PluginsListHandler` | Plugin list |
| `GET /admin/plugins/{name}` | `PluginDetailHandler` | Plugin detail |
| `GET /admin/import` | `ImportPageHandler` | Import wizard |
| `POST /admin/import` | `ImportSubmitHandler` | Process import |
| `GET /admin/audit` | `AuditLogHandler` | Audit log |
| `GET /admin/settings` | `SettingsHandler` | Settings page |
| `POST /admin/settings` | `SettingsUpdateHandler` | Update settings |

### Registration with Permission Middleware

Admin HTML routes use the same `RequireResourcePermission` and `RequirePermission` wrappers as the JSON API. The admin panel lives inside the `DefaultMiddlewareChain` (which runs `HTTPAuthenticationMiddleware` and `PermissionInjector`), so permission context is already available.

```go
// admin.go

func RegisterAdminRoutes(mux *http.ServeMux, mgr *config.Manager, driver db.DbDriver, pc *middleware.PermissionCache) {
    // Static assets (no auth required, no CSRF)
    staticFS, _ := fs.Sub(staticFiles, "static")
    mux.Handle("GET /admin/static/", http.StripPrefix("/admin/static/", http.FileServer(http.FS(staticFS))))

    // Auth pages (no auth required, no permission check)
    // Login POST gets CSRF + rate limiting but no session auth (user is not yet authenticated)
    loginLimiter := middleware.NewRateLimiter(rate.Limit(10.0/60.0), 10) // 10 attempts/min per IP
    mux.HandleFunc("GET /admin/login", handlers.LoginPageHandler())
    mux.Handle("POST /admin/login", loginLimiter(CSRFMiddleware()(http.HandlerFunc(handlers.LoginSubmitHandler(mgr)))))
    mux.HandleFunc("POST /admin/logout", handlers.LogoutHandler(mgr))

    adminAuth := AdminAuthMiddleware(mgr)
    csrf := CSRFMiddleware()

    // mutating wraps auth + CSRF + permission — use for all POST/DELETE handlers
    mutating := func(permission string, h http.Handler) http.Handler {
        return adminAuth(csrf(middleware.RequirePermission(permission)(h)))
    }
    // viewing wraps auth + permission only (no CSRF needed for GET requests)
    viewing := func(resource string, h http.Handler) http.Handler {
        return adminAuth(middleware.RequirePermission(resource+":read")(h))
    }

    // Dashboard
    mux.Handle("GET /admin/", adminAuth(http.HandlerFunc(handlers.DashboardHandler(driver))))

    // Content
    mux.Handle("GET /admin/content", viewing("content", http.HandlerFunc(handlers.ContentListHandler(driver))))
    mux.Handle("GET /admin/content/{id}", viewing("content", http.HandlerFunc(handlers.ContentEditHandler(driver))))
    mux.Handle("POST /admin/content", mutating("content:create", http.HandlerFunc(handlers.ContentCreateHandler(driver))))
    mux.Handle("POST /admin/content/{id}", mutating("content:update", http.HandlerFunc(handlers.ContentUpdateHandler(driver))))
    mux.Handle("DELETE /admin/content/{id}", mutating("content:delete", http.HandlerFunc(handlers.ContentDeleteHandler(driver))))
    mux.Handle("POST /admin/content/reorder", mutating("content:update", http.HandlerFunc(handlers.ContentReorderHandler(driver))))
    mux.Handle("POST /admin/content/move", mutating("content:update", http.HandlerFunc(handlers.ContentMoveHandler(driver))))

    // Schema — datatypes
    mux.Handle("GET /admin/schema/datatypes", viewing("datatypes", http.HandlerFunc(handlers.DatatypesListHandler(driver))))
    mux.Handle("GET /admin/schema/datatypes/{id}", viewing("datatypes", http.HandlerFunc(handlers.DatatypeDetailHandler(driver))))
    mux.Handle("POST /admin/schema/datatypes", mutating("datatypes:create", http.HandlerFunc(handlers.DatatypeCreateHandler(driver))))
    mux.Handle("POST /admin/schema/datatypes/{id}", mutating("datatypes:update", http.HandlerFunc(handlers.DatatypeUpdateHandler(driver))))
    mux.Handle("DELETE /admin/schema/datatypes/{id}", mutating("datatypes:delete", http.HandlerFunc(handlers.DatatypeDeleteHandler(driver))))

    // Schema — fields
    mux.Handle("GET /admin/schema/fields", viewing("fields", http.HandlerFunc(handlers.FieldsListHandler(driver))))
    mux.Handle("GET /admin/schema/fields/{id}", viewing("fields", http.HandlerFunc(handlers.FieldDetailHandler(driver))))
    mux.Handle("POST /admin/schema/fields", mutating("fields:create", http.HandlerFunc(handlers.FieldCreateHandler(driver))))
    mux.Handle("POST /admin/schema/fields/{id}", mutating("fields:update", http.HandlerFunc(handlers.FieldUpdateHandler(driver))))
    mux.Handle("DELETE /admin/schema/fields/{id}", mutating("fields:delete", http.HandlerFunc(handlers.FieldDeleteHandler(driver))))

    // Schema — field types
    mux.Handle("GET /admin/schema/field-types", viewing("field_types", http.HandlerFunc(handlers.FieldTypesListHandler(driver))))

    // Media
    mux.Handle("GET /admin/media", viewing("media", http.HandlerFunc(handlers.MediaListHandler(driver))))
    mux.Handle("GET /admin/media/{id}", viewing("media", http.HandlerFunc(handlers.MediaDetailHandler(driver))))
    mux.Handle("POST /admin/media", mutating("media:create", http.HandlerFunc(handlers.MediaUploadHandler(driver))))
    mux.Handle("POST /admin/media/{id}", mutating("media:update", http.HandlerFunc(handlers.MediaUpdateHandler(driver))))
    mux.Handle("DELETE /admin/media/{id}", mutating("media:delete", http.HandlerFunc(handlers.MediaDeleteHandler(driver))))

    // Routes
    mux.Handle("GET /admin/routes", viewing("routes", http.HandlerFunc(handlers.RoutesListHandler(driver))))
    mux.Handle("GET /admin/routes/admin", viewing("routes", http.HandlerFunc(handlers.AdminRoutesListHandler(driver))))
    mux.Handle("POST /admin/routes", mutating("routes:create", http.HandlerFunc(handlers.RouteCreateHandler(driver))))
    mux.Handle("POST /admin/routes/{id}", mutating("routes:update", http.HandlerFunc(handlers.RouteUpdateHandler(driver))))
    mux.Handle("DELETE /admin/routes/{id}", mutating("routes:delete", http.HandlerFunc(handlers.RouteDeleteHandler(driver))))

    // Users
    mux.Handle("GET /admin/users", viewing("users", http.HandlerFunc(handlers.UsersListHandler(driver))))
    mux.Handle("GET /admin/users/{id}", viewing("users", http.HandlerFunc(handlers.UserDetailHandler(driver))))
    mux.Handle("POST /admin/users", mutating("users:create", http.HandlerFunc(handlers.UserCreateHandler(driver))))
    mux.Handle("POST /admin/users/{id}", mutating("users:update", http.HandlerFunc(handlers.UserUpdateHandler(driver))))
    mux.Handle("DELETE /admin/users/{id}", mutating("users:delete", http.HandlerFunc(handlers.UserDeleteHandler(driver))))

    // Roles
    mux.Handle("GET /admin/users/roles", viewing("roles", http.HandlerFunc(handlers.RolesListHandler(driver, pc))))
    mux.Handle("POST /admin/users/roles", mutating("roles:create", http.HandlerFunc(handlers.RoleCreateHandler(driver, pc))))
    mux.Handle("POST /admin/users/roles/{id}", mutating("roles:update", http.HandlerFunc(handlers.RoleUpdateHandler(driver, pc))))
    mux.Handle("DELETE /admin/users/roles/{id}", mutating("roles:delete", http.HandlerFunc(handlers.RoleDeleteHandler(driver, pc))))

    // Tokens
    mux.Handle("GET /admin/users/tokens", viewing("tokens", http.HandlerFunc(handlers.TokensListHandler(driver))))
    mux.Handle("POST /admin/users/tokens", mutating("tokens:create", http.HandlerFunc(handlers.TokenCreateHandler(driver))))
    mux.Handle("DELETE /admin/users/tokens/{id}", mutating("tokens:delete", http.HandlerFunc(handlers.TokenDeleteHandler(driver))))

    // Sessions
    mux.Handle("GET /admin/sessions", viewing("sessions", http.HandlerFunc(handlers.SessionsListHandler(driver))))
    mux.Handle("DELETE /admin/sessions/{id}", mutating("sessions:delete", http.HandlerFunc(handlers.SessionDeleteHandler(driver))))

    // Plugins
    mux.Handle("GET /admin/plugins", viewing("plugins", http.HandlerFunc(handlers.PluginsListHandler(driver))))
    mux.Handle("GET /admin/plugins/{name}", viewing("plugins", http.HandlerFunc(handlers.PluginDetailHandler(driver))))

    // Import
    mux.Handle("GET /admin/import", viewing("import", http.HandlerFunc(handlers.ImportPageHandler())))
    mux.Handle("POST /admin/import", mutating("import:create", http.HandlerFunc(handlers.ImportSubmitHandler(driver))))

    // Audit
    mux.Handle("GET /admin/audit", adminAuth(http.HandlerFunc(handlers.AuditLogHandler(driver))))

    // Settings
    mux.Handle("GET /admin/settings", viewing("config", http.HandlerFunc(handlers.SettingsHandler(mgr))))
    mux.Handle("POST /admin/settings", mutating("config:update", http.HandlerFunc(handlers.SettingsUpdateHandler(mgr))))

    // Root redirect
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/admin/", http.StatusFound)
    })
}
```

## Web Components Design

### Light DOM by Default

All `mcms-*` components use **Light DOM** (no Shadow DOM). Required because HTMX attribute inheritance (`hx-target`, `hx-swap`, `hx-headers`) does not cross Shadow DOM boundaries.

The block editor is the exception — it uses Shadow DOM because it is self-contained and communicates via the JSON API, not HTMX.

### Component Catalog

#### `<mcms-toast>` — Global toast notifications
Listens for HTMX `showToast` events triggered by response headers.

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `position` | `"top-right"` \| `"top-left"` \| `"bottom-right"` \| `"bottom-left"` | `"bottom-right"` | Screen position for toast stack |
| `duration` | number (ms) | `5000` | Auto-dismiss duration. `0` = persistent until dismissed. |

| Method | Signature | Description |
|--------|-----------|-------------|
| `show` | `show(message: string, type: "success" \| "error" \| "info")` | Programmatically show a toast |

Listens for HTMX event `showToast` with detail `{message: string, type: string}`. Example server trigger: `HX-Trigger: {"showToast": {"message": "Saved", "type": "success"}}`.

#### `<mcms-dialog>` — Modal dialog

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `open` | boolean (presence) | absent | Controls visibility. Set to open, remove to close. |
| `title` | string | `""` | Dialog title text |
| `confirm-label` | string | `"Confirm"` | Confirm button text |
| `cancel-label` | string | `"Cancel"` | Cancel button text |
| `destructive` | boolean (presence) | absent | Styles confirm button as danger |

| Event | Detail | Description |
|-------|--------|-------------|
| `mcms-dialog:confirm` | `{}` | Fired when confirm button clicked |
| `mcms-dialog:cancel` | `{}` | Fired when cancel button clicked or Escape pressed |

Dialog content is slotted as children (Light DOM — not `<slot>`, just child elements).

#### `<mcms-data-table>` — Table wrapper
Sorting, column visibility toggles, bulk selection. Does NOT fetch data.

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `sortable` | boolean (presence) | absent | Enables column sort headers |
| `selectable` | boolean (presence) | absent | Adds row checkboxes for bulk selection |

| Event | Detail | Description |
|-------|--------|-------------|
| `mcms-table:sort` | `{column: string, direction: "asc" \| "desc"}` | Fired when a sortable column header is clicked |
| `mcms-table:select` | `{selected: string[]}` | Fired when row selection changes. Array of row `data-id` values. |

Table reads sort column from `<th data-sort="column_name">`. The component does NOT sort data client-side — it dispatches the event and the consumer uses HTMX to re-fetch sorted data.

#### `<mcms-media-picker>` — Media browser/selector
Opens dialog to browse/search/select media.

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `open` | boolean (presence) | absent | Controls visibility |
| `accept` | string | `"*"` | MIME type filter (e.g., `"image/*"`) |
| `multiple` | boolean (presence) | absent | Allow selecting multiple items |

| Event | Detail | Description |
|-------|--------|-------------|
| `media-selected` | `{id: string, url: string, alt: string}` or `{items: Array<{id, url, alt}>}` if `multiple` | Fired when user confirms selection |
| `media-picker:cancel` | `{}` | Fired when picker is closed without selection |

Internally uses `hx-get="/admin/media?picker=true"` to load the media grid inside the picker dialog.

#### `<mcms-tree-nav>` — Content tree navigation
Collapsible tree with expand/collapse, selection, drag-and-drop reorder, and move-across-parents.

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `data-active` | string | `""` | Content ID of the currently-edited node (highlighted) |

| Event | Detail | Description |
|-------|--------|-------------|
| `tree-nav:reorder` | `{parent_id: string, ordered_ids: string[]}` | Fired after drag-reorder within same parent |
| `tree-nav:move` | `{content_id: string, new_parent_id: string, position: number}` | Fired after drag to different parent |

**Lazy loading:** Tree renders the root level on initial page load. Expanding a node triggers `hx-get="/admin/content/tree/{id}/children"` which returns `<li>` children as a partial. Previously-expanded nodes are cached in the DOM (no re-fetch on collapse/expand toggle).

**Selection:** Clicking a node navigates to `/admin/content/{id}` via standard `<a href>`. The currently-edited content ID is highlighted via a `data-active` attribute set by the page template.

**Reorder (drag-and-drop within same parent):**
1. User drags a node to a new position among its siblings
2. Component collects the new sibling order as an array of content IDs
3. Dispatches `POST /admin/content/reorder` via HTMX with `{parent_id, ordered_ids: [...]}`
4. Server updates `next_sibling_id`/`prev_sibling_id` pointers atomically
5. Response: re-rendered `<ul>` of the affected parent's children (partial swap of the parent `<ul>`)

**Move (drag to different parent):**
1. User drags a node onto a different parent node (visual drop-zone highlight)
2. Dispatches `POST /admin/content/move` via HTMX with `{content_id, new_parent_id, position}`
3. Server detaches from old parent, attaches to new parent, updates all sibling pointers
4. Response: OOB swap of both the old parent's `<ul>` and the new parent's `<ul>` so both branches update

**Error handling:** Failed reorder/move returns `HX-Trigger: showToast` with error message. The tree does NOT optimistically update — it waits for the server response before reflecting changes.

#### `<mcms-field-renderer>` — Field type widgets
Renders the correct input control based on the CMS field type. The `type` attribute selects the rendering mode:

| Field Type | Rendered Control | Notes |
|---|---|---|
| `text` | `<input type="text">` | Standard single-line text input |
| `textarea` | `<textarea>` | Multi-line plain text, auto-resizes |
| `richtext` | `<textarea>` with markdown preview toggle | Preview rendered client-side via a lightweight markdown-to-HTML function (no third-party editor library). Stored as markdown. |
| `number` | `<input type="number">` | Respects `min`/`max`/`step` from field config |
| `boolean` | `<input type="checkbox">` | Toggle switch styled via CSS |
| `date` | `<input type="datetime-local">` | Native browser date picker |
| `select` | `<select>` | Options populated from field config `choices` array |
| `media` | Read-only thumbnail + "Choose" button | "Choose" opens `<mcms-media-picker>`, listens for `media-selected` event, updates a hidden `<input>` with the selected media ID and shows the thumbnail |
| `reference` | Read-only label + "Choose" button | Opens a content search dialog, sets hidden `<input>` with selected content ID |

All field renderers emit a `field-change` custom event with `{name, value}` on input, enabling the parent form to track dirty state.

#### `<mcms-confirm>` — Inline confirmation button
First click shows "Are you sure?", second click triggers action.

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `label` | string | `"Delete"` | Initial button text |
| `confirm-label` | string | `"Confirm"` | Text shown after first click |
| `timeout` | number (ms) | `3000` | Time before reverting to initial state if not confirmed |

HTMX attributes (`hx-delete`, `hx-target`, `hx-swap`) are placed directly on the `<mcms-confirm>` element. On confirm (second click), the component triggers the HTMX request. The component itself is a `<button>` in Light DOM.

#### `<mcms-search>` — Debounced search input
Triggers HTMX requests with configurable delay.

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `placeholder` | string | `"Search..."` | Input placeholder text |
| `name` | string | `"q"` | Form field name for the search parameter |
| `value` | string | `""` | Initial search value (for preserving state on page load) |

HTMX attributes (`hx-get`, `hx-target`, `hx-trigger`) are placed directly on the `<mcms-search>` element and inherited by the inner `<input>`. The component renders a text input with a clear button that appears when the input has a value.

## HTMX Patterns

### Page Navigation
Standard `<a href>` links for top-level navigation. Browser history, bookmarks, and back button work naturally.

### In-Page Updates
```html
<input type="search" name="q"
       hx-get="/admin/users" hx-target="#users-table-body"
       hx-trigger="input changed delay:300ms" hx-push-url="true"
       placeholder="Search users...">
```

### Multi-Target Responses (OOB Swaps)
```go
// In handler after delete:
RenderWithOOB(w, r,
    partials.DatatypesTableRows(updatedItems),
    OOBSwap{TargetID: "pagination", Component: partials.Pagination(newPagination)},
)
w.Header().Set("HX-Trigger", `{"showToast": {"message": "Deleted", "type": "success"}}`)
```

### Toast Notifications
```go
w.Header().Set("HX-Trigger", `{"showToast": {"message": "Created", "type": "success"}}`)
```

### Form Validation Pattern

All mutation handlers (create/update) follow the same validation error flow. Validation errors re-render the form with per-field error messages via HTMX partial swap — never a full page reload.

**Standard pattern:**

```go
// handlers/datatypes.go

func DatatypeCreateHandler(driver db.DbDriver) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
        name := strings.TrimSpace(r.FormValue("name"))
        label := strings.TrimSpace(r.FormValue("label"))

        // Validate — collect all errors, don't fail on first
        errs := make(map[string]string)
        if name == "" {
            errs["name"] = "Name is required"
        }
        if label == "" {
            errs["label"] = "Label is required"
        }

        // Check uniqueness
        if name != "" {
            existing, _ := driver.GetDatatypeByName(name)
            if existing != nil {
                errs["name"] = "A datatype with this name already exists"
            }
        }

        // Validation failed: re-render the form partial with error state
        if len(errs) > 0 {
            w.WriteHeader(http.StatusUnprocessableEntity)
            Render(w, r, partials.DatatypeForm(name, label, errs, CSRFTokenFromContext(r.Context())))
            return
        }

        // Success: create, then swap in updated table rows + toast
        // ...
    }
}
```

**Form partial with error state:**

```go
// partials/datatype_form.templ

templ DatatypeForm(name string, label string, errs map[string]string, csrfToken string) {
    <form id="datatype-form"
          hx-post="/admin/schema/datatypes"
          hx-target="#datatype-form"
          hx-swap="outerHTML">
        @FormField("name", "Name", name, errs["name"])
        @FormField("label", "Label", label, errs["label"])
        @CSRFField(csrfToken)
        <button type="submit">Create</button>
    </form>
}
```

**Key rules:**
- The form's `hx-target` points to itself (`#datatype-form`) so validation errors swap the form in-place with error annotations
- Success responses use `HX-Redirect` or OOB swaps to update the list + show a toast — never re-render the form on success
- HTTP 422 for validation errors (HTMX processes the response body for 422 by default)
- All errors are collected before responding — never show one error at a time

### Media Upload Pattern

Media upload uses HTMX multipart form submission. The form includes `hx-encoding="multipart/form-data"` to enable file upload.

```html
<form hx-post="/admin/media"
      hx-target="#media-grid"
      hx-swap="afterbegin"
      hx-encoding="multipart/form-data">
    <input type="file" name="file" accept="image/*,video/*,audio/*,.pdf,.zip" required/>
    <input type="text" name="alt_text" placeholder="Alt text"/>
    @CSRFField(csrfToken)
    <button type="submit">Upload</button>
</form>
```

The `MediaUploadHandler` reuses the same multipart parsing logic from the existing JSON API handler (`internal/router/media.go`), calling `r.ParseMultipartForm(32 << 20)` and passing the file to the media processing pipeline. On success, it returns the new media item as a `partials.MediaGridItem` partial prepended to the grid, plus a toast via OOB swap.

## CSS Strategy

Single CSS file with design tokens. No Tailwind, no preprocessor.

```css
:root {
  --color-bg: #0a0a0a;
  --color-surface: #141414;
  --color-border: #262626;
  --color-text: #fafafa;
  --color-text-muted: #a1a1aa;
  --color-primary: #3b82f6;
  --color-danger: #ef4444;
  --color-success: #22c55e;
  --radius: 0.5rem;
  --sidebar-width: 16rem;
}
.htmx-swapping { opacity: 0; transition: opacity 300ms ease-out; }
.htmx-settling { opacity: 1; transition: opacity 300ms ease-in; }
```

Dark mode by default. Light mode via `<html class="light">` overriding tokens.

### CSS File Organization

The single `admin.css` file is organized into labeled sections to prevent merge conflicts during parallel Phase 3 work. Each section is delimited by a comment banner:

```css
/* ========================================
   SECTION: Design Tokens
   ======================================== */

/* ========================================
   SECTION: Layout (sidebar, topbar, admin-layout, admin-content)
   ======================================== */

/* ========================================
   SECTION: Components (shared: tables, forms, buttons, badges, pagination)
   Phase 1 owns this section
   ======================================== */

/* ========================================
   SECTION: Content Pages (Batch A)
   ======================================== */

/* ========================================
   SECTION: Schema Pages (Batch B)
   ======================================== */

/* ========================================
   SECTION: Users & Routes Pages (Batch C)
   ======================================== */

/* ========================================
   SECTION: Operations Pages (Batch D)
   ======================================== */
```

Phase 1 creates the file with Design Tokens, Layout, and Components sections populated. Each Phase 3 batch appends styles only to its designated section. This avoids merge conflicts since each batch writes to a different region of the file.

## Static Asset Caching

Embedded static files are served with cache-busting via a build version query parameter. The `admin.go` file reads the version from the binary's build info (set via ldflags) and exposes it to templates:

```go
// In base.templ — version query param for cache busting:
<link rel="stylesheet" href={ "/admin/static/css/admin.css?v=" + version }/>
<script src={ "/admin/static/js/htmx.min.js?v=" + version }></script>
<script src={ "/admin/static/js/admin.js?v=" + version } defer></script>
```

The static file handler sets `Cache-Control: public, max-age=31536000, immutable` — browsers cache aggressively, and the version query parameter forces re-fetch on deployment. During development (`just admin-watch`), the version is set to a timestamp so every page load fetches fresh assets.

## Authentication Flow

Same cookie-based session system. Admin routes run inside `DefaultMiddlewareChain`, which runs `HTTPAuthenticationMiddleware` (populates user in context if session is valid) and `PermissionInjector` (resolves role to permission set). Neither of these reject unauthenticated requests — they only populate context. `AdminAuthMiddleware` is the enforcement layer that redirects unauthenticated users.

1. **Unauthenticated `/admin/*`** -> redirect to `/admin/login?next={current_path}` (preserves intended destination)
2. **Login page** renders `login.templ` form, includes `next` as a hidden field
3. **POST `/admin/login`** -> rate-limited (10/min per IP) + CSRF-protected. `auth.CheckPassword`, create session, set cookie, redirect to `next` param (default `/admin/`). The `next` value is validated to start with `/admin/` to prevent open redirect.
4. **Logout** -> clear cookie, redirect to `/admin/login`
5. **Session expiry (regular request)** -> 302 redirect to `/admin/login?next={current_path}`
6. **Session expiry (HTMX request)** -> `HX-Redirect: /admin/login?next={current_path}` (HTMX follows the redirect, preserving the return URL)

### AdminAuthMiddleware Implementation

```go
// middleware.go

// AdminAuthMiddleware checks that the request has an authenticated user in context
// (placed there by HTTPAuthenticationMiddleware in DefaultMiddlewareChain).
// If no user is present, it redirects to the login page with a ?next= parameter
// so the user returns to their intended page after login.
func AdminAuthMiddleware(mgr *config.Manager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := middleware.AuthenticatedUser(r.Context())
            if user == nil {
                // Validate next parameter: must start with /admin/ to prevent open redirect
                nextPath := r.URL.Path
                if !strings.HasPrefix(nextPath, "/admin/") && nextPath != "/admin" {
                    nextPath = "/admin/"
                }
                loginURL := "/admin/login?next=" + url.QueryEscape(nextPath)

                if r.Header.Get("HX-Request") != "" {
                    // HTMX: use HX-Redirect so HTMX performs a full page redirect
                    w.Header().Set("HX-Redirect", loginURL)
                    w.WriteHeader(http.StatusUnauthorized)
                    return
                }
                http.Redirect(w, r, loginURL, http.StatusFound)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Block Editor Integration

The block editor is an existing `<block-editor>` web component. It uses Shadow DOM, custom drag-and-drop on pointer events, and the CMS sibling-pointer tree model. Events: `block-editor:change`, `block-editor:save`, `block-editor:error`.

Copied into `internal/admin/static/js/block-editor.js`. Loaded on the content edit page:

```go
// pages/content_edit.templ

templ ContentEdit(layout layouts.AdminData, content db.ContentData, fields []db.Fields, blockStateJSON string) {
    @layouts.Admin(layout) {
        <div class="content-editor-layout">
            <div class="page-fields">
                for _, f := range fields {
                    <mcms-field-renderer
                        type={ f.FieldType }
                        name={ f.Name }
                        value={ f.Value }
                        label={ f.Label }>
                    </mcms-field-renderer>
                }
            </div>
            <block-editor id="block-editor" data-state={ blockStateJSON }></block-editor>
            <div class="document-settings">
                @partials.DocumentSettings(content)
            </div>
        </div>
        <script type="module" src="/admin/static/js/block-editor.js"></script>
        <script type="module">
            const editor = document.getElementById('block-editor');
            const csrfMeta = document.querySelector('meta[name="csrf-token"]');
            editor.addEventListener('block-editor:save', (e) => {
                const headers = {'Content-Type': 'application/json'};
                if (csrfMeta) headers['X-CSRF-Token'] = csrfMeta.content;
                fetch('/api/v1/content/batch', {
                    method: 'POST',
                    credentials: 'same-origin',
                    headers,
                    body: JSON.stringify(e.detail),
                }).then(res => {
                    const toast = document.querySelector('mcms-toast');
                    if (res.ok) {
                        toast?.show('Saved', 'success');
                    } else {
                        editor.dispatchEvent(new CustomEvent('block-editor:error', {detail: {status: res.status}}));
                        toast?.show('Save failed (status ' + res.status + ')', 'error');
                    }
                }).catch(() => {
                    document.querySelector('mcms-toast')?.show('Save failed (network error)', 'error');
                });
            });
        </script>
    }
}
```

## Build Integration

templ requires a code generation step: `templ generate` compiles `.templ` files into `*_templ.go` files. These generated files are committed to the repo (same pattern as sqlc-generated code).

```bash
# justfile additions
admin-generate:
    templ generate -f internal/admin/

admin-watch:
    templ generate -f internal/admin/ --watch

dev: admin-generate
    go build -o modulacms-x86 ...
```

The generated `*_templ.go` files are committed so that `go build` works without the `templ` CLI installed. Only developers modifying `.templ` files need `templ` installed.

Install: `go install github.com/a-h/templ/cmd/templ@latest`

**Merge conflict mitigation:** Unlike sqlc-generated files (which change only on schema changes), templ-generated files churn frequently during active UI development. To minimize merge conflicts during parallel Phase 3 work:
- Each agent works on a separate batch (A/B/C/D) in separate directories (different handler files, different page/partial templ files)
- Generated `*_templ.go` files are 1:1 with their source `.templ` file, so different batches produce different generated files — no shared generated file conflicts
- CI verifies generated files are current: `templ generate && git diff --exit-code internal/admin/`
- If conflicts do occur in generated files, resolution is: delete the conflicting `*_templ.go`, run `templ generate`, commit the result

## Testing Strategy

### Test File Organization

| Test File | What It Covers | Estimated Tests |
|---|---|---|
| `csrf_test.go` | CSRF token generation, validation, cookie handling | ~6 |
| `middleware_test.go` | AdminAuthMiddleware redirects, rate limiting, session expiry | ~5 |
| `render_test.go` | Render error handling, RenderWithOOB, buffer-then-write | ~4 |
| `handlers/auth_test.go` | Login (valid, invalid, rate limited, CSRF), logout, redirect-back | ~8 |
| `handlers/content_test.go` | Content CRUD, tree reorder/move, block editor integration | ~10 |
| `handlers/datatypes_test.go` | Datatypes CRUD, validation errors, partial-vs-full | ~8 |
| `handlers/fields_test.go` | Fields CRUD, field config validation | ~6 |
| `handlers/media_test.go` | Media list, upload (multipart), delete | ~6 |
| `handlers/users_test.go` | Users CRUD, roles, tokens, sessions | ~8 |
| `handlers/routes_test.go` | Routes CRUD, admin routes | ~4 |
| `handlers/settings_test.go` | Settings read/update, config validation | ~4 |
| `partials_test.go` | Component render tests for key partials | ~6 |
| **Total** | | **~75** |

### Compile-Time Safety
templ components are Go functions. If a component references a nonexistent field or passes the wrong type, it fails at `go build` — no runtime template errors possible.

### Required Test Categories Per Handler

Every handler must have tests covering:

1. **Unauthenticated access** — verifies 302 redirect to `/admin/login`
2. **Insufficient permissions** — verifies 403 when user lacks the required permission
3. **Full page render** — verifies full HTML page (contains `<!DOCTYPE`) for non-HTMX requests
4. **HTMX partial render** — verifies partial (no `<!DOCTYPE`) when `HX-Request: true` header is set
5. **Mutation + CSRF** — (POST/DELETE only) verifies 403 without valid CSRF token
6. **Validation errors** — (create/update only) verifies 422 response with error messages in re-rendered form

### Handler Test Examples
```go
func TestDatatypesListHandler_Unauthenticated(t *testing.T) {
    req := httptest.NewRequest("GET", "/admin/schema/datatypes", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    if rec.Code != http.StatusFound {
        t.Errorf("expected 302, got %d", rec.Code)
    }
}

func TestDatatypesListHandler_HTMXPartial(t *testing.T) {
    req := httptest.NewRequest("GET", "/admin/schema/datatypes", nil)
    req.Header.Set("HX-Request", "true")
    // ... set auth + permission context ...
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    body := rec.Body.String()
    // Partial: should contain <tr> elements but NOT <!DOCTYPE
    if strings.Contains(body, "<!DOCTYPE") {
        t.Error("HTMX request should return partial, not full page")
    }
}

func TestDatatypeCreateHandler_ValidationError(t *testing.T) {
    form := url.Values{"name": {""}, "label": {""}}
    req := httptest.NewRequest("POST", "/admin/schema/datatypes", strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    // ... set auth + permission + CSRF context ...
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    if rec.Code != http.StatusUnprocessableEntity {
        t.Errorf("expected 422, got %d", rec.Code)
    }
    body := rec.Body.String()
    if !strings.Contains(body, "Name is required") {
        t.Error("expected validation error for name field")
    }
}
```

### CSRF Validation Tests
```go
func TestCSRF_RejectsMissingToken(t *testing.T) { ... }
func TestCSRF_RejectsMismatchedToken(t *testing.T) { ... }
func TestCSRF_AcceptsValidToken(t *testing.T) { ... }
func TestCSRF_AllowsGETWithoutToken(t *testing.T) { ... }
func TestCSRF_LoginPOSTRequiresToken(t *testing.T) { ... }
func TestCSRF_HeaderAndFormBothAccepted(t *testing.T) { ... }
```

### Login Rate Limiting Tests
```go
func TestLogin_RateLimited(t *testing.T) {
    // Send 11 login attempts in quick succession — 11th should be rate-limited
    for i := range 11 {
        // ... POST /admin/login with bad credentials ...
        if i == 10 {
            if rec.Code != http.StatusTooManyRequests {
                t.Errorf("expected 429 on attempt %d, got %d", i+1, rec.Code)
            }
        }
    }
}
```

### Component Render Tests
```go
func TestDatatypesTableRows_RendersAllItems(t *testing.T) {
    items := []db.Datatypes{
        {DatatypeID: types.NewDatatypeID(), Name: "page", Label: "Page"},
        {DatatypeID: types.NewDatatypeID(), Name: "post", Label: "Blog Post"},
    }
    var buf bytes.Buffer
    err := partials.DatatypesTableRows(items).Render(context.Background(), &buf)
    if err != nil {
        t.Fatal(err)
    }
    body := buf.String()
    if !strings.Contains(body, "page") || !strings.Contains(body, "Blog Post") {
        t.Error("expected both items in rendered output")
    }
}
```

## DbDriver Prerequisites

The following `DbDriver` methods are required by admin handlers but may not exist yet. These must be added to the `DbDriver` interface and implemented across all three database wrappers (SQLite, MySQL, PostgreSQL) **before Phase 3 begins**. They can be added during Phase 1 or as a dedicated pre-Phase-3 task.

| Method | Signature | Used By |
|--------|-----------|---------|
| `ListDatatypesPaginated` | `(PaginationParams) (*[]Datatypes, error)` | `DatatypesListHandler` |
| `CountDatatypes` | `() (*int64, error)` | `DatatypesListHandler` |
| `ListFieldsPaginated` | `(PaginationParams) (*[]Fields, error)` | `FieldsListHandler` |
| `CountFields` | `() (*int64, error)` | `FieldsListHandler` |
| `ListUsersPaginated` | `(PaginationParams) (*[]Users, error)` | `UsersListHandler` |
| `CountUsers` | `() (*int64, error)` | `UsersListHandler` |
| `ListMediaPaginated` | `(PaginationParams) (*[]Media, error)` | `MediaListHandler` |
| `CountMedia` | `() (*int64, error)` | `MediaListHandler` |
| `ListRoutesPaginated` | `(PaginationParams) (*[]Routes, error)` | `RoutesListHandler` |
| `CountRoutes` | `() (*int64, error)` | `RoutesListHandler` |
| `GetDatatypeByName` | `(string) (*Datatypes, error)` | Uniqueness validation |

If an existing method already covers the use case (e.g., `ListDatatypes` returns all items and you paginate in Go), note that approach in the handler instead. But for large tables (content, media, audit), server-side pagination via SQL `LIMIT/OFFSET` is required.

**Adding a new DbDriver method requires:** SQL query in all three dialect files under `sql/schema/`, `just sqlc` to regenerate, interface addition in `internal/db/db.go`, wrapper implementations in `internal/db/*.go` for all three structs.

## Mutation Handler HTMX Requirement

All mutation handlers (POST/DELETE) require HTMX. Non-HTMX POST requests (e.g., JS disabled, automated tests, curl) should receive a meaningful response rather than failing silently. Standard pattern for mutation handlers:

```go
// For create/update handlers: non-HTMX POST gets a redirect on success, plain error on failure
if r.Header.Get("HX-Request") == "" {
    // Standard form submission fallback: redirect to list page on success
    http.Redirect(w, r, "/admin/schema/datatypes", http.StatusSeeOther)
    return
}
// HTMX: partial swap + toast
```

This means create/update handlers work with both HTMX and standard form submissions (progressive enhancement for writes). DELETE handlers remain HTMX-only and return `405 Method Not Allowed` for non-HTMX requests, since there is no standard HTML form for DELETE.

## Migration Strategy

### Coexistence During Development

During Phases 1-4, the new admin panel is registered alongside the existing SPA. The new routes are registered on `/admin/` — since Go's `ServeMux` uses most-specific-match routing, the new explicit routes (e.g., `GET /admin/content`) take priority over the old catch-all SPA handler. Pages not yet implemented fall through to the SPA. Phase 5 removes the SPA handler entirely.

The block editor JS file should be copied into `internal/admin/static/js/` during Phase 1 setup so that the content edit page (Phase 3, Batch A) is testable without waiting for Phase 4.

### Phase 1: Foundation
- [ ] Install templ CLI, add `just admin-generate` and `just admin-watch`
- [ ] Create `internal/admin/` package structure
- [ ] Create `embed.go` with `//go:embed static/*` directive
- [ ] Implement `csrf.go` — double-submit cookie CSRF protection
- [ ] Implement `render.go` — Render (buffer-then-write), RenderWithOOB (buffer-then-write), NewAdminData
- [ ] Implement `middleware.go` — AdminAuthMiddleware (with `?next=` redirect-back), CSRFMiddleware, login rate limiter
- [ ] Create `layouts/base.templ` (with cache-busted asset URLs), `layouts/admin.templ`, `layouts/auth.templ`
- [ ] Create `layouts/types.go` — AdminData struct
- [ ] Create `components/sidebar.templ`, `components/topbar.templ`, `components/icon.templ`
- [ ] Create `pages/login.templ` (with `next` hidden field) and `handlers/auth.go`
- [ ] Implement `admin.go` — route registration with permission middleware, static file serving with `Cache-Control` headers
- [ ] Wire login POST with CSRF + rate limiting: `loginLimiter(CSRFMiddleware()(handler))`
- [ ] Wire into `mux.go` — call `RegisterAdminRoutes` (coexists with SPA handler during transition)
- [ ] Vendor `htmx.min.js`
- [ ] Copy `block-editor.js` from external repo into `internal/admin/static/js/`
- [ ] Create `admin.css` with design tokens and base layout styles
- [ ] Create `admin.js` with HTMX config, CSRF header injection (for both HTMX and raw fetch)
- [ ] Create shared partials used by all batches: `partials/pagination.templ`, `partials/empty_state.templ`, `partials/form_field.templ`, `partials/toast.templ`, `partials/delete_confirm.templ`
- [ ] Add required paginated list / count methods to `DbDriver` interface and all three wrappers (see DbDriver Prerequisites above)
- [ ] Run `templ generate`, verify `go build` succeeds
- [ ] Write CSRF tests (including login POST)
- [ ] Write login rate limiting tests
- [ ] Write middleware tests (auth redirects, session expiry for HTMX vs regular)
- [ ] Write render tests (Render error handling, RenderWithOOB buffer-then-write)

### Phase 2: Web Components (all Light DOM)
- [ ] `mcms-toast.js` — toast notification system
- [ ] `mcms-dialog.js` — modal dialog
- [ ] `mcms-confirm.js` — inline confirmation button
- [ ] `mcms-data-table.js` — sortable/selectable table wrapper
- [ ] `mcms-search.js` — debounced search input
- [ ] `mcms-field-renderer.js` — field type input widgets
- [ ] `mcms-media-picker.js` — media browser/selector
- [ ] `mcms-tree-nav.js` — content tree navigation

### Phase 3: Admin Pages (parallelizable across agents)
Each page needs: handler, page `.templ`, partial `.templ`(s), handler tests (auth, permissions, partial-vs-full, CSRF, validation), route registration.

**Shared partials** (`pagination.templ`, `empty_state.templ`, `form_field.templ`, `toast.templ`, `delete_confirm.templ`) are created in Phase 1. All batches import them — do not recreate.

**Test harness:** Handler tests use a real SQLite `DbDriver` created via `db.NewTestDatabase()` (same pattern as existing `internal/db` tests). This is an integration test approach — no mock `DbDriver`. Each test file creates its own database in `testdb/`, seeds required data, and cleans up. Auth and permission context are injected via `middleware.WithAuthenticatedUser(ctx, user)` and `middleware.WithPermissions(ctx, perms)`.

**Batch A — Core content management:**
- [ ] Dashboard (`handlers/dashboard.go`, `pages/dashboard.templ`)
- [ ] Content list + partials
- [ ] Content edit with block editor
- [ ] Media list + grid partials
- [ ] Media detail

**Batch B — Schema management:**
- [ ] Datatypes list + detail
- [ ] Fields list + detail
- [ ] Field types list

**Batch C — Routes and users:**
- [ ] Routes list + admin routes list
- [ ] Users list + detail
- [ ] Roles list + permission list

**Batch D — Operations and tools:**
- [ ] Tokens list
- [ ] Sessions list
- [ ] Plugins list + detail
- [ ] Import page
- [ ] Audit log
- [ ] Settings

### Phase 4: Polish
- [ ] Light/dark mode toggle
- [ ] Keyboard shortcuts
- [ ] Loading indicators (HTMX events)
- [ ] Error states (HTMX `htmx:responseError`)
- [ ] Empty states for all list views
- [ ] Mobile responsive layout

### Phase 5: Cutover
- [ ] Remove SPA serving code from `mux.go` (lines 389-419)
- [ ] Remove `admin/embed.go` (replaced by `internal/admin/embed.go`)
- [ ] Move React `admin/` directory to `.claude/.trash/admin-react/`
- [ ] Verify `/` root redirect still points to `/admin/` (handled by `RegisterAdminRoutes`)
- [ ] Update CLAUDE.md to reflect templ-based admin architecture
- [ ] Update `justfile` — remove React build commands, add `admin-generate`
- [ ] Add CI step: `templ generate && git diff --exit-code internal/admin/`

## Key Decisions

### Why templ instead of html/template?
Compile-time type safety, function composition for layouts, no runtime parse errors, no FuncMap, no template registry. templ components are just Go functions — the compiler catches every mistake that `html/template` would only surface at runtime.

### Why commit generated `*_templ.go` files?
Same rationale as sqlc-generated code: `go build` works without the code generation tool installed. Only developers editing `.templ` files need `templ`. CI can verify generated files are up to date with `templ generate && git diff --exit-code`.

### Why keep the JSON API?
The JSON API serves three SDK consumers (TypeScript, Go, Swift) plus external developers. The admin HTML handlers are a second presentation layer over the same `DbDriver` operations.

### Why Light DOM for Web Components?
HTMX attribute inheritance does not cross Shadow DOM boundaries. Block editor is the exception (Shadow DOM, JSON API communication).

### Why double-submit cookie for CSRF?
No third-party dependency. Works with HTMX (header injection via `htmx:configRequest`) and standard forms (hidden field). `gorilla/csrf` is archived.

## File Count Estimate

| Category | Files | Notes |
|---|---|---|
| Go handlers | ~16 | One per resource area |
| Go infrastructure | ~5 | admin.go, render.go, csrf.go, middleware.go, embed.go |
| Layout templ | 3 | base, admin, auth |
| Page templ | ~22 | Full pages |
| Partial templ | ~20 | HTMX swap targets (includes form partials for validation) |
| Component templ | ~5 | sidebar, topbar, icon, data_table, status_badge |
| Generated Go | ~50 | *_templ.go (one per .templ file, committed) |
| Go tests | ~12 | csrf, middleware, render, auth, per-handler test files |
| Web Components JS | ~8 | mcms-* files (Light DOM) |
| Block editor | 1 | block-editor.js (Shadow DOM) |
| CSS | 1 | admin.css |
| JS glue | 1 | admin.js |
| Vendored libs | 1 | htmx.min.js |
| Layout types | 1 | layouts/types.go |
| **Total** | **~146** | (~50 are generated, ~96 hand-written) |
