# ModulaCMS JSON Output Examples

**Created:** 2026-01-16
**Purpose:** Show what ModulaCMS API JSON responses look like

---

## Overview

ModulaCMS outputs hierarchical JSON representing content trees. The structure is:

```
Root
└── Node (content item)
    ├── Datatype (defines what type of content: blog post, page, product, etc.)
    │   ├── Info (datatype metadata)
    │   └── Content (content_data with tree pointers)
    ├── Fields[] (array of field values)
    │   ├── Info (field definition: label, type, data)
    │   └── Content (actual field value)
    └── Nodes[] (child nodes, recursive)
```

---

## JSON Structure

### Root Level

```json
{
  "root": {
    "datatype": { ... },
    "fields": [ ... ],
    "nodes": [ ... ]
  }
}
```

---

## Complete Example: Blog Post

```json
{
  "root": {
    "datatype": {
      "info": {
        "datatype_id": 1,
        "parent_id": null,
        "label": "Blog Post",
        "type": "article",
        "author_id": 1,
        "date_created": "2026-01-15T10:30:00Z",
        "date_modified": "2026-01-16T14:22:00Z",
        "history": null
      },
      "content": {
        "content_data_id": 42,
        "parent_id": null,
        "first_child_id": 43,
        "next_sibling_id": null,
        "prev_sibling_id": null,
        "route_id": 1,
        "datatype_id": 1,
        "author_id": 1,
        "date_created": "2026-01-15T10:30:00Z",
        "date_modified": "2026-01-16T14:22:00Z",
        "history": null
      }
    },
    "fields": [
      {
        "info": {
          "field_id": 1,
          "parent_id": 1,
          "label": "Title",
          "data": "",
          "type": "text",
          "author_id": 1,
          "date_created": "2026-01-01T00:00:00Z",
          "date_modified": null,
          "history": null
        },
        "content": {
          "content_field_id": 101,
          "route_id": 1,
          "content_data_id": 42,
          "field_id": 1,
          "field_value": "Why ModulaCMS is Different",
          "author_id": 1,
          "date_created": "2026-01-15T10:30:00Z",
          "date_modified": "2026-01-16T14:22:00Z",
          "history": null
        }
      },
      {
        "info": {
          "field_id": 2,
          "parent_id": 1,
          "label": "Slug",
          "data": "",
          "type": "text",
          "author_id": 1,
          "date_created": "2026-01-01T00:00:00Z",
          "date_modified": null,
          "history": null
        },
        "content": {
          "content_field_id": 102,
          "route_id": 1,
          "content_data_id": 42,
          "field_id": 2,
          "field_value": "why-modulacms-is-different",
          "author_id": 1,
          "date_created": "2026-01-15T10:30:00Z",
          "date_modified": null,
          "history": null
        }
      },
      {
        "info": {
          "field_id": 3,
          "parent_id": 1,
          "label": "Body",
          "data": "",
          "type": "markdown",
          "author_id": 1,
          "date_created": "2026-01-01T00:00:00Z",
          "date_modified": null,
          "history": null
        },
        "content": {
          "content_field_id": 103,
          "route_id": 1,
          "content_data_id": 42,
          "field_id": 3,
          "field_value": "# Introduction\n\nModulaCMS takes a different approach...",
          "author_id": 1,
          "date_created": "2026-01-15T10:30:00Z",
          "date_modified": "2026-01-16T14:22:00Z",
          "history": null
        }
      },
      {
        "info": {
          "field_id": 4,
          "parent_id": 1,
          "label": "Featured Image",
          "data": "",
          "type": "image",
          "author_id": 1,
          "date_created": "2026-01-01T00:00:00Z",
          "date_modified": null,
          "history": null
        },
        "content": {
          "content_field_id": 104,
          "route_id": 1,
          "content_data_id": 42,
          "field_id": 4,
          "field_value": "https://cdn.example.com/images/blog-hero.jpg",
          "author_id": 1,
          "date_created": "2026-01-15T10:30:00Z",
          "date_modified": null,
          "history": null
        }
      },
      {
        "info": {
          "field_id": 5,
          "parent_id": 1,
          "label": "Published",
          "data": "",
          "type": "boolean",
          "author_id": 1,
          "date_created": "2026-01-01T00:00:00Z",
          "date_modified": null,
          "history": null
        },
        "content": {
          "content_field_id": 105,
          "route_id": 1,
          "content_data_id": 42,
          "field_id": 5,
          "field_value": "true",
          "author_id": 1,
          "date_created": "2026-01-15T10:30:00Z",
          "date_modified": "2026-01-16T14:22:00Z",
          "history": null
        }
      }
    ],
    "nodes": [
      {
        "datatype": {
          "info": {
            "datatype_id": 2,
            "parent_id": 1,
            "label": "Comment",
            "type": "comment",
            "author_id": 1,
            "date_created": "2026-01-01T00:00:00Z",
            "date_modified": null,
            "history": null
          },
          "content": {
            "content_data_id": 43,
            "parent_id": 42,
            "first_child_id": null,
            "next_sibling_id": null,
            "prev_sibling_id": null,
            "route_id": 1,
            "datatype_id": 2,
            "author_id": 5,
            "date_created": "2026-01-16T09:15:00Z",
            "date_modified": null,
            "history": null
          }
        },
        "fields": [
          {
            "info": {
              "field_id": 10,
              "parent_id": 2,
              "label": "Comment Text",
              "data": "",
              "type": "text",
              "author_id": 1,
              "date_created": "2026-01-01T00:00:00Z",
              "date_modified": null,
              "history": null
            },
            "content": {
              "content_field_id": 201,
              "route_id": 1,
              "content_data_id": 43,
              "field_id": 10,
              "field_value": "Great article! This explains exactly what I've been looking for.",
              "author_id": 5,
              "date_created": "2026-01-16T09:15:00Z",
              "date_modified": null,
              "history": null
            }
          },
          {
            "info": {
              "field_id": 11,
              "parent_id": 2,
              "label": "Commenter Name",
              "data": "",
              "type": "text",
              "author_id": 1,
              "date_created": "2026-01-01T00:00:00Z",
              "date_modified": null,
              "history": null
            },
            "content": {
              "content_field_id": 202,
              "route_id": 1,
              "content_data_id": 43,
              "field_id": 11,
              "field_value": "Jane Developer",
              "author_id": 5,
              "date_created": "2026-01-16T09:15:00Z",
              "date_modified": null,
              "history": null
            }
          }
        ],
        "nodes": []
      }
    ]
  }
}
```

