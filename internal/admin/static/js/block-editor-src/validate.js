// validate.js — State validation (dev-mode only)

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
