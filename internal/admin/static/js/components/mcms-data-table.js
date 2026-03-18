/**
 * @module mcms-data-table
 * @description Data table wrapper component providing column sorting and row selection
 * for standard HTML `<table>` elements. Operates in the Light DOM so that global styles
 * (Tailwind utilities, design-token CSS custom properties) apply directly to all injected
 * elements.
 *
 * Sorting is enabled by adding the `sortable` attribute to the host element and
 * `data-sort="column_name"` to sortable `<th>` headers. Clicking a header toggles between
 * ascending and descending order and dispatches an `mcms-table:sort` event. The component
 * does not perform client-side sorting — it emits the event so the server (typically via
 * HTMX) can re-render the table body with the correct order.
 *
 * Selection is enabled by adding the `selectable` attribute. The component injects a
 * "select all" checkbox into the header row and individual row checkboxes into each `<tr>`
 * in `<tbody>`. It manages the tri-state header checkbox (checked, unchecked, indeterminate)
 * and dispatches `mcms-table:select` events with the list of selected row IDs. A
 * `MutationObserver` watches `<tbody>` for HTMX content swaps and automatically adds
 * checkboxes to newly inserted rows.
 *
 * @example <caption>Sortable table</caption>
 * <mcms-data-table sortable>
 *   <table>
 *     <thead>
 *       <tr>
 *         <th data-sort="name">Name</th>
 *         <th data-sort="created_at" data-direction="desc">Created</th>
 *         <th>Actions</th>
 *       </tr>
 *     </thead>
 *     <tbody>
 *       <tr><td>Page One</td><td>2026-01-15</td><td>Edit</td></tr>
 *     </tbody>
 *   </table>
 * </mcms-data-table>
 *
 * @example <caption>Selectable table with row IDs</caption>
 * <mcms-data-table selectable>
 *   <table>
 *     <thead><tr><th>Name</th></tr></thead>
 *     <tbody>
 *       <tr data-id="01ABC"><td>Item A</td></tr>
 *       <tr data-id="01DEF"><td>Item B</td></tr>
 *     </tbody>
 *   </table>
 * </mcms-data-table>
 *
 * @example <caption>Both sorting and selection</caption>
 * <mcms-data-table sortable selectable>
 *   <table>
 *     <thead><tr><th data-sort="title">Title</th></tr></thead>
 *     <tbody>
 *       <tr data-id="01GHI"><td>Draft Post</td></tr>
 *     </tbody>
 *   </table>
 * </mcms-data-table>
 */

/**
 * @event mcms-table:sort
 * @type {CustomEvent}
 * @description Fired when a sortable column header is clicked. The table does not perform
 * client-side sorting; consumers should listen for this event and trigger a server request
 * (e.g., via HTMX) to reload the table body in the new order.
 * @property {Object} detail
 * @property {string} detail.column - The value of the `data-sort` attribute on the clicked `<th>`.
 * @property {string} detail.direction - The new sort direction: `"asc"` or `"desc"`.
 */

/**
 * @event mcms-table:select
 * @type {CustomEvent}
 * @description Fired whenever the set of selected rows changes (individual row checkbox
 * toggled, or "select all" toggled). Consumers can use this to enable/disable bulk action
 * buttons or update a selection count display.
 * @property {Object} detail
 * @property {string[]} detail.selected - Array of `data-id` values from currently selected rows.
 */

