import { describe, it, expect } from 'vitest';
import {
        unlink, insertBefore, insertAfter, insertAsFirstChild, insertAsLastChild,
        addBlock, addBlockFromDatatype, removeBlock, moveBlock,
        indentBlock, outdentBlock, duplicateBlock, moveBlockUp, moveBlockDown,
} from './tree-ops.js';
import { validateState } from './validate.js';
import { getChildren, getRootList, getDescendantCount } from './tree-queries.js';

// --- Helpers ---

function makeBlock(id, overrides) {
        return {
                id, type: 'text',
                parentId: null, firstChildId: null,
                prevSiblingId: null, nextSiblingId: null,
                label: 'Block ' + id, fields: [],
                ...overrides,
        };
}

function makeState(blocks, rootId) {
        var state = { blocks: {}, rootId: rootId || null, selectedBlockId: null, dirty: false };
        for (var b of blocks) state.blocks[b.id] = b;
        return state;
}

function expectValid(state) {
        expect(validateState(state)).toEqual([]);
}

// Builds a root chain: ids[0] -> ids[1] -> ... (all at root level)
function rootChain(...ids) {
        var blocks = ids.map(function(id, i) {
                return makeBlock(id, {
                        prevSiblingId: i > 0 ? ids[i - 1] : null,
                        nextSiblingId: i < ids.length - 1 ? ids[i + 1] : null,
                });
        });
        return makeState(blocks, ids[0]);
}

// --- unlink ---

describe('unlink', function() {
        it('removes a middle sibling from the chain', function() {
                var state = rootChain('a', 'b', 'c');
                unlink(state, 'b');

                expect(state.blocks.a.nextSiblingId).toBe('c');
                expect(state.blocks.c.prevSiblingId).toBe('a');
                expect(state.blocks.b.parentId).toBeNull();
                expect(state.blocks.b.prevSiblingId).toBeNull();
                expect(state.blocks.b.nextSiblingId).toBeNull();
        });

        it('removes the first sibling and updates rootId', function() {
                var state = rootChain('a', 'b', 'c');
                unlink(state, 'a');

                expect(state.rootId).toBe('b');
                expect(state.blocks.b.prevSiblingId).toBeNull();
        });

        it('removes the last sibling', function() {
                var state = rootChain('a', 'b', 'c');
                unlink(state, 'c');

                expect(state.blocks.b.nextSiblingId).toBeNull();
        });

        it('removes the only root block', function() {
                var state = rootChain('a');
                unlink(state, 'a');

                expect(state.rootId).toBeNull();
        });

        it('removes a first child and updates parent.firstChildId', function() {
                var state = makeState([
                        makeBlock('parent', { firstChildId: 'c1' }),
                        makeBlock('c1', { parentId: 'parent', nextSiblingId: 'c2' }),
                        makeBlock('c2', { parentId: 'parent', prevSiblingId: 'c1' }),
                ], 'parent');

                unlink(state, 'c1');
                expect(state.blocks.parent.firstChildId).toBe('c2');
                expect(state.blocks.c2.prevSiblingId).toBeNull();
        });
});

// --- insertBefore / insertAfter ---

describe('insertBefore', function() {
        it('inserts before the first root block', function() {
                var state = rootChain('a', 'b');
                state.blocks.x = makeBlock('x');
                insertBefore(state, 'x', 'a');

                expect(state.rootId).toBe('x');
                expect(state.blocks.x.nextSiblingId).toBe('a');
                expect(state.blocks.a.prevSiblingId).toBe('x');
                expectValid(state);
        });

        it('inserts before a middle block', function() {
                var state = rootChain('a', 'b', 'c');
                state.blocks.x = makeBlock('x');
                insertBefore(state, 'x', 'b');

                expect(state.blocks.a.nextSiblingId).toBe('x');
                expect(state.blocks.x.prevSiblingId).toBe('a');
                expect(state.blocks.x.nextSiblingId).toBe('b');
                expect(state.blocks.b.prevSiblingId).toBe('x');
                expectValid(state);
        });
});

