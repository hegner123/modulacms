# Type Safety

Use branded ID types, entity types, and enums from `@modulacms/types` for compile-time safety across both SDKs.

## Branded ID Types

All entity IDs are branded string types that prevent you from mixing up IDs at the type level. You cannot pass a `UserID` where a `ContentID` belongs, even though both are strings at runtime.

```typescript
import type { ContentID, UserID, DatatypeID } from '@modulacms/types'

function getContent(id: ContentID): Promise<ContentData> { /* ... */ }

const userId = 'abc123' as UserID
const contentId = 'def456' as ContentID

getContent(contentId)  // OK
getContent(userId)     // Compile error: UserID is not assignable to ContentID
```

### How Branding Works

Branded types use a phantom property that exists only at the type level:

```typescript
type Brand<T, B extends string> = T & { readonly __brand: B }
type ContentID = Brand<string, 'ContentID'>
```

The `__brand` property never exists at runtime -- it serves purely as a compile-time discriminator. Create a branded value with a type assertion:

```typescript
const id = '01HXYZ...' as ContentID
```

### Available ID Types

| Type | Entity |
|------|--------|
| `UserID` | User accounts |
| `ContentID` | Public content data nodes |
| `ContentFieldID` | Public content field values |
| `ContentRelationID` | Public content relations |
| `ContentVersionID` | Public content version snapshots |
| `DatatypeID` | Datatype (schema) definitions |
| `FieldID` | Field (schema) definitions |
| `MediaID` | Media assets |
| `RoleID` | Permission roles |
| `PermissionID` | Access control permissions |
| `RolePermissionID` | Role-permission junction rows |
| `FieldTypeID` | Field type lookup entries |
| `RouteID` | Public routes |
| `SessionID` | Active sessions |
| `UserOauthID` | OAuth connections |
| `AdminContentID` | Admin content data nodes |
| `AdminContentFieldID` | Admin content field values |
| `AdminContentRelationID` | Admin content relations |
| `AdminContentVersionID` | Admin content version snapshots |
| `AdminDatatypeID` | Admin datatype definitions |
| `AdminFieldID` | Admin field definitions |
| `AdminRouteID` | Admin routes |
| `AdminFieldTypeID` | Admin field type lookup entries |
| `LocaleID` | Locales |
| `WebhookID` | Webhooks |
| `WebhookDeliveryID` | Webhook deliveries |

### Value Types

These branded types are not entity identifiers but carry semantic meaning:

| Type | Underlying | Purpose |
|------|-----------|---------|
| `Slug` | `string` | URL-safe route slug |
| `Email` | `string` | Email address |
| `URL` | `string` | URL string |

## Entity Types

Entity types represent the data structures the API returns. `@modulacms/types` defines them, and both SDKs re-export them.

### Core Entities

| Type | Package | Description |
|------|---------|-------------|
| `ContentData` | `@modulacms/types` | Public content tree node with sibling pointers |
| `ContentField` | `@modulacms/types` | Public content field value |
| `ContentRelation` | `@modulacms/types` | Relation between public content nodes |
| `ContentVersion` | `@modulacms/types` | Version snapshot of public content |
| `Datatype` | `@modulacms/types` | Content type schema definition |
| `Field` | `@modulacms/types` | Field schema definition |
| `FieldTypeInfo` | `@modulacms/types` | Field type lookup entry |
| `Route` | `@modulacms/types` | Public route |
| `Media` | `@modulacms/types` | Media asset |
| `MediaDimension` | `@modulacms/types` | Media dimension preset |
| `Locale` | `@modulacms/types` | Locale definition |
| `Webhook` | `@modulacms/types` | Webhook configuration |
| `WebhookDelivery` | `@modulacms/types` | Webhook delivery attempt |

### Content Tree Types

| Type | Package | Description |
|------|---------|-------------|
| `ContentTree` | `@modulacms/types` | Root of a content tree (`{ root: ContentNode }`) |
| `ContentNode` | `@modulacms/types` | Tree node with datatype, fields, and child nodes |
| `NodeDatatype` | `@modulacms/types` | Datatype definition paired with content instance |
| `NodeField` | `@modulacms/types` | Field definition paired with content field value |

### Query Types

