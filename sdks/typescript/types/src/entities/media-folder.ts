/**
 * Media folder entity types.
 *
 * @module entities/media-folder
 */

import type { MediaFolderID } from '../ids.js'

/**
 * A folder for organizing media assets.
 */
export type MediaFolder = {
  /** Unique identifier for this media folder. */
  folder_id: MediaFolderID
  /** Display name of the folder. */
  name: string
  /** ID of the parent folder, or `null` for root-level folders. */
  parent_id: MediaFolderID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * Parameters for creating a new media folder.
 */
export type CreateMediaFolderParams = {
  /** Display name for the folder. */
  name: string
  /** ID of the parent folder, or `null` for root-level. */
  parent_id: MediaFolderID | null
}

/**
 * Parameters for updating a media folder (rename or move).
 */
export type UpdateMediaFolderParams = {
  /** ID of the folder to update. */
  folder_id: MediaFolderID
  /** Updated name for the folder. */
  name: string
  /** Updated parent folder ID, or `null` to move to root. */
  parent_id: MediaFolderID | null
}
