# Datatypes Screen Grid Layout

Rewrite the datatypes screen as a 3-phase GridScreen with hierarchical tree browsing,
field selection, and per-property field editing. Replaces the legacy 3-panel layout.
Handles both regular and admin mode via `AdminMode` flag (same struct, branching on mode).

Split into two deliverables:
- **Deliverable A** (Steps 1-6, 8-12): GridScreen migration with Phase 1 + Phase 2,
  keeping existing edit-all-at-once field dialog. Field reordering included.
- **Deliverable B** (Step 7): Phase 3 per-property field editing with new dialog types.

Deliverable A is a proven migration pattern. Deliverable B is a new feature that deserves
its own design pass after A is stable.

## Phases

### Phase 1: Datatype Browse

```
+---------------+-------------------------------+
|               | Details                       |
|  Datatypes    | type, parent, ID, field count |
|  (span 4)     | (0.30)                        |
|               +-------------------------------+
|  tree view    | Field Preview                 |
|  collapsible  | numbered list, type badges    |
|  searchable   | (0.70)                        |
|               |                               |
+---------------+-------------------------------+
```

- Left (span 4): hierarchical datatype tree built from `parent_id`. Collapsible nodes.
  Inline search via `/` (ActionSearch) filters by label. Wider span accommodates deep nesting.
- Center top (0.30): selected datatype metadata -- label, type, parent, ID, field count, dates.
- Center bottom (0.70): read-only field preview for the selected datatype -- label + type per field.

Key actions:
- Up/Down: navigate tree
- Enter/Right: select datatype, transition to Phase 2
- Expand/Collapse: toggle tree node expand state
- `/`: toggle inline search
- `n`: new datatype
- `e`: edit datatype (dialog)
- `d`: delete datatype (confirm dialog)
- Back/Esc: pop history
- Tab: cycle focus between cells

### Phase 2: Field Selection

```
+-------------+---------------------------------+
|             | Field Properties                |
|  Fields     | Label:    Hero Section          |
|  (span 3)   | Type:     group                 |
|             | Required: false                 |
|  cursor     | Sort:     1                     |
|  list       | (0.65)                          |
|             +---------------------------------+
|             | Context                         |
|             | Parent > Datatype breadcrumb    |
|             | (0.35)                          |
+-------------+---------------------------------+
```

- Left (span 3): field list for the selected datatype, ordered by sort_order.
  Reorder up/down supported (ActionReorderUp/ActionReorderDown).
- Center top (0.65): read-only property preview for the cursor-highlighted field.
- Center bottom (0.35): breadcrumb showing datatype parent chain, datatype label,
  total field count, and context-sensitive key hints.

Key actions:
- Up/Down: navigate field list
- Shift+Up/Shift+Down: reorder field (swap sort_order)
- Enter: enter field, transition to Phase 3 (focus center)
- `n`: new field (dialog)
- `e`: edit field (form dialog, all properties at once)
- `d`: delete field (confirm dialog)
- Back/Esc: return to Phase 1 (swap grid back)
- Tab: cycle focus between cells

### Phase 3: Field Edit (Deliverable B)

```
+-------------+---------------------------------+
|             | Field Properties (focused)      |
|  Fields     |   Label:       [Hero Section]   |
|  (span 3)   |   Type:        group            |
|             |   Sort Order:  1                |
|  (dimmed)   |   Data:        {}               |
|             |   Validation:  {}               |
|             |   UI Config:   {...}            |
|             |   Translatable: false           |
|             |   Required:    false            |
|             +---------------------------------+
|             | Context                         |
|             |                                 |
+-------------+---------------------------------+
```

Same grid as Phase 2. Focus shifts to the properties cell. Each property is a
cursor-navigable row. Enter on a row opens the appropriate edit dialog:

- Label, Name, Data, Validation, Roles: text input dialog (new: `ShowEditFieldPropertyDialogCmd`)
- Type: select dialog with FieldType enum (new: `ShowFieldTypeSelectDialogCmd`)
- Sort Order: number input dialog (new, or variant of text input)
- Translatable: toggle (immediate, no dialog)
- UI Config: opens existing UIConfig form dialog (`ShowUIConfigFormDialogCmd` / `ShowEditUIConfigFormDialogCmd`)

