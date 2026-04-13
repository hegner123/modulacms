# Admin Content Tree Rewrite

Replace the flat-list content management (`content_list` + `content_edit` + standalone block editor) with a two-panel layout driven by a route-level tree sidebar, a combined page view, and a slide-in field editing drawer.

**Scope:** Public content only (`content_data`, `routes`). Admin content (`admin_content_data`, `admin_routes`) is deferred to a follow-up phase after the public content tree is proven stable. The admin variant requires different ID types (`AdminContentID` vs `ContentID`), different handler signatures, and different DB methods, so it is a separate implementation pass, not a find-and-replace.

## Current Architecture (What Exists)

**Handlers:**
- `ContentListHandler` in `internal/admin/handlers/content.go` (line 72) -- flat paginated list with search
- `ContentEditHandler` (line 473) -- loads single content node, builds tree JSON via `buildTreeJSON`, renders `<block-editor>` + field panel
- `ContentTreeSaveHandler` (line 1057) -- bulk creates/updates/deletes from block editor
- `ContentCreateHandler` (line 583) -- creates content + optional route + fields
- `ContentReorderHandler` (line 843), `ContentMoveHandler` (line 921) -- tree mutations

**Templates:**
- `content_list.templ` -- table-based list with search, status filter, pagination, create dialog (dropdown select for datatype)
- `content_edit.templ` -- `<block-editor>` tree panel (left) + `<mcms-field-renderer>` field panel (right), locale tabs, version history dialog, publish controls

**Web Components:**
- `<block-editor>` (95KB bundled, `block-editor-src/`) -- full in-memory tree state with pointer-based linked list, drag-and-drop, datatype picker, save diffing, history/undo
- `<mcms-tree-nav>` (39KB, `components/mcms-tree-nav.js`) -- server-rendered `<ul>/<li>` tree with lazy loading, drag-and-drop reorder/reparent, active node highlight. Already has lazy-load via `htmx.ajax('GET', '/admin/content/tree/' + nodeId + '/children')`
- `<mcms-field-renderer>` (56KB) -- renders all field types (text, richtext, media, reference, plugin, etc.)
- `<mcms-dialog>` (13KB) -- modal wrapper with focus trap, auto-close on HTMX success

**Patterns:**
- HTMX navigation: `IsNavHTMX` returns true when `HX-Target == "main-content"`. Content-only component rendered + `HX-Trigger: pageTitle` for title update.
- OOB swaps via `RenderWithOOB` for updating dialogs alongside main content.
- `AdminData` layout struct carries title, permissions, CSRF, user, version.
- Media folder tree (`partials/media_folder_tree.templ`) is the existing reference for templ-based tree sidebar layout. Note: it does NOT use `<mcms-tree-nav>` -- it is a hand-built templ tree. The new content tree sidebar will be the first templ-rendered usage of `<mcms-tree-nav>`.

## Design

### Two-Panel Layout

**Panel 1 (Sidebar/Tree) -- two modes:**

1. **Route Tree Mode (default):** Top level shows routes as tree nodes. Routed content with page icons, unrouted/global content with globe icons below a separator. Search input at top for filtering. Expanding a route lazy-loads its content tree via HTMX.

2. **Block Tree Mode (after selecting a page):** Sidebar transforms to show the block-level tree for the selected page. "Back to routes" link at top returns to route tree mode. This mirrors the TUI's content navigation.

**Panel 2 (Main Content Area) -- two modes:**

1. **Page View:** After clicking a leaf content node, the main panel shows a combined view of ALL content blocks as summary cards. Each card shows datatype label and truncated field previews. Cards are ordered by tree traversal (DFS). Indentation reflects tree depth.

2. **Drawer Edit:** Clicking a block card slides in a drawer from the right with `<mcms-field-renderer>` elements for that block's fields. Save updates the block and refreshes the card in-place via OOB swap.

### Interaction Model (Routed Content)

