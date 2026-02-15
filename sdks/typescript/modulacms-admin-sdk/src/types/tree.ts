/**
 * Admin content tree response types.
 * Used by the admin tree endpoint to return fully resolved content hierarchies.
 *
 * @module types/tree
 */

import type { AdminContentID, FieldType } from './common.js'
import type { AdminRoute } from './admin.js'

/**
 * A resolved field within a content tree node.
 * Contains the field's label, type, and current value.
 */
export type ContentTreeField = {
  /** Human-readable field label. */
  field_label: string
  /** The data type of this field. */
  field_type: FieldType
  /** Current field value, or `null` if unset. */
  field_value: string | null
}

/**
 * A node in the admin content tree with its resolved fields and children.
 * Represents a fully hydrated content item including its tree position,
 * datatype metadata, field values, and recursive child nodes.
 */
export type ContentTreeNode = {
  /** Unique identifier for this content node. */
  content_data_id: AdminContentID
  /** Parent node ID, or `null` for root nodes. */
  parent_id: AdminContentID | null
  /** First child node ID, or `null`. */
  first_child_id: string | null
  /** Next sibling node ID, or `null`. */
  next_sibling_id: string | null
  /** Previous sibling node ID, or `null`. */
  prev_sibling_id: string | null
  /** Label of the datatype that defines this node's schema. */
  datatype_label: string
  /** Category of the datatype. */
  datatype_type: string
  /** Resolved field values for this node. */
  fields: ContentTreeField[]
  /** Recursively resolved child nodes. */
  children: ContentTreeNode[]
}

/**
 * Complete admin tree response containing the route metadata and
 * the full content tree rooted at that route.
 */
export type AdminTreeResponse = {
  /** The admin route this tree belongs to. */
  route: AdminRoute
  /** Root-level content nodes with recursively resolved children. */
  tree: ContentTreeNode[]
}

/**
 * Output format for the admin tree endpoint.
 *
 * - `'raw'` - Native ModulaCMS tree structure ({@link AdminTreeResponse}).
 * - `'contentful'` - Contentful-compatible JSON.
 * - `'sanity'` - Sanity.io-compatible JSON.
 * - `'strapi'` - Strapi-compatible JSON.
 * - `'wordpress'` - WordPress-compatible JSON.
 * - `'clean'` - ModulaCMS clean export format.
 */
export type TreeFormat = 'contentful' | 'sanity' | 'strapi' | 'wordpress' | 'clean' | 'raw'
