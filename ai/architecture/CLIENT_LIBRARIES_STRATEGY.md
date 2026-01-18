# ModulaCMS Client Libraries Strategy

**Created:** 2026-01-16
**Purpose:** Design client libraries that transform verbose JSON into developer-friendly objects

---

## The Problem: Verbose JSON

### What ModulaCMS Returns (Raw API)

```json
{
  "root": {
    "datatype": {
      "info": { "datatype_id": 1, "label": "Blog Post", "type": "article" },
      "content": { "content_data_id": 42, "route_id": 1, "author_id": 1 }
    },
    "fields": [
      {
        "info": { "field_id": 1, "label": "Title", "type": "text" },
        "content": { "field_value": "Hello World" }
      },
      {
        "info": { "field_id": 2, "label": "Slug", "type": "text" },
        "content": { "field_value": "hello-world" }
      },
      {
        "info": { "field_id": 3, "label": "Body", "type": "markdown" },
        "content": { "field_value": "# Introduction\n\nThis is the content..." }
      },
      {
        "info": { "field_id": 4, "label": "Published", "type": "boolean" },
        "content": { "field_value": "true" }
      }
    ],
    "nodes": []
  }
}
```

### What Developers Want to Write

```typescript
// Instead of this horror:
const title = data.root.fields.find(f => f.info.label === 'Title')?.content.field_value
const slug = data.root.fields.find(f => f.info.label === 'Slug')?.content.field_value
const body = data.root.fields.find(f => f.info.label === 'Body')?.content.field_value
const published = data.root.fields.find(f => f.info.label === 'Published')?.content.field_value === 'true'

// Developers want this:
const { title, slug, body, published } = post
```

---

## The Solution: Client Libraries

### Transformation Layers

```
ModulaCMS API (verbose JSON)
         ↓
Client Library (transformation)
         ↓
Clean Objects (developer-friendly)
```

---

## Client Library #1: JavaScript/TypeScript SDK

### What It Does

Transforms verbose ModulaCMS JSON into clean objects:

```typescript
import { ModulaCMS } from '@modulacms/client'

const cms = new ModulaCMS({
  apiUrl: 'https://api.example.com'
})

// Fetch and transform
const post = await cms.get('blog/hello-world')

// Clean object:
console.log(post)
// {
//   id: 42,
//   type: 'Blog Post',
//   title: 'Hello World',
//   slug: 'hello-world',
//   body: '# Introduction\n\nThis is the content...',
//   published: true,
//   _meta: {
//     author_id: 1,
//     date_created: '2026-01-15T10:30:00Z',
//     date_modified: '2026-01-16T14:22:00Z'
//   },
//   comments: []
// }
```

### Implementation

```typescript
// @modulacms/client/src/transformer.ts

export class ModulaCMSTransformer {

  /**
   * Transform verbose ModulaCMS JSON into clean object
   */
  transform(data: RawResponse): Record<string, any> {
    if (!data.root) return {}

    const node = data.root
    const result: Record<string, any> = {}

    // Add metadata
    result.id = node.datatype.content.content_data_id
    result.type = node.datatype.info.label
    result._meta = {
      datatype_id: node.datatype.info.datatype_id,
      route_id: node.datatype.content.route_id,
      author_id: node.datatype.content.author_id,
      date_created: node.datatype.content.date_created,
      date_modified: node.datatype.content.date_modified,
    }

    // Transform fields into properties
    for (const field of node.fields) {
      const key = this.fieldLabelToKey(field.info.label)
      const value = this.parseFieldValue(field.content.field_value, field.info.type)
      result[key] = value
    }

    // Transform child nodes
    if (node.nodes && node.nodes.length > 0) {
      const childrenKey = this.pluralize(node.nodes[0].datatype.info.label)
      result[childrenKey] = node.nodes.map(child => this.transformNode(child))
    }

    return result
  }

  /**
   * Convert field label to camelCase property key
   * "Title" → "title"
   * "Featured Image" → "featuredImage"
   * "SEO Meta Description" → "seoMetaDescription"
   */
  private fieldLabelToKey(label: string): string {
    return label
      .split(' ')
      .map((word, i) => i === 0 ? word.toLowerCase() : word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join('')
  }

  /**
   * Parse field value based on type
   */
  private parseFieldValue(value: string, type: string): any {
    switch (type) {
      case 'boolean':
        return value === 'true'
      case 'number':
        return parseFloat(value)
      case 'integer':
        return parseInt(value, 10)
      case 'json':
        try {
          return JSON.parse(value)
        } catch {
          return value
        }
      case 'date':
        return new Date(value)
      default:
        return value
    }
  }

  /**
   * Recursively transform child node
   */
  private transformNode(node: Node): Record<string, any> {
    return this.transform({ root: node })
  }

  /**
   * Simple pluralization
   */
  private pluralize(word: string): string {
    const lower = word.toLowerCase()
    if (lower.endsWith('s')) return lower
    if (lower.endsWith('y')) return lower.slice(0, -1) + 'ies'
    return lower + 's'
  }
}
```

