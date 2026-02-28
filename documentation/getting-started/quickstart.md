# Quickstart

ModulaCMS is a headless CMS that runs as a single binary with three servers: HTTP (REST API + admin panel), HTTPS, and SSH (terminal-based management TUI). This guide gets you from clone to running in under five minutes.

## Prerequisites

- **Go 1.24+** with CGO enabled (`CGO_ENABLED=1`)
- **just** command runner ([installation](https://github.com/casey/just#installation))
- **GCC or Clang** (CGO compiles the SQLite driver)
- Linux or macOS (Windows is not supported)

SQLite is the default database -- no external database server needed.

## Build

Clone the repository and build the development binary:

```bash
git clone https://github.com/hegner123/modulacms.git
cd modulacms
just dev
```

This produces a `modula-x86` binary in the project root with version info embedded via ldflags.

## Run

```bash
./modula-x86 serve
```

On first run with no `config.json` present, ModulaCMS automatically:

1. Creates a `config.json` with default settings.
2. Creates a `modula.db` SQLite database.
3. Creates database tables and inserts bootstrap data (roles, permissions, system user).
4. Generates and logs a system admin password.

Watch the startup logs for a line like:

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
- **Password:** the generated password from the startup logs

The admin panel provides a full management interface for content, schema, media, users, roles, and settings.

### SSH TUI

Connect to the terminal-based management interface:

```bash
ssh localhost -p 2233
```

The TUI provides the same management capabilities as the admin panel in a terminal interface.

### REST API

The API is available at `/api/v1/`. Authenticate first, then make requests:

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

## Next Steps

- [Installation Guide](installation.md) -- production builds, Docker, database options
- [Configuration Reference](configuration.md) -- all config.json fields and options

## Stopping the Server

Press `Ctrl+C` once for graceful shutdown (30-second timeout). Press `Ctrl+C` a second time to force immediate exit.
