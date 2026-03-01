# ModulaCMS vs Competitors: Full Comparison

**Updated:** 2026-02-28

Honest, dimension-by-dimension comparison of ModulaCMS against Strapi v5, Contentful, Payload CMS 3.x, Sanity, Directus v11, WordPress 6.x, Wagtail 7.3, and DatoCMS. Each section rates ModulaCMS as **leads**, **on par**, or **behind**.

---

## 1. Architecture & Deployment

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Language | Go | Node.js/TS | N/A (SaaS) | TypeScript | N/A (SaaS studio) | Node.js/TS | PHP | Python/Django | N/A (SaaS) |
| Self-hosted | Yes | Yes | No | Yes | Studio only | Yes | Yes | Yes | No |
| Single binary | Yes | No | N/A | No | N/A | No | No | No | N/A |
| Built-in HTTPS | Let's Encrypt autocert | No | N/A | No | N/A | No | No (plugin/proxy) | No (proxy) | N/A |
| Serverless | No | No | N/A | Yes (Next.js) | N/A | No | No | No | N/A |

**ModulaCMS leads.** No competitor offers a single binary with three concurrent servers (HTTP, HTTPS with autocert, SSH). Zero-dependency deployment is unique. Go's compiled binary eliminates the Node.js/Python/PHP runtime dependency chain entirely. The built-in Let's Encrypt autocert means production TLS without a reverse proxy.

**Trade-off:** Payload CMS has an edge in the Next.js ecosystem with its App Router integration and serverless compatibility. Contentful, Sanity, and DatoCMS remove deployment burden entirely as managed SaaS.

---

## 2. Database Support

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| SQLite | Yes | Yes | N/A | Yes | N/A | Yes | Experimental | Dev only | N/A |
| MySQL | Yes | Yes | N/A | No | N/A | Yes | Yes (only) | Yes | N/A |
| PostgreSQL | Yes | Yes | N/A | Yes | N/A | Yes | No | Yes | N/A |
| MongoDB | No | No | N/A | Yes | N/A | No | No | No | N/A |
| MS SQL | No | No | N/A | No | N/A | Yes | No | No | N/A |
| OracleDB | No | No | N/A | No | N/A | Yes | No | No | N/A |
| Switchable per deploy | Yes | No | N/A | Yes | N/A | Yes | No | No | N/A |
| Existing DB wrapping | No | No | N/A | No | N/A | Yes | No | No | N/A |

**ModulaCMS leads** among self-hosted options for the tri-database pattern -- SQLite, MySQL, and PostgreSQL from a single codebase with type-safe sqlc-generated queries per backend. Only Directus matches this breadth (7 databases), but Directus uses runtime introspection over Knex.js rather than compile-time type safety.

**Behind Directus** in total database count (3 vs 7) and in database introspection (Directus can wrap any existing SQL database without migration; ModulaCMS requires its own schema).

---

## 3. Content Modeling

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Field types | 15 predefined + unlimited custom (DB-driven registry) | 15+ (fixed set + custom fields via plugins) | 13 (fixed) | 23 (fixed set) | 18+ (fixed set + custom input components) | 30+ interfaces over fixed DB types | Untyped key-value (ACF/Meta Box add 30-40 typed fields) | 20+ block types (extensible via subclassing) | 21 (fixed set + plugin editors) |
| Field type extensibility | Full CRUD on `field_types` table; open `FieldType` string type accepts any value; admin UI for management | Custom fields via plugin SDK (cannot be relation/media/component/dynamic zone) | Custom editors via App Framework (field type set is fixed) | Custom field types not supported (fixed set) | Custom input components replace UI only; schema types are fixed | Custom interfaces (UI) over fixed DB column types | Plugins define new meta field patterns (not DB-level types) | Custom blocks by subclassing `Block` base class | Plugin field editors replace UI; underlying types fixed |
| Composable blocks | Yes (`<block-editor>` web component: datatype-backed, nestable, drag-and-drop, sibling-pointer tree) | Yes (dynamic zones) | No (refs only) | Yes (blocks, arrays) | Yes (objects, arrays) | Yes (repeater, M2A) | Yes (Gutenberg blocks) | Yes (StreamField) | Yes (modular content) |
| Code-first | No (API/UI) | Dual (UI + JSON) | UI only | Yes (TypeScript) | Yes (JS/TS) | UI only | Code (register_post_type) | Yes (Python) | UI only |
| Tree structures | Native (sibling pointers) | No | No | Plugin (nested-docs) | No | No | No | Native (treebeard) | Native (parent-child) |
| Conditional fields | No | Yes (v5.17) | No | Yes | Yes | Yes | No | No | No |

