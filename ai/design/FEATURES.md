# Admin Panel Feature Specification

Reference spec for all ModulaCMS admin panel designs. Every design variant must implement these features.

## Authentication

### Login
- Email/password form
- Rate-limited (10/min per IP)
- Redirect to dashboard if already authenticated
- CSRF protected

### Forgot Password
- Email input form
- Sends reset link via email
- Success message regardless of email existence (security)

### Reset Password
- Token-validated form
- New password + confirmation inputs
- Redirect to login on success

### Logout
- Session destruction
- Redirect to login

---

## Navigation

### Sidebar (12 items, permission-gated)

| Item | Path | Permission |
|------|------|------------|
| Dashboard | `/admin/` | None |
| Content | `/admin/content` | `content:read` |
| Media | `/admin/media` | `media:read` |
| Datatypes | `/admin/schema/datatypes` | `datatypes:read` |
| Routes | `/admin/routes` | `routes:read` |
| Users | `/admin/users` | `users:read` |
| Roles | `/admin/users/roles` | `roles:read` |
| Tokens | `/admin/users/tokens` | `tokens:read` |
| Plugins | `/admin/plugins` | `plugins:read` |
| Import | `/admin/import` | `import:create` |
| Audit | `/admin/audit` | None |
| Settings | `/admin/settings` | `config:read` |

Conditionally visible:
- **Locales** (`/admin/settings/locales`) — shown only when i18n enabled, requires `locale:read`

### Topbar
- Logo / site title
- Sidebar collapse toggle
- Theme toggle (light/dark)
- User dropdown: username display, logout action

---

## Dashboard

- Overview metrics (content count, media count, recent activity)
- Quick action links

---

## Content Management

### Content List
- Paginated table of top-level content items
- Columns: title (resolved from route title > title field > truncated ID), status badge, datatype, author, modified date
- Filters: status (all/draft/published), search by title/slug
- Row actions: edit, view, delete
- Create button: opens dialog with datatype selector + optional parent

### Content Editor
- **Tree panel** (left): hierarchical block tree with expand/collapse, drag-drop reordering
- **Field panel** (right): dynamic field inputs rendered per field type
- **Toolbar**: save, publish/unpublish, version history, metadata
- **Metadata display**: author, modified date, status badge, datatype label
- **i18n**: locale selector dropdown (when enabled), per-locale publishing

### Content Operations
- **Create**: select datatype, optional parent node
- **Update**: batch field value updates
- **Delete**: recursive (removes all descendants), confirmation required
- **Reorder**: drag-drop updates sibling pointers (next/prev)
- **Move**: reassign to new parent node
- **Tree save**: bulk pointer update (parent, first_child, next/prev sibling)

### Publishing
- **Publish**: creates version snapshot, sets status to published
- **Unpublish**: reverts status to draft
- **Per-locale publish**: when i18n enabled, publish/unpublish per locale
- Permission: `content:publish`

### Version History
- Paginated list of snapshots (label, date, author)
- Side-by-side diff comparison (select two versions)
- Manual snapshot creation with custom label
- Restore from any previous version

---

## Media Management

### Media List
- Grid layout (thumbnails for images, icons for other types)
- Pagination with configurable page size
- Upload button: opens file upload dialog
- **Picker mode**: minimal grid without full layout, for embedding in field inputs

### Media Detail
- Preview (image/video player/document icon)
- Edit form: alt text, caption, description, display name
- Focal point selector (x, y coordinates for crop guidance)
- Usage info: which content items reference this media
- Delete button (blocked if referenced by content)

### Media Upload
- Multipart file upload
- Drag-drop file input with image preview
- Progress indicator during upload
- Auto-generates preset image dimensions
- Fields: file, alt, caption, description, display_name

### Media Admin
- Storage health check (S3/local status, capacity)
- Cleanup unreferenced files
- Permission: `media:admin`

---

## Schema Management

### Datatypes

#### Datatypes List
- Paginated table
- Columns: label, type, created, modified
- Create button: opens form dialog (name, label, type)
- Row actions: edit, delete

#### Datatype Detail
- Edit form: name, label, type
- Linked fields section: ordered list of fields with sort position
- Add field button: opens inline form (name, label, type, data/validation/ui_config as JSON)
- Unlink field action per row
- Delete blocked if content exists using this datatype

### Fields

#### Fields List
- Paginated table
- Columns: label, type, linked datatypes (count), created
- Row actions: edit, delete

#### Field Detail
- Edit form: name, label, type
- JSON editors: data config, validation config, UI config
- Translatable checkbox (visible only when i18n enabled)
- Linked datatypes list (read-only, shows which datatypes use this field)
- Delete: unlinks from all datatypes (cascade)

### Field Types Reference (read-only)
- Table listing all 14 supported field types with descriptions:

| Type | Description |
|------|-------------|
| text | Single-line plain text |
| textarea | Multi-line plain text |
| richtext | Rich text with configurable toolbar |
| number | Integer or decimal |
| boolean | True/false toggle |
| date | Date only |
| datetime | Date and time |
| select | Dropdown from predefined options |
| media | Media file reference |
| relation | Reference to another content entry |
| json | Freeform JSON |
| slug | URL-friendly identifier (auto-generated) |
| email | Email with format validation |
| url | URL with format validation |

---

## Routes

### Routes List
- Paginated table
- Columns: slug, title, type (public/admin), created, modified
- Create button: opens form dialog (slug, title, status)
- Row actions: edit, delete, browse (public URL link)
- Delete blocked if content assigned to route

