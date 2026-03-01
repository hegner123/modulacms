# Wagtail CMS Features (v7.3)

## Runtime Architecture
- Built on **Django** (Python web framework). Wagtail 7.3 (current stable, released Feb 2, 2026) requires Django 5.1+
- Runs as a standard **WSGI** application. ASGI supported but Wagtail does not natively implement async views, so WSGI (Gunicorn, uWSGI) is recommended
- Process model: standard Django request/response cycle. Each WSGI worker handles one request at a time
- **Background tasks**: Starting in Wagtail 6.4, `django-tasks` integrated for background worker support. Pluggable backend interface for task queues (Celery, Django RQ, custom). The `django-tasks` framework promoted into Django core (Django 6.0)
- **Scheduled publishing** requires the `publish_scheduled` management command run periodically via cron (recommended: every 5-10 minutes). No built-in scheduler daemon

## Database Support
- Uses the **Django ORM** for all database access:
  - **PostgreSQL** (recommended for production, full-text search via `django.contrib.postgres`)
  - **MySQL / MariaDB** (supported and tested)
  - **SQLite** with JSON1 extension (development only)
- **Search backends** (separate from database):
  - **Database search backend** (default): native full-text search (PostgreSQL FTS, SQLite FTS5, MySQL FULLTEXT)
  - **Elasticsearch 7, 8, 9**: full-featured with faceting, relevance ranking, autocomplete
  - **OpenSearch 2, 3**: officially supported
  - `update_index` management command rebuilds search index

## Content Management
- **Page model**: All pages inherit from `wagtail.models.Page`. Page tree via **django-treebeard** using **Materialised Path** algorithm. Fields: `path`, `depth`, `numchild`. Tree operations O(1) to O(log n). `fixtree` command repairs corrupted tree state
- **StreamField**: Block-based freeform content stored as JSON. Add, remove, reorder, and nest blocks of different types. Drag-and-drop reordering (6.4), block previews (6.4)
- **Snippets**: Django models registered with `@register_snippet`. Optional mixins: `RevisionMixin`, `DraftStateMixin`, `PreviewableMixin`, `LockableMixin`, `WorkflowMixin`
- **Images**: Managed via image library with upload, tagging, focal point setting, alt text (accessibility-focused `ImageBlock` added 6.3)
- **Documents**: File upload and management with configurable serving method
- **Orderable inline panels**: Models with `ParentalKey` + `Orderable` for drag-and-drop child reordering via `InlinePanel`
- **Custom models**: Any Django model managed via `ModelViewSet` or `SnippetViewSet`

## Content Schema / Content Modeling
- **Code-first approach**: Content types defined as Python classes inheriting from `Page` or Django `Model`. No GUI schema builder
- **Django field types**: All standard fields available (`CharField`, `TextField`, `IntegerField`, `DateField`, `ForeignKey`, `BooleanField`, `URLField`, `EmailField`, etc.)
- **StreamField block types**:
  - **Text/input**: `CharBlock`, `TextBlock`, `EmailBlock`, `IntegerBlock`, `FloatBlock`, `DecimalBlock`, `RegexBlock`, `URLBlock`, `BooleanBlock`, `DateBlock`, `TimeBlock`, `DateTimeBlock`, `BlockQuoteBlock`
  - **Rich content**: `RichTextBlock` (Draftail WYSIWYG), `RawHTMLBlock`
  - **Choice**: `ChoiceBlock`, `MultipleChoiceBlock`
  - **Chooser**: `PageChooserBlock`, `DocumentChooserBlock`, `ImageChooserBlock`, `ImageBlock` (accessibility-focused, 6.3), `SnippetChooserBlock`, `EmbedBlock`
  - **Structural**: `StructBlock`, `ListBlock`, `StreamBlock` (nestable), `StaticBlock`
  - **Contrib**: `TableBlock`, `TypedTableBlock` (cells are arbitrary block types)
  - **Grouping**: `BlockGroup` (visual grouping, 7.3)
  - Custom blocks by subclassing any block type
- **Panel types**: `FieldPanel`, `InlinePanel`, `MultiFieldPanel`, `FieldRowPanel`, `TabbedInterface`, `HelpPanel`, `FormSubmissionsPanel`
- **Page types**: Each `Page` subclass is a distinct type. Restrict child types via `subpage_types` and parent types via `parent_page_types`

## Content Delivery / API
- **Wagtail API v2**: Built-in REST API powered by **Django REST Framework**:
  - `/api/v2/pages/` -- pages with type filtering, field selection, search
  - `/api/v2/images/` -- image metadata and rendition URLs
  - `/api/v2/documents/` -- document metadata and download URLs