**ModulaCMS leads** in field type extensibility. The `field_types` table is a first-class data-driven registry with full CRUD (create, read, update, delete via API and admin UI). The Go `FieldType` is an open `string` type -- not a closed enum -- so any custom value is accepted at the type system level. The 15 predefined types (`text`, `textarea`, `number`, `date`, `datetime`, `boolean`, `select`, `media`, `relation`, `json`, `richtext`, `slug`, `email`, `url`, `content_tree_ref`) are bootstrap seed data, not hard limits. Users and plugins can register unlimited additional types. Most competitors have fixed type sets where "custom fields" means replacing the UI editor widget but not defining genuinely new field types at the data layer.

**ModulaCMS leads** in tree structures. The sibling-pointer tree (parent_id, first_child_id, next_sibling_id, prev_sibling_id) provides O(1) navigation and reordering -- a first-class primitive, not a plugin. Only Wagtail (treebeard materialised path) and DatoCMS (parent-child) have comparable native trees.

**ModulaCMS on par** in composable content blocks. The `<block-editor>` web component provides a full block composition system built on the same sibling-pointer tree architecture as the content tree. Blocks are backed by datatypes (the schema system), so editors insert blocks from any registered datatype -- each block carries its own `datatypeId`, `fields[]`, and metadata. Blocks are nestable (container blocks accept children, max depth 8), reorderable via drag-and-drop and keyboard shortcuts (Tab/Shift+Tab for indent/outdent, arrows, Enter, Delete, Ctrl+Shift+D for duplicate), and support per-block field data. The type system is open: 4 built-in rendering types (text, heading, image, container) plus automatic fallback for any database-defined type. This is architecturally comparable to Wagtail's StreamField and Strapi's Dynamic Zones, with the distinction that ModulaCMS's blocks use the same pointer-based tree primitive as the content tree itself rather than a separate JSON blob format.

**Behind** in conditional fields. Strapi, Payload, Sanity, and Directus all support showing/hiding fields based on other field values. ModulaCMS does not.

---

## 4. Content Delivery API

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| REST | Yes | Yes | Yes | Yes | No (GROQ/GraphQL) | Yes | Yes | Yes (DRF) | Yes (CMA only) |
| GraphQL | No (by design) | Plugin | Yes | Yes (auto) | Yes | Yes (auto) | Plugin (WPGraphQL) | Plugin (grapple) | Yes (CDA) |
| Output format transform | 6 formats | No | No | No | No | Export (CSV/XML/YAML) | No | No | No |
| Filtering operators | 7 operators, in-memory, type-aware (number/boolean/date/string), zero SQL injection surface | 20+ (SQL-built) | 12+ (SQL-built) | 15+ (SQL-built) | GROQ (unlimited) | 30+ (SQL-built) | Limited | Basic | 10+ |
| Query endpoint | `GET /api/v1/query/{datatype}` with filter/sort/pagination params | Per-content-type REST endpoints | Per-content-type REST + GraphQL | Per-collection REST + GraphQL | GROQ + GraphQL | Per-collection REST + GraphQL | `/wp/v2/{type}` | `/api/v2/pages/` | GraphQL CDA |
| Sync/delta API | No | No | Yes | No | Listener API | WebSocket | No | No | Yes (SSE) |
| Real-time | No | No | No | No | Yes (SSE) | Yes (WebSocket) | No | No | Yes (SSE) |

**ModulaCMS leads** in output format transformation. Serving content in Contentful, Sanity, Strapi, WordPress, clean, or raw format from a single API is unique. This allows frontend teams to switch CMS providers or consume content in their preferred format without backend changes. No other CMS offers this.

**ModulaCMS on par to leading** in API filtering. The in-memory query package processes all filters, sorts, and pagination in Go rather than building dynamic SQL. This eliminates SQL injection surface entirely, avoids maintaining dialect-specific query builders across three databases, and makes the operator set unlimited -- adding a filter operator is a Go function, not a SQL clause tested across SQLite/MySQL/PostgreSQL. Go's performance budget (25K+ req/s, microsecond-level in-memory operations) makes this architecturally viable for CMS-scale datasets.

**GraphQL deliberately excluded.** ModulaCMS is built on three core values: **Performance**, **Flexibility**, and **Transparency**. GraphQL undermines all three. **Performance**: GraphQL adds a query parsing/validation layer, introduces N+1 resolver chains, and uses POST-only requests that bypass HTTP caching -- all for a problem ModulaCMS already solves with output format transformers and in-memory filtering. **Flexibility**: GraphQL locks the API into a rigid schema-first contract where the server must maintain resolver parity with every schema change across all three database backends -- the opposite of the lightweight, data-driven field type extensibility ModulaCMS provides. **Transparency**: GraphQL queries are opaque blobs sent in POST bodies, invisible to standard HTTP logging, caching layers, and debugging tools. REST endpoints with query parameters are inspectable, cacheable, and shareable as URLs. The 6 output format transformers already deliver the "get exactly what you need" value proposition that GraphQL promises, without the overhead.

