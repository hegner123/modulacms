// <mcms-field-renderer> -- Field type input widgets (Light DOM)
// Supports types: text, textarea, richtext, boolean, number, date, select, media, reference
// Dispatches: field-change custom event with {name, value}
class McmsFieldRenderer extends HTMLElement {
    constructor() {
        super();
        this._previewMode = false;
    }

    connectedCallback() {
        var type = this.getAttribute('type') || 'text';
        var name = this.getAttribute('name') || '';
        var value = this.getAttribute('value') || '';
        var label = this.getAttribute('label') || name;

        var wrapper = document.createElement('div');
        wrapper.className = 'form-field field-renderer-wrapper';

        var labelEl = document.createElement('label');
        labelEl.textContent = label;
        labelEl.setAttribute('for', 'field-' + name);
        wrapper.appendChild(labelEl);

        switch (type) {
            case 'richtext':
                this._buildRichtext(wrapper, name, value);
                break;
            case 'textarea':
                this._buildTextarea(wrapper, name, value);
                break;
            case 'boolean':
                this._buildBoolean(wrapper, name, value);
                break;
            case 'number':
                this._buildNumber(wrapper, name, value);
                break;
            case 'date':
                this._buildDate(wrapper, name, value);
                break;
            case 'select':
                this._buildSelect(wrapper, name, value);
                break;
            case 'media':
                this._buildMedia(wrapper, name, value);
                break;
            case 'reference':
                this._buildReference(wrapper, name, value);
                break;
            default:
                this._buildText(wrapper, name, value);
        }

        this.appendChild(wrapper);
    }

    _buildText(wrapper, name, value) {
        var input = document.createElement('input');
        input.type = 'text';
        input.name = name;
        input.id = 'field-' + name;
        input.value = value;
        wrapper.appendChild(input);
        this._attachChangeListener(input, name);
    }

    _buildTextarea(wrapper, name, value) {
        var textarea = document.createElement('textarea');
        textarea.name = name;
        textarea.id = 'field-' + name;
        textarea.rows = 4;
        textarea.textContent = value;
        wrapper.appendChild(textarea);
        this._autoResize(textarea);
        this._attachChangeListener(textarea, name);
    }

    _buildRichtext(wrapper, name, value) {
        var self = this;

        // Toolbar with preview toggle
        var toolbar = document.createElement('div');
        toolbar.className = 'field-renderer-richtext-toolbar';

        var toggleBtn = document.createElement('button');
        toggleBtn.type = 'button';
        toggleBtn.className = 'btn btn-sm btn-ghost';
        toggleBtn.textContent = 'Preview';
        toolbar.appendChild(toggleBtn);
        wrapper.appendChild(toolbar);

        // Textarea for editing
        var textarea = document.createElement('textarea');
        textarea.name = name;
        textarea.id = 'field-' + name;
        textarea.rows = 8;
        textarea.className = 'field-renderer-richtext-input';
        textarea.textContent = value;
        wrapper.appendChild(textarea);

        // Preview container (hidden initially)
        var preview = document.createElement('div');
        preview.className = 'field-renderer-richtext-preview';
        preview.style.display = 'none';
        wrapper.appendChild(preview);

        toggleBtn.addEventListener('click', function() {
            self._previewMode = !self._previewMode;
            if (self._previewMode) {
                preview.innerHTML = self._markdownToHtml(textarea.value);
                preview.style.display = '';
                textarea.style.display = 'none';
                toggleBtn.textContent = 'Edit';
            } else {
                preview.style.display = 'none';
                textarea.style.display = '';
                toggleBtn.textContent = 'Preview';
            }
        });

        this._autoResize(textarea);
        this._attachChangeListener(textarea, name);
    }

