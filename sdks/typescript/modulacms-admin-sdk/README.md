# @modulacms/admin-sdk

TypeScript SDK for the ModulaCMS admin API. Provides fully typed CRUD operations for all admin resources, authentication, media uploads, content tree retrieval, plugin management, server configuration, and bulk imports from external CMS platforms.

## Features

- Zero runtime dependencies -- built on native `fetch`
- Dual ESM + CJS builds via tsup
- Branded ID types prevent mixing incompatible identifiers at compile time
- HTTPS enforced by default; `allowInsecure` flag for local development
- 30s default timeout with per-request `AbortSignal` support
- Targets Node.js 18+ and modern browsers

## Install

```bash
npm install @modulacms/admin-sdk
# or
pnpm add @modulacms/admin-sdk
```

## Quick Start

```ts
import { createAdminClient, isApiError } from '@modulacms/admin-sdk'

const client = createAdminClient({
  baseUrl: 'https://cms.example.com',
  apiKey: process.env.CMS_API_KEY,
})

// Authentication
const me = await client.auth.me()

// CRUD operations
const users = await client.users.list()
const route = await client.adminRoutes.get(slug)

// Paginated listing
const page = await client.contentData.listPaginated({ limit: 20, offset: 0 })
console.log(page.data, page.total)

// Count
const total = await client.users.count()

// Media upload
const media = await client.mediaUpload.upload(file)

// Content tree
const tree = await client.adminTree.get(slug)

// Plugin management
const plugins = await client.plugins.list()

// Bulk import
const result = await client.import.contentful(exportData)
```

## Configuration

```ts
import type { ClientConfig } from '@modulacms/admin-sdk'

const config: ClientConfig = {
  baseUrl: 'https://cms.example.com', // Required. HTTPS enforced unless allowInsecure is set.
  apiKey: 'sk_live_abc123',           // Optional. Bearer token for server-side auth.
  defaultTimeout: 15000,              // Optional. Default: 30000ms.
  credentials: 'include',            // Optional. Default: 'include'. Controls fetch credentials mode.
  allowInsecure: false,               // Optional. Set true for http:// in development.
}
```

## Error Handling

The SDK throws `ApiError` for non-2xx responses and native `TypeError` for network failures.

```ts
import { isApiError } from '@modulacms/admin-sdk'

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

For media uploads, additional error guards are available:

```ts
import { isDuplicateMedia, isFileTooLarge, isInvalidMediaPath } from '@modulacms/admin-sdk'

try {
  await client.mediaUpload.upload(file, { path: 'products/shoes' })
} catch (err) {
  if (isDuplicateMedia(err)) { /* 409 conflict */ }
  if (isFileTooLarge(err)) { /* 400 size limit */ }
  if (isInvalidMediaPath(err)) { /* 400 bad path */ }
}
```

## Available Resources

### Standard CRUD Resources

Each resource provides `list()`, `get(id)`, `create(params)`, `update(params)`, `remove(id)`, `listPaginated({ limit, offset })`, and `count()`:

| Property | Entity | ID Type |
|---|---|---|
| `adminRoutes` | Admin routes | `Slug` (get/list) / `AdminRouteID` (remove) |
| `adminContentData` | Admin content nodes | `AdminContentID` |
| `adminContentFields` | Admin content field values | `AdminContentFieldID` |
| `adminDatatypes` | Admin datatype definitions | `AdminDatatypeID` |
| `adminFields` | Admin field definitions | `AdminFieldID` |
| `adminDatatypeFields` | Admin datatype-field mappings | `string` |
| `contentData` | Public content nodes | `ContentID` |
| `contentFields` | Public content field values | `ContentFieldID` |
| `datatypes` | Datatype schemas | `DatatypeID` |
| `fields` | Field schemas | `FieldID` |
| `datatypeFields` | Datatype-field mappings | `string` |
| `routes` | Public routes | `RouteID` |
| `media` | Media assets | `MediaID` |
| `mediaDimensions` | Media dimension presets | `string` |
| `users` | User accounts | `UserID` |
| `roles` | Permission roles | `RoleID` |
| `permissions` | Permissions | `PermissionID` |
| `rolePermissions` | Role-permission mappings | `RolePermissionID` |
| `tokens` | API tokens | `string` |
| `usersOauth` | OAuth connections | `UserOauthID` |
| `tables` | Custom tables | `string` |

`adminRoutes` also provides `listOrdered()` which returns routes sorted by position.

`rolePermissions` additionally provides `listByRole(roleId)` to get all permissions for a specific role.

### Authentication

```ts
// Login with email + password
const session = await client.auth.login({ email, password })

// Get current user
const me = await client.auth.me()

// Register new user
const user = await client.auth.register({ name, email, password })

// Reset password
await client.auth.reset({ user_id, password })

