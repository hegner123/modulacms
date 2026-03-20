// tree-ops.js — State-mutating tree operations

import { getTypeConfig, MAX_DEPTH } from './config.js';
import { generateId } from './id.js';
import { collectDescendants, getDepth, findLastSibling } from './tree-queries.js';

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
        const config = getTypeConfig(type);
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

export function addBlockFromDatatype(state, datatype, position, targetId) {
        var id = generateId();
        var block = {
                id: id,
                type: datatype.type,
                parentId: null,
                firstChildId: null,
                prevSiblingId: null,
                nextSiblingId: null,
                datatypeId: datatype.id,
                label: datatype.label,
                authorId: '',
                routeId: '',
                status: 'draft',
                dateCreated: new Date().toISOString(),
                dateModified: new Date().toISOString(),
                fields: [],
        };
        state.blocks[id] = block;

        if (!state.rootId) {
                state.rootId = id;
        } else if (position === 'before' && targetId) {
                insertBefore(state, id, targetId);
        } else if (position === 'after' && targetId) {
                insertAfter(state, id, targetId);
        } else if (position === 'inside' && targetId) {
                insertAsFirstChild(state, id, targetId);
        } else {
                var lastId = findLastSibling(state, state.rootId);
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
        const config = getTypeConfig(prevSibling.type);
        if (!config.canHaveChildren) return false;

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

                // Clone all properties from the original, then override pointers.
                // This preserves datatypeId, authorId, routeId, status, fields, etc.
                const clone = Object.assign({}, original, {
                        id: newId,
                        parentId: null,
                        firstChildId: null,
                        prevSiblingId: null,
                        nextSiblingId: null,
                });
                // Deep-clone fields array so edits to the clone don't affect the original.
                if (Array.isArray(original.fields)) {
                        clone.fields = original.fields.map(function(f) { return Object.assign({}, f); });
                }
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
