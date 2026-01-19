# Clean JSON Output Examples (After Transformation)

**Created:** 2026-01-16
**Purpose:** Show what ModulaCMS JSON looks like AFTER client library transformation

---

## Concept: Before and After

```
Raw ModulaCMS API (verbose)
         ↓
Client Library Transformer
         ↓
Clean JSON (developer-friendly)
```

---

## Example 1: Blog Post

### Raw ModulaCMS JSON (What API Returns)

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
        "date_created": "2026-01-01T00:00:00Z",
        "date_modified": null,
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
          "field_value": "https://cdn.example.com/images/hero.jpg",
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
              "field_value": "Great article!",
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

### Clean JSON (After Client Library Transformation)

```json
{
  "id": 42,
  "type": "Blog Post",
  "title": "Why ModulaCMS is Different",
  "slug": "why-modulacms-is-different",
  "body": "# Introduction\n\nModulaCMS takes a different approach...",
  "featuredImage": "https://cdn.example.com/images/hero.jpg",
  "published": true,
  "comments": [
    {
      "id": 43,
      "type": "Comment",
      "commentText": "Great article!",
      "commenterName": "Jane Developer",
      "_meta": {
        "authorId": 5,
        "dateCreated": "2026-01-16T09:15:00Z",
        "dateModified": null
      }
    }
  ],
  "_meta": {
    "authorId": 1,
    "routeId": 1,
    "dateCreated": "2026-01-15T10:30:00Z",
    "dateModified": "2026-01-16T14:22:00Z"
  }
}
```

**Size comparison:**
- Raw: ~145 lines
- Clean: ~28 lines
- **80% reduction** in JSON size

---

## Example 2: E-commerce Product

### Raw ModulaCMS JSON

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
        "first_child_id": 101,
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
        "info": { "field_id": 20, "label": "Product Name", "type": "text" },
        "content": { "field_value": "Classic T-Shirt" }
      },
      {
        "info": { "field_id": 21, "label": "SKU", "type": "text" },
        "content": { "field_value": "TSHIRT-001" }
      },
      {
        "info": { "field_id": 22, "label": "Price", "type": "number" },
        "content": { "field_value": "29.99" }
      },
      {
        "info": { "field_id": 23, "label": "Description", "type": "markdown" },
        "content": { "field_value": "Premium cotton t-shirt. Soft and durable." }
      },
      {
        "info": { "field_id": 24, "label": "Inventory", "type": "integer" },
        "content": { "field_value": "150" }
      },
      {
        "info": { "field_id": 25, "label": "In Stock", "type": "boolean" },
        "content": { "field_value": "true" }
      }
    ],
    "nodes": [
      {
        "datatype": {
          "info": { "datatype_id": 4, "label": "Product Variant", "type": "variant" },
          "content": { "content_data_id": 101 }
        },
        "fields": [
          {
            "info": { "field_id": 30, "label": "Size", "type": "text" },
            "content": { "field_value": "Medium" }
          },
          {
            "info": { "field_id": 31, "label": "Color", "type": "text" },
            "content": { "field_value": "Blue" }
          },
          {
            "info": { "field_id": 32, "label": "Price Modifier", "type": "number" },
            "content": { "field_value": "0.00" }
          },
          {
            "info": { "field_id": 33, "label": "Stock", "type": "integer" },
            "content": { "field_value": "45" }
          }
        ],
        "nodes": []
      },
      {
        "datatype": {
          "info": { "datatype_id": 4, "label": "Product Variant", "type": "variant" },
          "content": { "content_data_id": 102 }
        },
        "fields": [
          {
            "info": { "field_id": 30, "label": "Size", "type": "text" },
            "content": { "field_value": "Large" }
          },
          {
            "info": { "field_id": 31, "label": "Color", "type": "text" },
            "content": { "field_value": "Red" }
          },
          {
            "info": { "field_id": 32, "label": "Price Modifier", "type": "number" },
            "content": { "field_value": "2.00" }
          },
          {
            "info": { "field_id": 33, "label": "Stock", "type": "integer" },
            "content": { "field_value": "30" }
          }
        ],
        "nodes": []
      }
    ]
  }
}
```

### Clean JSON (After Transformation)

```json
{
  "id": 100,
  "type": "Product",
  "productName": "Classic T-Shirt",
  "sku": "TSHIRT-001",
  "price": 29.99,
  "description": "Premium cotton t-shirt. Soft and durable.",
  "inventory": 150,
  "inStock": true,
  "variants": [
    {
      "id": 101,
      "type": "Product Variant",
      "size": "Medium",
      "color": "Blue",
      "priceModifier": 0.00,
      "stock": 45
    },
    {
      "id": 102,
      "type": "Product Variant",
      "size": "Large",
      "color": "Red",
      "priceModifier": 2.00,
      "stock": 30
    }
  ],
  "_meta": {
    "authorId": 1,
    "routeId": 2,
    "dateCreated": "2026-01-10T12:00:00Z",
    "dateModified": "2026-01-15T10:30:00Z"
  }
}
```

**Notice:**
- ✅ Field labels → camelCase keys
- ✅ Type coercion: `"29.99"` → `29.99` (number)
- ✅ Type coercion: `"150"` → `150` (integer)
- ✅ Type coercion: `"true"` → `true` (boolean)
- ✅ Child nodes → `variants` array
- ✅ Removed all schema definitions (info)
- ✅ Removed tree pointers (first_child_id, etc.)

---

## Example 3: Documentation Site

### Clean JSON: Category with Articles

```json
{
  "id": 200,
  "type": "Documentation Category",
  "categoryName": "Getting Started",
  "icon": "rocket",
  "order": 1,
  "articles": [
    {
      "id": 201,
      "type": "Documentation Article",
      "title": "Installation",
      "slug": "installation",
      "content": "## Installation\n\nRun `npm install @modulacms/client`",
      "order": 1,
      "sections": [
        {
          "id": 301,
          "type": "Article Section",
          "heading": "Requirements",
          "body": "Node.js 18+ required",
          "codeSnippet": null
        },
        {
          "id": 302,
          "type": "Article Section",
          "heading": "Installation Steps",
          "body": "Follow these steps:",
          "codeSnippet": "npm install @modulacms/client"
        }
      ]
    },
    {
      "id": 202,
      "type": "Documentation Article",
      "title": "Quick Start",
      "slug": "quick-start",
      "content": "## Quick Start\n\nGet started in 5 minutes",
      "order": 2,
      "sections": []
    }
  ],
  "_meta": {
    "authorId": 1,
    "routeId": 3,
    "dateCreated": "2026-01-05T00:00:00Z",
    "dateModified": "2026-01-10T15:00:00Z"
  }
}
```

**Three-level hierarchy:**
```
Category
└── Articles
    └── Sections
