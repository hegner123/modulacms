# Read-Only SDK

Fetch published content from frontend applications using the lightweight read-only `@modulacms/sdk` client.

## ModulaClient

Create an instance with `new ModulaClient(config)`:

```typescript
import { ModulaClient } from '@modulacms/sdk'

const client = new ModulaClient({
  baseUrl: 'https://cms.example.com',
  apiKey: 'your-read-only-key',
})
```

All methods are async and return promises.

## getPage

The primary method for frontend content delivery. Resolves a route slug to a fully rendered content tree.

```typescript
const page = await client.getPage('about')
```

### GetPageOptions

Pass an options object as the second argument:

```typescript
const page = await client.getPage<MyPageType>('about', {
  format: 'clean',
  locale: 'fr',
  validate: isMyPageType,
})
```

| Option | Type | Description |
|--------|------|-------------|
| `format` | `ContentFormat` | Output format: `'contentful'`, `'sanity'`, `'strapi'`, `'wordpress'`, `'clean'`, `'raw'`. Overrides `defaultFormat` from config. |
| `locale` | `string` | Locale code (e.g. `'en'`, `'fr'`). Filters field values to the specified locale. |
| `validate` | `Validator<T>` | Runtime type guard for response validation. |

### Validator Pattern

The `Validator<T>` type is a type-narrowing function:

```typescript
type Validator<T> = (data: unknown) => data is T
```

When you provide a validator, `getPage` runs it against the parsed response and throws if validation fails. This gives you runtime safety on top of compile-time types.

```typescript
interface BlogPage {
  root: {
    datatype: { info: { label: string } }
    fields: Array<{ info: { name: string }; content: { field_value: string } }>
  }
}

function isBlogPage(data: unknown): data is BlogPage {
  return (
    typeof data === 'object' &&
    data !== null &&
    'root' in data
  )
}

const page = await client.getPage<BlogPage>('blog', {
  validate: isBlogPage,
})
// page is typed as BlogPage
```

## Content Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `getPage` | `<T>(slug: string, options?: GetPageOptions<T>) => Promise<T>` | Fetch rendered content tree by route slug. |
| `listRoutes` | `() => Promise<Route[]>` | List all public routes. |
| `getRoute` | `(id: string) => Promise<Route>` | Get a single route by ID. |
| `listContentData` | `() => Promise<ContentData[]>` | List all public content data nodes. |
| `getContentData` | `(id: string) => Promise<ContentData>` | Get a single content data node. |
| `listContentFields` | `() => Promise<ContentField[]>` | List all public content field values. |
| `getContentField` | `(id: string) => Promise<ContentField>` | Get a single content field value. |
| `listMedia` | `() => Promise<Media[]>` | List all media assets. |
| `getMedia` | `(id: string) => Promise<Media>` | Get a single media asset. |
| `listMediaPaginated` | `(params: PaginationParams) => Promise<PaginatedResponse<Media>>` | List media with pagination. |
| `listMediaDimensions` | `() => Promise<MediaDimension[]>` | List media dimension presets. |
| `getMediaDimension` | `(id: string) => Promise<MediaDimension>` | Get a single media dimension preset. |
| `listDatatypes` | `() => Promise<Datatype[]>` | List all datatype definitions. |
| `getDatatype` | `(id: string) => Promise<Datatype>` | Get a single datatype definition. |
| `listFields` | `() => Promise<Field[]>` | List all field definitions. |
| `getField` | `(id: string) => Promise<Field>` | Get a single field definition. |
| `queryContent` | `(datatype: string, params?: QueryParams) => Promise<QueryResult>` | Query content items by datatype name. |

## queryContent

Query content by datatype with filtering, sorting, and pagination:

```typescript
const result = await client.queryContent('blog-post', {
  sort: '-published_at',
  limit: 10,
  offset: 0,
  locale: 'en',
  status: 'published',
  filters: {
    category: 'news',
    'price[gte]': '10',
  },
})

console.log(`${result.data.length} of ${result.total} items`)
```

### QueryParams

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sort` | `string` | -- | Sort field. Prefix with `-` for descending. |
| `limit` | `number` | `20` | Max items to return (server max: 100). |
| `offset` | `number` | `0` | Items to skip for pagination. |
| `locale` | `string` | -- | Locale code for i18n content. |
| `status` | `string` | `'published'` | Content status filter. |
| `filters` | `Record<string, string>` | -- | Field filters with optional operator suffixes. |

Filter operators: `[eq]`, `[ne]`, `[gt]`, `[gte]`, `[lt]`, `[lte]`, `[like]`, `[in]`. A bare field name is treated as `[eq]`.

## Content Tree Structure

`getPage` returns a content tree in this structure (when you don't specify a format override):

```typescript
type ContentTree = {
  root: ContentNode
}

type ContentNode = {
  datatype: {
    info: Datatype      // schema definition
    content: ContentData // tree pointers, status, dates
  }
  fields: Array<{
    info: Field          // field definition (name, type, validation)
    content: ContentField // stored value
  }>
  nodes?: ContentNode[] // child nodes
}
```

Each node pairs its schema (`info`) with its content instance (`content`). Walk `nodes` recursively to traverse the tree.

## Framework Examples

### Next.js (App Router)

```typescript
import { ModulaClient } from '@modulacms/sdk'

const client = new ModulaClient({
  baseUrl: process.env.CMS_URL!,
  apiKey: process.env.CMS_API_KEY,
})

export default async function Page({ params }: { params: { slug: string } }) {
  const page = await client.getPage(params.slug)
  return <div>{JSON.stringify(page)}</div>
}
```

### Nuxt 3

```typescript
import { ModulaClient } from '@modulacms/sdk'

const client = new ModulaClient({
  baseUrl: useRuntimeConfig().cmsUrl,
})

const { data: page } = await useAsyncData('page', () =>
  client.getPage('home')
)
```

### SvelteKit

```typescript
import { ModulaClient } from '@modulacms/sdk'
import type { PageServerLoad } from './$types'

const client = new ModulaClient({
  baseUrl: import.meta.env.CMS_URL,
})

export const load: PageServerLoad = async ({ params }) => {
  const page = await client.getPage(params.slug)
  return { page }
}
```

## Error Handling

All `ModulaClient` methods throw `ModulaError` on failure. See [Error Handling](/docs/sdks/typescript/error-handling) for details.

```typescript
import { ModulaError } from '@modulacms/sdk'

try {
  const page = await client.getPage('nonexistent')
} catch (err) {
  if (err instanceof ModulaError) {
    console.error(`API error ${err.status}: ${err.message}`)
  }
}
```