---

## Key JSON Fields Explained

### Datatype Object

**info** - Datatype definition (template/schema)
- `datatype_id` - Unique ID for this datatype
- `parent_id` - Parent datatype (for inheritance)
- `label` - Human-readable name ("Blog Post", "Product", etc.)
- `type` - Type identifier ("article", "page", "product")
- `author_id` - Who created this datatype
- `date_created` - When datatype was created
- `date_modified` - Last modification time
- `history` - Change history (JSON)

**content** - Actual content instance using this datatype
- `content_data_id` - Unique ID for this content item
- `parent_id` - Parent content item (tree structure)
- `first_child_id` - First child in tree (O(1) traversal)
- `next_sibling_id` - Next sibling (O(1) traversal)
- `prev_sibling_id` - Previous sibling (O(1) traversal)
- `route_id` - Which route/site this belongs to
- `datatype_id` - References the datatype
- `author_id` - Who created this content
- `date_created` - When content was created
- `date_modified` - Last modification time
- `history` - Change history (JSON)

### Field Object

**info** - Field definition (schema)
- `field_id` - Unique ID for this field definition
- `parent_id` - Parent datatype
- `label` - Field name ("Title", "Body", "Price")
- `data` - Default value or field metadata
- `type` - Field type ("text", "markdown", "number", "boolean", "image")
- `author_id` - Who created this field
- `date_created` - When field was defined
- `date_modified` - Last modification time
- `history` - Change history (JSON)

**content** - Actual field value
- `content_field_id` - Unique ID for this value
- `route_id` - Which route/site
- `content_data_id` - Which content item this belongs to
- `field_id` - References the field definition
- `field_value` - **THE ACTUAL VALUE** (stored as string, typed by field.type)
- `author_id` - Who set this value
- `date_created` - When value was set
- `date_modified` - Last modification time
- `history` - Change history (JSON)

---

## Simplified Example: Product

