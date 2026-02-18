/**
 * Content tree types returned by slug-based content delivery.
 * These represent the assembled, hierarchical content tree that the CMS
 * builds from flat database rows when serving a page by slug.
 *
 * @module entities/tree
 */

import type { ContentData, ContentField } from './content.js'
import type { Datatype, Field } from './schema.js'

/**
 * A datatype paired with its content instance data.
 * `info` is the type definition; `content` is the specific content node's instance data.
 */
export type NodeDatatype = {
  /** Datatype definition (label, type, IDs). */
  info: Datatype
  /** Content instance data (tree pointers, route, status, dates). */
  content: ContentData
}

/**
 * A field definition paired with its content field value.
 * `info` is the field schema; `content` holds the actual stored value.
 */
export type NodeField = {
  /** Field definition (label, type, validation, ui_config). */
  info: Field
  /** Content field value instance. */
  content: ContentField
}

/**
 * A single node in the content tree hierarchy.
 * Each node has a datatype, field values, and optional child nodes.
 */
export type ContentNode = {
  /** The datatype definition and content instance for this node. */
  datatype: NodeDatatype
  /** Field values belonging to this node. */
  fields: NodeField[]
  /** Child nodes. Omitted or empty when this is a leaf node. */
  nodes?: ContentNode[]
}

/**
 * The top-level content tree returned by the slug handler.
 * Contains a single root node that forms the entry point of the hierarchy.
 */
export type ContentTree = {
  /** The root node of the content tree. */
  root: ContentNode
}
