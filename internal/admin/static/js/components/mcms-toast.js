// <mcms-toast> — Global toast notification system (Light DOM)
class McmsToast extends HTMLElement {
    constructor() {
        super();
        this._container = null;
    }

    connectedCallback() {
        const position = this.getAttribute('position') || 'bottom-right';
        this._container = document.createElement('div');
        this._container.className = 'toast-container ' + position;
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
        toast.className = 'toast toast-' + (type || 'info');
        toast.setAttribute('role', 'alert');
        toast.innerHTML = '<span class="toast-message">' + this._escape(message) + '</span>' +
            '<button class="toast-close" aria-label="Close">&times;</button>';
        toast.querySelector('.toast-close').addEventListener('click', () => toast.remove());
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
