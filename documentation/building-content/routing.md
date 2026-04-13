# Routing

Routes map URL slugs to content trees, making your content addressable and deliverable over HTTP.

## What is a route?

A **route** binds a unique slug to a content tree. Each route has a slug, a title, and a status. When a frontend requests content for a slug, ModulaCMS resolves the route, assembles the content tree, and returns it as JSON.

A **slug** is the unique identifier for a route. Slugs are freeform text -- a domain name (`example.com`), a path (`/about`), a keyword (`homepage`), or any string that fits your use case. The only constraint is uniqueness across all routes.

```
Route: slug = "homepage"
  +-- Content tree root ("Page")
      +-- title: "Welcome"
      +-- Hero Section
      |   +-- heading: "Hello"
      +-- Cards
          +-- Card
          +-- Card
```

## Create a route

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

### Route status

Routes use an integer status field:

| Value | Meaning |
|-------|---------|
| 0 | Inactive |
| 1 | Active |

### Set up a route's content tree

After creating a route, create a root content node for it. The root node's datatype must have `type = "_root"`:

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

This gives the route an empty content tree ready for child nodes. See [content trees](/docs/building-content/content-trees) for details on building out the tree.

## Manage routes

### List routes

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

### Get a single route

```bash
curl "http://localhost:8080/api/v1/routes/?q=01JNRW9P2DKTZ6Q4M8W3B5J7CL" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Update a route

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

### Delete a route

```bash
curl -X DELETE "http://localhost:8080/api/v1/routes/?q=01JNRW9P2DKTZ6Q4M8W3B5J7CL" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Deleting a route cascades: all content data and content field records belonging to that route are permanently deleted. There is no undo. Back up route content before deletion.

## Deliver content by slug

The public content delivery endpoint resolves a slug and returns the assembled content tree:

```bash
curl http://localhost:8080/api/v1/content/homepage
```

This endpoint is public (no authentication required).

### Output formats

The response structure can match other CMS conventions. Set the default in `modula.config.json`:

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
| `contentful` | Contentful-compatible response format |
| `sanity` | Sanity.io-compatible response format |
| `strapi` | Strapi-compatible response format |
| `wordpress` | WordPress REST API-compatible response format |

> **Good to know**: If you're migrating from another CMS, use the matching output format so your existing frontend code works with minimal changes.

## Build navigation from routes

Routes are the foundation of your site's navigation. Fetch all active routes and map them into menu items.

### Flat nav menu

**TypeScript SDK:**

```typescript
const routes = await client.listRoutes()

const nav = routes
  .filter(r => r.status === 1)
  .map(r => ({
    label: r.title,
    href: `/${r.slug}`,
  }))
```

**Go SDK:**

```go
routes, err := client.Routes.List(ctx)
if err != nil {
    // handle error
}

type NavItem struct {
    Label string
    Href  string
}

var nav []NavItem
for _, r := range routes {
    if r.Status != 1 {
        continue
    }
    nav = append(nav, NavItem{
        Label: r.Title,
        Href:  "/" + string(r.Slug),
    })
}
```

### Hierarchical nav menu

Routes are flat, but you can use slug conventions to build nested navigation. Slugs like `blog`, `blog/tutorials`, and `blog/news` imply a hierarchy.

**TypeScript SDK:**

```typescript
interface NavNode {
  label: string
  href: string
  children: NavNode[]
}

const routes = await client.listRoutes()

const nodes = new Map<string, NavNode>()
const roots: NavNode[] = []

for (const r of routes.filter(r => r.status === 1)) {
  const node: NavNode = {
    label: r.title,
    href: `/${r.slug}`,
    children: [],
  }
  nodes.set(r.slug, node)

  const lastSlash = r.slug.lastIndexOf('/')
  if (lastSlash > 0) {
    const parentSlug = r.slug.substring(0, lastSlash)
    const parent = nodes.get(parentSlug)
    if (parent) {
      parent.children.push(node)
      continue
    }
  }

  roots.push(node)
}
```

**Go SDK:**

```go
type NavNode struct {
    Label    string
    Href     string
    Children []*NavNode
}

routes, err := client.Routes.List(ctx)
if err != nil {
    // handle error
}

nodes := make(map[string]*NavNode)
var roots []*NavNode

for _, r := range routes {
    if r.Status != 1 {
        continue
    }

    slug := string(r.Slug)
    node := &NavNode{
        Label: r.Title,
        Href:  "/" + slug,
    }
    nodes[slug] = node

    lastSlash := strings.LastIndex(slug, "/")
    if lastSlash > 0 {
        parentSlug := slug[:lastSlash]
        if parent, ok := nodes[parentSlug]; ok {
            parent.Children = append(parent.Children, node)
            continue
        }
    }

    roots = append(roots, node)
}
```

### Fetch content for a route

Once you have a route slug, fetch its full content tree:

**curl (public):**

```bash
curl http://localhost:8080/api/v1/content/homepage
```

**Go SDK:**

```go
page, err := client.Content.GetPage(ctx, "homepage", "clean", "")
```

**TypeScript SDK:**

```typescript
const page = await client.getPage('homepage', { format: 'clean' })
```

> **Good to know**: Admin routes are separate from content-facing routes. Use `/api/v1/routes` for public navigation and `/api/v1/adminroutes` for admin navigation.

## Configuration

Route-related configuration in `modula.config.json`:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `output_format` | string | `""` (raw) | Default output format for content delivery |
| `client_site` | string | -- | Client site URL used in output format transformations |
| `space_id` | string | -- | Space identifier used in Contentful-style output format |
| `composition_max_depth` | integer | 10 | Maximum depth for composing referenced content subtrees |

## API reference

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/routes` | `routes:read` | List all routes (supports `limit` and `offset`) |
| POST | `/api/v1/routes` | `routes:create` | Create a route |
| GET | `/api/v1/routes/` | `routes:read` | Get a single route (`?q=ROUTE_ID`) |
| PUT | `/api/v1/routes/` | `routes:update` | Update a route |
| DELETE | `/api/v1/routes/` | `routes:delete` | Delete a route and all its content (cascade) |
| GET | `/api/v1/content/{slug}` | Public | Deliver assembled content tree for a slug |

## Next steps

- [Serving your frontend](/docs/building-content/serving-your-frontend) -- wire up a frontend framework to ModulaCMS
- [Querying content](/docs/building-content/querying) -- query content by datatype with filters and pagination
- [Media](/docs/building-content/media) -- upload and serve responsive images
