# Search and Filter

Recipes for querying and filtering content by datatype, field values, and status. The `/query/{datatype}` endpoint returns content entries matching a datatype name with optional field filters, sorting, and pagination.

For background on content modeling, see [content modeling](../guides/content-modeling.md).

## Query Parameters Reference

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `sort` | string | server default | Field name to sort by. Prefix with `-` for descending. |
| `limit` | int | `20` | Maximum items to return (server max: `100`). |
| `offset` | int | `0` | Number of items to skip for pagination. |
| `locale` | string | default locale | Locale code for internationalized content. |
| `status` | string | `published` | Content status filter (`published`, `draft`). |

## Filter Operators

Field filters are passed as query parameters. The key is the field name, optionally suffixed with a bracket operator.

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

## Query by Datatype Slug

The primary query use case. Given a datatype name (e.g., `blog-post`), returns all published content of that type.

**curl:**

```bash
curl "http://localhost:8080/api/v1/query/blog-post" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

Response:

```json
{
  "data": [
    {
      "content_data_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
      "datatype_id": "01HXK3M1E6PWZF9C5TNBQ4H7KJ",
      "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
      "status": "published",
      "date_created": "2026-01-15T10:00:00Z",
      "date_modified": "2026-02-01T14:30:00Z",
      "published_at": "2026-02-01T14:30:00Z",
      "fields": {
        "title": "Getting Started with ModulaCMS",
        "body": "<p>Welcome to the guide...</p>",
        "category": "tutorials",
        "featured_image": "01HXK6R4J9TYEM3A7WBQN5F2PK"
      }
    }
  ],
  "total": 42,
  "limit": 20,
  "offset": 0,
  "datatype": {
    "name": "blog-post",
    "label": "Blog Post"
  }
}
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

**TypeScript SDK (read-only):**

```typescript
const result = await client.queryContent('blog-post')

console.log(`Found ${result.total} items`)
for (const item of result.data) {
  console.log(`  ${item.content_data_id}: ${item.fields.title}`)
}
```

**TypeScript SDK (admin):**

```typescript
const result = await admin.query.query('blog-post')
```

## Filter by Field Value

Pass field filters as query parameters. Bare field names use exact match.

**curl:**

```bash
curl "http://localhost:8080/api/v1/query/blog-post?category=news" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Filters: map[string]string{
        "category": "news",
    },
})
```

**TypeScript SDK (read-only):**

```typescript
const result = await client.queryContent('blog-post', {
  filters: { category: 'news' },
})
```

**TypeScript SDK (admin):**

```typescript
const result = await admin.query.query('blog-post', {
  filters: { category: 'news' },
})
```

## Sort by Date Descending

Prefix the sort field with `-` for descending order.

**curl:**

```bash
curl "http://localhost:8080/api/v1/query/blog-post?sort=-date_created" \
  -H "Authorization: Bearer YOUR_API_KEY"
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

## Paginate Results

Use `limit` and `offset` for page-based pagination.

**curl:**

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
page := 2 // 1-indexed

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

## Combine Filters

Pass multiple filter parameters to combine them with AND logic.

**curl:**

```bash
curl "http://localhost:8080/api/v1/query/blog-post?category=news&status=published&sort=-date_created&limit=5" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Status: "published",
    Sort:   "-date_created",
    Limit:  5,
    Filters: map[string]string{
        "category": "news",
    },
})
```

**TypeScript SDK:**

```typescript
const result = await client.queryContent('blog-post', {
  status: 'published',
  sort: '-date_created',
  limit: 5,
  filters: {
    category: 'news',
  },
})
```

## Full-Text Search with "like" Operator

Use the `[like]` operator suffix for SQL LIKE pattern matching. `%` matches any sequence of characters.

**curl:**

```bash
# Title contains "tutorial"
curl "http://localhost:8080/api/v1/query/blog-post?title[like]=%25tutorial%25" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

Note: `%25` is the URL encoding of `%`.

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Filters: map[string]string{
        "title[like]": "%tutorial%",
    },
})
```

**TypeScript SDK:**

```typescript
const result = await client.queryContent('blog-post', {
  filters: {
    'title[like]': '%tutorial%',
  },
})
```

## Range Filters

Use `[gt]`, `[gte]`, `[lt]`, `[lte]` for range queries. Combine two operators on the same field for a range.

**curl:**

```bash
# Price between 10 and 50
curl "http://localhost:8080/api/v1/query/product?price[gte]=10&price[lte]=50" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "product", &modula.QueryParams{
    Filters: map[string]string{
        "price[gte]": "10",
        "price[lte]": "50",
    },
})
```

**TypeScript SDK:**

```typescript
const result = await client.queryContent('product', {
  filters: {
    'price[gte]': '10',
    'price[lte]': '50',
  },
})
```

## Match Any of Multiple Values

Use the `[in]` operator with comma-separated values.

**curl:**

```bash
curl "http://localhost:8080/api/v1/query/blog-post?category[in]=news,tutorials,updates" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Filters: map[string]string{
        "category[in]": "news,tutorials,updates",
    },
})
```

**TypeScript SDK:**

```typescript
const result = await client.queryContent('blog-post', {
  filters: {
    'category[in]': 'news,tutorials,updates',
  },
})
```

## Query with Locale

Filter content by locale for internationalized sites.

**curl:**

```bash
curl "http://localhost:8080/api/v1/query/blog-post?locale=fr&sort=-date_created" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Locale: "fr",
    Sort:   "-date_created",
})
```

**TypeScript SDK:**

```typescript
const result = await client.queryContent('blog-post', {
  locale: 'fr',
  sort: '-date_created',
})
```

## Query Draft Content

By default, queries return published content only. Pass `status=draft` to see unpublished content (requires authentication).

**curl:**

```bash
curl "http://localhost:8080/api/v1/query/blog-post?status=draft" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Status: "draft",
})
```

**TypeScript SDK:**

```typescript
const result = await admin.query.query('blog-post', {
  status: 'draft',
})
```

## Next Steps

- [Fetching Content](fetching-content.md) -- retrieve content by slug
- [Building Navigation](building-navigation.md) -- build menus from routes
