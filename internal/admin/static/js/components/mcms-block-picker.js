/**
 * @module mcms-block-picker
 * @description Standalone telescope/fzf-style datatype picker for adding content blocks.
 * Extracted from the block editor's picker.js mixin. Fetches datatypes from
 * /admin/api/datatypes, groups them into categories, and fires a custom event
 * on selection.
 *
 * @attr {string} parent-id - The content_data_id to create the new block under.
 * @attr {string} root-datatype-id - Optional. The root datatype ID for category grouping.
 *
 * @fires mcms-block-picker:select - { datatypeId, datatypeLabel, datatypeType, parentId }
 *
 * @example
 * <mcms-block-picker id="picker" parent-id="01ABC..." root-datatype-id="01XYZ..."></mcms-block-picker>
 * document.getElementById('picker').open();
 */

// --- Datatype cache (shared across instances) ---
var _dtCache = { data: null, fetchedAt: 0, ttl: 5 * 60 * 1000, pending: null };

function hasBase(type, base) {
    return type === base || (type.length > base.length && type.substring(0, base.length) === base && type.charAt(base.length) === '_');
}

function isSystemType(type) {
    return hasBase(type, '_root') || hasBase(type, '_nested_root') || hasBase(type, '_system_log');
}

function fetchDatatypes() {
    var now = Date.now();
    if (_dtCache.data && (now - _dtCache.fetchedAt) < _dtCache.ttl) {
        return Promise.resolve(_dtCache.data);
    }
    if (_dtCache.pending) return _dtCache.pending;

    _dtCache.pending = fetch('/admin/api/datatypes', { credentials: 'same-origin' })
        .then(function(res) {
            if (!res.ok) throw new Error('Failed to fetch datatypes: ' + res.status);
            return res.json();
        })
        .then(function(datatypes) {
            _dtCache.data = datatypes.map(function(dt) {
                return { id: dt.datatype_id, parentId: dt.parent_id || null, name: dt.name, label: dt.label, type: dt.type };
            });
            _dtCache.fetchedAt = Date.now();
            _dtCache.pending = null;
            return _dtCache.data;
        })
        .catch(function(err) {
            _dtCache.pending = null;
            throw err;
        });
    return _dtCache.pending;
}

function fetchDatatypesGrouped(rootDatatypeId) {
    return fetchDatatypes().then(function(datatypes) {
        var childrenOf = {};
        var byId = {};
        for (var i = 0; i < datatypes.length; i++) {
            var dt = datatypes[i];
            byId[dt.id] = dt;
            var pid = dt.parentId || '_none';
            if (!childrenOf[pid]) childrenOf[pid] = [];
            childrenOf[pid].push(dt);
        }

        function collectChildren(parentId, baseDepth) {
            var result = [];
            var kids = childrenOf[parentId];
            if (!kids) return result;
            for (var j = 0; j < kids.length; j++) {
                var kid = kids[j];
                if (isSystemType(kid.type)) continue;
                result.push({ id: kid.id, name: kid.name, label: kid.label, type: kid.type, depth: baseDepth });
                var gc = collectChildren(kid.id, baseDepth + 1);
                for (var k = 0; k < gc.length; k++) result.push(gc[k]);
            }
            return result;
        }

        var categories = [];

        if (rootDatatypeId && byId[rootDatatypeId]) {
            var rootDt = byId[rootDatatypeId];
            var rootItems = collectChildren(rootDatatypeId, 0);
            if (rootItems.length > 0) categories.push({ name: rootDt.label, items: rootItems });
        }

        var collectionItems = [];
        for (var ci = 0; ci < datatypes.length; ci++) {
            if (hasBase(datatypes[ci].type, '_collection')) {
                var kids2 = collectChildren(datatypes[ci].id, 0);
                for (var ck = 0; ck < kids2.length; ck++) collectionItems.push(kids2[ck]);
            }
        }
        if (collectionItems.length > 0) categories.push({ name: 'Collections', items: collectionItems });

        var globalItems = [];
        for (var gi = 0; gi < datatypes.length; gi++) {
            var gdt = datatypes[gi];
            if (hasBase(gdt.type, '_global')) {
                globalItems.push({ id: gdt.id, name: gdt.name, label: gdt.label, type: gdt.type, depth: 0 });
                var gkids = collectChildren(gdt.id, 1);
                for (var gk = 0; gk < gkids.length; gk++) globalItems.push(gkids[gk]);
            }
        }
        if (globalItems.length > 0) categories.push({ name: 'Global', items: globalItems });

        return { categories: categories };
    });
}

// --- Component ---
class McmsBlockPicker extends HTMLElement {
    constructor() {
        super();
        this._open = false;
        this._query = '';
        this._selectedIndex = 0;
        this._data = null;
        this._backdrop = null;
        this._pickerEl = null;
        this._keyHandler = null;
    }

