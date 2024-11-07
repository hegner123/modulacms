/**
 *@class CreateFieldForm
  @method private submitForm
 */
class CreateFieldForm extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.root = this.shadowRoot
    }

    /**
     * @method submitForm
     * @param {SubmitEvent} event
     * @param {HTMLFormElement} form 
     * @returns {any}
     * @description Method called on form submision to POST form to api
     */
    async #submitForm(event, form) {
        event.preventDefault()
        const formData = this.#getFormDataAsJson(form)
        const url = "https://localhost/admin/field/add"
        try {
            const response = await fetch(url, {
                method: "POST", 
                headers: {
                    "Content-Type": "application/json" 
                },
                body: formData
            });

            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }

            const data = await response.json();
            console.log("Form submitted successfully:", data);
        } catch (error) {
            console.error("Error submitting the form:", error);
        }
    }
    /**
    * @method getFormDataAsJson
    * @param {HTMLFormElement} form 
    */
    #getFormDataAsJson(form) {
        const formData = new FormData(form);
        const formObject = {};
        for (let [key, value] of formData.entries()) {
            formObject[key] = value;
        }
        return JSON.stringify(formObject);
    }

    /**
    *@method connectedCallback
    *@description part of the custom component api

    */
    connectedCallback() {
        // Called when the component is added to the DOM
        this.render();
        const root = this
        const form = this.shadowRoot.getElementById('createFieldForm');
        const spinner = this.shadowRoot.getElementById('loadingSpinner');
        const modal = this.shadowRoot.getElementById('successModal');
        const errors = this.shadowRoot.querySelectorAll(".error")
        const sub = this.#submitForm
        form.addEventListener('submit', async function(event) {
            event.preventDefault();
            spinner.style.display = 'block';
            let res = sub.call(root, event, form)
            if (res === "OK"){
                spinner.style.display = 'none';
                modal.style.display = 'block';
                setTimeout(() => {
                    modal.style.display = 'none';
                }, 2000);
            } else {
                errors.forEach(span=>{
                    console.log(span)
                })


            }

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
