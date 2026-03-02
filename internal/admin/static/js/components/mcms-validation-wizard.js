// <mcms-validation-wizard> -- Visual builder for composable validation rules (Light DOM)
// Attributes: value (initial JSON), textarea-id (sync target)
// Dispatches: validation-changed custom event on every rule change
class McmsValidationWizard extends HTMLElement {
    constructor() {
        super();
        this._built = false;
        this._rules = [];
        this._nextId = 1;
        this._selectedId = null;
        this._insertTargetId = null;
        this._testValue = '';
        this._testTimer = null;
        this._container = null;
        this._ruleListEl = null;
        this._editorEl = null;
        this._testResultsEl = null;
        this._warningEl = null;
    }

    connectedCallback() {
        if (this._built) return;
        this._built = true;
        this._parseInitialValue();
        this._build();
        this._renderAll();
    }

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

    _getEntry(id) {
        var loc = this._findById(this._rules, id);
        return loc ? loc.entries[loc.index] : null;
    }

    // ========================================
    // Build DOM
    // ========================================
    _build() {
        var container = document.createElement('div');
        container.className = 'vw-container';

        // Warning banner
        var warning = document.createElement('div');
        warning.className = 'vw-warning';
        warning.textContent = 'Could not parse existing validation config. Starting with empty rules.';
        warning.hidden = !this._showWarning;
        container.appendChild(warning);
        this._warningEl = warning;

        // Rule list section
        var listSection = document.createElement('div');
        listSection.className = 'vw-section';

        var listHeader = document.createElement('div');
        listHeader.className = 'vw-section-header';

        var listTitle = document.createElement('span');
        listTitle.className = 'vw-section-title';
        listTitle.textContent = 'Rule Builder';
        listHeader.appendChild(listTitle);

        var addBtn = document.createElement('div');
        addBtn.className = 'vw-add-wrapper';

        var addButton = document.createElement('button');
        addButton.type = 'button';
        addButton.className = 'btn btn-sm btn-primary';
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
        ruleList.className = 'vw-rule-list';
        listSection.appendChild(ruleList);
        this._ruleListEl = ruleList;

        container.appendChild(listSection);

        // Rule editor section
        var editorSection = document.createElement('div');
        editorSection.className = 'vw-section vw-editor-section';
        editorSection.hidden = true;

        var editorHeader = document.createElement('div');
        editorHeader.className = 'vw-section-header';
        var editorTitle = document.createElement('span');
        editorTitle.className = 'vw-section-title';
        editorTitle.textContent = 'Rule Editor';
        editorHeader.appendChild(editorTitle);
        editorSection.appendChild(editorHeader);

        var editorContent = document.createElement('div');
        editorContent.className = 'vw-editor-content';
        editorSection.appendChild(editorContent);
        this._editorEl = editorContent;
        this._editorSection = editorSection;

        container.appendChild(editorSection);

        // Test panel section
        var testSection = document.createElement('div');
        testSection.className = 'vw-section';

        var testHeader = document.createElement('div');
        testHeader.className = 'vw-section-header';
        var testTitle = document.createElement('span');
        testTitle.className = 'vw-section-title';
        testTitle.textContent = 'Test';
        testHeader.appendChild(testTitle);
        testSection.appendChild(testHeader);

        var testInput = document.createElement('input');
        testInput.type = 'text';
        testInput.className = 'vw-test-input';
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
        testResults.className = 'vw-test-results';
        testSection.appendChild(testResults);
        this._testResultsEl = testResults;

        container.appendChild(testSection);
        this._container = container;
        this.appendChild(container);

        // Close dropdowns when clicking outside
        document.addEventListener('click', function(e) {
            if (!e.target.closest('.vw-add-wrapper')) {
                var dropdowns = this.querySelectorAll('.vw-dropdown.open');
                for (var i = 0; i < dropdowns.length; i++) {
                    dropdowns[i].classList.remove('open');
                }
            }
        }.bind(this));
    }

