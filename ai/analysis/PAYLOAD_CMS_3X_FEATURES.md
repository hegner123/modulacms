# Payload CMS 3.x Feature Inventory

Comprehensive technical feature inventory for Payload CMS version 3.x (latest stable as of February 2026). Based on official documentation, GitHub source, and multiple independent sources.

---

## 1. Runtime Architecture

- **Language:** TypeScript (full stack). Config, hooks, access control, custom components -- all TypeScript.
- **Framework:** Built natively on Next.js App Router. Payload installs directly into your Next.js `/app` directory. The admin panel and HTTP layer (REST, GraphQL) are Next.js routes.
- **Process Model:** Single Node.js process running Next.js. No separate CMS server -- Payload IS the Next.js app.
- **Server Setup:** Payload's admin panel, REST API, and GraphQL API all run as Next.js route handlers within the same application.
- **React Version:** Uses the React compiler for optimized builds. Admin panel uses React Server Components by default.
- **Dependencies:** 27 dependencies in 3.0 (down from 88 in 2.0). Modular architecture -- you import only what you need.
- **Runtime Requirements:** Node.js (LTS). Can also run on serverless platforms (Vercel, Cloudflare) with caveats around connection pooling and timeouts.
- **Non-Next.js Usage:** Can be used with Astro, SvelteKit, Remix, or any Node.js environment, though Next.js is the primary integration target.

## 2. Database Support

- **Adapter Pattern:** Database-agnostic via a `db` property in `payload.config.ts`. Each adapter is an npm package installed separately.
- **Officially Supported Adapters:**
  - **MongoDB** (`@payloadcms/db-mongodb`) -- uses Mongoose ORM. Documents stored as single MongoDB documents regardless of complexity (arrays, blocks, localization all in one document). No migrations needed for schema changes.
  - **PostgreSQL** (`@payloadcms/db-postgres`) -- uses Drizzle ORM. Supports transactions, read replicas, custom schemas, and migrations. Arrays and blocks each get their own relational tables.
  - **SQLite** (`@payloadcms/db-sqlite`) -- uses Drizzle ORM with libSQL. Supports Turso for edge deployment. Point Field not yet supported on SQLite.
- **Configuration:** Set via `db` property in config: `db: mongooseAdapter({ url: process.env.DATABASE_URL })`
- **Feature Parity:** Nearly all Payload features work identically across all three adapters (localization, arrays, blocks, etc.). Exception: Point Field missing from SQLite.
- **Direct DB Access:** After initialization, the full Drizzle client is exposed for direct SQL queries, bypassing Payload's abstraction when needed.
- **Development Mode:** In dev, Drizzle can `push` schema changes directly to the database. For production, you create formal migrations.

## 3. Content Management

- **Collections:** Primary content containers. Defined in `payload.config.ts` with fields, hooks, access control, admin UI config, versioning, and auth settings. Correspond to database tables/collections.
- **Globals:** Singleton documents (e.g., site settings, navigation, footer). Same field system as collections but only one document per Global.
- **CRUD Operations:** Full create, read, update, delete across Local API, REST API, and GraphQL. All three APIs share the same query language.
- **Relationships:** `relationship` field type supports one-to-one, one-to-many, and polymorphic relationships (referencing multiple collections). Configurable with `hasMany`, `relationTo` (single collection or array of collections).
- **Join Field:** Virtual field (no data stored) that surfaces the reverse side of a relationship. Operates at the database level for performance. Enables bidirectional relationship traversal without data duplication.
- **Ordering:** When `orderable` is enabled on a collection, Payload uses fractional indexing for efficient document ordering.
- **Tree Structures:** Not a built-in primitive. Achieved by adding a self-referencing `relationship` field (e.g., `parent`) on a collection. The `@payloadcms/plugin-nested-docs` plugin adds parent-child relationships, breadcrumbs, and tree navigation.
- **Blocks Field:** Stores an array of heterogeneous objects. Each "block" has its own schema. Blocks can be mixed in any order. In relational DBs, each distinct block type gets its own table.
- **Array Field:** Stores an array of homogeneous objects (same schema for each item). In relational DBs, each array field gets its own table.
- **Trash / Soft Delete:** Built-in since v3.49.0. `trash: true` on a collection enables soft deletes. Automatically adds `deletedAt` and `deletedBy` fields. Access control can distinguish between trashing and permanent deletion.

## 4. Content Schema / Content Modeling

- **Config Approach:** Code-first. All schemas defined in TypeScript via `payload.config.ts`. No GUI schema builder -- you write code.

### Data Fields (store values in database)

