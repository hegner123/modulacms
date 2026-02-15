# ModulaCMS REST API Contract

**Version:** v1
**Base URL:** `/api/v1`
**Last Updated:** 2026-01-30

---

## Authentication

All non-public endpoints require authentication via one of two methods:

### Cookie Authentication (Primary)

Session cookie set after login or OAuth callback. The cookie name is configured in `config.json`.

### API Key Authentication (Fallback)

When no valid session cookie is present, the server checks for a Bearer token:

```
Authorization: Bearer <api_key>
```

The token must exist in the `tokens` table with `token_type = "api_key"`, not be revoked, and not be expired.

---

## Common Patterns

| Pattern | Description |
|---------|-------------|
| IDs | All primary keys are ULID strings (e.g., `"01HXK4N2F8..."`) |
| Collection endpoint | `GET /api/v1/{resource}` — list all |
| Item endpoint | `/api/v1/{resource}/` — operate on single item |
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

---

## Auth Endpoints

Rate limited: 10 requests/minute per IP. CORS enabled.

### POST /api/v1/auth/login

```json
// Request
{ "email": "string", "password": "string" }

// Response 200
{ "user_id": "ulid", "email": "string", "username": "string", "created_at": "timestamp" }
```

Sets an HTTP-only session cookie. Returns 401 for invalid credentials.

### POST /api/v1/auth/logout

Clears the session cookie. Returns 200 with `{"message": "Logged out successfully"}` regardless of auth state.

### GET /api/v1/auth/me

```json
// Response 200
{ "user_id": "ulid", "email": "string", "username": "string", "name": "string", "role": "string" }

// Response 401
{ "error": "Not authenticated" }
```

### POST /api/v1/auth/register

Creates a new user. Request/response follows the Users POST format.

### POST /api/v1/auth/reset

Updates user password. Request/response follows the Users PUT format.

### GET /api/v1/auth/oauth/login

Initiates OAuth flow with PKCE. Redirects to configured OAuth provider.

### GET /api/v1/auth/oauth/callback

OAuth provider redirect target. Validates state, exchanges code for token via PKCE, provisions user, creates session, sets cookie, redirects to configured success URL.

Query parameters: `code`, `state` (set by provider).

---

## Admin Content

### Admin Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/adminroutes` | List all admin routes |
| GET | `/api/v1/adminroutes?ordered=true` | List admin routes sorted by Order field |
| GET | `/api/v1/adminroutes/?q={slug}` | Get admin route by slug |
| POST | `/api/v1/adminroutes` | Create admin route |
| PUT | `/api/v1/adminroutes/` | Update admin route |
| DELETE | `/api/v1/adminroutes/?q={ulid}` | Delete admin route |

The `ordered=true` variant finds each route's root content node, reads its "Order" field value, and sorts numerically. Routes without an Order value appear last.

### Admin Tree

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/tree/{slug}` | Get admin content tree by route slug |

Returns the full admin content tree for a route, built from admin content data, datatypes, content fields, and field definitions. Supports `?format=` query parameter (contentful, sanity, strapi, wordpress, clean, raw).

Returns 404 if the slug does not match an admin route.

### Admin Content Data

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admincontentdatas` | List all |
| GET | `/api/v1/admincontentdatas/?q={ulid}` | Get by ID |
| POST | `/api/v1/admincontentdatas` | Create |
| PUT | `/api/v1/admincontentdatas/` | Update |
| DELETE | `/api/v1/admincontentdatas/?q={ulid}` | Delete |

### Admin Content Fields

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admincontentfields` | List all |
| GET | `/api/v1/admincontentfields/?q={ulid}` | Get by ID |
| POST | `/api/v1/admincontentfields` | Create |
| PUT | `/api/v1/admincontentfields/` | Update |
| DELETE | `/api/v1/admincontentfields/?q={ulid}` | Delete |

### Admin Datatypes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admindatatypes` | List all |
| GET | `/api/v1/admindatatypes/?q={ulid}` | Get by ID |
| POST | `/api/v1/admindatatypes` | Create |
| PUT | `/api/v1/admindatatypes/` | Update |
| DELETE | `/api/v1/admindatatypes/?q={ulid}` | Delete |

### Admin Fields

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/adminfields` | List all |
| GET | `/api/v1/adminfields/?q={ulid}` | Get by ID |
| POST | `/api/v1/adminfields` | Create |
| PUT | `/api/v1/adminfields/` | Update |
| DELETE | `/api/v1/adminfields/?q={ulid}` | Delete |

---

## Public Content

### Content Data

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/contentdata` | List all |
| GET | `/api/v1/contentdata/?q={ulid}` | Get by ID |
| POST | `/api/v1/contentdata` | Create |
| PUT | `/api/v1/contentdata/` | Update |
| DELETE | `/api/v1/contentdata/?q={ulid}` | Delete |

### Content Fields

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/contentfields` | List all |
| GET | `/api/v1/contentfields/?q={ulid}` | Get by ID |
| POST | `/api/v1/contentfields` | Create |
| PUT | `/api/v1/contentfields/` | Update |
| DELETE | `/api/v1/contentfields/?q={ulid}` | Delete |

---

## Schema Management

### Datatypes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/datatype` | List all |
| GET | `/api/v1/datatype/?q={ulid}` | Get by ID |
| POST | `/api/v1/datatype` | Create |
| PUT | `/api/v1/datatype/` | Update |
| DELETE | `/api/v1/datatype/?q={ulid}` | Delete |

