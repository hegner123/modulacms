# Quickstart

ModulaCMS is a headless CMS that runs as a single binary with three servers: HTTP (REST API + admin panel), HTTPS, and SSH (terminal-based management TUI). This guide gets you from zero to running in under five minutes.

## Install

Install the `modula` binary to your PATH so you can run it from any directory:

```bash
# Build from source
git clone https://github.com/hegner123/modulacms.git
cd modulacms
just build
```

Then copy the binary to a directory in your PATH:

```bash
cp out/bin/modula-x86 /usr/local/bin/modula
```

Verify the installation:

```bash
modula version
```

### Prerequisites

- **Go 1.24+** with CGO enabled (`CGO_ENABLED=1`)
- **just** command runner ([installation](https://github.com/casey/just#installation))
- **GCC or Clang** (CGO compiles the SQLite driver)
- Linux or macOS (Windows is not supported)

## Create a Project

Navigate to where you want your project to live and run init:

```bash
mkdir mysite && cd mysite
modula init
```

`modula init` runs the install wizard, then automatically registers the project in `~/.modula/configs.json` using the directory name as the project name and `local` as the environment. If it's the first project registered, it becomes the default.

For non-interactive setup:

```bash
modula init --yes --admin-password your-password
```

To use a custom project name:

```bash
modula init --name my-site --admin-password pw
```

This creates `config.json` and `modula.db` in the current directory, creates all database tables, seeds bootstrap data (roles, permissions, system user), and registers the project.

## Start the Server

```bash
modula serve
```

Watch the startup logs for the system admin credentials if you used the interactive installer without setting a password:

```
Generated system admin password  email=system@modulacms.local  password=<random-string>
```

Save this password. You need it to log in.

## Default Ports

| Server | Address | Purpose |
|--------|---------|---------|
| HTTP | `localhost:8080` | REST API + admin panel |
| HTTPS | `localhost:4000` | TLS-secured API (requires certificates) |
| SSH | `localhost:2233` | Terminal TUI for content management |

## Connect

### Web Admin Panel

Open [http://localhost:8080/admin/](http://localhost:8080/admin/) in your browser. Log in with:

- **Email:** `system@modulacms.local`
- **Password:** the password from setup

### SSH TUI

```bash
ssh localhost -p 2233
```

### REST API

```bash
# Log in and capture the session cookie
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "system@modulacms.local", "password": "YOUR_PASSWORD"}' \
  -c cookies.txt

# List content types
curl http://localhost:8080/api/v1/datatype \
  -b cookies.txt

# Health check (no auth required)
curl http://localhost:8080/api/v1/health
```

## Manage From Anywhere

Since `modula init` registered the project, you can already manage it from any directory:

```bash
cd ~
modula connect              # uses default project + env
modula connect mysite       # explicit project, default env
modula connect mysite local # explicit project + env
```

## Multiple Environments

Each project supports multiple environments (local, staging, production), each pointing to a different `config.json`:

```bash
modula connect set mysite local ./config.json
modula connect set mysite staging /srv/mysite/staging/config.json
modula connect set mysite prod /srv/mysite/prod/config.json
```

For remote environments, the `config.json` uses `remote_url` and `remote_api_key` instead of `db_driver` -- the TUI connects over HTTPS via the Go SDK instead of a direct database connection.

## Next Steps

- [Installation Guide](installation.md) -- build options, Docker, database backends, remote connections
- [Configuration Reference](configuration.md) -- all config.json fields and options

## Stopping the Server

Press `Ctrl+C` once for graceful shutdown (30-second timeout). Press `Ctrl+C` a second time to force immediate exit.
