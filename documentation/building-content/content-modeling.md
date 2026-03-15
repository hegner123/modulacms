# Content Modeling

Define your content schema by creating datatypes and fields, then populate them with structured data that ModulaCMS delivers to your frontend.

## Datatypes

A **datatype** is a content type definition. It describes a kind of content -- a "Blog Post," a "Hero Section," a "Product Card." Think of it as a "post type" in WordPress or a "content type" in Contentful.

Every datatype has a human-readable `label` (shown in the admin panel and TUI) and a `type` that categorizes its role.

### Datatype types

The `type` field controls how ModulaCMS treats the datatype. Types starting with `_` are reserved and trigger built-in behavior. All other values are user-defined.

**Reserved types:**

| Type | Purpose |
|------|---------|
| `_root` | Tree entry point for route-based content. Every route's content tree starts with one `_root` node. |
| `_reference` | Embeds shared content from another tree. ModulaCMS resolves the referenced content and attaches it at delivery time. |
| `_collection` | Marks content as a queryable collection. Clients can filter and paginate children via the query API. |
| `_global` | Tree entry point for site-wide content (menus, footers, settings). Not tied to a route -- accessed via the `/globals` endpoint. |
| `_plugin` | Plugin-provided content. Uses the `_plugin_{name}` namespace (e.g., `_plugin_analytics`). |

ModulaCMS rejects datatype creation if you use an unrecognized `_`-prefixed type.

**User-defined types:**

Any string not starting with `_` is a valid user-defined type. Use descriptive strings to organize your datatypes:

| Example Type | Use Case |
|-------------|----------|
| `section` | Layout sections (Hero, Footer, Sidebar) |
| `container` | Grouping containers (Cards, Tabs, Accordion) |
| `card` | Individual card components |
| `element` | Atomic UI elements |

User-defined types are pass-through -- ModulaCMS stores them but doesn't assign them special behavior. They help you categorize datatypes in the admin interface.

### Create a datatype

Create a datatype by sending a POST request to `/api/v1/datatype`:

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

Note the `datatype_id` from the response -- you'll use it when creating fields.

### Constrain datatypes by parent

A datatype can reference a parent datatype. This enforces a structural rule: instances of the child datatype can only appear under instances of the parent datatype in the content tree.

For example, a "Featured Card" datatype with `parent_id` pointing to a "Cards Container" datatype can only be nested under a "Cards Container" in the tree.

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

## Fields

A **field** is a single property within a datatype. Fields define what data an editor fills in when creating content -- a "Title" of type `text`, a "Body" of type `richtext`, a "Featured Image" of type `media`.

### Create a field

Create a field and assign it to a datatype using the `parent_id`:

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

The `parent_id` is the datatype this field belongs to. The `data` property accepts a JSON string containing metadata such as validation rules, placeholder text, and help text.

### Field types

Each field type determines the editor component shown in the admin panel and TUI, and tells consumers how to interpret the stored value.

| Type | What it does | Stored value |
|------|-------------|--------------|
| `text` | Single-line text input | Plain text |
| `textarea` | Multi-line plain text input | Plain text |
| `richtext` | Rich text / HTML editor | HTML string |
| `number` | Numeric input | Number as string |
| `date` | Date picker | ISO 8601 date (YYYY-MM-DD) |
| `datetime` | Date and time picker | ISO 8601 datetime (RFC 3339 UTC) |
| `boolean` | True/false toggle | `"true"` or `"false"` |
| `select` | Dropdown from predefined options (configured in `data`) | Selected option value |
| `media` | Media asset picker | Media ID |
| `_id` | Content node picker. On `_reference` datatypes, ModulaCMS resolves this value and attaches the referenced content at delivery time. | Content data ID |
| `json` | Structured JSON editor | JSON string |
| `slug` | URL-safe slug input (lowercase, numbers, hyphens) | Slug string |
| `email` | Email input with format validation | Email address |
| `url` | URL input with format validation | URL string |
| `plugin` | Plugin-provided editor with custom input UI | Opaque string (plugin decides format) |

> **Good to know**: All field values are stored as strings regardless of type. Numbers become their string representation, booleans become `"true"` or `"false"`, and references become ID strings. The field type tells your frontend how to interpret the value.

### Field properties