### Fields

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/fields` | List all |
| GET | `/api/v1/fields/?q={ulid}` | Get by ID |
| POST | `/api/v1/fields` | Create |
| PUT | `/api/v1/fields/` | Update |
| DELETE | `/api/v1/fields/?q={ulid}` | Delete |

---

## Routing

### Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/routes` | List all |
| GET | `/api/v1/routes/?q={ulid}` | Get by ID |
| POST | `/api/v1/routes` | Create |
| PUT | `/api/v1/routes/` | Update |
| DELETE | `/api/v1/routes/?q={ulid}` | Delete |

---

## Media

### Media Items

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/media` | List all |
| GET | `/api/v1/media/?q={ulid}` | Get by ID |
| POST | `/api/v1/media` | Create metadata |
| PUT | `/api/v1/media/` | Update metadata |
| DELETE | `/api/v1/media/?q={ulid}` | Delete |

### Media Upload

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/mediaupload/` | Upload file to S3 storage |

Content-Type: `multipart/form-data`. Form field: `file` (max 10 MB).

Process: validate no duplicate filename, optimize images, upload to S3, create media record.

### Media Dimensions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/mediadimensions` | List all |
| GET | `/api/v1/mediadimensions/?q={ulid}` | Get by ID |
| POST | `/api/v1/mediadimensions` | Create |
| PUT | `/api/v1/mediadimensions/` | Update |
| DELETE | `/api/v1/mediadimensions/?q={ulid}` | Delete |

---

## Users and Access Control

### Users

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/users` | List all |
| GET | `/api/v1/users/?q={ulid}` | Get by ID |
| POST | `/api/v1/users` | Create |
| PUT | `/api/v1/users/` | Update |
| DELETE | `/api/v1/users/?q={ulid}` | Delete |

### Roles

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/roles` | List all |
| GET | `/api/v1/roles/?q={ulid}` | Get by ID |
| POST | `/api/v1/roles` | Create |
| PUT | `/api/v1/roles/` | Update |
| DELETE | `/api/v1/roles/?q={ulid}` | Delete |

### Tokens

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/tokens` | List all |
| GET | `/api/v1/tokens/?q={ulid}` | Get by ID |
| POST | `/api/v1/tokens` | Create |
| PUT | `/api/v1/tokens/` | Update |
| DELETE | `/api/v1/tokens/?q={ulid}` | Delete |

### User OAuth

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/usersoauth` | List all |
| GET | `/api/v1/usersoauth/?q={ulid}` | Get by ID |
| POST | `/api/v1/usersoauth` | Create |
| PUT | `/api/v1/usersoauth/` | Update |
| DELETE | `/api/v1/usersoauth/?q={ulid}` | Delete |

### Sessions

| Method | Path | Description |
|--------|------|-------------|
| PUT | `/api/v1/sessions/` | Update session |
| DELETE | `/api/v1/sessions/?q={ulid}` | Delete session |

GET and POST are not allowed on session endpoints. Use `/api/v1/auth/login` and `/api/v1/auth/logout` instead.

### SSH Keys

Require authentication. Users can only manage their own keys.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/ssh-keys` | List authenticated user's SSH keys |
| POST | `/api/v1/ssh-keys` | Add SSH key |
| DELETE | `/api/v1/ssh-keys/{id}` | Delete SSH key by ID |

**POST request:**
```json
{ "public_key": "ssh-ed25519 AAAA...", "label": "my laptop" }
```

**POST response (201):**
Full SSH key record.

**GET response (200):**
```json
[
  {
    "ssh_key_id": "ulid",
    "key_type": "ssh-ed25519",
    "fingerprint": "SHA256:...",
    "label": "my laptop",
    "date_created": "timestamp",
    "last_used": "timestamp"
  }
]
```

GET response omits the full public key. DELETE returns 204 No Content. Returns 403 if the key belongs to a different user.

---

## Database Metadata

### Tables

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/tables` | List all |
| GET | `/api/v1/tables/?q={ulid}` | Get by ID |
| POST | `/api/v1/tables` | Create |
| PUT | `/api/v1/tables/` | Update |
| DELETE | `/api/v1/tables/?q={ulid}` | Delete |

---

## Import

Import endpoints parse CMS-specific JSON and create ModulaCMS content. All accept POST only.

**Note:** Database insertion is currently stubbed. Parsing works; records are not persisted.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/import/contentful` | Import Contentful format |
| POST | `/api/v1/import/sanity` | Import Sanity.io format |
| POST | `/api/v1/import/strapi` | Import Strapi format |
| POST | `/api/v1/import/wordpress` | Import WordPress format |
| POST | `/api/v1/import/clean` | Import ModulaCMS native format |
| POST | `/api/v1/import?format={fmt}` | Bulk import with format param |

Valid formats: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`.

**Response (201):**
```json
{
  "success": true,
  "datatypes_created": 5,
  "fields_created": 20,
  "content_created": 100,
  "message": "string",
  "errors": []
}
```

---

## Content Delivery

### GET /{slug}

Catch-all handler. Looks up a route by slug, builds the content tree, and returns it in the configured output format.

Query parameter: `?format=` overrides the default. Valid: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`.

---

## Implementation Notes

- All handlers use `db.ConfigDB(c)` which returns a singleton connection pool (initialized at startup via `db.InitDB()`). Handlers do not open or close connections.
- Request bodies are decoded with `json.NewDecoder(r.Body)`. Response bodies are encoded with `json.NewEncoder(w)`.
- Auth endpoints have CORS middleware and rate limiting (10 req/min). Resource endpoints use `HandleFunc` directly.
- The router uses `net/http.ServeMux` (standard library). See `internal/router/mux.go`.
