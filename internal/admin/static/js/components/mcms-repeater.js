/**
 * @module mcms-repeater
 * @description Repeater field component for editing arrays of objects in settings forms.
 *
 * Reads initial data from the `data-value` attribute (JSON array) and renders
 * editable rows with one input per field. Users can add and remove rows.
 * On form submit, the component serializes all rows to a hidden input as JSON.
 *
 * This is a Light DOM web component -- all rendered elements are direct children
 * of the host element, accessible to global CSS and HTMX processing.
 *
 * Supported field types:
 *   - "text" (default): standard text input
 *   - "password": password input
 *   - "tables": renders a "Configure Tables" button that opens a dialog with
 *     checkboxes for selecting database tables. The dialog ID is referenced via
 *     the field's "dialog" property, which should match a server-rendered
 *     mcms-dialog element on the page.
 *
 * @example
 * <mcms-repeater
 *   name="deploy_environments"
 *   data-value='[{"name":"production","url":"https://cms.example.com","api_key":"mcms_KEY","tables":["content_data"]}]'
 *   data-fields='[{"key":"name","label":"Name","type":"text"},{"key":"url","label":"URL","type":"text"},{"key":"api_key","label":"API Key","type":"password"},{"key":"tables","label":"Tables","type":"tables"}]'>
 * </mcms-repeater>
 *
 * @attr {string} name - Form field name. The hidden input uses this name.
 * @attr {string} data-value - JSON array of objects (initial data).
 * @attr {string} data-fields - JSON array of field definitions: {key, label, type, placeholder}.
 */
class McmsRepeater extends HTMLElement {
  connectedCallback() {
    this._fields = JSON.parse(this.getAttribute('data-fields') || '[]');
    this._rows = JSON.parse(this.getAttribute('data-value') || '[]');
    this._name = this.getAttribute('name') || 'repeater';
    this._render();
  }

  _render() {
    this.innerHTML = '';

    // Hidden input that holds the serialized JSON for form submission.
    var hidden = document.createElement('input');
    hidden.type = 'hidden';
    hidden.name = this._name;
    this._hidden = hidden;
    this.appendChild(hidden);
    this._sync();

    // Rows container.
    var container = document.createElement('div');
    container.className = 'space-y-3';
    this._container = container;
    this.appendChild(container);

    for (var i = 0; i < this._rows.length; i++) {
      this._addRowEl(this._rows[i], i);
    }

    // Add button.
    var addBtn = document.createElement('button');
    addBtn.type = 'button';
    addBtn.className = 'mt-3 inline-flex items-center gap-x-1.5 rounded-md bg-white/5 px-3 py-1.5 text-sm font-medium text-white ring-1 ring-white/10 ring-inset hover:bg-white/10';
    addBtn.textContent = '+ Add Environment';
    addBtn.addEventListener('click', this._addRow.bind(this));
    this.appendChild(addBtn);
  }

  _addRowEl(data, index) {
    var row = document.createElement('div');
    row.className = 'flex items-start gap-x-3 rounded-lg bg-white/[0.03] p-3 ring-1 ring-white/10';
    row.dataset.index = index;

    var fieldsWrap = document.createElement('div');
    fieldsWrap.className = 'grid flex-1 grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-3';

    var self = this;
    for (var f = 0; f < this._fields.length; f++) {
      var field = this._fields[f];
      var wrap = document.createElement('div');

      var label = document.createElement('label');
      label.className = 'block text-xs font-medium text-gray-400 mb-1';
      label.textContent = field.label;
      wrap.appendChild(label);

      if (field.type === 'tables') {
        this._renderTablesField(wrap, data, index, field);
      } else {
        var input = document.createElement('input');
        input.type = field.type || 'text';
        input.placeholder = field.placeholder || '';
        input.value = data[field.key] || '';
        input.className = 'block w-full rounded-md bg-white/5 px-3 py-1.5 text-sm text-white outline-1 -outline-offset-1 outline-white/10 placeholder:text-gray-500 focus:outline-2 focus:-outline-offset-2 focus:outline-[var(--color-primary)]';
        input.dataset.key = field.key;
        input.dataset.rowIndex = index;
        input.addEventListener('input', function(e) {
          var ri = parseInt(e.target.dataset.rowIndex, 10);
          self._rows[ri][e.target.dataset.key] = e.target.value;
          self._sync();
        });
        wrap.appendChild(input);
      }

      fieldsWrap.appendChild(wrap);
    }

    row.appendChild(fieldsWrap);

    // Remove button.
    var removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'mt-5 shrink-0 rounded-md p-1.5 text-gray-500 hover:text-red-400 hover:bg-white/5';
    removeBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-4"><path fill-rule="evenodd" d="M8.75 1A2.75 2.75 0 0 0 6 3.75v.443c-.795.077-1.584.176-2.365.298a.75.75 0 1 0 .23 1.482l.149-.022.841 10.518A2.75 2.75 0 0 0 7.596 19h4.807a2.75 2.75 0 0 0 2.742-2.53l.841-10.52.149.023a.75.75 0 0 0 .23-1.482A41.03 41.03 0 0 0 14 4.193V3.75A2.75 2.75 0 0 0 11.25 1h-2.5ZM10 4c.84 0 1.673.025 2.5.075V3.75c0-.69-.56-1.25-1.25-1.25h-2.5c-.69 0-1.25.56-1.25 1.25v.325C8.327 4.025 9.16 4 10 4ZM8.58 7.72a.75.75 0 0 0-1.5.06l.3 7.5a.75.75 0 1 0 1.5-.06l-.3-7.5Zm4.34.06a.75.75 0 1 0-1.5-.06l-.3 7.5a.75.75 0 1 0 1.5.06l.3-7.5Z" clip-rule="evenodd"/></svg>';
    removeBtn.dataset.rowIndex = index;
    removeBtn.addEventListener('click', function(e) {
      var btn = e.target.closest('button');
      var ri = parseInt(btn.dataset.rowIndex, 10);
      self._rows.splice(ri, 1);
      self._render();
    });
    row.appendChild(removeBtn);

    this._container.appendChild(row);
  }