**Behind** in live preview. ModulaCMS has preview mode (`?preview=true`) but no push mechanism to update an open preview tab when content changes. Competitors offer SSE/WebSocket-based live updates (Sanity, Directus, DatoCMS) primarily for dev-time editor experience -- in production, content changes flow through webhooks → SSG/ISR, not real-time push to end users.

---

## 5. Publishing & Versioning

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Draft/publish | Snapshot-based | Row duplication | Version numbers | Row (_status) | Dual documents | Status field | Post statuses | Draft/live revisions | Draft/published/updated |
| Scheduled publish | Yes (background ticker) | Enterprise (Releases) | Yes (Lite+) | Yes (Jobs Queue) | Yes (Growth+) | No | Yes (WP-Cron) | Yes (cron command) | Yes |
| Version history | Yes (snapshot) | Enterprise | Yes | Yes (autosave) | Yes (plan-gated retention) | Yes (revisions) | Yes (revisions) | Yes (revisions) | Yes |
| Version restore | Yes (field-values only) | Enterprise | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| Optimistic locking | Yes (revision counter + 409) | updatedAt check | sys.version | No | No | No | No | No | No |
| Named/labeled versions | Yes | No | No | No | No | Yes | No | No | No |
| Configurable retention | Yes (per-content cap) | Retention duration | No documented limit | No | Plan-gated (3d/90d/365d) | No | wp_revisions_to_keep | purge_revisions command | No |
| Review workflows | No | Enterprise | Roles only | Access control | No | No (permissions-based) | Plugin (PublishPress) | Yes (multi-stage) | Enterprise |

**ModulaCMS leads** in optimistic locking (revision counter + HTTP 409 is more reliable than timestamp comparison), snapshot-based versioning (avoids tree duplication), labeled versions, and configurable per-content retention caps with smart pruning (never deletes published snapshot or user-labeled versions).

**Behind** in review workflows. ModulaCMS has permission-gated publishing only. Wagtail has the strongest workflow system (multi-stage, custom task types, assignable to page subtrees). Strapi Enterprise and DatoCMS Enterprise offer configurable stage pipelines.

---

## 6. Internationalization

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Approach | Locale column on fields | Row per locale | All locales in JSON | Field-level | Document per locale | Junction table | Plugin (WPML/Polylang) | Separate tree per locale | Per-field toggle |
| Per-field translatable | Yes | No (all-or-nothing per type) | Yes | Yes | Custom | Yes | Plugin-dependent | Yes (TranslatableMixin) | Yes |
| Fallback chains | Yes (2-hop, cycle-validated) | Default locale only | Yes (configurable chains) | Yes | No built-in | Yes | Plugin-dependent | No (manual copy) | No built-in |
| `locale=*` all-locales | Yes | No | Yes | No | No | No | No | No | No |
| Per-locale publishing | Yes (independent snapshots) | Yes | Premium | No | Yes | N/A | Plugin-dependent | No | Yes |
| Non-translatable fields | Single row (locale=''), auto-included | Duplicated per locale | Single field | N/A | Per-document | Duplicated per locale | Plugin-dependent | Per-field toggle | Per-field toggle |

**ModulaCMS leads** in shared-tree i18n (tree structure shared, only field values vary by locale -- avoids the 5x row multiplication of Strapi), non-translatable field handling (single row with `locale=''` auto-included in all snapshots), `locale=*` multi-locale response (only Contentful matches this), and fallback chain validation (2-hop cap with cycle prevention).

**On par** with Contentful and DatoCMS in per-field localization. **Behind** Wagtail's wagtail-localize in translation workflow features (PO file export, machine translation backends, segment-based translation tracking).

---

## 7. Media Management

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Upload + optimization | Yes (auto WebP) | Yes (responsive) | CDN transforms | Yes (imageSizes) | CDN (hotspot/crop) | Yes (sharp, on-the-fly) | Plugin-dependent | Yes (renditions) | Imgix pipeline |
| Focal point crop | Yes | No | Face detection | Yes | Yes (hotspot) | No | No | Yes | Yes |
| On-the-fly transforms | No (preset-based) | No | Yes (Images API) | No (preset-based) | Yes (CDN params) | Yes (URL params) | No | Yes (template tags) | Yes (imgix params) |
| Video management | No | No | No | No | Yes (Mux) | No | Embed only | No | Yes (Mux) |
| Folders | No | Yes | Tags only | No (collections) | No | Yes (virtual) | Yes | Collections | Yes |

**On par** for basic media management -- upload, optimization, focal point, preset dimensions, S3 storage. ModulaCMS's auto-WebP output and configurable dimension presets are solid.

**Behind** in on-the-fly transforms. Directus (sharp URL params), Contentful (Images API), Sanity (CDN params), and DatoCMS (imgix) all support dynamic image transformation via URL parameters without predefining presets. ModulaCMS requires presets defined upfront.

**Behind** in video. Only Sanity and DatoCMS offer video streaming (Mux integration). ModulaCMS has no video management.

---

