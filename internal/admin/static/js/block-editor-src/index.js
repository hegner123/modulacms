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
import { EDITOR_CSS } from './styles.js';
import { fetchDatatypes } from './cache.js';
import { domPatchMethods } from './dom-patches.js';
import { dragMethods } from './drag.js';

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
const isBrowser = typeof window !== 'undefined' && typeof CSSStyleSheet !== 'undefined';

if (isBrowser) {
        const sheet = new CSSStyleSheet();
        sheet.replaceSync(EDITOR_CSS);

        class BlockEditor extends HTMLElement {
                static get observedAttributes() {
                        return ['data-state'];
                }

                constructor() {
                        super();
                        this.attachShadow({ mode: 'open' });
                        this.shadowRoot.adoptedStyleSheets = [sheet];

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
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();
                        this._render();
                }

                // ---- Rendering ----

                _render() {
                        this.shadowRoot.innerHTML = '';
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

                        this.shadowRoot.appendChild(container);
                }

                _renderHeader() {
                        const header = document.createElement('div');
                        header.className = 'editor-header';

                        const saveBtn = document.createElement('button');
                        saveBtn.textContent = 'Save';
                        saveBtn.className = 'save-btn';
                        saveBtn.dataset.action = 'save';
                        saveBtn.disabled = !this._state?.dirty;
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

                        const deleteBtn = document.createElement('button');
                        deleteBtn.className = 'block-delete-btn';
                        deleteBtn.textContent = 'Delete';
                        deleteBtn.dataset.action = 'delete';
                        deleteBtn.dataset.blockId = block.id;
                        el.appendChild(deleteBtn);

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
                                { label: '\u2192', action: 'toolbar-indent', title: 'Indent (Tab)' },
                                { label: '\u2190', action: 'toolbar-outdent', title: 'Outdent (Shift+Tab)' },
                                { label: 'Dup', action: 'toolbar-duplicate', title: 'Duplicate (Ctrl+Shift+D)' },
                                { label: 'Del', action: 'toolbar-delete', title: 'Delete' },
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

                _openInsertDialog(position, targetId) {
                        var self = this;
                        fetchDatatypes().then(function(datatypes) {
                                self._showDialog(datatypes, position, targetId);
                        }).catch(function(err) {
                                console.error('[block-editor] Failed to load datatypes:', err);
                        });
                }

                _showDialog(datatypes, position, targetId) {
                        var self = this;

                        // Backdrop
                        var backdrop = document.createElement('div');
                        backdrop.className = 'insert-dialog-backdrop';

                        // Panel
                        var panel = document.createElement('div');
                        panel.className = 'insert-dialog-panel';

                        // Title
                        var title = document.createElement('div');
                        title.className = 'insert-dialog-title';
                        title.textContent = 'Select Content Type';
                        panel.appendChild(title);

                        if (datatypes.length === 0) {
                                var emptyMsg = document.createElement('div');
                                emptyMsg.className = 'insert-dialog-empty';
                                emptyMsg.textContent = 'No content types available';
                                panel.appendChild(emptyMsg);
                        } else {
                                // Options list
                                for (var i = 0; i < datatypes.length; i++) {
                                        var dt = datatypes[i];
                                        var option = document.createElement('button');
                                        option.className = 'insert-dialog-option';
                                        option.dataset.datatypeIdx = String(i);

                                        var labelSpan = document.createElement('span');
                                        labelSpan.className = 'insert-dialog-option-label';
                                        labelSpan.textContent = dt.label;
                                        option.appendChild(labelSpan);

                                        var typeSpan = document.createElement('span');
                                        typeSpan.className = 'insert-dialog-option-type';
                                        typeSpan.textContent = dt.type;
                                        option.appendChild(typeSpan);

                                        option.addEventListener('click', (function(datatype) {
                                                return function() {
                                                        self._closeDialog();
                                                        self._onDatatypeSelected(datatype, position, targetId);
                                                };
                                        })(dt));

                                        panel.appendChild(option);
                                }
                        }

                        // Cancel button
                        var cancelBtn = document.createElement('button');
                        cancelBtn.className = 'insert-dialog-cancel';
                        cancelBtn.textContent = 'Cancel';
                        cancelBtn.addEventListener('click', function() {
                                self._closeDialog();
                        });
                        panel.appendChild(cancelBtn);

                        backdrop.appendChild(panel);

                        // Close on backdrop click
                        backdrop.addEventListener('click', function(e) {
                                if (e.target === backdrop) {
                                        self._closeDialog();
                                }
                        });

                        // Close on Escape
                        this._dialogEscHandler = function(e) {
                                if (e.key === 'Escape') {
                                        self._closeDialog();
                                }
                        };
                        window.addEventListener('keydown', this._dialogEscHandler);

                        this._dialogBackdrop = backdrop;
                        this.shadowRoot.appendChild(backdrop);
                }

                _closeDialog() {
                        if (this._dialogBackdrop) {
                                this._dialogBackdrop.remove();
                                this._dialogBackdrop = null;
                        }
                        if (this._dialogEscHandler) {
                                window.removeEventListener('keydown', this._dialogEscHandler);
                                this._dialogEscHandler = null;
                        }
                }

                _onDatatypeSelected(datatype, position, targetId) {
                        if (!this._state) return;

                        var id = addBlockFromDatatype(this._state, datatype, position, targetId);
                        this._devValidate();

                        // Full re-render to correctly place insert buttons around the new block
                        this._render();

                        // Select the new block
                        this._selectBlock(id);

                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                composed: true,
                                detail: { action: 'add', blockId: id },
                        }));
                }

                _renderError(message, detail) {
                        this.shadowRoot.innerHTML = '';
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
                        this.shadowRoot.appendChild(container);
                }

                // ---- Event Handling ----

                _handleClick(e) {
                        const target = e.target;
                        const action = target.dataset?.action;
                        if (!action) return;

                        if (action === 'insert') {
                                var position = target.dataset.position;
                                var targetId = target.dataset.targetId || null;
                                this._openInsertDialog(position, targetId);
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
                        const blockList = this.shadowRoot.querySelector('.block-list');
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

                        // Check for children — confirm if has descendants
                        const descendantCount = getDescendantCount(this._state, blockId);
                        if (descendantCount > 0) {
                                const confirmed = confirm(`Delete "${block.label}" and ${descendantCount} children?`);
                                if (!confirmed) return;
                        }

                        // Remember parent before removal so we can clean up empty children containers
                        const parentId = block.parentId;

                        const removedIds = removeBlock(this._state, blockId);
                        this._devValidate();

                        // Clear selection if the selected block was among those removed
                        if (this._state.selectedBlockId && removedIds.includes(this._state.selectedBlockId)) {
                                this._state.selectedBlockId = null;
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

                        // Select new (or deselect if clicking the same block)
                        if (this._state.selectedBlockId === blockId) {
                                this._state.selectedBlockId = null;
                                return;
                        }

                        this._state.selectedBlockId = blockId;
                        const el = this._elementRegistry.get(blockId);
                        if (el) {
                                el.classList.add('selected');
                        }

                        this.dispatchEvent(new CustomEvent('block-editor:select', {
                                bubbles: true,
                                composed: true,
                                detail: { blockId },
                        }));
                }

                // ---- Keyboard Shortcuts ----

                _onKeyDown(e) {
                        if (!this._state) return;

                        // Tab = indent, Shift+Tab = outdent
                        if (e.key === 'Tab') {
                                const blockId = this._state.selectedBlockId;
                                if (!blockId) return; // No selection — let browser handle Tab normally

                                e.preventDefault();

                                if (e.shiftKey) {
                                        this._doOutdentBlock(blockId);
                                } else {
                                        this._doIndentBlock(blockId);
                                }
                                return;
                        }

                        // Arrow Up/Down = move selection through depth-first traversal
                        if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
                                const blockId = this._state.selectedBlockId;
                                if (!blockId) return; // No selection — let browser handle arrows normally

                                const order = getBlockTraversalOrder(this._state);
                                if (order.length === 0) return;

                                const currentIndex = order.indexOf(blockId);
                                if (currentIndex === -1) return;

                                let nextIndex;
                                if (e.key === 'ArrowUp') {
                                        nextIndex = currentIndex - 1;
                                } else {
                                        nextIndex = currentIndex + 1;
                                }

                                // Clamp to bounds — do not wrap around
                                if (nextIndex < 0 || nextIndex >= order.length) return;

                                e.preventDefault();
                                this._selectBlock(order[nextIndex]);
                                return;
                        }

                        // Ctrl+Shift+D / Cmd+Shift+D = duplicate selected block
                        if (e.key === 'd' || e.key === 'D') {
                                if ((e.ctrlKey || e.metaKey) && e.shiftKey) {
                                        const blockId = this._state.selectedBlockId;
                                        if (!blockId) return;

                                        e.preventDefault();
                                        this._doDuplicateBlock(blockId);
                                        return;
                                }
                        }

                        // Delete / Backspace = delete selected block
                        if (e.key === 'Delete' || e.key === 'Backspace') {
                                const blockId = this._state.selectedBlockId;
                                if (!blockId) return; // No selection — let browser handle normally

                                e.preventDefault();
                                this._doDeleteBlock(blockId);
                                return;
                        }

                        // Enter = add new block after selected (default type 'text')
                        if (e.key === 'Enter') {
                                const blockId = this._state.selectedBlockId;
                                if (!blockId) return; // No selection — let browser handle normally

                                e.preventDefault();
                                this._doAddBlockAfter(blockId, 'text');
                                return;
                        }
                }

                _updateSaveButton() {
                        const saveBtn = this.shadowRoot.querySelector('[data-action="save"]');
                        if (saveBtn) {
                                saveBtn.disabled = !this._state?.dirty;
                        }
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

        // Assign prototype mixins: DOM patch helpers + drag-and-drop
        Object.assign(BlockEditor.prototype, dragMethods, domPatchMethods);

        customElements.define('block-editor', BlockEditor);
}
