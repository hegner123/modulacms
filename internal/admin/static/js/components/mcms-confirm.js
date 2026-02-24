// <mcms-confirm> — Inline confirmation button (Light DOM)
class McmsConfirm extends HTMLElement {
    connectedCallback() {
        const label = this.getAttribute('label') || 'Delete';
        const confirmLabel = this.getAttribute('confirm-label') || 'Confirm';
        const timeout = parseInt(this.getAttribute('timeout') || '3000', 10);

        const btn = document.createElement('button');
        btn.className = 'btn btn-danger btn-sm';
        btn.textContent = label;
        btn.type = 'button';

        let armed = false;
        let timer = null;

        btn.addEventListener('click', (e) => {
            if (!armed) {
                e.preventDefault();
                e.stopPropagation();
                armed = true;
                btn.textContent = confirmLabel;
                btn.classList.add('armed');
                timer = setTimeout(() => {
                    armed = false;
                    btn.textContent = label;
                    btn.classList.remove('armed');
                }, timeout);
            } else {
                clearTimeout(timer);
                // Copy HTMX attributes from host to button and trigger
                for (const attr of this.attributes) {
                    if (attr.name.startsWith('hx-')) {
                        btn.setAttribute(attr.name, attr.value);
                    }
                }
                htmx.process(btn);
                htmx.trigger(btn, 'click');
            }
        });

        this.appendChild(btn);
    }
}

customElements.define('mcms-confirm', McmsConfirm);
