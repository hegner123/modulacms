# AGENTS.md

Agent-oriented reference for working in the ModulaCMS codebase. Everything here is derived from actual source code, configuration, and workflows.

---

## Project Overview

ModulaCMS is a headless CMS written in Go that ships as a **single binary** running three concurrent servers:

1. **HTTP** (default `:8080`) — REST API + admin panel
2. **HTTPS** (default `:8443`) — autocert via Let's Encrypt
3. **SSH** (default `2233`) — Bubbletea TUI via Charmbracelet Wish

Content is managed through the SSH TUI, a server-rendered HTMX admin panel, or the REST API and delivered to frontends over HTTP/HTTPS.

**Requirements**: Go 1.25+, CGO enabled (SQLite via `mattn/go-sqlite3`), Linux/macOS only.

---

## Build System

**Uses `just` (justfile), not Make.** All commands below use `just <recipe>`.

### Essential Commands

| Command | Purpose |
|---------|---------|
| `just dev` | Build local dev binary (`modula-x86`) with version ldflags |
| `just run` | Build and run |
| `just run-admin` | Build and run with `air` hot reload (rebuilds on `.go` and `.templ` changes) |
| `just dev-admin` | Build with `-tags dev` for live CSS/JS from disk (no embed, no cache) |
| `just build` | Build production binary to `out/bin/` |
| `just check` | Compile-check `cmd` and `internal` packages without producing artifacts |
| `just clean` | Remove build artifacts |
| `just vendor` | Update vendor directory |

### Testing

| Command | Purpose |
|---------|---------|
| `just test` | Run all tests (creates/cleans `testdb/*.db` and `backups/*.zip`) |
| `just coverage` | Run tests with coverage report |
| `just test-integration` | S3 integration tests (requires MinIO: `just test-minio` first) |
| `just test-integration-db` | Cross-backend DB integration tests (requires `just docker-infra`) |
| `just test-minio` | Start MinIO container for integration tests |
| `just test-minio-down` | Stop MinIO |
| `just template-test` | Run the template test |

Single package or test:
```bash
go test -v ./internal/db
go test -v ./internal/db -run TestSpecificName
```

Tests use SQLite databases in `testdb/` that are created and cleaned per run. The test helper `testIntegrationDB(t)` creates a fresh WAL-mode SQLite DB in `t.TempDir()` with all tables bootstrapped. `testSeededDB(t)` seeds a full FK-satisfying graph (permissions → roles → users → routes → datatypes → fields → content).

### Code Generation

| Command | Purpose |
|---------|---------|
| `just sqlc` | Generate `sqlc.yml` then run `sqlc generate` (runs in `sql/` directory) |
| `just sqlc-config` | Generate `sqlc.yml` from shared definitions only |
| `just sqlc-config-verify` | Verify `sqlc.yml` is up-to-date (CI) |
| `just dbgen` | Generate db wrapper code from entity definitions |
| `just dbgen-entity Users` | Generate a single entity by name |
| `just dbgen-verify` | Verify generated db wrappers are up-to-date (CI) |

### Admin Panel

| Command | Purpose |
|---------|---------|
| `just admin generate` | Regenerate templ Go code |
| `just admin watch` | Watch `.templ` files for changes |
| `just admin verify` | Verify generated files are up-to-date (CI) |
| `just admin bundle` | Bundle block editor JS via esbuild |
| `just admin bundle-watch` | Watch and rebundle block editor JS |
| `just admin bundle-verify` | Verify bundle is up-to-date (CI) |

### Linting

| Command | Purpose |
|---------|---------|
| `just lint` | Run all linters (Go, Dockerfile, YAML) |
| `just lint-go` | `golangci-lint` via Docker |
| `just lint-dockerfile` | hadolint via Docker |
| `just lint-yaml` | yamllint via Docker |

### SDKs

