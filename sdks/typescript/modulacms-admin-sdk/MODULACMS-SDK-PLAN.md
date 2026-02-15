# ModulaCMS TypeScript SDK - Implementation Plan

> **Canonical document.** This is the authoritative implementation plan. Files in `ai/docs/` (`PLAN.md`, `PLAN-RC.md`) are earlier drafts and must not override decisions made here.

## Context

Build a TypeScript SDK for the ModulaCMS admin API from scratch. The project has `ai/docs/API_CONTRACT.md` (endpoint spec) and `ai/docs/STRUCT_REFERENCE.md` (Go struct definitions). The SDK provides typed functions for all admin API endpoints, targeting both browser and Node.js via native `fetch`.

**User choices:** Admin scope only, native fetch, Bun tooling, functional style (no classes), ESM-only (no CJS).

**Initial version:** `0.1.0` (pre-stability, minor = breaking under `0.x` semver).

---

## File Structure

```
modulacms-sdk/
├── src/
│   ├── index.ts              # createClient factory + re-exports
│   ├── http.ts               # Fetch wrapper: auth, errors, base URL
│   ├── resource.ts           # Generic CRUD resource factory
│   ├── types/
│   │   ├── index.ts          # Re-exports all types (using `export type` for type-only exports)
│   │   ├── common.ts         # Brand utility, branded ID types, Timestamp, RequestOptions, error shapes + isApiError guard
│   │   ├── auth.ts           # Login/me request & response types
│   │   ├── admin.ts          # AdminRoute, AdminContentData, AdminContentField, AdminDatatype, AdminField, AdminDatatypeField, AdminContentRelation + create/update params
│   │   ├── content.ts        # ContentData, ContentField, ContentRelation + create/update params
│   │   ├── schema.ts         # Datatype, Field, DatatypeField + create/update params (DatatypeField used in tree responses, no standalone endpoint)
│   │   ├── media.ts          # Media, MediaDimension types
│   │   ├── users.ts          # User, Role, Token (SENSITIVE), UserOauth (SENSITIVE), Session, SshKey types + create/update params
│   │   ├── routing.ts        # Route + create/update params
│   │   ├── tables.ts         # Table + create/update params
│   │   ├── import.ts         # Import request/response types
│   │   └── tree.ts           # Admin tree response types (ContentTreeNode, ContentTreeField)
│   ├── resources/
│   │   ├── auth.ts           # Auth endpoints (login, logout, me, register, reset)
│   │   ├── admin-tree.ts     # GET /admin/tree/{slug} with format param
│   │   ├── media-upload.ts   # POST multipart/form-data upload
│   │   ├── sessions.ts       # PUT/DELETE only (no list/get/create)
│   │   ├── ssh-keys.ts       # Different URL pattern (path param delete, no get-single/update)
│   │   └── import.ts         # POST-only import endpoints
├── package.json
├── tsconfig.json
├── ai/docs/API_CONTRACT.md   # (exists)
├── ai/docs/STRUCT_REFERENCE.md # (exists)
└── CLAUDE.md                 # (exists)
```

---

## Architecture

### HTTP Layer (`src/http.ts`)

A `createHttpClient(config)` factory that returns a configured fetch wrapper. **`HttpClient` is internal -- not exported in the public API.** Consumers interact only through the typed resource methods on `ModulaCMSClient`.

```typescript
type RequestOptions = {
  signal?: AbortSignal
}

// Internal type -- not exported
type HttpClient = {
  get: <T>(path: string, params?: Record<string, string>, opts?: RequestOptions) => Promise<T>
  post: <T>(path: string, body?: Record<string, unknown>, opts?: RequestOptions) => Promise<T>
  put: <T>(path: string, body?: Record<string, unknown>, opts?: RequestOptions) => Promise<T>
  del: (path: string, params?: Record<string, string>, opts?: RequestOptions) => Promise<void>
  raw: (path: string, init: RequestInit) => Promise<Response>  // for media upload
}
```

