# Contentful Features

## Runtime Architecture
- **SaaS-only** -- Contentful is a fully managed cloud SaaS platform. There is no self-hosted option.
- **Infrastructure**: Hosted on Amazon Web Services (AWS). The platform is a composite of dozens of individual microservices (authentication, resource counting, asset transformation, content delivery, etc.).
- **CDN**: Content Delivery API and Images API are served through Fastly CDN (80+ global edge locations, 6 continents). Asset delivery also uses Amazon CloudFront. Cache purge after content update completes in under 150 milliseconds.
- **Edge caching**: CDN-cached CDA requests are not rate-limited. Advanced caching options (Premium only) include:
  - **Stale-when-rate-limited**: serves stale cached data instead of 429 errors during traffic spikes.
  - **Stale-while-revalidate**: returns stale data for 1 minute after publish while refreshing in background.
- **Data residency**: EU data residency available (primary: AWS Ireland `eu-west-1`, secondary: AWS Frankfurt `eu-central-1`). CDN still caches globally; EU residency applies to data storage, not delivery.
- **Multi-Region Delivery Infrastructure (MRDI)**: Premium feature. Content replicated across primary and secondary regions for active-active delivery. 99.99% availability SLA covering CDA, GraphQL, Images API, and asset CDN.
- **API endpoints**:
  - CDA: `cdn.contentful.com` (read-only, published content)
  - CMA: `api.contentful.com` (read-write, management)
  - CPA: `preview.contentful.com` (read-only, draft + published)
  - GraphQL: `graphql.contentful.com`
  - Images: `images.ctfassets.net`
  - User Management API and SCIM API also available.

## Database Support
- **Fully managed** -- users have zero access to or control over the underlying database. Contentful does not publicly disclose its internal database technology.
- The database is sharded across many server instances. Content is stored as structured JSON documents exposed through APIs.
- Users cannot bring their own database, choose a database engine, or run queries against any backing store.
- Data is accessible only through the REST/GraphQL APIs.

## Content Management
- **CRUD operations**: Full create, read, update, delete via CMA (REST). All mutations require CMA authentication.
- **Content types**: Define content schemas (up to 1,000 per environment). Each has up to 50 fields. Content types have a `displayField` for entry title representation.
- **Entries**: Individual content items conforming to a content type. Up to 5,000,000 records per space (entries + assets combined). Max entry size: 2MB. Max 1,000 links (references, media, rich text) per entry.
- **Assets**: Binary files (images, documents, video). Three fixed fields: name, description, file. Max asset size: 50MB (Free) / 1,000MB (paid). Max image size: 100MB, 300 megapixels.
- **References**: Link fields point to other entries or assets. Supported as single reference, array of references, or embedded in Rich Text. Reference resolution depth: up to 10 levels via `include` parameter on CDA.
- **Cross-space references**: Link entries across different spaces within an organization. Supports references from up to 3 spaces/environments. Available for standard reference fields and Rich Text embedded entries.
- **Ordering**: Array fields support ordered lists of symbols or links. Entry ordering in API responses via `order` parameter on any field.
- **Rich Text**: Structured JSON document model (not flat HTML/Markdown). Node types include: paragraph, heading-1 through heading-6, unordered-list, ordered-list, list-item, blockquote, hr, table, table-row, table-cell, table-header-cell, hyperlink, embedded-entry-block, embedded-entry-inline, embedded-asset-block, text. Max 200,000 characters / 1MB payload. Toolbar can be restricted per field to limit allowed formatting. Supports tables and cross-space references as embedded entries.
- **Archive/Unarchive**: Entries and assets can be archived (hidden but preserved) and unarchived.
- **Bulk actions**: Publish, unpublish, or validate up to 200 items per bulk action. Max 5 concurrent bulk actions per space. Bulk actions run asynchronously.
- **Tags**: Up to 1,000 tags per environment. Up to 100 tags per entry or asset.
- **Taxonomy**: Structured hierarchical classification system with concept schemes (up to 20 per org), concepts (up to 6,000 per org, 2,000 per scheme, 50 per entry). Preferred and alternative labels. Taxonomy concepts deliverable via CDA. AI-powered tagging suggestions available.