## 8. Authentication & Authorization

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Auth methods | Password, session, API token, OAuth, SSH keys | JWT, OAuth, API tokens | API keys, PATs, OAuth, SAML | JWT, API keys, any collection | SAML, API tokens | Local, OAuth, OIDC, LDAP, SAML, JWT | Cookie, app passwords | Django auth, SSO | API tokens, SAML |
| Custom roles | Yes | Yes (free) | Premium only | Custom (function-based) | Enterprise | Yes | Yes | Yes (Django groups) | Yes |
| Field-level perms | Yes (dual-layer: `roles` column on fields, NULL = unrestricted) | Yes | Premium | Yes (per-field functions) | Enterprise | Yes | No | No | Enterprise |
| Content-level perms | resource:operation (layer 1) + field-level roles (layer 2) | Custom conditions | Tags-based (Premium) | Yes (query constraints) | GROQ filters (Enterprise) | Yes (item filters) | Meta capabilities | Page tree + collections | Per-creator rules |

**ModulaCMS leads** in authentication breadth. SSH public key auth for TUI access is unique -- no other CMS has it. Password, session cookies, API tokens, OAuth (Google/GitHub/Azure), and SSH keys cover all access patterns.

**On par** in permission granularity with a simpler design. ModulaCMS uses a dual-layer approach: layer 1 is `resource:operation` RBAC (can you touch content at all), layer 2 is a `roles` column on each field -- a nullable JSON array of role IDs. NULL means all roles can edit; a populated array restricts editing to listed roles. This achieves the same field-level permission outcome as Payload (per-field access functions), Directus (per-field toggles in role matrix), and Strapi (per-field unticking) with a single column and a single check, rather than a separate permission matrix or function-based access control layer. Available on the free tier -- Contentful, Sanity, and DatoCMS gate field-level permissions behind Premium/Enterprise plans.

---

## 9. Admin Panel

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Tech | HTMX + templ (server-rendered) | React SPA | React SPA | React + Next.js | React SPA | Vue.js 3 SPA | PHP + Gutenberg (React) | Django + Stimulus/React | Web SPA |
| Customizable | Web components | Injection zones | App Framework | Deep component replacement | Structure builder + plugins | Extensions (Vue) | Hooks + admin pages | Hooks + ViewSets | Plugins (iframe) |
| Live preview | No | Yes (2025) | Yes | Yes (iframe + postMessage) | Yes (Presentation tool) | Yes | Yes (basic) | Yes (PreviewableMixin) | No |
| Real-time collab | No | No | Yes | No | Yes (Google Docs-style) | Yes (v11.15) | No | Concurrent editing (6.2) | No |
| Dark mode | No | No | No | Yes | Yes | Yes | No | Yes | No |
| Dashboards/analytics | No | No | Yes (Analytics) | No | No | Yes (Insights) | Dashboard widgets | Reports | No |

**ModulaCMS leads** in having a server-rendered admin panel (HTMX + templ) -- faster initial load, no JS bundle, works without JS for basic operations. Also unique in having **dual admin interfaces** (web panel + SSH TUI).

**Behind** in admin panel richness. No live preview, no real-time collaboration, no dark mode, no dashboards/analytics. Sanity's real-time collaboration is best-in-class. Contentful and Directus offer the deepest customization frameworks.

---

## 10. SSH TUI

**ModulaCMS is unique.** No other CMS offers a terminal-based management interface accessible over SSH. The Bubbletea TUI provides full content management, schema editing, deploy sync, and configuration -- all from a terminal. This enables headless server management without a browser, CI/CD integration via SSH, and accessibility from any device with an SSH client.

No competitor has anything comparable.

---

## 11. Plugin System

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Approach | Lua (gopher-lua) | Node.js (Koa) | App Framework (iframe) | Config transforms | React plugins | Extensions SDK | PHP (hooks) | Django apps | Plugins SDK (iframe) |
| Sandboxed | Yes (safe stdlib subset) | No | Yes (iframe) | No | No | Optional (Sandbox SDK) | No | No | Yes (iframe) |
| Marketplace | No | Yes | Yes | No | Community | Yes | Yes (59,000+) | No | Yes (100+) |
| Custom DB tables | Yes (prefixed) | Yes | No | Yes (collections) | No (Content Lake) | Yes (collections) | Yes (wp tables) | Yes (Django models) | No |
| Hot reload | Yes (configurable) | Dev mode | N/A | Dev mode | Dev mode | Yes (auto-reload) | No | No | N/A |
| Circuit breaker | Yes | No | N/A | No | N/A | No | No | No | N/A |

**ModulaCMS leads** in plugin isolation. Sandboxed Lua VMs with no filesystem/network access, per-plugin table prefixing, circuit breakers, and hot reload create a production-safe plugin runtime. WordPress's 59,000+ plugins demonstrate the value of ecosystems but with zero sandboxing -- a single bad plugin can crash the site.

