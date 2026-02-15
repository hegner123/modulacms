# modulacms-admin-sdk

[![CI](https://github.com/hegner123/modulacms-admin-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/hegner123/modulacms-admin-sdk/actions/workflows/ci.yml)

TypeScript SDK for the ModulaCMS admin API. Provides fully typed CRUD operations for all admin resources, authentication, media uploads, content tree retrieval, and bulk imports from external CMS platforms.

## Features

- Zero runtime dependencies -- built on native `fetch`
- Branded ID types prevent mixing incompatible identifiers at compile time
- HTTPS enforced by default; `allowInsecure` flag for local development
- 30s default timeout with per-request `AbortSignal` support
- ESM-only, targets Node.js 18+ and modern browsers

## Install

```bash
bun add modulacms-admin-sdk
```

## Quick Start

```ts
import { createAdminClient, isApiError } from 'modulacms-admin-sdk'

const client = createAdminClient({
  baseUrl: 'https://cms.example.com',
  apiKey: process.env.CMS_API_KEY,
})

// Authentication
const me = await client.auth.me()

// CRUD operations
const users = await client.users.list()
const route = await client.adminRoutes.get(slug)

// Media upload
const media = await client.mediaUpload.upload(file)

// Content tree
const tree = await client.adminTree.get(slug)

// Bulk import
const result = await client.import.contentful(exportData)
```

## Configuration

```ts
import type { ClientConfig } from 'modulacms-admin-sdk'

const config: ClientConfig = {
  baseUrl: 'https://cms.example.com', // Required. HTTPS enforced unless allowInsecure is set.
  apiKey: 'sk_live_abc123',           // Optional. Bearer token for server-side auth.
  defaultTimeout: 15000,              // Optional. Default: 30000ms.
  credentials: 'include',             // Optional. Default: 'include'.
  allowInsecure: false,               // Optional. Set true for http:// in development.
}
```

## Error Handling

The SDK throws `ApiError` for non-2xx responses and native `TypeError` for network failures.

```ts
import { isApiError } from 'modulacms-admin-sdk'

try {
  await client.users.get(userId)
} catch (err) {
  if (isApiError(err)) {
    console.error(`API ${err.status}: ${err.message}`, err.body)
  } else {
    console.error('Network error:', err)
  }
}
```

## Available Resources

### Standard CRUD

Each resource provides `list()`, `get(id)`, `create(params)`, `update(params)`, and `remove(id)`:

| Property | Entity | ID Type |
|---|---|---|
| `adminRoutes` | Admin routes | `Slug` / `AdminRouteID` |
| `adminContentData` | Admin content nodes | `AdminContentID` |
| `adminContentFields` | Admin content field values | `AdminContentFieldID` |
| `adminDatatypes` | Admin datatype definitions | `AdminDatatypeID` |
| `adminFields` | Admin field definitions | `AdminFieldID` |
| `contentData` | Public content nodes | `ContentID` |
| `contentFields` | Public content field values | `ContentFieldID` |
| `datatypes` | Datatype schemas | `DatatypeID` |
| `fields` | Field schemas | `FieldID` |
| `routes` | Public routes | `RouteID` |
| `media` | Media assets | `MediaID` |
| `mediaDimensions` | Media dimension presets | `string` |
| `users` | User accounts | `UserID` |
| `roles` | Permission roles | `RoleID` |
| `tokens` | API tokens | `string` |
| `usersOauth` | OAuth connections | `UserOauthID` |
| `tables` | Tables | `string` |

### Specialized Resources

| Property | Methods | Description |
|---|---|---|
| `auth` | `login`, `logout`, `me`, `register`, `reset` | Authentication |
| `adminTree` | `get(slug, format?)` | Content tree retrieval |
| `mediaUpload` | `upload(file)` | Multipart file upload |
| `sessions` | `update`, `remove` | Session management |
| `sshKeys` | `list`, `create`, `remove` | SSH key management |
| `import` | `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `bulk` | Bulk import |

## Development

```bash
bun install
bun run build       # Compile to dist/
bun run typecheck   # Type-check without emitting
bun test            # Run tests
```

## License

MIT
