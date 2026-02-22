/**
 * Media asset entity types and their create/update parameter shapes.
 * Entity types are re-exported from @modulacms/types; param types are local.
 *
 * @module types/media
 */

import type { MediaID, URL, UserID } from './common.js'

// Re-export shared entity types
export type { Media, MediaDimension } from '@modulacms/types'

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/** Parameters for creating a media record via `POST /media`. */
export type CreateMediaParams = {
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

/** Parameters for creating a media dimension preset via `POST /mediadimensions`. */
export type CreateMediaDimensionParams = {
  /** Label for this dimension, or `null`. */
  label: string | null
  /** Width in pixels, or `null`. */
  width: number | null
  /** Height in pixels, or `null`. */
  height: number | null
  /** Aspect ratio string, or `null`. */
  aspect_ratio: string | null
}

// ---------------------------------------------------------------------------
// Update params
// ---------------------------------------------------------------------------

/** Parameters for updating a media record via `PUT /media/`. */
export type UpdateMediaParams = {
  /** ID of the media asset to update. */
  media_id: MediaID
  /** Updated filename, or `null`. */
  name: string | null
  /** Updated display name, or `null`. */
  display_name: string | null
  /** Updated alt text, or `null`. */
  alt: string | null
  /** Updated caption, or `null`. */
  caption: string | null
  /** Updated description, or `null`. */
  description: string | null
  /** Updated CSS class, or `null`. */
  class: string | null
  /** Updated MIME type, or `null`. */
  mimetype: string | null
  /** Updated dimensions JSON, or `null`. */
  dimensions: string | null
  /** Updated URL. */
  url: URL
  /** Updated srcset, or `null`. */
  srcset: string | null
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for updating a media dimension preset via `PUT /mediadimensions/`. */
export type UpdateMediaDimensionParams = {
  /** ID of the dimension preset to update. */
  md_id: string
  /** Updated label, or `null`. */
  label: string | null
  /** Updated width, or `null`. */
  width: number | null
  /** Updated height, or `null`. */
  height: number | null
  /** Updated aspect ratio, or `null`. */
  aspect_ratio: string | null
}

// ---------------------------------------------------------------------------
// Media health/cleanup
// ---------------------------------------------------------------------------

/** Response from `GET /media/health` showing orphaned S3 objects. */
export type MediaHealthResponse = {
  /** Total number of objects in the media bucket. */
  total_objects: number
  /** Number of objects with a corresponding database record. */
  tracked_keys: number
  /** S3 keys with no database record. */
  orphaned_keys: string[]
  /** Count of orphaned keys. */
  orphan_count: number
}

/** Response from `DELETE /media/cleanup` after deleting orphaned S3 objects. */
export type MediaCleanupResponse = {
  /** S3 keys that were successfully deleted. */
  deleted: string[]
  /** Number of keys deleted. */
  deleted_count: number
  /** S3 keys that failed to delete. */
  failed: string[]
  /** Number of keys that failed to delete. */
  failed_count: number
}
