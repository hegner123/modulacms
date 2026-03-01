# Sanity CMS Feature Inventory (Studio v3)

Comprehensive technical feature inventory of Sanity CMS, covering Sanity Studio v3 and the Content Lake platform. Research conducted February 2026 from official documentation and multiple independent sources.

---

## 1. Runtime Architecture

- **Hosted Content Lake**: All content is stored in Sanity's managed Content Lake (a cloud-hosted NoSQL document store). You cannot self-host the Content Lake; it is always managed by Sanity.
- **Self-hosted Studio**: Sanity Studio is a React-based SPA that compiles to static HTML/CSS/JS. It can be deployed to Sanity's own hosting (`sanity deploy`) or self-hosted on Vercel, Netlify, or any static file host. The Studio communicates with the Content Lake exclusively via HTTP APIs.
- **Two API endpoints**:
  - `api.sanity.io` -- live, uncached API; always returns the freshest data; used for mutations, listeners, and static builds.
  - `apicdn.sanity.io` -- CDN-distributed, cached API; opt-in; caches GET/HEAD/OPTIONS requests and POST requests to `/data/query` and `/graphql` (read-only endpoints). Responses cached until invalidated; stale content served for up to 2 hours if Content Lake is unavailable.
- **CDN infrastructure**: Sanity's API CDN has edge nodes in Mumbai, Sydney, Saint-Ghislain (Belgium), Sao Paulo, Oregon, Iowa, and Northern Virginia. An additional short-lived global CDN layer sits in front with points of presence on all continents. The short-lived global CDN does not cache private datasets or POST queries.
- **GROQ query engine**: GROQ (Graph-Relational Object Queries) is Sanity's proprietary query language. Queries are executed server-side in the Content Lake before results are returned. GROQ operates as a pipeline: data flows left-to-right through filters, projections, and ordering. Max query execution time: 1 minute.
- **Real-time sync**: Uses Server-Sent Events (SSE) protocol. Clients connect to `/data/listen/{dataset}?query=<GROQ>` and receive `welcome`, `mutation`, `channelError`, and `disconnect` events. Mutation events include `documentId`, `transition` (update/appear/disappear), `identity` (user), `mutations` array, full `result` document, revision IDs, and timestamp. Max listener connection lifetime: 30 minutes. Max concurrent listeners: 1,000 (Free), 5,000 (Growth), 10,000 (Enterprise).
- **Live Content API**: A specialized real-time API for rendering published content changes instantly on front ends. Listeners and the Live Content API are always routed to `api.sanity.io` (never cached).

## 2. Database / Content Lake

- **Document store**: The Content Lake is a schema-on-read NoSQL document store. Every piece of content is stored as a JSON document with system fields: `_id`, `_type`, `_rev`, `_createdAt`, `_updatedAt`.
- **No self-hosting**: The Content Lake is always Sanity-managed. You cannot run it on your own infrastructure.
- **NDJSON export format**: Datasets can be exported as `.tar.gz` archives containing all documents in a single NDJSON (Newline Delimited JSON) file plus asset files (images, files) in separate folders.
- **Document limits**:
  - Max documents per dataset: 10 million (customizable for Enterprise)
  - Max JSON document size: 32 MB
  - Max total JSON size in dataset: 10 GB
  - Max attributes per document: 1,000 (Free/Growth), 8,000 (Enterprise)
  - Max unique attributes per dataset: 2,000 (Free), 10,000 (Growth), custom (Enterprise)
  - Max attribute nesting depth: 20 levels
  - Max searchable term length: 1,024 UTF-8 characters
- **Real-time collaboration engine**: Edits are sent as patches (mutations) to the Content Lake in real-time. All patches are stored as transactions, forming the revision history. Multiple users can edit the same document simultaneously with presence indicators showing who is editing what. Changes auto-sync without manual saves.
- **Datasets**: A dataset is a collection of JSON documents (analogous to a database). Free plans get 2 public datasets; Growth gets 2 (public or private); Enterprise gets unlimited. Private datasets require authentication for all reads.
- **Transactions**: Mutations are atomic -- all mutations within a single API call either succeed or fail together. Transaction IDs are tracked for real-time listeners.

## 3. Content Management

- **Documents**: The fundamental unit of content. Each document has a `_type` field mapping to a schema definition, a ULID-based `_id`, and system metadata fields (`_rev`, `_createdAt`, `_updatedAt`).
- **CRUD**: Full create/read/update/delete via HTTP Mutation API. Mutations support `create`, `createOrReplace`, `createIfNotExists`, `patch` (set, unset, inc, dec, insert), and `delete` operations. Max mutation request body: 4 MB. Max mutation rate: 25 req/s. Max concurrent mutations: 100.
- **Portable Text**: Sanity's rich text format. An open specification (published on GitHub) that stores block content as a JSON array of typed objects. Blocks contain child spans with marks (decorators like bold/italic, or annotations like links with structured data). Fully extensible -- custom block types and inline objects can be embedded in the text stream. Serializer libraries exist for React, Vue, Svelte, and plain HTML.
- **References**: Typed references between documents using `_ref` (document ID) and `_type: "reference"`. Strong references prevent deletion of the referenced document. Weak references (`_weak: true`) allow deletion. Cross-dataset references available on Enterprise plans.
- **Ordering**: Documents in arrays can be ordered. Array items have `_key` fields for stable identity. GROQ supports `order()` for query-time sorting by any field.
- **Arrays**: First-class array type supporting arrays of any schema type (strings, objects, references, blocks, images, etc.). Array items are identified by `_key` for stable ordering and real-time collaboration.
- **Objects**: Inline structured data objects within documents. No separate document identity -- objects live within their parent document.
- **Image hotspot/crop**: Images support editor-defined hotspot (elliptical focal area) and crop (rectangular exclusion zones). Hotspot/crop metadata is stored per-use in the image field (not on the asset), so the same image can have different crops in different contexts. The `@sanity/image-url` library applies hotspot/crop data automatically when generating URLs.

