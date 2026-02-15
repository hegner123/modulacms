/**
 * Schema-level entity types: datatypes, fields, and their junction table.
 * These define the content model structure used by both admin and public content.
 *
 * @module entities/schema
 */

import type {
  ContentID,
  DatatypeID,
  FieldID,
  UserID,
} from '../ids.js'
import type { FieldType } from '../enums.js'

/**
 * A datatype (content type) definition that describes the schema of a content category.
 */
export type Datatype = {
  /** Unique identifier for this datatype. */
  datatype_id: DatatypeID
  /** Parent content ID for hierarchical datatypes, or `null`. */
  parent_id: ContentID | null
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
 * A field definition belonging to a datatype.
 * Defines the name, type, validation, and UI configuration of a content field.
 */
export type Field = {
  /** Unique identifier for this field definition. */
  field_id: FieldID
  /** Parent datatype ID this field belongs to, or `null`. */
  parent_id: DatatypeID | null
  /** Human-readable field label. */
  label: string
  /** Additional field metadata (JSON-encoded). */
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
 * Junction record linking a datatype to a field with ordering.
 */
export type DatatypeField = {
  /** Unique identifier for this junction record. */
  id: string
  /** The datatype in the relationship. */
  datatype_id: DatatypeID
  /** The field in the relationship. */
  field_id: FieldID
  /** Display ordering position within the datatype. */
  sort_order: number
}
