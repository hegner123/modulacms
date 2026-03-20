/**
 * @module mcms-command-palette
 * @description Command palette (Light DOM) for the ModulaCMS admin panel.
 *
 * Opens on Cmd+K / Ctrl+K. Searches across admin pages and config settings.
 * Pages are sourced from a data attribute; settings are fetched from the
 * search index API on first open and cached.
 *
 * Results are navigated with arrow keys and Enter. Selecting a page does
 * HTMX SPA navigation; selecting a setting navigates to /admin/settings
 * and scrolls to the field.
 *
 * @example
 * <mcms-command-palette></mcms-command-palette>
 */

class McmsCommandPalette extends HTMLElement {
    constructor() {
        super();
        this._built = false;
        this._open = false;
        this._activeIndex = -1;
        this._results = [];
        this._pages = [];
        this._settings = null; // cached search index
        this._query = '';
        this._boundKeyDown = this._onGlobalKeyDown.bind(this);
    }

    connectedCallback() {
        if (!this._built) this._build();
        document.addEventListener('keydown', this._boundKeyDown);
    }

    disconnectedCallback() {
        document.removeEventListener('keydown', this._boundKeyDown);
    }

    _build() {
        this._built = true;

        // Parse pages from nav items embedded as JSON data attribute,
        // or fall back to scraping the sidebar.
        this._pages = this._parsePages();

        // Build DOM
        this._backdrop = document.createElement('div');
        this._backdrop.className = 'fixed inset-0 bg-gray-900/50 z-[9998] hidden';
        this._backdrop.addEventListener('click', () => this.close());

        this._panel = document.createElement('div');
        this._panel.className = 'fixed inset-0 z-[9999] overflow-y-auto p-4 sm:p-6 md:p-20 hidden';
        this._panel.setAttribute('role', 'dialog');
        this._panel.setAttribute('aria-modal', 'true');
        this._panel.setAttribute('aria-label', 'Command palette');

        const container = document.createElement('div');
        container.className = 'mx-auto max-w-xl transform overflow-hidden rounded-xl bg-white/5 backdrop-blur-xl shadow-2xl ring-1 ring-white/10 transition-all';

        // Search input row
        const inputRow = document.createElement('div');
        inputRow.className = 'relative';

        const searchIcon = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        searchIcon.setAttribute('viewBox', '0 0 20 20');
        searchIcon.setAttribute('fill', 'currentColor');
        searchIcon.setAttribute('aria-hidden', 'true');
        searchIcon.setAttribute('class', 'pointer-events-none absolute left-4 top-3.5 size-5 text-gray-500');
        const iconPath = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        iconPath.setAttribute('fill-rule', 'evenodd');
        iconPath.setAttribute('clip-rule', 'evenodd');
        iconPath.setAttribute('d', 'M9 3.5a5.5 5.5 0 1 0 0 11 5.5 5.5 0 0 0 0-11ZM2 9a7 7 0 1 1 12.452 4.391l3.328 3.329a.75.75 0 1 1-1.06 1.06l-3.329-3.328A7 7 0 0 1 2 9Z');
        searchIcon.appendChild(iconPath);

        this._input = document.createElement('input');
        this._input.type = 'text';
        this._input.placeholder = 'Search pages and settings...';
        this._input.className = 'h-12 w-full border-0 bg-transparent pr-4 pl-11 text-sm text-white outline-none placeholder:text-gray-500 focus:ring-0';
        this._input.addEventListener('input', () => this._onInput());
        this._input.addEventListener('keydown', (e) => this._onInputKeyDown(e));

        inputRow.appendChild(searchIcon);
        inputRow.appendChild(this._input);

        // Results container
        this._resultsList = document.createElement('div');
        this._resultsList.className = 'max-h-80 scroll-py-2 overflow-y-auto';
        this._resultsList.setAttribute('role', 'listbox');
        this._resultsList.id = 'command-palette-results';
        this._input.setAttribute('role', 'combobox');
        this._input.setAttribute('aria-expanded', 'false');
        this._input.setAttribute('aria-controls', 'command-palette-results');
        this._input.setAttribute('aria-autocomplete', 'list');

        // Empty state
        this._emptyState = document.createElement('div');
        this._emptyState.className = 'px-6 py-14 text-center text-sm hidden';
        this._emptyState.innerHTML = '<p class="font-semibold text-white">No results found</p><p class="mt-2 text-gray-400">Try a different search term.</p>';

        // Footer
        this._footer = document.createElement('div');
        this._footer.className = 'flex flex-wrap items-center border-t border-white/10 px-4 py-2.5 text-xs text-gray-400';
        this._footer.innerHTML = '<kbd class="mr-1 rounded border border-white/10 bg-white/5 px-1.5 py-0.5 font-mono text-gray-300">&uarr;&darr;</kbd> navigate <kbd class="mx-1 rounded border border-white/10 bg-white/5 px-1.5 py-0.5 font-mono text-gray-300">&crarr;</kbd> select <kbd class="ml-1 rounded border border-white/10 bg-white/5 px-1.5 py-0.5 font-mono text-gray-300">esc</kbd> close';

        container.appendChild(inputRow);
        container.appendChild(this._resultsList);
        container.appendChild(this._emptyState);
        container.appendChild(this._footer);
        this._panel.appendChild(container);

        this.appendChild(this._backdrop);
        this.appendChild(this._panel);
    }

