/**
 * Globals resource for fetching published global content trees.
 *
 * @module resources/globals
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'

// ---------------------------------------------------------------------------
// Resource type
// ---------------------------------------------------------------------------

/** Globals operations available on `client.globals`. */
export type GlobalsResource = {
  /**
   * List all published global content trees.
   *
   * Global content items are root nodes typed as `_global` whose published
   * trees are available site-wide (e.g. navigation, footer, site settings).
   *
   * @param opts - Optional request options.
   * @returns The globals response. Shape depends on datatype schema configuration.
   */
  list: (opts?: RequestOptions) => Promise<unknown>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create the globals resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link GlobalsResource} with a `list` method.
 * @internal
 */
function createGlobalsResource(http: HttpClient): GlobalsResource {
  return {
    list(opts?: RequestOptions): Promise<unknown> {
      return http.get<unknown>('/globals', undefined, opts)
    },
  }
}

export { createGlobalsResource }
