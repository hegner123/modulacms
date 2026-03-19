/**
 * @module mcms-validation-wizard
 * @description Visual builder for composable validation rules, used in the admin panel
 * to configure field validation without writing JSON by hand.
 *
 * Provides a three-panel UI:
 * 1. **Rule Builder** -- a tree view of configured rules (individual rules and
 *    AND/OR groups) with add/delete controls.
 * 2. **Rule Editor** -- a detail panel for editing the selected rule's parameters
 *    (operator, value/class, comparison, negate, custom message).
 * 3. **Test Panel** -- a live test input that evaluates all rules against a user-provided
 *    value and shows pass/fail results in real time.
 *
 * Rules are serialized as JSON and synchronized to a hidden `<textarea>` identified
 * by the `textarea-id` attribute. The JSON schema follows the CMS validation config
 * format: `{ "rules": [ { "rule": { "op": "...", ... } }, { "group": { "all_of": [...] } } ] }`.
 *
 * Supported rule operators:
 * - `required` -- value must be non-empty
 * - `contains` -- value must contain a literal string or character class
 * - `starts_with` -- value must start with a literal string or character class
 * - `ends_with` -- value must end with a literal string or character class
 * - `equals` -- value must exactly equal a string
 * - `length` -- Unicode character count compared against a threshold
 * - `count` -- count of substring occurrences or character class matches compared against a threshold
 * - `range` -- numeric value compared against a threshold
 * - `item_count` -- comma-separated or JSON array item count compared against a threshold
 * - `one_of` -- value must be one of an enumerated set
 *
 * Grouping operators:
 * - `all_of` (AND) -- all child rules must pass
 * - `any_of` (OR) -- at least one child rule must pass
 *
 * This is a Light DOM web component -- all rendered elements are direct children of the
 * host element, making them accessible to global CSS.
 *
 * @example
 * <!-- Validation wizard bound to a textarea -->
 * <textarea id="validation-json" class="sr-only"></textarea>
 * <mcms-validation-wizard
 *     value='{"rules":[{"rule":{"op":"required"}},{"rule":{"op":"length","cmp":"gte","n":3}}]}'
 *     textarea-id="validation-json"
 * ></mcms-validation-wizard>
 *
 * @example
 * <!-- Empty wizard (no initial rules) -->
 * <textarea id="val-config" class="sr-only"></textarea>
 * <mcms-validation-wizard textarea-id="val-config"></mcms-validation-wizard>
 *
 * @example
 * <!-- Listen for rule changes -->
 * <mcms-validation-wizard id="wizard" textarea-id="val-json"></mcms-validation-wizard>
 * <script>
 *   document.getElementById('wizard').addEventListener('validation-changed', (e) => {
 *     console.log('Updated validation JSON:', e.detail.json);
 *   });
 * </script>
 */

/**
 * @event validation-changed
 * @description Fired whenever the validation rule configuration changes (rule added,
 * deleted, or edited). Bubbles up through the DOM.
 * @type {CustomEvent}
 * @property {Object} detail
 * @property {string} detail.json - The complete validation configuration as a JSON string
 *   in the format `{ "rules": [...] }`, with internal `_id` fields stripped.
 */

/**
 * Visual validation rule builder web component.
 *
 * Provides an interactive UI for constructing composable validation rules as a tree
 * structure. Rules can be individual operations (required, contains, length, etc.)
 * or groups (all_of/any_of) that contain nested rules. The component synchronizes
 * the rule configuration to a `<textarea>` element and dispatches events on changes.
 *
 * @extends HTMLElement
 *
 * @attr {string} value - Initial validation configuration as a JSON string. Expected
 *   format: `{ "rules": [...] }`. Empty string, `"{}"`, or invalid JSON starts with
 *   an empty rule set (and shows a warning if the JSON was non-empty but unparseable).
 * @attr {string} textarea-id - The `id` of the `<textarea>` element to sync the
 *   serialized rule JSON to. Updated on every rule change.
 *
 * @fires validation-changed
 */
class McmsValidationWizard extends HTMLElement {
    constructor() {
        super();

        /**
         * Guard flag to prevent re-building the DOM on repeated connectedCallback calls
         * (e.g., when the element is moved in the DOM).
         * @type {boolean}
         * @private
         */
        this._built = false;

        /**
         * The root array of rule entries. Each entry is either `{ rule: {...}, _id: N }`
         * or `{ group: { all_of: [...] | any_of: [...] }, _id: N }`.
         * @type {Array<Object>}
         * @private
         */
        this._rules = [];

        /**
         * Auto-incrementing counter for assigning unique IDs to rule entries.
         * Used internally for selection, lookup, and DOM keying.
         * @type {number}
         * @private
         */
        this._nextId = 1;

        /**
         * The `_id` of the currently selected rule entry, or `null` if none is selected.
         * Controls which rule's editor panel is displayed.
         * @type {number|null}
         * @private
         */
        this._selectedId = null;

        /**
         * When adding a rule from a group's "+" button, this holds the group's `_id`
         * so the new rule is inserted into that group's children instead of the root.
         * Reset to `null` after insertion.
         * @type {number|null}
         * @private
         */
        this._insertTargetId = null;

        /**
         * The current value in the test input field.
         * @type {string}
         * @private
         */
        this._testValue = '';

        /**
         * Debounce timer ID for the test input. Test results are re-evaluated
         * 200ms after the user stops typing.
         * @type {number|null}
         * @private
         */
        this._testTimer = null;

        /**
         * Reference to the root container div.
         * @type {HTMLDivElement|null}
         * @private
         */
        this._container = null;

        /**
         * Reference to the rule list container where rule items are rendered.
         * @type {HTMLDivElement|null}
         * @private
         */
        this._ruleListEl = null;

        /**
         * Reference to the editor content container where rule editing fields are rendered.
         * @type {HTMLDivElement|null}
         * @private
         */
        this._editorEl = null;

        /**
         * Reference to the test results container where pass/fail indicators are rendered.
         * @type {HTMLDivElement|null}
         * @private
         */
        this._testResultsEl = null;

        /**
         * Reference to the warning banner element shown when initial JSON parsing fails.
         * @type {HTMLDivElement|null}
         * @private
         */
        this._warningEl = null;
    }

    /**
     * Lifecycle callback invoked when the element is inserted into the DOM.
     *
     * Parses the initial `value` attribute, builds the three-panel UI (rule list,
     * editor, test panel), and renders all sections. Only runs once per element
     * instance (guarded by `_built` flag).
     */
    connectedCallback() {
        if (this._built) return;
        this._built = true;
        this._parseInitialValue();
        this._build();
        this._renderAll();
    }

    /**
     * Parses the `value` attribute as JSON and populates the internal `_rules` array.
     * If the value is empty, `"{}"`, or invalid JSON, starts with an empty rule set.
     * Sets `_showWarning` to `true` if the attribute contained non-empty, unparseable content,
     * so the warning banner is displayed.
     *
     * @private
     */
    _parseInitialValue() {
        var raw = this.getAttribute('value') || '';
        if (!raw || raw === '{}') {
            this._rules = [];
            return;
        }
        try {
            var parsed = JSON.parse(raw);
            if (!parsed.rules || !Array.isArray(parsed.rules)) {
                this._rules = [];
                this._showWarning = true;
                return;
            }
            this._rules = parsed.rules;
            this._assignIds(this._rules);
        } catch (e) {
            this._rules = [];
            this._showWarning = true;
        }
    }

    // ========================================
    // ID System
    // ========================================

