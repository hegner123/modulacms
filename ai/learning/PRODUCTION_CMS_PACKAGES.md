# Production CMS Package Comparison

**Purpose:** Compare ModulaCMS packages with production CMS systems (WordPress, Umbraco, Laravel CMS)
**Created:** 2026-01-16
**Status:** Gap analysis and learning resource

---

## How to Use This Document

This document shows what packages/features production CMS systems have that ModulaCMS currently lacks. Each section includes:
- What it is and why it exists
- Which CMSs have it
- Whether ModulaCMS needs it (and why/why not)
- Implementation complexity if you decide to add it

---

## ModulaCMS Current Packages (21 total)

✅ **What ModulaCMS Has:**

### Core Infrastructure
- `auth` - OAuth authentication
- `config` - Configuration management
- `middleware` - HTTP middleware (CORS, sessions, cookies)
- `router` - HTTP routing
- `utility` - Logging and utilities

### Data Layer
- `db` - Database abstraction interface
- `db-sqlite` - SQLite driver
- `db-mysql` - MySQL driver
- `db-psql` - PostgreSQL driver
- `model` - Data models and business logic

### Content Management
- `cli` - Terminal UI for content management
- `media` - Image processing and optimization
- `backup` - Database and media backup/restore

### Storage & Integration
- `bucket` - S3-compatible object storage
- `plugin` - Lua plugin system

### Deployment & Operations
- `install` - Installation wizard
- `update` - Self-update system
- `deploy` - Deployment utilities
- `file` - File operations
- `flags` - Command-line flags

---

## Missing Packages: Content Management

### 1. Content Versioning/Revisions

**What:** Track every change to content, allow reverting to previous versions
**Examples:** WordPress revisions, Umbraco version history, Git-like content versioning

**Why CMSs have it:**
- Undo mistakes (accidental deletion, bad edits)
- Content auditing and compliance
- Compare versions side-by-side
- Restore from any point in history

**Typical features:**
- Automatic version creation on save
- Manual version snapshots
- Diff viewing (what changed?)
- Restore to previous version
- Version pruning (keep last N versions)
- Author and timestamp tracking

**Do you need it?**
- **Must-have:** If content is mission-critical or legally regulated
- **Nice-to-have:** For multi-user content teams
- **Skip:** For simple sites with few editors

**Implementation complexity:** Medium
- Database: Add `content_versions` table
- Store full content snapshots or diffs (JSON patches)
- UI: Version timeline, restore buttons
- Storage consideration: Versions multiply storage needs

**Example table structure:**
```sql
CREATE TABLE content_versions (
    id INTEGER PRIMARY KEY,
    content_data_id INTEGER NOT NULL,
    version_number INTEGER NOT NULL,
    content_snapshot JSON NOT NULL,  -- Full content or diff
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_message TEXT,
    FOREIGN KEY (content_data_id) REFERENCES content_data(id)
);
```

---

### 2. Workflow/Publishing States

**What:** Content lifecycle management (draft → review → scheduled → published)
**Examples:** WordPress publish/draft/pending, Umbraco content workflow, custom approval flows

**Why CMSs have it:**
- Editorial review process
- Content approval workflows
- Scheduled publishing (future dates)
- Unpublish/archive content
- Prevent accidental publication

**Typical states:**
- **Draft** - Work in progress, not public
- **Pending Review** - Submitted for approval
- **Scheduled** - Publish at future date/time
- **Published** - Live and public
- **Archived** - No longer active but preserved

**Do you need it?**
- **Must-have:** Editorial teams with approval process
- **Nice-to-have:** Scheduled content (blog posts, announcements)
- **Skip:** Single-author sites or always-published content

**Implementation complexity:** Medium
- Database: Add `status` and `publish_date` columns to `content_data`
- Logic: Filter queries by status and date
- Cron: Scheduled publishing job (check `publish_date`)
- Permissions: Who can publish vs draft?

**Example fields:**
```sql
ALTER TABLE content_data ADD COLUMN status VARCHAR(20) DEFAULT 'draft';
ALTER TABLE content_data ADD COLUMN publish_date TIMESTAMP;
ALTER TABLE content_data ADD COLUMN unpublish_date TIMESTAMP;
```