// Logout
await client.auth.logout()
```

### Content Tree

Retrieve the full content tree for a route, optionally in a specific CMS format:

```ts
// Raw format (default)
const tree = await client.adminTree.get('my-page')

// Formatted for a specific CMS structure
const formatted = await client.adminTree.get('my-page', 'contentful')
// Supported formats: 'raw' | 'contentful' | 'sanity' | 'strapi' | 'wordpress' | 'clean'
```

Returns `AdminTreeResponse` with `{ route, tree }` where `tree` is an array of `ContentTreeNode` with recursive `children` and resolved `fields`.

### Media Upload

```ts
const media = await client.mediaUpload.upload(file)

// With custom path prefix
const media = await client.mediaUpload.upload(file, { path: 'products/shoes' })
```

The returned `Media` object's `srcset` field may be `null` initially while async image processing completes.

### Sessions

```ts
await client.sessions.update({ session_id, ...params })
await client.sessions.remove(sessionId)
```

Sessions are created implicitly via login -- there is no `create` method.

### SSH Keys

```ts
// List keys (public_key omitted for security)
const keys = await client.sshKeys.list()

// Create a key (returns full key including public_key)
const key = await client.sshKeys.create({ name, public_key })

// Remove a key
await client.sshKeys.remove(keyId)
```

### Plugins

```ts
// List all plugins
const plugins = await client.plugins.list()

// Get detailed info (circuit breaker state, schema drift, etc.)
const info = await client.plugins.get('my-plugin')

// Lifecycle management
await client.plugins.enable('my-plugin')
await client.plugins.disable('my-plugin')
await client.plugins.reload('my-plugin')

// Orphaned table cleanup
const dryRun = await client.plugins.cleanupDryRun()
const result = await client.plugins.cleanupDrop({ tables: ['orphaned_table'] })
```

### Plugin Routes and Hooks

```ts
// List plugin-registered routes
const routes = await client.pluginRoutes.list()

// Approve or revoke route registrations
await client.pluginRoutes.approve([{ plugin: 'my-plugin', method: 'GET', path: '/custom' }])
await client.pluginRoutes.revoke([{ plugin: 'my-plugin', method: 'GET', path: '/custom' }])

// Same pattern for hooks
const hooks = await client.pluginHooks.list()
await client.pluginHooks.approve([{ plugin: 'my-plugin', event: 'content:create' }])
await client.pluginHooks.revoke([{ plugin: 'my-plugin', event: 'content:create' }])
```

### Server Configuration

```ts
// Get current config (sensitive values redacted)
const config = await client.config.get()

// Get a specific category
const dbConfig = await client.config.get('database')

// Update config
const result = await client.config.update({ port: 9090 })
if (result.restart_required) {
  console.log('Server restart needed')
}

// Get field metadata (labels, categories, sensitivity flags)
const meta = await client.config.meta()
```

### Bulk Import

Import content from external CMS platforms:

```ts
const result = await client.import.contentful(exportData)
const result = await client.import.sanity(exportData)
const result = await client.import.strapi(exportData)
const result = await client.import.wordpress(exportData)
const result = await client.import.clean(exportData)

// Generic bulk import with format parameter
const result = await client.import.bulk('contentful', exportData)
```

Returns `ImportResponse` with `{ success, datatypes_created, fields_created, content_created, message, errors }`.

## Request Options

All methods accept an optional `RequestOptions` parameter for per-request control:

```ts
const controller = new AbortController()

const users = await client.users.list({ signal: controller.signal })

// Cancel the request
controller.abort()
```

## Types

All types are exported from the package entry point:

```ts
import type {
  // Client
  ClientConfig,
  ModulaCMSAdminClient,
  CrudResource,

  // Branded IDs
  UserID, ContentID, MediaID, RoleID, DatatypeID, FieldID, RouteID,
  AdminContentID, AdminRouteID, AdminDatatypeID, AdminFieldID,
  PermissionID, RolePermissionID, SessionID, UserOauthID,
  Slug, Email, URL,

  // Enums
  ContentStatus, // 'draft' | 'published' | 'archived' | 'pending'
  FieldType,     // 'text' | 'textarea' | 'number' | 'date' | 'boolean' | 'media' | ...

  // Pagination
  PaginationParams,
  PaginatedResponse,

  // Errors
  ApiError,
  RequestOptions,
} from '@modulacms/admin-sdk'
```

## Development

This package is part of the `sdks/typescript/` pnpm workspace monorepo. From the repository root:

```bash
just sdk-install    # pnpm install (workspace root)
just sdk-build      # Build all packages (types first, then SDKs)
just sdk-test       # Run all SDK tests (Vitest)
just sdk-typecheck  # Typecheck all packages
```

Or within this package directly:

```bash
pnpm run build       # Compile to dist/
pnpm run typecheck   # Type-check without emitting
pnpm test            # Run tests
```

## License

MIT
