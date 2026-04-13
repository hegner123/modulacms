# Querying Content

Query published content by datatype with filters, sorting, and pagination.

## Query endpoint

The query endpoint returns content items of a specific datatype:

```
GET /api/v1/query/{datatype_slug}
```

This is a public endpoint. It returns published content by default.

```bash
curl "http://localhost:8080/api/v1/query/blog-post"
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", nil)
if err != nil {
    // handle error
}

fmt.Printf("Found %d items (showing %d)\n", result.Total, len(result.Data))
for _, item := range result.Data {
    fmt.Printf("  %s: %s\n", item.ContentDataID, item.Fields["title"])
}
```

**TypeScript SDK:**

```typescript
const result = await client.queryContent('blog-post')

console.log(`Found ${result.total} items`)
for (const item of result.data) {
  console.log(`  ${item.content_data_id}: ${item.fields.title}`)
}
```

## Query parameters

All parameters are optional. Omitting all parameters returns the first 20 published items in default sort order.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `sort` | string | -- | Sort field. Prefix with `-` for descending. |
| `limit` | int | 20 | Maximum items to return. Clamped to 100. |
| `offset` | int | 0 | Items to skip for pagination. |
| `locale` | string | default locale | Locale code for internationalized content. |
| `status` | string | `published` | Content status filter (`published`, `draft`). |
| `{field}` | string | -- | Exact match filter on a field name. |
| `{field}[op]` | string | -- | Operator-based filter (see filter syntax). |

## Filter syntax

Pass field filters as query parameters. A bare field name matches exactly. Append an operator in brackets for comparison operations.

| Operator | Syntax | Description |
|----------|--------|-------------|
| `eq` (default) | `field=value` or `field[eq]=value` | Exact match |
| `ne` | `field[ne]=value` | Not equal |
| `gt` | `field[gt]=value` | Greater than |
| `gte` | `field[gte]=value` | Greater than or equal |
| `lt` | `field[lt]=value` | Less than |
| `lte` | `field[lte]=value` | Less than or equal |
| `like` | `field[like]=%pattern%` | SQL LIKE pattern match |
| `in` | `field[in]=a,b,c` | Match any of the listed values |

Multiple filters combine with AND logic.

> **Good to know**: Filter and sort field names must match the `name` property of the field definition on the datatype, not the `label`. Use the schema API to look up field names.

## Filter by field value

**curl:**

```bash
curl "http://localhost:8080/api/v1/query/blog-post?category=tutorials"
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Filters: map[string]string{
        "category": "tutorials",
    },
})
```

**TypeScript SDK:**

```typescript
const result = await client.queryContent('blog-post', {
  filters: { category: 'tutorials' },
})
```

### Combine filters

Pass multiple filter parameters. All filters apply with AND logic.

```bash
curl "http://localhost:8080/api/v1/query/blog-post?category=news&sort=-date_created&limit=5"
```

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Sort:  "-date_created",
    Limit: 5,
    Filters: map[string]string{
        "category": "news",
    },
})
```

```typescript
const result = await client.queryContent('blog-post', {
  sort: '-date_created',
  limit: 5,
  filters: { category: 'news' },
})
```

### Range filters

Use `[gt]`, `[gte]`, `[lt]`, `[lte]` for range queries. Combine two operators on the same field for a range.

```bash
curl "http://localhost:8080/api/v1/query/product?price[gte]=10&price[lte]=50&sort=price"
```

```go
result, err := client.Query.Query(ctx, "product", &modula.QueryParams{
    Filters: map[string]string{
        "price[gte]": "10",
        "price[lte]": "50",
    },
    Sort: "price",
})
```

```typescript
const result = await client.queryContent('product', {
  filters: {
    'price[gte]': '10',
    'price[lte]': '50',
  },
  sort: 'price',
})
```

### Pattern match with "like"

Use the `[like]` operator for SQL LIKE pattern matching. `%` matches any sequence of characters.

```bash
curl "http://localhost:8080/api/v1/query/blog-post?title[like]=%25tutorial%25"
```

> **Good to know**: `%25` is the URL encoding of `%`.

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Filters: map[string]string{
        "title[like]": "%tutorial%",
    },
})
```

```typescript
const result = await client.queryContent('blog-post', {
  filters: { 'title[like]': '%tutorial%' },
})
```

### Match multiple values with "in"

