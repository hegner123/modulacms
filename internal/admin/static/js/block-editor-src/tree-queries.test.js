import { describe, it, expect } from 'vitest';
import {
        getChildren, getRootList, isDescendant, getDepth, findLastSibling,
        getBlockTraversalOrder, collectDescendants, getDescendantCount,
} from './tree-queries.js';

// --- Helpers ---

function makeBlock(id, overrides) {
        return {
                id, type: 'text',
                parentId: null, firstChildId: null,
                prevSiblingId: null, nextSiblingId: null,
                label: 'Block ' + id,
                ...overrides,
        };
}

function makeState(blocks, rootId) {
        var state = { blocks: {}, rootId: rootId || null, selectedBlockId: null, dirty: false };
        for (var b of blocks) state.blocks[b.id] = b;
        return state;
}

// --- Test states ---

// Empty state
function emptyState() {
        return makeState([], null);
}

// Single root: a
function singleRoot() {
        return makeState([makeBlock('a')], 'a');
}

// Root chain: a -> b -> c
function rootChain() {
        return makeState([
                makeBlock('a', { nextSiblingId: 'b' }),
                makeBlock('b', { prevSiblingId: 'a', nextSiblingId: 'c' }),
                makeBlock('c', { prevSiblingId: 'b' }),
        ], 'a');
}

// Tree structure:
//   a
//   ├── b
//   │   └── d
//   └── c
//   e (second root)
function treeState() {
        return makeState([
                makeBlock('a', { firstChildId: 'b', nextSiblingId: 'e' }),
                makeBlock('b', { parentId: 'a', nextSiblingId: 'c', firstChildId: 'd' }),
                makeBlock('c', { parentId: 'a', prevSiblingId: 'b' }),
                makeBlock('d', { parentId: 'b' }),
                makeBlock('e', { prevSiblingId: 'a' }),
        ], 'a');
}

// --- getChildren ---

describe('getChildren', function() {
        it('returns empty array for block with no children', function() {
                var state = singleRoot();
                expect(getChildren(state, 'a')).toEqual([]);
        });

        it('returns empty array for nonexistent block', function() {
                var state = singleRoot();
                expect(getChildren(state, 'nonexistent')).toEqual([]);
        });

        it('returns children in sibling order', function() {
                var state = treeState();
                var children = getChildren(state, 'a');
                expect(children.map(function(b) { return b.id; })).toEqual(['b', 'c']);
        });

        it('returns single child', function() {
                var state = treeState();
                var children = getChildren(state, 'b');
                expect(children.map(function(b) { return b.id; })).toEqual(['d']);
        });
});

// --- getRootList ---

describe('getRootList', function() {
        it('returns empty array for empty state', function() {
                expect(getRootList(emptyState())).toEqual([]);
        });

        it('returns single root', function() {
                var roots = getRootList(singleRoot());
                expect(roots.map(function(b) { return b.id; })).toEqual(['a']);
        });

        it('returns all roots in order', function() {
                var roots = getRootList(rootChain());
                expect(roots.map(function(b) { return b.id; })).toEqual(['a', 'b', 'c']);
        });

        it('returns roots excluding children', function() {
                var roots = getRootList(treeState());
                expect(roots.map(function(b) { return b.id; })).toEqual(['a', 'e']);
        });
});

// --- isDescendant ---

describe('isDescendant', function() {
        it('returns true for direct child', function() {
                var state = treeState();
                expect(isDescendant(state, 'b', 'a')).toBe(true);
        });

        it('returns true for deep descendant', function() {
                var state = treeState();
                expect(isDescendant(state, 'd', 'a')).toBe(true);
        });

        it('returns false for non-descendant', function() {
                var state = treeState();
                expect(isDescendant(state, 'e', 'a')).toBe(false);
        });

        it('returns false for self', function() {
                var state = treeState();
                expect(isDescendant(state, 'a', 'a')).toBe(false);
        });

        it('returns false for ancestor (reverse direction)', function() {
                var state = treeState();
                expect(isDescendant(state, 'a', 'b')).toBe(false);
        });

        it('returns false for sibling', function() {
                var state = treeState();
                expect(isDescendant(state, 'c', 'b')).toBe(false);
        });
});

// --- getDepth ---

describe('getDepth', function() {
        it('returns 0 for root blocks', function() {
                var state = treeState();
                expect(getDepth(state, 'a')).toBe(0);
                expect(getDepth(state, 'e')).toBe(0);
        });

        it('returns 1 for direct children', function() {
                var state = treeState();
                expect(getDepth(state, 'b')).toBe(1);
                expect(getDepth(state, 'c')).toBe(1);
        });

        it('returns 2 for grandchildren', function() {
                var state = treeState();
                expect(getDepth(state, 'd')).toBe(2);
        });
});

// --- findLastSibling ---

describe('findLastSibling', function() {
        it('returns the block itself when no next sibling', function() {
                var state = singleRoot();
                expect(findLastSibling(state, 'a')).toBe('a');
        });

        it('walks to the end of the sibling chain', function() {
                var state = rootChain();
                expect(findLastSibling(state, 'a')).toBe('c');
        });

        it('returns the start when it is already the last', function() {
                var state = rootChain();
                expect(findLastSibling(state, 'c')).toBe('c');
        });
});

// --- getBlockTraversalOrder ---

describe('getBlockTraversalOrder', function() {
        it('returns empty array for empty state', function() {
                expect(getBlockTraversalOrder(emptyState())).toEqual([]);
        });

        it('returns single block', function() {
                expect(getBlockTraversalOrder(singleRoot())).toEqual(['a']);
        });

        it('returns flat list for root chain', function() {
                expect(getBlockTraversalOrder(rootChain())).toEqual(['a', 'b', 'c']);
        });

        it('returns DFS order for tree', function() {
                // Tree: a(b(d), c), e
                // DFS: a, b, d, c, e
                var order = getBlockTraversalOrder(treeState());
                expect(order).toEqual(['a', 'b', 'd', 'c', 'e']);
        });
});

// --- collectDescendants ---

describe('collectDescendants', function() {
        it('returns empty array for leaf block', function() {
                var state = singleRoot();
                expect(collectDescendants(state, 'a')).toEqual([]);
        });

        it('returns all descendants in DFS order', function() {
                var state = treeState();
                var desc = collectDescendants(state, 'a');
                expect(desc).toContain('b');
                expect(desc).toContain('c');
                expect(desc).toContain('d');
                expect(desc).toHaveLength(3);
        });

        it('returns only subtree descendants', function() {
                var state = treeState();
                var desc = collectDescendants(state, 'b');
                expect(desc).toEqual(['d']);
        });

        it('returns empty for nonexistent block', function() {
                var state = singleRoot();
                expect(collectDescendants(state, 'nonexistent')).toEqual([]);
        });
});

// --- getDescendantCount ---

describe('getDescendantCount', function() {
        it('returns 0 for leaf', function() {
                var state = singleRoot();
                expect(getDescendantCount(state, 'a')).toBe(0);
        });

        it('matches collectDescendants length', function() {
                var state = treeState();
                expect(getDescendantCount(state, 'a')).toBe(collectDescendants(state, 'a').length);
                expect(getDescendantCount(state, 'b')).toBe(collectDescendants(state, 'b').length);
        });

        it('counts nested descendants', function() {
                var state = treeState();
                expect(getDescendantCount(state, 'a')).toBe(3); // b, c, d
                expect(getDescendantCount(state, 'b')).toBe(1); // d
        });
});
