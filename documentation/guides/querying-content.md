# Querying Content

The query API provides filtered, sorted, paginated access to content by datatype. This is the primary endpoint for frontend applications that need to list, search, and paginate content -- blog post lists, product catalogs, news feeds, or any collection of content items.

## Concepts

**Datatype slug** -- The machine-readable name of a datatype (set when creating the datatype). Queries are scoped to a single datatype, identified by its slug in the URL path.

**Field filter** -- A key-value pair that restricts results to content where a named field matches a condition. Filters support exact match and operator-based comparisons.

**Sort** -- A field name that determines result ordering. Prefix with `-` for descending order. Sortable fields include custom fields and built-in fields (`date_created`, `date_modified`, `published_at`).

## Query Endpoint

```
GET /api/v1/query/{datatype_slug}
```

This is a public endpoint. It returns published content by default.

## Query Parameters

All parameters are optional. Omitting all parameters returns the first 20 published items in default sort order.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `sort` | string | -- | Sort field. Prefix with `-` for descending. |
| `limit` | int | 20 | Maximum items to return. Clamped to 100. |
| `offset` | int | 0 | Items to skip for pagination. |
| `locale` | string | default locale | Locale code for internationalized content. |
| `status` | string | `published` | Content status filter (`published`, `draft`). |
| `{field}` | string | -- | Exact match filter on a field name. |
| `{field}[op]` | string | -- | Operator-based filter (see below). |

### Filter Syntax

Filters are passed as query parameters. A bare field name matches exactly. Append an operator in brackets for comparison operations.

**Supported operators:**

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equal (same as bare field name) | `?category[eq]=news` |
| `ne` | Not equal | `?status[ne]=draft` |
| `gt` | Greater than | `?price[gt]=10` |
| `gte` | Greater than or equal | `?price[gte]=10` |
| `lt` | Less than | `?price[lt]=100` |
| `lte` | Less than or equal | `?price[lte]=100` |
| `like` | SQL LIKE pattern match (`%` is wildcard) | `?title[like]=%tutorial%` |
| `in` | Match any in comma-separated list | `?tag[in]=go,rust,zig` |

Multiple filters can be combined. All filters are applied with AND logic.

### Sort Syntax

Pass the field name to sort ascending, or prefix with `-` to sort descending. Only one sort field is supported per request.

```
?sort=title          # ascending by title
?sort=-published_at  # descending by publish date (newest first)
?sort=-date_created  # descending by creation date
```

## Response Structure

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
    "name": "blog-posts",
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
| `content_data_id` | ULID of the content item |
| `datatype_id` | ULID of the datatype |
| `author_id` | ULID of the author |
| `status` | Content status (`published`, `draft`, etc.) |
| `date_created` | ISO 8601 creation timestamp |
| `date_modified` | ISO 8601 last modification timestamp |
| `published_at` | ISO 8601 publication timestamp (empty if never published) |
| `fields` | Map of field name to field value (all values are strings) |

All field values are serialized as strings regardless of their underlying field type. Parse numeric, boolean, JSON, and date values according to the field type metadata from the schema API.

## Examples

### Recent Blog Posts

Fetch the 10 most recently published blog posts:

```bash
curl "http://localhost:8080/api/v1/query/blog-posts?sort=-published_at&limit=10&status=published"
```

### Filter by Category

Fetch tutorials, paginated:

```bash
curl "http://localhost:8080/api/v1/query/blog-posts?category=tutorials&limit=20&offset=0"
```

### Range Filter

Fetch products priced between 10 and 50:

```bash
curl "http://localhost:8080/api/v1/query/products?price[gte]=10&price[lte]=50&sort=price"
```

### Pattern Match

Fetch posts with "tutorial" in the title:

```bash
curl "http://localhost:8080/api/v1/query/blog-posts?title[like]=%25tutorial%25"
```

Note: `%25` is the URL-encoded form of `%` (the SQL LIKE wildcard).

### Multi-Value Filter

