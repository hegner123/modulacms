/**
 * Public route entity type and its create/update parameter shapes.
 * Entity type is re-exported from @modulacms/types; param types are local.
 *
 * @module types/routing
 */

import type { RouteID, Slug, UserID } from './common.js'

// Re-export shared entity type
export type { Route } from '@modulacms/types'

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/** Parameters for creating a new public route via `POST /routes`. */
export type CreateRouteParams = {
  /** Unique identifier to assign (client-generated). */
  route_id: RouteID
  /** URL-safe slug. */
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

// ---------------------------------------------------------------------------
// Update params
// ---------------------------------------------------------------------------

/**
 * Parameters for updating a public route via `PUT /routes/`.
 * The `slug_2` field carries the current slug for the WHERE clause,
 * allowing the `slug` field to be changed in the same operation.
 */
export type UpdateRouteParams = {
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
