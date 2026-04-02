# Quickstart

Get a ModulaCMS project running in three steps. This guide assumes you have the `modula` binary installed. If not, follow the [Installation guide](installation.md) first.

## 1. Create a project

```bash
mkdir mysite && cd mysite
modula init
```

`modula init` creates a `modula/` directory with config files, generates TLS certificates, initializes the SQLite database, seeds default roles and permissions, and registers the project. Each step is idempotent -- safe to run again if interrupted.

For non-interactive setup (CI pipelines):

```bash
modula init --mode ci --admin-password your-password
```

## 2. Start the server

```bash
modula serve
```

ModulaCMS starts three servers:

| Server | Address | Purpose |
|--------|---------|---------|
| HTTP | `localhost:8080` | REST API + admin panel |
| HTTPS | `localhost:4000` | TLS-secured API |
| SSH | `localhost:2233` | Terminal UI |

> **Good to know**: If you did not set a password during init, check the startup logs for the generated admin credentials.

## 3. Connect

**Web admin panel** -- open [http://localhost:8080/admin/](http://localhost:8080/admin/) and log in with `system@modulacms.local` and your password.

**Terminal UI** -- run `ssh localhost -p 2233`.

**REST API:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "system@modulacms.local", "password": "YOUR_PASSWORD"}' \
  -c cookies.txt

curl http://localhost:8080/api/v1/datatype -b cookies.txt
```

Once the project is registered, you can manage it from any directory:

```bash
modula serve mysite         # start the server for this project
modula tui mysite           # launch TUI for this project
modula connect mysite       # connect via SSH/TUI
```

## Next steps

- [Configuration](configuration.md) -- customize ports, database, S3 storage, OAuth, and more
- [Content Modeling](../building-content/content-modeling.md) -- design datatypes and fields for your content
- [SDK Overview](../sdks/overview.md) -- choose a client library for your frontend