- Prepends `baseUrl + /api/v1` to all paths
- Sets `Content-Type: application/json` for JSON requests
- Attaches `Authorization: Bearer <apiKey>` header when apiKey is provided
- Sets `credentials` to the value from `ClientConfig` (default `'include'` for cookie auth)
- Applies `AbortSignal.timeout(defaultTimeout)` to every request (default 30s), merged with any per-request `signal` via `AbortSignal.any([timeoutSignal, opts.signal])`. This preserves the default timeout even when consumers provide their own signal for manual cancellation. If no per-request signal is provided, only the timeout signal is used.
- Validates `Content-Type: application/json` on response headers before calling `response.json()`. If the response is not JSON (e.g., HTML error page from a proxy), sets `ApiError.body` to `undefined` and `message` to the HTTP status text. This produces clearer errors than letting `.json()` fail with "Unexpected token <".
- Throws `ApiError` on non-2xx responses (callers use `isApiError()` type guard in catch blocks)
- Network errors (connection reset, DNS failure) propagate as native `TypeError` -- callers distinguish network failures (not `isApiError`) from API failures (`isApiError`) in catch blocks
- Passes the merged signal through to the underlying `fetch` call

**Design notes:**
- **Trust model:** The generic `<T>` return on get/post/put is an intentional tradeoff. This SDK is designed for use against a trusted, first-party ModulaCMS server where the API contract is known and version-matched. Server responses are not validated at runtime -- `response.json()` results are returned as-is with the declared type. If the SDK is ever used against an untrusted server or a server running a different API version, response data may not match the declared types. Consumers requiring runtime validation should use the `raw` method and validate themselves.
- Since `HttpClient` is internal, the generic `<T>` never leaks to consumers -- they only see concrete return types like `Promise<User>` from the resource methods.
- Request bodies use `Record<string, unknown>` because it's the simplest type that covers all request payloads including fields with `| null` values and branded ID types (whose `__brand` property exists only in the type system and is erased at runtime).

### CRUD Factory (`src/resource.ts`)

A single `createResource<Entity, CreateParams, UpdateParams, Id = string>()` that covers the standard CRUD resources. The `Id` type parameter defaults to `string` but can be overridden for resources with branded ID types:

```typescript
type CrudResource<Entity, CreateParams, UpdateParams, Id = string> = {
  list: (opts?: RequestOptions) => Promise<Entity[]>
  get: (id: Id, opts?: RequestOptions) => Promise<Entity>
  create: (params: CreateParams, opts?: RequestOptions) => Promise<Entity>
  update: (params: UpdateParams, opts?: RequestOptions) => Promise<Entity>
  remove: (id: Id, opts?: RequestOptions) => Promise<void>
}
```

The factory takes `(http: HttpClient, path: string)` and builds all 5 methods from the path.

**URL construction:** The API uses a consistent pattern for item operations. The factory builds URLs as follows:
- `list()` -> `GET /{path}` (no trailing slash)
- `get(id)` -> `GET /{path}/?q={id}` (trailing slash + query param)
- `create(params)` -> `POST /{path}` (no trailing slash, params as JSON body)
- `update(params)` -> `PUT /{path}/` (trailing slash, params as JSON body)
- `remove(id)` -> `DELETE /{path}/?q={id}` (trailing slash + query param)

**Update method ID handling:** The `UpdateParams` type for each resource includes a unique identifier field. For most resources this is the entity's primary key ID (matching the Go backend where update params embed the ID). The factory serializes the full `UpdateParams` object as the PUT request body. The ID is part of the params, not a separate argument.

**Exceptions:** `UpdateAdminRouteParams` and `UpdateRouteParams` do NOT include the entity's primary key. They use a `slug_2: Slug` field as the WHERE clause identifier (the old/previous slug, allowing the slug itself to be changed while locating the record by its previous value). See `STRUCT_REFERENCE.md` lines 762-763 and 818-819.

All methods throw `ApiError` on non-2xx responses (including 404 on `get`). This is the SDK convention -- callers handle errors via try/catch with the `isApiError()` type guard.

