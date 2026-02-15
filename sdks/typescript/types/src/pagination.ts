/**
 * Pagination types for paginated API responses.
 *
 * @module pagination
 */

/** Parameters for paginated list requests. */
export type PaginationParams = {
  /** Maximum number of items to return. */
  limit: number
  /** Number of items to skip before starting to collect the result set. */
  offset: number
}

/**
 * Envelope returned by the API when pagination query parameters are present.
 *
 * @typeParam T - The entity type contained in the `data` array.
 */
export type PaginatedResponse<T> = {
  /** The page of entities. */
  data: T[]
  /** Total number of entities matching the query (across all pages). */
  total: number
  /** The limit that was applied. */
  limit: number
  /** The offset that was applied. */
  offset: number
}
