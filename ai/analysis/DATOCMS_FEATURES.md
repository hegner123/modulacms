# DatoCMS Features

## Runtime Architecture
- **Deployment Model**: Proprietary SaaS, fully managed. No self-hosted option
- **Infrastructure**: Kubernetes on AWS EKS (EC2), auto-scaling. All servers in EU (Ireland). Terraform-managed
- **CDN**: Cloudflare-based global CDN. DDoS mitigation and WAF (OWASP top 10 + custom rules), < 1ms added latency
- **Content Delivery API**: GraphQL at `https://graphql.datocms.com`. CDN-cached, globally distributed
- **Content Management API**: REST (JSON:API) at `https://site-api.datocms.com`. Not cached (write-path)
- **Real-time Updates API**: SSE at `https://graphql-listen.datocms.com`. Ephemeral channel URL (15-second TTL) streams `update` events when query results change
- **Asset CDN**: Imgix-powered image CDN. Mux-powered video streaming
- **TLS**: All connections encrypted, internal and external
- **Rate Limiting**: Separate CDA and CMA budgets

## Database / Storage
- **Fully managed**, opaque to users. No direct database access. Data encrypted at rest. Backups retained up to 14 days
- **AWS-hosted** (specifics not publicly disclosed). Likely PostgreSQL
- **Asset Storage**: Imgix CDN for images, Mux for video. No raw S3 access
- **Data Residency**: All servers and databases in EU (Ireland)

## Content Management
- **Records**: Individual content items created from models. Support draft/published lifecycle, versioning, localization
- **CRUD**: Full CRUD via CMA (REST, JSON:API). GraphQL CDA is read-only
- **Models**: Blueprints for content types. Regular models, single-instance (singletons), block models (`modular_block: true`)
- **Modular Content**: Dynamic layouts from predefined block models. Ordered lists of block records
- **Structured Text**: Rich text as portable JSON (DAST -- DatoCMS Abstract Syntax Tree, based on unist). Supports paragraphs, headings (h1-h6), lists, blockquotes, code blocks, inline formatting, hyperlinks, record links, inline record references, embedded blocks, **inline blocks** (2025), **custom text styles** (2025), slash commands, markdown shortcuts, drag-and-drop
- **Links**: Single link and multiple links (ordered references). Validators control linkable models
- **Ordering**: Manual (drag-and-drop), hierarchical (parent-child tree), automatic (by field value)
- **Tree Structures**: Native first-class support. `parent` and `children` attributes. Drag-and-drop tree reordering. DatoCMS claims to be the only headless CMS with built-in trees
- **Bulk Operations**: Bulk publish/unpublish, bulk delete, batch stage transitions

## Content Schema / Content Modeling
- **Field Types** (21 types): `string`, `text`, `structured_text`, `rich_text` (modular content), `single_block`, `integer`, `float`, `boolean`, `date`, `date_time`, `color`, `file`, `gallery`, `video`, `seo`, `slug`, `json`, `lat_lon`, `link`, `links`
- **Validations**: `required`, `unique`, `length`/`items_count`, `format` (URL, email, regex), `enum`, `number_range`, `date_range`, `file_size`, `extension`, `image_dimensions`, `alt_title_required`, `rich_text_blocks`, `structured_text_blocks`/`inline_blocks`/`links`, `accepted_item_types`, linked record publish/unpublish behavior
- **Fieldsets**: Visual grouping in editor UI
- **Appearance Settings**: Editor widget selection, parameters, addons (supplementary UI below field)
- **Field Extensions**: Plugins provide custom editors and addons
- **Default Values**: Configurable per field, locale-aware
- **Per-field localization**: Individual toggle per field
- **Deep Filtering**: Toggle on block fields for GraphQL deep filtering

## Content Delivery / API

**Content Delivery API (CDA)**:
- GraphQL at `https://graphql.datocms.com`. CDN-backed
- Auto-generated schema. Model `artist` produces `artist` (single) and `allArtists` (collection) queries
- Headers: `X-Include-Drafts: true`, `X-Environment: <name>`

