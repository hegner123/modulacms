# Block Editor — Technical Analysis

Comprehensive analysis of the ModulaCMS block editor: state management, bugs, styling, UX, integration patterns, extension architecture, and market position.

## 1. State Management

### Current Architecture

The editor uses a **mutable flat hashmap** (`state.js`):

```javascript
{ blocks: {}, rootId: null, selectedBlockId: null, dirty: false }
```

All tree operations in `tree-ops.js` mutate state directly. There is no immutability, no copy-on-write, and no transaction boundary. A failed operation mid-execution leaves state partially mutated.

**Snapshot cloning** (`History._cloneState`, history.js:15) performs a shallow-plus-one-level clone: it spreads each block object and maps the `fields` array with shallow field spreads. The same clone logic is duplicated in `BlockEditor._cloneBlocks` (index.js:122) — a DRY violation where one could be updated without the other.

**Memory**: Each undo entry stores a full copy of every block. With 50 entries and 200 blocks, this is manageable. With 1000+ blocks it becomes significant. No structural sharing — changing one block's label snapshots all blocks.

### How Major Editors Handle State

**ProseMirror** — Immutable document tree. Mutations produce `Transaction` objects containing invertible `Step` operations. New `EditorState` is derived from old state + transaction. Unchanged subtrees are shared between old and new states (structural sharing).

**Lexical (Meta)** — Mutable model behind a controlled gate. All mutations happen inside `editor.update()` callbacks. The editor snapshots before and diffs after, giving transaction semantics (atomicity via try/catch + rollback) and operation batching.

**Gutenberg** — Redux-based via `@wordpress/data`. Immutable state, action dispatch, selector memoization. Undo/redo via action replay or state snapshots.

**Slate.js** — All changes go through `Transforms` producing invertible, serializable operations. Editor normalizes after each change.

**Editor.js** — Flat JSON array, full-state snapshots for undo (similar to current approach).

### Patterns That Would Improve This Editor

**Operation-based history** (from ProseMirror/Lexical): Record discrete operations (`{ type: 'move', blockId, from, to }`) with inverses. Reduces memory from O(n * history_size) to O(ops * history_size) and enables collaborative editing later.

**Transaction boundaries** (from Lexical): Wrap mutations in `withTransaction(state, fn)` that clones state before `fn`, rolls back on throw, and only fires events/patches after success. Currently if `moveBlock` succeeds but DOM patching fails in `_executeDrop` (drag.js:367-368), state is mutated with no rollback.

**Structural sharing** (from ProseMirror/Redux): Only clone blocks that actually changed. A move operation on 500 blocks changes ~4-6 pointer fields. Copying all 500 is wasteful.

**Controlled mutation gate** (from Lexical): Single `editor.apply(operation)` method centralizing dirty marking, history, validation, DOM patching, and events. Currently scattered across `dom-patches.js`, `drag.js`, `picker.js`, and `index.js`.

### CRDT Relevance

The sibling-pointer model is structurally similar to Yjs's linked-list CRDT, making it a natural fit for collaborative editing if ever needed. The main prerequisite would be switching from full-state snapshots to an operation log.

---

## 2. Bugs

### Race Conditions

**BUG-1: Picker async fetch + stale reference** (picker.js:18-25)
`_openPicker` sets `_pickerOpen = true`, then issues an async `fetchDatatypesGrouped`. If the user rapidly opens and closes the picker, the `.then` callback calls `_renderPicker()` even after `_closePicker()` set the flag to `false`. No guard in the fetch callback. Result: phantom picker appears after dismissal.

**BUG-2: Drag zombie on concurrent undo** (drag.js:90, index.js:373)
`setPointerCapture` captures the pointer on the block-item element. If a concurrent undo (via window-level keydown) triggers `_render()` which clears `innerHTML`, the captured element is removed from the DOM. `pointermove`/`pointerup` events stop firing. `this._drag` remains non-null, overlay stays in DOM, auto-scroll keeps running. Only recoverable via Escape key. No `lostpointercapture` handler exists.

