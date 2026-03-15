import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { History } from './history.js';

function makeState(overrides) {
        return {
                blocks: {
                        'a': { id: 'a', parentId: null, firstChildId: 'b', nextSiblingId: null, prevSiblingId: null, fields: [{ fieldId: 'f1', value: 'hello' }] },
                        'b': { id: 'b', parentId: 'a', firstChildId: null, nextSiblingId: null, prevSiblingId: null, fields: [{ fieldId: 'f2', value: 42 }] },
                },
                rootId: 'a',
                selectedBlockId: 'a',
                dirty: false,
                ...overrides,
        };
}

describe('History', () => {
        beforeEach(() => {
                vi.useFakeTimers();
        });

        afterEach(() => {
                vi.useRealTimers();
        });

        it('pushUndo/popUndo symmetry — LIFO order', () => {
                var h = new History(50);
                var s1 = makeState({ selectedBlockId: 'a' });
                var s2 = makeState({ selectedBlockId: 'b' });
                var s3 = makeState({ selectedBlockId: null });

                h.pushUndo(s1);
                h.pushUndo(s2);
                h.pushUndo(s3);

                var current = makeState();
                var e3 = h.popUndo(current);
                expect(e3.selectedBlockId).toBe(null);

                var e2 = h.popUndo(current);
                expect(e2.selectedBlockId).toBe('b');

                var e1 = h.popUndo(current);
                expect(e1.selectedBlockId).toBe('a');

                expect(h.canUndo).toBe(false);
        });

        it('popUndo pushes to redo stack, popRedo pushes to undo stack', () => {
                var h = new History(50);
                var s1 = makeState({ selectedBlockId: 'a' });
                h.pushUndo(s1);

                var current = makeState({ selectedBlockId: 'b' });
                var undoEntry = h.popUndo(current);

                expect(h.canUndo).toBe(false);
                expect(h.canRedo).toBe(true);

                var redoEntry = h.popRedo(makeState());
                expect(redoEntry.selectedBlockId).toBe('b');
                expect(h.canUndo).toBe(true);
                expect(h.canRedo).toBe(false);
        });

        it('max size eviction — oldest entry removed when exceeding maxSize', () => {
                var h = new History(50);
                for (var i = 0; i < 51; i++) {
                        h.pushUndo(makeState({ selectedBlockId: String(i) }));
                }
                // Should have 50, not 51
                var count = 0;
                var current = makeState();
                while (h.canUndo) {
                        h.popUndo(current);
                        count++;
                }
                expect(count).toBe(50);
        });

        it('clear() resets both stacks', () => {
                var h = new History(50);
                h.pushUndo(makeState());
                h.pushUndo(makeState());
                // Pop one to redo
                h.popUndo(makeState());

                expect(h.canUndo).toBe(true);
                expect(h.canRedo).toBe(true);

                h.clear();
                expect(h.canUndo).toBe(false);
                expect(h.canRedo).toBe(false);
        });

        it('discardLastUndo() removes the most recent entry', () => {
                var h = new History(50);
                h.pushUndo(makeState({ selectedBlockId: 'a' }));
                h.pushUndo(makeState({ selectedBlockId: 'b' }));

                h.discardLastUndo();

                var entry = h.popUndo(makeState());
                expect(entry.selectedBlockId).toBe('a');
                expect(h.canUndo).toBe(false);
        });

        it('remapIds() rewrites block map keys, block.id, pointer fields, rootId, and selectedBlockId', () => {
                var h = new History(50);
                var state = makeState();
                h.pushUndo(state);

                var idMap = { 'a': 'A1', 'b': 'B1' };
                h.remapIds(idMap);

                var entry = h.popUndo(makeState());
                var blocks = entry.snapshot.blocks;

                // Old keys should be gone
                expect(blocks['a']).toBeUndefined();
                expect(blocks['b']).toBeUndefined();

                // New keys should exist
                expect(blocks['A1']).toBeDefined();
                expect(blocks['B1']).toBeDefined();

                // block.id remapped
                expect(blocks['A1'].id).toBe('A1');
                expect(blocks['B1'].id).toBe('B1');

                // Pointer fields remapped
                expect(blocks['A1'].firstChildId).toBe('B1');
                expect(blocks['B1'].parentId).toBe('A1');

                // rootId remapped
                expect(entry.snapshot.rootId).toBe('A1');

                // selectedBlockId remapped
                expect(entry.selectedBlockId).toBe('A1');
        });

        it('field batching — pushFieldChange within 500ms window does not create additional snapshots', () => {
                var h = new History(50);
                var state = makeState();

                h.pushFieldChange(state);
                expect(h.canUndo).toBe(true);

                // Within 500ms, subsequent calls are no-ops
                vi.advanceTimersByTime(200);
                h.pushFieldChange(state);
                h.pushFieldChange(state);

                // Should still have only 1 entry
                var count = 0;
                var current = makeState();
                while (h.canUndo) {
                        h.popUndo(current);
                        count++;
                }
                expect(count).toBe(1);
        });

        it('field batching — after 500ms, a new call creates a new snapshot', () => {
                var h = new History(50);
                var state = makeState();

                h.pushFieldChange(state);
                vi.advanceTimersByTime(500);

                h.pushFieldChange(state);

                var count = 0;
                var current = makeState();
                while (h.canUndo) {
                        h.popUndo(current);
                        count++;
                }
                expect(count).toBe(2);
        });

        it('pushUndo resets field batch timer', () => {
                var h = new History(50);
                var state = makeState();

                h.pushFieldChange(state);
                expect(h.inFieldBatch).toBe(true);

                // Structural mutation resets batch
                h.pushUndo(state);
                expect(h.inFieldBatch).toBe(false);

                // Next field change should create a new entry
                h.pushFieldChange(state);
                var count = 0;
                var current = makeState();
                while (h.canUndo) {
                        h.popUndo(current);
                        count++;
                }
                expect(count).toBe(3);
        });

        it('mutation isolation — snapshot is independent from original state', () => {
                var h = new History(50);
                var state = makeState();

                h.pushUndo(state);

                // Mutate the original state
                state.blocks['a'].fields[0].value = 'mutated';
                state.blocks['c'] = { id: 'c', parentId: null, firstChildId: null, nextSiblingId: null, prevSiblingId: null };

                // Pop the snapshot and verify isolation
                var entry = h.popUndo(makeState());
                expect(entry.snapshot.blocks['a'].fields[0].value).toBe('hello');
                expect(entry.snapshot.blocks['c']).toBeUndefined();
        });

        it('pushUndo clears redo stack', () => {
                var h = new History(50);
                h.pushUndo(makeState());
                h.popUndo(makeState());
                expect(h.canRedo).toBe(true);

                h.pushUndo(makeState());
                expect(h.canRedo).toBe(false);
        });

        it('remapIds works on redo stack too', () => {
                var h = new History(50);
                var state = makeState();
                h.pushUndo(state);
                h.popUndo(makeState());
                expect(h.canRedo).toBe(true);

                h.remapIds({ 'a': 'X', 'b': 'Y' });

                var entry = h.popRedo(makeState());
                expect(entry.snapshot.blocks['X']).toBeDefined();
                expect(entry.snapshot.blocks['Y']).toBeDefined();
                expect(entry.snapshot.rootId).toBe('X');
        });
});
