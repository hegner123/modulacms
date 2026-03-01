# WordPress Feature Inventory (6.x Series)

Comprehensive technical feature inventory of WordPress as a CMS platform. Covers the latest stable releases in the 6.x series (through 6.9, released December 2025). Distinguishes between core (native) features and plugin-provided features.

---

## 1. Runtime Architecture

- **Language:** PHP 8.1+ required (8.2+ recommended; WordPress 7.0 expected to require 8.2+)
- **Web Server:** Runs behind Apache (mod_php or PHP-FPM) or Nginx (PHP-FPM). No built-in HTTP server; relies entirely on external web server process.
- **Process Model:** PHP shares-nothing architecture -- each HTTP request spawns (or reuses) a PHP process, loads the entire WordPress bootstrap (`wp-load.php` -> `wp-config.php` -> `wp-settings.php`), executes, and exits. No long-running daemon; no persistent in-process state between requests. Process lifecycle managed by PHP-FPM (FastCGI Process Manager) worker pools or Apache's mod_php.
- **Bootstrap Sequence:** Every request loads: `wp-config.php` (constants) -> database connection -> `wp-settings.php` (loads active plugins, active theme `functions.php`, sets up query) -> template resolution -> output.
- **WP-Cron (Pseudo-Cron):**
  - NOT a real system cron. Triggered by page visits -- on each request, WordPress checks if any scheduled events are due.
  - Scheduling API: `wp_schedule_event()` (recurring), `wp_schedule_single_event()` (one-time). Intervals: `hourly`, `twicedaily`, `daily`, or custom via `cron_schedules` filter.
  - **Limitations:** On low-traffic sites, events may never fire. On high-traffic sites, redundant checks cause overhead. Page caching can prevent `wp-cron.php` from executing.
  - **Alternative:** Disable WP-Cron (`define('DISABLE_WP_CRON', true)`) and use a real system cron (`*/5 * * * * curl -s https://example.com/wp-cron.php`) or WP-CLI (`wp cron event run --due-now`).
  - WordPress 6.9 improved WP-Cron shutdown process and spawning efficiency.
- **Object Cache:**
  - Native `WP_Object_Cache` stores data in PHP memory per-request only (non-persistent by default).
  - Persistent backends via drop-in `object-cache.php`: Redis (via Redis Object Cache plugin) or Memcached. When persistent cache is detected, the Transients API automatically routes through it.
  - Cache groups, cache invalidation, and `wp_cache_*()` functions: `wp_cache_set()`, `wp_cache_get()`, `wp_cache_delete()`, `wp_cache_flush()`.

## 2. Database Support

- **Supported Databases:** MySQL 5.7+ and MariaDB 10.4+ only. No native support for PostgreSQL, SQLite, or any other RDBMS.
  - **SQLite Exception:** WordPress Playground uses SQLite via a compatibility layer (plugin translates MySQL queries to SQLite via AST parser). The SQLite Database Integration plugin passes 99% of WordPress unit tests as of late 2025, but is not production-supported in core.
- **Database Abstraction (`wpdb` class):**
  - Global `$wpdb` object wraps `mysqli` extension. NOT an ORM -- it's a thin query helper.
  - **Read Methods:**
    - `$wpdb->get_var($sql)` -- single scalar value
    - `$wpdb->get_row($sql)` -- single row (OBJECT, ARRAY_A, or ARRAY_N)
    - `$wpdb->get_col($sql)` -- single column as array
    - `$wpdb->get_results($sql)` -- multiple rows
  - **Write Methods:**
    - `$wpdb->insert($table, $data, $format)` -- INSERT, returns rows affected
    - `$wpdb->update($table, $data, $where, $format, $where_format)` -- UPDATE
    - `$wpdb->delete($table, $where, $where_format)` -- DELETE
    - `$wpdb->replace($table, $data, $format)` -- INSERT or UPDATE on PK conflict
    - `$wpdb->query($sql)` -- arbitrary SQL (CREATE, ALTER, DROP, etc.)
  - **Security:** `$wpdb->prepare($sql, ...$args)` -- parameterized queries using `%s` (string), `%d` (integer), `%f` (float) placeholders. NOT PDO-style named params.
  - **Table Prefix:** `$wpdb->prefix` (default `wp_`), configurable per install. Must be used for all table references.
  - **Error Handling:** `$wpdb->last_error`, `$wpdb->last_query`, `$wpdb->show_errors()`, `$wpdb->suppress_errors()`.
- **Higher-Level Query API (`WP_Query`):**
  - Object-oriented query builder for posts/CPTs. Parameters: `post_type`, `post_status`, `posts_per_page`, `paged`, `orderby`, `order`, `author`, `category`, `tag`, `s` (search), etc.
  - `meta_query` -- filter by postmeta (since 3.1). Supports `key`, `value`, `compare` (`=`, `!=`, `>`, `<`, `>=`, `<=`, `LIKE`, `NOT LIKE`, `IN`, `NOT IN`, `BETWEEN`, `NOT BETWEEN`, `EXISTS`, `NOT EXISTS`), `type` (NUMERIC, CHAR, DATE, etc.), nested arrays with `relation` (AND/OR).
  - `tax_query` -- filter by taxonomy terms. Similar structure to meta_query.
  - `date_query` -- filter by date fields (post_date, post_modified). Supports `before`, `after`, `year`, `month`, `day`, `hour`, `minute`, `second`.
  - Multiple orderby since WordPress 4.0: `'orderby' => ['meta_value_num' => 'ASC', 'date' => 'DESC']`.
  - `WP_Meta_Query`, `WP_Tax_Query`, `WP_Date_Query` -- internal classes that generate SQL JOINs/WHEREs.
- **Custom Tables:** Plugins can create arbitrary tables via `$wpdb->query()` with `dbDelta()` for schema migrations (compares existing schema to desired schema and applies diffs). No migration framework -- `dbDelta()` is the only built-in tool and it's notoriously fragile with whitespace/formatting.
- **Core Table Schema (12 default tables):**
  - `wp_posts` -- posts, pages, CPTs, revisions, attachments, nav_menu_items
  - `wp_postmeta` -- key-value metadata for posts (EAV pattern)
  - `wp_terms`, `wp_term_taxonomy`, `wp_term_relationships` -- taxonomy system
  - `wp_termmeta` -- term metadata (since 4.4)
  - `wp_comments`, `wp_commentmeta` -- comments and their metadata
  - `wp_users`, `wp_usermeta` -- users and user metadata (EAV)
  - `wp_options` -- site options (key-value, autoloaded or not)
  - `wp_links` -- blogroll (legacy, rarely used)

## 3. Content Management

- **Posts:** The fundamental content unit. All content types share the `wp_posts` table with a `post_type` discriminator.
- **Pages:** Hierarchical content type (supports parent-child). Uses the same `wp_posts` table with `post_type = 'page'`.
- **Custom Post Types (CPTs):** Registered via `register_post_type()`. Any number of custom types. Arguments include `public`, `hierarchical`, `supports` (title, editor, thumbnail, excerpt, comments, revisions, custom-fields, page-attributes, etc.), `has_archive`, `rewrite`, `show_in_rest` (enables Gutenberg + REST API), `menu_icon`, `capability_type`, `taxonomies`.
- **Custom Fields (Post Meta):** Key-value pairs stored in `wp_postmeta`. Native UI is minimal (simple text key/value input). The real power comes from plugins (see Content Schema section).
- **Meta Boxes:** PHP-rendered admin UI panels attached to post edit screens via `add_meta_box()`. Custom HTML/JS. Being superseded by block editor panels but still fully supported.
- **Taxonomies:**
  - **Categories:** Hierarchical (parent-child), required default taxonomy for posts.
  - **Tags:** Flat (non-hierarchical), default taxonomy for posts.
  - **Custom Taxonomies:** Registered via `register_taxonomy()`. Can be hierarchical or flat, attached to any post type. Stored in `wp_terms` + `wp_term_taxonomy` + `wp_term_relationships`.
