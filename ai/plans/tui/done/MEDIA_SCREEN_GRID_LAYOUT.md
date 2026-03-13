# Media Screen Grid Layout

Migrate the media screen from legacy 2-panel layout to the 12-column grid system.
Add a URL-path-derived folder tree and inline search/filter.

## Layout

```
┌──────────┬─────────────────────────┐
│  Media   │       Summary           │
│  Tree    │  name, mime, dims, url  │
│          │                         │
│  / ___   ├─────────────────────────┤
│  (3)     │       Metadata          │
│          │  alt, caption, desc,    │
│          │  focal, author, dates   │
└──────────┴─────────────────────────┘
```

### Grid definition

```go
var mediaGrid = Grid{
    Columns: []GridColumn{
        {Span: 3, Cells: []GridCell{
            {Height: 1.0, Title: "Media"},
        }},
        {Span: 9, Cells: []GridCell{
            {Height: 0.40, Title: "Summary"},
            {Height: 0.60, Title: "Metadata"},
        }},
    },
}
```

- Column 1 (span 3): scrollable folder tree with inline search input
- Column 2 (span 9): two stacked cells
  - Summary (40%): identity fields — display name, filename, mimetype, dimensions, URL, srcset
  - Metadata (60%): editorial fields — alt, caption, description, focal point, class, author, created, modified

## Steps

### Step 1: Create media tree node types

File: `internal/tui/media_tree.go`

Define the tree node structure and builder that converts a flat `[]db.Media` into a
URL-path-grouped folder tree using sibling pointers.

```go
type MediaNodeKind int

const (
    MediaNodeFile   MediaNodeKind = iota  // actual media item (leaf)
    MediaNodeFolder                       // virtual folder from shared path prefix
)

type MediaTreeNode struct {
    Kind        MediaNodeKind
    Label       string        // folder segment name or filename
    Path        string        // full path for sorting and grouping
    Depth       int
    Expand      bool
    Media       *db.Media     // non-nil only for MediaNodeFile
    FirstChild  *MediaTreeNode
    NextSibling *MediaTreeNode
    PrevSibling *MediaTreeNode
}
```

Builder function:

```go
func BuildMediaTree(items []db.Media) []*MediaTreeNode
```

Algorithm:
1. Extract path from each `media.URL` (strip scheme+host, or use the path portion)
2. Sort items by path segments lexicographically
3. Split each path into segments
4. Group by shared prefixes, creating `MediaNodeFolder` nodes for intermediate segments
5. Leaf nodes are `MediaNodeFile` with `Media` pointer set
6. Link children via `appendMediaChildNode` (same pattern as `appendChildNode`)
7. All folders default to `Expand: true`

Helper functions:

```go
func appendMediaChildNode(parent, child *MediaTreeNode)
func (n *MediaTreeNode) hasChildren() bool
func FlattenMediaTree(roots []*MediaTreeNode) []*MediaTreeNode
```

`FlattenMediaTree` walks depth-first respecting `Expand` state, returns only
navigable nodes (both folders and files — folders for expand/collapse, files for selection).

### Step 2: Add search/filter support

Add filtering to `media_tree.go`:

```go
func FilterMediaList(items []db.Media, query string) []db.Media
```

Case-insensitive `strings.Contains` match against:
- `media.Name`
- `media.DisplayName`
- `media.Mimetype`
- `media.URL` (path portion)

Returns the subset of items that match. The tree is rebuilt from filtered items,
so folder structure reflects only matching results.

### Step 3: Rewrite MediaScreen struct

File: `internal/tui/screen_media.go`

Replace the current struct with grid-based layout and search state.

```go
type MediaScreen struct {
    GridScreen
    Cursor       int
    MediaList    []db.Media          // full list from DB
    FilteredList []db.Media          // displayed subset (== MediaList when no filter)
    MediaTree    []*MediaTreeNode    // tree built from FilteredList
    FlatList     []*MediaTreeNode    // flattened for cursor navigation
    Searching    bool                // true when search input is active
    SearchInput  textinput.Model
    SearchQuery  string              // persisted query (set on enter)
}
```

Constructor:

```go
func NewMediaScreen(mediaList []db.Media) *MediaScreen
```

- Initialize `GridScreen` with `mediaGrid` and `FocusIndex: 0`
- Set `FilteredList = MediaList`
- Build tree and flatten
- Create `textinput.New()` with placeholder "filter..."

### Step 4: Implement Update logic

Rewrite `Update` in `screen_media.go`.

**Search mode** (`s.Searching == true`):
- All key input goes to `textinput.Model` except:
  - **Enter**: exit search, persist query in `SearchQuery`, keep filter active
  - **Esc**: exit search, clear query, reset `FilteredList = MediaList`, rebuild tree

On each keystroke in search mode:
1. Get `s.SearchInput.Value()`
2. `FilteredList = FilterMediaList(MediaList, value)`
3. Rebuild tree: `MediaTree = BuildMediaTree(FilteredList)`
4. Flatten: `FlatList = FlattenMediaTree(MediaTree)`
5. Reset cursor to 0

