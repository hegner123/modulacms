/**
 * Query parameter types for the content query API.
 */

/** Parameters for a content query request. */
export interface QueryParams {
  /** Sort field. Prefix with `-` for descending (e.g. `-published_at`). */
  sort?: string
  /** Maximum items to return (default 20, max 100). */
  limit?: number
  /** Number of items to skip for pagination. */
  offset?: number
  /** Locale code for i18n content filtering. */
  locale?: string
  /** Content status filter (default: `published`). */
  status?: string
  /** Field filters as key-value pairs. Use `field[op]` keys for operators. */
  filters?: Record<string, string>
}

/** A single content item in a query result. */
export interface QueryItem {
  content_data_id: string
  datatype_id: string
  author_id: string
  status: string
  date_created: string
  date_modified: string
  published_at: string
  fields: Record<string, string>
}

/** Datatype metadata in a query result. */
export interface QueryDatatype {
  name: string
  label: string
}

/** Response envelope for a content query. */
export interface QueryResult {
  data: QueryItem[]
  total: number
  limit: number
  offset: number
  datatype: QueryDatatype
}