    // --- Public API ---

    open() {
        if (this._open) return;
        this._open = true;
        this._backdrop.classList.remove('hidden');
        this._panel.classList.remove('hidden');
        this._input.value = '';
        this._query = '';
        this._activeIndex = -1;
        this._input.setAttribute('aria-expanded', 'true');
        this._renderInitial();
        // Delay focus so the dialog transition doesn't steal it.
        requestAnimationFrame(() => this._input.focus());
        this._loadSettings();
    }

    close() {
        if (!this._open) return;
        this._open = false;
        this._backdrop.classList.add('hidden');
        this._panel.classList.add('hidden');
        this._input.setAttribute('aria-expanded', 'false');
        this._resultsList.innerHTML = '';
    }

    toggle() {
        if (this._open) this.close(); else this.open();
    }

    // --- Data loading ---

    _parsePages() {
        // Scrape sidebar nav links.
        const links = document.querySelectorAll('[data-nav-items] a[href]');
        const pages = [];
        let currentSection = '';
        for (const link of links) {
            const href = link.getAttribute('href');
            const label = link.textContent.trim();
            if (!href || !label) continue;
            // Detect section from parent structure.
            const sectionEl = link.closest('[data-section]');
            const section = sectionEl ? sectionEl.dataset.section : '';
            if (section !== currentSection) currentSection = section;
            pages.push({
                type: 'page',
                label,
                href,
                section: currentSection,
                searchText: (label + ' ' + currentSection).toLowerCase(),
            });
        }
        // Fallback: if no sidebar nav found, use known pages.
        if (pages.length === 0) {
            return this._defaultPages();
        }

        // Add settings sections as searchable targets.
        pages.push(...this._parseSettingsSections());

        return pages;
    }

    _parseSettingsSections() {
        // Settings sections are rendered with [data-settings-section] attributes.
        // These are only present when the settings page is loaded, so we use
        // a static list that matches the settingsSection() calls in settings.templ.
        const sections = [
            { title: 'General', description: 'Ports, hosts, site URLs' },
            { title: 'Database', description: 'Driver and connection' },
            { title: 'Remote', description: 'Remote CMS sync' },
            { title: 'Storage (S3)', description: 'Media and backup storage' },
            { title: 'Backup', description: 'Backup paths' },
            { title: 'Authentication', description: 'Cookies and auth salt' },
            { title: 'OAuth', description: 'OAuth provider' },
            { title: 'CORS', description: 'Cross-origin settings' },
            { title: 'Email', description: 'SMTP and email providers' },
            { title: 'Content', description: 'Publishing, versioning, composition' },
            { title: 'Plugins', description: 'Plugin directory, VM limits, hooks' },
            { title: 'Observability', description: 'Error tracking, performance' },
            { title: 'Updates', description: 'Auto-update preferences' },
            { title: 'Deploy', description: 'Deploy environments, snapshots' },
            { title: 'MCP', description: 'Model Context Protocol server' },
            { title: 'Search', description: 'Full-text search index' },
            { title: 'Internationalization', description: 'Multi-language support' },
            { title: 'Webhooks', description: 'Webhook delivery and retries' },
            { title: 'Keybindings', description: 'TUI keyboard shortcuts' },
        ];
        return sections.map(s => {
            const id = 'section-' + s.title.toLowerCase().replace(/\s+/g, '-').replace(/[()]/g, '');
            return {
                type: 'section',
                label: s.title,
                description: s.description,
                sectionId: id,
                section: 'Settings',
                searchText: (s.title + ' ' + s.description + ' settings').toLowerCase(),
            };
        });
    }

