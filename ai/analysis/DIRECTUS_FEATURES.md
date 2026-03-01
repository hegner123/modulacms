# Directus Features (v11.15.x)

## Runtime Architecture
- **Language**: Node.js with TypeScript (full TypeScript codebase)
- **Framework**: Custom Express-like HTTP server; uses Knex.js for database abstraction, `pino` for logging, `sharp` for image processing, `nodemailer` for email
- **Process Model**: Single Node.js process. No worker threads or clustering built-in; horizontal scaling via multiple instances behind a load balancer with Redis for shared state (sessions, WebSocket coordination, collaborative editing)
- **Server Setup**: Runs as a single HTTP/HTTPS server serving both the REST/GraphQL API and the Vue.js-based admin panel (Data Studio) as static assets. WebSocket connections served from the same process for real-time features
- **Minimum Requirements**: Node.js 18+, a supported SQL database, and a `SECRET` environment variable for JWT signing
- **No native HTTPS termination** -- relies on reverse proxy (nginx, Caddy, etc.) for TLS in production

## Database Support
- **Officially Supported**: PostgreSQL (10+), MySQL (5.7.8+/8+), MariaDB (10.2.7+), SQLite 3, MS SQL (2017+), OracleDB, CockroachDB (21.1+)
- **Query Builder**: Knex.js underlying query builder. All `DB_*` environment variables passed directly to Knex's connection config
- **Database Introspection**: Uses `@directus/schema` (SchemaInspector) to introspect the physical database schema at runtime. Directus does not impose its own schema -- it mirrors the existing SQL database and layers metadata on top via `directus_*` system tables
- **Existing Database Support**: Can be installed on top of any existing SQL database without migration. The admin UI and API reflect the existing schema immediately
- **Dual-Schema Approach**: Physical database tables synchronized with metadata in `directus_*` system tables (directus_collections, directus_fields, directus_relations, etc.)

## Content Management
- **Collections**: Any database table becomes a "collection". Collections can be created via the Data Studio UI or by adding tables directly to the database
- **Items**: Full CRUD on items (rows) via REST API, GraphQL, SDK, or Data Studio. Batch operations supported (create, update, delete multiple items)
- **Relationships**:
  - **Many-to-One (M2O)**: Foreign key on the current collection
  - **One-to-Many (O2M)**: Inverse of M2O
  - **Many-to-Many (M2M)**: Auto-created junction collection with manual sort order support
  - **Many-to-Any (M2A)**: Polymorphic relationship. Junction table includes a column for related collection name. Used for page builders and reusable blocks
  - No explicit O2O type (achieved via M2O with unique constraint)
- **Ordering**: Manual sort order within relational fields (M2M, O2M, M2A)
- **Singleton Collections**: Bypass listing page, go directly to single item detail. Used for global settings
- **Archive (Soft Delete)**: Configurable per-collection archive field, archive value, unarchive value
- **Search**: Global search across items using `search` query parameter

## Content Schema / Content Modeling
- **UI-Driven Schema**: Collections, fields, and relationships created through Data Studio without SQL
- **Field Types (Database-Level)**: `String`, `Text`, `UUID`, `Integer`, `Big Integer`, `Float`, `Decimal`, `Boolean`, `DateTime`, `Date`, `Time`, `Timestamp`, `JSON`, `CSV`, `Hash`, `Binary`, Geometry types (`Point`, `LineString`, `Polygon`, `MultiPoint`, `MultiLineString`, `MultiPolygon`, `Geometry All`), `Alias` (virtual)
- **Interfaces (UI Input Components)**: 30+ built-in:
  - **Text & Numbers**: Input, Autocomplete, Block Editor, Code, Textarea, WYSIWYG (TinyMCE), Markdown, Tags
  - **Selection**: Toggle, Datetime picker, Repeater, Map (GeoJSON), Color picker, Dropdown, Dropdown Multiple, Icon picker, Checkboxes, Checkboxes Tree, Radio Buttons, Slider
  - **Relational**: File, Image, Files (M2M), Builder (M2A), Many to Many, One to Many, Tree View, Many to One, Translations
  - **Presentation**: Header, Divider, Button Links, Notice
  - **Groups**: Accordion, Detail Group, Raw Group
