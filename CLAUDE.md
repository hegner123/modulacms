# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ModulaCMS is a headless CMS written in Go that runs as a single binary with three concurrent servers: HTTP, HTTPS (with Let's Encrypt autocert), and SSH (running a Bubbletea TUI). Content is managed via the SSH TUI, web admin panel, or REST API and served to frontend clients over HTTP/HTTPS.

- **Go 1.24+** required, CGO enabled (for SQLite via mattn/go-sqlite3)
- Linux/macOS only
- Uses `just` (justfile) as the build runner, not Make
- **Project Board:** [GitHub Project (Roadmap)](https://github.com/users/hegner123/projects/2/views/1)

## TypeScript SDKs

The `sdks/typescript/` directory is a pnpm workspace monorepo containing three packages:

| Package | npm Name | Purpose |
|---------|----------|---------|
| `types/` | `@modulacms/types` | Shared entity types, branded IDs, enums |
| `modulacms-sdk/` | `@modulacms/sdk` | Read-only content delivery SDK |
| `modulacms-admin-sdk/` | `@modulacms/admin-sdk` | Full admin CRUD SDK |

Both SDKs depend on `@modulacms/types` via `workspace:*`. The admin SDK uses thin re-export files in `src/types/` that re-export shared entity types from `@modulacms/types` while keeping Create/Update param types local.

```bash
just sdk ts install     # pnpm install (workspace root)
just sdk ts build       # Build all packages (types first, then SDKs in parallel)
just sdk ts test        # Run all SDK tests (Vitest)
just sdk ts typecheck   # Typecheck all packages
just sdk ts clean       # Clean build artifacts
```

Tooling: TypeScript 5.7+, tsup (ESM+CJS dual builds), Vitest, pnpm 9+, Node 22+.

**CI:** `.github/workflows/sdks.yml` runs on `sdks/typescript/**`, `sdks/go/**`, and `sdks/swift/**` changes. The Go CI workflow (`.github/workflows/go.yml`) ignores `sdks/**` changes.

## Go SDK

The `sdks/go/` directory contains the Go SDK (`package modulacms`).

| File | Purpose |
|------|---------|
| `modulacms.go` | Client, ClientConfig, NewClient |
| `resource.go` | Generic `Resource[E, C, U, ID]` CRUD |
| `httpclient.go` | Internal HTTP transport |
| `types.go` | All entity structs + Create/Update params |
| `ids.go` | Branded ID types (ContentID, UserID, etc.) |
| `enums.go` | ContentStatus, FieldType, RouteType |
| `errors.go` | ApiError, IsNotFound, IsUnauthorized |
| `pagination.go` | PaginationParams, PaginatedResponse[T] |
| `auth.go` | AuthResource (login, logout, me, register, reset) |
| `media_upload.go` | Multipart file upload |
| `admin_tree.go` | Admin content tree |
| `content_delivery.go` | Public slug-based content delivery |
| `ssh_keys.go` | SSH key management |
| `sessions.go` | Session management |
| `import.go` | CMS data import |
| `content_batch.go` | Batch content updates |

Import path: `import modulacms "github.com/hegner123/modulacms/sdks/go"`

```bash
just sdk go test        # Run Go SDK tests
just sdk go vet         # Vet Go SDK
```

## Swift SDK

The `sdks/swift/` directory contains the Swift SDK as a single SPM package.

| File | Purpose |
|------|---------|
| `Client.swift` | ModulaCMSClient, ClientConfig |
| `Resource.swift` | Generic `Resource<E, C, U, ID>` CRUD |
| `HTTPClient.swift` | Internal URLSession transport (30s timeout, no cookies) |
| `Types.swift` | All entity structs + Create/Update params (explicit CodingKeys) |
| `IDs.swift` | ResourceID protocol + 30 branded types (26 IDs + 4 value types) |
| `Enums.swift` | ContentStatus, FieldType, RouteType |
| `Errors.swift` | APIError, isNotFound(), isUnauthorized() |
| `Pagination.swift` | PaginationParams, PaginatedResponse<T> |
| `JSON.swift` | Shared JSONDecoder/Encoder + JSONValue enum |
| `AuthResource.swift` | Login, logout, me, register, resetPassword |
| `MediaUploadResource.swift` | Multipart file upload (data + filename, no mimeType) |
| `AdminTreeResource.swift` | Admin content tree |
| `ContentDeliveryResource.swift` | Public slug-based content delivery |
| `SSHKeysResource.swift` | SSH key management |
| `SessionsResource.swift` | Session management |
| `ImportResource.swift` | Import from various CMS formats |
| `ContentBatchResource.swift` | Batch content updates |

Platforms: iOS 16+, macOS 13+, tvOS 16+, watchOS 9+. Swift 5.9+, zero dependencies.

```bash
just sdk swift build    # Build Swift SDK
just sdk swift test     # Run Swift SDK tests
just sdk swift clean    # Clean Swift SDK build artifacts
```

## Build & Development Commands

```bash
just dev                # Build local binary (modula-x86) with version info via ldflags
just run                # Build and run (depends on dev)
just run-admin          # Build and run with air hot reload (rebuilds on .go and .templ changes)
just dev-admin          # Build with -tags dev for live static assets (CSS/JS from disk, no embed)
just build              # Build production binary to out/bin/
just check              # Compile-check cmd and internal packages without producing artifacts
just clean              # Remove build artifacts
just vendor             # Update vendor directory
just watch              # Run with cosmtrek/air in Docker for automatic reload
```

## Testing

```bash
just test               # Run all tests (creates/cleans testdb/*.db and backups/*.zip)
just coverage           # Run tests with coverage report
just test-integration   # S3 integration tests (requires MinIO: just test-minio first)
just test-integration-db # Cross-backend DB integration tests (requires just docker-infra)
just test-minio         # Start MinIO container for integration tests
just test-minio-down    # Stop MinIO
just template-test      # Run the template test
just test-development   # Run tests for the development package

# Single package or test
go test -v ./internal/db
go test -v ./internal/db -run TestSpecificName
```

Tests use SQLite databases in `testdb/` that are created and cleaned per run.

## Linting

```bash
just lint               # Run all linters (go, dockerfile, yaml)
just lint-go            # golangci-lint via Docker
just lint-dockerfile    # Lint Dockerfile with hadolint via Docker
just lint-yaml          # Lint YAML files via Docker
```

## SQL Code Generation

```bash
just sqlc               # Generate sqlc.yml then run sqlc generate (runs in sql/ directory)
just sqlc-config        # Generate sqlc.yml from shared definitions only
just sqlc-config-verify # Verify sqlc.yml is up-to-date (for CI)
just dbgen              # Generate db wrapper code from entity definitions
just dbgen-entity Users # Generate a single entity by name
just dbgen-verify       # Verify generated db wrappers are up-to-date (for CI)
```

`just sqlc` runs `sqlc generate` against `sql/sqlc.yml`, which produces type-safe Go code in three packages:
- `internal/db-sqlite/` (package `mdb`)
- `internal/db-mysql/` (package `mdbm`)
- `internal/db-psql/` (package `mdbp`)

**Never edit files in these three directories by hand** -- they are overwritten by sqlc.

## Docker

Unified via `just dc <backend> <action>`. Backends: `full`, `sqlite`, `mysql`, `postgres`, `prod`.

```bash
# Actions: up, down, reset, dev, fresh, logs (+ destroy for full, minio-reset for postgres)
just dc full up         # Full stack (CMS + all databases + MinIO)
just dc full down       # Stop containers, keep volumes
just dc full reset      # Stop containers and delete volumes
just dc full dev        # Rebuild and restart CMS container only
just dc full fresh      # Reset volumes then rebuild everything
just dc full logs       # Tail CMS logs
just dc full destroy    # Remove containers, volumes, and images (full only)

just dc sqlite up       # SQLite stack
just dc mysql up        # MySQL stack
just dc postgres up     # PostgreSQL stack
just dc postgres minio-reset # Reset MinIO container (postgres only)
just dc prod up         # Production stack

just docker-infra       # Start infrastructure only (postgres, mysql, minio)
just docker-build       # Build standalone CMS image (for CI)
just docker-release     # Tag and push image with latest + version tags
```

## Deploy

```bash
just deploy             # Deploy to production (pull, build, health check, rollback on failure)
just status             # Show production container status
just logs               # Tail production CMS logs
just rollback           # Rollback CMS to previous image
```

## MCP Server

```bash
just mcp-build          # Build MCP server binary (mcp/ directory)
just mcp-install        # Build and install to /usr/local/bin/modula-mcp
```

## Plugins

```bash
just plugin list              # List installed plugins
just plugin init <name>       # Scaffold a new plugin
just plugin validate <path>   # Validate a plugin manifest
just plugin info <name>       # Show plugin details
just plugin reload <name>     # Reload a plugin
just plugin enable <name>     # Enable a plugin
just plugin disable <name>    # Disable a plugin
```

## Dealer

```bash
just dealer up          # Build and start dealer container
just dealer down        # Stop dealer container
just dealer reset       # Stop and delete dealer volumes
just dealer destroy     # Stop, delete volumes, and remove images
just dealer rebuild     # Force rebuild dealer container
just dealer logs        # Tail dealer logs
```

## Misc

```bash
just dump               # Dump sqlite db to SQL file
```

## Architecture

### Runtime Servers (cmd/serve.go)

The `serve` command starts three concurrent servers:
1. **HTTP** on configurable port (default `:8080`)
2. **HTTPS** on configurable port (default `:8443`) with autocert
3. **SSH** on configurable port (default `2222`) via Charmbracelet Wish, piping into the Bubbletea TUI

Graceful shutdown: first SIGINT/SIGTERM triggers shutdown; second forces exit.

### Request Flow

```
Client -> Middleware Chain (CORS, Sessions, Auth, Rate Limit, Permissions, Audit) -> stdlib ServeMux (Go 1.22+) -> Permission Guards -> Handlers -> DbDriver interface -> Database-specific driver -> SQL
```

### Tri-Database Pattern

ModulaCMS supports SQLite, MySQL, and PostgreSQL interchangeably via `config.json`'s `db_driver` field. The architecture:

1. **sqlc generates** per-database Go code from SQL queries in `sql/schema/` into `internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/`
2. **`internal/db/db.go`** defines the `DbDriver` interface (150+ methods) and three wrapper structs (`Database`, `MysqlDatabase`, `PsqlDatabase`) that each implement it
3. **Wrapper methods** in `internal/db/*.go` (e.g., `datatype.go`, `content_data.go`, `media.go`) convert between sqlc-generated types and application-level types, handling NULL conversions and type width differences (SQLite uses int64, MySQL/PostgreSQL use int32)
4. **`db.DefaultDriver`** is set at startup based on config and injected into handlers

### ULID-Based Typed IDs (internal/db/types/)

All entity IDs are 26-character ULIDs wrapped in distinct Go types (`DatatypeID`, `ContentID`, `UserID`, `FieldID`, `MediaID`, etc.) that implement `driver.Valuer`, `sql.Scanner`, and `json.Marshaler`. This provides compile-time type safety -- you cannot pass a `UserID` where a `ContentID` is expected.

Generate new IDs: `types.NewContentID()`, validate: `id.Validate()`, extract timestamp: `id.Time()`.

### Content Tree Structure

Content uses sibling pointers for O(1) navigation and reordering:
- `parent_id` -- parent node
- `first_child_id` -- leftmost child
- `next_sibling_id` / `prev_sibling_id` -- doubly-linked sibling list

### Audited Commands (internal/db/audited/)

Database mutations use an audited command pattern that atomically records `change_events` rows with operation type, old/new JSON values, request metadata, and timestamps for audit trail and replication.

### TUI (internal/tui/)

Built with Charmbracelet Bubbletea (Elm Architecture: Model-Update-View). Key files:
- `model.go` -- central state machine
- `commands.go` -- async action handlers returning `tea.Cmd`
- `page_builders.go` -- screen rendering
- `dialog.go` / `form_dialog.go` -- modal dialog system
- `table_model.go` -- table rendering

### RBAC Authorization (internal/middleware/authorization.go)

Role-based access control with `resource:operation` granular permissions. Three bootstrap roles: **admin** (all 47 permissions, bypasses checks), **editor** (28 permissions), **viewer** (3 read-only permissions).

Key components:
- `PermissionCache` -- in-memory role-to-permissions map, loaded at startup, refreshed every 60s via `StartPeriodicRefresh`. Uses build-then-swap for lock-free reads.
- `PermissionInjector` -- middleware that resolves the authenticated user's role to a `PermissionSet` in context.
- `RequirePermission("resource:operation")` -- single permission check middleware.
- `RequireResourcePermission("resource")` -- auto-maps HTTP method to operation (GET→read, POST→create, PUT/PATCH→update, DELETE→delete).
- `RequireAnyPermission(...)` / `RequireAllPermissions(...)` -- OR/AND logic variants.
- Admin bypass via `ContextIsAdmin()` boolean, not wildcard in PermissionSet.
- Fail-closed: missing PermissionSet in context → 403.
- `ValidatePermissionLabel(label)` -- validates `resource:operation` format (no regex, character-by-character).

The `role_permissions` junction table (`sql/schema/26_role_permissions/`) maps roles to permissions. System-protected roles and permissions cannot be deleted or renamed. Bootstrap data is seeded by `CreateBootstrapData`.

### Admin Panel (internal/admin/)

Server-rendered HTMX + templ admin UI. No React/SPA — all pages are rendered server-side with HTMX for interactivity.

```
internal/admin/           # CSRFMiddleware, AdminAuthMiddleware, StaticFS, CacheControl, embed
internal/admin/handlers/  # Render, RenderWithOOB, NewAdminData, CSRFTokenFromContext, handlers
internal/admin/layouts/   # templ layouts (base, admin, auth) + AdminData type
internal/admin/pages/     # templ full-page components
internal/admin/partials/  # templ HTMX swap targets
internal/admin/components/# templ shared UI components (sidebar, topbar, icon)
internal/admin/static/    # CSS, JS, web components (go:embed)
```

Key patterns:
- **templ** compiles `.templ` files to type-safe Go code. Run `just admin generate` to regenerate. Generated `*_templ.go` files are committed (like sqlc).
- **HTMX** drives all interactions. `HX-Request` header distinguishes partial vs full page renders. `HX-Trigger` for toast notifications.
- **CSRF** uses double-submit cookie pattern. Cookie `csrf_token` set on GET, validated on POST via `X-CSRF-Token` header or `_csrf` form field.
- **Light DOM Web Components** (`mcms-*`) for dialog, data-table, field-renderer, media-picker, tree-nav, toast, confirm, search.
- **Import cycle resolution**: `handlers` owns Render/CSRFTokenFromContext, `admin` owns CSRFContextKey. `PaginationPageData` lives in `partials` to avoid cycle between handlers and pages.
- **Route registration** in `mux.go` via `registerAdminRoutes()` with `mutating()` and `viewing()` middleware helpers.

```bash
just admin generate      # Regenerate templ Go code
just admin watch         # Watch .templ files for changes
just admin verify        # Verify generated files are up-to-date (CI)
just admin bundle        # Bundle block editor JS via esbuild
just admin bundle-watch  # Watch and rebundle block editor JS
just admin bundle-verify # Verify bundle is up-to-date (CI)
```

### HTTP Router (internal/router/)

Uses stdlib `http.ServeMux` (Go 1.22+ pattern routing). Endpoints are registered in `internal/router/mux.go` with per-route permission middleware wrappers. Response formats are configurable per-request (contentful, sanity, strapi, wordpress, clean, raw).

`NewModulacmsMux` takes `(mgr, bridge, driver, pc)` where `pc` is `*middleware.PermissionCache`. Every admin endpoint is wrapped with either `RequireResourcePermission` or `RequirePermission`. Public routes (auth, OAuth, slug delivery) have no permission guards.

### Configuration (internal/config/)

Loaded from `config.json` at project root. Key fields: `db_driver`, `port`, `ssl_port`, `ssh_port`, `bucket_*` (S3), `oauth_*`, `cors_*`, `plugin_*`, `observability_*`. If no config exists at startup, auto-setup runs with defaults.

## SQL Schema Organization

Schemas live in `sql/schema/` as numbered directories (0-26+). Each directory contains six files:

```
sql/schema/{N}_{name}/
  schema.sql / schema_mysql.sql / schema_psql.sql      # DDL
  queries.sql / queries_mysql.sql / queries_psql.sql    # sqlc-annotated queries
```

SQL dialect differences:
- **Placeholders:** SQLite/MySQL use `?`, PostgreSQL uses `$1, $2, $3`
- **Auto-increment:** SQLite `INTEGER PRIMARY KEY`, MySQL `AUTO_INCREMENT`, PostgreSQL `SERIAL`
- **RETURNING:** Supported by SQLite and PostgreSQL, not MySQL (use `:exec` + `LastInsertId`)

After modifying schema or queries: run `just sqlc`, then update the `DbDriver` interface in `internal/db/db.go`, implement methods on all three wrapper structs, and add the table to each struct's `CreateAllTables()`.

Combined schemas (`sql/all_schema*.sql`) are used for fresh installs; regenerate with helper scripts in `sql/schema/`.

## Package Map

| Package | Purpose |
|---------|---------|
| `cmd/` | Cobra CLI commands (serve, install, tui, connect, deploy, pipeline, cert, db, config, backup, update, version) |
| `internal/db/` | DbDriver interface, wrapper structs, application-level types, query builder |
| `internal/db/types/` | ULID-based typed IDs, enums, timestamps, nullable wrappers, field configs |
| `internal/db/audited/` | Audited command pattern for change event recording |
| `internal/db-sqlite/`, `db-mysql/`, `db-psql/` | sqlc-generated code (do not edit) |
| `internal/admin/` | HTMX admin panel: CSRF, auth middleware, static file embed |
| `internal/admin/handlers/` | Admin page handlers (render, auth, CRUD for all resources) |
| `internal/admin/layouts/` | templ layouts (base, admin, auth) and AdminData type |
| `internal/admin/pages/` | templ full-page components (~23 pages) |
| `internal/admin/partials/` | templ HTMX swap targets (~19 partials) |
| `internal/admin/components/` | templ shared UI: sidebar, topbar, icon, status_badge |
| `internal/admin/static/` | CSS, JS, HTMX, web components (go:embed) |
| `internal/tui/` | Bubbletea TUI (40+ files, Elm Architecture) |
| `internal/router/` | HTTP route registration with stdlib ServeMux |
| `internal/middleware/` | CORS, rate limiting, sessions, audit logging, RBAC authorization |
| `internal/auth/` | Authentication (password + OAuth with Google/GitHub/Azure) |
| `internal/config/` | Config struct, file provider, defaults |
| `internal/media/` | Image optimization, preset dimensions, S3 upload |
| `internal/backup/` | Backup/restore (SQL dump + media, stored locally or in S3) |
| `internal/model/` | Domain structs (Root, Node, Datatype, Field) |
| `internal/install/` | Setup wizard and bootstrap checks |
| `internal/plugin/` | Lua plugin system via gopher-lua |
| `internal/registry/` | Project registry (`~/.modula/configs.json`) for `connect` command |
| `internal/remote/` | RemoteDriver -- DbDriver over Go SDK (HTTPS) with retry and connection tracking |
| `internal/deploy/` | Content sync: export, import, push, pull, snapshots between environments |
| `internal/utility/` | Logging (slog), version info, helpers |
| `sdks/typescript/types/` | `@modulacms/types` -- shared TypeScript entity types, branded IDs, enums |
| `sdks/typescript/modulacms-sdk/` | `@modulacms/sdk` -- read-only content delivery TypeScript SDK |
| `sdks/typescript/modulacms-admin-sdk/` | `@modulacms/admin-sdk` -- full admin CRUD TypeScript SDK |
| `sdks/go/` | `modulacms` -- Go SDK for content delivery and admin CRUD |
| `sdks/swift/` | `ModulaCMS` -- Swift SDK for Apple platforms (iOS/macOS/tvOS/watchOS) |

## Workflow Guides

Detailed workflow documentation lives in `ai/workflows/`:
- `ADDING_TABLES.md` -- step-by-step for adding a new database table across all three backends
- `ADDING_FEATURES.md` -- end-to-end feature development flow
- `CREATING_TUI_SCREENS.md` -- adding new Bubbletea screens
- `TESTING.md` -- testing strategies and patterns

Additional AI-context documentation in `ai/`:
- `architecture/` -- system design (distributed architecture, multi-database, plugins, TUI)
- `domain/` -- business domain (content model, trees, datatypes/fields, media, routes)
- `api/` -- REST API contract
- `packages/` -- per-package docs (audited commands, model, transform)
- `plugins/` -- Lua plugin system (guide, API reference, configuration, pipeline plan)
- `operations/` -- deployment and local HTTPS setup
- `reference/` -- observability and troubleshooting
- `analysis/` -- gap analysis (middleware topics, production packages, schema flexibility)
- `plans/` -- future work (composed endpoints, concurrent editing, DB query, plugin SDK, etc.)
- `sqlc/` -- comprehensive sqlc reference docs
- `bubble-references/` -- Bubbletea component reference
- `marketing/` -- positioning and messaging
- `archive/` -- completed plans (historical reference)

User-facing documentation lives in `documentation/` at project root.

## Key Conventions

- The `DbDriver` interface in `internal/db/db.go` is the contract for all database operations. Any new query must be added to this interface and implemented on all three wrapper structs.
- Audited mutations take `(context.Context, audited.AuditContext, Params)` and record change events.
- NULL handling uses helper functions (`NullStringToString`, `StringToNullString`, `NullTimeToString`, etc.) in `internal/db/convert.go`.
- Config is passed through via `config.Config` struct, not environment variables (though env vars can be referenced in config.json via `${VAR}` syntax).
- The TUI communicates via Bubbletea messages -- state changes happen in `Update()`, rendering in `View()`, side effects in `tea.Cmd` functions.
- Permission labels follow the `resource:operation` format (e.g., `content:read`, `media:create`). Use `middleware.ValidatePermissionLabel()` before storing new labels. System-protected roles/permissions (bootstrap data) cannot be deleted or renamed via API.
- Route handlers that modify roles or permissions must trigger an async `pc.Load(driver)` to refresh the PermissionCache. Registration always assigns the `viewer` role; non-admins cannot set or change roles.
