# Build a Custom Admin Interface

ModulaCMS exposes a complete REST API for every administrative operation, so you can build a custom admin interface in any framework or language.

## Authenticate your interface

Your interface must authenticate before calling admin endpoints. ModulaCMS supports two authentication methods.

### Cookie-based sessions

Authenticate with email and password. The server returns an HTTP-only session cookie valid for 24 hours.

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "your-password"}'
```

Include the session cookie in subsequent requests. See [Authentication and access control](authentication.md) for session management details.

### API keys

For server-to-server or automated integrations, use an API key as a Bearer token:

```bash
curl http://localhost:8080/api/v1/media \
  -H "Authorization: Bearer mcms_01JMKX5V6QNPZ3R8W4T2YH9B0D"
```

API keys inherit the permissions of the user they belong to.

### Check the current user

Verify the authenticated session and retrieve the user's role:

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

```json
{
  "user_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "email": "admin@example.com",
  "username": "admin",
  "name": "Admin User",
  "role": "admin"
}
```

Use the role to determine what the user can see and do. If the role is `admin`, the user has full access. Otherwise, query the role's permissions to enable or disable UI elements.

## Common API patterns

All admin endpoints live under `/api/v1`. Request and response bodies use JSON. IDs are 26-character ULID strings.

### Standard CRUD

Most resources follow the same endpoint structure:

| Operation | Method | Path | Example |
|-----------|--------|------|---------|
| List all | GET | `/api/v1/{resource}` | `/api/v1/datatype` |
| Get one | GET | `/api/v1/{resource}/?q={id}` | `/api/v1/datatype/?q=01HXK...` |
| Create | POST | `/api/v1/{resource}` | `/api/v1/datatype` |
| Update | PUT | `/api/v1/{resource}/?q={id}` | `/api/v1/datatype/?q=01HXK...` |
| Delete | DELETE | `/api/v1/{resource}/?q={id}` | `/api/v1/datatype/?q=01HXK...` |

The `?q=` query parameter identifies the item for single-resource operations.

### Paginate results

List endpoints support pagination via `limit` and `offset` query parameters:

```bash
curl "http://localhost:8080/api/v1/contentdata?limit=25&offset=50" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

- Default limit: 50 items
- Maximum limit: 1000 items
- Without pagination parameters, the endpoint returns all items

### Handle errors

All errors return a JSON body with an `error` field:

```json
{"error": "not found"}
```

| Status | Meaning |
|--------|---------|
| 400 | Invalid input or missing required fields |
| 401 | No valid session or API key |
| 403 | Authenticated but missing required permission |
| 404 | Resource not found |
| 409 | Duplicate resource (e.g., same filename for media) |
| 500 | Internal server error |

## Set up the schema

Before creating content, define datatypes (content schemas) and fields (data definitions).

### List available field types

Field types define the kind of data a field holds (text, number, image, etc.). ModulaCMS ships with built-in field types.

```bash
curl http://localhost:8080/api/v1/fieldtypes \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Create a field

Fields define the individual pieces of data within a datatype. Each field references a field type.

```bash
curl -X POST http://localhost:8080/api/v1/fields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "label": "Title",
    "field_type_id": "01HXK4N2F8...",
    "required": true
  }'
```

### Create a datatype

Datatypes define the structure of your content.

```bash
curl -X POST http://localhost:8080/api/v1/datatype \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "label": "Blog Post",
    "type": "page"
  }'
```

### Link fields to datatypes

Fields are associated with a datatype via the `parent_id` field when creating the field:

```bash
curl -X POST http://localhost:8080/api/v1/fields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "label": "Title",
    "type": "text",
    "parent_id": "01HXK4N2F8..."
  }'
```

### Retrieve full datatypes

Fetch a datatype with all its linked fields in a single response:

```bash
curl http://localhost:8080/api/v1/datatype/full \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## Manage content

Content entries are instances of a datatype. Each entry holds field values defined by its datatype's schema.

### Create content

```bash
curl -X POST http://localhost:8080/api/v1/contentdata \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "datatype_id": "01HXK4N2F8...",
    "status": "draft",
    "author_id": "01JMKW8N3Q...",
    "parent_id": "",
    "first_child_id": "",
    "next_sibling_id": "",
    "prev_sibling_id": ""
  }'
```

### Set content field values

After creating a content entry, set its field values:

```bash
curl -X POST http://localhost:8080/api/v1/contentfields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01HXK4N2FA...",
    "field_id": "01HXK4N2F9...",
    "value": "My First Blog Post"
  }'
```

### Batch update fields

Update multiple content fields in a single request:

```bash
curl -X POST http://localhost:8080/api/v1/content/batch \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "updates": [
      {"content_field_id": "01HXK...", "value": "Updated Title"},
      {"content_field_id": "01HXK...", "value": "Updated Body"}
    ]
  }'
```

## Work with content trees

Content is organized in a tree structure with parent-child relationships.

### Get the tree

Retrieve the assembled content tree:

```bash
curl http://localhost:8080/api/v1/content/tree/01HXK4N2F8... \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Save the tree

The tree save endpoint accepts batch creates, updates, and deletes in a single request:

```bash
curl -X POST http://localhost:8080/api/v1/content/tree \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_id": "01HXK4N2F8...",
    "creates": [
      {
        "client_id": "temp-uuid-1",
        "datatype_id": "01HXK4N2F8...",
        "parent_id": "01HXK4N2F9..."
      }
    ],
    "updates": [
      {
        "content_data_id": "01HXK4N2FA...",
        "parent_id": "01HXK4N2F8..."
      }
    ],
    "deletes": ["01HXK4N2FB..."]
  }'
```

The response includes an `id_map` that maps your temporary client IDs to server-generated ULIDs. Use this to update your local state after the save.

### Reorder and move nodes

Reorder siblings within the same parent:

```bash
curl -X POST http://localhost:8080/api/v1/contentdata/reorder \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01HXK...",
    "prev_sibling_id": "01HXK...",
    "next_sibling_id": "01HXK..."
  }'
```

Move a node to a different parent:

```bash
curl -X POST http://localhost:8080/api/v1/contentdata/move \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01HXK...",
    "parent_id": "01HXK..."
  }'
```

## Use the admin content API

ModulaCMS has a second, independent content system for admin panel configuration. Admin content endpoints mirror the public content endpoints with `admin` prefixes:

| Public endpoint | Admin equivalent |
|-----------------|-----------------|
| `/api/v1/contentdata` | `/api/v1/admincontentdatas` |
| `/api/v1/contentfields` | `/api/v1/admincontentfields` |
| `/api/v1/datatype` | `/api/v1/admindatatypes` |
| `/api/v1/fields` | `/api/v1/adminfields` |
| `/api/v1/fieldtypes` | `/api/v1/adminfieldtypes` |
| `/api/v1/routes` | `/api/v1/adminroutes` |

Every operation shown in this guide works identically on admin endpoints. To create an admin screen, create an admin datatype via `/api/v1/admindatatypes`, attach admin fields via `/api/v1/adminfields`, and populate admin content via `/api/v1/admincontentdatas`.

> **Good to know**: Admin content and public content are independent systems. Changing admin content does not affect your site's public content, and vice versa.

## Create routes

Routes map URL slugs to content trees. Create routes to make content accessible:

```bash
curl -X POST http://localhost:8080/api/v1/routes \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "blog",
    "type": "page",
    "content_data_id": "01HXK4N2F8..."
  }'
```

List all routes with their associated content:

```bash
curl http://localhost:8080/api/v1/routes/full \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## Manage media

### Upload files

Media upload uses `multipart/form-data`:

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/image.jpg"
```

### Organize media into folders

```bash
# Create a folder
curl -X POST http://localhost:8080/api/v1/media-folders \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"name": "Photos", "parent_id": ""}'

# Get the full folder tree
curl http://localhost:8080/api/v1/media-folders/tree \
  -H "Cookie: session=YOUR_SESSION_COOKIE"

# Move media to a folder
curl -X POST http://localhost:8080/api/v1/media/move \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"media_ids": ["01HXK...", "01HXK..."], "folder_id": "01HXK..."}'
```

## Publish content

Content starts in `draft` status and must be published to appear on the public site.

```bash
# Publish content
curl -X POST http://localhost:8080/api/v1/content/publish \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"content_data_id": "01HXK..."}'

# Unpublish content
curl -X POST http://localhost:8080/api/v1/content/unpublish \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"content_data_id": "01HXK..."}'
```

## Check permissions

Every admin endpoint requires a `resource:operation` permission. Query the user's role permissions to determine which actions to display:

```bash
curl "http://localhost:8080/api/v1/role-permissions/role/?q=ROLE_ID" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

If the user's role lacks a required permission, the API returns 403. Your interface should hide or disable actions the user cannot perform.

> **Good to know**: The API supports CORS on auth and content delivery endpoints. For a custom admin interface on a different domain, add your domain to the CORS allowed origins in `modula.config.json`.

> **Good to know**: Every create, update, and delete operation produces a change event in the audit trail with the authenticated user's ID, request ID, and IP address.
