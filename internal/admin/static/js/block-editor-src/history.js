// history.js — Undo/redo stack manager for block editor
//
// Snapshot-based: clones the full { blocks, rootId } state before each mutation.
// Field edits are batched into 500ms windows to avoid one undo entry per keystroke.

export class History {
        constructor(maxSize) {
                this._maxSize = maxSize || 50;
                this._undoStack = [];
                this._redoStack = [];
                this._fieldBatchTimer = null;
                this._inFieldBatch = false;
        }

        _cloneState(state) {
                var blocks = {};
                var keys = Object.keys(state.blocks);
                for (var i = 0; i < keys.length; i++) {
                        var key = keys[i];
                        var block = state.blocks[key];
                        blocks[key] = {
                                ...block,
                                fields: block.fields ? block.fields.map(function(f) { return { ...f }; }) : undefined,
                        };
                }
                return { blocks: blocks, rootId: state.rootId };
        }

        pushUndo(state) {
                clearTimeout(this._fieldBatchTimer);
                this._inFieldBatch = false;

                var entry = {
                        snapshot: this._cloneState(state),
                        selectedBlockId: state.selectedBlockId,
                };
                this._undoStack.push(entry);
                this._redoStack = [];
                if (this._undoStack.length > this._maxSize) {
                        this._undoStack.shift();
                }
        }

        discardLastUndo() {
                this._undoStack.pop();
        }

        pushFieldChange(state) {
                if (this._inFieldBatch) return;
                this.pushUndo(state);
                this._inFieldBatch = true;
                var self = this;
                this._fieldBatchTimer = setTimeout(function() {
                        self._inFieldBatch = false;
                }, 500);
        }

        get inFieldBatch() {
                return this._inFieldBatch;
        }

        popUndo(currentState) {
                if (this._undoStack.length === 0) return null;
                var redoEntry = {
                        snapshot: this._cloneState(currentState),
                        selectedBlockId: currentState.selectedBlockId,
                };
                this._redoStack.push(redoEntry);
                return this._undoStack.pop();
        }

        popRedo(currentState) {
                if (this._redoStack.length === 0) return null;
                var undoEntry = {
                        snapshot: this._cloneState(currentState),
                        selectedBlockId: currentState.selectedBlockId,
                };
                this._undoStack.push(undoEntry);
                return this._redoStack.pop();
        }

        get canUndo() {
                return this._undoStack.length > 0;
        }

        get canRedo() {
                return this._redoStack.length > 0;
        }

        clear() {
                this._undoStack = [];
                this._redoStack = [];
                clearTimeout(this._fieldBatchTimer);
                this._inFieldBatch = false;
        }

        remapIds(idMap) {
                var clientIds = Object.keys(idMap);
                var stacks = [this._undoStack, this._redoStack];
                for (var s = 0; s < stacks.length; s++) {
                        var stack = stacks[s];
                        for (var e = 0; e < stack.length; e++) {
                                var entry = stack[e];
                                var snapshot = entry.snapshot;

                                // Rekey blocks map for remapped IDs
                                for (var c = 0; c < clientIds.length; c++) {
                                        var clientId = clientIds[c];
                                        if (snapshot.blocks[clientId] !== undefined) {
                                                var block = snapshot.blocks[clientId];
                                                delete snapshot.blocks[clientId];
                                                block.id = idMap[clientId];
                                                snapshot.blocks[idMap[clientId]] = block;
                                        }
                                }

                                // Remap pointer fields on ALL blocks
                                var blockKeys = Object.keys(snapshot.blocks);
                                for (var b = 0; b < blockKeys.length; b++) {
                                        var blk = snapshot.blocks[blockKeys[b]];
                                        if (blk.parentId && idMap[blk.parentId] !== undefined) {
                                                blk.parentId = idMap[blk.parentId];
                                        }
                                        if (blk.firstChildId && idMap[blk.firstChildId] !== undefined) {
                                                blk.firstChildId = idMap[blk.firstChildId];
                                        }
                                        if (blk.nextSiblingId && idMap[blk.nextSiblingId] !== undefined) {
                                                blk.nextSiblingId = idMap[blk.nextSiblingId];
                                        }
                                        if (blk.prevSiblingId && idMap[blk.prevSiblingId] !== undefined) {
                                                blk.prevSiblingId = idMap[blk.prevSiblingId];
                                        }
                                }

                                // Remap rootId
                                if (snapshot.rootId && idMap[snapshot.rootId] !== undefined) {
                                        snapshot.rootId = idMap[snapshot.rootId];
                                }

                                // Remap selectedBlockId
                                if (entry.selectedBlockId && idMap[entry.selectedBlockId] !== undefined) {
                                        entry.selectedBlockId = idMap[entry.selectedBlockId];
                                }
                        }
                }
        }
}
