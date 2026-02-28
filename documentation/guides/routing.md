# Routing

Routes connect a slug to a content tree. Each route maps a unique URL identifier to a single content data hierarchy, making content addressable and deliverable over HTTP.

## Concepts

**Route** -- A record that binds a slug (a unique string identifier) to a content tree. Each route has a slug, a title, a status, and an author. Content belonging to a route is accessed by resolving the slug to a `route_id`, then loading the content tree for that route.

**Slug** -- The unique identifier for a route. Slugs are freeform text -- they can be a domain name (`example.com`), a path (`/about`), a keyword (`homepage`), or any string that makes sense for your use case. The only constraint is uniqueness across all routes.

**Content isolation** -- Every content query in the system is scoped by `route_id`. Content in one route never appears in queries for another route. This isolation is enforced at the database level.

**Admin routes** -- A separate, parallel route system for internal admin content. Admin routes live in their own table and are distinct from content-facing routes. Admin content cannot mix with route content.

## How Routes Work

A route is a lightweight connector: it associates a slug with the content tree that lives under it.

```
Route: slug = "homepage"
  └── content_data (_root "Page")     <- the content tree for this route
      ├── content_field: title = "Welcome"
      ├── content_data ("Hero Section")
      │   └── content_field: heading = "Hello"
      └── content_data ("Cards")
          ├── content_data ("Card")
          └── content_data ("Card")
```

When a client requests content for a slug, the system:

1. Resolves the slug to a `route_id`.
2. Loads all `content_data` records for that route.
3. Loads all `content_fields`, `datatypes`, and `fields` for that route.
4. Assembles the content tree with the `_root`-typed node as root.
5. Resolves any `_reference` nodes by composing referenced content trees (see [content-modeling.md](content-modeling.md#tree-composition)).
6. Returns the assembled tree as JSON.

## Creating a Route

Create a route with `POST /api/v1/routes`:

```bash
curl -X POST http://localhost:8080/api/v1/routes \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "homepage",
    "title": "Home Page",
    "status": 1
  }'
```

Response (HTTP 201):

```json
{
  "route_id": "01JNRW9P2DKTZ6Q4M8W3B5J7CL",
  "slug": "homepage",
  "title": "Home Page",
  "status": 1,
  "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "date_created": "2026-02-27T10:00:00Z",
  "date_modified": "2026-02-27T10:00:00Z"
}
```

The slug must be unique across all routes.

### Route Status

Routes use an integer status field:

| Value | Meaning |
|-------|---------|
| 0 | Inactive |
| 1 | Active |

### Setting Up a Route's Content Tree

After creating a route, create a root content data node for it. The root node's datatype must have `type = "_root"`:

```bash
curl -X POST http://localhost:8080/api/v1/contentdata \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "route_id": "01JNRW9P2DKTZ6Q4M8W3B5J7CL",
    "datatype_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
    "status": "draft"
  }'
```

This gives the route an empty content tree ready for child nodes. See [content-trees.md](content-trees.md) for details on building out the tree.

## Managing Routes

### Listing Routes

```bash
curl http://localhost:8080/api/v1/routes \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

With pagination:

```bash
curl "http://localhost:8080/api/v1/routes?limit=20&offset=0" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Paginated responses include total count:

```json
{
  "data": [
    {"route_id": "...", "slug": "homepage", "title": "Home Page", "status": 1}
  ],
  "total": 4,
  "limit": 20,
  "offset": 0
}
```

### Getting a Single Route

```bash
curl "http://localhost:8080/api/v1/routes/?q=01JNRW9P2DKTZ6Q4M8W3B5J7CL" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

The `q` parameter is the route ID (26-character ULID).

### Updating a Route

```bash
curl -X PUT http://localhost:8080/api/v1/routes/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "route_id": "01JNRW9P2DKTZ6Q4M8W3B5J7CL",
    "slug": "homepage",
    "title": "Updated Home Page Title",
    "status": 1
  }'
```

### Deleting a Route

```bash
curl -X DELETE "http://localhost:8080/api/v1/routes/?q=01JNRW9P2DKTZ6Q4M8W3B5J7CL" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Deleting a route cascades: all content data and content field records belonging to that route are permanently deleted. There is no undo. Back up route content before deletion.

## Delivering Content by Slug

The public content delivery endpoint resolves a slug and returns the assembled content tree:

```bash
curl http://localhost:8080/api/v1/content/homepage
```

This endpoint is public (no authentication required).

### Output Formats

The response structure can be configured to match other CMS conventions. Set the default in `config.json`:

```json
{
  "output_format": "clean"
}
```

Or override per-request with the `format` query parameter:

```bash
curl "http://localhost:8080/api/v1/content/homepage?format=contentful"
```

| Format | Description |
|--------|-------------|
| `raw` | Native ModulaCMS tree structure (default) |
| `clean` | Simplified structure with flat field values |
| `contentful` | Mimics Contentful's response format |
| `sanity` | Mimics Sanity's response format |
| `strapi` | Mimics Strapi's response format |
| `wordpress` | Mimics WordPress REST API format |

## Configuration

Route-related configuration in `config.json`:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `output_format` | string | `""` (raw) | Default output format for content delivery |
| `client_site` | string | -- | Client site URL used in output format transformations |
| `space_id` | string | -- | Space identifier used in Contentful-style output format |
| `composition_max_depth` | integer | 10 | Maximum depth for composing referenced content subtrees |

## API Reference

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/routes` | `routes:read` | List all routes (supports `limit` and `offset`) |
| POST | `/api/v1/routes` | `routes:create` | Create a route |
| GET | `/api/v1/routes/` | `routes:read` | Get a single route (`?q=ROUTE_ID`) |
| PUT | `/api/v1/routes/` | `routes:update` | Update a route |
| DELETE | `/api/v1/routes/` | `routes:delete` | Delete a route and all its content (cascade) |
| GET | `/api/v1/content/{slug}` | Public | Deliver assembled content tree for a slug |

Admin routes follow the same pattern under `/api/v1/adminroutes`.

## Notes

- **Cascade deletion.** Deleting a route deletes all content data and content field records for that route. The `ON DELETE CASCADE` foreign key enforces this at the database level.
- **Slug uniqueness.** Route slugs are unique across the entire installation. Attempting to create a duplicate slug returns an error.
- **Content fields denormalize route_id.** Content field records carry their own `route_id` for query performance. When creating content fields, always include the `route_id` to match the parent content data's route.
- **Admin routes** are a separate system. The endpoints `/api/v1/adminroutes` and `/api/v1/admincontentdatas` manage content for the admin interface. Admin routes do not appear in the public content delivery endpoint.
