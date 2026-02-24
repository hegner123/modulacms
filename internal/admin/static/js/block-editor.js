// block-editor.js — <block-editor> web component
// Phase 6: Hover toolbar, duplicate, moveUp/Down + keyboard shortcuts

// ============================================================
// Block Type Configuration
// ============================================================

export const BLOCK_TYPE_CONFIG = {
        text: { label: 'Text', canHaveChildren: false },
        heading: { label: 'Heading', canHaveChildren: false },
        image: { label: 'Image', canHaveChildren: false },
        container: { label: 'Container', canHaveChildren: true },
};

export const MAX_DEPTH = 8;

// ============================================================
// ID Generation
// ============================================================

export function generateId() {
        return crypto.randomUUID();
}

// ============================================================
// State Creation
// ============================================================

export function createState() {
        return {
                blocks: {},
                rootId: null,
                selectedBlockId: null,
                dirty: false,
        };
}

// ============================================================
// Tree Operations — mutate state in place
// ============================================================

export function unlink(state, blockId) {
        const block = state.blocks[blockId];
        if (!block) return;

        if (block.prevSiblingId) {
                state.blocks[block.prevSiblingId].nextSiblingId = block.nextSiblingId;
        }
        if (block.nextSiblingId) {
                state.blocks[block.nextSiblingId].prevSiblingId = block.prevSiblingId;
        }

        // If block is firstChild of parent, update parent
        if (block.parentId) {
                const parent = state.blocks[block.parentId];
                if (parent && parent.firstChildId === blockId) {
                        parent.firstChildId = block.nextSiblingId;
                }
        }

        // If block is head of root list, update rootId
        if (state.rootId === blockId) {
                state.rootId = block.nextSiblingId;
        }

        block.parentId = null;
        block.prevSiblingId = null;
        block.nextSiblingId = null;
}

export function insertBefore(state, blockId, targetId) {
        const block = state.blocks[blockId];
        const target = state.blocks[targetId];

        block.parentId = target.parentId;
        block.nextSiblingId = target.id;
        block.prevSiblingId = target.prevSiblingId;

        if (target.prevSiblingId) {
                state.blocks[target.prevSiblingId].nextSiblingId = block.id;
        }
        target.prevSiblingId = block.id;

        // If target was firstChild of parent, block becomes firstChild
        if (target.parentId) {
                const parent = state.blocks[target.parentId];
                if (parent && parent.firstChildId === targetId) {
                        parent.firstChildId = block.id;
                }
        }

        // If target was root head, block becomes root head
        if (state.rootId === targetId) {
                state.rootId = block.id;
        }
}

export function insertAfter(state, blockId, targetId) {
        const block = state.blocks[blockId];
        const target = state.blocks[targetId];

        block.parentId = target.parentId;
        block.prevSiblingId = target.id;
        block.nextSiblingId = target.nextSiblingId;

        if (target.nextSiblingId) {
                state.blocks[target.nextSiblingId].prevSiblingId = block.id;
        }
        target.nextSiblingId = block.id;
}

export function insertAsFirstChild(state, blockId, parentId) {
        const block = state.blocks[blockId];
        const parent = state.blocks[parentId];

        block.parentId = parent.id;
        block.prevSiblingId = null;
        block.nextSiblingId = parent.firstChildId;

        if (parent.firstChildId) {
                state.blocks[parent.firstChildId].prevSiblingId = block.id;
        }
        parent.firstChildId = block.id;
}

export function insertAsLastChild(state, blockId, parentId) {
        const parent = state.blocks[parentId];
        if (!parent.firstChildId) {
                insertAsFirstChild(state, blockId, parentId);
                return;
        }
        const lastId = findLastSibling(state, parent.firstChildId);
        insertAfter(state, blockId, lastId);
}

export function addBlock(state, type, afterId) {
        const id = generateId();
        const config = BLOCK_TYPE_CONFIG[type];
        const block = {
                id,
                type,
                parentId: null,
                firstChildId: null,
                prevSiblingId: null,
                nextSiblingId: null,
                label: config.label + ' Block',
        };
        state.blocks[id] = block;

        if (!state.rootId) {
                // Empty state — this becomes the root
                state.rootId = id;
        } else if (afterId) {
                insertAfter(state, id, afterId);
        } else {
                // Append to end of root list
                const lastId = findLastSibling(state, state.rootId);
                insertAfter(state, id, lastId);
        }

        state.dirty = true;
        return id;
}

export function removeBlock(state, blockId) {
        const block = state.blocks[blockId];
        if (!block) return [];

        // 1. Collect all descendant IDs via DFS (while pointers are intact)
        const removed = collectDescendants(state, blockId);
        removed.push(blockId);

        // 2. Unlink block from its sibling chain
        unlink(state, blockId);

        // 3. Delete block + all descendants from state.blocks
        for (const id of removed) {
                delete state.blocks[id];
        }

        state.dirty = true;
        return removed;
}

export function moveBlock(state, blockId, targetId, position) {
        if (blockId === targetId) return;
        const block = state.blocks[blockId];
        const target = state.blocks[targetId];
        if (!block || !target) return;

        unlink(state, blockId);

        if (position === 'before') {
                insertBefore(state, blockId, targetId);
        } else if (position === 'after') {
                insertAfter(state, blockId, targetId);
        } else if (position === 'inside') {
                insertAsFirstChild(state, blockId, targetId);
        }

        state.dirty = true;
}

/**
 * Indent a block: make it the last child of its previous sibling.
 * Returns true if the operation was performed, false if it was a no-op.
 */
export function indentBlock(state, blockId) {
        const block = state.blocks[blockId];
        if (!block) return false;

        // Must have a previous sibling to indent into
        if (!block.prevSiblingId) return false;

        const prevSibling = state.blocks[block.prevSiblingId];
        if (!prevSibling) return false;

        // Check max depth constraint (depths 0 through MAX_DEPTH-1 are valid; MAX_DEPTH levels total)
        if (getDepth(state, blockId) + 1 >= MAX_DEPTH) return false;

        // Check that the previous sibling's type allows children
        const config = BLOCK_TYPE_CONFIG[prevSibling.type];
        if (!config || !config.canHaveChildren) return false;

        unlink(state, blockId);
        insertAsLastChild(state, blockId, prevSibling.id);
        state.dirty = true;
        return true;
}

/**
 * Outdent a block: move it out of its parent to become the next sibling
 * of that parent. Younger siblings of the block are reparented as children
 * of the block.
 * Returns true if the operation was performed, false if it was a no-op.
 */
export function outdentBlock(state, blockId) {
        const block = state.blocks[blockId];
        if (!block) return false;

        // Must have a parent to outdent from
        if (!block.parentId) return false;

        const parent = state.blocks[block.parentId];
        if (!parent) return false;
        const parentId = parent.id;

        // Step 2: Collect younger siblings (block.nextSiblingId chain)
        const youngerSiblings = [];
        let walkId = block.nextSiblingId;
        while (walkId) {
                youngerSiblings.push(walkId);
                walkId = state.blocks[walkId].nextSiblingId;
        }

        // Step 3: Detach younger siblings from parent
        if (block.prevSiblingId) {
                // Older siblings stay with parent; cut the chain after the older sibling
                state.blocks[block.prevSiblingId].nextSiblingId = null;
        } else {
                // Block is firstChild of parent. All children are leaving.
                parent.firstChildId = null;
        }

        // Step 4: Sever block's link to younger siblings before unlinking.
        // This prevents unlink's firstChild logic from re-pointing parent.firstChildId
        // to the now-detached younger siblings.
        if (youngerSiblings.length > 0) {
                block.nextSiblingId = null;
                state.blocks[youngerSiblings[0]].prevSiblingId = null;
        }

        // Step 5: Unlink block from parent's child list
        unlink(state, blockId);

        // Step 6: Reparent younger siblings under block
        if (youngerSiblings.length > 0) {
                // Set parentId on all younger siblings
                for (const sibId of youngerSiblings) {
                        state.blocks[sibId].parentId = blockId;
                }

                if (block.firstChildId) {
                        // Block already has children; append younger siblings after last existing child
                        const lastChildId = findLastSibling(state, block.firstChildId);
                        state.blocks[lastChildId].nextSiblingId = youngerSiblings[0];
                        state.blocks[youngerSiblings[0]].prevSiblingId = lastChildId;
                } else {
                        // Block has no existing children
                        block.firstChildId = youngerSiblings[0];
                        // prevSiblingId of first younger sibling is already null from step 4
                }
        }

        // Step 7: Insert block after its former parent
        insertAfter(state, blockId, parentId);

        state.dirty = true;
        return true;
}

/**
 * Duplicate a block and its entire subtree with new IDs.
 * The clone is inserted after the original in the sibling chain.
 * Returns the new root block's ID, or null if the block does not exist.
 */
