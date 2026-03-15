 Block Editor Undo/Redo Plan

 Context

 The block editor (<block-editor> web component) supports adding, removing, moving, indenting, outdenting, duplicating blocks and editing field values — but has no undo/redo. Users can't
 revert accidental deletes, moves, or field edits. This plan adds Ctrl+Z / Ctrl+Shift+Z undo/redo using a snapshot-based approach.

 Why snapshots, not command pattern: Operations like outdentBlock involve 7+ steps of pointer surgery across multiple blocks. Writing correct reverses for all operations is error-prone and
 fragile. The state is small enough (tens to low hundreds of blocks) that cloning it before each operation is cheap, and the full re-render path (_render()) already exists and works.

 Files to Modify

 ┌──────────────────────────────────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │                           File                           │                                                            Change                                                            │
 ├──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/static/js/block-editor-src/history.js     │ NEW — History class (undo/redo stack manager)                                                                                │
 ├──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/static/js/block-editor-src/index.js       │ Import History, wire up constructor, keyboard shortcuts, _undo/_redo/_restoreSnapshot/remapIds methods, field batch in       │
 │                                                          │ setFieldValue, clear in set state and _initState                                                                             │
 ├──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/static/js/block-editor-src/dom-patches.js │ Add this._history.pushUndo() before each mutation in 6 _do* methods, with discardLastUndo() on no-op                         │
 ├──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/static/js/block-editor-src/drag.js        │ Add snapshot before moveBlock() in _executeDrop                                                                              │
 ├──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/static/js/block-editor-src/picker.js      │ Add snapshot before addBlockFromDatatype() in _pickerInsertBlock                                                             │
 └──────────────────────────────────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 No server-side changes. No admin.js changes (it already calls editor.remapIds(idMap) after save).

 Implementation Steps

 Step 1: Create history.js

 Export: `export class History { ... }`

 New file exporting History class with:

 - constructor(maxSize=50) — Initialize undo/redo stacks (`this._undoStack = []`, `this._redoStack = []`), field batch timer ID (`this._fieldBatchTimer = null`), batch flag (`this._inFieldBatch = false`), max size
 - _cloneState(state) — Clone ONLY `{ blocks, rootId }` — does NOT include `selectedBlockId` or `dirty` (those are ephemeral view state stored at the entry level, not in the snapshot). Build a new blocks object: for each key in state.blocks, clone the block via `{ ...block, fields: block.fields ? block.fields.map(f => ({ ...f })) : undefined }`. Field values are always primitives (strings, numbers, booleans — JSON is stored as a string). Field objects must be shallow-copied because setFieldValue mutates `.value` in place — a bare `[...block.fields]` copies array references and would corrupt snapshots.
 - pushUndo(state) — Cancel any pending field batch timer (`clearTimeout(this._fieldBatchTimer); this._inFieldBatch = false;`), then create entry `{ snapshot: this._cloneState(state), selectedBlockId: state.selectedBlockId }`. Push to undo stack, clear redo stack, enforce max size (shift oldest if length > maxSize). Resetting the field batch ensures that field edits after a structural mutation always start a fresh batch.
 - discardLastUndo() — Pop last entry from undo stack when mutation turns out to be a no-op
 - pushFieldChange(state) — If `this._inFieldBatch` is true, return immediately (no-op). Otherwise: call `this.pushUndo(state)` to clone and push the snapshot, then set `this._inFieldBatch = true` and start `this._fieldBatchTimer = setTimeout(() => { this._inFieldBatch = false; }, 500)`. Reuses `pushUndo` internally to avoid duplicating push/clear/enforce-max logic.
 - inFieldBatch getter — Returns `this._inFieldBatch`
 - popUndo(currentState) — Clone currentState and push to redo stack, pop and return last undo entry
 - popRedo(currentState) — Clone currentState and push to undo stack, pop and return last redo entry
 - canUndo/canRedo getters — Return `this._undoStack.length > 0` / `this._redoStack.length > 0`
 - clear() — Reset both stacks to `[]`, cancel pending field batch timer (`clearTimeout(this._fieldBatchTimer)`), reset `this._inFieldBatch = false`
 - remapIds(idMap) — Iterate over `Object.keys(idMap)`. For each clientId, check each entry in both stacks: if `snapshot.blocks[clientId]` exists, delete old key and set `snapshot.blocks[idMap[clientId]] = block`, update `block.id = idMap[clientId]`. For each block in the snapshot (all blocks, not just remapped ones), remap pointer fields (parentId, firstChildId, nextSiblingId, prevSiblingId) through idMap if the pointer value exists in idMap. Remap snapshot.rootId and entry.selectedBlockId through idMap if they exist in idMap.

 Each stack entry: { snapshot: { blocks, rootId }, selectedBlockId }.

 Step 2: Wire up index.js

 Import: `import { History } from './history.js';` (add after existing imports at top of file)

 Constructor (after `this._rootDatatypeId = null;`, before closing brace): `this._history = new History(50);`

 `set state()` setter (after `this._state = newState;`): Add `if (this._history) this._history.clear();`

 `_initState` has two state-assignment paths. Add `if (this._history) this._history.clear();` in BOTH:
  1. After `this._state = createState();` in the early-return path (empty/null `data-state` attribute, line ~175)
  2. After `this._state = newState;` in the parsed-state path (line ~215)

 `disconnectedCallback` (after the existing removeEventListener calls): Add `if (this._history) this._history.clear();` to cancel any pending field batch timer and release snapshot memory on unmount.

 `setFieldValue` method (before `field.value = value;`): Add `if (!this._history.inFieldBatch) this._history.pushFieldChange(this._state);`

 Keyboard shortcuts in `_onKeyDown` (insert after `if (!this._state) return;`, before `var blockId = this._state.selectedBlockId;`). Note: the picker guard (`if (this._pickerOpen)`) is above these lines, so Ctrl+Z while the picker is open is consumed by `_onPickerKeyDown` — this is correct, undo does not fire during picker interaction:

 ```
 if ((e.ctrlKey || e.metaKey) && !e.shiftKey && e.key === 'z') {
     e.preventDefault();
     if (!this._drag) this._undo();
     return;
 }
 if ((e.ctrlKey || e.metaKey) && ((e.shiftKey && (e.key === 'z' || e.key === 'Z')) || e.key === 'y')) {
     e.preventDefault();
     if (!this._drag) this._redo();
     return;
 }
 ```

 Both undo and redo call e.preventDefault() to suppress browser-native undo/redo.

 New methods on class:

 - _undo() — Guard canUndo, call popUndo(this._state), call _restoreSnapshot(entry)
 - _redo() — Guard canRedo, call popRedo(this._state), call _restoreSnapshot(entry)
 - _restoreSnapshot(entry) — In this order:
   1. `this._state.blocks = entry.snapshot.blocks`
   2. `this._state.rootId = entry.snapshot.rootId`
   3. `this._state.selectedBlockId = null`
   4. `this._state.dirty = true` (ensures save button reflects unsaved state; `_render()` and `_selectBlock` do not modify `dirty`)
   5. `this._render()` (full re-render, which clears and rebuilds registries)
   6. If `entry.selectedBlockId` exists in `this._state.blocks`, call `this._selectBlock(entry.selectedBlockId)`
   7. Dispatch `block-editor:change` with `{ action: 'undo-redo' }`
 - remapIds(idMap) — In this order:
   1. For each clientId in Object.keys(idMap): if `this._elementRegistry` has clientId, get the element, delete old key, set new key `this._elementRegistry.set(idMap[clientId], el)`, update `el.dataset.blockId = idMap[clientId]`
   2. Same for `this._wrapperRegistry`: get wrapper, delete old key, set new key, update `wrapper.dataset.blockId = idMap[clientId]`
   3. Call `this._history.remapIds(idMap)`
   4. If `this._state.selectedBlockId` exists in idMap, remap it

   No `this._history` null guard needed — `History` is always constructed before `remapIds` can be called by `admin.js` (which guards with `typeof editor.remapIds === 'function'`).

   Note: this is the FIRST implementation of `remapIds` on BlockEditor. Previously `admin.js` skipped the call because the `typeof` guard returned false. After this change, `admin.js` will start calling it — this is intentional and required for history to survive saves.

   The DOM attribute update is required because remapEditorState in admin.js only mutates the JS state — without updating data-block-id on the actual elements, click handlers would read stale IDs.

 Step 3: Add snapshots in dom-patches.js

 For each of the 6 _do* methods listed below, add `this._history.pushUndo(this._state)` after ALL guard checks (including block-existence guards AND operation-specific guards like `!prevSiblingId`) but before the tree-ops mutation call. For methods where the mutation can return false (indent, outdent, moveUp, moveDown), follow with `if (!result) { this._history.discardLastUndo(); return; }`.

 Placement per method:
 - _doIndentBlock: after `if (!block) return;` guard, before `indentBlock()` call. Needs discardLastUndo on `!result`.
 - _doOutdentBlock: after `if (!block) return;` guard, before `outdentBlock()` call. Needs discardLastUndo on `!result`.
 - _doDuplicateBlock: after `if (!block) return;` guard, before `duplicateBlock()` call. No discardLastUndo needed — `duplicateBlock` cannot fail when the block existence guard has passed (the `if (!cloneId) return` path is unreachable).
 - _doMoveBlockUp: after `if (!prevSiblingId) return;` guard, before `moveBlockUp()` call. Needs discardLastUndo on `!result`.
 - _doMoveBlockDown: after `if (!nextSiblingId) return;` guard, before `moveBlockDown()` call. Needs discardLastUndo on `!result`.
 - _doAddBlockAfter: after guard checks, before `addBlock()` call. No discardLastUndo needed — `addBlock` always succeeds.

 Note: _doDeleteBlock is NOT included here — the delete snapshot is handled in Step 5 via _executeDeleteBlock (after the confirmation dialog).

 Step 4: Add snapshots in drag.js and picker.js

 - drag.js _executeDrop: `this._history.pushUndo(this._state)` after the `if (!this._drag || !this._state) return;` guard and the `oldParentId` read, but before the `moveBlock()` call
 - picker.js _pickerInsertBlock: `this._history.pushUndo(this._state)` after the guard checks and `var datatype = ...` preparation, but before the `addBlockFromDatatype()` call

 Step 5: Add snapshots in index.js mutation sites

 - _doAddBlock (before `addBlock()` call): `this._history.pushUndo(this._state)` — no discardLastUndo needed, addBlock always succeeds
 - _executeDeleteBlock (before `removeBlock()` call, after the `if (!block) return` guard): `this._history.pushUndo(this._state)` — no discardLastUndo needed, removeBlock always succeeds

 Step 6: Unit test History class

 Create `internal/admin/static/js/block-editor-src/history.test.js`. Use vitest. Import directly: `import { History } from './history.js';`

 Test cases:
 - pushUndo/popUndo symmetry (push N, pop N, get them back in reverse order)
 - pushUndo/popRedo symmetry
 - Max size eviction (push 51, verify only 50 remain, oldest is gone)
 - clear() resets both stacks and canUndo/canRedo return false
 - discardLastUndo() removes the most recent entry
 - remapIds() rewrites block map keys, block.id, pointer fields, rootId, and selectedBlockId
 - Field batching: pushFieldChange within 500ms window does not create additional snapshots; after 500ms (use vi.advanceTimersByTime) a new call does
