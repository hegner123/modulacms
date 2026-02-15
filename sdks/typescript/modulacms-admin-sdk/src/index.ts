/**
 * ModulaCMS Admin SDK entry point.
 *
 * Provides the {@link createAdminClient} factory function which returns a fully typed
 * {@link ModulaCMSAdminClient} with CRUD operations for all CMS resources, plus
 * specialized endpoints for authentication, media upload, content trees, and imports.
 *
 * @example
 * ```ts
 * import { createAdminClient, isApiError } from 'modulacms-admin-sdk'
 *
 * const client = createAdminClient({
 *   baseUrl: 'https://cms.example.com',
 *   apiKey: 'your-api-key',
 * })
 *
 * try {
 *   const users = await client.users.list()
 * } catch (err) {
 *   if (isApiError(err)) {
 *     console.error(`API ${err.status}: ${err.message}`)
 *   }
 * }
 * ```
 *
 * @packageDocumentation
 * @module modulacms-admin-sdk
 */

import { createHttpClient } from './http.js'
import { createResource } from './resource.js'
import { createAuthResource } from './resources/auth.js'
import { createAdminTreeResource } from './resources/admin-tree.js'
import { createMediaUploadResource } from './resources/media-upload.js'
import { createSessionsResource } from './resources/sessions.js'
import { createSshKeysResource } from './resources/ssh-keys.js'
import { createImportResource } from './resources/import.js'

import type { RequestOptions, AdminRouteID, Slug } from './types/common.js'
import type { AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams } from './types/admin.js'
import type { CrudResource } from './resource.js'

// Re-export all public types
export type { CrudResource } from './resource.js'
export { isApiError } from './types/common.js'
export type {
  Brand,
  UserID,
  AdminContentID,
  AdminContentFieldID,
  AdminContentRelationID,
  AdminDatatypeID,
  AdminFieldID,
  AdminRouteID,
  ContentID,
  ContentFieldID,
  ContentRelationID,
  DatatypeID,
  FieldID,
  MediaID,
  RoleID,
  RouteID,
  SessionID,
  UserOauthID,
  Slug,
  Email,
  URL,
  ContentStatus,
  FieldType,
  RequestOptions,
  ApiError,
  PaginationParams,
  PaginatedResponse,
} from './types/common.js'
export type { LoginRequest, LoginResponse, MeResponse } from './types/auth.js'
export type {
  AdminRoute,
  AdminContentData,
  AdminContentField,
  AdminDatatype,
  AdminField,
  AdminDatatypeField,
  AdminContentRelation,
  CreateAdminRouteParams,
  CreateAdminContentDataParams,
  CreateAdminContentFieldParams,
  CreateAdminDatatypeParams,
  CreateAdminFieldParams,
  CreateAdminDatatypeFieldParams,
  UpdateAdminRouteParams,
  UpdateAdminContentDataParams,
  UpdateAdminContentFieldParams,
  UpdateAdminDatatypeParams,
  UpdateAdminFieldParams,
  UpdateAdminDatatypeFieldParams,
} from './types/admin.js'
export type {
  ContentData,
  ContentField,
  ContentRelation,
  CreateContentDataParams,
  CreateContentFieldParams,
  UpdateContentDataParams,
  UpdateContentFieldParams,
} from './types/content.js'
export type {
  Datatype,
  Field,
  DatatypeField,
  CreateDatatypeParams,
  CreateDatatypeFieldParams,
  CreateFieldParams,
  UpdateDatatypeParams,
  UpdateDatatypeFieldParams,
  UpdateFieldParams,
} from './types/schema.js'
export type {
  Media,
  MediaDimension,
  CreateMediaParams,
  CreateMediaDimensionParams,
  UpdateMediaParams,
  UpdateMediaDimensionParams,
} from './types/media.js'
export type {
  User,
  Role,
  Token,
  UserOauth,
  Session,
  SshKey,
  SshKeyListItem,
  CreateUserParams,
  CreateRoleParams,
  CreateTokenParams,
  CreateUserOauthParams,
  CreateSshKeyRequest,
  UpdateUserParams,
  UpdateRoleParams,
  UpdateTokenParams,
  UpdateUserOauthParams,
  UpdateSessionParams,
} from './types/users.js'
export type { Route, CreateRouteParams, UpdateRouteParams } from './types/routing.js'
export type { Table, CreateTableParams, UpdateTableParams } from './types/tables.js'
export type { ImportFormat, ImportResponse } from './types/import.js'
export type {
  ContentTreeField,
  ContentTreeNode,
  AdminTreeResponse,
  TreeFormat,
} from './types/tree.js'