    _defaultPages() {
        const items = [
            { label: 'Dashboard', href: '/admin/', section: '' },
            { label: 'Content', href: '/admin/content', section: 'Content' },
            { label: 'Datatypes', href: '/admin/datatypes', section: 'Content' },
            { label: 'Field Types', href: '/admin/field-types', section: 'Content' },
            { label: 'Routes', href: '/admin/routes', section: 'Content' },
            { label: 'Media', href: '/admin/media', section: 'Media' },
            { label: 'Dimensions', href: '/admin/media/dimensions', section: 'Media' },
            { label: 'Users', href: '/admin/users', section: 'Users & Access' },
            { label: 'Roles', href: '/admin/users/roles', section: 'Users & Access' },
            { label: 'Tokens', href: '/admin/users/tokens', section: 'Users & Access' },
            { label: 'Plugins', href: '/admin/plugins', section: 'System' },
            { label: 'Tables', href: '/admin/tables', section: 'System' },
            { label: 'Import', href: '/admin/import', section: 'System' },
            { label: 'Audit', href: '/admin/audit', section: 'System' },
            { label: 'Settings', href: '/admin/settings', section: 'Settings' },
            { label: 'Locales', href: '/admin/settings/locales', section: 'Settings' },
            { label: 'Webhooks', href: '/admin/settings/webhooks', section: 'Settings' },
        ];
        return items.map(i => ({
            type: 'page',
            label: i.label,
            href: i.href,
            section: i.section,
            searchText: (i.label + ' ' + i.section).toLowerCase(),
        }));
    }

    async _loadSettings() {
        if (this._settings) return;
        try {
            const resp = await fetch('/api/v1/admin/config/search-index', {
                credentials: 'same-origin',
            });
            if (!resp.ok) return;
            const data = await resp.json();
            this._settings = data.map(s => ({
                type: 'setting',
                key: s.key,
                label: s.label,
                category: s.category_label,
                description: s.description,
                searchText: (s.key + ' ' + s.label + ' ' + s.description + ' ' + s.category_label + ' ' + (s.help_text || '')).toLowerCase(),
            }));
            // Re-render if query is active to include settings.
            if (this._query) this._onInput();
        } catch {
            // Settings search unavailable — page search still works.
        }
    }

    // --- Search ---

    _search(query) {
        if (!query) return [];

        const q = query.toLowerCase();
        const terms = q.split(/\s+/).filter(Boolean);
        const results = [];

        const match = (item) => {
            return terms.every(t => item.searchText.includes(t));
        };

        // Score: exact prefix on label > includes label > includes searchText
        const score = (item) => {
            const lbl = item.label.toLowerCase();
            if (lbl === q) return 100;
            if (lbl.startsWith(q)) return 80;
            if (item.key && item.key === q) return 90;
            if (lbl.includes(q)) return 60;
            return 40;
        };

        for (const page of this._pages) {
            if (match(page)) results.push({ ...page, score: score(page) });
        }
        if (this._settings) {
            for (const setting of this._settings) {
                if (match(setting)) results.push({ ...setting, score: score(setting) });
            }
        }

        const typeOrder = { page: 0, section: 1, setting: 2 };
        results.sort((a, b) => {
            // Pages before sections before settings at equal score.
            if (b.score !== a.score) return b.score - a.score;
            const ta = typeOrder[a.type] ?? 3;
            const tb = typeOrder[b.type] ?? 3;
            if (ta !== tb) return ta - tb;
            return a.label.localeCompare(b.label);
        });

        return results.slice(0, 20);
    }

    // --- Rendering ---

    _renderInitial() {
        // Show recent / top pages when no query.
        this._results = this._pages.slice(0, 8);
        this._renderResults('Pages');
    }

