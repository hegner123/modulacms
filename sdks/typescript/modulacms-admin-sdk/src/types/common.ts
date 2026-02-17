/**
 * Re-exports shared types from @modulacms/types.
 * This file preserves the import path `./common.js` used throughout the admin SDK.
 *
 * @module types/common
 */

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
  PermissionID,
  RolePermissionID,
  RouteID,
  SessionID,
  UserOauthID,
  Slug,
  Email,
  URL,
  ContentStatus,
  FieldType,
  PaginationParams,
  PaginatedResponse,
  RequestOptions,
  ApiError,
} from '@modulacms/types'

export { isApiError } from '@modulacms/types'
