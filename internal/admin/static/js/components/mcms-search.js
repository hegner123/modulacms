// <mcms-search> — Debounced search input (Light DOM)
class McmsSearch extends HTMLElement {
    connectedCallback() {
        const placeholder = this.getAttribute('placeholder') || 'Search...';
        const name = this.getAttribute('name') || 'q';
        const value = this.getAttribute('value') || '';

        const input = document.createElement('input');
        input.type = 'search';
        input.name = name;
        input.placeholder = placeholder;
        input.value = value;
        input.className = 'search-input';

        // Copy HTMX attributes from host
        for (const attr of this.attributes) {
            if (attr.name.startsWith('hx-')) {
                input.setAttribute(attr.name, attr.value);
            }
        }

        this.appendChild(input);
        if (typeof htmx !== 'undefined') {
            htmx.process(input);
        }
    }
}

customElements.define('mcms-search', McmsSearch);
