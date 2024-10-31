class LoginForm extends HTMLElement {
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
          font-family:  sans-serif;
        }
        .login-form {
            min-width:400px;
            max-width:500px;
            background:#555;
            border-radius:15px;
            padding:1.5rem;
            display:grid;
            grid-template-columns:1fr;
            grid-template-rows:max-content 1fr;
            gap:1rem;
        }
        .login-header h3{
            color:#fff;
            font-size:2rem;
            margin-block-end:0;
            margin-block-start:0;
        }
        .login-body{
            display:grid;
            grid-template-columns:1fr;
            grid-template-rows:1fr 1fr  max-content;
            gap:1rem;
        }
        .form-input-group{
            color:#fff;
            display:grid;
            grid-template-columns:1fr;
        }
        .form-input-group input{
            font-size:1rem;
            border:none;
            margin:0.25em 0 0;
            height:1.65rem;
            border-radius:5px;
            padding:0.5em;
        }
        .login-sm-link {
            padding:0.25em 0;
            width:fit-content;
            margin-top:0.5em;
            color:#008;
        }

        .login-form-submit{
            padding:15px;
            max-width:50%;
            border-radius:5px;
            font-weight:600;
            background:#fff;
            border:none;
            cursor:pointer;
        }
        .login-form-submit:hover{
            background:#f0f0f0;
            
        }
      </style>
        <div class="login-form">
            <div class="login-header">
                <h3>Login</h3>
            </div>
            <form action="/" class="login-body">
                    <div class="form-input-group">
                        <label class="login-user-input-label" for="useremail">Email / Username</label>
                        <input class="login-user-input" type="text" name="useremail">
                        <span class="login-user-input-error"></span>
                    </div>
                    <div class="form-input-group">
                        <label class="login-password-label" for="password">Password</label>
                        <input class="login-password" type="password" name="password">
                        <span class="login-password-error"></span>
                        <a href="/forgot-password" class="login-sm-link">Forgot password?</a>
                    </div>
                    <input class="login-form-submit" type="submit">
                </form>
        </div>
    `;
    }
}

// Define the new element
customElements.define('login-form', LoginForm);

