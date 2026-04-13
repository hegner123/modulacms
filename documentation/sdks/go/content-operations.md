# Content Operations

Read, write, and organize content through the Go SDK -- CRUD, content trees, publishing, batch updates, querying, and content delivery.

## CRUD via Generic Resources

Most content resources on the client are instances of `Resource[Entity, CreateParams, UpdateParams, ID]`. This generic type provides a consistent set of methods:

| Method | Signature | Description |
|--------|-----------|-------------|
| `List` | `(ctx) ([]E, error)` | Fetch all items |
| `Get` | `(ctx, id) (*E, error)` | Fetch one item by ID |
| `Create` | `(ctx, params) (*E, error)` | Create a new item |
| `Update` | `(ctx, params) (*E, error)` | Update an existing item |
| `Delete` | `(ctx, id) error` | Delete an item by ID |
| `ListPaginated` | `(ctx, PaginationParams) (*PaginatedResponse[E], error)` | Paginated listing |
| `Count` | `(ctx) (int64, error)` | Total count without fetching data |
| `RawList` | `(ctx, url.Values) (json.RawMessage, error)` | Raw listing with arbitrary query params |

### Creating Content Data

```go
parentID := modula.ContentID("01HX...")
datatypeID := modula.DatatypeID("01HX...")

item, err := client.ContentData.Create(ctx, modula.CreateContentDataParams{
    ParentID:   &parentID,
    DatatypeID: &datatypeID,
    Status:     modula.ContentStatusDraft,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created content: %s\n", item.ContentDataID)
```

### Setting Field Values

After creating a content data node, set its field values:

```go
field, err := client.ContentFields.Create(ctx, modula.CreateContentFieldParams{
    ContentDataID: item.ContentDataID,
    FieldID:       titleFieldID,
    Value:         "Hello World",
})
if err != nil {
    log.Fatal(err)
}
```

### Updating Content

```go
updated, err := client.ContentData.Update(ctx, modula.UpdateContentDataParams{
    ContentDataID: item.ContentDataID,
    Status:        modula.ContentStatusDraft,
})
```

### Deleting Content

```go
err := client.ContentData.Delete(ctx, contentID)
if modula.IsNotFound(err) {
    // Already deleted or never existed.
}
```

### Composite Operations

For operations that span multiple tables atomically, use the composite resources:

```go
// Create content with fields in a single request.
resp, err := client.ContentComposite.CreateWithFields(ctx, modula.ContentCreateParams{
    DatatypeID: datatypeID,
    ParentID:   &parentID,
    Fields: map[string]string{
        "title": "New Post",
        "body":  "Content here...",
    },
})

// Delete content and all its descendants recursively.
delResp, err := client.ContentComposite.DeleteRecursive(ctx, contentID)
fmt.Printf("Deleted %d nodes\n", delResp.Deleted)
```

## Content Trees

Content in ModulaCMS is organized as trees. Each route has its own content tree, and global trees provide site-wide content like navigation and footers without a route.

### Getting a Tree by Route

```go
nodes, err := client.ContentTree.GetByRoute(ctx, routeID)
if err != nil {
    log.Fatal(err)
}
for _, node := range nodes {
    fmt.Printf("%s: %s (%s)\n", node.ContentID, node.Title, node.Status)
}
```

Each `ContentTreeNode` contains the content ID, datatype ID, route ID, title, slug, status, and timestamps.

### Bulk Tree Operations with Save

The `ContentTree.Save` method accepts creates, updates, and deletes in a single atomic request. Use `TreeNodeCreate` for new nodes and `TreeNodeUpdate` for pointer changes:

```go
resp, err := client.ContentTree.Save(ctx, modula.TreeSaveRequest{
    ContentID: rootContentID,
    Creates: []modula.TreeNodeCreate{
        {
            ClientID:   "temp-1",
            DatatypeID: string(cardDatatypeID),
            ParentID:   modula.StringPtr(string(cardsContainerID)),
        },
    },
    Updates: []modula.TreeNodeUpdate{
        {
            ContentDataID: existingNodeID,
            ParentID:      modula.StringPtr(string(newParentID)),
        },
    },
    Deletes: []modula.ContentID{obsoleteNodeID},
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created: %d, Updated: %d, Deleted: %d\n",
    resp.Created, resp.Updated, resp.Deleted)

// resp.IDMap maps client-generated temp IDs to server-assigned IDs.
realID := resp.IDMap["temp-1"]
```

The `ClientID` field on `TreeNodeCreate` is a temporary identifier you assign. After the save, `resp.IDMap` maps each `ClientID` to the server-assigned `ContentID`. Use `modula.StringPtr()` to create `*string` values for pointer fields.

### Reordering and Moving Nodes

Reorder children within a parent:

```go
resp, err := client.ContentReorder.Reorder(ctx, modula.ContentReorderRequest{
    ParentID:   &parentID,
    OrderedIDs: []modula.ContentID{thirdID, firstID, secondID},
})
```

