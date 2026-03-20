# Block Editor

The block editor (`<block-editor>`) is a Web Component that implements a nested document tree with drag-and-drop, undo/redo, field editing, and keyboard shortcuts. It uses a sibling-pointer tree structure (parent, firstChild, prevSibling, nextSibling) for O(1) navigation and reordering. All state lives client-side; the editor communicates with the server only for saves and datatype fetches.

**Source:** `internal/admin/static/js/block-editor-src/`
**Bundle:** `internal/admin/static/js/block-editor.js` (built via `just admin bundle`)
**Styles:** `internal/admin/static/css/block-editor.css`
**Template:** `internal/admin/pages/content_edit.templ`

## Source Files

| File | Purpose |
|------|---------|
| `index.js` | BlockEditor class, lifecycle, rendering, event routing |
| `state.js` | State shape, initialization |
| `tree-ops.js` | Tree mutation operations (add, move, delete, indent, outdent, duplicate) |
| `tree-queries.js` | Read-only tree traversal (children, depth, descendants, DFS order) |
| `drag.js` | Pointer-based drag-and-drop (threshold, overlay, drop zones, auto-scroll) |
| `history.js` | Undo/redo snapshot stack with field-edit batching |
| `config.js` | Block type definitions, MAX_DEPTH constant |
| `dom-patches.js` | Surgical DOM updates after state mutations |
| `picker.js` | Block picker modal (telescope/fzf-style datatype search) |
| `id.js` | Client-side UUID generation |
| `cache.js` | Datatype fetch cache (5-minute TTL, request dedup) |
| `validate.js` | Pointer consistency checks (dev-mode) |
| `styles.js` | Injected styles (legacy, most styles now in block-editor.css) |

## State Shape

```javascript
{
  blocks: {              // Map<blockId, Block>
    "id-1": {
      id: "id-1",
      type: "text",            // "text", "heading", "image", "container", or custom
      parentId: null,           // null for root blocks
      firstChildId: null,
      prevSiblingId: null,
      nextSiblingId: "id-2",
      label: "Text Block",
      datatypeId: "01ABC...",   // datatype definition ID
      authorId: "01XYZ...",
      routeId: "01RTN...",
      status: "draft",
      dateCreated: "2026-03-20T...",
      dateModified: "2026-03-20T...",
      fields: [
        { fieldId: "01FLD...", label: "Title", value: "Hello" },
      ]
    },
  },
  rootId: "id-1",         // first root-level block
  selectedBlockId: null,   // currently selected block
  dirty: false             // true after mutation, cleared on save
}
```

## Tree Operations (`tree-ops.js`)

All functions mutate state in place. Caller is responsible for pushing undo snapshots before calling.

| Function | Signature | Purpose |
|----------|-----------|---------|
| `unlink` | `(state, blockId)` | Detach block from sibling chain, update parent pointers |
| `insertBefore` | `(state, blockId, targetId)` | Insert before target in sibling chain |
| `insertAfter` | `(state, blockId, targetId)` | Insert after target |
| `insertAsFirstChild` | `(state, blockId, parentId)` | Reparent as first child |
| `insertAsLastChild` | `(state, blockId, parentId)` | Reparent as last child |
| `addBlock` | `(state, type, afterId?) → id` | Create new block, insert at end or after afterId |
| `addBlockFromDatatype` | `(state, datatype, position, targetId) → id` | Create from datatype definition |
| `removeBlock` | `(state, blockId) → string[]` | Delete block and descendants, return removed IDs |
| `moveBlock` | `(state, blockId, targetId, position)` | Move to before/after/inside target |
| `indentBlock` | `(state, blockId) → boolean` | Make last child of previous sibling |
| `outdentBlock` | `(state, blockId) → boolean` | Move out of parent, reparent younger siblings |
| `duplicateBlock` | `(state, blockId) → string \| null` | Deep-clone subtree with new IDs |
| `moveBlockUp` | `(state, blockId) → boolean` | Swap with previous sibling |
| `moveBlockDown` | `(state, blockId) → boolean` | Swap with next sibling |

