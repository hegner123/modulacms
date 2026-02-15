/**
 * Import resource for bulk content migration from external CMS platforms.
 * Provides per-format convenience methods and a generic `bulk` method.
 *
 * @module resources/import
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'
import type { ImportFormat, ImportResponse } from '../types/import.js'

/**
 * Bulk import operations available on `client.import`.
 */
type ImportResource = {
  /**
   * Import content from Contentful export format.
   * @param data - Contentful-formatted export data.
   * @param opts - Optional request options.
   * @returns Import result with creation counts and any errors.
   */
  contentful: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>

  /**
   * Import content from Sanity.io export format.
   * @param data - Sanity-formatted export data.
   * @param opts - Optional request options.
   * @returns Import result with creation counts and any errors.
   */
  sanity: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>

  /**
   * Import content from Strapi export format.
   * @param data - Strapi-formatted export data.
   * @param opts - Optional request options.
   * @returns Import result with creation counts and any errors.
   */
  strapi: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>

  /**
   * Import content from WordPress export format.
   * @param data - WordPress-formatted export data.
   * @param opts - Optional request options.
   * @returns Import result with creation counts and any errors.
   */
  wordpress: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>

  /**
   * Import content from ModulaCMS clean export format.
   * @param data - Clean-formatted export data.
   * @param opts - Optional request options.
   * @returns Import result with creation counts and any errors.
   */
  clean: (data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>

  /**
   * Import content using a dynamic format specifier.
   * Routes to `POST /import?format={format}`.
   * @param format - The source CMS format.
   * @param data - Export data in the specified format.
   * @param opts - Optional request options.
   * @returns Import result with creation counts and any errors.
   */
  bulk: (format: ImportFormat, data: Record<string, unknown>, opts?: RequestOptions) => Promise<ImportResponse>
}

/**
 * Create the import resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns An {@link ImportResource} with per-format and bulk import methods.
 * @internal
 */
function createImportResource(http: HttpClient): ImportResource {
  return {
    contentful(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse> {
      return http.post<ImportResponse>('/import/contentful', data, opts)
    },

    sanity(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse> {
      return http.post<ImportResponse>('/import/sanity', data, opts)
    },

    strapi(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse> {
      return http.post<ImportResponse>('/import/strapi', data, opts)
    },

    wordpress(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse> {
      return http.post<ImportResponse>('/import/wordpress', data, opts)
    },

    clean(data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse> {
      return http.post<ImportResponse>('/import/clean', data, opts)
    },

    bulk(format: ImportFormat, data: Record<string, unknown>, opts?: RequestOptions): Promise<ImportResponse> {
      return http.post<ImportResponse>(`/import?format=${format}`, data, opts)
    },
  }
}

export type { ImportResource }
export { createImportResource }
