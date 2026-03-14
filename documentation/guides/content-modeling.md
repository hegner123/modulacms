# Content Modeling

ModulaCMS uses a dynamic schema system where content structure is defined at runtime, not in application code. You create content types (called datatypes), define their properties (fields), and assign fields to datatypes -- all through the API or admin panel. No code deployment, no database migrations, no restart required.

This guide covers how the schema system works and how to use it to model your content.

## Concepts

**Datatype** -- A content type definition that describes a kind of content. A "Blog Post" datatype, a "Page" datatype, a "Hero Section" datatype. Analogous to a "post type" in WordPress or a "content type" in Contentful. Datatypes are stored in the database and can be created or modified at any time.

**Field** -- A property definition that belongs to a datatype. A "Title" field of type `text`, a "Body" field of type `richtext`, a "Featured Image" field of type `media`. Each field has a `parent_id` that points to the datatype it belongs to.

**Content data** -- An instance of a datatype. When you create a blog post, you create a content data record that references the "Blog Post" datatype. Content data records form a tree structure (covered in [content-trees.md](content-trees.md)).

**Content field** -- The actual value for a field on a specific content data instance. If the field definition says "Title, type text," the content field holds the actual string "My First Post."

**Field type** -- The kind of input a field represents. Field types are stored in the `field_types` table and can be extended at runtime -- you are not limited to the built-in set. Custom field types can be registered and used alongside the defaults.

## How the Pieces Fit Together

The content model has two layers: **definitions** and **instances**.

Definitions (the schema):
- Datatypes define what kinds of content exist
- Fields define what properties those content types have, each assigned to a single datatype via `parent_id`

Instances (the content):
- Content data records are instances of datatypes
- Content field records hold the values for each field on each content data instance

```
datatypes
    |
    +--- fields (via fields.parent_id -> datatypes.datatype_id)
    |
    v
content_data (instances of datatypes)
    |
    v
content_fields (values for fields on instances)
```

Each content data record points to its datatype via `datatype_id`. Each content field record points to both its content data instance (`content_data_id`) and its field definition (`field_id`).

## Datatype Classifications

Every datatype has a `type` field that classifies its role. Types prefixed with underscore (`_`) are engine-reserved and trigger built-in behavior. All other values are user-defined pass-through strings.

### Reserved Types

