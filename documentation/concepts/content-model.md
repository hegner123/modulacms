# Content Model

ModulaCMS uses a schema-first content model. You define **datatypes** (content schemas) and **fields** (properties within those schemas), then create **content data** nodes that conform to those schemas. Content data nodes live in trees and store their values in **content fields**. A parallel admin content system provides CMS-internal configuration with the same structure.

## Datatypes

A datatype is a content schema definition -- analogous to a "post type" in WordPress or a "content type" in Strapi. Datatypes define what kind of content a node represents and which fields it contains.

```go
type Datatype struct {
    DatatypeID   DatatypeID  `json:"datatype_id"`
    ParentID     *DatatypeID `json:"parent_id"`
    Name         string      `json:"name"`
    Label        string      `json:"label"`
    Type         string      `json:"type"`
    AuthorID     *UserID     `json:"author_id"`
    DateCreated  Timestamp   `json:"date_created"`
    DateModified Timestamp   `json:"date_modified"`
}
```

| Field | Purpose |
|-------|---------|
| `DatatypeID` | ULID identifier, generated server-side |
| `ParentID` | Optional parent datatype for hierarchical organization |
| `Name` | Machine-readable identifier (e.g., `hero_section`) |
| `Label` | Human-readable display name (e.g., `Hero Section`) |
| `Type` | Categorization string that controls behavior |

### Datatype Types

The `Type` field categorizes the datatype. Types prefixed with underscore (`_`) are engine-reserved and trigger built-in behavior. All other values are user-defined pass-through strings.

**Reserved types** (engine-enforced):

| Type | Purpose |
|------|---------|
| `_root` | Tree entry point. Every route's content tree must have exactly one `_root` node. Its `parent_id` is always null. |
| `_reference` | Triggers tree composition. Resolves `_id` field values and attaches the referenced content trees as children. |
| `_nested_root` | Assigned at runtime during tree composition. When a `_reference` node's subtree is fetched, the fetcher replaces the subtree root's original type with `_nested_root` so the tree builder's root-finding logic (`IsRootType`) works recursively without modification. The `_nested_root` type persists in the delivered JSON output. |
| `_system_log` | Synthetic node injected when a reference cannot be resolved during composition. Contains error details. |
| `_collection` | Marks content as a queryable collection. Signals to clients that children support filtering and pagination via the query API. |
| `_global` | Singleton site-wide content (menus, footers, settings). Not tied to a route — delivered via the `/globals` endpoint. |
| `_plugin` | Plugin-provided content. Actual types use the `_plugin_{name}` namespace (e.g., `_plugin_analytics`), registered by the plugin system during initialization. |

**User-defined types** (pass-through, no engine behavior):

Any string not starting with `_` is a valid user-defined type. Use descriptive strings to categorize your component datatypes:

| Example Type | Use Case |
|-------------|----------|
| `section` | Layout sections (Hero, Footer, Sidebar) |
| `container` | Grouping containers (Cards, Tabs, Accordion) |
| `card` | Individual card components |
| `element` | Atomic UI elements |

Attempting to create a datatype with an unrecognized `_`-prefixed type returns an error.

## Fields

A field defines a single property within a datatype schema. Fields specify the data type, validation rules, and UI rendering hints for content entry.

```go
type Field struct {
    FieldID      FieldID     `json:"field_id"`
    ParentID     *DatatypeID `json:"parent_id"`
    SortOrder    int64       `json:"sort_order"`
    Name         string      `json:"name"`
    Label        string      `json:"label"`
    Data         string      `json:"data"`
    Validation   string      `json:"validation"`
    UIConfig     string      `json:"ui_config"`
    Type         FieldType   `json:"type"`
    Translatable int64       `json:"translatable"`
    Roles        []string    `json:"roles"`
    AuthorID     *UserID     `json:"author_id"`
    DateCreated  Timestamp   `json:"date_created"`
    DateModified Timestamp   `json:"date_modified"`
}
```

| Field | Purpose |
|-------|---------|
| `ParentID` | The datatype this field belongs to |
| `SortOrder` | Display ordering within the datatype's field list |
| `Name` | Machine-readable name (e.g., `featured_image`) |
| `Label` | Human-readable name (e.g., `Featured Image`) |
| `Data` | JSON blob with type-specific configuration (e.g., select options, reference target) |
| `Validation` | JSON blob with validation rules (e.g., required, min/max length) |
| `UIConfig` | JSON blob with admin UI rendering hints (e.g., widget type, placeholder text) |
| `Type` | The field type -- determines storage, validation, and UI widget |
| `Translatable` | Non-zero if the field supports per-locale values for i18n |
| `Roles` | Restrict visibility to specific roles. `nil` means unrestricted. |

### Field Types

The `FieldType` enum defines all available field types:

| Type | Description |
|------|-------------|
| `text` | Single-line plain text input |
| `textarea` | Multi-line plain text input |
| `number` | Numeric value (integer or decimal) |
| `date` | Calendar date without time (YYYY-MM-DD) |
| `datetime` | Date with time, serialized as RFC 3339 UTC |
| `boolean` | True/false value |
| `select` | Value from a predefined list of options (configured in the `Data` JSON blob) |
| `media` | Reference to a media asset (stores a `MediaID`) |
| `_id` | Content data ID (ULID). On `_reference` datatype nodes, the composition engine resolves this value to fetch and attach referenced subtrees at delivery time. |
| `json` | Arbitrary JSON data, preserved as-is without schema validation |
| `richtext` | Formatted text with markup (HTML or structured rich text) |
| `slug` | URL-safe slug, validated to contain only lowercase letters, numbers, and hyphens |
| `email` | Email address with format validation |
| `url` | URL string with format validation |

Custom field types can be registered via the FieldType API to extend the CMS.

