# ModulaCMS

A headless CMS written in Go that ships as a single binary with three concurrent servers: HTTP, HTTPS (with Let's Encrypt autocert), and SSH (running a Bubbletea terminal UI). Content is managed through the SSH TUI or REST API and delivered to frontend clients over HTTP/HTTPS in multiple output formats.

## Features

- **Single Binary** -- one compiled artifact runs everything: API server, admin TUI, and background services
- **Three Servers** -- HTTP, HTTPS (autocert), and SSH run concurrently with graceful shutdown
- **Terminal UI** -- full content management via SSH using Charmbracelet Bubbletea (Elm Architecture)
- **Tri-Database** -- SQLite, MySQL, and PostgreSQL supported interchangeably via a single `DbDriver` interface
- **Multi-Format Delivery** -- content responses can mimic Contentful, Sanity, Strapi, WordPress, or use a clean/raw format
- **RBAC** -- role-based access control with 47 granular permissions, three bootstrap roles, and an in-memory permission cache
- **Lua Plugins** -- sandboxed gopher-lua VMs with database access, content lifecycle hooks, custom HTTP routes, and hot-reload
- **Typed IDs** -- 26-character ULIDs wrapped in distinct Go types for compile-time safety across all entities
- **Audit Trail** -- every mutation atomically records old/new JSON values, user, IP, and timestamps in `change_events`
- **Media Pipeline** -- S3-compatible storage with image optimization, WebP conversion, responsive dimension presets, and focal points
- **OAuth** -- pluggable OAuth providers (Google, GitHub, Azure) alongside password authentication
- **Import** -- bulk data import from Contentful, Sanity, Strapi, WordPress, or ModulaCMS clean format
- **Backup/Restore** -- ZIP archives containing SQL dumps plus media, stored locally or in S3
- **SDKs** -- official TypeScript, Go, and Swift SDKs with zero external dependencies

## Requirements

- Go 1.24+ with CGO enabled (for SQLite via `mattn/go-sqlite3`)
- Linux or macOS
- [just](https://github.com/casey/just) as the build runner

## Quick Start

```bash
# Build and run (auto-creates config.json with defaults on first run)
just run

# Or build a local binary
just dev
./modulacms-x86 serve

# Interactive setup wizard
./modulacms serve --wizard
```

On first run without a `config.json`, ModulaCMS generates one with defaults, creates the database schema, bootstraps RBAC roles (admin, editor, viewer), and logs a random admin password. The SSH server starts immediately; HTTP/HTTPS start once the database is ready.

**Default Ports:**

| Server | Port |
|--------|------|
| HTTP   | 8080 |
| HTTPS  | 8443 |
| SSH    | 2222 |

Connect to the TUI: `ssh localhost -p 2222`

## Build & Development

```bash
just dev              # Build local binary with version info via ldflags
just run              # Build and run
just build            # Production binary to out/bin/
just check            # Compile-check without producing artifacts
just clean            # Remove build artifacts
just vendor           # Update vendor directory
```

## Testing

```bash
just test             # Run all tests (creates/cleans testdb/ and backups/)
just coverage         # Tests with coverage report
just lint             # Run all linters (Go, Dockerfile, YAML)

# Single package or test
go test -v ./internal/db
go test -v ./internal/db -run TestSpecificName

# S3 integration tests (requires MinIO)
just test-minio       # Start MinIO container
just test-integration # Run integration tests
just test-minio-down  # Stop MinIO
```

## Docker

```bash
just docker-up        # Full stack (CMS + all databases + MinIO)
just docker-dev       # Rebuild and restart CMS container only
just docker-infra     # Infrastructure only (Postgres, MySQL, MinIO)
just docker-down      # Stop containers, keep volumes
just docker-reset     # Stop containers and delete volumes
```

Per-database stacks:

```bash
just docker-sqlite-up   / just docker-sqlite-down   / just docker-sqlite-reset
just docker-mysql-up    / just docker-mysql-down     / just docker-mysql-reset
just docker-postgres-up / just docker-postgres-down  / just docker-postgres-reset
```

The Docker image exposes ports 8080 (HTTP), 4000 (HTTPS), and 2233 (SSH), with volumes for `/app/data`, `/app/certs`, `/app/.ssh`, `/app/backups`, and `/app/plugins`.

## Architecture

### Runtime

The `serve` command starts three concurrent servers sharing a single `DbDriver` instance:

```
HTTP  (default :8080)  ─┐
HTTPS (default :8443)  ─┤── stdlib ServeMux (Go 1.22+) ── Middleware Chain ── Handlers ── DbDriver
SSH   (default :2222)  ─┘   Charmbracelet Wish ── Bubbletea TUI ─────────────────────────── DbDriver
```

Graceful shutdown: first SIGINT/SIGTERM triggers a 30-second shutdown; second signal forces exit. Shutdown order: HTTP servers, plugin system, database connections.

### Request Flow

```
Client Request
  -> Request ID
  -> Logging
  -> CORS
  -> Authentication (cookie session or Bearer API key)
  -> Rate Limiting (auth endpoints: 10 req/min per IP)
  -> Permission Injection (RBAC)
  -> Route Handler
  -> DbDriver Interface
  -> Database-specific wrapper (SQLite / MySQL / PostgreSQL)
  -> sqlc-generated queries
```

### Tri-Database Pattern

One codebase supports three databases through a layered abstraction:

1. **SQL schemas** in `sql/schema/` define tables and queries per dialect (SQLite, MySQL, PostgreSQL)
2. **sqlc** generates type-safe Go code into `internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/`
3. **`DbDriver` interface** (~150 methods in `internal/db/db.go`) provides the contract
4. **Wrapper structs** (`Database`, `MysqlDatabase`, `PsqlDatabase`) implement the interface, converting between sqlc types and application types

Switch databases by setting `db_driver` in `config.json` to `"sqlite"`, `"mysql"`, or `"postgres"`.

### Content Model

Content uses a tree structure with sibling pointers for O(1) navigation and reordering:

- `parent_id` -- parent node
- `first_child_id` -- leftmost child
- `next_sibling_id` / `prev_sibling_id` -- doubly-linked sibling list

Content items have a status lifecycle: **draft** -> **pending** -> **published** -> **archived**.

### Data Model

27 schema directories define the full entity model:

| Entity Group | Tables |
|-------------|--------|
| **Content** | content_data, content_fields, content_relations, admin variants |
| **Schema** | datatypes, fields, datatype_fields, admin variants |
| **Media** | media, media_dimensions |
| **Routing** | routes, admin_routes |
| **Users & Auth** | users, roles, permissions, role_permissions, tokens, user_oauth, sessions, user_ssh_keys |
| **System** | backups, change_events, tables |

All entity IDs are 26-character ULIDs wrapped in distinct Go types (`ContentID`, `UserID`, `FieldID`, etc.) that provide compile-time type safety.

### RBAC Authorization

Role-based access control with `resource:operation` granular permissions:

| Role | Permissions | Description |
|------|-------------|-------------|
| **admin** | 47 (all) | Bypasses all permission checks |
| **editor** | 28 | CRUD on content, media, routes, datatypes, fields |
| **viewer** | 3 | Read-only: content, media, routes |

The `PermissionCache` maintains an in-memory role-to-permissions map with lock-free reads and 60-second periodic refresh. System-protected roles and permissions cannot be deleted or renamed.

### Audited Commands

All database mutations are wrapped in transactions that atomically record `change_events` rows capturing:
- Operation type (INSERT, UPDATE, DELETE)
- Old and new JSON values
- User ID, request ID, IP address
- Hybrid Logical Clock timestamps for distributed ordering

## API

All admin endpoints are prefixed with `/api/v1/` and follow standard REST conventions. Public content delivery uses slug-based routing.

### Authentication

```
POST   /api/v1/auth/login          # Session login
POST   /api/v1/auth/logout         # Session logout
GET    /api/v1/auth/me             # Current user profile
POST   /api/v1/auth/register       # Registration
POST   /api/v1/auth/reset          # Password reset
GET    /api/v1/auth/oauth/login    # OAuth flow initiation
GET    /api/v1/auth/oauth/callback # OAuth callback
```

### Content Management

```
GET|POST          /api/v1/contentdata            # List / Create
GET|PUT|DELETE    /api/v1/contentdata/{id}        # Get / Update / Delete
POST              /api/v1/content/batch           # Batch operations

GET|POST          /api/v1/contentfields           # Content field values
GET|POST          /api/v1/contentrelations        # Content relationships
```

Admin content mirrors exist at `/api/v1/admincontentdatas` and `/api/v1/admincontentfields` for draft management.

### Schema

```
GET|POST          /api/v1/datatype               # Datatypes
GET|POST          /api/v1/fields                 # Field definitions
GET|POST          /api/v1/datatypefields         # Datatype-field associations
GET|POST          /api/v1/tables                 # Custom tables
```

### Media

```
GET               /api/v1/media                  # List (paginated)
POST              /api/v1/media                  # Upload (multipart/form-data)
DELETE            /api/v1/media/{id}             # Delete
GET               /api/v1/media/health           # S3 connectivity check
DELETE            /api/v1/media/cleanup          # Remove orphaned S3 objects
GET|POST          /api/v1/mediadimensions        # Dimension presets
```

### Routes & Content Delivery

```
GET|POST          /api/v1/routes                 # Route management
GET               /{slug}                        # Public content delivery (format via query param)
GET               /api/v1/admin/tree/            # Admin content tree
```

The `format` query parameter controls response structure: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, or `raw`.

### Users & Access Control

```
GET|POST          /api/v1/users                  # User management
GET|POST          /api/v1/roles                  # Roles
GET|POST          /api/v1/permissions            # Permissions
GET|POST          /api/v1/role-permissions       # Role-permission mappings
GET|POST|DELETE   /api/v1/ssh-keys               # SSH key management
GET|POST|DELETE   /api/v1/sessions               # Session management
GET|POST|DELETE   /api/v1/tokens                 # API token management
```

### Import

```
POST   /api/v1/import/contentful   # Import from Contentful
POST   /api/v1/import/sanity       # Import from Sanity
POST   /api/v1/import/strapi       # Import from Strapi
POST   /api/v1/import/wordpress    # Import from WordPress
POST   /api/v1/import/clean        # Import ModulaCMS format
POST   /api/v1/import              # Bulk import
```

### Configuration & Plugins

```
GET               /api/v1/admin/config           # Get config (redacted)
PATCH             /api/v1/admin/config           # Update config
GET               /api/v1/admin/config/meta      # Config field metadata
GET               /api/v1/admin/plugins          # List plugins
GET               /api/v1/admin/plugins/routes   # Plugin route approval
GET               /api/v1/admin/plugins/hooks    # Plugin hook approval
```

## Terminal UI

The SSH-accessible TUI is built with Charmbracelet Bubbletea following the Elm Architecture (Model-Update-View). It provides 26+ screens for managing all CMS operations:

- **Content** -- browse, create, edit content with tree navigation
- **Datatypes & Fields** -- define and manage content schemas
- **Media** -- upload and manage media assets with file picker
- **Users** -- user management with role assignment
- **Routes** -- URL slug configuration
- **Plugins** -- browse, enable, disable, reload Lua plugins
- **Configuration** -- edit server configuration
- **Quick Start** -- guided setup wizard

The TUI uses a focus system (page, table, form, dialog) for keyboard input routing, custom form dialogs for data entry, and async commands for database operations.

## Lua Plugin System

Plugins extend ModulaCMS with sandboxed Lua scripts via gopher-lua.

### Plugin Structure

```
plugins/my-plugin/
  init.lua          # Entry point with plugin_info table
```

### Capabilities

- **Database** -- query builder with safe identifier validation, isolated per-plugin tables
- **Content Hooks** -- before/after hooks on create, update, delete, publish, archive
- **HTTP Routes** -- register custom endpoints with approval workflow
- **Logging** -- structured logging integrated with CMS slog
- **Schema** -- define custom tables with drift detection

### Safety

- Operation counting per VM checkout (default 1000 ops)
- Per-hook timeout (default 2000ms) and per-event timeout (default 5000ms)
- Circuit breaker: max consecutive failures trips the plugin
- Connection pool limits and request/response size limits
- Plugins cannot access core CMS tables or other plugins' tables
- Hot-reload with file watcher (opt-in for production)

### CLI

```bash
./modulacms plugin list               # List plugins
./modulacms plugin init my-plugin     # Create scaffold
./modulacms plugin validate ./path    # Validate structure
./modulacms plugin reload my-plugin   # Hot-reload (requires running server)
./modulacms plugin enable my-plugin   # Enable
./modulacms plugin disable my-plugin  # Disable
```

## Configuration

Configuration lives in `config.json` at the project root. Environment variables can be referenced as `${VAR}` or `${VAR:-default}`.

Key configuration categories:

| Category | Fields |
|----------|--------|
| **Database** | `db_driver`, `db_url`, `db_name`, `db_user`, `db_password` |
| **Server** | `port`, `ssl_port`, `ssh_port`, `environment`, `cert_dir` |
| **Auth** | `auth_salt`, `cookie_*`, `oauth_*` |
| **S3 Storage** | `bucket_endpoint`, `bucket_media`, `bucket_backup`, `bucket_access_key`, `bucket_secret_key` |
| **Email** | `email_enabled`, `email_provider` (smtp/sendgrid/ses/postmark), `email_from_*` |
| **CORS** | `cors_origins`, `cors_methods`, `cors_headers`, `cors_credentials` |
| **Output** | `output_format` (contentful/sanity/strapi/wordpress/clean/raw) |
| **Plugins** | `plugin_enabled`, `plugin_directory`, `plugin_max_vms`, `plugin_timeout`, `plugin_hot_reload` |
| **Observability** | `observability_enabled`, `observability_provider` (sentry/datadog/newrelic), `observability_dsn` |

Runtime configuration can be updated via the REST API (`PATCH /api/v1/admin/config`) with hot-reload support for applicable fields.

## SQL Code Generation

```bash
just sqlc    # Regenerate Go code from SQL queries
```

This runs `sqlc generate` against `sql/sqlc.yml`, producing type-safe Go code in three packages:
- `internal/db-sqlite/` (package `mdb`)
- `internal/db-mysql/` (package `mdbm`)
- `internal/db-psql/` (package `mdbp`)

Never edit files in these directories by hand -- they are overwritten by sqlc. After modifying schema or queries, run `just sqlc`, then update the `DbDriver` interface and implement methods on all three wrapper structs.

## SDKs

ModulaCMS provides official SDKs for TypeScript, Go, and Swift. All SDKs have zero external dependencies beyond their respective standard libraries.

### TypeScript

The `sdks/typescript/` directory is a pnpm workspace monorepo with three packages:

| Package | npm | Purpose |
|---------|-----|---------|
| `types/` | `@modulacms/types` | Shared entity types, 30 branded IDs, enums |
| `modulacms-sdk/` | `@modulacms/sdk` | Read-only content delivery client |
| `modulacms-admin-sdk/` | `@modulacms/admin-sdk` | Full admin CRUD client |

**Requirements:** TypeScript 5.7+, Node 18+ (Fetch API), pnpm 9+

**Install:**
```bash
npm install @modulacms/sdk             # Content delivery
npm install @modulacms/admin-sdk       # Admin operations
```

**Content Delivery:**
```typescript
import { ModulaClient } from "@modulacms/sdk"

const cms = new ModulaClient({
  baseUrl: "https://cms.example.com",
  defaultFormat: "clean",
})

const page = await cms.getPage("blog/hello-world")
const media = await cms.listMedia()
```

**Admin SDK:**
```typescript
import { createAdminClient } from "@modulacms/admin-sdk"

const client = createAdminClient({
  baseUrl: "https://cms.example.com",
  apiKey: "your-api-key",
})

await client.auth.login({ email: "admin@example.com", password: "pass" })
const users = await client.users.list()
const content = await client.contentData.create({ status: "draft", ... })
await client.mediaUpload.upload(file)
```

Both SDKs ship as dual ESM + CommonJS builds via tsup with full type declarations.

```bash
just sdk-install      # pnpm install (workspace root)
just sdk-build        # Build all packages
just sdk-test         # Run all SDK tests (Vitest)
just sdk-typecheck    # Typecheck all packages
```

### Go

The Go SDK provides a type-safe client with a generic `Resource[Entity, CreateParams, UpdateParams, ID]` pattern for CRUD operations.

**Import:** `github.com/hegner123/modulacms/sdks/go`

```go
import modulacms "github.com/hegner123/modulacms/sdks/go"

client, err := modulacms.NewClient(modulacms.ClientConfig{
    BaseURL: "https://cms.example.com",
    APIKey:  "your-api-key",
})

// Authentication
me, err := client.Auth.Me(ctx)

// CRUD with typed IDs and pagination
users, err := client.Users.ListPaginated(ctx, modulacms.PaginationParams{Limit: 20, Offset: 0})
content, err := client.ContentData.Create(ctx, modulacms.CreateContentDataParams{...})

// Media upload
media, err := client.MediaUpload.Upload(ctx, file, "photo.jpg", &modulacms.MediaUploadOptions{
    Path: "blog/headers",
})

// Content delivery
page, err := client.Content.GetPage(ctx, "blog/hello-world", "clean")

// Error classification
if modulacms.IsNotFound(err) { ... }
if modulacms.IsUnauthorized(err) { ... }
```

The SDK exposes 23+ typed resource endpoints including content, schema, media, users, roles, permissions, plugins, configuration, sessions, SSH keys, and bulk import.

```bash
just sdk-go-test      # Run Go SDK tests
just sdk-go-vet       # Vet Go SDK
```

### Swift

The Swift SDK is a zero-dependency Swift Package Manager package supporting Apple platforms.

**Platforms:** iOS 16+, macOS 13+, tvOS 16+, watchOS 9+
**Swift:** 5.9+

```swift
import ModulaCMS

let client = try ModulaCMSClient(config: ClientConfig(
    baseURL: "https://cms.example.com",
    apiKey: "your-api-key"
))

// Authentication
let login = try await client.auth.login(params: LoginParams(
    email: "admin@example.com",
    password: "pass"
))

// CRUD with branded IDs
let users = try await client.users.listPaginated(params: PaginationParams(limit: 20, offset: 0))
let content = try await client.contentData.create(params: CreateContentDataParams(...))

// Media upload
let media = try await client.mediaUpload.upload(
    data: imageData,
    filename: "photo.jpg",
    options: MediaUploadResource.UploadOptions(path: "photos")
)

// Content delivery
let page = try await client.content.getPage(slug: "blog/hello-world", format: "clean")

// Error handling
do {
    let user = try await client.users.get(id: userID)
} catch let error as APIError where isNotFound(error) {
    print("User not found")
}
```

The SDK uses async/await throughout, marks all types as `Sendable` for actor isolation, and provides 30 branded ID types via a `ResourceID` protocol. It covers 26 CRUD resources plus 13 specialized endpoints (auth, media upload, admin tree, plugins, config, etc.).

```bash
just sdk-swift-build  # Build Swift SDK
just sdk-swift-test   # Run Swift SDK tests
just sdk-swift-clean  # Clean build artifacts
```

## Backup & Restore

```bash
./modulacms backup create              # Create full backup (ZIP with SQL dump + metadata)
./modulacms backup restore backup.zip  # Restore from backup
./modulacms backup list                # List backup history
./modulacms backup delete {id}         # Delete backup record
```

Backups are ZIP archives containing a database-specific SQL dump and a JSON manifest with driver, timestamp, version, and node ID. Storage is local (`backups/` directory) or S3.

## CLI Commands

```
modulacms [--config=path] [--verbose] <command>

  serve              Start HTTP/HTTPS/SSH servers
  serve --wizard     Interactive setup before starting
  install            Interactive installation wizard
  install --yes      Non-interactive with defaults
  tui                Launch terminal UI standalone
  db init            Initialize database and bootstrap data
  db wipe            Drop all tables
  backup create      Create full backup
  backup restore     Restore from backup
  backup list        List backup history
  config show        Print config as JSON
  config validate    Validate config.json
  config set         Update config field
  cert generate      Generate self-signed certificates
  cert check         Verify certificate validity
  update check       Check for updates
  update install     Install new version
  plugin list        List plugins
  plugin init        Create plugin scaffold
  plugin validate    Validate plugin structure
  plugin reload      Hot-reload plugin
  plugin enable      Enable plugin
  plugin disable     Disable plugin
  version            Show version info
```

## CI/CD

Two GitHub Actions workflows:

- **Go** (`.github/workflows/go.yml`) -- runs on Go source changes (excludes `sdks/**`), tests with libwebp-dev on Ubuntu
- **SDKs** (`.github/workflows/sdks.yml`) -- runs on SDK changes, tests TypeScript (pnpm + Vitest), Go SDK, and Swift SDK (SPM build + test)

## Project Structure

```
cmd/                          Cobra CLI commands (serve, install, tui, db, backup, config, cert, plugin, version)
internal/
  auth/                       Authentication (bcrypt, OAuth, sessions)
  backup/                     Backup/restore (SQL dump + ZIP)
  cli/                        Bubbletea TUI (40+ files, Elm Architecture)
  config/                     Configuration loading, validation, hot-reload
  db/                         DbDriver interface, wrapper structs, application types
  db/types/                   ULID-based typed IDs, enums, nullable wrappers
  db/audited/                 Audited command pattern for change events
  db-sqlite/, db-mysql/, db-psql/  sqlc-generated code (do not edit)
  definitions/                CMS format definitions (Contentful, Sanity, Strapi, WordPress)
  install/                    Setup wizard and bootstrap
  media/                      Image optimization, S3 upload
  middleware/                  CORS, rate limiting, sessions, RBAC authorization
  model/                      Domain structs (Node, Datatype, Field)
  plugin/                     Lua plugin system (gopher-lua)
  router/                     HTTP route registration, slug handling, pagination
  transform/                  Content format transformers
  utility/                    Logging (slog), version info, helpers
sql/
  schema/                     27 numbered schema directories (DDL + queries per dialect)
  sqlc.yml                    sqlc configuration
sdks/
  typescript/
    types/                    @modulacms/types (shared types, branded IDs, enums)
    modulacms-sdk/            @modulacms/sdk (read-only content delivery)
    modulacms-admin-sdk/      @modulacms/admin-sdk (full admin CRUD)
  go/                         Go SDK (generic Resource[E,C,U,ID] pattern)
  swift/                      Swift SDK (SPM, zero dependencies, async/await)
deploy/
  docker/                     Docker Compose files per database
```

## License

See [LICENSE](LICENSE) for details.