### Usage

```typescript
// Next.js example
import { ModulaCMS } from '@modulacms/client'

const cms = new ModulaCMS({
  apiUrl: process.env.MODULACMS_API_URL
})

export async function getBlogPost(slug: string) {
  const raw = await fetch(`${process.env.MODULACMS_API_URL}/blog/${slug}`)
  const data = await raw.json()

  // Transform
  const transformer = new ModulaCMSTransformer()
  return transformer.transform(data)
}

// In component
const post = await getBlogPost('hello-world')

return (
  <article>
    <h1>{post.title}</h1>
    <div>{post.body}</div>
    {post.comments.map(comment => (
      <div key={comment.id}>{comment.commentText}</div>
    ))}
  </article>
)
```

---

## Client Library #2: Code Generation

### Generate TypeScript Types from Datatypes

Instead of generic transformers, generate specific types and functions for your schema:

```bash
modulacms-codegen --api https://api.example.com --output ./src/cms
```

**Generated output:**

```typescript
// src/cms/types.ts

export interface BlogPost {
  id: number
  title: string
  slug: string
  body: string
  featuredImage: string | null
  published: boolean
  _meta: ContentMetadata
  comments: Comment[]
}

export interface Comment {
  id: number
  commentText: string
  commenterName: string
  commenterEmail: string | null
  _meta: ContentMetadata
}

export interface Product {
  id: number
  productName: string
  price: number
  description: string
  sku: string
  inventory: number
  _meta: ContentMetadata
  variants: ProductVariant[]
}
```

```typescript
// src/cms/client.ts

import { BlogPost, Comment, Product } from './types'

export class CMSClient {
  constructor(private apiUrl: string) {}

  async getBlogPost(slug: string): Promise<BlogPost> {
    const response = await fetch(`${this.apiUrl}/blog/${slug}`)
    const data = await response.json()
    return transformBlogPost(data)
  }

  async listBlogPosts(): Promise<BlogPost[]> {
    const response = await fetch(`${this.apiUrl}/blog`)
    const data = await response.json()
    return data.posts.map(transformBlogPost)
  }

  async getProduct(sku: string): Promise<Product> {
    const response = await fetch(`${this.apiUrl}/products/${sku}`)
    const data = await response.json()
    return transformProduct(data)
  }
}

// Type-safe transformers (generated)
function transformBlogPost(raw: any): BlogPost {
  return {
    id: raw.root.datatype.content.content_data_id,
    title: getField(raw.root, 'Title'),
    slug: getField(raw.root, 'Slug'),
    body: getField(raw.root, 'Body'),
    featuredImage: getField(raw.root, 'Featured Image'),
    published: getField(raw.root, 'Published') === 'true',
    _meta: extractMetadata(raw.root),
    comments: raw.root.nodes.map(transformComment)
  }
}

function transformComment(raw: any): Comment {
  return {
    id: raw.datatype.content.content_data_id,
    commentText: getField(raw, 'Comment Text'),
    commenterName: getField(raw, 'Commenter Name'),
    commenterEmail: getField(raw, 'Commenter Email'),
    _meta: extractMetadata(raw)
  }
}
```

