# ModulaCMS

A headless CMS written in Go that ships as a single binary with three concurrent servers: HTTP, HTTPS (with Let's Encrypt autocert), and SSH (running a Bubbletea terminal UI). Content is managed through the SSH TUI, web admin panel, or REST API and delivered to frontend clients over HTTP/HTTPS in multiple output formats.

## Features

- **Single Binary** -- one compiled artifact runs everything: API server, admin panel, TUI, and background services
- **Three Servers** -- HTTP, HTTPS (autocert), and SSH run concurrently with graceful shutdown
- **Admin Panel** -- server-rendered HTMX + templ web interface with block editor, tree navigation, and content versioning
- **Terminal UI** -- full content management via SSH using Charmbracelet Bubbletea with responsive layouts, panel tabs, scroll indicators, and screen modes
- **Tri-Database** -- SQLite, MySQL, and PostgreSQL supported interchangeably via a single `DbDriver` interface
- **Multi-Format Delivery** -- content responses can mimic Contentful, Sanity, Strapi, WordPress, or use a clean/raw format
- **Publishing** -- publish, unpublish, schedule, version, and restore content with snapshot-based delivery
- **i18n** -- locale management, translatable fields, locale-aware content delivery with fallback chains
- **Webhooks** -- event-driven HTTP notifications with delivery tracking, retry, and test endpoints
- **RBAC** -- role-based access control with 47+ granular permissions, three bootstrap roles, and an in-memory permission cache
- **Field Validation** -- composable validation rules enforced at HTTP, admin panel, and TUI write paths
- **Cascade Operations** -- content creation with fields, recursive delete, datatype cascade, user reassignment, and media reference cleanup
- **Lua Plugins** -- sandboxed gopher-lua VMs with database access, content lifecycle hooks, custom HTTP routes, and hot-reload
- **Typed IDs** -- 26-character ULIDs wrapped in distinct Go types for compile-time safety across all entities
- **Audit Trail** -- every mutation atomically records old/new JSON values, user, IP, and timestamps in `change_events`
- **Media Pipeline** -- S3-compatible storage with image optimization, WebP conversion, responsive dimension presets, and focal points
- **OAuth** -- pluggable OAuth providers (Google, GitHub, Azure) alongside password authentication
- **Email** -- transactional email via SMTP, SendGrid, SES, or Postmark for password resets and notifications
- **Import** -- bulk data import from Contentful, Sanity, Strapi, WordPress, or ModulaCMS clean format
- **Backup/Restore** -- ZIP archives containing SQL dumps plus media, stored locally or in S3
- **Deploy** -- content sync between environments with export, import, push, pull, and snapshot management
- **Connect** -- project registry with multi-environment support; connect locally or remotely via the Go SDK
- **MCP Server** -- 40+ tool Model Context Protocol server for AI-assisted content management
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
./modula-x86 serve

# Interactive setup wizard
./modula serve --wizard
```

On first run without a `config.json`, ModulaCMS generates one with defaults, creates the database schema, bootstraps RBAC roles (admin, editor, viewer), and logs a random admin password. The SSH server starts immediately; HTTP/HTTPS start once the database is ready.

**Default Ports:**

| Server | Port |
|--------|------|
| HTTP   | 8080 |
| HTTPS  | 8443 |
| SSH    | 2222 |

Connect to the TUI: `ssh localhost -p 2222`

Access the admin panel: `http://localhost:8080/admin/`

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

# Cross-backend DB integration tests
just docker-infra         # Start Postgres, MySQL, MinIO
just test-integration-db  # Run cross-backend tests
```

## Docker

Unified via `just dc <backend> <action>`:

```bash
# Backends: full, sqlite, mysql, postgres, prod
# Actions: up, down, reset, dev, fresh, logs, destroy (full only), minio-reset (postgres only)

just dc full up       # Full stack (CMS + all databases + MinIO)
just dc full down     # Stop containers, keep volumes
just dc full reset    # Stop containers and delete volumes
just dc full dev      # Rebuild and restart CMS container only
just dc full fresh    # Reset volumes then rebuild everything
just dc full logs     # Tail CMS logs
just dc full destroy  # Remove containers, volumes, and images

just dc sqlite up     # SQLite stack
just dc mysql up      # MySQL stack
just dc postgres up   # PostgreSQL stack
```

Other Docker commands:

```bash
just docker-infra     # Infrastructure only (Postgres, MySQL, MinIO)
just docker-build     # Build standalone CMS image (for CI)
just docker-release   # Tag and push image with version tags
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
| **i18n** | locales |
| **Webhooks** | webhooks, webhook_deliveries |
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

## Admin Panel

Server-rendered HTMX + templ web interface. No SPA -- all pages are server-rendered with HTMX for interactivity.

- **Content** -- tree navigation, block editor with drag-and-drop, inline field editing
- **Schema** -- datatypes, fields, and field-datatype associations
- **Media** -- upload, browse, image preview with dimension presets
- **Users & Roles** -- user management, role assignment, permission configuration
- **Routes** -- URL slug management
- **Plugins** -- browse, enable, disable, view details
- **Webhooks** -- create, test, view delivery history
- **Locales** -- i18n configuration and locale management
- **Settings** -- server configuration
- **Audit Log** -- change event browser
- **Import** -- bulk import from external CMS platforms
- **Sessions & Tokens** -- active session and API token management

Light DOM web components (`mcms-*`) provide dialog, data-table, field-renderer, media-picker, tree-nav, toast, confirm, and search widgets.

```bash
just admin generate      # Regenerate templ Go code
just admin watch         # Watch .templ files for changes
just admin bundle        # Bundle block editor JS via esbuild
```

## API

All endpoints are prefixed with `/api/v1/` and follow standard REST conventions. Content delivery uses slug-based routing at `/api/v1/content/{slug}`.

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
POST              /api/v1/content/create          # Create content with fields (cascade)
POST              /api/v1/content/batch           # Batch operations
POST              /api/v1/contentdata/move        # Move node in tree
POST              /api/v1/contentdata/reorder     # Reorder siblings

GET|POST          /api/v1/contentfields           # Content field values
GET|POST          /api/v1/contentrelations        # Content relationships
```

### Publishing & Versioning

```
POST              /api/v1/content/publish         # Publish content (creates snapshot)
POST              /api/v1/content/unpublish       # Unpublish content
POST              /api/v1/content/schedule        # Schedule future publish
GET               /api/v1/content/versions        # List versions
POST              /api/v1/content/versions        # Create manual version
DELETE            /api/v1/content/versions/{id}   # Delete version
POST              /api/v1/content/restore         # Restore from version
```

Admin content mirrors exist at `/api/v1/admin/content/` for draft management.

### Content Delivery

```
GET               /api/v1/content/{slug}          # Published content by slug
GET               /api/v1/content/{slug}?preview=true  # Live draft (requires auth)
GET               /api/v1/content/{slug}?locale=en     # Locale-specific delivery
GET               /api/v1/content/{slug}?format=clean  # Format override
GET               /api/v1/globals                 # All global content trees
GET               /api/v1/query/{datatype}        # Query by datatype
```

The `format` query parameter controls response structure: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, or `raw`.

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

### Routes & Locales

```
GET|POST          /api/v1/routes                 # Route management
GET               /api/v1/locales                # Public locale list
CRUD              /api/v1/admin/locales          # Admin locale management
```

### Users & Access Control

```
GET|POST          /api/v1/users                  # User management
POST              /api/v1/users/reassign-delete  # Reassign content and delete user
GET|POST          /api/v1/roles                  # Roles
GET|POST          /api/v1/permissions            # Permissions
GET|POST          /api/v1/role-permissions       # Role-permission mappings
GET|POST|DELETE   /api/v1/ssh-keys               # SSH key management
GET|POST|DELETE   /api/v1/sessions               # Session management
GET|POST|DELETE   /api/v1/tokens                 # API token management
```

### Webhooks

```
CRUD              /api/v1/admin/webhooks         # Webhook management
POST              /api/v1/admin/webhooks/{id}/test       # Test delivery
GET               /api/v1/admin/webhooks/{id}/deliveries # Delivery history
POST              /api/v1/admin/webhooks/deliveries/{id}/retry # Retry delivery
```

### Import & Configuration

```
POST   /api/v1/import/contentful   # Import from Contentful
POST   /api/v1/import/sanity       # Import from Sanity
POST   /api/v1/import/strapi       # Import from Strapi
POST   /api/v1/import/wordpress    # Import from WordPress
POST   /api/v1/import/clean        # Import ModulaCMS format
POST   /api/v1/import              # Bulk import

GET               /api/v1/admin/config           # Get config (redacted)
PATCH             /api/v1/admin/config           # Update config
GET               /api/v1/admin/config/meta      # Config field metadata
GET               /api/v1/admin/plugins          # List plugins
GET               /api/v1/admin/plugins/routes   # Plugin route approval
```

## Terminal UI

The SSH-accessible TUI is built with Charmbracelet Bubbletea following the Elm Architecture (Model-Update-View). Each screen implements a `Screen` interface with its own state, update, and view methods.

- **Content** -- browse, create, edit content with tree navigation (regular and admin views)
- **Datatypes & Fields** -- define and manage content schemas
- **Media** -- upload and manage media assets with file picker
- **Users** -- user management with role assignment
- **Routes** -- URL slug configuration
- **Plugins** -- browse, enable, disable, reload Lua plugins
- **Webhooks** -- webhook management and delivery monitoring
- **Pipelines** -- pipeline entry management
- **Deploy** -- content sync between environments
- **Configuration** -- edit server configuration
- **Database** -- database info and table browser
- **Quick Start** -- guided setup wizard

UI features: responsive panel layouts with three screen modes (normal/wide/full), accordion focus, panel tabs, scroll indicators, adaptive statusbar, compact/full header with breadcrumbs.

## Connect System

The `connect` command provides a project registry for managing multiple CMS instances and environments from a single CLI. Each project can have multiple environments (local, dev, staging, prod), each pointing to a different `config.json`. The registry lives at `~/.modula/configs.json`.

### Local vs Remote

When a config has `db_driver` set, the TUI connects directly to the local database. When a config has `remote_url` and `remote_api_key` set instead, the TUI connects to a remote CMS server over HTTPS using the Go SDK as a `DbDriver` implementation. This means the same TUI works identically whether managing a local database or a remote server -- the `RemoteDriver` in `internal/remote/` implements the full `DbDriver` interface by delegating to SDK calls.

Remote connections include:
- **Health check** on startup (fails fast if server unreachable)
- **Connection tracking** with atomic status (connected/disconnected/unknown), reflected in the TUI status bar
- **Retry logic** for transient failures (502/503/504, timeouts) with a single retry after 1s delay
- **Graceful degradation** -- DDL and infrastructure methods return `ErrNotSupported` or `ErrRemoteMode`

### Registry Management

```bash
# Register a project environment
modula connect set mysite local ./config.json
modula connect set mysite prod /srv/mysite/config.json

# Connect to a project
modula connect                      # default project, default env
modula connect mysite               # mysite, default env
modula connect mysite prod          # mysite, prod env

# Manage defaults
modula connect default mysite       # set default project
modula connect default mysite prod  # set default env for project

# List and remove
modula connect list                 # show all projects and environments
modula connect remove mysite        # remove entire project
modula connect remove mysite --env dev  # remove single environment
```

### Resolution Order

1. Both name and env given: use that exact project + environment
2. Only name given: use that project's default environment
3. Neither given: use the default project's default environment
4. Registry empty: look for `config.json` in the current directory

### Config Fields

| Field | Purpose |
|-------|---------|
| `remote_url` | Base URL of the remote CMS server (e.g. `https://cms.example.com`) |
| `remote_api_key` | API key for authenticating SDK calls to the remote server |
| `db_driver` | Local database driver (`sqlite`, `mysql`, `postgres`) -- mutually exclusive with `remote_url` |

A config must have either `remote_url` or `db_driver` set (not both).

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
./modula plugin list               # List plugins
./modula plugin init my-plugin     # Create scaffold
./modula plugin validate ./path    # Validate structure
./modula plugin reload my-plugin   # Hot-reload (requires running server)
./modula plugin enable my-plugin   # Enable
./modula plugin disable my-plugin  # Disable
```

## Deploy

The `deploy` command syncs content between environments via JSON exports and the Go SDK.

```bash
modula deploy export -o export.json           # Export content to JSON
modula deploy import -f export.json           # Import from JSON
modula deploy pull --env prod                 # Download from remote
modula deploy push --env staging              # Upload to remote
modula deploy snapshot list                   # List import snapshots
modula deploy snapshot show <id>              # Show snapshot details
modula deploy snapshot restore <id>           # Restore from snapshot
modula deploy env list                        # List configured environments
modula deploy env test                        # Test environment connectivity
```

## MCP Server

ModulaCMS includes a Model Context Protocol server with 40+ tools for AI-assisted content management. The MCP server connects to the CMS via the Go SDK.

```bash
just mcp-build        # Build MCP server binary
just mcp-install      # Build and install to /usr/local/bin/modula-mcp
```

Tools cover content CRUD, content fields, batch operations, schema management, media, routes, users, roles, permissions, configuration, and import.

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
| **i18n** | `i18n_enabled`, `i18n_default_locale`, `i18n_fallback_chain` |
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
just sdk ts install      # pnpm install (workspace root)
just sdk ts build        # Build all packages
just sdk ts test         # Run all SDK tests (Vitest)
just sdk ts typecheck    # Typecheck all packages
```

### Go

The Go SDK provides a type-safe client with a generic `Resource[Entity, CreateParams, UpdateParams, ID]` pattern for CRUD operations.

**Import:** `github.com/hegner123/modulacms/sdks/go`

```go
import modula "github.com/hegner123/modulacms/sdks/go"

client, err := modula.NewClient(modula.ClientConfig{
    BaseURL: "https://cms.example.com",
    APIKey:  "your-api-key",
})

// Authentication
me, err := client.Auth.Me(ctx)

// CRUD with typed IDs and pagination
users, err := client.Users.ListPaginated(ctx, modula.PaginationParams{Limit: 20, Offset: 0})
content, err := client.ContentData.Create(ctx, modula.CreateContentDataParams{...})

// Media upload
media, err := client.MediaUpload.Upload(ctx, file, "photo.jpg", &modula.MediaUploadOptions{
    Path: "blog/headers",
})

// Content delivery
page, err := client.Content.GetPage(ctx, "blog/hello-world", "clean")

// Error classification
if modula.IsNotFound(err) { ... }
if modula.IsUnauthorized(err) { ... }
```

The SDK exposes 23+ typed resource endpoints including content, schema, media, users, roles, permissions, plugins, configuration, sessions, SSH keys, and bulk import.

```bash
just sdk go test      # Run Go SDK tests
just sdk go vet       # Vet Go SDK
```

### Swift

The Swift SDK is a zero-dependency Swift Package Manager package supporting Apple platforms.

**Platforms:** iOS 16+, macOS 13+, tvOS 16+, watchOS 9+
**Swift:** 5.9+

```swift
import Modula

let client = try ModulaClient(config: ClientConfig(
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
just sdk swift build  # Build Swift SDK
just sdk swift test   # Run Swift SDK tests
just sdk swift clean  # Clean build artifacts
```

## Backup & Restore

```bash
./modula backup create              # Create full backup (ZIP with SQL dump + metadata)
./modula backup restore backup.zip  # Restore from backup
./modula backup list                # List backup history
./modula backup delete {id}         # Delete backup record
```

Backups are ZIP archives containing a database-specific SQL dump and a JSON manifest with driver, timestamp, version, and node ID. Storage is local (`backups/` directory) or S3.

## CLI Commands

```
modula [--config=path] [--verbose] <command>

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
  connect            Launch TUI for a registered project (default project + env)
  connect <name>     Launch TUI for named project (default env)
  connect <name> <env>  Launch TUI for named project + specific env
  connect set        Register a project environment
  connect list       List registered projects and environments
  connect remove     Remove a project or environment
  connect default    Set the default project or environment
  deploy export      Export content data to JSON file
  deploy import      Import content from JSON export
  deploy pull        Download data from remote environment
  deploy push        Upload local data to remote environment
  deploy snapshot    List, show, or restore import snapshots
  deploy env         List and test configured deploy environments
  pipeline list      Show all pipeline entries
  pipeline show      Show pipelines for a specific table
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
cmd/                          Cobra CLI commands (serve, install, tui, connect, deploy, pipeline, db, backup, config, cert, plugin, version)
internal/
  admin/                      HTMX admin panel: CSRF, auth middleware, static file embed, web components
  admin/handlers/             Admin page handlers (render, auth, CRUD for all resources)
  admin/pages/                templ full-page components (~28 pages)
  admin/static/               CSS, JS, block editor, web components (go:embed)
  auth/                       Authentication (bcrypt, OAuth, sessions)
  backup/                     Backup/restore (SQL dump + ZIP)
  config/                     Configuration loading, validation, hot-reload
  db/                         DbDriver interface, wrapper structs, application types
  db/types/                   ULID-based typed IDs, enums, nullable wrappers
  db/audited/                 Audited command pattern for change events
  db-sqlite/, db-mysql/, db-psql/  sqlc-generated code (do not edit)
  definitions/                CMS format definitions (Contentful, Sanity, Strapi, WordPress)
  deploy/                     Content sync: export, import, push, pull, snapshots
  email/                      Transactional email (SMTP, SendGrid, SES, Postmark)
  install/                    Setup wizard and bootstrap
  media/                      Image optimization, S3 upload
  middleware/                  CORS, rate limiting, sessions, RBAC authorization
  model/                      Domain structs (Node, Datatype, Field)
  plugin/                     Lua plugin system (gopher-lua)
  publishing/                 Publish, snapshot, version management
  registry/                   Project registry (~/.modula/configs.json)
  remote/                     RemoteDriver -- DbDriver over Go SDK (HTTPS)
  router/                     HTTP route registration, slug handling, globals, pagination
  service/                    Service layer (SchemaService with composable store interfaces)
  transform/                  Content format transformers
  tui/                        Bubbletea TUI (Screen interface, PanelScreen base, 12+ screens)
  utility/                    Logging (slog), version info, helpers
  validation/                 Composable field validation rules
mcp/                          MCP server (40+ tools for AI-assisted CMS management)
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