// ---------------------------------------------------------------------------
// Imports for client type (type-only, for the ModulaCMSAdminClient shape)
// ---------------------------------------------------------------------------

import type {
  AdminContentData,
  AdminContentField,
  AdminDatatype,
  AdminField,
  AdminDatatypeField,
  CreateAdminContentDataParams,
  CreateAdminContentFieldParams,
  CreateAdminDatatypeParams,
  CreateAdminFieldParams,
  CreateAdminDatatypeFieldParams,
  UpdateAdminContentDataParams,
  UpdateAdminContentFieldParams,
  UpdateAdminDatatypeParams,
  UpdateAdminFieldParams,
  UpdateAdminDatatypeFieldParams,
} from './types/admin.js'
import type {
  ContentData,
  ContentField,
  CreateContentDataParams,
  CreateContentFieldParams,
  UpdateContentDataParams,
  UpdateContentFieldParams,
} from './types/content.js'
import type {
  Datatype,
  Field,
  DatatypeField,
  CreateDatatypeParams,
  CreateDatatypeFieldParams,
  CreateFieldParams,
  UpdateDatatypeParams,
  UpdateDatatypeFieldParams,
  UpdateFieldParams,
} from './types/schema.js'
import type {
  Media,
  MediaDimension,
  CreateMediaParams,
  CreateMediaDimensionParams,
  UpdateMediaParams,
  UpdateMediaDimensionParams,
} from './types/media.js'
import type {
  User,
  Role,
  Token,
  UserOauth,
  CreateUserParams,
  CreateRoleParams,
  CreateTokenParams,
  CreateUserOauthParams,
  UpdateUserParams,
  UpdateRoleParams,
  UpdateTokenParams,
  UpdateUserOauthParams,
} from './types/users.js'
import type { Route, CreateRouteParams, UpdateRouteParams } from './types/routing.js'
import type { Table, CreateTableParams, UpdateTableParams } from './types/tables.js'
import type {
  AdminContentID,
  AdminContentFieldID,
  AdminDatatypeID,
  AdminFieldID,
  ContentID,
  ContentFieldID,
  DatatypeID,
  FieldID,
  MediaID,
  RoleID,
  RouteID,
  UserID,
  UserOauthID,
} from './types/common.js'
import type { LoginRequest, LoginResponse, MeResponse } from './types/auth.js'
import type { Session, UpdateSessionParams, SshKey, SshKeyListItem, CreateSshKeyRequest } from './types/users.js'
import type { SessionID } from './types/common.js'
import type { AdminTreeResponse, TreeFormat } from './types/tree.js'
import type { ImportFormat, ImportResponse } from './types/import.js'

// ---------------------------------------------------------------------------
// Client config
// ---------------------------------------------------------------------------

/**
 * Configuration options for creating a {@link ModulaCMSAdminClient}.
 *
 * @example
 * ```ts
 * const config: ClientConfig = {
 *   baseUrl: 'https://cms.example.com',
 *   apiKey: 'sk_live_abc123',
 *   defaultTimeout: 15000,
 * }
 * ```
 */
export type ClientConfig = {
  /**
   * Base URL of the ModulaCMS server (e.g. `'https://cms.example.com'`).
   * Must be HTTPS unless {@link allowInsecure} is set to `true`.
   */
  baseUrl: string
  /**
   * API key for Bearer token authentication.
   * Intended for server-side use only; do not embed in client-side code.
   * Omit for cookie-based authentication.
   */
  apiKey?: string
  /**
   * Default request timeout in milliseconds.
   * @defaultValue 30000
   */
  defaultTimeout?: number
  /**
   * Fetch `credentials` mode for cookie handling.
   * @defaultValue `'include'`
   */
  credentials?: RequestCredentials
  /**
   * Allow `http://` base URLs. Required for local development.
   * Set to `true` to suppress the HTTPS enforcement error.
   *
   * @remarks Do not enable in production. HTTP transmits credentials in plaintext.
   * @defaultValue `false`
   */
  allowInsecure?: boolean
}

