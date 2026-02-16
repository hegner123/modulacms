/**
 * Plugin management resource providing list, info, reload, enable, disable,
 * and cleanup endpoints for the plugin system.
 *
 * @module resources/plugins
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'
import type {
  PluginListItem,
  PluginInfo,
  PluginActionResponse,
  PluginStateResponse,
  CleanupDryRunResponse,
  CleanupDropParams,
  CleanupDropResponse,
} from '../types/plugins.js'

/** Plugin management operations available on `client.plugins`. */
type PluginsResource = {
  /** List all installed plugins. */
  list: (opts?: RequestOptions) => Promise<PluginListItem[]>
  /** Get detailed info for a specific plugin. */
  get: (name: string, opts?: RequestOptions) => Promise<PluginInfo>
  /** Reload a plugin from disk. */
  reload: (name: string, opts?: RequestOptions) => Promise<PluginActionResponse>
  /** Enable a disabled plugin. */
  enable: (name: string, opts?: RequestOptions) => Promise<PluginStateResponse>
  /** Disable an active plugin. */
  disable: (name: string, opts?: RequestOptions) => Promise<PluginStateResponse>
  /** Dry-run cleanup to list orphaned tables. */
  cleanupDryRun: (opts?: RequestOptions) => Promise<CleanupDryRunResponse>
  /** Drop orphaned tables (destructive). */
  cleanupDrop: (params: CleanupDropParams, opts?: RequestOptions) => Promise<CleanupDropResponse>
}

/**
 * Create the plugins resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link PluginsResource}.
 * @internal
 */
function createPluginsResource(http: HttpClient): PluginsResource {
  return {
    async list(opts?: RequestOptions): Promise<PluginListItem[]> {
      const envelope = await http.get<{ plugins: PluginListItem[] }>('/admin/plugins', undefined, opts)
      return envelope.plugins
    },

    get(name: string, opts?: RequestOptions): Promise<PluginInfo> {
      return http.get<PluginInfo>(`/admin/plugins/${encodeURIComponent(name)}`, undefined, opts)
    },

    reload(name: string, opts?: RequestOptions): Promise<PluginActionResponse> {
      return http.post<PluginActionResponse>(`/admin/plugins/${encodeURIComponent(name)}/reload`, undefined, opts)
    },

    enable(name: string, opts?: RequestOptions): Promise<PluginStateResponse> {
      return http.post<PluginStateResponse>(`/admin/plugins/${encodeURIComponent(name)}/enable`, undefined, opts)
    },

    disable(name: string, opts?: RequestOptions): Promise<PluginStateResponse> {
      return http.post<PluginStateResponse>(`/admin/plugins/${encodeURIComponent(name)}/disable`, undefined, opts)
    },

    cleanupDryRun(opts?: RequestOptions): Promise<CleanupDryRunResponse> {
      return http.get<CleanupDryRunResponse>('/admin/plugins/cleanup', undefined, opts)
    },

    cleanupDrop(params: CleanupDropParams, opts?: RequestOptions): Promise<CleanupDropResponse> {
      return http.post<CleanupDropResponse>('/admin/plugins/cleanup', params as unknown as Record<string, unknown>, opts)
    },
  }
}

export type { PluginsResource }
export { createPluginsResource }