- **Filtering**: `?type=`, `?descendant_of=`, `?child_of=`, `?ancestor_of=`, `?site=`, `?locale=`, `?search=`, `?fields=`, custom field filters via `api_fields`
- **Ordering**: `?order=` with `-` prefix for descending
- **Pagination**: `?limit=` and `?offset=`. Response includes `meta.total_count`. Configurable max via `WAGTAILAPI_LIMIT_MAX`
- **Custom API endpoints**: Extend `BaseAPIViewSet` for any model
- **Headless mode**: Official documentation for headless deployment (6.4)
- **GraphQL**: Via **wagtail-grapple** (maintained by Wagtail core team). Auto-generates schema from Wagtail models including StreamField

## Content Publishing
- **Draft/Live workflow**: Pages and snippets (with `DraftStateMixin`) have distinct draft and live states
- **Publish/Unpublish**: Explicit actions. `live_revision` pointer tracks published version
- **Save draft**: Sets `has_unpublished_changes = True`
- **Scheduled publishing**: Set `go_live_at` datetime. `publish_scheduled` management command (cron) publishes at time. History view shows scheduled revisions with "Unschedule" option
- **Scheduled unpublishing**: Set `expire_at` datetime. Auto-unpublished by management command
- **Moderation workflow**: Configurable multi-stage approval:
  - **Workflow**: Named sequence of `WorkflowTask` instances, assigned to page trees or snippet models
  - **WorkflowTask**: Abstract base. Built-in: `GroupApprovalTask`. Custom types by subclassing
  - **WorkflowState/TaskState**: Track per-object workflow progress
  - Actions: approve, reject (with comment), cancel, resume after rejection
- **Concurrent editing notifications** (6.2): Notifications when multiple users edit same page/snippet. Periodic pings (default 10 seconds)
- **Autosave** (7.3): Automatic draft saving. Uses concurrent editing detection for conflicts. Configurable interval via `WAGTAIL_AUTOSAVE_INTERVAL`. Enabled by default
- **Page locking**: Lock by editors with lock permission. `WAGTAILADMIN_GLOBAL_EDIT_LOCK` prevents all others from editing

## Content Versioning
- **Revisions**: Every edit creates a `Revision` storing full JSON snapshot. Stored indefinitely unless purged
- **live_revision**: Points to currently published revision
- **latest_revision**: Points to most recent (may be draft)
- **has_unpublished_changes**: Boolean, `True` when latest differs from live
- **Revision comparison**: Side-by-side diff view showing field-level changes between any two revisions
- **Restore**: "Review this revision" then "Replace current version" creates new revision with old content (non-destructive)
- **PageLogEntry**: Audit model recording all page actions with timestamps and user
- **ModelLogEntry**: Generic audit model for non-page models
- **purge_revisions**: Management command to delete old revisions by age and model type. Preserves live, in-moderation, and scheduled revisions

## Internationalization (i18n)
- **Core i18n** (built-in since 2.11):
  - `TranslatableMixin`: Adds `locale` and `translation_key` (UUID) fields
  - `Locale` model: BCP-47 language codes
  - **Separate page tree per locale**: Each locale has its own complete page tree linked via `translation_key` UUID
  - URL routing via Django's `i18n_patterns()` with language prefix
  - `LocaleMiddleware` for browser language detection
- **wagtail.contrib.simple_translation** (built-in): Copy pages/snippets to new locale as drafts in source language. Manual translation and publish
- **wagtail-localize** (official external package):
  - **Segment-based translation**: Breaks content (including StreamField) into translatable string segments
  - In-admin translation editing UI
  - **PO file import/export** for professional translation workflows
  - **Machine translation**: DeepL, Google Cloud Translate, LibreTranslate (self-hostable)
  - Sync/translation workflows with per-segment status tracking
  - Background task support via Django RQ or django-tasks

## Media Management
- **Image library**: Upload, browse, search, tag, organize. Custom `Image` model via `WAGTAILIMAGES_IMAGE_MODEL`
- **Focal point**: Manual or automatic. Stored as x/y/width/height fields. Used by `fill` crop operation
- **Renditions**: Auto-generated on demand and cached. Operations:
  - `width-N`, `height-N`, `max-WxH`, `min-WxH`, `fill-WxH` (focal-point-centered crop)
  - `fill-WxH-cN` -- crop closeness to focal point
  - Format: `format-jpeg`, `format-png`, `format-gif`, `format-webp`, `format-avif`, lossless variants, `format-ico`
  - Quality: `jpegquality-N`, `webpquality-N`, `avifquality-N`
  - `bgcolor-RRGGBB` for transparent images
