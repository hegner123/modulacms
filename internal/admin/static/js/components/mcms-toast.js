/**
 * @module mcms-toast
 * @description Global toast notification system (Light DOM) for the ModulaCMS
 * admin panel.
 *
 * Renders a fixed-position container on the viewport and exposes a
 * {@link McmsToast#show} method to display timed, dismissible notifications.
 * Each toast is styled with a colored left border indicating its type
 * (success, error, or info) and includes a close button.
 *
 * The container is appended to `document.body` (not inside the custom element)
 * so that toasts render above all other content regardless of stacking context.
 * When the element is disconnected, the container is removed from the body.
 *
 * Typically a single `<mcms-toast>` element is placed once in the base layout
 * and referenced globally via its ID.
 *
 * @example <caption>Single global toast element in the admin layout</caption>
 * <mcms-toast id="toast" position="bottom-right" duration="5000"></mcms-toast>
 *
 * @example <caption>Showing notifications from JavaScript</caption>
 * const toast = document.getElementById('toast');
 * toast.show('Item created successfully', 'success');
 * toast.show('Failed to save changes', 'error');
 * toast.show('Processing your request...', 'info');
 *
 * @example <caption>Persistent toast (no auto-dismiss)</caption>
 * <mcms-toast id="toast" duration="0"></mcms-toast>
 * <script>
 *   // duration="0" disables auto-dismiss; user must click the close button
 *   document.getElementById('toast').show('Manual close only', 'info');
 * </script>
 */

/**
 * Toast notification custom element that manages a fixed-position container
 * of dismissible, auto-expiring notification messages.
 *
 * Each toast is rendered as an ARIA `role="alert"` element with a colored
 * left border, message text, and a close button. Toasts auto-dismiss after
 * the configured duration unless duration is set to `0`.
 *
 * The notification container is appended directly to `document.body` so it
 * floats above all page content. Position is controlled by the `position`
 * attribute.
 *
 * @extends HTMLElement
 *
 * @attr {string} [position="bottom-right"] - Corner of the viewport where toasts
 *   appear. Supported values: `"top-right"`, `"top-left"`, `"bottom-left"`,
 *   `"bottom-right"`. Read once during `connectedCallback`; changing after
 *   mount has no effect.
 * @attr {string} [duration="5000"] - Time in milliseconds before a toast
 *   auto-dismisses. Set to `"0"` to disable auto-dismiss (user must click
 *   the close button). Read each time {@link McmsToast#show} is called, so
 *   it can be changed between calls.
 *
 * @example
 * // Typical admin layout placement
 * <mcms-toast id="toast" position="bottom-right" duration="5000"></mcms-toast>
 *
 * // Trigger from HTMX response header (HX-Trigger: {"showToast": {...}})
 * document.body.addEventListener('showToast', (e) => {
 *   document.getElementById('toast').show(e.detail.message, e.detail.type);
 * });
 */
class McmsToast extends HTMLElement {
    constructor() {
        super();

        /**
         * Reference to the fixed-position container `<div>` appended to
         * `document.body`. Holds all active toast notification elements.
         * Created in {@link McmsToast#connectedCallback}, removed in
         * {@link McmsToast#disconnectedCallback}.
         * @type {HTMLDivElement|null}
         * @private
         */
        this._container = null;
    }

    /**
     * Called when the element is inserted into the DOM. Creates a fixed-position
     * container `<div>` and appends it to `document.body`. The container's
     * position class is determined by the `position` attribute (defaults to
     * `"bottom-right"`).
     *
     * The container is styled with:
     * - `z-index: 200` to float above modals and overlays
     * - `flex-col` layout with gap for stacking multiple toasts
     * - `max-w-sm` to constrain toast width on large screens
     * @returns {void}
     */
    connectedCallback() {
        const position = this.getAttribute('position') || 'bottom-right';
        this._container = document.createElement('div');
        this._container.className = 'fixed z-[200] flex flex-col gap-2 max-w-sm ' + (position === 'top-right' ? 'top-4 right-4' : position === 'top-left' ? 'top-4 left-4' : position === 'bottom-left' ? 'bottom-4 left-4' : 'bottom-4 right-4');
        document.body.appendChild(this._container);
    }

    /**
     * Called when the element is removed from the DOM. Removes the toast
     * container from `document.body` to prevent orphaned DOM nodes.
     * Safe to call if the container was never created or was already removed.
     * @returns {void}
     */
    disconnectedCallback() {
        if (this._container && this._container.parentNode) {
            this._container.parentNode.removeChild(this._container);
        }
    }

    /**
     * Displays a toast notification message in the container.
     *
     * Creates a styled `<div>` with `role="alert"` containing the escaped
     * message text and a close button. The left border color reflects the
     * toast type:
     * - `"success"` -- green border (`--color-success`)
     * - `"error"` -- red border (`--color-danger`)
     * - Any other value (including `"info"` or `undefined`) -- blue border (`--color-primary`)
     *
     * The toast auto-removes itself after the duration specified by the
     * `duration` attribute (default 5000ms). If duration is `0`, the toast
     * persists until the user clicks the close button.
     *
     * The close button is identified by both a `data-toast-close` attribute
     * and a class for dual-targeting flexibility.
     *
     * Does nothing if the container has not been created (element is not
     * connected to the DOM).
     *
     * @param {string} message - The notification text to display. HTML entities
     *   are escaped via {@link McmsToast#_escape} to prevent XSS.
     * @param {string} [type] - The notification type controlling border color.
     *   One of `"success"`, `"error"`, or any other value for the default
     *   info/primary style.
     * @returns {void}
     *
     * @example
     * const toast = document.getElementById('toast');
     *
     * // Success notification
     * toast.show('Changes saved', 'success');
     *
     * // Error notification
     * toast.show('Upload failed: file too large', 'error');
     *
     * // Info/default notification
     * toast.show('Syncing content...', 'info');
     */
    show(message, type) {
        if (!this._container) return;
        const duration = parseInt(this.getAttribute('duration') || '5000', 10);
        const toast = document.createElement('div');
        var borderColor = type === 'success' ? 'border-l-[var(--color-success)]' : type === 'error' ? 'border-l-[var(--color-danger)]' : 'border-l-[var(--color-primary)]';
        toast.className = 'flex items-start gap-3 rounded-lg border border-[var(--color-border)] border-l-4 ' + borderColor + ' bg-[var(--color-surface)] p-4 shadow-lg';
        toast.setAttribute('role', 'alert');
        // DUAL: data-toast-close + class
        toast.innerHTML = '<span class="flex-1 text-sm text-[var(--color-text)]">' + this._escape(message) + '</span>' +
            '<button class="ml-auto cursor-pointer bg-transparent border-none text-[var(--color-text-muted)] hover:text-[var(--color-text)]" data-toast-close aria-label="Close">&times;</button>';
        toast.querySelector('[data-toast-close]').addEventListener('click', () => toast.remove());
        this._container.appendChild(toast);
        if (duration > 0) {
            setTimeout(() => { if (toast.parentNode) toast.remove(); }, duration);
        }
    }

    /**
     * Escapes a string for safe insertion into HTML. Creates a temporary
     * `<div>`, sets its `textContent` (which escapes HTML entities), then
     * reads back the `innerHTML`. This prevents XSS when rendering
     * user-provided or server-provided notification messages.
     * @private
     * @param {string} str - The raw string to escape.
     * @returns {string} The HTML-escaped string safe for innerHTML insertion.
     */
    _escape(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
}

customElements.define('mcms-toast', McmsToast);
