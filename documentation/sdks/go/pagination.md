# Pagination

The Go SDK supports paginated listings for all generic `Resource` types. This guide covers the pagination types, the `ListPaginated` method, iteration patterns, and counting.

## Types

### PaginationParams

```go
type PaginationParams struct {
    Limit  int64 // Maximum number of items to return.
    Offset int64 // Number of items to skip.
}
```

Both fields are sent as query parameters. A zero `Limit` uses the server default. A zero `Offset` starts from the beginning.

### PaginatedResponse

```go
type PaginatedResponse[T any] struct {
    Data   []T   `json:"data"`   // Items for the current page.
    Total  int64 `json:"total"`  // Total items across all pages.
    Limit  int64 `json:"limit"`  // Page size used.
    Offset int64 `json:"offset"` // Items skipped.
}
```

`Total` is the total number of items matching the query, not just the current page. Use it for building pagination controls in your UI.

## ListPaginated vs List

Every generic `Resource` exposes both `List` and `ListPaginated`:

| Method | Returns | Use when |
|--------|---------|----------|
| `List(ctx)` | `([]E, error)` | You want all items and know the dataset is small. |
| `ListPaginated(ctx, params)` | `(*PaginatedResponse[E], error)` | You need pagination metadata or the dataset may be large. |

Prefer `ListPaginated` unless you are certain the total count is small (under a few hundred items). `List` fetches the server's default page, which may not include all items.

## Basic Usage

```go
page, err := client.Datatypes.ListPaginated(ctx, modula.PaginationParams{
    Limit:  25,
    Offset: 0,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Page 1: %d items, %d total\n", len(page.Data), page.Total)
```

## Iteration Pattern

To iterate through all items across multiple pages:

```go
var allUsers []modula.User

limit := int64(50)
offset := int64(0)

for {
    page, err := client.Users.ListPaginated(ctx, modula.PaginationParams{
        Limit:  limit,
        Offset: offset,
    })
    if err != nil {
        return fmt.Errorf("list users at offset %d: %w", offset, err)
    }
    allUsers = append(allUsers, page.Data...)

    offset += limit
    if offset >= page.Total {
        break
    }
}

fmt.Printf("Fetched %d users total\n", len(allUsers))
```

### Iteration with Early Exit

When processing items as you go and you do not need the full list in memory:

```go
limit := int64(100)
offset := int64(0)

for {
    page, err := client.ContentData.ListPaginated(ctx, modula.PaginationParams{
        Limit:  limit,
        Offset: offset,
    })
    if err != nil {
        return err
    }
    for _, item := range page.Data {
        if err := processItem(ctx, item); err != nil {
            return fmt.Errorf("process %s: %w", item.ContentDataID, err)
        }
    }
    offset += limit
    if offset >= page.Total {
        break
    }
}
```

## Counting Without Fetching Data

To get the total count without transferring any entity data:

```go
total, err := client.ContentData.Count(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%d content items\n", total)
```

`Count` makes a lightweight request that returns only the total. Use it for dashboard summaries, progress indicators, or deciding whether to paginate at all.

## Combining with QueryResource

The `QueryResource` has its own pagination built into `QueryParams`, separate from the generic `ListPaginated` method:

```go
result, err := client.Query.Query(ctx, "blog-posts", &modula.QueryParams{
    Limit:  10,
    Offset: 20,
    Sort:   "-date_created",
    Status: "published",
})
// result.Total, result.Limit, result.Offset work the same way.
```

See [Content Operations](content-operations.md#querying-content) for full query documentation.

## Next Steps

- [Content Operations](content-operations.md) -- CRUD, trees, publishing, querying
- [Error Handling](error-handling.md) -- handling pagination errors
- [Reference](reference.md) -- full resource index