```
Route Tree          Block Tree              Page View                Drawer
(sidebar)           (sidebar)               (main panel)             (slide-in)

 /blog         ->   ▾ Hero          ->     [Hero Card]         ->   [title: ___]
 /about              ▾ Row                 [Row Card]               [subtitle: ___]
 /pricing              Card                  [Card Card]            [image: ___]
 ─────────             Card                  [Card Card]            [Save] [Cancel]
 Navigation            Text Block          [Text Block Card]
 Footer              ▸ Sidebar
```

### Interaction Model (Unrouted/Global Content)

Unrouted content (navigation menus, footers, shared blocks) uses `_global` or `_reference` datatypes and has no route association. These appear below a "Site-wide" separator in the sidebar with a globe icon instead of a page icon.

**Clicking an unrouted node skips the route-expansion step** and goes directly to the block-tree + page-view mode. The unrouted node IS the root of its content tree (there is no route to expand first). The sidebar transitions to block-tree mode showing the global node's children. The main panel shows the page view with block summary cards.

```
Route Tree          Block Tree              Page View                Drawer
(sidebar)           (sidebar)               (main panel)             (slide-in)

 /blog               ▾ Menu Link 1  ->     [Menu Link 1 Card]  ->  [label: ___]
 /about              ▾ Menu Link 2         [Menu Link 2 Card]      [url: ___]
 /pricing              Submenu             [Submenu Card]          [Save] [Cancel]
 ─── Site-wide ───
>Navigation      ->  Back to routes
 Footer
```

The breadcrumb for unrouted content reads: `Content Tree > Navigation > Menu Link 1` (no route slug segment).

### Sidebar State Reconstruction

When the user directly navigates to `/admin/content-tree/page/{contentID}` (via URL, browser refresh, or deep link), the server reconstructs the sidebar in block-tree mode for that content node's tree. The handler loads the content node, reads its `root_id` field to find the tree root in one query (do not walk `parent_id` up the chain). If `root_id` is null, the content node IS the root. Load the root's `route_id` to resolve the route slug for the breadcrumb. For unrouted content (`route_id` is null), use the root node's datatype label as the breadcrumb segment. The sidebar renders the block-tree with the current node highlighted. The "Back to routes" link is always present in block-tree mode.

### Atomicity Model

The new design saves one block's fields at a time via the drawer, not a batched save of all changes. This is the same atomicity model as the TUI. Implications:

- **Partial saves are expected.** Editing block A and saving, then editing block B and failing validation, leaves the tree with A updated and B unchanged. This is acceptable because each save is a complete, valid operation on a single block.
- **Content versions are not created per-field-save.** Versions are created by the explicit "Publish" action, which snapshots the entire tree as-is. The drawer save updates live content fields; the publish flow captures a point-in-time snapshot.
- **Tree structure changes (reorder, reparent, delete) are separate operations** from field edits, each with their own server round-trip.

## Phase 0: Foundation (Layout Shell + Types + Routes)

**Goal:** Create the two-panel layout container, new page-level types, and new route registrations without breaking existing pages. Existing content_list and content_edit continue to work during development.

### Files to create

`internal/admin/pages/content_tree.templ` -- New top-level page template with two-panel layout:
- Outer container: `flex` layout with sidebar (w-72 shrink-0) and main area (flex-1)
- Sidebar slot: `<div id="content-tree-sidebar">` for route tree
- Main slot: `<div id="content-main-panel">` for page view / empty state
- Breadcrumb bar above main panel: `<nav id="content-breadcrumb">`
- OOB targets for drawer and dialogs

`internal/admin/pages/content_tree_types.go` -- Types consumed by templ templates:
- `RouteTreeNode`: RouteID, Slug, Title, DatatypeLabel, HasChildren bool, IsGlobal bool, ContentDataID
- `ContentBlockSummary`: ContentDataID, DatatypeLabel, DatatypeType, Fields []BlockFieldSummary, ChildCount int
- `BlockFieldSummary`: Label, Type, Value (truncated for preview)
- `ContentBreadcrumb`: segments for route > page > block

