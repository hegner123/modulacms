// ModulaCMS Admin — HTMX config and glue code

// ========================================
// Theme: persist light/dark preference
// ========================================
(function initTheme() {
    const saved = localStorage.getItem('mcms-theme');
    if (saved === 'light') {
        document.documentElement.classList.add('light');
    } else if (saved === 'dark') {
        document.documentElement.classList.remove('light');
    }
    // If no preference saved, default is dark (no .light class)
})();

function toggleTheme() {
    const isLight = document.documentElement.classList.toggle('light');
    localStorage.setItem('mcms-theme', isLight ? 'light' : 'dark');
    // Update toggle button icon
    const btn = document.querySelector('.theme-toggle');
    if (btn) {
        btn.textContent = isLight ? '\u263E' : '\u2600'; // ☾ / ☀
    }
}

// ========================================
// Richtext: fetch global toolbar config
// ========================================
(function() {
    fetch('/admin/api/config/richtext-toolbar', { credentials: 'same-origin' })
        .then(function(res) { return res.ok ? res.json() : null; })
        .then(function(toolbar) { if (toolbar) window.__mcmsRichtextToolbar = toolbar; })
        .catch(function() {});
})();

// ========================================
// CSRF: inject token into all HTMX requests
// ========================================
document.body.addEventListener('htmx:configRequest', (e) => {
    const meta = document.querySelector('meta[name="csrf-token"]');
    if (meta && meta.content) {
        e.detail.headers['X-CSRF-Token'] = meta.content;
    }
});

// ========================================
// Loading Bar: top-of-page progress indicator
// ========================================
(function initLoadingBar() {
    const bar = document.createElement('div');
    bar.className = 'loading-bar';
    bar.id = 'loading-bar';
    document.body.appendChild(bar);

    document.body.addEventListener('htmx:beforeRequest', () => {
        bar.classList.remove('done');
        bar.classList.add('active');
    });

    document.body.addEventListener('htmx:afterRequest', () => {
        bar.classList.remove('active');
        bar.classList.add('done');
        setTimeout(() => {
            bar.classList.remove('done');
            // Reset width for next request
            bar.style.width = '0';
            requestAnimationFrame(() => { bar.style.width = ''; });
        }, 300);
    });
})();

// ========================================
// Error handling: HTMX response errors
// ========================================
document.body.addEventListener('htmx:responseError', (e) => {
    const toast = document.querySelector('mcms-toast');
    if (!toast) return;

    const status = e.detail.xhr.status;
    if (status === 403) {
        toast.show('Permission denied', 'error');
    } else if (status === 422) {
        // Validation errors are handled by the response body swap
    } else if (status === 429) {
        toast.show('Too many requests. Please wait.', 'error');
    } else if (status >= 500) {
        toast.show('Server error. Please try again.', 'error');
    } else if (status === 0) {
        toast.show('Network error. Check your connection.', 'error');
    }
});

// Handle toast events from HTMX response headers (HX-Trigger)
document.body.addEventListener('showToast', (e) => {
    const toast = document.querySelector('mcms-toast');
    if (toast && e.detail) {
        toast.show(e.detail.message || 'Action completed', e.detail.type || 'info');
    }
});

// Allow 422 responses to be swapped (validation errors)
document.body.addEventListener('htmx:beforeSwap', (e) => {
    if (e.detail.xhr.status === 422) {
        e.detail.shouldSwap = true;
        e.detail.isError = false;
    }
});

// Handle network errors (timeout, connection refused)
document.body.addEventListener('htmx:sendError', () => {
    const toast = document.querySelector('mcms-toast');
    if (toast) {
        toast.show('Unable to reach server. Please try again.', 'error');
    }
});

// ========================================
// Keyboard Shortcuts
// ========================================
const shortcuts = [
    { key: '/', desc: 'Focus search', action: focusSearch },
    { key: 'n', desc: 'New item', action: clickNewButton },
    { key: 'g h', desc: 'Go to dashboard', action: () => navigateTo('/admin/') },
    { key: 'g c', desc: 'Go to content', action: () => navigateTo('/admin/content') },
    { key: 'g m', desc: 'Go to media', action: () => navigateTo('/admin/media') },
    { key: 'g u', desc: 'Go to users', action: () => navigateTo('/admin/users') },
    { key: 'g s', desc: 'Go to settings', action: () => navigateTo('/admin/settings') },
    { key: '?', desc: 'Show shortcuts', action: toggleShortcutsHelp },
    { key: 'Escape', desc: 'Close dialog/overlay', action: closeTopOverlay },
    { key: '~ ~', desc: 'Focus block editor', action: focusBlockEditor },
];

let pendingPrefix = '';
let prefixTimer = null;

