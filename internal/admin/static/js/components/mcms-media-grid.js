/**
 * @module mcms-media-grid
 * @description Self-rendering media grid with folder drag-and-drop support (Light DOM).
 *
 * Reads media items from the `data-items` attribute (JSON array) and renders an
 * interactive grid of files and folders. Supports drag-to-folder moves, desktop
 * file upload via drag-and-drop, multi-select (click, Ctrl+click, Shift+click),
 * bulk delete, configurable grid density, and upload skeleton placeholders.
 *
 * The component re-initializes HTMX on rendered content and triggers lucide icon
 * creation after HTMX swaps into `#main-content`.
 *
 * @example <caption>Basic usage with server-rendered data</caption>
 * <mcms-media-grid
 *   data-items='[{"id":"01J...","type":"file","name":"photo.jpg","displayName":"Photo","mimetype":"image/jpeg","url":"/media/photo.jpg","alt":"A photo"}]'
 *   folder-id=""
 *   upload-url="/admin/media"
 *   detail-url="/admin/media/">
 *   <div id="pagination">...</div>
 * </mcms-media-grid>
 *
 * @example <caption>Listening for events</caption>
 * document.querySelector('mcms-media-grid')
 *   .addEventListener('mcms-media-grid:upload', function(e) {
 *     console.log('Uploaded:', e.detail.filename, e.detail.mediaId);
 *   });
 */

/**
 * @event mcms-media-grid:upload
 * @description Fired after a file is successfully uploaded via drag-and-drop.
 * @type {CustomEvent}
 * @property {Object} detail
 * @property {string} detail.mediaId - The server-assigned media ID (from `X-Media-ID` header).
 * @property {string} detail.url - The media URL (from `X-Media-URL` header).
 * @property {string} detail.filename - The original filename of the uploaded file.
 */

/**
 * @event mcms-media-grid:move
 * @description Fired after a media item is successfully moved to a folder via drag-and-drop.
 * @type {CustomEvent}
 * @property {Object} detail
 * @property {string} detail.mediaId - The ID of the media item that was moved.
 * @property {string} detail.folderId - The ID of the destination folder (empty string for root).
 */

/**
 * @event mcms-media-grid:select
 * @description Fired when the selection changes (click, Ctrl+click, or Shift+click on grid items).
 * @type {CustomEvent}
 * @property {Object} detail
 * @property {string[]} detail.selected - Array of currently selected media IDs.
 */

/**
 * @event mcms-media-grid:pick
 * @description Fired in picker mode when file selection changes. Contains full item metadata.
 * @type {CustomEvent}
 * @property {Object} detail
 * @property {Array<{id: string, url: string, alt: string, name: string}>} detail.selected - Selected items.
 */

/**
 * @event mcms-media-grid:navigate
 * @description Fired in picker mode when a folder is clicked (instead of HTMX navigation).
 * @type {CustomEvent}
 * @property {Object} detail
 * @property {string} detail.folderId - The folder ID to navigate into.
 * @property {string} detail.folderUrl - The URL for the folder content.
 */

/**
 * Media grid component that renders files and folders with drag-and-drop interactions.
 *
 * The grid reads its data from the `data-items` attribute, which must contain a JSON
 * array of item objects. Each item has a `type` field (`"file"` or `"folder"`) and
 * associated metadata (id, name, displayName, mimetype, url, alt).
 *
 * Grid density is persisted via `mcms.settings` under the key `media.gridSize` with
 * four levels (1-4) mapping to different responsive column counts.
 *
 * @extends HTMLElement
 *
 * @fires mcms-media-grid:upload
 * @fires mcms-media-grid:move
 * @fires mcms-media-grid:select
 * @fires mcms-media-grid:pick
 * @fires mcms-media-grid:navigate
 *
 * @attr {string} data-items - JSON array of media items to render. Each item must have
 *   `id`, `type` ("file"|"folder"), `name`, `displayName`. File items also use `mimetype`,
 *   `url`, and `alt`. Folder items use `id` and `displayName`.
 * @attr {string} folder-id - Current folder ID. Empty string or absent means root folder.
 *   Used when uploading files to associate them with the current folder.
 * @attr {string} upload-url - URL endpoint for file uploads. Defaults to `/admin/media`.
 * @attr {string} detail-url - URL prefix for media detail page links. Defaults to `/admin/media/`.
 *   The media ID is appended to form the full URL.
 * @attr {string} mode - Grid interaction mode. `"picker"` changes click behavior: clicking a file
 *   toggles selection instead of navigating to detail; clicking a folder dispatches a navigate event
 *   instead of HTMX navigation. Absent or `"default"` preserves the standard behavior.
 * @attr {boolean} multiple - In picker mode, allows selecting multiple files. Without this attribute,
 *   only one file can be selected at a time.
 *
 * @example
 * <mcms-media-grid
 *   data-items='[...]'
 *   folder-id="01HX..."
 *   upload-url="/admin/media"
 *   detail-url="/admin/media/">
 * </mcms-media-grid>
 */
class McmsMediaGrid extends HTMLElement {
    constructor() {
        super();
        /** @type {Array<Object>} Parsed media items from the data-items attribute */
        this._items = [];
        /** @type {Set<string>} Set of currently selected media IDs */
        this._selected = new Set();
        /** @type {string|null} Media ID of the item currently being dragged */
        this._dragId = null;
    }