---

### 3. Search & Indexing

**What:** Full-text search across content with ranking and relevance
**Examples:** WordPress search, Elasticsearch integration, Algolia, Meilisearch

**Why CMSs have it:**
- Users need to find content quickly
- Filter and faceted search
- Search across multiple fields (title, body, metadata)
- Relevance ranking (best matches first)

**Typical features:**
- Full-text search (match words in content)
- Fuzzy matching (typo tolerance)
- Faceted search (filter by category, date, author)
- Search suggestions (autocomplete)
- Search analytics (what people search for)
- Highlighting (show matching text snippets)

**Do you need it?**
- **Must-have:** Large content libraries (100+ pages)
- **Nice-to-have:** Any site with search functionality
- **Skip:** Small sites where browsing is sufficient

**Implementation complexity:** Low to High
- **Low:** SQLite/MySQL full-text search (built-in, limited features)
- **Medium:** PostgreSQL full-text search (better than SQLite/MySQL)
- **High:** Elasticsearch/Meilisearch integration (best features, separate service)

**SQLite example:**
```sql
-- Create full-text search virtual table
CREATE VIRTUAL TABLE content_fts USING fts5(
    title,
    body,
    content_data_id
);

-- Search query
SELECT * FROM content_fts WHERE content_fts MATCH 'search term';
```

**Package structure:**
```
internal/search/
├── search.go          # Search interface
├── sqlite_fts.go      # SQLite FTS5 implementation
├── postgres_fts.go    # PostgreSQL full-text search
└── elasticsearch.go   # Optional: Elasticsearch integration
```

---

### 4. Import/Export

**What:** Bulk content migration, data portability
**Examples:** WordPress WXR export, CSV import, JSON API export

**Why CMSs have it:**
- Migrate from other platforms
- Bulk content operations
- Data portability
- Backup in portable format
- Integration with external tools

**Typical formats:**
- **CSV** - Simple tabular data
- **JSON** - Structured data with relationships
- **XML** - WordPress WXR format
- **Markdown** - Content-only format
- **SQL dumps** - Database-level export (ModulaCMS has via backup)

**Do you need it?**
- **Must-have:** If migrating from existing CMS
- **Nice-to-have:** For backup/restore in portable format
- **Skip:** Greenfield projects with no migration needs

**Implementation complexity:** Medium
- Export: Query database, serialize to format
- Import: Parse format, validate, insert to database
- Mapping: Handle field differences between systems
- Relationships: Preserve parent/child, references

**Package structure:**
```
internal/import/
├── importer.go        # Import interface
├── csv.go            # CSV import
├── json.go           # JSON import
├── wordpress.go      # WordPress WXR import
└── validation.go     # Import validation

internal/export/
├── exporter.go       # Export interface
├── csv.go           # CSV export
├── json.go          # JSON export
└── api.go           # JSON API format
```

---

## Missing Packages: User Experience

### 5. Comments/Discussion System

**What:** User-generated comments on content
**Examples:** WordPress comments, Disqus integration, moderation system

**Why CMSs have it:**
- Community engagement
- User feedback
- Discussion threads
- Social proof

**Typical features:**
- Threaded replies (comment hierarchy)
- Moderation (approve/reject/spam)
- User notifications (reply alerts)
- Rich text or Markdown support
- Voting (upvote/downvote)
- Spam filtering (Akismet-style)

**Do you need it?**
- **Must-have:** Community sites, blogs with engagement
- **Nice-to-have:** Marketing sites
- **Skip:** Headless CMS (let client implement), corporate sites

**Implementation complexity:** Medium
- Database: `comments` table with parent/child relationships
- Moderation: Status field, admin approval
- Notifications: Email alerts (requires email package)
- Spam: Integrate third-party service or simple rules

**ModulaCMS consideration:**
Since ModulaCMS is headless, comments might be better in the client application. However, an API for managing comments would be useful.

---

### 6. Menu/Navigation Builder

**What:** Visual menu editor for site navigation
**Examples:** WordPress menus, drag-and-drop menu builders