**BUG-3: Kebab double-open** (index.js:736-746)
Outside-click handler registration is inside `requestAnimationFrame`. During the frame gap, a fast second click on a different kebab button could open two menus simultaneously.

### Data Loss

**BUG-4: `duplicateBlock` drops content fields** (tree-ops.js:324-332)
The clone only copies `id`, `type`, `parentId`, `firstChildId`, `prevSiblingId`, `nextSiblingId`, and `label`. Does **not** copy `datatypeId`, `authorId`, `routeId`, `status`, `dateCreated`, `dateModified`, or `fields`. Blocks created via `addBlockFromDatatype` (line 130-165) include these properties, so duplicating them loses all content data.

### DOM-State Desync

**BUG-5: Silent DOM patch skip** (drag.js:355-438)
If `_wrapperRegistry.get(blockId)` returns null (race condition, concurrent undo, or missed registration), the state mutation at line 368 has already happened. The `if (blockWrapper && targetWrapper)` guard on line 375 skips DOM patching silently. State says block X is after block Y, but DOM shows X in its original position. No error reported. Only fixed on next full `_render()`.

**BUG-6: `_render()` does not cancel active drag** (index.js:454-497)
`_render()` clears registries and destroys DOM but does not call `_cleanupDrag()`. An active drag session is left in zombie state.

**BUG-7: `_doAddBlock` ignores empty-state element** (index.js:938-955)
`addBlock` appends to root list. `_doAddBlock` appends wrapper to `.block-list`. If the list contains the empty-state centered plus button, the new block appears after it instead of replacing it.

### Tree Operation Edge Cases

**BUG-8: `findLastSibling` infinite loop on corrupted state** (tree-queries.js:47-53)
If a sibling chain has a cycle (A.next -> B, B.next -> A), this loops forever. No cycle detection. `validateState` catches cycles, but `findLastSibling` is called during mutations before validation runs.

**BUG-9: `insertBefore`/`insertAfter` without prior `unlink`** (tree-ops.js:36-75)
These set pointers but never unlink first. If called on an already-linked block without `unlink`, the block exists in two sibling chains simultaneously. Callers are responsible, but direct callers could forget.

### Recovery & History

**BUG-10: No schema version in recovery data** (index.js:234)
Recovery JSON has `{ state, timestamp }` but no version marker. If the block format changes between releases, recovered data produces blocks with missing properties. `validateState` catches pointer issues but not missing content fields.

**BUG-11: Non-atomic ID remap** (history.js:107-115)
If two client IDs remap to the same server ID (server bug), the second write silently overwrites the first. No duplicate detection.

### Browser Compatibility

**COMPAT-1: `crypto.randomUUID()`** (id.js:4) — Fails in non-secure contexts (HTTP). Not in Safari < 15.4, Chrome < 92, Firefox < 95.

**COMPAT-2: `color-mix()` CSS** (block-editor.css:247) — Not in Safari < 16.2, Chrome < 111. Declaration ignored; container blocks get no background.

---

## 3. Styling

### Current Approach

CSS custom properties scoped to `block-editor` element (block-editor.css:4-43). Local aliases (`--bg`, `--border`, `--accent`) bridge from the admin panel's design tokens. No Shadow DOM — selectors scope via element prefix.

**Strengths**: Theming propagates automatically. No build step. Natural scoping.

**Weaknesses**:
- **Hardcoded badge colors** (lines 348-366): `#2a1e45`, `#cba6f7` etc. are dark-theme-only, not using design tokens. Light theme would fail.
- **Mixed spacing systems**: Editor uses `--sp-1` through `--sp-8`, but picker uses `--space-2` directly. The mapping at line 33 is incorrect: `--sp-6: var(--space-5)`.
- **Magic z-index**: Context menu uses `z-index: 9999` (line 396) instead of `var(--z-*)` tokens.
- **No responsive nesting**: 8 levels x 24px = 192px indentation. On narrow viewports, deep blocks push content off-screen. No container queries.

