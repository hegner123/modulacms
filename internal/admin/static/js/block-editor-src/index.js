// index.js — <block-editor> web component entry point
//
// Imports all modules, defines the BlockEditor class, assigns prototype mixins,
// and registers the custom element. Pure function re-exports sit outside the
// isBrowser guard for test compatibility.

import { BLOCK_TYPE_CONFIG, getTypeConfig, MAX_DEPTH } from './config.js';
import { generateId } from './id.js';
import { createState } from './state.js';
import {
        unlink, insertBefore, insertAfter, insertAsFirstChild, insertAsLastChild,
        addBlock, addBlockFromDatatype, removeBlock, moveBlock,
        indentBlock, outdentBlock, duplicateBlock, moveBlockUp, moveBlockDown,
} from './tree-ops.js';
import {
        getChildren, getRootList, isDescendant, getDepth, findLastSibling,
        getBlockTraversalOrder, collectDescendants, getDescendantCount,
} from './tree-queries.js';
import { validateState } from './validate.js';
import { domPatchMethods } from './dom-patches.js';
import { dragMethods } from './drag.js';
import { pickerMethods } from './picker.js';

// ============================================================
// Re-exports — pure functions available for testing outside browser
// ============================================================

export {
        BLOCK_TYPE_CONFIG, getTypeConfig, MAX_DEPTH,
        generateId,
        createState,
        unlink, insertBefore, insertAfter, insertAsFirstChild, insertAsLastChild,
        addBlock, addBlockFromDatatype, removeBlock, moveBlock,
        indentBlock, outdentBlock, duplicateBlock, moveBlockUp, moveBlockDown,
        getChildren, getRootList, isDescendant, getDepth, findLastSibling,
        getBlockTraversalOrder, collectDescendants, getDescendantCount,
        validateState,
};

// ============================================================
// Web Component (browser-only)
// ============================================================

// Guard browser-only code for testability in Node/vitest
const isBrowser = typeof window !== 'undefined';

