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

        var headerLeft = document.createElement('div');
        headerLeft.className = 'media-picker-header-left';

        var title = document.createElement('h2');
        title.className = 'media-picker-title';
        title.textContent = 'Select Media';
        headerLeft.appendChild(title);

        var uploadBtn = document.createElement('button');
        uploadBtn.type = 'button';
        uploadBtn.className = 'btn btn-sm btn-primary';
        uploadBtn.textContent = 'Upload';
        uploadBtn.addEventListener('click', this._toggleUpload.bind(this));
        headerLeft.appendChild(uploadBtn);

        header.appendChild(headerLeft);

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

        // Upload zone (hidden by default)
        var uploadZone = document.createElement('div');
        uploadZone.className = 'media-picker-upload';
        uploadZone.hidden = true;

        var dropzone = document.createElement('div');
        dropzone.className = 'media-picker-dropzone';
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
            dropzone.classList.add('dragover');
        });
        dropzone.addEventListener('dragleave', function() {
            dropzone.classList.remove('dragover');
        });
        dropzone.addEventListener('drop', function(e) {
            e.preventDefault();
            dropzone.classList.remove('dragover');
            if (e.dataTransfer.files && e.dataTransfer.files[0]) {
                self._uploadFile(e.dataTransfer.files[0]);
            }
        });

        dropzone.appendChild(fileInput);
        uploadZone.appendChild(dropzone);

        var uploadStatus = document.createElement('div');
        uploadStatus.className = 'media-picker-upload-status';
        uploadZone.appendChild(uploadStatus);

        panel.appendChild(uploadZone);
        this._uploadZone = uploadZone;
        this._uploadStatus = uploadStatus;
        this._dropzone = dropzone;

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
            this._pendingSelectId = null;
            // Collapse upload zone
            if (this._uploadZone) {
                this._uploadZone.hidden = true;
            }
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

    _toggleUpload() {
        if (!this._uploadZone) return;
        this._uploadZone.hidden = !this._uploadZone.hidden;
        if (!this._uploadZone.hidden) {
            this._uploadStatus.textContent = '';
            this._uploadStatus.className = 'media-picker-upload-status';
        }
    }

    _uploadFile(file) {
        if (!file) return;
        var self = this;
        var formData = new FormData();
        formData.append('file', file);

        var csrfMeta = document.querySelector('meta[name="csrf-token"]');
        var csrfToken = csrfMeta ? csrfMeta.content : '';

        this._dropzone.classList.add('uploading');
        this._uploadStatus.textContent = 'Uploading...';
        this._uploadStatus.className = 'media-picker-upload-status';

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
            self._dropzone.classList.remove('uploading');
            self._uploadStatus.textContent = 'Upload complete';
            self._uploadZone.hidden = true;
            self._loadMedia('');

            // After grid refreshes, auto-select the new item
            if (mediaId && self._grid) {
                var onSwap = function() {
                    self._grid.removeEventListener('htmx:afterSwap', onSwap);
                    var newItem = self._grid.querySelector('.media-picker-item[data-media-id="' + mediaId + '"]');
                    if (newItem) {
                        self._toggleItem(newItem);
                    }
                    self._pendingSelectId = null;
                };
                self._grid.addEventListener('htmx:afterSwap', onSwap);
            }
            return response.text();
        }).catch(function(err) {
            self._dropzone.classList.remove('uploading');
            self._uploadStatus.textContent = err.message || 'Upload failed';
            self._uploadStatus.className = 'media-picker-upload-status media-picker-upload-error';
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
