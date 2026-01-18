# ModulaCMS Transform Package

**Created:** 2026-01-16
**Purpose:** Transform ModulaCMS JSON to match popular CMS formats (Contentful, Sanity, Strapi, WordPress)

---

## The Vision: Drop-in Replacement

### The Problem

Switching CMSs requires rewriting frontend code:

```typescript
// Using Contentful
const entry = await contentful.getEntry('42')
const title = entry.fields.title

// Switching to ModulaCMS (raw)
const response = await fetch('https://api.modulacms.com/content/42')
const data = await response.json()
const title = data.root.fields.find(f => f.info.label === 'Title')?.content.field_value
// 😭 Have to rewrite ALL frontend code
```

### The Solution

ModulaCMS transformers match existing CMS formats:

```typescript
// Using Contentful
const entry = await contentful.getEntry('42')
const title = entry.fields.title

// Switching to ModulaCMS with Contentful transformer
import { getEntry } from '@modulacms/contentful'
const entry = await getEntry('42')
const title = entry.fields.title
// 🎉 Zero frontend changes!
```

---

## Package Structure

```
@modulacms/transform
├── src/
│   ├── core/
│   │   ├── transformer.ts       # Base transformer class
│   │   └── types.ts             # Common types
│   ├── formats/
│   │   ├── contentful.ts        # Contentful-compatible output
│   │   ├── sanity.ts            # Sanity-compatible output
│   │   ├── strapi.ts            # Strapi-compatible output
│   │   ├── wordpress.ts         # WordPress REST API compatible
│   │   ├── clean.ts             # Clean ModulaCMS format
│   │   └── graphql.ts           # GraphQL-compatible output
│   ├── clients/
│   │   ├── contentful.ts        # Drop-in Contentful client
│   │   ├── sanity.ts            # Drop-in Sanity client
│   │   ├── strapi.ts            # Drop-in Strapi client
│   │   └── wordpress.ts         # Drop-in WordPress client
│   └── index.ts
├── package.json
└── README.md
```

---

## Transformer: Contentful Format

### Contentful JSON Structure

```json
{
  "sys": {
    "id": "42",
    "type": "Entry",
    "contentType": {
      "sys": {
        "id": "blogPost",
        "type": "Link",
        "linkType": "ContentType"
      }
    },
    "space": { "sys": { "id": "space123" } },
    "createdAt": "2026-01-15T10:30:00Z",
    "updatedAt": "2026-01-16T14:22:00Z",
    "revision": 2
  },
  "fields": {
    "title": "Why ModulaCMS is Different",
    "slug": "why-modulacms-is-different",
    "body": "# Introduction\n\nModulaCMS takes a different approach...",
    "featuredImage": {
      "sys": {
        "id": "asset123",
        "type": "Link",
        "linkType": "Asset"
      }
    },
    "published": true
  }
}
```

### Implementation