**Content Management API (CMA)**:
- REST (JSON:API) at `https://site-api.datocms.com`
- Full CRUD on all resources

**Filtering**: `eq`, `neq`, `lt`, `lte`, `gt`, `gte`, `in`, `notIn`, `exists`, `isBlank`, `matches` (full-text). Filter by `_status`, `_createdAt`, `_updatedAt`, `_publishedAt`, `_isValid`

**Pagination**: Offset-based with `first` (max 500) and `skip`. `_allXxxMeta { count }` for totals. Auto-pagination utility in `@datocms/cda-client`

**Responsive Images**: `responsiveImage` object with `src`, `srcSet`, `width`, `height`, `alt`, `title`, `base64` (LQIP), `bgColor`, `sizes`. Full imgix parameter support

**Video**: `muxPlaybackId`, `streamingUrl` (HLS), `mp4Url`, dimensions, duration, `thumbnailUrl`, `thumbhash`

**Real-time Updates API**: Same GraphQL queries via SSE. Streams `update` events on content changes. Auto-reconnect in official clients

**Cache Tags**: CDA responses include tags for granular cache invalidation. `cda_cache_tags.invalidate` webhook event

## Content Publishing
- **Draft/Published**: Opt-in per model. States: `draft`, `published`, `updated` (dirty published)
- **Allow Saving Invalid Drafts**: Per-model toggle. Validations enforced only at publish
- **Scheduled Publishing/Unpublishing**: Calendar picker, per-record. Auto-triggers build triggers. Per-locale scheduling
- **Publication Workflow** (Enterprise): Customizable state machine. Named stages, per-role transitions, batch stage transitions
- **Bulk Publish/Unpublish**: Multi-select batch operations
- **Locale-Based Publishing**: Publish/unpublish individual locales independently
- **Dependency Tracking on Unpublish**: Checks for linking published records, can warn or block

## Content Versioning
- **Automatic Versioning**: Every save and publish creates new version
- **Version History**: Timeline view with date and author
- **Version Comparison (Diff)**: Visual field-level diff between any two versions
- **Restore**: Any previous version restorable
- **Current Version ID**: Exposed as `meta.current_version` in CMA

## Internationalization (i18n)
- **Per-Field Localization**: Each field individually configured. No all-or-nothing model toggle
- **Locale Management**: ISO locale codes at project level
- **Editor UI**: One tab per locale when localized fields exist
- **Optional/Required Locales**: Per-model "All locales required?" toggle
- **Locale-Based Publishing**: Independent per-locale publish/unpublish with scheduling
- **Translator Roles** (Enterprise): Per-role locale permissions, globally or per-model
- **No built-in fallback locales**: Handled on frontend/client side
- **2026 Pricing**: 5 default locales per plan, add-ons available

## Media / Asset Management
- **Image CDN (Imgix)**: Every image on Imgix. On-the-fly transforms: crop, resize, format conversion (auto WebP/AVIF), quality, blur, sharpen, watermark, hundreds of parameters
- **Automatic Image Optimization**: Configurable project-wide imgix optimization
- **Responsive Images**: `responsiveImage` GraphQL object with `srcSet`, `sizes`, `base64` (LQIP), dominant colors, BlurHash, ThumbHash
- **Focal Point**: Set in asset editor. Respected by crop operations
- **Video Management (Mux)**: Auto-encoded for streaming. HLS, MP4 (high/medium/low), thumbnails, ThumbHash. Video encoding costs removed in 2026
- **Upload from URL**: Via CMA
- **Asset Metadata**: Title, alt text, copyright/notes, custom metadata. Alt/title can be required
- **AI-Powered Alt Text**: Marketplace plugin for image alt text generation
- **Folders/Tags**: Asset organization
- **Image Components**: Official `<Image />` for React, Vue, Svelte. `<VideoPlayer />` wrapping Mux

## Authentication
- **API Tokens**: Created in Settings. Linked to Role. Flags: CDA access, CDA preview mode, CMA access
- **Role-Based Tokens**: Custom roles restrict GraphQL schema visibility entirely
- **SSO (SAML 2.0 + SCIM 2.0)**: Enterprise. Okta, OneLogin, Azure AD. Automated provisioning
- **2FA**: Available for all users. Can be enforced per-project
- **Organizations**: Share account ownership without sharing credentials