## Tree Queries (`tree-queries.js`)

All functions are read-only.

| Function | Returns | Purpose |
|----------|---------|---------|
| `getChildren(state, parentId)` | `Block[]` | Direct children via sibling chain |
| `getRootList(state)` | `Block[]` | All root-level blocks |
| `isDescendant(state, candidateId, ancestorId)` | `boolean` | Ancestry check |
| `getDepth(state, blockId)` | `number` | Parent chain length (root = 0) |
| `findLastSibling(state, blockId)` | `string` | Last block in sibling chain |
| `getBlockTraversalOrder(state)` | `string[]` | Depth-first traversal |
| `collectDescendants(state, blockId)` | `string[]` | All descendant IDs |
| `getDescendantCount(state, blockId)` | `number` | Count descendants |

## DOM Structure

### Template embedding

```html
<block-editor
  id="content-block-editor"
  data-content-id="01ABC..."
  data-root-datatype-id="01DTY..."
  data-state='{"blocks":{...},"rootId":"..."}'>
</block-editor>
```

Attributes:
- `data-state` — JSON string consumed and removed on init
- `data-content-id` — for recovery key and save identification
- `data-root-datatype-id` — root datatype for block picker hierarchy

### Rendered DOM

```
block-editor
└─ .editor-container [data-editor-container]
   ├─ .editor-header
   │  ├─ .collapse-controls
   │  │  ├─ button[data-action="expand-all"]
   │  │  └─ button[data-action="collapse-all"]
   │  └─ button.save-btn[data-action="save"]
   │
   └─ .block-list
      ├─ .block-wrapper[data-block-id]          ← per block
      │  ├─ .block-item[data-block-id]          ← header (draggable)
      │  │  ├─ button.block-chevron             ← collapse toggle (▾/▸)
      │  │  ├─ span.block-type-badge            ← colored type label
      │  │  ├─ span.block-label                 ← block name
      │  │  ├─ span.child-count-badge           ← descendant count (if > 0)
      │  │  ├─ button.block-kebab               ← context menu (⋮)
      │  │  └─ .block-content-preview           ← field value previews
      │  │     └─ .preview-field (repeated)
      │  │        ├─ .preview-field-label
      │  │        └─ .preview-field-value
      │  │
      │  └─ .children-container[data-parent-id] ← created on demand
      │     └─ .block-wrapper (recursive)
      │
      └─ .insert-empty                          ← shown when no blocks
         └─ button.insert-btn--empty[data-action="insert"]
```

### Overlays (appended during interaction)

- `.drag-overlay` — cloned block header, `position: fixed; top: 0; left: 0`, positioned via `transform: translate()`
- `.drop-indicator` — 2px accent line, `position: absolute` in `.block-list`
- `.block-context-menu` — kebab menu, `position: fixed`, auto-flips at viewport edges
- `.block-picker-backdrop > .block-picker` — modal with search input and category results

## Drag-and-Drop

### Flow

1. **Pointer down** on `.block-item` — record start position, attach pre-threshold handlers
2. **Pre-threshold move** — track distance; if < 8px, ignore (will become a click)
3. **Threshold crossed (8px)** — `_startDrag()`:
   - Compute grab offset from block's top-left corner
   - Clone block header as `.drag-overlay` (position: fixed, top: 0, left: 0)
   - Position via `transform: translate(clientX - grabOffsetX, clientY - grabOffsetY)`
   - Set pointer capture on block header
   - Add `.dragging` class (opacity 0.3)
4. **Drag move** — update overlay transform, start auto-scroll if near edges, compute drop zone
5. **Drop zone detection** (`_computeDropZone`) — for each visible `.block-item`:
   - Top 25% → `before`, bottom 25% → `after`, middle → `inside`
   - Coerce `inside` to `after` if type can't have children or depth limit reached
   - Skip descendants of dragged block (prevents circular nesting)