## 4. Content Schema / Content Modeling

- **Code-first schema**: Schemas are defined in JavaScript or TypeScript files as plain objects using `defineType()`, `defineField()`, and `defineArrayMember()` helper functions for type safety. No GUI schema builder -- schema is code.
- **Schema types** (complete list):
  - `string` -- short text, optional list of predefined options
  - `number` -- numeric values
  - `text` -- multi-line text (textarea)
  - `boolean` -- true/false
  - `date` -- date only (no time)
  - `datetime` -- exact date and time
  - `slug` -- URL-safe string, auto-generated from another field
  - `url` -- URL values
  - `reference` -- typed reference to another document
  - `image` -- image upload with optional hotspot/crop, metadata extraction
  - `file` -- arbitrary file upload
  - `object` -- inline structured data (no document identity)
  - `array` -- ordered list of any other types
  - `block` -- Portable Text rich text block
  - `span` -- text span within a block (inline)
  - `geopoint` -- geographic coordinates (latitude, longitude, altitude)
  - `document` -- top-level content type with its own ID, revision history, and publishing state
  - `crossDatasetReference` -- reference to documents in other datasets (Enterprise)
  - `globalDocumentReference` -- reference to documents across projects
- **Validations**: Built-in validation methods per type (e.g., `required()`, `min()`, `max()`, `length()`, `regex()` for strings; `positive()`, `integer()` for numbers). Custom validation via `custom()` with async support. Validation rules can reference sibling fields via `rule.valueOfField()`. Document-level validation can access all fields. Validation can report errors (blocks publishing) or warnings (informational). `rule.skip()` for conditionally hidden fields.
- **Custom input components**: Full React component override for any field input. Uses the Field Component API for customizing title, description, validation display, and presence indicators. Custom components receive the field value and onChange callback.
- **Conditional fields**: Fields can be conditionally hidden or shown using the `hidden` property, which accepts a boolean or a callback function receiving the document and current user context. Conditional validation can skip rules for hidden fields using `rule.skip()`.
- **Fieldsets**: Fields can be grouped into collapsible fieldsets within a document type using the `fieldsets` property. Fieldsets can be set to collapsed by default.
- **Groups**: Schema fields can be organized into tab-like groups within the document editor for cleaner navigation.
- **Initial value templates**: Define default values for new documents. Templates can be parameterized (e.g., pre-set a category when creating from a filtered list).

## 5. Content Delivery / API

- **GROQ queries**: Primary query method. Syntax: `*[filter]{projection}`. Supports chaining filters, projections, ordering, slicing, and functions. Endpoint: `GET /v{date}/data/query/{dataset}?query=<GROQ>`.
- **GraphQL API**: Deployed per-dataset via `sanity graphql deploy`. Supports gen1/gen2/gen3 API generations (default gen3). Auto-generates types and filtering from schema. Supports pagination via `take` and `skip`. Supports GraphQL Playground. Not real-time; query-only.
- **Content Lake HTTP API**: RESTful API for mutations (`/data/mutate`), queries (`/data/query`), document retrieval (`/data/doc`), history, and listening (`/data/listen`). Versioned by date (e.g., `v2021-06-07`).
- **CDN API**: `apicdn.sanity.io` for cached reads. Caches GROQ query results and GraphQL responses. Max cached POST body: 300 KB. Responses >10 MB not cached. Non-200 responses not cached. Authenticated request caching segmented per token.
- **Filtering**: GROQ filters support equality, comparison, logical operators, string matching (`match`), array containment, `in` operator, `references()` function, `defined()`, `dateTime()` functions, and more. Filters can operate on nested fields, array items, and referenced documents.
- **Projections**: GROQ projections reshape response data. Can select specific fields, rename fields, compute derived values, inline referenced documents (joins), and construct arbitrary JSON structures. Reduces payload size and eliminates client-side data transformation.
- **Joins**: GROQ supports dereferencing references inline using the `->` operator. Follow references to include related document data in a single query. Supports chained dereferences and filtering within joins.
- **Pagination**: GROQ uses array slicing: `[0...10]` for offset-based pagination. Also supports keyset/cursor-based pagination using filters on sorted fields (e.g., `*[_type == "post" && _createdAt < $lastDate] | order(_createdAt desc) [0...10]`). GraphQL uses `take`/`skip` parameters.
- **Ordering**: GROQ `order()` function sorts by any field(s), ascending or descending. Multiple sort keys supported. Ordering is applied before slicing for efficient pagination.
- **Perspectives**: Query-level toggle to return either draft+published content (for previewing) or only published content (for production). Set via `perspective` parameter on queries.
- **Rate limits**: 500 req/s global API call rate per IP. 25 req/s mutation rate. 25 req/s upload rate. Max concurrent queries: 500. Max concurrent mutations: 100. CDN unlimited for cached responses.

