// dom-patches.js — DOM mutation helper methods (prototype mixin)
//
// Methods in this object are assigned to BlockEditor.prototype via Object.assign.
// They access the class instance via `this` (e.g., this._state, this._wrapperRegistry,
// this._elementRegistry, this._renderBlockWrapper, this._selectBlock, this._devValidate,
// this._updateSaveButton, this.dispatchEvent).

import { indentBlock, outdentBlock, duplicateBlock, moveBlockUp, moveBlockDown, addBlock } from './tree-ops.js';
import { getDepth, getChildren, getDescendantCount } from './tree-queries.js';
import { validateState } from './validate.js';

export const domPatchMethods = {
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

                this._history.pushUndo(this._state);
                const result = indentBlock(this._state, blockId);
                if (!result) { this._history.discardLastUndo(); return; }

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
                        composed: true,
                        detail: { action: 'indent', blockId },
                }));
        },

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

                this._history.pushUndo(this._state);
                const result = outdentBlock(this._state, blockId);
                if (!result) { this._history.discardLastUndo(); return; }

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
                        composed: true,
                        detail: { action: 'outdent', blockId },
                }));
        },

        /**
         * Duplicate a block and its entire subtree. Renders the cloned subtree
         * and inserts it into the DOM after the original's wrapper.
         */
        _doDuplicateBlock(blockId) {
                if (!this._state) return;
                if (!blockId) return;
                const block = this._state.blocks[blockId];
                if (!block) return;

                this._history.pushUndo(this._state);
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
                        composed: true,
                        detail: { action: 'duplicate', blockId },
                }));
        },

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

                this._history.pushUndo(this._state);
                const result = moveBlockUp(this._state, blockId);
                if (!result) { this._history.discardLastUndo(); return; }

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
                        composed: true,
                        detail: { action: 'moveUp', blockId },
                }));
        },

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

                this._history.pushUndo(this._state);
                const result = moveBlockDown(this._state, blockId);
                if (!result) { this._history.discardLastUndo(); return; }

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
                        composed: true,
                        detail: { action: 'moveDown', blockId },
                }));
        },

        /**
         * Add a new block after the specified block (or at end of root list).
         * Used by Enter key and toolbar. Default type is 'text'.
         */
        _doAddBlockAfter(afterBlockId, type) {
                if (!this._state) return;
                if (!afterBlockId) return;

                const afterBlock = this._state.blocks[afterBlockId];
                if (!afterBlock) return;

                this._history.pushUndo(this._state);
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
                        composed: true,
                        detail: { action: 'add', blockId: id },
                }));
        },

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
        },

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
        },

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
        },
};