| Field Type | Description |
|---|---|
| `text` | Single-line string. Supports `hasMany` for multiple strings. |
| `textarea` | Multi-line string. |
| `number` | Integer or float. Supports `hasMany`, min/max validation. |
| `email` | Email string with built-in validation. |
| `select` | Dropdown/multi-select. Options defined in config. Supports `hasMany`. |
| `checkbox` | Boolean true/false. |
| `date` | Date/datetime with configurable picker. |
| `point` | GeoJSON point (longitude, latitude). Not supported on SQLite. |
| `radio` | Radio button group (single select from options). |
| `code` | Code editor field with syntax highlighting (Monaco editor). |
| `json` | Arbitrary JSON. Supports JSON Schema for validation and editor typeahead. |
| `richText` | Rich text via Lexical editor. Highly configurable feature set. |
| `relationship` | Reference to document(s) in other collection(s). Supports polymorphic relationships. |
| `upload` | Reference to an upload-enabled collection. Displays file/image preview. |
| `array` | Ordered list of sub-documents with identical schema. |
| `blocks` | Ordered list of sub-documents with heterogeneous schemas (each block type has its own fields). |
| `group` | Named group of fields stored under a single key. Creates a nested object. |
| `join` | Virtual field -- surfaces reverse relationships. No data stored. |

### Presentational / Layout Fields (no data stored)

| Field Type | Description |
|---|---|
| `tabs` (unnamed) | Organizes fields into tabbed UI. No data nesting. |
| `tabs` (named) | Like group but rendered as tabs. Creates nested object. |
| `row` | Arranges fields horizontally in the admin panel. |
| `collapsible` | Wraps fields in a collapsible section. |
| `ui` | Blank field slot for injecting custom React components. |

### Virtual Fields

- Any field can be marked `virtual: true` to indicate it is not stored in the database but is computed or derived.

### Field-Level Features

- **Validation:** Custom validation functions per field. Runs both client-side and server-side. Built-in validations for email, required, min/max, etc.
- **Conditional Logic:** `admin.condition` function on any field -- show/hide based on sibling field values.
- **Access Control:** Per-field `access.read` and `access.update` functions returning boolean.
- **Default Values:** Static or function-based defaults.
- **Hooks:** Field-level `beforeChange`, `afterChange`, `beforeRead`, `afterRead` hooks.
- **Custom Components:** Any field's admin UI can be completely replaced with custom React components.
- **Localization:** Per-field `localized: true` flag.
- **Indexing:** `index: true` for database-level indexes on frequently queried fields.
- **Hidden Fields:** `hidden: true` removes from API responses. `admin.hidden: true` hides from admin UI only.

## 5. Content Delivery / API

### Three APIs

- **Local API:** Direct-to-database, no HTTP overhead. Used in server components and server functions. Full type safety. Fastest option.
- **REST API:** Standard HTTP endpoints auto-generated for every collection and global. Pattern: `GET /api/{collection-slug}`.
- **GraphQL:** Full GraphQL API with auto-generated schema, queries, and mutations. Includes GraphQL Playground. Only initialized if used (lazy loading in 3.0).

### Query Language (shared across all three APIs)

**Operators:**

| Operator | Description |
|---|---|
| `equals` | Exact match |
| `not_equals` | Not equal |
| `greater_than` | Numeric/date comparison |
| `greater_than_equal` | Numeric/date comparison |
| `less_than` | Numeric/date comparison |
| `less_than_equal` | Numeric/date comparison |
| `like` | Case-insensitive, all words must be present in any order |
| `contains` | Case-insensitive substring match |
| `in` | Value in comma-delimited list |
| `not_in` | Value NOT in comma-delimited list |
| `all` | Must contain all values (MongoDB only) |
| `exists` | Field exists (true) or not (false) |
| `near` | Point field distance query: `longitude,latitude,maxDistance,minDistance` |
| `within` | Point field GeoJSON area containment |
| `intersects` | Point field GeoJSON area intersection |

- **AND/OR Logic:** Queries support nested `and` / `or` arrays for complex boolean logic.
- **Nested Properties:** Dot notation for querying into relationships: `'artists.featured': { exists: true }`.

### Pagination

- `limit` -- number of documents per page
- `page` -- page number
- `pagination: false` -- disable pagination, return all results
- Response includes: `docs`, `totalDocs`, `totalPages`, `page`, `pagingCounter`, `hasPrevPage`, `hasNextPage`, `prevPage`, `nextPage`

### Depth Control

- `depth` parameter controls how many levels of relationships are populated (resolved to full documents vs. returning just IDs).
- Default depth is configurable globally and per-collection.
- GraphQL depth is determined by query shape, not a parameter.

### Select & Populate

- **Select API:** Specify exactly which fields to return. Reduces response size and skips processing of unselected fields (including their hooks).
- **Populate:** Control which fields are returned when populating relationships. `defaultPopulate` on relationship fields enforces specific fields globally.

