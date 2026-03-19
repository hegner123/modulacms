/**
 * Admin media folder entity types and their create/update parameter shapes.
 * Entity types are re-exported from @modulacms/types; param types are local.
 *
 * @module types/admin-media-folders
 */

import type { AdminMediaFolderID } from './common.js'

// Re-export shared entity types
export type { AdminMediaFolder } from '@modulacms/types'

// Re-export param types from @modulacms/types
export type { CreateAdminMediaFolderParams, UpdateAdminMediaFolderParams } from '@modulacms/types'

import type { AdminMediaFolder } from '@modulacms/types'

/** A node in the admin media folder tree with recursive children. */
export type AdminMediaFolderTreeNode = AdminMediaFolder & {
  /** Child folders nested under this folder. */
  children: AdminMediaFolderTreeNode[]
}