These dialog commands are **new** and do not exist yet. They are scoped to Deliverable B.
Each returns a result message that dispatches a single-property update command.

Back/Esc returns focus to Phase 2 (field list).

## Grid Definitions

```go
var datatypeBrowseGrid = Grid{
    Columns: []GridColumn{
        {Span: 4, Cells: []GridCell{
            {Height: 1.0, Title: "Datatypes"},
        }},
        {Span: 8, Cells: []GridCell{
            {Height: 0.30, Title: "Details"},
            {Height: 0.70, Title: "Fields"},
        }},
    },
}

var datatypeFieldGrid = Grid{
    Columns: []GridColumn{
        {Span: 3, Cells: []GridCell{
            {Height: 1.0, Title: "Fields"},
        }},
        {Span: 9, Cells: []GridCell{
            {Height: 0.65, Title: "Properties"},
            {Height: 0.35, Title: "Context"},
        }},
    },
}
```

## Data Structures

### DatatypeTreeNode

File: `internal/tui/datatype_tree.go`

Hierarchical tree node built from `parent_id` relationships. Uses a `Children` slice
(not sibling pointers). Datatypes have simple parent-child grouping sorted by label,
which doesn't need the insertion-ordered sibling pointer pattern that `MediaTreeNode`
uses for URL-path-derived folder structures.

```go
type DatatypeNodeKind int

const (
    DatatypeNodeItem DatatypeNodeKind = iota
    DatatypeNodeGroup // has children
)

type DatatypeTreeNode struct {
    Kind       DatatypeNodeKind
    Label      string
    Depth      int
    Expand     bool
    Datatype   *db.Datatypes      // nil for admin mode
    AdminDT    *db.AdminDatatypes  // nil for regular mode
    FieldCount int                 // cached count from field preview fetch
    Children   []*DatatypeTreeNode // child datatypes (sorted by label)
}
```

Builder functions:
- `BuildDatatypeTree(items []db.Datatypes) []*DatatypeTreeNode` -- builds tree from flat list
- `BuildAdminDatatypeTree(items []db.AdminDatatypes) []*DatatypeTreeNode` -- admin variant
- `FlattenDatatypeTree(roots []*DatatypeTreeNode) []*DatatypeTreeNode` -- depth-first flat list, respects `Expand`
- `FilterDatatypeList(items []db.Datatypes, query string) []db.Datatypes` -- case-insensitive label match (pre-build filter)
- `FilterAdminDatatypeList(items []db.AdminDatatypes, query string) []db.AdminDatatypes`

### FieldProperty

File: `internal/tui/field_property.go`

Navigable key-value map for Phase 2 (read-only preview) and Phase 3 (editable).

```go
type FieldProperty struct {
    Key      string // display label: "Label", "Type", "Sort Order", etc.
    Value    string // display value
    Field    string // struct field name for dispatch: "Label", "Type", "SortOrder", etc.
    Editable bool   // false for read-only props (ID, dates)
}

// FieldPropertiesFromField builds the property list for a regular field.
func FieldPropertiesFromField(f db.Fields) []FieldProperty

// FieldPropertiesFromAdminField builds the property list for an admin field.
func FieldPropertiesFromAdminField(f db.AdminFields) []FieldProperty
```

Property rows for `db.Fields`:

| Key           | Field         | Editable | Dialog Type       |
|---------------|---------------|----------|-------------------|
| Label         | Label         | yes      | text input        |
| Name          | Name          | yes      | text input        |
| Type          | Type          | yes      | select (enum)     |
| Sort Order    | SortOrder     | yes      | number input      |
| Data          | Data          | yes      | text input (JSON) |
| Validation    | Validation    | yes      | text input (JSON) |
| UI Config     | UIConfig      | yes      | UIConfig dialog   |
| Translatable  | Translatable  | yes      | toggle (inline)   |
| Roles         | Roles         | yes      | text input        |
| ID            | FieldID       | no       | --                |
| Author        | AuthorID      | no       | --                |
| Created       | DateCreated   | no       | --                |
| Modified      | DateModified  | no       | --                |