/**
 * Data table web component with column sorting and row selection.
 *
 * Wraps a standard HTML `<table>` and enhances it with interactive sorting indicators
 * and checkbox-based row selection. Does not render its own table markup — it expects
 * a `<table>` with `<thead>` and `<tbody>` as children.
 *
 * **Observed attributes:** None (does not use `observedAttributes`).
 *
 * **Host element attributes:**
 * - `sortable` — Enables column sorting. Sortable columns are identified by `<th data-sort="...">`.
 * - `selectable` — Enables row selection with checkboxes.
 *
 * **Data attributes on `<th>` elements (sorting):**
 * - `data-sort` — Column identifier used in the `mcms-table:sort` event detail. Presence of
 *   this attribute makes the header clickable.
 * - `data-direction` — Current sort direction: `"asc"` or `"desc"`. Set by the component on
 *   click; can be pre-set in markup for initial sort state.
 *
 * **CSS classes applied by the component (DUAL pattern: data attribute + class):**
 * - `sort-asc` / `sort-desc` — Applied to the active sort `<th>` for CSS-based sort indicators.
 *
 * **Data attributes on `<tr>` elements (selection):**
 * - `data-id` — Row identifier included in the `mcms-table:select` event's `selected` array.
 *
 * @extends HTMLElement
 * @fires mcms-table:sort
 * @fires mcms-table:select
 *
 * @example
 * const table = document.querySelector('mcms-data-table');
 *
 * table.addEventListener('mcms-table:sort', (e) => {
 *     console.log('Sort by', e.detail.column, e.detail.direction);
 * });
 *
 * table.addEventListener('mcms-table:select', (e) => {
 *     console.log('Selected IDs:', e.detail.selected);
 * });
 *
 * // Programmatic access
 * const ids = table.getSelected();   // ['01ABC', '01DEF']
 * table.clearSelection();            // Unchecks all checkboxes
 */
class McmsDataTable extends HTMLElement {
    constructor() {
        super();

        /**
         * Bound reference to `_onClick` for adding/removing the delegated click listener.
         * @type {Function}
         * @private
         */
        this._boundOnClick = this._onClick.bind(this);

        /**
         * Bound reference to `_onHeaderCheckbox` for the "select all" checkbox change handler.
         * @type {Function}
         * @private
         */
        this._boundOnHeaderCheckbox = this._onHeaderCheckbox.bind(this);

        /**
         * Bound reference to `_onRowCheckbox` for individual row checkbox change handlers.
         * @type {Function}
         * @private
         */
        this._boundOnRowCheckbox = this._onRowCheckbox.bind(this);
    }

    /**
     * Called when the element is inserted into the DOM. Conditionally initializes sorting
     * (if `sortable` attribute is present) and selection (if `selectable` attribute is present).
     */
    connectedCallback() {
        if (this.hasAttribute('sortable')) {
            this._initSorting();
        }
        if (this.hasAttribute('selectable')) {
            this._initSelection();
        }
    }

    /**
     * Called when the element is removed from the DOM. Removes the delegated click listener
     * used for sorting. The MutationObserver (if created) is not explicitly disconnected
     * here since it observes a child node that will also be removed.
     */
    disconnectedCallback() {
        this.removeEventListener('click', this._boundOnClick);
    }

    /**
     * Sets up column sorting by registering a delegated click listener and applying initial
     * sort indicator classes. Makes all `<th>` elements with a `data-sort` attribute
     * clickable (cursor pointer, no text selection). If a `<th>` already has a
     * `data-direction` attribute, applies the corresponding CSS class (`sort-asc` or
     * `sort-desc`) for the initial visual indicator.
     * @private
     */
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

    /**
     * Delegated click handler for sorting. When a `<th>` with `data-sort` is clicked,
     * toggles the sort direction between `"asc"` and `"desc"` (defaulting to `"asc"` if
     * no direction was set), clears sort indicators from all other headers, applies the
     * new indicator, and dispatches the `mcms-table:sort` event.
     *
     * @param {MouseEvent} e - The click event.
     * @private
     */
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

    /**
     * Applies the appropriate CSS class (`sort-asc` or `sort-desc`) to a `<th>` element
     * based on the sort direction. Removes both classes first to ensure a clean state.
     * This follows the DUAL pattern where both a `data-direction` attribute and a CSS
     * class are maintained for styling flexibility.
     *
     * @param {HTMLTableCellElement} th - The table header cell to update.
     * @param {string} direction - The sort direction: `"asc"` or `"desc"`.
     * @private
     */
    // DUAL: data-direction + class (sort-asc/sort-desc used by CSS)
    _applySortClass(th, direction) {
        th.classList.remove('sort-asc', 'sort-desc');
        if (direction === 'asc') {
            th.classList.add('sort-asc');
        } else if (direction === 'desc') {
            th.classList.add('sort-desc');
        }
    }

