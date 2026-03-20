/**
 * @module mcms-field-renderer
 * @description Light DOM web component that renders form input widgets for different CMS field
 * types. Supports text, textarea, richtext (with markdown toolbar and preview), boolean, number,
 * date, select, media (with picker integration), reference (with content search dialog), and
 * plugin-provided custom fields.
 *
 * The richtext toolbar is configurable per-field via the `toolbar` attribute, globally via
 * `window.__mcmsRichtextToolbar`, or falls back to a built-in default set. Markdown preview
 * uses a lightweight built-in converter (no external dependencies).
 *
 * Plugin fields support two rendering modes: "inline" loads a web component from the plugin's
 * HTTP route, while "overlay" opens the plugin UI in an iframe dialog.
 *
 * All field types dispatch a `field-change` custom event when their value changes, allowing
 * parent forms to react to edits uniformly.
 *
 * @example <caption>Basic text field</caption>
 * <mcms-field-renderer type="text" name="title" value="Hello" label="Title"></mcms-field-renderer>
 *
 * @example <caption>Richtext field with custom toolbar</caption>
 * <mcms-field-renderer
 *   type="richtext"
 *   name="body"
 *   value="# Welcome"
 *   label="Body"
 *   toolbar='["bold","italic","link","preview"]'>
 * </mcms-field-renderer>
 *
 * @example <caption>Select field with choices</caption>
 * <mcms-field-renderer
 *   type="select"
 *   name="status"
 *   value="draft"
 *   label="Status"
 *   choices='[{"value":"draft","label":"Draft"},{"value":"published","label":"Published"}]'>
 * </mcms-field-renderer>
 *
 * @example <caption>Media field with thumbnail</caption>
 * <mcms-field-renderer
 *   type="media"
 *   name="hero_image"
 *   value="01HXYZ..."
 *   label="Hero Image"
 *   media-url="/uploads/hero.jpg"
 *   media-alt="Hero banner">
 * </mcms-field-renderer>
 *
 * @example <caption>Number field with constraints</caption>
 * <mcms-field-renderer type="number" name="count" value="5" min="0" max="100" step="1"></mcms-field-renderer>
 *
 * @example <caption>Plugin field (inline mode)</caption>
 * <mcms-field-renderer
 *   type="plugin"
 *   name="color"
 *   value="#ff0000"
 *   data-plugin-name="color-picker"
 *   data-plugin-interface="color-field"
 *   data-plugin-mode="inline">
 * </mcms-field-renderer>
 */

/**
 * Toolbar action definitions for richtext fields.
 *
 * Each entry maps an action name to its configuration object. Actions with `prefix`/`suffix`
 * insert markdown syntax around the selected text (or a placeholder). Actions with a `handler`
 * key use special logic (e.g., link insertion, preview toggle).
 *
 * @constant {Object.<string, ToolbarAction>}
 *
 * @typedef {Object} ToolbarAction
 * @property {string} label - Button text displayed in the toolbar.
 * @property {string} title - Tooltip text for the button.
 * @property {string} [prefix] - Markdown prefix inserted before selected text.
 * @property {string} [suffix] - Markdown suffix inserted after selected text.
 * @property {string} [placeholder] - Default text inserted when no text is selected.
 * @property {boolean} [line] - When true, the prefix is inserted at the beginning of the
 *   current line rather than wrapping the selection inline.
 * @property {string} [handler] - Special handler name. Supported values: "link" (opens link
 *   insertion logic), "preview" (toggles markdown preview mode).
 */
var TOOLBAR_ACTIONS = {
    bold:    { label: 'B',       title: 'Bold',           prefix: '**', suffix: '**', placeholder: 'bold text' },
    italic:  { label: 'I',       title: 'Italic',          prefix: '*',  suffix: '*',  placeholder: 'italic text' },
    h1:      { label: 'H1',      title: 'Heading 1',       prefix: '# ', suffix: '',   placeholder: 'Heading', line: true },
    h2:      { label: 'H2',      title: 'Heading 2',       prefix: '## ', suffix: '',  placeholder: 'Heading', line: true },
    h3:      { label: 'H3',      title: 'Heading 3',       prefix: '### ', suffix: '', placeholder: 'Heading', line: true },
    link:    { label: 'Link',    title: 'Insert Link',     handler: 'link' },
    ul:      { label: 'UL',      title: 'Unordered List',  prefix: '- ',  suffix: '',  placeholder: 'list item', line: true },
    ol:      { label: 'OL',      title: 'Ordered List',    prefix: '1. ', suffix: '',  placeholder: 'list item', line: true },
    preview: { label: 'Preview', title: 'Toggle Preview',  handler: 'preview' }
};

/**
 * Default toolbar action list used when no per-field `toolbar` attribute and no
 * `window.__mcmsRichtextToolbar` global configuration is provided.
 *
 * @constant {string[]}
 */