    _buildBoolean(wrapper, name, value) {
        var container = document.createElement('div');
        container.className = 'field-renderer-boolean';

        var input = document.createElement('input');
        input.type = 'checkbox';
        input.name = name;
        input.id = 'field-' + name;
        input.checked = value === 'true';
        container.appendChild(input);

        // Move the label into the container for inline layout
        var existingLabel = wrapper.querySelector('label');
        if (existingLabel) {
            existingLabel.removeAttribute('for');
            var inlineLabel = document.createElement('label');
            inlineLabel.setAttribute('for', 'field-' + name);
            inlineLabel.textContent = existingLabel.textContent;
            container.appendChild(inlineLabel);
        }

        wrapper.appendChild(container);

        var self = this;
        input.addEventListener('change', function() {
            self._emitChange(name, String(input.checked));
        });
    }

    _buildNumber(wrapper, name, value) {
        var input = document.createElement('input');
        input.type = 'number';
        input.name = name;
        input.id = 'field-' + name;
        input.value = value;

        // Respect min, max, step from host element attributes
        var min = this.getAttribute('min');
        var max = this.getAttribute('max');
        var step = this.getAttribute('step');
        if (min !== null) input.min = min;
        if (max !== null) input.max = max;
        if (step !== null) input.step = step;

        wrapper.appendChild(input);
        this._attachChangeListener(input, name);
    }

    _buildDate(wrapper, name, value) {
        var input = document.createElement('input');
        input.type = 'datetime-local';
        input.name = name;
        input.id = 'field-' + name;
        input.value = value;
        wrapper.appendChild(input);
        this._attachChangeListener(input, name);
    }

    _buildSelect(wrapper, name, value) {
        var select = document.createElement('select');
        select.name = name;
        select.id = 'field-' + name;

        // Parse choices from attribute
        var choicesAttr = this.getAttribute('choices');
        if (choicesAttr) {
            try {
                var choices = JSON.parse(choicesAttr);
                if (Array.isArray(choices)) {
                    for (var i = 0; i < choices.length; i++) {
                        var option = document.createElement('option');
                        option.value = choices[i].value || '';
                        option.textContent = choices[i].label || choices[i].value || '';
                        if (option.value === value) {
                            option.selected = true;
                        }
                        select.appendChild(option);
                    }
                }
            } catch (err) {
                // Invalid JSON, leave select empty
            }
        }

        wrapper.appendChild(select);
        this._attachChangeListener(select, name);
    }

    _buildMedia(wrapper, name, value) {
        var container = document.createElement('div');
        container.className = 'field-renderer-media';

        // Hidden input for storing media ID
        var hidden = document.createElement('input');
        hidden.type = 'hidden';
        hidden.name = name;
        hidden.id = 'field-' + name;
        hidden.value = value;
        container.appendChild(hidden);

        // Thumbnail preview area
        var thumbContainer = document.createElement('div');
        thumbContainer.className = 'field-renderer-media-thumb';
        var mediaUrl = this.getAttribute('media-url') || '';
        var mediaAlt = this.getAttribute('media-alt') || '';
        if (mediaUrl) {
            var img = document.createElement('img');
            img.src = mediaUrl;
            img.alt = this._escapeText(mediaAlt);
            thumbContainer.appendChild(img);
        } else if (value) {
            var placeholder = document.createElement('span');
            placeholder.className = 'field-renderer-media-id';
            placeholder.textContent = 'Media: ' + value;
            thumbContainer.appendChild(placeholder);
        } else {
            var empty = document.createElement('span');
            empty.className = 'field-renderer-media-empty';
            empty.textContent = 'No media selected';
            thumbContainer.appendChild(empty);
        }
        container.appendChild(thumbContainer);

        // Choose button
        var chooseBtn = document.createElement('button');
        chooseBtn.type = 'button';
        chooseBtn.className = 'btn btn-sm';
        chooseBtn.textContent = 'Choose';
        container.appendChild(chooseBtn);

        wrapper.appendChild(container);

        var self = this;
        chooseBtn.addEventListener('click', function() {
            // Find or create a media picker
            var picker = document.querySelector('mcms-media-picker');
            if (!picker) {
                picker = document.createElement('mcms-media-picker');
                document.body.appendChild(picker);
            }

            // Listen for selection
            var onSelected = function(e) {
                picker.removeEventListener('media-selected', onSelected);
                picker.removeEventListener('media-picker:cancel', onCancel);

                var detail = e.detail || {};
                hidden.value = detail.id || '';

                // Update thumbnail
                thumbContainer.textContent = '';
                if (detail.url) {
                    var newImg = document.createElement('img');
                    newImg.src = detail.url;
                    newImg.alt = self._escapeText(detail.alt || '');
                    thumbContainer.appendChild(newImg);
                } else if (detail.id) {
                    var idSpan = document.createElement('span');
                    idSpan.className = 'field-renderer-media-id';
                    idSpan.textContent = 'Media: ' + detail.id;
                    thumbContainer.appendChild(idSpan);
                }

                self._emitChange(name, hidden.value);
            };

            var onCancel = function() {
                picker.removeEventListener('media-selected', onSelected);
                picker.removeEventListener('media-picker:cancel', onCancel);
            };

            picker.addEventListener('media-selected', onSelected);
            picker.addEventListener('media-picker:cancel', onCancel);
            picker.setAttribute('open', '');
        });
    }

