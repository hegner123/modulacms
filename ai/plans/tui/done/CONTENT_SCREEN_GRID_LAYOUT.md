# Content Screen Grid Layout Plan

## Overview

Redesign the content screen to use the new 12-column grid layout system, replacing the legacy 3-panel layout. The screen has two phases -- content select and tree browsing -- each with a 2-column grid. Field values in the preview are read-only; editing is handled via dialogs.

## Resolved Decisions

- **`_global` type**: Already registered as `DatatypeTypeGlobal` in `internal/db/types/reserved.go`. Standalone content uses `_global` (and potentially other non-routed root types). Partitioning: `RouteID.Valid` for routed, `!RouteID.Valid` for standalone.
- **Field editing**: No inline field editing in tree phase. Field values are immutable in the content display. Users edit content values via the existing form dialog system.
- **Expand vs select**: Separate keybindings. Expand/collapse toggles child visibility; select (enter) navigates into the content tree for that item.
- **Version list**: Renders in a footer row within the tree browsing grid (see Grid Definitions).
- **Batch field queries**: `ListContentFieldsByRoute` and `ListAdminContentFieldsByRoute` already exist on DbDriver. No new SQL queries needed for batch loading.
- **ScreenMode**: `ScreenFull` is handled by `GridScreen.RenderGrid` (renders only focused cell). `ScreenWide` accordion behavior is dropped -- the 2-column grid is already compact enough.
- **Standalone content selection**: Pressing enter on a standalone item (no route) loads the content tree using `GetContentDataDescendants(contentDataID)` directly. A new `LoadStandaloneTreeCmd(contentDataID)` command wraps this. The tree phase works identically for routed and standalone content once the tree is loaded.
- **Admin mode toggle in select phase**: Toggling admin mode (ctrl+a) triggers a full rebuild of `SelectTree` and `FlatSelectList` from the appropriate data source (`AdminRootContentSummary` when admin, `RootContentSummary` when client). Cursor resets to 0.

## Content Select Phase

Two-column layout. Left column is a slug-grouped hierarchical content list. Right column stacks detail view and stats.

```
+--------------+-------------------------------+
| Content      | Details                       |
| (3/12)       | (9/12, 0.65)                  |
|              |                               |
| -- Pages --  | slug, title, datatype, status |
| /            | author, dates, route info     |
| v /about     |                               |
|   /about/v.. |                               |
| /contact     +-------------------------------+
|              | Stats                         |
| -- Global -- | (9/12, 0.35)                  |
| Main Menu    | counts, key hints             |
| Footer       |                               |
+--------------+-------------------------------+
```

### Slug Tree Model