describe('insertAfter', function() {
        it('inserts after the last root block', function() {
                var state = rootChain('a', 'b');
                state.blocks.x = makeBlock('x');
                insertAfter(state, 'x', 'b');

                expect(state.blocks.b.nextSiblingId).toBe('x');
                expect(state.blocks.x.prevSiblingId).toBe('b');
                expect(state.blocks.x.nextSiblingId).toBeNull();
                expectValid(state);
        });

        it('inserts after a middle block', function() {
                var state = rootChain('a', 'b', 'c');
                state.blocks.x = makeBlock('x');
                insertAfter(state, 'x', 'b');

                expect(state.blocks.b.nextSiblingId).toBe('x');
                expect(state.blocks.x.nextSiblingId).toBe('c');
                expect(state.blocks.c.prevSiblingId).toBe('x');
                expectValid(state);
        });
});

// --- insertAsFirstChild / insertAsLastChild ---

describe('insertAsFirstChild', function() {
        it('adds first child to a childless parent', function() {
                var state = makeState([
                        makeBlock('parent', { type: 'container' }),
                ], 'parent');
                state.blocks.child = makeBlock('child');

                insertAsFirstChild(state, 'child', 'parent');

                expect(state.blocks.parent.firstChildId).toBe('child');
                expect(state.blocks.child.parentId).toBe('parent');
                expectValid(state);
        });

        it('pushes existing first child to second position', function() {
                var state = makeState([
                        makeBlock('parent', { type: 'container', firstChildId: 'c1' }),
                        makeBlock('c1', { parentId: 'parent' }),
                ], 'parent');
                state.blocks.c0 = makeBlock('c0');

                insertAsFirstChild(state, 'c0', 'parent');

                expect(state.blocks.parent.firstChildId).toBe('c0');
                expect(state.blocks.c0.nextSiblingId).toBe('c1');
                expect(state.blocks.c1.prevSiblingId).toBe('c0');
                expectValid(state);
        });
});

describe('insertAsLastChild', function() {
        it('adds to an empty parent (delegates to insertAsFirstChild)', function() {
                var state = makeState([
                        makeBlock('parent', { type: 'container' }),
                ], 'parent');
                state.blocks.child = makeBlock('child');

                insertAsLastChild(state, 'child', 'parent');

                expect(state.blocks.parent.firstChildId).toBe('child');
                expect(state.blocks.child.parentId).toBe('parent');
                expectValid(state);
        });

        it('appends after existing children', function() {
                var state = makeState([
                        makeBlock('parent', { type: 'container', firstChildId: 'c1' }),
                        makeBlock('c1', { parentId: 'parent', nextSiblingId: 'c2' }),
                        makeBlock('c2', { parentId: 'parent', prevSiblingId: 'c1' }),
                ], 'parent');
                state.blocks.c3 = makeBlock('c3');

                insertAsLastChild(state, 'c3', 'parent');

                expect(state.blocks.c2.nextSiblingId).toBe('c3');
                expect(state.blocks.c3.prevSiblingId).toBe('c2');
                expect(state.blocks.c3.parentId).toBe('parent');
                expectValid(state);
        });
});

// --- addBlock ---

describe('addBlock', function() {
        it('creates a root block in empty state', function() {
                var state = makeState([], null);
                var id = addBlock(state, 'text');

                expect(state.rootId).toBe(id);
                expect(state.blocks[id].type).toBe('text');
                expect(state.dirty).toBe(true);
                expectValid(state);
        });

        it('appends to end of root list when no afterId', function() {
                var state = rootChain('a');
                var id = addBlock(state, 'heading');

                expect(state.blocks.a.nextSiblingId).toBe(id);
                expect(state.blocks[id].prevSiblingId).toBe('a');
                expect(state.rootId).toBe('a');
                expectValid(state);
        });

        it('inserts after the specified block', function() {
                var state = rootChain('a', 'b');
                var id = addBlock(state, 'text', 'a');

                expect(state.blocks.a.nextSiblingId).toBe(id);
                expect(state.blocks[id].nextSiblingId).toBe('b');
                expect(state.blocks.b.prevSiblingId).toBe(id);
                expectValid(state);
        });
});

// --- addBlockFromDatatype ---

