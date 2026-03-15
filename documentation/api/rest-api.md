# REST API Reference

ModulaCMS exposes a JSON REST API at the base path `/api/v1`. Use it to manage content, media, users, roles, and configuration programmatically. All request and response bodies use `application/json` unless noted otherwise.

## Authentication

Every endpoint except public content delivery and the auth login/register routes requires authentication. The server checks two methods in order:

1. **Session cookie** -- Set automatically after a successful login or OAuth callback. The cookie name is configured in `modula.config.json`.
2. **API key** -- When no valid session cookie is present, include a Bearer token in the `Authorization` header. The token must exist in the `tokens` table, must not be revoked, and must not be expired.

```
Authorization: Bearer 01HXK4N2F8RJZGP6VTQY3MCSW9
```

## Common Patterns

| Pattern | Description |
|---------|-------------|
| IDs | All primary keys are 26-character ULID strings (e.g., `"01HXK4N2F8RJZGP6VTQY3MCSW9"`) |
| Collection endpoint | `GET /api/v1/{resource}` returns all items |
| Item endpoint | `/api/v1/{resource}/` operates on a single item |
| Item identification | `?q={ulid}` query parameter for GET, PUT, DELETE on item endpoints |
| Content-Type | `application/json` for all request and response bodies |
| Timestamps | RFC 3339 UTC (e.g., `"2026-01-30T12:00:00Z"`) |

### Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success (GET, PUT, DELETE) |
| 201 | Created (POST) |
| 204 | No Content (DELETE with no body) |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 405 | Method Not Allowed |
| 409 | Conflict (duplicate resource) |
| 500 | Internal Server Error |

## Auth Endpoints

Auth endpoints are rate limited to 10 requests per minute per IP. CORS is enabled on these routes.

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "your-password"}'
```

Response (200):

```json
{
  "user_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
  "email": "admin@example.com",
  "username": "admin",
  "created_at": "2026-01-30T12:00:00Z"
}
```

Sets an HTTP-only session cookie. Returns 401 for invalid credentials.

### Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Clears the session cookie. Returns 200 with `{"message": "Logged out successfully"}` regardless of auth state.

### Current User

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Response (200):

```json
{
  "user_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
  "email": "admin@example.com",
  "username": "admin",
  "name": "Admin User",
  "role": "admin"
}
```

Returns 401 with `{"error": "Not authenticated"}` when no valid session exists.

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "username": "newuser", "password": "secure-password"}'
```

Creates a new user. Request and response follow the same format as the Users POST endpoint. New users are assigned the `viewer` role by default.

### Password Reset

```bash
curl -X POST http://localhost:8080/api/v1/auth/reset \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "new-password"}'
```

Updates the user password. Request and response follow the Users PUT format.

### OAuth

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/auth/oauth/login` | Initiates OAuth flow with PKCE. Redirects to the configured OAuth provider. |
| GET | `/api/v1/auth/oauth/callback` | OAuth provider redirect target. Validates state, exchanges code for token via PKCE, creates or provisions the user, creates a session, sets the cookie, and redirects to the configured success URL. |

The callback receives `code` and `state` query parameters from the OAuth provider.

## Health

```bash
curl http://localhost:8080/api/v1/health
```

Returns a JSON health check response. No authentication required.

## Content Endpoints

### Content Data

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/contentdata` | List all content data |
| GET | `/api/v1/contentdata/?q={ulid}` | Get content data by ID |
| POST | `/api/v1/contentdata` | Create content data |
| PUT | `/api/v1/contentdata/` | Update content data |
| DELETE | `/api/v1/contentdata/?q={ulid}` | Delete content data |

### Content Fields

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/contentfields` | List all content fields |
| GET | `/api/v1/contentfields/?q={ulid}` | Get content field by ID |
| POST | `/api/v1/contentfields` | Create content field |
| PUT | `/api/v1/contentfields/` | Update content field |
| DELETE | `/api/v1/contentfields/?q={ulid}` | Delete content field |

### Admin Content Data

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admincontentdatas` | List all admin content data |
| GET | `/api/v1/admincontentdatas/?q={ulid}` | Get admin content data by ID |
| POST | `/api/v1/admincontentdatas` | Create admin content data |
| PUT | `/api/v1/admincontentdatas/` | Update admin content data |
| DELETE | `/api/v1/admincontentdatas/?q={ulid}` | Delete admin content data |

