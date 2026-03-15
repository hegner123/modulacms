# Admin Panel

ModulaCMS includes a server-rendered admin panel built with HTMX and templ. There is no React or single-page application -- all pages are rendered on the server and delivered as HTML. HTMX handles interactive updates by swapping page fragments without full reloads. The panel is accessible at `/admin/` and provides management screens for content, schema, media, users, roles, sessions, tokens, routes, plugins, import, and settings.

## Accessing the Admin Panel

Navigate to `http://localhost:8080/admin/` in your browser. The root URL (`/`) redirects to `/admin/` automatically.

If you are not authenticated, you are redirected to the login page at `/admin/login`. After logging in, you are redirected back to the page you originally requested (tracked via the `?next=` query parameter).

### Login

The admin panel uses cookie-based sessions for authentication. Enter your email and password on the login page. On success, the server sets an HTTP-only session cookie and redirects you to the dashboard.

The login endpoint is rate-limited to 10 attempts per minute per IP address.

### Logout

Click the logout action in the admin panel, or POST to `/admin/logout`. This clears the session cookie.

## Dashboard

The dashboard at `/admin/` is the landing page after login. It requires authentication but no specific permission -- any authenticated user can view it.

## Management Screens

Each screen corresponds to a resource in ModulaCMS. Access is controlled by RBAC permissions -- if your role lacks the required permission, you receive a 403 response.

### Content

**Path:** `/admin/content`

List, create, edit, and delete content entries. Content is organized in a tree structure with parent-child relationships. The content screen supports:

- Listing content with tree navigation
- Creating new content entries
- Editing content entries and their fields
- Reordering content within the tree
- Moving content between parent nodes
- Saving the full content tree structure

| Action | Required Permission |
|--------|-------------------|
| View list | `content:read` |
| Create | `content:create` |
| Edit | `content:update` |
| Delete | `content:delete` |
| Reorder / move / tree save | `content:update` |

### Schema: Datatypes

**Path:** `/admin/schema/datatypes`

Datatypes define the structure of your content. Each datatype has a label, type, and a set of linked fields. The schema screen lets you manage datatypes and their field assignments.

- List all datatypes
- View datatype details with linked fields
- Create, edit, and delete datatypes
- Add fields to a datatype

| Action | Required Permission |
|--------|-------------------|
| View list | `datatypes:read` |
| Create | `datatypes:create` |
| Edit | `datatypes:update` |
| Delete | `datatypes:delete` |
| Add field to datatype | `fields:create` |

### Schema: Fields

**Path:** `/admin/schema/fields/{id}`

Fields define individual pieces of data within a datatype (text, number, image, etc.). Field management is accessed through the datatype detail screen rather than a standalone list.

- View field details
- Edit field configuration
- Delete fields

| Action | Required Permission |
|--------|-------------------|
| View | `fields:read` |
| Edit | `fields:update` |
| Delete | `fields:delete` |

### Schema: Field Types

**Path:** `/admin/schema/field-types`

View the available field types (text, number, image, etc.) that can be used when creating fields.

| Action | Required Permission |
|--------|-------------------|
| View list | `field_types:read` |

### Media

**Path:** `/admin/media`

Upload, view, edit, and delete media assets. The media screen integrates with S3-compatible storage for file management.

- List all media with thumbnails
- View media details (metadata, URLs, srcset)
- Upload new media files
- Edit metadata (display name, alt text, caption, description)
- Delete media (removes S3 objects and database record)

| Action | Required Permission |
|--------|-------------------|
| View list | `media:read` |
| Upload | `media:create` |
| Edit metadata | `media:update` |
| Delete | `media:delete` |

### Routes

**Path:** `/admin/routes`

Routes map URL slugs to content. Manage public-facing routes and admin routes separately.

- List public routes at `/admin/routes`
- List admin routes at `/admin/routes/admin`
- Create, edit, and delete routes

| Action | Required Permission |
|--------|-------------------|
| View list | `routes:read` |
| Create | `routes:create` |
| Edit | `routes:update` |
| Delete | `routes:delete` |

### Users

**Path:** `/admin/users`

Manage user accounts. View user details, create new users, edit profiles, and delete accounts.

| Action | Required Permission |
|--------|-------------------|
| View list | `users:read` |
| Create | `users:create` |
| Edit | `users:update` |
| Delete | `users:delete` |

### Roles

**Path:** `/admin/users/roles`

Manage roles and their permission assignments. Create custom roles, view role details with assigned permissions, and modify role-permission mappings.

- List all roles
- Create new roles with a permission selection form
- View role details with assigned permissions
- Edit role label and permissions
- Delete non-system-protected roles

| Action | Required Permission |
|--------|-------------------|
| View list | `roles:read` |
| Create | `roles:create` |
| Edit | `roles:update` |
| Delete | `roles:delete` |

### Tokens

**Path:** `/admin/users/tokens`

Manage API keys and other tokens. Create new tokens and delete existing ones.

| Action | Required Permission |
|--------|-------------------|
| View list | `tokens:read` |
| Create | `tokens:create` |
| Delete | `tokens:delete` |

