# modulacms-admin-sdk

TypeScript SDK for interacting with the ModulaCMS API. Provides typed client for content management, authentication, media uploads, and bulk imports.

## Overview

The modulacms-admin-sdk provides a production-grade TypeScript client for ModulaCMS servers. It handles authentication via API keys or session cookies, supports both secure HTTPS and local HTTP connections, and exposes a fully typed interface for all CMS operations including CRUD for content, schema management, user administration, and bulk content migration from external platforms. All API calls use the Fetch API with configurable timeouts and abort signal support.

## Installation

Install the package from npm and import the factory function and types.

```typescript
import { createAdminClient, isApiError } from 'modulacms-admin-sdk'
import type { ClientConfig, ModulaCMSAdminClient } from 'modulacms-admin-sdk'
```

## Exports

The package exports a factory function, type guard, and all public types.

### function createAdminClient

`function createAdminClient(config: ClientConfig): ModulaCMSAdminClient`

Creates a configured ModulaCMS client instance. Validates the base URL for HTTPS unless `allowInsecure` is set. Rejects empty API keys. Returns a client object with typed resource properties for all CMS endpoints.

Throws Error if baseUrl is not a valid URL.
Throws Error if baseUrl uses http without allowInsecure set to true.
Throws Error if apiKey is an empty string.

```typescript
const client = createAdminClient({
  baseUrl: 'https://cms.example.com',
  apiKey: process.env.CMS_API_KEY,
})
const users = await client.users.list()
```

### function isApiError

`function isApiError(err: unknown): err is ApiError`

Type guard that narrows an unknown caught value to ApiError. Returns false for network errors, timeouts, and non-SDK exceptions. Use this in catch blocks to differentiate API errors from network failures.

```typescript
try {
  await client.users.get(userId)
} catch (err) {
  if (isApiError(err)) {
    console.error(`API ${err.status}: ${err.message}`)
  }
}
```

## Types

### interface ClientConfig

Configuration options for createAdminClient.

```typescript
type ClientConfig = {
  baseUrl: string
  apiKey?: string
  defaultTimeout?: number
  credentials?: RequestCredentials
  allowInsecure?: boolean
}
```

baseUrl is the ModulaCMS server origin. Must be HTTPS unless allowInsecure is true.
apiKey is the Bearer token for server-side authentication. Omit for cookie-based auth.
defaultTimeout in milliseconds defaults to 30000.
credentials mode defaults to include for cookie handling.
allowInsecure permits HTTP connections for local development. Do not enable in production.

### interface ModulaCMSAdminClient

The main client object returned by createAdminClient. Each property is a resource with CRUD methods or specialized endpoints.

All CRUD resources expose list, get, create, update, remove, listPaginated, and count methods.

```typescript
const client = createAdminClient(config)
await client.users.list()
await client.auth.login({ email, password })
await client.mediaUpload.upload(file)
const tree = await client.adminTree.get(slug)
```

### interface CrudResource

Standard CRUD operations for an API resource. Type parameters are Entity, CreateParams, UpdateParams, and Id which defaults to string.

```typescript
type CrudResource<Entity, CreateParams, UpdateParams, Id = string> = {
  list: (opts?: RequestOptions) => Promise<Entity[]>
  get: (id: Id, opts?: RequestOptions) => Promise<Entity>
  create: (params: CreateParams, opts?: RequestOptions) => Promise<Entity>
  update: (params: UpdateParams, opts?: RequestOptions) => Promise<Entity>
  remove: (id: Id, opts?: RequestOptions) => Promise<void>
  listPaginated: (params: PaginationParams, opts?: RequestOptions) => Promise<PaginatedResponse<Entity>>
  count: (opts?: RequestOptions) => Promise<number>
}
```

listPaginated returns a paginated envelope when limit or offset are present. count returns the total entity count by issuing a zero-limit paginated request.

### interface RequestOptions

Per-request configuration passed to every SDK method. Contains an optional AbortSignal to cancel requests. The signal is merged with the default timeout signal.

```typescript
const controller = new AbortController()
await client.users.list({ signal: controller.signal })
controller.abort()
```

### interface ApiError

Structured error returned by the SDK when the API responds with a non-2xx status. Distinguished by the _tag discriminant field set to ApiError. Network-level failures throw native TypeError instead.

```typescript
type ApiError = {
  readonly _tag: 'ApiError'
  status: number
  message: string
  body?: unknown
}
```

status is the HTTP status code. message is the HTTP status text. body contains the parsed JSON response if the server returned application/json.

### type Brand