**Behind** in ecosystem. No marketplace. WordPress dominates (59,000+ plugins). Strapi, Contentful, Directus, and DatoCMS all have curated marketplaces. ModulaCMS plugins must be built from scratch.

---

## 12. Webhooks

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Built-in | **No** | Yes | Yes | Via hooks | Yes (GROQ-powered) | Yes (Flows) | No (plugin) | No (signals) | Yes |
| HMAC signing | Planned | No | Yes | No | Yes | No | N/A | N/A | No |
| Retries | Planned | No | No | No | Auto-retry | No | N/A | N/A | Yes (7 retries) |
| Delivery logs | Planned | No | Yes (500/webhook) | No | No | No | N/A | N/A | Yes |

**ModulaCMS is behind.** Webhooks are the single biggest feature gap. Every headless CMS ships webhooks (except WordPress and Wagtail natively). Without them, published content changes cannot trigger CDN invalidation, SSG rebuilds, or external notifications. This is Phase 3 in the roadmap.

---

## 13. SDKs

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| TypeScript | Yes (3 packages) | Yes (1) | Yes (CDA + CMA) | No official | Yes | Yes (composable) | No official | No official | Yes (CDA + CMA + React/Vue/Svelte) |
| Go | Yes | No | No | No | No | No | No | No | No |
| Swift | Yes | No | Yes | No | No | No | No | No | No |
| Python | No | No | Yes | No | Yes | No | No | No | No |
| Ruby | No | No | Yes | No | Yes | No | No | No | Yes |
| Java/.NET | No | No | Yes (both) | No | No | No | No | No | No |
| PHP | No | No | Yes | No | No | No | No | No | No |

**ModulaCMS leads** in SDK breadth for a self-hosted CMS. Three languages (TypeScript with 3 packages, Go, Swift) with branded IDs and full CRUD coverage. No other self-hosted CMS offers official Go or Swift SDKs. Payload has no official SDK at all. Strapi has one JS client.

**Behind** Contentful (8 languages, separate CDA + CMA SDKs per language) in total language coverage.

---

## 14. MCP Server

| Has MCP | ModulaCMS | Contentful | Directus |
|---------|-----------|------------|----------|
| | Yes | Yes | Yes (v11.13) |

**On par.** ModulaCMS, Contentful, and Directus are the only CMSs with official MCP servers for AI assistant integration. Most competitors have none.

---

## 15. Import / Migration

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Import from competitors | Contentful, Sanity, Strapi, WordPress | No | No | No | No | No (DB-level) | WXR format | dumpdata/loaddata | Contentful, WordPress |
| Output format compat | 6 formats | No | No | No | No | CSV/XML/YAML | No | No | No |
| Deploy sync | Yes (push/pull/snapshot) | transfer CLI | Environment aliases | DB migrations | Dataset copy | Schema snapshot/apply | No | No | Environment forking |

**ModulaCMS leads** in migration-friendliness. Built-in importers for 4 competitor formats plus bidirectional transform support. The 6 output format transformers mean frontends built for Contentful/Sanity/Strapi/WordPress can consume ModulaCMS content without modification. Deploy sync with push/pull/snapshots and conflict policies is more complete than most competitors.

---

## 16. Backup & Restore

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Built-in backup | Yes (SQL + media → zip) | CLI export (tar) | CLI export (JSON) | No | Dataset export (NDJSON) | Schema snapshot only | No (plugin) | No | No (env fork) |
| Incremental | Enum (not impl) | No | No | No | No | No | Plugin | No | No |
| S3 upload | Yes | No | No | N/A | N/A | No | Plugin | No | N/A |
| Backup history in DB | Yes | No | No | No | No | No | No | No | No |

**ModulaCMS leads.** Built-in backup with SQL dump + media archive to zip, S3 upload, backup history tracking in DB, and verification status. Most competitors either have no built-in backup (Payload, WordPress, Wagtail, DatoCMS) or offer limited export-only tools (Strapi, Contentful, Directus schema-only).

---

## 17. Observability & Auditing

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| Structured logging | Yes (slog) | Yes (Winston) | N/A | Configurable | N/A | Yes (pino) | WP_DEBUG | Django logging | N/A |
| Audit trail | Yes (change_events, free) | Enterprise only | Premium only | Version history | Enterprise | Yes (activity, free) | Plugin | Yes (PageLogEntry, free) | Enterprise only |
| HLC timestamps | Yes | No | No | No | No | No | No | No | No |
| Request IDs | Yes | No | N/A | No | N/A | No | No | No | N/A |
| Observability providers | Sentry, Datadog, NR | Community plugins | N/A | No | N/A | No | Plugin | No | N/A |

**ModulaCMS leads** in audit trail for a self-hosted CMS. The `change_events` table captures all DB mutations atomically with old/new JSON values, HLC timestamps for distributed ordering, and full request metadata -- available on the free tier. Strapi, Contentful, Sanity, and DatoCMS all gate audit logs behind enterprise/premium plans. Only Directus and Wagtail match by offering free audit trails.

