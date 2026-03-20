# REST API Reference

ModulaCMS exposes a JSON REST API at the base path `/api/v1` for managing content, media, users, roles, and configuration programmatically.

## Authentication

Every endpoint except public content delivery and the auth login/register routes requires authentication. The server checks two methods in this order:

1. **Session cookie** -- Set automatically after a successful login or OAuth callback. You configure the cookie name in `modula.config.json`.
2. **API key** -- When no valid session cookie is present, include a Bearer token in the `Authorization` header. The token must not be revoked or expired.

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

The server rate limits auth endpoints to 10 requests per minute per IP and enables CORS on these routes.

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

The server sets an HTTP-only session cookie. Returns 401 for invalid credentials.

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

Creates a new user with the same request/response format as the Users POST endpoint. The server assigns the `viewer` role by default.

### Password Reset

```bash
curl -X POST http://localhost:8080/api/v1/auth/reset \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "new-password"}'
```

Updates the user password. Uses the same request/response format as Users PUT.

### OAuth

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/auth/oauth/login` | Initiates OAuth flow with PKCE. Redirects to the configured OAuth provider. |
| GET | `/api/v1/auth/oauth/callback` | OAuth provider redirect target. Validates state, exchanges code for token via PKCE, creates or provisions the user, creates a session, sets the cookie, and redirects to the configured success URL. |

The OAuth provider sends `code` and `state` query parameters to the callback.

### Request Password Reset

```bash
curl -X POST http://localhost:8080/api/v1/auth/request-password-reset \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

Initiates a password reset email flow. Sends a reset token to the provided email address. Returns 200 regardless of whether the email exists (to prevent enumeration).

### Confirm Password Reset

```bash
curl -X POST http://localhost:8080/api/v1/auth/confirm-password-reset \
  -H "Content-Type: application/json" \
  -d '{"token": "reset-token-from-email", "password": "new-password"}'
```

Completes the password reset using a token received via email.

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
| GET | `/api/v1/contentdata/full` | List all content data with full details |
| GET | `/api/v1/contentdata/by-route` | List content data filtered by route |
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
| GET | `/api/v1/admincontentdatas/full` | List all admin content data with full details |
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

These endpoints handle composite creation, batch updates, tree operations, and node reordering:

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/api/v1/content/create` | `content:create` | Create content with fields (composite) |
| POST | `/api/v1/content/batch` | `content:update` | Batch update content fields |
| POST | `/api/v1/content/tree` | `content:update` | Save content tree structure |
| GET | `/api/v1/content/tree/{routeID}` | `content:read` | Get content tree by route ID |
| POST | `/api/v1/contentdata/reorder` | `content:update` | Reorder content nodes |
| POST | `/api/v1/contentdata/move` | `content:update` | Move a content node to a new parent |
| POST | `/api/v1/admincontentdatas/reorder` | `content:update` | Reorder admin content nodes |
| POST | `/api/v1/admincontentdatas/move` | `content:update` | Move an admin content node |
| POST | `/api/v1/admin/content/heal` | `content:update` | Repair content tree inconsistencies |

### Content Versions (Non-Admin)

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/contentversions` | `content:read` | List content versions (filtered by content_id) |
| GET | `/api/v1/content/versions` | `content:read` | List content versions |
| GET | `/api/v1/content/versions/` | `content:read` | Get specific version |
| POST | `/api/v1/content/versions` | `content:update` | Create a version snapshot |
| DELETE | `/api/v1/content/versions/` | `content:delete` | Delete a version |
| POST | `/api/v1/content/restore` | `content:update` | Restore content from a version |