```typescript
// src/formats/contentful.ts

import { ModulaCMSNode, ModulaCMSRoot } from '../core/types'

export interface ContentfulEntry {
  sys: {
    id: string
    type: 'Entry'
    contentType: {
      sys: {
        id: string
        type: 'Link'
        linkType: 'ContentType'
      }
    }
    space?: {
      sys: {
        id: string
        type: 'Link'
        linkType: 'Space'
      }
    }
    createdAt: string
    updatedAt: string
    revision?: number
  }
  fields: Record<string, any>
}

export class ContentfulTransformer {

  transform(data: ModulaCMSRoot, spaceId?: string): ContentfulEntry {
    const node = data.root

    return {
      sys: {
        id: node.datatype.content.content_data_id.toString(),
        type: 'Entry',
        contentType: {
          sys: {
            id: this.toContentfulId(node.datatype.info.label),
            type: 'Link',
            linkType: 'ContentType'
          }
        },
        ...(spaceId && {
          space: {
            sys: {
              id: spaceId,
              type: 'Link',
              linkType: 'Space'
            }
          }
        }),
        createdAt: node.datatype.content.date_created || new Date().toISOString(),
        updatedAt: node.datatype.content.date_modified || node.datatype.content.date_created || new Date().toISOString(),
        revision: 1
      },
      fields: this.extractFields(node)
    }
  }

  private extractFields(node: ModulaCMSNode): Record<string, any> {
    const fields: Record<string, any> = {}

    for (const field of node.fields) {
      const key = this.fieldLabelToKey(field.info.label)
      const value = this.parseFieldValue(field.content.field_value, field.info.type)

      // Handle references (images, links, etc.)
      if (field.info.type === 'image' || field.info.type === 'asset') {
        fields[key] = {
          sys: {
            id: this.extractAssetId(value),
            type: 'Link',
            linkType: 'Asset'
          }
        }
      } else {
        fields[key] = value
      }
    }

    return fields
  }

  private fieldLabelToKey(label: string): string {
    return label
      .split(' ')
      .map((word, i) => i === 0 ? word.toLowerCase() : word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join('')
  }

  private parseFieldValue(value: string, type: string): any {
    switch (type) {
      case 'boolean':
        return value === 'true'
      case 'number':
      case 'decimal':
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
      case 'datetime':
        return value
      default:
        return value
    }
  }

  private toContentfulId(label: string): string {
    return label.toLowerCase().replace(/\s+/g, '')
  }

  private extractAssetId(url: string): string {
    // Extract ID from URL or generate hash
    const match = url.match(/\/([^\/]+)\.(jpg|jpeg|png|gif|webp|svg)/)
    return match ? match[1] : this.hashString(url)
  }

  private hashString(str: string): string {
    let hash = 0
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i)
      hash = ((hash << 5) - hash) + char
      hash = hash & hash
    }
    return Math.abs(hash).toString(36)
  }
}
```

### Drop-in Client

```typescript
// src/clients/contentful.ts

import { ContentfulTransformer, ContentfulEntry } from '../formats/contentful'

export interface ContentfulClientConfig {
  space: string
  accessToken: string
  apiUrl: string  // ModulaCMS API URL
}

export class ContentfulClient {
  private transformer: ContentfulTransformer
  private apiUrl: string
  private space: string

  constructor(config: ContentfulClientConfig) {
    this.transformer = new ContentfulTransformer()
    this.apiUrl = config.apiUrl
    this.space = config.space
  }

  /**
   * Get entry by ID (Contentful-compatible API)
   */
  async getEntry(entryId: string): Promise<ContentfulEntry> {
    const response = await fetch(`${this.apiUrl}/content/${entryId}`)
    const data = await response.json()
    return this.transformer.transform(data, this.space)
  }

  /**
   * Get entries with query (Contentful-compatible API)
   */
  async getEntries(query?: {
    content_type?: string
    limit?: number
    skip?: number
    order?: string
  }): Promise<{ items: ContentfulEntry[] }> {
    const params = new URLSearchParams()
    if (query?.content_type) params.append('content_type', query.content_type)
    if (query?.limit) params.append('limit', query.limit.toString())
    if (query?.skip) params.append('skip', query.skip.toString())
    if (query?.order) params.append('order', query.order)

    const response = await fetch(`${this.apiUrl}/content?${params}`)
    const data = await response.json()

    return {
      items: data.items.map((item: any) => this.transformer.transform({ root: item }, this.space))
    }
  }
}

/**
 * Create Contentful-compatible client
 */
export function createClient(config: ContentfulClientConfig): ContentfulClient {
  return new ContentfulClient(config)
}
```

### Usage (Drop-in Replacement)

```typescript
// Before (using Contentful)
import { createClient } from 'contentful'

const client = createClient({
  space: 'your_space',
  accessToken: 'your_token'
})

const entry = await client.getEntry('42')
console.log(entry.fields.title)

// After (using ModulaCMS with Contentful transformer)
import { createClient } from '@modulacms/contentful'

const client = createClient({
  space: 'modulacms',
  accessToken: 'not_used',
  apiUrl: 'https://api.modulacms.com'
})

const entry = await client.getEntry('42')
console.log(entry.fields.title)  // Same code!
```

---

## Transformer: Sanity Format

### Sanity JSON Structure