## Authorization / RBAC
- **Default Roles**: Admin (full), Editor (records only)
- **Custom Roles**: Unlimited with granular permissions
- **Role Inheritance**: Hierarchical permission inheritance
- **Project-Wide Permissions**: Models/fields/navigation, languages, deployment, roles, webhooks, API tokens, shared filters
- **Environment Access**: All, primary only, or sandbox only per role
- **Content-Level Permissions**: Per environment, per model, per creator. Actions: View, Create, Edit, Publish, Delete, Take over, Move to stage. Additive or subtractive rules. Creator/role restrictions
- **Locale Permissions** (Enterprise): Per-role locale editing restrictions
- **Workflow Stage Permissions**: Per-role transition control
- **Asset Permissions**: View, Create, Edit metadata/replace, Delete, Edit creator
- **Fail-closed**: Everything not explicitly allowed is denied

## Admin UI
- **Web-Based**: Browser-based at `*.admin.datocms.com`
- **Content Editor**: Field-type-specific editors, locale tabs, validation feedback, draft/publish/schedule
- **Schema Builder**: Visual model/field editor with drag-and-drop, fieldset grouping
- **Media Browser**: Upload, search, filter, focal point, metadata, folders/tags
- **Build Triggers**: Netlify, Vercel, Travis CI, GitLab CI, CircleCI, custom webhook URLs. Auto-trigger on publish
- **Environment Switcher**: Top bar. Permission-controlled visibility
- **Plugin UI Extensions**: Custom field editors, field addons, sidebar panels, custom pages, asset sources, upload adapters

## Plugin / Extension System
- **Architecture**: Plugins as web apps in sandboxed `<iframe>`. Plugin SDK (`@datocms/plugins-sdk`) with hooks. Full TypeScript support. React UI library (`datocms-react-ui`)
- **Extension Types**: Field editors, field addons, sidebar panels, custom pages, asset sources, upload adapters, build trigger extensions
- **Marketplace**: 100+ open-source community plugins. Free installation
- **Private Plugins**: Organization-specific, plan-limited

## Webhooks
- **Events**: Record (create/update/delete/publish/unpublish), Model (create/update/delete), Upload (CRUD), Build trigger (deploy events), Environment (deploy events), Maintenance Mode, SSO User, CDA Cache Tags invalidate
- **Payload**: JSON with `site_id`, `webhook_id`, `environment`, `entity_type`, `event_type`, `entity`, `previous_entity` (for updates)
- **Custom Payloads**: Mustache-templated body and URL
- **Authentication**: HTTP basic auth + custom headers (no HMAC signing documented)
- **Automatic Retries**: Optional. Up to 7 retries with exponential backoff (2min, 6min, 30min, 1hr, 5hr, 1day, 2days)
- **Timeouts**: Connection: 2s, total: 8s
- **Delivery Logs**: Activity log filterable by status/event/date. Manual resend capability
- **Per-Environment**: Payload includes environment field

## SDKs & Client Libraries
- **JS/TS (CMA)**: `@datocms/cma-client-node`, `@datocms/cma-client-browser`. Auto-pagination, rate limit handling, retry logic
- **JS/TS (CDA)**: `@datocms/cda-client`. `executeQuery`, `executeQueryWithAutoPagination`
- **React**: `react-datocms` -- Image, VideoPlayer, StructuredText components, `useQuerySubscription` hook
- **Vue.js**: `vue-datocms`
- **Svelte**: `@datocms/svelte`
- **Ruby**: `datocms-client` gem
- **Real-time**: `datocms-listen` (SSE client)
- **Structured Text Utilities**: `datocms-structured-text-utils`, `datocms-structured-text-to-html-string`, `datocms-html-to-structured-text`, `datocms-contentful-to-structured-text`

