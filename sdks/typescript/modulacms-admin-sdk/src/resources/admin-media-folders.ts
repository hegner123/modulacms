/**
 * Admin media folders resource for managing admin media folder hierarchy.
 *
 * Provides CRUD operations plus tree retrieval and folder media listing
 * at `/api/v1/adminmedia-folders`.
 *
 * @module resources/admin-media-folders
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions, PaginationParams, PaginatedResponse, AdminMediaFolderID } from '../types/common.js'
import type { AdminMediaFolder } from '../types/admin-media-folders.js'
import type { CreateAdminMediaFolderParams, UpdateAdminMediaFolderParams } from '../types/admin-media-folders.js'
import type { AdminMediaFolderTreeNode } from '../types/admin-media-folders.js'
import type { AdminMedia } from '../types/admin-media.js'

/**
 * Admin media folder resource operations available on `client.adminMediaFolders`.
 */
type AdminMediaFoldersResource = {
  /** List admin media folders (root-level by default, or children of a parent). */
  list: (parentId?: AdminMediaFolderID, opts?: RequestOptions) => Promise<AdminMediaFolder[]>
  /** Get a single admin media folder by ID. */
  get: (id: AdminMediaFolderID, opts?: RequestOptions) => Promise<AdminMediaFolder>
  /** Create a new admin media folder. */
  create: (params: CreateAdminMediaFolderParams, opts?: RequestOptions) => Promise<AdminMediaFolder>
  /** Update an admin media folder (rename or move). */
  update: (id: AdminMediaFolderID, params: UpdateAdminMediaFolderParams, opts?: RequestOptions) => Promise<AdminMediaFolder>
  /** Delete an admin media folder. Fails if folder has children or media. */
  remove: (id: AdminMediaFolderID, opts?: RequestOptions) => Promise<void>
  /** Get the full admin media folder tree hierarchy. */
  tree: (opts?: RequestOptions) => Promise<AdminMediaFolderTreeNode[]>
  /** List admin media items in a folder (paginated). */
  listMedia: (id: AdminMediaFolderID, params?: PaginationParams, opts?: RequestOptions) => Promise<PaginatedResponse<AdminMedia>>
}

/**
 * Create the admin media folders resource.
 *
 * @param http - Configured HTTP client.
 * @returns An {@link AdminMediaFoldersResource} with CRUD, tree, and media listing.
 * @internal
 */
function createAdminMediaFoldersResource(http: HttpClient): AdminMediaFoldersResource {
  return {
    list(parentId?: AdminMediaFolderID, opts?: RequestOptions): Promise<AdminMediaFolder[]> {
      const params: Record<string, string> = {}
      if (parentId !== undefined) {
        params['parent_id'] = String(parentId)
      }
      return http.get<AdminMediaFolder[]>('/adminmedia-folders', Object.keys(params).length > 0 ? params : undefined, opts)
    },

    get(id: AdminMediaFolderID, opts?: RequestOptions): Promise<AdminMediaFolder> {
      return http.get<AdminMediaFolder>('/adminmedia-folders/' + encodeURIComponent(String(id)), undefined, opts)
    },

    create(params: CreateAdminMediaFolderParams, opts?: RequestOptions): Promise<AdminMediaFolder> {
      return http.post<AdminMediaFolder>('/adminmedia-folders', params as unknown as Record<string, unknown>, opts)
    },

    update(id: AdminMediaFolderID, params: UpdateAdminMediaFolderParams, opts?: RequestOptions): Promise<AdminMediaFolder> {
      return http.put<AdminMediaFolder>('/adminmedia-folders/' + encodeURIComponent(String(id)), params as unknown as Record<string, unknown>, opts)
    },

    remove(id: AdminMediaFolderID, opts?: RequestOptions): Promise<void> {
      return http.del('/adminmedia-folders/' + encodeURIComponent(String(id)), undefined, opts)
    },

    tree(opts?: RequestOptions): Promise<AdminMediaFolderTreeNode[]> {
      return http.get<AdminMediaFolderTreeNode[]>('/adminmedia-folders/tree', undefined, opts)
    },

    listMedia(id: AdminMediaFolderID, params?: PaginationParams, opts?: RequestOptions): Promise<PaginatedResponse<AdminMedia>> {
      const q: Record<string, string> = {}
      if (params) {
        q['limit'] = String(params.limit)
        q['offset'] = String(params.offset)
      }
      return http.get<PaginatedResponse<AdminMedia>>('/adminmedia-folders/' + encodeURIComponent(String(id)) + '/media', Object.keys(q).length > 0 ? q : undefined, opts)
    },
  }
}

export type { AdminMediaFoldersResource }
export { createAdminMediaFoldersResource }