```

**All nested cleanly** without verbose datatype/field structure.

---

## Example 4: Multi-Language Content (Future)

### Clean JSON: Blog Post with Translations

```json
{
  "id": 42,
  "type": "Blog Post",
  "locale": "en",
  "title": "Why ModulaCMS is Different",
  "slug": "why-modulacms-is-different",
  "body": "# Introduction\n\nModulaCMS takes a different approach...",
  "translations": [
    {
      "locale": "es",
      "title": "Por qué ModulaCMS es diferente",
      "slug": "por-que-modulacms-es-diferente",
      "body": "# Introducción\n\nModulaCMS adopta un enfoque diferente..."
    },
    {
      "locale": "fr",
      "title": "Pourquoi ModulaCMS est différent",
      "slug": "pourquoi-modulacms-est-different",
      "body": "# Introduction\n\nModulaCMS adopte une approche différente..."
    }
  ],
  "_meta": {
    "authorId": 1,
    "routeId": 1,
    "dateCreated": "2026-01-15T10:30:00Z",
    "dateModified": "2026-01-16T14:22:00Z"
  }
}
```

**Note:** This would use the flexible schema (language as field) but client library makes it clean.

---

## Example 5: Custom Admin Panel Content

### Clean JSON: Dashboard Widget

```json
{
  "id": 500,
  "type": "Dashboard Widget",
  "widgetName": "Analytics Overview",
  "widgetType": "chart",
  "position": {
    "row": 1,
    "column": 1,
    "width": 6,
    "height": 4
  },
  "config": {
    "chartType": "line",
    "dataSource": "/api/analytics/pageviews",
    "refreshInterval": 300
  },
  "permissions": ["admin", "editor"],
  "_meta": {
    "authorId": 1,
    "routeId": 10,
    "dateCreated": "2026-01-12T08:00:00Z",
    "dateModified": "2026-01-14T11:30:00Z"
  }
}
```

**Demonstrates:**
- JSON field type (position, config parsed from strings)
- Admin content (route_id 10 = admin panel)
- Complex nested data

---

## Example 6: List of Items

### Clean JSON: Blog Posts List

```json
{
  "posts": [
    {
      "id": 42,
      "type": "Blog Post",
      "title": "Why ModulaCMS is Different",
      "slug": "why-modulacms-is-different",
      "excerpt": "ModulaCMS takes a different approach...",
      "featuredImage": "https://cdn.example.com/images/hero.jpg",
      "published": true,
      "publishedAt": "2026-01-15T10:30:00Z",
      "_meta": {
        "authorId": 1,
        "dateCreated": "2026-01-15T10:30:00Z"
      }
    },
    {
      "id": 43,
      "type": "Blog Post",
      "title": "Getting Started with ModulaCMS",
      "slug": "getting-started",
      "excerpt": "Learn how to install and configure ModulaCMS...",
      "featuredImage": "https://cdn.example.com/images/getting-started.jpg",
      "published": true,
      "publishedAt": "2026-01-16T08:00:00Z",
      "_meta": {
        "authorId": 1,
        "dateCreated": "2026-01-16T08:00:00Z"
      }
    }
  ],
  "pagination": {
    "total": 45,
    "page": 1,
    "perPage": 10,
    "totalPages": 5
  }
}
```

**For lists:**
- Array of clean objects
- Pagination metadata separate
- No verbose datatype info repeated

---

## Transformation Rules

### 1. Field Labels → camelCase Keys

```
"Title" → "title"
"Featured Image" → "featuredImage"
"SEO Meta Description" → "seoMetaDescription"
"Product Name" → "productName"
```

### 2. Type Coercion

```json
// Input (all strings)
{
  "field_value": "true",        // type: "boolean"
  "field_value": "42.99",       // type: "number"
  "field_value": "150",         // type: "integer"
  "field_value": "2026-01-15",  // type: "date"
  "field_value": "{\"x\": 1}",  // type: "json"
}