**Benefits:**
- ✅ Fully type-safe (TypeScript knows your schema)
- ✅ Autocomplete in IDE
- ✅ Compile-time errors if schema changes
- ✅ No runtime overhead finding fields
- ✅ Optimized transformers

---

## Client Library #3: GraphQL Layer

### ModulaCMS GraphQL Wrapper

Build a GraphQL API on top of ModulaCMS:

```graphql
# schema.graphql

type BlogPost {
  id: Int!
  title: String!
  slug: String!
  body: String!
  featuredImage: String
  published: Boolean!
  dateCreated: String
  dateModified: String
  comments: [Comment!]!
}

type Comment {
  id: Int!
  commentText: String!
  commenterName: String!
  commenterEmail: String
}

type Query {
  blogPost(slug: String!): BlogPost
  blogPosts(limit: Int, offset: Int): [BlogPost!]!
  product(sku: String!): Product
}
```

**Usage:**

```typescript
import { gql } from '@apollo/client'

const GET_BLOG_POST = gql`
  query GetBlogPost($slug: String!) {
    blogPost(slug: $slug) {
      title
      body
      published
      comments {
        commentText
        commenterName
      }
    }
  }
`

const { data } = useQuery(GET_BLOG_POST, { variables: { slug: 'hello-world' } })

return <h1>{data.blogPost.title}</h1>
```

**Implementation:**

```typescript
// GraphQL resolver
const resolvers = {
  Query: {
    blogPost: async (_: any, { slug }: { slug: string }) => {
      const response = await fetch(`${MODULACMS_API}/blog/${slug}`)
      const data = await response.json()
      return transformBlogPost(data)
    }
  }
}
```

**Benefits:**
- ✅ Query only fields you need
- ✅ No over-fetching
- ✅ Strong typing with GraphQL codegen
- ✅ Client-side caching (Apollo, urql)
- ✅ Familiar to many developers

---

## Client Library #4: React Hooks

### Custom hooks for React/Next.js

```typescript
// @modulacms/react

import { useQuery } from '@tanstack/react-query'
import { transformBlogPost } from './transformers'

export function useBlogPost(slug: string) {
  return useQuery({
    queryKey: ['blogPost', slug],
    queryFn: async () => {
      const response = await fetch(`${process.env.MODULACMS_API}/blog/${slug}`)
      const data = await response.json()
      return transformBlogPost(data)
    }
  })
}

export function useBlogPosts() {
  return useQuery({
    queryKey: ['blogPosts'],
    queryFn: async () => {
      const response = await fetch(`${process.env.MODULACMS_API}/blog`)
      const data = await response.json()
      return data.posts.map(transformBlogPost)
    }
  })
}
```

**Usage:**

```tsx
import { useBlogPost } from '@modulacms/react'

export default function BlogPostPage({ params }: { params: { slug: string } }) {
  const { data: post, isLoading, error } = useBlogPost(params.slug)

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>Error: {error.message}</div>

  return (
    <article>
      <h1>{post.title}</h1>
      <div>{post.body}</div>
    </article>
  )
}
```

**Benefits:**
- ✅ Built-in loading/error states
- ✅ Automatic caching and refetching
- ✅ React Suspense support
- ✅ Clean component code

---

## Client Library #5: Next.js Integration

### Helper functions for Next.js Server Components

```typescript
// @modulacms/nextjs

import { cache } from 'react'
import { transformBlogPost } from './transformers'

export const getBlogPost = cache(async (slug: string) => {
  const response = await fetch(`${process.env.MODULACMS_API}/blog/${slug}`, {
    next: { revalidate: 3600 } // ISR: revalidate every hour
  })
  const data = await response.json()
  return transformBlogPost(data)
})

export const getBlogPosts = cache(async () => {
  const response = await fetch(`${process.env.MODULACMS_API}/blog`, {
    next: { revalidate: 300 } // ISR: revalidate every 5 minutes
  })
  const data = await response.json()
  return data.posts.map(transformBlogPost)
})

export async function generateStaticParams() {
  const posts = await getBlogPosts()
  return posts.map(post => ({ slug: post.slug }))
}
```

