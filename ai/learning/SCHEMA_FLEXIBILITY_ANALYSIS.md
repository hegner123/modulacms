# Schema Flexibility Analysis

**Purpose:** Identify which "missing packages" can be implemented using ModulaCMS's existing flexible schema
**Created:** 2026-01-16
**Key Insight:** ModulaCMS's dynamic datatype/field system eliminates the need for many specialized packages

---

## ModulaCMS Schema Architecture Recap

### Core Components

**Dynamic Schema System:**
```
datatypes (define structure)
    ↓
fields (define field types)
    ↓
content_data (actual content items in tree)
    ↓
content_fields (field values)
```

**Key Features:**
1. **Datatypes define structure** - Like "Post", "Page", "Product"
2. **Fields are reusable** - Title, Body, Image fields can be shared
3. **Content is tree-based** - Hierarchical with sibling pointers
4. **Multi-site support** - Routes are tree roots
5. **Field types** - Text, richtext, image, number, date, etc.

**Advantages:**
- Add new content types without code changes
- Flexible field definitions
- Reuse fields across datatypes
- Tree structure for hierarchies

---

## Features That Can Use Existing Schema

These features require **NO new packages or tables** - just clever use of existing schema.

---

### ✅ 1. SEO Fields (TRIVIAL)

**Implementation:** Add fields to existing datatypes

**How it works:**
```sql
-- Create SEO fields (one-time setup)
INSERT INTO fields (name, slug, field_type) VALUES
    ('Meta Title', 'meta_title', 'text'),
    ('Meta Description', 'meta_description', 'textarea'),
    ('OG Image', 'og_image', 'image'),
    ('Canonical URL', 'canonical_url', 'text'),
    ('Robots', 'robots', 'select');  -- index,follow / noindex,nofollow

-- Attach to any datatype (e.g., "Page" datatype)
INSERT INTO datatypes_fields (datatype_id, field_id, position)
VALUES (1, <meta_title_field_id>, 10);
```

**Usage:**
- Content editors fill in SEO fields when creating pages
- API returns SEO metadata with content
- Client renders meta tags