## Content Schema / Content Modeling

**Field types** (13 types):

| Name | API Type | JSON Type | Max Size |
|------|----------|-----------|----------|
| Short Text | `Symbol` | String | 256 chars |
| Long Text | `Text` | String | 50,000 chars |
| Rich Text | `RichText` | Object | 200,000 chars / 1MB |
| Integer | `Integer` | Number | -2^53 to 2^53 |
| Decimal | `Number` | Number | -2^53 to 2^53 |
| Date/Time | `Date` | String | ISO 8601 format |
| Location | `Location` | Object | lat/lon coordinates |
| Boolean | `Boolean` | Boolean | true/false |
| Media (Asset) | `Link` (linkType: Asset) | Object | Single asset reference |
| Reference (Entry) | `Link` (linkType: Entry) | Object | Single entry reference |
| Array | `Array` | Array | List of Symbols or Links |
| JSON Object | `Object` | Object | Arbitrary JSON |
| Multiple media/references | `Array` of `Link` | Array | Multiple asset or entry references |

**Validations** (per field type):

| Validation | Applicable To |
|-----------|---------------|
| Size (min/max length) | Text, Symbol, Object, Multiple symbols/entries/assets |
| Predefined values (up to 50) | Text, Symbol, Number, Decimal |
| Regex pattern matching | Text, Symbol |
| Date range | Date |
| Number range | Number, Decimal |
| Number of items (min/max count) | Multiple symbols/entries/assets |
| Asset file size (min/max) | Asset, Multiple assets |
| Image dimensions (min/max w/h) | Asset, Multiple assets |
| Type constraint (restrict linked content type or MIME group) | Entry, Multiple entries, Asset, Multiple assets |
| Required | All field types |
| Unique | Symbol |

Built-in regex presets: email, URL, US date, European date, US phone, US zip code, 12-hour time, 24-hour time.

**Appearances / Editors**: Fields can have assigned editor widgets. Default editors include: single line, multi line, dropdown, tags, list, checkbox, radio, boolean, rating, number, URL, JSON editor, location picker, date picker, markdown, slug, entry/asset reference picker, rich text. Custom editors can be built via App Framework.

**Content model templates**: Reusable content type templates (up to 200 versions per template on paid plans).

## Content Delivery / API

**Content Delivery API (CDA)**:
- Read-only REST API for published content at `cdn.contentful.com`.
- Rate limits: 55 req/sec (Free), 78 req/sec (paid). Cached requests do not count.
- **Filtering**: `all`, `in`, `nin`, `exists`, `match`, `gt`, `gte`, `lt`, `lte`, `ne`, `near`, `within` operators. Full-text search on Text and RichText fields. Relational queries across linked entries.
- **Sorting**: `order` parameter, prefix with `-` for descending.
- **Pagination**: Offset-based (`skip` + `limit`, max 1,000 items per request) and cursor-based pagination (opaque cursor tokens with `next`/`prev` links).
- **Include/Expansion**: `include` parameter (0-10) resolves linked entries/assets. Default depth is 1.
- **Locale**: `locale` query parameter to request specific locale or `locale=*` for all locales.
- Max response size: 7MB. Max URI length: 7,600 characters.

**Content Management API (CMA)**:
- Read-write REST API at `api.contentful.com`.
- Rate limits: 7 req/sec (Free), 10 req/sec (paid).
- Max request size: 1MB.
- Supports all CRUD operations on entries, assets, content types, locales, webhooks, roles, environments, etc.

**Content Preview API (CPA)**:
- Same interface as CDA but includes draft (unpublished) content.
- Rate limits: 14 req/sec (Free), 20 req/sec (paid).

**GraphQL Content API**:
- Available at `graphql.contentful.com/content/v1/spaces/{SPACE_ID}`.
- Supports both CDA and CPA data.
- Rate limit: 55 req/sec.
- Max request size: 8KB.
- Auto-generated schema from content model.

**Sync API**:
- Initial sync returns all published content (optionally filtered by content type).
- Subsequent calls with `syncToken` return only deltas (created, updated, deleted items).
- Pagination via `nextPageUrl` (max 1,000 items per page).
- Designed for keeping external data stores in sync.