```json
{
  "_id": "42",
  "_type": "blogPost",
  "_createdAt": "2026-01-15T10:30:00Z",
  "_updatedAt": "2026-01-16T14:22:00Z",
  "_rev": "v1",
  "title": "Why ModulaCMS is Different",
  "slug": {
    "current": "why-modulacms-is-different",
    "_type": "slug"
  },
  "body": [
    {
      "_type": "block",
      "children": [
        { "_type": "span", "text": "ModulaCMS takes a different approach..." }
      ]
    }
  ],
  "featuredImage": {
    "_type": "image",
    "asset": {
      "_ref": "image-asset123",
      "_type": "reference"
    }
  },
  "published": true
}
```

### Implementation

```typescript
// src/formats/sanity.ts

import { ModulaCMSNode, ModulaCMSRoot } from '../core/types'

export interface SanityDocument {
  _id: string
  _type: string
  _createdAt: string
  _updatedAt: string
  _rev?: string
  [key: string]: any
}

export class SanityTransformer {

  transform(data: ModulaCMSRoot): SanityDocument {
    const node = data.root

    const doc: SanityDocument = {
      _id: node.datatype.content.content_data_id.toString(),
      _type: this.toSanityType(node.datatype.info.label),
      _createdAt: node.datatype.content.date_created || new Date().toISOString(),
      _updatedAt: node.datatype.content.date_modified || node.datatype.content.date_created || new Date().toISOString(),
      _rev: 'v1'
    }

    // Add fields
    for (const field of node.fields) {
      const key = this.fieldLabelToKey(field.info.label)
      const value = this.transformField(field)
      doc[key] = value
    }

    // Add child nodes as arrays
    if (node.nodes && node.nodes.length > 0) {
      const childType = this.pluralize(this.toSanityType(node.nodes[0].datatype.info.label))
      doc[childType] = node.nodes.map(child => this.transformNode(child))
    }

    return doc
  }

  private transformField(field: any): any {
    const type = field.info.type
    const value = field.content.field_value

    // Handle Sanity-specific field types
    if (field.info.label.toLowerCase().includes('slug')) {
      return {
        current: value,
        _type: 'slug'
      }
    }

    if (type === 'image' || type === 'asset') {
      return {
        _type: 'image',
        asset: {
          _ref: this.extractAssetRef(value),
          _type: 'reference'
        }
      }
    }

    if (type === 'markdown' || type === 'richtext') {
      // Convert markdown to Sanity portable text (simplified)
      return this.markdownToPortableText(value)
    }

    if (type === 'boolean') {
      return value === 'true'
    }

    if (type === 'number' || type === 'decimal') {
      return parseFloat(value)
    }

    if (type === 'integer') {
      return parseInt(value, 10)
    }

    return value
  }

  private markdownToPortableText(markdown: string): any[] {
    // Simplified conversion - in production, use proper markdown parser
    const paragraphs = markdown.split('\n\n')
    return paragraphs.map(para => ({
      _type: 'block',
      children: [
        {
          _type: 'span',
          text: para.replace(/^#+\s/, '') // Remove markdown headers
        }
      ]
    }))
  }

  private transformNode(node: ModulaCMSNode): SanityDocument {
    return this.transform({ root: node })
  }

  private fieldLabelToKey(label: string): string {
    return label
      .split(' ')
      .map((word, i) => i === 0 ? word.toLowerCase() : word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join('')
  }

  private toSanityType(label: string): string {
    return label.toLowerCase().replace(/\s+/g, '')
  }

  private pluralize(word: string): string {
    if (word.endsWith('s')) return word
    if (word.endsWith('y')) return word.slice(0, -1) + 'ies'
    return word + 's'
  }

  private extractAssetRef(url: string): string {
    const match = url.match(/\/([^\/]+)\.(jpg|jpeg|png|gif|webp|svg)/)
    return match ? `image-${match[1]}` : `image-${this.hashString(url)}`
  }

  private hashString(str: string): string {
    let hash = 0
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i)
      hash = ((hash << 5) - hash) + char
      hash = hash & hash
    }
    return Math.abs(hash).toString(36)
  }
}
```

### Drop-in Client