    /**
     * Recursively assigns unique `_id` values to an array of rule entries and their
     * nested group children. Used during initial parsing and is not called after
     * construction (new entries get IDs in `_addRule`).
     *
     * @param {Array<Object>} entries - The array of rule entries to assign IDs to.
     * @private
     */
    _assignIds(entries) {
        for (var i = 0; i < entries.length; i++) {
            entries[i]._id = this._nextId++;
            if (entries[i].group) {
                var g = entries[i].group;
                if (g.all_of) this._assignIds(g.all_of);
                if (g.any_of) this._assignIds(g.any_of);
            }
        }
    }

    /**
     * Recursively searches for a rule entry by its `_id` in a nested entry array.
     *
     * @param {Array<Object>} entries - The array of entries to search.
     * @param {number} id - The `_id` value to find.
     * @returns {{ entries: Array<Object>, index: number }|null} An object containing the
     *   parent array and the index of the found entry, or `null` if not found.
     * @private
     */
    _findById(entries, id) {
        for (var i = 0; i < entries.length; i++) {
            if (entries[i]._id === id) return { entries: entries, index: i };
            if (entries[i].group) {
                var g = entries[i].group;
                var children = g.all_of || g.any_of;
                if (children) {
                    var found = this._findById(children, id);
                    if (found) return found;
                }
            }
        }
        return null;
    }

    /**
     * Retrieves a rule entry object by its `_id`.
     *
     * @param {number} id - The `_id` value to look up.
     * @returns {Object|null} The entry object, or `null` if not found.
     * @private
     */
    _getEntry(id) {
        var loc = this._findById(this._rules, id);
        return loc ? loc.entries[loc.index] : null;
    }

    // ========================================
    // Build DOM
    // ========================================

    /**
     * Constructs the entire component DOM structure: warning banner, rule list section
     * with header and "+ Add" button, rule editor section (initially hidden), and test
     * panel section with live input and results area.
     *
     * Attaches a document-level click listener to close any open add-rule dropdowns
     * when clicking outside of them.
     *
     * @private
     */
    _build() {
        var container = document.createElement('div');
        container.className = 'flex flex-col gap-4';

        // Warning banner
        var warning = document.createElement('div');
        warning.className = 'rounded-md border border-[var(--color-danger)] bg-[var(--color-danger)]/10 px-4 py-3 text-sm text-[var(--color-danger)]';
        warning.textContent = 'Could not parse existing validation config. Starting with empty rules.';
        warning.hidden = !this._showWarning;
        container.appendChild(warning);
        this._warningEl = warning;

        // Rule list section
        var listSection = document.createElement('div');
        listSection.className = 'rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)]';

        var listHeader = document.createElement('div');
        listHeader.className = 'flex items-center justify-between border-b border-[var(--color-border)] px-4 py-3';

        var listTitle = document.createElement('span');
        listTitle.className = 'text-sm font-semibold text-[var(--color-text)]';
        listTitle.textContent = 'Rule Builder';
        listHeader.appendChild(listTitle);

        var addBtn = document.createElement('div');
        addBtn.className = 'relative';
        addBtn.setAttribute('data-add-wrapper', '');

        var addButton = document.createElement('button');
        addButton.type = 'button';
        addButton.className = 'inline-flex items-center justify-center rounded-md px-3 py-1.5 text-xs font-medium bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors cursor-pointer border-none';
        addButton.textContent = '+ Add';
        addButton.addEventListener('click', function(e) {
            e.stopPropagation();
            this._insertTargetId = null;
            this._toggleDropdown(addBtn);
        }.bind(this));
        addBtn.appendChild(addButton);
        addBtn.appendChild(this._buildDropdown());
        listHeader.appendChild(addBtn);

        listSection.appendChild(listHeader);

        var ruleList = document.createElement('div');
        ruleList.className = 'px-4 py-2';
        listSection.appendChild(ruleList);
        this._ruleListEl = ruleList;

        container.appendChild(listSection);

        // Rule editor section
        var editorSection = document.createElement('div');
        editorSection.className = 'rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)]';
        editorSection.hidden = true;

        var editorHeader = document.createElement('div');
        editorHeader.className = 'flex items-center justify-between border-b border-[var(--color-border)] px-4 py-3';
        var editorTitle = document.createElement('span');
        editorTitle.className = 'text-sm font-semibold text-[var(--color-text)]';
        editorTitle.textContent = 'Rule Editor';
        editorHeader.appendChild(editorTitle);
        editorSection.appendChild(editorHeader);

        var editorContent = document.createElement('div');
        editorContent.className = 'px-4 py-3 flex flex-col gap-3';
        editorSection.appendChild(editorContent);
        this._editorEl = editorContent;
        this._editorSection = editorSection;

        container.appendChild(editorSection);

        // Test panel section
        var testSection = document.createElement('div');
        testSection.className = 'rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)]';

        var testHeader = document.createElement('div');
        testHeader.className = 'flex items-center justify-between border-b border-[var(--color-border)] px-4 py-3';
        var testTitle = document.createElement('span');
        testTitle.className = 'text-sm font-semibold text-[var(--color-text)]';
        testTitle.textContent = 'Test';
        testHeader.appendChild(testTitle);
        testSection.appendChild(testHeader);

        var testInput = document.createElement('input');
        testInput.type = 'text';
        testInput.className = 'mx-4 mt-3 rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none';
        testInput.placeholder = 'Enter a test value...';
        testInput.addEventListener('input', function(e) {
            this._testValue = e.target.value;
            clearTimeout(this._testTimer);
            this._testTimer = setTimeout(function() {
                this._renderTestResults();
            }.bind(this), 200);
        }.bind(this));
        testSection.appendChild(testInput);

        var testResults = document.createElement('div');
        testResults.className = 'px-4 py-3 flex flex-col gap-1';
        testSection.appendChild(testResults);
        this._testResultsEl = testResults;

        container.appendChild(testSection);
        this._container = container;
        this.appendChild(container);

        // Close dropdowns when clicking outside
        document.addEventListener('click', function(e) {
            if (!e.target.closest('[data-add-wrapper]')) {
                var dropdowns = this.querySelectorAll('[data-dropdown][data-open]');
                for (var i = 0; i < dropdowns.length; i++) {
                    dropdowns[i].removeAttribute('data-open');
                    dropdowns[i].classList.add('hidden'); // DUAL: data-open + class
                }
            }
        }.bind(this));
    }

