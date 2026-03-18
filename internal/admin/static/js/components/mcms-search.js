/**
 * @module mcms-search
 * @description Light DOM web component that renders a debounced search input with HTMX
 * integration. The component creates a styled `<input type="search">` element and
 * automatically forwards any `hx-*` attributes from the host element to the inner input,
 * enabling HTMX-powered server-side search without manual wiring.
 *
 * This component is designed to work with HTMX's `hx-trigger` attribute for debounced
 * search. A typical setup uses `hx-trigger="input changed delay:300ms"` on the host
 * element, which is forwarded to the inner input and processed by HTMX after the component
 * calls `htmx.process()`.
 *
 * The component renders into the Light DOM (no Shadow DOM), so the inner input inherits
 * page-level styles and participates in standard form submission.
 *
 * @example <caption>Basic search input</caption>
 * <mcms-search placeholder="Search content..." name="q"></mcms-search>
 *
 * @example <caption>HTMX-powered search with debounce</caption>
 * <mcms-search
 *   placeholder="Search..."
 *   name="q"
 *   value=""
 *   hx-get="/admin/content"
 *   hx-target="#results"
 *   hx-trigger="input changed delay:300ms"
 *   hx-swap="innerHTML">
 * </mcms-search>
 *
 * @example <caption>Pre-populated search with custom name</caption>
 * <mcms-search placeholder="Filter users..." name="filter" value="admin"></mcms-search>
 */

/**
 * A debounced search input web component with automatic HTMX attribute forwarding.
 *
 * On connection, the component reads its configuration from HTML attributes, creates a
 * styled `<input type="search">`, copies all `hx-*` attributes from the host element
 * to the inner input, and calls `htmx.process()` to activate HTMX behaviors on the
 * dynamically created element.
 *
 * @extends HTMLElement
 *
 * @attr {string} [placeholder="Search..."] - Placeholder text displayed in the empty input.
 * @attr {string} [name="q"] - The input's `name` attribute, used in form submissions and
 *   HTMX request parameters.
 * @attr {string} [value=""] - The initial value of the search input.
 * @attr {string} [hx-*] - Any HTMX attribute (e.g., `hx-get`, `hx-target`, `hx-trigger`,
 *   `hx-swap`, `hx-indicator`, `hx-push-url`, `hx-include`) is forwarded from the host
 *   element to the inner `<input>` element. This allows full HTMX configuration on the
 *   custom element tag without needing to access the inner input directly.
 *
 * @example <caption>Programmatic creation</caption>
 * const search = document.createElement('mcms-search');
 * search.setAttribute('placeholder', 'Find items...');
 * search.setAttribute('hx-get', '/api/search');
 * search.setAttribute('hx-target', '#results');
 * search.setAttribute('hx-trigger', 'input changed delay:300ms');
 * container.appendChild(search);
 */
class McmsSearch extends HTMLElement {
    /**
     * Called when the element is inserted into the DOM. Creates the inner `<input type="search">`
     * element, applies styling, copies all `hx-*` attributes from the host element to the
     * input, appends it to the Light DOM, and initializes HTMX processing on the input if
     * HTMX is available.
     *
     * The HTMX attribute forwarding iterates over all attributes on the host element and
     * copies any whose name starts with `hx-` to the inner input. This enables declarative
     * HTMX configuration on the `<mcms-search>` tag itself.
     *
     * After appending the input, `htmx.process()` is called to ensure HTMX recognizes and
     * activates its attributes on the dynamically created element. If HTMX is not loaded,
     * this step is silently skipped and the input functions as a standard search field.
     */
    connectedCallback() {
        const placeholder = this.getAttribute('placeholder') || 'Search...';
        const name = this.getAttribute('name') || 'q';
        const value = this.getAttribute('value') || '';

        const input = document.createElement('input');
        input.type = 'search';
        input.name = name;
        input.placeholder = placeholder;
        input.value = value;
        input.className = 'w-full rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]';

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
