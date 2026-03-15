# Glossary

Definitions of key terms and concepts used throughout ModulaCMS documentation.

## Content Model

**Content Data** -- A single content entry in the content tree. Each content data record belongs to a route, has a datatype, and is authored by a user. Content data records form a tree structure through parent-child relationships. Identified by a ULID.

**Content Field** -- A value associated with a content data record. Each content field links a field definition to a content data record and stores the actual data (text, number, date, etc.). A content data record typically has many content fields, one for each field in its datatype.

**Datatype** -- A content type definition, analogous to a "model" or "content type" in other CMS platforms. A datatype defines the structure of content by specifying which fields it contains. Examples: "Blog Post", "Product", "Page".

**Field** -- A field definition that describes a single property of a datatype. Fields have a label (e.g., "Title"), a data type (e.g., "string"), and a field type (e.g., "text", "rich_text", "number"). Each field belongs to a datatype.

**Content Tree** -- The hierarchical structure of content within a route. Content nodes can be nested under other nodes, forming a tree. You can reorder siblings and move nodes between parents through the API, TUI, or admin panel.

## Routing

**Route** -- A named entry point that maps a slug (e.g., "blog", "products") to a content tree for delivery. Each route has a slug, a title, a status, and an author. When a client requests `GET /blog`, the API looks up the route with slug "blog" and returns its published content tree. Global trees exist independently and are served via `/api/v1/globals`.

**Admin Route** -- A route that manages the admin panel's own content structure. Admin routes use a parallel set of tables to store configuration and UI content separately from public content.

**Site** -- In ModulaCMS, the `client_site` and `admin_site` configuration fields define the primary domain and admin domain. Routes serve as the multi-site mechanism: different routes can represent different sections or sites within a single ModulaCMS instance.

## Identity and IDs

**ULID** -- Universally Unique Lexicographically Sortable Identifier. A 26-character string (e.g., `01HXK4N2F8RJZGP6VTQY3MCSW9`) used as the primary key for all entities in ModulaCMS. ULIDs are time-ordered (sortable by creation time), globally unique, and URL-safe.

## Database

**Database Backend** -- ModulaCMS supports three database engines interchangeably: SQLite, MySQL, and PostgreSQL. Select the active backend with the `db_driver` field in `modula.config.json`. Application behavior is identical across all three.

**Audit Trail** -- ModulaCMS records an audit event alongside every content mutation, capturing the operation type, old and new values, request metadata, and a timestamp. This provides a complete history of changes for compliance and debugging.

## Access Control

**RBAC** -- Role-Based Access Control. ModulaCMS assigns each user a role, and each role has a set of permissions. Permissions follow the `resource:operation` format (e.g., `content:read`, `media:create`, `users:delete`).

Three bootstrap roles are created on first run:
- **admin** -- All permissions. Bypasses permission checks entirely.
- **editor** -- 36 permissions covering content, media, routes, datatypes, fields, and field types.
- **viewer** -- 5 read-only permissions (`content:read`, `media:read`, `routes:read`, `field_types:read`, `admin_field_types:read`).

System-protected roles and permissions cannot be deleted or renamed through the API.

**Permission Label** -- A string in the format `resource:operation` that identifies a specific permission. Examples: `content:create`, `roles:update`, `media:delete`.

## Interfaces

**TUI** -- Terminal User Interface. ModulaCMS runs an SSH server (default port 2233) that presents an interactive terminal interface. Connect with `ssh user@host -p 2233` to manage content, datatypes, fields, routes, users, and media through a keyboard-driven UI.

**Admin Panel** -- A server-rendered web interface for managing your CMS. Pages load as complete HTML with interactive updates powered by HTMX -- no JavaScript framework or SPA. The admin panel requires authentication and provides CSRF protection.

**REST API** -- The JSON API at `/api/v1` used by frontend applications, SDKs, and external integrations. All endpoints follow consistent patterns: collection endpoints at `/api/v1/{resource}`, item endpoints at `/api/v1/{resource}/?q={ulid}`, and standard HTTP methods for CRUD operations.

**Content Delivery** -- The public-facing endpoint at `GET /{slug}` that returns content trees in configurable output formats. Your frontend application calls this endpoint to fetch content. Supports `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, and `raw` formats via the `?format=` query parameter.

## Configuration

**modula.config.json** -- The configuration file loaded at startup. Contains all runtime settings: database connection, server ports, TLS certificates, S3 storage credentials, OAuth configuration, CORS policy, and more. Environment variables can be referenced using `${VAR}` syntax. If no config file exists on first run, the setup wizard creates one with defaults.

**Environment** -- The `environment` field in `modula.config.json` controls TLS behavior:
- `local` -- Uses self-signed certificates from `cert_dir`
- `http-only` -- Disables HTTPS entirely
- `development`, `staging`, `production` -- Uses automatic Let's Encrypt certificates

## Infrastructure

**Backup** -- ModulaCMS can create and restore backups containing a SQL dump of the database and all media files. Backups can be stored locally or in S3-compatible storage. Configured via the `backup_option` field in `modula.config.json`.

**Media Dimensions** -- Named width/height presets that define the sizes generated when an image is uploaded. The optimization pipeline produces a cropped and scaled variant for each preset where the source image is large enough. Presets that would require upscaling are skipped.

**S3-Compatible Storage** -- Object storage for media files, accessed through the AWS S3 API. Any S3-compatible provider works: AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, Cloudflare R2. Configured through `bucket_*` fields in `modula.config.json`.