```bash
just sdk ts install     # pnpm install (workspace root)
just sdk ts build       # Build all packages (types first, then SDKs in parallel)
just sdk ts test        # Run all SDK tests (Vitest)
just sdk ts typecheck   # Typecheck all packages
just sdk go test        # Run Go SDK tests
just sdk go vet         # Vet Go SDK
just sdk swift build    # Build Swift SDK
just sdk swift test     # Run Swift SDK tests
```

### Docker

Unified via `just dc <backend> <action>`. Backends: `full`, `sqlite`, `mysql`, `postgres`, `prod`.

```bash
just dc full up         # Full stack (CMS + all databases + MinIO)
just dc full down       # Stop containers, keep volumes
just dc full reset      # Stop containers and delete volumes
just dc full fresh      # Reset volumes then rebuild everything
just dc sqlite up       # SQLite stack
just dc mysql up        # MySQL stack
just dc postgres up     # PostgreSQL stack
just docker-infra       # Start infrastructure only (postgres, mysql, minio)
```

### Deploy

```bash
just deploy             # Deploy to production (pull, build, health check, rollback on failure)
just status             # Show production container status
just logs               # Tail production CMS logs
just rollback           # Rollback CMS to previous image
```

### Plugins

```bash
just plugin list              # List installed plugins
just plugin init <name>       # Scaffold a new plugin
just plugin validate <path>   # Validate a plugin manifest
just plugin enable <name>     # Enable a plugin
just plugin disable <name>    # Disable a plugin
```

---

## Code Organization

```
cmd/                          # Cobra CLI commands (serve, install, init, tui, connect, deploy, pipeline, cert, db, config, backup, update, version, plugin, mcp)
internal/
  db/                         # DbDriver interface, wrapper structs, application types, query builder
    types/                    # ULID-based typed IDs, enums, timestamps, nullable wrappers, field configs
    audited/                  # Audited command pattern for atomic change event recording
    dbmetrics/                # SQL driver-level instrumentation, query metrics recording
    sql/                      # Embedded SQL for migrations
  db-sqlite/                  # sqlc-generated SQLite code (DO NOT EDIT)
  db-mysql/                   # sqlc-generated MySQL code (DO NOT EDIT)
  db-psql/                    # sqlc-generated PostgreSQL code (DO NOT EDIT)
  admin/                      # HTMX admin panel: CSRF, auth middleware, static file embed
    handlers/                 # Admin page handlers (render, auth, CRUD for all resources)
    layouts/                  # templ layouts (base, admin, auth) and AdminData type
    pages/                    # templ full-page components (~45 pages)
    partials/                 # templ HTMX swap targets (~46 partials)
    components/               # templ shared UI: sidebar, topbar, nav, icon, status_badge
    static/                   # CSS, JS, HTMX, web components (go:embed)
  tui/                        # Bubbletea TUI (140+ files, Elm Architecture)
  router/                     # HTTP route registration with stdlib ServeMux (Go 1.22+ patterns)
  middleware/                  # CORS, rate limiting, sessions, panic recovery, HTTP metrics, RBAC authorization
  auth/                       # Authentication (password + OAuth with Google/GitHub/Azure)
  config/                     # Config struct, file provider, defaults
  media/                      # Image optimization, preset dimensions, S3 upload
  backup/                     # Backup/restore (SQL dump + media)
  definitions/                # Schema definitions for installations and code generation
  deploy/                     # Deployment client/server, export/import, snapshot
  install/                    # Setup wizard and bootstrap checks
  mcp/                        # MCP server for AI tool integration
  model/                      # Domain structs (Root, Node, Datatype, Field)
  plugin/                     # Lua plugin system via gopher-lua
  publishing/                 # Snapshot publishing, version history
  webhooks/                   # Webhook dispatcher, events, signing
  query/                      # Content query: filter, sort, paginate, transform
  tree/                       # Content tree operations
  transform/                  # Response format transformers (Contentful, Sanity, Strapi, WordPress, Clean, Raw)
  validation/                 # Input validation rules and type validators
  bucket/                     # S3-compatible storage client
  service/                    # Service layer for business logic orchestration
  utility/                    # Logging (slog), version info, helpers, metrics, observability
  email/                      # Email service (SMTP, SendGrid, SES, Postmark)
  tui/                        # TUI layout framework (header, panel, statusbar, layers)
  update/                     # Self-update checker
sdks/
  typescript/                 # pnpm workspace monorepo: @modulacms/types, @modulacms/tree, @modulacms/sdk, @modulacms/admin-sdk, @modulacms/plugin-sdk, @modulacms/admin-ui
  go/                         # Go SDK (modulacms package)
  swift/                      # Swift SDK (SPM package, Apple platforms)
sql/
  schema/                     # 38 numbered directories (0-37), each with 6 files (3 schema + 3 queries per DB engine)
  sqlc.yml                    # sqlc configuration (generated — do not hand-edit)
  all_schema*.sql             # Combined schemas for fresh installs
tools/
  dbgen/                      # DB wrapper code generator
  sqlcgen/                    # sqlc.yml generator from shared definitions
  transform_cud/              # CUD transform generator
  transform_bootstrap/        # Bootstrap data generator
mcp/                          # MCP server binary
ai/                           # AI-context documentation (architecture, workflows, domain, API, plugins, plans)
deploy/                       # Docker compose files and deployment configs
```