```typescript
// src/clients/sanity.ts

import { SanityTransformer, SanityDocument } from '../formats/sanity'

export interface SanityClientConfig {
  projectId: string
  dataset: string
  apiVersion?: string
  apiUrl: string  // ModulaCMS API URL
}

export class SanityClient {
  private transformer: SanityTransformer
  private apiUrl: string

  constructor(config: SanityClientConfig) {
    this.transformer = new SanityTransformer()
    this.apiUrl = config.apiUrl
  }

  /**
   * Fetch documents using GROQ-like syntax (simplified)
   */
  async fetch(query: string): Promise<SanityDocument[]> {
    // Parse simple GROQ queries
    // *[_type == "blogPost"] → /content?type=blogPost
    const typeMatch = query.match(/_type\s*==\s*["']([^"']+)["']/)
    const type = typeMatch ? typeMatch[1] : null

    const url = type
      ? `${this.apiUrl}/content?type=${type}`
      : `${this.apiUrl}/content`

    const response = await fetch(url)
    const data = await response.json()

    if (Array.isArray(data.items)) {
      return data.items.map((item: any) => this.transformer.transform({ root: item }))
    }

    return [this.transformer.transform(data)]
  }

  /**
   * Get document by ID
   */
  async getDocument(id: string): Promise<SanityDocument> {
    const response = await fetch(`${this.apiUrl}/content/${id}`)
    const data = await response.json()
    return this.transformer.transform(data)
  }
}

/**
 * Create Sanity-compatible client
 */
export function createClient(config: SanityClientConfig): SanityClient {
  return new SanityClient(config)
}
```

### Usage (Drop-in Replacement)

```typescript
// Before (using Sanity)
import { createClient } from '@sanity/client'

const client = createClient({
  projectId: 'your_project',
  dataset: 'production',
  apiVersion: '2024-01-01'
})

const posts = await client.fetch(`*[_type == "blogPost"]`)

// After (using ModulaCMS with Sanity transformer)
import { createClient } from '@modulacms/sanity'

const client = createClient({
  projectId: 'modulacms',
  dataset: 'production',
  apiVersion: '2024-01-01',
  apiUrl: 'https://api.modulacms.com'
})

const posts = await client.fetch(`*[_type == "blogPost"]`)  // Same code!
```

---

## Transformer: Strapi Format

### Strapi JSON Structure

```json
{
  "data": {
    "id": 42,
    "attributes": {
      "title": "Why ModulaCMS is Different",
      "slug": "why-modulacms-is-different",
      "body": "# Introduction\n\nModulaCMS takes a different approach...",
      "published": true,
      "createdAt": "2026-01-15T10:30:00Z",
      "updatedAt": "2026-01-16T14:22:00Z",
      "featuredImage": {
        "data": {
          "id": 123,
          "attributes": {
            "url": "https://cdn.example.com/hero.jpg",
            "name": "hero.jpg"
          }
        }
      }
    }
  },
  "meta": {}
}
```

### Implementation