    /**
     * Lifecycle callback invoked when the element is inserted into the DOM.
     * Sets positioning styles, reads item data, renders the drop overlay and grid,
     * and binds all event listeners for drag-and-drop, click selection, and HTMX swaps.
     */
    connectedCallback() {
        this.style.position = 'relative';
        this.style.display = 'block';
        this._readData();
        this._renderDropOverlay();
        this._render();
        this._bind();
        this._bindFolders();
    }

    /**
     * Lifecycle callback invoked when the element is removed from the DOM.
     * Removes all event listeners bound during connectedCallback.
     */
    disconnectedCallback() {
        this._unbind();
        this._unbindFolders();
    }

    /**
     * Returns true when the grid is in picker mode.
     * @returns {boolean}
     */
    _isPickerMode() {
        return this.getAttribute('mode') === 'picker';
    }

    // ---- Data ----

    /**
     * Parses the `data-items` attribute JSON into the internal `_items` array.
     * Logs a warning if the attribute is missing and an error if the JSON is malformed.
     * On success, logs the count of folders and files loaded.
     */
    _readData() {
        var json = this.getAttribute('data-items');
        if (!json) {
            console.warn('[mcms-media-grid] no data-items attribute found');
            this._items = [];
            return;
        }
        try {
            this._items = JSON.parse(json);
            var folders = this._items.filter(function(i) { return i.type === 'folder'; }).length;
            var files = this._items.filter(function(i) { return i.type === 'file'; }).length;
            console.log('[mcms-media-grid] loaded data:', this._items.length, 'items (' + folders + ' folders, ' + files + ' files)');
        } catch (e) {
            console.error('[mcms-media-grid] JSON parse error:', e);
            this._items = [];
        }
    }

    // ---- Drop Overlay ----

    /**
     * Creates and appends the full-screen drop overlay element used for desktop file uploads.
     * The overlay displays an upload icon and instructional text, and handles its own
     * dragover, dragleave, and drop events. Only created once (idempotent).
     */
    _renderDropOverlay() {
        if (this._overlay) return;
        var overlay = document.createElement('div');
        overlay.setAttribute('data-drop-overlay', '');
        overlay.className = 'pointer-events-none absolute inset-0 z-10 hidden flex-col items-center justify-center rounded-lg border-2 border-dashed border-white/20 bg-gray-900/80';

        overlay.innerHTML = '<svg class="size-12 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">'
            + '<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />'
            + '</svg>'
            + '<p class="mt-3 text-sm font-medium text-white">Drop files to upload</p>'
            + '<p class="mt-1 text-xs text-gray-400">Files will be uploaded to the current folder</p>';

        // Overlay must accept drag events when visible
        var self = this;
        overlay.addEventListener('dragover', function(e) {
            e.preventDefault();
            e.stopPropagation();
            e.dataTransfer.dropEffect = 'copy';
        });
        overlay.addEventListener('dragleave', function(e) {
            if (!overlay.contains(e.relatedTarget)) {
                self._hideDropOverlay();
            }
        });
        overlay.addEventListener('drop', function(e) {
            e.preventDefault();
            e.stopPropagation();
            self._hideDropOverlay();
            console.log('[mcms-media-grid] overlay drop, files:', e.dataTransfer.files.length);
            if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
                for (var i = 0; i < e.dataTransfer.files.length; i++) {
                    console.log('[mcms-media-grid] overlay queuing upload:', e.dataTransfer.files[i].name);
                    self._uploadFile(e.dataTransfer.files[i]);
                }
            }
        });

