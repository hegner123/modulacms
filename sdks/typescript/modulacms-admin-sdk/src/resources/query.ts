/**
 * Content query resource for fetching filtered, sorted, paginated content
 * lists by datatype name via `GET /query/{datatype}`.
 *
 * @module resources/query
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'
import type { QueryParams, QueryResult } from '@modulacms/types'

/**
 * Content query operations available on `client.query`.
 */
type QueryResource = {
  /**
   * Query content items by datatype name.
   *
   * @param datatype - Datatype name to query (e.g. `"blog_post"`).
   * @param params - Optional query parameters (filters, sort, limit, offset, locale, status).
   * @param opts - Optional request options.
   * @returns A paginated query result envelope.
   */
  query: (datatype: string, params?: QueryParams, opts?: RequestOptions) => Promise<QueryResult>
}

/**
 * Create the content query resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link QueryResource} with a `query` method.
 * @internal
 */
function createQueryResource(http: HttpClient): QueryResource {
  return {
    query(datatype: string, params?: QueryParams, opts?: RequestOptions): Promise<QueryResult> {
      const p: Record<string, string> = {}
      if (params?.sort) p.sort = params.sort
      if (params?.limit !== undefined) p.limit = String(params.limit)
      if (params?.offset !== undefined) p.offset = String(params.offset)
      if (params?.locale) p.locale = params.locale
      if (params?.status) p.status = params.status
      if (params?.filters) {
        for (const [key, value] of Object.entries(params.filters)) {
          p[key] = value
        }
      }
      const queryParams = Object.keys(p).length > 0 ? p : undefined
      return http.get<QueryResult>(`/query/${encodeURIComponent(datatype)}`, queryParams, opts)
    },
  }
}

export type { QueryResource }
export { createQueryResource }
