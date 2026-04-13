/**
 * @module mcms-drawer
 * @description Slide-in drawer component (Light DOM) for the ModulaCMS admin panel.
 *
 * Slides in from the right edge of the viewport. Includes a semi-transparent
 * backdrop, Escape-to-close, Tab focus trapping, and HTMX integration
 * (auto-close on successful request from within the drawer).
 *
 * The drawer opens when its innerHTML is populated (via HTMX swap into
 * #content-drawer). It listens for htmx:afterSwap on itself and auto-opens
 * when content arrives. The drawer closes and clears its content when dismissed.
 *
 * @example <caption>HTMX-driven drawer</caption>
 * <div id="content-drawer">
 *   <mcms-drawer>
 *     <div class="p-6">Drawer content here</div>
 *   </mcms-drawer>
 * </div>
 *
 * @fires mcms-drawer:open When the drawer opens
 * @fires mcms-drawer:close When the drawer closes
 */
class McmsDrawer extends HTMLElement {
    constructor() {
        super();
        this._isOpen = false;
        this._focusableSelector = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
        this._onKeyDown = this._onKeyDown.bind(this);
        this._onBackdropClick = this._onBackdropClick.bind(this);
        this._onHtmxSuccess = this._onHtmxSuccess.bind(this);
        this._previousFocus = null;
    }

    connectedCallback() {
        // Auto-open if the drawer has content on connection
        if (this.children.length > 0) {
            this.open();
        }

        // Listen for HTMX success events to auto-close
        document.body.addEventListener('htmx:afterRequest', this._onHtmxSuccess);

        // Process HTMX attributes on content
        if (typeof htmx !== 'undefined') {
            htmx.process(this);
        }
    }

    disconnectedCallback() {
        document.body.removeEventListener('htmx:afterRequest', this._onHtmxSuccess);
        document.removeEventListener('keydown', this._onKeyDown);
    }

    open() {
        if (this._isOpen) return;
        this._isOpen = true;
        this._previousFocus = document.activeElement;

        // Build backdrop if not present
        if (!this._backdrop) {
            this._backdrop = document.createElement('div');
            this._backdrop.className = 'fixed inset-0 bg-black/40 transition-opacity duration-200';
            this._backdrop.style.zIndex = 'var(--z-drawer, 90)';
            this._backdrop.style.opacity = '0';
            this._backdrop.addEventListener('click', this._onBackdropClick);
        }

        // Insert backdrop before drawer
        this.parentNode.insertBefore(this._backdrop, this);

        // Style the drawer panel
        this.style.position = 'fixed';
        this.style.top = '0';
        this.style.right = '0';
        this.style.bottom = '0';
        this.style.width = '80vw';
        this.style.maxWidth = '80vw';
        this.style.zIndex = 'var(--z-drawer, 90)';
        this.style.transform = 'translateX(100%)';
        this.style.transition = 'transform 200ms ease-out';
        this.style.overflowY = 'auto';

        // Trigger animation on next frame
        requestAnimationFrame(function() {
            this._backdrop.style.opacity = '1';
            this.style.transform = 'translateX(0)';
        }.bind(this));

        document.addEventListener('keydown', this._onKeyDown);
        this.dispatchEvent(new CustomEvent('mcms-drawer:open', { bubbles: true }));

        // Focus first focusable element
        requestAnimationFrame(function() {
            var focusable = this.querySelector(this._focusableSelector);
            if (focusable) focusable.focus();
        }.bind(this));
    }

    close() {
        if (!this._isOpen) return;
        this._isOpen = false;

        this.style.transform = 'translateX(100%)';
        if (this._backdrop) {
            this._backdrop.style.opacity = '0';
        }

        document.removeEventListener('keydown', this._onKeyDown);

        // Remove backdrop and clear content after animation
        var self = this;
        setTimeout(function() {
            if (self._backdrop && self._backdrop.parentNode) {
                self._backdrop.parentNode.removeChild(self._backdrop);
            }
            // Clear drawer container content
            var container = document.getElementById('content-drawer');
            if (container) {
                container.innerHTML = '';
            }
        }, 200);

        if (this._previousFocus && typeof this._previousFocus.focus === 'function') {
            this._previousFocus.focus();
        }

        this.dispatchEvent(new CustomEvent('mcms-drawer:close', { bubbles: true }));
    }

    _onKeyDown(e) {
        if (e.key === 'Escape') {
            e.preventDefault();
            this.close();
            return;
        }
        // Focus trap
        if (e.key === 'Tab') {
            var focusable = Array.from(this.querySelectorAll(this._focusableSelector));
            if (focusable.length === 0) return;
            var first = focusable[0];
            var last = focusable[focusable.length - 1];
            if (e.shiftKey) {
                if (document.activeElement === first) {
                    e.preventDefault();
                    last.focus();
                }
            } else {
                if (document.activeElement === last) {
                    e.preventDefault();
                    first.focus();
                }
            }
        }
    }

    _onBackdropClick() {
        this.close();
    }

    _onHtmxSuccess(e) {
        // Do NOT auto-close on form submissions. The drawer uses auto-save
        // on debounce, so closing on every save would be disruptive.
        // The user closes the drawer explicitly via Close button or Escape.
    }
}

if (!customElements.get('mcms-drawer')) {
    customElements.define('mcms-drawer', McmsDrawer);
}