| Type | Purpose |
|------|---------|
| `_root` | Tree entry point. Every route's content tree must have exactly one root node whose datatype is `_root`. |
| `_reference` | Triggers tree composition. Resolves `_id` field values and attaches the referenced content trees as children. See [Tree Composition](#tree-composition). |
| `_nested_root` | Assigned at runtime during tree composition. When a `_reference` node's subtree is fetched, the fetcher replaces the subtree root's original type with `_nested_root` so the tree builder's root-finding logic (`IsRootType`) works recursively without modification. The `_nested_root` type persists in the delivered JSON output. |
| `_system_log` | Synthetic node injected when a reference cannot be resolved. Contains error details. |
| `_collection` | Marks content as a queryable collection. Signals to clients that children support filtering and pagination via the query API. |
| `_global` | Singleton site-wide content (menus, footers, settings). Not tied to a route — delivered via the `/globals` endpoint. |
| `_plugin` | Plugin-provided content. Actual types use the `_plugin_{name}` namespace (e.g., `_plugin_analytics`), registered by the plugin system during initialization. |

### User-Defined Types

Any string that does not start with `_` is a valid user-defined type. Use descriptive strings to categorize your component datatypes:

| Example Type | Use Case |
|-------------|----------|
| `"section"` | Layout sections (Hero, Footer, Sidebar) |
| `"container"` | Grouping containers (Cards, Tabs, Accordion) |
| `"card"` | Individual card components |
| `"element"` | Atomic UI elements |

User types are pass-through -- the engine stores them but does not assign them special behavior. They are useful for organizing datatypes in the admin interface.

### Parent-Constrained Datatypes

A datatype can have a `parent_id` pointing to another datatype. This enforces a structural rule: instances of the child datatype can only appear as children of instances of the parent datatype in the content tree.

For example, a "Featured Card" datatype with `parent_id` pointing to a "Cards" datatype can only be nested under a "Cards" instance in the content tree.

## Creating Datatypes

Create a datatype by posting to `/api/v1/datatype`:

```bash
curl -X POST http://localhost:8080/api/v1/datatype \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "label": "Blog Post",
    "type": "_root"
  }'
```

Response (HTTP 201):

```json
{
  "datatype_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
  "parent_id": null,
  "label": "Blog Post",
  "type": "_root",
  "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "date_created": "2026-02-27T10:00:00Z",
  "date_modified": "2026-02-27T10:00:00Z"
}
```

To create a child-constrained datatype, include `parent_id`:

```bash
curl -X POST http://localhost:8080/api/v1/datatype \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "label": "Featured Card",
    "type": "card",
    "parent_id": "01JNRW6A7BMXY4K9P2Q5TH3JCR"
  }'
```

## Creating Fields

Create a field definition by posting to `/api/v1/fields`. The `parent_id` assigns the field to a datatype:

```bash
curl -X POST http://localhost:8080/api/v1/fields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
    "label": "Post Title",
    "type": "text",
    "data": "{\"required\": true, \"maxLength\": 200, \"placeholder\": \"Enter post title\"}",
    "validation": "{}",
    "ui_config": "{}"
  }'
```

The `parent_id` points to the datatype this field belongs to. The `data` field accepts a JSON string containing metadata such as validation rules, placeholder text, and help text. The `type` field can be any type registered in the `field_types` table -- including the built-in types and any custom types you have added.

### Built-in Field Types

The built-in admin panel and TUI have editor components for the following field types only. These are the default types that ship with ModulaCMS. The `field_types` table exists to support custom admin panels that implement their own editor components for additional field types.

Each field type determines two things: the editor component rendered in the admin UI, and (for some types) backend behavior during content delivery. The stored value is always a string -- the field type tells consumers how to interpret it.

| Type | Editor Component | Stored Value |
|------|-----------------|--------------|
| `text` | Single-line text input | Plain text |
| `textarea` | Multi-line plain text input | Plain text |
| `richtext` | Rich text / HTML editor | HTML string |
| `number` | Numeric input | Number as string |
| `date` | Date picker | ISO 8601 date |
| `datetime` | Date and time picker | ISO 8601 datetime |
| `boolean` | True/false toggle | `"true"` or `"false"` |
| `select` | Dropdown selection | Selected option value |
| `media` | Media asset picker | Media ID (ULID) |
| `_id` | Content node picker | Content data ID (ULID). On `_reference` datatype nodes, the composition engine resolves this value to fetch and attach referenced subtrees at delivery time. See [Tree Composition](#tree-composition). |
| `json` | Structured JSON editor | JSON string |
| `slug` | URL-safe slug input | Slug string |
| `email` | Email input | Email address |
| `url` | URL input | URL string |
| `plugin` | Plugin-provided editor (inline or overlay) | Opaque string (plugin decides format) |

### Custom Field Types

The `field_types` table is an extension point for adding new field types beyond the built-in set. When a field uses a custom type and its `ui_config` column is blank, the admin panel and TUI fall back to a plain text input. The `ui_config` JSON column can specify an editor component to use for the custom type, allowing the built-in admin panel and TUI to render a specialized editor without requiring a custom frontend.

Register a custom field type:

```bash
curl -X POST http://localhost:8080/api/v1/fieldtypes \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "color",
    "label": "Color Picker"
  }'
```

Custom admin panels (built with the SDKs or as standalone frontends) can also use the `field_types` table to discover registered types and render their own editor components. The backend stores the type string and field value as-is -- interpretation and validation for custom types is the responsibility of the rendering layer.

### Field Column Purposes

| Column | Purpose |
|--------|---------|
| `parent_id` | The datatype this field belongs to (references `datatypes.datatype_id`) |
| `label` | Human-readable display name shown in the UI |
| `data` | Generic text column with no strictly defined purpose. Some output format transformers read from it. Available for any application-specific metadata. |
| `validation` | JSON blob containing composable validation rules. Uses a planned rule-based validation system where rules are composed declaratively -- no regex. |
| `ui_config` | JSON blob containing TUI display configuration. Controls how the field renders in the SSH TUI admin interface. |
| `type` | Field type string, referencing a registered type in the `field_types` table |

## Viewing a Datatype with Its Fields

The composed view endpoint returns a datatype with all its field definitions in a single response:

```bash
curl "http://localhost:8080/api/v1/datatype/full?q=01JNRW5V6QNPZ3R8W4T2YH9B0D" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

This is useful when building content editors that need to render forms based on the datatype schema.

## Tree Composition

Tree composition allows a content tree to include content from other content trees by reference. This is how shared content (navigation menus, footers, sidebars) gets embedded into multiple pages without duplication.

### How It Works

1. Create a datatype with `type = "_reference"` (e.g., "Menu Reference").
2. Add one or more `_id` fields to it. Each field holds the `content_data_id` of a content node to reference.
3. Place an instance of the `_reference` datatype in your content tree.
4. When the content delivery endpoint assembles the tree, it detects `_reference` nodes, fetches the referenced content trees, and attaches them as children of the reference node.

### Example: Shared Navigation Menu

Suppose you have a main navigation menu that should appear on every page.

**Step 1: Create the menu content.** Create a `_root` datatype called "Main Menu" with fields for menu items. Create a route for it (e.g., slug `main-menu`) and populate it with content.

```
Route: slug = "main-menu"
└── content_data (_root "Main Menu")
    ├── content_data ("Menu Item")
    │   └── content_field: label = "Home"
    ├── content_data ("Menu Item")
    │   └── content_field: label = "About"
    └── content_data ("Menu Item")
        └── content_field: label = "Contact"
```

**Step 2: Create a reference datatype.** Create a datatype with `type = "_reference"` and label "Menu Reference". Add an `_id` field to it:

```bash
# Create the _reference datatype
curl -X POST http://localhost:8080/api/v1/datatype \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"label": "Menu Reference", "type": "_reference"}'

# Create an _id field on it
curl -X POST http://localhost:8080/api/v1/fields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "MENU_REFERENCE_DATATYPE_ID",
    "label": "menu",
    "type": "_id",
    "data": "{}",
    "validation": "{}",
    "ui_config": "{}"
  }'
```

**Step 3: Use the reference in page content trees.** In each page's content tree, add an instance of the "Menu Reference" datatype. Set its `_id` field value to the `content_data_id` of the Main Menu's root node.

```
Route: slug = "homepage"
└── content_data (_root "Page")
    ├── content_data (_reference "Menu Reference")   <- references Main Menu
    │   └── content_field: menu = "01JNRW..."        <- content_data_id of Main Menu root
    ├── content_data ("Hero Section")
    │   └── content_field: heading = "Welcome"
    └── content_data ("Footer")
```

**Step 4: Content delivery resolves the reference.** When a client requests `GET /api/v1/content/homepage`, the composition engine:

1. Detects the `_reference` node ("Menu Reference").
2. Reads its `_id` field value (`01JNRW...`).
3. Fetches and builds the referenced content tree (the Main Menu).
4. Attaches the menu tree as a child of the reference node.
5. Returns the fully composed tree.

The menu content lives in one place but appears in every page that references it. Update the menu once, and all pages reflect the change.

### Composition Behavior

- **Depth-bounded.** Composition recurses up to `composition_max_depth` levels (default 10, configurable in `modula.config.json`). If a referenced tree itself contains `_reference` nodes, those are also resolved, up to the depth limit.
- **Concurrent.** Sibling references at the same level are resolved concurrently (up to 10 goroutines) for performance.
- **Circular reference detection.** If a reference would create a cycle (A references B which references A), a `_system_log` node is injected instead with an error message explaining the cycle.
- **Graceful degradation.** If a referenced content node does not exist or cannot be fetched, a `_system_log` node is injected with error details rather than failing the entire request. The rest of the tree is returned normally.

### _system_log Nodes

When composition encounters an error, it injects a synthetic `_system_log` node with four fields:

| Field | Description |
|-------|-------------|
| `error` | Error type: `circular_reference`, `max_depth`, `not_found`, or `build_failed` |
| `message` | Human-readable description of the problem and how to fix it |
| `reference_id` | The `content_data_id` that could not be resolved |
| `parent_label` | The label of the `_reference` datatype that triggered the error |

These nodes are visible in the JSON output and can be used by frontends to display appropriate error states or be filtered out in production.

## Example: Building a Blog Post Schema

Here is the full sequence for creating a "Blog Post" datatype with five fields.

**1. Create the datatype:**

```bash
curl -X POST http://localhost:8080/api/v1/datatype \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"label": "Blog Post", "type": "_root"}'
```

Note the `datatype_id` from the response. For this example, assume `01JNRW5V6QNPZ3R8W4T2YH9B0D`.

**2. Create the fields (each assigned to the datatype via `parent_id`):**

```bash
# Title field
curl -X POST http://localhost:8080/api/v1/fields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
    "label": "Title",
    "type": "text",
    "data": "{\"required\": true, \"maxLength\": 200}",
    "validation": "{}",
    "ui_config": "{}"
  }'

