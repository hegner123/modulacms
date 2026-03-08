# TUI Package Fix Plan

Date: 2026-03-07
Status: Phases 1-8 complete.
Source: TUI package review (125 files, ~700KB)

## Overview

Eight phases, ordered by risk reduction and dependency. Each phase is independently shippable — later phases don't require earlier ones (except Phase 1, which unblocks safe work in Phase 2-3).

Run `just check` after every file change. Run `just test` after each phase.

---

## Phase 1: Fix Bugs (Critical)

**Goal:** Eliminate panics, data corruption risk, and silent failures.

### 1A. Panic guard in update.go

**File:** `update.go:682`

Add nil check before `m.ActiveScreen.PageIndex()`. The surrounding `if m.ActiveScreen != nil` block is violated by handlers like `UpdateDeployCms` that can set ActiveScreen to nil mid-dispatch.

```go
// Before
if _, ok := msg.(HomeDashboardDataMsg); ok {
    utility.DefaultLogger.Fdebug(fmt.Sprintf("... page=%d)", m.ActiveScreen.PageIndex()))
}

// After
if _, ok := msg.(HomeDashboardDataMsg); ok && m.ActiveScreen != nil {
    utility.DefaultLogger.Fdebug(fmt.Sprintf("... page=%d)", m.ActiveScreen.PageIndex()))
}
```

### 1B. Remove duplicate file picker handling

**Files:** `update.go:655-678` and `update.go:748-782`

Two identical file picker blocks exist — one in the Screen path, one in the legacy fallback. Keep only the Screen path version (655-678). The fallback version (748-782) can be deleted once Phase 5 removes the legacy path.

Interim fix: extract to a helper method:

```go
func (m Model) handleFilePicker(msg tea.Msg) (Model, tea.Cmd, bool) {
    // Returns (model, cmd, handled)
    // Single implementation, called from both paths
}
```

### 1C. Propagate tree operation errors

**Files:** `commands.go` (HandleMoveContent, HandleDeleteContent, HandleReorderSibling) and `admin_commands.go` (admin variants)

Currently: nested DB updates log errors but continue execution, leaving the tree in a partial state.

Fix: accumulate errors and return a failure message if any step fails after the first.

```go
// Pattern for each multi-step tree operation:
var errs []error

if err := d.UpdateContentData(ctx, ac, updateA); err != nil {
    errs = append(errs, fmt.Errorf("update node A: %w", err))
}
if err := d.UpdateContentData(ctx, ac, updateB); err != nil {
    errs = append(errs, fmt.Errorf("update node B: %w", err))
}

if len(errs) > 0 {
    return ActionResultMsg{
        Title:   "Partial Failure",
        Message: fmt.Sprintf("Tree operation incomplete: %s", errors.Join(errs...)),
        IsError: true,
    }
}
```

Locations (all follow same pattern):
- `commands.go`: HandleDeleteContent (~line 728, 752, 776), HandleMoveContent (~line 864, 888, 912), HandleReorderSibling (~line 2018, 2042, 2065)
- `admin_commands.go`: HandleDeleteAdminContent, HandleMoveAdminContent, HandleAdminReorderSibling (parallel locations)

### 1D. Fix NumberBubble validation

**File:** `bubble_number.go:40-43`

Current check allows minus at any cursor position. Fix:

```go
case r == '-':
    if b.input.Value() != "" {
        return b, nil // only allow minus as first character of empty input
    }
```

### 1E. Add audit context to remote media upload

**File:** `commands.go` ~line 1927

The remote upload path bypasses AuditContext. Pass it through.

---

## Phase 2: Extract Shared Tree Operations

**Goal:** Eliminate the biggest source of duplication and make tree operations testable.

### New file: `tree_ops.go`

Extract from `commands.go` and `admin_commands.go`:

```
DetachFromSiblings(ctx, ac, driver, node)          error
AttachAsLastChild(ctx, ac, driver, source, parent)  error
SwapSiblingOrder(ctx, ac, driver, nodeA, nodeB)     error
UpdateParentFirstChild(ctx, ac, driver, parent, newFirstChild) error
```

Each function:
- Takes `context.Context`, `audited.AuditContext`, `db.DbDriver`, and typed IDs
- Returns error (no silent swallowing)
- Is unit-testable against a test DB

### New file: `tree_ops_test.go`