All methods accept an optional `RequestOptions` parameter (last argument) for `AbortSignal` support.

**Pagination:** The current API contract does not document `limit`/`offset` support on standard list endpoints. The `list` method initially takes no filter params. If the backend adds pagination support, a `ListParams` type will be added in a minor version bump.

Resources using this factory:

| API Path | Client Property | Id Type |
|----------|----------------|---------|
| `adminroutes` | `client.adminRoutes` | `Slug` for get, `AdminRouteID` for remove (split -- see below) |
| `admincontentdatas` | `client.adminContentData` | `AdminContentID` |
| `admincontentfields` | `client.adminContentFields` | `AdminContentFieldID` |
| `admindatatypes` | `client.adminDatatypes` | `AdminDatatypeID` |
| `adminfields` | `client.adminFields` | `AdminFieldID` |
| `contentdata` | `client.contentData` | `ContentID` |
| `contentfields` | `client.contentFields` | `ContentFieldID` |
| `datatype` | `client.datatypes` | `DatatypeID` |
| `fields` | `client.fields` | `FieldID` |
| `routes` | `client.routes` | `RouteID` |
| `media` | `client.media` | `MediaID` |
| `mediadimensions` | `client.mediaDimensions` | `string` (MdID is unbranded) |
| `users` | `client.users` | `UserID` |
| `roles` | `client.roles` | `RoleID` |
| `tokens` | `client.tokens` | `string` (Token ID is unbranded) |
| `usersoauth` | `client.usersOauth` | `UserOauthID` |
| `tables` | `client.tables` | `string` (Table ID is unbranded) |

**Admin routes** gets an extended version with `listOrdered()` and `getBySlug(slug: Slug)`. Its base CRUD uses:
- `get` accepts `Slug` (the API contract uses `GET /adminroutes/?q={slug}`)
- `remove` accepts `AdminRouteID` (the API contract uses `DELETE /adminroutes/?q={ulid}`)

This means admin routes has a split identifier: `CrudResource<AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams, Slug>` for most operations, but `remove` is overridden to accept `AdminRouteID` instead.

### Special Resources

Resources that do not follow the standard 5-method CRUD pattern:

**`src/resources/auth.ts`** -- Auth endpoints:

| Method | Signature | HTTP | Returns |
|--------|-----------|------|---------|
| `login` | `(params: LoginRequest, opts?) => Promise<LoginResponse>` | `POST /auth/login` | User summary; also sets session cookie |
| `logout` | `(opts?) => Promise<void>` | `POST /auth/logout` | Server returns `{message}` but SDK discards it |
| `me` | `(opts?) => Promise<MeResponse>` | `GET /auth/me` | Current user info |
| `register` | `(params: CreateUserParams, opts?) => Promise<User>` | `POST /auth/register` | Full User entity (HTTP 201) |
| `reset` | `(params: UpdateUserParams, opts?) => Promise<string>` | `POST /auth/reset` | Message string: `"Successfully updated {username}\n"` |

`register` uses `CreateUserParams` (same as `POST /users`). The `hash` field expects a pre-hashed password, not plaintext. `reset` uses `UpdateUserParams` (same as `PUT /users/`), which is a full user update -- it requires `user_id` to identify the target user plus all entity fields.

**`src/resources/sessions.ts`** -- Session management (no list/get/create):

| Method | Signature | HTTP |
|--------|-----------|------|
| `update` | `(params: UpdateSessionParams, opts?) => Promise<Session>` | `PUT /sessions/` |
| `remove` | `(id: SessionID, opts?) => Promise<void>` | `DELETE /sessions/?q={id}` |

**`src/resources/ssh-keys.ts`** -- SSH key management (no get-single, no update):

| Method | Signature | HTTP |
|--------|-----------|------|
| `list` | `(opts?) => Promise<SshKeyListItem[]>` | `GET /ssh-keys` |
| `create` | `(params: CreateSshKeyRequest, opts?) => Promise<SshKey>` | `POST /ssh-keys` |
| `remove` | `(id: string, opts?) => Promise<void>` | `DELETE /ssh-keys/{id}` |

