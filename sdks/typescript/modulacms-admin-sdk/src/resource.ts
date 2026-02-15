/**
 * Generic CRUD resource factory that generates standard list/get/create/update/remove
 * methods for any entity type using a shared {@link HttpClient}.
 *
 * @module resource
 * @internal
 */

import type { HttpClient } from './http.js'
import type { RequestOptions, PaginationParams, PaginatedResponse } from './types/common.js'

// ---------------------------------------------------------------------------
// CRUD resource type
// ---------------------------------------------------------------------------

/**
 * Standard CRUD (Create, Read, Update, Delete) operations for an API resource.
 *
 * @typeParam Entity - The entity type returned by the API.
 * @typeParam CreateParams - Parameters required to create a new entity.
 * @typeParam UpdateParams - Parameters required to update an existing entity.
 * @typeParam Id - The identifier type used for get/remove (defaults to `string`).
 *
 * @example
 * ```ts
 * const user = await client.users.get(userId)
 * const allUsers = await client.users.list()
 * ```
 */
type CrudResource<Entity, CreateParams, UpdateParams, Id = string> = {
  /**
   * List all entities of this resource type.
   * @param opts - Optional request options.
   * @returns Array of all entities.
   */
  list: (opts?: RequestOptions) => Promise<Entity[]>

  /**
   * Get a single entity by its identifier.
   * @param id - The entity's unique identifier.
   * @param opts - Optional request options.
   * @returns The matching entity.
   */
  get: (id: Id, opts?: RequestOptions) => Promise<Entity>

  /**
   * Create a new entity.
   * @param params - The creation parameters.
   * @param opts - Optional request options.
   * @returns The newly created entity.
   */
  create: (params: CreateParams, opts?: RequestOptions) => Promise<Entity>

  /**
   * Update an existing entity.
   * @param params - The update parameters (must include the entity identifier).
   * @param opts - Optional request options.
   * @returns The updated entity.
   */
  update: (params: UpdateParams, opts?: RequestOptions) => Promise<Entity>

  /**
   * Remove an entity by its identifier.
   * @param id - The entity's unique identifier.
   * @param opts - Optional request options.
   */
  remove: (id: Id, opts?: RequestOptions) => Promise<void>

  /**
   * List entities with pagination. When `limit` or `offset` are present the
   * server returns a paginated envelope instead of a plain array.
   * @param params - Pagination parameters (limit and offset).
   * @param opts - Optional request options.
   * @returns A paginated envelope containing the data page and total count.
   */
  listPaginated: (params: PaginationParams, opts?: RequestOptions) => Promise<PaginatedResponse<Entity>>

  /**
   * Return the total number of entities of this resource type.
   * Implemented as a zero-limit paginated request.
   * @param opts - Optional request options.
   * @returns The total entity count.
   */
  count: (opts?: RequestOptions) => Promise<number>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create a standard CRUD resource bound to a specific API path.
 *
 * URL patterns:
 * - `list` : `GET /api/v1/{path}`
 * - `get`  : `GET /api/v1/{path}/?q={id}`
 * - `create`: `POST /api/v1/{path}`
 * - `update`: `PUT /api/v1/{path}/`
 * - `remove`: `DELETE /api/v1/{path}/?q={id}`
 *
 * @typeParam Entity - The entity type.
 * @typeParam CreateParams - Creation parameter type.
 * @typeParam UpdateParams - Update parameter type.
 * @typeParam Id - Identifier type (defaults to `string`).
 * @param http - The configured HTTP client.
 * @param path - The API resource path (e.g. `'users'`, `'adminroutes'`).
 * @returns A {@link CrudResource} with all five CRUD operations.
 * @internal
 */
function createResource<Entity, CreateParams, UpdateParams, Id = string>(
  http: HttpClient,
  path: string,
): CrudResource<Entity, CreateParams, UpdateParams, Id> {
  return {
    list(opts?: RequestOptions): Promise<Entity[]> {
      return http.get<Entity[]>(`/${path}`, undefined, opts)
    },

    get(id: Id, opts?: RequestOptions): Promise<Entity> {
      return http.get<Entity>(`/${path}/`, { q: String(id) }, opts)
    },

    create(params: CreateParams, opts?: RequestOptions): Promise<Entity> {
      return http.post<Entity>(`/${path}`, params as Record<string, unknown>, opts)
    },

    update(params: UpdateParams, opts?: RequestOptions): Promise<Entity> {
      return http.put<Entity>(`/${path}/`, params as Record<string, unknown>, opts)
    },

    remove(id: Id, opts?: RequestOptions): Promise<void> {
      return http.del(`/${path}/`, { q: String(id) }, opts)
    },

    listPaginated(params: PaginationParams, opts?: RequestOptions): Promise<PaginatedResponse<Entity>> {
      return http.get<PaginatedResponse<Entity>>(`/${path}`, {
        limit: String(params.limit),
        offset: String(params.offset),
      }, opts)
    },

    async count(opts?: RequestOptions): Promise<number> {
      const result = await this.listPaginated({ limit: 0, offset: 0 }, opts)
      return result.total
    },
  }
}

export type { CrudResource }
export { createResource }
