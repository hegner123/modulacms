/**
 * Admin search resource for rebuilding the search index.
 *
 * @module resources/search
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

/** Response from the search index rebuild endpoint (`POST /admin/search/rebuild`). */
export type SearchRebuildResponse = {
  /** Status of the rebuild operation. */
  status: string
  /** Number of documents indexed. */
  documents: number
  /** Number of unique terms in the index. */
  terms: number
  /** Memory consumed by the index in bytes. */
  mem_bytes: number
}

// ---------------------------------------------------------------------------
// Resource type
// ---------------------------------------------------------------------------

/** Admin search operations available on `client.search`. */
export type SearchResource = {
  /** Rebuild the full-text search index. */
  rebuild: (opts?: RequestOptions) => Promise<SearchRebuildResponse>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create the admin search resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link SearchResource} with a `rebuild` method.
 * @internal
 */
function createSearchResource(http: HttpClient): SearchResource {
  return {
    rebuild(opts?: RequestOptions): Promise<SearchRebuildResponse> {
      return http.post<SearchRebuildResponse>('/admin/search/rebuild', {} as Record<string, unknown>, opts)
    },
  }
}

export { createSearchResource }