### Admin Content Fields

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admincontentfields` | List all admin content fields |
| GET | `/api/v1/admincontentfields/?q={ulid}` | Get admin content field by ID |
| POST | `/api/v1/admincontentfields` | Create admin content field |
| PUT | `/api/v1/admincontentfields/` | Update admin content field |
| DELETE | `/api/v1/admincontentfields/?q={ulid}` | Delete admin content field |

### Admin Content Tree

```bash
curl http://localhost:8080/api/v1/admin/tree/blog \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Returns the full content tree for an admin route identified by its slug. Supports the `?format=` query parameter to choose the output shape: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, or `raw`.

Returns 404 if no admin route matches the slug.

### Content Operations

These endpoints handle batch updates, tree operations, and node reordering:

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/content/batch` | Batch update content fields |
| POST | `/api/v1/content/tree` | Save content tree structure |
| POST | `/api/v1/contentdata/reorder` | Reorder content nodes |
| POST | `/api/v1/contentdata/move` | Move a content node to a new parent |
| POST | `/api/v1/admincontentdatas/reorder` | Reorder admin content nodes |
| POST | `/api/v1/admincontentdatas/move` | Move an admin content node |
| POST | `/api/v1/admin/content/heal` | Repair content tree inconsistencies |

## Schema Management

### Datatypes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/datatype` | List all datatypes |
| GET | `/api/v1/datatype/?q={ulid}` | Get datatype by ID |
| POST | `/api/v1/datatype` | Create datatype |
| PUT | `/api/v1/datatype/` | Update datatype |
| DELETE | `/api/v1/datatype/?q={ulid}` | Delete datatype |

### Fields

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/fields` | List all fields |
| GET | `/api/v1/fields/?q={ulid}` | Get field by ID |
| POST | `/api/v1/fields` | Create field |
| PUT | `/api/v1/fields/` | Update field |
| DELETE | `/api/v1/fields/?q={ulid}` | Delete field |

### Admin Datatypes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admindatatypes` | List all admin datatypes |
| GET | `/api/v1/admindatatypes/?q={ulid}` | Get admin datatype by ID |
| POST | `/api/v1/admindatatypes` | Create admin datatype |
| PUT | `/api/v1/admindatatypes/` | Update admin datatype |
| DELETE | `/api/v1/admindatatypes/?q={ulid}` | Delete admin datatype |

### Admin Fields

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/adminfields` | List all admin fields |
| GET | `/api/v1/adminfields/?q={ulid}` | Get admin field by ID |
| POST | `/api/v1/adminfields` | Create admin field |
| PUT | `/api/v1/adminfields/` | Update admin field |
| DELETE | `/api/v1/adminfields/?q={ulid}` | Delete admin field |

### Datatype Fields (Junction)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/datatypefields` | List all datatype-field associations |
| GET | `/api/v1/datatypefields/?q={ulid}` | Get association by ID |
| POST | `/api/v1/datatypefields` | Link a field to a datatype |
| PUT | `/api/v1/datatypefields/` | Update a datatype-field association |
| DELETE | `/api/v1/datatypefields/?q={ulid}` | Unlink a field from a datatype |

### Admin Datatype Fields (Junction)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admindatatypefields` | List all admin datatype-field associations |
| GET | `/api/v1/admindatatypefields/?q={ulid}` | Get association by ID |
| POST | `/api/v1/admindatatypefields` | Link an admin field to an admin datatype |
| PUT | `/api/v1/admindatatypefields/` | Update an admin datatype-field association |
| DELETE | `/api/v1/admindatatypefields/?q={ulid}` | Unlink an admin field from an admin datatype |

### Field Types

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/fieldtypes` | List all field types |
| GET | `/api/v1/fieldtypes/?q={ulid}` | Get field type by ID |
| POST | `/api/v1/fieldtypes` | Create field type |
| PUT | `/api/v1/fieldtypes/` | Update field type |
| DELETE | `/api/v1/fieldtypes/?q={ulid}` | Delete field type |

### Admin Field Types

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/adminfieldtypes` | List all admin field types |
| GET | `/api/v1/adminfieldtypes/?q={ulid}` | Get admin field type by ID |
| POST | `/api/v1/adminfieldtypes` | Create admin field type |
| PUT | `/api/v1/adminfieldtypes/` | Update admin field type |
| DELETE | `/api/v1/adminfieldtypes/?q={ulid}` | Delete admin field type |