### Publishing (Non-Admin)

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/api/v1/content/publish` | `content:publish` | Publish content |
| POST | `/api/v1/content/unpublish` | `content:publish` | Unpublish content |
| POST | `/api/v1/content/schedule` | `content:publish` | Schedule content for future publication |

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
| GET | `/api/v1/routes/full` | List all routes with full details |
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
| GET | `/api/v1/media` | List all media items (responses include computed `download_url` field) |
| GET | `/api/v1/media/?q={ulid}` | Get media item by ID |
| GET | `/api/v1/media/full` | List all media items with author names |
| GET | `/api/v1/media/{id}/download` | Download file (302 redirect to pre-signed S3 URL with Content-Disposition: attachment) |
| GET | `/api/v1/media/references?q={ulid}` | Scan for content fields referencing a media asset |
| GET | `/api/v1/media/health` | Check for orphaned files in S3 bucket (requires `media:admin`) |
| POST | `/api/v1/media` | Create media metadata |
| PUT | `/api/v1/media/` | Update media metadata |
| DELETE | `/api/v1/media/?q={ulid}` | Delete media item |
| DELETE | `/api/v1/media/cleanup` | Delete orphaned files from S3 (requires `media:admin`) |

### Media Upload

Upload a file by sending a multipart form POST to the same `/api/v1/media` endpoint:

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/image.jpg"
```

Content-Type: `multipart/form-data`. Form field name: `file`. Maximum upload size: 10 MB (configurable via `max_upload_size` in `modula.config.json`).

The server validates that no file with the same name already exists, optimizes images at each configured dimension preset, uploads all variants to S3, and creates the media record.

### Media Dimensions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/mediadimensions` | List all dimension presets |
| GET | `/api/v1/mediadimensions/?q={ulid}` | Get dimension preset by ID |
| POST | `/api/v1/mediadimensions` | Create dimension preset |
| PUT | `/api/v1/mediadimensions/` | Update dimension preset |
| DELETE | `/api/v1/mediadimensions/?q={ulid}` | Delete dimension preset |

### Media Folders

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/media-folders` | `media:read` | List root folders (or children via `?parent_id={ulid}`) |
| GET | `/api/v1/media-folders/tree` | `media:read` | Get full folder hierarchy as nested tree |
| POST | `/api/v1/media-folders` | `media:create` | Create folder |
| GET | `/api/v1/media-folders/{id}` | `media:read` | Get folder by ID |
| PUT | `/api/v1/media-folders/{id}` | `media:update` | Update folder |
| DELETE | `/api/v1/media-folders/{id}` | `media:delete` | Delete folder (rejects if non-empty) |
| GET | `/api/v1/media-folders/{id}/media` | `media:read` | List media in folder (supports pagination) |
| POST | `/api/v1/media/move` | `media:update` | Batch move media items to a folder or to root |

**Create a folder:**

```bash
curl -X POST http://localhost:8080/api/v1/media-folders \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"name": "Photos", "parent_id": ""}'
```

Response (201): The created media folder record.

**Get folder tree:**

```bash
curl http://localhost:8080/api/v1/media-folders/tree \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Returns a nested JSON array where each node contains `folder_id`, `name`, `parent_id`, `date_created`, `date_modified`, and a `children` array.

**Move media to a folder:**

```bash
curl -X POST http://localhost:8080/api/v1/media/move \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"media_ids": ["01HXK4N2F8...", "01HXK4N2F9..."], "folder_id": "01HXK4N2FA..."}'
```

Set `folder_id` to `null` or omit it to move media items back to root. Maximum batch size is 100 items.

> **Good to know**: The server enforces a maximum folder depth of 10 levels and unique names within each parent. Deleting a folder returns 409 Conflict if it contains child folders or media items.

## Users and Access Control

### Users

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/users` | List all users |
| GET | `/api/v1/users/?q={ulid}` | Get user by ID |
| GET | `/api/v1/users/full` | List all users with role details |
| GET | `/api/v1/users/full/` | Get user with role details by ID |
| POST | `/api/v1/users` | Create user |
| PUT | `/api/v1/users/` | Update user |
| DELETE | `/api/v1/users/?q={ulid}` | Delete user |
| POST | `/api/v1/users/reassign-delete` | Reassign content and delete user |
| GET | `/api/v1/users/sessions` | List sessions for authenticated user |

### Roles

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/roles` | List all roles |
| GET | `/api/v1/roles/?q={ulid}` | Get role by ID |
| POST | `/api/v1/roles` | Create role |
| PUT | `/api/v1/roles/` | Update role |
| DELETE | `/api/v1/roles/?q={ulid}` | Delete role |

> **Good to know**: System-protected roles (admin, editor, viewer) cannot be deleted or renamed.

### Permissions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/permissions` | List all permissions |
| GET | `/api/v1/permissions/?q={ulid}` | Get permission by ID |
| POST | `/api/v1/permissions` | Create permission |
| PUT | `/api/v1/permissions/` | Update permission |
| DELETE | `/api/v1/permissions/?q={ulid}` | Delete permission |