```json
{
  "root": {
    "datatype": {
      "info": {
        "datatype_id": 3,
        "parent_id": null,
        "label": "Product",
        "type": "product",
        "author_id": 1,
        "date_created": "2026-01-01T00:00:00Z",
        "date_modified": null,
        "history": null
      },
      "content": {
        "content_data_id": 100,
        "parent_id": null,
        "first_child_id": null,
        "next_sibling_id": null,
        "prev_sibling_id": null,
        "route_id": 2,
        "datatype_id": 3,
        "author_id": 1,
        "date_created": "2026-01-10T12:00:00Z",
        "date_modified": "2026-01-15T10:30:00Z",
        "history": null
      }
    },
    "fields": [
      {
        "info": {
          "field_id": 20,
          "parent_id": 3,
          "label": "Product Name",
          "data": "",
          "type": "text",
          "author_id": 1,
          "date_created": "2026-01-01T00:00:00Z",
          "date_modified": null,
          "history": null
        },
        "content": {
          "content_field_id": 301,
          "route_id": 2,
          "content_data_id": 100,
          "field_id": 20,
          "field_value": "ModulaCMS Self-Hosted License",
          "author_id": 1,
          "date_created": "2026-01-10T12:00:00Z",
          "date_modified": null,
          "history": null
        }
      },
      {
        "info": {
          "field_id": 21,
          "parent_id": 3,
          "label": "Price",
          "data": "",
          "type": "number",
          "author_id": 1,
          "date_created": "2026-01-01T00:00:00Z",
          "date_modified": null,
          "history": null
        },
        "content": {
          "content_field_id": 302,
          "route_id": 2,
          "content_data_id": 100,
          "field_id": 21,
          "field_value": "0.00",
          "author_id": 1,
          "date_created": "2026-01-10T12:00:00Z",
          "date_modified": null,
          "history": null
        }
      },
      {
        "info": {
          "field_id": 22,
          "parent_id": 3,
          "label": "Description",
          "data": "",
          "type": "markdown",
          "author_id": 1,
          "date_created": "2026-01-01T00:00:00Z",
          "date_modified": null,
          "history": null
        },
        "content": {
          "content_field_id": 303,
          "route_id": 2,
          "content_data_id": 100,
          "field_id": 22,
          "field_value": "Open source headless CMS. Deploy anywhere. MIT License.",
          "author_id": 1,
          "date_created": "2026-01-10T12:00:00Z",
          "date_modified": "2026-01-15T10:30:00Z",
          "history": null
        }
      }
    ],
    "nodes": []
  }
}
```

---

## Null Values

ModulaCMS uses custom null handling for JSON:

```json
{
  "parent_id": null,          // NullInt64 - no parent
  "first_child_id": 42,       // NullInt64 - has child
  "date_modified": null,      // NullString - never modified
  "history": null             // NullString - no history
}
```

**Why custom nulls?**
- SQL `NULL` values serialize properly to JSON `null`
- Valid values serialize to actual values (not `{Valid: true, Int64: 42}`)

---

## Tree Structure Pointers

ModulaCMS uses **sibling pointers** for O(1) tree traversal:

```json
{
  "content": {
    "content_data_id": 42,
    "parent_id": 10,           // Parent node
    "first_child_id": 50,      // First child (linked list head)
    "next_sibling_id": 43,     // Next sibling
    "prev_sibling_id": 41      // Previous sibling
  }
}
```

**Tree visualization:**
```
Node 10 (parent)
├── Node 41 (child 1) ← prev_sibling_id
├── Node 42 (child 2) ← THIS NODE
│   └── Node 50 (grandchild) ← first_child_id
└── Node 43 (child 3) ← next_sibling_id
```

**Benefits:**
- O(1) get first child
- O(1) get next/prev sibling
- O(1) insert before/after
- No recursive queries for siblings

---

## Multiple Content Items (List)

When fetching multiple items (e.g., all blog posts), wrap in array:

```json
{
  "posts": [
    {
      "datatype": { ... },
      "fields": [ ... ],
      "nodes": [ ... ]
    },
    {
      "datatype": { ... },
      "fields": [ ... ],
      "nodes": [ ... ]
    }
  ]
}
```

---

## Next.js Usage Example

```typescript
// Fetch blog post
const response = await fetch('https://api.example.com/blog/why-modulacms-is-different')
const data = await response.json()

// Extract fields
const title = data.root.fields.find(f => f.info.label === 'Title')?.content.field_value
const slug = data.root.fields.find(f => f.info.label === 'Slug')?.content.field_value
const body = data.root.fields.find(f => f.info.label === 'Body')?.content.field_value
const published = data.root.fields.find(f => f.info.label === 'Published')?.content.field_value === 'true'

// Extract child nodes (comments)
const comments = data.root.nodes.map(node => {
  const text = node.fields.find(f => f.info.label === 'Comment Text')?.content.field_value
  const name = node.fields.find(f => f.info.label === 'Commenter Name')?.content.field_value
  return { text, name }
})

// Render
return (
  <article>
    <h1>{title}</h1>
    <div dangerouslySetInnerHTML={{ __html: marked(body) }} />
    <section>
      <h2>Comments</h2>
      {comments.map((comment, i) => (
        <div key={i}>
          <strong>{comment.name}</strong>: {comment.text}
        </div>
      ))}
    </section>
  </article>
)
```

