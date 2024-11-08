
class SettingsComponent extends HTMLElement {
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
        this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          font-family: Arial, sans-serif;
        }
      </style>
       <h2>Settings</h2> 
         <form>
                <label for="port">Port</label>
                <input type="text" id="port" name="port" value="8080">
                <span class="error error-1"></span>

                <label for="ssl_port">SSL Port</label>
                <input type="text" id="ssl_port" name="ssl_port" value="443">
                <span class="error error-2"></span>

                <label for="client_site">Client Site</label>
                <input type="text" id="client_site" name="client_site" value="example.com">
                <span class="error error-3"></span>

                <label for="db_url">DB URL</label>
                <input type="text" id="db_url" name="db_url" value="default">
                <span class="error error-4"></span>

                <label for="db_name">DB Name</label>
                <input type="text" id="db_name" name="db_name" value="default">
                <span class="error error-5"></span>

                <label for="db_password">DB Password</label>
                <input type="text" id="db_password" name="db_password" value="none">
                <span class="error error-6"></span>

                <label for="bucket_url">Bucket URL</label>
                <input type="text" id="bucket_url" name="bucket_url" value="local">
                <span class="error error-7"></span>

                <label for="bucket_password">Bucket Password</label>
                <input type="text" id="bucket_password" name="bucket_password" value="none">
                <span class="error error-8"></span>
            </form>
    `;
    }
}

// Define the new element
customElements.define('settings-component', SettingsComponent);