```typescript
// src/formats/strapi.ts

import { ModulaCMSNode, ModulaCMSRoot } from '../core/types'

export interface StrapiResponse {
  data: StrapiEntry | StrapiEntry[]
  meta: Record<string, any>
}

export interface StrapiEntry {
  id: number
  attributes: Record<string, any>
}

export class StrapiTransformer {

  transform(data: ModulaCMSRoot): StrapiResponse {
    const node = data.root

    return {
      data: {
        id: node.datatype.content.content_data_id,
        attributes: this.extractAttributes(node)
      },
      meta: {}
    }
  }

  transformMany(items: ModulaCMSRoot[]): StrapiResponse {
    return {
      data: items.map(item => ({
        id: item.root.datatype.content.content_data_id,
        attributes: this.extractAttributes(item.root)
      })),
      meta: {
        pagination: {
          page: 1,
          pageSize: items.length,
          pageCount: 1,
          total: items.length
        }
      }
    }
  }

  private extractAttributes(node: ModulaCMSNode): Record<string, any> {
    const attributes: Record<string, any> = {}

    // Add timestamps
    attributes.createdAt = node.datatype.content.date_created || new Date().toISOString()
    attributes.updatedAt = node.datatype.content.date_modified || attributes.createdAt

    // Add fields
    for (const field of node.fields) {
      const key = this.fieldLabelToKey(field.info.label)
      const value = this.transformField(field)
      attributes[key] = value
    }

    // Add relations (child nodes)
    if (node.nodes && node.nodes.length > 0) {
      const childType = this.pluralize(this.fieldLabelToKey(node.nodes[0].datatype.info.label))
      attributes[childType] = {
        data: node.nodes.map(child => ({
          id: child.datatype.content.content_data_id,
          attributes: this.extractAttributes(child)
        }))
      }
    }

    return attributes
  }

  private transformField(field: any): any {
    const type = field.info.type
    const value = field.content.field_value

    // Handle media fields (Strapi format)
    if (type === 'image' || type === 'asset') {
      const filename = value.split('/').pop() || 'file'
      return {
        data: {
          id: this.hashString(value),
          attributes: {
            url: value,
            name: filename,
            alternativeText: field.info.label
          }
        }
      }
    }

    // Type coercion
    if (type === 'boolean') return value === 'true'
    if (type === 'number' || type === 'decimal') return parseFloat(value)
    if (type === 'integer') return parseInt(value, 10)

    return value
  }

  private fieldLabelToKey(label: string): string {
    return label
      .split(' ')
      .map((word, i) => i === 0 ? word.toLowerCase() : word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join('')
  }

  private pluralize(word: string): string {
    if (word.endsWith('s')) return word
    if (word.endsWith('y')) return word.slice(0, -1) + 'ies'
    return word + 's'
  }

  private hashString(str: string): string {
    let hash = 0
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i)
      hash = ((hash << 5) - hash) + char
      hash = hash & hash
    }
    return Math.abs(hash)
  }
}
```

---

## Transformer: WordPress REST API Format

### WordPress JSON Structure

```json
{
  "id": 42,
  "date": "2026-01-15T10:30:00",
  "date_gmt": "2026-01-15T10:30:00",
  "modified": "2026-01-16T14:22:00",
  "modified_gmt": "2026-01-16T14:22:00",
  "slug": "why-modulacms-is-different",
  "status": "publish",
  "type": "post",
  "link": "https://example.com/blog/why-modulacms-is-different",
  "title": {
    "rendered": "Why ModulaCMS is Different"
  },
  "content": {
    "rendered": "<h1>Introduction</h1><p>ModulaCMS takes a different approach...</p>",
    "protected": false
  },
  "excerpt": {
    "rendered": "<p>ModulaCMS takes a different approach...</p>",
    "protected": false
  },
  "author": 1,
  "featured_media": 123,
  "comment_status": "open",
  "ping_status": "open",
  "meta": {},
  "acf": {
    "featured_image": "https://cdn.example.com/hero.jpg",
    "published": true
  }
}
```

### Implementation