### How Gutenberg Handles Styling

`theme.json` declares design tokens → CSS custom properties. `block.json` per block type declares supported style attributes. CSS modules for editor styles. Design tokens cascade from theme to block to inline.

### Animation Gaps

- No FLIP animation for drop reordering (blocks teleport to new position)
- No entry/exit animations for block add/remove
- No pulse/fade on drop indicator
- The overlay uses `position: fixed` + `transform: translate()` — correct and performant, but actual blocks feel abrupt on reorder

### Accessibility (WCAG 2.1 AA)

**Present**: Focusable container (`tabindex="0"`), comprehensive keyboard navigation (vim + arrows + tab), picker returns focus on close.

**Missing**:
- No `role="tree"`, `role="treeitem"`, `role="group"` — screen readers see flat divs
- No `aria-expanded` on collapse chevrons
- No `aria-selected` on selected block
- No `aria-label` on editor container
- No `aria-live` region for announcing mutations
- No keyboard mechanism for drag-and-drop reordering
- `opacity: 0.3` on dragging block (line 122) reduces contrast below AA thresholds
- `opacity: 0.4` on disabled save (line 77) may fail contrast

**WCAG failures**: 1.3.1 (tree structure not conveyed via ARIA), 2.1.1 (drag-and-drop not keyboard-accessible), 4.1.2 (interactive elements lack accessible names).

---

## 4. UX Controls

### Drag-and-Drop Comparison

| Aspect | Current (Pointer Events) | HTML5 DnD | dnd-kit | pragmatic-drag-and-drop |
|--------|--------------------------|-----------|---------|--------------------------|
| Touch support | Yes (pointer events) | No | Yes | Yes |
| Keyboard DnD | No | No | Yes (built-in) | No |
| Custom drag preview | Yes (cloned element) | Limited | Yes | Limited |
| Tree support | Custom three-zone | Manual | Sortable preset | Tree adapter |
| ARIA announcements | No | No | Yes (built-in) | No |
| Framework | None (vanilla) | None | React | Framework-agnostic |

Current weaknesses: No keyboard DnD, 8px threshold too small for touch (should be 10-15px), `_computeDropZone` scans all blocks every frame (O(n) per `pointermove`).

### Selection Model

Single selection only. Missing: Shift+click range select, Ctrl+click toggle select, Shift+arrow range extend. All tree editors that support batch operations (Figma, VS Code explorer, Gutenberg) offer multi-select.

### Inline vs Panel Editing

Panel editing (current) is the right choice for a headless CMS with complex structured fields. Inline editing (Gutenberg, Notion) is better for content-heavy workflows but makes tree navigation visually noisy. The split panel layout correctly separates structure from data.

### Toolbar Patterns

The kebab menu requires two clicks for every action. Gutenberg uses a floating toolbar above the selected block for frequent operations. Adding a compact inline toolbar (move up/down arrows, drag handle, type indicator) would reduce friction without abandoning the kebab for less-common actions.

---

## 5. UI Integration Patterns

### Current Pattern

Unidirectional events upward (`block-editor:change`, `block-editor:select`, `block-editor:save`), imperative methods downward (`commitSave(idMap)`, `setFieldValue(blockId, fieldId, value)`). No shared stores.

### Autosave Comparison

| Editor | Autosave Strategy |
|--------|-------------------|
| Gutenberg | Server autosave every 60s + sessionStorage local backup |
| Notion | Save on every keystroke with debouncing |
| Google Docs | Continuous via OT, no save button |
| Current | Manual save only. sessionStorage recovery after 30s inactivity |

### Field Validation