**Advantages:**
- No code changes needed
- Per-datatype control (some types need SEO, others don't)
- Flexible field configuration

**Example API response:**
```json
{
  "id": 123,
  "title": "Welcome to Our Site",
  "body": "...",
  "fields": {
    "meta_title": "Welcome - Best Site Ever",
    "meta_description": "Discover amazing content...",
    "og_image": "https://s3.../social.jpg"
  }
}
```

---

### ✅ 2. Menu/Navigation Builder (EASY)

**Implementation:** Use existing tree structure with a "Menu" datatype

**How it works:**
```sql
-- Create Menu datatype
INSERT INTO datatypes (name, slug) VALUES ('Menu Item', 'menu_item');

-- Create menu fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('Label', 'label', 'text'),
    ('Link Type', 'link_type', 'select'),  -- internal / external / custom
    ('Target Content', 'target_content_id', 'reference'),
    ('External URL', 'external_url', 'text'),
    ('CSS Class', 'css_class', 'text'),
    ('Icon', 'icon', 'text');

-- Create a menu route
INSERT INTO routes (name, slug, datatype_id)
VALUES ('Main Menu', 'main-menu', <menu_item_datatype_id>);
```

**Tree structure provides:**
- Hierarchical menus (dropdowns)
- Ordering via sibling pointers
- Nesting (unlimited levels)
- Fast queries (O(1) tree operations)

**Menu example:**
```
Main Menu (route)
  ├─ Home (menu_item)
  ├─ About (menu_item)
  │   ├─ Team (menu_item)
  │   └─ History (menu_item)
  ├─ Products (menu_item)
  │   ├─ Software (menu_item)
  │   └─ Hardware (menu_item)
  └─ Contact (menu_item)
```

**API returns nested structure:**
```json
{
  "menu": "main-menu",
  "items": [
    {
      "label": "Home",
      "link": "/",
      "children": []
    },
    {
      "label": "About",
      "link": "/about",
      "children": [
        {"label": "Team", "link": "/about/team"},
        {"label": "History", "link": "/about/history"}
      ]
    }
  ]
}
```

**Advantages:**
- Reuse tree navigation logic
- Multiple menus (header, footer, sidebar) via multiple routes
- Dynamic menu editing via TUI/API
- No special menu code needed

---

### ✅ 3. Comments System (EASY)

**Implementation:** Use tree structure with "Comment" datatype

**How it works:**
```sql
-- Create Comment datatype
INSERT INTO datatypes (name, slug) VALUES ('Comment', 'comment');

-- Create comment fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('Comment Body', 'body', 'richtext'),
    ('Author Name', 'author_name', 'text'),
    ('Author Email', 'author_email', 'email'),
    ('Author IP', 'author_ip', 'text'),
    ('Status', 'status', 'select'),  -- approved / pending / spam
    ('Parent Content ID', 'parent_content_id', 'number');

-- Comments can be on a dedicated route or linked to content
```

**Tree structure provides:**
- Threaded replies (comment hierarchy)
- Ordering (newest first, oldest first)
- Fast queries

**Comment thread example:**
```
Post: "Why Go is Great" (content_data)

Comments route for this post:
  ├─ Comment #1: "Great article!"
  │   ├─ Reply #1.1: "I agree!"
  │   └─ Reply #1.2: "Thanks for sharing"
  ├─ Comment #2: "Disagree, Rust is better"
  │   └─ Reply #2.1: "Both are good"
  └─ Comment #3: "What about Python?"
```

**API structure:**
```json
{
  "comments": [
    {
      "id": 1,
      "body": "Great article!",
      "author": "John Doe",
      "status": "approved",
      "replies": [
        {"id": 2, "body": "I agree!", "author": "Jane"},
        {"id": 3, "body": "Thanks for sharing", "author": "Bob"}
      ]
    }
  ]
}
```

**Moderation:**
- Add `status` field (approved/pending/spam)
- Filter comments by status in queries
- TUI interface for moderation

**Advantages:**
- Threaded replies for free (tree structure)
- No new tables needed
- Moderation via existing field system

---

### ✅ 4. Tags/Categories (MEDIUM DIFFICULTY)

**Implementation:** Create Tag/Category datatypes, link via reference fields

**Approach 1: Simple (One-to-Many)**
```sql
-- Create Tag datatype
INSERT INTO datatypes (name, slug) VALUES ('Tag', 'tag');

-- Add tag fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('Tag Name', 'name', 'text'),
    ('Tag Slug', 'slug', 'text');

-- On content datatypes, add a tag reference field
INSERT INTO fields (name, slug, field_type) VALUES
    ('Tags', 'tag_ids', 'multi_reference');  -- If multi_reference type exists
```

**Challenge:** Many-to-many relationships
- Content can have multiple tags
- Tags can be on multiple content items
- Current schema doesn't natively support many-to-many

**Approach 2: Creative (Store as JSON or Comma-Separated)**
```sql
-- Add tags field to content
INSERT INTO fields (name, slug, field_type) VALUES
    ('Tags', 'tags', 'text');  -- Store as "go,programming,cms"

-- Or use JSON field type
INSERT INTO fields (name, slug, field_type) VALUES
    ('Tags', 'tags', 'json');  -- Store as ["go", "programming", "cms"]
```

**Query by tag:**
```sql
-- SQLite JSON query
SELECT * FROM content_data cd
JOIN content_fields cf ON cd.id = cf.content_data_id
WHERE cf.field_id = <tags_field_id>
  AND json_each(cf.field_value) LIKE '%go%';

-- Or comma-separated
WHERE cf.field_value LIKE '%go%';
```

**Approach 3: Junction Table (Best, requires schema addition)**
```sql
-- Add a junction table (minor schema addition)
CREATE TABLE content_tags (
    content_data_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (content_data_id, tag_id)
);
```

**Verdict:**
- Simple tags: Use JSON or comma-separated (works today)
- Proper taxonomy: Needs junction table (minor addition)

**Categories (Hierarchical):**
- Create Category datatype
- Use tree structure for category hierarchy
- Link content via reference field

```
Categories Route:
  ├─ Technology
  │   ├─ Programming
  │   │   ├─ Go
  │   │   └─ Python
  │   └─ Hardware
  └─ Lifestyle
      ├─ Health
      └─ Travel
```

---

### ✅ 5. Redirects (EASY)

**Implementation:** Redirects as a datatype

**How it works:**
```sql
-- Create Redirect datatype
INSERT INTO datatypes (name, slug) VALUES ('Redirect', 'redirect');

-- Create redirect fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('Source URL', 'source_url', 'text'),
    ('Destination URL', 'destination_url', 'text'),
    ('Status Code', 'status_code', 'select'),  -- 301 / 302
    ('Active', 'active', 'boolean');

-- Create a route for redirects
INSERT INTO routes (name, slug, datatype_id)
VALUES ('Redirects', 'redirects', <redirect_datatype_id>);
```

**Middleware checks redirects:**
```go
func CheckRedirects(w http.ResponseWriter, r *http.Request, c *config.Config) {
    // Query redirects route for matching source_url
    redirects := db.GetContentByRoute("redirects")

    for _, redirect := range redirects {
        sourceURL := redirect.GetField("source_url")
        if r.URL.Path == sourceURL {
            destURL := redirect.GetField("destination_url")
            statusCode := redirect.GetField("status_code")

            http.Redirect(w, r, destURL, statusCode)
            return
        }
    }
}
```

**Advantages:**
- Manage redirects via TUI/API (no code deployments)
- Search/filter redirects like any content
- Audit trail (who created redirect, when)

---

### ✅ 6. Form Builder (MEDIUM)

**Implementation:** Forms as datatypes, submissions as content

**How it works:**
```sql
-- Create Form datatype
INSERT INTO datatypes (name, slug) VALUES ('Form', 'form');

-- Form fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('Form Name', 'name', 'text'),
    ('Form Fields Definition', 'fields_json', 'json'),  -- Field definitions
    ('Success Message', 'success_message', 'richtext'),
    ('Email Recipient', 'email_to', 'email');

-- Create Form Submission datatype
INSERT INTO datatypes (name, slug) VALUES ('Form Submission', 'form_submission');

-- Submission fields
INSERT INTO fields (name, slug, field_type) VALUES
    ('Form ID', 'form_id', 'reference'),
    ('Submission Data', 'data', 'json'),  -- All form field values
    ('Submitted At', 'submitted_at', 'datetime'),
    ('Submitter IP', 'ip_address', 'text');
```

**Form definition (JSON):**
```json
{
  "form_name": "Contact Form",
  "fields_json": [
    {
      "name": "full_name",
      "label": "Full Name",
      "type": "text",
      "required": true
    },
    {
      "name": "email",
      "label": "Email",
      "type": "email",
      "required": true
    },
    {
      "name": "message",
      "label": "Message",
      "type": "textarea",
      "required": true
    }
  ]
}
```

**Form submission flow:**
1. Client fetches form definition from API
2. Renders form based on JSON definition
3. Submits data to `/api/forms/{form_id}/submit`
4. Server creates form_submission content item
5. Optional: Send email notification

**Advantages:**
- No-code form creation
- Submissions stored as content (queryable, exportable)
- Form definitions versioned like any content

**Limitations:**
- Complex validation rules harder to express
- File uploads need special handling
- No visual form builder (JSON editing)

---

### ✅ 7. Content Workflow States (TRIVIAL)

**Implementation:** Add columns to content_data table

**Schema change:**
```sql
-- Add to existing content_data table
ALTER TABLE content_data ADD COLUMN status VARCHAR(20) DEFAULT 'draft';
ALTER TABLE content_data ADD COLUMN publish_date TIMESTAMP;
ALTER TABLE content_data ADD COLUMN unpublish_date TIMESTAMP;
```

**States:**
- `draft` - Work in progress, not public
- `scheduled` - Publish at future date
- `published` - Live and public
- `archived` - No longer active

**Query logic:**
```go
// Get published content only
func GetPublishedContent(routeID int) []*ContentData {
    now := time.Now()

    // Content is published if:
    // 1. status = 'published'
    // 2. publish_date is in past (or null)
    // 3. unpublish_date is in future (or null)

    return db.Query(`
        SELECT * FROM content_data
        WHERE route_id = ?
          AND status = 'published'
          AND (publish_date IS NULL OR publish_date <= ?)
          AND (unpublish_date IS NULL OR unpublish_date > ?)
    `, routeID, now, now)
}
```

**Scheduled publishing:**
- Cron job checks for content where `publish_date <= now AND status = 'scheduled'`
- Updates status to `published`

**Advantages:**
- No new tables needed
- Simple queries
- Standard CMS workflow

---

### ✅ 8. Related Content / Content References (EASY)

**Implementation:** Reference field type

**How it works:**
```sql
-- Add reference field to datatype
INSERT INTO fields (name, slug, field_type) VALUES
    ('Related Posts', 'related_posts', 'multi_reference'),
    ('Author', 'author', 'reference'),  -- Single reference
    ('Featured Image', 'featured_image_id', 'reference');  -- Reference to media table
```

**Store reference in content_fields:**
```sql
-- Store content_data ID as field value
INSERT INTO content_fields (content_data_id, field_id, field_value)
VALUES (123, <related_posts_field_id>, '456,789,101');  -- Comma-separated IDs

-- Or JSON for multi-reference
VALUES (123, <related_posts_field_id>, '[456, 789, 101]');
```

**Query with relations:**
```sql
-- Get content with related posts
SELECT
    cd.id,
    cd.title,
    GROUP_CONCAT(related.title) as related_titles
FROM content_data cd
LEFT JOIN content_fields cf ON cd.id = cf.content_data_id
LEFT JOIN content_data related ON FIND_IN_SET(related.id, cf.field_value)
WHERE cf.field_id = <related_posts_field_id>
GROUP BY cd.id;
```

**Use cases:**
- Related articles
- Author attribution (reference to User)
- Product variants
- Cross-references

---

## Features That Need Schema Extensions

These require **minor schema additions** but no new packages.

---

### 🔶 9. Tags/Categories (Proper Implementation)

**Why minor schema addition:**
- Many-to-many relationships not natively supported
- Need junction table

**Schema addition:**
```sql
CREATE TABLE content_tags (
    content_data_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,  -- References content_data where datatype = 'tag'
    PRIMARY KEY (content_data_id, tag_id),
    FOREIGN KEY (content_data_id) REFERENCES content_data(id),
    FOREIGN KEY (tag_id) REFERENCES content_data(id)
);

CREATE INDEX idx_content_tags_content ON content_tags(content_data_id);
CREATE INDEX idx_content_tags_tag ON content_tags(tag_id);
```

**Query content by tag:**
```sql
SELECT cd.* FROM content_data cd
JOIN content_tags ct ON cd.id = ct.content_data_id
WHERE ct.tag_id = ?;
```

**Verdict:**
- Minimal schema change (one junction table)
- Tags are still content items (manageable via TUI)
- Flexible and scalable

---

### 🔶 10. Search Indexing (Database Feature)

**Why database feature:**
- SQLite FTS5, MySQL FULLTEXT, PostgreSQL full-text search are built-in
- No new tables needed, just indexes

**SQLite FTS5 Implementation:**
```sql
-- Create full-text search virtual table
CREATE VIRTUAL TABLE content_fts USING fts5(
    title,
    body,
    content_data_id UNINDEXED
);

-- Populate from content_data and content_fields
INSERT INTO content_fts (title, body, content_data_id)
SELECT
    cd.title,
    GROUP_CONCAT(cf.field_value, ' ') as body,
    cd.id
FROM content_data cd
LEFT JOIN content_fields cf ON cd.id = cf.content_data_id
GROUP BY cd.id;

-- Search query
SELECT cd.* FROM content_data cd
JOIN content_fts fts ON cd.id = fts.content_data_id
WHERE content_fts MATCH 'search query'
ORDER BY rank;
```

**PostgreSQL Full-Text Search:**
```sql
-- Add tsvector column
ALTER TABLE content_data ADD COLUMN search_vector tsvector;

-- Create index
CREATE INDEX idx_search_vector ON content_data USING GIN(search_vector);

-- Update trigger to maintain search_vector
CREATE TRIGGER update_search_vector
BEFORE INSERT OR UPDATE ON content_data
FOR EACH ROW EXECUTE FUNCTION
  tsvector_update_trigger(search_vector, 'pg_catalog.english', title, body);

-- Search query
SELECT * FROM content_data
WHERE search_vector @@ to_tsquery('search & query')
ORDER BY ts_rank(search_vector, to_tsquery('search & query')) DESC;
```

**Verdict:**
- No new package needed
- Use database built-in features
- Moderate complexity (maintain indexes)

---

## Features That Need New Packages

These **cannot** be implemented with just schema flexibility.

---

### ❌ 1. Cache System

**Why needs new package:**
- Requires in-memory storage (Go maps, Redis)
- Not stored in database (defeats purpose of cache)
- Needs TTL, eviction policies

**Cannot use schema because:**
- Cache is transient (not persistent data)
- Performance-critical (must be in-memory)

---

### ❌ 2. Background Jobs / Queue

**Why needs new package:**
- Requires job queue (channel, Redis queue, database queue)
- Worker pool management
- Retry logic, failure handling

**Cannot use schema because:**
- Jobs are ephemeral (not content)
- Need concurrent execution
- Requires separate worker processes

---

### ❌ 3. Email / Notifications

**Why needs new package:**
- Requires SMTP client or API integration (SendGrid, Mailgun)
- Email templates (separate from content)
- Delivery tracking

**Cannot use schema because:**
- External service integration
- Not stored as content (transactional)

---

### ❌ 4. Webhooks

**Why needs new package:**
- Requires HTTP client to POST events
- Delivery queue and retry logic
- Signature generation (HMAC)

**Could partially use schema:**
- Webhook configurations (URLs, events) could be datatypes
- Delivery logs could be content

**But still needs:**
- Event listener/dispatcher
- HTTP client for delivery
- Background job system

---

### ❌ 5. Scheduled Tasks / Cron

**Why needs new package:**
- Requires time-based triggers (goroutine with ticker, cron library)
- Task scheduling logic
- Not event-driven (time-driven)

**Cannot use schema because:**
- Cron is infrastructure, not content
- Needs always-running process

---

### ❌ 6. Content Versioning (Automatic)

**Why needs new package:**
- Requires automatic snapshots on save (hooks/triggers)
- Diff generation between versions
- Version storage strategy

**Could use schema:**
- Versions as separate content items
- But awkward and manual

**Why doesn't fit:**
- Versioning should be transparent (automatic)
- Users shouldn't manually create versions
- Needs hooks on content save

---

### ❌ 7. Audit Trail

**Why needs new package:**
- Requires interception of all operations (middleware)
- High write volume (performance concern)
- Time-series data (different access patterns)

**Could use schema:**
- Audit logs as content items
- But would pollute content tree

**Why doesn't fit:**
- Audit data != content data
- Different retention policies
- Different query patterns (by user, by time)

---

### ❌ 8. Video Processing

**Why needs new package:**
- Requires FFmpeg or external service
- Long-running operations (transcoding)
- Multiple output formats

**Cannot use schema because:**
- External binary execution
- Resource-intensive operations
- Not just data storage

---

### ❌ 9. Advanced Image Transformations

**Why needs new package:**
- On-demand image manipulation
- URL-based parameters (`/image.jpg?w=800`)
- Real-time processing

**ModulaCMS has basic image processing:**
- Current `internal/media` does preset dimensions
- Could extend for on-demand transformations

**Verdict:**
- Extend existing `media` package
- Not a schema problem

---

## Summary Table

| Feature | Schema Alone? | Complexity | Notes |
|---------|--------------|------------|-------|
| **SEO Fields** | ✅ Yes | Trivial | Just add fields to datatypes |
| **Menus** | ✅ Yes | Easy | Use tree structure, "Menu" datatype |
| **Comments** | ✅ Yes | Easy | Use tree structure for threading |
| **Redirects** | ✅ Yes | Easy | Redirects as datatype, middleware checks |
| **Form Builder** | ✅ Yes | Medium | Forms + submissions as datatypes |
| **Workflow States** | ✅ Yes | Trivial | Add columns to content_data |
| **Related Content** | ✅ Yes | Easy | Reference field type |
| **Tags (Simple)** | ✅ Yes | Easy | JSON or comma-separated field |
| **Tags (Proper)** | 🔶 Minor addition | Medium | Need junction table |
| **Search** | 🔶 DB feature | Medium | Use SQLite FTS5 / PostgreSQL FTS |
| **Categories** | ✅ Yes | Easy | Use tree structure |
| **Cache** | ❌ No | N/A | Needs in-memory storage |
| **Background Jobs** | ❌ No | N/A | Needs queue system |
| **Email** | ❌ No | N/A | Needs SMTP/API client |
| **Webhooks** | ❌ No | N/A | Needs HTTP client, delivery system |
| **Cron/Scheduling** | ❌ No | N/A | Needs time-based triggers |
| **Versioning** | ❌ No | N/A | Needs automatic snapshots |
| **Audit Trail** | ❌ No | N/A | Different data model |
| **Video Processing** | ❌ No | N/A | Needs FFmpeg/external service |

---

## Recommended Quick Wins

### Implement These Today (Zero Code, Just Configuration)

1. **SEO Fields** - 5 minutes
   - Add meta_title, meta_description, og_image fields
   - Attach to Page/Post datatypes
   - Return in API

2. **Related Content** - 10 minutes
   - Add reference fields to datatypes
   - Store IDs in field_value
   - Query related content

3. **Categories** - 15 minutes
   - Create Category datatype
   - Create categories route
   - Add category tree
   - Link content via reference field

### Implement These This Week (Minimal Code)

4. **Workflow States** - 1 hour
   - Add status, publish_date columns to content_data
   - Update queries to filter by status
   - Add scheduled publishing cron

5. **Redirects** - 2 hours
   - Create Redirect datatype
   - Add middleware to check redirects
   - Test redirect flow

6. **Menus** - 2 hours
   - Create Menu Item datatype
   - Create menu routes
   - Add API endpoint to return nested menu structure

### Implement These This Month (Medium Effort)

7. **Search** - 1 day
   - Create FTS virtual table (SQLite) or add tsvector (PostgreSQL)
   - Maintain search index on content changes
   - Add search API endpoint

8. **Comments** - 2 days
   - Create Comment datatype
   - Build comment tree structure
   - Add moderation fields
   - Add comment API endpoints

9. **Tags (Proper)** - 2 days
   - Add content_tags junction table
   - Create Tag datatype
   - Update content API to include tags
   - Add tag filtering

10. **Form Builder** - 3 days
    - Create Form and Form Submission datatypes
    - Add form rendering API
    - Add submission endpoint
    - Add form management UI in TUI

---

## Architecture Insight

**ModulaCMS's killer feature:**
The flexible datatype/field system means many "features" are just **configuration, not code**.

Traditional CMSs need separate packages/plugins for:
- Custom post types → ModulaCMS: datatypes
- Custom fields → ModulaCMS: fields
- Taxonomies → ModulaCMS: datatypes + tree
- Menus → ModulaCMS: datatypes + tree
- SEO → ModulaCMS: fields
- Forms → ModulaCMS: datatypes

**This is powerful because:**
1. No plugin dependencies
2. No version compatibility issues
3. Consistent API (everything is content)
4. Unified management (TUI manages everything)
5. Searchable/queryable (all data in same system)

**Trade-off:**
Some features need specialized packages (cache, email, jobs), but content-related features can use schema.

---

## Next Steps

1. **Quick wins:** Implement SEO fields, related content, categories this week
2. **High impact:** Add workflow states and search this month
3. **User requests:** Build menus, comments, forms as users need them
4. **Infrastructure:** Add cache, jobs, email when performance requires it

The schema flexibility means you can **validate features with users before building packages**. Create datatypes, test with real data, then optimize with dedicated code if needed.

---

**Last Updated:** 2026-01-16
