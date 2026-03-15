/**
 * Media folder entity types and their create/update parameter shapes.
 * Entity types are re-exported from @modulacms/types; param types are local.
 *
 * @module types/media-folders
 */

import type { MediaFolderID, MediaID, PaginatedResponse } from './common.js'

// Re-export shared entity type
export type { MediaFolder } from '@modulacms/types'

// Re-export param types from @modulacms/types
export type { CreateMediaFolderParams, UpdateMediaFolderParams } from '@modulacms/types'

import type { MediaFolder } from '@modulacms/types'
import type { Media } from '@modulacms/types'

/** A node in the media folder tree with recursive children. */
export type MediaFolderTreeNode = MediaFolder & {
  /** Child folders nested under this folder. */
  children: MediaFolderTreeNode[]
}

/** Parameters for batch-moving media items to a folder. */
export type MoveMediaParams = {
  /** IDs of media items to move. */
  media_ids: MediaID[]
  /** Target folder ID, or `null` to move to root. */
  folder_id: MediaFolderID | null
}

/** Response from the batch move media endpoint. */
export type MoveMediaResponse = {
  /** Number of media items moved. */
  moved: number
}
