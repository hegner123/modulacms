// picker.js — Block picker prototype mixin (telescope/fzf style)
//
// Methods in this object are assigned to BlockEditor.prototype via Object.assign.
// They access the class instance via `this`.

import { fetchDatatypesGrouped } from './cache.js';
import { addBlockFromDatatype } from './tree-ops.js';

export var pickerMethods = {

        _openPicker: function(insertTargetId, position) {
                this._pickerOpen = true;
                this._pickerInsertTarget = insertTargetId;
                this._pickerInsertPosition = position || 'after';
                this._pickerQuery = '';
                this._pickerSelectedIndex = 0;

                var self = this;
                fetchDatatypesGrouped(this._rootDatatypeId).then(function(grouped) {
                        self._pickerData = grouped;
                        self._renderPicker();
                }).catch(function(err) {
                        console.error('[block-editor] Failed to load datatypes for picker:', err);
                        self._closePicker();
                });
        },

        _closePicker: function() {
                this._pickerOpen = false;
                if (this._pickerBackdrop) {
                        this._pickerBackdrop.remove();
                        this._pickerBackdrop = null;
                }
                this._pickerEl = null;
                // Remove document-level keyboard listener
                if (this._pickerEscHandler) {
                        document.removeEventListener('keydown', this._pickerEscHandler, true);
                        this._pickerEscHandler = null;
                }
                // Return focus to editor container
                var container = this.querySelector('[data-editor-container]');
                if (container) container.focus();
        },

        _renderPicker: function() {
                // Remove any existing picker
                if (this._pickerBackdrop) {
                        this._pickerBackdrop.remove();
                }

                // Backdrop — fixed overlay, click to dismiss
                var backdrop = document.createElement('div');
                backdrop.className = 'block-picker-backdrop';

                var picker = document.createElement('div');
                picker.className = 'block-picker';

                // Results area (scrollable, on top)
                var results = document.createElement('div');
                results.className = 'block-picker-results';
                picker.appendChild(results);

                // Input area (bottom)
                var inputBar = document.createElement('div');
                inputBar.className = 'block-picker-input';

                var prompt = document.createElement('span');
                prompt.className = 'block-picker-prompt';
                prompt.textContent = '>';
                inputBar.appendChild(prompt);

                var queryDisplay = document.createElement('span');
                queryDisplay.className = 'block-picker-query';
                queryDisplay.textContent = this._pickerQuery;
                inputBar.appendChild(queryDisplay);

                picker.appendChild(inputBar);
                backdrop.appendChild(picker);

                this._pickerEl = picker;
                this._pickerBackdrop = backdrop;

                // Click on backdrop (not picker) dismisses
                var self = this;
                backdrop.addEventListener('mousedown', function(e) {
                        if (e.target === backdrop) {
                                self._closePicker();
                        }
                });

                document.body.appendChild(backdrop);

                this._renderPickerResults();

                // Document-level keydown (capture phase) — handles all picker
                // keys regardless of focus
                this._pickerEscHandler = function(e) {
                        if (!self._pickerOpen) return;
                        self._onPickerKeyDown(e);
                };
                document.addEventListener('keydown', this._pickerEscHandler, true);
        },

        /**
         * Build the flat list of selectable items from picker data,
         * filtered by the current query. Returns an array of
         * { id, label, type, depth, categoryName } objects.
         */
        _getPickerItems: function() {
                if (!this._pickerData) return [];
                var categories = this._pickerData.categories;
                var query = this._pickerQuery.toLowerCase();
                var items = [];

                for (var c = 0; c < categories.length; c++) {
                        var cat = categories[c];
                        var catItems = [];

                        for (var i = 0; i < cat.items.length; i++) {
                                var item = cat.items[i];
                                if (query && item.label.toLowerCase().indexOf(query) === -1 && (!item.name || item.name.toLowerCase().indexOf(query) === -1)) continue;
                                catItems.push(item);
                        }

                        if (catItems.length === 0) continue;

                        // Add category header marker
                        items.push({ isHeader: true, name: cat.name });

                        for (var j = 0; j < catItems.length; j++) {
                                items.push(catItems[j]);
                        }
                }

                return items;
        },

        _renderPickerResults: function() {
                var resultsEl = this._pickerEl.querySelector('.block-picker-results');
                if (!resultsEl) return;
                resultsEl.innerHTML = '';

                var items = this._getPickerItems();
                var selectableIndex = 0;

                for (var i = 0; i < items.length; i++) {
                        var item = items[i];

                        if (item.isHeader) {
                                var header = document.createElement('div');
                                header.className = 'block-picker-header';
                                header.textContent = item.name;
                                resultsEl.appendChild(header);
                                continue;
                        }

                        var row = document.createElement('div');
                        row.className = 'block-picker-item';
                        row.dataset.selectableIndex = String(selectableIndex);
                        row.dataset.datatypeId = item.id;
                        row.dataset.datatypeLabel = item.label;
                        row.dataset.datatypeType = item.type;

                        if (item.depth > 0) {
                                row.style.paddingLeft = (12 + item.depth * 16) + 'px';
                        }

                        if (selectableIndex === this._pickerSelectedIndex) {
                                row.classList.add('block-picker-item--selected');
                        }

                        row.textContent = item.label;

                        // Click to insert
                        var self = this;
                        row.addEventListener('mousedown', (function(idx) {
                                return function(e) {
                                        e.preventDefault();
                                        e.stopPropagation();
                                        self._pickerSelectedIndex = idx;
                                        self._pickerInsertBlock();
                                };
                        })(selectableIndex));

                        resultsEl.appendChild(row);
                        selectableIndex++;
                }

                // Scroll selected into view
                var selected = resultsEl.querySelector('.block-picker-item--selected');
                if (selected) {
                        selected.scrollIntoView({ block: 'nearest' });
                }

                // Update query display
                var queryEl = this._pickerEl.querySelector('.block-picker-query');
                if (queryEl) {
                        queryEl.textContent = this._pickerQuery;
                }
        },

        /**
         * Get the list of selectable (non-header) items from picker data.
         */
        _getSelectableItems: function() {
                if (!this._pickerData) return [];
                var categories = this._pickerData.categories;
                var query = this._pickerQuery.toLowerCase();
                var items = [];

                for (var c = 0; c < categories.length; c++) {
                        var cat = categories[c];
                        for (var i = 0; i < cat.items.length; i++) {
                                var item = cat.items[i];
                                if (query && item.label.toLowerCase().indexOf(query) === -1 && (!item.name || item.name.toLowerCase().indexOf(query) === -1)) continue;
                                items.push(item);
                        }
                }

                return items;
        },

        _onPickerKeyDown: function(e) {
                e.preventDefault();
                e.stopPropagation();

                if (e.key === 'Escape') {
                        this._closePicker();
                        return;
                }

                var selectableItems = this._getSelectableItems();
                var maxIndex = selectableItems.length - 1;

                if (e.key === 'ArrowUp') {
                        this._pickerSelectedIndex = Math.max(0, this._pickerSelectedIndex - 1);
                        this._renderPickerResults();
                        return;
                }

                if (e.key === 'ArrowDown') {
                        this._pickerSelectedIndex = Math.min(maxIndex, this._pickerSelectedIndex + 1);
                        this._renderPickerResults();
                        return;
                }

                if (e.key === 'Enter') {
                        this._pickerInsertBlock();
                        return;
                }

                if (e.key === 'Backspace') {
                        if (this._pickerQuery.length > 0) {
                                this._pickerQuery = this._pickerQuery.slice(0, -1);
                                this._pickerSelectedIndex = 0;
                                this._renderPickerResults();
                        }
                        return;
                }

                // Printable character (single char, no ctrl/meta)
                if (e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
                        this._pickerQuery += e.key;
                        this._pickerSelectedIndex = 0;
                        this._renderPickerResults();
                        return;
                }
        },

        _pickerInsertBlock: function() {
                var selectableItems = this._getSelectableItems();
                if (selectableItems.length === 0) return;

                var idx = this._pickerSelectedIndex;
                if (idx < 0 || idx >= selectableItems.length) return;

                var item = selectableItems[idx];
                var datatype = { id: item.id, label: item.label, type: item.type };

                this._history.pushUndo(this._state);
                var id = addBlockFromDatatype(
                        this._state,
                        datatype,
                        this._pickerInsertPosition,
                        this._pickerInsertTarget
                );

                this._closePicker();
                this._devValidate();
                this._render();
                this._selectBlock(id);

                this.dispatchEvent(new CustomEvent('block-editor:change', {
                        bubbles: true,
                        composed: true,
                        detail: { action: 'add', blockId: id },
                }));
        },
};
