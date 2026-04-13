# TypeScript SDK -- Getting Started

Install and configure the ModulaCMS TypeScript SDK for content delivery or admin operations.

| Package | npm Name | Purpose |
|---------|----------|---------|
| `types/` | `@modulacms/types` | Shared entity types, branded IDs, enums, error types |
| `modulacms-sdk/` | `@modulacms/sdk` | Read-only content delivery for frontend apps |
| `modulacms-admin-sdk/` | `@modulacms/admin-sdk` | Full admin CRUD for management tools and CI/CD |

Both SDKs depend on `@modulacms/types`. Installing either SDK from npm pulls in `@modulacms/types` automatically.

## Installation

Install the read-only SDK for frontend content delivery:

```bash
npm install @modulacms/sdk
```

Install the admin SDK for full CRUD operations:

```bash
npm install @modulacms/admin-sdk
```

> **Good to know**: Both packages ship as dual ESM + CJS builds. TypeScript 5.7+ recommended. Node 22+ for server-side use; any modern browser for client-side.

## Creating a Read-Only Client

The read-only SDK provides `ModulaClient` for fetching published content. This is the client you use in frontend applications.

```typescript
import { ModulaClient } from '@modulacms/sdk'

const client = new ModulaClient({
  baseUrl: 'https://cms.example.com',
  apiKey: 'your-api-key',
})

// Fetch a page by slug
const page = await client.getPage('about')
```

### ModulaClientConfig

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `baseUrl` | `string` | Yes | -- | ModulaCMS server URL. |
| `apiKey` | `string` | No | -- | Bearer token for API authentication. |
| `defaultFormat` | `ContentFormat` | No | -- | Default output format for `getPage`. |
| `timeout` | `number` | No | -- | Request timeout in milliseconds. |
| `credentials` | `RequestCredentials` | No | -- | Fetch `credentials` mode for cookie handling. |

## Creating an Admin Client

The admin SDK provides `createAdminClient` for full read/write access to every API resource.

```typescript
import { createAdminClient } from '@modulacms/admin-sdk'

const client = createAdminClient({
  baseUrl: 'https://cms.example.com',
  apiKey: 'sk_live_abc123',
})

// List all users
const users = await client.users.list()

// Get the current user
const me = await client.auth.me()
```

### ClientConfig

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `baseUrl` | `string` | Yes | -- | ModulaCMS server URL. Must be HTTPS unless `allowInsecure` is true. |
| `apiKey` | `string` | No | -- | Bearer token for API authentication. Omit for cookie-based auth. |
| `defaultTimeout` | `number` | No | `30000` | Request timeout in milliseconds. |
| `credentials` | `RequestCredentials` | No | `'include'` | Fetch `credentials` mode for cookie handling. |
| `allowInsecure` | `boolean` | No | `false` | Allow `http://` URLs. Required for local development. |

`createAdminClient` validates the config at construction time and throws if:

- `baseUrl` is not a valid URL.
- `baseUrl` uses `http://` without `allowInsecure: true`.
- `apiKey` is an empty string (omit it instead).

## Local Development

For local development against an HTTP server:

```typescript
const client = createAdminClient({
  baseUrl: 'http://localhost:8080',
  allowInsecure: true,
})
```

## Server-Side Rendering

Both SDKs use the global `fetch` function. In Node.js 22+, this is built-in. For older Node versions or custom HTTP requirements, the read-only SDK accepts any `fetch`-compatible function via the config.

## First Requests

### Fetching a Page (Read-Only SDK)

```typescript
import { ModulaClient } from '@modulacms/sdk'

const client = new ModulaClient({
  baseUrl: 'https://cms.example.com',
})

// Get the content tree for a route slug
const page = await client.getPage('home')

// List all routes
const routes = await client.listRoutes()

// Query content by datatype
const posts = await client.queryContent('blog-post', {
  sort: '-published_at',
  limit: 10,
})
```

### Admin Operations (Admin SDK)

```typescript
import { createAdminClient } from '@modulacms/admin-sdk'

const client = createAdminClient({
  baseUrl: 'https://cms.example.com',
  apiKey: process.env.CMS_API_KEY,
})

// Authenticate
const loginResponse = await client.auth.login({
  email: 'admin@example.com' as any,
  password: 'password',
})

// CRUD on datatypes
const datatypes = await client.datatypes.list()
const dt = await client.datatypes.get(datatypeId)

// Upload media
const media = await client.mediaUpload.upload(file)

// Publish content
await client.publishing.publish({ content_data_id: contentId })
```

## Next Steps

- [Read-Only SDK](/docs/sdks/typescript/read-only-sdk) -- content delivery for frontend apps
- [Admin SDK](/docs/sdks/typescript/admin-sdk) -- full CRUD operations
- [Type Safety](/docs/sdks/typescript/type-safety) -- branded IDs and entity types
- [Error Handling](/docs/sdks/typescript/error-handling) -- error types and patterns
- [Reference](/docs/sdks/typescript/reference) -- quick reference tables