A TUI-side data structure built from the flat `[]ContentDataTopLevel` list on fetch result receipt (`RootContentSummaryFetchResultsMsg` / `AdminContentDataFetchResultsMsg`). Rebuilt whenever the underlying list changes or when admin mode toggles. Retained (not nil'd) when entering tree phase so that returning via back does not require a rebuild.

```go
// ContentSelectNodeKind discriminates the three node types in the select tree.
type ContentSelectNodeKind int

const (
    NodeContent ContentSelectNodeKind = iota // selectable content item
    NodeSection                              // non-selectable section header ("Pages", "Standalone")
    NodeGroup                                // collapsible group (slug prefix or datatype label)
)

type ContentSelectNode struct {
    Kind     ContentSelectNodeKind
    Label    string                  // display label: "/" or "about" or "Main Menu"
    Slug     string                  // full slug for sorting (empty for standalone)
    Depth    int                     // visual indent level
    Expand   bool                    // collapsible group state (only meaningful for NodeGroup)
    Content  *db.ContentDataTopLevel // non-nil only for NodeContent
    Children []*ContentSelectNode    // nested items
}
```

For admin mode, `ContentSelectNode.Content` is nil and a separate `AdminContent *db.AdminContentDataTopLevel` field is used:

```go
type ContentSelectNode struct {
    Kind         ContentSelectNodeKind
    Label        string
    Slug         string
    Depth        int
    Expand       bool
    Content      *db.ContentDataTopLevel      // non-nil for regular mode NodeContent
    AdminContent *db.AdminContentDataTopLevel  // non-nil for admin mode NodeContent
    Children     []*ContentSelectNode
}
```

Invariant: exactly one of `Content` / `AdminContent` is non-nil for `NodeContent` nodes, based on `ContentScreen.AdminMode`. Both are nil for `NodeSection` and `NodeGroup`.

### Flat Cursor Indexing

`FlattenSelectTree(roots []*ContentSelectNode) []*ContentSelectNode` produces a flat slice for cursor navigation by walking the tree depth-first, respecting `Expand` state (collapsed children are omitted). **Only `NodeContent` and `NodeGroup` items are included; `NodeSection` items are excluded.** This means the cursor index always maps to a selectable/expandable item -- no skip logic is needed in the up/down navigation handlers.

When the user presses enter on a `NodeContent` item, the `Content` (or `AdminContent`) pointer provides the data needed to enter tree phase. When the user presses enter on a `NodeGroup` item, nothing happens (expand/collapse uses a separate keybinding).

Up/down increments/decrements the cursor index. Boundary clamping at 0 and `len(FlatSelectList)-1`. No wrap.

### Slug Sorting Algorithm

1. **Partition** items into routed (`RouteID.Valid == true`) and standalone (`RouteID.Valid == false`).
2. **Routed items** -- multi-phase lexicographic sort by slug segments:
   - Split each slug by `/` into segments (filter empty strings from leading `/`).
   - Special case: `/` (single slash, homepage) sorts first (segment count = 0).
   - Compare segment[0] first, then segment[1] if tied, etc.
   - Items sharing a first segment are grouped under a collapsible parent.
   - Example ordering: `/` < `/about` < `/about/team` < `/about/visit` < `/blog` < `/contact`
3. **Standalone items** -- group by `DatatypeLabel` alphabetically, then sort by content title/ID within each group. Standalone section appears below all routed items.

### Section Rendering

```
-- Pages ----------------------------
  / Homepage              [Page] *
  v /about                [Page] *
    /about/visit          [Page] o
    /about/team           [Page] o
  /contact                [Page] *
  /blog                   [Blog] *

-- Standalone -----------------------
  v Menu
    Main Menu             [Menu] *
    Footer                [Menu] *
  v Config
    Site Settings         [Config] *
```

- `*` = published, `o` = draft
- Section headers ("Pages", "Standalone") render visually but are excluded from the flat cursor list.
- Group nodes with children are collapsible (v/> toggle via expand keybinding).
- `/about` can be both a content item AND a collapsible parent. Expand keybinding toggles children; enter navigates into that content's tree.

### Keybindings (Select Phase)

| Key | Action |
|-----|--------|
| up/down (j/k) | Move cursor |
| enter | Select NodeContent item, enter tree phase |
| space or right | Expand/collapse NodeGroup (also works on NodeContent with children) |
| n | New content |
| d | Delete content |
| ctrl+a | Toggle admin/client mode (rebuilds select tree) |
| esc | Back to home |
| q | Quit |

### Scroll Support

The left column (content list) uses `CellContent.TotalLines` and `CellContent.ScrollOffset` to enable scrolling within the Panel. `TotalLines` = length of flat visible list. `ScrollOffset` = `ClampScroll(cursor, totalLines, innerHeight)`. Same pattern used by the old 3-panel layout.

## Tree Browsing Phase

Two-column layout. Left is the content tree. Right is the full document preview showing all nodes' field values (read-only).

```
+--------------+-------------------------------+
| Tree         | Preview                       |
| (3/12)       | (9/12)                        |
|              |                               |
| v Row        | Row                           |
|   v Column   |   Column                      |
|     Hero  <- |     Hero Content ============ |
|   v Column   |       title: Welcome          |
|     Cards    |       subtitle: Building...   |
|              |       image: hero.jpg         |
|              |   Column                      |
|              |     Featured Cards ---------- |
|              |       title: What We Offer    |
|              |       card_1: Speed           |
|              |       card_2: Quality         |
+--------------+-------------------------------+
```

### Preview Rendering Rules

- Every node in the tree renders in document order (depth-first traversal).
- Section title indentation reflects tree depth.
- The selected node's section gets a highlighted border/accent box.
- Non-selected sections are dimmed.
- Structural nodes (Row, Column) render the same way: title at their indent depth, fields below if they have any. Since structural nodes have few/no fields, they add minimal visual noise.
- Field values are read-only. No field cursor, no field selection, no inline editing.

### Preview Scroll

The preview pane uses `CellContent.TotalLines` and `CellContent.ScrollOffset`. When the tree cursor moves, calculate the line offset of the selected node's section in the rendered preview string and set `ScrollOffset` to bring it into view. This is a line-count calculation during rendering, not a separate viewport widget.

### Version List

When the user triggers version view (keybinding), the tree browsing grid swaps to a 3-cell variant with a footer row for versions:

```
+--------------+-------------------------------+
| Tree         | Preview                       |
| (3/12)       | (9/12, 0.65)                  |
|              |                               |
+--------------+-------------------------------+
| Versions                                     |
| (12/12, 0.35)                                |
+----------------------------------------------+
```

The version footer is full-width. Since the grid system renders columns side-by-side, the footer is rendered outside the grid as a separate panel joined vertically in `View()`.

Height calculation when `ShowVersionList == true`:
```go
gridHeight := int(float64(ctx.Height) * 0.65)
footerHeight := ctx.Height - gridHeight
```
Pass `gridHeight` as height to `RenderGrid`. Render the version panel as `Panel{Width: ctx.Width, Height: footerHeight, ...}` and join vertically with `lipgloss.JoinVertical`.

Selecting a version triggers the existing restore dialog. Pressing back/esc returns to the normal 2-cell tree grid.

### Keybindings (Tree Phase)

| Key | Action |
|-----|--------|
| up/down (j/k) | Move tree cursor |
| enter/space | Expand/collapse node |
| e | Edit content (opens form dialog for selected node) |
| n | New child content (dialog) |
| d | Delete content (dialog) |
| c | Copy content |
| m | Move content (dialog) |
| +/- | Reorder siblings |
| p | Publish/unpublish (dialog) |
| v | Show version list |
| L | Switch locale |
| ctrl+a | Toggle admin/client mode |
| esc | Back to select phase |
| q | Quit |

## Data Changes

### 1. Add DatatypeType to Top-Level Queries

Add `DatatypeType string` field to both `ContentDataTopLevel` and `AdminContentDataTopLevel`.

Update the SQL queries in all 3 backends (`just sqlc` is mandatory since query files change):
- `ListContentDataTopLevelPaginated` (sqlite, mysql, psql)
- `ListAdminContentDataTopLevelPaginated` (sqlite, mysql, psql)

The JOIN already reaches the datatypes table for `DatatypeLabel`. Adding `datatypes.type AS datatype_type` to the SELECT is a one-line change per query.

Update ALL map functions that return `ContentDataTopLevel` or `AdminContentDataTopLevel`. This includes:
- `mapContentDataTopLevel` (3 backends)
- `mapContentDataTopLevelByStatus` (3 backends)
- `mapAdminContentDataTopLevel` (3 backends)
- Any other functions returning these types -- search for all such functions and update them all.

### 2. Batch Field Loading

Use existing queries to load ALL fields for a tree in one pass after tree load:
- `ListContentFieldsByRoute(routeID)` -- already on DbDriver, all 3 backends
- `ListAdminContentFieldsByRoute(adminRouteID)` -- already on DbDriver, all 3 backends
- Locale-aware variants: `ListContentFieldsByRouteAndLocale`, `ListAdminContentFieldsByRouteAndLocale`

**For standalone content** (no route), use `ListContentFieldsByContentDataIDs` if available, or iterate the tree nodes and batch by their content IDs.

The mapping from raw `db.ContentFields` to `ContentFieldDisplay` currently happens in `commands.go` (search for `ContentFieldDisplay{`) and `admin_constructors.go` (search for `AdminContentFieldDisplay{`). Extract only the canonical field mapping logic (field definitions + content field values). The lightweight join-based path in `admin_constructors.go` remains as-is for other use cases.

```go
// MapContentFieldsToDisplay maps raw DB content fields + field definitions
// to display structs, keyed by ContentDataID.
func MapContentFieldsToDisplay(
    contentFields []db.ContentFields,
    fieldDefs []db.Fields,
) map[types.ContentID][]ContentFieldDisplay
```

The function joins content fields with field definitions (for Label, Type, Validation, Data) and groups by `ContentDataID`.

Field definitions come from the datatypes in the tree. A single tree can contain nodes of multiple datatypes. The batch loader calls `ListFieldsByDatatypeID` once per distinct datatype ID found in the tree's nodes, then merges the results into a single `[]db.Fields` slice passed to the mapping function.

Admin variant (explicit types):
```go
func MapAdminContentFieldsToDisplay(
    contentFields []db.AdminContentFields,
    fieldDefs []db.AdminFields,
) map[types.AdminContentID][]AdminContentFieldDisplay
```

**Commands and messages:**

Create `BatchLoadContentFieldsCmd(routeID types.RouteID, locale string)` tea.Cmd:
1. Calls `ListContentFieldsByRoute` (or `ListContentFieldsByRouteAndLocale` if locale non-empty)
2. Collects distinct datatype IDs from the tree, calls `ListFieldsByDatatypeID` per datatype
3. Maps via `MapContentFieldsToDisplay`
4. Returns `BatchContentFieldsLoadedMsg{Fields: map[types.ContentID][]ContentFieldDisplay}`
5. On error, returns `FetchErrMsg{Error: err}` (existing error handler in `model.go` displays it)

Create `BatchLoadAdminContentFieldsCmd(adminRouteID types.AdminRouteID, locale string)` tea.Cmd:
1. Calls `ListAdminContentFieldsByRoute` (or locale variant)
2. Collects distinct admin datatype IDs, calls field defs per datatype
3. Maps via `MapAdminContentFieldsToDisplay`
4. Returns `BatchAdminContentFieldsLoadedMsg{Fields: map[types.AdminContentID][]AdminContentFieldDisplay}`
5. On error, returns `FetchErrMsg{Error: err}`

Define message types in `commands.go`:
```go
type BatchContentFieldsLoadedMsg struct {
    Fields map[types.ContentID][]ContentFieldDisplay
}

type BatchAdminContentFieldsLoadedMsg struct {
    Fields map[types.AdminContentID][]AdminContentFieldDisplay
}
```

Wire into tree load: after `TreeLoadedMsg`, dispatch `BatchLoadContentFieldsCmd`. After `AdminTreeLoadedMsg`, dispatch `BatchLoadAdminContentFieldsCmd`. The ContentScreen `Update` handler stores the map on `s.AllFields` / `s.AllAdminFields`.

### 3. Standalone Content Identification

Standalone content = items where `RouteID.Valid == false`. The `DatatypeType` field (from Step 1) is used for display grouping in the standalone section (group by datatype label), but the partition logic itself only needs `RouteID.Valid`.

`_global` items will naturally fall into standalone because they have no route. Other unrouted `_root` items (if any exist) also appear in standalone.

## Grid Definitions

```go
// Content select phase: 2 columns, stacked right
var contentSelectGrid = Grid{
    Columns: []GridColumn{
        {Span: 3, Cells: []GridCell{
            {Height: 1.0, Title: "Content"},
        }},
        {Span: 9, Cells: []GridCell{
            {Height: 0.65, Title: "Details"},
            {Height: 0.35, Title: "Stats"},
        }},
    },
}

// Tree browsing phase: 2 columns, single right
var contentTreeGrid = Grid{
    Columns: []GridColumn{
        {Span: 3, Cells: []GridCell{
            {Height: 1.0, Title: "Tree"},
        }},
        {Span: 9, Cells: []GridCell{
            {Height: 1.0, Title: "Preview"},
        }},
    },
}
```

The version footer grid is not a separate `Grid` definition. Instead, when `ShowVersionList` is true, `View()` uses `contentTreeGrid` but overrides the height passed to `RenderGrid` and appends the version footer panel via `lipgloss.JoinVertical`.

## Implementation Steps

### Step 1: SQL + Struct Changes
- Add `DatatypeType string` to `ContentDataTopLevel` and `AdminContentDataTopLevel`
- Update SQL query files (3 backends x 2 tables) to SELECT `datatypes.type`
- Run `just sqlc` to regenerate
- Update ALL map functions returning these types (search codebase -- includes `mapContentDataTopLevel`, `mapContentDataTopLevelByStatus`, `mapAdminContentDataTopLevel`, and any variants)

### Step 2: Extract Field Display Mapping + Batch Commands
- Extract `MapContentFieldsToDisplay` into `internal/tui/content_field_map.go`
- Extract `MapAdminContentFieldsToDisplay` into the same file
- Source: canonical mapping logic from `commands.go` (search `ContentFieldDisplay{`) and `admin_constructors.go` (search `AdminContentFieldDisplay{`)
- Create `BatchLoadContentFieldsCmd` and `BatchLoadAdminContentFieldsCmd` in `commands.go`
- Define `BatchContentFieldsLoadedMsg` and `BatchAdminContentFieldsLoadedMsg` in `commands.go`
- Error path: return `FetchErrMsg{Error: err}`
- Wire: dispatch batch load after `TreeLoadedMsg` / `AdminTreeLoadedMsg`

### Step 3: ContentSelectNode Slug Tree Builder
- New file: `internal/tui/content_select_tree.go`
- Define `ContentSelectNodeKind` enum and `ContentSelectNode` struct (with both `Content` and `AdminContent` fields)
- `BuildContentSelectTree(items []db.ContentDataTopLevel) []*ContentSelectNode`
- `BuildAdminContentSelectTree(items []db.AdminContentDataTopLevel) []*ContentSelectNode`
- Slug parsing, multi-phase segment sorting, section partitioning
- `FlattenSelectTree(nodes []*ContentSelectNode) []*ContentSelectNode`
  - Walks depth-first, respects `Expand` state
  - Excludes `NodeSection` items -- cursor always maps to selectable/expandable item
- Unit tests: `internal/tui/content_select_tree_test.go`
  - Sorting edge cases: duplicate prefixes, single `/`, empty slugs, mixed routed/standalone
  - Flatten with collapsed groups
  - Section header exclusion
  - Admin content node population

### Step 4: Grid Layouts + GridScreen Migration
- Define `contentSelectGrid` and `contentTreeGrid` in `screen_content.go`
- Migrate `ContentScreen` to embed `GridScreen` instead of using raw `PanelFocus`
- Remove `PanelFocus` field (replaced by `GridScreen.FocusIndex`)
- Remove `handleFieldPanelKeys` entirely -- there is no field panel in the new design. All field editing is initiated from the tree panel via the 'e' keybinding opening the dialog system.
- Swap grid based on phase: `s.Grid = contentSelectGrid` when entering select, `s.Grid = contentTreeGrid` when entering tree
- Add fields to ContentScreen:
  - `SelectTree []*ContentSelectNode` -- built on fetch, retained across phases
  - `FlatSelectList []*ContentSelectNode` -- flattened for cursor, rebuilt on expand/collapse and on return from tree phase
  - `AllFields map[types.ContentID][]ContentFieldDisplay` -- batch loaded after tree load
  - `AllAdminFields map[types.AdminContentID][]AdminContentFieldDisplay`

### Step 5: Rewrite Select Phase Rendering
- Replace `renderRouteList()` with slug tree rendering using `FlatSelectList`
  - Render section headers inline (they appear visually but are not in `FlatSelectList`)
  - Render `NodeGroup` items with expand/collapse indicator
  - Render `NodeContent` items with cursor, datatype badge, and publish status
- Replace `renderRouteDetail()` with detail view from `ContentSelectNode.Content` (or `.AdminContent`)
- Replace `renderRouteActions()` with stats cell (content count, route count, datatype breakdown)
- Wire expand/collapse keybinding on group nodes (rebuilds `FlatSelectList`)
- Enter on `NodeContent`:
  - If `Content.RouteID.Valid`: use existing route select logic (`ReloadContentTreeCmd`)
  - If `!Content.RouteID.Valid`: use new `LoadStandaloneTreeCmd(Content.ContentDataID)` which calls `GetContentDataDescendants` and builds the tree
  - Same pattern for admin via `AdminContent`
- Scroll support via `CellContent.TotalLines` / `CellContent.ScrollOffset`

### Step 6: Rewrite Tree Phase Preview
- Replace `renderContentPreview()` with full document preview
- Traverse entire tree depth-first, render each node as a section:
  - Title line at tree depth indent (e.g., `"  Column"`, `"    Hero Content"`)
  - Field values below title, indented one level deeper
  - Selected node's section: accent-colored border/box
  - Other sections: dimmed/faint style
- Use `AllFields` / `AllAdminFields` map to look up fields per node (no per-cursor DB calls)
- Calculate line offset of selected section for scroll auto-positioning
- Remove `renderFields()`, `renderRegularFields()`, `renderAdminFields()`, `renderTreeActions()`

### Step 7: Version List in Footer
- When version keybinding is pressed, set `ShowVersionList = true`
- `View()` when `ShowVersionList`:
  - `gridHeight := int(float64(ctx.Height) * 0.65)`
  - `footerHeight := ctx.Height - gridHeight`
  - Render grid with `gridHeight`, render version panel as `Panel{Width: ctx.Width, Height: footerHeight}`
  - Join with `lipgloss.JoinVertical`
- Version panel reuses existing `renderVersionList()` / `renderAdminVersionList()`
- Esc/back in version mode: set `ShowVersionList = false`, restore normal grid
- Version cursor navigation works on the version list, not the tree

### Step 8: Remove Old Layout Code
- Remove `CONTENT` and `ADMINCONTENT` entries from `pageLayouts` map in `pages.go`
- Remove `PanelFocus`-based panel cycling from ContentScreen
- Remove `handleFieldPanelKeys` entirely (no field panel in new design)
- Remove per-cursor field loading (`loadFieldsForCurrentNode`, `loadAdminFieldsForCurrentNode`)
- Remove `renderTreeActions()`
- Remove old `View()` method with manual panel width calculations
- `layoutForPage` is no longer called for these pages; `PageIndex()` still returns the constants for navigation/history

### Step 9: Update Key Hints and Navigation
- Rewrite `ContentScreen.KeyHints` method to use phase-based conditions:
  - If `ShowVersionList`: version hints (up/down, select/restore, back)
  - If `inTreePhase()`: tree hints (up/down, expand, edit, new, delete, copy, move, publish, versions, locale, back, quit)
  - Else (select phase): select hints (up/down, expand, select, new, delete, admin toggle, back, quit)
- No `PanelFocus` branches -- the `FocusIndex` from GridScreen determines which cell has keyboard focus, but key hints are phase-based, not cell-based
- Admin toggle (ctrl+a) works in all phases; in select phase it triggers select tree rebuild

## File Impact Summary

| File | Change |
|------|--------|
| `sql/schema/16_content_data/queries*.sql` | Add `datatypes.type` to top-level content queries |
| `sql/schema/17_admin_content_data/queries*.sql` | Add `datatypes.type` to admin top-level queries |
| `internal/db/content_data_custom.go` | Add `DatatypeType` field, update ALL map functions |
| `internal/db/admin_content_data_custom.go` | Add `DatatypeType` field, update ALL map functions |
| `internal/tui/content_select_tree.go` | NEW: ContentSelectNodeKind, ContentSelectNode, slug tree builder, sort, flatten |
| `internal/tui/content_select_tree_test.go` | NEW: slug sort + tree builder tests |
| `internal/tui/content_field_map.go` | NEW: MapContentFieldsToDisplay, MapAdminContentFieldsToDisplay |
| `internal/tui/screen_content.go` | Embed GridScreen, swap grids, batch field state, standalone tree load, admin toggle rebuild |
| `internal/tui/screen_content_view.go` | Rewrite View(), select rendering, preview rendering, version footer |
| `internal/tui/pages.go` | Remove CONTENT/ADMINCONTENT from pageLayouts |
| `internal/tui/commands.go` | Add BatchLoad*Cmd, BatchLoad*Msg types, LoadStandaloneTreeCmd |
| `internal/tui/admin_constructors.go` | Replace inline admin field mapping with extracted function |

## Dependencies

- Grid system (`grid.go`, `grid_screen.go`) -- already implemented
- Panel rendering (`panel.go`) -- already supports accent override and scroll indicators
- Dialog system -- already handles content field editing
- AppContext with ActiveAccent -- already implemented
- `_global` reserved type -- already registered
- `ListContentFieldsByRoute` / `ListAdminContentFieldsByRoute` -- already on DbDriver
- `GetContentDataDescendants` -- already on DbDriver (used for standalone tree loading)
