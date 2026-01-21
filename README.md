# ModulaCMS

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
- Linux or macOS (Windows is not currently supported)
- For production: MySQL or PostgreSQL recommended

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/hegner123/modulacms.git
cd modulacms

# Build for development
make dev

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

### Build Commands

```bash
make dev        # Build local binary (./modulacms-x86)
make build      # Build production binaries (x86 + AMD64)
make run        # Build and run
make clean      # Remove build artifacts
```

### Testing

> **Note:** Test suite is a work in progress. Coverage is incomplete.

```bash
make test              # Run all tests
make test-development  # Run development package tests
make coverage          # Run tests with coverage report
```

### Database

```bash
make sqlc      # Generate Go code from SQL queries
make dump      # Dump SQLite database to SQL file
make docker-db # Start database containers
```

### Code Quality

```bash
make lint      # Run all linters
make lint-go   # Lint Go code
make vendor    # Update vendor directory
```

## Architecture

```
cmd/main.go              - Application entry point
internal/
├── cli/                 - TUI implementation (Bubbletea)
├── router/              - REST API handlers
├── db/                  - Database interface
├── db-sqlite/           - SQLite driver
├── db-mysql/            - MySQL driver
├── db-psql/             - PostgreSQL driver
├── model/               - Business logic
├── auth/                - OAuth authentication
├── backup/              - Backup/restore
├── bucket/              - S3 storage integration
├── config/              - Configuration management
├── media/               - Media processing
├── middleware/          - HTTP/SSH middleware
├── plugin/              - Lua plugin system
└── utility/             - Shared utilities
sql/
├── schema/              - Database migrations
├── mysql/               - MySQL queries
└── postgres/            - PostgreSQL queries
```

### Key Technologies

- **TUI Framework**: [Charmbracelet Bubbletea](https://github.com/charmbracelet/bubbletea) (Elm Architecture)
- **SSH Server**: [Charmbracelet Wish](https://github.com/charmbracelet/wish)
- **Forms**: [Charmbracelet Huh](https://github.com/charmbracelet/huh)
- **Styling**: [Charmbracelet Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Database**: [sqlc](https://sqlc.dev/) for type-safe SQL
- **Plugins**: [gopher-lua](https://github.com/yuin/gopher-lua)

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
4. Run linters (`make lint`)
5. Commit your changes
6. Push to the branch
7. Open a Pull Request

### Code Style

- Use `go fmt` for formatting
- Use tabs for indentation
- Follow existing naming conventions
- Run `make lint` before committing

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