# Body field
curl -X POST http://localhost:8080/api/v1/fields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
    "label": "Body",
    "type": "richtext",
    "data": "{\"required\": true}",
    "validation": "{}",
    "ui_config": "{}"
  }'

# Excerpt field
curl -X POST http://localhost:8080/api/v1/fields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
    "label": "Excerpt",
    "type": "textarea",
    "data": "{\"maxLength\": 300}",
    "validation": "{}",
    "ui_config": "{}"
  }'

# Featured Image field
curl -X POST http://localhost:8080/api/v1/fields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
    "label": "Featured Image",
    "type": "media",
    "data": "{}",
    "validation": "{}",
    "ui_config": "{}"
  }'

# Publish Date field
curl -X POST http://localhost:8080/api/v1/fields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
    "label": "Publish Date",
    "type": "datetime",
    "data": "{\"required\": true}",
    "validation": "{}",
    "ui_config": "{}"
  }'
```

**3. Verify the schema:**

```bash
curl "http://localhost:8080/api/v1/datatype/full?q=01JNRW5V6QNPZ3R8W4T2YH9B0D" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

The response contains the datatype and all five field definitions.

## Creating Content Instances

Once a datatype schema is defined, create content instances with `/api/v1/contentdata`:

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

Then set field values with `/api/v1/contentfields`:

```bash
curl -X POST http://localhost:8080/api/v1/contentfields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "field_id": "01JNRW7K8CNQZ5P3R9W6TJ4MAS",
    "field_value": "My First Blog Post",
    "route_id": "01JNRW9P2DKTZ6Q4M8W3B5J7CL"
  }'
```

Content status values: `draft`, `published`.

## Modifying Schemas at Runtime

You can add fields to an existing datatype, remove fields, or modify them at any time.

**Adding a field** to an existing datatype: create the field with `parent_id` set to the target datatype. The new field appears immediately in the content editor. Existing content instances do not automatically have values for the new field -- they get populated when edited.

**Removing a field**: delete the field definition. This cascades -- all content field values for that field across all content instances are deleted.

**Deleting a datatype** is blocked if content instances reference it. Delete or reassign the content first.

## API Reference

All endpoints require authentication and the appropriate permission.

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/datatype` | `datatypes:read` | List all datatypes |
| POST | `/api/v1/datatype` | `datatypes:create` | Create a datatype |
| GET | `/api/v1/datatype/` | `datatypes:read` | Get a single datatype (`?q=ID`) |
| GET | `/api/v1/datatype/full` | `datatypes:read` | Get a datatype with all fields (`?q=ID`) |
| PUT | `/api/v1/datatype/` | `datatypes:update` | Update a datatype |
| DELETE | `/api/v1/datatype/` | `datatypes:delete` | Delete a datatype (`?q=ID`) |
| GET | `/api/v1/fields` | `fields:read` | List all fields |
| POST | `/api/v1/fields` | `fields:create` | Create a field |
| GET | `/api/v1/fields/` | `fields:read` | Get a single field (`?q=ID`) |
| PUT | `/api/v1/fields/` | `fields:update` | Update a field |
| DELETE | `/api/v1/fields/` | `fields:delete` | Delete a field (`?q=ID`) |
| GET | `/api/v1/contentdata` | `content:read` | List all content data |
| POST | `/api/v1/contentdata` | `content:create` | Create a content data instance |
| GET | `/api/v1/contentdata/` | `content:read` | Get a single content data instance (`?q=ID`) |
| PUT | `/api/v1/contentdata/` | `content:update` | Update a content data instance |
| DELETE | `/api/v1/contentdata/` | `content:delete` | Delete a content data instance (`?q=ID`) |
| GET | `/api/v1/contentfields` | `content:read` | List all content fields |
| POST | `/api/v1/contentfields` | `content:create` | Create a content field value |
| GET | `/api/v1/contentfields/` | `content:read` | Get a single content field (`?q=ID`) |
| PUT | `/api/v1/contentfields/` | `content:update` | Update a content field value |
| DELETE | `/api/v1/contentfields/` | `content:delete` | Delete a content field value (`?q=ID`) |

All list endpoints support pagination with `limit` and `offset` query parameters.

## Notes

- All IDs are 26-character ULIDs, time-sortable and globally unique.
- The `data` column on fields is a generic text column with no enforced structure. Some output format transformers read from it.
- The `validation` column stores composable validation rules (JSON). The `ui_config` column stores TUI display configuration (JSON). Neither is enforced by the backend at write time.
- Deleting a field definition cascades to all content field values for that field.
- Deleting a datatype cascades to its field definitions.
- Content data instances belong to a route. Every content query is scoped by `route_id`. See [routing.md](routing.md) for details.
- Types prefixed with `_` are reserved for engine use. Attempting to create a datatype with an unrecognized `_`-prefixed type returns an error.
