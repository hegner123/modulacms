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
];

for (const src of componentFiles) {
    const script = document.createElement('script');
    script.src = src;
    script.type = 'module';
    document.head.appendChild(script);
}