**Usage:**

```tsx
// app/blog/[slug]/page.tsx

import { getBlogPost, generateStaticParams } from '@modulacms/nextjs'

export { generateStaticParams }

export default async function BlogPostPage({ params }: { params: { slug: string } }) {
  const post = await getBlogPost(params.slug)

  return (
    <article>
      <h1>{post.title}</h1>
      <div>{post.body}</div>
    </article>
  )
}
```

**Benefits:**
- ✅ Next.js ISR/SSG built-in
- ✅ Automatic deduplication with `cache()`
- ✅ Server Components ready
- ✅ TypeScript support

---

## Comparison: Other CMS Client Libraries

### Contentful

```typescript
// Contentful has great client library
import { createClient } from 'contentful'

const client = createClient({
  space: 'your_space',
  accessToken: 'your_token'
})

const entry = await client.getEntry('entry_id')
console.log(entry.fields.title) // Clean access
```

### Sanity

```typescript
// Sanity has GROQ query language
import { createClient } from '@sanity/client'

const client = createClient({
  projectId: 'your_project',
  dataset: 'production'
})

const posts = await client.fetch(`
  *[_type == "post"] {
    title,
    slug,
    body
  }
`)
```

### Strapi

```typescript
// Strapi has auto-generated REST/GraphQL
const response = await fetch('https://api.example.com/api/posts?populate=*')
const data = await response.json()

// Still verbose, but better than raw
data.data.forEach(post => {
  console.log(post.attributes.title)
})
```

### ModulaCMS (Raw)

```typescript
// Without client library
const response = await fetch('https://api.example.com/blog/hello-world')
const data = await response.json()

const title = data.root.fields.find(f => f.info.label === 'Title')?.content.field_value
// 😭 Painful
```

### ModulaCMS (With Client Library)

```typescript
// With client library
import { getBlogPost } from '@modulacms/client'

const post = await getBlogPost('hello-world')
console.log(post.title) // 🎉 Clean!
```

---

## Implementation Priority

### Phase 1: Basic Transformer (JavaScript/TypeScript)

**Package:** `@modulacms/client`

```bash
npm install @modulacms/client
```

**Features:**
- Generic transformer (works with any schema)
- Field label → camelCase keys
- Type coercion (string → boolean, number, date)
- Child nodes → nested arrays
- TypeScript types for raw API

**Timeline:** 2-4 weeks

---

### Phase 2: Code Generator

**Package:** `@modulacms/codegen`

```bash
npm install -g @modulacms/codegen
modulacms-codegen --api https://api.example.com --output ./src/cms
```

**Features:**
- Introspect API to discover datatypes/fields
- Generate TypeScript interfaces
- Generate type-safe transformers
- Generate client class with methods
- Watch mode (regenerate on schema changes)

**Timeline:** 4-6 weeks

---

### Phase 3: Framework Integrations

**Packages:**
- `@modulacms/react` - React hooks
- `@modulacms/nextjs` - Next.js helpers
- `@modulacms/vue` - Vue composables
- `@modulacms/svelte` - Svelte stores

**Timeline:** 2 weeks per framework

---

### Phase 4: GraphQL Wrapper (Optional)

**Package:** `@modulacms/graphql`

**Features:**
- Auto-generate GraphQL schema from datatypes
- Resolvers that fetch and transform
- Subscriptions for real-time (future)

**Timeline:** 6-8 weeks

---

## Why This Matters

### Developer Experience

**Without client library:**
```typescript
// 10 lines of boilerplate for every field
const title = data.root.fields.find(f => f.info.label === 'Title')?.content.field_value
const slug = data.root.fields.find(f => f.info.label === 'Slug')?.content.field_value
const body = data.root.fields.find(f => f.info.label === 'Body')?.content.field_value
const published = data.root.fields.find(f => f.info.label === 'Published')?.content.field_value === 'true'
const image = data.root.fields.find(f => f.info.label === 'Featured Image')?.content.field_value
// ... repeat for every content type
```

**With client library:**
```typescript
// 1 line
const { title, slug, body, published, featuredImage } = await cms.get('blog/hello-world')
```