- **Displays**: Separate components for rendering values in list/table views
- **Validations**: Per-field using filter syntax. Client-side and server-side. Custom error messages
- **Conditional Fields**: Show/hide fields based on other field values using filter-rule syntax
- **Groups/Sections**: Accordion, Detail Group, Raw Group. Conditional logic at group level
- **Translations Interface**: Special relational interface auto-creating junction collection and `languages` collection
- **Field Width**: Half or full width per field

## Content Delivery / API
- **REST API**: Auto-generated for every collection at `/items/{collection}`. System endpoints at `/server`, `/users`, `/roles`, `/files`, `/activity`, `/revisions`
- **GraphQL API**: Auto-generated at `/graphql` (user-scoped) and `/graphql/system`. Queries, mutations, and subscriptions
- **Query Parameters**:
  - `fields` -- dot notation, wildcards (`*`, `*.*`, `*.*.*`)
  - `filter` -- 30+ operators
  - `search` -- full-text search
  - `sort` -- ascending default, `-` prefix for descending
  - `limit`, `offset`, `page`
  - `aggregate` -- `count`, `countDistinct`, `sum`, `avg`, `min`, `max`, `countAll`
  - `groupBy` -- group results
  - `deep` -- nested query parameters on relational fields
  - `alias` -- rename fields or fetch same nested data with different filters
  - `export` -- CSV, JSON, XML, or YAML
- **Filter Operators**: `_eq`, `_neq`, `_lt`, `_lte`, `_gt`, `_gte`, `_in`, `_nin`, `_null`, `_nnull`, `_contains`, `_ncontains`, `_icontains`, `_nicontains`, `_starts_with`, `_istarts_with`, `_ends_with`, `_iends_with`, `_between`, `_nbetween`, `_empty`, `_nempty`, `_intersects`, `_nintersects`, `_intersects_bbox`, `_nintersects_bbox`, `_regex`, `_some`, `_none`
- **Logical Operators**: `_and`, `_or` -- nestable
- **Dynamic Variables**: `$CURRENT_USER`, `$CURRENT_ROLE`, `$NOW`, `$NOW(<adjustment>)`
- **Functions**: `year()`, `month()`, `week()`, `day()`, `weekday()`, `hour()`, `minute()`, `second()`, `count()`
- **Deep Filtering**: `$FOLLOW(target-collection, relation-field)` for indirect relations

## Content Publishing
- **No native publishing workflow built-in**. Implemented through archive/status field pattern + permissions
- **Archive/Status Field**: Default status options: `draft`, `published`, `archived`. Custom statuses can be added
- **Workflow via Permissions**: Multi-stage publishing achieved by role-based permissions restricting status transitions
- **Content Versioning** provides draft-without-publish capability

## Content Versioning
- **Content Versioning** (introduced v10.7): Named versions that can be drafted and edited independently of the main item
- **Revisions**: Every mutation creates a revision record in `directus_revisions`. Stores full item state plus delta
- **Version Promotion**: A version can be "promoted" to become the main item
- **Activity Timeline**: Sidebar showing chronological list of all revisions
- **Diff/Comparison**: Content Comparison Modal with side-by-side view. Toggle between comparing against previous or latest revision
- **Restore**: Any revision can be applied/restored from comparison modal
- **Delta Field** (v11.1.2+): `directus_versions` `delta` field combines all revisions for pruning support

## Internationalization (i18n)
- **Translations Interface**: Special relational interface handling multilingual content via junction tables. Auto-creates `languages` collection and collection-specific translations junction table
- **Content Translations**: Translatable fields stored relationally, not as duplicated columns per language
- **Language Direction**: LTR and RTL per language
- **Default Language**: Configurable at field level; can default to current user's language
- **Interface Translations**: Data Studio UI supports 50+ languages
- **Deep Parameter**: Filter translations by language code in API requests
- **Progress Indicators**: Visual translation completeness indicators

## Media / Asset Management
- **File Library**: Central file management in `directus_files` with metadata (title, description, tags, location, dimensions, filesize, MIME type)
- **Storage Adapters**: Local filesystem, Amazon S3 (and S3-compatible: MinIO, DigitalOcean Spaces, Backblaze B2, Cloudflare R2), Google Cloud Storage, Azure Blob Storage. Multiple storage locations simultaneously
- **Image Transformations**: On-the-fly via `sharp`. URL parameters: `width`, `height`, `fit` (cover, contain, inside, outside), `quality` (1-100), `format` (jpg, png, webp, avif, tiff), `withoutEnlargement`, blur, flatten, tint/recolor, flip, flop
- **Transformation Presets**: Named presets via `?key=preset-name`. Can restrict to presets only
- **Folder Organization**: Virtual folders (not mirrored to storage). Nested hierarchy. Permissions on folders
- **Metadata**: Automatic EXIF/IPTC extraction. Custom metadata fields addable