export function duplicateBlock(state, blockId) {
        const block = state.blocks[blockId];
        if (!block) return null;

        // Deep-clone the block and its subtree with new IDs.
        // idMap tracks original -> clone ID for rewiring pointers.
        const idMap = new Map();

        function cloneSubtree(originalId) {
                const original = state.blocks[originalId];
                if (!original) return null;

                const newId = generateId();
                idMap.set(originalId, newId);

                const clone = {
                        id: newId,
                        type: original.type,
                        parentId: null,
                        firstChildId: null,
                        prevSiblingId: null,
                        nextSiblingId: null,
                        label: original.label,
                };
                state.blocks[newId] = clone;

                // Clone children recursively via sibling chain
                let childId = original.firstChildId;
                let prevCloneChildId = null;
                while (childId) {
                        const cloneChildId = cloneSubtree(childId);
                        state.blocks[cloneChildId].parentId = newId;

                        if (!prevCloneChildId) {
                                clone.firstChildId = cloneChildId;
                        } else {
                                state.blocks[prevCloneChildId].nextSiblingId = cloneChildId;
                                state.blocks[cloneChildId].prevSiblingId = prevCloneChildId;
                        }

                        prevCloneChildId = cloneChildId;
                        childId = state.blocks[childId].nextSiblingId;
                }

                return newId;
        }

        const cloneId = cloneSubtree(blockId);

        // Insert clone after original in sibling chain
        insertAfter(state, cloneId, blockId);

        state.dirty = true;
        return cloneId;
}

/**
 * Move a block before its previous sibling. No-op if no previous sibling.
 * IMPORTANT: Captures prevSiblingId BEFORE calling unlink (unlink clears pointers).
 * Returns true if the operation was performed, false if it was a no-op.
 */
export function moveBlockUp(state, blockId) {
        const block = state.blocks[blockId];
        if (!block) return false;

        // Capture before unlink clears the pointer
        const prevSiblingId = block.prevSiblingId;
        if (!prevSiblingId) return false;

        unlink(state, blockId);
        insertBefore(state, blockId, prevSiblingId);

        state.dirty = true;
        return true;
}

/**
 * Move a block after its next sibling. No-op if no next sibling.
 * IMPORTANT: Captures nextSiblingId BEFORE calling unlink (unlink clears pointers).
 * Returns true if the operation was performed, false if it was a no-op.
 */
export function moveBlockDown(state, blockId) {
        const block = state.blocks[blockId];
        if (!block) return false;

        // Capture before unlink clears the pointer
        const nextSiblingId = block.nextSiblingId;
        if (!nextSiblingId) return false;

        unlink(state, blockId);
        insertAfter(state, blockId, nextSiblingId);

        state.dirty = true;
        return true;
}

// ============================================================
// Tree Queries
// ============================================================

export function getChildren(state, parentId) {
        const parent = state.blocks[parentId];
        if (!parent || !parent.firstChildId) return [];
        const children = [];
        let currentId = parent.firstChildId;
        while (currentId) {
                const child = state.blocks[currentId];
                children.push(child);
                currentId = child.nextSiblingId;
        }
        return children;
}

export function getRootList(state) {
        if (!state.rootId) return [];
        const roots = [];
        let currentId = state.rootId;
        while (currentId) {
                const block = state.blocks[currentId];
                roots.push(block);
                currentId = block.nextSiblingId;
        }
        return roots;
}

export function isDescendant(state, candidateId, ancestorId) {
        let currentId = state.blocks[candidateId]?.parentId;
        while (currentId) {
                if (currentId === ancestorId) return true;
                currentId = state.blocks[currentId]?.parentId;
        }
        return false;
}

export function getDepth(state, blockId) {
        let depth = 0;
        let currentId = state.blocks[blockId]?.parentId;
        while (currentId) {
                depth++;
                currentId = state.blocks[currentId]?.parentId;
        }
        return depth;
}

export function findLastSibling(state, blockId) {
        let currentId = blockId;
        while (state.blocks[currentId]?.nextSiblingId) {
                currentId = state.blocks[currentId].nextSiblingId;
        }
        return currentId;
}

/**
 * Depth-first traversal of all blocks. Walks from rootId, visiting each
 * block's firstChildId before its nextSiblingId.
 * Returns an array of block IDs in traversal order.
 */
export function getBlockTraversalOrder(state) {
        const order = [];
        if (!state.rootId) return order;

        const stack = [state.rootId];
        while (stack.length > 0) {
                const id = stack.pop();
                const block = state.blocks[id];
                if (!block) continue;
                order.push(id);
                // Push nextSibling first so firstChild is processed first (LIFO)
                if (block.nextSiblingId) stack.push(block.nextSiblingId);
                if (block.firstChildId) stack.push(block.firstChildId);
        }
        return order;
}

function collectDescendants(state, blockId) {
        const result = [];
        const block = state.blocks[blockId];
        if (!block || !block.firstChildId) return result;

        const stack = [block.firstChildId];
        while (stack.length > 0) {
                const id = stack.pop();
                const child = state.blocks[id];
                if (!child) continue;
                result.push(id);
                if (child.nextSiblingId) stack.push(child.nextSiblingId);
                if (child.firstChildId) stack.push(child.firstChildId);
        }
        return result;
}

export function getDescendantCount(state, blockId) {
        return collectDescendants(state, blockId).length;
}

// ============================================================
// State Validation (dev-mode only)
// ============================================================

export function validateState(state) {
        const errors = [];
        const reachable = new Set();

        // Check reciprocal sibling pointers
        for (const [id, block] of Object.entries(state.blocks)) {
                if (block.nextSiblingId) {
                        const next = state.blocks[block.nextSiblingId];
                        if (!next) {
                                errors.push(`Block ${id}: nextSiblingId "${block.nextSiblingId}" not found in blocks`);
                        } else if (next.prevSiblingId !== id) {
                                errors.push(`Block ${id}: nextSiblingId "${block.nextSiblingId}" has prevSiblingId "${next.prevSiblingId}", expected "${id}"`);
                        }
                }
                if (block.prevSiblingId) {
                        const prev = state.blocks[block.prevSiblingId];
                        if (!prev) {
                                errors.push(`Block ${id}: prevSiblingId "${block.prevSiblingId}" not found in blocks`);
                        } else if (prev.nextSiblingId !== id) {
                                errors.push(`Block ${id}: prevSiblingId "${block.prevSiblingId}" has nextSiblingId "${prev.nextSiblingId}", expected "${id}"`);
                        }
                }
        }

        // Check parent-firstChild consistency
        for (const [id, block] of Object.entries(state.blocks)) {
                if (block.firstChildId) {
                        const firstChild = state.blocks[block.firstChildId];
                        if (!firstChild) {
                                errors.push(`Block ${id}: firstChildId "${block.firstChildId}" not found in blocks`);
                        } else if (firstChild.parentId !== id) {
                                errors.push(`Block ${id}: firstChildId "${block.firstChildId}" has parentId "${firstChild.parentId}", expected "${id}"`);
                        }
                }
        }

        // Check root chain has parentId = null, and track reachability
        if (state.rootId) {
                const visited = new Set();
                let currentId = state.rootId;
                while (currentId) {
                        if (visited.has(currentId)) {
                                errors.push(`Cycle detected in root chain at block "${currentId}"`);
                                break;
                        }
                        visited.add(currentId);
                        const block = state.blocks[currentId];
                        if (!block) {
                                errors.push(`Root chain references non-existent block "${currentId}"`);
                                break;
                        }
                        if (block.parentId !== null) {
                                errors.push(`Root chain block "${currentId}" has parentId "${block.parentId}", expected null`);
                        }
                        reachable.add(currentId);
                        // Mark children as reachable (recursively)
                        markChildrenReachable(state, currentId, reachable, errors);
                        currentId = block.nextSiblingId;
                }
        }

        // Check all blocks are reachable
        for (const id of Object.keys(state.blocks)) {
                if (!reachable.has(id)) {
                        errors.push(`Block "${id}" is not reachable from rootId or any firstChildId chain`);
                }
        }

        // Check parent consistency for children
        for (const [id, block] of Object.entries(state.blocks)) {
                if (block.parentId) {
                        const parent = state.blocks[block.parentId];
                        if (!parent) {
                                errors.push(`Block "${id}" has parentId "${block.parentId}" which does not exist`);
                        }
                }
        }

        // Check sibling chains for cycles
        for (const [id, block] of Object.entries(state.blocks)) {
                if (block.firstChildId) {
                        const visited = new Set();
                        let childId = block.firstChildId;
                        while (childId) {
                                if (visited.has(childId)) {
                                        errors.push(`Cycle in child chain of block "${id}" at "${childId}"`);
                                        break;
                                }
                                visited.add(childId);
                                const child = state.blocks[childId];
                                if (!child) break;
                                if (child.parentId !== id) {
                                        errors.push(`Child "${childId}" of block "${id}" has parentId "${child.parentId}", expected "${id}"`);
                                }
                                childId = child.nextSiblingId;
                        }
                }
        }

        return errors;
}

function markChildrenReachable(state, parentId, reachable, errors) {
        const parent = state.blocks[parentId];
        if (!parent || !parent.firstChildId) return;

        const visited = new Set();
        let childId = parent.firstChildId;
        while (childId) {
                if (visited.has(childId)) break; // cycle, already reported
                visited.add(childId);
                reachable.add(childId);
                const child = state.blocks[childId];
                if (!child) break;
                markChildrenReachable(state, childId, reachable, errors);
                childId = child.nextSiblingId;
        }
}

// ============================================================
// CSS (inlined for adoptedStyleSheets)
// ============================================================

