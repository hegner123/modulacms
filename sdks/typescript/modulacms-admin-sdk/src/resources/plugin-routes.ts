/**
 * Plugin route approval resource providing list, approve, and revoke
 * endpoints for managing plugin-registered HTTP routes.
 *
 * @module resources/plugin-routes
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'
import type { PluginRoute, RouteApprovalItem } from '../types/plugins.js'

/** Plugin route approval operations available on `client.pluginRoutes`. */
type PluginRoutesResource = {
  /** List all plugin-registered routes with approval status. */
  list: (opts?: RequestOptions) => Promise<PluginRoute[]>
  /** Approve one or more plugin routes. */
  approve: (routes: RouteApprovalItem[], opts?: RequestOptions) => Promise<void>
  /** Revoke approval for one or more plugin routes. */
  revoke: (routes: RouteApprovalItem[], opts?: RequestOptions) => Promise<void>
}

/**
 * Create the plugin routes resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link PluginRoutesResource}.
 * @internal
 */
function createPluginRoutesResource(http: HttpClient): PluginRoutesResource {
  return {
    async list(opts?: RequestOptions): Promise<PluginRoute[]> {
      const envelope = await http.get<{ routes: PluginRoute[] }>('/admin/plugins/routes', undefined, opts)
      return envelope.routes
    },

    async approve(routes: RouteApprovalItem[], opts?: RequestOptions): Promise<void> {
      await http.post<{ ok: true }>('/admin/plugins/routes/approve', { routes } as unknown as Record<string, unknown>, opts)
    },

    async revoke(routes: RouteApprovalItem[], opts?: RequestOptions): Promise<void> {
      await http.post<{ ok: true }>('/admin/plugins/routes/revoke', { routes } as unknown as Record<string, unknown>, opts)
    },
  }
}

export type { PluginRoutesResource }
export { createPluginRoutesResource }