---

## TypeScript Types

```typescript
interface Root {
  root: Node
}

interface Node {
  datatype: Datatype
  fields: Field[]
  nodes: Node[]
}

interface Datatype {
  info: DatatypeInfo
  content: ContentData
}

interface DatatypeInfo {
  datatype_id: number
  parent_id: number | null
  label: string
  type: string
  author_id: number
  date_created: string | null
  date_modified: string | null
  history: string | null
}

interface ContentData {
  content_data_id: number
  parent_id: number | null
  first_child_id: number | null
  next_sibling_id: number | null
  prev_sibling_id: number | null
  route_id: number
  datatype_id: number
  author_id: number
  date_created: string | null
  date_modified: string | null
  history: string | null
}

interface Field {
  info: FieldInfo
  content: ContentField
}

interface FieldInfo {
  field_id: number
  parent_id: number | null
  label: string | any  // Can be string or other types
  data: string
  type: string  // "text" | "markdown" | "number" | "boolean" | "image" | etc.
  author_id: number
  date_created: string | null
  date_modified: string | null
  history: string | null
}

interface ContentField {
  content_field_id: number
  route_id: number | null
  content_data_id: number
  field_id: number
  field_value: string  // Stored as string, typed by field.info.type
  author_id: number
  date_created: string | null
  date_modified: string | null
  history: string | null
}
```

---

## Key Insights

### 1. Separation of Schema and Data

**Schema (info)** - Reusable definitions
- Datatypes define content types
- Fields define field definitions

**Data (content)** - Actual content instances
- ContentData instances using datatypes
- ContentFields storing actual values

**Benefits:**
- Change schema without migrating data
- Reuse datatypes across routes
- Add fields without code changes

### 2. Everything is a String (Until It's Not)

All field values stored as strings:
```json
{
  "field_value": "true"       // boolean
  "field_value": "42.99"      // number
  "field_value": "Hello"      // text
  "field_value": "# Markdown" // markdown
  "field_value": "https://..." // image URL
}
```

Frontend types them based on `field.info.type`:
```typescript
if (field.info.type === 'boolean') return field.content.field_value === 'true'
if (field.info.type === 'number') return parseFloat(field.content.field_value)
// etc.
```

### 3. Hierarchical and Relational

**Hierarchical** - Tree structure via nodes
```json
"nodes": [
  { "datatype": { ... }, "nodes": [ ... ] }  // Recursive
]
```

**Relational** - IDs reference other records
```json
{
  "content_data_id": 42,
  "datatype_id": 1,        // References datatype
  "field_id": 10,          // References field definition
  "parent_id": 10,         // References parent content
  "route_id": 1            // References route
}
```

### 4. Dual Content Model

Same structure for public and admin content:

**Public content:**
- `datatype` → `content_data` → `content_fields`
- Routes: blog, pages, products

**Admin content:**
- `admin_datatype` → `admin_content_data` → `admin_content_fields`
- Admin routes: dashboard, settings, analytics

**Same JSON structure, different trees!**

---

## Why This Structure?

### Advantages

1. **Flexible** - Any content type without code changes
2. **Hierarchical** - Nested content naturally represented
3. **Fast** - O(1) tree operations via sibling pointers
4. **Queryable** - Relational structure for complex queries
5. **Versioned** - History field tracks changes
6. **Multi-tenant** - Routes separate content by site
7. **Typed** - Field types provide validation hints
8. **Cacheable** - Entire tree in one response

### Trade-offs

1. **Verbose** - More JSON than traditional flat structure
2. **Parsing** - Frontend must extract fields by label
3. **Type coercion** - Strings need parsing (number, boolean)

**Solution:** Client libraries to simplify extraction

---

## Real-World Use Cases

### Blog Platform

```
Root: Blog Post
├── Fields: title, slug, body, featured_image, published
└── Nodes: Comments
    └── Fields: comment_text, commenter_name, commenter_email
```

### E-commerce

```
Root: Product
├── Fields: name, price, description, sku, inventory
└── Nodes: Variants
    ├── Fields: size, color, price_modifier
    └── Nodes: Images
        └── Fields: url, alt_text
```

### Documentation Site

```
Root: Category
├── Fields: category_name, icon
└── Nodes: Articles
    ├── Fields: title, slug, content, order
    └── Nodes: Sections
        └── Fields: heading, body, code_snippet
```

### Multi-site Agency

```
Route 1: Client A website
Route 2: Client B website
Route 3: Client A admin panel
Route 4: Client B admin panel

Same backend, different content trees!
```

---

**Last Updated:** 2026-01-16