// Output (typed)
{
  "booleanField": true,
  "numberField": 42.99,
  "integerField": 150,
  "dateField": "2026-01-15T00:00:00Z",
  "jsonField": { "x": 1 }
}
```

### 3. Child Nodes → Arrays

```json
// Input: nodes array with "Comment" datatype
"nodes": [
  { "datatype": { "info": { "label": "Comment" } }, ... },
  { "datatype": { "info": { "label": "Comment" } }, ... }
]

// Output: "comments" array (pluralized, lowercased)
"comments": [
  { ... },
  { ... }
]
```

### 4. Metadata Consolidation

```json
// Input: scattered across datatype.info and datatype.content
"datatype": {
  "info": { "datatype_id": 1, "label": "Blog Post" },
  "content": {
    "content_data_id": 42,
    "author_id": 1,
    "date_created": "...",
    "date_modified": "..."
  }
}

// Output: consolidated in _meta
"_meta": {
  "authorId": 1,
  "routeId": 1,
  "dateCreated": "...",
  "dateModified": "..."
}
```

### 5. Remove Schema Info

```json
// Input: every field has "info" with schema
"fields": [
  {
    "info": { "field_id": 1, "label": "Title", "type": "text", ... },
    "content": { "field_value": "Hello" }
  }
]

// Output: just the value
"title": "Hello"
```

### 6. Remove Tree Pointers

```json
// Input: tree navigation pointers
"content": {
  "parent_id": 10,
  "first_child_id": 50,
  "next_sibling_id": 43,
  "prev_sibling_id": 41
}

// Output: tree structure represented by nesting
"comments": [ ... ]  // Children nested, no pointers
```

---

## Comparison: Other CMS Clean JSON

### Contentful

```json
{
  "sys": {
    "id": "42",
    "type": "Entry",
    "contentType": { "sys": { "id": "blogPost" } },
    "createdAt": "2026-01-15T10:30:00Z",
    "updatedAt": "2026-01-16T14:22:00Z"
  },
  "fields": {
    "title": "Why ModulaCMS is Different",
    "slug": "why-modulacms-is-different",
    "body": "# Introduction...",
    "featuredImage": { "sys": { "id": "asset123" } }
  }
}
```

**Still has some metadata noise**, but cleaner than raw ModulaCMS.

### Sanity

```json
{
  "_id": "42",
  "_type": "blogPost",
  "_createdAt": "2026-01-15T10:30:00Z",
  "_updatedAt": "2026-01-16T14:22:00Z",
  "title": "Why ModulaCMS is Different",
  "slug": { "current": "why-modulacms-is-different" },
  "body": "# Introduction...",
  "featuredImage": { "_ref": "image-123" }
}
```

**Very clean**, underscore prefix for system fields.

### Strapi

```json
{
  "id": 42,
  "attributes": {
    "title": "Why ModulaCMS is Different",
    "slug": "why-modulacms-is-different",
    "body": "# Introduction...",
    "featuredImage": {
      "data": {
        "id": 123,
        "attributes": { "url": "https://..." }
      }
    },
    "createdAt": "2026-01-15T10:30:00Z",
    "updatedAt": "2026-01-16T14:22:00Z"
  }
}
```

**Nested "attributes" pattern**, somewhat verbose.

### ModulaCMS (Clean)

```json
{
  "id": 42,
  "type": "Blog Post",
  "title": "Why ModulaCMS is Different",
  "slug": "why-modulacms-is-different",
  "body": "# Introduction...",
  "featuredImage": "https://cdn.example.com/images/hero.jpg",
  "_meta": {
    "authorId": 1,
    "routeId": 1,
    "dateCreated": "2026-01-15T10:30:00Z",
    "dateModified": "2026-01-16T14:22:00Z"
  }
}
```

**Flat structure, underscore prefix for metadata**, competitive with best CMSs.

---

## Developer Experience Comparison

### Without Client Library

```typescript
// Nightmare code
const title = data.root.fields.find(f => f.info.label === 'Title')?.content.field_value
const slug = data.root.fields.find(f => f.info.label === 'Slug')?.content.field_value
const body = data.root.fields.find(f => f.info.label === 'Body')?.content.field_value
const published = data.root.fields.find(f => f.info.label === 'Published')?.content.field_value === 'true'