## Cursor Management

The screen uses `GridScreen.Cursor` for the active list and `PropertyCursor` for
Phase 3 property navigation. Saved cursor fields handle phase transitions.

### Cursor fields on DatatypesScreen

```go
// Embedded GridScreen provides:
//   Cursor    int  -- active list cursor (datatypes in Phase 1, fields in Phase 2/3)
//   CursorMax int  -- updated on phase transition and data refresh
//   FocusIndex int -- cell focus (Tab cycling)

// Additional fields:
SavedDTCursor int // preserved datatype cursor when entering Phase 2
PropertyCursor int // property row cursor in Phase 3
```

### Phase transition cursor behavior

| Transition | Cursor action |
|------------|---------------|
| Phase 1 -> Phase 2 | `SavedDTCursor = Cursor`, `Cursor = 0`, `CursorMax = fieldsLen()-1` |
| Phase 2 -> Phase 3 | `PropertyCursor = 0`, `FocusIndex = 1` (properties cell) |
| Phase 3 -> Phase 2 | `PropertyCursor` discarded, `FocusIndex = 0` (field list cell) |
| Phase 2 -> Phase 1 | `Cursor = SavedDTCursor`, `CursorMax = flatDTLen()-1`, swap grid |

### Key routing per phase

Phase 1 and Phase 2 use `HandleCommonKeys` for quit/back/cursor movement since
`Cursor` always points at the active list. `CursorMax` is updated on each phase
transition and data refresh.

Phase 3 **does not call `HandleCommonKeys`**. It intercepts Up/Down/Enter/Back
directly and routes them to `PropertyCursor`. Only Quit is delegated. This avoids
`HandleCommonKeys` moving `GridScreen.Cursor` when the user intends to navigate
properties.

```go
// Phase 3 key handling pseudocode:
if Phase == 2 { // Phase 3 = edit
    if ActionUp:   PropertyCursor-- (clamp 0)
    if ActionDown: PropertyCursor++ (clamp len(Properties)-1)
    if Enter:      dispatch edit dialog for Properties[PropertyCursor]
    if Back/Esc:   Phase = 1, FocusIndex = 0
    if Quit:       tea.Quit
    return  // do NOT fall through to HandleCommonKeys
}
```

## Command Pattern

All new commands use `ctx.DB` (Screen pattern), not `db.ConfigDB()` (legacy Model pattern).
This ensures remote mode compatibility and follows the direction of migrated screens.

Audit context is built from `AppContext` fields:

```go
// Inside a tea.Cmd closure:
ac := middleware.AuditContextFromCLI(*ctx.Config, ctx.UserID)
err := ctx.DB.UpdateFieldSortOrder(context.Background(), ac, db.UpdateFieldSortOrderParams{...})
```

Both `UpdateFieldSortOrder` and `UpdateAdminFieldSortOrder` exist on `DbDriver` with
full tri-database implementations (confirmed in `internal/db/field_custom.go` and
`internal/db/admin_field_custom.go`). No new DB layer work is needed for reordering.

## Field Preview Fetch Strategy

In Phase 1, field previews are fetched for the cursor-highlighted datatype. Fetch
triggers on:
- **Explicit cursor movement** (Up/Down): always fetch
- **Expand/Collapse**: fetch only if the cursor position changed (i.e., the node the
  cursor points to after flatten is a different datatype than before)
- **Search filter change**: rebuild tree, clamp cursor, fetch for new cursor position

This avoids unnecessary fetches on expand/collapse when the cursor stays on the same node.

## Steps -- Deliverable A

### Step 1: Create datatype tree node types and builder

File: `internal/tui/datatype_tree.go`

- Define `DatatypeNodeKind`, `DatatypeTreeNode` (with `Children` slice)
- Implement `BuildDatatypeTree` -- index by ID, group by `parent_id`, build tree, assign
  `Kind = DatatypeNodeGroup` for nodes with children, sort children by label