const EDITOR_CSS = `
:host {
  display: block;
  --font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-family: var(--font-family);
  --bg: #1e1e2e;
  --border: #383850;
  --border-hover: #4a4a6a;
  --text: #cdd6f4;
  --text-muted: #6c7086;
  --accent: #89b4fa;
  --danger: #f38ba8;
  --danger-hover: #e06688;
  --toolbar-bg: #181825;
  --block-bg: #1e1e2e;
  --error-bg: #302028;
  --error-border: #6e3040;
  --error-text: #f38ba8;
  --btn-hover-bg: #252536;
  --save-active-text: #1e1e2e;
  --save-hover-bg: #7aa8ec;
  --save-disabled-bg: #3a4a6a;
  --save-disabled-text: #6c7086;
  --badge-bg: #1e2a45;
  --badge-text: #89b4fa;
  --sp-1: 0.25rem;
  --sp-2: 0.5rem;
  --sp-3: 0.75rem;
  --sp-4: 1rem;
  --sp-6: 1.5rem;
  --sp-8: 2rem;
  --radius-sm: 3px;
  --radius-md: 4px;
  --radius-lg: 6px;
  --radius-xl: 8px;
  --fs-xs: 0.6875rem;
  --fs-sm: 0.75rem;
  --fs-md: 0.8125rem;
  --fs-base: 0.875rem;
}

.editor-container {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  overflow: hidden;
  outline: none;
}

.editor-container:focus-visible {
  border-color: var(--accent);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-2) var(--sp-3);
  background: var(--toolbar-bg);
  border-bottom: 1px solid var(--border);
}

.toolbar button {
  padding: var(--sp-2) var(--sp-3);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg);
  color: var(--text);
  font-size: var(--fs-md);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.toolbar button:hover {
  border-color: var(--border-hover);
  background: var(--btn-hover-bg);
}

.toolbar button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.toolbar button.save-btn {
  margin-inline-start: auto;
  background: var(--accent);
  color: var(--save-active-text);
  border-color: var(--accent);
}

.toolbar button.save-btn:hover:not(:disabled) {
  background: var(--save-hover-bg);
}

.toolbar button.save-btn:disabled {
  background: var(--save-disabled-bg);
  border-color: var(--save-disabled-bg);
  color: var(--save-disabled-text);
}

.block-list {
  padding: var(--sp-3);
  min-height: 4rem;
}

.block-list:empty::after {
  content: "No blocks. Use the toolbar to add a block.";
  display: block;
  padding: var(--sp-8);
  text-align: center;
  color: var(--text-muted);
  font-size: var(--fs-base);
}

.block-item {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-3) var(--sp-3);
  margin-bottom: var(--sp-2);
  background: var(--block-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  transition: border-color 0.15s;
}

.block-item:hover {
  border-color: var(--border-hover);
}

.block-item.selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 1px var(--accent);
}

.block-type-badge {
  display: inline-block;
  padding: var(--sp-1) var(--sp-2);
  border-radius: var(--radius-sm);
  font-size: var(--fs-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  background: var(--badge-bg);
  color: var(--badge-text);
  flex-shrink: 0;
}

.block-label {
  flex: 1;
  font-size: var(--fs-base);
  color: var(--text);
}

.block-delete-btn {
  padding: var(--sp-1) var(--sp-2);
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  font-size: var(--fs-sm);
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
  flex-shrink: 0;
}

.block-delete-btn:hover {
  color: var(--danger);
  background: var(--error-bg);
}

.error-container {
  padding: var(--sp-6);
  margin: var(--sp-3);
  background: var(--error-bg);
  border: 1px solid var(--error-border);
  border-radius: var(--radius-lg);
}

.error-message {
  color: var(--error-text);
  font-size: var(--fs-base);
  font-weight: 500;
  margin-bottom: var(--sp-3);
}

.error-detail {
  color: var(--error-text);
  font-size: var(--fs-sm);
  opacity: 0.7;
  margin-bottom: var(--sp-3);
  white-space: pre-wrap;
  font-family: monospace;
}

.error-container button {
  padding: var(--sp-2) var(--sp-3);
  border: 1px solid var(--error-border);
  border-radius: var(--radius-md);
  background: var(--bg);
  color: var(--error-text);
  font-size: var(--fs-md);
  cursor: pointer;
}

.children-container {
  padding-inline-start: 0;
  padding-top: var(--sp-1);
}

.block-wrapper {
  margin-bottom: var(--sp-2);
}

/* ---- Phase 3: Drag and Drop with Nesting ---- */

.block-item {
  cursor: grab;
  user-select: none;
}

.block-item.dragging {
  opacity: 0.3;
}

.block-item.drop-inside {
  border-color: var(--accent);
  box-shadow: 0 0 0 1px var(--accent), inset 0 0 0 1px var(--accent);
}

.drag-overlay {
  position: fixed;
  pointer-events: none;
  opacity: 0.7;
  z-index: 9999;
  will-change: transform;
}

.drop-indicator {
  position: absolute;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--accent);
  pointer-events: none;
  z-index: 100;
  border-radius: 1px;
}

.block-list {
  position: relative;
}

.child-count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.25rem;
  height: 1.25rem;
  padding: 0 var(--sp-1);
  border-radius: var(--radius-sm);
  font-size: var(--fs-xs);
  font-weight: 600;
  background: var(--badge-bg);
  color: var(--badge-text);
  flex-shrink: 0;
}

/* ---- Phase 5: Block Type Rendering ---- */

.block-item {
  flex-wrap: wrap;
}

.block-type-content {
  width: 100%;
  margin-top: var(--sp-2);
  order: 99;
}

/* Text block: paragraph placeholder lines */
.block-type-content--text {
  display: flex;
  flex-direction: column;
  gap: var(--sp-1);
}

.block-type-content--text .text-line {
  height: 0.5rem;
  background: var(--border);
  border-radius: 2px;
  opacity: 0.4;
}

.block-type-content--text .text-line:last-child {
  width: 60%;
}

/* Heading block: large bold label */
.block-type-content--heading {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.01em;
}

/* Image block: dashed rect placeholder */
.block-type-content--image {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 5rem;
  border: 2px dashed var(--border);
  border-radius: var(--radius-md);
  color: var(--text-muted);
  font-size: var(--fs-sm);
}

/* Container block: labeled box with visible children area */
.block-item--container {
  border-color: var(--accent);
  border-style: dashed;
  background: color-mix(in srgb, var(--accent) 4%, var(--block-bg));
}

.block-type-content--container {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Type-specific badge colors */
.block-type-badge--heading {
  background: #2a1e45;
  color: #cba6f7;
}

.block-type-badge--image {
  background: #1e3a2a;
  color: #a6e3a1;
}

.block-type-badge--container {
  background: #1e2a45;
  color: #89b4fa;
}

.block-type-badge--text {
  background: #2a2a1e;
  color: #f9e2af;
}

/* ---- Phase 6: Hover Toolbar ---- */

.block-item {
  position: relative;
}

.block-hover-toolbar {
  display: none;
  position: absolute;
  top: var(--sp-1);
  right: var(--sp-1);
  z-index: 50;
  gap: 2px;
  padding: 2px;
  background: var(--toolbar-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.block-item:hover > .block-hover-toolbar,
.block-hover-toolbar:hover {
  display: flex;
}

.block-item.dragging > .block-hover-toolbar {
  display: none;
}

.block-hover-toolbar button {
  padding: 2px var(--sp-2);
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.15s, background 0.15s;
}

.block-hover-toolbar button:hover {
  color: var(--text);
  background: var(--btn-hover-bg);
}

.block-hover-toolbar button[data-action="toolbar-delete"]:hover {
  color: var(--danger);
  background: var(--error-bg);
}
`;

// ============================================================
// Web Component
// ============================================================

// Guard browser-only code for testability in Node/vitest
const isBrowser = typeof window !== 'undefined' && typeof CSSStyleSheet !== 'undefined';