- Mutation isolation: push a snapshot, then mutate a block's field value and add a new block to the original state object. Pop the snapshot and verify the field has its original value and the new block does not exist in the restored snapshot

 Step 7: Bundle and test

 Run `just admin bundle` to rebuild block-editor.js, then run `just admin test` to execute the vitest tests including the new history.test.js.

 Edge Cases Handled

 - Undo after save/ID-remap: remapIds() rewrites all snapshots; history preserved across saves
 - Field batching: 500ms window groups rapid keystrokes into one undo step
- Structural mutation during field batch: pushUndo resets the field batch timer, so a field edit after a structural mutation (e.g., delete) always creates a new undo entry rather than being silently absorbed into the prior batch
- Collapse/expand state: intentionally excluded from snapshots — collapse is view state, not content state. Undo restores content structure only.
 - No-op mutations: discardLastUndo() prevents empty undo entries when indent/outdent/moveUp/moveDown fail
 - Delete confirmation dialog: Snapshot pushed in _executeDeleteBlock (after confirmation), not _doDeleteBlock
 - Undo during drag: Guarded — Ctrl+Z ignored while this._drag is active
 - Ctrl+Z in field inputs: Field panel inputs are outside the editor container, so browser native undo handles them (no conflict)
 - Block no longer exists after undo: Selection cleared gracefully if restored block ID missing
 - HTMX navigation: _initState clears history on reinit
 - Component unmount: disconnectedCallback calls clear() to cancel pending field batch timer and free snapshot memory
 - Always-succeed mutations: _doAddBlock, _doAddBlockAfter, _doDuplicateBlock, and _executeDeleteBlock do not need discardLastUndo fallback
- remapIds complexity: O(snapshots * blocks) for pointer remapping, acceptable given the 50-snapshot cap and typical block counts under 200

 Verification

 1. just admin bundle — Rebuild JS bundle
 2. Start dev server, navigate to content edit page with block editor
 3. Test each operation + undo + redo:
   - Add block → Ctrl+Z (block removed) → Ctrl+Shift+Z (block returns)
   - Delete block → Ctrl+Z (block restored) → Ctrl+Shift+Z (deleted again)
   - Indent/outdent/move up/move down → undo each
   - Drag block to new position → undo
   - Insert from picker → undo
   - Edit field value, pause, edit again → undo steps through batches
 4. Save → undo past save point → verify IDs correct
 5. Undo 50+ times → verify oldest entries evicted
