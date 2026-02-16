/**
 * Plugin hook approval resource providing list, approve, and revoke
 * endpoints for managing plugin-registered lifecycle hooks.
 *
 * @module resources/plugin-hooks
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'
import type { PluginHook, HookApprovalItem } from '../types/plugins.js'

/** Plugin hook approval operations available on `client.pluginHooks`. */
type PluginHooksResource = {
  /** List all plugin-registered hooks with approval status. */
  list: (opts?: RequestOptions) => Promise<PluginHook[]>
  /** Approve one or more plugin hooks. */
  approve: (hooks: HookApprovalItem[], opts?: RequestOptions) => Promise<void>
  /** Revoke approval for one or more plugin hooks. */
  revoke: (hooks: HookApprovalItem[], opts?: RequestOptions) => Promise<void>
}

/**
 * Create the plugin hooks resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link PluginHooksResource}.
 * @internal
 */
function createPluginHooksResource(http: HttpClient): PluginHooksResource {
  return {
    async list(opts?: RequestOptions): Promise<PluginHook[]> {
      const envelope = await http.get<{ hooks: PluginHook[] }>('/admin/plugins/hooks', undefined, opts)
      return envelope.hooks
    },

    async approve(hooks: HookApprovalItem[], opts?: RequestOptions): Promise<void> {
      await http.post<{ ok: true }>('/admin/plugins/hooks/approve', { hooks } as unknown as Record<string, unknown>, opts)
    },

    async revoke(hooks: HookApprovalItem[], opts?: RequestOptions): Promise<void> {
      await http.post<{ ok: true }>('/admin/plugins/hooks/revoke', { hooks } as unknown as Record<string, unknown>, opts)
    },
  }
}

export type { PluginHooksResource }
export { createPluginHooksResource }