- Implement `BuildAdminDatatypeTree` -- same logic for admin datatypes
- Implement `FlattenDatatypeTree` -- recursive depth-first traversal, skip children of
  collapsed nodes (Expand=false), set Depth on each node
- Implement `FilterDatatypeList` / `FilterAdminDatatypeList` -- case-insensitive label
  substring match, returns filtered flat list (tree is rebuilt from filtered list)

File: `internal/tui/datatype_tree_test.go`

- Test tree building: flat list with parents produces correct hierarchy
- Test flatten: collapsed nodes hide children, expanded nodes show them
- Test flatten: depth values are correct for nested nodes
- Test filter: partial label match, case insensitivity
- Test empty input: no panic on nil/empty slices

### Step 2: Create field property map

File: `internal/tui/field_property.go`

- Define `FieldProperty` struct
- Implement `FieldPropertiesFromField(db.Fields) []FieldProperty`
- Implement `FieldPropertiesFromAdminField(db.AdminFields) []FieldProperty`
- Editable properties first, read-only (ID, author, dates) last
- UIConfig value: display "(none)" for empty/`{}`, otherwise compact summary

File: `internal/tui/field_property_test.go`

- Test property generation: correct keys, values, editability flags
- Test UIConfig display value (JSON summary or "(none)")
- Test admin field properties include correct ID type

### Step 3: Rewrite DatatypesScreen struct

File: `internal/tui/screen_datatypes.go`

Replace the existing struct with GridScreen-based version:

```go
type DatatypesScreen struct {
    GridScreen
    AdminMode bool

    // Phase tracking
    Phase          int // 0=browse, 1=fields, 2=edit (Phase 3 is Deliverable B)
    PropertyCursor int // cursor within property list (Phase 3, Deliverable B)
    SavedDTCursor  int // preserved datatype cursor when in Phase 2+

    // Phase 1: Datatype browse
    Datatypes      []db.Datatypes
    AdminDatatypes []db.AdminDatatypes
    DatatypeTree   []*DatatypeTreeNode
    FlatDTList     []*DatatypeTreeNode
    Searching      bool
    SearchInput    textinput.Model
    SearchQuery    string

    // Phase 2+3: Field selection and editing
    SelectedDTNode *DatatypeTreeNode // the datatype entered in Phase 2
    Fields         []db.Fields
    AdminFields    []db.AdminFields
    Properties     []FieldProperty // built from selected field (read-only in Phase 2)
}
```

- Constructor `NewDatatypesScreen` initializes with `datatypeBrowseGrid`, builds tree
- `rebuildTree()` rebuilds tree + flat list from current data (filtered if searching)
- `flatDTLen()` returns `len(FlatDTList)`
- `fieldsLen()` returns field count for current mode (regular or admin)

### Step 4: Phase 1 -- Datatype browse (Update + View)

File: `internal/tui/screen_datatypes.go`

Update method for Phase 1 (browse):
- `HandleFocusNav` for Tab cycling between cells
- Arrow keys navigate `FlatDTList` via `Cursor`, with field preview fetch on move.
  Track previous cursor position -- only fetch if the underlying datatype changed.
- Expand/Collapse: toggle node `Expand`, re-flatten, clamp cursor, fetch if cursor
  now points at a different datatype
- Enter/Right on a leaf or item node: set `SelectedDTNode`, `SavedDTCursor = Cursor`,
  fetch fields, swap to `datatypeFieldGrid`, set `Phase = 1`, `Cursor = 0`,
  `CursorMax = fieldsLen()-1`, `FocusIndex = 0`
- `/` toggles inline search (like media screen): set `Searching`, focus `SearchInput`,
  on change filter + rebuild tree + clamp cursor
- `n`/`e`/`d` dispatch to existing datatype CRUD dialogs
- `HandleCommonKeys` for quit/back

View method for Phase 1:
- Cell 0: render tree with indentation, expand markers, cursor, search input at top
- Cell 1: render datatype details (label, type, parent chain, ID, field count, dates)
- Cell 2: render field preview list (numbered, type badge per field)

### Step 5: Phase 2 -- Field selection (Update + View)

File: `internal/tui/screen_datatypes.go`

