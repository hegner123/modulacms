# ModulaCMS Content Model Reference

This document describes the five core tables that define ModulaCMS's content modeling system: **routes**, **datatypes**, **fields**, **content_data**, and **content_fields**. Use this as your authoritative reference when designing templated CMS structures that users can install.

---

## Table Schemas

### routes

A route is a URL slug that maps to a content tree. Every piece of published content lives behind a route.

```sql
CREATE TABLE routes (
    route_id     TEXT PRIMARY KEY NOT NULL CHECK (length(route_id) = 26),  -- ULID
    slug         TEXT NOT NULL UNIQUE,   -- URL path (e.g. "/about", "/blog/my-post")
    title        TEXT NOT NULL,
    status       INTEGER NOT NULL,       -- publication state
    author_id    TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

### datatypes

A datatype is a **content type definition** — a blueprint that describes a kind of content (Page, Post, Hero Section, Card, etc.). Datatypes form a type-level hierarchy via `parent_id`.

```sql
CREATE TABLE datatypes (
    datatype_id   TEXT PRIMARY KEY NOT NULL CHECK (length(datatype_id) = 26),  -- ULID
    parent_id     TEXT DEFAULT NULL REFERENCES datatypes ON DELETE SET NULL,   -- type-level parent
    label         TEXT NOT NULL,          -- human-readable name (e.g. "Hero Section")
    type          TEXT NOT NULL,          -- "ROOT" or a free-form category string
    author_id     TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

### fields

A field is a **field definition** attached to a datatype. Fields describe the shape of data a datatype expects (title, body, image, etc.).

```sql
CREATE TABLE fields (
    field_id     TEXT PRIMARY KEY NOT NULL CHECK (length(field_id) = 26),  -- ULID
    parent_id    TEXT DEFAULT NULL REFERENCES datatypes ON DELETE SET NULL, -- owning datatype
    label        TEXT DEFAULT 'unlabeled' NOT NULL,  -- human-readable name
    data         TEXT NOT NULL,           -- generic text for frontend use (alias, metadata)
    validation   TEXT NOT NULL,           -- JSON blob: ModulaCMS validation schema
    ui_config    TEXT NOT NULL,           -- JSON blob: ModulaCMS TUI UI config schema
    type         TEXT NOT NULL CHECK (type IN (
        'text', 'textarea', 'number', 'date', 'datetime', 'boolean',
        'select', 'media', 'relation', 'json', 'richtext', 'slug',
        'email', 'url'
    )),
    author_id    TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

### content_data

A content_data row is an **instance** of a datatype — a single node in the content tree. Nodes are linked into a tree using parent pointers and a doubly-linked sibling list for O(1) navigation and reordering.

```sql
CREATE TABLE content_data (
    content_data_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_data_id) = 26),  -- ULID
    parent_id       TEXT REFERENCES content_data ON DELETE SET NULL,       -- parent node
    first_child_id  TEXT REFERENCES content_data ON DELETE SET NULL,       -- leftmost child
    next_sibling_id TEXT REFERENCES content_data ON DELETE SET NULL,       -- next in sibling list
    prev_sibling_id TEXT REFERENCES content_data ON DELETE SET NULL,       -- previous in sibling list
    route_id        TEXT NOT NULL REFERENCES routes ON DELETE CASCADE,     -- owning route
    datatype_id     TEXT NOT NULL REFERENCES datatypes ON DELETE SET NULL, -- definition pointer
    author_id       TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    status          TEXT NOT NULL DEFAULT 'draft',  -- draft | published | archived | pending
    date_created    TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified   TEXT DEFAULT CURRENT_TIMESTAMP
);
```

### content_fields

A content_fields row holds the **value** for a single field on a single content_data node. It is the instance-level data that pairs with a field definition.

```sql
CREATE TABLE content_fields (
    content_field_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_field_id) = 26),  -- ULID
    route_id         TEXT REFERENCES routes ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id  TEXT NOT NULL REFERENCES content_data ON UPDATE CASCADE ON DELETE CASCADE,
    field_id         TEXT NOT NULL REFERENCES fields ON UPDATE CASCADE ON DELETE CASCADE,
    field_value      TEXT NOT NULL,        -- the actual content value
    author_id        TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    date_created     TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified    TEXT DEFAULT CURRENT_TIMESTAMP
);
```

---

## Entity Relationships

```
route (1) ──────────< content_data (many)
                           │
                           ├── parent_id ────────> content_data (self-referential tree)
                           ├── first_child_id ───> content_data (sibling list head)
                           ├── next_sibling_id ──> content_data (sibling list forward)
                           ├── prev_sibling_id ──> content_data (sibling list backward)
                           │
                           ├── datatype_id ──────> datatype (1) ──< fields (many)
                           │                         │
                           │                         └── parent_id ──> datatype (type-level parent)
                           │
                           └───────< content_fields (many)
                                          │
                                          └── field_id ──> field (definition pointer)
