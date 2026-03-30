/**
 * Environment resource for retrieving the current CMS environment info.
 *
 * @module resources/environment
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

/** Response from the environment endpoint (`GET /environment`). */
export type EnvironmentResponse = {
  /** Environment name (e.g. `"production"`, `"development"`). */
  environment: string
  /** Deployment stage identifier. */
  stage: string
}

// ---------------------------------------------------------------------------
// Resource type
// ---------------------------------------------------------------------------

/** Environment operations available on `client.environment`. */
export type EnvironmentResource = {
  /** Get the current environment and stage information. */
  get: (opts?: RequestOptions) => Promise<EnvironmentResponse>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create the environment resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns An {@link EnvironmentResource} with a `get` method.
 * @internal
 */
function createEnvironmentResource(http: HttpClient): EnvironmentResource {
  return {
    get(opts?: RequestOptions): Promise<EnvironmentResponse> {
      return http.get<EnvironmentResponse>('/environment', undefined, opts)
    },
  }
}

export { createEnvironmentResource }