        this._overlay = overlay;
        this.appendChild(overlay);
    }

    /**
     * Shows the drop overlay by removing the `hidden` and `pointer-events-none` classes
     * and adding the `flex` class to make it visible and interactive.
     */
    _showDropOverlay() {
        if (this._overlay) {
            this._overlay.classList.remove('hidden', 'pointer-events-none');
            this._overlay.classList.add('flex');
        }
    }

    /**
     * Hides the drop overlay by removing the `flex` class and restoring
     * `hidden` and `pointer-events-none` classes.
     */
    _hideDropOverlay() {
        if (this._overlay) {
            this._overlay.classList.remove('flex');
            this._overlay.classList.add('hidden', 'pointer-events-none');
        }
    }

    // ---- Render ----

    /**
     * Returns the Tailwind CSS grid class string for the current grid density level.
     * Reads the saved preference from `mcms.settings` under key `media.gridSize`.
     * Falls back to level `'3'` if no preference is stored or the value is unrecognized.
     *
     * Grid density levels:
     * - `'1'`: 2 cols -> 4 cols (sm) -> 6 cols (lg) -- most items visible
     * - `'2'`: 2 cols -> 3 cols (sm) -> 5 cols (lg)
     * - `'3'`: 2 cols -> 3 cols (sm) -> 4 cols (lg) -- default
     * - `'4'`: 1 col  -> 2 cols (sm) -> 3 cols (lg) -- largest thumbnails
     *
     * @returns {string} Space-separated Tailwind CSS class string for the grid layout.
     */
    _savedGridClass() {
        var colMap = {
            '1': 'grid grid-cols-2 gap-x-4 gap-y-8 sm:grid-cols-4 sm:gap-x-6 lg:grid-cols-6 xl:gap-x-8',
            '2': 'grid grid-cols-2 gap-x-4 gap-y-8 sm:grid-cols-3 sm:gap-x-6 lg:grid-cols-5 xl:gap-x-8',
            '3': 'grid grid-cols-2 gap-x-4 gap-y-8 sm:grid-cols-3 sm:gap-x-6 lg:grid-cols-4 xl:gap-x-8',
            '4': 'grid grid-cols-1 gap-x-4 gap-y-8 sm:grid-cols-2 sm:gap-x-6 lg:grid-cols-3 xl:gap-x-8'
        };
        var saved = mcms.settings.get('media.gridSize', '3');
        if (colMap[saved]) return colMap[saved];
        return colMap['3'];
    }

    /**
     * Renders the media grid list element and populates it with folder and file items.
     * Creates the `<ul>` container if it does not already exist, inserts it before
     * the pagination element (if present), and fills it with HTML from `_renderFolder`
     * and `_renderFile`. Shows an empty state message when no items exist.
     * Calls `htmx.process()` on the list to activate HTMX attributes in rendered content.
     */
    _render() {
        var list = this.querySelector('[data-media-grid-list]');
        if (!list) {
            list = document.createElement('ul');
            list.setAttribute('role', 'list');
            list.setAttribute('data-media-grid-list', '');
            list.className = this._savedGridClass();
            var pagination = this.querySelector('#pagination');
            if (pagination) {
                this.insertBefore(list, pagination);
            } else {
                this.appendChild(list);
            }
        }

        if (this._items.length === 0) {
            list.innerHTML = '<li class="col-span-full"><div class="rounded-lg border border-dashed border-white/10 px-6 py-10 text-center"><p class="text-sm text-gray-400">No media found. Upload your first file to get started.</p></div></li>';
            return;
        }

        var detailUrl = this.getAttribute('detail-url') || '/admin/media/';
        var html = '';
        for (var i = 0; i < this._items.length; i++) {
            var item = this._items[i];

            if (item.type === 'folder') {
                html += this._renderFolder(item);
            } else {
                html += this._renderFile(item, detailUrl);
            }
        }
        list.innerHTML = html;
        console.log('[mcms-media-grid] rendered', this._items.length, 'items to grid');

        if (typeof htmx !== 'undefined') htmx.process(list);
    }

    /**
     * Renders the HTML for a single folder item in the grid.
     * Includes a folder icon, HTMX-enabled navigation link, delete confirmation button
     * (via `<mcms-confirm>`), and the folder display name. The folder `<li>` is marked
     * with `data-media-folder-item` and `data-folder-id` attributes for drag-and-drop targeting.
     *
     * @param {Object} item - The folder data object.
     * @param {string} item.id - The folder ID (ULID).
     * @param {string} item.displayName - Human-readable folder name.
     * @returns {string} HTML string for the folder grid item.
     */
    _renderFolder(item) {
        var folderUrl = '/admin/media?folder_id=' + encodeURIComponent(item.id);
        var deleteUrl = '/admin/media-folders/' + encodeURIComponent(item.id);
        var isPicker = this._isPickerMode();

        var navElement;
        if (isPicker) {
            // In picker mode, use a button that dispatches a navigate event instead of HTMX link
            navElement = '<button type="button" data-folder-nav data-folder-id="' + this._esc(item.id) + '" data-folder-url="' + this._esc(folderUrl) + '" class="absolute inset-0 cursor-pointer border-none bg-transparent focus:outline-hidden">'
                + '<span class="sr-only">Open folder ' + this._esc(item.displayName) + '</span>'
                + '</button>';
        } else {
            navElement = '<a href="' + this._esc(folderUrl) + '" hx-get="' + this._esc(folderUrl) + '" hx-target="#main-content" hx-push-url="true" class="absolute inset-0 focus:outline-hidden">'
                + '<span class="sr-only">Open folder ' + this._esc(item.displayName) + '</span>'
                + '</a>';
        }

        var deleteButton = '';
        if (!isPicker) {
            deleteButton = '<div class="absolute top-2 right-2 z-10 hidden group-hover/folder:block">'
                + '<mcms-confirm label="Delete folder" message="Delete this folder? It must be empty." icon=\'<svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>\' button-class="flex size-7 cursor-pointer items-center justify-center rounded-full bg-red-500/90 text-white shadow-sm hover:bg-red-400" hx-delete="' + this._esc(deleteUrl) + '" hx-swap="none"></mcms-confirm>'
                + '</div>';
        }

        var folderCardClass = isPicker
            ? 'group overflow-hidden rounded-lg bg-gray-800'
            : 'group overflow-hidden rounded-lg bg-gray-800 focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-[var(--color-primary)]';

        return '<li class="group/folder relative" data-media-folder-item data-folder-id="' + this._esc(item.id) + '">'
            + '<div class="' + folderCardClass + '">'
            + '<div class="pointer-events-none flex aspect-square items-center justify-center rounded-lg outline -outline-offset-1 outline-white/10 group-hover:opacity-75">'
            + '<svg class="size-12 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12.75V12A2.25 2.25 0 0 1 4.5 9.75h15A2.25 2.25 0 0 1 21.75 12v.75m-8.69-6.44-2.12-2.12a1.5 1.5 0 0 0-1.061-.44H4.5A2.25 2.25 0 0 0 2.25 6v12a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9a2.25 2.25 0 0 0-2.25-2.25h-5.379a1.5 1.5 0 0 1-1.06-.44Z" /></svg>'
            + '</div>'
            + navElement + '</div>'
            + deleteButton
            + '<p class="pointer-events-none mt-2 block truncate text-sm font-medium text-white">' + this._esc(item.displayName) + '</p>'
            + '<p class="pointer-events-none block text-sm font-medium text-gray-400">Folder</p>'
            + '</li>';
    }

    /**
     * Renders the HTML for a single file item in the grid.
     * Image files (mimetype starting with `image/`) display a lazy-loaded thumbnail;
     * non-image files display a bold uppercase file extension badge. Each item is
     * wrapped in a draggable `<li>` with `data-media-grid-item` and `data-media-id`
     * attributes, and includes an HTMX-enabled link to the detail page.
     *
     * @param {Object} item - The file data object.
     * @param {string} item.id - The media ID (ULID).
     * @param {string} item.name - The original filename (used to extract file extension).
     * @param {string} item.displayName - Human-readable display name.
     * @param {string} item.mimetype - MIME type (e.g., `"image/jpeg"`, `"application/pdf"`).
     * @param {string} item.url - URL to the media file (used as image `src`).
     * @param {string} item.alt - Alt text for image files.
     * @param {string} detailUrl - URL prefix for the detail page link.
     * @returns {string} HTML string for the file grid item.
     */
    _renderFile(item, detailUrl) {
        var isImg = item.mimetype && item.mimetype.indexOf('image/') === 0;
        var ext = 'FILE';
        if (item.name) {
            var dot = item.name.lastIndexOf('.');
            if (dot >= 0) ext = item.name.substring(dot + 1).toUpperCase();
        }
        var isPicker = this._isPickerMode();

        var preview;
        if (isImg) {
            preview = '<img src="' + this._esc(item.url) + '" alt="' + this._esc(item.alt) + '" loading="lazy" class="pointer-events-none aspect-square w-full rounded-lg object-cover outline -outline-offset-1 outline-white/10 group-hover:opacity-75" />';
        } else {
            preview = '<div class="pointer-events-none flex aspect-square items-center justify-center rounded-lg outline -outline-offset-1 outline-white/10 group-hover:opacity-75"><span class="text-lg font-bold uppercase text-gray-400">' + this._esc(ext) + '</span></div>';
        }

        var actionElement;
        if (isPicker) {
            // In picker mode, use a button that toggles selection instead of navigating
            actionElement = '<button type="button" data-picker-select data-media-id="' + this._esc(item.id) + '" data-media-url="' + this._esc(item.url) + '" data-media-alt="' + this._esc(item.alt) + '" class="absolute inset-0 cursor-pointer border-none bg-transparent focus:outline-hidden">'
                + '<span class="sr-only">Select ' + this._esc(item.displayName) + '</span>'
                + '</button>';
        } else {
            actionElement = '<a href="' + this._esc(detailUrl + item.id) + '" hx-get="' + this._esc(detailUrl + item.id) + '" hx-target="#main-content" hx-push-url="true" class="absolute inset-0 focus:outline-hidden">'
                + '<span class="sr-only">View details for ' + this._esc(item.displayName) + '</span>'
                + '</a>';
        }

        var cardClass = isPicker
            ? 'group overflow-hidden rounded-lg bg-gray-800'
            : 'group overflow-hidden rounded-lg bg-gray-800 focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-[var(--color-primary)]';

        return '<li id="media-item-' + this._esc(item.id) + '" class="relative" data-media-grid-item draggable="' + (isPicker ? 'false' : 'true') + '" data-media-id="' + this._esc(item.id) + '" data-media-url="' + this._esc(item.url) + '" data-media-alt="' + this._esc(item.alt) + '" data-media-name="' + this._esc(item.displayName) + '">'
            + '<div class="' + cardClass + '">'
            + preview
            + actionElement + '</div>'
            + '<p class="pointer-events-none mt-2 block truncate text-sm font-medium text-white">' + this._esc(item.displayName) + '</p>'
            + '<p class="pointer-events-none block text-sm font-medium text-gray-400">' + this._esc(item.mimetype) + '</p>'
            + '</li>';
    }

    /**
     * Escapes a string for safe insertion into HTML by creating a text node
     * and reading its parent's innerHTML. Returns an empty string for falsy input.
     *
     * @param {string} s - The string to HTML-escape.
     * @returns {string} The HTML-escaped string.
     */
    _esc(s) {
        if (!s) return '';
        var d = document.createElement('div');
        d.appendChild(document.createTextNode(s));
        return d.innerHTML;
    }

    // ---- Grid Events ----

    /**
     * Binds all drag-and-drop and click event listeners to the component.
     *
     * Registered handlers:
     * - `dragstart`: Initiates media item drag with a custom ghost image (thumbnail or
     *   file extension badge). Sets the `application/x-mcms-media` data transfer type
     *   with the media ID and reduces the dragged item's opacity.
     * - `dragend`: Cleans up drag state, removes the custom ghost element, and clears
     *   all `data-drop-target` attributes from folder elements.
     * - `dragover`: Handles two drag scenarios -- internal media items being dragged onto
     *   folders (move) or other items (reorder), and desktop files being dragged onto the
     *   grid (upload). Shows the drop overlay for desktop file drags.
     * - `dragleave`: Removes drop target highlights and hides the overlay when the drag
     *   leaves the component bounds.
     * - `drop`: Processes drops for internal media moves (to folders or reorder) and
     *   desktop file uploads.
     * - `click`: Handles item selection with support for single click (exclusive select),
     *   Ctrl/Cmd+click (toggle individual), and Shift+click (range select). Ignores
     *   clicks on links and buttons. Dispatches `mcms-media-grid:select` and updates
     *   the bulk action bar.
     */
    _bind() {
        var self = this;
        this._onDragStart = function(e) {
            var item = e.target.closest('[data-media-grid-item]');
            if (!item) return;
            var id = item.getAttribute('data-media-id');
            if (!id) return;
            self._dragId = id;
            e.dataTransfer.setData('application/x-mcms-media', id);
            e.dataTransfer.effectAllowed = 'move';
            item.setAttribute('data-dragging', '');

            // Build a custom drag image from the grid item's image or file badge
            var img = item.querySelector('img');
            var ghost = document.createElement('div');
            ghost.style.cssText = 'position:absolute;top:-9999px;left:-9999px;width:100px;height:100px;border-radius:8px;overflow:hidden;box-shadow:0 8px 24px rgba(0,0,0,0.5);pointer-events:none;';
            if (img) {
                var clone = img.cloneNode(false);
                clone.style.cssText = 'width:100%;height:100%;object-fit:cover;display:block;';
                clone.removeAttribute('loading');
                ghost.appendChild(clone);
            } else {
                var ext = item.querySelector('.text-lg');
                ghost.style.cssText += 'display:flex;align-items:center;justify-content:center;background:#1f2937;';
                var label = document.createElement('span');
                label.textContent = ext ? ext.textContent : 'FILE';
                label.style.cssText = 'color:#9ca3af;font-weight:700;font-size:14px;text-transform:uppercase;';
                ghost.appendChild(label);
            }
            document.body.appendChild(ghost);
            self._dragGhost = ghost;
            e.dataTransfer.setDragImage(ghost, 50, 50);

            requestAnimationFrame(function() { item.style.opacity = '0.4'; });
        };
        this._onDragEnd = function(e) {
            var item = e.target.closest('[data-media-grid-item]');
            if (item) { item.removeAttribute('data-dragging'); item.style.opacity = ''; }
            self._dragId = null;
            // Remove the custom drag ghost from the DOM
            if (self._dragGhost) {
                self._dragGhost.remove();
                self._dragGhost = null;
            }
            // Clean up all folder drop targets
            var targets = document.querySelectorAll('[data-drop-target]');
            for (var i = 0; i < targets.length; i++) {
                targets[i].removeAttribute('data-drop-target');
                targets[i].classList.remove('media-folder-drop-target');
            }
        };
        this._onGridDragOver = function(e) {
            if (e.dataTransfer.types.indexOf('application/x-mcms-media') >= 0) {
                // Drop onto folder card = move into folder
                var folder = e.target.closest('[data-media-folder-item]');
                if (folder) {
                    e.preventDefault();
                    e.dataTransfer.dropEffect = 'move';
                    folder.setAttribute('data-drop-target', '');
                    return;
                }
                // Drop onto another media item = reorder
                var target = e.target.closest('[data-media-grid-item]');
                if (target && target.getAttribute('data-media-id') !== self._dragId) {
                    e.preventDefault();
                    e.dataTransfer.dropEffect = 'move';
                    target.setAttribute('data-drop-target', '');
                }
                return;
            }
            // File drop from desktop
            if (e.dataTransfer.types.indexOf('Files') >= 0) {
                e.preventDefault();
                e.dataTransfer.dropEffect = 'copy';
                self._showDropOverlay();
            }
        };
        this._onGridDragLeave = function(e) {
            var target = e.target.closest('[data-media-grid-item], [data-media-folder-item]');
            if (target) target.removeAttribute('data-drop-target');
            if (e.target === self || !self.contains(e.relatedTarget)) {
                self._hideDropOverlay();
            }
        };
        this._onGridDrop = function(e) {
            self._hideDropOverlay();
            if (e.dataTransfer.types.indexOf('application/x-mcms-media') >= 0) {
                var mediaId = e.dataTransfer.getData('application/x-mcms-media');
                console.log('[mcms-media-grid] drop event, mediaId:', mediaId, 'target:', e.target.tagName, 'target classes:', e.target.className);
                // Drop onto folder card = move into folder
                var folder = e.target.closest('[data-media-folder-item]');
                console.log('[mcms-media-grid] closest folder item:', folder, 'folder-id:', folder ? folder.getAttribute('data-folder-id') : 'N/A');
                if (folder) {
                    e.preventDefault();
                    folder.removeAttribute('data-drop-target');
                    var folderId = folder.getAttribute('data-folder-id') || '';
                    console.log('[mcms-media-grid] moving to folder, folderId:', JSON.stringify(folderId), 'empty?', folderId === '');
                    self._moveToFolder(mediaId, folderId);
                    return;
                }
                // Drop onto another media item = reorder
                var target = e.target.closest('[data-media-grid-item]');
                if (!target) return;
                e.preventDefault();
                target.removeAttribute('data-drop-target');
                var dragEl = self.querySelector('[data-dragging]');
                if (!dragEl || dragEl === target) return;
                var list = self.querySelector('[data-media-grid-list]');
                if (!list) return;
                var items = Array.from(list.querySelectorAll('[data-media-grid-item]'));
                var dragIdx = items.indexOf(dragEl);
                var dropIdx = items.indexOf(target);
                if (dragIdx < dropIdx) {
                    list.insertBefore(dragEl, target.nextSibling);
                } else {
                    list.insertBefore(dragEl, target);
                }
                return;
            }
            // File drop from desktop (grid handler)
            if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
                e.preventDefault();
                console.log('[mcms-media-grid] grid drop, files:', e.dataTransfer.files.length);
                for (var i = 0; i < e.dataTransfer.files.length; i++) {
                    console.log('[mcms-media-grid] grid queuing upload:', e.dataTransfer.files[i].name);
                    self._uploadFile(e.dataTransfer.files[i]);
                }
            }
        };
        this._onGridClick = function(e) {
            // ---- Picker mode: folder navigation ----
            if (self._isPickerMode()) {
                var folderNav = e.target.closest('[data-folder-nav]');
                if (folderNav) {
                    e.preventDefault();
                    e.stopPropagation();
                    self.dispatchEvent(new CustomEvent('mcms-media-grid:navigate', {
                        bubbles: true,
                        detail: {
                            folderId: folderNav.getAttribute('data-folder-id') || '',
                            folderUrl: folderNav.getAttribute('data-folder-url') || ''
                        }
                    }));
                    return;
                }

                // ---- Picker mode: file selection ----
                var pickerBtn = e.target.closest('[data-picker-select]');
                if (pickerBtn) {
                    e.preventDefault();
                    e.stopPropagation();
                    var pickerItem = pickerBtn.closest('[data-media-grid-item]');
                    if (!pickerItem) return;
                    var pickId = pickerItem.getAttribute('data-media-id');
                    if (!pickId) return;

                    var isMultiple = self.hasAttribute('multiple');

                    if (pickerItem.hasAttribute('data-selected')) {
                        // Deselect
                        self._selected.delete(pickId);
                        pickerItem.removeAttribute('data-selected');
                    } else {
                        // Select (deselect others first in single-select mode)
                        if (!isMultiple) {
                            self._clearSelection();
                        }
                        self._selected.add(pickId);
                        pickerItem.setAttribute('data-selected', '');
                    }
                    self._dispatchPickEvent();
                    return;
                }
                return;
            }

            // ---- Default mode: existing selection behavior ----
            if (e.target.closest('a, button')) return;
            var item = e.target.closest('[data-media-grid-item]');
            if (!item) return;
            var id = item.getAttribute('data-media-id');
            if (!id) return;

            if (e.ctrlKey || e.metaKey) {
                if (self._selected.has(id)) { self._selected.delete(id); item.removeAttribute('data-selected'); }
                else { self._selected.add(id); item.setAttribute('data-selected', ''); }
            } else if (e.shiftKey) {
                var all = Array.from(self.querySelectorAll('[data-media-grid-item]'));
                var ids = all.map(function(el) { return el.getAttribute('data-media-id'); });
                var last = Array.from(self._selected).pop();
                var si = last ? ids.indexOf(last) : 0;
                var ei = ids.indexOf(id);
                if (si < 0) si = 0;
                if (si > ei) { var t = si; si = ei; ei = t; }
                for (var j = si; j <= ei; j++) { self._selected.add(ids[j]); all[j].setAttribute('data-selected', ''); }
            } else {
                self._clearSelection();
                self._selected.add(id);
                item.setAttribute('data-selected', '');
            }
            self.dispatchEvent(new CustomEvent('mcms-media-grid:select', { bubbles: true, detail: { selected: Array.from(self._selected) } }));
            self._updateBulkBar();
        };

        this.addEventListener('dragstart', this._onDragStart);
        this.addEventListener('dragend', this._onDragEnd);
        this.addEventListener('dragover', this._onGridDragOver);
        this.addEventListener('dragleave', this._onGridDragLeave);
        this.addEventListener('drop', this._onGridDrop);
        this.addEventListener('click', this._onGridClick);
    }

    /**
     * Removes all drag-and-drop and click event listeners that were bound in `_bind()`.
     */
    _unbind() {
        this.removeEventListener('dragstart', this._onDragStart);
        this.removeEventListener('dragend', this._onDragEnd);
        this.removeEventListener('dragover', this._onGridDragOver);
        this.removeEventListener('dragleave', this._onGridDragLeave);
        this.removeEventListener('drop', this._onGridDrop);
        this.removeEventListener('click', this._onGridClick);
    }

    // ---- Folder Move ----

    /**
     * Moves a media item to a folder by sending a POST request to the server.
     * On success, removes the item from the DOM, shows a success toast, and dispatches
     * the `mcms-media-grid:move` event. On failure, shows an error toast.
     *
     * The request includes CSRF token, HX-Request header (for server-side HTMX detection),
     * and sends the folder_id as URL-encoded form data.
     *
     * @param {string} mediaId - The ID of the media item to move.
     * @param {string} folderId - The destination folder ID (empty string for root).
     */
    _moveToFolder(mediaId, folderId) {
        console.log('[mcms-media-grid] move media', mediaId, 'to folder', folderId);
        var csrf = this._getCSRF();
        var self = this;
        var moveUrl = '/admin/media/move/' + encodeURIComponent(mediaId);
        console.log('[mcms-media-grid] POST', moveUrl);

        fetch(moveUrl, {
            method: 'POST',
            headers: {
                'X-CSRF-Token': csrf,
                'HX-Request': 'true',
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: 'folder_id=' + encodeURIComponent(folderId),
            credentials: 'same-origin'
        }).then(function(res) {
            console.log('[mcms-media-grid] move response:', res.status, res.statusText);
            if (res.ok) {
                self._showToast('Media moved to folder', 'success');
                var item = document.getElementById('media-item-' + mediaId);
                if (item) item.remove();
                self.dispatchEvent(new CustomEvent('mcms-media-grid:move', {
                    bubbles: true,
                    detail: { mediaId: mediaId, folderId: folderId }
                }));
            } else {
                res.text().then(function(body) { console.error('[mcms-media-grid] move failed:', res.status, body); });
                self._showToast('Failed to move media', 'error');
            }
        }).catch(function(err) {
            console.error('[mcms-media-grid] move network error:', err);
            self._showToast('Network error while moving media', 'error');
        });
    }

    // ---- HTMX Integration ----

    /**
     * Binds a global `htmx:afterSwap` listener on `document.body` that re-initializes
     * lucide icons after HTMX swaps content into `#main-content`. This ensures folder
     * navigation and other HTMX-driven page transitions render icons correctly.
     */
    _bindFolders() {
        var self = this;
        this._onAfterSwap = function(e) {
            if (e.detail.target && e.detail.target.id === 'main-content') {
                if (typeof lucide !== 'undefined') {
                    setTimeout(function() { lucide.createIcons(); }, 10);
                }
            }
        };
        document.body.addEventListener('htmx:afterSwap', this._onAfterSwap);
    }

    /**
     * Removes the global `htmx:afterSwap` listener that was bound in `_bindFolders()`.
     */
    _unbindFolders() {
        document.body.removeEventListener('htmx:afterSwap', this._onAfterSwap);
    }

    // ---- Upload ----

    /**
     * Uploads a single file to the server via a multipart POST request.
     * Inserts an animated skeleton placeholder into the grid while uploading.
     * On success, dispatches `mcms-media-grid:upload` and triggers an HTMX reload
     * of the current page to refresh the grid with the new item. On failure, shows
     * an error toast.
     *
     * The upload includes the file, the current folder ID (from the `folder-id` attribute),
     * and a CSRF token in both the form body and request header.
     *
     * @param {File} file - The File object to upload (from drag-and-drop or file input).
     */
    _uploadFile(file) {
        var url = this.getAttribute('upload-url') || '/admin/media';
        var csrf = this._getCSRF();
        var folderId = this.getAttribute('folder-id') || '';
        var fd = new FormData();
        fd.append('file', file);
        fd.append('folder_id', folderId);
        fd.append('_csrf', csrf);

        var skeletonId = this._insertUploadSkeleton(file.name);
        var self = this;
        fetch(url, {
            method: 'POST',
            headers: { 'X-CSRF-Token': csrf, 'HX-Request': 'true' },
            body: fd,
            credentials: 'same-origin'
        }).then(function(res) {
            self._removeUploadSkeleton(skeletonId);
            if (res.ok) {
                var mediaId = res.headers.get('X-Media-ID') || '';
                var mediaUrl = res.headers.get('X-Media-URL') || '';
                self._showToast('Uploaded: ' + file.name, 'success');
                self.dispatchEvent(new CustomEvent('mcms-media-grid:upload', { bubbles: true, detail: { mediaId: mediaId, url: mediaUrl, filename: file.name } }));
                if (typeof htmx !== 'undefined') {
                    htmx.ajax('GET', window.location.pathname + window.location.search, { target: '#main-content', swap: 'innerHTML' });
                }
            } else {
                self._showToast('Failed to upload: ' + file.name, 'error');
            }
        }).catch(function(err) {
            self._removeUploadSkeleton(skeletonId);
            self._showToast('Network error uploading: ' + file.name, 'error');
        });
    }

    // ---- Picker Events ----

    /**
     * Dispatches the `mcms-media-grid:pick` event with full metadata for all selected items.
     * Reads `data-media-url`, `data-media-alt`, and `data-media-name` from the DOM elements.
     */
    _dispatchPickEvent() {
        var selected = [];
        var items = this.querySelectorAll('[data-media-grid-item][data-selected]');
        for (var i = 0; i < items.length; i++) {
            selected.push({
                id: items[i].getAttribute('data-media-id') || '',
                url: items[i].getAttribute('data-media-url') || '',
                alt: items[i].getAttribute('data-media-alt') || '',
                name: items[i].getAttribute('data-media-name') || ''
            });
        }
        this.dispatchEvent(new CustomEvent('mcms-media-grid:pick', {
            bubbles: true,
            detail: { selected: selected }
        }));
    }

    // ---- Helpers ----

    /**
     * Clears all selected items by emptying the `_selected` set and removing
     * the `data-selected` attribute from all grid items in the DOM.
     */
    _clearSelection() {
        this._selected.clear();
        var items = this.querySelectorAll('[data-selected]');
        for (var i = 0; i < items.length; i++) items[i].removeAttribute('data-selected');
    }

    /**
     * Reads the CSRF token from the `<meta name="csrf-token">` element in the document head.
     *
     * @returns {string} The CSRF token value, or an empty string if the meta tag is not found.
     */
    _getCSRF() {
        var meta = document.querySelector('meta[name="csrf-token"]');
        return meta ? meta.content : '';
    }

    /**
     * Shows a toast notification by delegating to the `<mcms-toast>` component in the DOM.
     *
     * @param {string} msg - The message text to display.
     * @param {string} type - The toast type (e.g., `"success"`, `"error"`).
     */
    _showToast(msg, type) {
        var t = document.querySelector('mcms-toast');
        if (t) t.show(msg, type);
    }

    // ---- Display Size ----

    /**
     * Changes the grid density and persists the preference.
     * Updates the grid list element's CSS classes to match the new density level
     * and saves the setting via `mcms.settings.set()`.
     *
     * @param {string} level - The grid density level (`"1"` through `"4"`).
     *   Level `"1"` shows the most columns (smallest thumbnails); level `"4"` shows
     *   the fewest columns (largest thumbnails).
     *
     * @example
     * document.querySelector('mcms-media-grid').setGridSize('2');
     */
    setGridSize(level) {
        var list = this.querySelector('[data-media-grid-list]');
        if (!list) return;
        mcms.settings.set('media.gridSize', level);
        list.className = this._savedGridClass();
    }

    // ---- Bulk Delete ----

    /**
     * Sends a bulk delete request for all currently selected media items.
     * Posts a JSON body with `{ ids: [...] }` to `/admin/media/bulk-delete`.
     * On success, clears the selection, hides the bulk bar, and triggers an HTMX
     * reload of the current page. Parses the `HX-Trigger` response header to
     * display server-sent toast messages.
     */
    _bulkDelete() {
        var ids = Array.from(this._selected);
        if (ids.length === 0) return;
        var csrf = this._getCSRF();
        var self = this;

        fetch('/admin/media/bulk-delete', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': csrf,
                'HX-Request': 'true',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ ids: ids }),
            credentials: 'same-origin'
        }).then(function(res) {
            if (res.ok) {
                self._clearSelection();
                self._updateBulkBar();
                if (typeof htmx !== 'undefined') {
                    htmx.ajax('GET', window.location.pathname + window.location.search, { target: '#main-content', swap: 'innerHTML' });
                }
            }
            // Toast is sent via HX-Trigger header and handled by htmx
            var trigger = res.headers.get('HX-Trigger');
            if (trigger) {
                try {
                    var parsed = JSON.parse(trigger);
                    if (parsed.showToast) self._showToast(parsed.showToast.message, parsed.showToast.type);
                } catch (e) {
                    // header parsing failed, ignore
                }
            }
        }).catch(function(err) {
            self._showToast('Network error during bulk delete', 'error');
        });
    }

    // ---- Bulk Bar ----

    /**
     * Updates the bulk action bar visibility and selection count display.
     * Looks for `#media-bulk-bar` and `#bulk-count` elements in the DOM (expected
     * to be rendered by the server). Shows the bar when items are selected,
     * hides it when the selection is empty.
     */
    _updateBulkBar() {
        var bar = document.getElementById('media-bulk-bar');
        var count = document.getElementById('bulk-count');
        if (!bar || !count) return;
        var n = this._selected.size;
        count.textContent = n;
        if (n > 0) {
            bar.classList.remove('hidden');
            bar.classList.add('flex');
        } else {
            bar.classList.remove('flex');
            bar.classList.add('hidden');
        }
    }

    // ---- Upload Skeleton ----

    /**
     * Inserts an animated skeleton placeholder at the beginning of the grid list
     * to indicate an upload is in progress. The skeleton displays a spinning icon
     * and the filename with an "Uploading..." label.
     *
     * @param {string} filename - The name of the file being uploaded, displayed below the spinner.
     * @returns {string} The DOM ID of the inserted skeleton element, used to remove it later.
     */
    _insertUploadSkeleton(filename) {
        var list = this.querySelector('[data-media-grid-list]');
        if (!list) return;
        var id = 'upload-skeleton-' + Date.now();
        var li = document.createElement('li');
        li.id = id;
        li.className = 'relative animate-pulse';
        li.innerHTML = '<div class="overflow-hidden rounded-lg bg-gray-800">'
            + '<div class="flex aspect-square items-center justify-center rounded-lg bg-gray-700">'
            + '<svg class="size-8 animate-spin text-gray-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v4m0 12v4m-7.07-3.93l2.83-2.83m8.49-8.49l2.83-2.83M2 12h4m12 0h4M4.93 4.93l2.83 2.83m8.49 8.49l2.83 2.83" /></svg>'
            + '</div></div>'
            + '<p class="mt-2 block truncate text-sm font-medium text-gray-500">' + this._esc(filename) + '</p>'
            + '<p class="block text-sm font-medium text-gray-600">Uploading...</p>';
        list.prepend(li);
        return id;
    }

    /**
     * Removes an upload skeleton placeholder from the DOM by its ID.
     *
     * @param {string} skeletonId - The DOM ID of the skeleton element to remove
     *   (as returned by `_insertUploadSkeleton`).
     */
    _removeUploadSkeleton(skeletonId) {
        var el = document.getElementById(skeletonId);
        if (el) el.remove();
    }

    // ---- Public API ----

    /**
     * Returns an array of currently selected media IDs.
     *
     * @returns {string[]} Array of selected media ID strings.
     *
     * @example
     * var ids = document.querySelector('mcms-media-grid').getSelected();
     * console.log('Selected:', ids); // ["01HX...", "01HY..."]
     */
    getSelected() { return Array.from(this._selected); }

    /**
     * Clears the current selection and hides the bulk action bar.
     *
     * @example
     * document.querySelector('mcms-media-grid').clearSelection();
     */
    clearSelection() { this._clearSelection(); this._updateBulkBar(); }

    /**
     * Returns an array of selected items with full metadata (picker mode).
     * Reads data attributes from the DOM to build each item object.
     *
     * @returns {Array<{id: string, url: string, alt: string, name: string}>}
     *
     * @example
     * var items = document.querySelector('mcms-media-grid').getSelectedItems();
     * console.log(items); // [{id: "01HX...", url: "/media/...", alt: "...", name: "..."}]
     */
    getSelectedItems() {
        var selected = [];
        var items = this.querySelectorAll('[data-media-grid-item][data-selected]');
        for (var i = 0; i < items.length; i++) {
            selected.push({
                id: items[i].getAttribute('data-media-id') || '',
                url: items[i].getAttribute('data-media-url') || '',
                alt: items[i].getAttribute('data-media-alt') || '',
                name: items[i].getAttribute('data-media-name') || ''
            });
        }
        return selected;
    }

    /**
     * Re-reads the `data-items` attribute and re-renders the grid.
     * Use this after the server updates the attribute value (e.g., via HTMX swap).
     *
     * @example
     * var grid = document.querySelector('mcms-media-grid');
     * grid.setAttribute('data-items', JSON.stringify(newItems));
     * grid.refresh();
     */
    refresh() { this._readData(); this._render(); }
}

customElements.define('mcms-media-grid', McmsMediaGrid);