## Datatype-Field Junction

Fields are linked to datatypes through a junction table, allowing fields to be shared across multiple datatypes.

```go
type DatatypeField struct {
    ID           DatatypeFieldID `json:"id"`
    DatatypeID   DatatypeID      `json:"datatype_id"`
    FieldID      FieldID         `json:"field_id"`
    SortOrder    int64           `json:"sort_order"`
    DateCreated  Timestamp       `json:"date_created"`
    DateModified Timestamp       `json:"date_modified"`
}
```

`SortOrder` on the junction record controls the display ordering of fields within a specific datatype. This is separate from the field's own `SortOrder`, which is its default ordering.

## Content Data

Content data represents the actual content -- a node in a [content tree](tree-structure.md). Each content data node has a datatype (its schema), an optional route (its URL), and tree pointers for navigation.

```go
type ContentData struct {
    ContentDataID ContentID     `json:"content_data_id"`
    ParentID      *ContentID    `json:"parent_id"`
    FirstChildID  *string       `json:"first_child_id"`
    NextSiblingID *string       `json:"next_sibling_id"`
    PrevSiblingID *string       `json:"prev_sibling_id"`
    RouteID       *RouteID      `json:"route_id"`
    DatatypeID    *DatatypeID   `json:"datatype_id"`
    AuthorID      *UserID       `json:"author_id"`
    Status        ContentStatus `json:"status"`
    PublishedAt   *Timestamp    `json:"published_at,omitempty"`
    PublishedBy   *UserID       `json:"published_by,omitempty"`
    PublishAt     *Timestamp    `json:"publish_at,omitempty"`
    Revision      int64         `json:"revision"`
    DateCreated   Timestamp     `json:"date_created"`
    DateModified  Timestamp     `json:"date_modified"`
}
```

The four tree pointers (`ParentID`, `FirstChildID`, `NextSiblingID`, `PrevSiblingID`) form a doubly-linked sibling list. See [Tree Structure](tree-structure.md) for details.

`Status` is either `draft` or `published`. See [Publishing Lifecycle](publishing-lifecycle.md) for the full workflow.

`Revision` is an incrementing counter that tracks how many times the content node has been modified.

## Content Fields

Content fields store the actual values for a content data node. Each content field links to a field definition and a content data node.

```go
type ContentField struct {
    ContentFieldID ContentFieldID `json:"content_field_id"`
    RouteID        *RouteID       `json:"route_id"`
    ContentDataID  *ContentID     `json:"content_data_id"`
    FieldID        *FieldID       `json:"field_id"`
    FieldValue     string         `json:"field_value"`
    Locale         string         `json:"locale"`
    AuthorID       *UserID        `json:"author_id"`
    DateCreated    Timestamp      `json:"date_created"`
    DateModified   Timestamp      `json:"date_modified"`
}
```

All values are stored as strings in `FieldValue`, regardless of the field type. Numbers are stored as their string representation, booleans as `"true"` or `"false"`, media references as the MediaID string, and JSON as a serialized JSON string.

`Locale` identifies the language/region for this field value. See [Localization](localization.md) for how translatable fields work.

`RouteID` is denormalized on content field records for query performance. Always include the route ID when creating content fields.

## Content Relations

Content relations represent directional references between content nodes through `_id`-type fields.

```go
type ContentRelation struct {
    ContentRelationID ContentRelationID `json:"content_relation_id"`
    SourceContentID   ContentID         `json:"source_content_id"`
    TargetContentID   ContentID         `json:"target_content_id"`
    FieldID           FieldID           `json:"field_id"`
    SortOrder         int64             `json:"sort_order"`
    DateCreated       Timestamp         `json:"date_created"`
}
```

Relations are created when a content node's `_id`-type field references another content node. `SortOrder` controls the display ordering when multiple relations exist on the same field.

At content delivery time, `_id` fields can compose the referenced node's subtree inline. The composition depth limit is configurable via `composition_max_depth` in `config.json` (default: 10).

## Admin Content System

ModulaCMS maintains a parallel admin content system for CMS-internal configuration. Admin content uses a separate set of tables with identical structure:

| User Content | Admin Content |
|-------------|---------------|
| `Datatype` | `AdminDatatype` |
| `Field` | `AdminField` |
| `DatatypeField` | `AdminDatatypeField` |
| `ContentData` | `AdminContentData` |
| `ContentField` | `AdminContentField` |
| `ContentRelation` | `AdminContentRelation` |
| `Route` | `AdminRoute` |
| `FieldTypeInfo` | `AdminFieldTypeInfo` |
| `ContentVersion` | `AdminContentVersion` |

The admin content system allows the CMS to manage its own internal content (dashboard configuration, system pages, navigation menus) using the same tree-based structure as user content, without namespace collisions with user-defined schemas.

Admin content is managed through separate API endpoints under `/api/v1/admin/` and has its own publishing lifecycle.

## Entity Relationships

```
Datatype (schema)
  |
  +--< DatatypeField >-- Field (property definition)
  |
  +--< ContentData (tree node)
         |
         +--< ContentField (value, keyed by FieldID + Locale)
         |
         +--< ContentRelation (references to other ContentData)
         |
         +-- Route (URL mapping)
```

A datatype defines the schema. Fields define the properties. DatatypeField links them together. Content data nodes use a datatype as their schema and store values in content fields. Content relations connect nodes through `_id`-type fields. Routes give content data nodes addressable URLs.

## ID System

All entities use 26-character ULID identifiers wrapped in distinct Go types for compile-time safety. A `ContentID` cannot be passed where a `UserID` is expected. ULIDs encode a millisecond-precision timestamp, making them naturally sortable by creation time.

See the [Glossary](../reference/glossary.md) for a complete list of ID types.