6. **Drop indicator** — line for before/after, highlight class for inside
7. **Pointer up** — execute drop: push undo, mutate state via `moveBlock()`, patch DOM
8. **Cleanup** — remove overlay, drop indicator, release pointer, clear drag state

### Auto-scroll

During drag, if pointer is within 40px of the editor container's top or bottom edge, `requestAnimationFrame` scrolls proportionally (max 12px/frame). Stops when pointer leaves edge zone.

### Depth limit

`MAX_DEPTH = 8`. Enforced in drop zone coercion and indent operations. Blocks cannot be nested deeper than 8 levels.

## Undo/Redo (`history.js`)

### Architecture

- Two stacks: undo (50 entries max) and redo
- Each entry: `{ snapshot: { blocks, rootId }, selectedBlockId }`
- Snapshots are deep clones of state (blocks + fields)
- Every mutation pushes an undo snapshot before modifying state

### Field edit batching

`pushFieldChange(state)` groups rapid keystrokes into a single undo entry. First call pushes a snapshot; subsequent calls within 500ms are no-ops. After 500ms silence, the batch closes and the next edit starts a new entry.

### Keyboard shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+Z / Cmd+Z | Undo |
| Ctrl+Shift+Z / Cmd+Shift+Z / Ctrl+Y | Redo |

### Restore flow

`_restoreSnapshot(entry)` replaces `_state.blocks` and `rootId`, does a full re-render, restores selection, and dispatches `block-editor:change`.

### ID remapping

After save, `History.remapIds(idMap)` updates all client UUIDs to server ULIDs in every snapshot on both stacks.

## Block Picker

A telescope/fzf-style modal for inserting blocks by datatype.

### Opening

Triggered by Enter key, kebab menu "Add" items, or the empty-state `+` button. Opens with an insert target (blockId + position) or `root` for empty editors.

### Data flow

1. Fetch datatypes from `/admin/api/datatypes` (cached 5 minutes, deduped)
2. Group into categories: Root Type children, Collections, Global
3. Filter system types: `_root`, `_nested_root`, `_system_log`, `_reference`

### UI

- Backdrop overlay (click to dismiss)
- Results area with category headers and selectable items (indented by depth)
- Input bar with `>` prompt and live query text
- Keyboard: arrows to navigate, Enter to insert, Escape to close, type to filter

### Insertion

Selected datatype → `addBlockFromDatatype(state, datatype, position, targetId)` → re-render → auto-select new block.

## Save and Recovery

### Save

1. User clicks Save → `serialize()` produces JSON of blocks + rootId
2. Editor emits `block-editor:save` event with `{ state: json }`
3. Host application POSTs to server, receives idMap in response
4. Host calls `editor.commitSave(idMap)`:
   - Remaps all client UUIDs to server ULIDs in state, registries, DOM attributes, and history
   - Snapshots state as new baseline
   - Clears dirty flag and recovery data

### Crash recovery

- On every mutation, schedules a 30-second debounced write to `sessionStorage` (key: `mcms-block-recovery:{contentId}`)
- On editor mount, checks for recovery data and shows a banner with Restore/Dismiss
- On `beforeunload` with dirty state, flushes recovery and shows browser confirmation dialog
- Recovery cleared on successful save or dismiss

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| J / ↓ | Navigate down (DFS order) |
| K / ↑ | Navigate up (DFS order) |
| H / ← | Navigate to parent |
| L / → | Navigate to first child |
| Tab | Select next block |
| Shift+Tab | Select previous block |
| > (Shift+.) | Indent selected block |
| < (Shift+,) | Outdent selected block |
| Enter | Open block picker (insert after selected) |
| Delete / Backspace | Delete selected block |
| Ctrl+Shift+D / Cmd+Shift+D | Duplicate selected block |
| Ctrl+Z / Cmd+Z | Undo |
| Ctrl+Shift+Z / Cmd+Shift+Z / Ctrl+Y | Redo |
| Escape | Cancel drag or close picker |