## 6. Content Publishing

- **Draft/published model**: Every document can exist in two states simultaneously. The draft has an ID prefixed with `drafts.` (e.g., `drafts.abc123`); the published version has the bare ID (`abc123`). Editing creates/updates the draft; publishing copies the draft to the published document. Unauthenticated API requests only see published documents.
- **Live edit mode**: Per-type opt-in via `liveEdit: true` in schema. Changes are published immediately without going through the draft workflow.
- **Drafts can be disabled globally**: Set `document.drafts.enabled: false` in `sanity.config.ts` to disable all draft creation, limiting editing to live edit documents, API mutations, and Content Releases.
- **Scheduled Drafts**: Available on all paid plans (Growth+). Schedule a document to publish at a specific future date and time. Accessible from the document actions menu next to the publish button. Max scheduled publishing execution time: 1 minute.
- **Content Releases**: Enterprise-only add-on. Group multiple document changes into a single release that can be previewed, validated, scheduled, and published atomically. Release types: ASAP, At time (scheduled), and Undecided. Max 1,000 document versions per release. Max 100 MB total JSON per release. Releases batch in 10 MB sets. Rollback releases available. Preview overlapping releases available.
- **Workflow**: No built-in multi-step editorial workflow (draft -> review -> approve -> publish). The Contributor role can write drafts but not publish. Custom workflows can be built using document actions, custom plugins, or Content Releases.
- **Unpublish**: Documents can be unpublished (removing the published version while keeping the draft).

## 7. Content Versioning

- **Revision history**: Every edit is stored as a transaction (patch). Together, transactions form the complete revision history. Viewable in Studio via the History panel.
- **History retention**:
  - Free: 3 days
  - Growth: 90 days
  - Enterprise: 365 days (custom retention available)
  - Revisions older than the retention cutoff are truncated into a single revision item. Latest published and draft versions are always retained regardless of plan.
- **History events**: Each revision item shows: Published, Unpublished, Edited, or Truncated. Timestamps and user identity are recorded.
- **Restore**: You can restore a document to any previous revision within the retention window from the History panel.
- **Diff view**: Studio shows field-level diffs between revisions, highlighting what changed between any two points in the history.
- **Who changed what**: Each transaction records the identity of the user who made the change, viewable in the history timeline.
- **Full audit trail and History API**: Enterprise-only. Programmatic access to document history. Activity feed with event records: 90 days (Growth), 365 days (Enterprise).

## 8. Internationalization (i18n)

- **No built-in i18n system**: Localization is achieved through content modeling patterns and optional plugins. Sanity does not charge for locales; unlimited locales on all plans.
- **Document-level i18n** (via `@sanity/document-internationalization` plugin):
  - Creates a separate document per language
  - Each document has a `language` field
  - A `translation.metadata` document stores references linking all language versions together
  - Allows independent publishing per language
  - Best for documents where most/all fields differ by language, especially Portable Text
  - Provides Studio UI for creating translations and navigating between language versions
- **Field-level i18n** (via `sanity-plugin-internationalized-array` plugin):
  - Single document with localized fields stored as arrays (each item has `_key` = language code and `value` = content)
  - Uses fewer unique attributes than object-based approach (important given attribute limits)
  - Requires publishing all languages simultaneously
  - Best for documents with a mix of localized and shared fields
  - Custom UI renders each language input inline (no popup dialogs)
- **Object-based field localization**: Alternative pattern using an object with a field per language (e.g., `title.en`, `title.fr`). Simpler but consumes more unique attributes per language.
- **Language filter plugin**: `@sanity/language-filter` hides languages an editor does not need, cleaning up the Studio UI.
- **AI-powered translation**: The AI Assist plugin offers one-click translation between languages using LLMs.
- **Translation service adapters**: Official plugins for Transifex and Smartling integration.
- **GROQ querying**: Localized content can be queried with `coalesce()` for fallback chains and parameterized queries for dynamic locale selection.

## 9. Media / Asset Management

- **Image pipeline**: Globally distributed asset CDN serves images with on-demand transformations via URL parameters. Supports resize, crop, fit modes (clip, crop, fill, fillmax, max, scale, min), format conversion (jpg, png, webp, auto), quality adjustment, blur, flip, orientation, and DPR.
- **Image hotspot**: Elliptical focal area defined by editors. The pipeline uses the hotspot center as the focal point when cropping to different aspect ratios. Metadata stored per image field usage, not per asset.
- **Image crop**: Rectangular crop regions set by editors. Applied before hotspot-based focal cropping. Stored per image field usage.
- **Auto-format**: Images can be served in the optimal format for the requesting browser (e.g., WebP where supported) via `auto=format`.
- **Supported upload formats**: PNG, JPG/JPEG, BMP, GIF, TIFF, SVG, PSD, WebP, HEIF.
- **Asset limits**: Max image size: 256 megapixels. Max output dimensions from transforms: 8,192 pixels. Max animated image size with transforms: 256 megapixels (width x height x frame count). Max upload duration: 5 minutes (dataset assets), 1 hour (Media Library assets). Max upload rate: 25 req/s.
- **File uploads**: Arbitrary file type uploads with metadata storage. Files are referenced by asset documents.
- **Asset metadata**: Images automatically extract EXIF data, color palette, dimensions, and file size. Metadata stored in the image asset document.
- **Signed URLs**: Enterprise feature. Signed URLs include a cryptographic signature validating access and exact transformations, preventing unauthorized use and hotlinking.
- **Media Library**: Enterprise add-on. Centralized digital asset management across projects within an organization.
- **Media browser plugin**: Built-in media browser available on all plans for browsing and selecting existing assets.
- **Asset references**: Images and files are stored as separate asset documents. The image/file field stores a reference to the asset plus context-specific metadata (hotspot, crop, alt text). Same asset can be reused across documents with different per-use metadata.
- **@sanity/image-url**: Official JS library that generates CDN URLs from image references, automatically applying hotspot/crop data and transformation parameters.