    _onInput() {
        this._query = this._input.value.trim();
        this._activeIndex = -1;
        if (!this._query) {
            this._renderInitial();
            return;
        }
        this._results = this._search(this._query);
        if (this._results.length === 0) {
            this._resultsList.innerHTML = '';
            this._resultsList.classList.add('hidden');
            this._emptyState.classList.remove('hidden');
        } else {
            this._emptyState.classList.add('hidden');
            this._resultsList.classList.remove('hidden');
            this._renderResults();
        }
    }

    _renderResults(defaultGroupLabel) {
        this._resultsList.innerHTML = '';
        this._resultsList.classList.remove('hidden');
        this._emptyState.classList.add('hidden');

        // Group results by type.
        let currentType = '';
        let globalIdx = 0;
        const fragment = document.createDocumentFragment();

        for (const item of this._results) {
            // Group header when type changes.
            const groupLabel = item.type === 'page' ? (defaultGroupLabel || 'Pages')
                : item.type === 'section' ? 'Settings Sections'
                : 'Settings';
            if (item.type !== currentType) {
                currentType = item.type;
                const header = document.createElement('h3');
                header.className = 'px-4 pt-3 pb-1.5 text-xs font-semibold text-gray-400';
                header.textContent = groupLabel;
                fragment.appendChild(header);
            }

            const el = document.createElement('div');
            el.setAttribute('role', 'option');
            el.setAttribute('aria-selected', 'false');
            el.dataset.index = globalIdx;
            el.className = 'group flex cursor-default items-center gap-3 px-4 py-2 text-sm text-gray-300 select-none';

            if (item.type === 'page') {
                el.innerHTML = `<svg class="size-5 shrink-0 text-gray-500 group-aria-selected:text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 0 0 3 8.25v10.5A2.25 2.25 0 0 0 5.25 21h10.5A2.25 2.25 0 0 0 18 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25"/></svg>`
                    + `<span class="flex-auto truncate">${this._esc(item.label)}</span>`
                    + (item.section ? `<span class="shrink-0 text-xs text-gray-500 group-aria-selected:text-gray-300">${this._esc(item.section)}</span>` : '');
            } else if (item.type === 'section') {
                // Hash/anchor icon for settings sections.
                el.innerHTML = `<svg class="size-5 shrink-0 text-gray-500 group-aria-selected:text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M5.25 8.25h15m-16.5 7.5h15m-1.8-13.5-3.9 19.5m-2.1-19.5-3.9 19.5"/></svg>`
                    + `<span class="flex-auto truncate">${this._esc(item.label)}</span>`
                    + `<span class="shrink-0 text-xs text-gray-500 group-aria-selected:text-gray-300">${this._esc(item.description)}</span>`;
            } else {
                el.innerHTML = `<svg class="size-5 shrink-0 text-gray-500 group-aria-selected:text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z"/><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/></svg>`
                    + `<span class="flex-auto truncate">${this._esc(item.label)}</span>`
                    + `<span class="shrink-0 text-xs text-gray-500 group-aria-selected:text-gray-300">${this._esc(item.category)}</span>`;
            }

            el.addEventListener('click', () => this._select(item));
            el.addEventListener('mouseenter', () => {
                this._activeIndex = parseInt(el.dataset.index, 10);
                this._highlightActive();
            });
            fragment.appendChild(el);
            globalIdx++;
        }

        this._resultsList.appendChild(fragment);
    }

    _highlightActive() {
        const items = this._resultsList.querySelectorAll('[role="option"]');
        for (const item of items) {
            const idx = parseInt(item.dataset.index, 10);
            const selected = idx === this._activeIndex;
            item.setAttribute('aria-selected', selected ? 'true' : 'false');
            if (selected) {
                item.classList.add('bg-white/10', 'text-white');
                item.scrollIntoView({ block: 'nearest' });
            } else {
                item.classList.remove('bg-white/10', 'text-white');
            }
        }
        this._input.setAttribute('aria-activedescendant',
            this._activeIndex >= 0 ? '' : '');
    }

    // --- Selection ---

    _select(item) {
        this.close();
        if (item.type === 'page') {
            this._navigateTo(item.href);
        } else if (item.type === 'setting') {
            this._navigateToSetting(item.key);
        } else if (item.type === 'section') {
            this._navigateToSection(item.sectionId);
        }
    }

