# Strapi v5 Features

## Runtime Architecture
- **Language:** JavaScript/TypeScript. The v5 codebase has been almost entirely rewritten in TypeScript (TypeScript 5.0+ support).
- **Framework:** Built on **Koa.js** (upgraded from Koa v2 to v3 in Strapi 5). Koa handles the HTTP request/response pipeline with middleware cascading (LIFO execution order).
- **Runtime:** **Node.js 18+** required. Only maintenance and LTS versions of Node.js are supported.
- **Process Model:** Single long-running Node.js process. Strapi is designed to run as an always-on service. It is explicitly **not suited for serverless** environments due to slow cold boot times (several seconds to initialize).
- **Admin Panel Bundler:** Vite (default in v5, replacing Webpack from v4). Webpack is still available as an alternative bundler.
- **Dev Mode:** `strapi develop` runs with hot-reload. The admin panel watches for changes by default in v5 (no separate `watch-admin` command needed).
- **Docker:** Supported via the community `@strapi-community/dockerize` CLI package that generates `Dockerfile` and `docker-compose.yml`. No official first-party Docker image.

## Database Support
- **Supported databases:** PostgreSQL, MySQL/MariaDB, SQLite
- **Default:** SQLite (for local development)
- **NOT supported:** MongoDB, any NoSQL database, cloud-native databases (Aurora, Cloud SQL, etc.)
- **Query layer:** **Knex.js** as the SQL query builder. Bookshelf.js historically provided ORM capabilities on top of Knex. Direct Knex access is available via `strapi.db.connection`.
- **Configuration:** `config/database.js` (or `.ts`). Accepts Knex.js connection options including connection pooling settings.
- **Transactions:** Supported. `strapi.db.transaction()` wraps operations; nested transactions use the outer transaction implicitly. `onCommit` and `onRollback` hooks are available. Knex queries require explicit `.transacting(trx)`.
- **Migrations:** Experimental. JavaScript files in `./database/migrations/` run automatically at startup in alphabetical order. The `up()` function runs in a database transaction. **No down migrations** -- reverting requires manual intervention.

## Content Management
- **Content types:** Two kinds -- **Collection types** (multiple entries) and **Single types** (one entry per type).
- **CRUD:** Full create, read, update, delete via admin panel, REST API, GraphQL, and the Document Service API (programmatic).
- **Components:** Reusable field groups that can be added to content types, dynamic zones, or nested in other components. Defined as `type: 'component'` with a `repeatable` parameter.
- **Dynamic Zones:** Flexible content areas where editors choose from a list of allowed components at authoring time. Defined as `type: 'dynamiczone'` with a `components` array. Different components cannot share the same field name with different types.
- **Relations:** One-way, one-to-one, one-to-many, many-to-one, many-to-many, and many-way relations between collection types.
- **Ordering:** Entries can be sorted in list views. Repeatable components maintain order. No built-in tree structure (parent/child/sibling pointers) for content -- content is flat by type.
- **Tree structures:** No native tree/hierarchy support for content entries. Hierarchical content requires custom implementation via self-referencing relations.
- **Bulk operations:** Bulk publish and bulk unpublish from the Content Manager list view (select multiple entries).
- **Conditional fields:** Added in v5.17. Fields can be shown/hidden dynamically based on dropdown selections, checkbox values, or other field state.

## Content Schema / Content Modeling
- **Definition approach:** Dual -- content types can be defined via the **admin panel Content-Type Builder** (generates JSON schema files) or by editing **schema JSON files** directly in code (`./src/api/[api-name]/content-types/[content-type-name]/schema.json`).
- **Strapi AI:** Built-in AI assistant for content modeling that can generate content type structures.
- **Field types available:**
  - **Text** (short text, single line)
  - **Rich Text (Blocks)** -- block-based editor with live rendering, images, code blocks
  - **Number** (integer, float, decimal, big integer)
  - **Boolean** (toggle)
  - **Date** (date, time, or datetime picker)
  - **Email** (with format validation)
  - **Password** (hashed)
  - **UID** (auto-generated unique identifier, optionally based on another field)
  - **Enumeration** (predefined dropdown list)
  - **JSON** (arbitrary JSON data)
  - **Media** (single or multiple file references)
  - **Relation** (link to another content type)
  - **Component** (embed a component)
  - **Dynamic Zone** (flexible component selection)
  - **Custom fields** (plugin-defined; cannot create relation, media, component, or dynamic zone custom field types)
