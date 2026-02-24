/**
 * Deploy sync resource for cross-instance content synchronization.
 * Provides health check, export, and import operations.
 *
 * @module resources/deploy
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'
import type { DeployHealthResponse, DeploySyncPayload, DeploySyncResult } from '../types/deploy.js'

/**
 * Deploy sync operations available on `client.deploy`.
 */
type DeployResource = {
  /**
   * Check deploy endpoint health and version info.
   * @param opts - Optional request options.
   * @returns Health response with status, version, and node ID.
   */
  health: (opts?: RequestOptions) => Promise<DeployHealthResponse>

  /**
   * Export CMS data as a sync payload. Optionally limit to specific tables.
   * @param tables - Table names to export. Omit or pass empty array for default table set.
   * @param opts - Optional request options.
   * @returns The full sync payload.
   */
  export: (tables?: string[], opts?: RequestOptions) => Promise<DeploySyncPayload>

  /**
   * Import a sync payload into this instance.
   * @param payload - The sync payload to import.
   * @param dryRun - Set to true to validate without making database changes.
   * @param opts - Optional request options.
   * @returns Import result with affected tables, row counts, and any errors.
   */
  importPayload: (payload: DeploySyncPayload, dryRun?: boolean, opts?: RequestOptions) => Promise<DeploySyncResult>
}

/**
 * Create the deploy resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link DeployResource} with health, export, and import methods.
 * @internal
 */
function createDeployResource(http: HttpClient): DeployResource {
  return {
    health(opts?: RequestOptions): Promise<DeployHealthResponse> {
      return http.get<DeployHealthResponse>('/deploy/health', undefined, opts)
    },

    export(tables?: string[], opts?: RequestOptions): Promise<DeploySyncPayload> {
      const body = (tables !== undefined && tables.length > 0)
        ? { tables } as unknown as Record<string, unknown>
        : undefined
      return http.post<DeploySyncPayload>('/deploy/export', body, opts)
    },

    importPayload(payload: DeploySyncPayload, dryRun?: boolean, opts?: RequestOptions): Promise<DeploySyncResult> {
      const path = dryRun ? '/deploy/import?dry_run=true' : '/deploy/import'
      return http.post<DeploySyncResult>(path, payload as Record<string, unknown>, opts)
    },
  }
}

export type { DeployResource }
export { createDeployResource }