    /**
     * Creates a dropdown menu element containing buttons for each available rule operator
     * and group type. Includes a visual separator between individual rules and group types.
     *
     * Available operators: required, contains, starts_with, ends_with, equals, length,
     * count, range, item_count, one_of. Group types: all_of (AND), any_of (OR).
     *
     * Each button click calls `_addRule()` with the operator name and closes all open
     * dropdowns.
     *
     * @returns {HTMLDivElement} The dropdown menu element with `[data-dropdown]` attribute.
     * @private
     */
    _buildDropdown() {
        var dropdown = document.createElement('div');
        dropdown.className = 'absolute right-0 top-full z-10 mt-1 hidden min-w-[10rem] rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] py-1 shadow-lg';
        dropdown.setAttribute('data-dropdown', '');

        var ops = [
            { op: 'required', label: 'Required' },
            { op: 'contains', label: 'Contains' },
            { op: 'starts_with', label: 'Starts with' },
            { op: 'ends_with', label: 'Ends with' },
            { op: 'equals', label: 'Equals' },
            { op: 'length', label: 'Length' },
            { op: 'count', label: 'Count' },
            { op: 'range', label: 'Range' },
            { op: 'item_count', label: 'Item count' },
            { op: 'one_of', label: 'One of' },
            { op: '_sep', label: '' },
            { op: 'all_of', label: 'All must pass (AND)' },
            { op: 'any_of', label: 'Any must pass (OR)' },
        ];

        for (var i = 0; i < ops.length; i++) {
            var item = ops[i];
            if (item.op === '_sep') {
                var sep = document.createElement('div');
                sep.className = 'my-1 border-t border-[var(--color-border)]';
                dropdown.appendChild(sep);
                continue;
            }
            var btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'block w-full text-left px-4 py-1.5 text-sm text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] cursor-pointer bg-transparent border-none';
            btn.textContent = item.label;
            btn.dataset.op = item.op;
            btn.addEventListener('click', function(e) {
                e.stopPropagation();
                var op = e.target.dataset.op;
                this._addRule(op);
                // Close all dropdowns
                var dropdowns = this.querySelectorAll('[data-dropdown][data-open]');
                for (var j = 0; j < dropdowns.length; j++) {
                    dropdowns[j].removeAttribute('data-open');
                    dropdowns[j].classList.add('hidden'); // DUAL: data-open + class
                }
            }.bind(this));
            dropdown.appendChild(btn);
        }

        return dropdown;
    }

    /**
     * Toggles the visibility of a dropdown menu within an add-wrapper element.
     * Closes all other open dropdowns first. Uses a dual state approach: the
     * `data-open` attribute tracks logical state while the `hidden` class controls
     * CSS visibility.
     *
     * @param {HTMLElement} wrapper - The `[data-add-wrapper]` element containing
     *   the dropdown to toggle.
     * @private
     */
    _toggleDropdown(wrapper) {
        var dd = wrapper.querySelector('[data-dropdown]');
        if (!dd) return;
        // Close all other dropdowns first
        var allDd = this.querySelectorAll('[data-dropdown][data-open]');
        for (var i = 0; i < allDd.length; i++) {
            if (allDd[i] !== dd) {
                allDd[i].removeAttribute('data-open');
                allDd[i].classList.add('hidden'); // DUAL: data-open + class
            }
        }
        var isOpen = dd.hasAttribute('data-open');
        if (isOpen) {
            dd.removeAttribute('data-open');
            dd.classList.add('hidden'); // DUAL: data-open + class
        } else {
            dd.setAttribute('data-open', '');
            dd.classList.remove('hidden'); // DUAL: data-open + class
        }
    }

    // ========================================
    // Add / Delete Rules
    // ========================================

    /**
     * Creates a new rule entry with the specified operator and adds it to the rule tree.
     *
     * For group operators (`all_of`, `any_of`), creates a group entry with an empty
     * children array. For individual operators, creates a rule entry with sensible
     * defaults for that operator type (e.g., `cmp: 'gte'`, `n: 0` for length rules).
     *
     * If `_insertTargetId` is set (from a group's "+" button), the new entry is added
     * to that group's children. Otherwise, it is appended to the root rule array.
     *
     * After adding, selects the new entry and triggers a full re-render.
     *
     * @param {string} op - The operator name (e.g., `"required"`, `"contains"`,
     *   `"all_of"`, `"any_of"`).
     * @private
     */
    _addRule(op) {
        var entry;
        if (op === 'all_of' || op === 'any_of') {
            var group = {};
            group[op] = [];
            entry = { group: group };
        } else {
            var rule = { op: op };
            // Set defaults based on op
            if (op === 'length' || op === 'item_count') {
                rule.cmp = 'gte';
                rule.n = 0;
            } else if (op === 'count') {
                rule.cmp = 'gte';
                rule.n = 0;
                rule.class = 'uppercase';
            } else if (op === 'range') {
                rule.cmp = 'gte';
                rule.n = 0;
            } else if (op === 'contains' || op === 'starts_with' || op === 'ends_with') {
                rule.value = '';
            } else if (op === 'equals') {
                rule.value = '';
            } else if (op === 'one_of') {
                rule.values = [];
            }
            entry = { rule: rule };
        }

        entry._id = this._nextId++;

        if (this._insertTargetId !== null) {
            // Insert into a group
            var parent = this._getEntry(this._insertTargetId);
            if (parent && parent.group) {
                var children = parent.group.all_of || parent.group.any_of;
                if (children) {
                    children.push(entry);
                }
            }
        } else {
            this._rules.push(entry);
        }

        this._insertTargetId = null;
        this._selectedId = entry._id;
        this._onRulesChanged();
    }

    /**
     * Deletes a rule entry by its `_id` from the rule tree (including from within groups).
     * If the deleted entry was selected, clears the selection. Triggers a full re-render.
     *
     * @param {number} id - The `_id` of the entry to delete.
     * @private
     */
    _deleteRule(id) {
        var loc = this._findById(this._rules, id);
        if (!loc) return;
        loc.entries.splice(loc.index, 1);
        if (this._selectedId === id) {
            this._selectedId = null;
        }
        this._onRulesChanged();
    }

    // ========================================
    // Render
    // ========================================

    /**
     * Re-renders all three panels: rule list, editor, and test results.
     *
     * @private
     */
    _renderAll() {
        this._renderRuleList();
        this._renderEditor();
        this._renderTestResults();
    }

    /**
     * Renders the rule list panel. If no rules exist, shows an empty-state message.
     * Otherwise, delegates to `_renderEntries()` to build the nested rule tree.
     *
     * @private
     */
    _renderRuleList() {
        if (!this._ruleListEl) return;
        this._ruleListEl.innerHTML = '';
        if (this._rules.length === 0) {
            var empty = document.createElement('div');
            empty.className = 'py-6 text-center text-sm text-[var(--color-text-dim)]';
            empty.textContent = 'No rules configured. Click "+ Add" to create a rule.';
            this._ruleListEl.appendChild(empty);
            return;
        }
        this._renderEntries(this._rules, this._ruleListEl, 0);
    }

    /**
     * Recursively renders an array of rule entries into a container element.
     * Delegates to `_renderRuleItem()` for individual rules and `_renderGroupItem()`
     * for groups.
     *
     * @param {Array<Object>} entries - The array of rule entries to render.
     * @param {HTMLElement} container - The DOM element to append rendered items to.
     * @param {number} depth - The current nesting depth (used for visual indentation
     *   in groups).
     * @private
     */
    _renderEntries(entries, container, depth) {
        for (var i = 0; i < entries.length; i++) {
            var entry = entries[i];
            if (entry.rule) {
                this._renderRuleItem(entry, container);
            } else if (entry.group) {
                this._renderGroupItem(entry, container, depth);
            }
        }
    }

    /**
     * Renders a single rule entry as a row in the rule list with a summary label
     * and a delete button. Clicking the label selects the rule and opens its editor.
     * Highlights the row when it is the currently selected entry.
     *
     * Uses dual state approach: `data-selected` attribute + `selected` class.
     *
     * @param {Object} entry - The rule entry object with `{ rule: {...}, _id: N }`.
     * @param {HTMLElement} container - The DOM element to append the row to.
     * @private
     */
    _renderRuleItem(entry, container) {
        var row = document.createElement('div');
        row.className = 'flex items-center justify-between rounded-md px-3 py-2 text-sm transition-colors hover:bg-[var(--color-surface-hover)]';
        if (entry._id === this._selectedId) {
            row.setAttribute('data-selected', '');
            row.classList.add('selected'); // DUAL: data-selected + class
        }

        var label = document.createElement('span');
        label.className = 'flex-1 cursor-pointer text-[var(--color-text)]';
        label.textContent = this._ruleSummary(entry.rule);
        label.addEventListener('click', function() {
            this._selectedId = entry._id;
            this._renderAll();
        }.bind(this));
        row.appendChild(label);

        var del = document.createElement('button');
        del.type = 'button';
        del.className = 'ml-2 cursor-pointer bg-transparent border-none text-[var(--color-text-muted)] hover:text-[var(--color-danger)] text-lg leading-none';
        del.textContent = '\u00d7';
        del.title = 'Remove rule';
        del.addEventListener('click', function(e) {
            e.stopPropagation();
            this._deleteRule(entry._id);
        }.bind(this));
        row.appendChild(del);

        container.appendChild(row);
    }