### Locale Parameters

- `locale` query parameter returns content in specified locale.
- `fallback-locale` parameter controls fallback behavior. Can be a locale code, `'null'`, `'false'`, or `'none'` to disable fallback.

### Sorting

- `sort` parameter accepts field names. Prefix with `-` for descending order.

## 6. Content Publishing

- **Draft/Publish System:** Enabled per-collection via `versions.drafts: true`. Adds a `_status` field to documents with values `'draft'` or `'published'`.
- **Save Actions:** When drafts are enabled, the admin UI replaces the "Save" button with "Save Draft" and "Publish" actions.
- **Autosave:** When `versions.drafts.autosave: true` is enabled, the admin UI automatically saves changes as draft versions as you edit. Payload creates one autosave version and updates it (not a new version per keystroke).
- **Scheduled Publishing:** Built-in. Set a future date to publish or unpublish a document. Requires the Jobs Queue to be configured (jobs must be processed for scheduled events to fire).
- **Scheduled Unpublishing:** Documents can be scheduled to revert to draft status after a given date.

## 7. Content Versioning

- **Version History:** Enabled per-collection via `versions: true` (or `versions: { maxPerDoc: N }`). All versions stored in a separate collection/table.
- **Version Storage:** Each change creates a new version record. For relational DBs, versions get their own table.
- **Restore:** Any historical version can be restored to become the current document.
- **Compare/Diff:** Admin UI provides a compare view that highlights line-level differences between versions.
- **Retention:** `maxPerDoc` controls how many versions to keep per document. Older versions are pruned automatically.
- **Autosave Drafts:** Autosave creates/updates a single draft version to avoid version bloat. Only one autosave version exists at a time.
- **Version Metadata:** Each version records the user who made the change and the timestamp.
- **API Access:** Versions are queryable via the API -- you can list, retrieve, and restore versions programmatically.

## 8. Internationalization (i18n)

### Content Localization

- **Strategy:** Field-level localization. Not document-level -- each field can independently opt into localization with `localized: true`.
- **Configuration:** Set in `payload.config.ts` via `localization` key with `locales` array, `defaultLocale`, and `fallback` boolean.
- **Locale Definition:** Each locale has a `label`, `code`, and optional `fallbackLocale` (per-locale fallback chains).
- **Fallback:** If `fallback: true` globally, missing translations automatically fall back to the fallback locale. Can be disabled per-request.
- **API Support:** `locale` and `fallback-locale` query parameters on REST/GraphQL. Local API accepts `locale` option.
- **Admin UI:** Locale switcher in the admin panel. Editors can toggle between locales while editing.

### Admin UI i18n

- **Separate from content localization.** The `i18n` config controls the language of the admin panel itself (labels, buttons, messages).
- **Built-in translations** for the admin interface in multiple languages.
- **Custom translations** can be added via the i18n config.

## 9. Media / Upload Management

- **Upload Collections:** Any collection can be upload-enabled by adding `upload: true` (or upload config object). This adds file upload capabilities to that collection.
- **Image Resizing:** Define `imageSizes` array in upload config. Payload auto-generates resized versions using `sharp`. Each size specifies `name`, `width`, `height`, and `fit` mode.
- **Focal Point:** Admin UI provides a focal point selector when `imageSizes` or `resizeOptions` are defined. Setting a focal point regenerates all image sizes centered on that point.
- **Image Cropping:** Built-in crop tool in the admin panel.
- **MIME Filtering:** `mimeTypes` array in upload config restricts allowed file types.
- **File Size Limits:** Configurable max file size.
- **Admin Thumbnails:** Upload collections display thumbnails in the admin list view and document views.
- **Local Storage:** Files stored on the local filesystem by default. Configurable `staticDir`.
- **Disable Local Storage:** `disableLocalStorage: true` when using cloud storage exclusively.

### Cloud Storage Adapters

- **Plugin:** `@payloadcms/plugin-cloud-storage` provides the adapter pattern.
- **Official Adapters:**
  - Amazon S3 (`@payloadcms/storage-s3`)
  - Google Cloud Storage (`@payloadcms/storage-gcs`)
  - Azure Blob Storage (`@payloadcms/storage-azure`)
  - Vercel Blob Storage (`@payloadcms/storage-vercel-blob`)
- All automatically resized images are also uploaded to cloud storage.

## 10. Authentication

