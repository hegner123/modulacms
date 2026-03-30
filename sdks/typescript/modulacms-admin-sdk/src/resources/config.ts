/**
 * Config management resource providing get, update, and meta endpoints.
 *
 * The config system manages all runtime settings for a ModulaCMS instance.
 * Settings are organized into categories (e.g. `'server'`, `'database'`,
 * `'media'`, `'auth'`, `'cors'`, `'plugins'`, `'observability'`).
 *
 * @remarks
 * **Hot-reloadable vs restart-required:** Each config field has a `hot_reloadable`
 * flag in its metadata. Hot-reloadable fields take effect immediately after update
 * (e.g. CORS origins, rate limits). Fields that are not hot-reloadable require a
 * server restart to take effect (e.g. database driver, port numbers, SSL settings).
 *
 * When an update includes non-hot-reloadable fields, the response includes a
 * `restart_required` array listing which fields need a restart. The update is still
 * persisted immediately -- the restart is only needed for the new values to take effect
 * in the running process.
 *
 * **Sensitive fields:** Some config fields (e.g. S3 secret keys, OAuth client secrets)
 * are marked `sensitive` in metadata. These are redacted in GET responses (returned as
 * `"********"`) but can be written via update.
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
  SearchIndexResponse,
} from '../types/config.js'

/** Config management operations available on `client.config`. */
type ConfigResource = {
  /**
   * Retrieve the current configuration, optionally filtered by category.
   *
   * Sensitive fields (passwords, secret keys) are redacted in the response
   * and returned as `"********"`.
   *
   * @param category - Optional category name to filter results (e.g. `'server'`,
   *   `'database'`, `'media'`, `'auth'`, `'cors'`). When omitted, all categories
   *   are returned.
   * @param opts - Optional request configuration (headers, signal, etc.).
   * @returns The current config values as a flat key-value object.
   * @throws {ApiError} 401 if not authenticated, 403 if missing `config:read` permission.
   *
   * @example
   * ```ts
   * // Get all config
   * const all = await client.config.get()
   * console.log(all.config.port) // 8080
   *
   * // Get only media-related config
   * const media = await client.config.get('media')
   * ```
   */
  get: (category?: string, opts?: RequestOptions) => Promise<ConfigGetResponse>
  /**
   * Update one or more configuration fields.
   *
   * Only the fields included in the `updates` object are changed; all other
   * fields retain their current values (partial update / merge semantics).
   *
   * @param updates - A flat object of config keys to new values. Keys must match
   *   the `json_key` values from the metadata registry. Unknown keys are rejected.
   * @param opts - Optional request configuration.
   * @returns The updated config, plus `restart_required` listing any fields that
   *   need a server restart and `warnings` for any non-fatal issues.
   * @throws {ApiError} 400 if a key is unknown or a value fails validation.
   * @throws {ApiError} 401 if not authenticated, 403 if missing `config:update` permission.
   *
   * @remarks
   * Check the `restart_required` array in the response. If non-empty, the updated
   * values have been persisted but will not take effect until the server is restarted.
   * Hot-reloadable fields take effect immediately.
   *
   * @example
   * ```ts
   * const result = await client.config.update({
   *   cors_allowed_origins: 'https://example.com,https://app.example.com',
   *   rate_limit_rps: 100,
   * })
   *
   * if (result.restart_required?.length) {
   *   console.warn('Restart needed for:', result.restart_required)
   * }
   * ```
   */
  update: (updates: Record<string, unknown>, opts?: RequestOptions) => Promise<ConfigUpdateResponse>
  /**
   * Retrieve the field metadata registry describing all available config fields.
   *
   * Each entry includes the field's JSON key, display label, category, validation
   * constraints, and behavioral flags (`hot_reloadable`, `sensitive`, `required`).
   *
   * @param opts - Optional request configuration.
   * @returns The list of all config field metadata entries and available category names.
   * @throws {ApiError} 401 if not authenticated, 403 if missing `config:read` permission.
   *
   * @remarks
   * Use the `categories` array in the response to populate category filter dropdowns.
   * Use `hot_reloadable` to inform users whether a change takes effect immediately
   * or requires a restart.
   *
   * @example
   * ```ts
   * const meta = await client.config.meta()
   *
   * // Find all fields that need a restart
   * const restartFields = meta.fields.filter(f => !f.hot_reloadable)
   *
   * // List available categories
   * console.log(meta.categories) // ['server', 'database', 'media', ...]
   * ```
   */
  meta: (opts?: RequestOptions) => Promise<ConfigMetaResponse>
  /**
   * Retrieve the current search index status.
   *
   * @param opts - Optional request configuration.
   * @returns The search index status including document count and memory usage.
   * @throws {ApiError} 401 if not authenticated, 403 if missing `config:read` permission.
   */
  searchIndex: (opts?: RequestOptions) => Promise<SearchIndexResponse>
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

    searchIndex(opts?: RequestOptions): Promise<SearchIndexResponse> {
      return http.get<SearchIndexResponse>('/admin/config/search-index', undefined, opts)
    },
  }
}

export type { ConfigResource }
export { createConfigResource }