> **Good to know**: System-protected permissions cannot be deleted or renamed.

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

You can only manage your own SSH keys. All SSH key endpoints require authentication.

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

> **Good to know**: GET omits the full public key. DELETE returns 204 No Content. Attempting to delete another user's key returns 403.

## Database Metadata

### Tables

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/tables` | List all table metadata |
| GET | `/api/v1/tables/?q={ulid}` | Get table metadata by ID |
| POST | `/api/v1/tables` | Create table metadata |
| PUT | `/api/v1/tables/` | Update table metadata |
| DELETE | `/api/v1/tables/?q={ulid}` | Delete table metadata |

## Admin Media

Admin media items are stored separately from public media and power the admin panel UI (icons, backgrounds, branding, etc.). The admin media API mirrors the public media API but operates on the admin bucket.

### Admin Media Items

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/adminmedia` | List all admin media items |
| GET | `/api/v1/adminmedia/?q={ulid}` | Get admin media item by ID |
| GET | `/api/v1/adminmedia/{id}/download` | Download admin media file (302 redirect to pre-signed S3 URL) |
| POST | `/api/v1/adminmedia` | Upload or create admin media metadata |
| POST | `/api/v1/adminmedia/move` | Batch move admin media items to a folder or to root |
| PUT | `/api/v1/adminmedia/` | Update admin media metadata |
| DELETE | `/api/v1/adminmedia/?q={ulid}` | Delete admin media item |

Upload works identically to the public media endpoint: send a multipart form POST with the `file` field.

All admin media endpoints require `media:*` permissions.

### Admin Media Folders

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/adminmedia-folders` | `media:read` | List admin media folders (or children via `?parent_id={ulid}`) |
| GET | `/api/v1/adminmedia-folders/tree` | `media:read` | Get full admin folder hierarchy as nested tree |
| POST | `/api/v1/adminmedia-folders` | `media:create` | Create admin media folder |
| GET | `/api/v1/adminmedia-folders/{id}` | `media:read` | Get admin folder by ID |
| PUT | `/api/v1/adminmedia-folders/{id}` | `media:update` | Update admin folder |
| DELETE | `/api/v1/adminmedia-folders/{id}` | `media:delete` | Delete admin folder (rejects if non-empty) |
| GET | `/api/v1/adminmedia-folders/{id}/media` | `media:read` | List media in admin folder (supports pagination) |

The admin media folder structure follows the same rules as public media folders: maximum folder depth of 10 levels and unique names within each parent.

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
| GET | `/api/v1/admin/config/search-index` | Get search index configuration |

## Metrics

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/metrics` | Returns a JSON snapshot of all collected metrics (requires `config:read` permission) |

## Import

Import endpoints parse CMS-specific JSON and create ModulaCMS content from it. All import endpoints accept POST.

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

## Admin Publishing

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/api/v1/admin/content/publish` | `content:publish` | Publish admin content |
| POST | `/api/v1/admin/content/unpublish` | `content:publish` | Unpublish admin content |
| POST | `/api/v1/admin/content/schedule` | `content:publish` | Schedule admin content for future publication |

## Admin Content Versions

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/admin/content/versions` | `content:read` | List admin content versions |
| GET | `/api/v1/admin/content/versions/` | `content:read` | Get specific admin content version |
| POST | `/api/v1/admin/content/versions` | `content:update` | Create an admin content version snapshot |
| DELETE | `/api/v1/admin/content/versions/` | `content:delete` | Delete an admin content version |
| POST | `/api/v1/admin/content/restore` | `content:update` | Restore admin content from a version |

