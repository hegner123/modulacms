/**
 * Config management resource providing get, update, and meta endpoints.
 *
 * @module resources/config
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'
import type {
  ConfigGetResponse,
  ConfigUpdateResponse,
  ConfigMetaResponse,
} from '../types/config.js'

/** Config management operations available on `client.config`. */
type ConfigResource = {
  /** Get the current config (redacted). Optional category filter. */
  get: (category?: string, opts?: RequestOptions) => Promise<ConfigGetResponse>
  /** Update config fields. Body is a JSON object with fields to change. */
  update: (updates: Record<string, unknown>, opts?: RequestOptions) => Promise<ConfigUpdateResponse>
  /** Get field metadata registry. */
  meta: (opts?: RequestOptions) => Promise<ConfigMetaResponse>
}

/**
 * Create the config resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link ConfigResource}.
 * @internal
 */
function createConfigResource(http: HttpClient): ConfigResource {
  return {
    get(category?: string, opts?: RequestOptions): Promise<ConfigGetResponse> {
      const params = category ? { category } : undefined
      return http.get<ConfigGetResponse>('/admin/config', params, opts)
    },

    update(updates: Record<string, unknown>, opts?: RequestOptions): Promise<ConfigUpdateResponse> {
      return http.patch<ConfigUpdateResponse>('/admin/config', updates, opts)
    },

    meta(opts?: RequestOptions): Promise<ConfigMetaResponse> {
      return http.get<ConfigMetaResponse>('/admin/config/meta', undefined, opts)
    },
  }
}

export type { ConfigResource }
export { createConfigResource }
