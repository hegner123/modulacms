# Serving Your Frontend

Connect a frontend framework to ModulaCMS and render your content as real pages.

## The pattern

Every frontend integration follows the same pattern:

1. Create a catch-all route in your framework.
2. Extract the slug from the URL.
3. Call the ModulaCMS content delivery API with that slug.
4. Handle the response: render content, follow redirects, or show a 404.

The content delivery endpoint is public and requires no authentication:

```
GET /api/v1/content/{slug}
```

## Next.js (App Router)

Create a catch-all route at `app/[[...slug]]/page.tsx`:

```typescript
import { ModulaClient } from '@modulacms/sdk'

const client = new ModulaClient({
  baseUrl: process.env.CMS_URL!,
  apiKey: process.env.CMS_API_KEY!,
})

export default async function Page({ params }: { params: { slug?: string[] } }) {
  const slug = params.slug?.join('/') || 'homepage'

  try {
    const page = await client.getPage(slug, { format: 'clean' })
    return <ContentRenderer content={page} />
  } catch (err) {
    if (isNotFound(err)) {
      return notFound()
    }
    throw err
  }
}
```

> **Good to know**: The `[[...slug]]` syntax in Next.js matches both `/` (no slug segments) and `/about/team` (multiple segments). Map the empty case to your homepage slug.

## Nuxt

Create a catch-all route at `pages/[...slug].vue`:

```vue
<script setup lang="ts">
const route = useRoute()
const slug = (route.params.slug as string[])?.join('/') || 'homepage'

const { data: page, error } = await useFetch(
  `${useRuntimeConfig().public.cmsUrl}/api/v1/content/${slug}?format=clean`
)

if (error.value?.statusCode === 404) {
  throw createError({ statusCode: 404, statusMessage: 'Page Not Found' })
}
</script>

<template>
  <ContentRenderer :content="page" />
</template>
```

## SvelteKit

Create a catch-all route at `src/routes/[...slug]/+page.server.ts`:

```typescript
import { error } from '@sveltejs/kit'
import type { PageServerLoad } from './$types'

export const load: PageServerLoad = async ({ params, fetch }) => {
  const slug = params.slug || 'homepage'
  const res = await fetch(
    `${process.env.CMS_URL}/api/v1/content/${slug}?format=clean`
  )

  if (res.status === 404) {
    throw error(404, 'Page not found')
  }

  return { page: await res.json() }
}
```

## Fetch content by slug

The content delivery endpoint returns the full published content tree for a route.

**curl:**

```bash
curl http://localhost:8080/api/v1/content/homepage
```

**Go SDK:**

```go
import modula "github.com/hegner123/modulacms/sdks/go"

client, err := modula.NewClient(modula.ClientConfig{
    BaseURL: "http://localhost:8080",
    APIKey:  "YOUR_API_KEY",
})
if err != nil {
    // handle error
}

raw, err := client.Content.GetPage(ctx, "homepage", "", "")
if err != nil {
    if modula.IsNotFound(err) {
        // 404 -- no published content at this slug
    }
    // handle error
}
```

**TypeScript SDK:**

```typescript
import { ModulaClient, isNotFound } from '@modulacms/sdk'

const client = new ModulaClient({
  baseUrl: 'https://cms.example.com',
  apiKey: 'YOUR_API_KEY',
})

try {
  const page = await client.getPage('homepage')
} catch (err) {
  if (isNotFound(err)) {
    // 404
  }
}
```

### Choose an output format

The `format` parameter controls the response shape. Use the format that best fits your frontend:

| Format | Description |
|--------|-------------|
| `raw` | Native ModulaCMS tree structure (default) |
| `clean` | Simplified structure with flat field values |
| `contentful` | Contentful-compatible response format |
| `sanity` | Sanity.io-compatible response format |
| `strapi` | Strapi-compatible response format |
| `wordpress` | WordPress REST API-compatible response format |

```bash
curl "http://localhost:8080/api/v1/content/homepage?format=clean"
```