## Routing

### Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/routes` | List all routes |
| GET | `/api/v1/routes/?q={ulid}` | Get route by ID |
| POST | `/api/v1/routes` | Create route |
| PUT | `/api/v1/routes/` | Update route |
| DELETE | `/api/v1/routes/?q={ulid}` | Delete route |

### Admin Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/adminroutes` | List all admin routes |
| GET | `/api/v1/adminroutes?ordered=true` | List admin routes sorted by Order field |
| GET | `/api/v1/adminroutes/?q={slug}` | Get admin route by slug |
| POST | `/api/v1/adminroutes` | Create admin route |
| PUT | `/api/v1/adminroutes/` | Update admin route |
| DELETE | `/api/v1/adminroutes/?q={ulid}` | Delete admin route |

The `ordered=true` variant reads each route's root content node "Order" field value and sorts routes numerically. Routes without an Order value appear last.

## Media

### Media Items

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/media` | List all media items |
| GET | `/api/v1/media/?q={ulid}` | Get media item by ID |
| POST | `/api/v1/media` | Create media metadata |
| PUT | `/api/v1/media/` | Update media metadata |
| DELETE | `/api/v1/media/?q={ulid}` | Delete media item |

### Media Upload

Upload a file by sending a multipart form POST to the same `/api/v1/media` endpoint:

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/image.jpg"
```

Content-Type: `multipart/form-data`. Form field name: `file`. Maximum upload size: 10 MB (configurable via `max_upload_size` in `modula.config.json`).

The upload pipeline validates that no file with the same name already exists, optimizes images at each configured dimension preset, uploads all variants to S3, and creates the media database record.

### Media Dimensions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/mediadimensions` | List all dimension presets |
| GET | `/api/v1/mediadimensions/?q={ulid}` | Get dimension preset by ID |
| POST | `/api/v1/mediadimensions` | Create dimension preset |
| PUT | `/api/v1/mediadimensions/` | Update dimension preset |
| DELETE | `/api/v1/mediadimensions/?q={ulid}` | Delete dimension preset |

## Users and Access Control

### Users

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/users` | List all users |
| GET | `/api/v1/users/?q={ulid}` | Get user by ID |
| POST | `/api/v1/users` | Create user |
| PUT | `/api/v1/users/` | Update user |
| DELETE | `/api/v1/users/?q={ulid}` | Delete user |

### Roles

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/roles` | List all roles |
| GET | `/api/v1/roles/?q={ulid}` | Get role by ID |
| POST | `/api/v1/roles` | Create role |
| PUT | `/api/v1/roles/` | Update role |
| DELETE | `/api/v1/roles/?q={ulid}` | Delete role |

System-protected roles (admin, editor, viewer) cannot be deleted or renamed.

### Permissions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/permissions` | List all permissions |
| GET | `/api/v1/permissions/?q={ulid}` | Get permission by ID |
| POST | `/api/v1/permissions` | Create permission |
| PUT | `/api/v1/permissions/` | Update permission |
| DELETE | `/api/v1/permissions/?q={ulid}` | Delete permission |

System-protected permissions cannot be deleted or renamed.

### Role Permissions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/role-permissions` | List all role-permission associations |
| GET | `/api/v1/role-permissions/?q={ulid}` | Get association by ID |
| POST | `/api/v1/role-permissions` | Assign a permission to a role |
| PUT | `/api/v1/role-permissions/` | Update a role-permission association |
| DELETE | `/api/v1/role-permissions/?q={ulid}` | Remove a permission from a role |
| GET | `/api/v1/role-permissions/role/?q={ulid}` | List all permissions for a specific role |

### Tokens

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/tokens` | List all tokens |
| GET | `/api/v1/tokens/?q={ulid}` | Get token by ID |
| POST | `/api/v1/tokens` | Create token |
| PUT | `/api/v1/tokens/` | Update token |
| DELETE | `/api/v1/tokens/?q={ulid}` | Delete token |

### User OAuth

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/usersoauth` | List all OAuth associations |
| GET | `/api/v1/usersoauth/?q={ulid}` | Get OAuth association by ID |
| POST | `/api/v1/usersoauth` | Create OAuth association |
| PUT | `/api/v1/usersoauth/` | Update OAuth association |
| DELETE | `/api/v1/usersoauth/?q={ulid}` | Delete OAuth association |

