/**
 * Content delivery resource for fetching rendered content trees by slug
 * via `GET /content/{slug}`.
 *
 * @module resources/content-delivery
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions, ContentFormat, Slug } from '../types/common.js'
import type { ContentTree } from '../types/tree.js'

/**
 * Content delivery operations available on `client.contentDelivery`.
 */
type ContentDeliveryResource = {
  /**
   * Fetch a rendered content tree by route slug.
   *
   * The API resolves the slug to a route, builds the full content tree,
   * and returns it in the requested output format.
   *
   * When `format` is omitted, returns the server default format.
   * For specific formats (`'contentful'`, `'sanity'`, etc.), returns a
   * platform-specific JSON structure as `Record<string, unknown>`.
   *
   * @param slug - Route slug (e.g. `"about"`, `"blog"`). Leading slashes are stripped.
   * @param format - Optional output format.
   * @param opts - Optional request options.
   * @returns The content tree in the requested format.
   */
  getPage: (slug: Slug, format?: ContentFormat, opts?: RequestOptions) => Promise<ContentTree | Record<string, unknown>>
}

/**
 * Create the content delivery resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link ContentDeliveryResource} with a `getPage` method.
 * @internal
 */
function createContentDeliveryResource(http: HttpClient): ContentDeliveryResource {
  return {
    getPage(slug: Slug, format?: ContentFormat, opts?: RequestOptions): Promise<ContentTree | Record<string, unknown>> {
      const trimmed = slug.startsWith('/') ? slug.slice(1) : slug
      const params: Record<string, string> | undefined = format ? { format } : undefined
      return http.get<ContentTree | Record<string, unknown>>(`/content/${trimmed}`, params, opts)
    },
  }
}

export type { ContentDeliveryResource }
export { createContentDeliveryResource }