```typescript
const page = await client.getPage('homepage', { format: 'clean' })
```

```go
raw, err := client.Content.GetPage(ctx, "homepage", "clean", "")
```

> **Good to know**: If you're migrating from another CMS, use the matching output format so your existing frontend code works with minimal changes.

### Request localized content

Pass a locale code to get translated content. When a translation doesn't exist, ModulaCMS falls back along the locale's fallback chain (e.g., `fr-CA` falls back to `fr`, then to the default locale).

```bash
curl "http://localhost:8080/api/v1/content/homepage?locale=fr"
```

```typescript
const page = await client.getPage('homepage', { format: 'clean', locale: 'fr' })
```

```go
raw, err := client.Content.GetPage(ctx, "homepage", "clean", "fr")
```

## Build navigation from routes

Fetch all active routes to build your site's navigation menu.

**TypeScript SDK:**

```typescript
const routes = await client.listRoutes()

const nav = routes
  .filter(r => r.status === 1)
  .map(r => ({
    label: r.title,
    href: `/${r.slug}`,
  }))
```

**Go SDK:**

```go
routes, err := client.Routes.List(ctx)
if err != nil {
    // handle error
}

type NavItem struct {
    Label string
    Href  string
}

var nav []NavItem
for _, r := range routes {
    if r.Status != 1 {
        continue
    }
    nav = append(nav, NavItem{
        Label: r.Title,
        Href:  "/" + string(r.Slug),
    })
}
```

> **Good to know**: Build nested navigation by using slug conventions. Slugs like `blog`, `blog/tutorials`, and `blog/news` imply a hierarchy you can parse on the client. See [routing](routing.md) for a full hierarchical nav example.

## Search and filter content

Use the query endpoint to list, search, and filter content by datatype. This is how you build blog indexes, product catalogs, and filtered content lists.

```bash
curl "http://localhost:8080/api/v1/query/blog-post?sort=-published_at&limit=10"
```

**TypeScript SDK:**

```typescript
const result = await client.queryContent('blog-post', {
  sort: '-published_at',
  limit: 10,
})

for (const item of result.data) {
  console.log(`${item.fields.title} (${item.published_at})`)
}
```

**Go SDK:**

```go
result, err := client.Query.Query(ctx, "blog-post", &modula.QueryParams{
    Sort:  "-published_at",
    Limit: 10,
})
if err != nil {
    // handle error
}

for _, item := range result.Data {
    fmt.Printf("%s (%s)\n", item.Fields["title"], item.PublishedAt)
}
```

Filter by field values:

```typescript
const tutorials = await client.queryContent('blog-post', {
  filters: { category: 'tutorials' },
  sort: '-published_at',
  limit: 20,
})
```

Pattern match with the `like` operator:

```typescript
const results = await client.queryContent('blog-post', {
  filters: { 'title[like]': '%tutorial%' },
})
```

> **Good to know**: See [querying content](querying.md) for the full filter syntax, all operators, and pagination patterns.

## Handle errors

All SDKs return typed errors you can inspect for HTTP status codes.

**TypeScript SDK:**

```typescript
import { isNotFound, isUnauthorized } from '@modulacms/sdk'

try {
  const page = await client.getPage('nonexistent')
} catch (err) {
  if (isNotFound(err)) {
    // 404 -- render your 404 page
  }
  if (isUnauthorized(err)) {
    // 401 -- invalid or missing API key
  }
}
```

**Go SDK:**

```go
raw, err := client.Content.GetPage(ctx, "nonexistent", "", "")
if err != nil {
    if modula.IsNotFound(err) {
        // 404
    }
    if modula.IsUnauthorized(err) {
        // 401
    }
    var apiErr *modula.ApiError
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

## Next steps

- [Routing](routing.md) -- create routes and configure output formats
- [Querying content](querying.md) -- filter, sort, and paginate content by datatype
- [Media](media.md) -- serve responsive images in your frontend
