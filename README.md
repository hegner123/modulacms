# ModulaCMS

<picture>
  <source  srcset="https://modulacms.com/examples/demo.gif">
    <img  src="https://modulacms.com/examples/demo.gif">
</picture>

## What This Is

A headless CMS that lets you build whatever admin panel your client wants without rebuilding the entire backend.

WordPress locks you into the WordPress admin. Laravel locks you into the Laravel admin. When a high-value client shows up with a dump truck of money asking for custom admin features that don't fit your CMS... you're stuck. You either hack around limitations, rebuild from scratch, or lose the contract.

ModulaCMS is a single Go binary that serves content over HTTP and gives you SSH access for backend management. The CMS doesn't dictate what the admin interface looks like. You build it however the client wants. React, Vue, vanilla JS, mobile app, whatever. The binary just serves content and lets you manage it.

Two separate HTTP endpoints from one server:

```json
{
  "client_site": "modulacms.com",
  "admin_site": "admin.modulacms.com"
}
```

Your public site hits one endpoint. Your custom admin dashboard hits another. Both backed by the same database, same binary, same SSH management interface. The client gets the admin UX they're paying for, not the one your CMS forces on you.

---

## Why Agencies Need This

**Agencies live and die by their admin panel.**

Your clients don't spend most of their time looking at the beautiful frontend you built. They spend it in the admin interface updating content, managing pages, uploading images. The admin UX is what they interact with daily.

When a client comes in with enterprise requirements... Azure Active Directory SSO, custom workflows, branded admin interface matching their internal tools, mobile content management... and you're locked into WordPress or Strapi or whatever you built on, you're in trouble.

You either tell them "sorry, can't do that" (lose the contract), hack around the CMS's limitations (fragile, breaks on updates), or rebuild the entire admin interface (expensive, kills your margin).

**ModulaCMS decouples the admin panel from the CMS.**

The binary serves content over two HTTP endpoints. You build whatever admin interface the client needs. They want drag-and-drop page builders? Build it. Admin interface matching their brand guidelines? Build it. Mobile app for content management? Build it. Integration with their internal systems? Build it.

The CMS is just a backend. Fast, flexible, enterprise-capable. The admin UX is yours to define.

---

## One Binary, Three Servers, Written In Go

ModulaCMS is a single compiled Go binary that runs:

1. **HTTP server** (fallback if HTTPS config breaks)
2. **HTTPS server** (production with Let's Encrypt built-in)
3. **SSH server** (developer/ops management TUI)

Both `client_site` and `admin_site` requests hit the same Go HTTP/HTTPS servers. The routing middleware differentiates based on hostname. One binary, one process, minimal resource footprint.

**Why Go matters:**

WordPress (PHP), Drupal (PHP), Strapi (Node.js), Django CMS (Python)... they're all interpreted languages with runtime overhead. WordPress serving 100 requests/second might need 2GB RAM and aggressive caching. ModulaCMS handling the same load uses a fraction of that and doesn't break a sweat.

Go compiles to a native binary. No interpreter. No JIT warmup. No garbage collection pauses killing your p99 latency. Concurrent by default with goroutines handling thousands of connections without thread pool tuning.

**ModulaCMS is faster than any other CMS platform on the market.** Not "fast for a CMS." Actually fast.

---

## SSH-Based Content Management For Developers

The SSH TUI isn't the client's admin panel. It's **your** power tool.

Connect to the CMS the same way you connect to your server:

```bash
ssh admin@your-cms.com
```

You get a full TUI with keyboard navigation, forms, tree visualization, and real-time updates. Built with Charmbracelet's Bubbletea framework (same tech behind lazygit). Three-column layout when editing content:

**Column 1:** Tree navigation showing content hierarchy
**Column 2:** Live preview of selected content
**Column 3:** Field editor for the current item

Database operations, content structure changes, emergency fixes, deployment tasks. Fast, scriptable, accessible from anywhere. When you need to write long-form content, pop into your actual editor (vim, emacs, whatever) instead of fighting with a textarea.

The client gets whatever beautiful web-based admin panel you build for them. You get SSH access for backend management. Different tools for different users.

---

## Enterprise OAuth Without Custom Auth Layers

Enterprise clients need SSO. Azure Active Directory, Okta, JumpCloud, Google Workspace. They're not going to let employees create "yet another password" for your CMS.

Most CMSs give you OAuth for "Sign in with GitHub" or "Sign in with Google" and call it done. Enterprise providers require custom configuration or get locked behind enterprise pricing tiers.

ModulaCMS doesn't hardcode OAuth providers. You configure the endpoints:

```json
{
  "oauth_client_id": "...",
  "oauth_client_secret": "...",
  "oauth_endpoint": {
    "oauth_auth_url": "https://login.microsoftonline.com/...",
    "oauth_token_url": "https://login.microsoftonline.com/..."
  }
}
```

Azure Active Directory? Point it at Azure's OAuth endpoints. Okta? Okta's endpoints. JumpCloud? JumpCloud's endpoints. Google Workspace with custom domain? Configure it. Custom internal OAuth server? Works.

Your enterprise client with strict SSO requirements doesn't force you to rebuild authentication. Change the config, restart, done.

---

## File Uploads Never Go To Mystery Filesystem

Here's a real problem: you're running Directus on Azure App Service. Users upload files. The files go... somewhere. Azure App Service doesn't give you SSH access to the filesystem. No FTP either. The files are in some ephemeral storage location you can't access. Good luck debugging when uploads break.

ModulaCMS makes this impossible. File uploads require S3-compatible storage. Not optional. Not defaulting to local filesystem that might work until you deploy to a managed platform. Required.

```json
{
  "bucket_url": "us-iad-10.linodeobjects.com",
  "bucket_endpoint": "modulacms.us-iad-10.linodeobjects.com",
  "bucket_media": "media-bucket",
  "bucket_backup": "backup-bucket"
}
```

You always know where uploads are going. It's right there in the config. Files go to S3-compatible storage that you can access directly, set up CDN for, configure lifecycle policies on, and migrate by changing two config values.

Separate buckets for media versus backups. Different access patterns, different lifecycle policies, different cost optimization strategies.

Client wants to switch from AWS to Linode to save costs? Change `bucket_url` and `bucket_endpoint`. Want to move media to a CDN-backed bucket? Change the bucket. The file paths are in the database, the files are in object storage, migrations are configuration changes.

---

## The Tree Structure That Actually Scales

Most CMSs store hierarchical content using simple parent_id references (adjacency list pattern). That works until you need to query deeply nested structures or maintain sibling order. Then you're writing recursive queries or loading entire trees into memory.

ModulaCMS uses sibling pointers. Each content node stores four references:

```sql
CREATE TABLE content_data (
    content_data_id INTEGER PRIMARY KEY,
    parent_id INTEGER REFERENCES content_data ON DELETE SET NULL,
    first_child_id INTEGER REFERENCES content_data ON DELETE SET NULL,
    next_sibling_id INTEGER REFERENCES content_data ON DELETE SET NULL,
    prev_sibling_id INTEGER REFERENCES content_data ON DELETE SET NULL,
    -- ... other fields
);
```

This is the same pattern filesystems use. It gives you O(1) operations when combined with a NodeIndex map:

```go
type TreeRoot struct {
    Root      *TreeNode
    NodeIndex map[int64]*TreeNode  // O(1) lookup by ID
    Orphans   map[int64]*TreeNode
}
```

Need to find a node? O(1) map lookup.
Need to insert a sibling? O(1) pointer update.
Need to delete a node? O(1) pointer rewiring.

The tree loader handles edge cases you'd normally have to debug manually. Orphaned nodes (parent loaded after child), circular references, missing parents. It loads in three phases:

```go
// Phase 1: Create all nodes and populate index
func (page *TreeRoot) createAllNodes(rows *[]db.GetContentTreeByRouteRow, stats *LoadStats) error

// Phase 2: Assign hierarchy for nodes with existing parents
func (page *TreeRoot) assignImmediateHierarchy(stats *LoadStats) error

// Phase 3: Iteratively resolve orphaned nodes
func (page *TreeRoot) resolveOrphans(stats *LoadStats) error
```

The orphan resolution is iterative with circular reference detection. If you somehow create a loop (node A's parent is node B, node B's parent is node A), the loader catches it and reports it instead of hanging.

```go
func (page *TreeRoot) hasCircularReference(node *TreeNode, visited map[int64]bool) bool {
    nodeID := node.Instance.ContentDataID

    if visited[nodeID] {
        return true // Found cycle
    }

    if !node.Instance.ParentID.Valid {
        return false // Reached root
    }

    visited[nodeID] = true
    parentID := node.Instance.ParentID.Int64
    parent := page.NodeIndex[parentID]

    if parent == nil {
        return false // Parent doesn't exist (not circular, just missing)
    }

    return page.hasCircularReference(parent, visited)
}
```

This matters when you're dealing with client-facing content trees. You can nest Pages inside Sections inside Categories to arbitrary depth without worrying about query performance degrading.

---

## Lazy Loading For Large Content Trees

Initial problem: loading a full content tree with 500+ nodes to show a root menu kills performance.

Traditional approach: `SELECT * FROM content_data WHERE route_id = ?` returns everything. You load 500 rows, build the entire tree in memory, then render 5 top-level items. Waste of bandwidth, waste of memory, slow initial load.

ModulaCMS uses lazy loading. Initially, load only root content and its immediate children (maybe 18 rows instead of 500). When the user expands a node in the TUI, load that node's children on demand.

```sql
-- Initial load: just root and first level
SELECT cd.*, dt.label as datatype_label, dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
AND (cd.parent_id IS NULL OR cd.parent_id IN (
    SELECT content_data_id FROM content_data
    WHERE parent_id IS NULL AND route_id = ?
))
ORDER BY cd.parent_id NULLS FIRST, cd.content_data_id
```

When a user expands a node, the TUI fires an `ExpandNodeMsg` that triggers loading children if they haven't been loaded yet:

```go
case ExpandNodeMsg:
    model.ContentTree.ExpandedNodes[m.NodeID] = true

    if m.LoadChildren && !model.LazyLoadStates[m.NodeID] {
        model.LazyLoadStates[m.NodeID] = true

        return model, LoadNodeChildrenCmd{
            NodeID: m.NodeID,
            Depth:  1, // load one level deeper
        }
    }
```

Performance difference for a typical page tree:

**Without lazy loading:** 500 nodes × 200 bytes = 100KB in memory
**With lazy loading:** 18 nodes × 200 bytes = 4KB initially, load more as needed

That's a 25x reduction in initial memory footprint. More importantly, it's instant feedback. The interface shows content immediately instead of waiting for a massive tree to load.

---

## Flexible Content Modeling With Datatypes

You're not locked into "pages" and "posts." You define datatypes.

```sql
CREATE TABLE datatypes (
    datatype_id INTEGER PRIMARY KEY,
    parent_id INTEGER REFERENCES datatypes ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,  -- 'ROOT', 'ELEMENT', etc.
    -- ... metadata fields
);

CREATE TABLE fields (
    field_id INTEGER PRIMARY KEY,
    parent_id INTEGER REFERENCES datatypes ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    data TEXT NOT NULL,
    type TEXT NOT NULL,  -- 'text', 'richtext', 'image', etc.
    -- ... metadata fields
);
```

A datatype is a content schema. A "Page" datatype might have Title (text), Favicon (image), and Description (text) fields. A "Product" datatype might have Name (text), Price (number), SKU (text), and Images (image) fields.

Datatypes can be nested. A "Page" datatype can have child datatypes like "Hero Section" and "Content Row." Build your content structure to match your actual site architecture instead of shoehorning everything into a blog post format.

When you create content, you pick a ROOT datatype (Page, Post, Form, or whatever you've defined). The TUI loads the fields for that datatype and presents a form. Fill it out, save, done.

```go
// When user selects "Page" datatype for new content
func SelectRootType(rootID int64) ContentCreationFlow {
    root := database.GetDatatype(rootID)
    fields := database.GetFieldsForDatatype(rootID)

    return ContentCreationFlow{
        SelectedRoot: root,
        FormFields:   fields, // Title, Favicon, Description for Page
    }
}
```

This approach gives you WordPress-like flexibility without WordPress's bloat or its opinionated post/page distinction.

---

## Draft and Published Status (Planned Feature)

Here's a pain point most CMSs get wrong: draft/published status is all-or-nothing at the page level.

You're redesigning a section of a page. The Hero is done and ready to go live. The new About section is half-finished. With traditional CMSs, you either publish the whole page (broken About section goes live) or leave it all in draft (finished Hero stays hidden).

ModulaCMS is designed to handle status at the field and section level. Publish the Hero, leave the About section in draft. Publish individual fields while keeping others as drafts.

```sql
-- Planned schema addition
ALTER TABLE content_data ADD COLUMN status ENUM('draft', 'published') DEFAULT 'published';
ALTER TABLE content_fields ADD COLUMN status ENUM('draft', 'published') DEFAULT 'published';
```

Tree example:

```
├── Hero Section (published)
│   ├── Title (published)
│   ├── Subtitle (draft) ← Being updated
│   └── Background Image (published)
├── Content Section (draft) ← Entire section being redesigned
│   ├── Paragraph 1 (draft)
│   └── Call-to-Action (draft)
└── Footer (published)
```

The public site queries `WHERE status = 'published'` and serves only ready content. The admin TUI loads all statuses and shows indicators for what's draft versus published. You can publish a single field, a section, or the whole tree with keyboard shortcuts.

This feature is specced and ready to implement. The schema migration is written, the query patterns are documented, the TUI commands are designed.

---

## Multi-Database Support

SQLite for small projects and development. MySQL or Postgres for production.

```go
type DbDriver interface {
    GetContentTree(routeID int64) (*[]ContentNode, error)
    CreateContent(data ContentData) error
    UpdateField(fieldID int64, value string) error
    // ... 150+ methods
}
```

Three implementations: `db-sqlite`, `db-mysql`, `db-psql`. Same interface, different backends. Switch by changing a config value.

The SQL is generated by `sqlc`, which takes SQL queries and outputs type-safe Go code:

```sql
-- name: GetContentTree :many
SELECT cd.*, dt.label, dt.type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?;
```

Becomes:

```go
func (q *Queries) GetContentTree(ctx context.Context, routeID int64) ([]GetContentTreeRow, error) {
    // Generated code with full type safety
}
```

No string concatenation, no SQL injection risk, no runtime query errors. If the SQL is invalid, `sqlc` catches it at compile time.

Schema migrations are embedded in the binary using `//go:embed` directives. The schema files live in `sql/schema/` organized by numbered directories (1_permissions, 2_roles, etc). On startup, ModulaCMS checks which migrations have run and applies missing ones.

This means deployment is a single binary. No separate migration runner, no schema files to sync, no version mismatches.

---

## S3-Compatible Object Storage

Images and media go to S3-compatible storage. Not just AWS S3, also Linode Object Storage, DigitalOcean Spaces, Backblaze B2, anything that speaks the S3 API.

Upload an image in the TUI, it goes to your bucket. Reference it in a field, the path is stored. The CMS tracks metadata (dimensions, file size, mime type) in a local database table for quick lookups.

```sql
CREATE TABLE media (
    media_id INTEGER PRIMARY KEY,
    bucket TEXT NOT NULL,
    key TEXT NOT NULL,
    filename TEXT NOT NULL,
    -- ... metadata
);
```

Why external storage instead of database BLOBs? Databases are bad at serving files. S3 is built for it. You get CDN integration, automatic backups, and you don't bloat your database with binary data.

---

## Backup and Restore

Backups are zip archives containing SQL dumps and media files.

```go
type Backup struct {
    Timestamp  time.Time
    Database   []byte  // SQL dump
    MediaFiles []File  // From S3
    Plugins    []File  // Lua scripts
}
```

Run a backup command in the TUI, get a timestamped zip file. Restore command extracts and replays the SQL dump, re-uploads media to your bucket, reinstalls plugins.

This makes environment syncing trivial. Backup production, restore to local dev, keep working. Backup local, restore to staging, test before deploying. The entire CMS state is in one portable file.

---

## The Elm Architecture TUI

The interface follows the Elm Architecture pattern: Model-Update-View.

**Model** is application state. The content tree, current page, active form, user session, everything.

```go
type Model struct {
    ContentTree     *TreeRoot
    CurrentPage     PageType
    ActiveForm      *Form
    UserSession     *Session
    // ... more state
}
```

**Update** is a message handler. User presses a key, that fires a message. Database query completes, that's a message. The update function receives messages and returns a new model plus optional commands.

```go
func Update(msg Msg, model Model) (Model, Cmd) {
    switch m := msg.(type) {
    case ExpandNodeMsg:
        model.ContentTree.ExpandedNodes[m.NodeID] = true
        return model, LoadNodeChildrenCmd{NodeID: m.NodeID}
    case NodeChildrenLoadedMsg:
        model.ContentTree.Nodes[m.NodeID].Children = m.Children
        return model, NoCmd()
    }
}
```

**View** renders the model to the terminal. The TUI library (Bubbletea) calls your view function every frame, you return a string (with ANSI escape codes for colors/layout), it gets drawn.

```go
func View(model Model) string {
    switch model.CurrentPage {
    case CMSPage:
        return renderCMSPage(model)  // Three-column layout
    case FormPage:
        return renderFormPage(model) // Dynamic form
    }
}
```

This architecture keeps state management sane. All state lives in Model. All state changes go through Update. View is pure (same model always renders the same output). No scattered mutations, no "just update this field over here" bugs.

It's the same pattern React uses (setState → render), just applied to a TUI instead of a web frontend.

---

## External Editor Integration

Long-form content in a terminal? Pop into your real editor.

The TUI detects when you're editing a richtext field and offers to open an external editor. Hit 'e', it spawns `$EDITOR` (vim, emacs, nano, whatever you have configured) with a temp file. You edit in your actual editor with all your plugins and keybindings. Save and quit, the content is pulled back into the CMS.

This sounds small but it's huge. Ever tried writing a blog post in a browser textarea? Ever lost work because the session timed out? Ever fought with a WYSIWYG editor that mangles your HTML?

External editor support means you write content in the tool you already know, with the workflow you already have.

---

## The Fat Config Philosophy

Every painful part of CMS deployment is in the config:

```json
{
  "environment": "production",
  "db_driver": "postgres",
  "db_url": "localhost:5432",
  "bucket_url": "us-iad-10.linodeobjects.com",
  "bucket_media": "media-bucket",
  "bucket_backup": "backup-bucket",
  "oauth_client_id": "...",
  "oauth_endpoint": {
    "oauth_auth_url": "https://login.microsoftonline.com/...",
    "oauth_token_url": "https://login.microsoftonline.com/..."
  },
  "cors_origins": ["https://admin.example.com"],
  "ssl_port": "443",
  "client_site": "example.com",
  "admin_site": "admin.example.com"
}
```

Yes, it's a fat config. Database credentials, OAuth providers, S3 buckets, CORS origins, SSL ports, site domains, all in one file.

But you configure it once per environment. The binary loads it. You never think about it again. No hunting through scattered nginx configs, no separate certbot cron jobs, no database settings hidden in web UI, no "where did I configure OAuth again?"

Everything that can break deployment is visible in one file. When something breaks, you know where to look.

---

## Let's Encrypt Built In, Localhost HTTPS Out Of The Box

No nginx reverse proxy needed. No certbot. No manual cert renewal. Let's Encrypt is built into the binary using Go's `autocert.Manager`:

```go
manager := autocert.Manager{
    Prompt: autocert.AcceptTOS,
    HostPolicy: autocert.HostWhitelist(
        configuration.Client_Site,
        configuration.Admin_Site,
    ),
}
```

The binary requests certs for both domains, handles ACME challenges, renews automatically. HTTPS just works.

**But here's the clever part:** ModulaCMS runs both HTTP and HTTPS servers from the same binary.

```go
httpServer := &http.Server{
    Addr:    configuration.Client_Site + configuration.Port,
    Handler: middlewareHandler,
}

httpsServer := &http.Server{
    Addr:      configuration.Client_Site + configuration.SSL_Port,
    TLSConfig: manager.TLSConfig(),
    Handler:   middlewareHandler,
}
```

Let's Encrypt cert expired? DNS misconfigured? TLS settings broken? **HTTP still works.** You can still SSH in and diagnose the issue instead of being completely locked out.

And localhost gets HTTPS for development. No browser security warnings, no "refused to connect" errors, no hacky workarounds that break every 6 months when browsers change their localhost cert handling.

---

## Media Optimization Like WordPress

Upload one image, get multiple sizes automatically.

```go
// Get dimension presets from database
dimensionsPTR, err := d.ListMediaDimensions()

// Crop and scale to each preset
for _, dx := range dimensions {
    cropWidth := int(dx.Width.Int64)
    cropHeight := int(dx.Height.Int64)
    // Center crop and scale
    cropRect := image.Rect(x0, y0, x0+cropWidth, y0+cropHeight)
    img := image.NewRGBA(dstRect)
    in.Scale(img, dstRect, dImg, cropRect, draw.Over, nil)

    filename := fmt.Sprintf("%s-%vx%v%s", baseName, width, height, ext)
}
```

You define dimension presets in the database (1200x800, 800x600, 400x300, whatever your design needs). Upload an image, ModulaCMS generates all sizes with center cropping and scaling using Go's `image/draw` package.

Supports PNG, JPEG, GIF, WebP. The optimized images go straight to your S3 bucket. Reference them in responsive images with proper srcset attributes.

Just like WordPress's image handling, but you control the presets and the files go to object storage instead of the server filesystem.

---

## Lua Plugins: Output Adapters, Import Adapters, Custom Logic

The plugin system uses Lua via `gopher-lua` for embedded scripting. Extend the CMS without recompiling the binary.

**Output Adapters:**

Your client has an existing Next.js site built to consume WordPress's JSON API. They want to switch to ModulaCMS but don't want to rewrite the frontend.

Write a Lua plugin that transforms ModulaCMS output into WordPress JSON format:

```lua
function transform_to_wordpress(content)
    -- Transform ModulaCMS content structure
    -- into WordPress API response format
    return {
        id = content.content_data_id,
        title = { rendered = content.fields.title },
        content = { rendered = content.fields.body },
        -- ... match WordPress schema
    }
end
```

The frontend keeps hitting the same API endpoints, gets the same JSON structure, doesn't know the backend changed.

**Import Adapters:**

Client wants to migrate from WordPress/Drupal/Contentful but has 10 years of content. Write a Lua plugin that transforms their database export into ModulaCMS format:

```lua
function import_wordpress_post(wp_post)
    -- Transform WordPress post structure
    -- into ModulaCMS datatypes and fields
    create_content({
        datatype_id = get_datatype_id("Post"),
        fields = {
            { field_id = get_field_id("Title"), value = wp_post.post_title },
            { field_id = get_field_id("Body"), value = wp_post.post_content },
            -- ... map all fields
        }
    })
end
```

Entire database migrations become plugin scripts instead of custom migration tools.

**Custom Business Logic:**

Content transformations, custom field types, workflow automation, validation rules. All without touching the core binary.

**Why Lua?**

Lua is widely adopted and has a low barrier to entry, making ModulaCMS plugin development accessible to a large pool of developers.

---

## The Tech Stack

**Go:** Single binary, fast compile times, great concurrency, minimal runtime overhead.

**Charmbracelet ecosystem:** wish (SSH server), bubbletea (TUI framework), huh (forms), lipgloss (styling). Modern terminal UI libraries with active development.

**sqlc:** Type-safe SQL code generation. Write SQL, get Go functions with full type checking.

**Lua:** Embedded scripting for plugins using gopher-lua. Extend the CMS without recompiling.

**Standard library routing:** Uses `net/http` ServeMux for HTTP routing. No framework overhead, just the Go standard library.

---

## What's Built, What's Planned

**Working now:**

Database abstraction layer with SQLite, MySQL, Postgres support. Schema migrations embedded in binary. Tree structure with sibling pointers and O(1) operations. Lazy loading for content trees. SSH server with TUI for content navigation. Form system for editing fields. External editor integration. S3-compatible storage for media. Backup/restore system. OAuth configuration (partial implementation).

**Planned (designed and specced):**

Atomic draft/published status at field and section level. Full OAuth implementation (GitHub, Google, custom providers). Plugin system completion (structure exists, needs API finalization). Admin UI for backup/restore operations. Bucket testing UI for S3 configuration.

The core architecture is solid. The tree operations are proven. The TUI framework is there. What's left is polish and completing partially-implemented features.

---

## Getting Started

Clone the repo and build:

```bash
git clone https://github.com/yourusername/modulacms.git
cd modulacms
make build
```

Run with SQLite for local development:

```bash
./modulacms --cli
```

The TUI will guide you through initial setup. Create a route (like a site), define datatypes (like Page, Post), add fields to those datatypes (Title, Body, Image), then start creating content.

For production, point it at MySQL or Postgres by editing the config. Configure S3-compatible storage for media. Set up OAuth if you want third-party authentication. Start the SSH server and connect.

Check `CLAUDE.md` for full build commands, testing, and code style guidelines.

---

## Why This Exists

**Because agencies shouldn't be held hostage by their CMS's admin panel.**

When a high-value client walks in with custom requirements... Azure AD integration, branded admin interface, custom workflows, mobile content management... you shouldn't have to choose between losing the contract and rebuilding from scratch.

ModulaCMS decouples the admin UX from the backend. Build whatever interface the client needs. The CMS just serves content.

**Because enterprise features shouldn't be locked behind enterprise pricing.**

OAuth for any provider, S3-compatible storage, multi-database support. It's all in the config. Your startup client and your enterprise client use the same binary, just configured differently.

**Because deployment shouldn't require a DevOps team.**

One binary. Let's Encrypt built in. HTTP fallback if HTTPS breaks. All config in one file. Deploy to a VPS, AWS, Linode, DigitalOcean, wherever. It just runs.

**Because you shouldn't lose files to mystery filesystem locations.**

S3-compatible storage is required, not optional. You always know where uploads go. Managed platforms without SSH access aren't a problem anymore.

**Because Go is faster than PHP/Node.js/Python.**

Compiled binary, concurrent by default, minimal memory footprint. ModulaCMS is faster than any other CMS platform on the market.

**ModulaCMS is for agencies building custom client experiences without rebuilding the entire backend every time.**