SSH keys use a **path parameter** for delete (`/ssh-keys/{id}`), not a query parameter. This is the only resource with this pattern.

The GET list response omits the full public key. The SDK defines two types:
```typescript
// Full entity (returned by POST create)
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

// List item (returned by GET list -- no public_key)
type SshKeyListItem = {
  ssh_key_id: string
  key_type: string
  fingerprint: string
  label: string
  date_created: string
  last_used: string
}

// Create request (subset of fields)
type CreateSshKeyRequest = {
  public_key: string
  label: string
}
```

**`src/resources/media-upload.ts`** -- File upload:

| Method | Signature | HTTP |
|--------|-----------|------|
| `upload` | `(file: File \| Blob, opts?) => Promise<Media>` | `POST /mediaupload/` |

Sends `multipart/form-data` with a `file` field (max 10 MB). Uses `HttpClient.raw()` for the multipart request. Returns a `Media` entity (HTTP 200).

**Important:** The response is sent before S3 upload and image optimization complete asynchronously. The `srcset` field will be `null` in the initial response. Consumers needing the optimized srcset should poll `client.media.get(id)` after a delay.

**`src/resources/admin-tree.ts`** -- Admin content tree:

| Method | Signature | HTTP |
|--------|-----------|------|
| `get` | `(slug: Slug, format?: TreeFormat, opts?) => Promise<AdminTreeResponse \| Record<string, unknown>>` | `GET /admin/tree/{slug}?format={format}` |

Uses a **path parameter** for the slug (not query param). When `format` is `'raw'` or omitted, returns `AdminTreeResponse`. For other formats (contentful, sanity, etc.), returns `Record<string, unknown>` since those shapes are defined by third-party CMS conventions.

**`src/resources/import.ts`** -- CMS import:

| Method | Signature | HTTP |
|--------|-----------|------|
| `contentful` | `(data: Record<string, unknown>, opts?) => Promise<ImportResponse>` | `POST /import/contentful` |
| `sanity` | `(data: Record<string, unknown>, opts?) => Promise<ImportResponse>` | `POST /import/sanity` |
| `strapi` | `(data: Record<string, unknown>, opts?) => Promise<ImportResponse>` | `POST /import/strapi` |
| `wordpress` | `(data: Record<string, unknown>, opts?) => Promise<ImportResponse>` | `POST /import/wordpress` |
| `clean` | `(data: Record<string, unknown>, opts?) => Promise<ImportResponse>` | `POST /import/clean` |
| `bulk` | `(format: ImportFormat, data: Record<string, unknown>, opts?) => Promise<ImportResponse>` | `POST /import?format={fmt}` |

### Client Entry Point (`src/index.ts`)

```typescript
type ClientConfig = {
  baseUrl: string
  apiKey?: string
  defaultTimeout?: number             // ms, applied as AbortSignal.timeout() on all requests. Default: 30000 (30s)
  credentials?: RequestCredentials    // 'include' | 'same-origin' | 'omit'. Default: 'include'
  allowInsecure?: boolean             // must be true to use http:// baseUrl. Default: false
}

function createClient(config: ClientConfig): ModulaCMSClient
```

`createClient` validates config at construction time:
- Throws if `baseUrl` is not a valid URL
- Throws if `baseUrl` uses `http://` and `allowInsecure` is not explicitly `true` (prevents accidental plaintext credential transmission)
- Throws if `apiKey` is provided but empty string
- Defaults `defaultTimeout` to `30000` (30 seconds) if not provided -- every request gets an `AbortSignal.timeout()` to prevent indefinite hangs
- Defaults `credentials` to `'include'` for cookie auth, but consumers can set `'same-origin'` or `'omit'` based on their deployment context

**Security note:** `apiKey` should only be used in server-side (Node.js) contexts. Never embed API keys in browser bundles -- use cookie authentication for browser clients instead.

**`ModulaCMSClient` type** -- the complete client shape:

