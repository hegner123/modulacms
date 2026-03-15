// <mcms-toast> — Global toast notification system (Light DOM)
class McmsToast extends HTMLElement {
    constructor() {
        super();
        this._container = null;
    }

    connectedCallback() {
        const position = this.getAttribute('position') || 'bottom-right';
        this._container = document.createElement('div');
        this._container.className = 'fixed z-[200] flex flex-col gap-2 max-w-sm ' + (position === 'top-right' ? 'top-4 right-4' : position === 'top-left' ? 'top-4 left-4' : position === 'bottom-left' ? 'bottom-4 left-4' : 'bottom-4 right-4');
        document.body.appendChild(this._container);
    }

    disconnectedCallback() {
        if (this._container && this._container.parentNode) {
            this._container.parentNode.removeChild(this._container);
        }
    }

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

    _escape(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
}

customElements.define('mcms-toast', McmsToast);