Update method for Phase 2 (field selection):
- `HandleFocusNav` for Tab cycling between cells
- Arrow keys navigate field list via `Cursor`. On cursor change, rebuild `Properties`
  from the highlighted field (read-only preview in center panel).
- Reorder up/down (ActionReorderUp/ActionReorderDown): swap `sort_order` of current
  field with adjacent field. Uses `ctx.DB.UpdateFieldSortOrder` /
  `ctx.DB.UpdateAdminFieldSortOrder` via new standalone command functions (not
  `Model.HandleReorderField`). On success, re-fetch fields and preserve cursor position.
- Enter: transition to Phase 3 -- set `Phase = 2`, `PropertyCursor = 0`,
  `FocusIndex = 1` (properties cell). (Deliverable B activates this. In Deliverable A,
  Enter on a field opens the existing edit-all-at-once field dialog instead.)
- `n`/`e`/`d` dispatch to field CRUD dialogs
- Back/Esc: swap grid back to `datatypeBrowseGrid`, `Phase = 0`,
  `Cursor = SavedDTCursor`, `CursorMax = flatDTLen()-1`, `FocusIndex = 0`
- `HandleCommonKeys` for quit/back/cursor (Cursor = field list cursor here)

View method for Phase 2:
- Cell 0: render field list with cursor, sort order, type badge
- Cell 1: render property preview (key-value pairs, read-only)
- Cell 2: render context breadcrumb (parent chain > datatype label) + key hints

### Step 6: Field reordering commands

File: `internal/tui/commands_datatypes.go` (new file)

New standalone command functions using `ctx.DB` pattern:

```go
// ReorderFieldCmd swaps sort_order between two adjacent fields.
// Uses ctx.DB, not db.ConfigDB(). Returns DatatypeFieldsFetchMsg to reload.
func ReorderFieldCmd(
    cfg *config.Config,
    userID types.UserID,
    driver db.DbDriver,
    fieldA types.FieldID, orderA int64,
    fieldB types.FieldID, orderB int64,
    datatypeID types.NullableDatatypeID,
) tea.Cmd

// ReorderAdminFieldCmd is the admin variant.
func ReorderAdminFieldCmd(
    cfg *config.Config,
    userID types.UserID,
    driver db.DbDriver,
    fieldA types.AdminFieldID, orderA int64,
    fieldB types.AdminFieldID, orderB int64,
    datatypeID types.NullableAdminDatatypeID,
) tea.Cmd
```

These accept the `db.DbDriver` directly (from `ctx.DB`) and build audit context via
`middleware.AuditContextFromCLI(*cfg, userID)`. On success, return a fetch message to
reload the field list. On error, return `ActionResultMsg` with error detail.

### Step 7: Phase 3 -- Per-property field editing (Deliverable B)

**This step is deferred to Deliverable B.** It requires:
- New dialog commands: `ShowEditFieldPropertyDialogCmd`, `ShowFieldTypeSelectDialogCmd`
- New message types for single-property update results
- New `UpdateFieldPropertyCmd` / `UpdateAdminFieldPropertyCmd` command functions
- Phase 3 key handling that bypasses `HandleCommonKeys` (see Cursor Management section)
- Error handling for partial property updates (re-fetch field on any update, success or failure)

Until Deliverable B, Enter on a field in Phase 2 opens the existing edit-all-at-once
field dialog (`ShowEditFieldDialogCmd` / `ShowEditAdminFieldDialogCmd`). The `Properties`
list is rendered read-only in Phase 2's center panel.

### Step 8: Wire fetch messages and data refresh

File: `internal/tui/screen_datatypes.go`

Consolidate existing fetch handlers:
- `AllDatatypesFetchMsg` / `AllDatatypesFetchResultsMsg` -- rebuild tree after fetch
- `AdminAllDatatypesFetchMsg` / `AdminAllDatatypesFetchResultsMsg` -- admin variant
- `DatatypeFieldsFetchMsg` / `DatatypeFieldsFetchResultsMsg` -- update field list,
  rebuild properties if in Phase 2/3
