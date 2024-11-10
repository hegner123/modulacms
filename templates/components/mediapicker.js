class MediaPicker extends HTMLElement {
    constructor() {
        super();
        // Attach a shadow DOM for encapsulation
        this.attachShadow({ mode: 'open' });
        this.document = this.shadowRoot
        this.image = ""
    }

    async fetchMedia() {
        this.images={}
        const url = "http://localhost:8500/api/media/get"
        try {
            const response = await fetch(url)
            if (response) {
                this.images = await response.json()
                this.displayMedia()
            }
        } catch (err) {
            throw new Error(err)
        }
    }
    displayMedia() {

        this.mediaDisplay = this.shadowRoot.querySelector("#display")
        if (this.images.length > 0){
            for(let i=0;i < this.images.length;i++){
               const wrapper = document.createElement("div") 
                const image = document.createElement("img")
                const label = document.createElement("span")
                image.setAttribute("src",this.images[i].url)
                wrapper.append(image)
                label.innerHTML=this.images[i].name
                wrapper.append(label)
                this.mediaDisplay.append(wrapper)
                

            }
        }

    }






    connectedCallback() {
        // Called when the component is added to the DOM
        console.dir(this)
        this.render();
        if (this.document) {
            this.modal = this.shadowRoot.querySelector("#modal")
            this.openButton = this.shadowRoot.querySelector("#open")
            this.closeButton = this.shadowRoot.querySelector("#close")
            this.fetchButton = this.shadowRoot.querySelector("#fetch")
        }
        this.openButton.addEventListener("click", () => {
            this.modal.showModal()
        })
        this.closeButton.addEventListener("click", () => {
            this.modal.close()
        })

        this.fetchButton.addEventListener("click", () => {
            this.fetchMedia()
        })

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
        .image-gallery{
            display:grid;
            grid-template-columns:repeat(auto-fit,300px);
            grid-gap:2rem;
            width:90vw;
            height:90vh;
        }
        img{
            width:100%;
        }
      </style>
      <dialog id="modal">
      <button id="fetch">Fetch</button>
      <div id="display" class="image-gallery"></div>
      <button id="close">Close</button>
      </dialog>
      <button id="open">Open</button>
    `;
    }
}

// Define the new element
customElements.define('media-picker', MediaPicker);