Table-driven tests for each operation:
- Detach only child, detach first child, detach middle child, detach last child
- Attach to empty parent, attach to parent with existing children
- Swap adjacent siblings, swap with boundary (first/last)

### Refactor existing commands

Replace inline pointer manipulation in:
- `HandleMoveContent` / `HandleMoveAdminContent` → call `DetachFromSiblings` + `AttachAsLastChild`
- `HandleDeleteContent` / `HandleDeleteAdminContent` → call `DetachFromSiblings` + `UpdateParentFirstChild`
- `HandleReorderSibling` / `HandleAdminReorderSibling` → call `SwapSiblingOrder`

Expected reduction: ~600 lines removed from commands.go, ~400 from admin_commands.go.

---

## Phase 3: Split Large Files

**Goal:** No file over 1000 lines. Each file has a single responsibility.

### 3A. Split update_dialog.go (3,586 lines)

Target files:

| New File | Content | Est. Lines |
|----------|---------|------------|
| `update_dialog.go` | Switch dispatch + generic handlers | ~400 |
| `update_dialog_delete.go` | All DIALOGDELETE* cases | ~400 |
| `update_dialog_publish.go` | Publish/unpublish/restore confirmations | ~300 |
| `update_dialog_form.go` | FormDialogAcceptMsg routing | ~500 |
| `update_dialog_content.go` | Content-specific dialog operations | ~500 |
| `update_dialog_helpers.go` | UpdateRouteFromDialog, UpdateDatatypeFromDialog, etc. | ~600 |

Additionally, extract the repeated delete confirmation pattern into a generic handler:

```go
func handleDeleteConfirmation[T any](
    m Model,
    ctx *T,
    clearCtx func(),
    deleteCmd func(T) tea.Cmd,
) (Model, tea.Cmd) {
    if ctx == nil {
        return m, tea.Batch(OverlayClearCmd(), FocusSetCmd(PAGEFOCUS))
    }
    c := *ctx
    clearCtx()
    return m, tea.Batch(
        OverlayClearCmd(),
        FocusSetCmd(PAGEFOCUS),
        LoadingStartCmd(),
        deleteCmd(c),
    )
}
```

This replaces 13 near-identical cases with one generic function.

### 3B. Split form_dialog.go (2,477 lines)

| New File | Content | Est. Lines |
|----------|---------|------------|
| `form_dialog.go` | FormDialogModel struct + Update + View + generic constructors | ~600 |
| `form_dialog_content.go` | ContentFormDialogModel struct + Update + View | ~800 |
| `form_dialog_constructors.go` | All New*Dialog / NewEdit*Dialog functions | ~600 |

### 3C. Split commands.go (2,400+ lines)

| New File | Content | Est. Lines |
|----------|---------|------------|
| `commands.go` | CRUD commands (create, update, delete for non-tree entities) | ~800 |
| `commands_tree.go` | HandleMoveContent, HandleReorderSibling, HandleCopyContent, HandleDeleteContent (refactored to use tree_ops.go) | ~400 |
| `commands_media.go` | HandleMediaUpload, media-related commands | ~300 |
| `commands_content.go` | Content field batch loading, content instances | ~400 |

---

## Phase 4: Clean Model Struct

**Goal:** Remove stale state, simplify DialogContext.

### 4A. Remove dead fields from Model

Delete from `model.go`:
- `Paginator`, `PageMod`, `MaxRows` — legacy pagination, unused by Screen pages
- `FormState`, `TableState` — never updated in Screen flows
- `QueryResults`, `Time` — completely unused
- `Cursor`, `FocusIndex` — orphaned (AppContext owns these)

After deletion: search for references with `checkfor`, fix any remaining uses (likely in legacy paths that should also be removed in Phase 5).

### 4B. Replace DialogContext with sum type

Current: 42-pointer struct where at most 1 field is non-nil at a time.

Replace with:

```go
type DialogContext struct {
    Active any // nil = no dialog context; type-switch to extract
}

// Usage in update_dialog.go:
switch ctx := m.DCtx.Active.(type) {
case *DeleteContentContext:
    // handle
case *PublishContentContext:
    // handle
}
```

Or keep typed but use a single field with a discriminator:

```go
type DialogContext struct {
    Kind   DialogAction
    Active any
}
```

This eliminates 42 fields and makes "clear all context" trivial: `m.DCtx.Active = nil`.

### 4C. Clear DCtx on dialog cancel

