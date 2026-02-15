/**
 * Table entity type and its create/update parameter shapes.
 *
 * @module types/tables
 */

import type { UserID } from './common.js'

// ---------------------------------------------------------------------------
// Entity type
// ---------------------------------------------------------------------------

/**
 * A named table entity in the CMS.
 */
export type Table = {
  /** Unique identifier for this table. */
  id: string
  /** Human-readable label. */
  label: string
  /** ID of the user who created this table, or `null`. */
  author_id: UserID | null
}

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/** Parameters for creating a new table via `POST /tables`. */
export type CreateTableParams = {
  /** Human-readable label for the new table. */
  label: string
}

// ---------------------------------------------------------------------------
// Update params
// ---------------------------------------------------------------------------

/** Parameters for updating a table via `PUT /tables/`. */
export type UpdateTableParams = {
  /** Updated label. */
  label: string
  /** ID of the table to update. */
  id: string
}
