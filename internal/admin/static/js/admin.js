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
    if (key === 'g') {
        pendingPrefix = 'g';
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
    if (!stateStr) return;

    var state;
    try {
        state = JSON.parse(stateStr);
    } catch (parseErr) {
        showBlockEditorToast('Failed to parse editor state', 'error');
        return;
    }

    if (!state.blocks) return;

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

    if (creates.length === 0 && updates.length === 0 && deletes.length === 0) {
        showBlockEditorToast('No structural changes to save', 'info');
        return;
    }

    var body = {
        content_id: contentId,
        creates: creates,
        updates: updates,
        deletes: deletes
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
            return;
        }
        var resp;
        try {
            resp = JSON.parse(xhr.responseText);
        } catch (ignored) {
            showBlockEditorToast('Content structure saved', 'success');
            editor.setAttribute('data-state', stateStr);
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
            showBlockEditorToast(parts.length > 0 ? parts.join(', ') : 'Content structure saved', 'success');
        }
        // Update data-state so next save diffs correctly
        editor.setAttribute('data-state', stateStr);
    };
    xhr.onerror = function() {
        showBlockEditorToast('Network error saving content tree', 'error');
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

function showBlockEditorToast(message, type) {
    var toast = document.querySelector('mcms-toast');
    if (toast) {
        toast.show(message, type);
    }
}