The editor performs no field-level validation. `setFieldValue` (index.js:344) writes any value unchecked. The editor could accept a validation schema per datatype and show inline errors.

---

## 6. Extension Architecture

### Current Config

```javascript
BLOCK_TYPE_CONFIG = {
    text:      { label: 'Text',      canHaveChildren: false },
    heading:   { label: 'Heading',   canHaveChildren: false },
    image:     { label: 'Image',     canHaveChildren: false },
    container: { label: 'Container', canHaveChildren: true },
};
```

Unknown types fallback to `canHaveChildren: true`. The picker uses server datatypes, not this config. This is extensible by accident.

### What an Extension API Would Need

1. **`registerBlockType(name, config)`** — icon, badge color, preview renderer, constraints (max children, allowed child types, required fields)
2. **Custom field editors** — plugin-provided web components for editing
3. **Lifecycle hooks** — `onBeforeAdd`, `onAfterRemove`, `onBeforeMove` for validation and side effects
4. **Schema declaration** — block data shape for validation and serialization
5. **Transform rules** — converting between block types while preserving compatible fields

Gutenberg's `block.json` + `registerBlockType` API is the gold standard reference.

---

## 7. Market Position

### Feature Matrix

| Feature | ModulaCMS | Gutenberg | Editor.js | Notion | Sanity PT |
|---------|-----------|-----------|-----------|--------|-----------|
| Nested tree structure | Yes | Yes | No | Yes | Yes |
| Drag-and-drop | Yes | Yes | Yes | Yes | Limited |
| Undo/redo | Yes | Yes | Yes | Yes | Yes |
| Collaborative editing | No | No | No | Yes | Yes |
| Custom block types | Via datatypes | registerBlockType | Tool API | No | Schema |
| Keyboard navigation | Yes (vim) | Yes | Limited | Yes | Limited |
| Block picker | Telescope/fzf | Inserter panel | Toolbox | Slash commands | No |
| Inline editing | No | Yes | Yes | Yes | No |
| Multi-select | No | Yes | No | Yes | No |
| ARIA tree semantics | No | Yes | No | Partial | Partial |
| Framework | None | React | Vanilla | React | React |
| Bundle size | ~15KB | ~500KB | ~35KB | N/A | ~150KB |

### Unique Strengths

1. **Sibling-pointer tree** — Same data structure as server content tree. Zero conversion between editor and API.
2. **Zero dependencies** — Pure vanilla JS web component. Embeddable in any frontend.
3. **Server-driven block types** — Datatypes defined by administrators, not code.
4. **Telescope/fzf picker** — Power-user-friendly keyboard-driven insertion.
5. **Headless CMS fit** — Pure structural editor, not WYSIWYG. Rendering is the frontend's job.

### Key Gaps

1. Multi-block selection for batch operations
2. Keyboard-accessible reordering (Alt+Arrow shortcuts)
3. ARIA tree semantics for screen readers
4. Block transforms (change type preserving fields)
5. Copy/paste blocks
6. Autosave
7. Drop animations

---

## Priority Summary

### Critical Bugs (Data Loss / Broken State)
1. **BUG-4**: `duplicateBlock` drops `fields`, `datatypeId`, and all content properties
2. **BUG-2**: Drag zombie on concurrent undo — no `lostpointercapture` handler
3. **BUG-5**: Silent DOM-state desync when registry lookup fails

### High Priority (UX / Accessibility)
4. ARIA tree roles (`role="tree"`, `role="treeitem"`, `aria-expanded`)
5. Keyboard block reordering (Alt+Up/Down or Ctrl+Shift+Up/Down)
6. Add `lostpointercapture` event handler to recover from zombie drags
7. Guard picker `.then` callback against stale `_pickerOpen` state

### Architecture
8. Centralize mutations through single `apply(operation)` gate
9. Transaction boundaries (snapshot before, rollback on failure)
10. Clone all block properties in `duplicateBlock`