- **Built-in Auth:** Any collection can be auth-enabled by adding `auth: true`. This adds email/password fields, login/logout/token endpoints, and session management.
- **No Separate User Service:** Auth is a collection feature, not a separate module. You can have multiple auth-enabled collections (e.g., `users`, `admins`, `customers`).
- **HTTP-Only Cookies:** Default authentication method. Secure against XSS. Cannot be read by client-side JavaScript.
- **JWT (JSON Web Tokens):** Generated on login, refresh, and me operations. Can be used as alternative to cookies. Configurable expiration.
- **API Keys:** Non-expiring keys per user. Enabled via `auth.useAPIKey: true`. Authenticated via `Authorization: {collection-slug} API-Key {key}` header.
- **Session Support:** `auth.strategies` supports custom authentication strategies. Sessions enabled by default (`auth.disableSession: false`).
- **Operations:**
  - Login (`POST /api/{collection}/login`)
  - Logout (`POST /api/{collection}/logout`)
  - Me (`GET /api/{collection}/me`)
  - Refresh token (`POST /api/{collection}/refresh-token`)
  - Forgot password (`POST /api/{collection}/forgot-password`)
  - Reset password (`POST /api/{collection}/reset-password`)
  - Verify email (`POST /api/{collection}/verify/{token}`)
  - Unlock account (`POST /api/{collection}/unlock`)
- **Email Verification:** `auth.verify: true` sends verification emails on registration.
- **Max Login Attempts:** `auth.maxLoginAttempts` and `auth.lockTime` for brute force protection.
- **OAuth:** Not built into core. Implemented via custom strategies or community plugins. Enterprise SSO (SAML/OAuth 2.0) available for enterprise customers.

## 11. Authorization / Access Control

- **Function-Based:** Access control is defined via functions, not static roles. Functions return either a boolean (allow/deny) or a query constraint (filter documents the user can access).
- **Collection-Level Access:** Functions for `create`, `read`, `update`, `delete` operations. Each receives the authenticated user and request context.
  ```ts
  access: {
    read: ({ req: { user } }) => user?.role === 'admin' ? true : { author: { equals: user?.id } },
    create: ({ req: { user } }) => Boolean(user),
    update: ({ req: { user } }) => user?.role === 'admin',
    delete: ({ req: { user } }) => user?.role === 'admin',
  }
  ```
- **Field-Level Access:** `access.read` and `access.update` on individual fields. Return boolean only (no query constraints at field level).
- **Global Access:** Same pattern applied to Globals.
- **Query Constraints:** Collection access functions can return a `Where` query instead of a boolean. This filters results at the database level -- users only see documents matching the constraint.
- **Role Patterns:** Payload does not impose a role system. You build your own RBAC by checking user properties in access functions. Common pattern: add a `role` field to your users collection.
- **Cascading Access Check:** Payload evaluates all access functions on login and returns a permissions object reflecting what the user can do across the entire application.
- **Admin UI Enforcement:** Access control is enforced in the admin panel -- fields and collections the user cannot access are hidden automatically.

## 12. Admin Panel

- **Technology:** React-based, built on Next.js App Router. Uses React Server Components by default.
- **Custom Components:** Nearly every part of the admin UI can be replaced with custom React components via config. Server Components by default; Client Components supported with `'use client'` directive.
- **Root Components:** `admin.components.views` allows replacing or adding top-level views (Dashboard, Account, etc.).
- **Dashboard Customization:**
  - `admin.components.beforeDashboard` -- inject components before default dashboard content
  - `admin.components.afterDashboard` -- inject after
  - Full dashboard replacement via custom view
- **Custom Views:** Create entirely new admin pages or replace built-in ones (list view, edit view, etc.).
- **Live Preview:** Admin panel renders an iframe loading your frontend. Communicates via `window.postMessage`. Two modes:
  - **Server-side:** Calls `router.refresh()` to re-render with fresh data from Local API. Recommended for Next.js.
  - **Client-side:** Listens for postMessage events and re-renders client-side.
- **Custom CSS:** `custom.scss` file for global style overrides. Payload uses CSS variables extensively -- override `--theme-elevation-*` and color variables for complete theming.
- **Branding:** Custom logo and icon components via `admin.components.graphics.Logo` and `admin.components.graphics.Icon`.
- **Dark Mode:** Built-in. Users choose light/dark/auto (OS detection). Preferences persisted per user across sessions. CSS variables auto-adapt to theme.
- **White-Labeling:** Full white-label support. Replace all branding, logos, colors, and views.
- **Tailwind CSS:** Can be themed with Tailwind CSS 4 via custom CSS integration.
- **UI Component Library:** `@payloadcms/ui` package provides reusable components (buttons, modals, drawers, etc.) for use in custom components.

## 13. Plugin / Extension System

- **Architecture:** Plugins are functions that receive a Payload config and return a modified config. They can add collections, globals, fields, hooks, admin views, endpoints, or anything else that exists in the config.
  ```ts
  const myPlugin = (incomingConfig: Config): Config => {
    // modify and return config
  }
  ```
- **Composability:** Plugins can be stacked. Each plugin transforms the config sequentially.