- **Template tag**: `{% image page.photo fill-400x400 format-webp as photo %}`
- **Feature detection**: Optional OpenCV/Rustface-based face detection for automatic focal points
- **Document management**: Upload, tag, organize. Custom model. Configurable serving method
- **Collections**: Hierarchical folder-like organization. Permissions propagate down tree. Control add/edit/delete/choose per collection

## Authentication
- Built on **Django's auth system**. Session-based authentication
- **Custom user models** via `AUTH_USER_MODEL`
- **SSO**: Via `social-auth-app-django` or `django-allauth` (Google, GitHub, Azure AD, etc.)
- **2FA**: Via **wagtail-2fa** (TOTP, supports Authy/Google Authenticator/1Password). `WAGTAIL_2FA_REQUIRED` forces all admin users
- **Password management**: Configurable via `WAGTAIL_PASSWORD_MANAGEMENT_ENABLED`, `WAGTAIL_PASSWORD_RESET_ENABLED`
- **Frontend login**: Configurable URL and template for private pages

## Authorization / Permissions
- **Page-level permissions**: Attached at any tree point, propagate downward. Types: Add, Edit, Publish, Bulk delete, Lock
- **Collection-based permissions**: For images and documents. Types: Add, Edit, Delete, Choose, Collection management
- **Group-based**: All permissions assigned through Django groups
- **Custom permissions**: Via `register_permissions` hook
- **Workflow permissions**: Restricted to specific groups via `GroupApprovalTask`
- **Private pages**: Restrict subtrees by shared password, login required (specific groups or any authenticated user)
- **Admin access**: Django `is_staff` flag. Superusers bypass all checks

## Admin Panel (Wagtail Admin)
- **Tech stack**: Hybrid Django templates + Stimulus (JavaScript). Complex components use React (sidebar, Draftail, StreamField chooser)
- **Slim sidebar** (since 2.16): Collapsible navigation
- **Dark mode** (since 5.0): Light, dark, or system preferences
- **Draftail rich text editor**: Based on Draft.js. Extensible via `register_rich_text_features` hook. Three extension types: inline styles, blocks, entities. Plus Controls, Decorators, Plugins APIs. Toolbar configurable per field
- **Reports** (built-in): Locked pages, Aging pages, Site history, Workflows, Workflow tasks. Exportable. Custom reports via `register_reports_menu_item` hook
- **Content metrics** (6.2): Reading time, readability score, word count
- **Accessibility checker** (5.0+): Built-in Axe-based checker. Alt text, heading structure, color contrast
- **Snippets listing**: Search, filtering, bulk actions via `SnippetViewSet`
- **Custom admin views**: Via `register_admin_viewset` or `register_admin_urls` hooks

## Extension System
- **Django apps as extensions**: Any Django app integrates via `INSTALLED_APPS`
- **Hooks system**: 60+ named hooks. Categories:
  - Admin modules, editor workflow, page explorer, choosers, snippets, page serving, document serving, bulk actions, editor interface, audit, images, users
- **Custom StreamField blocks**: Subclass existing blocks or `Block` base. Custom React editing components
- **ViewSets**: `ModelViewSet` and `SnippetViewSet` for registering arbitrary Django models
- **Custom bulk actions**: Via `register_bulk_action` hook
- **Form builder** (`wagtail.contrib.forms`): Editors build forms via admin UI. Submissions stored and viewable

## Webhooks
- **No built-in webhook system**
- **Django signals**: `page_published`, `page_unpublished`, `pre_page_move`, `post_page_move`, standard Django signals. Connect to custom handlers
- **Hooks** provide alternative extension points
- **Community packages**: `wagtail-webhook` exists but not officially maintained

## SDKs & Client Libraries
- **No official SDK** for any language. REST/GraphQL consumed directly
- **Community**: `wagtail-js` (npm), `wagtail-spa-integration` (PyPI), `wagtail-headless-preview` (Torchbox)

## CLI / Management Commands
- `wagtail start <project_name>` -- scaffold new project
- `publish_scheduled` -- publish/unpublish objects with scheduled dates (run via cron)
- `fixtree` -- repair page tree errors
- `move_pages` -- relocate pages between subtrees
- `update_index` -- rebuild search index
- `purge_revisions` -- delete old revisions (preserves live/moderation/scheduled)
- `purge_embeds` -- clear cached embeds
- `rebuild_references_index` -- populate cross-reference tracking
- `wagtail_update_image_renditions` -- regenerate renditions
- Standard Django commands: `createsuperuser`, `migrate`, `collectstatic`, `dumpdata`, `loaddata`

## Deployment
- **Standard Django deployment**: Gunicorn/uWSGI behind Nginx/Apache
- **Docker**: Well-documented. `bakerydemo` reference setup
- **PaaS**: Heroku, Fly.io, PythonAnywhere, Digital Ocean App Platform
- **Static site generation**: `wagtail-bakery` exports as static HTML
- **Media storage**: Local filesystem default. S3/cloud via `django-storages`
- **Frontend cache invalidation**: `wagtail.contrib.frontend_cache` (Varnish, Squid, Cloudflare, CloudFront)

