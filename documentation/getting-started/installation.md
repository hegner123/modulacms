# Installation

ModulaCMS compiles to a single binary. Install it to your PATH and use it to create and manage projects from any directory. The project registry at `~/.modula/configs.json` tracks all your projects and environments, so you can switch between them without navigating to each project directory.

## System Requirements

| Requirement | Details |
|-------------|---------|
| Go | 1.24 or later |
| CGO | Must be enabled (`CGO_ENABLED=1`) |
| C compiler | GCC or Clang (for the SQLite driver) |
| OS | Linux or macOS |
| Build runner | `just` ([installation](https://github.com/casey/just#installation)) |

CGO is required because the SQLite driver (`mattn/go-sqlite3`) is a C library. Even if you plan to use MySQL or PostgreSQL, the binary still compiles with the SQLite driver.

## Installing the Binary

### Build from Source

```bash
git clone https://github.com/hegner123/modulacms.git
cd modulacms
just build
```

This produces the binary at `out/bin/modula-x86`. Copy it to your PATH:

```bash
cp out/bin/modula-x86 /usr/local/bin/modula
```

Verify:

```bash
modula version
```

### Development Build

For contributors working on ModulaCMS itself:

```bash
just dev    # builds modula-x86 in project root
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

## Creating a Project

Once `modula` is in your PATH, create a project anywhere:

```bash
mkdir ~/projects/mysite && cd ~/projects/mysite
modula init
```

`modula init` does two things:

1. **Runs the install wizard** -- same as `modula install`. Prompts for database driver, connection details, ports, and admin credentials. Creates `config.json`, initializes the database, and seeds bootstrap data.
2. **Registers the project** -- adds an entry to `~/.modula/configs.json` with the project name (defaults to the directory name) and environment `local` pointing to the new `config.json`. Sets it as the default project if it's the first one registered.

For automated/CI setups:

```bash
modula init --yes --admin-password your-password
```

To use a custom project name instead of the directory name:

```bash
modula init --name my-site --admin-password pw
```

### What Init Creates

1. **`config.json`** with your chosen settings.
2. **Database tables** for all CMS entities.
3. **Bootstrap data** including three roles (admin, editor, viewer), 47 permissions, and a system admin user.
4. **Validation** of the database setup.
5. **Registry entry** in `~/.modula/configs.json` mapping the project name to this directory's config.

The system admin credentials are printed to the log:

```
Generated system admin password  email=system@modulacms.local  password=<random-string>
```

### Install Without Registration

`modula install` runs the same wizard without the registry step. Use this when you don't need the project registered (e.g., Docker containers, ephemeral environments):

```bash
modula install --yes --admin-password your-password
```

## Project Registry

The registry at `~/.modula/configs.json` maps project names to environments, each pointing to a `config.json` path. `modula init` populates this automatically. You can also manage it manually.

### Manual Registration

```bash
modula connect set mysite local ~/projects/mysite/config.json
```

This creates a project named "mysite" with a "local" environment. The first environment added becomes the default.

### Set Defaults

```bash
modula connect default mysite              # default project
modula connect default mysite local        # default env for a project
```

### Use From Anywhere

```bash
modula connect                             # default project, default env
modula connect mysite                      # specific project, default env
modula connect mysite staging              # specific project + env
```

The `connect` command resolves the config path from the registry, changes to the project directory, and launches the TUI. If the config has `remote_url` set, it connects over HTTPS via the SDK. Otherwise, it opens a local database connection.

### Manage the Registry

```bash
modula connect list                        # show all projects + envs
modula connect set mysite prod /srv/config.json  # add environment
modula connect remove mysite               # remove entire project
modula connect remove mysite --env staging # remove one environment
```

### Auto-Detection

If no project is specified and no default is set, `modula connect` checks for a `config.json` in the current working directory as a fallback.

## Multiple Environments

A typical setup has multiple environments per project:

```bash
# Local development (direct database)
modula connect set mysite local ~/projects/mysite/config.json

# Staging (remote connection over HTTPS)
modula connect set mysite staging ~/projects/mysite/staging.json

# Production (remote connection over HTTPS)
modula connect set mysite prod ~/projects/mysite/prod.json
```

**Local config** (`config.json`) uses `db_driver` for a direct database connection:

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

When connecting with a remote config, the TUI operates over the REST API instead of a direct database connection. All features work the same way -- the `RemoteDriver` implements the same `DbDriver` interface.

## Docker

ModulaCMS provides Docker Compose configurations for different database backends. All compose files are in `deploy/docker/`.

### Full Stack (All Databases + MinIO)

```bash
just dc full up
```

Starts ModulaCMS with PostgreSQL, MySQL, MinIO (S3-compatible storage), and builds the CMS container.

### Single Database Stacks

```bash
just dc sqlite up     # SQLite (minimal, no external database)
just dc mysql up      # MySQL
just dc postgres up   # PostgreSQL
```

### Docker Stack Management

| Command | Description |
|---------|-------------|
| `just dc <backend> up` | Build and start containers |
| `just dc <backend> down` | Stop containers (keep volumes) |
| `just dc <backend> reset` | Stop containers and delete volumes |
| `just dc <backend> dev` | Rebuild and restart only the CMS container |
| `just dc <backend> fresh` | Reset volumes, then rebuild and start |
| `just dc <backend> logs` | Tail CMS container logs |

Replace `<backend>` with `full`, `sqlite`, `mysql`, or `postgres`.

### Infrastructure Only

To start just the database and storage containers without the CMS (useful when running the binary locally):

```bash
just docker-infra
```

## Database Setup

### SQLite (Default)

SQLite requires no setup. On first run, ModulaCMS creates a `modula.db` file in the working directory:

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

For local HTTPS development, generate self-signed certificates:

```bash
modula cert generate
```

This creates `localhost.crt` and `localhost.key` in the certificate directory. In production, ModulaCMS uses Let's Encrypt autocert for automatic certificate provisioning when the `environment` is not set to `local` or `docker`.

## CLI Reference

| Command | Description |
|---------|-------------|
| `init` | Install wizard + register project in registry |
| `init --name <name>` | Init with custom project name (default: directory name) |
| `init --yes --admin-password <pw>` | Non-interactive init with defaults |
| `serve` | Start all servers (HTTP, HTTPS, SSH) |
| `serve --wizard` | Interactive setup wizard before starting |
| `install` | Run installation wizard only (no registry) |
| `install --yes --admin-password <pw>` | Non-interactive install with defaults |
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

Global flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `config.json` | Path to configuration file |
| `-v`, `--verbose` | `false` | Enable debug logging |

## Notes

- **Backup tools.** PostgreSQL backups require `pg_dump` in your PATH. MySQL backups require `mysqldump`. SQLite backups need no external tools.
- **S3 storage is optional.** Media uploads require an S3-compatible storage provider (AWS S3, MinIO, etc.), but the CMS starts without it -- media upload endpoints return errors until storage is configured.
- **OAuth is optional.** The CMS functions with local authentication only. OAuth (Google, GitHub, Azure) can be configured later.
- **Remote connections.** The `connect` command with a remote config uses the Go SDK over HTTPS. All TUI features work identically whether connected locally or remotely.