---

## Architecture Patterns

### Tri-Database Pattern

ModulaCMS supports SQLite, MySQL, and PostgreSQL interchangeably via `config.json`'s `db_driver` field:

1. **sqlc generates** per-database Go code from SQL queries in `sql/schema/` into `internal/db-sqlite/` (package `mdb`), `internal/db-mysql/` (package `mdbm`), `internal/db-psql/` (package `mdbp`)
2. **`internal/db/db.go`** defines the `DbDriver` interface (400+ methods across 24 embedded repository interfaces) and three wrapper structs (`Database`, `MysqlDatabase`, `PsqlDatabase`)
3. **Wrapper methods** in `internal/db/*.go` convert between sqlc types and application types, handling NULL conversions and type width differences (SQLite uses int64, MySQL/PostgreSQL use int32)
4. **`db.DefaultDriver`** is set at startup based on config and injected into handlers

**Never edit files in `internal/db-sqlite/`, `internal/db-mysql/`, or `internal/db-psql/`** — they are overwritten by sqlc.

### ULID-Based Typed IDs

All entity IDs are 26-character ULIDs wrapped in distinct Go types in `internal/db/types/types_ids.go`:

```go
type DatatypeID string
type ContentID string
type UserID string
type FieldID string
type MediaID string
// ... 30+ types total
```

Each implements `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`. This provides **compile-time type safety** — you cannot pass a `UserID` where a `ContentID` is expected.

```go
types.NewContentID()     // Generate new ID
id.Validate()            // Validate ULID format
id.Time()                // Extract embedded timestamp
id.IsZero()              // Check if empty
```

### Audited Commands

Database mutations use an audited command pattern (`internal/db/audited/`) that atomically records `change_events` rows:

```go
audited.Create[T](cmd CreateCommand[T]) (T, error)
audited.Update[T](cmd UpdateCommand[T]) (T, error)
audited.Delete(cmd DeleteCommand) error
```

Each command runs inside a transaction with the entity operation + audit record. Hooks (before/after) can abort or extend the operation.

Mutation handlers always take `(context.Context, audited.AuditContext, Params)`:
```go
driver.CreateDatatype(ctx, auditCtx, CreateDatatypeParams{...})
```

### Code Generation (dbgen)

Generated files follow the pattern `*_gen.go` and have the header:
```go
// Code generated by tools/dbgen; DO NOT EDIT.
```

Each generated file contains:
- **Structs**: Entity struct, Create/Update params, any list params
- **SQLite methods** on `Database`
- **MySQL methods** on `MysqlDatabase`
- **PostgreSQL methods** on `PsqlDatabase`
- **MapString* function** for TUI display conversion