    /**
     * Sets up row selection by injecting a "select all" checkbox into the first header row
     * and individual row checkboxes into each existing `<tbody>` row. Also starts a
     * `MutationObserver` on `<tbody>` to automatically add checkboxes to rows inserted
     * by HTMX content swaps.
     *
     * Requires the table to have `<table>`, `<thead>` with at least one `<tr>`, and
     * `<tbody>`. If any of these are missing, selection is silently not initialized.
     * @private
     */
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

    /**
     * Injects a checkbox `<td>` as the first cell of a table row. The checkbox's
     * `data-row-id` attribute is set from the row's `data-id` attribute, linking the
     * checkbox to the row's identity for selection tracking.
     *
     * @param {HTMLTableRowElement} row - The `<tr>` element to add a checkbox to.
     * @private
     */
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

    /**
     * Handles the "select all" header checkbox change event. When checked, checks all
     * row checkboxes; when unchecked, unchecks all. Then emits a selection change event.
     *
     * @param {Event} e - The change event from the header checkbox.
     * @private
     */
    _onHeaderCheckbox(e) {
        var checked = e.target.checked;
        var rowCheckboxes = this.querySelectorAll('[data-role="row-checkbox"]');
        for (var i = 0; i < rowCheckboxes.length; i++) {
            rowCheckboxes[i].checked = checked;
        }
        this._emitSelectEvent();
    }

    /**
     * Handles individual row checkbox change events. Updates the header "select all"
     * checkbox to reflect the aggregate state:
     * - **All checked** — header checkbox is checked, not indeterminate.
     * - **None checked** — header checkbox is unchecked, not indeterminate.
     * - **Some checked** — header checkbox is unchecked and set to indeterminate.
     *
     * Then emits a selection change event.
     * @private
     */
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

    /**
     * Collects all checked row IDs and dispatches the `mcms-table:select` custom event.
     * Called by both `_onHeaderCheckbox` and `_onRowCheckbox` after any selection change.
     * @private
     */
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

    /**
     * Creates a `MutationObserver` that watches `<tbody>` for new `<tr>` elements added
     * by HTMX content swaps. When a new row is detected (one without a row checkbox),
     * a checkbox is automatically injected. After any mutation batch, the header "select all"
     * checkbox is reset to unchecked/non-indeterminate since the row set has changed.
     *
     * The observer reference is stored as `this._observer` for potential cleanup.
     *
     * @param {HTMLTableSectionElement} tbody - The `<tbody>` element to observe.
     * @param {HTMLInputElement} headerCheckbox - The "select all" checkbox to reset on content swap.
     * @private
     */
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

    /**
     * Returns the `data-id` values of all currently selected (checked) rows.
     *
     * @returns {string[]} Array of row ID strings. Empty array if no rows are selected.
     *
     * @example
     * const table = document.querySelector('mcms-data-table');
     * const selected = table.getSelected();
     * // => ['01H9ABCDEF', '01H9GHIJKL']
     *
     * if (selected.length > 0) {
     *     fetch('/admin/content/bulk-delete', {
     *         method: 'POST',
     *         body: JSON.stringify({ ids: selected })
     *     });
     * }
     */
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

    /**
     * Clears all row selections and resets the header "select all" checkbox to unchecked.
     * Does not dispatch an `mcms-table:select` event (call is considered a programmatic
     * reset, not a user interaction).
     *
     * @example
     * const table = document.querySelector('mcms-data-table');
     * table.clearSelection();
     * // All checkboxes are now unchecked
     */
    clearSelection() {
        var all = this.querySelectorAll('[data-role="row-checkbox"], [data-role="select-all"]');
        for (var i = 0; i < all.length; i++) {
            all[i].checked = false;
            all[i].indeterminate = false;
        }
    }
}

customElements.define('mcms-data-table', McmsDataTable);