    _buildDropdown() {
        var dropdown = document.createElement('div');
        dropdown.className = 'vw-dropdown';

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
                sep.className = 'vw-dropdown-sep';
                dropdown.appendChild(sep);
                continue;
            }
            var btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'vw-dropdown-item';
            btn.textContent = item.label;
            btn.dataset.op = item.op;
            btn.addEventListener('click', function(e) {
                e.stopPropagation();
                var op = e.target.dataset.op;
                this._addRule(op);
                // Close all dropdowns
                var dropdowns = this.querySelectorAll('.vw-dropdown.open');
                for (var j = 0; j < dropdowns.length; j++) {
                    dropdowns[j].classList.remove('open');
                }
            }.bind(this));
            dropdown.appendChild(btn);
        }

        return dropdown;
    }

    _toggleDropdown(wrapper) {
        var dd = wrapper.querySelector('.vw-dropdown');
        if (!dd) return;
        // Close all other dropdowns first
        var allDd = this.querySelectorAll('.vw-dropdown.open');
        for (var i = 0; i < allDd.length; i++) {
            if (allDd[i] !== dd) allDd[i].classList.remove('open');
        }
        dd.classList.toggle('open');
    }

    // ========================================
    // Add / Delete Rules
    // ========================================
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
    _renderAll() {
        this._renderRuleList();
        this._renderEditor();
        this._renderTestResults();
    }

    _renderRuleList() {
        if (!this._ruleListEl) return;
        this._ruleListEl.innerHTML = '';
        if (this._rules.length === 0) {
            var empty = document.createElement('div');
            empty.className = 'vw-empty';
            empty.textContent = 'No rules configured. Click "+ Add" to create a rule.';
            this._ruleListEl.appendChild(empty);
            return;
        }
        this._renderEntries(this._rules, this._ruleListEl, 0);
    }

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

    _renderRuleItem(entry, container) {
        var row = document.createElement('div');
        row.className = 'vw-rule-item';
        if (entry._id === this._selectedId) row.classList.add('selected');

        var label = document.createElement('span');
        label.className = 'vw-rule-label';
        label.textContent = this._ruleSummary(entry.rule);
        label.addEventListener('click', function() {
            this._selectedId = entry._id;
            this._renderAll();
        }.bind(this));
        row.appendChild(label);

        var del = document.createElement('button');
        del.type = 'button';
        del.className = 'vw-rule-delete';
        del.textContent = '\u00d7';
        del.title = 'Remove rule';
        del.addEventListener('click', function(e) {
            e.stopPropagation();
            this._deleteRule(entry._id);
        }.bind(this));
        row.appendChild(del);

        container.appendChild(row);
    }

    _renderGroupItem(entry, container, depth) {
        var g = entry.group;
        var isAllOf = !!g.all_of;
        var children = isAllOf ? g.all_of : g.any_of;

        var group = document.createElement('div');
        group.className = 'vw-group';
        if (entry._id === this._selectedId) group.classList.add('selected');

        var header = document.createElement('div');
        header.className = 'vw-group-header';

        var label = document.createElement('span');
        label.className = 'vw-group-label';
        label.textContent = isAllOf ? 'All must pass' : 'Any must pass';
        label.addEventListener('click', function() {
            this._selectedId = entry._id;
            this._renderAll();
        }.bind(this));
        header.appendChild(label);

        var actions = document.createElement('div');
        actions.className = 'vw-group-actions';

        var addWrapper = document.createElement('div');
        addWrapper.className = 'vw-add-wrapper';

        var addBtn = document.createElement('button');
        addBtn.type = 'button';
        addBtn.className = 'btn btn-sm btn-ghost';
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
        del.className = 'vw-rule-delete';
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
        body.className = 'vw-group-body';
        if (children.length === 0) {
            var empty = document.createElement('div');
            empty.className = 'vw-empty vw-empty-sm';
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

    _targetLabel(rule) {
        if (rule.value) return '"' + rule.value + '"';
        if (rule.class) return rule.class;
        return '(unset)';
    }

    _cmpSymbol(cmp) {
        var symbols = { eq: '=', neq: '\u2260', gt: '>', gte: '\u2265', lt: '<', lte: '\u2264' };
        return symbols[cmp] || cmp || '?';
    }

    _formatN(n) {
        if (n === null || n === undefined) return '?';
        if (n === Math.floor(n)) return String(Math.floor(n));
        return String(n);
    }

    // ========================================
    // Rule Editor
    // ========================================
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
            p.className = 'vw-editor-info';
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
        opLabel.className = 'vw-editor-op';
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

    _buildValueOrClass(rule) {
        var frag = document.createElement('div');
        frag.className = 'vw-editor-field';

        var hasClass = !!rule.class;

        // Radio toggle
        var radioRow = document.createElement('div');
        radioRow.className = 'vw-radio-row';

        var radioName = 'vw-vc-' + this._selectedId;

        var valRadio = document.createElement('label');
        valRadio.className = 'vw-radio-label';
        var valInput = document.createElement('input');
        valInput.type = 'radio';
        valInput.name = radioName;
        valInput.checked = !hasClass;
        valRadio.appendChild(valInput);
        valRadio.appendChild(document.createTextNode(' Literal value'));
        radioRow.appendChild(valRadio);

        var clsRadio = document.createElement('label');
        clsRadio.className = 'vw-radio-label';
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
        valueRow.className = 'vw-editor-field';
        valueRow.hidden = hasClass;
        var valLabel = document.createElement('label');
        valLabel.textContent = 'Value';
        valLabel.className = 'vw-label';
        valueRow.appendChild(valLabel);
        var valField = document.createElement('input');
        valField.type = 'text';
        valField.className = 'vw-input';
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
        classRow.className = 'vw-editor-field';
        classRow.hidden = !hasClass;
        var clsLabel = document.createElement('label');
        clsLabel.textContent = 'Class';
        clsLabel.className = 'vw-label';
        classRow.appendChild(clsLabel);
        var clsSelect = document.createElement('select');
        clsSelect.className = 'vw-select';
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

    _buildValueInput(rule) {
        var field = document.createElement('div');
        field.className = 'vw-editor-field';
        var label = document.createElement('label');
        label.textContent = 'Value';
        label.className = 'vw-label';
        field.appendChild(label);
        var input = document.createElement('input');
        input.type = 'text';
        input.className = 'vw-input';
        input.value = rule.value || '';
        input.addEventListener('input', function(e) {
            rule.value = e.target.value;
            this._onEditorInput();
        }.bind(this));
        field.appendChild(input);
        return field;
    }

    _buildCmpSelect(rule) {
        var field = document.createElement('div');
        field.className = 'vw-editor-field';
        var label = document.createElement('label');
        label.textContent = 'Comparison';
        label.className = 'vw-label';
        field.appendChild(label);

        var select = document.createElement('select');
        select.className = 'vw-select';
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

    _buildNInput(rule, allowDecimals) {
        var field = document.createElement('div');
        field.className = 'vw-editor-field';
        var label = document.createElement('label');
        label.textContent = 'N';
        label.className = 'vw-label';
        field.appendChild(label);
        var input = document.createElement('input');
        input.type = 'number';
        input.className = 'vw-input';
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

    _buildNegate(rule) {
        var field = document.createElement('div');
        field.className = 'vw-editor-field vw-checkbox-field';
        var label = document.createElement('label');
        label.className = 'vw-checkbox-label';
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

    _buildValuesTextarea(rule) {
        var field = document.createElement('div');
        field.className = 'vw-editor-field';
        var label = document.createElement('label');
        label.textContent = 'Values (one per line)';
        label.className = 'vw-label';
        field.appendChild(label);
        var textarea = document.createElement('textarea');
        textarea.className = 'vw-textarea';
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

    _buildMessageInput(rule) {
        var field = document.createElement('div');
        field.className = 'vw-editor-field';
        var label = document.createElement('label');
        label.textContent = 'Custom message (optional)';
        label.className = 'vw-label';
        field.appendChild(label);
        var input = document.createElement('input');
        input.type = 'text';
        input.className = 'vw-input';
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
    _renderTestResults() {
        if (!this._testResultsEl) return;
        this._testResultsEl.innerHTML = '';

        if (this._rules.length === 0) return;

        var results = evaluateEntries(this._testValue, this._rules);
        this._renderTestEntries(results, this._testResultsEl);
    }

    _renderTestEntries(results, container) {
        for (var i = 0; i < results.length; i++) {
            var r = results[i];
            var row = document.createElement('div');
            row.className = 'vw-test-row ' + (r.pass ? 'pass' : 'fail');

            var icon = document.createElement('span');
            icon.className = 'vw-test-icon';
            icon.textContent = r.pass ? '\u2713' : '\u2717';
            row.appendChild(icon);

            var msg = document.createElement('span');
            msg.className = 'vw-test-msg';
            msg.textContent = r.message;
            row.appendChild(msg);

            container.appendChild(row);

            if (r.children && r.children.length > 0) {
                var nested = document.createElement('div');
                nested.className = 'vw-test-nested';
                this._renderTestEntries(r.children, nested);
                container.appendChild(nested);
            }
        }
    }

    // ========================================
    // Sync
    // ========================================
    _onRulesChanged() {
        this._syncTextarea();
        this._renderRuleList();
        this._renderEditor();
        this._renderTestResults();
    }

    _onEditorInput() {
        this._syncTextarea();
        this._renderRuleList();
        this._renderTestResults();
    }

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

function cmpLabel(cmp) {
    var labels = { eq: 'exactly', neq: 'not', gt: 'more than', gte: 'at least', lt: 'less than', lte: 'at most' };
    return labels[cmp] || cmp;
}

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

function formatN(n) {
    if (n === null || n === undefined) return '?';
    if (n === Math.floor(n)) return String(Math.floor(n));
    return String(n);
}

function targetLabel(rule) {
    if (rule.value) return '"' + rule.value + '"';
    if (rule.class) return rule.class + ' characters';
    return '(unset)';
}

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

customElements.define('mcms-validation-wizard', McmsValidationWizard);
