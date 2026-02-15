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
 * - `'archived'` - Removed from public view but retained.
 * - `'pending'` - Awaiting review or approval.
 */
export type ContentStatus = 'draft' | 'published' | 'archived' | 'pending'

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
  | 'relation'
  | 'json'
  | 'richtext'
  | 'slug'
  | 'email'
  | 'url'

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