Custom (hand-written) extensions go in `*_custom.go` files alongside the generated ones.

### Content Tree Structure

Content uses sibling pointers for O(1) navigation and reordering:
- `parent_id` — parent node
- `first_child_id` — leftmost child
- `next_sibling_id` / `prev_sibling_id` — doubly-linked sibling list

### Request Flow

```
Client → Middleware Chain (Recovery, RequestID, ClientIP, UserAgent, Logging, HTTPMetrics, CORS, Auth, PublicEndpoint, PermissionInjector)
       → stdlib ServeMux (Go 1.22+ pattern routing)
       → Permission Guards
       → Handlers
       → DbDriver interface
       → Database-specific driver
       → SQL
```

### RBAC Authorization

Role-based access control with `resource:operation` granular permissions (`internal/middleware/authorization.go`):

- **Three bootstrap roles**: admin (all 72 permissions, bypasses checks), editor (36), viewer (5 read-only)
- **`PermissionCache`** — in-memory role-to-permissions map, loaded at startup, refreshed every 60s. Build-then-swap for lock-free reads.
- **Permission middleware**: `RequirePermission("resource:operation")`, `RequireResourcePermission("resource")` (auto-maps HTTP method → operation), `RequireAnyPermission(...)`, `RequireAllPermissions(...)`
- **Admin bypass** via `ContextIsAdmin()` boolean, not wildcard
- **Fail-closed**: missing `PermissionSet` in context → 403
- Permission labels follow `resource:operation` format. Validate with `middleware.ValidatePermissionLabel(label)`.

---

## Admin Panel Patterns

Server-rendered HTMX + templ. No React/SPA.

### Handler Pattern

Handlers are closure factories that capture dependencies:

```go
func DatatypesListHandler(driver db.DbDriver) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse pagination, query params
        // Fetch data via driver
        // HTMX partial vs full page render
    }
}
```

### Route Registration

Admin routes are registered in `internal/router/mux.go` via `registerAdminRoutes()`:

```go
mutating := func(permission string, h http.HandlerFunc) http.Handler {
    return adminAuth(csrf(middleware.RequirePermission(permission)(http.HandlerFunc(h))))
}
viewing := func(resource string, h http.HandlerFunc) http.Handler {
    return adminAuth(csrf(middleware.RequirePermission(resource+":read")(http.HandlerFunc(h))))
}

mux.Handle("GET /admin/content", viewing("content", ContentListHandler(driver, mgr)))
mux.Handle("POST /admin/content", mutating("content:create", ContentCreateHandler(driver, mgr)))
```

### HTMX Rendering

Handlers distinguish between HTMX partial requests and full page loads:
- **HTMX nav** (`HX-Request` header present): render partial content + OOB swap targets
- **Full page**: wrap content in admin layout (sidebar, topbar, base HTML)

### CSRF

Double-submit cookie pattern:
- Cookie `csrf_token` set on GET, scoped to `/admin/`
- Validated on POST via `X-CSRF-Token` header or `_csrf` form field
- Token stored in context via `admin.CSRFContextKey{}`

### templ Templates

`.templ` files compile to type-safe Go code. Generated `*_templ.go` files are committed.

```
layouts/base.templ         → HTML shell, CSS/JS includes, cache busting via version
pages/content_list.templ   → Full page components
partials/pagination.templ  → HTMX swap targets (table rows, forms, dialogs)
components/sidebar.templ   → Shared UI components
```

Regenerate with `just admin generate`. Watch mode: `just admin watch`.

### Static Assets

- **Build tags**: `//go:build !dev` embeds `static/*` into binary; `-tags dev` serves from disk
- **Web components**: Light DOM components prefixed `mcms-*` (command-palette, confirm, data-table, dialog, field-renderer, file-input, focal-point, media-grid, media-picker, publish-button, scroll, search, toast, tree-nav, validation-wizard)
- **Block editor**: Source in `static/js/block-editor-src/`, bundled via esbuild to `static/js/block-editor.js`
- **CSS**: tailwind.css → block-editor.css (Tailwind migration complete, legacy CSS files removed)

