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
        this._uploadZone = null;
        this._uploadStatus = null;
        this._dropzone = null;
        this._selected = [];
        this._pendingSelectId = null;
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
        backdrop.className = 'fixed inset-0 z-[110] bg-black/60';
        backdrop.addEventListener('click', function(e) {
            if (e.target === backdrop) {
                this._cancel();
            }
        }.bind(this));

        // Panel
        var panel = document.createElement('div');
        panel.className = 'fixed inset-4 z-[110] flex flex-col rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] shadow-lg';
        panel.setAttribute('role', 'dialog');
        panel.setAttribute('aria-modal', 'true');
        panel.setAttribute('aria-label', 'Media picker');

        // Header
        var header = document.createElement('div');
        header.className = 'flex items-center justify-between border-b border-[var(--color-border)] px-5 py-4';

        var headerLeft = document.createElement('div');
        headerLeft.className = 'flex items-center gap-3';

        var title = document.createElement('h2');
        title.className = 'text-lg font-semibold text-[var(--color-text)]';
        title.textContent = 'Select Media';
        headerLeft.appendChild(title);

        var uploadBtn = document.createElement('button');
        uploadBtn.type = 'button';
        uploadBtn.className = 'inline-flex items-center justify-center rounded-md px-3 py-1.5 text-xs font-medium bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors cursor-pointer border-none';
        uploadBtn.textContent = 'Upload';
        uploadBtn.addEventListener('click', this._toggleUpload.bind(this));
        headerLeft.appendChild(uploadBtn);

        header.appendChild(headerLeft);

        var closeBtn = document.createElement('button');
        closeBtn.type = 'button';
        closeBtn.className = 'text-[var(--color-text-muted)] hover:text-[var(--color-text)] cursor-pointer bg-transparent border-none text-lg';
        closeBtn.textContent = '\u00d7';
        closeBtn.setAttribute('aria-label', 'Close');
        closeBtn.addEventListener('click', this._cancel.bind(this));
        header.appendChild(closeBtn);

        panel.appendChild(header);

        // Search
        var searchWrapper = document.createElement('div');
        searchWrapper.className = 'px-5 py-2';

        var searchInput = document.createElement('input');
        searchInput.type = 'search';
        searchInput.placeholder = 'Search media...';
        searchInput.className = 'w-full rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]';
        searchWrapper.appendChild(searchInput);
        panel.appendChild(searchWrapper);

        this._searchInput = searchInput;

        // Upload zone (hidden by default)
        var uploadZone = document.createElement('div');
        uploadZone.className = 'px-5 py-2';
        uploadZone.hidden = true;

        var dropzone = document.createElement('div');
        dropzone.className = 'flex flex-col items-center justify-center rounded-lg border-2 border-dashed border-[var(--color-border)] bg-[var(--color-bg)] p-8 text-center text-sm text-[var(--color-text-muted)] transition-colors cursor-pointer';
        dropzone.textContent = 'Drop a file here or click to browse';

        var fileInput = document.createElement('input');
        fileInput.type = 'file';
        fileInput.style.display = 'none';
        var accept = this.getAttribute('accept');
        if (accept) {
            fileInput.accept = accept;
        }

        var self = this;
        dropzone.addEventListener('click', function() {
            fileInput.click();
        });
        fileInput.addEventListener('change', function() {
            if (fileInput.files && fileInput.files[0]) {
                self._uploadFile(fileInput.files[0]);
                fileInput.value = '';
            }
        });
        dropzone.addEventListener('dragover', function(e) {
            e.preventDefault();
            dropzone.setAttribute('data-dragover', '');
            dropzone.classList.add('dragover'); // DUAL: data-dragover + class
        });
        dropzone.addEventListener('dragleave', function() {
            dropzone.removeAttribute('data-dragover');
            dropzone.classList.remove('dragover'); // DUAL: data-dragover + class
        });
        dropzone.addEventListener('drop', function(e) {
            e.preventDefault();
            dropzone.removeAttribute('data-dragover');
            dropzone.classList.remove('dragover'); // DUAL: data-dragover + class
            if (e.dataTransfer.files && e.dataTransfer.files[0]) {
                self._uploadFile(e.dataTransfer.files[0]);
            }
        });

        dropzone.appendChild(fileInput);
        uploadZone.appendChild(dropzone);

        var uploadStatus = document.createElement('div');
        uploadStatus.className = 'mt-2 text-sm text-[var(--color-text-muted)]';
        uploadZone.appendChild(uploadStatus);

        panel.appendChild(uploadZone);
        this._uploadZone = uploadZone;
        this._uploadStatus = uploadStatus;
        this._dropzone = dropzone;

        // Grid container
        var grid = document.createElement('div');
        grid.className = 'grid grid-cols-[repeat(auto-fill,minmax(8rem,1fr))] gap-3 overflow-y-auto p-4 flex-1';
        grid.id = 'media-picker-grid-' + this._uniqueId();
        panel.appendChild(grid);
        this._grid = grid;

        // Actions
        var actions = document.createElement('div');
        actions.className = 'flex justify-end gap-3 border-t border-[var(--color-border)] px-5 py-4';

        var cancelBtn = document.createElement('button');
        cancelBtn.type = 'button';
        cancelBtn.className = 'inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] transition-colors cursor-pointer border-none';
        cancelBtn.textContent = 'Cancel';
        cancelBtn.addEventListener('click', this._cancel.bind(this));
        actions.appendChild(cancelBtn);

        var confirmBtn = document.createElement('button');
        confirmBtn.type = 'button';
        confirmBtn.className = 'inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors cursor-pointer border-none';
        confirmBtn.textContent = 'Select';
        confirmBtn.addEventListener('click', this._confirmSelection.bind(this));
        actions.appendChild(confirmBtn);

        panel.appendChild(actions);
        backdrop.appendChild(panel);

        this._backdrop = backdrop;
        this.appendChild(backdrop);

        // Set up search debounce
        var debounceTimer = null;
        searchInput.addEventListener('input', function() {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(function() {
                self._loadMedia(searchInput.value.trim());
            }, 300);
        });

        // Delegate clicks on grid items
        grid.addEventListener('click', function(e) {
            var item = e.target.closest('[data-media-picker-item]');
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
            this._pendingSelectId = null;
            // Collapse upload zone
            if (this._uploadZone) {
                this._uploadZone.hidden = true;
            }
            // Clear previous selections visually
            if (this._grid) {
                var items = this._grid.querySelectorAll('[data-media-picker-item][data-selected]');
                for (var i = 0; i < items.length; i++) {
                    items[i].removeAttribute('data-selected');
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
            var allItems = this._grid.querySelectorAll('[data-media-picker-item][data-selected]');
            for (var i = 0; i < allItems.length; i++) {
                allItems[i].removeAttribute('data-selected');
                allItems[i].classList.remove('selected'); // DUAL: data-selected + class
            }
            this._selected = [];
        }

        if (item.hasAttribute('data-selected')) {
            item.removeAttribute('data-selected');
            item.classList.remove('selected'); // DUAL: data-selected + class
            this._selected = this._selected.filter(function(s) { return s.id !== mediaId; });
        } else {
            item.setAttribute('data-selected', '');
            item.classList.add('selected'); // DUAL: data-selected + class
            this._selected.push({ id: mediaId, url: mediaUrl, alt: mediaAlt });
        }
    }

    _toggleUpload() {
        if (!this._uploadZone) return;
        this._uploadZone.hidden = !this._uploadZone.hidden;
        if (!this._uploadZone.hidden) {
            this._uploadStatus.textContent = '';
            this._uploadStatus.className = 'mt-2 text-sm text-[var(--color-text-muted)]';
        }
    }

    _uploadFile(file) {
        if (!file) return;
        var self = this;
        var formData = new FormData();
        formData.append('file', file);

        var csrfMeta = document.querySelector('meta[name="csrf-token"]');
        var csrfToken = csrfMeta ? csrfMeta.content : '';

        this._dropzone.setAttribute('data-uploading', '');
        this._dropzone.classList.add('uploading'); // DUAL: data-uploading + class
        this._uploadStatus.textContent = 'Uploading...';
        this._uploadStatus.className = 'mt-2 text-sm text-[var(--color-text-muted)]';

        fetch('/admin/media', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': csrfToken,
                'HX-Request': 'true'
            },
            body: formData
        }).then(function(response) {
            if (!response.ok) {
                throw new Error('Upload failed (' + response.status + ')');
            }
            var mediaId = response.headers.get('X-Media-ID');
            var mediaUrl = response.headers.get('X-Media-URL');
            self._pendingSelectId = mediaId;
            self._dropzone.removeAttribute('data-uploading');
            self._dropzone.classList.remove('uploading'); // DUAL: data-uploading + class
            self._uploadStatus.textContent = 'Upload complete';
            self._uploadZone.hidden = true;
            self._loadMedia('');

            // After grid refreshes, auto-select the new item
            if (mediaId && self._grid) {
                var onSwap = function() {
                    self._grid.removeEventListener('htmx:afterSwap', onSwap);
                    var newItem = self._grid.querySelector('[data-media-picker-item][data-media-id="' + mediaId + '"]');
                    if (newItem) {
                        self._toggleItem(newItem);
                    }
                    self._pendingSelectId = null;
                };
                self._grid.addEventListener('htmx:afterSwap', onSwap);
            }
            return response.text();
        }).catch(function(err) {
            self._dropzone.removeAttribute('data-uploading');
            self._dropzone.classList.remove('uploading'); // DUAL: data-uploading + class
            self._uploadStatus.textContent = err.message || 'Upload failed';
            self._uploadStatus.className = 'mt-2 text-sm text-[var(--color-danger)]';
        });
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