- **Menu System:** `wp_nav_menu()` with menu locations registered by themes. Menu items stored as `nav_menu_item` post type. Drag-and-drop admin UI. Supports posts, pages, categories, custom links, and custom post types as menu items.
- **Comments:** Built-in threaded comment system. Supports moderation (approve, spam, trash), comment metadata, avatars (Gravatar), pingbacks/trackbacks (legacy). Configurable: require login, require approval, close after N days, nesting depth.
- **Revisions:** See Section 8 (Content Versioning).
- **Post Formats:** Predefined display hints: standard, aside, gallery, link, image, quote, status, video, audio, chat. Theme must declare support. Rarely used in modern WordPress -- largely superseded by blocks.
- **Block Patterns:** Predefined block layouts that users can insert. Themes and plugins register patterns via `register_block_pattern()`. Pattern categories via `register_block_pattern_category()`. Since 6.0, patterns can be created from the editor and stored as `wp_block` posts. WordPress.org hosts a pattern directory.
- **Reusable Blocks (Synced Patterns):** Blocks saved as `wp_block` post type. Edits propagate everywhere the block is used. Since 6.3, renamed to "Synced Patterns."

## 4. Content Schema / Content Modeling

- **Custom Post Types (`register_post_type()`):**
  - Registration arguments control: labels, public visibility, REST API exposure, admin UI, menu position, capability type, rewrite rules, query var, hierarchical structure, supported features (title, editor, thumbnail, excerpt, comments, trackbacks, custom-fields, page-attributes, revisions).
  - `capability_type` maps to standard WordPress capabilities (edit, delete, publish, read) per CPT. With `map_meta_cap => true`, WordPress maps meta capabilities to primitive capabilities automatically.
  - REST API: `show_in_rest => true` exposes the CPT at `/wp/v2/{rest_base}`. Custom `rest_controller_class` for advanced usage.
- **Custom Taxonomies (`register_taxonomy()`):**
  - Can be hierarchical (like categories) or flat (like tags).
  - Supports: labels, public visibility, REST API, rewrite, `show_in_quick_edit`, `show_admin_column`, default term, sort.
  - Can be shared across multiple post types.
- **Custom Fields / Post Meta (native):**
  - Simple key-value storage in `wp_postmeta`. No schema enforcement, no field types, no validation at the DB level.
  - Native UI: plain text input for key and value on post edit screen (hidden by default in Gutenberg).
  - `register_meta()` can define type, description, default, schema, sanitize/auth callbacks, and REST API exposure.
  - Block Bindings (since 6.5): `"source": "core/post-meta"` connects block attributes directly to post meta fields without PHP template code.
- **Advanced Custom Fields (ACF) -- Plugin:**
  - 30+ field types: Text, Textarea, Number, Range, Email, URL, Password, Image, File, Gallery, oEmbed, Select, Checkbox, Radio Button, Button Group, True/False, Link, Post Object, Page Link, Relationship, Taxonomy, User, Google Map, Date Picker, Date Time Picker, Time Picker, Color Picker, Message, Accordion, Tab, Group, Repeater (Pro), Flexible Content (Pro), Clone (Pro).
  - Field groups with conditional display rules (show on post type X, page template Y, user role Z, etc.).
  - Stores data in `wp_postmeta` (same as native custom fields) -- fully compatible.
  - ACF Blocks: custom Gutenberg blocks defined via `block.json` with `acf` configuration key. Can store data in postmeta or inline in block content.
  - Now "Secure Custom Fields" after WordPress.org fork (October 2024).
- **Meta Box -- Plugin:**
  - Alternative to ACF. 40+ field types. Uses JSON Schema for field definitions.
  - Custom tables support (can store meta outside `wp_postmeta`).
  - MB Custom Post Types & Custom Taxonomies for code-free CPT/taxonomy registration.
- **Block Schema (`block.json`):**
  - Every block declares its metadata in `block.json`: name, title, category, icon, description, supports, attributes (with type/default/source), example, styles, editorScript, editorStyle, style, viewScript.
  - Attributes define the block's data model. Types: `string`, `number`, `boolean`, `object`, `array`, `null`, `integer`. Sources: `html`, `text`, `attribute`, `query`, `meta`.
  - `"usesContext"` and `"providesContext"` for block communication.

## 5. Content Delivery / API