    /**
     * Renders a group entry (all_of or any_of) as a bordered card with a header
     * showing the group type, a "+" button for adding child rules, a delete button,
     * and a body containing the nested child entries (or an empty-state message).
     *
     * Uses dual state approach: `data-selected` attribute + `selected` class.
     *
     * @param {Object} entry - The group entry object with `{ group: { all_of: [...] | any_of: [...] }, _id: N }`.
     * @param {HTMLElement} container - The DOM element to append the group card to.
     * @param {number} depth - The current nesting depth for recursive rendering.
     * @private
     */
    _renderGroupItem(entry, container, depth) {
        var g = entry.group;
        var isAllOf = !!g.all_of;
        var children = isAllOf ? g.all_of : g.any_of;

        var group = document.createElement('div');
        group.className = 'rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] my-1';
        if (entry._id === this._selectedId) {
            group.setAttribute('data-selected', '');
            group.classList.add('selected'); // DUAL: data-selected + class
        }

        var header = document.createElement('div');
        header.className = 'flex items-center justify-between px-3 py-2 border-b border-[var(--color-border)]';

        var label = document.createElement('span');
        label.className = 'text-sm font-medium text-[var(--color-text)] cursor-pointer';
        label.textContent = isAllOf ? 'All must pass' : 'Any must pass';
        label.addEventListener('click', function() {
            this._selectedId = entry._id;
            this._renderAll();
        }.bind(this));
        header.appendChild(label);

        var actions = document.createElement('div');
        actions.className = 'flex items-center gap-1';

        var addWrapper = document.createElement('div');
        addWrapper.className = 'relative';
        addWrapper.setAttribute('data-add-wrapper', '');

        var addBtn = document.createElement('button');
        addBtn.type = 'button';
        addBtn.className = 'inline-flex items-center justify-center rounded-md px-2 py-1 text-xs font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] transition-colors cursor-pointer border-none';
        addBtn.textContent = '+';
        addBtn.title = 'Add rule to group';
        addBtn.addEventListener('click', function(e) {
            e.stopPropagation();
            this._insertTargetId = entry._id;
            this._toggleDropdown(addWrapper);
        }.bind(this));
        addWrapper.appendChild(addBtn);
        addWrapper.appendChild(this._buildDropdown());
        actions.appendChild(addWrapper);

        var del = document.createElement('button');
        del.type = 'button';
        del.className = 'ml-2 cursor-pointer bg-transparent border-none text-[var(--color-text-muted)] hover:text-[var(--color-danger)] text-lg leading-none';
        del.textContent = '\u00d7';
        del.title = 'Remove group';
        del.addEventListener('click', function(e) {
            e.stopPropagation();
            this._deleteRule(entry._id);
        }.bind(this));
        actions.appendChild(del);

        header.appendChild(actions);
        group.appendChild(header);

        var body = document.createElement('div');
        body.className = 'px-3 py-2';
        if (children.length === 0) {
            var empty = document.createElement('div');
            empty.className = 'py-3 text-center text-xs text-[var(--color-text-dim)]';
            empty.textContent = 'Empty group. Click "+" to add rules.';
            body.appendChild(empty);
        } else {
            this._renderEntries(children, body, depth + 1);
        }
        group.appendChild(body);

        container.appendChild(group);
    }

    // ========================================
    // Rule Summary
    // ========================================

    /**
     * Generates a human-readable summary string for a rule, used as the label in
     * the rule list. Includes negation prefix, operator name, target value/class,
     * comparison symbol, and threshold as appropriate.
     *
     * @param {Object} rule - The rule object with `op` and operator-specific fields.
     * @returns {string} A concise summary string (e.g., `"length >= 3"`, `"not contains \"foo\""`).
     * @private
     */
    _ruleSummary(rule) {
        if (!rule) return '???';
        var neg = rule.negate ? 'not ' : '';
        switch (rule.op) {
            case 'required':
                return 'required';
            case 'contains':
                return neg + 'contains ' + this._targetLabel(rule);
            case 'starts_with':
                return neg + 'starts with ' + this._targetLabel(rule);
            case 'ends_with':
                return neg + 'ends with ' + this._targetLabel(rule);
            case 'equals':
                return neg + 'equals "' + (rule.value || '') + '"';
            case 'length':
                return 'length ' + this._cmpSymbol(rule.cmp) + ' ' + this._formatN(rule.n);
            case 'count':
                return 'count ' + this._targetLabel(rule) + ' ' + this._cmpSymbol(rule.cmp) + ' ' + this._formatN(rule.n);
            case 'range':
                return 'range ' + this._cmpSymbol(rule.cmp) + ' ' + this._formatN(rule.n);
            case 'item_count':
                return 'item count ' + this._cmpSymbol(rule.cmp) + ' ' + this._formatN(rule.n);
            case 'one_of':
                var vals = (rule.values || []).join(', ');
                return neg + 'one of: ' + (vals || '(none)');
            default:
                return rule.op;
        }
    }

    /**
     * Returns a display label for the rule's target -- either a quoted literal value
     * or the character class name.
     *
     * @param {Object} rule - The rule object with optional `value` and `class` fields.
     * @returns {string} Display label (e.g., `'"hello"'`, `'uppercase'`, `'(unset)'`).
     * @private
     */
    _targetLabel(rule) {
        if (rule.value) return '"' + rule.value + '"';
        if (rule.class) return rule.class;
        return '(unset)';
    }

    /**
     * Converts a comparison operator code to its Unicode symbol for display.
     *
     * @param {string} cmp - The comparison code (`"eq"`, `"neq"`, `"gt"`, `"gte"`,
     *   `"lt"`, `"lte"`).
     * @returns {string} The Unicode symbol (e.g., `"="`, `"\u2260"`, `">"`, `"\u2265"`,
     *   `"<"`, `"\u2264"`) or the raw code if unknown.
     * @private
     */
    _cmpSymbol(cmp) {
        var symbols = { eq: '=', neq: '\u2260', gt: '>', gte: '\u2265', lt: '<', lte: '\u2264' };
        return symbols[cmp] || cmp || '?';
    }

    /**
     * Formats a numeric value for display. Returns `"?"` for null/undefined,
     * removes trailing `.0` for integers, and converts to string otherwise.
     *
     * @param {number|null|undefined} n - The number to format.
     * @returns {string} The formatted string.
     * @private
     */
    _formatN(n) {
        if (n === null || n === undefined) return '?';
        if (n === Math.floor(n)) return String(Math.floor(n));
        return String(n);
    }

    // ========================================
    // Rule Editor
    // ========================================