    open() {
        if (this._open) return;
        this._open = true;
        this._query = '';
        this._selectedIndex = 0;

        var self = this;
        var rootDtId = this.getAttribute('root-datatype-id') || '';
        fetchDatatypesGrouped(rootDtId).then(function(grouped) {
            if (!self._open) return;
            self._data = grouped;
            self._render();
        }).catch(function(err) {
            console.error('[mcms-block-picker] Failed to load datatypes:', err);
            if (self._open) self.close();
        });
    }

    close() {
        this._open = false;
        if (this._backdrop) {
            this._backdrop.remove();
            this._backdrop = null;
        }
        this._pickerEl = null;
        if (this._keyHandler) {
            document.removeEventListener('keydown', this._keyHandler, true);
            this._keyHandler = null;
        }
    }

    _render() {
        if (this._backdrop) this._backdrop.remove();

        var backdrop = document.createElement('div');
        backdrop.className = 'block-picker-backdrop';

        var picker = document.createElement('div');
        picker.className = 'block-picker';

        var results = document.createElement('div');
        results.className = 'block-picker-results';
        picker.appendChild(results);

        var inputBar = document.createElement('div');
        inputBar.className = 'block-picker-input';
        var prompt = document.createElement('span');
        prompt.className = 'block-picker-prompt';
        prompt.textContent = '>';
        inputBar.appendChild(prompt);
        var queryDisplay = document.createElement('span');
        queryDisplay.className = 'block-picker-query';
        queryDisplay.textContent = this._query;
        inputBar.appendChild(queryDisplay);
        picker.appendChild(inputBar);

        backdrop.appendChild(picker);
        this._pickerEl = picker;
        this._backdrop = backdrop;

        var self = this;
        backdrop.addEventListener('mousedown', function(e) {
            if (e.target === backdrop) self.close();
        });

        document.body.appendChild(backdrop);
        this._renderResults();

        this._keyHandler = function(e) {
            if (!self._open) return;
            self._onKeyDown(e);
        };
        document.addEventListener('keydown', this._keyHandler, true);
    }

    _getItems() {
        if (!this._data) return [];
        var categories = this._data.categories;
        var query = this._query.toLowerCase();
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
            items.push({ isHeader: true, name: cat.name });
            for (var j = 0; j < catItems.length; j++) items.push(catItems[j]);
        }
        return items;
    }

    _getSelectable() {
        if (!this._data) return [];
        var categories = this._data.categories;
        var query = this._query.toLowerCase();
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
    }

    _renderResults() {
        var resultsEl = this._pickerEl.querySelector('.block-picker-results');
        if (!resultsEl) return;
        resultsEl.innerHTML = '';

        var items = this._getItems();
        var selectableIndex = 0;
        var self = this;

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
            if (item.depth > 0) {
                row.style.paddingLeft = (12 + item.depth * 16) + 'px';
            }
            if (selectableIndex === this._selectedIndex) {
                row.classList.add('block-picker-item--selected');
            }
            row.textContent = item.label;

            row.addEventListener('mousedown', (function(idx) {
                return function(e) {
                    e.preventDefault();
                    e.stopPropagation();
                    self._selectedIndex = idx;
                    self._selectItem();
                };
            })(selectableIndex));

            resultsEl.appendChild(row);
            selectableIndex++;
        }

        var selected = resultsEl.querySelector('.block-picker-item--selected');
        if (selected) selected.scrollIntoView({ block: 'nearest' });

        var queryEl = this._pickerEl.querySelector('.block-picker-query');
        if (queryEl) queryEl.textContent = this._query;
    }

    _onKeyDown(e) {
        e.preventDefault();
        e.stopPropagation();

        if (e.key === 'Escape') { this.close(); return; }

        var selectable = this._getSelectable();
        var maxIndex = selectable.length - 1;

        if (e.key === 'ArrowUp') {
            this._selectedIndex = Math.max(0, this._selectedIndex - 1);
            this._renderResults();
            return;
        }
        if (e.key === 'ArrowDown') {
            this._selectedIndex = Math.min(maxIndex, this._selectedIndex + 1);
            this._renderResults();
            return;
        }
        if (e.key === 'Enter') { this._selectItem(); return; }
        if (e.key === 'Backspace') {
            if (this._query.length > 0) {
                this._query = this._query.slice(0, -1);
                this._selectedIndex = 0;
                this._renderResults();
            }
            return;
        }
        if (e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
            this._query += e.key;
            this._selectedIndex = 0;
            this._renderResults();
        }
    }

    _selectItem() {
        var selectable = this._getSelectable();
        if (selectable.length === 0) return;
        var idx = this._selectedIndex;
        if (idx < 0 || idx >= selectable.length) return;

        var item = selectable[idx];
        var parentId = this.getAttribute('parent-id') || '';

        this.close();

        this.dispatchEvent(new CustomEvent('mcms-block-picker:select', {
            bubbles: true,
            detail: {
                datatypeId: item.id,
                datatypeLabel: item.label,
                datatypeType: item.type,
                parentId: parentId
            }
        }));
    }
}

if (!customElements.get('mcms-block-picker')) {
    customElements.define('mcms-block-picker', McmsBlockPicker);
}