document.addEventListener('keydown', (e) => {
    // Skip if user is typing in an input
    const tag = e.target.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || e.target.isContentEditable) {
        if (e.key === 'Escape') {
            e.target.blur();
        }
        return;
    }

    // Skip if modifier keys are held (except shift for ?)
    if (e.ctrlKey || e.metaKey || e.altKey) return;

    const key = e.key;

    // Handle two-key sequences (g h, g c, etc.)
    if (pendingPrefix) {
        const combo = pendingPrefix + ' ' + key;
        clearTimeout(prefixTimer);
        pendingPrefix = '';
        const match = shortcuts.find(s => s.key === combo);
        if (match) {
            e.preventDefault();
            match.action();
        }
        return;
    }

    // Start a prefix sequence
    if (key === 'g' || key === '~') {
        pendingPrefix = key;
        prefixTimer = setTimeout(() => { pendingPrefix = ''; }, 500);
        return;
    }

    // Single-key shortcuts
    const match = shortcuts.find(s => s.key === key);
    if (match) {
        e.preventDefault();
        match.action();
    }
});

function focusSearch() {
    const input = document.querySelector('mcms-search input, .search-input, input[type="search"]');
    if (input) {
        input.focus();
        input.select();
    }
}

function clickNewButton() {
    const btn = document.querySelector('[data-shortcut="new"], .page-header .btn-primary');
    if (btn) btn.click();
}

function navigateTo(url) {
    // Use HTMX SPA navigation instead of full page reload
    if (typeof htmx !== 'undefined') {
        htmx.ajax('GET', url, {target: '#main-content', swap: 'innerHTML show:window:top'});
        history.pushState({}, '', url);
        return;
    }
    window.location.href = url;
}

function toggleShortcutsHelp() {
    let overlay = document.getElementById('shortcuts-overlay');
    if (overlay) {
        overlay.remove();
        return;
    }

    overlay = document.createElement('div');
    overlay.id = 'shortcuts-overlay';
    overlay.className = 'shortcuts-overlay';
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) overlay.remove();
    });

    const panel = document.createElement('div');
    panel.className = 'shortcuts-panel';
    panel.innerHTML = '<h2>Keyboard Shortcuts</h2><div class="shortcuts-list">' +
        shortcuts.map(s => {
            const keys = s.key.split(' ').map(k =>
                '<span class="shortcut-key">' + escapeHtml(k) + '</span>'
            ).join(' ');
            return '<div class="shortcut-row"><span class="shortcut-keys">' + keys + '</span><span class="shortcut-desc">' + escapeHtml(s.desc) + '</span></div>';
        }).join('') +
        '</div>';

    overlay.appendChild(panel);
    document.body.appendChild(overlay);
}

function focusBlockEditor() {
    var editor = document.querySelector('block-editor');
    if (!editor) return;
    var container = editor.querySelector('.editor-container');
    if (container) container.focus();
}