    /**
     * Renders the rule editor panel for the currently selected entry.
     *
     * If no entry is selected, hides the editor section. For group entries, shows a
     * read-only description of the group type (AND/OR). For rule entries, renders
     * the appropriate editing fields based on the operator: value/class toggle,
     * comparison select, numeric input, negate checkbox, values textarea, and always
     * a custom message input at the bottom.
     *
     * @private
     */
    _renderEditor() {
        if (!this._editorEl) return;
        this._editorEl.innerHTML = '';

        if (this._selectedId === null) {
            this._editorSection.hidden = true;
            return;
        }

        var entry = this._getEntry(this._selectedId);
        if (!entry) {
            this._editorSection.hidden = true;
            return;
        }

        this._editorSection.hidden = false;

        if (entry.group) {
            // Group editor: just show the type
            var p = document.createElement('p');
            p.className = 'text-sm text-[var(--color-text-muted)]';
            var isAllOf = !!entry.group.all_of;
            p.textContent = isAllOf
                ? 'AND group: all child rules must pass.'
                : 'OR group: at least one child rule must pass.';
            this._editorEl.appendChild(p);
            return;
        }

        var rule = entry.rule;
        if (!rule) return;

        // Op label
        var opLabel = document.createElement('div');
        opLabel.className = 'text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wider';
        opLabel.textContent = 'Operation: ' + rule.op;
        this._editorEl.appendChild(opLabel);

        var self = this;

        // Fields based on op
        switch (rule.op) {
            case 'required':
                // No editable fields
                break;

            case 'contains':
            case 'starts_with':
            case 'ends_with':
                this._editorEl.appendChild(this._buildValueOrClass(rule));
                this._editorEl.appendChild(this._buildNegate(rule));
                break;

            case 'equals':
                this._editorEl.appendChild(this._buildValueInput(rule));
                this._editorEl.appendChild(this._buildNegate(rule));
                break;

            case 'length':
                this._editorEl.appendChild(this._buildCmpSelect(rule));
                this._editorEl.appendChild(this._buildNInput(rule, false));
                break;

            case 'count':
                this._editorEl.appendChild(this._buildValueOrClass(rule));
                this._editorEl.appendChild(this._buildCmpSelect(rule));
                this._editorEl.appendChild(this._buildNInput(rule, false));
                break;

            case 'range':
                this._editorEl.appendChild(this._buildCmpSelect(rule));
                this._editorEl.appendChild(this._buildNInput(rule, true));
                break;

            case 'item_count':
                this._editorEl.appendChild(this._buildCmpSelect(rule));
                this._editorEl.appendChild(this._buildNInput(rule, false));
                break;

            case 'one_of':
                this._editorEl.appendChild(this._buildValuesTextarea(rule));
                this._editorEl.appendChild(this._buildNegate(rule));
                break;
        }

        // Message input (all ops)
        this._editorEl.appendChild(this._buildMessageInput(rule));
    }

    /**
     * Builds the value/class toggle editor for rules that support matching against
     * either a literal string value or a character class (uppercase, lowercase, digits,
     * symbols, spaces).
     *
     * Renders radio buttons to switch between "Literal value" (text input) and
     * "Character class" (select dropdown) modes. Switching modes clears the other
     * field's value in the rule object.
     *
     * @param {Object} rule - The rule object being edited. Mutated directly on input.
     * @returns {HTMLDivElement} The container element with the toggle and both input fields.
     * @private
     */
    _buildValueOrClass(rule) {
        var frag = document.createElement('div');
        frag.className = 'flex flex-col gap-2';

        var hasClass = !!rule.class;

        // Radio toggle
        var radioRow = document.createElement('div');
        radioRow.className = 'flex gap-4';

        var radioName = 'vw-vc-' + this._selectedId;

        var valRadio = document.createElement('label');
        valRadio.className = 'flex items-center gap-1.5 text-sm text-[var(--color-text)] cursor-pointer';
        var valInput = document.createElement('input');
        valInput.type = 'radio';
        valInput.name = radioName;
        valInput.checked = !hasClass;
        valRadio.appendChild(valInput);
        valRadio.appendChild(document.createTextNode(' Literal value'));
        radioRow.appendChild(valRadio);

        var clsRadio = document.createElement('label');
        clsRadio.className = 'flex items-center gap-1.5 text-sm text-[var(--color-text)] cursor-pointer';
        var clsInput = document.createElement('input');
        clsInput.type = 'radio';
        clsInput.name = radioName;
        clsInput.checked = hasClass;
        clsRadio.appendChild(clsInput);
        clsRadio.appendChild(document.createTextNode(' Character class'));
        radioRow.appendChild(clsRadio);

        frag.appendChild(radioRow);

        // Value input
        var valueRow = document.createElement('div');
        valueRow.className = 'flex flex-col gap-1';
        valueRow.hidden = hasClass;
        var valLabel = document.createElement('label');
        valLabel.textContent = 'Value';
        valLabel.className = 'text-xs font-medium text-[var(--color-text-muted)]';
        valueRow.appendChild(valLabel);
        var valField = document.createElement('input');
        valField.type = 'text';
        valField.className = 'rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-1.5 text-sm text-[var(--color-text)] outline-none';
        valField.value = rule.value || '';
        valField.addEventListener('input', function(e) {
            rule.value = e.target.value;
            rule.class = '';
            this._onEditorInput();
        }.bind(this));
        valueRow.appendChild(valField);
        frag.appendChild(valueRow);

        // Class select
        var classRow = document.createElement('div');
        classRow.className = 'flex flex-col gap-1';
        classRow.hidden = !hasClass;
        var clsLabel = document.createElement('label');
        clsLabel.textContent = 'Class';
        clsLabel.className = 'text-xs font-medium text-[var(--color-text-muted)]';
        classRow.appendChild(clsLabel);
        var clsSelect = document.createElement('select');
        clsSelect.className = 'rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-1.5 text-sm text-[var(--color-text)] outline-none';
        var classes = ['uppercase', 'lowercase', 'digits', 'symbols', 'spaces'];
        for (var i = 0; i < classes.length; i++) {
            var opt = document.createElement('option');
            opt.value = classes[i];
            opt.textContent = classes[i];
            opt.selected = rule.class === classes[i];
            clsSelect.appendChild(opt);
        }
        clsSelect.addEventListener('change', function(e) {
            rule.class = e.target.value;
            rule.value = '';
            this._onEditorInput();
        }.bind(this));
        classRow.appendChild(clsSelect);
        frag.appendChild(classRow);

        // Toggle handler
        valInput.addEventListener('change', function() {
            valueRow.hidden = false;
            classRow.hidden = true;
            rule.class = '';
            this._onEditorInput();
        }.bind(this));

        clsInput.addEventListener('change', function() {
            valueRow.hidden = true;
            classRow.hidden = false;
            rule.value = '';
            if (!rule.class) {
                rule.class = 'uppercase';
                clsSelect.value = 'uppercase';
            }
            this._onEditorInput();
        }.bind(this));

        return frag;
    }

    /**
     * Builds a simple text input field for the rule's `value` property.
     * Used by the `equals` operator editor.
     *
     * @param {Object} rule - The rule object being edited. Mutated directly on input.
     * @returns {HTMLDivElement} The labeled input field container.
     * @private
     */
    _buildValueInput(rule) {
        var field = document.createElement('div');
        field.className = 'flex flex-col gap-1';
        var label = document.createElement('label');
        label.textContent = 'Value';
        label.className = 'text-xs font-medium text-[var(--color-text-muted)]';
        field.appendChild(label);
        var input = document.createElement('input');
        input.type = 'text';
        input.className = 'rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-1.5 text-sm text-[var(--color-text)] outline-none';
        input.value = rule.value || '';
        input.addEventListener('input', function(e) {
            rule.value = e.target.value;
            this._onEditorInput();
        }.bind(this));
        field.appendChild(input);
        return field;
    }