---

## 18. Pricing & Licensing

| | ModulaCMS | Strapi | Contentful | Payload | Sanity | Directus | WordPress | Wagtail | DatoCMS |
|-|-----------|--------|------------|---------|--------|----------|-----------|---------|---------|
| License | Open source | MIT (open core) | Proprietary SaaS | MIT | Proprietary SaaS | BSL 1.1 | GPLv2 | BSD 3-Clause | Proprietary SaaS |
| Enterprise-gated features | None | SSO, audit, workflows, history | Custom roles, SSO, audit | Audit (cloud) | Custom roles, releases, SSO, audit | Revenue threshold | None | None | Workflows, audit, SSO, translator roles |
| Managed hosting | No | Strapi Cloud | Yes (only) | Payload Cloud | Yes (Content Lake) | Directus Cloud | WordPress.com | No | Yes (only) |

**ModulaCMS leads** in feature parity across tiers. Every feature (RBAC, audit trail, version history, i18n, publishing) is available for free. Strapi gates 5 features behind paid plans. Contentful gates 8+ features. DatoCMS gates 4+. Sanity gates custom roles. Only WordPress and Wagtail match on having no feature gating.

**Behind** in managed hosting. ModulaCMS is self-hosted only. Strapi Cloud, Payload Cloud, Directus Cloud, and the SaaS platforms (Contentful, Sanity, DatoCMS) all offer managed hosting that removes operational burden.

---

## Summary Scorecard

| Dimension | Rating | Notes |
|-----------|--------|-------|
| Architecture & deployment | **Leads** | Single binary, built-in HTTPS, SSH TUI -- unmatched operational simplicity |
| Database support | **Leads** | Tri-database with compile-time safety. Behind Directus in total count |
| Content modeling | **Leads** | Extensible field type registry, native tree structures, block editor with datatype-backed composable blocks; behind only in conditional fields |
| Content delivery API | **Leads** | 6 output format transforms unique; in-memory filtering (no SQL injection); GraphQL excluded by design (violates core values); behind only in real-time |
| Publishing & versioning | **Leads** | Snapshot-based, optimistic locking, labeled versions, configurable retention |
| Internationalization | **Leads** | Shared-tree i18n, non-translatable field handling, locale=* |
| Media management | **On par** | Solid basics; behind in on-the-fly transforms and video |
| Auth & RBAC | **Leads** | SSH key auth unique; dual-layer field-level perms (roles column) simpler than competitors; free tier -- competitors gate this behind enterprise |
| Admin panel | **Mixed** | Server-rendered + TUI is unique; behind in live preview, collab, dark mode |
| SSH TUI | **Leads** | Unique -- no competitor has anything comparable |
| Plugin system | **Leads** | Sandboxed Lua with circuit breakers; behind in ecosystem size |
| Webhooks | **Behind** | Biggest gap. Every headless CMS has them. Phase 3 planned |
| SDKs | **Leads** | 3 languages (TS/Go/Swift) with branded IDs. Best for self-hosted |
| MCP server | **Leads** | Only 3 CMSs have official MCP servers |
| Import/migration | **Leads** | 4 importer formats + 6 output transforms. Unique |
| Backup & restore | **Leads** | Built-in with history tracking. Most competitors have none |
| Observability & auditing | **Leads** | Free audit trail with HLC timestamps. Competitors gate this |
| Performance & infrastructure | **Leads** | 25K+ req/s on 50MB RAM. Single binary, no runtime. Competitors need 2-4GB for 10-1,000 req/s |
| Admin panel as CMS content | **Leads** | Unique dual content model. No competitor has a parallel admin content system |
| Pricing | **Leads** | Zero feature gating. Everything free |

### Critical gaps to close (in priority order)

1. **Webhooks** (Phase 3) -- table-stakes for any headless CMS
2. **Review workflows** (Phase 4) -- needed for team content governance
3. **Live preview** -- push-based preview updates for dev-time editor experience (SSE/WebSocket to open preview tab)

## 19. Performance, Infrastructure & Scaling

### Throughput

| | Throughput (req/s) | Context |
|-|-------------------|---------|
| **ModulaCMS (Go)** | 25,000-120,000+ | Go stdlib HTTP with real DB backend. TechEmpower Round 23: GoFrame 658K RPS (JSON), Fiber 12M RPS (plaintext). Sub-50ms p95 at 10K RPS in production tuning scenarios |
| **Strapi v5** | ~10 | Payload's GraphQL benchmark: 102ms avg/request, 100 queries in 10.2s. Complex documents with 60+ relationships |
| **Payload 3.x** | ~66 | Same benchmark: 15ms avg/request, 100 queries in 1.5s. 3x faster than Directus, 7x faster than Strapi |
| **Directus** | ~22 | Same benchmark: 45ms avg/request, 100 queries in 4.5s. Without caching: ~1.5s avg. With caching: ~72ms |
| **WordPress** | 150-290 (uncached), 3,200-15,800 (cached) | PHP 8.3. Cached numbers with OpenLiteSpeed (15.8K) or Nginx (3.2K). Full-page cache reduces TTFB 50-90% |
| **Wagtail/Django** | 100-1,000+ | Depends on view complexity. Django + Gunicorn: 9K-14K on optimized Arm64 infrastructure. GIL limits per-worker parallelism |
| **Contentful** (SaaS) | 55-78 req/s (rate limit), CDN unlimited | CDN-cached responses exempt from rate limits. Fastly-backed |
| **Sanity** (SaaS) | 500 req/s per IP, CDN unlimited | 25 mutations/s. 500 concurrent queries per dataset |
| **DatoCMS** (SaaS) | 40 req/s per token, CDN unlimited | 20 req/s CMA. Queries >8KB compressed bypass CDN cache |