function closeTopOverlay() {
    // Close shortcuts help
    const shortcuts = document.getElementById('shortcuts-overlay');
    if (shortcuts) { shortcuts.remove(); return; }

    // Close any open dialog
    const dialog = document.querySelector('mcms-dialog[open]');
    if (dialog) { dialog.close(); return; }

    // Close any open media picker
    const picker = document.querySelector('mcms-media-picker[open]');
    if (picker) { picker.close(); }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// ========================================
// Mobile Sidebar Toggle
// ========================================
function toggleSidebar() {
    const sidebar = document.querySelector('.sidebar');
    const overlay = document.querySelector('.sidebar-overlay');
    if (!sidebar) return;

    const isOpen = sidebar.classList.toggle('open');
    if (overlay) overlay.classList.toggle('open', isOpen);
}

// Close sidebar when clicking overlay
document.addEventListener('click', (e) => {
    if (e.target.classList.contains('sidebar-overlay')) {
        toggleSidebar();
    }
});

// ========================================
// Lucide Icons: render on load and after HTMX swaps
// ========================================
function initLucideIcons() {
    if (typeof lucide !== 'undefined') {
        lucide.createIcons();
    }
}
document.addEventListener('DOMContentLoaded', initLucideIcons);
document.body.addEventListener('htmx:afterSwap', initLucideIcons);

// ========================================
// Roles Sidebar: active state toggle
// ========================================
document.addEventListener('click', function(e) {
    var item = e.target.closest('.roles-sidebar-item');
    if (!item) return;
    var nav = item.closest('.roles-sidebar-nav');
    if (!nav) return;
    nav.querySelectorAll('.roles-sidebar-item').forEach(function(el) {
        el.classList.remove('active');
    });
    item.classList.add('active');
});

// ========================================
// Permissions Matrix: toggle-all logic
// ========================================
function initPermMatrixToggles() {
    document.querySelectorAll('.perm-matrix').forEach(function(table) {
        // Set initial state of toggle checkboxes
        table.querySelectorAll('[data-toggle-col]').forEach(function(toggle) {
            updateColToggle(table, toggle);
        });
        table.querySelectorAll('[data-toggle-row]').forEach(function(toggle) {
            updateRowToggle(table, toggle);
        });
    });
}

function updateColToggle(table, toggle) {
    var col = toggle.getAttribute('data-toggle-col');
    var boxes = table.querySelectorAll('input[data-operation="' + col + '"]');
    if (boxes.length === 0) {
        toggle.checked = false;
        toggle.indeterminate = false;
        return;
    }
    var checked = 0;
    for (var i = 0; i < boxes.length; i++) {
        if (boxes[i].checked) checked++;
    }
    toggle.checked = checked === boxes.length;
    toggle.indeterminate = checked > 0 && checked < boxes.length;
}

function updateRowToggle(table, toggle) {
    var res = toggle.getAttribute('data-toggle-row');
    var boxes = table.querySelectorAll('input[data-resource="' + res + '"]');
    if (boxes.length === 0) {
        toggle.checked = false;
        toggle.indeterminate = false;
        return;
    }
    var checked = 0;
    for (var i = 0; i < boxes.length; i++) {
        if (boxes[i].checked) checked++;
    }
    toggle.checked = checked === boxes.length;
    toggle.indeterminate = checked > 0 && checked < boxes.length;
}

document.addEventListener('change', function(e) {
    var target = e.target;
    if (!target.closest('.perm-matrix')) return;
    var table = target.closest('.perm-matrix');

    // Column toggle clicked
    var col = target.getAttribute('data-toggle-col');
    if (col) {
        var boxes = table.querySelectorAll('input[data-operation="' + col + '"]');
        for (var i = 0; i < boxes.length; i++) {
            boxes[i].checked = target.checked;
        }
        // Update all row toggles since columns changed
        table.querySelectorAll('[data-toggle-row]').forEach(function(t) {
            updateRowToggle(table, t);
        });
        return;
    }

    // Row toggle clicked
    var res = target.getAttribute('data-toggle-row');
    if (res) {
        var boxes = table.querySelectorAll('input[data-resource="' + res + '"]');
        for (var i = 0; i < boxes.length; i++) {
            boxes[i].checked = target.checked;
        }
        // Update all column toggles since rows changed
        table.querySelectorAll('[data-toggle-col]').forEach(function(t) {
            updateColToggle(table, t);
        });
        return;
    }

    // Individual permission checkbox changed
    if (target.getAttribute('data-resource') && target.getAttribute('data-operation')) {
        table.querySelectorAll('[data-toggle-col]').forEach(function(t) {
            updateColToggle(table, t);
        });
        table.querySelectorAll('[data-toggle-row]').forEach(function(t) {
            updateRowToggle(table, t);
        });
    }
});

document.addEventListener('DOMContentLoaded', initPermMatrixToggles);
document.body.addEventListener('htmx:afterSwap', initPermMatrixToggles);

// ========================================
// Load Web Components
// ========================================
const componentFiles = [
    '/admin/static/js/components/mcms-toast.js',
    '/admin/static/js/components/mcms-dialog.js',
    '/admin/static/js/components/mcms-confirm.js',
    '/admin/static/js/components/mcms-data-table.js',
    '/admin/static/js/components/mcms-search.js',
    '/admin/static/js/components/mcms-field-renderer.js',
    '/admin/static/js/components/mcms-media-picker.js',
    '/admin/static/js/components/mcms-tree-nav.js',
    '/admin/static/js/components/mcms-file-input.js',
    '/admin/static/js/components/mcms-scroll.js',
    '/admin/static/js/components/mcms-validation-wizard.js',
    '/admin/static/js/components/mcms-media-tree.js',
];

for (const src of componentFiles) {
    const script = document.createElement('script');
    script.src = src;
    script.type = 'module';
    document.head.appendChild(script);
}

// ========================================
// SPA Navigation: sidebar active state
// ========================================
function updateSidebarActive(path) {
    if (!path) return;
    document.querySelectorAll('.sidebar-link').forEach(function(link) {
        var href = link.getAttribute('hx-get') || link.getAttribute('href');
        if (!href) return;
        var active = false;
        if (href === '/admin/' || href === '/admin') {
            active = (path === '/admin/' || path === '/admin');
        } else if (path.indexOf(href) === 0) {
            active = (path.length === href.length || path.charAt(href.length) === '/');
        }
        if (active) {
            link.classList.add('active');
        } else {
            link.classList.remove('active');
        }
    });
}

document.body.addEventListener('htmx:pushedIntoHistory', function(e) {
    updateSidebarActive(e.detail.path);
});

document.body.addEventListener('htmx:historyRestore', function(e) {
    updateSidebarActive(e.detail.path);
});

// ========================================
// SPA Navigation: page title from HX-Trigger
// ========================================
document.body.addEventListener('pageTitle', function(e) {
    if (e.detail && e.detail.value) {
        document.title = e.detail.value + ' \u2014 ModulaCMS';
    }
});

// ========================================
// SPA Navigation: close mobile sidebar after nav swap
// ========================================
document.body.addEventListener('htmx:afterSwap', function(e) {
    if (e.detail.target && e.detail.target.id === 'main-content') {
        var sidebar = document.querySelector('.sidebar');
        var overlay = document.querySelector('.sidebar-overlay');
        if (sidebar && sidebar.classList.contains('open')) {
            sidebar.classList.remove('open');
            if (overlay) overlay.classList.remove('open');
        }
    }
});

// ========================================
// Split button dropdown toggle
// ========================================
document.addEventListener('click', function(e) {
    var toggle = e.target.closest('.btn-split-toggle');
    if (toggle) {
        var split = toggle.closest('.btn-split');
        if (split) split.classList.toggle('open');
        return;
    }
    // Close any open split menus when clicking elsewhere
    document.querySelectorAll('.btn-split.open').forEach(function(el) {
        el.classList.remove('open');
    });
});

// ========================================
// Version history dialog: move restore form into action row
// ========================================
(function() {
    var moved = false;
    var observer = new MutationObserver(function() {
        if (moved) return;
        var dialog = document.getElementById('version-history-dialog');
        if (!dialog) return;
        var form = dialog.querySelector('.version-restore-form');
        var actions = dialog.querySelector('.dialog-actions');
        if (form && actions) {
            actions.insertBefore(form, actions.firstChild);
            moved = true;
        }
    });
    var dialog = document.getElementById('version-history-dialog');
    if (dialog) {
        observer.observe(dialog, { childList: true, subtree: true });
    }
})();

// ========================================
// Version list: click to select, restore button
// ========================================
document.addEventListener('click', function(e) {
    var item = e.target.closest('.version-item[data-version-id]');
    if (!item) return;
    var list = item.closest('.version-list');
    if (!list) return;
    list.querySelectorAll('.version-item').forEach(function(el) {
        el.classList.remove('selected');
    });
    item.classList.add('selected');
    var dialog = document.getElementById('version-history-dialog');
    if (!dialog) return;
    var hiddenInput = dialog.querySelector('#version-restore-id');
    if (hiddenInput) hiddenInput.value = item.dataset.versionId;
    var restoreBtn = dialog.querySelector('#version-restore-form button[type="submit"]');
    if (restoreBtn) restoreBtn.disabled = false;
});

// ========================================
// Clickable table rows: [data-href] on <tr>
// ========================================
document.addEventListener('click', function(e) {
    if (e.target.closest('button, a, input, select, textarea')) return;
    var row = e.target.closest('tr.clickable-row[data-href]');
    if (!row) return;
    var href = row.getAttribute('data-href');
    if (!href) return;
    if (typeof htmx !== 'undefined') {
        htmx.ajax('GET', href, {target: '#main-content', swap: 'innerHTML show:window:top'});
        history.pushState({}, '', href);
        return;
    }
    window.location.href = href;
});

// ========================================
// Dialog openers: [data-opens-dialog] buttons
// ========================================

document.addEventListener('click', function(e) {
    var btn = e.target.closest('[data-opens-dialog]');
    if (!btn) return;
    var dialogId = btn.getAttribute('data-opens-dialog');
    var dialog = document.getElementById(dialogId);
    if (dialog) dialog.setAttribute('open', '');
});

// Delete dialogs: mcms-dialog with [data-delete-url]
document.addEventListener('mcms-dialog:confirm', function(e) {
    var dialog = e.target.closest('mcms-dialog[data-delete-url]');
    if (!dialog) return;
    var url = dialog.getAttribute('data-delete-url');
    var target = dialog.getAttribute('data-delete-target') || '#main-content';
    htmx.ajax('DELETE', url, {target: target, swap: 'outerHTML'});
});

// ========================================
// Block Editor: field panel wiring
// ========================================

// Save initial panel content (root content fields) for restoration on deselect
var _rootPanelHTML = '';
var _rootPanelTitle = '';
(function() {
    var panel = document.getElementById('panel-content');
    var title = document.getElementById('panel-title');
    if (panel) _rootPanelHTML = panel.innerHTML;
    if (title) _rootPanelTitle = title.textContent;
})();

function clearFieldPanel() {
    var panel = document.getElementById('panel-content');
    var title = document.getElementById('panel-title');
    if (panel) panel.innerHTML = _rootPanelHTML;
    if (title) title.textContent = _rootPanelTitle;
}

// Datatype fields cache for newly created blocks
var _dtFieldsCache = {};

function fetchDatatypeFields(datatypeId) {
    if (_dtFieldsCache[datatypeId]) {
        return Promise.resolve(_dtFieldsCache[datatypeId]);
    }
    return fetch('/admin/api/datatypes/' + datatypeId + '/fields', { credentials: 'same-origin' })
        .then(function(res) {
            if (!res.ok) throw new Error('Failed to fetch fields: ' + res.status);
            return res.json();
        })
        .then(function(fields) {
            _dtFieldsCache[datatypeId] = fields;
            return fields;
        });
}

// Block selection listener — opens field panel
document.addEventListener('block-editor:select', function(e) {
    var panel = document.getElementById('panel-content');
    var title = document.getElementById('panel-title');
    if (!panel) return;
    var editor = document.getElementById('content-block-editor');
    if (!editor) return;
    var blockId = e.detail && e.detail.blockId;
    if (!blockId) { clearFieldPanel(); return; }
    var block = editor.getBlock(blockId);
    if (!block) { clearFieldPanel(); return; }
    if (title) title.textContent = block.label || 'Block Fields';
    var fields = editor.getFieldData(blockId);
    if (fields && fields.length > 0) {
        renderFieldPanel(panel, blockId, block, fields);
    } else if (block.datatypeId) {
        fetchDatatypeFields(block.datatypeId).then(function(defs) {
            var empty = defs.map(function(f) {
                return { contentFieldId: '', fieldId: f.fieldId, label: f.label, type: f.type, value: '', toolbar: f.toolbar || null };
            });
            block.fields = empty;
            editor._state.dirty = true;
            editor._updateSaveButton();
            renderFieldPanel(panel, blockId, block, empty);
        });
    } else {
        panel.innerHTML = '<p class="text-muted">No fields for this block type.</p>';
    }
});

// Render field inputs into the panel
function renderFieldPanel(panel, blockId, block, fields) {
    panel.innerHTML = '';
    for (var i = 0; i < fields.length; i++) {
        var f = fields[i];
        var el = document.createElement('mcms-field-renderer');
        el.setAttribute('type', f.type);
        el.setAttribute('name', f.fieldId);
        el.setAttribute('value', f.value || '');
        el.setAttribute('label', f.label);
        el.dataset.blockId = blockId;
        el.dataset.fieldId = f.fieldId;
        if (f.type === 'richtext') {
            var tb = f.toolbar;
            if (!tb && block.datatypeId && _dtFieldsCache[block.datatypeId]) {
                var cached = _dtFieldsCache[block.datatypeId];
                for (var j = 0; j < cached.length; j++) {
                    if (cached[j].fieldId === f.fieldId && cached[j].toolbar) {
                        tb = cached[j].toolbar;
                        break;
                    }
                }
            }
            if (tb) el.setAttribute('toolbar', JSON.stringify(tb));
        }
        panel.appendChild(el);
    }
}

// Field change listener — updates block state or tracks root content field changes
document.addEventListener('field-change', function(e) {
    var renderer = e.target.closest('mcms-field-renderer');
    if (!renderer) return;

    // Block field: update editor state
    if (renderer.dataset.blockId) {
        var editor = document.getElementById('content-block-editor');
        if (editor) editor.setFieldValue(renderer.dataset.blockId, renderer.dataset.fieldId, e.detail.value);
        return;
    }

    // Root content field: track as pending save
    if (renderer.dataset.contentDataId) {
        if (!_pendingRootFieldUpdates) _pendingRootFieldUpdates = {};
        _pendingRootFieldUpdates[renderer.dataset.fieldId] = {
            content_data_id: renderer.dataset.contentDataId,
            content_field_id: renderer.dataset.contentFieldId || '',
            field_id: renderer.dataset.fieldId,
            value: e.detail.value
        };
        // Mark block editor dirty so Save button activates
        var editor = document.getElementById('content-block-editor');
        if (editor && editor._state) {
            editor._state.dirty = true;
            editor._updateSaveButton();
        }
    }
});

var _pendingRootFieldUpdates = null;

// ========================================
// Block Editor: auto-save
// ========================================
(function() {
    var INTERVAL = 30000; // 30 seconds
    var DEBOUNCE = 2000;  // 2 seconds after input
    var debounceTimer = null;
    var intervalTimer = null;
    var saving = false;

    function autoSave() {
        if (saving) return;
        var editor = document.getElementById('content-block-editor');
        if (!editor) return;
        var hasPendingRoot = _pendingRootFieldUpdates && Object.keys(_pendingRootFieldUpdates).length > 0;
        if (!editor.dirty && !hasPendingRoot) return;
        saving = true;
        editor.save();
    }

    // Reset saving flag when save completes
    document.addEventListener('block-editor:save-complete', function() {
        saving = false;
    });

    // Debounce: 2s after any field change
    document.addEventListener('field-change', function() {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(autoSave, DEBOUNCE);
    });

    // Also debounce on block tree mutations (add, move, delete, indent, outdent)
    document.addEventListener('block-editor:change', function() {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(autoSave, DEBOUNCE);
    });

    // Interval: every 30s as a safety net
    function startInterval() {
        if (intervalTimer) return;
        intervalTimer = setInterval(autoSave, INTERVAL);
    }

    // Start interval when a block editor exists on the page
    if (document.getElementById('content-block-editor')) {
        startInterval();
    }
    // Also start after HTMX swaps in case the editor loads via SPA navigation
    document.body.addEventListener('htmx:afterSwap', function() {
        if (document.getElementById('content-block-editor')) {
            startInterval();
        } else {
            clearInterval(intervalTimer);
            intervalTimer = null;
        }
    });
})();

// ========================================
// Block Editor: save wiring
// ========================================

// nullableID converts an empty string to null for the batch API's nullable ID fields.
// Non-empty strings pass through as-is.
function nullableID(val) {
    return val ? val : null;
}

document.addEventListener('block-editor:save', function(e) {
    var editor = e.target;
    if (!editor || !editor.id) return;

    var stateStr = e.detail && e.detail.state;
    if (!stateStr) {
        editor.dispatchEvent(new CustomEvent('block-editor:save-complete', { bubbles: true, composed: true, detail: { success: false } }));
        return;
    }

    var state;
    try {
        state = JSON.parse(stateStr);
    } catch (parseErr) {
        showBlockEditorToast('Failed to parse editor state', 'error');
        editor.dispatchEvent(new CustomEvent('block-editor:save-complete', { bubbles: true, composed: true, detail: { success: false } }));
        return;
    }

    if (!state.blocks) {
        editor.dispatchEvent(new CustomEvent('block-editor:save-complete', { bubbles: true, composed: true, detail: { success: false } }));
        return;
    }

    // Read initial state from data-state attribute to diff against
    var initialStr = editor.getAttribute('data-state');
    var initial = {};
    if (initialStr) {
        try {
            var parsed = JSON.parse(initialStr);
            initial = parsed.blocks || {};
        } catch (ignored) {
            // If initial parse fails, treat all blocks as changed
        }
    }

    var contentId = editor.getAttribute('data-content-id') || '';

    // Detect creates: blocks in current state with no matching initial entry
    var creates = [];
    // Detect updates: existing blocks with changed pointer fields
    var updates = [];
    var currentKeys = Object.keys(state.blocks);
    for (var i = 0; i < currentKeys.length; i++) {
        var id = currentKeys[i];
        var block = state.blocks[id];
        var orig = initial[id];
        if (!orig) {
            // New block — needs create. datatypeId is set by the block editor's
            // datatype picker; blocks without one are skipped.
            if (!block.datatypeId) continue;
            creates.push({
                client_id: block.id,
                datatype_id: block.datatypeId,
                parent_id: nullableID(block.parentId),
                first_child_id: nullableID(block.firstChildId),
                next_sibling_id: nullableID(block.nextSiblingId),
                prev_sibling_id: nullableID(block.prevSiblingId)
            });
            continue;
        }
        if ((block.parentId || '') !== (orig.parentId || '') ||
            (block.firstChildId || '') !== (orig.firstChildId || '') ||
            (block.nextSiblingId || '') !== (orig.nextSiblingId || '') ||
            (block.prevSiblingId || '') !== (orig.prevSiblingId || '')) {
            updates.push({
                content_data_id: block.id,
                parent_id: nullableID(block.parentId),
                first_child_id: nullableID(block.firstChildId),
                next_sibling_id: nullableID(block.nextSiblingId),
                prev_sibling_id: nullableID(block.prevSiblingId)
            });
        }
    }

    // Detect deletes: blocks in initial state but absent from current state
    var deletes = [];
    var initialKeys = Object.keys(initial);
    for (var j = 0; j < initialKeys.length; j++) {
        if (!state.blocks[initialKeys[j]]) {
            deletes.push(initialKeys[j]);
        }
    }

    // Compute field updates: changed or new field values
    var fieldUpdates = [];
    for (var fi = 0; fi < currentKeys.length; fi++) {
        var fid = currentKeys[fi];
        var fblock = state.blocks[fid];
        var forig = initial[fid];
        if (!fblock.fields) continue;
        for (var fj = 0; fj < fblock.fields.length; fj++) {
            var f = fblock.fields[fj];
            var origField = null;
            if (forig && forig.fields) {
                for (var fk = 0; fk < forig.fields.length; fk++) {
                    if (forig.fields[fk].fieldId === f.fieldId) {
                        origField = forig.fields[fk];
                        break;
                    }
                }
            }
            if (!origField || origField.value !== f.value) {
                fieldUpdates.push({
                    content_data_id: fblock.id,
                    content_field_id: f.contentFieldId || '',
                    field_id: f.fieldId,
                    value: f.value
                });
            }
        }
    }

    // Include pending root content field updates
    if (_pendingRootFieldUpdates) {
        var rootKeys = Object.keys(_pendingRootFieldUpdates);
        for (var ri = 0; ri < rootKeys.length; ri++) {
            fieldUpdates.push(_pendingRootFieldUpdates[rootKeys[ri]]);
        }
    }

    if (creates.length === 0 && updates.length === 0 && deletes.length === 0 && fieldUpdates.length === 0) {
        showBlockEditorToast('No changes to save', 'info');
        editor.dispatchEvent(new CustomEvent('block-editor:save-complete', { bubbles: true, composed: true, detail: { success: true } }));
        return;
    }

    var body = {
        content_id: contentId,
        creates: creates,
        updates: updates,
        deletes: deletes,
        field_updates: fieldUpdates
    };

    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/admin/content/tree');
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.setRequestHeader('HX-Request', 'true');

    // Include CSRF token
    var meta = document.querySelector('meta[name="csrf-token"]');
    if (meta && meta.content) {
        xhr.setRequestHeader('X-CSRF-Token', meta.content);
    }

    xhr.onload = function() {
        if (xhr.status !== 200) {
            showBlockEditorToast('Save failed: server error', 'error');
            editor.dispatchEvent(new CustomEvent('block-editor:save-complete', { bubbles: true, composed: true, detail: { success: false } }));
            return;
        }
        var resp;
        try {
            resp = JSON.parse(xhr.responseText);
        } catch (ignored) {
            showBlockEditorToast('Content structure saved', 'success');
            editor.setAttribute('data-state', stateStr);
            editor.dispatchEvent(new CustomEvent('block-editor:save-complete', { bubbles: true, composed: true, detail: { success: true } }));
            return;
        }

        // Remap client UUIDs to server ULIDs in editor state so the next
        // save diff treats them as existing blocks, not new creates.
        if (resp.id_map && Object.keys(resp.id_map).length > 0) {
            remapEditorState(editor, state, resp.id_map);
            // Rebuild stateStr from the remapped state
            stateStr = JSON.stringify(state);
        }

        if (resp.errors && resp.errors.length > 0) {
            showBlockEditorToast('Partial save: ' + resp.errors.length + ' error(s)', 'error');
        } else {
            var parts = [];
            if (resp.created > 0) parts.push(resp.created + ' created');
            if (resp.updated > 0) parts.push(resp.updated + ' updated');
            if (resp.deleted > 0) parts.push(resp.deleted + ' deleted');
            if (resp.fields_updated > 0) parts.push(resp.fields_updated + ' field(s) saved');
            showBlockEditorToast(parts.length > 0 ? parts.join(', ') : 'Content saved', 'success');
        }
        // Clear pending root field updates on success
        _pendingRootFieldUpdates = null;
        // Update data-state so next save diffs correctly
        editor.setAttribute('data-state', stateStr);
        var saveSuccess = !(resp.errors && resp.errors.length > 0);
        editor.dispatchEvent(new CustomEvent('block-editor:save-complete', { bubbles: true, composed: true, detail: { success: saveSuccess } }));
    };
    xhr.onerror = function() {
        showBlockEditorToast('Network error saving content tree', 'error');
        editor.dispatchEvent(new CustomEvent('block-editor:save-complete', { bubbles: true, composed: true, detail: { success: false } }));
    };
    xhr.send(JSON.stringify(body));
});

// remapEditorState replaces client UUIDs with server ULIDs throughout the
// editor's in-memory state and block map. This ensures the next save treats
// previously-new blocks as existing (update path, not create path).
function remapEditorState(editor, state, idMap) {
    var clientIds = Object.keys(idMap);
    for (var i = 0; i < clientIds.length; i++) {
        var clientId = clientIds[i];
        var serverId = idMap[clientId];
        var block = state.blocks[clientId];
        if (!block) continue;

        // Update the block's own ID
        block.id = serverId;

        // Move the block entry to the new key
        state.blocks[serverId] = block;
        delete state.blocks[clientId];

        // Update rootId if it pointed to this client ID
        if (state.rootId === clientId) {
            state.rootId = serverId;
        }
    }

    // Second pass: remap pointer fields across ALL blocks that may reference
    // client IDs (both newly created and existing blocks).
    var allKeys = Object.keys(state.blocks);
    for (var j = 0; j < allKeys.length; j++) {
        var b = state.blocks[allKeys[j]];
        if (b.parentId && idMap[b.parentId]) b.parentId = idMap[b.parentId];
        if (b.firstChildId && idMap[b.firstChildId]) b.firstChildId = idMap[b.firstChildId];
        if (b.nextSiblingId && idMap[b.nextSiblingId]) b.nextSiblingId = idMap[b.nextSiblingId];
        if (b.prevSiblingId && idMap[b.prevSiblingId]) b.prevSiblingId = idMap[b.prevSiblingId];
    }

    // Notify the block editor component to update its internal references
    if (typeof editor.remapIds === 'function') {
        editor.remapIds(idMap);
    }
}

// ========================================
// Dialog Utilities
// ========================================

// showConfirmDialog creates a temporary mcms-dialog, opens it, and returns
// a Promise that resolves true on confirm, false on cancel.
function showConfirmDialog(options) {
    console.log('[confirm-debug] showConfirmDialog called', options);
    var defined = customElements.get('mcms-dialog');
    console.log('[confirm-debug] mcms-dialog defined:', !!defined);
    return new Promise(function(resolve) {
        var dialog = document.createElement('mcms-dialog');
        console.log('[confirm-debug] created dialog element, constructor:', dialog.constructor.name);
        dialog.setAttribute('title', options.title || 'Confirm');
        dialog.setAttribute('confirm-label', options.confirmLabel || 'Confirm');
        dialog.setAttribute('cancel-label', options.cancelLabel || 'Cancel');
        if (options.destructive) dialog.setAttribute('destructive', '');

        if (options.message) {
            var p = document.createElement('p');
            p.textContent = options.message;
            dialog.appendChild(p);
        }

        function cleanup(result) {
            console.log('[confirm-debug] dialog resolved:', result);
            dialog.removeEventListener('mcms-dialog:confirm', onConfirm);
            dialog.removeEventListener('mcms-dialog:cancel', onCancel);
            dialog.remove();
            resolve(result);
        }

        function onConfirm() { cleanup(true); }
        function onCancel() { cleanup(false); }

        dialog.addEventListener('mcms-dialog:confirm', onConfirm);
        dialog.addEventListener('mcms-dialog:cancel', onCancel);

        document.body.appendChild(dialog);
        console.log('[confirm-debug] dialog appended to body, setting open');
        dialog.setAttribute('open', '');
    });
}

// ========================================
// HTMX: replace native confirm() with mcms-dialog
// ========================================

document.body.addEventListener('htmx:confirm', function(e) {
    console.log('[confirm-debug] htmx:confirm fired', {
        question: e.detail.question,
        path: e.detail.path,
        target: e.target,
        tagName: e.target.tagName,
        classList: e.target.className,
        hxConfirmAttr: e.target.getAttribute('hx-confirm'),
        defaultPrevented: e.defaultPrevented,
    });
    // Only intercept requests that have an hx-confirm attribute
    if (!e.detail.question) {
        console.log('[confirm-debug] no question, skipping (native confirm will fire)');
        return;
    }
    e.preventDefault();
    console.log('[confirm-debug] preventDefault called, native confirm blocked');

    var question = e.detail.question;
    var path = e.detail.path || '';
    var questionLower = question.toLowerCase();

    // Infer dialog style from the question text and request path
    var destructive = questionLower.indexOf('delete') !== -1
        || questionLower.indexOf('revoke') !== -1
        || path.indexOf('/unpublish') !== -1;

    var title = 'Confirm';
    var confirmLabel = 'Confirm';

    if (questionLower.indexOf('delete') !== -1) {
        title = 'Delete';
        confirmLabel = 'Delete';
    } else if (questionLower.indexOf('revoke') !== -1) {
        title = 'Revoke';
        confirmLabel = 'Revoke';
    } else if (path.indexOf('/unpublish') !== -1) {
        title = 'Unpublish';
        confirmLabel = 'Unpublish';
    } else if (path.indexOf('/publish') !== -1) {
        title = 'Publish';
        confirmLabel = 'Publish';
    } else if (questionLower.indexOf('restore') !== -1) {
        title = 'Restore';
        confirmLabel = 'Restore';
    }

    console.log('[confirm-debug] opening dialog:', { title: title, confirmLabel: confirmLabel, destructive: destructive });

    showConfirmDialog({
        title: title,
        message: question,
        confirmLabel: confirmLabel,
        destructive: destructive,
    }).then(function(confirmed) {
        console.log('[confirm-debug] dialog promise resolved, confirmed:', confirmed);
        if (!confirmed) return;

        // Save block editor before publish/unpublish
        if (path.indexOf('/publish') !== -1) {
            var editor = document.getElementById('content-block-editor');
            console.log('[confirm-debug] publish path detected, editor:', !!editor, 'dirty:', editor && editor.dirty);
            if (editor && editor.dirty) {
                console.log('[confirm-debug] editor dirty, saving before publish');
                function onSaveComplete(saveEvent) {
                    editor.removeEventListener('block-editor:save-complete', onSaveComplete);
                    console.log('[confirm-debug] save complete, success:', saveEvent.detail && saveEvent.detail.success);
                    if (saveEvent.detail && saveEvent.detail.success) {
                        e.detail.issueRequest();
                    }
                }
                editor.addEventListener('block-editor:save-complete', onSaveComplete);
                editor.save();
                return;
            }
        }

        console.log('[confirm-debug] issuing request');
        e.detail.issueRequest();
    });
});

// Log all htmx:confirm events at the document level (captures before body)
document.addEventListener('htmx:confirm', function(e) {
    console.log('[confirm-debug] htmx:confirm at DOCUMENT level', {
        question: e.detail.question,
        target: e.target.tagName + '.' + e.target.className,
        defaultPrevented: e.defaultPrevented,
    });
}, true);

function showBlockEditorToast(message, type) {
    var toast = document.querySelector('mcms-toast');
    if (toast) {
        toast.show(message, type);
    }
}