### Hooks System

**Collection Hooks:**
- `beforeOperation` -- modify operation arguments
- `beforeValidate` -- add/format data before validation
- `beforeChange` -- modify data before save (after validation)
- `beforeRead` -- modify document before output transformation
- `beforeDelete` -- execute logic before deletion
- `afterChange` -- execute logic after create/update (receives `doc`, `previousDoc`)
- `afterRead` -- modify document as last step before return
- `afterDelete` -- execute logic after deletion
- `afterOperation` -- modify operation result
- `afterError` -- error handling/logging (Sentry, DataDog, etc.)

**Auth Hooks (on auth-enabled collections):**
- `beforeLogin`, `afterLogin`, `afterLogout`, `afterRefresh`, `afterMe`, `afterForgotPassword`
- `refresh` -- replace default refresh behavior
- `me` -- replace default me behavior

**Global Hooks:** Same lifecycle hooks as collections (minus delete-related hooks).

**Field Hooks:** `beforeChange`, `afterChange`, `beforeRead`, `afterRead` at the individual field level.

**Root Hooks:** Not collection/global specific. For application-level side effects.

**Hook Context:** Custom `context` object passed between hooks within the same request. Useful for sharing data (e.g., fetch from 3rd party in `beforeChange`, use in `afterChange` without re-fetching).

### Official Plugins

| Package | Purpose |
|---|---|
| `@payloadcms/plugin-seo` | Injects SEO fields (title, description, Open Graph) into collections/globals |
| `@payloadcms/plugin-nested-docs` | Parent-child hierarchies, breadcrumbs, tree navigation |
| `@payloadcms/plugin-form-builder` | Dynamic form creation with field types, validation, submission handling, email notifications |
| `@payloadcms/plugin-redirects` | URL redirect management |
| `@payloadcms/plugin-search` | Cross-collection search with automatic indexing |
| `@payloadcms/plugin-stripe` | Stripe payment integration with two-way sync |
| `@payloadcms/plugin-cloud-storage` | Abstracts cloud storage providers for uploads |
| `@payloadcms/richtext-lexical` | Lexical-based rich text editor (default in 3.x) |

## 14. Webhooks

- **No dedicated webhook system.** Payload does not have a built-in webhook configuration UI or webhook delivery system.
- **Implementation via Hooks:** Webhook-like behavior is implemented using collection/global hooks (e.g., `afterChange` hook that sends an HTTP request to an external URL).
- **Custom Endpoints:** In 3.x with Next.js, custom server-side logic is handled via Next.js API Routes rather than custom CMS endpoints.
- **Third-Party Integration:** Hooks can call any external service (GitHub Actions, Slack, Zapier, etc.) via standard HTTP requests.

## 15. SDKs & Client Libraries

- **Official SDK:** `@payloadcms/sdk` -- TypeScript SDK for querying the REST API with full type safety. Currently in **beta** (may have breaking changes in minor versions).
  - Supports: `find`, `findByID`, `create`, `update`, `delete`, `count`, `auth.login`
  - Type-safe based on generated types from your config
- **Local API:** The primary "SDK" for server-side usage. Direct-to-database, no HTTP. Used in React Server Components, server functions, and Node.js scripts.
- **REST Client Patterns:** `qs-esm` package recommended for building complex query strings for the REST API.
- **No Official SDKs for Other Languages:** No Go, Swift, Python, etc. SDKs. Community-maintained options may exist.

## 16. CLI

- **`create-payload-app`:** Interactive scaffolding tool. Creates a new Payload + Next.js project with template selection and database adapter configuration.
- **`payload` CLI Commands:**
  - `payload migrate` -- run pending migrations
  - `payload migrate:create` -- create a new migration file
  - `payload migrate:down` -- roll back all migrations
  - `payload migrate:refresh` -- drop all, re-run all migrations
  - `payload migrate:reset` -- reset migration state
  - `payload migrate:status` -- show migration status
  - `payload generate:types` -- generate TypeScript interfaces from config (outputs `payload-types.ts`)
  - `payload generate:db-schema` -- generate database schema
  - `payload generate:importmap` -- generate import map
  - `payload jobs:run` -- manually process queued jobs
  - `payload run` -- start the application
  - `payload info` -- display Payload project information

## 17. Deployment

- **Self-Hosted:** Deploy anywhere Node.js runs. Standard Node.js server deployment.
- **Docker:** Multi-stage Docker builds supported. Set `output: 'standalone'` in `next.config.js`.
- **Vercel:** Works but with caveats. Serverless functions have timeout limits and connection pooling issues with Drizzle. Better suited for frontend-only Vercel + separate Payload server.
- **Cloudflare:** One-click hosting option available.
- **Payload Cloud:** Official hosted platform. Managed infrastructure. See Pricing section.
- **Serverless Considerations:** Payload's connection pool is designed for persistent servers. Serverless creates/destroys connections at rates the pool was not designed for. Admin operations (bulk imports, migrations, media processing) may exceed serverless timeout limits.
- **Recommended Self-Host:** Persistent VPS (Railway, Render, DigitalOcean, Hetzner) with managed database. Payload as its own Node process.

