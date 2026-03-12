import { describe, it, expect } from 'vitest';
import { validateState } from './validate.js';

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

// --- Valid states ---

describe('validateState with valid states', function() {
        it('passes for empty state', function() {
                expect(validateState(makeState([], null))).toEqual([]);
        });

        it('passes for single root block', function() {
                var state = makeState([makeBlock('a')], 'a');
                expect(validateState(state)).toEqual([]);
        });

        it('passes for root chain', function() {
                var state = makeState([
                        makeBlock('a', { nextSiblingId: 'b' }),
                        makeBlock('b', { prevSiblingId: 'a', nextSiblingId: 'c' }),
                        makeBlock('c', { prevSiblingId: 'b' }),
                ], 'a');
                expect(validateState(state)).toEqual([]);
        });

        it('passes for parent-child tree', function() {
                var state = makeState([
                        makeBlock('parent', { firstChildId: 'c1' }),
                        makeBlock('c1', { parentId: 'parent', nextSiblingId: 'c2' }),
                        makeBlock('c2', { parentId: 'parent', prevSiblingId: 'c1' }),
                ], 'parent');
                expect(validateState(state)).toEqual([]);
        });

        it('passes for deeply nested tree', function() {
                var state = makeState([
                        makeBlock('a', { firstChildId: 'b' }),
                        makeBlock('b', { parentId: 'a', firstChildId: 'c' }),
                        makeBlock('c', { parentId: 'b', firstChildId: 'd' }),
                        makeBlock('d', { parentId: 'c' }),
                ], 'a');
                expect(validateState(state)).toEqual([]);
        });
});

// --- Broken pointers ---

describe('validateState detects broken pointers', function() {
        it('detects broken nextSiblingId reciprocal', function() {
                var state = makeState([
                        makeBlock('a', { nextSiblingId: 'b' }),
                        makeBlock('b'), // prevSiblingId should be 'a' but is null
                ], 'a');
                var errors = validateState(state);
                expect(errors.length).toBeGreaterThan(0);
                expect(errors.some(function(e) { return e.includes('nextSiblingId'); })).toBe(true);
        });

        it('detects broken prevSiblingId reciprocal', function() {
                var state = makeState([
                        makeBlock('a'), // nextSiblingId should be 'b' but is null
                        makeBlock('b', { prevSiblingId: 'a' }),
                ], 'a');
                var errors = validateState(state);
                expect(errors.length).toBeGreaterThan(0);
                expect(errors.some(function(e) { return e.includes('prevSiblingId'); })).toBe(true);
        });

        it('detects nextSiblingId pointing to nonexistent block', function() {
                var state = makeState([
                        makeBlock('a', { nextSiblingId: 'ghost' }),
                ], 'a');
                var errors = validateState(state);
                expect(errors.length).toBeGreaterThan(0);
                expect(errors.some(function(e) { return e.includes('not found'); })).toBe(true);
        });

        it('detects firstChildId pointing to nonexistent block', function() {
                var state = makeState([
                        makeBlock('a', { firstChildId: 'ghost' }),
                ], 'a');
                var errors = validateState(state);
                expect(errors.length).toBeGreaterThan(0);
        });

        it('detects firstChild with wrong parentId', function() {
                var state = makeState([
                        makeBlock('parent', { firstChildId: 'child' }),
                        makeBlock('child', { parentId: 'other' }), // wrong parent
                        makeBlock('other'),
                ], 'parent');
                var errors = validateState(state);
                expect(errors.length).toBeGreaterThan(0);
                expect(errors.some(function(e) { return e.includes('parentId'); })).toBe(true);
        });
});

// --- Orphaned blocks ---

describe('validateState detects orphaned blocks', function() {
        it('detects a block not reachable from rootId', function() {
                var state = makeState([
                        makeBlock('a'),
                        makeBlock('orphan'), // not linked to anything
                ], 'a');
                var errors = validateState(state);
                expect(errors.length).toBeGreaterThan(0);
                expect(errors.some(function(e) { return e.includes('not reachable'); })).toBe(true);
        });
});

// --- Root chain consistency ---

describe('validateState detects root chain issues', function() {
        it('detects root block with non-null parentId', function() {
                var state = makeState([
                        makeBlock('a', { parentId: 'someone', nextSiblingId: 'b' }),
                        makeBlock('b', { prevSiblingId: 'a' }),
                ], 'a');
                var errors = validateState(state);
                expect(errors.length).toBeGreaterThan(0);
                expect(errors.some(function(e) { return e.includes('parentId'); })).toBe(true);
        });

        it('detects rootId pointing to nonexistent block', function() {
                var state = makeState([], 'ghost');
                var errors = validateState(state);
                expect(errors.length).toBeGreaterThan(0);
        });
});

// --- Child chain consistency ---

describe('validateState detects child chain issues', function() {
        it('detects child with wrong parentId in child chain', function() {
                var state = makeState([
                        makeBlock('parent', { firstChildId: 'c1' }),
                        makeBlock('c1', { parentId: 'parent', nextSiblingId: 'c2' }),
                        makeBlock('c2', { prevSiblingId: 'c1', parentId: 'wrong' }), // wrong parent
                ], 'parent');
                var errors = validateState(state);
                expect(errors.length).toBeGreaterThan(0);
        });
});