```bash
curl "http://localhost:8080/api/v1/query/blog-post?category[in]=news,tutorials,updates"
```

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Filters: map[string]string{
        "category[in]": "news,tutorials,updates",
    },
})
```

```typescript
const result = await client.queryContent('blog-post', {
  filters: { 'category[in]': 'news,tutorials,updates' },
})
```

## Sort results

Pass the field name to sort ascending, or prefix with `-` to sort descending. One sort field per request.

```
?sort=title          # ascending by title
?sort=-published_at  # descending by publish date (newest first)
?sort=-date_created  # descending by creation date
```

**curl:**

```bash
curl "http://localhost:8080/api/v1/query/blog-post?sort=-date_created"
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Sort: "-date_created",
})
```

**TypeScript SDK:**

```typescript
const result = await client.queryContent('blog-post', {
  sort: '-date_created',
})
```

## Paginate results

Use `limit` and `offset` for page-based pagination.

```bash
# Page 1 (items 1-10)
curl "http://localhost:8080/api/v1/query/blog-post?limit=10&offset=0"

# Page 2 (items 11-20)
curl "http://localhost:8080/api/v1/query/blog-post?limit=10&offset=10"

# Page 3 (items 21-30)
curl "http://localhost:8080/api/v1/query/blog-post?limit=10&offset=20"
```

**Go SDK:**

```go
pageSize := 10
page := 2

result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Limit:  pageSize,
    Offset: (page - 1) * pageSize,
})
if err != nil {
    // handle error
}

totalPages := (int(result.Total) + pageSize - 1) / pageSize
fmt.Printf("Page %d of %d (%d total items)\n", page, totalPages, result.Total)
```

**TypeScript SDK:**

```typescript
const pageSize = 10
const page = 2

const result = await client.queryContent('blog-post', {
  limit: pageSize,
  offset: (page - 1) * pageSize,
})

const totalPages = Math.ceil(result.total / result.limit)
console.log(`Page ${page} of ${totalPages} (${result.total} total items)`)
```

> **Good to know**: The `limit` parameter is clamped to a server-side maximum of 100. Values above 100 are silently reduced.

## Query localized content

Pass a locale code to filter content by locale.

```bash
curl "http://localhost:8080/api/v1/query/blog-post?locale=fr&sort=-date_created"
```

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Locale: "fr",
    Sort:   "-date_created",
})
```

```typescript
const result = await client.queryContent('blog-post', {
  locale: 'fr',
  sort: '-date_created',
})
```

## Query draft content

By default, queries return published content only. Pass `status=draft` to see unpublished content (requires authentication).

```bash
curl "http://localhost:8080/api/v1/query/blog-post?status=draft" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Status: "draft",
})
```

```typescript
const result = await admin.query.query('blog-post', {
  status: 'draft',
})
```

## Response structure

```json
{
  "data": [
    {
      "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
      "datatype_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
      "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
      "status": "published",
      "date_created": "2026-02-27T10:00:00Z",
      "date_modified": "2026-03-01T15:30:00Z",
      "published_at": "2026-03-01T15:30:00Z",
      "fields": {
        "title": "Getting Started with ModulaCMS",
        "body": "<p>Welcome to the guide...</p>",
        "category": "tutorials",
        "views": "142"
      }
    }
  ],
  "total": 47,
  "limit": 20,
  "offset": 0,
  "datatype": {
    "name": "blog-post",
    "label": "Blog Post"
  }
}
```

| Field | Description |
|-------|-------------|
| `data` | Array of content items matching the query |
| `total` | Total number of matching items across all pages |
| `limit` | Page size applied to this query |
| `offset` | Number of items skipped before this page |
| `datatype` | Metadata about the queried datatype (name and label) |

Each item in `data` contains:

| Field | Description |
|-------|-------------|
| `content_data_id` | Unique identifier for the content item |
| `datatype_id` | Identifier of the datatype |
| `author_id` | Identifier of the author |
| `status` | Content status (`published`, `draft`) |
| `date_created` | ISO 8601 creation timestamp |
| `date_modified` | ISO 8601 last modification timestamp |
| `published_at` | ISO 8601 publication timestamp (empty if never published) |
| `fields` | Map of field name to field value (all values are strings) |

> **Good to know**: All field values are serialized as strings regardless of their underlying field type. Parse numeric, boolean, JSON, and date values according to the field type metadata from the schema API.

## Next steps

- [Serving your frontend](/docs/building-content/serving-your-frontend) -- wire up a frontend framework to display queried content
- [Routing](/docs/building-content/routing) -- create routes and deliver content by slug
- [Media](/docs/building-content/media) -- upload and serve responsive images