## 18. Email

- **Adapter Pattern:** Email functionality via `email` property in config. Adapters are separate packages.
- **Official Adapters:**
  - `@payloadcms/email-nodemailer` -- wraps Nodemailer. Supports any Nodemailer transport (SMTP, SendGrid, Resend, etc.). Most common choice.
  - `@payloadcms/email-resend` -- lightweight Resend adapter. Preferred for Vercel deployments (smaller bundle than Nodemailer).
- **Third-Party Adapters:** Community adapter for SendGrid (`@zapal-tech/payload-email-sendgrid`).
- **Built-in Email Use Cases:** Password reset, email verification, form submissions (via form-builder plugin).
- **Custom Email:** Single transporter configured globally. For separate bulk vs. transactional email, use hooks to send via different services.
- **Templates:** Email templates can be customized via the config. HTML email support.

## 19. Backup & Restore

- **No Built-in Backup System.** Payload does not have a dedicated backup/restore feature.
- **Migration System:** Serves as the primary mechanism for database schema management. Not a data backup tool.
- **Recommended Practice:** Use database-native tools (`pg_dump`, `mongodump`, etc.) for data backups before migrations.
- **MongoDB:** Schema-less nature means migrations are rarely needed. Data backup via `mongodump`.
- **PostgreSQL/SQLite:** Migrations required for production schema changes. Backup via `pg_dump` or file copy.

## 20. Configuration System

- **Config File:** `payload.config.ts` at project root (customizable via `PAYLOAD_CONFIG_PATH` env var).
- **TypeScript-First:** Strongly typed config. IDE provides autocomplete and type checking.
- **`buildConfig` Function:** Main config is wrapped in `buildConfig()` which validates and processes the config.
- **Key Config Properties:**
  - `collections` -- array of collection configs
  - `globals` -- array of global configs
  - `db` -- database adapter
  - `editor` -- rich text editor (Lexical by default)
  - `email` -- email adapter
  - `admin` -- admin panel customization (components, meta, theme, etc.)
  - `localization` -- locale configuration
  - `i18n` -- admin UI language
  - `plugins` -- array of plugin functions
  - `serverURL` -- base URL of the application
  - `secret` -- JWT signing secret
  - `cors` -- CORS origins
  - `csrf` -- CSRF protection config
  - `rateLimit` -- rate limiting config
  - `upload` -- global upload settings
  - `graphQL` -- GraphQL configuration (disable, custom schema extensions)
  - `jobs` -- jobs queue configuration
  - `telemetry` -- opt-in/out of anonymous telemetry
- **Environment Variables:** Accessed via `process.env` in config. Client-side env vars prefixed with `NEXT_PUBLIC_`. `.env` file support.

## 21. Observability / Logging

- **Built-in Logging:** Payload has internal logging. Custom log levels configurable.
- **`afterError` Hook:** Collection-level and root-level error hooks for sending errors to external services (Sentry, DataDog, etc.).
- **No Built-in Observability Dashboard.** Relies on external monitoring tools.
- **Community Solutions:**
  - `payload-auditor` plugin -- event tracking, auditing, operation logging with per-collection and per-operation granularity.
  - `customLogger` in payload-auditor for tailored log output.
- **Custom Logging:** Implementable via hooks at any lifecycle point.

## 22. Audit Trail

- **Version History as Audit Trail:** When versioning is enabled, every change creates a version record that includes:
  - The full document state at that point
  - Which user made the change
  - Timestamp of the change
- **Compare View:** Admin UI visualizes differences between versions with line-level delta highlighting.
- **Enterprise Audit Logs:** Payload Enterprise offers dedicated audit log functionality for compliance requirements. This is a paid feature.
- **Limitations:** Core versioning tracks document-level changes but does not provide a dedicated audit log with operation type, IP address, or request metadata out of the box. For those, use community plugins or hooks.

## 23. Testing

- **Local API for Testing:** The Local API is the primary testing tool. Initialize Payload in test setup (`payload.init()`) and call operations directly without HTTP.
- **Test Pattern:** Use `beforeAll` to initialize Payload with a test database, then use `payload.find()`, `payload.create()`, etc. in test cases.
- **No Official Testing Framework:** Payload does not ship a testing library. Use Jest, Vitest, or any Node.js test framework.
- **Direct Database Access:** `payload.db` exposes the database client for direct queries in tests, bypassing hooks and validation.
- **REST API Testing:** Custom fetch implementations allow testing REST endpoints without a running dev server.

