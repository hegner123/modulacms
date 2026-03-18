/**
 * @module mcms-dialog
 * @description Modal dialog wrapper (Light DOM) for the ModulaCMS admin panel.
 *
 * Drop-in replacement for the native `<dialog>` element. Wraps its children
 * in a native `<dialog>` with a semi-transparent backdrop, auto-centering,
 * keyboard focus trapping (Tab / Shift+Tab), and Escape-to-close.
 *
 * The component does NOT restructure children -- content (title, form, buttons)
 * is authored in templ and preserved as-is inside the created `<dialog>`.
 *
 * After building its internal structure, the component calls `htmx.process()`
 * on the dialog so that any HTMX attributes on moved children are re-bound.
 * It also listens for `htmx:afterRequest` on `document.body` and auto-closes
 * the dialog when a successful HTMX request originates from an element inside it.
 *
 * @example <caption>Basic dialog with server-rendered content</caption>
 * <mcms-dialog id="create-dialog" aria-labelledby="dialog-title">
 *   <div class="p-6">
 *     <h2 id="dialog-title">Create Item</h2>
 *     <form hx-post="/admin/items" hx-target="#item-list">
 *       <input name="name" type="text" />
 *       <button type="submit">Save</button>
 *     </form>
 *   </div>
 * </mcms-dialog>
 *
 * @example <caption>Opening and closing programmatically</caption>
 * document.getElementById('create-dialog').open();
 * document.getElementById('create-dialog').close();
 *
 * @example <caption>Auto-open on page load</caption>
 * <mcms-dialog id="welcome" open>
 *   <div class="p-6">Welcome back!</div>
 * </mcms-dialog>
 */

/**
 * @event mcms-dialog:open
 * @type {CustomEvent}
 * @description Fired on the `<mcms-dialog>` element when the dialog is opened
 * via the {@link McmsDialog#open} method. Bubbles up the DOM tree.
 *
 * @example
 * dialog.addEventListener('mcms-dialog:open', (e) => {
 *   console.log('Dialog opened:', e.target.id);
 * });
 */

/**
 * @event mcms-dialog:close
 * @type {CustomEvent}
 * @description Fired on the `<mcms-dialog>` element when the dialog is closed
 * via the {@link McmsDialog#close} method. Bubbles up the DOM tree.
 *
 * @example
 * dialog.addEventListener('mcms-dialog:close', (e) => {
 *   console.log('Dialog closed:', e.target.id);
 * });
 */

/**
 * Modal dialog custom element that wraps its children in a native `<dialog>`.
 *
 * Provides backdrop click-to-close, Escape key handling, Tab focus trapping,
 * HTMX integration (auto-close on successful request), and accessibility
 * attributes (aria-labelledby forwarding).
 *
 * @extends HTMLElement
 * @fires mcms-dialog:open When the dialog is opened
 * @fires mcms-dialog:close When the dialog is closed
 *
 * @attr {string} [id] - Element ID, used for programmatic access and debug logging.
 * @attr {boolean} [open] - When present at connection time, the dialog auto-opens.
 *   This is a one-time check in `connectedCallback`, not a reactive attribute.
 * @attr {string} [aria-labelledby] - Forwarded to the internal `<dialog>` element
 *   to associate the dialog with its title for screen readers.
 *
 * @example
 * // Open from a button click handler
 * <button onclick="document.getElementById('my-dialog').open()">Open</button>
 *
 * <mcms-dialog id="my-dialog">
 *   <div class="p-6">
 *     <h2>Dialog Title</h2>
 *     <p>Dialog content here.</p>
 *     <button onclick="document.getElementById('my-dialog').close()">Cancel</button>
 *   </div>
 * </mcms-dialog>
 */