```typescript
type ModulaCMSClient = {
  // Auth (special resource)
  auth: {
    login: (params: LoginRequest, opts?: RequestOptions) => Promise<LoginResponse>
    logout: (opts?: RequestOptions) => Promise<void>
    me: (opts?: RequestOptions) => Promise<MeResponse>
    register: (params: CreateUserParams, opts?: RequestOptions) => Promise<User>
    reset: (params: UpdateUserParams, opts?: RequestOptions) => Promise<string>
  }

  // CRUD resources (via factory)
  adminRoutes: CrudResource<AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams, Slug> & {
    listOrdered: (opts?: RequestOptions) => Promise<AdminRoute[]>
    remove: (id: AdminRouteID, opts?: RequestOptions) => Promise<void>  // overrides Slug default
  }
  adminContentData: CrudResource<AdminContentData, CreateAdminContentDataParams, UpdateAdminContentDataParams, AdminContentID>
  adminContentFields: CrudResource<AdminContentField, CreateAdminContentFieldParams, UpdateAdminContentFieldParams, AdminContentFieldID>
  adminDatatypes: CrudResource<AdminDatatype, CreateAdminDatatypeParams, UpdateAdminDatatypeParams, AdminDatatypeID>
  adminFields: CrudResource<AdminField, CreateAdminFieldParams, UpdateAdminFieldParams, AdminFieldID>
  contentData: CrudResource<ContentData, CreateContentDataParams, UpdateContentDataParams, ContentID>
  contentFields: CrudResource<ContentField, CreateContentFieldParams, UpdateContentFieldParams, ContentFieldID>
  datatypes: CrudResource<Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID>
  fields: CrudResource<Field, CreateFieldParams, UpdateFieldParams, FieldID>
  routes: CrudResource<Route, CreateRouteParams, UpdateRouteParams, RouteID>
  media: CrudResource<Media, CreateMediaParams, UpdateMediaParams, MediaID>
  mediaDimensions: CrudResource<MediaDimension, CreateMediaDimensionParams, UpdateMediaDimensionParams>
  users: CrudResource<User, CreateUserParams, UpdateUserParams, UserID>
  roles: CrudResource<Role, CreateRoleParams, UpdateRoleParams, RoleID>
  tokens: CrudResource<Token, CreateTokenParams, UpdateTokenParams>
  usersOauth: CrudResource<UserOauth, CreateUserOauthParams, UpdateUserOauthParams, UserOauthID>
  tables: CrudResource<Table, CreateTableParams, UpdateTableParams>

  // Special resources
  adminTree: {
    get: (slug: Slug, format?: TreeFormat, opts?: RequestOptions) => Promise<AdminTreeResponse | Record<string, unknown>>
  }
  mediaUpload: {
    upload: (file: File | Blob, opts?: RequestOptions) => Promise<Media>
  }
  sessions: {
    update: (params: UpdateSessionParams, opts?: RequestOptions) => Promise<Session>
    remove: (id: SessionID, opts?: RequestOptions) => Promise<void>
  }
  sshKeys: {
    list: (opts?: RequestOptions) => Promise<SshKeyListItem[]>
    create: (params: CreateSshKeyRequest, opts?: RequestOptions) => Promise<SshKey>
    remove: (id: string, opts?: RequestOptions) => Promise<void>
  }
  import: {
    contentful: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    sanity: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    strapi: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    wordpress: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    clean: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    bulk: (format: ImportFormat, data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
  }
}
```

### Type Mapping (Go -> TypeScript)

All ID types use branded types to prevent cross-domain misuse. An admin SDK wiring together 17+ entity types with plain string IDs everywhere is a misuse magnet -- branded types catch bugs like passing a `UserID` where a `RouteID` is expected at compile time, with zero runtime cost.