## 24. Migration

- **Automatic Schema Management (Dev):** In development, Drizzle `push` mode automatically syncs your config changes to the database without migration files.
- **Migration Commands:**
  - `payload migrate:create` -- generate a migration file based on schema diff
  - `payload migrate` -- apply pending migrations
  - `payload migrate:down` -- roll back all migrations
  - `payload migrate:fresh` -- drop all tables, re-run all migrations
  - `payload migrate:reset` -- reset migration tracking
  - `payload migrate:status` -- show which migrations have been applied
- **MongoDB:** Rarely needs migrations. Schema changes are reflected automatically since MongoDB is schema-less.
- **PostgreSQL/SQLite:** Migrations required for production deployments. Generated by Drizzle based on config changes.
- **Production Workflow:** Disable `push: false` in production. Only apply changes via migration files.

## 25. Pricing / Licensing

- **License:** MIT. Free and open source forever. No feature gating in the core product.
- **Self-Hosting:** Free. You pay only for infrastructure (hosting, database, storage).
- **Payload Cloud Pricing (Managed Hosting):**

| Plan | Price | Compute | Database | File Storage | Notes |
|---|---|---|---|---|---|
| Standard | $35/month | 512MB RAM, serverless | 3GB | 30GB | 50M serverless RPUs |
| Pro | $199/month | Dedicated cluster, HA | 30GB | 150GB | High-availability compute |
| Enterprise | Custom (from $10,000/year) | Custom | Custom | Custom | SSO, AI features, visual editor, A/B testing |

- **Overage Pricing:** $0.50/GB database, $0.20/GB bandwidth, $0.02/GB file storage.
- **Enterprise Features (paid):** SSO (SAML/OAuth 2.0), AI Auto-Embedding, Publishing Workflows, Visual Editor, Static A/B Testing, dedicated audit logs.
- **Note:** Payload was acquired by Figma, which may affect future pricing and cloud offerings.

## 26. Unique Features

### Local API (Bypass HTTP)

- Direct database access from server-side code with zero HTTP overhead.
- Full type safety with generated types.
- Available in React Server Components, Next.js server functions, and any Node.js code.
- Same operations as REST/GraphQL but orders of magnitude faster for server-side use.

### Lexical Rich Text Editor

- Built on Meta's Lexical framework.
- **23 official features** including:
  - Text formatting: Bold, Italic, Underline, Strikethrough, Subscript, Superscript, InlineCode
  - Structure: Paragraph, Heading (H1-H6 configurable), Blockquote, HorizontalRule
  - Lists: Ordered, Unordered, Checklist
  - Media: Upload (all file types), Relationship (block-level document references)
  - Layout: Align (left/center/right/justify), Indent
  - Links: Internal and external with auto-URL detection
  - Toolbars: InlineToolbar (floating, on selection), FixedToolbar (persistent, top)
  - Advanced: BlocksFeature (embed Payload Blocks in rich text), TextStateFeature (custom inline styles), EXPERIMENTAL_TableFeature, TreeViewFeature (debug)
- **Custom Features:** Build your own Lexical features and share with the community.
- **Converters:** Built-in converters for HTML, JSX (React), Markdown, and MDX. Custom converters for Blocks/Inline Blocks.
- **Inline Blocks:** Components rendered inline within text (e.g., dynamic price display from API).
- **Markdown Support:** Many features support markdown shortcuts (e.g., `**bold**`, `# heading`, `- list`).
- **Keyboard Shortcuts:** Standard shortcuts (Ctrl/Cmd+B for bold, etc.).

### Next.js App Router Integration

- Payload installs directly into your `/app` directory.
- REST API and GraphQL endpoints are Next.js route handlers.
- Admin panel is a Next.js route group.
- Local API accessible in any server component or server function.
- Share layouts, middleware, and routing between CMS and frontend.
- Single deployment for both CMS and website.

### TypeScript Type Generation

- `payload generate:types` creates `payload-types.ts` from your config.
- Auto-generates interfaces for all collections, globals, and their fields.
- Relationships typed to reference correct collection interfaces.
- Declare statement auto-added for type inference within Payload operations.
- JSON Schema to TypeScript conversion under the hood.

### Live Preview

- Renders frontend in an iframe within the admin panel.
- Real-time preview as content is edited.
- Server-side mode (recommended): `router.refresh()` on save, re-renders with fresh Local API data.
- Client-side mode: `window.postMessage` events on every change for instant preview.
- Works with any frontend framework in the iframe.

### Jobs Queue