```

**Definition vs Instance:**
- `datatypes` + `fields` = **definitions** (the blueprint/schema)
- `content_data` + `content_fields` = **instances** (the actual content)
- `content_data.datatype_id` points to its definition
- `content_fields.field_id` points to its definition
- `content_fields.content_data_id` points to its owning node

---

## Rules and Constraints

### 1. ROOT Datatypes

A datatype with `type = "ROOT"` is the **only** type that can serve as the root node of a content tree. The tree-building algorithm identifies the root by checking `Datatype.Info.Type == "ROOT"`. Every route's content tree must have exactly one content_data node whose datatype has `type = "ROOT"`.

ROOT datatypes represent first-party page-level content types: **Page**, **Post**, **Case Study**, **Class**, **Service**, etc.

### 2. Parent-Constrained Datatypes

When a datatype has a non-null `parent_id`, it can **only** be used as a child of a content_data node whose datatype matches that parent. This enforces structural rules at the type level.

Example: A "Featured Card" datatype with `parent_id` pointing to the "Cards" datatype can only appear nested under a "Cards" instance in the content tree.

### 3. Unparented Non-ROOT Datatypes

A datatype can have `parent_id = NULL` and `type != "ROOT"`. This is a currently undefined state reserved for future use. When designing templates, avoid creating datatypes in this state unless you have a specific reason and document the intent.

### 4. Fields Belong to Datatypes

Fields are assigned to a specific datatype via `fields.parent_id → datatypes.datatype_id`. Fields are **not designed to be reusable** across datatypes — each datatype owns its own set of field definitions.

A datatype can have **any number of fields** (zero or more).

### 5. Field Column Purposes

| Column | Purpose |
|--------|---------|
| `label` | Human-readable display name shown in the UI |
| `data` | Generic text column for frontend use — could serve as an alias, CSS class hint, or any frontend-specific metadata |
| `validation` | JSON blob containing validation instructions following the ModulaCMS validation schema |
| `ui_config` | JSON blob containing TUI display configuration following the ModulaCMS UI config schema |
| `type` | Enumerated field type for type safety. One of: `text`, `textarea`, `number`, `date`, `datetime`, `boolean`, `select`, `media`, `relation`, `json`, `richtext`, `slug`, `email`, `url` |

### 6. Route-to-Content Binding

A route is linked to an instance of a ROOT datatype via a content_data row where `content_data.route_id = routes.route_id` and the content_data node's datatype has `type = "ROOT"`.

### 7. Content Delivery Builds a Tree

When a client accesses a route through the content delivery endpoint `GET /{slug}`, the system:

1. Resolves the slug to a `route_id`
2. Fetches all `content_data` rows for that route
3. Fetches the `datatype` definition for each content_data node
4. Fetches all `content_fields` for that route
5. Fetches the `field` definition for each content_field
6. Assembles a tree: the ROOT-typed node becomes the root, all other nodes are linked via `parent_id`, children are ordered via the `first_child_id → next_sibling_id` pointer chain
7. Fields are attached to their owning node via matching `content_data_id`

### 8. Tree Navigation is O(1)

The content_data table uses sibling pointers (`parent_id`, `first_child_id`, `next_sibling_id`, `prev_sibling_id`) for O(1) tree navigation and reordering. The tree builder follows the `first_child_id → next_sibling_id` chain to produce correctly ordered child lists, with cycle detection.

### 9. Content Fields Hold Values, Fields Hold Keys

The `fields` table defines **what** a field is (name, type, validation). The `content_fields` table holds the **value** for that field on a specific content_data instance. Think of `fields` as the column definition and `content_fields` as the cell value.

### 10. Planned: Content Reference Fields

It is planned that a `content_fields.field_value` can hold the ID of a `content_data` node. When the tree builder encounters such a reference, it will build that referenced node's subtree and inline its JSON into the field value. This is planned for features like menus and cross-page content embedding.

---

## Tree Structure Example

A "Page" route at `/about` with a Hero section and a Cards section containing two cards:

```
Route: /about
└── content_data (ROOT "Page")          ← datatype.type = "ROOT"
    ├── content_field: title = "About Us"
    ├── content_field: meta_description = "Learn about our company"
    │
    ├── content_data ("Hero Section")    ← child node, ordered via sibling pointers
    │   ├── content_field: heading = "Welcome"
    │   ├── content_field: subheading = "We build things"
    │   └── content_field: background_image = "hero.jpg"
    │
    └── content_data ("Cards")           ← next sibling of Hero Section
        ├── content_data ("Featured Card")  ← parent-constrained to Cards
        │   ├── content_field: title = "Service 1"
        │   └── content_field: description = "We do this"
        │
        └── content_data ("Featured Card")  ← next sibling of first card
            ├── content_field: title = "Service 2"
            └── content_field: description = "We do that"