Nominal type utility that adds a compile-time tag to a base type. Prevents accidental assignment between structurally identical but semantically different types. The brand is erased at runtime.

```typescript
type Brand<T, B extends string> = T & { readonly __brand: B }
type OrderID = Brand<string, 'OrderID'>
const id = 'abc' as OrderID
```

### Branded ID Types

All entity identifiers are branded strings to prevent mixing incompatible IDs. The brands include UserID, AdminContentID, AdminContentFieldID, AdminContentRelationID, AdminDatatypeID, AdminFieldID, AdminRouteID, ContentID, ContentFieldID, ContentRelationID, DatatypeID, FieldID, MediaID, RoleID, RouteID, SessionID, and UserOauthID.

Slug, Email, and URL are also branded strings for type safety.

### type ContentStatus

Lifecycle status of a content item. Values are draft, published, archived, or pending.

draft means work in progress and not publicly visible.
published means live and publicly accessible.
archived means removed from public view but retained.
pending means awaiting review or approval.

### type FieldType

Supported field types for content schema definitions. Determines the editor widget and validation rules. Values are text, textarea, number, date, datetime, boolean, select, media, relation, json, richtext, slug, email, or url.

### interface PaginationParams

Parameters for paginated list requests.

```typescript
type PaginationParams = {
  limit: number
  offset: number
}
```

limit is the maximum number of items to return. offset is the number of items to skip before collecting the result set.

### interface PaginatedResponse

Envelope returned by the API when pagination query parameters are present. Contains the data page, total count, and the limit and offset that were applied.

```typescript
type PaginatedResponse<T> = {
  data: T[]
  total: number
  limit: number
  offset: number
}
```

## Authentication

The auth resource provides login, logout, session identity retrieval, user registration, and password reset endpoints.

### method auth.login

`auth.login(params: LoginRequest, opts?: RequestOptions): Promise<LoginResponse>`

Authenticate with email and password. Returns the authenticated user's identity. Throws ApiError with status 401 on invalid credentials.

```typescript
const response = await client.auth.login({
  email: 'user@example.com' as Email,
  password: 'secret',
})
console.log(response.user_id)
```

### method auth.logout

`auth.logout(opts?: RequestOptions): Promise<void>`

End the current authenticated session. Returns void on success.

### method auth.me

`auth.me(opts?: RequestOptions): Promise<MeResponse>`

Retrieve the currently authenticated user's identity and role. Throws ApiError with status 401 if not authenticated.

```typescript
const me = await client.auth.me()
console.log(me.username, me.role)
```

### method auth.register

`auth.register(params: CreateUserParams, opts?: RequestOptions): Promise<User>`

Register a new user account. Returns the created user entity.

### method auth.reset

`auth.reset(params: UpdateUserParams, opts?: RequestOptions): Promise<string>`

Reset a user's password or account details. Accepts updated user parameters including the new password hash. Returns a confirmation message string.

### interface LoginRequest

Credentials payload sent to login.

```typescript
type LoginRequest = {
  email: Email
  password: string
}
```

password is plaintext and transmitted over HTTPS only unless allowInsecure is set.

### interface LoginResponse

Successful login response containing the authenticated user's basic identity.

```typescript
type LoginResponse = {
  user_id: UserID
  email: Email
  username: string
  created_at: string
}
```

created_at is an ISO 8601 timestamp.

### interface MeResponse

Response from the me endpoint representing the currently authenticated user.

```typescript
type MeResponse = {
  user_id: UserID
  email: Email
  username: string
  name: string
  role: string
}
```

## Admin Routes

Admin routes serve as top-level containers for admin content trees. The adminRoutes resource uses Slug for get and list operations and AdminRouteID for remove operations.

### method adminRoutes.list

`adminRoutes.list(opts?: RequestOptions): Promise<AdminRoute[]>`

List all admin routes. Returns an array of AdminRoute entities.

### method adminRoutes.listOrdered

`adminRoutes.listOrdered(opts?: RequestOptions): Promise<AdminRoute[]>`

List all admin routes sorted by server-defined order. Returns an array of AdminRoute entities.

### method adminRoutes.get

`adminRoutes.get(slug: Slug, opts?: RequestOptions): Promise<AdminRoute>`

Retrieve a single admin route by its slug.

### method adminRoutes.create

`adminRoutes.create(params: CreateAdminRouteParams, opts?: RequestOptions): Promise<AdminRoute>`

Create a new admin route. Returns the created entity.

### method adminRoutes.update

`adminRoutes.update(params: UpdateAdminRouteParams, opts?: RequestOptions): Promise<AdminRoute>`

Update an admin route. The slug_2 field carries the current slug for the WHERE clause allowing slug to be changed in the same operation.