Note: Token permissions (`tokens:read`, `tokens:create`, `tokens:delete`) are enforced at the route level but are not included in the bootstrap editor or viewer roles. Only admin users (who bypass permission checks) can access token management by default. To grant token access to non-admin roles, create the corresponding permissions and assign them.

### Sessions

**Path:** `/admin/sessions`

View active sessions and terminate them.

| Action | Required Permission |
|--------|-------------------|
| View list | `sessions:read` |
| Delete | `sessions:delete` |

### Plugins

**Path:** `/admin/plugins`

View installed plugins and their details.

| Action | Required Permission |
|--------|-------------------|
| View list | `plugins:read` |
| View detail | `plugins:read` |

Note: Plugin permissions (`plugins:read`, `plugins:admin`) are enforced at the route level but are not included in the bootstrap roles. Only admin users can access plugin management by default.

### Import

**Path:** `/admin/import`

Import content from other CMS platforms. The import screen accepts data in Contentful, Sanity, Strapi, WordPress, and generic bulk formats.

| Action | Required Permission |
|--------|-------------------|
| View page | `import:read` |
| Submit import | `import:create` |

Note: Import permissions (`import:read`, `import:create`) are enforced at the route level but are not included in the bootstrap roles. Only admin users can access import functionality by default.

### Audit Log

**Path:** `/admin/audit`

View the audit log of all changes made through the admin panel. The audit log screen requires authentication but does not require a specific permission.

### Settings

**Path:** `/admin/settings`

View and update the CMS configuration.

| Action | Required Permission |
|--------|-------------------|
| View settings | `config:read` |
| Update settings | `config:update` |

## CSRF Protection

The admin panel uses double-submit cookie CSRF protection. Here is how it works:

1. On GET requests, the server sets a `csrf_token` cookie (scoped to `/admin/`, readable by JavaScript, SameSite=Strict).
2. The token value is also embedded in a `<meta>` tag in the page `<head>`.
3. On mutating requests (POST, PUT, PATCH, DELETE), the server validates that the token from the cookie matches the token submitted via either:
   - The `X-CSRF-Token` HTTP header, or
   - The `_csrf` form field.

HTMX requests include the CSRF token automatically via the `X-CSRF-Token` header, read from the `csrf_token` cookie by client-side JavaScript.

If the CSRF token is missing or does not match, the server responds with HTTP 403. The token is reused across GET requests within the same session to prevent desynchronization between the cookie and the `<meta>` tag during HTMX partial page navigations.

## Web Components

The admin panel uses Light DOM web components with the `mcms-` prefix for interactive UI elements. These are vanilla JavaScript custom elements -- no framework required.

| Component | Purpose |
|-----------|---------|
| `mcms-dialog` | Modal dialog for forms and confirmations |
| `mcms-data-table` | Sortable, filterable data table |
| `mcms-field-renderer` | Renders field inputs based on field type configuration |
| `mcms-media-picker` | Media browser and selection widget |
| `mcms-tree-nav` | Tree navigation for content hierarchy |
| `mcms-toast` | Toast notification system (triggered via `HX-Trigger` response headers) |
| `mcms-confirm` | Confirmation dialog for destructive actions |
| `mcms-search` | Search input with filtering |
| `mcms-file-input` | File upload input with preview |
| `mcms-scroll` | Scrollable container with overflow handling |
| `mcms-validation-wizard` | Multi-step form validation wizard |

These components are embedded in the binary and served from `/admin/static/`. Static assets are served with aggressive cache headers (`Cache-Control: public, max-age=31536000, immutable`).

## HTMX Interaction Model

The admin panel does not use client-side routing or a JavaScript framework. Page interactions follow this pattern:

1. Full page loads return a complete HTML document with the admin layout.
2. HTMX partial requests (identified by the `HX-Request` header) return only the HTML fragment that needs to be swapped.
3. After mutating operations (create, update, delete), the server triggers toast notifications via the `HX-Trigger` response header.
4. Navigation within the admin panel uses HTMX `hx-get` and `hx-push-url` to update the URL and swap the main content area.
5. If the session expires during an HTMX request, the server returns an `HX-Redirect` header pointing to the login page instead of a 302 redirect (which HTMX would follow transparently).

## Notes

- **No JavaScript framework.** The admin panel uses vanilla JavaScript, HTMX, and web components. There is no React, Vue, or other SPA framework.
- **Server-rendered templates.** All HTML is generated by templ, a type-safe Go template engine. Templates compile to Go code at build time.
- **Embedded static assets.** CSS, JavaScript, and web component files are embedded in the Go binary via `go:embed`. No external CDN or separate static file server is needed.
- **Permission enforcement.** Every admin panel route is wrapped with permission middleware. Viewing pages requires `resource:read`; mutating actions require the corresponding `resource:create`, `resource:update`, or `resource:delete` permission.
- **Session-based authentication.** The admin panel authenticates via the same cookie-based sessions as the REST API. API key authentication is not supported for admin panel access.