    _buildReference(wrapper, name, value) {
        var container = document.createElement('div');
        container.className = 'field-renderer-reference';

        // Hidden input for storing content ID
        var hidden = document.createElement('input');
        hidden.type = 'hidden';
        hidden.name = name;
        hidden.id = 'field-' + name;
        hidden.value = value;
        container.appendChild(hidden);

        // Label showing current value
        var refLabel = document.createElement('span');
        refLabel.className = 'field-renderer-reference-label';
        var refTitle = this.getAttribute('ref-title') || '';
        if (refTitle) {
            refLabel.textContent = refTitle;
        } else if (value) {
            refLabel.textContent = 'Content: ' + value;
        } else {
            refLabel.textContent = 'No content selected';
        }
        container.appendChild(refLabel);

        // Choose button
        var chooseBtn = document.createElement('button');
        chooseBtn.type = 'button';
        chooseBtn.className = 'btn btn-sm';
        chooseBtn.textContent = 'Choose';
        container.appendChild(chooseBtn);

        wrapper.appendChild(container);

        var self = this;
        chooseBtn.addEventListener('click', function() {
            // Build a minimal content search dialog
            var dialog = document.createElement('mcms-dialog');
            dialog.setAttribute('title', 'Select Content');
            dialog.setAttribute('confirm-label', 'Select');
            dialog.setAttribute('cancel-label', 'Cancel');

            var searchContainer = document.createElement('div');
            searchContainer.className = 'field-renderer-reference-search';

            var searchInput = document.createElement('input');
            searchInput.type = 'text';
            searchInput.placeholder = 'Search content...';
            searchInput.className = 'field-renderer-reference-search-input';
            searchContainer.appendChild(searchInput);

            var resultsList = document.createElement('div');
            resultsList.className = 'field-renderer-reference-results';
            searchContainer.appendChild(resultsList);

            dialog.appendChild(searchContainer);
            document.body.appendChild(dialog);

            var selectedContentId = '';
            var selectedTitle = '';

            // Search via HTMX or basic fetch
            var debounceTimer = null;
            searchInput.addEventListener('input', function() {
                clearTimeout(debounceTimer);
                debounceTimer = setTimeout(function() {
                    var query = searchInput.value.trim();
                    if (query.length < 2) {
                        resultsList.textContent = '';
                        return;
                    }
                    if (typeof htmx !== 'undefined') {
                        htmx.ajax('GET', '/admin/content?q=' + encodeURIComponent(query) + '&picker=true', {
                            target: resultsList,
                            swap: 'innerHTML'
                        });
                    }
                }, 300);
            });

            resultsList.addEventListener('click', function(e) {
                var item = e.target.closest('[data-content-id]');
                if (!item) return;

                // Deselect others
                var allItems = resultsList.querySelectorAll('[data-content-id]');
                for (var i = 0; i < allItems.length; i++) {
                    allItems[i].classList.remove('selected');
                }
                item.classList.add('selected');
                selectedContentId = item.getAttribute('data-content-id');
                selectedTitle = item.textContent.trim();
            });

            dialog.addEventListener('mcms-dialog:confirm', function() {
                if (selectedContentId) {
                    hidden.value = selectedContentId;
                    refLabel.textContent = selectedTitle || 'Content: ' + selectedContentId;
                    self._emitChange(name, selectedContentId);
                }
                dialog.remove();
            });

            dialog.addEventListener('mcms-dialog:cancel', function() {
                dialog.remove();
            });

            dialog.setAttribute('open', '');
        });
    }

