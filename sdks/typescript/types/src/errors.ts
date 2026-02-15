/**
 * API error type and type guard for ModulaCMS SDK error handling.
 *
 * @module errors
 */

/**
 * Structured error returned by the SDK when the API responds with a non-2xx status.
 * Distinguished from other errors by the `_tag` discriminant field.
 *
 * Network-level failures (DNS, connection refused) throw native `TypeError`
 * instead -- use {@link isApiError} to differentiate.
 *
 * @example
 * ```ts
 * try {
 *   await client.users.get(id)
 * } catch (err) {
 *   if (isApiError(err)) {
 *     console.error(`API ${err.status}: ${err.message}`, err.body)
 *   }
 * }
 * ```
 */
export type ApiError = {
  /** Discriminant tag. Always `'ApiError'`. */
  readonly _tag: 'ApiError'
  /** HTTP status code from the response. */
  status: number
  /** HTTP status text (e.g. `'Not Found'`). */
  message: string
  /** Parsed JSON response body, if the server returned `application/json`. */
  body?: unknown
}

/**
 * Type guard that narrows an unknown caught value to {@link ApiError}.
 * Returns `false` for network errors, timeouts, and non-SDK exceptions.
 *
 * @param err - The caught error value to check.
 * @returns `true` if `err` is an {@link ApiError} with `_tag === 'ApiError'`.
 *
 * @example
 * ```ts
 * catch (err) {
 *   if (isApiError(err)) {
 *     // err is ApiError here
 *   }
 * }
 * ```
 */
export function isApiError(err: unknown): err is ApiError {
  return (
    typeof err === 'object' &&
    err !== null &&
    '_tag' in err &&
    (err as ApiError)._tag === 'ApiError'
  )
}