```typescript
// src/formats/wordpress.ts

import { ModulaCMSNode, ModulaCMSRoot } from '../core/types'
import { marked } from 'marked'

export interface WordPressPost {
  id: number
  date: string
  date_gmt: string
  modified: string
  modified_gmt: string
  slug: string
  status: string
  type: string
  link: string
  title: { rendered: string }
  content: { rendered: string; protected: boolean }
  excerpt: { rendered: string; protected: boolean }
  author: number
  featured_media: number
  comment_status: string
  ping_status: string
  meta: Record<string, any>
  acf?: Record<string, any>
}

export class WordPressTransformer {
  private siteUrl: string

  constructor(siteUrl: string = 'https://example.com') {
    this.siteUrl = siteUrl
  }

  transform(data: ModulaCMSRoot): WordPressPost {
    const node = data.root

    const fields = this.extractFields(node)
    const slug = fields.slug || this.generateSlug(fields.title || '')
    const type = this.toWordPressType(node.datatype.info.label)

    return {
      id: node.datatype.content.content_data_id,
      date: node.datatype.content.date_created || new Date().toISOString(),
      date_gmt: node.datatype.content.date_created || new Date().toISOString(),
      modified: node.datatype.content.date_modified || node.datatype.content.date_created || new Date().toISOString(),
      modified_gmt: node.datatype.content.date_modified || node.datatype.content.date_created || new Date().toISOString(),
      slug: slug,
      status: fields.published ? 'publish' : 'draft',
      type: type,
      link: `${this.siteUrl}/${type}/${slug}`,
      title: {
        rendered: fields.title || ''
      },
      content: {
        rendered: this.renderContent(fields.body || fields.content || ''),
        protected: false
      },
      excerpt: {
        rendered: this.generateExcerpt(fields.body || fields.content || fields.excerpt || ''),
        protected: false
      },
      author: node.datatype.content.author_id,
      featured_media: fields.featuredImage ? this.hashString(fields.featuredImage) : 0,
      comment_status: 'open',
      ping_status: 'open',
      meta: {},
      acf: this.extractACF(fields)
    }
  }

  private extractFields(node: ModulaCMSNode): Record<string, any> {
    const fields: Record<string, any> = {}

    for (const field of node.fields) {
      const key = this.fieldLabelToKey(field.info.label)
      const value = this.parseFieldValue(field.content.field_value, field.info.type)
      fields[key] = value
    }

    return fields
  }

  private extractACF(fields: Record<string, any>): Record<string, any> {
    // Put custom fields in ACF object (Advanced Custom Fields format)
    const acf: Record<string, any> = {}
    const standardFields = ['title', 'slug', 'body', 'content', 'excerpt', 'published']

    for (const [key, value] of Object.entries(fields)) {
      if (!standardFields.includes(key)) {
        acf[key] = value
      }
    }

    return acf
  }

  private renderContent(content: string): string {
    // Convert markdown to HTML if needed
    if (content.includes('#') || content.includes('**')) {
      return marked(content)
    }
    return `<p>${content}</p>`
  }

  private generateExcerpt(content: string, length: number = 150): string {
    const text = content.replace(/#/g, '').replace(/\*\*/g, '').trim()
    const excerpt = text.substring(0, length)
    return `<p>${excerpt}${text.length > length ? '...' : ''}</p>`
  }

  private generateSlug(title: string): string {
    return title
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '')
  }

  private toWordPressType(label: string): string {
    const type = label.toLowerCase().replace(/\s+/g, '_')
    if (type.includes('post') || type.includes('article')) return 'post'
    if (type.includes('page')) return 'page'
    return type
  }

  private fieldLabelToKey(label: string): string {
    return label
      .split(' ')
      .map((word, i) => i === 0 ? word.toLowerCase() : word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join('')
  }

  private parseFieldValue(value: string, type: string): any {
    if (type === 'boolean') return value === 'true'
    if (type === 'number' || type === 'decimal') return parseFloat(value)
    if (type === 'integer') return parseInt(value, 10)
    return value
  }

  private hashString(str: string): number {
    let hash = 0
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i)
      hash = ((hash << 5) - hash) + char
      hash = hash & hash
    }
    return Math.abs(hash)
  }
}
```

---

## Usage Examples

### Example 1: Migrate from Contentful

```typescript
// Old Contentful code (no changes needed!)
import { createClient } from 'contentful'

const client = createClient({
  space: 'your_space',
  accessToken: 'your_token'
})

const posts = await client.getEntries({ content_type: 'blogPost' })

posts.items.forEach(post => {
  console.log(post.fields.title)
  console.log(post.fields.body)
})
```

```typescript
// Switch to ModulaCMS (only change import!)
import { createClient } from '@modulacms/contentful'

const client = createClient({
  space: 'modulacms',
  accessToken: 'not_needed',
  apiUrl: 'https://api.modulacms.com'
})

const posts = await client.getEntries({ content_type: 'blogPost' })

posts.items.forEach(post => {
  console.log(post.fields.title)  // Same code!
  console.log(post.fields.body)   // Same code!
})
```

**Result:** Zero frontend changes, just swap the import!

---

### Example 2: Migrate from Sanity

```typescript
// Old Sanity code
import { createClient } from '@sanity/client'

const client = createClient({
  projectId: 'abc123',
  dataset: 'production'
})

const posts = await client.fetch(`*[_type == "blogPost"]{ title, body }`)
```

```typescript
// Switch to ModulaCMS (only change import!)
import { createClient } from '@modulacms/sanity'

const client = createClient({
  projectId: 'modulacms',
  dataset: 'production',
  apiUrl: 'https://api.modulacms.com'
})

const posts = await client.fetch(`*[_type == "blogPost"]{ title, body }`)  // Same!
```