Handler-internal types (defined in `internal/admin/handlers/content_tree.go`, not the pages package):
- `DrawerFieldData`: wraps existing `db.ContentFieldWithFieldRow` + metadata for drawer rendering

### New routes

Add to `internal/router/mux.go` inside `registerAdminRoutes()`:

```
GET  /admin/content-tree                          viewing("content:read", ContentTreePageHandler(driver, mgr))
GET  /admin/content-tree/sidebar                  viewing("content:read", ContentTreeSidebarHandler(driver, mgr))
GET  /admin/content-tree/sidebar/{routeID}        viewing("content:read", ContentTreeRouteChildrenHandler(driver))
GET  /admin/content-tree/page/{contentID}         viewing("content:read", ContentTreePageViewHandler(driver, mgr))
GET  /admin/content-tree/drawer/{contentID}       viewing("content:read", ContentTreeDrawerHandler(driver, mgr))
POST /admin/content-tree/drawer/{contentID}       mutating("content:update", ContentTreeDrawerSaveHandler(driver, mgr))
POST /admin/content-tree/create                   mutating("content:create", ContentTreeCreateHandler(driver, mgr))
GET  /admin/content-tree/create-options/{parentID} viewing("content:read", ContentTreeCreateOptionsHandler(driver))
DELETE /admin/content-tree/block/{contentID}       mutating("content:delete", ContentTreeDeleteBlockHandler(svc, mgr))
```

Handler signatures:
```go
ContentTreePageHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc
ContentTreeSidebarHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc
ContentTreeRouteChildrenHandler(driver db.DbDriver) http.HandlerFunc
ContentTreePageViewHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc
ContentTreeDrawerHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc
ContentTreeDrawerSaveHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc
ContentTreeCreateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc
ContentTreeCreateOptionsHandler(driver db.DbDriver) http.HandlerFunc
ContentTreeDeleteBlockHandler(svc *service.Registry, mgr *config.Manager) http.HandlerFunc
```

### New SQL query (prerequisite for Phase 2)

Add `ListContentFieldsByContentIDs` to `sql/schema/17_content_fields/`. This batch query loads all content fields (with field metadata) for a list of content_data_id values in a single query, eliminating the N+1 pattern where `buildTreeJSON` fires one `ListContentFieldsWithFieldByContentData` per block. Run `just sqlc` after adding.

SQLite (`queries.sql`):
```sql
-- name: ListContentFieldsByContentIDs :many
SELECT cf.*, f.label as field_label, f.type as field_type, f.config as field_config
FROM content_fields cf
LEFT JOIN fields f ON cf.field_id = f.field_id
WHERE cf.content_data_id IN (sqlc.slice('content_ids'));
```

MySQL (`queries_mysql.sql`): Same as SQLite (both use `sqlc.slice`).

PostgreSQL (`queries_psql.sql`):
```sql
-- name: ListContentFieldsByContentIDs :many
SELECT cf.*, f.label as field_label, f.type as field_type, f.config as field_config
FROM content_fields cf
LEFT JOIN fields f ON cf.field_id = f.field_id
WHERE cf.content_data_id = ANY($1::text[]);
```

### Files to modify

- `internal/admin/static/css/input.css` -- Add drawer z-index token: `--z-drawer: 90;` (sits between `--z-sidebar: 60` and `--z-dialog: 100` so confirmation dialogs opened from the drawer appear on top). Verify that `z-[var(--z-drawer)]` compiles in `just admin generate`. If Tailwind does not resolve CSS variable references in arbitrary values, use `z-[90]` directly.

**Checkpoint:** `just check` compiles. New `/admin/content-tree` route serves a skeleton two-panel page with empty sidebar and main area. Existing `/admin/content` still works unchanged.

## Phase 1: Route Tree Sidebar (HTMX + templ)

