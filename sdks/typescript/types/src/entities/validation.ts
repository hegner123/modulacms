/**
 * Validation entity types for reusable validation configs referenced by fields.
 *
 * @module entities/validation
 */

import type { ValidationID, AdminValidationID, UserID } from '../ids.js'

/**
 * A reusable validation configuration that can be referenced by field definitions.
 */
export type Validation = {
  /** Unique identifier for this validation config. */
  validation_id: ValidationID
  /** Human-readable name for this validation. */
  name: string
  /** Description of what this validation does. */
  description: string
  /** JSON-encoded validation rules configuration. */
  config: string
  /** ID of the user who created this validation, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modified timestamp. */
  date_modified: string
}

/**
 * An admin-side reusable validation configuration referenced by admin field definitions.
 */
export type AdminValidation = {
  /** Unique identifier for this admin validation config. */
  admin_validation_id: AdminValidationID
  /** Human-readable name for this validation. */
  name: string
  /** Description of what this validation does. */
  description: string
  /** JSON-encoded validation rules configuration. */
  config: string
  /** ID of the user who created this validation, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modified timestamp. */
  date_modified: string
}
