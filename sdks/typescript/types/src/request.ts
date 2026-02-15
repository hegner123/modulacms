/**
 * Per-request configuration types.
 *
 * @module request
 */

/**
 * Optional per-request configuration passed to every SDK method.
 *
 * @example
 * ```ts
 * const controller = new AbortController()
 * const users = await client.users.list({ signal: controller.signal })
 * ```
 */
export type RequestOptions = {
  /** An {@link AbortSignal} to cancel the request. Merged with the default timeout signal. */
  signal?: AbortSignal
}
