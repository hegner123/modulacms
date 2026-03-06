# Redirects

Add a redirects system to ModulaCMS. Redirects are data, not behavior — the CMS
stores redirect mappings and returns them as structured responses. The client
decides how to execute them.

## Design Philosophy

ModulaCMS is a data authority, not a behavior authority. The slug resolution
endpoint returns a typed response: either content or a redirect instruction.
The CMS never issues HTTP 301/302 responses for redirect entries — that decision
belongs to the client application or SDK.

```
GET /api/v1/content/old-blog

CMS checks: redirects table -> routes table -> content tree

Response (redirect):
{
  "type": "redirect",
  "redirect": {
    "source": "/old-blog",
    "target": "/blog",
    "status_code": 301,
    "preserve_query": true
  }
}

Response (content):
{
  "type": "content",
  ...existing tree payload...
}
```

SDK users get convenience (auto-follow or typed result). Raw API users get
transparent data and full control.

## Schema

### redirects (public content redirects)

```sql
CREATE TABLE IF NOT EXISTS redirects (
    redirect_id TEXT PRIMARY KEY NOT NULL CHECK (length(redirect_id) = 26),
    source_path TEXT NOT NULL UNIQUE,
    target_path TEXT NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 301,
    active INTEGER NOT NULL DEFAULT 1,
    preserve_query INTEGER NOT NULL DEFAULT 1,
    note TEXT,
    author_id TEXT REFERENCES users(user_id) ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_redirects_source ON redirects(source_path);
CREATE INDEX IF NOT EXISTS idx_redirects_author ON redirects(author_id);

CREATE TRIGGER IF NOT EXISTS update_redirects_modified
    AFTER UPDATE ON redirects
    FOR EACH ROW
    BEGIN
        UPDATE redirects SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE redirect_id = NEW.redirect_id;
    END;
```

### admin_redirects (admin-side redirects)

Same structure with `admin_redirect_id` PK and `admin_` prefix, following the
existing dual-table convention (routes/admin_routes, content_data/admin_content_data).

### Fields

| Field | Type | Purpose |
|-------|------|---------|
| redirect_id | TEXT (ULID) | Primary key |
| source_path | TEXT UNIQUE | Incoming URL path to match (exact) |
| target_path | TEXT | Destination path or full URL |
| status_code | INTEGER | 301 (permanent), 302 (temporary), 307, 308 |
| active | INTEGER | Toggle without deleting (1=active, 0=inactive) |
| preserve_query | INTEGER | Pass through query string to target (1=yes, 0=no) |
| note | TEXT | Editorial note ("migrated from WP", "SEO rename") |
| author_id | TEXT | Who created it |
| date_created | TEXT | Timestamp |
| date_modified | TEXT | Timestamp |

### What's NOT in the table

- **hit_count / last_hit_at** — The CMS doesn't track analytics. Clients handle
  their own tracking. This is consistent with the data-not-behavior principle.
- **regex patterns** — Exact path matching only. Keeps the lookup O(1) from an
  in-memory map. Prefix matching is a potential future addition but not v1.
- **priority / ordering** — Source paths are unique, so there's no ambiguity.
  If prefix matching is added later, priority becomes necessary.

### Validation rules

- `source_path` must start with `/`
- `source_path` must not collide with an existing route slug
- `target_path` must not equal `source_path` (no self-redirects)
- `status_code` must be one of: 301, 302, 307, 308
- Creating a redirect for a path that already has a route should warn or fail

## Router Integration

### Redirect cache

In-memory `map[string]Redirect` loaded at startup, refreshed on mutation.
Follows the same pattern as `PermissionCache` — build-then-swap for lock-free reads.

```go
type RedirectCache struct {
    mu       sync.RWMutex
    bySource map[string]db.Redirects
}

func (rc *RedirectCache) Lookup(path string) (db.Redirects, bool)
func (rc *RedirectCache) Load(driver db.DbDriver) error
```

### Slug handler modification

