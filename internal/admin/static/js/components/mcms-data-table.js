// <mcms-data-table> -- Table wrapper with sorting and selection (Light DOM)
// Supports: sortable, selectable attributes
// Dispatches: mcms-table:sort, mcms-table:select custom events
class McmsDataTable extends HTMLElement {
    constructor() {
        super();
        this._boundOnClick = this._onClick.bind(this);
        this._boundOnHeaderCheckbox = this._onHeaderCheckbox.bind(this);
        this._boundOnRowCheckbox = this._onRowCheckbox.bind(this);
    }

    connectedCallback() {
        if (this.hasAttribute('sortable')) {
            this._initSorting();
        }
        if (this.hasAttribute('selectable')) {
            this._initSelection();
        }
    }

    disconnectedCallback() {
        this.removeEventListener('click', this._boundOnClick);
    }

    _initSorting() {
        this.addEventListener('click', this._boundOnClick);
        // Apply initial sort indicators from data attributes
        var headers = this.querySelectorAll('th[data-sort]');
        for (var i = 0; i < headers.length; i++) {
            var th = headers[i];
            th.style.cursor = 'pointer';
            th.style.userSelect = 'none';
            var dir = th.dataset.direction;
            if (dir) {
                this._applySortClass(th, dir);
            }
        }
    }

    _onClick(e) {
        var th = e.target.closest('th[data-sort]');
        if (!th) return;

        var column = th.dataset.sort;
        var current = th.dataset.direction || '';
        var direction = current === 'asc' ? 'desc' : 'asc';

        // Clear all sort directions
        var headers = this.querySelectorAll('th[data-sort]');
        for (var i = 0; i < headers.length; i++) {
            headers[i].removeAttribute('data-direction');
            // DUAL: data-direction + class (sort-asc/sort-desc used by CSS)
            headers[i].classList.remove('sort-asc', 'sort-desc');
        }

        // Set new direction
        th.dataset.direction = direction;
        this._applySortClass(th, direction);

        this.dispatchEvent(new CustomEvent('mcms-table:sort', {
            bubbles: true,
            detail: { column: column, direction: direction }
        }));
    }

    // DUAL: data-direction + class (sort-asc/sort-desc used by CSS)
    _applySortClass(th, direction) {
        th.classList.remove('sort-asc', 'sort-desc');
        if (direction === 'asc') {
            th.classList.add('sort-asc');
        } else if (direction === 'desc') {
            th.classList.add('sort-desc');
        }
    }

    _initSelection() {
        var table = this.querySelector('table');
        if (!table) return;

        var thead = table.querySelector('thead');
        var tbody = table.querySelector('tbody');
        if (!thead || !tbody) return;

        // Add header checkbox
        var headerRow = thead.querySelector('tr');
        if (!headerRow) return;

        var headerTh = document.createElement('th');
        headerTh.className = 'w-10 px-4 py-3';
        var headerCheckbox = document.createElement('input');
        headerCheckbox.type = 'checkbox';
        headerCheckbox.className = 'rounded border-[var(--color-border)]';
        headerCheckbox.setAttribute('data-role', 'select-all');
        headerCheckbox.setAttribute('aria-label', 'Select all rows');
        headerCheckbox.addEventListener('change', this._boundOnHeaderCheckbox);
        headerTh.appendChild(headerCheckbox);
        headerRow.insertBefore(headerTh, headerRow.firstChild);

        // Add row checkboxes
        var rows = tbody.querySelectorAll('tr');
        for (var i = 0; i < rows.length; i++) {
            this._addRowCheckbox(rows[i]);
        }

        // Watch for HTMX swaps that replace tbody rows
        this._observeBodyChanges(tbody, headerCheckbox);
    }

    _addRowCheckbox(row) {
        var td = document.createElement('td');
        td.className = 'w-10 px-4 py-3';
        var checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.className = 'rounded border-[var(--color-border)]';
        checkbox.setAttribute('data-role', 'row-checkbox');
        checkbox.setAttribute('aria-label', 'Select row');

        var rowId = row.getAttribute('data-id') || '';
        checkbox.setAttribute('data-row-id', rowId);
        checkbox.addEventListener('change', this._boundOnRowCheckbox);

        td.appendChild(checkbox);
        row.insertBefore(td, row.firstChild);
    }

    _onHeaderCheckbox(e) {
        var checked = e.target.checked;
        var rowCheckboxes = this.querySelectorAll('[data-role="row-checkbox"]');
        for (var i = 0; i < rowCheckboxes.length; i++) {
            rowCheckboxes[i].checked = checked;
        }
        this._emitSelectEvent();
    }

    _onRowCheckbox() {
        // Update header checkbox state
        var rowCheckboxes = this.querySelectorAll('[data-role="row-checkbox"]');
        var headerCheckbox = this.querySelector('[data-role="select-all"]');
        if (!headerCheckbox) return;

        var allChecked = true;
        var noneChecked = true;
        for (var i = 0; i < rowCheckboxes.length; i++) {
            if (rowCheckboxes[i].checked) {
                noneChecked = false;
            } else {
                allChecked = false;
            }
        }

        headerCheckbox.checked = allChecked && rowCheckboxes.length > 0;
        headerCheckbox.indeterminate = !allChecked && !noneChecked;
        this._emitSelectEvent();
    }

    _emitSelectEvent() {
        var selected = [];
        var rowCheckboxes = this.querySelectorAll('[data-role="row-checkbox"]');
        for (var i = 0; i < rowCheckboxes.length; i++) {
            if (rowCheckboxes[i].checked) {
                var rowId = rowCheckboxes[i].getAttribute('data-row-id');
                if (rowId) {
                    selected.push(rowId);
                }
            }
        }
        this.dispatchEvent(new CustomEvent('mcms-table:select', {
            bubbles: true,
            detail: { selected: selected }
        }));
    }

    _observeBodyChanges(tbody, headerCheckbox) {
        // After HTMX swaps new rows into tbody, add checkboxes to them
        var self = this;
        var observer = new MutationObserver(function(mutations) {
            for (var m = 0; m < mutations.length; m++) {
                var addedNodes = mutations[m].addedNodes;
                for (var n = 0; n < addedNodes.length; n++) {
                    var node = addedNodes[n];
                    if (node.nodeType === 1 && node.tagName === 'TR' && !node.querySelector('[data-role="row-checkbox"]')) {
                        self._addRowCheckbox(node);
                    }
                }
            }
            // Reset header checkbox after content swap
            if (headerCheckbox) {
                headerCheckbox.checked = false;
                headerCheckbox.indeterminate = false;
            }
        });
        observer.observe(tbody, { childList: true });
        // Store observer reference for potential cleanup
        this._observer = observer;
    }

    // Public API: get currently selected row IDs
    getSelected() {
        var selected = [];
        var rowCheckboxes = this.querySelectorAll('[data-role="row-checkbox"]:checked');
        for (var i = 0; i < rowCheckboxes.length; i++) {
            var rowId = rowCheckboxes[i].getAttribute('data-row-id');
            if (rowId) {
                selected.push(rowId);
            }
        }
        return selected;
    }

    // Public API: clear all selections
    clearSelection() {
        var all = this.querySelectorAll('[data-role="row-checkbox"], [data-role="select-all"]');
        for (var i = 0; i < all.length; i++) {
            all[i].checked = false;
            all[i].indeterminate = false;
        }
    }
}

customElements.define('mcms-data-table', McmsDataTable);
