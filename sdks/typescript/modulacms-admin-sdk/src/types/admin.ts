/**
 * Admin-side entity types and their create/update parameter shapes.
 * These mirror the admin content management tables in the ModulaCMS database.
 *
 * @module types/admin
 */

import type {
  AdminContentFieldID,
  AdminContentID,
  AdminContentRelationID,
  AdminDatatypeID,
  AdminFieldID,
  AdminRouteID,
  ContentStatus,
  FieldType,
  Slug,
  UserID,
} from './common.js'

// ---------------------------------------------------------------------------
// Entity types
// ---------------------------------------------------------------------------

/**
 * An admin-side route that serves as the top-level container for admin content trees.
 * Identified by both a unique ID and a URL slug.
 */
export type AdminRoute = {
  /** Unique identifier for this admin route. */
  admin_route_id: AdminRouteID
  /** URL-safe slug for this route. */
  slug: Slug
  /** Human-readable title of the route. */
  title: string
  /** Numeric status flag (0 = inactive, 1 = active). */
  status: number
  /** ID of the user who created this route, or `null` if system-generated. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * An admin content data node in the content tree.
 * Uses a linked-list tree structure (parent, first child, siblings).
 */
export type AdminContentData = {
  /** Unique identifier for this content node. */
  admin_content_data_id: AdminContentID
  /** Parent node ID, or `null` for root nodes. */
  parent_id: AdminContentID | null
  /** ID of this node's first child, or `null` if it has no children. */
  first_child_id: string | null
  /** ID of the next sibling node, or `null` if this is the last sibling. */
  next_sibling_id: string | null
  /** ID of the previous sibling node, or `null` if this is the first sibling. */
  prev_sibling_id: string | null
  /** The admin route this content belongs to. */
  admin_route_id: string
  /** The datatype that defines this content's schema, or `null` if untyped. */
  admin_datatype_id: AdminDatatypeID | null
  /** ID of the user who created this content, or `null`. */
  author_id: UserID | null
  /** Publication lifecycle status. */
  status: ContentStatus
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * A field value belonging to an admin content data node.
 * Links a content node to a specific field definition and stores the value.
 */
export type AdminContentField = {
  /** Unique identifier for this field value. */
  admin_content_field_id: AdminContentFieldID
  /** The admin route this field belongs to, or `null`. */
  admin_route_id: string | null
  /** The content data node this field value belongs to. */
  admin_content_data_id: string
  /** The field definition this value corresponds to, or `null`. */
  admin_field_id: AdminFieldID | null
  /** The stored value as a string (serialized based on field type). */
  admin_field_value: string
  /** ID of the user who set this value, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * An admin-side datatype definition that describes the schema of a content type.
 */
export type AdminDatatype = {
  /** Unique identifier for this datatype. */
  admin_datatype_id: AdminDatatypeID
  /** Parent content ID for hierarchical datatypes, or `null`. */
  parent_id: AdminContentID | null
  /** Human-readable label for this datatype. */
  label: string
  /** Datatype category (e.g. `'page'`, `'component'`). */
  type: string
  /** ID of the user who created this datatype, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * An admin-side field definition that belongs to a datatype.
 * Defines the name, type, validation, and UI configuration of a content field.
 */
export type AdminField = {
  /** Unique identifier for this field definition. */
  admin_field_id: AdminFieldID
  /** Parent datatype ID this field belongs to, or `null`. */
  parent_id: AdminDatatypeID | null
  /** Human-readable field label. */
  label: string
  /** Additional field data (JSON-encoded metadata). */
  data: string
  /** Validation rules (JSON-encoded). */
  validation: string
  /** UI widget configuration (JSON-encoded). */
  ui_config: string
  /** The data type of this field. */
  type: FieldType
  /** ID of the user who created this field, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * Junction record linking an admin datatype to an admin field.
 * Read-only; no dedicated API endpoint exists for this entity.
 */
export type AdminDatatypeField = {
  /** Unique identifier for this junction record. */
  id: string
  /** The datatype in the relationship. */
  admin_datatype_id: AdminDatatypeID
  /** The field in the relationship. */
  admin_field_id: AdminFieldID
}

/** Parameters for creating a new admin datatype-field junction record via `POST /admindatatypefields`. */
export type CreateAdminDatatypeFieldParams = {
  /** The admin datatype in the relationship. */
  admin_datatype_id: AdminDatatypeID
  /** The admin field in the relationship. */
  admin_field_id: AdminFieldID
}

/** Parameters for updating an admin datatype-field junction record via `PUT /admindatatypefields/`. */
export type UpdateAdminDatatypeFieldParams = {
  /** Unique identifier for this junction record. */
  id: string
  /** The admin datatype in the relationship. */
  admin_datatype_id: AdminDatatypeID
  /** The admin field in the relationship. */
  admin_field_id: AdminFieldID
}

/**
 * A directional relation between two admin content nodes via a specific field.
 */
export type AdminContentRelation = {
  /** Unique identifier for this relation. */
  admin_content_relation_id: AdminContentRelationID
  /** The content node that owns this relation. */
  source_content_id: AdminContentID
  /** The content node being referenced. */
  target_content_id: AdminContentID
  /** The field definition that established this relation. */
  admin_field_id: AdminFieldID
  /** Display ordering position. */
  sort_order: number
  /** ISO 8601 creation timestamp. */
  date_created: string
}

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/** Parameters for creating a new admin route via `POST /adminroutes`. */
export type CreateAdminRouteParams = {
  /** URL-safe slug for the new route. */
  slug: Slug
  /** Human-readable title. */
  title: string
  /** Numeric status flag. */
  status: number
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for creating a new admin content data node via `POST /admincontentdatas`. */
export type CreateAdminContentDataParams = {
  /** Parent node ID, or `null` for root nodes. */
  parent_id: AdminContentID | null
  /** First child node ID, or `null`. */
  first_child_id: string | null
  /** Next sibling node ID, or `null`. */
  next_sibling_id: string | null
  /** Previous sibling node ID, or `null`. */
  prev_sibling_id: string | null
  /** Admin route this content belongs to. */
  admin_route_id: string
  /** Datatype ID defining this content's schema, or `null`. */
  admin_datatype_id: AdminDatatypeID | null
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** Publication lifecycle status. */
  status: ContentStatus
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for creating a new admin content field value via `POST /admincontentfields`. */
export type CreateAdminContentFieldParams = {
  /** Admin route this field belongs to, or `null`. */
  admin_route_id: string | null
  /** Content data node this field value belongs to. */
  admin_content_data_id: string
  /** Field definition this value corresponds to, or `null`. */
  admin_field_id: AdminFieldID | null
  /** The field value as a serialized string. */
  admin_field_value: string
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for creating a new admin datatype definition via `POST /admindatatypes`. */
export type CreateAdminDatatypeParams = {
  /** Parent content ID, or `null`. */
  parent_id: AdminContentID | null
  /** Human-readable label. */
  label: string
  /** Datatype category. */
  type: string
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for creating a new admin field definition via `POST /adminfields`. */
export type CreateAdminFieldParams = {
  /** Parent datatype ID, or `null`. */
  parent_id: AdminDatatypeID | null
  /** Human-readable field label. */
  label: string
  /** Additional field metadata (JSON-encoded). */
  data: string
  /** Validation rules (JSON-encoded). */
  validation: string
  /** UI configuration (JSON-encoded). */
  ui_config: string
  /** The data type of this field. */
  type: FieldType
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

// ---------------------------------------------------------------------------
// Update params
// ---------------------------------------------------------------------------

/**
 * Parameters for updating an admin route via `PUT /adminroutes/`.
 * The `slug_2` field carries the current slug for the WHERE clause,
 * allowing the `slug` field to be changed in the same operation.
 */
export type UpdateAdminRouteParams = {
  /** New slug value (may differ from `slug_2` if renaming). */
  slug: Slug
  /** Updated title. */
  title: string
  /** Updated status flag. */
  status: number
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
  /** Current slug used to locate the record (WHERE clause). */
  slug_2: Slug
}

/** Parameters for updating an admin content data node via `PUT /admincontentdatas/`. */
export type UpdateAdminContentDataParams = {
  /** ID of the content node to update. */
  admin_content_data_id: AdminContentID
  /** Updated parent node ID, or `null`. */
  parent_id: AdminContentID | null
  /** Updated first child ID, or `null`. */
  first_child_id: string | null
  /** Updated next sibling ID, or `null`. */
  next_sibling_id: string | null
  /** Updated previous sibling ID, or `null`. */
  prev_sibling_id: string | null
  /** Admin route this content belongs to. */
  admin_route_id: string
  /** Updated datatype ID, or `null`. */
  admin_datatype_id: AdminDatatypeID | null
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** Updated publication status. */
  status: ContentStatus
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for updating an admin content field value via `PUT /admincontentfields/`. */
export type UpdateAdminContentFieldParams = {
  /** ID of the field value to update. */
  admin_content_field_id: AdminContentFieldID
  /** Admin route, or `null`. */
  admin_route_id: string | null
  /** Content data node this field belongs to. */
  admin_content_data_id: string
  /** Field definition, or `null`. */
  admin_field_id: AdminFieldID | null
  /** Updated field value. */
  admin_field_value: string
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for updating an admin datatype definition via `PUT /admindatatypes/`. */
export type UpdateAdminDatatypeParams = {
  /** ID of the datatype to update. */
  admin_datatype_id: AdminDatatypeID
  /** Updated parent content ID, or `null`. */
  parent_id: AdminContentID | null
  /** Updated label. */
  label: string
  /** Updated category type. */
  type: string
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for updating an admin field definition via `PUT /adminfields/`. */
export type UpdateAdminFieldParams = {
  /** ID of the field to update. */
  admin_field_id: AdminFieldID
  /** Updated parent datatype ID, or `null`. */
  parent_id: AdminDatatypeID | null
  /** Updated label. */
  label: string
  /** Updated metadata (JSON-encoded). */
  data: string
  /** Updated validation rules (JSON-encoded). */
  validation: string
  /** Updated UI configuration (JSON-encoded). */
  ui_config: string
  /** Updated field type. */
  type: FieldType
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

// ---------------------------------------------------------------------------
// Admin content data move/reorder
// ---------------------------------------------------------------------------

/** Parameters for moving an admin content data node to a new parent via `POST /admincontentdatas/move`. */
export type MoveAdminContentDataParams = {
  /** ID of the node to move. */
  node_id: AdminContentID
  /** New parent node ID, or `null` for root level. */
  new_parent_id: AdminContentID | null
  /** 0-indexed position within the new parent's children. */
  position: number
}

/** Response from the move admin content data endpoint. */
export type MoveAdminContentDataResponse = {
  /** ID of the moved node. */
  node_id: AdminContentID
  /** Previous parent node ID, or `null`. */
  old_parent_id: AdminContentID | null
  /** New parent node ID, or `null`. */
  new_parent_id: AdminContentID | null
  /** Position within the new parent's children. */
  position: number
}

/** Parameters for reordering admin content data siblings via `POST /admincontentdatas/reorder`. */
export type ReorderAdminContentDataParams = {
  /** Parent node ID, or `null` for root-level siblings. */
  parent_id: AdminContentID | null
  /** Ordered list of sibling admin content data IDs in the desired sequence. */
  ordered_ids: AdminContentID[]
}

/** Response from the reorder admin content data endpoint. */
export type ReorderAdminContentDataResponse = {
  /** Number of nodes updated. */
  updated: number
  /** Parent node ID, or `null`. */
  parent_id: AdminContentID | null
}
