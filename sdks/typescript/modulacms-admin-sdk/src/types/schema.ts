/**
 * Schema-level entity types: datatypes, fields, and their junction table.
 * Entity types are re-exported from @modulacms/types; param types are local.
 *
 * @module types/schema
 */

import type {
  AdminFieldTypeID,
  ContentID,
  DatatypeID,
  FieldID,
  FieldType,
  FieldTypeID,
  UserID,
} from './common.js'

// Re-export shared entity types
export type { Datatype, Field, FieldTypeInfo, AdminFieldTypeInfo, DatatypeField } from '@modulacms/types'

// ---------------------------------------------------------------------------
// View types (composed responses from /datatype/full)
// ---------------------------------------------------------------------------

/** Author summary embedded in a full view response. */
export type AuthorView = {
  /** Unique identifier for this user. */
  user_id: UserID
  /** Login username. */
  username: string
  /** Display name. */
  name: string
  /** Email address. */
  email: string
  /** Role label. */
  role: string
}

/** A field definition with sort order, used in the datatype full view. */
export type DatatypeFieldView = {
  /** Unique identifier for the field. */
  field_id: FieldID
  /** Human-readable label. */
  label: string
  /** The data type of this field. */
  type: FieldType
  /** Additional metadata (JSON-encoded). */
  data: string
  /** Validation rules (JSON-encoded). */
  validation: string
  /** UI configuration (JSON-encoded). */
  ui_config: string
  /** Display ordering position. */
  sort_order: number
}

/** A fully composed datatype response from `GET /datatype/full`. */
export type DatatypeFullView = {
  /** Unique identifier for this datatype. */
  datatype_id: DatatypeID
  /** Human-readable label. */
  label: string
  /** Datatype category. */
  type: string
  /** Parent datatype ID, or `null`. */
  parent_id: DatatypeID | null
  /** Author who created this datatype. */
  author?: AuthorView
  /** All field definitions belonging to this datatype. */
  fields: DatatypeFieldView[]
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/** Parameters for creating a new datatype-field junction record via `POST /datatypefields`. */
export type CreateDatatypeFieldParams = {
  /** The datatype in the relationship. */
  datatype_id: DatatypeID
  /** The field in the relationship. */
  field_id: FieldID
  /** Display ordering position within the datatype. */
  sort_order: number
}

/** Parameters for creating a new datatype via `POST /datatype`. */
export type CreateDatatypeParams = {
  /** Unique identifier to assign (client-generated). */
  datatype_id: DatatypeID
  /** Parent content ID, or `null`. */
  parent_id: ContentID | null
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

/** Parameters for creating a new field definition via `POST /fields`. */
export type CreateFieldParams = {
  /** Unique identifier to assign (client-generated). */
  field_id: FieldID
  /** Parent datatype ID, or `null`. */
  parent_id: DatatypeID | null
  /** Human-readable label. */
  label: string
  /** Additional metadata (JSON-encoded). */
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

/** Parameters for updating a datatype-field junction record via `PUT /datatypefields/`. */
export type UpdateDatatypeFieldParams = {
  /** Unique identifier for this junction record. */
  id: string
  /** The datatype in the relationship. */
  datatype_id: DatatypeID
  /** The field in the relationship. */
  field_id: FieldID
  /** Display ordering position within the datatype. */
  sort_order: number
}

/** Parameters for updating a datatype via `PUT /datatype/`. */
export type UpdateDatatypeParams = {
  /** ID of the datatype to update. */
  datatype_id: DatatypeID
  /** Updated parent content ID, or `null`. */
  parent_id: ContentID | null
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

/** Parameters for updating a field definition via `PUT /fields/`. */
export type UpdateFieldParams = {
  /** ID of the field to update. */
  field_id: FieldID
  /** Updated parent datatype ID, or `null`. */
  parent_id: DatatypeID | null
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
// Field type lookup params
// ---------------------------------------------------------------------------

/** Parameters for creating a new field type via `POST /fieldtypes`. */
export type CreateFieldTypeParams = {
  /** Machine-readable type key. */
  type: string
  /** Human-readable label. */
  label: string
}

/** Parameters for updating a field type via `PUT /fieldtypes/`. */
export type UpdateFieldTypeParams = {
  /** ID of the field type to update. */
  field_type_id: FieldTypeID
  /** Updated type key. */
  type: string
  /** Updated label. */
  label: string
}

/** Parameters for creating a new admin field type via `POST /adminfieldtypes`. */
export type CreateAdminFieldTypeParams = {
  /** Machine-readable type key. */
  type: string
  /** Human-readable label. */
  label: string
}

/** Parameters for updating an admin field type via `PUT /adminfieldtypes/`. */
export type UpdateAdminFieldTypeParams = {
  /** ID of the admin field type to update. */
  admin_field_type_id: AdminFieldTypeID
  /** Updated type key. */
  type: string
  /** Updated label. */
  label: string
}
