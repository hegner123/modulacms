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
  AdminContentVersionID,
  AdminDatatypeID,
  AdminFieldID,
  AdminRouteID,
  ContentID,
  ContentFieldID,
  ContentRelationID,
  ContentVersionID,
  LocaleID,
  DatatypeID,
  FieldID,
  MediaID,
  MediaFolderID,
  AdminMediaID,
  AdminMediaFolderID,
  RoleID,
  PermissionID,
  RolePermissionID,
  FieldTypeID,
  AdminFieldTypeID,
  RouteID,
  SessionID,
  UserOauthID,
  WebhookID,
  WebhookDeliveryID,
  ValidationID,
  AdminValidationID,
  Slug,
  Email,
  URL,
  ContentStatus,
  ContentFormat,
  FieldType,
  PaginationParams,
  PaginatedResponse,
  RequestOptions,
  ApiError,
} from '@modulacms/types'

export { isApiError } from '@modulacms/types'
