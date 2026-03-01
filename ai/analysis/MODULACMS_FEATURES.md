# ModulaCMS Features

## Runtime Architecture
- Single binary with three concurrent servers: HTTP, HTTPS (Let's Encrypt autocert), SSH (Bubbletea TUI)
- Graceful shutdown (first signal → shutdown, second → force exit)
- Configurable ports for all three servers
- Self-signed certificate generation for local dev
- Interactive setup wizard (`--wizard` flag)
- Auto-setup with defaults if no config exists
- Build version info via ldflags (Version, GitCommit, BuildDate)

## Tri-Database Pattern
- SQLite (CGO, mattn/go-sqlite3)
- MySQL
- PostgreSQL
- Selectable per deployment via `config.json`
- Unified `DbDriver` interface (150+ methods) implemented by all three
- sqlc-generated type-safe code per backend

## Content Management
- CRUD for content data entries
- Content tree with sibling pointers for O(1) navigation/reordering (`parent_id`, `first_child_id`, `next_sibling_id`, `prev_sibling_id`)
- Content reordering within a parent
- Content move (cross-parent relocation)
- Content tree bulk save (bulk pointer updates + deletes)
- Content tree heal (repair malformed IDs)
- Pagination and status filtering (draft/published)
- Batch content updates (multiple items in one request)
- CRUD for content field values
- Separate admin content tree with its own content data + fields

## Content Schema
- CRUD for datatypes (content type definitions)
- Datatype full read (datatype + all fields in one response)
- Field assignment to datatypes
- CRUD for fields
- **Data-driven field type registry** (`field_types` table with full CRUD via API and admin UI)
- Open `FieldType` string type -- not a closed enum; accepts any value at the type system level
- 15 predefined (bootstrap) types: `text`, `textarea`, `number`, `date`, `datetime`, `boolean`, `select`, `media`, `relation`, `json`, `richtext`, `slug`, `email`, `url`, `content_tree_ref`
- Unlimited custom field types registrable via API, admin panel, or plugins
- Separate `admin_field_types` table for admin content tree field types (also CRUD)
- Configurable richtext toolbar buttons (bold, italic, h1–h3, link, ul, ol, preview, etc.)
- Separate admin datatypes, admin fields, admin field types
- Sort order on fields and admin fields

## Content Relations
- Cross-content references via content relations table
- Admin content relations table

## Content Delivery (Public API)
- Slug-based content delivery (`/api/v1/content/<slug>`)
- Serves from published snapshots (immutable, fast)
- Preview mode for authenticated users (`?preview=true`, sets `X-Robots-Tag: noindex`)
- Configurable output format per-request: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`

## Query API (`/api/v1/query/{datatype}`)
- In-memory query pipeline: resolve datatype → fetch content → filter by status → batch fields → build indexes → filter → sort → paginate
- **Zero SQL injection surface** -- all filtering, sorting, and pagination processed in Go, not dynamic SQL
- 7 filter operators with type-aware comparison (number, boolean, date, string)
- Multi-field sorting with type-aware comparison, missing values sort to end
- Configurable pagination (default 20, max 100) with offset support
- QueryTransformer interface for pluggable output formats (RawQueryTransformer produces JSON envelope)
- Tri-database compatible -- single Go implementation works identically across SQLite/MySQL/PostgreSQL
- Batch field loading via `ListContentFieldsByContentDataIDs` with `QSelect` + `In()` condition
- SDK support: Go (`QueryResource`), TypeScript (`queryContent()`), Swift (`QueryResource`)

## Content Publishing
- Publish content (creates snapshot, sets status to `published`)
- Unpublish content (resets to `draft`)
- Schedule publish (set `publish_at` for future publishing)
- Background scheduler with configurable tick interval (default 60s)
- Catches up on overdue scheduled items at startup
- Publish/unpublish/schedule for admin content tree separately

## Content Version History
- Automatic version snapshot on publish
- Manual version creation with optional label
- List, get, delete versions per content item
- Restore content from a specific version
- Version compare (side-by-side diff in admin panel)
- Configurable max versions per content item (default 50, 0 = unlimited)
- Hourly background prune sweep for excess versions

## Content Tree Composition
- Composed/nested tree read (full hierarchical tree for a route)
- Configurable composition max depth (default 10)
- Tree composition references via `content_tree_ref` field type

## Media Management
- Upload (multipart form), list (paginated), get, update metadata, delete
- Media health check (verifies storage backend)
- Media cleanup (orphan removal)
- Automatic image optimization on upload
- Multi-resolution output from single upload
- Configurable dimension presets (`media_dimensions` table)
- Focal point-aware crop
- Supported formats: JPEG, PNG, WebP, GIF → output to WebP
- Dimension validation (max 10,000x10,000 px, 50 megapixels)
- Configurable max upload size (default 10 MB)

## Storage Backends
- Local filesystem storage
- S3-compatible object storage (AWS S3, MinIO, DigitalOcean Spaces, etc.)
- Configurable region, bucket, endpoint, credentials, ACL, path-style, public URL
- Separate buckets for media and backups

## Routes
- CRUD for routes (URL → content tree mappings)
- Route types: `static`, `dynamic`, `api`, `redirect`
- Admin routes (separate admin-facing route registry)

## Authentication
- Password-based with bcrypt (cost 12)
- Session cookie authentication (configurable name, duration, Secure, SameSite)
- Session management (list, delete sessions)
- API token authentication (create, delete tokens)
- OAuth 2.0: Google, GitHub, Azure (configurable client ID/secret, scopes, URLs)
- SSH public key authentication for TUI access (per-user key management)
- Registration endpoint
- Password reset via email (request → confirmation link → new password)

## Authorization (RBAC)
- Dual-layer permission model:
  - **Layer 1: Resource-level** -- `resource:operation` permissions (e.g., `content:read`, `media:create`)
  - **Layer 2: Field-level** -- `roles` column on `fields` and `admin_fields` tables; nullable JSON array of role IDs. NULL = all roles can edit, populated array = restricted to listed roles
- Bootstrap roles: **admin** (47 permissions, bypasses checks), **editor** (28), **viewer** (3 read-only)
- `PermissionCache`: in-memory, lock-free reads, 60-second periodic refresh
- Middleware: `RequirePermission`, `RequireAnyPermission`, `RequireAllPermissions`, `RequireResourcePermission` (auto-maps HTTP method)
- Admin bypass via `ContextIsAdmin()` boolean
- Fail-closed (missing PermissionSet → 403)
- System-protected roles/permissions (bootstrap data cannot be deleted/renamed)
- Permission label validation (character-by-character, no regex)
- Registration assigns `viewer`; non-admins cannot set roles

## Middleware Stack
- Request ID generation
- Structured HTTP request/response logging
- CORS (configurable origins, methods, headers, credentials)
- Session authentication
- Public endpoint protection
- Permission set injection
- Rate limiting (token bucket, per-IP, configurable rate/burst — auth: 10 req/min)
- Audit context injection (user ID, client IP, request ID, plugin hook runner)
- CSRF protection (double-submit cookie pattern)
- Admin authentication middleware

## Block Editor (`<block-editor>` Web Component)
- Composable content blocks backed by the datatype/field schema system
- Same sibling-pointer tree architecture as the content tree (`parentId`, `firstChildId`, `prevSiblingId`, `nextSiblingId`)
- Blocks created from any registered datatype -- each block carries `datatypeId`, per-block `fields[]`, author, route, status, timestamps
- Nestable blocks: container blocks accept children, configurable max depth (default 8)
- Drag-and-drop reordering with auto-scroll, drop indicators, and descendant-drop prevention
- Keyboard shortcuts: Tab/Shift+Tab (indent/outdent), Arrow Up/Down (selection traversal), Enter (new block), Delete/Backspace (remove), Ctrl+Shift+D (duplicate)
- Hover toolbar per block: Move Up, Move Down, Indent, Outdent, Duplicate, Delete
- Open type system: 4 built-in rendering types (text, heading, image, container) + automatic fallback for any database-defined type (assumes `canHaveChildren: true`)
- Datatype cache with 5-minute TTL and in-flight deduplication
- State validation (pointer consistency checks) with dev-mode warnings
- Dirty tracking with beforeunload guard
- Serializable state (`blocks` + `rootId` as JSON) emitted via `block-editor:save` custom event
- Selection events via `block-editor:select` and mutation events via `block-editor:change`
- Full subtree duplication with ID regeneration
- Bundled via esbuild from modular source (`block-editor-src/`: index, config, state, tree-ops, tree-queries, validate, cache, dom-patches, drag, styles, id)

## Admin Panel (HTMX + templ)
- Server-rendered (no SPA/React), HTMX for interactivity
- templ type-safe Go HTML templates
- Light DOM Web Components: `mcms-dialog`, `mcms-data-table`, `mcms-field-renderer`, `mcms-media-picker`, `mcms-tree-nav`, `mcms-toast`, `mcms-confirm`, `mcms-search`, `block-editor`
- CSRF on all mutating requests
- Static assets embedded with `go:embed`
- Cache-Control on static assets
- Pages: Dashboard, Content list/edit, Version history/compare, Datatypes list/detail, Fields detail, Field types list, Media list/detail, Routes list, Admin routes list, Users list/detail, Roles list/detail/new, Tokens list, Sessions list, Plugins list/detail, Import, Audit log, Settings, Demo, Forgot password, Reset password
- Admin JSON API endpoints for block editor and richtext toolbar config

## Plugin System (Lua)
- Lua plugins via gopher-lua
- Per-plugin VM pool (channel-based, configurable size, default 4 VMs)
- Sandboxed VM (safe stdlib subset only — no io, os, package, debug)
- Frozen read-only module proxies
- Plugin tables prefixed with `plugin_<name>_`, no cross-plugin access
- Hot reload (configurable, off by default)
- Circuit breaker per plugin (configurable failure threshold, reset interval)
- Background file watcher for hot reload
- Lifecycle: `on_init()`, `on_shutdown()`, `plugin_info` manifest
- **db module**: `define_table`, `insert`, `query`, `update`, `delete`, `ulid()` — column types, indexes, foreign keys, unique constraints
- **http module**: `handle(method, path, handler_fn)` — register HTTP routes
- **hooks module**: `on(event, table, handler_fn)` — content lifecycle hooks
- **core module**: gated access to core CMS tables (three-layer gating)
- **log module**: `info`, `warn`, `error`
- Before-hooks (synchronous, can block/modify) and after-hooks (async, non-blocking)
- Events: `before_insert`, `after_insert`, `before_update`, `after_update`, `before_delete`, `after_delete`
- Table wildcard `"*"` for all-table hooks
- Priority ordering, per-hook timeout, per-event chain timeout
- Schema drift detection
- VM health reporting
- Multi-instance sync via DB state polling

## Plugin CLI (`modula plugin`)
- `list`, `init`, `validate`, `install`, `info`, `reload`, `enable`, `disable`
- `approve` / `revoke` for routes and hooks (individual or `--all-routes`/`--all-hooks`)
- `--yes` flag for non-interactive CI/CD
- `--token` flag for API token override

## Pipeline CLI (`modula pipeline`)
- `list`, `show <table>`, `enable <id>`, `disable <id>`, `remove <id>`

## CMS Import
- Import from: Contentful, Sanity, Strapi, WordPress, Clean (ModulaCMS native), Bulk
- Bidirectional transform support
- Output formats configurable when reading back

## Deploy Sync
- Export content tables to JSON (selectable tables, gzip support)
- Import exported JSON (optional pre-import backup, dry-run mode)
- Snapshot system: versioned import snapshots stored on disk (list, show details, restore)
- Pull from remote environment (--dry-run, --skip-backup, --tables)
- Push to remote environment (--dry-run, --tables)
- Named environments in config with URL + API key
- Environment management: `env list`, `env test`
- `--json` output for machine-readable output
- User reference resolution across deployments
- Conflict policies: `lww` (Last Write Wins), `manual`

## Backup & Restore
- Full backup: SQL dump + configured paths → zip archive
- Backup types: full, incremental, differential (enum)
- Backup status tracking: pending, in_progress, completed, failed
- Backup history in DB (ID, type, status, timestamps, duration, size, path, triggered_by, metadata)
- S3 upload of backup archives
- Verification status tracking
- CLI: `backup create`, `backup restore`, `backup list`

## SSH TUI (Terminal UI)
- Bubbletea Elm Architecture served over SSH via Wish
- SSH public key + session-based authentication
- Content management (browse trees, create/edit/delete nodes, edit field values)
- Schema management (datatypes, fields)
- Route management
- User management
- Media management
- Admin panel (admin content trees, admin fields)
- Deploy panel (multi-environment push/pull with status)
- Update panel (check for and apply CMS updates)
- Database form dialog (interactive DB config)
- UI config form dialog
- Filter and navigation history
- Field bubble components: text, textarea, number, boolean, select, slug, email, url
- Quickstart flow for new installs
- Table model for tabular display
- Custom theme/style system
- Configurable key bindings
- ASCII art title banners
- Status bar with mode/context

## CLI Commands
- `modula serve` (start servers, `--wizard`, `--config`)
- `modula install` (interactive setup wizard)
- `modula db init` / `db wipe` / `db wipe-redeploy` / `db reset` / `db export`
- `modula cert generate` (self-signed SSL, optionally trust)
- `modula backup create` / `backup restore` / `backup list`
- `modula plugin` (full plugin management)
- `modula pipeline` (pipeline management)
- `modula deploy` (full deploy sync)
- `modula update` (check/apply binary updates from GitHub releases, platform-aware)
- `modula version` (version, commit, build date)
- `modula tui` (launch TUI directly without SSH)
- `modula config` (config management)

## Configuration System
- Hot-reloadable config via `config.Manager` + `FileProvider`
- `${VAR}` and `${VAR:-default}` environment variable expansion
- All server ports, database credentials, S3 storage, auth, cookies, OAuth, CORS, plugins, observability, email, deploy, composition, publishing, versioning, richtext toolbar, keybindings, backup, style customization, logging

## Email System
- Providers: Disabled, SMTP (TLS), SendGrid, AWS SES, Postmark
- Transactional email (password reset)
- Configurable from address, from name, reply-to
- HTML email with MIME handling

## Observability
- Configurable provider: Sentry, Datadog, New Relic, etc.
- DSN, environment, release, sample rate, traces sample rate, PII control, debug mode, server name, flush interval, global tags
- Structured logging via `slog`
- Request IDs on all requests (`X-Request-ID`)
- Authorization events logged with full context

## Auditing
- `change_events` table records all DB mutations atomically
- Captures: operation type, old/new JSON values, request metadata, timestamps, user ID, node ID, action type, source
- HLC (Hybrid Logical Clock) timestamps for distributed ordering
- Audit log viewable in admin panel with pagination

## ULID-Based Typed IDs
- 30+ branded ID types (ContentID, UserID, FieldID, MediaID, DatatypeID, RouteID, RoleID, PermissionID, etc.)
- Each implements `driver.Valuer`, `sql.Scanner`, `json.Marshaler`
- Compile-time type safety
- `NewXxxID()`, `.Validate()`, `.Time()` (extract timestamp)
- Nullable ID wrappers for optional foreign keys

## SDKs
- **TypeScript**: `@modulacms/types` (shared types, branded IDs, enums), `@modulacms/sdk` (read-only delivery), `@modulacms/admin-sdk` (full admin CRUD) — pnpm workspace, tsup (ESM+CJS), Vitest
- **Go**: `modulacms` package — generic CRUD, all entity types, branded IDs, enums, errors, pagination, auth, media upload, content delivery, SSH keys, sessions, import, batch
- **Swift**: SPM package — iOS 16+, macOS 13+, tvOS 16+, watchOS 9+ — generic CRUD, all entity types, branded IDs, URLSession transport, zero dependencies

## MCP Server
- Model Context Protocol server exposing ModulaCMS to AI assistants via the Go SDK
- Tools for: content, admin content, schema, media, routes, users, RBAC, sessions, tokens, SSH keys, import, deploy, config, plugins, health, OAuth

## Code Generation
- **sqlc**: SQL → type-safe Go (three backends)
- **dbgen**: custom codegen from entity definitions in `internal/definitions/`
- **templ**: `.templ` → type-safe Go HTML
- **esbuild**: block editor JS bundling

## Testing
- `just test` (creates/cleans testdb/*.db)
- `just coverage` (coverage report)
- `just test-integration` (S3 via MinIO)
- `just test-integration-db` (cross-backend MySQL + PostgreSQL)
- SDK tests (TypeScript, Go, Swift)
- Admin render tests
- Plugin capability drift test

## Deployment & Docker
- Dockerfile for production
- Docker Compose stacks: full, sqlite, mysql, postgres, production
- Infrastructure-only stack (postgres + mysql + minio)
- Production deploy via `just deploy` (SSH remote: pull, build, health-check, rollback on failure)
- `just status`, `just logs`, `just rollback`
- CI: `.github/workflows/go.yml` (Go), `.github/workflows/sdks.yml` (all SDKs)
- Linting: golangci-lint, hadolint, yamllint

## Database Schema (33 schema directories)
- `backups`, `change_events`, `wipe`, `permissions`, `roles`, `media_dimension`, `users`, `admin_routes`, `routes`, `datatypes`, `fields`, `admin_datatypes`, `admin_fields`, `tokens`, `user_oauth`, `tables`, `media`, `sessions`, `content_data`, `content_fields`, `admin_content_data`, `admin_content_fields`, `joins`, `user_ssh_keys`, `content_relations`, `admin_content_relations`, `role_permissions`, `field_types`, `admin_field_types`, `plugins`, `pipelines`, `content_versions`, `admin_content_versions`