Move a node to a different parent at a specific position:

```go
resp, err := client.ContentReorder.Move(ctx, modula.ContentMoveRequest{
    NodeID:      nodeID,
    NewParentID: &targetParentID,
    Position:    0, // Insert as first child
})
```

## Publishing

The `Publishing` resource manages the content lifecycle: draft, published, and scheduled.

> **Good to know**: The client exposes two publishing resources -- `Publishing` for public content and `AdminPublishing` for admin content -- with identical methods.

### Publish

Transition content from draft to published, creating a version snapshot:

```go
resp, err := client.Publishing.Publish(ctx, modula.PublishRequest{
    ContentDataID: contentID,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Published at %s, version %s\n", resp.PublishedAt, resp.VersionID)
```

### Unpublish

Revert published content back to draft:

```go
resp, err := client.Publishing.Unpublish(ctx, modula.PublishRequest{
    ContentDataID: contentID,
})
```

### Schedule

Set content to publish at a future time:

```go
resp, err := client.Publishing.Schedule(ctx, modula.ScheduleRequest{
    ContentDataID: contentID,
    PublishAt:     "2026-04-01T09:00:00Z",
})
```

### Version History

List all version snapshots for a content item:

```go
versions, err := client.Publishing.ListVersions(ctx, string(contentID))
for _, v := range versions {
    fmt.Printf("Version %s: created %s\n", v.ContentVersionID, v.DateCreated)
}
```

Get a specific version:

```go
version, err := client.Publishing.GetVersion(ctx, string(versionID))
```

### Restore a Version

Replace the current draft with a previous version's field values:

```go
resp, err := client.Publishing.Restore(ctx, modula.RestoreRequest{
    ContentDataID: contentID,
    VersionID:     versionID,
})
```

## Batch Updates

Send multiple content updates in a single request:

```go
raw, err := client.ContentBatch.Update(ctx, batchPayload)
```

Structure the batch payload based on your update needs. The SDK returns the response as `json.RawMessage` for flexibility.

## Content Delivery

The `Content` resource provides the public-facing endpoint for fetching rendered pages by slug. This is the primary read path for frontend applications:

```go
raw, err := client.Content.GetPage(ctx, "/about", "clean", "en-US")
if err != nil {
    log.Fatal(err)
}
// raw is json.RawMessage -- unmarshal into your application-specific type.
```

Parameters:
- **slug** -- The URL path of the content page (e.g., `"/about"`, `"/blog/my-post"`).
- **format** -- Response format: `"contentful"`, `"sanity"`, `"strapi"`, `"wordpress"`, `"clean"`, or `"raw"`. Each formats the tree differently for compatibility with various frontend frameworks.
- **locale** -- Locale code (e.g., `"en-US"`, `"fr-FR"`). Empty string uses the default locale.

## Querying Content

The `Query` resource provides filtered, sorted, paginated content queries by datatype name. This is the primary way to build collection pages (blog listing, product catalog, etc.):

```go
result, err := client.Query.Query(ctx, "blog-posts", &modula.QueryParams{
    Sort:   "-date_created",
    Limit:  10,
    Offset: 0,
    Status: "published",
    Locale: "en-US",
    Filters: map[string]string{
        "category": "tutorials",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Showing %d of %d %s items\n",
    len(result.Data), result.Total, result.Datatype.Label)

for _, item := range result.Data {
    fmt.Printf("  %s: %s\n", item.ContentDataID, item.Fields["title"])
}
```

### QueryParams Fields

| Field | Type | Description |
|-------|------|-------------|
| `Sort` | `string` | Sort field. Prefix with `-` for descending (e.g., `"-date_created"`). |
| `Limit` | `int` | Maximum items to return. Server default when 0. |
| `Offset` | `int` | Items to skip for pagination. |
| `Locale` | `string` | Locale code filter. Empty uses default locale. |
| `Status` | `string` | Filter by content status (`"published"`, `"draft"`). |
| `Filters` | `map[string]string` | Field-level filters. Keys are field names, values are matched exactly. |

Use `Total` from the response for building pagination controls. The response also includes `Datatype` metadata (name and label) for the queried type.

## Content Healing

Detect and repair structural inconsistencies in content trees:

```go
// Dry run first to see what would change.
report, err := client.ContentHeal.Heal(ctx, true)
fmt.Printf("Scanned %d nodes, found %d repairs needed\n",
    report.ContentDataScanned, len(report.ContentDataRepairs))

// Apply repairs.
report, err = client.ContentHeal.Heal(ctx, false)
```

## Next Steps

- [Error Handling](/docs/sdks/go/error-handling) -- error classification and recovery patterns
- [Pagination](/docs/sdks/go/pagination) -- iteration patterns for large datasets
- [Reference](/docs/sdks/go/reference) -- full resource index