## CLI
- **Package**: `@datocms/cli` (`npx datocms` or `dato`)
- **Environment Management**: `environments:list`, `environments:fork`, `environments:promote`, `environments:rename`, `environments:destroy`
- **Maintenance Mode**: `maintenance:on` (with `--force`), `maintenance:off`
- **Schema Migrations**: `migrations:new`, `migrations:run`, `--autogenerate` (diff between environments), TypeScript support
- **Recommended Workflow**: Fork primary to sandbox, write migration scripts, maintenance mode on, run migrations, test, promote, maintenance off

## Deployment / Environments
- **Primary Environment**: One per project, named `main` by default
- **Sandbox Environments**: Deep copies (forks). Independent copies of models, records, uploads, plugins, config
- **Promotion**: Instantly promote sandbox to primary. No service interruption
- **Force Sandbox**: Block all users from editing primary schema directly
- **Build Triggers**: Integration with Netlify, Vercel, Travis, GitLab, CircleCI, custom webhooks
- **2026 Limits**: 3 environments per project (reduced from 8). Extra at EUR 39/month

## Email
- **No built-in email system**. Only invitation and account management emails
- **Integrations**: Via webhooks + external services

## Backup & Restore
- **No built-in backup tool**
- **Environment Forking**: Recommended approach for state snapshots
- **Database Backups**: Internal, retained 14 days, not user-accessible
- **CMA Export**: Programmatic via API

## Configuration
- **Dashboard Settings**: Locales, timezone, SEO, appearance, deployment, SSO, tokens, roles, webhooks, environments, plugins, maintenance
- **API-Based**: All configuration manageable via CMA
- **Scripted Migrations**: Version-controlled TypeScript/JavaScript migration scripts

## Observability / Logging
- **API Usage Dashboard**: Call counts, bandwidth, record counts vs plan limits
- **Rate Limit Monitoring**: API headers + dashboard
- **Status Page**: Public platform health monitoring
- **Webhook Activity Log**: Delivery tracking with status, timestamps, payload

## Audit Trail
- **Audit Logs** (Enterprise only): Complete JSON history of project events. Available via API for SIEM integration
- **Version History**: Per-record timeline with author/timestamp and diff. All plans
- **Webhook Payloads**: `previous_entity` provides before/after diff

## Content Migration
- **CMA REST API**: Full programmatic access for migrations
- **Scripted Migrations**: CLI-based with `--autogenerate` schema diff
- **Contentful Import**: Built-in `datocms contentful:import`
- **WordPress Import**: CLI support
- **Structured Text Migration**: HTML-to-DAST and Contentful-to-DAST converters

## Pricing / Licensing
- **Proprietary SaaS**: No self-hosted option
- **Free Plan**: 3 projects, 2 editors, 300 records, 10 models, 10 GB traffic/month, 100K API calls/month
- **Professional**: EUR 149/month (annual). 10 collaborators, 100 models, 5 locales, 3 environments, 50K video streaming minutes
- **Enterprise**: Custom pricing. Workflows, Audit Logs, Translator roles, SAML/SCIM SSO, SLA
- **Discounts**: 50% for education/nonprofits, 30% for agencies
- **ISO 27001 Certified**

## Unique Features
- **Structured Text (DAST)**: Rich text as portable JSON AST based on unist. Embeds typed blocks and record links in text. Inline blocks (2025). Most distinctive technical feature
- **Real-time Updates API**: SSE-based live content streaming. Same queries as CDA. No polling
- **Imgix Image Pipeline**: Hundreds of on-the-fly parameters. Pre-computed `responsiveImage` with BlurHash/ThumbHash/dominant color
- **Mux Video Integration**: First-class streaming with HLS, adaptive bitrate, thumbnails. Encoding costs removed 2026
- **Built-in Tree Structures**: Native hierarchical content with `children`/`parent` GraphQL fields
- **Sandbox Environments**: Git-branch-like content environments. Fork, test, promote
- **CDN-Backed GraphQL**: Edge CDN serving, sub-100ms global response times
- **Locale-Based Publishing**: Independent per-locale publish with scheduling
- **SEO Meta Tags Field**: Dedicated field type bundling title, description, OG image, Twitter card, no-index
- **Cache Tag Invalidation Webhook**: Precise ISR with Next.js/Nuxt
- **ISO 27001 Certification**
