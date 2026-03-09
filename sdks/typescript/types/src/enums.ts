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
 * Lifecycle status of a plugin in the persistent registry.
 *
 * - `'installed'` - Registered but not active.
 * - `'enabled'` - Active and running.
 */
export type PluginStatus = 'installed' | 'enabled'

/**
 * Database-level mutation type recorded in change_events for audit trail.
 */
export type Operation = 'INSERT' | 'UPDATE' | 'DELETE'

/**
 * Business-level action recorded in audit log entries.
 */
export type Action = 'create' | 'update' | 'delete' | 'publish'

/**
 * Conflict resolution policy for a datatype in distributed or concurrent editing.
 *
 * - `'lww'` - Last write wins (simple, possible data loss).
 * - `'manual'` - Flags conflicts for human resolution.
 */
export type ConflictPolicy = 'lww' | 'manual'

/**
 * Type of backup operation.
 */
export type BackupType = 'full' | 'incremental' | 'differential'

/**
 * Status of a backup operation.
 */
export type BackupStatus = 'pending' | 'in_progress' | 'completed' | 'failed'

/**
 * Status of a backup verification check.
 */
export type VerificationStatus = 'pending' | 'verified' | 'failed'

/**
 * Status of a backup set (collection of backups).
 */
export type BackupSetStatus = 'pending' | 'complete' | 'partial'

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