## Email
- Built on **Django's email framework**
- **SMTP**: Standard Django settings
- **Notification emails**: Sent on workflow actions (submit, approve, reject). `WAGTAILADMIN_NOTIFICATION_FROM_EMAIL` configurable
- **Notification format**: Plain text default, HTML via `WAGTAILADMIN_NOTIFICATION_USE_HTML`
- **Form builder emails**: `AbstractEmailForm` sends submissions to configured addresses

## Backup & Restore
- **No built-in CMS backup tool**
- **Django dumpdata/loaddata**: JSON fixtures. Known issues with StreamField
- **Database-level**: `pg_dump`/`pg_restore`, `mysqldump`, file copy. More reliable
- **Media backup**: Separate filesystem/S3 backup required
- **Revision recovery**: Deleted pages sometimes recoverable via revisions

## Configuration System
- **Django settings.py**: Primary configuration. All Wagtail settings are Python constants
- **Key WAGTAIL_* settings**: `WAGTAIL_SITE_NAME`, `WAGTAILADMIN_BASE_URL`, `WAGTAIL_I18N_ENABLED`, `WAGTAIL_CONTENT_LANGUAGES`, `WAGTAILSEARCH_BACKENDS`, `WAGTAILIMAGES_IMAGE_MODEL`, `WAGTAIL_WORKFLOW_ENABLED`, `WAGTAIL_AUTOSAVE_INTERVAL`, `WAGTAILADMIN_COMMENTS_ENABLED`, `WAGTAILADMIN_GLOBAL_EDIT_LOCK`
- **Environment variables**: No built-in loading. Via `django-environ` or `python-decouple`
- **Site settings** (`wagtail.contrib.settings`): Admin-editable. `BaseSiteSetting` (per-site) and `BaseGenericSetting` (global)

## Search
- **Unified API**: `search()` method on QuerySets across all backends
- **Search backends**: Database (default), Elasticsearch 7/8/9, OpenSearch 2/3
- **Custom search fields**: `index.SearchField` (full-text), `index.AutocompleteField` (partial), `index.FilterField` (exact), `index.RelatedFields` (across relations)
- **Autocomplete**: Separate `autocomplete()` method
- **Search promotions** (`wagtail.contrib.search_promotions`): Editor-curated "Editors' Picks" for specific terms
- **Search query logging**: For analytics. Configurable retention

## Observability / Logging
- **Django logging**: Standard Python `logging` module
- **PageLogEntry/ModelLogEntry**: Record all actions with timestamp, user, action type, data
- **No built-in metrics/APM**. Standard Django monitoring approaches apply

## Audit Trail
- **PageLogEntry**: 25+ action types: create, edit, delete, publish, unpublish, schedule, lock, unlock, rename, revert, copy, move, reorder, workflow events
- **ModelLogEntry**: Same system for non-page models
- **Site History report**: All logged actions, filterable by user, action type, date range. Exportable
- **Custom log actions**: Via `register_log_actions` hook

## Pricing / Licensing
- **BSD 3-Clause License**: Fully open source. Free for any purpose including commercial
- **No paid tier**: No premium features, no enterprise edition
- **Commercial support**: From Torchbox (creators) and network of agencies

## Unique Features
- **StreamField**: Block-based freeform content with nested, reorderable, mixed-type blocks. JSON storage. Drag-and-drop (6.4)
- **Page tree with django-treebeard**: Materialised path for efficient tree operations. Enforced parent/child type constraints
- **Workflow system**: Multi-stage, multi-reviewer approval. Custom task types. Assignable to page subtrees or snippet models
- **Draftail editor**: Draft.js-based with Python-hook extensibility. Three extension categories plus Controls, Decorators, Plugins
- **TableBlock / TypedTableBlock**: HTML table editor. TypedTableBlock allows cells to contain arbitrary StreamField block types
- **Form builder** (`wagtail.contrib.forms`): Editors create forms without developer involvement
- **Contrib packages**: `routablepage`, `redirects`, `search_promotions`, `sitemaps`, `frontend_cache`, `simple_translation`, `settings`, `typed_table_block`
- **Content metrics** (6.2): Reading time, readability, word count
- **Accessibility checker**: Built-in Axe-based checking (5.0+)
- **Concurrent editing + Autosave** (6.2/7.3): Real-time multi-editor awareness with automatic draft saving
- **Snippet mixins**: Modular feature composition (RevisionMixin, DraftStateMixin, PreviewableMixin, LockableMixin, WorkflowMixin)
