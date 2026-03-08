# Fetching Content

Recipes for retrieving published content from ModulaCMS. Content is served through slug-based routing -- each route maps a slug to a content tree. The content delivery endpoint assembles the tree, resolves references, and returns structured JSON.

For background on how routes and content trees work, see [routing](../guides/routing.md) and [content trees](../guides/content-trees.md).

## Get a Page by Slug

The primary frontend use case. Given a route slug, returns the full published content tree.

**curl:**

```bash
curl http://localhost:8080/api/v1/content/homepage
```

**Go SDK:**

```go
import (
    "context"
    "encoding/json"
    "fmt"

    modula "github.com/hegner123/modulacms/sdks/go"
)

client, err := modula.NewClient(modula.ClientConfig{
    BaseURL: "http://localhost:8080",
    APIKey:  "01HXK4N2F8RJZGP6VTQY3MCSW9",
})
if err != nil {
    // handle error
}

raw, err := client.Content.GetPage(ctx, "homepage", "", "")
if err != nil {
    if modula.IsNotFound(err) {
        fmt.Println("no published content at this slug")
        return
    }
    // handle error
}

fmt.Println(string(raw))
```

**TypeScript SDK (read-only):**

```typescript
import { ModulaClient } from '@modulacms/sdk'

const client = new ModulaClient({
  baseUrl: 'https://cms.example.com',
  apiKey: 'YOUR_API_KEY',
})

const page = await client.getPage('homepage')
console.log(page)
```

**TypeScript SDK (admin):**

```typescript
import { createAdminClient } from '@modulacms/admin-sdk'

const admin = createAdminClient({
  baseUrl: 'https://cms.example.com',
  apiKey: 'YOUR_API_KEY',
})

const page = await admin.contentDelivery.getPage('homepage')
console.log(page)
```

## Get a Page with a Specific Output Format

The `format` parameter controls the response shape. Supported values: `clean`, `raw`, `contentful`, `sanity`, `strapi`, `wordpress`.

| Format | Description |
|--------|-------------|
| `clean` | Flat object with resolved fields and children |
| `raw` | Internal tree structure with all IDs and metadata |
| `contentful` | Contentful-compatible structure |
| `sanity` | Sanity.io-compatible structure |
| `strapi` | Strapi-compatible structure |
| `wordpress` | WordPress-compatible structure |

**curl:**

```bash
curl "http://localhost:8080/api/v1/content/homepage?format=clean"
```

**Go SDK:**

```go
raw, err := client.Content.GetPage(ctx, "homepage", "clean", "")
```

**TypeScript SDK (read-only):**

```typescript
const page = await client.getPage('homepage', { format: 'clean' })
```

**TypeScript SDK (admin):**

```typescript
const page = await admin.contentDelivery.getPage('homepage', 'clean')
```

## Get a Page with Locale

Request content translated to a specific locale. When the locale has no translation, the CMS falls back along the locale's fallback chain (e.g., `fr-CA` falls back to `fr`, then to the default locale).

**curl:**

```bash
curl "http://localhost:8080/api/v1/content/homepage?locale=fr"
```

**Go SDK:**

```go
raw, err := client.Content.GetPage(ctx, "homepage", "", "fr")
```

**TypeScript SDK (read-only):**

```typescript
const page = await client.getPage('homepage', { locale: 'fr' })
```

Combine format and locale:

```bash
curl "http://localhost:8080/api/v1/content/homepage?format=clean&locale=fr"
```

```go
raw, err := client.Content.GetPage(ctx, "homepage", "clean", "fr")
```

```typescript
const page = await client.getPage('homepage', { format: 'clean', locale: 'fr' })
```

## Get a Single Content Data Entry by ID

Fetch a single content node by its ULID. Useful when you already have the content ID from a relation or query result.

**curl:**

```bash
curl "http://localhost:8080/api/v1/contentdata/?q=01HXK4N2F8RJZGP6VTQY3MCSW9" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
content, err := client.ContentData.Get(ctx, modula.ContentID("01HXK4N2F8RJZGP6VTQY3MCSW9"))
```

**TypeScript SDK (admin):**

```typescript
const content = await admin.contentData.get('01HXK4N2F8RJZGP6VTQY3MCSW9' as ContentID)
```

## List All Content for a Route

List all content data entries, then filter by route ID client-side. Or use the content delivery endpoint to get the full assembled tree for a route.

**curl (list all content data):**

```bash
curl http://localhost:8080/api/v1/contentdata \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
allContent, err := client.ContentData.List(ctx)
if err != nil {
    // handle error
}

// Filter to a specific route
targetRouteID := modula.RouteID("01HXK4N2F8RJZGP6VTQY3MCSW9")
for _, c := range allContent {
    if c.RouteID != nil && *c.RouteID == targetRouteID {
        fmt.Printf("Content: %s (status: %s)\n", c.ContentDataID, c.Status)
    }
}
```

**TypeScript SDK (admin):**

```typescript
const allContent = await admin.contentData.list()

const targetRouteID = '01HXK4N2F8RJZGP6VTQY3MCSW9'
const routeContent = allContent.filter(c => c.route_id === targetRouteID)
```

## Get Content Fields for a Content Data Entry

Retrieve all field values for a specific content node. Fields are stored separately from content data and linked by `content_data_id`.

**curl:**

```bash
curl http://localhost:8080/api/v1/contentfields \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
allFields, err := client.ContentFields.List(ctx)
if err != nil {
    // handle error
}

targetContentID := modula.ContentID("01HXK4N2F8RJZGP6VTQY3MCSW9")
for _, f := range allFields {
    if f.ContentDataID != nil && *f.ContentDataID == targetContentID {
        fmt.Printf("Field %s = %s\n", f.FieldID, f.FieldValue)
    }
}
```

**TypeScript SDK (read-only):**

```typescript
const fields = await client.listContentFields()

const targetID = '01HXK4N2F8RJZGP6VTQY3MCSW9'
const nodeFields = fields.filter(f => f.content_data_id === targetID)
for (const f of nodeFields) {
  console.log(`${f.field_id} = ${f.field_value}`)
}
```

**TypeScript SDK (admin):**

```typescript
const fields = await admin.contentFields.list()
const nodeFields = fields.filter(f => f.content_data_id === targetID)
```

## Error Handling

All SDKs return typed errors that can be inspected for HTTP status codes.

**Go SDK:**

```go
raw, err := client.Content.GetPage(ctx, "nonexistent", "", "")
if err != nil {
    if modula.IsNotFound(err) {
        // 404 -- no published content at this slug
    }
    if modula.IsUnauthorized(err) {
        // 401 -- invalid or missing API key
    }
    var apiErr *modula.ApiError
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

**TypeScript SDK:**

```typescript
import { ApiError, isNotFound, isUnauthorized } from '@modulacms/sdk'

try {
  const page = await client.getPage('nonexistent')
} catch (err) {
  if (isNotFound(err)) {
    // 404
  }
  if (isUnauthorized(err)) {
    // 401
  }
}
```