---

### Example 3: Migrate from WordPress

```typescript
// Old WordPress REST API code
const response = await fetch('https://example.com/wp-json/wp/v2/posts/42')
const post = await response.json()

console.log(post.title.rendered)
console.log(post.content.rendered)
console.log(post.acf.custom_field)
```

```typescript
// Switch to ModulaCMS
import { WordPressClient } from '@modulacms/wordpress'

const client = new WordPressClient({
  siteUrl: 'https://example.com',
  apiUrl: 'https://api.modulacms.com'
})

const post = await client.getPost(42)

console.log(post.title.rendered)     // Same!
console.log(post.content.rendered)   // Same!
console.log(post.acf.custom_field)   // Same!
```

---

## Marketing: The Killer Feature

### Headline

> **"Switch from Contentful to ModulaCMS without changing a single line of frontend code."**

### How It Works

1. **Install transformer package**
   ```bash
   npm install @modulacms/contentful
   ```

2. **Change one line** (the import)
   ```typescript
   // Before
   import { createClient } from 'contentful'

   // After
   import { createClient } from '@modulacms/contentful'
   ```

3. **Update config** (point to ModulaCMS API)
   ```typescript
   const client = createClient({
     space: 'modulacms',
     accessToken: 'your_modulacms_token',
     apiUrl: 'https://api.modulacms.com'  // Add this
   })
   ```

4. **Everything else stays the same**
   - Same method calls: `client.getEntry()`, `client.getEntries()`
   - Same field access: `entry.fields.title`
   - Same data structure
   - Zero frontend refactoring

### Why This Matters

**Traditional CMS migration:**
- Weeks of frontend refactoring
- High risk of bugs
- Expensive developer time
- Rewrite components, utilities, hooks

**ModulaCMS migration with transformers:**
- Hours, not weeks
- Low risk (same API)
- Cheap (minimal developer time)
- No frontend changes

**Result:** "Try ModulaCMS without risk"

---

## Implementation Plan

### Phase 1: Core Transformers (4 weeks)

**Package:** `@modulacms/transform`

**Deliverables:**
- Base transformer class
- Contentful transformer
- Sanity transformer
- Strapi transformer
- WordPress transformer
- Clean transformer (ModulaCMS native)
- Full test suite

---

### Phase 2: Client Libraries (4 weeks)

**Packages:**
- `@modulacms/contentful` - Drop-in Contentful replacement
- `@modulacms/sanity` - Drop-in Sanity replacement
- `@modulacms/strapi` - Drop-in Strapi replacement
- `@modulacms/wordpress` - Drop-in WordPress replacement

**Each includes:**
- Client class matching original API
- TypeScript types
- Documentation
- Migration guide
- Examples

---

### Phase 3: Advanced Features (4 weeks)

**GraphQL support:**
- `@modulacms/graphql` - GraphQL schema generation
- Apollo Server integration
- Type codegen

**Framework integrations:**
- `@modulacms/next-contentful` - Next.js + Contentful format
- `@modulacms/next-sanity` - Next.js + Sanity format

**CLI tools:**
- `modulacms-migrate contentful` - Automated migration
- `modulacms-migrate sanity`
- Content import/export

---

### Phase 4: Community & Ecosystem (Ongoing)

**Open source strategy:**
- Accept community transformers
- Shopify transformer
- Drupal transformer
- Ghost transformer
- Custom CMS transformers

**Marketplace:**
- Pre-built migration scripts
- Professional migration services
- Schema mapping tools

---

## Package API Design

### Installation

```bash
# Core transformer package
npm install @modulacms/transform

# Or use format-specific packages
npm install @modulacms/contentful
npm install @modulacms/sanity
npm install @modulacms/strapi
npm install @modulacms/wordpress
```

### Core API