| Property | Purpose |
|----------|---------|
| `label` | Human-readable name shown in the editor UI |
| `type` | Determines the editor component and value format |
| `data` | JSON string for type-specific metadata (placeholder text, select options, help text) |
| `validation` | JSON string for validation rules (required, min/max length) |
| `ui_config` | JSON string controlling how the field renders in the admin interface |
| `translatable` | Whether the field stores per-locale values for localization |
| `roles` | Restrict visibility to specific user roles. Unrestricted by default. |

### Register custom field types

You can add field types beyond the built-in set. Register a custom type, and it becomes available for any field:

```bash
curl -X POST http://localhost:8080/api/v1/fieldtypes \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "color",
    "label": "Color Picker"
  }'
```

When a field uses a custom type without a `ui_config`, the admin panel falls back to a plain text input. Set `ui_config` to specify a specialized editor component.

## View a datatype with its fields

The composed view endpoint returns a datatype with all its field definitions in a single response:

```bash
curl "http://localhost:8080/api/v1/datatype/full?q=01JNRW5V6QNPZ3R8W4T2YH9B0D" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Use this when building content editors that need to render forms based on the datatype schema.

## Example: build a blog post schema

Here is the full sequence for creating a "Blog Post" datatype with five fields.

**1. Create the datatype:**

```bash
curl -X POST http://localhost:8080/api/v1/datatype \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"label": "Blog Post", "type": "_root"}'
```

**2. Create the fields:**

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

## Modify schemas at runtime

Add, remove, or change fields on an existing datatype at any time. No restart or migration required.

**Add a field** -- Create a new field with `parent_id` set to the target datatype. The field appears immediately in the content editor. Existing content doesn't automatically have values for the new field -- values get populated when editors update the content.

**Remove a field** -- Delete the field definition. ModulaCMS cascades the deletion to all content field values for that field across all content instances.

**Delete a datatype** -- Deletion is blocked if content instances reference the datatype. Delete or reassign the content first.

> **Good to know**: Deleting a field permanently removes all stored values for that field. Deleting a datatype cascades to its field definitions.

## Shared content with references

Content trees can include content from other trees by reference. This is how shared content (navigation menus, footers, sidebars) gets embedded into multiple pages without duplication.

1. Create a datatype with `type = "_reference"` (e.g., "Menu Reference").
2. Add an `_id` field to it that points to the content you want to embed.
3. Place an instance of the reference datatype in your content tree.
4. When a frontend client requests the page, ModulaCMS detects the reference, fetches the referenced content, and attaches it as children of the reference node.

The referenced content lives in one place. Update it once, and every page that references it reflects the change.

> **Good to know**: Configure the maximum reference depth via `composition_max_depth` in `modula.config.json` (default: 10). If a reference can't be resolved, ModulaCMS returns the rest of the tree normally and includes error details in the response.

## API reference

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/datatype` | `datatypes:read` | List all datatypes |
| POST | `/api/v1/datatype` | `datatypes:create` | Create a datatype |
| GET | `/api/v1/datatype/` | `datatypes:read` | Get a datatype (`?q=ID`) |
| GET | `/api/v1/datatype/full` | `datatypes:read` | Get a datatype with all fields (`?q=ID`) |
| PUT | `/api/v1/datatype/` | `datatypes:update` | Update a datatype |
| DELETE | `/api/v1/datatype/` | `datatypes:delete` | Delete a datatype (`?q=ID`) |
| GET | `/api/v1/fields` | `fields:read` | List all fields |
| POST | `/api/v1/fields` | `fields:create` | Create a field |
| GET | `/api/v1/fields/` | `fields:read` | Get a field (`?q=ID`) |
| PUT | `/api/v1/fields/` | `fields:update` | Update a field |
| DELETE | `/api/v1/fields/` | `fields:delete` | Delete a field (`?q=ID`) |
| GET | `/api/v1/fieldtypes` | `fields:read` | List registered field types |
| POST | `/api/v1/fieldtypes` | `fields:create` | Register a custom field type |

All list endpoints support pagination with `limit` and `offset` query parameters.

> **Good to know**: All IDs are 26-character ULIDs -- time-sortable and globally unique.

## Next steps

Now that you have a content schema, [create content and organize it into trees](creating-content.md).