### Hot Reload (Development)

`just run-admin` uses `air` with this workflow:
1. `air` watches `.go` and `.templ` files
2. Pre-cmd runs `templ generate`
3. Builds with `just dev-admin` (`-tags dev`)
4. Runs `./out/modulacms-dev serve --config out/config.json`

---

## TUI Patterns

Built with Charmbracelet Bubbletea (Elm Architecture: Model → Update → View).

Key files in `internal/tui/`:
- `model.go` — central state machine (`Model` struct with focus, state, navigation)
- `update*.go` — state transition handlers (`Update()` dispatches by focus and message type)
- `view.go` / `panel_view.go` — screen rendering
- `commands.go` — async action handlers returning `tea.Cmd`
- `dialog.go` / `form_dialog.go` — modal dialog system
- `table_model.go` — table rendering
- `constructors.go` — page/form constructors

Focus system: `PAGEFOCUS`, `TABLEFOCUS`, `FORMFOCUS`, `DIALOGFOCUS` — determines which component receives keypresses.

---

## SQL Schema

Schemas live in `sql/schema/` as 38 numbered directories (0-37). Each directory contains up to 6 files:

```
sql/schema/{N}_{name}/
  schema.sql / schema_mysql.sql / schema_psql.sql      # DDL
  queries.sql / queries_mysql.sql / queries_psql.sql    # sqlc-annotated queries
```

SQL dialect differences:
- **Placeholders**: SQLite/MySQL use `?`, PostgreSQL uses `$1, $2, $3`
- **Auto-increment**: SQLite `INTEGER PRIMARY KEY`, MySQL `AUTO_INCREMENT`, PostgreSQL `SERIAL`
- **RETURNING**: Supported by SQLite and PostgreSQL, not MySQL (use `:exec` + `LastInsertId`)

After modifying schema or queries:
1. Run `just sqlc`
2. Update the `DbDriver` interface in `internal/db/db.go`
3. Implement methods on all three wrapper structs
4. Add the table to each struct's `CreateAllTables()`

---

## Adding a New Database Table

Detailed guide: `ai/workflows/ADDING_TABLES.md`. Summary:

1. Pick the next number (`ls sql/schema/` to find max)
2. Create 6 SQL files (schema + queries × 3 engines)
3. Add sqlc overrides if needed
4. `just sqlc` to generate Go code
5. Add typed ID in `internal/db/types/types_ids.go` (if new entity)
6. Add entity definition in `tools/dbgen/definitions.go`
7. `just dbgen` to generate wrapper code
8. Add new methods to `DbDriver` interface in `internal/db/db.go`
9. Update `CreateAllTables()` and `DropAllTables()` on all three wrapper structs
10. Add optional `_custom.go` for hand-written logic
11. Write tests

---

## Adding a Feature End-to-End

Detailed guide: `ai/workflows/ADDING_FEATURES.md`. Summary:

1. **Schema** — add SQL files in `sql/schema/`
2. **sqlc** — run `just sqlc`
3. **DbDriver** — add to interface + implement on all 3 wrappers
4. **Business logic** — add in appropriate `internal/` package
5. **TUI** — add Bubbletea screens if needed
6. **API** — add router handlers in `internal/router/`
7. **Admin** — add templ pages/partials + handlers
8. **Tests** — unit + integration
9. **Docs** — update API docs and AI context

---

## Testing Patterns

### Test Helpers (`internal/db/test_helpers_test.go`)

```go
db := testIntegrationDB(t)     // Fresh SQLite with all tables
db := testSeededDB(t)          // Full FK-satisfying seed data

auditCtx := testAuditCtx       // Pre-built audit context
auditCtx := testAuditCtxWithUser(userID)
```

### Compile-Time Interface Checks

Entity test files verify all 3 drivers implement audited commands:

```go
var _ audited.CreateCommand[mdb.Datatypes] = (*db.DatatypeCreateCmd)(nil)
var _ audited.CreateCommand[mdbm.Datatypes] = (*db.DatatypeCreateMysqlCmd)(nil)
var _ audited.CreateCommand[mdbp.Datatypes] = (*db.DatatypeCreatePsqlCmd)(nil)
```

### Test Fixtures

Each entity has `*TestFixture()` and `*UpdateParams()` helper functions that produce fully-populated test structs.

### Integration Tests

- **DB integration** (`-tags integration`): cross-backend tests requiring `just docker-infra`
- **S3 integration** (`-tags integration`): media tests requiring `just test-minio`
- Tests create isolated DBs via `t.TempDir()` — no shared state

---

## SDKs

### TypeScript (`sdks/typescript/`)

pnpm workspace monorepo with six packages:

| Package | npm Name | Purpose |
|---------|----------|---------|
| `types/` | `@modulacms/types` | Shared entity types, branded IDs, enums |
| `tree/` | `@modulacms/tree` | Content tree utilities |
| `modulacms-sdk/` | `@modulacms/sdk` | Read-only content delivery SDK |
| `modulacms-admin-sdk/` | `@modulacms/admin-sdk` | Full admin CRUD SDK |
| `plugin-sdk/` | `@modulacms/plugin-sdk` | Plugin UI SDK (Web Components, zero deps, browser-only) |
| `admin-ui/` | `@modulacms/admin-ui` | Admin panel TypeScript (block editor state) |

Tooling: TypeScript 5.7+, tsup (ESM+CJS dual builds), Vitest, pnpm 9+, Node 22+.

### Go SDK (`sdks/go/`)

Import path: `import modulacms "github.com/hegner123/modulacms/sdks/go"`

Generic `Resource[E, C, U, ID]` pattern for CRUD operations.

### Swift SDK (`sdks/swift/`)

SPM package. Platforms: iOS 16+, macOS 13+, tvOS 16+, watchOS 9+. Swift 5.9+, zero dependencies.

---

## CI/CD

### Go CI (`.github/workflows/go.yml`)

- Triggers on `main`, `develop`, `dev` push and `main` PRs (ignores `sdks/**`)
- **test** → **build** (matrix: darwin amd64/arm64, linux amd64/arm64) → **release** (on tags) → **deploy-dev** (on develop/dev branches)
- CGO_ENABLED=1, requires `libwebp-dev` on Linux
- Build uses vendored dependencies (`-mod vendor`)

### SDK CI (`.github/workflows/sdks.yml`)

- Triggers on `sdks/**` changes
- **TypeScript**: pnpm 9 + Node 22 → install → build types → typecheck → build → test
- **Go SDK**: Go 1.25 → vet → test
- **Swift SDK**: macOS 14 + Xcode 15.4 → build → test

---

## Configuration

Loaded from `config.json` at project root. Key fields:

- `db_driver` — `sqlite`, `mysql`, `psql`
- `port`, `ssl_port`, `ssh_port` — server ports
- `bucket_*` — S3 settings
- `oauth_*` — OAuth provider settings (Google, GitHub, Azure)
- `cors_*` — CORS configuration
- `plugin_*` — Plugin system settings
- `observability_*` — Metrics and tracing

Environment variables can be referenced via `${VAR}` syntax in config.json.

---

## Key Conventions