// Accessing child nodes
const comments = data.root.nodes.map(node => {
  const text = node.fields.find(f => f.info.label === 'Comment Text')?.content.field_value
  const name = node.fields.find(f => f.info.label === 'Commenter Name')?.content.field_value
  return { text, name }
})
```

### With Client Library (Clean JSON)

```typescript
// Beautiful code
const { title, slug, body, published, comments } = post

// Accessing child properties
comments.map(comment => ({
  text: comment.commentText,
  name: comment.commenterName
}))
```

**10x less code, 100x more readable.**

---

## Real-World Usage

### Next.js Server Component

```tsx
// app/blog/[slug]/page.tsx

import { getBlogPost } from '@modulacms/nextjs'

export default async function BlogPostPage({ params }: { params: { slug: string } }) {
  const post = await getBlogPost(params.slug)  // Returns clean JSON

  return (
    <article>
      <h1>{post.title}</h1>
      <img src={post.featuredImage} alt={post.title} />
      <div dangerouslySetInnerHTML={{ __html: marked(post.body) }} />

      {post.published && (
        <time dateTime={post._meta.dateCreated}>
          Published {new Date(post._meta.dateCreated).toLocaleDateString()}
        </time>
      )}

      <section>
        <h2>Comments ({post.comments.length})</h2>
        {post.comments.map(comment => (
          <div key={comment.id}>
            <strong>{comment.commenterName}</strong>
            <p>{comment.commentText}</p>
          </div>
        ))}
      </section>
    </article>
  )
}
```

**Zero field extraction logic.** Just clean property access.

### React Component with Hooks

```tsx
import { useBlogPost } from '@modulacms/react'

export function BlogPost({ slug }: { slug: string }) {
  const { data: post, isLoading, error } = useBlogPost(slug)

  if (isLoading) return <Spinner />
  if (error) return <Error message={error.message} />

  return (
    <article>
      <h1>{post.title}</h1>
      <p>{post.body}</p>
      <CommentList comments={post.comments} />
    </article>
  )
}
```

**Zero boilerplate.** Clean, typed, cached.

---

## Summary: Clean JSON Characteristics

### Properties of Clean ModulaCMS JSON

1. **Flat field structure** - No nested `info`/`content` objects
2. **camelCase keys** - Field labels transformed to valid JS keys
3. **Typed values** - Strings coerced to boolean, number, date, JSON
4. **Nested children** - Child nodes as arrays, not tree pointers
5. **Metadata separate** - All metadata in `_meta` object
6. **No schema info** - Schema definitions removed (not needed at runtime)
7. **No tree pointers** - parent_id, next_sibling_id removed (hierarchy via nesting)
8. **Consistent IDs** - Top-level `id` property (not buried in content)

### Size Reduction

- Raw ModulaCMS JSON: ~145 lines (blog post example)
- Clean JSON: ~28 lines
- **80% reduction**

### Cognitive Load Reduction

- Raw: 5-10 lines per field extraction
- Clean: 1 line property access
- **90% reduction in code**

---

## Conclusion

**Clean ModulaCMS JSON looks like:**
- Flat, simple objects
- camelCase keys matching field labels
- Properly typed values
- Nested arrays for children
- Metadata in `_meta`
- Competitive with Contentful, Sanity, Strapi

**The transformation:**
- Client library handles complexity
- Backend stays generic and flexible
- Frontend gets clean, typed data
- Developers stay happy and productive

**This is essential for adoption.**

---

**Last Updated:** 2026-01-16
