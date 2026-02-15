/**
 * Schema-level entity types: datatypes, fields, and their junction table.
 * Entity types are re-exported from @modulacms/types; param types are local.
 *
 * @module types/schema
 */

import type {
  ContentID,
  DatatypeID,
  FieldID,
  FieldType,
  UserID,
} from './common.js'

// Re-export shared entity types
export type { Datatype, Field, DatatypeField } from '@modulacms/types'

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