- **Build runner**: `just`, not `make`
- **Module path**: `github.com/hegner123/modulacms`
- **Vendor mode**: dependencies are vendored; use `just vendor` after `go get`
- **Generated code**: `*_gen.go` (dbgen), `*_templ.go` (templ), `internal/db-{sqlite,mysql,psql}/` (sqlc) — never hand-edit
- **Custom extensions**: `*_custom.go` files alongside generated ones
- **ID types**: always use typed IDs (`types.ContentID`, `types.UserID`), never raw strings
- **Audit context**: all mutations take `audited.AuditContext` for change event recording
- **NULL handling**: use helpers in `internal/db/convert.go` (`NullStringToEmpty`, `StringToNullString`, etc.)
- **Error logging**: `utility.DefaultLogger` (slog-based)
- **Permission labels**: `resource:operation` format (e.g., `content:read`, `media:create`)
- **Router**: stdlib `http.ServeMux` with Go 1.22+ method-pattern routing (`"GET /api/v1/..."`)
- **Admin handler pattern**: closure factories capturing `db.DbDriver` and returning `http.HandlerFunc`
- **HTMX-aware responses**: check `HX-Request` header to decide partial vs full render
- **Config access**: via `*config.Manager` (supports hot reload), not global state
- **Template regeneration**: run `just admin generate` after changing any `.templ` file
- **Permissions on routes**: `viewing("resource", h)` for reads, `mutating("resource:action", h)` for writes

---

## Gotchas and Non-Obvious Patterns

1. **Three implementations for everything**: any new DB query needs implementation on `Database` (SQLite), `MysqlDatabase`, and `PsqlDatabase` wrapper structs. Missing one means compile failure.

2. **sqlc.yml is generated**: don't edit `sql/sqlc.yml` directly. Edit the definitions in `tools/sqlcgen/` and run `just sqlc-config`.

3. **Type width differences**: SQLite uses `int64`, MySQL/PostgreSQL use `int32`. Wrapper methods handle the conversion.

4. **Build tags for admin assets**: production (`//go:build !dev`) embeds static files. Development (`-tags dev`) serves from disk for hot reload. Both `admin.go` and `embed.go` have build-tagged variants.

5. **Import cycles in admin**: `handlers` owns `Render`/`CSRFTokenFromContext`, `admin` owns `CSRFContextKey`. `PaginationPageData` lives in `partials` to avoid cycle between handlers and pages.

6. **Permission cache refresh**: route handlers that modify roles or permissions must trigger `pc.Load(driver)` to refresh the in-memory `PermissionCache`.

7. **CSRF token flow**: the cookie is `HttpOnly=false` (JS-readable) with `SameSite=Strict`. GET requests reuse existing tokens to prevent HTMX partial navigation desync.

8. **testdb/ and backups/ directories**: tests expect these to exist. The `just test` recipe creates and cleans them. If running `go test` directly, ensure they exist.

9. **templ files are committed**: both `.templ` source and generated `*_templ.go` are committed. CI verifies they're in sync via `just admin verify`.

10. **Block editor bundle is committed**: `internal/admin/static/js/block-editor.js` is bundled output. Source is in `block-editor-src/`. CI verifies via `just admin bundle-verify`.

11. **Admin panel HTMX patterns**: full page loads render the complete layout; HTMX navigation requests (`HX-Request` header) return only the content partial + OOB swap targets. Both paths go through the same handler.

12. **Cobra CLI structure**: `cmd/main.go` is the entrypoint. `cmd/root.go` defines the root command. Subcommands: `serve`, `install`, `init`, `tui`, `connect`, `deploy`, `pipeline`, `cert`, `db`, `config`, `backup`, `update`, `version`, `plugin`, `mcp`.

---

## Documentation

Extensive AI-context documentation lives in `ai/`:

| Directory | Content |
|-----------|---------|
| `ai/workflows/` | Step-by-step guides: adding tables, features, TUI screens, testing |
| `ai/architecture/` | System design: distributed arch, multi-database, plugins, TUI |
| `ai/domain/` | Business domain: content model, trees, datatypes/fields, media, routes |
| `ai/api/` | REST API contract |
| `ai/packages/` | Per-package docs (audited commands, model, transform) |
| `ai/plugins/` | Lua plugin system guide, API reference, configuration |
| `ai/plans/` | Future work: composed endpoints, concurrent editing, plugin SDK |
| `ai/sqlc/` | Comprehensive sqlc reference |

User-facing documentation is in `documentation/` at project root.
