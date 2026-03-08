# Building Navigation

Recipes for building navigation menus from ModulaCMS routes. Routes are lightweight connectors between a slug and a content tree. The routes API returns the data needed to build nav menus, breadcrumbs, and sitemaps.

For background on how routes work, see [routing](../guides/routing.md).

## Route Types

Routes in ModulaCMS have a slug, title, and status. The `status` field is a numeric flag:

| Status | Meaning |
|--------|---------|
| `0` | Inactive (hidden) |
| `1` | Active (visible) |

Admin routes are a separate system from content-facing routes. Use `/api/v1/routes` for public navigation and `/api/v1/adminroutes` for admin navigation.

## List All Routes

**curl:**

```bash
curl http://localhost:8080/api/v1/routes \
  -H "Authorization: Bearer YOUR_API_KEY"
```

Response:

```json
[
  {
    "route_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
    "slug": "homepage",
    "title": "Home Page",
    "status": 1,
    "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
    "date_created": "2026-01-15T10:00:00Z",
    "date_modified": "2026-01-15T10:00:00Z"
  },
  {
    "route_id": "01HXK5P3G7SWAH2C6VRXE8Q1KN",
    "slug": "about",
    "title": "About Us",
    "status": 1,
    "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
    "date_created": "2026-01-15T10:05:00Z",
    "date_modified": "2026-01-15T10:05:00Z"
  }
]
```

**Go SDK:**

```go
routes, err := client.Routes.List(ctx)
if err != nil {
    // handle error
}

for _, r := range routes {
    fmt.Printf("%s -> /%s\n", r.Title, r.Slug)
}
```

**TypeScript SDK (read-only):**

```typescript
const routes = await client.listRoutes()

for (const r of routes) {
  console.log(`${r.title} -> /${r.slug}`)
}
```

**TypeScript SDK (admin):**

```typescript
const routes = await admin.routes.list()
```

## Filter Routes by Status

Only show active routes in navigation.

**Go SDK:**

```go
routes, err := client.Routes.List(ctx)
if err != nil {
    // handle error
}

var activeRoutes []modula.Route
for _, r := range routes {
    if r.Status == 1 {
        activeRoutes = append(activeRoutes, r)
    }
}
```

**TypeScript SDK:**

```typescript
const routes = await client.listRoutes()
const activeRoutes = routes.filter(r => r.status === 1)
```

## Build a Nav Menu from Routes

Map routes into a navigation structure suitable for rendering.

**Go SDK:**

```go
type NavItem struct {
    Label string
    Href  string
}

routes, err := client.Routes.List(ctx)
if err != nil {
    // handle error
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

// Use nav to render your menu
for _, item := range nav {
    fmt.Printf("<a href=\"%s\">%s</a>\n", item.Href, item.Label)
}
```

**TypeScript SDK:**

```typescript
interface NavItem {
  label: string
  href: string
}

const routes = await client.listRoutes()

const nav: NavItem[] = routes
  .filter(r => r.status === 1)
  .map(r => ({
    label: r.title,
    href: `/${r.slug}`,
  }))

// Render as HTML
const html = nav
  .map(item => `<a href="${item.href}">${item.label}</a>`)
  .join('\n')
```

## Build a Hierarchical Nav Menu

Routes are flat, but you can use slug conventions to build nested navigation. For example, slugs like `blog`, `blog/tutorials`, and `blog/news` imply a hierarchy.

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

// Build lookup by slug
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

    // Find parent by trimming the last segment
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

## Ordered Admin Routes

Admin routes support server-side ordering via the `?ordered=true` parameter. This reads each route's root content node "Order" field and sorts numerically.

**curl:**

```bash
curl "http://localhost:8080/api/v1/adminroutes?ordered=true" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
// Admin routes use the AdminRoutes resource
adminRoutes, err := client.AdminRoutes.List(ctx)
if err != nil {
    // handle error
}
```

**TypeScript SDK (admin):**

```typescript
// Ordered admin routes for sidebar navigation
const adminRoutes = await admin.adminRoutes.listOrdered()

const adminNav = adminRoutes.map(r => ({
  label: r.title,
  href: `/admin/${r.slug}`,
}))
```

## Get the Content Tree for a Route

Once you have a route slug, you can fetch its full content tree via the content delivery endpoint or the admin tree endpoint.

**curl (public content delivery):**

```bash
curl http://localhost:8080/api/v1/content/homepage
```

**curl (admin tree):**

```bash
curl http://localhost:8080/api/v1/admin/tree/homepage \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
// Public content
page, err := client.Content.GetPage(ctx, "homepage", "clean", "")

// Admin tree
tree, err := client.AdminTree.Get(ctx, "homepage", "")
```

**TypeScript SDK:**

```typescript
// Public content
const page = await admin.contentDelivery.getPage('homepage')

// Admin tree
const tree = await admin.adminTree.get('homepage' as Slug)
```

## Next Steps

- [Fetching Content](fetching-content.md) -- full content retrieval recipes
- [Search and Filter](search-and-filter.md) -- query content by datatype and field values
