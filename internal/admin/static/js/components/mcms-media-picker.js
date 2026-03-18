/**
 * @module mcms-media-picker
 * @description Modal media browser and selector dialog (Light DOM).
 *
 * Uses `<mcms-media-grid mode="picker">` internally for browsing and selecting media.
 * Loads grid content from the server via fetch, supports folder navigation within the
 * picker, file upload, search, and single/multi-select modes.
 *
 * The picker builds its entire DOM imperatively in `connectedCallback` (no template or
 * shadow DOM). Visibility is controlled by the `open` attribute -- adding it opens the
 * dialog, removing it closes it.
 *
 * @example <caption>Single-select media picker</caption>
 * <mcms-media-picker accept="image/*"></mcms-media-picker>
 *
 * <script>
 *   var picker = document.querySelector('mcms-media-picker');
 *   picker.open();
 *   picker.addEventListener('media-selected', function(e) {
 *     console.log('Selected:', e.detail.id, e.detail.url);
 *   });
 * </script>
 *
 * @example <caption>Multi-select media picker</caption>
 * <mcms-media-picker multiple></mcms-media-picker>
 *
 * <script>
 *   var picker = document.querySelector('mcms-media-picker');
 *   picker.open();
 *   picker.addEventListener('media-selected', function(e) {
 *     e.detail.items.forEach(function(item) {
 *       console.log('Selected:', item.id, item.url, item.alt);
 *     });
 *   });
 * </script>
 */

/**
 * @event media-selected
 * @description Fired when the user confirms their media selection by clicking the "Select" button.
 *   The detail shape depends on whether the `multiple` attribute is set.
 * @type {CustomEvent}
 * @property {Object} detail - Selection data.
 * @property {string} [detail.id] - Media ID (single-select mode only).
 * @property {string} [detail.url] - Media URL (single-select mode only).
 * @property {string} [detail.alt] - Media alt text (single-select mode only).
 * @property {Array<Object>} [detail.items] - Array of selected items (multi-select mode only).
 * @property {string} [detail.items[].id] - Media ID of each selected item.
 * @property {string} [detail.items[].url] - Media URL of each selected item.
 * @property {string} [detail.items[].alt] - Media alt text of each selected item.
 */

/**
 * @event media-picker:cancel
 * @description Fired when the user cancels the picker by clicking the close button,
 *   the cancel button, the backdrop, or pressing the Escape key.
 * @type {CustomEvent}
 */

/**
 * Modal media browser and selector dialog component.
 *
 * Builds an accessible modal dialog with:
 * - A header with title, breadcrumb, upload toggle button, and close button
 * - A collapsible upload zone with drag-and-drop and click-to-browse file input
 * - A scrollable container holding a `<mcms-media-grid mode="picker">` element
 * - Footer actions with Cancel and Select buttons
 *
 * The grid is loaded from the server via fetch. Folder navigation dispatches events
 * from the grid that the picker intercepts to load new folder content. File selection
 * is handled entirely by the grid in picker mode.
 *
 * @extends HTMLElement
 *
 * @fires media-selected
 * @fires media-picker:cancel
 *
 * @attr {boolean} open - When present, the picker dialog is visible.
 * @attr {string} accept - File type filter for the upload input and server query.
 * @attr {boolean} multiple - Enables multi-select mode. Passed through to the grid.
 */
class McmsMediaPicker extends HTMLElement {
    constructor() {
        super();
        /** @type {boolean} Whether the DOM has been built */
        this._built = false;
        /** @type {HTMLElement|null} The full-screen backdrop overlay element */
        this._backdrop = null;
        /** @type {HTMLElement|null} The scrollable grid container */
        this._gridContainer = null;
        /** @type {HTMLElement|null} The collapsible upload zone wrapper */
        this._uploadZone = null;
        /** @type {HTMLElement|null} The upload status text element */
        this._uploadStatus = null;
        /** @type {HTMLElement|null} The drag-and-drop zone element */
        this._dropzone = null;
        /** @type {HTMLElement|null} The breadcrumb container */
        this._breadcrumb = null;
        /** @type {HTMLElement|null} The Select button */
        this._selectBtn = null;
        /** @type {Array<{id: string, url: string, alt: string, name: string}>} Currently selected items */
        this._selected = [];
        /** @type {string} Current folder ID being browsed */
        this._currentFolderId = '';
        /** @type {Array<{id: string, name: string}>} Folder navigation stack for breadcrumb */
        this._folderStack = [];
        /** @type {Function} Bound keydown handler for Escape key */
        this._boundKeyDown = this._onKeyDown.bind(this);
        /** @type {Function} Bound handler for grid pick events */
        this._boundOnPick = this._onGridPick.bind(this);
        /** @type {Function} Bound handler for grid navigate events */
        this._boundOnNavigate = this._onGridNavigate.bind(this);
    }