- `AdminDatatypeFieldsFetchMsg` / `AdminDatatypeFieldsFetchResultsMsg` -- admin variant
- Data refresh messages (`AllDatatypesSet`, etc.) -- rebuild tree, clamp cursor
- Field reorder result messages -- re-fetch fields, preserve cursor position
  (field at cursor may have moved; clamp to new bounds)

### Step 9: Extract view rendering

File: `internal/tui/screen_datatypes_view.go`

Move all `render*` methods to a separate view file (pattern matches `screen_media_view.go`):
- `View(ctx AppContext) string` -- dispatches to phase-appropriate cell assembly
- `renderDatatypeTree() string` -- Phase 1 tree with indent, expand icons, cursor, search
- `renderDatatypeDetails() string` -- Phase 1 details cell
- `renderFieldPreview() string` -- Phase 1 field list preview
- `renderFieldList() string` -- Phase 2 field list with cursor, sort order badges
- `renderFieldProperties() string` -- Phase 2 property key-value display (read-only)
- `renderContext() string` -- Phase 2 breadcrumb + key hints

### Step 10: Update page layout and constructors

File: `internal/tui/pages.go`

- Remove `DATATYPES` and `ADMINDATATYPES` entries from `pageLayouts` map
  (GridScreen handles its own layout, legacy entries no longer needed)

File: `internal/tui/constructors.go`

- Update `screenForPage` and any constructor functions to use new `NewDatatypesScreen`
  signature and initialize with tree data

### Step 11: Key hints

File: `internal/tui/screen_datatypes_view.go`

`KeyHints` returns phase-appropriate hints:

Phase 1 (browse):
```
j/k nav | enter select | / search | n new | e edit | d del | tab panel | esc back
```

Phase 2 (fields):
```
j/k nav | J/K reorder | enter edit | n new | e edit all | d del | tab panel | esc back
```

Phase 3 (Deliverable B):
```
j/k nav | enter edit | esc back
```

### Step 12: Tests

File: `internal/tui/screen_datatypes_test.go`

- Phase transitions: browse -> fields -> browse (Phase 3 deferred to Deliverable B)
- Cursor bounds in each phase
- `SavedDTCursor` preserved across Phase 1 -> Phase 2 -> Phase 1 round-trip
- Grid swap on phase transition (`datatypeBrowseGrid` / `datatypeFieldGrid`)
- Search filtering in Phase 1: filter, rebuild, clamp, unfocused hides search
- Field reorder: cursor preserved after sort_order swap
- Property list generation for regular and admin fields
- Admin mode branching: correct IDs, correct fetch messages, correct CRUD dialogs
- Expand/collapse: flatten respects Expand, cursor clamped after collapse
- Empty states: no datatypes, no fields, no properties

## Migration Notes

- The existing `DatatypesScreen` uses `FocusPanel` (TreePanel/ContentPanel/RoutePanel).
  The new version replaces this with `GridScreen.FocusIndex` for cell focus and `Phase`
  for state machine transitions.
- The existing dual cursor (`Cursor` + `FieldCursor`) is replaced by `GridScreen.Cursor`
  (always the active list) + `SavedDTCursor` (preserved on phase transition) +
  `PropertyCursor` (Phase 3 only, Deliverable B).
- The existing `renderActions()` panels (regular + admin) are replaced by the Context
  cell which shows breadcrumb and dynamic key hints.
- The `u` key for UIConfig is removed as a standalone action. In Deliverable A, UIConfig
  is visible in the read-only property preview. In Deliverable B, it becomes a property
  row that opens the existing UIConfig dialog on Enter.
- All existing CRUD dialog commands (`ShowEditDatatypeDialogCmd`, `ShowDeleteFieldDialogCmd`,
  `ShowEditFieldDialogCmd`, etc.) are reused as-is. No new dialog commands in Deliverable A.
- New field reorder commands use `ctx.DB` (Screen pattern), not `db.ConfigDB()` (legacy
  Model pattern). This ensures remote mode compatibility.
- `HandleReorderField` on `Model` is NOT used. New standalone functions accept `db.DbDriver`
  directly and build audit context from `AppContext` fields.
