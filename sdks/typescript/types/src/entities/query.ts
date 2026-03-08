/**
 * Query parameter types for the ModulaCMS content query API.
 *
 * The query system provides filtered, sorted, paginated access to content
 * of a specific datatype. Queries are scoped to a single datatype (identified
 * by slug in the URL path) and return content items with their field values
 * flattened into a key-value map.
 *
 * @remarks
 * All query parameters are optional. Omitting all parameters returns the
 * first 20 published content items in default sort order.
 *
 * @module entities/query
 */

/**
 * Parameters for a content query request.
 *
 * Passed as query string parameters to `GET /api/v1/query/{datatype_slug}`.
 *
 * @remarks
 * **Filter syntax:** The `filters` record supports both exact-match and operator-based
 * filtering. For exact match, use the field name as the key. For operator-based filtering,
 * use the format `field[op]` where `op` is one of:
 * - `eq` -- equal (same as exact match)
 * - `ne` -- not equal
 * - `gt` -- greater than
 * - `gte` -- greater than or equal
 * - `lt` -- less than
 * - `lte` -- less than or equal
 * - `like` -- SQL LIKE pattern match (use `%` as wildcard)
 * - `in` -- comma-separated list of values
 *
 * **Sort syntax:** Prefix the field name with `-` for descending order. Only one
 * sort field is supported per request.
 *
 * **Pagination:** Use `limit` and `offset` for cursor-free pagination.
 * The response includes `total` for calculating page counts.
 *
 * @example
 * ```ts
 * // Fetch the 10 most recently published blog posts
 * const params: QueryParams = {
 *   sort: '-published_at',
 *   limit: 10,
 *   status: 'published',
 * }
 * ```
 *
 * @example
 * ```ts
 * // Filter by category with pagination
 * const params: QueryParams = {
 *   filters: { category: 'tutorials' },
 *   limit: 20,
 *   offset: 40, // page 3
 * }
 * ```
 *
 * @example
 * ```ts
 * // Operator-based filtering: posts with views > 100
 * const params: QueryParams = {
 *   filters: { 'views[gt]': '100' },
 *   sort: '-views',
 * }
 * ```
 *
 * @example
 * ```ts
 * // Multi-locale query for French content
 * const params: QueryParams = {
 *   locale: 'fr',
 *   status: 'published',
 * }
 * ```
 */
export interface QueryParams {
  /**
   * Sort field name. Prefix with `-` for descending order.
   *
   * Sortable fields include any field defined on the datatype, plus the
   * built-in fields `date_created`, `date_modified`, and `published_at`.
   *
   * @example `'title'` -- ascending by title
   * @example `'-published_at'` -- descending by publish date (newest first)
   */
  sort?: string
  /**
   * Maximum number of items to return.
   *
   * @defaultValue `20`
   * @remarks Clamped to a server-side maximum of `100`. Values above 100 are silently reduced.
   */
  limit?: number
  /**
   * Number of items to skip for offset-based pagination.
   *
   * @defaultValue `0`
   * @remarks Use in combination with {@link QueryResult.total} to calculate page boundaries.
   */
  offset?: number
  /**
   * Locale code for filtering internationalized content (e.g. `'en'`, `'fr'`, `'de'`).
   *
   * When set, only field values matching this locale are included in the response.
   * When omitted, the server returns the default locale.
   */
  locale?: string
  /**
   * Content status filter.
   *
   * Common values: `'published'`, `'draft'`, `'archived'`.
   *
   * @defaultValue `'published'`
   */
  status?: string
  /**
   * Field filters as key-value pairs.
   *
   * Keys are field names, optionally suffixed with an operator in brackets.
   * Values are always strings (the server handles type coercion).
   *
   * @remarks
   * Supported operator suffixes: `[eq]`, `[ne]`, `[gt]`, `[gte]`, `[lt]`, `[lte]`, `[like]`, `[in]`.
   * A bare field name (no operator) is treated as `[eq]`.
   *
   * @example `{ category: 'news' }` -- exact match
   * @example `{ 'price[gte]': '10', 'price[lte]': '50' }` -- range filter
   * @example `{ 'tag[in]': 'go,rust,zig' }` -- match any of the listed values
   * @example `{ 'title[like]': '%tutorial%' }` -- pattern match
   */
  filters?: Record<string, string>
}

/**
 * A single content item in a query result.
 *
 * Represents one row of content data with its field values flattened into
 * a string-keyed map. Field keys correspond to the `name` property of the
 * field definitions on the content's datatype.
 *
 * @remarks
 * All field values are serialized as strings regardless of their underlying
 * {@link FieldType}. Consumers should parse numeric, boolean, JSON, and date
 * field values according to the field type metadata from the schema API.
 */
export interface QueryItem {
  /** ULID of this content data record. */
  content_data_id: string
  /** ULID of the datatype that defines this content's schema. */
  datatype_id: string
  /** ULID of the user who created this content. */
  author_id: string
  /** Publication lifecycle status (e.g. `'published'`, `'draft'`, `'archived'`). */
  status: string
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
  /** ISO 8601 timestamp when this content was published, or empty string if unpublished. */
  published_at: string
  /**
   * Field values keyed by field name.
   *
   * Keys are the `name` property of the datatype's field definitions.
   * Values are always strings; parse according to the field's {@link FieldType}.
   *
   * @example `{ title: 'Hello World', body: '<p>Content</p>', views: '42' }`
   */
  fields: Record<string, string>
}

/**
 * Datatype metadata included in a query result.
 *
 * Provides the identity of the datatype that was queried, so consumers
 * can display or route based on the content type without a separate
 * schema lookup.
 */
export interface QueryDatatype {
  /** Machine-readable datatype name (matches the slug used in the query URL). */
  name: string
  /** Human-readable display label for this datatype. */
  label: string
}

/**
 * Response envelope for a content query.
 *
 * Contains the matched content items, pagination metadata, and the
 * identity of the queried datatype.
 *
 * @remarks
 * Use {@link total}, {@link limit}, and {@link offset} to implement
 * pagination controls. The total number of pages is `Math.ceil(total / limit)`.
 *
 * @example
 * ```ts
 * const result: QueryResult = await client.query('blog-posts', {
 *   sort: '-published_at',
 *   limit: 10,
 * })
 *
 * console.log(`Showing ${result.data.length} of ${result.total} items`)
 * const totalPages = Math.ceil(result.total / result.limit)
 * const currentPage = Math.floor(result.offset / result.limit) + 1
 * ```
 */
export interface QueryResult {
  /** Array of content items matching the query. May be empty if no results match. */
  data: QueryItem[]
  /** Total number of items matching the query (ignoring limit/offset). */
  total: number
  /** The limit that was applied to this query (echoed from the request, or the default). */
  limit: number
  /** The offset that was applied to this query (echoed from the request, or 0). */
  offset: number
  /** Metadata about the datatype that was queried. */
  datatype: QueryDatatype
}
