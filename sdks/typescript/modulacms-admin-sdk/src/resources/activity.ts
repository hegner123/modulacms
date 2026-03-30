/**
 * Activity resource for retrieving recent change events.
 *
 * @module resources/activity
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

/** A recent activity/change event with actor information. */
export type ActivityItem = {
  /** Unique change event ID. */
  id: string
  /** The operation type (e.g. `"create"`, `"update"`, `"delete"`). */
  operation: string
  /** The entity table name. */
  table_name: string
  /** The affected entity ID. */
  entity_id: string
  /** The user who performed the action. */
  actor_id: string
  /** Actor display name, if available. */
  actor_name?: string
  /** ISO 8601 timestamp of the event. */
  timestamp: string
  /** Additional metadata about the change. */
  metadata?: Record<string, unknown>
}

// ---------------------------------------------------------------------------
// Resource type
// ---------------------------------------------------------------------------

/** Activity operations available on `client.activity`. */
export type ActivityResource = {
  /**
   * List recent activity events.
   * @param limit - Maximum number of events to return. Server default applies if omitted.
   * @param opts - Optional request options.
   */
  listRecent: (limit?: number, opts?: RequestOptions) => Promise<ActivityItem[]>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create the activity resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns An {@link ActivityResource} with a `listRecent` method.
 * @internal
 */
function createActivityResource(http: HttpClient): ActivityResource {
  return {
    listRecent(limit?: number, opts?: RequestOptions): Promise<ActivityItem[]> {
      const params: Record<string, string> | undefined = limit !== undefined
        ? { limit: String(limit) }
        : undefined
      return http.get<ActivityItem[]>('/activity/recent', params, opts)
    },
  }
}

export { createActivityResource }