**Note:** Self-hosted CMS numbers for Strapi, Payload, and Directus are from Payload's comparative benchmark (complex documents, 60+ relationships, local database, sequential queries). Not peak synthetic throughput. Go baseline is for realistic HTTP workloads with database queries.

### Memory & CPU

| | Min RAM | Recommended RAM | Min CPU | Typical production footprint |
|-|---------|----------------|---------|------------------------------|
| **ModulaCMS (Go)** | ~50MB | 512MB-1GB | 1 core | Single binary, 50-200MB depending on connections. GOMEMLIMIT for containers |
| **Strapi v5** | 2GB | 4GB+ | 1 core | 500MB-2GB+ before traffic. 32GB+ disk recommended |
| **Payload 3.x** | 1GB | 2-8GB | 1 vCPU | Build process memory-intensive (may crash <2GB). 27 dependencies |
| **Directus** | 512MB | 2GB (4GB for large images) | 0.25 vCPU | Image processing spikes. Redis recommended |
| **WordPress** | 512MB | 1-2GB | 1 core | Each PHP-FPM worker: 30-60MB. Plugin count multiplies memory |
| **Wagtail** | 512MB | 1-2GB | 2 cores | Per Gunicorn worker: 50-150MB. 4-8 workers typical (200MB-1.2GB) |
| **Contentful** | N/A | N/A | N/A | SaaS -- no infrastructure to manage |
| **Sanity** | N/A | N/A | N/A | SaaS -- Studio is a static React app |
| **DatoCMS** | N/A | N/A | N/A | SaaS -- no infrastructure to manage |

### Required Infrastructure

