# Block Editor: State Management Refactor

## Problem

The block editor stores its entire state tree (all blocks, pointers, and field values) as a JSON string in the `data-state` HTML attribute. This attribute serves three roles:

1. **Server hydration** — templ renders the initial tree JSON into the attribute
2. **Change detection** — on save, admin.js reads `data-state` as the baseline to diff against
3. **Baseline update** — after save, admin.js writes the remapped state back to `data-state`

This means the full state is serialized in the DOM at all times, parsed on every save, and re-serialized after every save. For large layouts (200+ blocks with rich field values), this is unsustainable — the attribute grows to hundreds of KB, the `attributeChangedCallback` fires spuriously, and the DOM carries dead weight.

## Solution

Move the diff baseline from a DOM attribute to an in-memory JS property. Use `data-state` for initial hydration only (read once, then discard). Add crash recovery via debounced `sessionStorage` writes.

## Files Changed

| File | Change |
|------|--------|
| `block-editor-src/index.js` | Remove `data-state` from `observedAttributes`. Add `_baseline`, `commitSave(idMap)`, `getBaseline()`, crash recovery methods. Consume `data-state` once in `_initState()` then remove it. |
| `block-editor-src/history.js` | No changes (already has `remapIds`). |
| `admin.js` | Replace `editor.getAttribute('data-state')` with `editor.getBaseline()`. Replace `editor.setAttribute('data-state', ...)` with `editor.commitSave(idMap)`. Remove `remapEditorState()` (logic moves into component). |
| `content_edit.templ` | No changes (server still renders `data-state` for initial hydration). |

## Implementation Details

### 1. `_baseline` property (index.js)

A deep clone of `{ blocks, rootId }` taken at two points:
- End of `_initState()` after successful parse + validation
- Inside `commitSave()` after remap

```
_cloneBlocks(state) → deep clone of { blocks: {...}, rootId }
```

Uses the same clone logic as `History._cloneState()` — shallow copy each block, spread-copy each field array entry.

### 2. `commitSave(idMap)` method (index.js)

Called by admin.js after a successful save response. Executes in this exact order:

1. Remap `_state` block IDs and pointers (existing `remapEditorState` logic)
2. Remap `_elementRegistry` and `_wrapperRegistry` keys (existing `remapIds` logic)
3. Remap `_history` stacks via `this._history.remapIds(idMap)`
4. Remap `_state.selectedBlockId` if it was a client ID
5. Set `_baseline` to deep clone of current `_state`
6. Set `_state.dirty = false`
7. Update save button state
8. Clear crash recovery snapshot (successful save = no recovery needed)

When `idMap` is null or empty, steps 1-4 are skipped (save with no new blocks).

### 3. `getBaseline()` method (index.js)

Returns `_baseline` — the frozen state object that admin.js diffs against. This replaces `editor.getAttribute('data-state')` + `JSON.parse()`.

### 4. `_initState()` changes (index.js)

After successful parse + validation + state assignment:
- `this._baseline = this._cloneBlocks(this._state)`
- `this.removeAttribute('data-state')` — free the DOM
- Check for crash recovery snapshot and offer restore if found

### 5. `observedAttributes` change (index.js)

Remove `data-state` from the static `observedAttributes` array. Delete the `attributeChangedCallback` method entirely. The attribute is only read in `connectedCallback` → `_initState()`.

### 6. Crash recovery (index.js)

**Storage key:** `mcms-block-recovery:{contentId}` where `contentId` comes from `data-content-id` attribute.

**Write:** Debounced (30 seconds) after any mutation that sets `dirty = true`. Also writes on `visibilitychange` when `document.hidden && dirty`. Uses `this.serialize()` to get current state. Wrapped in try/catch — `QuotaExceededError` silently ignored.

**Read:** In `_initState()`, after hydrating from `data-state`, check if a recovery snapshot exists for this content ID. If found and its timestamp is newer than the page load, show a restore banner. User can restore or dismiss.

**Clear:** On successful `commitSave()`. On `beforeunload` when `!dirty`. On user dismiss of restore banner.

**Keep:** On `beforeunload` when `dirty` (user navigated away with unsaved changes — still worth offering recovery next time).

### 7. admin.js save handler changes

**Before (reads baseline from DOM):**
```js
var initialStr = editor.getAttribute('data-state');
var initial = {};
if (initialStr) {
    var parsed = JSON.parse(initialStr);
    initial = parsed.blocks || {};
}
```

**After (reads baseline from JS property):**
```js
var baseline = editor.getBaseline();
var initial = baseline ? baseline.blocks : {};
```

**Before (updates DOM after save):**
```js
remapEditorState(editor, state, resp.id_map);
stateStr = JSON.stringify(state);
editor.setAttribute('data-state', stateStr);
```

**After (calls component method):**
```js
editor.commitSave(resp.id_map || {});
```

The `remapEditorState()` function in admin.js is deleted — its logic is absorbed into `commitSave()`.

### 8. `isDirty()` — flag, not comparison

The existing `get dirty()` getter returns `this._state?.dirty ?? false`. This remains unchanged. The `dirty` flag is set to `true` by tree-ops and field edits, and cleared by `commitSave()`. No deep comparison ever runs for dirty detection.

## Order of Operations

1. Add `_cloneBlocks()` helper to index.js
2. Add `_baseline`, `getBaseline()`, `commitSave(idMap)` to index.js
3. Modify `_initState()` — set baseline, remove attribute, check recovery
4. Remove `observedAttributes` and `attributeChangedCallback`
5. Add crash recovery methods (debounced write, restore check, clear)
6. Modify `_onBeforeUnload` to handle recovery cleanup
7. Update admin.js save handler to use new API
8. Delete `remapEditorState()` from admin.js
9. Rebuild bundle with `just admin bundle`