## Authentication
- **Local**: Email/password with configurable password rules
- **OAuth 2.0**: Generic provider support via environment variables
- **OpenID Connect**: Standard OIDC support
- **LDAP**: Active Directory integration with auto-provisioning and role mapping
- **SAML**: SAML 2.0 SSO support
- **External JWT Validation**: Accept JWTs from external systems
- **Token Types**: Temporary (JWT), Refresh tokens, Static tokens (per-user permanent), Session tokens (cookie-based)
- **2FA**: TOTP two-factor authentication

## Authorization / Access Control
- **Policies** (v11+): Reusable permission sets attachable to roles or users. Additive
- **Roles**: Contain policies. Hierarchical (roles within roles)
- **CRUDS Permissions**: Per-collection Create, Read, Update, Delete, Share. Each with custom filters
- **Field-Level Permissions**: Toggle fields per CRUDS action
- **Item-Level Permissions**: Custom filter rules per action (e.g., "only read items where author equals $CURRENT_USER")
- **Validation Permissions**: Filter rules validating values before create/update
- **Presets**: Default values auto-applied when a role creates items
- **Public Role**: Defines unauthenticated API permissions. Disabled by default
- **Admin Role**: Bypasses all permission checks
- **Fail-Closed**: Missing permissions default to denied

## Admin Panel (Data Studio)
- **Framework**: Vue.js 3 SPA served from Directus server
- **Modules**: Content (Explore), Files, Users, Insights, Settings. Custom modules via extensions
- **Layouts**: Table, Cards, Calendar, Map, Kanban
- **Insights (Dashboards)**: Drag-and-drop dashboard builder. Panels: metric, list, time-series, bar chart, pie chart, custom. Auto-refresh. Exportable/importable
- **Bookmarks/Presets**: Save layout configurations as named bookmarks. User or global scope
- **Shares**: Public links to share specific items or filtered views without accounts
- **Theming**: Dark/light mode. Custom themes via extensions
- **Branding**: Project name, color, logo, foreground/background images, custom CSS, public notes
- **Live Preview**: Preview content on external frontends within Data Studio
- **Collaborative Editing** (v11.15+): Multiple users edit same item with real-time presence and field-level locking. WebSockets + Redis
- **AI Assistant** (v11.15 GA): Multi-provider (OpenAI, Anthropic, Google Gemini, Ollama). Context-aware. Integrated into visual editor
- **Visual Editor / Studio Module**: WYSIWYG page editing with live preview split pane

## Extension System
- **Extension Types**:
  - **App Extensions** (frontend, Vue.js): Interfaces, Displays, Layouts, Modules, Panels, Themes
  - **API Extensions** (backend, Node.js): Hooks, Endpoints, Operations
  - **Bundles**: Package multiple extensions together
- **Extension SDK**: `@directus/extensions-sdk` with CLI, Vue composables, TypeScript types
- **Sandboxed vs Non-Sandboxed**: Sandboxed API extensions run in isolated environment with restricted imports. Non-sandboxed have full Node.js access
- **Marketplace**: Registry of extensions. Installable from Data Studio UI (sandboxed) or file system (non-sandboxed)

## Webhooks
- **Legacy Webhooks**: Configurable via Data Studio. Name, method, URL, status, events, collections, payload, custom headers
- **Replaced by Flows**: Legacy system superseded by Flows automation with more flexible functionality
- **Actions vs Filters**: "Action" hooks fire after (non-blocking), "filter" hooks fire before (can modify/reject)

## Flows (Automation)
- **Visual Workflow Editor**: Drag-and-drop flow builder. One trigger, chain of operations
- **Triggers** (5 types): Event Hook, Webhook, Schedule (CRON), Another Flow, Manual
- **Operations** (built-in): Create/Update/Delete/Read Data, Send Email, Send Notification, Run Script, Webhook/Request, Condition, Transform Payload, Log, Trigger Flow, Sleep
- **Data Chain**: Each operation receives and modifies a shared data payload
- **Custom Operations**: Via extensions
- **Conditional Logic**: Branch with resolve/reject paths