describe('addBlockFromDatatype', function() {
        it('inserts before a target', function() {
                var state = rootChain('a', 'b');
                var dt = { id: 'dt1', label: 'Hero', type: 'container' };
                var id = addBlockFromDatatype(state, dt, 'before', 'b');

                expect(state.blocks[id].type).toBe('container');
                expect(state.blocks[id].label).toBe('Hero');
                expect(state.blocks[id].datatypeId).toBe('dt1');
                expect(state.blocks.a.nextSiblingId).toBe(id);
                expect(state.blocks[id].nextSiblingId).toBe('b');
                expectValid(state);
        });

        it('inserts inside a target as first child', function() {
                var state = makeState([
                        makeBlock('parent', { type: 'container' }),
                ], 'parent');
                var dt = { id: 'dt2', label: 'Card', type: 'text' };
                var id = addBlockFromDatatype(state, dt, 'inside', 'parent');

                expect(state.blocks.parent.firstChildId).toBe(id);
                expect(state.blocks[id].parentId).toBe('parent');
                expectValid(state);
        });

        it('appends to root when no position or target', function() {
                var state = rootChain('a');
                var dt = { id: 'dt3', label: 'Footer', type: 'text' };
                var id = addBlockFromDatatype(state, dt, null, null);

                expect(state.blocks.a.nextSiblingId).toBe(id);
                expectValid(state);
        });
});

// --- removeBlock ---

describe('removeBlock', function() {
        it('removes a leaf block', function() {
                var state = rootChain('a', 'b', 'c');
                var removed = removeBlock(state, 'b');

                expect(removed).toEqual(['b']);
                expect(state.blocks.b).toBeUndefined();
                expect(state.blocks.a.nextSiblingId).toBe('c');
                expect(state.blocks.c.prevSiblingId).toBe('a');
                expectValid(state);
        });

        it('removes a block and all descendants', function() {
                var state = makeState([
                        makeBlock('a', { nextSiblingId: 'b' }),
                        makeBlock('b', { prevSiblingId: 'a', firstChildId: 'c' }),
                        makeBlock('c', { parentId: 'b', firstChildId: 'd' }),
                        makeBlock('d', { parentId: 'c' }),
                ], 'a');

                var removed = removeBlock(state, 'b');

                expect(removed).toContain('b');
                expect(removed).toContain('c');
                expect(removed).toContain('d');
                expect(removed).toHaveLength(3);
                expect(state.blocks.b).toBeUndefined();
                expect(state.blocks.c).toBeUndefined();
                expect(state.blocks.d).toBeUndefined();
                expect(state.blocks.a.nextSiblingId).toBeNull();
                expectValid(state);
        });

        it('removes the only root block', function() {
                var state = rootChain('a');
                removeBlock(state, 'a');

                expect(state.rootId).toBeNull();
                expect(Object.keys(state.blocks)).toHaveLength(0);
                expectValid(state);
        });

        it('updates rootId when removing the first root', function() {
                var state = rootChain('a', 'b');
                removeBlock(state, 'a');

                expect(state.rootId).toBe('b');
                expect(state.blocks.b.prevSiblingId).toBeNull();
                expectValid(state);
        });
});

// --- moveBlock ---

describe('moveBlock', function() {
        it('moves a block before another', function() {
                var state = rootChain('a', 'b', 'c');
                moveBlock(state, 'c', 'a', 'before');

                var roots = getRootList(state);
                expect(roots.map(function(b) { return b.id; })).toEqual(['c', 'a', 'b']);
                expectValid(state);
        });

        it('moves a block after another', function() {
                var state = rootChain('a', 'b', 'c');
                moveBlock(state, 'a', 'c', 'after');

                var roots = getRootList(state);
                expect(roots.map(function(b) { return b.id; })).toEqual(['b', 'c', 'a']);
                expectValid(state);
        });

        it('moves a block inside a container', function() {
                var state = makeState([
                        makeBlock('a', { nextSiblingId: 'container' }),
                        makeBlock('container', { prevSiblingId: 'a', type: 'container' }),
                ], 'a');

                moveBlock(state, 'a', 'container', 'inside');

                expect(state.rootId).toBe('container');
                expect(state.blocks.container.firstChildId).toBe('a');
                expect(state.blocks.a.parentId).toBe('container');
                expectValid(state);
        });

        it('is a no-op when blockId equals targetId', function() {
                var state = rootChain('a', 'b');
                moveBlock(state, 'a', 'a', 'after');

                // State should be unchanged (dirty not set by no-op)
                var roots = getRootList(state);
                expect(roots.map(function(b) { return b.id; })).toEqual(['a', 'b']);
        });
});

