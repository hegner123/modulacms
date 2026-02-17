# Response Assembly Pattern

This document describes the three-tier response assembly pattern used in ModulaCMS for composing API responses with varying levels of nesting.

## Three Assembly Tiers

| Tier | Use Case | Queries | Example |
|------|----------|---------|---------|
| **Flat** | Single-entity CRUD | 1 query | `GET /api/v1/content/data?q={id}` |
| **Composed** | Entity with embedded relations | 2-4 bounded queries (detail) or JOINs (list) | Future composed endpoints |
| **Tree** | Recursive hierarchy with sibling ordering | JOINs + BuildNodes + transform | `GET /api/v1/admin/tree/{slug}` |

### When to Use Each Tier

- **Flat**: Default for all CRUD endpoints. Returns the entity as stored, with ID references for related objects. No additional queries needed.
- **Composed**: When the response needs embedded related objects (e.g., `author: {}`, `datatype: {}`, `fields: [...]`). Use when clients need the full picture in a single request without tree traversal.
- **Tree**: Only for recursive parent-child hierarchies with sibling pointer ordering. Uses the BuildNodes pipeline in `internal/transform/`.

## View Types

View types represent the composed response shape. They live in `internal/db/views.go`.

### Naming Convention

- Suffix: `*View` (e.g., `ContentDataView`, `AuthorView`)
- Nullable one-to-one relations: pointer type with `omitempty` (nil when FK is null, omitted from JSON)
- One-to-many relations: slice type, always initialized to empty slice (JSON `[]` not `null`)
- Sensitive fields excluded (e.g., `AuthorView` omits `hash`)

### Existing View Types

| Type | Purpose | Source Entity |
|------|---------|---------------|
| `AuthorView` | Safe user subset (no hash) | `Users` |
| `DatatypeView` | Embedded datatype summary | `Datatypes` |
| `FieldView` | Field definition paired with value | `ContentFields` + `Fields` |
| `ContentDataView` | Full composed content response | `ContentData` + relations |

## Mapper Functions

Mappers convert entity types to view types. They are standalone functions in `internal/db/views.go`.

| Function | Converts |
|----------|----------|
| `MapAuthorView(Users)` | Users -> AuthorView |
| `MapDatatypeView(Datatypes)` | Datatypes -> DatatypeView |
| `MapFieldView(ContentFields, Fields)` | ContentFields + Fields -> FieldView |
| `MapFieldViewFromRow(ContentFieldWithFieldRow)` | JOIN row -> FieldView |

## Assembly Functions

Assembly functions compose view types from raw entities. They live in `internal/db/assemble.go` as standalone functions (not methods on DbDriver -- the logic is identical across all 3 database backends).

### Detail Assembly (single item)

```go
view, err := AssembleContentDataView(driver, contentID)
```

Uses bounded sequential queries:
1. `GetContentData` -- main entity
2. `GetUser` -- author (if author_id is set)
3. `GetDatatype` -- datatype (if datatype_id is set)
4. `ListContentFieldsWithFieldByContentData` -- fields with definitions (single JOIN query)

### List Assembly (multiple items)

For list endpoints, avoid N+1 by using:
1. A single query to fetch all items
2. `GroupBy` to group JOIN rows by primary entity ID
3. `AssembleFieldViews` to convert grouped rows to view slices

```go
grouped := GroupBy(rows, func(r ContentFieldWithFieldRow) types.ContentID {
    return r.ContentDataID.ID
})
for id, fieldRows := range grouped {
    fields := AssembleFieldViews(fieldRows)
    // ... compose view
}
```

## Adding a New View Type

1. **Define the view struct** in `internal/db/views.go` with `*View` suffix
2. **Add a mapper function** (e.g., `MapFooView(Foo) FooView`)
3. **If a JOIN query is needed**: add SQL to `sql/schema/22_joins/` (all 3 dialects), run `just sqlc`, add wrapper type to `internal/db/getTree.go`, add to `DbDriver` interface, implement on all 3 wrappers
4. **Add an assembly function** in `internal/db/assemble.go`
5. **Wire into handler** in `internal/router/`
6. **Add tests** in `internal/db/assemble_test.go`

## Key Files

| File | Purpose |
|------|---------|
| `internal/db/views.go` | View types and mapper functions |
| `internal/db/assemble.go` | Assembly functions and GroupBy helper |
| `internal/db/getTree.go` | JOIN wrapper types (ContentFieldWithFieldRow, etc.) |
| `internal/db/db.go` | DbDriver interface (ListContentFieldsWithFieldByContentData) |
| `sql/schema/22_joins/` | JOIN queries for all 3 SQL dialects |