```typescript
type Brand<T, B extends string> = T & { readonly __brand: B }

// Each Go ID type becomes a distinct branded string
type UserID = Brand<string, 'UserID'>
type AdminContentID = Brand<string, 'AdminContentID'>
type AdminContentFieldID = Brand<string, 'AdminContentFieldID'>
type AdminContentRelationID = Brand<string, 'AdminContentRelationID'>
type AdminDatatypeID = Brand<string, 'AdminDatatypeID'>
type AdminFieldID = Brand<string, 'AdminFieldID'>
type AdminRouteID = Brand<string, 'AdminRouteID'>
type ContentID = Brand<string, 'ContentID'>
type ContentFieldID = Brand<string, 'ContentFieldID'>
type ContentRelationID = Brand<string, 'ContentRelationID'>
type DatatypeID = Brand<string, 'DatatypeID'>
type FieldID = Brand<string, 'FieldID'>
type MediaID = Brand<string, 'MediaID'>
type RoleID = Brand<string, 'RoleID'>
type RouteID = Brand<string, 'RouteID'>
type SessionID = Brand<string, 'SessionID'>
type UserOauthID = Brand<string, 'UserOauthID'>
type Slug = Brand<string, 'Slug'>
type Email = Brand<string, 'Email'>
type URL = Brand<string, 'URL'>
```

| Go Type | TypeScript |
|---------|-----------|
| `types.AdminContentID`, `types.UserID`, etc. | Branded `string` per ID domain (see list above) |
| `types.NullableContentID`, `types.NullableUserID`, etc. | `BrandedID \| null` (e.g., `AdminContentID \| null`) |
| `sql.NullString` | `string \| null` |
| `sql.NullInt64` | `number \| null` |
| `types.Timestamp` | `string` (RFC 3339, not branded -- RFC 3339 compliance is a server responsibility, not a client-side type constraint) |
| `types.ContentStatus` | `'draft' \| 'published' \| 'archived' \| 'pending'` |
| `types.FieldType` | `'text' \| 'textarea' \| 'number' \| 'date' \| 'datetime' \| 'boolean' \| 'select' \| 'media' \| 'relation' \| 'json' \| 'richtext' \| 'slug' \| 'email' \| 'url'` |
| `types.Slug` | `Brand<string, 'Slug'>` |
| `types.Email` | `Brand<string, 'Email'>` |
| `types.URL` | `Brand<string, 'URL'>` |
| `types.JSONData` | `Record<string, unknown>` |
| `int64` | `number` |
| `bool` | `boolean` |

JSON field names use `snake_case` matching the API.

### Auth Endpoint Types (`src/types/auth.ts`)

```typescript
type LoginRequest = {
  email: Email
  password: string
}

type LoginResponse = {
  user_id: UserID
  email: Email
  username: string
  created_at: string  // Timestamp
}

type MeResponse = {
  user_id: UserID
  email: Email
  username: string
  name: string
  role: string
}
```

`auth.register` uses `CreateUserParams` directly (defined in `users.ts`). The `hash` field expects a **pre-hashed password**, not plaintext. `auth.register` returns `Promise<User>` (HTTP 201).

`auth.reset` uses `UpdateUserParams` directly (defined in `users.ts`). This is a full user update that requires `user_id` to identify the target user. `auth.reset` returns `Promise<string>` -- the server responds with a message: `"Successfully updated {username}\n"`.

`auth.logout` returns `Promise<void>`. The server responds with `{message: "Logged out successfully"}` but the SDK discards it.

OAuth endpoints (`GET /auth/oauth/login`, `GET /auth/oauth/callback`) are browser-redirect flows, not JSON API calls. The SDK deliberately excludes them -- they are initiated by navigating the browser to the URL, not by calling an SDK method.

### Import Endpoint Types (`src/types/import.ts`)

```typescript
type ImportFormat = 'contentful' | 'sanity' | 'strapi' | 'wordpress' | 'clean'

type ImportResponse = {
  success: boolean
  datatypes_created: number
  fields_created: number
  content_created: number
  message: string
  errors: string[]
}
```

Import request bodies are format-specific JSON payloads. The SDK accepts `Record<string, unknown>` for import request bodies since the schema varies by format and is defined by the source CMS. Invalid payloads are rejected by the server with a 400 response.