| | Runtime | Database | Additional requirements |
|-|---------|----------|------------------------|
| **ModulaCMS** | Single binary | SQLite, MySQL, or PostgreSQL | Nothing else. Built-in HTTPS (Let's Encrypt), built-in SSH server. Optional: reverse proxy, S3 |
| **Strapi** | Node.js 18+ | SQLite, PostgreSQL, or MySQL | Reverse proxy (Nginx). SSD recommended |
| **Payload** | Node.js (Next.js) | MongoDB or PostgreSQL | Reverse proxy or Vercel/Netlify. File storage adapter |
| **Directus** | Node.js 18+ | PostgreSQL, MySQL, MariaDB, SQLite, MS SQL, OracleDB, CockroachDB | Redis (recommended, required for horizontal scaling). Reverse proxy |
| **WordPress** | PHP 8.3+ | MySQL 8.0+ or MariaDB 10.6+ | Web server (Apache/Nginx). PHP extensions. OPcache. Cache plugin. Object cache (Redis/Memcached) |
| **Wagtail** | Python 3.x + Gunicorn | PostgreSQL (recommended), MySQL, SQLite | Nginx reverse proxy. Redis (caching + Celery). Elasticsearch (optional search). Celery (async tasks) |

### Scaling

| | Approach | Horizontal scaling complexity |
|-|----------|------------------------------|
| **ModulaCMS** | Stateless binary behind load balancer. Vertical also effective (goroutine concurrency). No shared state required between instances | Low -- add instances, point at same DB |
| **Strapi** | PM2 cluster mode or Kubernetes. Stateless API servers | Moderate -- needs load balancer, shared storage for uploads |
| **Payload** | Standard Next.js scaling. Serverless on Vercel/Netlify | Low-Moderate -- serverless simplifies, but DB connection pooling needed |
| **Directus** | Multiple instances + Redis + load balancer. Redis required for session/WebSocket sync | Moderate -- Redis dependency adds operational complexity |
| **WordPress** | PHP-FPM tuning, then horizontal with shared DB + shared uploads (NFS/S3) + shared sessions (Redis) + DB replication | High -- many moving parts (shared filesystem, session store, DB replicas, object cache, page cache) |
| **Wagtail** | Gunicorn workers + Celery workers + horizontal app servers + shared DB + S3 media + Varnish/Squid cache proxy | High -- similar to WordPress complexity with Python-specific tooling |

**ModulaCMS leads decisively.** A Go binary with built-in HTTPS serving 25K+ req/s on 50MB of RAM is a different category from Node.js/Python/PHP CMSs requiring 2-4GB RAM to serve 10-1,000 req/s. The infrastructure requirements table tells the story: ModulaCMS needs a binary and a database. Competitors need a language runtime, package manager, web server, reverse proxy, cache layer, and often Redis. WordPress and Wagtail horizontal scaling requires 5-6 separate infrastructure components coordinated together.

The SaaS platforms (Contentful, Sanity, DatoCMS) trade raw performance for zero infrastructure management, but their API rate limits (40-78 req/s for uncached requests) are lower than what a single ModulaCMS instance delivers.

---

## 20. Admin Panel as CMS Content (Dual Content Model)

| | Has dual content model | Admin content managed as CMS content | Approach |
|-|----------------------|-------------------------------------|----------|
| **ModulaCMS** | **Yes** | **Yes** | Complete parallel content system: `admin_content_data`, `admin_datatypes`, `admin_fields`, `admin_field_types`, `admin_content_fields`, `admin_routes`, `admin_content_relations`, `admin_content_versions` -- all with same sibling-pointer tree, locale support, publishing, and versioning as the public content model |
| **Strapi** | No | No | Fixed React SPA. Code-level customization via injection zones and `src/admin/app.js`. Admin structure is hardcoded |
| **Contentful** | No | No | Fixed web app. App Framework allows iframe-based extensions at predefined locations. Admin consumes CMA/CDA through custom apps but this is emergent, not designed |
| **Payload** | No (closest competitor) | Partial | Admin is a Next.js app consuming its own Local API. Globals (singletons) serve as admin-specific content. Custom views (RSC) can query collections directly. But no structural distinction between admin and public content types |
| **Sanity** | No | Partial | Studio is a React app querying the Content Lake via GROQ. Structure Builder configures admin layout via code. Dashboard widgets can display GROQ query results. But layout config is code, not content |
| **Directus** | No (second-closest) | Partial | Insights dashboards query collections and can even create items. Custom modules have full SDK access. But admin layout is auto-generated from DB schema, not content-managed |
| **WordPress** | No | Plugin-only | Fixed wp-admin. Plugins (WP Adminify, Ultimate Dashboard) can create admin pages via block editor. Options API provides key-value settings. Not a parallel content model |
| **Wagtail** | No | Partial | `BaseSiteSetting` / `BaseGenericSetting` provide admin-editable singleton models. Custom admin views via hooks can query ORM. Closest to admin-as-content in spirit but limited to flat settings, not a full content tree |
| **DatoCMS** | No | Partial | Plugin SDK custom pages run in iframes with CMA access. Can query records and render them. But no structural distinction for admin content |

**ModulaCMS is unique.** No other CMS implements a true dual content model. ModulaCMS's admin content system is a complete parallel:

- **`admin_datatypes`** -- content type definitions for admin panel content (parallel to `datatypes`)
- **`admin_fields`** + **`admin_field_types`** -- field definitions and extensible type registry (parallel to `fields` + `field_types`)
- **`admin_content_data`** -- content entries in a sibling-pointer tree with draft/publish, scheduling, and revision tracking (parallel to `content_data`)
- **`admin_content_fields`** -- per-field values with locale support and unique locale constraint (parallel to `content_fields`)
- **`admin_routes`** -- URL-to-content-tree mappings for admin pages (parallel to `routes`)
- **`admin_content_relations`** -- cross-content references within admin content (parallel to `content_relations`)
- **`admin_content_versions`** -- version history for admin content (parallel to `content_versions`)

Every capability of the public content system -- tree navigation, field extensibility, localization, publishing, versioning -- is available for admin panel content. The admin panel literally eats its own dog food: its content is CMS-managed content defined through the same primitives (datatypes, fields, content tree) as the public-facing content.

The closest any competitor comes is **Payload CMS**, where the admin panel is architecturally a Next.js app consuming its own Local API, and you could manually create "admin" collections with `admin.hidden`. But even Payload requires you to build the rendering layer yourself -- the CMS does not provide a parallel admin content model with its own schema, tree, and delivery system.

---

### Competitive position

ModulaCMS occupies a unique niche: **operationally simple, architecturally sophisticated, completely free**. No other CMS combines single-binary deployment, tri-database support, snapshot-based versioning, shared-tree i18n, six output format transformers, SSH TUI, three SDKs, MCP server, sandboxed plugins, and zero feature gating.

The closest architectural peers are **Directus** (SQL-based, self-hosted, database introspection) and **Wagtail** (Django-based, page tree, workflows, BSD). ModulaCMS exceeds both in operational simplicity and exceeds Directus in versioning/i18n. Wagtail leads in workflow maturity and StreamField flexibility.

After Phases 3-4 ship (webhooks + review workflows), ModulaCMS will match or exceed Strapi v5 Enterprise in the publishing/versioning/i18n domain while remaining fully free and self-hosted.
