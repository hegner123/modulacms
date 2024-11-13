
class MediaUpload extends HTMLElement {
    constructor() {
        super();
        // Attach a shadow DOM for encapsulation
        this.attachShadow({ mode: 'open' });
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
          <h4>File upload</h4>
          <form type="submit" hx-post="/admin/media/create" hx-encoding="multipart/form-data">
              <label>Upload</label>
              <input id="media-upload"  type="file"/>
              <button>Submit</button>
          </form>
      </div>
    `;
    }
}

// Define the new element
customElements.define('media-upload', MediaUpload);