In `internal/router/slugs.go`, redirect lookup happens before route resolution.
This is the insertion point at line 62:

```go
// Before:
route, err := d.GetRouteID(slug)

// After:
if redir, found := redirectCache.Lookup(slug); found && redir.Active == 1 {
    // Return redirect response (200 with type discriminator)
    writeRedirectResponse(w, redir)
    return nil
}
route, err := d.GetRouteID(slug)
```

The redirect response is a 200 OK with a JSON body. Not a 301/302 HTTP redirect.
The client interprets the response type and decides what to do.

### Response envelope

The slug endpoint currently returns a bare content tree. This needs a wrapper
to discriminate between content and redirect responses:

```go
type SlugResponse struct {
    Type     string    `json:"type"`               // "content" or "redirect"
    Redirect *Redirect `json:"redirect,omitempty"`
    // content fields remain at top level for backwards compatibility,
    // or nested under a "content" key (breaking change — needs decision)
}

type Redirect struct {
    Source        string `json:"source"`
    Target        string `json:"target"`
    StatusCode    int    `json:"status_code"`
    PreserveQuery bool   `json:"preserve_query"`
}
```

**Decision needed:** Should existing content responses be wrapped in a `"content"`
key, or should the `"type"` field be added alongside existing fields? The latter
is backwards-compatible but less clean. The former is a breaking change to the
content delivery API.

Option A (backwards-compatible):
```json
{ "type": "content", "data": { ...existing tree... } }
```

Option B (additive — no wrapper, just add "type" field):
```json
{ "type": "content", ...existing tree fields... }
```

Option C (redirect-only envelope — content stays unchanged):
Return redirects with a different HTTP status (e.g., 300) or a header
(`X-ModulaCMS-Redirect: true`) so clients can detect it without parsing.
Content responses remain exactly as they are today.

## API Endpoints

### CRUD

```
GET    /api/v1/redirects          — list all redirects
POST   /api/v1/redirects          — create redirect
GET    /api/v1/redirects/{id}     — get redirect by ID
PUT    /api/v1/redirects/{id}     — update redirect
DELETE /api/v1/redirects/{id}     — delete redirect
```

Permission resource: `redirects` (maps to `redirects:read`, `redirects:create`,
`redirects:update`, `redirects:delete`).

Admin variants:
```
GET    /api/v1/adminredirects     — list admin redirects
POST   /api/v1/adminredirects     — create admin redirect
...
```

### Bulk operations (future)

```
POST /api/v1/redirects/import     — bulk CSV/JSON import
GET  /api/v1/redirects/export     — bulk export
POST /api/v1/redirects/validate   — dry-run validation (check conflicts)
```

## SDK Changes

### TypeScript SDK

```typescript
// Types package
interface Redirect {
  redirect_id: RedirectID;
  source_path: string;
  target_path: string;
  status_code: 301 | 302 | 307 | 308;
  active: boolean;
  preserve_query: boolean;
  note: string | null;
  author_id: UserID | null;
  date_created: string;
  date_modified: string;
}

// Content delivery SDK — slug resolution returns union type
type SlugResult =
  | { type: "content"; data: ContentTree }
  | { type: "redirect"; redirect: RedirectInfo };

// Admin SDK — CRUD
client.redirects.list()
client.redirects.get(id)
client.redirects.create({ source_path, target_path, ... })
client.redirects.update(id, { ... })
client.redirects.delete(id)
```

### Go SDK

```go
type SlugResult struct {
    Type     string        `json:"type"`
    Content  *ContentTree  `json:"data,omitempty"`
    Redirect *RedirectInfo `json:"redirect,omitempty"`
}

// Resource[Redirect, CreateRedirectParams, UpdateRedirectParams, RedirectID]
client.Redirects.List(ctx, params)
client.Redirects.Create(ctx, params)
```

### Swift SDK

Same pattern — `SlugResult` enum with associated values.

## TUI Screen

### Grid layout