## 10. Authentication

- **Token-based auth**: Sanity uses bearer tokens in the `Authorization` HTTP header for all authenticated API requests.
- **Studio auth**: Cookie-based authentication handled automatically by the Studio when users log in. Login providers: email/password, Google, GitHub.
- **Personal API tokens**: Generated on login via `sanity login`. Found via `sanity debug --secrets`. Last 1 year (shorter with SAML SSO). One per user per project.
- **Robot tokens (project-level)**: Created in project settings (Settings > API > Tokens). Scoped to a single project. Can be read-only (Viewer), read+write (Editor), or custom role. Displayed once on creation. Used for server-to-server integration.
- **Robot tokens (organization-level)**: Created in organization settings. Access to manage multiple projects, deploy SDK apps, or access organization-wide Sanity apps.
- **SAML SSO**: Available as Growth add-on ($1,399/month) or included with Enterprise. Supports any SAML 2.0 identity provider: Okta, Google Workspace, Azure AD/Entra ID, PingIdentity, OneLogin, etc. Role mapping from IdP groups to Sanity roles using pattern matching. When SSO is enabled, all users authenticate through the organization's IdP.
- **Third-party login**: Google and GitHub login for Studio access (no SAML required).
- **Custom authentication**: Enterprise feature. Replace Sanity's user database with your own authentication system (Active Directory, Kerberos, etc.). Sanity provides APIs to register users, manage permissions, and create sessions.
- **Token types**: Viewer (read-only), Editor (read+write), Deploy Studio (deploy only).

## 11. Authorization / RBAC

- **Default roles** (by plan):
  - Free: Administrator, Viewer (2 roles)
  - Growth: Administrator, Editor, Developer, Contributor, Viewer (5 roles)
  - Enterprise: All default roles + custom roles
- **Role descriptions**:
  - **Administrator**: Read/write all datasets, full project settings access
  - **Editor**: Read/write all datasets, limited project settings (can modify datasets but not create new ones)
  - **Developer**: Read/write all datasets, developer-level project settings access
  - **Contributor**: Read/write draft content only, cannot publish, no project settings access
  - **Viewer**: Read-only all datasets, no project settings (free on all plans, does not consume a paid seat)
- **Custom roles**: Enterprise only. Created via API (`POST /roles`). Attach pre-defined or custom permissions. Permission format: `{company}.{resourceType}.{objectName}.{action}`.
- **Document-level access control**: Enterprise only. Create custom permission resources using the `sanity.document.filter` schema with a GROQ filter to target specific document types or conditions. Custom document filter permissions include: create, read, update, manage, manageHistory, and history actions.
- **Filter-based access**: GROQ filters in custom permissions restrict which documents a role can access (e.g., only `_type == "post"` documents).
- **API for roles management**: Full REST API for creating, reading, updating, and deleting roles and permissions. `GET /permissions` to list available permission resources.
- **Member management**: Users are added via invitation or access request. A user can have multiple roles across projects in an organization. Max 1,000 users per project.

## 12. Sanity Studio (Admin UI)

- **React-based**: Built with React. Compiles to static files. Runs entirely client-side after loading.
- **Open source**: Studio source code is on GitHub (`sanity-io/sanity`).
- **Structure Builder**: API for customizing document browsing, list views, document views, and menu organization. Define custom navigation trees, filtered lists per content type, split panes, and custom document views.
- **Structure tool**: The default document management tool. Shows lists of documents organized by type, with detail panes for editing. Fully customizable via Structure Builder API.
- **Document views**: Configure custom views for documents (e.g., form, preview, JSON inspector, custom React components shown as tabs).
- **Custom tools**: Add entirely new top-level navigation items (tools) to the Studio. A tool is a React component registered in the Studio config.
- **Theming**: Studio supports custom themes via the `theme` property in `sanity.config.ts`. Control colors, typography, and branding.
- **Branding**: Customize Studio name, icon, and navigation branding via config.
- **Presentation tool**: Embeds a live preview of your actual website inside the Studio. Editors click elements in the preview to navigate to the corresponding content in the editing sidebar. Changes update in real-time.
- **Visual Editing**: Overlays on your rendered website that link content elements back to their source in the Studio. Powered by Content Source Maps. Uses `@sanity/overlays` library.
- **Real-time multiplayer editing**: Multiple users can edit the same document simultaneously. Presence indicators show who is editing which fields. Changes sync in real-time without conflicts.
- **Comments**: Available on Growth+ plans. Comment on specific fields within documents. Threaded discussions.
- **Task management**: Available on Growth+ plans. Assign tasks to team members within the Studio context.
- **Keyboard shortcuts**: Common keyboard shortcuts documented for navigation and editing.
- **Free hosting**: Sanity hosts your Studio at `{project-name}.sanity.studio` for free on all plans via `sanity deploy`.