**Goal:** The sidebar shows all routes as a collapsible tree with search. Routed content first with page icons, unrouted/global content below a separator with globe icons.

### Files to create

`internal/admin/handlers/content_tree.go` -- New handler file:
- `ContentTreePageHandler` -- Loads route tree, renders full page or HTMX partial
- `ContentTreeSidebarHandler` -- Returns sidebar HTML (for search filtering)
  - Calls `driver.ListRoutes()` for all routes
  - Groups into routed (with route slug/title) and unrouted (globals, menus)
  - Search query param filters by title/slug substring match in Go (all routes are loaded regardless for the tree; in-memory filtering is fine at this scale). Do NOT add a new DB query for search in this phase.
- `ContentTreeRouteChildrenHandler` -- Lazy-loads content tree under a route
  - Takes route ID from path, builds nested `<ul>/<li>` markup from tree

`internal/admin/partials/content_tree_sidebar.templ` -- Sidebar partial templates:
- `ContentTreeSidebar(routes []RouteTreeNode, globals []RouteTreeNode, search string, csrfToken string)`
- Search input: `hx-get="/admin/content-tree/sidebar"`, `hx-target="#content-tree-sidebar"`, `hx-trigger="input changed delay:300ms"`
- Routed section: each route in `<mcms-tree-nav>` with lazy-load
- Separator + global section with globe icons
- "+" button on each tree node for contextual create

`internal/admin/partials/content_tree_children.templ` -- Lazy-loaded subtree partial:
- `ContentTreeChildren(nodes []ContentTreeChildNode)` -- Nested `<ul>/<li>` for a route's content tree

### Files to modify

`internal/admin/static/js/components/mcms-tree-nav.js`:
- Add configurable URL pattern via `data-children-url` attribute. Format: URL template with `{id}` placeholder, e.g., `data-children-url="/admin/content-tree/sidebar/{id}"`. JS replaces `{id}` with `encodeURIComponent(nodeId)`. Falls back to current hardcoded URL (`/admin/content/tree/{id}/children`) when attribute is absent.
- Add `data-leaf` boolean attribute support on `<li>` nodes. When present: click fires `hx-get="/admin/content-tree/page/{nodeId}"` targeting `#content-main-panel` with `hx-push-url="true"`. No expand/collapse toggle rendered. When absent: current expand/collapse behavior unchanged.

**Checkpoint:** `/admin/content-tree` shows the two-panel layout with a populated route tree. Expanding a route lazy-loads its content subtree. Search filters routes. Clicking leaf nodes is wired but shows empty main panel.

## Phase 2: Combined Page View (Block Summary Cards)

**Goal:** Clicking a leaf content node loads a visual summary of ALL blocks into the main panel. The sidebar transforms to show the block-level tree.

### Files to create

`internal/admin/partials/content_page_view.templ` -- Page view partial:
- Header: page title, status badge, publish button, version history menu
- Block cards: iterate blocks in DFS tree order, render each as a card
- Each card: `data-content-id`, click opens drawer
- Cards show: datatype label badge, truncated field previews, child count indicator
- Indentation via `margin-left` based on tree depth
- "+" buttons between/after cards for adding new blocks

`internal/admin/partials/content_block_card.templ` -- Individual block card:
- `ContentBlockCard(block ContentBlockSummary, depth int)` -- Single block summary
- Datatype label as colored badge
- First 2-3 field values as truncated previews
- Hover shows "Edit" overlay
- Click: `hx-get="/admin/content-tree/drawer/{contentID}"` into `#content-drawer`

`internal/admin/partials/content_breadcrumb.templ`:
- `ContentBreadcrumb(crumbs ContentBreadcrumb)` -- `Content Tree > /blog > Blog Post > [block name]`
- Each segment is a clickable link

### Handler additions to content_tree.go