```
+----------------+--------------------------------------+
|   Redirects    |  Details                             |
|                |  source, target, status, query,      |
|  /old  -> /new |  note, author, dates                 |
|  /foo  -> /bar |                                      |
|  (4)           +--------------------------------------+
|                |  Info                                 |
|                |  status code explanation,             |
|                |  active/inactive, conflict warnings   |
+----------------+--------------------------------------+
```

```go
var redirectsGrid = Grid{
    Columns: []GridColumn{
        {Span: 4, Cells: []GridCell{
            {Height: 1.0, Title: "Redirects"},
        }},
        {Span: 8, Cells: []GridCell{
            {Height: 0.60, Title: "Details"},
            {Height: 0.40, Title: "Info"},
        }},
    },
}
```

- Column 1 (span 4): scrollable redirect list with source -> target preview.
  Wider than media/content (span 3) because paths need room.
  Inactive redirects shown dimmed.
- Column 2 top (60%): full details — source path, target path, status code,
  preserve query, note, author, created, modified, ID
- Column 2 bottom (40%): contextual info — human-readable status code meaning
  ("301 Permanent — search engines update their index"), active/inactive state,
  warnings (e.g., "target path does not match any route")

### List rendering

```
 -> /old-blog        -> /blog          301
    /legacy/about    -> /about         301
    /temp-landing    -> /promo         302  (inactive)
```

Each line: cursor, source (left-aligned), target (right portion), status code.
Inactive entries rendered in dim style.

### Inline search

`/` enters search mode (same as media screen). Filters by source_path and
target_path. Uses `textinput.Model` and `config.ActionSearch`.

### Key bindings

```go
func (s *RedirectsScreen) KeyHints(km config.KeyMap) []KeyHint {
    if s.Searching {
        return []KeyHint{
            {"type", "filter"},
            {"enter", "accept"},
            {"esc", "clear"},
        }
    }
    return []KeyHint{
        {km.HintString(config.ActionSearch), "search"},
        {km.HintString(config.ActionSelect), "select"},
        {km.HintString(config.ActionNew), "new"},
        {km.HintString(config.ActionEdit), "edit"},
        {km.HintString(config.ActionDelete), "del"},
        {km.HintString(config.ActionNextPanel), "panel"},
        {km.HintString(config.ActionBack), "back"},
        {km.HintString(config.ActionQuit), "quit"},
    }
}
```

### Form dialog

Create/edit redirect uses the existing form dialog system:

- Source path (text input, required, must start with `/`)
- Target path (text input, required)
- Status code (select: 301, 302, 307, 308)
- Preserve query (toggle: yes/no)
- Note (text input, optional)
- Active (toggle: yes/no — edit only)

### Struct

```go
type RedirectsScreen struct {
    GridScreen
    AdminMode    bool
    Cursor       int
    Redirects    []db.Redirects
    AdminRedirects []db.AdminRedirects
    Searching    bool
    SearchInput  textinput.Model
    SearchQuery  string
    FilteredList []db.Redirects          // or admin variant
}
```

## Nav Changes

Replace the routes screen entry in the TUI navigation with redirects:

```
Before:  Home | Content | Routes | Admin Routes | Media | ...
After:   Home | Content | Redirects | Media | ...
```

Routes CRUD moves into the content screen as actions on route nodes in the
select tree (edit slug/title, delete route). Admin routes follow the same
pattern for admin content.

The routes screen code (`screen_routes.go`) is not deleted — it can be kept
for the admin panel or as a fallback, but removed from the TUI nav menu.

## Implementation Steps

### Phase 1: Schema and database layer

1. Create `sql/schema/N_redirects/` with schema + queries for all 3 backends
2. Create `sql/schema/N_admin_redirects/` (same pattern)
3. Run `just sqlc` to generate Go code
4. Run `just dbgen-entity Redirects` and `just dbgen-entity AdminRedirects`
5. Add methods to `DbDriver` interface and all three wrapper structs
6. Add `CreateAllTables` entries