  _renderTablesField(wrap, data, index, field) {
    var self = this;
    var tables = data[field.key] || [];
    var count = Array.isArray(tables) ? tables.length : 0;

    // Badge showing count.
    var badge = document.createElement('span');
    badge.className = 'text-xs text-gray-500';
    badge.textContent = count > 0 ? count + ' table' + (count !== 1 ? 's' : '') + ' selected' : 'Default (content only)';

    // Button to open dialog.
    var btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'inline-flex items-center gap-x-1.5 rounded-md bg-white/5 px-3 py-1.5 text-sm font-medium text-white ring-1 ring-white/10 ring-inset hover:bg-white/10';
    btn.textContent = 'Configure';
    btn.addEventListener('click', function() {
      self._openTablesDialog(index, field.key, tables);
    });

    var row = document.createElement('div');
    row.className = 'flex items-center gap-x-3';
    row.appendChild(btn);
    row.appendChild(badge);
    wrap.appendChild(row);
  }

  _openTablesDialog(rowIndex, fieldKey, currentTables) {
    var self = this;
    // Find the server-rendered dialog element.
    var dialogId = 'deploy-tables-settings-dialog';
    var dialog = document.getElementById(dialogId);
    if (!dialog) return;

    // Update checkboxes to match current row's tables.
    var checkboxes = dialog.querySelectorAll('input[type="checkbox"][name="tables"]');
    var selected = new Set(Array.isArray(currentTables) ? currentTables : []);
    for (var i = 0; i < checkboxes.length; i++) {
      checkboxes[i].checked = selected.has(checkboxes[i].value);
    }

    // Wire up the done button to save selections.
    var doneBtn = dialog.querySelector('[data-action="save-tables"]');
    if (doneBtn) {
      // Clone to remove old listeners.
      var newBtn = doneBtn.cloneNode(true);
      doneBtn.parentNode.replaceChild(newBtn, doneBtn);
      newBtn.addEventListener('click', function() {
        var checked = dialog.querySelectorAll('input[type="checkbox"][name="tables"]:checked');
        var tableNames = [];
        for (var j = 0; j < checked.length; j++) {
          tableNames.push(checked[j].value);
        }
        self._rows[rowIndex][fieldKey] = tableNames;
        self._sync();
        self._render();
        dialog.close();
      });
    }

    dialog.open();
  }

  _addRow() {
    var obj = {};
    for (var f = 0; f < this._fields.length; f++) {
      var field = this._fields[f];
      if (field.type === 'tables') {
        obj[field.key] = [];
      } else {
        obj[field.key] = '';
      }
    }
    this._rows.push(obj);
    this._render();
  }

  _sync() {
    // Filter out completely empty rows before serializing.
    var nonEmpty = [];
    for (var i = 0; i < this._rows.length; i++) {
      var row = this._rows[i];
      var hasValue = false;
      for (var k in row) {
        var v = row[k];
        if (Array.isArray(v)) {
          if (v.length > 0) { hasValue = true; break; }
        } else if (v !== '') {
          hasValue = true; break;
        }
      }
      if (hasValue) nonEmpty.push(row);
    }
    this._hidden.value = JSON.stringify(nonEmpty);
  }
}

customElements.define('mcms-repeater', McmsRepeater);
