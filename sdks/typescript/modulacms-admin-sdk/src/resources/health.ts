/**
 * Health check resource for monitoring CMS instance status.
 *
 * @module resources/health
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

/** Response from the health check endpoint (`GET /health`). */
export type HealthResponse = {
  /** Overall system health status. */
  status: 'ok' | 'degraded'
  /** Environment name (e.g. `"production"`, `"staging"`). */
  environment: string
  /** Per-subsystem boolean checks. */
  checks: {
    database: boolean
    storage: boolean
    plugins: boolean
  }
  /** Optional detail strings for each subsystem (present when degraded). */
  details: {
    database?: string
    storage?: string
    plugins?: string
  }
}

// ---------------------------------------------------------------------------
// Resource type
// ---------------------------------------------------------------------------

/** Health check operations available on `client.health`. */
export type HealthResource = {
  /** Check the health status of the CMS instance. */
  check: (opts?: RequestOptions) => Promise<HealthResponse>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create the health resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link HealthResource} with a `check` method.
 * @internal
 */
function createHealthResource(http: HttpClient): HealthResource {
  return {
    check(opts?: RequestOptions): Promise<HealthResponse> {
      return http.get<HealthResponse>('/health', undefined, opts)
    },
  }
}

export { createHealthResource }