class McmsDialog extends HTMLElement {
    constructor() {
        super();

        /**
         * Reference to the internal native `<dialog>` element created during build.
         * @type {HTMLDialogElement|null}
         * @private
         */
        this._dialog = null;

        /**
         * Tracks whether {@link McmsDialog#_build} has been called to prevent
         * duplicate builds if the element is disconnected and reconnected.
         * @type {boolean}
         * @private
         */
        this._built = false;

        /**
         * Bound reference to {@link McmsDialog#_onKeyDown} for adding and removing
         * the document-level keydown listener without losing the `this` context.
         * @type {function(KeyboardEvent): void}
         * @private
         */
        this._boundKeyDown = this._onKeyDown.bind(this);

        /**
         * Bound reference to {@link McmsDialog#_onBackdropClick} for the dialog
         * click listener, used to detect clicks on the backdrop pseudo-element.
         * @type {function(MouseEvent): void}
         * @private
         */
        this._boundBackdropClick = this._onBackdropClick.bind(this);

        /**
         * CSS selector string matching all focusable elements within the dialog.
         * Used by the Tab-key focus trap in {@link McmsDialog#_onKeyDown}.
         * @type {string}
         * @private
         */
        this._focusableSelector = 'button, [href], input:not([type="hidden"]), select, textarea, [tabindex]:not([tabindex="-1"])';
    }

    /**
     * Called when the element is inserted into the DOM. Triggers the one-time
     * build of the internal `<dialog>` element and auto-opens the dialog if
     * the host element has the `open` attribute.
     * @returns {void}
     */
    connectedCallback() {
        if (!this._built) this._build();
        // Auto-open if the host element has the open attribute
        if (this.hasAttribute('open')) {
            this.open();
        }
    }

    /**
     * Called when the element is removed from the DOM. Cleans up the
     * document-level keydown listener to prevent memory leaks and stale
     * event handling.
     * @returns {void}
     */
    disconnectedCallback() {
        document.removeEventListener('keydown', this._boundKeyDown);
    }

    /**
     * One-time setup that creates the internal `<dialog>` element, moves all
     * existing children into it, wires up event listeners (backdrop click,
     * native close, HTMX afterRequest), and re-processes HTMX attributes
     * on the moved children.
     *
     * The native `<dialog>` is styled with Tailwind classes for fixed
     * positioning, rounded corners, dark background, and a semi-transparent
     * backdrop overlay.
     *
     * Copies `aria-labelledby` from the host element to the `<dialog>` for
     * screen reader accessibility.
     * @private
     * @returns {void}
     */
    _build() {
        this._built = true;
        console.log('[mcms-dialog] build:', this.id || '(no id)', 'open:', this.hasAttribute('open'));

        // Collect existing children
        var children = [];
        while (this.firstChild) {
            children.push(this.removeChild(this.firstChild));
        }

        // Create native dialog
        var dialog = document.createElement('dialog');
        dialog.className = 'fixed inset-0 m-auto rounded-lg bg-gray-800 p-0 shadow-xl backdrop:bg-gray-900/75 sm:w-full sm:max-w-lg';

        // Copy relevant attributes from host to dialog
        var ariaLabel = this.getAttribute('aria-labelledby');
        if (ariaLabel) dialog.setAttribute('aria-labelledby', ariaLabel);

        // Re-attach children inside dialog
        for (var i = 0; i < children.length; i++) {
            dialog.appendChild(children[i]);
        }

        dialog.addEventListener('click', this._boundBackdropClick);
        dialog.addEventListener('close', function() {
            document.removeEventListener('keydown', this._boundKeyDown);
        }.bind(this));

        // Close on successful HTMX request from any element inside the dialog
        var self = this;
        document.body.addEventListener('htmx:afterRequest', function(e) {
            if (e.detail.successful && dialog.contains(e.detail.elt)) {
                self.close();
            }
        });

        this._dialog = dialog;
        this.appendChild(dialog);

        // Re-process HTMX attributes on moved children
        if (typeof htmx !== 'undefined') {
            htmx.process(dialog);
        }

        // Native <dialog> without open attribute is hidden by default,
        // but ensure it stays closed until .open() is called
        if (!this.hasAttribute('open')) {
            dialog.removeAttribute('open');
        }
    }