**Images API**:
- No authentication required.
- Transformations: resize (`w`, `h`), crop (`fit` parameter with face detection), format conversion (`fm=webp`, `fm=avif`, `fm=jpg`, `fm=png`), quality (`q=1-100`), rounded corners / circle crop, background color.
- Max source image: 100MB, 300 megapixels. AVIF source max: 9 megapixels.
- Served via CDN.

## Content Publishing
- **Draft/Published lifecycle**: New entries start as drafts. Must be explicitly published to appear in CDA. Entries can be in states: Draft, Changed (published but has newer draft), Published, Archived.
- **Scheduled publishing**: Schedule individual entries/assets for future publish or unpublish. Up to 500 scheduled actions per environment. Max scheduling horizon: 24 months. Max 200 scheduled actions executable per minute. Available on Lite and Premium plans.
- **Releases**: Group up to 200 entities (entries + assets) into a release for coordinated publishing. Types: Scheduled (for a future date) and Ideation (for collaboration). Max 2 release actions per minute. Available on Premium plans.
- **Timeline Releases**: Schedule and preview multiple future versions of the same entry. Up to 80 total timeline releases per environment. Up to 30 active scheduled releases. Max scheduling horizon: 4 months.
- **Calendar view**: Visual calendar showing all scheduled releases and individual publish/unpublish events.
- **Locale-based publishing** (Premium): Publish content per locale independently.
- **Bulk actions**: Publish, unpublish, or validate up to 200 items atomically in a single API call.

## Content Versioning
- **Auto-incrementing version numbers**: Every save increments the entry's `sys.version` counter. The `X-Contentful-Version` header is required on CMA updates for optimistic concurrency control.
- **Version history**: View all previously published versions in the web app's Versions sidebar widget.
- **Snapshot comparison**: Side-by-side diff view comparing any previous version with the current version. Changed sections highlighted; unchanged sections greyed out.
- **Restore**: Roll back to any previous published version with a confirmation dialog. Restoring creates a new version (does not destroy intermediate versions).
- **Author tracking**: Each version records which user published it and when.
- **Snapshots API**: Programmatic access to entry snapshots via CMA (`/entries/{id}/snapshots`).
- **Retention**: Contentful does not publicly document a retention limit for version history. Deleted spaces/environments can be recovered if requested within 25 days.

## Internationalization (i18n)
- **Locales**: Up to 500 locales per environment (and per organization). Configure via web app or CMA.
- **Default locale**: Every space has one default locale. Cannot be removed.
- **Per-field localization**: Localization is toggled per field on each content type ("Enable localization of this field"). Non-localized fields share a single value across all locales. Allows mixing localized (e.g., title, body) and non-localized (e.g., date, price) fields on the same entry.
- **Fallback locales**: Each locale can specify a fallback locale, forming a fallback chain/tree. Example: `de-CH` falls back to `de-DE`, which falls back to `en-US`. If a field value is not set (not null, not empty string -- literally absent), CDA returns the fallback locale's value.
- **Fallback behavior nuance**: Setting a field to `null` or empty string prevents fallback. Only an absent value triggers fallback.
- **Locale visibility**: Toggle which locale fields are visible in the entry editor UI.
- **API locale parameters**: CDA supports `locale={code}` for a single locale or `locale=*` for all locales in the response.
- **Locale-based publishing** (Premium): Publish locales independently.
- **Translator role**: Built-in role that restricts editing to specific locale(s).
- **Pricing impact**: Free plan: 2 locales. Lite: 3 locales. Premium: custom.

