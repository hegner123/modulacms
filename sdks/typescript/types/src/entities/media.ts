/**
 * Media asset entity types.
 *
 * @module entities/media
 */

import type { MediaID, URL, UserID } from '../ids.js'

/**
 * A media asset (image, video, document, etc.) stored in the CMS.
 */
export type Media = {
  /** Unique identifier for this media asset. */
  media_id: MediaID
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
  /** ID of the user who uploaded this asset, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * A named dimension preset for responsive media rendering.
 */
export type MediaDimension = {
  /** Unique identifier for this dimension preset. */
  md_id: string
  /** Human-readable label (e.g. `'thumbnail'`, `'hero'`), or `null`. */
  label: string | null
  /** Target width in pixels, or `null` for unconstrained. */
  width: number | null
  /** Target height in pixels, or `null` for unconstrained. */
  height: number | null
  /** Aspect ratio string (e.g. `'16:9'`), or `null`. */
  aspect_ratio: string | null
}