    _navigateTo(url) {
        if (typeof htmx !== 'undefined') {
            htmx.ajax('GET', url, { target: '#main-content', swap: 'innerHTML show:window:top' });
            history.pushState({}, '', url);
        } else {
            window.location.href = url;
        }
    }

    _navigateToSetting(key) {
        const settingsPath = '/admin/settings';
        const currentPath = window.location.pathname;

        const scrollToField = () => {
            // Try to find and scroll to the field input by name.
            requestAnimationFrame(() => {
                const el = document.querySelector(`[name="${key}"]`)
                    || document.getElementById(key);
                if (el) {
                    el.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    el.focus();
                    // Brief highlight effect.
                    const wrapper = el.closest('.space-y-1, .sm\\:grid') || el.parentElement;
                    if (wrapper) {
                        wrapper.classList.add('ring-2', 'ring-[var(--color-primary)]', 'rounded-lg');
                        setTimeout(() => {
                            wrapper.classList.remove('ring-2', 'ring-[var(--color-primary)]', 'rounded-lg');
                        }, 2000);
                    }
                }
            });
        };

        if (currentPath === settingsPath || currentPath === settingsPath + '/') {
            scrollToField();
        } else {
            // Navigate to settings first, then scroll.
            if (typeof htmx !== 'undefined') {
                htmx.ajax('GET', settingsPath, { target: '#main-content', swap: 'innerHTML show:window:top' });
                history.pushState({}, '', settingsPath);
                // Wait for HTMX swap, then scroll.
                document.body.addEventListener('htmx:afterSettle', function onSwap() {
                    document.body.removeEventListener('htmx:afterSettle', onSwap);
                    scrollToField();
                });
            } else {
                window.location.href = settingsPath + '#' + key;
            }
        }
    }

    _navigateToSection(sectionId) {
        const settingsPath = '/admin/settings';
        const currentPath = window.location.pathname;

        const scrollToSection = () => {
            requestAnimationFrame(() => {
                const el = document.getElementById(sectionId);
                if (el) {
                    el.scrollIntoView({ behavior: 'smooth', block: 'start' });
                    // Brief highlight on the section heading.
                    el.classList.add('ring-2', 'ring-[var(--color-primary)]', 'rounded-lg', 'p-2', '-m-2');
                    setTimeout(() => {
                        el.classList.remove('ring-2', 'ring-[var(--color-primary)]', 'rounded-lg', 'p-2', '-m-2');
                    }, 2000);
                }
            });
        };

        if (currentPath === settingsPath || currentPath === settingsPath + '/') {
            scrollToSection();
        } else {
            if (typeof htmx !== 'undefined') {
                htmx.ajax('GET', settingsPath, { target: '#main-content', swap: 'innerHTML show:window:top' });
                history.pushState({}, '', settingsPath);
                document.body.addEventListener('htmx:afterSettle', function onSwap() {
                    document.body.removeEventListener('htmx:afterSettle', onSwap);
                    scrollToSection();
                });
            } else {
                window.location.href = settingsPath + '#' + sectionId;
            }
        }
    }

    // --- Keyboard ---

    _onGlobalKeyDown(e) {
        // Cmd+K / Ctrl+K to toggle.
        if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
            e.preventDefault();
            this.toggle();
            return;
        }
        // Escape to close.
        if (e.key === 'Escape' && this._open) {
            e.preventDefault();
            e.stopPropagation();
            this.close();
        }
    }

    _onInputKeyDown(e) {
        const count = this._results.length;
        if (count === 0) return;

        if (e.key === 'ArrowDown') {
            e.preventDefault();
            this._activeIndex = (this._activeIndex + 1) % count;
            this._highlightActive();
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            this._activeIndex = (this._activeIndex - 1 + count) % count;
            this._highlightActive();
        } else if (e.key === 'Enter') {
            e.preventDefault();
            if (this._activeIndex >= 0 && this._activeIndex < count) {
                this._select(this._results[this._activeIndex]);
            }
        }
    }

    // --- Utilities ---

    _esc(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
}

if (!customElements.get('mcms-command-palette')) {
    customElements.define('mcms-command-palette', McmsCommandPalette);
}