## Media / Asset Management
- **Upload**: Assets uploaded via CMA or web app. Three fixed fields: title, description, file. File is locale-aware (different files per locale supported).
- **Asset size limits**: 50MB (Free), 1,000MB (paid plans).
- **Image API transformations**: Resize, crop (with face detection), format conversion (WebP, AVIF, JPEG, PNG, GIF, 8-bit PNG), quality adjustment (1-100%), background color, rounded corners, circle/ellipse crop, progressive JPEG.
- **CDN delivery**: All assets served via Fastly/CloudFront CDN. Images from `images.ctfassets.net`, other assets from `assets.ctfassets.net`.
- **Image editing in-app**: Basic image editing (crop, rotate).
- **Folders**: No traditional folder structure for assets. Organization via tags and taxonomy.
- **Tags**: Assets support up to 100 tags each.
- **MIME type validation**: Reference fields can restrict accepted asset types (image, video, audio, PDF, archive, code, etc.).
- **Embargoed assets** (Enterprise spaces): Restrict asset access with embargo rules.

## Authentication
- **Content Delivery API**: Authenticated via API keys (space-level). Two types: CDA access token and CPA preview token. Sent as `Authorization: Bearer {token}` header or `access_token` query parameter.
- **Content Management API**: Authenticated via either:
  - **Personal Access Tokens (PATs)**: User-scoped tokens with same permissions as the user. Can set expiration dates.
  - **OAuth 2.0 tokens**: For apps needing access to multiple users' data. Contentful implements OAuth provider; apps redirect users through Contentful's OAuth flow.
- **Images API**: No authentication required.
- **API key limits**: 100 per organization (Free/Lite), 200 per organization (paid).
- **SSO/SAML**: SAML 2.0 SSO with any compliant IdP (Okta, Azure AD, OneLogin, Ping Identity, Auth0, G Suite). Supports SP-initiated and IdP-initiated SSO. JIT user provisioning. SSO enforcement available. Premium plan only.
- **SCIM**: System for Cross-domain Identity Management protocol for automated user and team provisioning from IdP.
- **Two-factor authentication**: Not natively provided; handled by SSO identity provider if SSO is enforced.

## Authorization / RBAC

**Organization-level roles** (4 roles):

| Role | Org Settings | Spaces | Apps | Invoices |
|------|-------------|--------|------|----------|
| Owner | Full | Own spaces | Yes | Yes |
| Admin | Full (except invoices) | Own spaces | Yes | No |
| Developer | No (except Apps, Taxonomy) | Own spaces | Yes | No |
| Member | No | Own spaces | No | No |

**Space-level roles** (5 built-in + custom):
- **Admin**: Full control over space (content, settings, API keys, roles).
- **Editor**: Create, edit, publish, archive, delete all content (master environment).
- **Author**: Create and edit all content, cannot publish.
- **Translator**: Edit content in assigned locale(s) only.
- **Freelancer**: Create content and edit/delete own content only. Cannot publish. (Premium only).

**Custom roles** (Premium only):
- Whitelist approach -- all permissions denied by default, must be explicitly granted.
- Configure permissions across: Content, Media, Environments.
- **Allow and Deny rules**: Specify which actions, entry types, and content types are allowed or denied.
- **Tags-based access**: Restrict access based on entry/asset tags.
- **Field-level permissions**: Grant edit access to specific fields of specific content types.
- **Locale-level permissions**: Restrict editing to specific locales.
- **Content type scoping**: Restrict permissions to specific content types.
- Up to 250 custom roles per space.

## Web App (Admin UI)
- **Tech stack**: React-based single-page application. Design system: Forma 36 (open-source React component library).
- **Content modeling UI**: Visual editor for creating/editing content types. Drag-and-drop field reordering. JSON preview of content model.
- **Entry editor**: Customizable editor with field widgets. Default sidebar widgets: Status, Preview, Links, Translation, Versions, Users. Sidebar is reconfigurable.
- **Real-time collaboration**: Multiple editors can work on the same entry simultaneously with live presence indicators.
- **Search**: Full-text search across entries. Filter by content type, status, author, tags.
- **Customization**: Full customization via App Framework. Field editors, sidebar widgets, entry editors, and full-page apps can all be replaced with custom React apps running in sandboxed iframes.

## App Framework / Extensions
- **App Framework**: Apps are HTML/CSS/JS SPAs running in sandboxed iframes within the Contentful web app.
- **App locations** (where apps can render):
  - **Field**: Replace a field's editor widget.
  - **Sidebar**: Add widgets to the entry sidebar.
  - **Entry editor**: Replace the entire entry editing experience.
  - **Dialog**: Modal dialogs invoked by other app locations.
  - **Page**: Full-page apps accessible from the main navigation.
  - **Config**: App configuration screen.