All shortcuts are scoped to `.editor-container`. Picker keys override editor keys when the picker is open.

## Custom Events

| Event | Detail | When |
|-------|--------|------|
| `block-editor:change` | `{ action, blockId, [targetId, position] }` | After any mutation |
| `block-editor:select` | `{ blockId }` or `{ blockId: null }` | Block selection change |
| `block-editor:save` | `{ state: JSON string }` | Save button clicked |
| `block-editor:error` | `{ message, error }` | Init parse or validation error |

## Field Panel Integration

The editor and the field panel (right side of the content edit page) communicate via events and method calls:

1. **Selection** — editor emits `block-editor:select` → host fetches fields for selected block → renders `<mcms-field-renderer>` components in `#panel-content`
2. **Field change** — user edits a field renderer → host calls `editor.setFieldValue(blockId, fieldId, value)` → state updated, undo batched, preview refreshed
3. **Content preview** — each block header shows truncated field values (120 char limit, ULID refs shown as "(ref)")

## Collapse/Expand

- `_collapsedBlocks` Set tracks which blocks are collapsed
- Toggle adds/removes `collapsed` class on `.block-wrapper`
- CSS hides `.children-container` and `.block-content-preview` for collapsed blocks
- Chevron character: ▾ (expanded) / ▸ (collapsed)
- Expand All / Collapse All buttons in editor header

## Kebab Context Menu

Right-side `⋮` button on each block header opens a fixed-position menu:

```
Add
  ├─ Add Before
  ├─ Add After
  └─ Add Inside (if type supports children and depth < MAX)
Move
  ├─ Move Up / Move Down
  ├─ Indent / Outdent
─────────
  ├─ Duplicate
  └─ Delete (destructive)
```

Auto-flips up if clipped at viewport bottom. Dismissed on outside click or Escape.

## Block Type Configuration

```javascript
{
  text:      { label: 'Text',      canHaveChildren: false },
  heading:   { label: 'Heading',   canHaveChildren: false },
  image:     { label: 'Image',     canHaveChildren: false },
  container: { label: 'Container', canHaveChildren: true },
}
```

Unknown types default to `canHaveChildren: true`. The `canHaveChildren` flag controls:
- Whether "Add Inside" appears in the kebab menu
- Whether "inside" drop zone is available during drag
- Whether indent operation is allowed (target must accept children)

## ID Generation and Remapping

- Client-side: `crypto.randomUUID()` generates UUID v4 for new blocks
- Server-side: ULIDs generated on save
- `commitSave(idMap)` remaps all UUIDs to ULIDs across state, registries, DOM attributes, and history stacks

## Indentation

Block nesting depth is visualized via `marginInlineStart = depth * 24px` on `.block-wrapper`. After any move/indent/outdent, `_updateDescendantDepths(blockId)` recursively updates all descendants.

## Validation (`validate.js`)

`validateState(state)` checks pointer consistency (dev-mode only):
- Reciprocal sibling pointers (A.next = B ↔ B.prev = A)
- Parent-child consistency (parent.firstChild has correct parentId)
- Root chain reachability
- No cycles in sibling chains
- All references point to existing blocks

Returns array of error messages. Called after every mutation via `_devValidate()`.

## CSS Architecture

All styles scoped to `block-editor` element selector. Key CSS custom properties:

```css
block-editor {
  --bg, --text, --border, --accent, --danger
  --badge-bg, --badge-text
  --sp-1 through --sp-8        /* spacing scale */
  --fs-xs through --fs-base    /* font sizes */
  --radius-sm through --radius-xl
}
```

The `.content-editor-layout` grid (defined in block-editor.css) places the editor and field panel side by side at 5fr / 7fr. Both sections expand to their full content height — the single scroll region is `#main-content`.
