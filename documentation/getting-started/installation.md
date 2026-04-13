# Installation

Install ModulaCMS from source and configure it for your environment.

## Quick Start

```bash
git clone https://github.com/hegner123/modulacms.git
cd modulacms && just build
cp out/bin/modula /usr/local/bin/modula
mkdir ~/mysite && cd ~/mysite
modula init
modula serve
```

For full details on each step, read on.

## System Requirements

| Requirement | Details |
|-------------|---------|
| Go | 1.25 or later |
| CGO | Must be enabled (`CGO_ENABLED=1`) |
| C compiler | GCC or Clang (for the SQLite driver) |
| OS | Linux or macOS |
| Build runner | `just` ([installation](https://github.com/casey/just#installation)) |

> **Good to know**: CGO is required because the SQLite driver (`mattn/go-sqlite3`) is a C library. Even if you use MySQL or PostgreSQL, the binary still compiles with the SQLite driver.

## Build from Source

```bash
git clone https://github.com/hegner123/modulacms.git
cd modulacms
just build
```

This produces the binary at `out/bin/modula`. Copy it to your PATH:

```bash
cp out/bin/modula /usr/local/bin/modula
```

Verify:

```bash
modula version
```

### Development Build

For contributors working on ModulaCMS itself:

```bash
just dev    # builds modula in project root
just run    # builds and immediately runs
just check  # compile-check without producing a binary
```

| Command | Description |
|---------|-------------|
| `just dev` | Build development binary in project root |
| `just run` | Build and run immediately |
| `just check` | Compile-check without producing a binary |
| `just clean` | Remove build artifacts |
| `just vendor` | Update vendored dependencies |

## Create a Project

Once `modula` is in your PATH, create a project anywhere:

```bash
mkdir ~/projects/mysite && cd ~/projects/mysite
modula init
```

`modula init` is idempotent -- each step checks whether its output already exists and skips if so. Safe to run multiple times. It performs six steps:

1. **Loads or creates the project registry** at `~/.modula/configs.json`.
2. **Creates a `modula/` project directory** (or uses the current directory if `modula.config.json` already exists there).
3. **Writes config files** -- a base `modula.config.json` plus three environment overlays (`modula.local.config.json`, `modula.dev.config.json`, `modula.prod.config.json`).
4. **Registers configs** in the project registry and sets `local` as the default environment.
5. **Generates localhost TLS certificates** in a `certs/` directory.
6. **Creates and seeds the SQLite database** -- skipped for external databases (MySQL/PostgreSQL) or if the `.db` file already exists.

### Init modes

| Mode | Flag | Behavior |
|------|------|----------|
| Interactive | (default) | Prompts for project name and admin password |
| CI | `--mode ci` or `--yes` | No prompts; requires `--admin-password` |
| Container | `--mode container` | No prompts; skips database creation entirely |

```bash
modula init                                       # interactive
modula init --mode ci --admin-password s3cret!    # CI pipeline
modula init --mode container                      # Docker entrypoint
modula init --name my-site                        # custom project name
```

### What init creates

```
mysite/
  modula/
    modula.config.json           # base config (shared defaults)
    modula.local.config.json     # local overlay (SQLite, dev ports)
    modula.dev.config.json       # dev overlay
    modula.prod.config.json      # prod overlay (PostgreSQL, env vars)
    certs/                       # self-signed localhost certificates
    modula.db                    # SQLite database (if applicable)
```

The database is seeded with bootstrap data: three roles (admin, editor, viewer), 72 permissions, and a system admin user.

When no `--admin-password` is provided outside interactive mode, ModulaCMS generates a random password and prints it to the log:

```
Generated system admin password  email=system@modula.local  password=<random-string>
```

### Check project status

After init, verify everything is in place:

```bash
modula status
```

This shows the base config, certificates, registered environments, and available commands for the project in the current directory.

## Project Registry

The registry at `~/.modula/configs.json` maps project names to environments, each pointing to a `modula.config.json` path. `modula init` populates this automatically. You can also manage it manually.

### Register manually

```bash
modula connect set mysite local ~/projects/mysite/modula.config.json
```

This creates a project named "mysite" with a "local" environment. The first environment added becomes the default.

### Set defaults

```bash
modula connect default mysite              # default project
modula connect default mysite local        # default env for a project
```

### Connect from anywhere

```bash
modula connect                             # default project, default env
modula connect mysite                      # specific project, default env
modula connect mysite staging              # specific project + env
```

The `connect` command resolves the config path from the registry, changes to the project directory, and launches the TUI. If the config has `remote_url` set, it connects over HTTPS via the SDK instead of opening a local database connection.

### Manage the registry

```bash
modula connect list                        # show all projects + envs
modula connect set mysite prod /srv/modula.config.json  # add environment
modula connect remove mysite               # remove entire project
modula connect remove mysite --env staging # remove one environment
```

> **Good to know**: If no project is specified and no default is set, `modula connect` checks for a `modula.config.json` in the current working directory as a fallback.

## Multiple Environments

A typical setup has multiple environments per project:

```bash
# Local development (direct database)
modula connect set mysite local ~/projects/mysite/modula.config.json

# Staging (remote connection over HTTPS)
modula connect set mysite staging ~/projects/mysite/staging.json

# Production (remote connection over HTTPS)
modula connect set mysite prod ~/projects/mysite/prod.json
```

**Local config** (`modula.config.json`) uses `db_driver` for a direct database connection:

```json
{
  "db_driver": "sqlite",
  "db_url": "./modula.db"
}
```

**Remote config** (`staging.json`, `prod.json`) uses `remote_url` for an HTTPS connection via the Go SDK:

```json
{
  "remote_url": "https://staging.mysite.com",
  "remote_api_key": "your-api-key"
}
```

When you connect with a remote config, the TUI operates over the REST API. All features work the same way whether you connect locally or remotely.

## Docker

ModulaCMS provides Docker Compose configurations for different database backends. All compose files live in `deploy/docker/`.

### Full stack (all databases + MinIO)

```bash
just dc full up
```

Starts ModulaCMS with PostgreSQL, MySQL, MinIO (S3-compatible storage), and the CMS container.

### Single database stacks

```bash
just dc sqlite up     # SQLite (minimal, no external database)
just dc mysql up      # MySQL
just dc postgres up   # PostgreSQL
```

### Docker stack management

| Command | Description |
|---------|-------------|
| `just dc <backend> up` | Build and start containers |
| `just dc <backend> down` | Stop containers (keep volumes) |
| `just dc <backend> reset` | Stop containers and delete volumes |
| `just dc <backend> dev` | Rebuild and restart only the CMS container |
| `just dc <backend> fresh` | Reset volumes, then rebuild and start |
| `just dc <backend> logs` | Tail CMS container logs |

Replace `<backend>` with `full`, `sqlite`, `mysql`, or `postgres`.

### Infrastructure only

Start the database and storage containers without the CMS (useful when running the binary locally):

```bash
just docker-infra
```

## Database Setup

### SQLite (default)

SQLite requires no setup. ModulaCMS creates a `modula.db` file in the working directory on first run:

```json
{
  "db_driver": "sqlite",
  "db_url": "./modula.db",
  "db_name": "modula.db"
}
```

### MySQL

```json
{
  "db_driver": "mysql",
  "db_url": "localhost:3306",
  "db_name": "modulacms",
  "db_username": "modula",
  "db_password": "your-password"
}
```

### PostgreSQL

```json
{
  "db_driver": "postgres",
  "db_url": "localhost:5432",
  "db_name": "modulacms",
  "db_username": "modula",
  "db_password": "your-password"
}
```

## TLS Certificates

Generate self-signed certificates for local HTTPS development:

```bash
modula cert generate
```

This creates `localhost.crt` and `localhost.key` in the certificate directory. Staging and production environments use Let's Encrypt autocert for automatic certificate provisioning. Development environments use these self-signed certificates. Local environments skip TLS entirely.

## CLI Reference

| Command | Description |
|---------|-------------|
| `init` | Initialize project (idempotent, safe to re-run) |
| `init --mode ci --admin-password <pw>` | Non-interactive init for CI |
| `init --mode container` | Init for Docker (skips DB creation) |
| `init --name <name>` | Init with custom project name |
| `status` | Show project status for the current directory |
| `serve` | Start all servers (HTTP, HTTPS, SSH) |
| `serve --wizard` | Interactive setup wizard before starting |
| `connect` | Launch TUI for a registered project |
| `connect set <name> <env> <path>` | Register a project environment |
| `connect list` | List all registered projects |
| `connect remove <name>` | Remove a project from registry |
| `connect default <name> [env]` | Set default project or environment |
| `version` | Show version, commit, and build date |
| `cert generate` | Generate self-signed TLS certificates |
| `config show` | Display current configuration |
| `config validate` | Validate configuration file |
| `backup` | Database backup operations |
| `db` | Database management commands |
| `tui` | Launch the TUI without starting servers |
| `plugin` | Plugin management (list, init, validate, info, reload, enable, disable) |
| `deploy` | Deployment operations |

### Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `modula.config.json` | Path to configuration file |
| `--overlay` | | Overlay config file (merged on top of `--config`) |
| `-v`, `--verbose` | `false` | Enable debug logging |
| `-y`, `--yes` | `false` | Auto-accept all prompts (equivalent to `--mode ci`) |

> **Good to know**: S3 storage is optional. Media upload endpoints return errors until you configure storage, but the CMS starts without it.

## Next steps

- [Your First Project](/docs/getting-started/first-project) -- the three-step path from init to connected
- [Configuration](/docs/getting-started/configuration) -- all `modula.config.json` fields and options
- [Content Modeling](/docs/building-content/content-modeling) -- design datatypes and fields for your content
