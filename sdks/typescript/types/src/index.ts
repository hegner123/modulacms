/**
 * @modulacms/types -- shared type definitions for all ModulaCMS TypeScript SDKs.
 *
 * @packageDocumentation
 */

// Brand utility
export type { Brand } from './brand.js'

// Branded ID types
export type {
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
  PermissionID,
  RolePermissionID,
  RouteID,
  SessionID,
  UserOauthID,
  Slug,
  Email,
  URL,
} from './ids.js'

// Common primitives
export type { ULID, Timestamp, NullableString, NullableNumber } from './common.js'

// Enums and runtime constants
export type { ContentStatus, FieldType, ContentFormat } from './enums.js'
export { CONTENT_FORMATS } from './enums.js'

// Pagination
export type { PaginationParams, PaginatedResponse } from './pagination.js'

// Errors
export type { ApiError } from './errors.js'
export { isApiError } from './errors.js'

// Request options
export type { RequestOptions } from './request.js'

// Entity types
export type { ContentData, ContentField, ContentRelation } from './entities/content.js'
export type { Datatype, Field, DatatypeField } from './entities/schema.js'
export type { Media, MediaDimension } from './entities/media.js'
export type { Route } from './entities/routing.js'
export type { NodeDatatype, NodeField, ContentNode, ContentTree } from './entities/tree.js'