- **REST API (built-in since 4.7):**
  - Namespace: `wp/v2`. Base URL: `/wp-json/wp/v2/`.
  - **Default Endpoints:**
    - `/posts` -- CRUD for posts
    - `/pages` -- CRUD for pages
    - `/media` -- CRUD for attachments/media
    - `/users` -- CRUD for users (authentication required for write)
    - `/comments` -- CRUD for comments
    - `/categories` -- CRUD for categories
    - `/tags` -- CRUD for tags
    - `/types` -- list registered post types
    - `/taxonomies` -- list registered taxonomies
    - `/statuses` -- list post statuses
    - `/settings` -- site settings (admin only)
    - `/search` -- unified search across content types
    - `/block-types` -- registered block types
    - `/blocks` -- reusable blocks / synced patterns
    - `/themes` -- active theme info
    - `/plugins` -- installed plugins (admin only)
    - `/menus`, `/menu-items`, `/menu-locations` -- navigation menus
    - `/templates`, `/template-parts` -- block theme templates
    - `/global-styles` -- global styles configuration
    - `/pattern-directory/patterns` -- pattern directory
  - **Filtering:** `?per_page=`, `?page=`, `?search=`, `?after=`, `?before=`, `?author=`, `?categories=`, `?tags=`, `?status=`, `?orderby=`, `?order=`, `?slug=`, `?_fields=` (field selection), `?_embed` (include linked resources).
  - **Authentication Methods:**
    - Cookie + Nonce (default for logged-in users in wp-admin; nonce via `X-WP-Nonce` header)
    - Application Passwords (since 5.6, native; Basic Auth over HTTPS)
    - OAuth 2.0 (via plugin: WP OAuth Server, or Automattic's OAuth2 for WordPress.com)
    - JWT (via plugin: JWT Authentication for WP REST API, Simple JWT Login)
  - **Custom Endpoints:** `register_rest_route($namespace, $route, $args)` in `rest_api_init` action. Supports per-endpoint permission callbacks, argument validation/sanitization, custom schema.
  - **Batch API (since 5.6):** `/wp/v2/batch` endpoint for sending multiple requests in one HTTP call.
  - **Response format:** JSON only. No XML, no content negotiation.
- **WPGraphQL -- Plugin:**
  - Adds a `/graphql` endpoint. Full schema auto-generated from registered post types, taxonomies, menus, settings.
  - Supports queries, mutations, fragments, variables, pagination (cursor-based).
  - WPGraphQL + ACF extension for exposing ACF fields in GraphQL.
  - Popular for headless WordPress with Next.js, Gatsby, Astro, etc.
- **oEmbed:**
  - WordPress is both an oEmbed **consumer** (auto-embeds URLs from whitelisted providers) and an oEmbed **provider** (other sites can embed WordPress content).
  - Whitelisted providers (40+): YouTube, Vimeo, Twitter/X, Instagram, Facebook, Spotify, SoundCloud, Flickr, TikTok, Reddit, Tumblr, Kickstarter, Imgur, Issuu, SlideShare, DailyMotion, TED, WordPress.tv, Amazon Kindle, Wolfram Cloud, and more.
  - Custom providers via `wp_oembed_add_provider()`.
  - Discovery: WordPress checks for oEmbed discovery links (`<link>` tags) on pasted URLs.

## 6. Block Editor (Gutenberg)

- **Block Count:** 113 core blocks as of WordPress 6.8/6.9.
- **Block Categories:**
  - **Text:** Paragraph, Heading (H1-H6), List, Quote, Classic (TinyMCE fallback), Code, Preformatted, Pullquote, Table, Verse, Footnotes, Details.
  - **Media:** Image, Gallery, Audio, Cover, File, Media & Text, Video.
  - **Design:** Buttons, Columns, Group, Row, Stack, More, Page Break, Separator, Spacer.
  - **Widgets:** Archives, Calendar, Categories List, Custom HTML, Latest Comments, Latest Posts, Page List, RSS, Search, Social Icons, Tag Cloud, Shortcode.
  - **Theme:** Navigation, Site Logo, Site Title, Site Tagline, Query Loop, Post Title, Post Excerpt, Post Featured Image, Post Content, Post Author, Post Date, Post Terms, Post Comments, Login/Out, Template Part, Header, Footer, Post Navigation Link, Comments Pagination, etc.
  - **Embeds:** YouTube, Twitter, Vimeo, WordPress, SoundCloud, Spotify, Flickr, Dailymotion, Reddit, TikTok, Pinterest, Amazon Kindle, etc. (30+ embed blocks wrapping oEmbed providers).
- **Custom Blocks:** Built with `@wordpress/create-block` scaffold. Defined via `block.json` + React JSX for `edit` (editor) and `save` (frontend) functions. Server-side rendered blocks via `render_callback` in PHP.
- **Block Patterns:** Pre-built compositions of multiple blocks. Registered via `register_block_pattern()` or by placing pattern files in `patterns/` directory. WordPress.org pattern directory for community patterns.
- **Full-Site Editing (FSE):**
  - Block themes use HTML template files (not PHP) in `templates/` and `parts/` directories.
  - Site Editor (`/wp-admin/site-editor.php`): visual editing of all templates (index, single, page, archive, 404, search, etc.) and template parts (header, footer, sidebar).
  - Template hierarchy follows the same precedence as classic themes (e.g., `single-{post_type}.html`, `single.html`, `singular.html`, `index.html`).
  - Style Variations: alternative `theme.json` files in `styles/` directory.
- **theme.json (version 3, since 6.6):**
  - Central configuration for block themes AND classic themes.
  - **Settings:** color palettes, gradients, duotone, font sizes, font families, spacing sizes, custom properties, appearance tools, layout (contentSize, wideSize), typography (fluid, line-height, text-decoration, letter-spacing), border, shadow, dimensions.
  - **Styles:** global and per-block styles for color, typography, spacing, border, shadow, outline, dimensions, filter, background.
  - **templateParts:** declares template parts with area assignments (header, footer, general).
  - **customTemplates:** registers custom page/post templates.
  - **patterns:** references pattern slugs from the pattern directory.
- **Block Themes vs Classic Themes:**
  - **Block themes:** HTML templates, `theme.json`, Site Editor, no PHP template files, no Customizer, no widget areas, no `functions.php` for display logic.
  - **Classic themes:** PHP templates, Customizer, widget areas, `functions.php`, optional `theme.json` for editor settings.
  - **Hybrid:** Classic themes can add `theme.json` for partial block editor support. Since WordPress 6.9, classic themes load block CSS on demand by default.

## 7. Content Publishing

- **Core Post Statuses:**
  - `draft` -- saved but not published, only visible to author and editors
  - `auto-draft` -- automatically created when starting a new post, before first save
  - `pending` -- submitted for review, awaiting editor/admin approval
  - `publish` -- live and visible to all visitors
  - `future` -- scheduled for future publication (automatically transitions to `publish` at scheduled time via WP-Cron)
  - `private` -- published but visible only to users with `read_private_posts` capability (admins, editors)
  - `trash` -- soft-deleted, recoverable for 30 days (configurable via `EMPTY_TRASH_DAYS` constant)
  - `inherit` -- used by revisions and attachments, inherits parent status
- **Custom Post Statuses:** `register_post_status()` allows defining additional statuses. However, Gutenberg support for custom statuses is incomplete -- they work in classic editor but have limited UI in the block editor. Plugins like PublishPress Statuses fill this gap.
- **Scheduled Publishing:** Set a future date/time on any post. WordPress uses WP-Cron to transition `future` to `publish` at the scheduled time. If WP-Cron fails (low traffic, caching), scheduled posts can miss their publication time ("missed schedule" issue).
- **Bulk Actions:** Posts list table supports bulk edit (status, author, categories, tags, sticky, comments, pings) and bulk trash/delete/restore.
- **Quick Edit:** Inline editing of post title, slug, date, status, categories, tags, and other fields directly from the posts list table.
- **Editorial Workflow (native):** Limited to draft -> pending -> publish flow. No native approval chains, editorial comments, or multi-step workflows.
- **Editorial Workflow (plugins):**
  - **PublishPress / PublishPress Statuses:** Custom statuses (Pitch, Assigned, In Progress, Approved), editorial comments, content calendar, notifications, editorial metadata.
  - **Edit Flow:** Custom statuses, editorial comments, editorial metadata, content calendar, user groups, notifications. (Less actively maintained.)
- **Sticky Posts:** Pin posts to the top of the blog page. Simple boolean flag per post.
- **Post Locking:** When two users edit the same post, WordPress shows a lock dialog. One user gets the lock; others see a "take over" option.

## 8. Content Versioning

- **Revisions System (native since 2.6):**
  - Every save of a published or draft post creates a revision stored as a child post with `post_type = 'revision'` and `post_status = 'inherit'`.
  - Revisions store: post_title, post_content, post_excerpt, post_author (who made the revision).
  - **Autosave:** WordPress autosaves every 60 seconds (configurable via `AUTOSAVE_INTERVAL` constant). Maximum ONE autosave per user per post. Autosaves are stored as revisions with `post_name = {parent_id}-autosave-v1`.
  - Regular revisions: `post_name = {parent_id}-revision-v1`, incrementing.
- **Revision Comparison (Diff):**
  - Visual diff screen shows additions (green), deletions (red), and unchanged content.
  - Slider UI to browse through revision history.
  - "Compare any two revisions" mode with two handles on the slider.
  - Compares title, content, and excerpt.
- **Revision Limits:**
  - Default: unlimited revisions (every save creates one).
  - `WP_POST_REVISIONS` constant in `wp-config.php`: `true` (unlimited), `false` or `0` (disable), integer N (keep N revisions per post). Old revisions auto-deleted when limit reached.
  - `wp_revisions_to_keep` filter for per-post-type or per-post control.
- **Revision Restore:** One-click "Restore This Revision" button replaces current content with selected revision. Creates a new revision recording the restore action.
- **Limitations:**
  - Revisions only track post_title, post_content, post_excerpt. NOT post meta, taxonomy terms, or featured images.
  - No branching, no merge, no named versions/tags.
  - No diff for custom fields or block-level changes.
  - Revision storage can bloat the database significantly (hundreds of revisions per post over time).
  - **Plugin:** WP Revisions Control, Revision Strike (for cleanup).

## 9. Internationalization (i18n)

- **Interface Translation (native):**
  - Full gettext-based i18n system: `__()`, `_e()`, `_x()`, `_n()`, `_nx()`, `esc_html__()`, `esc_attr__()`, etc.
  - JavaScript: `wp.i18n.__()`, `wp.i18n._x()`, `wp.i18n._n()` via `@wordpress/i18n` package.
  - Translation files: `.po`/`.mo` (classic) and `.json` (JavaScript, since 5.0).
  - Community translations via translate.wordpress.org for core, themes, and plugins.
  - Admin language can be set per-user (since 4.7). Site language set in Settings > General.
- **Content Translation (NOT native):**
  - WordPress has NO built-in multi-language content management. The core only handles one language per site (interface translation only).
  - **Approaches (all plugin-based):**
    - **Separate posts per language** (WPML, Polylang): creates duplicate posts linked via metadata. Each translation is a full post with its own URL.
    - **Post meta storage** (TranslatePress): translations stored in a custom table, original content in `wp_posts`. Frontend translation via string replacement.
    - **Multisite per language:** each language as a separate site in a multisite network (MultilingualPress plugin).
  - **WPML (plugin, commercial):**
    - Most popular multi-language plugin. Paid license required ($39-$159/year).
    - Creates translation entries in custom `icl_translations` table linking posts across languages.
    - Translation management dashboard, professional translation service integrations.
    - URL modes: subdirectory (`/en/`, `/fr/`), subdomain, domain per language, URL parameter.
    - String translation for theme/plugin strings.
    - Adds 15-30% database overhead.
  - **Polylang (plugin, free + pro):**
    - Uses WordPress taxonomies for language assignment (lighter approach).
    - Free version: manual content translation, language switcher, URL management.
    - Pro version: frontend translation, string translation, Lingotek integration.
    - Lighter database impact (~5% overhead).
  - **TranslatePress (plugin, free + pro):**
    - Visual frontend translation editor (translate directly on the page).
    - Stores translations in a dedicated `trp_translations` table.
    - Automatic translation via Google Translate or DeepL API.

## 10. Media Management

- **Media Library:**
  - Grid and list views. Filtering by type (image, audio, video, document), date, and upload source.
  - Drag-and-drop upload. Multiple file upload.
  - Attachments stored as `post_type = 'attachment'` in `wp_posts`. File metadata in `wp_postmeta`.
  - Media details: title, caption, alt text, description. Editable per-attachment.
- **Image Sizes:**
  - Default sizes: Thumbnail (150x150, hard crop), Medium (300x300 max), Medium Large (768px width), Large (1024x1024 max), Full (original).
  - Since 5.3: additional `1536x1536` and `2048x2048` sizes auto-generated.
  - Custom sizes via `add_image_size($name, $width, $height, $crop)`. Hard crop (exact dimensions) or soft crop (proportional).
  - `srcset` and `sizes` attributes automatically generated for responsive images (since 4.4).
  - Since 5.3: automatic big image threshold -- images larger than 2560px are scaled down; original kept as `-scaled` variant.
- **Image Editing (native):**
  - Built-in editor: crop (with aspect ratio options), rotate (90-degree increments), flip (horizontal/vertical), scale.
  - Can apply edits to specific sizes (thumbnail only, all except thumbnail, all sizes).
  - Uses GD library or ImageMagick (if available, preferred).
- **PDF Preview:** Generates thumbnail previews of PDF first page (requires Ghostscript or ImageMagick on server).
- **Audio/Video:** Native HTML5 `<audio>` and `<video>` players. Playlist support. Metadata extraction (ID3 tags for audio, basic video metadata).
- **EXIF Handling:** Reads EXIF data from JPEG images (camera, aperture, shutter speed, ISO, date, GPS). Stored in attachment metadata. Optional auto-rotation based on EXIF orientation (since 5.3, enabled by default).
- **Image Optimization (native):** Limited -- WordPress compresses JPEGs at quality 82 (filterable via `wp_editor_set_quality`). No WebP/AVIF conversion native (WebP support added in 5.8 for uploads; AVIF in 6.5). No lazy loading markup natively until 5.5 (`loading="lazy"`).
- **Image Optimization (plugins):** ShortPixel, Imagify, Smush, EWWW Image Optimizer -- provide lossy/lossless compression, WebP/AVIF conversion, CDN serving, bulk optimization.
- **External Media:** No native support for external media URLs or DAM integration. Plugins available (External Media without Import, WP Offload Media for S3/GCS/Azure).

## 11. Authentication

- **Cookie-Based (native, default):**
  - `wp_signon()` sets authentication cookies (`wordpress_logged_in_{hash}`, `wordpress_{hash}`).
  - Cookies are HttpOnly, Secure (when HTTPS), scoped to the site path.
  - Session token system: multiple concurrent sessions per user, viewable/revokable in user profile.
  - Password hashing: PHPass (portable bcrypt-based hashing).
- **Application Passwords (native, since 5.6):**
  - Per-user generated passwords for API access. Used with HTTP Basic Auth over HTTPS.
  - Scoped per-application (name label). Revocable individually.
  - Intended for REST API and XML-RPC authentication.
  - Stored hashed in `wp_usermeta`.
- **OAuth 2.0 (plugin):**
  - NOT native. Requires plugin: WP OAuth Server, OAuth2 Complete for WordPress.
  - WordPress.com uses OAuth2 natively for its hosted API.
  - Grants: authorization code, client credentials, refresh tokens.
- **JWT (plugin):**
  - NOT native. Plugins: JWT Authentication for WP-REST-API, Simple JWT Login.
  - Typically issues JWT on `/wp-json/jwt-auth/v1/token` endpoint.
  - Stateless authentication for SPAs and mobile apps.
- **Two-Factor Authentication (plugin):**
  - NOT native. Plugins: Two Factor (official feature plugin), WP 2FA, Wordfence Login Security.
  - Methods: TOTP (authenticator apps), email codes, backup codes, FIDO U2F/WebAuthn (hardware keys).
- **LDAP/Active Directory (plugin):**
  - NOT native. Plugins: Simple LDAP Login, AuthLDAP, miniOrange LDAP.
  - Maps LDAP groups to WordPress roles.
- **SAML/SSO (plugin):**
  - NOT native. Plugins: miniOrange SAML SSO, OneLogin SAML SSO.

## 12. Authorization / Roles & Capabilities

- **Default Roles (6 roles):**
  - **Super Admin** -- multisite only. Access to network admin and all sites. Can install/remove plugins/themes network-wide.
  - **Administrator** -- full single-site access. All capabilities within one site. Can install plugins, edit themes, manage users, change settings.
  - **Editor** -- publish and manage ALL posts (own and others). Manage categories, tags, links. Moderate comments. Cannot install plugins or change settings.
  - **Author** -- publish and manage own posts only. Can upload files.
  - **Contributor** -- write and manage own posts but CANNOT publish. Submitted posts go to "pending review." Cannot upload files.
  - **Subscriber** -- read-only. Can manage own profile. Default role for new registrations.
- **Capabilities System:**
  - ~70 primitive capabilities in core (e.g., `edit_posts`, `publish_posts`, `delete_others_posts`, `manage_options`, `upload_files`, `edit_theme_options`, `moderate_comments`, `manage_categories`, `edit_users`, `install_plugins`, `activate_plugins`, `edit_files`, `unfiltered_html`, `export`, `import`).
  - Meta capabilities: abstract capabilities (`edit_post`, `delete_post`, `read_post`) mapped to primitive capabilities via `map_meta_cap()` based on context (e.g., "can this user edit THIS specific post?" checks ownership, post status, post type).
  - Custom capabilities via `add_cap()` on role objects.
  - Custom roles via `add_role($role, $display_name, $capabilities)`, `remove_role()`.
  - Roles stored in `wp_options` table (serialized array). Capabilities stored per-role.
  - Per-user capability overrides via `WP_User::add_cap()` (stored in `wp_usermeta`).
- **Per-Post-Type Capabilities:**
  - `register_post_type()` with `capability_type` generates type-specific capabilities (e.g., for CPT "book": `edit_books`, `publish_books`, `delete_books`, `edit_others_books`, `read_private_books`).
  - `map_meta_cap => true` required for proper capability mapping.
- **Capability Check Functions:** `current_user_can($capability)`, `user_can($user, $capability)`, `author_can($post, $capability)`.
- **Plugins:** User Role Editor, Members (by developer Justin Tadlock) -- GUI for editing roles/capabilities.

## 13. Admin Panel (wp-admin)

- **Dashboard (`/wp-admin/`):** Widget-based dashboard with drag-and-drop layout. Default widgets: At a Glance (post/page/comment counts), Activity (recent posts, comments), Quick Draft, WordPress Events and News. Custom dashboard widgets via `wp_add_dashboard_widget()`.
- **Admin Bar (Toolbar):** Persistent top bar on frontend and backend for logged-in users. Contains: site name link, New (post/page/media/user), Edit (current page), Comments, Updates, user menu. Extensible via `admin_bar_menu` hook and `WP_Admin_Bar::add_node()`. Togglable per-user.
- **Screen Options:** Per-screen panel (top-right) controlling: visible columns, items per page, view mode (compact/extended), visible meta boxes. State saved per-user in `wp_usermeta`.
- **Custom Admin Pages:** `add_menu_page()`, `add_submenu_page()`, `add_options_page()`, `add_management_page()`. Full control over slug, capability requirement, render callback, icon, position.
- **Admin Notices:** `admin_notices` hook. Standard classes: `notice-success`, `notice-error`, `notice-warning`, `notice-info`. Dismissible notices via `is-dismissible` class.
- **Help Tabs:** Contextual help via `WP_Screen::add_help_tab()`. Per-screen documentation panels.
- **List Tables (`WP_List_Table`):** Standardized data tables for posts, pages, users, comments, plugins, etc. Features: sortable columns, bulk actions, search, pagination, row actions, inline edit. Custom columns via `manage_{post_type}_posts_columns` and `manage_{post_type}_posts_custom_column` filters.
- **Settings API:** `register_setting()`, `add_settings_section()`, `add_settings_field()`. Handles validation, sanitization, nonce verification, and rendering of settings pages with standard WordPress UI.

## 14. Plugin System

- **Hook System (Actions + Filters):**
  - **Actions:** `do_action($tag, ...$args)` / `add_action($tag, $callback, $priority, $accepted_args)`. Execute side effects at specific points. ~2,500+ action hooks in core.
  - **Filters:** `apply_filters($tag, $value, ...$args)` / `add_filter($tag, $callback, $priority, $accepted_args)`. Transform values through a pipeline. ~2,200+ filter hooks in core.
  - Priority system (default 10, lower = earlier). `remove_action()` / `remove_filter()` to unhook.
  - All hooks fire synchronously within the PHP request lifecycle. No async/event-queue mechanism.
- **Plugin API:**
  - Plugins are PHP files in `wp-content/plugins/`. Single-file or directory-based.
  - Plugin header: `Plugin Name`, `Version`, `Author`, `Description`, `Requires PHP`, `Requires at least` (WP version), `Requires Plugins` (since 6.5).
  - Activation/deactivation hooks: `register_activation_hook()`, `register_deactivation_hook()`.
  - Uninstall: `register_uninstall_hook()` or `uninstall.php` file.
- **Plugin Directory:** 59,000+ free plugins on wordpress.org/plugins. Hosted SVN repository with readme.txt standard, version tagging, and automated security scanning.
- **Must-Use Plugins (mu-plugins):**
  - Located in `wp-content/mu-plugins/`. Always active, cannot be deactivated via admin UI.
  - Load before regular plugins (alphabetically). No activation/deactivation hooks.
  - No subdirectory autoloading (only top-level PHP files load automatically).
  - Use case: site-critical functionality, security rules, custom object cache setup.
- **Drop-in Plugins:**
  - Special filenames in `wp-content/` that override core behavior:
    - `advanced-cache.php` -- page caching (requires `WP_CACHE` constant)
    - `object-cache.php` -- persistent object cache backend (replaces `WP_Object_Cache`)
    - `db.php` -- custom database class (wraps/replaces `$wpdb`)
    - `db-error.php` -- custom database error page
    - `maintenance.php` -- custom maintenance mode page
    - `php-error.php` -- custom PHP error page
    - `fatal-error-handler.php` -- custom fatal error handler (since 5.2)
    - `install.php` -- custom install script
    - `sunrise.php` -- multisite early loading (requires `SUNRISE` constant)
- **Plugin Dependencies (since 6.5):**
  - `Requires Plugins` header declares plugin dependencies by slug.
  - WordPress blocks activation if dependencies are not installed/active.
  - Does NOT support mu-plugin dependencies.
- **Auto-Updates:**
  - Plugin and theme auto-updates since 5.5 (opt-in per plugin/theme in admin UI).
  - Minor core auto-updates since 3.7 (enabled by default).
  - Major core auto-updates since 5.6 (opt-in; default for new installs).
  - Configurable via `auto_update_plugin`, `auto_update_theme` filters and `WP_AUTO_UPDATE_CORE` constant.

## 15. Theme System

- **Classic Themes:**
  - PHP template files (`index.php`, `single.php`, `page.php`, `archive.php`, `header.php`, `footer.php`, `sidebar.php`, etc.).
  - **Template Hierarchy:** WordPress selects the most specific template available. For a single post: `single-{post_type}-{slug}.php` > `single-{post_type}.php` > `single.php` > `singular.php` > `index.php`.
  - `functions.php` -- theme setup, hooks, feature declarations.
  - Customizer for theme options (colors, header image, menus, widgets).
  - Widget areas via `register_sidebar()` and `dynamic_sidebar()`.
  - Template tags: `the_title()`, `the_content()`, `the_excerpt()`, `the_post_thumbnail()`, `get_template_part()`, etc.
- **Block Themes:**
  - HTML template files in `templates/` and `parts/` directories.
  - `theme.json` as the central configuration (replaces Customizer, `add_theme_support()` calls, and much of `functions.php`).
  - Site Editor for full visual editing of all templates.
  - No `functions.php` required (though still supported for hooks/enqueuing).
  - No widget areas, no Customizer. Everything is blocks.
  - Default theme since WordPress 5.9: Twenty Twenty-Two (first default block theme). Twenty Twenty-Three, Twenty Twenty-Four, Twenty Twenty-Five followed.
- **Child Themes:**
  - Inherit all templates and functionality from parent theme.
  - Override specific templates by placing same-named files in child theme.
  - `style.css` with `Template:` header pointing to parent.
  - `functions.php` loads BEFORE parent's `functions.php`.
  - Block themes: child themes can override `theme.json`, templates, and parts.
- **Starter Content:** `add_theme_support('starter-content', $content)` -- provides default pages, posts, widgets, menus, and theme mods for new sites. Only activates on fresh installs.
- **Theme Marketplace:** wordpress.org/themes (11,000+ free themes). Commercial theme marketplaces: ThemeForest, Elegant Themes, StudioPress, Flavor themes, etc.

## 16. Webhooks

- **Native Webhooks:** WordPress has NO built-in webhook system (no outgoing HTTP notifications on content events).
- **Action Hooks as Alternative:** Developers can attach custom code to action hooks (`save_post`, `wp_insert_comment`, `user_register`, etc.) to make outgoing HTTP requests using `wp_remote_post()`. This is DIY, not a configurable webhook system.
- **WooCommerce Webhooks:** WooCommerce (plugin) has a built-in webhook system with admin UI for configuring delivery URLs, topics (order.created, product.updated, etc.), and secrets. This is WooCommerce-specific, not core WordPress.
- **WP Webhooks (plugin):**
  - Bi-directional webhook system. Triggers (outgoing) and actions (incoming).
  - Triggers: user login/registration/update/delete, post create/update/delete, comment events, email events, custom action hooks.
  - Actions: create/update/delete users, posts, comments; fire custom hooks; execute PHP.
  - Authentication: API Key, Bearer Token, Basic Auth per webhook.
  - JSON payloads. Retry logic. Logging.
- **Other Solutions:** Zapier/Make integrations via their respective WordPress plugins; Uncanny Automator; AutomatorWP.

## 17. WP-CLI

- **Overview:** Official command-line interface for WordPress. Installable as a standalone PHAR (`wp-cli.phar`) or via Composer/Homebrew.
- **Core Commands:**
  - `wp core` -- download, install, update, verify-checksums, version, is-installed, multisite-convert, multisite-install
  - `wp db` -- create, drop, reset, optimize, repair, cli (opens MySQL shell), query, export, import, search, size, tables, prefix
  - `wp plugin` -- install, activate, deactivate, delete, update, list, search, status, get, toggle, verify-checksums, auto-updates
  - `wp theme` -- install, activate, delete, update, list, search, status, get, mod, enable/disable (multisite)
  - `wp post` -- create, update, delete, get, list, generate, meta (get/set/delete/list)
  - `wp user` -- create, update, delete, get, list, generate, import-csv, meta, add-role, remove-role, set-role, add-cap, remove-cap, list-caps, session (destroy)
  - `wp comment` -- create, update, delete, get, list, approve, unapprove, spam, unspam, trash, untrash, generate, meta, recount
  - `wp option` -- get, set, delete, list, update, add, pluck, patch
  - `wp cache` -- add, delete, flush, get, set, type (shows active cache backend)
  - `wp cron` -- event list, event run, event schedule, event unschedule, event delete, schedule list, test (verifies cron spawning)
  - `wp search-replace` -- find/replace strings in database with serialization-safe handling. Supports `--dry-run`, `--precise`, `--regex`, `--all-tables`.
  - `wp scaffold` -- plugin, theme, child-theme, post-type, taxonomy, block, _s (starter theme). Generates boilerplate files.
  - `wp export` -- export content to WXR (WordPress eXtended RSS) XML format
  - `wp import` -- import WXR files
  - `wp eval` -- execute arbitrary PHP code
  - `wp eval-file` -- execute a PHP file
  - `wp rewrite` -- flush, list, structure
  - `wp transient` -- delete, get, set, type, list
  - `wp media` -- regenerate (regenerate image sizes), import, fix-orientation
  - `wp menu` -- create, delete, list, item add/update/delete/list, location assign/list/remove
  - `wp widget` -- add, update, delete, deactivate, list, move, reset
  - `wp config` -- create, get, set, delete, list, has, path, shuffle-salts, edit
  - `wp language` -- core install/list/activate/update, plugin install/list/update, theme install/list/update
  - `wp maintenance-mode` -- activate, deactivate, status, is-active
  - `wp site` -- create, delete, list, archive/unarchive, activate/deactivate, spam/not-spam, empty, switch-language (multisite)
  - `wp super-admin` -- add, remove, list (multisite)
- **Extensibility:** Custom commands via `WP_CLI::add_command()`. Package system for third-party command bundles.
- **Scripting:** `--format=json` output for machine parsing. Exit codes for automation.

## 18. Deployment

- **Traditional Hosting:**
  - Upload via FTP/SFTP. Database import via phpMyAdmin. Manual `wp-config.php` configuration.
  - One-click installers (Softaculous, Fantastico) on shared hosting.
  - No built-in deployment pipeline, version control integration, or staging.
- **Managed WordPress Hosting:**
  - **WP Engine:** Git push deploy, staging environments, EverCache (page + object cache), automated backups, Global Edge Security (Cloudflare). From $20/month.
  - **Kinsta:** Google Cloud Platform, LXD containers, staging with push/pull, SSH/WP-CLI, Redis, Git. From $35/month.
  - **Flywheel:** Staging, SFTP, SSL, CDN, nightly backups. From $15/month.
  - **Pantheon:** Git-based workflow, dev/test/live environments, Terminus CLI, Redis, Fastly CDN.
  - **WordPress.com (Automattic):** Hosted WordPress with various plan tiers. Limited plugin/theme control on lower plans; Business ($33/month) and Commerce ($70/month) plans allow full plugin/theme installation.
- **Bedrock (Roots):**
  - Modern WordPress boilerplate. Composer-managed dependencies (WordPress core, plugins, themes as Composer packages).
  - 12-factor app principles: `.env` file for environment-specific config, `config/` directory instead of `wp-config.php`, separate `web/` directory as document root.
  - `wp-content/` renamed to `app/` with `mu-plugins/`, `plugins/`, `themes/`, `uploads/` subdirectories.
  - Version control friendly: only committed code in repo, dependencies via `composer.lock`.
- **Docker:**
  - Official `wordpress` Docker image (Apache + PHP + WordPress). Typically paired with `mysql`/`mariadb` container.
  - Docker Compose stacks for local development.
- **wp-env:**
  - Official local development environment tool from WordPress. Requires Docker and Node.js.
  - `@wordpress/env` package. `.wp-env.json` configuration for plugins, themes, WordPress version, PHP version.
  - `wp-env start`, `wp-env stop`, `wp-env clean`, `wp-env run <command>`.
  - Creates isolated WordPress instances for plugin/theme development and testing.
- **Other Local Tools:** Local by Flywheel (GUI), DevKinsta, MAMP, XAMPP, Lando, Valet (macOS).

## 19. Email

- **`wp_mail()` (native):**
  - WordPress wraps PHP's `mail()` function via PHPMailer library (bundled in core).
  - Supports HTML emails, attachments, CC/BCC, custom headers.
  - Default: sends from `wordpress@yourdomain.com` using PHP `mail()` (relies on server's MTA -- sendmail/postfix).
  - Filters: `wp_mail_from`, `wp_mail_from_name`, `wp_mail_content_type`, `phpmailer_init` (for SMTP configuration).
- **SMTP (plugin-based):**
  - `wp_mail()` uses PHP `mail()` by default, which often results in emails going to spam or not being delivered.
  - **WP Mail SMTP** (4M+ active installs): reconfigures PHPMailer to use SMTP. Supports: Gmail/Google Workspace, Outlook 365, SendGrid, Mailgun, Amazon SES, Brevo (Sendinblue), SMTP.com, Postmark, SparkPost, Zoho Mail, generic SMTP.
  - **FluentSMTP:** Free alternative. Multiple connection profiles, email logging, auto-retry.
  - Features via plugins: email logging, delivery tracking, rate limiting, backup mailer (failover), email queue.
- **Transactional Email Providers:** SendLayer, Brevo (300 free/day), Mailgun (5,000 free/month for 3 months), SendGrid, Amazon SES, Postmark. All integrated via SMTP plugins.
- **Core Email Events:** New user registration, password reset, comment notification, update notifications, post author notifications, admin email change confirmation.

## 20. Backup & Restore

- **Native Backup:** WordPress has NO built-in backup system.
- **Native Export/Import (WXR):**
  - Tools > Export: exports content to WXR (WordPress eXtended RSS) XML format. Exports: posts, pages, CPTs, comments, categories, tags, custom taxonomies, users (as authors), media references (URLs, not files).
  - Tools > Import: imports WXR files. Can remap authors. Downloads and imports media files from URLs.
  - Limitations: WXR does NOT include theme settings, plugin settings, widget configurations, Customizer settings, or the database options table. Not a complete backup.
- **Backup Plugins:**
  - **UpdraftPlus** (3M+ active installs): scheduled backups (files + database), remote storage (S3, Google Drive, Dropbox, Azure, FTP, email), one-click restore, incremental backups (premium), migration/clone, backup encryption (premium).
  - **Duplicator** (1.5M+ installs): packages entire site into archive + installer.php. Used primarily for migration. Scheduled backups in Pro version.
  - **All-in-One WP Migration** (5M+ installs): exports entire site as `.wpress` file. Simple drag-and-drop restore. Free version limited to 512MB import size; Unlimited extension is paid.
  - **BackWPup** (800K+ installs): database and file backups, scheduled via WP-Cron or system cron, multiple storage destinations, one-click restore (premium).
  - **BlogVault/Jepack Backup:** Real-time backup services with off-site storage.
- **Managed Hosting Backups:** WP Engine, Kinsta, Flywheel, Pantheon all provide automated daily backups with one-click restore. Not WordPress core features.

## 21. Configuration System

- **`wp-config.php`:**
  - PHP file loaded at the start of every request. Defines PHP constants.
  - **Database:** `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_CHARSET`, `DB_COLLATE`, `$table_prefix`.
  - **Security Keys/Salts:** `AUTH_KEY`, `SECURE_AUTH_KEY`, `LOGGED_IN_KEY`, `NONCE_KEY`, `AUTH_SALT`, `SECURE_AUTH_SALT`, `LOGGED_IN_SALT`, `NONCE_SALT`. Used for cookie signing and nonce generation.
  - **Debug:** `WP_DEBUG` (true/false), `WP_DEBUG_LOG` (true or path), `WP_DEBUG_DISPLAY`, `SCRIPT_DEBUG` (load unminified core JS/CSS), `SAVEQUERIES` (log all DB queries).
  - **Core Behavior:** `WP_HOME`, `WP_SITEURL`, `WP_CONTENT_DIR`, `WP_CONTENT_URL`, `WP_PLUGIN_DIR`, `UPLOADS`, `ABSPATH`.
  - **Caching:** `WP_CACHE` (enables `advanced-cache.php` drop-in).
  - **Cron:** `DISABLE_WP_CRON`, `ALTERNATE_WP_CRON`, `WP_CRON_LOCK_TIMEOUT`.
  - **Memory:** `WP_MEMORY_LIMIT` (default `40M`), `WP_MAX_MEMORY_LIMIT` (admin, default `256M`).
  - **Revisions:** `WP_POST_REVISIONS`, `AUTOSAVE_INTERVAL`.
  - **Trash:** `EMPTY_TRASH_DAYS` (default 30).
  - **Updates:** `WP_AUTO_UPDATE_CORE` (`true`, `false`, `minor`), `AUTOMATIC_UPDATER_DISABLED`.
  - **Multisite:** `WP_ALLOW_MULTISITE`, `MULTISITE`, `SUBDOMAIN_INSTALL`, `DOMAIN_CURRENT_SITE`, `PATH_CURRENT_SITE`, `SITE_ID_CURRENT_SITE`, `BLOG_ID_CURRENT_SITE`, `SUNRISE`.
  - **File Editing:** `DISALLOW_FILE_EDIT` (disables theme/plugin editor), `DISALLOW_FILE_MODS` (disables all file modifications including updates).
  - **Filesystem:** `FS_METHOD` (`direct`, `ftpext`, `ssh2`, `ftpsockets`), `FTP_*` constants.
- **Options Table (`wp_options`):**
  - Key-value store for all site settings. ~150 core options.
  - `get_option()`, `update_option()`, `add_option()`, `delete_option()`.
  - `autoload` column: options marked `yes` are loaded into memory on every request (cached as one query). Performance-critical.
  - Serialized PHP arrays/objects supported as values.
- **Customizer API:**
  - Live-preview configuration interface for themes. Being deprecated in favor of Site Editor for block themes, but still functional.
  - Objects: Panels > Sections > Controls > Settings.
  - `$wp_customize->add_panel()`, `add_section()`, `add_setting()`, `add_control()`.
  - Transport modes: `refresh` (reload preview) or `postMessage` (instant JS-based update).
  - Selective refresh: partial DOM updates via `WP_Customize_Selective_Refresh`.
  - Built-in controls: text, textarea, checkbox, radio, select, dropdown-pages, color, media, image, cropped-image.
  - Settings stored in `wp_options` (theme_mods_{theme_name}) or custom locations.
- **Site Health (`/wp-admin/site-health.php`):**
  - Since WordPress 5.1. Two tabs: Status (tests) and Info (system data).
  - Status checks: HTTPS, PHP version, plugin/theme updates, background updates, loopback request, REST API, debug mode, database extensions, page cache, persistent object cache, etc.
  - Info sections: WordPress constants, server (PHP, MySQL, web server), filesystem permissions, database, active theme, active plugins, media handling, and more.
  - Extensible via `site_status_tests` filter.

## 22. Multisite

- **Network Architecture:**
  - Single WordPress installation serves multiple sites. Shared codebase, shared `wp_users` table, per-site content tables (`wp_2_posts`, `wp_2_options`, etc.).
  - Enable via `WP_ALLOW_MULTISITE` constant, then Network Setup wizard.
  - URL modes: subdirectory (`example.com/site1/`) or subdomain (`site1.example.com`). Must choose at setup time; cannot change after.
- **Network Admin (`/wp-admin/network/`):**
  - Separate admin area for the Super Admin. Manages: sites, users, themes, plugins, settings, updates.
  - Create/edit/delete sites. Activate/deactivate/delete users across network.
- **Shared Resources:**
  - **Users:** Single `wp_users` table. Users can have different roles on different sites. Network admin assigns users to sites.
  - **Themes:** Installed once, network-enabled by Super Admin. Site admins choose from network-enabled themes.
  - **Plugins:** Installed by Super Admin. Can be "Network Activated" (active on all sites) or enabled for per-site activation.
- **Domain Mapping (native since 4.5):**
  - Map custom domains to individual sites in the network. Previously required a plugin (WordPress MU Domain Mapping), now built into core.
  - DNS and web server configuration required externally.
- **Per-Site Configuration:**
  - Each site has its own `wp_N_options` table. Independent settings, active theme (from network-enabled list), active plugins (if allowed by network admin).
  - Site admins CANNOT install themes or plugins. Can only activate network-enabled ones.
  - Enable per-site plugin management: Network Settings > "Plugins" checkbox.
- **`sunrise.php` Drop-in:** Loaded before multisite initialization. Used for domain mapping customization, early redirects, and custom logic before site resolution.
- **Limitations:** All sites share the same WordPress version and database server. Schema changes affect all sites. Large networks (1000+ sites) require careful database optimization. No built-in per-site database isolation.

## 23. Caching

- **Object Cache:**
  - `WP_Object_Cache` class. In-memory key-value store during request lifecycle.
  - Functions: `wp_cache_get($key, $group)`, `wp_cache_set($key, $data, $group, $expire)`, `wp_cache_delete()`, `wp_cache_flush()`, `wp_cache_add()` (only if key doesn't exist).
  - Cache groups for namespacing (e.g., `posts`, `options`, `users`). `wp_cache_add_non_persistent_groups()` for groups that should never persist.
  - **Persistent backends** (via `object-cache.php` drop-in):
    - **Redis:** Rich data types, persistence, pub/sub, LUA scripting, eviction policies. Plugin: Redis Object Cache (200K+ installs).
    - **Memcached:** Simple key-value, multi-threaded, distributed. Plugin: W3 Total Cache or Memcached drop-in.
  - WordPress core makes ~500-2000+ `wp_cache_get()` calls per request depending on page complexity.
- **Transients API:**
  - `set_transient($key, $value, $expiration)`, `get_transient($key)`, `delete_transient($key)`.
  - Without persistent object cache: stored in `wp_options` table with expiration (auto-cleaned by cron).
  - With persistent object cache: automatically routed through object cache (bypasses database).
  - Site transients (`set_site_transient()`) -- network-wide in multisite.
- **Page Cache:**
  - NOT built-in to WordPress core.
  - `WP_CACHE` constant + `advanced-cache.php` drop-in enables page caching plugins.
  - Plugins: WP Super Cache (Automattic), W3 Total Cache, WP Fastest Cache, LiteSpeed Cache, WP Rocket (commercial).
  - Concept: serve full HTML output from cache, bypassing PHP/WordPress bootstrap entirely. Huge performance impact (10-100x faster).
  - Managed hosts (Kinsta, WP Engine, Pantheon) provide server-level page caching.
- **Fragment Caching:** No native API. Implemented via object cache manually (cache rendered HTML fragments). Some caching plugins offer this.
- **Browser Caching:** Controlled via `.htaccess` (Apache) or server config (Nginx). Plugins like W3 Total Cache add appropriate `Cache-Control`, `Expires`, `ETag` headers.

## 24. SEO

- **Native SEO Features (limited):**
  - XML Sitemaps: built-in since WordPress 5.5. Auto-generated at `/wp-sitemap.xml`. Includes posts, pages, CPTs, categories, tags, custom taxonomies, author archives. Configurable via `wp_sitemaps_*` filters. Basic -- no priority, frequency, or image sitemaps.
  - `<title>` tag: `add_theme_support('title-tag')` -- WordPress generates page titles from post title, site name, etc. (since 4.1).
  - Canonical URLs: `rel="canonical"` output by `wp_head()` (since 2.9).
  - Robots meta: basic `noindex` for search results, attachment pages, and preview pages.
  - No native: meta descriptions, Open Graph tags, Twitter Cards, structured data/JSON-LD, breadcrumbs, redirects, 301 management, keyword tracking, content analysis, readability scoring, internal linking suggestions.
- **SEO Plugins:**
  - **Yoast SEO** (5M+ installs): Meta titles/descriptions, Open Graph, Twitter Cards, XML sitemaps (extended), breadcrumbs, canonical URLs, redirects (premium), content analysis (keyword density, readability), schema/structured data (basic in free, full in premium), IndexNow integration.
  - **Rank Math** (3M+ installs): 13 schema types free (Article, Product, Recipe, FAQ, HowTo, etc.), advanced sitemap control, 404 monitor, redirections, role manager, image SEO, local SEO, WooCommerce SEO, Google Analytics/Search Console integration. More features in free tier than Yoast.
  - **All in One SEO** (3M+ installs): Schema generator (any schema type), smart sitemaps, social media integration, local SEO, WooCommerce SEO, TruSEO scoring, link assistant.

## 25. Observability / Logging

- **WP_DEBUG System:**
  - `WP_DEBUG = true` -- displays all PHP errors, notices, warnings, and deprecated function usage.
  - `WP_DEBUG_LOG = true` -- logs errors to `wp-content/debug.log`. Can specify custom path: `WP_DEBUG_LOG = '/path/to/debug.log'`.
  - `WP_DEBUG_DISPLAY = false` -- suppresses error display while still logging (recommended for production).
  - `SAVEQUERIES = true` -- stores all database queries in `$wpdb->queries` array with time and caller info. Performance impact.
  - `SCRIPT_DEBUG = true` -- loads unminified `.dev.js` and `.dev.css` core files for debugging.
- **Error Handling:**
  - Since 5.2: Recovery Mode -- detects fatal errors, sends admin email with one-time login link, auto-pauses the offending plugin/theme.
  - `fatal-error-handler.php` drop-in for custom fatal error handling.
  - `wp_die()` for controlled error pages.
- **Site Health (native):**
  - Status tests covering server configuration, PHP version, database connectivity, background updates, REST API availability, loopback requests, HTTPS, debug mode.
  - Info tab provides system details for debugging.
- **Query Monitor (plugin, de facto standard):**
  - Debug panel showing: database queries (count, time, duplicates, caller), PHP errors/warnings/notices, hooks (fired and attached callbacks), HTTP API calls, enqueued scripts/styles, transients, capability checks, conditional checks, environment info.
  - Per-block query data in block editor.
  - Does NOT persist logs (request-scoped only).
- **Debug Log Manager (plugin):** View, filter, and clear `debug.log` from admin. One-click toggle for WP_DEBUG modes.
- **New Relic / Datadog / Application Monitoring:** No native integration. Requires PHP agent installation at server level or plugins like WP New Relic.

## 26. Audit Trail

- **Native Audit Logging:** WordPress has NO built-in audit log or change tracking (beyond revisions for post content).
- **WP Activity Log (plugin, 200K+ installs):**
  - Tracks 400+ events: user logins/logouts/failed logins, post/page CRUD, plugin/theme changes, user profile changes, settings changes, widget changes, menu changes, media operations, taxonomy changes, permalink changes, multisite network events.
  - Per-event: who, what, when, IP address, object ID.
  - Configurable retention (unlimited in premium, 3/6/12 months).
  - Premium: real-time alerts (email, SMS, Slack), external log storage (database, syslog), reports, session management, search and filter.
- **Simple History (plugin, 200K+ installs):**
  - Tracks: post edits, user logins/logouts/profile updates, plugin/theme activations, option changes, media uploads/edits/deletes, Gravity Forms submissions, User Switching events.
  - Lightweight. Dashboard widget + dedicated log page.
  - RSS feed of events. Stealth mode (hidden from admin UI).
  - Stores in default WordPress database (`simple_history` custom tables).
  - No setup required -- starts logging immediately on activation.
- **Stream (plugin):** Tracks user activity with connectors for core, ACF, BuddyPress, Gravity Forms, Jetpack, WooCommerce. External storage connectors.

## 27. Pricing / Licensing

- **License:** GPLv2 (or later). "Free as in freedom" -- use, modify, distribute, sell without restriction.
- **WordPress.org (self-hosted):**
  - Software: free. No per-site, per-user, or per-feature licensing.
  - Costs: domain ($10-20/year), hosting ($3-30+/month), premium plugins ($0-300+/year each), premium themes ($0-100+/year).
  - Total typical cost: $50-500+/year depending on hosting and plugins.
- **WordPress.com (Automattic-hosted):**
  - Free: wordpress.com subdomain, limited storage, ads displayed, no custom plugins.
  - Personal ($4/month): custom domain, no ads, email support.
  - Premium ($8/month): advanced design, CSS customization, monetization.
  - Business ($25/month): plugin/theme installation, SFTP/SSH, staging, 50GB storage.
  - Commerce ($45/month): WooCommerce, payment processing, all Business features.
  - Enterprise (VIP): custom pricing, dedicated infrastructure, SLA, 24/7 support.
- **Plugin/Theme Economics:**
  - GPL: all WordPress plugins and themes must be GPL-licensed (derivative works inherit license).
  - Revenue model: free plugin + premium version (freemium), annual license for support/updates, SaaS tie-in (API keys), one-time purchase.
  - WordPress.org directory: free plugins only (no payment required to list). Plugin reviews by WordPress.org team.
  - Commercial marketplaces: CodeCanyon/ThemeForest (Envato), EDD, Freemius, Gumroad.
  - GPL allows redistribution: "GPL clubs" resell premium plugins legally but without support/updates from original developer.

## 28. Unique Features / Ecosystem

- **Market Share:** Powers ~43% of all websites on the internet (W3Techs, 2025). ~63% of sites using a known CMS. By far the dominant CMS platform.
- **Plugin Ecosystem:** 59,000+ free plugins on wordpress.org. Thousands more commercial plugins. Covers virtually every conceivable website feature (ecommerce, forums, LMS, membership, booking, CRM, etc.).
- **Gutenberg / Full-Site Editing (FSE):**
  - WordPress's most significant architectural evolution. Transitioning from a document editor to a full site builder.
  - Four phases: 1) Editor (done), 2) Customization/FSE (mostly done), 3) Collaboration (in progress, 2025-2026), 4) Multilingual (future).
  - Phase 3 focus: real-time collaboration, multi-user editing, editorial workflows.