## Locales

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/admin/locales` | `locale:read` | List all locales |
| GET | `/api/v1/admin/locales/` | `locale:read` | Get locale by ID |
| POST | `/api/v1/admin/locales` | `locale:create` | Create locale |
| PUT | `/api/v1/admin/locales/` | `locale:update` | Update locale |
| DELETE | `/api/v1/admin/locales/` | `locale:delete` | Delete locale |

## Webhooks

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/admin/webhooks` | `webhook:read` | List all webhooks |
| POST | `/api/v1/admin/webhooks` | `webhook:create` | Create webhook |
| GET | `/api/v1/admin/webhooks/{id}` | `webhook:read` | Get webhook by ID |
| PUT | `/api/v1/admin/webhooks/{id}` | `webhook:update` | Update webhook |
| DELETE | `/api/v1/admin/webhooks/{id}` | `webhook:delete` | Delete webhook |
| POST | `/api/v1/admin/webhooks/{id}/test` | `webhook:update` | Send test delivery |
| GET | `/api/v1/admin/webhooks/{id}/deliveries` | `webhook:read` | List deliveries for webhook |
| POST | `/api/v1/admin/webhooks/deliveries/{id}/retry` | `webhook:update` | Retry a failed delivery |

## Translations

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/api/v1/admin/contentdata/{id}/translations` | `content:create` | Create translation for content |
| POST | `/api/v1/admin/admincontentdata/{id}/translations` | `content:create` | Create translation for admin content |

## Validations

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/validations` | `validations:read` | List all validations |
| POST | `/api/v1/validations` | `validations:create` | Create validation |
| GET | `/api/v1/validations/search` | `validations:read` | Search validations |
| GET | `/api/v1/validations/{id}` | `validations:read` | Get validation by ID |
| PUT | `/api/v1/validations/{id}` | `validations:update` | Update validation |
| DELETE | `/api/v1/validations/{id}` | `validations:delete` | Delete validation |

## Admin Validations

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/admin/validations` | `admin_validations:read` | List all admin validations |
| POST | `/api/v1/admin/validations` | `admin_validations:create` | Create admin validation |
| GET | `/api/v1/admin/validations/search` | `admin_validations:read` | Search admin validations |
| GET | `/api/v1/admin/validations/{id}` | `admin_validations:read` | Get admin validation by ID |
| PUT | `/api/v1/admin/validations/{id}` | `admin_validations:update` | Update admin validation |
| DELETE | `/api/v1/admin/validations/{id}` | `admin_validations:delete` | Delete admin validation |

## Activity

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/activity/recent` | `audit:read` | Get recent activity feed |

## Public Locales

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/locales` | List enabled locales (no authentication required) |

## Content Delivery via Slug

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/content/{slug}` | Get content tree by route slug (no authentication required) |

Supports `?format=` query parameter: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`.

## Search

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/search` | Full-text search across published content (no auth required) |
| POST | `/api/v1/admin/search/rebuild` | Re-index all documents (requires `search:update` permission) |

**Search query parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `q` | Yes | Search query string |
| `type` | No | Filter by content type (e.g. `"blog_post"`) |
| `locale` | No | Filter by locale code |
| `limit` | No | Maximum results (default 20) |
| `offset` | No | Pagination offset |
| `prefix` | No | Enable prefix matching (default true, set `"false"` for exact) |

## Globals

```bash
curl http://localhost:8080/api/v1/globals
```

Returns all published global content trees. No authentication required. Global content items are root nodes typed as `_global` whose trees are available site-wide (e.g. navigation, footer, site settings).

## Query

```bash
curl "http://localhost:8080/api/v1/query/blog_post?sort=-published_at&limit=10"
```

Query content items by datatype name with optional filtering, sorting, and pagination. No authentication required.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `sort` | No | Sort field, prefix `-` for descending (e.g. `-published_at`) |
| `limit` | No | Maximum results (default 20, max 100) |
| `offset` | No | Pagination offset |
| `locale` | No | Locale code filter |
| `status` | No | Content status filter (default `published`) |
| `{field}` | No | Field filters as key-value pairs (supports `[eq]`, `[ne]`, `[gt]`, `[gte]`, `[lt]`, `[lte]`, `[like]`, `[in]` operators) |

## Public Content Delivery

### GET /{slug}

The top-level public content delivery endpoint. Given a route slug, the server builds the full content tree and returns it in the configured output format.

```bash
curl http://localhost:8080/blog
```

Override the output format with the `?format=` query parameter:

```bash
curl http://localhost:8080/blog?format=clean
```

Valid formats: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`.

Returns 404 if no route matches the slug.

> **Good to know**: Every admin endpoint requires authentication and role-based permission checks. Public routes (auth, OAuth, content delivery by slug) have no permission guards.