// --- indentBlock ---

describe('indentBlock', function() {
        it('makes a block the last child of its previous sibling', function() {
                var state = rootChain('a', 'b');
                // a needs to be a container type to accept children
                state.blocks.a.type = 'container';

                var result = indentBlock(state, 'b');

                expect(result).toBe(true);
                expect(state.blocks.a.firstChildId).toBe('b');
                expect(state.blocks.b.parentId).toBe('a');
                expect(state.rootId).toBe('a');
                expectValid(state);
        });

        it('returns false when there is no previous sibling', function() {
                var state = rootChain('a', 'b');
                var result = indentBlock(state, 'a');
                expect(result).toBe(false);
        });

        it('returns false when previous sibling cannot have children', function() {
                var state = rootChain('a', 'b');
                // type 'text' cannot have children
                state.blocks.a.type = 'text';

                var result = indentBlock(state, 'b');
                expect(result).toBe(false);
        });

        it('appends as last child when prev sibling already has children', function() {
                var state = makeState([
                        makeBlock('a', { type: 'container', firstChildId: 'c1', nextSiblingId: 'b' }),
                        makeBlock('c1', { parentId: 'a' }),
                        makeBlock('b', { prevSiblingId: 'a' }),
                ], 'a');

                indentBlock(state, 'b');

                expect(state.blocks.c1.nextSiblingId).toBe('b');
                expect(state.blocks.b.prevSiblingId).toBe('c1');
                expect(state.blocks.b.parentId).toBe('a');
                expectValid(state);
        });
});

// --- outdentBlock ---

describe('outdentBlock', function() {
        it('moves a child to become next sibling of its parent', function() {
                var state = makeState([
                        makeBlock('parent', { type: 'container', firstChildId: 'child' }),
                        makeBlock('child', { parentId: 'parent' }),
                ], 'parent');

                var result = outdentBlock(state, 'child');

                expect(result).toBe(true);
                expect(state.blocks.parent.nextSiblingId).toBe('child');
                expect(state.blocks.child.parentId).toBeNull();
                expect(state.blocks.parent.firstChildId).toBeNull();
                expectValid(state);
        });

        it('returns false when block is at root level', function() {
                var state = rootChain('a');
                var result = outdentBlock(state, 'a');
                expect(result).toBe(false);
        });

        it('reparents younger siblings under the outdented block', function() {
                var state = makeState([
                        makeBlock('parent', { type: 'container', firstChildId: 'c1' }),
                        makeBlock('c1', { parentId: 'parent', nextSiblingId: 'c2' }),
                        makeBlock('c2', { parentId: 'parent', prevSiblingId: 'c1', nextSiblingId: 'c3' }),
                        makeBlock('c3', { parentId: 'parent', prevSiblingId: 'c2' }),
                ], 'parent');

                outdentBlock(state, 'c2');

                // c2 is now after parent at root level
                expect(state.blocks.parent.nextSiblingId).toBe('c2');
                expect(state.blocks.c2.parentId).toBeNull();
                // c3 was reparented under c2
                expect(state.blocks.c2.firstChildId).toBe('c3');
                expect(state.blocks.c3.parentId).toBe('c2');
                // c1 stays with parent
                expect(state.blocks.parent.firstChildId).toBe('c1');
                expectValid(state);
        });
});

// --- duplicateBlock ---

describe('duplicateBlock', function() {
        it('duplicates a leaf block with a new ID', function() {
                var state = rootChain('a', 'b');
                var cloneId = duplicateBlock(state, 'a');

                expect(cloneId).not.toBe('a');
                expect(state.blocks[cloneId].type).toBe('text');
                expect(state.blocks[cloneId].label).toBe('Block a');
                expect(state.blocks.a.nextSiblingId).toBe(cloneId);
                expect(state.blocks[cloneId].nextSiblingId).toBe('b');
                expectValid(state);
        });

        it('deep-clones a subtree with new IDs', function() {
                var state = makeState([
                        makeBlock('root', { type: 'container', firstChildId: 'c1' }),
                        makeBlock('c1', { parentId: 'root', firstChildId: 'gc1' }),
                        makeBlock('gc1', { parentId: 'c1' }),
                ], 'root');

                var cloneId = duplicateBlock(state, 'root');
                var clone = state.blocks[cloneId];

                expect(cloneId).not.toBe('root');
                expect(clone.firstChildId).not.toBeNull();
                expect(clone.firstChildId).not.toBe('c1');

                // Verify clone child exists and has its own grandchild
                var cloneChild = state.blocks[clone.firstChildId];
                expect(cloneChild.parentId).toBe(cloneId);
                expect(cloneChild.firstChildId).not.toBeNull();
                expect(cloneChild.firstChildId).not.toBe('gc1');

                var cloneGrandchild = state.blocks[cloneChild.firstChildId];
                expect(cloneGrandchild.parentId).toBe(cloneChild.id);
                expectValid(state);
        });

        it('returns null for nonexistent block', function() {
                var state = rootChain('a');
                var result = duplicateBlock(state, 'nonexistent');
                expect(result).toBeNull();
        });
});