    /**
     * Handles click events on the internal `<dialog>` element to detect
     * backdrop clicks. When a `<dialog>` is opened with `showModal()`, clicks
     * on the `::backdrop` pseudo-element register as clicks on the `<dialog>`
     * element itself (since the pseudo-element is not a real DOM node). If the
     * click target is the dialog element directly (not a child), it is treated
     * as a backdrop click and the dialog is closed.
     * @private
     * @param {MouseEvent} e - The click event from the `<dialog>` element.
     * @returns {void}
     */
    _onBackdropClick(e) {
        // Click on the dialog backdrop (the ::backdrop pseudo-element triggers
        // clicks on the dialog element itself at the click coordinates).
        // If the click target is the dialog element, it's a backdrop click.
        if (e.target === this._dialog) {
            this.close();
        }
    }

    /**
     * Document-level keydown handler active only while the dialog is open.
     * Implements two behaviors:
     *
     * 1. **Escape key** -- prevents the native dialog Escape behavior (which
     *    would close without dispatching our custom event) and calls
     *    {@link McmsDialog#close} instead.
     * 2. **Tab key** -- implements a focus trap. When the user tabs past the
     *    last focusable element, focus wraps to the first. When Shift+Tab
     *    moves before the first focusable element, focus wraps to the last.
     *    This keeps keyboard focus contained within the modal.
     * @private
     * @param {KeyboardEvent} e - The keydown event from the document.
     * @returns {void}
     */
    _onKeyDown(e) {
        if (e.key === 'Escape') {
            e.preventDefault();
            this.close();
            return;
        }

        if (e.key === 'Tab' && this._dialog) {
            var focusable = this._dialog.querySelectorAll(this._focusableSelector);
            if (focusable.length === 0) return;
            var first = focusable[0];
            var last = focusable[focusable.length - 1];
            if (e.shiftKey) {
                if (document.activeElement === first) { e.preventDefault(); last.focus(); }
            } else {
                if (document.activeElement === last) { e.preventDefault(); first.focus(); }
            }
        }
    }

    // Public API

    /**
     * Opens the dialog as a modal (with backdrop). Adds a document-level
     * keydown listener for Escape and Tab focus trapping, dispatches the
     * `mcms-dialog:open` custom event, and focuses the first focusable
     * element inside the dialog on the next animation frame.
     *
     * If the internal `<dialog>` has not been built yet (e.g., the element
     * is not connected to the DOM), logs a warning and returns early.
     *
     * @fires mcms-dialog:open
     * @returns {void}
     *
     * @example
     * document.getElementById('edit-dialog').open();
     */
    open() {
        console.log('[mcms-dialog] open:', this.id || '(no id)');
        if (!this._dialog) { console.warn('[mcms-dialog] no internal dialog element'); return; }
        this._dialog.showModal();
        document.addEventListener('keydown', this._boundKeyDown);
        this.dispatchEvent(new CustomEvent('mcms-dialog:open', { bubbles: true }));
        // Focus first focusable
        var self = this;
        requestAnimationFrame(function() {
            var focusable = self._dialog.querySelectorAll(self._focusableSelector);
            if (focusable.length > 0) focusable[0].focus();
        });
    }

    /**
     * Closes the dialog. Calls the native `<dialog>.close()` method, removes
     * the document-level keydown listener, and dispatches the
     * `mcms-dialog:close` custom event.
     *
     * Safe to call even if the dialog is already closed or the internal
     * `<dialog>` element does not exist.
     *
     * @fires mcms-dialog:close
     * @returns {void}
     *
     * @example
     * document.getElementById('edit-dialog').close();
     */
    close() {
        console.log('[mcms-dialog] close:', this.id || '(no id)');
        if (!this._dialog) return;
        this._dialog.close();
        document.removeEventListener('keydown', this._boundKeyDown);
        this.dispatchEvent(new CustomEvent('mcms-dialog:close', { bubbles: true }));
    }

    /**
     * Alias for {@link McmsDialog#open} to match the native `<dialog>` API.
     * Code that calls `dialog.showModal()` will work interchangeably with
     * this custom element.
     *
     * @returns {void}
     *
     * @example
     * document.getElementById('edit-dialog').showModal();
     */
    showModal() { this.open(); }
}

customElements.define('mcms-dialog', McmsDialog);