### Admin Tree Response Type (`src/types/tree.ts`)

The `GET /admin/tree/{slug}` endpoint returns a nested content tree. The response is assembled from multiple database queries (content data, datatypes, fields, field values). Based on the `GetRouteTreeByRouteIDRow` struct:

```typescript
type ContentTreeField = {
  field_label: string
  field_type: FieldType
  field_value: string | null
}

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

type AdminTreeResponse = {
  route: AdminRoute
  tree: ContentTreeNode[]
}

type TreeFormat = 'contentful' | 'sanity' | 'strapi' | 'wordpress' | 'clean' | 'raw'
```

**Note:** The exact response shape depends on the `?format=` parameter. The `raw` format returns the structure above. Other formats (contentful, sanity, etc.) transform the tree into the target CMS's schema. The SDK types cover the `raw` format; other formats return `Record<string, unknown>` since their shapes are defined by third-party CMS conventions.

### Error Handling

```typescript
type ApiError = {
  readonly _tag: 'ApiError'  // discriminant -- prevents false positives from other objects with status + message
  status: number
  message: string
  body?: unknown  // raw parsed JSON from error response, or undefined if response was not JSON
}

function isApiError(err: unknown): err is ApiError {
  return (
    typeof err === 'object' &&
    err !== null &&
    '_tag' in err &&
    (err as ApiError)._tag === 'ApiError'
  )
}
```

Thrown on any non-2xx response. The `http.ts` layer constructs `ApiError` objects with `_tag: 'ApiError'` set. The `_tag` discriminant is necessary because SDK consumers frequently use multiple HTTP libraries (Axios, native fetch, etc.) in the same codebase. Without it, any object with `{status: number, message: string}` would pass the type guard -- including Axios errors and Node HTTP response objects.

When the response Content-Type is not JSON (e.g., HTML 502 from a proxy), `body` is `undefined` and `message` is set to the HTTP status text. The SDK does not parse or validate error response body structure -- `body` contains the raw parsed JSON. Common server error shapes include `{error: string}` for auth errors and `{errors: [{field, message}]}` for validation errors, but consumers should narrow the type themselves.

Since the SDK uses functional style (no classes), `ApiError` is a plain object -- callers use the exported `isApiError()` type guard to narrow in catch blocks instead of `instanceof`.

Network errors (connection reset, DNS failure) are not wrapped -- they propagate as native errors. Callers can distinguish: `isApiError(err)` means an HTTP response was received; `!isApiError(err)` means the request never completed.

---

## Deliberately Excluded

The following entities exist in `STRUCT_REFERENCE.md` but have no corresponding API endpoints in `API_CONTRACT.md`. They are excluded from the SDK's resource layer but their types are still defined in `src/types/` for use in nested responses (e.g., admin tree):

| Entity | Reason for Exclusion | Types Defined? |
|--------|---------------------|----------------|
| Backup, BackupSet, BackupVerification | Internal server operations, no API endpoints | No |
| ChangeEvent | Internal audit log, no API endpoints | No |
| Permissions | No CRUD API endpoints exposed | No |
| AdminDatatypeFields | Junction table, no standalone endpoint. Appears in tree responses. | Yes (in `admin.ts`) |
| AdminContentRelations | No standalone endpoint. Appears in tree responses. | Yes (in `admin.ts`) |
| ContentRelations | No standalone endpoint. Appears in tree responses. | Yes (in `content.ts`) |
| DatatypeFields | Junction table, no standalone endpoint. Appears in tree responses. | Yes (in `schema.ts`) |

**Content delivery endpoint** (`GET /{slug}`) is excluded per the "admin scope only" design choice. This is a public-facing endpoint for frontend content consumption, not admin management.

**OAuth redirect endpoints** (`GET /auth/oauth/login`, `GET /auth/oauth/callback`) are browser-redirect flows, not JSON API calls. They are initiated by navigating to the URL directly, not through SDK method calls.

