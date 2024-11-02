class ParagraphComponent extends HTMLElement {
    constructor() {
        super();
        // Attach a shadow DOM for encapsulation
        this.attachShadow({ mode: 'open' });
        this.field
        this.value
        this.validation
    }

    connectedCallback() {
        // Called when the component is added to the DOM
        this.render();
    }

    render() {
        // Simple template for your component
        this.shadowRoot.innerHTML = `
      <style>
        /* Scoped styles go here */
        :host {
          display: block;
          font-family: Arial, sans-serif;
        }
      </style>
      <div>
        Hello, World! This is a bare-bones web component.
      </div>
    `;
    }
}

// Define the new element
customElements.define('paragraph-component', ParagraphComponent);