- **Validations:** `required`, `unique`, `minLength`, `maxLength`, `min`, `max`, `regex` pattern matching, private (hidden from API). Database-level constraints via `column` config (`unique: true`, `notNullable`).
- **Note on unique + drafts:** Unique validation is intentionally skipped for draft entries when Draft & Publish is enabled; duplicates are only caught at publish time.

## Content Delivery / API
- **REST API:** Built-in, enabled by default. Auto-generated CRUD endpoints per content type at `/api/:pluralApiId`.
- **GraphQL API:** Available via the `@strapi/plugin-graphql` package (not built-in, must be installed). Powered by **Apollo Server 4** with **Nexus** for code-first schema generation. Auto-generates queries, mutations, and subscriptions per content type.
- **Filtering operators (REST):** `$eq`, `$eqi`, `$ne`, `$nei`, `$in`, `$notIn`, `$lt`, `$lte`, `$gt`, `$gte`, `$between`, `$contains`, `$notContains`, `$containsi`, `$notContainsi`, `$startsWith`, `$startsWithi`, `$endsWith`, `$endsWithi`, `$null`, `$notNull`. Logical: `$and`, `$or`, `$not`.
- **Deep filtering:** Filter on nested relation fields (e.g., `filters[chef][restaurants][stars][$eq]=5`). Performance warning for deeply nested filters.
- **Sorting:** `sort` parameter with `:asc` or `:desc` per field. Multiple sort fields supported.
- **Pagination:** Two modes (cannot mix): Page-based (`page` + `pageSize`) or offset-based (`start` + `limit`).
- **Population:** `populate` parameter to include relations, media, components, dynamic zones. Supports nested population, selective field population, and combining with filters/sort.
- **Field selection:** `fields` parameter to return only specific fields.
- **Locale parameter:** Top-level `locale` query parameter (simplified from v4's `plugins[i18n][locale]`).
- **Status parameter:** Filter by `draft` or `published` status.
- **Document Service API:** Internal programmatic API (replacement for v4 Entity Service for high-level operations). Handles documents with locale/status variants. Middleware can be attached for before/after hooks.
- **Query Engine API:** Lower-level database query API used under the hood.

## Content Publishing
- **Draft & Publish:** Built-in (core feature, free). Content exists in two states: Draft and Published. A third status "Modified" indicates a published entry with newer draft changes.
- **Published tab is read-only** -- serves as a snapshot of what's currently live.
- **Scheduled publishing:** Not directly on individual entries. Instead, entries are grouped into **Releases** which can be scheduled for a specific date, time, and timezone. Releases feature is available on Growth and Enterprise plans.
- **Releases:** Containers that group entries (across different content types and locales) for simultaneous publish/unpublish. Can be published manually or scheduled.
- **Bulk publish/unpublish:** Select multiple entries from list view and publish or unpublish simultaneously.
- **Review Workflows:** **Enterprise only.** Custom multi-stage approval pipelines (default stages: To do, In progress, Ready to review, Reviewed). Stages are configurable, reorderable, and deletable. Role-based stage transition controls. Entries can be assigned to specific admin users for review.

## Content Versioning
- **Content History:** Available on **Growth and Enterprise plans** (not free Community). Stores previous document versions with timestamps, user attribution, and status labels (Draft, Modified, Published).
- **Restore:** Restoring a version overwrites the current draft with that version's content, setting the document to "Modified" status.
- **Comparison:** The history view shows fields and content for each selected version. Unknown or renamed fields from previous versions appear under an "Unknown fields" section.
- **Diff:** Field-level comparison showing what changed between versions. No side-by-side diff UI documented -- it is a version browser showing one version at a time.
- **Retention:** Self-hosted Enterprise customers can configure retention duration (default 90 days). No documented retention configuration for Growth plan.
- **No built-in branching or named versions.** Content History is a linear timeline of changes.

## Internationalization (i18n)
- **Status:** Built into Strapi 5 core (was a separate plugin in v4). No plugin installation required.
- **Approach:** **Row per locale.** Each locale version of a document is stored as a separate database row. The `documentId` (24-char alphanumeric string) groups all locale/status variants of a single document.
- **Scaling:** Adding more locales requires more rows, not schema changes.
- **Per-field translation:** All fields in a content type are translatable when i18n is enabled for that type. Dynamic zones and components can differ structurally per locale (different components, different ordering).
- **Locale management:** Configured in admin panel Settings. Default locale is `en`.
- **Fallback:** The `en` locale serves as the fallback -- if a translation is not found, `en` content is used.
- **API usage:** `?locale=fr` as a top-level query parameter (REST) or `locale` argument (GraphQL).
- **Content Manager:** Locale switcher in the editor to create/view translations per locale.
- **Limitations:** No per-field locale opt-out documented (it's all-or-nothing for a content type once i18n is enabled on that type). No configurable fallback chains beyond the default locale.

## Media Management
- **Media Library:** Built-in centralized asset manager. Supports images, audio, video, and documents.
- **Upload providers:**
  - **Local** (default): Files stored in `public/uploads/`
  - **AWS S3:** Official `@strapi/provider-upload-aws-s3`
  - **Cloudinary:** Official `@strapi/provider-upload-cloudinary`
  - Custom providers can be created
- **Private providers:** Providers can implement `isPrivate()` and `getSignedUrl(file)` for secure signed URL access.
- **Responsive images:** Automatic size optimization without quality loss. Generates multiple responsive formats: small, medium, and large thumbnails by default. Configurable in settings.
- **Folder organization:** Full folder system -- create, edit, move, delete folders. Assets can be browsed, filtered, sorted, and searched within folders.
- **Metadata:** File name, alternative text, caption, and location fields per asset.
- **View modes:** Grid view and list view toggle.
- **Migration:** Assets can be migrated between providers (local to S3, S3 to Cloudinary, etc.).

## Authentication
- **Admin panel authentication:** JWT-based. Separate from end-user auth.
- **End-user authentication (Users & Permissions plugin):**
  - **Email/password** (default, always enabled)
  - **JWT tokens** for stateless auth. Two JWT management modes configurable in `config/plugins`.
  - **Social/OAuth 2.0 providers:** Google, Facebook, GitHub, Discord, Twitch, Instagram, Auth0, AWS Cognito, LinkedIn, Patreon, Reddit, Twitter, Keycloak. Custom providers can be added manually.
- **API Tokens:** Built into Strapi CMS. Scoped tokens (read-only, full-access, or custom). Managed in admin panel. Used for REST and GraphQL authentication without user credentials.
- **SSO (Single Sign-On):** **Enterprise only** (or SSO add-on). Supports SAML, OAuth 2.0, OpenID Connect. Providers: Azure AD, Okta, Auth0, Keycloak, and others. Configuration in `config/admin`.
- **Security practices:** Token expiration recommended at 15 minutes, refresh token rotation, read-only tokens for public access.

## Authorization / RBAC
- **Admin panel RBAC:**
  - Default roles: **Super Admin**, **Editor**, **Author**
  - Custom roles: Unlimited custom roles available (was Enterprise-only pre-v4.8, now free)
  - **Permission granularity:** Per content type, per CRUD operation (Create, Read, Update, Delete, Publish), per locale, per plugin
  - **Field-level permissions:** Toggle permissions per field per content type per role. Untick specific fields to deny access.
  - **Custom conditions:** Per-permission conditions can be defined to add contextual restrictions (e.g., "only own entries").
  - **Programmatic conditions:** Custom RBAC conditions can be registered in code (`config/admin`).
- **End-user permissions (Users & Permissions plugin):**
  - **Public** and **Authenticated** roles by default
  - Custom roles for end users
  - Per-endpoint permission toggles
- **Content-level permissions:** Through custom conditions (e.g., restrict to entries created by the user). No native row-level security out of the box beyond conditions.

## Admin Panel
- **Tech stack:** React single-page application. Vite bundler (default; Webpack available). TypeScript.
- **Customization:**
  - `src/admin/app.js|ts` for configuration (logo, theme, locales, menu)
  - Homepage customization (configurable dashboard)
  - Injection zones for plugins to add components
  - Custom admin panel pages and routes
  - Logo, favicon, and color theme overrides
- **Plugin Marketplace:** Built into admin panel. Browse, search, and install plugins and providers. v4 and v5 plugins are **not cross-compatible** (providers are compatible across both).
- **Content Manager features:** List view with filters/sort/search, bulk actions, locale switcher, Draft/Published tabs, relation management, component/dynamic zone editors, media picker.
- **Live Preview:** Built-in content preview with real-time updates as you type (added 2025). Requires frontend integration configuration.
- **Responsive admin panel:** Added in 2025 for mobile/tablet use.
- **Locales:** Admin panel UI can be translated. Locales/translations configurable for the admin interface itself.

## Plugin / Extension System
- **Plugin architecture:** Plugins have both a server-side (Koa routes, controllers, services, content types) and admin-side (React components, pages, settings).
- **Admin Panel API hooks:** `register()`, `bootstrap()`, `registerTrads()` for injecting React components and translations.
- **Admin APIs:** Menu API, Settings API, Injection Zone API, Reducer API, Hook API for extending navigation, settings panels, and UI.
- **Server APIs:** Routes, controllers, services, policies, middlewares per plugin.
- **Document Service Middleware:** The **recommended** approach in v5 for extending behavior (replacing lifecycle hooks for most use cases). Middleware wraps Document Service methods with before/after logic. Can affect multiple content types.
- **Lifecycle hooks:** Still exist but are now intended for low-level database activity hooks only. Events: `beforeCreate`, `afterCreate`, `beforeUpdate`, `afterUpdate`, `beforeDelete`, `afterDelete`, `beforeFindOne`, `afterFindOne`, `beforeFindMany`, `afterFindMany`, `beforeCount`, `afterCount`.
- **Custom fields:** Registered in the `register()` lifecycle function via `strapi.customFields.register()`. Cannot create relation, media, component, or dynamic zone custom field types.
- **Custom APIs:** Add custom routes, controllers, services in `./src/api/[api-name]/`.
- **Marketplace:** In-admin browser for discovering and installing community and official plugins.

## Webhooks
- **Configuration:** Via admin panel (Settings > Webhooks) or programmatically. Default headers configurable in `config/server`.
- **Events:**
  - **Entry:** `entry.create`, `entry.update`, `entry.delete`, `entry.publish`, `entry.unpublish`
  - **Media:** Upload events
- **Payload:** JSON with `event`, `createdAt`, `model`, and `entry` fields. In v5, media fields in entry webhooks contain only the Asset ID (not full media records).
- **Signing:** Not built-in. Recommended to implement HMAC-SHA256 signing manually with a shared secret, timestamp, and signature headers.
- **Retries:** Not built-in. Documentation recommends implementing retry logic on the receiving end or using external services like QStash for guaranteed delivery.
- **Delivery logs:** No built-in delivery log or retry dashboard. External tooling needed.
- **Custom webhook events:** Can be registered programmatically.

## SDKs & Client Libraries
- **Official:** `@strapi/client` -- JavaScript/TypeScript client library. Supports API token authentication and type-safe content queries. Import: `import { strapi } from '@strapi/client'`.
- **No official SDKs for:** Python, Go, Swift, PHP, Ruby, Java, or any other language.
- **Community libraries:** `strapi-sdk-js` (community JS SDK), various unofficial wrappers exist but none are officially maintained.
- **Type generation:** `strapi ts:generate-types` generates TypeScript types from your content type schemas.

## CLI
- **Project creation:** `npx create-strapi-app@latest` (the `strapi new` command was removed in v5).
- **Development:** `strapi develop` (dev mode with hot-reload), `strapi build` (build admin panel), `strapi start` (production, no hot-reload).
- **Code generation:** `strapi generate` -- interactive generator for content types, controllers, policies, middlewares, services, and plugins.
- **TypeScript types:** `strapi ts:generate-types` (`--debug`, `--silent`, `--out-dir` options).
- **Inspection:** `strapi routes:list`, `strapi content-types:list`, `strapi policies:list`, `strapi middlewares:list`.
- **Data management:** `strapi export`, `strapi import`, `strapi transfer`.
- **Other:** `strapi version`, `strapi help`, `strapi console` (REPL).
- **Cloud CLI:** Separate `strapi cloud` commands for Strapi Cloud deployment (login, deploy, logs, etc.).
- **Removed in v5:** `strapi install`, `strapi uninstall`, `strapi new`, `strapi watch-admin`.

## Deployment
- **Self-hosted:** Runs on any server with Node.js 18+. Standard deployment to VPS, bare metal, PaaS.
- **Strapi Cloud:** Managed hosting with bundled database, asset storage (CDN), and email. Plans: Free, Essential, Pro, Scale.
- **Docker:** Supported via community tooling (`@strapi-community/dockerize`). Not an official first-party Docker image.
- **Serverless:** **Not recommended.** Strapi's boot time (several seconds) makes cold starts impractical. Every request would take seconds, not milliseconds.
- **Hosting platforms documented:** DigitalOcean, Render, Heroku, AWS (EC2, not Lambda), Azure, Google Cloud Run, Railway.
- **Requirements:** Node.js 18+, supported database, persistent filesystem for SQLite/local uploads.

## Email
- **Built-in:** Email feature for transactional messages. Configured in `config/plugins.js|ts`.
- **Default provider:** Local SMTP (Sendmail).
- **Official providers:**
  - `@strapi/provider-email-sendgrid`
  - `@strapi/provider-email-mailgun`
  - `@strapi/provider-email-amazon-ses`
  - `@strapi/provider-email-nodemailer`
- **Configuration:** Provider name, provider options (API keys from env vars), settings (defaultFrom, defaultReplyTo, testAddress).
- **Strapi Cloud:** Has a built-in email service included in Cloud hosting.
- **Limitations:** Transactional email only (password resets, confirmations, etc.). No marketing email, templates, or campaign management built-in.

## Backup & Restore
- **Export:** `strapi export` CLI command. Exports content (entities + relations), files (media assets), project configuration, and schemas. Output is a `.tar` archive with optional `.gz` compression and `.enc` encryption. `--only` flag to export selectively (content, files, config).
- **Import:** `strapi import` CLI command. **Destructive** -- deletes all existing data (database + uploads) before importing. Source and target schemas must match exactly. Supports `.tar`, `.tar.gz`, `.tar.gz.enc`.
- **Transfer:** `strapi transfer` for direct instance-to-instance data transfer. Requires `transfer.token.salt` in target's `config/admin`. Transfer tokens authorize the operation.
- **No admin panel UI** for backup/restore. CLI-only.
- **No incremental backups.** Full export/import only.
- **No automated backup scheduling** built-in (Strapi Cloud has backup features).

## Configuration System
- **File-based:** Configuration files in `./config/` directory. JavaScript or TypeScript files.
  - `config/server.js|ts` -- host, port, app keys, logging
  - `config/database.js|ts` -- database connection
  - `config/admin.js|ts` -- admin panel, JWT secrets, SSO, transfer tokens
  - `config/plugins.js|ts` -- plugin configuration
  - `config/middlewares.js|ts` -- middleware stack
  - `config/api.js|ts` -- API settings
- **Environment variables:** `.env` file at project root. Accessed via `process.env.{VAR}` or the `env()` utility (with default values and type casting).
- **Environment-specific overrides:** `./config/env/{NODE_ENV}/` directory. Files merge into base config. `NODE_ENV` defaults to `development`.
- **Key env vars:** `HOST`, `PORT`, `APP_KEYS`, `API_TOKEN_SALT`, `ADMIN_JWT_SECRET`, `TRANSFER_TOKEN_SALT`, `ENCRYPTION_KEY`, `DATABASE_CLIENT`, `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_NAME`, `DATABASE_USERNAME`, `DATABASE_PASSWORD`.
- **No admin panel configuration editor** for server/database config. Some plugin settings are configurable via admin UI.
- **Custom env path:** Set `ENV_PATH` environment variable to change `.env` file location.

## Observability / Logging
- **Logging library:** **Winston.** Configurable in `config/server.js|ts`.
- **Log levels:** `error`, `warn`, `info`, `http`, `verbose`, `debug`, `silly`. Default level in v5 is `http` (changed from `info` in v4).
- **Transports:** Console (default), file-based logging, custom transports (e.g., `winston-daily-rotate-file` for rotation).
- **Monitoring integrations:** No built-in monitoring dashboard. Integrations documented for Prometheus, Grafana, Datadog, Amazon CloudWatch, Splunk, ELK Stack, Sentry.
- **Request logging:** Koa middleware logs HTTP requests at `http` level.
- **No built-in APM, tracing, or metrics endpoint.**

## Audit Trail
- **Audit Logs:** **Enterprise only.** Captures every administrative action in a searchable, filterable history.
- **Tracked information:** Action type (create, update, delete, etc.), date/time, user who performed the action, detailed modal with IP address, request body, response body.
- **Filtering:** By action, date, user.
- **Retention:** Configurable for self-hosted Enterprise (default 90 days).
- **Shareable:** Individual log entries can be shared via link with team members.
- **No audit logging in Community or Growth plans.** Community users must build custom audit logging via lifecycle hooks or Document Service middleware.

## Testing
- **No built-in test runner or test utilities shipped with Strapi.**
- **Recommended stack:** Jest (test runner, assertions) + Supertest (HTTP endpoint testing).
- **Unit testing:** Mock the `strapi` object for isolated controller/service testing.
- **Integration/E2E testing:** Boot a Strapi instance, use Supertest to hit REST endpoints.
- **Plugin testing:** Jest + Supertest installed as dev dependencies in the plugin directory.
- **No test database management, fixtures, or factory utilities built-in.**

## Transfer / Migration
- **strapi transfer:** Direct data transfer between two running Strapi instances. Uses transfer tokens for authorization. Transfers content, files, configuration, and schemas.
- **strapi export / strapi import:** Tar-based archive for offline transfer. Import is destructive (wipes target first). Schemas must match between source and target.
- **v4 to v5 migration:** Official migration guide and tooling. Codemods available. Breaking changes documented (50+ documented breaking changes).
- **Database migrations:** `./database/migrations/` with `up()` functions. No `down()` migrations. Experimental feature.
- **Content type changes:** Strapi auto-syncs database tables with content type schemas at boot. No manual migration needed for schema changes made through the Content-Type Builder.
- **No environment promotion workflow** (no staging-to-production content sync tool beyond manual export/import/transfer).

## Pricing / Licensing
- **License:** MIT license for the Community Edition (open source).
- **Open Core model:** Community Edition is free and open source. Enterprise features require paid plans.

**Self-Hosted Plans:**

| Plan | Price | Key Features |
|------|-------|--------------|
| Community | Free | Full CMS, REST + GraphQL API, RBAC, i18n, Draft & Publish |
| Growth | $15/month per admin seat | Releases, Content History, Basic Support |
| Enterprise | Custom pricing (per admin seat) | Everything in Growth + SSO, Audit Logs, Review Workflows, SLA support |

**Strapi Cloud Plans:**

| Plan | Price |
|------|-------|
| Free | $0 (limited resources) |
| Essential | Starting price varies |
| Pro | Higher tier |
| Scale | Highest tier |

**Enterprise-gated features (not available in free Community):**
- SSO (Enterprise or SSO add-on)
- Audit Logs (Enterprise only)
- Review Workflows (Enterprise only)
- Content History (Growth or Enterprise)
- Releases / Scheduled Publishing (Growth or Enterprise)

**Free in Community:**
- Full REST + GraphQL APIs
- Draft & Publish
- RBAC with custom roles
- i18n
- Media Library
- Plugin system
- Users & Permissions with social auth
- Webhooks
- CLI data management (export/import/transfer)
