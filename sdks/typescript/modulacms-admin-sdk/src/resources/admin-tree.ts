/**
 * Admin tree resource for retrieving fully resolved content hierarchies
 * via `GET /admin/tree/{slug}`.
 *
 * @module resources/admin-tree
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions, Slug } from '../types/common.js'
import type { AdminTreeResponse, TreeFormat } from '../types/tree.js'

/**
 * Admin tree operations available on `client.adminTree`.
 */
type AdminTreeResource = {
  /**
   * Retrieve the full content tree for an admin route.
   *
   * When `format` is `'raw'` or omitted, returns an {@link AdminTreeResponse}.
   * For other formats (`'contentful'`, `'sanity'`, etc.), returns a
   * platform-specific JSON structure as `Record<string, unknown>`.
   *
   * @param slug - The admin route slug to retrieve the tree for.
   * @param format - Optional output format. Defaults to `'raw'`.
   * @param opts - Optional request options.
   * @returns The content tree in the requested format.
   */
  get: (slug: Slug, format?: TreeFormat, opts?: RequestOptions) => Promise<AdminTreeResponse | Record<string, unknown>>
}

/**
 * Create the admin tree resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns An {@link AdminTreeResource} with a `get` method.
 * @internal
 */
function createAdminTreeResource(http: HttpClient): AdminTreeResource {
  return {
    get(slug: Slug, format?: TreeFormat, opts?: RequestOptions): Promise<AdminTreeResponse | Record<string, unknown>> {
      const params: Record<string, string> | undefined = format ? { format } : undefined
      return http.get<AdminTreeResponse | Record<string, unknown>>(`/admin/tree/${slug}`, params, opts)
    },
  }
}

export type { AdminTreeResource }
export { createAdminTreeResource }