Currently DCtx fields are cleared on accept but not on cancel. Add clearing to all cancel paths in update_dialog.go:

```go
case DialogCancelMsg:
    m.DCtx.Active = nil // with the new sum type, this is one line
    return m, tea.Batch(OverlayClearCmd(), FocusSetCmd(PAGEFOCUS))
```

---

## Phase 5: Remove Legacy Paths

**Goal:** Single dispatch path through Screen interface.

### 5A. Remove legacy fallback in update.go

Delete `update.go:690-717` (the `else` block when `m.ActiveScreen == nil`). All pages should use the Screen interface. If any page still lacks a Screen implementation, implement it first.

Verify: `screenForPage()` returns a non-nil Screen for every `PageIndex` constant. If any return nil, those pages need Screen implementations before this step.

### 5B. Audit all `if m.ActiveScreen != nil` guards

After 5A, ActiveScreen should never be nil after initialization. Remove the 9+ nil checks in update.go that are now unnecessary. Keep only the one in `Init()` where ActiveScreen might not be set yet.

### 5C. Remove legacy rendering in panel_view.go

If `cmsPanelContent()` has cases for pages that now have Screen.View(), remove those legacy render cases.

---

## Phase 6: Consolidate Components

**Goal:** Reduce bubble component duplication, add test coverage.

### 6A. Generic TextInputBubble

Create `bubble_textinput.go`:

```go
type TextInputBubble struct {
    input    textinput.Model
    label    string
    validate func(string) error // nil = no validation
}

func NewTextInputBubble(label, placeholder string, validate func(string) error) TextInputBubble
```

Then reduce `bubble_text.go`, `bubble_email.go`, `bubble_url.go`, `bubble_slug.go` to thin constructors:

```go
func NewEmailBubble() TextInputBubble {
    return NewTextInputBubble("Email", "user@example.com", validateEmail)
}
```

### 6B. Add bubble component tests

New file: `bubble_test.go`

Table-driven tests for:
- NumberBubble: minus position, decimal filtering, max value
- DatePickerBubble: month/year rollover, leap year, minute wrap
- SelectBubble: empty options, cursor bounds, option replacement
- BooleanBubble: toggle behavior

### 6C. Fix SelectBubble and BooleanBubble SetWidth

Either implement real width handling or document why it's a no-op.

### 6D. Standardize Focus() return

All FieldBubble implementations should return `tea.Cmd`. BooleanBubble and SelectBubble currently return nil — acceptable but should be explicit about it.

---

## Phase 7: Code Cleanup

**Goal:** Remove dead code, fix minor issues, improve consistency.

### 7A. Remove dead code

- `GetContentInstances()` in commands.go (returns empty `tea.Batch()`)
- Unused message types: `CmsAddNewContentDataMsg`, `CmsAddNewContentFieldsMsg`, `CmsBuildDefineDatatypeFormMsg`, `TablesFetch`, `FetchHeadersRows`, `GetColumns`
- No-op handler: `BuildTreeFromRouteMsg` → `return m, nil`
- Commented-out functions in debug.go (lines 99-107, 330-398)

### 7B. Build-tag debug.go

Add `//go:build debug` to debug.go so the 400 lines of debug helpers don't compile into production binaries. Create a `debug_stub.go` with `//go:build !debug` containing empty stubs for any debug methods called from non-debug code.

### 7C. Fix minor issues

- `field_property.go`: Delete duplicate `uiConfigSummary()` (identical to `compactJSON()`)
- `type_registry.go`: Log or error on duplicate key registration
- Extract dialog width constant: `const dialogBorderPadding = 6`
- `parse.go`: Leave as-is (boilerplate but correct; code generation would add build complexity for marginal gain)
- Fix inconsistent message naming: rename `FieldReorderResultMsg` → `FieldReorderedMsg` and `AdminFieldReorderResultMsg` → `AdminFieldReorderedMsg`
- `Init()`: Only fire `HomeDashboardFetchCmd` when `m.Page == HOMEPAGE`

### 7D. Add default case logging in update_dialog.go

```go
default:
    if m.Logger != nil {
        m.Logger.Fwarn(fmt.Sprintf("UpdateDialog: unhandled message type %T", msg))
    }
    return m, nil
```

---

## Phase 8: Consolidate Admin/Regular Handlers

**Goal:** Reduce the ~50% duplication between admin and regular command/dialog paths.

This is the largest refactor and should be done last, after Phases 1-7 have stabilized the codebase.