**Browse mode** (`s.Searching == false`):
- `/` — enter search mode, focus textinput
- `enter` on folder node — toggle `Expand`, re-flatten
- `enter` on file node — no-op (details shown automatically)
- `n` (ActionNew) — open file picker for upload
- `d` (ActionDelete) — delete selected media (only on file nodes)
- `up/down` — cursor navigation within `FlatList`
- `tab/shift-tab` — cell focus cycling via `GridScreen.HandleFocusNav`
- `h` (ActionBack) — history pop

**Message handlers** (same as current):
- `MediaFetchMsg` — fetch from DB
- `MediaFetchResultsMsg` — set `MediaList`, rebuild tree
- `MediaListSet` — refresh after CMS operation

### Step 5: Implement View rendering

File: `internal/tui/screen_media_view.go`

Split view methods into a separate file (matches `screen_content_view.go` pattern).

**`View(ctx AppContext) string`**:
- Build 3 `CellContent` entries (tree, summary, metadata)
- Call `s.RenderGrid(ctx, cells)`

**`renderMediaTree(ctx AppContext) CellContent`**:
- Walk `FlatList`, render each node with indent based on `Depth`
- Folder nodes: prefix with `▸`/`▾` for collapsed/expanded
- File nodes: prefix with cursor indicator (`->` or spaces)
- If `s.Searching`, append a search input line at the bottom: `/ ` + `s.SearchInput.View()`
- If `s.SearchQuery != ""` and not searching, show filter badge in content
- Set `TotalLines` and `ScrollOffset` via `ClampScroll`
- Panel title: `"Media"` or `"Media (N/M)"` when filtered

**`renderMediaSummary() string`**:
- If no file node selected, return "No media selected"
- Selected = `FlatList[Cursor]` when it's a file node; otherwise find nearest file
- Display: display name, filename, mimetype, dimensions, URL, srcset

**`renderMediaMetadata() string`**:
- Same selected item as summary
- Display: alt, caption, description, focal X/Y, class, author ID, created, modified

### Step 6: Update pages.go and remove legacy layout

In `internal/tui/pages.go`:
- Remove or comment out `MEDIA` from `pageLayouts` map (grid replaces it)

### Step 7: Update key hints

```go
func (s *MediaScreen) KeyHints(km config.KeyMap) []KeyHint {
    if s.Searching {
        return []KeyHint{
            {"type", "filter"},
            {"enter", "accept"},
            {"esc", "clear"},
        }
    }
    return []KeyHint{
        {km.HintString(config.ActionSearch), "search"},
        {km.HintString(config.ActionNew), "upload"},
        {km.HintString(config.ActionDelete), "del"},
        {"enter", "expand"},
        {km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
        {km.HintString(config.ActionNextPanel), "panel"},
        {km.HintString(config.ActionBack), "back"},
        {km.HintString(config.ActionQuit), "quit"},
    }
}
```

Note: if `config.ActionSearch` doesn't exist yet, add `ActionSearch` to the keymap
with default binding `/`.

### Step 8: Write tests

File: `internal/tui/media_tree_test.go`

Test cases:
- **Empty list**: `BuildMediaTree(nil)` returns nil
- **Single file**: one node, no folders
- **Shared prefix**: `/images/a.jpg` and `/images/b.jpg` grouped under `images/` folder
- **Nested folders**: `/a/b/c/file.jpg` creates `a/ > b/ > c/ > file.jpg`
- **Mixed depths**: some files at root, some nested
- **Flatten respects expand**: collapsed folder hides children
- **FilterMediaList**: matches on name, display name, mimetype, URL path; case-insensitive

File: `internal/tui/screen_media_test.go`

Test cases:
- **NewMediaScreen**: initial state has tree built, cursor at 0, not searching
- **Search mode toggle**: `/` enters, esc clears, enter persists
- **Filter rebuilds tree**: verify tree structure changes when filter applied
- **Cursor bounds**: cursor stays in range after filter reduces list
- **Expand/collapse**: enter on folder toggles expand, re-flattens

## Dependencies

- `internal/tui/grid.go` — Grid, GridColumn, GridCell, CellContent (exists)
- `internal/tui/grid_screen.go` — GridScreen, HandleFocusNav, RenderGrid (exists)
- `internal/tui/panel.go` — Panel, ClampScroll, PanelInnerHeight (exists)
- `github.com/charmbracelet/bubbles/textinput` — search input (already vendored)

## Files modified

| File | Change |
|------|--------|
| `internal/tui/media_tree.go` | **NEW** — tree node types, builder, flatten, filter |
| `internal/tui/screen_media.go` | Rewrite struct, Update, constructor |
| `internal/tui/screen_media_view.go` | **NEW** — View, renderMediaTree, renderMediaSummary, renderMediaMetadata |
| `internal/tui/pages.go` | Remove MEDIA from pageLayouts |
| `internal/tui/media_tree_test.go` | **NEW** — tree builder + filter tests |
| `internal/tui/screen_media_test.go` | **NEW** — screen behavior tests |
| `internal/tui/keys.go` (if needed) | Add ActionSearch keybinding |
