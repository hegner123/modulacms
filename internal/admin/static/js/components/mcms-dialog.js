// <mcms-dialog> -- Modal dialog (Light DOM)
// Supports: open, title, confirm-label, cancel-label, destructive attributes
// Dispatches: mcms-dialog:confirm, mcms-dialog:cancel custom events
class McmsDialog extends HTMLElement {
    constructor() {
        super();
        this._backdrop = null;
        this._panel = null;
        this._originalChildren = [];
        this._built = false;
        this._boundKeyDown = this._onKeyDown.bind(this);
        this._focusableSelector = 'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])';
    }

    static get observedAttributes() {
        return ['open', 'title', 'confirm-label', 'cancel-label', 'destructive'];
    }

    connectedCallback() {
        this._build();
        this._syncVisibility();
    }

    disconnectedCallback() {
        document.removeEventListener('keydown', this._boundKeyDown);
        this._built = false;
    }

    attributeChangedCallback(name) {
        if (!this._built) return;

        if (name === 'open') {
            this._syncVisibility();
            return;
        }
        if (name === 'title') {
            var titleEl = this.querySelector('[data-dialog-title]');
            if (titleEl) {
                titleEl.textContent = this.getAttribute('title') || '';
            }
            return;
        }
        if (name === 'confirm-label') {
            var confirmBtn = this.querySelector('[data-dialog-confirm]');
            if (confirmBtn) {
                confirmBtn.textContent = this.getAttribute('confirm-label') || 'Confirm';
            }
            return;
        }
        if (name === 'cancel-label') {
            var cancelBtn = this.querySelector('[data-dialog-cancel]');
            if (cancelBtn) {
                cancelBtn.textContent = this.getAttribute('cancel-label') || 'Cancel';
            }
            return;
        }
        if (name === 'destructive') {
            var btn = this.querySelector('[data-dialog-confirm]');
            if (btn) {
                if (this.hasAttribute('destructive')) {
                    btn.classList.remove('bg-[var(--color-primary)]', 'hover:bg-[var(--color-primary-hover)]');
                    btn.classList.add('bg-[var(--color-danger)]', 'hover:bg-[var(--color-danger-hover)]');
                } else {
                    btn.classList.remove('bg-[var(--color-danger)]', 'hover:bg-[var(--color-danger-hover)]');
                    btn.classList.add('bg-[var(--color-primary)]', 'hover:bg-[var(--color-primary-hover)]');
                }
            }
        }
    }

    _build() {
        if (this._built) return;
        this._built = true;

        // Capture existing children before we restructure
        this._originalChildren = [];
        while (this.firstChild) {
            this._originalChildren.push(this.removeChild(this.firstChild));
        }

        // Build the dialog structure
        var backdrop = document.createElement('div');
        backdrop.className = 'fixed inset-0 z-[100] bg-black/60';
        backdrop.addEventListener('click', function(e) {
            if (e.target === backdrop) {
                this._cancel();
            }
        }.bind(this));

        var panel = document.createElement('div');
        panel.className = 'relative mx-auto mt-[10vh] w-full max-w-lg rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] shadow-lg';
        panel.setAttribute('role', 'dialog');
        panel.setAttribute('aria-modal', 'true');

        var titleText = this.getAttribute('title') || '';
        if (titleText) {
            var titleEl = document.createElement('h2');
            titleEl.className = 'border-b border-[var(--color-border)] px-5 py-4 text-lg font-semibold text-[var(--color-text)]';
            titleEl.setAttribute('data-dialog-title', '');
            titleEl.textContent = titleText;
            panel.appendChild(titleEl);
        }

        var body = document.createElement('div');
        body.className = 'px-5 py-4';
        for (var i = 0; i < this._originalChildren.length; i++) {
            body.appendChild(this._originalChildren[i]);
        }
        panel.appendChild(body);

        var actions = document.createElement('div');
        actions.className = 'flex justify-end gap-3 border-t border-[var(--color-border)] px-5 py-4';
        actions.setAttribute('data-dialog-actions', '');

        var cancelLabel = this.getAttribute('cancel-label') || 'Cancel';
        var cancelBtn = document.createElement('button');
        cancelBtn.type = 'button';
        cancelBtn.className = 'inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] transition-colors cursor-pointer border-none';
        cancelBtn.setAttribute('data-dialog-cancel', '');
        cancelBtn.textContent = cancelLabel;
        cancelBtn.addEventListener('click', this._cancel.bind(this));
        actions.appendChild(cancelBtn);

        var confirmLabel = this.getAttribute('confirm-label') || 'Confirm';
        var isDestructive = this.hasAttribute('destructive');
        var confirmBtn = document.createElement('button');
        confirmBtn.type = 'button';
        confirmBtn.className = 'inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium transition-colors cursor-pointer border-none ' + (isDestructive ? 'bg-[var(--color-danger)] text-white hover:bg-[var(--color-danger-hover)]' : 'bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)]');
        confirmBtn.setAttribute('data-dialog-confirm', '');
        confirmBtn.textContent = confirmLabel;
        confirmBtn.addEventListener('click', this._confirm.bind(this));
        actions.appendChild(confirmBtn);

        panel.appendChild(actions);
        backdrop.appendChild(panel);

        this._backdrop = backdrop;
        this._panel = panel;
        this.appendChild(backdrop);
    }

    _syncVisibility() {
        var isOpen = this.hasAttribute('open');
        if (this._backdrop) {
            this._backdrop.style.display = isOpen ? '' : 'none';
        }
        if (isOpen) {
            document.addEventListener('keydown', this._boundKeyDown);
            // Focus first focusable element inside panel
            var self = this;
            requestAnimationFrame(function() {
                self._trapFocusInside();
            });
        } else {
            document.removeEventListener('keydown', this._boundKeyDown);
        }
    }

    _trapFocusInside() {
        if (!this._panel) return;
        var focusable = this._panel.querySelectorAll(this._focusableSelector);
        if (focusable.length > 0) {
            focusable[0].focus();
        }
    }

    _onKeyDown(e) {
        if (!this.hasAttribute('open')) return;

        if (e.key === 'Escape') {
            e.preventDefault();
            this._cancel();
            return;
        }

        // Tab trap
        if (e.key === 'Tab') {
            var focusable = this._panel.querySelectorAll(this._focusableSelector);
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

    _confirm() {
        this.dispatchEvent(new CustomEvent('mcms-dialog:confirm', { bubbles: true }));
        this.removeAttribute('open');
    }

    _cancel() {
        this.dispatchEvent(new CustomEvent('mcms-dialog:cancel', { bubbles: true }));
        this.removeAttribute('open');
    }

    // Public API: open the dialog programmatically
    open() {
        this.setAttribute('open', '');
    }

    // Public API: close the dialog programmatically
    close() {
        this.removeAttribute('open');
    }
}

customElements.define('mcms-dialog', McmsDialog);
