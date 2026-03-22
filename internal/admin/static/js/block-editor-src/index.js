// index.js — <block-editor> web component entry point
//
// Imports all modules, defines the BlockEditor class, assigns prototype mixins,
// and registers the custom element. Pure function re-exports sit outside the
// isBrowser guard for test compatibility.

import { MAX_DEPTH } from './config.js';
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
import { History } from './history.js';

// ============================================================
// Re-exports — pure functions available for testing outside browser
// ============================================================

export {
        MAX_DEPTH,
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
                constructor() {
                        super();

                        this._state = null;
                        this._baseline = null; // frozen snapshot for save diffing
                        this._elementRegistry = new Map(); // blockId -> block-item element
                        this._wrapperRegistry = new Map(); // blockId -> block-wrapper element
                        this._collapsedBlocks = new Set(); // blockId set for collapsed blocks
                        this._beforeUnloadHandler = this._onBeforeUnload.bind(this);
                        this._visibilityHandler = this._onVisibilityChange.bind(this);
                        this._recoveryTimer = null;
                        this._recoveryDebounceMs = 30000;

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
                        this._history = new History(50);
                }

                get dev() {
                        return this.hasAttribute('data-dev');
                }

                get state() {
                        return this._state;
                }

                set state(newState) {
                        this._state = newState;
                        if (this._history) this._history.clear();
                        this._state.dirty = false;
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();
                        this._collapsedBlocks.clear();
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

                // ---- Baseline (diff target for saves) ----

                _cloneBlocks(state) {
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

                getBaseline() {
                        return this._baseline;
                }

                // commitSave is called by the save handler after a successful server
                // response. It remaps client UUIDs to server ULIDs, snapshots the
                // current state as the new baseline, and clears the dirty flag.
                // Order matters — see BLOCK_EDITOR_STATE_PLAN.md §2.
                commitSave(idMap) {
                        if (!this._state) return;

                        // 1. Remap live state block IDs and pointers
                        if (idMap && Object.keys(idMap).length > 0) {
                                var clientIds = Object.keys(idMap);
                                for (var i = 0; i < clientIds.length; i++) {
                                        var clientId = clientIds[i];
                                        var serverId = idMap[clientId];
                                        var block = this._state.blocks[clientId];
                                        if (!block) continue;
                                        block.id = serverId;
                                        this._state.blocks[serverId] = block;
                                        delete this._state.blocks[clientId];
                                        if (this._state.rootId === clientId) {
                                                this._state.rootId = serverId;
                                        }
                                }
                                // Second pass: remap pointer fields across ALL blocks
                                var allKeys = Object.keys(this._state.blocks);
                                for (var j = 0; j < allKeys.length; j++) {
                                        var b = this._state.blocks[allKeys[j]];
                                        if (b.parentId && idMap[b.parentId]) b.parentId = idMap[b.parentId];
                                        if (b.firstChildId && idMap[b.firstChildId]) b.firstChildId = idMap[b.firstChildId];
                                        if (b.nextSiblingId && idMap[b.nextSiblingId]) b.nextSiblingId = idMap[b.nextSiblingId];
                                        if (b.prevSiblingId && idMap[b.prevSiblingId]) b.prevSiblingId = idMap[b.prevSiblingId];
                                }

                                // 2. Remap element/wrapper registries + all descendant data-block-id attrs
                                for (var k = 0; k < clientIds.length; k++) {
                                        var cid = clientIds[k];
                                        var sid = idMap[cid];
                                        if (this._elementRegistry.has(cid)) {
                                                var el = this._elementRegistry.get(cid);
                                                this._elementRegistry.delete(cid);
                                                this._elementRegistry.set(sid, el);
                                                el.dataset.blockId = sid;
                                        }
                                        if (this._wrapperRegistry.has(cid)) {
                                                var wrapper = this._wrapperRegistry.get(cid);
                                                this._wrapperRegistry.delete(cid);
                                                this._wrapperRegistry.set(sid, wrapper);
                                                wrapper.dataset.blockId = sid;
                                                // Update child elements (kebab button, chevron, etc.)
                                                var descendants = wrapper.querySelectorAll('[data-block-id="' + cid + '"]');
                                                for (var d = 0; d < descendants.length; d++) {
                                                        descendants[d].dataset.blockId = sid;
                                                }
                                        }
                                }

                                // 3. Remap history stacks
                                this._history.remapIds(idMap);

                                // 4. Remap selectedBlockId
                                if (this._state.selectedBlockId && idMap[this._state.selectedBlockId] !== undefined) {
                                        this._state.selectedBlockId = idMap[this._state.selectedBlockId];
                                }
                        }

                        // 5. Snapshot current state as new baseline
                        this._baseline = this._cloneBlocks(this._state);

                        // 6. Clear dirty flag
                        this._state.dirty = false;
                        this._updateSaveButton();

                        // 7. Clear crash recovery (successful save = nothing to recover)
                        this._clearRecovery();
                }

                // ---- Crash Recovery ----

                _recoveryKey() {
                        var contentId = this.getAttribute('data-content-id');
                        return contentId ? 'mcms-block-recovery:' + contentId : null;
                }

                _scheduleRecoveryWrite() {
                        clearTimeout(this._recoveryTimer);
                        var self = this;
                        this._recoveryTimer = setTimeout(function() {
                                self._writeRecovery();
                        }, this._recoveryDebounceMs);
                }

                _writeRecovery() {
                        var key = this._recoveryKey();
                        if (!key || !this._state || !this._state.dirty) return;
                        try {
                                sessionStorage.setItem(key, JSON.stringify({
                                        state: this.serialize(),
                                        timestamp: Date.now(),
                                }));
                        } catch (e) {
                                // QuotaExceededError or SecurityError — silently ignore
                        }
                }

                _clearRecovery() {
                        clearTimeout(this._recoveryTimer);
                        var key = this._recoveryKey();
                        if (!key) return;
                        try {
                                sessionStorage.removeItem(key);
                        } catch (e) {
                                // SecurityError — silently ignore
                        }
                }

                _checkRecovery() {
                        var key = this._recoveryKey();
                        if (!key) return;
                        var raw;
                        try {
                                raw = sessionStorage.getItem(key);
                        } catch (e) {
                                return;
                        }
                        if (!raw) return;

                        var recovery;
                        try {
                                recovery = JSON.parse(raw);
                        } catch (e) {
                                this._clearRecovery();
                                return;
                        }

                        if (!recovery.state || !recovery.timestamp) {
                                this._clearRecovery();
                                return;
                        }

                        var self = this;
                        var age = Date.now() - recovery.timestamp;
                        var ageStr = age < 60000
                                ? Math.round(age / 1000) + 's ago'
                                : Math.round(age / 60000) + 'm ago';

                        var banner = document.createElement('div');
                        banner.className = 'recovery-banner';
                        banner.innerHTML =
                                '<span>Unsaved changes recovered (' + ageStr + ')</span>' +
                                '<button class="recovery-restore">Restore</button>' +
                                '<button class="recovery-dismiss">Dismiss</button>';

                        banner.querySelector('.recovery-restore').addEventListener('click', function() {
                                var parsed;
                                try {
                                        parsed = JSON.parse(recovery.state);
                                } catch (e) {
                                        self._clearRecovery();
                                        banner.remove();
                                        return;
                                }
                                var restored = {
                                        blocks: parsed.blocks || {},
                                        rootId: parsed.rootId || null,
                                        selectedBlockId: null,
                                        dirty: true,
                                };
                                var errors = validateState(restored);
                                if (errors.length > 0) {
                                        self._clearRecovery();
                                        banner.remove();
                                        return;
                                }
                                self._state = restored;
                                self._history.clear();
                                self._elementRegistry.clear();
                                self._wrapperRegistry.clear();
                                self._collapsedBlocks.clear();
                                self._render();
                                self._updateSaveButton();
                                self._clearRecovery();
                        });

                        banner.querySelector('.recovery-dismiss').addEventListener('click', function() {
                                self._clearRecovery();
                                banner.remove();
                        });

                        this.prepend(banner);
                }

                _onVisibilityChange() {
                        if (document.hidden && this._state && this._state.dirty) {
                                this._writeRecovery();
                        }
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
                                if (!this._history.inFieldBatch) this._history.pushFieldChange(this._state);
                                field.value = value;
                                this._state.dirty = true;
                                this._updateSaveButton();
                                this._updateContentPreview(blockId);
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
                        document.addEventListener('visibilitychange', this._visibilityHandler);
                        this._initState();
                }

                disconnectedCallback() {
                        window.removeEventListener('beforeunload', this._beforeUnloadHandler);
                        window.removeEventListener('keydown', this._escapeHandler);
                        document.removeEventListener('visibilitychange', this._visibilityHandler);
                        clearTimeout(this._recoveryTimer);
                        if (this._history) this._history.clear();
                }

                // ---- State Initialization ----

                _initState() {
                        const stateAttr = this.getAttribute('data-state');

                        // No attribute or empty string — start empty
                        if (stateAttr === null || stateAttr === '') {
                                this._state = createState();
                                this._baseline = null;
                                if (this._history) this._history.clear();
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
                        this._baseline = this._cloneBlocks(newState);
                        if (this._history) this._history.clear();
                        this._rootDatatypeId = this.getAttribute('data-root-datatype-id') || null;
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();
                        this._collapsedBlocks.clear();

                        // Consume the attribute and free the DOM
                        this.removeAttribute('data-state');

                        this._render();

                        // Check for crash recovery snapshot after render
                        this._checkRecovery();
                }

                // ---- Rendering ----

                _render() {
                        // Cancel any active drag before destroying the DOM it references.
                        if (this._drag) {
                                this._cleanupDrag();
                        }

                        this.innerHTML = '';
                        this._elementRegistry.clear();
                        this._wrapperRegistry.clear();

                        const container = document.createElement('div');
                        container.className = 'editor-container';
                        container.setAttribute('data-editor-container', '');

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

                        const collapseControls = document.createElement('div');
                        collapseControls.className = 'collapse-controls';

                        const expandAllBtn = document.createElement('button');
                        expandAllBtn.textContent = 'Expand All';
                        expandAllBtn.className = 'collapse-btn';
                        expandAllBtn.dataset.action = 'expand-all';
                        collapseControls.appendChild(expandAllBtn);

                        const collapseAllBtn = document.createElement('button');
                        collapseAllBtn.textContent = 'Collapse All';
                        collapseAllBtn.className = 'collapse-btn';
                        collapseAllBtn.dataset.action = 'collapse-all';
                        collapseControls.appendChild(collapseAllBtn);

                        header.appendChild(collapseControls);

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

                        for (var i = 0; i < rootList.length; i++) {
                                const block = rootList[i];
                                const wrapper = this._renderBlockWrapper(block, 0);
                                container.appendChild(wrapper);
                        }
                }

                /**
                 * Render a block-wrapper div containing the block-item header and
                 * optionally a children-container. The wrapper is indented by depth.
                 */
                _renderBlockWrapper(block, depth) {
                        const wrapper = document.createElement('div');
                        wrapper.className = 'block-wrapper';
                        if (this._collapsedBlocks.has(block.id)) {
                                wrapper.classList.add('collapsed');
                        }
                        wrapper.dataset.blockId = block.id;
                        wrapper.style.marginInlineStart = (depth * 24) + 'px';

                        const header = this._renderBlockHeader(block);
                        wrapper.appendChild(header);

                        const children = getChildren(this._state, block.id);
                        if (children.length > 0) {
                                const childContainer = document.createElement('div');
                                childContainer.className = 'children-container';
                                childContainer.dataset.parentId = block.id;

                                for (var ci = 0; ci < children.length; ci++) {
                                        const child = children[ci];
                                        childContainer.appendChild(this._renderBlockWrapper(child, depth + 1));
                                }
                                wrapper.appendChild(childContainer);
                        }

                        this._wrapperRegistry.set(block.id, wrapper);
                        return wrapper;
                }

                /**
                 * Render the block-item header element (label, child count, delete button,
                 * type-specific content).
                 */
                _renderBlockHeader(block) {
                        const el = document.createElement('div');
                        el.className = 'block-item';
                        el.dataset.blockId = block.id;

                        // Collapse chevron
                        const chevron = document.createElement('button');
                        chevron.className = 'block-chevron';
                        chevron.dataset.action = 'toggle-collapse';
                        chevron.dataset.blockId = block.id;
                        chevron.textContent = this._collapsedBlocks.has(block.id) ? '\u25B8' : '\u25BE';
                        chevron.title = 'Toggle collapse';
                        el.appendChild(chevron);

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

                        // Kebab menu button (three dots)
                        const kebab = document.createElement('button');
                        kebab.className = 'block-kebab';
                        kebab.dataset.action = 'kebab-menu';
                        kebab.dataset.blockId = block.id;
                        kebab.title = 'Block actions';
                        kebab.innerHTML = '\u22EE'; // vertical ellipsis
                        el.appendChild(kebab);

                        // Content preview (actual field values)
                        const preview = this._renderContentPreview(block);
                        if (preview) {
                                el.appendChild(preview);
                        }

                        this._elementRegistry.set(block.id, el);
                        return el;
                }

                // ---- Kebab Context Menu ----

                _openKebabMenu(blockId, anchorEl) {
                        console.log('[kebab-debug] _openKebabMenu — blockId:', blockId, 'anchorEl:', anchorEl.tagName + '.' + anchorEl.className, 'anchor in DOM:', document.contains(anchorEl));
                        this._closeKebabMenu();

                        const block = this._state.blocks[blockId];
                        if (!block) {
                                console.log('[kebab-debug] _openKebabMenu — block not found in state, aborting');
                                return;
                        }

                        const canNest = getDepth(this._state, blockId) < MAX_DEPTH - 1;

                        var menu = document.createElement('div');
                        menu.className = 'block-context-menu';

                        var sections = [
                                {
                                        label: 'Add',
                                        items: [
                                                { label: 'Add Before', action: 'insert', position: 'before', blockId: blockId },
                                                { label: 'Add After', action: 'insert', position: 'after', blockId: blockId },
                                                canNest ? { label: 'Add Inside', action: 'insert', position: 'inside', blockId: blockId } : null,
                                        ],
                                },
                                {
                                        label: 'Move',
                                        items: [
                                                { label: 'Move Up', action: 'toolbar-move-up', blockId: blockId },
                                                { label: 'Move Down', action: 'toolbar-move-down', blockId: blockId },
                                                { label: 'Indent', action: 'toolbar-indent', blockId: blockId },
                                                { label: 'Outdent', action: 'toolbar-outdent', blockId: blockId },
                                        ],
                                },
                                {
                                        label: null,
                                        items: [
                                                { label: 'Duplicate', action: 'toolbar-duplicate', blockId: blockId },
                                                { label: 'Delete', action: 'toolbar-delete', blockId: blockId, destructive: true },
                                        ],
                                },
                        ];

                        for (var si = 0; si < sections.length; si++) {
                                var section = sections[si];
                                if (si > 0) {
                                        var sep = document.createElement('div');
                                        sep.className = 'context-menu-separator';
                                        menu.appendChild(sep);
                                }
                                if (section.label) {
                                        var heading = document.createElement('div');
                                        heading.className = 'context-menu-heading';
                                        heading.textContent = section.label;
                                        menu.appendChild(heading);
                                }
                                for (var ii = 0; ii < section.items.length; ii++) {
                                        var item = section.items[ii];
                                        if (!item) continue;
                                        var btn = document.createElement('button');
                                        btn.className = 'context-menu-item';
                                        if (item.destructive) btn.classList.add('context-menu-item--destructive');
                                        btn.textContent = item.label;
                                        btn.dataset.action = item.action;
                                        btn.dataset.blockId = item.blockId;
                                        if (item.position) btn.dataset.position = item.position;
                                        if (item.position) btn.dataset.targetId = item.blockId;
                                        menu.appendChild(btn);
                                }
                        }

                        // Position fixed to viewport, directly below the kebab button
                        var rect = anchorEl.getBoundingClientRect();
                        menu.style.position = 'fixed';
                        menu.style.top = (rect.bottom + 4) + 'px';
                        menu.style.left = rect.left + 'px';

                        this._kebabMenu = menu;
                        // Route menu clicks back to the block editor's _handleClick
                        var self = this;
                        menu.addEventListener('click', function(evt) {
                                var item = evt.target.closest('[data-action]');
                                console.log('[kebab-debug] menu click listener — e.target:', evt.target.tagName + '.' + evt.target.className, 'nodeType:', evt.target.nodeType, 'closest [data-action]:', item ? item.dataset.action + ' blockId:' + item.dataset.blockId : 'null', 'menu in DOM:', document.contains(menu));
                                if (item) self._handleClick(evt);
                        });

                        document.body.appendChild(menu);
                        console.log('[kebab-debug] _openKebabMenu — menu appended to body, position:', menu.style.top, menu.style.left, 'items:', menu.querySelectorAll('[data-action]').length);

                        // Close on outside click. Register synchronously but skip the
                        // current event cycle via an armed flag to avoid catching the
                        // opening click. This prevents a second kebab click during the
                        // rAF gap from opening two menus simultaneously.
                        var self = this;
                        var armed = false;
                        requestAnimationFrame(function() { armed = true; });
                        self._kebabOutsideHandler = function(evt) {
                                if (!armed) return;
                                var inside = menu.contains(evt.target);
                                if (!inside) {
                                        self._closeKebabMenu();
                                }
                        };
                        document.addEventListener('pointerdown', self._kebabOutsideHandler, true);

                        // Viewport overflow: flip up if clipped at bottom, nudge left if clipped at right
                        requestAnimationFrame(function() {
                                var menuRect = menu.getBoundingClientRect();
                                if (menuRect.bottom > window.innerHeight - 8) {
                                        menu.style.top = (rect.top - menuRect.height - 4) + 'px';
                                }
                                if (menuRect.right > window.innerWidth - 8) {
                                        menu.style.left = (rect.right - menuRect.width) + 'px';
                                }
                        });
                }

                _closeKebabMenu() {
                        console.log('[kebab-debug] _closeKebabMenu — menu exists:', !!this._kebabMenu, 'outsideHandler exists:', !!this._kebabOutsideHandler);
                        if (this._kebabMenu) {
                                this._kebabMenu.remove();
                                this._kebabMenu = null;
                        }
                        if (this._kebabOutsideHandler) {
                                document.removeEventListener('pointerdown', this._kebabOutsideHandler, true);
                                this._kebabOutsideHandler = null;
                        }
                }

                /**
                 * Render content preview showing field labels and values.
                 * Returns null only when the block has no fields at all.
                 */
                _renderContentPreview(block) {
                        var fields = block.fields || [];
                        if (fields.length === 0) return null;

                        var preview = document.createElement('div');
                        preview.className = 'block-content-preview';

                        for (var i = 0; i < fields.length; i++) {
                                var f = fields[i];
                                var label = f.label || f.fieldId || 'Field';
                                var value = (f.value || '').trim();

                                var el = document.createElement('div');
                                el.className = 'preview-field';

                                var labelSpan = document.createElement('span');
                                labelSpan.className = 'preview-field-label';
                                labelSpan.textContent = label;
                                el.appendChild(labelSpan);

                                if (!value) {
                                        var emptySpan = document.createElement('span');
                                        emptySpan.className = 'preview-field-empty';
                                        emptySpan.textContent = '\u2014';
                                        el.appendChild(emptySpan);
                                } else {
                                        // Truncate ULID-looking values to just the type hint
                                        var isUlid = value.length === 26 && /^[0-9A-HJKMNP-TV-Z]{26}$/.test(value);
                                        var valueSpan = document.createElement('span');
                                        valueSpan.className = 'preview-field-value';
                                        if (isUlid) {
                                                valueSpan.textContent = '(ref)';
                                                valueSpan.classList.add('preview-field-ref');
                                        } else if (value.length > 120) {
                                                valueSpan.textContent = value.substring(0, 120) + '\u2026';
                                        } else {
                                                valueSpan.textContent = value;
                                        }
                                        el.appendChild(valueSpan);
                                }

                                preview.appendChild(el);
                        }

                        return preview;
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

                _renderError(message, detail) {
                        this.innerHTML = '';
                        const container = document.createElement('div');
                        container.className = 'editor-container';
                        container.setAttribute('data-editor-container', '');

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
                        const target = e.target.closest('[data-action]');
                        if (!target) {
                                console.log('[kebab-debug] _handleClick — no [data-action] found, e.target:', e.target.tagName, e.target.className, 'nodeType:', e.target.nodeType);
                                return;
                        }
                        const action = target.dataset.action;
                        console.log('[kebab-debug] _handleClick — action:', action, 'blockId:', target.dataset.blockId, 'target element:', target.tagName + '.' + target.className, 'event source:', e.currentTarget === this._kebabMenu ? 'context-menu' : 'editor-container');

                        if (action === 'kebab-menu') {
                                console.log('[kebab-debug] → kebab-menu branch — blockId:', target.dataset.blockId, 'anchor rect:', JSON.stringify(target.getBoundingClientRect().toJSON()));
                                e.stopPropagation();
                                this._openKebabMenu(target.dataset.blockId, target);
                                return;
                        } else if (action === 'insert') {
                                console.log('[kebab-debug] → insert branch — position:', target.dataset.position, 'targetId:', target.dataset.targetId);
                                this._closeKebabMenu();
                                var position = target.dataset.position;
                                var targetId = target.dataset.targetId || null;
                                this._openPicker(targetId, position);
                                return;
                        } else if (action === 'add') {
                                console.log('[kebab-debug] → add branch (legacy) — blockType:', target.dataset.blockType);
                                this._closeKebabMenu();
                                const blockType = target.dataset.blockType || 'text';
                                this._doAddBlock(blockType);
                        } else if (action === 'delete' || action === 'toolbar-delete') {
                                console.log('[kebab-debug] → delete branch — blockId:', target.dataset.blockId, 'block exists:', !!this._state?.blocks[target.dataset.blockId]);
                                this._closeKebabMenu();
                                this._doDeleteBlock(target.dataset.blockId);
                        } else if (action === 'save') {
                                this.save();
                        } else if (action === 'toolbar-move-up') {
                                console.log('[kebab-debug] → move-up branch — blockId:', target.dataset.blockId);
                                this._closeKebabMenu();
                                this._doMoveBlockUp(target.dataset.blockId);
                        } else if (action === 'toolbar-move-down') {
                                console.log('[kebab-debug] → move-down branch — blockId:', target.dataset.blockId);
                                this._closeKebabMenu();
                                this._doMoveBlockDown(target.dataset.blockId);
                        } else if (action === 'toolbar-indent') {
                                console.log('[kebab-debug] → indent branch — blockId:', target.dataset.blockId);
                                this._closeKebabMenu();
                                this._doIndentBlock(target.dataset.blockId);
                        } else if (action === 'toolbar-outdent') {
                                console.log('[kebab-debug] → outdent branch — blockId:', target.dataset.blockId);
                                this._closeKebabMenu();
                                this._doOutdentBlock(target.dataset.blockId);
                        } else if (action === 'toolbar-duplicate') {
                                console.log('[kebab-debug] → duplicate branch — blockId:', target.dataset.blockId);
                                this._closeKebabMenu();
                                this._doDuplicateBlock(target.dataset.blockId);
                        } else if (action === 'toggle-collapse') {
                                this._toggleCollapse(target.dataset.blockId);
                        } else if (action === 'expand-all') {
                                this._expandAll();
                        } else if (action === 'collapse-all') {
                                this._collapseAll();
                        }
                }

                _doAddBlock(type) {
                        this._history.pushUndo(this._state);
                        const id = addBlock(this._state, type);
                        this._devValidate();

                        // Patch DOM — append new block wrapper to block list (root level, depth 0)
                        const block = this._state.blocks[id];
                        const wrapper = this._renderBlockWrapper(block, 0);
                        const blockList = this.querySelector('.block-list');

                        // Remove empty-state insert button if present.
                        const emptyInsert = blockList.querySelector('.insert-empty');
                        if (emptyInsert) emptyInsert.remove();

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

                        this._history.pushUndo(this._state);

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
                                this._collapsedBlocks.delete(id);
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

                        if ((e.ctrlKey || e.metaKey) && !e.shiftKey && e.key === 'z') {
                                e.preventDefault();
                                if (!this._drag) this._undo();
                                return;
                        }
                        if ((e.ctrlKey || e.metaKey) && ((e.shiftKey && (e.key === 'z' || e.key === 'Z')) || e.key === 'y')) {
                                e.preventDefault();
                                if (!this._drag) this._redo();
                                return;
                        }

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
                        // Save button is always enabled — dirty tracking is for
                        // beforeunload, crash recovery, and save diffing.
                        if (this._state && this._state.dirty) {
                                this._scheduleRecoveryWrite();
                        }
                }

                // ---- Collapse / Expand ----

                _toggleCollapse(blockId) {
                        if (!blockId) return;
                        var wrapper = this._wrapperRegistry.get(blockId);
                        if (!wrapper) return;

                        if (this._collapsedBlocks.has(blockId)) {
                                this._collapsedBlocks.delete(blockId);
                                wrapper.classList.remove('collapsed');
                        } else {
                                this._collapsedBlocks.add(blockId);
                                wrapper.classList.add('collapsed');
                        }

                        var chevron = wrapper.querySelector(':scope > .block-item > .block-chevron');
                        if (chevron) {
                                chevron.textContent = this._collapsedBlocks.has(blockId) ? '\u25B8' : '\u25BE';
                        }
                }

                _expandAll() {
                        this._collapsedBlocks.clear();
                        this._render();
                }

                _collapseAll() {
                        for (var id in this._state.blocks) {
                                this._collapsedBlocks.add(id);
                        }
                        this._render();
                }

                _updateContentPreview(blockId) {
                        var el = this._elementRegistry.get(blockId);
                        if (!el) return;
                        var block = this._state.blocks[blockId];
                        if (!block) return;

                        var existingPreview = el.querySelector('.block-content-preview');
                        var newPreview = this._renderContentPreview(block);

                        if (existingPreview && newPreview) {
                                existingPreview.replaceWith(newPreview);
                        } else if (existingPreview && !newPreview) {
                                existingPreview.remove();
                        } else if (!existingPreview && newPreview) {
                                el.appendChild(newPreview);
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
                                // Flush recovery snapshot before the tab closes
                                this._writeRecovery();
                                e.preventDefault();
                                e.returnValue = 'You have unsaved changes.';
                        } else {
                                // Clean navigation — clear any stale recovery data
                                this._clearRecovery();
                        }
                }

                // ---- Undo / Redo ----

                _undo() {
                        if (!this._history.canUndo) return;
                        var entry = this._history.popUndo(this._state);
                        this._restoreSnapshot(entry);
                }

                _redo() {
                        if (!this._history.canRedo) return;
                        var entry = this._history.popRedo(this._state);
                        this._restoreSnapshot(entry);
                }

                _restoreSnapshot(entry) {
                        this._state.blocks = entry.snapshot.blocks;
                        this._state.rootId = entry.snapshot.rootId;
                        this._state.selectedBlockId = null;
                        this._state.dirty = true;
                        this._render();
                        if (entry.selectedBlockId && this._state.blocks[entry.selectedBlockId]) {
                                this._selectBlock(entry.selectedBlockId);
                        }
                        this.dispatchEvent(new CustomEvent('block-editor:change', {
                                bubbles: true,
                                composed: true,
                                detail: { action: 'undo-redo' },
                        }));
                }

                // remapIds is kept for backward compatibility but delegates to commitSave.
                // Prefer calling commitSave(idMap) directly from save handlers.
                remapIds(idMap) {
                        this.commitSave(idMap);
                }
        }

        // Assign prototype mixins: DOM patch helpers, drag-and-drop, picker
        Object.assign(BlockEditor.prototype, dragMethods, domPatchMethods, pickerMethods);

        customElements.define('block-editor', BlockEditor);
}