- **App SDK**: JavaScript/TypeScript SDK for interacting with the Contentful web app from within iframe apps.
- **Forma 36**: Official React component library for building apps with consistent Contentful look-and-feel.
- **App Marketplace**: Curated marketplace of ready-to-install apps (Cloudinary, Netlify, Google Analytics, image focal point, AI tools, etc.).
- **App Identity**: Apps have their own identity for authenticated backend calls.
- **App Events**: Apps can subscribe to content lifecycle events.
- **App Actions**: Apps can expose callable actions that other apps, automations, and AI can invoke.
- **Contentful Functions**: Serverless functions running on Contentful's infrastructure.
- **App definitions**: 10 per organization (Free), 250 per organization (paid).
- **App installations**: 10 per environment (Free), 50 per environment (paid).
- **Open-source field editors**: All default field editor React components published as open source (`@contentful/field-editors`).

## Webhooks
- **Configuration**: Up to 20 webhooks per space (Free/Lite), 100 per space (paid). Configure via web app or CMA.
- **URL target**: Any HTTPS endpoint. Custom HTTP headers can be added. HTTP basic auth supported.
- **Content events**:
  - Entry: create, save, auto_save, archive, unarchive, publish, unpublish, delete
  - Asset: create, save, auto_save, archive, unarchive, publish, unpublish, delete
  - ContentType: create, save, publish, unpublish, delete
  - Task: create, save, delete
  - Release: (via action events)
- **Action events**: Release actions (create, execute), Scheduled actions (create, save, execute, delete), Bulk actions.
- **Filters**: Filter by Environment ID, Content Type ID, Entity ID, User ID (created by), User ID (updated by). Comparison operators: equal, not equal, in, not in, regexp, not regexp. Multiple filters per webhook (AND logic).
- **Webhook transformations**: Customize the payload body using JSONPath and Handlebars-like templates.
- **Signing/verification**: HMAC-SHA256 request signing. Signing secret generated per space. Verification function checks signature + timestamp with configurable TTL (default 30 seconds). Prevents replay attacks.
- **Idempotency**: `X-Contentful-Idempotency-Key` header (SHA256 hash) for deduplication.
- **Retries**: No automatic retries. Failed webhooks (including 30-second timeout) are not retried.
- **Activity log**: Up to 500 log entries per webhook (FIFO). Logs include full JSON request, response body, response status, and headers.
- **Templates**: Pre-built webhook templates for common integrations (Slack, Netlify, CircleCI, AWS, etc.).

## SDKs & Client Libraries

**Official SDKs** (by language):

| Language | CDA SDK | CMA SDK |
|----------|---------|---------|
| JavaScript/TypeScript | `contentful.js` (100% TypeScript, Node + browser) | `contentful-management.js` |
| Python | `contentful.py` | `contentful-management.py` |
| Ruby | `contentful.rb` | `contentful-management.rb` |
| Java | `contentful.java` | `contentful-management.java` |
| .NET | `contentful.net` (CDA + CMA in one) | (combined) |
| iOS/Swift | `contentful.swift` | -- |
| Android | `contentful.java` (Android variant) | -- |
| PHP | `contentful.php` | -- |

**Additional libraries**: `@contentful/rich-text-types`, `@contentful/rich-text-react-renderer`, `@contentful/rich-text-html-renderer`, `@contentful/content-source-maps`, `@contentful/live-preview`, `@contentful/experiences-sdk-react`.

## CLI
- **Package**: `contentful-cli` (npm).
- **Authentication**: OAuth browser flow or personal access token. `contentful login` opens browser for auth.
- **Space management**: `contentful space list`, `contentful space create`, `contentful space delete`, `contentful space use`.
- **Content export**: `contentful space export` -- exports to JSON. Options: `--include-drafts`, `--include-archived`, `--skip-content-model`, `--skip-content`, `--skip-roles`, `--skip-tags`, `--skip-webhooks`, `--skip-editor-interfaces`.
- **Content import**: `contentful space import` -- imports from JSON. Options: `--content-file`, `--skip-content-model`, `--skip-locales`, `--skip-content-publishing`.
- **Environment management**: `contentful space environment create`, `contentful space environment use`.
- **Environment aliases**: `contentful space alias create`, `contentful space alias update`.
- **Migration scripting**: `contentful space migration --space-id <id> <script.js>`. JavaScript migration scripts using the `contentful-migration` module.

