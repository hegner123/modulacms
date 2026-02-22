/**
 * Branded ID types for compile-time safety across all ModulaCMS entities.
 *
 * @module ids
 */

import type { Brand } from './brand.js'

/** Unique identifier for a user account. */
export type UserID = Brand<string, 'UserID'>

/** Unique identifier for an admin content data node. */
export type AdminContentID = Brand<string, 'AdminContentID'>

/** Unique identifier for an admin content field value. */
export type AdminContentFieldID = Brand<string, 'AdminContentFieldID'>

/** Unique identifier for an admin content relation. */
export type AdminContentRelationID = Brand<string, 'AdminContentRelationID'>

/** Unique identifier for an admin datatype definition. */
export type AdminDatatypeID = Brand<string, 'AdminDatatypeID'>

/** Unique identifier for an admin field definition. */
export type AdminFieldID = Brand<string, 'AdminFieldID'>

/** Unique identifier for an admin route. */
export type AdminRouteID = Brand<string, 'AdminRouteID'>

/** Unique identifier for a public content data node. */
export type ContentID = Brand<string, 'ContentID'>

/** Unique identifier for a public content field value. */
export type ContentFieldID = Brand<string, 'ContentFieldID'>

/** Unique identifier for a public content relation. */
export type ContentRelationID = Brand<string, 'ContentRelationID'>

/** Unique identifier for a datatype (schema) definition. */
export type DatatypeID = Brand<string, 'DatatypeID'>

/** Unique identifier for a field (schema) definition. */
export type FieldID = Brand<string, 'FieldID'>

/** Unique identifier for a media asset. */
export type MediaID = Brand<string, 'MediaID'>

/** Unique identifier for a user role. */
export type RoleID = Brand<string, 'RoleID'>

/** Unique identifier for a permission. */
export type PermissionID = Brand<string, 'PermissionID'>

/** Unique identifier for a role-permission junction row. */
export type RolePermissionID = Brand<string, 'RolePermissionID'>

/** Unique identifier for a field type lookup entry. */
export type FieldTypeID = Brand<string, 'FieldTypeID'>

/** Unique identifier for an admin field type lookup entry. */
export type AdminFieldTypeID = Brand<string, 'AdminFieldTypeID'>

/** Unique identifier for a public route. */
export type RouteID = Brand<string, 'RouteID'>

/** Unique identifier for an active session. */
export type SessionID = Brand<string, 'SessionID'>

/** Unique identifier for a user OAuth connection. */
export type UserOauthID = Brand<string, 'UserOauthID'>

/** URL-safe slug string used to identify routes. */
export type Slug = Brand<string, 'Slug'>

/** Branded email address string. */
export type Email = Brand<string, 'Email'>

/** Branded URL string. */
export type URL = Brand<string, 'URL'>