`ContentTreePageViewHandler`:
- Loads content node by ID
- Factor out `buildPageBlocks(driver, contentID, roleID, isAdmin) []ContentBlockSummary`
- Walks tree in DFS order, collecting block summaries with truncated field previews
- Returns `ContentPageView` partial for HTMX
- Response includes OOB swap for sidebar: transforms to block-level tree with "Back to routes" link

### Tree Structure Mutations

The sidebar block-tree and page view both expose tree mutation controls:

**Delete:** Each block card has a delete icon (visible on hover). Clicking opens an `mcms-dialog` confirmation ("Delete this block and its N children?"). On confirm, `hx-delete="/admin/content-tree/block/{contentID}"` fires. The handler must recursively delete children in application code (walk the tree, delete leaves first) because the `content_data` self-referential FKs use `ON DELETE SET NULL`, not CASCADE. Deleting a parent without deleting children first would orphan them. `content_fields.content_data_id` uses `ON DELETE CASCADE` (deleting a content_data row deletes its direct fields). However, `content_fields.root_id` uses `ON DELETE SET NULL`, so always delete leaf nodes first (bottom-up) to ensure each node's direct fields cascade-delete before the parent is removed. Server returns OOB swap removing the card from the page view and the node from the sidebar tree.

**Reorder:** Drag-and-drop in the sidebar block-tree (already supported by `mcms-tree-nav`). On drop, `mcms-tree-nav` fires the existing reorder endpoint (`POST /admin/content/reorder`, hardcoded at mcms-tree-nav.js:894). Server returns OOB swap refreshing the page view card order.

**Reparent/Move:** Drag a node in the sidebar block-tree to a new parent (already supported by `mcms-tree-nav`). On drop, fires the existing move endpoint (`POST /admin/content/move`, hardcoded at mcms-tree-nav.js:720). Server returns OOB swap refreshing both the sidebar tree and the page view.

Note: The move/reorder URLs in `mcms-tree-nav.js` are hardcoded (unlike the children-URL which Phase 1 makes configurable). This is intentional -- the new content tree pages operate on the same `content_data` table and reuse the same backend handlers. No URL change needed.

All structural mutations refresh the page view via OOB swap to keep card order in sync with the tree.

### Error and Empty States

- **No blocks:** Page view shows an empty state card ("This page has no content blocks yet. Click + to add one.") using the existing `partials/empty_state.templ` pattern.
- **Drawer load failure:** If the drawer GET returns an error, show a toast with the error message. Do not open an empty drawer.
- **Drawer save validation error:** Server returns 422 with field-level error messages. The drawer stays open and renders inline errors next to the relevant `<mcms-field-renderer>` elements (matching the existing content_edit validation pattern).
- **Tree load failure:** If lazy-loading a route's children fails, show an inline error message in the tree node ("Failed to load. Click to retry.") with `hx-trigger="click"` to re-attempt.
- **No routes:** Sidebar shows an empty state ("No routes found. Create a route first.") with a link to `/admin/routes`.

**Checkpoint:** Clicking a leaf content node shows block summary cards. Sidebar shows block-level tree. Breadcrumb shows navigation path. Cards are clickable (drawer not yet implemented). Tree mutations (delete, reorder, move) work via existing handlers with OOB refreshes.

## Phase 3: Field Edit Drawer

**Goal:** Clicking a block card slides in a drawer from the right with field editors. Save updates the block and refreshes the page view.

### Files to create

`internal/admin/static/js/components/mcms-drawer.js` -- Drawer web component (NEW):
- Light DOM component (matches project convention)
- `open()` / `close()` methods
- Slide-in animation via CSS transitions (`transform: translateX(100%)` -> `translateX(0)`)
- Optional backdrop (semi-transparent)
- Escape key closes
- `mcms-drawer:open` / `mcms-drawer:close` custom events
- Auto-close on HTMX success (same pattern as mcms-dialog)
- Focus trap within drawer when open