### Sessions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/sessions` | List all sessions |
| PUT | `/api/v1/sessions/` | Update session |
| DELETE | `/api/v1/sessions/?q={ulid}` | Delete session |

Use `/api/v1/auth/login` and `/api/v1/auth/logout` to create and destroy your own sessions.

### SSH Keys

Users can only manage their own SSH keys. All SSH key endpoints require authentication.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/ssh-keys` | List the authenticated user's SSH keys |
| POST | `/api/v1/ssh-keys` | Add an SSH key |
| DELETE | `/api/v1/ssh-keys/{id}` | Delete an SSH key by ID |

**Add an SSH key:**

```bash
curl -X POST http://localhost:8080/api/v1/ssh-keys \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"public_key": "ssh-ed25519 AAAAC3Nza...", "label": "my laptop"}'
```

Response (201): The full SSH key record.

**List SSH keys:**

```bash
curl http://localhost:8080/api/v1/ssh-keys \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Response (200):

```json
[
  {
    "ssh_key_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
    "key_type": "ssh-ed25519",
    "fingerprint": "SHA256:...",
    "label": "my laptop",
    "date_created": "2026-01-30T12:00:00Z",
    "last_used": "2026-02-15T08:30:00Z"
  }
]
```

The GET response omits the full public key. DELETE returns 204 No Content. Attempting to delete another user's key returns 403.

## Database Metadata

### Tables

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/tables` | List all table metadata |
| GET | `/api/v1/tables/?q={ulid}` | Get table metadata by ID |
| POST | `/api/v1/tables` | Create table metadata |
| PUT | `/api/v1/tables/` | Update table metadata |
| DELETE | `/api/v1/tables/?q={ulid}` | Delete table metadata |

## Media Administration

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/media/health` | Check S3 storage connectivity |
| DELETE | `/api/v1/media/cleanup` | Remove orphaned media files from S3 |

## Deploy

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/deploy/health` | Deployment health check |
| POST | `/api/v1/deploy/export` | Export site data |
| POST | `/api/v1/deploy/import` | Import site data |

## Configuration

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/config` | Get current configuration |
| PATCH | `/api/v1/admin/config` | Update configuration fields |
| GET | `/api/v1/admin/config/meta` | Get configuration field metadata (types, descriptions, defaults) |

## Metrics

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/metrics` | Returns a JSON snapshot of all collected metrics (requires `config:read` permission) |

## Import

Import endpoints parse CMS-specific JSON and create ModulaCMS content from it. All import endpoints accept POST only.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/import/contentful` | Import Contentful format |
| POST | `/api/v1/import/sanity` | Import Sanity.io format |
| POST | `/api/v1/import/strapi` | Import Strapi format |
| POST | `/api/v1/import/wordpress` | Import WordPress format |
| POST | `/api/v1/import/clean` | Import ModulaCMS native format |
| POST | `/api/v1/import?format={fmt}` | Bulk import with format parameter |

Valid formats: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`.

Response (201):

```json
{
  "success": true,
  "datatypes_created": 5,
  "fields_created": 20,
  "content_created": 100,
  "message": "Import completed successfully",
  "errors": []
}
```

## Content Delivery

### GET /{slug}

The public content delivery endpoint. Given a route slug, it builds the full content tree and returns it in the configured output format.

```bash
curl http://localhost:8080/blog
```

Override the output format with the `?format=` query parameter:

```bash
curl http://localhost:8080/blog?format=clean
```

Valid formats: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`.

Returns 404 if no route matches the slug.

## Notes

- All handlers use a singleton database connection pool initialized at startup. Handlers do not open or close individual connections.
- Auth endpoints have CORS middleware and rate limiting (10 requests per minute per IP).
- The router uses Go 1.22+ pattern routing via `net/http.ServeMux`.
- Every admin endpoint requires authentication and is gated by role-based permission checks. Public routes (auth, OAuth, content delivery by slug) have no permission guards.
