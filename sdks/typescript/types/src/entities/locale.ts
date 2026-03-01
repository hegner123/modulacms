/**
 * Locale entity type for i18n support.
 *
 * @module entities/locale
 */

import type { LocaleID } from '../ids.js'

/**
 * A locale configuration for i18n content.
 */
export type Locale = {
  /** Unique identifier for this locale. */
  locale_id: LocaleID
  /** BCP 47 language code (e.g. `"en"`, `"fr"`, `"de"`). */
  code: string
  /** Human-readable label for this locale (e.g. `"English"`, `"French"`). */
  label: string
  /** Whether this is the default locale. */
  is_default: boolean
  /** Whether this locale is currently enabled. */
  is_enabled: boolean
  /** BCP 47 code of the fallback locale, or empty string if none. */
  fallback_code: string
  /** Display ordering position. */
  sort_order: number
  /** ISO 8601 creation timestamp. */
  date_created: string
}