### Route Detail
- Edit form: slug, title, status code
- Slug rename support

---

## User Management

### Users List
- Full table (no pagination — expected low user count)
- Columns: username, email, role, created, status
- Create button: opens form dialog (username, email, password, role dropdown)
- Row actions: edit, delete

### User Detail
- Edit form: username, email, role (dropdown)
- Change password form (separate section)
- OAuth connections display (read-only: Google, GitHub, Azure)
- Delete button (blocked: cannot delete self, cannot delete last admin)

---

## Roles & Permissions

### Roles List (split layout)
- **Left panel**: scrollable list of roles, create button
- **Right panel**: selected role detail

### Role Detail
- Edit form: name, label
- System-protected roles (admin, editor, viewer) cannot be renamed or deleted
- Permission matrix: grid of resources x operations with checkboxes
- Toggle individual permissions on/off (async save)

### Permission Matrix

| Resource | Operations |
|----------|-----------|
| content | read, create, update, delete, publish |
| datatypes | read, create, update, delete |
| fields | read, create, update, delete |
| media | read, create, update, delete, admin |
| routes | read, create, update, delete |
| users | read, create, update, delete |
| roles | read, create, update, delete |
| tokens | read, create, delete |
| sessions | read, delete |
| field_types | read |
| admin_tree | read |
| config | read, update |
| plugins | read, admin |
| deploy | read, create |
| import | create |
| locale | read, create, update, delete |
| ssh_keys | read, create, delete |

**47 total permissions.**

### Bootstrap Roles (system-protected)
- **admin**: all 47 permissions, bypasses checks
- **editor**: 28 permissions (content CRUD + media + publishing)
- **viewer**: 3 read-only permissions

---

## API Tokens

### Tokens List
- Paginated table
- Columns: token (truncated with copy button), created, expires
- Create button: generates random 32-byte hex token, displays full value once
- Row actions: copy full token, delete (revoke)

---

## Sessions

### Sessions List
- Table of active HTTP sessions
- Columns: username, IP address, created, expires
- Row action: revoke (immediate logout for that session)

---

## Plugins (placeholder)

### Plugins List
- List installed plugins
- Enable/disable toggle per plugin
- Configure plugin settings

### Plugin Route Management
- List plugin-registered HTTP routes with approval status
- Approve/revoke routes
- Permission: `plugins:admin`

---

## Import

### Import Page
- File upload form (max 32 MB)
- Format selector dropdown:
  - Contentful
  - Sanity
  - Strapi
  - WordPress (JSON export)
  - Clean (generic JSON)
- Submit: bulk import, toast notification on result

---

## Audit Log

### Audit List (read-only)
- Paginated table, reverse chronological
- Columns: operation (create/update/delete), entity type, entity ID, author, timestamp
- Expandable row: JSON diff (old values / new values)
- Metadata: IP address, user agent, request ID
- No CRUD — read-only history

---

## Settings

### Settings Form
- Grouped config sections:

| Section | Fields |
|---------|--------|
| General | environment, port, ssl_port, ssh_host, ssh_port, client_site, admin_site, log_path, output_format, node_id, space_id |
| Database | db_driver, db_url, db_name, db_username, db_password |
| Storage | bucket_region, bucket_media, bucket_backups, bucket_cdn_url |
| OAuth | google (id, secret), github (id, secret), azure (id, secret) |
| CORS | cors_origins, cors_allow_credentials |
| Email | smtp_host, smtp_port, smtp_user, smtp_password, smtp_from, smtp_reply_to |
| Richtext | richtext_toolbar (configurable list) |
| i18n | i18n_enabled, i18n_default_locale |

- Hot-reloadable: changes apply without restart
- Toast notification on save

---

## Locale Settings (conditional on i18n)

### Locales List
- Table: code, name, default flag, enabled flag
- Create button: form dialog (code, name, is_default, is_enabled)
- Row actions: edit, delete
- Cannot delete the default locale

---

## Shared UI Patterns

Every design must implement these interaction patterns:

### Dialogs
- Modal dialogs for create/edit forms
- Confirmation dialogs for destructive actions (delete)
- Escape key closes dialog

### Toasts
- Position configurable (default: top-right)
- Types: success, error, info
- Auto-dismiss with configurable duration

### Tables
- Sortable columns
- Pagination (prev/next, page numbers)
- Empty state message when no data

### Forms
- Label + input + validation message pattern
- CSRF token on all mutating forms
- Submit button with loading state

### Field Renderer
- Dynamic input rendering based on field type
- Must support all 14 field types
- Richtext: configurable toolbar
- Media: inline picker (opens media grid in modal)
- Relation: content selector
- JSON: code editor or textarea
- Select: dropdown from options defined in field config

### Tree Navigation (content editor)
- Hierarchical expand/collapse
- Drag-drop reordering
- Node selection highlights active block
- Visual indicators for draft vs published nodes

### Search
- Debounced input
- Filters table/grid results

### Theme
- Dark and light modes
- Toggle persisted to localStorage
- CSS custom property based (design tokens)

---

## Security Requirements

- CSRF: double-submit cookie on all mutating requests
- Session-based authentication (cookie)
- Permission checks on every handler (fail-closed: missing permission = 403)
- Rate limiting on auth endpoints
- Admin bypass via role flag, not wildcard permissions
- No credentials/tokens/PII in logs
