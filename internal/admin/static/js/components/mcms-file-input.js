// <mcms-file-input> -- Styled file input with drag-and-drop (Light DOM)
// Attributes: name, accept, multiple, required, label, max-size (bytes)
// Dispatches: file-change custom event with { name, files }
class McmsFileInput extends HTMLElement {
    constructor() {
        super();
        this._files = [];
        this._input = null;
        this._dropZone = null;
        this._fileList = null;
    }

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

    _showError(message) {
        var err = document.createElement('div');
        err.className = 'text-sm text-[var(--color-danger)] py-1';
        err.textContent = message;
        this._fileList.appendChild(err);
        setTimeout(function() { if (err.parentNode) err.remove(); }, 5000);
    }

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
