# ModulaCMS

[![CI/CD](https://github.com/hegner123/modulacms/actions/workflows/go.yml/badge.svg)](https://github.com/hegner123/modulacms/actions/workflows/go.yml)

A headless CMS written in Go that serves content over HTTP/HTTPS and provides SSH access for backend management.

## Overview

ModulaCMS decouples the admin panel from the CMS backend, allowing agencies to build custom admin interfaces while using a fast, flexible backend. It runs as a single binary with concurrent HTTP, HTTPS, and SSH servers.

## Features

- **Single Binary Deployment** - HTTP, HTTPS, and SSH servers in one executable
- **Multi-Database Support** - SQLite, MySQL, and PostgreSQL
- **Tree-Based Content** - Sibling-pointer tree structure with O(1) operations
- **SSH Terminal UI** - Full TUI for developer/ops management via SSH
- **S3-Compatible Storage** - Media upload and optimization with any S3-compatible provider
- **Let's Encrypt Integration** - Automatic SSL certificate management
- **OAuth Authentication** - Configurable OAuth provider support
- **Dynamic Content Schema** - Define datatypes and fields without code changes
- **Lua Plugin System** - Extend functionality with Lua scripts
- **CMS Migration** - Import from Contentful, Sanity, Strapi, WordPress
- **Backup/Restore** - Built-in backup system with S3 storage option

## Requirements

- Go 1.24+
- CGO enabled (required for SQLite driver)
- [just](https://github.com/casey/just) command runner
- Linux or macOS (Windows is not currently supported)
- For production: MySQL or PostgreSQL recommended
- For SDK development: Node.js 22+, pnpm 9+

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/hegner123/modulacms.git
cd modulacms

# Build for development
just dev

# Run the installation wizard
./modulacms-x86 --install
```

### Running

```bash
# Start all servers (HTTP, HTTPS, SSH)
./modulacms-x86

# Start TUI in CLI mode (local, no SSH)
./modulacms-x86 --cli

# Show version
./modulacms-x86 --version

# Check for updates
./modulacms-x86 --update

# Generate self-signed SSL certificates for local development
./modulacms-x86 --gen-certs

# Use custom config file
./modulacms-x86 --config=/path/to/config.json
```

### Connecting via SSH

```bash
ssh -p 2222 user@localhost
```

## Configuration

ModulaCMS uses a JSON configuration file. Create `config.json` in the project root:

```json
{
  "port": ":8080",
  "ssl_port": ":8443",
  "ssh_port": "2222",
  "ssh_host": "0.0.0.0",
  "environment": "local",
  "client_site": "localhost",
  "admin_site": "admin.localhost",
  "db_driver": "sqlite",
  "db_url": "modula.db",
  "cert_dir": "./certs",
  "bucket_endpoint": "s3.amazonaws.com",
  "bucket_media": "media-bucket",
  "bucket_backup": "backup-bucket",
  "bucket_access_key": "",
  "bucket_secret_key": "",
  "oauth_provider_name": "google",
  "oauth_client_id": "",
  "oauth_client_secret": "",
  "oauth_redirect_url": "http://localhost:8080/api/v1/auth/oauth",
  "oauth_scopes": ["profile", "email"],
  "oauth_endpoint": {
    "oauth_auth_url": "https://accounts.google.com/o/oauth2/auth",
    "oauth_token_url": "https://oauth2.googleapis.com/token"
  },
  "cors_origins": ["http://localhost:3000"],
  "cors_methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
  "cors_headers": ["Content-Type", "Authorization"],
  "cors_credentials": true
}
```

### Database Drivers

| Driver | `db_driver` | `db_url` Example |
|--------|-------------|------------------|
| SQLite | `sqlite` | `./modula.db` |
| MySQL | `mysql` | `user:pass@tcp(localhost:3306)/dbname` |
| PostgreSQL | `psql` | `postgres://user:pass@localhost:5432/dbname?sslmode=disable` |

## API

ModulaCMS provides a RESTful API at `/api/v1`.

### Authentication

```
POST /api/v1/auth/register  - Register new user
POST /api/v1/auth/reset     - Reset password
GET  /api/v1/auth/oauth     - OAuth callback
```

### Content Management

```
GET    /api/v1/admincontentdatas     - List content
POST   /api/v1/admincontentdatas     - Create content
GET    /api/v1/admincontentdatas/?q= - Get content by ID
PUT    /api/v1/admincontentdatas/?q= - Update content
DELETE /api/v1/admincontentdatas/?q= - Delete content
```

### Schema Management

```
GET/POST   /api/v1/datatypes   - Manage content types
GET/POST   /api/v1/fields      - Manage fields
```

### Media

```
POST /api/v1/upload  - Upload media files
```

See [API Documentation](ai/api/API_CONTRACT.md) for complete reference.

## Development

ModulaCMS uses [just](https://github.com/casey/just) as its command runner. Run `just` to see all available recipes.

### Build Commands

```bash
just dev        # Build local binary (./modulacms-x86) with version info
just build      # Build production binary to out/bin/
just run        # Build and run
just check      # Compile-check without producing artifacts
just clean      # Remove build artifacts
```

### Testing

```bash
just test              # Run all Go tests
just coverage          # Run tests with coverage report
just test-integration  # S3 integration tests (requires MinIO: just test-minio first)
```

### Database

```bash
just sqlc      # Generate Go code from SQL queries
just dump      # Dump SQLite database to SQL file
```

### Code Quality

```bash
just lint      # Run all linters (go, dockerfile, yaml)
just lint-go   # Lint Go code via Docker
just vendor    # Update vendor directory
```

### TypeScript SDKs

The `sdks/typescript/` directory contains a pnpm workspace with three packages:

| Package | Description |
|---------|-------------|
| `@modulacms/types` | Shared entity types, branded IDs, enums |
| `@modulacms/sdk` | Read-only content delivery SDK |
| `@modulacms/admin-sdk` | Full admin CRUD SDK |

```bash
just sdk-install    # Install dependencies (pnpm)
just sdk-build      # Build all packages
just sdk-test       # Run all SDK tests (Vitest)
just sdk-typecheck  # Typecheck all packages
just sdk-clean      # Clean build artifacts
```

## Architecture

```
cmd/                     - Cobra CLI commands (serve, install, tui, etc.)
internal/
├── cli/                 - TUI implementation (Bubbletea)
├── router/              - REST API handlers (stdlib ServeMux)
├── db/                  - Database interface (DbDriver, wrapper structs)
├── db/types/            - ULID-based typed IDs, enums, field configs
├── db/audited/          - Audited command pattern for change events
├── db-sqlite/           - SQLite driver (sqlc-generated, do not edit)
├── db-mysql/            - MySQL driver (sqlc-generated, do not edit)
├── db-psql/             - PostgreSQL driver (sqlc-generated, do not edit)
├── model/               - Domain structs (Root, Node, Datatype, Field)
├── auth/                - OAuth authentication (Google/GitHub/Azure)
├── backup/              - Backup/restore (SQL dump + media, local or S3)
├── bucket/              - S3 storage integration
├── config/              - Configuration management
├── media/               - Image optimization, preset dimensions, S3 upload
├── middleware/           - CORS, rate limiting, sessions, audit logging
├── plugin/              - Lua plugin system (gopher-lua)
└── utility/             - Logging (slog), version info, helpers
sql/
└── schema/              - Numbered schema directories (DDL + sqlc queries)
sdks/
└── typescript/          - pnpm workspace (Node 22+, pnpm 9+)
    ├── types/           - @modulacms/types (shared entity types, branded IDs)
    ├── modulacms-sdk/   - @modulacms/sdk (read-only content delivery)
    └── modulacms-admin-sdk/ - @modulacms/admin-sdk (full admin CRUD)
```

### Key Technologies

- **Build Runner**: [just](https://github.com/casey/just)
- **TUI Framework**: [Charmbracelet Bubbletea](https://github.com/charmbracelet/bubbletea) (Elm Architecture)
- **SSH Server**: [Charmbracelet Wish](https://github.com/charmbracelet/wish)
- **Forms**: [Charmbracelet Huh](https://github.com/charmbracelet/huh)
- **Styling**: [Charmbracelet Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Database**: [sqlc](https://sqlc.dev/) for type-safe SQL
- **Plugins**: [gopher-lua](https://github.com/yuin/gopher-lua)
- **SDK Build**: [tsup](https://tsup.egoist.dev/) (dual ESM+CJS), [Vitest](https://vitest.dev/) for tests
- **SDK Workspace**: [pnpm](https://pnpm.io/) workspace monorepo

## Documentation

Additional documentation is available in the `ai/` directory:

- [Architecture](ai/architecture/) - System design and patterns
- [API Contract](ai/api/API_CONTRACT.md) - Complete REST API reference
- [Workflows](ai/workflows/) - Development guides
- [Packages](ai/packages/) - Internal package documentation
- [Reference](ai/reference/) - Quick start, patterns, troubleshooting

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/name`)
3. Make your changes
4. Run linters (`just lint`)
5. Commit your changes
6. Push to the branch
7. Open a Pull Request

### Code Style

- Use `go fmt` for formatting
- Use tabs for indentation
- Follow existing naming conventions
- Run `just lint` before committing

## License

ModulaCMS is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

This means:
- You can use, modify, and distribute this software freely
- If you modify ModulaCMS and run it on a server, you **must** make your modified source code available
- Any derivative works must also be licensed under AGPL-3.0
- Copyright and license notices must be preserved

See [LICENSE](LICENSE) for the full license text.

## Links

- [Repository](https://github.com/hegner123/modulacms)
- [Issue Tracker](https://github.com/hegner123/modulacms/issues)