### Phase 2: Redirect cache and slug handler

7. Create `internal/middleware/redirect_cache.go` — RedirectCache with
   Load, Lookup, periodic refresh
8. Modify `internal/router/slugs.go` — add redirect check before route lookup
9. Define response envelope type (pending decision on Option A/B/C)
10. Wire RedirectCache into server startup in `cmd/serve.go`
11. Add `redirects` permission resource to bootstrap data

### Phase 3: API endpoints

12. Create `internal/router/redirects.go` — CRUD handlers
13. Register routes in `internal/router/mux.go` with permission middleware
14. Write API tests

### Phase 4: TUI screen

15. Create `internal/tui/screen_redirects.go` — struct, Update, constructor
16. Create `internal/tui/screen_redirects_view.go` — View, render methods
17. Add `REDIRECTS` page index to `pages.go`
18. Add form dialog types for create/edit redirect
19. Add fetch/result messages
20. Update navigation — replace routes with redirects in menu
21. Write screen tests

### Phase 5: Route management migration

22. Add route edit action to content select tree (edit slug/title from content screen)
23. Add route delete action to content select tree
24. Remove routes/admin routes from TUI nav menu
25. Keep routes screen code for admin panel use

### Phase 6: SDK updates

26. Add Redirect types to all three SDK type packages
27. Add SlugResult union type to content delivery SDKs
28. Add Redirects resource to admin SDKs
29. Update SDK tests

### Phase 7: Admin panel

30. Add redirects page to admin panel (HTMX + templ)
31. Add redirect CRUD partials
32. Register admin routes in `registerAdminRoutes()`

## Open Questions

1. **Response envelope** — Option A, B, or C? This affects backwards
   compatibility for existing API consumers.

2. **Route collision detection** — Should creating a redirect for a path that
   matches an existing route slug be an error, a warning, or allowed (redirect
   takes precedence)?

3. **Admin redirects** — Do admin routes need their own redirect table, or is
   one redirect table sufficient? Admin content uses separate admin_routes, but
   redirects might be simpler since they're not tied to content trees.

4. **Prefix matching** — Is exact-only sufficient for v1, or do users need
   `/old-section/* -> /new-section/*` from day one?

5. **External targets** — Should `target_path` support full URLs
   (`https://example.com/page`) for cross-domain redirects, or only relative
   paths within the same CMS?

## Files

| File | Change |
|------|--------|
| `sql/schema/N_redirects/` | **NEW** — schema + queries (3 backends) |
| `sql/schema/N_admin_redirects/` | **NEW** — schema + queries (3 backends) |
| `internal/db/redirect_gen.go` | **NEW** — generated by dbgen |
| `internal/db/admin_redirect_gen.go` | **NEW** — generated by dbgen |
| `internal/db/db.go` | Add redirect methods to DbDriver interface |
| `internal/middleware/redirect_cache.go` | **NEW** — in-memory cache |
| `internal/router/slugs.go` | Add redirect check before route lookup |
| `internal/router/redirects.go` | **NEW** — CRUD handlers |
| `internal/router/mux.go` | Register redirect endpoints |
| `internal/tui/screen_redirects.go` | **NEW** — TUI screen |
| `internal/tui/screen_redirects_view.go` | **NEW** — View rendering |
| `internal/tui/pages.go` | Add REDIRECTS page index |
| `internal/tui/messages.go` | Add fetch/result messages |
| `internal/tui/update_dialog.go` | Add redirect form dialog handling |
| `cmd/serve.go` | Wire RedirectCache into server startup |
| `sdks/typescript/types/` | Add Redirect types |
| `sdks/typescript/modulacms-sdk/` | Add SlugResult type |
| `sdks/typescript/modulacms-admin-sdk/` | Add Redirects resource |
| `sdks/go/types.go` | Add Redirect types |
| `sdks/go/modulacms.go` | Add Redirects resource |
| `sdks/swift/Types.swift` | Add Redirect types |