## Deployment / Environments
- **Master environment**: Every space has a default "master" environment.
- **Sandbox environments**: Created by cloning any existing environment. Cloning is fast (< 1 minute). Copies content types, entries, assets, locales, tags, UI extensions. Does NOT copy: workflows, scheduled releases, tasks, space memberships.
- **Environment limits**: Up to 151 environments per space.
- **Environment creation throttle**: Max 12 environment creations per 5-minute window.
- **Environment aliases**: Pointers that resolve to a target environment. Retargeting an alias is near-instant (< 250ms). Enables blue-green deployment patterns.
- **Deployment workflow**: Clone master to dev, apply migrations, test, retarget master alias to new environment.
- **Environment isolation**: Each environment has independent content, content types, locales, tags, and app installations.

## Email
- **Limited native email**: Workflow step actions can send email notifications. System notifications (invitations, password resets) sent by Contentful.
- **No email templating/sending service**.
- **Integration options**: Mailgun app in Marketplace, webhook-based integration with email services, third-party automation (Zapier, Make, n8n).

## Backup & Restore
- **Export/Import**: `contentful space export` produces a JSON dump. `contentful space import` restores from JSON. Also available as npm libraries (`contentful-export`, `contentful-import`).
- **Environment cloning as backup**: Clone the master environment before making changes. No expiration on environments.
- **Automated backups**: No built-in scheduled backup. Recommended: AWS Lambda + CloudWatch running `contentful-export` on a schedule.
- **Deleted space/environment recovery**: Contentful can recover within 25 days. After 25 days, permanently lost.
- **Limitations**: Scheduled releases, tasks, workflows, and space memberships NOT included in export/import.

## Configuration
- **Space settings**: Name, description, space ID (immutable).
- **Environment settings**: Locales, content types, API keys (scoped to environment), webhook configurations.
- **API keys**: CDA and CPA access tokens managed per space. Multiple API keys per space, each targeting specific environments.
- **App configuration**: Each installed app has its own configuration page.
- **Organization-level settings**: Spaces, users, teams, apps, taxonomy manager, CMA tokens, usage, invoices, SSO/SCIM, audit logs.

## Observability / Logging
- **Usage dashboard**: Organization-level showing API call counts, CDN bandwidth, record counts, and asset bandwidth by space and time period.
- **API call tracking**: Usage metrics track CDA, CMA, CPA, and GraphQL calls against plan quotas.
- **Rate limit headers**: `X-Contentful-RateLimit-Second-Limit`, `X-Contentful-RateLimit-Second-Remaining`, `X-Contentful-RateLimit-Reset`.
- **Webhook activity logs**: Per-webhook request/response logs (up to 500 entries per webhook).
- **Status page**: `https://www.contentfulstatus.com/`
- **Contentful Analytics**: Content performance analytics integrating Google Analytics 4 data into the entry sidebar.
- **Function logs**: Serverless function execution logs retained for 30 days.
- **No built-in APM integration**.

## Audit Trail
- **Audit logs** (Premium/Enterprise): Capture all CMA actions across the organization. JSON events in OCSF (Open Cybersecurity Schema Framework) format.
- **Export destinations**: AWS S3, Azure Blob Storage, or Google Cloud Storage. Multi-cloud export supported.
- **Retention**: Controlled by the customer's storage configuration.
- **SIEM integration**: OCSF format enables integration with Splunk, Elastic, etc.
- **Not available on Free or Lite plans**.