```

### JSON Output (simplified)

```json
{
  "root": {
    "datatype": {
      "info": { "label": "Page", "type": "ROOT" },
      "content": { "content_data_id": "...", "status": "published" }
    },
    "fields": [
      { "info": { "label": "title", "type": "text" }, "content": { "field_value": "About Us" } },
      { "info": { "label": "meta_description", "type": "textarea" }, "content": { "field_value": "Learn about our company" } }
    ],
    "nodes": [
      {
        "datatype": { "info": { "label": "Hero Section", "type": "section" } },
        "fields": [
          { "info": { "label": "heading", "type": "text" }, "content": { "field_value": "Welcome" } },
          { "info": { "label": "subheading", "type": "text" }, "content": { "field_value": "We build things" } },
          { "info": { "label": "background_image", "type": "media" }, "content": { "field_value": "hero.jpg" } }
        ],
        "nodes": []
      },
      {
        "datatype": { "info": { "label": "Cards", "type": "container" } },
        "fields": [],
        "nodes": [
          {
            "datatype": { "info": { "label": "Featured Card", "type": "card" } },
            "fields": [
              { "info": { "label": "title", "type": "text" }, "content": { "field_value": "Service 1" } },
              { "info": { "label": "description", "type": "textarea" }, "content": { "field_value": "We do this" } }
            ],
            "nodes": []
          },
          {
            "datatype": { "info": { "label": "Featured Card", "type": "card" } },
            "fields": [
              { "info": { "label": "title", "type": "text" }, "content": { "field_value": "Service 2" } },
              { "info": { "label": "description", "type": "textarea" }, "content": { "field_value": "We do that" } }
            ],
            "nodes": []
          }
        ]
      }
    ]
  }
}
```

---

## Designing Templates

When designing installable CMS templates, produce a structured definition containing:

1. **Datatypes**: Each with a `label`, `type` (use `"ROOT"` only for the page-level type), and optionally a `parent_id` referencing another datatype's label to enforce nesting rules.

2. **Fields per Datatype**: For each datatype, define its fields with `label`, `type` (from the 14 allowed values), `data` (frontend metadata), `validation` (JSON), and `ui_config` (JSON). Each field is unique to its datatype.

3. **Tree Structure**: Define the default content_data hierarchy — which datatype instances appear as children of which, and in what order. The root must be a ROOT-typed datatype.

4. **Default Field Values**: Optionally provide default `field_value` entries for content_fields to pre-populate the template.

### Template Constraints

- Exactly one datatype must have `type = "ROOT"` — this is the page-level type
- Datatypes with a `parent_id` can only be instantiated under their declared parent
- Field types must be one of: `text`, `textarea`, `number`, `date`, `datetime`, `boolean`, `select`, `media`, `relation`, `json`, `richtext`, `slug`, `email`, `url`
- All IDs are 26-character ULIDs generated at creation time
- Content status values: `draft`, `published`, `archived`, `pending`
- A route slug must be unique across the system
