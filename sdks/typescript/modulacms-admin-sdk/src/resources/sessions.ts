/**
 * Sessions resource providing update and remove operations only.
 * Sessions are created implicitly via login; no list, get, or create endpoints exist.
 *
 * @module resources/sessions
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions, SessionID } from '../types/common.js'
import type { Session, UpdateSessionParams } from '../types/users.js'

/**
 * Session management operations available on `client.sessions`.
 */
type SessionsResource = {
  /**
   * Update an existing session's metadata.
   * @param params - Session update parameters (must include `session_id`).
   * @param opts - Optional request options.
   * @returns The updated session entity.
   */
  update: (params: UpdateSessionParams, opts?: RequestOptions) => Promise<Session>

  /**
   * Remove (invalidate) a session by its ID.
   * @param id - The session ID to remove.
   * @param opts - Optional request options.
   */
  remove: (id: SessionID, opts?: RequestOptions) => Promise<void>
}

/**
 * Create the sessions resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link SessionsResource} with update and remove methods.
 * @internal
 */
function createSessionsResource(http: HttpClient): SessionsResource {
  return {
    update(params: UpdateSessionParams, opts?: RequestOptions): Promise<Session> {
      return http.put<Session>('/sessions/', params as Record<string, unknown>, opts)
    },

    remove(id: SessionID, opts?: RequestOptions): Promise<void> {
      return http.del('/sessions/', { q: String(id) }, opts)
    },
  }
}

export type { SessionsResource }
export { createSessionsResource }