    _autoResize(textarea) {
        var resize = function() {
            textarea.style.height = 'auto';
            textarea.style.height = textarea.scrollHeight + 'px';
        };
        textarea.addEventListener('input', resize);
        // Initial resize after DOM render
        requestAnimationFrame(resize);
    }

    _attachChangeListener(input, name) {
        var self = this;
        input.addEventListener('input', function() {
            var val = input.type === 'checkbox' ? String(input.checked) : input.value;
            self._emitChange(name, val);
        });
    }

    _emitChange(name, value) {
        this.dispatchEvent(new CustomEvent('field-change', {
            bubbles: true,
            detail: { name: name, value: value }
        }));
    }

    _escapeText(str) {
        var div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    // Simple markdown-to-HTML converter (no external deps)
    // Supports: headers (# ## ###), bold (**), italic (*), links [text](url),
    // unordered lists (- or *), ordered lists (1.), paragraphs
    _markdownToHtml(md) {
        var lines = md.split('\n');
        var html = [];
        var inUl = false;
        var inOl = false;

        for (var i = 0; i < lines.length; i++) {
            var line = lines[i];

            // Close lists if the line is not a list item
            var isUnorderedItem = /^[\s]*[-*]\s+/.test(line);
            var isOrderedItem = /^[\s]*\d+\.\s+/.test(line);

            if (!isUnorderedItem && inUl) {
                html.push('</ul>');
                inUl = false;
            }
            if (!isOrderedItem && inOl) {
                html.push('</ol>');
                inOl = false;
            }

            // Headers
            if (line.match(/^### /)) {
                html.push('<h3>' + this._inlineMarkdown(this._escapeText(line.substring(4))) + '</h3>');
                continue;
            }
            if (line.match(/^## /)) {
                html.push('<h2>' + this._inlineMarkdown(this._escapeText(line.substring(3))) + '</h2>');
                continue;
            }
            if (line.match(/^# /)) {
                html.push('<h1>' + this._inlineMarkdown(this._escapeText(line.substring(2))) + '</h1>');
                continue;
            }

            // Unordered list
            if (isUnorderedItem) {
                if (!inUl) {
                    html.push('<ul>');
                    inUl = true;
                }
                var content = line.replace(/^[\s]*[-*]\s+/, '');
                html.push('<li>' + this._inlineMarkdown(this._escapeText(content)) + '</li>');
                continue;
            }

            // Ordered list
            if (isOrderedItem) {
                if (!inOl) {
                    html.push('<ol>');
                    inOl = true;
                }
                var olContent = line.replace(/^[\s]*\d+\.\s+/, '');
                html.push('<li>' + this._inlineMarkdown(this._escapeText(olContent)) + '</li>');
                continue;
            }

            // Empty line
            if (line.trim() === '') {
                html.push('');
                continue;
            }

            // Paragraph
            html.push('<p>' + this._inlineMarkdown(this._escapeText(line)) + '</p>');
        }

        // Close any open lists
        if (inUl) html.push('</ul>');
        if (inOl) html.push('</ol>');

        return html.join('\n');
    }

    // Process inline markdown: bold, italic, links
    // Input is already HTML-escaped, so we work with the escaped form
    _inlineMarkdown(text) {
        // Bold: **text** (escaped: we match literal ** pairs)
        text = text.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        // Italic: *text* (single asterisk, not already consumed by bold)
        text = text.replace(/\*(.+?)\*/g, '<em>$1</em>');
        // Links: [text](url) -- need to handle HTML-escaped brackets
        // After _escapeText, [] and () are not escaped, only <, >, &, " are
        text = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>');
        return text;
    }
}

customElements.define('mcms-field-renderer', McmsFieldRenderer);