- **WordPress Playground:**
  - Run WordPress entirely in the browser via WebAssembly. No server required.
  - PHP compiled to WASM via Emscripten. MySQL replaced with SQLite. Server APIs replaced with JavaScript.
  - Use cases: instant demos, plugin/theme testing, documentation, education, CI/CD testing.
  - Available at playground.wordpress.net. Embeddable in any webpage.
  - PHP 8.3 default (since July 2025), OpCache enabled, XDebug support (experimental), 64-bit integer support, EXIF parsing.
  - SQLite integration passes 99% of WordPress unit tests.
  - Multi-worker support in Node.js for concurrent request handling.
- **REST API as Headless CMS:**
  - WordPress can serve as a headless backend with any frontend framework (Next.js, Nuxt, Gatsby, Astro, SvelteKit, etc.).
  - WPGraphQL plugin provides GraphQL alternative to REST.
  - Faust.js (by WP Engine) -- headless WordPress framework for Next.js.
  - Common pattern: WordPress admin for content management + decoupled React/Vue/Svelte frontend.
- **WordPress.com Hosting:**
  - Automattic operates the largest WordPress hosting infrastructure.
  - WordPress VIP: enterprise-tier hosting for large publishers (Time, TechCrunch, Salesforce, Facebook Newsroom).
