// tree-queries.js — Read-only tree traversal functions

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
        const visited = new Set();
        visited.add(currentId);
        while (state.blocks[currentId]?.nextSiblingId) {
                currentId = state.blocks[currentId].nextSiblingId;
                if (visited.has(currentId)) {
                        console.error('[block-editor] Cycle detected in sibling chain at block:', currentId);
                        break;
                }
                visited.add(currentId);
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

export function collectDescendants(state, blockId) {
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