    /**
     * Declares which attributes trigger `attributeChangedCallback`.
     * @returns {string[]}
     */
    static get observedAttributes() {
        return ['open'];
    }

    /**
     * Lifecycle callback invoked when the element is inserted into the DOM.
     */
    connectedCallback() {
        this._build();
        this._syncVisibility();
    }

    /**
     * Lifecycle callback invoked when the element is removed from the DOM.
     */
    disconnectedCallback() {
        document.removeEventListener('keydown', this._boundKeyDown);
        this._built = false;
    }

    /**
     * Lifecycle callback invoked when an observed attribute changes.
     * @param {string} name
     */
    attributeChangedCallback(name) {
        if (!this._built) return;
        if (name === 'open') {
            this._syncVisibility();
        }
    }

    /**
     * Imperatively builds the dialog DOM tree. Only runs once.
     */
    _build() {
        if (this._built) return;
        this._built = true;

        var self = this;

        // Backdrop
        var backdrop = document.createElement('div');
        backdrop.className = 'fixed inset-0 z-[110] bg-black/60';
        backdrop.addEventListener('click', function(e) {
            if (e.target === backdrop) {
                self._cancel();
            }
        });

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
        uploadBtn.addEventListener('click', function() { self._toggleUpload(); });
        headerLeft.appendChild(uploadBtn);

        header.appendChild(headerLeft);

        var closeBtn = document.createElement('button');
        closeBtn.type = 'button';
        closeBtn.className = 'text-[var(--color-text-muted)] hover:text-[var(--color-text)] cursor-pointer bg-transparent border-none text-lg';
        closeBtn.textContent = '\u00d7';
        closeBtn.setAttribute('aria-label', 'Close');
        closeBtn.addEventListener('click', function() { self._cancel(); });
        header.appendChild(closeBtn);

        panel.appendChild(header);

        // Breadcrumb bar
        var breadcrumb = document.createElement('div');
        breadcrumb.className = 'flex items-center gap-1 border-b border-[var(--color-border)] px-5 py-2 text-sm';
        breadcrumb.innerHTML = '<button type="button" class="text-[var(--color-text-muted)] hover:text-[var(--color-text)] cursor-pointer bg-transparent border-none text-sm" data-breadcrumb-root>All Media</button>';
        this._breadcrumb = breadcrumb;
        panel.appendChild(breadcrumb);

        // Breadcrumb root click
        breadcrumb.addEventListener('click', function(e) {
            var root = e.target.closest('[data-breadcrumb-root]');
            if (root) {
                self._currentFolderId = '';
                self._folderStack = [];
                self._loadGrid('');
                self._updateBreadcrumb();
                return;
            }
            var crumb = e.target.closest('[data-breadcrumb-folder]');
            if (crumb) {
                var targetId = crumb.getAttribute('data-breadcrumb-folder');
                // Trim the folder stack to this point
                var idx = -1;
                for (var i = 0; i < self._folderStack.length; i++) {
                    if (self._folderStack[i].id === targetId) { idx = i; break; }
                }
                if (idx >= 0) {
                    self._folderStack = self._folderStack.slice(0, idx + 1);
                }
                self._currentFolderId = targetId;
                self._loadGrid(targetId);
                self._updateBreadcrumb();
            }
        });

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

        dropzone.addEventListener('click', function() { fileInput.click(); });
        fileInput.addEventListener('change', function() {
            if (fileInput.files && fileInput.files[0]) {
                self._uploadFile(fileInput.files[0]);
                fileInput.value = '';
            }
        });
        dropzone.addEventListener('dragover', function(e) {
            e.preventDefault();
            dropzone.setAttribute('data-dragover', '');
        });
        dropzone.addEventListener('dragleave', function() {
            dropzone.removeAttribute('data-dragover');
        });
        dropzone.addEventListener('drop', function(e) {
            e.preventDefault();
            dropzone.removeAttribute('data-dragover');
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

        // Grid container (scrollable, holds the mcms-media-grid)
        var gridContainer = document.createElement('div');
        gridContainer.className = 'flex-1 overflow-y-auto p-4';
        panel.appendChild(gridContainer);
        this._gridContainer = gridContainer;

        // Footer actions
        var actions = document.createElement('div');
        actions.className = 'flex justify-end gap-3 border-t border-[var(--color-border)] px-5 py-4';

        var cancelBtn = document.createElement('button');
        cancelBtn.type = 'button';
        cancelBtn.className = 'inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] transition-colors cursor-pointer border-none';
        cancelBtn.textContent = 'Cancel';
        cancelBtn.addEventListener('click', function() { self._cancel(); });
        actions.appendChild(cancelBtn);

        var selectBtn = document.createElement('button');
        selectBtn.type = 'button';
        selectBtn.className = 'inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors cursor-pointer border-none disabled:opacity-50 disabled:cursor-not-allowed';
        selectBtn.textContent = 'Select';
        selectBtn.disabled = true;
        selectBtn.addEventListener('click', function() { self._confirmSelection(); });
        actions.appendChild(selectBtn);
        this._selectBtn = selectBtn;

        panel.appendChild(actions);
        backdrop.appendChild(panel);

        this._backdrop = backdrop;
        this.appendChild(backdrop);
    }

    /**
     * Synchronizes the dialog visibility with the `open` attribute state.
     */
    _syncVisibility() {
        var isOpen = this.hasAttribute('open');
        if (this._backdrop) {
            this._backdrop.style.display = isOpen ? '' : 'none';
        }
        if (isOpen) {
            document.addEventListener('keydown', this._boundKeyDown);
            this._selected = [];
            this._currentFolderId = '';
            this._folderStack = [];
            this._updateSelectBtn();
            // Collapse upload zone
            if (this._uploadZone) {
                this._uploadZone.hidden = true;
            }
            // Load root grid
            this._loadGrid('');
            this._updateBreadcrumb();
        } else {
            document.removeEventListener('keydown', this._boundKeyDown);
            // Clean up grid event listeners
            this._unbindGridEvents();
        }
    }

    /**
     * Loads the media grid for the given folder from the server.
     * Fetches HTML containing `<mcms-media-grid mode="picker">` and swaps it
     * into the grid container.
     *
     * @param {string} folderId - Folder ID to load, empty string for root.
     */
    _loadGrid(folderId) {
        if (!this._gridContainer) return;
        var self = this;

        var url = '/admin/media?picker=true';
        if (folderId) {
            url += '&folder_id=' + encodeURIComponent(folderId);
        }
        var accept = this.getAttribute('accept');
        if (accept) {
            url += '&accept=' + encodeURIComponent(accept);
        }

        // Show loading state
        this._gridContainer.innerHTML = '<div class="flex items-center justify-center py-12"><p class="text-sm text-[var(--color-text-muted)]">Loading media...</p></div>';

        fetch(url, {
            headers: { 'HX-Request': 'true' },
            credentials: 'same-origin'
        }).then(function(res) {
            if (!res.ok) {
                throw new Error('Failed to load media (' + res.status + ')');
            }
            return res.text();
        }).then(function(html) {
            self._unbindGridEvents();
            self._gridContainer.innerHTML = html;

            // Find the grid element and configure it
            var grid = self._gridContainer.querySelector('mcms-media-grid');
            if (grid) {
                // Pass through the multiple attribute
                if (self.hasAttribute('multiple')) {
                    grid.setAttribute('multiple', '');
                }
                // Bind events
                self._bindGridEvents();
            }

            // Reset selection state for the new grid
            self._selected = [];
            self._updateSelectBtn();
        }).catch(function(err) {
            console.error('[mcms-media-picker] load error:', err);
            self._gridContainer.innerHTML = '<div class="flex items-center justify-center py-12"><p class="text-sm text-red-400">Failed to load media</p></div>';
        });
    }

    /**
     * Binds event listeners on the grid element inside the container.
     */
    _bindGridEvents() {
        if (!this._gridContainer) return;
        this._gridContainer.addEventListener('mcms-media-grid:pick', this._boundOnPick);
        this._gridContainer.addEventListener('mcms-media-grid:navigate', this._boundOnNavigate);
    }

    /**
     * Removes grid event listeners.
     */
    _unbindGridEvents() {
        if (!this._gridContainer) return;
        this._gridContainer.removeEventListener('mcms-media-grid:pick', this._boundOnPick);
        this._gridContainer.removeEventListener('mcms-media-grid:navigate', this._boundOnNavigate);
    }

    /**
     * Handles the `mcms-media-grid:pick` event to track selection.
     * @param {CustomEvent} e
     */
    _onGridPick(e) {
        this._selected = e.detail.selected || [];
        this._updateSelectBtn();
    }

    /**
     * Handles the `mcms-media-grid:navigate` event for folder navigation.
     * @param {CustomEvent} e
     */
    _onGridNavigate(e) {
        var folderId = e.detail.folderId || '';
        if (!folderId) return;

        // Try to get the folder name from the grid item
        var folderName = folderId;
        var grid = this._gridContainer ? this._gridContainer.querySelector('mcms-media-grid') : null;
        if (grid) {
            var folderEl = grid.querySelector('[data-media-folder-item][data-folder-id="' + folderId + '"]');
            if (folderEl) {
                var nameEl = folderEl.querySelector('.text-white');
                if (nameEl) {
                    folderName = nameEl.textContent || folderId;
                }
            }
        }

        this._folderStack.push({ id: folderId, name: folderName });
        this._currentFolderId = folderId;
        this._loadGrid(folderId);
        this._updateBreadcrumb();
    }

    /**
     * Updates the breadcrumb display to reflect the current folder path.
     */
    _updateBreadcrumb() {
        if (!this._breadcrumb) return;
        var html = '<button type="button" class="text-[var(--color-text-muted)] hover:text-[var(--color-text)] cursor-pointer bg-transparent border-none text-sm" data-breadcrumb-root>All Media</button>';

        for (var i = 0; i < this._folderStack.length; i++) {
            var f = this._folderStack[i];
            var isLast = i === this._folderStack.length - 1;
            html += '<span class="text-[var(--color-text-muted)]">/</span>';
            if (isLast) {
                html += '<span class="text-sm font-medium text-[var(--color-text)]">' + this._escHtml(f.name) + '</span>';
            } else {
                html += '<button type="button" class="text-[var(--color-text-muted)] hover:text-[var(--color-text)] cursor-pointer bg-transparent border-none text-sm" data-breadcrumb-folder="' + this._escHtml(f.id) + '">' + this._escHtml(f.name) + '</button>';
            }
        }

        this._breadcrumb.innerHTML = html;
    }

    /**
     * Updates the Select button enabled/disabled state.
     */
    _updateSelectBtn() {
        if (this._selectBtn) {
            this._selectBtn.disabled = this._selected.length === 0;
        }
    }

    /**
     * Toggles the upload zone visibility.
     */
    _toggleUpload() {
        if (!this._uploadZone) return;
        this._uploadZone.hidden = !this._uploadZone.hidden;
        if (!this._uploadZone.hidden) {
            this._uploadStatus.textContent = '';
            this._uploadStatus.className = 'mt-2 text-sm text-[var(--color-text-muted)]';
        }
    }

    /**
     * Uploads a file to the server and refreshes the grid on success.
     * @param {File} file
     */
    _uploadFile(file) {
        if (!file) return;
        var self = this;
        var formData = new FormData();
        formData.append('file', file);
        if (this._currentFolderId) {
            formData.append('folder_id', this._currentFolderId);
        }

        var csrfMeta = document.querySelector('meta[name="csrf-token"]');
        var csrfToken = csrfMeta ? csrfMeta.content : '';

        this._dropzone.setAttribute('data-uploading', '');
        this._uploadStatus.textContent = 'Uploading...';
        this._uploadStatus.className = 'mt-2 text-sm text-[var(--color-text-muted)]';

        fetch('/admin/media', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': csrfToken,
                'HX-Request': 'true'
            },
            body: formData,
            credentials: 'same-origin'
        }).then(function(response) {
            if (!response.ok) {
                throw new Error('Upload failed (' + response.status + ')');
            }
            self._dropzone.removeAttribute('data-uploading');
            self._uploadStatus.textContent = 'Upload complete';
            self._uploadZone.hidden = true;
            // Reload grid to show the new item
            self._loadGrid(self._currentFolderId);
            return response.text();
        }).catch(function(err) {
            self._dropzone.removeAttribute('data-uploading');
            self._uploadStatus.textContent = err.message || 'Upload failed';
            self._uploadStatus.className = 'mt-2 text-sm text-[var(--color-danger)]';
        });
    }

    /**
     * Confirms the current selection and closes the picker.
     */
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

    /**
     * Cancels the picker.
     */
    _cancel() {
        this.dispatchEvent(new CustomEvent('media-picker:cancel', {
            bubbles: true
        }));
        this.removeAttribute('open');
    }

    /**
     * Handles Escape key.
     * @param {KeyboardEvent} e
     */
    _onKeyDown(e) {
        if (!this.hasAttribute('open')) return;
        if (e.key === 'Escape') {
            e.preventDefault();
            this._cancel();
        }
    }

    /**
     * Escapes a string for safe HTML insertion.
     * @param {string} s
     * @returns {string}
     */
    _escHtml(s) {
        if (!s) return '';
        var d = document.createElement('div');
        d.appendChild(document.createTextNode(s));
        return d.innerHTML;
    }

    // ---- Public API ----

    /**
     * Opens the media picker dialog.
     */
    open() {
        this.setAttribute('open', '');
    }

    /**
     * Closes the media picker dialog programmatically (no cancel event).
     */
    close() {
        this.removeAttribute('open');
    }
}

customElements.define('mcms-media-picker', McmsMediaPicker);