Fetch posts tagged with any of three languages:

```bash
curl "http://localhost:8080/api/v1/query/blog-posts?tag[in]=go,rust,zig"
```

### Localized Content

Fetch French translations of blog posts:

```bash
curl "http://localhost:8080/api/v1/query/blog-posts?locale=fr&status=published&sort=-published_at"
```

### Pagination

Calculate page boundaries using the response metadata:

```
Total items: 47
Page size:   20
Total pages: ceil(47 / 20) = 3
Current page: floor(offset / limit) + 1

Page 1: ?limit=20&offset=0
Page 2: ?limit=20&offset=20
Page 3: ?limit=20&offset=40
```

## SDK Examples

### Go

```go
import modula "github.com/hegner123/modulacms/sdks/go"

client, _ := modula.NewClient(modula.ClientConfig{
    BaseURL: "http://localhost:8080",
    APIKey:  "mcms_YOUR_API_KEY",
})

// Recent published posts
result, err := client.Query.Query(ctx, "blog-posts", &modula.QueryParams{
    Sort:   "-published_at",
    Limit:  10,
    Status: "published",
})

// Filtered by category with pagination
result, err = client.Query.Query(ctx, "blog-posts", &modula.QueryParams{
    Filters: map[string]string{
        "category": "tutorials",
    },
    Limit:  20,
    Offset: 40, // page 3
})

// Range filter on a numeric field
result, err = client.Query.Query(ctx, "products", &modula.QueryParams{
    Filters: map[string]string{
        "price[gte]": "10",
        "price[lte]": "50",
    },
    Sort: "price",
})

// Localized content
result, err = client.Query.Query(ctx, "blog-posts", &modula.QueryParams{
    Locale: "fr",
    Status: "published",
    Sort:   "-published_at",
})

// Iterate results
for _, item := range result.Data {
    title := item.Fields["title"]
    body := item.Fields["body"]
    // ...
}

// Pagination
totalPages := (result.Total + result.Limit - 1) / result.Limit
currentPage := result.Offset/result.Limit + 1
```

### TypeScript

```typescript
import { ModulaCMSAdmin } from '@modulacms/admin-sdk'

const client = new ModulaCMSAdmin({
  baseUrl: 'http://localhost:8080',
  apiKey: 'mcms_YOUR_API_KEY',
})

// Recent published posts
const result = await client.query('blog-posts', {
  sort: '-published_at',
  limit: 10,
  status: 'published',
})

// Filtered by category
const tutorials = await client.query('blog-posts', {
  filters: { category: 'tutorials' },
  limit: 20,
  offset: 40,
})

// Operator-based filtering
const affordable = await client.query('products', {
  filters: {
    'price[gte]': '10',
    'price[lte]': '50',
  },
  sort: 'price',
})

// Multi-value filter
const tagged = await client.query('blog-posts', {
  filters: { 'tag[in]': 'go,rust,zig' },
})

// Pagination
const totalPages = Math.ceil(result.total / result.limit)
const currentPage = Math.floor(result.offset / result.limit) + 1
```

## API Reference

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/query/{datatype_slug}` | Public | Query content by datatype with filters, sort, and pagination |

## Notes

- **String values.** All field values in the response are strings, regardless of field type. Numeric, boolean, and date fields must be parsed by the consumer according to the field type metadata.
- **Default status.** When no `status` parameter is provided, only `published` content is returned. To include drafts, explicitly set `?status=draft` or omit the status filter by passing an empty string.
- **Limit clamping.** The `limit` parameter is clamped to a server-side maximum of 100. Values above 100 are silently reduced.
- **Single sort field.** Only one sort field per request is supported. To sort by multiple criteria, sort in your application after fetching.
- **Filter AND logic.** Multiple filters are combined with AND. There is no OR support at the query parameter level -- use the `in` operator for OR-like behavior on a single field.
- **Field name matching.** Filter and sort field names must match the `name` property of the field definition on the datatype (not the `label`). Use the schema API to look up field names.
