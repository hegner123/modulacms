/**
 * Admin media asset entity types and their create/update parameter shapes.
 * Entity types are re-exported from @modulacms/types; param types are local.
 *
 * @module types/admin-media
 */

import type { AdminMediaID, AdminMediaFolderID, URL, UserID } from './common.js'

// Re-export shared entity types
export type { AdminMedia } from '@modulacms/types'

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/** Parameters for creating an admin media record via `POST /adminmedia` (multipart upload). */
export type CreateAdminMediaParams = {
  /** Internal filename, or `null`. */
  name: string | null
  /** Display name, or `null`. */
  display_name: string | null
  /** Alt text, or `null`. */
  alt: string | null
  /** Caption, or `null`. */
  caption: string | null
  /** Description, or `null`. */
  description: string | null
  /** CSS class, or `null`. */
  class: string | null
  /** MIME type, or `null`. */
  mimetype: string | null
  /** JSON-encoded dimensions, or `null`. */
  dimensions: string | null
  /** Primary asset URL. */
  url: URL
  /** Srcset value, or `null`. */
  srcset: string | null
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

// ---------------------------------------------------------------------------
// Update params
// ---------------------------------------------------------------------------

/** Parameters for updating admin media metadata via `PUT /adminmedia/`. */
export type UpdateAdminMediaParams = {
  /** ID of the admin media asset to update. */
  admin_media_id: AdminMediaID
  /** Updated display name. */
  display_name: string
  /** Updated alt text. */
  alt: string
  /** Updated caption. */
  caption: string
  /** Updated description. */
  description: string
  /** Updated focal point X coordinate (0.0-1.0), or `null`. */
  focal_x: number | null
  /** Updated focal point Y coordinate (0.0-1.0), or `null`. */
  focal_y: number | null
  /** Updated folder ID, or empty string to move to root, or `null` to keep current. */
  folder_id?: string | null
}

// ---------------------------------------------------------------------------
// Move params
// ---------------------------------------------------------------------------

/** Parameters for batch-moving admin media items to a folder. */
export type MoveAdminMediaParams = {
  /** IDs of admin media items to move. */
  media_ids: string[]
  /** Target folder ID, or `null` to move to root. */
  folder_id: string | null
}

/** Response from the batch move admin media endpoint. */
export type MoveAdminMediaResponse = {
  /** Number of admin media items moved. */
  moved: number
}