    /**
     * Builds a select dropdown for the rule's comparison operator (`cmp`).
     * Options: equals, not equals, greater than, at least, less than, at most.
     *
     * @param {Object} rule - The rule object being edited. Mutated directly on change.
     * @returns {HTMLDivElement} The labeled select field container.
     * @private
     */
    _buildCmpSelect(rule) {
        var field = document.createElement('div');
        field.className = 'flex flex-col gap-1';
        var label = document.createElement('label');
        label.textContent = 'Comparison';
        label.className = 'text-xs font-medium text-[var(--color-text-muted)]';
        field.appendChild(label);

        var select = document.createElement('select');
        select.className = 'rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-1.5 text-sm text-[var(--color-text)] outline-none';
        var cmps = [
            { val: 'eq', label: 'equals (=)' },
            { val: 'neq', label: 'not equals (\u2260)' },
            { val: 'gt', label: 'greater than (>)' },
            { val: 'gte', label: 'at least (\u2265)' },
            { val: 'lt', label: 'less than (<)' },
            { val: 'lte', label: 'at most (\u2264)' },
        ];
        for (var i = 0; i < cmps.length; i++) {
            var opt = document.createElement('option');
            opt.value = cmps[i].val;
            opt.textContent = cmps[i].label;
            opt.selected = rule.cmp === cmps[i].val;
            select.appendChild(opt);
        }
        select.addEventListener('change', function(e) {
            rule.cmp = e.target.value;
            this._onEditorInput();
        }.bind(this));
        field.appendChild(select);
        return field;
    }

    /**
     * Builds a numeric input field for the rule's threshold value (`n`).
     *
     * For integer-only fields (length, count, item_count), sets `step="1"` and `min="0"`.
     * For decimal-capable fields (range), sets `step="any"` to allow floating-point input.
     *
     * @param {Object} rule - The rule object being edited. Mutated directly on input.
     * @param {boolean} allowDecimals - When `true`, allows floating-point values via
     *   `step="any"`. When `false`, restricts to non-negative integers.
     * @returns {HTMLDivElement} The labeled numeric input field container.
     * @private
     */
    _buildNInput(rule, allowDecimals) {
        var field = document.createElement('div');
        field.className = 'flex flex-col gap-1';
        var label = document.createElement('label');
        label.textContent = 'N';
        label.className = 'text-xs font-medium text-[var(--color-text-muted)]';
        field.appendChild(label);
        var input = document.createElement('input');
        input.type = 'number';
        input.className = 'rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-1.5 text-sm text-[var(--color-text)] outline-none';
        if (!allowDecimals) {
            input.step = '1';
            input.min = '0';
        } else {
            input.step = 'any';
        }
        input.value = (rule.n !== null && rule.n !== undefined) ? rule.n : '';
        input.addEventListener('input', function(e) {
            var v = e.target.value;
            if (v === '') {
                rule.n = null;
            } else {
                rule.n = parseFloat(v);
            }
            this._onEditorInput();
        }.bind(this));
        field.appendChild(input);
        return field;
    }

    /**
     * Builds a checkbox input for the rule's `negate` flag. When checked, the rule
     * result is inverted (pass becomes fail and vice versa).
     *
     * @param {Object} rule - The rule object being edited. Mutated directly on change.
     * @returns {HTMLDivElement} The labeled checkbox container.
     * @private
     */
    _buildNegate(rule) {
        var field = document.createElement('div');
        field.className = 'flex items-center gap-2';
        var label = document.createElement('label');
        label.className = 'flex items-center gap-1.5 text-sm text-[var(--color-text)] cursor-pointer';
        var checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.checked = !!rule.negate;
        checkbox.addEventListener('change', function(e) {
            rule.negate = e.target.checked;
            this._onEditorInput();
        }.bind(this));
        label.appendChild(checkbox);
        label.appendChild(document.createTextNode(' Negate'));
        field.appendChild(label);
        return field;
    }

    /**
     * Builds a textarea for the `one_of` rule's `values` array. Values are displayed
     * and edited as one value per line. Empty and whitespace-only lines are filtered out.
     *
     * @param {Object} rule - The rule object being edited. Its `values` array is
     *   mutated directly on input.
     * @returns {HTMLDivElement} The labeled textarea container.
     * @private
     */
    _buildValuesTextarea(rule) {
        var field = document.createElement('div');
        field.className = 'flex flex-col gap-1';
        var label = document.createElement('label');
        label.textContent = 'Values (one per line)';
        label.className = 'text-xs font-medium text-[var(--color-text-muted)]';
        field.appendChild(label);
        var textarea = document.createElement('textarea');
        textarea.className = 'rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-1.5 text-sm text-[var(--color-text)] outline-none font-mono';
        textarea.rows = 4;
        textarea.value = (rule.values || []).join('\n');
        textarea.addEventListener('input', function(e) {
            var lines = e.target.value.split('\n');
            rule.values = [];
            for (var i = 0; i < lines.length; i++) {
                var trimmed = lines[i].trim();
                if (trimmed !== '') rule.values.push(trimmed);
            }
            this._onEditorInput();
        }.bind(this));
        field.appendChild(textarea);
        return field;
    }

    /**
     * Builds a text input for the rule's optional custom validation error message.
     * If left empty, the rule evaluation functions generate a default message based
     * on the operator and parameters.
     *
     * @param {Object} rule - The rule object being edited. Mutated directly on input.
     * @returns {HTMLDivElement} The labeled input field container.
     * @private
     */
    _buildMessageInput(rule) {
        var field = document.createElement('div');
        field.className = 'flex flex-col gap-1';
        var label = document.createElement('label');
        label.textContent = 'Custom message (optional)';
        label.className = 'text-xs font-medium text-[var(--color-text-muted)]';
        field.appendChild(label);
        var input = document.createElement('input');
        input.type = 'text';
        input.className = 'rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-1.5 text-sm text-[var(--color-text)] outline-none';
        input.placeholder = 'Leave empty for auto-generated message';
        input.value = rule.message || '';
        input.addEventListener('input', function(e) {
            rule.message = e.target.value;
            this._onEditorInput();
        }.bind(this));
        field.appendChild(input);
        return field;
    }

    // ========================================
    // Test Panel
    // ========================================

    /**
     * Evaluates all rules against the current test value and renders the pass/fail
     * results in the test panel. Each result shows a check or cross icon with the
     * validation message. Group results include nested child results.
     *
     * Uses the standalone `evaluateEntries()` function for actual rule evaluation.
     *
     * @private
     */
    _renderTestResults() {
        if (!this._testResultsEl) return;
        this._testResultsEl.innerHTML = '';

        if (this._rules.length === 0) return;

        var results = evaluateEntries(this._testValue, this._rules);
        this._renderTestEntries(results, this._testResultsEl);
    }

