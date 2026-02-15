/**
 * Public route entity type.
 *
 * @module entities/routing
 */

import type { RouteID, Slug, UserID } from '../ids.js'

/**
 * A public-facing route that maps a URL slug to a content tree.
 */
export type Route = {
  /** Unique identifier for this route. */
  route_id: RouteID
  /** URL-safe slug for this route. */
  slug: Slug
  /** Human-readable title. */
  title: string
  /** Numeric status flag (0 = inactive, 1 = active). */
  status: number
  /** ID of the user who created this route, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}
