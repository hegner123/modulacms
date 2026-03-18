/**
 * @module mcms-confirm
 * @description Confirmation dialog component that renders a styled trigger button paired
 * with an `<mcms-dialog>` for user confirmation before destructive or significant actions.
 *
 * HTMX attributes (`hx-*`) placed on the host element are automatically forwarded to the
 * confirm button inside the dialog, preventing accidental requests when clicking the
 * trigger button. After a successful HTMX request from the confirm button, the dialog
 * closes automatically.
 *
 * This is a Light DOM web component -- all rendered elements are direct children of the
 * host element, making them accessible to global CSS and HTMX processing.
 *
 * @example
 * <!-- Basic danger confirmation (default variant) -->
 * <mcms-confirm
 *     label="Delete"
 *     message="This will permanently delete this item."
 *     hx-delete="/admin/items/123"
 *     hx-target="#item-row-123"
 *     hx-swap="outerHTML"
 * ></mcms-confirm>
 *
 * @example
 * <!-- Primary variant with custom action label -->
 * <mcms-confirm
 *     label="Publish"
 *     message="This will make the content publicly visible."
 *     action-label="Confirm Publish"
 *     variant="primary"
 *     hx-post="/admin/content/456/publish"
 *     hx-target="#main-content"
 * ></mcms-confirm>
 *
 * @example
 * <!-- With icon content and custom button classes -->
 * <mcms-confirm
 *     label="Remove"
 *     icon="<svg>...</svg>"
 *     message="Remove this field from the datatype?"
 *     button-class="btn btn-icon btn-danger"
 *     hx-delete="/admin/fields/789"
 * ></mcms-confirm>
 */

/**
 * Trigger button and confirmation dialog web component.
 *
 * Renders a button that, when clicked, opens a confirmation dialog containing a message
 * and Cancel/Confirm buttons. The confirm button receives all `hx-*` attributes that
 * were originally on the host element, so the HTMX request fires only after explicit
 * user confirmation.
 *
 * @extends HTMLElement
 *
 * @attr {string} label - Text displayed on the trigger button. Default: `"Delete"`.
 * @attr {string} message - Confirmation message shown inside the dialog body.
 *   Default: `"Are you sure?"`.
 * @attr {string} action-label - Text for the confirm button inside the dialog.
 *   Default: same as `label`.
 * @attr {string} variant - Visual style variant. `"danger"` renders red buttons (default),
 *   `"primary"` renders blue/brand-colored buttons.
 * @attr {string} button-class - When set, overrides the trigger button's CSS classes
 *   entirely, bypassing variant-based styling.
 * @attr {string} icon - HTML string (typically an SVG) to render inside the trigger
 *   button instead of text. When set, the `label` is used as `aria-label` instead.
 * @attr {string} hx-* - Any HTMX attribute (e.g., `hx-delete`, `hx-post`, `hx-target`,
 *   `hx-swap`) is removed from the host and forwarded to the confirm button inside
 *   the dialog. This prevents HTMX from firing on the trigger button click.
 *
 * @example
 * const confirm = document.querySelector('mcms-confirm');
 * // The component self-initializes on connectedCallback -- no JS API needed.
 * // Interaction is entirely through attributes and HTMX.
 */
class McmsConfirm extends HTMLElement {
    /**
     * Lifecycle callback invoked when the element is inserted into the DOM.
     *
     * Reads all configuration attributes, collects and removes `hx-*` attributes from
     * the host, builds the trigger button and `<mcms-dialog>` with Cancel/Confirm buttons,
     * wires click handlers, processes HTMX on the confirm button, and listens for
     * `htmx:afterRequest` to auto-close the dialog on success.
     */
    connectedCallback() {
        var label = this.getAttribute('label') || 'Delete';
        var message = this.getAttribute('message') || 'Are you sure?';
        var actionLabel = this.getAttribute('action-label') || label;
        var variant = this.getAttribute('variant') || 'danger';
        var buttonClass = this.getAttribute('button-class');

        // Collect and remove hx-* attributes from host (prevent HTMX from
        // processing them on the host element, which would fire on any click)
        var hxAttrs = [];
        var toRemove = [];
        for (var i = 0; i < this.attributes.length; i++) {
            var attr = this.attributes[i];
            if (attr.name.indexOf('hx-') === 0) {
                hxAttrs.push({ name: attr.name, value: attr.value });
                toRemove.push(attr.name);
            }
        }
        for (var r = 0; r < toRemove.length; r++) {
            this.removeAttribute(toRemove[r]);
        }

        // Trigger button
        var icon = this.getAttribute('icon') || '';
        var triggerBtn = document.createElement('button');
        triggerBtn.type = 'button';
        if (buttonClass) {
            triggerBtn.className = buttonClass;
        } else if (variant === 'primary') {
            triggerBtn.className = 'rounded-md bg-[var(--color-primary)] px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-[var(--color-primary-hover)]';
        } else {
            triggerBtn.className = 'rounded-md bg-red-500 px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-red-400';
        }
        if (icon) {
            triggerBtn.innerHTML = icon;
            triggerBtn.setAttribute('aria-label', label);
        } else {
            triggerBtn.textContent = label;
        }

        // Dialog
        var dialog = document.createElement('mcms-dialog');

        var body = document.createElement('div');
        body.className = 'p-6';
        var msg = document.createElement('p');
        msg.className = 'text-sm text-gray-400';
        msg.textContent = message;
        body.appendChild(msg);
        dialog.appendChild(body);

        var actions = document.createElement('div');
        actions.className = 'flex justify-end gap-x-3 border-t border-white/10 bg-white/5 px-6 py-3';

        var cancelBtn = document.createElement('button');
        cancelBtn.type = 'button';
        cancelBtn.className = 'rounded-md bg-white/10 px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-white/20';
        cancelBtn.textContent = 'Cancel';
        cancelBtn.addEventListener('click', function() { dialog.close(); });
        actions.appendChild(cancelBtn);

        var confirmBtn = document.createElement('button');
        confirmBtn.type = 'button';
        if (variant === 'primary') {
            confirmBtn.className = 'rounded-md bg-[var(--color-primary)] px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-[var(--color-primary-hover)]';
        } else {
            confirmBtn.className = 'rounded-md bg-red-500 px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-red-400';
        }
        confirmBtn.textContent = actionLabel;

        // Forward hx-* attributes to the confirm button
        for (var j = 0; j < hxAttrs.length; j++) {
            confirmBtn.setAttribute(hxAttrs[j].name, hxAttrs[j].value);
        }

        actions.appendChild(confirmBtn);
        dialog.appendChild(actions);

        // Wire trigger
        triggerBtn.addEventListener('click', function() { dialog.open(); });

        this.appendChild(triggerBtn);
        this.appendChild(dialog);

        // Let HTMX discover the hx-* attributes on the confirm button
        if (typeof htmx !== 'undefined') {
            htmx.process(confirmBtn);
        }

        // Close dialog after successful HTMX request
        document.body.addEventListener('htmx:afterRequest', function(e) {
            if (e.detail.elt === confirmBtn && e.detail.successful) {
                dialog.close();
            }
        });
    }
}

customElements.define('mcms-confirm', McmsConfirm);