    /**
     * Recursively renders test result entries into a container. Each result is displayed
     * as a colored row (green for pass, red for fail) with a check/cross icon and message.
     * Group results are followed by their nested child results in an indented container.
     *
     * @param {Array<Object>} results - Array of result objects from `evaluateEntries()`,
     *   each with `{ pass: boolean, message: string, children?: Array }`.
     * @param {HTMLElement} container - The DOM element to append result rows to.
     * @private
     */
    _renderTestEntries(results, container) {
        for (var i = 0; i < results.length; i++) {
            var r = results[i];
            var row = document.createElement('div');
            row.className = 'flex items-center gap-2 rounded px-2 py-1 text-sm ' + (r.pass ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]');

            var icon = document.createElement('span');
            icon.className = 'font-bold';
            icon.textContent = r.pass ? '\u2713' : '\u2717';
            row.appendChild(icon);

            var msg = document.createElement('span');
            msg.className = 'flex-1';
            msg.textContent = r.message;
            row.appendChild(msg);

            container.appendChild(row);

            if (r.children && r.children.length > 0) {
                var nested = document.createElement('div');
                nested.className = 'pl-4 flex flex-col gap-1';
                this._renderTestEntries(r.children, nested);
                container.appendChild(nested);
            }
        }
    }

    // ========================================
    // Sync
    // ========================================

    /**
     * Called when the rule tree structure changes (rule added or deleted).
     * Syncs the textarea, then re-renders the rule list, editor, and test results.
     *
     * @private
     */
    _onRulesChanged() {
        this._syncTextarea();
        this._renderRuleList();
        this._renderEditor();
        this._renderTestResults();
    }

    /**
     * Called when a rule's properties change in the editor (value, comparison, negate, etc.).
     * Syncs the textarea, then re-renders the rule list summaries and test results.
     * Does not re-render the editor itself (to preserve input focus).
     *
     * @private
     */
    _onEditorInput() {
        this._syncTextarea();
        this._renderRuleList();
        this._renderTestResults();
    }

    /**
     * Serializes the current rule tree to JSON (with internal `_id` fields stripped)
     * and writes it to the target textarea. Also dispatches a `validation-changed`
     * custom event with the JSON string in `detail.json`.
     *
     * @private
     */
    _syncTextarea() {
        var textareaId = this.getAttribute('textarea-id');
        if (!textareaId) return;
        var textarea = document.getElementById(textareaId);
        if (!textarea) return;

        var cleaned = this._stripIds(this._rules);
        var json = JSON.stringify({ rules: cleaned }, null, 2);
        textarea.value = json;

        this.dispatchEvent(new CustomEvent('validation-changed', {
            bubbles: true,
            detail: { json: json }
        }));
    }

    /**
     * Creates a deep copy of the rule entries array with all internal `_id` fields
     * removed and empty/falsy properties stripped. This produces clean JSON suitable
     * for storage and server-side consumption.
     *
     * Properties that are empty string, `false`, `null`, `undefined`, or empty arrays
     * are omitted from the output.
     *
     * @param {Array<Object>} entries - The array of rule entries to clean.
     * @returns {Array<Object>} A new array with cleaned entries (no `_id` fields,
     *   no empty properties).
     * @private
     */
    _stripIds(entries) {
        var result = [];
        for (var i = 0; i < entries.length; i++) {
            var entry = entries[i];
            var clean = {};
            if (entry.rule) {
                var r = {};
                var keys = ['op', 'value', 'values', 'class', 'cmp', 'n', 'negate', 'message'];
                for (var k = 0; k < keys.length; k++) {
                    var key = keys[k];
                    var val = entry.rule[key];
                    if (val === undefined || val === null || val === '' || val === false) continue;
                    if (key === 'values' && Array.isArray(val) && val.length === 0) continue;
                    if (key === 'n' && val === null) continue;
                    r[key] = val;
                }
                clean.rule = r;
            }
            if (entry.group) {
                var g = {};
                if (entry.group.all_of) g.all_of = this._stripIds(entry.group.all_of);
                if (entry.group.any_of) g.any_of = this._stripIds(entry.group.any_of);
                clean.group = g;
            }
            result.push(clean);
        }
        return result;
    }
}

// ========================================
// Client-Side Rule Evaluation (standalone functions)
// ========================================
// These functions are used by the test panel to evaluate rules against a test value
// in real time. They mirror the server-side validation logic so users can preview
// results before saving.

/**
 * Classifies whether a character belongs to a named character class.
 *
 * Supported classes:
 * - `"uppercase"` -- ASCII uppercase letters (A-Z)
 * - `"lowercase"` -- ASCII lowercase letters (a-z)
 * - `"digits"` -- ASCII digits (0-9)
 * - `"spaces"` -- whitespace characters (space, tab, newline, carriage return)
 * - `"symbols"` -- any character that is not a letter, digit, or whitespace
 *
 * @param {string} c - A single character to classify.
 * @param {string} cls - The character class name.
 * @returns {boolean} `true` if the character belongs to the specified class.
 */
function classifyChar(c, cls) {
    var code = c.charCodeAt(0);
    if (cls === 'uppercase') return code >= 65 && code <= 90;
    if (cls === 'lowercase') return code >= 97 && code <= 122;
    if (cls === 'digits') return code >= 48 && code <= 57;
    if (cls === 'spaces') return c === ' ' || c === '\t' || c === '\n' || c === '\r';
    if (cls === 'symbols') {
        return !(code >= 65 && code <= 90) && !(code >= 97 && code <= 122) &&
               !(code >= 48 && code <= 57) && c !== ' ' && c !== '\t' && c !== '\n' && c !== '\r';
    }
    return false;
}

/**
 * Converts a comparison operator code to a human-readable label for validation messages.
 *
 * @param {string} cmp - The comparison code (`"eq"`, `"neq"`, `"gt"`, `"gte"`, `"lt"`, `"lte"`).
 * @returns {string} The label (e.g., `"exactly"`, `"not"`, `"more than"`, `"at least"`,
 *   `"less than"`, `"at most"`).
 */
function cmpLabel(cmp) {
    var labels = { eq: 'exactly', neq: 'not', gt: 'more than', gte: 'at least', lt: 'less than', lte: 'at most' };
    return labels[cmp] || cmp;
}

/**
 * Performs a numeric comparison between two values using a comparison operator code.
 *
 * @param {number} actual - The actual value (left-hand side).
 * @param {number} expected - The expected threshold (right-hand side).
 * @param {string} cmp - The comparison operator (`"eq"`, `"neq"`, `"gt"`, `"gte"`,
 *   `"lt"`, `"lte"`).
 * @returns {boolean} The result of the comparison. Returns `false` for unknown operators.
 */
function compareCmp(actual, expected, cmp) {
    switch (cmp) {
        case 'eq': return actual === expected;
        case 'neq': return actual !== expected;
        case 'gt': return actual > expected;
        case 'gte': return actual >= expected;
        case 'lt': return actual < expected;
        case 'lte': return actual <= expected;
        default: return false;
    }
}

/**
 * Counts the number of items in a value string. First attempts to parse the value
 * as a JSON array (if it starts with `[`). Falls back to splitting by commas and
 * counting non-empty trimmed segments.
 *
 * @param {string} value - The value string to count items in.
 * @returns {number} The number of items. Returns `0` for empty/falsy input.
 */
function countItems(value) {
    if (!value) return 0;
    if (value.charAt(0) === '[') {
        try {
            var arr = JSON.parse(value);
            if (Array.isArray(arr)) return arr.length;
        } catch (e) {
            // fall through to comma split
        }
    }
    var parts = value.split(',');
    var count = 0;
    for (var i = 0; i < parts.length; i++) {
        if (parts[i].trim() !== '') count++;
    }
    return count;
}

/**
 * Formats a numeric value for display in validation messages. Returns `"?"` for
 * null/undefined, removes trailing `.0` for integers.
 *
 * @param {number|null|undefined} n - The number to format.
 * @returns {string} The formatted string.
 */
function formatN(n) {
    if (n === null || n === undefined) return '?';
    if (n === Math.floor(n)) return String(Math.floor(n));
    return String(n);
}

