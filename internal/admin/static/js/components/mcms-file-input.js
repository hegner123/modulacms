/**
 * @module mcms-file-input
 * @description Styled file input component with drag-and-drop support, file previews,
 * size validation, and a removable file list.
 *
 * Renders a visually rich drop zone that replaces the native file input. Supports both
 * click-to-browse and drag-and-drop workflows. Selected files are displayed in a list
 * with thumbnails (for images) or file extension badges (for other types), file size,
 * and a remove button. Files exceeding the optional `max-size` limit are rejected with
 * a temporary error message.
 *
 * The native `<input type="file">` is kept in sync with the displayed file list via
 * the `DataTransfer` API, ensuring form submissions include the correct files.
 *
 * This is a Light DOM web component -- all rendered elements are direct children of the
 * host element, making them accessible to global CSS and form handling.
 *
 * @example
 * <!-- Single image upload with size limit -->
 * <mcms-file-input
 *     name="avatar"
 *     accept="image/*"
 *     label="Profile Photo"
 *     max-size="5242880"
 * ></mcms-file-input>
 *
 * @example
 * <!-- Multiple file upload -->
 * <mcms-file-input
 *     name="attachments"
 *     accept=".pdf,.doc,.docx"
 *     multiple
 *     required
 *     label="Documents"
 * ></mcms-file-input>
 *
 * @example
 * <!-- Listen for file changes -->
 * <mcms-file-input name="media" id="my-input"></mcms-file-input>
 * <script>
 *   document.getElementById('my-input').addEventListener('file-change', (e) => {
 *     console.log('Selected files:', e.detail.files);
 *     console.log('Input name:', e.detail.name);
 *   });
 * </script>
 */

/**
 * @event file-change
 * @description Fired when the set of selected files changes, either through the file
 * picker, drag-and-drop, or file removal. Bubbles up through the DOM.
 * @type {CustomEvent}
 * @property {Object} detail
 * @property {string} detail.name - The `name` attribute value of this file input.
 * @property {File[]} detail.files - Array of currently selected `File` objects.
 */

/**
 * Styled file input with drag-and-drop, previews, and validation.
 *
 * Builds a drop zone with a hidden native `<input type="file">`, a visual file list
 * showing thumbnails or extension badges, and individual remove buttons. The native
 * input's `FileList` is kept synchronized via `DataTransfer` so form submissions work
 * correctly.
 *
 * @extends HTMLElement
 *
 * @attr {string} name - The form field name for the file input. Default: `"file"`.
 * @attr {string} accept - Comma-separated list of accepted MIME types and/or file
 *   extensions (e.g., `"image/*,.pdf"`). Applied to both the native input and
 *   drag-and-drop filtering.
 * @attr {boolean} multiple - When present, allows selecting multiple files.
 * @attr {boolean} required - When present, marks the native input as required.
 * @attr {string} label - Optional label text displayed above the drop zone.
 * @attr {string} max-size - Maximum file size in bytes. Files exceeding this limit
 *   are rejected with a temporary error message. Default: `0` (no limit).
 *
 * @fires file-change
 */
class McmsFileInput extends HTMLElement {
    constructor() {
        super();

        /**
         * Array of currently selected File objects.
         * @type {File[]}
         * @private
         */
        this._files = [];

        /**
         * Reference to the hidden native file input element.
         * @type {HTMLInputElement|null}
         * @private
         */
        this._input = null;

        /**
         * Reference to the drop zone container element.
         * @type {HTMLDivElement|null}
         * @private
         */
        this._dropZone = null;

        /**
         * Reference to the file list container where file items are rendered.
         * @type {HTMLDivElement|null}
         * @private
         */
        this._fileList = null;
    }