**Impact:** 10x less code, 10x faster development

---

### Adoption

**Barrier to entry:**
- ❌ Raw JSON: "This is too complex, I'll use Contentful"
- ✅ Client library: "This is clean, I'll try ModulaCMS"

**Client libraries = Adoption**

---

### Competition

| CMS | Client Library Quality | Adoption |
|-----|----------------------|----------|
| Contentful | ⭐⭐⭐⭐⭐ Excellent | Very High |
| Sanity | ⭐⭐⭐⭐⭐ Excellent (GROQ) | High |
| Strapi | ⭐⭐⭐ Good | Medium |
| WordPress (headless) | ⭐⭐ Okay (WPGraphQL) | High (legacy) |
| **ModulaCMS (raw)** | ⭐ None | Low |
| **ModulaCMS (with libs)** | ⭐⭐⭐⭐⭐ Excellent | **High potential** |

**Client libraries are NOT optional for adoption.**

---

## Architecture Philosophy

### Backend: Simple and Generic

```
ModulaCMS backend:
- Generic datatype/field system
- Tree structure
- Metadata (author, dates, history)
- Raw JSON output
```

**Philosophy:** "Sophistication through simplicity"

**Result:** Backend stays small, fast, maintainable

---

### Client Libraries: Developer Experience

```
Client libraries:
- Hide complexity
- Clean APIs
- Type safety
- Framework integrations
```

**Philosophy:** "Backend is minimal, clients add sugar"

**Result:** Developers get best of both worlds

---

## Real-World Example

### Agency Workflow (Today)

```typescript
// Custom transformer per project
function extractFields(data: any) {
  const result: any = {}
  for (const field of data.root.fields) {
    const key = field.info.label.toLowerCase().replace(/ /g, '_')
    result[key] = field.content.field_value
  }
  return result
}

// Every agency rebuilds this
```

### Agency Workflow (With Client Library)

```typescript
// Install once
npm install @modulacms/client

// Use everywhere
import { ModulaCMS } from '@modulacms/client'
const cms = new ModulaCMS({ apiUrl: process.env.API_URL })
const post = await cms.get('blog/hello-world')
```

**Savings:** 100+ hours per agency project

---

## Community Contributions

### Open Source Strategy

**Core maintained by ModulaCMS team:**
- `@modulacms/client` (JavaScript/TypeScript transformer)
- `@modulacms/codegen` (code generator)

**Community can build:**
- `@modulacms/react` (React hooks)
- `@modulacms/vue` (Vue composables)
- `@modulacms/svelte` (Svelte stores)
- `@modulacms/angular` (Angular services)
- `@modulacms/python` (Python SDK)
- `@modulacms/go` (Go SDK)
- `@modulacms/php` (PHP SDK)

**Why this works:**
- Transformation logic is simple
- Raw API is well-documented
- Contributors add value for their ecosystems
- Ecosystem grows organically

---

## Conclusion

### Yes, Frontend Translators Make Total Sense

**Reasons:**

1. **Developer Experience** - Raw JSON is too verbose
2. **Type Safety** - Generated types prevent errors
3. **Framework Integration** - Hooks, composables, helpers
4. **Adoption** - Clean APIs → more users
5. **Competition** - Every major CMS has client libraries
6. **Philosophy** - Backend stays simple, clients add sugar

### The Strategy

**Backend (ModulaCMS core):**
- Generic, flexible, powerful
- Verbose JSON with full metadata
- Simple, maintainable codebase

**Client Libraries:**
- Transform verbose → clean
- Add framework integrations
- Type safety and codegen
- Community contributions

**Result:**
- Backend: Sophisticated through simplicity
- Frontend: Clean, typed, framework-integrated
- Developers: Happy and productive
- Agencies: Save hundreds of hours

### Next Steps

1. Build `@modulacms/client` (basic transformer)
2. Build `@modulacms/codegen` (type generator)
3. Document transformation patterns
4. Encourage community framework integrations
5. Measure adoption impact

**Client libraries are not optional. They're essential for adoption.**

---

**Last Updated:** 2026-01-16