**Why CMSs have it:**
- Non-technical users build navigation
- Reorder menu items easily
- Multi-level menus (dropdowns)
- Different menus for different areas (header, footer, sidebar)

**Typical features:**
- Drag-and-drop reordering
- Nested menu items (hierarchy)
- Menu locations (header, footer)
- Custom links (external URLs)
- Menu item metadata (icon, description)

**Do you need it?**
- **Must-have:** Non-technical content editors
- **Nice-to-have:** If menus change frequently
- **Skip:** Developer-managed navigation, single menu structure

**Implementation complexity:** Low to Medium
- Database: `menus` and `menu_items` tables
- Tree structure: Similar to content tree (parent_id)
- API: CRUD operations for menus
- UI: Drag-and-drop in TUI is hard; better as client feature

**ModulaCMS consideration:**
Could use existing tree structure for menus, just add a `menu` route type.

---

### 7. Taxonomy (Tags/Categories)

**What:** Flexible classification beyond hierarchical tree
**Examples:** WordPress tags and categories, taxonomy vocabularies

**Why CMSs have it:**
- Multiple classification schemes (tags AND categories)
- Many-to-many relationships (post has multiple tags)
- Faceted filtering (find by tag + category)
- Organizational flexibility

**Typical features:**
- Categories (hierarchical, single-select)
- Tags (flat, multi-select)
- Custom taxonomies (locations, products, etc.)
- Taxonomy archives (all posts with tag "Go")

**Do you need it?**
- **Must-have:** Content needs multiple classification schemes
- **Nice-to-have:** Blogs, product catalogs
- **Skip:** Simple hierarchical content (tree is enough)

**Implementation complexity:** Medium
- Database: `taxonomies`, `terms`, `content_terms` junction table
- Queries: Join to find content by term
- UI: Tag picker, category selector

**ModulaCMS consideration:**
Current tree structure is hierarchical only. Tags would require new tables:

```sql
CREATE TABLE taxonomies (
    id INTEGER PRIMARY KEY,
    name VARCHAR(100),      -- "tags", "categories", "locations"
    slug VARCHAR(100),
    hierarchical BOOLEAN    -- Can terms have children?
);

CREATE TABLE terms (
    id INTEGER PRIMARY KEY,
    taxonomy_id INTEGER NOT NULL,
    name VARCHAR(100),
    slug VARCHAR(100),
    parent_id INTEGER,      -- For hierarchical taxonomies
    FOREIGN KEY (taxonomy_id) REFERENCES taxonomies(id)
);

CREATE TABLE content_terms (
    content_data_id INTEGER NOT NULL,
    term_id INTEGER NOT NULL,
    PRIMARY KEY (content_data_id, term_id)
);
```

---

## Missing Packages: Technical Infrastructure

### 8. Cache System

**What:** Store computed results to avoid expensive operations
**Examples:** Object cache, page cache, Redis, Memcached

**Why CMSs have it:**
- Reduce database queries
- Faster page loads
- Handle traffic spikes
- Reduce server load

**Typical cache types:**
- **Object cache** - Database query results, computed data
- **Page cache** - Full rendered pages (not applicable to headless)
- **API response cache** - Cache JSON responses
- **Fragment cache** - Cache parts of responses

**Do you need it?**
- **Must-have:** High traffic sites (1000+ requests/min)
- **Nice-to-have:** Any production site
- **Skip:** Development, low-traffic sites

**Implementation complexity:** Low to Medium
- **Low:** In-memory cache (Go map with TTL)
- **Medium:** Redis/Memcached integration

**Package structure:**
```
internal/cache/
├── cache.go          # Cache interface
├── memory.go         # In-memory cache (go map)
├── redis.go          # Redis implementation
└── noop.go          # No-op cache for development
```

**Interface example:**
```go
type Cache interface {
    Get(key string) ([]byte, error)
    Set(key string, value []byte, ttl time.Duration) error
    Delete(key string) error
    Clear() error
}
```

---

### 9. Queue/Background Jobs

**What:** Process long-running tasks asynchronously
**Examples:** Image processing, email sending, imports, exports

**Why CMSs have it:**
- Don't block HTTP requests
- Retry failed operations
- Process in parallel
- Schedule tasks