// ---------------------------------------------------------------------------
// ModulaCMSAdminClient type
// ---------------------------------------------------------------------------

/**
 * The main SDK client providing typed access to all ModulaCMS API resources.
 *
 * Created via {@link createAdminClient}. Each property corresponds to an API resource
 * with standard CRUD operations or specialized endpoints.
 *
 * @example
 * ```ts
 * const client = createAdminClient({ baseUrl: 'https://cms.example.com', apiKey: 'key' })
 *
 * // Standard CRUD
 * const users = await client.users.list()
 * const user = await client.users.get(userId)
 *
 * // Authentication
 * const me = await client.auth.me()
 *
 * // Media upload
 * const media = await client.mediaUpload.upload(file)
 *
 * // Content tree
 * const tree = await client.adminTree.get(slug)
 *
 * // Bulk import
 * const result = await client.import.contentful(exportData)
 * ```
 */
export type ModulaCMSAdminClient = {
  /** Authentication endpoints (login, logout, me, register, reset). */
  auth: {
    /** Authenticate with email and password. */
    login: (params: LoginRequest, opts?: RequestOptions) => Promise<LoginResponse>
    /** End the current session. */
    logout: (opts?: RequestOptions) => Promise<void>
    /** Get the currently authenticated user's identity. */
    me: (opts?: RequestOptions) => Promise<MeResponse>
    /** Register a new user account. */
    register: (params: CreateUserParams, opts?: RequestOptions) => Promise<User>
    /** Reset a user's password or account details. */
    reset: (params: UpdateUserParams, opts?: RequestOptions) => Promise<string>
  }

  /**
   * Admin routes resource with split-identifier pattern.
   * Uses {@link Slug} for `get` and `list`, {@link AdminRouteID} for `remove`.
   */
  adminRoutes: Omit<CrudResource<AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams, Slug>, 'remove'> & {
    /** List all admin routes sorted by server-defined order. */
    listOrdered: (opts?: RequestOptions) => Promise<AdminRoute[]>
    /** Remove an admin route by its unique ID. */
    remove: (id: AdminRouteID, opts?: RequestOptions) => Promise<void>
  }
  /** Admin content data nodes (tree structure). */
  adminContentData: CrudResource<AdminContentData, CreateAdminContentDataParams, UpdateAdminContentDataParams, AdminContentID>
  /** Admin content field values. */
  adminContentFields: CrudResource<AdminContentField, CreateAdminContentFieldParams, UpdateAdminContentFieldParams, AdminContentFieldID>
  /** Admin datatype definitions. */
  adminDatatypes: CrudResource<AdminDatatype, CreateAdminDatatypeParams, UpdateAdminDatatypeParams, AdminDatatypeID>
  /** Admin field definitions. */
  adminFields: CrudResource<AdminField, CreateAdminFieldParams, UpdateAdminFieldParams, AdminFieldID>
  /** Public content data nodes (tree structure). */
  contentData: CrudResource<ContentData, CreateContentDataParams, UpdateContentDataParams, ContentID>
  /** Public content field values. */
  contentFields: CrudResource<ContentField, CreateContentFieldParams, UpdateContentFieldParams, ContentFieldID>
  /** Datatype (content type) schema definitions. */
  datatypes: CrudResource<Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID>
  /** Field schema definitions. */
  fields: CrudResource<Field, CreateFieldParams, UpdateFieldParams, FieldID>
  /** Public routes. */
  routes: CrudResource<Route, CreateRouteParams, UpdateRouteParams, RouteID>
  /** Media assets. */
  media: CrudResource<Media, CreateMediaParams, UpdateMediaParams, MediaID>
  /** Media dimension presets for responsive rendering. */
  mediaDimensions: CrudResource<MediaDimension, CreateMediaDimensionParams, UpdateMediaDimensionParams>
  /** User accounts. */
  users: CrudResource<User, CreateUserParams, UpdateUserParams, UserID>
  /** Permission roles. */
  roles: CrudResource<Role, CreateRoleParams, UpdateRoleParams, RoleID>
  /** API tokens. SENSITIVE - contains bearer credentials. */
  tokens: CrudResource<Token, CreateTokenParams, UpdateTokenParams>
  /** OAuth connections. SENSITIVE - contains OAuth tokens. */
  usersOauth: CrudResource<UserOauth, CreateUserOauthParams, UpdateUserOauthParams, UserOauthID>
  /** Datatype-field junction records (schema-level). */
  datatypeFields: CrudResource<DatatypeField, CreateDatatypeFieldParams, UpdateDatatypeFieldParams>
  /** Admin datatype-field junction records. */
  adminDatatypeFields: CrudResource<AdminDatatypeField, CreateAdminDatatypeFieldParams, UpdateAdminDatatypeFieldParams>
  /** Named tables. */
  tables: CrudResource<Table, CreateTableParams, UpdateTableParams>

  /** Admin content tree retrieval with optional format conversion. */
  adminTree: {
    /** Get the full content tree for an admin route by slug. */
    get: (slug: Slug, format?: TreeFormat, opts?: RequestOptions) => Promise<AdminTreeResponse | Record<string, unknown>>
  }
  /** File upload to the media library via multipart/form-data. */
  mediaUpload: {
    /** Upload a file and receive the created media entity. */
    upload: (file: File | Blob, opts?: RequestOptions) => Promise<Media>
  }
  /** Session management (update and remove only; sessions are created via login). */
  sessions: {
    /** Update session metadata. */
    update: (params: UpdateSessionParams, opts?: RequestOptions) => Promise<Session>
    /** Invalidate a session. */
    remove: (id: SessionID, opts?: RequestOptions) => Promise<void>
  }
  /** SSH key management (list, create, remove). */
  sshKeys: {
    /** List SSH keys (summaries without public key material). */
    list: (opts?: RequestOptions) => Promise<SshKeyListItem[]>
    /** Register a new SSH public key. */
    create: (params: CreateSshKeyRequest, opts?: RequestOptions) => Promise<SshKey>
    /** Remove an SSH key by ID. */
    remove: (id: string, opts?: RequestOptions) => Promise<void>
  }
  /** Bulk content import from external CMS platforms. */
  import: {
    /** Import from Contentful export format. */
    contentful: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    /** Import from Sanity.io export format. */
    sanity: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    /** Import from Strapi export format. */
    strapi: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    /** Import from WordPress export format. */
    wordpress: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    /** Import from ModulaCMS clean format. */
    clean: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
    /** Import using a dynamic format specifier. */
    bulk: (format: ImportFormat, data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
  }
}

// ---------------------------------------------------------------------------
// createAdminClient factory
// ---------------------------------------------------------------------------

/**
 * Create a configured ModulaCMS client instance.
 *
 * Validates the configuration, sets up the HTTP layer with authentication
 * and timeout handling, and returns a {@link ModulaCMSAdminClient} with typed
 * access to all API resources.
 *
 * @param config - Client configuration (base URL, API key, timeout, etc.).
 * @returns A fully configured {@link ModulaCMSAdminClient}.
 * @throws `Error` if `baseUrl` is not a valid URL.
 * @throws `Error` if `baseUrl` uses `http://` without `allowInsecure: true`.
 * @throws `Error` if `apiKey` is an empty string.
 *
 * @example
 * ```ts
 * // Production
 * const client = createAdminClient({
 *   baseUrl: 'https://cms.example.com',
 *   apiKey: process.env.CMS_API_KEY,
 * })
 *
 * // Local development
 * const devClient = createAdminClient({
 *   baseUrl: 'http://localhost:8080',
 *   allowInsecure: true,
 * })
 * ```
 */
export function createAdminClient(config: ClientConfig): ModulaCMSAdminClient {
  // Validate baseUrl
  let parsed: globalThis.URL
  try {
    parsed = new globalThis.URL(config.baseUrl)
  } catch {
    throw new Error(`Invalid baseUrl: ${config.baseUrl}`)
  }

  if (parsed.protocol === 'http:' && config.allowInsecure !== true) {
    throw new Error(
      'baseUrl uses http:// which transmits credentials in plaintext. ' +
      'Set allowInsecure: true to allow this, or use https://.',
    )
  }

  if (config.apiKey !== undefined && config.apiKey === '') {
    throw new Error('apiKey must not be an empty string')
  }

  const defaultTimeout = config.defaultTimeout ?? 30000
  const credentials = config.credentials ?? 'include'

  // Strip trailing slash from baseUrl for consistent URL building
  const baseUrl = config.baseUrl.replace(/\/+$/, '')

  const http = createHttpClient({
    baseUrl,
    ...(config.apiKey !== undefined ? { apiKey: config.apiKey } : {}),
    defaultTimeout,
    credentials,
  })

  // Build admin routes with split-identifier pattern
  const baseAdminRoutes = createResource<AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams, Slug>(
    http,
    'adminroutes',
  )

  const adminRoutes: ModulaCMSAdminClient['adminRoutes'] = {
    ...baseAdminRoutes,

    listOrdered(opts?: RequestOptions): Promise<AdminRoute[]> {
      return http.get<AdminRoute[]>('/adminroutes', { ordered: 'true' }, opts)
    },

    remove(id: AdminRouteID, opts?: RequestOptions): Promise<void> {
      return http.del('/adminroutes/', { q: String(id) }, opts)
    },
  }

  return {
    auth: createAuthResource(http),
    adminRoutes,
    adminContentData: createResource<AdminContentData, CreateAdminContentDataParams, UpdateAdminContentDataParams, AdminContentID>(http, 'admincontentdatas'),
    adminContentFields: createResource<AdminContentField, CreateAdminContentFieldParams, UpdateAdminContentFieldParams, AdminContentFieldID>(http, 'admincontentfields'),
    adminDatatypes: createResource<AdminDatatype, CreateAdminDatatypeParams, UpdateAdminDatatypeParams, AdminDatatypeID>(http, 'admindatatypes'),
    adminFields: createResource<AdminField, CreateAdminFieldParams, UpdateAdminFieldParams, AdminFieldID>(http, 'adminfields'),
    contentData: createResource<ContentData, CreateContentDataParams, UpdateContentDataParams, ContentID>(http, 'contentdata'),
    contentFields: createResource<ContentField, CreateContentFieldParams, UpdateContentFieldParams, ContentFieldID>(http, 'contentfields'),
    datatypes: createResource<Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID>(http, 'datatype'),
    fields: createResource<Field, CreateFieldParams, UpdateFieldParams, FieldID>(http, 'fields'),
    routes: createResource<Route, CreateRouteParams, UpdateRouteParams, RouteID>(http, 'routes'),
    media: createResource<Media, CreateMediaParams, UpdateMediaParams, MediaID>(http, 'media'),
    mediaDimensions: createResource<MediaDimension, CreateMediaDimensionParams, UpdateMediaDimensionParams>(http, 'mediadimensions'),
    users: createResource<User, CreateUserParams, UpdateUserParams, UserID>(http, 'users'),
    roles: createResource<Role, CreateRoleParams, UpdateRoleParams, RoleID>(http, 'roles'),
    tokens: createResource<Token, CreateTokenParams, UpdateTokenParams>(http, 'tokens'),
    usersOauth: createResource<UserOauth, CreateUserOauthParams, UpdateUserOauthParams, UserOauthID>(http, 'usersoauth'),
    datatypeFields: createResource<DatatypeField, CreateDatatypeFieldParams, UpdateDatatypeFieldParams>(http, 'datatypefields'),
    adminDatatypeFields: createResource<AdminDatatypeField, CreateAdminDatatypeFieldParams, UpdateAdminDatatypeFieldParams>(http, 'admindatatypefields'),
    tables: createResource<Table, CreateTableParams, UpdateTableParams>(http, 'tables'),
    adminTree: createAdminTreeResource(http),
    mediaUpload: createMediaUploadResource(http, defaultTimeout, credentials, config.apiKey),
    sessions: createSessionsResource(http),
    sshKeys: createSshKeysResource(http),
    import: createImportResource(http),
  }
}
