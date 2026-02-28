# Installation

ModulaCMS is a Go application that compiles to a single binary. You can build it from source or run it with Docker. SQLite works out of the box with zero configuration; MySQL and PostgreSQL require a running database server.

## System Requirements

| Requirement | Details |
|-------------|---------|
| Go | 1.24 or later |
| CGO | Must be enabled (`CGO_ENABLED=1`) |
| C compiler | GCC or Clang (for the SQLite driver) |
| OS | Linux or macOS |
| Build runner | `just` ([installation](https://github.com/casey/just#installation)) |

CGO is required because the SQLite driver (`mattn/go-sqlite3`) is a C library. Even if you plan to use MySQL or PostgreSQL, the binary still compiles with the SQLite driver.

## Building from Source

### Development Build

```bash
just dev
```

This produces `modula-x86` in the project root. The build:

- Runs `templ generate` for the admin panel templates.
- Bundles admin panel JavaScript assets via esbuild.
- Embeds version, git commit, and build date via ldflags.
- Uses vendored dependencies (`-mod vendor`).

### Production Build

```bash
just build
```

This produces the binary at `out/bin/modula-x86`. Same build process as `just dev` but outputs to the `out/bin/` directory for cleaner deployment workflows.

### Build and Run

```bash
just run
```

This runs `just dev` and then immediately executes the binary.

### Verify the Build

```bash
./modula-x86 version
```

This prints the embedded version, commit hash, and build date.

### Other Build Commands

| Command | Description |
|---------|-------------|
| `just check` | Compile-check without producing a binary |
| `just clean` | Remove build artifacts (`bin/`, `out/`) |
| `just vendor` | Update vendored dependencies |

## Docker

ModulaCMS provides Docker Compose configurations for different database backends. All compose files are in `deploy/docker/`.

### Full Stack (All Databases + MinIO)

```bash
just dc full up
```

Starts ModulaCMS with PostgreSQL, MySQL, MinIO (S3-compatible storage), and builds the CMS container. This is useful for development and testing across all backends.

### Single Database Stacks

```bash
# SQLite (minimal, no external database)
just dc sqlite up

# MySQL
just dc mysql up

# PostgreSQL
just dc postgres up
```

Each stack includes only the CMS container and the relevant database (SQLite needs no external container). The SQLite stack also includes a MinIO container for media storage.

### Docker Stack Management

All stacks support the same set of actions:

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

To start just the database and storage containers without the CMS (useful when running the CMS binary locally):

```bash
just docker-infra
```

This starts PostgreSQL, MySQL, and MinIO containers.

## Database Setup

### SQLite (Default)

SQLite requires no setup. On first run, ModulaCMS creates a `modula.db` file in the working directory. The default `config.json` uses these settings:

```json
{
  "db_driver": "sqlite",
  "db_url": "./modula.db",
  "db_name": "modula.db"
}
```

### MySQL

Provide a running MySQL instance and configure `config.json`:

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

Provide a running PostgreSQL instance and configure `config.json`:

```json
{
  "db_driver": "postgres",
  "db_url": "localhost:5432",
  "db_name": "modulacms",
  "db_username": "modula",
  "db_password": "your-password"
}
```

## First-Run Setup

When you run `./modula-x86 serve` and no `config.json` exists, ModulaCMS runs automatic setup:

1. **Creates `config.json`** with default settings (SQLite, development environment, default ports).
2. **Creates database tables** for all CMS entities.
3. **Inserts bootstrap data** including three roles (admin, editor, viewer), 47 permissions, and a system admin user.
4. **Validates** the database setup.

The system admin credentials are printed to the log output:

```
Generated system admin password  email=system@modulacms.local  password=<random-string>
```

### Interactive Setup

For an interactive setup wizard that lets you choose database driver, connection details, and other settings:

```bash
./modula-x86 serve --wizard
```

The wizard prompts for each configuration option with validation and retry support (up to 3 attempts).

### Standalone Install Command

You can also run the installation separately from the server:

```bash
./modula-x86 install
```

This runs the interactive installer without starting the servers.

## TLS Certificates

For local HTTPS development, generate self-signed certificates:

```bash
./modula-x86 cert generate
```

This creates `localhost.crt` and `localhost.key` in the certificate directory (default: `./`). In production, ModulaCMS uses Let's Encrypt autocert for automatic certificate provisioning when the `environment` is not set to `local` or `docker`.

## CLI Commands

The binary provides several management commands beyond `serve`:

| Command | Description |
|---------|-------------|
| `serve` | Start all servers (HTTP, HTTPS, SSH) |
| `serve --wizard` | Interactive setup wizard before starting |
| `install` | Run installation wizard (without starting servers) |
| `version` | Show version, commit, and build date |
| `cert generate` | Generate self-signed TLS certificates |
| `config show` | Display current configuration |
| `config validate` | Validate configuration file |
| `backup` | Database backup operations |
| `db` | Database management commands |
| `tui` | Launch the TUI without starting servers |
| `plugin` | Plugin management (list, init, validate, info, reload, enable, disable) |
| `deploy` | Deployment operations |

Global flags available on all commands:

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `config.json` | Path to configuration file |
| `-v`, `--verbose` | `false` | Enable debug logging |

## Notes

- **Backup tools.** PostgreSQL backups require `pg_dump` in your PATH. MySQL backups require `mysqldump`. SQLite backups need no external tools. The installer warns if these are missing.
- **S3 storage is optional.** Media uploads require an S3-compatible storage provider (AWS S3, MinIO, DigitalOcean Spaces, etc.), but the CMS starts and functions without it -- media upload endpoints will return errors until storage is configured.
- **OAuth is optional.** The CMS functions with local authentication only. OAuth (Google, GitHub, Azure) can be configured later in `config.json`.
