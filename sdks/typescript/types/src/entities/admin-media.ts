/**
 * Admin media asset and admin media folder entity types.
 *
 * @module entities/admin-media
 */

import type { AdminMediaID, AdminMediaFolderID, URL, UserID } from '../ids.js'

/**
 * An admin media asset (image, video, document, etc.) stored in the CMS
 * for use in the admin panel UI.
 */
export type AdminMedia = {
  /** Unique identifier for this admin media asset. */
  admin_media_id: AdminMediaID
  /** Internal filename, or `null`. */
  name: string | null
  /** Human-readable display name, or `null`. */
  display_name: string | null
  /** Alternative text for accessibility, or `null`. */
  alt: string | null
  /** Caption text, or `null`. */
  caption: string | null
  /** Extended description, or `null`. */
  description: string | null
  /** CSS class name for styling, or `null`. */
  class: string | null
  /** MIME type (e.g. `'image/png'`), or `null`. */
  mimetype: string | null
  /** JSON-encoded dimension data, or `null`. */
  dimensions: string | null
  /** Primary URL where the asset is served. */
  url: URL
  /** Responsive `srcset` attribute value, or `null`. */
  srcset: string | null
  /** Focal point X coordinate (0.0-1.0), or `null`. */
  focal_x: number | null
  /** Focal point Y coordinate (0.0-1.0), or `null`. */
  focal_y: number | null
  /** ID of the user who uploaded this asset, or `null`. */
  author_id: UserID | null
  /** ID of the admin folder containing this asset, or `null` for root. */
  folder_id: AdminMediaFolderID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
  /** Relative URL for downloading the file with the correct display filename. */
  download_url: string
}

/**
 * An admin media folder for organizing admin media assets.
 */
export type AdminMediaFolder = {
  /** Unique identifier for this admin media folder. */
  admin_folder_id: AdminMediaFolderID
  /** Display name of the folder. */
  name: string
  /** ID of the parent folder, or `null` for root-level folders. */
  parent_id: AdminMediaFolderID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * Parameters for creating a new admin media folder.
 */
export type CreateAdminMediaFolderParams = {
  /** Display name for the folder. */
  name: string
  /** ID of the parent folder, or empty string for root-level. */
  parent_id: string
}

/**
 * Parameters for updating an admin media folder (rename or move).
 */
export type UpdateAdminMediaFolderParams = {
  /** Updated name for the folder, or omit to keep current. */
  name?: string
  /** Updated parent folder ID, or empty string to move to root, or omit to keep current. */
  parent_id?: string
}