**Bulk tree save** is not included in v0.1. The ModulaCMS backend will add endpoints for posting a raw tree and running transactional updates. The SDK will add corresponding methods when those endpoints are available.

---

## Implementation Order

1. **Project setup** - `package.json`, `tsconfig.json`, bun config
2. **Common types** - `src/types/common.ts` (Brand utility, all branded ID types, ContentStatus, FieldType, error shapes)
3. **Entity types** - All files in `src/types/` (translate every struct from STRUCT_REFERENCE.md, including create/update params)
4. **HTTP layer** - `src/http.ts`
5. **CRUD factory** - `src/resource.ts`
6. **Special resources** - `src/resources/auth.ts`, `admin-tree.ts`, `media-upload.ts`, `sessions.ts`, `ssh-keys.ts`, `import.ts`
7. **Client factory** - `src/index.ts` (wires everything together, defines `ModulaCMSClient` type)
8. **Build verification** - `bun run build`, confirm types export correctly

---

## Build / Publish Config

**package.json:**
- `name`: `modulacms-sdk`
- `version`: `0.1.0`
- `type`: `module`
- `main`: `./dist/index.js`
- `types`: `./dist/index.d.ts`
- `sideEffects`: `false` (enables tree-shaking)
- `files`: `["dist"]` (only publish built output)
- `scripts`:
  - `build`: `tsc`
  - `typecheck`: `tsc --noEmit`
- Zero runtime dependencies

**`"exports"` map:**
```json
{
  ".": {
    "types": "./dist/index.d.ts",
    "default": "./dist/index.js"
  }
}
```

**ESM-only build pipeline:** `tsc` emits both `.js` (ESM) and `.d.ts` (declarations) in a single pass. No bundler step required. Source files use `.js` extensions on relative imports (e.g., `import { foo } from './http.js'`) for Node.js ESM compatibility. CJS is deliberately unsupported -- all target environments (Node >= 18, modern browsers, bundlers) have native ESM support.

**tsconfig.json:**
- `target`: `ES2022`
- `module`: `NodeNext`
- `moduleResolution`: `nodenext`
- `strict`: `true`
- `declaration`: `true`
- `outDir`: `dist`

---

## Security Considerations

### Transport Security
- `createClient` requires HTTPS by default. HTTP URLs are rejected unless `allowInsecure: true` is explicitly set. This prevents accidental plaintext transmission of API keys and session cookies.

### Credential Handling
- **API keys** should only be used in server-side (Node.js) contexts. The SDK documentation must warn against embedding API keys in browser bundles. Browser clients should use cookie authentication instead.
- **`credentials` is configurable** (default `'include'`). Consumers deploying the SDK cross-origin or in contexts where cookie leakage is a concern can set `'same-origin'` or `'omit'`.

### Sensitive Response Data
- The `Token` entity type includes the raw `token` string (API key value). The `UserOauth` entity type includes `access_token` and `refresh_token` (OAuth secrets). Consumers must treat responses from `client.tokens` and `client.usersOauth` as sensitive -- do not log, serialize to localStorage, or expose in client-side state without redaction.

### Trust Model
- The SDK assumes a trusted, first-party, version-matched ModulaCMS server. Server responses are not validated at runtime. If the server is compromised, returns unexpected shapes, or runs a different API version, the SDK will pass through malformed data as typed objects. Consumers requiring defense against untrusted servers should use the `raw` method and perform their own validation.

### Request Safety
- All requests have a default 30-second timeout via `AbortSignal.timeout()` to prevent resource exhaustion from hung connections. The timeout is preserved even when consumers provide a custom `AbortSignal` (merged via `AbortSignal.any()`).
- The HTTP layer validates `Content-Type: application/json` on responses before JSON parsing to produce clear error messages when proxies return HTML.

---

## Verification

1. `bun run build` succeeds with no errors
2. `bun run typecheck` passes
3. Import `createClient` from built output and confirm types are available
4. Manual smoke test: create a client, call `client.auth.me()` against a running ModulaCMS instance