**Typical use cases:**
- Image optimization (resize, compress)
- Send bulk emails
- Generate reports
- Import large datasets
- Video transcoding
- Webhook delivery

**Do you need it?**
- **Must-have:** Long-running operations (>5 seconds)
- **Nice-to-have:** Any background processing
- **Skip:** Simple sites with fast operations

**Implementation complexity:** Medium
- **Simple:** Goroutines with channels
- **Production:** Persistent queue (Redis, database-backed)
- Libraries: `github.com/hibiken/awork`, `github.com/gocraft/work`

**Package structure:**
```
internal/queue/
├── queue.go          # Queue interface
├── memory.go         # In-memory queue (goroutines)
├── redis.go          # Redis-backed queue
└── worker.go         # Worker pool implementation
```

---

### 10. Scheduled Tasks/Cron

**What:** Run tasks at specific times or intervals
**Examples:** Publish scheduled content, cleanup old data, send digests

**Why CMSs have it:**
- Scheduled publishing
- Periodic maintenance
- Batch operations
- Time-based automation

**Typical tasks:**
- Publish scheduled content (check `publish_date`)
- Delete old sessions
- Generate sitemaps
- Send email digests
- Backup database
- Cleanup temp files

**Do you need it?**
- **Must-have:** Scheduled publishing, periodic tasks
- **Nice-to-have:** Automation, maintenance
- **Skip:** Simple sites with no time-based operations

**Implementation complexity:** Low to Medium
- **Simple:** Goroutine with `time.Ticker`
- **Better:** Cron library `github.com/robfig/cron`
- **Production:** External cron calls HTTP endpoint

**Package structure:**
```
internal/scheduler/
├── scheduler.go      # Scheduler interface
├── cron.go          # Cron implementation
└── tasks/           # Predefined tasks
    ├── publish.go   # Publish scheduled content
    ├── cleanup.go   # Delete old data
    └── backup.go    # Automated backups
```

**Example:**
```go
// Run every minute to check for content to publish
scheduler.AddFunc("* * * * *", func() {
    publishScheduledContent()
})
```

---

### 11. Email/Notifications

**What:** Send emails and notifications to users
**Examples:** SMTP integration, transactional email (SendGrid, Mailgun)

**Why CMSs have it:**
- User registration emails
- Password reset
- Comment notifications
- Content published alerts
- Newsletter/digest emails