// --- moveBlockUp / moveBlockDown ---

describe('moveBlockUp', function() {
        it('swaps with previous sibling', function() {
                var state = rootChain('a', 'b', 'c');
                var result = moveBlockUp(state, 'b');

                expect(result).toBe(true);
                var roots = getRootList(state);
                expect(roots.map(function(b) { return b.id; })).toEqual(['b', 'a', 'c']);
                expectValid(state);
        });

        it('returns false when already first', function() {
                var state = rootChain('a', 'b');
                var result = moveBlockUp(state, 'a');
                expect(result).toBe(false);
        });

        it('updates rootId when moving to first position', function() {
                var state = rootChain('a', 'b');
                moveBlockUp(state, 'b');

                expect(state.rootId).toBe('b');
                expectValid(state);
        });
});

describe('moveBlockDown', function() {
        it('swaps with next sibling', function() {
                var state = rootChain('a', 'b', 'c');
                var result = moveBlockDown(state, 'b');

                expect(result).toBe(true);
                var roots = getRootList(state);
                expect(roots.map(function(b) { return b.id; })).toEqual(['a', 'c', 'b']);
                expectValid(state);
        });

        it('returns false when already last', function() {
                var state = rootChain('a', 'b');
                var result = moveBlockDown(state, 'b');
                expect(result).toBe(false);
        });
});

// --- Complex scenarios ---

describe('complex operations', function() {
        it('survives a sequence of add, move, indent, outdent, delete', function() {
                var state = makeState([], null);

                // Add three blocks
                var id1 = addBlock(state, 'container');
                var id2 = addBlock(state, 'text');
                var id3 = addBlock(state, 'heading');

                expectValid(state);
                expect(getRootList(state)).toHaveLength(3);

                // Indent id2 into id1
                indentBlock(state, id2);
                expectValid(state);
                expect(state.blocks[id2].parentId).toBe(id1);

                // Move id3 inside id1 (as first child, before id2)
                moveBlock(state, id3, id2, 'before');
                expectValid(state);
                expect(getChildren(state, id1).map(function(b) { return b.id; })).toEqual([id3, id2]);

                // Outdent id3 from id1
                outdentBlock(state, id3);
                expectValid(state);
                expect(state.blocks[id3].parentId).toBeNull();
                // id2 was a younger sibling, should be reparented under id3
                expect(state.blocks[id2].parentId).toBe(id3);

                // Delete id3 (and its child id2)
                var removed = removeBlock(state, id3);
                expectValid(state);
                expect(removed).toContain(id3);
                expect(removed).toContain(id2);
                expect(Object.keys(state.blocks)).toHaveLength(1);
                expect(state.rootId).toBe(id1);
        });

        it('duplicate then move preserves structure', function() {
                var state = makeState([
                        makeBlock('a', { type: 'container', firstChildId: 'b', nextSiblingId: 'c' }),
                        makeBlock('b', { parentId: 'a' }),
                        makeBlock('c', { prevSiblingId: 'a' }),
                ], 'a');

                var cloneId = duplicateBlock(state, 'a');
                expectValid(state);

                // Move clone inside original
                moveBlock(state, cloneId, 'a', 'inside');
                expectValid(state);

                expect(state.blocks[cloneId].parentId).toBe('a');
                // a should now have b and clone as children
                var children = getChildren(state, 'a');
                expect(children.length).toBeGreaterThanOrEqual(2);
        });
});
