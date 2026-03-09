/**
 * Enum union types and runtime constants for ModulaCMS.
 *
 * @module enums
 */

/**
 * Lifecycle status of a content item.
 *
 * - `'draft'` - Work in progress, not publicly visible.
 * - `'published'` - Live and publicly accessible.
 */
export type ContentStatus = 'draft' | 'published'

/**
 * Supported field types for content schema definitions.
 * Determines the editor widget and validation rules applied to a field.
 */
export type FieldType =
  | 'text'
  | 'textarea'
  | 'number'
  | 'date'
  | 'datetime'
  | 'boolean'
  | 'select'
  | 'media'
  | '_id'
  | 'json'
  | 'richtext'
  | 'slug'
  | 'email'
  | 'url'

/**
 * Classifies how a route maps incoming URL patterns to content or behavior.
 *
 * - `'static'` - Fixed URL path mapped to a single content item.
 * - `'dynamic'` - URL pattern with parameters resolved at request time.
 * - `'api'` - Custom API endpoint, typically registered by a plugin.
 * - `'redirect'` - Redirect to another URL (returns metadata, not a 301).
 */
export type RouteType = 'static' | 'dynamic' | 'api' | 'redirect'

/**
 * All supported content output format identifiers.
 * Use this array for runtime validation of format values.
 *
 * @example
 * if (CONTENT_FORMATS.includes(userInput)) { ... }
 */
export const CONTENT_FORMATS = ["contentful", "sanity", "strapi", "wordpress", "clean", "raw"] as const;

/**
 * A content output format accepted by the `?format=` query parameter.
 * Derived from {@link CONTENT_FORMATS} to keep the runtime array and type in sync.
 */
export type ContentFormat = (typeof CONTENT_FORMATS)[number];