**Typical features:**
- SMTP server configuration
- Email templates (HTML + plain text)
- Transactional email service integration
- Email queue (don't block requests)
- Bulk email (newsletters)
- Email tracking (opens, clicks)

**Do you need it?**
- **Must-have:** User registration, password reset
- **Nice-to-have:** Notifications, alerts
- **Skip:** Headless API-only CMS

**Implementation complexity:** Medium
- SMTP: Standard library `net/smtp`
- Templates: `html/template` for emails
- Queueing: Background job system (see #9)
- Services: SendGrid, Mailgun, AWS SES SDKs

**Package structure:**
```
internal/email/
├── email.go          # Email interface
├── smtp.go           # SMTP implementation
├── sendgrid.go       # SendGrid service
├── templates/        # Email templates
│   ├── welcome.html
│   └── reset.html
└── queue.go          # Email queue
```

---

### 12. Webhooks

**What:** HTTP callbacks when events occur
**Examples:** Content published → notify external service, new user → send to CRM

**Why CMSs have it:**
- Integrate with external services
- Event-driven architecture
- Decouple systems
- Automation

**Typical events:**
- Content created/updated/deleted
- User registered
- Comment posted
- Media uploaded
- Form submitted

**Typical features:**
- Configure webhook URLs
- Event subscriptions (choose which events)
- Retry failed webhooks
- Webhook signatures (HMAC validation)
- Delivery logs

**Do you need it?**
- **Must-have:** Integration with external services
- **Nice-to-have:** Extensibility, automation
- **Skip:** Standalone systems with no integrations

**Implementation complexity:** Medium
- Database: `webhooks` and `webhook_deliveries` tables
- Delivery: Background jobs to POST events
- Retry: Exponential backoff
- Security: HMAC signatures

**Package structure:**
```
internal/webhooks/
├── webhook.go        # Webhook management
├── delivery.go       # Webhook delivery
├── events.go         # Event definitions
└── signatures.go     # HMAC signing
```

---

## Missing Packages: SEO & Marketing

### 13. SEO Management

**What:** Search engine optimization tools and metadata
**Examples:** Meta tags, Open Graph, Twitter Cards, sitemaps

**Why CMSs have it:**
- Search engine visibility
- Social media previews
- Structured data (Schema.org)
- Analytics integration

**Typical features:**
- Meta title and description per page
- Open Graph tags (Facebook, LinkedIn)
- Twitter Card tags
- Canonical URLs
- Robots.txt management
- XML sitemap generation
- Schema.org structured data
- Redirect management

**Do you need it?**
- **Must-have:** Public-facing content sites
- **Nice-to-have:** Marketing sites
- **Skip:** Internal apps, APIs

**Implementation complexity:** Low to Medium
- Database: Add SEO fields to `content_data`
- Sitemap: Generate XML from content tree
- API: Return SEO metadata with content

**ModulaCMS consideration:**
As a headless CMS, SEO data should be in the API response. Client implements meta tags.

**Example fields:**
```sql
ALTER TABLE content_data ADD COLUMN meta_title VARCHAR(255);
ALTER TABLE content_data ADD COLUMN meta_description TEXT;
ALTER TABLE content_data ADD COLUMN og_image VARCHAR(500);
ALTER TABLE content_data ADD COLUMN canonical_url VARCHAR(500);
ALTER TABLE content_data ADD COLUMN robots VARCHAR(50);  -- index,follow
```

---

### 14. Redirects Management

**What:** URL redirect rules (301, 302)
**Examples:** Old URL → new URL, broken link management

**Why CMSs have it:**
- Preserve SEO when URLs change
- Fix broken links
- URL restructuring
- Domain changes

**Typical features:**
- Redirect rules (source → destination)
- Redirect types (301 permanent, 302 temporary)
- Wildcard redirects (pattern matching)
- Redirect chains detection
- Bulk import redirects

**Do you need it?**
- **Must-have:** URL structure changes, migrations
- **Nice-to-have:** SEO-focused sites
- **Skip:** New sites with stable URLs

**Implementation complexity:** Low
- Database: `redirects` table
- Middleware: Check incoming URL against redirects
- Response: Return 301/302 with Location header

**Package structure:**
```
internal/redirects/
├── redirects.go      # Redirect management
├── matcher.go        # URL pattern matching
└── middleware.go     # HTTP redirect middleware
```

**Example table:**
```sql
CREATE TABLE redirects (
    id INTEGER PRIMARY KEY,
    source_url VARCHAR(500) NOT NULL,
    destination_url VARCHAR(500) NOT NULL,
    status_code INTEGER DEFAULT 301,  -- 301 or 302
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Missing Packages: Access Control & Security

### 15. Granular Permissions (ACL)

**What:** Fine-grained access control beyond roles
**Examples:** User can edit own posts, permissions per content item

**Why CMSs have it:**
- Complex permission requirements
- Per-resource permissions
- Organizational hierarchy (departments, teams)
- Delegation (assign editor to specific pages)

**Typical patterns:**
- **Role-based** (ModulaCMS has basic roles)
- **Permission-based** (create, read, update, delete permissions)
- **Resource-based** (permissions per content item)
- **Attribute-based** (conditions: if author == user)

**Do you need it?**
- **Must-have:** Multi-tenant, complex organizations
- **Nice-to-have:** Medium-size content teams
- **Skip:** Small teams, simple role structure

**Implementation complexity:** High
- Database: `permissions`, `user_permissions`, `role_permissions`
- Logic: Check permissions on every operation
- Inheritance: User permissions + role permissions
- Caching: Permission lookups are frequent

**ModulaCMS current state:**
Has `permissions` and `roles` tables, but basic implementation.

---

### 16. Audit Trail

**What:** Log all actions for security and compliance
**Examples:** Who edited what, when, and from where

**Why CMSs have it:**
- Security investigations
- Compliance (HIPAA, SOC2, GDPR)
- Change tracking
- User accountability

**What to log:**
- User actions (login, logout, create, update, delete)
- Resource accessed
- Timestamp
- IP address
- User agent
- Before/after values (for updates)
- Success/failure

**Do you need it?**
- **Must-have:** Regulated industries, compliance requirements
- **Nice-to-have:** Security-conscious applications
- **Skip:** Small personal sites

**Implementation complexity:** Medium
- Database: `audit_log` table (can grow large)
- Logging: Intercept all operations
- Storage: Time-series database or log aggregation
- Querying: Fast searches by user, resource, date

**Package structure:**
```
internal/audit/
├── audit.go          # Audit logging
├── logger.go         # Audit event logger
└── query.go          # Audit log queries
```

---

## Missing Packages: Media & Assets

### 17. Advanced Image Transformations

**What:** On-demand image manipulation (resize, crop, filters)
**Examples:** WordPress image sizes, Cloudinary, imgproxy

**Why CMSs have it:**
- Responsive images (different sizes for different devices)
- Thumbnails
- Cropping and focal point
- Image optimization (WebP conversion)
- Filters (grayscale, blur)

**ModulaCMS current state:**
Has basic image optimization in `internal/media` (scaling, cropping).

**What's missing:**
- On-demand transformations (URL parameters: `/image.jpg?w=800&h=600`)
- Advanced crops (focal point, smart crop)
- Filters and effects
- Format conversion (automatic WebP)
- Art direction crops (different crops per breakpoint)

**Do you need it?**
- **Must-have:** Image-heavy sites, responsive design
- **Nice-to-have:** Better performance, bandwidth savings
- **Skip:** Text-heavy sites, simple layouts

**Implementation complexity:** Medium to High
- **Medium:** URL-based transformations with caching
- **High:** Full image manipulation service (Cloudinary-like)

---

### 18. Video Processing

**What:** Video upload, transcoding, streaming
**Examples:** Automatic transcoding, multiple formats, HLS streaming

**Why CMSs have it:**
- Video content management
- Multiple format support (MP4, WebM)
- Adaptive streaming (HLS, DASH)
- Thumbnail generation
- Compression

**Do you need it?**
- **Must-have:** Video-heavy sites
- **Nice-to-have:** Occasional video content
- **Skip:** No video content, use YouTube/Vimeo

**Implementation complexity:** High
- External services: AWS MediaConvert, Mux, Cloudflare Stream
- Self-hosted: FFmpeg integration (complex)
- Storage: Large file handling
- Streaming: HLS/DASH serving

**ModulaCMS consideration:**
Probably better to integrate with external video services (Mux, Vimeo) than build in-house.

---

## Missing Packages: Developer Experience

### 19. API Documentation

**What:** Auto-generated API docs
**Examples:** OpenAPI/Swagger, GraphQL introspection

**Why CMSs have it:**
- Developer onboarding
- API discoverability
- Testing interface (Swagger UI)
- Client SDK generation

**Do you need it?**
- **Must-have:** Public API, third-party integrations
- **Nice-to-have:** Internal development
- **Skip:** Private, internal-only API

**Implementation complexity:** Low to Medium
- OpenAPI: `github.com/swaggo/swag` (annotations → OpenAPI spec)
- GraphQL: Built-in introspection
- Docs UI: Swagger UI, Redoc, GraphQL Playground

---

### 20. Form Builder

**What:** Create custom forms without code
**Examples:** Contact forms, surveys, lead capture

**Why CMSs have it:**
- Non-technical form creation
- Flexible field types
- Validation rules
- Email notifications
- Data collection

**Do you need it?**
- **Must-have:** Marketing sites, lead generation
- **Nice-to-have:** User feedback, surveys
- **Skip:** Fixed forms, developer-managed forms

**Implementation complexity:** High
- Database: Dynamic form definitions
- Rendering: Generate form HTML from definition
- Validation: Dynamic rules
- Storage: Form submissions

**ModulaCMS consideration:**
As a headless CMS, forms might be better handled by the client. But a form data storage API could be useful.

---

## Missing Packages: Analytics & Reporting

### 21. Analytics Integration

**What:** Track content performance, user behavior
**Examples:** Google Analytics, Matomo, custom analytics

**Why CMSs have it:**
- Content performance metrics
- User behavior tracking
- Conversion tracking
- Popular content reports

**Do you need it?**
- **Must-have:** Marketing sites, data-driven decisions
- **Nice-to-have:** Understanding user behavior
- **Skip:** Internal tools, no tracking needs

**Implementation complexity:** Low
- Integration: Inject tracking scripts (client-side)
- Server-side: Log events to analytics service
- Reporting: Dashboard or integrate with external tools

---

### 22. Content Analytics

**What:** Internal analytics about content usage
**Examples:** View counts, popular pages, search terms

**Why CMSs have it:**
- Content strategy decisions
- Editor insights
- Performance metrics

**Features:**
- Page view counts
- Popular content rankings
- Search term tracking
- User engagement metrics

**Do you need it?**
- **Nice-to-have:** Content teams, data-driven editing
- **Skip:** Small sites, external analytics sufficient

---

## Missing Packages: Localization

### 23. Multi-language/i18n

**What:** Content in multiple languages
**Examples:** WordPress WPML, Polylang, language switching

**Why CMSs have it:**
- International audiences
- Regional content
- Localized marketing

**Typical approaches:**
- **Separate content per language** (duplicate content items)
- **Translation tables** (link translations together)
- **Field-level translations** (each field has multiple language versions)

**Do you need it?**
- **Must-have:** Multi-country sites, international businesses
- **Nice-to-have:** Future international expansion
- **Skip:** Single-language sites

**Implementation complexity:** Medium to High
- Database: Add language field or separate translation tables
- API: Return content in requested language
- Fallback: Show default language if translation missing

---

## Priority Implementation Guide

### Tier 1: Essential for Production CMS (Implement First)

1. **Cache System** - Performance and scalability
2. **Search & Indexing** - User experience
3. **Content Workflow** - Draft/published states
4. **Scheduled Tasks** - Publish scheduled content
5. **Email System** - Password reset, notifications

**Why these first:**
These directly impact core CMS functionality and user experience.

---

### Tier 2: Professional Features (Implement Second)

6. **Content Versioning** - Undo mistakes, compliance
7. **SEO Management** - Visibility and marketing
8. **Queue/Background Jobs** - Performance for heavy operations
9. **Webhooks** - Integration and extensibility
10. **Redirects** - SEO and URL management

**Why these second:**
Professional sites need these, but they're not blocking basic functionality.

---

### Tier 3: Advanced Features (Implement Third)

11. **Audit Trail** - Security and compliance
12. **Granular Permissions** - Complex organizations
13. **Taxonomy (Tags)** - Flexible content organization
14. **Import/Export** - Migration and portability
15. **Comments** - Community engagement

**Why these third:**
Nice to have, but depend on specific use cases.

---

### Tier 4: Specialized Features (Implement As Needed)

16. **Video Processing** - Video-heavy sites only
17. **Form Builder** - Marketing and lead gen
18. **Analytics** - Data-driven content strategy
19. **Multi-language** - International sites
20. **Menu Builder** - Non-technical editors

**Why these last:**
Very specific needs, not all CMSs require them.

---

## What WordPress Has (That ModulaCMS Doesn't)

**WordPress packages/features:**
- ✅ Revisions (version history)
- ✅ Post status workflow (draft, pending, published)
- ✅ Comments (built-in)
- ✅ Taxonomies (categories, tags)
- ✅ Menus (navigation builder)
- ✅ Widgets (modular content blocks)
- ✅ Shortcodes (embeddable content)
- ✅ Cron (scheduled tasks)
- ✅ Transients (object cache API)
- ✅ Multisite (multiple sites, one install)
- ✅ Theme system (not needed for headless)
- ✅ Plugin ecosystem (ModulaCMS has Lua plugins)
- ✅ Media library (ModulaCMS has S3 media)
- ✅ Custom post types (ModulaCMS has datatypes)
- ✅ Custom fields (ModulaCMS has fields)
- ✅ User roles (ModulaCMS has basic roles)
- ✅ XML-RPC API (outdated, not needed)
- ✅ REST API (ModulaCMS should have this)
- ✅ Gutenberg editor (not needed for headless)
- ✅ Search (built-in, basic)

---

## What Umbraco Has (That ModulaCMS Doesn't)

**Umbraco packages/features:**
- ✅ Content versions (full history)
- ✅ Publish workflow (save, publish, unpublish)
- ✅ Language variants (multi-language content)
- ✅ Content scheduling (future publishing)
- ✅ Forms (form builder)
- ✅ Media management (similar to ModulaCMS)
- ✅ Document types (similar to ModulaCMS datatypes)
- ✅ Member groups (user management)
- ✅ Content templates (blueprints)
- ✅ Macros (reusable components)
- ✅ Examine (Lucene search integration)
- ✅ Cache management
- ✅ Relations (content relationships)
- ✅ Audit trail (change tracking)
- ✅ Health checks (system monitoring)

---

## What Laravel CMS Packages Have

**Common Laravel CMS features:**
- ✅ Filament/Nova admin panels (ModulaCMS has TUI)
- ✅ Eloquent ORM (ModulaCMS uses sqlc)
- ✅ Queue system (Redis, database)
- ✅ Cache (Redis, Memcached)
- ✅ Notifications (email, SMS, Slack)
- ✅ Broadcasting (WebSockets, Pusher)
- ✅ Task scheduling (cron)
- ✅ File storage (local, S3, similar to ModulaCMS)
- ✅ Authentication (similar to ModulaCMS)
- ✅ Authorization gates (permissions)
- ✅ API resources (JSON responses)
- ✅ Validation (request validation)
- ✅ Middleware (ModulaCMS has basic)
- ✅ Testing tools (ModulaCMS has Go tests)

---

## What ModulaCMS Does BETTER

**Advantages over traditional CMSs:**

1. **Single binary deployment** - No PHP/web server setup
2. **Multi-database support** - SQLite, MySQL, PostgreSQL
3. **Built-in HTTPS** - Let's Encrypt autocert
4. **SSH management** - Secure remote content management
5. **Tree-based structure** - O(1) tree operations with sibling pointers
6. **Headless-first** - No template cruft, pure API
7. **Go performance** - Fast, low memory usage
8. **No plugin dependencies** - Everything built-in except Lua plugins
9. **Embedded schema** - Migrations in binary
10. **S3-native** - Object storage built-in, not afterthought

---

## Recommendations for ModulaCMS

### Must Add (Critical Gaps)

1. **Cache system** - Essential for production performance
2. **Search/indexing** - Users expect search functionality
3. **Content workflow** - Draft/published states
4. **Scheduled publishing** - Basic CMS requirement

### Should Add (Professional Features)

5. **Content versions** - Undo functionality, compliance
6. **Email system** - Password reset, notifications
7. **Background jobs** - Don't block HTTP requests
8. **SEO fields** - Meta tags, Open Graph

### Nice to Have (Depends on Use Case)

9. **Webhooks** - Integration with external services
10. **Tags/taxonomy** - Flexible categorization
11. **Redirects** - URL management
12. **Audit logging** - Security and compliance

### Skip (Not Needed for Headless)

- Theme/template system (headless)
- Frontend editor (headless)
- Comment moderation UI (API only)
- Form builder UI (API only)

---

## Implementation Priority Matrix

```
Priority    | Complexity | Impact | Features
------------|-----------|--------|------------------------------------------
Critical    | Low       | High   | Content workflow, scheduled tasks
Critical    | Medium    | High   | Cache, search, email
Important   | Medium    | High   | Versions, webhooks, background jobs
Important   | Low       | Medium | SEO fields, redirects
Nice        | Medium    | Medium | Tags, audit log, import/export
Nice        | High      | Medium | Granular ACL, multi-language
Optional    | High      | Low    | Video processing, analytics, form builder
```

---

## Next Steps

1. Review this document and identify which packages match your use cases
2. Start with Tier 1 features (cache, search, workflow, scheduling, email)
3. Implement one package at a time, fully tested
4. Document each package (like existing MIDDLEWARE_PACKAGE.md docs)
5. Consider which features should be in core vs plugins

Remember: **Don't build everything at once.** Add features as you encounter real needs from users.

---

**Last Updated:** 2026-01-16