- Built-in job queue system for background processing.
- **Tasks:** Define reusable task handlers in config.
- **Workflows:** Chain tasks in order with retry from point of failure.
- **Queues:** Segment jobs into named groups.
- **Scheduling:** Cron-style `schedule` attributes for recurring tasks.
- **Delayed Execution:** `waitUntil` for future-dated jobs.
- **Auto-Run:** Configure automatic queue processing on a cron schedule.
- **Serverless Support:** Vercel Cron can trigger job processing via API endpoint.
- **Use Cases:** Scheduled publishing, email sending, data sync, media processing.

### Multi-Tenancy

- First-class multi-tenancy support for SaaS applications.
- Tenant isolation at the data level.

---

## Sources

- [Payload CMS Official Documentation](https://payloadcms.com/docs/getting-started/what-is-payload)
- [Payload 3.0 Announcement](https://payloadcms.com/posts/blog/payload-30-the-first-cms-that-installs-directly-into-any-nextjs-app)
- [Fields Overview](https://payloadcms.com/docs/fields/overview)
- [Database Overview](https://payloadcms.com/docs/database/overview)
- [PostgreSQL Adapter](https://payloadcms.com/docs/database/postgres)
- [SQLite Adapter](https://payloadcms.com/docs/database/sqlite)
- [Queries Overview](https://payloadcms.com/docs/queries/overview)
- [Versions](https://payloadcms.com/docs/versions/overview)
- [Drafts](https://payloadcms.com/docs/versions/drafts)
- [Autosave](https://payloadcms.com/docs/versions/autosave)
- [Localization](https://payloadcms.com/docs/configuration/localization)
- [i18n](https://payloadcms.com/docs/configuration/i18n)
- [Uploads](https://payloadcms.com/docs/upload/overview)
- [Storage Adapters](https://payloadcms.com/docs/upload/storage-adapters)
- [Authentication Overview](https://payloadcms.com/docs/authentication/overview)
- [Access Control Overview](https://payloadcms.com/docs/access-control/overview)
- [Collection Access Control](https://payloadcms.com/docs/access-control/collections)
- [Field-level Access Control](https://payloadcms.com/docs/access-control/fields)
- [Admin Panel](https://payloadcms.com/docs/admin/overview)
- [Custom Components](https://payloadcms.com/docs/admin/components)
- [Custom Views](https://payloadcms.com/docs/custom-components/custom-views)
- [Root Components](https://payloadcms.com/docs/custom-components/root-components)
- [Customizing CSS](https://payloadcms.com/docs/admin/customizing-css)
- [Plugins Overview](https://payloadcms.com/docs/plugins/overview)
- [Collection Hooks](https://payloadcms.com/docs/hooks/collections)
- [Field Hooks](https://payloadcms.com/docs/hooks/fields)
- [Hooks Context](https://payloadcms.com/docs/hooks/context)
- [Live Preview](https://payloadcms.com/docs/live-preview/overview)
- [Server-side Live Preview](https://payloadcms.com/docs/live-preview/server)
- [Rich Text Editor](https://payloadcms.com/docs/rich-text/overview)
- [Official Rich Text Features](https://payloadcms.com/docs/rich-text/official-features)
- [Lexical Converters](https://payloadcms.com/docs/rich-text/converters)
- [REST API](https://payloadcms.com/docs/rest-api/overview)
- [Local API](https://payloadcms.com/docs/local-api/overview)
- [GraphQL Overview](https://payloadcms.com/docs/graphql/overview)
- [Select](https://payloadcms.com/docs/queries/select)
- [Depth](https://payloadcms.com/docs/queries/depth)
- [Pagination](https://payloadcms.com/docs/queries/pagination)
- [Email Functionality](https://payloadcms.com/docs/email/overview)
- [Migrations](https://payloadcms.com/docs/database/migrations)
- [Generating TypeScript](https://payloadcms.com/docs/typescript/generating-types)
- [The Payload Config](https://payloadcms.com/docs/configuration/overview)
- [Production Deployment](https://payloadcms.com/docs/production/deployment)
- [Jobs Queue](https://payloadcms.com/docs/jobs-queue/overview)
- [Trash](https://payloadcms.com/docs/trash/overview)
- [Stripe Plugin](https://payloadcms.com/docs/plugins/stripe)
- [Join Field](https://payloadcms.com/docs/fields/join)
- [Point Field](https://payloadcms.com/docs/fields/point)
- [Payload Cloud](https://payloadcms.com/cloud-terms)
- [Payload is Open Source (MIT)](https://payloadcms.com/posts/blog/open-source)
- [GitHub Repository](https://github.com/payloadcms/payload)
- [Trash Support & Job Scheduling Release](https://payloadcms.com/posts/releases/new-in-payload-trash-support-job-scheduling-and-dx-enhancements)
- [Enterprise Audit Logs](https://payloadcms.com/enterprise/audit-logs)
