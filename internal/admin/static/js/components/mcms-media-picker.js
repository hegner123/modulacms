// <mcms-media-picker> -- Media browser/selector dialog (Light DOM)
// Supports: open, accept, multiple attributes
// Dispatches: media-selected, media-picker:cancel custom events
class McmsMediaPicker extends HTMLElement {
    constructor() {
        super();
        this._built = false;
        this._backdrop = null;
        this._searchInput = null;
        this._grid = null;
        this._selected = [];
        this._boundKeyDown = this._onKeyDown.bind(this);
    }

    static get observedAttributes() {
        return ['open'];
    }

    connectedCallback() {
        this._build();
        this._syncVisibility();
    }

    disconnectedCallback() {
        document.removeEventListener('keydown', this._boundKeyDown);
        this._built = false;
    }

    attributeChangedCallback(name) {
        if (!this._built) return;
        if (name === 'open') {
            this._syncVisibility();
        }
    }

    _build() {
        if (this._built) return;
        this._built = true;

        // Backdrop
        var backdrop = document.createElement('div');
        backdrop.className = 'media-picker-backdrop';
        backdrop.addEventListener('click', function(e) {
            if (e.target === backdrop) {
                this._cancel();
            }
        }.bind(this));

        // Panel
        var panel = document.createElement('div');
        panel.className = 'media-picker-panel';
        panel.setAttribute('role', 'dialog');
        panel.setAttribute('aria-modal', 'true');
        panel.setAttribute('aria-label', 'Media picker');

        // Header
        var header = document.createElement('div');
        header.className = 'media-picker-header';

        var title = document.createElement('h2');
        title.className = 'media-picker-title';
        title.textContent = 'Select Media';
        header.appendChild(title);

        var closeBtn = document.createElement('button');
        closeBtn.type = 'button';
        closeBtn.className = 'btn btn-ghost btn-sm media-picker-close';
        closeBtn.textContent = '\u00d7';
        closeBtn.setAttribute('aria-label', 'Close');
        closeBtn.addEventListener('click', this._cancel.bind(this));
        header.appendChild(closeBtn);

        panel.appendChild(header);

        // Search
        var searchWrapper = document.createElement('div');
        searchWrapper.className = 'media-picker-search';

        var searchInput = document.createElement('input');
        searchInput.type = 'search';
        searchInput.placeholder = 'Search media...';
        searchInput.className = 'media-picker-search-input';
        searchWrapper.appendChild(searchInput);
        panel.appendChild(searchWrapper);

        this._searchInput = searchInput;

        // Grid container
        var grid = document.createElement('div');
        grid.className = 'media-picker-grid';
        grid.id = 'media-picker-grid-' + this._uniqueId();
        panel.appendChild(grid);
        this._grid = grid;

        // Actions
        var actions = document.createElement('div');
        actions.className = 'media-picker-actions';

        var cancelBtn = document.createElement('button');
        cancelBtn.type = 'button';
        cancelBtn.className = 'btn';
        cancelBtn.textContent = 'Cancel';
        cancelBtn.addEventListener('click', this._cancel.bind(this));
        actions.appendChild(cancelBtn);

        var confirmBtn = document.createElement('button');
        confirmBtn.type = 'button';
        confirmBtn.className = 'btn btn-primary';
        confirmBtn.textContent = 'Select';
        confirmBtn.addEventListener('click', this._confirmSelection.bind(this));
        actions.appendChild(confirmBtn);

        panel.appendChild(actions);
        backdrop.appendChild(panel);

        this._backdrop = backdrop;
        this.appendChild(backdrop);

        // Set up search debounce
        var self = this;
        var debounceTimer = null;
        searchInput.addEventListener('input', function() {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(function() {
                self._loadMedia(searchInput.value.trim());
            }, 300);
        });

        // Delegate clicks on grid items
        grid.addEventListener('click', function(e) {
            var item = e.target.closest('.media-picker-item');
            if (!item) return;
            self._toggleItem(item);
        });
    }

    _syncVisibility() {
        var isOpen = this.hasAttribute('open');
        if (this._backdrop) {
            this._backdrop.style.display = isOpen ? '' : 'none';
        }
        if (isOpen) {
            document.addEventListener('keydown', this._boundKeyDown);
            this._selected = [];
            // Clear previous selections visually
            if (this._grid) {
                var items = this._grid.querySelectorAll('.media-picker-item.selected');
                for (var i = 0; i < items.length; i++) {
                    items[i].classList.remove('selected');
                }
            }
            // Load initial media
            this._loadMedia('');
            // Focus search input
            var self = this;
            requestAnimationFrame(function() {
                if (self._searchInput) {
                    self._searchInput.value = '';
                    self._searchInput.focus();
                }
            });
        } else {
            document.removeEventListener('keydown', this._boundKeyDown);
        }
    }

    _loadMedia(query) {
        if (!this._grid) return;

        var url = '/admin/media?picker=true';
        if (query) {
            url += '&q=' + encodeURIComponent(query);
        }
        var accept = this.getAttribute('accept');
        if (accept) {
            url += '&accept=' + encodeURIComponent(accept);
        }

        if (typeof htmx !== 'undefined') {
            htmx.ajax('GET', url, {
                target: this._grid,
                swap: 'innerHTML'
            });
        }
    }

    _toggleItem(item) {
        var isMultiple = this.hasAttribute('multiple');
        var mediaId = item.getAttribute('data-media-id') || '';
        var mediaUrl = item.getAttribute('data-media-url') || '';
        var mediaAlt = item.getAttribute('data-media-alt') || '';

        if (!isMultiple) {
            // Single select: deselect all others
            var allItems = this._grid.querySelectorAll('.media-picker-item.selected');
            for (var i = 0; i < allItems.length; i++) {
                allItems[i].classList.remove('selected');
            }
            this._selected = [];
        }

        if (item.classList.contains('selected')) {
            item.classList.remove('selected');
            this._selected = this._selected.filter(function(s) { return s.id !== mediaId; });
        } else {
            item.classList.add('selected');
            this._selected.push({ id: mediaId, url: mediaUrl, alt: mediaAlt });
        }
    }

    _confirmSelection() {
        if (this._selected.length === 0) {
            this._cancel();
            return;
        }

        var isMultiple = this.hasAttribute('multiple');
        var detail;

        if (isMultiple) {
            detail = { items: this._selected.slice() };
        } else {
            detail = this._selected[0];
        }

        this.dispatchEvent(new CustomEvent('media-selected', {
            bubbles: true,
            detail: detail
        }));
        this.removeAttribute('open');
    }

    _cancel() {
        this.dispatchEvent(new CustomEvent('media-picker:cancel', {
            bubbles: true
        }));
        this.removeAttribute('open');
    }

    _onKeyDown(e) {
        if (!this.hasAttribute('open')) return;
        if (e.key === 'Escape') {
            e.preventDefault();
            this._cancel();
        }
    }

    _uniqueId() {
        return 'mp-' + Math.random().toString(36).substring(2, 9);
    }

    // Public API
    open() {
        this.setAttribute('open', '');
    }

    close() {
        this.removeAttribute('open');
    }
}

customElements.define('mcms-media-picker', McmsMediaPicker);