## Content Migration
- **contentful-migration**: Official migration tool, integrated into CLI.
- **Migration scripts**: JavaScript files that export a function receiving a `migration` object.
- **Schema operations**: `createContentType()`, `editContentType()`, `deleteContentType()`, `createField()`, `editField()`, `deleteField()`, `changeFieldControl()`, `moveField()`.
- **Content transformation**: `transformEntries()` -- iterate, read, transform, write. `deriveLinkedEntries()` -- create new linked entries from existing data.
- **CMA access**: Migration scripts can call CMA directly via `migration.makeRequest()`.
- **Idempotency**: Not automatically tracked. Developers must implement their own tracking.

## Pricing / Licensing

| Feature | Free | Lite ($300/mo) | Premium (Custom) |
|---------|------|----------------|-----------------|
| Price | $0 | $300/month | Custom (~$60K+/year base) |
| Users | 10 | 20 | Custom |
| Roles | 2 (Admin, Editor) | 3 | Custom + Freelancer + Custom roles |
| Locales | 2 | 3 | Custom (up to 500) |
| API calls/month | 100K | 1M | Unlimited |
| CDN bandwidth | 50 GB/month | 100 GB/month | Custom |
| Max asset size | 50 MB | 50 MB | 1,000 MB |
| Spaces | 1 Starter | 1 Starter + 1 Lite | Unlimited |
| Records per space | 25 (Starter) | Up to 10,000 (Lite) | Custom (up to 5M) |
| Environments | Limited | Limited | Up to 151 |
| Scheduled publishing | No | Yes | Yes |
| Custom roles | No | No | Yes |
| SSO/SAML | No | No | Yes |
| Audit logs | No | No | Yes |
| Locale-based publishing | No | No | Yes |

**Free plan restriction**: "May only be used to test and learn about Contentful's product; it may not be used to support commercial use cases."

## Unique Features

**Contentful Studio / Experience Builder**:
- Visual drag-and-drop canvas for assembling web experiences using registered design system components.
- Experiences SDK (React, Next.js, Gatsby) for registering custom components, breakpoints, and design tokens.
- Built-in and custom components. Persona-driven: developers register, marketers assemble, editors bind content.

**AI Actions**:
- AI-powered content operations built into the platform.
- Models: GPT-4o, GPT-4o mini, Claude 3.5 Sonnet, and others.
- Pre-built templates: translation, content adaptation, alt text generation, proofreading, SEO optimization, tone rewriting, meta description generation, product description generation.
- Custom templates with user-defined instructions.
- Bulk AI actions: process up to 200 entries concurrently.
- Token limits: 128K tokens (GPT-4o), 200K tokens (Sonnet 3.5) per action instruction.

**Content Source Maps**:
- Metadata embedded in API responses linking content to its source fields in Contentful.
- Uses steganography (invisible Unicode characters) in text strings.
- Enables automatic "click to edit" links in preview/production.
- Works with Vercel Content Link for automatic inspector mode in Next.js. Premium plan only.

**Live Preview**:
- Real-time preview of content changes in your frontend application.
- Two modes: **Live Updates** (field changes reflected instantly) and **Inspector Mode** (click to jump to field in editor).
- SDK: `@contentful/live-preview` (React, Next.js).

**Contentful Functions**:
- Serverless functions running on Contentful's infrastructure.
- Types: filter, transformation, handler, GraphQL field resolvers, App Event handlers.
- Limits: 50 functions/app, 20M executions/org/month, 30s max execution, 128MB memory.

**Personalization**:
- AI-native personalization engine. Audience segmentation, A/B testing, conversion tracking.
- Data connections: Google Analytics, Segment, Contentsquare, Google Tag Manager.
- SOC 2 Type 2 compliance.

**Automation Builder**:
- No-code visual automation builder with triggers, conditions, and actions.
- Actions: AI Action, Slack, Microsoft Teams, Task, Email.
- Up to 100 automations per environment, 1,000 steps per execution.

**MCP Server**:
- Official MCP server for AI agent integration (`@contentful/mcp-server`).
- Remote hosted version (Beta): `https://mcp.contentful.com/mcp`.
- Enables AI agents to read/write content, manage models, work with assets, invoke AI Actions, publish content.

**Sync API**:
- Delta sync for keeping external data stores synchronized. Initial + incremental via sync tokens.

**Timeline Releases**:
- Schedule multiple future versions of the same entry with preview capability.