/**
 * Returns a display label for a rule's target value or character class, used in
 * default validation messages.
 *
 * @param {Object} rule - The rule object with optional `value` and `class` fields.
 * @returns {string} Display label (e.g., `'"hello"'`, `'uppercase characters'`, `'(unset)'`).
 */
function targetLabel(rule) {
    if (rule.value) return '"' + rule.value + '"';
    if (rule.class) return rule.class + ' characters';
    return '(unset)';
}

/**
 * Generates a default human-readable validation message for a rule based on its
 * operator and parameters. Used when no custom message is set on the rule.
 *
 * @param {Object} rule - The rule object with `op` and operator-specific fields.
 * @returns {string} A descriptive validation message (e.g., `"must be at least 3 characters"`,
 *   `"must contain uppercase characters"`).
 */
function defaultMessage(rule) {
    var neg = rule.negate;
    switch (rule.op) {
        case 'required':
            return 'is required';
        case 'contains':
            return (neg ? 'must not contain ' : 'must contain ') + targetLabel(rule);
        case 'starts_with':
            return (neg ? 'must not start with ' : 'must start with ') + targetLabel(rule);
        case 'ends_with':
            return (neg ? 'must not end with ' : 'must end with ') + targetLabel(rule);
        case 'equals':
            return (neg ? 'must not equal "' : 'must equal "') + (rule.value || '') + '"';
        case 'length':
            return 'must be ' + cmpLabel(rule.cmp) + ' ' + formatN(rule.n) + ' characters';
        case 'count':
            return 'must have ' + cmpLabel(rule.cmp) + ' ' + formatN(rule.n) + ' of ' + targetLabel(rule);
        case 'range':
            return 'value must be ' + cmpLabel(rule.cmp) + ' ' + formatN(rule.n);
        case 'item_count':
            return 'must have ' + cmpLabel(rule.cmp) + ' ' + formatN(rule.n) + ' items';
        case 'one_of':
            var vals = (rule.values || []).join(', ');
            return (neg ? 'must not be one of: ' : 'must be one of: ') + vals;
        default:
            return 'validation failed';
    }
}

/**
 * Evaluates a single validation rule against a string value.
 *
 * Handles all supported operators: required, contains, starts_with, ends_with,
 * equals, length (Unicode-aware via `Array.from`), count, range (numeric parsing),
 * item_count, and one_of. Applies the `negate` flag to invert results for all
 * operators except `required`.
 *
 * @param {string} value - The string value to validate.
 * @param {Object} rule - The rule object with `op` and operator-specific fields.
 * @returns {{ pass: boolean, message: string }} The evaluation result with a pass/fail
 *   boolean and the validation message (custom or auto-generated).
 */
function evaluateRule(value, rule) {
    var passed = false;

    switch (rule.op) {
        case 'required':
            passed = value.length > 0;
            break;

        case 'contains':
            if (rule.value) {
                passed = value.indexOf(rule.value) !== -1;
            } else if (rule.class) {
                for (var i = 0; i < value.length; i++) {
                    if (classifyChar(value.charAt(i), rule.class)) { passed = true; break; }
                }
            }
            break;

        case 'starts_with':
            if (rule.value) {
                passed = value.indexOf(rule.value) === 0;
            } else if (rule.class && value.length > 0) {
                passed = classifyChar(value.charAt(0), rule.class);
            }
            break;

        case 'ends_with':
            if (rule.value) {
                passed = value.length >= rule.value.length &&
                         value.substring(value.length - rule.value.length) === rule.value;
            } else if (rule.class && value.length > 0) {
                passed = classifyChar(value.charAt(value.length - 1), rule.class);
            }
            break;

        case 'equals':
            passed = value === (rule.value || '');
            break;

        case 'length':
            if (rule.cmp && rule.n !== null && rule.n !== undefined) {
                // Use Array.from for proper Unicode rune count
                var len = Array.from(value).length;
                passed = compareCmp(len, rule.n, rule.cmp);
            }
            break;

        case 'count':
            if (rule.cmp && rule.n !== null && rule.n !== undefined) {
                var cnt = 0;
                if (rule.value) {
                    var pos = 0;
                    while (true) {
                        var idx = value.indexOf(rule.value, pos);
                        if (idx === -1) break;
                        cnt++;
                        pos = idx + rule.value.length;
                    }
                } else if (rule.class) {
                    for (var ci = 0; ci < value.length; ci++) {
                        if (classifyChar(value.charAt(ci), rule.class)) cnt++;
                    }
                }
                passed = compareCmp(cnt, rule.n, rule.cmp);
            }
            break;

        case 'range':
            if (rule.cmp && rule.n !== null && rule.n !== undefined) {
                var f = parseFloat(value);
                if (isNaN(f)) {
                    return { pass: false, message: rule.message || 'must be a number' };
                }
                passed = compareCmp(f, rule.n, rule.cmp);
            }
            break;

        case 'item_count':
            if (rule.cmp && rule.n !== null && rule.n !== undefined) {
                passed = compareCmp(countItems(value), rule.n, rule.cmp);
            }
            break;

        case 'one_of':
            var vals = rule.values || [];
            for (var oi = 0; oi < vals.length; oi++) {
                if (value === vals[oi]) { passed = true; break; }
            }
            break;

        default:
            passed = true;
    }

    // Apply negate (except required which doesn't support negate)
    if (rule.negate && rule.op !== 'required') {
        passed = !passed;
    }

    var msg = rule.message || defaultMessage(rule);
    return { pass: passed, message: msg };
}

/**
 * Evaluates an array of rule entries (rules and groups) against a string value.
 *
 * Individual rules are evaluated with `evaluateRule()`. Groups evaluate their children
 * recursively: `all_of` groups pass only if every child passes (AND logic), `any_of`
 * groups pass if at least one child passes (OR logic). Empty `any_of` groups vacuously
 * pass.
 *
 * @param {string} value - The string value to validate against all entries.
 * @param {Array<Object>} entries - Array of rule entries, each either
 *   `{ rule: {...} }` or `{ group: { all_of: [...] | any_of: [...] } }`.
 * @returns {Array<{ pass: boolean, message: string, children?: Array }>} Array of
 *   evaluation results. Group results include a `children` array with nested results.
 */
function evaluateEntries(value, entries) {
    var results = [];
    for (var i = 0; i < entries.length; i++) {
        var entry = entries[i];
        if (entry.rule) {
            results.push(evaluateRule(value, entry.rule));
        } else if (entry.group) {
            var g = entry.group;
            var isAllOf = !!g.all_of;
            var children = isAllOf ? g.all_of : g.any_of;
            var childResults = evaluateEntries(value, children || []);

            var groupPass;
            if (isAllOf) {
                groupPass = true;
                for (var ai = 0; ai < childResults.length; ai++) {
                    if (!childResults[ai].pass) { groupPass = false; break; }
                }
            } else {
                groupPass = false;
                for (var oi = 0; oi < childResults.length; oi++) {
                    if (childResults[oi].pass) { groupPass = true; break; }
                }
                // Empty any_of group: vacuously pass
                if (childResults.length === 0) groupPass = true;
            }

            results.push({
                pass: groupPass,
                message: (isAllOf ? 'all_of' : 'any_of') + ' (' + (groupPass ? 'pass' : 'fail') + ')',
                children: childResults
            });
        }
    }
    return results;
}

if (!customElements.get('mcms-validation-wizard')) {
    customElements.define('mcms-validation-wizard', McmsValidationWizard);
}