var TOOLBAR_FALLBACK = ['bold', 'italic', 'h1', 'h2', 'h3', 'link', 'ul', 'ol', 'preview'];

/**
 * Light DOM web component that renders the appropriate form input widget for a given CMS
 * field type. Each instance creates a labeled input element matching the specified `type`
 * attribute and dispatches `field-change` events on user edits.
 *
 * Supported field types:
 * - **text** -- Standard single-line text input (also the default for unknown types).
 * - **textarea** -- Multi-line text input with auto-resize.
 * - **richtext** -- Textarea with a configurable markdown toolbar and live preview toggle.
 * - **boolean** -- Checkbox with inline label layout.
 * - **number** -- Numeric input respecting `min`, `max`, and `step` attributes.
 * - **date** -- `datetime-local` input.
 * - **select** -- Dropdown populated from a JSON `choices` attribute.
 * - **media** -- Hidden input with thumbnail preview and `<mcms-media-picker>` integration.
 * - **reference** -- Hidden input with content search dialog via `<mcms-dialog>` and HTMX.
 * - **plugin** -- Plugin-provided custom field loaded as a web component or iframe.
 *
 * @extends HTMLElement
 * @fires McmsFieldRenderer#field-change
 *
 * @attr {string} type - The field type to render. One of: "text", "textarea", "richtext",
 *   "boolean", "number", "date", "select", "media", "reference", "plugin". Defaults to "text".
 * @attr {string} name - The field name, used as the input's `name` attribute and included in
 *   `field-change` event details.
 * @attr {string} value - The initial value of the field.
 * @attr {string} [label] - Display label for the field. Defaults to the `name` attribute value.
 * @attr {string} [toolbar] - JSON array of toolbar action names for richtext fields. Overrides
 *   `window.__mcmsRichtextToolbar` and the built-in fallback.
 * @attr {string} [min] - Minimum value for number fields. Passed through to the input element.
 * @attr {string} [max] - Maximum value for number fields. Passed through to the input element.
 * @attr {string} [step] - Step increment for number fields. Passed through to the input element.
 * @attr {string} [choices] - JSON array of `{value, label}` objects for select fields.
 * @attr {string} [media-url] - URL of the current media thumbnail for media fields.
 * @attr {string} [media-alt] - Alt text for the current media thumbnail.
 * @attr {string} [ref-title] - Display title for the currently selected reference content.
 * @attr {string} [data-plugin-name] - Plugin identifier for plugin fields.
 * @attr {string} [data-plugin-interface] - Plugin interface name for plugin fields.
 * @attr {string} [data-plugin-mode] - Rendering mode for plugin fields: "inline" (default)
 *   loads a web component, "overlay" opens an iframe dialog.
 *
 * @example <caption>Listening for field changes</caption>
 * document.querySelector('mcms-field-renderer')
 *   .addEventListener('field-change', function(e) {
 *     console.log('Field changed:', e.detail.name, '=', e.detail.value);
 *   });
 */
// Tailwind UI dark-mode input styles (from tailwind-ui/forms/input-groups).
var FIELD_INPUT_CLASS = 'block w-full rounded-md bg-white/5 px-3 py-1.5 text-base text-white outline-1 -outline-offset-1 outline-white/10 placeholder:text-gray-500 focus:outline-2 focus:-outline-offset-2 focus:outline-[var(--color-primary)] sm:text-sm/6';

// Tailwind UI dark-mode textarea styles (from tailwind-ui/forms/textareas).
var FIELD_TEXTAREA_CLASS = 'block w-full rounded-md bg-white/5 px-3 py-1.5 text-base text-white outline-1 -outline-offset-1 outline-white/10 placeholder:text-gray-500 focus:outline-2 focus:-outline-offset-2 focus:outline-[var(--color-primary)] sm:text-sm/6';

// Tailwind UI dark-mode select styles (from tailwind-ui/forms/select-menus).
var FIELD_SELECT_CLASS = 'w-full appearance-none rounded-md bg-white/5 py-1.5 pr-8 pl-3 text-base text-white outline-1 -outline-offset-1 outline-white/10 *:bg-gray-800 focus:outline-2 focus:-outline-offset-2 focus:outline-[var(--color-primary)] sm:text-sm/6';

// Tailwind UI dark-mode checkbox styles (from tailwind-ui/forms/checkboxes).
var FIELD_CHECKBOX_CLASS = 'size-4 appearance-none rounded-sm border border-white/10 bg-white/5 checked:border-[var(--color-primary)] checked:bg-[var(--color-primary)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--color-primary)] forced-colors:appearance-auto';

class McmsFieldRenderer extends HTMLElement {
    constructor() {
        super();
        /**
         * Whether the richtext preview mode is currently active.
         * @type {boolean}
         * @private
         */
        this._previewMode = false;
    }