    /**
     * Lifecycle callback invoked when the element is inserted into the DOM.
     *
     * Reads configuration attributes, builds the drop zone with an upload icon and
     * browse prompt, creates the hidden native file input, attaches the file list
     * container, and binds all event handlers for click, keyboard, file selection,
     * and drag-and-drop interactions.
     */
    connectedCallback() {
        var name = this.getAttribute('name') || 'file';
        var accept = this.getAttribute('accept') || '';
        var multiple = this.hasAttribute('multiple');
        var required = this.hasAttribute('required');
        var label = this.getAttribute('label') || '';

        var wrapper = document.createElement('div');
        wrapper.className = 'flex flex-col gap-1';

        if (label) {
            var labelEl = document.createElement('label');
            labelEl.className = 'block text-sm/6 font-medium text-white';
            labelEl.textContent = label;
            wrapper.appendChild(labelEl);
        }

        // Hidden native file input
        var input = document.createElement('input');
        input.type = 'file';
        input.name = name;
        input.className = 'sr-only';
        if (accept) input.accept = accept;
        if (multiple) input.multiple = true;
        if (required) input.required = true;
        this._input = input;

        // Drop zone
        var dropZone = document.createElement('div');
        dropZone.className = 'flex flex-col items-center justify-center rounded-lg border-2 border-dashed border-[var(--color-border)] bg-[var(--color-bg)] p-8 text-center text-sm text-[var(--color-text-muted)] transition-colors cursor-pointer';
        dropZone.setAttribute('tabindex', '0');
        dropZone.setAttribute('role', 'button');
        dropZone.setAttribute('aria-label', 'Choose file or drag and drop');
        this._dropZone = dropZone;

        var icon = document.createElement('div');
        icon.className = 'mb-2 text-[var(--color-text-dim)]';
        icon.innerHTML = '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>';
        dropZone.appendChild(icon);

        var prompt = document.createElement('div');
        prompt.className = 'text-sm text-[var(--color-text-muted)]';
        var browseText = document.createElement('span');
        browseText.className = 'text-[var(--color-primary)] font-medium cursor-pointer';
        browseText.textContent = 'Choose file';
        prompt.appendChild(browseText);
        prompt.appendChild(document.createTextNode(' or drag and drop'));
        dropZone.appendChild(prompt);

        if (accept) {
            var hint = document.createElement('div');
            hint.className = 'mt-1 text-xs text-[var(--color-text-dim)]';
            hint.textContent = accept;
            dropZone.appendChild(hint);
        }

        dropZone.appendChild(input);

        wrapper.appendChild(dropZone);

        // File list container
        var fileList = document.createElement('div');
        fileList.className = 'mt-3 space-y-2';
        this._fileList = fileList;
        wrapper.appendChild(fileList);

        this.appendChild(wrapper);

        this._bindEvents(dropZone, input);
    }