### 8A. Parameterized tree operations (DONE)

Created `content_ops.go` with a `treeOps` abstraction that normalizes tree pointer
manipulation for both regular and admin content. This eliminated the 6 duplicated
tree functions in `tree_ops.go` (~800 lines) and replaced them with 4 unified
functions (~390 lines):

- `treeOps` struct with `getNode`/`updateNode` function fields
- `newContentTreeOps(d)` / `newAdminTreeOps(d)` constructors
- `contentToTreeNode()` / `adminContentToTreeNode()` conversion helpers
- `detachFromSiblings()` — unified from `detachContentFromSiblings` + `detachAdminContentFromSiblings`
- `attachAsLastChild()` — unified from `attachContentAsLastChild` + `attachAdminContentAsLastChild`
- `swapSiblings()` — unified from `swapContentSiblings` + `swapAdminContentSiblings`
- `spliceAfter()` — new helper used by both copy handlers

Callers updated: `HandleDeleteContent`, `HandleMoveContent`, `HandleReorderSibling`,
`HandleCopyContent` (commands_tree.go) and `HandleDeleteAdminContent`,
`HandleMoveAdminContent`, `HandleAdminReorderSibling`, `HandleCopyAdminContent`
(admin_commands.go).

Full handler unification (merging each handler pair into a single function) was
evaluated but deferred — the adapter layer for different message types and return
types adds complexity that outweighs the ~280 lines of handler duplication savings.

### 8B. Parameterized dialog handlers (DONE)

Consolidated 5 admin dialog action/context pairs into their regular counterparts:

- `DeleteAdminContentContext` → merged into `DeleteContentContext` with `AdminMode bool`
- `PublishAdminContentContext` → merged into `PublishContentContext` with `AdminMode bool`
- `RestoreAdminVersionContext` → merged into `RestoreVersionContext` with `AdminMode bool`
- `DeleteAdminContentFieldContext` → merged into `DeleteContentFieldContext` with `AdminMode bool`
- 5 admin dialog action constants removed (`DIALOGDELETEADMINCONTENT`, `DIALOGPUBLISHADMINCONTENT`,
  `DIALOGUNPUBLISHADMINCONTENT`, `DIALOGRESTOREADMINVERSION`, `DIALOGDELETEADMINCONTENTFIELD`)

Admin show-dialog handlers in `update_dialog_admin.go` now store regular context structs
with `AdminMode: true` and use regular dialog action constants. The accept handler in
`update_dialog_delete.go` checks `AdminMode` to produce the correct confirmed message type.

### 8C. Consolidate message types (DONE)

Replaced 11 parallel admin/regular result message types with unified types carrying
`AdminMode bool`. Eliminated types:

- `AdminContentCreatedMsg` → `ContentCreatedMsg` (also renamed `ContentDataID` → `ContentID`)
- `AdminContentDeletedMsg` → `ContentDeletedMsg` (upgraded from `string` to typed IDs)
- `AdminContentReorderedMsg` → `ContentReorderedMsg`
- `AdminContentCopiedMsg` → `ContentCopiedMsg`
- `AdminContentMovedMsg` → `ContentMovedMsg`
- `AdminPublishCompletedMsg` → `PublishCompletedMsg`
- `AdminUnpublishCompletedMsg` → `UnpublishCompletedMsg`
- `AdminVersionRestoredMsg` → `VersionRestoredMsg`
- `AdminContentFieldUpdatedMsg` → `ContentFieldUpdatedMsg`
- `AdminContentFieldAddedMsg` → `ContentFieldAddedMsg`
- `AdminContentFieldDeletedMsg` → `ContentFieldDeletedMsg`

Producers in `admin_commands.go` cast admin typed IDs to regular types
(safe — both are `string` aliases). Consumers in `screen_content.go`,
`update_cms.go`, and `admin_update_cms.go` use `msg.AdminMode` to select
admin vs regular reload commands. Removed ~80 lines of dead admin result
handlers from `admin_update_cms.go`.

---

## Verification

After each phase:
1. `just check` — compile verification
2. `just test` — unit tests pass
3. Manual TUI smoke test via `just run` — navigate all screens, create/edit/delete content, reorder siblings

After all phases:
- Line count for `internal/tui/` should drop by ~2,000-3,000 lines
- No file over 1,000 lines
- Zero silent DB error swallowing in tree operations
- All bubble components have test coverage
