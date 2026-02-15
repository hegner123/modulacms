/**
 * Import request and response types for bulk content migration from external CMS platforms.
 *
 * @module types/import
 */

/**
 * Supported CMS platform formats for content import.
 *
 * - `'contentful'` - Contentful export format.
 * - `'sanity'` - Sanity.io export format.
 * - `'strapi'` - Strapi export format.
 * - `'wordpress'` - WordPress export format.
 * - `'clean'` - ModulaCMS native clean format.
 */
export type ImportFormat = 'contentful' | 'sanity' | 'strapi' | 'wordpress' | 'clean'

/**
 * Server response from a bulk import operation.
 * Contains creation counts and any errors encountered during import.
 */
export type ImportResponse = {
  /** Whether the import completed without fatal errors. */
  success: boolean
  /** Number of datatype definitions created. */
  datatypes_created: number
  /** Number of field definitions created. */
  fields_created: number
  /** Number of content items created. */
  content_created: number
  /** Human-readable summary message. */
  message: string
  /** List of non-fatal error messages encountered during import. */
  errors: string[]
}