## 13. Plugin / Extension System

- **Plugins**: Encapsulations of workspace config. A plugin can register schemas, tools, document actions, document badges, form components, and more. Installed via npm and added to the `plugins` array in `sanity.config.ts`.
- **Community plugin ecosystem**: Sanity Exchange (sanity.io/exchange) hosts community and official plugins. Categories include content management, media, localization, dashboards, and Studio customization.
- **Document actions**: Customize the actions available in the document action menu (publish, delete, unpublish, duplicate, etc.). Add custom actions or remove/reorder existing ones via the `document.actions` config callback.
- **Document badges**: Add visual badges/labels to documents in list views and the editor. Configured via `document.badges` callback.
- **Custom input components**: Replace any field's default input with a custom React component. Uses the Component API (input, field, item, preview).
- **Custom tools**: Register new top-level navigation tools in the Studio via the `tools` property.
- **Structure Builder customization**: Plugins can provide custom structure items, document views, and list configurations.
- **Notable official plugins**: `@sanity/vision` (GROQ query playground), `@sanity/document-internationalization`, `@sanity/scheduled-publishing` (deprecated in favor of built-in Scheduled Drafts), `@sanity/language-filter`, `@sanity/color-input`, `@sanity/code-input`, `@sanity/table`, AI Assist.

## 14. Webhooks

- **GROQ-powered webhooks**: Define outgoing HTTP requests triggered by Content Lake changes. Configured in project settings (sanity.io/manage), via CLI, or via API.
- **Trigger events**: Create, Update, Delete, or any combination.
- **GROQ filters**: Specify which documents trigger the webhook using GROQ filter syntax. Supports the `delta::` namespace to filter on document state before/after change. Supports `before()` and `after()` functions. Does not support sub-queries or cross-dataset reference dereferencing.
- **GROQ projections**: Define the webhook payload using GROQ projections. Shape the outgoing JSON body using document fields, `delta::` namespace values, and `before()`/`after()` functions. If empty, sends the entire document after the change.
- **HTTP methods**: POST, PUT, PATCH, DELETE, or GET.
- **Custom headers**: Add arbitrary HTTP headers (e.g., `Authorization: Bearer <token>`).
- **Secret signing**: Optional shared secret hashed and included in request headers. Follows the same standard as Stripe. Verified with the `@sanity/webhook` toolkit library.
- **Default headers**: `sanity-transaction-id`, `sanity-transaction-time`, `sanity-dataset`, `sanity-document-id`, `sanity-project-id`, `sanity-webhook-id`, `sanity-operation`, `idempotency-key`, `content-type: application/json`.
- **Idempotency**: Every webhook delivery includes an `idempotency-key` header for deduplication, following the IETF draft standard.
- **Retry policy**: HTTP 429 and 500-range responses retried with exponential back-off for up to 30 minutes. 400-range (except 429) treated as permanently undeliverable. Max 1 concurrent request per webhook. 30-second request timeout.
- **Draft/version filtering**: By default, `drafts.` and `versions.` prefixed documents are ignored. Can be opted in via settings.
- **Webhook limits**: Free: 2 webhooks. Growth: 4 webhooks. Enterprise: custom.
- **Attempts log**: API endpoint for viewing delivery attempts and message queue status.
- **Shareable configs**: Webhook configurations can be exported as shareable URLs.
- **API versioning**: Default `v2021-03-25`, overridable via API.

## 15. SDKs & Client Libraries