    /**
     * Called when the element is inserted into the DOM. Reads configuration from attributes,
     * creates a wrapper container with a label, and delegates to the appropriate builder
     * method based on the `type` attribute. The built widget is appended to the element's
     * Light DOM.
     */
    connectedCallback() {
        // Guard: skip if already rendered (re-connection via HTMX swap/history restore)
        if (this.querySelector('.flex.flex-col.gap-1')) return;

        var type = this.getAttribute('type') || 'text';
        var name = this.getAttribute('name') || '';
        var value = this.getAttribute('value') || '';
        var label = this.getAttribute('label') || name;

        var wrapper = document.createElement('div');
        wrapper.className = 'flex flex-col gap-1';

        var labelEl = document.createElement('label');
        labelEl.className = 'block text-sm/6 font-medium text-white';
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
            case 'plugin':
                this._buildPluginField(wrapper, name, value);
                break;
            default:
                this._buildText(wrapper, name, value);
        }

        this.appendChild(wrapper);
    }

    /**
     * Builds a standard single-line text input and appends it to the wrapper.
     *
     * @param {HTMLDivElement} wrapper - The container element to append the input to.
     * @param {string} name - The field name, used for the input's `name` and `id` attributes.
     * @param {string} value - The initial value of the input.
     * @private
     */
    _buildText(wrapper, name, value) {
        var input = document.createElement('input');
        input.type = 'text';
        input.name = name;
        input.id = 'field-' + name;
        input.value = value;
        input.className = FIELD_INPUT_CLASS;
        wrapper.appendChild(input);
        this._attachChangeListener(input, name);
    }

    /**
     * Builds a multi-line textarea with auto-resize behavior and appends it to the wrapper.
     * The textarea automatically grows to fit its content as the user types.
     *
     * @param {HTMLDivElement} wrapper - The container element to append the textarea to.
     * @param {string} name - The field name, used for the textarea's `name` and `id` attributes.
     * @param {string} value - The initial text content of the textarea.
     * @private
     */
    _buildTextarea(wrapper, name, value) {
        var textarea = document.createElement('textarea');
        textarea.name = name;
        textarea.id = 'field-' + name;
        textarea.rows = 4;
        textarea.className = FIELD_TEXTAREA_CLASS;
        textarea.textContent = value;
        wrapper.appendChild(textarea);
        this._autoResize(textarea);
        this._attachChangeListener(textarea, name);
    }

    /**
     * Builds a richtext editing area with a configurable markdown toolbar and live preview
     * toggle. The toolbar action list is resolved in priority order:
     *   1. Per-field `toolbar` attribute (JSON array of action names)
     *   2. Global `window.__mcmsRichtextToolbar` array
     *   3. Built-in {@link TOOLBAR_FALLBACK} default
     *
     * Each toolbar button either wraps/inserts markdown syntax (via {@link _insertMarkdown})
     * or triggers a special handler (link insertion via {@link _insertLink}, preview toggle).
     *
     * The preview mode renders the textarea's markdown content as HTML using the built-in
     * {@link _markdownToHtml} converter and swaps the textarea for a read-only preview div.
     *
     * @param {HTMLDivElement} wrapper - The container element to append the richtext editor to.
     * @param {string} name - The field name, used for the textarea's `name` and `id` attributes.
     * @param {string} value - The initial markdown content.
     * @private
     */
    _buildRichtext(wrapper, name, value) {
        var self = this;

        // Resolve toolbar: per-field attribute > global config > hardcoded fallback
        var toolbarItems = null;
        var tbAttr = this.getAttribute('toolbar');
        if (tbAttr) {
            try { toolbarItems = JSON.parse(tbAttr); } catch (e) { toolbarItems = null; }
        }
        if (!toolbarItems && window.__mcmsRichtextToolbar) {
            toolbarItems = window.__mcmsRichtextToolbar;
        }
        if (!toolbarItems) {
            toolbarItems = TOOLBAR_FALLBACK;
        }

        // Toolbar container
        var toolbar = document.createElement('div');
        toolbar.className = 'flex flex-wrap gap-1 border-b border-[var(--color-border)] pb-2 mb-2';
        wrapper.appendChild(toolbar);

        // Textarea for editing
        var textarea = document.createElement('textarea');
        textarea.name = name;
        textarea.id = 'field-' + name;
        textarea.rows = 8;
        textarea.className = FIELD_TEXTAREA_CLASS + ' font-mono';
        textarea.textContent = value;
        wrapper.appendChild(textarea);

        // Preview container (hidden initially)
        var preview = document.createElement('div');
        preview.className = 'rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] prose prose-sm max-w-none';
        preview.style.display = 'none';
        wrapper.appendChild(preview);

        // Build toolbar buttons from resolved action list
        var previewBtn = null;
        for (var i = 0; i < toolbarItems.length; i++) {
            var actionName = toolbarItems[i];
            var action = TOOLBAR_ACTIONS[actionName];
            if (!action) continue;

            var btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'inline-flex items-center justify-center rounded-md px-2 py-1 text-xs font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] transition-colors cursor-pointer border-none';
            btn.setAttribute('data-action', actionName);
            btn.textContent = action.label;
            btn.title = action.title;
            toolbar.appendChild(btn);

            if (action.handler === 'preview') {
                previewBtn = btn;
            } else if (action.handler === 'link') {
                (function(a) {
                    btn.addEventListener('click', function() {
                        self._insertLink(textarea, name);
                    });
                })(action);
            } else {
                (function(a) {
                    btn.addEventListener('click', function() {
                        self._insertMarkdown(textarea, name, a.prefix, a.suffix, a.placeholder, a.line);
                    });
                })(action);
            }
        }

        // Preview toggle
        if (previewBtn) {
            previewBtn.addEventListener('click', function() {
                self._previewMode = !self._previewMode;
                if (self._previewMode) {
                    preview.innerHTML = self._markdownToHtml(textarea.value);
                    preview.style.display = '';
                    textarea.style.display = 'none';
                    previewBtn.textContent = 'Edit';
                } else {
                    preview.style.display = 'none';
                    textarea.style.display = '';
                    previewBtn.textContent = 'Preview';
                }
            });
        }

        this._autoResize(textarea);
        this._attachChangeListener(textarea, name);
    }

    /**
     * Inserts markdown formatting syntax into a textarea at the current cursor position or
     * around the current text selection. After insertion, the selection is adjusted to
     * highlight the meaningful content (not the syntax characters), and an `input` event
     * is dispatched to trigger change listeners.
     *
     * For line-start actions (e.g., headings, lists), the prefix is inserted at the
     * beginning of the line containing the cursor, not at the cursor position itself.
     *
     * @param {HTMLTextAreaElement} textarea - The textarea to insert markdown into.
     * @param {string} name - The field name (used for change event propagation).
     * @param {string} prefix - The markdown syntax to insert before the text (e.g., "**", "# ").
     * @param {string} suffix - The markdown syntax to insert after the text (e.g., "**", "").
     * @param {string} placeholder - Default text to insert when no text is selected.
     * @param {boolean} lineStart - When true, inserts the prefix at the beginning of the
     *   current line instead of at the cursor position.
     * @private
     */
    _insertMarkdown(textarea, name, prefix, suffix, placeholder, lineStart) {
        textarea.focus();
        var start = textarea.selectionStart;
        var end = textarea.selectionEnd;
        var text = textarea.value;
        var selected = text.substring(start, end);

        if (lineStart) {
            // Line-start actions: find the beginning of the current line
            var lineBegin = text.lastIndexOf('\n', start - 1) + 1;
            if (selected) {
                textarea.value = text.substring(0, lineBegin) + prefix + text.substring(lineBegin);
                textarea.selectionStart = start + prefix.length;
                textarea.selectionEnd = end + prefix.length;
            } else {
                var insert = prefix + placeholder;
                textarea.value = text.substring(0, lineBegin) + insert + text.substring(lineBegin);
                textarea.selectionStart = lineBegin + prefix.length;
                textarea.selectionEnd = lineBegin + insert.length;
            }
        } else {
            if (selected) {
                var wrapped = prefix + selected + suffix;
                textarea.value = text.substring(0, start) + wrapped + text.substring(end);
                textarea.selectionStart = start + prefix.length;
                textarea.selectionEnd = start + prefix.length + selected.length;
            } else {
                var insert = prefix + placeholder + suffix;
                textarea.value = text.substring(0, start) + insert + text.substring(end);
                textarea.selectionStart = start + prefix.length;
                textarea.selectionEnd = start + prefix.length + placeholder.length;
            }
        }

        textarea.dispatchEvent(new Event('input', { bubbles: true }));
    }

    /**
     * Inserts a markdown link at the current cursor position or wraps the selected text as
     * the link label. The cursor is positioned to select "url" so the user can immediately
     * type the destination URL.
     *
     * If text is selected, it becomes the link label: `[selected text](url)`.
     * If no text is selected, a placeholder is used: `[link text](url)`.
     *
     * @param {HTMLTextAreaElement} textarea - The textarea to insert the link into.
     * @param {string} name - The field name (used for change event propagation).
     * @private
     */
    _insertLink(textarea, name) {
        textarea.focus();
        var start = textarea.selectionStart;
        var end = textarea.selectionEnd;
        var text = textarea.value;
        var selected = text.substring(start, end);

        var linkText = selected || 'link text';
        var insert = '[' + linkText + '](url)';
        textarea.value = text.substring(0, start) + insert + text.substring(end);

        // Position cursor on "url"
        var urlStart = start + 1 + linkText.length + 2; // [linkText](
        textarea.selectionStart = urlStart;
        textarea.selectionEnd = urlStart + 3; // "url"

        textarea.dispatchEvent(new Event('input', { bubbles: true }));
    }

    /**
     * Builds a checkbox input with an inline label layout. The original block-level label
     * created by `connectedCallback` is replaced with a label positioned beside the checkbox
     * for a more natural boolean field appearance.
     *
     * The value "true" (string) is treated as checked; all other values are unchecked.
     * Change events emit the string "true" or "false".
     *
     * @param {HTMLDivElement} wrapper - The container element to append the checkbox to.
     * @param {string} name - The field name, used for the input's `name` and `id` attributes.
     * @param {string} value - The initial value; "true" sets the checkbox as checked.
     * @private
     */
    _buildBoolean(wrapper, name, value) {
        var container = document.createElement('div');
        container.className = 'flex items-center gap-2';

        // Tailwind UI checkbox pattern: grid overlay for custom check SVG
        var checkWrap = document.createElement('div');
        checkWrap.className = 'group grid size-4 grid-cols-1';

        var input = document.createElement('input');
        input.type = 'checkbox';
        input.name = name;
        input.id = 'field-' + name;
        input.checked = value === 'true';
        input.className = FIELD_CHECKBOX_CLASS + ' col-start-1 row-start-1';
        checkWrap.appendChild(input);

        var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svg.setAttribute('viewBox', '0 0 14 14');
        svg.setAttribute('fill', 'none');
        svg.setAttribute('class', 'pointer-events-none col-start-1 row-start-1 size-3.5 self-center justify-self-center stroke-white');
        var path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('d', 'M3 8L6 11L11 3.5');
        path.setAttribute('stroke-width', '2');
        path.setAttribute('stroke-linecap', 'round');
        path.setAttribute('stroke-linejoin', 'round');
        path.setAttribute('class', 'opacity-0 group-has-checked:opacity-100');
        svg.appendChild(path);
        checkWrap.appendChild(svg);

        container.appendChild(checkWrap);

        // Move the label into the container for inline layout
        var existingLabel = wrapper.querySelector('label');
        if (existingLabel) {
            existingLabel.removeAttribute('for');
            var inlineLabel = document.createElement('label');
            inlineLabel.setAttribute('for', 'field-' + name);
            inlineLabel.className = 'text-sm/6 font-medium text-white select-none';
            inlineLabel.textContent = existingLabel.textContent;
            container.appendChild(inlineLabel);
        }

        wrapper.appendChild(container);

        var self = this;
        input.addEventListener('change', function() {
            self._emitChange(name, String(input.checked));
        });
    }

    /**
     * Builds a numeric input that respects optional `min`, `max`, and `step` constraint
     * attributes from the host element. These attributes are read from the
     * `<mcms-field-renderer>` element and forwarded to the inner `<input type="number">`.
     *
     * @param {HTMLDivElement} wrapper - The container element to append the input to.
     * @param {string} name - The field name, used for the input's `name` and `id` attributes.
     * @param {string} value - The initial numeric value.
     * @private
     */
    _buildNumber(wrapper, name, value) {
        var input = document.createElement('input');
        input.type = 'number';
        input.name = name;
        input.id = 'field-' + name;
        input.value = value;
        input.className = FIELD_INPUT_CLASS;

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

    /**
     * Builds a datetime-local input for date/time selection.
     *
     * @param {HTMLDivElement} wrapper - The container element to append the input to.
     * @param {string} name - The field name, used for the input's `name` and `id` attributes.
     * @param {string} value - The initial datetime value in ISO format (e.g., "2026-03-16T12:00").
     * @private
     */
    _buildDate(wrapper, name, value) {
        var input = document.createElement('input');
        input.type = 'datetime-local';
        input.name = name;
        input.id = 'field-' + name;
        input.value = value;
        input.className = FIELD_INPUT_CLASS;
        wrapper.appendChild(input);
        this._attachChangeListener(input, name);
    }

    /**
     * Builds a `<select>` dropdown populated from the `choices` attribute. The `choices`
     * attribute must be a JSON array of objects with `value` and optional `label` properties.
     * The option matching the current `value` is automatically selected.
     *
     * If the `choices` attribute contains invalid JSON, the select is left empty.
     *
     * @param {HTMLDivElement} wrapper - The container element to append the select to.
     * @param {string} name - The field name, used for the select's `name` and `id` attributes.
     * @param {string} value - The initially selected option value.
     * @private
     */
    _buildSelect(wrapper, name, value) {
        var select = document.createElement('select');
        select.name = name;
        select.id = 'field-' + name;
        select.className = FIELD_SELECT_CLASS;

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

    /**
     * Builds a media field with a thumbnail preview area, a hidden input for the media ID,
     * and a "Choose" button that opens an `<mcms-media-picker>` component. If no picker
     * exists in the DOM, one is created and appended to `document.body`.
     *
     * The thumbnail preview displays one of three states:
     * - An `<img>` if `media-url` is set on the host element.
     * - A text ID placeholder if a `value` is set but no URL is available.
     * - "No media selected" if no value is set.
     *
     * When the user selects media from the picker, the hidden input and thumbnail are
     * updated, and a `field-change` event is emitted.
     *
     * @param {HTMLDivElement} wrapper - The container element to append the media field to.
     * @param {string} name - The field name, used for the hidden input's `name` and `id`.
     * @param {string} value - The initial media ID (ULID string).
     * @private
     */
    _buildMedia(wrapper, name, value) {
        var container = document.createElement('div');
        container.className = 'flex items-center gap-3';

        // Hidden input for storing media ID
        var hidden = document.createElement('input');
        hidden.type = 'hidden';
        hidden.name = name;
        hidden.id = 'field-' + name;
        hidden.value = value;
        container.appendChild(hidden);

        // Thumbnail preview area
        var thumbContainer = document.createElement('div');
        thumbContainer.className = 'w-24 h-24 shrink-0 rounded-md overflow-hidden bg-[var(--color-bg)] border border-[var(--color-border)] flex items-center justify-center';
        var mediaUrl = this.getAttribute('media-url') || '';
        var mediaAlt = this.getAttribute('media-alt') || '';
        if (mediaUrl) {
            var img = document.createElement('img');
            img.src = mediaUrl;
            img.alt = this._escapeText(mediaAlt);
            img.className = 'w-full h-full object-cover';
            thumbContainer.appendChild(img);
        } else if (value) {
            var placeholder = document.createElement('span');
            placeholder.className = 'text-[10px] leading-tight text-center text-[var(--color-text-muted)] break-all px-1.5';
            placeholder.textContent = value;
            thumbContainer.appendChild(placeholder);
        } else {
            var empty = document.createElement('span');
            empty.className = 'text-xs text-center text-[var(--color-text-dim)]';
            empty.textContent = 'No media';
            empty.style.lineHeight = '1.3';
            thumbContainer.appendChild(empty);
        }
        container.appendChild(thumbContainer);

        // Choose button
        var chooseBtn = document.createElement('button');
        chooseBtn.type = 'button';
        chooseBtn.className = 'inline-flex items-center justify-center rounded-md px-3 py-1.5 text-xs font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] transition-colors cursor-pointer border border-[var(--color-border)]';
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
                    newImg.className = 'w-full h-full object-cover';
                    thumbContainer.appendChild(newImg);
                } else if (detail.id) {
                    var idSpan = document.createElement('span');
                    idSpan.className = 'text-[10px] leading-tight text-center text-[var(--color-text-muted)] break-all px-1.5';
                    idSpan.textContent = detail.id;
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

    /**
     * Builds a content reference field with a hidden input for the content ID, a display
     * label showing the current selection, and a "Choose" button that opens an
     * `<mcms-dialog>` with a debounced search input.
     *
     * The search uses HTMX to fetch content results from `/admin/content?q=...&picker=true`.
     * Results are rendered as clickable items with `data-content-id` attributes. The user
     * selects an item (highlighted with both `data-selected` attribute and `selected` class),
     * then confirms via the dialog's confirm button.
     *
     * Search input is debounced at 300ms and requires a minimum of 2 characters.
     *
     * @param {HTMLDivElement} wrapper - The container element to append the reference field to.
     * @param {string} name - The field name, used for the hidden input's `name` and `id`.
     * @param {string} value - The initial content ID (ULID string).
     * @private
     */
    _buildReference(wrapper, name, value) {
        var container = document.createElement('div');
        container.className = 'flex items-center gap-3';

        // Hidden input for storing content ID
        var hidden = document.createElement('input');
        hidden.type = 'hidden';
        hidden.name = name;
        hidden.id = 'field-' + name;
        hidden.value = value;
        container.appendChild(hidden);

        // Label showing current value
        var refLabel = document.createElement('span');
        refLabel.className = 'text-sm text-[var(--color-text-muted)]';
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
        chooseBtn.className = 'inline-flex items-center justify-center rounded-md px-3 py-1.5 text-xs font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] transition-colors cursor-pointer border border-[var(--color-border)]';
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
            searchContainer.className = 'flex flex-col gap-3';

            var searchInput = document.createElement('input');
            searchInput.type = 'text';
            searchInput.placeholder = 'Search content...';
            searchInput.className = 'w-full rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]';
            searchContainer.appendChild(searchInput);

            var resultsList = document.createElement('div');
            resultsList.className = 'max-h-60 overflow-y-auto rounded-md border border-[var(--color-border)] bg-[var(--color-bg)]';
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
                    allItems[i].removeAttribute('data-selected');
                    allItems[i].classList.remove('selected'); // DUAL: data-selected + class
                }
                item.setAttribute('data-selected', '');
                item.classList.add('selected'); // DUAL: data-selected + class
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

    /**
     * Attaches auto-resize behavior to a textarea element. On each `input` event, the
     * textarea's height is reset to `auto` and then set to its `scrollHeight`, causing
     * it to grow (and shrink) to fit its content. An initial resize is performed on the
     * next animation frame to handle pre-populated values.
     *
     * @param {HTMLTextAreaElement} textarea - The textarea to attach auto-resize behavior to.
     * @private
     */
    _autoResize(textarea) {
        var resize = function() {
            textarea.style.height = 'auto';
            textarea.style.height = textarea.scrollHeight + 'px';
        };
        textarea.addEventListener('input', resize);
        // Initial resize after DOM render
        requestAnimationFrame(resize);
    }

    /**
     * Attaches an `input` event listener to a form element that dispatches a `field-change`
     * custom event on the host element whenever the input value changes. For checkbox inputs,
     * the emitted value is the string "true" or "false"; for all others, it is the input's
     * string value.
     *
     * @param {HTMLInputElement|HTMLTextAreaElement|HTMLSelectElement} input - The form element
     *   to listen on.
     * @param {string} name - The field name included in the `field-change` event detail.
     * @private
     */
    _attachChangeListener(input, name) {
        var self = this;
        input.addEventListener('input', function() {
            var val = input.type === 'checkbox' ? String(input.checked) : input.value;
            self._emitChange(name, val);
        });
    }

    /**
     * Dispatches a `field-change` custom event on this element. The event bubbles up the
     * DOM tree so parent forms and containers can listen for changes from any nested field
     * renderer.
     *
     * @param {string} name - The field name that changed.
     * @param {string} value - The new field value.
     * @private
     */
    _emitChange(name, value) {
        this.dispatchEvent(new CustomEvent('field-change', {
            bubbles: true,
            detail: { name: name, value: value }
        }));
    }

    /**
     * Escapes a plain text string for safe insertion into HTML. Uses a temporary DOM element
     * to leverage the browser's built-in escaping of `<`, `>`, `&`, and `"` characters.
     *
     * @param {string} str - The plain text string to escape.
     * @returns {string} The HTML-escaped string.
     * @private
     */
    _escapeText(str) {
        var div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    /**
     * Converts a markdown string to HTML using a lightweight built-in parser. This converter
     * has no external dependencies and supports a subset of markdown commonly used in CMS
     * content editing.
     *
     * Supported syntax:
     * - **Headers:** `#`, `##`, `###` at the start of a line.
     * - **Bold:** `**text**` rendered as `<strong>`.
     * - **Italic:** `*text*` rendered as `<em>`.
     * - **Links:** `[text](url)` rendered as `<a>` with `target="_blank"` and `rel="noopener"`.
     * - **Unordered lists:** Lines starting with `- ` or `* ` grouped into `<ul>`.
     * - **Ordered lists:** Lines starting with `1. ` (any digit) grouped into `<ol>`.
     * - **Paragraphs:** Non-empty lines not matching other patterns wrapped in `<p>`.
     *
     * All text content is HTML-escaped via {@link _escapeText} before inline markdown
     * processing to prevent XSS.
     *
     * @param {string} md - The raw markdown string to convert.
     * @returns {string} The resulting HTML string.
     * @private
     */
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

    /**
     * Processes inline markdown formatting within an already HTML-escaped string. Converts
     * bold (`**text**`), italic (`*text*`), and link (`[text](url)`) syntax to their HTML
     * equivalents.
     *
     * Bold is processed before italic to prevent `**text**` from being partially consumed
     * by the italic pattern. Links open in a new tab with `rel="noopener"` for security.
     *
     * Note: The input is expected to be pre-escaped via {@link _escapeText}. The characters
     * `[`, `]`, `(`, `)`, and `*` are not affected by the escaping (only `<`, `>`, `&`, `"`
     * are escaped), so markdown syntax is preserved through the escaping step.
     *
     * @param {string} text - An HTML-escaped string containing inline markdown syntax.
     * @returns {string} The string with inline markdown converted to HTML tags.
     * @private
     */
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

    /**
     * Builds a plugin-provided custom field. Plugin fields delegate rendering to an external
     * plugin identified by `data-plugin-name` and `data-plugin-interface` attributes.
     *
     * Two rendering modes are supported:
     *
     * **Inline mode** (default): Dynamically loads a `<script>` from the plugin's HTTP route
     * (`/api/v1/plugins/{name}/interface/{interface}/component.js`), then creates a custom
     * element with the tag `plugin-{name}-{interface}`. The plugin component receives `value`
     * and `name` attributes and is expected to dispatch `field-change` events when its value
     * changes. If the script fails to load, a plain text input fallback is shown.
     *
     * **Overlay mode**: Shows the current value as text with an "Edit" button. Clicking the
     * button opens an `<mcms-dialog>` containing an iframe pointing to the plugin's interface
     * URL (`/api/v1/plugins/{name}/interface/{interface}/`). The plugin UI communicates
     * value changes back via `postMessage`.
     *
     * If `data-plugin-name` or `data-plugin-interface` are missing, a raw text input fallback
     * is rendered with a "(plugin field -- raw value)" placeholder.
     *
     * @param {HTMLDivElement} wrapper - The container element to append the plugin field to.
     * @param {string} name - The field name, used for the hidden input's `name` and `id`.
     * @param {string} value - The initial field value (opaque string managed by the plugin).
     * @private
     */
    _buildPluginField(wrapper, name, value) {
        var pluginName = this.getAttribute('data-plugin-name') || '';
        var pluginInterface = this.getAttribute('data-plugin-interface') || '';
        var mode = this.getAttribute('data-plugin-mode') || 'inline';

        var container = document.createElement('div');
        container.className = 'flex items-center gap-3';

        // Hidden input for the field value (opaque string).
        var hidden = document.createElement('input');
        hidden.type = 'hidden';
        hidden.name = name;
        hidden.id = 'field-' + name;
        hidden.value = value;
        container.appendChild(hidden);

        if (!pluginName || !pluginInterface) {
            // No plugin config — show raw text input as fallback.
            var fallback = document.createElement('input');
            fallback.type = 'text';
            fallback.name = name;
            fallback.id = 'field-' + name;
            fallback.value = value;
            fallback.placeholder = '(plugin field — raw value)';
            wrapper.appendChild(fallback);
            this._attachChangeListener(fallback, name);
            return;
        }

        if (mode === 'inline') {
            // Inline mode: load web component from plugin HTTP route.
            var componentUrl = '/api/v1/plugins/' + encodeURIComponent(pluginName) + '/interface/' + encodeURIComponent(pluginInterface) + '/component.js';
            var componentTag = 'plugin-' + pluginName + '-' + pluginInterface;

            // Attempt to load the plugin component script dynamically.
            var script = document.createElement('script');
            script.src = componentUrl;
            script.onerror = function() {
                // Fallback: show text input if component not available.
                var fallbackInput = document.createElement('input');
                fallbackInput.type = 'text';
                fallbackInput.value = value;
                fallbackInput.placeholder = '(plugin editor unavailable)';
                container.appendChild(fallbackInput);
                var self2 = this;
                fallbackInput.addEventListener('input', function() {
                    hidden.value = fallbackInput.value;
                    self2.dispatchEvent(new CustomEvent('field-change', {
                        bubbles: true,
                        detail: { name: name, value: fallbackInput.value }
                    }));
                }.bind(self2));
            }.bind(this);
            script.onload = function() {
                var el = document.createElement(componentTag);
                el.setAttribute('value', value);
                el.setAttribute('name', name);
                container.appendChild(el);
                // Listen for field-change from the plugin component.
                el.addEventListener('field-change', function(e) {
                    hidden.value = e.detail.value;
                    this.dispatchEvent(new CustomEvent('field-change', {
                        bubbles: true,
                        detail: { name: name, value: e.detail.value }
                    }));
                }.bind(this));
            }.bind(this);
            document.head.appendChild(script);
        } else {
            // Overlay mode: show button with current value, open modal on click.
            var display = document.createElement('span');
            display.className = 'text-sm text-[var(--color-text-muted)]';
            display.textContent = value || '(not set)';
            container.appendChild(display);

            var editBtn = document.createElement('button');
            editBtn.type = 'button';
            editBtn.textContent = 'Edit';
            editBtn.className = 'inline-flex items-center justify-center rounded-md px-3 py-1.5 text-xs font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] transition-colors cursor-pointer border border-[var(--color-border)]';
            container.appendChild(editBtn);

            var self = this;
            editBtn.addEventListener('click', function() {
                var dialog = document.createElement('mcms-dialog');
                dialog.setAttribute('title', pluginInterface);
                dialog.setAttribute('confirm-label', 'Done');
                dialog.setAttribute('cancel-label', 'Cancel');

                var iframe = document.createElement('iframe');
                iframe.src = '/api/v1/plugins/' + encodeURIComponent(pluginName) + '/interface/' + encodeURIComponent(pluginInterface) + '/';
                iframe.style.width = '100%';
                iframe.style.height = '400px';
                iframe.style.border = 'none';
                dialog.appendChild(iframe);
                document.body.appendChild(dialog);

                dialog.addEventListener('confirm', function() {
                    // The plugin UI communicates via postMessage.
                    dialog.remove();
                });
                dialog.addEventListener('cancel', function() {
                    dialog.remove();
                });
            });
        }

        wrapper.appendChild(container);
    }
}

/**
 * Dispatched when any field's value changes. Bubbles up from the `<mcms-field-renderer>`
 * element through the DOM tree.
 *
 * @event McmsFieldRenderer#field-change
 * @type {CustomEvent}
 * @property {Object} detail - The event detail object.
 * @property {string} detail.name - The name of the field that changed.
 * @property {string} detail.value - The new value of the field. For boolean fields, this is
 *   the string "true" or "false". For all other types, it is the input's string value.
 */

customElements.define('mcms-field-renderer', McmsFieldRenderer);