if (isBrowser) {
        const sheet = new CSSStyleSheet();
        sheet.replaceSync(EDITOR_CSS);

        class BlockEditor extends HTMLElement {
                static get observedAttributes() {
                        return ['data-state'];
                }

                constructor() {
                        super();
                        this.attachShadow({ mode: 'open' });
                        this.shadowRoot.adoptedStyleSheets = [sheet];

                        this._state = null;
                        this._elementRegistry = new Map(); // blockId -> block-item element
                        this._wrapperRegistry = new Map(); // blockId -> block-wrapper element
                        this._beforeUnloadHandler = this._onBeforeUnload.bind(this);

                        // Phase 3: Drag state
                        this._drag = null; // active drag session object, or null
                        this._dropIndicator = null;
                        this._escapeHandler = this._onEscapeKey.bind(this);
                        this._autoScrollRaf = null; // requestAnimationFrame ID for auto-scroll
                        this._lastPointerY = 0; // track pointer Y for auto-scroll

                        // Phase 4: Keyboard handler
                        this._keydownHandler = this._onKeyDown.bind(this);
                }

                get dev() {
                        return this.hasAttribute('data-dev');
                }

                get state() {
                        return this._state;
                }

                set state(newState) {
                        this._state = newState;
                        this._state.dirty = false;
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();
                        this._render();
                }

                get dirty() {
                        return this._state ? this._state.dirty : false;
                }

                serialize() {
                        if (!this._state) return '{}';
                        return JSON.stringify({
                                blocks: this._state.blocks,
                                rootId: this._state.rootId,
                        });
                }

                getBlock(id) {
                        return this._state?.blocks[id] ?? null;
                }

                save() {
                        if (!this._state) return;
                        const serialized = this.serialize();
                        this._state.dirty = false;
                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:save', {
                                bubbles: true,
                                detail: { state: serialized },
                        }));
                }

                // ---- Lifecycle ----

                connectedCallback() {
                        window.addEventListener('beforeunload', this._beforeUnloadHandler);
                        window.addEventListener('keydown', this._escapeHandler);
                        this._initState();
                }

                disconnectedCallback() {
                        window.removeEventListener('beforeunload', this._beforeUnloadHandler);
                        window.removeEventListener('keydown', this._escapeHandler);
                }

                attributeChangedCallback(name, oldValue, newValue) {
                        if (name === 'data-state' && this.isConnected) {
                                this._initState();
                        }
                }

                // ---- State Initialization ----

                _initState() {
                        const stateAttr = this.getAttribute('data-state');

                        // No attribute or empty string — start empty
                        if (stateAttr === null || stateAttr === '') {
                                this._state = createState();
                                this._render();
                                return;
                        }

                        // Parse JSON
                        let parsed;
                        try {
                                parsed = JSON.parse(stateAttr);
                        } catch (parseError) {
                                this._renderError('Invalid block data', parseError.message);
                                this.dispatchEvent(new CustomEvent('block-editor:error', {
                                        bubbles: true,
                                        detail: { message: 'Invalid block data', error: parseError },
                                }));
                                return;
                        }

                        // Build state from parsed data
                        const newState = {
                                blocks: parsed.blocks || {},
                                rootId: parsed.rootId || null,
                                selectedBlockId: null,
                                dirty: false,
                        };

                        // Validate
                        const validationErrors = validateState(newState);
                        if (validationErrors.length > 0) {
                                const message = 'Block data has inconsistent pointers: ' + validationErrors[0];
                                this._renderError(message, validationErrors.join('\n'));
                                this.dispatchEvent(new CustomEvent('block-editor:error', {
                                        bubbles: true,
                                        detail: { message, error: new Error(validationErrors.join('; ')) },
                                }));
                                return;
                        }

                        this._state = newState;
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();
                        this._render();
                }

                // ---- Rendering ----

                _render() {
                        this.shadowRoot.innerHTML = '';
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();

                        const container = document.createElement('div');
                        container.className = 'editor-container';

                        // Toolbar
                        const toolbar = this._renderToolbar();
                        container.appendChild(toolbar);

                        // Block list
                        const blockList = document.createElement('div');
                        blockList.className = 'block-list';
                        this._renderBlocksInto(blockList);
                        container.appendChild(blockList);

                        // Make container focusable for keyboard events
                        container.setAttribute('tabindex', '0');

                        // Delegated click handler
                        container.addEventListener('click', (e) => this._handleClick(e));

                        // Delegated pointerdown for drag initiation
                        container.addEventListener('pointerdown', (e) => this._onPointerDown(e));

                        // Keyboard shortcuts (scoped to editor container)
                        container.addEventListener('keydown', this._keydownHandler);

                        this.shadowRoot.appendChild(container);
                }

                _renderToolbar() {
                        const toolbar = document.createElement('div');
                        toolbar.className = 'toolbar';

                        // One button per block type
                        for (const [type, config] of Object.entries(BLOCK_TYPE_CONFIG)) {
                                const btn = document.createElement('button');
                                btn.textContent = '+ ' + config.label;
                                btn.dataset.action = 'add';
                                btn.dataset.blockType = type;
                                toolbar.appendChild(btn);
                        }

                        const saveBtn = document.createElement('button');
                        saveBtn.textContent = 'Save';
                        saveBtn.className = 'save-btn';
                        saveBtn.dataset.action = 'save';
                        saveBtn.disabled = !this._state?.dirty;
                        toolbar.appendChild(saveBtn);

                        return toolbar;
                }

                _renderBlocksInto(container) {
                        if (!this._state) return;
                        const rootList = getRootList(this._state);
                        for (const block of rootList) {
                                const wrapper = this._renderBlockWrapper(block, 0);
                                container.appendChild(wrapper);
                        }
                }

                /**
                 * Render a block-wrapper div containing the block-item header and
                 * optionally a children-container. The wrapper is indented by depth.
                 */
                _renderBlockWrapper(block, depth) {
                        const wrapper = document.createElement('div');
                        wrapper.className = 'block-wrapper';
                        wrapper.dataset.blockId = block.id;
                        wrapper.style.marginInlineStart = (depth * 24) + 'px';

                        const header = this._renderBlockHeader(block);
                        wrapper.appendChild(header);

                        // Render children if any
                        const children = getChildren(this._state, block.id);
                        if (children.length > 0) {
                                const childContainer = document.createElement('div');
                                childContainer.className = 'children-container';
                                childContainer.dataset.parentId = block.id;
                                for (const child of children) {
                                        childContainer.appendChild(this._renderBlockWrapper(child, depth + 1));
                                }
                                wrapper.appendChild(childContainer);
                        }

                        this._wrapperRegistry.set(block.id, wrapper);
                        return wrapper;
                }

                /**
                 * Render the block-item header element (badge, label, child count, delete button,
                 * type-specific content).
                 */
                _renderBlockHeader(block) {
                        const el = document.createElement('div');
                        el.className = 'block-item';
                        el.dataset.blockId = block.id;

                        // Add type-specific class for container styling
                        if (block.type === 'container') {
                                el.classList.add('block-item--container');
                        }

                        const badge = document.createElement('span');
                        badge.className = 'block-type-badge block-type-badge--' + block.type;
                        badge.textContent = BLOCK_TYPE_CONFIG[block.type]?.label || block.type;
                        el.appendChild(badge);

                        const label = document.createElement('span');
                        label.className = 'block-label';
                        label.textContent = block.label;
                        el.appendChild(label);

                        // Child count badge
                        const childCount = getDescendantCount(this._state, block.id);
                        if (childCount > 0) {
                                const countBadge = document.createElement('span');
                                countBadge.className = 'child-count-badge';
                                countBadge.textContent = String(childCount);
                                countBadge.title = childCount + ' descendant' + (childCount === 1 ? '' : 's');
                                el.appendChild(countBadge);
                        }

                        const deleteBtn = document.createElement('button');
                        deleteBtn.className = 'block-delete-btn';
                        deleteBtn.textContent = 'Delete';
                        deleteBtn.dataset.action = 'delete';
                        deleteBtn.dataset.blockId = block.id;
                        el.appendChild(deleteBtn);

                        // Hover toolbar (child of .block-item to avoid pointerleave flicker)
                        const hoverToolbar = this._renderHoverToolbar(block.id);
                        el.appendChild(hoverToolbar);

                        // Type-specific content area
                        const content = this._renderTypeContent(block);
                        if (content) {
                                el.appendChild(content);
                        }

                        this._elementRegistry.set(block.id, el);
                        return el;
                }

                /**
                 * Render the hover toolbar for a block header.
                 * Contains: Move Up, Move Down, Indent, Outdent, Duplicate, Delete.
                 * All buttons have data-action attributes so _onPointerDown excludes
                 * them from drag initiation.
                 */
                _renderHoverToolbar(blockId) {
                        const toolbar = document.createElement('div');
                        toolbar.className = 'block-hover-toolbar';

                        const actions = [
                                { label: '\u2191', action: 'toolbar-move-up', title: 'Move Up' },
                                { label: '\u2193', action: 'toolbar-move-down', title: 'Move Down' },
                                { label: '\u2192', action: 'toolbar-indent', title: 'Indent (Tab)' },
                                { label: '\u2190', action: 'toolbar-outdent', title: 'Outdent (Shift+Tab)' },
                                { label: 'Dup', action: 'toolbar-duplicate', title: 'Duplicate (Ctrl+Shift+D)' },
                                { label: 'Del', action: 'toolbar-delete', title: 'Delete' },
                        ];

                        for (const def of actions) {
                                const btn = document.createElement('button');
                                btn.textContent = def.label;
                                btn.title = def.title;
                                btn.dataset.action = def.action;
                                btn.dataset.blockId = blockId;
                                toolbar.appendChild(btn);
                        }

                        return toolbar;
                }

                /**
                 * Render type-specific content for a block.
                 * text = paragraph placeholder lines
                 * heading = large bold label
                 * image = dashed rect placeholder
                 * container = labeled children area indicator
                 */
                _renderTypeContent(block) {
                        const wrapper = document.createElement('div');
                        wrapper.className = 'block-type-content block-type-content--' + block.type;

                        if (block.type === 'text') {
                                // Three placeholder lines simulating paragraph text
                                for (let i = 0; i < 3; i++) {
                                        const line = document.createElement('div');
                                        line.className = 'text-line';
                                        wrapper.appendChild(line);
                                }
                                return wrapper;
                        }

                        if (block.type === 'heading') {
                                wrapper.textContent = block.label;
                                return wrapper;
                        }

                        if (block.type === 'image') {
                                wrapper.textContent = 'Image placeholder';
                                return wrapper;
                        }

                        if (block.type === 'container') {
                                const childCount = getDescendantCount(this._state, block.id);
                                if (childCount > 0) {
                                        wrapper.textContent = childCount + ' block' + (childCount === 1 ? '' : 's') + ' inside';
                                } else {
                                        wrapper.textContent = 'Drop blocks here';
                                }
                                return wrapper;
                        }

                        return null;
                }

                /**
                 * Render children of a parent block into its children container.
                 * Used during DOM patching when a new children container is created.
                 */
                _renderChildrenInto(childContainer, parentId, parentDepth) {
                        const children = getChildren(this._state, parentId);
                        for (const child of children) {
                                const wrapper = this._renderBlockWrapper(child, parentDepth + 1);
                                childContainer.appendChild(wrapper);
                        }
                }

                _renderError(message, detail) {
                        this.shadowRoot.innerHTML = '';
                        const container = document.createElement('div');
                        container.className = 'editor-container';

                        const errorDiv = document.createElement('div');
                        errorDiv.className = 'error-container';

                        const msgEl = document.createElement('div');
                        msgEl.className = 'error-message';
                        msgEl.textContent = message;
                        errorDiv.appendChild(msgEl);

                        if (detail) {
                                const detailEl = document.createElement('div');
                                detailEl.className = 'error-detail';
                                detailEl.textContent = detail;
                                errorDiv.appendChild(detailEl);
                        }

                        container.appendChild(errorDiv);
                        this.shadowRoot.appendChild(container);
                }

                // ---- Event Handling ----

                _handleClick(e) {
                        const target = e.target;
                        const action = target.dataset?.action;
                        if (!action) return;

                        if (action === 'add') {
                                const blockType = target.dataset.blockType || 'text';
                                this._doAddBlock(blockType);
                        } else if (action === 'delete' || action === 'toolbar-delete') {
                                this._doDeleteBlock(target.dataset.blockId);
                        } else if (action === 'save') {
                                this.save();
                        } else if (action === 'toolbar-move-up') {
                                this._doMoveBlockUp(target.dataset.blockId);
                        } else if (action === 'toolbar-move-down') {
                                this._doMoveBlockDown(target.dataset.blockId);
                        } else if (action === 'toolbar-indent') {
                                this._doIndentBlock(target.dataset.blockId);
                        } else if (action === 'toolbar-outdent') {
                                this._doOutdentBlock(target.dataset.blockId);
                        } else if (action === 'toolbar-duplicate') {
                                this._doDuplicateBlock(target.dataset.blockId);
                        }
                }

                _doAddBlock(type) {
                        const id = addBlock(this._state, type);
                        this._devValidate();

                        // Patch DOM — append new block wrapper to block list (root level, depth 0)
                        const block = this._state.blocks[id];
                        const wrapper = this._renderBlockWrapper(block, 0);
                        const blockList = this.shadowRoot.querySelector('.block-list');
                        blockList.appendChild(wrapper);

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                detail: { action: 'add', blockId: id },
                        }));
                }

                _doDeleteBlock(blockId) {
                        const block = this._state.blocks[blockId];
                        if (!block) return;

                        // Check for children — confirm if has descendants
                        const descendantCount = getDescendantCount(this._state, blockId);
                        if (descendantCount > 0) {
                                const confirmed = confirm(`Delete "${block.label}" and ${descendantCount} children?`);
                                if (!confirmed) return;
                        }

                        // Remember parent before removal so we can clean up empty children containers
                        const parentId = block.parentId;

                        const removedIds = removeBlock(this._state, blockId);
                        this._devValidate();

                        // Clear selection if the selected block was among those removed
                        if (this._state.selectedBlockId && removedIds.includes(this._state.selectedBlockId)) {
                                this._state.selectedBlockId = null;
                        }

                        // Patch DOM — remove wrapper elements (wrapper contains both header and children)
                        for (const id of removedIds) {
                                const wrapper = this._wrapperRegistry.get(id);
                                if (wrapper) {
                                        wrapper.remove();
                                        this._wrapperRegistry.delete(id);
                                }
                                const el = this._elementRegistry.get(id);
                                if (el) {
                                        this._elementRegistry.delete(id);
                                }
                        }

                        // Children container cleanup: if parent lost its last child, remove the empty container
                        this._cleanupEmptyChildrenContainer(parentId);

                        // Update parent's child count badge
                        this._updateChildCountBadge(parentId);

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                detail: { action: 'remove', blockId },
                        }));
                }

                // ---- Phase 4: Selection ----

                _selectBlock(blockId) {
                        if (!this._state) return;

                        // Deselect previous
                        if (this._state.selectedBlockId) {
                                const prevEl = this._elementRegistry.get(this._state.selectedBlockId);
                                if (prevEl) {
                                        prevEl.classList.remove('selected');
                                }
                        }

                        // Select new (or deselect if clicking the same block)
                        if (this._state.selectedBlockId === blockId) {
                                this._state.selectedBlockId = null;
                                return;
                        }

                        this._state.selectedBlockId = blockId;
                        const el = this._elementRegistry.get(blockId);
                        if (el) {
                                el.classList.add('selected');
                        }

                        this.dispatchEvent(new CustomEvent('block-editor:select', {
                                bubbles: true,
                                detail: { blockId },
                        }));
                }

                // ---- Phase 4: Keyboard Shortcuts ----

                _onKeyDown(e) {
                        if (!this._state) return;

                        // Tab = indent, Shift+Tab = outdent
                        if (e.key === 'Tab') {
                                const blockId = this._state.selectedBlockId;
                                if (!blockId) return; // No selection — let browser handle Tab normally

                                e.preventDefault();

                                if (e.shiftKey) {
                                        this._doOutdentBlock(blockId);
                                } else {
                                        this._doIndentBlock(blockId);
                                }
                                return;
                        }

                        // Arrow Up/Down = move selection through depth-first traversal
                        if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
                                const blockId = this._state.selectedBlockId;
                                if (!blockId) return; // No selection — let browser handle arrows normally

                                const order = getBlockTraversalOrder(this._state);
                                if (order.length === 0) return;

                                const currentIndex = order.indexOf(blockId);
                                if (currentIndex === -1) return;

                                let nextIndex;
                                if (e.key === 'ArrowUp') {
                                        nextIndex = currentIndex - 1;
                                } else {
                                        nextIndex = currentIndex + 1;
                                }

                                // Clamp to bounds — do not wrap around
                                if (nextIndex < 0 || nextIndex >= order.length) return;

                                e.preventDefault();
                                this._selectBlock(order[nextIndex]);
                                return;
                        }

                        // Ctrl+Shift+D / Cmd+Shift+D = duplicate selected block
                        if (e.key === 'd' || e.key === 'D') {
                                if ((e.ctrlKey || e.metaKey) && e.shiftKey) {
                                        const blockId = this._state.selectedBlockId;
                                        if (!blockId) return;

                                        e.preventDefault();
                                        this._doDuplicateBlock(blockId);
                                        return;
                                }
                        }

                        // Delete / Backspace = delete selected block
                        if (e.key === 'Delete' || e.key === 'Backspace') {
                                const blockId = this._state.selectedBlockId;
                                if (!blockId) return; // No selection — let browser handle normally

                                e.preventDefault();
                                this._doDeleteBlock(blockId);
                                return;
                        }

                        // Enter = add new block after selected (default type 'text')
                        if (e.key === 'Enter') {
                                const blockId = this._state.selectedBlockId;
                                if (!blockId) return; // No selection — let browser handle normally

                                e.preventDefault();
                                this._doAddBlockAfter(blockId, 'text');
                                return;
                        }
                }

                /**
                 * Indent a block: make it the last child of its previous sibling.
                 * Handles both state mutation and DOM patching.
                 */
                _doIndentBlock(blockId) {
                        if (!this._state) return;

                        const block = this._state.blocks[blockId];
                        if (!block) return;

                        // Remember old parent before mutation for cleanup
                        const oldParentId = block.parentId;

                        // The previous sibling will become the new parent
                        const newParentId = block.prevSiblingId;

                        const result = indentBlock(this._state, blockId);
                        if (!result) return;

                        this._devValidate();

                        // Patch DOM: move block wrapper into new parent's children container
                        const blockWrapper = this._wrapperRegistry.get(blockId);
                        const newParentWrapper = this._wrapperRegistry.get(newParentId);

                        if (blockWrapper && newParentWrapper) {
                                // Get or create new parent's children container
                                let childContainer = newParentWrapper.querySelector(':scope > .children-container');
                                if (!childContainer) {
                                        childContainer = document.createElement('div');
                                        childContainer.className = 'children-container';
                                        childContainer.dataset.parentId = newParentId;
                                        newParentWrapper.appendChild(childContainer);
                                }

                                // Append as last child
                                childContainer.appendChild(blockWrapper);

                                // Update indentation for block and its descendants
                                const newDepth = getDepth(this._state, blockId);
                                blockWrapper.style.marginInlineStart = (newDepth * 24) + 'px';
                                this._updateDescendantDepths(blockId);
                        }

                        // Clean up old parent's empty children container
                        this._cleanupEmptyChildrenContainer(oldParentId);

                        // Update child count badges
                        this._updateChildCountBadge(oldParentId);
                        this._updateChildCountBadge(newParentId);

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                detail: { action: 'indent', blockId },
                        }));
                }

                /**
                 * Outdent a block: move it out of its parent to become the next sibling
                 * of that parent. Younger siblings are reparented under the block.
                 * Handles both state mutation and DOM patching.
                 */
                _doOutdentBlock(blockId) {
                        if (!this._state) return;

                        const block = this._state.blocks[blockId];
                        if (!block) return;

                        // Remember old parent and younger siblings before mutation
                        const oldParentId = block.parentId;
                        if (!oldParentId) return; // Already root

                        // Collect younger sibling IDs before state mutation
                        const youngerSiblingIds = [];
                        let walkId = block.nextSiblingId;
                        while (walkId) {
                                youngerSiblingIds.push(walkId);
                                walkId = this._state.blocks[walkId].nextSiblingId;
                        }

                        const result = outdentBlock(this._state, blockId);
                        if (!result) return;

                        this._devValidate();

                        // Patch DOM
                        const blockWrapper = this._wrapperRegistry.get(blockId);
                        const oldParentWrapper = this._wrapperRegistry.get(oldParentId);

                        if (blockWrapper && oldParentWrapper) {
                                // Move block wrapper to after old parent wrapper in the grandparent container
                                const nextSibling = oldParentWrapper.nextElementSibling;
                                if (nextSibling) {
                                        oldParentWrapper.parentNode.insertBefore(blockWrapper, nextSibling);
                                } else {
                                        oldParentWrapper.parentNode.appendChild(blockWrapper);
                                }

                                // Update block's indentation
                                const newDepth = getDepth(this._state, blockId);
                                blockWrapper.style.marginInlineStart = (newDepth * 24) + 'px';

                                // Move younger sibling wrappers into block's children container
                                if (youngerSiblingIds.length > 0) {
                                        let childContainer = blockWrapper.querySelector(':scope > .children-container');
                                        if (!childContainer) {
                                                childContainer = document.createElement('div');
                                                childContainer.className = 'children-container';
                                                childContainer.dataset.parentId = blockId;
                                                blockWrapper.appendChild(childContainer);
                                        }

                                        for (const sibId of youngerSiblingIds) {
                                                const sibWrapper = this._wrapperRegistry.get(sibId);
                                                if (sibWrapper) {
                                                        childContainer.appendChild(sibWrapper);
                                                }
                                        }
                                }

                                // Update indentation for all descendants (including reparented younger siblings)
                                this._updateDescendantDepths(blockId);
                        }

                        // Clean up old parent's empty children container
                        this._cleanupEmptyChildrenContainer(oldParentId);

                        // Update child count badges
                        this._updateChildCountBadge(oldParentId);
                        this._updateChildCountBadge(blockId);

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                detail: { action: 'outdent', blockId },
                        }));
                }

                // ---- Phase 6: Duplicate, Move Up, Move Down ----

                /**
                 * Duplicate a block and its entire subtree. Renders the cloned subtree
                 * and inserts it into the DOM after the original's wrapper.
                 */
                _doDuplicateBlock(blockId) {
                        if (!this._state) return;
                        if (!blockId) return;
                        const block = this._state.blocks[blockId];
                        if (!block) return;

                        const cloneId = duplicateBlock(this._state, blockId);
                        if (!cloneId) return;

                        this._devValidate();

                        // Patch DOM: render the cloned subtree and insert after original wrapper
                        const originalWrapper = this._wrapperRegistry.get(blockId);
                        if (originalWrapper) {
                                const cloneBlock = this._state.blocks[cloneId];
                                const depth = getDepth(this._state, cloneId);
                                const cloneWrapper = this._renderBlockWrapper(cloneBlock, depth);

                                const nextSibling = originalWrapper.nextElementSibling;
                                if (nextSibling) {
                                        originalWrapper.parentNode.insertBefore(cloneWrapper, nextSibling);
                                } else {
                                        originalWrapper.parentNode.appendChild(cloneWrapper);
                                }
                        }

                        // Update parent's child count badge
                        const parentId = this._state.blocks[cloneId]?.parentId;
                        this._updateChildCountBadge(parentId);

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                detail: { action: 'duplicate', blockId },
                        }));
                }

                /**
                 * Move a block before its previous sibling.
                 * Handles state mutation and surgical DOM patching.
                 */
                _doMoveBlockUp(blockId) {
                        if (!this._state) return;
                        if (!blockId) return;
                        const block = this._state.blocks[blockId];
                        if (!block) return;

                        // Remember the previous sibling's wrapper before state mutation
                        const prevSiblingId = block.prevSiblingId;
                        if (!prevSiblingId) return;

                        const result = moveBlockUp(this._state, blockId);
                        if (!result) return;

                        this._devValidate();

                        // Patch DOM: move block wrapper before the previous sibling's wrapper
                        const blockWrapper = this._wrapperRegistry.get(blockId);
                        const prevWrapper = this._wrapperRegistry.get(prevSiblingId);

                        if (blockWrapper && prevWrapper) {
                                prevWrapper.parentNode.insertBefore(blockWrapper, prevWrapper);
                        }

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                detail: { action: 'moveUp', blockId },
                        }));
                }

                /**
                 * Move a block after its next sibling.
                 * Handles state mutation and surgical DOM patching.
                 */
                _doMoveBlockDown(blockId) {
                        if (!this._state) return;
                        if (!blockId) return;
                        const block = this._state.blocks[blockId];
                        if (!block) return;

                        // Remember the next sibling's wrapper before state mutation
                        const nextSiblingId = block.nextSiblingId;
                        if (!nextSiblingId) return;

                        const result = moveBlockDown(this._state, blockId);
                        if (!result) return;

                        this._devValidate();

                        // Patch DOM: move block wrapper after the next sibling's wrapper
                        const blockWrapper = this._wrapperRegistry.get(blockId);
                        const nextWrapper = this._wrapperRegistry.get(nextSiblingId);

                        if (blockWrapper && nextWrapper) {
                                const afterNext = nextWrapper.nextElementSibling;
                                if (afterNext) {
                                        nextWrapper.parentNode.insertBefore(blockWrapper, afterNext);
                                } else {
                                        nextWrapper.parentNode.appendChild(blockWrapper);
                                }
                        }

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                detail: { action: 'moveDown', blockId },
                        }));
                }

                /**
                 * Add a new block after the specified block (or at end of root list).
                 * Used by Enter key and toolbar. Default type is 'text'.
                 */
                _doAddBlockAfter(afterBlockId, type) {
                        if (!this._state) return;
                        if (!afterBlockId) return;

                        const afterBlock = this._state.blocks[afterBlockId];
                        if (!afterBlock) return;

                        const id = addBlock(this._state, type || 'text', afterBlockId);
                        this._devValidate();

                        // Patch DOM: render new block wrapper and insert after the target wrapper
                        const block = this._state.blocks[id];
                        const depth = getDepth(this._state, id);
                        const wrapper = this._renderBlockWrapper(block, depth);
                        const afterWrapper = this._wrapperRegistry.get(afterBlockId);

                        if (afterWrapper) {
                                const nextSibling = afterWrapper.nextElementSibling;
                                if (nextSibling) {
                                        afterWrapper.parentNode.insertBefore(wrapper, nextSibling);
                                } else {
                                        afterWrapper.parentNode.appendChild(wrapper);
                                }
                        }

                        // Select the new block
                        this._selectBlock(id);

                        // Update parent's child count badge
                        const parentId = this._state.blocks[id]?.parentId;
                        this._updateChildCountBadge(parentId);

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                detail: { action: 'add', blockId: id },
                        }));
                }

                // ---- Phase 3: Drag and Drop with Nesting ----

                _onPointerDown(e) {
                        // Only primary button (left click / touch)
                        if (e.button !== 0) return;

                        // Don't drag when clicking delete button or toolbar buttons
                        if (e.target.closest('[data-action]')) return;

                        // Find the block header (.block-item) that was clicked
                        const blockItem = e.target.closest('.block-item');
                        if (!blockItem) return;

                        const blockId = blockItem.dataset.blockId;
                        if (!blockId) return;

                        // Phase 3: Allow dragging ALL blocks (root and nested)
                        const block = this._state?.blocks[blockId];
                        if (!block) return;

                        // Record start position for threshold detection
                        const startX = e.clientX;
                        const startY = e.clientY;

                        // Pre-threshold handlers attached to the block header itself
                        const onPreMove = (moveEvent) => {
                                try {
                                        this._onPreThresholdMove(moveEvent, blockItem, blockId, startX, startY, onPreMove, onPreUp);
                                } catch (err) {
                                        console.error('[block-editor] Error in pre-threshold move:', err);
                                        blockItem.removeEventListener('pointermove', onPreMove);
                                        blockItem.removeEventListener('pointerup', onPreUp);
                                }
                        };

                        const onPreUp = () => {
                                // Pointer released before threshold — this was a click, not a drag
                                blockItem.removeEventListener('pointermove', onPreMove);
                                blockItem.removeEventListener('pointerup', onPreUp);
                                this._selectBlock(blockId);
                        };

                        blockItem.addEventListener('pointermove', onPreMove);
                        blockItem.addEventListener('pointerup', onPreUp);
                }

                _onPreThresholdMove(e, blockItem, blockId, startX, startY, preMoveFn, preUpFn) {
                        const dx = e.clientX - startX;
                        const dy = e.clientY - startY;
                        const distance = Math.sqrt(dx * dx + dy * dy);

                        if (distance < 8) return; // Below 8px threshold

                        // Threshold crossed — remove pre-threshold listeners
                        blockItem.removeEventListener('pointermove', preMoveFn);
                        blockItem.removeEventListener('pointerup', preUpFn);

                        // Start the real drag
                        this._startDrag(e, blockItem, blockId, startX, startY);
                }

                _startDrag(e, blockItem, blockId, startX, startY) {
                        try {
                                // Capture pointer on the block header
                                blockItem.setPointerCapture(e.pointerId);

                                // Compute grab offset from the block's top-left corner
                                const blockRect = blockItem.getBoundingClientRect();
                                const grabOffsetX = startX - blockRect.left;
                                const grabOffsetY = startY - blockRect.top + 985;

                                // Create drag overlay (cloned block header)
                                const overlay = this._createDragOverlay(blockItem, e.clientX, e.clientY, grabOffsetX, grabOffsetY);

                                // Create drag-move and drag-end handlers bound to this drag session
                                const onDragMove = (moveEvent) => {
                                        try {
                                                this._onDragMove(moveEvent);
                                        } catch (err) {
                                                console.error('[block-editor] Error in drag move:', err);
                                                this._cleanupDrag();
                                        }
                                };

                                const onDragEnd = (upEvent) => {
                                        try {
                                                this._onDragEnd(upEvent);
                                        } catch (err) {
                                                console.error('[block-editor] Error in drag end:', err);
                                                this._cleanupDrag();
                                        }
                                };

                                // Store drag session state
                                this._drag = {
                                        blockId,
                                        blockItem,
                                        overlay,
                                        grabOffsetX,
                                        grabOffsetY,
                                        pointerId: e.pointerId,
                                        onDragMove,
                                        onDragEnd,
                                        dropTarget: null, // { blockId, position: 'before'|'after'|'inside' }
                                };

                                // Mark original block as dragging
                                blockItem.classList.add('dragging');

                                // Attach drag-mode handlers to the block header (capture routes all events here)
                                blockItem.addEventListener('pointermove', onDragMove);
                                blockItem.addEventListener('pointerup', onDragEnd);
                                blockItem.addEventListener('pointercancel', onDragEnd);
                        } catch (err) {
                                console.error('[block-editor] Error starting drag:', err);
                                this._cleanupDrag();
                        }
                }

                _createDragOverlay(blockItem, clientX, clientY, grabOffsetX, grabOffsetY) {
                        const overlay = blockItem.cloneNode(true);
                        overlay.className = 'block-item drag-overlay';
                        // Match the original width so it does not collapse
                        overlay.style.width = blockItem.getBoundingClientRect().width + 'px';
                        overlay.style.transform = 'translate(' + (clientX - grabOffsetX) + 'px, ' + (clientY - grabOffsetY) + 'px)';

                        // Append to shadow root so it renders above everything
                        this.shadowRoot.appendChild(overlay);
                        return overlay;
                }

                _onDragMove(e) {
                        if (!this._drag) return;

                        const { overlay, grabOffsetX, grabOffsetY } = this._drag;

                        // Update overlay position using viewport-relative clientX/clientY
                        overlay.style.transform = 'translate(' + (e.clientX - grabOffsetX) + 'px, ' + (e.clientY - grabOffsetY) + 'px)';

                        // Track pointer for auto-scroll
                        this._lastPointerY = e.clientY;

                        // Start auto-scroll if not already running
                        this._startAutoScroll();

                        // Compute drop zone using three-zone detection
                        const dropTarget = this._computeDropZone(e.clientX, e.clientY);
                        this._drag.dropTarget = dropTarget;

                        if (dropTarget) {
                                this._updateDropIndicator(dropTarget);
                        } else {
                                this._removeDropIndicator();
                                this._removeDropInsideHighlight();
                        }
                }

                /**
                 * Three-zone drop detection against block header rects (not full subtree).
                 * Works on ALL blocks (root and nested).
                 * beforeZone = Math.max(rect.height * 0.25, 10)
                 * afterZone = Math.max(rect.height * 0.25, 10)
                 * relativeY in beforeZone -> "before", in afterZone -> "after", else -> "inside"
                 *
                 * Coercions:
                 * - "inside" coerced to "after" if canHaveChildren is false
                 * - "inside" coerced to "after" if target depth + 1 > 8 (max nesting)
                 * - Skip entirely if target is descendant of dragged block
                 */
                _computeDropZone(clientX, clientY) {
                        if (!this._drag) return null;

                        // Get ALL block-item headers in the shadow DOM
                        const allBlockItems = this.shadowRoot.querySelectorAll('.block-item');
                        let closestTarget = null;
                        let closestDistance = Infinity;

                        for (const item of allBlockItems) {
                                const itemBlockId = item.dataset.blockId;
                                if (!itemBlockId) continue;

                                // Skip the block being dragged
                                if (itemBlockId === this._drag.blockId) continue;

                                // Skip if item is a drag overlay
                                if (item.classList.contains('drag-overlay')) continue;

                                // Skip if target is a descendant of the dragged block (prevents circular nesting)
                                if (isDescendant(this._state, itemBlockId, this._drag.blockId)) continue;

                                const rect = item.getBoundingClientRect();

                                // Check if cursor is within this block header's vertical range
                                if (clientY >= rect.top && clientY <= rect.bottom) {
                                        const beforeZone = Math.max(rect.height * 0.25, 10);
                                        const afterZone = Math.max(rect.height * 0.25, 10);
                                        const relativeY = clientY - rect.top;

                                        let position;
                                        if (relativeY < beforeZone) {
                                                position = 'before';
                                        } else if (relativeY > rect.height - afterZone) {
                                                position = 'after';
                                        } else {
                                                position = 'inside';
                                        }

                                        // Coerce "inside" to "after" if block type cannot have children
                                        if (position === 'inside') {
                                                const targetBlock = this._state.blocks[itemBlockId];
                                                const config = BLOCK_TYPE_CONFIG[targetBlock?.type];
                                                if (!config || !config.canHaveChildren) {
                                                        position = 'after';
                                                }
                                        }

                                        // Coerce "inside" to "after" if max nesting depth (8) would be exceeded
                                        if (position === 'inside') {
                                                const targetDepth = getDepth(this._state, itemBlockId);
                                                if (targetDepth + 1 > 8) {
                                                        position = 'after';
                                                }
                                        }

                                        return { blockId: itemBlockId, position };
                                }

                                // Track closest block for edge cases (cursor above/below all)
                                const distToTop = Math.abs(clientY - rect.top);
                                const distToBottom = Math.abs(clientY - rect.bottom);
                                const minDist = Math.min(distToTop, distToBottom);
                                if (minDist < closestDistance) {
                                        closestDistance = minDist;
                                        closestTarget = { item, rect, distToTop, distToBottom };
                                }
                        }

                        // Cursor is outside all block headers — snap to nearest block
                        if (closestTarget) {
                                const { item, rect } = closestTarget;
                                const itemBlockId = item.dataset.blockId;
                                if (clientY < rect.top) {
                                        return { blockId: itemBlockId, position: 'before' };
                                }
                                return { blockId: itemBlockId, position: 'after' };
                        }

                        return null;
                }

                _updateDropIndicator(dropTarget) {
                        const targetEl = this._elementRegistry.get(dropTarget.blockId);
                        if (!targetEl) return;

                        // Handle "inside" — highlight the target block header, no line indicator
                        if (dropTarget.position === 'inside') {
                                this._removeDropIndicator();
                                this._applyDropInsideHighlight(dropTarget.blockId);
                                return;
                        }

                        // "before" or "after" — show line indicator, remove inside highlight
                        this._removeDropInsideHighlight();

                        const blockList = this.shadowRoot.querySelector('.block-list');
                        if (!blockList) return;

                        // Create indicator if it does not exist
                        if (!this._dropIndicator) {
                                this._dropIndicator = document.createElement('div');
                                this._dropIndicator.className = 'drop-indicator';
                                blockList.appendChild(this._dropIndicator);
                        }

                        const targetRect = targetEl.getBoundingClientRect();
                        const listRect = blockList.getBoundingClientRect();

                        // Position relative to block list container
                        if (dropTarget.position === 'before') {
                                this._dropIndicator.style.top = (targetRect.top - listRect.top - 1) + 'px';
                        } else {
                                this._dropIndicator.style.top = (targetRect.bottom - listRect.top - 1) + 'px';
                        }
                }

                _applyDropInsideHighlight(blockId) {
                        // Remove previous highlight
                        this._removeDropInsideHighlight();

                        const targetEl = this._elementRegistry.get(blockId);
                        if (targetEl) {
                                targetEl.classList.add('drop-inside');
                        }
                }

                _removeDropInsideHighlight() {
                        const highlighted = this.shadowRoot.querySelector('.drop-inside');
                        if (highlighted) {
                                highlighted.classList.remove('drop-inside');
                        }
                }

                _removeDropIndicator() {
                        if (this._dropIndicator) {
                                this._dropIndicator.remove();
                                this._dropIndicator = null;
                        }
                }

                _onDragEnd(e) {
                        if (!this._drag) return;

                        const { dropTarget } = this._drag;

                        if (dropTarget) {
                                this._executeDrop(dropTarget);
                        }

                        this._cleanupDrag();
                }

                /**
                 * Execute a drop: mutate state, patch DOM, handle nesting.
                 * For "inside" drops: move block's wrapper into target's children container.
                 * For "before"/"after": move block's wrapper next to target's wrapper.
                 */
                _executeDrop(dropTarget) {
                        if (!this._drag || !this._state) return;

                        const { blockId } = this._drag;
                        const { blockId: targetId, position } = dropTarget;

                        // Remember the old parent before state mutation so we can clean up empty children containers
                        const oldParentId = this._state.blocks[blockId]?.parentId;

                        // Mutate state
                        moveBlock(this._state, blockId, targetId, position);
                        this._devValidate();

                        // Patch DOM: move the block wrapper to its new position
                        const blockWrapper = this._wrapperRegistry.get(blockId);
                        const targetWrapper = this._wrapperRegistry.get(targetId);

                        if (blockWrapper && targetWrapper) {
                                if (position === 'inside') {
                                        // Get or create target's children container
                                        let childContainer = targetWrapper.querySelector(':scope > .children-container');
                                        if (!childContainer) {
                                                childContainer = document.createElement('div');
                                                childContainer.className = 'children-container';
                                                childContainer.dataset.parentId = targetId;
                                                targetWrapper.appendChild(childContainer);
                                        }

                                        // Update block wrapper indentation to new depth
                                        const newDepth = getDepth(this._state, blockId);
                                        blockWrapper.style.marginInlineStart = (newDepth * 24) + 'px';

                                        // Also update indentation for all descendant wrappers
                                        this._updateDescendantDepths(blockId);

                                        // Insert as first child in the container
                                        if (childContainer.firstElementChild) {
                                                childContainer.insertBefore(blockWrapper, childContainer.firstElementChild);
                                        } else {
                                                childContainer.appendChild(blockWrapper);
                                        }
                                } else if (position === 'before') {
                                        // Insert before the target wrapper in its parent container
                                        targetWrapper.parentNode.insertBefore(blockWrapper, targetWrapper);

                                        // Update indentation
                                        const newDepth = getDepth(this._state, blockId);
                                        blockWrapper.style.marginInlineStart = (newDepth * 24) + 'px';
                                        this._updateDescendantDepths(blockId);
                                } else {
                                        // "after" — insert after target wrapper
                                        const nextSibling = targetWrapper.nextElementSibling;
                                        if (nextSibling) {
                                                targetWrapper.parentNode.insertBefore(blockWrapper, nextSibling);
                                        } else {
                                                targetWrapper.parentNode.appendChild(blockWrapper);
                                        }

                                        // Update indentation
                                        const newDepth = getDepth(this._state, blockId);
                                        blockWrapper.style.marginInlineStart = (newDepth * 24) + 'px';
                                        this._updateDescendantDepths(blockId);
                                }
                        }

                        // Children container cleanup: if old parent lost its last child, remove empty container
                        this._cleanupEmptyChildrenContainer(oldParentId);

                        // Update child count badges for affected parents
                        this._updateChildCountBadge(oldParentId);
                        this._updateChildCountBadge(targetId);
                        // If the block itself gained/lost children representation, update it too
                        this._updateChildCountBadge(blockId);

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                detail: { action: 'move', blockId, targetId, position },
                        }));
                }

                /**
                 * Recursively update marginInlineStart on descendant wrappers to match their new depth.
                 */
                _updateDescendantDepths(parentBlockId) {
                        const children = getChildren(this._state, parentBlockId);
                        for (const child of children) {
                                const childWrapper = this._wrapperRegistry.get(child.id);
                                if (childWrapper) {
                                        const depth = getDepth(this._state, child.id);
                                        childWrapper.style.marginInlineStart = (depth * 24) + 'px';
                                }
                                this._updateDescendantDepths(child.id);
                        }
                }

                /**
                 * Remove an empty children container div from a parent's wrapper.
                 * This prevents visual artifacts (empty indented space) and incorrect
                 * elementFromPoint hit testing (empty container intercepting pointer events).
                 */
                _cleanupEmptyChildrenContainer(parentId) {
                        if (!parentId) return;
                        const parentBlock = this._state?.blocks[parentId];
                        if (!parentBlock) return;

                        // Only clean up if the parent truly has no children in state
                        if (parentBlock.firstChildId !== null) return;

                        const parentWrapper = this._wrapperRegistry.get(parentId);
                        if (!parentWrapper) return;

                        const childContainer = parentWrapper.querySelector(':scope > .children-container');
                        if (childContainer) {
                                childContainer.remove();
                        }
                }

                /**
                 * Update or remove the child count badge on a block's header element.
                 */
                _updateChildCountBadge(blockId) {
                        if (!blockId) return;
                        const block = this._state?.blocks[blockId];
                        if (!block) return;

                        const headerEl = this._elementRegistry.get(blockId);
                        if (!headerEl) return;

                        const existingBadge = headerEl.querySelector('.child-count-badge');
                        const childCount = getDescendantCount(this._state, blockId);

                        if (childCount > 0) {
                                if (existingBadge) {
                                        existingBadge.textContent = String(childCount);
                                        existingBadge.title = childCount + ' descendant' + (childCount === 1 ? '' : 's');
                                } else {
                                        // Insert badge before the delete button
                                        const countBadge = document.createElement('span');
                                        countBadge.className = 'child-count-badge';
                                        countBadge.textContent = String(childCount);
                                        countBadge.title = childCount + ' descendant' + (childCount === 1 ? '' : 's');
                                        const deleteBtn = headerEl.querySelector('.block-delete-btn');
                                        if (deleteBtn) {
                                                headerEl.insertBefore(countBadge, deleteBtn);
                                        } else {
                                                headerEl.appendChild(countBadge);
                                        }
                                }
                        } else {
                                if (existingBadge) {
                                        existingBadge.remove();
                                }
                        }
                }

                // ---- Auto-scroll ----

                _startAutoScroll() {
                        if (this._autoScrollRaf !== null) return; // Already running

                        const editorContainer = this.shadowRoot.querySelector('.editor-container');
                        if (!editorContainer) return;

                        const EDGE_ZONE = 40; // pixels from edge that triggers scrolling
                        const MAX_SPEED = 12; // max pixels per frame

                        const scrollStep = () => {
                                if (!this._drag) {
                                        this._stopAutoScroll();
                                        return;
                                }

                                const containerRect = editorContainer.getBoundingClientRect();
                                const pointerY = this._lastPointerY;

                                const distFromTop = pointerY - containerRect.top;
                                const distFromBottom = containerRect.bottom - pointerY;

                                if (distFromTop < EDGE_ZONE && distFromTop >= 0) {
                                        // Scroll up — speed proportional to closeness to edge
                                        const speed = Math.round(MAX_SPEED * (1 - distFromTop / EDGE_ZONE));
                                        editorContainer.scrollTop = editorContainer.scrollTop - speed;
                                } else if (distFromBottom < EDGE_ZONE && distFromBottom >= 0) {
                                        // Scroll down
                                        const speed = Math.round(MAX_SPEED * (1 - distFromBottom / EDGE_ZONE));
                                        editorContainer.scrollTop = editorContainer.scrollTop + speed;
                                } else {
                                        // Pointer moved out of edge zone — stop auto-scroll
                                        this._stopAutoScroll();
                                        return;
                                }

                                this._autoScrollRaf = requestAnimationFrame(scrollStep);
                        };

                        this._autoScrollRaf = requestAnimationFrame(scrollStep);
                }

                _stopAutoScroll() {
                        if (this._autoScrollRaf !== null) {
                                cancelAnimationFrame(this._autoScrollRaf);
                                this._autoScrollRaf = null;
                        }
                }

                _onEscapeKey(e) {
                        if (e.key === 'Escape' && this._drag) {
                                this._cleanupDrag();
                        }
                }

                _cleanupDrag() {
                        if (!this._drag) return;

                        const { blockItem, overlay, pointerId, onDragMove, onDragEnd } = this._drag;

                        // Stop auto-scroll
                        this._stopAutoScroll();

                        // Release pointer capture
                        try {
                                if (blockItem.hasPointerCapture(pointerId)) {
                                        blockItem.releasePointerCapture(pointerId);
                                }
                        } catch (captureErr) {
                                // Pointer capture may already be released; safe to ignore
                        }

                        // Remove drag-mode handlers
                        blockItem.removeEventListener('pointermove', onDragMove);
                        blockItem.removeEventListener('pointerup', onDragEnd);
                        blockItem.removeEventListener('pointercancel', onDragEnd);

                        // Remove overlay
                        if (overlay && overlay.parentNode) {
                                overlay.remove();
                        }

                        // Restore original block opacity
                        blockItem.classList.remove('dragging');

                        // Remove drop indicator and inside highlight
                        this._removeDropIndicator();
                        this._removeDropInsideHighlight();

                        // Clear drag state
                        this._drag = null;
                }

                _updateSaveButton() {
                        const saveBtn = this.shadowRoot.querySelector('[data-action="save"]');
                        if (saveBtn) {
                                saveBtn.disabled = !this._state?.dirty;
                        }
                }

                // ---- Dev-mode validation ----

                _devValidate() {
                        if (!this.dev) return;
                        const errors = validateState(this._state);
                        if (errors.length > 0) {
                                console.warn('[block-editor] Validation errors after mutation:', errors);
                        } else {
                                console.log('[block-editor] State valid');
                        }
                }

                // ---- beforeunload ----

                _onBeforeUnload(e) {
                        if (this._state?.dirty) {
                                e.preventDefault();
                                e.returnValue = 'You have unsaved changes.';
                        }
                }
        }

        customElements.define('block-editor', BlockEditor);
}