```typescript
import {
  ContentfulTransformer,
  SanityTransformer,
  StrapiTransformer,
  WordPressTransformer,
  CleanTransformer
} from '@modulacms/transform'

// Transform to Contentful format
const contentful = new ContentfulTransformer()
const entry = contentful.transform(modulaCMSData)

// Transform to Sanity format
const sanity = new SanityTransformer()
const doc = sanity.transform(modulaCMSData)

// Transform to clean format
const clean = new CleanTransformer()
const obj = clean.transform(modulaCMSData)
```

### Client API

```typescript
// Contentful-compatible client
import { createClient } from '@modulacms/contentful'

const client = createClient({
  space: 'modulacms',
  accessToken: 'token',
  apiUrl: 'https://api.modulacms.com'
})

await client.getEntry('id')
await client.getEntries({ content_type: 'blogPost' })

// Sanity-compatible client
import { createClient } from '@modulacms/sanity'

const client = createClient({
  projectId: 'modulacms',
  dataset: 'production',
  apiUrl: 'https://api.modulacms.com'
})

await client.fetch(`*[_type == "blogPost"]`)
await client.getDocument('id')
```

---

## Testing Strategy

### Unit Tests

Test each transformer independently:

```typescript
describe('ContentfulTransformer', () => {
  it('transforms ModulaCMS data to Contentful format', () => {
    const transformer = new ContentfulTransformer()
    const result = transformer.transform(modulaCMSData)

    expect(result.sys.id).toBe('42')
    expect(result.sys.type).toBe('Entry')
    expect(result.fields.title).toBe('Why ModulaCMS is Different')
  })

  it('handles nested nodes', () => {
    // Test child node transformation
  })

  it('handles different field types', () => {
    // Test boolean, number, date, etc.
  })
})
```

### Integration Tests

Test against real Contentful/Sanity SDKs:

```typescript
describe('Contentful Client Compatibility', () => {
  it('matches Contentful client API', async () => {
    const modulaClient = createModulaContentfulClient(config)
    const result = await modulaClient.getEntry('42')

    // Verify structure matches Contentful
    expect(result).toHaveProperty('sys')
    expect(result).toHaveProperty('fields')
    expect(result.sys).toHaveProperty('contentType')
  })
})
```

### Comparison Tests

Compare output with actual CMS responses:

```typescript
describe('Format Compatibility', () => {
  it('matches Contentful response structure', () => {
    const actualContentfulResponse = /* real Contentful data */
    const modulaCMSTransformed = transformer.transform(modulaCMSData)

    // Should have same structure
    expect(Object.keys(modulaCMSTransformed)).toEqual(Object.keys(actualContentfulResponse))
  })
})
```

---

## Documentation Strategy

### For Each Transformer

**README includes:**
1. What it does
2. Installation
3. Quick start (5 lines of code)
4. API reference
5. Migration guide
6. Field mapping table
7. Known limitations
8. Examples

### Migration Guides

**Step-by-step:**
1. Install package
2. Update imports
3. Update config
4. Test locally
5. Deploy

**With code diffs:**
```diff
- import { createClient } from 'contentful'
+ import { createClient } from '@modulacms/contentful'

  const client = createClient({
    space: 'your_space',
    accessToken: 'your_token',
+   apiUrl: 'https://api.modulacms.com'
  })
```

---

## Conclusion

### The Vision

**ModulaCMS becomes the easiest CMS to migrate TO:**
- Switch from Contentful: 1 hour
- Switch from Sanity: 1 hour
- Switch from Strapi: 1 hour
- Switch from WordPress: 2 hours

**No other CMS offers this.**

### The Strategy

**Backend:** Generic, flexible ModulaCMS
**Transformers:** Match any CMS format
**Result:** Drop-in replacement for any popular CMS

### The Impact

**Adoption barrier eliminated:**
- "Try ModulaCMS without rewriting frontend"
- "Switch from Contentful in production, zero downtime"
- "Keep your Contentful code, use ModulaCMS backend"

**Marketing gold:**
> "The only CMS you can adopt without changing your frontend code."

### Next Steps

1. Build `@modulacms/transform` core package
2. Build `@modulacms/contentful` (highest priority - most users)
3. Build `@modulacms/sanity` (second priority - popular with agencies)
4. Document migration paths
5. Create migration tools
6. Measure adoption impact

**This is a game-changer for adoption.**

---

**Last Updated:** 2026-01-16