## SDKs & Client Libraries
- **Official**: `@directus/sdk` (TypeScript/JavaScript). Composable client: `createDirectus(url).with(rest()).with(graphql()).with(realtime()).with(authentication())`. Modular, tree-shakeable. Node.js, browsers, edge runtimes
- **Community**: PHP (`directus/sdk`), Dart/Flutter, Go

## CLI
- `bootstrap` -- Initialize database + first admin. Idempotent
- `database install` -- Create system tables
- `database migrate:latest` -- Run pending migrations
- `schema snapshot` -- Export schema to YAML/JSON
- `schema apply` -- Apply schema snapshot (diff-based)
- `schema diff` -- Show differences
- `users create`, `roles create`, `count <collection>`
- Custom migrations in `extensions/migrations/`

## Deployment
- **Self-Hosted**: Standard Node.js application
- **Docker**: Official image `directus/directus`
- **Directus Cloud**: Professional ($99/mo), Business ($499/mo), Enterprise (custom)
- **Scaling**: Stateless; horizontal via multiple instances + shared DB + Redis

## Email
- **Transport**: Built-in via `nodemailer`. SMTP, Sendmail, third-party services
- **Templates**: LiquidJS templating engine. Custom templates in `EMAIL_TEMPLATES_PATH`
- **Flow Integration**: "Send Email" operation in Flows

## Backup & Restore
- **Schema Snapshot**: `schema snapshot` exports data model as YAML/JSON. Does NOT include content data
- **Schema Apply**: Diff-based application of snapshots
- **No built-in content/data backup**. Database backups handled externally
- **No built-in media backup**

## Configuration System
- **Primary Method**: Environment variables via `.env` file or system env
- **200+ configurable variables**: General, Database, Auth, Storage, Email, Cache, CORS, Rate Limiting, Extensions, Telemetry, Logging, Security
- **Runtime Settings**: Some in `directus_settings` table, editable via Data Studio

## Observability / Logging
- **Logger**: `pino` for structured JSON logging
- **HTTP Logger**: `pino-http` for request logging
- **Log Levels**: fatal, error, warn, info, debug, trace
- **Real-time Logs**: WebSocket endpoint at `/websocket/logs` for streaming
- **Telemetry**: Configurable anonymous usage data

## Audit Trail
- **Activity Collection** (`directus_activity`): Logs every action (Create, Update, Delete, Comment, Login). User, action, timestamp, IP, user agent, collection, item ID
- **Revisions Collection** (`directus_revisions`): Full item state at time of mutation plus delta
- **Per-Item Timeline**: Sidebar showing revision history
- **Configurable Tracking**: Per-collection toggle (activity + revisions, activity only, or nothing)
- **Read-Only**: Activity and revisions cannot be deleted

## Pricing / Licensing
- **License**: BSL 1.1 (Business Source License) since v10.0.0
  - Non-production use always free
  - Organizations < $5M annual revenue: free production use
  - Code converts to GPLv3 after 3 years
  - Older versions (v9.x-) remain GPLv3
- **Self-Hosted**: Free for < $5M revenue. Enterprise: custom pricing
- **Directus Cloud**: Professional (~$99/mo), Business (~$499/mo), Enterprise (custom)

## Unique / Notable Features
- **Database Introspection**: Wraps any existing SQL database without schema modifications. Admin UI and APIs generated dynamically. Core differentiator
- **Real-Time via WebSockets**: Full bidirectional WebSocket support. CRUD over WebSocket. GraphQL subscriptions
- **Insights / Dashboards**: Built-in analytics with drag-and-drop panels
- **Flows Automation**: Visual no-code workflows with 5 triggers, 12+ operations
- **Data Sharing**: Public links for filtered views without accounts
- **Collaborative Editing** (v11.15): Native real-time with field-level locking
- **AI Assistant** (v11.15 GA): Multi-provider (OpenAI, Anthropic, Google, Ollama)
- **MCP Support** (v11.13): Native Model Context Protocol server for AI tools
- **Deployment Module** (v11.15): Trigger Vercel deployments from Data Studio
- **Geospatial Support**: First-class GeoJSON with Map interface and spatial filters
- **Response Format Agnostic**: Export to CSV, XML, YAML via `export` parameter