- **@sanity/client (JavaScript/TypeScript)**: Official client. Works in browsers, Node.js, Bun, Deno, and Edge Runtime. Supports GROQ queries, mutations, listeners, image URL generation, file uploads. Configurable `useCdn` toggle. Current version: v7.x.
- **PHP client**: Official Sanity PHP client for server-side PHP applications.
- **Rust client**: Community-supported Sanity Rust client.
- **LINQ (C#) client**: Community-supported Sanity LINQ client for .NET.
- **Flutter client**: Community-supported Sanity Flutter client.
- **No official Python, Ruby, or Swift SDK listed on docs page** (as of February 2026). Community or third-party clients may exist on npm/GitHub but are not listed in official documentation.
- **@sanity/image-url**: Library for generating image CDN URLs with transformations from image references.
- **@sanity/webhook**: Toolkit for verifying webhook signatures and handling webhook payloads.
- **Portable Text serializers**: Official libraries for React (`@portabletext/react`), Vue, Svelte, and HTML.
- **Framework integrations**:
  - `next-sanity` -- Next.js toolkit (live previews, visual editing, image components)
  - Nuxt module
  - Astro integration
- **Content Source Maps**: Open specification for annotating JSON fragments with metadata about their origin (field, document, dataset). Enables Visual Editing overlays. Returned by the GROQ API when requested.
- **Sanity MCP Server**: Official MCP (Model Context Protocol) server for connecting AI agents (Claude, Cursor, etc.) to Sanity content.

## 16. CLI

- **Installation**: `npm install -g sanity` or use via `npx sanity@latest [command]`.
- **Core commands**:
  - `sanity init` -- initialize a new Studio project
  - `sanity dev` -- start local development server (Vite-based)
  - `sanity build` -- build Studio for production
  - `sanity deploy` -- build and deploy Studio to Sanity hosting
  - `sanity undeploy` -- remove deployed Studio
  - `sanity start` -- alias for dev
  - `sanity login` / `sanity logout` -- authenticate CLI
- **Dataset management**:
  - `sanity dataset create <name>` -- create dataset
  - `sanity dataset delete <name>` -- delete dataset
  - `sanity dataset list` -- list datasets
  - `sanity dataset export <dataset> [file]` -- export to NDJSON + assets tarball
  - `sanity dataset import <file> <dataset>` -- import from export file
  - `sanity dataset copy <source> <target>` -- server-side copy (Enterprise)
- **GraphQL**:
  - `sanity graphql deploy` -- deploy GraphQL API from schema (options: `--dry-run`, `--tag`, `--dataset`, `--generation gen1|gen2|gen3`, `--playground`)
  - `sanity graphql list` -- list deployed APIs
  - `sanity graphql undeploy` -- remove deployed API
- **CORS**: `sanity cors add/delete/list` -- manage allowed origins
- **Documents**: `sanity documents get <id>` -- fetch a document, `sanity documents query <GROQ>` -- run a query
- **Migration**: `sanity migration create` -- scaffold a content migration, `sanity migration run` -- execute it (dry run by default, `--no-dry-run` to apply)
- **Backup** (Enterprise): `sanity backup enable/disable/list/download`
- **Webhooks**: `sanity hook create/delete/list` -- manage webhooks
- **Debug**: `sanity debug --secrets` -- show project info and auth tokens
- **Config**: `sanity.cli.js` or `sanity.cli.ts` for project-level CLI settings (projectId, dataset, server port, GraphQL config, Vite config).

## 17. Deployment

- **Studio deployment to Sanity hosting**: `sanity deploy` builds static files and uploads to `{name}.sanity.studio`. Free on all plans.
- **Self-hosted Studio**: Since Studio compiles to static HTML/CSS/JS, it can be hosted on Vercel, Netlify, AWS S3/CloudFront, or any static hosting. Automatic deploys via CI/CD when pushing to Git repositories.
- **Content Lake is always managed**: You cannot self-host or run the Content Lake. All data storage, querying, and real-time infrastructure is provided by Sanity's cloud.
- **No server-side runtime required**: Studio is purely client-side. All data operations go through Sanity's HTTP APIs.

## 18. Email

- **No built-in email system**: Sanity does not include any email sending, templating, or transactional email functionality. Email workflows require external services triggered via webhooks or custom integrations.

## 19. Backup & Restore

- **Manual export** (all plans): `sanity dataset export <dataset>` creates a `.tar.gz` containing NDJSON documents + assets. Can be imported to another dataset with `sanity dataset import`.
- **Managed backups** (Enterprise only):
  - Enabled via `sanity backup enable --dataset <name>`
  - Automatic daily backups
  - Stored offsite by Sanity for data redundancy
  - Daily backups retained for 365 days; weekly backups retained for an additional 2 years
  - Contains all documents and assets (but NOT comments or document history/timeline)
  - Download via `sanity backup download --dataset <name> --id <backup-id>`
  - List via `sanity backup list --dataset <name>` (default 30, max 100, supports `--after`/`--before` date filters)
  - Restore via `sanity dataset import <backup-file> <dataset>`
- **Dataset copy/clone**: Enterprise feature. Server-side `sanity dataset copy` without downloading locally. Cloud cloning available.
- **No automated backup on Free/Growth**: Must set up your own automation (e.g., GitHub Actions running `sanity dataset export` on a schedule).

## 20. Configuration

- **sanity.config.ts**: Central configuration file for Studio. Defines project ID, dataset, plugins, schema, document actions, document badges, tools, theme, and more. Uses `defineConfig()` helper.
- **sanity.cli.ts**: CLI-specific configuration. Defines project ID, dataset, dev server port, GraphQL deployment settings, and Vite config extensions.
- **Environment variables**: Must be prefixed with `SANITY_STUDIO_` to be available in Studio code (browser-side). Loaded from `.env`, `.env.local`, `.env.<mode>`, `.env.<mode>.local`. Mode determined by `SANITY_ACTIVE_ENV` (defaults to `development` for `sanity dev`, `production` for `sanity deploy`). Access via `process.env.SANITY_STUDIO_*`.
- **Workspaces**: A single Studio can serve multiple workspaces (different project/dataset combinations) from the same codebase. Configured as an array in `defineConfig()`.
- **API versioning**: API endpoints are versioned by date (e.g., `v2021-06-07`). Pinning to a date ensures stable behavior regardless of future API changes.

## 21. Observability / Logging

- **Usage dashboard**: Available in Sanity Manage (sanity.io/manage). Shows API requests, CDN requests, bandwidth, asset storage, and document counts relative to plan quotas.
- **Request logs**: API request logging for insights into content interaction and bandwidth usage. Basic logs on Free/Growth; Advanced logs on Enterprise.
- **Activity feed**: Enterprise feature. Event records with 90-day retention (Growth) or 365-day retention (Enterprise).
- **Datadog integration**: Official Sanity integration for forwarding activity logs to Datadog via webhooks.
- **Quota monitoring**: Plan usage tracked against limits. Free/Growth Trial plans do not allow overages (service blocked). Paid Growth plans auto-bill overages.
- **No built-in APM or tracing**: No server-side observability since the Content Lake is fully managed. Monitoring is limited to usage dashboards and request logs.

## 22. Audit Trail

- **Document history timeline**: Every edit is recorded as a transaction with timestamp and user identity. Viewable in Studio History panel.
- **Who changed what**: Each revision records the `identity` (user ID) of the author.
- **Transaction log**: Mutations include `transactionId` and `timestamp`. Available through real-time listeners and document history API.
- **Full audit trail**: Enterprise only. Comprehensive audit log of content changes and events. Custom retention periods.
- **Audit log export**: Enterprise plans can export logs to Google Cloud Storage as compressed NDJSON for SIEM integration.
- **SOC 2, GDPR, CCPA, PCI DSS compliance**: All plans include baseline compliance. Enterprise adds audit trails and custom controls.
- **Limitations**: Document history timeline is subject to retention limits (3/90/365 days by plan). Comments and timeline data are not included in backup exports.

## 23. Content Migration

- **@sanity/migrate tooling**: Built into the CLI. `sanity migration create` scaffolds a migration file. `sanity migration run` executes it. Always runs in dry-run mode unless `--no-dry-run` is specified.
- **Migration definitions**: Code-based migrations with a GROQ filter to select target documents and named helper functions that define transformations. Supports async functions for fetching external data.
- **GROQ-based transforms**: Migration filters use GROQ syntax (simple filters, no joins) to select documents for transformation.
- **Dataset export/import workflow**: Export dataset, modify NDJSON files (change `_type`, `_id`, update references), re-import. Suitable for type renaming and ID migration.
- **`sanity exec`**: Run arbitrary scripts against a dataset using the Sanity client. Useful for ad-hoc data transformations.
- **Schema evolution**: Schema-on-read architecture means schema changes do not require database migrations. Old documents remain valid until explicitly migrated. Studio handles undefined fields gracefully.
- **Dataset management**: `sanity dataset create/delete/list/export/import/copy`. Copy is server-side (Enterprise only).
- **Best practice**: Always export a backup before running migrations.

## 24. Pricing / Licensing

- **Free plan ($0/forever)**:
  - 20 user seats (Admin + Viewer roles only)
  - 2 public datasets
  - 10,000 documents
  - 2,000 unique attributes per dataset
  - 1,000,000 API CDN requests/month (note: pricing page says 1M, some docs say 5M -- check current pricing page)
  - 250,000 API requests/month
  - 100 GB bandwidth/month
  - 100 GB asset storage
  - 2 GROQ-powered webhooks
  - 3-day history retention
  - Content Agent (100 AI credits/month)
  - No overages allowed (blocked at limit)
- **Growth plan ($15/seat/month)**:
  - Up to 50 seats (Admin, Editor, Developer, Contributor, Viewer roles)
  - 2 datasets (public or private)
  - 25,000 documents
  - 10,000 unique attributes per dataset
  - 1,000,000 API CDN requests/month
  - 250,000 API requests/month
  - 100 GB bandwidth/month
  - 100 GB asset storage
  - 4 GROQ-powered webhooks
  - 90-day history retention
  - Comments and Tasks
  - Scheduled Drafts
  - AI Assist (100 AI credits/month)
  - Pay-as-you-go overages (CDN: $1/250k, API: $1/25k, Bandwidth: $0.30/GB, Assets: $0.50/GB)
  - Viewers are free (do not consume paid seats)
- **Growth add-ons**:
  - SAML SSO: $1,399/month
  - Dedicated Support: $799/month
  - Increased Quota: $299/month (extends to 50k docs, 5M CDN, 1M API, 500 GB bandwidth, 500 GB assets)
  - Extra Datasets: $999/dataset/month (up to 2 additional)
- **Enterprise (custom pricing)**:
  - Custom seat count
  - Custom roles and access control
  - Custom dataset count (unlimited)
  - SAML SSO included
  - Media Library (add-on)
  - Dedicated support with >99.9% uptime SLA
  - Onboarding program
  - Custom history retention
  - Custom usage quotas
  - Content Releases
  - Cloud cloning, managed backups, dataset hot swap
  - Cross-dataset references
  - Advanced dataset management
  - Full audit trail and History API
  - High Frequency CDN
  - Custom CDN domains
  - Enterprise SSO
  - Custom access control
  - Invoicing and organization-level billing
- **Gated features summary**:
  - Free-only exclusions: Comments, Tasks, Scheduled Drafts, AI Assist, Editor/Developer/Contributor roles, private datasets
  - Growth exclusions: Custom roles, Content Releases, SAML SSO (unless add-on), managed backups, cloud cloning, cross-dataset references, audit log export, custom CDN domains, High Frequency CDN
  - Enterprise-only: Custom roles, Content Releases, managed backups, full audit trail, custom access control, Media Library, enterprise SSO, cloud cloning

## 25. Unique Features

- **Real-time collaboration (Google Docs-style)**: Multiple users edit the same document simultaneously with live presence indicators, automatic conflict resolution, and instant sync. No save button -- changes are persisted as they are made. Cursor/field-level presence shows who is editing what.
- **Portable Text**: Open specification (GitHub: `portabletext/portabletext`) for structured rich text as JSON arrays. Framework-agnostic -- render in React, Vue, Svelte, HTML, or any custom renderer. Fully extensible with custom block types and inline objects. Separates content structure from presentation, unlike HTML-based rich text editors.
- **GROQ query language**: Declarative, pipeline-based query language purpose-built for JSON document stores. Combines the flexibility of GraphQL (ask for exactly what you need) with the simplicity of REST (single endpoint). Supports filtering, projection, ordering, slicing, joins via dereference (`->`), functions, and the `delta::` namespace for change-aware queries in webhooks.
- **Content Source Maps**: Open specification for annotating JSON API responses with metadata about content origin (field path, document ID, dataset). Enables Visual Editing overlays without any manual annotation by developers. Returned as part of GROQ API responses when requested.
- **Visual Editing**: Click-to-edit overlays on your production website. Powered by Content Source Maps and `@sanity/overlays`. Works with any front-end framework (Next.js, Nuxt, Astro, SvelteKit, etc.). Editors click any content element on the rendered page to jump directly to editing it in Studio.
- **Presentation tool**: Embeds a live, interactive preview of your website directly inside Sanity Studio. Side-by-side editing: make changes in the form, see them reflected immediately in the embedded preview. Click elements in the preview to navigate to the corresponding document/field in the editor.
- **AI Assist / Content Agent**: AI-powered content generation, translation, and transformation integrated into Studio. Content Agent understands the content schema and dataset. Can make bulk changes across hundreds of documents, staging everything for review. Agent Actions API allows programmatic AI-powered content workflows via API calls. MCP server enables AI agent (Claude, Cursor) integration. Credits: 100/month on Free/Growth, 500/month on Enterprise. Credit consumption: 4 per query, 2 per action.
- **Scheduled Publishing**: Two tiers: Scheduled Drafts (Growth+, single document) and Content Releases (Enterprise, multiple documents published atomically). Content Releases support ASAP, scheduled, and undecided timing.
- **Structure Builder**: Programmatic API for fully customizing the Studio's document management UI. Define custom navigation hierarchies, filtered views, split panes, and document type groupings. Used by plugins and project code alike.
- **Schema-on-read architecture**: Content Lake does not enforce schema on write. Schema definitions live in Studio code. Content models can evolve without database migrations or downtime. Old documents remain accessible and queryable.
- **Perspectives**: Query-level toggle between `published` (production) and `previewDrafts` (includes draft changes) views of content. Single query works for both production and preview without changing the query itself.
- **Compute and Functions**: Serverless functions that run in Sanity's infrastructure. Available on all plans (500K invocations/month included). Used for Agent Actions, custom AI workflows, and server-side content processing.

---

Sources:
- [Sanity Content Lake Docs](https://www.sanity.io/docs/content-lake)
- [GROQ Introduction](https://www.sanity.io/docs/content-lake/groq-introduction)
- [Schema Types](https://www.sanity.io/docs/studio/schema-types)
- [Validation](https://www.sanity.io/docs/studio/validation)
- [Localization](https://www.sanity.io/docs/studio/localization)
- [Drafts](https://www.sanity.io/docs/content-lake/drafts)
- [History Experience](https://www.sanity.io/docs/user-guides/history-experience)
- [Content Releases](https://www.sanity.io/docs/user-guides/content-releases)
- [Scheduled Drafts](https://www.sanity.io/docs/studio/scheduled-drafts-user-guide)
- [Authentication](https://www.sanity.io/docs/content-lake/http-auth)
- [Roles and Permissions](https://www.sanity.io/docs/content-lake/roles-concepts)
- [SAML SSO](https://www.sanity.io/docs/developer-guides/sso-saml)
- [Image Transformations](https://www.sanity.io/docs/apis-and-sdks/image-urls)
- [Presenting Images](https://www.sanity.io/docs/apis-and-sdks/presenting-images)
- [GROQ-Powered Webhooks](https://www.sanity.io/docs/content-lake/webhooks)
- [Libraries and Tooling](https://www.sanity.io/docs/libraries)
- [CLI](https://www.sanity.io/docs/apis-and-sdks/cli)
- [Studio Deployment](https://www.sanity.io/docs/studio/deployment)
- [Backups](https://www.sanity.io/docs/content-lake/backups)
- [API CDN](https://www.sanity.io/docs/content-lake/api-cdn)
- [Real-time Updates](https://www.sanity.io/docs/content-lake/realtime-updates)
- [Technical Limits](https://www.sanity.io/docs/content-lake/technical-limits)
- [Pricing](https://www.sanity.io/pricing)
- [Visual Editing](https://www.sanity.io/docs/visual-editing/introduction-to-visual-editing)
- [Content Source Maps](https://www.sanity.io/docs/visual-editing/content-source-maps)
- [Content Agent](https://www.sanity.io/content-agent)
- [Schema and Content Migrations](https://www.sanity.io/docs/content-lake/schema-and-content-migrations)
- [Portable Text Specification](https://github.com/portabletext/portabletext)
- [Structure Builder](https://www.sanity.io/docs/studio/structure-introduction)
- [Studio Customization](https://www.sanity.io/docs/studio/studio-customization)
- [Environment Variables](https://www.sanity.io/docs/studio/environment-variables)
- [Paginating with GROQ](https://www.sanity.io/docs/developer-guides/paginating-with-groq)
- [Request Logs](https://www.sanity.io/docs/platform-management/request-logs)
- [Spring Release 2025](https://www.sanity.io/blog/what-s-new-may-2025)
- [Sanity Pricing Comparison (Webstacks)](https://www.webstacks.com/blog/sanity-pricing-plans-compared)
