/**
 * Metrics resource for retrieving server metrics snapshots.
 *
 * @module resources/metrics
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

/** Server metrics snapshot from `GET /admin/metrics`. */
export type MetricsSnapshot = Record<string, unknown>

// ---------------------------------------------------------------------------
// Resource type
// ---------------------------------------------------------------------------

/** Metrics operations available on `client.metrics`. */
export type MetricsResource = {
  /** Get the current server metrics snapshot. */
  get: (opts?: RequestOptions) => Promise<MetricsSnapshot>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create the metrics resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link MetricsResource} with a `get` method.
 * @internal
 */
function createMetricsResource(http: HttpClient): MetricsResource {
  return {
    get(opts?: RequestOptions): Promise<MetricsSnapshot> {
      return http.get<MetricsSnapshot>('/admin/metrics', undefined, opts)
    },
  }
}

export { createMetricsResource }