`internal/admin/partials/content_drawer.templ` -- Drawer partial:
- `ContentDrawer(content db.ContentData, fields []db.ContentFieldWithFieldRow, meta DrawerMeta)`
- Fixed position, right-0, full height, w-96, z-[var(--z-drawer)]
- Header: block datatype label, close button
- Body: `<form>` with `<mcms-field-renderer>` for each field (reuse pattern from content_edit.templ)
- Footer: Save button, Cancel button
- Save: `hx-post="/admin/content-tree/drawer/{contentID}"` with `hx-swap="none"` (server returns OOB updates)
- On success: close drawer, OOB-swap updated block card, show toast

### Handler additions to content_tree.go

`ContentTreeDrawerHandler` (GET) -- Loads content fields, filtered by role, returns `ContentDrawer` partial

`ContentTreeDrawerSaveHandler` (POST) -- The drawer form uses `name="field_{content_field_id}"` for each field input (matching `mcms-field-renderer`'s existing name attribute convention). The handler loads the content node's fields from DB, iterates them, and reads `r.FormValue("field_" + cf.ContentFieldID.String())` for each. Fields not present in the form submission are skipped (not zeroed). Returns OOB swaps:
- Updated block card: `<div id="block-card-{id}" hx-swap-oob="true">`
- Toast trigger: `HX-Trigger: showToast`
- Drawer close trigger: `HX-Trigger: drawerClose`

### Files to modify

- `internal/admin/static/css/input.css` -- Add drawer styles
- `internal/admin/layouts/base.templ` -- Add `<script>` for mcms-drawer.js

**Checkpoint:** Full edit flow works: navigate tree -> select page -> click block -> edit fields in drawer -> save -> see updated card. Drawer opens/closes with animation.

## Phase 4: Create Flow (Datatype Cards)

**Goal:** "+" buttons open a contextual create panel showing datatype cards. Selecting a card creates the content node and navigates to it.

### Files to create

`internal/admin/partials/content_create_cards.templ` -- Datatype card grid:
- `ContentCreateCards(datatypes []db.Datatypes, parentID string, csrfToken string)`
- Grid of cards: datatype label, type badge, description
- Clicking a card: `hx-post="/admin/content-tree/create"` with hidden fields for parent_id, datatype_id
- Cards are contextual to parent datatype

### Handler additions to content_tree.go

`ContentTreeCreateOptionsHandler` -- Returns contextual datatype cards:
- Loads parent content node and its datatype
- Falls back to `driver.ListDatatypesRoot()` if no constraints
- Returns `ContentCreateCards` partial

`ContentTreeCreateHandler` -- Creates content and redirects:
- Extract the content creation core (content_data INSERT + route association + field scaffolding) from `ContentCreateHandler` (lines 583-708) into a shared helper function `createContentNode(ctx, driver, ac, params)`. The new handler calls this helper with `parent_id` and `datatype_id` from the POST body. Do NOT duplicate the existing handler's form-parsing logic for slug, title, or field values -- the card-based flow does not collect those.
- Auto-populate slug from parent route's slug
- On success: redirect to page view for the new content

**Checkpoint:** "+" buttons show datatype cards. Selecting a card creates content with scaffolded fields. Full create-edit cycle works.

## Phase 5: Transition and Cleanup

**Goal:** Wire the new content tree as the primary interface. Redirect old URLs.

### Files to modify

`internal/admin/components/nav.go` -- Change Content nav item href to `/admin/content-tree`

`internal/admin/static/js/components/mcms-command-palette.js` -- Update hardcoded `/admin/content` URL (line 220) to `/admin/content-tree` so the command palette navigates directly without a 302 redirect.

`internal/router/mux.go`:
- Replace the existing `mux.Handle("GET /admin/content", ...)` and `mux.Handle("GET /admin/content/{id}", ...)` handler registrations with 302 redirect handlers pointing to `/admin/content-tree` and `/admin/content-tree/page/{id}` respectively. Do not register both old handlers and redirects at the same path (Go ServeMux does not allow duplicate patterns).
- API routes (`/api/v1/content/...`) are unchanged.

`internal/admin/static/js/admin.js` -- Field saves now go through drawer's `hx-post`, not client-side state diffing. Tree structure changes (reorder, reparent) still use existing handlers via HTMX from sidebar tree's drag-and-drop.

### Deprecate (keep but mark)

- `pages/content_list.templ` -- "Deprecated: replaced by content_tree.templ"
- `pages/content_edit.templ` -- "Deprecated: replaced by content_tree page view + drawer"
- Old handler functions in `content.go` -- Keep functional, mark deprecated

**Checkpoint:** Content nav opens tree view. Old URLs redirect. All operations work through new UI.

## Phase 6: Block Editor Retirement

**Goal:** The existing `<block-editor>` JS (95KB) is no longer used by the new design.

**What the block editor currently does:**
1. Renders block tree as a list (tree-ops.js, dom-patches.js) -> **replaced by server-side rendering + mcms-tree-nav**
2. Drag-and-drop reorder/reparent (drag.js) -> **replaced by mcms-tree-nav drag-and-drop**
3. Datatype picker for adding blocks (picker.js, cache.js) -> **replaced by datatype cards partial**
4. Client-side tree state with undo/redo (state.js, history.js) -> **no longer needed (server-authoritative)**
5. Diffs state against baseline for save (index.js) -> **replaced by per-field drawer saves**
6. Validates tree integrity (validate.js) -> **no longer needed client-side**

**Decision:** Do NOT remove the block editor code yet. The new content tree view does not load it. The old `content_edit.templ` remains functional as a fallback. Future work can fully retire `<block-editor>` once the new flow is proven stable.

`internal/admin/pages/content_tree.templ` -- Does NOT include `block-editor.js` or `block-editor.css`.

Both `block-editor.css` (line 13) and `block-editor.js` (line 18, 95KB) are currently loaded unconditionally in `base.templ`. Move both the `<script type="module" src="block-editor.js">` tag and `<link rel="stylesheet" href="block-editor.css">` from `base.templ` into `content_edit.templ` directly (the only consumer). `<script type="module">` works in `<body>` per HTML spec. This avoids changing `Base()`'s signature, which would require updating every page that calls it.

**Checkpoint:** New content tree pages load without block editor bundle. Page weight is significantly reduced.

## Parallelization Guide

| Agent | Phase | Dependencies |
|-------|-------|-------------|
| A | Phase 0 (layout shell + types + routes) | None |
| B | Phase 1 (sidebar handler + partials) | Phase 0 |
| C | Phase 3 mcms-drawer.js web component | None (standalone) |
| D | Phase 3 drawer CSS | None (standalone) |

After Phase 0:
- Phase 1 and Phase 2 can run in parallel
- Phase 3 JS component can be built independently of Phase 1/2
- Phase 4 depends on Phase 1 (sidebar) + Phase 2 (page view)
- Phase 5 depends on all prior phases
- Phase 6 is informational/cleanup, can run anytime after Phase 2

## Risk Assessment

1. **mcms-tree-nav URL hardcoding** -- Lazy-load URL is hardcoded at line 375. Must be made configurable via `data-children-url` attribute before Phase 1. Low risk.

2. **Two tree contexts in sidebar** -- Sidebar alternates between route-level and block-level tree. Handle via OOB swap of entire sidebar content (matches existing `RenderWithOOB` pattern).

3. **Drawer is a new component** -- No slide-in panel exists yet. Adapt focus trapping and HTMX integration from `mcms-dialog`.

4. **Field save atomicity** -- Current block editor batches all changes into one POST. New drawer saves individual blocks' fields. This is simpler and more reliable (no complex diffing), and matches how the TUI works. Tree structure changes and field changes are separate operations.

5. **Server-authoritative tree** -- Every tree mutation requires a server round-trip. This is the HTMX way. Mitigate with optimistic UI updates in mcms-tree-nav (already has this for reorders).