if (isBrowser) {
        class BlockEditor extends HTMLElement {
                static get observedAttributes() {
                        return ['data-state'];
                }

                constructor() {
                        super();

                        this._state = null;
                        this._elementRegistry = new Map(); // blockId -> block-item element
                        this._wrapperRegistry = new Map(); // blockId -> block-wrapper element
                        this._beforeUnloadHandler = this._onBeforeUnload.bind(this);

                        // Drag state
                        this._drag = null; // active drag session object, or null
                        this._dropIndicator = null;
                        this._escapeHandler = this._onEscapeKey.bind(this);
                        this._autoScrollRaf = null; // requestAnimationFrame ID for auto-scroll
                        this._lastPointerY = 0; // track pointer Y for auto-scroll

                        // Keyboard handler
                        this._keydownHandler = this._onKeyDown.bind(this);

                        // Pointer-select guard: prevents focus auto-select from
                        // fighting with click-to-select during pointer interactions
                        this._pointerSelectActive = false;

                        // Picker state
                        this._pickerOpen = false;
                        this._pickerEl = null;
                        this._pickerBackdrop = null;
                        this._pickerInsertTarget = null;
                        this._pickerInsertPosition = 'after';
                        this._pickerQuery = '';
                        this._pickerSelectedIndex = 0;
                        this._pickerData = null;
                        this._rootDatatypeId = null;
                }

                get dev() {
                        return this.hasAttribute('data-dev');
                }

                get state() {
                        return this._state;
                }

                set state(newState) {
                        this._state = newState;
                        this._state.dirty = false;
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();
                        this._render();
                }

                get dirty() {
                        return this._state ? this._state.dirty : false;
                }

                serialize() {
                        if (!this._state) return '{}';
                        return JSON.stringify({
                                blocks: this._state.blocks,
                                rootId: this._state.rootId,
                        });
                }

                getBlock(id) {
                        return this._state?.blocks[id] ?? null;
                }

                getFieldData(blockId) {
                        return this._state?.blocks[blockId]?.fields || [];
                }

                setFieldValue(blockId, fieldId, value) {
                        const block = this._state?.blocks[blockId];
                        if (!block?.fields) return;
                        const field = block.fields.find(f => f.fieldId === fieldId);
                        if (field) {
                                field.value = value;
                                this._state.dirty = true;
                                this._updateSaveButton();
                        }
                }

                save() {
                        if (!this._state) return;
                        const serialized = this.serialize();
                        this._state.dirty = false;
                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:save', {
                                bubbles: true,
                                composed: true,
                                detail: { state: serialized },
                        }));
                }

                // ---- Lifecycle ----

                connectedCallback() {
                        window.addEventListener('beforeunload', this._beforeUnloadHandler);
                        window.addEventListener('keydown', this._escapeHandler);
                        this._initState();
                }

                disconnectedCallback() {
                        window.removeEventListener('beforeunload', this._beforeUnloadHandler);
                        window.removeEventListener('keydown', this._escapeHandler);
                }

                attributeChangedCallback(name, oldValue, newValue) {
                        if (name === 'data-state' && this.isConnected) {
                                this._initState();
                        }
                }

                // ---- State Initialization ----

                _initState() {
                        const stateAttr = this.getAttribute('data-state');

                        // No attribute or empty string — start empty
                        if (stateAttr === null || stateAttr === '') {
                                this._state = createState();
                                this._render();
                                return;
                        }

                        // Parse JSON
                        let parsed;
                        try {
                                parsed = JSON.parse(stateAttr);
                        } catch (parseError) {
                                this._renderError('Invalid block data', parseError.message);
                                this.dispatchEvent(new CustomEvent('block-editor:error', {
                                        bubbles: true,
                                        composed: true,
                                        detail: { message: 'Invalid block data', error: parseError },
                                }));
                                return;
                        }

                        // Build state from parsed data
                        const newState = {
                                blocks: parsed.blocks || {},
                                rootId: parsed.rootId || null,
                                selectedBlockId: null,
                                dirty: false,
                        };

                        // Validate
                        const validationErrors = validateState(newState);
                        if (validationErrors.length > 0) {
                                const message = 'Block data has inconsistent pointers: ' + validationErrors[0];
                                this._renderError(message, validationErrors.join('\n'));
                                this.dispatchEvent(new CustomEvent('block-editor:error', {
                                        bubbles: true,
                                        composed: true,
                                        detail: { message, error: new Error(validationErrors.join('; ')) },
                                }));
                                return;
                        }

                        this._state = newState;
                        this._rootDatatypeId = this.getAttribute('data-root-datatype-id') || null;
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();
                        this._render();
                }

                // ---- Rendering ----

                _render() {
                        this.innerHTML = '';
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();

                        const container = document.createElement('div');
                        container.className = 'editor-container';

                        // Header (save button)
                        const header = this._renderHeader();
                        container.appendChild(header);

                        // Block list (with inline insert buttons)
                        const blockList = document.createElement('div');
                        blockList.className = 'block-list';
                        this._renderBlocksInto(blockList);
                        container.appendChild(blockList);

                        // Make container focusable for keyboard events
                        container.setAttribute('tabindex', '0');

                        // Delegated click handler
                        container.addEventListener('click', (e) => this._handleClick(e));

                        // Delegated pointerdown for drag initiation
                        container.addEventListener('pointerdown', (e) => this._onPointerDown(e));

                        // Keyboard shortcuts (scoped to editor container)
                        container.addEventListener('keydown', this._keydownHandler);

                        // Auto-select first block when editor gains focus via keyboard (tab).
                        // Skipped during pointer interactions — click-to-select handles those.
                        var self = this;
                        container.addEventListener('focus', function() {
                                if (self._pointerSelectActive) return;
                                if (!self._state.selectedBlockId) {
                                        var order = getBlockTraversalOrder(self._state);
                                        if (order.length > 0) self._selectBlock(order[0]);
                                }
                        });

                        this.appendChild(container);
                }

                _renderHeader() {
                        const header = document.createElement('div');
                        header.className = 'editor-header';

                        const saveBtn = document.createElement('button');
                        saveBtn.textContent = 'Save';
                        saveBtn.className = 'save-btn';
                        saveBtn.dataset.action = 'save';
                        header.appendChild(saveBtn);

                        return header;
                }

                _renderBlocksInto(container) {
                        if (!this._state) return;
                        const rootList = getRootList(this._state);

                        if (rootList.length === 0) {
                                // Empty state: show centered plus button
                                container.appendChild(this._renderEmptyState());
                                return;
                        }

                        // Insert button before the first block
                        container.appendChild(this._renderInsertButton('before', rootList[0].id));

                        for (var i = 0; i < rootList.length; i++) {
                                const block = rootList[i];
                                const wrapper = this._renderBlockWrapper(block, 0);
                                container.appendChild(wrapper);

                                // Insert button after each block
                                container.appendChild(this._renderInsertButton('after', block.id));
                        }
                }

                /**
                 * Render a block-wrapper div containing the block-item header and
                 * optionally a children-container. The wrapper is indented by depth.
                 */
                _renderBlockWrapper(block, depth) {
                        const wrapper = document.createElement('div');
                        wrapper.className = 'block-wrapper';
                        wrapper.dataset.blockId = block.id;
                        wrapper.style.marginInlineStart = (depth * 24) + 'px';

                        const header = this._renderBlockHeader(block);
                        wrapper.appendChild(header);

                        // Render children or inside-insert button for container types
                        const typeConfig = getTypeConfig(block.type);
                        const children = getChildren(this._state, block.id);
                        if (children.length > 0) {
                                const childContainer = document.createElement('div');
                                childContainer.className = 'children-container';
                                childContainer.dataset.parentId = block.id;

                                // Insert button before first child
                                childContainer.appendChild(this._renderInsertButton('before', children[0].id));

                                for (var ci = 0; ci < children.length; ci++) {
                                        const child = children[ci];
                                        childContainer.appendChild(this._renderBlockWrapper(child, depth + 1));
                                        // Insert button after each child
                                        childContainer.appendChild(this._renderInsertButton('after', child.id));
                                }
                                wrapper.appendChild(childContainer);
                        } else if (typeConfig.canHaveChildren) {
                                // Empty container — show insert button inside
                                const insideBtn = this._renderInsertButton('inside', block.id);
                                insideBtn.classList.add('insert-btn--inside');
                                wrapper.appendChild(insideBtn);
                        }

                        this._wrapperRegistry.set(block.id, wrapper);
                        return wrapper;
                }

                /**
                 * Render the block-item header element (badge, label, child count, delete button,
                 * type-specific content).
                 */
                _renderBlockHeader(block) {
                        const el = document.createElement('div');
                        el.className = 'block-item';
                        el.dataset.blockId = block.id;

                        // Add type-specific class for container styling
                        if (block.type === 'container') {
                                el.classList.add('block-item--container');
                        }

                        const badge = document.createElement('span');
                        badge.className = 'block-type-badge block-type-badge--' + block.type;
                        badge.textContent = getTypeConfig(block.type).label;
                        el.appendChild(badge);

                        const label = document.createElement('span');
                        label.className = 'block-label';
                        label.textContent = block.label;
                        el.appendChild(label);

                        // Child count badge
                        const childCount = getDescendantCount(this._state, block.id);
                        if (childCount > 0) {
                                const countBadge = document.createElement('span');
                                countBadge.className = 'child-count-badge';
                                countBadge.textContent = String(childCount);
                                countBadge.title = childCount + ' descendant' + (childCount === 1 ? '' : 's');
                                el.appendChild(countBadge);
                        }

                        // Hover toolbar (child of .block-item to avoid pointerleave flicker)
                        const hoverToolbar = this._renderHoverToolbar(block.id);
                        el.appendChild(hoverToolbar);

                        // Type-specific content area
                        const content = this._renderTypeContent(block);
                        if (content) {
                                el.appendChild(content);
                        }

                        this._elementRegistry.set(block.id, el);
                        return el;
                }

                /**
                 * Render the hover toolbar for a block header.
                 * Contains: Move Up, Move Down, Indent, Outdent, Duplicate, Delete.
                 * All buttons have data-action attributes so _onPointerDown excludes
                 * them from drag initiation.
                 */
                _renderHoverToolbar(blockId) {
                        const toolbar = document.createElement('div');
                        toolbar.className = 'block-hover-toolbar';

                        const actions = [
                                { label: '\u2191', action: 'toolbar-move-up', title: 'Move Up' },
                                { label: '\u2193', action: 'toolbar-move-down', title: 'Move Down' },
                                { label: '\u2192', action: 'toolbar-indent', title: 'Indent (>)' },
                                { label: '\u2190', action: 'toolbar-outdent', title: 'Outdent (<)' },
                                { label: '\u2398', action: 'toolbar-duplicate', title: 'Duplicate (Ctrl+Shift+D)' },
                                { label: '\u2715', action: 'toolbar-delete', title: 'Delete' },
                        ];

                        for (const def of actions) {
                                const btn = document.createElement('button');
                                btn.textContent = def.label;
                                btn.title = def.title;
                                btn.dataset.action = def.action;
                                btn.dataset.blockId = blockId;
                                toolbar.appendChild(btn);
                        }

                        return toolbar;
                }

                /**
                 * Render type-specific content for a block.
                 * text = paragraph placeholder lines
                 * heading = large bold label
                 * image = dashed rect placeholder
                 * container = labeled children area indicator
                 */
                _renderTypeContent(block) {
                        const wrapper = document.createElement('div');
                        wrapper.className = 'block-type-content block-type-content--' + block.type;

                        if (block.type === 'text') {
                                // Three placeholder lines simulating paragraph text
                                for (let i = 0; i < 3; i++) {
                                        const line = document.createElement('div');
                                        line.className = 'text-line';
                                        wrapper.appendChild(line);
                                }
                                return wrapper;
                        }

                        if (block.type === 'heading') {
                                wrapper.textContent = block.label;
                                return wrapper;
                        }

                        if (block.type === 'image') {
                                wrapper.textContent = 'Image placeholder';
                                return wrapper;
                        }

                        if (block.type === 'container') {
                                const childCount = getDescendantCount(this._state, block.id);
                                if (childCount > 0) {
                                        wrapper.textContent = childCount + ' block' + (childCount === 1 ? '' : 's') + ' inside';
                                } else {
                                        wrapper.textContent = 'Drop blocks here';
                                }
                                return wrapper;
                        }

                        return null;
                }

                /**
                 * Render children of a parent block into its children container.
                 * Used during DOM patching when a new children container is created.
                 */
                _renderChildrenInto(childContainer, parentId, parentDepth) {
                        const children = getChildren(this._state, parentId);
                        for (const child of children) {
                                const wrapper = this._renderBlockWrapper(child, parentDepth + 1);
                                childContainer.appendChild(wrapper);
                        }
                }

                // ---- Insert Buttons & Dialog ----

                _renderEmptyState() {
                        var container = document.createElement('div');
                        container.className = 'insert-empty';
                        var btn = document.createElement('button');
                        btn.className = 'insert-btn insert-btn--empty';
                        btn.title = 'Add content block';
                        btn.innerHTML = '+';
                        btn.dataset.action = 'insert';
                        btn.dataset.position = 'root';
                        container.appendChild(btn);
                        return container;
                }

                _renderInsertButton(position, targetId) {
                        var btn = document.createElement('button');
                        btn.className = 'insert-btn';
                        btn.title = 'Insert block';
                        btn.innerHTML = '+';
                        btn.dataset.action = 'insert';
                        btn.dataset.position = position;
                        btn.dataset.targetId = targetId || '';
                        return btn;
                }


                _renderError(message, detail) {
                        this.innerHTML = '';
                        const container = document.createElement('div');
                        container.className = 'editor-container';

                        const errorDiv = document.createElement('div');
                        errorDiv.className = 'error-container';

                        const msgEl = document.createElement('div');
                        msgEl.className = 'error-message';
                        msgEl.textContent = message;
                        errorDiv.appendChild(msgEl);

                        if (detail) {
                                const detailEl = document.createElement('div');
                                detailEl.className = 'error-detail';
                                detailEl.textContent = detail;
                                errorDiv.appendChild(detailEl);
                        }

                        container.appendChild(errorDiv);
                        this.appendChild(container);
                }

                // ---- Event Handling ----

                _handleClick(e) {
                        const target = e.target;
                        const action = target.dataset?.action;
                        if (!action) return;

                        if (action === 'insert') {
                                var position = target.dataset.position;
                                var targetId = target.dataset.targetId || null;
                                this._openPicker(targetId, position);
                                return;
                        } else if (action === 'add') {
                                const blockType = target.dataset.blockType || 'text';
                                this._doAddBlock(blockType);
                        } else if (action === 'delete' || action === 'toolbar-delete') {
                                this._doDeleteBlock(target.dataset.blockId);
                        } else if (action === 'save') {
                                this.save();
                        } else if (action === 'toolbar-move-up') {
                                this._doMoveBlockUp(target.dataset.blockId);
                        } else if (action === 'toolbar-move-down') {
                                this._doMoveBlockDown(target.dataset.blockId);
                        } else if (action === 'toolbar-indent') {
                                this._doIndentBlock(target.dataset.blockId);
                        } else if (action === 'toolbar-outdent') {
                                this._doOutdentBlock(target.dataset.blockId);
                        } else if (action === 'toolbar-duplicate') {
                                this._doDuplicateBlock(target.dataset.blockId);
                        }
                }

                _doAddBlock(type) {
                        const id = addBlock(this._state, type);
                        this._devValidate();

                        // Patch DOM — append new block wrapper to block list (root level, depth 0)
                        const block = this._state.blocks[id];
                        const wrapper = this._renderBlockWrapper(block, 0);
                        const blockList = this.querySelector('.block-list');
                        blockList.appendChild(wrapper);

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                composed: true,
                                detail: { action: 'add', blockId: id },
                        }));
                }

                _doDeleteBlock(blockId) {
                        const block = this._state.blocks[blockId];
                        if (!block) return;

                        const descendantCount = getDescendantCount(this._state, blockId);
                        if (descendantCount > 0) {
                                var self = this;
                                showConfirmDialog({
                                        title: 'Delete Block',
                                        message: 'Delete "' + block.label + '" and ' + descendantCount + ' children?',
                                        confirmLabel: 'Delete',
                                        destructive: true,
                                }).then(function(confirmed) {
                                        if (confirmed) self._executeDeleteBlock(blockId);
                                });
                                return;
                        }

                        this._executeDeleteBlock(blockId);
                }

                _executeDeleteBlock(blockId) {
                        const block = this._state.blocks[blockId];
                        if (!block) return;

                        const parentId = block.parentId;

                        const removedIds = removeBlock(this._state, blockId);
                        this._devValidate();

                        // Clear selection if the selected block was among those removed
                        if (this._state.selectedBlockId && removedIds.includes(this._state.selectedBlockId)) {
                                this._state.selectedBlockId = null;
                                this.dispatchEvent(new CustomEvent('block-editor:select', {
                                        bubbles: true, composed: true,
                                        detail: { blockId: null },
                                }));
                        }

                        // Patch DOM — remove wrapper elements (wrapper contains both header and children)
                        for (const id of removedIds) {
                                const wrapper = this._wrapperRegistry.get(id);
                                if (wrapper) {
                                        wrapper.remove();
                                        this._wrapperRegistry.delete(id);
                                }
                                const el = this._elementRegistry.get(id);
                                if (el) {
                                        this._elementRegistry.delete(id);
                                }
                        }

                        // Children container cleanup: if parent lost its last child, remove the empty container
                        this._cleanupEmptyChildrenContainer(parentId);

                        // Update parent's child count badge
                        this._updateChildCountBadge(parentId);

                        this._updateSaveButton();
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                composed: true,
                                detail: { action: 'remove', blockId },
                        }));
                }

                // ---- Selection ----

                _selectBlock(blockId) {
                        if (!this._state) return;

                        // Deselect previous
                        if (this._state.selectedBlockId) {
                                const prevEl = this._elementRegistry.get(this._state.selectedBlockId);
                                if (prevEl) {
                                        prevEl.classList.remove('selected');
                                }
                        }

                        // Deselect if clicking the same block
                        if (this._state.selectedBlockId === blockId) {
                                this._state.selectedBlockId = null;
                                this.dispatchEvent(new CustomEvent('block-editor:select', {
                                        bubbles: true, composed: true,
                                        detail: { blockId: null },
                                }));
                                return;
                        }

                        this._state.selectedBlockId = blockId;
                        const el = this._elementRegistry.get(blockId);
                        if (el) {
                                el.classList.add('selected');
                                el.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
                        }

                        this.dispatchEvent(new CustomEvent('block-editor:select', {
                                bubbles: true,
                                composed: true,
                                detail: { blockId },
                        }));
                }

                // ---- Keyboard Shortcuts ----

                _onKeyDown(e) {
                        if (this._pickerOpen) {
                                this._onPickerKeyDown(e);
                                return;
                        }
                        if (!this._state) return;
                        var blockId = this._state.selectedBlockId;
                        var noMod = !e.ctrlKey && !e.metaKey && !e.altKey;

                        // Tab / Shift+Tab = cycle selection through DFS order
                        if (e.key === 'Tab') {
                                e.preventDefault();
                                this._navigateDFS(e.shiftKey);
                                return;
                        }

                        // ArrowUp/ArrowDown or j/k = DFS navigation
                        if (e.key === 'ArrowDown' || (e.key === 'j' && noMod)) {
                                e.preventDefault();
                                this._navigateDFS(false);
                                return;
                        }
                        if (e.key === 'ArrowUp' || (e.key === 'k' && noMod)) {
                                e.preventDefault();
                                this._navigateDFS(true);
                                return;
                        }

                        // h / ArrowLeft = navigate to parent
                        if (e.key === 'ArrowLeft' || (e.key === 'h' && noMod)) {
                                if (!blockId) return;
                                var parentId = this._state.blocks[blockId].parentId;
                                if (parentId) {
                                        e.preventDefault();
                                        this._selectBlock(parentId);
                                }
                                return;
                        }

                        // l / ArrowRight = navigate to first child
                        if (e.key === 'ArrowRight' || (e.key === 'l' && noMod)) {
                                if (!blockId) return;
                                var childId = this._state.blocks[blockId].firstChildId;
                                if (childId) {
                                        e.preventDefault();
                                        this._selectBlock(childId);
                                }
                                return;
                        }

                        // > (Shift+.) = indent selected block
                        if (e.key === '>' && noMod) {
                                if (!blockId) return;
                                e.preventDefault();
                                this._doIndentBlock(blockId);
                                return;
                        }

                        // < (Shift+,) = outdent selected block
                        if (e.key === '<' && noMod) {
                                if (!blockId) return;
                                e.preventDefault();
                                this._doOutdentBlock(blockId);
                                return;
                        }

                        // Ctrl+Shift+D / Cmd+Shift+D = duplicate selected block
                        if (e.key === 'd' || e.key === 'D') {
                                if ((e.ctrlKey || e.metaKey) && e.shiftKey) {
                                        if (!blockId) return;
                                        e.preventDefault();
                                        this._doDuplicateBlock(blockId);
                                        return;
                                }
                        }

                        // Delete / Backspace = delete selected block
                        if (e.key === 'Delete' || e.key === 'Backspace') {
                                if (!blockId) return;
                                e.preventDefault();
                                this._doDeleteBlock(blockId);
                                return;
                        }

                        // Enter = open block picker with selected node as insert target
                        if (e.key === 'Enter') {
                                if (!blockId) return;
                                e.preventDefault();
                                this._openPicker(blockId, 'after');
                                return;
                        }
                }

                /**
                 * Navigate DFS order: select next (backward=false) or previous (backward=true) block.
                 * Auto-selects first or last block if nothing is currently selected.
                 */
                _navigateDFS(backward) {
                        var order = getBlockTraversalOrder(this._state);
                        if (order.length === 0) return;

                        var blockId = this._state.selectedBlockId;
                        if (!blockId) {
                                this._selectBlock(backward ? order[order.length - 1] : order[0]);
                                return;
                        }

                        var currentIndex = order.indexOf(blockId);
                        if (currentIndex === -1) return;

                        var nextIndex = backward ? currentIndex - 1 : currentIndex + 1;
                        if (nextIndex < 0 || nextIndex >= order.length) return;

                        this._selectBlock(order[nextIndex]);
                }

                _updateSaveButton() {
                        // no-op: save button is always enabled.
                        // Dirty tracking remains for autosave and beforeunload.
                }

                // ---- Dev-mode validation ----

                _devValidate() {
                        if (!this.dev) return;
                        const errors = validateState(this._state);
                        if (errors.length > 0) {
                                console.warn('[block-editor] Validation errors after mutation:', errors);
                        } else {
                                console.log('[block-editor] State valid');
                        }
                }

                // ---- beforeunload ----

                _onBeforeUnload(e) {
                        if (this._state?.dirty) {
                                e.preventDefault();
                                e.returnValue = 'You have unsaved changes.';
                        }
                }
        }

        // Assign prototype mixins: DOM patch helpers, drag-and-drop, picker
        Object.assign(BlockEditor.prototype, dragMethods, domPatchMethods, pickerMethods);

        customElements.define('block-editor', BlockEditor);
}