| Type | Package | Description |
|------|---------|-------------|
| `QueryParams` | `@modulacms/types` | Query parameters (sort, limit, offset, locale, status, filters) |
| `QueryResult` | `@modulacms/types` | Paginated query response |
| `QueryItem` | `@modulacms/types` | Single content item in a query result |
| `QueryDatatype` | `@modulacms/types` | Datatype metadata in a query result |

### Admin-Only Types (from `@modulacms/admin-sdk`)

The admin SDK defines additional types for resources that only exist in the admin API:

| Type | Description |
|------|-------------|
| `AdminRoute` | Admin route definition |
| `AdminContentData` | Admin content tree node |
| `AdminContentField` | Admin content field value |
| `AdminDatatype` | Admin datatype definition |
| `AdminField` | Admin field definition |
| `User` | User account (includes `hash` -- server-side only) |
| `Role` | Permission role |
| `Permission` | Access control permission |
| `RolePermission` | Role-permission junction |
| `Token` | API token (SENSITIVE) |
| `UserOauth` | OAuth connection (SENSITIVE) |
| `Session` | Active session |
| `SshKey` | Full SSH key (includes public key material) |
| `SshKeyListItem` | SSH key summary (excludes public key) |
| `Table` | Named table |

### Create and Update Param Types

Each entity has corresponding `Create*Params` and `Update*Params` types. These are local to the admin SDK (not re-exported from `@modulacms/types`) because they are only used for write operations.

```typescript
import type { CreateDatatypeParams, UpdateDatatypeParams } from '@modulacms/admin-sdk'
```

Update types use full-replacement -- provide all fields. Some have special patterns:

- Route updates use a `slug_2` field to carry the current slug for the WHERE clause, allowing renames.
- User updates have an optional `password` field -- omit it to keep the existing password.

### View Types (Safe for Display)

The admin SDK provides "view" types that omit sensitive fields:

| View Type | Omits |
|-----------|-------|
| `UserWithRoleLabel` | `hash` |
| `UserFullView` | `hash`, uses safe sub-views for oauth/keys/sessions/tokens |
| `UserOauthView` | `access_token`, `refresh_token` |
| `UserSshKeyView` | `public_key` |
| `SessionView` | `session_data` |
| `TokenView` | `token` (bearer credential) |

## Enums

```typescript
import type { ContentStatus, FieldType, ContentFormat } from '@modulacms/types'
import { CONTENT_FORMATS } from '@modulacms/types'
```

### ContentStatus

```typescript
type ContentStatus = 'draft' | 'published'
```

### FieldType

```typescript
type FieldType =
  | 'text' | 'textarea' | 'number' | 'date' | 'datetime'
  | 'boolean' | 'select' | 'media' | '_id' | 'json'
  | 'richtext' | 'slug' | 'email' | 'url'
```

### ContentFormat

```typescript
type ContentFormat = 'contentful' | 'sanity' | 'strapi' | 'wordpress' | 'clean' | 'raw'
```

Use the `CONTENT_FORMATS` array for runtime validation:

```typescript
if (CONTENT_FORMATS.includes(userInput)) {
  // userInput is ContentFormat
}
```

## Pagination Types

```typescript
import type { PaginationParams, PaginatedResponse } from '@modulacms/types'

type PaginationParams = {
  limit: number
  offset: number
}

type PaginatedResponse<T> = {
  data: T[]
  total: number
  limit: number
  offset: number
}
```

## Re-Export Strategy

The admin SDK re-exports shared entity types from `@modulacms/types`, so you can import shared types from either package:

```typescript
// Both work
import type { ContentData, Datatype } from '@modulacms/types'
import type { ContentData, Datatype } from '@modulacms/admin-sdk'
```

Import Create and Update param types from `@modulacms/admin-sdk` -- they are not re-exported to the types package.

## Using Types Independently

You can import `@modulacms/types` directly without either SDK. This is useful for:

- Shared type definitions in a monorepo
- Backend code that processes CMS data without making HTTP calls
- Type-safe API response handling in custom HTTP clients

```typescript
import type {
  ContentID,
  DatatypeID,
  ContentData,
  Datatype,
  Field,
  ContentStatus,
  FieldType,
  ApiError,
  PaginatedResponse,
} from '@modulacms/types'
```
