
class CreateFieldForm extends HTMLElement {
    constructor() {
        super();
        // Attach a shadow DOM for encapsulation
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        // Called when the component is added to the DOM
        this.render();
        const form = document.getElementById('createFieldForm');
        const spinner = document.getElementById('loadingSpinner');
        const modal = document.getElementById('successModal');

        form.addEventListener('submit', function(event) {
            event.preventDefault();

            // Show loading spinner
            spinner.style.display = 'block';

            // Simulate form submission delay
            setTimeout(() => {
                spinner.style.display = 'none';
                modal.style.display = 'block';

                setTimeout(() => {
                    modal.style.display = 'none';
                    form.reset();
                }, 2000);
            }, 1500);
        });
    }

    render() {
        // Simple template for your component
        this.shadowRoot.innerHTML = `
    <div class="form-container">
        <div class="form-header">Create Field</div>
        <form id="createFieldForm">
            <div class="form-row">
                <div class="form-group">
                    <label for="routeId">Route ID</label>
                    <input type="number" id="routeId" name="routeId" required>
                    <span class="error" id="RouteError"></span>
                </div>
                <div class="form-group">
                    <label for="author">Author</label>
                    <input type="text" id="author" name="author">
                    <span class="error" id="authorError"></span>
                </div>
            </div>

            <div class="form-row">
                <div class="form-group">
                    <label for="authorId">Author ID</label>
                    <input type="text" id="authorId" name="authorId">
                    <span class="error" id="authorIdError"></span>
                </div>
                <div class="form-group">
                    <label for="key">Key</label>
                    <input type="text" id="key" name="key">
                    <span class="error" id="keyError"></span>
                </div>
            </div>

            <div class="form-row">
                <div class="form-group">
                    <label for="data">Data</label>
                    <input type="text" id="data" name="data">
                    <span class="error" id="dataError"></span>
                </div>
                <div class="form-group">
                    <label for="dateCreated">Date Created</label>
                    <input type="date" id="dateCreated" name="dateCreated">
                    <span class="error" id="dateCreatedError"></span>
                </div>
            </div>

            <div class="form-row">
                <div class="form-group">
                    <label for="dateModified">Date Modified</label>
                    <input type="date" id="dateModified" name="dateModified">
                    <span class="error" id="dateModifiedError"></span>
                </div>
                <div class="form-group">
                    <label for="component">Component</label>
                    <input type="text" id="component" name="component">
                    <span class="error" id="componentError"></span>
                </div>
            </div>

            <div class="form-row">
                <div class="form-group">
                    <label for="tags">Tags</label>
                    <input type="text" id="tags" name="tags">
                    <span class="error" id="tagsError"></span>
                </div>
                <div class="form-group">
                    <label for="parent">Parent</label>
                    <input type="text" id="parent" name="parent">
                    <span class="error" id="parentError"></span>
                </div>
            </div>

            <button type="submit" class="submit-button">Submit</button>
        </form>

        <div class="spinner" id="loadingSpinner">Loading...</div>
        <div class="modal" id="successModal">Field created successfully!</div>
    </div>
    `;
    }
}

// Define the new element
customElements.define('create-field', CreateFieldForm);
window.addEventListener("load", app)
function app(e) {
    e.preventDefault()
}