    /**
     * Binds all interactive event handlers to the drop zone and file input.
     *
     * Attaches click and keyboard handlers to the drop zone for opening the native
     * file picker, a change handler on the file input, and dragenter/dragover/dragleave/drop
     * handlers for drag-and-drop support. Drag state is indicated via both a `data-dragover`
     * attribute and a `file-input-dragover` CSS class (dual approach for CSS/JS targeting).
     *
     * Dropped files are filtered against the `accept` attribute and limited to one file
     * when `multiple` is not set. The filtered files are transferred to the native input
     * via the `DataTransfer` API.
     *
     * @param {HTMLDivElement} dropZone - The drop zone container element.
     * @param {HTMLInputElement} input - The hidden native file input element.
     * @private
     */
    _bindEvents(dropZone, input) {
        var self = this;

        // Click the drop zone to open file picker
        dropZone.addEventListener('click', function(e) {
            if (e.target === input) return;
            input.click();
        });

        // Keyboard: Enter or Space opens file picker
        dropZone.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                input.click();
            }
        });

        // File input change
        input.addEventListener('change', function() {
            self._handleFiles(input.files);
        });

        // Drag-and-drop
        dropZone.addEventListener('dragenter', function(e) {
            e.preventDefault();
            e.stopPropagation();
            // DUAL: data-dragover + class
            dropZone.dataset.dragover = '';
            dropZone.classList.add('file-input-dragover');
        });

        dropZone.addEventListener('dragover', function(e) {
            e.preventDefault();
            e.stopPropagation();
            // DUAL: data-dragover + class
            dropZone.dataset.dragover = '';
            dropZone.classList.add('file-input-dragover');
        });

        dropZone.addEventListener('dragleave', function(e) {
            e.preventDefault();
            e.stopPropagation();
            // DUAL: data-dragover + class
            delete dropZone.dataset.dragover;
            dropZone.classList.remove('file-input-dragover');
        });

        dropZone.addEventListener('drop', function(e) {
            e.preventDefault();
            e.stopPropagation();
            // DUAL: data-dragover + class
            delete dropZone.dataset.dragover;
            dropZone.classList.remove('file-input-dragover');

            var dt = e.dataTransfer;
            if (!dt || !dt.files || dt.files.length === 0) return;

            // Transfer dropped files to the native input via DataTransfer
            var transfer = new DataTransfer();
            var accept = self.getAttribute('accept') || '';
            var multiple = self.hasAttribute('multiple');

            for (var i = 0; i < dt.files.length; i++) {
                if (!multiple && i > 0) break;
                if (accept && !self._matchesAccept(dt.files[i], accept)) continue;
                transfer.items.add(dt.files[i]);
            }

            input.files = transfer.files;
            self._handleFiles(transfer.files);
        });
    }

    /**
     * Processes a list of selected files, applying size validation and updating the UI.
     *
     * Clears the current file list, validates each file against the `max-size` attribute,
     * adds valid files to the internal array, renders file items in the list, and
     * dispatches a `file-change` event with the updated file set.
     *
     * @param {FileList} fileListObj - The FileList from the native input or DataTransfer.
     * @private
     */
    _handleFiles(fileListObj) {
        var self = this;
        var maxSize = parseInt(this.getAttribute('max-size') || '0', 10);
        this._files = [];
        this._fileList.textContent = '';

        if (!fileListObj || fileListObj.length === 0) return;

        for (var i = 0; i < fileListObj.length; i++) {
            var file = fileListObj[i];

            if (maxSize > 0 && file.size > maxSize) {
                this._showError('File "' + file.name + '" exceeds max size (' + this._formatSize(maxSize) + ')');
                continue;
            }

            this._files.push(file);
            this._renderFileItem(file, i);
        }

        this.dispatchEvent(new CustomEvent('file-change', {
            bubbles: true,
            detail: { name: this.getAttribute('name') || 'file', files: this._files }
        }));
    }

    /**
     * Renders a single file item in the file list with thumbnail/icon, name, size,
     * and a remove button.
     *
     * For image files, generates a thumbnail using `FileReader.readAsDataURL()`.
     * For non-image files, displays the file extension in uppercase as a badge.
     * The remove button calls `_removeFile()` to update the file list.
     *
     * @param {File} file - The File object to render.
     * @param {number} index - The index of this file in the `_files` array, used
     *   for removal targeting.
     * @private
     */
    _renderFileItem(file, index) {
        var self = this;
        var item = document.createElement('div');
        item.className = 'flex items-center gap-3 rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 text-sm';

        // Thumbnail or file icon
        var thumb = document.createElement('div');
        thumb.className = 'w-10 h-10 rounded overflow-hidden flex items-center justify-center bg-[var(--color-bg)] shrink-0';

        if (file.type && file.type.indexOf('image/') === 0) {
            var img = document.createElement('img');
            img.alt = file.name;
            var reader = new FileReader();
            reader.onload = function(e) { img.src = e.target.result; };
            reader.readAsDataURL(file);
            thumb.appendChild(img);
        } else {
            var ext = document.createElement('span');
            ext.className = 'text-[0.625rem] font-bold text-[var(--color-text-dim)] uppercase';
            var parts = file.name.split('.');
            ext.textContent = parts.length > 1 ? parts[parts.length - 1].toUpperCase() : 'FILE';
            thumb.appendChild(ext);
        }
        item.appendChild(thumb);

        // File info
        var info = document.createElement('div');
        info.className = 'flex flex-col flex-1 min-w-0';

        var nameEl = document.createElement('span');
        nameEl.className = 'text-sm text-[var(--color-text)] truncate';
        nameEl.textContent = file.name;
        info.appendChild(nameEl);

        var sizeEl = document.createElement('span');
        sizeEl.className = 'text-xs text-[var(--color-text-dim)]';
        sizeEl.textContent = this._formatSize(file.size);
        info.appendChild(sizeEl);

        item.appendChild(info);

        // Remove button
        var removeBtn = document.createElement('button');
        removeBtn.type = 'button';
        removeBtn.className = 'ml-auto cursor-pointer bg-transparent border-none text-[var(--color-text-muted)] hover:text-[var(--color-danger)] text-lg leading-none';
        removeBtn.setAttribute('aria-label', 'Remove file');
        removeBtn.innerHTML = '&times;';
        removeBtn.addEventListener('click', function() {
            self._removeFile(index);
        });
        item.appendChild(removeBtn);

        this._fileList.appendChild(item);
    }

    /**
     * Removes a file from the selection by index, rebuilds the native input's FileList
     * via DataTransfer, re-renders the file list, and dispatches a `file-change` event.
     *
     * @param {number} index - The index of the file to remove from the `_files` array.
     * @private
     */
    _removeFile(index) {
        this._files.splice(index, 1);

        // Rebuild the native input's FileList
        var transfer = new DataTransfer();
        for (var i = 0; i < this._files.length; i++) {
            transfer.items.add(this._files[i]);
        }
        this._input.files = transfer.files;

        // Re-render list
        this._fileList.textContent = '';
        for (var j = 0; j < this._files.length; j++) {
            this._renderFileItem(this._files[j], j);
        }

        this.dispatchEvent(new CustomEvent('file-change', {
            bubbles: true,
            detail: { name: this.getAttribute('name') || 'file', files: this._files }
        }));
    }

    /**
     * Displays a temporary error message in the file list area. The message
     * auto-removes after 5 seconds.
     *
     * @param {string} message - The error message text to display.
     * @private
     */
    _showError(message) {
        var err = document.createElement('div');
        err.className = 'text-sm text-[var(--color-danger)] py-1';
        err.textContent = message;
        this._fileList.appendChild(err);
        setTimeout(function() { if (err.parentNode) err.remove(); }, 5000);
    }

    /**
     * Checks whether a file matches the `accept` attribute specification.
     *
     * Supports three formats:
     * - Exact MIME type match (e.g., `"application/pdf"`)
     * - Wildcard MIME type match (e.g., `"image/*"` matches any `image/` type)
     * - File extension match (e.g., `".pdf"`, case-insensitive)
     *
     * @param {File} file - The file to check.
     * @param {string} accept - Comma-separated accept string (e.g., `"image/*,.pdf"`).
     * @returns {boolean} `true` if the file matches at least one accept pattern.
     * @private
     */
    _matchesAccept(file, accept) {
        var types = accept.split(',');
        for (var i = 0; i < types.length; i++) {
            var t = types[i].trim();
            if (!t) continue;
            // MIME wildcard (e.g. image/*)
            if (t.indexOf('/') !== -1) {
                if (t.endsWith('/*')) {
                    if (file.type.indexOf(t.replace('/*', '/')) === 0) return true;
                } else {
                    if (file.type === t) return true;
                }
            }
            // Extension (e.g. .pdf)
            if (t.charAt(0) === '.') {
                if (file.name.toLowerCase().endsWith(t.toLowerCase())) return true;
            }
        }
        return false;
    }

    /**
     * Formats a byte count into a human-readable size string.
     *
     * Uses binary units (KB = 1024 bytes). Integers are displayed without decimals;
     * values above 1 KB are shown with one decimal place.
     *
     * @param {number} bytes - The byte count to format.
     * @returns {string} Formatted size string (e.g., `"0 B"`, `"1.5 MB"`, `"256 KB"`).
     * @private
     *
     * @example
     * this._formatSize(0);        // "0 B"
     * this._formatSize(1024);     // "1.0 KB"
     * this._formatSize(5242880);  // "5.0 MB"
     */
    _formatSize(bytes) {
        if (bytes === 0) return '0 B';
        var units = ['B', 'KB', 'MB', 'GB'];
        var i = 0;
        var size = bytes;
        while (size >= 1024 && i < units.length - 1) {
            size /= 1024;
            i++;
        }
        return (i === 0 ? size : size.toFixed(1)) + ' ' + units[i];
    }
}

customElements.define('mcms-file-input', McmsFileInput);
