// drag.js — Drag-and-drop methods (prototype mixin)
//
// Methods in this object are assigned to BlockEditor.prototype via Object.assign.
// They access the class instance via `this` (e.g., this._state, this._drag,
// this._wrapperRegistry, this._elementRegistry,
// this._updateDescendantDepths, this._cleanupEmptyChildrenContainer,
// this._updateChildCountBadge, this._selectBlock, this._devValidate,
// this._updateSaveButton, this._renderBlockWrapper, this._renderChildrenInto).

import { moveBlock } from './tree-ops.js';
import { isDescendant, getDepth, getChildren } from './tree-queries.js';
import { MAX_DEPTH } from './config.js';

export const dragMethods = {
        _onPointerDown(e) {
                // Only primary button (left click / touch)
                if (e.button !== 0) return;

                // Don't drag when clicking delete button or toolbar buttons
                if (e.target.closest('[data-action]')) {
                        console.log('[kebab-debug] pointerdown skipped drag — target has [data-action]:', e.target.closest('[data-action]').dataset.action, 'blockId:', e.target.closest('[data-action]').dataset.blockId);
                        return;
                }

                // Find the block header (.block-item) that was clicked
                const blockItem = e.target.closest('.block-item');
                if (!blockItem) return;

                const blockId = blockItem.dataset.blockId;
                if (!blockId) return;

                // Content preview clicks: select immediately, no drag
                if (e.target.closest('.block-content-preview')) {
                        this._selectBlock(blockId);
                        return;
                }

                // Allow dragging ALL blocks (root and nested)
                const block = this._state?.blocks[blockId];
                if (!block) return;

                // Guard focus auto-select from firing during this pointer interaction
                this._pointerSelectActive = true;

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
                        this._pointerSelectActive = false;
                        this._selectBlock(blockId);
                };

                blockItem.addEventListener('pointermove', onPreMove);
                blockItem.addEventListener('pointerup', onPreUp);
        },

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
        },

        _startDrag(e, blockItem, blockId, startX, startY) {
                try {
                        // Capture pointer on the block header
                        blockItem.setPointerCapture(e.pointerId);

                        // Compute grab offset from the block's top-left corner
                        const blockRect = blockItem.getBoundingClientRect();
                        const grabOffsetX = startX - blockRect.left;
                        // NOTE: The overlay uses position:fixed + transform, so coordinates are
                        // viewport-relative. clientX/Y minus the grab offset places the overlay
                        // at the correct position under the pointer.
                        const grabOffsetY = startY - blockRect.top;

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

                        // Handle lost pointer capture (element removed from DOM during drag,
                        // e.g. by a concurrent undo/redo triggering _render()).
                        const onLostCapture = () => {
                                blockItem.removeEventListener('lostpointercapture', onLostCapture);
                                this._cleanupDrag();
                        };

                        // Attach drag-mode handlers to the block header (capture routes all events here)
                        blockItem.addEventListener('pointermove', onDragMove);
                        blockItem.addEventListener('pointerup', onDragEnd);
                        blockItem.addEventListener('pointercancel', onDragEnd);
                        blockItem.addEventListener('lostpointercapture', onLostCapture);
                } catch (err) {
                        console.error('[block-editor] Error starting drag:', err);
                        this._cleanupDrag();
                }
        },

        _createDragOverlay(blockItem, clientX, clientY, grabOffsetX, grabOffsetY) {
                const overlay = blockItem.cloneNode(true);
                overlay.className = 'block-item drag-overlay';
                // Match the original width so it does not collapse
                overlay.style.width = blockItem.getBoundingClientRect().width + 'px';
                overlay.style.transform = 'translate(' + (clientX - grabOffsetX) + 'px, ' + (clientY - grabOffsetY) + 'px)';

                // Append to the block-editor element
                this.appendChild(overlay);
                return overlay;
        },

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
        },

        /**
         * Three-zone drop detection against block header rects (not full subtree).
         * Works on ALL blocks (root and nested).
         * beforeZone = Math.max(rect.height * 0.25, 10)
         * afterZone = Math.max(rect.height * 0.25, 10)
         * relativeY in beforeZone -> "before", in afterZone -> "after", else -> "inside"
         *
         * Coercions:
         * - "inside" coerced to "after" if canHaveChildren is false
         * - "inside" coerced to "after" if target depth + 1 > MAX_DEPTH
         * - Skip entirely if target is descendant of dragged block
         */
        _computeDropZone(clientX, clientY) {
                if (!this._drag) return null;

                // Get ALL block-item headers in the shadow DOM
                const allBlockItems = this.querySelectorAll('.block-item');
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

                                // Coerce "inside" to "after" if max nesting depth would be exceeded
                                if (position === 'inside') {
                                        const targetDepth = getDepth(this._state, itemBlockId);
                                        if (targetDepth + 1 > MAX_DEPTH) {
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
        },

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

                const blockList = this.querySelector('.block-list');
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
        },

        _applyDropInsideHighlight(blockId) {
                // Remove previous highlight
                this._removeDropInsideHighlight();

                const targetEl = this._elementRegistry.get(blockId);
                if (targetEl) {
                        targetEl.classList.add('drop-inside');
                }
        },

        _removeDropInsideHighlight() {
                const highlighted = this.querySelector('.drop-inside');
                if (highlighted) {
                        highlighted.classList.remove('drop-inside');
                }
        },

        _removeDropIndicator() {
                if (this._dropIndicator) {
                        this._dropIndicator.remove();
                        this._dropIndicator = null;
                }
        },

        _onDragEnd(e) {
                if (!this._drag) return;

                const { dropTarget } = this._drag;

                if (dropTarget) {
                        this._executeDrop(dropTarget);
                }

                this._cleanupDrag();
        },

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

                // Snapshot for undo before state mutation
                this._history.pushUndo(this._state);

                // Mutate state
                moveBlock(this._state, blockId, targetId, position);
                this._devValidate();

                // Patch DOM: move the block wrapper to its new position
                const blockWrapper = this._wrapperRegistry.get(blockId);
                const targetWrapper = this._wrapperRegistry.get(targetId);

                if (!blockWrapper || !targetWrapper) {
                        // Registry out of sync with state — DOM patch impossible.
                        // Fall back to full re-render to reconcile.
                        console.warn('[block-editor] Registry miss during drop, falling back to full render');
                        this._render();
                        this._selectBlock(blockId);
                        this._updateSaveButton();
                        return;
                }

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
                        composed: true,
                        detail: { action: 'move', blockId, targetId, position },
                }));
        },

        // ---- Auto-scroll ----

        _startAutoScroll() {
                if (this._autoScrollRaf !== null) return; // Already running

                const editorContainer = this.querySelector('[data-editor-container]');
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
        },

        _stopAutoScroll() {
                if (this._autoScrollRaf !== null) {
                        cancelAnimationFrame(this._autoScrollRaf);
                        this._autoScrollRaf = null;
                }
        },

        _onEscapeKey(e) {
                if (e.key === 'Escape' && this._drag) {
                        this._cleanupDrag();
                }
        },

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

                // Clear pointer and drag state
                this._pointerSelectActive = false;
                this._drag = null;
        },
};