### method adminRoutes.remove

`adminRoutes.remove(id: AdminRouteID, opts?: RequestOptions): Promise<void>`

Remove an admin route by its unique ID.

### interface AdminRoute

An admin-side route entity.

```typescript
type AdminRoute = {
  admin_route_id: AdminRouteID
  slug: Slug
  title: string
  status: number
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

status is a numeric flag where 0 is inactive and 1 is active. date_created and date_modified are ISO 8601 timestamps.

## Admin Content Data

Admin content data nodes form the content tree using a linked-list structure with parent_id, first_child_id, next_sibling_id, and prev_sibling_id pointers.

### interface AdminContentData

An admin content data node in the content tree.

```typescript
type AdminContentData = {
  admin_content_data_id: AdminContentID
  parent_id: AdminContentID | null
  first_child_id: string | null
  next_sibling_id: string | null
  prev_sibling_id: string | null
  admin_route_id: string
  admin_datatype_id: AdminDatatypeID | null
  author_id: UserID | null
  status: ContentStatus
  date_created: string
  date_modified: string
}
```

parent_id is null for root nodes. admin_datatype_id defines the content schema or is null if untyped.

### interface CreateAdminContentDataParams

Parameters for creating a new admin content data node via POST to admincontentdatas.

```typescript
type CreateAdminContentDataParams = {
  parent_id: AdminContentID | null
  first_child_id: string | null
  next_sibling_id: string | null
  prev_sibling_id: string | null
  admin_route_id: string
  admin_datatype_id: AdminDatatypeID | null
  author_id: UserID | null
  status: ContentStatus
  date_created: string
  date_modified: string
}
```

### interface UpdateAdminContentDataParams

Parameters for updating an admin content data node via PUT to admincontentdatas.

```typescript
type UpdateAdminContentDataParams = {
  admin_content_data_id: AdminContentID
  parent_id: AdminContentID | null
  first_child_id: string | null
  next_sibling_id: string | null
  prev_sibling_id: string | null
  admin_route_id: string
  admin_datatype_id: AdminDatatypeID | null
  author_id: UserID | null
  status: ContentStatus
  date_created: string
  date_modified: string
}
```

admin_content_data_id identifies the node to update.

## Admin Content Fields

Admin content fields store field values for admin content data nodes. Each field links to a content node and a field definition.

### interface AdminContentField

A field value belonging to an admin content data node.

```typescript
type AdminContentField = {
  admin_content_field_id: AdminContentFieldID
  admin_route_id: string | null
  admin_content_data_id: string
  admin_field_id: AdminFieldID | null
  admin_field_value: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

admin_field_value is stored as a serialized string based on the field type.

### interface CreateAdminContentFieldParams

Parameters for creating a new admin content field value via POST to admincontentfields.

```typescript
type CreateAdminContentFieldParams = {
  admin_route_id: string | null
  admin_content_data_id: string
  admin_field_id: AdminFieldID | null
  admin_field_value: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

### interface UpdateAdminContentFieldParams

Parameters for updating an admin content field value via PUT to admincontentfields.

```typescript
type UpdateAdminContentFieldParams = {
  admin_content_field_id: AdminContentFieldID
  admin_route_id: string | null
  admin_content_data_id: string
  admin_field_id: AdminFieldID | null
  admin_field_value: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

admin_content_field_id identifies the field value to update.

## Admin Datatypes

Admin datatypes describe the schema of admin content types. Each datatype defines the structure and category of content nodes.

### interface AdminDatatype

An admin-side datatype definition.

```typescript
type AdminDatatype = {
  admin_datatype_id: AdminDatatypeID
  parent_id: AdminContentID | null
  label: string
  type: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

label is human-readable. type indicates the datatype category such as page or component.

### interface CreateAdminDatatypeParams

Parameters for creating a new admin datatype via POST to admindatatypes.

```typescript
type CreateAdminDatatypeParams = {
  parent_id: AdminContentID | null
  label: string
  type: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

### interface UpdateAdminDatatypeParams

Parameters for updating an admin datatype via PUT to admindatatypes.

```typescript
type UpdateAdminDatatypeParams = {
  admin_datatype_id: AdminDatatypeID
  parent_id: AdminContentID | null
  label: string
  type: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

admin_datatype_id identifies the datatype to update.

## Admin Fields

Admin fields define the name, type, validation, and UI configuration of content fields within datatypes.

### interface AdminField

An admin-side field definition.

```typescript
type AdminField = {
  admin_field_id: AdminFieldID
  parent_id: AdminDatatypeID | null
  label: string
  data: string
  validation: string
  ui_config: string
  type: FieldType
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

data, validation, and ui_config are JSON-encoded strings. type determines the field's data type and editor widget.

### interface CreateAdminFieldParams

Parameters for creating a new admin field via POST to adminfields.

```typescript
type CreateAdminFieldParams = {
  parent_id: AdminDatatypeID | null
  label: string
  data: string
  validation: string
  ui_config: string
  type: FieldType
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

### interface UpdateAdminFieldParams

Parameters for updating an admin field via PUT to adminfields.

```typescript
type UpdateAdminFieldParams = {
  admin_field_id: AdminFieldID
  parent_id: AdminDatatypeID | null
  label: string
  data: string
  validation: string
  ui_config: string
  type: FieldType
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

admin_field_id identifies the field to update.

## Admin Datatype Fields

AdminDatatypeField is a junction record linking an admin datatype to an admin field. No dedicated list or get endpoints exist.

### interface AdminDatatypeField

Junction record linking an admin datatype to an admin field.

```typescript
type AdminDatatypeField = {
  id: string
  admin_datatype_id: AdminDatatypeID
  admin_field_id: AdminFieldID
}
```

### interface CreateAdminDatatypeFieldParams

Parameters for creating a junction record via POST to admindatatypefields.

```typescript
type CreateAdminDatatypeFieldParams = {
  admin_datatype_id: AdminDatatypeID
  admin_field_id: AdminFieldID
}
```

### interface UpdateAdminDatatypeFieldParams

Parameters for updating a junction record via PUT to admindatatypefields.

```typescript
type UpdateAdminDatatypeFieldParams = {
  id: string
  admin_datatype_id: AdminDatatypeID
  admin_field_id: AdminFieldID
}
```

id identifies the junction record to update.

## Admin Content Relations

AdminContentRelation represents a directional relation between two admin content nodes via a specific field.

### interface AdminContentRelation

A directional relation between admin content nodes.

```typescript
type AdminContentRelation = {
  admin_content_relation_id: AdminContentRelationID
  source_content_id: AdminContentID
  target_content_id: AdminContentID
  admin_field_id: AdminFieldID
  sort_order: number
  date_created: string
}
```

source_content_id owns the relation. target_content_id is the referenced node. sort_order determines display ordering.

## Public Content Data

ContentData represents public content nodes in the published content tree. Uses the same linked-list structure as admin content data.

### interface ContentData

A public content data node.

```typescript
type ContentData = {
  content_data_id: ContentID
  parent_id: ContentID | null
  first_child_id: string | null
  next_sibling_id: string | null
  prev_sibling_id: string | null
  route_id: RouteID | null
  datatype_id: DatatypeID | null
  author_id: UserID | null
  status: ContentStatus
  date_created: string
  date_modified: string
}
```

route_id links to the public route this content belongs to or is null.

### interface CreateContentDataParams

Parameters for creating a public content data node via POST to contentdata.

```typescript
type CreateContentDataParams = {
  route_id: RouteID | null
  parent_id: ContentID | null
  first_child_id: string | null
  next_sibling_id: string | null
  prev_sibling_id: string | null
  datatype_id: DatatypeID | null
  author_id: UserID | null
  status: ContentStatus
  date_created: string
  date_modified: string
}
```

### interface UpdateContentDataParams

Parameters for updating a public content data node via PUT to contentdata.

```typescript
type UpdateContentDataParams = {
  content_data_id: ContentID
  parent_id: ContentID | null
  first_child_id: string | null
  next_sibling_id: string | null
  prev_sibling_id: string | null
  route_id: RouteID | null
  datatype_id: DatatypeID | null
  author_id: UserID | null
  status: ContentStatus
  date_created: string
  date_modified: string
}
```

content_data_id identifies the node to update.

## Public Content Fields

ContentField stores field values for public content data nodes.

### interface ContentField

A field value belonging to a public content data node.

```typescript
type ContentField = {
  content_field_id: ContentFieldID
  route_id: RouteID | null
  content_data_id: ContentID | null
  field_id: FieldID | null
  field_value: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

field_value is stored as a serialized string.

### interface CreateContentFieldParams

Parameters for creating a public content field value via POST to contentfields.

```typescript
type CreateContentFieldParams = {
  route_id: RouteID | null
  content_data_id: ContentID | null
  field_id: FieldID | null
  field_value: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

### interface UpdateContentFieldParams

Parameters for updating a public content field value via PUT to contentfields.

```typescript
type UpdateContentFieldParams = {
  content_field_id: ContentFieldID
  route_id: RouteID | null
  content_data_id: ContentID | null
  field_id: FieldID | null
  field_value: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

content_field_id identifies the field value to update.

## Public Content Relations

ContentRelation represents a directional relation between two public content nodes.

### interface ContentRelation

A directional relation between public content nodes.

```typescript
type ContentRelation = {
  content_relation_id: ContentRelationID
  source_content_id: ContentID
  target_content_id: ContentID
  field_id: FieldID
  sort_order: number
  date_created: string
}
```

source_content_id owns the relation. target_content_id is the referenced node.

## Datatypes

Datatype entities describe the schema of public content types.

### interface Datatype

A datatype schema definition.

```typescript
type Datatype = {
  datatype_id: DatatypeID
  parent_id: ContentID | null
  label: string
  type: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

label is human-readable. type indicates the content category.

### interface CreateDatatypeParams

Parameters for creating a datatype via POST to datatype.

```typescript
type CreateDatatypeParams = {
  datatype_id: DatatypeID
  parent_id: ContentID | null
  label: string
  type: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

datatype_id is client-generated.

### interface UpdateDatatypeParams

Parameters for updating a datatype via PUT to datatype.

```typescript
type UpdateDatatypeParams = {
  datatype_id: DatatypeID
  parent_id: ContentID | null
  label: string
  type: string
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

## Fields

Field entities define schema-level field definitions for datatypes.

### interface Field

A field schema definition.

```typescript
type Field = {
  field_id: FieldID
  parent_id: DatatypeID | null
  label: string
  data: string
  validation: string
  ui_config: string
  type: FieldType
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

data, validation, and ui_config are JSON-encoded.

### interface CreateFieldParams

Parameters for creating a field via POST to fields.

```typescript
type CreateFieldParams = {
  field_id: FieldID
  parent_id: DatatypeID | null
  label: string
  data: string
  validation: string
  ui_config: string
  type: FieldType
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

field_id is client-generated.

### interface UpdateFieldParams

Parameters for updating a field via PUT to fields.

```typescript
type UpdateFieldParams = {
  field_id: FieldID
  parent_id: DatatypeID | null
  label: string
  data: string
  validation: string
  ui_config: string
  type: FieldType
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

## Datatype Fields

DatatypeField is a junction record linking a datatype to a field with sort ordering.

### interface DatatypeField

Junction record linking a datatype to a field.

```typescript
type DatatypeField = {
  id: string
  datatype_id: DatatypeID
  field_id: FieldID
  sort_order: number
}
```

sort_order determines field display order within the datatype.

### interface CreateDatatypeFieldParams

Parameters for creating a junction record via POST to datatypefields.

```typescript
type CreateDatatypeFieldParams = {
  datatype_id: DatatypeID
  field_id: FieldID
  sort_order: number
}
```

### interface UpdateDatatypeFieldParams

Parameters for updating a junction record via PUT to datatypefields.

```typescript
type UpdateDatatypeFieldParams = {
  id: string
  datatype_id: DatatypeID
  field_id: FieldID
  sort_order: number
}
```

## Public Routes

Route entities map URL slugs to public content trees.

### interface Route

A public-facing route.

```typescript
type Route = {
  route_id: RouteID
  slug: Slug
  title: string
  status: number
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

status is 0 for inactive and 1 for active.

### interface CreateRouteParams

Parameters for creating a public route via POST to routes.

```typescript
type CreateRouteParams = {
  slug: Slug
  title: string
  status: number
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

route_id is server-generated and not included in create params.

### interface UpdateRouteParams

Parameters for updating a public route via PUT to routes.

```typescript
type UpdateRouteParams = {
  slug: Slug
  title: string
  status: number
  author_id: UserID | null
  date_created: string
  date_modified: string
  slug_2: Slug
}
```

slug_2 carries the current slug for the WHERE clause allowing slug to be changed.

## Media

Media entities represent uploaded assets such as images, videos, and documents.

### interface Media

A media asset stored in the CMS.

```typescript
type Media = {
  media_id: MediaID
  name: string | null
  display_name: string | null
  alt: string | null
  caption: string | null
  description: string | null
  class: string | null
  mimetype: string | null
  dimensions: string | null
  url: URL
  srcset: string | null
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

url is the primary asset URL. srcset is the responsive srcset attribute value or null. srcset may be null initially after upload while async processing completes.

### interface CreateMediaParams

Parameters for creating a media record via POST to media.

```typescript
type CreateMediaParams = {
  name: string | null
  display_name: string | null
  alt: string | null
  caption: string | null
  description: string | null
  class: string | null
  mimetype: string | null
  dimensions: string | null
  url: URL
  srcset: string | null
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

### interface UpdateMediaParams

Parameters for updating a media record via PUT to media.

```typescript
type UpdateMediaParams = {
  media_id: MediaID
  name: string | null
  display_name: string | null
  alt: string | null
  caption: string | null
  description: string | null
  class: string | null
  mimetype: string | null
  dimensions: string | null
  url: URL
  srcset: string | null
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

media_id identifies the asset to update.

## Media Dimensions

MediaDimension entities define named dimension presets for responsive media rendering.

### interface MediaDimension

A named dimension preset.

```typescript
type MediaDimension = {
  md_id: string
  label: string | null
  width: number | null
  height: number | null
  aspect_ratio: string | null
}
```

label is human-readable such as thumbnail or hero. width and height are in pixels or null for unconstrained.

### interface CreateMediaDimensionParams

Parameters for creating a dimension preset via POST to mediadimensions.

```typescript
type CreateMediaDimensionParams = {
  label: string | null
  width: number | null
  height: number | null
  aspect_ratio: string | null
}
```

### interface UpdateMediaDimensionParams

Parameters for updating a dimension preset via PUT to mediadimensions.

```typescript
type UpdateMediaDimensionParams = {
  md_id: string
  label: string | null
  width: number | null
  height: number | null
  aspect_ratio: string | null
}
```

md_id identifies the preset to update.

## Media Upload

The mediaUpload resource provides file upload via multipart form-data.

### method mediaUpload.upload

`mediaUpload.upload(file: File | Blob, opts?: RequestOptions): Promise<Media>`

Upload a file to the media library. Sends a multipart form-data POST request. Returns the created media entity. The srcset field may be null initially because responsive image variants are generated asynchronously. Poll client.media.get after a delay to retrieve the processed srcset.

Throws ApiError on non-2xx responses.
Throws TypeError on network failure.

```typescript
const file = new File(['content'], 'image.png', { type: 'image/png' })
const media = await client.mediaUpload.upload(file)
console.log(media.media_id, media.url)
```

## Users

User entities represent registered accounts with authentication credentials and role assignments.

### interface User

A registered user account.

```typescript
type User = {
  user_id: UserID
  username: string
  name: string
  email: Email
  hash: string
  role: string
  date_created: string
  date_modified: string
}
```

hash is the password hash. role is the assigned role label.

### interface CreateUserParams

Parameters for registering a new user via POST to auth register.

```typescript
type CreateUserParams = {
  username: string
  name: string
  email: Email
  hash: string
  role: string
  date_created: string
  date_modified: string
}
```

### interface UpdateUserParams

Parameters for updating a user via PUT to users.

```typescript
type UpdateUserParams = {
  user_id: UserID
  username: string
  name: string
  email: Email
  hash: string
  role: string
  date_created: string
  date_modified: string
}
```

user_id identifies the user to update.

## Roles

Role entities define permission sets that can be assigned to users.

### interface Role

A permission role.

```typescript
type Role = {
  role_id: RoleID
  label: string
  permissions: string
}
```

label is human-readable. permissions is a JSON-encoded permissions map.

### interface CreateRoleParams

Parameters for creating a role via POST to roles.

```typescript
type CreateRoleParams = {
  label: string
  permissions: string
}
```

### interface UpdateRoleParams

Parameters for updating a role via PUT to roles.

```typescript
type UpdateRoleParams = {
  role_id: RoleID
  label: string
  permissions: string
}
```

role_id identifies the role to update.

## Tokens

Token entities represent API tokens or refresh tokens issued to users. The token field contains a bearer credential and must be treated as a secret.

### interface Token

An API token or refresh token. SENSITIVE.

```typescript
type Token = {
  id: string
  user_id: UserID | null
  token_type: string
  token: string
  issued_at: string
  expires_at: string
  revoked: boolean
}
```

token is SENSITIVE and must never be logged or exposed. token_type indicates access or refresh. issued_at and expires_at are ISO 8601 timestamps.

### interface CreateTokenParams

Parameters for creating a token via POST to tokens. SENSITIVE.

```typescript
type CreateTokenParams = {
  user_id: UserID | null
  token_type: string
  token: string
  issued_at: string
  expires_at: string
  revoked: boolean
}
```

### interface UpdateTokenParams

Parameters for updating a token via PUT to tokens. SENSITIVE.

```typescript
type UpdateTokenParams = {
  token: string
  issued_at: string
  expires_at: string
  revoked: boolean
  id: string
}
```

id identifies the token to update.

## User OAuth

UserOauth entities link users to external OAuth providers. Contains access_token and refresh_token which are SENSITIVE.

### interface UserOauth

An OAuth connection. SENSITIVE.

```typescript
type UserOauth = {
  user_oauth_id: UserOauthID
  user_id: UserID | null
  oauth_provider: string
  oauth_provider_user_id: string
  access_token: string
  refresh_token: string
  token_expires_at: string
  date_created: string
}
```

oauth_provider is the provider name such as google or github. access_token and refresh_token are SENSITIVE and must never be logged or exposed.

### interface CreateUserOauthParams

Parameters for creating an OAuth connection via POST to usersoauth. SENSITIVE.

```typescript
type CreateUserOauthParams = {
  user_id: UserID | null
  oauth_provider: string
  oauth_provider_user_id: string
  access_token: string
  refresh_token: string
  token_expires_at: string
  date_created: string
}
```

### interface UpdateUserOauthParams

Parameters for updating an OAuth connection via PUT to usersoauth. SENSITIVE.

```typescript
type UpdateUserOauthParams = {
  access_token: string
  refresh_token: string
  token_expires_at: string
  user_oauth_id: UserOauthID
}
```

user_oauth_id identifies the OAuth connection to update.

## Sessions

Session entities track active user sessions. Sessions are created implicitly via login. No list get or create endpoints exist.

### interface Session

An active user session.

```typescript
type Session = {
  session_id: SessionID
  user_id: UserID | null
  created_at: string
  expires_at: string
  last_access: string | null
  ip_address: string | null
  user_agent: string | null
  session_data: string | null
}
```

session_data is JSON-encoded session payload or null.

### method sessions.update

`sessions.update(params: UpdateSessionParams, opts?: RequestOptions): Promise<Session>`

Update session metadata. Returns the updated session entity.

### method sessions.remove

`sessions.remove(id: SessionID, opts?: RequestOptions): Promise<void>`

Invalidate a session by its ID.

### interface UpdateSessionParams

Parameters for updating a session via PUT to sessions.

```typescript
type UpdateSessionParams = {
  session_id: SessionID
  user_id: UserID | null
  created_at: string
  expires_at: string
  last_access: string | null
  ip_address: string | null
  user_agent: string | null
  session_data: string | null
}
```

session_id identifies the session to update.

## SSH Keys

SSH key management provides list create and remove operations. List returns summaries without public key material. Create returns the full key including public_key.

### interface SshKey

A full SSH key record including public key material.

```typescript
type SshKey = {
  ssh_key_id: string
  user_id: string | null
  public_key: string
  key_type: string
  fingerprint: string
  label: string
  date_created: string
  last_used: string
}
```

public_key is the public key material such as ssh-ed25519 AAAA. key_type is the algorithm such as ssh-ed25519 or ssh-rsa. fingerprint is the key fingerprint such as SHA256.

### interface SshKeyListItem

Summary SSH key record returned by list. Omits public_key for security.

```typescript
type SshKeyListItem = {
  ssh_key_id: string
  key_type: string
  fingerprint: string
  label: string
  date_created: string
  last_used: string
}
```

### method sshKeys.list

`sshKeys.list(opts?: RequestOptions): Promise<SshKeyListItem[]>`

List all SSH keys for the authenticated user. Returns summary items without public key material.

### method sshKeys.create

`sshKeys.create(params: CreateSshKeyRequest, opts?: RequestOptions): Promise<SshKey>`

Register a new SSH public key. Returns the full SSH key record including public_key.

### method sshKeys.remove

`sshKeys.remove(id: string, opts?: RequestOptions): Promise<void>`

Remove an SSH key by its ID. Uses path-parameter deletion with DELETE to ssh-keys slash id.

### interface CreateSshKeyRequest

Parameters for registering a new SSH key via POST to ssh-keys.

```typescript
type CreateSshKeyRequest = {
  public_key: string
  label: string
}
```

public_key is the public key material to register. label is human-readable.

## Tables

Table entities represent named tables in the CMS.

### interface Table

A named table entity.

```typescript
type Table = {
  id: string
  label: string
  author_id: UserID | null
}
```

label is human-readable.

### interface CreateTableParams

Parameters for creating a table via POST to tables.

```typescript
type CreateTableParams = {
  label: string
}
```

### interface UpdateTableParams

Parameters for updating a table via PUT to tables.

```typescript
type UpdateTableParams = {
  label: string
  id: string
}
```

id identifies the table to update.

## Admin Tree

The adminTree resource provides fully resolved content hierarchies for admin routes.

### method adminTree.get

`adminTree.get(slug: Slug, format?: TreeFormat, opts?: RequestOptions): Promise<AdminTreeResponse | Record<string, unknown>>`

Retrieve the full content tree for an admin route by slug. When format is raw or omitted returns an AdminTreeResponse. For other formats such as contentful or sanity returns a platform-specific JSON structure as Record string unknown.

```typescript
const tree = await client.adminTree.get('home' as Slug)
console.log(tree.route.title, tree.tree.length)

const contentful = await client.adminTree.get('home' as Slug, 'contentful')
```

### interface ContentTreeField

A resolved field within a content tree node.

```typescript
type ContentTreeField = {
  field_label: string
  field_type: FieldType
  field_value: string | null
}
```

field_label is human-readable. field_type determines the data type. field_value is the current value or null if unset.

### interface ContentTreeNode

A node in the admin content tree with resolved fields and children.

```typescript
type ContentTreeNode = {
  content_data_id: AdminContentID
  parent_id: AdminContentID | null
  first_child_id: string | null
  next_sibling_id: string | null
  prev_sibling_id: string | null
  datatype_label: string
  datatype_type: string
  fields: ContentTreeField[]
  children: ContentTreeNode[]
}
```

datatype_label and datatype_type describe the node's schema. fields contains resolved field values. children contains recursively resolved child nodes.

### interface AdminTreeResponse

Complete admin tree response containing route metadata and the content tree.

```typescript
type AdminTreeResponse = {
  route: AdminRoute
  tree: ContentTreeNode[]
}
```

route is the admin route this tree belongs to. tree contains root-level content nodes with recursively resolved children.

### type TreeFormat

Output format for the admin tree endpoint. Values are raw, contentful, sanity, strapi, wordpress, or clean.

raw returns native ModulaCMS tree structure. Other formats return platform-specific JSON.

## Import

The import resource provides bulk content migration from external CMS platforms.

### method import.contentful

`import.contentful(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse>`

Import content from Contentful export format. Returns import result with creation counts and errors.

### method import.sanity

`import.sanity(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse>`

Import content from Sanity.io export format.

### method import.strapi

`import.strapi(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse>`

Import content from Strapi export format.

### method import.wordpress

`import.wordpress(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse>`

Import content from WordPress export format.

### method import.clean

`import.clean(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse>`

Import content from ModulaCMS native clean format.

### method import.bulk

`import.bulk(format: ImportFormat, data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse>`

Import content using a dynamic format specifier. Routes to POST import with query parameter format.

```typescript
const result = await client.import.bulk('contentful', exportData)
console.log(result.content_created, result.errors)
```

### type ImportFormat

Supported CMS platform formats for content import. Values are contentful, sanity, strapi, wordpress, or clean.

### interface ImportResponse

Server response from a bulk import operation.

```typescript
type ImportResponse = {
  success: boolean
  datatypes_created: number
  fields_created: number
  content_created: number
  message: string
  errors: string[]
}
```

success indicates whether the import completed without fatal errors. datatypes_created, fields_created, and content_created are creation counts. message is a human-readable summary. errors contains non-fatal error messages encountered during import.

## Error Handling

All SDK methods throw ApiError on non-2xx responses and TypeError on network failures. Use the isApiError type guard to differentiate.

```typescript
try {
  await client.users.get(userId)
} catch (err) {
  if (isApiError(err)) {
    if (err.status === 404) {
      console.error('User not found')
    } else if (err.status === 401) {
      console.error('Not authenticated')
    } else {
      console.error(`API error ${err.status}: ${err.message}`)
    }
  } else {
    console.error('Network or timeout error', err)
  }
}
```

ApiError includes the HTTP status code, status text, and optional parsed JSON body. Network errors timeout errors and abort errors throw native TypeError or DOMException.

## Request Cancellation

Every SDK method accepts an optional RequestOptions parameter with an AbortSignal. The signal is merged with the default timeout signal so either signal aborting will cancel the request.

```typescript
const controller = new AbortController()
const promise = client.users.list({ signal: controller.signal })
controller.abort()
try {
  await promise
} catch (err) {
  console.log('Request aborted')
}
```

The default timeout is 30000 milliseconds unless overridden in ClientConfig.

## Pagination

All CRUD resources expose listPaginated and count methods. When limit or offset parameters are present the server returns a paginated envelope instead of a plain array.

```typescript
const page = await client.users.listPaginated({ limit: 10, offset: 0 })
console.log(page.data, page.total, page.limit, page.offset)

const total = await client.users.count()
console.log(`Total users: ${total}`)
```

listPaginated returns PaginatedResponse containing data array, total count, and the limit and offset that were applied. count is implemented as a zero-limit paginated request and returns the total entity count.