- **Backwards Compatibility:**
  - WordPress maintains extreme backwards compatibility. Plugins from 10+ years ago often still work.
  - `_deprecated_function()`, `_deprecated_argument()` -- formal deprecation process with long transition periods.
  - This is both a strength (stability, trust) and a weakness (legacy code, slower evolution).
- **Community:**
  - WordCamps: hundreds of annual conferences worldwide.
  - Contributor days, Five for the Future program (companies contribute 5% of resources to WordPress development).
  - Make.wordpress.org: organized contributor teams (Core, Design, Mobile, Polyglots, Meta, Plugins, Themes, Documentation, Community, etc.).
- **Limitations / Criticisms:**
  - Monolithic PHP architecture -- no microservices, no async processing, no background workers without workarounds.
  - Everything-is-a-post model: `wp_posts` table stores posts, pages, revisions, attachments, menu items, reusable blocks, custom CSS, changesets, templates, template parts, navigation, fonts -- leads to bloated tables and semantic confusion.
  - EAV pattern for metadata (`wp_postmeta`, `wp_usermeta`) -- poor query performance at scale, no type enforcement, no indexing on values by default.
  - No native multi-tenancy beyond Multisite (which shares codebase and database).
  - Security: large attack surface due to plugin ecosystem. WordPress is the most targeted CMS.
  - No built-in: audit logging, webhooks, multi-language content, backup, deployment pipeline, job queue, real-time features.

---

*Document compiled February 2026. Based on WordPress 6.7-6.9 (latest stable), official developer documentation, and community sources.*
