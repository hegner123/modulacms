# @modulacms/types

Shared TypeScript type definitions for the ModulaCMS SDK ecosystem. This package provides branded ID types, entity interfaces, enums, pagination, and error types consumed by both [`@modulacms/sdk`](https://www.npmjs.com/package/@modulacms/sdk) (content delivery) and [`@modulacms/admin-sdk`](https://www.npmjs.com/package/@modulacms/admin-sdk) (admin CRUD).

## Install

```bash
npm install @modulacms/types
# or
pnpm add @modulacms/types
```

## Usage

```ts
import type {
  ContentID,
  UserID,
  ContentData,
  Field,
  PaginatedResponse,
  ApiError,
} from '@modulacms/types'

import { isApiError, CONTENT_FORMATS } from '@modulacms/types'
```

## What's included

### Branded ID types

All entity identifiers are nominal (branded) string types. Two values with the same underlying string but different brands are not assignable to each other, catching misuse at compile time.

```ts
import type { ContentID, UserID } from '@modulacms/types'

const contentId = '01HXK4N2F8QZJV3K7M1Y9ABCDE' as ContentID
const userId = '01HXK4N2F8QZJV3K7M1Y9ABCDE' as UserID

// contentId = userId  // compile error -- different brands
```

Available ID types:

| Type | Entity |
|------|--------|
| `UserID` | User account |
| `ContentID` | Public content node |
| `ContentFieldID` | Public content field value |
| `ContentRelationID` | Public content relation |
| `DatatypeID` | Datatype (schema) definition |
| `FieldID` | Field (schema) definition |
| `MediaID` | Media asset |
| `RoleID` | User role |
| `PermissionID` | Permission |
| `RolePermissionID` | Role-permission junction |
| `RouteID` | Public route |
| `SessionID` | Active session |
| `UserOauthID` | User OAuth connection |
| `AdminContentID` | Admin content node |
| `AdminContentFieldID` | Admin content field value |
| `AdminContentRelationID` | Admin content relation |
| `AdminDatatypeID` | Admin datatype definition |
| `AdminFieldID` | Admin field definition |
| `AdminRouteID` | Admin route |

Branded value types: `Slug`, `Email`, `URL`.

The `Brand<T, B>` utility type is also exported for creating custom branded types.

### Enums

```ts
import type { ContentStatus, FieldType, ContentFormat } from '@modulacms/types'
import { CONTENT_FORMATS } from '@modulacms/types'
```

**`ContentStatus`** -- `'draft' | 'published' | 'archived' | 'pending'`

**`FieldType`** -- `'text' | 'textarea' | 'number' | 'date' | 'datetime' | 'boolean' | 'select' | 'media' | 'relation' | 'json' | 'richtext' | 'slug' | 'email' | 'url'`

**`ContentFormat`** -- `'contentful' | 'sanity' | 'strapi' | 'wordpress' | 'clean' | 'raw'`

`CONTENT_FORMATS` is a runtime `as const` array for validating format strings.

### Entity types

**Content** -- `ContentData`, `ContentField`, `ContentRelation`

Content nodes form a linked-list tree using `parent_id`, `first_child_id`, `next_sibling_id`, and `prev_sibling_id` pointers for O(1) navigation and reordering.

**Schema** -- `Datatype`, `Field`, `DatatypeField`

Datatypes define content categories. Fields define individual data entries within a datatype. `DatatypeField` is the junction record linking them with sort order.

**Media** -- `Media`, `MediaDimension`

Media assets with support for responsive `srcset` and dimension presets.

**Routing** -- `Route`

Maps URL slugs to content trees.

**Tree** -- `NodeDatatype`, `NodeField`, `ContentNode`, `ContentTree`

The assembled hierarchical content tree returned by slug-based content delivery. `ContentTree` contains a `root` `ContentNode`, which recursively nests child nodes, each with their datatype definition, content instance data, and field values.

### Pagination

```ts
import type { PaginationParams, PaginatedResponse } from '@modulacms/types'

const params: PaginationParams = { limit: 20, offset: 0 }
// PaginatedResponse<T> = { data: T[], total: number, limit: number, offset: number }
```

### Error handling

```ts
import type { ApiError } from '@modulacms/types'
import { isApiError } from '@modulacms/types'

try {
  await client.content.get(id)
} catch (err) {
  if (isApiError(err)) {
    console.error(`API ${err.status}: ${err.message}`)
  }
}
```

`ApiError` has a `_tag: 'ApiError'` discriminant to distinguish API errors from network-level failures.

### Request options

```ts
import type { RequestOptions } from '@modulacms/types'

// Pass to any SDK method for per-request configuration
const options: RequestOptions = { signal: AbortSignal.timeout(5000) }
```

### Common primitives

`ULID` -- string alias for ULID identifiers.
`Timestamp` -- ISO 8601 UTC timestamp string.
`NullableString`, `NullableNumber` -- nullable variants for optional API fields.

## Build

Dual ESM + CJS output via [tsup](https://tsup.egoist.dev). TypeScript declarations are included.

```bash
pnpm build       # Build to dist/
pnpm typecheck   # Type-check without emitting
pnpm clean        # Remove build artifacts
```

## License

MIT
